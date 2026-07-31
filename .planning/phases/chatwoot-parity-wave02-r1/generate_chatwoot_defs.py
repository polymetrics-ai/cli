#!/usr/bin/env python3
"""Generate Chatwoot connector-local parity metadata from official OpenAPI sources.

This script is a planning artifact for issue #148 wave02-r1. It fetches public
Chatwoot OpenAPI JSON without credentials and rewrites only Chatwoot-owned
connector definition/docs files.
"""
from __future__ import annotations

import copy
import json
import re
import urllib.request
from collections import Counter, defaultdict
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[3]
DEFS = ROOT / "internal/connectors/defs/chatwoot"
DOCS = ROOT / "docs/connectors/chatwoot"
PHASE = Path(__file__).resolve().parent

SOURCES = {
    "chatwoot_application_openapi": "https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/tag_groups/application_swagger.json",
    "chatwoot_platform_openapi": "https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/tag_groups/platform_swagger.json",
    "chatwoot_client_openapi": "https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/tag_groups/client_swagger.json",
    "chatwoot_other_openapi": "https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/tag_groups/other_swagger.json",
}
SOURCE_DOC = {
    "chatwoot_application_openapi": SOURCES["chatwoot_application_openapi"],
    "chatwoot_platform_openapi": SOURCES["chatwoot_platform_openapi"],
    "chatwoot_client_openapi": SOURCES["chatwoot_client_openapi"],
    "chatwoot_other_openapi": SOURCES["chatwoot_other_openapi"],
}
REVIEWED_AT = "2026-07-31"

CURRENT_WRITE_NAMES = {
    ("POST", "/api/v1/accounts/{account_id}/contacts"): "create_contact",
    ("PUT", "/api/v1/accounts/{account_id}/contacts/{id}"): "update_contact",
    ("POST", "/api/v1/accounts/{account_id}/conversations"): "create_conversation",
    ("POST", "/api/v1/accounts/{account_id}/conversations/{conversation_id}/messages"): "send_message",
    ("POST", "/api/v1/accounts/{account_id}/conversations/{conversation_id}/toggle_status"): "toggle_conversation_status",
    ("POST", "/api/v1/accounts/{account_id}/labels"): "create_label",
}
STREAM_COVERAGE = {
    ("GET", "/api/v1/accounts/{account_id}/conversations"): "conversations",
    ("GET", "/api/v1/accounts/{account_id}/contacts"): "contacts",
    ("GET", "/api/v1/accounts/{account_id}/inboxes"): "inboxes",
    ("GET", "/api/v1/accounts/{account_id}/agents"): "agents",
    ("GET", "/api/v1/accounts/{account_id}/teams"): "teams",
    ("GET", "/api/v1/accounts/{account_id}/labels"): "labels",
    ("GET", "/api/v1/accounts/{account_id}/conversations/{conversation_id}/messages"): "messages",
}
DIRECT_READ_PATH_HINTS = (
    "/search",
    "/filter",
    "/reports",
    "/summary_reports",
    "/statistics",
    "/reporting_events",
    "/custom_filters",
    "/profile",
    "/survey/",
)
CDC_HINTS = ("audit_logs", "reporting_events", "webhook")
MUTATION_METHODS = {"POST", "PUT", "PATCH", "DELETE"}
WRITE_KIND_BY_METHOD = {"POST": "create", "PUT": "update", "PATCH": "update", "DELETE": "delete"}


def load_json(path: Path) -> Any:
    return json.loads(path.read_text())


def write_json(path: Path, value: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, indent=2, sort_keys=False) + "\n")


def fetch_openapi() -> dict[str, Any]:
    out: dict[str, Any] = {}
    for source_id, url in SOURCES.items():
        req = urllib.request.Request(url, headers={"User-Agent": "polymetrics-chatwoot-parity-worker"})
        with urllib.request.urlopen(req, timeout=30) as resp:
            out[source_id] = json.loads(resp.read())
    return out


def iter_operations(specs: dict[str, Any]) -> list[dict[str, Any]]:
    ops: list[dict[str, Any]] = []
    for source_id, spec in specs.items():
        group = source_id.removeprefix("chatwoot_").removesuffix("_openapi")
        for path, item in spec.get("paths", {}).items():
            common_params = item.get("parameters", [])
            for method_l, op in item.items():
                method = method_l.upper()
                if method not in {"GET", "POST", "PUT", "PATCH", "DELETE"}:
                    continue
                ops.append(
                    {
                        "source_id": source_id,
                        "group": group,
                        "method": method,
                        "path": path,
                        "operation_id": op.get("operationId") or "",
                        "summary": op.get("summary") or op.get("description") or "Chatwoot operation",
                        "description": op.get("description") or "",
                        "tags": op.get("tags") or [],
                        "operation": op,
                        "path_item_parameters": common_params,
                    }
                )
    ops.sort(key=lambda o: (o["source_id"], o["path"], o["method"], o["operation_id"]))
    return ops


def ref_name(ref: str) -> str:
    return ref.rsplit("/", 1)[-1]


def resolve_ref(spec: dict[str, Any], obj: Any) -> Any:
    if isinstance(obj, dict) and "$ref" in obj:
        name = ref_name(obj["$ref"])
        for section in ("schemas", "parameters", "requestBodies"):
            if name in spec.get("components", {}).get(section, {}):
                return resolve_ref(spec, spec["components"][section][name])
    return obj


