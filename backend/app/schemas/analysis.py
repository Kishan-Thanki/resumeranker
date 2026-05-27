"""Pydantic models that mirror the frontend's contract verbatim.

Source of truth: frontend/src/lib/types.ts

Python is snake_case internally; `serialization_alias` produces camelCase JSON
to match the frontend exactly. FastAPI route decorators must use
`response_model_by_alias=True` (set globally below in the routes module).
"""

from datetime import datetime
from typing import Literal

from pydantic import BaseModel, ConfigDict, Field

AnalysisStatus = Literal["queued", "processing", "completed", "failed"]
SectionId = Literal["skills", "experience", "education", "leadership"]
MatchStrength = Literal["strong", "partial", "weak", "none"]
EvidenceSource = Literal["jd", "resume"]


class Evidence(BaseModel):
    model_config = ConfigDict(populate_by_name=True)

    text: str
    source: EvidenceSource
    location: str | None = None


class RequirementMatch(BaseModel):
    model_config = ConfigDict(populate_by_name=True)

    id: str
    requirement: str
    jd_evidence: Evidence = Field(serialization_alias="jdEvidence")
    matched: bool
    match_strength: MatchStrength = Field(serialization_alias="matchStrength")
    resume_evidence: list[Evidence] = Field(
        default_factory=list, serialization_alias="resumeEvidence"
    )
    note: str | None = None


class SectionScore(BaseModel):
    model_config = ConfigDict(populate_by_name=True)

    id: SectionId
    label: str
    score: int = Field(ge=0, le=100)
    requirements: list[RequirementMatch]


class AnalysisResult(BaseModel):
    model_config = ConfigDict(populate_by_name=True)

    id: str
    created_at: datetime = Field(serialization_alias="createdAt")
    jd_title: str = Field(serialization_alias="jdTitle")
    resume_name: str = Field(serialization_alias="resumeName")
    status: AnalysisStatus
    sections: list[SectionScore] = Field(default_factory=list)
    error_message: str | None = Field(default=None, serialization_alias="errorMessage")
