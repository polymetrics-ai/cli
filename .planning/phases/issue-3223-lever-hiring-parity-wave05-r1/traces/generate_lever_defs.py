#!/usr/bin/env python3
"""Generate connector-local Lever Hiring parity definitions from audited operation inventory."""
from __future__ import annotations

import json
import os
import re
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[4]
DEF = ROOT / "internal/connectors/defs/lever-hiring"
PHASE = ROOT / ".planning/phases/issue-3223-lever-hiring-parity-wave05-r1"
HTTP_OPS = json.loads((PHASE / "sources/lever-official-http-ops.json").read_text())
DOCS_URL = "https://hire.lever.co/developer/documentation"

WEBHOOK_EVENTS = [
    "applicationCreated",
    "candidateHired",
    "candidateStageChange",
    "candidateArchiveChange",
    "candidateDeleted",
    "interviewCreated",
    "interviewUpdated",
    "interviewDeleted",
    "contactCreated",
    "contactUpdated",
]


def api_path(path: str) -> str:
    return re.sub(r":([A-Za-z_][A-Za-z0-9_]*)", r"{\1}", path)


def action_path(path: str) -> tuple[str, list[str]]:
    fields: list[str] = []

    def repl(match: re.Match[str]) -> str:
        name = match.group(1)
        fields.append(name)
        return "{{ record.%s }}" % name

    return re.sub(r":([A-Za-z_][A-Za-z0-9_]*)", repl, path), fields


def slug(value: str) -> str:
    value = value.strip().lower().replace("&", " and ")
    value = re.sub(r"[^a-z0-9]+", "_", value).strip("_")
    return value or "operation"


def kebab(value: str) -> str:
    return slug(value).replace("_", "-")


FIELD_VALUE_ITEM_SCHEMA: dict[str, Any] = {
    "type": "object",
    "required": ["id", "value"],
    "additionalProperties": False,
    "properties": {
        "id": {"type": "string"},
        "value": {
            "type": ["string", "number", "integer", "boolean", "array", "null"],
            "items": {
                "type": ["string", "number", "integer", "boolean", "object", "null"],
                "additionalProperties": False,
                "properties": {
                    "score": {"type": ["integer", "number", "null"]},
                    "comment": {"type": ["string", "null"]},
                    "text": {"type": ["string", "null"]},
                },
            },
        },
    },
}

FORM_FIELD_ITEM_SCHEMA: dict[str, Any] = {
    "type": "object",
    "required": ["type", "text"],
    "additionalProperties": False,
    "properties": {
        "id": {"type": ["string", "null"]},
        "type": {
            "type": "string",
            "enum": [
                "application-file-upload",
                "code",
                "currency",
                "date",
                "dropdown",
                "file-upload",
                "multiple choice",
                "multiple-choice",
                "multiple select",
                "multiple-select",
                "note",
                "score",
                "score system",
                "score-system",
                "scorecard",
                "text",
                "textarea",
                "university",
                "yes-no",
                "yes/no",
            ],
        },
        "text": {"type": "string"},
        "description": {"type": ["string", "null"]},
        "required": {"type": "boolean"},
        "prompt": {"type": ["string", "null"]},
        "options": {
            "type": "array",
            "items": {
                "type": "object",
                "required": ["text"],
                "additionalProperties": False,
                "properties": {
                    "text": {"type": "string"},
                    "optionId": {"type": ["string", "null"]},
                },
            },
        },
        "scores": {
            "type": "array",
            "items": {
                "type": "object",
                "required": ["text"],
                "additionalProperties": False,
                "properties": {
                    "text": {"type": "string"},
                    "description": {"type": ["string", "null"]},
                },
            },
        },
    },
}


def prop_schema(name: str) -> dict[str, Any]:
    if name == "fieldValues":
        return {"type": "array", "items": FIELD_VALUE_ITEM_SCHEMA}
    if name == "fields":
        return {"type": "array", "items": FORM_FIELD_ITEM_SCHEMA}
    if name in {"links", "tags", "sources", "postingIds", "options", "subfields"}:
        return {"type": "array", "items": {"type": ["object", "string", "number", "integer", "boolean", "null"]}}
    if name in {"secretByDefault", "backfill", "isRequired"}:
        return {"type": "boolean"}
    if name.endswith("At") or name in {"headcountTotal", "headcountHired", "min", "max"}:
        return {"type": ["integer", "number", "null"]}
    return {"type": ["string", "null"]}


