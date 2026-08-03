package connectors

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type missingIconConnector struct{ Sample }

func (missingIconConnector) Name() string { return "missing-icon" }

func (missingIconConnector) Metadata() Metadata {
	return Metadata{Name: "missing-icon", DisplayName: "Missing Icon"}
}

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

func TestConnectorIconRegistryProjectsCompleteMetadata(t *testing.T) {
	apple, ok := ConnectorIconFor("apple-search-ads")
	if !ok {
		t.Fatal("apple-search-ads icon not found")
	}
	want := ConnectorIcon{
		ID:             "simple-icons-apple",
		Path:           "icons/simple-icons/apple.svg",
		Title:          "Apple",
		SimpleIconSlug: "apple",
		SimpleIconHex:  "000000",
		Source:         IconSourceSimpleIcons,
		License:        "CC0-1.0",
		ReviewStatus:   IconReviewSimpleIconsCC0Trademark,
		ReviewURL:      "https://simpleicons.org/?q=Apple",
		Match:          "curated-alias",
		MatchedBy:      "apple",
	}
	if apple != want {
		t.Fatalf("apple-search-ads icon = %+v, want %+v", apple, want)
	}

	section := iconSection(Manifest{Metadata: Metadata{Icon: &ConnectorIcon{
		ID:             "projected",
		Path:           "icons/projected.svg",
		Title:          "Projected Icon",
		SimpleIconSlug: "projected-icon",
		SimpleIconHex:  "ABCDEF",
		Source:         IconSourceSimpleIcons,
		License:        "CC0-1.0",
		Attribution:    "Example attribution",
		ReviewStatus:   IconReviewSimpleIconsCC0Trademark,
		ReviewURL:      "https://example.invalid/review",
		Match:          "exact-name-or-slug",
		MatchedBy:      "projected-icon",
	}}})
	wantLines := []string{
		"id: projected",
		"asset: icons/projected.svg",
		"title: Projected Icon",
		"simple_icon_slug: projected-icon",
		"simple_icon_hex: ABCDEF",
		"source: simple-icons",
		"license: CC0-1.0",
		"attribution: Example attribution",
		"review_status: cc0_with_trademark_caveat",
		"review_url: https://example.invalid/review",
		"match: exact-name-or-slug",
		"matched_by: projected-icon",
	}
	if strings.Join(section.Lines, "\n") != strings.Join(wantLines, "\n") {
		t.Fatalf("icon section = %q, want %q", section.Lines, wantLines)
	}
}

func TestConnectorIconMetadataOmitsAbsentOptionalFields(t *testing.T) {
	icon, ok := ConnectorIconFor("100ms")
	if !ok {
		t.Fatal("100ms icon not found")
	}
	section := iconSection(Manifest{Metadata: Metadata{Icon: &icon}})
	want := []string{
		"id: pm-sample",
		"asset: icons/pm-sample.svg",
		"source: polymetrics",
		"review_status: polymetrics",
		"review_url: https://github.com/polymetrics-ai/cli",
	}
	if strings.Join(section.Lines, "\n") != strings.Join(want, "\n") {
		t.Fatalf("100ms icon section = %q, want %q", section.Lines, want)
	}
	encoded, err := json.Marshal(icon)
	if err != nil {
		t.Fatalf("marshal 100ms icon: %v", err)
	}
	wantJSON := `{"id":"pm-sample","path":"icons/pm-sample.svg","source":"polymetrics","review_status":"polymetrics","review_url":"https://github.com/polymetrics-ai/cli"}`
	if string(encoded) != wantJSON {
		t.Fatalf("100ms icon JSON = %s, want %s", encoded, wantJSON)
	}
}

func TestConnectorIconRegistryContainsOnlyBareKeys(t *testing.T) {
	for _, entry := range ConnectorIconEntries() {
		if strings.HasPrefix(entry.Connector, "source-") || strings.HasPrefix(entry.Connector, "destination-") {
			t.Fatalf("connector icon registry contains legacy prefixed key %q", entry.Connector)
		}
	}
}

func TestMetadataWithIconUsesCanonicalRegistryOnly(t *testing.T) {
	meta := MetadataWithIcon(Metadata{
		Name: "missing-icon",
		Icon: &ConnectorIcon{
			ID:           "pm-missing-icon",
			Path:         "icons/pm-sample.svg",
			Source:       IconSourcePolymetrics,
			ReviewStatus: IconReviewPolymetrics,
		},
	})
	if meta.Icon != nil {
		t.Fatalf("MetadataWithIcon() used non-canonical metadata icon: %+v", meta.Icon)
	}
}

func TestRegistryIconCoverageRequiresExplicitCanonicalEntry(t *testing.T) {
	registry := NewEmptyRegistry()
	registry.Register(missingIconConnector{})
	err := registry.ValidateIconCoverage()
	if err == nil || !strings.Contains(err.Error(), `missing explicit icon registry entry for connector "missing-icon"`) {
		t.Fatalf("ValidateIconCoverage() error = %v", err)
	}
}

func TestMustValidateIconCoverageRevalidatesAfterRegistration(t *testing.T) {
	registry := NewEmptyRegistry()
	registry.RegisterBuiltins()
	registry.MustValidateIconCoverage()
	registry.MustValidateIconCoverage()

	defer func() {
		if recover() == nil {
			t.Fatal("MustValidateIconCoverage() did not panic for a connector registered after validation")
		}
	}()
	registry.Register(missingIconConnector{})
	registry.MustValidateIconCoverage()
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

func TestConnectorIconPathOwnershipReportsRuntimeBuiltinDisposition(t *testing.T) {
	for _, path := range []string{
		"docs/connectors/icons/pm-file.svg",
		"website/public/connectors/icons/pm-file.svg",
		"docs/connectors/icons/pm-warehouse.svg",
	} {
		t.Run(path, func(t *testing.T) {
			owner, err := ConnectorIconOwnerForPath(path)
			if !errors.Is(err, ErrConnectorIconPathRuntimeBuiltin) {
				t.Fatalf("ConnectorIconOwnerForPath(%q) error = %v, want runtime builtin disposition", path, err)
			}
			if owner != "" {
				t.Fatalf("ConnectorIconOwnerForPath(%q) owner = %q, want no connector ownership", path, owner)
			}
			if strings.Contains(err.Error(), "undeclared") {
				t.Fatalf("ConnectorIconOwnerForPath(%q) error = %v, want a declared-but-unowned disposition", path, err)
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
