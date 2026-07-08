# ResumeRanker - Unified Gateway & Web Container

This directory contains the entire frontend architecture for ResumeRanker.

## Architecture Overview

Instead of hosting the reverse proxy (Caddy) and the frontend files separately, we have merged them into an immutable, stateless container:

- **Caddy Edge Server:** Handles automatic SSL (Let's Encrypt), static file serving, and reverse proxy routing to the Go API.
- **Vanilla Frontend:** Pure ES6 Javascript, HTML, and CSS (no build steps, no bloat).

### Directory Structure

```text
web/
├── public/              # (Source Code) All HTML, CSS, and JS files live here
├── Caddyfile            # (Config) The routing and domain logic for Caddy
├── Dockerfile           # (Config) Builds the secure Alpine container
├── entrypoint.sh        # (Config) Drops root privileges and boots Caddy
└── build.sh             # (Helper) Script to instantly compile the image
```

## Advanced Security

This container is built with production-grade security:

1. **Root Lockdown:** The `root` user password is permanently locked (`passwd -l`).
2. **Dedicated User:** A non-root `admin` user is created.
3. **Privilege Dropping:** The container starts as root, but `entrypoint.sh` instantly drops privileges and runs Caddy entirely as the `admin` user via `su-exec`.
4. **Port Binding:** Caddy is granted the `cap_net_bind_service` Linux kernel capability, allowing it to bind to ports 80/443 without being root.

## Building and Running

### 1. Build the Image

You can instantly compile the image using the provided helper script:

```bash
./build.sh
```

### 2. Run in Production

Because the container is completely dynamic, you pass your environment variables into it at runtime. It requires no hardcoding.

```bash
docker run -d \
  -p 80:80 \
  -p 443:443 \
  -e DOMAIN="resumeranker.kishanthanki.dev" \
  -e API_UPSTREAM="api-server:8080" \
  resumeranker-web:latest
```

- **`DOMAIN`**: Caddy will listen for this domain and procure an SSL certificate automatically.
- **`API_UPSTREAM`**: Caddy will automatically route all `/api/*` traffic to this internal destination (e.g., your Go API Docker container).

_Note: The frontend Javascript relies entirely on Caddy's routing. It uses the relative path `/api/v1` and requires no custom injection._
