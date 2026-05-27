import logging
from datetime import UTC, datetime
from typing import Annotated

import redis.asyncio as redis_async
from fastapi import APIRouter, Depends, HTTPException, Request, Response, status
from sqlalchemy.ext.asyncio import AsyncSession

from app.config import settings
from app.db import get_db
from app.deps import get_current_user, get_redis, get_session_token
from app.models.user import User
from app.schemas.auth import (
    AcceptPolicyBody,
    MeResponse,
    RequestLinkBody,
    RequestLinkResponse,
    UserPublic,
    VerifyBody,
    VerifyResponse,
)
from app.security.cookies import clear_session_cookie, set_session_cookie
from app.services import audit_service, auth_service
from app.services.email_service import send_magic_link
from app.services.rate_limit import RateLimitExceeded, check_and_increment

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/auth", tags=["auth"])
me_router = APIRouter(tags=["auth"])


@router.post("/request-link", response_model=RequestLinkResponse)
async def request_link(
    body: RequestLinkBody,
    request: Request,
    db: Annotated[AsyncSession, Depends(get_db)],
    redis: Annotated[redis_async.Redis, Depends(get_redis)],
) -> RequestLinkResponse:
    # Anti-abuse: 5 req/hour per email AND 20 req/hour per IP.
    email = body.email.lower()
    client_ip = request.client.host if request.client else "unknown"
    user_agent = request.headers.get("user-agent")
    try:
        await check_and_increment(redis, f"rl:auth:email:{email}", max_count=5, window_seconds=3600)
        await check_and_increment(
            redis, f"rl:auth:ip:{client_ip}", max_count=20, window_seconds=3600
        )
    except RateLimitExceeded as exc:
        # Same 200 response — never reveal rate-limit state to callers.
        # But DO log it server-side so abuse patterns are discoverable
        # via the audit table.
        await audit_service.log_event(
            audit_service.EventType.RATELIMIT_HIT,
            email=email,
            ip_address=client_ip,
            user_agent=user_agent,
            details={"scope": "auth.request_link", "key": str(exc)},
        )
        return RequestLinkResponse(ok=True)

    await audit_service.log_event(
        audit_service.EventType.AUTH_MAGIC_LINK_REQUESTED,
        email=email,
        ip_address=client_ip,
        user_agent=user_agent,
    )
    raw_token = await auth_service.create_magic_link(
        db, email, accepted_policy_version=body.accepted_policy_version
    )
    magic_url = f"{settings.app_base_url}/auth/verify?token={raw_token}"
    try:
        await send_magic_link(email, magic_url, requested_at=datetime.now(UTC))
    except Exception:
        logger.exception("magic-link email send failed", extra={"email": email})
        # Still 200 — enumeration prevention.

    return RequestLinkResponse(ok=True)


@router.post("/verify", response_model=VerifyResponse)
async def verify(
    body: VerifyBody,
    request: Request,
    response: Response,
    db: Annotated[AsyncSession, Depends(get_db)],
) -> VerifyResponse:
    client_ip = request.client.host if request.client else "unknown"
    user_agent = request.headers.get("user-agent")
    try:
        user = await auth_service.verify_magic_link(db, body.token)
    except auth_service.AuthError as exc:
        # Auditing failed verifies is the whole point of an audit log —
        # repeated failures from one IP signal brute force or scraper.
        await audit_service.log_event(
            audit_service.EventType.AUTH_SIGNIN_FAILED,
            ip_address=client_ip,
            user_agent=user_agent,
            details={"reason": "invalid_link"},
        )
        # Single generic error so callers can't distinguish not-found / expired / consumed.
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="invalid link") from exc

    session_token = await auth_service.create_session(db, user)
    set_session_cookie(response, session_token)
    return VerifyResponse(user=UserPublic.model_validate(user))


@router.post("/sign-out", status_code=status.HTTP_204_NO_CONTENT)
async def sign_out(
    response: Response,
    db: Annotated[AsyncSession, Depends(get_db)],
    token: Annotated[str, Depends(get_session_token)],
) -> None:
    await auth_service.revoke_session(db, token)
    clear_session_cookie(response)


@me_router.get("/me", response_model=MeResponse)
async def me(user: Annotated[User, Depends(get_current_user)]) -> User:
    return user


@me_router.post("/me/accept-policy", response_model=MeResponse)
async def accept_policy(
    body: AcceptPolicyBody,
    user: Annotated[User, Depends(get_current_user)],
    db: Annotated[AsyncSession, Depends(get_db)],
) -> User:
    """Re-acceptance flow: user confirms the new ToS/Privacy version.

    The frontend hits this when it detects the user's stored
    `acceptedPolicyVersion` is older than its own `CURRENT_POLICY_VERSION`.
    Returns the updated user record so the frontend can dismiss the modal
    and stop blocking the rest of the app.
    """
    return await auth_service.accept_policy(db, user, body.accepted_policy_version)


@me_router.delete("/me", status_code=status.HTTP_204_NO_CONTENT)
async def delete_me(
    response: Response,
    user: Annotated[User, Depends(get_current_user)],
    db: Annotated[AsyncSession, Depends(get_db)],
) -> None:
    """Hard-delete the authenticated user and all their data.

    GDPR / DPDP / CCPA "right to erasure". Cascades to sessions,
    magic_links, and analyses via FK ON DELETE CASCADE. We also clear
    the session cookie on the response — the cookie's token belongs to
    a session row that no longer exists and would return 401 on the
    next authenticated request anyway, but clearing it explicitly
    avoids one round-trip's worth of confusion in the client.

    Logs the deletion event with the email so we have an auditable
    record of the request, then the email itself is gone from the
    database. Per `AGENTS.md` rule #8 logging an email is allowed.
    """
    logger.info("account deletion requested", extra={"email": user.email})
    await auth_service.delete_user(db, user)
    clear_session_cookie(response)
