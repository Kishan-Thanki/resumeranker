<div align="center">
  <h1>ResumeRanker</h1>
  <p><b>AI-Powered Resume Parsing and Candidate Ranking Platform</b></p>
</div>

---

## Architecture Overview

ResumeRanker is built using a modern, polyglot microservice architecture. It strictly separates the web-facing logic from the heavy CPU/AI computation.

### The 4 Core Components:
- **`web/`**: The Frontend interface, served blazingly fast via Caddy.
- **`api/`**: The Go (Golang) monolithic backend. It acts as the gatekeeper, handling JWT authentication, PostgreSQL interactions, quotas, and routing.
- **`analysis/`**: The pure-Python AI Engine. It receives PDFs via gRPC, parses them on background threads, and orchestrates requests to external LLMs (Gemini/OpenAI) using strict Pydantic JSON schemas.
- **`database/`**: The isolated PostgreSQL database container.

### Shared Contracts:
- **`proto/`**: The absolute source of truth. Contains the `analysis.proto` file that both the Go API (Client) and Python Engine (Server) use to generate their networking code, guaranteeing flawless HTTP/2 communication.

---

## How to Spin Up

To manually spin up the full local development environment, you must build and run each Docker component separately. Run the following commands from the root of the project:

### 1. Database
```bash
./database/build.sh
./database/run.sh
```

### 2. Python AI Engine
*(Important: Create your `analysis/env/dev.env` with your LLM API Key first)*
```bash
./analysis/build.sh
./analysis/run.sh
```

### 3. Go API (Includes Migrations)
```bash
./api/build.sh
./api/run.sh
```

### 4. Web Frontend
```bash
./web/build.sh
./web/run.sh
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
