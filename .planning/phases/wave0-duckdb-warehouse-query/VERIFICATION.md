# Local Verification

- CI detected: no (local `make verify` / `make verify-duckdb` are the gates)
- Verified by: orchestrator (independent of backend agent) — ran both lanes + e2e + read the engine code.

| Check | Status | Command | Result |
| --- | --- | --- | --- |
| Default lane (CGO-free) | passed | `make verify` | exit 0; smoke ok |
| Default is CGO-free | passed | `CGO_ENABLED=0 go list -deps ./cmd/pm \| grep -c duckdb` | 0 (go-duckdb not compiled) |
| Tagged lane (CGO) | passed | `make verify-duckdb` | exit 0 |
| DuckDB analytics tests | passed | `go test -tags duckdb ./internal/app/ -run TestDuckDB` | TestDuckDBJoinAndAggregate PASS, TestDuckDBSelectOnlyRejectsMutation PASS |
| Red-first tests now green | passed | both lanes | seam + duckdb tests pass, unweakened |

## End-to-end evidence (tagged binary)
- `pm query run --sql "SELECT c.name, SUM(o.amount) AS total FROM orders o JOIN customers c USING(customer_id) GROUP BY c.name ORDER BY total DESC"` → `[{name:Ada,total:150},{name:Grace,total:70}]` — real JOIN + GROUP BY aggregation over the JSONL warehouse.
- `pm query run --sql "DROP VIEW orders"` → `error: only SELECT/WITH queries are allowed` (SELECT-only safety).

## Orchestrator review (VERIFY)
GO (with one warning). Reviewed `internal/app/query_engine_duckdb.go`:
- SELECT-only: rejects empty/`;`-chained/non-SELECT-WITH + whole-word DDL/DML tokens (insert..set). Sound.
- Injection safety: view names validated `^[A-Za-z0-9_]+$`; file paths quote-escaped literals; user SQL never interpolated into `CREATE VIEW`. Read functions bounded to the warehouse dir.
- Correctness: `normalizeValue` collapses HUGEINT (`*big.Int` from SUM) to int64/float64, []byte→string, time→RFC3339 — rows interchangeable with the JSONL engine. `hasTopLevelLimit` is depth-aware.

## WARNING (build-contract change)
go-duckdb v1.8.5 declares `go 1.24`, so the module `go` directive was bumped `1.23 → 1.24` and
`toolchain go1.24.0`. This affects the DEFAULT build too: with `GOTOOLCHAIN=local` on a sub-1.24 Go,
direct `go build`/`go test` now fail. Mitigation in place: the Makefile exports `GOTOOLCHAIN ?= auto`,
so `make verify`/`make verify-duckdb` auto-fetch go1.24.0 and both pass. Decision for the owner: keep
go1.24 (current), or pin an older go-duckdb that declares ≤go1.23 to preserve the go1.23 floor.
