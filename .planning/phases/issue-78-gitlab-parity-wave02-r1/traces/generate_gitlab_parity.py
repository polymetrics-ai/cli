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
    "etl_read": 397,
    "reverse_etl_write": 637,
    "direct_read_query_search": 6,
    "binary_file": 89,
    "cdc_changefeed": 15,
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
}
FORCE_ETL = {
    ("GET", "/api/v4/groups/{id}/packages"),
    ("GET", "/api/v4/projects/{id}/packages"),
    ("GET", "/api/v4/projects/{id}/packages/{package_id}"),
    ("GET", "/api/v4/projects/{id}/packages/{package_id}/pipelines"),
    ("GET", "/api/v4/projects/{id}/packages/terraform/modules/{module_name}/{module_system}"),
    ("GET", "/api/v4/projects/{id}/packages/terraform/modules/{module_name}/{module_system}/*module_version"),
    ("GET", "/api/v4/packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}/download"),
    ("GET", "/api/v4/packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}"),
    ("GET", "/api/v4/packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}/*module_version/download"),
    ("GET", "/api/v4/packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}/*module_version"),
}
FORCE_BINARY = {
    ("GET", "/api/v4/geo/retrieve/{replicable_name}/{replicable_id}"),
    ("GET", "/api/v4/projects/{id}/export/download"),
    ("GET", "/api/v4/projects/{id}/export_relations/download"),
    ("GET", "/api/v4/projects/{id}/terraform/state/{name}"),
    ("GET", "/api/v4/projects/{id}/terraform/state/{name}/versions/{serial}"),
}
# Package deletions are safer to represent as named destructive reverse-ETL plans than binary operations.
FORCE_REVERSE = {
    ("POST", "/api/v4/geo/node_proxy/{id}/graphql"),
    ("POST", "/api/v4/glql"),
    ("POST", "/api/v4/markdown"),
    ("DELETE", "/api/v4/projects/{id}/packages/{package_id}"),
    ("DELETE", "/api/v4/projects/{id}/packages/{package_id}/package_files/{package_file_id}"),
}
BINARY_TERMS = (
    "download",
    "artifact",
    "archive",
    "raw",
    "blobs",
    "dependency_proxy",
)
METADATA_READ_SUMMARY_PATTERNS = tuple(
    re.compile(pattern)
    for pattern in (
        r"^list\b",
        r"^get all\b",
        r"^retrieve details\b",
        r"\bmetadata endpoint\b",
        r"\bmetadata service\b",
        r"\bfeed service index\b",
        r"\bservice index\b",
        r"\bfeed enumerate\b",
        r"\bfind packages\b",
        r"\bsearch\b",
        r"^verify\b",
        r"^check\b",
        r"\bauthenticate\b",
        r"\bavailability\b",
        r"\bget all tags\b",
        r"\ball tags\b",
        r"\bpackage revisions\b",
        r"\brecipe revision\b",
        r"\blatest recipe\b",
    )
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
    "certificate",
    "certificates",
    "variables",
    "keys",
    "deploy_key",
    "access_key",
    "api_key",
    "private_key",
    "service_account_key",
    "access_tokens",
    "secure_files",
    "secure file",
    "webhook token",
    "integration token",
)
SECRET_TEXT_PATTERN = re.compile(
    r"(^|[^a-z0-9])"
    r"(secrets?|tokens?|passwords?|credentials?|certificates?|private[_ -]?tokens?|"
    r"webhook[_ -]?tokens?|integration[_ -]?tokens?|api[_ -]?keys?|access[_ -]?keys?|"
    r"private[_ -]?keys?|service[_ -]?account[_ -]?keys?|deploy[_ -]?keys?|secure[_ -]?files?)"
    r"([^a-z0-9]|$)"
)
STREAM_COVERAGE = {
    ("GET", "/api/v4/projects"): "projects",
    ("GET", "/api/v4/groups"): "groups",
    ("GET", "/api/v4/issues"): "issues",
}
SUPPLEMENTAL_STREAM_ROWS = [
    {
        "method": "GET",
        "path": "/users",
        "covered_by": {"stream": "users"},
        "source_url": "https://docs.gitlab.com/ee/api/users/#list-users",
        "notes": "Existing fixture-backed top-level users stream; absent from the pinned OpenAPI v2 source.",
    }
]


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


