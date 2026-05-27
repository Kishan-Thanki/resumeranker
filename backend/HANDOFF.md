# Resume-Ranker — Backend Handoff

A web tool that lets a job-seeking candidate upload their resume and a job
description and receive a section-by-section match analysis with evidence
quotes. **No top-line score** — section breakdowns only.

This document is a snapshot for a fresh AI session. The frontend at
`../frontend/` defines the contract; this backend implements it. The two
stacks are **independent** — each has its own `Dockerfile`, `compose.yml`,
and image, and is started from its own directory. They communicate only
over HTTP from the user's browser, so they don't share a Docker network.

Security posture: see [`SECURITY.md`](SECURITY.md) in this directory.

---

## Status

| | Status |
| --- | --- |
| API + worker stack | **Live in dev compose.** All endpoints implemented and verified end-to-end. |
| Auth flow | Magic-link → session works (with stdout email fallback when `RESEND_API_KEY` is unset). |
| Analysis pipeline | `POST /analyses` → arq queue → worker → `completed` with populated `SectionScore[]`. |
| LLM | **Stub mode** active by default. Real calls activate when `LLM_API_KEY` is set; provider determined by `LLM_MODEL` (default `gemini/gemini-2.5-flash`). |
| Tests | 8 unit/integration tests + 25 eval pairs (eval suite is `@pytest.mark.eval`-gated). All green. |
| `mypy --strict` | **Clean** (54 source files, 0 issues). |
| `ruff` | **Clean.** |
| Production image | Builds at 554 MB. Verified to start. |

---

## Stack

Python 3.12 · FastAPI · SQLAlchemy 2 async + asyncpg · PostgreSQL 16
(`pgvector/pgvector:pg16`) · Redis 7 · arq · litellm · instructor ·
pdfplumber · Resend (stdout fallback) · Alembic · uv · `python:3.12-slim`.
`ruff` + `mypy --strict`. Everything in Docker.

## Repo placement

```text
resume-ranker/
├── frontend/                  (SvelteKit, complete)
└── backend/                   ← this
    ├── BACKEND_BRIEF.md       (full spec)
    ├── HANDOFF.md             (you are here)
    ├── SECURITY.md            (defenses + threat model for this stack)
    ├── AGENTS.md              (15 rules)
    ├── Dockerfile             (prod multi-stage)
    ├── Dockerfile.dev         (dev with hot-reload)
    ├── compose.yml            (api + worker + postgres + redis)
    ├── .env.example
    ├── pyproject.toml, uv.lock
    ├── alembic.ini, migrations/
    └── app/
        ├── main.py            (FastAPI app, CORS, router mounting)
        ├── config.py          (pydantic-settings; `llm_stub_mode`, `email_stub_mode`)
        ├── db.py              (async engine; NullPool under pytest)
        ├── deps.py            (get_db, get_redis, get_current_user, get_bearer_token)
        ├── worker.py          (arq WorkerSettings + run_analysis)
        ├── models/            (User, MagicLink, Session, Analysis)
        ├── schemas/           (auth.py, analysis.py — analysis.py mirrors frontend types.ts)
        ├── routes/            (health, auth, analyses)
        ├── services/
        │   ├── auth_service.py        (magic links, sessions, hashing)
        │   ├── email_service.py       (Resend + stdout fallback)
        │   ├── pdf_service.py         (pdfplumber extraction, <100 chars → reject)
        │   ├── analysis_create.py     (queued-row insertion)
        │   ├── analysis_service.py    (DB ↔ Pydantic, ownership scoping)
        │   ├── job_queue.py           (arq enqueue wrapper)
        │   ├── llm_service.py         (litellm + instructor; stub short-circuit)
        │   ├── llm_stub.py            (canned fixtures by input hash)
        │   ├── llm_cache.py           (Redis cache by SHA-256(input))
        │   ├── llm_cost.py            (token + cost accounting per analysis)
        │   └── rate_limit.py          (Redis sliding-window counter)
        ├── domain/
        │   ├── base.py        (Domain Protocol)
        │   └── tech.py        (TechDomain — only domain in v1; PROMPT_VERSION="tech-v1")
        └── tests/
            ├── conftest.py    (resets engine + redis singletons between tests)
            ├── test_health.py, test_auth.py  (default suite)
            └── evals/         (5 JDs + 5 resumes + 25 expected ranges; marked @pytest.mark.eval)
```

