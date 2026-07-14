import io
import re
import pypdfium2 as pdfium

def scrub_pii(text: str) -> str:
    """Aggressively removes emails and phone numbers while keeping links."""
    email_pattern = r'\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b'
    text = re.sub(email_pattern, '[EMAIL REDACTED]', text)
    
    phone_pattern = r'(\+?\d{1,3}[\s.-]?)?(\(?\d{3}\)?[\s.-]?)?\d{3}[\s.-]?\d{4}\b'
    text = re.sub(phone_pattern, '[PHONE REDACTED]', text)
    
    linkedin_pattern = r'(?i)(linkedin\.com/in/)([a-zA-Z0-9_-]+)'
    text = re.sub(linkedin_pattern, r'\1[REDACTED]', text)
    
    github_pattern = r'(?i)(github\.com/)([a-zA-Z0-9_-]+)'
    text = re.sub(github_pattern, r'\1[REDACTED]', text)
    
    return text

def extract_text_from_pdf_bytes(pdf_bytes: bytes) -> str:
    """Extracts text using pypdfium2 directly from raw bytes and scrubs PII."""
    pages_text = []
    
    with pdfium.PdfDocument(pdf_bytes) as pdf:
        for i in range(len(pdf)):
            page = pdf[i]
            textpage = page.get_textpage()
            text = textpage.get_text_range()
            pages_text.append(text or "")
                
    cleaned_text = "\n".join(pages_text).strip()
    if len(cleaned_text) < 100:
        raise ValueError("PDF contains insufficient text for analysis (less than 100 characters).")
        
    return scrub_pii(cleaned_text)
