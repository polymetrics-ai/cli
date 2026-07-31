#!/usr/bin/env python3
"""Generate the connector-local Crisp operation ledger bundle from official docs."""
from __future__ import annotations

import json
import pathlib
import re
import sys
from collections import Counter
from dataclasses import dataclass, field
from typing import Any

import requests
from bs4 import BeautifulSoup, Tag

DOCS_URL = "https://docs.crisp.chat/references/rest-api/v1/"
AUTH_GUIDE = "https://docs.crisp.chat/guides/rest-api/authentication/"
ROOT = pathlib.Path(__file__).resolve().parents[3]
OUT = ROOT / "internal/connectors/defs/crisp"
PHASE = ROOT / ".planning/phases/issue-204-crisp-parity-wave02-r1"

QUERY_CHUNK_RE = re.compile(r"\{[?&]([^}]+)\}")
PATH_VAR_RE = re.compile(r"\{([^?&}][^}]*)\}")

CDC_PATHS = {
    "/v1/website/{website_id}/conversation/{session_id}/events/{page_number}",
    "/v1/website/{website_id}/people/suggest/events/{page_number}",
    "/v1/website/{website_id}/people/events/{people_id}",
    "/v1/website/{website_id}/people/events/{people_id}/list/{page_number}",
}

BINARY_PATHS = {
    "/v1/website/{website_id}/conversation/{session_id}/files/{page_number}",
    "/v1/website/{website_id}/conversation/{session_id}/transcript",
    "/v1/website/{website_id}/people/export/profiles",
    "/v1/website/{website_id}/people/import/profiles",
    "/v1/website/{website_id}/helpdesk/locale/{locale}/import",
    "/v1/website/{website_id}/helpdesk/locale/{locale}/export",
    "/v1/bucket/url/generate",
    "/v1/media/animation/list/{page_number}{?per_page}{?search_query}{&list_id}",
}

DIRECT_PATHS = {
    "/v1/website/{website_id}/conversations/suggest/segments/{page_number}",
    "/v1/website/{website_id}/conversations/suggest/segment",
    "/v1/website/{website_id}/conversations/suggest/data/{page_number}",
    "/v1/website/{website_id}/conversations/suggest/data",
    "/v1/website/{website_id}/conversation/{session_id}/report",
    "/v1/website/{website_id}/people/suggest/segments/{page_number}",
    "/v1/website/{website_id}/people/suggest/segment",
    "/v1/website/{website_id}/people/suggest/data/{page_number}",
    "/v1/website/{website_id}/people/suggest/data",
    "/v1/website/{website_id}/people/suggest/event",
    "/v1/website/{website_id}/analytics/generate",
    "/v1/website/{website_id}/batch/report",
}

@dataclass
class Param:
    name: str
    raw_type: str
    required: bool
    description: str = ""
    enum: list[str] = field(default_factory=list)

@dataclass
class Operation:
    category: str
    title: str
    anchor: str
    method: str
    raw_path: str
    path: str
    uri_params: list[Param]
    body_params: list[Param]
    response_params: list[Param]
    description: str
    lane: str
    operation_id: str
    source_url: str


def slugify(value: str) -> str:
    value = value.lower()
    value = re.sub(r"[^a-z0-9]+", "-", value)
    value = value.strip("-")
    return value or "operation"


def canonical_path(raw: str) -> str:
    return QUERY_CHUNK_RE.sub("", raw)


def query_params(raw: str) -> set[str]:
    out: set[str] = set()
    for chunk in QUERY_CHUNK_RE.findall(raw):
        for part in chunk.split(","):
            part = part.strip()
            if part:
                out.add(part)
    return out


def path_vars(path: str) -> set[str]:
    return {m.group(1) for m in PATH_VAR_RE.finditer(path)}


def clean_text(text: str) -> str:
    text = text.replace("â­", "⭐")
    return re.sub(r"\s+", " ", text).strip()