def merge_all_of(spec: dict[str, Any], schema: dict[str, Any], seen: set[str] | None = None) -> dict[str, Any]:
    schema = resolve_ref(spec, schema)
    if not isinstance(schema, dict):
        return {"type": "object", "additionalProperties": True}
    if "allOf" not in schema:
        return schema
    merged: dict[str, Any] = {"type": "object", "properties": {}, "required": []}
    for part in schema.get("allOf", []):
        part = merge_all_of(spec, part, seen)
        if part.get("properties"):
            merged.setdefault("properties", {}).update(part["properties"])
        if part.get("required"):
            merged.setdefault("required", []).extend(part["required"])
    if schema.get("properties"):
        merged.setdefault("properties", {}).update(schema["properties"])
    if schema.get("required"):
        merged.setdefault("required", []).extend(schema["required"])
    if merged.get("required"):
        merged["required"] = sorted(set(merged["required"]))
    return merged


def schema_type(schema: dict[str, Any]) -> Any:
    typ = schema.get("type")
    if isinstance(typ, list):
        non_null = [t for t in typ if t != "null"]
        return non_null[0] if non_null else "string"
    if typ:
        return typ
    if "properties" in schema:
        return "object"
    if "items" in schema:
        return "array"
    if "enum" in schema:
        value = schema["enum"][0] if schema["enum"] else "fixture"
        if isinstance(value, bool):
            return "boolean"
        if isinstance(value, int):
            return "integer"
        if isinstance(value, float):
            return "number"
    return "string"


def simplify_schema(spec: dict[str, Any], schema: Any, depth: int = 0) -> dict[str, Any]:
    schema = merge_all_of(spec, resolve_ref(spec, schema if isinstance(schema, dict) else {}))
    if not isinstance(schema, dict):
        return {"type": "string"}
    typ = schema_type(schema)
    out: dict[str, Any] = {}
    if isinstance(typ, list):
        out["type"] = typ
    else:
        out["type"] = typ
    if schema.get("description"):
        out["description"] = str(schema["description"]).strip()[:240]
    if schema.get("enum") and all(isinstance(x, (str, int, float, bool)) or x is None for x in schema["enum"]):
        out["enum"] = schema["enum"]
    if schema.get("format") in {"date-time", "date", "email", "uri"}:
        out["format"] = schema["format"]
    if typ == "object":
        props = schema.get("properties") or {}
        if props and depth < 2:
            out["properties"] = {snake(k): simplify_schema(spec, v, depth + 1) for k, v in props.items()}
            req = [snake(r) for r in schema.get("required", []) if r in props]
            if req:
                out["required"] = req
            out["additionalProperties"] = bool(schema.get("additionalProperties", False))
        else:
            out["additionalProperties"] = True
    elif typ == "array":
        out["items"] = simplify_schema(spec, schema.get("items", {"type": "string"}), depth + 1)
    return out


def request_body_schema(spec: dict[str, Any], op: dict[str, Any]) -> dict[str, Any] | None:
    rb = op.get("requestBody")
    if not rb:
        return None
    rb = resolve_ref(spec, rb)
    content = rb.get("content", {}) if isinstance(rb, dict) else {}
    media = content.get("application/json") or content.get("application/json; charset=utf-8")
    if not media and content:
        # Chatwoot currently uses JSON for these APIs; fall back only to keep the ledger typed.
        media = next(iter(content.values()))
    if not media:
        return None
    schema = media.get("schema") or {}
    return simplify_schema(spec, schema)


def parameters_for(spec: dict[str, Any], op_item: dict[str, Any], op: dict[str, Any]) -> list[dict[str, Any]]:
    params = []
    for raw in list(op_item.get("path_item_parameters", [])) + list(op.get("parameters", [])):
        p = resolve_ref(spec, raw)
        if isinstance(p, dict):
            params.append(p)
    return params


def path_params_for(spec: dict[str, Any], op_item: dict[str, Any]) -> list[dict[str, Any]]:
    return [p for p in parameters_for(spec, op_item, op_item["operation"]) if p.get("in") == "path"]


def query_params_for(spec: dict[str, Any], op_item: dict[str, Any]) -> list[dict[str, Any]]:
    return [p for p in parameters_for(spec, op_item, op_item["operation"]) if p.get("in") == "query"]


def snake(value: str) -> str:
    value = re.sub(r"([a-z0-9])([A-Z])", r"\1_\2", value)
    value = re.sub(r"[^A-Za-z0-9]+", "_", value).strip("_").lower()
    return value or "value"


def action_name_for(op_item: dict[str, Any], used: set[str]) -> str:
    key = (op_item["method"], op_item["path"])
    if key in CURRENT_WRITE_NAMES:
        name = CURRENT_WRITE_NAMES[key]
    else:
        base = snake(op_item["operation_id"] or f"{op_item['method']}_{op_item['path']}")
        # Drop noisy articles while preserving verb-first names.
        base = re.sub(r"\b(a|an|the)_", "", base)
        name = base
    if name in used:
        name = f"{op_item['group']}_{name}"
    suffix = 2
    original = name
    while name in used:
        name = f"{original}_{suffix}"
        suffix += 1
    used.add(name)
    return name


def operation_id_for(op_item: dict[str, Any], used: set[str]) -> str:
    fallback = f"{op_item['method']}_{op_item['path']}"
    base = f"chatwoot.{snake(op_item['operation_id'] or fallback)}"
    if base in used:
        base = f"chatwoot.{op_item['group']}.{base.removeprefix('chatwoot.')}"
    original = base
    i = 2
    while base in used:
        base = f"{original}.{i}"
        i += 1
    used.add(base)
    return base


def templated_path(path: str) -> tuple[str, list[str]]:
    path_fields: list[str] = []
    def repl(match: re.Match[str]) -> str:
        name = match.group(1)
        if name == "account_id":
            return "{{ config.account_id }}"
        field = snake(name)
        if field not in path_fields:
            path_fields.append(field)
        return "{{ record." + field + " }}"
    return re.sub(r"\{([^}]+)\}", repl, path), path_fields


