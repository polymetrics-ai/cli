package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestValidateConnectorDocsRejectsStaleIconMetadata(t *testing.T) {
	dir := t.TempDir()
	registry := appRegistry()
	if err := writeConnectorDocs(dir, registry); err != nil {
		t.Fatalf("write connector docs: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		corrupt func(*testing.T, []byte) []byte
		want    string
	}{
		{
			name: "catalog json",
			path: filepath.Join(dir, "catalog", "all-connectors.json"),
			corrupt: corruptCatalogIcon("apple-search-ads", func(icon *connectors.ConnectorIcon) {
				icon.Path = "icons/pm-sample.svg"
			}),
			want: "icon metadata",
		},
		{
			name: "catalog json missing canonical review URL",
			path: filepath.Join(dir, "catalog", "all-connectors.json"),
			corrupt: corruptCatalogIcon("100ms", func(icon *connectors.ConnectorIcon) {
				icon.ReviewURL = ""
			}),
			want: "icon metadata",
		},
		{
			name: "catalog json obsolete same-path provenance",
			path: filepath.Join(dir, "catalog", "all-connectors.json"),
			corrupt: corruptCatalogIcon("convex", func(icon *connectors.ConnectorIcon) {
				icon.Source = connectors.IconSourceUpstream
				icon.ReviewStatus = connectors.IconReviewUpstreamSeeded
				icon.ReviewURL = "https://docs.convex.dev/http-api/"
			}),
			want: "icon metadata",
		},
		{
			name: "catalog markdown",
			path: filepath.Join(dir, "catalog", "all-connectors.md"),
			corrupt: replaceGeneratedMetadata(
				"icons/simple-icons/apple.svg",
				"icons/pm-sample.svg",
			),
			want: "icon path",
		},
		{
			name: "manual",
			path: filepath.Join(dir, "apple-search-ads", "MANUAL.md"),
			corrupt: replaceGeneratedMetadata(
				"ICON\n  asset: icons/simple-icons/apple.svg\n",
				"ICON\n  asset: icons/pm-sample.svg\n",
			),
			want: "icon metadata",
		},
		{
			name: "manual missing canonical review URL",
			path: filepath.Join(dir, "100ms", "MANUAL.md"),
			corrupt: replaceGeneratedMetadata(
				"  review_url: https://github.com/polymetrics-ai/cli\n",
				"",
			),
			want: "icon metadata",
		},
		{
			name: "manual obsolete same-path provenance",
			path: filepath.Join(dir, "convex", "MANUAL.md"),
			corrupt: replaceGeneratedMetadata(
				"  source: official\n  review_status: official_verified\n  review_url: https://docs.convex.dev/\n",
				"  source: upstream_registry\n  review_status: upstream_seeded\n  review_url: https://docs.convex.dev/http-api/\n",
			),
			want: "icon metadata",
		},
		{
			name: "skill",
			path: filepath.Join(dir, "apple-search-ads", "SKILL.md"),
			corrupt: replaceGeneratedMetadata(
				"## Icon\n\n- asset: icons/simple-icons/apple.svg\n",
				"## Icon\n\n- asset: icons/pm-sample.svg\n",
			),
			want: "icon metadata",
		},
		{
			name: "skill missing canonical review URL",
			path: filepath.Join(dir, "100ms", "SKILL.md"),
			corrupt: replaceGeneratedMetadata(
				"- review_url: https://github.com/polymetrics-ai/cli\n",
				"",
			),
			want: "icon metadata",
		},
		{
			name: "skill obsolete same-path provenance",
			path: filepath.Join(dir, "convex", "SKILL.md"),
			corrupt: replaceGeneratedMetadata(
				"- source: official\n- review_status: official_verified\n- review_url: https://docs.convex.dev/\n",
				"- source: upstream_registry\n- review_status: upstream_seeded\n- review_url: https://docs.convex.dev/http-api/\n",
			),
			want: "icon metadata",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original, err := os.ReadFile(test.path)
			if err != nil {
				t.Fatalf("read generated doc: %v", err)
			}
			corrupted := test.corrupt(t, original)
			if err := os.WriteFile(test.path, corrupted, 0o644); err != nil {
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
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validation error = %q, want %q", err, test.want)
			}
		})
	}
}

func corruptCatalogIcon(name string, corrupt func(*connectors.ConnectorIcon)) func(*testing.T, []byte) []byte {
	return func(t *testing.T, data []byte) []byte {
		t.Helper()
		var defs []connectors.Definition
		if err := json.Unmarshal(data, &defs); err != nil {
			t.Fatalf("decode generated catalog: %v", err)
		}
		found := false
		for i := range defs {
			if defs[i].Name != name {
				continue
			}
			if defs[i].Icon == nil {
				t.Fatalf("connector %s generated without icon", name)
			}
			corrupt(defs[i].Icon)
			found = true
			break
		}
		if !found {
			t.Fatalf("generated catalog missing connector %s", name)
		}
		corrupted, err := json.MarshalIndent(defs, "", "  ")
		if err != nil {
			t.Fatalf("encode corrupted catalog: %v", err)
		}
		return append(corrupted, '\n')
	}
}

func replaceGeneratedMetadata(from, to string) func(*testing.T, []byte) []byte {
	return func(t *testing.T, data []byte) []byte {
		t.Helper()
		if !strings.Contains(string(data), from) {
			t.Fatalf("generated doc missing fixture %q", from)
		}
		return []byte(strings.ReplaceAll(string(data), from, to))
	}
}
