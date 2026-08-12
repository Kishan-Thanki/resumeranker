#!/usr/bin/env bash

set -euo pipefail

NETWORK_NAME="resumeranker-net"
NETWORK_MONITOR_NAME="resumeranker-network-monitor"

if ! docker info >/dev/null 2>&1; then
    echo "Docker is not installed or the Docker daemon is not running."
    exit 1
fi

if ! docker network inspect "${NETWORK_NAME}" >/dev/null 2>&1; then
    echo "Docker network does not exist: ${NETWORK_NAME}"
    echo "Start the Analysis Engine first with:"
    echo "  ./scripts/run-container.sh"
    exit 1
fi

if docker ps -a --format '{{.Names}}' | grep -Fxq "${NETWORK_MONITOR_NAME}"; then
    echo "Removing existing network monitor: ${NETWORK_MONITOR_NAME}"
    docker rm -f "${NETWORK_MONITOR_NAME}" >/dev/null 2>&1 || true
fi

echo "Starting Docker network monitor..."
echo
echo "Network : ${NETWORK_NAME}"
echo "Monitor : ${NETWORK_MONITOR_NAME}"
echo
echo "Capturing ALL traffic visible on ${NETWORK_NAME}."
echo "Press Ctrl+C to stop."
echo

docker run \
    --rm \
    --name "${NETWORK_MONITOR_NAME}" \
    --network "${NETWORK_NAME}" \
    --cap-add NET_RAW \
    --cap-add NET_ADMIN \
    nicolaka/netshoot \
    tcpdump \
        -nn \
        -tttt \
        -i any
