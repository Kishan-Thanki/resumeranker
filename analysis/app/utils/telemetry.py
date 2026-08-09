import asyncio
import time
from collections.abc import Awaitable, Callable
from functools import wraps
from typing import Any, ParamSpec, TypeVar

from app.logger import setup_logger
from app.schemas import LLMUsage, aggregate_llm_usage

P = ParamSpec("P")
T = TypeVar("T", bound=tuple[Any, LLMUsage])

metrics_logger = setup_logger("metrics")


def track_llm_cost(step_name: str):
    """
    Decorator that records telemetry for each LLM pipeline stage.
    Emits structured metrics including latency, token usage,
    queue wait time, retry count, and success/failure/cancelled status.
    """

    def decorator(func: Callable[P, Awaitable[T]]) -> Callable[P, Awaitable[T]]:
        @wraps(func)
        async def wrapper(*args: P.args, **kwargs: P.kwargs) -> T:
            start_time = time.perf_counter()
            try:
                result = await func(*args, **kwargs)
                _, usage = result
                aggregated = aggregate_llm_usage([usage])
                metrics_logger.info(
                    "LLM step completed",
                    extra={
                        "step": step_name,
                        "status": "success",
                        "prompt_tokens": aggregated["prompt_tokens"],
                        "completion_tokens": aggregated["completion_tokens"],
                        "total_tokens": aggregated["total_tokens"],
                        "latency_ms": round(
                            (time.perf_counter() - start_time) * 1000,
                            2,
                        ),
                        "queue_wait_ms": round(
                            aggregated["queue_wait_seconds"] * 1000,
                            2,
                        ),
                        "retries": aggregated["retries"],
                    },
                )
                return result
            except asyncio.CancelledError:
                metrics_logger.warning(
                    "LLM step cancelled",
                    extra={
                        "step": step_name,
                        "status": "cancelled",
                        "latency_ms": round(
                            (time.perf_counter() - start_time) * 1000,
                            2,
                        ),
                    },
                )
                raise
            except Exception:
                metrics_logger.exception(
                    "LLM step failed",
                    extra={
                        "step": step_name,
                        "status": "failure",
                        "latency_ms": round(
                            (time.perf_counter() - start_time) * 1000,
                            2,
                        ),
                    },
                )
                raise

        return wrapper

    return decorator
