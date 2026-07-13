# Project Context: ResumeRanker Analysis Engine

You are looking at the `@analysis/` directory of the ResumeRanker project.

## Purpose
This engine receives a Job Description PDF and a Resume PDF, reads them, and uses an LLM to score how well the candidate matches the job.

## Core Technologies
- **Python 3.12+**
- **gRPC (`grpcio`, `grpcio-tools`)**: For networking.
- **Pydantic V2**: For rigorous data validation and JSON schema generation.
- **Instructor**: To force the LLM to return data matching the Pydantic schemas.
- **LiteLLM**: To proxy the LLM calls so we can switch between Gemini/OpenAI effortlessly.
- **uv**: For blazing fast, deterministic dependency management.
- **pdfplumber**: For raw text extraction from PDFs.

## The 3-Step LLM Pipeline (`llm_service.py`)
1. **Extract Job Requirements:** Ask the LLM to read the JD and output an array of `ExtractedRequirement`.
2. **Extract Resume Claims:** Ask the LLM to read the Resume and output an array of `ResumeClaim` (with evidence).
3. **Score the Match:** Send the extracted JD requirements and the extracted Resume claims to the LLM, and ask it to output a `FinalAnalysisResult` containing individual scores (0-100) and rationale.
