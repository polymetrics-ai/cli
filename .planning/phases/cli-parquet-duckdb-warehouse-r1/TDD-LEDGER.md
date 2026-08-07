# TDD ledger — Parquet tables, DuckDB engine

Phase: `cli-parquet-duckdb-warehouse-r1`
Branch: `fm/cli-parquet-duckdb-warehouse-r1`
Builds on: #3901 (`fcff76a73`)

## GSD provenance

**Manual-GSD fallback**, as recorded in `PLAN.md`. `scripts/gsd doctor` is healthy on this branch
and `scripts/gsd sources plan-phase` resolves; the lifecycle ran inline rather than by spawning GSD
role agents, because the canonical delivery contract forbids spawning an orchestrator, planner,
reviewer or verifier for a single-worker job. `PLAN.md`, this ledger and `VERIFICATION.md` are the
phase artifacts that lifecycle would have produced, written as the work happened rather than
retrofitted.

## Red → Green

Every assertion is on **returned records or on-disk content**. Not one asserts exit status: a
status check would have passed against both defects #3901 fixed, and it would have passed against
the write-path gap found here.

### Red — cycle 1

Committed at `test(warehouse): reproduce the JSONL table format and the SELECT * query ceiling`.
Only the new constants and the refusal type were declared, so the tests compiled and failed on
behaviour rather than on a missing symbol. Full log: `traces/red-1.log`.

```
--- FAIL: TestWarehouseMaterializesTablesAsParquet
    materialized table .../tables/records.jsonl has extension ".jsonl", want ".parquet"
--- FAIL: TestTablePathIsASingleParquetFile
    TablePath() = ".../tables/records.jsonl", want ".../tables/records.parquet"
--- FAIL: TestAppendModeRebuildsFromWALWithoutDuplicating
    read_parquet(.../records.jsonl): No magic bytes found at end of file
--- FAIL: TestZeroRowSyncMaterializesAReadableEmptyTable
    empty table .../records.jsonl has extension ".jsonl", want ".parquet"
--- FAIL: TestPreParquetJSONLTableIsRefusedAndLeftOnDisk
    Tables() error = <nil>, want *warehouse.LegacyTableFormatError
--- FAIL: TestQuerySQLAggregatesOverParquetTables
    QueryEngineName() = "jsonl", want "duckdb"
--- FAIL: TestReverseETLReadsAParquetSourceTable
    reverse source table .../records.jsonl has extension ".jsonl", want ".parquet"
```

The reverse-ETL test is worth calling out. Written to drive plan → approve → run, it **passed**
against JSONL — proving only that reverse ETL works, which was never in doubt. The brief said
verify it against Parquet rather than assume it, so the test was strengthened to assert the file it
is about to read really is Parquet. That is what turned it red.

### Green — cycle 1

`feat(warehouse)!: materialize tables as Parquet and make DuckDB the query engine`.

| Behaviour | Test |
|---|---|
| A sync leaves a real Parquet file holding the rows, WAL still JSONL | `TestWarehouseMaterializesTablesAsParquet` |
| A table is one file, never a directory | `TestTablePathIsASingleParquetFile` |
| Append mode rebuilds from the WAL without duplicating rows | `TestAppendModeRebuildsFromWALWithoutDuplicating` |
| A zero-row sync still materializes a readable, listed table | `TestZeroRowSyncMaterializesAReadableEmptyTable` |
| A pre-Parquet JSONL table is refused, named, left on disk | `TestPreParquetJSONLTableIsRefusedAndLeftOnDisk` |
| DuckDB answers GROUP BY, WHERE, aggregates; writes still refused | `TestQuerySQLAggregatesOverParquetTables` |
| Reverse ETL reads a Parquet source table end to end | `TestReverseETLReadsAParquetSourceTable` |

The Parquet assertions are made with an **independent reader** (`database/sql` over `go-duckdb`
directly, in the test), not through the writer's own code path. A test that read the file back
through the code that wrote it would prove only self-consistency.

### Red — cycle 2, found by running the binary

Driving the built `pm` through the migration case surfaced a gap the unit tests had not: the
**read** path refused a pre-Parquet warehouse, but the **write** path did not. `pm etl run` reported
`records_loaded: 3`, `status: completed`, exit 0 — into a warehouse where every read was refused.
An operator was told at once that the sync worked and that the table cannot be read.

