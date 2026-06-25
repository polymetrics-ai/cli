# SUMMARY — DuckDB Warehouse Query (build-tag gated)

Status: **completed_with_warnings** (functional goal met, both lanes green; one build-contract warning).

## What shipped
- **Pluggable query-engine seam** in `internal/app`: `sqlQueryEngine` interface + `App.sqlEngine`
  (set in `Open()`); `App.QuerySQL` delegates; `App.QueryEngineName()` reports the active engine.
- **Default engine** (`query_engine_default.go`, `//go:build !duckdb`): `jsonlEngine` reproduces
  today's `select * from <table>` behavior exactly. Default build stays pure-Go / CGO-free.
- **DuckDB engine** (`query_engine_duckdb.go`, `//go:build duckdb`): in-memory DuckDB, `read_ndjson_auto`
  views per warehouse table, `validateSelectOnly`, LIMIT-wrap, generic row scan + value normalization.
  Real JOIN/GROUP BY/window SQL over the JSONL warehouse.
- **Dependency**: `github.com/marcboeker/go-duckdb v1.8.5` (CGO; bundled static libs; imported only
  from the tagged file). **Two verify lanes**: `make verify` (default, CGO-free) and `make verify-duckdb`.

## Verification
- `make verify` exit 0 (default, CGO-free — go-duckdb not compiled). `make verify-duckdb` exit 0.
- e2e: `pm query run --sql "...JOIN...GROUP BY..."` → Ada=150, Grace=70; `DROP VIEW` rejected.
- TDD: 3 behavior tasks red-confirmed before code, now green.
- Orchestrator review GO: SELECT-only + identifier/path injection safety verified.

## Warning (decision pending)
go-duckdb v1.8.5 forces `go 1.24` on the whole module (its go.mod declares go 1.24), bumping the
default build's minimum Go from 1.23 → 1.24. Mitigated via `GOTOOLCHAIN ?= auto` in the Makefile
(auto-fetches go1.24.0). Owner decision: accept go1.24, or pin an older go-duckdb to keep the go1.23
floor (older DuckDB). See VERIFICATION.md + ADR.md.

## Boundary / deferred
- Warehouse storage stays JSONL (DuckDB queries via views). Parquet/native-DuckDB storage deferred.
- DuckDB is query-only here; not wired into ETL write path.

## Next
- Owner: confirm go1.24 acceptance (or request older-go-duckdb pin).
- Wave 1 connectors (GA) on the github template + connsdk + registryset codegen.