def path_schema_for(param: dict[str, Any]) -> dict[str, Any]:
    sch = param.get("schema") or {}
    typ = schema_type(sch if isinstance(sch, dict) else {})
    out: dict[str, Any] = {"type": "integer" if typ == "integer" else "string"}
    desc = param.get("description") or param.get("name")
    if desc:
        out["description"] = str(desc).strip()[:200]
    if typ != "integer":
        out["pattern"] = "^[\\s\\S]+$"
    return out


def merge_record_schema(path_params: list[dict[str, Any]], path_fields: list[str], body_schema: dict[str, Any] | None, require_body: bool) -> dict[str, Any]:
    props: dict[str, Any] = {}
    required: list[str] = []
    for p in path_params:
        raw_name = p.get("name")
        if raw_name == "account_id":
            continue
        name = snake(raw_name or "id")
        props[name] = path_schema_for(p)
        if name in path_fields and name not in required:
            required.append(name)
    if body_schema and body_schema.get("type") == "object" and body_schema.get("properties"):
        for k, v in body_schema.get("properties", {}).items():
            if k not in props:
                props[k] = v
        body_required = body_schema.get("required") or []
        # If the provider marks requestBody required but omits field-level required,
        # allow a sparse update body for PATCH/PUT and require one property for create.
        for r in body_required:
            if r in props and r not in required:
                required.append(r)
    elif body_schema and body_schema.get("type") not in (None, "object"):
        props["payload"] = body_schema
        if require_body:
            required.append("payload")
    schema: dict[str, Any] = {
        "$schema": "http://json-schema.org/draft-07/schema#",
        "type": "object",
        "properties": props,
        "additionalProperties": False,
    }
    if required:
        schema["required"] = required
    if body_schema and body_schema.get("type") == "object" and not body_schema.get("required") and require_body:
        non_path = [k for k in props if k not in path_fields]
        if non_path:
            schema["minProperties"] = len(path_fields) + 1
    return schema


def sample_for_schema(schema: dict[str, Any], name: str = "value") -> Any:
    if "enum" in schema and schema["enum"]:
        for val in schema["enum"]:
            if val is not None:
                return val
    typ = schema_type(schema)
    if typ == "integer":
        return 1
    if typ == "number":
        return 1.5
    if typ == "boolean":
        return True
    if typ == "array":
        return [sample_for_schema(schema.get("items", {"type": "string"}), name)]
    if typ == "object":
        props = schema.get("properties") or {}
        req = schema.get("required") or list(props)[:1]
        out = {}
        for k in req:
            out[k] = sample_for_schema(props.get(k, {"type": "string"}), k)
        return out
    if "email" in name:
        return "fixture@example.invalid"
    if "url" in name:
        return "https://fixture.example.invalid/resource"
    if "date" in name or name.endswith("_at"):
        return "2026-01-01T00:00:00Z"
    if name.endswith("id") or name == "id":
        return "fixture-id"
    return "fixture-value"


def sample_record(schema: dict[str, Any], path_fields: list[str]) -> dict[str, Any]:
    props = schema.get("properties", {})
    required = schema.get("required") or list(props)
    record = {k: sample_for_schema(props.get(k, {"type": "string"}), k) for k in required}
    # Ensure path fields are always present and scalar.
    for f in path_fields:
        prop = props.get(f, {"type": "string"})
        record[f] = 1 if schema_type(prop) == "integer" else f"fixture-{f}"
    # Include one optional body property for update actions so dry-run bodies are non-empty.
    for k, prop in props.items():
        if k not in path_fields and k not in record:
            record[k] = sample_for_schema(prop, k)
            break
    return record


def write_fixture_for(action: dict[str, Any], record: dict[str, Any]) -> dict[str, Any]:
    body_type = action.get("body_type") or "json"
    path = action["path"]
    for f in action.get("path_fields", []):
        path = path.replace("{{ record." + f + " }}", str(record[f]))
    # Conformance materializes config.account_id with this synthetic value.
    path = path.replace("{{ config.account_id }}", "synthetic-conformance-value")
    expect: dict[str, Any] = {"method": action["method"], "path": path}
    if body_type == "none":
        body_fields = action.get("body_fields", [])
        if body_fields:
            expect["body"] = {k: record[k] for k in body_fields if k in record}
    else:
        body = {k: v for k, v in record.items() if k not in set(action.get("path_fields", []))}
        if body:
            expect["body"] = body
    response_status = 200 if action["method"] != "POST" else 201
    return {"record": record, "expect": expect, "response": {"status": response_status, "body": {"id": 1, "status": "ok"}}}


def has_query_write_gap(specs: dict[str, Any], op_item: dict[str, Any]) -> bool:
    spec = specs[op_item["source_id"]]
    return bool(query_params_for(spec, op_item))


def is_direct_post(op_item: dict[str, Any]) -> bool:
    return op_item["method"] == "POST" and any(x in op_item["path"] for x in ("/filter", "/search"))


def should_generate_write(specs: dict[str, Any], op_item: dict[str, Any]) -> bool:
    if op_item["method"] not in MUTATION_METHODS:
        return False
    if op_item["group"] == "client":
        return False
    if is_direct_post(op_item):
        return False
    if has_query_write_gap(specs, op_item):
        return False
    return True


def risk_for(op_item: dict[str, Any]) -> str:
    path = op_item["path"].lower()
    method = op_item["method"]
    text = f"{op_item['operation_id']} {op_item['summary']} {' '.join(op_item['tags'])}".lower()
    if method == "DELETE" or "delete" in text or "remove" in text:
        return "critical" if "account" in path or op_item["group"] == "platform" else "high"
    if op_item["group"] == "platform" or any(x in text for x in ("agent", "user", "role", "account")):
        return "high"
    if any(x in text for x in ("webhook", "assignment", "sla", "portal", "article")):
        return "medium"
    return "medium" if method in MUTATION_METHODS else "low"


