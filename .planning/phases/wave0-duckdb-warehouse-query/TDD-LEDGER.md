# TDD Ledger

Phase: wave0-duckdb-warehouse-query

Record failing test evidence before production code for every behavior-adding task.

---

## b-engine-seam
- Test: `internal/app/query_engine_test.go` → `TestQuerySQLEngineSeamPreservesSelectAll`
- Command: `go test ./internal/app/ -run TestQuerySQLEngineSeamPreservesSelectAll`
- Red output (before code):
  ```
  internal/app/query_engine_test.go:49:14: a.QueryEngineName undefined (type *app.App has no field or method QueryEngineName)
  FAIL  polymetrics/internal/app [build failed]
  ```
- Green output (after code):
  ```
  --- PASS: TestQuerySQLEngineSeamPreservesSelectAll (0.03s)
  ok    polymetrics/internal/app
  ```
- Note: test file scoped to the default lane via `//go:build !duckdb` (assertion `=="jsonl"` unchanged);
  shared `seedWarehouseTable` helper moved to untagged `query_engine_helpers_test.go` so the duckdb test still compiles.
- Status: green

## b-duckdb-engine
- Test: `internal/app/query_engine_duckdb_test.go` (`//go:build duckdb`) → `TestDuckDBJoinAndAggregate`, `TestDuckDBSelectOnlyRejectsMutation`
- Command: `CGO_ENABLED=1 go test -tags duckdb ./internal/app/ -run TestDuckDB`
- Red output (before code):
  ```
  internal/app/query_engine_duckdb_test.go:26:14: a.QueryEngineName undefined (type *app.App has no field or method QueryEngineName)
  FAIL  polymetrics/internal/app [build failed]
  ```
- Green output (after code, `CGO_ENABLED=1 go test -tags duckdb ./internal/app/ -run TestDuckDB`):
  ```
  --- PASS: TestDuckDBJoinAndAggregate (0.04s)
  --- PASS: TestDuckDBSelectOnlyRejectsMutation (0.03s)
  ok    polymetrics/internal/app
  ```
- Status: green

## b-verify-lane
- Covered by `t-both-green`: `make verify-duckdb` does not exist yet, and the duckdb-tagged tests
  above fail to build until the engine lands. The lane is exercised green only after b-duckdb-engine.
- Green output (after code): `make verify-duckdb` exits 0; default `make verify` exits 0.
- Status: green
