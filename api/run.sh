#!/bin/bash
set -e

cd "$(dirname "$0")"

docker network create resumeranker-net || true

echo "Running Database Migrations..."
docker run --rm \
  --network resumeranker-net \
  -v $(pwd)/migrations:/migrations \
  migrate/migrate \
  -path=/migrations/ \
  -database "postgres://postgres:postgres@resumeranker-db:5432/resumeranker?sslmode=disable" up

echo "Stopping existing API container..."
docker stop resumeranker-api || true
docker rm resumeranker-api || true

echo "Starting API Server container..."
docker run -d \
  --name resumeranker-api \
  --network resumeranker-net \
  --env-file env/dev.env \
  -e DATABASE_URL="postgres://postgres:postgres@resumeranker-db:5432/resumeranker?sslmode=disable" \
  -p 9080:8080 \
  resumeranker-api-image

echo "API container (resumeranker-api) started successfully on port 9080."
