"""Canned LLM fixtures used when LLM_API_KEY is not configured.

The shape mirrors what the real LLM service produces. Outputs are
deterministic by-hash of the input so different JD/resume pairs give
different (but plausible) fixtures.

Content was originally adapted from the frontend's mock-analysis file to stay consistent
with what the frontend has been demonstrating so far.
"""

import hashlib
from typing import Literal

from app.schemas.analysis import (
    Evidence,
    RequirementMatch,
    SectionScore,
)

Flavor = Literal[
    "strong", "partial", "weak", "frontend-fit", "specialty-mismatch", "promising-grad"
]

FLAVORS: tuple[Flavor, ...] = (
    "strong",
    "partial",
    "weak",
    "frontend-fit",
    "specialty-mismatch",
    "promising-grad",
)


def _pick_flavor(jd_text: str, resume_text: str) -> Flavor:
    digest = hashlib.sha256(f"{jd_text}\n---\n{resume_text}".encode()).digest()
    idx = digest[0] % len(FLAVORS)
    return FLAVORS[idx]


def _strong() -> list[SectionScore]:
    return [
        SectionScore(
            id="skills",
            label="Skills",
            score=92,
            requirements=[
                RequirementMatch(
                    id="r-1",
                    requirement="4+ years of professional Python experience",
                    jd_evidence=Evidence(
                        text="Strong Python fundamentals with 4+ years of professional experience writing production code.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Backend Engineer, Stride Financial — Python, Django, PostgreSQL (Sep 2021 – present, 4.5 years).",
                            source="resume",
                            location="Experience",
                        )
                    ],
                ),
                RequirementMatch(
                    id="r-2",
                    requirement="Production experience with PostgreSQL",
                    jd_evidence=Evidence(
                        text="Comfortable designing PostgreSQL schemas and writing performant queries against large tables.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Owned the 2 TB Postgres cluster powering ledger reconciliation; rewrote three hot queries with composite indexes and CTEs, cutting p95 from 4.1s to 280ms.",
                            source="resume",
                            location="Projects",
                        )
                    ],
                ),
                RequirementMatch(
                    id="r-3",
                    requirement="Docker and container-based deployments",
                    jd_evidence=Evidence(
                        text="Familiarity with Docker; you should be able to write a Dockerfile without help.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Migrated 12 services from bare-metal to multi-stage Docker images; reduced average image size by 64%.",
                            source="resume",
                            location="Projects",
                        )
                    ],
                ),
            ],
        ),
        SectionScore(
            id="experience",
            label="Experience",
            score=88,
            requirements=[
                RequirementMatch(
                    id="r-4",
                    requirement="Has owned a production service end-to-end",
                    jd_evidence=Evidence(
                        text="You will be the primary on-call engineer for at least one production service.",
                        source="jd",
                        location="About the role",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Sole owner of the reconciliation service: design, implementation, on-call rotation (every fourth week), incident postmortems.",
                            source="resume",
                            location="Experience",
                        )
                    ],
                ),
                RequirementMatch(
                    id="r-5",
                    requirement="Experience working in a regulated or financial domain",
                    jd_evidence=Evidence(
                        text="Bonus: prior work in fintech, payments, or another regulated industry.",
                        source="jd",
                        location="Nice to have",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Stride Financial is a UK FCA-authorised lender; all production changes go through a documented change-control process.",
                            source="resume",
                            location="Experience",
                        )
                    ],
                ),
            ],
        ),
        SectionScore(
            id="education",
            label="Education",
            score=95,
            requirements=[
                RequirementMatch(
                    id="r-6",
                    requirement="Bachelor's degree in Computer Science or equivalent practical experience",
                    jd_evidence=Evidence(
                        text="BSc in Computer Science or equivalent practical experience.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="BSc Computer Science, University of Manchester (2017–2020), First Class Honours.",
                            source="resume",
                            location="Education",
                        )
                    ],
                ),
            ],
        ),
        SectionScore(
            id="leadership",
            label="Leadership Signals",
            score=80,
            requirements=[
                RequirementMatch(
                    id="r-7",
                    requirement="Has led a cross-team technical initiative",
                    jd_evidence=Evidence(
                        text="Examples of driving a project that required coordination across multiple teams.",
                        source="jd",
                        location="About the role",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Led the migration from synchronous webhook delivery to a queued, retried system. Required sign-off from Risk, Compliance and Platform.",
                            source="resume",
                            location="Projects",
                        )
                    ],
                ),
                RequirementMatch(
                    id="r-8",
                    requirement="Has written publicly-visible technical content",
                    jd_evidence=Evidence(
                        text="Bonus: a blog, conference talk, or open-source contribution we can read.",
                        source="jd",
                        location="Nice to have",
                    ),
                    matched=False,
                    match_strength="none",
                    resume_evidence=[],
                ),
            ],
        ),
    ]


