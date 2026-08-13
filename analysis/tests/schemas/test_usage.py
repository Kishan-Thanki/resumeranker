from app.schemas.usage import LLMUsage, aggregate_llm_usage


class TestAggregateLLMUsage:
    def test_aggregates_single_usage(self) -> None:
        usages: list[LLMUsage] = [
            {
                "prompt_tokens": 100,
                "completion_tokens": 50,
                "queue_wait_seconds": 0.25,
                "retries": 1,
            }
        ]

        result = aggregate_llm_usage(usages)

        assert result == {
            "prompt_tokens": 100,
            "completion_tokens": 50,
            "total_tokens": 150,
            "queue_wait_seconds": 0.25,
            "retries": 1,
            "call_count": 1,
        }

    def test_aggregates_multiple_usages(self) -> None:
        usages: list[LLMUsage] = [
            {
                "prompt_tokens": 100,
                "completion_tokens": 50,
                "queue_wait_seconds": 0.25,
                "retries": 1,
            },
            {
                "prompt_tokens": 200,
                "completion_tokens": 75,
                "queue_wait_seconds": 0.50,
                "retries": 0,
            },
            {
                "prompt_tokens": 300,
                "completion_tokens": 125,
                "queue_wait_seconds": 1.25,
                "retries": 2,
            },
        ]

        result = aggregate_llm_usage(usages)

        assert result["prompt_tokens"] == 600
        assert result["completion_tokens"] == 250
        assert result["total_tokens"] == 850
        assert result["queue_wait_seconds"] == 2.0
        assert result["retries"] == 3
        assert result["call_count"] == 3

    def test_empty_usage_list_returns_zero_values(self) -> None:
        result = aggregate_llm_usage([])

        assert result == {
            "prompt_tokens": 0,
            "completion_tokens": 0,
            "total_tokens": 0,
            "queue_wait_seconds": 0.0,
            "retries": 0,
            "call_count": 0,
        }

    def test_rounds_aggregated_queue_wait_time(self) -> None:
        usages: list[LLMUsage] = [
            {
                "prompt_tokens": 10,
                "completion_tokens": 5,
                "queue_wait_seconds": 0.12345,
                "retries": 0,
            },
            {
                "prompt_tokens": 20,
                "completion_tokens": 10,
                "queue_wait_seconds": 0.67891,
                "retries": 1,
            },
        ]

        result = aggregate_llm_usage(usages)

        assert result["queue_wait_seconds"] == 0.8024

    def test_total_tokens_is_prompt_plus_completion(self) -> None:
        usages: list[LLMUsage] = [
            {
                "prompt_tokens": 123,
                "completion_tokens": 456,
                "queue_wait_seconds": 0.0,
                "retries": 0,
            },
            {
                "prompt_tokens": 10,
                "completion_tokens": 20,
                "queue_wait_seconds": 0.0,
                "retries": 0,
            },
        ]

        result = aggregate_llm_usage(usages)

        assert result["total_tokens"] == (
            result["prompt_tokens"] + result["completion_tokens"]
        )

    def test_call_count_matches_number_of_usages(self) -> None:
        usages: list[LLMUsage] = [
            {
                "prompt_tokens": 1,
                "completion_tokens": 2,
                "queue_wait_seconds": 0.0,
                "retries": 0,
            }
            for _ in range(5)
        ]

        result = aggregate_llm_usage(usages)

        assert result["call_count"] == 5

    def test_retries_are_summed(self) -> None:
        usages: list[LLMUsage] = [
            {
                "prompt_tokens": 1,
                "completion_tokens": 1,
                "queue_wait_seconds": 0.0,
                "retries": 3,
            },
            {
                "prompt_tokens": 1,
                "completion_tokens": 1,
                "queue_wait_seconds": 0.0,
                "retries": 4,
            },
        ]

        result = aggregate_llm_usage(usages)

        assert result["retries"] == 7