def generic_stream_schema(name: str, extra_props: list[str] | None = None) -> dict[str, Any]:
    props: dict[str, Any] = {
        "id": {"type": "string"},
        "text": {"type": ["string", "null"]},
        "name": {"type": ["string", "null"]},
        "createdAt": {"type": ["integer", "null"]},
        "updatedAt": {"type": ["integer", "null"]},
    }
    for prop in extra_props or []:
        props[prop] = {"type": ["string", "null"]}
    return {
        "$schema": "http://json-schema.org/draft-07/schema#",
        "title": name,
        "type": "object",
        "x-primary-key": ["id"],
        "required": ["id"],
        "properties": props,
    }


def stream_record(name: str, **extra: str) -> dict[str, Any]:
    record: dict[str, Any] = {
        "id": f"{name}_fixture_1",
        "text": f"Fixture {name.replace('_', ' ')}",
        "name": f"Fixture {name.replace('_', ' ')}",
        "createdAt": 1767225601000,
        "updatedAt": 1767225602000,
    }
    record.update(extra)
    return record


def page(path: str, data: list[dict[str, Any]], query: dict[str, str] | None = None) -> dict[str, Any]:
    request: dict[str, Any] = {"method": "GET", "path": path}
    if query:
        request["query"] = query
    return {
        "request": request,
        "response": {"status": 200, "body": {"data": data, "hasNext": False, "next": None}},
    }


# Stream definitions. Existing streams stay first to preserve the current mandatory fixture behavior.
streams: list[dict[str, Any]] = [
    {"name": "opportunities", "path": "/opportunities", "query": {"limit": "100"}, "records": {"path": "data"}, "schema": "schemas/opportunities.json"},
    {"name": "postings", "path": "/postings", "query": {"limit": "100"}, "records": {"path": "data"}, "schema": "schemas/postings.json"},
    {"name": "users", "path": "/users", "query": {"limit": "100"}, "records": {"path": "data"}, "schema": "schemas/users.json"},
    {"name": "requisitions", "path": "/requisitions", "query": {"limit": "100"}, "records": {"path": "data"}, "schema": "schemas/requisitions.json"},
    {"name": "stages", "path": "/stages", "query": {"limit": "100"}, "records": {"path": "data"}, "schema": "schemas/stages.json"},
]

simple_streams = [
    ("deleted_applications", "/applications/deleted"),
    ("archive_reasons", "/archive_reasons"),
    ("audit_events", "/audit_events"),
    ("disposition_stages", "/disposition_stages"),
    ("feedback_templates", "/feedback_templates"),
    ("deleted_opportunities", "/opportunities/deleted"),
    ("deleted_postings", "/postings/deleted"),
    ("form_templates", "/form_templates"),
    ("requisition_fields", "/requisition_fields"),
    ("sources", "/sources"),
]

fanout_streams = [
    ("applications", "/opportunities/{{ fanout.id }}/applications", "/opportunities", "opportunity_id"),
    ("opportunity_feedback", "/opportunities/{{ fanout.id }}/feedback", "/opportunities", "opportunity_id"),
    ("opportunity_interviews", "/opportunities/{{ fanout.id }}/interviews", "/opportunities", "opportunity_id"),
    ("opportunity_notes", "/opportunities/{{ fanout.id }}/notes", "/opportunities", "opportunity_id"),
    ("opportunity_offers", "/opportunities/{{ fanout.id }}/offers", "/opportunities", "opportunity_id"),
    ("opportunity_file_actions", "/opportunities/{{ fanout.id }}/file_actions", "/opportunities", "opportunity_id"),
    ("opportunity_panels", "/opportunities/{{ fanout.id }}/panels", "/opportunities", "opportunity_id"),
    ("opportunity_forms", "/opportunities/{{ fanout.id }}/forms", "/opportunities", "opportunity_id"),
    ("opportunity_referrals", "/opportunities/{{ fanout.id }}/referrals", "/opportunities", "opportunity_id"),
    ("posting_users", "/postings/{{ fanout.id }}/users", "/postings", "posting_id"),
]

