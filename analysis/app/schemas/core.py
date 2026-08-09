from typing import Literal

from pydantic import BaseModel, Field, ValidationInfo, model_validator

# Valid section IDs are NOT fixed globally — they're defined per domain via
# DomainStrategy.section_taxonomy() (see base.py). Kept as a plain str
# here rather than a Literal, since the actual valid set is only known at
# runtime once a domain strategy is selected. Enforcement happens in the
# model_validators in extraction.py / analysis.py, when a domain's
# taxonomy is passed in via validation context.
SectionId = str

MatchStrength = Literal["strong", "partial", "weak", "none"]
EvidenceSource = Literal["jd", "resume"]


class Evidence(BaseModel):
    text: str = Field(
        min_length=1,
        max_length=400,
        description="The exact verbatim quote from the source text. NEVER paraphrase.",
    )
    source: EvidenceSource = Field(
        description="The source document this evidence came from."
    )
    location: str | None = Field(
        default=None,
        description="The literal heading name or section from the source where this was found.",
    )

    @model_validator(mode="after")
    def verify_verbatim(self, info: ValidationInfo) -> "Evidence":
        """
        Optional verbatim check. Pass validation context like:

            Evidence.model_validate(
                data,
                context={"source_texts": {"jd": jd_text, "resume": resume_text}},
            )

        No-op if no context (or no entry for this source) is supplied, so
        existing call sites without context still work.
        """
        context = info.context or {}
        source_texts: dict[str, str] = context.get("source_texts", {})
        source_doc = source_texts.get(self.source)
        if source_doc is not None:
            normalized_quote = " ".join(self.text.split())
            normalized_doc = " ".join(source_doc.split())
            if normalized_quote not in normalized_doc:
                raise ValueError(
                    f"Evidence text not found verbatim in {self.source} source: {self.text!r}"
                )
        return self
