from __future__ import annotations

import os
from dataclasses import dataclass


MAX_FILE_SIZE_BYTES = 10 * 1024 * 1024
SECTION_ORDER = ["skills", "experience", "education", "project"]
SECTION_ORDER_INDEX = {section_id: index for index, section_id in enumerate(SECTION_ORDER)}

MATCH_LABELS: dict[str, str] = {
    "strong": "Strong",
    "partial": "Partial",
    "weak": "Weak",
    "none": "No Match",
}

VALID_MATCH_STRENGTHS = frozenset(MATCH_LABELS)


@dataclass(frozen=True)
class ConnectionConfig:
    host: str
    port: int
    timeout_seconds: float

    @property
    def address(self) -> str:
        return f"{self.host}:{self.port}"


def parse_int_env(name: str, default: int) -> int:
    raw = os.getenv(name, "").strip()
    if not raw:
        return default
    try:
        return int(raw)
    except ValueError:
        return default


def parse_float_env(name: str, default: float) -> float:
    raw = os.getenv(name, "").strip()
    if not raw:
        return default
    try:
        value = float(raw)
    except ValueError:
        return default
    return value if value > 0 else default


def load_default_connection() -> ConnectionConfig:
    host = os.getenv("ANALYSIS_GRPC_HOST", "localhost").strip() or "localhost"
    port = parse_int_env("ANALYSIS_GRPC_PORT", 50051)
    timeout_seconds = parse_float_env("ANALYSIS_GRPC_TIMEOUT_SECONDS", 180.0)
    return ConnectionConfig(host=host, port=port, timeout_seconds=timeout_seconds)
