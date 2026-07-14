import pytest
from unittest.mock import patch, MagicMock
from app.services.pdf_service import extract_text_from_pdf_bytes, scrub_pii

def test_extract_text_success():
    """Unit test for successful PDF extraction using mocked pypdfium2."""
    fake_pdf = MagicMock()
    fake_page = MagicMock()
    fake_textpage = MagicMock()
    fake_textpage.get_text_range.return_value = "This is a valid resume with enough characters to pass the one hundred character minimum length requirement."
    fake_page.get_textpage.return_value = fake_textpage
    fake_pdf.__len__.return_value = 1
    fake_pdf.__getitem__.return_value = fake_page
    
    with patch("app.services.pdf_service.pdfium.PdfDocument") as mock_pdf_cls:
        mock_pdf_cls.return_value.__enter__.return_value = fake_pdf
        text = extract_text_from_pdf_bytes(b"fakebytes")
        assert "valid resume" in text

def test_extract_text_insufficient_length():
    """Unit test to ensure tiny/corrupt PDFs are rejected cleanly."""
    fake_pdf = MagicMock()
    fake_page = MagicMock()
    fake_textpage = MagicMock()
    fake_textpage.get_text_range.return_value = "Too short."
    fake_page.get_textpage.return_value = fake_textpage
    fake_pdf.__len__.return_value = 1
    fake_pdf.__getitem__.return_value = fake_page
    
    with patch("app.services.pdf_service.pdfium.PdfDocument", return_value=fake_pdf):
        with pytest.raises(ValueError, match="less than 100 characters"):
            extract_text_from_pdf_bytes(b"fakebytes")

def test_scrub_pii():
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
    
    # Assert PII is redacted
    assert "john.doe@example.com" not in scrubbed
    assert "[EMAIL REDACTED]" in scrubbed
    
    assert "+1 555-123-4567" not in scrubbed
    assert "(555) 987-6543" not in scrubbed
    assert "123-456-7890" not in scrubbed
    assert scrubbed.count("[PHONE REDACTED]") == 3
    
    # Assert professional links have paths redacted
    assert "https://linkedin.com/in/[REDACTED]" in scrubbed
    assert "https://github.com/[REDACTED]" in scrubbed
