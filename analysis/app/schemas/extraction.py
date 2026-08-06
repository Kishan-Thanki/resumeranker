from pydantic import BaseModel, Field

from .core import Evidence, SectionId


class ExtractedRequirement(BaseModel):
    id: str = Field(
        min_length=1,
        description="A unique identifier for this requirement (e.g., req-1)",
    )
    section: SectionId = Field(
        description="The broad category this requirement falls under"
    )
    requirement: str = Field(
        description="A short, concise imperative phrase capturing what's required (max 15 words)"
    )
    jd_evidence: Evidence = Field(
        description="The exact evidence proving this requirement exists in the Job Description"
    )


class ResumeClaim(BaseModel):
    id: str = Field(
        min_length=1, description="A unique identifier for this claim (e.g., claim-1)"
    )
    section: SectionId = Field(description="The broad category this claim falls under")
    claim: str = Field(
        description="A short, concise statement summarizing the candidate's claim (max 15 words)"
    )
    resume_evidence: Evidence = Field(
        description="The exact evidence proving this claim exists in the candidate's Resume"
    )
