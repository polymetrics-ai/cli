package main

import (
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
