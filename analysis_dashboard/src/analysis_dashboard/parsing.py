from __future__ import annotations

import json
from typing import Any, cast

from streamlit.runtime.uploaded_file_manager import UploadedFile

from analysis_dashboard.config import (
    MAX_FILE_SIZE_BYTES,
    VALID_MATCH_STRENGTHS,
)

ERROR_MESSAGES = {
    "ERR_INFRA_CONNECTION_DROP": "We couldn't complete the analysis right now. Please try again.",
    "ERR_INFRA_TIMEOUT": "The analysis took too long to finish. Please try again.",
    "ERR_INFRA_UPSTREAM_OUTAGE": "We couldn't complete the analysis right now. Please try again shortly.",
    "ERR_RATE_LIMIT_EXCEEDED": "Analysis is busy right now. Please wait a moment and try again.",
    "ERR_AUTH_KEY_INVALID": "We couldn't complete the analysis right now. Please contact support.",
    "ERR_MODEL_DEPRECATED_OR_NOT_FOUND": "We couldn't complete the analysis right now. Please contact support.",
    "ERR_PDF_PARSE": "We couldn't read one of the uploaded PDFs. Please upload another PDF and try again.",
    "ERR_CONTEXT_WINDOW_OVERFLOW": "The uploaded documents are too long. Please upload shorter PDFs and try again.",
    "ERR_INPUT_EXTRACTION_ARTIFACT": "We couldn't read the content of one of the PDFs. Please upload another PDF and try again.",
    "ERR_BAD_REQUEST": "Please check both uploaded PDFs and try again.",
    "ERR_SCHEMA_DRIFT_VALIDATION": "We couldn't complete the analysis right now. Please try again shortly.",
    "ERR_SECURITY_CONTENT_POLICY": "We couldn't process these documents. Please review them and try again.",
    "ERR_ASSEMBLY_MISMATCH": "We couldn't complete the analysis right now. Please try again or contact support.",
    "ERR_PIPELINE_TIMEOUT": "Analysis took too long to complete. Please try again.",
    "ERR_PROVIDER_API_FAILURE": "We couldn't complete the analysis right now. Please try again shortly.",
    "ERR_PIPELINE_INTERNAL": "We couldn't complete the analysis right now. Please try again or contact support.",
    "ERR_INTERNAL": "We couldn't complete the analysis right now. Please try again or contact support.",
}


def format_analysis_error(error_message: object) -> tuple[str, str]:
    """Convert service or transport errors into concise user-facing text."""
    raw_message = safe_str(error_message, "The analysis service returned an unknown error.")

    if raw_message.startswith("Could not communicate with the analysis service:"):
        return (
            "GRPC_ERROR",
            "We couldn't connect to the analysis service. Please try again shortly.",
        )

    if ":" in raw_message:
        code, _details = raw_message.split(":", 1)
        code = code.strip()
        if code in ERROR_MESSAGES:
            return code, ERROR_MESSAGES[code]

    return "UNKNOWN_ERROR", "We couldn't complete the analysis right now. Please try again."


def safe_str(value: object, default: str = "") -> str:
    if value is None:
        return default
    return str(value).strip() or default


def clamp_int(value: object, default: int = 0, minimum: int = 0, maximum: int = 100) -> int:
    try:
        if isinstance(value, bool):
            parsed = int(value)
        elif isinstance(value, int):
            parsed = value
        elif isinstance(value, str):
            parsed = int(value.strip())
        else:
            parsed = default
    except ValueError:
        parsed = default
    return max(minimum, min(maximum, parsed))


def coerce_string_list(value: object) -> list[str]:
    if value is None:
        return []
    if isinstance(value, str):
        item = value.strip()
        return [item] if item else []
    if not isinstance(value, list):
        return []

    result: list[str] = []
    for item in value:
        text = safe_str(item)
        if text:
            result.append(text)
    return result


def normalize_evidence_list(value: object) -> list[dict[str, Any]]:
    if isinstance(value, dict):
        items: list[object] = [value]
    elif isinstance(value, list):
        items = value
    else:
        return []

    return [cast(dict[str, Any], item) for item in items if isinstance(item, dict)]


def normalize_strength(value: object) -> str:
    strength = safe_str(value, "none").lower()
    return strength if strength in VALID_MATCH_STRENGTHS else "none"


def parse_analysis_result(result_json: str) -> dict[str, Any]:
    if not result_json.strip():
        raise ValueError("The analysis service returned an empty result.")

    try:
        parsed = json.loads(result_json)
    except json.JSONDecodeError as exc:
        raise ValueError("The analysis service returned invalid JSON.") from exc

    if not isinstance(parsed, dict):
        raise TypeError("The analysis service returned an unexpected result format.")

    data = cast(dict[str, Any], parsed)

    complete_analysis = data.get("completeAnalysis")
    if not isinstance(complete_analysis, str):
        raise TypeError("The analysis response is missing a valid executive summary.")

    sections_raw = data.get("sections")
    if not isinstance(sections_raw, list):
        raise TypeError("The analysis response contains invalid section data.")

    seen_ids: set[str] = set()

    for section in sections_raw:
        if not isinstance(section, dict):
            raise TypeError("The analysis response contains a malformed section entry.")

        section_id = safe_str(section.get("id"))
        if section_id:
            if section_id in seen_ids:
                raise ValueError(
                    f"Duplicate section id returned by analysis service: {section_id}"
                )
            seen_ids.add(section_id)

        requirements_raw = section.get("requirements")
        if requirements_raw is None:
            continue
        if not isinstance(requirements_raw, list):
            raise TypeError("The analysis response contains invalid requirement data.")

        for requirement in requirements_raw:
            if not isinstance(requirement, dict):
                raise TypeError(
                    "The analysis response contains a malformed requirement entry."
                )

    return data


def validate_pdf_upload(uploaded_file: UploadedFile | None, label: str) -> None:
    if uploaded_file is None:
        raise ValueError(f"{label} PDF was not uploaded.")

    if uploaded_file.type != "application/pdf":
        raise ValueError(f"{label} must be a PDF file.")

    if uploaded_file.size is None or uploaded_file.size <= 0:
        raise ValueError(f"{label} PDF is empty.")

    if uploaded_file.size > MAX_FILE_SIZE_BYTES:
        raise ValueError(f"{label} is larger than the 10 MB upload limit.")


def build_requirement_metrics(sections: list[dict[str, Any]]) -> dict[str, int]:
    metrics = {
        "total": 0,
        "matched": 0,
        "strong": 0,
        "partial": 0,
        "weak": 0,
        "none": 0,
        "jd_evidence": 0,
        "resume_evidence": 0,
    }

    for section in sections:
        requirements = section.get("requirements", [])
        if not isinstance(requirements, list):
            continue

        for requirement in requirements:
            if not isinstance(requirement, dict):
                continue

            metrics["total"] += 1
            strength = normalize_strength(requirement.get("matchStrength"))
            if bool(requirement.get("matched", strength in {"strong", "partial"})):
                metrics["matched"] += 1

            metrics[strength] += 1

            jd_evidence = normalize_evidence_list(requirement.get("jdEvidence"))
            resume_evidence = normalize_evidence_list(requirement.get("resumeEvidence"))
            if jd_evidence:
                metrics["jd_evidence"] += 1
            if resume_evidence:
                metrics["resume_evidence"] += 1

    return metrics
