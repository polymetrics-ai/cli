#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[4]
DEFS = ROOT / "internal/connectors/defs/ashby"
SYNC_TOKEN_FOUNDATION = "ashby-sync-token-checkpoint-foundation"
STREAM_ARRAY_FOUNDATION = "connector-stream-repeatable-array-foundation"
REFERRAL_FORM_FOUNDATION = "ashby-referral-form-info-side-effect-foundation"
APPLICATION_FORM_FOUNDATION = "ashby-application-form-typed-multipart-foundation"
BLOCKED_PATHS = {
    "/referralForm.info": {
        "model": "admin_reverse_etl",
        "risk": "medium",
        "reason": f"fetches the default referral form but conditionally creates one when absent; blocked pending {REFERRAL_FORM_FOUNDATION} so an apparent read cannot mutate Ashby without a typed confirmation path",
    },
    "/applicationForm.submit": {
        "model": "sensitive_reverse_etl",
        "risk": "high",
        "reason": f"requires multipart/form-data with typed file parts for application fields; blocked pending {APPLICATION_FORM_FOUNDATION} because the Ashby write executor is JSON-only",
    },
}
DIRECT_READS = {
    "/user.interviewerSettings": {
        "id": "ashby.direct.user.interviewer.settings",
        "summary": "user.interviewerSettings",
        "description": "Get interviewer settings for a user.",
        "source_url": "https://developers.ashbyhq.com/reference/userinterviewersettings",
        "risk": "low",
        "command": "user interviewer-settings",
        "command_risk": "bounded JSON direct read; credential-marked response fields are redacted, and non-credential interviewer settings remain complete in trusted live local output",
    },
    "/report.generate": {
        "id": "ashby.direct.report.generate",
        "summary": "report.generate",
        "description": "Generates a new report or polls the status of an existing report generation.",
        "source_url": "https://developers.ashbyhq.com/reference/reportgenerate",
        "risk": "medium",
        "command": "report generate",
        "command_summary": "Start an Ashby report generation or check an existing request.",
        "command_risk": "bounded JSON direct read that starts or polls a documented Ashby report generation and returns at most 1 MiB of redacted JSON; the connector does not fetch returned report URLs or poll automatically",
    },
}
BLOCKED_COMMANDS = {
    "referral-form info",
    "application-form submit",
}
DOCUMENTED_MAPS = {
    "create_department": {("properties", "extraData"): True},
    "update_department": {("properties", "extraData"): True},
    "create_location": {("properties", "extraData"): True},
    "create_interview_schedule": {("properties", "interviewEvents", "items", "properties", "extraData"): True},
    "create_survey_submission": {("properties", "submittedValues"): True},
}


