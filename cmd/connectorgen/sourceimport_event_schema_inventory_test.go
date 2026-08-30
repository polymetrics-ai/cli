package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestSourceImportV3EventSchemaInventoryRejectsInvalidSelectors keeps the
// event-schema inventory a closed source selector, rather than a second copy of
// an OpenAPI component schema in the lock. The selector may name only a schema
// in one declared source document.
func TestSourceImportV3EventSchemaInventoryRejectsInvalidSelectors(t *testing.T) {
	tests := []struct {
		name       string
		documents  []sourceImportEventInventoryFixtureDocument
		inventory  map[string]any
		wantParse  string
		wantImport string
	}{
		{
			name: "unknown source document",
			documents: []sourceImportEventInventoryFixtureDocument{{
				ID:      "alpha",
				Schemas: map[string]any{"EventResponse": map[string]any{"type": "object"}},
			}},
			inventory: sourceImportEventInventory("missing", "EventResponse"),
			wantParse: "unknown source document",
		},
		{
			name: "duplicate schema name",
			documents: []sourceImportEventInventoryFixtureDocument{{
				ID:      "alpha",
				Schemas: map[string]any{"EventResponse": map[string]any{"type": "object"}},
			}},
			inventory: map[string]any{
				"source_document": "alpha",
				"schemas": []any{
					sourceImportEventSchemaSelectorWire("EventResponse"),
					sourceImportEventSchemaSelectorWire("EventResponse"),
				},
			},
			wantParse: "duplicates schema name",
		},
		{
			name: "missing schema name",
			documents: []sourceImportEventInventoryFixtureDocument{{
				ID:      "alpha",
				Schemas: map[string]any{"EventResponse": map[string]any{"type": "object"}},
			}},
			inventory: map[string]any{
				"source_document": "alpha",
				"schemas": []any{map[string]any{
					"source_location": `components.schemas["EventResponse"]`,
				}},
			},
			wantParse: "schema name",
		},
		{
			name: "non schema selector",
			documents: []sourceImportEventInventoryFixtureDocument{{
				ID:      "alpha",
				Schemas: map[string]any{"EventResponse": map[string]any{"type": "object"}},
			}},
			inventory: map[string]any{
				"source_document": "alpha",
				"schemas": []any{map[string]any{
					"name":            "EventResponse",
					"source_location": `components.parameters["EventResponse"]`,
				}},
			},
			wantParse: "canonical schema source location",
		},
		{
			name: "selected source entry is not a schema object",
			documents: []sourceImportEventInventoryFixtureDocument{{
				ID:      "alpha",
				Schemas: map[string]any{"EventResponse": "not-a-schema"},
			}},
			inventory:  sourceImportEventInventory("alpha", "EventResponse"),
			wantImport: "does not resolve to an object schema",
		},
		{
			name: "selector does not belong to bound source document",
			documents: []sourceImportEventInventoryFixtureDocument{
				{ID: "alpha", Schemas: map[string]any{"EventResponse": map[string]any{"type": "object"}}},
				{ID: "beta", Schemas: map[string]any{"OtherResponse": map[string]any{"type": "object"}}},
			},
			inventory:  sourceImportEventInventory("beta", "EventResponse"),
			wantImport: "does not resolve to an object schema",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, artifacts := sourceImportEventInventoryFixture(t, test.documents, test.inventory)
			lock, err := parseSourceImportLock(raw, "fixture")
			if test.wantParse != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantParse) {
					t.Fatalf("parse event-schema inventory error = %v, want %q", err, test.wantParse)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse event-schema inventory fixture: %v", err)
			}
			_, err = importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(_ context.Context, sourceURL string) ([]byte, error) {
				raw, ok := artifacts[sourceURL]
				if !ok {
					return nil, fmt.Errorf("unexpected source URL %q", sourceURL)
				}
				return raw, nil
			}), defaultSourceImportLimits())
			if err == nil || !strings.Contains(err.Error(), test.wantImport) {
				t.Fatalf("import event-schema inventory error = %v, want %q", err, test.wantImport)
			}
		})
	}
}

