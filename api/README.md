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

### Intelligent Analysis Engine

Submit a resume and a job description, and the `analysis` module orchestrates a secure connection to our external AI/LLM providers (the Engine). It normalizes the AI's output to generate a structured, highly contextual score and feedback report for downstream services to consume.

### Internal API Key & Quota Management

To prevent rogue internal services from blowing through our AI provider budgets, access to the engine is governed by the `apikey` service. Each internal consumer (e.g., frontend app, Slack bot, scheduled cron job) is issued an API key with a strict token quota.

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

Internal services use their issued API key to hit the engine:

```bash
POST /api/v1/analyze/resume
Header: Authorization: Bearer <INTERNAL_API_KEY>

{
  "resume_text": "Experienced software engineer with 10 years in Go...",
  "job_description": "Looking for a senior backend developer..."
}
```

**Response:**

```json
{
  "model": "engine-v1",
  "result": "{\"score\": 95, \"feedback\": \"Excellent match.\"}",
  "total_tokens": 150
}
```

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
go run ./cmd/api
```
