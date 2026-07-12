# Resume Ranker

ResumeRanker is an AI-powered resume parsing and candidate ranking API.

## Project Structure
The project is split into three fully isolated Docker components:
- `database/`: PostgreSQL database image and setup.
- `api/`: Go backend API and database migrations.
- `web/`: Frontend interface served via Caddy.

## How to Spin Up

To manually spin up the full environment, you must build and run each component separately. Run the following commands from the root of the project:

### 1. Database
```bash
./database/build.sh
./database/run.sh
```

### 2. API (Includes Migrations)
```bash
./api/build.sh
./api/run.sh
```

### 3. Web
```bash
./web/build.sh
./web/run.sh
```

The Web UI will be available at `https://localhost:8443` (HTTP on `8080`) and the API serves traffic at `http://localhost:9080`. The Database is mapped to `5432`.
