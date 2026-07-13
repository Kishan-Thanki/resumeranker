import pytest
from unittest.mock import patch, MagicMock
from app.services.pdf_service import extract_text_from_pdf_bytes

def test_extract_text_success():
    """Unit test for successful PDF extraction using mocked pdfplumber."""
    fake_pdf = MagicMock()
    fake_page = MagicMock()
    fake_page.extract_text.return_value = "This is a valid resume with enough characters to pass the one hundred character minimum length requirement."
    fake_pdf.__enter__.return_value.pages = [fake_page]
    
    with patch("app.services.pdf_service.pdfplumber.open", return_value=fake_pdf):
        text = extract_text_from_pdf_bytes(b"fakebytes")
        assert "valid resume" in text

def test_extract_text_insufficient_length():
    """Unit test to ensure tiny/corrupt PDFs are rejected cleanly."""
    fake_pdf = MagicMock()
    fake_page = MagicMock()
    fake_page.extract_text.return_value = "Too short."
    fake_pdf.__enter__.return_value.pages = [fake_page]
    
    with patch("app.services.pdf_service.pdfplumber.open", return_value=fake_pdf):
        with pytest.raises(ValueError, match="less than 100 characters"):
            extract_text_from_pdf_bytes(b"fakebytes")
