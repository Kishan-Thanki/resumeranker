import os
import re
from contextlib import ExitStack

import phonenumbers
import pypdfium2 as pdfium

MIN_PDF_CHARS = int(os.getenv("MIN_PDF_CHARS", "100"))
MAX_PDF_PAGES = int(os.getenv("MAX_PDF_PAGES", "20"))
BENCHMARK_MODE = os.getenv("BENCHMARK_MODE", "").lower() == "true"
DEFAULT_PHONE_REGION = os.getenv("DEFAULT_PHONE_REGION", "IN")

EMAIL_PATTERN = re.compile(r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b")
LINKEDIN_PATTERN = re.compile(r"(?i)(linkedin\.com/in/)([A-Za-z0-9_-]+)")
GITHUB_PATTERN = re.compile(r"(?i)(github\.com/)([A-Za-z0-9_-]+)")


def _redact_phone_numbers(text: str) -> str:
    """
    Redacts phone numbers using libphonenumber (the `phonenumbers` package).
    Processes matches in reverse order so string offsets remain valid.
    """
    matches = list(
        phonenumbers.PhoneNumberMatcher(
            text,
            DEFAULT_PHONE_REGION,
            leniency=phonenumbers.Leniency.POSSIBLE,
        )
    )
    for match in reversed(matches):
        text = text[: match.start] + "[PHONE REDACTED]" + text[match.end :]
    return text


def scrub_pii(text: str) -> str:
    """Redacts personally identifiable information from extracted text."""
    text = LINKEDIN_PATTERN.sub(r"\1[REDACTED]", text)
    text = GITHUB_PATTERN.sub(r"\1[REDACTED]", text)
    text = _redact_phone_numbers(text)
    text = EMAIL_PATTERN.sub("[EMAIL REDACTED]", text)
    return text


def extract_text_from_pdf_bytes(pdf_bytes: bytes) -> str:
    """Extract text safely from PDF bytes using pypdfium2 and redact common PII."""
    if BENCHMARK_MODE:
        return (
            "Benchmark dummy text covering the minimum character "
            "requirement for performance testing. " * 10
        )

    pages_text: list[str] = []
    try:
        with pdfium.PdfDocument(pdf_bytes) as pdf:
            page_count = len(pdf)
            if page_count > MAX_PDF_PAGES:
                raise ValueError(
                    f"PDF has too many pages (found {page_count}, "
                    f"maximum {MAX_PDF_PAGES}). The document exceeds the allowed limit."
                )

            for page in pdf:
                with ExitStack() as stack:
                    stack.callback(page.close)
                    textpage = page.get_textpage()
                    stack.callback(textpage.close)
                    pages_text.append(textpage.get_text_range() or "")
    except ValueError:
        raise
    except Exception as e:
        raise ValueError(
            "Could not process PDF (it may be corrupted, password-protected, "
            f"or not a valid PDF): {e}"
        ) from e

    cleaned_text = "\n".join(pages_text).strip()
    if len(cleaned_text) < MIN_PDF_CHARS:
        raise ValueError(
            f"PDF contains insufficient text (found {len(cleaned_text)}, "
            f"required {MIN_PDF_CHARS})."
        )
    return scrub_pii(cleaned_text)
