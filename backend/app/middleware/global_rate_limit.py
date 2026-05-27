"""Per-IP global rate limit middleware.

Catches abusive traffic that bypasses the endpoint-specific limiters (e.g.
flooding `/health`, brute-forcing `/auth/verify`, oracle probing on `/me`).
The endpoint-specific limits in routes/auth.py and routes/analyses.py still
apply on top of this.

Implementation: Redis fixed-window counter keyed by client IP. The window
is 60 seconds. The 429 response uses the generic error shape installed by
app/error_handlers.py so it doesn't leak a framework signature.
"""

from __future__ import annotations

import logging
from collections.abc import Awaitable, Callable

from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import Response

from app.deps import get_redis_client

logger = logging.getLogger(__name__)


def _client_ip(request: Request) -> str:
    """Best-effort client IP extraction.

    When the app sits behind a trusted reverse proxy that sets
    `X-Forwarded-For`, we honor the leftmost value. Otherwise we fall back
    to the direct peer address.
    """
    fwd = request.headers.get("x-forwarded-for")
    if fwd:
        # Take the first hop — proxies append to the right.
        return fwd.split(",")[0].strip()
    return request.client.host if request.client else "unknown"


class GlobalRateLimitMiddleware(BaseHTTPMiddleware):
    def __init__(
        self,
        app: object,
        *,
        max_requests_per_minute: int = 100,
        window_seconds: int = 60,
    ) -> None:
        super().__init__(app)  # type: ignore[arg-type]
        self._max = max_requests_per_minute
        self._window = window_seconds

    async def dispatch(
        self,
        request: Request,
        call_next: Callable[[Request], Awaitable[Response]],
    ) -> Response:
        ip = _client_ip(request)
        key = f"rl:global:ip:{ip}"
        redis = get_redis_client()
        try:
            pipe = redis.pipeline()
            pipe.incr(key)
            pipe.ttl(key)
            count_res, ttl_res = await pipe.execute()
            count = int(count_res)
            if int(ttl_res) < 0:
                await redis.expire(key, self._window)
            if count > self._max:
                # Generic shape matches error_handlers.py.
                return JSONResponse(
                    status_code=429,
                    content={"error": "rate_limited"},
                )
        except Exception:
            # If Redis is briefly down, fail-open rather than 500ing every
            # request. The endpoint-specific limiters in auth/analyses
            # still provide defense in depth.
            logger.warning("global rate-limit redis check failed; allowing request")
        return await call_next(request)
