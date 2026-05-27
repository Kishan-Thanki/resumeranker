# resume-ranker backend — AI assistant context

FastAPI + arq + Postgres + Redis. Section-by-section resume↔JD matching API,
async via worker. This file loads automatically when working in `backend/`.

For full project context, see the repo-root `CONTEXT.md`. Per-stack rules and
detail in this folder's `AGENTS.md` (15 rules), `HANDOFF.md`, and
`SECURITY.md` — all three are required reading on first session.

@AGENTS.md
@HANDOFF.md

## Stack at a glance

Python 3.12 · FastAPI · SQLAlchemy 2 async + asyncpg · Postgres 16 (pgvector)
· Redis 7 · arq · litellm + instructor · pdfplumber · Resend · Alembic · uv
· `python:3.12-slim` · `mypy --strict` + `ruff`.

## Containers & names

```
api        FastAPI / uvicorn :8000
worker     arq queue consumer (same image as api, different CMD)
postgres   pgvector/pgvector:pg16, volume resume-ranker-postgres-data
redis      redis:7-alpine, volume resume-ranker-redis-data
```

- **Compose project name:** `resume-ranker-backend` (group label in Docker Desktop).
- Container names are **short** intentionally; they collide if other containers on the host use the same names.
- Network: `resume-ranker` (shared with frontend if it's also up; not required).

## Workflow

All commands run **inside the container** via `docker compose exec ...`. Never
host-side `uv` / `python`.

```bash
# Start (from backend/)
docker compose up -d --build

# Tests + lint
docker compose exec api pytest
docker compose exec api pytest -m eval        # LLM eval suite
docker compose exec api ruff check .
docker compose exec api ruff format --check .
docker compose exec api mypy app

# Migrations
docker compose exec api alembic upgrade head
docker compose exec api alembic revision --autogenerate -m "describe"

# Magic-link debugging (stub mode prints links to api stdout)
docker logs api | grep "Link:"
```

## Architecture rule

**The api never calls the LLM.** Only the worker invokes
`app/services/llm_service.py`. Routes are thin (validate → enqueue → DB read).
If a diff puts LLM calls into `app/routes/...`, that's a bug. Per `AGENTS.md`
rule #10.

## Mode flags

Two "stub modes" are the default for safe local dev. They're toggled by
env-var values in `.env`:

| Env var | `replace-me` → stub mode | Real value → live mode |
|---|---|---|
| `LLM_API_KEY` | Worker pulls canned fixtures from `app/services/llm_stub.py` | Real litellm calls to whichever provider `LLM_MODEL` points at (e.g. `gemini/gemini-2.5-flash`) |
| `RESEND_API_KEY` | Magic-link URLs print to `docker logs api` | Real email via Resend |

Stub mode adds 1–2.5 s artificial latency so timing can't distinguish stub
from real (anti-fingerprint). See `SECURITY.md`.

## Top-3 backend-specific gotchas

1. **`mypy --strict` is non-negotiable.** No bare `Any`. If you must, add
   `# type: ignore[<rule>]` with a comment explaining why.
2. **Async DB pool needs care under pytest.** `app/db.py` uses `NullPool` when
   running under pytest (otherwise pool-bound connections die between
   per-test event loops). Don't change this without reading the comment.
3. **Singleton Redis client also resets per test.** `app/tests/conftest.py`
   handles it. Same pattern.


## Files I should always read on cold start

| Path | Purpose |
|---|---|
| `AGENTS.md` | The 15 rules. Source of truth for style/architecture. |
| `HANDOFF.md` | Stack, deploy, deviations, open items. |
| `SECURITY.md` | Defenses + deliberate trade-offs (opaque tokens, no JWT). |
| `app/config.py` | Every env var the backend reads. |
| `app/main.py` | CORS allow-list, router mounting, exception handlers. |
| `pyproject.toml` | Dep versions, ruff/mypy/pytest config. |
