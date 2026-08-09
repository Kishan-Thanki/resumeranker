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

    queue_wait_seconds and retries are TOTALS across all calls here —
    same field names as LLMUsage, different scope (per-call there,
    summed here). Build this with aggregate_llm_usage(), not by hand,
    so total_tokens can't drift from prompt_tokens + completion_tokens.
    """

    prompt_tokens: int
    completion_tokens: int
    total_tokens: int
    queue_wait_seconds: float
    retries: int
    call_count: int


def aggregate_llm_usage(usages: list[LLMUsage]) -> AggregatedLLMUsage:
    """
    Sum per-call usage into a single aggregate.

    total_tokens is derived here, never set independently, so it can't
    disagree with prompt_tokens + completion_tokens. call_count is included
    since it's needed to turn any of these sums into a per-call average later.
    """
    prompt_tokens = sum(u["prompt_tokens"] for u in usages)
    completion_tokens = sum(u["completion_tokens"] for u in usages)
    return AggregatedLLMUsage(
        prompt_tokens=prompt_tokens,
        completion_tokens=completion_tokens,
        total_tokens=prompt_tokens + completion_tokens,
        queue_wait_seconds=sum(u["queue_wait_seconds"] for u in usages),
        retries=sum(u["retries"] for u in usages),
        call_count=len(usages),
    )