for name, path_value in simple_streams:
    streams.append({
        "name": name,
        "path": path_value,
        "query": {"limit": "100"},
        "records": {"path": "data"},
        "projection": "passthrough",
        "schema": f"schemas/{name}.json",
    })

for name, path_value, parent_path, stamp in fanout_streams:
    streams.append({
        "name": name,
        "path": path_value,
        "query": {"limit": "100"},
        "records": {"path": "data"},
        "projection": "passthrough",
        "fan_out": {
            "ids_from": {"request": {"path": parent_path, "records_path": "data", "id_field": "id"}},
            "into": {"path_var": "id"},
            "stamp_field": stamp,
        },
        "schema": f"schemas/{name}.json",
    })

streams_json = {
    "base": {
        "url": "{{ config.base_url }}",
        "user_agent": "polymetrics-go-cli",
        "auth": [
            {"mode": "bearer", "token": "{{ secrets.access_token }}", "when": "{{ secrets.access_token }}"},
            {"mode": "basic", "username": "{{ secrets.api_key }}", "password": "", "when": "{{ secrets.api_key }}"},
        ],
        "pagination": {"type": "cursor", "cursor_param": "offset", "token_path": "next", "stop_path": "hasNext", "page_size": 100},
        "check": {"method": "GET", "path": "/postings", "query": {"limit": "1"}},
        "error_map": [
            {"status": 401, "hint": "lever-hiring credentials are missing or invalid; re-run pm credentials add lever-hiring"},
            {"status": 403, "hint": "lever-hiring account lacks access for this request"},
            {"status": 429, "class": "rate_limited", "hint": "lever-hiring rate limit exceeded; the connector will retry with backoff"},
        ],
    },
    "streams": streams,
}

(DEF / "streams.json").write_text(json.dumps(streams_json, indent=2) + "\n")

# Schemas and fixtures for new streams.
for name, _path in simple_streams:
    (DEF / "schemas" / f"{name}.json").write_text(json.dumps(generic_stream_schema(name), indent=2) + "\n")
    d = DEF / "fixtures/streams" / name
    d.mkdir(parents=True, exist_ok=True)
    (d / "page_1.json").write_text(json.dumps(page(_path, [stream_record(name)], {"limit": "100"}), indent=2) + "\n")

for name, path_value, parent_path, stamp in fanout_streams:
    (DEF / "schemas" / f"{name}.json").write_text(json.dumps(generic_stream_schema(name, [stamp]), indent=2) + "\n")
    d = DEF / "fixtures/streams" / name
    d.mkdir(parents=True, exist_ok=True)
    parent_id = "posting_fixture_1" if parent_path == "/postings" else "opportunity_fixture_1"
    child_path = path_value.replace("{{ fanout.id }}", parent_id)
    (d / "page_1.json").write_text(json.dumps(page(parent_path, [{"id": parent_id}], None), indent=2) + "\n")
    (d / "page_2.json").write_text(json.dumps(page(child_path, [stream_record(name, **{stamp: parent_id})], {"limit": "100"}), indent=2) + "\n")

