package engine

import (
	"testing"

	"polymetrics.ai/internal/connectors/defs"
)

func TestXeroOperationsLedgerMetrics(t *testing.T) {
	bundle, err := Load(defs.FS, "xero")
	if err != nil {
		t.Fatalf("load xero bundle: %v", err)
	}

	byKind := map[string]int{}
	for _, op := range bundle.Operations {
		byKind[op.Kind]++
	}
	if got, want := len(bundle.Operations), 92; got != want {
		t.Fatalf("Xero operations = %d, want %d", got, want)
	}
	if got, want := byKind["binary_download"], 26; got != want {
		t.Fatalf("Xero binary_download operations = %d, want %d", got, want)
	}
	if got, want := byKind["rest_read"], 22; got != want {
		t.Fatalf("Xero rest_read operations = %d, want %d", got, want)
	}
	if got, want := byKind["file_upload"], 22; got != want {
		t.Fatalf("Xero file_upload operations = %d, want %d", got, want)
	}
	if got, want := byKind["rest_write"], 22; got != want {
		t.Fatalf("Xero rest_write operations = %d, want %d", got, want)
	}
}
