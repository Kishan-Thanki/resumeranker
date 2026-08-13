from typing import TypedDict


class LLMUsage(TypedDict):
    """Normalized usage statistics for a single LLM call.

    Token fields represent usage reported by the configured LLM provider.
    Optional fields use None when the provider does not expose that metric.
    Execution fields are owned by this service rather than the provider.
    """

    input_tokens: int | None
    output_tokens: int | None
    total_tokens: int | None

    reasoning_tokens: int | None
    cached_input_tokens: int | None
    cache_creation_input_tokens: int | None
    cache_read_input_tokens: int | None

    queue_wait_seconds: float
    retries: int


class AggregatedLLMUsage(TypedDict):
    """
    Normalized usage aggregated across multiple LLM calls.
    Built via aggregate_llm_usage().
    """

    input_tokens: int | None
    output_tokens: int | None
    total_tokens: int | None

    reasoning_tokens: int | None
    cached_input_tokens: int | None
    cache_creation_input_tokens: int | None
    cache_read_input_tokens: int | None

    queue_wait_seconds: float
    retries: int
    call_count: int


def _sum_optional(values: list[int | None]) -> int | None:
    """Sum values when usage is available; preserve None when it is unavailable."""
    present_values = [value for value in values if value is not None]

    if not present_values:
        return None

    return sum(present_values)


def aggregate_llm_usage(
    usages: list[LLMUsage],
) -> AggregatedLLMUsage:
    """Aggregate normalized usage across multiple LLM calls."""

    input_tokens = _sum_optional([usage["input_tokens"] for usage in usages])
    output_tokens = _sum_optional([usage["output_tokens"] for usage in usages])
    total_tokens = _sum_optional([usage["total_tokens"] for usage in usages])
    reasoning_tokens = _sum_optional([usage["reasoning_tokens"] for usage in usages])
    cached_input_tokens = _sum_optional(
        [usage["cached_input_tokens"] for usage in usages]
    )
    cache_creation_input_tokens = _sum_optional(
        [usage["cache_creation_input_tokens"] for usage in usages]
    )
    cache_read_input_tokens = _sum_optional(
        [usage["cache_read_input_tokens"] for usage in usages]
    )

    return AggregatedLLMUsage(
        input_tokens=input_tokens,
        output_tokens=output_tokens,
        total_tokens=total_tokens,
        reasoning_tokens=reasoning_tokens,
        cached_input_tokens=cached_input_tokens,
        cache_creation_input_tokens=cache_creation_input_tokens,
        cache_read_input_tokens=cache_read_input_tokens,
        queue_wait_seconds=round(
            sum(usage["queue_wait_seconds"] for usage in usages),
            4,
        ),
        retries=sum(usage["retries"] for usage in usages),
        call_count=len(usages),
    )
