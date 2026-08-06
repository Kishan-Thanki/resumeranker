# ============================================================================
# IMPORTS
# ============================================================================
import asyncio
import os
import random
import time
from collections.abc import Awaitable, Callable
from typing import Any, NoReturn, ParamSpec, TypedDict, TypeVar

import instructor
from instructor.core import InstructorRetryException
from litellm import acompletion
from litellm.exceptions import (
    APIConnectionError,
    APIError,
    ContextWindowExceededError,
    RateLimitError,
    Timeout,
)
from pydantic import BaseModel, ValidationError

from app.domain.base import DomainStrategy
from app.domain.tech import TechDomain
from app.logger import logger
from app.schemas import ExtractedRequirement, ResumeClaim
from app.utils.telemetry import track_llm_cost

# ============================================================================
# CONSTANTS & ENVIRONMENT
# ============================================================================
MAX_RATE_LIMIT_RETRIES = 5
BACKOFF_BASE_SECONDS = 2

DEFAULT_TOKEN_COUNT = 0
DEFAULT_RETRY_COUNT = 0
DEFAULT_QUEUE_WAIT_SECONDS = 0.0

LLM_API_KEY = os.environ["LLM_API_KEY"]
LLM_MODEL = os.environ["LLM_MODEL"]
MAX_CONCURRENT_LLM_REQUESTS = int(os.environ["MAX_CONCURRENT_LLM_REQUESTS"])

llm_semaphore = asyncio.Semaphore(MAX_CONCURRENT_LLM_REQUESTS)

# ============================================================================
# TYPE ALIASES & UTILITY TYPES
# ============================================================================
P = ParamSpec("P")
T = TypeVar("T")
R = TypeVar("R", bound=BaseModel)
S = TypeVar("S", bound=BaseModel)


class LLMUsage(TypedDict):
    """Structured usage statistics for a single LLM call."""

    prompt_tokens: int
    completion_tokens: int
    queue_wait_seconds: float
    retries: int


class AggregatedLLMUsage(TypedDict):
    """Aggregated usage statistics across multiple LLM calls."""

    prompt_tokens: int
    completion_tokens: int
    total_tokens: int
    queue_wait_seconds: float
    retries: int


# ============================================================================
# CUSTOM EXCEPTIONS
# ============================================================================
class AnalysisEngineError(Exception):
    """Base exception for analysis engine."""

    def __init__(self, code: str, message: str):
        self.code = code
        self.message = message
        super().__init__(self.message)


# ============================================================================
# LLM EXCEPTION HANDLING & BACKOFF
# ============================================================================
def _handle_llm_exception(exc: Exception, stage: str) -> NoReturn:
    """
    Translates LLM-specific exceptions into domain-specific AnalysisEngineErrors.
    """
    logger.exception("LLM exception", extra={"stage": stage})
    if isinstance(exc, RateLimitError):
        raise AnalysisEngineError(
            "ERR_RATE_LIMIT", "The AI service is currently experiencing high load."
        )
    if isinstance(exc, APIConnectionError):
        raise AnalysisEngineError(
            "ERR_API_CONNECTION", "Failed to connect to the AI service provider."
        )
    if isinstance(exc, ContextWindowExceededError):
        raise AnalysisEngineError(
            "ERR_CONTEXT_EXCEEDED", "The provided documents are too large."
        )
    if isinstance(exc, Timeout):
        raise AnalysisEngineError("ERR_LLM_TIMEOUT", "The AI service timed out.")
    if isinstance(exc, (InstructorRetryException, ValidationError)):
        raise AnalysisEngineError(
            "ERR_LLM_VALIDATION", "The AI service returned an invalid format."
        )
    if isinstance(exc, APIError):
        raise AnalysisEngineError(
            "ERR_AI_PROVIDER", "The AI service provider returned an error."
        )
    raise AnalysisEngineError(
        "ERR_LLM_INTERNAL", f"An internal error occurred during {stage}."
    )


