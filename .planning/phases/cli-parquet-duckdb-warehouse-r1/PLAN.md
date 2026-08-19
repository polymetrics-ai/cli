# Plan — Parquet tables, DuckDB engine

Phase: `cli-parquet-duckdb-warehouse-r1`
Branch: `fm/cli-parquet-duckdb-warehouse-r1`
Builds on: #3901 (`fcff76a73`, per-connection warehouse nesting)

## GSD provenance

**Manual-GSD fallback**, recorded here rather than left implicit.

`scripts/gsd doctor` is healthy on this branch and `scripts/gsd sources plan-phase` resolves to
`.gsd/commands.json`, `.gsd/upstream.lock.json` and `.gsd/official-docs/COMMANDS.md`. The lifecycle
was run inline — discuss, plan, execute test-first, verify — rather than by spawning GSD role
agents, because the canonical delivery contract forbids spawning an orchestrator, planner, reviewer
or verifier for a single-worker job and the dispatch brief carries the discuss-phase content
already. This file plus `TDD-LEDGER.md` and `VERIFICATION.md` are the phase artifacts that
lifecycle would have produced.

A supervisor brief cannot waive the lifecycle. This one did not try to; it named the contract as
authoritative over itself.

## Goal

Make Parquet the warehouse's table format and DuckDB its query engine, in one change, across all
three read paths: ETL materialization, reverse ETL, and query.

Out of scope, parked by the captain: search, embeddings, vector indexes. Nothing here pre-empts
them.

## What is already true, verified in the code before relying on it

- `internal/warehouse/layout.go` owns the layout. Tables live at
  `<workspace>/<connector>/<connection>/tables/<table>.jsonl`; the WAL at `wal/<stream>.jsonl`.
- The WAL is opened `O_APPEND` and fsynced per batch (`local_warehouse.go:109`, `:257`). Parquet
  files cannot be appended to once closed, so **the WAL stays JSONL**. This is what makes the table
  format switchable at all, not a compromise.
- The deduped table is already rebuilt wholesale from the WAL and renamed into place
  (`materializeDedupedFinal`, `local_warehouse.go:365-403`; rename at `:290`). Confirmed: the format
  is a derived detail.
- Non-deduped append modes stream into the table `O_APPEND` (`local_warehouse.go:122-135`). This is
  the one path that is **not** already wholesale, and it has to become wholesale.
- The shipped release binary is CGO-free with `newSQLEngine` returning `jsonlEngine`
  (`query_engine_default.go`), whose entire capability is `SELECT * FROM <table> [LIMIT n]`
  (`util.go:417 parseSelectAll`). `.goreleaser.yaml` sets `CGO_ENABLED=0` and no `duckdb` tag, so
  **no released `pm` has ever had a real SQL engine.**

## Decision 1 — a table is a single Parquet file, not a directory

`tables/<table>.parquet`. Settled with measurement, not preference. Trace:
`traces/layout-decision-experiment.log`.

The directory form `tables/<table>/part-0000.parquet` was proposed for three benefits. Measured:

| claimed benefit | measured result |
|---|---|
| parallel read across parts | **not real.** 1 M rows, whole-row scan: single file 609 ms, 9-part glob 600 ms — 1.5%, inside noise. DuckDB already parallelises across the 9 row groups *inside* the single file; parts add no parallelism it did not have. |
| parallel write across parts | **not real, and costly.** 1 M rows: single file 173 ms, 9 parts 2.517 s — 14.5× slower. There is no parallel producer to exploit: the materializer folds one WAL sequentially by primary key. |
| a home for things describing the data | **already available.** The search research sites indexes at `index/<table>.duckdb`, a sibling *inside the same connection directory*, which inherits `owner.json` and `AssertOwnedTable` identically. A table directory is not needed for it. |

And the directory form costs something concrete that the file form gives for free:

```
file  rename(tmp -> existing file):            err=<nil>          content after swap: "new"
dir   rename(tmp -> existing non-empty dir):   err=file exists
dir   between the two renames: parts visible=0, stat(table dir)=no such file or directory
```

A directory cannot be swapped into place atomically. The two-rename workaround opens a window in
which a concurrent reader sees **no table at all** while its rows sit intact on disk — a silent
wrong answer, which is the exact failure class #3901 exists to remove. Single-file `os.Rename` has
no such window.

**Partitioning is not foreclosed.** `read_parquet` takes a glob, and file-versus-directory is
decidable with one `os.Stat` — so a future partitioned form is *detectable, never guessed*, the same
principle #3901 established. The retrofit is bounded as long as table resolution goes through
`warehouse.Tables`/`FindTable` rather than an ad-hoc glob, which this change preserves.

## Decision 2 — DuckDB is unconditional; the build tag goes

The captain ruled DuckDB a must. Keeping `-tags duckdb` optional would mean the two builds write
**different table formats**, which is the install-time drift that `AGENTS.md` *Command Surface Must
Stay Executable* exists to prevent. The alternative — a second, pure-Go Parquet implementation for
the CGO-free build — adds a dependency that duplicates what DuckDB already does and can drift from
it.

So: one binary, one format, one engine. `query_engine_default.go` is deleted, `jsonlEngine` with it.
DuckDB writes and reads every Parquet table, so there is exactly one Parquet implementation in the
process.