def auth_scopes_for(op_item: dict[str, Any]) -> list[str]:
    if op_item["group"] == "client":
        return ["public_client"]
    if op_item["group"] == "platform":
        return ["platform_app_api_access_token"]
    return ["user_api_access_token"]


def operation_model_for(op_item: dict[str, Any]) -> str:
    if op_item["method"] in MUTATION_METHODS:
        if op_item["method"] == "DELETE" or "delete" in op_item["operation_id"].lower() or "remove" in op_item["summary"].lower():
            return "destructive_action"
        if op_item["group"] == "platform":
            return "admin_reverse_etl"
        return "sensitive_reverse_etl"
    if op_item["method"] == "GET" or is_direct_post(op_item):
        return "direct_read"
    return "disallowed"


def operation_reason(op_item: dict[str, Any], write_generated: bool) -> str:
    if write_generated:
        return "covered by a named Chatwoot reverse ETL write action; execution still requires plan, preview, approval token, and destructive confirmation when marked"
    if op_item["group"] == "client":
        return "planned/blocked: Chatwoot public client API uses inbox/contact identifiers and no userApiKey security scheme; connector needs a separate no-auth/client-safe contract before execution"
    if is_direct_post(op_item):
        return "planned/blocked: provider search/filter POST requires bounded direct-read command contract with typed query/body flags and redacted output before execution"
    if has_query_write_gap({op_item["source_id"]: {}}, op_item):
        return "planned/blocked: endpoint requires query parameters on a write path; current connector-local write action schema has no query-param execution contract"
    if op_item["method"] == "GET":
        return "planned/blocked: documented read operation is represented but not executable until a typed stream or direct-read command, fixture, and schema are authored"
    return "planned/blocked: documented operation lacks a safe connector-local execution contract in this slice"


def generate_streams() -> None:
    streams = load_json(DEFS / "streams.json")
    streams["base"]["url"] = "{{ config.base_url }}"
    streams["base"]["check"] = {"method": "GET", "path": "/api/v1/accounts/{{ config.account_id }}/agents"}
    for stream in streams["streams"]:
        if stream["path"].startswith("/api/"):
            continue
        if stream["name"] == "messages":
            stream["path"] = "/api/v1/accounts/{{ config.account_id }}/conversations/{{ fanout.id }}/messages"
            stream["fan_out"]["ids_from"]["request"]["path"] = "/api/v1/accounts/{{ config.account_id }}/conversations"
        else:
            stream["path"] = "/api/v1/accounts/{{ config.account_id }}" + stream["path"]
    write_json(DEFS / "streams.json", streams)


def update_fixture_paths() -> None:
    for path in (DEFS / "fixtures/streams").glob("*/*.json"):
        data = load_json(path)
        req = data.get("request")
        if req and isinstance(req.get("path"), str) and not req["path"].startswith("/api/"):
            req["path"] = "/api/v1/accounts/synthetic-conformance-value" + req["path"]
        write_json(path, data)
    check = load_json(DEFS / "fixtures/check.json")
    check.setdefault("request", {"method": "GET", "path": "/api/v1/accounts/synthetic-conformance-value/agents"})
    check["request"] = {"method": "GET", "path": "/api/v1/accounts/synthetic-conformance-value/agents"}
    write_json(DEFS / "fixtures/check.json", check)


def generate_writes(specs: dict[str, Any], ops: list[dict[str, Any]]) -> tuple[list[dict[str, Any]], dict[tuple[str, str], str]]:
    used: set[str] = set()
    actions: list[dict[str, Any]] = []
    coverage: dict[tuple[str, str], str] = {}
    fixtures_dir = DEFS / "fixtures/writes"
    fixtures_dir.mkdir(parents=True, exist_ok=True)
    for op_item in ops:
        if not should_generate_write(specs, op_item):
            continue
        spec = specs[op_item["source_id"]]
        method = op_item["method"]
        action_name = action_name_for(op_item, used)
        path, path_fields = templated_path(op_item["path"])
        body_schema = request_body_schema(spec, op_item["operation"])
        path_params = path_params_for(spec, op_item)
        require_body = bool(op_item["operation"].get("requestBody", {}).get("required")) if isinstance(op_item["operation"].get("requestBody"), dict) else bool(op_item["operation"].get("requestBody"))
        record_schema = merge_record_schema(path_params, path_fields, body_schema, require_body)
        action: dict[str, Any] = {
            "name": action_name,
            "kind": WRITE_KIND_BY_METHOD.get(method, "custom"),
            "method": method,
            "path": path,
            "record_schema": record_schema,
            "risk": f"Chatwoot {op_item['group']} API mutation ({method} {op_item['path']}); reverse ETL plan, preview, explicit approval, and execute are required",
        }
        if path_fields:
            action["path_fields"] = path_fields
        if method == "DELETE":
            action["body_type"] = "none"
            action["delete"] = {"idempotent": True, "missing_ok_status": [404]}
            action["redact_fields"] = path_fields[:]
            action["risk"] = f"destructive Chatwoot mutation ({method} {op_item['path']}); requires reverse ETL plan, preview, explicit approval, and typed --confirm destructive before execute"
            action["confirm"] = "destructive"
        elif risk_for(op_item) in {"high", "critical"} or op_item["group"] == "platform":
            action["confirm"] = "destructive"
            action["risk"] += "; admin/elevated mutation is typed-confirmation gated"
        actions.append(action)
        coverage[(method, op_item["path"])] = action_name
        write_json(fixtures_dir / f"{action_name}.json", write_fixture_for(action, sample_record(record_schema, path_fields)))
    # Remove stale old fixture names only if their action no longer exists.
    valid = {f"{a['name']}.json" for a in actions}
    for old in fixtures_dir.glob("*.json"):
        if old.name not in valid:
            old.unlink()
    write_json(DEFS / "writes.json", {"actions": actions})
    return actions, coverage


