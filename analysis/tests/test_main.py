import asyncio
from types import SimpleNamespace
from typing import Any

import pytest
from app import main


class FakeServer:
    def __init__(self) -> None:
        self.started = False
        self.start_calls = 0
        self.stop_calls: list[int] = []
        self.wait_for_termination_calls = 0
        self.options = None
        self.compression = None
        self.registered = False
        self.registered_method_handlers: dict[str, dict[str, object]] = {}

    def add_insecure_port(self, address: str) -> int:
        self.bound_address = address
        return 50051

    def add_generic_rpc_handlers(self, handlers: tuple[object, ...]) -> None:
        self.registered = True
        self.rpc_handlers = handlers

    def add_registered_method_handlers(
        self,
        service_name: str,
        handlers: dict[str, object],
    ) -> None:
        self.registered_method_handlers = {
            service_name: handlers,
        }

    async def start(self) -> None:
        self.started = True
        self.start_calls += 1

    async def wait_for_termination(self) -> None:
        self.wait_for_termination_calls += 1
        await asyncio.Event().wait()

    async def stop(self, grace: int) -> None:
        self.stop_calls.append(grace)


class FakeGrpcAio:
    def __init__(self, fake_server: FakeServer) -> None:
        self.fake_server = fake_server
        self.server_kwargs: dict[str, Any] | None = None

    def server(self, **kwargs: Any) -> FakeServer:
        self.server_kwargs = kwargs
        return self.fake_server


class FakeGrpc:
    class Compression:
        Gzip = "gzip"

    def __init__(self, server: FakeServer) -> None:
        self.aio = FakeGrpcAio(server)


