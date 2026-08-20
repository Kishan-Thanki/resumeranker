from __future__ import annotations

import json

import pytest
from analysis_dashboard.config import MAX_FILE_SIZE_BYTES

from analysis_dashboard import parsing

from .conftest import FakeUploadedFile


def test_parse_analysis_result_normalizes_and_sorts_sections(
    sample_analysis_result_json: str,
) -> None:
    parsed = parsing.parse_analysis_result(sample_analysis_result_json)

    assert parsed["completeAnalysis"] == "Strong overall fit for the role."
    assert [section["id"] for section in parsed["sections"]] == [
        "project",
        "skills",
        "experience",
        "education",
    ]

    skills = parsed["sections"][1]
    requirement = skills["requirements"][0]

    assert requirement["matchStrength"] == "strong"
    assert requirement["supportingClaimIds"] == ["claim-1"]
    assert isinstance(requirement["jdEvidence"], list)
    assert isinstance(requirement["resumeEvidence"], dict)


def test_parse_analysis_result_preserves_service_payload_values() -> None:
    payload = {
        "completeAnalysis": "  Summary exactly as returned  ",
        "sections": [
            {
                "id": "custom",
                "label": "Custom label",
                "score": "not-normalized",
                "review": None,
                "requirements": [],
            }
        ],
    }

    parsed = parsing.parse_analysis_result(json.dumps(payload))

    assert parsed == payload


def test_parse_analysis_result_rejects_duplicate_sections() -> None:
    payload = {
        "completeAnalysis": "Summary",
        "sections": [
            {"id": "skills", "label": "Skills", "score": 80, "review": "ok"},
            {"id": "skills", "label": "Skills", "score": 70, "review": "dup"},
        ],
    }

    with pytest.raises(ValueError, match="Duplicate section id"):
        parsing.parse_analysis_result(json.dumps(payload))


def test_parse_analysis_result_rejects_invalid_json() -> None:
    with pytest.raises(ValueError, match="invalid JSON"):
        parsing.parse_analysis_result("{not-json}")


def test_build_requirement_metrics_counts_matches(
    sample_analysis_result_json: str,
) -> None:
    parsed = parsing.parse_analysis_result(sample_analysis_result_json)
    metrics = parsing.build_requirement_metrics(parsed["sections"])

    assert metrics == {
        "total": 4,
        "matched": 4,
        "strong": 2,
        "partial": 2,
        "weak": 0,
        "none": 0,
        "jd_evidence": 4,
        "resume_evidence": 4,
    }


def test_validate_pdf_upload_accepts_pdf() -> None:
    parsing.validate_pdf_upload(FakeUploadedFile(), "Resume")


def test_validate_pdf_upload_rejects_missing_file() -> None:
    with pytest.raises(ValueError, match="was not uploaded"):
        parsing.validate_pdf_upload(None, "Resume")


def test_validate_pdf_upload_rejects_wrong_type() -> None:
    upload = FakeUploadedFile(type="text/plain")

    with pytest.raises(ValueError, match="must be a PDF file"):
        parsing.validate_pdf_upload(upload, "Resume")


def test_validate_pdf_upload_rejects_oversize_file() -> None:
    upload = FakeUploadedFile(size=MAX_FILE_SIZE_BYTES + 1)

    with pytest.raises(ValueError, match="larger than the 10 MB upload limit"):
        parsing.validate_pdf_upload(upload, "Resume")


def test_format_analysis_error_hides_provider_details() -> None:
    code, message = parsing.format_analysis_error(
        "ERR_SCHEMA_DRIFT_VALIDATION: The AI service failed to produce valid structured output: GeminiException 503"
    )

    assert code == "ERR_SCHEMA_DRIFT_VALIDATION"
    assert message == (
        "We couldn't complete the analysis right now. Please try again shortly."
    )


def test_format_analysis_error_handles_transport_failure() -> None:
    code, message = parsing.format_analysis_error(
        "Could not communicate with the analysis service: [UNAVAILABLE] connection refused"
    )

    assert code == "GRPC_ERROR"
    assert message == (
        "We couldn't connect to the analysis service. Please try again shortly."
    )


def test_format_analysis_error_handles_unknown_code() -> None:
    code, message = parsing.format_analysis_error("ERR_NEW_FAILURE: provider detail")

    assert code == "UNKNOWN_ERROR"
    assert message == "We couldn't complete the analysis right now. Please try again."
