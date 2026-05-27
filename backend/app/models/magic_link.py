import uuid
from datetime import datetime

from sqlalchemy import DateTime, ForeignKey, Text
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column

from app.models.base import Base, TimestampMixin, UUIDPkMixin


class MagicLink(Base, UUIDPkMixin, TimestampMixin):
    __tablename__ = "magic_links"

    user_id: Mapped[uuid.UUID | None] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("users.id", ondelete="CASCADE"),
        nullable=True,
    )
    email: Mapped[str] = mapped_column(Text, nullable=False)
    token_hash: Mapped[str] = mapped_column(Text, nullable=False, index=True)
    expires_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False)
    consumed_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)

    # Carries the policy version the requester accepted on /auth, until
    # the magic-link is verified and the version is copied onto the User
    # row. Nullable for the brief window between this migration applying
    # and the new request-link handler shipping (effectively never in
    # production — the same deploy contains both — but safer for tests
    # and partial rollouts).
    accepted_policy_version: Mapped[str | None] = mapped_column(Text, nullable=True)
