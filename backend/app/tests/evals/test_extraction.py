"""Structural-validity eval: JD extraction returns reasonable requirements per section.

Runs against the live LLM service (stub mode or real, depending on env).
Marked `eval` so it's excluded from default pytest runs.
"""

from __future__ import annotations

import pytest

from app.domain.tech import TechDomain
from app.services import llm_service


@pytest.mark.eval
@pytest.mark.parametrize(
    "jd_slug",
    [
        "jd-backend-senior",
        "jd-platform-eng",
        "jd-ml-senior",
        "jd-frontend-mid",
        "jd-data-eng-junior",
    ],
)
async def test_jd_extraction_structural(
    jd_slug: str,
    jd_texts: dict[str, str],
) -> None:
    """Each JD should produce >= 6 requirements spanning all four sections."""
    domain = TechDomain()
    requirements = await llm_service.extract_jd_requirements(jd_texts[jd_slug], domain)

    assert len(requirements) >= 6, f"{jd_slug}: too few requirements ({len(requirements)})"

    sections_seen = {r.section for r in requirements}
    required = {"skills", "experience", "education", "leadership"}
    missing = required - sections_seen
    assert not missing, f"{jd_slug}: missing sections {missing}"

    # Evidence quotes must be verbatim from the JD (substring check).
    jd_text = jd_texts[jd_slug]
    for r in requirements:
        assert r.jd_evidence.text in jd_text, (
            f"{jd_slug}: non-verbatim evidence quote: {r.jd_evidence.text!r}"
        )
