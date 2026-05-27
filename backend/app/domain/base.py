from typing import Protocol

from app.schemas.analysis import SectionId


class Domain(Protocol):
    """A domain plugs role-specific prompts and section taxonomies into the
    matching pipeline. v1 ships only `tech`; the seam exists so adding
    `medical`, `sales`, etc. doesn't require touching services."""

    name: str
    prompt_version: str

    def jd_extraction_prompt(self) -> str: ...
    def resume_extraction_prompt(self) -> str: ...
    def scoring_prompt(self) -> str: ...
    def section_taxonomy(self) -> list[SectionId]: ...
    def section_weights(self) -> dict[SectionId, float]: ...
