#!/usr/bin/env python3
"""Generate connector-local GitLab parity metadata for issue #78.

The generator consumes the pinned public GitLab OpenAPI v2 YAML and writes only
`internal/connectors/defs/gitlab/**` plus planning evidence. It intentionally
keeps unsupported operations blocked/planned instead of fabricating runtime
execution.
"""
from __future__ import annotations

import json
import re
import textwrap
import urllib.request
from collections import Counter
from copy import deepcopy
from pathlib import Path
from typing import Any

import yaml

ROOT = Path(__file__).resolve().parents[4]
DEF_DIR = ROOT / "internal/connectors/defs/gitlab"
OPENAPI_URL = "https://gitlab.com/gitlab-org/gitlab/-/raw/9cd04099eb59d87335798e4f57a2bc5a2622e4cc/doc/api/openapi/openapi_v2.yaml"
MASTER_BRANCH_URL = "https://gitlab.com/api/v4/projects/gitlab-org%2Fgitlab/repository/branches/master"
METHODS = {"get", "post", "put", "patch", "delete", "head", "options"}
PARENT_COUNTS = {
    "etl_read": 308,
    "reverse_etl_write": 498,
    "direct_read_query_search": 6,
    "binary_file": 298,
    "cdc_changefeed": 34,
    "excluded_not_applicable": 2,
}

EXCLUDED = {
    ("POST", "/api/v4/integrations/slack/interactions"),
    ("POST", "/api/v4/integrations/slack/options"),
}
DIRECT = {
    ("GET", "/api/v4/groups/{id}/(-/)search"),
    ("GET", "/api/v4/projects/{id}/(-/)search"),
    ("GET", "/api/v4/search"),
    ("POST", "/api/v4/geo/node_proxy/{id}/graphql"),
    ("POST", "/api/v4/glql"),
    ("POST", "/api/v4/markdown"),
}
# Four package metadata endpoints are durable JSON record reads, not binary payload transfers.
FORCE_ETL = {
    ("GET", "/api/v4/groups/{id}/packages"),
    ("GET", "/api/v4/projects/{id}/packages"),
    ("GET", "/api/v4/projects/{id}/packages/{package_id}"),
    ("GET", "/api/v4/projects/{id}/packages/{package_id}/pipelines"),
}
# Project export downloads are binary even though other project_import status/import operations remain writes/ETL.
FORCE_BINARY = {
    ("GET", "/api/v4/projects/{id}/export/download"),
    ("GET", "/api/v4/projects/{id}/export_relations/download"),
}
# Package deletions are safer to represent as named destructive reverse-ETL plans than binary operations.
FORCE_REVERSE = {
    ("DELETE", "/api/v4/projects/{id}/packages/{package_id}"),
    ("DELETE", "/api/v4/projects/{id}/packages/{package_id}/package_files/{package_file_id}"),
}
BINARY_TERMS = (
    "upload",
    "download",
    "artifact",
    "archive",
    "raw",
    "blobs",
    "packages",
    "package",
    "registry",
    "terraform",
    "dependency_proxy",
)
ADMIN_TERMS = (
    "admin",
    "application",
    "license",
    "runners",
    "runner",
    "hooks",
    "impersonation",
    "broadcast",
    "feature_flags",
    "settings",
)
SECRET_TERMS = (
    "secret",
    "token",
    "password",
    "credential",
    "variables",
    "keys",
    "deploy_key",
    "access_tokens",
)
STREAM_COVERAGE = {
    ("GET", "/api/v4/projects"): "projects",
    ("GET", "/api/v4/groups"): "groups",
    # The legacy fixture-backed stream reads the top-level /users API, which is
    # documented in GitLab REST docs but absent from this pinned OpenAPI source;
    # use the official project-users collection row to keep the stream represented
    # without adding a non-OpenAPI 1,147th api_surface row.
    ("GET", "/api/v4/projects/{id}/users"): "users",
    ("GET", "/api/v4/issues"): "issues",
}


