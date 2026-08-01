#!/usr/bin/env python3
"""Local GitLab generated surface count check for issue #78."""
from __future__ import annotations

import json
import sys
from collections import Counter
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[4]
DEF_DIR = ROOT / "internal/connectors/defs/gitlab"
EXPECTED = {
    "gitlab.etl_read": 308,
    "gitlab.reverse_etl_write": 640,
    "gitlab.direct_read_query_search": 3,
    "gitlab.binary_file": 178,
    "gitlab.cdc_changefeed": 15,
    "gitlab.excluded_not_applicable": 2,
}
WRITE_METHODS = {"POST", "PUT", "PATCH", "DELETE"}
TYPED_DESTRUCTIVE_APPROVAL = "typed_confirmation + plan_preview_approval_execute"


def row_key(row: dict[str, Any]) -> tuple[str, str] | None:
    rest = row.get("rest")
    if isinstance(rest, dict):
        return str(rest.get("method")), str(rest.get("path"))
    composite = row.get("composite")
    if isinstance(composite, dict):
        steps = composite.get("steps")
        if isinstance(steps, list) and steps:
            method, path, *_ = str(steps[0]).split(" ", 2)
            return method, path
    return None


def cli_api_surface_paths(cli: dict[str, Any]) -> list[str]:
    paths: list[str] = []
    for command in cli.get("commands", []):
        if not isinstance(command, dict):
            continue
        for row in command.get("api_surface", []):
            if isinstance(row, dict):
                paths.append(str(row.get("path")))
    return paths


def main() -> int:
    api = json.loads((DEF_DIR / "api_surface.json").read_text())
    operations = json.loads((DEF_DIR / "operations.json").read_text())["operations"]
    cli = json.loads((DEF_DIR / "cli_surface.json").read_text())
    api_rows = api.get("endpoints", [])
    counts = Counter(op["audit_event"] for op in operations)
    api_keys = {(row.get("method"), row.get("path")) for row in api_rows if isinstance(row, dict)}
    operation_keys = {key for op in operations if (key := row_key(op)) is not None}
    cli_by_operation = {command.get("operation"): command for command in cli.get("commands", []) if isinstance(command, dict)}
    problems: list[str] = []
    if len(api_rows) != 1147:
        problems.append(f"api_surface endpoints={len(api_rows)}, want 1147")
    if len(operations) != 1146:
        problems.append(f"operations rows={len(operations)}, want 1146")
    if len(cli.get("commands", [])) != 1147:
        problems.append(f"cli commands={len(cli.get('commands', []))}, want 1147")
    for key, want in EXPECTED.items():
        got = counts[key]
        if got != want:
            problems.append(f"{key}={got}, want {want}")
    covered_streams = sorted(
        row["covered_by"]["stream"]
        for row in api_rows
        if isinstance(row, dict) and row.get("covered_by", {}).get("stream")
    )
    if covered_streams != ["groups", "issues", "projects", "users"]:
        problems.append(f"covered_streams={covered_streams!r}")
    top_level_users = [row for row in api_rows if isinstance(row, dict) and (row.get("method"), row.get("path")) == ("GET", "/users")]
    if len(top_level_users) != 1 or top_level_users[0].get("covered_by", {}).get("stream") != "users":
        problems.append("GET /users is not the users stream coverage row")
    project_users = [row for row in api_rows if isinstance(row, dict) and (row.get("method"), row.get("path")) == ("GET", "/projects/{id}/users")]
    if len(project_users) != 1 or project_users[0].get("covered_by"):
        problems.append("GET /projects/{id}/users must remain planned, not stream-covered")
    prefixed_api = sorted(path for _, path in api_keys if isinstance(path, str) and path.startswith("/api/v4"))[:5]
    prefixed_ops = sorted(path for _, path in operation_keys if path.startswith("/api/v4"))[:5]
    prefixed_cli = sorted(path for path in cli_api_surface_paths(cli) if path.startswith("/api/v4"))[:5]
    if prefixed_api:
        problems.append(f"api_surface has /api/v4-prefixed paths {prefixed_api}")
    if prefixed_ops:
        problems.append(f"operations has /api/v4-prefixed paths {prefixed_ops}")
    if prefixed_cli:
        problems.append(f"cli_surface has /api/v4-prefixed paths {prefixed_cli}")
    missing_api = sorted(operation_keys - api_keys)[:5]
    if missing_api:
        problems.append(f"operations paths missing from api_surface {missing_api}")
    for op in operations:
        key = row_key(op)
        if key is None:
            problems.append(f"{op.get('id')} has no REST/composite key")
            continue
        method, _ = key
        audit_event = op.get("audit_event")
        if method in WRITE_METHODS and audit_event not in {"gitlab.reverse_etl_write", "gitlab.excluded_not_applicable"}:
            problems.append(f"{op.get('id')} write method classified as {audit_event}")
        if method == "DELETE":
            command = cli_by_operation.get(op.get("id"), {})
            if op.get("destructive") is not True:
                problems.append(f"{op.get('id')} DELETE missing destructive=true")
            if op.get("approval") != TYPED_DESTRUCTIVE_APPROVAL:
                problems.append(f"{op.get('id')} DELETE approval={op.get('approval')!r}")
            if audit_event != "gitlab.reverse_etl_write":
                problems.append(f"{op.get('id')} DELETE audit_event={audit_event}")
            if op.get("output_policy") == "binary_bounded":
                problems.append(f"{op.get('id')} DELETE still uses binary_bounded output")
            if command.get("intent") != "reverse_etl":
                problems.append(f"{op.get('id')} DELETE cli intent={command.get('intent')!r}")
    if problems:
        print("FAIL GitLab generated surface counts")
        for problem in problems[:50]:
            print(f"- {problem}")
        if len(problems) > 50:
            print(f"- ... {len(problems) - 50} more")
        return 1
    print("PASS GitLab generated surface counts")
    print(
        json.dumps(
            {
                "api_surface": len(api_rows),
                "operations": len(operations),
                "cli_commands": len(cli.get("commands", [])),
                "counts": dict(counts),
            },
            indent=2,
            sort_keys=True,
        )
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