def generate_operations(specs: dict[str, Any], ops: list[dict[str, Any]], write_coverage: dict[tuple[str, str], str]) -> tuple[list[dict[str, Any]], dict[tuple[str, str], str]]:
    used: set[str] = set()
    operation_ids: dict[tuple[str, str], str] = {}
    operations: list[dict[str, Any]] = []
    for op_item in ops:
        spec = specs[op_item["source_id"]]
        op_id = operation_id_for(op_item, used)
        operation_ids[(op_item["method"], op_item["path"])] = op_id
        write_generated = (op_item["method"], op_item["path"]) in write_coverage
        kind = "rest_write" if op_item["method"] in MUTATION_METHODS else "rest_read"
        risk = risk_for(op_item)
        approval = "none: read-only operation metadata; execution requires a declared stream or bounded direct-read command"
        if op_item["method"] in MUTATION_METHODS:
            approval = "reverse ETL plan -> preview -> explicit approval token -> execute"
            if risk in {"high", "critical"} or op_item["method"] == "DELETE" or op_item["group"] == "platform":
                approval += "; typed --confirm destructive required"
        rest: dict[str, Any] = {"method": op_item["method"], "path": op_item["path"], "max_bytes": 1048576}
        if op_item["method"] in {"POST", "PUT", "PATCH"}:
            body_schema = request_body_schema(spec, op_item["operation"])
            if body_schema:
                rest["content_type"] = "application/json"
                rest["body_schema"] = body_schema
        qps = query_params_for(spec, op_item)
        if qps:
            rest["query"] = {snake(p.get("name", "query")): "{{ query." + snake(p.get("name", "query")) + " }}" for p in qps}
        op_spec: dict[str, Any] = {
            "id": op_id,
            "kind": kind,
            "summary": op_item["summary"],
            "description": op_item["description"][:500] if op_item["description"] else op_item["summary"],
            "source_url": SOURCE_DOC[op_item["source_id"]],
            "risk": "critical" if risk == "critical" else risk,
            "approval": approval,
            "output_policy": "json_redacted",
            "auth_scopes": auth_scopes_for(op_item),
            "mutation_class": "none",
            "rest": rest,
        }
        if op_item["method"] in MUTATION_METHODS:
            if op_item["method"] == "DELETE" or "delete" in op_item["operation_id"].lower() or "remove" in op_item["summary"].lower():
                op_spec["mutation_class"] = "destructive"
                op_spec["destructive"] = True
            elif op_item["group"] == "platform":
                op_spec["mutation_class"] = "admin"
            elif op_item["method"] == "POST":
                op_spec["mutation_class"] = "create"
            else:
                op_spec["mutation_class"] = "update"
        operations.append(op_spec)
    write_json(DEFS / "operations.json", {"operations": operations})
    return operations, operation_ids


def generate_api_surface(ops: list[dict[str, Any]], write_coverage: dict[tuple[str, str], str], operation_ids: dict[tuple[str, str], str]) -> None:
    endpoints: list[dict[str, Any]] = []
    for op_item in ops:
        key = (op_item["method"], op_item["path"])
        ep: dict[str, Any] = {"method": op_item["method"], "path": op_item["path"]}
        if key in STREAM_COVERAGE:
            ep["covered_by"] = {"stream": STREAM_COVERAGE[key]}
        elif key in write_coverage:
            ep["covered_by"] = {"write": write_coverage[key]}
        else:
            model = operation_model_for(op_item)
            risk = risk_for(op_item)
            ep["operation"] = {
                "model": model,
                "status": "blocked",
                "risk": "critical" if risk == "critical" else risk,
                "blocked_by_default": True,
                "reason": blocked_reason(op_item),
                "source_url": SOURCE_DOC[op_item["source_id"]],
                "notes": f"typed operation metadata id: {operation_ids.get(key, '')}",
            }
        endpoints.append(ep)
    surface = {
        "api": "Chatwoot official OpenAPI tag groups (application, platform, client, other)",
        "docs": ", ".join(SOURCES.values()),
        "reviewed_at": REVIEWED_AT,
        "operation_ledger_version": 1,
        "scope": "Every operation from the four official Chatwoot OpenAPI tag-group sources is represented exactly once. Application/platform mutations executable in this slice use named reverse ETL write actions; public client/no-auth and query-param write gaps remain blocked until a separate safe contract exists.",
        "endpoints": endpoints,
    }
    write_json(DEFS / "api_surface.json", surface)


def blocked_reason(op_item: dict[str, Any]) -> str:
    if op_item["group"] == "client":
        return "planned/blocked: public client API requires inbox/contact identifiers and no userApiKey auth; a separate client-safe no-auth connector contract is required before execution"
    if is_direct_post(op_item):
        return "planned/blocked: provider search/filter POST needs bounded direct-read command metadata, typed body/query flags, and redacted output before execution"
    if op_item["method"] in MUTATION_METHODS and op_item["path"].endswith("/custom_filters"):
        return "planned/blocked: custom filter creation requires a query parameter (`filter_type`) on a write; shared write query-param support is not available in this connector-local slice"
    if op_item["method"] == "GET":
        if any(h in op_item["path"] for h in DIRECT_READ_PATH_HINTS):
            return "planned/blocked: bounded direct/provider query read is documented but not executable until connector-local command flags and redaction evidence are completed"
        return "planned/blocked: ETL stream candidate requires typed schema and fixture evidence before execution can be claimed"
    return "planned/blocked: safe execution contract is absent in this connector-local slice"


def command_group(path: str) -> str:
    parts = [p for p in path.split("/") if p and not p.startswith("{")]
    for p in parts:
        if p in {"api", "v1", "v2", "accounts", "platform", "public", "inboxes"}:
            continue
        return snake(p)
    return "operations"