def load_openapi() -> dict[str, Any]:
    with urllib.request.urlopen(OPENAPI_URL, timeout=60) as response:
        return yaml.safe_load(response.read())


def dump_json(path: Path, data: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, indent=2, ensure_ascii=False) + "\n")


def clean_text(value: Any, fallback: str) -> str:
    if not isinstance(value, str) or not value.strip():
        return fallback
    value = re.sub(r"\s+", " ", value.strip())
    return value[:280]


def operation_id(method: str, path: str, used: set[str]) -> str:
    normalized = path.lower()
    normalized = normalized.replace("/api/v4", "api-v4", 1)
    normalized = re.sub(r"[^a-z0-9]+", "-", normalized).strip("-")
    base = f"gitlab.{method.lower()}-{normalized}"
    if not re.match(r"^[a-z0-9][a-z0-9._-]*$", base):
        base = "gitlab." + re.sub(r"[^a-z0-9._-]+", "-", base).strip(".-_")
    candidate = base
    i = 2
    while candidate in used:
        candidate = f"{base}-{i}"
        i += 1
    used.add(candidate)
    return candidate


def relative_path(path: str) -> str:
    if path.startswith("/api/v4"):
        stripped = path[len("/api/v4") :]
        return stripped or "/"
    return path


def op_blob(op: dict[str, Any]) -> str:
    return " ".join(
        [
            op["path"],
            op.get("operationId") or "",
            op.get("summary") or "",
            " ".join(op.get("tags") or []),
        ]
    ).lower()


def lane_for(op: dict[str, Any]) -> str:
    key = (op["method"], op["path"])
    if key in EXCLUDED:
        return "excluded_not_applicable"
    if key in DIRECT:
        return "direct_read_query_search"
    blob = op_blob(op)
    if "event" in blob or "hook" in blob:
        if not op["path"].startswith("/api/v4/usage_data/") and key != (
            "GET",
            "/api/v4/groups/{id}/audit_events/{audit_event_id}",
        ):
            return "cdc_changefeed"
    if key in FORCE_ETL:
        return "etl_read"
    if key in FORCE_REVERSE:
        return "reverse_etl_write"
    if key in FORCE_BINARY:
        return "binary_file"
    first_tag = (op.get("tags") or [""])[0]
    produces = op.get("produces") or ()
    consumes = op.get("consumes") or ()
    non_json = any(("json" not in item and "*/*" not in item) for item in tuple(produces) + tuple(consumes))
    if first_tag != "project_import" and (non_json or any(term in blob for term in BINARY_TERMS)):
        return "binary_file"
    if op["method"] == "GET":
        return "etl_read"
    return "reverse_etl_write"


def is_admin(op: dict[str, Any]) -> bool:
    blob = op_blob(op)
    return any(term in blob for term in ADMIN_TERMS)


def is_secret_sensitive(op: dict[str, Any]) -> bool:
    blob = op_blob(op)
    return any(term in blob for term in SECRET_TERMS)


def is_destructive(op: dict[str, Any], lane: str) -> bool:
    blob = op_blob(op)
    return lane == "reverse_etl_write" and (
        op["method"] == "DELETE"
        or any(term in blob for term in ("delete", "remove", "destroy", "purge", "revoke", "reset", "erase"))
    )


def risk_for(op: dict[str, Any], lane: str) -> str:
    if lane == "excluded_not_applicable":
        return "none"
    if lane == "etl_read":
        return "low"
    if lane in {"direct_read_query_search", "binary_file", "cdc_changefeed"}:
        return "medium"
    if is_destructive(op, lane):
        return "critical" if is_admin(op) else "high"
    if is_admin(op) or is_secret_sensitive(op):
        return "high"
    return "medium"


def is_write_method(op: dict[str, Any]) -> bool:
    return op["method"] not in {"GET", "HEAD"}


