#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")/.."

IMAGE_NAME="resumeranker-analysis"
CONTAINER_NAME="resumeranker-analysis"
NETWORK_NAME="resumeranker-net"

if ! docker info >/dev/null 2>&1; then
    echo "Docker is not installed or the Docker daemon is not running."
    exit 1
fi

if [[ ! -f ".env" ]]; then
    echo ".env file not found."
    exit 1
fi

# Read PORT from .env so the published host port matches the
# port the application uses inside the container.
PORT="$(grep -E '^PORT=' .env | tail -n1 | cut -d= -f2- | tr -d '\r' || true)"

if [[ -z "${PORT}" ]]; then
    echo "PORT not found in .env file."
    exit 1
fi

if [[ ! "${PORT}" =~ ^[0-9]+$ ]]; then
    echo "Invalid PORT in .env: ${PORT}"
    exit 1
fi

if ! docker network inspect "${NETWORK_NAME}" >/dev/null 2>&1; then
    echo "Creating Docker network: ${NETWORK_NAME}"
    docker network create "${NETWORK_NAME}"
fi

if docker ps -a --format '{{.Names}}' | grep -Fxq "${CONTAINER_NAME}"; then
    echo "Removing existing container: ${CONTAINER_NAME}"
    docker stop "${CONTAINER_NAME}" >/dev/null 2>&1 || true
    docker rm "${CONTAINER_NAME}" >/dev/null 2>&1 || true
fi

echo "Starting Analysis Engine..."

docker run \
    --detach \
    --name "${CONTAINER_NAME}" \
    --network "${NETWORK_NAME}" \
    --env-file .env \
    --publish "${PORT}:${PORT}" \
    "${IMAGE_NAME}"

echo
echo "Analysis Engine started successfully."
echo "Container : ${CONTAINER_NAME}"
echo "Image     : ${IMAGE_NAME}"
echo "Port      : ${PORT}"
