# resume-ranker frontend — AI assistant context

SvelteKit 2 (Svelte 5 runes mode) on `@sveltejs/adapter-node`. Renders the
section-by-section match UI. This file loads automatically when working in
`frontend/`.

For full project context, see the repo-root `CONTEXT.md`. Per-stack rules and
detail in this folder's `AGENTS.md` (14 rules), `HANDOFF.md`, and
`SECURITY.md` — all three are required reading on first session.

@AGENTS.md
@HANDOFF.md

## Stack at a glance

SvelteKit 2 · Svelte 5 (runes) · `@sveltejs/adapter-node` · TypeScript ·
TailwindCSS 4 · shadcn-svelte · bits-ui · `@lucide/svelte` · svelte-sonner
(toasts) · mode-watcher (dark-mode FOUC guard) · vitest +
@testing-library/svelte. Built via Vite 8 + oxc-minify.

## Container & workflow

```
resume-ranker-frontend     adapter-node server, host :5173 → container :3000
```

**No `docker compose` for the frontend** — it runs via a `run.sh` wrapper so
Docker Desktop shows it flat (no project-group nesting). Intentional.

```bash
./run.sh           # build + start (idempotent)
./run.sh stop      # stop + remove container
./run.sh logs      # follow
./run.sh status    # live CPU / mem
```

For in-container commands (svelte-check, vitest, anything `pnpm` runs):

```bash
docker exec resume-ranker-frontend pnpm exec svelte-check --tsconfig ./tsconfig.json
docker exec resume-ranker-frontend pnpm exec vitest run
```

**Never run host-side `pnpm`/`node` for runtime work.** Host install of
`pnpm install --frozen-lockfile` for IDE intellisense is the only exception
(`node_modules/` is gitignored, doesn't affect Docker builds).

## Architecture rules

- **All data fetching goes through `src/lib/api.ts`.** No ad-hoc `fetch()` in
  components. Per `AGENTS.md`.
- **Svelte auto-escape only.** No `{@html ...}` anywhere — instant block in
  reviews.
- **CSP lives in `svelte.config.js`** (`kit.csp.mode = 'hash'`), NOT in
  `hooks.server.ts`. The two would collide and break hydration.
- **Source maps OFF in production.** Pinned in `vite.config.ts`
  (`build.sourcemap = false`). Don't suggest enabling.

## Top-3 frontend-specific gotchas

1. **mode-watcher inline script needs a hard-coded sha256 in CSP.** Located
   in `svelte.config.js` `script-src` directive. If you upgrade mode-watcher
   and CSP errors appear in the browser console, the new expected hash is in
   the error message — paste it in.
2. **`VITE_*` env vars are baked at build time.** Changing
   `VITE_API_BASE_URL` requires a rebuild (`./run.sh` re-runs the
   docker build).
3. **`svelte-check` is the source of truth, not `tsc`.** `tsc --noEmit` can't
   resolve `.svelte` re-exports. The IDE may show TS errors that
   `svelte-check` doesn't — trust svelte-check.

## Env vars

| Var | Read at | Default |
|---|---|---|
| `VITE_API_BASE_URL` | build (Vite inlines into bundle + CSP) | `http://localhost:8000` |
| `HOST` / `PORT` | runtime (adapter-node) | `0.0.0.0` / `3000` |


## Files I should always read on cold start

| Path | Purpose |
|---|---|
| `AGENTS.md` | The 14 rules. Style + architecture. |
| `HANDOFF.md` | Stack, deploy, deviations, env vars table. |
| `SECURITY.md` | CSP, headers, anti-fingerprint, deliberate trade-offs. |
| `svelte.config.js` | CSP directives + mode-watcher hash. |
| `src/hooks.server.ts` | Security headers (no CSP — that's in svelte.config.js). |
| `src/lib/api.ts` | The only `fetch()` surface. |
| `package.json` | Pinned deps, scripts. |
