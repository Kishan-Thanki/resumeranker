"""PDF text extraction via pdfplumber."""

import io
import re

import pdfplumber

MIN_EXTRACTED_CHARS = 100


class PdfExtractionError(Exception):
    """Raised when the PDF can't be parsed or yields too little text to use."""


def extract_text_from_pdf(file_bytes: bytes) -> str:
    """Extract text from a PDF blob, normalize whitespace, return one string.

    Raises PdfExtractionError on parse failure or if the result has fewer than
    100 chars (we assume image-only / scanned PDFs in v1 and reject them).
    """
    try:
        pages: list[str] = []
        with pdfplumber.open(io.BytesIO(file_bytes)) as pdf:
            for page in pdf.pages:
                page_text = page.extract_text() or ""
                if page_text.strip():
                    pages.append(page_text)
    except Exception as exc:
        raise PdfExtractionError(f"could not parse PDF: {exc}") from exc

    combined = "\n\n".join(pages)
    # Collapse runs of whitespace but preserve paragraph breaks.
    combined = re.sub(r"[ \t]+", " ", combined)
    combined = re.sub(r"\n{3,}", "\n\n", combined).strip()

    if len(combined) < MIN_EXTRACTED_CHARS:
        raise PdfExtractionError(
            "PDF text too short — the file may be image-based (scanned) "
            "or empty. v1 doesn't support OCR; upload a text-based PDF."
        )
    return combined
