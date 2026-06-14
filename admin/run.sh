#!/usr/bin/env bash
# Build and run the admin container without docker compose.

set -euo pipefail

IMAGE="resume-ranker-admin:latest"
NAME="resume-ranker-admin"
HOST_PORT="5174"
CONTAINER_PORT="3000"

MEMORY="${ADMIN_MEMORY:-256m}"
MEMORY_SWAP="${ADMIN_MEMORY_SWAP:-${MEMORY}}"

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
    docker rm -f "$NAME" 2>/dev/null || true

    echo "Building $IMAGE ..."
    docker build -t "$IMAGE" .

    echo "Starting $NAME ..."
    docker run -d \
      --name "$NAME" \
      --restart unless-stopped \
      --network resume-ranker \
      --env-file .env \
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
