# ResumeRanker Analysis Dashboard

Standalone Streamlit frontend for ResumeRanker Analysis.

This package is intentionally self-sufficient:

- it only depends on the shared protobuf contract
- it talks directly to the analysis gRPC service
- it does not depend on the older `web/` frontend or HTTP dashboard

The project uses a standard `src/` layout, with the real importable package
under `src/analysis_dashboard/`.

## Architecture

```text
src/analysis_dashboard/
├── app.py         # Streamlit entrypoint
├── main.py        # Page orchestration
├── config.py      # Runtime configuration and constants
├── grpc_client.py # gRPC transport
├── parsing.py     # JSON and upload validation
├── rendering.py   # Streamlit rendering helpers
├── state.py       # Session-state helpers
└── styles.py      # Inline CSS
```

## What It Shows

- executive summary from `completeAnalysis`
- section-level scores and reviews
- requirement-level verdicts
- JD and resume evidence side by side
- report-style overview and section views

## Configuration

The UI reads environment variables from `analysis_dashboard/.env` when present.

Required runtime settings:

- `ANALYSIS_GRPC_HOST`
- `ANALYSIS_GRPC_PORT`

Optional settings:

- `ANALYSIS_GRPC_TIMEOUT_SECONDS`
- `ANALYSIS_GRPC_ADDRESS` if you want to document a single endpoint for humans

## Run

```bash
cd analysis_dashboard
python -m pip install --no-build-isolation -e .
streamlit run src/analysis_dashboard/app.py
```

Make sure the analysis service is already running and reachable on the configured host and port.
