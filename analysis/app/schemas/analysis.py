from pydantic import BaseModel, ConfigDict, Field, field_validator
from .core import SectionId, MatchStrength, Evidence

class RequirementMatch(BaseModel):
    model_config = ConfigDict(populate_by_name=True)

    id: str = Field(description="The ID of the JD requirement being evaluated")
    requirement: str = Field(description="The text of the requirement")
    jd_evidence: Evidence = Field(serialization_alias="jdEvidence")
    matched: bool = Field(description="True if the candidate meets this requirement at least partially, False if weak or none")
    match_strength: MatchStrength = Field(serialization_alias="matchStrength", description="How strongly the candidate's resume meets this requirement")
    resume_evidence: list[Evidence] = Field(
        default_factory=list, 
        serialization_alias="resumeEvidence",
        description="The resume claims that prove the candidate meets this requirement (leave empty if unmatched)"
    )
    note: str | None = Field(default=None, description="A strictly factual statement about what is missing. Never give advice.")

class SectionScore(BaseModel):
    model_config = ConfigDict(populate_by_name=True)

    id: SectionId
    label: str = Field(description="The human-readable label for this section (e.g., 'Experience')")
    score: int = Field(ge=0, le=100, description="The final calculated percentage score for this section (0-100)")
    requirements: list[RequirementMatch]

    @field_validator('score')
    @classmethod
    def validate_score_bounds(cls, v: int) -> int:
        if not (0 <= v <= 100):
            raise ValueError("Score must be between 0 and 100")
        return v
