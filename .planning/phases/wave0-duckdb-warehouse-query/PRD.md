# PRD — DuckDB Warehouse Query (build-tag gated)

## Problem
"Query the extracted data and run analysis" needs real analytical SQL (joins, aggregations, window
functions). Today `app.QuerySQL` (internal/app/app.go:584) only parses `select * from <table>` and
reads JSONL — no joins/aggregations. The query/analyze step of the extract→query→analyze→act loop
is therefore very limited.

## Goal
Add a DuckDB-backed SQL query engine over the existing JSONL warehouse, **behind a `duckdb` build
tag**, so `go build -tags duckdb` enables full read-only analytical SQL while the **default build
stays pure-Go / CGO-free** with today's JSONL behavior. Approved dependency: `github.com/marcboeker/
go-duckdb` (CGO; bundles prebuilt static libs for darwin/arm64).

## Non-Goals
- Changing the ETL write path / warehouse storage format (stays JSONL this phase; Parquet/native
  DuckDB storage is a later optimization).
- Connectors, CDC, or catalog work.

## Success Metrics
- Default build (`go build ./...`, `make verify`) unchanged and green — no CGO, no go-duckdb compiled.
- `-tags duckdb` build: `QuerySQL` runs JOIN + GROUP BY + window queries over warehouse tables and
  returns correct rows; SELECT-only safety rejects INSERT/UPDATE/DELETE/DDL.
- New `make verify-duckdb` lane green (CGO build + tagged tests).

## Constraints
- SELECT-only enforcement (no mutation via query); no secrets in scope.
- Default and tagged builds both green; reverse-ETL unchanged.
