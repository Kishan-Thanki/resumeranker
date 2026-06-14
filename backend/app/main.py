import logging

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from starlette.middleware.httpsredirect import HTTPSRedirectMiddleware

from app import error_handlers
from app.config import settings
from app.middleware.global_rate_limit import GlobalRateLimitMiddleware
from app.routes import analyses, auth, contact, health, iam

# Application logging setup.
#
# Without an explicit `basicConfig`, our `logging.getLogger(__name__)` calls
# fall through to the default root logger which has no handler — the lines
# get silently dropped. Configure the root once at import time so every
# module's logger emits to stdout/stderr (uvicorn captures it into
# `docker logs`).
#
# SQLAlchemy + the asyncpg pool are noisy at INFO level (one line per
# query / connection event). Pinning them to WARNING keeps our actual app
# logs readable.
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)-7s %(name)s: %(message)s",
)
logging.getLogger("sqlalchemy.engine").setLevel(logging.WARNING)
logging.getLogger("sqlalchemy.pool").setLevel(logging.WARNING)
logging.getLogger("asyncio").setLevel(logging.WARNING)

# In development we set `echo=True` on the SQLAlchemy engine so SQL is
# visible during local debugging. echo=True attaches SQLAlchemy's own
# StreamHandler to `sqlalchemy.engine.Engine` — without disabling
# propagation, every SQL line gets printed twice (once by echo, once by
# our root handler). Cutting propagation here lets us keep both behaviors
# without duplicates.
logging.getLogger("sqlalchemy.engine.Engine").propagate = False

# Hide the auto-generated OpenAPI surface from anonymous callers in any env
# that isn't `development`. The schema reveals every endpoint, every request
# body, every status code — fine for local dev, low-signal-high-noise risk
# in production.
_is_dev = settings.environment == "development"
_docs_url = "/docs" if _is_dev else None
_redoc_url = "/redoc" if _is_dev else None
_openapi_url = "/openapi.json" if _is_dev else None

app = FastAPI(
    title="Resume-Ranker API",
    version="0.1.0",
    description="Backend for Resume-Ranker. Matches frontend's AnalysisResult contract.",
    docs_url=_docs_url,
    redoc_url=_redoc_url,
    openapi_url=_openapi_url,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=[settings.app_base_url, "http://localhost:5174", "http://127.0.0.1:5174"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Global per-IP cap. Catches abusive flooding that the endpoint-specific
# limiters don't (e.g. /health, /me probing, /auth/verify brute force).
# 100 requests/minute is generous for a legitimate human session.
app.add_middleware(GlobalRateLimitMiddleware, max_requests_per_minute=100)

# Force HTTPS in production. Local dev runs over http://localhost so this
# is gated off. When deployed behind a TLS-terminating proxy (Cloudflare,
# Caddy, nginx), the proxy forwards `X-Forwarded-Proto: https` and
# Starlette respects that header — direct http:// hits get a 307 to the
# https:// equivalent. Defense-in-depth alongside the HSTS header set
# by the frontend's hooks.server.ts.
if settings.environment != "development":
    app.add_middleware(HTTPSRedirectMiddleware)

error_handlers.register(app)

app.include_router(health.router)
app.include_router(auth.router)
app.include_router(auth.me_router)
app.include_router(analyses.router)
app.include_router(contact.router)
app.include_router(iam.router)
