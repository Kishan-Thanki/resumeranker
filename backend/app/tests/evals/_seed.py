"""Source-of-truth fixture content for the eval harness.

Run as a script to materialize the file tree:

    docker compose exec api uv run python -m app.tests.evals._seed

This keeps the 35 fixture files (5 JDs + 5 resumes + 25 expected) in sync
with their authored data. Edit the constants below to retune fixtures.
"""

from __future__ import annotations

import json
import pathlib

HERE = pathlib.Path(__file__).parent
FIXTURES = HERE / "fixtures"
EXPECTED_DIR = HERE / "expected"


# --- 5 JD fixtures ----------------------------------------------------------

JDS: dict[str, str] = {
    "jd-backend-senior": """\
Senior Backend Engineer at Acme Financial

About the role
You will own one or more production services in our payments stack. You'll
participate in the on-call rotation and write blameless postmortems when
things go wrong. You will coordinate with Risk and Compliance on
infrastructure changes.

Requirements
- 4+ years of professional Python experience writing production code.
- Comfortable designing PostgreSQL schemas and writing performant queries
  against large tables.
- Familiarity with Docker; you should be able to write a Dockerfile without
  help.
- Practical experience deploying services to AWS (ECS/Fargate or EKS).
- BSc in Computer Science or equivalent practical experience.

Nice to have
- Prior work in fintech, payments, or another regulated industry.
- A blog, conference talk, or open-source contribution we can read.
""",
    "jd-platform-eng": """\
Platform Engineer at Nexus Robotics

About the role
You will be operating multi-cluster Kubernetes on AWS EKS and on-prem.
You'll participate in the on-call rotation for customer-facing services
and write blameless postmortems for incidents.

Requirements
- At least 3 years in an SRE, platform, or DevOps role.
- You should have hands-on Terraform experience and be comfortable
  refactoring a 50-module repo.
- Comfortable instrumenting services with Prometheus metrics and authoring
  Grafana dashboards.
- Most of our internal tooling is Go. Prior Go experience is required.
- Track record of owning CI/CD pipelines — failure budgets, flaky-test
  triage, deploy gates.
- Degree in a relevant field or demonstrable self-taught equivalent.

Nice to have
- Examples of introducing a new tool or workflow that other teams adopted.
""",
    "jd-ml-senior": """\
Senior Machine Learning Engineer at Lumen AI

About the role
You will work directly with our research team on PyTorch (and increasingly
JAX) training code. You'll own ML projects end-to-end: scoping, dataset
work, training, evaluation, deployment. You will set the technical
direction for two junior MLEs joining in the next quarter.

Requirements
- Practical experience with distributed training: DDP, FSDP, or DeepSpeed
  across at least 8 GPUs.
- Has shipped a model to production users and operated it under real load
  (SageMaker, Triton, or similar).
- At least 5 years in a role focused primarily on machine learning, deep
  learning, or research engineering.
- MSc or PhD in CS, Math, Physics, or a related quantitative field.
  Equivalent demonstrated research output also accepted.
- Strong Python is expected — the whole team writes Python every day.

About you
- You can point to a feature in a shipping product where the ML you built
  is doing the work.
- You should be able to read a recent NeurIPS paper and reason about the
  loss function.
""",
    "jd-frontend-mid": """\
Mid-level Frontend Engineer at Quill Publishing

About the role
You will build customer-facing features in our reader app and contribute to
our shared design system. You'll review pull requests from junior
engineers and pair on tricky UI work.

Requirements
- 3+ years building production React applications with TypeScript.
- Comfortable working inside a design-system repo (we use shadcn-style
  primitives and Tailwind).
- Has shipped accessible UI work — keyboard navigation, screen-reader
  testing, focus management.
- Comfortable with modern frontend tooling (Vite, pnpm, Playwright).
- Bachelor's degree or equivalent self-directed learning.

Nice to have
- Experience mentoring more junior engineers and reviewing their work.
- Open-source contributions to a design system or component library.
""",
    "jd-data-eng-junior": """\
Junior Data Engineer at Brightline Analytics

About the role
You will build and maintain ETL pipelines moving data from our application
databases into the warehouse. You'll write SQL transforms, monitor
pipeline runs, and triage failures.

Requirements
- 0–2 years professional experience. Recent graduates encouraged to apply.
- Solid SQL fundamentals; can write a multi-CTE query and reason about
  query plans.
- Working knowledge of Python for scripting and data manipulation.
- Familiarity with at least one orchestration tool (Airflow, Dagster,
  Prefect — any of them is fine).
- Undergraduate degree in Computer Science, Mathematics, or similar.

About you
- You want to learn data engineering at a place where senior engineers
  actually review your code.
""",
}


