"""LLM service. Single point of contact with the model.

Stub-mode short-circuit: when LLM_API_KEY isn't set to a real value,
returns canned fixtures without making any network calls (see llm_stub).

When a real key is set, dispatches to whichever provider `LLM_MODEL`
points at (Gemini, Anthropic, OpenAI, etc.) via litellm. We pass the
key explicitly with the `api_key` kwarg rather than letting litellm
auto-discover from per-provider env vars — keeps our config to a single
`LLM_API_KEY` regardless of which provider we choose.

`instructor` wraps every call for typed/validated/retried structured
output. Each call is also cached in Redis by (model, prompt_version,
input) and recorded for cost accounting per analysis.
"""

import asyncio
import logging
import random
from typing import Protocol

from pydantic import BaseModel, Field, TypeAdapter

from app.config import settings
from app.domain.base import Domain
from app.schemas.analysis import Evidence, SectionId, SectionScore
from app.services import llm_cache, llm_cost, llm_stub

logger = logging.getLogger(__name__)


# Stub-mode artificial latency. Real Anthropic Haiku calls take ~1–3 s per
# step; stub responses are sub-millisecond. Without this padding, an
# observer timing requests could trivially detect "this server is in stub
# mode" (and therefore that the response content is canned). Sleeping a
# random duration in the same band makes the two modes indistinguishable
# from outside.
_STUB_MIN_DELAY = 1.0
_STUB_MAX_DELAY = 2.5


async def _stub_latency() -> None:
    await asyncio.sleep(random.uniform(_STUB_MIN_DELAY, _STUB_MAX_DELAY))


class LLMError(Exception):
    """Raised when the LLM provider call fails after retries."""

def _format_llm_error(phase: str, exc: Exception) -> str:
    """Map raw provider errors into human-friendly messages for the UI."""
    exc_name = type(exc).__name__
    if "ServiceUnavailable" in exc_name or "503" in str(exc):
        return f"{phase} failed: The AI model is currently experiencing high demand. Please try again in a few moments."
    if "RateLimit" in exc_name or "429" in str(exc):
        return f"{phase} failed: We're sending too many requests to the AI right now. Please wait a bit and try again."
    if "Timeout" in exc_name:
        return f"{phase} failed: The AI model took too long to respond. Please try again."
    if "APIConnection" in exc_name:
        return f"{phase} failed: Could not connect to the AI service. Please try again."
    
    # If it's an unrecognized error, show a generic message to avoid leaking raw JSON/stacktraces
    # but still log the raw error for debugging.
    logger.error("Raw LLM error in %s: %s", phase, exc, exc_info=exc)
    return f"{phase} failed: An unexpected error occurred with the AI provider."


# --- Internal types (NOT in the wire contract) -------------------------------


class ExtractedRequirement(BaseModel):
    id: str = Field(min_length=1)
    section: SectionId
    requirement: str
    jd_evidence: Evidence


class ResumeClaim(BaseModel):
    id: str = Field(min_length=1)
    section: SectionId
    claim: str
    resume_evidence: Evidence


_REQ_LIST = TypeAdapter(list[ExtractedRequirement])
_CLAIM_LIST = TypeAdapter(list[ResumeClaim])
_SECTION_LIST = TypeAdapter(list[SectionScore])


# --- LLM client (lazy; only created when not in stub mode) -------------------


class _Completions(Protocol):
    async def create(self, **kwargs: object) -> object: ...


_client: object | None = None


def _get_client() -> object:
    global _client
    if _client is None:
        import instructor
        from litellm import acompletion

        # `Mode.JSON` works across providers (Gemini, Anthropic, OpenAI),
        # whereas the default `Mode.TOOLS` (OpenAI-style function-calling)
        # makes Gemini return empty `choices` arrays with v2.5 models —
        # the model produces valid output but instructor's TOOLS-mode
        # parser can't recognize it. JSON mode asks the model to return a
        # JSON string matching the Pydantic schema, which every provider
        # litellm supports handles cleanly.
        _client = instructor.from_litellm(acompletion, mode=instructor.Mode.JSON)
    return _client


# --- Cost helpers ------------------------------------------------------------


async def _record_call_from_response(analysis_id: str | None, model: str, raw: object) -> None:
    """Best-effort token extraction from a litellm/openai-style response."""
    if analysis_id is None:
        return
    prompt_tokens = 0
    completion_tokens = 0
    usage = getattr(raw, "usage", None)
    if usage is not None:
        prompt_tokens = int(getattr(usage, "prompt_tokens", 0) or 0)
        completion_tokens = int(getattr(usage, "completion_tokens", 0) or 0)
    await llm_cost.record_call(analysis_id, model, prompt_tokens, completion_tokens)


# --- Public pipeline functions ----------------------------------------------


