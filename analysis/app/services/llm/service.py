# service.py

from __future__ import annotations

import asyncio
import functools
import os
import random
import time
from collections import deque
from collections.abc import Awaitable, Callable
from typing import Any, NoReturn, ParamSpec, Protocol, TypeVar, cast

import instructor
from instructor.core import InstructorRetryException
from litellm import acompletion
from litellm.exceptions import (
    APIConnectionError,
    APIError,
    AuthenticationError,
    BadRequestError,
    ContentPolicyViolationError,
    ContextWindowExceededError,
    InternalServerError,
    NotFoundError,
    PermissionDeniedError,
    RateLimitError,
    ServiceUnavailableError,
    Timeout,
    UnprocessableEntityError,
)
from pydantic import BaseModel, ConfigDict, Field, ValidationError

from app.domain.base import DomainStrategy
from app.domain.tech import TechDomain
from app.logger import logger
from app.schemas import (
    AggregatedLLMUsage,
    ExtractedRequirement,
    LLMUsage,
    MatchVerdict,
    RequirementMatch,
    ResumeClaim,
    SectionScore,
    SectionVerdict,
    aggregate_llm_usage,
)
from app.utils.telemetry import track_llm_cost

# ============================================================================
# EXCEPTIONS
# ============================================================================


class AnalysisEngineError(Exception):
    """Base domain exception for analysis engine errors."""

    def __init__(
        self,
        code: str,
        message: str,
        retryable: bool = False,
    ) -> None:
        self.code = code
        self.message = message
        self.retryable = retryable
        super().__init__(self.message)


def handle_llm_exception(
    exc: Exception,
    stage: str,
) -> NoReturn:
    """Translate LLM-specific exceptions across operational layers into
    AnalysisEngineErrors.
    """
    if isinstance(exc, AnalysisEngineError):
        raise exc

    logger.error(
        "LLM exception occurred during pipeline execution",
        extra={
            "stage": stage,
            "exception_type": type(exc).__name__,
        },
        exc_info=exc,
    )

    # LAYER 1: Infrastructure & Network Failures

    if isinstance(exc, APIConnectionError):
        raise AnalysisEngineError(
            "ERR_INFRA_CONNECTION_DROP",
            "Failed to connect to the AI model host. Network drop or connection closed.",
            retryable=True,
        ) from exc

    if isinstance(exc, Timeout):
        raise AnalysisEngineError(
            "ERR_INFRA_TIMEOUT",
            f"The request timed out during stage '{stage}'. "
            "The model host took too long to respond.",
            retryable=True,
        ) from exc

    if isinstance(exc, (ServiceUnavailableError, InternalServerError)):
        raise AnalysisEngineError(
            "ERR_INFRA_UPSTREAM_OUTAGE",
            "The AI service is temporarily unavailable or experiencing hardware/OOM issues.",
            retryable=True,
        ) from exc

    # LAYER 2: Rate, Quota & Lifecycle Failures

    if isinstance(exc, RateLimitError):
        raise AnalysisEngineError(
            "ERR_RATE_LIMIT_EXCEEDED",
            "Rate limit or token budget reached. Please wait before retrying.",
            retryable=True,
        ) from exc

    if isinstance(exc, (AuthenticationError, PermissionDeniedError)):
        raise AnalysisEngineError(
            "ERR_AUTH_KEY_INVALID",
            "Authentication failed. The API key or model permissions are invalid.",
            retryable=False,
        ) from exc

    if isinstance(exc, NotFoundError):
        raise AnalysisEngineError(
            "ERR_MODEL_DEPRECATED_OR_NOT_FOUND",
            "The requested LLM model version or endpoint was not found.",
            retryable=False,
        ) from exc

    # LAYER 3: Input & Context Engineering Failures

    if isinstance(exc, ContextWindowExceededError):
        raise AnalysisEngineError(
            "ERR_CONTEXT_WINDOW_OVERFLOW",
            "The document or total input tokens exceed the model's context length limit.",
            retryable=False,
        ) from exc

    if isinstance(exc, UnprocessableEntityError):
        raise AnalysisEngineError(
            "ERR_INPUT_EXTRACTION_ARTIFACT",
            "The input text contained unprocessable encoding or corrupted file artifacts.",
            retryable=False,
        ) from exc

    # LAYER 4: Non-Deterministic & Behavioral Model Failures

    if isinstance(exc, (InstructorRetryException, ValidationError)):
        detail_msg = str(exc)

        if isinstance(exc, InstructorRetryException) and exc.__cause__:
            detail_msg = f"Validation cause: {exc.__cause__}"

        raise AnalysisEngineError(
            "ERR_SCHEMA_DRIFT_VALIDATION",
            f"The AI service failed to produce valid structured output "
            f"for stage '{stage}': {detail_msg}",
            retryable=True,
        ) from exc

    # LAYER 6: Security & Data Governance Failures

    if isinstance(exc, ContentPolicyViolationError):
        raise AnalysisEngineError(
            "ERR_SECURITY_CONTENT_POLICY",
            "The input text or response was flagged by safety guards or content filters.",
            retryable=False,
        ) from exc

    # LAYER 5: Orchestration, State & Generic Fallbacks

    if isinstance(exc, BadRequestError):
        raise AnalysisEngineError(
            "ERR_BAD_REQUEST",
            f"The request sent during '{stage}' was invalid or malformed.",
            retryable=False,
        ) from exc

    if isinstance(exc, APIError):
        raise AnalysisEngineError(
            "ERR_PROVIDER_API_FAILURE",
            f"An unexpected upstream provider error occurred during stage '{stage}'.",
            retryable=True,
        ) from exc

    raise AnalysisEngineError(
        "ERR_PIPELINE_INTERNAL",
        f"An unhandled execution failure occurred during pipeline stage '{stage}'.",
        retryable=False,
    ) from exc


