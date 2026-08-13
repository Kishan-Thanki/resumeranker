import asyncio

import pytest
from app import servicer
from app.domain.tech import DynamicFinalAnalysisResult
from app.pb import analysis_pb2
from app.schemas import SectionScore
from app.services.llm import AnalysisEngineError


def make_final_analysis() -> DynamicFinalAnalysisResult:
    return DynamicFinalAnalysisResult(
        complete_analysis="Strong overall fit for the role.",
        sections=[
            SectionScore(
                id="skills",
                label="Skills",
                score=90,
                review="Strong technical coverage.",
            ),
            SectionScore(
                id="experience",
                label="Experience",
                score=85,
                review="Relevant production experience.",
            ),
            SectionScore(
                id="education",
                label="Education",
                score=80,
                review="Education is aligned.",
            ),
            SectionScore(
                id="project",
                label="Project",
                score=70,
                review="Some relevant project evidence is present.",
            ),
        ],
    )


def make_usage() -> dict[str, int | float]:
    return {
        "prompt_tokens": 500,
        "completion_tokens": 150,
        "total_tokens": 650,
        "queue_wait_seconds": 1.25,
        "retries": 2,
        "call_count": 3,
    }


def make_request() -> analysis_pb2.AnalyzeRequest:
    return analysis_pb2.AnalyzeRequest(
        resume_pdf=b"resume-pdf",
        job_description_pdf=b"jd-pdf",
        request_id="request-123",
    )


