import pytest
from app.domain.tech import (
    JD_EXTRACTION_PROMPT,
    PROMPT_VERSION,
    RESUME_EXTRACTION_PROMPT,
    SCORING_PROMPT,
    VALID_SECTIONS,
    DynamicFinalAnalysisResult,
    TechDomain,
    TechScoringResult,
)
from app.schemas import MatchVerdict, SectionScore, SectionVerdict


class TestTechDomain:
    def test_name(self, tech_domain: TechDomain) -> None:
        assert tech_domain.name == "tech"

    def test_prompt_version(self, tech_domain: TechDomain) -> None:
        assert tech_domain.prompt_version == PROMPT_VERSION
        assert tech_domain.prompt_version == "tech-v4"

    def test_jd_extraction_prompt(self, tech_domain: TechDomain) -> None:
        assert tech_domain.jd_extraction_prompt() == JD_EXTRACTION_PROMPT
        assert "software-engineering job description" in JD_EXTRACTION_PROMPT

    def test_resume_extraction_prompt(self, tech_domain: TechDomain) -> None:
        assert tech_domain.resume_extraction_prompt() == RESUME_EXTRACTION_PROMPT
        assert "candidate's resume" in RESUME_EXTRACTION_PROMPT

    def test_scoring_prompt(self, tech_domain: TechDomain) -> None:
        assert tech_domain.scoring_prompt() == SCORING_PROMPT
        assert "strong / partial / weak / none" in SCORING_PROMPT

    def test_section_taxonomy(self, tech_domain: TechDomain) -> None:
        assert tech_domain.section_taxonomy() == [
            "skills",
            "experience",
            "education",
            "project",
        ]

    def test_valid_sections_matches_taxonomy(self, tech_domain: TechDomain) -> None:
        assert set(tech_domain.section_taxonomy()) == VALID_SECTIONS

    def test_section_weights(self, tech_domain: TechDomain) -> None:
        assert tech_domain.section_weights() == {
            "skills": 0.40,
            "experience": 0.40,
            "project": 0.15,
            "education": 0.05,
        }

    def test_section_weights_sum_to_one(self, tech_domain: TechDomain) -> None:
        assert sum(tech_domain.section_weights().values()) == pytest.approx(1.0)

    def test_section_weights_cover_exact_taxonomy(
        self,
        tech_domain: TechDomain,
    ) -> None:
        assert set(tech_domain.section_weights()) == set(
            tech_domain.section_taxonomy()
        )

    def test_section_weights_are_positive(self, tech_domain: TechDomain) -> None:
        assert all(weight > 0 for weight in tech_domain.section_weights().values())

    def test_get_scoring_schema(self, tech_domain: TechDomain) -> None:
        assert tech_domain.get_scoring_schema() is TechScoringResult

    def test_get_final_schema(self, tech_domain: TechDomain) -> None:
        assert tech_domain.get_final_schema() is DynamicFinalAnalysisResult

    def test_constructor_validates_section_weights(self) -> None:
        domain = TechDomain()

        domain.validate_section_weights()


class TestTechScoringResult:
    def test_accepts_complete_scoring_result(self) -> None:
        result = TechScoringResult(
            complete_analysis="Strong overall fit for the role.",
            match_verdicts=[
                MatchVerdict(
                    id="req-1",
                    match_strength="strong",
                    supporting_claim_ids=["claim-1"],
                )
            ],
            section_verdicts=[
                SectionVerdict(
                    id="skills",
                    score=90,
                    review="Strong technical coverage.",
                ),
                SectionVerdict(
                    id="experience",
                    score=85,
                    review="Relevant engineering experience.",
                ),
                SectionVerdict(
                    id="education",
                    score=70,
                    review="Education is reasonably aligned.",
                ),
                SectionVerdict(
                    id="project",
                    score=60,
                    review="Some project evidence is present.",
                ),
            ],
        )

        assert result.complete_analysis == "Strong overall fit for the role."
        assert len(result.match_verdicts) == 1
        assert len(result.section_verdicts) == 4

    def test_optional_skill_lists_default_to_empty(self) -> None:
        result = TechScoringResult(
            complete_analysis="Overall fit is moderate.",
            match_verdicts=[],
            section_verdicts=[
                SectionVerdict(
                    id="skills",
                    score=0,
                    review="No evidence.",
                ),
                SectionVerdict(
                    id="experience",
                    score=0,
                    review="No evidence.",
                ),
                SectionVerdict(
                    id="education",
                    score=0,
                    review="No evidence.",
                ),
                SectionVerdict(
                    id="project",
                    score=0,
                    review="No evidence.",
                ),
            ],
        )

        assert result.matched_skills == []
        assert result.missing_critical_skills == []

    def test_accepts_skill_lists(self) -> None:
        result = TechScoringResult(
            complete_analysis="Good fit.",
            match_verdicts=[],
            section_verdicts=[
                SectionVerdict(
                    id=section,
                    score=50,
                    review="Mixed evidence.",
                )
                for section in VALID_SECTIONS
            ],
            matched_skills=["Python", "FastAPI"],
            missing_critical_skills=["Kubernetes"],
        )

        assert result.matched_skills == ["Python", "FastAPI"]
        assert result.missing_critical_skills == ["Kubernetes"]

    def test_accepts_exactly_all_four_sections(self) -> None:
        result = TechScoringResult(
            complete_analysis="Complete section coverage.",
            match_verdicts=[],
            section_verdicts=[
                SectionVerdict(
                    id="skills",
                    score=50,
                    review="Mixed.",
                ),
                SectionVerdict(
                    id="experience",
                    score=50,
                    review="Mixed.",
                ),
                SectionVerdict(
                    id="education",
                    score=50,
                    review="Mixed.",
                ),
                SectionVerdict(
                    id="project",
                    score=50,
                    review="Mixed.",
                ),
            ],
        )

        assert {section.id for section in result.section_verdicts} == VALID_SECTIONS

    def test_rejects_duplicate_section_ids(self) -> None:
        section_verdicts = [
            SectionVerdict(
                id="skills",
                score=80,
                review="Good.",
            ),
            SectionVerdict(
                id="skills",
                score=70,
                review="Duplicate.",
            ),
            SectionVerdict(
                id="education",
                score=60,
                review="Okay.",
            ),
            SectionVerdict(
                id="project",
                score=50,
                review="Okay.",
            ),
        ]

        with pytest.raises(ValueError, match="duplicate section ids"):
            TechScoringResult(
                complete_analysis="Invalid result.",
                match_verdicts=[],
                section_verdicts=section_verdicts,
            )

    def test_rejects_missing_section(self) -> None:
        section_verdicts = [
            SectionVerdict(
                id="skills",
                score=80,
                review="Good.",
            ),
            SectionVerdict(
                id="experience",
                score=70,
                review="Good.",
            ),
            SectionVerdict(
                id="education",
                score=60,
                review="Okay.",
            ),
        ]

        with pytest.raises(
            ValueError,
            match="section_verdicts must cover exactly",
        ):
            TechScoringResult(
                complete_analysis="Incomplete result.",
                match_verdicts=[],
                section_verdicts=section_verdicts,
            )

    def test_rejects_unexpected_section(self) -> None:
        section_verdicts = [
            SectionVerdict(
                id="skills",
                score=80,
                review="Good.",
            ),
            SectionVerdict(
                id="experience",
                score=70,
                review="Good.",
            ),
            SectionVerdict(
                id="education",
                score=60,
                review="Okay.",
            ),
            SectionVerdict(
                id="project",
                score=50,
                review="Okay.",
            ),
        ]

        section_verdicts[-1] = SectionVerdict(
            id="other",
            score=50,
            review="Unexpected.",
        )

        with pytest.raises(
            ValueError,
            match="section_verdicts must cover exactly",
        ):
            TechScoringResult(
                complete_analysis="Invalid taxonomy.",
                match_verdicts=[],
                section_verdicts=section_verdicts,
            )