def command_path_for(op_item: dict[str, Any], action_name: str | None = None) -> str:
    group = command_group(op_item["path"])
    verb = "delete" if op_item["method"] == "DELETE" else ("create" if op_item["method"] == "POST" else "update" if op_item["method"] in {"PATCH", "PUT"} else "get")
    if action_name and action_name.startswith("toggle_"):
        verb = snake(op_item["operation_id"]).replace("toggle_status_of_", "toggle_")
    return f"{group} {verb}"


def flag_type_for_schema(schema: dict[str, Any]) -> str:
    typ = schema_type(schema)
    if typ == "integer" or typ == "number":
        return "integer"
    if typ == "boolean":
        return "boolean"
    if typ == "array":
        return "string_array"
    if schema.get("enum"):
        return "enum"
    return "string"


def flags_for_action(action: dict[str, Any]) -> list[dict[str, Any]]:
    schema = action["record_schema"]
    props = schema.get("properties", {})
    required = set(schema.get("required", [])) | set(action.get("path_fields", []))
    flags = []
    for name in sorted(required):
        prop = props.get(name, {"type": "string"})
        flag: dict[str, Any] = {"name": name.replace("_", "-"), "type": flag_type_for_schema(prop), "summary": prop.get("description") or f"Chatwoot field {name}.", "maps_to": f"record.{name}"}
        if flag["type"] == "enum" and prop.get("enum"):
            flag["values"] = [str(v) for v in prop["enum"] if v is not None]
        if flag["type"] == "string":
            flag["allow_empty"] = False
        flags.append(flag)
    return flags[:12]


def generate_cli_surface(ops: list[dict[str, Any]], actions: list[dict[str, Any]], write_coverage: dict[tuple[str, str], str], operation_ids: dict[tuple[str, str], str]) -> None:
    action_by_name = {a["name"]: a for a in actions}
    groups: dict[str, list[str]] = defaultdict(list)
    commands: list[dict[str, Any]] = []
    stream_labels = {
        "conversations": ("conversations list", "Read Chatwoot conversations through the declared ETL stream.", "/api/v1/accounts/{account_id}/conversations"),
        "contacts": ("contacts list", "Read Chatwoot contacts through the declared ETL stream.", "/api/v1/accounts/{account_id}/contacts"),
        "inboxes": ("inboxes list", "Read Chatwoot inboxes through the declared ETL stream.", "/api/v1/accounts/{account_id}/inboxes"),
        "agents": ("agents list", "Read Chatwoot account agents through the declared ETL stream.", "/api/v1/accounts/{account_id}/agents"),
        "teams": ("teams list", "Read Chatwoot teams through the declared ETL stream.", "/api/v1/accounts/{account_id}/teams"),
        "labels": ("labels list", "Read Chatwoot account labels through the declared ETL stream.", "/api/v1/accounts/{account_id}/labels"),
        "messages": ("messages list", "Read Chatwoot conversation messages through the declared fan-out ETL stream.", "/api/v1/accounts/{account_id}/conversations/{conversation_id}/messages"),
    }
    for stream, (path, summary, api_path) in stream_labels.items():
        groups[path.split()[0]].append(path)
        commands.append({
            "path": path,
            "summary": summary,
            "intent": "etl",
            "availability": "implemented",
            "stream": stream,
            "source_cli_path": f"GET {api_path}",
            "source_url": SOURCES["chatwoot_application_openapi"],
            "api_surface": [{"method": "GET", "path": api_path}],
            "examples": [f"pm chatwoot {path} --credential chatwoot-prod --limit 25 --json"],
        })
    for op_item in ops:
        key = (op_item["method"], op_item["path"])
        action_name = write_coverage.get(key)
        if not action_name:
            # Add docs-only command rows for blocked direct/client/admin gaps without executable targets.
            if op_item["method"] == "GET" and any(h in op_item["path"] for h in DIRECT_READ_PATH_HINTS):
                cp = f"{command_group(op_item['path'])} planned"
                groups[cp.split()[0]].append(cp)
                commands.append({
                    "path": cp,
                    "summary": f"Planned bounded direct read for {op_item['summary']}.",
                    "intent": "direct_read",
                    "availability": "planned",
                    "operation": operation_ids.get((op_item["method"], op_item["path"]), ""),
                    "source_cli_path": f"{op_item['method']} {op_item['path']}",
                    "source_url": SOURCE_DOC[op_item["source_id"]],
                    "api_surface": [{"method": op_item["method"], "path": op_item["path"]}],
                    "output_policy": "json_redacted",
                    "notes": blocked_reason(op_item),
                })
            continue
        action = action_by_name[action_name]
        cp = command_path_for(op_item, action_name)
        # Avoid duplicate command paths by appending operation id tail.
        seen_paths = {c["path"] for c in commands}
        if cp in seen_paths:
            cp = f"{cp} {snake(op_item['operation_id']).split('_')[-1]}"
        groups[cp.split()[0]].append(cp)
        implemented = action_name in set(CURRENT_WRITE_NAMES.values()) or action.get("method") == "DELETE"
        cmd = {
            "path": cp,
            "summary": f"Plan {op_item['summary']} through a named Chatwoot reverse ETL action.",
            "intent": "reverse_etl",
            "availability": "implemented" if implemented else "planned",
            "write": action_name,
            "source_cli_path": f"{op_item['method']} {op_item['path']}",
            "source_url": SOURCE_DOC[op_item["source_id"]],
            "api_surface": [{"method": op_item["method"], "path": op_item["path"]}],
            "risk": action["risk"],
            "approval": "Plan first, inspect preview output, then execute only with the generated approval token" + (" and --confirm destructive" if action.get("confirm") == "destructive" else "") + ".",
            "examples": [f"pm chatwoot {cp} --credential chatwoot-prod --preview --json"],
        }
        if cmd["availability"] == "implemented":
            cmd["flags"] = flags_for_action(action)
        commands.append(cmd)
    group_objs = []
    for gid, paths in sorted(groups.items()):
        unique = []
        for p in paths:
            if p not in unique:
                unique.append(p)
        group_objs.append({"id": gid, "title": gid.replace("_", " ").title(), "commands": unique})
    cli = {
        "tagline": "Inspect, read, and safely plan typed Chatwoot account/platform operations.",
        "usage": "pm chatwoot <command> [flags]",
        "source_cli": {"name": "Chatwoot OpenAPI", "docs": SOURCES["chatwoot_application_openapi"], "reference": "application/platform/client/other tag groups", "source": "provider_api"},
        "groups": group_objs,
        "global_flags": [
            {"name": "credential", "type": "string", "summary": "Credential name for the Chatwoot request."},
            {"name": "config", "type": "string_array", "summary": "Connector config override as key=value; never pass secret values here."},
            {"name": "json", "type": "boolean", "summary": "Emit machine-readable JSON output."},
            {"name": "limit", "type": "integer", "summary": "Maximum records to emit from stream commands."},
            {"name": "preview", "type": "boolean", "summary": "Preview a reverse-ETL write without making a network mutation."},
            {"name": "approve", "type": "string", "summary": "Approval token required to execute a reverse-ETL plan."},
            {"name": "confirm", "type": "string", "summary": "Typed confirmation challenge for destructive/admin reverse-ETL writes."},
        ],
        "commands": commands,
        "help_topics": [
            {"name": "destructive-confirmation", "summary": "Chatwoot DELETE/destructive/admin operations are represented only as named actions with plan -> preview -> approval -> execute and typed --confirm destructive gates, or as blocked operation rows."},
            {"name": "public-client-api", "summary": "Public client API operations remain planned/blocked until a separate no-auth inbox/contact contract exists."},
        ],
    }
    write_json(DEFS / "cli_surface.json", cli)


