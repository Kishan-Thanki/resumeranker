# Maintainability note:
# This service intentionally keeps gRPC orchestration, PDF processing,
# concurrency, and related error handling together for simplicity.
# If these responsibilities grow significantly as the service evolves,
# consider extracting concerns such as PDF processing/concurrency,
# executor lifecycle, error handling, or metrics into dedicated modules.

import asyncio
import functools
import multiprocessing as mp
import os
import time
import weakref
from concurrent.futures import ProcessPoolExecutor
from typing import Any

from app.logger import logger, request_id_var
from app.pb import analysis_pb2, analysis_pb2_grpc
from app.services.llm import AnalysisEngineError, analyze_fit
from app.services.pdf import extract_text_from_pdf_bytes


def _get_max_pdf_parsers() -> int:
    return int(os.getenv("MAX_CONCURRENT_PDF_PARSERS", "4"))


def _get_llm_config() -> tuple[str, str]:
    model = os.getenv("LLM_MODEL", "gpt-4o")
    provider = (
        model.split("/", 1)[0] if "/" in model else os.getenv("LLM_PROVIDER", "unknown")
    )
    return model, provider


ANALYZE_TIMEOUT_SECONDS = float(os.getenv("ANALYZE_TIMEOUT_SECONDS", "120"))


spawn_context = mp.get_context("spawn")
pdf_executor = ProcessPoolExecutor(
    max_workers=_get_max_pdf_parsers(),
    mp_context=spawn_context,
)

_pdf_bouncers: weakref.WeakValueDictionary[int, asyncio.Semaphore] = (
    weakref.WeakValueDictionary()
)


def _get_pdf_bouncer() -> asyncio.Semaphore:
    """
    Returns the semaphore for the current running event loop, creating one if
    it doesn't exist yet. Keyed by loop id in a WeakValueDictionary so the
    semaphore is automatically dropped when the loop is garbage collected.
    This prevents stale references across event loops (e.g., in pytest teardown
    where each test gets a fresh loop).
    """
    loop = asyncio.get_running_loop()
    bouncer = _pdf_bouncers.get(id(loop))
    if bouncer is None:
        bouncer = asyncio.Semaphore(_get_max_pdf_parsers())
        _pdf_bouncers[id(loop)] = bouncer
    return bouncer


async def shutdown_pdf_executor() -> None:
    """
    Shuts down the PDF worker process pool. Call this during application
    shutdown (see main.py's serve()), after the gRPC server itself has
    stopped accepting new work.
    """
    logger.info("Shutting down PDF worker pool...")
    loop = asyncio.get_running_loop()
    await loop.run_in_executor(
        None,
        functools.partial(pdf_executor.shutdown, wait=True, cancel_futures=True),
    )


def _cancel_pending(*tasks: "asyncio.Task[Any]") -> None:
    """Cancels any of the given tasks that haven't finished yet."""
    for task in tasks:
        if not task.done():
            task.cancel()


async def _parse_pdf_safely(
    pdf_bytes: bytes, doc_name: str
) -> tuple[str, dict[str, float]]:
    """
    Safely extracts text from PDF bytes within established concurrency limits.
    Uses the loop-aware semaphore to bound concurrent CPU-intensive parsing.
    """
    bouncer = _get_pdf_bouncer()
    wait_start = time.perf_counter()

    async with bouncer:
        queue_wait = time.perf_counter() - wait_start
        try:
            parse_start = time.perf_counter()
            loop = asyncio.get_running_loop()
            text = await loop.run_in_executor(
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
        except Exception as exc:
            logger.exception(
                "Failed to parse PDF",
                extra={"document": doc_name},
            )
            raise AnalysisEngineError(
                "ERR_PDF_PARSE",
                f"Failed to parse {doc_name}: {exc}",
                retryable=False,
            ) from exc


class AnalysisEngineServicer(analysis_pb2_grpc.AnalysisEngineServicer):
    """
    gRPC service implementation for the Analysis Engine.
    """

    async def Analyze(self, request: Any, context: Any) -> analysis_pb2.AnalyzeResponse:
        """
        Extracts text from the supplied PDFs and performs AI analysis.
        """
        request_id_var.set(request.request_id)
        start_time = time.perf_counter()
        llm_model, llm_provider = _get_llm_config()
        logger.info("gRPC Analyze request started")

        jd_pdf_task = asyncio.create_task(
            _parse_pdf_safely(request.job_description_pdf, "Job Description")
        )
        resume_pdf_task = asyncio.create_task(
            _parse_pdf_safely(request.resume_pdf, "Resume")
        )

        # Stage 1: PDF Parsing
        try:
            (jd_text, jd_metrics), (resume_text, resume_metrics) = await asyncio.gather(
                jd_pdf_task, resume_pdf_task
            )
            pdf_queue_wait_seconds = (
                jd_metrics["queue_wait"] + resume_metrics["queue_wait"]
            )
            pdf_duration_seconds = jd_metrics["duration"] + resume_metrics["duration"]
        except AnalysisEngineError as exc:
            _cancel_pending(jd_pdf_task, resume_pdf_task)
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
        except Exception:  # noqa: BLE001
            _cancel_pending(jd_pdf_task, resume_pdf_task)
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

        # Stage 2: LLM Analysis
        try:
            final_analysis, usage = await asyncio.wait_for(
                analyze_fit(jd_text, resume_text),
                timeout=ANALYZE_TIMEOUT_SECONDS,
            )
        except asyncio.TimeoutError:
            latency = time.perf_counter() - start_time
            logger.warning(
                "LLM analysis timed out",
                extra={
                    "analysis.success": False,
                    "analysis.stage": "llm",
                    "analysis.error_code": "ERR_PIPELINE_TIMEOUT",
                    "analysis.latency_seconds": latency,
                    "analysis.timeout_seconds": ANALYZE_TIMEOUT_SECONDS,
                },
            )
            return analysis_pb2.AnalyzeResponse(
                success=False,
                error_message=(
                    f"ERR_PIPELINE_TIMEOUT: Analysis exceeded the "
                    f"{ANALYZE_TIMEOUT_SECONDS}s time limit."
                ),
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
        except Exception:  # noqa: BLE001
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

        input_tokens = usage["input_tokens"]
        output_tokens = usage["output_tokens"]
        total_tokens = usage["total_tokens"]
        llm_queue_wait = usage["queue_wait_seconds"]
        llm_retries = usage["retries"]

        latency = time.perf_counter() - start_time
        logger.info(
            "gRPC Analyze request completed successfully",
            extra={
                "analysis.success": True,
                "analysis.latency_seconds": latency,
                "analysis.provider": llm_provider,
                "analysis.model": llm_model,
                "analysis.input_tokens": input_tokens,
                "analysis.output_tokens": output_tokens,
                "analysis.total_tokens": total_tokens,
                "analysis.pdf_queue_wait_seconds": pdf_queue_wait_seconds,
                "analysis.pdf_duration_seconds": pdf_duration_seconds,
                "analysis.llm_queue_wait_seconds": llm_queue_wait,
                "analysis.llm_retries": llm_retries,
            },
        )
        return analysis_pb2.AnalyzeResponse(
            success=True,
            result_json=final_analysis.model_dump_json(
                by_alias=True,
                exclude_none=True,
            ),
            model=llm_model,
            input_tokens=input_tokens or 0,
            output_tokens=output_tokens or 0,
            total_tokens=total_tokens or 0,
        )
