"""Contact-form delivery.

When a visitor submits the /contact form, the message is delivered to
the admin inbox configured by `CONTACT_TO_EMAIL`. The visitor never
sees that address — they just get a generic "thanks" response.

Stub mode (no real `CONTACT_TO_EMAIL` or no `RESEND_API_KEY`) writes
the message to stdout instead, so local dev works without sending
real email.
"""

import logging
from html import escape

from app.config import settings

logger = logging.getLogger(__name__)


_SUBJECT = "Resume Ranker — Contact form message"


def _contact_from_email() -> str:
    """Build a contact-form-specific From address.

    Derives `contact@<domain>` from the verified `RESEND_FROM_EMAIL`
    domain (e.g. if RESEND_FROM_EMAIL is `login@mail.example.com`,
    we send contact-form forwards from `contact@mail.example.com`).
    No new DNS / Resend verification needed — Resend lets us use any
    local-part on a verified subdomain.

    Why a separate from-address: the same `RESEND_FROM_EMAIL` is used
    by magic-link auth emails, where `login@` is semantically right.
    Reusing it for contact-form forwards puts a misleading "login@"
    label on a non-login email — cosmetic but ugly. This helper gives
    contact-form mail a more accurate label without adding env vars.
    """
    _local, _, domain = settings.resend_from_email.partition("@")
    # Defensive: fall back to the configured from-address if parsing
    # fails (e.g. someone set an env var without an @).
    if not domain:
        return settings.resend_from_email
    return f"contact@{domain}"


def _contact_stub_mode() -> bool:
    """True when either the admin destination or Resend itself is in stub
    mode. Either way we can't deliver the message, so we log it."""
    return settings.email_stub_mode or settings.contact_to_email.strip() in {
        "",
        "replace-me",
        "your-key-here",
    }


def _render_admin_html(name: str, sender_email: str, message: str) -> str:
    """Plain branded HTML for the admin's inbox. The visitor's input is
    HTML-escaped because we're injecting it into an HTML body."""
    return f"""<!doctype html>
<html lang="en">
<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#f4f4f5;margin:0;padding:24px;">
  <div style="max-width:520px;margin:0 auto;background:#ffffff;border:1px solid #e4e4e7;border-radius:12px;padding:24px;">
    <div style="font-size:13px;color:#71717a;letter-spacing:-0.01em;">Resume Ranker · Contact form</div>
    <h1 style="margin:8px 0 16px 0;font-size:18px;font-weight:600;color:#18181b;letter-spacing:-0.02em;">New message from {escape(name)}</h1>
    <table style="font-size:13px;color:#52525b;line-height:1.6;border-collapse:collapse;width:100%;">
      <tr><td style="padding:4px 0;color:#71717a;width:80px;">From:</td><td style="padding:4px 0;color:#18181b;">{escape(name)} &lt;{escape(sender_email)}&gt;</td></tr>
    </table>
    <div style="margin-top:16px;padding:16px;background:#f4f4f5;border-radius:8px;font-size:14px;line-height:1.6;color:#18181b;white-space:pre-wrap;">{escape(message)}</div>
    <p style="margin-top:24px;font-size:12px;color:#a1a1aa;">Reply directly to this email to respond — Reply-To is set to the visitor's address.</p>
  </div>
</body>
</html>"""


def _render_admin_text(name: str, sender_email: str, message: str) -> str:
    return (
        "Resume Ranker — Contact form message\n"
        f"\nFrom: {name} <{sender_email}>\n"
        "\n--- Message ---\n"
        f"{message}\n"
        "\n--- End ---\n"
        "\nReply directly to this email to respond.\n"
    )


async def send_contact_message(name: str, sender_email: str, message: str) -> None:
    """Deliver a contact-form message to the admin inbox.

    `sender_email` is whoever filled in the form. `name` and `message`
    are their inputs (must be escaped before HTML rendering — handled
    in `_render_admin_html`). The visitor's email is set as `reply_to`
    so the admin's reply goes straight back to them, but they never
    see the admin address.
    """
    if _contact_stub_mode():
        print(
            "\n"
            "================ CONTACT FORM (stub mode) ================\n"
            f"From:    {name} <{sender_email}>\n"
            f"To:      {settings.contact_to_email}\n"
            "Message:\n"
            f"{message}\n"
            "(set CONTACT_TO_EMAIL + RESEND_API_KEY in .env to send real emails)\n"
            "==========================================================\n",
            flush=True,
        )
        return

    import resend

    resend.api_key = settings.resend_api_key
    resend.Emails.send(
        {
            "from": _contact_from_email(),
            "to": settings.contact_to_email,
            "reply_to": sender_email,
            "subject": _SUBJECT,
            "html": _render_admin_html(name, sender_email, message),
            "text": _render_admin_text(name, sender_email, message),
        }
    )
    # Never log message content (could be sensitive). Log only the routing.
    logger.info("contact message forwarded", extra={"sender": sender_email})
