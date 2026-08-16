# ResumeRanker Analysis Service

The Analysis Service is the AI-powered analysis engine of ResumeRanker. It receives a job description and a candidate resume as PDFs, extracts structured information from both, evaluates the candidate's fit against the job requirements, and returns a structured analysis.

It is implemented as an asynchronous gRPC service and is designed to remain independent of any specific LLM provider or model.

## What It Does

The service performs the following high-level pipeline:

```text
Job Description PDF ─┐
                     ├─→ PDF Text Extraction
Resume PDF ──────────┘
                            ↓
                     PII Scrubbing
                            ↓
              ┌─────────────┴─────────────┐
              ↓                           ↓
       JD Requirement                Resume Claim
         Extraction                   Extraction
              └─────────────┬─────────────┘
                            ↓
                       ID Assignment
                            ↓
                      Claims Scoring
                            ↓
                     Result Assembly
                            ↓
                    Structured Analysis
```

The service does not directly determine the final result from raw PDF text. Instead, it first converts the documents into structured requirements and claims, then uses those stable structures for scoring and final assembly.

## Architecture

The service is divided into several main areas:

```text
app/
├── domain/
│   └── tech.py
├── schemas/
├── services/
│   ├── pdf/
│   └── llm/
├── pb/
├── servicer.py
├── logger.py
└── main.py
```

### Domain

The domain layer defines the analysis strategy and the rules used for the current technical/software-engineering domain.

It provides:

* extraction prompts
* scoring prompts
* section taxonomy
* section weights
* scoring and final-result schemas

The rest of the service is written against domain abstractions rather than hard-coding the analysis pipeline around one particular domain.

### PDF Processing

The PDF service converts uploaded PDF bytes into text using PDFium.

Before the extracted text reaches the LLM stages, common personally identifiable information is scrubbed from the text.

The resulting text is therefore the source text used by the LLM analysis pipeline.

### LLM Pipeline

The LLM layer is provider- and model-agnostic.

The current implementation uses LiteLLM as the provider abstraction and Instructor for structured model output.

The analysis consists of three LLM stages:

1. **JD Extraction**
   Extracts structured job requirements.

2. **Resume Extraction**
   Extracts structured candidate claims.

3. **Claims Scoring**
   Evaluates how the candidate claims match the extracted requirements and produces section-level evaluations.

Each extracted requirement and claim receives a stable identifier so that LLM-produced verdicts can be correlated reliably during final assembly.

### Result Assembly

The LLM produces verdicts, but the service does not trust response ordering to associate those verdicts with source data.

The assembly layer validates that requirement verdicts:

* contain no duplicates
* contain no unknown requirement IDs
* cover the extracted requirements exactly

It then combines the validated verdicts with the original structured requirement and claim data to create the domain-specific final result.

## Provider and Model Independence

The service is intentionally not built around Gemini, OpenAI, Anthropic, or any single model.

The configured provider and model are resolved at runtime, while LiteLLM provides the provider abstraction underneath the analysis pipeline.

Usage information is normalized into a provider-neutral representation such as:

```text
input_tokens
output_tokens
total_tokens
reasoning_tokens
cached_input_tokens
cache_creation_input_tokens
cache_read_input_tokens
```

Provider-specific billing information is treated as telemetry rather than part of the core analysis contract.

## Reliability and Execution Controls

The service includes several safeguards around LLM execution:

* bounded concurrent LLM requests
* rolling request-rate limiting
* exponential backoff for transient provider failures
* support for `Retry-After`
* structured exception mapping into domain-level error codes
* validation retries for structured model responses
* explicit cancellation handling
* bounded concurrent PDF processing

This allows provider failures, rate limits, invalid structured responses, and infrastructure errors to be handled separately from application-level analysis failures.

## Observability

The service emits structured logs for:

* request lifecycle
* PDF processing
* LLM stage execution
* provider/model information
* normalized token usage
* queue wait time
* retries
* latency
* success and failure status

LLM execution telemetry is kept separate from provider-specific billing concerns.

## API

The service exposes a gRPC API defined by the shared protobuf contract.

The primary operation is:

```text
AnalysisEngine.Analyze
```

It accepts:

```text
resume_pdf
job_description_pdf
request_id
```

and returns:

```text
success
error_message
result_json
model
input_tokens
output_tokens
total_tokens
```

The result itself is returned as structured JSON generated from the domain-specific final schema.

## Current Validation

The service has a comprehensive automated test suite covering schemas, domain logic, PDF processing, LLM orchestration, result assembly, gRPC orchestration, and application lifecycle.

It also includes a real-provider end-to-end test that exercises the service through its Dockerized gRPC interface and validates the complete PDF → LLM → assembly pipeline.

## Deployment Model

The service is packaged as a Python 3.12 application and runs as a containerized asynchronous gRPC server.

Dependency resolution is locked through `uv.lock`, and the container uses the same locked dependency set for reproducible builds.

The Analysis Service is therefore the component that turns raw job-description and resume documents into a validated, structured assessment of candidate fit while keeping the analysis pipeline independent of any specific LLM provider or model.
