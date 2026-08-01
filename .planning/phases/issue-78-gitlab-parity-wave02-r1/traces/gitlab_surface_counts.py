#!/usr/bin/env python3
"""Local GitLab generated surface count check for issue #78."""
from __future__ import annotations

import json
import re
import sys
from collections import Counter
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[4]
DEF_DIR = ROOT / "internal/connectors/defs/gitlab"
EXPECTED = {
    "gitlab.etl_read": 397,
    "gitlab.reverse_etl_write": 637,
    "gitlab.direct_read_query_search": 6,
    "gitlab.binary_file": 89,
    "gitlab.cdc_changefeed": 15,
    "gitlab.excluded_not_applicable": 2,
}
WRITE_METHODS = {"POST", "PUT", "PATCH", "DELETE"}
TYPED_DESTRUCTIVE_APPROVAL = "typed_confirmation + plan_preview_approval_execute"
METADATA_REST_ROWS = {
    ("GET", "/groups/{id}/registry/repositories"),
    ("GET", "/projects/{id}/registry/repositories"),
    ("GET", "/groups/{id}/uploads"),
    ("GET", "/projects/{id}/uploads"),
    ("GET", "/projects/{id}/packages/{package_id}/package_files"),
    ("GET", "/group/{id}/-/packages/composer/packages"),
    ("GET", "/projects/{id}/packages/terraform/modules/{module_name}/{module_system}"),
    ("GET", "/projects/{id}/packages/terraform/modules/{module_name}/{module_system}/{module_version}"),
    ("GET", "/packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}/versions"),
    ("GET", "/packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}/download"),
    ("GET", "/packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}"),
    ("GET", "/packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}/{module_version}/download"),
    ("GET", "/packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}/{module_version}"),
}
TERRAFORM_MODULE_FILE_ROWS = {
    ("GET", "/packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}/{module_version}/file"),
}
SECRET_SENSITIVE_ROWS = {
    ("PUT", "/groups/{id}/integrations/campfire"),
    ("PUT", "/projects/{id}/integrations/campfire"),
    ("PUT", "/projects/{id}/services/campfire"),
    ("GET", "/projects/{id}/secure_files"),
    ("POST", "/projects/{id}/secure_files"),
    ("GET", "/projects/{id}/secure_files/{secure_file_id}"),
    ("GET", "/projects/{id}/secure_files/{secure_file_id}/download"),
}
GEO_BINARY_ROWS = {
    ("GET", "/geo/retrieve/{replicable_name}/{replicable_id}"),
}


def row_key(row: dict[str, Any]) -> tuple[str, str] | None:
    rest = row.get("rest")
    if isinstance(rest, dict):
        return str(rest.get("method")), str(rest.get("path"))
    binary = row.get("binary")
    if isinstance(binary, dict):
        return str(binary.get("method")), str(binary.get("path"))
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


