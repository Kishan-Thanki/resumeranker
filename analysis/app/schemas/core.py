from typing import Literal

from pydantic import BaseModel, ConfigDict, Field

SectionId = str
MatchStrength = Literal["strong", "partial", "weak", "none"]
EvidenceSource = Literal["jd", "resume"]


class Evidence(BaseModel):
    model_config = ConfigDict(populate_by_name=True)

    text: str = Field(
        description="The exact verbatim quote from the source text. NEVER paraphrase."
    )
    source: EvidenceSource = Field(
        description="The source document this evidence came from."
    )
    location: str | None = Field(
        default=None,
        description="The literal heading name or section from the source where this was found.",
    )
