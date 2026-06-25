# SPEC — DuckDB Warehouse Query (build-tag gated)

## Engine abstraction (internal/app)
Add an interface and build-tagged construction:
```go
type sqlQueryEngine interface {
    QuerySQL(ctx context.Context, sql string, limit int) ([]connectors.Record, error)
    Name() string // "jsonl" | "duckdb"
}
```
- `query_engine_default.go` (`//go:build !duckdb`): `func newSQLEngine(a *App) sqlQueryEngine`
  returns a `jsonlEngine` that reproduces TODAY's behavior exactly (parseSelectAll → QueryTable).
- `query_engine_duckdb.go` (`//go:build duckdb`): `func newSQLEngine(a *App) sqlQueryEngine`
  returns a `duckdbEngine{warehouseDir}`.
- `App` gains a `sqlEngine sqlQueryEngine` field set in `Open()` via `newSQLEngine(a)`.
- `App.QuerySQL` delegates to `a.sqlEngine.QuerySQL(ctx, sql, limit)`. `QueryTable` stays a direct
  JSONL read (works in both builds; used by reverse-ETL source reads) — unchanged.

## duckdbEngine behavior
1. Open in-memory DuckDB: `sql.Open("duckdb", "")` (database/sql; driver registered by go-duckdb).
2. For each `<table>.jsonl` in the warehouse dir, register a view:
   `CREATE VIEW "<table>" AS SELECT * FROM read_json_auto('<abs path>', format='newline_delimited')`.
   (Table identifiers validated against `[a-zA-Z0-9_]+`; path passed as a literal — quote-escaped.)
3. Validate the user SQL is SELECT-only (`validateSelectOnly`): must start with SELECT or WITH;
   reject `;`, and the tokens insert/update/delete/drop/alter/create/attach/copy/pragma/call/export.
4. If no LIMIT present and limit>0, wrap: `SELECT * FROM (<sql>) LIMIT <limit>`.
5. Run query; scan each row into `connectors.Record` keyed by column name; normalize DuckDB values
   (int64/float64/bool/string/time.Time/[]byte/nil) to JSON-friendly Go types matching existing rows.
6. Close the connection per call (engine is stateless; views are per-query over current files).

## Dependency
`go get github.com/marcboeker/go-duckdb@v1.8.5` + `go mod tidy`. Imported ONLY from
`query_engine_duckdb.go` (tag-gated), so default `go build ./...` never compiles it (CGO-free).

## Makefile
Add:
```
verify-duckdb: ; CGO_ENABLED=1 go build -tags duckdb ./cmd/pm && CGO_ENABLED=1 go test -tags duckdb ./...
```
Leave `verify` untouched (default, CGO-free).

## Out of scope
Write-path/storage changes, connectors, CDC.
