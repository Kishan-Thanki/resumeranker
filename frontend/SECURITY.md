# Security — Frontend

Security posture for the Resume-Ranker frontend (SvelteKit + adapter-node):
what we defend against, how, and what we don't (yet).

The backend's defenses (auth, rate limiting, input validation, LLM
guardrails) live in `../backend/SECURITY.md`. The two stacks are
independent — both docs together make up the full posture.

If you find a vulnerability, see **Reporting** at the bottom.

---

## Threat model

| Asset | Worst-case impact |
|---|---|
| Session token in `localStorage` | Account takeover via XSS / phishing |
| Resume / JD text rendered in the UI | XSS if not auto-escaped |
| Frontend bundle | Source disclosure (file paths, internal logic) via leaked source maps |
| The Node server itself | Limited blast radius — it's stateless and only serves the SPA shell |

Adversaries we design against:

1. **Automated scanners** probing for default headers, exposed source maps, framework fingerprints.
2. **XSS attempts** — content reflected from API responses (JD/resume text quotes) is the primary surface.
3. **Clickjacking / framing** attempts.

Out of scope for v1:

- Targeted attacks by well-resourced adversaries.
- Side-channel / timing attacks.
- Compromised user device (we can't protect `localStorage` from a hostile extension).

---

## Defenses in place

### Output sanitization
- **Svelte auto-escape** for all user-controlled text. JD and resume quotes render as text, never HTML.
- **No `{@html ...}`** anywhere in the codebase.
- API responses are typed (`src/lib/types.ts`); fields render through `{value}` interpolation only.

### Auth surface
- **Bearer token in `Authorization` header**, stored in `localStorage` under `resume-ranker:session-token`.
- **No cookies** → **CSRF is structurally impossible**.
- Trade-off: `localStorage` is XSS-reachable. CSP (below) is the primary mitigation; Svelte auto-escape is the second layer.

### HTTP security headers
Set by `src/hooks.server.ts` on every response:

| Header | Value | Why |
|---|---|---|
| `Content-Security-Policy` | tight `default-src 'self'`; `connect-src` only allows the backend origin (read from `VITE_API_BASE_URL` at server startup) | Defense vs XSS, data exfiltration. The `connect-src` allowlist means even if an attacker injects a script, it can't beacon out. |
| `X-Content-Type-Options` | `nosniff` | Prevent MIME-type confusion |
| `X-Frame-Options` | `DENY` | Clickjacking protection |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Limit referrer leakage |
| `Permissions-Policy` | disables camera, mic, geolocation, payment, usb | Reduce attack surface |
| `Strict-Transport-Security` | 2 years + `includeSubDomains; preload` | Force HTTPS (effective once behind TLS proxy) |

### Containers
- **Non-root user (`app`)** in the Dockerfile.
- **Minimal base image** — `node:22-alpine`.
- **OS security patches** applied at build time (`apk upgrade`).
- **Bundled `npm` removed** from the image — we use pnpm via corepack; npm's bundled deps were the only HIGH CVE.

### Dependency scanning
- **`pnpm audit`** clean of moderate/high (1 known low remains — `cookie@0.6.0` via `@sveltejs/kit`; upstream constraint).
- **Trivy on the frontend image** — 0 fixable HIGH/CRITICAL.

### Anti-fingerprinting
- **`X-SvelteKit-Page`, `X-Powered-By`, `ETag` headers stripped** in `src/hooks.server.ts` on every response — adversaries can't read the framework signature from headers.
- **No source maps in production build.** `vite.config.ts` pins `build.sourcemap = false`; minified bundle has no `.map` siblings so JS sources aren't recoverable.

### Production deployment hardening
- This image is **stateless** — no DB, no sessions on the server, no file uploads to disk. It can be killed and restarted at any time.
- HTTPS termination should happen at the proxy / CDN (Cloudflare, Caddy, etc.) — HSTS is already set so the browser will refuse downgrades once a single HTTPS connect succeeds.
- **Hosting opacity (recommended at deploy time):** put Cloudflare in front of the origin. DNS A-record proxied (orange cloud) → origin IP never resolves publicly.

---

## What's still imperfect

| Gap | Severity | Why deferred |
|---|---|---|
| `cookie@0.6.0` transitive vuln from `@sveltejs/kit` | Low | Upstream pins `cookie ^0.6.0`. Patched on `Vary` cookie parsing. SvelteKit will bump eventually. |
| Trivy flags some unfixable OS-level CVEs in `node:22-alpine` (no patch upstream) | Various | Cannot patch until upstream issues fixes. Re-scan periodically. |
| No automated dep-update bot (Dependabot / Renovate) | Low | Manual `pnpm audit` for now. Add when repo lands on GitHub. |
| Session token in `localStorage` (XSS-reachable) | Low–Medium | Trade-off for "no cookies → no CSRF". CSP + Svelte auto-escape are the mitigations. A hostile extension can always read `localStorage`. |
| `VITE_API_BASE_URL` is inlined at build time | Low | Same image can't be repointed without rebuild. Move to runtime `window.__ENV__` if multi-env single-image is needed. |

---

## Operational checklist before any deploy

1. Set `VITE_API_BASE_URL` to the public HTTPS backend origin **before** running `pnpm build` — it gets baked into the bundle.
2. Verify `vite.config.ts` still has `build.sourcemap = false`.
3. Run `pnpm audit` and Trivy on the built image; fail the deploy on any new HIGH/CRITICAL.
4. Behind a TLS-terminating proxy (Caddy/Coolify/Cloudflare). HSTS is set unconditionally so plain-HTTP origins will lock browsers out after first visit.
5. Verify the CSP `connect-src` in production responses actually includes the prod backend origin (`curl -I https://<frontend>/` and read the header).

---

## Reporting

Found a vulnerability? Do **not** open a public issue. Email
`security@<your-domain>` (replace with real address before publishing the
repo) with:

- A description and reproduction steps.
- The branch / commit you tested against.
- Your expected vs. actual behavior.

We aim to acknowledge within 72 hours.