def normalize_gitlab_path(path: str) -> str:
    path = re.sub(r"\(\*([A-Za-z0-9_]+)/\)", r"{\1}/", path)
    return re.sub(r"\*([A-Za-z0-9_]+)", r"{\1}", path)


def relative_path(path: str) -> str:
    path = normalize_gitlab_path(path)
    if path.startswith("/api/v4"):
        stripped = path[len("/api/v4") :]
        return stripped or "/"
    return path


def path_param_names(path: str) -> set[str]:
    return {match.group(1) for match in re.finditer(r"\{([A-Za-z0-9_]+)\}", normalize_gitlab_path(path))}


def op_blob(op: dict[str, Any]) -> str:
    return " ".join(
        [
            op["path"],
            op.get("operationId") or "",
            op.get("summary") or "",
            " ".join(op.get("tags") or []),
        ]
    ).lower()


def path_has_file_payload(path: str, summary: str) -> bool:
    normalized = path.lower()
    lowered_summary = summary.lower()
    if normalized.endswith("/artifacts/tree") or "/artifacts/tree/" in normalized:
        return False
    if re.search(r"/(download|raw|archive|snapshot)(/|$)", normalized):
        return True
    if "/blobs/" in normalized or "/artifact" in normalized or "/artifacts/" in normalized:
        return True
    if "/uploads/" in normalized and re.search(r"\{(upload_id|secret|filename)\}", normalized):
        return True
    if "/packages/debian/pool/" in normalized or "/packages/rpm/repodata/" in normalized:
        return True
    if "/packages/debian/dists/" in normalized and re.search(r"/(release|inrelease|packages|sources)(?:$|[/{*])", normalized):
        return True
    if re.search(r"\.(tgz|zip|tar|gz|gem|whl|jar|pom|sha1|md5|sha256|module|rpm|deb|apk|gpg|yaml|yml|asc)(?:$|[/{*])", normalized):
        return True
    return bool(re.search(r"\bretrieve (?:a |the |specific )?(?:recipe|package) file\b", lowered_summary))


def summary_marks_metadata_read(summary: str) -> bool:
    lowered = summary.lower()
    return any(pattern.search(lowered) for pattern in METADATA_READ_SUMMARY_PATTERNS)


def is_binary_payload(op: dict[str, Any]) -> bool:
    key = (op["method"], op["path"])
    if key in FORCE_ETL:
        return False
    if key in FORCE_BINARY:
        return True
    if op["method"] != "GET":
        return False
    if path_has_file_payload(op["path"], op.get("summary") or ""):
        return True
    if summary_marks_metadata_read(op.get("summary") or ""):
        return False
    first_tag = (op.get("tags") or [""])[0]
    produces = op.get("produces") or ()
    consumes = op.get("consumes") or ()
    non_json = any(("json" not in item and "*/*" not in item) for item in tuple(produces) + tuple(consumes))
    if first_tag != "project_import" and non_json:
        return True
    blob = op_blob(op)
    return any(term in blob for term in BINARY_TERMS)


def lane_for(op: dict[str, Any]) -> str:
    key = (op["method"], op["path"])
    if key in EXCLUDED:
        return "excluded_not_applicable"
    if key in FORCE_REVERSE or is_write_method(op):
        return "reverse_etl_write"
    if key in DIRECT:
        return "direct_read_query_search"
    if op["method"] == "HEAD":
        return "direct_read_query_search"
    blob = op_blob(op)
    if "event" in blob or "hook" in blob:
        if not op["path"].startswith("/api/v4/usage_data/") and key != (
            "GET",
            "/api/v4/groups/{id}/audit_events/{audit_event_id}",
        ):
            return "cdc_changefeed"
    if is_binary_payload(op):
        return "binary_file"
    if op["method"] == "GET":
        return "etl_read"
    return "reverse_etl_write"