# Supported write actions: no raw query parameters, no multipart/file bytes, closed schemas.
write_specs = [
    ("update_feedback", "update", "PUT", "/opportunities/:opportunity/feedback/:feedback", ["completedAt", "fieldValues"], "Updates a feedback form for an opportunity."),
    ("delete_feedback", "delete", "DELETE", "/opportunities/:opportunity/feedback/:feedback", [], "Deletes a feedback form from an opportunity."),
    ("create_feedback_template", "create", "POST", "/feedback_templates", ["text", "instructions", "group", "fields"], "Creates a feedback template."),
    ("update_feedback_template", "update", "PUT", "/feedback_templates/:feedback_template", ["text", "instructions", "group", "fields"], "Updates a feedback template."),
    ("delete_feedback_template", "delete", "DELETE", "/feedback_templates/:feedback_template", [], "Deletes a feedback template."),
    ("create_form_template", "create", "POST", "/form_templates", ["text", "instructions", "group", "secretByDefault", "fields"], "Creates a profile form template."),
    ("update_form_template", "update", "PUT", "/form_templates/:form_template", ["text", "instructions", "group", "secretByDefault", "fields"], "Updates a profile form template."),
    ("delete_form_template", "delete", "DELETE", "/form_templates/:form_template", [], "Deletes a profile form template."),
    ("delete_note", "delete", "DELETE", "/opportunities/:opportunity/notes/:noteId", [], "Deletes a note from an opportunity."),
    ("delete_requisition", "delete", "DELETE", "/requisitions/:requisition", [], "Deletes a requisition."),
    ("delete_requisition_field_options", "delete", "DELETE", "/requisition_fields/:requisition_field/options", [], "Deletes dropdown options for a requisition field."),
    ("delete_requisition_field", "delete", "DELETE", "/requisition_fields/:requisition_field", [], "Deletes a requisition field."),
    ("deactivate_user", "update", "POST", "/users/:user/deactivate", [], "Deactivates a Lever user."),
    ("reactivate_user", "update", "POST", "/users/:user/reactivate", [], "Reactivates a Lever user."),
]

actions = []
write_fixtures = DEF / "fixtures/writes"
write_fixtures.mkdir(parents=True, exist_ok=True)
for name, kind, method, raw_path, body_fields, risk in write_specs:
    resolved_path, path_fields = action_path(raw_path)
    props = {field: {"type": "string", "pattern": "^[A-Za-z0-9._-]+$"} for field in path_fields}
    for field in body_fields:
        props[field] = prop_schema(field)
    required = list(path_fields)
    if kind == "create" and "text" in body_fields:
        required.append("text")
    if kind == "create" and "fields" in body_fields:
        required.append("fields")
    schema: dict[str, Any] = {"type": "object", "required": required, "additionalProperties": False, "properties": props}
    if body_fields and kind != "create":
        schema["minProperties"] = len(path_fields) + 1
    action: dict[str, Any] = {
        "name": name,
        "kind": kind,
        "method": method,
        "path": resolved_path,
        "path_fields": path_fields,
        "record_schema": schema,
        "risk": risk + " Reverse ETL writes require plan, preview, explicit approval, and execute.",
    }
    if kind == "delete" or name == "deactivate_user":
        action["body_type"] = "none"
        action["confirm"] = "destructive"
    actions.append(action)

    record: dict[str, Any] = {field: f"{field}_fixture_1" for field in path_fields}
    for field in body_fields:
        if field == "fieldValues":
            record[field] = [{"id": "field_fixture_1", "value": "Fixture value"}]
        elif field == "fields" and "feedback_template" in name:
            record[field] = [{"type": "score-system", "text": "Rating", "required": True}]
        elif field == "fields":
            record[field] = [{
                "type": "date",
                "text": "Start Date",
                "description": "Please enter a desired start date.",
                "required": True,
            }]
        elif field == "secretByDefault":
            record[field] = False
        elif field.endswith("At"):
            record[field] = 1767225601000
        else:
            record[field] = f"Fixture {field}"
    expected_path = raw_path
    for field in path_fields:
        expected_path = expected_path.replace(":" + field, record[field])
    expect: dict[str, Any] = {"method": method, "path": expected_path}
    body = {k: v for k, v in record.items() if k not in path_fields}
    if body:
        expect["body"] = body
    (write_fixtures / f"{name}.json").write_text(json.dumps({"record": record, "expect": expect}, indent=2) + "\n")

(DEF / "writes.json").write_text(json.dumps({"actions": actions}, indent=2) + "\n")

# Direct-read operations/commands for fixed JSON GETs that are not stream-backed, binary/file-family, or webhook lifecycle.
stream_api_paths = {api_path(p) for _, p in simple_streams}
stream_api_paths.update({"/opportunities", "/postings", "/users", "/requisitions", "/stages"})
stream_api_paths.update({api_path(p.replace("{{ fanout.id }}", ":opportunity")) for _, p, parent, _ in fanout_streams if parent == "/opportunities"})
stream_api_paths.add("/postings/{posting}/users")
write_api_keys = {(method, api_path(path)) for _name, _kind, method, path, _body, _risk in write_specs}

