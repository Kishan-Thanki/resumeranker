# Future Architecture

This document serves as a living record of architectural brainstorms and edge cases we discussed during the design of the Analysis Engine.

## [P1] Dynamic Domain Strategy Routing

**Current State:** The backend fully supports dynamic domains (via the abstract `DomainStrategy` class which dynamically generates its own Pydantic schemas). However, the engine still hardcodes `domain = TechDomain()` in `llm_service.py` for all incoming requests.

**The Problem:** When we introduce `SalesDomain` or `HealthcareDomain`, the Engine needs to know which domain to invoke.

**Future Solutions:**

1. **Explicit (UI):** Add an "Industry" dropdown to the frontend, pass it through the gRPC `AnalyzeRequest`, and use a Python Factory pattern to load the correct domain class.
2. **Implicit (Router LLM):** Pass the first 500 characters of the Job Description to a cheap, ultra-fast model (like `gemini-2.5-flash-8b`) to classify the industry automatically. This provides a magical user experience and prevents users from intentionally selecting the wrong domain.
