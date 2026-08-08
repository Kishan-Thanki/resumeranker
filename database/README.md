# ResumeRanker - Database Layer

This directory contains the local, containerized infrastructure for the ResumeRanker application.

## Architecture

We utilize `docker-compose` to manage our local backend database, ensuring a consistent development environment that mirrors production (Neon).

- **PostgreSQL**: Uses the official `postgres:18-alpine` image. Data is persisted using a named Docker volume (`resumeranker-db-data`), ensuring data survives container restarts.

## Directory Structure

```text
database/
├── docker-compose.db.yml
└── run.sh
```

## Running the Infrastructure

There is no "build" step required, as we pull the official, production-grade image directly.

### 1. Launch the Services

Spin up the database container. If the volume does not exist, Docker will create it automatically.

```bash
./run.sh
```

### 2. Connection Details

The service is exposed on your host machine at the following port:

* **PostgreSQL**: `localhost:5432`

### 3. Stopping the Services

To stop the container while keeping the data volume intact:

```bash
docker compose -f docker-compose.db.yml down
```

*Note: Database schema migrations are handled by the API container (`/api`) upon startup using `golang-migrate`.*