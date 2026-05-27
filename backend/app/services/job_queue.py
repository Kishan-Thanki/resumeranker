"""Thin wrapper around arq's enqueue. Keeps route code free of arq imports."""

from arq import create_pool
from arq.connections import ArqRedis

from app.worker import WorkerSettings

_pool: ArqRedis | None = None


async def get_pool() -> ArqRedis:
    global _pool
    if _pool is None:
        _pool = await create_pool(WorkerSettings.redis_settings)
    return _pool


async def enqueue_analysis(analysis_id: str) -> None:
    pool = await get_pool()
    await pool.enqueue_job(
        "run_analysis",
        analysis_id,
        _queue_name=WorkerSettings.queue_name,
    )
