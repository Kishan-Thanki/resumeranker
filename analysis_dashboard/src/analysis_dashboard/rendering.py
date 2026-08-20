from __future__ import annotations

import html
from typing import Any

import streamlit as st
from streamlit.runtime.uploaded_file_manager import UploadedFile

from analysis_dashboard.config import MATCH_LABELS, SECTION_ORDER, ConnectionConfig
from analysis_dashboard.parsing import (
    build_requirement_metrics,
    clamp_int,
    coerce_string_list,
    normalize_evidence_list,
    normalize_strength,
    safe_str,
)
from analysis_dashboard.styles import APP_STYLES


def apply_app_styles() -> None:
    st.markdown(APP_STYLES, unsafe_allow_html=True)


def render_stat_card(label: str, value: str, note: str | None = None) -> None:
    note_html = f'<div class="stat-note">{html.escape(note)}</div>' if note else ""
    st.markdown(
        f"""
        <div class="stat-card">
            <div class="stat-label">{html.escape(label)}</div>
            <div class="stat-value">{html.escape(value)}</div>
            {note_html}
        </div>
        """,
        unsafe_allow_html=True,
    )


def render_overall_score(sections: list[dict[str, Any]]) -> None:
    scores = [
        clamp_int(section.get("score"), 0, 0, 100)
        for section in sections
        if isinstance(section, dict)
    ]
    overall_score = round(sum(scores) / len(scores)) if scores else 0
    score_tone = (
        "strong"
        if overall_score >= 80
        else "partial"
        if overall_score >= 60
        else "weak"
        if overall_score >= 40
        else "none"
    )
    score_label = {
        "strong": "Strong alignment",
        "partial": "Promising alignment",
        "weak": "Limited alignment",
        "none": "Low alignment",
    }[score_tone]

    st.markdown(
        f"""
        <div class="overall-score-panel">
            <div class="score-ring score-{score_tone}" style="--score: {overall_score * 3.6}deg">
                <div class="score-ring-inner">
                    <div class="score-value">{overall_score}</div>
                    <div class="score-unit">/ 100</div>
                </div>
            </div>
            <div class="overall-score-copy">
                <div class="eyebrow">Overall match</div>
                <h2>{html.escape(score_label)}</h2>
                <p>Average alignment across the analyzed sections.</p>
            </div>
        </div>
        """,
        unsafe_allow_html=True,
    )


def badge_html(text: str, tone: str) -> str:
    tone_class = tone if tone in {"strong", "partial", "weak", "none"} else "none"
    return f'<span class="badge badge-{tone_class}">{html.escape(text)}</span>'


def render_evidence_box(title: str, evidence_items: list[dict[str, Any]], empty_text: str) -> None:
    with st.container(border=True):
        st.markdown(f"**{title}**")

        if not evidence_items:
            st.caption(empty_text)
            return

        for evidence in evidence_items:
            text = safe_str(evidence.get("text"))
            if not text:
                continue

            location = safe_str(evidence.get("location"), "Unspecified location")
            st.markdown(
                f"""
                <div class="evidence-item">
                    <div class="evidence-location">{html.escape(location)}</div>
                    <p class="evidence-text">{html.escape(text)}</p>
                </div>
                """,
                unsafe_allow_html=True,
            )


