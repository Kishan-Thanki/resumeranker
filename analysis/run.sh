#!/bin/bash
set -e

cd "$(dirname "$0")"

docker network create resumeranker-net || true

echo "Stopping existing Analysis container..."
docker stop resumeranker-analysis || true
docker rm resumeranker-analysis || true

echo "Starting Analysis Engine container..."

docker run -d \
  --name resumeranker-analysis \
  --network resumeranker-net \
  --env-file env/dev.env \
  -p 8001:8001 \
  resumeranker-analysis-image

echo "Analysis container (resumeranker-analysis) started successfully on port 8001."
