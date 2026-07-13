<div align="center">
  <h1>ResumeRanker</h1>
  <p><b>AI-Powered Resume Parsing and Candidate Ranking Platform</b></p>
</div>

---

## Architecture Overview

ResumeRanker is built using a modern, polyglot microservice architecture. It strictly separates the web-facing logic from the heavy CPU/AI computation.

### The 4 Core Components:
- **`web/`**: The Frontend interface (Vanilla ES6/HTML/CSS) served blazingly fast via Caddy.
- **`api/`**: The Go (Golang) monolithic backend. It acts as the gateway, translating internal Python errors into semantic HTTP status codes, propagating `RequestID` distributed tracing, and handling auth/PostgreSQL.
- **`analysis/`**: The pure-Python AI Engine running on `uvloop`. It receives PDFs via gRPC, parses them securely via `Semaphore` thread bouncers, and orchestrates LLM requests.
- **`database/`**: The isolated PostgreSQL database container.

### Shared Contracts:
- **`proto/`**: The absolute source of truth. Contains the `analysis.proto` file that both the Go API (Client) and Python Engine (Server) use to generate their networking code, guaranteeing flawless HTTP/2 communication.

---

## How to Spin Up

We use Docker Compose to orchestrate the entire polyglot stack seamlessly.

### 1. Configure Environment
Before starting, ensure you have your environment variables configured.
*(Important: Create your `analysis/env/dev.env` with your LLM API Key first)*

### 2. Boot the Cluster
Run the following from the root of the project to build and launch all 4 microservices simultaneously:

```bash
docker compose up --build -d
```

To view the aggregated, distributed-traced logs across the entire system:
```bash
docker compose logs -f
```

---

## Network Mapping
- **Web UI**: `https://localhost:8443` (HTTP fallback on `8080`)
- **Go API**: `http://localhost:9080`
- **Python Engine (gRPC)**: `http://localhost:8001`
- **Postgres Database**: `5432`

---

## Developer Documentation

Each core component has its own deeply detailed `README.md` and `.agents/` documentation folder. 

If you are a developer (or an AI Agent) diving into a specific stack, **you must read the `README.md` inside that specific directory** before modifying code:
- Go API Rules: `api/.agents/AGENTS.md`
- Python AI Rules: `analysis/.agents/AGENTS.md`
