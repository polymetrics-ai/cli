#!/usr/bin/env python3
"""Generate the PostgreSQL api_surface.json operation ledger.

The ledger is source-first and reconciles to the landed official audit record:
SQL commands (183), frontend/backend protocol messages (52), streaming
replication commands (8), logical replication messages (20), total 263.
"""

from __future__ import annotations

import datetime as dt
import html
import json
import re
import sys
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
OUT = ROOT / "internal/connectors/defs/postgres/api_surface.json"
SUMMARY = ROOT / ".planning/phases/postgres-parity-wave04/api-surface-summary.json"
BASE = "https://www.postgresql.org/docs/current/"

URLS = {
    "sql": BASE + "sql-commands.html",
    "protocol": BASE + "protocol-message-formats.html",
    "streaming": BASE + "protocol-replication.html",
    "logical": BASE + "protocol-logicalrep-message-formats.html",
}

SQL_ETL = {"DECLARE"}
SQL_DIRECT = {"EXPLAIN", "FETCH", "SELECT", "SELECT INTO", "SHOW", "VALUES"}
SQL_BINARY = {"COPY"}
SQL_CDC = {"CREATE SUBSCRIPTION", "ALTER SUBSCRIPTION", "DROP SUBSCRIPTION"}
SQL_EXCLUDED = {
    "ABORT",
    "BEGIN",
    "CHECKPOINT",
    "COMMIT",
    "COMMIT PREPARED",
    "DEALLOCATE",
    "DISCARD",
    "END",
    "EXECUTE",
    "LOCK",
    "MOVE",
    "PREPARE",
    "PREPARE TRANSACTION",
    "RELEASE SAVEPOINT",
    "RESET",
    "ROLLBACK",
    "ROLLBACK PREPARED",
    "ROLLBACK TO SAVEPOINT",
    "SAVEPOINT",
    "SET",
    "SET CONSTRAINTS",
    "SET TRANSACTION",
    "START TRANSACTION",
}

PROTOCOL_DIRECT = {
    "Bind",
    "BindComplete",
    "CommandComplete",
    "DataRow",
    "Describe",
    "EmptyQueryResponse",
    "Execute",
    "FunctionCall",
    "NoData",
    "ParameterDescription",
    "Parse",
    "ParseComplete",
    "PortalSuspended",
    "Query",
    "ReadyForQuery",
    "RowDescription",
}
PROTOCOL_BINARY = {
    "CopyData",
    "CopyDone",
    "CopyFail",
    "CopyInResponse",
    "CopyOutResponse",
    "CopyBothResponse",
    "FunctionCallResponse",
}
PROTOCOL_CDC = {"NotificationResponse"}

STREAMING_COMMANDS = [
    "IDENTIFY_SYSTEM",
    "TIMELINE_HISTORY",
    "CREATE_REPLICATION_SLOT",
    "ALTER_REPLICATION_SLOT",
    "READ_REPLICATION_SLOT",
    "START_REPLICATION",
    "DROP_REPLICATION_SLOT",
    "BASE_BACKUP",
]

LOGICAL_EXCLUDED = {"TupleData", "Type"}

COVERED_WRITES = {
    "INSERT": "insert_row",
    "UPDATE": "update_row",
    "MERGE": "upsert_row",
    "DELETE": "delete_row",
    "TRUNCATE": "truncate_table",
}


def fetch(url: str) -> str:
    with urllib.request.urlopen(url, timeout=30) as resp:
        return resp.read().decode("utf-8", errors="replace")


def strip_tags(s: str) -> str:
    return re.sub(r"\s+", " ", html.unescape(re.sub(r"<.*?>", " ", s))).strip()


def parse_sql() -> list[tuple[str, str]]:
    raw = fetch(URLS["sql"])
    seen: dict[str, str] = {}
    for href, name in re.findall(r'<a href="(sql-[^"]+\.html)">([^<]+)</a>', raw):
        name = html.unescape(name).strip()
        if name and name not in seen:
            seen[name] = BASE + href
    out = sorted(seen.items(), key=lambda kv: kv[0])
    if len(out) != 183:
        raise SystemExit(f"SQL parse count {len(out)} != 183")
    return out


