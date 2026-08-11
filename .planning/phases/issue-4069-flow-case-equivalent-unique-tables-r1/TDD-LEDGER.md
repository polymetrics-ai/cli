# #4069 TDD ledger

**Fresh delivery lineage:** 0 / 5 corrections used
**Specification owner:** #4066 at its terminal 5 / 5; this is not loop 6
**Starting head:** `659efd8a0d69f26b55fcbd3c02150e995c159519`

| Slice | RED contract | GREEN contract | Status |
|---|---|---|---|
| 1. Two exact-unique canonical equivalents | Real Parquet `acme/records` and `globex/RECORDS` cause generic `SELECT 1` to fail while an omitted flow surfaces raw DuckDB duplicate-view text. | Scoped reads return their owner rows, generic `SELECT 1` succeeds, and omitted flow reads return typed `*warehouse.AmbiguousTableError` with the flow `connection` remedy. | GREEN 2026-08-12 |
| 2. Existing #4066 boundary | The correction must not alter the existing alias, three-table case-variant, exact ambiguity, action-source, reverse/read, or schedule behavior. | Short focused fence passes before CPU pause; full affected/race selectors run after the transport CPU gate. | Targeted GREEN; full fence pending CPU gate |
| 3. Refactor | No refactor is accepted before GREEN. | The policy remains snapshot-derived, deterministic, and free of SQL-text filtering or extra inventory scans. | GREEN review pending CPU gate |

## RED command

```text
go test -timeout 20m ./internal/cli -run '^TestFlowCaseEquivalentUniqueTablesPreserveGenericSQLAndTypedAmbiguity$' -count=1
```

Expected RED: exit 1. The fixture must establish that both exact spellings are
individually unique, the selected reads work, then show `SELECT 1` fails at
`register view` with DuckDB's canonical duplicate error and the omitted flow
does not contain `*warehouse.AmbiguousTableError`.

## GREEN command

```text
go test -timeout 20m ./internal/cli -run '^TestFlowCaseEquivalentUniqueTablesPreserveGenericSQLAndTypedAmbiguity$' -count=1
```

Expected GREEN: exit 0. The records remain isolated by selected connection,
the unrelated generic query succeeds, and each omitted-flow assertion exposes
the typed ambiguity plus the truthful flow manifest remedy with no checkpoint
or returned rows.

## Evidence rules

The committed RED must be captured before any production implementation edit.
The only allowed local state is temporary test-owned Parquet/DuckDB data; no
credential, provider, reverse-ETL, action dispatch, or transport state is
created. Trace files contain command/result summaries only and never secrets.

## Recorded RED

At production head `f30dfcefc`, the command exited 1 as required. The selected
`acme/records` and `globex/RECORDS` subtests both passed and returned only their
own row. The unrelated generic `SELECT 1` failed before execution with:

```text
register view "records": Catalog Error: View with name "records" already exists!
```

Both omitted-flow forms then failed the typed-error assertion because the same
raw DuckDB error reached the flow engine instead of an
`*warehouse.AmbiguousTableError`. The non-secret command record is
`traces/red-case-equivalent-unique-tables.txt`.

## Recorded GREEN

The exact targeted command passed after the policy began deriving
case-equivalent exact-name groups from the existing resolver snapshot. The
fixture retained both selected owner rows, generic `SELECT 1` returned one row,
and omitted unquoted `records` plus quoted `RECORDS` each returned an error
chain containing `*warehouse.AmbiguousTableError` for `records` and the flow
`connection` remedy. The rejected paths returned no flow rows and wrote no
success checkpoint.

The short #4066 flow matrix, focused app query matrix, and schedule re-entry
selectors also passed. The full affected-package and race matrix remains
intentionally deferred to Firstmate's transport CPU gate. The non-secret
records are in `traces/green-case-equivalent-unique-tables.txt`.
