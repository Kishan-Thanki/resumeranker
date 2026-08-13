import pytest
from app.schemas.core import Evidence
from app.schemas.extraction import ExtractedRequirement, ResumeClaim
from pydantic import ValidationError


class TestExtractedRequirement:
    def test_accepts_valid_requirement(self, jd_evidence) -> None:
        requirement = ExtractedRequirement(
            id="req-1",
            section="skills",
            requirement="Strong Python experience",
            jd_evidence=jd_evidence,
        )

        assert requirement.id == "req-1"
        assert requirement.section == "skills"
        assert requirement.requirement == "Strong Python experience"
        assert requirement.jd_evidence.source == "jd"

    def test_id_defaults_to_none(self, jd_evidence) -> None:
        requirement = ExtractedRequirement(
            section="skills",
            requirement="Strong Python experience",
            jd_evidence=jd_evidence,
        )

        assert requirement.id is None

    def test_strips_whitespace_from_fields(self) -> None:
        requirement = ExtractedRequirement(
            id="  req-1  ",
            section=" skills ",
            requirement="  Strong Python experience  ",
            jd_evidence=Evidence(
                text="  Strong Python experience  ",
                source="jd",
            ),
        )

        assert requirement.id == "req-1"
        assert requirement.section == "skills"
        assert requirement.requirement == "Strong Python experience"
        assert requirement.jd_evidence.text == "Strong Python experience"

    def test_rejects_non_jd_evidence(self, resume_evidence) -> None:
        with pytest.raises(
            ValidationError,
            match="jd_evidence.source must be 'jd'",
        ):
            ExtractedRequirement(
                id="req-1",
                section="skills",
                requirement="Strong Python experience",
                jd_evidence=resume_evidence,
            )

    def test_rejects_requirement_longer_than_maximum(self, jd_evidence) -> None:
        with pytest.raises(ValidationError):
            ExtractedRequirement(
                id="req-1",
                section="skills",
                requirement="x" * 151,
                jd_evidence=jd_evidence,
            )

    def test_accepts_requirement_at_maximum_length(self, jd_evidence) -> None:
        requirement = ExtractedRequirement(
            id="req-1",
            section="skills",
            requirement="x" * 150,
            jd_evidence=jd_evidence,
        )

        assert len(requirement.requirement) == 150

    def test_accepts_known_section_without_validation_context(
        self,
        jd_evidence,
    ) -> None:
        requirement = ExtractedRequirement(
            id="req-1",
            section="not-a-tech-section",
            requirement="Some requirement",
            jd_evidence=jd_evidence,
        )

        assert requirement.section == "not-a-tech-section"

    def test_rejects_section_outside_domain_taxonomy_when_context_is_supplied(
        self,
        jd_evidence,
    ) -> None:
        with pytest.raises(
            ValidationError,
            match="section 'not-a-tech-section' is not in this domain's section taxonomy",
        ):
            ExtractedRequirement.model_validate(
                {
                    "id": "req-1",
                    "section": "not-a-tech-section",
                    "requirement": "Some requirement",
                    "jd_evidence": {
                        "text": "Some requirement",
                        "source": "jd",
                    },
                },
                context={
                    "valid_sections": {
                        "skills",
                        "experience",
                        "education",
                        "project",
                    },
                },
            )

    def test_accepts_section_in_supplied_domain_taxonomy(
        self,
        jd_evidence,
    ) -> None:
        requirement = ExtractedRequirement.model_validate(
            {
                "id": "req-1",
                "section": "experience",
                "requirement": "5+ years of Python",
                "jd_evidence": {
                    "text": jd_evidence.text,
                    "source": "jd",
                },
            },
            context={
                "valid_sections": {
                    "skills",
                    "experience",
                    "education",
                    "project",
                },
            },
        )

        assert requirement.section == "experience"


class TestResumeClaim:
    def test_accepts_valid_claim(self, resume_evidence) -> None:
        claim = ResumeClaim(
            id="claim-1",
            section="experience",
            claim="Six years of Python development",
            resume_evidence=resume_evidence,
        )

        assert claim.id == "claim-1"
        assert claim.section == "experience"
        assert claim.claim == "Six years of Python development"
        assert claim.resume_evidence.source == "resume"

    def test_id_defaults_to_none(self, resume_evidence) -> None:
        claim = ResumeClaim(
            section="skills",
            claim="Production FastAPI experience",
            resume_evidence=resume_evidence,
        )

        assert claim.id is None

    def test_strips_whitespace_from_fields(self) -> None:
        claim = ResumeClaim(
            id="  claim-1  ",
            section=" experience ",
            claim="  Production Python experience  ",
            resume_evidence=Evidence(
                text="  Built production Python services  ",
                source="resume",
                location="  Experience  ",
            ),
        )

        assert claim.id == "claim-1"
        assert claim.section == "experience"
        assert claim.claim == "Production Python experience"
        assert claim.resume_evidence.text == "Built production Python services"
        assert claim.resume_evidence.location == "Experience"

    def test_rejects_non_resume_evidence(self, jd_evidence) -> None:
        with pytest.raises(
            ValidationError,
            match="resume_evidence.source must be 'resume'",
        ):
            ResumeClaim(
                id="claim-1",
                section="experience",
                claim="Python development",
                resume_evidence=jd_evidence,
            )

    def test_rejects_claim_longer_than_maximum(self, resume_evidence) -> None:
        with pytest.raises(ValidationError):
            ResumeClaim(
                id="claim-1",
                section="skills",
                claim="x" * 151,
                resume_evidence=resume_evidence,
            )

    def test_accepts_claim_at_maximum_length(self, resume_evidence) -> None:
        claim = ResumeClaim(
            id="claim-1",
            section="skills",
            claim="x" * 150,
            resume_evidence=resume_evidence,
        )

        assert len(claim.claim) == 150

    def test_accepts_unknown_section_without_validation_context(
        self,
        resume_evidence,
    ) -> None:
        claim = ResumeClaim(
            id="claim-1",
            section="not-a-tech-section",
            claim="Some claim",
            resume_evidence=resume_evidence,
        )

        assert claim.section == "not-a-tech-section"

    def test_rejects_section_outside_domain_taxonomy_when_context_is_supplied(
        self,
        resume_evidence,
    ) -> None:
        with pytest.raises(
            ValidationError,
            match="section 'not-a-tech-section' is not in this domain's section taxonomy",
        ):
            ResumeClaim.model_validate(
                {
                    "id": "claim-1",
                    "section": "not-a-tech-section",
                    "claim": "Some claim",
                    "resume_evidence": {
                        "text": resume_evidence.text,
                        "source": "resume",
                    },
                },
                context={
                    "valid_sections": {
                        "skills",
                        "experience",
                        "education",
                        "project",
                    },
                },
            )

    def test_accepts_section_in_supplied_domain_taxonomy(
        self,
        resume_evidence,
    ) -> None:
        claim = ResumeClaim.model_validate(
            {
                "id": "claim-1",
                "section": "project",
                "claim": "Built a FastAPI service",
                "resume_evidence": {
                    "text": "Built a FastAPI service",
                    "source": "resume",
                },
            },
            context={
                "valid_sections": {
                    "skills",
                    "experience",
                    "education",
                    "project",
                },
            },
        )

        assert claim.section == "project"
