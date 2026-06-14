# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS base

WORKDIR /app

# --- deps stage: install all deps (including dev) for the build ---
FROM base AS deps
COPY package.json package-lock.json .npmrc* ./
# In case admin does not have .npmrc, the * makes it optional
RUN npm install

# --- build stage: produce the SvelteKit Node output ---
FROM base AS build
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

# --- runtime stage: single image, dev deps kept --------------------------
FROM node:22-alpine AS runner

RUN apk --no-cache upgrade
RUN rm -rf /usr/local/lib/node_modules/npm /usr/local/bin/npm /usr/local/bin/npx

WORKDIR /app



RUN addgroup -S app && adduser -S app -G app
RUN chown app:app /app
USER app

COPY --from=deps --chown=app:app /app/node_modules ./node_modules
COPY --from=build --chown=app:app /app/build ./build
COPY --from=build --chown=app:app /app/.svelte-kit ./.svelte-kit
COPY --chown=app:app . .

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000
EXPOSE 3000

CMD ["node", "build"]