def is_admin(op: dict[str, Any]) -> bool:
    blob = op_blob(op)
    return any(term in blob for term in ADMIN_TERMS)


def text_marks_secret(value: Any) -> bool:
    if not isinstance(value, str):
        return False
    lowered = value.lower()
    if any(term in lowered for term in SECRET_TERMS):
        return True
    normalized = re.sub(r"[_/.-]+", " ", lowered)
    return bool(SECRET_TEXT_PATTERN.search(normalized))


def schema_marks_secret(value: Any, spec: dict[str, Any] | None = None, seen: set[str] | None = None) -> bool:
    if isinstance(value, dict):
        ref = value.get("$ref")
        if spec is not None and isinstance(ref, str):
            seen = seen or set()
            if ref in seen:
                return False
            try:
                return schema_marks_secret(resolve_ref(ref, spec), spec, seen | {ref})
            except KeyError:
                return False
        for key, item in value.items():
            if key == "$ref":
                continue
            if text_marks_secret(str(key)) or schema_marks_secret(item, spec, seen):
                return True
        return False
    if isinstance(value, list):
        return any(schema_marks_secret(item, spec, seen) for item in value)
    return text_marks_secret(value)


def parameter_marks_secret(param: dict[str, Any], spec: dict[str, Any] | None = None) -> bool:
    return schema_marks_secret(param, spec)


def is_secret_sensitive(op: dict[str, Any], spec: dict[str, Any] | None = None) -> bool:
    if text_marks_secret(op_blob(op)) or text_marks_secret(op.get("description")):
        return True
    return any(parameter_marks_secret(param, spec) for param in op.get("parameters") or [] if isinstance(param, dict))


def is_destructive(op: dict[str, Any], lane: str) -> bool:
    blob = op_blob(op)
    return lane == "reverse_etl_write" and (
        op["method"] == "DELETE"
        or any(term in blob for term in ("delete", "remove", "destroy", "purge", "revoke", "reset", "erase"))
    )


def risk_for(op: dict[str, Any], lane: str, spec: dict[str, Any] | None = None) -> str:
    if lane == "excluded_not_applicable":
        return "none"
    if is_destructive(op, lane):
        return "critical" if is_admin(op) else "high"
    if is_admin(op) or is_secret_sensitive(op, spec):
        return "high"
    if lane == "etl_read":
        return "low"
    if lane in {"direct_read_query_search", "binary_file", "cdc_changefeed"}:
        return "medium"
    return "medium"


def is_write_method(op: dict[str, Any]) -> bool:
    return op["method"] not in {"GET", "HEAD"}


def mutation_class_for(op: dict[str, Any], lane: str, spec: dict[str, Any] | None = None) -> str:
    if lane in {"excluded_not_applicable", "direct_read_query_search"} or not is_write_method(op):
        return "none"
    if op["method"] == "DELETE":
        return "delete"
    if is_admin(op):
        return "admin"
    if is_secret_sensitive(op, spec):
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


def cli_type_for_schema(schema: dict[str, Any]) -> str:
    if schema.get("enum"):
        return "enum"
    typ = schema.get("type", "string")
    if typ == "boolean":
        return "boolean"
    if typ == "integer":
        return "integer"
    if typ == "array":
        return "string_array"
    return "string"


def cli_flag_name(location: str, name: str) -> str:
    normalized = re.sub(r"([a-z0-9])([A-Z])", r"\1-\2", name).lower()
    normalized = normalized.replace("_", "-")
    normalized = re.sub(r"[^a-z0-9]+", "-", normalized).strip("-") or "value"
    return f"{location}-{normalized}"


