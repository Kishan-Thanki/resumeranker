#!/bin/sh
set -e

echo "Starting Caddy as user 'admin'..."

export HOME=/home/admin

exec su-exec admin caddy run --config /etc/caddy/Caddyfile --adapter caddyfile