def parse_protocol() -> list[tuple[str, str]]:
    raw = fetch(URLS["protocol"])
    out: list[tuple[str, str]] = []
    for anchor, term in re.findall(r'<dt id="(PROTOCOL-MESSAGE-FORMATS-[^"]+)"[^>]*>.*?<span class="term">(.*?)</span>', raw, re.S):
        name = re.sub(r" \([A-Z &]+\)$", "", strip_tags(term))
        out.append((name, URLS["protocol"] + "#" + anchor))
    if len(out) != 52:
        raise SystemExit(f"protocol parse count {len(out)} != 52")
    return out


def parse_logical() -> list[tuple[str, str]]:
    raw = fetch(URLS["logical"])
    out: list[tuple[str, str]] = []
    for anchor, term in re.findall(r'<dt id="(PROTOCOL-LOGICALREP-MESSAGE-FORMATS-[^"]+)"[^>]*>.*?<span class="term">(.*?)</span>', raw, re.S):
        name = strip_tags(term)
        out.append((name, URLS["logical"] + "#" + anchor))
    if len(out) != 20:
        raise SystemExit(f"logical parse count {len(out)} != 20")
    return out


def endpoint(method: str, path: str, *, source: str | None = None, covered_by: dict | None = None, operation: dict | None = None) -> dict:
    row = {"method": method, "path": path}
    if covered_by:
        row["covered_by"] = covered_by
    if operation:
        if source:
            operation["source_url"] = source
        row["operation"] = operation
    return row


def op(model: str, risk: str, reason: str, notes: str = "") -> dict:
    row = {"model": model, "status": "blocked", "risk": risk, "blocked_by_default": True, "reason": reason}
    if notes:
        row["notes"] = notes
    return row


def sql_row(name: str, url: str) -> tuple[dict, str]:
    if name in COVERED_WRITES:
        return endpoint("SQL", name, source=url, covered_by={"write": COVERED_WRITES[name]}), "reverse_etl_write"
    if name in SQL_ETL:
        return endpoint("SQL", name, source=url, operation=op("direct_read", "medium", "DECLARE is a cursor primitive; pm exposes bounded table Read instead of raw cursor SQL.")), "etl_read"
    if name in SQL_DIRECT:
        return endpoint("SQL", name, source=url, operation=op("direct_read", "medium", "Raw SQL read/query text is not exposed; use catalog/read with bounded streams.")), "direct_read_query_search"
    if name in SQL_BINARY:
        return endpoint("SQL", name, source=url, operation=op("binary_read", "high", "COPY can stream arbitrary binary/text payloads and requires a separate bounded binary contract.")), "binary_file"
    if name in SQL_CDC:
        return endpoint("SQL", name, source=url, operation=op("local_workflow", "high", "Logical replication subscriptions require live CDC orchestration and are blocked in the fixture-only connector.")), "cdc_changefeed"
    if name in SQL_EXCLUDED:
        return endpoint("SQL", name, source=url, operation=op("disallowed", "low", "Transaction/session/control SQL is not a connector data operation and is counted as not applicable.")), "excluded_not_applicable"
    model = "destructive_action" if name.startswith(("DROP", "TRUNCATE")) else "admin_reverse_etl"
    risk = "critical" if model == "destructive_action" else "high"
    return endpoint("SQL", name, source=url, operation=op(model, risk, "Official SQL command is not exposed because raw/admin SQL write variants need command-specific closed schemas; generic SQL writes are forbidden.")), "reverse_etl_write"


def protocol_row(name: str, url: str) -> tuple[dict, str]:
    if name in PROTOCOL_DIRECT:
        return endpoint("PROTOCOL_MESSAGE", name, source=url, operation=op("direct_read", "medium", "Frontend/backend protocol message is handled internally by pgx or blocked; pm does not expose raw protocol messages.")), "direct_read_query_search"
    if name in PROTOCOL_BINARY:
        return endpoint("PROTOCOL_MESSAGE", name, source=url, operation=op("binary_read", "high", "COPY/function binary payload messages are not exposed without a bounded binary contract.")), "binary_file"
    if name in PROTOCOL_CDC:
        return endpoint("PROTOCOL_MESSAGE", name, source=url, operation=op("local_workflow", "medium", "LISTEN/NOTIFY-style change notifications are not exposed as a CDC surface in this fixture-only connector.")), "cdc_changefeed"
    return endpoint("PROTOCOL_MESSAGE", name, source=url, operation=op("disallowed", "low", "Authentication, startup, error, sync, and lifecycle protocol messages are internal connector mechanics, not user-facing operations.")), "excluded_not_applicable"


