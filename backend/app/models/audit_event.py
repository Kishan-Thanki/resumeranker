import uuid

from sqlalchemy import ForeignKey, Index, Text
from sqlalchemy.dialects.postgresql import JSONB, UUID
from sqlalchemy.orm import Mapped, mapped_column

from app.models.base import Base, TimestampMixin, UUIDPkMixin


class AuditEvent(Base, UUIDPkMixin, TimestampMixin):
    """Append-only log of security-relevant events.

    Outlives the rows it references — `user_id` uses ON DELETE SET NULL so
    that when a user deletes their account, the audit trail of THAT very
    deletion (and everything leading up to it) survives. The `email`
    column is duplicated alongside `user_id` for exactly the same reason:
    once `user_id` is nulled out, the email keeps the row attributable to
    a recognizable identifier during incident response.

    `details` is a free-form JSONB so per-event-type extras (failure
    reasons, rate-limit keys, etc.) can be added without schema changes.

    Read pattern: `psql` queries scoped by event_type + time range during
    incident response. Indexes are tuned for that — most-common queries
    are "show me all of TYPE in the last N hours" or "everything tied to
    EMAIL in the last N days".
    """

    __tablename__ = "audit_events"

    # Short tag for the kind of event. Examples (kept in code, not a FK):
    #   auth.magic_link_requested      magic-link form submission
    #   auth.signin                    successful verify → session created
    #   auth.signin_failed             verify rejected (not-found / expired / consumed)
    #   auth.signout                   user explicitly signed out
    #   account.created                first time we saw this email
    #   account.deleted                hard-delete via DELETE /me
    #   ratelimit.hit                  any rate-limit threshold tripped
    #   contact.submitted              /contact message accepted
    #   contact.honeypot               /contact bot-fill detected
    event_type: Mapped[str] = mapped_column(Text, nullable=False)

    # Nullable because (a) failed signins have no known user_id, and
    # (b) the FK uses ON DELETE SET NULL so the row survives the user.
    user_id: Mapped[uuid.UUID | None] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("users.id", ondelete="SET NULL"),
        nullable=True,
    )

    # Always set when an email is involved (request-link / verify /
    # signup / delete). Persisted even after `user_id` is nulled so we
    # can still attribute the event to a human.
    email: Mapped[str | None] = mapped_column(Text, nullable=True)

    # Best-effort client identification. Stored as plain text rather than
    # INET because we may see X-Forwarded-For with a list, and CIDR
    # querying isn't a use case here.
    ip_address: Mapped[str | None] = mapped_column(Text, nullable=True)

    user_agent: Mapped[str | None] = mapped_column(Text, nullable=True)

    # Free-form per-event-type extras. Example payloads:
    #   {"reason": "expired"}           on auth.signin_failed
    #   {"key": "rl:auth:email:...",    on ratelimit.hit
    #    "scope": "email", "max": 5}
    details: Mapped[dict[str, object]] = mapped_column(
        JSONB,
        nullable=False,
        default=dict,
        server_default="{}",
    )

    __table_args__ = (
        # Most common query: "all events of type X, newest first"
        Index("ix_audit_events_type_created_at", "event_type", "created_at"),
        # Second most common: "everything tied to this email, newest first"
        Index("ix_audit_events_email_created_at", "email", "created_at"),
        # User-scoped audit (when user_id is non-null and still maps to a user)
        Index("ix_audit_events_user_id_created_at", "user_id", "created_at"),
    )
