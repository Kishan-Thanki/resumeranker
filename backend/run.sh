#!/usr/bin/env bash
# Build and run the backend cluster using docker compose.
#
# This script exists to provide a uniform developer experience alongside
# the frontend/run.sh script.
#
# Usage:
#   ./run.sh                  # build + start in background
#   ./run.sh stop             # stop all backend containers
#   ./run.sh logs             # follow logs of all backend containers
#   ./run.sh logs api         # follow logs of just the API container

set -euo pipefail

cd "$(dirname "$0")"

case "${1:-up}" in
  stop)
    echo "Stopping backend cluster..."
    docker compose down
    ;;
  logs)
    if [ $# -eq 2 ]; then
      docker compose logs -f "$2"
    else
      docker compose logs -f
    fi
    ;;
  up|"")
    echo "Building and starting backend cluster..."
    docker compose up -d --build
    echo ""
    echo "Backend is up!"
    echo "Logs:   ./run.sh logs [service]"
    echo "Stop:   ./run.sh stop"
    ;;
  *)
    echo "Usage: $0 [up|stop|logs]" >&2
    exit 2
    ;;
esac
