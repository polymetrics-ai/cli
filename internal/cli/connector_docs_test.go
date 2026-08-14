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
			name: "catalog json icon ID",
			path: filepath.Join(dir, "catalog", "all-connectors.json"),
			corrupt: corruptCatalogIcon("apple-search-ads", func(icon *connectors.ConnectorIcon) {
				icon.ID = "stale-apple"
			}),
			want: "icon metadata",
		},
		{
			name: "catalog json license attribution",
			path: filepath.Join(dir, "catalog", "all-connectors.json"),
			corrupt: corruptCatalogIcon("apple-search-ads", func(icon *connectors.ConnectorIcon) {
				icon.License = "stale-license"
			}),
			want: "icon metadata",
		},
		{
			name: "catalog json match metadata",
			path: filepath.Join(dir, "catalog", "all-connectors.json"),
			corrupt: corruptCatalogIcon("apple-search-ads", func(icon *connectors.ConnectorIcon) {
				icon.MatchedBy = "stale-match"
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
				"  asset: icons/simple-icons/apple.svg\n",
				"  asset: icons/pm-sample.svg\n",
			),
			want: "icon metadata",
		},
		{
			name: "manual icon ID",
			path: filepath.Join(dir, "apple-search-ads", "MANUAL.md"),
			corrupt: replaceGeneratedMetadata(
				"  id: simple-icons-apple\n",
				"  id: stale-apple\n",
			),
			want: "icon metadata",
		},
		{
			name: "manual license attribution",
			path: filepath.Join(dir, "apple-search-ads", "MANUAL.md"),
			corrupt: replaceGeneratedMetadata(
				"  license: CC0-1.0\n",
				"  license: stale-license\n",
			),
			want: "icon metadata",
		},
		{
			name: "manual match metadata",
			path: filepath.Join(dir, "apple-search-ads", "MANUAL.md"),
			corrupt: replaceGeneratedMetadata(
				"  matched_by: apple\n",
				"  matched_by: stale-match\n",
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
				"- asset: icons/simple-icons/apple.svg\n",
				"- asset: icons/pm-sample.svg\n",
			),
			want: "icon metadata",
		},
		{
			name: "skill icon ID",
			path: filepath.Join(dir, "apple-search-ads", "SKILL.md"),
			corrupt: replaceGeneratedMetadata(
				"- id: simple-icons-apple\n",
				"- id: stale-apple\n",
			),
			want: "icon metadata",
		},
		{
			name: "skill license attribution",
			path: filepath.Join(dir, "apple-search-ads", "SKILL.md"),
			corrupt: replaceGeneratedMetadata(
				"- license: CC0-1.0\n",
				"- license: stale-license\n",
			),
			want: "icon metadata",
		},
		{
			name: "skill match metadata",
			path: filepath.Join(dir, "apple-search-ads", "SKILL.md"),
			corrupt: replaceGeneratedMetadata(
				"- matched_by: apple\n",
				"- matched_by: stale-match\n",
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

func TestValidateConnectorDocsRejectsStaleGeneratedContent(t *testing.T) {
	dir := t.TempDir()
	registry := appRegistry()
	if err := writeConnectorDocs(dir, registry); err != nil {
		t.Fatalf("write connector docs: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read connector docs: %v", err)
	}
	manualPath := ""
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(dir, entry.Name(), "MANUAL.md")
		if _, err := os.Stat(candidate); err == nil {
			manualPath = candidate
			break
		}
	}
	if manualPath == "" {
		t.Fatal("generated connector docs contain no manual")
	}
	manual, err := os.ReadFile(manualPath)
	if err != nil {
		t.Fatalf("read generated manual: %v", err)
	}
	if err := os.WriteFile(manualPath, append(manual, []byte("\nstale\n")...), 0o644); err != nil {
		t.Fatalf("write stale generated manual: %v", err)
	}

	err = validateConnectorDocs(dir, registry)
	if err == nil {
		t.Fatal("validateConnectorDocs accepted stale generated content")
	}
	if !strings.Contains(err.Error(), "manual is stale") {
		t.Fatalf("validation error = %q, want stale manual error", err)
	}
}

func TestGeneratedConnectorIconBlockRequiresExactUniqueHeading(t *testing.T) {
	tests := []struct {
		name     string
		document string
		heading  string
		want     string
		wantErr  string
	}{
		{
			name:     "valid exact manual section",
			document: "DESCRIPTION\nConnector details.\n\nICON\n  id: canonical\n\nSECURITY\nSafe.\n",
			heading:  "ICON\n",
			want:     "  id: canonical",
		},
		{
			name:     "manual ignores favicon substring",
			document: "DESCRIPTION\nFAVICON\n  id: shadow\n\nICON\n  id: canonical\n\nSECURITY\nSafe.\n",
			heading:  "ICON\n",
			want:     "  id: canonical",
		},
		{
			name:     "skill ignores embedded heading fragment",
			document: "## Description\n\nFAV## Icon\n\n- id: shadow\n\n## Icon\n\n- id: canonical\n\n## Agent Rules\n",
			heading:  "## Icon\n\n",
			want:     "- id: canonical",
		},
		{
			name:     "duplicate exact sections",
			document: "ICON\n  id: first\n\nICON\n  id: second\n\nSECURITY\nSafe.\n",
			heading:  "ICON\n",
			wantErr:  "duplicate sections",
		},
		{
			name:     "skill rejects duplicate heading without blank line",
			document: "## Icon\n- id: shadow\n\n## Icon\n\n- id: canonical\n\n## Agent Rules\n",
			heading:  "## Icon\n\n",
			wantErr:  "duplicate sections",
		},
		{
			name:     "missing exact section",
			document: "DESCRIPTION\nFAVICON\n  id: shadow\n\nSECURITY\nSafe.\n",
			heading:  "ICON\n",
			wantErr:  "missing section",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := generatedConnectorIconBlock(test.document, test.heading)
			if test.wantErr != "" {
				if err == nil {
					t.Fatalf("generatedConnectorIconBlock() error = nil, want %q", test.wantErr)
				}
				if !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("generatedConnectorIconBlock() error = %q, want %q", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("generatedConnectorIconBlock() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("generatedConnectorIconBlock() = %q, want %q", got, test.want)
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
