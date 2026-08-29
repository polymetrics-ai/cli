package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSourceImportLockAcceptsAsanaSourceOperationEnrichment(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "internal", "connectors", "defs", "asana", "sources", "asana-operation-source-lock.json"))
	if err != nil {
		t.Fatalf("read retained Asana source lock: %v", err)
	}

	lock, err := parseSourceImportLock(raw, "asana")
	if err != nil {
		t.Fatalf("parse retained Asana source lock with source operation enrichment: %v", err)
	}
	if got, want := len(lock.Rest.Operations), 249; got != want {
		t.Fatalf("retained Asana REST operations = %d, want %d", got, want)
	}
}

func TestParseSourceImportLockSourceOperationEnrichmentStaysClosed(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		mutate  func(t *testing.T, operation map[string]any)
		wantErr string
	}{
		{
			name: "accepts documented enrichment",
		},
		{
			name: "rejects undeclared operation member",
			mutate: func(_ *testing.T, operation map[string]any) {
				operation["undocumented"] = true
			},
			wantErr: `unknown field "undocumented"`,
		},
		{
			name: "rejects undeclared inline parameter member",
			mutate: func(t *testing.T, operation map[string]any) {
				parameters, ok := operation["parameters"].([]any)
				if !ok || len(parameters) == 0 {
					t.Fatal("fixture has no inline parameters")
				}
				parameter, ok := parameters[0].(map[string]any)
				if !ok {
					t.Fatal("fixture parameter is not an object")
				}
				parameter["undocumented"] = true
			},
			wantErr: `unknown field "undocumented"`,
		},
		{
			name:   "rejects non-object enrichment",
			target: "operation",
			mutate: func(_ *testing.T, operation map[string]any) {
				operation["source_operation"] = []any{}
			},
			wantErr: "source_operation",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, _ := sourceImportSourceOperationCompatibilityFixture(t)
			if test.mutate != nil {
				var document map[string]any
				if err := json.Unmarshal(raw, &document); err != nil {
					t.Fatalf("decode source lock fixture: %v", err)
				}
				rest, ok := document["rest"].(map[string]any)
				if !ok {
					t.Fatal("fixture REST document is not an object")
				}
				operations, ok := rest["operations"].([]any)
				if !ok || len(operations) != 1 {
					t.Fatalf("fixture REST operations = %#v, want one operation", rest["operations"])
				}
				lockedOperation, ok := operations[0].(map[string]any)
				if !ok {
					t.Fatal("fixture REST operation is not an object")
				}
				if test.target == "operation" {
					test.mutate(t, lockedOperation)
				} else {
					enrichment, ok := lockedOperation["source_operation"].(map[string]any)
					if !ok {
						t.Fatal("fixture source operation enrichment is not an object")
					}
					test.mutate(t, enrichment)
				}
				var err error
				raw, err = json.Marshal(document)
				if err != nil {
					t.Fatalf("encode mutated source lock fixture: %v", err)
				}
			}

			_, err := parseSourceImportLock(raw, "alpha")
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("parse source lock with documented enrichment: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("parse source lock error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func TestParseSourceImportLockSourceContractEnrichmentStaysClosed(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(t *testing.T, contract map[string]any)
		wantErr string
	}{
		{
			name: "rejects undeclared contract member",
			mutate: func(_ *testing.T, contract map[string]any) {
				contract["undocumented"] = true
			},
			wantErr: `unknown field "undocumented"`,
		},
		{
			name: "rejects undeclared component member",
			mutate: func(t *testing.T, contract map[string]any) {
				components, ok := contract["components"].(map[string]any)
				if !ok {
					t.Fatal("fixture source contract components are not an object")
				}
				components["undocumented"] = map[string]any{}
			},
			wantErr: `unknown field "undocumented"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, _ := sourceImportSourceOperationCompatibilityFixture(t)
			var document map[string]any
			if err := json.Unmarshal(raw, &document); err != nil {
				t.Fatalf("decode source lock fixture: %v", err)
			}
			contract, ok := document["source_contract"].(map[string]any)
			if !ok {
				t.Fatal("fixture source contract is not an object")
			}
			test.mutate(t, contract)
			raw, err := json.Marshal(document)
			if err != nil {
				t.Fatalf("encode mutated source lock fixture: %v", err)
			}

			_, err = parseSourceImportLock(raw, "alpha")
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("parse source lock error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func TestSourceOperationEnrichmentCannotRelaxRetainedArtifactHash(t *testing.T) {
	raw, artifact := sourceImportSourceOperationCompatibilityFixture(t)
	lock, err := parseSourceImportLock(raw, "alpha")
	if err != nil {
		t.Fatalf("parse source lock with source operation enrichment: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return artifact, nil
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import hash-pinned source lock with enrichment: %v", err)
	}
	if got, want := len(result.Operations), 1; got != want {
		t.Fatalf("imported operation count = %d, want %d", got, want)
	}

	mismatched := append([]byte(nil), artifact...)
	mismatched = append(mismatched, ' ')
	_, err = importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
		return mismatched, nil
	}), defaultSourceImportLimits())
	if err == nil || !strings.Contains(err.Error(), "does not match locked bytes and SHA-256") {
		t.Fatalf("retained artifact hash error = %v, want locked-byte and SHA-256 mismatch", err)
	}
}

func sourceImportSourceOperationCompatibilityFixture(t *testing.T) ([]byte, []byte) {
	t.Helper()
	artifact := []byte(`{"openapi":"3.0.3","info":{"title":"fixture","version":"1"},"paths":{"/items/{item_id}":{"post":{"operationId":"createItem","parameters":[{"name":"item_id","in":"path","required":true,"schema":{"type":"string"}},{"name":"limit","in":"query","schema":{"type":"integer"}}],"requestBody":{"required":true,"content":{"application/json":{"schema":{"type":"object"}}}},"responses":{"200":{"description":"ok"}}}}}}`)
	digest := sha256.Sum256(artifact)
	document := map[string]any{
		"schema_version": 2,
		"connector":      "alpha",
		"source_contract": map[string]any{
			"openapi": "3.0.3",
			"servers": []any{map[string]any{
				"url":         "https://api.fixtures.polymetrics.invalid/v1",
				"description": "fixture server",
			}},
			"security": []any{map[string]any{"token": []any{}}},
			"components": map[string]any{
				"parameters":      map[string]any{"limit": map[string]any{"name": "limit", "in": "query"}},
				"responses":       map[string]any{"ok": map[string]any{"description": "ok"}},
				"schemas":         map[string]any{"item": map[string]any{"type": "object"}},
				"securitySchemes": map[string]any{"token": map[string]any{"type": "http"}},
			},
		},
		"rest": map[string]any{
			"source_url":   "https://fixtures.polymetrics.invalid/alpha-openapi.json",
			"sha256":       hex.EncodeToString(digest[:]),
			"bytes":        len(artifact),
			"openapi":      "3.0.3",
			"info_version": "1",
			"operations": []any{map[string]any{
				"id":              "alpha.rest.createItem",
				"protocol":        "rest",
				"method":          "POST",
				"path":            "/items/{item_id}",
				"operation_id":    "createItem",
				"deprecated":      false,
				"source_location": `paths["/items/{item_id}"].post`,
				"source_operation": map[string]any{
					"summary":     "Create an item",
					"description": "Creates one item.",
					"tags":        []any{"items"},
					"security":    []any{map[string]any{"token": []any{}}},
					"path_parameters": []any{map[string]any{
						"$ref": "#/components/parameters/item_id",
					}},
					"parameters": []any{map[string]any{
						"name":        "limit",
						"in":          "query",
						"description": "The number of items.",
						"required":    false,
						"example":     10,
						"schema":      map[string]any{"type": "integer"},
						"style":       "form",
						"explode":     true,
					}},
					"requestBody": map[string]any{
						"description": "Item input.",
						"required":    true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema":   map[string]any{"type": "object"},
								"examples": map[string]any{"one": map[string]any{"value": map[string]any{"name": "example"}}},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "ok",
							"content": map[string]any{
								"application/json": map[string]any{"schema": map[string]any{"type": "object"}},
							},
						},
					},
				},
			}},
		},
		"counts": map[string]any{"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1},
	}
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("encode source lock fixture: %v", err)
	}
	return raw, artifact
}
