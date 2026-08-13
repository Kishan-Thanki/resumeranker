import asyncio
from types import SimpleNamespace

import app.services.llm.service as llm_service
import pytest


class FakeTransientError(Exception):
    """Synthetic transient provider failure for retry tests."""

    response: object | None = None


class FakeConnectionError(Exception):
    """Synthetic connection failure for exception-mapping tests."""


class FakeTimeoutError(Exception):
    """Synthetic timeout failure for exception-mapping tests."""


class TestRequireId:
    def test_returns_existing_id(self) -> None:
        assert llm_service.require_id("req-1", "requirement") == "req-1"

    def test_rejects_missing_id(self) -> None:
        with pytest.raises(
            ValueError,
            match="requirement id was never assigned",
        ):
            llm_service.require_id(None, "requirement")


class TestAssignIds:
    def test_assigns_requirement_ids_and_claim_ids(
        self,
        sample_requirements,
        sample_claims,
    ) -> None:
        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        llm_service.assign_ids(jd_response, resume_response)

        assert [requirement.id for requirement in jd_response.requirements] == [
            "req-1",
            "req-2",
            "req-3",
        ]
        assert [claim.id for claim in resume_response.claims] == [
            "claim-1",
            "claim-2",
            "claim-3",
        ]

    def test_reassigns_existing_ids_deterministically(
        self,
        sample_requirements,
        sample_claims,
    ) -> None:
        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        jd_response.requirements[0].id = "old-req-id"
        resume_response.claims[0].id = "old-claim-id"

        llm_service.assign_ids(jd_response, resume_response)

        assert jd_response.requirements[0].id == "req-1"
        assert resume_response.claims[0].id == "claim-1"


class TestBuildUsage:
    def test_builds_usage_from_object_response(self) -> None:
        raw = SimpleNamespace(
            usage=SimpleNamespace(
                prompt_tokens=123,
                completion_tokens=45,
            ),
            _hidden_params={
                "retries": 2,
            },
        )

        usage = llm_service.build_usage(
            raw,
            queue_wait_seconds=1.25,
        )

        assert usage == {
            "input_tokens": 123,
            "output_tokens": 45,
            "total_tokens": None,
            "reasoning_tokens": None,
            "cached_input_tokens": None,
            "cache_creation_input_tokens": None,
            "cache_read_input_tokens": None,
            "queue_wait_seconds": 1.25,
            "retries": 2,
        }

    def test_builds_usage_from_dictionary_response(self) -> None:
        raw = SimpleNamespace(
            usage={
                "prompt_tokens": 200,
                "completion_tokens": 80,
            },
            _hidden_params={
                "retries": 1,
            },
        )

        usage = llm_service.build_usage(
            raw,
            queue_wait_seconds=0.5,
        )

        assert usage == {
            "input_tokens": 200,
            "output_tokens": 80,
            "total_tokens": None,
            "reasoning_tokens": None,
            "cached_input_tokens": None,
            "cache_creation_input_tokens": None,
            "cache_read_input_tokens": None,
            "queue_wait_seconds": 0.5,
            "retries": 1,
        }

    def test_defaults_missing_usage_values_to_none(self) -> None:
        raw = SimpleNamespace(
            usage=None,
            _hidden_params={},
        )

        usage = llm_service.build_usage(
            raw,
            queue_wait_seconds=0.0,
        )

        assert usage == {
            "input_tokens": None,
            "output_tokens": None,
            "total_tokens": None,
            "reasoning_tokens": None,
            "cached_input_tokens": None,
            "cache_creation_input_tokens": None,
            "cache_read_input_tokens": None,
            "queue_wait_seconds": 0.0,
            "retries": 0,
        }

    def test_preserves_none_usage_values(self) -> None:
        raw = SimpleNamespace(
            usage=SimpleNamespace(
                prompt_tokens=None,
                completion_tokens=None,
                total_tokens=None,
            ),
            _hidden_params={
                "retries": None,
            },
        )

        usage = llm_service.build_usage(
            raw,
            queue_wait_seconds=0.25,
        )

        assert usage == {
            "input_tokens": None,
            "output_tokens": None,
            "total_tokens": None,
            "reasoning_tokens": None,
            "cached_input_tokens": None,
            "cache_creation_input_tokens": None,
            "cache_read_input_tokens": None,
            "queue_wait_seconds": 0.25,
            "retries": 0,
        }


