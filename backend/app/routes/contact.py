"""POST /contact — the contact-form endpoint.

Unauthenticated. Validated. Rate-limited per IP. Honeypot-protected.
Always returns 200 {"ok": true} regardless of whether the message was
actually delivered, so probers / spammers can't learn whether an inbox
exists or rate-limit state. Real failures are logged server-side.
"""

import logging
from typing import Annotated

import redis.asyncio as redis_async
from fastapi import APIRouter, Depends, Request
from sqlalchemy.ext.asyncio import AsyncSession

from app.db import get_db
from app.deps import get_redis
from app.schemas.contact import ContactMessageBody, ContactMessageResponse
from app.services import audit_service, contact_service
from app.services.rate_limit import RateLimitExceeded, check_and_increment

logger = logging.getLogger(__name__)
router = APIRouter(tags=["contact"])


@router.post("/contact", response_model=ContactMessageResponse)
async def submit_contact(
    body: ContactMessageBody,
    request: Request,
    db: Annotated[AsyncSession, Depends(get_db)],
    redis: Annotated[redis_async.Redis, Depends(get_redis)],
) -> ContactMessageResponse:
    client_ip = request.client.host if request.client else "unknown"
    user_agent = request.headers.get("user-agent")

    # Honeypot — bots autofilling all fields tip themselves off.
    if body.website.strip():
        await audit_service.log_event(
            audit_service.EventType.CONTACT_HONEYPOT,
            email=body.email.lower().strip(),
            ip_address=client_ip,
            user_agent=user_agent,
        )
        return ContactMessageResponse(ok=True)

    # Per-IP rate limit: 3 messages per hour. Tight, because real users
    # rarely submit more than once per visit. Same window as the
    # auth/request-link per-IP limit for consistency.
    try:
        await check_and_increment(
            redis,
            f"rl:contact:ip:{client_ip}",
            max_count=3,
            window_seconds=3600,
        )
    except RateLimitExceeded as exc:
        # Same 200 — enumeration prevention.
        await audit_service.log_event(
            audit_service.EventType.RATELIMIT_HIT,
            email=body.email.lower().strip(),
            ip_address=client_ip,
            user_agent=user_agent,
            details={"scope": "contact", "key": str(exc)},
        )
        return ContactMessageResponse(ok=True)

    try:
        await contact_service.send_contact_message(
            name=body.name.strip(),
            sender_email=body.email.lower().strip(),
            message=body.message.strip(),
        )
        await audit_service.log_event(
            audit_service.EventType.CONTACT_SUBMITTED,
            email=body.email.lower().strip(),
            ip_address=client_ip,
            user_agent=user_agent,
        )
    except Exception:
        # Log the failure but still return 200. The visitor doesn't need
        # to know our email infra is having a moment.
        logger.exception("contact send failed", extra={"sender": body.email})

    # `db` is unused today but kept on the signature so adding an
    # audit row later is a one-line change, not a route-signature change.
    _ = db

    return ContactMessageResponse(ok=True)
