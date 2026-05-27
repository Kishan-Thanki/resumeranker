# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS base
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /app

# --- deps stage: install all deps (including dev) for the build ---
FROM base AS deps
COPY package.json pnpm-lock.yaml .npmrc ./
RUN pnpm install --frozen-lockfile

# --- build stage: produce the SvelteKit Node output ---
FROM base AS build
# The backend origin is BAKED IN at build time — Vite inlines
# `import.meta.env.VITE_API_BASE_URL` into the JS bundle, and
# svelte.config.js reads it into the CSP `connect-src`. It cannot be
# changed at runtime. Defaults to localhost for local dev; production
# builds MUST pass the real API origin:
#   docker build --build-arg VITE_API_BASE_URL=https://api.example.com .
ARG VITE_API_BASE_URL=http://localhost:8000
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN pnpm build

# --- runtime stage: single image, dev deps kept --------------------------
# We reuse the `deps` node_modules (with vitest, svelte-check, etc.) so the
# team can run `docker compose exec web pnpm run test` and similar against
# the live container. The image is ~30 MB larger than a prod-only install
# but the test workflow stays intact in this single-image setup.
FROM node:22-alpine AS runner

# Apply OS security patches. `node:22-alpine` ships with already-known
# CVEs in Alpine packages that have fixes in the apk repo but aren't yet
# rolled into the tag. `apk upgrade` pulls them in.
RUN apk --no-cache upgrade

# Strip the bundled npm binary. We use pnpm via corepack; npm isn't called
# anywhere in this image. The bundled npm ships its own (older) copy of
# picomatch which Trivy flags HIGH (CVE-2026-33671).
RUN rm -rf /usr/local/lib/node_modules/npm /usr/local/bin/npm /usr/local/bin/npx

WORKDIR /app

# corepack so `pnpm` is available inside the running container for ad-hoc
# scripts (vitest, svelte-check, etc.).
RUN corepack enable && corepack prepare pnpm@latest --activate

# Non-root user for security
RUN addgroup -S app && adduser -S app -G app
# WORKDIR creates /app as root. Hand it to `app` so vitest/svelte-check can
# write their cache dirs when invoked via `docker compose exec web ...`.
RUN chown app:app /app
USER app

COPY --from=deps --chown=app:app /app/node_modules ./node_modules
COPY --from=build --chown=app:app /app/build ./build
# Carry the SvelteKit-generated TS config + types so svelte-check works
# against the running image (dev workflow inside the single image).
COPY --from=build --chown=app:app /app/.svelte-kit ./.svelte-kit
COPY --chown=app:app . .

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000
EXPOSE 3000

CMD ["node", "build"]