def parse_param(key: Tag) -> Param:
    name_el = key.find("div", class_="request-format-path")
    type_el = key.find("div", class_="request-format-type")
    if name_el is None or type_el is None:
        return Param("unknown", "string", False)
    required = type_el.find("span", class_="request-format-required") is not None
    type_copy = BeautifulSoup(str(type_el), "html.parser")
    for span in type_copy.find_all("span"):
        span.extract()
    raw_type = clean_text(type_copy.get_text(" ")) or "string"
    label = key.find("div", class_="request-format-label")
    description = ""
    enum: list[str] = []
    if label is not None:
        # First paragraph is the useful prose summary. Remaining code blocks are enum values/examples.
        first_p = label.find("p")
        if first_p is not None:
            description = clean_text(first_p.get_text(" "))
        if raw_type.startswith("enum") or raw_type.startswith("array[enum"):
            enum = [clean_text(c.get_text(" ")) for c in label.find_all("code")]
            enum = [e for e in enum if e]
    return Param(clean_text(name_el.get_text(" ")), raw_type, required, description, enum)


def parse_params(fmt: Tag | None) -> tuple[str, list[Param], list[Param], list[Param]]:
    if fmt is None:
        return "", [], [], []
    description = ""
    sections: dict[str, list[Param]] = {"URI Parameters": [], "Request Body": [], "Response Data": []}
    current = ""
    for child in fmt.children:
        if not isinstance(child, Tag):
            continue
        classes = child.get("class", [])
        if "markdown" in classes and not description:
            description = clean_text(child.get_text(" "))
            continue
        if "request-format-title" in classes:
            current = clean_text(child.get_text(" "))
            sections.setdefault(current, [])
            continue
        if "request-format-keys" in classes and current:
            for key in child.find_all("div", class_="request-format-key", recursive=False):
                sections.setdefault(current, []).append(parse_param(key))
    return description, sections.get("URI Parameters", []), sections.get("Request Body", []), sections.get("Response Data", [])


def param_schema(p: Param) -> dict[str, Any]:
    raw = p.raw_type
    schema: dict[str, Any]
    if raw.startswith("array["):
        inner = raw[len("array["):-1]
        if inner.startswith("enum"):
            items: dict[str, Any] = {"type": "string"}
            if p.enum:
                items["enum"] = p.enum
        elif inner == "number":
            items = {"type": "number"}
        elif inner == "object":
            items = {"type": "object", "additionalProperties": True}
        else:
            items = {"type": "string"}
        schema = {"type": "array", "items": items}
    elif raw.startswith("enum"):
        typ = "number" if "number" in raw else "string"
        schema = {"type": typ}
        if p.enum and typ == "string":
            schema["enum"] = p.enum
    elif raw == "number":
        schema = {"type": "number"}
    elif raw == "boolean":
        schema = {"type": "boolean"}
    elif raw == "object":
        schema = {"type": "object", "additionalProperties": True}
    else:
        schema = {"type": "string"}
    if p.description:
        schema["description"] = p.description[:240]
    return schema


def object_schema(params: list[Param], include_optional: bool = True) -> dict[str, Any]:
    props = {p.name: param_schema(p) for p in params if include_optional or p.required}
    required = [p.name for p in params if p.required and p.name in props]
    schema: dict[str, Any] = {"type": "object", "properties": props, "additionalProperties": False}
    if required:
        schema["required"] = required
    return schema


def lane_for(method: str, raw_path: str) -> str:
    if method == "HEAD":
        return "head_check"
    if raw_path in CDC_PATHS:
        return "cdc_changefeed"
    if raw_path in BINARY_PATHS:
        return "binary_file"
    if raw_path in DIRECT_PATHS:
        return "direct_read_query_search"
    if method == "GET":
        return "etl_read"
    return "reverse_etl_write"


