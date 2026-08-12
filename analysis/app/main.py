import asyncio
import os
import signal

import grpc
import uvloop

from app.logger import logger
from app.pb import analysis_pb2_grpc
from app.servicer import AnalysisEngineServicer, shutdown_pdf_executor


def load_config() -> dict[str, int]:
    """Safely parse and validate environment variables at startup."""
    try:
        return {
            "port": int(os.getenv("PORT", "50051")),
            "max_concurrent_streams": int(
                os.getenv("GRPC_MAX_CONCURRENT_STREAMS", "100")
            ),
            "keepalive_time_ms": int(os.getenv("GRPC_KEEPALIVE_TIME_MS", "60000")),
            "keepalive_timeout_ms": int(
                os.getenv("GRPC_KEEPALIVE_TIMEOUT_MS", "20000")
            ),
            "max_message_length": int(
                os.getenv("GRPC_MAX_MESSAGE_LENGTH", str(16 * 1024 * 1024))
            ),
        }
    except ValueError as e:
        logger.error("Invalid numeric value in gRPC environment variables", exc_info=e)
        raise RuntimeError("Failed to parse gRPC configuration") from e


async def serve() -> None:
    """Creates, configures, and runs the asynchronous gRPC server."""
    config = load_config()

    server = grpc.aio.server(
        options=(
            ("grpc.max_concurrent_streams", config["max_concurrent_streams"]),
            ("grpc.keepalive_time_ms", config["keepalive_time_ms"]),
            ("grpc.keepalive_timeout_ms", config["keepalive_timeout_ms"]),
            ("grpc.max_receive_message_length", config["max_message_length"]),
            ("grpc.max_send_message_length", config["max_message_length"]),
        ),
        compression=grpc.Compression.Gzip,
    )

    analysis_pb2_grpc.add_AnalysisEngineServicer_to_server(
        AnalysisEngineServicer(),
        server,
    )

    port = config["port"]
    bound_port = server.add_insecure_port(f"[::]:{port}")
    if bound_port == 0:
        logger.error(
            "Failed to bind gRPC server to port",
            extra={"port": port, "config": config},
        )
        raise RuntimeError(f"Failed to bind gRPC server to port {port}")

    logger.info(
        "Starting gRPC server",
        extra={"port": bound_port, "compression": "gzip", **config},
    )

    await server.start()

    stop_event = asyncio.Event()

    def signal_handler() -> None:
        logger.info("Received termination signal")
        stop_event.set()

    loop = asyncio.get_running_loop()
    for sig in (signal.SIGINT, signal.SIGTERM):
        try:
            loop.add_signal_handler(sig, signal_handler)
        except NotImplementedError:
            pass

    try:
        server_wait_task = asyncio.create_task(server.wait_for_termination())
        stop_event_task = asyncio.create_task(stop_event.wait())

        _, pending = await asyncio.wait(
            [server_wait_task, stop_event_task],
            return_when=asyncio.FIRST_COMPLETED,
        )

        for task in pending:
            task.cancel()

        for task in pending:
            try:
                await task
            except asyncio.CancelledError:
                pass

    finally:
        logger.info("Stopping gRPC server with grace period...")
        try:
            await asyncio.shield(server.stop(grace=10))
        except asyncio.CancelledError:
            logger.debug("Server stop interrupted during task cancellation.")
        except Exception:  # noqa: BLE001
            logger.exception("Error stopping gRPC server")

        logger.info("Shutting down PDF executor pool...")
        try:
            await asyncio.shield(shutdown_pdf_executor())
        except asyncio.CancelledError:
            logger.debug("PDF executor shutdown interrupted during task cancellation.")
        except Exception:  # noqa: BLE001
            logger.exception("Error shutting down PDF executor pool")

        logger.info("Server cleanup complete.")


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
