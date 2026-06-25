# TEST-PLAN — DuckDB Warehouse Query

## Gates
- Default: `make verify` (CGO-free; gofmt+vet+`go test ./...`+build+smoke) — must stay green/unchanged.
- Tagged: `make verify-duckdb` = `CGO_ENABLED=1 go build -tags duckdb ./cmd/pm` + `go test -tags duckdb ./...`.

## Red-first tests
1. **Engine seam (default)** `internal/app/query_engine_test.go`: assert `App.QuerySQL("select * from
   t")` still returns the table rows after the seam is introduced (behavior preserved). Red until the
   `sqlEngine` seam + default engine exist.
2. **DuckDB analytics** `internal/app/query_engine_duckdb_test.go` (`//go:build duckdb`):
   - Seed `orders.jsonl` (order_id, customer_id, amount) + `customers.jsonl` (customer_id, name).
   - `SELECT c.name, SUM(o.amount) AS total FROM orders o JOIN customers c USING(customer_id) GROUP BY c.name ORDER BY total DESC` → assert aggregated totals.
   - A window-function query (e.g. ROW_NUMBER) returns ranked rows.
   - SELECT-only: `INSERT INTO orders ...` and `DROP VIEW orders` are rejected with an error.
   Red until the duckdb engine exists (`go test -tags duckdb ./internal/app/` fails to compile).

## Parity / regression
- All existing internal/app, internal/connectors, internal/cli tests stay green in BOTH builds.
- `parseSelectAll`-style `select * from <table>` still works in both builds.

## Evidence
Red captured in TDD-LEDGER.md (Status: red-confirmed) before code; green after.
