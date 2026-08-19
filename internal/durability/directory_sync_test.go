package durability_test

import (
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/durability"
)

func TestSyncDirectory(t *testing.T) {
	if err := durability.SyncDirectory(t.TempDir()); err != nil {
		t.Fatalf("SyncDirectory() error = %v", err)
	}
}

func TestSyncDirectoryReportsMissingDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing")
	if err := durability.SyncDirectory(path); err == nil {
		t.Fatal("SyncDirectory() error = nil, want missing directory error")
	}
}