def risk_for(op: Operation) -> str:
    text = f"{op.title} {op.path} {op.category}".lower()
    if op.method == "DELETE":
        if op.path in {"/v1/website/{website_id}", "/v1/website/{website_id}/expunge"}:
            return "critical"
        return "high"
    if op.lane == "binary_file":
        return "high"
    if op.lane == "cdc_changefeed" and op.method != "GET":
        return "high"
    if op.method in {"POST", "PUT", "PATCH"}:
        if any(word in text for word in ["delete", "remove", "block", "unsubscribe", "dispatch", "bill", "data privacy", "expunge"]):
            return "high"
        return "medium" if op.lane == "direct_read_query_search" else "high"
    if any(word in text for word in ["conversation", "people", "profile", "analytics", "campaign", "subscription", "operator"]):
        return "medium"
    return "low"


def mutation_class(op: Operation) -> str:
    text = f"{op.title} {op.path}".lower()
    if op.method == "DELETE":
        return "destructive"
    if any(word in text for word in ["delete", "remove", "expunge", "unsubscribe", "block"]):
        return "destructive"
    if op.method == "POST" and any(word in text for word in ["create", "add", "send", "import", "dispatch", "subscribe", "redeem", "report"]):
        return "create"
    if op.method == "PUT":
        return "update"
    if op.method == "PATCH":
        return "update"
    return "update"


def api_model(op: Operation) -> str:
    if op.method == "DELETE":
        return "destructive_action"
    if op.method in {"POST", "PUT", "PATCH"}:
        if risk_for(op) in {"high", "critical"} or "admin" in op.category.lower():
            return "admin_reverse_etl" if "admin" in op.category.lower() else "sensitive_reverse_etl"
        return "sensitive_reverse_etl"
    if op.lane == "binary_file":
        return "binary_read"
    return "direct_read"


def blocked_reason(op: Operation) -> str:
    if op.method == "DELETE":
        return "Crisp DELETE/destructive operation is in scope but blocked until a named reverse-ETL action supplies schema, redaction, idempotency notes, plan -> preview -> explicit approval -> execute, and typed destructive confirmation."
    if op.method in {"POST", "PUT", "PATCH"}:
        return "Crisp non-read REST operation is in scope but blocked until a named typed action supplies schema, redaction, risk text, fixture evidence, and plan -> preview -> explicit approval -> execute."
    if op.lane == "etl_read":
        return "Crisp ETL/read operation is in scope but blocked until a connector-owned fixture-backed stream, schema, pagination/cursor policy, and conformance evidence are added."
    if op.lane == "cdc_changefeed":
        return "Crisp changefeed/event operation is in scope but blocked pending CDC/changefeed truthfulness foundations #2986/#2988 and connector-owned replay evidence."
    if op.lane == "direct_read_query_search":
        return "Crisp bounded provider search/query/direct operation is in scope but blocked pending shared provider search/query foundation #2985 and connector-owned typed command evidence."
    if op.lane == "binary_file":
        return "Crisp binary/file/import/export operation is in scope but blocked until a connector-owned fixed-target binary policy supplies size caps, path safety, redaction, and fixture evidence."
    if op.lane == "head_check":
        return "Crisp HEAD existence-check operation is documented in the current REST API but is non-data and blocked until a typed check/direct-read contract exists."
    return "Crisp operation is in scope but blocked until a connector-owned safe execution contract exists."


def operation_kind(op: Operation) -> str:
    if op.method == "HEAD":
        return "composite"
    if op.lane == "binary_file" and op.method == "GET":
        return "binary_download"
    if op.method == "GET":
        return "rest_read"
    return "rest_write"


def path_param_fields(op: Operation) -> list[Param]:
    vars_ = path_vars(op.path)
    return [p for p in op.uri_params if p.name in vars_]


def query_param_fields(op: Operation) -> list[Param]:
    q = query_params(op.raw_path)
    return [p for p in op.uri_params if p.name in q or (p.name not in path_vars(op.path) and p.name not in path_vars(op.raw_path))]


