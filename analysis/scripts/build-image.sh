#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")/.."

IMAGE_NAME="resumeranker-analysis"

if ! docker info >/dev/null 2>&1; then
    echo "Docker is not installed or the Docker daemon is not running."
    exit 1
fi

echo "Building Docker image: ${IMAGE_NAME}..."

docker build -t "${IMAGE_NAME}" .

echo
echo "Build completed successfully."
echo "Image: ${IMAGE_NAME}"
