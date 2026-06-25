# Agent Trace: backend

Phase: wave0-duckdb-warehouse-query · Stage: EXECUTE · Date: 2026-06-25

## Rendered Prompt Or Prompt Reference

BACKEND role: implement the DuckDB warehouse-query engine behind a build tag, test-first, keeping both
build configurations green. Authoritative design: `SPEC.md`. Plan: `PLAN.md`. Constraints in `ADR.md`
and `THREAT-MODEL.md`.

## Files Inspected

- `.planning/phases/wave0-duckdb-warehouse-query/{SPEC,PLAN,ADR,THREAT-MODEL,TDD-LEDGER}.md`
- `internal/app/query_engine_test.go`, `internal/app/query_engine_duckdb_test.go` (red tests)
- `internal/app/app.go` (App struct, `Open`, `QueryTable`, old `QuerySQL`)
- `internal/app/util.go` (`parseSelectAll`), `internal/connectors/connectors.go` (`Record`)
- `Makefile`, `go.mod`

## Actions Taken

### Wave A — engine seam (default build, no new dep)
- `internal/app/app.go`: added `sqlEngine sqlQueryEngine` field + `sqlQueryEngine` interface
  (`QuerySQL(ctx, sql, limit) ([]connectors.Record, error)` / `Name() string`); set
  `a.sqlEngine = newSQLEngine(a)` in `Open()`; `QuerySQL` now delegates to `a.sqlEngine.QuerySQL`;
  added `QueryEngineName()`.
- Added `internal/app/query_engine_default.go` (`//go:build !duckdb`): `newSQLEngine` -> `jsonlEngine`
  reproducing the historical `parseSelectAll` -> `QueryTable` behavior; `Name()=="jsonl"`. CGO-free.

### Wave B — DuckDB engine (tagged)
- `go get github.com/marcboeker/go-duckdb@v1.8.5` + `go mod tidy` (toolchain auto-fetch; see go.mod note).
- Added `internal/app/query_engine_duckdb.go` (`//go:build duckdb`): `newSQLEngine` -> `duckdbEngine`
  with `warehouseDir = filepath.Join(a.projectDir, "warehouse")`; `Name()=="duckdb"`. In-memory DuckDB via
  `database/sql` (`sql.Open("duckdb", "")`, blank import of go-duckdb). Behavior:
  - `validateSelectOnly`: must start (trimmed, case-insensitive) with `select`/`with`; rejects `;` and the
    whole-word tokens insert/update/delete/drop/alter/create/attach/copy/pragma/call/export/install/load/set.
  - Per `*.jsonl` in warehouse dir: table = filename sans ext, validated `^[A-Za-z0-9_]+$`; empty files
    skipped; `CREATE VIEW "<table>" AS SELECT * FROM read_ndjson_auto('<abs path>')` with the path passed
    as a single-quote-escaped string literal (no user-SQL interpolation into view creation).
  - Top-level `LIMIT` detection (paren-depth aware); if absent and `limit>0`, wrap
    `SELECT * FROM (<sql>) AS _pm_q LIMIT <n>`.
  - Generic row scan (`rows.Columns()` + `*any`) into `connectors.Record`; `normalizeValue`:
    []byte->string, time.Time->RFC3339, `*big.Int`->int64 (or float64 if it overflows), uint64->int64/float64.
- `Makefile`: added `verify-duckdb` target (TAB recipe: CGO build of `./cmd/pm` + `go test -tags duckdb ./...`),
  added it to `.PHONY`, `verify` left unchanged; added `export GOTOOLCHAIN ?= auto` (toolchain note below).

### Test scoping (no assertion changes)
- `internal/app/query_engine_test.go`: added `//go:build !duckdb` so the default-only seam test (asserts
  `QueryEngineName()=="jsonl"`) does not run under `-tags duckdb` where the engine is "duckdb".
- Moved the shared `seedWarehouseTable` helper into untagged `internal/app/query_engine_helpers_test.go`
  so the duckdb-tagged test still compiles. No test logic/assertions weakened.

## Commands Run

- `gofmt -l internal cmd` -> clean
- `go vet ./...` -> clean; `CGO_ENABLED=1 go vet -tags duckdb ./internal/app/` -> clean
- `go test ./...` -> all ok; `make verify` -> exit 0
- `CGO_ENABLED=0 go build ./...` -> ok; `go list -deps ./cmd/pm | grep duckdb` -> 0 packages
- `make verify-duckdb` -> exit 0
- Type probe: confirmed DuckDB `SUM(int)` returns `*big.Int` (drove `normalizeValue`).

## Findings

- go-duckdb v1.8.5's module declares `go 1.24`. Go MVS forces the main module's `go` directive to
  `>= 1.24`, so adding the (human-approved) dependency bumped `go.mod`: `go 1.23.0`/`toolchain go1.23.4`
  -> `go 1.24`/`toolchain go1.24.0`. Transitive (indirect) deps added: apache/arrow-go/v18, goccy/go-json,
  google/flatbuffers, klauspost/compress, klauspost/cpuid/v2, pierrec/lz4/v4, zeebo/xxh3,
  go-viper/mapstructure/v2 (+ upgrades to golang.org/x/{exp,mod,tools}, grpc, genproto). go.sum updated.
- DuckDB `SUM` over integers returns `*big.Int`; the test's `toFloat` does not handle it, so
  `normalizeValue` collapses `*big.Int` to int64 (float64 on overflow) — rows match JSONL-engine types.

## Handoff Summary

Both build lanes are green. Default build is pure-Go / CGO-free and never compiles go-duckdb (verified via
`CGO_ENABLED=0` build and an empty duckdb dep set in `go list -deps ./cmd/pm`). The tagged lane links the
prebuilt static lib and passes the JOIN/aggregate and SELECT-only tests. Reverse-ETL and `QueryTable` paths
untouched.

## Verification Evidence

`make verify` (tail):
```
test -s "$SMOKE_DIR/.polymetrics/outbox/customers_to_outbox.jsonl"; \
printf 'smoke ok: %s\n' "$SMOKE_DIR"
smoke ok: /var/folders/.../tmp.q6Dizi9N4Y
```
exit 0.

`make verify-duckdb` (tail):
```
CGO_ENABLED=1 go build -tags duckdb ./cmd/pm
CGO_ENABLED=1 go test -tags duckdb ./...
?   	polymetrics/cmd/pm	[no test files]
ok  	polymetrics/internal/app	1.025s
ok  	polymetrics/internal/cli	2.959s
...
ok  	polymetrics/internal/vault	(cached)
```
exit 0.

Tagged tests verbose:
```
--- PASS: TestDuckDBJoinAndAggregate (0.04s)
--- PASS: TestDuckDBSelectOnlyRejectsMutation (0.03s)
```
Default seam test: `--- PASS: TestQuerySQLEngineSeamPreservesSelectAll`.

## Unresolved Risks

- Toolchain: go.mod now requires `go 1.24`. An ambient `GOTOOLCHAIN=local` pinned to go1.23.x can no longer
  build the module at all (default lane included). Mitigated by `export GOTOOLCHAIN ?= auto` in the Makefile
  so both verify targets fetch go1.24.0 when needed. CI/dev environments hard-pinning `GOTOOLCHAIN=local`
  to <1.24 must install go1.24+ or allow auto. This is a direct consequence of the approved go-duckdb add,
  not a code defect — flagging for the build owner.