# ============================================================================
# BACKOFF / RETRY
# ============================================================================


P = ParamSpec("P")
T = TypeVar("T")

MAX_RATE_LIMIT_RETRIES = max(
    0,
    int(os.getenv("MAX_RATE_LIMIT_RETRIES", "3")),
)

BACKOFF_BASE_SECONDS = float(os.getenv("BACKOFF_BASE_SECONDS", "1.0"))

TRANSIENT_ERRORS = (
    RateLimitError,
    APIConnectionError,
    ServiceUnavailableError,
    InternalServerError,
    Timeout,
)


def _parse_retry_after(exc: Exception) -> float | None:
    """Extract Retry-After header from an exception response, if present."""
    response = getattr(exc, "response", None)

    if response is None or not hasattr(response, "headers"):
        return None

    headers = response.headers
    retry_after_str = headers.get("Retry-After") or headers.get("retry-after")

    if retry_after_str:
        try:
            return float(retry_after_str)
        except (ValueError, TypeError):
            return None

    return None


async def execute_with_backoff(
    api_call: Callable[P, Awaitable[T]],
    *args: P.args,
    **kwargs: P.kwargs,
) -> T:
    """Execute an LLM API call with adaptive exponential backoff
    for transient errors.
    """
    for attempt in range(MAX_RATE_LIMIT_RETRIES + 1):
        try:
            return await api_call(*args, **kwargs)

        except TRANSIENT_ERRORS as exc:
            if attempt == MAX_RATE_LIMIT_RETRIES:
                raise

            retry_after = _parse_retry_after(exc)

            if retry_after is not None:
                delay = retry_after

                logger.warning(
                    "Transient LLM failure (Retry-After header respected).",
                    extra={
                        "error_type": type(exc).__name__,
                        "delay_seconds": delay,
                        "attempt": attempt + 1,
                        "max_retries": MAX_RATE_LIMIT_RETRIES,
                    },
                )
            else:
                delay = (BACKOFF_BASE_SECONDS * (2**attempt)) + random.uniform(0, 1)

                logger.warning(
                    "Transient LLM failure. Retrying with exponential backoff.",
                    extra={
                        "error_type": type(exc).__name__,
                        "delay_seconds": delay,
                        "attempt": attempt + 1,
                        "max_retries": MAX_RATE_LIMIT_RETRIES,
                        "backoff_base": BACKOFF_BASE_SECONDS,
                    },
                )

            await asyncio.sleep(delay)

    raise RuntimeError("Unreachable")


# ============================================================================
# LLM RESPONSE SCHEMAS
# ============================================================================


class JDRequirementsResponse(BaseModel):
    model_config = ConfigDict(str_strip_whitespace=True)

    requirements: list[ExtractedRequirement] = Field(
        min_length=1,
        description="List of extracted requirements from the job description.",
    )


