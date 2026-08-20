from __future__ import annotations

from types import SimpleNamespace
from typing import Any

from analysis_dashboard.config import ConnectionConfig

from analysis_dashboard import main


class _Context:
    def __init__(self, owner: "FakeMainStreamlit") -> None:
        self.owner = owner

    def __enter__(self) -> "_Context":
        return self

    def __exit__(self, exc_type, exc, tb) -> None:
        return None

    def update(self, **kwargs: Any) -> None:
        self.owner.status_updates.append(kwargs)

    def write(self, _message: str) -> None:
        return None


class FakeMainStreamlit:
    def __init__(self) -> None:
        self.session_state: dict[str, Any] = {}
        self.errors: list[str] = []
        self.captions: list[str] = []
        self.status_updates: list[dict[str, Any]] = []
        self.tabs_called_with: list[str] | None = None

    def markdown(self, message: str, **kwargs: Any) -> None:
        return None

    def subheader(self, _message: str) -> None:
        return None

    def error(self, message: str) -> None:
        self.errors.append(message)

    def warning(self, _message: str) -> None:
        return None

    def caption(self, message: str) -> None:
        self.captions.append(message)

    def status(self, label: str, **kwargs: Any) -> _Context:
        return _Context(self)

    def tabs(self, labels: list[str]) -> list[_Context]:
        self.tabs_called_with = labels
        return [_Context(self) for _ in labels]


def _configure_app(monkeypatch, fake_st: FakeMainStreamlit, *, submit_pressed: bool) -> None:
    connection = ConnectionConfig(host="analysis", port=50051, timeout_seconds=30.0)
    files = (object(), object()) if submit_pressed else (None, None)

    monkeypatch.setattr(main, "st", fake_st)
    monkeypatch.setattr(main, "apply_app_styles", lambda: None)
    monkeypatch.setattr(main, "render_sidebar", lambda _connection: (connection, *files, submit_pressed, False))
    monkeypatch.setattr(main, "render_empty_state", lambda _connection: None)
    monkeypatch.setattr(main, "validate_pdf_upload", lambda *_args: None)
    monkeypatch.setattr(main, "reset_analysis_state", lambda: fake_st.session_state.update(
        analysis_result=None,
        analysis_metadata=None,
        analysis_error=None,
        analysis_request_id=None,
    ))


def test_run_app_displays_friendly_service_error_once(monkeypatch) -> None:
    fake_st = FakeMainStreamlit()
    _configure_app(monkeypatch, fake_st, submit_pressed=True)
    response = SimpleNamespace(success=False, error_message="ERR_SCHEMA_DRIFT_VALIDATION: Gemini 503")
    monkeypatch.setattr(
        main,
        "call_analysis_service",
        lambda *_args: (response, None, "ui-test", 1.0),
    )

    main.run_app()

    assert fake_st.errors == [
        "We couldn't complete the analysis right now. Please try again shortly."
    ]
    assert fake_st.captions == []
    assert fake_st.status_updates[-1] == {
        "label": "Analysis could not be completed",
        "state": "error",
        "expanded": False,
    }


def test_run_app_handles_missing_uploads_without_calling_service(monkeypatch) -> None:
    fake_st = FakeMainStreamlit()
    _configure_app(monkeypatch, fake_st, submit_pressed=False)
    service_called = False

    def fail_if_called(*_args):
        nonlocal service_called
        service_called = True
        raise AssertionError("service should not be called")

    monkeypatch.setattr(main, "call_analysis_service", fail_if_called)
    monkeypatch.setattr(main, "render_sidebar", lambda _connection: (
        ConnectionConfig(host="analysis", port=50051, timeout_seconds=30.0),
        None,
        None,
        True,
        False,
    ))

    main.run_app()

    assert fake_st.errors == ["Please upload both the job description and the resume."]
    assert service_called is False


def test_run_app_stores_success_metadata(monkeypatch) -> None:
    fake_st = FakeMainStreamlit()
    _configure_app(monkeypatch, fake_st, submit_pressed=True)
    response = SimpleNamespace(
        success=True,
        result_json='{"completeAnalysis": "Good fit", "sections": []}',
        model="test-model",
        input_tokens=10,
        output_tokens=5,
        total_tokens=15,
    )
    monkeypatch.setattr(
        main,
        "call_analysis_service",
        lambda *_args: (response, None, "ui-success", 2.5),
    )
    monkeypatch.setattr(
        main,
        "parse_analysis_result",
        lambda _result_json: {"completeAnalysis": "Good fit", "sections": []},
    )
    monkeypatch.setattr(main, "render_overview", lambda *_args: None)
    monkeypatch.setattr(main, "render_section", lambda *_args: None)

    main.run_app()

    assert fake_st.session_state["analysis_error"] is None
    assert fake_st.session_state["analysis_metadata"] == {
        "request_id": "ui-success",
        "model": "test-model",
        "input_tokens": 10,
        "output_tokens": 5,
        "total_tokens": 15,
        "elapsed_seconds": 2.5,
    }
    assert fake_st.tabs_called_with == ["Overview", "Sections"]
    assert fake_st.status_updates[-1] == {
        "label": "Analysis complete",
        "state": "complete",
        "expanded": False,
    }