def rest_block(op: Operation) -> dict[str, Any]:
    block: dict[str, Any] = {"method": op.method, "path": op.path, "max_bytes": 1048576}
    qparams = query_param_fields(op)
    if qparams:
        # RESTOperationSpec.Query is map[string]string in the current loader.
        # Keep typed detail in CLI flags/body_schema; store compact query type hints here.
        block["query"] = {p.name: p.raw_type for p in qparams}
    if op.body_params:
        block["body_schema"] = object_schema(op.body_params)
        block["content_type"] = "application/json"
    elif op.method == "POST":
        # The current operations loader requires POST rest_read operations to
        # declare an explicit schema even when the provider docs show no body.
        block["body_schema"] = {"type": "object", "properties": {}, "additionalProperties": False}
        block["content_type"] = "application/json"
    return block


def operation_spec(op: Operation) -> dict[str, Any]:
    kind = operation_kind(op)
    risk = risk_for(op)
    spec: dict[str, Any] = {
        "id": op.operation_id,
        "kind": kind,
        "summary": f"{op.title} ({op.method} {op.path})",
        "description": op.description or f"Crisp {op.title} operation.",
        "source_url": op.source_url,
        "risk": "none" if risk == "low" and op.method == "HEAD" else risk,
        "approval": "none",
        "output_policy": "json_redacted",
        "auth_scopes": ["crisp_rest_api_v1", f"lane:{op.lane}"],
        "mutation_class": "none",
        "destructive": False,
        "audit_event": f"crisp.{op.lane}",
    }
    if kind == "composite":
        spec["composite"] = {"steps": [f"planned:{op.method} {op.path}"]}
    elif kind == "binary_download":
        spec["binary"] = {"method": "GET", "path": op.path, "max_bytes": 16777216, "allow_overwrite": False, "extract_archives": False}
        spec["output_policy"] = "binary_redacted"
    else:
        spec["rest"] = rest_block(op)
    if kind == "rest_write":
        spec["mutation_class"] = mutation_class(op)
        spec["destructive"] = spec["mutation_class"] in {"delete", "destructive", "admin"} or op.method == "DELETE"
        if spec["destructive"]:
            spec["approval"] = "reverse ETL plan -> preview -> explicit approval -> execute; typed destructive confirmation required"
        else:
            spec["approval"] = "reverse ETL plan -> preview -> explicit approval -> execute"
    elif op.method == "DELETE":
        spec["mutation_class"] = "destructive"
        spec["destructive"] = True
        spec["approval"] = "typed destructive confirmation required before any future execute path"
    if risk in {"high", "critical"} or spec.get("destructive"):
        spec["secret_sensitive"] = True
        spec["sensitive_policy"] = {
            "input_mode": "env_or_stdin",
            "redact_fields": ["token_id", "token_key", "website_id", "session_id", "people_id", "plugin_id"],
            "preflight": "No live execution in this bundle; future execution must validate typed fields and redact identifiers before preview.",
            "transform": "none",
            "approval_mode": "typed_confirmation",
        }
    return spec


def flag_type(p: Param) -> str:
    schema = param_schema(p)
    typ = schema.get("type")
    if typ == "boolean":
        return "boolean"
    if typ == "number":
        return "integer" if p.raw_type == "enum[number]" else "string"
    if typ == "array":
        return "string_array"
    if "enum" in schema:
        return "enum"
    return "string"


MUTATION_METHODS = {"DELETE", "PUT", "PATCH", "POST"}


def cli_flags(op: Operation) -> list[dict[str, Any]]:
    flags: list[dict[str, Any]] = []
    reverse = op.lane == "reverse_etl_write" or op.method in MUTATION_METHODS
    seen: set[str] = set()
    for p in path_param_fields(op):
        target = f"record.{p.name}" if reverse else f"path.{p.name}"
        flags.append(flag_for_param(p, target))
        seen.add((p.name, target))
    for p in query_param_fields(op):
        target = f"query.{p.name}"
        flags.append(flag_for_param(p, target))
        seen.add((p.name, target))
    for p in op.body_params:
        target = f"record.{p.name}" if reverse else f"body.{p.name}"
        key = (p.name, target)
        if key in seen:
            continue
        flags.append(flag_for_param(p, target))
        seen.add(key)
    return flags