class ResumeClaimsResponse(BaseModel):
    model_config = ConfigDict(str_strip_whitespace=True)

    claims: list[ResumeClaim] = Field(
        min_length=1,
        description="List of extracted claims from the candidate's resume.",
    )


# ============================================================================
# ID MANAGEMENT
# ============================================================================


def require_id(entity_id: str | None, kind: str) -> str:
    """
    Narrows an Optional id to str.

    ExtractedRequirement.id / ResumeClaim.id are typed str | None because
    they start unset -- assign_ids() fills them in before this is ever called.

    Raising here turns "somehow assign_ids() was skipped" into an immediate,
    clear error instead of a None dict key silently causing a confusing
    downstream KeyError.
    """
    if entity_id is None:
        raise ValueError(f"{kind} id was never assigned — assign_ids() must run first")

    return entity_id


def assign_ids(
    jd_reqs: JDRequirementsResponse,
    res_claims: ResumeClaimsResponse,
) -> None:
    """
    Assigns real, unique ids to every extracted requirement and claim,
    mutating in place.

    These ids get embedded in the JSON sent to the LLM for scoring, and
    echoed back in MatchVerdict/SectionVerdict so verdicts can be correlated
    to the right requirement without relying on response order.
    """
    for i, req in enumerate(jd_reqs.requirements, start=1):
        req.id = f"req-{i}"

    for i, claim in enumerate(res_claims.claims, start=1):
        claim.id = f"claim-{i}"


# ============================================================================
# LLM RATE LIMITING / EXECUTION
# ============================================================================


R = TypeVar("R", bound=BaseModel)

MAX_CONCURRENT_LLM_REQUESTS = int(os.getenv("MAX_CONCURRENT_LLM_REQUESTS", "5"))

LLM_MAX_REQUESTS_PER_MINUTE = int(os.getenv("LLM_MAX_REQUESTS_PER_MINUTE", "4"))

MAX_VALIDATION_RETRIES = int(os.getenv("MAX_VALIDATION_RETRIES", "2"))

DEFAULT_TOKEN_COUNT = 0
DEFAULT_RETRY_COUNT = 0

RATE_LIMIT_WINDOW_SECONDS = 60.0


class LLMRateLimiter:
    """Limits LLM provider admission to a maximum number of attempts
    within a rolling window.
    """

    def __init__(
        self,
        max_requests: int,
        window_seconds: float = RATE_LIMIT_WINDOW_SECONDS,
    ) -> None:
        if max_requests <= 0:
            raise ValueError("LLM_MAX_REQUESTS_PER_MINUTE must be greater than 0")

        if window_seconds <= 0:
            raise ValueError("Rate-limit window must be greater than 0")

        self._max_requests = max_requests
        self._window_seconds = window_seconds
        self._request_times: deque[float] = deque()
        self._lock = asyncio.Lock()

    async def acquire(self) -> float:
        wait_start = time.perf_counter()

        while True:
            async with self._lock:
                now = time.monotonic()

                while self._request_times:
                    oldest_request = self._request_times[0]

                    if now - oldest_request >= self._window_seconds:
                        self._request_times.popleft()
                    else:
                        break

                if len(self._request_times) < self._max_requests:
                    self._request_times.append(now)
                    return time.perf_counter() - wait_start

                oldest_request = self._request_times[0]

                wait_seconds = self._window_seconds - (now - oldest_request)

            await asyncio.sleep(max(wait_seconds, 0.0))


llm_semaphore = asyncio.Semaphore(MAX_CONCURRENT_LLM_REQUESTS)

llm_rate_limiter = LLMRateLimiter(max_requests=LLM_MAX_REQUESTS_PER_MINUTE)


