from unittest.mock import MagicMock, patch

import pytest
from app.services.pdf_service import (
    extract_text_from_pdf_bytes,
    scrub_pii,
)


def _mock_pdf(text: str):
    """Create a mocked one-page PDF returning the supplied text."""
    fake_pdf = MagicMock()
    fake_page = MagicMock()
    fake_textpage = MagicMock()

    fake_textpage.get_text_range.return_value = text
    fake_textpage.close.return_value = None

    fake_page.get_textpage.return_value = fake_textpage
    fake_pdf.__iter__.return_value = iter([fake_page])

    return fake_pdf, fake_textpage


def test_extract_text_success():
    """Extracts text successfully from a valid PDF."""

    fake_pdf, fake_textpage = _mock_pdf(
        "This is a valid resume with enough characters to pass the one hundred "
        "character minimum length requirement."
    )

    with (
        patch("app.services.pdf_service.MIN_PDF_CHARS", 100),
        patch("app.services.pdf_service.pdfium.PdfDocument") as mock_pdf_cls,
    ):
        mock_pdf_cls.return_value.__enter__.return_value = fake_pdf

        text = extract_text_from_pdf_bytes(b"fakebytes")

    assert "valid resume" in text
    fake_textpage.close.assert_called_once()


def test_extract_text_insufficient_length():
    """Rejects PDFs that do not contain enough extractable text."""

    fake_pdf, fake_textpage = _mock_pdf("Too short.")

    with (
        patch("app.services.pdf_service.MIN_PDF_CHARS", 100),
        patch("app.services.pdf_service.pdfium.PdfDocument") as mock_pdf_cls,
        pytest.raises(
            ValueError,
            match=r"PDF contains insufficient text",
        ),
    ):
        mock_pdf_cls.return_value.__enter__.return_value = fake_pdf
        extract_text_from_pdf_bytes(b"fakebytes")

    fake_textpage.close.assert_called_once()


def test_scrub_pii():
    """Redacts common personally identifiable information."""

    text = """
John Doe
Email: john.doe@example.com
Phone: +1 555-123-4567
Other phone: (555) 987-6543
No code phone: 123-456-7890
LinkedIn: https://linkedin.com/in/johndoe
GitHub: https://github.com/johndoe
"""

    scrubbed = scrub_pii(text)

    assert "john.doe@example.com" not in scrubbed
    assert "[EMAIL REDACTED]" in scrubbed

    assert "+1 555-123-4567" not in scrubbed
    assert "(555) 987-6543" not in scrubbed
    assert "123-456-7890" not in scrubbed
    assert scrubbed.count("[PHONE REDACTED]") == 3

    assert "https://linkedin.com/in/[REDACTED]" in scrubbed
    assert "https://github.com/[REDACTED]" in scrubbed
