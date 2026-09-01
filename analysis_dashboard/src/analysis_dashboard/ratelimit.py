"""Per-visitor throttling for the expensive analyze action.

Caddy sits in front of this app, but Streamlit serves one long-lived
WebSocket per browser tab, so a proxy-level rate limit only slows down new
connections, not repeated button clicks on an already-open tab. This module
gates the actual gRPC call so a single visitor can't drive unbounded LLM
cost, keyed by the client IP Caddy forwards.

In-memory only: fine for a single dashboard process. If this is ever scaled
to multiple replicas behind a load balancer, move the counters to a shared
store (e.g. Redis) instead.
"""

from __future__ import annotations

import threading
import time
from collections import deque

import streamlit as st

from analysis_dashboard.config import parse_bool_env, parse_float_env, parse_int_env

MAX_REQUESTS_PER_WINDOW = parse_int_env("RATE_LIMIT_MAX_REQUESTS", 5)
WINDOW_SECONDS = parse_float_env("RATE_LIMIT_WINDOW_SECONDS", 600.0)
DAILY_REQUEST_LIMIT = parse_int_env("DAILY_REQUEST_LIMIT", 10)
DAILY_WINDOW_SECONDS = parse_float_env("DAILY_WINDOW_SECONDS", 86400.0)
HUMAN_CHECK_ENABLED = parse_bool_env("HUMAN_CHECK_ENABLED", False)
HUMAN_CHECK_THRESHOLD = parse_int_env("HUMAN_CHECK_THRESHOLD", 3)

_lock = threading.Lock()
_hits: dict[str, deque[float]] = {}
_daily_hits: dict[str, deque[float]] = {}
_blocked_attempts: dict[str, int] = {}


def _client_key() -> str:
    headers = st.context.headers or {}
    forwarded_for = headers.get("X-Forwarded-For", "")
    if forwarded_for:
        return forwarded_for.split(",")[0].strip()
    return headers.get("X-Real-Ip", "") or "unknown"


def _check_window_limit(
    hits: deque[float],
    max_requests: int,
    window_seconds: float,
    now: float,
) -> tuple[bool, float]:
    while hits and now - hits[0] > window_seconds:
        hits.popleft()

    if len(hits) >= max_requests:
        retry_after = window_seconds - (now - hits[0])
        return False, max(retry_after, 0.0)

    hits.append(now)
    return True, 0.0


def should_require_human_check() -> bool:
    """Return whether the current client should pass a light challenge."""
    if not parse_bool_env("HUMAN_CHECK_ENABLED", False):
        return False

    key = _client_key()
    threshold = parse_int_env("HUMAN_CHECK_THRESHOLD", 3)
    with _lock:
        blocked_count = _blocked_attempts.get(key, 0)
        return blocked_count >= threshold


def check_rate_limit() -> tuple[bool, float, bool]:
    """Return (allowed, seconds_until_retry, requires_human_check) for the current visitor."""
    key = _client_key()
    now = time.monotonic()

    burst_max = parse_int_env("RATE_LIMIT_MAX_REQUESTS", 5)
    burst_window = parse_float_env("RATE_LIMIT_WINDOW_SECONDS", 600.0)
    daily_limit = parse_int_env("DAILY_REQUEST_LIMIT", 10)
    daily_window = parse_float_env("DAILY_WINDOW_SECONDS", 86400.0)

    with _lock:
        burst_hits = _hits.setdefault(key, deque())
        burst_allowed, burst_retry_after = _check_window_limit(
            burst_hits,
            burst_max,
            burst_window,
            now,
        )
        if not burst_allowed:
            _blocked_attempts[key] = _blocked_attempts.get(key, 0) + 1
            return False, burst_retry_after, should_require_human_check()

        daily_hits = _daily_hits.setdefault(key, deque())
        daily_allowed, daily_retry_after = _check_window_limit(
            daily_hits,
            daily_limit,
            daily_window,
            now,
        )
        if not daily_allowed:
            _blocked_attempts[key] = _blocked_attempts.get(key, 0) + 1
            return False, daily_retry_after, should_require_human_check()

        _blocked_attempts[key] = 0
        return True, 0.0, False
