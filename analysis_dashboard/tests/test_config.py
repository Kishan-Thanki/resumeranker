from __future__ import annotations

from analysis_dashboard import config


def test_load_default_connection_uses_defaults(monkeypatch) -> None:
    monkeypatch.delenv("ANALYSIS_GRPC_HOST", raising=False)
    monkeypatch.delenv("ANALYSIS_GRPC_PORT", raising=False)
    monkeypatch.delenv("ANALYSIS_GRPC_TIMEOUT_SECONDS", raising=False)

    connection = config.load_default_connection()

    assert connection.host == "localhost"
    assert connection.port == 50051
    assert connection.timeout_seconds == 180.0
    assert connection.address == "localhost:50051"


def test_load_default_connection_uses_environment(monkeypatch) -> None:
    monkeypatch.setenv("ANALYSIS_GRPC_HOST", "analysis.internal")
    monkeypatch.setenv("ANALYSIS_GRPC_PORT", "60051")
    monkeypatch.setenv("ANALYSIS_GRPC_TIMEOUT_SECONDS", "42.5")

    connection = config.load_default_connection()

    assert connection.host == "analysis.internal"
    assert connection.port == 60051
    assert connection.timeout_seconds == 42.5
