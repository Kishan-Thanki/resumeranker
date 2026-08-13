from unittest.mock import AsyncMock

import app.services.llm.service as llm_service
import pytest


class TestExtractJdRequirements:
    @pytest.mark.asyncio
    async def test_extracts_jd_requirements_using_domain_prompt(
        self,
        tech_domain,
        sample_jd_text,
        sample_requirements,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        usage = {
            "input_tokens": 100,
            "output_tokens": 50,
            "total_tokens": 150,
            "reasoning_tokens": None,
            "cached_input_tokens": None,
            "cache_creation_input_tokens": None,
            "cache_read_input_tokens": None,
            "queue_wait_seconds": 0.25,
            "retries": 1,
        }

        mock_call_llm = AsyncMock(
            return_value=(response, usage),
        )
        monkeypatch.setattr(
            llm_service,
            "call_llm",
            mock_call_llm,
        )

        result, returned_usage = await llm_service.extract_jd_requirements(
            client=object(),
            llm_model="test-model",
            llm_api_key="test-key",
            domain=tech_domain,
            jd_text=sample_jd_text,
        )

        assert result is response
        assert returned_usage == usage

        mock_call_llm.assert_awaited_once()

        assert mock_call_llm.await_args is not None
        kwargs = mock_call_llm.await_args.kwargs

        assert kwargs["response_model"] is llm_service.JDRequirementsResponse
        assert kwargs["client"] is not None
        assert kwargs["model"] == "test-model"
        assert kwargs["api_key"] == "test-key"
        assert kwargs["stage"] == "JD Extraction"

        assert kwargs["messages"] == [
            {
                "role": "system",
                "content": tech_domain.jd_extraction_prompt(),
            },
            {
                "role": "user",
                "content": sample_jd_text,
            },
        ]

        assert kwargs["context"] == {
            "source_texts": {
                "jd": sample_jd_text,
            },
            "valid_sections": tech_domain.section_taxonomy(),
        }

    @pytest.mark.asyncio
    async def test_strips_surrounding_whitespace_before_sending_jd(
        self,
        tech_domain,
        sample_requirements,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        raw_jd = "  \n  Python backend experience required.  \n "
        cleaned_jd = "Python backend experience required."

        response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        usage = {
            "input_tokens": 100,
            "output_tokens": 50,
            "total_tokens": 150,
            "reasoning_tokens": None,
            "cached_input_tokens": None,
            "cache_creation_input_tokens": None,
            "cache_read_input_tokens": None,
            "queue_wait_seconds": 0.25,
            "retries": 1,
        }

        mock_call_llm = AsyncMock(
            return_value=(response, usage),
        )
        monkeypatch.setattr(
            llm_service,
            "call_llm",
            mock_call_llm,
        )

        await llm_service.extract_jd_requirements(
            client=object(),
            llm_model="test-model",
            llm_api_key="test-key",
            domain=tech_domain,
            jd_text=raw_jd,
        )

        assert mock_call_llm.await_args is not None
        kwargs = mock_call_llm.await_args.kwargs

        assert kwargs["messages"][1]["content"] == cleaned_jd
        assert kwargs["context"]["source_texts"]["jd"] == cleaned_jd

    @pytest.mark.asyncio
    @pytest.mark.parametrize(
        "jd_text",
        ["", "   ", "\n\t  "],
    )
    async def test_rejects_empty_jd_text(
        self,
        tech_domain,
        jd_text: str,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        mock_call_llm = AsyncMock()
        monkeypatch.setattr(
            llm_service,
            "call_llm",
            mock_call_llm,
        )

        with pytest.raises(llm_service.AnalysisEngineError) as exc_info:
            await llm_service.extract_jd_requirements(
                client=object(),
                llm_model="test-model",
                llm_api_key="test-key",
                domain=tech_domain,
                jd_text=jd_text,
            )

        assert exc_info.value.code == "ERR_BAD_REQUEST"
        assert exc_info.value.retryable is False
        assert "Job description text cannot be empty" in exc_info.value.message

        mock_call_llm.assert_not_awaited()


class TestExtractResumeClaims:
    @pytest.mark.asyncio
    async def test_extracts_resume_claims_using_domain_prompt(
        self,
        tech_domain,
        sample_resume_text,
        sample_claims,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )
        usage = {
            "input_tokens": 100,
            "output_tokens": 50,
            "total_tokens": 150,
            "reasoning_tokens": None,
            "cached_input_tokens": None,
            "cache_creation_input_tokens": None,
            "cache_read_input_tokens": None,
            "queue_wait_seconds": 0.25,
            "retries": 1,
        }

        mock_call_llm = AsyncMock(
            return_value=(response, usage),
        )
        monkeypatch.setattr(
            llm_service,
            "call_llm",
            mock_call_llm,
        )

        result, returned_usage = await llm_service.extract_resume_claims(
            client=object(),
            llm_model="test-model",
            llm_api_key="test-key",
            domain=tech_domain,
            resume_text=sample_resume_text,
        )

        assert result is response
        assert returned_usage == usage

        mock_call_llm.assert_awaited_once()

        assert mock_call_llm.await_args is not None
        kwargs = mock_call_llm.await_args.kwargs

        assert kwargs["response_model"] is llm_service.ResumeClaimsResponse
        assert kwargs["client"] is not None
        assert kwargs["model"] == "test-model"
        assert kwargs["api_key"] == "test-key"
        assert kwargs["stage"] == "Resume Extraction"

        assert kwargs["messages"] == [
            {
                "role": "system",
                "content": tech_domain.resume_extraction_prompt(),
            },
            {
                "role": "user",
                "content": sample_resume_text,
            },
        ]

        assert kwargs["context"] == {
            "source_texts": {
                "resume": sample_resume_text,
            },
            "valid_sections": tech_domain.section_taxonomy(),
        }

    @pytest.mark.asyncio
    async def test_strips_surrounding_whitespace_before_sending_resume(
        self,
        tech_domain,
        sample_claims,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        raw_resume = "  \n  Senior Python Engineer.  \n "
        cleaned_resume = "Senior Python Engineer."

        response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )
        usage = {
            "input_tokens": 100,
            "output_tokens": 50,
            "total_tokens": 150,
            "reasoning_tokens": None,
            "cached_input_tokens": None,
            "cache_creation_input_tokens": None,
            "cache_read_input_tokens": None,
            "queue_wait_seconds": 0.25,
            "retries": 1,
        }

        mock_call_llm = AsyncMock(
            return_value=(response, usage),
        )
        monkeypatch.setattr(
            llm_service,
            "call_llm",
            mock_call_llm,
        )

        await llm_service.extract_resume_claims(
            client=object(),
            llm_model="test-model",
            llm_api_key="test-key",
            domain=tech_domain,
            resume_text=raw_resume,
        )

        assert mock_call_llm.await_args is not None
        kwargs = mock_call_llm.await_args.kwargs

        assert kwargs["messages"][1]["content"] == cleaned_resume
        assert kwargs["context"]["source_texts"]["resume"] == cleaned_resume

    @pytest.mark.asyncio
    @pytest.mark.parametrize(
        "resume_text",
        ["", "   ", "\n\t  "],
    )
    async def test_rejects_empty_resume_text(
        self,
        tech_domain,
        resume_text: str,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        mock_call_llm = AsyncMock()
        monkeypatch.setattr(
            llm_service,
            "call_llm",
            mock_call_llm,
        )

        with pytest.raises(llm_service.AnalysisEngineError) as exc_info:
            await llm_service.extract_resume_claims(
                client=object(),
                llm_model="test-model",
                llm_api_key="test-key",
                domain=tech_domain,
                resume_text=resume_text,
            )

        assert exc_info.value.code == "ERR_BAD_REQUEST"
        assert exc_info.value.retryable is False
        assert "Resume text cannot be empty" in exc_info.value.message

        mock_call_llm.assert_not_awaited()


class TestExtractionResponseModels:
    def test_jd_response_requires_at_least_one_requirement(
        self,
        sample_requirements,
    ) -> None:
        response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )

        assert len(response.requirements) == len(sample_requirements)

        with pytest.raises(ValueError):
            llm_service.JDRequirementsResponse(
                requirements=[],
            )

    def test_resume_response_requires_at_least_one_claim(
        self,
        sample_claims,
    ) -> None:
        response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        assert len(response.claims) == len(sample_claims)

        with pytest.raises(ValueError):
            llm_service.ResumeClaimsResponse(
                claims=[],
            )