def main() -> None:
    streams_go_path = ROOT / "internal/connectors/native/ashby/streams_gen.go"
    streams_go = streams_go_path.read_text()
    sync_token_streams = set(
        re.findall(r'^\s*"([^"]+)":.*requestFields: map\[string\]string\{[^\n]*"syncToken"', streams_go, re.MULTILINE)
    )

    spec = read_json(DEFS / "spec.json")
    spec["required"] = [name for name in spec.get("required", []) if name != "start_date"]
    spec.get("properties", {}).pop("start_date", None)
    write_json(DEFS / "spec.json", spec)

    streams = read_json(DEFS / "streams.json")
    incremental_streams = {stream["name"] for stream in streams.get("streams", []) if stream.get("incremental")}
    streams["streams"] = [stream for stream in streams.get("streams", []) if stream.get("name") != "referral_form_info"]
    for stream in streams["streams"]:
        stream.pop("incremental", None)
        schema_path = DEFS / stream["schema"]
        schema = read_json(schema_path)
        schema.pop("x-cursor-field", None)
        write_json(schema_path, schema)
    write_json(DEFS / "streams.json", streams)
    remove_path(DEFS / "schemas/referral_form_info.json")
    remove_path(DEFS / "fixtures/streams/referral_form_info/page_1.json")

    writes = read_json(DEFS / "writes.json")
    actions_by_path = {action.get("path"): action for action in writes.get("actions", [])}
    removed_write_paths = set(DIRECT_READS) | {"/applicationForm.submit"}
    writes["actions"] = [action for action in writes.get("actions", []) if action.get("path") not in removed_write_paths]
    for action in writes["actions"]:
        close_modeled_objects(action.get("record_schema"))
        for path, additional_properties in DOCUMENTED_MAPS.get(action.get("name"), {}).items():
            node = at_path(action.get("record_schema"), path)
            if isinstance(node, dict):
                node["additionalProperties"] = additional_properties
    write_json(DEFS / "writes.json", writes)

    operations = read_json(DEFS / "operations.json")
    operations_by_id = {operation.get("id"): operation for operation in operations.get("operations", [])}
    for path, rule in DIRECT_READS.items():
        if rule["id"] in operations_by_id:
            continue
        action = actions_by_path.get(path)
        if action is None:
            raise RuntimeError(f"missing source write action for {path}")
        operations["operations"].append({
            "id": rule["id"],
            "kind": "rest_read",
            "summary": rule["summary"],
            "description": rule["description"],
            "source_url": rule["source_url"],
            "risk": rule["risk"],
            "approval": "none",
            "output_policy": "json_redacted",
            "rest": {
                "method": action["method"],
                "path": path,
                "content_type": "application/json",
                "max_bytes": 1048576,
                "body": {},
                "body_schema": action["record_schema"],
            },
        })
    write_json(DEFS / "operations.json", operations)
    direct_required_by_path = {}
    for operation in operations.get("operations", []):
        rest = operation.get("rest") or {}
        body_schema = rest.get("body_schema") or {}
        direct_required_by_path[rest.get("path")] = set(body_schema.get("required", []))

    surface = read_json(DEFS / "cli_surface.json")
    commands = []
    direct_by_command = {rule["command"]: (path, rule) for path, rule in DIRECT_READS.items()}
    for command in surface.get("commands", []):
        if command.get("path") in BLOCKED_COMMANDS:
            continue
        direct = direct_by_command.get(command.get("path"))
        if direct is not None:
            path, rule = direct
            action = actions_by_path.get(path)
            if action is None:
                required = direct_required_by_path.get(path, set())
            else:
                required = set(action.get("record_schema", {}).get("required", []))
            command["intent"] = "direct_read"
            command["availability"] = "implemented"
            command.pop("write", None)
            command.pop("redact_fields", None)
            command["operation"] = rule["id"]
            if "command_summary" in rule:
                command["summary"] = rule["command_summary"]
            command["output_policy"] = "json_redacted"
            command["risk"] = rule["command_risk"]
            command["approval"] = "none"
            command["notes"] = "Fixed Ashby POST direct read; no raw method/path/body override is exposed."
            for flag in command.get("flags", []):
                maps_to = flag.get("maps_to", "")
                if maps_to.startswith("record."):
                    maps_to = "body." + maps_to.removeprefix("record.")
                    flag["maps_to"] = maps_to
                root = maps_to.removeprefix("body.").split(".", 1)[0]
                if root in required:
                    flag["required"] = True
                else:
                    flag.pop("required", None)
        elif command.get("intent") == "etl":
            flags = command.get("flags", [])
            removed = [flag for flag in flags if flag.get("type") == "string_array"]
            command["flags"] = [flag for flag in flags if flag.get("type") != "string_array"]
            notes = command.get("notes", "")
            if removed and STREAM_ARRAY_FOUNDATION not in notes:
                names = ", ".join("--" + flag["name"] for flag in removed)
                notes += f" Repeatable array request variants ({names}) are blocked pending {STREAM_ARRAY_FOUNDATION}."
            if command.get("stream") in sync_token_streams:
                for sentence in (
                    f" Opaque syncToken incremental checkpointing is blocked pending {SYNC_TOKEN_FOUNDATION}; this stream is full-refresh only.",
                    f" Incremental execution is blocked pending {SYNC_TOKEN_FOUNDATION}; this stream is full-refresh only.",
                    f" Opaque syncToken checkpointing is blocked pending {SYNC_TOKEN_FOUNDATION}; this stream is full-refresh only.",
                ):
                    notes = notes.replace(sentence, "")
                notes += f" Opaque syncToken checkpointing is blocked pending {SYNC_TOKEN_FOUNDATION}; this stream is full-refresh only."
                command["summary"] = f"Full-refresh-only Ashby {command['path']} read. Opaque syncToken checkpointing is unavailable pending {SYNC_TOKEN_FOUNDATION}."
            elif command.get("stream") in incremental_streams and SYNC_TOKEN_FOUNDATION not in notes:
                notes += f" Incremental execution is blocked pending {SYNC_TOKEN_FOUNDATION}; this stream is full-refresh only."
            command["notes"] = notes
        command["summary"] = terminal_summary(command.get("summary", ""), command.get("path", "Ashby command"))
        commands.append(command)
    surface["commands"] = commands
    commands_by_intent = {
        "streams": [command["path"] for command in commands if command.get("intent") == "etl"],
        "direct_reads": [command["path"] for command in commands if command.get("intent") == "direct_read"],
        "reverse_etl": [command["path"] for command in commands if command.get("intent") == "reverse_etl"],
    }
    for group in surface.get("groups", []):
        if group.get("id") in commands_by_intent:
            group["commands"] = commands_by_intent[group["id"]]
    write_json(DEFS / "cli_surface.json", surface)

    api_surface = read_json(DEFS / "api_surface.json")
    for endpoint in api_surface.get("endpoints", []):
        path = endpoint.get("path")
        if path in BLOCKED_PATHS:
            rule = BLOCKED_PATHS[path]
            endpoint.pop("covered_by", None)
            endpoint.pop("excluded", None)
            endpoint["operation"] = {
                "model": rule["model"],
                "status": "blocked",
                "risk": rule["risk"],
                "blocked_by_default": True,
                "reason": rule["reason"],
                "source_url": "https://developers.ashbyhq.com/reference/referralforminfo" if path == "/referralForm.info" else "https://developers.ashbyhq.com/reference/applicationformsubmit-1",
                "notes": "referralForm.info" if path == "/referralForm.info" else "applicationForm.submit",
            }
        elif path in DIRECT_READS:
            endpoint.pop("operation", None)
            endpoint.pop("excluded", None)
            endpoint["covered_by"] = {"direct_read": DIRECT_READS[path]["command"]}
    api_surface["scope"] = "Complete public Ashby inventory: REST operations plus OpenAPI webhook events. Supported REST read/write/direct surfaces are fixed and typed; conditional side-effect reads, multipart submissions, inbound partner/webhook, and presigned external file-transfer workflows remain blocked with source-backed reasons."
    write_json(DEFS / "api_surface.json", api_surface)

    docs_path = DEFS / "docs.md"
    docs = docs_path.read_text()
    docs = docs.replace(
        "Ashby list and info reads are fixed POST endpoints with documented body fields only. The native connector owns Ashby's cursor-in-body pagination, applies page-size and max-pages bounds, and supports client-side incremental filtering when generated stream metadata explicitly declares an incremental cursor.",
        f"Ashby list and info reads are fixed POST endpoints with documented body fields only. The native connector owns Ashby's cursor-in-body pagination and applies page-size, max-pages, and repeated-cursor bounds. Streams are full-refresh only until `{SYNC_TOKEN_FOUNDATION}` supplies an Ashby-owned persisted opaque-token state seam; timestamp fields are not used as lossy substitutes. Repeatable array stream flags are withheld until `{STREAM_ARRAY_FOUNDATION}` preserves every supplied value.",
    )
    docs = docs.replace(
        "Reverse ETL writes are typed action names with closed top-level JSON schemas and the normal plan → preview → explicit approval → execute gate.",
        "Reverse ETL writes are typed action names with recursively closed modeled JSON schemas and the normal plan → preview → explicit approval → execute gate. Explicitly documented map-valued fields retain their map schemas; all other modeled objects reject undeclared fields.",
    )
    docs = re.sub(r"- Implemented ETL(?:/changefeed)? streams: \d+", f"- Implemented ETL streams: {len(streams['streams'])}", docs)
    docs = re.sub(r"- Implemented bounded direct reads/search/file metadata operations: \d+", f"- Implemented bounded direct reads/search/file metadata operations: {len(operations['operations'])}", docs)
    docs = re.sub(r"- Implemented reverse-ETL write actions: \d+", f"- Implemented reverse-ETL write actions: {len(writes['actions'])}", docs)
    docs = re.sub(
        r"- Reverse-ETL CLI commands with scalar flags: \d+; partial nested-object flag surfaces: \d+",
        f"- Reverse-ETL CLI commands with scalar flags: {sum(1 for command in commands if command.get('intent') == 'reverse_etl' and command.get('availability') == 'implemented')}; partial nested-object flag surfaces: {sum(1 for command in commands if command.get('intent') == 'reverse_etl' and command.get('availability') == 'partial')}",
        docs,
    )
    blocked_count = sum(1 for endpoint in api_surface.get("endpoints", []) if endpoint.get("operation") is not None)
    docs = re.sub(r"- Blocked/non-executable ledger rows: \d+", f"- Blocked/non-executable ledger rows: {blocked_count}", docs)
    docs = docs.replace(
        "timestamp fields are not used as lossy substitutes. Repeatable array stream flags",
        "timestamp fields are not used as lossy substitutes. Runtime help replaces provider incremental descriptions with full-refresh-only blocker text for every documented sync-token request. Repeatable array stream flags",
    )
    known_limits = f"Blocked rows are still documented in `api_surface.json`: inbound assessment-partner APIs and webhook events are not pull-executable by a CLI connector, and `file.createFileUploadHandle` remains blocked until a reviewed bounded binary/file workflow can safely return and consume presigned upload handles. `referralForm.info` is blocked pending `{REFERRAL_FORM_FOUNDATION}` because it conditionally creates a default form. `applicationForm.submit` is blocked pending `{APPLICATION_FORM_FOUNDATION}` because the documented request requires multipart form data and typed file parts. Opaque incremental state is blocked pending `{SYNC_TOKEN_FOUNDATION}`, and repeatable array stream-command variants are blocked pending `{STREAM_ARRAY_FOUNDATION}`. `hiringTeamRole.list` defaults to `namesOnly=true`; the `namesOnly=false` object-result variant is blocked pending variant-schema foundation `ashby_hiring_team_role_list_names_only_false`. Fixture replay covers every implemented stream with synthetic values only; no live Ashby credentials or provider calls were used."
    docs = re.sub(r"## Known limits\n\n[\s\S]*$", f"## Known limits\n\n{known_limits}\n", docs)
    docs_path.write_text(docs)

    streams_go = re.sub(r'cursorField: "[^"]*"', 'cursorField: ""', streams_go)
    streams_go = re.sub(r'^\s*"referral_form_info":.*\n', "", streams_go, flags=re.MULTILINE)
    streams_go_path.write_text(streams_go)


