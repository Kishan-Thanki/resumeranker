import math
from abc import ABC, abstractmethod

from pydantic import BaseModel

from app.schemas import SectionId


class DomainStrategy(ABC):
    """
    Abstract Base Class defining the contract for all industry domains.
    Any new domain (e.g., Tech, Sales, Healthcare) must implement these methods.
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
        match section_taxonomy(), and values must sum to 1.0.
        """

    @abstractmethod
    def get_scoring_schema(self) -> type[BaseModel]:
        """
        Dynamically generates and returns the schema the LLM actually
        fills in during scoring (verdicts and reviews).
        """

    @abstractmethod
    def get_final_schema(self) -> type[BaseModel]:
        """Dynamically generates and returns the FinalAnalysisResult Pydantic schema for this domain."""

    def validate_section_weights(self) -> None:
        """
        Concrete sanity check for subclasses. Call once per strategy at
        registration or app startup. Catches misconfigured domains before
        they produce incorrect final scores.
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

        invalid_weights = {k: v for k, v in weights.items() if v <= 0}
        if invalid_weights:
            raise ValueError(
                f"{self.name}: section_weights must be positive (> 0), got: {invalid_weights}"
            )

        total = sum(weights.values())
        if not math.isclose(total, 1.0, abs_tol=1e-6):
            raise ValueError(
                f"{self.name}: section_weights must sum to 1.0, got {total}"
            )
