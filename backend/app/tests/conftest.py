"""Shared pytest fixtures. Provides an httpx AsyncClient pointed at the
in-process FastAPI app for fast in-container integration tests."""

import contextlib
from collections.abc import AsyncIterator

import pytest
from httpx import ASGITransport, AsyncClient

from app import deps
from app.db import engine
from app.main import app


@pytest.fixture(autouse=True)
async def _reset_singletons_between_tests() -> AsyncIterator[None]:
    """Module-level engine + redis singletons bind to whichever event loop
    first uses them. pytest-asyncio spins a fresh loop per test by default,
    so we tear both down between tests."""
    yield
    await engine.dispose()
    if deps._redis_client is not None:
        with contextlib.suppress(Exception):
            await deps._redis_client.aclose()
        deps._redis_client = None


@pytest.fixture
async def client() -> AsyncIterator[AsyncClient]:
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as c:
        yield c
