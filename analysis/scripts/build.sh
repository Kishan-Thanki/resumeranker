#!/bin/bash
set -e

cd "$(dirname "$0")"

echo "Building Analysis Engine Docker image..."
docker build -t resumeranker-analysis-image .
echo "Build complete."
