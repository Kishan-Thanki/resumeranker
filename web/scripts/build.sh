#!/bin/bash

set -e

cd "$(dirname "$0")/.."

IMAGE_NAME="resumeranker-web"
TAG="latest"

echo "====================================="
echo "Building Docker Image: $IMAGE_NAME:$TAG"
echo "====================================="

docker build -t ${IMAGE_NAME}:${TAG} .

echo ""
echo "Build completed successfully!"
echo ""
echo "To test this container, run:"
echo "docker run --rm -it \\"
echo "  --network resumeranker-net \\"
echo "  -p 8080:80 -p 8443:443 \\"
echo "  -e PUBLIC_DOMAIN=\"localhost\" \\"
echo "  -e API_DOMAIN=\"api.localhost\" \\"
echo "  -e API_UPSTREAM=\"resumeranker-api:8080\" \\"
echo "  -e WEB_ROOT=\"/var/www/html\" \\"
echo "  ${IMAGE_NAME}:${TAG}"
echo ""