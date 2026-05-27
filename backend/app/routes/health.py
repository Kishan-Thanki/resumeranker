from typing import Annotated

import redis.asyncio as redis_async
from fastapi import APIRouter, Depends, Response, status
from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession

from app.db import get_db
from app.deps import get_redis

router = APIRouter(tags=["health"])


async def _check_db(db: AsyncSession) -> bool:
    try:
        await db.execute(text("SELECT 1"))
    except Exception:
        return False
    return True


async def _check_redis(redis: redis_async.Redis) -> bool:
    try:
        return bool(await redis.ping())
    except Exception:
        return False


@router.get("/health")
async def health(
    response: Response,
    db: Annotated[AsyncSession, Depends(get_db)],
    redis: Annotated[redis_async.Redis, Depends(get_redis)],
) -> dict[str, str]:
    """Public liveness probe.

    Returns `{"status": "ok"}` when both DB and Redis are reachable, or a
    503 with `{"status": "degraded"}` otherwise. Intentionally does NOT
    reveal which internal dependency failed — that information goes to the
    server logs and to operators via `docker compose ps` / `logs`, not to
    anonymous callers.
    """
    db_ok = await _check_db(db)
    redis_ok = await _check_redis(redis)
    if db_ok and redis_ok:
        return {"status": "ok"}
    response.status_code = status.HTTP_503_SERVICE_UNAVAILABLE
    return {"status": "degraded"}
