from typing import TypedDict


class LLMUsage(TypedDict):
    """Usage statistics for a single LLM call."""

    prompt_tokens: int
    completion_tokens: int
    queue_wait_seconds: float
    retries: int


class AggregatedLLMUsage(TypedDict):
    """
    Usage statistics summed across multiple LLM calls.
    Built via aggregate_llm_usage().
    """

    prompt_tokens: int
    completion_tokens: int
    total_tokens: int
    queue_wait_seconds: float
    retries: int
    call_count: int


def aggregate_llm_usage(usages: list[LLMUsage]) -> AggregatedLLMUsage:
    """Sum per-call usage into a single aggregate."""
    prompt_tokens = sum(u["prompt_tokens"] for u in usages)
    completion_tokens = sum(u["completion_tokens"] for u in usages)
    queue_wait = round(sum(u["queue_wait_seconds"] for u in usages), 4)

    return AggregatedLLMUsage(
        prompt_tokens=prompt_tokens,
        completion_tokens=completion_tokens,
        total_tokens=prompt_tokens + completion_tokens,
        queue_wait_seconds=queue_wait,
        retries=sum(u["retries"] for u in usages),
        call_count=len(usages),
    )