def path_param_names(path: str) -> set[str]:
    return set(re.findall(r"\{([A-Za-z0-9_]+)\}", path))


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
    splat_api = sorted(path for _, path in api_keys if isinstance(path, str) and "*" in path)[:5]
    splat_ops = sorted(path for _, path in operation_keys if "*" in path)[:5]
    splat_cli = sorted(path for path in cli_api_surface_paths(cli) if "*" in path)[:5]
    if splat_api:
        problems.append(f"api_surface has unresolved splat paths {splat_api}")
    if splat_ops:
        problems.append(f"operations has unresolved splat paths {splat_ops}")
    if splat_cli:
        problems.append(f"cli_surface has unresolved splat paths {splat_cli}")
    missing_api = sorted(operation_keys - api_keys)[:5]
    if missing_api:
        problems.append(f"operations paths missing from api_surface {missing_api}")
    for command in cli.get("commands", []):
        if not isinstance(command, dict):
            continue
        names: set[str] = set()
        for row in command.get("api_surface", []):
            if isinstance(row, dict):
                names.update(path_param_names(str(row.get("path"))))
        for flag in command.get("flags", []):
            if not isinstance(flag, dict):
                continue
            maps_to = str(flag.get("maps_to", ""))
            if maps_to.startswith("query.") and maps_to.removeprefix("query.") in names:
                problems.append(f"{command.get('operation') or command.get('path')} path param mapped as {maps_to}")
                break
    for op in operations:
        key = row_key(op)
        if key is None:
            problems.append(f"{op.get('id')} has no REST/binary/composite key")
            continue
        method, _ = key
        audit_event = op.get("audit_event")
        command = cli_by_operation.get(op.get("id"), {})
        rest = op.get("rest")
        if isinstance(rest, dict):
            query = rest.get("query")
            if isinstance(query, dict):
                for query_key, query_value in query.items():
                    if str(query_key).startswith(("path.", "query.", "body.")) or "|required" in str(query_value) or "|optional" in str(query_value):
                        problems.append(f"{op.get('id')} rest.query contains parameter contract {query_key}={query_value}")
                        break
        if method in WRITE_METHODS and audit_event not in {"gitlab.reverse_etl_write", "gitlab.excluded_not_applicable"}:
            problems.append(f"{op.get('id')} write method classified as {audit_event}")
        if key in METADATA_REST_ROWS:
            if op.get("kind") != "rest_read" or audit_event == "gitlab.binary_file":
                problems.append(f"{op.get('id')} metadata row classified as {op.get('kind')}/{audit_event}")
            if op.get("output_policy") == "binary_file_bounded":
                problems.append(f"{op.get('id')} metadata row uses binary output policy")
        if key in TERRAFORM_MODULE_FILE_ROWS:
            if op.get("kind") != "binary_download" or audit_event != "gitlab.binary_file":
                problems.append(f"{op.get('id')} terraform module file row classified as {op.get('kind')}/{audit_event}")
        if key in SECRET_SENSITIVE_ROWS:
            if op.get("secret_sensitive") is not True:
                problems.append(f"{op.get('id')} expected secret_sensitive=true")
            if op.get("risk") != "high":
                problems.append(f"{op.get('id')} secret row risk={op.get('risk')!r}")
            if not isinstance(op.get("sensitive_policy"), dict):
                problems.append(f"{op.get('id')} missing sensitive_policy")
        if key in GEO_BINARY_ROWS:
            if op.get("kind") != "binary_download" or audit_event != "gitlab.binary_file":
                problems.append(f"{op.get('id')} geo retrieve row classified as {op.get('kind')}/{audit_event}")
        if op.get("secret_sensitive") is True:
            if op.get("risk") in {"low", "none"}:
                problems.append(f"{op.get('id')} secret-sensitive risk={op.get('risk')!r}")
            if method == "GET" and not command.get("risk"):
                problems.append(f"{op.get('id')} secret-sensitive CLI risk missing")
        if method == "HEAD":
            if audit_event != "gitlab.direct_read_query_search":
                problems.append(f"{op.get('id')} HEAD audit_event={audit_event}")
            if op.get("kind") != "composite":
                problems.append(f"{op.get('id')} HEAD kind={op.get('kind')}")
            if command.get("intent") != "direct_read":
                problems.append(f"{op.get('id')} HEAD cli intent={command.get('intent')!r}")
            if "Reverse ETL" in str(command.get("approval", "")) or "mutation" in str(command.get("risk", "")):
                problems.append(f"{op.get('id')} HEAD still has write safety text")
        if audit_event == "gitlab.binary_file":
            binary = op.get("binary")
            if op.get("kind") != "binary_download":
                problems.append(f"{op.get('id')} binary kind={op.get('kind')}")
            if not isinstance(binary, dict):
                problems.append(f"{op.get('id')} missing binary block")
            elif binary.get("method") != "GET" or not binary.get("path"):
                problems.append(f"{op.get('id')} invalid binary block {binary}")
            if op.get("rest") is not None:
                problems.append(f"{op.get('id')} binary row still has rest block")
            if op.get("output_policy") != "binary_file_bounded":
                problems.append(f"{op.get('id')} binary output_policy={op.get('output_policy')!r}")
            if command.get("output_policy") != "binary_file_bounded":
                problems.append(f"{op.get('id')} binary cli output_policy={command.get('output_policy')!r}")
        if method == "DELETE":
            if op.get("destructive") is not True:
                problems.append(f"{op.get('id')} DELETE missing destructive=true")
            if op.get("approval") != TYPED_DESTRUCTIVE_APPROVAL:
                problems.append(f"{op.get('id')} DELETE approval={op.get('approval')!r}")
            if audit_event != "gitlab.reverse_etl_write":
                problems.append(f"{op.get('id')} DELETE audit_event={audit_event}")
            if op.get("output_policy") in {"binary_bounded", "binary_file_bounded"}:
                problems.append(f"{op.get('id')} DELETE uses binary output policy")
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