def build_usage(
    raw: Any,
    queue_wait_seconds: float,
) -> LLMUsage:
    """Normalize provider usage from a LiteLLM response."""

    usage = getattr(raw, "usage", None)

    def _extract_value(
        obj: Any,
        key: str,
        default: Any = None,
    ) -> Any:
        if obj is None:
            return default

        if isinstance(obj, dict):
            return obj.get(key, default)

        return getattr(obj, key, default)

    prompt_tokens = _extract_value(
        usage,
        "prompt_tokens",
    )

    completion_tokens = _extract_value(
        usage,
        "completion_tokens",
    )

    total_tokens = _extract_value(
        usage,
        "total_tokens",
    )

    completion_details = _extract_value(
        usage,
        "completion_tokens_details",
    )

    prompt_details = _extract_value(
        usage,
        "prompt_tokens_details",
    )

    reasoning_tokens = _extract_value(
        completion_details,
        "reasoning_tokens",
    )

    cached_input_tokens = _extract_value(
        prompt_details,
        "cached_tokens",
    )

    cache_creation_input_tokens = _extract_value(
        prompt_details,
        "cache_write_tokens",
    )

    cache_read_input_tokens = _extract_value(
        usage,
        "cache_read_input_tokens",
    )

    hidden_params = getattr(raw, "_hidden_params", {}) or {}

    retries = _extract_value(
        hidden_params,
        "retries",
        DEFAULT_RETRY_COUNT,
    )

    return {
        "input_tokens": (int(prompt_tokens) if prompt_tokens is not None else None),
        "output_tokens": (
            int(completion_tokens) if completion_tokens is not None else None
        ),
        "total_tokens": (int(total_tokens) if total_tokens is not None else None),
        "reasoning_tokens": (
            int(reasoning_tokens) if reasoning_tokens is not None else None
        ),
        "cached_input_tokens": (
            int(cached_input_tokens) if cached_input_tokens is not None else None
        ),
        "cache_creation_input_tokens": (
            int(cache_creation_input_tokens)
            if cache_creation_input_tokens is not None
            else None
        ),
        "cache_read_input_tokens": (
            int(cache_read_input_tokens)
            if cache_read_input_tokens is not None
            else None
        ),
        "queue_wait_seconds": float(queue_wait_seconds),
        "retries": int(retries) if retries is not None else 0,
    }


async def call_llm(
    response_model: type[R],
    messages: list[dict[str, str]],
    client: Any,
    model: str,
    api_key: str,
    stage: str,
    context: dict[str, Any] | None = None,
) -> tuple[R, LLMUsage]:
    """Centralized LLM execution with rate limiting, backoff,
    and exception mapping.
    """
    total_queue_wait_seconds = 0.0
    provider_attempt = 0

    async def _guarded_call() -> Any:
        nonlocal total_queue_wait_seconds
        nonlocal provider_attempt

        provider_attempt += 1

        rate_wait_seconds = await llm_rate_limiter.acquire()
        total_queue_wait_seconds += rate_wait_seconds

        semaphore_wait_start = time.perf_counter()

        async with llm_semaphore:
            total_queue_wait_seconds += time.perf_counter() - semaphore_wait_start

            logger.info(
                "LLM provider attempt started",
                extra={
                    "stage": stage,
                    "provider_attempt": provider_attempt,
                    "model": model,
                },
            )

            try:
                result = await client.chat.completions.create_with_completion(
                    api_key=api_key,
                    model=model,
                    messages=messages,
                    response_model=response_model,
                    max_retries=MAX_VALIDATION_RETRIES,
                    context=context,
                )

                logger.info(
                    "LLM provider attempt completed",
                    extra={
                        "stage": stage,
                        "provider_attempt": provider_attempt,
                        "model": model,
                    },
                )

                return result

            except Exception as exc:
                logger.warning(
                    "LLM provider attempt failed, evaluating retries",
                    extra={
                        "stage": stage,
                        "provider_attempt": provider_attempt,
                        "model": model,
                        "error_type": type(exc).__name__,
                        "error": str(exc),
                    },
                )
                raise

    try:
        resp, raw = await execute_with_backoff(_guarded_call)

    except Exception as exc:  # noqa: BLE001
        handle_llm_exception(exc, stage)

    usage = build_usage(
        raw,
        total_queue_wait_seconds,
    )

    return resp, usage


# ============================================================================
# EXTRACTION
# ============================================================================


@track_llm_cost("jd_extraction")
async def extract_jd_requirements(
    client: Any,
    llm_model: str,
    llm_api_key: str,
    domain: DomainStrategy,
    jd_text: str,
) -> tuple[JDRequirementsResponse, LLMUsage]:
    """
    Extracts structured requirements from job description text using the LLM.
    """
    cleaned_jd = jd_text.strip() if jd_text else ""

    if not cleaned_jd:
        raise AnalysisEngineError(
            "ERR_BAD_REQUEST",
            "Job description text cannot be empty or whitespace.",
            retryable=False,
        )

    messages = [
        {
            "role": "system",
            "content": domain.jd_extraction_prompt(),
        },
        {
            "role": "user",
            "content": cleaned_jd,
        },
    ]

    return await call_llm(
        response_model=JDRequirementsResponse,
        messages=messages,
        client=client,
        model=llm_model,
        api_key=llm_api_key,
        stage="JD Extraction",
        context={
            "source_texts": {
                "jd": cleaned_jd,
            },
            "valid_sections": domain.section_taxonomy(),
        },
    )


