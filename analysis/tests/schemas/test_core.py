import pytest
from app.schemas.core import Evidence, normalize_text
from pydantic import ValidationError


class TestNormalizeText:
    def test_normalizes_unicode_and_whitespace(self) -> None:
        text = '  Senior\u2014level   Python  engineer\twith\n  "quotes"  '

        result = normalize_text(text)

        assert result == 'Senior-level Python engineer with "quotes"'

    def test_normalizes_smart_quotes(self) -> None:
        text = "“Python” and ‘FastAPI’"

        result = normalize_text(text)

        assert result == '"Python" and \'FastAPI\''

    def test_normalizes_dash_variants(self) -> None:
        text = "Python–FastAPI—PostgreSQL"

        result = normalize_text(text)

        assert result == "Python-FastAPI-PostgreSQL"

    def test_normalizes_unicode_compatibility_characters(self) -> None:
        text = "Ｐｙｔｈｏｎ"

        result = normalize_text(text)

        assert result == "Python"

    def test_empty_text_becomes_empty_string(self) -> None:
        assert normalize_text("") == ""


class TestEvidence:
    def test_accepts_valid_jd_evidence(self) -> None:
        evidence = Evidence(
            text="5+ years of Python experience",
            source="jd",
            location="Requirements",
        )

        assert evidence.text == "5+ years of Python experience"
        assert evidence.source == "jd"
        assert evidence.location == "Requirements"

    def test_accepts_valid_resume_evidence(self) -> None:
        evidence = Evidence(
            text="Built production APIs with FastAPI",
            source="resume",
            location="Experience",
        )

        assert evidence.source == "resume"

    def test_strips_whitespace_from_fields(self) -> None:
        evidence = Evidence(
            text="  Python experience  ",
            source="jd",
            location="  Requirements  ",
        )

        assert evidence.text == "Python experience"
        assert evidence.location == "Requirements"

    def test_rejects_empty_evidence_text(self) -> None:
        with pytest.raises(ValidationError):
            Evidence(
                text="",
                source="jd",
            )

    def test_rejects_invalid_evidence_source(self) -> None:
        with pytest.raises(ValidationError):
            Evidence(
                text="Python experience",
                source="other",  # type: ignore[arg-type]
            )

    def test_verifies_evidence_against_raw_source_text(self) -> None:
        source_text = (
            "The candidate must have 5+ years of Python experience "
            "and strong backend development skills."
        )

        evidence = Evidence.model_validate(
            {
                "text": "5+ years of Python experience",
                "source": "jd",
            },
            context={
                "source_texts": {
                    "jd": source_text,
                },
            },
        )

        assert evidence.text == "5+ years of Python experience"

    def test_verifies_evidence_using_normalized_source_text(self) -> None:
        source_text = "Senior\u2014level   Python engineer"

        evidence = Evidence.model_validate(
            {
                "text": "Senior-level Python engineer",
                "source": "jd",
            },
            context={
                "source_texts": {
                    "jd": source_text,
                },
            },
        )

        assert evidence.text == "Senior-level Python engineer"

    def test_rejects_evidence_not_found_in_source(self) -> None:
        with pytest.raises(
            ValidationError,
            match="Evidence text not found verbatim in jd source",
        ):
            Evidence.model_validate(
                {
                    "text": "Java experience",
                    "source": "jd",
                },
                context={
                    "source_texts": {
                        "jd": "The candidate must have Python experience.",
                    },
                },
            )

    def test_uses_precomputed_normalized_source_text(self) -> None:
        evidence = Evidence.model_validate(
            {
                "text": "Senior-level Python engineer",
                "source": "jd",
            },
            context={
                "normalized_source_texts": {
                    "jd": "Senior-level Python engineer with FastAPI",
                },
                "source_texts": {
                    "jd": "This raw source should not be needed.",
                },
            },
        )

        assert evidence.text == "Senior-level Python engineer"

    def test_skips_verbatim_check_when_no_source_context_is_provided(self) -> None:
        evidence = Evidence(
            text="This source text is unavailable during validation",
            source="jd",
        )

        assert evidence.text == "This source text is unavailable during validation"

    def test_verifies_against_the_correct_source_document(self) -> None:
        with pytest.raises(
            ValidationError,
            match="Evidence text not found verbatim in resume source",
        ):
            Evidence.model_validate(
                {
                    "text": "Python experience",
                    "source": "resume",
                },
                context={
                    "source_texts": {
                        "jd": "Python experience",
                        "resume": "Java experience",
                    },
                },
            )

    def test_allows_optional_location_to_be_none(self) -> None:
        evidence = Evidence(
            text="Python experience",
            source="jd",
            location=None,
        )

        assert evidence.location is None
