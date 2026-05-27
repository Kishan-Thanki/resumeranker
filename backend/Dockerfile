# syntax=docker/dockerfile:1.7

FROM python:3.12-slim AS base
# Apply OS security patches on top of the base image. The published
# `python:3.12-slim` tag ships with already-known CVEs in OS packages
# (systemd, ncurses, etc.) that have fixes in apt but aren't yet rolled
# into the tag. This upgrade pulls them in.
RUN apt-get update \
    && apt-get -y upgrade \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
RUN pip install --no-cache-dir uv==0.5.*
WORKDIR /app

# --- deps stage: full deps (incl. dev) into /app/.venv ---------------------
# Dev deps (pytest, ruff, mypy) are kept in the final image so the team
# can run `docker compose exec api pytest` against the live container. The
# image is ~100 MB larger than a prod-only sync but the dev workflow stays
# intact in this single-image setup.
FROM base AS deps
COPY pyproject.toml uv.lock ./
RUN uv sync --frozen

# --- runtime stage: minimal final image ------------------------------------
FROM python:3.12-slim AS runner

# Apply OS security patches on the runtime base too (separate FROM, so
# the `base` stage's upgrade doesn't carry over).
RUN apt-get update \
    && apt-get -y upgrade \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Non-root user
RUN groupadd -r app && useradd -r -g app app

WORKDIR /app
# WORKDIR creates /app as root. Hand it to `app` so ruff/mypy/pytest can
# write their cache dirs when invoked via `docker compose exec api ...`.
RUN chown app:app /app

COPY --from=deps --chown=app:app /app/.venv ./.venv
COPY --chown=app:app . .

ENV PATH="/app/.venv/bin:$PATH" \
    PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1

USER app
EXPOSE 8000

# Default CMD runs the API. The worker container in compose.yml overrides
# this with `arq app.worker.WorkerSettings`. Same image, two entrypoints.
# `--no-server-header` + `--no-date-header` strip fingerprintable defaults.
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000", "--no-server-header", "--no-date-header"]