def mutation_class_for(op: dict[str, Any], lane: str) -> str:
    if lane in {"excluded_not_applicable", "direct_read_query_search"} or not is_write_method(op):
        return "none"
    if op["method"] == "DELETE":
        return "delete"
    if is_admin(op):
        return "admin"
    if is_secret_sensitive(op):
        return "secret"
    if op["method"] == "POST":
        return "create"
    if op["method"] in {"PUT", "PATCH"}:
        return "update"
    return "update"


def operation_model_for(op: dict[str, Any], lane: str) -> str:
    if lane == "binary_file":
        return "binary_read"
    if lane in {"etl_read", "direct_read_query_search", "cdc_changefeed"}:
        return "direct_read"
    if is_destructive(op, lane):
        return "destructive_action"
    if is_admin(op):
        return "admin_reverse_etl"
    return "sensitive_reverse_etl"


def json_type_for_param(param: dict[str, Any]) -> dict[str, Any]:
    typ = param.get("type", "string")
    schema: dict[str, Any] = {"type": typ}
    if "format" in param:
        schema["format"] = param["format"]
    if "enum" in param:
        schema["enum"] = param["enum"]
    if typ == "array":
        schema["items"] = param.get("items") or {"type": "string"}
    return schema


def resolve_ref(ref: str, spec: dict[str, Any]) -> Any:
    if not ref.startswith("#/"):
        return {"$ref": ref}
    cur: Any = spec
    for part in ref[2:].split("/"):
        cur = cur[part]
    return cur


def clean_schema(schema: Any, spec: dict[str, Any], depth: int = 0) -> Any:
    if depth > 3:
        return {"type": "object"}
    if not isinstance(schema, dict):
        return {"type": "object"}
    if "$ref" in schema:
        return clean_schema(resolve_ref(str(schema["$ref"]), spec), spec, depth + 1)
    out: dict[str, Any] = {}
    for key in ("type", "format", "enum", "required", "additionalProperties", "minimum", "maximum", "minLength", "maxLength"):
        if key in schema:
            out[key] = deepcopy(schema[key])
    if "items" in schema:
        out["items"] = clean_schema(schema["items"], spec, depth + 1)
    if "properties" in schema and isinstance(schema["properties"], dict):
        props = {}
        for name, child in sorted(schema["properties"].items()):
            props[name] = clean_schema(child, spec, depth + 1)
        out["properties"] = props
    if "description" in schema and isinstance(schema["description"], str):
        out["description"] = clean_text(schema["description"], "")
    return out or {"type": "object"}


def parameter_contract(op: dict[str, Any], spec: dict[str, Any]) -> tuple[dict[str, str], dict[str, Any] | None]:
    query_contract: dict[str, str] = {}
    body_schema: dict[str, Any] | None = None
    form_props: dict[str, Any] = {}
    form_required: list[str] = []
    for param in op.get("parameters") or []:
        if not isinstance(param, dict):
            continue
        location = param.get("in")
        name = param.get("name")
        if location in {"path", "query"} and name:
            schema = json_type_for_param(param)
            typ = schema.get("type", "string")
            required = "required" if location == "path" or param.get("required") else "optional"
            query_contract[f"{location}.{name}"] = f"{typ}|{required}"
        elif location == "body":
            body_schema = clean_schema(param.get("schema") or {"type": "object"}, spec)
        elif location == "formData" and name:
            form_props[name] = json_type_for_param(param)
            if param.get("required"):
                form_required.append(name)
    if form_props:
        body_schema = {"type": "object", "properties": form_props}
        if form_required:
            body_schema["required"] = sorted(form_required)
    return dict(sorted(query_contract.items())), body_schema


