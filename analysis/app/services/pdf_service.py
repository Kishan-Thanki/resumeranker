import io
import pdfplumber

def extract_text_from_pdf_bytes(pdf_bytes: bytes) -> str:
    """Extracts text using pdfplumber directly from raw bytes."""
    with pdfplumber.open(io.BytesIO(pdf_bytes)) as pdf:
        pages_text = [page.extract_text() or "" for page in pdf.pages]
                
    cleaned_text = "\n".join(pages_text).strip()
    if len(cleaned_text) < 100:
        raise ValueError("PDF contains insufficient text for analysis (less than 100 characters).")
        
    return cleaned_text
