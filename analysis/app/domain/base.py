import math
from abc import ABC, abstractmethod

from pydantic import BaseModel

from app.schemas import SectionId


class DomainStrategy(ABC):
    """
    Abstract Base Class defining the contract for all industry domains.
    Any new domain (e.g., Sales, Healthcare) must implement these methods.
    """

    @property
    @abstractmethod
    def name(self) -> str:
        """The identifier name of the domain (e.g., 'tech', 'sales')"""

    @property
    @abstractmethod
    def prompt_version(self) -> str:
        """The version of the prompts being used"""

    @abstractmethod
    def jd_extraction_prompt(self) -> str:
        """System prompt for extracting requirements from the Job Description"""

    @abstractmethod
    def resume_extraction_prompt(self) -> str:
        """System prompt for extracting claims from the Resume"""

    @abstractmethod
    def scoring_prompt(self) -> str:
        """System prompt for scoring the extracted claims against the requirements"""

    @abstractmethod
    def section_taxonomy(self) -> list[SectionId]:
        """Returns the list of valid section IDs for this domain"""

    @abstractmethod
    def section_weights(self) -> dict[SectionId, float]:
        """
        Returns the scoring weight for each section. Keys must exactly
        match section_taxonomy(), and values must sum to 1.0 — call
        validate_section_weights() once at startup to check this rather
        than discovering a misconfigured domain from a wrong score.
        """

    @abstractmethod
    def get_scoring_schema(self) -> type[BaseModel]:
        """
        Dynamically generates and returns the schema the LLM actually
        fills in during scoring: verdicts only (match strength,
        supporting claim ids, section scores/reviews) -- not the fully
        assembled final result. Distinct from get_final_schema(): this
        is what the LLM produces; get_final_schema() is what your code
        assembles from it afterward.
        """

    @abstractmethod
    def get_final_schema(self) -> type[BaseModel]:
        """Dynamically generates and returns the FinalAnalysisResult Pydantic schema for this domain."""

    def validate_section_weights(self) -> None:
        """
        Concrete sanity check for subclasses. Call once per strategy at
        registration or app startup, not on every scoring call. Catches a
        misconfigured domain (typo'd section key, weights that don't sum
        to 1.0) before it silently produces a wrong final score.
        """
        taxonomy = set(self.section_taxonomy())
        weights = self.section_weights()
        weight_keys = set(weights)
        if weight_keys != taxonomy:
            missing = taxonomy - weight_keys
            extra = weight_keys - taxonomy
            raise ValueError(
                f"{self.name}: section_weights keys must match section_taxonomy exactly "
                f"(missing: {missing or None}, unexpected: {extra or None})"
            )
        total = sum(weights.values())
        if not math.isclose(total, 1.0, abs_tol=1e-6):
            raise ValueError(
                f"{self.name}: section_weights must sum to 1.0, got {total}"
            )
