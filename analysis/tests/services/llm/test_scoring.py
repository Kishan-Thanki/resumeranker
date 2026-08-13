from typing import Any

import app.services.llm.service as llm_service
import pytest
from app.domain.tech import TechScoringResult
from app.schemas import MatchVerdict, SectionVerdict


def make_scoring_result() -> TechScoringResult:
    return TechScoringResult(
        complete_analysis="Strong overall fit for the role.",
        match_verdicts=[
            MatchVerdict(
                id="req-1",
                match_strength="strong",
                supporting_claim_ids=["claim-1"],
            ),
            MatchVerdict(
                id="req-2",
                match_strength="strong",
                supporting_claim_ids=["claim-2"],
            ),
            MatchVerdict(
                id="req-3",
                match_strength="partial",
                supporting_claim_ids=["claim-3"],
                note="The evidence covers part of the requirement.",
            ),
        ],
        section_verdicts=[
            SectionVerdict(
                id="skills",
                score=90,
                review="Strong technical skill coverage.",
            ),
            SectionVerdict(
                id="experience",
                score=95,
                review="Relevant production experience is well demonstrated.",
            ),
            SectionVerdict(
                id="education",
                score=90,
                review="The requested education is present.",
            ),
            SectionVerdict(
                id="project",
                score=40,
                review="Project evidence is limited.",
            ),
        ],
        matched_skills=["Python", "FastAPI", "PostgreSQL"],
        missing_critical_skills=["Kubernetes"],
    )

class TestScoreClaims:
    @pytest.mark.asyncio
    async def test_sends_requirements_and_claims_to_llm(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )
        scoring_result = make_scoring_result()

        usage = {
            "prompt_tokens": 500,
            "completion_tokens": 150,
            "queue_wait_seconds": 0.75,
            "retries": 1,
        }

        captured: dict[str, Any] = {}

        async def fake_call_llm(*args: Any, **kwargs: Any):
            captured.update(kwargs)
            return scoring_result, usage

        monkeypatch.setattr(
            llm_service,
            "call_llm",
            fake_call_llm,
        )

        result, returned_usage = await llm_service.score_claims(
            client=object(),
            llm_model="test-model",
            llm_api_key="test-key",
            domain=tech_domain,
            jd_reqs=jd_response,
            res_claims=resume_response,
            response_model=TechScoringResult,
        )

        assert result is scoring_result
        assert returned_usage == usage

        assert captured["response_model"] is TechScoringResult
        assert captured["model"] == "test-model"
        assert captured["api_key"] == "test-key"
        assert captured["stage"] == "Claims Scoring"

        messages = captured["messages"]
        assert isinstance(messages, list)
        assert len(messages) == 2

        system_message = messages[0]
        user_message = messages[1]

        assert system_message == {
            "role": "system",
            "content": tech_domain.scoring_prompt(),
        }

        assert user_message["role"] == "user"

        scoring_input = user_message["content"]

        assert isinstance(scoring_input, str)

        expected_jd_json = jd_response.model_dump_json(exclude_none=True)
        expected_resume_json = resume_response.model_dump_json(
            exclude_none=True
        )

        assert "JD Requirements:\n" in scoring_input
        assert expected_jd_json in scoring_input
        assert "Resume Claims:\n" in scoring_input
        assert expected_resume_json in scoring_input

        assert captured["context"] == {
            "valid_sections": tech_domain.section_taxonomy(),
        }

    @pytest.mark.asyncio
    async def test_preserves_requirement_and_claim_ids_in_scoring_input(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        scoring_result = make_scoring_result()
        usage = {
            "prompt_tokens": 100,
            "completion_tokens": 50,
            "queue_wait_seconds": 0.0,
            "retries": 0,
        }

        captured: dict[str, Any] = {}

        async def fake_call_llm(*args: Any, **kwargs: Any):
            captured.update(kwargs)
            return scoring_result, usage

        monkeypatch.setattr(
            llm_service,
            "call_llm",
            fake_call_llm,
        )

        await llm_service.score_claims(
            client=object(),
            llm_model="test-model",
            llm_api_key="test-key",
            domain=tech_domain,
            jd_reqs=jd_response,
            res_claims=resume_response,
            response_model=TechScoringResult,
        )

        messages = captured["messages"]
        scoring_input = messages[1]["content"]

        assert isinstance(scoring_input, str)

        for requirement in sample_requirements:
            assert requirement.id in scoring_input

        for claim in sample_claims:
            assert claim.id in scoring_input

    @pytest.mark.asyncio
    async def test_uses_domain_scoring_schema_passed_by_caller(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        scoring_schema = tech_domain.get_scoring_schema()

        scoring_result = make_scoring_result()
        usage = {
            "prompt_tokens": 100,
            "completion_tokens": 50,
            "queue_wait_seconds": 0.0,
            "retries": 0,
        }

        captured: dict[str, Any] = {}

        async def fake_call_llm(*args: Any, **kwargs: Any):
            captured.update(kwargs)
            return scoring_result, usage

        monkeypatch.setattr(
            llm_service,
            "call_llm",
            fake_call_llm,
        )

        await llm_service.score_claims(
            client=object(),
            llm_model="test-model",
            llm_api_key="test-key",
            domain=tech_domain,
            jd_reqs=jd_response,
            res_claims=resume_response,
            response_model=scoring_schema,
        )

        assert captured["response_model"] is scoring_schema

    @pytest.mark.asyncio
    async def test_passes_validation_context_for_domain_sections(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        scoring_result = make_scoring_result()
        usage = {
            "prompt_tokens": 100,
            "completion_tokens": 50,
            "queue_wait_seconds": 0.0,
            "retries": 0,
        }

        captured: dict[str, Any] = {}

        async def fake_call_llm(*args: Any, **kwargs: Any):
            captured.update(kwargs)
            return scoring_result, usage

        monkeypatch.setattr(
            llm_service,
            "call_llm",
            fake_call_llm,
        )

        await llm_service.score_claims(
            client=object(),
            llm_model="test-model",
            llm_api_key="test-key",
            domain=tech_domain,
            jd_reqs=jd_response,
            res_claims=resume_response,
            response_model=TechScoringResult,
        )

        assert captured["context"] == {
            "valid_sections": [
                "skills",
                "experience",
                "education",
                "project",
            ],
        }

    @pytest.mark.asyncio
    async def test_returns_usage_unchanged(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        scoring_result = make_scoring_result()

        usage = {
            "prompt_tokens": 777,
            "completion_tokens": 222,
            "queue_wait_seconds": 3.125,
            "retries": 2,
        }

        async def fake_call_llm(*args: Any, **kwargs: Any):
            return scoring_result, usage

        monkeypatch.setattr(
            llm_service,
            "call_llm",
            fake_call_llm,
        )

        _, returned_usage = await llm_service.score_claims(
            client=object(),
            llm_model="test-model",
            llm_api_key="test-key",
            domain=tech_domain,
            jd_reqs=jd_response,
            res_claims=resume_response,
            response_model=TechScoringResult,
        )

        assert returned_usage is usage
