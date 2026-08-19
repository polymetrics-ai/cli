package app_test

import (
	"context"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/warehouse"
)

// seedWarehouseTable writes records as a warehouse table under the project's
// warehouse dir (<root>/.polymetrics/warehouse/<table>.parquet). These are the
// unattributed root-level tables no connection owns, seeded by hand rather than
// by a sync — so they go through the same Parquet writer a sync would use, and
// the test never depends on a second Parquet implementation.
func seedWarehouseTable(t *testing.T, root, table string, records []map[string]any) {
	t.Helper()
	path := filepath.Join(root, ".polymetrics", "warehouse", table+warehouse.TableFileExt)
	rows := make([]warehouse.Row, 0, len(records))
	for _, record := range records {
		rows = append(rows, warehouse.Row(record))
	}
	if err := warehouse.WriteTable(context.Background(), path, rows); err != nil {
		t.Fatalf("seed warehouse table %s: %v", table, err)
	}
}
