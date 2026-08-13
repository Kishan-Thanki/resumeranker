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

    Token metrics are provider-normalized. A value of None means the
    provider did not report that metric.
    """

    def decorator(func: Callable[P, Awaitable[T]]) -> Callable[P, Awaitable[T]]:
        @wraps(func)
        async def wrapper(*args: P.args, **kwargs: P.kwargs) -> T:
            start_time = time.perf_counter()

            def get_latency_ms() -> float:
                return round((time.perf_counter() - start_time) * 1000, 2)

            try:
                result = await func(*args, **kwargs)
                _, usage = result
                aggregated = aggregate_llm_usage([usage])

                metrics_logger.info(
                    "LLM step completed",
                    extra={
                        "step": step_name,
                        "status": "success",
                        "input_tokens": aggregated["input_tokens"],
                        "output_tokens": aggregated["output_tokens"],
                        "total_tokens": aggregated["total_tokens"],
                        "reasoning_tokens": aggregated["reasoning_tokens"],
                        "cached_input_tokens": aggregated["cached_input_tokens"],
                        "cache_creation_input_tokens": aggregated[
                            "cache_creation_input_tokens"
                        ],
                        "cache_read_input_tokens": aggregated[
                            "cache_read_input_tokens"
                        ],
                        "latency_ms": get_latency_ms(),
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
                        "latency_ms": get_latency_ms(),
                    },
                )
                raise

            except Exception:
                metrics_logger.exception(
                    "LLM step failed",
                    extra={
                        "step": step_name,
                        "status": "failure",
                        "latency_ms": get_latency_ms(),
                    },
                )
                raise

        return wrapper

    return decorator