@track_llm_cost("resume_extraction")
async def extract_resume_claims(
    client: Any,
    llm_model: str,
    llm_api_key: str,
    domain: DomainStrategy,
    resume_text: str,
) -> tuple[ResumeClaimsResponse, LLMUsage]:
    """
    Extracts structured claims from resume text using the LLM.
    """
    cleaned_resume = resume_text.strip() if resume_text else ""

    if not cleaned_resume:
        raise AnalysisEngineError(
            "ERR_BAD_REQUEST",
            "Resume text cannot be empty or whitespace.",
            retryable=False,
        )

    messages = [
        {
            "role": "system",
            "content": domain.resume_extraction_prompt(),
        },
        {
            "role": "user",
            "content": cleaned_resume,
        },
    ]

    return await call_llm(
        response_model=ResumeClaimsResponse,
        messages=messages,
        client=client,
        model=llm_model,
        api_key=llm_api_key,
        stage="Resume Extraction",
        context={
            "source_texts": {
                "resume": cleaned_resume,
            },
            "valid_sections": domain.section_taxonomy(),
        },
    )


# ============================================================================
# SCORING
# ============================================================================


class ScoringResultShape(Protocol):
    """
    Structural shape every domain's get_scoring_schema() must produce.

    get_scoring_schema() is typed as -> type[BaseModel] because each
    domain builds its own class dynamically (see TechDomain), so the
    type checker has no static way to know match_verdicts /
    section_verdicts / complete_analysis exist on the result.
    """

    complete_analysis: str
    match_verdicts: list[MatchVerdict]
    section_verdicts: list[SectionVerdict]


S = TypeVar("S", bound=BaseModel)


@track_llm_cost("scoring")
async def score_claims(
    client: Any,
    llm_model: str,
    llm_api_key: str,
    domain: DomainStrategy,
    jd_reqs: JDRequirementsResponse,
    res_claims: ResumeClaimsResponse,
    response_model: type[S],
) -> tuple[S, LLMUsage]:
    """
    Orchestrates the LLM to produce verdicts scoring the candidate's fit
    against JD requirements and resume claims.

    response_model here is the domain's scoring schema (verdicts only),
    not the fully assembled final result.
    """
    scoring_input = (
        f"JD Requirements:\n"
        f"{jd_reqs.model_dump_json(exclude_none=True)}\n\n"
        f"Resume Claims:\n"
        f"{res_claims.model_dump_json(exclude_none=True)}"
    )

    messages = [
        {
            "role": "system",
            "content": domain.scoring_prompt(),
        },
        {
            "role": "user",
            "content": scoring_input,
        },
    ]

    return await call_llm(
        response_model=response_model,
        messages=messages,
        client=client,
        model=llm_model,
        api_key=llm_api_key,
        stage="Claims Scoring",
        context={
            "valid_sections": domain.section_taxonomy(),
        },
    )


# ============================================================================
# FINAL RESULT ASSEMBLY
# ============================================================================


