from datetime import datetime

from sqlalchemy import DateTime, Text
from sqlalchemy.orm import Mapped, mapped_column

from app.models.base import Base, TimestampMixin, UUIDPkMixin


class User(Base, UUIDPkMixin, TimestampMixin):
    __tablename__ = "users"

    email: Mapped[str] = mapped_column(Text, unique=True, nullable=False, index=True)

    # Click-wrap consent record: which version of ToS+Privacy this user
    # accepted, and when. Set at user creation time (during /auth/verify),
    # copied from the magic_link row that proved email ownership.
    # Nullable for migration backfill of pre-policy users (see Alembic).
    accepted_policy_version: Mapped[str | None] = mapped_column(Text, nullable=True)
    accepted_policy_at: Mapped[datetime | None] = mapped_column(
        DateTime(timezone=True), nullable=True
    )
