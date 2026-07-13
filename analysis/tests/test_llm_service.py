import pytest
import os
from unittest.mock import patch, AsyncMock

from app.services.llm_service import analyze_fit, JDRequirementsResponse, ResumeClaimsResponse
from app.schemas import (
    FinalAnalysisResult, 
    SectionsAnalysis, 
    SectionScore
)

class FakeUsage:
    prompt_tokens = 10
    completion_tokens = 5

class FakeRaw:
    usage = FakeUsage()

final_resp = FinalAnalysisResult(
    complete_analysis="Candidate is a good fit.",
    sections_analysis=SectionsAnalysis(skills="good", experience="good", education="good", leadership="good"),
    sections=[
        SectionScore(id="skills", label="Skills", score=90, requirements=[])
    ]
)

async def custom_create_with_completion(*args, **kwargs):
    """Dynamically route the mock based on which Pydantic schema instructor is asking for."""
    model = kwargs.get("response_model")
    if model == JDRequirementsResponse:
        return JDRequirementsResponse(requirements=[]), FakeRaw()
    elif model == ResumeClaimsResponse:
        return ResumeClaimsResponse(claims=[]), FakeRaw()
    elif model == FinalAnalysisResult:
        return final_resp, FakeRaw()
    raise ValueError(f"Unknown model requested: {model}")

@pytest.mark.asyncio
@patch("app.services.llm_service.instructor.from_litellm")
@patch.dict(os.environ, {"LLM_API_KEY": "fake_key", "LLM_MODEL": "fake_model"})
async def test_analyze_fit_pipeline(mock_from_litellm):
    """
    Integration test for the entire LLM pipeline.
    This intercepts all network calls so we can test the gather/scoring logic for free.
    """
    mock_client = AsyncMock()
    mock_client.chat.completions.create_with_completion.side_effect = custom_create_with_completion
    mock_from_litellm.return_value = mock_client
    
    result, usage = await analyze_fit("Fake JD", "Fake Resume")
    
    assert result.complete_analysis == "Candidate is a good fit."
    assert result.sections[0].score == 90
    
    assert usage["total_tokens"] == 45
    assert usage["prompt_tokens"] == 30
    assert usage["completion_tokens"] == 15