def streaming_row(name: str) -> tuple[dict, str]:
    url = URLS["streaming"]
    if name == "BASE_BACKUP":
        return endpoint("REPLICATION_COMMAND", name, source=url, operation=op("binary_read", "critical", "Base backups stream database files and are outside the bounded connector read contract.")), "binary_file"
    return endpoint("REPLICATION_COMMAND", name, source=url, operation=op("local_workflow", "high", "Streaming replication commands require live replication privileges and pglogrepl orchestration; blocked pending a separate CDC gate.")), "cdc_changefeed"


def logical_row(name: str, url: str) -> tuple[dict, str]:
    if name in LOGICAL_EXCLUDED:
        return endpoint("LOGICAL_REPLICATION_MESSAGE", name, source=url, operation=op("disallowed", "low", "Logical type/tuple metadata is a message component rather than a standalone connector operation.")), "excluded_not_applicable"
    return endpoint("LOGICAL_REPLICATION_MESSAGE", name, source=url, operation=op("local_workflow", "high", "Logical replication messages are decoder/CDC protocol events only; live CDC is blocked pending the gated pglogrepl dependency.")), "cdc_changefeed"


def main() -> int:
    rows: list[dict] = []
    lanes: dict[str, int] = {}
    surfaces: dict[str, int] = {}

    def add(row: dict, lane: str, surface: str) -> None:
        rows.append(row)
        lanes[lane] = lanes.get(lane, 0) + 1
        surfaces[surface] = surfaces.get(surface, 0) + 1

    for name, url in parse_sql():
        row, lane = sql_row(name, url)
        add(row, lane, "sql_commands")
    for name, url in parse_protocol():
        row, lane = protocol_row(name, url)
        add(row, lane, "frontend_backend_protocol_messages")
    for name in STREAMING_COMMANDS:
        row, lane = streaming_row(name)
        add(row, lane, "streaming_replication_commands")
    for name, url in parse_logical():
        row, lane = logical_row(name, url)
        add(row, lane, "logical_replication_messages")

    expected_lanes = {
        "etl_read": 1,
        "reverse_etl_write": 149,
        "direct_read_query_search": 22,
        "binary_file": 9,
        "cdc_changefeed": 29,
        "excluded_not_applicable": 53,
    }
    expected_surfaces = {
        "sql_commands": 183,
        "frontend_backend_protocol_messages": 52,
        "streaming_replication_commands": 8,
        "logical_replication_messages": 20,
    }
    if len(rows) != 263 or lanes != expected_lanes or surfaces != expected_surfaces:
        raise SystemExit(f"count mismatch rows={len(rows)} lanes={lanes} surfaces={surfaces}")

    today = dt.date.today().isoformat()
    payload = {
        "api": "PostgreSQL 18 SQL commands and wire/replication protocol surfaces",
        "docs": "https://www.postgresql.org/docs/current/",
        "reviewed_at": today,
        "operation_ledger_version": 1,
        "scope": "PostgreSQL 18 official docs bounded to the SQL Commands index, frontend/backend protocol message formats, streaming replication protocol commands, and logical replication message formats. Generic SQL functions, extension APIs, system catalogs, and local decoder/interface stubs are not counted. This connector exposes dynamic catalog/read plus five bounded row/table reverse-ETL actions only; all other official operations are blocked or evidence-backed not applicable.",
        "endpoints": rows,
    }
    OUT.write_text(json.dumps(payload, indent=2, sort_keys=False) + "\n")
    SUMMARY.write_text(json.dumps({"total": len(rows), "lanes": lanes, "surfaces": surfaces, "covered_writes": COVERED_WRITES}, indent=2, sort_keys=True) + "\n")
    print(f"wrote {OUT} with {len(rows)} rows")
    print(json.dumps({"lanes": lanes, "surfaces": surfaces}, sort_keys=True))
    return 0


if __name__ == "__main__":
    sys.exit(main())
