"""Application-layer encryption for sensitive TEXT columns.

We use Fernet (AES-128-CBC + HMAC-SHA256, with versioned envelope) from
the `cryptography` package. Plaintext leaves SQLAlchemy on the way in,
ciphertext lands in Postgres; on the way out, ciphertext is decrypted
back to plaintext transparently.

Why app-layer and not pgcrypto:
- Keeps the key out of the database server. Even a full DB dump
  (volume snapshot, backup file, accidental psql `SELECT *`) yields
  nothing without the key, which lives in the API container's env.
- Works the same in dev (sqlite in tests), staging, and prod.
- No Postgres extension to install or version-pin.

Stub mode: when `settings.encryption_stub_mode` is true (no real key
configured), the type passes values through unchanged. This keeps
local-dev and CI flowing without forcing every contributor to generate
a key. Production deployments MUST set `RESUME_ENCRYPTION_KEY`.

Rotation: not supported in v1. Bumping the key invalidates every
existing row. When rotation is needed, write a one-shot migration that
constructs a `MultiFernet([new_key, old_key])`, reads → decrypts with
the old key → re-encrypts with the new key, then drops the old key.
"""

from __future__ import annotations

from functools import lru_cache
from typing import Any

from cryptography.fernet import Fernet, InvalidToken
from sqlalchemy import Text
from sqlalchemy.engine import Dialect
from sqlalchemy.types import TypeDecorator

from app.config import settings


@lru_cache(maxsize=1)
def _fernet() -> Fernet | None:
    """Lazily construct the Fernet instance.

    Cached so we don't re-parse the key on every column access. Returns
    None in stub mode so callers can short-circuit to plaintext.
    """
    if settings.encryption_stub_mode:
        return None
    return Fernet(settings.resume_encryption_key.encode("utf-8"))


class EncryptedText(TypeDecorator[str]):
    """SQLAlchemy column type that transparently encrypts/decrypts TEXT values.

    The underlying column is still TEXT (ciphertext is urlsafe-base64,
    so it's ASCII-safe). Length is roughly 1.4x plaintext + 100 bytes
    of Fernet envelope overhead — Text has no length limit so this is
    fine for resume / JD bodies up to a few hundred KB.

    Tolerates legacy plaintext rows: if decryption fails (InvalidToken),
    the value is assumed to be pre-encryption plaintext and returned
    as-is. This lets the encryption migration run lazily — rows get
    re-encrypted on the next UPDATE, and reads keep working in the
    meantime. After the backfill migration runs, this fallback becomes
    dead code but is cheap to keep for defense-in-depth.
    """

    impl = Text
    cache_ok = True

    def process_bind_param(self, value: str | None, dialect: Dialect) -> str | None:
        if value is None:
            return None
        cipher = _fernet()
        if cipher is None:
            return value
        return cipher.encrypt(value.encode("utf-8")).decode("ascii")

    def process_result_value(self, value: Any, dialect: Dialect) -> str | None:
        if value is None:
            return None
        cipher = _fernet()
        if cipher is None:
            # Stub mode — value is plaintext (or, if the DB was populated
            # before stub mode was toggled on, ciphertext we can't read).
            return str(value)
        try:
            return cipher.decrypt(str(value).encode("ascii")).decode("utf-8")
        except InvalidToken:
            # Legacy plaintext row — return as-is. The next write will
            # encrypt it.
            return str(value)