async def extract_jd_requirements(
    jd_text: str,
    domain: Domain,
    *,
    analysis_id: str | None = None,
) -> list[ExtractedRequirement]:
    if settings.llm_stub_mode:
        await _stub_latency()
        sections = llm_stub.stub_sections(jd_text, "")
        result: list[ExtractedRequirement] = []
        for s in sections:
            for r in s.requirements:
                result.append(
                    ExtractedRequirement(
                        id=r.id, section=s.id, requirement=r.requirement, jd_evidence=r.jd_evidence
                    )
                )
        return result

    cached = await llm_cache.get_cached("jd", settings.llm_model, domain.prompt_version, jd_text)
    if cached is not None:
        logger.info("cache hit: jd extraction")
        return _REQ_LIST.validate_python(cached)

    client = _get_client()
    try:
        response = await client.chat.completions.create(  # type: ignore[attr-defined]
            model=settings.llm_model,
            api_key=settings.llm_api_key,
            response_model=list[ExtractedRequirement],
            messages=[
                {"role": "system", "content": domain.jd_extraction_prompt()},
                {"role": "user", "content": jd_text},
            ],
            max_retries=2,
        )
    except Exception as exc:
        raise LLMError(_format_llm_error("JD extraction", exc)) from exc

    await _record_call_from_response(analysis_id, settings.llm_model, response)
    payload = [r.model_dump(mode="json") for r in response]
    await llm_cache.set_cached("jd", settings.llm_model, domain.prompt_version, jd_text, payload)
    return list(response)


async def extract_resume_claims(
    resume_text: str,
    domain: Domain,
    *,
    analysis_id: str | None = None,
) -> list[ResumeClaim]:
    if settings.llm_stub_mode:
        await _stub_latency()
        sections = llm_stub.stub_sections("", resume_text)
        result: list[ResumeClaim] = []
        for s in sections:
            for r in s.requirements:
                for ev in r.resume_evidence:
                    result.append(
                        ResumeClaim(
                            id=f"c-{r.id}",
                            section=s.id,
                            claim=r.requirement,
                            resume_evidence=ev,
                        )
                    )
        return result

    cached = await llm_cache.get_cached(
        "resume", settings.llm_model, domain.prompt_version, resume_text
    )
    if cached is not None:
        logger.info("cache hit: resume extraction")
        return _CLAIM_LIST.validate_python(cached)

    client = _get_client()
    try:
        response = await client.chat.completions.create(  # type: ignore[attr-defined]
            model=settings.llm_model,
            api_key=settings.llm_api_key,
            response_model=list[ResumeClaim],
            messages=[
                {"role": "system", "content": domain.resume_extraction_prompt()},
                {"role": "user", "content": resume_text},
            ],
            max_retries=2,
        )
    except Exception as exc:
        raise LLMError(_format_llm_error("Resume extraction", exc)) from exc

    await _record_call_from_response(analysis_id, settings.llm_model, response)
    payload = [r.model_dump(mode="json") for r in response]
    await llm_cache.set_cached(
        "resume", settings.llm_model, domain.prompt_version, resume_text, payload
    )
    return list(response)


async def score_requirements_against_claims(
    requirements: list[ExtractedRequirement],
    claims: list[ResumeClaim],
    domain: Domain,
    *,
    jd_text: str,
    resume_text: str,
    analysis_id: str | None = None,
) -> list[SectionScore]:
    if settings.llm_stub_mode:
        await _stub_latency()
        return llm_stub.stub_sections(jd_text, resume_text)

    # Scoring cache key includes both texts (different resume → different scoring
    # even with the same JD requirements).
    cache_input = f"{jd_text}\n--RESUME--\n{resume_text}"
    cached = await llm_cache.get_cached(
        "score", settings.llm_model, domain.prompt_version, cache_input
    )
    if cached is not None:
        logger.info("cache hit: scoring")
        return _SECTION_LIST.validate_python(cached)

    client = _get_client()
    payload_in = {
        "requirements": [r.model_dump() for r in requirements],
        "claims": [c.model_dump() for c in claims],
        "section_weights": domain.section_weights(),
    }
    try:
        response = await client.chat.completions.create(  # type: ignore[attr-defined]
            model=settings.llm_model,
            api_key=settings.llm_api_key,
            response_model=list[SectionScore],
            messages=[
                {"role": "system", "content": domain.scoring_prompt()},
                {"role": "user", "content": str(payload_in)},
            ],
            max_retries=2,
        )
    except Exception as exc:
        raise LLMError(_format_llm_error("Scoring", exc)) from exc

    await _record_call_from_response(analysis_id, settings.llm_model, response)
    payload_out = [s.model_dump(mode="json") for s in response]
    await llm_cache.set_cached(
        "score", settings.llm_model, domain.prompt_version, cache_input, payload_out
    )
    return list(response)
