from collections.abc import AsyncIterator
from typing import Annotated

import redis.asyncio as redis_async
from fastapi import Cookie, Depends, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession

from app.config import settings
from app.db import get_db
from app.models.user import User
from app.security.cookies import SESSION_COOKIE_NAME
from app.services import auth_service

_redis_client: redis_async.Redis | None = None


def get_redis_client() -> redis_async.Redis:
    """Lazily create a singleton async redis client."""
    global _redis_client
    if _redis_client is None:
        _redis_client = redis_async.Redis.from_url(
            settings.redis_url,
            encoding="utf-8",
            decode_responses=True,
        )
    return _redis_client


async def get_redis() -> AsyncIterator[redis_async.Redis]:
    """FastAPI dependency yielding the shared async redis client."""
    yield get_redis_client()


async def get_session_token(
    session: Annotated[str | None, Cookie(alias=SESSION_COOKIE_NAME)] = None,
) -> str:
    """Pull the session token off the HttpOnly cookie set at /auth/verify.

    Returns the raw token (which is what `auth_service.get_session_user`
    expects — it hashes internally). Raises 401 when the cookie is
    missing, with the same generic error shape as an invalid one so
    callers can't distinguish "never signed in" from "session expired".
    """
    if not session:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="invalid or expired session",
        )
    return session


async def get_current_user(
    db: Annotated[AsyncSession, Depends(get_db)],
    token: Annotated[str, Depends(get_session_token)],
) -> User:
    user = await auth_service.get_session_user(db, token)
    if user is None:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="invalid or expired session",
        )
    return user


class RequireRole:
    def __init__(self, allowed_roles: list[str]) -> None:
        self.allowed_roles = allowed_roles

    def __call__(self, user: Annotated[User, Depends(get_current_user)]) -> User:
        if user.role.value not in self.allowed_roles:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="insufficient permissions",
            )
        return user
