# Security — Backend

Security posture for the Resume-Ranker backend (FastAPI + arq + Postgres + Redis):
what we defend against, how, and what we don't (yet).

The frontend's defenses (CSP, header policy, output escaping) live in
`../frontend/SECURITY.md`. The two stacks are independent — both docs
together make up the full posture.

If you find a vulnerability, see **Reporting** at the bottom.

---

## Threat model

| Asset | Worst-case impact |
| --- | --- |
| User accounts (email + sessions) | Account takeover → impersonation |
| Uploaded resumes | PII disclosure (we extract text and discard the file, but the text is persisted) |
| Job descriptions | Low — typically not sensitive |
| LLM API key (when configured) | Cost abuse, exfiltration of analyses |
| Postgres data at rest | Bulk PII disclosure if DB is exfiltrated |
| Backend RCE / SQLi | Full system compromise |

Adversaries we design against, in rough order of concern:

1. **Automated scanners / opportunistic attackers** — script kiddies probing public IPs for default creds, known CVEs.
2. **Account-takeover attempts** — credential stuffing, brute-force, magic-link replay.
3. **Insider misuse** — a logged-in user trying to access another user's data.
4. **Bot-driven abuse** — mass signups, mass analysis runs, rate-limit evasion.

Out of scope for v1:

- Targeted attacks by well-resourced adversaries.
- Side-channel / timing attacks.
- Physical attacks on the host.
- Supply-chain compromise of upstream packages (we run audits but trust them).

---

## Defenses in place

### Authentication

- **Passwordless magic-link** — 32 random bytes per token, SHA-256-hashed in DB, never stored or logged in raw form.
- **Single-use** — `consumed_at` timestamp; replay is rejected.
- **15-minute expiry** — short window for stolen-link risk.
- **Single error response** for not-found / expired / consumed — no enumeration.
- **Session tokens** — same 32-byte random, hashed, 30-day sliding expiry, revocable on sign-out.
- **Bearer-token in `Authorization` header** — no cookies, **CSRF is structurally impossible**.

### Rate limiting (Redis sliding-window)

- **Global per-IP cap**: 100 requests / minute across every endpoint. Catches abusive flooding that bypasses endpoint-specific limits (e.g. `/health` floods, `/auth/verify` brute force, `/me` token-oracle probing). Returns the generic `{"error": "rate_limited"}` shape so attackers can't infer the limiter type. Implemented in `app/middleware/global_rate_limit.py` against Redis.
- `POST /auth/request-link`: 5/hour per email, 20/hour per IP. Even when blocked, the response is still 200 `{"ok": true}` so attackers can't enumerate.
- `POST /analyses`: 10/day per user (`MAX_RESUMES_PER_USER_PER_DAY`).

### Input validation

- **Pydantic at every API boundary** — request bodies, query params, multipart fields. No `Any`. No free dicts.
- **PDF validation** — MIME type + `.pdf` extension + 10 MB cap. `pdfplumber` rejects non-PDFs gracefully.
- **JD text** — 50,000-char cap.
- **No `eval` / `exec` / shell interpolation** anywhere in the codebase.
- **SQLAlchemy parameterized queries** everywhere → SQL injection structurally impossible.

### LLM responses

- **`instructor` + Pydantic models** for every LLM call — the model's output must match a schema or the call retries / errors.
- **Stub mode by default** — no real LLM calls unless `LLM_API_KEY` is set to a real value. The active provider is determined by `LLM_MODEL` (e.g. `gemini/gemini-2.5-flash`, `anthropic/claude-haiku-4-5`).
- **Stub mode adds 1–2.5 s of randomized latency per call** so request timing can't be used to distinguish stub-vs-real mode (i.e. attackers can't tell that responses are canned).

### Containers

- **Non-root user (`app`)** in the backend Dockerfile.
- **Minimal base image** — `python:3.12-slim`.
- **OS security patches** applied at build time (`apt-get upgrade`).
- **Postgres credentials in `.env`** — not committed.
- **Internal-only Postgres + Redis** — host ports `15432` and `16379` are dev convenience; in production they're not exposed via the reverse proxy.

