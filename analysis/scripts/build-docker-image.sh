#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")/.."

IMAGE_NAME="resumeranker-analysis"

if ! command -v docker >/dev/null 2>&1; then
    echo "ERROR: Docker is not installed."
    exit 1
fi

if ! docker info >/dev/null 2>&1; then
    echo "ERROR: Docker is installed but the Docker daemon is not running."
    exit 1
fi

echo "Building Docker image: ${IMAGE_NAME}"

docker build \
    --tag "${IMAGE_NAME}:latest" \
    .

echo
echo "Build completed successfully."
echo "Image: ${IMAGE_NAME}:latest"