def close_modeled_objects(node: Any) -> None:
    if isinstance(node, list):
        for item in node:
            close_modeled_objects(item)
        return
    if not isinstance(node, dict):
        return
    if isinstance(node.get("properties"), dict) or node.get("type") == "object":
        node["additionalProperties"] = False
    for value in node.values():
        close_modeled_objects(value)


def terminal_summary(description: str, fallback: str) -> str:
    source_was_truncated = len(description) >= 160
    text = description.strip() or fallback
    text = re.sub(r"\[([^\]]+)\]\([^)]*\)", r"\1", text)
    text = re.sub(r"`([^`]*)`", r"\1", text)
    text = re.sub(r"\((?:ref:|authentication#)[^)]*$", "", text)
    text = re.sub(r"(?m)^\s*>\s?", "", text)
    text = text.replace("**", "").replace("__", "")
    text = text.translate(str.maketrans("", "", "[]`*"))
    text = re.sub(r"\s+", " ", text).strip()
    needs_abbreviation = source_was_truncated or (text and text[-1] not in ".!?")
    if len(text) <= 160 and not needs_abbreviation:
        return text
    candidate = text[:161]
    sentence_end = max(candidate.rfind(". "), candidate.rfind("! "), candidate.rfind("? "))
    if sentence_end >= 40:
        return candidate[:sentence_end + 1]
    word_end = candidate[:157].rfind(" ")
    if word_end < 40:
        word_end = 157
    return candidate[:word_end].rstrip(" ,;:-") + "..."


def at_path(node: Any, path: tuple[str, ...]) -> Any:
    for part in path:
        if not isinstance(node, dict):
            return None
        node = node.get(part)
    return node


def remove_path(path: Path) -> None:
    path.unlink(missing_ok=True)
    parent = path.parent
    if parent.exists() and not any(parent.iterdir()):
        parent.rmdir()


def read_json(path: Path) -> Any:
    return json.loads(path.read_text())


def write_json(path: Path, value: Any) -> None:
    path.write_text(json.dumps(value, indent=2) + "\n")


if __name__ == "__main__":
    main()