func TestSourceImportRetainedAsanaV3EventInventoryResolvesExactlyFiveSchemas(t *testing.T) {
	defsRoot := filepath.Join("..", "..", "internal", "connectors", "defs")
	lockPath := filepath.Join(defsRoot, "asana", "sources", "asana-operation-source-lock.json")
	raw, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read retained Asana v3 lock: %v", err)
	}
	baselineRaw := sourceImportRemoveEventSchemaInventory(t, raw)
	baseline, err := parseSourceImportLock(baselineRaw, "asana")
	if err != nil {
		t.Fatalf("parse retained Asana v3 lock without event-schema inventory: %v", err)
	}
	lock, err := parseSourceImportLock(raw, "asana")
	if err != nil {
		t.Fatalf("parse retained Asana v3 event-schema inventory: %v", err)
	}
	if got, want := sourceImportEventInventoryOperationTuples(lock), sourceImportEventInventoryOperationTuples(baseline); !reflect.DeepEqual(got, want) {
		t.Fatalf("event-schema inventory changed source operation identities: got %v, want %v", got, want)
	}
	inventory := lock.Rest.EventSchemaInventory
	if inventory == nil || inventory.SourceDocument != "asana-openapi" || len(inventory.Schemas) != 5 {
		t.Fatalf("Asana event-schema inventory = %+v, want five selectors for asana-openapi", inventory)
	}
	fetcher, err := newConnectorSourceImportRetainedArtifactFetcher(defsRoot, "asana", defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("construct retained Asana artifact fetcher: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, fetcher, defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("resolve retained Asana v3 event-schema inventory: %v", err)
	}
	if got, want := len(result.Operations), 249; got != want {
		t.Fatalf("retained Asana operation count after event-schema resolution = %d, want %d", got, want)
	}
}

type sourceImportEventInventoryFixtureDocument struct {
	ID      string
	Schemas map[string]any
}

func sourceImportEventInventoryFixture(t *testing.T, documents []sourceImportEventInventoryFixtureDocument, inventory map[string]any) ([]byte, map[string][]byte) {
	t.Helper()
	fixtureDocuments := make([]sourceImportV3FixtureDocument, 0, len(documents))
	artifacts := make(map[string][]byte, len(documents))
	for _, document := range documents {
		path := fmt.Sprintf("/%s", document.ID)
		operationID := "get" + strings.Title(document.ID)
		artifactDocument := map[string]any{
			"openapi": "3.0.3",
			"info":    map[string]any{"title": document.ID, "version": "1"},
			"paths": map[string]any{
				path: map[string]any{"get": map[string]any{
					"operationId": operationID,
					"responses":   map[string]any{"200": map[string]any{"description": "ok"}},
				}},
			},
			"components": map[string]any{"schemas": document.Schemas},
		}
		artifact, err := json.Marshal(artifactDocument)
		if err != nil {
			t.Fatalf("encode %s fixture artifact: %v", document.ID, err)
		}
		url := fmt.Sprintf("https://fixtures.polymetrics.invalid/%s.openapi.json", document.ID)
		fixtureDocuments = append(fixtureDocuments, sourceImportV3FixtureDocument{
			ID:          document.ID,
			Path:        path,
			OperationID: operationID,
			Artifact:    artifact,
			ArtifactURL: url,
		})
		artifacts[url] = artifact
	}
	raw := sourceImportV3FixtureLock(t, "fixture", fixtureDocuments)
	return sourceImportAddEventSchemaInventory(t, raw, inventory), artifacts
}

func sourceImportAddEventSchemaInventory(t *testing.T, raw []byte, inventory map[string]any) []byte {
	t.Helper()
	var lock map[string]any
	if err := json.Unmarshal(raw, &lock); err != nil {
		t.Fatalf("decode source lock fixture: %v", err)
	}
	rest, ok := lock["rest"].(map[string]any)
	if !ok {
		t.Fatalf("source lock REST = %T, want object", lock["rest"])
	}
	rest["event_schema_inventory"] = inventory
	encoded, err := json.Marshal(lock)
	if err != nil {
		t.Fatalf("encode source lock fixture: %v", err)
	}
	return encoded
}

func sourceImportRemoveEventSchemaInventory(t *testing.T, raw []byte) []byte {
	t.Helper()
	var lock map[string]any
	if err := json.Unmarshal(raw, &lock); err != nil {
		t.Fatalf("decode source lock fixture: %v", err)
	}
	rest, ok := lock["rest"].(map[string]any)
	if !ok {
		t.Fatalf("source lock REST = %T, want object", lock["rest"])
	}
	delete(rest, "event_schema_inventory")
	encoded, err := json.Marshal(lock)
	if err != nil {
		t.Fatalf("encode source lock fixture: %v", err)
	}
	return encoded
}

func sourceImportEventInventory(sourceDocument, name string) map[string]any {
	return map[string]any{
		"source_document": sourceDocument,
		"schemas":         []any{sourceImportEventSchemaSelectorWire(name)},
	}
}

func sourceImportEventSchemaSelectorWire(name string) map[string]any {
	return map[string]any{
		"name":            name,
		"source_location": `components.schemas["` + name + `"]`,
	}
}

func sourceImportEventInventoryOperationTuples(lock sourceImportLock) []string {
	tuples := []string{}
	for _, document := range lock.Rest.SourceDocuments {
		for _, operation := range document.Operations {
			tuples = append(tuples, operation.ID+"\x00"+operation.Method+"\x00"+operation.Path+"\x00"+operation.OperationID+"\x00"+operation.SourceLocation)
		}
	}
	return tuples
}
