# Future Architecture & Scaling Roadmap

This document serves as a living record of architectural brainstorms and edge cases we discussed during the initial MVP design of the Python Analysis Engine. These are features we intentionally deferred to avoid premature optimization, but have clear implementation paths when the application scales.

They are organized by priority (highest impact to lowest impact).

---

## [P1] Architectural Scaling (Post-Launch)

_These require structural changes to how Go and Python communicate._

### 1. Background Task Queueing (Celery / Redis)

**Current State:** We protect against API rate limits natively using `asyncio.Semaphore`. The Go API holds the gRPC connection open until Python finishes (30-60 seconds).
**The Problem:** Holding hundreds of synchronous HTTP/gRPC connections open for 60 seconds is an anti-pattern that leads to timeouts at massive scale.
**Future Solution:**

1. The Go API drops the PDF payloads into a message broker (Redis, RabbitMQ).
2. Go immediately returns a `202 Accepted` to the Frontend with a `job_id`.
3. Stateless Python workers pull jobs, process them, and fire a webhook back to Go upon completion.

### 2. Dynamic Domain Strategy Loading

**Current State:** Hardcoded for MVP: `domain: DomainStrategy = TechDomain()`.
**The Problem:** When we introduce `SalesDomain` or `HealthcareDomain`, the Engine needs to dynamically swap them.
**Future Solutions:**

- **Explicit (UI):** Add an "Industry" field to the gRPC `AnalyzeRequest` and use a Python Factory pattern.
- **Implicit (Router LLM):** Pass the first 500 characters of the JD to a cheap, ultra-fast model (`gemini-2.5-flash-8b`) to classify the industry automatically, preventing users from trolling the UI.

---

## [P2] Safety & Engine Migrations (Long-Term)

_These should only be addressed if specific edge-cases arise in production._

### 3. The `pdfplumber` Dilemma (CPU vs Accuracy)

**Current State:** We use `pdfplumber` running on background threads. It takes 1-3 seconds to parse a PDF.
**The Problem:** `pdfplumber` is notoriously slow because it maps pixel coordinates for tables.
**Future Solution:** If CPU load becomes too expensive, migrate to a pure-text C-based extractor.

- _Option A:_ `PyMuPDF` (10x faster, but uses a restrictive AGPL-3.0 commercial license).
- _Option B:_ `pypdf` or `pdfium2` (Slightly slower than PyMuPDF, but commercially safe BSD licenses).