def collect_operations(spec: dict[str, Any]) -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    for path, item in spec.get("paths", {}).items():
        if not isinstance(item, dict):
            continue
        common_params = item.get("parameters") or []
        for method, operation in item.items():
            if method.lower() not in METHODS or not isinstance(operation, dict):
                continue
            op = deepcopy(operation)
            op["method"] = method.upper()
            op["path"] = path
            op["parameters"] = common_params + (op.get("parameters") or [])
            op["produces"] = tuple(op.get("produces") or item.get("produces") or spec.get("produces") or [])
            op["consumes"] = tuple(op.get("consumes") or item.get("consumes") or spec.get("consumes") or [])
            rows.append(op)
    # Preserve official path order while making repeated runs deterministic.
    return rows


def api_surface_row(op: dict[str, Any], lane: str, op_id: str) -> dict[str, Any]:
    key = (op["method"], op["path"])
    row: dict[str, Any] = {"method": op["method"], "path": op["path"]}
    if key in STREAM_COVERAGE:
        row["covered_by"] = {"stream": STREAM_COVERAGE[key]}
        return row
    source_url = OPENAPI_URL
    if lane == "excluded_not_applicable":
        row["operation"] = {
            "model": "disallowed",
            "status": "blocked",
            "risk": "low",
            "blocked_by_default": True,
            "reason": "GitLab Slack integration callback endpoint consumed by GitLab internals; not a user-invoked connector ETL, direct-read, binary, CDC, or reverse-ETL operation.",
            "source_url": OPENAPI_URL,
            "notes": "Counted as excluded/not-applicable in the parent ledger.",
        }
        return row
    risk = risk_for(op, lane)
    row["operation"] = {
        "model": operation_model_for(op, lane),
        "status": "blocked",
        "risk": "critical" if risk == "critical" else ("high" if risk == "high" else ("medium" if risk == "medium" else "low")),
        "blocked_by_default": True,
        "reason": f"planned GitLab {lane} operation; connector-local typed metadata exists as {op_id}, but no verified executable stream/action/command is claimed in this wave",
        "source_url": source_url,
        "notes": "DELETE/destructive/admin operations stay in scope; execution requires a named action with destructive confirmation plus plan -> preview -> explicit approval -> execute evidence.",
    }
    return row


def operation_row(op: dict[str, Any], lane: str, op_id: str, spec: dict[str, Any]) -> dict[str, Any]:
    summary = clean_text(op.get("summary"), op.get("operationId") or f"{op['method']} {op['path']}")
    risk = risk_for(op, lane)
    destructive = is_destructive(op, lane)
    admin = is_admin(op)
    secret_sensitive = is_secret_sensitive(op)
    if lane == "excluded_not_applicable":
        approval = "not_applicable"
    elif is_write_method(op):
        approval = "typed_confirmation + plan_preview_approval_execute" if destructive or admin else "plan_preview_approval_execute"
    else:
        approval = "none; planned bounded connector command/stream required before execution"
    output_policy = "binary_bounded" if lane == "binary_file" else "json_redacted"
    if lane == "excluded_not_applicable" or op["method"] == "HEAD":
        kind = "composite"
    elif is_write_method(op) and not (lane == "direct_read_query_search" and op["method"] == "POST"):
        kind = "rest_write"
    else:
        kind = "rest_read"
    query_contract, body_schema = parameter_contract(op, spec)
    rest: dict[str, Any] = {
        "method": op["method"],
        "path": relative_path(op["path"]),
        "max_bytes": 16 * 1024 * 1024 if lane == "binary_file" else (10 * 1024 * 1024 if lane in {"direct_read_query_search", "cdc_changefeed"} else 1024 * 1024),
    }
    consumes = list(op.get("consumes") or [])
    if consumes:
        rest["content_type"] = consumes[0]
    elif body_schema is not None or (kind == "rest_read" and op["method"] == "POST"):
        rest["content_type"] = "application/json"
    if query_contract:
        rest["query"] = query_contract
    if body_schema is not None:
        rest["body_schema"] = body_schema
    elif kind == "rest_read" and op["method"] == "POST":
        rest["body_schema"] = {"type": "object"}
    row: dict[str, Any] = {
        "id": op_id,
        "kind": kind,
        "summary": summary,
        "description": f"GitLab {lane} operation from the pinned OpenAPI v2 source. This row is typed connector-local metadata; it is not executable until a verified stream/action/command covers it.",
        "source_url": OPENAPI_URL,
        "risk": risk,
        "approval": approval,
        "output_policy": output_policy,
        "auth_scopes": ["GitLab token scopes per official operation and target resource permissions"],
        "audit_event": f"gitlab.{lane}",
        "mutation_class": mutation_class_for(op, lane),
    }
    if kind == "composite":
        row["composite"] = {"steps": [f"{op['method']} {relative_path(op['path'])} planned metadata only"]}
    else:
        row["rest"] = rest
    if destructive:
        row["destructive"] = True
    if secret_sensitive:
        row["secret_sensitive"] = True
        row["sensitive_policy"] = {
            "input_mode": "env_or_file",
            "redact_fields": ["value", "token", "secret", "key", "password"],
            "preflight": "secret-shaped inputs must be supplied outside prompt text and redacted from previews/errors",
            "transform": "none",
            "approval_mode": "typed_confirmation",
        }
    return row


