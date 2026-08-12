#!/usr/bin/env bash

set -euo pipefail

CONTAINER_NAME="resumeranker-analysis"
NETWORK_MONITOR_NAME="resumeranker-network-monitor"

if ! docker info >/dev/null 2>&1; then
    echo "Docker is not installed or the Docker daemon is not running."
    exit 1
fi

if ! docker inspect "${CONTAINER_NAME}" >/dev/null 2>&1; then
    echo "Analysis Engine container does not exist: ${CONTAINER_NAME}"
    echo "Start the Analysis Engine first with:"
    echo "  ./scripts/run-container.sh"
    exit 1
fi

if ! docker ps --format '{{.Names}}' | grep -Fxq "${CONTAINER_NAME}"; then
    echo "Analysis Engine container is not running: ${CONTAINER_NAME}"
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
echo "Target  : ${CONTAINER_NAME}"
echo "Monitor : ${NETWORK_MONITOR_NAME}"
echo
echo "Capturing traffic from the Analysis Engine network namespace."
echo "Monitor is running in the background."
echo
echo "View traffic:"
echo "  docker logs -f ${NETWORK_MONITOR_NAME}"
echo
echo "Stop monitor:"
echo "  docker rm -f ${NETWORK_MONITOR_NAME}"
echo

docker run \
    --detach \
    --rm \
    --name "${NETWORK_MONITOR_NAME}" \
    --network "container:${CONTAINER_NAME}" \
    --cap-add NET_RAW \
    --cap-add NET_ADMIN \
    nicolaka/netshoot \
    tcpdump \
        -nn \
        -tttt \
        -i eth0
