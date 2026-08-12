from pydantic import BaseModel, ConfigDict, Field, ValidationInfo, model_validator

from .core import Evidence, SectionId


class ExtractedRequirement(BaseModel):
    model_config = ConfigDict(str_strip_whitespace=True)

    id: str | None = Field(
        default=None,
        description="Leave unset. Assign this after extraction (e.g. via enumerate() over the parsed list) rather than trusting the model to keep a consistent counter.",
    )
    section: SectionId = Field(
        description="The broad category this requirement falls under"
    )
    requirement: str = Field(
        max_length=150,
        description="A short, concise imperative phrase capturing what's required (max 15 words)",
    )
    jd_evidence: Evidence = Field(
        description="The exact evidence proving this requirement exists in the Job Description"
    )

    @model_validator(mode="after")
    def check_evidence_source(self) -> "ExtractedRequirement":
        if self.jd_evidence.source != "jd":
            raise ValueError(
                f"jd_evidence.source must be 'jd', got {self.jd_evidence.source!r}"
            )
        return self

    @model_validator(mode="after")
    def check_section_taxonomy(self, info: ValidationInfo) -> "ExtractedRequirement":
        """
        Optional section-taxonomy check. Pass the active domain's valid
        sections via validation context:

            ExtractedRequirement.model_validate(
                data,
                context={"valid_sections": strategy.section_taxonomy()},
            )

        No-op if no context is supplied.
        """
        context = info.context or {}
        valid_sections = context.get("valid_sections")
        if valid_sections is not None and self.section not in valid_sections:
            raise ValueError(
                f"section {self.section!r} is not in this domain's section taxonomy"
            )
        return self


class ResumeClaim(BaseModel):
    model_config = ConfigDict(str_strip_whitespace=True)

    id: str | None = Field(
        default=None,
        description="Leave unset. Assign this after extraction (e.g. via enumerate() over the parsed list) rather than trusting the model to keep a consistent counter.",
    )
    section: SectionId = Field(description="The broad category this claim falls under")
    claim: str = Field(
        max_length=150,
        description="A short, concise statement summarizing the candidate's claim (max 15 words)",
    )
    resume_evidence: Evidence = Field(
        description="The exact evidence proving this claim exists in the candidate's Resume"
    )

    @model_validator(mode="after")
    def check_evidence_source(self) -> "ResumeClaim":
        if self.resume_evidence.source != "resume":
            raise ValueError(
                f"resume_evidence.source must be 'resume', got {self.resume_evidence.source!r}"
            )
        return self

    @model_validator(mode="after")
    def check_section_taxonomy(self, info: ValidationInfo) -> "ResumeClaim":
        context = info.context or {}
        valid_sections = context.get("valid_sections")
        if valid_sections is not None and self.section not in valid_sections:
            raise ValueError(
                f"section {self.section!r} is not in this domain's section taxonomy"
            )
        return self
