"""End-to-end eval: full matching pipeline for each (JD, resume) pair.

For each of the 25 pairs, runs JD extraction → resume claims → scoring, then
asserts each section score lands inside its hand-tuned expected range.

In stub mode the stub is deterministic-by-hash, so this trivially passes for
whichever flavor the hash picks. Real LLM mode is where this test gets its
signal.
"""

from __future__ import annotations

import pytest

from app.domain.tech import TechDomain
from app.services import llm_service


@pytest.mark.eval
async def test_pipeline_scores_in_range(
    expected_pairs: list[tuple[str, str, dict[str, dict[str, int]]]],
    jd_texts: dict[str, str],
    resume_texts: dict[str, str],
) -> None:
    """At least 80% of (pair, section) cells should hit the expected range."""
    domain = TechDomain()
    total_cells = 0
    hits = 0

    for jd_slug, resume_slug, section_ranges in expected_pairs:
        jd_text = jd_texts[jd_slug]
        resume_text = resume_texts[resume_slug]

        requirements = await llm_service.extract_jd_requirements(jd_text, domain)
        claims = await llm_service.extract_resume_claims(resume_text, domain)
        sections = await llm_service.score_requirements_against_claims(
            requirements,
            claims,
            domain,
            jd_text=jd_text,
            resume_text=resume_text,
        )

        # `s.id` is a Literal section name; `section_ranges` keys come from JSON
        # so they're plain str. Compare via str → str for typing.
        score_by_id: dict[str, int] = {str(s.id): s.score for s in sections}
        for section_id, bounds in section_ranges.items():
            total_cells += 1
            score = score_by_id.get(section_id)
            if score is None:
                continue
            if bounds["min"] <= score <= bounds["max"]:
                hits += 1

    hit_rate = hits / total_cells if total_cells else 0.0
    assert hit_rate >= 0.80, (
        f"only {hits}/{total_cells} = {hit_rate:.0%} section scores in expected range; "
        "tune prompts or expected ranges"
    )