```
--- FAIL: TestSyncRefusesAPreParquetWarehouseInsteadOfReportingSuccess
    RunETL() into a pre-Parquet warehouse error = <nil>, want *warehouse.LegacyTableFormatError
```

This is exactly the failure the brief warned about: a test asserting `exit 0` would have passed.

### Green — cycle 2

`warehouse.CheckLegacyTableFormat` added to `runWarehouseETL`, to `Warehouse.Write`, and to
`Warehouse.ValidateWrite` — the last because `ValidateWrite` exists to predict `Write`, and a
refusal that only arrived at write time would leave an approved reverse plan that can never run.
`TestWarehouseValidateWriteAgreesWithWrite` gained a pre-Parquet warehouse case so the two cannot
drift.

The refusal was also extended to **root-level** `*.jsonl`, which the first cut missed: those are the
unattributed direct writes, and a Parquet-only glob would have silently reported them absent — the
same silent-absence failure, one directory up.

### Red — cycle 3, a bounded schema inference

Reviewing the write path for portability turned up a second defect the earlier tests could not
reach, because every one of them used a handful of rows. The staged rows are converted by a JSON
reader that infers the schema from a **bounded prefix — 20,480 rows by default**. A connector field
that is sparse, or that a provider started sending partway through a backfill, first appears after
it.

```
--- FAIL: TestWriteTableKeepsAColumnThatFirstAppearsLate
    WriteTable() error = write parquet table late_column.parquet: Invalid Input Error:
    JSON transform error ... in line 25000: Object {"id":"r24999","late_field":"only here"}
    has unknown key "late_field"
--- FAIL: TestWriteTableKeepsAColumnWhoseTypeWidensLate
    WriteTable() error = ... in line 25000: Failed to cast value to numerical: "not a number"
```

The failure is loud rather than silent, which is the one mercy here — but it means **any table
larger than the sample with a late-appearing field cannot be materialized at all**. A real backfill
would hit it. The 74-row live GitHub sync could not have.

### Green — cycle 3

`sample_size=-1` on the JSON read, so inference reads every staged row. The cost is one extra pass
over a file already on disk. Row counts in both tests are deliberately past the default rather than
round numbers, so the test fails if any bound is reintroduced at any size.

### Green on arrival — the one guard that was not a fix

`TestTablePathsSurviveShellAndSQLMetacharactersInTheRoot` passed the first time it ran, and is
recorded as a guard rather than a red/green cycle so the ledger stays honest about which tests
found something.

It covers the one part of a table path pm does not generate: the **warehouse root**, supplied by the
operator with `--config path=...`, which `SafePathPart` never sees. Every Windows path contains
backslashes; a root can legitimately contain a quote or a space anywhere. All of those reach a SQL
string literal. `back\slash`, `back\the\table`, `quo'te` and `with space` all round-trip
correctly — so the literals are carried as data, not interpreted. Worth pinning precisely because
the code is correct today and a future edit could quietly stop being.

## Tests changed, and why none was weakened

| Test | Change | Intent preserved |
|---|---|---|
| `TestQuerySQLEngineSeamPreservesSelectAll` | asserts `duckdb`, not `jsonl` | Same behaviour pinned — `select * from <table>` still answers with the same rows. It now runs in the only build there is, rather than only in the untagged one. |
| `TestQueryTableStopsAtLimitBeforeLaterDecodeError` → `...BeforeReadingTheRest` | rewritten | The old fixture was a JSONL file with a malformed second line; a clean return was indirect proof the read stopped. Parquet validates its footer up front, so that fixture cannot exist. The contract is now asserted **directly** — the emit callback never sees the row after the limit — which is stronger, not weaker. |
| `readWarehouseTableRows`, `writeWarehouseRows`, `seedWarehouseTableRows`, `seedWarehouseTable`, layout and connector fixtures | read/write Parquet | Fixtures now go through the real table writer, so a fixture can no longer drift from the format a sync produces. Assertions untouched. |
| `internal/cli/testdata/golden_transcripts.json` | regenerated | Help text changed; the golden is the record of it. |

`make verify-duckdb` was removed from the Makefile. It existed only to exercise the tagged build,
which is now the default build — keeping it would run the identical suite twice. No test lost
coverage: `verify` now covers what `verify-duckdb` did.

**No test was weakened, skipped, or deleted.**

## Proof against the real binary

