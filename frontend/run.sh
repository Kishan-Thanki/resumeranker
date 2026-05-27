#!/usr/bin/env bash
# Build and run the frontend container without docker compose.
#
# We deliberately avoid compose so Docker Desktop shows the container at
# the top level (no project group nesting). The trade-off is no
# `compose ps/logs/down` — use the docker CLI equivalents below.
#
# Usage:
#   ./run.sh                  # build + start (re-creates if already running)
#   ./run.sh stop             # stop and remove the container
#   ./run.sh logs             # follow logs
#   ./run.sh status           # show resource usage
#
# Resource caps mirror what was in compose.yml. SvelteKit adapter-node
# idles at ~60 MB; cap is 256 MiB. Bump if you ever see OOM kills:
#   docker inspect resume-ranker-frontend | grep OOMKilled

set -euo pipefail

IMAGE="resume-ranker-frontend:latest"
NAME="resume-ranker-frontend"
HOST_PORT="5173"
CONTAINER_PORT="3000"

# Memory cap. The runtime needs ~60 MB idle, so 256 MiB is generous for
# serving traffic. svelte-check and vitest need ~1 GiB to run, so when
# you want to exec dev-time tools inside the container, start it with
# `FRONTEND_MEMORY=2g ./run.sh` (or run `docker update --memory 2g
# resume-ranker-frontend` on the live container before exec'ing).
MEMORY="${FRONTEND_MEMORY:-256m}"
MEMORY_SWAP="${FRONTEND_MEMORY_SWAP:-${MEMORY}}"

cd "$(dirname "$0")"

case "${1:-up}" in
  stop)
    docker rm -f "$NAME" 2>/dev/null || true
    echo "Stopped and removed $NAME"
    ;;
  logs)
    docker logs -f "$NAME"
    ;;
  status)
    docker stats --no-stream "$NAME"
    ;;
  up|"")
    # Remove any existing container so we get a clean start (mirrors
    # `docker compose up -d` behavior).
    docker rm -f "$NAME" 2>/dev/null || true

    echo "Building $IMAGE ..."
    docker build -t "$IMAGE" .

    echo "Starting $NAME ..."
    docker run -d \
      --name "$NAME" \
      --restart unless-stopped \
      -p "${HOST_PORT}:${CONTAINER_PORT}" \
      --memory "$MEMORY" \
      --memory-swap "$MEMORY_SWAP" \
      --memory-reservation 64m \
      --cpus 0.5 \
      --pids-limit 100 \
      "$IMAGE"

    echo ""
    echo "Up on http://localhost:${HOST_PORT}/"
    echo "Logs:   ./run.sh logs"
    echo "Stop:   ./run.sh stop"
    ;;
  *)
    echo "Usage: $0 [up|stop|logs|status]" >&2
    exit 2
    ;;
esac
