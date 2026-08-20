from __future__ import annotations

from typing import Any

from analysis_dashboard.config import ConnectionConfig

from analysis_dashboard import rendering


class _Context:
    def __init__(self, owner: "FakeRenderingStreamlit") -> None:
        self.owner = owner

    def __enter__(self) -> "_Context":
        return self

    def __exit__(self, exc_type, exc, tb) -> None:
        return None


class FakeRenderingStreamlit:
    def __init__(self) -> None:
        self.calls: list[tuple[str, Any]] = []
        self.uploaders: list[dict[str, Any]] = []
        self.form_submit_result = False
        self.clear_result = False
        self.sidebar = _Context(self)

    def markdown(self, message: str, **kwargs: Any) -> None:
        self.calls.append(("markdown", message))

    def caption(self, message: str) -> None:
        self.calls.append(("caption", message))

    def columns(self, spec: Any) -> list[_Context]:
        count = spec if isinstance(spec, int) else len(spec)
        return [_Context(self) for _ in range(count)]

    def container(self, **kwargs: Any) -> _Context:
        self.calls.append(("container", kwargs))
        return _Context(self)

    def expander(self, label: str, **kwargs: Any) -> _Context:
        self.calls.append(("expander", label))
        return _Context(self)

    def form(self, name: str, **kwargs: Any) -> _Context:
        self.calls.append(("form", name))
        return _Context(self)

    def file_uploader(self, label: str, **kwargs: Any) -> None:
        self.uploaders.append({"label": label, **kwargs})
        return None

    def form_submit_button(self, label: str, **kwargs: Any) -> bool:
        self.calls.append(("form_submit_button", label))
        return self.form_submit_result

    def button(self, label: str, **kwargs: Any) -> bool:
        self.calls.append(("button", label))
        return self.clear_result


def test_render_sidebar_configures_two_single_pdf_uploaders(monkeypatch) -> None:
    fake_st = FakeRenderingStreamlit()
    monkeypatch.setattr(rendering, "st", fake_st)
    connection = ConnectionConfig(host="analysis", port=50051, timeout_seconds=30.0)

    result = rendering.render_sidebar(connection)

    assert result == (connection, None, None, False, False)
    assert [uploader["label"] for uploader in fake_st.uploaders] == [
        "Job description PDF",
        "Resume PDF",
    ]
    assert all(uploader["type"] == ["pdf"] for uploader in fake_st.uploaders)
    assert all(uploader["accept_multiple_files"] is False for uploader in fake_st.uploaders)


def test_badge_html_allows_only_known_tones() -> None:
    assert "badge-strong" in rendering.badge_html("Strong", "strong")
    assert "badge-none" in rendering.badge_html("Unknown", "unexpected")
    assert "<script>" not in rendering.badge_html("<script>alert(1)</script>", "strong")


def test_render_stat_card_escapes_values_and_optional_note(monkeypatch) -> None:
    fake_st = FakeRenderingStreamlit()
    monkeypatch.setattr(rendering, "st", fake_st)

    rendering.render_stat_card("<Label>", "<Value>", "<Note>")

    message = fake_st.calls[0][1]
    assert "&lt;Label&gt;" in message
    assert "&lt;Value&gt;" in message
    assert "&lt;Note&gt;" in message


def test_render_overall_score_averages_section_scores(monkeypatch) -> None:
    fake_st = FakeRenderingStreamlit()
    monkeypatch.setattr(rendering, "st", fake_st)

    rendering.render_overall_score(
        [
            {"score": 90},
            {"score": 70},
        ]
    )

    message = fake_st.calls[0][1]
    assert "score-strong" in message
    assert ">80<" in message
    assert "Strong alignment" in message


def test_render_evidence_box_escapes_content(monkeypatch) -> None:
    fake_st = FakeRenderingStreamlit()
    monkeypatch.setattr(rendering, "st", fake_st)

    rendering.render_evidence_box(
        "Resume Evidence",
        [{"location": "<location>", "text": "<b>evidence</b>"}],
        "No evidence",
    )

    rendered = "\n".join(str(value) for _, value in fake_st.calls)
    assert "&lt;location&gt;" in rendered
    assert "&lt;b&gt;evidence&lt;/b&gt;" in rendered
    assert "<b>evidence</b>" not in rendered
