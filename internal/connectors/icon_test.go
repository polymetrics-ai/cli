package connectors

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConnectorIconRegistryResolvesBareNames(t *testing.T) {
	github, ok := ConnectorIconFor("github")
	if !ok {
		t.Fatal("github icon not found")
	}
	if github.Path != "icons/github.svg" || github.Source != IconSourceUpstream {
		t.Fatalf("github icon = %+v, want upstream github.svg", github)
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