def render_requirement(requirement: dict[str, Any], index: int) -> None:
    title = safe_str(requirement.get("requirement"), "Unnamed requirement")
    strength = normalize_strength(requirement.get("matchStrength"))
    matched = bool(requirement.get("matched", strength in {"strong", "partial"}))
    note = safe_str(requirement.get("note"))
    claim_ids = coerce_string_list(
        requirement.get("supportingClaimIds") or requirement.get("supporting_claim_ids")
    )
    jd_evidence = normalize_evidence_list(requirement.get("jdEvidence"))
    resume_evidence = normalize_evidence_list(requirement.get("resumeEvidence"))

    with st.expander(f"{index}. {title}", expanded=False):
        top_cols = st.columns(3)
        with top_cols[0]:
            st.markdown(badge_html(MATCH_LABELS[strength], strength), unsafe_allow_html=True)
        with top_cols[1]:
            st.markdown(
                badge_html("Matched" if matched else "Unmatched", strength),
                unsafe_allow_html=True,
            )
        with top_cols[2]:
            st.markdown(
                badge_html(
                    f"{len(jd_evidence)} JD / {len(resume_evidence)} resume evidence",
                    "none",
                ),
                unsafe_allow_html=True,
            )

        if claim_ids:
            st.caption("Supporting claim ids: " + ", ".join(claim_ids))

        if note:
            st.info(note)

        evidence_cols = st.columns(2)
        with evidence_cols[0]:
            render_evidence_box(
                "Job Description Evidence",
                jd_evidence,
                "No JD evidence was returned for this requirement.",
            )
        with evidence_cols[1]:
            render_evidence_box(
                "Resume Evidence",
                resume_evidence,
                "No resume evidence was returned for this requirement.",
            )


def render_section(section: dict[str, Any]) -> None:
    section_id = safe_str(section.get("id"), "section")
    label = safe_str(section.get("label"), section_id.title())
    score = clamp_int(section.get("score"), 0, 0, 100)
    review = safe_str(section.get("review"), "No review was returned for this section.")
    requirements = section.get("requirements", [])
    if not isinstance(requirements, list):
        requirements = []

    matched_count = sum(
        1
        for requirement in requirements
        if isinstance(requirement, dict) and bool(requirement.get("matched", False))
    )
    section_tone = "strong" if score >= 80 else "partial" if score >= 60 else "weak" if score >= 40 else "none"

    with st.container(border=True):
        header_cols = st.columns([3, 1, 1])
        with header_cols[0]:
            st.markdown(f"### {label}")
            st.caption(f"Section id: {section_id}")
        with header_cols[1]:
            st.markdown(badge_html(f"{score}%", section_tone), unsafe_allow_html=True)
        with header_cols[2]:
            st.markdown(
                badge_html(f"{matched_count}/{len(requirements)} matched", section_tone),
                unsafe_allow_html=True,
            )

        st.progress(score / 100.0)
        st.write(review)

        if requirements:
            st.markdown("#### Requirements")
            for index, requirement in enumerate(requirements, start=1):
                if isinstance(requirement, dict):
                    render_requirement(requirement, index)
        else:
            st.caption("No requirements were returned for this section.")