def update_metadata(actions: list[dict[str, Any]]) -> None:
    metadata = load_json(DEFS / "metadata.json")
    metadata["description"] = "Reads Chatwoot account data and safely plans typed reverse ETL mutations across the official application and platform APIs. Public client/no-auth operations are recorded as planned until a separate safe contract exists."
    metadata["capabilities"]["write"] = True
    metadata["capabilities"]["query"] = False
    metadata["risk"] = {
        "read": "external Chatwoot API reads of account-scoped conversation, contact, inbox, label, team, agent, and message data; additional official reads are ledgered as planned until typed stream/direct-read fixtures exist",
        "write": f"{len(actions)} named Chatwoot reverse ETL actions; DELETE/destructive/admin actions require typed destructive confirmation plus plan, preview, approval, execute",
        "approval": "reverse ETL mutations require plan -> preview -> explicit approval token -> execute; destructive/admin actions additionally require --confirm destructive",
    }
    write_json(DEFS / "metadata.json", metadata)


def update_spec() -> None:
    spec = load_json(DEFS / "spec.json")
    props = spec.setdefault("properties", {})
    props["base_url"]["default"] = "https://app.chatwoot.com"
    props["base_url"]["description"] = "Chatwoot server root URL, without /api/v1/accounts."
    props["account_id"]["description"] = "Chatwoot account ID for application and platform account-scoped endpoints."
    props["api_access_token"]["description"] = "Chatwoot user or platform-app API access token supplied via env/stdin; never inline."
    write_json(DEFS / "spec.json", spec)


def docs_for(actions: list[dict[str, Any]], surface: dict[str, Any], operations: list[dict[str, Any]]) -> str:
    counts = Counter()
    for ep in surface["endpoints"]:
        if ep.get("covered_by", {}).get("stream"):
            counts["streams"] += 1
        elif ep.get("covered_by", {}).get("write"):
            counts["writes"] += 1
        elif ep.get("operation"):
            counts["blocked"] += 1
    destructive = sum(1 for a in actions if a.get("confirm") == "destructive")
    return f"""# Chatwoot connector\n\n## Overview\n\nThe Chatwoot connector reads account-scoped conversations, contacts, inboxes, labels, agents, teams, and messages from the official Chatwoot Application API. This wave also records every operation from the official application, platform, client, and other OpenAPI tag-group sources in `api_surface.json` and `operations.json`.\n\nOfficial operation ledger counts: {len(surface['endpoints'])} total rows; {counts['streams']} stream-covered rows; {counts['writes']} reverse-ETL write-covered rows; {counts['blocked']} planned/blocked rows.\n\n## Auth setup\n\nUse `pm credentials add <name> --connector chatwoot` with `--from-env api_access_token=ENV` or `--value-stdin api_access_token`. Do not paste tokens into chat, shell history, docs, fixtures, or JSON examples. `base_url` is the Chatwoot server root URL (default `https://app.chatwoot.com`) and `account_id` scopes application/platform account endpoints.\n\nThe public client API documented by Chatwoot does not use the same user/platform token contract; those operations are represented as planned/blocked until a separate no-auth inbox/contact-safe connector contract exists.\n\n## Streams notes\n\nImplemented fixture-backed streams remain account-scoped and bounded: `conversations`, `contacts`, `inboxes`, `agents`, `teams`, `labels`, and fan-out `messages`. Additional documented GET/report/search/changefeed operations are present exactly once in the ledger and remain planned/blocked until typed stream/direct-read schemas and fixtures are authored.\n\nThe connector uses bounded page-number pagination for paginated streams and fixture replay only in local conformance. No live provider call is made by this bundle.\n\n## Write actions & risks\n\nThis bundle declares {len(actions)} named reverse-ETL write actions for safely expressible Chatwoot application/platform mutations. Every write must follow plan -> preview -> explicit approval token -> execute. {destructive} DELETE/destructive/admin/elevated actions carry `confirm: destructive` and require the typed `--confirm destructive` challenge before execution. Delete actions are modeled as idempotent for 404 responses where the resource is already absent.\n\nPublic client writes and write endpoints requiring query-parameter execution support remain planned/blocked rather than exposed through a raw API escape hatch. There is no generic HTTP method/path/body tool.\n\n## Known limits\n\n- Live certification is not claimed; no credentials or provider writes were used.\n- Public client API operations need a separate no-auth, inbox/contact-bounded contract before execution.\n- `POST /api/v1/accounts/{{account_id}}/custom_filters` requires a query parameter on a write path; this connector-local slice records it as blocked because the shared write engine has no query-param write contract.\n- Additional GET/report/changefeed operations beyond the seven fixture-backed streams are operation-ledger rows until typed stream or direct-read fixtures prove safe execution.\n- Generated operation schemas are bounded metadata for planning and validation; provider-specific edge semantics still require live-safe certification before certification claims.\n"""


