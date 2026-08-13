import json
import os
from pathlib import Path

import grpc
import pytest
from app.pb import analysis_pb2, analysis_pb2_grpc
from dotenv import load_dotenv

pytestmark = pytest.mark.e2e


FIXTURES_DIR = Path(__file__).parent / "fixtures"
JD_PDF = FIXTURES_DIR / "sample_jd.pdf"
RESUME_PDF = FIXTURES_DIR / "sample_resume.pdf"

load_dotenv(Path(__file__).parents[2] / ".env")

HOST = os.getenv("E2E_GRPC_HOST", "127.0.0.1")
PORT = os.getenv("E2E_GRPC_PORT", os.getenv("PORT", "50051"))
TARGET = f"{HOST}:{PORT}"

def _require_e2e_configuration() -> None:
    required = ("LLM_API_KEY", "LLM_MODEL")

    missing = [
        variable
        for variable in required
        if not os.getenv(variable)
    ]

    if missing:
        pytest.fail(
            "Real E2E test requires: "
            + ", ".join(missing)
        )

    if not JD_PDF.is_file():
        pytest.fail(f"JD fixture not found: {JD_PDF}")

    if not RESUME_PDF.is_file():
        pytest.fail(f"Resume fixture not found: {RESUME_PDF}")


@pytest.mark.asyncio
async def test_real_analysis_end_to_end() -> None:
    _require_e2e_configuration()

    jd_pdf = JD_PDF.read_bytes()
    resume_pdf = RESUME_PDF.read_bytes()

    request_id = "e2e-real-analysis-001"

    async with grpc.aio.insecure_channel(TARGET) as channel:
        stub = analysis_pb2_grpc.AnalysisEngineStub(channel)

        response = await stub.Analyze(
            analysis_pb2.AnalyzeRequest(
                resume_pdf=resume_pdf,
                job_description_pdf=jd_pdf,
                request_id=request_id,
            )
        )

    assert response.success is True, response.error_message
    assert response.error_message == ""

    assert response.model
    assert response.input_tokens > 0
    assert response.output_tokens > 0
    assert response.total_tokens == (
        response.input_tokens + response.output_tokens
    )

    assert response.result_json

    result = json.loads(response.result_json)

    assert isinstance(result, dict)
    assert "completeAnalysis" in result
    assert isinstance(result["completeAnalysis"], str)
    assert result["completeAnalysis"].strip()

    assert "sections" in result
    assert isinstance(result["sections"], list)

    sections = result["sections"]
    section_ids = {section["id"] for section in sections}

    assert section_ids == {
        "skills",
        "experience",
        "education",
        "project",
    }

    for section in sections:
        assert isinstance(section["label"], str)
        assert 0 <= section["score"] <= 100
        assert isinstance(section["review"], str)
        assert isinstance(section["requirements"], list)

        for requirement in section["requirements"]:
            assert requirement["id"]
            assert requirement["requirement"]

            assert requirement["jdEvidence"]["source"] == "jd"
            assert requirement["jdEvidence"]["text"]

            assert requirement["matchStrength"] in {
                "strong",
                "partial",
                "weak",
                "none",
            }

            assert isinstance(requirement["resumeEvidence"], list)

            for evidence in requirement["resumeEvidence"]:
                assert evidence["source"] == "resume"
                assert evidence["text"]

            assert isinstance(requirement["matched"], bool)