def is_binary_family(path: str) -> bool:
    return "/files" in path or "/resumes" in path or path.endswith("/download") or path == "/uploads"

def is_webhook(path: str) -> bool:
    return path.startswith("/webhooks")


def sensitive_direct_read_policy(path: str) -> dict[str, Any] | None:
    if path == "/v1/eeo/responses/pii":
        return {"redact_fields": [
            "eeoResponses",
            "applicationArchivedAt",
            "applicationArchivedBy",
            "appliedAt",
            "currentStage",
            "contact",
            "gender",
            "race",
            "veteran",
            "disability",
            "disabilitySignatureDate",
            "eeoSurveyRespondedAt",
            "hiredDate",
            "hiringManager",
            "opportunityId",
            "origin",
            "posting",
            "requisitionCodes",
            "source",
        ]}
    if path.startswith("/surveys/diversity/"):
        return {"redact_fields": [
            "candidateLocations",
            "countryCodes",
            "fields",
            "instructions",
            "text",
            "options",
            "gender",
            "ethnicity",
            "race",
            "veteran",
            "disability",
            "sexualOrientation",
            "genderIdentity",
        ]}
    return None


direct_ops: list[dict[str, Any]] = []
cli_commands: list[dict[str, Any]] = []
api_rows: list[dict[str, Any]] = []

# Commands for streams.
for st in streams:
    cmd_path = f"{kebab(st['name'])} list"
    cli_commands.append({
        "path": cmd_path,
        "summary": f"Read Lever Hiring {st['name'].replace('_', ' ')} as ETL records.",
        "intent": "etl",
        "availability": "implemented",
        "stream": st["name"],
        "api_surface": [{"method": "GET", "path": next((api_path(p) for n, p in simple_streams if n == st['name']), None) or ""}],
        "examples": [f"pm lever-hiring {cmd_path} --json --limit 25"],
    })
# Fix api_surface for existing/fanout stream commands.
stream_command_paths: dict[str, str] = {
    "opportunities": "/opportunities", "postings": "/postings", "users": "/users", "requisitions": "/requisitions", "stages": "/stages",
    "applications": "/opportunities/{opportunity}/applications",
    "opportunity_feedback": "/opportunities/{opportunity}/feedback",
    "opportunity_interviews": "/opportunities/{opportunity}/interviews",
    "opportunity_notes": "/opportunities/{opportunity}/notes",
    "opportunity_offers": "/opportunities/{opportunity}/offers",
    "opportunity_file_actions": "/opportunities/{opportunity}/file_actions",
    "opportunity_panels": "/opportunities/{opportunity}/panels",
    "opportunity_forms": "/opportunities/{opportunity}/forms",
    "opportunity_referrals": "/opportunities/{opportunity}/referrals",
    "posting_users": "/postings/{posting}/users",
}
for cmd in cli_commands:
    if cmd["stream"] in stream_command_paths:
        cmd["api_surface"] = [{"method": "GET", "path": stream_command_paths[cmd["stream"]]}]

