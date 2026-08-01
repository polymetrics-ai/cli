#!/usr/bin/env python3
"""GitLab official inventory parity check for issue #78.

Fetches the pinned public GitLab OpenAPI source and compares its unique operation
count to connector-local api_surface/operations metadata. This is a planning and
verification helper only; it performs no credentialed provider calls.
"""
from __future__ import annotations

import json
import re
import sys
import urllib.request
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[4]
OPENAPI_URL = "https://gitlab.com/gitlab-org/gitlab/-/raw/9cd04099eb59d87335798e4f57a2bc5a2622e4cc/doc/api/openapi/openapi_v2.yaml"
METHODS = {"get", "post", "put", "patch", "delete", "head", "options"}
SURFACE = ROOT / "internal/connectors/defs/gitlab/api_surface.json"
OPERATIONS = ROOT / "internal/connectors/defs/gitlab/operations.json"
SUPPLEMENTAL_SURFACE = {("GET", "/users")}


def normalize_gitlab_path(path: str) -> str:
    path = re.sub(r"\(\*([A-Za-z0-9_]+)/\)", r"{\1}/", path)
    return re.sub(r"\*([A-Za-z0-9_]+)", r"{\1}", path)


def relative_path(path: str) -> str:
    path = normalize_gitlab_path(path)
    if path.startswith("/api/v4"):
        stripped = path[len("/api/v4") :]
        return stripped or "/"
    return path


def load_official_operations() -> set[tuple[str, str]]:
    with urllib.request.urlopen(OPENAPI_URL, timeout=60) as response:
        spec = yaml.safe_load(response.read())
    seen: set[tuple[str, str]] = set()
    for path, item in spec.get("paths", {}).items():
        if not isinstance(item, dict):
            continue
        for method in item:
            if method.lower() in METHODS:
                seen.add((method.upper(), relative_path(path)))
    return seen


def operation_keys(rows: list[dict[str, object]]) -> set[tuple[str, str]]:
    keys: set[tuple[str, str]] = set()
    for row in rows:
        rest = row.get("rest")
        if isinstance(rest, dict):
            keys.add((str(rest.get("method")), str(rest.get("path"))))
            continue
        binary = row.get("binary")
        if isinstance(binary, dict):
            keys.add((str(binary.get("method")), str(binary.get("path"))))
            continue
        composite = row.get("composite")
        if isinstance(composite, dict):
            steps = composite.get("steps")
            if isinstance(steps, list) and steps:
                method, path, *_ = str(steps[0]).split(" ", 2)
                keys.add((method, path))
    return keys


def main() -> int:
    official = load_official_operations()
    surface_rows = json.loads(SURFACE.read_text()).get("endpoints", [])
    operations_rows = json.loads(OPERATIONS.read_text()).get("operations", []) if OPERATIONS.exists() else []
    surface_keys = {(row.get("method"), row.get("path")) for row in surface_rows if isinstance(row, dict)}
    operations_keys = operation_keys(operations_rows)
    problems = []
    if len(surface_rows) != len(official) + len(SUPPLEMENTAL_SURFACE):
        problems.append(
            f"local api_surface row count {len(surface_rows)} != official operation count {len(official)} plus supplemental {len(SUPPLEMENTAL_SURFACE)}"
        )
    if operations_keys != official:
        missing = sorted(official - operations_keys)[:5]
        extra = sorted(operations_keys - official)[:5]
        problems.append(f"operations path set mismatch missing={missing} extra={extra}")
    if not SUPPLEMENTAL_SURFACE.issubset(surface_keys):
        problems.append(f"missing supplemental api_surface rows {sorted(SUPPLEMENTAL_SURFACE - surface_keys)}")
    expected_surface = official | SUPPLEMENTAL_SURFACE
    if surface_keys != expected_surface:
        missing = sorted(expected_surface - surface_keys)[:5]
        extra = sorted(surface_keys - expected_surface)[:5]
        problems.append(f"api_surface path set mismatch missing={missing} extra={extra}")
    if problems:
        print("FAIL GitLab inventory parity")
        for problem in problems:
            print(f"- {problem}")
        return 1
    print(
        f"PASS GitLab inventory parity: {len(official)} official operations plus {len(SUPPLEMENTAL_SURFACE)} supplemental stream row represented"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