# --- 5 resume fixtures ------------------------------------------------------

RESUMES: dict[str, str] = {
    "resume-backend-senior": """\
Priya Sharma — Senior Backend Engineer

Experience
Backend Engineer, Stride Financial — Python, Django, PostgreSQL
(Sep 2021 – present, 4.5 years)
- Sole owner of the reconciliation service: design, implementation,
  on-call rotation (every fourth week), incident postmortems.
- Stride Financial is a UK FCA-authorised lender; all production changes
  go through a documented change-control process.
- Onboarded two graduate hires in 2024; pair-programmed during their
  first month.

Projects
- Owned the 2 TB Postgres cluster powering ledger reconciliation; rewrote
  three hot queries with composite indexes and CTEs, cutting p95 from
  4.1s to 280ms.
- Migrated 12 services from bare-metal to multi-stage Docker images;
  reduced average image size by 64%.
- Day-to-day work in AWS: ECS Fargate, RDS, S3, CloudWatch. Wrote the
  Terraform that provisions our staging environment.
- Led the migration from synchronous webhook delivery to a queued,
  retried system. Required sign-off from Risk, Compliance and Platform.

Education
BSc Computer Science, University of Manchester (2017–2020),
First Class Honours.

Skills
Python, PostgreSQL, Docker, AWS, Terraform, Django, SQLAlchemy.
""",
    "resume-devops-generalist": """\
James Okafor — DevOps Engineer

Experience
DevOps Engineer, Quill Press — 3 years 2 months (Feb 2023 – present).
- Secondary on-call for the publishing platform; primary on-call sits
  with the application engineering team.
- Wrote three blameless postmortems in the last year. Two were for
  incidents I primarily resolved.

Projects
- Wrote Terraform modules for VPC, RDS, and ALB across two AWS accounts.
  ~12 modules total.
- Designed and maintained the GitHub Actions workflows for 8 services:
  lint, type-check, unit, integration, deploy. Cut average CI time from
  18 min to 6 min.
- Maintained CloudWatch dashboards and alerts for the publishing platform.

Skills
Python, Bash, Docker, Docker Compose, Terraform, AWS, GitHub Actions,
CloudWatch.

Education
BSc Geography, University of Leeds (2018–2021). Self-taught into
infrastructure; completed AWS Solutions Architect Associate (2023).
""",
    "resume-junior-fullstack": """\
Alex Chen — Software Engineer

Experience
Software Engineer, Brightline Analytics (Jan 2025 – present, 4 months)
Software Engineer Intern, Brightline Analytics (Jun 2024 – Dec 2024,
6 months, returning intern).
- Python (Flask) is the primary stack at Brightline; I write Python for
  both backend services and analysis scripts.
- Shipped a churn-prediction notebook shared with the analytics team.

Education
BSc Computer Science, Auckland University of Technology (2019–2022).
Coursework: Linear Algebra, Calculus I–II, Probability & Statistics
(undergraduate, 2022).

Skills
Python, Flask, JavaScript, React, scikit-learn, pandas, SQL, Git.
""",
    "resume-frontend-mid": """\
Maya Lopez — Frontend Engineer

Experience
Frontend Engineer, Quill Press (Mar 2022 – present, 3 years).
- Built features in the reader app: TypeScript + React + Tailwind.
- Reviewed pull requests from two junior engineers; mentored both
  through their first six months.
- Set up Playwright end-to-end testing for the reader app.

Projects
- Migrated the design-system repo from styled-components to Tailwind +
  shadcn primitives. Coordinated rollout across four product teams.
- Owned a11y compliance for the new checkout flow: keyboard navigation,
  screen-reader testing with VoiceOver and NVDA, focus management.

Skills
TypeScript, React, Tailwind CSS, Vite, pnpm, Playwright, shadcn,
HTML / CSS, accessibility tooling.

Education
BA Design and Computing, Goldsmiths (2017–2020).
""",
    "resume-recent-grad": """\
Sam Patel — Recent Graduate

Education
BSc Computer Science, University of Waterloo (2021–2025).
GPA 3.7 / 4.0. Coursework: Algorithms, Databases, Distributed Systems,
Linear Algebra.

Projects
- Built a multi-user todo app with Flask + SQLite for my final-year
  software-engineering course.
- Built a SQL-driven movie recommender for a databases course assignment.
  Wrote ~15 queries against the IMDB dataset using window functions
  and CTEs.
- Contributed a small bug fix to a Python testing library on GitHub
  (one merged PR).

Internships
- Software Engineering Intern, smaller startup (May 2024 – Aug 2024, 4
  months). Wrote Python scripts to migrate customer data; gained some
  exposure to Airflow.

Skills
Python, SQL, Java (school only), Git, basic Docker.
""",
}


