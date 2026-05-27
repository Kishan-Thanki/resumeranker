import sys
from collections.abc import AsyncIterator

from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine
from sqlalchemy.pool import NullPool

from app.config import settings

# Under pytest, use NullPool so each connection is opened on the current event
# loop and discarded on release. The default pool binds connections to the
# loop they were created on, which breaks pytest-asyncio's per-test loops.
_under_pytest = "pytest" in sys.modules
_engine_kwargs: dict[str, object] = {
    "echo": settings.environment == "development" and not _under_pytest,
    "pool_pre_ping": True,
}
if _under_pytest:
    _engine_kwargs["poolclass"] = NullPool

engine = create_async_engine(settings.database_url, **_engine_kwargs)

async_session_factory = async_sessionmaker(engine, expire_on_commit=False, class_=AsyncSession)


async def get_db() -> AsyncIterator[AsyncSession]:
    """FastAPI dependency yielding an async DB session."""
    async with async_session_factory() as session:
        yield session
