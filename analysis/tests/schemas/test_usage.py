from app.schemas.usage import LLMUsage, aggregate_llm_usage


class TestAggregateLLMUsage:
    def test_aggregates_single_usage(self) -> None:
        usages: list[LLMUsage] = [
            {
                "input_tokens": 100,
                "output_tokens": 50,
                "total_tokens": 150,
                "reasoning_tokens": 20,
                "cached_input_tokens": 10,
                "cache_creation_input_tokens": 5,
                "cache_read_input_tokens": 3,
                "queue_wait_seconds": 0.25,
                "retries": 1,
            }
        ]

        result = aggregate_llm_usage(usages)

        assert result == {
            "input_tokens": 100,
            "output_tokens": 50,
            "total_tokens": 150,
            "reasoning_tokens": 20,
            "cached_input_tokens": 10,
            "cache_creation_input_tokens": 5,
            "cache_read_input_tokens": 3,
            "queue_wait_seconds": 0.25,
            "retries": 1,
            "call_count": 1,
        }

    def test_aggregates_multiple_usages(self) -> None:
        usages: list[LLMUsage] = [
            {
                "input_tokens": 100,
                "output_tokens": 50,
                "total_tokens": 150,
                "reasoning_tokens": 20,
                "cached_input_tokens": 10,
                "cache_creation_input_tokens": 5,
                "cache_read_input_tokens": 3,
                "queue_wait_seconds": 0.25,
                "retries": 1,
            },
            {
                "input_tokens": 200,
                "output_tokens": 75,
                "total_tokens": 275,
                "reasoning_tokens": 30,
                "cached_input_tokens": 20,
                "cache_creation_input_tokens": 8,
                "cache_read_input_tokens": 4,
                "queue_wait_seconds": 0.50,
                "retries": 0,
            },
            {
                "input_tokens": 300,
                "output_tokens": 125,
                "total_tokens": 425,
                "reasoning_tokens": 40,
                "cached_input_tokens": 30,
                "cache_creation_input_tokens": 12,
                "cache_read_input_tokens": 6,
                "queue_wait_seconds": 1.25,
                "retries": 2,
            },
        ]

        result = aggregate_llm_usage(usages)

        assert result["input_tokens"] == 600
        assert result["output_tokens"] == 250
        assert result["total_tokens"] == 850
        assert result["reasoning_tokens"] == 90
        assert result["cached_input_tokens"] == 60
        assert result["cache_creation_input_tokens"] == 25
        assert result["cache_read_input_tokens"] == 13
        assert result["queue_wait_seconds"] == 2.0
        assert result["retries"] == 3
        assert result["call_count"] == 3

    def test_empty_usage_list_returns_none_for_unavailable_token_metrics(
        self,
    ) -> None:
        result = aggregate_llm_usage([])

        assert result == {
            "input_tokens": None,
            "output_tokens": None,
            "total_tokens": None,
            "reasoning_tokens": None,
            "cached_input_tokens": None,
            "cache_creation_input_tokens": None,
            "cache_read_input_tokens": None,
            "queue_wait_seconds": 0.0,
            "retries": 0,
            "call_count": 0,
        }

    def test_preserves_zero_values_when_provider_reports_zero(self) -> None:
        usages: list[LLMUsage] = [
            {
                "input_tokens": 0,
                "output_tokens": 0,
                "total_tokens": 0,
                "reasoning_tokens": 0,
                "cached_input_tokens": 0,
                "cache_creation_input_tokens": 0,
                "cache_read_input_tokens": 0,
                "queue_wait_seconds": 0.0,
                "retries": 0,
            }
        ]

        result = aggregate_llm_usage(usages)

        assert result["input_tokens"] == 0
        assert result["output_tokens"] == 0
        assert result["total_tokens"] == 0
        assert result["reasoning_tokens"] == 0
        assert result["cached_input_tokens"] == 0
        assert result["cache_creation_input_tokens"] == 0
        assert result["cache_read_input_tokens"] == 0

    def test_preserves_none_when_metric_is_unavailable_in_every_call(
        self,
    ) -> None:
        usages: list[LLMUsage] = [
            {
                "input_tokens": 100,
                "output_tokens": 50,
                "total_tokens": 150,
                "reasoning_tokens": None,
                "cached_input_tokens": None,
                "cache_creation_input_tokens": None,
                "cache_read_input_tokens": None,
                "queue_wait_seconds": 0.25,
                "retries": 1,
            },
            {
                "input_tokens": 200,
                "output_tokens": 75,
                "total_tokens": 275,
                "reasoning_tokens": None,
                "cached_input_tokens": None,
                "cache_creation_input_tokens": None,
                "cache_read_input_tokens": None,
                "queue_wait_seconds": 0.50,
                "retries": 0,
            },
        ]

        result = aggregate_llm_usage(usages)

        assert result["input_tokens"] == 300
        assert result["output_tokens"] == 125
        assert result["total_tokens"] == 425
        assert result["reasoning_tokens"] is None
        assert result["cached_input_tokens"] is None
        assert result["cache_creation_input_tokens"] is None
        assert result["cache_read_input_tokens"] is None

    def test_sums_values_when_some_calls_report_none(self) -> None:
        usages: list[LLMUsage] = [
            {
                "input_tokens": 100,
                "output_tokens": 50,
                "total_tokens": 150,
                "reasoning_tokens": 20,
                "cached_input_tokens": None,
                "cache_creation_input_tokens": 5,
                "cache_read_input_tokens": None,
                "queue_wait_seconds": 0.25,
                "retries": 1,
            },
            {
                "input_tokens": 200,
                "output_tokens": 75,
                "total_tokens": 275,
                "reasoning_tokens": None,
                "cached_input_tokens": 10,
                "cache_creation_input_tokens": None,
                "cache_read_input_tokens": 4,
                "queue_wait_seconds": 0.50,
                "retries": 0,
            },
        ]

        result = aggregate_llm_usage(usages)

        assert result["reasoning_tokens"] == 20
        assert result["cached_input_tokens"] == 10
        assert result["cache_creation_input_tokens"] == 5
        assert result["cache_read_input_tokens"] == 4

    def test_total_tokens_is_aggregated_from_provider_reported_totals(
        self,
    ) -> None:
        usages: list[LLMUsage] = [
            {
                "input_tokens": 123,
                "output_tokens": 456,
                "total_tokens": 579,
                "reasoning_tokens": None,
                "cached_input_tokens": None,
                "cache_creation_input_tokens": None,
                "cache_read_input_tokens": None,
                "queue_wait_seconds": 0.0,
                "retries": 0,
            },
            {
                "input_tokens": 10,
                "output_tokens": 20,
                "total_tokens": 30,
                "reasoning_tokens": None,
                "cached_input_tokens": None,
                "cache_creation_input_tokens": None,
                "cache_read_input_tokens": None,
                "queue_wait_seconds": 0.0,
                "retries": 0,
            },
        ]

        result = aggregate_llm_usage(usages)

        assert result["total_tokens"] == 609

    def test_call_count_matches_number_of_usages(self) -> None:
        usages: list[LLMUsage] = [
            {
                "input_tokens": 1,
                "output_tokens": 2,
                "total_tokens": 3,
                "reasoning_tokens": None,
                "cached_input_tokens": None,
                "cache_creation_input_tokens": None,
                "cache_read_input_tokens": None,
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
                "input_tokens": 1,
                "output_tokens": 1,
                "total_tokens": 2,
                "reasoning_tokens": None,
                "cached_input_tokens": None,
                "cache_creation_input_tokens": None,
                "cache_read_input_tokens": None,
                "queue_wait_seconds": 0.0,
                "retries": 3,
            },
            {
                "input_tokens": 1,
                "output_tokens": 1,
                "total_tokens": 2,
                "reasoning_tokens": None,
                "cached_input_tokens": None,
                "cache_creation_input_tokens": None,
                "cache_read_input_tokens": None,
                "queue_wait_seconds": 0.0,
                "retries": 4,
            },
        ]

        result = aggregate_llm_usage(usages)

        assert result["retries"] == 7

    def test_queue_wait_is_rounded_to_four_decimal_places(self) -> None:
        usages: list[LLMUsage] = [
            {
                "input_tokens": 10,
                "output_tokens": 5,
                "total_tokens": 15,
                "reasoning_tokens": None,
                "cached_input_tokens": None,
                "cache_creation_input_tokens": None,
                "cache_read_input_tokens": None,
                "queue_wait_seconds": 0.12345,
                "retries": 0,
            },
            {
                "input_tokens": 20,
                "output_tokens": 10,
                "total_tokens": 30,
                "reasoning_tokens": None,
                "cached_input_tokens": None,
                "cache_creation_input_tokens": None,
                "cache_read_input_tokens": None,
                "queue_wait_seconds": 0.67891,
                "retries": 1,
            },
        ]

        result = aggregate_llm_usage(usages)

        assert result["queue_wait_seconds"] == 0.8024
