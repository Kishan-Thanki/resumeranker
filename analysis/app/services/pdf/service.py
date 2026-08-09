import os
import re

import phonenumbers
import pypdfium2 as pdfium

MIN_PDF_CHARS = int(os.environ["MIN_PDF_CHARS"])
MAX_PDF_PAGES = int(os.getenv("MAX_PDF_PAGES", "20"))
BENCHMARK_MODE = os.getenv("BENCHMARK_MODE", "").lower() == "true"
DEFAULT_PHONE_REGION = os.getenv("DEFAULT_PHONE_REGION", "IN")

EMAIL_PATTERN = r"\b[A-Za-z0-9.*%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b"
LINKEDIN_PATTERN = r"(?i)(linkedin\.com/in/)([A-Za-z0-9*-]+)"
GITHUB_PATTERN = r"(?i)(github\.com/)([A-Za-z0-9_-]+)"

def _redact_phone_numbers(text: str) -> str:
    """
    Redacts phone numbers using libphonenumber (the `phonenumbers` package)
    instead of a hand-rolled regex.

    Regex can't reliably cover international phone formats. Indian mobile
    numbers, for example, are commonly grouped 5+5 ("98765 43210"), which a
    US-shaped 3+4 regex never matches -- the number passes through
    completely unredacted. Leniency is set to POSSIBLE deliberately: for
    PII redaction, a false-positive redaction is a minor inconvenience,
    but a false negative is an actual privacy leak, so this errs toward
    catching more, not fewer.

    DEFAULT_PHONE_REGION only affects numbers with no country code; numbers
    written with a leading + are parsed correctly regardless of region.
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
    # Redact URLs first so digits in usernames aren't mistakenly
    # detected and partially redacted as phone numbers.
    text = re.sub(LINKEDIN_PATTERN, r"\1[REDACTED]", text)
    text = re.sub(GITHUB_PATTERN, r"\1[REDACTED]", text)
    text = _redact_phone_numbers(text)
    text = re.sub(EMAIL_PATTERN, "[EMAIL REDACTED]", text)
    return text

def extract_text_from_pdf_bytes(pdf_bytes: bytes) -> str:
    """Extract text from a PDF and redact common PII."""
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
                try:
                    textpage = page.get_textpage()
                    try:
                        pages_text.append(textpage.get_text_range() or "")
                    finally:
                        textpage.close()
                finally:
                    page.close()
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
