import uuid
from datetime import datetime

from sqlalchemy import DateTime, ForeignKey, Index, Text, func
from sqlalchemy.dialects.postgresql import JSONB, UUID
from sqlalchemy.orm import Mapped, mapped_column

from app.models.base import Base, TimestampMixin, UUIDPkMixin
from app.models.encrypted import EncryptedText


class Analysis(Base, UUIDPkMixin, TimestampMixin):
    __tablename__ = "analyses"

    user_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("users.id", ondelete="CASCADE"),
        nullable=False,
    )
    status: Mapped[str] = mapped_column(Text, nullable=False)
    jd_title: Mapped[str] = mapped_column(Text, nullable=False)
    resume_name: Mapped[str] = mapped_column(Text, nullable=False)
    # Encrypted at rest via Fernet; see app.models.encrypted.EncryptedText.
    # Stored as TEXT (ciphertext is urlsafe-base64) so the migration is a
    # type swap, not a column rebuild.
    jd_text: Mapped[str] = mapped_column(EncryptedText, nullable=False)
    resume_text: Mapped[str] = mapped_column(EncryptedText, nullable=False)
    sections: Mapped[list[dict[str, object]]] = mapped_column(
        JSONB,
        nullable=False,
        default=list,
        server_default="[]",
    )
    error_message: Mapped[str | None] = mapped_column(Text, nullable=True)
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        server_default=func.now(),
        onupdate=func.now(),
        nullable=False,
    )

    __table_args__ = (Index("ix_analyses_user_id_created_at", "user_id", "created_at"),)