def cli_command_for_operation(op: dict[str, Any], lane: str, op_id: str) -> dict[str, Any]:
    method = op["method"].lower()
    slug = re.sub(r"[^a-z0-9]+", "-", relative_path(op["path"]).lower()).strip("-") or "root"
    if lane == "etl_read":
        intent = "etl"
        availability = "implemented" if (op["method"], op["path"]) in STREAM_COVERAGE else "planned"
    elif lane == "reverse_etl_write":
        intent = "reverse_etl"
        availability = "planned"
    elif lane in {"direct_read_query_search", "binary_file", "cdc_changefeed"}:
        intent = "direct_read"
        availability = "planned"
    else:
        intent = "docs_only"
        availability = "excluded"
    cmd: dict[str, Any] = {
        "path": f"operation {method} {slug}",
        "summary": clean_text(op.get("summary"), f"{op['method']} {op['path']}"),
        "intent": intent,
        "availability": availability,
        "operation": op_id,
        "source_url": OPENAPI_URL,
        "notes": f"Typed metadata for GitLab {lane}; not a generic raw API escape hatch. Execution stays blocked until a connector-local stream/action/command has verification evidence.",
    }
    key = (op["method"], op["path"])
    if key in STREAM_COVERAGE:
        stream = STREAM_COVERAGE[key]
        cmd.pop("operation")
        cmd["path"] = f"{stream} list"
        cmd["summary"] = f"List GitLab {stream} as ETL records."
        cmd["stream"] = stream
        cmd["api_surface"] = [{"method": op["method"], "path": op["path"]}]
        cmd["notes"] = "Implemented fixture-backed GitLab ETL stream; use --limit for bounded output and saved credentials/config for live reads."
        cmd["examples"] = [f"pm gitlab {stream} list --json --limit 25"]
    elif lane != "excluded_not_applicable":
        cmd["api_surface"] = [{"method": op["method"], "path": op["path"]}]
    if intent == "reverse_etl":
        cmd["risk"] = "Planned GitLab mutation; execution requires a named reverse-ETL action with typed schema, redaction, and safety evidence."
        cmd["approval"] = "Reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive/admin writes also require typed confirmation."
    if intent == "direct_read":
        cmd["output_policy"] = "json_redacted"
        cmd["risk"] = "Planned bounded read/query/binary metadata; no live provider call is made by this metadata."
    return cmd