# API rows in official order.
for op in HTTP_OPS:
    method, raw_path = op["method"], op["path"]
    path = api_path(raw_path)
    key = (method, path)
    title = op.get("title", f"{method} {path}")
    source_url = f"{DOCS_URL}#{op.get('id')}" if op.get("id") else DOCS_URL
    if method == "GET" and path in stream_api_paths:
        stream_name = next((name for name, spath in stream_command_paths.items() if spath == path), None)
        if stream_name is None:
            stream_name = next((n for n, p in simple_streams if api_path(p) == path), None)
        api_rows.append({"method": method, "path": path, "covered_by": {"stream": stream_name}})
        continue
    if key in write_api_keys:
        write_name = next(name for name, _kind, m, p, _body, _risk in write_specs if m == method and api_path(p) == path)
        api_rows.append({"method": method, "path": path, "covered_by": {"write": write_name}})
        continue
    if method == "GET" and not is_binary_family(path) and not is_webhook(path):
        direct_slug = slug(title)
        if "?" in path:
            direct_slug = direct_slug + "_" + slug(path.split("?", 1)[1])
        op_id = "lever-hiring." + direct_slug
        cmd_path = "direct " + direct_slug.replace("_", "-")
        path_params = re.findall(r"\{([A-Za-z_][A-Za-z0-9_]*)\}", path)
        flags = [{"name": kebab(param), "type": "string", "summary": f"Lever {param} identifier.", "maps_to": f"path.{param}", "allow_empty": False} for param in path_params]
        policy = "clinical_json_redacted" if "eeo" in path.lower() or "diversity" in path.lower() else "json_redacted"
        direct_op = {
            "id": op_id,
            "kind": "rest_read",
            "summary": title,
            "description": "Fixed-target Lever JSON direct read with bounded response size and redacted JSON output.",
            "source_url": source_url,
            "risk": "high" if policy == "clinical_json_redacted" else "medium",
            "approval": "none: read-only fixed-target operation with bounded redacted output",
            "output_policy": policy,
            "rest": {"method": "GET", "path": path, "max_bytes": 1048576},
        }
        sensitive_policy = sensitive_direct_read_policy(path)
        if sensitive_policy:
            direct_op["sensitive_policy"] = sensitive_policy
        direct_ops.append(direct_op)
        cli_commands.append({
            "path": cmd_path,
            "summary": title,
            "intent": "direct_read",
            "availability": "implemented",
            "operation": op_id,
            "output_policy": policy,
            "flags": flags,
            "api_surface": [{"method": "GET", "path": path}],
            "examples": ["pm lever-hiring " + cmd_path + (" " + " ".join(f"--{f['name']} fixture-{f['name']}" for f in flags) if flags else "") + " --json"],
        })
        api_rows.append({"method": method, "path": path, "covered_by": {"direct_read": cmd_path}})
        continue
    # Blocked operation rows.
    if is_binary_family(path):
        model, risk, reason = "binary_read", "high", "Lever file/resume/upload/download operation requires a bounded binary or multipart transfer executor; no generic file byte passthrough is exposed."
    elif is_webhook(path):
        model, risk, reason = "local_workflow", "high", "Lever webhook subscription lifecycle is a CDC/changefeed control surface; blocked pending webhook/CDC foundation #2986/#2988 and no live subscription configuration is performed."
    elif method in {"DELETE"}:
        model, risk, reason = "destructive_action", "critical", "Destructive Lever operation is blocked until a closed schema, typed confirmation, redaction, and provider idempotency evidence are authored."
    else:
        model, risk, reason = "sensitive_reverse_etl", "high", "Lever mutation is blocked because the documented request uses required query/form parameters or a body shape the current declarative write contract cannot express without a generic passthrough."
    api_rows.append({"method": method, "path": path, "operation": {"model": model, "status": "blocked", "risk": risk, "blocked_by_default": True, "reason": reason, "source_url": source_url}})

# Add webhook event rows as CDC/changefeed operation ledger entries.
for event in WEBHOOK_EVENTS:
    api_rows.append({
        "method": "WEBHOOK",
        "path": f"lever-webhook:{event}",
        "operation": {
            "model": "local_workflow",
            "status": "blocked",
            "risk": "medium",
            "blocked_by_default": True,
            "reason": "Lever webhook event payload ingestion requires the shared webhook/CDC receiver and state foundation (#2986/#2988); no live subscription or receiver is configured by this connector bundle.",
            "source_url": DOCS_URL + "#event-payloads",
        },
    })

(DEF / "operations.json").write_text(json.dumps({"operations": direct_ops}, indent=2) + "\n")
(DEF / "api_surface.json").write_text(json.dumps({
    "api": "Lever Data API and webhook events (https://api.lever.co/v1)",
    "docs": DOCS_URL,
    "reviewed_at": "2026-08-01",
    "operation_ledger_version": 1,
    "scope": "Current official Lever Developer documentation: 107 HTTP operations plus 10 webhook trigger/event names. Implemented rows are fixed-target streams, bounded direct reads, or typed reverse-ETL writes; unsupported binary, query-parameter write, and webhook/CDC shapes are blocked with source evidence.",
    "endpoints": api_rows,
}, indent=2) + "\n")