class TestGetLlmConfig:
    def test_reads_model_and_provider_from_environment(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setenv("LLM_MODEL", "openai/gpt-test")
        monkeypatch.setenv("LLM_PROVIDER", "fallback-provider")

        model, provider = servicer._get_llm_config()

        assert model == "openai/gpt-test"
        assert provider == "openai"

    def test_uses_explicit_provider_when_model_has_no_provider_prefix(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setenv("LLM_MODEL", "gpt-test")
        monkeypatch.setenv("LLM_PROVIDER", "openai")

        model, provider = servicer._get_llm_config()

        assert model == "gpt-test"
        assert provider == "openai"

    def test_defaults_when_environment_values_are_missing(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.delenv("LLM_MODEL", raising=False)
        monkeypatch.delenv("LLM_PROVIDER", raising=False)

        model, provider = servicer._get_llm_config()

        assert model == "gpt-4o"
        assert provider == "unknown"


class TestGetMaxPdfParsers:
    def test_returns_configured_value(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setenv("MAX_CONCURRENT_PDF_PARSERS", "7")

        assert servicer._get_max_pdf_parsers() == 7

    def test_returns_default_value(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.delenv("MAX_CONCURRENT_PDF_PARSERS", raising=False)

        assert servicer._get_max_pdf_parsers() == 4


class TestCancelPending:
    @pytest.mark.asyncio
    async def test_cancels_only_pending_tasks(self) -> None:
        async def never_finishes() -> None:
            await asyncio.Event().wait()

        pending_task = asyncio.create_task(never_finishes())

        completed_task = asyncio.create_task(asyncio.sleep(0))
        await completed_task

        servicer._cancel_pending(
            pending_task,
            completed_task,
        )

        await asyncio.sleep(0)

        assert pending_task.cancelled() is True
        assert completed_task.done() is True

        with pytest.raises(asyncio.CancelledError):
            await pending_task


class TestAnalysisEngineServicer:
    @pytest.mark.asyncio
    async def test_successful_analysis_returns_complete_response(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        request = make_request()
        final_analysis = make_final_analysis()
        usage = make_usage()

        async def fake_parse(
            pdf_bytes: bytes,
            doc_name: str,
        ) -> tuple[str, dict[str, float]]:
            if doc_name == "Job Description":
                return (
                    "JD extracted text",
                    {
                        "queue_wait": 0.10,
                        "duration": 0.50,
                    },
                )

            return (
                "Resume extracted text",
                {
                    "queue_wait": 0.20,
                    "duration": 0.70,
                },
            )

        async def fake_analyze_fit(
            jd_text: str,
            resume_text: str,
        ) -> tuple[DynamicFinalAnalysisResult, dict[str, int | float]]:
            assert jd_text == "JD extracted text"
            assert resume_text == "Resume extracted text"
            return final_analysis, usage

        monkeypatch.setattr(
            servicer,
            "_parse_pdf_safely",
            fake_parse,
        )
        monkeypatch.setattr(
            servicer,
            "analyze_fit",
            fake_analyze_fit,
        )
        monkeypatch.setenv(
            "LLM_MODEL",
            "openai/gpt-test",
        )

        service = servicer.AnalysisEngineServicer()

        response = await service.Analyze(
            request,
            object(),
        )

        assert response.success is True
        assert response.error_message == ""
        assert response.model == "openai/gpt-test"
        assert response.prompt_tokens == 500
        assert response.completion_tokens == 150
        assert response.total_tokens == 650

        assert (
            response.result_json
            == final_analysis.model_dump_json(
                by_alias=True,
                exclude_none=True,
            )
        )

        assert servicer.request_id_var.get() == "request-123"

    @pytest.mark.asyncio
    async def test_successful_analysis_propagates_provider_name(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        request = make_request()
        final_analysis = make_final_analysis()

        async def fake_parse(
            _pdf_bytes: bytes,
            _doc_name: str,
        ) -> tuple[str, dict[str, float]]:
            return (
                "Extracted text",
                {
                    "queue_wait": 0.0,
                    "duration": 0.1,
                },
            )

        async def fake_analyze_fit(
            _jd_text: str,
            _resume_text: str,
        ) -> tuple[DynamicFinalAnalysisResult, dict[str, int | float]]:
            return (
                final_analysis,
                make_usage(),
            )

        monkeypatch.setattr(
            servicer,
            "_parse_pdf_safely",
            fake_parse,
        )
        monkeypatch.setattr(
            servicer,
            "analyze_fit",
            fake_analyze_fit,
        )
        monkeypatch.setenv(
            "LLM_MODEL",
            "anthropic/test-model",
        )

        service = servicer.AnalysisEngineServicer()

        response = await service.Analyze(
            request,
            object(),
        )

        assert response.success is True
        assert response.model == "anthropic/test-model"

    @pytest.mark.asyncio
    async def test_pdf_error_returns_expected_failure_response(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        request = make_request()

        async def fake_parse(
            _pdf_bytes: bytes,
            doc_name: str,
        ) -> tuple[str, dict[str, float]]:
            if doc_name == "Job Description":
                raise AnalysisEngineError(
                    code="ERR_PDF_PARSE",
                    message="Job Description could not be processed.",
                    retryable=False,
                )

            return (
                "Resume extracted text",
                {
                    "queue_wait": 0.0,
                    "duration": 0.1,
                },
            )

        analyze_mock_called = False

        async def fake_analyze_fit(
            _jd_text: str,
            _resume_text: str,
        ) -> tuple[DynamicFinalAnalysisResult, dict[str, int | float]]:
            nonlocal analyze_mock_called
            analyze_mock_called = True
            return make_final_analysis(), make_usage()

        monkeypatch.setattr(
            servicer,
            "_parse_pdf_safely",
            fake_parse,
        )
        monkeypatch.setattr(
            servicer,
            "analyze_fit",
            fake_analyze_fit,
        )

        service = servicer.AnalysisEngineServicer()

        response = await service.Analyze(
            request,
            object(),
        )

        assert response.success is False
        assert response.result_json == ""
        assert response.model == ""
        assert response.prompt_tokens == 0
        assert response.completion_tokens == 0
        assert response.total_tokens == 0
        assert (
            response.error_message
            == "ERR_PDF_PARSE: Job Description could not be processed."
        )
        assert analyze_mock_called is False

    @pytest.mark.asyncio
    async def test_unexpected_pdf_error_returns_generic_failure_response(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        request = make_request()

        async def fake_parse(
            _pdf_bytes: bytes,
            _doc_name: str,
        ) -> tuple[str, dict[str, float]]:
            raise RuntimeError("unexpected PDF failure")

        monkeypatch.setattr(
            servicer,
            "_parse_pdf_safely",
            fake_parse,
        )

        service = servicer.AnalysisEngineServicer()

        response = await service.Analyze(
            request,
            object(),
        )

        assert response.success is False
        assert (
            response.error_message
            == "ERR_INTERNAL: An unexpected error occurred while "
            "processing the documents."
        )

    @pytest.mark.asyncio
    async def test_llm_analysis_error_returns_expected_failure_response(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        request = make_request()

        async def fake_parse(
            _pdf_bytes: bytes,
            doc_name: str,
        ) -> tuple[str, dict[str, float]]:
            return (
                f"{doc_name} text",
                {
                    "queue_wait": 0.0,
                    "duration": 0.1,
                },
            )

        async def fake_analyze_fit(
            _jd_text: str,
            _resume_text: str,
        ) -> tuple[DynamicFinalAnalysisResult, dict[str, int | float]]:
            raise AnalysisEngineError(
                code="ERR_RATE_LIMIT_EXCEEDED",
                message="LLM rate limit reached.",
                retryable=True,
            )

        monkeypatch.setattr(
            servicer,
            "_parse_pdf_safely",
            fake_parse,
        )
        monkeypatch.setattr(
            servicer,
            "analyze_fit",
            fake_analyze_fit,
        )

        service = servicer.AnalysisEngineServicer()

        response = await service.Analyze(
            request,
            object(),
        )

        assert response.success is False
        assert (
            response.error_message
            == "ERR_RATE_LIMIT_EXCEEDED: LLM rate limit reached."
        )
        assert response.result_json == ""
        assert response.model == ""

    @pytest.mark.asyncio
    async def test_unexpected_llm_error_returns_generic_failure_response(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        request = make_request()

        async def fake_parse(
            _pdf_bytes: bytes,
            _doc_name: str,
        ) -> tuple[str, dict[str, float]]:
            return (
                "Extracted text",
                {
                    "queue_wait": 0.0,
                    "duration": 0.1,
                },
            )

        async def fake_analyze_fit(
            _jd_text: str,
            _resume_text: str,
        ) -> tuple[DynamicFinalAnalysisResult, dict[str, int | float]]:
            raise RuntimeError("unexpected LLM failure")

        monkeypatch.setattr(
            servicer,
            "_parse_pdf_safely",
            fake_parse,
        )
        monkeypatch.setattr(
            servicer,
            "analyze_fit",
            fake_analyze_fit,
        )

        service = servicer.AnalysisEngineServicer()

        response = await service.Analyze(
            request,
            object(),
        )

        assert response.success is False
        assert (
            response.error_message
            == "ERR_INTERNAL: An unexpected error occurred during AI analysis."
        )

    @pytest.mark.asyncio
    async def test_pdf_requests_are_started_for_both_documents(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        request = make_request()
        started: list[str] = []

        async def fake_parse(
            pdf_bytes: bytes,
            doc_name: str,
        ) -> tuple[str, dict[str, float]]:
            started.append(doc_name)

            await asyncio.sleep(0)

            return (
                f"{doc_name} text",
                {
                    "queue_wait": 0.0,
                    "duration": 0.1,
                },
            )

        async def fake_analyze_fit(
            _jd_text: str,
            _resume_text: str,
        ) -> tuple[DynamicFinalAnalysisResult, dict[str, int | float]]:
            return make_final_analysis(), make_usage()

        monkeypatch.setattr(
            servicer,
            "_parse_pdf_safely",
            fake_parse,
        )
        monkeypatch.setattr(
            servicer,
            "analyze_fit",
            fake_analyze_fit,
        )

        service = servicer.AnalysisEngineServicer()

        response = await service.Analyze(
            request,
            object(),
        )

        assert response.success is True
        assert set(started) == {
            "Job Description",
            "Resume",
        }

    @pytest.mark.asyncio
    async def test_analysis_cancellation_is_propagated(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        request = make_request()

        async def fake_parse(
            _pdf_bytes: bytes,
            _doc_name: str,
        ) -> tuple[str, dict[str, float]]:
            return (
                "Extracted text",
                {
                    "queue_wait": 0.0,
                    "duration": 0.1,
                },
            )

        async def fake_analyze_fit(
            _jd_text: str,
            _resume_text: str,
        ) -> tuple[DynamicFinalAnalysisResult, dict[str, int | float]]:
            raise asyncio.CancelledError

        monkeypatch.setattr(
            servicer,
            "_parse_pdf_safely",
            fake_parse,
        )
        monkeypatch.setattr(
            servicer,
            "analyze_fit",
            fake_analyze_fit,
        )

        service = servicer.AnalysisEngineServicer()

        with pytest.raises(asyncio.CancelledError):
            await service.Analyze(
                request,
                object(),
            )
