from __future__ import annotations

import streamlit as st

from analysis_dashboard.config import load_default_connection
from analysis_dashboard.grpc_client import call_analysis_service
from analysis_dashboard.parsing import (
    format_analysis_error,
    parse_analysis_result,
    safe_str,
    validate_pdf_upload,
)
from analysis_dashboard.rendering import (
    apply_app_styles,
    render_empty_state,
    render_overview,
    render_section,
    render_sidebar,
)
from analysis_dashboard.state import reset_analysis_state

DEFAULT_CONNECTION = load_default_connection()


def run_app() -> None:
    if "analysis_result" not in st.session_state:
        reset_analysis_state()

    apply_app_styles()
    connection, jd_file, resume_file, submit_pressed, clear_pressed = render_sidebar(
        DEFAULT_CONNECTION
    )

    st.markdown(
        f"""
        <div class="hero">
            <div class="hero-kicker">ResumeRanker analysis workspace</div>
            <h1 class="hero-title">Analyze a resume against a job description</h1>
            <p class="hero-subtitle">
                Evaluate candidate fit against job requirements and generate structured alignment reports.
            </p>
        </div>
        """,
        unsafe_allow_html=True,
    )

    if clear_pressed:
        reset_analysis_state()
        st.rerun()

    if submit_pressed:
        reset_analysis_state()
        if jd_file is None or resume_file is None:
            st.error("Please upload both the job description and the resume.")
        else:
            try:
                validate_pdf_upload(jd_file, "Job description")
                validate_pdf_upload(resume_file, "Resume")
            except ValueError as exc:
                st.error(str(exc))
            else:
                with st.status("Running ResumeRanker analysis...", expanded=True) as status:
                    response, grpc_error, request_id, elapsed_seconds = call_analysis_service(
                        connection,
                        jd_file,
                        resume_file,
                    )

                    st.session_state["analysis_request_id"] = request_id

                    if grpc_error is not None:
                        st.session_state["analysis_error"] = grpc_error
                        status.update(
                            label="Analysis could not be completed",
                            state="error",
                            expanded=False,
                        )
                    elif response is None:
                        st.session_state["analysis_error"] = (
                            "The analysis service returned no response."
                        )
                        status.update(
                            label="Analysis could not be completed",
                            state="error",
                            expanded=False,
                        )
                    elif not response.success:
                        error_message = safe_str(
                            response.error_message,
                            "Unknown analysis error.",
                        )
                        st.session_state["analysis_error"] = error_message
                        status.update(
                            label="Analysis could not be completed",
                            state="error",
                            expanded=False,
                        )
                    else:
                        try:
                            parsed_result = parse_analysis_result(response.result_json)
                        except (TypeError, ValueError) as exc:
                            st.session_state["analysis_error"] = str(exc)
                            status.update(
                                label="Analysis could not be completed",
                                state="error",
                                expanded=False,
                            )
                        else:
                            st.session_state["analysis_result"] = parsed_result
                            st.session_state["analysis_metadata"] = {
                                "request_id": request_id,
                                "model": safe_str(response.model, "Unknown"),
                                "input_tokens": int(response.input_tokens),
                                "output_tokens": int(response.output_tokens),
                                "total_tokens": int(response.total_tokens),
                                "elapsed_seconds": elapsed_seconds,
                            }
                            st.session_state["analysis_error"] = None
                            status.update(
                                label="Analysis complete",
                                state="complete",
                                expanded=False,
                            )

    stored_result = st.session_state.get("analysis_result")
    stored_metadata = st.session_state.get("analysis_metadata")
    stored_error = st.session_state.get("analysis_error")

    if isinstance(stored_error, str) and stored_error:
        _error_code, user_message = format_analysis_error(stored_error)
        st.error(user_message)

    if isinstance(stored_result, dict) and isinstance(stored_metadata, dict):
        overview_tab, sections_tab = st.tabs(["Overview", "Sections"])

        with overview_tab:
            render_overview(stored_result, stored_metadata)

        with sections_tab:
            st.subheader("Section Details")
            sections = stored_result.get("sections", [])
            if not isinstance(sections, list) or not sections:
                st.warning("No section data was returned.")
            else:
                for section in sections:
                    if isinstance(section, dict):
                        render_section(section)

    else:
        render_empty_state(DEFAULT_CONNECTION)


def main() -> None:
    run_app()


if __name__ == "__main__":
    main()
