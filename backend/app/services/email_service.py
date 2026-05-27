"""Email delivery.

Stub mode (no real RESEND_API_KEY): writes the magic-link URL to stdout
instead of sending email. This is the brief-prescribed dev workflow when
no Resend account is configured.

Real mode renders a branded HTML email with a plain-text fallback. The
HTML template is hand-rolled (no Jinja) because v1 has exactly one
transactional email — a template engine is overkill here.

Design notes:
- Inline CSS only. Gmail strips <style> tags. We keep a single <style>
  block for dark-mode support, but every visible style is also inlined
  so the email renders correctly in the strict clients too.
- System font stack. Renders consistently across macOS Mail, Gmail,
  Outlook, mobile clients.
- Table-based layout. Still the most reliable across email clients.
- Mobile-friendly: max-width 480px, generous padding.
- No personal branding (no domain or social handle in the body) — the
  brand identity stays "Resume Ranker" only.
- Privacy policy URL is built from `settings.app_base_url` so it
  automatically resolves to the correct origin per environment.
"""

import logging
from datetime import UTC, datetime
from html import escape

from app.config import settings

logger = logging.getLogger(__name__)


_SUBJECT = "Sign in to Resume Ranker"
_PREHEADER = "Your sign-in link for Resume Ranker. Valid for 15 minutes, single use."


def _format_request_time(requested_at: datetime) -> str:
    """Render the request timestamp for the email body.

    Backend doesn't reliably know the user's timezone, so UTC is the safe
    universal choice. Format is readable, unambiguous, with the zone label
    so the user can mentally convert if needed.
    """
    if requested_at.tzinfo is None:
        requested_at = requested_at.replace(tzinfo=UTC)
    return requested_at.astimezone(UTC).strftime("%b %d, %Y at %H:%M UTC")


def _privacy_url() -> str:
    """Build the privacy-policy URL for the email footer.

    Uses `settings.app_base_url`, which is the frontend's public origin
    (configured per environment via the APP_BASE_URL env var). The
    frontend serves /privacy as a static route.
    """
    base = settings.app_base_url.rstrip("/")
    return f"{base}/privacy"


def _render_html(magic_link_url: str, requested_at: datetime) -> str:
    """Branded HTML body for the magic-link email.

    `magic_link_url` is HTML-escaped for safety in `href` and visible text.
    URLs we generate shouldn't contain HTML metacharacters anyway, but we
    escape defensively.
    """
    safe_url = escape(magic_link_url, quote=True)
    safe_privacy = escape(_privacy_url(), quote=True)
    safe_time = escape(_format_request_time(requested_at), quote=False)
    return f"""<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="color-scheme" content="light dark">
  <meta name="supported-color-schemes" content="light dark">
  <title>Sign in to Resume Ranker</title>
  <style>
    /* Dark mode opt-in for clients that respect prefers-color-scheme
       (Apple Mail; Gmail handles its own auto-darkening). */
    @media (prefers-color-scheme: dark) {{
      .body {{ background-color: #09090b !important; }}
      .card {{ background-color: #18181b !important; border-color: #27272a !important; }}
      .text-primary {{ color: #fafafa !important; }}
      .text-muted {{ color: #a1a1aa !important; }}
      .button-bg {{ background-color: #fafafa !important; }}
      .button-link {{ color: #09090b !important; }}
      .divider {{ border-color: #27272a !important; }}
      .footer-text {{ color: #71717a !important; }}
      .privacy-link {{ color: #a1a1aa !important; }}
    }}
  </style>
</head>
<body class="body" style="margin:0;padding:0;background-color:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;-webkit-font-smoothing:antialiased;">

  <!-- Pre-header: visible in inbox previews, hidden in the email body -->
  <div style="display:none;max-height:0;overflow:hidden;mso-hide:all;font-size:1px;line-height:1px;color:#f4f4f5;">
    {_PREHEADER}
  </div>

  <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%" style="background-color:#f4f4f5;">
    <tr>
      <td align="center" style="padding:40px 20px;">

        <!-- Main card -->
        <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%" class="card" style="max-width:480px;background-color:#ffffff;border:1px solid #e4e4e7;border-radius:12px;">

          <tr>
            <td style="padding:32px 32px 0 32px;">
              <div class="text-primary" style="font-size:14px;font-weight:600;color:#18181b;letter-spacing:-0.01em;">Resume Ranker</div>
            </td>
          </tr>

          <tr>
            <td style="padding:24px 32px 0 32px;">
              <h1 class="text-primary" style="margin:0;font-size:20px;font-weight:600;color:#18181b;letter-spacing:-0.02em;line-height:1.3;">Sign in to your account</h1>
            </td>
          </tr>

          <tr>
            <td style="padding:12px 32px 0 32px;">
              <p class="text-muted" style="margin:0;font-size:14px;line-height:1.6;color:#52525b;">
                Click the button below to sign in. This link is valid for 15&nbsp;minutes and can be used once.
              </p>
            </td>
          </tr>

          <tr>
            <td style="padding:24px 32px 0 32px;" align="left">
              <table role="presentation" cellpadding="0" cellspacing="0" border="0">
                <tr>
                  <td class="button-bg" style="background-color:#18181b;border-radius:8px;">
                    <a href="{safe_url}" class="button-link" style="display:inline-block;padding:12px 24px;font-size:14px;font-weight:500;color:#ffffff;text-decoration:none;letter-spacing:-0.01em;">
                      Sign in to Resume Ranker
                    </a>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <tr>
            <td style="padding:24px 32px 0 32px;">
              <p class="text-muted" style="margin:0;font-size:12px;line-height:1.5;color:#71717a;">
                Requested on {safe_time}.
              </p>
            </td>
          </tr>

          <tr>
            <td style="padding:32px 32px 0 32px;">
              <div class="divider" style="border-top:1px solid #e4e4e7;line-height:1px;font-size:1px;">&nbsp;</div>
            </td>
          </tr>

          <tr>
            <td style="padding:24px 32px 8px 32px;">
              <p class="text-muted" style="margin:0;font-size:12px;line-height:1.5;color:#71717a;">
                If you didn't request this email, you can safely ignore it. Your account remains secure.
              </p>
            </td>
          </tr>

          <tr>
            <td style="padding:8px 32px 32px 32px;">
              <p class="text-muted" style="margin:0;font-size:12px;line-height:1.5;color:#71717a;">
                This inbox isn't monitored, so please don't reply.
              </p>
            </td>
          </tr>

        </table>

        <!-- Footer below the card -->
        <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%" style="max-width:480px;">
          <tr>
            <td style="padding:16px 32px;text-align:center;">
              <p class="footer-text" style="margin:0 0 4px 0;font-size:11px;line-height:1.5;color:#a1a1aa;">
                Resume Ranker &middot; Section-by-section resume analysis
              </p>
              <p style="margin:0;font-size:11px;line-height:1.5;color:#a1a1aa;">
                <a href="{safe_privacy}" class="privacy-link" style="color:#71717a;text-decoration:underline;">Privacy policy</a>
              </p>
            </td>
          </tr>
        </table>

      </td>
    </tr>
  </table>
</body>
</html>"""


