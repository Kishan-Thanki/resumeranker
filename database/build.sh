#!/bin/bash
set -e

cd "$(dirname "$0")"

echo "Building database image..."
docker build -t resumeranker-db-image .
echo "Build completed successfully."
