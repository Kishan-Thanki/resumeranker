#!/bin/bash
set -e

docker network create resumeranker-net || true

echo "Stopping existing web container..."
docker stop resumeranker-web || true
docker rm resumeranker-web || true

echo "Starting Web Frontend container..."
docker run -d \
  --name resumeranker-web \
  --network resumeranker-net \
  --restart unless-stopped \
  -p 8080:80 \
  -p 8443:443 \
  -e PUBLIC_DOMAIN="http://localhost" \
  -e API_DOMAIN="http://api.localhost" \
  -e API_UPSTREAM="resumeranker-api:8080" \
  -e WEB_ROOT="/var/www/html" \
  resumeranker-web:latest

echo "Web container (resumeranker-web) started successfully on port 8080 (HTTP) and 8443 (HTTPS)."
