"""encrypt existing resume and jd text rows

Revision ID: d62b68a21d7b
Revises: 524f3d7eb66b
Create Date: 2026-05-24 17:22:51.355396

Data-only migration. The column type is still TEXT — switching from
sa.Text to app.models.encrypted.EncryptedText is a Python-level type
swap, not a schema change. This migration walks the existing rows and
encrypts `resume_text` and `jd_text` in place, so reads after the type
swap go through the same code path regardless of when the row was
created.

Idempotent: rows that already look like Fernet ciphertext (decrypt
cleanly) are left untouched. Safe to re-run.

Skip-mode: if RESUME_ENCRYPTION_KEY is unset (stub mode), this is a
no-op. Local dev / CI without a key still upgrades cleanly.
"""
from typing import Sequence, Union

from alembic import op
from cryptography.fernet import Fernet, InvalidToken
from sqlalchemy import text

from app.config import settings


# revision identifiers, used by Alembic.
revision: str = 'd62b68a21d7b'
down_revision: Union[str, None] = '524f3d7eb66b'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    if settings.encryption_stub_mode:
        # No key configured — leave rows as plaintext. The EncryptedText
        # type passes values through unchanged in stub mode.
        return

    cipher = Fernet(settings.resume_encryption_key.encode("utf-8"))
    bind = op.get_bind()
    rows = bind.execute(text("SELECT id, jd_text, resume_text FROM analyses")).fetchall()

    for row_id, jd_text, resume_text in rows:
        encrypted_jd = _encrypt_if_plaintext(cipher, jd_text)
        encrypted_resume = _encrypt_if_plaintext(cipher, resume_text)
        if encrypted_jd is None and encrypted_resume is None:
            continue
        bind.execute(
            text(
                "UPDATE analyses SET "
                "jd_text = COALESCE(:jd, jd_text), "
                "resume_text = COALESCE(:resume, resume_text) "
                "WHERE id = :id"
            ),
            {"id": row_id, "jd": encrypted_jd, "resume": encrypted_resume},
        )


def downgrade() -> None:
    if settings.encryption_stub_mode:
        return

    cipher = Fernet(settings.resume_encryption_key.encode("utf-8"))
    bind = op.get_bind()
    rows = bind.execute(text("SELECT id, jd_text, resume_text FROM analyses")).fetchall()

    for row_id, jd_text, resume_text in rows:
        try:
            jd_plain = cipher.decrypt(jd_text.encode("ascii")).decode("utf-8")
            resume_plain = cipher.decrypt(resume_text.encode("ascii")).decode("utf-8")
        except (InvalidToken, AttributeError):
            continue
        bind.execute(
            text("UPDATE analyses SET jd_text = :jd, resume_text = :resume WHERE id = :id"),
            {"id": row_id, "jd": jd_plain, "resume": resume_plain},
        )


def _encrypt_if_plaintext(cipher: Fernet, value: str | None) -> str | None:
    """Return ciphertext if `value` looks like plaintext; None if it's
    already Fernet ciphertext (so we don't double-encrypt) or null."""
    if value is None:
        return None
    try:
        cipher.decrypt(value.encode("ascii"))
    except (InvalidToken, ValueError):
        # Couldn't decrypt → it's plaintext. Encrypt it.
        return cipher.encrypt(value.encode("utf-8")).decode("ascii")
    # Decrypted cleanly → already encrypted, skip.
    return None
