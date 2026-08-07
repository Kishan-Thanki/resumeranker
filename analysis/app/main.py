import asyncio
import os

import grpc
import uvloop

from app.logger import logger
from app.pb import analysis_pb2_grpc
from app.servicer import AnalysisEngineServicer

PORT = os.environ["PORT"]
GRPC_MAX_CONCURRENT_STREAMS = int(os.environ["GRPC_MAX_CONCURRENT_STREAMS"])
GRPC_KEEPALIVE_TIME_MS = int(os.environ["GRPC_KEEPALIVE_TIME_MS"])
GRPC_KEEPALIVE_TIMEOUT_MS = int(os.environ["GRPC_KEEPALIVE_TIMEOUT_MS"])


async def serve() -> None:
    """
    Creates, configures, and runs the asynchronous gRPC server.

    Configuration is loaded from environment variables and validated eagerly at
    startup. The server remains running until termination and performs a
    graceful shutdown, allowing in-flight RPCs to complete.
    """

    server = grpc.aio.server(
        options=(
            ("grpc.max_concurrent_streams", GRPC_MAX_CONCURRENT_STREAMS),
            ("grpc.keepalive_time_ms", GRPC_KEEPALIVE_TIME_MS),
            ("grpc.keepalive_timeout_ms", GRPC_KEEPALIVE_TIMEOUT_MS),
        ),
        compression=grpc.Compression.Gzip,
    )

    analysis_pb2_grpc.add_AnalysisEngineServicer_to_server(
        AnalysisEngineServicer(),
        server,
    )

    bound_port = server.add_insecure_port(f"[::]:{PORT}")
    if bound_port == 0:
        raise RuntimeError(f"Failed to bind gRPC server to port {PORT}")

    logger.info(
        "Starting gRPC server",
        extra={
            "port": bound_port,
            "compression": "gzip",
            "max_concurrent_streams": GRPC_MAX_CONCURRENT_STREAMS,
        },
    )

    try:
        await server.start()
        await server.wait_for_termination()

    finally:
        logger.info("Stopping gRPC server...")

    try:
        await server.stop(grace=10)
    except asyncio.CancelledError:
        logger.debug("Shutdown interrupted during server.stop()")


def main() -> None:
    """Application entry point."""

    uvloop.install()

    try:
        asyncio.run(serve())

    except KeyboardInterrupt:
        logger.info("Shutdown requested by user.")

    except Exception:
        logger.exception("Failed to start gRPC server")
        raise


if __name__ == "__main__":
    main()
