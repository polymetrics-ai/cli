package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestBuildIconEntryUsesUpstreamIconMetadata(t *testing.T) {
	entry, ok, err := buildIconEntry(map[string]any{
		"public":           true,
		"dockerRepository": "registry/source-github",
		"documentationUrl": "https://example.com/integrations/sources/github",
		"icon":             "github.svg",
		"iconUrl":          "https://example.com/source-github/icon.svg",
		"externalDocumentationUrls": []any{map[string]any{
			"url": "https://docs.github.com/rest",
		}},
	})
	if err != nil || !ok {
		t.Fatalf("buildIconEntry() ok=%t err=%v", ok, err)
	}
	if entry.Connector != "github" || entry.ID != "github" || entry.Path != "icons/github.svg" {
		t.Fatalf("entry identity = %+v", entry)
	}
	if entry.Source != "upstream_registry" || entry.ReviewStatus != "upstream_seeded" || entry.ReviewURL != "" {
		t.Fatalf("entry provenance = %+v", entry)
	}
}

func TestBuildIconEntryScopesGenericIconNames(t *testing.T) {
	entry, ok, err := buildIconEntry(map[string]any{
		"public":           true,
		"dockerRepository": "registry/source-demo",
		"documentationUrl": "https://example.com/integrations/sources/demo",
		"icon":             "icon.svg",
		"iconUrl":          "https://example.com/source-demo/icon.svg",
	})
	if err != nil || !ok {
		t.Fatalf("buildIconEntry() ok=%t err=%v", ok, err)
	}
	if entry.ID != "demo" || entry.Path != "icons/demo.svg" {
		t.Fatalf("entry identity = %+v", entry)
	}
}

func TestBuildIconEntriesAddsReviewedFallbackForImplementedDefinitions(t *testing.T) {
	entries, assets, err := buildIconEntries(registryFile{Sources: []map[string]any{{
		"public":           true,
		"dockerRepository": "registry/source-demo",
		"documentationUrl": "https://example.com/integrations/sources/demo",
		"icon":             "demo.svg",
		"iconUrl":          "https://example.com/source-demo/icon.svg",
	}}}, buildOptions{ImplementedConnectors: map[string]bool{"demo": true, "missing": true}})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries len = %d, want 2: %+v", len(entries), entries)
	}
	if len(assets) != 1 || assets[0].Path != "icons/demo.svg" {
		t.Fatalf("assets = %+v", assets)
	}
	byConnector := map[string]iconEntry{}
	for _, entry := range entries {
		byConnector[entry.Connector] = entry
		if strings.HasPrefix(entry.Connector, "source-") || strings.HasPrefix(entry.Connector, "destination-") {
			t.Fatalf("buildIconEntries emitted prefixed key: %+v", entry)
		}
	}
	if got := byConnector["demo"]; got.Path != "icons/demo.svg" || got.Source != connectors.IconSourceUpstream {
		t.Fatalf("demo entry = %+v", got)
	}
	if got := byConnector["missing"]; got.Path != "icons/pm-sample.svg" || got.Source != connectors.IconSourcePolymetrics {
		t.Fatalf("missing fallback entry = %+v", got)
	}
}

