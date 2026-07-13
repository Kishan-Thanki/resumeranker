from typing import Literal

from pydantic import BaseModel, ConfigDict, Field, field_validator

SectionId = Literal["skills", "experience", "education", "leadership"]
MatchStrength = Literal["strong", "partial", "weak", "none"]
EvidenceSource = Literal["jd", "resume"]

class Evidence(BaseModel):
    model_config = ConfigDict(populate_by_name=True)

    text: str = Field(description="The exact verbatim quote from the source text. NEVER paraphrase.")
    source: EvidenceSource = Field(description="The source document this evidence came from.")
    location: str | None = Field(default=None, description="The literal heading name or section from the source where this was found.")

class ExtractedRequirement(BaseModel):
    id: str = Field(min_length=1, description="A unique identifier for this requirement (e.g., req-1)")
    section: SectionId = Field(description="The broad category this requirement falls under")
    requirement: str = Field(description="A short, concise imperative phrase capturing what's required (max 15 words)")
    jd_evidence: Evidence = Field(description="The exact evidence proving this requirement exists in the Job Description")

class ResumeClaim(BaseModel):
    id: str = Field(min_length=1, description="A unique identifier for this claim (e.g., claim-1)")
    section: SectionId = Field(description="The broad category this claim falls under")
    claim: str = Field(description="A short, concise statement summarizing the candidate's claim (max 15 words)")
    resume_evidence: Evidence = Field(description="The exact evidence proving this claim exists in the candidate's Resume")

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

class SectionsAnalysis(BaseModel):
    model_config = ConfigDict(populate_by_name=True)
    
    skills: str = Field(description="Qualitative review of skills gap (1-2 sentences max)")
    experience: str = Field(description="Qualitative review of experience gap (1-2 sentences max)")
    education: str = Field(description="Qualitative review of education gap (1-2 sentences max)")
    leadership: str = Field(description="Qualitative review of leadership gap (1-2 sentences max)")

class FinalAnalysisResult(BaseModel):
    model_config = ConfigDict(populate_by_name=True)

    complete_analysis: str = Field(serialization_alias="completeAnalysis", description="Executive summary of the candidate's overall fit (2-3 sentences max)")
    sections_analysis: SectionsAnalysis = Field(serialization_alias="sectionsAnalysis")
    sections: list[SectionScore]

class EngineRequest(BaseModel):
    resume_pdf: str 
    job_description_pdf: str 

class EngineResponse(BaseModel):
    result_json: str
    model: str
    prompt_tokens: int
    completion_tokens: int
    total_tokens: int
