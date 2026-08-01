package defs

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestProductionEmbedLoadsRuntimeBundles(t *testing.T) {
	bundles, err := engine.LoadAll(FS)
	if err != nil {
		t.Fatalf("LoadAll(FS): %v", err)
	}
	if len(bundles) == 0 {
		t.Fatal("LoadAll(FS) returned zero bundles")
	}

	var github *engine.Bundle
	for i := range bundles {
		if bundles[i].Name == "github" {
			github = &bundles[i]
			break
		}
	}
	if github == nil {
		t.Fatal("LoadAll(FS) missing github bundle")
	}
	if github.Metadata.Name != "github" {
		t.Fatalf("github metadata name = %q", github.Metadata.Name)
	}
	if len(github.Streams) == 0 {
		t.Fatal("github bundle has zero streams")
	}
	if github.Docs == "" {
		t.Fatal("github bundle docs are empty")
	}
	if github.Surface != nil {
		t.Fatal("production embed should not include api_surface.json")
	}
	if github.Fixtures != nil {
		t.Fatal("production embed should not include fixtures")
	}
}

func TestAirtableRuntimeBundleSafetyContract(t *testing.T) {
	bundles, err := engine.LoadAll(FS)
	if err != nil {
		t.Fatalf("LoadAll(FS): %v", err)
	}
	var airtable *engine.Bundle
	for i := range bundles {
		if bundles[i].Name == "airtable" {
			airtable = &bundles[i]
			break
		}
	}
	if airtable == nil {
		t.Fatal("LoadAll(FS) missing airtable bundle")
	}
	if airtable.Metadata.Capabilities.Query || airtable.Metadata.Capabilities.CDC {
		t.Fatalf("airtable capabilities query=%v cdc=%v, want both false", airtable.Metadata.Capabilities.Query, airtable.Metadata.Capabilities.CDC)
	}

	valid := []connectors.Record{{"records": []any{"recA1B2C3D4E5F6G7"}}}
	if err := engine.ValidateWrite(context.Background(), *airtable, connectors.WriteRequest{Action: "delete_multiple_records"}, valid); err != nil {
		t.Fatalf("ValidateWrite valid delete_multiple_records: %v", err)
	}
	unsafe := []connectors.Record{{"records": []any{"recA&records[]=recB"}}}
	err = engine.ValidateWrite(context.Background(), *airtable, connectors.WriteRequest{Action: "delete_multiple_records"}, unsafe)
	if err == nil {
		t.Fatal("ValidateWrite unsafe delete_multiple_records = nil, want error")
	}
	if !strings.Contains(err.Error(), "pattern") {
		t.Fatalf("ValidateWrite unsafe error = %q, want pattern validation", err.Error())
	}
}

func TestProductionEmbedExcludesConformanceArtifacts(t *testing.T) {
	for _, path := range []string{"github/api_surface.json", "github/fixtures"} {
		if _, err := fs.Stat(FS, path); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("fs.Stat(%q) err = %v, want fs.ErrNotExist", path, err)
		}
	}
}
