import pytest
from app.schemas.analysis import (
    MatchVerdict,
    RequirementMatch,
    SectionScore,
    SectionVerdict,
)
from app.schemas.core import Evidence
from app.schemas.extraction import ExtractedRequirement, ResumeClaim
from pydantic import ValidationError

TECH_SECTIONS = {
    "skills",
    "experience",
    "education",
    "project",
}


class TestMatchVerdict:
    def test_accepts_valid_verdict(self) -> None:
        verdict = MatchVerdict(
            id="req-1",
            match_strength="strong",
            supporting_claim_ids=["claim-1", "claim-2"],
            note="Direct production evidence is present.",
        )

        assert verdict.id == "req-1"
        assert verdict.match_strength == "strong"
        assert verdict.supporting_claim_ids == ["claim-1", "claim-2"]
        assert verdict.note == "Direct production evidence is present."

    def test_defaults_supporting_claim_ids_to_empty_list(self) -> None:
        verdict = MatchVerdict(
            id="req-1",
            match_strength="none",
        )

        assert verdict.supporting_claim_ids == []

    def test_defaults_note_to_none(self) -> None:
        verdict = MatchVerdict(
            id="req-1",
            match_strength="none",
        )

        assert verdict.note is None

    def test_rejects_invalid_match_strength(self) -> None:
        with pytest.raises(ValidationError):
            MatchVerdict(
                id="req-1",
                match_strength="excellent",  # type: ignore[arg-type]
            )

    def test_strips_whitespace_from_string_fields(self) -> None:
        verdict = MatchVerdict(
            id="  req-1  ",
            match_strength="strong",
            supporting_claim_ids=["claim-1"],
            note="  Direct evidence.  ",
        )

        assert verdict.id == "req-1"
        assert verdict.note == "Direct evidence."

    def test_rejects_note_longer_than_maximum(self) -> None:
        with pytest.raises(ValidationError):
            MatchVerdict(
                id="req-1",
                match_strength="weak",
                note="x" * 301,
            )


class TestSectionVerdict:
    def test_accepts_valid_section_verdict(self) -> None:
        verdict = SectionVerdict(
            id="skills",
            score=85,
            review="Strong technical coverage with a few gaps.",
        )

        assert verdict.id == "skills"
        assert verdict.score == 85
        assert verdict.review == "Strong technical coverage with a few gaps."

    def test_accepts_score_boundaries(self) -> None:
        minimum = SectionVerdict(
            id="skills",
            score=0,
            review="No evidence.",
        )
        maximum = SectionVerdict(
            id="skills",
            score=100,
            review="Complete evidence.",
        )

        assert minimum.score == 0
        assert maximum.score == 100

    def test_rejects_negative_score(self) -> None:
        with pytest.raises(ValidationError):
            SectionVerdict(
                id="skills",
                score=-1,
                review="Invalid score.",
            )

    def test_rejects_score_above_100(self) -> None:
        with pytest.raises(ValidationError):
            SectionVerdict(
                id="skills",
                score=101,
                review="Invalid score.",
            )

    def test_rejects_review_longer_than_maximum(self) -> None:
        with pytest.raises(ValidationError):
            SectionVerdict(
                id="skills",
                score=80,
                review="x" * 401,
            )

    def test_taxonomy_is_not_checked_without_context(self) -> None:
        verdict = SectionVerdict(
            id="not-a-real-section",  # type: ignore[arg-type]
            score=50,
            review="Accepted without taxonomy context.",
        )

        assert verdict.id == "not-a-real-section"

    def test_accepts_section_in_supplied_taxonomy(self) -> None:
        verdict = SectionVerdict.model_validate(
            {
                "id": "experience",
                "score": 90,
                "review": "Strong relevant experience.",
            },
            context={
                "valid_sections": TECH_SECTIONS,
            },
        )

        assert verdict.id == "experience"

    def test_rejects_section_outside_supplied_taxonomy(self) -> None:
        with pytest.raises(
            ValidationError,
            match="section id 'not-a-real-section' is not in this domain's section taxonomy",
        ):
            SectionVerdict.model_validate(
                {
                    "id": "not-a-real-section",
                    "score": 50,
                    "review": "Invalid section.",
                },
                context={
                    "valid_sections": TECH_SECTIONS,
                },
            )


