//go:build !duckdb

package app_test

import (
	"context"
	"testing"

	"polymetrics/internal/app"
)

// TestQuerySQLEngineSeamPreservesSelectAll is the red-first test for the
// pluggable query-engine seam: after introducing App.sqlEngine, the default
// build must preserve today's `select * from <table>` behavior and report the
// jsonl engine.
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

	if got := a.QueryEngineName(); got != "jsonl" {
		t.Fatalf("QueryEngineName() = %q, want jsonl (default build)", got)
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
