from __future__ import annotations

from typing import Any, Self

from analysis_dashboard import legal


class FakeLegalStreamlit:
    def __init__(self, *, checkbox_value: bool = False, button_value: bool = False) -> None:
        self.session_state: dict[str, Any] = {}
        self.checkbox_value = checkbox_value
        self.button_value = button_value
        self.rerun_called = False

    def title(self, _message: str) -> None:
        return None

    def caption(self, _message: str) -> None:
        return None

    def markdown(self, _message: str, **_kwargs: Any) -> None:
        return None

    def container(self, **_kwargs: Any):
        return _Context()

    def expander(self, _label: str):
        return _Context()

    def checkbox(self, _label: str, **_kwargs: Any) -> bool:
        return self.checkbox_value

    def button(self, _label: str, **_kwargs: Any) -> bool:
        return self.button_value

    def rerun(self) -> None:
        self.rerun_called = True


class _Context:
    def __enter__(self) -> Self:
        return self

    def __exit__(self, exc_type, exc, tb) -> None:
        return None


def test_has_accepted_terms_false_when_not_set(monkeypatch) -> None:
    fake_st = FakeLegalStreamlit()
    fake_st.session_state = {}
    monkeypatch.setattr(legal, "st", fake_st)

    assert legal.has_accepted_terms() is False


def test_has_accepted_terms_false_when_version_mismatch(monkeypatch) -> None:
    fake_st = FakeLegalStreamlit()
    fake_st.session_state = {"terms_accepted": True, "terms_accepted_version": "old-version"}
    monkeypatch.setattr(legal, "st", fake_st)

    assert legal.has_accepted_terms() is False


def test_has_accepted_terms_true_when_current_version(monkeypatch) -> None:
    fake_st = FakeLegalStreamlit()
    fake_st.session_state = {
        "terms_accepted": True,
        "terms_accepted_version": legal.TERMS_VERSION,
    }
    monkeypatch.setattr(legal, "st", fake_st)

    assert legal.has_accepted_terms() is True


def test_render_terms_gate_returns_true_when_already_accepted(monkeypatch) -> None:
    fake_st = FakeLegalStreamlit()
    fake_st.session_state = {
        "terms_accepted": True,
        "terms_accepted_version": legal.TERMS_VERSION,
    }
    monkeypatch.setattr(legal, "st", fake_st)

    assert legal.render_terms_gate() is True


def test_render_terms_gate_blocks_until_checkbox_and_button(monkeypatch) -> None:
    fake_st = FakeLegalStreamlit(checkbox_value=False, button_value=False)
    fake_st.session_state = {}
    monkeypatch.setattr(legal, "st", fake_st)

    assert legal.render_terms_gate() is False
    assert "terms_accepted" not in fake_st.session_state
    assert fake_st.rerun_called is False


def test_render_terms_gate_accepts_when_checked_and_clicked(monkeypatch) -> None:
    fake_st = FakeLegalStreamlit(checkbox_value=True, button_value=True)
    fake_st.session_state = {}
    monkeypatch.setattr(legal, "st", fake_st)

    result = legal.render_terms_gate()

    assert fake_st.session_state["terms_accepted"] is True
    assert fake_st.session_state["terms_accepted_version"] == legal.TERMS_VERSION
    assert fake_st.rerun_called is True
    assert result is False
