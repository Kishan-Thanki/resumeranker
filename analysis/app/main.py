import asyncio
import os
import grpc
import time

import sys

# gRPC Python 3 import bug workaround: Add the pb directory to the path 
# so the auto-generated analysis_pb2_grpc.py can resolve 'import analysis_pb2'
pb_dir = os.path.join(os.path.dirname(__file__), 'pb')
sys.path.insert(0, pb_dir)

from app.pb import analysis_pb2
from app.pb import analysis_pb2_grpc

from app.services.pdf_service import extract_text_from_pdf_bytes
from app.services.llm_service import analyze_fit, AnalysisEngineError
from app.logger import logger, request_id_var

MAX_CONCURRENT_PDFS = int(os.environ.get("MAX_CONCURRENT_PDF_PARSERS", 4))
pdf_bouncer = asyncio.Semaphore(MAX_CONCURRENT_PDFS)

async def _parse_pdf_safely(pdf_bytes: bytes, doc_name: str) -> str:
    """Throttles PDF parsing to prevent OOM memory crashes."""
    async with pdf_bouncer:
        try:
            return await asyncio.to_thread(extract_text_from_pdf_bytes, pdf_bytes)
        except Exception as e:
            logger.exception(f"Failed to parse PDF document: {doc_name}")
            raise AnalysisEngineError("ERR_PDF_PARSE", f"Failed to parse {doc_name}. Please ensure it is a valid text-based PDF.")

class AnalysisEngineServicer(analysis_pb2_grpc.AnalysisEngineServicer):
    async def Analyze(self, request, context):
        request_id_var.set(request.request_id)
        start_time = time.time()
        logger.info("gRPC Analyze request started")
        
        try:
            jd_text, resume_text = await asyncio.gather(
                _parse_pdf_safely(request.job_description_pdf, "Job Description"),
                _parse_pdf_safely(request.resume_pdf, "Resume")
            )
        except AnalysisEngineError as e:
            latency = time.time() - start_time
            logger.warning(f"gRPC Analyze request failed during PDF parsing: {e.code}", extra={"analysis.error_code": e.code, "analysis.latency_seconds": latency})
            return analysis_pb2.AnalyzeResponse(
                success=False,
                error_message=f"{e.code}: {e.message}"
            )
        except Exception as e:
            latency = time.time() - start_time
            logger.exception("Unexpected error during PDF parsing phase", extra={"analysis.error_code": "ERR_INTERNAL", "analysis.latency_seconds": latency})
            return analysis_pb2.AnalyzeResponse(
                success=False,
                error_message="ERR_INTERNAL: An unexpected error occurred while processing the documents."
            )

        try:
            final_analysis, usage = await analyze_fit(jd_text, resume_text)
        except AnalysisEngineError as e:
            latency = time.time() - start_time
            logger.warning(f"gRPC Analyze request failed during AI processing: {e.code}", extra={"analysis.error_code": e.code, "analysis.latency_seconds": latency})
            return analysis_pb2.AnalyzeResponse(
                success=False,
                error_message=f"{e.code}: {e.message}"
            )
        except Exception as e:
            latency = time.time() - start_time
            logger.exception("Unexpected error during LLM processing phase", extra={"analysis.error_code": "ERR_INTERNAL", "analysis.latency_seconds": latency})
            return analysis_pb2.AnalyzeResponse(
                success=False,
                error_message="ERR_INTERNAL: An unexpected error occurred during AI analysis."
            )

        latency = time.time() - start_time
        model_name = os.environ.get("LLM_MODEL", "unknown")
        
        logger.info("gRPC Analyze request completed successfully", extra={
            "analysis.latency_seconds": latency,
            "analysis.model": model_name,
            "analysis.prompt_tokens": usage.get("prompt_tokens", 0),
            "analysis.completion_tokens": usage.get("completion_tokens", 0),
            "analysis.total_tokens": usage.get("total_tokens", 0)
        })

        return analysis_pb2.AnalyzeResponse(
            success=True,
            result_json=final_analysis.model_dump_json(by_alias=True, exclude_none=True),
            model=model_name,
            prompt_tokens=usage.get("prompt_tokens", 0),
            completion_tokens=usage.get("completion_tokens", 0),
            total_tokens=usage.get("total_tokens", 0)
        )

async def serve():
    port = os.environ.get("PORT", "8001")
    server = grpc.aio.server(
        options=(
            ('grpc.max_concurrent_streams', 100),
            ('grpc.keepalive_time_ms', 10000),
            ('grpc.keepalive_timeout_ms', 5000),
            ('grpc.default_compression_algorithm', 2),
            ('grpc.default_compression_level', 2),
        ),
        compression=grpc.Compression.Gzip
    )
    analysis_pb2_grpc.add_AnalysisEngineServicer_to_server(AnalysisEngineServicer(), server)
    
    server.add_insecure_port(f"[::]:{port}")
    logger.info(f"gRPC Server starting on port {port}")
    
    await server.start()
    await server.wait_for_termination()

if __name__ == "__main__":
    import uvloop
    uvloop.install()
    
    asyncio.run(serve())
