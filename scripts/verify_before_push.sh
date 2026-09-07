#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "== Workflow YAML validation =="
python3 scripts/check_workflows.py

echo "== Docker Compose validation =="
docker compose config >/dev/null

echo "== Analysis tests =="
(cd analysis && uv run pytest -q -m "not e2e")

echo "== Dashboard tests =="
(cd analysis_dashboard && uv run pytest -q)

echo "== Pre-push checks passed =="