class TestParseRetryAfter:
    def test_reads_retry_after_header(self) -> None:
        exc = FakeTransientError()
        exc.response = SimpleNamespace(
            headers={
                "Retry-After": "3.5",
            }
        )

        assert llm_service._parse_retry_after(exc) == 3.5

    def test_reads_lowercase_retry_after_header(self) -> None:
        exc = FakeTransientError()
        exc.response = SimpleNamespace(
            headers={
                "retry-after": "2",
            }
        )

        assert llm_service._parse_retry_after(exc) == 2.0

    def test_returns_none_when_header_is_missing(self) -> None:
        exc = FakeTransientError()
        exc.response = SimpleNamespace(headers={})

        assert llm_service._parse_retry_after(exc) is None

    def test_returns_none_for_invalid_header(self) -> None:
        exc = FakeTransientError()
        exc.response = SimpleNamespace(
            headers={
                "Retry-After": "not-a-number",
            }
        )

        assert llm_service._parse_retry_after(exc) is None

    def test_returns_none_when_response_is_missing(self) -> None:
        exc = FakeTransientError()

        assert llm_service._parse_retry_after(exc) is None


class TestExecuteWithBackoff:
    @pytest.mark.asyncio
    async def test_returns_success_without_retry(self) -> None:
        calls = 0

        async def api_call() -> str:
            nonlocal calls
            calls += 1
            return "success"

        result = await llm_service.execute_with_backoff(api_call)

        assert result == "success"
        assert calls == 1

    @pytest.mark.asyncio
    async def test_retries_transient_failure_then_succeeds(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        calls = 0
        sleep_delays: list[float] = []

        async def api_call() -> str:
            nonlocal calls
            calls += 1

            if calls == 1:
                raise FakeTransientError("temporary failure")

            return "success"

        async def fake_sleep(delay: float) -> None:
            sleep_delays.append(delay)

        monkeypatch.setattr(
            llm_service,
            "TRANSIENT_ERRORS",
            (FakeTransientError,),
        )
        monkeypatch.setattr(
            llm_service.asyncio,
            "sleep",
            fake_sleep,
        )
        monkeypatch.setattr(
            llm_service,
            "MAX_RATE_LIMIT_RETRIES",
            2,
        )
        monkeypatch.setattr(
            llm_service,
            "BACKOFF_BASE_SECONDS",
            1.0,
        )
        monkeypatch.setattr(
            llm_service.random,
            "uniform",
            lambda _start, _end: 0.0,
        )

        result = await llm_service.execute_with_backoff(api_call)

        assert result == "success"
        assert calls == 2
        assert sleep_delays == [1.0]

    @pytest.mark.asyncio
    async def test_retries_until_maximum_then_raises(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        calls = 0
        sleep_delays: list[float] = []

        async def api_call() -> str:
            nonlocal calls
            calls += 1
            raise FakeTransientError("still failing")

        async def fake_sleep(delay: float) -> None:
            sleep_delays.append(delay)

        monkeypatch.setattr(
            llm_service,
            "TRANSIENT_ERRORS",
            (FakeTransientError,),
        )
        monkeypatch.setattr(
            llm_service.asyncio,
            "sleep",
            fake_sleep,
        )
        monkeypatch.setattr(
            llm_service,
            "MAX_RATE_LIMIT_RETRIES",
            2,
        )
        monkeypatch.setattr(
            llm_service,
            "BACKOFF_BASE_SECONDS",
            1.0,
        )
        monkeypatch.setattr(
            llm_service.random,
            "uniform",
            lambda _start, _end: 0.0,
        )

        with pytest.raises(FakeTransientError, match="still failing"):
            await llm_service.execute_with_backoff(api_call)

        assert calls == 3
        assert sleep_delays == [1.0, 2.0]

    @pytest.mark.asyncio
    async def test_respects_retry_after_header(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        calls = 0
        sleep_delays: list[float] = []

        async def api_call() -> str:
            nonlocal calls
            calls += 1

            if calls == 1:
                exc = FakeTransientError("rate limited")
                exc.response = SimpleNamespace(
                    headers={
                        "Retry-After": "4.5",
                    }
                )
                raise exc

            return "success"

        async def fake_sleep(delay: float) -> None:
            sleep_delays.append(delay)

        monkeypatch.setattr(
            llm_service,
            "TRANSIENT_ERRORS",
            (FakeTransientError,),
        )
        monkeypatch.setattr(
            llm_service.asyncio,
            "sleep",
            fake_sleep,
        )
        monkeypatch.setattr(
            llm_service,
            "MAX_RATE_LIMIT_RETRIES",
            1,
        )

        result = await llm_service.execute_with_backoff(api_call)

        assert result == "success"
        assert calls == 2
        assert sleep_delays == [4.5]

    @pytest.mark.asyncio
    async def test_non_transient_exception_is_not_retried(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        calls = 0

        async def api_call() -> str:
            nonlocal calls
            calls += 1
            raise ValueError("bad request")

        monkeypatch.setattr(
            llm_service,
            "TRANSIENT_ERRORS",
            (FakeTransientError,),
        )

        with pytest.raises(ValueError, match="bad request"):
            await llm_service.execute_with_backoff(api_call)

        assert calls == 1


class TestLLMRateLimiter:
    @pytest.mark.asyncio
    async def test_allows_requests_within_limit(self) -> None:
        limiter = llm_service.LLMRateLimiter(
            max_requests=2,
            window_seconds=60.0,
        )

        first_wait = await limiter.acquire()
        second_wait = await limiter.acquire()

        assert first_wait >= 0
        assert second_wait >= 0
        assert len(limiter._request_times) == 2

    @pytest.mark.asyncio
    async def test_releases_capacity_after_window_expires(
        self,
    ) -> None:
        limiter = llm_service.LLMRateLimiter(
            max_requests=1,
            window_seconds=0.001,
        )

        await limiter.acquire()

        await asyncio.sleep(0.005)

        wait = await limiter.acquire()

        assert wait >= 0
        assert len(limiter._request_times) == 1

    @pytest.mark.asyncio
    async def test_rejects_invalid_max_requests(self) -> None:
        with pytest.raises(
            ValueError,
            match="LLM_MAX_REQUESTS_PER_MINUTE must be greater than 0",
        ):
            llm_service.LLMRateLimiter(max_requests=0)

    @pytest.mark.asyncio
    async def test_rejects_invalid_window(self) -> None:
        with pytest.raises(
            ValueError,
            match="Rate-limit window must be greater than 0",
        ):
            llm_service.LLMRateLimiter(
                max_requests=1,
                window_seconds=0,
            )


class TestHandleLLMException:
    def test_preserves_existing_analysis_engine_error(self) -> None:
        original = llm_service.AnalysisEngineError(
            code="ERR_EXISTING",
            message="Already mapped",
            retryable=True,
        )

        with pytest.raises(llm_service.AnalysisEngineError) as exc_info:
            llm_service.handle_llm_exception(
                original,
                "test-stage",
            )

        assert exc_info.value is original

    def test_maps_connection_failure(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setattr(
            llm_service,
            "APIConnectionError",
            FakeConnectionError,
        )

        with pytest.raises(llm_service.AnalysisEngineError) as exc_info:
            llm_service.handle_llm_exception(
                FakeConnectionError("connection dropped"),
                "JD Extraction",
            )

        assert exc_info.value.code == "ERR_INFRA_CONNECTION_DROP"
        assert exc_info.value.retryable is True

    def test_maps_timeout_failure(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setattr(
            llm_service,
            "Timeout",
            FakeTimeoutError,
        )

        with pytest.raises(llm_service.AnalysisEngineError) as exc_info:
            llm_service.handle_llm_exception(
                FakeTimeoutError("timed out"),
                "Resume Extraction",
            )

        assert exc_info.value.code == "ERR_INFRA_TIMEOUT"
        assert exc_info.value.retryable is True

    def test_maps_unknown_exception_to_internal_error(self) -> None:
        with pytest.raises(llm_service.AnalysisEngineError) as exc_info:
            llm_service.handle_llm_exception(
                RuntimeError("unexpected failure"),
                "Scoring",
            )

        assert exc_info.value.code == "ERR_PIPELINE_INTERNAL"
        assert exc_info.value.retryable is False
        assert "Scoring" in exc_info.value.message


class TestAnalysisEngineError:
    def test_stores_error_information(self) -> None:
        error = llm_service.AnalysisEngineError(
            code="ERR_TEST",
            message="Test error",
            retryable=True,
        )

        assert error.code == "ERR_TEST"
        assert error.message == "Test error"
        assert error.retryable is True
        assert str(error) == "Test error"
