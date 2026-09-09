package agentcontract

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCodeProjectionsAreRequiredAndGenerated(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	updated, err := SyncProjections(root, contract)
	if err != nil || updated != 8 {
		t.Fatalf("SyncProjections created %d projections: %v", updated, err)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("generated projections failed drift check: %v", err)
	}
	count := 0
	for _, target := range contract.Projections {
		if target.Harness != "opencode" {
			continue
		}
		count++
		if !target.Required || target.RenderMode != opencodeMarkdownYAMLFrontmatter {
			t.Fatalf("OpenCode target %#v is not a required canonical projection", target)
		}
	}
	if count != 2 {
		t.Fatalf("canonical contract registers %d OpenCode projections, want 2", count)
	}
	missing := filepath.Join(root, ".opencode", "agents", "pm-delivery-worker.md")
	if err := os.Remove(missing); err != nil {
		t.Fatal(err)
	}
	if err := CheckProjections(root, contract); err == nil {
		t.Fatal("CheckProjections accepted a missing registered OpenCode projection")
	}
}