def assemble_final_result(
    domain: DomainStrategy,
    jd_reqs: JDRequirementsResponse,
    res_claims: ResumeClaimsResponse,
    scoring_result: BaseModel,
) -> BaseModel:
    """
    Merges the LLM's verdicts (match strength, supporting claim ids,
    section scores/reviews) with the already-known, already-verified
    requirement/claim data into the domain's real final schema.

    Raises clearly if match_verdicts don't exactly cover the given
    requirement ids (missing, duplicate, or unrecognized), rather than
    silently dropping or misassigning one.
    """
    result = cast(
        ScoringResultShape,
        scoring_result,
    )

    requirements_by_id = {
        require_id(req.id, "requirement"): req for req in jd_reqs.requirements
    }

    claims_by_id = {require_id(claim.id, "claim"): claim for claim in res_claims.claims}

    match_verdicts = result.match_verdicts
    verdict_ids = [v.id for v in match_verdicts]

    if len(verdict_ids) != len(set(verdict_ids)):
        raise AnalysisEngineError(
            "ERR_ASSEMBLY_MISMATCH",
            f"Duplicate requirement IDs found in match verdicts: {verdict_ids}",
            retryable=False,
        )

    if set(verdict_ids) != set(requirements_by_id):
        missing = set(requirements_by_id) - set(verdict_ids)
        extra = set(verdict_ids) - set(requirements_by_id)

        raise AnalysisEngineError(
            "ERR_ASSEMBLY_MISMATCH",
            "Match verdicts must cover exactly the extracted "
            "requirement IDs. "
            f"Missing: {missing or None}, "
            f"Unexpected: {extra or None}",
            retryable=False,
        )

    requirement_matches = [
        RequirementMatch.from_verdict(
            requirements_by_id[verdict.id],
            verdict,
            claims_by_id,
        )
        for verdict in match_verdicts
    ]

    matches_by_section: dict[str, list[RequirementMatch]] = {}

    for match in requirement_matches:
        original_section = requirements_by_id[match.id].section

        matches_by_section.setdefault(
            original_section,
            [],
        ).append(match)

    seen_sections: set[str] = set()
    section_scores: list[SectionScore] = []

    for sv in result.section_verdicts:
        seen_sections.add(sv.id)

        section_scores.append(
            SectionScore(
                id=sv.id,
                label=sv.id.replace("_", " ").title(),
                score=sv.score,
                review=sv.review,
                requirements=matches_by_section.get(
                    sv.id,
                    [],
                ),
            )
        )

    for section_id, matches in matches_by_section.items():
        if section_id not in seen_sections:
            section_scores.append(
                SectionScore(
                    id=section_id,
                    label=section_id.replace("_", " ").title(),
                    score=0,
                    review=("Section evaluation was omitted by scoring model."),
                    requirements=matches,
                )
            )

    final_schema = domain.get_final_schema()

    return final_schema(
        complete_analysis=result.complete_analysis,
        sections=section_scores,
    )


# ============================================================================
# PIPELINE CONFIGURATION
# ============================================================================


@functools.lru_cache(maxsize=1)
def get_domain() -> DomainStrategy:
    """
    Cached so domain initialization and taxonomy validation run once.
    """
    return TechDomain()


@functools.lru_cache(maxsize=1)
def get_client() -> Any:
    """Cached instructor-wrapped LiteLLM client."""
    return instructor.from_litellm(
        acompletion,
        mode=instructor.Mode.JSON,
    )


def _get_llm_config() -> tuple[str, str]:
    """Retrieves LLM model and API key dynamically at runtime."""
    api_key = os.getenv("LLM_API_KEY", "")
    model = os.getenv("LLM_MODEL", "")

    return model, api_key


# ============================================================================
# MAIN PIPELINE
# ============================================================================


async def analyze_fit(
    jd_text: str,
    resume_text: str,
) -> tuple[BaseModel, AggregatedLLMUsage]:
    """
    Orchestrates the full analysis pipeline:

    1. Resolve domain and LLM client.
    2. Extract JD requirements and resume claims in parallel.
    3. Assign stable IDs.
    4. Score resume claims against JD requirements.
    5. Assemble the final domain-specific result.
    6. Aggregate LLM usage.
    """
    domain: DomainStrategy = get_domain()
    client = get_client()

    llm_model, llm_api_key = _get_llm_config()

    logger.debug("Starting parallel extraction for JD and Resume...")

    (
        (jd_reqs, jd_usage),
        (
            res_claims,
            res_usage,
        ),
    ) = await asyncio.gather(
        extract_jd_requirements(
            client,
            llm_model,
            llm_api_key,
            domain,
            jd_text,
        ),
        extract_resume_claims(
            client,
            llm_model,
            llm_api_key,
            domain,
            resume_text,
        ),
    )

    assign_ids(
        jd_reqs,
        res_claims,
    )

    scoring_schema = domain.get_scoring_schema()

    scoring_result, score_usage = await score_claims(
        client,
        llm_model,
        llm_api_key,
        domain,
        jd_reqs,
        res_claims,
        response_model=scoring_schema,
    )

    final_analysis = assemble_final_result(
        domain,
        jd_reqs,
        res_claims,
        scoring_result,
    )

    aggregated = aggregate_llm_usage(
        [
            jd_usage,
            res_usage,
            score_usage,
        ]
    )

    return final_analysis, aggregated