def cli_flag_for_param(location: str, name: str, param: dict[str, Any], schema: dict[str, Any], required: bool) -> dict[str, Any]:
    flag: dict[str, Any] = {
        "name": cli_flag_name(location, name),
        "type": cli_type_for_schema(schema),
        "summary": clean_text(param.get("description"), f"{location} parameter {name}"),
        "maps_to": f"{location}.{name}",
    }
    if schema.get("enum"):
        flag["values"] = [str(value) for value in schema["enum"]]
    if schema.get("format") == "date-time":
        flag["format"] = "date-time"
    if required and flag["type"] == "string":
        flag["allow_empty"] = False
    return flag


def add_cli_flag(flags: list[dict[str, Any]], flag: dict[str, Any], seen_maps_to: set[str], seen_names: set[str]) -> None:
    maps_to = str(flag.get("maps_to"))
    if maps_to in seen_maps_to:
        return
    base_name = str(flag["name"])
    name = base_name
    i = 2
    while name in seen_names:
        name = f"{base_name}-{i}"
        i += 1
    flag["name"] = name
    seen_maps_to.add(maps_to)
    seen_names.add(name)
    flags.append(flag)


def parameter_contract(op: dict[str, Any], spec: dict[str, Any]) -> tuple[list[dict[str, Any]], dict[str, str], dict[str, Any] | None]:
    flags: list[dict[str, Any]] = []
    fixed_query: dict[str, str] = {}
    seen_maps_to: set[str] = set()
    seen_names: set[str] = set()
    body_schema: dict[str, Any] | None = None
    form_props: dict[str, Any] = {}
    form_required: list[str] = []
    path_names = path_param_names(op["path"])
    for param in op.get("parameters") or []:
        if not isinstance(param, dict):
            continue
        location = param.get("in")
        name = param.get("name")
        if location in {"path", "query"} and name:
            name = str(name)
            effective_location = "path" if name in path_names else str(location)
            schema = json_type_for_param(param)
            enum = schema.get("enum")
            if effective_location == "query" and isinstance(enum, list) and len(enum) == 1:
                fixed_query[name] = str(enum[0])
                continue
            required = effective_location == "path" or bool(param.get("required"))
            add_cli_flag(flags, cli_flag_for_param(effective_location, name, param, schema, required), seen_maps_to, seen_names)
        elif location == "body":
            body_schema = clean_schema(param.get("schema") or {"type": "object"}, spec)
        elif location == "formData" and name:
            schema = json_type_for_param(param)
            form_props[str(name)] = schema
            required = bool(param.get("required"))
            if required:
                form_required.append(str(name))
            add_cli_flag(flags, cli_flag_for_param("body", str(name), param, schema, required), seen_maps_to, seen_names)
    for name in sorted(path_names):
        if f"path.{name}" in seen_maps_to:
            continue
        add_cli_flag(
            flags,
            cli_flag_for_param("path", name, {"description": f"path parameter {name}"}, {"type": "string"}, True),
            seen_maps_to,
            seen_names,
        )
    if form_props:
        body_schema = {"type": "object", "properties": form_props}
        if form_required:
            body_schema["required"] = sorted(form_required)
    return flags, dict(sorted(fixed_query.items())), body_schema


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


def api_operation_notes(lane: str) -> str:
    if lane == "reverse_etl_write":
        return "DELETE/destructive/admin operations stay in scope; execution requires a named action with destructive confirmation plus plan -> preview -> explicit approval -> execute evidence."
    if lane == "binary_file":
        return "Binary/file transfer rows stay blocked until a bounded connector-local download command has fixtures and size-limit evidence."
    if lane == "cdc_changefeed":
        return "Changefeed/audit/webhook rows stay blocked until connector-local CDC or webhook fixtures and delivery semantics are verified."
    return "Non-mutating planned read/search/metadata rows stay blocked until connector-local command fixtures and bounded response evidence are verified."