### CORS

- Single allow-list: `settings.app_base_url`. No wildcards.
- `allow_credentials=True` but credentials are bearer headers, not cookies — origin must match exactly.

### Observability

- `/health` endpoint reports DB + Redis status, no internal details.
- Logs do NOT include: tokens, password hashes, API keys, full session tokens, resume text contents.
- Logs DO include: timestamps, log levels, email addresses (per brief — emails can be logged).

### Dependency scanning

- **`pip-audit`** clean — 0 vulns across prod and dev deps.
- **Trivy on the backend image** — 0 fixable HIGH/CRITICAL.

### Anti-fingerprinting (don't tell adversaries which stack we run)

- **`Server: uvicorn` header stripped** via `uvicorn --no-server-header`. No backend framework signature in response headers.
- **`Date` header stripped** (`--no-date-header`) to remove one more well-known default-set header.
- **Generic error bodies.** All HTTP errors return `{"error": "<slug>"}` — never FastAPI's default `{"detail": ...}` and never Pydantic's validation-error array. Installed via `app/error_handlers.py`. Full error detail still goes to server logs.
- **Pydantic `RequestValidationError` wrapped** to return `{"error": "unprocessable"}` instead of the recognizable `{"detail":[{"loc":...,"msg":...,"type":...}]}` FastAPI default.
- **`/health` returns only `{"status": "ok"}` or `{"status": "degraded"}`** — internal service names (db, redis) never leak to anonymous callers. Per-service health stays in `docker compose ps` / logs for operators.

### Production deployment hardening

- `/docs`, `/redoc`, `/openapi.json` are only mounted when `ENVIRONMENT=development`. Production returns 404.
- Postgres and Redis host port publishing is for local-dev convenience only; behind a Coolify/Caddy reverse proxy in production they're internal-only.
- `APP_BASE_URL` must be set to the public HTTPS origin for magic-link emails to point at the right URL.
- **Hosting opacity (recommended at deploy time):** Cloudflare in front of the origin. DNS A-record proxied (orange cloud) → origin IP never resolves publicly. Cloudflare's ASN replaces the origin's. Lock the VM firewall to allow only Cloudflare's published IP ranges on :80/:443.

---

## What's still imperfect

| Gap | Severity | Why deferred |
| --- | --- | --- |
| Trivy flags some unfixable OS-level CVEs (no patch available upstream) | Various | Cannot patch until upstream issues fixes. Re-scan periodically. |
| No WAF in front | Medium | Add Cloudflare in front of the VM during deploy (free tier). |
| No audit log table (who signed in when, who created what analysis) | Low | Backend stdout is the de-facto audit log; ship to a log service in production. |
| No automated dep-update bot (Dependabot / Renovate) | Low | Manual `pip-audit` for now. Add Dependabot when repo lands on GitHub. |
| No fuzzing of file upload path | Low | `pdfplumber` is well-tested; size cap limits blast radius. |
| Per-user session limit unenforced | Low | A user can have unlimited concurrent sessions. Add max-N per user in v2. |

---

## Operational checklist before any deploy

1. Regenerate `POSTGRES_PASSWORD` to a long random value (`openssl rand -hex 32`) and update `DATABASE_URL`.
2. Set `ENVIRONMENT=production` (this hides `/docs` and `/openapi.json`).
3. Set `APP_BASE_URL` to the public HTTPS origin.
4. Set real `LLM_API_KEY` (provider per `LLM_MODEL`) and `RESEND_API_KEY` (or leave at `replace-me` to stay in stub mode).
5. Ensure no port mappings expose Postgres or Redis publicly — behind the proxy only.
6. Re-run `pip-audit` and Trivy in CI on every commit.

---

## Reporting

Found a vulnerability? Do **not** open a public issue. Email
`security@<your-domain>` (replace with real address before publishing the
repo) with:

- A description and reproduction steps.
- The branch / commit you tested against.
- Your expected vs. actual behavior.

We aim to acknowledge within 72 hours.
