package defs

import (
	"io/fs"
	"path"
	"strings"
	"testing"
)

func TestRuntimeEmbedContainsExecutionJSONOnly(t *testing.T) {
	forbiddenBase := map[string]bool{
		"api_surface.json":                   true,
		"certification.json":                 true,
		"declaration_admission_sources.json": true,
		"docs.md":                            true,
		"enabled_connector_contract.json":    true,
		"operation_endpoint_ledger.json":     true,
		"source.lock.json":                   true,
	}
	err := fs.WalkDir(FS, ".", func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		base := path.Base(filePath)
		topLevelConnectorArtifact := strings.Count(filePath, "/") == 1
		if (topLevelConnectorArtifact && forbiddenBase[base]) || strings.Contains(filePath, "/sources/") || strings.Contains(filePath, "source-lock") {
			t.Errorf("runtime embed contains authoring/admission artifact %q", filePath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk runtime embed: %v", err)
	}
}