def api_surface_row(op: dict[str, Any], lane: str, op_id: str, spec: dict[str, Any]) -> dict[str, Any]:
    key = (op["method"], op["path"])
    row: dict[str, Any] = {"method": op["method"], "path": relative_path(op["path"])}
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
    risk = risk_for(op, lane, spec)
    row["operation"] = {
        "model": operation_model_for(op, lane),
        "status": "blocked",
        "risk": "critical" if risk == "critical" else ("high" if risk == "high" else ("medium" if risk == "medium" else "low")),
        "blocked_by_default": True,
        "reason": f"planned GitLab {lane} operation; connector-local typed metadata exists as {op_id}, but no verified executable stream/action/command is claimed in this wave",
        "source_url": source_url,
        "notes": api_operation_notes(lane),
    }
    return row


def operation_row(op: dict[str, Any], lane: str, op_id: str, spec: dict[str, Any]) -> dict[str, Any]:
    summary = clean_text(op.get("summary"), op.get("operationId") or f"{op['method']} {op['path']}")
    risk = risk_for(op, lane, spec)
    destructive = is_destructive(op, lane)
    admin = is_admin(op)
    secret_sensitive = is_secret_sensitive(op, spec)
    if lane == "excluded_not_applicable":
        approval = "not_applicable"
    elif is_write_method(op):
        approval = "typed_confirmation + plan_preview_approval_execute" if destructive or admin else "plan_preview_approval_execute"
    else:
        approval = "none; planned bounded connector command/stream required before execution"
    output_policy = "binary_file_bounded" if lane == "binary_file" else "json_redacted"
    if lane == "excluded_not_applicable":
        kind = "composite"
    elif op["method"] == "HEAD":
        kind = "composite"
    elif lane == "binary_file" and op["method"] == "GET":
        kind = "binary_download"
    elif is_write_method(op) and not (lane == "direct_read_query_search" and op["method"] == "POST"):
        kind = "rest_write"
    else:
        kind = "rest_read"
    _flags, fixed_query, body_schema = parameter_contract(op, spec)
    rest: dict[str, Any] | None = None
    if kind in {"rest_read", "rest_write"}:
        rest = {
            "method": op["method"],
            "path": relative_path(op["path"]),
            "max_bytes": 10 * 1024 * 1024 if lane in {"direct_read_query_search", "cdc_changefeed"} else 1024 * 1024,
        }
        consumes = list(op.get("consumes") or [])
        if consumes:
            rest["content_type"] = consumes[0]
        elif body_schema is not None or (kind == "rest_read" and op["method"] == "POST"):
            rest["content_type"] = "application/json"
        if fixed_query:
            rest["query"] = fixed_query
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
        "mutation_class": mutation_class_for(op, lane, spec),
    }
    if kind == "composite":
        row["composite"] = {"steps": [f"{op['method']} {relative_path(op['path'])} planned metadata only"]}
    elif kind == "binary_download":
        row["binary"] = {
            "method": op["method"],
            "path": relative_path(op["path"]),
            "max_bytes": 16 * 1024 * 1024,
            "allow_overwrite": False,
            "extract_archives": False,
        }
    else:
        row["rest"] = rest
    if destructive:
        row["destructive"] = True
    if secret_sensitive:
        row["secret_sensitive"] = True
        row["sensitive_policy"] = {
            "input_mode": "env_or_file",
            "redact_fields": ["value", "token", "secret", "key", "password", "credential", "certificate", "private_key", "secure_file", "file"],
            "preflight": "secret-shaped inputs must be supplied outside prompt text and redacted from previews/errors",
            "transform": "none",
            "approval_mode": "typed_confirmation",
        }
    return row


def api_surface_supplemental_row(row: dict[str, Any]) -> dict[str, Any]:
    return {
        "method": row["method"],
        "path": row["path"],
        "covered_by": deepcopy(row["covered_by"]),
    }


