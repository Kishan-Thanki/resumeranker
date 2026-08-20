#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")/.."

IMAGE_NAME="resumeranker-analysis-dashboard:latest"
CONTAINER_NAME="resumeranker-analysis-dashboard"
NETWORK_NAME="resumeranker-net"

if ! command -v docker >/dev/null 2>&1; then
    echo "ERROR: Docker is not installed."
    exit 1
fi

if ! docker info >/dev/null 2>&1; then
    echo "ERROR: Docker is installed but the Docker daemon is not running."
    exit 1
fi

if [[ ! -f ".env" ]]; then
    echo "ERROR: .env file not found."
    exit 1
fi

PORT="$(grep -E '^PORT=' .env | tail -n1 | cut -d= -f2- | tr -d '\r' || true)"

if [[ -z "${PORT}" ]]; then
    echo "ERROR: PORT not found in .env file."
    exit 1
fi

if [[ ! "${PORT}" =~ ^[0-9]+$ ]]; then
    echo "ERROR: Invalid PORT in .env: ${PORT}"
    exit 1
fi

if (( PORT < 1 || PORT > 65535 )); then
    echo "ERROR: PORT must be between 1 and 65535: ${PORT}"
    exit 1
fi

if ! docker network inspect "${NETWORK_NAME}" >/dev/null 2>&1; then
    echo "Creating Docker network: ${NETWORK_NAME}"
    docker network create "${NETWORK_NAME}" >/dev/null
fi

if docker ps -a --format '{{.Names}}' | grep -Fxq "${CONTAINER_NAME}"; then
    echo "Removing existing container: ${CONTAINER_NAME}"

    docker rm \
        --force \
        "${CONTAINER_NAME}" \
        >/dev/null
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
echo "Network   : ${NETWORK_NAME}"
echo "Port      : ${PORT}"
