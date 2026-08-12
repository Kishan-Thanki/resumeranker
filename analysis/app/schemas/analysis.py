from pydantic import (
    BaseModel,
    ConfigDict,
    Field,
    ValidationInfo,
    computed_field,
    field_validator,
)

from .core import Evidence, MatchStrength, SectionId
from .extraction import ExtractedRequirement, ResumeClaim


class MatchVerdict(BaseModel):
    """The LLM judgment on requirement matching."""

    model_config = ConfigDict(str_strip_whitespace=True)

    id: str = Field(
        description=(
            "The id of the ExtractedRequirement this verdict answers for. "
            "Must exactly match one of the requirement ids provided."
        )
    )
    match_strength: MatchStrength = Field(
        description="How strongly the candidate's resume meets this requirement",
    )
    supporting_claim_ids: list[str] = Field(
        default_factory=list,
        description=(
            "The id(s) of the resume claim(s) proving the requirement is met."
        ),
    )
    note: str | None = Field(
        default=None,
        max_length=300,
        description="Factual summary of missing aspects.",
    )


class SectionVerdict(BaseModel):
    """The LLM judgment on a section's score and review."""

    model_config = ConfigDict(str_strip_whitespace=True)

    id: SectionId = Field(description="The section id this verdict is for.")
    score: int = Field(
        ge=0,
        le=100,
        description="Assessed score for this section (0-100).",
    )
    review: str = Field(
        max_length=400,
        description="Qualitative review (1-2 sentences) of section gaps or strengths.",
    )

    @field_validator("id")
    @classmethod
    def check_section_taxonomy(cls, v: SectionId, info: ValidationInfo) -> SectionId:
        context = info.context or {}
        valid_sections = context.get("valid_sections")
        if valid_sections is not None and v not in valid_sections:
            raise ValueError(
                f"section id {v!r} is not in this domain's section taxonomy"
            )
        return v


class RequirementMatch(BaseModel):
    model_config = ConfigDict(str_strip_whitespace=True)

    id: str
    requirement: str
    jd_evidence: Evidence = Field(
        serialization_alias="jdEvidence",
        description="Exact evidence proving requirement exists in JD",
    )
    match_strength: MatchStrength = Field(
        serialization_alias="matchStrength",
        description="How strongly requirement is met",
    )
    resume_evidence: list[Evidence] = Field(
        default_factory=list,
        serialization_alias="resumeEvidence",
        description="Resume claims proving requirement is met",
    )
    note: str | None = Field(
        default=None,
        max_length=300,
        description="Factual statement of what is missing.",
    )

    @computed_field
    @property
    def matched(self) -> bool:
        return self.match_strength in ("strong", "partial")

    @field_validator("jd_evidence")
    @classmethod
    def check_jd_evidence_source(cls, v: Evidence) -> Evidence:
        if v.source != "jd":
            raise ValueError(f"jd_evidence.source must be 'jd', got {v.source!r}")
        return v

    @field_validator("resume_evidence")
    @classmethod
    def check_resume_evidence_source(cls, v: list[Evidence]) -> list[Evidence]:
        for evidence in v:
            if evidence.source != "resume":
                raise ValueError(
                    f"resume_evidence entries must have source='resume', got {evidence.source!r}"
                )
        return v

    @classmethod
    def from_verdict(
        cls,
        requirement: ExtractedRequirement,
        verdict: MatchVerdict,
        claims_by_id: dict[str, ResumeClaim],
    ) -> "RequirementMatch":
        if requirement.id is None:
            raise ValueError("requirement.id must be assigned before matching.")
        if verdict.id != requirement.id:
            raise ValueError(
                f"verdict.id {verdict.id!r} does not match requirement.id {requirement.id!r}"
            )
        resume_evidence: list[Evidence] = []
        for claim_id in verdict.supporting_claim_ids:
            claim = claims_by_id.get(claim_id)
            if claim is None:
                raise ValueError(
                    f"verdict for requirement {requirement.id!r} references "
                    f"unknown resume claim id {claim_id!r}"
                )
            resume_evidence.append(claim.resume_evidence)
        return cls(
            id=requirement.id,
            requirement=requirement.requirement,
            jd_evidence=requirement.jd_evidence,
            match_strength=verdict.match_strength,
            resume_evidence=resume_evidence,
            note=verdict.note,
        )


class SectionScore(BaseModel):
    model_config = ConfigDict(str_strip_whitespace=True)

    id: SectionId
    label: str = Field(description="Human-readable label for section")
    score: int = Field(ge=0, le=100, description="Section percentage score (0-100)")
    review: str = Field(max_length=400, description="Qualitative review of section")
    requirements: list[RequirementMatch] = Field(
        default_factory=list,
        description="Matched requirements for section",
    )

    @field_validator("id")
    @classmethod
    def check_section_taxonomy(cls, v: SectionId, info: ValidationInfo) -> SectionId:
        context = info.context or {}
        valid_sections = context.get("valid_sections")
        if valid_sections is not None and v not in valid_sections:
            raise ValueError(
                f"section id {v!r} is not in this domain's section taxonomy"
            )
        return v