## Docker resource naming

The compose project name (`resume-ranker-backend`) is the group label in
Docker Desktop. Container names are intentionally **short** because the
group already disambiguates them — Docker Desktop shows the group at the
top and the four short containers inside it.

| Resource | Name |
| --- | --- |
| Compose project (group in Docker Desktop) | `resume-ranker-backend` |
| Container (api) | `api` |
| Container (worker) | `worker` |
| Container (postgres) | `postgres` |
| Container (redis) | `redis` |
| Network | `resume-ranker` |
| Volume (postgres) | `resume-ranker-postgres-data` |
| Volume (redis) | `resume-ranker-redis-data` |

**Heads-up — name collision risk:** short container names like `postgres`
and `redis` are common. If you already have a `postgres` or `redis`
container running on the host (from another project), `docker compose up`
will error with "container name already in use". Stop the other one first
or revert to prefixed names in `compose.yml`.
| Image (dev) | `resume-ranker-backend:dev` (~780 MB) |
| Image (prod) | `resume-ranker-backend:prod` (~550 MB) |

Host port mappings (changed from brief because 5432 is in use on this dev box):

- Postgres `localhost:15432` → container `5432`
- Redis `localhost:16379` → container `6379`
- API `localhost:8000` → container `8000`

Internal container-to-container traffic uses default 5432/6379 via the
shared `resume-ranker-backend` network.

## HTTP API (everything implemented and verified)

Auth is a **HttpOnly session cookie** set by `/auth/verify` (not a bearer
header). The browser sends it automatically on same-site requests; the
frontend uses `credentials: 'include'`.

| Method | Path | Auth | Notes |
| --- | --- | --- | --- |
| GET | `/health` | none | 200 with `db: ok / redis: ok`, 503 otherwise. |
| POST | `/auth/request-link` | none | Body `{email, acceptedPolicyVersion}`. Always 200 `{ok: true}`. Rate-limited 5/hr per email + 20/hr per IP. Email goes to stdout in stub mode. |
| POST | `/auth/verify` | none | Body `{token}`. Sets the session cookie via `Set-Cookie`; returns `{user}` (no token in the body). 400 with same generic message on any failure. |
| POST | `/auth/sign-out` | cookie | 204. Clears the session cookie. |
| GET | `/me` | cookie | Returns `{id, email, acceptedPolicyVersion}`. 401 if invalid/expired/revoked. |
| POST | `/me/accept-policy` | cookie | Body `{acceptedPolicyVersion}`. Records re-acceptance; returns the updated `/me` payload. |
| DELETE | `/me` | cookie | 204. Hard-deletes the user; cascades to sessions, magic_links, analyses; clears the cookie. |
| GET | `/analyses` | cookie | `AnalysisResult[]`, newest first. Camel-case JSON matches frontend types. |
| GET | `/analyses/{id}` | cookie | `AnalysisResult` or 404 (same response for missing-vs-not-owned). |
| POST | `/analyses` | cookie | Multipart: `resume` (PDF), `jd_input_type=pdf\|text`, `jd_pdf` or `jd_text`. Returns 201 with `queued` row. Per-user quota `MAX_RESUMES_PER_USER_PER_DAY=10` + service-wide `MAX_ANALYSES_PER_DAY_GLOBAL`. |
| DELETE | `/analyses/{id}` | cookie | 204. Owner-scoped hard delete; 404 for missing-or-not-owned (no existence leak). |
| POST | `/contact` | none | Body `{name, email, message, website}`. Always 200; honeypot + 3/hr per-IP limit. |

## LLM stub mode

`app/config.py` exposes `settings.llm_stub_mode`: true when
`LLM_API_KEY` is unset, empty, `"replace-me"`, or `"your-key-here"`.

When true, `app/services/llm_service.py` short-circuits each of the three
public functions and pulls a deterministic-by-input fixture from
`app/services/llm_stub.py`. Pipeline shape, JSON output, and frontend
rendering are identical to real-LLM mode.

