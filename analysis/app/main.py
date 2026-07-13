import asyncio
import logging
import grpc
import os

from app.pb import analysis_pb2
from app.pb import analysis_pb2_grpc

from app.services.pdf_service import extract_text_from_pdf_bytes
from app.services.llm_service import analyze_fit

MAX_CONCURRENT_PDFS = int(os.environ.get("MAX_CONCURRENT_PDF_PARSERS", 4))
pdf_bouncer = asyncio.Semaphore(MAX_CONCURRENT_PDFS)

async def _parse_pdf_safely(pdf_bytes: bytes) -> str:
    """Throttles PDF parsing to prevent OOM memory crashes."""
    async with pdf_bouncer:
        return await asyncio.to_thread(extract_text_from_pdf_bytes, pdf_bytes)

class AnalysisEngineServicer(analysis_pb2_grpc.AnalysisEngineServicer):
    async def Analyze(self, request, context):
        try:
            jd_text, resume_text = await asyncio.gather(
                _parse_pdf_safely(request.job_description_pdf),
                _parse_pdf_safely(request.resume_pdf)
            )
        except Exception as e:
            logging.error(f"Failed to process PDF: {e}")
            return analysis_pb2.AnalyzeResponse(
                success=False,
                error_message=f"Invalid PDF input: {str(e)}"
            )

        try:
            final_analysis, usage = await analyze_fit(jd_text, resume_text)
        except Exception as e:
            logging.error(f"LLM processing failed: {e}")
            return analysis_pb2.AnalyzeResponse(
                success=False,
                error_message=f"LLM Engine failed: {str(e)}"
            )

        return analysis_pb2.AnalyzeResponse(
            success=True,
            result_json=final_analysis.model_dump_json(by_alias=True, exclude_none=True),
            model=os.environ["LLM_MODEL"],
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
    logging.info(f"gRPC Server starting on port {port}")
    
    await server.start()
    await server.wait_for_termination()

if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    
    import uvloop
    uvloop.install()
    
    asyncio.run(serve())
