# resume-ranker — AI assistant context (repo root)

Section-by-section resume↔JD matching with evidence quotes. **No top-line
score.** This file loads automatically into every AI session and gives
you the project's "must-know" baseline. Stack-specific context lives in
`backend/CONTEXT.md` and `frontend/CONTEXT.md`, which auto-load when working in
those folders.

## Layout

```text
resume-ranker/
├── backend/                 FastAPI + arq + Postgres + Redis (own image)
├── frontend/                SvelteKit (adapter-node)         (own image)
├── .github/                 dormant CI (gitleaks, pip-audit, pnpm audit, Trivy)
└── .pre-commit-config.yaml  gitleaks secret scanner
```

The two stacks are **fully independent** — each has its own `Dockerfile`,
image, and lifecycle. Only the browser bridges them (CORS + bearer-token
fetch). There is no root `compose.yml`.

## The hard rules

1. **Docker-first.** Never `pnpm` / `uv` / `python` on the host. All commands
   go through `docker compose exec ...` (backend) or `docker exec
   resume-ranker-frontend ...` (frontend). Host install of `pnpm install` for
   IDE intellisense is the only exception.
2. **Backend never calls the LLM.** API is thin (validate + DB + enqueue);
   only the worker invokes litellm. Per `backend/AGENTS.md` rule #10.
3. **Frontend never renders raw HTML.** Svelte auto-escape only. No
   `{@html ...}` anywhere. Per `frontend/AGENTS.md`.
4. **Quality gates are mandatory.** Both stacks must pass `mypy --strict` (backend),
   `svelte-check` (frontend), and their respective test suites before any
   change ships.

## How the stacks run

```bash
# Backend (4 containers: api, worker, postgres, redis)
cd backend && docker compose up -d --build

# Frontend (1 container, no compose by design — flat in Docker Desktop)
cd frontend && ./run.sh
```

Backend exposes `http://localhost:8000`. Frontend serves `http://localhost:5173`.

## Top-5 gotchas (with one-line fixes)

1. **CSP / mode-watcher hash.** Frontend CSP is in `svelte.config.js`
   (`mode: 'hash'`). mode-watcher's inline script has a hard-coded sha256 in
   there — when you upgrade mode-watcher, the browser console reports the new
   expected hash; paste it into the directives.
2. **`VITE_*` env vars are baked at build time.** Changing
   `VITE_API_BASE_URL` requires `./run.sh` (rebuild).
3. **Short backend container names** (`api`, `worker`, `postgres`, `redis`)
   collide with anything else on the host using those names. Stop the other
   one first.
4. **Backend stub modes.** `LLM_API_KEY=replace-me` and
   `RESEND_API_KEY=replace-me` (default in `.env`) keep the LLM and email in
   stub mode. Magic-link URLs print to api logs:
   `docker logs api | grep "Link:"`. The active LLM provider is
   determined by `LLM_MODEL` (e.g. `gemini/gemini-2.5-flash`).
5. **VSCode Python venv auto-activation is globally disabled** (user
   preference). Don't re-enable it; see auto-memory entry `user_ide_setup`.

## Where to find more

| Topic | File |
| --- | --- |
| Backend rules (15) | `backend/AGENTS.md` |
| Backend handoff (stack, deploy, deviations) | `backend/HANDOFF.md` |
| Backend security posture | `backend/SECURITY.md` |
| Frontend rules (14) | `frontend/AGENTS.md` |
| Frontend handoff | `frontend/HANDOFF.md` |
| Frontend security posture | `frontend/SECURITY.md` |
| Cross-session accumulated knowledge | `~/.ai/projects/<this-project>/memory/MEMORY.md` |