Beyond the unit tests, the built `pm` was driven end to end. Transcript: `traces/e2e-run.log`.

- `pm etl run` materializes `tables/sample_customers.parquet` beside `wal/customers.jsonl` and
  `owner.json`; `file(1)` reports **Apache Parquet**. Table 1,784 bytes against a 2,152-byte WAL on
  3 rows — the compression only pays at scale, and is not claimed to here.
- `pm query run --sql` answers `GROUP BY`, `WHERE` + aggregate, `ORDER BY` with projection, and a
  CTE with a window function — every one of which the shipped `SELECT * FROM t [LIMIT n]` engine
  could not express. `delete from sample_customers` is still refused.
- `pm reverse plan` → `preview` → `run` against the Parquet table wrote all three mapped records to
  the outbox, asserted on their contents.
- **Interop, the actual case for Parquet:** the table was read by two programs that are not `pm` —
  the standalone `duckdb` CLI and Python `pyarrow`. pyarrow reports 3 rows, 1 row group, `ZSTD`,
  `created_by: DuckDB version v1.1.3`.
- **Migration:** a `tables/*.jsonl` planted in a healthy warehouse makes reads refuse by name; the
  file is byte-identical afterwards; removing it restores reads. This is where cycle 2's red came
  from.
- `make smoke` — the full init → credentials → connection → catalog → ETL → reverse plan/preview/run
  path — passes against the Parquet layout.

## Proof against the live GitHub connector

The captain added this to the definition of done mid-flight: prove it against live GitHub, not
fixtures, strictly read-only. Done — full transcript in `traces/live-github-proof.md`.

- `pm etl run` against the public `golang/example` repository landed **74 pull requests** and
  **11 labels** in Parquet. Compared row by row against the API: **IDENTICAL** on every asserted
  field.
- `pm reverse plan/preview/run` out of that Parquet table wrote all 74 records to the local outbox,
  and every one was compared back to the live API: **0 disagreements, 0 missing**.
- The WAL-replay change was exercised against live data: append grew 11 → 22 → 33 rows over three
  runs with 11 distinct names throughout; deduped held at 74 rows with **0** duplicate primary keys
  over repeated runs.
- Read-only throughout. The token came from `gh auth token`, was never printed or logged, and
  carries `repo` scope — so the reverse destination was the local outbox, never GitHub. The
  repository was re-read afterwards and is unchanged.

The live run is proof the path works; it is **not** proof it scales. At 74 rows it could not have
reached the schema-inference bound cycle 3 found, which is why that defect was caught by a written
test rather than by this exercise. Both kinds of evidence were needed.

### Red — cycle 4, found by CI on the pushed branch

Three test fixtures hand-wrote a **root-level warehouse JSONL table**, which cycle 2's refusal
correctly rejects. They were missed locally because `go test ./internal/connectors/` does not
include `./internal/connectors/certify/`, and because the root-level half of the refusal landed
after the earlier `internal/cli` run.

```
--- FAIL: TestSampleOutboxWriteLifecycleAgainstRealCLI
    reverse plan create: exit 1 stderr=error: warehouse tables are stored as Parquet,
    but 1 table(s) are still JSONL (.../warehouse/cert_write_seed_sample.jsonl)
--- FAIL: TestReverseETLToGitHubCreatesPullRequestAfterApproval
--- FAIL: TestGitHubCommandWriteUsesReversePlanApproval
```

The refusal is right and the fixtures were wrong: they wrote a format the binary under test
refuses. All three now seed through `warehouse.WriteTable`, so a fixture can no longer drift from
the format a sync produces — the same treatment the layout and connector fixtures already got.

The lesson is recorded rather than just fixed: **`./internal/connectors/` does not cover
`./internal/connectors/certify/`.** Scope local runs with `/...` or name the subpackage.

### Green — cycle 4

`internal/connectors/certify` and `internal/cli` both green.

## Local verification

`gofmt` · `go vet ./...` · `go build ./cmd/pm` · `internal/app` (218s) · `internal/warehouse` ·
`internal/connectors` · `internal/cli` · `make tidy-check` · `make docs-check` · `make smoke` ·
`make agent-contract-check` · `make connectorgen-validate` · `make connectorgen-surface-sync` ·
`make connector-boundary` · `make release-workflow-check` · `make lint`.

`go test` was run with `-timeout 20m` throughout, per `AGENTS.md`.
