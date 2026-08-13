from typing import cast

import app.services.llm.service as llm_service
import pytest
from app.domain.tech import DynamicFinalAnalysisResult, TechScoringResult
from app.schemas import MatchVerdict, SectionScore, SectionVerdict
from pydantic import BaseModel, ConfigDict


def build_scoring_result(
    *,
    match_verdicts: list[MatchVerdict] | None = None,
    section_verdicts: list[SectionVerdict] | None = None,
    complete_analysis: str = "Strong overall fit for the role.",
) -> TechScoringResult:
    if match_verdicts is None:
        match_verdicts = [
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
                match_strength="partial",
                supporting_claim_ids=["claim-3"],
                note="The evidence covers part of the requirement.",
            ),
        ]

    if section_verdicts is None:
        section_verdicts = [
            SectionVerdict(
                id="skills",
                score=90,
                review="Strong technical skill coverage.",
            ),
            SectionVerdict(
                id="experience",
                score=95,
                review="Relevant production experience is well demonstrated.",
            ),
            SectionVerdict(
                id="education",
                score=90,
                review="The requested education is present.",
            ),
            SectionVerdict(
                id="project",
                score=40,
                review="Project evidence is limited.",
            ),
        ]

    return TechScoringResult(
        complete_analysis=complete_analysis,
        match_verdicts=match_verdicts,
        section_verdicts=section_verdicts,
    )


def assemble_tech_result(
    tech_domain,
    jd_response,
    resume_response,
    scoring_result,
) -> DynamicFinalAnalysisResult:
    """Narrow the generic BaseModel return to the concrete Tech result."""
    return cast(
        DynamicFinalAnalysisResult,
        llm_service.assemble_final_result(
            tech_domain,
            jd_response,
            resume_response,
            scoring_result,
        ),
    )