def build_cli_surface(ops: list[dict[str, Any]], lanes: dict[tuple[str, str], str], ids: dict[tuple[str, str], str]) -> dict[str, Any]:
    commands = [cli_command_for_operation(op, lanes[(op["method"], op["path"])], ids[(op["method"], op["path"])]) for op in ops]
    return {
        "tagline": "Inspect GitLab streams and planned typed operations without exposing raw API escape hatches.",
        "usage": "pm gitlab <command> [flags]",
        "source_cli": {
            "name": "glab / GitLab API",
            "docs": "https://docs.gitlab.com/ee/api/rest/",
            "reference": OPENAPI_URL,
            "source": "provider_api",
        },
        "groups": [
            {"id": "streams", "title": "Implemented fixture-backed streams", "commands": ["projects", "groups", "users", "issues"]},
            {"id": "operations", "title": "Planned typed operation metadata", "commands": ["operation"]},
        ],
        "global_flags": [
            {"name": "credential", "type": "string", "summary": "Credential name to use for a future GitLab command."},
            {"name": "connection", "type": "string", "summary": "Alias for --credential."},
            {"name": "config", "type": "string_array", "summary": "Connector config override as key=value."},
            {"name": "json", "type": "boolean", "summary": "Emit machine-readable JSON output."},
            {"name": "limit", "type": "integer", "summary": "Maximum PM ETL records to emit; does not control GitLab provider-side totals."},
            {"name": "max-bytes", "type": "integer", "summary": "Maximum future direct-read or binary response bytes; planned operations declare bounded metadata."},
            {"name": "preview", "type": "boolean", "summary": "Preview a future reverse-ETL write command without making a network mutation."},
            {"name": "approve", "type": "string", "summary": "Approval token required to execute a future reverse-ETL plan."},
            {"name": "confirm", "type": "string", "summary": "Typed confirmation challenge required for future destructive reverse-ETL writes."},
        ],
        "commands": commands,
        "help_topics": [
            {"name": "gitlab", "summary": "GitLab connector streams and planned typed operation metadata."},
            {"name": "gitlab safety", "summary": "GitLab write, destructive, direct-read, binary, and changefeed safety gates."},
        ],
    }


