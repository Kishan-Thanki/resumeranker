from httpx import AsyncClient


async def test_health_ok(client: AsyncClient) -> None:
    """Health returns minimal shape — no internal service names leaked."""
    resp = await client.get("/health")
    assert resp.status_code == 200
    body = resp.json()
    assert body == {"status": "ok"}
    # Explicitly verify no internal-detail keys leak to anonymous callers.
    assert "checks" not in body
    assert "db" not in body
    assert "redis" not in body
