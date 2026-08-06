import asyncio
import os
import time
from concurrent.futures import ProcessPoolExecutor

from app.logger import logger, request_id_var
from app.pb import analysis_pb2, analysis_pb2_grpc
from app.services.llm_service import AnalysisEngineError, analyze_fit
from app.services.pdf_service import extract_text_from_pdf_bytes

MAX_CONCURRENT_PDF_PARSERS = int(os.environ["MAX_CONCURRENT_PDF_PARSERS"])
LLM_MODEL = os.environ["LLM_MODEL"]
LLM_PROVIDER = os.environ.get("LLM_PROVIDER", "gemini")

pdf_bouncer = asyncio.Semaphore(MAX_CONCURRENT_PDF_PARSERS)
pdf_executor = ProcessPoolExecutor(max_workers=MAX_CONCURRENT_PDF_PARSERS)


async def _parse_pdf_safely(pdf_bytes: bytes, doc_name: str) -> tuple[str, dict]:
    """
    Safely extracts text from PDF bytes within established concurrency limits.

    Uses the global semaphore to bound concurrent CPU-intensive parsing.
    Parsing itself runs inside a ProcessPoolExecutor to keep the asyncio event
    loop responsive.

    Returns:
        (extracted_text, metrics)
    """
    wait_start = time.perf_counter()

    async with pdf_bouncer:
        queue_wait = time.perf_counter() - wait_start

        try:
            parse_start = time.perf_counter()

            text = await asyncio.get_running_loop().run_in_executor(
                pdf_executor,
                extract_text_from_pdf_bytes,
                pdf_bytes,
            )

            duration = time.perf_counter() - parse_start

            return text, {
                "queue_wait": queue_wait,
                "duration": duration,
            }

        except asyncio.CancelledError:
            raise

        except Exception:
            logger.exception(
                "Failed to parse PDF",
                extra={"document": doc_name},
            )
            raise AnalysisEngineError(
                "ERR_PDF_PARSE",
                f"Failed to parse {doc_name}. Please ensure it is a valid text-based PDF.",
            )


class AnalysisEngineServicer(analysis_pb2_grpc.AnalysisEngineServicer):
    """
    gRPC service implementation for the Analysis Engine.
    """

    async def Analyze(self, request, context):
        """
        Extracts text from the supplied PDFs and performs AI analysis.
        """

        request_id_var.set(request.request_id)

        start_time = time.perf_counter()

        logger.info("gRPC Analyze request started")

        try:
            (jd_text, jd_metrics), (resume_text, resume_metrics) = await asyncio.gather(
                _parse_pdf_safely(request.job_description_pdf, "Job Description"),
                _parse_pdf_safely(request.resume_pdf, "Resume"),
            )

            pdf_queue_wait_seconds = (
                jd_metrics["queue_wait"] + resume_metrics["queue_wait"]
            )

            pdf_duration_seconds = jd_metrics["duration"] + resume_metrics["duration"]

        except AnalysisEngineError as exc:
            latency = time.perf_counter() - start_time

            logger.warning(
                "PDF processing failed",
                extra={
                    "analysis.success": False,
                    "analysis.stage": "pdf",
                    "analysis.error_code": exc.code,
                    "analysis.latency_seconds": latency,
                },
            )

            return analysis_pb2.AnalyzeResponse(
                success=False,
                error_message=f"{exc.code}: {exc.message}",
            )

        except asyncio.CancelledError:
            raise

        except Exception:
            latency = time.perf_counter() - start_time

            logger.exception(
                "Unexpected PDF processing error",
                extra={
                    "analysis.success": False,
                    "analysis.stage": "pdf",
                    "analysis.error_code": "ERR_INTERNAL",
                    "analysis.latency_seconds": latency,
                },
            )

            return analysis_pb2.AnalyzeResponse(
                success=False,
                error_message=(
                    "ERR_INTERNAL: An unexpected error occurred while "
                    "processing the documents."
                ),
            )

        try:
            final_analysis, usage = await analyze_fit(
                jd_text,
                resume_text,
            )

        except AnalysisEngineError as exc:
            latency = time.perf_counter() - start_time

            logger.warning(
                "LLM processing failed",
                extra={
                    "analysis.success": False,
                    "analysis.stage": "llm",
                    "analysis.error_code": exc.code,
                    "analysis.latency_seconds": latency,
                },
            )

            return analysis_pb2.AnalyzeResponse(
                success=False,
                error_message=f"{exc.code}: {exc.message}",
            )

        except asyncio.CancelledError:
            raise

        except Exception:
            latency = time.perf_counter() - start_time

            logger.exception(
                "Unexpected LLM processing error",
                extra={
                    "analysis.success": False,
                    "analysis.stage": "llm",
                    "analysis.error_code": "ERR_INTERNAL",
                    "analysis.latency_seconds": latency,
                },
            )

            return analysis_pb2.AnalyzeResponse(
                success=False,
                error_message=(
                    "ERR_INTERNAL: An unexpected error occurred during AI analysis."
                ),
            )

        latency = time.perf_counter() - start_time

        logger.info(
            "gRPC Analyze request completed successfully",
            extra={
                "analysis.success": True,
                "analysis.latency_seconds": latency,
                "analysis.provider": LLM_PROVIDER,
                "analysis.model": LLM_MODEL,
                "analysis.prompt_tokens": usage["prompt_tokens"],
                "analysis.completion_tokens": usage["completion_tokens"],
                "analysis.total_tokens": usage["total_tokens"],
                "analysis.pdf_queue_wait_seconds": pdf_queue_wait_seconds,
                "analysis.pdf_duration_seconds": pdf_duration_seconds,
                "analysis.llm_queue_wait_seconds": usage["queue_wait_seconds"],
                "analysis.llm_retries": usage["retries"],
            },
        )

        return analysis_pb2.AnalyzeResponse(
            success=True,
            result_json=final_analysis.model_dump_json(
                by_alias=True,
                exclude_none=True,
            ),
            model=LLM_MODEL,
            prompt_tokens=usage["prompt_tokens"],
            completion_tokens=usage["completion_tokens"],
            total_tokens=usage["total_tokens"],
        )