# CLI surface.
write_by_name = {a[0]: a for a in write_specs}
for action in actions:
    raw = next(item for item in write_specs if item[0] == action["name"])
    _name, _kind, method, raw_path, body_fields, risk = raw
    path_fields = action.get("path_fields", [])
    flags = []
    for field in path_fields + body_fields:
        ftype = "boolean" if field == "secretByDefault" else "integer" if field.endswith("At") else "string_array" if field in {"fields", "fieldValues"} else "string"
        flags.append({"name": kebab(field), "type": ftype, "summary": f"Lever {field} value.", "maps_to": f"record.{field}", "allow_empty": False} if ftype == "string" else {"name": kebab(field), "type": ftype, "summary": f"Lever {field} value.", "maps_to": f"record.{field}"})
    cmd_path = kebab(action["name"]) + " plan"
    cli_commands.append({
        "path": cmd_path,
        "summary": risk,
        "intent": "reverse_etl",
        "availability": "implemented",
        "write": action["name"],
        "flags": flags,
        "api_surface": [{"method": method, "path": api_path(raw_path)}],
        "risk": action["risk"],
        "approval": "Reverse ETL writes require plan, preview, explicit approval, then execute.",
        "examples": ["pm lever-hiring " + cmd_path + " --preview --json"],
    })

cli_surface = {
    "tagline": "Read Lever Hiring records and safely plan typed Lever mutations.",
    "usage": "pm lever-hiring <command> [flags]",
    "source_cli": {"name": "Lever API", "docs": DOCS_URL, "reference": "Lever Developer documentation fetched 2026-08-01", "source": "provider_api"},
    "groups": [
        {"id": "etl", "title": "ETL streams", "commands": sorted({c["path"].split()[0] for c in cli_commands if c["intent"] == "etl"})},
        {"id": "direct", "title": "Bounded direct reads", "commands": ["direct"]},
        {"id": "writes", "title": "Reverse ETL write plans", "commands": sorted({c["path"].split()[0] for c in cli_commands if c["intent"] == "reverse_etl"})},
    ],
    "global_flags": [
        {"name": "credential", "type": "string", "summary": "Credential name to use for the Lever request."},
        {"name": "connection", "type": "string", "summary": "Alias for --credential."},
        {"name": "config", "type": "string_array", "summary": "Connector config override as key=value; never pass secret values here."},
        {"name": "json", "type": "boolean", "summary": "Emit machine-readable JSON output."},
        {"name": "limit", "type": "integer", "summary": "Maximum ETL records to emit."},
        {"name": "max-bytes", "type": "integer", "summary": "Maximum bounded direct-read response bytes."},
        {"name": "preview", "type": "boolean", "summary": "Preview a reverse-ETL write command without making a network mutation."},
        {"name": "approve", "type": "string", "summary": "Approval token required to execute a reverse-ETL plan."},
        {"name": "confirm", "type": "string", "summary": "Typed confirmation challenge for destructive reverse-ETL writes."},
    ],
    "commands": cli_commands,
    "help_topics": [
        {"name": "operation-ledger", "summary": "Lever operation ledger covers 107 HTTP operations and 10 webhook events from the official documentation."},
        {"name": "write-safety", "summary": "Lever write commands use reverse ETL plan -> preview -> approval -> execute; destructive actions require typed confirmation."},
    ],
}
(DEF / "cli_surface.json").write_text(json.dumps(cli_surface, indent=2) + "\n")

# Metadata and docs.
metadata = json.loads((DEF / "metadata.json").read_text())
metadata["description"] = "Reads Lever Hiring opportunities, postings, users, requisitions, stages, and related hiring resources; exposes bounded direct reads and selected typed reverse-ETL write plans."
metadata["capabilities"]["write"] = True
metadata["risk"]["write"] = "selected Lever mutations are exposed only through reverse ETL plan, preview, explicit approval, and execute; destructive actions require typed confirmation"
(DEF / "metadata.json").write_text(json.dumps(metadata, indent=2) + "\n")

