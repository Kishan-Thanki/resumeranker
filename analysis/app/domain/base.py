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
        pass
        
    @property
    @abstractmethod
    def prompt_version(self) -> str:
        """The version of the prompts being used"""
        pass

    @abstractmethod
    def jd_extraction_prompt(self) -> str:
        """System prompt for extracting requirements from the Job Description"""
        pass

    @abstractmethod
    def resume_extraction_prompt(self) -> str:
        """System prompt for extracting claims from the Resume"""
        pass

    @abstractmethod
    def scoring_prompt(self) -> str:
        """System prompt for scoring the extracted claims against the requirements"""
        pass
        
    @abstractmethod
    def section_taxonomy(self) -> list[SectionId]:
        """Returns the list of valid section IDs for this domain"""
        pass

    @abstractmethod
    def section_weights(self) -> dict[SectionId, float]:
        """Returns the scoring weight for each section (must sum to 1.0)"""
        pass

    @abstractmethod
    def get_final_schema(self) -> type[BaseModel]:
        """Dynamically generates and returns the FinalAnalysisResult Pydantic schema for this domain."""
        pass