class TestAssembleFinalResult:
    def test_builds_final_result_from_scoring_output(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
    ) -> None:
        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )
        scoring_result = build_scoring_result()

        result = assemble_tech_result(
            tech_domain,
            jd_response,
            resume_response,
            scoring_result,
        )

        assert result.complete_analysis == "Strong overall fit for the role."
        assert len(result.sections) == 4

        sections_by_id = {
            section.id: section
            for section in result.sections
        }

        assert set(sections_by_id) == {
            "skills",
            "experience",
            "education",
            "project",
        }

    def test_creates_requirement_matches_with_correct_evidence(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
    ) -> None:
        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        result = assemble_tech_result(
            tech_domain,
            jd_response,
            resume_response,
            build_scoring_result(),
        )

        all_matches = [
            match
            for section in result.sections
            for match in section.requirements
        ]

        assert len(all_matches) == 3

        matches_by_id = {
            match.id: match
            for match in all_matches
        }

        assert matches_by_id["req-1"].jd_evidence == (
            sample_requirements[0].jd_evidence
        )
        assert matches_by_id["req-1"].resume_evidence == [
            sample_claims[0].resume_evidence
        ]

        assert matches_by_id["req-2"].jd_evidence == (
            sample_requirements[1].jd_evidence
        )
        assert matches_by_id["req-2"].resume_evidence == [
            sample_claims[1].resume_evidence
        ]

        assert matches_by_id["req-3"].jd_evidence == (
            sample_requirements[2].jd_evidence
        )
        assert matches_by_id["req-3"].resume_evidence == [
            sample_claims[2].resume_evidence
        ]

    def test_places_requirement_matches_under_original_requirement_section(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
    ) -> None:
        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        result = assemble_tech_result(
            tech_domain,
            jd_response,
            resume_response,
            build_scoring_result(),
        )

        sections_by_id = {
            section.id: section
            for section in result.sections
        }

        assert [m.id for m in sections_by_id["skills"].requirements] == [
            "req-1"
        ]
        assert [m.id for m in sections_by_id["experience"].requirements] == [
            "req-2"
        ]
        assert [m.id for m in sections_by_id["education"].requirements] == [
            "req-3"
        ]
        assert sections_by_id["project"].requirements == []

    def test_copies_section_scores_and_reviews(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
    ) -> None:
        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        result = assemble_tech_result(
            tech_domain,
            jd_response,
            resume_response,
            build_scoring_result(),
        )

        sections_by_id = {
            section.id: section
            for section in result.sections
        }

        assert sections_by_id["skills"].score == 90
        assert sections_by_id["skills"].review == (
            "Strong technical skill coverage."
        )

        assert sections_by_id["experience"].score == 95
        assert sections_by_id["education"].score == 90
        assert sections_by_id["project"].score == 40

    def test_builds_human_readable_section_labels(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
    ) -> None:
        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        result = assemble_tech_result(
            tech_domain,
            jd_response,
            resume_response,
            build_scoring_result(),
        )

        labels = {
            section.id: section.label
            for section in result.sections
        }

        assert labels == {
            "skills": "Skills",
            "experience": "Experience",
            "education": "Education",
            "project": "Project",
        }

    def test_rejects_duplicate_requirement_verdict_ids(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
    ) -> None:
        duplicate_verdicts = [
            MatchVerdict(
                id="req-1",
                match_strength="strong",
                supporting_claim_ids=["claim-1"],
            ),
            MatchVerdict(
                id="req-1",
                match_strength="partial",
                supporting_claim_ids=["claim-1"],
            ),
            MatchVerdict(
                id="req-3",
                match_strength="none",
                supporting_claim_ids=[],
            ),
        ]

        scoring_result = build_scoring_result(
            match_verdicts=duplicate_verdicts,
        )

        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        with pytest.raises(
            llm_service.AnalysisEngineError,
            match="Duplicate requirement IDs found in match verdicts",
        ) as exc_info:
            llm_service.assemble_final_result(
                tech_domain,
                jd_response,
                resume_response,
                scoring_result,
            )

        assert exc_info.value.code == "ERR_ASSEMBLY_MISMATCH"
        assert exc_info.value.retryable is False

    def test_rejects_missing_requirement_verdict(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
    ) -> None:
        missing_verdicts = [
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
        ]

        scoring_result = build_scoring_result(
            match_verdicts=missing_verdicts,
        )

        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        with pytest.raises(
            llm_service.AnalysisEngineError,
            match="Match verdicts must cover exactly",
        ) as exc_info:
            llm_service.assemble_final_result(
                tech_domain,
                jd_response,
                resume_response,
                scoring_result,
            )

        assert exc_info.value.code == "ERR_ASSEMBLY_MISMATCH"

    def test_rejects_unexpected_requirement_verdict_id(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
    ) -> None:
        unexpected_verdicts = [
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
                id="req-999",
                match_strength="strong",
                supporting_claim_ids=["claim-3"],
            ),
        ]

        scoring_result = build_scoring_result(
            match_verdicts=unexpected_verdicts,
        )

        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        with pytest.raises(
            llm_service.AnalysisEngineError,
            match="Match verdicts must cover exactly",
        ) as exc_info:
            llm_service.assemble_final_result(
                tech_domain,
                jd_response,
                resume_response,
                scoring_result,
            )

        assert exc_info.value.code == "ERR_ASSEMBLY_MISMATCH"

    def test_rejects_unassigned_requirement_id(
        self,
        tech_domain,
        sample_claims,
    ) -> None:
        from app.schemas import Evidence, ExtractedRequirement

        requirement_without_id = ExtractedRequirement(
            section="skills",
            requirement="Python experience",
            jd_evidence=Evidence(
                text="Python experience",
                source="jd",
            ),
        )

        jd_response = llm_service.JDRequirementsResponse(
            requirements=[requirement_without_id],
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        scoring_result = build_scoring_result(
            match_verdicts=[
                MatchVerdict(
                    id="req-1",
                    match_strength="strong",
                    supporting_claim_ids=["claim-1"],
                )
            ],
        )

        with pytest.raises(
            ValueError,
            match="requirement id was never assigned",
        ):
            llm_service.assemble_final_result(
                tech_domain,
                jd_response,
                resume_response,
                scoring_result,
            )

    def test_uses_domain_final_schema(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
    ) -> None:
        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        result = assemble_tech_result(
            tech_domain,
            jd_response,
            resume_response,
            build_scoring_result(),
        )

        assert isinstance(
            result,
            tech_domain.get_final_schema(),
        )

    def test_final_result_serializes_with_expected_aliases(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
    ) -> None:
        jd_response = llm_service.JDRequirementsResponse(
            requirements=sample_requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        result = assemble_tech_result(
            tech_domain,
            jd_response,
            resume_response,
            build_scoring_result(),
        )

        payload = result.model_dump(
            by_alias=True,
            exclude_none=True,
        )

        assert "completeAnalysis" in payload
        assert "sections" in payload

        first_requirement = next(
            match
            for section in payload["sections"]
            for match in section["requirements"]
            if match["id"] == "req-1"
        )

        assert "jdEvidence" in first_requirement
        assert "resumeEvidence" in first_requirement
        assert first_requirement["matched"] is True

    def test_missing_section_verdict_gets_zero_score_fallback(
        self,
        tech_domain,
        sample_requirements,
        sample_claims,
    ) -> None:
        from app.schemas import Evidence, ExtractedRequirement

        project_requirement = ExtractedRequirement(
            id="req-4",
            section="project",
            requirement="Provide architecture design experience",
            jd_evidence=Evidence(
                text="Architecture design experience is required.",
                source="jd",
            ),
        )

        requirements = [
            *sample_requirements,
            project_requirement,
        ]

        jd_response = llm_service.JDRequirementsResponse(
            requirements=requirements,
        )
        resume_response = llm_service.ResumeClaimsResponse(
            claims=sample_claims,
        )

        section_verdicts = [
            SectionVerdict(
                id="skills",
                score=90,
                review="Strong technical skill coverage.",
            ),
            SectionVerdict(
                id="experience",
                score=95,
                review="Relevant experience.",
            ),
            SectionVerdict(
                id="education",
                score=90,
                review="Education is aligned.",
            ),
        ]

        class PartialScoringResult(BaseModel):
            model_config = ConfigDict(str_strip_whitespace=True)

            complete_analysis: str
            match_verdicts: list[MatchVerdict]
            section_verdicts: list[SectionVerdict]

        scoring_result = PartialScoringResult(
            complete_analysis="Partial section evaluation.",
            match_verdicts=[
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
                    match_strength="partial",
                    supporting_claim_ids=["claim-3"],
                ),
                MatchVerdict(
                    id="req-4",
                    match_strength="none",
                    supporting_claim_ids=[],
                ),
            ],
            section_verdicts=section_verdicts,
        )

        result = assemble_tech_result(
            tech_domain,
            jd_response,
            resume_response,
            scoring_result,
        )

        project_section = next(
            section
            for section in result.sections
            if section.id == "project"
        )

        assert isinstance(project_section, SectionScore)
        assert project_section.score == 0
        assert (
            project_section.review
            == "Section evaluation was omitted by scoring model."
        )
        assert [match.id for match in project_section.requirements] == ["req-4"]
