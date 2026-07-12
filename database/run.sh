#!/bin/bash
set -e

echo "Ensuring docker network exists..."
docker network create resumeranker-net || true

echo "Ensuring docker volume exists..."
docker volume create resumeranker-db-data || true

echo "Stopping and removing existing db container..."
docker stop resumeranker-db || true
docker rm resumeranker-db || true

echo "Starting database container..."
docker run -d \
  --name resumeranker-db \
  --network resumeranker-net \
  -v resumeranker-db-data:/var/lib/postgresql/data \
  -p 5432:5432 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=resumeranker \
  resumeranker-db-image

echo "Database (resumeranker-db) started successfully on port 5432."
