import time
from collections.abc import Awaitable, Callable
from functools import wraps
from typing import Any, ParamSpec, TypeVar

from app.logger import setup_logger
from app.schemas.llm import LLMUsage

P = ParamSpec("P")
T = TypeVar("T", bound=tuple[Any, LLMUsage])

metrics_logger = setup_logger("metrics")


def track_llm_cost(step_name: str):
    """
    Decorator that records telemetry for each LLM pipeline stage.

    Emits structured metrics including latency, token usage,
    queue wait time, retry count, and success/failure status.
    """

    def decorator(func: Callable[P, Awaitable[T]]) -> Callable[P, Awaitable[T]]:
        @wraps(func)
        async def wrapper(*args: P.args, **kwargs: P.kwargs) -> T:
            start_time = time.perf_counter()

            try:
                result = await func(*args, **kwargs)
                _, usage = result

                prompt_tokens = usage["prompt_tokens"]
                completion_tokens = usage["completion_tokens"]

                metrics_logger.info(
                    "LLM step completed",
                    extra={
                        "step": step_name,
                        "status": "success",
                        "prompt_tokens": prompt_tokens,
                        "completion_tokens": completion_tokens,
                        "total_tokens": prompt_tokens + completion_tokens,
                        "latency_ms": round(
                            (time.perf_counter() - start_time) * 1000,
                            2,
                        ),
                        "queue_wait_ms": round(
                            usage["queue_wait_seconds"] * 1000,
                            2,
                        ),
                        "retries": usage["retries"],
                    },
                )

                return result

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
