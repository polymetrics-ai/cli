package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFoundationEvidenceHistoricalManifestRequiresExactGraph(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	manifest := readEvidenceFixture(t, filepath.Join(root, "data", "cli-current-foundations-main-integration-r1", "evidence-manifest.json"))
	if err := validateFoundationEvidence(manifest, nil, nil); err == nil || !strings.Contains(err.Error(), "exact repository graph") {
		t.Fatalf("historical evidence error = %v, want exact graph refusal", err)
	}
}

func readEvidenceFixture(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