def update_docs(actions: list[dict[str, Any]], surface: dict[str, Any], operations: list[dict[str, Any]]) -> None:
    docs_md = docs_for(actions, surface, operations)
    (DEFS / "docs.md").write_text(docs_md)
    # Lightweight manual/SKILL docs stay Chatwoot-owned and mirror the generated connector guide.
    manual = f"""# pm connectors inspect chatwoot\n\n```text\nNAME\n  pm connectors inspect chatwoot - Chatwoot connector manual\n\nSYNOPSIS\n  pm connectors inspect chatwoot\n  pm connectors inspect chatwoot --json\n\nDESCRIPTION\n  Reads fixture-backed Chatwoot account streams and safely plans named reverse ETL actions. The operation ledger covers all {len(surface['endpoints'])} official Chatwoot OpenAPI operations.\n\nICON\n  asset: icons/pm-sample.svg\n  source: polymetrics\n  review_status: polymetrics\n  review_url: https://github.com/polymetrics-ai/cli\n\nSECURITY\n  No secret values belong in chat, shell history, docs, fixtures, or JSON output. Reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive/admin actions require --confirm destructive.\n\nIMPLEMENTED STREAMS\n  conversations, contacts, inboxes, agents, teams, labels, messages\n\nREVERSE ETL ACTIONS\n  {len(actions)} named actions are declared in writes.json; destructive/admin count: {sum(1 for a in actions if a.get('confirm') == 'destructive')}. Inspect structured connector JSON for the full action list.\n\nKNOWN LIMITS\n  Public client API, write-query gaps, and non-fixture-backed read/report/changefeed operations remain planned/blocked in api_surface.json with evidence. No live certification is claimed.\n\nAGENT WORKFLOW\n  Inspect first with --json. Never request credentials in chat. Use plan, preview, approval token, and --confirm destructive for destructive/admin write execution.\n\nEXIT STATUS\n  0 success\n  1 runtime error\n  2 usage error\n```\n"""
    skill = f"""---\nname: pm-chatwoot\ndescription: Safe Chatwoot connector usage for pm.\n---\n\n# Chatwoot connector\n\nUse `pm connectors inspect chatwoot --json` before selecting streams or write actions.\n\n## Agent Rules\n\n- Streams: conversations, contacts, inboxes, agents, teams, labels, messages.\n- Writes: {len(actions)} named reverse-ETL actions in `internal/connectors/defs/chatwoot/writes.json`.\n- Safety: plan -> preview -> explicit approval -> execute; destructive/admin actions require typed `--confirm destructive`.\n- Never ask for or print `api_access_token`; load it from environment or stdin.\n- Public client API and non-fixture-backed direct/report/changefeed operations remain planned/blocked in `api_surface.json`; do not route through a raw HTTP escape hatch.\n"""
    DOCS.mkdir(parents=True, exist_ok=True)
    (DOCS / "MANUAL.md").write_text(manual)
    (DOCS / "SKILL.md").write_text(skill)


def update_certification() -> None:
    cert = {
        "schema_version": 1,
        "source": {
            "default_stream": "conversations",
            "source_credential_defaults": {"base_url": "https://app.chatwoot.com", "account_id": "1"},
            "live_unavailable": [
                {"kind": "missing_credentials", "contains": ["api_access_token"]},
                {"kind": "not_certified", "contains": ["no live Chatwoot certification was run in wave02-r1"]},
            ],
        },
        "direct_read_candidates": [],
        "binary_candidates": [],
        "write_pairings": [],
    }
    write_json(DEFS / "certification.json", cert)


def main() -> None:
    specs = fetch_openapi()
    ops = iter_operations(specs)
    generate_streams()
    update_fixture_paths()
    actions, write_coverage = generate_writes(specs, ops)
    operations, operation_ids = generate_operations(specs, ops, write_coverage)
    generate_api_surface(ops, write_coverage, operation_ids)
    generate_cli_surface(ops, actions, write_coverage, operation_ids)
    update_metadata(actions)
    update_spec()
    update_certification()
    surface = load_json(DEFS / "api_surface.json")
    update_docs(actions, surface, operations)
    summary = {
        "official_operations": len(ops),
        "write_actions": len(actions),
        "destructive_or_admin_confirm_gated": sum(1 for a in actions if a.get("confirm") == "destructive"),
        "api_surface": Counter("covered" if e.get("covered_by") else "blocked" for e in surface["endpoints"]),
        "blocked_reasons": Counter(e.get("operation", {}).get("reason", "covered") for e in surface["endpoints"]),
    }
    (PHASE / "generation-summary.json").write_text(json.dumps(summary, indent=2, default=dict) + "\n")
    print(json.dumps(summary, indent=2, default=dict))


if __name__ == "__main__":
    main()