def render_overview(result: dict[str, Any], metadata: dict[str, Any]) -> None:
    complete_analysis = safe_str(
        result.get("completeAnalysis"),
        "No executive summary was returned.",
    )
    sections = result.get("sections", [])
    if not isinstance(sections, list):
        sections = []

    metrics = build_requirement_metrics(sections)

    render_overall_score(sections)

    st.markdown("### Executive Summary")
    st.info(complete_analysis)

    result_skills = coerce_string_list(result.get("matchedSkills") or result.get("matched_skills"))
    missing_skills = coerce_string_list(
        result.get("missingCriticalSkills") or result.get("missing_critical_skills")
    )

    if result_skills or missing_skills:
        skill_cols = st.columns(2)
        with skill_cols[0]:
            matched_html = "".join(
                f'<span class="skill-chip skill-chip-positive">{html.escape(skill)}</span>'
                for skill in result_skills
            )
            st.markdown(
                "<div class=\"skill-panel\"><div class=\"panel-label\">Matched skills</div>"
                + (matched_html or '<span class="muted-text">None returned</span>')
                + "</div>",
                unsafe_allow_html=True,
            )
        with skill_cols[1]:
            missing_html = "".join(
                f'<span class="skill-chip skill-chip-negative">{html.escape(skill)}</span>'
                for skill in missing_skills
            )
            st.markdown(
                "<div class=\"skill-panel\"><div class=\"panel-label\">Missing critical skills</div>"
                + (missing_html or '<span class="muted-text">None returned</span>')
                + "</div>",
                unsafe_allow_html=True,
            )

    st.markdown("### Requirement Coverage")
    coverage_cols = st.columns(6)
    with coverage_cols[0]:
        render_stat_card("Requirements", str(metrics["total"]), "Across all sections")
    with coverage_cols[1]:
        render_stat_card("Matched", str(metrics["matched"]), "Strong + partial")
    with coverage_cols[2]:
        render_stat_card("Strong", str(metrics["strong"]))
    with coverage_cols[3]:
        render_stat_card("Partial", str(metrics["partial"]))
    with coverage_cols[4]:
        render_stat_card("Weak", str(metrics["weak"]))
    with coverage_cols[5]:
        render_stat_card("No match", str(metrics["none"]))

    st.markdown("### Section Snapshot")
    if not sections:
        st.warning("The analysis service returned no sections.")
        return

    section_ids = [safe_str(section.get("id")) for section in sections if isinstance(section, dict)]
    missing_ids = [section_id for section_id in SECTION_ORDER if section_id not in section_ids]
    extra_ids = [section_id for section_id in section_ids if section_id not in SECTION_ORDER]

    if missing_ids or extra_ids:
        st.warning(
            "The response section taxonomy did not fully match the expected tech domain. "
            f"Missing: {missing_ids or 'none'}; Unexpected: {extra_ids or 'none'}."
        )

    preview_cols = st.columns(min(4, len(sections)))
    for column, section in zip(preview_cols, sections):
        if not isinstance(section, dict):
            continue

        section_id = safe_str(section.get("id"), "section")
        label = safe_str(section.get("label"), section_id.title())
        score = clamp_int(section.get("score"), 0, 0, 100)
        review = safe_str(section.get("review"), "No review was returned for this section.")
        requirements = section.get("requirements", [])
        requirement_count = len(requirements) if isinstance(requirements, list) else 0
        matched_count = (
            sum(
                1
                for requirement in requirements
                if isinstance(requirement, dict) and bool(requirement.get("matched", False))
            )
            if isinstance(requirements, list)
            else 0
        )
        tone = "strong" if score >= 80 else "partial" if score >= 60 else "weak" if score >= 40 else "none"

        with column:
            render_stat_card(label, f"{score}%", f"{matched_count}/{requirement_count} matched")
            st.markdown(badge_html(section_id, tone), unsafe_allow_html=True)
            st.caption(review if len(review) <= 140 else review[:137] + "...")


def render_empty_state(connection: ConnectionConfig) -> None:
    st.markdown(
        """
        <div class="hero">
            <div class="hero-kicker">ResumeRanker analysis</div>
            <h1 class="hero-title">Evidence-backed candidate fit</h1>
            <p class="hero-subtitle">
                Upload a job description and a candidate resume to evaluate alignment across core technical areas.
            </p>
        </div>
        """,
        unsafe_allow_html=True,
    )


def render_sidebar(
    connection: ConnectionConfig,
) -> tuple[ConnectionConfig, UploadedFile | None, UploadedFile | None, bool, bool]:
    with st.sidebar:
        st.markdown("## Documents")

        with st.form("analysis-upload-form", clear_on_submit=False):
            jd_file = st.file_uploader(
                "Job description PDF",
                type=["pdf"],
                accept_multiple_files=False,
                help="Upload the target job description as a PDF.",
            )

            resume_file = st.file_uploader(
                "Resume PDF",
                type=["pdf"],
                accept_multiple_files=False,
                help="Upload the candidate resume as a PDF.",
            )

            submit_pressed = st.form_submit_button(
                "Run analysis",
                type="primary",
                use_container_width=True,
            )

        clear_pressed = st.button("Clear last result", use_container_width=True)

    return connection, jd_file, resume_file, submit_pressed, bool(clear_pressed)