def _partial() -> list[SectionScore]:
    return [
        SectionScore(
            id="skills",
            label="Skills",
            score=58,
            requirements=[
                RequirementMatch(
                    id="r-1",
                    requirement="Production Kubernetes experience",
                    jd_evidence=Evidence(
                        text="You will be operating multi-cluster Kubernetes on AWS EKS and on-prem.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=False,
                    match_strength="none",
                    resume_evidence=[],
                    note="The JD requires Kubernetes; the resume mentions Docker but not Kubernetes.",
                ),
                RequirementMatch(
                    id="r-2",
                    requirement="Terraform / IaC at scale",
                    jd_evidence=Evidence(
                        text="You should have hands-on Terraform experience and be comfortable refactoring a 50-module repo.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="partial",
                    resume_evidence=[
                        Evidence(
                            text="Wrote Terraform modules for VPC, RDS, and ALB across two AWS accounts. ~12 modules total.",
                            source="resume",
                            location="Projects",
                        )
                    ],
                    note="Resume shows ~12 modules; the JD implies a much larger codebase.",
                ),
                RequirementMatch(
                    id="r-3",
                    requirement="CI/CD pipeline ownership",
                    jd_evidence=Evidence(
                        text="Track record of owning CI/CD pipelines — failure budgets, flaky-test triage, deploy gates.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Designed and maintained the GitHub Actions workflows for 8 services: lint, type-check, unit, integration, deploy. Cut average CI time from 18 min to 6 min.",
                            source="resume",
                            location="Projects",
                        )
                    ],
                ),
            ],
        ),
        SectionScore(
            id="experience",
            label="Experience",
            score=65,
            requirements=[
                RequirementMatch(
                    id="r-4",
                    requirement="3+ years platform / infrastructure focused work",
                    jd_evidence=Evidence(
                        text="At least 3 years in an SRE, platform, or DevOps role.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="DevOps Engineer, Quill Press — 3 years 2 months (Feb 2023 – present).",
                            source="resume",
                            location="Experience",
                        )
                    ],
                ),
            ],
        ),
        SectionScore(
            id="education",
            label="Education",
            score=70,
            requirements=[
                RequirementMatch(
                    id="r-5",
                    requirement="Bachelor's degree or equivalent self-directed learning",
                    jd_evidence=Evidence(
                        text="Degree in a relevant field or demonstrable self-taught equivalent.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="partial",
                    resume_evidence=[
                        Evidence(
                            text="BSc Geography, University of Leeds (2018–2021). Self-taught into infrastructure; completed AWS Solutions Architect Associate (2023).",
                            source="resume",
                            location="Education",
                        )
                    ],
                ),
            ],
        ),
        SectionScore(
            id="leadership",
            label="Leadership Signals",
            score=50,
            requirements=[
                RequirementMatch(
                    id="r-6",
                    requirement="Has written or co-authored incident postmortems",
                    jd_evidence=Evidence(
                        text="Comfortable writing blameless postmortems and presenting them to the wider team.",
                        source="jd",
                        location="About the role",
                    ),
                    matched=True,
                    match_strength="partial",
                    resume_evidence=[
                        Evidence(
                            text="Wrote three blameless postmortems in the last year. Two were for incidents I primarily resolved.",
                            source="resume",
                            location="Experience",
                        )
                    ],
                ),
            ],
        ),
    ]


def _weak() -> list[SectionScore]:
    return [
        SectionScore(
            id="skills",
            label="Skills",
            score=22,
            requirements=[
                RequirementMatch(
                    id="r-1",
                    requirement="Deep learning research or production experience with PyTorch or JAX",
                    jd_evidence=Evidence(
                        text="You will work directly with our research team on PyTorch (and increasingly JAX) training code.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=False,
                    match_strength="none",
                    resume_evidence=[],
                    note="The JD requires PyTorch or JAX; the resume lists scikit-learn and pandas only.",
                ),
                RequirementMatch(
                    id="r-2",
                    requirement="Python (general-purpose, production-quality)",
                    jd_evidence=Evidence(
                        text="Strong Python is expected — the whole team writes Python every day.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="partial",
                    resume_evidence=[
                        Evidence(
                            text="Python (Flask) is the primary stack at Brightline; I write Python for both backend services and analysis scripts.",
                            source="resume",
                            location="Experience",
                        )
                    ],
                    note="Python experience is web-framework focused; JD context is research / training code.",
                ),
            ],
        ),
        SectionScore(
            id="experience",
            label="Experience",
            score=18,
            requirements=[
                RequirementMatch(
                    id="r-3",
                    requirement="5+ years in an applied ML or research engineering role",
                    jd_evidence=Evidence(
                        text="At least 5 years in a role focused primarily on machine learning, deep learning, or research engineering.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=False,
                    match_strength="none",
                    resume_evidence=[],
                    note="The JD requires 5+ years in an ML role; the resume shows 1.5 years total post-graduation, in a generalist software role.",
                ),
            ],
        ),
        SectionScore(
            id="education",
            label="Education",
            score=35,
            requirements=[
                RequirementMatch(
                    id="r-4",
                    requirement="MSc or PhD in a quantitative field, or equivalent published work",
                    jd_evidence=Evidence(
                        text="MSc or PhD in CS, Math, Physics, or a related quantitative field. Equivalent demonstrated research output also accepted.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=False,
                    match_strength="none",
                    resume_evidence=[],
                    note="The JD requires MSc/PhD or research output; the resume shows only a BSc with no publications.",
                ),
                RequirementMatch(
                    id="r-5",
                    requirement="Bachelor's in a technical field",
                    jd_evidence=Evidence(
                        text="Undergraduate degree in Computer Science, Mathematics, or similar.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="BSc Computer Science, Auckland University of Technology (2019–2022).",
                            source="resume",
                            location="Education",
                        )
                    ],
                ),
            ],
        ),
        SectionScore(
            id="leadership",
            label="Leadership Signals",
            score=30,
            requirements=[
                RequirementMatch(
                    id="r-6",
                    requirement="Has owned an ML project from scoping to deployment",
                    jd_evidence=Evidence(
                        text="Track record of owning a project end-to-end: scoping, dataset work, training, evaluation, deployment.",
                        source="jd",
                        location="About the role",
                    ),
                    matched=False,
                    match_strength="none",
                    resume_evidence=[],
                    note="The JD requires end-to-end ML ownership; the resume shows only individual contributions within larger product features.",
                ),
            ],
        ),
    ]


def _frontend_fit() -> list[SectionScore]:
    """Mid-level frontend candidate against a frontend role. Strong skill
    overlap but lighter on the leadership signals the JD asked for."""
    return [
        SectionScore(
            id="skills",
            label="Skills",
            score=82,
            requirements=[
                RequirementMatch(
                    id="r-1",
                    requirement="3+ years building production React applications with TypeScript",
                    jd_evidence=Evidence(
                        text="3+ years building production React applications with TypeScript.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Frontend Engineer, Quill Press (Mar 2022 - present, 3 years). Built features in the reader app: TypeScript + React + Tailwind.",
                            source="resume",
                            location="Experience",
                        )
                    ],
                ),
                RequirementMatch(
                    id="r-2",
                    requirement="Has shipped accessible UI work",
                    jd_evidence=Evidence(
                        text="Has shipped accessible UI work - keyboard navigation, screen-reader testing, focus management.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Owned a11y compliance for the new checkout flow: keyboard navigation, screen-reader testing with VoiceOver and NVDA, focus management.",
                            source="resume",
                            location="Projects",
                        )
                    ],
                ),
                RequirementMatch(
                    id="r-3",
                    requirement="Comfortable with modern frontend tooling (Vite, pnpm, Playwright)",
                    jd_evidence=Evidence(
                        text="Comfortable with modern frontend tooling (Vite, pnpm, Playwright).",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Set up Playwright end-to-end testing for the reader app.",
                            source="resume",
                            location="Experience",
                        )
                    ],
                ),
            ],
        ),
        SectionScore(
            id="experience",
            label="Experience",
            score=78,
            requirements=[
                RequirementMatch(
                    id="r-4",
                    requirement="Working inside a design-system repo",
                    jd_evidence=Evidence(
                        text="Comfortable working inside a design-system repo (we use shadcn-style primitives and Tailwind).",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Migrated the design-system repo from styled-components to Tailwind + shadcn primitives. Coordinated rollout across four product teams.",
                            source="resume",
                            location="Projects",
                        )
                    ],
                ),
            ],
        ),
        SectionScore(
            id="education",
            label="Education",
            score=72,
            requirements=[
                RequirementMatch(
                    id="r-5",
                    requirement="Bachelor's degree or equivalent self-directed learning",
                    jd_evidence=Evidence(
                        text="Bachelor's degree or equivalent self-directed learning.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="BA Design and Computing, Goldsmiths (2017-2020).",
                            source="resume",
                            location="Education",
                        )
                    ],
                ),
            ],
        ),
        SectionScore(
            id="leadership",
            label="Leadership Signals",
            score=62,
            requirements=[
                RequirementMatch(
                    id="r-6",
                    requirement="Mentoring more junior engineers",
                    jd_evidence=Evidence(
                        text="Experience mentoring more junior engineers and reviewing their work.",
                        source="jd",
                        location="Nice to have",
                    ),
                    matched=True,
                    match_strength="partial",
                    resume_evidence=[
                        Evidence(
                            text="Reviewed pull requests from two junior engineers; mentored both through their first six months.",
                            source="resume",
                            location="Experience",
                        )
                    ],
                ),
                RequirementMatch(
                    id="r-7",
                    requirement="Open-source contributions to a design system",
                    jd_evidence=Evidence(
                        text="Open-source contributions to a design system or component library.",
                        source="jd",
                        location="Nice to have",
                    ),
                    matched=False,
                    match_strength="none",
                    resume_evidence=[],
                ),
            ],
        ),
    ]