def cli_command_for_stream(stream: str, method: str, path: str, source_url: str = OPENAPI_URL) -> dict[str, Any]:
    return {
        "path": f"{stream} list",
        "summary": f"List GitLab {stream} as ETL records.",
        "intent": "etl",
        "availability": "implemented",
        "stream": stream,
        "source_url": source_url,
        "api_surface": [{"method": method, "path": path}],
        "notes": "Implemented fixture-backed GitLab ETL stream; use --limit for bounded output and saved credentials/config for live reads.",
        "examples": [f"pm gitlab {stream} list --json --limit 25"],
    }


def cli_command_for_operation(op: dict[str, Any], lane: str, op_id: str, spec: dict[str, Any]) -> dict[str, Any]:
    method = op["method"].lower()
    slug = re.sub(r"[^a-z0-9]+", "-", relative_path(op["path"]).lower()).strip("-") or "root"
    operation_risk = risk_for(op, lane, spec)
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
        return cli_command_for_stream(stream, op["method"], relative_path(op["path"]))
    if lane != "excluded_not_applicable":
        cmd["api_surface"] = [{"method": op["method"], "path": relative_path(op["path"])}]
    flags, _fixed_query, _body_schema = parameter_contract(op, spec)
    if flags:
        cmd["flags"] = flags
    if intent == "etl" and availability == "planned" and operation_risk in {"high", "critical"}:
        cmd["output_policy"] = "json_redacted"
        cmd["risk"] = "Planned sensitive/admin GitLab read metadata; no live provider call is made by this metadata."
    if intent == "reverse_etl":
        cmd["risk"] = "Planned GitLab mutation; execution requires a named reverse-ETL action with typed schema, redaction, and safety evidence."
        cmd["approval"] = "Reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive/admin writes also require typed confirmation."
    if intent == "direct_read":
        if lane == "binary_file":
            cmd["output_policy"] = "binary_file_bounded"
            cmd["risk"] = "Planned bounded binary/file metadata; no live provider call is made by this metadata."
        else:
            cmd["output_policy"] = "json_redacted"
            cmd["risk"] = "Planned bounded read/query/metadata; no live provider call is made by this metadata."
    return cmd