Consequence, stated rather than allowed to drift: **`pm` now needs a C toolchain to build from
source.** It keeps zero *runtime* dependencies — no network, no services — which is what
"dependency-free" means to a user. `AGENTS.md` gets that line.

## Decision 3 — platform matrix

Measured on darwin/arm64 with the release ldflags (`-s -w` + version stamps):

| build | bytes | |
|---|---|---|
| `CGO_ENABLED=0`, no tag (today's release) | 79,040,146 | 75.4 MiB |
| `CGO_ENABLED=1 -tags duckdb` | 111,529,074 | 106.4 MiB |
| **delta** | **+32,488,928** | **+31.0 MiB, +41.1%** |

Matches the research's independently measured +31.0 MB / +42%.

`go-duckdb@v1.8.5` ships `deps/windows_amd64/libduckdb.a` and its production constraint is
`darwin || (linux && (amd64 || arm64)) || (freebsd && amd64) || (windows && amd64)`. The only
`!windows` tag in the module is on a test file. **`windows/arm64` is the only target with no bundled
library.**

This machine has no mingw-w64 and no zig, so windows/amd64 cannot be built here — the same toolchain
gap that produced a wrong answer in the research's earlier round. **So it is verified in CI, on a
native `windows-latest` runner, and the drop is limited to `windows/arm64` only.** If that job
fails, the casualty is all of Windows rather than Windows-on-ARM, which is a materially different
trade and reopens this decision.

## Decision 4 — existing warehouses

**No released `pm` has ever written a nested-layout table.** The last release is `v0.1.1`
(`.release-please-manifest.json`, `CHANGELOG.md`); #3901 landed after it and carries no tag. So:

- **Released warehouses** are the flat legacy layout. `CheckLegacyLayout` already refuses them with
  a rebuild instruction. **This change compounds nothing** — the same single rebuild that #3901's
  release note already asks for produces Parquet tables directly.
- **Unreleased-`main` warehouses** are nested with JSONL tables. A `tables/*.jsonl` is detected and
  **refused by name**; it is neither read nor deleted. Reading it would work today and be silently
  stale the moment a sync writes the Parquet beside it — two files, one table name, one of them
  wrong. The WAL is untouched, so re-running the sync rebuilds the table losslessly.

Detect, never guess. Refuse rather than silently rewrite. Nothing is deleted on the operator's
behalf.

## Work items

1. `internal/warehouse`: `TableFileExt = ".parquet"`; `TablePath` returns it; `Tables` globs
   `*.parquet`; a pre-Parquet `tables/*.jsonl` raises a named refusal.
2. `internal/warehouse/parquet.go`: the single Parquet read/write implementation, DuckDB-backed,
   with zero rows preserving today's empty-table semantics.
3. `internal/app/local_warehouse.go`: materialize **every** mode wholesale from the WAL. Deduped
   keeps its primary-key fold and sorted output; append modes keep WAL order. `RecordsLoaded`
   semantics unchanged.
4. `internal/connectors`: `Warehouse.Read` reads Parquet through the shared implementation.
5. `internal/app`: delete the engine build tag and `jsonlEngine`; `registerViews` uses
   `read_parquet`.
6. Build: `Makefile`, `.goreleaser.yaml`, release and verify workflows — CGO on, native runners,
   `windows/arm64` dropped, windows/amd64 DuckDB build verified.
7. Docs/website/help parity per `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.

## Required skills loaded

`golang-how-to` routing per `.agents/agentic-delivery/references/required-skills-routing.md`, then
`golang-testing` (test-first, table-driven, `-timeout 20m`), `golang-error-handling` (`%w`,
`errors.As` on the named refusal types), `golang-database` (`database/sql` over DuckDB, row
scanning, resource lifecycle), `golang-safety` (defer-in-loop, nil handling on scanned values),
`golang-security` (no SQL interpolation of user input; every path a validated literal).

## Test plan — red first, on returned records

Every assertion is on **returned records or on-disk content**. Not one asserts exit status: a
`exit 0` check would have passed against both defects fixed in #3901.

| Behaviour | Test |
|---|---|
| A sync materializes a real Parquet file holding the rows | `TestWarehouseMaterializesTablesAsParquet` |
| The table is one file, never a directory | `TestTablePathIsASingleParquetFile` |
| DuckDB answers aggregates the old ceiling could not express | `TestQuerySQLAggregatesOverParquetTables` |
| Reverse ETL reads a Parquet source table end to end | `TestReverseETLReadsAParquetSourceTable` |
| Append mode rebuilds from the WAL without duplicating rows | `TestAppendModeRebuildsFromWALWithoutDuplicating` |
| A pre-Parquet JSONL table is refused, named, and left on disk | `TestPreParquetJSONLTableIsRefusedAndLeftOnDisk` |
| A zero-row sync still materializes a readable, listed table | `TestZeroRowSyncMaterializesAReadableEmptyTable` |
| Two connections still keep their own rows, now in Parquet | existing #3901 isolation tests, unweakened |

No test is weakened, skipped, or deleted. #3901's isolation suite must keep passing against the new
format; where it reads table files by extension the helper learns Parquet rather than loosening its
assertion.
