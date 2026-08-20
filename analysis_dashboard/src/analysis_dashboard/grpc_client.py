from __future__ import annotations

import time
import uuid

import grpc
from streamlit.runtime.uploaded_file_manager import UploadedFile

from analysis_dashboard.pb import analysis_pb2, analysis_pb2_grpc
from analysis_dashboard.config import ConnectionConfig


def call_analysis_service(
    config: ConnectionConfig,
    jd_file: UploadedFile,
    resume_file: UploadedFile,
) -> tuple[analysis_pb2.AnalyzeResponse | None, str | None, str, float]:
    request_id = f"ui-{uuid.uuid4().hex}"
    request = analysis_pb2.AnalyzeRequest(
        request_id=request_id,
        job_description_pdf=jd_file.getvalue(),
        resume_pdf=resume_file.getvalue(),
    )

    start = time.perf_counter()
    try:
        with grpc.insecure_channel(config.address) as channel:
            stub = analysis_pb2_grpc.AnalysisEngineStub(channel)
            response = stub.Analyze(request, timeout=config.timeout_seconds)
    except grpc.RpcError as exc:
        elapsed_seconds = time.perf_counter() - start
        return (
            None,
            "Could not communicate with the analysis service: "
            f"[{exc.code().name}] {exc.details()}",
            request_id,
            elapsed_seconds,
        )

    elapsed_seconds = time.perf_counter() - start
    return response, None, request_id, elapsed_seconds
