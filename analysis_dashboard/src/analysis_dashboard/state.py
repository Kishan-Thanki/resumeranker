from __future__ import annotations

import streamlit as st


def reset_analysis_state() -> None:
    st.session_state["analysis_result"] = None
    st.session_state["analysis_metadata"] = None
    st.session_state["analysis_error"] = None
    st.session_state["analysis_request_id"] = None
