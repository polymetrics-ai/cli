package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/app"
)

func TestQueryTableStopsAtLimitBeforeLaterDecodeError(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	path := filepath.Join(root, ".polymetrics", "warehouse", "broken_after_first.jsonl")
	if err := os.WriteFile(path, []byte("{\"id\":\"ok\"}\n{broken\n"), 0o600); err != nil {
		t.Fatalf("write warehouse fixture: %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	rows, err := a.QueryTable(ctx, app.QueryTableRequest{Table: "broken_after_first", Limit: 1})
	if err != nil {
		t.Fatalf("QueryTable() error = %v", err)
	}
	if len(rows) != 1 || rows[0]["id"] != "ok" {
		t.Fatalf("rows = %+v, want first row only", rows)
	}
}
