import asyncio
import os
import grpc
import sys

pb_dir = os.path.join(os.path.dirname(__file__), 'pb')
sys.path.insert(0, pb_dir)

from app.pb import analysis_pb2_grpc
from app.logger import logger
from app.servicer import AnalysisEngineServicer

async def serve():
    """
    Initializes and starts the asynchronous gRPC server.
    
    Configures critical server options such as max concurrent streams and 
    keepalive timeouts to ensure high availability and resilience.
    Binds the AnalysisEngineServicer to the server and listens on the specified PORT.
    """
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
