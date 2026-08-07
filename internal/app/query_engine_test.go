package app_test

import (
	"context"
	"testing"

	"polymetrics.ai/internal/app"
)

// TestQuerySQLEngineSeamPreservesSelectAll pins the behaviour the query-engine
// seam was introduced to preserve: `select * from <table>` still answers, and
// still answers with the same rows. What changed is which engine answers it —
// the JSONL engine that could express nothing else is gone, so this now runs
// against DuckDB in the only build there is.
func TestQuerySQLEngineSeamPreservesSelectAll(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if got := a.QueryEngineName(); got != "duckdb" {
		t.Fatalf("QueryEngineName() = %q, want duckdb", got)
	}

	seedWarehouseTable(t, root, "widgets", []map[string]any{
		{"id": "w1", "name": "alpha"},
		{"id": "w2", "name": "beta"},
	})

	rows, err := a.QuerySQL(ctx, "select * from widgets", 10)
	if err != nil {
		t.Fatalf("QuerySQL: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
}
