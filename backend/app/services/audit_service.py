"""Audit-log writer.

Appends a row to `audit_events` for security-relevant events. Designed
to be **independent of the caller's transaction**:

- Uses its own DB session via `async_session_factory()` so a parent
  rollback never erases the audit trail.
- Catches every exception so a failure to log NEVER blocks the action
  being audited (e.g., if Postgres is briefly down, we still let the
  request complete and just lose this one audit row).

This matches the security-audit principle: "log what happened, even
when what happened failed." A failed signin attempt is exactly what
we want to capture; an undelivered audit row is acceptable noise.
"""

from __future__ import annotations

import logging
import uuid

from app.db import async_session_factory
from app.models.audit_event import AuditEvent

logger = logging.getLogger(__name__)


# Canonical event-type tags. Kept here as a single source of truth so
# callers reference constants rather than free strings; typos become
# easier to catch in review.
class EventType:
    AUTH_MAGIC_LINK_REQUESTED = "auth.magic_link_requested"
    AUTH_SIGNIN = "auth.signin"
    AUTH_SIGNIN_FAILED = "auth.signin_failed"
    AUTH_SIGNOUT = "auth.signout"
    ACCOUNT_CREATED = "account.created"
    ACCOUNT_DELETED = "account.deleted"
    POLICY_REACCEPTED = "policy.reaccepted"
    RATELIMIT_HIT = "ratelimit.hit"
    CONTACT_SUBMITTED = "contact.submitted"
    CONTACT_HONEYPOT = "contact.honeypot"


async def log_event(
    event_type: str,
    *,
    user_id: uuid.UUID | None = None,
    email: str | None = None,
    ip_address: str | None = None,
    user_agent: str | None = None,
    details: dict[str, object] | None = None,
) -> None:
    """Append one row to `audit_events`. Never raises.

    Uses its own session so the audit row is independent of any caller
    transaction. This is intentional — we want to record `auth.signin_failed`
    even when the parent verify code path is rolling back its own state.

    All callers pass keyword arguments; positional invocation is rejected
    by the function signature (`*`) to prevent argument-order mistakes
    (`user_id` and `email` are both string-like in practice).
    """
    try:
        async with async_session_factory() as session:
            session.add(
                AuditEvent(
                    event_type=event_type,
                    user_id=user_id,
                    email=email.lower() if email else None,
                    ip_address=ip_address,
                    user_agent=user_agent,
                    details=details or {},
                )
            )
            await session.commit()
    except Exception:
        # NEVER let an audit-log failure block the audited action.
        # Log to stdout so ops can see the problem even though the row
        # didn't land.
        logger.exception(
            "audit log write failed",
            extra={"event_type": event_type, "email": email},
        )
