# ResumeRanker - Database & Cache Layer

This directory contains the local, containerized infrastructure for the ResumeRanker application.

## Architecture

We utilize `docker-compose` to manage our local backend dependencies, ensuring a consistent development environment.

- **PostgreSQL**: Uses the official `postgres:16-alpine` image. Data is persisted using a named Docker volume (`resumeranker-db-data`), ensuring data survives container restarts.
- **Redis**: Uses the `redis:alpine` image to handle caching and session management.
- **Networking**: Services are isolated within the `resumeranker-net` network but are exposed to your host machine.

## Directory Structure

```text
database/
├── docker-compose.db.yml
└── run.sh
```

## Running the Infrastructure

There is no "build" step required, as we pull the official, production-grade images directly.

### 1. Launch the Services

Spin up both the database and the cache container. If the volume or network does not exist, Docker will create them automatically.

```bash
./run.sh
```

### 2. Connection Details

The services are exposed on your host machine at the following ports:

- **PostgreSQL**: `localhost:5432`
- **Redis**: `localhost:6379`

### 3. Stopping the Services

To stop the containers while keeping the data volume and network intact:

```bash
docker-compose -f docker-compose.db.yml down
```

---

_Note: Database schema migrations are handled by the API container (`/api`) upon startup using `golang-migrate`._
