from pydantic import BaseModel, EmailStr, Field


class ContactMessageBody(BaseModel):
    """POST /contact request body.

    All three primary fields are required. The `website` field is a
    honeypot — humans see no input for it on the form; bots
    autofilling all visible fields plus invisible ones will leak. Any
    non-empty value gets silently rejected by the route.
    """

    name: str = Field(min_length=1, max_length=100)
    email: EmailStr
    # Keep the cap generous enough for a real support request, narrow
    # enough that a malicious payload can't flood our email provider.
    message: str = Field(min_length=10, max_length=5000)
    # Honeypot — must remain empty. Not rendered as an input on the
    # form. Bots that fill every field tip themselves off.
    website: str = ""


class ContactMessageResponse(BaseModel):
    """We return the same shape on success and rate-limit / honeypot
    rejection — visitors never learn whether their submission actually
    reached an inbox. Spam-prevention by-design."""

    ok: bool = True
