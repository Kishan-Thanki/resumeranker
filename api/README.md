<div align="center">
  <h1>ResumeRanker Core Service</h1>
  <p><b>Internal Engine for AI-Powered Resume Analysis</b></p>
</div>

---

## Overview

ResumeRanker Core is the internal backend service responsible for orchestrating AI analysis of candidate resumes against job descriptions. Designed exclusively for internal infrastructure, it acts as the centralized brain for our HR systems, ATS plugins, and internal tooling.

Built with performance, security, and scalability in mind, it safely abstracts external AI models away from our client-facing applications.

---

## System Capabilities

### Intelligent Analysis Orchestration

Submit a resume and a job description to the API, and the `analysis` module orchestrates a secure HTTP/2 gRPC stream to our isolated **Python AI Engine**. The Go backend acts as the gateway, enforcing quotas and authorization, while delegating the heavy CPU/LLM processing to the Python microservice. It natively injects a `RequestID` into the gRPC payload to enable **Distributed Tracing** across both languages, and translates raw gRPC Python errors into semantic HTTP status codes (e.g. 429 Rate Limit, 502 Bad Gateway).

### Internal API Key & Quota Management

To prevent rogue internal services from blowing through our AI provider budgets, access to the engine is governed by the `apikey` service. Each internal consumer (e.g., frontend app, Slack bot) is issued an API key with a strict token quota.

### Decoupled Security Model (RBAC)

We employ a dual-authentication strategy:

- **Service-to-Service:** Headless execution via API Keys.
- **Administrative Access:** JWT Bearer tokens for internal admins managing quotas, generating keys, and viewing audit logs.

### Comprehensive Audit Trails

For compliance and internal debugging, every critical event (admin login, key generation, and AI analysis success/failure) is asynchronously logged by the decoupled `audit` service, providing a 100% transparent history of system behavior.

---

## Internal API Quickstart

### 1. Issue a Service Key

_(Requires Admin JWT Authentication)_
Generate a key for a new internal service or consumer:

```bash
POST /api/v1/keys/generate
{
  "user_id": 1,
  "quota": 100000
}
```

### 2. Request Analysis

Internal services use their issued API key to hit the engine with multipart form-data (PDFs):

```bash
POST /api/v1/analyze/resume
Header: Authorization: Bearer <INTERNAL_API_KEY>
Content-Type: multipart/form-data

-F "resume_pdf=@/path/to/resume.pdf"
-F "job_description_pdf=@/path/to/jd.pdf"
```

**Response:**

```json
{
  "model": "gemini-2.5-flash",
  "result": { ... }, // Structured JSON Score Payload from Python
  "total_tokens": 450
}
```

**Semantic Error Responses:**
- `400 Bad Request`: Invalid/corrupt PDFs (`ErrInvalidPDF`).
- `429 Too Many Requests`: Upstream provider rate limits (`ErrRateLimit`).
- `502 Bad Gateway`: Python engine failed validation or threw an internal error.
- `504 Gateway Timeout`: Python or the LLM provider took too long to respond.

---

## Infrastructure & Setup

The service is built on a high-performance Go 1.21+ stack using Clean Architecture.

- **Routing:** `go-chi/chi/v5`
- **Database:** PostgreSQL via `pgxpool`
- **Migrations:** `golang-migrate`

### Running Locally

```bash
# Ensure local postgres is running, then initialize schema:
make install-tools
make migrate-up

# Start the core engine
make run
```

### Testing

The service is strictly tested with a 100% standard library coverage strategy.

```bash
# Run standard unit tests
make test

# Run tests including Postgres database integration tests
make test-integration
```

---

## License

This project is proprietary and confidential. All rights are reserved. 
Please see the [LICENSE](./LICENSE) file for more details.
