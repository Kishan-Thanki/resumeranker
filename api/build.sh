#!/bin/bash
set -e

cd "$(dirname "$0")"

echo "Building API image..."
docker build -t resumeranker-api-image .
echo "Build completed successfully."
