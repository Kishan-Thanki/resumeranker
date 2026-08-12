import unicodedata
from typing import Literal

from pydantic import BaseModel, ConfigDict, Field, ValidationInfo, model_validator

SectionId = str
MatchStrength = Literal["strong", "partial", "weak", "none"]
EvidenceSource = Literal["jd", "resume"]


def normalize_text(text: str) -> str:
    """
    Normalizes unicode punctuation (smart quotes, em-dashes) and collapses
    all whitespace sequences into a single space for reliable substring matching.
    """
    text = unicodedata.normalize("NFKC", text)
    text = (
        text.replace("“", '"')
        .replace("”", '"')
        .replace("’", "'")
        .replace("‘", "'")
        .replace("–", "-")
        .replace("—", "-")
    )
    return " ".join(text.split())


class Evidence(BaseModel):
    model_config = ConfigDict(str_strip_whitespace=True)

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
        Verifies evidence text exists verbatim in source text if provided in context.

        Expected context format:
            context={
                "source_texts": {"jd": jd_text, "resume": resume_text},
                "normalized_source_texts": {"jd": norm_jd, "resume": norm_resume} # Optional pre-computed cache
            }
        """
        context = info.context or {}

        normalized_sources: dict[str, str] = context.get("normalized_source_texts", {})
        normalized_doc = normalized_sources.get(self.source)

        if normalized_doc is None:
            source_texts: dict[str, str] = context.get("source_texts", {})
            raw_doc = source_texts.get(self.source)
            if raw_doc is not None:
                normalized_doc = normalize_text(raw_doc)

        if normalized_doc is not None:
            normalized_quote = normalize_text(self.text)
            if normalized_quote not in normalized_doc:
                raise ValueError(
                    f"Evidence text not found verbatim in {self.source} source: {self.text!r}"
                )

        return self