To enable real LLM calls: edit `.env`, set `LLM_API_KEY=<your-key>` and
`LLM_MODEL` to the provider/model you want
(e.g. `gemini/gemini-2.5-flash`, `anthropic/claude-haiku-4-5`,
`openai/gpt-4o-mini`). No code change required — the key is passed
explicitly to litellm so any provider it supports works.

## Email stub mode

`app/config.py` exposes `settings.email_stub_mode`: same logic for
`RESEND_API_KEY`. When true, magic-link URLs print to the api container's
stdout — grep with `docker compose logs api | grep "Link:"`.

## Daily commands (from inside `backend/`)

```bash
# Start everything (api + worker + postgres + redis)
docker compose up -d --build

# Stop
docker compose down

# Hard reset including DB data
docker compose down -v

# Run a one-off
docker compose exec api uv run python -m app.scripts.something
docker compose exec api uv add some-package
docker compose exec api uv run pytest                  # default: excludes evals
docker compose exec api uv run pytest -m eval          # runs the LLM eval suite

# Migrations
docker compose exec api uv run alembic revision --autogenerate -m "describe change"
docker compose exec api uv run alembic upgrade head

# Production image
docker build -t resume-ranker-backend:prod .
docker run --rm -p 8000:8000 --env-file .env resume-ranker-backend:prod

# Regenerate eval fixtures (5 JDs + 5 resumes + 25 expected JSONs)
docker compose exec api uv run python -m app.tests.evals._seed
```

**Never run `uv` or `python` on the host.** Rule #13.

## Deviations from the brief (intentional)

| Deviation | Why |
| --- | --- |
| `Dockerfile.dev` runs `uv sync` at **build time** as the brief specified — this works because the venv lives in a named volume (`resume-ranker-backend-venv`) so the bind mount doesn't clobber it. | No frontend-style "runtime install" needed; uv's behavior plays nicer with Docker than pnpm's did. |
| Host ports remapped to **15432 / 16379** for postgres/redis | Default 5432 was in use on the dev box. Inside-container ports unchanged. |
| `app/db.py` uses `NullPool` when running under pytest | The default SQLAlchemy pool binds connections to whichever event loop first uses them; pytest-asyncio's per-test loops then crash with "Event loop is closed". |
| `app/tests/conftest.py` resets the singleton redis client between tests | Same reason as above. |
| Eval suite passes trivially in stub mode | The stub is deterministic-by-input, so all 25 pair scores land in the hand-tuned expected ranges. Real signal appears once a real LLM key is configured. |
| `reportlab` is a runtime dep | Needed by an in-container script that generates a test PDF for `POST /analyses` smoke tests. Could be moved to dev-deps if you're size-conscious. |

## Open items

1. **Real LLM key** — set `LLM_API_KEY` in `.env` (provider per `LLM_MODEL`, default `gemini/gemini-2.5-flash`) to flip out of stub mode. Then rerun `docker compose exec api pytest -m eval` to see the actual section-score quality.
2. **Real email** — set `RESEND_API_KEY` and verify a sending domain in Resend. The email body in `email_service.py` is intentionally minimal; production should add an HTML version.
3. **Frontend integration — DONE.** `frontend/src/lib/api.ts` calls this backend over CORS; the polling loop on `GET /analyses/{id}` runs every 2s while status is queued/processing. `VITE_API_BASE_URL` defaults to `http://localhost:8000`.
4. **`compose.prod.yml`** — overlay for production: worker container overrides CMD to `arq app.worker.WorkerSettings`. Not in v1.
5. **Object storage** — uploaded PDFs are extracted-and-discarded today; raw bytes never persisted. Fine for v1, may want S3 later.

## Rules in force

See `AGENTS.md` — 15 rules, copied verbatim from the brief.

The most load-bearing ones:

- API never calls litellm. Worker only. (#10)
- All LLM responses are typed via `instructor` + Pydantic. No free text. (#11)
- Prompt changes bump `PROMPT_VERSION` in `app/domain/tech.py`. (#12)
- Docker-first; no host-side `uv` or `python`. (#13)
