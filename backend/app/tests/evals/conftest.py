"""Eval-suite fixtures. Marked `eval`; excluded from the default pytest run."""

from __future__ import annotations

import json
import pathlib

import pytest

HERE = pathlib.Path(__file__).parent
JD_DIR = HERE / "fixtures" / "jds"
RESUME_DIR = HERE / "fixtures" / "resumes"
EXPECTED_DIR = HERE / "expected"


def _load_pairs() -> list[tuple[str, str, dict[str, dict[str, int]]]]:
    pairs: list[tuple[str, str, dict[str, dict[str, int]]]] = []
    for path in sorted(EXPECTED_DIR.glob("*.json")):
        payload = json.loads(path.read_text(encoding="utf-8"))
        pairs.append((payload["jd"], payload["resume"], payload["section_ranges"]))
    return pairs


@pytest.fixture(scope="session")
def jd_texts() -> dict[str, str]:
    return {p.stem: p.read_text(encoding="utf-8") for p in JD_DIR.glob("*.txt")}


@pytest.fixture(scope="session")
def resume_texts() -> dict[str, str]:
    return {p.stem: p.read_text(encoding="utf-8") for p in RESUME_DIR.glob("*.txt")}


@pytest.fixture(scope="session")
def expected_pairs() -> list[tuple[str, str, dict[str, dict[str, int]]]]:
    return _load_pairs()
