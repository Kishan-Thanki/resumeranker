# ResumeRanker

ResumeRanker compares a candidate resume with a job description and presents an evidence-backed fit report.

The project is organized as two independent Python applications connected through a shared gRPC contract:

```text
Browser
   |
   v
Streamlit Analysis Dashboard :8501
   |
   v
Analysis gRPC Service :50051
   |
   v
PDF extraction + LLM analysis
```

## Components

### Analysis Service

[`analysis/`](analysis/) is the backend analysis engine. It receives one resume PDF and one job-description PDF, extracts requirements and candidate claims, scores their alignment, and returns a structured JSON report.

### Analysis Dashboard

[`analysis_dashboard/`](analysis_dashboard/) is the Streamlit application where users upload documents and view the report through an Overview and Sections view.

### Shared Protocol

[`proto/analysis.proto`](proto/analysis.proto) defines the gRPC request and response used by the dashboard and analysis service.

## Quick Start With Docker

Requirements:

- Docker with Compose
- An LLM provider API key for real analysis

Create the local environment files:

```bash
cp analysis/.env.example analysis/.env
cp analysis_dashboard/.env.example analysis_dashboard/.env
```

Set `LLM_PROVIDER`, `LLM_MODEL`, and `LLM_API_KEY` in `analysis/.env`.

For Docker Compose, make sure `analysis_dashboard/.env` points to the service name:

```env
ANALYSIS_GRPC_HOST=analysis
ANALYSIS_GRPC_PORT=50051
ANALYSIS_GRPC_ADDRESS=analysis:50051
```

Start the application:

```bash
docker compose up -d --build
```

Open the dashboard:

```text
http://localhost:8501
```

Stop the application:

```bash
docker compose down
```

## Development

Each component has its own dependencies, lockfile, scripts, and tests.

Run analysis tests:

```bash
cd analysis
uv sync
uv run pytest -v -m "not e2e"
```

Run dashboard tests:

```bash
cd analysis_dashboard
uv sync
uv run pytest -v
```

The unit tests are local and do not require a real LLM request. The analysis E2E test requires a running service and external provider credentials.

## Project Layout

```text
analysis/             gRPC analysis backend
analysis_dashboard/   Streamlit user interface
proto/                shared protobuf contract
.github/workflows/    CI and release workflows
docker-compose.yml    local full-stack deployment
```
