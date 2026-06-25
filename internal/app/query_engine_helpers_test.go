package app_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// seedWarehouseTable writes records as a JSONL warehouse table under the project's
// warehouse dir (<root>/.polymetrics/warehouse/<table>.jsonl). It is shared by both
// the default-build seam test and the duckdb-tagged engine tests.
func seedWarehouseTable(t *testing.T, root, table string, records []map[string]any) {
	t.Helper()
	dir := filepath.Join(root, ".polymetrics", "warehouse")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir warehouse: %v", err)
	}
	f, err := os.Create(filepath.Join(dir, table+".jsonl"))
	if err != nil {
		t.Fatalf("create table file: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, rec := range records {
		if err := enc.Encode(rec); err != nil {
			t.Fatalf("encode record: %v", err)
		}
	}
}
