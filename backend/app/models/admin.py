import uuid
from datetime import datetime

from sqlalchemy import Boolean, DateTime, ForeignKey, Text
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.models.base import Base, TimestampMixin, UUIDPkMixin


class AdminRole(Base, UUIDPkMixin, TimestampMixin):
    __tablename__ = "admin_roles"

    name: Mapped[str] = mapped_column(Text, unique=True, nullable=False)
    # E.g. ["frontend_users:*", "settings:read"]
    policies: Mapped[list[str]] = mapped_column(JSONB, nullable=False, server_default="[]")

    admins: Mapped[list["Admin"]] = relationship(back_populates="role")


class Admin(Base, UUIDPkMixin, TimestampMixin):
    __tablename__ = "admins"

    email: Mapped[str] = mapped_column(Text, unique=True, nullable=False, index=True)
    password_hash: Mapped[str] = mapped_column(Text, nullable=False)
    is_active: Mapped[bool] = mapped_column(Boolean, nullable=False, server_default="true")

    role_id: Mapped[uuid.UUID] = mapped_column(ForeignKey("admin_roles.id", ondelete="RESTRICT"), nullable=False)
    role: Mapped["AdminRole"] = relationship(back_populates="admins")

    sessions: Mapped[list["AdminSession"]] = relationship(
        back_populates="admin", cascade="all, delete-orphan"
    )


class AdminSession(Base, UUIDPkMixin, TimestampMixin):
    __tablename__ = "admin_sessions"

    admin_id: Mapped[uuid.UUID] = mapped_column(ForeignKey("admins.id", ondelete="CASCADE"), nullable=False)
    token_hash: Mapped[str] = mapped_column(Text, unique=True, nullable=False, index=True)
    expires_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False)
    revoked_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)

    admin: Mapped["Admin"] = relationship(back_populates="sessions")


class AdminAuditEvent(Base, UUIDPkMixin, TimestampMixin):
    __tablename__ = "admin_audit_events"

    admin_id: Mapped[uuid.UUID | None] = mapped_column(ForeignKey("admins.id", ondelete="SET NULL"), nullable=True)
    action: Mapped[str] = mapped_column(Text, nullable=False, index=True)
    ip_address: Mapped[str] = mapped_column(Text, nullable=False)
    user_agent: Mapped[str | None] = mapped_column(Text, nullable=True)
    # Storing any relevant diffs or metadata here
    details: Mapped[dict | None] = mapped_column(JSONB, nullable=True)
