import os
import asyncio
import instructor
from litellm import acompletion
from litellm.exceptions import RateLimitError, APIConnectionError, APIError, ContextWindowExceededError, Timeout
from instructor.core import InstructorRetryException
from pydantic import ValidationError, BaseModel

from app.schemas import (
    ExtractedRequirement,
    ResumeClaim
)
from app.domain.base import DomainStrategy
from app.domain.tech import TechDomain
from app.logger import logger

class AnalysisEngineError(Exception):
    """Base exception for analysis engine."""
    def __init__(self, code: str, message: str):
        self.code = code
        self.message = message
        super().__init__(self.message)

class JDRequirementsResponse(BaseModel):
    requirements: list[ExtractedRequirement]

class ResumeClaimsResponse(BaseModel):
    claims: list[ResumeClaim]

MAX_CONCURRENT = int(os.environ.get("MAX_CONCURRENT_LLM_REQUESTS", 3))
llm_bouncer = asyncio.Semaphore(MAX_CONCURRENT)

def _handle_llm_exception(e: Exception, stage: str):
    logger.exception(f"LLM exception during {stage}")
    if isinstance(e, RateLimitError):
        raise AnalysisEngineError("ERR_RATE_LIMIT", "The AI service is currently experiencing high load. Please try again later.")
    elif isinstance(e, APIConnectionError):
        raise AnalysisEngineError("ERR_API_CONNECTION", "Failed to connect to the AI service provider.")
    elif isinstance(e, ContextWindowExceededError):
        raise AnalysisEngineError("ERR_CONTEXT_EXCEEDED", "The provided documents are too large for the AI to process.")
    elif isinstance(e, Timeout):
        raise AnalysisEngineError("ERR_LLM_TIMEOUT", "The AI service took too long to respond.")
    elif isinstance(e, (InstructorRetryException, ValidationError)):
        raise AnalysisEngineError("ERR_LLM_VALIDATION", "The AI service returned an invalid format.")
    elif isinstance(e, APIError):
        raise AnalysisEngineError("ERR_AI_PROVIDER", "The AI service provider returned an error.")
    else:
        raise AnalysisEngineError("ERR_LLM_INTERNAL", f"An internal error occurred during {stage}.")

async def _extract_jd_requirements(client, model: str, api_key: str, domain: DomainStrategy, jd_text: str) -> tuple[JDRequirementsResponse, dict]:
    msg = [
        {"role": "system", "content": domain.jd_extraction_prompt()},
        {"role": "user", "content": jd_text}
    ]
    
    async with llm_bouncer:
        try:
            resp, raw = await client.chat.completions.create_with_completion(
                api_key=api_key,
                model=model,
                messages=msg,
                response_model=JDRequirementsResponse,
                max_retries=1,
            )
        except Exception as e:
            _handle_llm_exception(e, "JD Extraction")
        
    usage = {
        "prompt_tokens": getattr(raw.usage, "prompt_tokens", 0),
        "completion_tokens": getattr(raw.usage, "completion_tokens", 0)
    }
    return resp, usage

async def _extract_resume_claims(client, model: str, api_key: str, domain: DomainStrategy, resume_text: str) -> tuple[ResumeClaimsResponse, dict]:
    msg = [
        {"role": "system", "content": domain.resume_extraction_prompt()},
        {"role": "user", "content": resume_text}
    ]
    
    async with llm_bouncer:
        try:
            resp, raw = await client.chat.completions.create_with_completion(
                api_key=api_key,
                model=model,
                messages=msg,
                response_model=ResumeClaimsResponse,
                max_retries=1,
            )
        except Exception as e:
            _handle_llm_exception(e, "Resume Extraction")
        
    usage = {
        "prompt_tokens": getattr(raw.usage, "prompt_tokens", 0),
        "completion_tokens": getattr(raw.usage, "completion_tokens", 0)
    }
    return resp, usage

async def _score_claims(client, model: str, api_key: str, domain: DomainStrategy, jd_reqs: JDRequirementsResponse, res_claims: ResumeClaimsResponse) -> tuple[BaseModel, dict]:
    user_prompt = f"JD Requirements:\n{jd_reqs.model_dump_json(indent=2)}\n\nResume Claims:\n{res_claims.model_dump_json(indent=2)}"
    msg = [
        {"role": "system", "content": domain.scoring_prompt()},
        {"role": "user", "content": user_prompt}
    ]
    
    async with llm_bouncer:
        try:
            resp, raw = await client.chat.completions.create_with_completion(
                api_key=api_key,
                model=model,
                messages=msg,
                response_model=domain.get_final_schema(),
                max_retries=1,
            )
        except Exception as e:
            _handle_llm_exception(e, "Claims Scoring")
        
    usage = {
        "prompt_tokens": getattr(raw.usage, "prompt_tokens", 0),
        "completion_tokens": getattr(raw.usage, "completion_tokens", 0)
    }
    return resp, usage

async def analyze_fit(jd_text: str, resume_text: str) -> tuple[BaseModel, dict[str, int]]:
    api_key = os.environ.get("LLM_API_KEY")
    if not api_key:
        raise AnalysisEngineError("ERR_MISSING_CONFIG", "LLM API key is missing from environment.")

    domain: DomainStrategy = TechDomain()
    model = os.environ.get("LLM_MODEL")
    if not model:
        raise AnalysisEngineError("ERR_MISSING_CONFIG", "LLM_MODEL is missing from environment.")
    
    client = instructor.from_litellm(acompletion, mode=instructor.Mode.JSON)

    logger.debug(f"Starting parallel extraction for JD (length: {len(jd_text)}) and Resume (length: {len(resume_text)})...")

    total_prompt_tokens = 0
    total_completion_tokens = 0
    
    jd_task = _extract_jd_requirements(client, model, api_key, domain, jd_text)
    res_task = _extract_resume_claims(client, model, api_key, domain, resume_text)
    
    (jd_reqs, jd_usage), (res_claims, res_usage) = await asyncio.gather(jd_task, res_task)
    
    logger.debug(f"Extraction completed. Extracted {len(jd_reqs.requirements)} JD requirements and {len(res_claims.claims)} Resume claims. Starting final claim scoring...")
    
    total_prompt_tokens += jd_usage["prompt_tokens"] + res_usage["prompt_tokens"]
    total_completion_tokens += jd_usage["completion_tokens"] + res_usage["completion_tokens"]

    final_analysis, score_usage = await _score_claims(client, model, api_key, domain, jd_reqs, res_claims)
    total_prompt_tokens += score_usage["prompt_tokens"]
    total_completion_tokens += score_usage["completion_tokens"]

    usage = {
        "prompt_tokens": total_prompt_tokens,
        "completion_tokens": total_completion_tokens,
        "total_tokens": total_prompt_tokens + total_completion_tokens
    }

    return final_analysis, usage