class TestLoadConfig:
    def test_loads_defaults(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        for variable in (
            "PORT",
            "GRPC_MAX_CONCURRENT_STREAMS",
            "GRPC_KEEPALIVE_TIME_MS",
            "GRPC_KEEPALIVE_TIMEOUT_MS",
            "GRPC_MAX_MESSAGE_LENGTH",
        ):
            monkeypatch.delenv(variable, raising=False)

        config = main.load_config()

        assert config == {
            "port": 50051,
            "max_concurrent_streams": 100,
            "keepalive_time_ms": 60000,
            "keepalive_timeout_ms": 20000,
            "max_message_length": 16 * 1024 * 1024,
        }

    def test_loads_environment_values(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setenv("PORT", "60000")
        monkeypatch.setenv("GRPC_MAX_CONCURRENT_STREAMS", "25")
        monkeypatch.setenv("GRPC_KEEPALIVE_TIME_MS", "30000")
        monkeypatch.setenv("GRPC_KEEPALIVE_TIMEOUT_MS", "5000")
        monkeypatch.setenv("GRPC_MAX_MESSAGE_LENGTH", "8388608")

        config = main.load_config()

        assert config == {
            "port": 60000,
            "max_concurrent_streams": 25,
            "keepalive_time_ms": 30000,
            "keepalive_timeout_ms": 5000,
            "max_message_length": 8388608,
        }

    @pytest.mark.parametrize(
        "variable",
        [
            "PORT",
            "GRPC_MAX_CONCURRENT_STREAMS",
            "GRPC_KEEPALIVE_TIME_MS",
            "GRPC_KEEPALIVE_TIMEOUT_MS",
            "GRPC_MAX_MESSAGE_LENGTH",
        ],
    )
    def test_rejects_invalid_numeric_environment_value(
        self,
        monkeypatch: pytest.MonkeyPatch,
        variable: str,
    ) -> None:
        monkeypatch.setenv(variable, "not-a-number")

        with pytest.raises(
            RuntimeError,
            match="Failed to parse gRPC configuration",
        ):
            main.load_config()


class TestServe:
    @pytest.mark.asyncio
    async def test_starts_server_and_shuts_down_gracefully(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        fake_server = FakeServer()
        fake_grpc = FakeGrpc(fake_server)

        async def fake_shutdown_pdf_executor() -> None:
            return None

        monkeypatch.setattr(
            main,
            "grpc",
            fake_grpc,
        )
        monkeypatch.setattr(
            main,
            "shutdown_pdf_executor",
            fake_shutdown_pdf_executor,
        )

        async def fake_stop_event_wait() -> None:
            await asyncio.sleep(0)

        class FakeEvent:
            def __init__(self) -> None:
                self.wait_calls = 0

            def set(self) -> None:
                pass

            async def wait(self) -> None:
                self.wait_calls += 1
                await fake_stop_event_wait()

        monkeypatch.setattr(
            main.asyncio,
            "Event",
            FakeEvent,
        )

        # Make the signal-driven event complete immediately.
        async def fake_wait_for_termination() -> None:
            await asyncio.sleep(0)

        fake_server.wait_for_termination = fake_wait_for_termination

        monkeypatch.setattr(
            main.asyncio,
            "get_running_loop",
            lambda: SimpleNamespace(
                add_signal_handler=lambda _sig, _handler: None,
            ),
        )

        await main.serve()

        assert fake_server.started is True
        assert fake_server.start_calls == 1
        assert fake_server.stop_calls == [10]
        assert fake_server.wait_for_termination_calls == 0

    @pytest.mark.asyncio
    async def test_server_configuration_is_passed_to_grpc(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        fake_server = FakeServer()
        fake_grpc = FakeGrpc(fake_server)

        monkeypatch.setattr(
            main,
            "grpc",
            fake_grpc,
        )

        async def fake_shutdown_pdf_executor() -> None:
            return None

        monkeypatch.setattr(
            main,
            "shutdown_pdf_executor",
            fake_shutdown_pdf_executor,
        )

        async def fake_wait_for_termination() -> None:
            await asyncio.sleep(0)

        fake_server.wait_for_termination = fake_wait_for_termination

        monkeypatch.setattr(
            main.asyncio,
            "get_running_loop",
            lambda: SimpleNamespace(
                add_signal_handler=lambda _sig, _handler: None,
            ),
        )

        await main.serve()

        assert fake_grpc.aio.server_kwargs is not None

        assert fake_grpc.aio.server_kwargs["compression"] == "gzip"

        options = fake_grpc.aio.server_kwargs["options"]

        assert (
            ("grpc.max_concurrent_streams", 100)
            in options
        )
        assert (
            ("grpc.keepalive_time_ms", 60000)
            in options
        )
        assert (
            ("grpc.keepalive_timeout_ms", 20000)
            in options
        )
        assert (
            ("grpc.max_receive_message_length", 16 * 1024 * 1024)
            in options
        )
        assert (
            ("grpc.max_send_message_length", 16 * 1024 * 1024)
            in options
        )

    @pytest.mark.asyncio
    async def test_fails_when_port_binding_returns_zero(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        class UnbindableServer(FakeServer):
            def add_insecure_port(self, address: str) -> int:
                self.bound_address = address
                return 0

        fake_server = UnbindableServer()
        fake_grpc = FakeGrpc(fake_server)

        monkeypatch.setattr(
            main,
            "grpc",
            fake_grpc,
        )

        with pytest.raises(
            RuntimeError,
            match="Failed to bind gRPC server to port 50051",
        ):
            await main.serve()

    @pytest.mark.asyncio
    async def test_server_start_failure_is_propagated(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        fake_server = FakeServer()
        fake_grpc = FakeGrpc(fake_server)

        async def failing_start() -> None:
            raise RuntimeError("startup failed")

        fake_server.start = failing_start

        monkeypatch.setattr(
            main,
            "grpc",
            fake_grpc,
        )

        monkeypatch.setattr(
            main.asyncio,
            "get_running_loop",
            lambda: SimpleNamespace(
                add_signal_handler=lambda _sig, _handler: None,
            ),
        )

        with pytest.raises(
            RuntimeError,
            match="startup failed",
        ):
            await main.serve()

        assert fake_server.stop_calls == []


class TestMain:
    def test_main_installs_uvloop(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        install_called = False
        run_called = False

        def fake_install() -> None:
            nonlocal install_called
            install_called = True

        async def fake_serve() -> None:
            return None

        def fake_run(coro):
            nonlocal run_called
            run_called = True
            coro.close()

        monkeypatch.setattr(
            main.uvloop,
            "install",
            fake_install,
        )
        monkeypatch.setattr(
            main,
            "serve",
            fake_serve,
        )
        monkeypatch.setattr(
            main.asyncio,
            "run",
            fake_run,
        )

        main.main()

        assert install_called is True
        assert run_called is True

    def test_main_logs_keyboard_interrupt(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        logged_messages: list[str] = []

        def fake_run(coro):
            coro.close()
            raise KeyboardInterrupt

        def fake_info(message: str, *args, **kwargs) -> None:
            logged_messages.append(message)

        monkeypatch.setattr(
            main.asyncio,
            "run",
            fake_run,
        )
        monkeypatch.setattr(
            main.logger,
            "info",
            fake_info,
        )

        main.main()

        assert "Shutdown requested by user." in logged_messages

    def test_main_reraises_unexpected_startup_exception(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        def fake_run(coro):
            coro.close()
            raise RuntimeError("startup exploded")

        logged_messages: list[str] = []

        def fake_exception(message: str, *args, **kwargs) -> None:
            logged_messages.append(message)

        monkeypatch.setattr(
            main.asyncio,
            "run",
            fake_run,
        )
        monkeypatch.setattr(
            main.logger,
            "exception",
            fake_exception,
        )

        with pytest.raises(
            RuntimeError,
            match="startup exploded",
        ):
            main.main()

        assert "Failed to start gRPC server" in logged_messages
