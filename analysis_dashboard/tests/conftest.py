from __future__ import annotations

import json
from dataclasses import dataclass, field
from typing import Any

import pytest


@dataclass
class FakeUploadedFile:
    name: str = "sample.pdf"
    type: str = "application/pdf"
    size: int = 128
    data: bytes = b"%PDF-1.4 fake pdf bytes"

    def getvalue(self) -> bytes:
        return self.data


class _NullContext:
    def __init__(self, owner: "FakeStreamlit") -> None:
        self.owner = owner

    def __enter__(self) -> "_NullContext":
        return self

    def __exit__(self, exc_type, exc, tb) -> None:
        return None

    def write(self, message: str) -> None:
        self.owner.calls.append(("write", message))

    def update(self, **kwargs: Any) -> None:
        self.owner.calls.append(("update", kwargs))


class FakeStreamlit:
    def __init__(self) -> None:
        self.session_state: dict[str, Any] = {}
        self.calls: list[tuple[str, Any]] = []
        self.rerun_called = False
        self.status_updates: list[tuple[str, Any]] = []

    def markdown(self, message: str, **kwargs: Any) -> None:
        self.calls.append(("markdown", message))

    def subheader(self, message: str) -> None:
        self.calls.append(("subheader", message))

    def info(self, message: str) -> None:
        self.calls.append(("info", message))

    def warning(self, message: str) -> None:
        self.calls.append(("warning", message))

    def error(self, message: str) -> None:
        self.calls.append(("error", message))

    def success(self, message: str) -> None:
        self.calls.append(("success", message))

    def caption(self, message: str) -> None:
        self.calls.append(("caption", message))

    def write(self, message: str) -> None:
        self.calls.append(("write", message))

    def progress(self, value: float) -> None:
        self.calls.append(("progress", value))

    def download_button(self, **kwargs: Any) -> None:
        self.calls.append(("download_button", kwargs))

    def json(self, value: Any) -> None:
        self.calls.append(("json", value))

    def rerun(self) -> None:
        self.rerun_called = True
        raise RuntimeError("rerun called")

    def columns(self, spec: Any) -> list[_NullContext]:
        if isinstance(spec, int):
            count = spec
        else:
            count = len(spec)
        return [_NullContext(self) for _ in range(count)]

    def tabs(self, labels: list[str]) -> list[_NullContext]:
        self.calls.append(("tabs", labels))
        return [_NullContext(self) for _ in labels]

    def container(self, border: bool = False) -> _NullContext:
        self.calls.append(("container", border))
        return _NullContext(self)

    def expander(self, label: str, expanded: bool = False) -> _NullContext:
        self.calls.append(("expander", {"label": label, "expanded": expanded}))
        return _NullContext(self)

    def status(self, label: str, expanded: bool = False) -> _NullContext:
        self.status_updates.append((label, expanded))
        return _NullContext(self)


@pytest.fixture
def sample_analysis_result_dict() -> dict[str, Any]:
    return {
        "completeAnalysis": "Strong overall fit for the role.",
        "matchedSkills": ["Python", "FastAPI", "PostgreSQL"],
        "missingCriticalSkills": ["Kafka"],
        "sections": [
            {
                "id": "project",
                "label": "Projects",
                "score": 74,
                "review": "Good project breadth with some room for deeper system design.",
                "requirements": [
                    {
                        "id": "req-project-1",
                        "requirement": "Show production project ownership",
                        "jdEvidence": {
                            "source": "jd",
                            "text": "Built backend systems in production.",
                            "location": "Responsibilities",
                        },
                        "matchStrength": "partial",
                        "resumeEvidence": [
                            {
                                "source": "resume",
                                "text": "Built production APIs with FastAPI and PostgreSQL.",
                                "location": "Experience",
                            }
                        ],
                        "supporting_claim_ids": ["claim-1"],
                        "note": "Projects are relevant but could be deeper.",
                    }
                ],
            },
            {
                "id": "skills",
                "label": "Skills",
                "score": 91,
                "review": "Very strong match on the required stack.",
                "requirements": [
                    {
                        "id": "req-skills-1",
                        "requirement": "Python and FastAPI experience",
                        "jdEvidence": [
                            {
                                "source": "jd",
                                "text": "Strong experience with FastAPI and PostgreSQL is required.",
                                "location": "Requirements",
                            }
                        ],
                        "matchStrength": "strong",
                        "resumeEvidence": {
                            "source": "resume",
                            "text": "Built production APIs with FastAPI and PostgreSQL.",
                            "location": "Experience",
                        },
                        "supportingClaimIds": ["claim-1"],
                    }
                ],
            },
            {
                "id": "experience",
                "label": "Experience",
                "score": 88,
                "review": "Experience level exceeds the requested minimum.",
                "requirements": [
                    {
                        "id": "req-exp-1",
                        "requirement": "5+ years of Python experience",
                        "jdEvidence": {
                            "source": "jd",
                            "text": "5+ years of Python experience.",
                            "location": "Requirements",
                        },
                        "matchStrength": "strong",
                        "resumeEvidence": [
                            {
                                "source": "resume",
                                "text": "Senior Backend Engineer with 6 years of Python development.",
                                "location": "Summary",
                            }
                        ],
                        "supportingClaimIds": ["claim-2"],
                    }
                ],
            },
            {
                "id": "education",
                "label": "Education",
                "score": 62,
                "review": "Education aligns well with the role requirements.",
                "requirements": [
                    {
                        "id": "req-edu-1",
                        "requirement": "Bachelor's degree in Computer Science",
                        "jdEvidence": {
                            "source": "jd",
                            "text": "Bachelor's degree in Computer Science is preferred.",
                            "location": "Requirements",
                        },
                        "matchStrength": "partial",
                        "resumeEvidence": [
                            {
                                "source": "resume",
                                "text": "Bachelor of Technology in Computer Science.",
                                "location": "Education",
                            }
                        ],
                        "supportingClaimIds": ["claim-3"],
                    }
                ],
            },
        ],
    }


@pytest.fixture
def sample_analysis_result_json(sample_analysis_result_dict: dict[str, Any]) -> str:
    return json.dumps(sample_analysis_result_dict)
