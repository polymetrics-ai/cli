package main

import "testing"

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
	if entry.Connector != "source-github" || entry.ID != "github" || entry.Path != "icons/github.svg" {
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
	if entry.ID != "source-demo" || entry.Path != "icons/source-demo.svg" {
		t.Fatalf("entry identity = %+v", entry)
	}
}

func TestBuildIconEntriesIncludesBuiltins(t *testing.T) {
	entries, assets, err := buildIconEntries(registryFile{Sources: []map[string]any{{
		"public":           true,
		"dockerRepository": "registry/source-demo",
		"documentationUrl": "https://example.com/integrations/sources/demo",
		"icon":             "demo.svg",
		"iconUrl":          "https://example.com/source-demo/icon.svg",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 6 {
		t.Fatalf("entries len = %d, want 6", len(entries))
	}
	if len(assets) != 1 || assets[0].Path != "icons/demo.svg" {
		t.Fatalf("assets = %+v", assets)
	}
	seenLocal := map[string]bool{}
	for _, entry := range entries {
		if entry.Source != "upstream_registry" {
			seenLocal[entry.Connector] = true
		}
	}
	for _, want := range []string{"file", "outbox", "sample", "searxng", "warehouse"} {
		if !seenLocal[want] {
			t.Fatalf("missing built-in icon entry %q in %+v", want, entries)
		}
	}
}