func TestBuildIconEntriesPreservesCuratedAttribution(t *testing.T) {
	entries, _, err := buildIconEntries(registryFile{}, buildOptions{
		ImplementedConnectors: map[string]bool{"demo": true},
		CuratedEntries: []iconEntry{{
			Connector:      "demo",
			ID:             "simple-icons-demo",
			Path:           "icons/simple-icons/demo.svg",
			Source:         connectors.IconSourceSimpleIcons,
			License:        "CC0-1.0",
			Attribution:    "Example attribution",
			ReviewStatus:   connectors.IconReviewSimpleIconsCC0Trademark,
			SimpleIconSlug: "demo",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Attribution != "Example attribution" {
		t.Fatalf("entries = %+v, want curated attribution", entries)
	}
}

func TestBuildIconEntriesPreservesCuratedRuntimeBuiltinOverride(t *testing.T) {
	entries, _, err := buildIconEntries(registryFile{}, buildOptions{
		ImplementedConnectors: map[string]bool{"demo": true},
		IncludeLocalBuiltins:  true,
		CuratedEntries: []iconEntry{{
			Connector:    "warehouse",
			ID:           "warehouse",
			Path:         "icons/warehouse.svg",
			Source:       connectors.IconSourceOfficial,
			ReviewStatus: connectors.IconReviewOfficial,
			ReviewURL:    "https://example.com/warehouse",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var got iconEntry
	for _, entry := range entries {
		if entry.Connector == "warehouse" {
			got = entry
		}
	}
	if got.Path != "icons/warehouse.svg" || got.Source != connectors.IconSourceOfficial || got.ReviewStatus != connectors.IconReviewOfficial {
		t.Fatalf("warehouse entry = %+v, want curated override retained", got)
	}
}

func TestBuildIconEntriesRejectsCuratedKeyWithoutOwner(t *testing.T) {
	_, _, err := buildIconEntries(registryFile{}, buildOptions{
		ImplementedConnectors: map[string]bool{"demo": true},
		IncludeLocalBuiltins:  true,
		CuratedEntries: []iconEntry{{
			Connector:    "retired",
			ID:           "retired",
			Path:         "icons/retired.svg",
			Source:       connectors.IconSourceOfficial,
			ReviewStatus: connectors.IconReviewOfficial,
		}},
	})
	if err == nil || !strings.Contains(err.Error(), `curated icon entry "retired" has no connector definition or runtime builtin owner`) {
		t.Fatalf("buildIconEntries curated owner error = %v", err)
	}
}

func TestBuildIconEntriesAllowsSharedPathWhenCuratedEntryHasNoSourceURL(t *testing.T) {
	entries, assets, err := buildIconEntries(registryFile{Sources: []map[string]any{
		{
			"public":           true,
			"dockerRepository": "registry/source-alpha",
			"documentationUrl": "https://example.com/integrations/sources/alpha",
			"icon":             "shared.svg",
			"iconUrl":          "https://example.com/shared.svg",
		},
		{
			"public":           true,
			"dockerRepository": "registry/source-beta",
			"documentationUrl": "https://example.com/integrations/sources/beta",
			"icon":             "shared.svg",
			"iconUrl":          "https://example.com/shared.svg",
		},
	}}, buildOptions{
		ImplementedConnectors: map[string]bool{"alpha": true, "beta": true},
		CuratedEntries: []iconEntry{{
			Connector:    "alpha",
			ID:           "shared",
			Path:         "icons/shared.svg",
			Source:       connectors.IconSourceOfficial,
			ReviewStatus: connectors.IconReviewOfficial,
		}},
	})
	if err != nil {
		t.Fatalf("buildIconEntries curated shared path: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %+v, want two connectors", entries)
	}
	if len(assets) != 1 || assets[0].SourceURL != "https://example.com/shared.svg" {
		t.Fatalf("assets = %+v, want the assigned upstream source URL", assets)
	}
}

func TestBuildIconEntriesRejectsDuplicateCuratedKeys(t *testing.T) {
	_, _, err := buildIconEntries(registryFile{}, buildOptions{
		ImplementedConnectors: map[string]bool{"demo": true},
		CuratedEntries: []iconEntry{
			{
				Connector:    "demo",
				ID:           "simple-icons-demo",
				Path:         "icons/simple-icons/demo.svg",
				Source:       connectors.IconSourceSimpleIcons,
				ReviewStatus: connectors.IconReviewSimpleIconsCC0Trademark,
			},
			{
				Connector:    "demo",
				ID:           "official-demo",
				Path:         "icons/demo.svg",
				Source:       connectors.IconSourceOfficial,
				ReviewStatus: connectors.IconReviewOfficial,
			},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate curated icon entry") {
		t.Fatalf("buildIconEntries duplicate curated key error = %v", err)
	}
}

func TestBuildIconEntriesRejectsAmbiguousSourceDestinationCollapse(t *testing.T) {
	_, _, err := buildIconEntries(registryFile{
		Sources: []map[string]any{{
			"public":           true,
			"dockerRepository": "registry/source-demo",
			"documentationUrl": "https://example.com/integrations/sources/demo",
			"icon":             "source.svg",
			"iconUrl":          "https://example.com/source-demo/source.svg",
		}},
		Destinations: []map[string]any{{
			"public":           true,
			"dockerRepository": "registry/destination-demo",
			"documentationUrl": "https://example.com/integrations/destinations/demo",
			"icon":             "destination.svg",
			"iconUrl":          "https://example.com/destination-demo/destination.svg",
		}},
	}, buildOptions{ImplementedConnectors: map[string]bool{"demo": true}})
	if err == nil || !strings.Contains(err.Error(), "ambiguous source/destination icon collapse") {
		t.Fatalf("buildIconEntries ambiguous collapse error = %v", err)
	}
}

func TestBuildIconEntriesAllowsCuratedRowToResolveConflictingSourceURLs(t *testing.T) {
	entries, assets, err := buildIconEntries(registryFile{
		Sources: []map[string]any{{
			"public":           true,
			"dockerRepository": "registry/source-demo",
			"documentationUrl": "https://example.com/integrations/sources/demo",
			"icon":             "demo.svg",
			"iconUrl":          "https://example.com/source-demo/demo.svg",
		}},
		Destinations: []map[string]any{{
			"public":           true,
			"dockerRepository": "registry/destination-demo",
			"documentationUrl": "https://example.com/integrations/destinations/demo",
			"icon":             "demo.svg",
			"iconUrl":          "https://example.com/destination-demo/demo.svg",
		}},
	}, buildOptions{
		ImplementedConnectors: map[string]bool{"demo": true},
		CuratedEntries: []iconEntry{{
			Connector:    "demo",
			ID:           "demo",
			Path:         "icons/demo.svg",
			Source:       connectors.IconSourceUpstream,
			ReviewStatus: connectors.IconReviewUpstreamSeeded,
			ReviewURL:    "https://example.com/review/demo",
		}},
	})
	if err != nil {
		t.Fatalf("buildIconEntries curated collapse: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %+v, want one curated demo entry", entries)
	}
	if got := entries[0]; got.Connector != "demo" || got.ReviewURL != "https://example.com/review/demo" || got.SourceURL != "" {
		t.Fatalf("entry = %+v, want the authored curated row", got)
	}
	if len(assets) != 0 {
		t.Fatalf("assets = %+v, want no silently selected conflicting upstream asset", assets)
	}
}

func TestBuildIconEntriesAllowsReviewedSameAssetCollapse(t *testing.T) {
	entries, assets, err := buildIconEntries(registryFile{
		Sources: []map[string]any{{
			"public":           true,
			"dockerRepository": "registry/source-demo",
			"documentationUrl": "https://example.com/integrations/sources/demo",
			"icon":             "demo.svg",
			"iconUrl":          "https://example.com/source-demo/demo.svg",
		}},
		Destinations: []map[string]any{{
			"public":           true,
			"dockerRepository": "registry/destination-demo",
			"documentationUrl": "https://example.com/integrations/destinations/demo",
			"icon":             "demo.svg",
			"iconUrl":          "https://example.com/source-demo/demo.svg",
		}},
	}, buildOptions{ImplementedConnectors: map[string]bool{"demo": true}})
	if err != nil {
		t.Fatalf("buildIconEntries same asset collapse: %v", err)
	}
	if len(entries) != 1 || entries[0].Connector != "demo" || entries[0].Path != "icons/demo.svg" {
		t.Fatalf("entries = %+v", entries)
	}
	if len(assets) != 1 || assets[0].Path != "icons/demo.svg" {
		t.Fatalf("assets = %+v", assets)
	}
}

func TestBuildIconEntriesRejectsSamePathWithDifferentSourceURLs(t *testing.T) {
	_, _, err := buildIconEntries(registryFile{
		Sources: []map[string]any{{
			"public":           true,
			"dockerRepository": "registry/source-demo",
			"documentationUrl": "https://example.com/integrations/sources/demo",
			"icon":             "demo.svg",
			"iconUrl":          "https://example.com/source-demo/demo.svg",
		}},
		Destinations: []map[string]any{{
			"public":           true,
			"dockerRepository": "registry/destination-demo",
			"documentationUrl": "https://example.com/integrations/destinations/demo",
			"icon":             "demo.svg",
			"iconUrl":          "https://example.com/destination-demo/demo.svg",
		}},
	}, buildOptions{ImplementedConnectors: map[string]bool{"demo": true}})
	if err == nil {
		t.Fatal("buildIconEntries accepted uncurated conflicting source URLs")
	}
	for _, want := range []string{
		"conflicting source URLs",
		"source-demo",
		"destination-demo",
		"https://example.com/source-demo/demo.svg",
		"https://example.com/destination-demo/demo.svg",
		`curated icon entry for "demo"`,
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("buildIconEntries source URL conflict error = %q, want %q", err, want)
		}
	}
}

func TestBuildIconEntriesRejectsSharedAssetPathSourceURLConflict(t *testing.T) {
	_, _, err := buildIconEntries(registryFile{Sources: []map[string]any{
		{
			"public":           true,
			"dockerRepository": "registry/source-alpha",
			"documentationUrl": "https://example.com/integrations/sources/alpha",
			"icon":             "shared.svg",
			"iconUrl":          "https://example.com/alpha/shared.svg",
		},
		{
			"public":           true,
			"dockerRepository": "registry/source-beta",
			"documentationUrl": "https://example.com/integrations/sources/beta",
			"icon":             "shared.svg",
			"iconUrl":          "https://example.com/beta/shared.svg",
		},
	}}, buildOptions{ImplementedConnectors: map[string]bool{"alpha": true, "beta": true}})
	if err == nil || !strings.Contains(err.Error(), "conflicting source URLs for shared icon path") {
		t.Fatalf("buildIconEntries shared path source URL conflict error = %v", err)
	}
}

func TestBuildIconEntriesAllowsSharedAssetPathWithIdenticalSourceURL(t *testing.T) {
	entries, assets, err := buildIconEntries(registryFile{Sources: []map[string]any{
		{
			"public":           true,
			"dockerRepository": "registry/source-alpha",
			"documentationUrl": "https://example.com/integrations/sources/alpha",
			"icon":             "shared.svg",
			"iconUrl":          "https://example.com/shared.svg",
		},
		{
			"public":           true,
			"dockerRepository": "registry/source-beta",
			"documentationUrl": "https://example.com/integrations/sources/beta",
			"icon":             "shared.svg",
			"iconUrl":          "https://example.com/shared.svg",
		},
	}}, buildOptions{ImplementedConnectors: map[string]bool{"alpha": true, "beta": true}})
	if err != nil {
		t.Fatalf("buildIconEntries shared path identical source URL: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %+v, want two connectors", entries)
	}
	if len(assets) != 1 || assets[0].Path != "icons/shared.svg" || assets[0].SourceURL != "https://example.com/shared.svg" {
		t.Fatalf("assets = %+v, want one shared asset", assets)
	}
}

func TestValidateBuiltIconEntryUsesSlashOrientedPaths(t *testing.T) {
	entry := iconEntry{
		Connector:    "demo",
		ID:           "demo",
		Path:         "icons/simple-icons/demo.svg",
		Source:       connectors.IconSourceSimpleIcons,
		ReviewStatus: connectors.IconReviewSimpleIconsCC0Trademark,
	}
	if err := validateBuiltIconEntry(entry); err != nil {
		t.Fatalf("validateBuiltIconEntry() rejected forward-slash path: %v", err)
	}

	entry.Path = `icons/simple-icons\demo.svg`
	if err := validateBuiltIconEntry(entry); err == nil {
		t.Fatal("validateBuiltIconEntry() accepted backslash path")
	}
}

func writeCuratedRegistry(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "icon_data.json")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write curated registry fixture: %v", err)
	}
	return path
}

func TestLoadCuratedIconEntriesRejectsEmptyConnectorKey(t *testing.T) {
	path := writeCuratedRegistry(t, `[{"connector":"","id":"demo","path":"icons/demo.svg","source":"official_site","review_status":"official"}]`)
	_, err := loadCuratedIconEntries(path)
	if err == nil {
		t.Fatal("loadCuratedIconEntries() accepted an empty curated connector key")
	}
	if !strings.Contains(err.Error(), "entry missing connector") || !strings.Contains(err.Error(), path) {
		t.Fatalf("loadCuratedIconEntries() error = %v, want connector-required error naming %q", err, path)
	}
}

func TestLoadCuratedIconEntriesRejectsSourcePrefixedConnectorKey(t *testing.T) {
	path := writeCuratedRegistry(t, `[{"connector":"source-github","id":"github","path":"icons/github.svg","source":"official_site","review_status":"official"}]`)
	_, err := loadCuratedIconEntries(path)
	if err == nil {
		t.Fatal("loadCuratedIconEntries() accepted a source-prefixed curated connector key")
	}
	if !strings.Contains(err.Error(), `"source-github"`) || !strings.Contains(err.Error(), "must be a bare connector identifier") || !strings.Contains(err.Error(), path) {
		t.Fatalf("loadCuratedIconEntries() error = %v, want bare-identifier error naming the key and %q", err, path)
	}
}

func TestLoadCuratedIconEntriesRejectsDestinationPrefixedConnectorKey(t *testing.T) {
	path := writeCuratedRegistry(t, `[{"connector":"destination-github","id":"github","path":"icons/github.svg","source":"official_site","review_status":"official"}]`)
	_, err := loadCuratedIconEntries(path)
	if err == nil {
		t.Fatal("loadCuratedIconEntries() accepted a destination-prefixed curated connector key")
	}
	if !strings.Contains(err.Error(), `"destination-github"`) || !strings.Contains(err.Error(), "must be a bare connector identifier") || !strings.Contains(err.Error(), path) {
		t.Fatalf("loadCuratedIconEntries() error = %v, want bare-identifier error naming the key and %q", err, path)
	}
}

func TestLoadCuratedIconEntriesAcceptsBareConnectorKey(t *testing.T) {
	path := writeCuratedRegistry(t, `[{"connector":"github","id":"github","path":"icons/github.svg","source":"official_site","review_status":"official","review_url":"https://github.com"}]`)
	entries, err := loadCuratedIconEntries(path)
	if err != nil {
		t.Fatalf("loadCuratedIconEntries() unexpected error = %v", err)
	}
	if len(entries) != 1 || entries[0].Connector != "github" || entries[0].ReviewURL != "https://github.com" {
		t.Fatalf("entries = %+v, want single preserved bare github entry", entries)
	}
}

func TestBuildIconEntryStillCollapsesPrefixedRawUpstreamSlug(t *testing.T) {
	entry, ok, err := buildIconEntry(map[string]any{
		"public":           true,
		"dockerRepository": "registry/destination-github",
		"documentationUrl": "https://example.com/integrations/destinations/github",
		"icon":             "github.svg",
		"iconUrl":          "https://example.com/destination-github/icon.svg",
	})
	if err != nil || !ok {
		t.Fatalf("buildIconEntry() ok=%t err=%v", ok, err)
	}
	if entry.Connector != "github" {
		t.Fatalf("entry.Connector = %q, want raw upstream destination- prefix collapsed to bare %q", entry.Connector, "github")
	}
}
