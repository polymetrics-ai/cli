package connectors

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConnectorCatalogIncludesIconMetadata(t *testing.T) {
	catalog := ConnectorCatalog()
	if got, want := len(catalog), 646; got != want {
		t.Fatalf("catalog len = %d, want %d", got, want)
	}
	for _, entry := range catalog {
		if entry.Icon == nil {
			t.Fatalf("%s missing icon metadata", entry.Slug)
		}
		if entry.Icon.ID == "" || entry.Icon.Path == "" || entry.Icon.Source == "" || entry.Icon.ReviewStatus == "" {
			t.Fatalf("%s incomplete icon metadata: %+v", entry.Slug, entry.Icon)
		}
		if !strings.HasPrefix(entry.Icon.Path, "icons/") || !strings.HasSuffix(entry.Icon.Path, ".svg") {
			t.Fatalf("%s icon path = %q, want icons/*.svg", entry.Slug, entry.Icon.Path)
		}
	}

	github, ok := ConnectorDefinitionBySlug("source-github")
	if !ok {
		t.Fatal("source-github not found")
	}
	if github.Icon == nil || github.Icon.Path != "icons/github.svg" || github.Icon.Source != IconSourceUpstream {
		t.Fatalf("source-github icon = %+v, want upstream github.svg", github.Icon)
	}
}

func TestRegistryListIncludesBuiltinIcons(t *testing.T) {
	registry := NewRegistry()
	want := map[string]bool{"sample": true, "file": true, "warehouse": true, "outbox": true}
	for _, meta := range registry.List() {
		if !want[meta.Name] {
			continue
		}
		delete(want, meta.Name)
		if meta.Icon == nil || meta.Icon.Path == "" {
			t.Fatalf("%s missing built-in icon: %+v", meta.Name, meta.Icon)
		}
		if meta.Icon.Source != "polymetrics" || meta.Icon.ReviewStatus != "polymetrics" {
			t.Fatalf("%s icon source = %+v, want polymetrics", meta.Name, meta.Icon)
		}
	}
	if len(want) > 0 {
		t.Fatalf("built-in connectors not listed: %+v", want)
	}
}

func TestValidateConnectorIconsReportsMissingMetadata(t *testing.T) {
	err := ValidateConnectorIcons(t.TempDir(), []ConnectorDefinition{{Slug: "source-missing", Name: "Missing", Type: ConnectorTypeSource}}, nil)
	if err == nil || !strings.Contains(err.Error(), "connector icon source-missing: missing icon registry entry") {
		t.Fatalf("ValidateConnectorIcons() error = %v", err)
	}
}

func TestValidateConnectorIconsRejectsUnsafeSVG(t *testing.T) {
	dir := t.TempDir()
	iconsDir := filepath.Join(dir, "icons")
	if err := os.MkdirAll(iconsDir, 0o755); err != nil {
		t.Fatalf("mkdir icons: %v", err)
	}
	if err := os.WriteFile(filepath.Join(iconsDir, "unsafe.svg"), []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`), 0o644); err != nil {
		t.Fatalf("write unsafe icon: %v", err)
	}
	defs := []ConnectorDefinition{{
		Slug: "source-unsafe",
		Name: "Unsafe",
		Type: ConnectorTypeSource,
		Icon: &ConnectorIcon{ID: "unsafe", Path: "icons/unsafe.svg", Source: IconSourceUpstream, ReviewStatus: IconReviewUpstreamSeeded},
	}}
	err := ValidateConnectorIcons(dir, defs, nil)
	if err == nil || !strings.Contains(err.Error(), `connector icon source-unsafe: svg contains forbidden content "<script"`) {
		t.Fatalf("ValidateConnectorIcons() error = %v", err)
	}
}

func TestValidateConnectorIconSVGContentRejectsEventHandlersAndExternalReferences(t *testing.T) {
	cases := []struct {
		name string
		svg  string
		want string
	}{
		{name: "event handler", svg: `<svg xmlns="http://www.w3.org/2000/svg"><path onload = "alert(1)" d="M0 0"/></svg>`, want: "forbidden event handler"},
		{name: "href", svg: `<svg xmlns="http://www.w3.org/2000/svg"><image href = "https://example.com/icon.svg"/></svg>`, want: "forbidden external href"},
		{name: "url", svg: `<svg xmlns="http://www.w3.org/2000/svg"><path style="fill: url( https://example.com/a.svg )"/></svg>`, want: "forbidden external url()"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateConnectorIconSVGContent("source-test", []byte(tc.svg))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ValidateConnectorIconSVGContent() error = %v, want %q", err, tc.want)
			}
		})
	}
}
