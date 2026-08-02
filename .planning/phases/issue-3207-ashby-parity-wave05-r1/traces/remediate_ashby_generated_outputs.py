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
DOCUMENTED_MAPS = {
    "create_department": {("properties", "extraData"): True},
    "update_department": {("properties", "extraData"): True},
    "create_location": {("properties", "extraData"): True},
    "create_interview_schedule": {("properties", "interviewEvents", "items", "properties", "extraData"): True},
    "create_survey_submission": {("properties", "submittedValues"): True},
}


def main() -> None:
    spec = read_json(DEFS / "spec.json")
    spec["required"] = [name for name in spec.get("required", []) if name != "start_date"]
    spec.get("properties", {}).pop("start_date", None)
    write_json(DEFS / "spec.json", spec)

    streams = read_json(DEFS / "streams.json")
    incremental_streams = {
        stream["name"] for stream in streams.get("streams", []) if stream.get("incremental")
    }
    for stream in streams.get("streams", []):
        stream.pop("incremental", None)
        schema_path = DEFS / stream["schema"]
        schema = read_json(schema_path)
        schema.pop("x-cursor-field", None)
        write_json(schema_path, schema)
    write_json(DEFS / "streams.json", streams)

    writes = read_json(DEFS / "writes.json")
    for action in writes.get("actions", []):
        close_modeled_objects(action.get("record_schema"))
        for path, additional_properties in DOCUMENTED_MAPS.get(action.get("name"), {}).items():
            node = at_path(action.get("record_schema"), path)
            if isinstance(node, dict):
                node["additionalProperties"] = additional_properties
    write_json(DEFS / "writes.json", writes)

    surface = read_json(DEFS / "cli_surface.json")
    for command in surface.get("commands", []):
        if command.get("intent") != "etl":
            continue
        flags = command.get("flags", [])
        removed = [flag for flag in flags if flag.get("type") == "string_array"]
        command["flags"] = [flag for flag in flags if flag.get("type") != "string_array"]
        notes = command.get("notes", "")
        if removed and STREAM_ARRAY_FOUNDATION not in notes:
            names = ", ".join("--" + flag["name"] for flag in removed)
            notes += f" Repeatable array request variants ({names}) are blocked pending {STREAM_ARRAY_FOUNDATION}."
        if command.get("stream") in incremental_streams and SYNC_TOKEN_FOUNDATION not in notes:
            notes += f" Incremental execution is blocked pending {SYNC_TOKEN_FOUNDATION}; this stream is full-refresh only."
        command["notes"] = notes
    write_json(DEFS / "cli_surface.json", surface)

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
    docs = docs.replace(
        "`hiringTeamRole.list` defaults to `namesOnly=true`;",
        f"Opaque incremental state is blocked pending `{SYNC_TOKEN_FOUNDATION}`, and repeatable array stream-command variants are blocked pending `{STREAM_ARRAY_FOUNDATION}`. `hiringTeamRole.list` defaults to `namesOnly=true`;",
    )
    docs_path.write_text(docs)

    streams_go_path = ROOT / "internal/connectors/native/ashby/streams_gen.go"
    streams_go = streams_go_path.read_text()
    streams_go = re.sub(r'cursorField: "[^"]*"', 'cursorField: ""', streams_go)
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


def at_path(node: Any, path: tuple[str, ...]) -> Any:
    for part in path:
        if not isinstance(node, dict):
            return None
        node = node.get(part)
    return node


def read_json(path: Path) -> Any:
    return json.loads(path.read_text())


def write_json(path: Path, value: Any) -> None:
    path.write_text(json.dumps(value, indent=2) + "\n")


if __name__ == "__main__":
    main()
