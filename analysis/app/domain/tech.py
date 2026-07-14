"""Tech domain prompts. Hand-tuned for software engineering roles."""

from pydantic import BaseModel, Field, ConfigDict, create_model
from app.schemas import SectionId, SectionScore

PROMPT_VERSION = "tech-v4"

JD_EXTRACTION_PROMPT = """\
You extract structured requirements from a software-engineering job description.

For each distinct requirement, produce:
- A short imperative phrase capturing what's required (e.g. "5+ years Python").
- The section it belongs to: experience, project, education, or skills.
- The exact verbatim quote from the JD that the requirement is drawn from.
  Include nothing beyond the JD's own words. Maximum 240 characters.
- An optional location label that reflects the JD's own grouping. Use
  the literal heading name from the JD when present:
    * "Requirements" / "Must-have" / "Responsibilities" -> use "Requirements"
    * "Nice to have" / "Bonus" / "Preferred" -> use "Nice to have"
    * "About the role" / "What you'll do" / intro paragraph -> use "About the role"
  This lets the UI distinguish hard-requirements from preferred-bonuses.

Be exhaustive. Always extract, even when stated alongside other requirements:
- **Years-of-experience / seniority** requirements (e.g. "5+ years backend
  engineering", "senior-level", "10+ years in production systems"). These
  belong in the `experience` section.
- **Education-level** requirements (e.g. "Bachelor's degree in CS", "Master's
  preferred"). Belong in `education`.
- **Specific named technologies** -- frameworks, libraries, languages,
  databases, cloud providers, tools. Belong in `skills`.
- **Projects / Open-source contributions / Architecture design / Portfolios**
  requirements. Belong in `project`.

When a single sentence states multiple distinct signals -- e.g. "5+ years of
backend engineering experience with Python or Go" -- extract each signal as
its own requirement (one for years-of-experience, one for the language).
Both can share the same verbatim quote as evidence.

Aim for 8-16 requirements total. Spread them across the four sections in
roughly the proportions the JD actually emphasizes. Do not invent
requirements just to fill a section, but also do not skip core requirements
because they overlap stylistically with others.

Hard rules:
- The evidence text MUST be a verbatim quote from the JD. No paraphrasing.
- Never include requirements that aren't actually in the JD.
- Do not include a top-line / overall score concept anywhere.
"""

RESUME_EXTRACTION_PROMPT = """\
You extract structured claims from a candidate's resume.

For each distinct claim the resume makes, produce:
- A short statement capturing the claim (e.g. "4.5 years Python at Stride Financial").
- The section it belongs to: experience, project, education, or skills.
- The exact verbatim quote from the resume that supports the claim.
  Maximum 240 characters.
- An optional location label (e.g. "Experience", "Education", "Skills").

Hard rules:
- The evidence text MUST be a verbatim quote from the resume. No paraphrasing.
- Do not include claims not actually made by the resume.
"""

