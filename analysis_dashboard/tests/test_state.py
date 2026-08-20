from __future__ import annotations

from analysis_dashboard import state


def test_reset_analysis_state_clears_all_session_values(monkeypatch) -> None:
    session_state = {
        "analysis_result": {"sections": []},
        "analysis_metadata": {"model": "test-model"},
        "analysis_error": "old error",
        "analysis_request_id": "old-request",
    }
    monkeypatch.setattr(state.st, "session_state", session_state)

    state.reset_analysis_state()

    assert session_state == {
        "analysis_result": None,
        "analysis_metadata": None,
        "analysis_error": None,
        "analysis_request_id": None,
    }
