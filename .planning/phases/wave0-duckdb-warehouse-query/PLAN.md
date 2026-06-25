# PLAN — DuckDB Warehouse Query (build-tag gated)

Behavior tasks require red-first evidence in TDD-LEDGER.md.

## Wave A — engine seam (default build, no new dep)
- [ ] id: t-engine-seam type: test — Red test (default build): `App.QuerySQL` delegates to a
      pluggable `sqlEngine`; existing `select * from <table>` behavior preserved.
- [ ] id: b-engine-seam type: behavior — Add `sqlQueryEngine` interface + `query_engine_default.go`
      (`//go:build !duckdb`) jsonlEngine reproducing current parseSelectAll→QueryTable; wire
      `App.sqlEngine` in `Open()`; `QuerySQL` delegates. No behavior change in default build.

## Wave B — DuckDB engine (tagged)
- [ ] id: t-duckdb-query type: test — Red test (`//go:build duckdb`, internal/app): write two JSONL
      tables to a temp warehouse, run a JOIN + GROUP BY via `QuerySQL`, assert the aggregate; plus a
      SELECT-only rejection test. Red until the duckdb engine exists.
- [ ] id: b-duckdb-engine type: behavior — Add dependency go-duckdb@v1.8.5; `query_engine_duckdb.go`
      (`//go:build duckdb`): in-memory DuckDB, `read_json_auto` views per warehouse table,
      `validateSelectOnly`, limit wrap, row scan → connectors.Record. `newSQLEngine` (duckdb variant).
- [ ] id: b-verify-lane type: behavior — Add `make verify-duckdb` (CGO build + `go test -tags duckdb ./...`).

## Wave C — Verification
- [ ] id: t-both-green type: test — Default `make verify` green (CGO-free, unchanged); `make
      verify-duckdb` green (JOIN/GROUP BY/window correct; SELECT-only enforced).

## Ordering
A → B → C. Wave A is dependency-free and preserves default behavior. Wave B adds the gated CGO path.
Heavy implementation delegated to a backend subagent; orchestrator writes red tests + gates + verify.

## Rollback
Remove the `duckdb`-tagged files + `go mod tidy` to drop the dependency; default build is unaffected.
