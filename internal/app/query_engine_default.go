//go:build !duckdb

package app

import (
	"context"

	"polymetrics/internal/connectors"
)

// newSQLEngine returns the default JSONL-backed query engine. It reproduces the
// historical QuerySQL behavior (parseSelectAll -> QueryTable) and keeps the
// default build pure-Go / CGO-free.
func newSQLEngine(a *App) sqlQueryEngine {
	return jsonlEngine{app: a}
}

// jsonlEngine answers SELECT * FROM <table> [LIMIT n] queries by reading the
// JSONL warehouse table directly, exactly as App.QuerySQL did before the engine
// seam was introduced.
type jsonlEngine struct {
	app *App
}

func (e jsonlEngine) Name() string { return "jsonl" }

func (e jsonlEngine) QuerySQL(ctx context.Context, sql string, limit int) ([]connectors.Record, error) {
	table, parsedLimit, err := parseSelectAll(sql)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = parsedLimit
	}
	return e.app.QueryTable(ctx, QueryTableRequest{Table: table, Limit: limit})
}