# --- 25 expected-range entries ---------------------------------------------
# Sections in order: skills, experience, education, leadership.
# Each cell is (min, max) for the section score.
# Hand-tuned per (jd, resume) pair: 4 numbers per section × 4 sections.

MatrixCell = dict[str, tuple[int, int]]

EXPECTED: dict[tuple[str, str], MatrixCell] = {
    # ---------- jd-backend-senior ------------------------------------------
    ("jd-backend-senior", "resume-backend-senior"): {
        "skills": (75, 100),
        "experience": (70, 100),
        "education": (80, 100),
        "leadership": (60, 95),
    },
    ("jd-backend-senior", "resume-devops-generalist"): {
        "skills": (40, 70),
        "experience": (40, 70),
        "education": (50, 80),
        "leadership": (30, 65),
    },
    ("jd-backend-senior", "resume-junior-fullstack"): {
        "skills": (25, 55),
        "experience": (10, 40),
        "education": (70, 95),
        "leadership": (5, 35),
    },
    ("jd-backend-senior", "resume-frontend-mid"): {
        "skills": (10, 35),
        "experience": (15, 45),
        "education": (50, 85),
        "leadership": (35, 70),
    },
    ("jd-backend-senior", "resume-recent-grad"): {
        "skills": (15, 45),
        "experience": (0, 25),
        "education": (75, 100),
        "leadership": (0, 25),
    },
    # ---------- jd-platform-eng --------------------------------------------
    ("jd-platform-eng", "resume-backend-senior"): {
        "skills": (35, 65),
        "experience": (45, 75),
        "education": (75, 100),
        "leadership": (50, 85),
    },
    ("jd-platform-eng", "resume-devops-generalist"): {
        "skills": (45, 75),
        "experience": (55, 85),
        "education": (60, 90),
        "leadership": (40, 70),
    },
    ("jd-platform-eng", "resume-junior-fullstack"): {
        "skills": (5, 30),
        "experience": (5, 30),
        "education": (60, 90),
        "leadership": (0, 25),
    },
    ("jd-platform-eng", "resume-frontend-mid"): {
        "skills": (5, 30),
        "experience": (10, 40),
        "education": (50, 85),
        "leadership": (30, 65),
    },
    ("jd-platform-eng", "resume-recent-grad"): {
        "skills": (5, 30),
        "experience": (0, 20),
        "education": (70, 100),
        "leadership": (0, 25),
    },
    # ---------- jd-ml-senior -----------------------------------------------
    ("jd-ml-senior", "resume-backend-senior"): {
        "skills": (15, 40),
        "experience": (15, 45),
        "education": (40, 70),
        "leadership": (45, 75),
    },
    ("jd-ml-senior", "resume-devops-generalist"): {
        "skills": (5, 30),
        "experience": (10, 35),
        "education": (30, 60),
        "leadership": (30, 60),
    },
    ("jd-ml-senior", "resume-junior-fullstack"): {
        "skills": (10, 35),
        "experience": (5, 30),
        "education": (30, 60),
        "leadership": (0, 25),
    },
    ("jd-ml-senior", "resume-frontend-mid"): {
        "skills": (0, 20),
        "experience": (5, 25),
        "education": (30, 60),
        "leadership": (25, 60),
    },
    ("jd-ml-senior", "resume-recent-grad"): {
        "skills": (5, 30),
        "experience": (0, 20),
        "education": (35, 65),
        "leadership": (0, 25),
    },
    # ---------- jd-frontend-mid --------------------------------------------
    ("jd-frontend-mid", "resume-backend-senior"): {
        "skills": (5, 30),
        "experience": (15, 45),
        "education": (75, 100),
        "leadership": (40, 75),
    },
    ("jd-frontend-mid", "resume-devops-generalist"): {
        "skills": (0, 20),
        "experience": (15, 40),
        "education": (50, 80),
        "leadership": (25, 55),
    },
    ("jd-frontend-mid", "resume-junior-fullstack"): {
        "skills": (15, 45),
        "experience": (10, 35),
        "education": (60, 90),
        "leadership": (0, 25),
    },
    ("jd-frontend-mid", "resume-frontend-mid"): {
        "skills": (75, 100),
        "experience": (65, 95),
        "education": (60, 90),
        "leadership": (60, 90),
    },
    ("jd-frontend-mid", "resume-recent-grad"): {
        "skills": (5, 30),
        "experience": (0, 20),
        "education": (75, 100),
        "leadership": (0, 25),
    },
    # ---------- jd-data-eng-junior -----------------------------------------
    ("jd-data-eng-junior", "resume-backend-senior"): {
        "skills": (55, 85),
        "experience": (60, 95),
        "education": (80, 100),
        "leadership": (50, 85),
    },
    ("jd-data-eng-junior", "resume-devops-generalist"): {
        "skills": (35, 65),
        "experience": (45, 75),
        "education": (55, 85),
        "leadership": (30, 65),
    },
    ("jd-data-eng-junior", "resume-junior-fullstack"): {
        "skills": (40, 70),
        "experience": (30, 65),
        "education": (75, 100),
        "leadership": (0, 30),
    },
    ("jd-data-eng-junior", "resume-frontend-mid"): {
        "skills": (10, 35),
        "experience": (20, 50),
        "education": (50, 85),
        "leadership": (30, 65),
    },
    ("jd-data-eng-junior", "resume-recent-grad"): {
        "skills": (45, 75),
        "experience": (15, 45),
        "education": (75, 100),
        "leadership": (0, 25),
    },
}