class TestDynamicFinalAnalysisResult:
    def test_accepts_complete_final_result(self) -> None:
        result = DynamicFinalAnalysisResult(
            complete_analysis="Strong overall fit.",
            sections=[
                SectionScore(
                    id="skills",
                    label="Skills",
                    score=90,
                    review="Strong.",
                ),
                SectionScore(
                    id="experience",
                    label="Experience",
                    score=85,
                    review="Strong.",
                ),
                SectionScore(
                    id="education",
                    label="Education",
                    score=70,
                    review="Good.",
                ),
                SectionScore(
                    id="project",
                    label="Project",
                    score=60,
                    review="Moderate.",
                ),
            ],
        )

        assert result.complete_analysis == "Strong overall fit."
        assert {section.id for section in result.sections} == VALID_SECTIONS

    def test_serializes_complete_analysis_using_camel_case_alias(self) -> None:
        result = DynamicFinalAnalysisResult(
            complete_analysis="Strong overall fit.",
            sections=[
                SectionScore(
                    id=section,
                    label=section.title(),
                    score=50,
                    review="Mixed.",
                )
                for section in VALID_SECTIONS
            ],
        )

        data = result.model_dump(by_alias=True)

        assert "completeAnalysis" in data
        assert "complete_analysis" not in data

    def test_rejects_duplicate_section_ids(self) -> None:
        sections = [
            SectionScore(
                id="skills",
                label="Skills",
                score=80,
                review="Good.",
            ),
            SectionScore(
                id="skills",
                label="Skills duplicate",
                score=70,
                review="Duplicate.",
            ),
            SectionScore(
                id="education",
                label="Education",
                score=60,
                review="Okay.",
            ),
            SectionScore(
                id="project",
                label="Project",
                score=50,
                review="Okay.",
            ),
        ]

        with pytest.raises(ValueError, match="duplicate section ids"):
            DynamicFinalAnalysisResult(
                complete_analysis="Invalid result.",
                sections=sections,
            )

    def test_rejects_missing_section(self) -> None:
        sections = [
            SectionScore(
                id="skills",
                label="Skills",
                score=80,
                review="Good.",
            ),
            SectionScore(
                id="experience",
                label="Experience",
                score=70,
                review="Good.",
            ),
            SectionScore(
                id="education",
                label="Education",
                score=60,
                review="Okay.",
            ),
        ]

        with pytest.raises(
            ValueError,
            match="sections must cover exactly",
        ):
            DynamicFinalAnalysisResult(
                complete_analysis="Incomplete result.",
                sections=sections,
            )

    def test_rejects_unexpected_section(self) -> None:
        sections = [
            SectionScore(
                id="skills",
                label="Skills",
                score=80,
                review="Good.",
            ),
            SectionScore(
                id="experience",
                label="Experience",
                score=70,
                review="Good.",
            ),
            SectionScore(
                id="education",
                label="Education",
                score=60,
                review="Okay.",
            ),
            SectionScore(
                id="other",
                label="Other",
                score=50,
                review="Unexpected.",
            ),
        ]

        with pytest.raises(
            ValueError,
            match="sections must cover exactly",
        ):
            DynamicFinalAnalysisResult(
                complete_analysis="Invalid taxonomy.",
                sections=sections,
            )
