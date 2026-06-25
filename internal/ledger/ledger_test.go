package ledger

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestJSONLedgerAppend(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.jsonl")
	l := JSONLedger{Path: path}
	err := l.Append(context.Background(), RunRecord{
		ID:             "run_test",
		Mode:           "dependency-free",
		Operation:      "etl",
		Status:         "completed",
		RecordsRead:    3,
		RecordsWritten: 3,
		Duration:       int64(time.Millisecond),
	})
	if err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(b), `"run_test"`) {
		t.Fatalf("ledger missing run id: %s", string(b))
	}
}
