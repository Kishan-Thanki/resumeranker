#!/usr/bin/env python3
import sys
from pathlib import Path

import yaml

root = Path(__file__).resolve().parents[1]
workflow_dir = root / ".github" / "workflows"
files = sorted(workflow_dir.glob("*.yml"))
if not files:
    print("No workflow YAML files found")
    sys.exit(1)

for path in files:
    try:
        yaml.safe_load(path.read_text())
        print(f"OK {path.relative_to(root)}")
    except yaml.YAMLError as exc:
        print(f"FAIL {path.relative_to(root)}: {exc}")
        sys.exit(1)

print(f"Validated {len(files)} workflow files")
