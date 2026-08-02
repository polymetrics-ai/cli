package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateConnectorDocsRejectsStaleIconMetadata(t *testing.T) {
	dir := t.TempDir()
	registry := appRegistry()
	if err := writeConnectorDocs(dir, registry); err != nil {
		t.Fatalf("write connector docs: %v", err)
	}

	tests := []struct {
		name string
		path string
		from string
		to   string
	}{
		{
			name: "catalog json",
			path: filepath.Join(dir, "catalog", "all-connectors.json"),
			from: `"path": "icons/simple-icons/apple.svg"`,
			to:   `"path": "icons/pm-sample.svg"`,
		},
		{
			name: "catalog markdown",
			path: filepath.Join(dir, "catalog", "all-connectors.md"),
			from: "icons/simple-icons/apple.svg",
			to:   "icons/pm-sample.svg",
		},
		{
			name: "manual",
			path: filepath.Join(dir, "apple-search-ads", "MANUAL.md"),
			from: "ICON\n  asset: icons/simple-icons/apple.svg\n",
			to:   "ICON\n  asset: icons/pm-sample.svg\n",
		},
		{
			name: "skill",
			path: filepath.Join(dir, "apple-search-ads", "SKILL.md"),
			from: "## Icon\n\n- asset: icons/simple-icons/apple.svg\n",
			to:   "## Icon\n\n- asset: icons/pm-sample.svg\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original, err := os.ReadFile(test.path)
			if err != nil {
				t.Fatalf("read generated doc: %v", err)
			}
			if !strings.Contains(string(original), test.from) {
				t.Fatalf("generated doc missing fixture %q", test.from)
			}
			corrupted := strings.ReplaceAll(string(original), test.from, test.to)
			if err := os.WriteFile(test.path, []byte(corrupted), 0o644); err != nil {
				t.Fatalf("write stale generated doc: %v", err)
			}
			defer func() {
				if err := os.WriteFile(test.path, original, 0o644); err != nil {
					t.Errorf("restore generated doc: %v", err)
				}
			}()

			err = validateConnectorDocs(dir, registry)
			if err == nil {
				t.Fatal("validateConnectorDocs accepted stale icon metadata")
			}
			if !strings.Contains(err.Error(), "icon path") {
				t.Fatalf("validation error = %q, want icon path mismatch", err)
			}
		})
	}
}