def build_docs(counts: Counter[str]) -> str:
    lines = [
        "# Overview",
        "",
        "Reads GitLab projects, groups, users, and issues through the GitLab REST API v4 and carries a complete connector-local operation ledger for the pinned official GitLab OpenAPI v2 source.",
        "",
        "Implemented fixture-backed streams: `projects`, `groups`, `users`, `issues`.",
        "",
        "The complete official ledger has 1,146 operations: 308 ETL/read, 498 reverse-ETL write/mutation, 6 direct/provider query/search, 298 binary/file, 34 CDC/changefeed/audit/webhook, and 2 excluded/not-applicable callback endpoints.",
        "",
        "Only the four streams are executable in this wave. `api_surface.json`, `operations.json`, and `cli_surface.json` keep every other operation represented as typed planned/blocked metadata until a future connector-local stream/action/command adds fixtures and execution evidence.",
        "",
        f"Official source: {OPENAPI_URL}",
        f"Branch provenance source: {MASTER_BRANCH_URL}",
        "",
        "## Auth setup",
        "",
        "Connection fields:",
        "",
        "- `access_token` (required, secret, string); GitLab personal access token or OAuth access token. Used only for Bearer auth and redacted by the connector runtime.",
        "- `base_url` (optional, string); default `https://gitlab.com/api/v4`; format `uri`; use `https://gitlab.example.com/api/v4` for self-managed instances or fixture replay.",
        "- `start_date` (optional, string); format `date-time`; used by the current read stream filters where upstream supports a date bound.",
        "- `page_size` (optional, string); default `50`; current streams send `per_page=50`.",
        "- `mode` (optional, string); fixture/live mode marker used by local harnesses.",
        "",
        "Requests use Bearer authentication from `secrets.access_token`. No fixture or metadata file contains credential values.",
        "",
        "Connection checks call GET `/user` against the configured API base URL.",
        "",
        "## Streams notes",
        "",
        "Default pagination follows RFC 5988 `Link` headers with `rel=next`; fixture pages are bounded and sanitized.",
        "",
        "- `projects`: GET `/projects`; records at the response root; sends `per_page=50`; optionally sends `last_activity_after` from `start_date`.",
        "- `groups`: GET `/groups`; records at the response root; sends `per_page=50`.",
        "- `users`: GET `/users`; records at the response root; sends `per_page=50`; optionally sends `created_after` from `start_date`. The pinned OpenAPI source does not enumerate this top-level users collection, so `api_surface.json` keeps the stream represented through the official project-users collection row without adding a non-OpenAPI operation.",
        "- `issues`: GET `/issues`; records at the response root; sends `per_page=50`; optionally sends `updated_after` from `start_date`; derives `author_id` from `author.id`.",
        "",
        "Planned ETL/read rows in `operations.json` are metadata only. They are not advertised as executable streams until each has a schema, fixture replay, and conformance evidence.",
        "",
        "## Write actions & risks",
        "",
        "No `writes.json` actions are executable in this wave, and connector metadata keeps `capabilities.write=false` until named actions and fixtures are added.",
        "",
        "The official ledger still includes all 498 mutation operations, including DELETE, destructive, admin, token/key/variable, hook, runner, package delete, and other high-risk operations. These are not blanket-excluded as unsafe. They are represented as planned/blocked typed metadata with risk, source URL, bounded request schemas where available, and approval notes.",
        "",
        "Before any GitLab mutation can execute it must become a named reverse-ETL action with:",
        "",
        "1. a bounded record schema and redaction contract;",
        "2. dry-run plan and preview evidence;",
        "3. explicit approval token;",
        "4. `confirm: \"destructive\"` and typed confirmation for destructive/admin actions;",
        "5. idempotency and cleanup notes where upstream behavior allows them;",
        "6. fixture/conformance evidence proving request shape without live credentials.",
        "",
        "No generic HTTP method/path/body, arbitrary GraphQL, shell, file, SQL write/read, extension, binary, or raw passthrough command is exposed.",
        "",
        "## Known limits",
        "",
        "- Fixture-backed implementation remains limited to 4 streams; all other official operations are planned/blocked metadata.",
        "- This connector is not live-certified; fixture success must not be reported as provider certification.",
        "- Direct/provider search/query rows depend on shared foundation #2985 before execution can be claimed.",
        "- Binary/file transfer rows depend on shared foundation #2987 before bounded download/upload execution can be claimed.",
        "- CDC/changefeed/audit/webhook rows depend on shared foundations #2986 and #2988 before CDC/changefeed claims can be made.",
        "- Destructive/admin write rows depend on per-action schemas, redaction, fixtures, and typed confirmation evidence before execution can be claimed.",
        "- The current top-level `/users` stream is fixture-backed legacy behavior; the pinned OpenAPI source omits that exact row, so this wave records the shared documentation/source mismatch and does not claim new user-stream parity beyond existing fixtures.",
        "- The two excluded rows are GitLab Slack integration callback endpoints (`/integrations/slack/interactions` and `/integrations/slack/options`), not user-invoked connector operations.",
        f"- Generated lane counts: etl_read={counts['etl_read']}, reverse_etl_write={counts['reverse_etl_write']}, direct_read_query_search={counts['direct_read_query_search']}, binary_file={counts['binary_file']}, cdc_changefeed={counts['cdc_changefeed']}, excluded_not_applicable={counts['excluded_not_applicable']}.",
    ]
    return "\n".join(lines) + "\n"


