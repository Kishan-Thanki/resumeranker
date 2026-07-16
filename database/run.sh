#!/bin/bash
set -e

SCRIPT_DIR="$(dirname "$0")"

echo "Starting database and cache services..."

docker-compose -f "$SCRIPT_DIR/docker-compose.db.yml" up -d

echo "Database and Redis are running."