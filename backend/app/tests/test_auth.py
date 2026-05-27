"""Quick end-to-end auth smoke test through the FastAPI app.

Talks to the dev DB and Redis (running in compose). Not isolated per-test —
sufficient for v1 confidence.
"""

import uuid

from httpx import AsyncClient

from app.security.cookies import SESSION_COOKIE_NAME


async def test_auth_full_flow(client: AsyncClient) -> None:
    email = f"pytest-{uuid.uuid4().hex[:8]}@example.com"

    # The /auth/request-link body now requires `acceptedPolicyVersion`
    # (the click-wrap consent the user checked on the /auth form).
    # See backend/app/schemas/auth.py:RequestLinkBody.
    resp = await client.post(
        "/auth/request-link",
        json={"email": email, "acceptedPolicyVersion": "test-version"},
    )
    assert resp.status_code == 200
    assert resp.json() == {"ok": True}

    # The magic-link token isn't returned to the client (only to stdout via
    # the email service stub). We have only the token_hash on disk, so we
    # can't complete /auth/verify from a DB read alone. Cover the
    # verify+session path via direct service calls instead.
    from app.db import async_session_factory
    from app.services import auth_service

    async with async_session_factory() as db:
        raw_token = await auth_service.create_magic_link(
            db, email, accepted_policy_version="test-version"
        )

    # /auth/verify sets the session cookie via Set-Cookie. The httpx client
    # stores it on its cookie jar transparently, so subsequent requests
    # carry it without any extra wiring.
    verify = await client.post("/auth/verify", json={"token": raw_token})
    assert verify.status_code == 200
    body = verify.json()
    # The response body deliberately does NOT contain `sessionToken` —
    # it lives only in the HttpOnly cookie.
    assert "sessionToken" not in body
    assert body["user"]["email"] == email
    # Cookie is set, HttpOnly, SameSite=lax.
    set_cookie = verify.headers.get("set-cookie", "")
    assert SESSION_COOKIE_NAME in set_cookie
    assert "HttpOnly" in set_cookie
    assert "SameSite=lax" in set_cookie

    # Subsequent /me uses the cookie automatically — no Authorization header.
    me = await client.get("/me")
    assert me.status_code == 200
    assert me.json()["email"] == email

    out = await client.post("/auth/sign-out")
    assert out.status_code == 204
    # Sign-out should issue a Set-Cookie with Max-Age=0 to clear the jar.
    clear_cookie = out.headers.get("set-cookie", "")
    assert SESSION_COOKIE_NAME in clear_cookie
    assert "Max-Age=0" in clear_cookie

    me_after = await client.get("/me")
    assert me_after.status_code == 401
