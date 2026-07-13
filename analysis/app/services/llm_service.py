import os
import asyncio
import instructor
from litellm import acompletion
from pydantic import BaseModel

from app.schemas import (
    ExtractedRequirement,
    ResumeClaim,
    FinalAnalysisResult
)
from app.domain.base import DomainStrategy
from app.domain.tech import TechDomain

class JDRequirementsResponse(BaseModel):
    requirements: list[ExtractedRequirement]

class ResumeClaimsResponse(BaseModel):
    claims: list[ResumeClaim]

MAX_CONCURRENT = int(os.environ.get("MAX_CONCURRENT_LLM_REQUESTS", 3))
llm_bouncer = asyncio.Semaphore(MAX_CONCURRENT)

async def _extract_jd_requirements(client, model: str, api_key: str, domain: DomainStrategy, jd_text: str) -> tuple[JDRequirementsResponse, dict]:
    msg = [
        {"role": "system", "content": domain.jd_extraction_prompt()},
        {"role": "user", "content": jd_text}
    ]
    
    async with llm_bouncer:
        resp, raw = await client.chat.completions.create_with_completion(
            api_key=api_key,
            model=model,
            messages=msg,
            response_model=JDRequirementsResponse,
            max_retries=3,     
            num_retries=5,     
        )
        
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
        resp, raw = await client.chat.completions.create_with_completion(
            api_key=api_key,
            model=model,
            messages=msg,
            response_model=ResumeClaimsResponse,
            max_retries=3,     
            num_retries=5,     
        )
        
    usage = {
        "prompt_tokens": getattr(raw.usage, "prompt_tokens", 0),
        "completion_tokens": getattr(raw.usage, "completion_tokens", 0)
    }
    return resp, usage

async def _score_claims(client, model: str, api_key: str, domain: DomainStrategy, jd_reqs: JDRequirementsResponse, res_claims: ResumeClaimsResponse) -> tuple[FinalAnalysisResult, dict]:
    user_prompt = f"JD Requirements:\n{jd_reqs.model_dump_json(indent=2)}\n\nResume Claims:\n{res_claims.model_dump_json(indent=2)}"
    msg = [
        {"role": "system", "content": domain.scoring_prompt()},
        {"role": "user", "content": user_prompt}
    ]
    
    async with llm_bouncer:
        resp, raw = await client.chat.completions.create_with_completion(
            api_key=api_key,
            model=model,
            messages=msg,
            response_model=FinalAnalysisResult,
            max_retries=3,     
            num_retries=5,     
        )
        
    usage = {
        "prompt_tokens": getattr(raw.usage, "prompt_tokens", 0),
        "completion_tokens": getattr(raw.usage, "completion_tokens", 0)
    }
    return resp, usage

async def analyze_fit(jd_text: str, resume_text: str) -> tuple[FinalAnalysisResult, dict[str, int]]:
    """
    Executes the 3-step LLM pipeline:
    1. Extract JD Requirements
    2. Extract Resume Claims
    3. Score Claims against Requirements
    """
    api_key = os.environ.get("LLM_API_KEY")
    if not api_key:
        raise ValueError("LLM API key is missing from environment. Cannot perform analysis.")

    domain: DomainStrategy = TechDomain()
    model = os.environ.get("LLM_MODEL")
    if not model:
        raise ValueError("LLM_MODEL is missing from environment. Cannot perform analysis.")
    
    client = instructor.from_litellm(acompletion)

    total_prompt_tokens = 0
    total_completion_tokens = 0
    
    jd_task = _extract_jd_requirements(client, model, api_key, domain, jd_text)
    res_task = _extract_resume_claims(client, model, api_key, domain, resume_text)
    
    (jd_reqs, jd_usage), (res_claims, res_usage) = await asyncio.gather(jd_task, res_task)
    
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
