package main

import (
	"path/filepath"
	"testing"
)

func TestRecurlyReviewBundleValidatesAndSurfaceIsSynchronized(t *testing.T) {
	dir := filepath.Join("..", "..", "internal", "connectors", "defs", "recurly")
	report, err := validatePath(dir)
	if err != nil {
		t.Fatalf("validatePath: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("Recurly validation findings: %+v", report.Findings)
	}

	stats, err := syncBundle(dir, true)
	if err != nil {
		t.Fatalf("syncBundle: %v", err)
	}
	if stats.total() != 0 {
		t.Fatalf("Recurly surface has %d derived metadata changes", stats.total())
	}
}
