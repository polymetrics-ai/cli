# THREAT-MODEL — DuckDB query engine

## Assets
- Warehouse data (extracted records in JSONL). Local filesystem paths.

## Risks & mitigations
- **SQL mutation via query path**: the query API must be read-only. Mitigation: `validateSelectOnly`
  rejects non-SELECT/WITH statements, `;` (statement chaining), and DDL/DML/`attach`/`copy`/`pragma`/
  `call`/`export` tokens. DuckDB views are created by the engine, not the caller.
- **Path / identifier injection**: table view names derive from filenames in the warehouse dir;
  validate identifiers against `[A-Za-z0-9_]+` and pass file paths as quote-escaped string literals to
  `read_json_auto`. Do not interpolate user input into `CREATE VIEW` names.
- **Local file disclosure via read_* functions**: only the engine constructs `read_json_auto` over the
  warehouse dir; user SQL references views, not arbitrary `read_csv`/`read_parquet('/etc/...')`. The
  SELECT-only validator plus the view-only schema bound the blast radius. (Future: disable DuckDB
  filesystem/`enable_external_access` if tighter sandboxing is needed.)
- **Resource exhaustion**: queries are LIMIT-wrapped when no limit is supplied; per-call connection.

## Non-applicable
No secrets, no network, no auth surface in this phase.
