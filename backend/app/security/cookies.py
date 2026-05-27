"""Session cookie helpers.

All session-token plumbing lives behind these helpers so the routes can
stay declarative ("set the cookie", "clear the cookie") without having
to remember the security attributes every time. Centralizing also means
we only have one place to audit when the cookie policy changes.
"""

from __future__ import annotations

from fastapi import Response

from app.config import settings
from app.services.auth_service import SESSION_TTL

# Name kept simple and generic — no framework signal, no client-visible
# usage (HttpOnly hides it from JS). Matches the historical localStorage
# key conceptually, just lives in a different store.
SESSION_COOKIE_NAME = "session"


def set_session_cookie(response: Response, raw_token: str) -> None:
    """Attach the session cookie to `response` with hardened attributes.

    Attribute rationale:
    - HttpOnly: blocks `document.cookie` access, so an XSS injection
      cannot exfiltrate the token. This is the whole reason for moving
      off localStorage.
    - Secure (prod only): browser refuses to send the cookie over plain
      HTTP. Gated on `environment != "development"` so local-dev still
      works on http://localhost.
    - SameSite=Lax: blocks the cookie on cross-site POST/DELETE,
      neutralizing CSRF for the state-changing endpoints without
      breaking the magic-link verify flow (the user clicks a link in
      their email which is a top-level GET → Lax permits it).
    - Path=/: cookie sent to every endpoint on the API origin.
    - Max-Age = SESSION_TTL: matches the server-side row expiry. The
      sliding-window refresh in `get_session_user` extends the DB-side
      expiry but the browser cookie keeps its original Max-Age; this is
      fine because expired-but-still-presented cookies are validated
      against the (refreshed) DB row.
    """
    response.set_cookie(
        key=SESSION_COOKIE_NAME,
        value=raw_token,
        max_age=int(SESSION_TTL.total_seconds()),
        httponly=True,
        secure=settings.environment != "development",
        samesite="lax",
        path="/",
    )


def clear_session_cookie(response: Response) -> None:
    """Tell the browser to drop the session cookie.

    Sets Max-Age=0 with the same path. Browsers require the same Path
    (and Domain, if used) on deletion as on creation — we use Path=/
    for both, no Domain, so they match.
    """
    response.delete_cookie(
        key=SESSION_COOKIE_NAME,
        path="/",
    )