def _specialty_mismatch() -> list[SectionScore]:
    """Senior backend candidate against a non-backend specialty role (e.g.
    ML or platform). Strong overall engineering signal, but the JD specialty
    requirements aren't met."""
    return [
        SectionScore(
            id="skills",
            label="Skills",
            score=42,
            requirements=[
                RequirementMatch(
                    id="r-1",
                    requirement="Specialty-specific technical stack",
                    jd_evidence=Evidence(
                        text="Day-to-day work in the team's specialty stack is required.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=False,
                    match_strength="none",
                    resume_evidence=[],
                    note="The JD requires a specialty stack the resume doesn't list. The candidate's strengths are in an adjacent area.",
                ),
                RequirementMatch(
                    id="r-2",
                    requirement="General-purpose Python production experience",
                    jd_evidence=Evidence(
                        text="Strong Python is expected - the whole team writes Python every day.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Backend Engineer, Stride Financial - Python, Django, PostgreSQL (Sep 2021 - present, 4.5 years).",
                            source="resume",
                            location="Experience",
                        )
                    ],
                ),
            ],
        ),
        SectionScore(
            id="experience",
            label="Experience",
            score=55,
            requirements=[
                RequirementMatch(
                    id="r-3",
                    requirement="Has owned a production service end-to-end",
                    jd_evidence=Evidence(
                        text="You will own a production service end-to-end.",
                        source="jd",
                        location="About the role",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Sole owner of the reconciliation service: design, implementation, on-call rotation, incident postmortems.",
                            source="resume",
                            location="Experience",
                        )
                    ],
                ),
                RequirementMatch(
                    id="r-4",
                    requirement="Domain experience in the team's specialty",
                    jd_evidence=Evidence(
                        text="Hands-on experience in the specialty area is required.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=False,
                    match_strength="none",
                    resume_evidence=[],
                    note="The resume shows backend product work; the JD asks for a different specialty.",
                ),
            ],
        ),
        SectionScore(
            id="education",
            label="Education",
            score=80,
            requirements=[
                RequirementMatch(
                    id="r-5",
                    requirement="Bachelor's in a technical field",
                    jd_evidence=Evidence(
                        text="Undergraduate degree in Computer Science, Mathematics, or similar.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="BSc Computer Science, University of Manchester (2017-2020), First Class Honours.",
                            source="resume",
                            location="Education",
                        )
                    ],
                ),
            ],
        ),
        SectionScore(
            id="leadership",
            label="Leadership Signals",
            score=65,
            requirements=[
                RequirementMatch(
                    id="r-6",
                    requirement="Has led a cross-team technical initiative",
                    jd_evidence=Evidence(
                        text="Examples of driving a project that required coordination across multiple teams.",
                        source="jd",
                        location="About the role",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Led the migration from synchronous webhook delivery to a queued, retried system. Required sign-off from Risk, Compliance and Platform.",
                            source="resume",
                            location="Projects",
                        )
                    ],
                ),
            ],
        ),
    ]


