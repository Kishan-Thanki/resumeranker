# Resume-Ranker — Frontend Handoff

A web tool that lets a job-seeking candidate upload their resume and a job
description and receive a section-by-section match analysis with evidence
quotes. **No top-line score** — only section breakdowns with side-by-side
evidence and factual gap statements.

This document is a snapshot for a fresh AI session. The backend at
`../backend/` implements the contract; this frontend renders it. The two
stacks are **independent** — each has its own `Dockerfile`, `compose.yml`,
and image, and is started from its own directory. They communicate only
over HTTP from the user's browser, so they don't share a Docker network.

Security posture: see [`SECURITY.md`](SECURITY.md) in this directory.

---

## Status

| | Status |
|---|---|
| Routes | **Done.** 6 routes + `+error.svelte`, wired to backend, polling, dark mode, mobile. |
| API client | `src/lib/api.ts` covers magic-link auth, sessions, analyses CRUD, file upload. |
| Tests | 40 vitest tests across 6 files. All green. |
| `svelte-check` | **Clean** (0 errors, 0 warnings). |
| Production bundle | Source maps OFF; minified by Vite's default `oxc-minify`. |

## Stack

SvelteKit 2 (Svelte 5 runes) · `@sveltejs/adapter-node` · TypeScript ·
TailwindCSS 4 · shadcn-svelte · bits-ui · `@lucide/svelte` · svelte-sonner
(toasts) · mode-watcher (dark mode) · vitest + @testing-library/svelte.
Everything runs in Docker via `node:22-alpine`.

## Repo placement

```
resume-ranker/
├── backend/                  (FastAPI + arq + Postgres + Redis)
└── frontend/                 ← this
    ├── BRIEF.md              (full UX spec)
    ├── HANDOFF.md            (you are here)
    ├── SECURITY.md           (frontend defenses)
    ├── AGENTS.md             (14 rules)
    ├── Dockerfile            (single multi-stage image)
    ├── run.sh                (build + run wrapper — see "Daily commands")
    ├── package.json, pnpm-lock.yaml
    ├── svelte.config.js      (adapter-node)
    ├── vite.config.ts        (build.sourcemap = false)
    ├── vitest.config.ts      (separate from vite.config — see comment in file)
    ├── tsconfig.json
    ├── tailwind.config.ts
    ├── components.json       (shadcn-svelte config)
    ├── static/               (favicon, etc.)
    └── src/
        ├── hooks.server.ts   (CSP + security headers + anti-fingerprint stripping)
        ├── app.css           (Tailwind imports, theme tokens)
        ├── routes/           (6 routes: /, /auth, /auth/sent, /auth/verify, /app, /app/new, /app/analysis/[id])
        ├── lib/
        │   ├── api.ts        (typed fetch client; reads VITE_API_BASE_URL)
        │   ├── types.ts      (AnalysisResult, SectionScore, RequirementMatch — mirror of backend)
        │   ├── components/   (TopBar, ScoreBar, SectionScoreCard, RequirementRow, EvidenceBlock, FileUpload, EmptyState, MatchBadge)
        │   ├── components/ui/ (shadcn-svelte primitives: button, card, input, label, textarea, dropdown-menu, progress, separator, sonner, tooltip)
        │   ├── stores/       (auth, analyses — Svelte stores that wrap api.ts)
        │   └── assets/       (favicon.svg)
        └── test-stubs/       (vitest setup for jsdom)
```

## Docker resource naming

This stack does **not** use `docker compose` — by design. Compose adds a
project-group label that makes Docker Desktop nest the container under a
folder, even when the project has only one container. Bypassing compose
keeps the container flat in the UI.

| Resource | Name |
|---|---|
| Container | `resume-ranker-frontend` |
| Image | `resume-ranker-frontend:latest` |
| Network | Docker default `bridge` (no custom network needed — the browser, not the container, talks to the backend) |
| Host port → Container | `5173 → 3000` |

The container internally serves on `3000` (adapter-node default with
`HOST=0.0.0.0`, `PORT=3000`). The host port stays `5173` so the dev URL
in the browser is unchanged.

Resource caps applied at `docker run` time (mirrors what backend's compose
file does for its services): 256 MiB memory cap, 64 MiB reservation, 0.5
CPU cores, 100 pids limit. SvelteKit idles at ~20 MiB — caps are
deliberately generous-with-headroom.

## Env vars

The frontend is configured by **two** build-time variables. Everything else is built into the image.

| Var | Read where | Read when | Default | Required in prod? |
|---|---|---|---|---|
| `VITE_API_BASE_URL` | `src/lib/api.ts` (inlined into the JS bundle) | Build time (`pnpm build`) | `http://localhost:8000` | **Yes** |
| `VITE_API_BASE_URL` | `svelte.config.js` (written into CSP `connect-src`) | Build time | `http://localhost:8000` | **Yes** |
| `HOST` | adapter-node | Runtime | `0.0.0.0` (set in Dockerfile) | — |
| `PORT` | adapter-node | Runtime | `3000` (set in Dockerfile) | — |

