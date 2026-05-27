"""Redis-backed sliding-window rate limiter.

Lightweight implementation: per-key counter with TTL. Good enough for the
v1 anti-abuse use cases (magic-link request and per-user analysis quota).
"""

import redis.asyncio as redis_async


class RateLimitExceeded(Exception):
    pass


async def check_and_increment(
    redis: redis_async.Redis,
    key: str,
    max_count: int,
    window_seconds: int,
) -> int:
    """Increment counter at `key`. Set TTL on first hit. Raise if over the
    limit. Returns the new count on success."""
    pipe = redis.pipeline()
    pipe.incr(key)
    pipe.ttl(key)
    count_result, ttl_result = await pipe.execute()
    count = int(count_result)
    if int(ttl_result) < 0:
        await redis.expire(key, window_seconds)
    if count > max_count:
        raise RateLimitExceeded(f"limit {max_count} per {window_seconds}s exceeded for {key}")
    return count
