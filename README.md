# ResumeRanker

ResumeRanker is an evidence-based hiring intelligence tool that compares a candidate resume against a job description and identifies how well the applicant matches the role.

It helps recruiters, hiring teams, and hiring managers answer a simple but critical question: whether a candidate is a strong fit for the job based on concrete evidence from the resume and the requirements, not just keyword overlap.

## What the product does

ResumeRanker analyzes:

- a candidate resume
- a target job description
- role-specific skills, requirements, and experience signals

Then it produces a structured assessment that highlights:

- skills match and gaps
- experience alignment
- strengths and weaknesses
- role-fit confidence backed by extracted evidence

The goal is to move hiring evaluation from intuition toward transparent, explainable comparison.

## System overview

```text
Browser
   |
   v
Caddy Reverse Proxy
   |
   v
Streamlit Analysis Dashboard
   |
   v
Analysis gRPC Service
   |
   v
Resume + job analysis pipeline
```

## Core components

### Analysis Service

[`analysis/`](analysis/) is the backend engine. It processes the resume and job description, extracts structured claims and requirements, compares them, and returns a scored fit result with supporting evidence.

### Analysis Dashboard

[`analysis_dashboard/`](analysis_dashboard/) is the user-facing interface. It presents the analysis in a readable format and helps users explore candidate fit across sections and evidence points.

### Caddy Proxy

[`proxy/`](proxy/) is the public entry point and reverse proxy. It fronts the dashboard and handles secure HTTP routing for the deployed application.

### Shared Contract

[`proto/analysis.proto`](proto/analysis.proto) defines the shared contract between the dashboard and analysis service, keeping the pipeline modular and consistent.

## Why it matters

Many hiring workflows are still driven by manual CV review, shallow keyword matching, or subjective interpretation. ResumeRanker introduces a more structured and evidence-backed approach to candidate evaluation.

It is designed for teams that want:

- faster screening
- clearer alignment signals
- less guesswork in early-stage hiring decisions
- a more transparent candidate assessment workflow

## Product focus

This repository is centered on a practical hiring decision support product:

- analyze candidate fit against job requirements
- expose the evidence behind the score
- support streamlined review in a dashboard experience
- keep the service architecture modular and extensible

---

## License

See [LICENSE](LICENSE) for the repository licensing terms.