def _promising_grad() -> list[SectionScore]:
    """Recent CS graduate against a junior-level role (e.g. junior data eng).
    Education + foundational skills score well, experience is light."""
    return [
        SectionScore(
            id="skills",
            label="Skills",
            score=60,
            requirements=[
                RequirementMatch(
                    id="r-1",
                    requirement="Solid SQL fundamentals",
                    jd_evidence=Evidence(
                        text="Solid SQL fundamentals; can write a multi-CTE query and reason about query plans.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Built a SQL-driven movie recommender for a databases course assignment. Wrote ~15 queries against the IMDB dataset using window functions and CTEs.",
                            source="resume",
                            location="Projects",
                        )
                    ],
                ),
                RequirementMatch(
                    id="r-2",
                    requirement="Working knowledge of Python",
                    jd_evidence=Evidence(
                        text="Working knowledge of Python for scripting and data manipulation.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="partial",
                    resume_evidence=[
                        Evidence(
                            text="Wrote Python scripts to migrate customer data; gained some exposure to Airflow.",
                            source="resume",
                            location="Internships",
                        )
                    ],
                ),
                RequirementMatch(
                    id="r-3",
                    requirement="Familiarity with at least one orchestration tool",
                    jd_evidence=Evidence(
                        text="Familiarity with at least one orchestration tool (Airflow, Dagster, Prefect - any of them is fine).",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="weak",
                    resume_evidence=[
                        Evidence(
                            text="...gained some exposure to Airflow.",
                            source="resume",
                            location="Internships",
                        )
                    ],
                    note='Exposure only; the JD says "any of them is fine" so this clears the bar narrowly.',
                ),
            ],
        ),
        SectionScore(
            id="experience",
            label="Experience",
            score=38,
            requirements=[
                RequirementMatch(
                    id="r-4",
                    requirement="0-2 years professional experience",
                    jd_evidence=Evidence(
                        text="0-2 years professional experience. Recent graduates encouraged to apply.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="Software Engineering Intern, smaller startup (May 2024 - Aug 2024, 4 months).",
                            source="resume",
                            location="Internships",
                        )
                    ],
                ),
            ],
        ),
        SectionScore(
            id="education",
            label="Education",
            score=85,
            requirements=[
                RequirementMatch(
                    id="r-5",
                    requirement="Undergraduate degree in CS, Math, or similar",
                    jd_evidence=Evidence(
                        text="Undergraduate degree in Computer Science, Mathematics, or similar.",
                        source="jd",
                        location="Requirements",
                    ),
                    matched=True,
                    match_strength="strong",
                    resume_evidence=[
                        Evidence(
                            text="BSc Computer Science, University of Waterloo (2021-2025). GPA 3.7 / 4.0. Coursework: Algorithms, Databases, Distributed Systems, Linear Algebra.",
                            source="resume",
                            location="Education",
                        )
                    ],
                ),
            ],
        ),
        SectionScore(
            id="leadership",
            label="Leadership Signals",
            score=25,
            requirements=[
                RequirementMatch(
                    id="r-6",
                    requirement="Eager to learn at a place with senior review",
                    jd_evidence=Evidence(
                        text="You want to learn data engineering at a place where senior engineers actually review your code.",
                        source="jd",
                        location="About you",
                    ),
                    matched=True,
                    match_strength="weak",
                    resume_evidence=[
                        Evidence(
                            text="Contributed a small bug fix to a Python testing library on GitHub (one merged PR).",
                            source="resume",
                            location="Projects",
                        )
                    ],
                    note="Open-source contribution shows engagement, but no formal mentee or leadership signals on the resume.",
                ),
            ],
        ),
    ]


_FIXTURES = {
    "strong": _strong,
    "partial": _partial,
    "weak": _weak,
    "frontend-fit": _frontend_fit,
    "specialty-mismatch": _specialty_mismatch,
    "promising-grad": _promising_grad,
}


def stub_sections(jd_text: str, resume_text: str) -> list[SectionScore]:
    """Deterministic-by-input stub. Returns a fully-populated SectionScore[]
    matching the wire contract. The worker can call this and the response
    looks identical to a real LLM-driven analysis."""
    flavor = _pick_flavor(jd_text, resume_text)
    return _FIXTURES[flavor]()