covered_streams = sum(1 for r in api_rows if r.get("covered_by", {}).get("stream"))
covered_direct = sum(1 for r in api_rows if r.get("covered_by", {}).get("direct_read"))
covered_writes = sum(1 for r in api_rows if r.get("covered_by", {}).get("write"))
blocked = sum(1 for r in api_rows if r.get("operation"))
fixture_streams = len(streams)
fixture_writes = len(actions)
(DEF / "docs.md").write_text(f"""# Overview

Reads Lever Hiring opportunities, postings, users, requisitions, stages, and related hiring records through the Lever Data API. This bundle also exposes fixed-target bounded direct reads and selected typed reverse-ETL write plans.

Service API documentation: {DOCS_URL}.

Current official operation ledger: 117 operations total (107 HTTP operations and 10 webhook trigger/event names). Implemented rows: {covered_streams} stream-backed reads, {covered_direct} bounded direct reads, and {covered_writes} typed writes. Blocked/planned rows: {blocked}. Certified rows: 0 (fixture-only; no live provider calls were made).

## Auth setup

Connection fields:

- `access_token` (optional, secret, string): Lever OAuth2 access token. Bearer authentication is preferred when both credential types are present.
- `api_key` (optional, secret, string): Lever API key sent as HTTP Basic username with a blank password.
- `base_url` (optional, string): default `https://api.lever.co/v1`; use `https://api.sandbox.lever.co/v1` or a test proxy only via config, never by editing operation paths.
- `mode` (optional, string): retained for credential-free fixture/conformance compatibility.

Secret fields are redacted: `access_token`, `api_key`. Connection checks call `GET /postings?limit=1`.

## Streams notes

The connector declares {len(streams)} stream-backed read surfaces. Top-level collection streams use Lever cursor pagination (`limit`, `offset`, `next`, `hasNext`). Opportunity/posting sub-resource streams use the engine fan-out contract to list parent IDs from `/opportunities` or `/postings`, then request the fixed child collection path once per parent ID and stamp the parent ID onto emitted records.

Implemented stream names: {', '.join('`'+s['name']+'`' for s in streams)}.

Scalar-list and binary/file-family operations that cannot be represented as durable object records are kept out of ETL streams and are either bounded direct reads or blocked in the operation ledger.

## Write actions & risks

The connector declares {len(actions)} typed write actions: {', '.join('`'+a['name']+'`' for a in actions)}.

Writes are only available through reverse ETL plan -> preview -> explicit approval -> execute. Destructive/no-body actions use `confirm: destructive`. The bundle does not expose arbitrary request bodies, raw query strings, generic method/path/body, file bytes, shell commands, or passthrough HTTP tools.

Many official Lever mutations remain blocked because the current generic write executor does not support provider query parameters such as `perform_as`, or because the official request body is not closed enough to expose without a connector-specific schema/hook.

## Known limits

- Fixture-only evidence: no live Lever credentials, provider calls, provider writes, or certification run were used.
- Webhook trigger ingestion and webhook subscription lifecycle rows are blocked pending the shared webhook/CDC foundation (#2986/#2988).
- Lever file/resume/upload/download rows are blocked pending a bounded binary/multipart transfer executor; no generic file byte passthrough is exposed.
- Mutations requiring documented query/form parameters (for example `perform_as`) are blocked until shared write-query support or a connector hook can express them without a raw query escape hatch.
- The generated operation ledger is based on the current official Lever Developer documentation fetched for this slice and should be re-audited when Lever changes that page.
""")

print(json.dumps({
    "streams": len(streams),
    "writes": len(actions),
    "direct_reads": len(direct_ops),
    "api_rows": len(api_rows),
    "covered_stream_rows": covered_streams,
    "covered_direct_rows": covered_direct,
    "covered_write_rows": covered_writes,
    "blocked_rows": blocked,
    "fixture_streams": fixture_streams,
    "fixture_writes": fixture_writes,
}, indent=2))
