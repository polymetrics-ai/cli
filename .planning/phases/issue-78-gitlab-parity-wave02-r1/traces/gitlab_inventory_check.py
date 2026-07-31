#!/usr/bin/env python3
"""GitLab official inventory parity check for issue #78.

Fetches the pinned public GitLab OpenAPI source and compares its unique operation
count to connector-local api_surface/operations metadata. This is a planning and
verification helper only; it performs no credentialed provider calls.
"""
from __future__ import annotations

import json
import sys
import urllib.request
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[4]
OPENAPI_URL = "https://gitlab.com/gitlab-org/gitlab/-/raw/9cd04099eb59d87335798e4f57a2bc5a2622e4cc/doc/api/openapi/openapi_v2.yaml"
METHODS = {"get", "post", "put", "patch", "delete", "head", "options"}
SURFACE = ROOT / "internal/connectors/defs/gitlab/api_surface.json"
OPERATIONS = ROOT / "internal/connectors/defs/gitlab/operations.json"


def load_official_count() -> int:
    with urllib.request.urlopen(OPENAPI_URL, timeout=60) as response:
        spec = yaml.safe_load(response.read())
    seen: set[tuple[str, str]] = set()
    for path, item in spec.get("paths", {}).items():
        if not isinstance(item, dict):
            continue
        for method in item:
            if method.lower() in METHODS:
                seen.add((method.upper(), path))
    return len(seen)


def main() -> int:
    official = load_official_count()
    surface_rows = len(json.loads(SURFACE.read_text()).get("endpoints", []))
    operations_rows = 0
    if OPERATIONS.exists():
        operations_rows = len(json.loads(OPERATIONS.read_text()).get("operations", []))
    problems = []
    if surface_rows != official:
        problems.append(f"local api_surface row count {surface_rows} != official operation count {official}")
    if OPERATIONS.exists() and operations_rows != official:
        problems.append(f"local operations row count {operations_rows} != official operation count {official}")
    elif not OPERATIONS.exists():
        problems.append("local operations.json missing")
    if problems:
        print("FAIL GitLab inventory parity")
        for problem in problems:
            print(f"- {problem}")
        return 1
    print(f"PASS GitLab inventory parity: {official} official operations represented")
    return 0


if __name__ == "__main__":
    sys.exit(main())
