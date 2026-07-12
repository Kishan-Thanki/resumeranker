# ResumeRanker - Database Container

This directory contains the isolated PostgreSQL database environment for the ResumeRanker application.

## Architecture

We use a dedicated, standalone PostgreSQL container built from the official `postgres:17` image. 

- **Data Persistence**: Data is persisted using a named Docker volume (`resumeranker-db-data`), ensuring data survives container restarts and removals.
- **Security**: The base image is the standard Debian-based PostgreSQL image, which provides a highly secure, heavily patched foundation. We avoid passing credentials in the Dockerfile directly; instead, they are securely injected at runtime.
- **Migrations**: Note that **database schema migrations are NOT executed here.** Migrations are automatically run by the API container (`/api`) upon startup using `golang-migrate`.

## Directory Structure

```text
database/
├── Dockerfile           # (Config) The PostgreSQL 17 base image
├── build.sh             # (Helper) Script to build the database image
└── run.sh               # (Helper) Script to spin up the database container and volume
```

## Building and Running

### 1. Build the Image

Compile the local database Docker image:
```bash
./build.sh
```

### 2. Run the Container

Spin up the database container on your local machine. This will automatically create the `resumeranker-net` Docker network and `resumeranker-db-data` volume if they don't already exist.

```bash
./run.sh
```

The database will be exposed on your host machine at `localhost:5432` and internally to other containers at `resumeranker-db:5432`.
