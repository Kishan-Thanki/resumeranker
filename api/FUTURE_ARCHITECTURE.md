# Future Architecture & Scaling Roadmap (Go API)

This document outlines the architectural roadmap for the Go API. These features were intentionally deferred during the MVP to maintain velocity but are essential for scaling the system to thousands of concurrent users.

They are organized by priority (highest impact to lowest impact).

---

## [P0] Critical Performance & Cost Upgrades

### 1. API Caching (Redis)
**Current State:** Every `AnalyzeRequest` results in a live gRPC call to the Python Engine, which burns LLM tokens and costs money.
**The Problem:** If a candidate uploads the exact same resume to the exact same job description twice (or if a recruiter refreshes the page), we pay the LLM cost twice.
**Future Solution:** 
Implement a Redis caching layer in the Go `analysis` service. 
1. SHA-256 hash the `ResumePDF` and `JobDescriptionPDF` bytes.
2. Check Redis for this composite hash.
3. If it exists, return the cached `ResultJson` instantly (Latency: 5ms, Cost: $0).
4. If it misses, call the Python gRPC server, return the result, and cache it with a 30-day TTL.

### 2. Client-Side gRPC Load Balancing
**Current State:** The Go API holds a single, static connection to `localhost:8001`.
**The Problem:** When we scale up to 5 parallel Python Docker containers to handle heavy LLM traffic, Go needs to distribute the load across all 5 containers.
**Future Solution:** Configure the Go `grpc.NewClient()` to use DNS-based Round Robin load balancing. 
```go
grpc.NewClient(
    "dns:///python-engine-headless-service:8001",
    grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
)
```

---

## [P1] Architectural Scaling (Asynchronous Workflows)

### 3. Background Task Queues (Redis / RabbitMQ)
**Current State:** The Go HTTP handler holds the client's connection open for 15-30 seconds waiting for the Python gRPC response.
**The Problem:** Long-lived synchronous HTTP connections waste memory, risk reverse-proxy (Caddy) timeouts, and create fragile user experiences if the network drops.
**Future Solution:** 
1. The Go HTTP handler saves the PDFs to an S3 bucket (or Postgres) and drops a message into a Redis Queue.
2. It instantly returns a `202 Accepted` to the Frontend with a `job_id`.
3. A background Go Worker (`goroutine`) pulls from Redis, makes the gRPC call to Python, and saves the `ResultJson` to PostgreSQL.
4. The Frontend polls a `GET /api/v1/jobs/{job_id}` endpoint until it is "Complete".

---

## [P2] User Experience (Long-Term)

### 4. Streaming gRPC & WebSockets
**Current State:** We use Unary gRPC (Send one request, wait, get one massive response).
**The Problem:** The user stares at a loading spinner for 15 seconds with no feedback.
**Future Solution:** 
Convert the `.proto` contract to use Server Streaming (`returns (stream AnalyzeResponse)`). 
1. Python yields partial progress (e.g., "JD Extracted...", "Resume Extracted...", "Finalizing Score...").
2. Go consumes this gRPC stream.
3. Go forwards these chunks to the React frontend via Server-Sent Events (SSE) or WebSockets, creating a dynamic, ChatGPT-like loading experience.
