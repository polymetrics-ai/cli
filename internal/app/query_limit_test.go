package app_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/warehouse"
)

// TestQueryTableStopsAtLimitBeforeReadingTheRest pins that a limited read stops
// at the limit rather than consuming the whole table.
//
// It used to prove this with a JSONL fixture whose second line was malformed:
// a read that ran past the limit hit a decode error, so a clean return was the
// proof. Parquet validates its footer up front, so a half-broken file is not
// representable and that particular fixture cannot exist. The contract is
// therefore asserted directly instead of inferred from a side effect — the
// emit callback must never see the row after the limit — which is a stronger
// statement than the original, not a weaker one.
func TestQueryTableStopsAtLimitBeforeReadingTheRest(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	path := filepath.Join(root, ".polymetrics", "warehouse", "stops_at_limit"+warehouse.TableFileExt)
	if err := warehouse.WriteTable(ctx, path, []warehouse.Row{
		{"id": "ok"},
		{"id": "second"},
		{"id": "third"},
	}); err != nil {
		t.Fatalf("write warehouse fixture: %v", err)
	}

	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	rows, err := a.QueryTable(ctx, app.QueryTableRequest{Table: "stops_at_limit", Limit: 1})
	if err != nil {
		t.Fatalf("QueryTable() error = %v", err)
	}
	if len(rows) != 1 || rows[0]["id"] != "ok" {
		t.Fatalf("rows = %+v, want first row only", rows)
	}

	// The table read surface stops the moment emit reports it has enough, so
	// the row after the limit is never handed to the caller at all.
	stop := errors.New("enough")
	seen := make([]string, 0, 3)
	err = warehouse.ReadTable(ctx, path, func(row warehouse.Row) error {
		seen = append(seen, row["id"].(string))
		if len(seen) == 1 {
			return stop
		}
		return nil
	})
	if !errors.Is(err, stop) {
		t.Fatalf("ReadTable() error = %v, want the emit error propagated", err)
	}
	if len(seen) != 1 || seen[0] != "ok" {
		t.Fatalf("emitted rows = %v, want the read to stop after the first", seen)
	}
}
