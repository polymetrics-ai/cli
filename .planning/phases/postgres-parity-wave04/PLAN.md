# PLAN — PostgreSQL parity wave 04 r1

## GSD command path and fallback

- Adapter preflight: `scripts/gsd doctor` passed in this worktree.
- Official prompt captured: `scripts/gsd prompt execute-phase postgres-parity-wave04-r1 --dry-run` -> `gsd-execute-phase-prompt.txt`.
- `scripts/gsd prompt programming-loop ...` is unavailable (`unknown GSD command`), so this phase uses manual GSD universal runtime fallback with TDD, plan, verification, and run-state artifacts.

## Required skills loaded

`gsd-core`, `context-mode`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-stretchr-testify`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-database`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-lint`.

## Scope guard

Allowed production scope: `internal/connectors/defs/postgres/**`, `internal/connectors/native/postgres/**`, postgres-owned tests/fixtures, generated connector docs/skills/catalog/website surfaces affected by postgres metadata, and `.planning/phases/postgres-parity-wave04/**`.

Non-goals: no live DB calls, no credentials/secrets, no new dependencies, no raw generic SQL read/write/query commands, no shell/file/extension/binary passthrough, no certification claim, no push/PR/merge/no-mistakes.

## Source inventory plan

1. Re-audit PostgreSQL 18 official docs: SQL Commands index, frontend/backend protocol message formats, streaming replication commands, logical replication message formats.
2. Reconcile against `/Users/karthiksivadas/karthik-agent-workspace/data/cli-official-api-parity-audit-r2/{audit.json,sources.json}`.
3. Generate `api_surface.json` with 263 operation rows matching landed audit counts: SQL=183, protocol=52, streaming=8, logical=20; lanes etl_read=1, reverse_etl_write=149, direct_read_query_search=22, binary_file=9, cdc_changefeed=29, excluded_not_applicable=53.

## Implementation slices

### Slice A — tests first

Add focused postgres tests for the missing expected behavior:

- metadata/definition/manifest exposes bounded reverse-ETL write actions,
- native writer validates closed row-DML/truncate schemas and blocks arbitrary SQL/unsafe identifiers,
- writer builds parameterized SQL templates without inlining values,
- fixture-mode writes return deterministic counts without DB calls,
- Read applies caller `req.Limit` before connector `read_limit`.

### Slice B — bounded write implementation

Add native postgres `Write`, `ValidateWrite`, and `DryRunWrite` for five closed schemas only:

- `insert_row`
- `update_row`
- `upsert_row`
- `delete_row`
- `truncate_table`

All identifiers are validated/quoted, values are bind parameters, delete/update/upsert require explicit key arrays, and truncate requires a typed `confirm_phrase` record field. No arbitrary statement text is accepted.

### Slice C — definition/docs/API surface

- Update metadata/docs/spec surfaces to show native tier-3 dynamic schema plus fixture-only row write support.
- Add `writes.json`, `examples.json`, `cli_surface.json` with command metadata that still routes through plan -> preview -> approval -> execute.
- Generate operation ledger rows from official docs and landed audit counts; covered rows only for the five implemented SQL write actions, all other official operations remain blocked or evidence-backed disallowed.

### Slice D — verification and issue addendum

- Run focused tests and schema/conformance gates before broad gates.
- Regenerate postgres docs/skills/catalog surfaces only if generators report drift.
- Append an idempotent captain-policy addendum to issues #3118-#3125 through `gh-axi`, preserving bodies and truthful counts.
- Make one clean local commit; do not push.

### Resume addendum — typed MERGE correction

After upstream termination, preserve the existing PostgreSQL parity changes and finish the `upsert_row` correction so official `MERGE` coverage is truthful:

- `upsert_row` must emit a closed-schema, parameterized PostgreSQL `MERGE` statement, not `INSERT ... ON CONFLICT`.
- Update the upsert unit test and write fixture template to the same typed `MERGE` SQL.
- Re-run connector-local tests, conformance, connectorgen validation, build, boundary, `make verify`, and `git diff --check` before committing.