def main() -> None:
    spec = load_openapi()
    ops = collect_operations(spec)
    used_ids: set[str] = set()
    ids: dict[tuple[str, str], str] = {}
    lanes: dict[tuple[str, str], str] = {}
    api_rows = []
    operation_rows = []
    counts: Counter[str] = Counter()
    for op in ops:
        key = (op["method"], op["path"])
        op_id = operation_id(op["method"], op["path"], used_ids)
        lane = lane_for(op)
        ids[key] = op_id
        lanes[key] = lane
        counts[lane] += 1
        api_rows.append(api_surface_row(op, lane, op_id))
        operation_rows.append(operation_row(op, lane, op_id, spec))
    if len(ops) != 1146:
        raise SystemExit(f"expected 1146 official operations, got {len(ops)}")
    if dict(counts) != PARENT_COUNTS:
        raise SystemExit(f"lane counts {dict(counts)} do not match parent counts {PARENT_COUNTS}")
    metadata = {
        "name": "gitlab",
        "display_name": "GitLab",
        "description": "Reads GitLab projects, groups, users, and issues; carries a complete planned official GitLab operation ledger for future typed streams, writes, direct reads, binary transfers, and changefeeds.",
        "integration_type": "api",
        "docs_url": "https://docs.gitlab.com/ee/api/rest/",
        "release_stage": "ga",
        "capabilities": {"check": True, "read": True, "write": False, "query": False, "cdc": False, "dynamic_schema": False},
        "batch": {"read_page_size": 50, "write_batch_size": 1},
        "rate_limit": {"requests_per_minute": 600},
        "risk": {
            "read": "external GitLab API reads for the four implemented streams; planned reads remain metadata-only until fixture-backed",
            "write": "no executable write actions in this wave; planned mutations require reverse ETL plan, preview, explicit approval, and destructive typed confirmation where applicable",
            "approval": "destructive/admin operations are in scope but blocked until connector-local named actions provide confirm=destructive and verified safety evidence",
        },
    }
    api_surface = {
        "api": "GitLab REST API v4 (pinned OpenAPI v2 source)",
        "docs": "https://docs.gitlab.com/ee/api/rest/",
        "reviewed_at": "2026-07-31",
        "operation_ledger_version": 1,
        "scope": "Complete official GitLab OpenAPI v2 ledger at gitlab-org/gitlab commit 9cd04099eb59d87335798e4f57a2bc5a2622e4cc. Four existing streams are executable; all other official operations are planned/blocked typed metadata until connector-local schemas, fixtures, commands, and safety evidence make them executable. DELETE/destructive/admin operations are included and require typed destructive confirmation plus plan -> preview -> explicit approval -> execute before execution.",
        "endpoints": api_rows,
    }
    operations = {"operations": operation_rows}
    cli_surface = build_cli_surface(ops, lanes, ids)
    certification = {
        "schema_version": 1,
        "source": {
            "default_stream": "projects",
            "source_credential_defaults": {"base_url": "https://gitlab.com/api/v4"},
            "live_unavailable": [
                {
                    "kind": "Error",
                    "contains": [
                        "http 401",
                        "http 403",
                        "http 404",
                        "status 401",
                        "status 403",
                        "status 404",
                        "token lacks access",
                        "interpolate: unresolved key",
                    ],
                }
            ],
        },
        "direct_read_candidates": [],
        "binary_candidates": [],
        "write_pairings": [],
    }
    dump_json(DEF_DIR / "metadata.json", metadata)
    dump_json(DEF_DIR / "api_surface.json", api_surface)
    dump_json(DEF_DIR / "operations.json", operations)
    dump_json(DEF_DIR / "cli_surface.json", cli_surface)
    dump_json(DEF_DIR / "certification.json", certification)
    (DEF_DIR / "docs.md").write_text(build_docs(counts))
    summary = {
        "official_operations": len(ops),
        "lane_counts": dict(counts),
        "covered_streams": sorted(STREAM_COVERAGE.values()),
        "blocked_or_planned": len(ops) - len(STREAM_COVERAGE) - counts["excluded_not_applicable"],
        "excluded_not_applicable": counts["excluded_not_applicable"],
        "source": OPENAPI_URL,
    }
    dump_json(ROOT / ".planning/phases/issue-78-gitlab-parity-wave02-r1/traces/gitlab_generation_summary.json", summary)
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
