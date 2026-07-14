# ResumeRanker - Analysis Engine

This is the core Machine Learning / AI backend for **ResumeRanker**. It is a high-performance gRPC server that uses `litellm` and `instructor` to extract structured JSON data from Job Descriptions and Resumes.

## Architecture Highlights

- **gRPC Only**: We do not use REST, FastAPI, or HTTP. All communication with the Go backend happens strictly over HTTP/2 gRPC streams for maximum throughput.

- **Provider Agnostic**: The engine uses `litellm`, meaning we can hot-swap from Gemini to OpenAI to Anthropic just by changing the `.env` file. No code changes required.
- **Extreme Concurrency**:
  - The gRPC server runs on `uvloop`, a drop-in C-based replacement for `asyncio` that handles tens of thousands of requests per second.

  - PDF extraction uses the blazingly fast, C++ based `pypdfium2` (BSD licensed). Extraction is safely offloaded to background threads and throttled via `asyncio.Semaphore` (`pdf_bouncer`) to prevent OOM crashes.
  - LLM API calls are executed concurrently using `asyncio.gather` to cut latency by 50%.

- **Dynamic Domain Schemas**: The engine uses a Domain Strategy pattern. Pydantic schemas (like `SectionsAnalysis`) are dynamically generated at runtime via `pydantic.create_model` based on the specific industry (e.g., Tech vs Healthcare), allowing infinite scaling of scoring rubrics without touching core code.
- **Zero-Cost Testing**: The integration test suite uses `@patch` to dynamically intercept LLM network calls, meaning the CI/CD pipeline runs in ~1 second and costs $0 in LLM API tokens.
- **Datadog-Ready Tracing**: Every log emitted automatically intercepts and appends the Go-originated `request_id` via Python `contextvars` to seamlessly support distributed tracing in ELK or Datadog without manual logger passing.

## Directory Structure

- `app/main.py`: The lightweight gRPC server entry point.
- `app/servicer.py`: The core gRPC Servicer that handles orchestration logic.
- `app/domain/`: The Strategy Pattern implementations for different industries (e.g., `TechDomain`).
- `app/pb/`: Generated Protobuf Python bindings.
- `app/schemas/`: Modular Pydantic V2 schemas (`core`, `extraction`, `analysis`, `api`) used by `instructor`.
- `app/services/`: Core logic for LLM orchestration and PII-scrubbed PDF parsing.
- `tests/`: Pure Unit and Integration tests.
- `benchmarks/`: Tooling and scripts for running zero-cost load tests using `ghz`.
- `bin/`: Developer scripts to build and run the Docker container.
- `env/`: Environment variable templates (dev and prod).

## How to Run Locally

### 1. Set up your environment

Copy the template and add your Gemini (or OpenAI) API Key.

```bash
cp env/dev.env.template env/dev.env
# Edit env/dev.env to include your API Key
```

### 2. Run the Docker Container

```bash
bash bin/build.sh
bash bin/run.sh
```

The gRPC server will now be listening on `localhost:8001`.

## How to Test

```bash
uv run pytest tests/
```

## Future Roadmap

Please refer to [FUTURE_ARCHITECTURE.md](./FUTURE_ARCHITECTURE.md) for our roadmap
