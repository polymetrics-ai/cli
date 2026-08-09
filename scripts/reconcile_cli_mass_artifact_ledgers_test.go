package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStageValidatedJSONBytesLeavesDestinationUntouchedForInvalidCandidate(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "ledger.json")
	const original = `{"preserve":"evidence"}` + "\n"
	if err := os.WriteFile(destination, []byte(original), 0o644); err != nil {
		t.Fatalf("seed destination: %v", err)
	}

	if _, err := stageValidatedJSONBytes(destination, []byte(`{"broken":`)); err == nil {
		t.Fatal("stage invalid JSON error = nil, want refusal before replacement")
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read destination after invalid candidate: %v", err)
	}
	if string(got) != original {
		t.Fatalf("destination = %q after invalid candidate, want unchanged %q", got, original)
	}
}
