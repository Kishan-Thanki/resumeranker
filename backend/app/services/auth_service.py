"""Auth service — no FastAPI imports.

Magic-link generation and verification, session creation, revocation, lookup.
Tokens are 32 random bytes, base64url-encoded. We store only SHA-256 hashes.
"""

import hashlib
import secrets
from datetime import UTC, datetime, timedelta

from sqlalchemy import select, update
from sqlalchemy.ext.asyncio import AsyncSession

from app.models.magic_link import MagicLink
from app.models.session import Session
from app.models.user import User
from app.services import audit_service

MAGIC_LINK_TTL = timedelta(minutes=15)
SESSION_TTL = timedelta(days=30)


class AuthError(Exception):
    """Raised when a magic link or session token is invalid. All failure modes
    bubble up as the same exception so callers can return a single generic
    error message and avoid leaking which condition failed."""


def _generate_token() -> str:
    return secrets.token_urlsafe(32)


def _hash_token(raw: str) -> str:
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()


def _now() -> datetime:
    return datetime.now(UTC)


async def _get_or_create_user(
    db: AsyncSession,
    email: str,
    accepted_policy_version: str | None = None,
) -> tuple[User, bool]:
    """Fetch the user by email, creating them if missing.

    When the user is newly created and `accepted_policy_version` is
    provided (carried from the verified magic-link row), we stamp the
    click-wrap consent record on the User row at creation time. Existing
    users keep their previously-recorded acceptance — we don't overwrite
    on every sign-in.
    """
    result = await db.execute(select(User).where(User.email == email))
    user = result.scalar_one_or_none()
    is_new = False
    if user is None:
        is_new = True
        user = User(email=email)
        if accepted_policy_version is not None:
            user.accepted_policy_version = accepted_policy_version
            user.accepted_policy_at = _now()
        db.add(user)
        await db.flush()
    return user, is_new


async def create_magic_link(
    db: AsyncSession,
    email: str,
    accepted_policy_version: str,
) -> str:
    """Returns the raw token. Caller is responsible for emailing it.

    `accepted_policy_version` is the policy version the user just
    affirmatively accepted on the /auth form. It travels with the magic
    link so that on verification we can stamp it onto the (possibly
    newly-created) User row.
    """
    raw = _generate_token()
    link = MagicLink(
        email=email,
        token_hash=_hash_token(raw),
        expires_at=_now() + MAGIC_LINK_TTL,
        accepted_policy_version=accepted_policy_version,
    )
    db.add(link)
    await db.commit()
    return raw


async def verify_magic_link(db: AsyncSession, raw_token: str) -> User:
    """Validate, consume, and return the associated user. Raises AuthError on
    any failure (not found / expired / consumed) — same exception type for all
    cases so we don't leak which condition failed."""
    token_hash = _hash_token(raw_token)
    result = await db.execute(select(MagicLink).where(MagicLink.token_hash == token_hash))
    link = result.scalar_one_or_none()
    if link is None:
        raise AuthError("invalid")
    now = _now()
    if link.consumed_at is not None:
        raise AuthError("invalid")
    if link.expires_at < now:
        raise AuthError("invalid")

    user, is_new = await _get_or_create_user(
        db, link.email, accepted_policy_version=link.accepted_policy_version
    )
    link.user_id = user.id
    link.consumed_at = now
    await db.commit()

    if is_new:
        await audit_service.log_event(
            audit_service.EventType.ACCOUNT_CREATED,
            user_id=user.id,
            email=link.email,
            details={"accepted_policy_version": link.accepted_policy_version},
        )

    return user


async def create_session(db: AsyncSession, user: User) -> str:
    raw = _generate_token()
    session = Session(
        user_id=user.id,
        token_hash=_hash_token(raw),
        expires_at=_now() + SESSION_TTL,
    )
    db.add(session)
    await db.commit()
    await audit_service.log_event(
        audit_service.EventType.AUTH_SIGNIN,
        user_id=user.id,
        email=user.email,
    )
    return raw


async def get_session_user(db: AsyncSession, raw_token: str) -> User | None:
    """Look up an active session, refresh sliding window, return the user.
    Returns None for any invalid/expired/revoked token."""
    token_hash = _hash_token(raw_token)
    result = await db.execute(
        select(Session, User)
        .join(User, Session.user_id == User.id)
        .where(Session.token_hash == token_hash)
    )
    row = result.first()
    if row is None:
        return None
    session: Session = row[0]
    user: User = row[1]
    now = _now()
    if session.revoked_at is not None or session.expires_at < now:
        return None
    # Sliding refresh
    session.expires_at = now + SESSION_TTL
    await db.commit()
    return user


async def revoke_session(db: AsyncSession, raw_token: str) -> None:
    token_hash = _hash_token(raw_token)
    # Look up the user_id BEFORE revoking so the audit row captures who
    # signed out. Done as a separate query for clarity; the cost is one
    # extra index lookup per signout (negligible).
    lookup = await db.execute(select(Session.user_id).where(Session.token_hash == token_hash))
    row = lookup.first()
    user_id = row[0] if row else None

    await db.execute(
        update(Session)
        .where(Session.token_hash == token_hash, Session.revoked_at.is_(None))
        .values(revoked_at=_now())
    )
    await db.commit()
    await audit_service.log_event(
        audit_service.EventType.AUTH_SIGNOUT,
        user_id=user_id,
    )


async def accept_policy(db: AsyncSession, user: User, version: str) -> User:
    """Record the user's re-acceptance of a (presumably newer) policy version.

    Used by the re-acceptance modal flow: when the frontend's bundled
    `CURRENT_POLICY_VERSION` is newer than the user's stored
    `accepted_policy_version`, the modal forces a re-click before any
    other UI works. This endpoint persists that click.

    The version string is stored verbatim — caller validates length and
    shape via Pydantic. Both `accepted_policy_version` and
    `accepted_policy_at` advance; the timestamp captures when the latest
    consent was recorded.
    """
    user.accepted_policy_version = version
    user.accepted_policy_at = _now()
    await db.commit()
    await audit_service.log_event(
        audit_service.EventType.POLICY_REACCEPTED,
        user_id=user.id,
        email=user.email,
        details={"version": version},
    )
    return user


async def delete_user(db: AsyncSession, user: User) -> None:
    """Hard-delete a user and all their data.

    Used for the GDPR/DPDP/CCPA "right to erasure" flow exposed via
    DELETE /me. We rely on `ON DELETE CASCADE` FKs to clean up child
    rows: sessions, magic_links, and analyses are all wiped automatically
    when the user row is removed.

    Hard delete (not soft) — the privacy policy promises the data is
    gone; soft-delete with `deleted_at` would technically still retain it.
    """
    # Capture identifiers before the row is gone; the audit_events FK
    # uses ON DELETE SET NULL so `user_id` survives but we lose the
    # ability to look up the email from the user row.
    user_id = user.id
    email = user.email
    await db.delete(user)
    await db.commit()
    await audit_service.log_event(
        audit_service.EventType.ACCOUNT_DELETED,
        user_id=user_id,
        email=email,
    )
