"""Per-analysis token + cost accounting.

Records each LLM call into a Redis hash keyed by analysis_id. The worker
logs a cost summary at the end of each run.

Prices are per the brief's note on Claude Haiku 4.5. Stub-mode calls record
zero tokens / zero cost.
"""

import logging

import redis.asyncio as redis_async

from app.deps import get_redis_client

logger = logging.getLogger(__name__)

# Approximate per-1M-token prices for claude-haiku-4-5 (USD).
# These are constants for v1; production would source them from a config
# refreshed against Anthropic's published pricing.
HAIKU_PROMPT_PER_M = 0.80
HAIKU_COMPLETION_PER_M = 4.00


def _estimate_cost_usd(model: str, prompt_tokens: int, completion_tokens: int) -> float:
    """Per-1M-token pricing. Conservative defaults for unknown models."""
    if "haiku" in model.lower():
        p, c = HAIKU_PROMPT_PER_M, HAIKU_COMPLETION_PER_M
    else:
        # Fallback assumption — log a warning but don't crash.
        logger.warning("unknown LLM model for cost estimate: %s", model)
        p, c = 3.00, 15.00
    return (prompt_tokens / 1_000_000) * p + (completion_tokens / 1_000_000) * c


async def record_call(
    analysis_id: str,
    model: str,
    prompt_tokens: int,
    completion_tokens: int,
) -> None:
    redis: redis_async.Redis = get_redis_client()
    cost = _estimate_cost_usd(model, prompt_tokens, completion_tokens)
    key = f"llm:cost:{analysis_id}"
    pipe = redis.pipeline()
    pipe.hincrby(key, "prompt_tokens", prompt_tokens)
    pipe.hincrby(key, "completion_tokens", completion_tokens)
    pipe.hincrbyfloat(key, "cost_usd", cost)
    pipe.hincrby(key, "call_count", 1)
    # 7-day TTL: cost data is only useful while we still care about the analysis.
    pipe.expire(key, 7 * 24 * 3600)
    await pipe.execute()


async def get_totals(analysis_id: str) -> dict[str, float]:
    redis: redis_async.Redis = get_redis_client()
    # redis-py's `hgetall` async type is union with sync; narrow at the boundary.
    raw_obj = await redis.hgetall(f"llm:cost:{analysis_id}")  # type: ignore[misc]
    raw: dict[str, str] = raw_obj or {}
    if not raw:
        return {"prompt_tokens": 0, "completion_tokens": 0, "cost_usd": 0.0, "call_count": 0}
    return {
        "prompt_tokens": int(raw.get("prompt_tokens", "0")),
        "completion_tokens": int(raw.get("completion_tokens", "0")),
        "cost_usd": float(raw.get("cost_usd", "0.0")),
        "call_count": int(raw.get("call_count", "0")),
    }