def build_cli_surface(ops: list[dict[str, Any]], lanes: dict[tuple[str, str], str], ids: dict[tuple[str, str], str], spec: dict[str, Any]) -> dict[str, Any]:
    commands = [cli_command_for_operation(op, lanes[(op["method"], op["path"])], ids[(op["method"], op["path"])], spec) for op in ops]
    for row in SUPPLEMENTAL_STREAM_ROWS:
        commands.append(
            cli_command_for_stream(
                row["covered_by"]["stream"],
                row["method"],
                row["path"],
                str(row.get("source_url") or OPENAPI_URL),
            )
        )
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
        f"The complete official ledger has 1,146 operations: {counts['etl_read']} ETL/read, {counts['reverse_etl_write']} reverse-ETL write/mutation, {counts['direct_read_query_search']} direct/provider query/search/metadata, {counts['binary_file']} binary/file read/transfer, {counts['cdc_changefeed']} CDC/changefeed/audit/webhook, and {counts['excluded_not_applicable']} excluded/not-applicable callback endpoints.",
        "",
        "Only the four streams are executable in this wave. `api_surface.json`, `operations.json`, and `cli_surface.json` keep every other operation represented as typed planned/blocked metadata until a future connector-local stream/action/command adds fixtures and execution evidence.",
        "",
        "`api_surface.json` also includes one connector-local supplemental GET `/users` row for the fixture-backed users stream because that documented top-level endpoint is absent from the pinned OpenAPI source; official GET `/projects/{id}/users` remains planned metadata.",
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
        "- `users`: GET `/users`; records at the response root; sends `per_page=50`; optionally sends `created_after` from `start_date`. The pinned OpenAPI source does not enumerate this top-level users collection, so `api_surface.json` uses a connector-local supplemental coverage row and leaves GET `/projects/{id}/users` as planned metadata.",
        "- `issues`: GET `/issues`; records at the response root; sends `per_page=50`; optionally sends `updated_after` from `start_date`; derives `author_id` from `author.id`.",
        "",
        "Planned ETL/read rows in `operations.json` are metadata only. They are not advertised as executable streams until each has a schema, fixture replay, and conformance evidence.",
        "",
        "## Write actions & risks",
        "",
        "No `writes.json` actions are executable in this wave, and connector metadata keeps `capabilities.write=false` until named actions and fixtures are added.",
        "",
        f"The official ledger still includes all {counts['reverse_etl_write']} mutation operations, including DELETE, destructive, admin, token/key/variable, hook, runner, package delete, and other high-risk operations. These are not blanket-excluded as unsafe. They are represented as planned/blocked typed metadata with risk, source URL, bounded request schemas where available, and approval notes.",
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
        "- Direct/provider search/query/metadata rows depend on shared foundation #2985 before execution can be claimed.",
        "- Binary/file transfer rows depend on shared foundation #2987 before bounded download/upload execution can be claimed.",
        "- CDC/changefeed/audit/webhook rows depend on shared foundations #2986 and #2988 before CDC/changefeed claims can be made.",
        "- Destructive/admin write rows depend on per-action schemas, redaction, fixtures, and typed confirmation evidence before execution can be claimed.",
        "- The current top-level `/users` stream is fixture-backed legacy behavior; the pinned OpenAPI source omits that exact row, so api surface coverage uses a connector-local supplemental row and does not mark project-scoped users as implemented.",
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
        api_rows.append(api_surface_row(op, lane, op_id, spec))
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
    api_rows_with_supplemental = api_rows + [api_surface_supplemental_row(row) for row in SUPPLEMENTAL_STREAM_ROWS]
    api_surface = {
        "api": "GitLab REST API v4 (pinned OpenAPI v2 source)",
        "docs": "https://docs.gitlab.com/ee/api/rest/",
        "reviewed_at": "2026-07-31",
        "operation_ledger_version": 1,
        "scope": "Complete official GitLab OpenAPI v2 ledger at gitlab-org/gitlab commit 9cd04099eb59d87335798e4f57a2bc5a2622e4cc plus one connector-local supplemental GET /users coverage row for the existing users stream. Four existing streams are executable; all official operations not covered by streams are planned/blocked typed metadata until connector-local schemas, fixtures, commands, and safety evidence make them executable. DELETE/destructive/admin operations are included and require typed destructive confirmation plus plan -> preview -> explicit approval -> execute before execution.",
        "endpoints": api_rows_with_supplemental,
    }
    operations = {"operations": operation_rows}
    cli_surface = build_cli_surface(ops, lanes, ids, spec)
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
    supplemental_streams = sorted(row["covered_by"]["stream"] for row in SUPPLEMENTAL_STREAM_ROWS)
    summary = {
        "official_operations": len(ops),
        "api_surface_endpoints": len(api_rows_with_supplemental),
        "lane_counts": dict(counts),
        "covered_streams": sorted(set(STREAM_COVERAGE.values()) | set(supplemental_streams)),
        "official_covered_streams": sorted(STREAM_COVERAGE.values()),
        "supplemental_stream_endpoints": deepcopy(SUPPLEMENTAL_STREAM_ROWS),
        "blocked_or_planned": len(ops) - len(STREAM_COVERAGE) - counts["excluded_not_applicable"],
        "excluded_not_applicable": counts["excluded_not_applicable"],
        "source": OPENAPI_URL,
    }
    dump_json(ROOT / ".planning/phases/issue-78-gitlab-parity-wave02-r1/traces/gitlab_generation_summary.json", summary)
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