SCORING_PROMPT = """\
You score how well a resume matches a job description, requirement by
requirement. You will receive a list of JD requirements and a list of resume
claims. For each requirement:

Decide if the resume matches it: **strong / partial / weak / none**.

- `strong` = the resume directly demonstrates the requirement, with concrete
  evidence in the candidate's own claims.
- `partial` = the resume covers PART of the requirement but not all of it.
  This is the right label whenever the JD lists multiple specific items and
  the resume only has some. Example: JD requires "FastAPI, gRPC, and OpenAPI",
  resume has FastAPI + gRPC but not OpenAPI -> `partial`.
- `weak` = the resume hints at adjacent experience but doesn't directly
  address the requirement.
- `none` = the resume doesn't address the requirement at all.

**Treat "or similar" and slash-separated alternatives as either/or.** When the
JD says "Kafka, RabbitMQ, or similar", any equivalent message queue (Redis,
arq, NATS, etc.) qualifies as `strong`. When the JD says "AWS / GCP", having
either qualifies as `strong`. Do not require all listed options when the JD
explicitly allows alternatives.

**Be conservative on Strong.** A Strong match needs at least one CONCRETE
piece of resume evidence -- a quoted sentence demonstrating the requirement.
General summary-line claims ("Strong in PostgreSQL") combined with NO
specific experience-section evidence are at best `partial`. If the only
evidence is a one-word entry in a Skills section, that's `partial`, not
`strong`. Reserve `strong` for cases where the resume shows the candidate
actually used the thing in production (work history, project, OSS).

**Comma-separated tool lists without "or similar" require ALL items for
strong.** JD says "Docker, Kubernetes, and CI/CD pipelines" -- resume must
demonstrate all three (Docker AND K8s AND a CI/CD tool) for `strong`. Two
out of three -> `partial`. One out of three -> `weak`. Same rule for
parenthetical lists: "(FastAPI, gRPC, OpenAPI)" needs all three for
`strong` unless joined by "or".

For each requirement:
- If matched, attach the resume claim(s) that prove the match. Include each
  one's verbatim evidence quote.
- If unmatched, leave the resume evidence empty and optionally add a one-line
  factual note about what the resume is missing. Never offer advice or
  suggest things the candidate should add.

Then group the requirements into the four sections (skills, experience, education, project) and assign each section a score 0-100 based on
proportion and strength of matched requirements. A section with all
`strong` matches -> close to 100. A section with mixed `partial` and `none` ->
in the 40-70 range. A section where the resume doesn't address any
requirement -> close to 0.

You must also provide two qualitative reviews:
1. `sections_analysis`: A qualitative review (1-2 sentences) for EACH of the four sections explaining the major gaps or strengths.
2. `complete_analysis`: A short executive summary (2-3 sentences) evaluating the candidate's overall fit for the role.

Hard rules:
- No top-line overall percentage -- section-level only.
- Evidence quotes must remain verbatim from their source.
- Notes are factual statements about the gap, never advice.
  GOOD: "The JD requires Kubernetes; the resume doesn't mention it."
  BAD:  "Consider learning Kubernetes."
"""


from app.domain.base import DomainStrategy

class TechDomain(DomainStrategy):
    """
    Implementation of the DomainStrategy for the Technology/IT industry.
    Provides hand-tuned prompts for evaluating software engineering and technical roles,
    placing heavy weight on exact technical skills and relevant engineering experience.
    """
    
    @property
    def name(self) -> str:
        return "tech"
        
    @property
    def prompt_version(self) -> str:
        return PROMPT_VERSION

    def jd_extraction_prompt(self) -> str:
        return JD_EXTRACTION_PROMPT

    def resume_extraction_prompt(self) -> str:
        return RESUME_EXTRACTION_PROMPT

    def scoring_prompt(self) -> str:
        return SCORING_PROMPT

    def section_taxonomy(self) -> list[SectionId]:
        return ["experience", "project", "education", "skills"]

    def section_weights(self) -> dict[SectionId, float]:
        """
        Based on modern technical recruiting standards (ATS filters + Hiring Managers)
        Skills (Keywords/Tech Stack) and Experience are the primary drivers.
        Education is increasingly de-emphasized in favor of Projects/Portfolios.
        """
        return {"experience": 0.40, "skills": 0.40, "project": 0.15, "education": 0.05}

    def get_final_schema(self) -> type[BaseModel]:
        fields = {}
        for sec in self.section_taxonomy():
            fields[sec] = (str, Field(description=f"Qualitative review of {sec} gap (1-2 sentences max)"))
            
        SectionsAnalysis = create_model('SectionsAnalysis', **fields, __config__=ConfigDict(populate_by_name=True))
        
        class DynamicFinalAnalysisResult(BaseModel):
            model_config = ConfigDict(populate_by_name=True)

            complete_analysis: str = Field(serialization_alias="completeAnalysis", description="Executive summary of the candidate's overall fit (2-3 sentences max)")
            sections_analysis: SectionsAnalysis = Field(serialization_alias="sectionsAnalysis")
            sections: list[SectionScore]
            
        return DynamicFinalAnalysisResult
