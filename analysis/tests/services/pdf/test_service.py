from unittest.mock import MagicMock

import app.services.pdf.service as pdf_service
import pytest


class TestScrubPii:
    def test_redacts_email_address(self) -> None:
        text = "Contact me at kishan@example.com for more information."

        result = pdf_service.scrub_pii(text)

        assert result == "Contact me at [EMAIL REDACTED] for more information."

    def test_redacts_linkedin_profile(self) -> None:
        text = "LinkedIn: https://linkedin.com/in/kishan-thanki"

        result = pdf_service.scrub_pii(text)

        assert result == "LinkedIn: https://linkedin.com/in/[REDACTED]"

    def test_redacts_github_profile(self) -> None:
        text = "GitHub: https://github.com/kishan-thanki"

        result = pdf_service.scrub_pii(text)

        assert result == "GitHub: https://github.com/[REDACTED]"

    def test_redacts_phone_number(self) -> None:
        text = "Phone: +91 9876543210"

        result = pdf_service.scrub_pii(text)

        assert "[PHONE REDACTED]" in result
        assert "9876543210" not in result

    def test_redacts_multiple_pii_types(self) -> None:
        text = (
            "Email: kishan@example.com\n"
            "Phone: +91 9876543210\n"
            "LinkedIn: https://linkedin.com/in/kishan\n"
            "GitHub: https://github.com/kishan"
        )

        result = pdf_service.scrub_pii(text)

        assert "[EMAIL REDACTED]" in result
        assert "[PHONE REDACTED]" in result
        assert "linkedin.com/in/[REDACTED]" in result
        assert "github.com/[REDACTED]" in result

        assert "kishan@example.com" not in result
        assert "9876543210" not in result
        assert "linkedin.com/in/kishan" not in result
        assert "github.com/kishan" not in result

    def test_preserves_non_pii_text(self) -> None:
        text = "Senior Backend Engineer with 6 years of Python experience."

        result = pdf_service.scrub_pii(text)

        assert result == text

    def test_redacts_multiple_phone_numbers_without_offset_corruption(self) -> None:
        text = "Primary: +91 9876543210 Secondary: +91 9123456789"

        result = pdf_service.scrub_pii(text)

        assert result.count("[PHONE REDACTED]") == 2
        assert "9876543210" not in result
        assert "9123456789" not in result


