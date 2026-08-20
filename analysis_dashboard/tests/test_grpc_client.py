from __future__ import annotations

from types import SimpleNamespace

import grpc
import pytest

from analysis_dashboard import grpc_client
from analysis_dashboard.config import ConnectionConfig

from .conftest import FakeUploadedFile


class _FakeChannel:
    def __enter__(self) -> "_FakeChannel":
        return self

    def __exit__(self, exc_type, exc, tb) -> None:
        return None


class _FakeRpcError(grpc.RpcError):
    def __init__(self, code: grpc.StatusCode, details: str) -> None:
        super().__init__()
        self._code = code
        self._details = details

    def code(self) -> grpc.StatusCode:
        return self._code

    def details(self) -> str:
        return self._details


def test_call_analysis_service_success(monkeypatch) -> None:
    response = SimpleNamespace(
        success=True,
        error_message="",
        result_json='{"completeAnalysis":"ok","sections":[]}',
        model="openai/gpt-test",
        input_tokens=10,
        output_tokens=5,
        total_tokens=15,
    )

    captured = {}

    class FakeStub:
        def __init__(self, channel) -> None:
            captured["channel"] = channel

        def Analyze(self, request, timeout):
            captured["request"] = request
            captured["timeout"] = timeout
            return response

    monkeypatch.setattr(grpc_client.grpc, "insecure_channel", lambda address: _FakeChannel())
    monkeypatch.setattr(grpc_client.analysis_pb2_grpc, "AnalysisEngineStub", FakeStub)

    result, error, request_id, elapsed_seconds = grpc_client.call_analysis_service(
        ConnectionConfig(host="localhost", port=50051, timeout_seconds=12.5),
        FakeUploadedFile(data=b"jd-bytes"),
        FakeUploadedFile(name="resume.pdf", data=b"resume-bytes"),
    )

    assert error is None
    assert result is response
    assert request_id.startswith("ui-")
    assert elapsed_seconds >= 0
    assert captured["timeout"] == 12.5
    assert captured["request"].job_description_pdf == b"jd-bytes"
    assert captured["request"].resume_pdf == b"resume-bytes"
    assert captured["request"].request_id == request_id


def test_call_analysis_service_formats_rpc_errors(monkeypatch) -> None:
    class FakeStub:
        def __init__(self, channel) -> None:
            pass

        def Analyze(self, request, timeout):
            raise _FakeRpcError(grpc.StatusCode.UNAVAILABLE, "service unavailable")

    monkeypatch.setattr(grpc_client.grpc, "insecure_channel", lambda address: _FakeChannel())
    monkeypatch.setattr(grpc_client.analysis_pb2_grpc, "AnalysisEngineStub", FakeStub)

    result, error, request_id, elapsed_seconds = grpc_client.call_analysis_service(
        ConnectionConfig(host="localhost", port=50051, timeout_seconds=3.0),
        FakeUploadedFile(),
        FakeUploadedFile(name="resume.pdf"),
    )

    assert result is None
    assert error == "Could not communicate with the analysis service: [UNAVAILABLE] service unavailable"
    assert request_id.startswith("ui-")
    assert elapsed_seconds >= 0