def flag_for_param(p: Param, maps_to: str) -> dict[str, Any]:
    f: dict[str, Any] = {
        "name": slugify(p.name),
        "type": flag_type(p),
        "summary": p.description or f"Crisp parameter {p.name}",
        "maps_to": maps_to,
    }
    if p.enum and f["type"] == "enum":
        f["values"] = p.enum[:64]
    if f["type"] == "string":
        f["allow_empty"] = False
        low = f["summary"].lower() + " " + p.name.lower()
        if "iso 8601" in low or "date" in low or "timestamp" in low:
            f["format"] = "date-time"
    return f


def cli_command(op: Operation) -> dict[str, Any]:
    group = {
        "etl_read": "etl",
        "direct_read_query_search": "direct",
        "binary_file": "binary",
        "reverse_etl_write": "reverse",
        "cdc_changefeed": "changefeed",
        "head_check": "checks",
    }[op.lane]
    mutating = op.method in MUTATION_METHODS
    if op.lane == "reverse_etl_write":
        intent = "reverse_etl"
    elif op.lane in {"direct_read_query_search", "binary_file", "cdc_changefeed"}:
        intent = "direct_write" if mutating else "direct_read"
    elif op.lane == "head_check":
        intent = "direct_read"
    else:
        intent = "etl"
    op_spec = operation_spec(op)
    cmd: dict[str, Any] = {
        "path": f"{group} {slugify(op.title)}",
        "summary": op.title,
        "intent": intent,
        "availability": "planned",
        "operation": op.operation_id,
        "source_url": op.source_url,
        "flags": cli_flags(op),
        "risk": risk_for(op),
        "approval": op_spec["approval"] if (intent == "reverse_etl" or mutating) else "none",
        "notes": f"{op.method} {op.path}; lane={op.lane}; planned fixed Crisp operation metadata only, not a raw API escape hatch.",
    }
    if intent in {"etl", "direct_read"}:
        cmd["output_policy"] = "json_redacted"
    return cmd


def parse_operations() -> list[Operation]:
    html = requests.get(DOCS_URL, timeout=30).text
    soup = BeautifulSoup(html, "html.parser")
    ops: list[Operation] = []
    category = ""
    for el in soup.find_all(["h1", "h2"]):
        if el.name == "h1":
            text = clean_text(el.get_text(" "))
            if text and text != "REST API Reference (V1)":
                category = text
            continue
        spec = el.find_next_sibling("div", class_="request-specification")
        if spec is None:
            continue
        target = spec.find("div", class_="request-target")
        if target is None:
            continue
        lines = [clean_text(x) for x in target.get_text("\n", strip=True).split("\n") if clean_text(x)]
        if len(lines) < 2:
            continue
        method = lines[0].upper()
        if method not in {"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}:
            continue
        raw_path = "".join(lines[1:]).replace(" ", "")
        path = canonical_path(raw_path)
        title = clean_text(el.get_text(" ")).removeprefix("⭐ ")
        anchor = el.get("id") or slugify(title)
        description, uri, body, response = parse_params(spec.find("div", class_="request-format"))
        lane = lane_for(method, raw_path)
        op_id = "crisp." + anchor
        ops.append(Operation(category, title, anchor, method, raw_path, path, uri, body, response, description, lane, op_id, DOCS_URL + "#" + anchor))
    seen: set[tuple[str, str]] = set()
    for op in ops:
        key = (op.method, op.path)
        if key in seen:
            raise RuntimeError(f"duplicate operation {key}")
        seen.add(key)
    return ops


def write_json(path: pathlib.Path, value: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, indent=2, ensure_ascii=False) + "\n")