class TestExtractTextFromPdfBytes:
    def test_benchmark_mode_returns_dummy_text(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setattr(pdf_service, "BENCHMARK_MODE", True)

        result = pdf_service.extract_text_from_pdf_bytes(b"not-a-real-pdf")

        assert len(result) >= pdf_service.MIN_PDF_CHARS
        assert "Benchmark dummy text" in result

    def test_extracts_text_from_pdf_pages(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setattr(pdf_service, "BENCHMARK_MODE", False)
        monkeypatch.setattr(pdf_service, "MIN_PDF_CHARS", 10)

        first_textpage = MagicMock()
        first_textpage.get_text_range.return_value = "First page text"

        second_textpage = MagicMock()
        second_textpage.get_text_range.return_value = "Second page text"

        first_page = MagicMock()
        first_page.get_textpage.return_value = first_textpage

        second_page = MagicMock()
        second_page.get_textpage.return_value = second_textpage

        pdf_document = MagicMock()
        pdf_document.__len__.return_value = 2
        pdf_document.__iter__.return_value = iter(
            [first_page, second_page]
        )

        pdf_document_cm = MagicMock()
        pdf_document_cm.__enter__.return_value = pdf_document
        pdf_document_cm.__exit__.return_value = False

        monkeypatch.setattr(
            pdf_service.pdfium,
            "PdfDocument",
            MagicMock(return_value=pdf_document_cm),
        )

        result = pdf_service.extract_text_from_pdf_bytes(b"fake-pdf")

        assert result == "First page text\nSecond page text"

        first_textpage.get_text_range.assert_called_once()
        second_textpage.get_text_range.assert_called_once()
        first_page.get_textpage.assert_called_once()
        second_page.get_textpage.assert_called_once()

        first_textpage.close.assert_called_once()
        second_textpage.close.assert_called_once()
        first_page.close.assert_called_once()
        second_page.close.assert_called_once()

    def test_skips_empty_textpage_results(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setattr(pdf_service, "BENCHMARK_MODE", False)
        monkeypatch.setattr(pdf_service, "MIN_PDF_CHARS", 10)

        first_textpage = MagicMock()
        first_textpage.get_text_range.return_value = None

        second_textpage = MagicMock()
        second_textpage.get_text_range.return_value = "Enough text"

        first_page = MagicMock()
        first_page.get_textpage.return_value = first_textpage

        second_page = MagicMock()
        second_page.get_textpage.return_value = second_textpage

        pdf_document = MagicMock()
        pdf_document.__len__.return_value = 2
        pdf_document.__iter__.return_value = iter(
            [first_page, second_page]
        )

        pdf_document_cm = MagicMock()
        pdf_document_cm.__enter__.return_value = pdf_document
        pdf_document_cm.__exit__.return_value = False

        monkeypatch.setattr(
            pdf_service.pdfium,
            "PdfDocument",
            MagicMock(return_value=pdf_document_cm),
        )

        result = pdf_service.extract_text_from_pdf_bytes(b"fake-pdf")

        assert result == "Enough text"

    def test_rejects_pdf_with_too_many_pages(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setattr(pdf_service, "BENCHMARK_MODE", False)
        monkeypatch.setattr(pdf_service, "MAX_PDF_PAGES", 2)

        pdf_document = MagicMock()
        pdf_document.__len__.return_value = 3

        pdf_document_cm = MagicMock()
        pdf_document_cm.__enter__.return_value = pdf_document
        pdf_document_cm.__exit__.return_value = False

        monkeypatch.setattr(
            pdf_service.pdfium,
            "PdfDocument",
            MagicMock(return_value=pdf_document_cm),
        )

        with pytest.raises(
            ValueError,
            match="PDF has too many pages",
        ):
            pdf_service.extract_text_from_pdf_bytes(b"fake-pdf")

    def test_rejects_pdf_with_insufficient_text(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setattr(pdf_service, "BENCHMARK_MODE", False)
        monkeypatch.setattr(pdf_service, "MIN_PDF_CHARS", 50)

        textpage = MagicMock()
        textpage.get_text_range.return_value = "Too short"

        page = MagicMock()
        page.get_textpage.return_value = textpage

        pdf_document = MagicMock()
        pdf_document.__len__.return_value = 1
        pdf_document.__iter__.return_value = iter([page])

        pdf_document_cm = MagicMock()
        pdf_document_cm.__enter__.return_value = pdf_document
        pdf_document_cm.__exit__.return_value = False

        monkeypatch.setattr(
            pdf_service.pdfium,
            "PdfDocument",
            MagicMock(return_value=pdf_document_cm),
        )

        with pytest.raises(
            ValueError,
            match="PDF contains insufficient text",
        ):
            pdf_service.extract_text_from_pdf_bytes(b"fake-pdf")

    def test_rejects_invalid_pdf(self) -> None:
        with pytest.raises(
            ValueError,
            match=(
                r"Could not process PDF \(it may be corrupted, "
                r"password-protected,"
            ),
        ):
            pdf_service.extract_text_from_pdf_bytes(b"definitely-not-a-pdf")

    def test_redacts_pii_after_pdf_text_extraction(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setattr(pdf_service, "BENCHMARK_MODE", False)
        monkeypatch.setattr(pdf_service, "MIN_PDF_CHARS", 10)

        textpage = MagicMock()
        textpage.get_text_range.return_value = (
            "Senior Engineer\n"
            "Email: kishan@example.com\n"
            "Phone: +91 9876543210"
        )

        page = MagicMock()
        page.get_textpage.return_value = textpage

        pdf_document = MagicMock()
        pdf_document.__len__.return_value = 1
        pdf_document.__iter__.return_value = iter([page])

        pdf_document_cm = MagicMock()
        pdf_document_cm.__enter__.return_value = pdf_document
        pdf_document_cm.__exit__.return_value = False

        monkeypatch.setattr(
            pdf_service.pdfium,
            "PdfDocument",
            MagicMock(return_value=pdf_document_cm),
        )

        result = pdf_service.extract_text_from_pdf_bytes(b"fake-pdf")

        assert "Senior Engineer" in result
        assert "[EMAIL REDACTED]" in result
        assert "[PHONE REDACTED]" in result
        assert "kishan@example.com" not in result
        assert "9876543210" not in result

    def test_converts_unexpected_pdf_exception_to_value_error(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setattr(pdf_service, "BENCHMARK_MODE", False)

        pdf_document_cm = MagicMock()
        pdf_document_cm.__enter__.side_effect = RuntimeError("PDFium exploded")
        pdf_document_cm.__exit__.return_value = False

        monkeypatch.setattr(
            pdf_service.pdfium,
            "PdfDocument",
            MagicMock(return_value=pdf_document_cm),
        )

        with pytest.raises(
            ValueError,
            match="Could not process PDF",
        ) as exc_info:
            pdf_service.extract_text_from_pdf_bytes(b"fake-pdf")

        assert "PDFium exploded" in str(exc_info.value)

    def test_preserves_value_error_from_page_limit_validation(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setattr(pdf_service, "BENCHMARK_MODE", False)
        monkeypatch.setattr(pdf_service, "MAX_PDF_PAGES", 1)

        pdf_document = MagicMock()
        pdf_document.__len__.return_value = 2

        pdf_document_cm = MagicMock()
        pdf_document_cm.__enter__.return_value = pdf_document
        pdf_document_cm.__exit__.return_value = False

        monkeypatch.setattr(
            pdf_service.pdfium,
            "PdfDocument",
            MagicMock(return_value=pdf_document_cm),
        )

        with pytest.raises(ValueError) as exc_info:
            pdf_service.extract_text_from_pdf_bytes(b"fake-pdf")

        assert "PDF has too many pages" in str(exc_info.value)

    def test_normalizes_whitespace_around_extracted_text(
        self,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        monkeypatch.setattr(pdf_service, "BENCHMARK_MODE", False)
        monkeypatch.setattr(pdf_service, "MIN_PDF_CHARS", 5)

        textpage = MagicMock()
        textpage.get_text_range.return_value = "   Useful resume text   "

        page = MagicMock()
        page.get_textpage.return_value = textpage

        pdf_document = MagicMock()
        pdf_document.__len__.return_value = 1
        pdf_document.__iter__.return_value = iter([page])

        pdf_document_cm = MagicMock()
        pdf_document_cm.__enter__.return_value = pdf_document
        pdf_document_cm.__exit__.return_value = False

        monkeypatch.setattr(
            pdf_service.pdfium,
            "PdfDocument",
            MagicMock(return_value=pdf_document_cm),
        )

        result = pdf_service.extract_text_from_pdf_bytes(b"fake-pdf")

        assert result == "Useful resume text"
