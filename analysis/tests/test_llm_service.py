import os
from typing import Any, cast
from unittest.mock import AsyncMock, patch

import pytest
from app.domain.tech import TechDomain
from app.schemas import SectionScore
from app.services.llm_service import (
    JDRequirementsResponse,
    ResumeClaimsResponse,
    analyze_fit,
)

# ---------------------------------------------------------------------
# Dynamic schema used by the analysis engine
# ---------------------------------------------------------------------

FinalAnalysisResult = cast(type[Any], TechDomain().get_final_schema())
SectionsAnalysis = FinalAnalysisResult.model_fields["sections_analysis"].annotation


class FakeUsage:
    prompt_tokens = 10
    completion_tokens = 5


class FakeRaw:
    usage = FakeUsage()


final_response = FinalAnalysisResult(
    complete_analysis="Candidate is a good fit.",
    sections_analysis=SectionsAnalysis(
        skills="good",
        experience="good",
        education="good",
        project="good",
    ),
    sections=[
        SectionScore(
            id="skills",
            label="Skills",
            score=90,
            requirements=[],
        )
    ],
)


async def custom_create_with_completion(*args, **kwargs):
    """
    Return the correct fake response depending on which Instructor
    response model is requested.
    """
    model = kwargs.get("response_model")

    if model is JDRequirementsResponse:
        return JDRequirementsResponse(requirements=[]), FakeRaw()

    if model is ResumeClaimsResponse:
        return ResumeClaimsResponse(claims=[]), FakeRaw()

    if getattr(model, "__name__", "") == "DynamicFinalAnalysisResult":
        return final_response, FakeRaw()

    raise ValueError(f"Unknown response model: {model!r}")


@pytest.mark.asyncio
@patch("app.services.llm_service.instructor.from_litellm")
@patch.dict(
    os.environ,
    {
        "LLM_API_KEY": "fake_key",
        "LLM_MODEL": "fake_model",
    },
)
async def test_analyze_fit_pipeline(mock_from_litellm):
    """
    Integration test for the complete LLM pipeline.

    All Instructor/LiteLLM calls are mocked, allowing the orchestration
    logic to be tested without making any external API requests.
    """
    mock_client = AsyncMock()
    mock_client.chat.completions.create_with_completion.side_effect = (
        custom_create_with_completion
    )
    mock_from_litellm.return_value = mock_client

    result, usage = await analyze_fit("Fake JD", "Fake Resume")

    result = cast(Any, result)

    assert result.complete_analysis == "Candidate is a good fit."
    assert result.sections[0].score == 90

    assert usage["prompt_tokens"] == 30
    assert usage["completion_tokens"] == 15
    assert usage["total_tokens"] == 45