class TestRequirementMatch:
    def test_accepts_valid_requirement_match(
        self,
        sample_requirement: ExtractedRequirement,
        sample_claim: ResumeClaim,
    ) -> None:
        match = RequirementMatch(
            id="req-1",
            requirement="Strong FastAPI and PostgreSQL experience",
            jd_evidence=sample_requirement.jd_evidence,
            match_strength="strong",
            resume_evidence=[sample_claim.resume_evidence],
            note=None,
        )

        assert match.id == "req-1"
        assert match.match_strength == "strong"
        assert len(match.resume_evidence) == 1

    def test_matched_is_true_for_strong(
        self,
        sample_requirement: ExtractedRequirement,
    ) -> None:
        match = RequirementMatch(
            id="req-1",
            requirement=sample_requirement.requirement,
            jd_evidence=sample_requirement.jd_evidence,
            match_strength="strong",
        )

        assert match.matched is True

    def test_matched_is_true_for_partial(
        self,
        sample_requirement: ExtractedRequirement,
    ) -> None:
        match = RequirementMatch(
            id="req-1",
            requirement=sample_requirement.requirement,
            jd_evidence=sample_requirement.jd_evidence,
            match_strength="partial",
        )

        assert match.matched is True

    @pytest.mark.parametrize("match_strength", ["weak", "none"])
    def test_matched_is_false_for_weak_and_none(
        self,
        sample_requirement: ExtractedRequirement,
        match_strength: str,
    ) -> None:
        match = RequirementMatch(
            id="req-1",
            requirement=sample_requirement.requirement,
            jd_evidence=sample_requirement.jd_evidence,
            match_strength=match_strength,  # type: ignore[arg-type]
        )

        assert match.matched is False

    def test_rejects_jd_evidence_from_resume(
        self,
        sample_claim: ResumeClaim,
    ) -> None:
        with pytest.raises(
            ValidationError,
            match="jd_evidence.source must be 'jd'",
        ):
            RequirementMatch(
                id="req-1",
                requirement="Python experience",
                jd_evidence=sample_claim.resume_evidence,
                match_strength="strong",
            )

    def test_rejects_resume_evidence_from_jd(
        self,
        sample_requirement: ExtractedRequirement,
    ) -> None:
        with pytest.raises(
            ValidationError,
            match="resume_evidence entries must have source='resume'",
        ):
            RequirementMatch(
                id="req-1",
                requirement=sample_requirement.requirement,
                jd_evidence=sample_requirement.jd_evidence,
                match_strength="strong",
                resume_evidence=[sample_requirement.jd_evidence],
            )

    def test_rejects_note_longer_than_maximum(
        self,
        sample_requirement: ExtractedRequirement,
    ) -> None:
        with pytest.raises(ValidationError):
            RequirementMatch(
                id="req-1",
                requirement=sample_requirement.requirement,
                jd_evidence=sample_requirement.jd_evidence,
                match_strength="none",
                note="x" * 301,
            )


class TestRequirementMatchFromVerdict:
    def test_builds_requirement_match_from_verdict(
        self,
        sample_requirement: ExtractedRequirement,
        sample_claims: list[ResumeClaim],
    ) -> None:
        verdict = MatchVerdict(
            id="req-1",
            match_strength="strong",
            supporting_claim_ids=["claim-1"],
            note="Direct evidence is present.",
        )

        claims_by_id = {
            claim.id: claim
            for claim in sample_claims
            if claim.id is not None
        }

        match = RequirementMatch.from_verdict(
            sample_requirement,
            verdict,
            claims_by_id,
        )

        assert match.id == "req-1"
        assert match.requirement == sample_requirement.requirement
        assert match.jd_evidence == sample_requirement.jd_evidence
        assert match.match_strength == "strong"
        assert match.note == "Direct evidence is present."
        assert len(match.resume_evidence) == 1
        assert match.resume_evidence[0] == sample_claims[0].resume_evidence

    def test_resolves_multiple_supporting_claims(
        self,
        sample_requirement: ExtractedRequirement,
        sample_claims: list[ResumeClaim],
    ) -> None:
        verdict = MatchVerdict(
            id="req-1",
            match_strength="strong",
            supporting_claim_ids=["claim-1", "claim-2"],
        )

        claims_by_id = {
            claim.id: claim
            for claim in sample_claims
            if claim.id is not None
        }

        match = RequirementMatch.from_verdict(
            sample_requirement,
            verdict,
            claims_by_id,
        )

        assert len(match.resume_evidence) == 2
        assert match.resume_evidence[0] == sample_claims[0].resume_evidence
        assert match.resume_evidence[1] == sample_claims[1].resume_evidence

    def test_allows_no_supporting_claims(
        self,
        sample_requirement: ExtractedRequirement,
    ) -> None:
        verdict = MatchVerdict(
            id="req-1",
            match_strength="none",
            supporting_claim_ids=[],
            note="The resume does not address this requirement.",
        )

        match = RequirementMatch.from_verdict(
            sample_requirement,
            verdict,
            {},
        )

        assert match.resume_evidence == []
        assert match.match_strength == "none"
        assert match.matched is False

    def test_rejects_missing_requirement_id(self) -> None:
        requirement = ExtractedRequirement(
            section="skills",
            requirement="Strong Python experience",
            jd_evidence=Evidence(
                text="Strong Python experience",
                source="jd",
            ),
        )

        verdict = MatchVerdict(
            id="req-1",
            match_strength="strong",
        )

        with pytest.raises(
            ValueError,
            match="requirement.id must be assigned before matching",
        ):
            RequirementMatch.from_verdict(
                requirement,
                verdict,
                {},
            )

    def test_rejects_mismatched_verdict_id(
        self,
        sample_requirement: ExtractedRequirement,
    ) -> None:
        verdict = MatchVerdict(
            id="req-999",
            match_strength="strong",
        )

        with pytest.raises(
            ValueError,
            match="does not match requirement.id",
        ):
            RequirementMatch.from_verdict(
                sample_requirement,
                verdict,
                {},
            )

    def test_rejects_unknown_supporting_claim_id(
        self,
        sample_requirement: ExtractedRequirement,
    ) -> None:
        verdict = MatchVerdict(
            id="req-1",
            match_strength="strong",
            supporting_claim_ids=["claim-999"],
        )

        with pytest.raises(
            ValueError,
            match="references unknown resume claim id 'claim-999'",
        ):
            RequirementMatch.from_verdict(
                sample_requirement,
                verdict,
                {},
            )


