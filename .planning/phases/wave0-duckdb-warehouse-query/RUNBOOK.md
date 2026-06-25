# RUNBOOK / Rollback — DuckDB query engine

## Build modes
- Default (recommended for distribution): `make build` / `go build ./cmd/pm` — pure-Go, CGO-free,
  JSONL query behavior. No go-duckdb compiled.
- Analytics: `CGO_ENABLED=1 go build -tags duckdb ./cmd/pm` — DuckDB-backed `pm query run --sql`.

## Verify
- `make verify` (default) and `make verify-duckdb` (tagged) both green.

## Rollback
- Delete `internal/app/query_engine_duckdb*.go`, run `go mod tidy` to drop go-duckdb, remove the
  `verify-duckdb` target. Default build path is unaffected (the seam + jsonl engine remain harmless).

## Operational notes
- The tagged binary requires the bundled DuckDB static lib (linked at build); the running binary has
  no external runtime dependency.
- Queries are read-only (SELECT/WITH); enforced by `validateSelectOnly`.
