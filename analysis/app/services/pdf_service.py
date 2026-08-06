import os
import re

import pypdfium2 as pdfium

MIN_PDF_CHARS = int(os.environ["MIN_PDF_CHARS"])

EMAIL_PATTERN = r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b"
PHONE_PATTERN = r"(\+?\d{1,3}[\s.-]?)?(\(?\d{3}\)?[\s.-]?)?\d{3}[\s.-]?\d{4}\b"
LINKEDIN_PATTERN = r"(?i)(linkedin\.com/in/)([A-Za-z0-9_-]+)"
GITHUB_PATTERN = r"(?i)(github\.com/)([A-Za-z0-9_-]+)"


def scrub_pii(text: str) -> str:
    """Redacts personally identifiable information from extracted text."""
    text = re.sub(EMAIL_PATTERN, "[EMAIL REDACTED]", text)
    text = re.sub(PHONE_PATTERN, "[PHONE REDACTED]", text)
    text = re.sub(LINKEDIN_PATTERN, r"\1[REDACTED]", text)
    text = re.sub(GITHUB_PATTERN, r"\1[REDACTED]", text)
    return text


def extract_text_from_pdf_bytes(pdf_bytes: bytes) -> str:
    """Extract text from a PDF and redact common PII."""

    pages_text: list[str] = []

    with pdfium.PdfDocument(pdf_bytes) as pdf:
        for page in pdf:
            textpage = page.get_textpage()
            try:
                pages_text.append(textpage.get_text_range() or "")
            finally:
                textpage.close()

    cleaned_text = "\n".join(pages_text).strip()

    if len(cleaned_text) < MIN_PDF_CHARS:
        raise ValueError(
            f"PDF contains insufficient text (found {len(cleaned_text)}, "
            f"required {MIN_PDF_CHARS})."
        )

    return scrub_pii(cleaned_text)