def main() -> int:
    ops = parse_operations()
    counts = Counter(op.lane for op in ops)
    method_counts = Counter(op.method for op in ops)
    category_counts = Counter(op.category for op in ops)
    OUT.mkdir(parents=True, exist_ok=True)

    metadata = {
        "name": "crisp",
        "display_name": "Crisp",
        "description": "Connector-local Crisp REST API V1 official operation ledger and planned typed command surface.",
        "integration_type": "api",
        "docs_url": DOCS_URL,
        "release_stage": "planned",
        "capabilities": {"check": False, "read": False, "write": False, "query": False, "cdc": False, "dynamic_schema": False},
        "batch": {"read_page_size": 20, "write_batch_size": 1},
        "rate_limit": {"requests_per_minute": 60},
        "risk": {
            "read": "No Crisp read execution is enabled in this connector-local ledger slice; official read, direct, binary, and changefeed operations are inventoried as blocked/planned rows.",
            "write": "No Crisp write execution is enabled in this connector-local ledger slice; DELETE/destructive/admin writes are included as typed destructive-confirmation candidates, not exclusions.",
            "approval": "Future reverse ETL actions must use plan -> preview -> explicit approval -> execute; destructive/admin actions must require typed destructive confirmation.",
        },
    }
    spec = {
        "$schema": "http://json-schema.org/draft-07/schema#",
        "title": "Crisp Connection Specification",
        "type": "object",
        "required": [],
        "properties": {
            "base_url": {"type": "string", "format": "uri", "default": "https://api.crisp.chat", "description": "Crisp REST API base URL."},
            "tier": {"type": "string", "enum": ["website", "plugin", "user"], "default": "website", "description": "Crisp X-Crisp-Tier header value for future typed operations."},
            "token_id": {"type": "string", "x-secret": True, "description": "Crisp token identifier supplied through credential storage, environment, or stdin; never inline in docs or issue text."},
            "token_key": {"type": "string", "x-secret": True, "description": "Crisp token key supplied through credential storage, environment, or stdin; never inline in docs or issue text."},
            "website_id": {"type": "string", "description": "Optional Crisp website/workspace identifier for future typed commands."},
            "plugin_id": {"type": "string", "description": "Optional Crisp plugin identifier for future typed plugin commands."},
            "start_date": {"type": "string", "format": "date-time", "description": "Optional RFC3339 lower bound for future incremental streams."},
            "page_size": {"type": "string", "default": "20", "description": "Optional Crisp page-size request value for future paginated operations; provider docs commonly allow 20-50."},
            "max_pages": {"type": "string", "default": "0", "description": "Maximum pages for future typed reads; 0 means unbounded under PM output limits."},
        },
    }
    streams = {
        "base": {
            "url": "{{ config.base_url }}",
            "user_agent": "polymetrics-go-cli",
            "headers": {"Accept": "application/json", "X-Crisp-Tier": "{{ config.tier }}"},
            "auth": [
                {"mode": "basic", "username": "{{ secrets.token_id }}", "password": "{{ secrets.token_key }}", "when": "{{ secrets.token_key }}"},
                {"mode": "none"},
            ],
            "pagination": {"type": "none"},
            "error_map": [
                {"status": 401, "hint": "Crisp credential is missing or expired; re-add it without printing token values."},
                {"status": 403, "hint": "Crisp credential lacks the required tier or scope for this operation."},
                {"status": 429, "class": "rate_limited", "hint": "Crisp rate limit reached; retry after the provider window resets."},
            ],
        },
        "streams": [],
    }
    api_surface = {
        "api": "Crisp REST API V1",
        "docs": DOCS_URL,
        "reviewed_at": "2026-07-31",
        "operation_ledger_version": 1,
        "scope": "Complete current Crisp REST API V1 documentation inventory parsed from the official reference. Current docs expose 234 method/path rows: the parent r2 audit allocated 220 non-HEAD rows and this ledger additionally records 14 HEAD existence-check rows as planned/blocked non-data operations. DELETE/destructive/admin operations are included in scope and are not blanket-excluded.",
        "endpoints": [],
    }
    for op in ops:
        api_surface["endpoints"].append({
            "method": op.method,
            "path": op.path,
            "operation": {
                "model": api_model(op),
                "status": "blocked",
                "risk": risk_for(op),
                "blocked_by_default": True,
                "reason": blocked_reason(op),
                "source_url": op.source_url,
                "notes": f"lane={op.lane}; operation_id={op.operation_id}; title={op.title}; category={op.category}; documented_path={op.raw_path}",
            },
        })
    operations = {"operations": [operation_spec(op) for op in ops]}
    cli_surface = {
        "tagline": "Review planned typed Crisp REST API operations without exposing a raw API escape hatch.",
        "usage": "pm crisp <etl|direct|binary|changefeed|reverse|checks> <operation> [flags]",
        "source_cli": {"name": "Crisp REST API V1", "docs": DOCS_URL, "reference": "official docs fetched 2026-07-31", "source": "provider_api"},
        "groups": [
            {"id": "etl", "title": "ETL/read operations (planned)", "commands": ["etl"]},
            {"id": "direct", "title": "Direct/provider search/query operations (planned)", "commands": ["direct"]},
            {"id": "binary", "title": "Binary/file/import/export operations (planned)", "commands": ["binary"]},
            {"id": "changefeed", "title": "Changefeed/event operations (planned)", "commands": ["changefeed"]},
            {"id": "reverse", "title": "Reverse ETL write operations (planned)", "commands": ["reverse"]},
            {"id": "checks", "title": "HEAD/existence checks (planned)", "commands": ["checks"]},
        ],
        "global_flags": [
            {"name": "credential", "type": "string", "summary": "Credential profile name; never pass token values as flags.", "maps_to": "config.credential", "allow_empty": False},
            {"name": "limit", "type": "integer", "summary": "Maximum PM records/items to emit when a future typed command supports pagination.", "maps_to": "query.limit"},
            {"name": "json", "type": "boolean", "summary": "Render machine-readable output for future implemented commands."},
        ],
        "commands": [cli_command(op) for op in ops],
        "help_topics": [
            {"name": "crisp safety", "summary": "Crisp planned operations require typed fixed-target commands; no raw API escape hatch is available."},
            {"name": "crisp reverse-etl", "summary": "Future Crisp writes must use plan -> preview -> explicit approval -> execute, with typed destructive confirmation for destructive/admin actions."},
        ],
    }
    certification = {"schema_version": 1, "direct_read_candidates": [], "binary_candidates": [], "write_pairings": []}

    docs = f"""# Crisp connector

## Overview

This bundle is the connector-local Crisp REST API V1 official operation ledger for parent issue #204. It inventories the official Crisp REST API reference at `{DOCS_URL}` and deduplicates operations by HTTP method and canonical path.

The current documentation parse found {len(ops)} documented method/path rows. Method counts: {', '.join(f'{k} {v}' for k, v in sorted(method_counts.items()))}. Lane counts: ETL/read {counts['etl_read']}, direct/provider-search/query {counts['direct_read_query_search']}, binary/file {counts['binary_file']}, changefeed/events {counts['cdc_changefeed']}, reverse/write {counts['reverse_etl_write']}, HEAD checks {counts['head_check']}. The parent r2 audit allocated 220 non-HEAD rows; this ledger additionally records the 14 current official HEAD existence checks as planned/blocked non-data operations.

No local row is claimed as implemented, fixture-tested, certified, or live-safe in this wave. `metadata.json` keeps `check`, `read`, `write`, `query`, and `cdc` false until executable connector-local evidence exists.

## Auth setup

Crisp REST API authentication uses Basic authentication with a token keypair plus the `X-Crisp-Tier` header. The official authentication guides document website-token requests as `Authorization: Basic BASE64(token_id:token_key)` with `X-Crisp-Tier: website`, and plugin-token requests as `Authorization: Basic BASE64(identifier:key)` with `X-Crisp-Tier: plugin`.

Future executable Crisp commands should use a credential profile containing the token identifier and token key. Do not pass token values in prompt text, command examples, issue bodies, fixtures, or logs. Both `token_id` and `token_key` are marked `x-secret` in `spec.json`; this ledger slice does not perform authenticated provider calls.

## Streams notes

No Crisp ETL streams are enabled yet. The {counts['etl_read']} ETL/read rows and {counts['cdc_changefeed']} changefeed/event rows are represented in `api_surface.json`, `operations.json`, and planned command metadata as blocked rows until future connector-local lanes add named streams, schemas, pagination/cursor policy, sanitized fixtures, and conformance evidence.

Provider search/query/direct operations remain blocked pending shared foundation #2985 and must stay distinct from warehouse-focused `pm query`. CDC/changefeed rows remain blocked pending #2986/#2988 and connector-owned replay evidence.

## Write actions & risks

No Crisp reverse ETL write actions are enabled yet. The {counts['reverse_etl_write']} reverse/write rows include DELETE, destructive, sensitive, and admin operations in scope. They are not blanket-excluded as unsafe. A future executable action must be named, schema-gated, redacted, risk documented, and routed through plan -> preview -> explicit approval -> execute; DELETE/destructive/admin actions must additionally require typed destructive confirmation such as `confirm: "destructive"` in `writes.json` plus idempotency notes where the provider supports safe retry/missing-resource semantics.

Binary/file/import/export rows are also blocked until a connector-owned fixed-target policy supplies size caps, path safety, redaction, and fixture/conformance evidence. No raw method/path/body, arbitrary GraphQL, shell, unrestricted file, generic SQL, or passthrough API command is exposed by this bundle.

## Known limits

- This is a complete documented operation ledger and planned typed command surface, not completed runtime parity.
- Shared provider search/query foundation #2985 is open, so search/query/direct commands remain planned and blocked.
- CDC truth/lab foundations #2986 and #2988 are open, so changefeed/event operations remain planned and blocked.
- `operations.json` preserves typed top-level request-body schemas parsed from the documentation, but nested provider-specific object fields remain bounded generic objects until future executable lanes add full schemas and fixtures.
- No live credentials, live provider calls, live writes, certification, VPS/Thaalam work, or merge activity were used to produce this bundle. Fixture-backed dynamic conformance has no executable stream or write action to run in this wave.
"""

    write_json(OUT / "metadata.json", metadata)
    write_json(OUT / "spec.json", spec)
    write_json(OUT / "streams.json", streams)
    write_json(OUT / "api_surface.json", api_surface)
    write_json(OUT / "operations.json", operations)
    write_json(OUT / "cli_surface.json", cli_surface)
    write_json(OUT / "certification.json", certification)
    (OUT / "docs.md").write_text(docs)

    sources = {
        "docs_url": DOCS_URL,
        "auth_guide": AUTH_GUIDE,
        "operation_count": len(ops),
        "method_counts": dict(sorted(method_counts.items())),
        "lane_counts": dict(sorted(counts.items())),
        "category_counts": dict(sorted(category_counts.items())),
        "head_rows_note": "Current official docs expose 14 HEAD existence-check rows in addition to the parent r2 audit's 220 non-HEAD rows.",
        "shared_foundation_blockers": ["#2985 provider search/query boundary", "#2986 CDC truthfulness", "#2988 CDC state/lab"],
    }
    write_json(PHASE / "crisp-source-inventory.json", sources)
    (PHASE / "SOURCES.md").write_text(
        "# SOURCES — Crisp parity wave02 r1\n\n"
        f"Official REST API reference: {DOCS_URL}\n\n"
        f"Authentication guide: {AUTH_GUIDE}\n\n"
        f"Generated operation rows: {len(ops)}. Method counts: {dict(sorted(method_counts.items()))}. Lane counts: {dict(sorted(counts.items()))}.\n\n"
        "The official documentation parse found 14 HEAD existence-check rows in addition to the parent r2 audit's 220 non-HEAD operation allocation. The GitHub issue count tables were preserved; the connector-local ledger records the HEAD rows as planned/blocked non-data operations.\n\n"
        "No live provider calls, credentials, writes, certification, VPS/Thaalam work, push, PR, or merge were used.\n"
    )
    print(f"wrote {OUT} with {len(ops)} operations; lanes={dict(sorted(counts.items()))}; methods={dict(sorted(method_counts.items()))}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
