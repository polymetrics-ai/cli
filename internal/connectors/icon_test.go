package connectors

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConnectorIconRegistryUsesExactBareNamesOnly(t *testing.T) {
	apify, ok := ConnectorIconFor("apify-dataset")
	if !ok {
		t.Fatal("apify-dataset icon not found")
	}
	if apify.Path != "icons/apify.svg" || apify.Source != IconSourceUpstream {
		t.Fatalf("apify-dataset icon = %+v, want upstream apify.svg", apify)
	}

	apple, ok := ConnectorIconFor("apple-search-ads")
	if !ok {
		t.Fatal("apple-search-ads icon not found")
	}
	if apple.Path != "icons/simple-icons/apple.svg" || apple.Source != IconSourceSimpleIcons {
		t.Fatalf("apple-search-ads icon = %+v, want curated Simple Icons apple.svg", apple)
	}

	if icon, ok := ConnectorIconFor("source-apify-dataset"); ok {
		t.Fatalf("legacy prefixed lookup unexpectedly resolved: %+v", icon)
	}
}

func TestConnectorIconRegistryContainsOnlyBareKeys(t *testing.T) {
	for _, entry := range ConnectorIconEntries() {
		if strings.HasPrefix(entry.Connector, "source-") || strings.HasPrefix(entry.Connector, "destination-") {
			t.Fatalf("connector icon registry contains legacy prefixed key %q", entry.Connector)
		}
	}
}

func TestConnectorIconPathOwnershipMapsCanonicalAndGeneratedCopies(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{path: "docs/connectors/icons/apify.svg", want: "apify-dataset"},
		{path: "website/public/connectors/icons/apify.svg", want: "apify-dataset"},
		{path: "docs/connectors/icons/simple-icons/apple.svg", want: "apple-search-ads"},
		{path: "website/public/connectors/icons/simple-icons/apple.svg", want: "apple-search-ads"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			got, err := ConnectorIconOwnerForPath(tc.path)
			if err != nil {
				t.Fatalf("ConnectorIconOwnerForPath(%q): %v", tc.path, err)
			}
			if got != tc.want {
				t.Fatalf("ConnectorIconOwnerForPath(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func TestConnectorIconPathOwnershipRejectsAmbiguousOrUndeclaredPaths(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
		want string
	}{
		{name: "ambiguous shared fallback", path: "docs/connectors/icons/pm-sample.svg", want: "ambiguous connector icon path"},
		{name: "undeclared canonical asset", path: "docs/connectors/icons/not-declared.svg", want: "undeclared connector icon path"},
		{name: "outside icon roots", path: "docs/connectors/README.md", want: "unsupported connector icon path"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ConnectorIconOwnerForPath(tc.path)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ConnectorIconOwnerForPath(%q) error = %v, want %q", tc.path, err, tc.want)
			}
		})
	}

	_, err := ValidateConnectorIconOwnershipPaths([]string{
		"docs/connectors/icons/apify.svg",
		"website/public/connectors/icons/apify.svg",
		"docs/connectors/icons/apify.svg",
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate connector icon path") {
		t.Fatalf("ValidateConnectorIconOwnershipPaths duplicate error = %v", err)
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
	err := ValidateConnectorIcons(t.TempDir(), []Definition{{Name: "missing", DisplayName: "Missing"}}, nil)
	if err == nil || !strings.Contains(err.Error(), "connector icon missing: missing icon registry entry") {
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
	defs := []Definition{{
		Name: "unsafe",
		Icon: &ConnectorIcon{ID: "unsafe", Path: "icons/unsafe.svg", Source: IconSourceUpstream, ReviewStatus: IconReviewUpstreamSeeded},
	}}
	err := ValidateConnectorIcons(dir, defs, nil)
	if err == nil || !strings.Contains(err.Error(), `connector icon unsafe: svg contains forbidden content "<script"`) {
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
			err := ValidateConnectorIconSVGContent("unsafe-test", []byte(tc.svg))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ValidateConnectorIconSVGContent() error = %v, want %q", err, tc.want)
			}
		})
	}
}