def _parse_retry_after(exc: RateLimitError) -> int | None:
    """Extract Retry-After header from a RateLimitError, if present."""
    if getattr(exc, "response", None) and hasattr(exc.response, "headers"):
        retry_after_str = exc.response.headers.get("Retry-After")
        if retry_after_str and retry_after_str.isdigit():
            return int(retry_after_str)
    return None


async def _execute_with_backoff(
    api_call: Callable[P, Awaitable[T]],
    *args: P.args,
    **kwargs: P.kwargs,
) -> T:
    """
    Executes an LLM API call with adaptive exponential backoff and jitter.
    Respects 'Retry-After' headers if provided by the LLM API.
    """
    for attempt in range(MAX_RATE_LIMIT_RETRIES + 1):
        try:
            return await api_call(*args, **kwargs)
        except RateLimitError as exc:
            if attempt == MAX_RATE_LIMIT_RETRIES:
                raise

            retry_after = _parse_retry_after(exc)
            if retry_after is not None:
                delay = retry_after
                logger.warning(
                    "Rate limited. Provider requested wait.",
                    extra={
                        "delay_seconds": delay,
                        "attempt": attempt + 1,
                        "max_retries": MAX_RATE_LIMIT_RETRIES,
                        "retry_after": retry_after,
                    },
                )
            else:
                delay = BACKOFF_BASE_SECONDS**attempt + random.uniform(0, 1)
                logger.warning(
                    "Rate limited. Exponential backoff.",
                    extra={
                        "delay_seconds": delay,
                        "attempt": attempt + 1,
                        "max_retries": MAX_RATE_LIMIT_RETRIES,
                        "backoff_base": BACKOFF_BASE_SECONDS,
                    },
                )
            await asyncio.sleep(delay)
    raise RuntimeError("Unreachable") 


def _build_usage(raw: Any, queue_wait_seconds: float) -> LLMUsage:
    """Constructs an LLMUsage dict from the raw LiteLLM response."""
    return {
        "prompt_tokens": getattr(raw.usage, "prompt_tokens", DEFAULT_TOKEN_COUNT),
        "completion_tokens": getattr(raw.usage, "completion_tokens", DEFAULT_TOKEN_COUNT),
        "queue_wait_seconds": queue_wait_seconds,
        "retries": getattr(raw, "_hidden_params", {}).get("retries", DEFAULT_RETRY_COUNT)
        if hasattr(raw, "_hidden_params")
        else DEFAULT_RETRY_COUNT,
    }


async def _call_llm(
    response_model: type[R],
    messages: list[dict[str, str]],
    client: Any,
    model: str,
    api_key: str,
    stage: str,
) -> tuple[R, LLMUsage]:
    """
    Centralised LLM call with concurrency limiting, backoff, and usage tracking.
    Returns the response model instance and usage stats.
    """
    wait_start = time.perf_counter()
    async with llm_semaphore:
        queue_wait_seconds = time.perf_counter() - wait_start
        try:
            resp, raw = await _execute_with_backoff(
                client.chat.completions.create_with_completion,
                api_key=api_key,
                model=model,
                messages=messages,
                response_model=response_model,
                max_retries=0,  
            )
        except Exception as exc:
            _handle_llm_exception(exc, stage)

    usage = _build_usage(raw, queue_wait_seconds)
    return resp, usage


# ============================================================================
# RESPONSE MODELS (Pydantic)
# ============================================================================
class JDRequirementsResponse(BaseModel):
    requirements: list[ExtractedRequirement]


class ResumeClaimsResponse(BaseModel):
    claims: list[ResumeClaim]


# ============================================================================
# EXTRACTION FUNCTIONS
# ============================================================================
@track_llm_cost("jd_extraction")
async def _extract_jd_requirements(
    client: Any,
    llm_model: str,
    llm_api_key: str,
    domain: DomainStrategy,
    jd_text: str,
) -> tuple[JDRequirementsResponse, LLMUsage]:
    """
    Extracts structured requirements from job description text using the LLM.
    """
    messages = [
        {"role": "system", "content": domain.jd_extraction_prompt()},
        {"role": "user", "content": jd_text},
    ]
    return await _call_llm(
        response_model=JDRequirementsResponse,
        messages=messages,
        client=client,
        model=llm_model,
        api_key=llm_api_key,
        stage="JD Extraction",
    )


