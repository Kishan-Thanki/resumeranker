# ResumeRanker - Python Analysis Engine

This is the core Machine Learning / AI backend for **ResumeRanker**. It is a pure Python, high-performance gRPC server that uses `litellm` and `instructor` to extract structured JSON data from Job Descriptions and Resumes.

## Architecture Highlights
- **gRPC Only**: We do not use REST, FastAPI, or HTTP. All communication with the Go backend happens strictly over HTTP/2 gRPC streams for maximum throughput.
- **Provider Agnostic**: The engine uses `litellm`, meaning we can hot-swap from Gemini to OpenAI to Anthropic just by changing the `.env` file. No code changes required.
- **Extreme Concurrency**: 
  - `pdfplumber` CPU bounds are offloaded to background threads.
  - LLM API calls are executed concurrently using `asyncio.gather` to cut latency by 50%.
- **Zero-Cost Testing**: The integration test suite uses `@patch` to dynamically intercept LLM network calls, meaning the CI/CD pipeline runs in ~1 second and costs $0 in LLM API tokens.

## Directory Structure
- `app/main.py`: The gRPC server entry point.
- `app/pb/`: Generated Protobuf Python bindings.
- `app/schemas/`: Pydantic V2 schemas used by `instructor` to enforce JSON outputs.
- `app/services/`: Core logic for LLM orchestration and PDF parsing.
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
Please refer to `FUTURE_ARCHITECTURE.md` for our roadmap on implementing `uvloop`, GZIP compression, and Celery task queues.
