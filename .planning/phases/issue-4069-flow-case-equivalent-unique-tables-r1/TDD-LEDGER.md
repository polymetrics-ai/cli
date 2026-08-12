# #4069 TDD ledger

**Fresh delivery lineage:** correction 1 / 5 is active; 0 / 5 corrections
were used before this correction.
**Specification owner:** #4066 at its terminal 5 / 5; this is not loop 6
**Starting head:** `659efd8a0d69f26b55fcbd3c02150e995c159519`
**Correction-1 canonical finish-plan SHA-256:**
`939f14f61defd993f8ad0335a5eb617d97083c9f73a6a75259d0e312ae8f408`

| Slice | RED contract | GREEN contract | Status |
|---|---|---|---|
| 1. Two exact-unique canonical equivalents | Real Parquet `acme/records` and `globex/RECORDS` cause generic `SELECT 1` to fail while an omitted flow surfaces raw DuckDB duplicate-view text. | Scoped reads return their owner rows, generic `SELECT 1` succeeds, and omitted flow reads return typed `*warehouse.AmbiguousTableError` with the flow `connection` remedy. | GREEN 2026-08-12 |
| 2. Existing #4066 boundary | The correction must not alter the existing alias, three-table case-variant, exact ambiguity, action-source, reverse/read, or schedule behavior. | Targeted pre-pause fence and resumed full affected/race selectors pass after the transport CPU gate. | GREEN 2026-08-12 |
| 3. Refactor | No refactor is accepted before GREEN. | The policy remains snapshot-derived, deterministic, and free of SQL-text filtering or extra inventory scans. | GREEN review 2026-08-12 |

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
selectors also passed before the CPU pause. After the gate cleared,
`internal/app`, `internal/cli`, `internal/flow`, `internal/warehouse`,
and `internal/schedule` all passed in full, and the focused
#3897/#4066/#4069 race matrix passed. The non-secret records are in
`traces/green-case-equivalent-unique-tables.txt` and
`traces/resume-broad-verification.txt`.

## Correction 1 / 5 — accepted same-owner policy

The audited d902 head leaves a same-owner destination inventory unguarded. This
is the first correction in the fresh #4069 lineage; the prior cross-owner RED
and GREEN rows remain historical evidence and must not be rewritten.

| Slice | RED contract | GREEN contract | Status |
|---|---|---|---|
| C1. New connection invariant | One local-warehouse request contains distinct streams whose effective tables are `records` and `RECORDS`; creation currently succeeds or persists. Capture state bytes/count/revision and require `errors.As` to the new type. | Creation rejects after defaults and before ID/save; all captured persisted state remains unchanged. Exact duplicate spelling and non-local destination controls preserve their current behavior. | RED recorded 2026-08-12 |
| C2. Legacy sync fence | A persisted same-owner collision opens unchanged, then its next stream run currently begins/persists a run and can mutate WAL/table state. | Open does not rewrite legacy state; either stream is rejected before `beginRun` and no run/checkpoint/stream/owner/directory/WAL/temp/Parquet state changes. | RED recorded 2026-08-12 |
| C3. SQL policy | Generic/selected bare and quoted collision references currently either bind a survivor or encounter raw registration behavior; a one-owner `AmbiguousTableError` is not a truthful remedy. | A dedicated typed same-owner collision is returned only for the colliding key; `SELECT 1`, unrelated tables, and a real generated-alias collision control remain executable. | RED recorded 2026-08-12 |
| C4. Flow and schedule boundary | Unscoped and selected bare/quoted collision flows can complete or report the wrong error/checkpoint outcome. | Each fails without a success checkpoint, including schedule re-entry; inherited cross-owner flow/action/reverse/schedule behavior remains green. | RED recorded 2026-08-12 |
| C5. Exact physical reads | A case-insensitive physical path can have one surviving spelling even though legacy state declares two. | Direct query, action, and reverse reads use only resolver-proven physical spellings and refuse the missing variant without aliasing it. | RED control passed 2026-08-12 |

### Correction 1 RED command set

The committed failing command record is
`traces/correction-1-same-owner-red.txt`. It includes the five named cases in
`PLAN.md`, selected app/CLI/flow/warehouse selectors, and the preserved
cross-owner/generated-alias controls. RED evidence uses real test-owned
Parquet/DuckDB state and asserts types plus persisted mutation boundaries, not
merely command status.

## Recorded correction 1 RED

At planning checkpoint `465e02911`, with the canonical finish-plan snapshot
SHA-256 recorded above, these focused commands exited 1 as expected:

```text
go test -timeout 20m ./internal/app -run 'Test(CreateConnectionRejectsSameOwnerCaseEquivalentDestinationTables|LegacySameOwnerCaseEquivalent)' -count=1
go test -timeout 20m ./internal/cli -run '^TestFlowLegacySameOwnerCaseEquivalentInventoryStopsAtTheTypedBoundary$' -count=1
```

The app suite proves the current code accepts a new `records`/`RECORDS`
inventory, starts a legacy run, and returns nil or an ordinary missing-table
error for generic, selected, quoted, and generated-alias SQL instead of the
required typed one-owner collision. The direct resolver-backed query/action/
reverse physical-spelling control passed. The flow suite proves omitted and
selected bare/quoted forms currently complete with nil error; its two-attempt
case is the schedule re-entry regression. Full non-secret output is recorded
in `traces/correction-1-same-owner-red.txt`. No production file is part of the
RED checkpoint.

### Correction 1 GREEN rule

GREEN may begin only after the committed RED checkpoint. The implementation
must share deterministic ASCII identifier-key semantics across creation, legacy
sync preflight, and query policy; it must not migrate/rewrite data, parse SQL,
Unicode-fold, reserve flat aliases, access credentials, or claim certification.