class TestSectionScore:
    def test_accepts_valid_section_score(self) -> None:
        section = SectionScore(
            id="skills",
            label="Skills",
            score=85,
            review="Strong technical coverage.",
            requirements=[],
        )

        assert section.id == "skills"
        assert section.label == "Skills"
        assert section.score == 85
        assert section.review == "Strong technical coverage."
        assert section.requirements == []

    def test_accepts_score_boundaries(self) -> None:
        minimum = SectionScore(
            id="skills",
            label="Skills",
            score=0,
            review="No evidence.",
        )
        maximum = SectionScore(
            id="skills",
            label="Skills",
            score=100,
            review="Complete evidence.",
        )

        assert minimum.score == 0
        assert maximum.score == 100

    def test_rejects_negative_score(self) -> None:
        with pytest.raises(ValidationError):
            SectionScore(
                id="skills",
                label="Skills",
                score=-1,
                review="Invalid.",
            )

    def test_rejects_score_above_100(self) -> None:
        with pytest.raises(ValidationError):
            SectionScore(
                id="skills",
                label="Skills",
                score=101,
                review="Invalid.",
            )

    def test_rejects_review_longer_than_maximum(self) -> None:
        with pytest.raises(ValidationError):
            SectionScore(
                id="skills",
                label="Skills",
                score=80,
                review="x" * 401,
            )

    def test_accepts_requirements(self, sample_requirement, sample_claim) -> None:
        requirement = RequirementMatch(
            id="req-1",
            requirement=sample_requirement.requirement,
            jd_evidence=sample_requirement.jd_evidence,
            match_strength="strong",
            resume_evidence=[sample_claim.resume_evidence],
        )

        section = SectionScore(
            id="skills",
            label="Skills",
            score=90,
            review="Strong coverage.",
            requirements=[requirement],
        )

        assert len(section.requirements) == 1
        assert section.requirements[0].id == "req-1"

    def test_taxonomy_is_not_checked_without_context(self) -> None:
        section = SectionScore(
            id="not-a-real-section",  # type: ignore[arg-type]
            label="Unknown",
            score=50,
            review="Accepted without taxonomy context.",
        )

        assert section.id == "not-a-real-section"

    def test_accepts_section_in_supplied_taxonomy(self) -> None:
        section = SectionScore.model_validate(
            {
                "id": "project",
                "label": "Project",
                "score": 75,
                "review": "Good project evidence.",
                "requirements": [],
            },
            context={
                "valid_sections": TECH_SECTIONS,
            },
        )

        assert section.id == "project"

    def test_rejects_section_outside_supplied_taxonomy(self) -> None:
        with pytest.raises(
            ValidationError,
            match="section id 'not-a-real-section' is not in this domain's section taxonomy",
        ):
            SectionScore.model_validate(
                {
                    "id": "not-a-real-section",
                    "label": "Unknown",
                    "score": 50,
                    "review": "Invalid section.",
                    "requirements": [],
                },
                context={
                    "valid_sections": TECH_SECTIONS,
                },
            )
