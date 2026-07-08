#!/bin/bash

set -e

IMAGE_NAME="resumeranker-web"
TAG="latest"

echo "====================================="
echo "Building Docker Image: $IMAGE_NAME:$TAG"
echo "====================================="

# Build the docker image
docker build -t ${IMAGE_NAME}:${TAG} .

echo ""
echo "Build completed successfully!"
echo ""
echo "To test this container, run:"
echo "docker run --rm -it \\"
echo "  --network resumeranker-net \\"
echo "  -p 8080:80 -p 8443:443 \\"
echo "  -e DOMAIN=\"localhost:8080\" \\"
echo "  -e API_UPSTREAM=\"rr-api:8080\" \\"
echo "  ${IMAGE_NAME}:${TAG}"
echo ""
