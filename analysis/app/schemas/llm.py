from typing import TypedDict


class LLMUsage(TypedDict):
    """Usage statistics for a single LLM call."""

    prompt_tokens: int
    completion_tokens: int
    queue_wait_seconds: float
    retries: int


class AggregatedLLMUsage(TypedDict):
    """Aggregated usage statistics across multiple LLM calls."""

    prompt_tokens: int
    completion_tokens: int
    total_tokens: int
    queue_wait_seconds: float
    retries: int