def write_files() -> None:
    (FIXTURES / "jds").mkdir(parents=True, exist_ok=True)
    (FIXTURES / "resumes").mkdir(parents=True, exist_ok=True)
    EXPECTED_DIR.mkdir(parents=True, exist_ok=True)

    for slug, body in JDS.items():
        (FIXTURES / "jds" / f"{slug}.txt").write_text(body, encoding="utf-8")
    for slug, body in RESUMES.items():
        (FIXTURES / "resumes" / f"{slug}.txt").write_text(body, encoding="utf-8")
    for (jd, resume), cell in EXPECTED.items():
        payload = {
            "jd": jd,
            "resume": resume,
            "section_ranges": {
                section: {"min": lo, "max": hi} for section, (lo, hi) in cell.items()
            },
            "min_requirements_per_section": {
                "skills": 1,
                "experience": 1,
                "education": 1,
                "leadership": 1,
            },
        }
        (EXPECTED_DIR / f"{jd}__{resume}.json").write_text(
            json.dumps(payload, indent=2), encoding="utf-8"
        )

    n_jd = len(list((FIXTURES / "jds").glob("*.txt")))
    n_resume = len(list((FIXTURES / "resumes").glob("*.txt")))
    n_expected = len(list(EXPECTED_DIR.glob("*.json")))
    print(f"wrote {n_jd} JDs, {n_resume} resumes, {n_expected} expected files")


if __name__ == "__main__":
    write_files()