@track_llm_cost("resume_extraction")
async def _extract_resume_claims(
    client: Any,
    llm_model: str,
    llm_api_key: str,
    domain: DomainStrategy,
    resume_text: str,
) -> tuple[ResumeClaimsResponse, LLMUsage]:
    """
    Extracts structured claims from resume text using the LLM.
    """
    messages = [
        {"role": "system", "content": domain.resume_extraction_prompt()},
        {"role": "user", "content": resume_text},
    ]
    return await _call_llm(
        response_model=ResumeClaimsResponse,
        messages=messages,
        client=client,
        model=llm_model,
        api_key=llm_api_key,
        stage="Resume Extraction",
    )


# ============================================================================
# SCORING FUNCTION
# ============================================================================
@track_llm_cost("scoring")
async def _score_claims(
    client: Any,
    llm_model: str,
    llm_api_key: str,
    domain: DomainStrategy,
    jd_reqs: JDRequirementsResponse,
    res_claims: ResumeClaimsResponse,
    response_model: type[S],
) -> tuple[S, LLMUsage]:
    """
    Orchestrates the LLM to score the candidate's fit based on JD requirements and resume claims.

    Note: The `response_model` must be obtained from the same `domain` to ensure consistency
    between the scoring prompt and the expected output schema.
    """
    scoring_input = (
        f"JD Requirements:\n"
        f"{jd_reqs.model_dump_json(indent=2)}\n\n"
        f"Resume Claims:\n"
        f"{res_claims.model_dump_json(indent=2)}"
    )
    messages = [
        {"role": "system", "content": domain.scoring_prompt()},
        {"role": "user", "content": scoring_input},
    ]
    return await _call_llm(
        response_model=response_model,
        messages=messages,
        client=client,
        model=llm_model,
        api_key=llm_api_key,
        stage="Claims Scoring",
    )


# ============================================================================
# AGGREGATION HELPER
# ============================================================================
def _aggregate_usage(*usages: LLMUsage) -> AggregatedLLMUsage:
    """Sums multiple LLMUsage dicts into a single AggregatedLLMUsage."""
    total_prompt = sum(u["prompt_tokens"] for u in usages)
    total_completion = sum(u["completion_tokens"] for u in usages)
    total_queue = sum(u["queue_wait_seconds"] for u in usages)
    total_retries = sum(u["retries"] for u in usages)
    return {
        "prompt_tokens": total_prompt,
        "completion_tokens": total_completion,
        "total_tokens": total_prompt + total_completion,
        "queue_wait_seconds": total_queue,
        "retries": total_retries,
    }


# ============================================================================
# PUBLIC API
# ============================================================================
async def analyze_fit(
    jd_text: str, resume_text: str
) -> tuple[BaseModel, AggregatedLLMUsage]:
    """
    Orchestrates the full analysis pipeline: parallel extraction of JD/Resume entities
    followed by sequential scoring of the fit.

    Returns:
        - The final analysis result (Pydantic model from the domain).
        - A dictionary with aggregated token usage, queue wait time, and retry counts.
    """
    domain: DomainStrategy = TechDomain()
    client = instructor.from_litellm(acompletion, mode=instructor.Mode.JSON)

    logger.debug("Starting parallel extraction for JD and Resume...")
    jd_task = _extract_jd_requirements(client, LLM_MODEL, LLM_API_KEY, domain, jd_text)
    res_task = _extract_resume_claims(
        client, LLM_MODEL, LLM_API_KEY, domain, resume_text
    )
    (jd_reqs, jd_usage), (res_claims, res_usage) = await asyncio.gather(
        jd_task, res_task
    )

    final_schema = domain.get_final_schema()
    final_analysis, score_usage = await _score_claims(
        client,
        LLM_MODEL,
        LLM_API_KEY,
        domain,
        jd_reqs,
        res_claims,
        response_model=final_schema,
    )

    aggregated = _aggregate_usage(jd_usage, res_usage, score_usage)
    return final_analysis, aggregated