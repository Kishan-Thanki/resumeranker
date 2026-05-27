"""Redis-backed cache for LLM responses, keyed by (model, prompt_version, input).

JD extraction and resume claims are the highest-value cache hits — the same
JD is often analyzed against multiple resumes, and the same resume against
multiple JDs.
"""

import hashlib
import json

import redis.asyncio as redis_async

from app.deps import get_redis_client

CACHE_TTL_SECONDS = 7 * 24 * 3600  # 7 days


def cache_key(model: str, prompt_version: str, input_text: str, kind: str) -> str:
    digest = hashlib.sha256(f"{kind}::{model}::{prompt_version}::{input_text}".encode()).hexdigest()
    return f"llm:cache:{kind}:{digest}"


async def get_cached(
    kind: str,
    model: str,
    prompt_version: str,
    input_text: str,
) -> list[dict[str, object]] | None:
    redis: redis_async.Redis = get_redis_client()
    key = cache_key(model, prompt_version, input_text, kind)
    raw = await redis.get(key)
    if raw is None:
        return None
    try:
        decoded: list[dict[str, object]] = json.loads(raw)
        return decoded
    except (TypeError, ValueError):
        return None


async def set_cached(
    kind: str,
    model: str,
    prompt_version: str,
    input_text: str,
    payload: list[dict[str, object]],
) -> None:
    redis: redis_async.Redis = get_redis_client()
    key = cache_key(model, prompt_version, input_text, kind)
    await redis.setex(key, CACHE_TTL_SECONDS, json.dumps(payload))