**Important quirk:** all `VITE_*` vars are inlined into the browser bundle at *build* time, so the same image can't be re-pointed at a different API without rebuilding. If you want one image for multiple environments, move the values into a runtime-injected `window.__ENV__` (small refactor).

## Daily commands (from inside `frontend/`)

```bash
# Build + start (re-creates if already running)
./run.sh

# Stop and remove
./run.sh stop

# Follow logs
./run.sh logs

# Live resource usage (CPU, memory)
./run.sh status

# Type-check + diagnostics (the source-of-truth check)
docker exec resume-ranker-frontend pnpm exec svelte-check --tsconfig ./tsconfig.json

# Unit tests
docker exec resume-ranker-frontend pnpm exec vitest run

# Production deploy — push the image to a registry and pull it on the target
docker build -t resume-ranker-frontend:latest .
docker tag resume-ranker-frontend:latest <registry>/<you>/resume-ranker-frontend:1.0
docker push <registry>/<you>/resume-ranker-frontend:1.0
# Then on the target host:
docker run -d --name resume-ranker-frontend --restart unless-stopped \
  -p 80:3000 -e VITE_API_BASE_URL=https://api.example.com \
  <registry>/<you>/resume-ranker-frontend:1.0
```

**Never run `pnpm`/`node` on the host.** Rule per `AGENTS.md`. (Exception:
host-side `pnpm install` is acceptable purely for IDE intellisense —
`node_modules/` is git/dockerignored, so it doesn't affect the build.)

## How the frontend talks to the backend

**Browser-side `fetch()` only.** The SvelteKit Node server in this image
serves the SPA shell; the browser then makes cross-origin calls to
`http://localhost:8000` (or whatever `VITE_API_BASE_URL` was at build time).

- **CORS:** the backend allow-list must include this frontend's origin
  (`settings.app_base_url` in `backend/app/config.py`).
- **Auth:** session token is in `localStorage` (`resume-ranker:session-token`) and sent as `Authorization: Bearer <token>` on every authed call. No cookies → CSRF is structurally impossible.
- **Polling:** `src/routes/app/analysis/[id]/+page.svelte` polls `GET /analyses/{id}` every **2 s** while status is `queued` or `processing`; gives up after **2 minutes** with a "still working" hint.

## E2E walkthrough (manual, real browser)

Assumes the backend is up on `http://localhost:8000` and this frontend on `http://localhost:5173`.

1. Open <http://localhost:5173/>
2. Click **Get started** → enter any email → submit.
3. In a separate terminal, in `../backend/`: `docker compose logs api | grep "Link:"` → grab the magic-link URL.
4. Paste the URL into the browser → lands on `/app`.
5. **New analysis** → paste a JD + attach a resume PDF → submit.
6. Detail page polls `queued → processing → completed` over ~5 s.
7. Sign out from the top-right menu.

## Deviations from the brief (intentional)

| Item | Brief | Reality | Why |
|---|---|---|---|
| Dev / prod image split | Brief had separate `Dockerfile.dev` + `compose.prod.yml` | **Single image, no overlays** | One version, production-parity locally. Rebuild for changes (~30 s). |
| Tailwind base color | `slate` or `zinc` | `zinc` | shadcn-svelte CLI dropped `slate`. |
| Icon package | `lucide-svelte` | `@lucide/svelte` | Renamed scoped package; same icon set. |
| Type-check command | `tsc --noEmit` | `svelte-check` | `tsc` can't resolve `.svelte` re-exports. |
| Single image carries dev deps | Brief: prod-only | **Includes vitest, svelte-check, etc.** | Lets `compose exec web ...` run tests against the live container. Image ~30 MB larger. |
| Vitest config | Same file as Vite | **Separate `vitest.config.ts`** | SvelteKit plugin forces SSR-condition resolution for `svelte`, which breaks `@testing-library/svelte`'s `mount()`. See comment in `vite.config.ts`. |

## Open items

1. **Backend integration — DONE.** `src/lib/api.ts` talks to the FastAPI backend over CORS; magic-link, sessions, analyses CRUD, polling all wired.
2. **Multi-env image** — if you ever need one image for staging vs prod, move `VITE_API_BASE_URL` from a build-time inline into a runtime `window.__ENV__` injection in `src/app.html`.
3. **ESLint / Prettier** — not configured. `svelte-check` is the source of truth for type + Svelte diagnostics. Add ESLint if you want stylistic rules.

## Rules in force

See `AGENTS.md` — 14 rules. Most load-bearing:
- Svelte auto-escape only. **No `{@html ...}`** anywhere.
- All data fetching goes through `src/lib/api.ts`. No ad-hoc `fetch()` in components.
- Docker-first; no host-side `pnpm` / `node` for runtime workflows.
- Source maps OFF in production builds (pinned in `vite.config.ts`).
