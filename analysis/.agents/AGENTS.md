# Agent Rules & Philosophies (Python Engine)

This file dictates how future AI Agents should interact with the `@analysis/` codebase.

## 1. Architectural Mandates
- **gRPC Only:** Do NOT suggest adding `FastAPI`, `Flask`, or `REST`. This is a pure gRPC microservice.
- **No Stubs/Fakes in Production:** Do not use stub endpoints. The Go backend expects real responses.
- **Provider Agnosticism:** Never hardcode `openai` or `gemini` SDKs. Always use `litellm` and `instructor` so the models can be hot-swapped via `.env`.
- **Pure Python Independence:** This engine is 100% self-sufficient. It does not care about the frontend or databases. It receives PDFs and returns JSON. Do not couple it to external states.

## 2. Dependency Management
- We use `uv` exclusively for dependency management.
- If you add a dependency, use `uv add <pkg>`.
- The `Dockerfile` MUST use `uv sync --frozen` to guarantee that the production build is perfectly deterministic and matches the `uv.lock` file.

## 3. Performance & Concurrency
- Never block the `asyncio` event loop.
- The server runs on `uvloop` for ultra-fast C-based asyncio operations.
- If you use a CPU-bound library (like `pdfplumber`), you MUST wrap it in `asyncio.to_thread` and throttle it with an `asyncio.Semaphore` to prevent OOM crashes.
- Always use `asyncio.gather` for independent network tasks to minimize latency.

## 4. Structured Logging & Distributed Tracing
- All logs must use the pre-configured `JSONFormatter` in `logger.py` to ensure Datadog compatibility.
- Never manually pass a logger object. Use `request_id_var.get()` or `logger.info(..., extra={...})`.
- The `request_id` contextvar binds the gRPC request ID to all async tasks in that scope.
- Do NOT log raw `jd_text` or `resume_text` to prevent GDPR violations.

## 5. Testing
- Do NOT burn LLM API tokens in the test suite. 
- You MUST use `unittest.mock.patch` to intercept `litellm` or `instructor` calls.
