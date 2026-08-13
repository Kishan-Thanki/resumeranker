from collections.abc import Iterator

import pytest
from app.domain.tech import TechDomain
from app.schemas import (
    Evidence,
    ExtractedRequirement,
    MatchVerdict,
    ResumeClaim,
    SectionScore,
    SectionVerdict,
)

# ============================================================================
# SOURCE TEXT FIXTURES
# ============================================================================


@pytest.fixture
def sample_jd_text() -> str:
    """Small deterministic JD text used by schema/PDF/LLM tests."""
    return (
        "We are looking for a Senior Backend Engineer with 5+ years "
        "of Python experience. Strong experience with FastAPI and PostgreSQL "
        "is required. Experience with AWS is preferred. "
        "Bachelor's degree in Computer Science is preferred."
    )


@pytest.fixture
def sample_resume_text() -> str:
    """Small deterministic resume text used by schema/LLM tests."""
    return (
        "Senior Backend Engineer with 6 years of Python development. "
        "Built production APIs with FastAPI and PostgreSQL at Acme Corp. "
        "Deployed backend services on AWS. "
        "Bachelor of Technology in Computer Science."
    )


# ============================================================================
# DOMAIN FIXTURES
# ============================================================================


@pytest.fixture
def tech_domain() -> TechDomain:
    """Fresh TechDomain instance for domain tests."""
    return TechDomain()


# ============================================================================
# EVIDENCE FIXTURES
# ============================================================================


@pytest.fixture
def jd_evidence() -> Evidence:
    """Valid JD evidence."""
    return Evidence(
        text="Strong experience with FastAPI and PostgreSQL is required.",
        source="jd",
        location="Requirements",
    )


@pytest.fixture
def resume_evidence() -> Evidence:
    """Valid resume evidence."""
    return Evidence(
        text="Built production APIs with FastAPI and PostgreSQL at Acme Corp.",
        source="resume",
        location="Experience",
    )


# ============================================================================
# EXTRACTION FIXTURES
# ============================================================================


@pytest.fixture
def sample_requirement(jd_evidence: Evidence) -> ExtractedRequirement:
    """A single valid extracted JD requirement."""
    return ExtractedRequirement(
        id="req-1",
        section="skills",
        requirement="Strong FastAPI and PostgreSQL experience",
        jd_evidence=jd_evidence,
    )


@pytest.fixture
def sample_requirements(jd_evidence: Evidence) -> list[ExtractedRequirement]:
    """Several valid JD requirements covering multiple sections."""
    return [
        ExtractedRequirement(
            id="req-1",
            section="skills",
            requirement="Strong FastAPI and PostgreSQL experience",
            jd_evidence=jd_evidence,
        ),
        ExtractedRequirement(
            id="req-2",
            section="experience",
            requirement="5+ years of Python experience",
            jd_evidence=Evidence(
                text="with 5+ years of Python experience.",
                source="jd",
                location="Requirements",
            ),
        ),
        ExtractedRequirement(
            id="req-3",
            section="education",
            requirement="Bachelor's degree in Computer Science",
            jd_evidence=Evidence(
                text="Bachelor's degree in Computer Science is preferred.",
                source="jd",
                location="Requirements",
            ),
        ),
    ]


@pytest.fixture
def sample_claim(resume_evidence: Evidence) -> ResumeClaim:
    """A single valid extracted resume claim."""
    return ResumeClaim(
        id="claim-1",
        section="skills",
        claim="Production FastAPI and PostgreSQL experience",
        resume_evidence=resume_evidence,
    )


@pytest.fixture
def sample_claims(resume_evidence: Evidence) -> list[ResumeClaim]:
    """Several valid resume claims covering multiple sections."""
    return [
        ResumeClaim(
            id="claim-1",
            section="skills",
            claim="Production FastAPI and PostgreSQL experience",
            resume_evidence=resume_evidence,
        ),
        ResumeClaim(
            id="claim-2",
            section="experience",
            claim="Six years of Python development",
            resume_evidence=Evidence(
                text="Senior Backend Engineer with 6 years of Python development.",
                source="resume",
                location="Experience",
            ),
        ),
        ResumeClaim(
            id="claim-3",
            section="education",
            claim="Computer Science bachelor's degree",
            resume_evidence=Evidence(
                text="Bachelor of Technology in Computer Science.",
                source="resume",
                location="Education",
            ),
        ),
    ]


# ============================================================================
# SCORING FIXTURES
# ============================================================================


@pytest.fixture
def sample_match_verdict() -> MatchVerdict:
    """A valid strong match verdict for one requirement."""
    return MatchVerdict(
        id="req-1",
        match_strength="strong",
        supporting_claim_ids=["claim-1"],
        note=None,
    )


@pytest.fixture
def sample_match_verdicts() -> list[MatchVerdict]:
    """Valid verdicts covering the sample requirement IDs."""
    return [
        MatchVerdict(
            id="req-1",
            match_strength="strong",
            supporting_claim_ids=["claim-1"],
        ),
        MatchVerdict(
            id="req-2",
            match_strength="strong",
            supporting_claim_ids=["claim-2"],
        ),
        MatchVerdict(
            id="req-3",
            match_strength="strong",
            supporting_claim_ids=["claim-3"],
        ),
    ]


@pytest.fixture
def sample_section_verdicts() -> list[SectionVerdict]:
    """Valid verdicts covering the complete Tech taxonomy."""
    return [
        SectionVerdict(
            id="skills",
            score=90,
            review="Strong coverage of the required technical skills.",
        ),
        SectionVerdict(
            id="experience",
            score=95,
            review="The resume demonstrates more than the required Python experience.",
        ),
        SectionVerdict(
            id="education",
            score=90,
            review="The candidate has the requested Computer Science education.",
        ),
        SectionVerdict(
            id="project",
            score=20,
            review="The available resume evidence does not address a specific project requirement.",
        ),
    ]


# ============================================================================
# FINAL RESULT FIXTURES
# ============================================================================


@pytest.fixture
def sample_section_score() -> SectionScore:
    """Minimal valid final-domain section score."""
    return SectionScore(
        id="skills",
        label="Skills",
        score=90,
        review="Strong technical skill coverage.",
        requirements=[],
    )


# ============================================================================
# ASSEMBLY HELPERS
# ============================================================================


@pytest.fixture
def claims_by_id(sample_claims: list[ResumeClaim]) -> dict[str, ResumeClaim]:
    """Convenience mapping used by RequirementMatch assembly tests."""
    return {
        claim.id: claim
        for claim in sample_claims
        if claim.id is not None
    }


@pytest.fixture
def requirements_by_id(
    sample_requirements: list[ExtractedRequirement],
) -> dict[str, ExtractedRequirement]:
    """Convenience mapping used by assembly tests."""
    return {
        requirement.id: requirement
        for requirement in sample_requirements
        if requirement.id is not None
    }


# ============================================================================
# ENVIRONMENT HELPERS
# ============================================================================


@pytest.fixture
def clean_llm_environment(monkeypatch: pytest.MonkeyPatch) -> Iterator[None]:
    """
    Removes LLM-related environment overrides for tests that need to verify
    application defaults.
    """
    variables = (
        "LLM_PROVIDER",
        "LLM_MODEL",
        "LLM_API_KEY",
        "MAX_CONCURRENT_LLM_REQUESTS",
        "LLM_MAX_REQUESTS_PER_MINUTE",
        "MAX_VALIDATION_RETRIES",
        "MAX_RATE_LIMIT_RETRIES",
        "BACKOFF_BASE_SECONDS",
    )

    for variable in variables:
        monkeypatch.delenv(variable, raising=False)

    yield