def _render_text(magic_link_url: str, requested_at: datetime) -> str:
    """Plain-text body for clients that don't render HTML, or aggressive spam
    filters that prefer multipart messages with a text alternative.

    The plain-text version keeps the URL inline because there's no button
    primitive in plain text — the URL itself is the action.
    """
    time_str = _format_request_time(requested_at)
    return (
        "Sign in to Resume Ranker\n"
        "\n"
        "Click the link below to sign in. This link is valid for 15 minutes "
        "and can be used once.\n"
        "\n"
        f"{magic_link_url}\n"
        "\n"
        f"Requested on {time_str}.\n"
        "\n"
        "If you didn't request this email, you can safely ignore it. Your "
        "account remains secure.\n"
        "\n"
        "This inbox isn't monitored, so please don't reply.\n"
        "\n"
        "--\n"
        "Resume Ranker — Section-by-section resume analysis\n"
        f"Privacy policy: {_privacy_url()}\n"
    )


async def send_magic_link(
    to_email: str,
    magic_link_url: str,
    requested_at: datetime | None = None,
) -> None:
    # Default the timestamp to now-UTC if the caller didn't supply one,
    # so legacy / test callers still work.
    if requested_at is None:
        requested_at = datetime.now(UTC)

    if settings.email_stub_mode:
        # stdout-friendly format the user can copy-paste during dev.
        print(
            "\n"
            "================ MAGIC LINK (stub mode) ================\n"
            f"To:    {to_email}\n"
            f"Link:  {magic_link_url}\n"
            f"At:    {_format_request_time(requested_at)}\n"
            "(set RESEND_API_KEY in .env to send real emails)\n"
            "========================================================\n",
            flush=True,
        )
        return

    # Real send via Resend. Imported lazily so dev environments without the
    # library installed still work (it's in deps though, just defensive).
    import resend

    resend.api_key = settings.resend_api_key
    resend.Emails.send(
        {
            "from": settings.resend_from_email,
            "to": to_email,
            "subject": _SUBJECT,
            "html": _render_html(magic_link_url, requested_at),
            "text": _render_text(magic_link_url, requested_at),
        }
    )
    logger.info("magic link email sent", extra={"to": to_email})
