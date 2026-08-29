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

func TestParseSourceImportLockAcceptsAsanaV3DocumentOwnedInventory(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "internal", "connectors", "defs", "asana", "sources", "asana-operation-source-lock.json"))
	if err != nil {
		t.Fatalf("read retained Asana source lock: %v", err)
	}

	lock, err := parseSourceImportLock(raw, "asana")
	if err != nil {
		t.Fatalf("parse retained Asana v3 source lock: %v", err)
	}
	if lock.SchemaVersion != 3 || len(lock.Rest.SourceDocuments) != 1 || len(lock.Rest.SourceDocuments[0].Operations) != 249 {
		t.Fatalf("retained Asana v3 inventory = schema %d documents=%d operations=%d, want schema 3/one document/249 operations", lock.SchemaVersion, len(lock.Rest.SourceDocuments), len(lock.Rest.SourceDocuments[0].Operations))
	}
	inventory := lock.Rest.EventSchemaInventory
	if inventory == nil || inventory.SourceDocument != "asana-openapi" || len(inventory.Schemas) != 5 {
		t.Fatalf("retained Asana v3 event-schema inventory = %+v, want five selectors for asana-openapi", inventory)
	}
}

func TestParseSourceImportLockRetainsBatchOneLegacyProviderFacts(t *testing.T) {
	tests := []struct {
		connector string
		wantREST  int
	}{
		{connector: "bitbucket", wantREST: 297},
		{connector: "circleci", wantREST: 111},
		{connector: "dockerhub", wantREST: 54},
		{connector: "gitlab", wantREST: 1752},
		{connector: "jira", wantREST: 617},
		{connector: "notion", wantREST: 49},
		{connector: "sentry", wantREST: 223},
		{connector: "stripe", wantREST: 589},
		{connector: "vercel", wantREST: 400},
	}

	for _, test := range tests {
		t.Run(test.connector, func(t *testing.T) {
			lockPath := filepath.Join("..", "..", "internal", "connectors", "defs", test.connector, "sources", test.connector+"-operation-source-lock.json")
			raw, err := os.ReadFile(lockPath)
			if err != nil {
				t.Fatalf("read Batch-1 source lock: %v", err)
			}
			lock, err := parseSourceImportLock(raw, test.connector)
			if err != nil {
				t.Fatalf("parse Batch-1 source lock: %v", err)
			}
			if got := len(lock.Rest.Operations); got != test.wantREST {
				t.Fatalf("REST source identities = %d, want %d", got, test.wantREST)
			}
			if contract := lock.SourceContract; contract == nil {
				t.Fatal("legacy source contract was dropped from the canonical source lock")
			}

			if test.connector != "gitlab" {
				return
			}
			contract := sourceImportProviderObjectForTest(t, lock.SourceContract.Raw, "GitLab source contract")
			servers := sourceImportProviderArrayForTest(t, contract["servers"], "GitLab servers")
			if len(servers) != 1 {
				t.Fatalf("GitLab source contract servers = %d, want one", len(servers))
			}
			server := sourceImportProviderObjectForTest(t, servers[0], "GitLab server")
			variables := sourceImportProviderObjectForTest(t, server["variables"], "GitLab server variables")
			hostname := sourceImportProviderObjectForTest(t, variables["hostname"], "GitLab hostname variable")
			if got := string(hostname["default"]); got != `"gitlab.com"` {
				t.Fatalf("GitLab hostname default = %s, want gitlab.com", got)
			}
		})
	}
}

func TestParseSourceImportLockRetainsSingularMediaExamples(t *testing.T) {
	tests := []struct {
		name        string
		example     any
		wantExample string
	}{
		{
			name:        "object",
			example:     map[string]any{"name": "fixture"},
			wantExample: `{"name":"fixture"}`,
		},
		{
			name:        "explicit null",
			example:     nil,
			wantExample: "null",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, _ := sourceImportSourceOperationCompatibilityFixture(t)
			var document map[string]any
			if err := json.Unmarshal(raw, &document); err != nil {
				t.Fatalf("decode source lock fixture: %v", err)
			}
			rest := document["rest"].(map[string]any)
			documentOperation := rest["operations"].([]any)[0].(map[string]any)
			enrichment := documentOperation["source_operation"].(map[string]any)
			documentResponses := enrichment["responses"].(map[string]any)
			documentResponse := documentResponses["200"].(map[string]any)
			documentContent := documentResponse["content"].(map[string]any)
			documentMedia := documentContent["application/json"].(map[string]any)
			documentMedia["example"] = test.example
			raw, err := json.Marshal(document)
			if err != nil {
				t.Fatalf("encode source lock fixture: %v", err)
			}

			lock, err := parseSourceImportLock(raw, "alpha")
			if err != nil {
				t.Fatalf("parse source lock with singular media example: %v", err)
			}
			parsedOperation := sourceImportProviderObjectForTest(t, lock.Rest.Operations[0].SourceOperation.Raw, "source operation")
			parsedResponses := sourceImportProviderObjectForTest(t, parsedOperation["responses"], "source operation responses")
			parsedResponse := sourceImportProviderObjectForTest(t, parsedResponses["200"], "source operation response")
			parsedContent := sourceImportProviderObjectForTest(t, parsedResponse["content"], "source operation response content")
			parsedMedia := sourceImportProviderObjectForTest(t, parsedContent["application/json"], "source operation response media")
			if got := string(parsedMedia["example"]); got != test.wantExample {
				t.Fatalf("singular media example = %s, want %s", got, test.wantExample)
			}
		})
	}
}

func sourceImportProviderObjectForTest(t *testing.T, raw json.RawMessage, name string) map[string]json.RawMessage {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	if object == nil {
		t.Fatalf("%s is not an object", name)
	}
	return object
}

func sourceImportProviderArrayForTest(t *testing.T, raw json.RawMessage, name string) []json.RawMessage {
	t.Helper()
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	return values
}

func TestParseSourceImportLockSourceOperationEnrichmentRetainsProviderFragments(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		mutate       func(t *testing.T, operation map[string]any)
		wantErr      string
		wantRetained string
	}{
		{
			name: "accepts documented enrichment",
		},
		{
			name: "retains provider operation member",
			mutate: func(_ *testing.T, operation map[string]any) {
				operation["x-provider-operation"] = true
			},
			wantRetained: `"x-provider-operation":true`,
		},
		{
			name: "retains provider parameter member",
			mutate: func(t *testing.T, operation map[string]any) {
				parameters, ok := operation["parameters"].([]any)
				if !ok || len(parameters) == 0 {
					t.Fatal("fixture has no inline parameters")
				}
				parameter, ok := parameters[0].(map[string]any)
				if !ok {
					t.Fatal("fixture parameter is not an object")
				}
				parameter["x-provider-parameter"] = true
			},
			wantRetained: `"x-provider-parameter":true`,
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

			lock, err := parseSourceImportLock(raw, "alpha")
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("parse source lock with documented enrichment: %v", err)
				}
				if test.wantRetained != "" && !strings.Contains(string(lock.Rest.Operations[0].SourceOperation.Raw), test.wantRetained) {
					t.Fatalf("source operation fragment %s does not retain %s", lock.Rest.Operations[0].SourceOperation.Raw, test.wantRetained)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("parse source lock error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func TestParseSourceImportLockSourceContractEnrichmentRetainsProviderFragments(t *testing.T) {
	tests := []struct {
		name         string
		mutate       func(t *testing.T, document map[string]any)
		wantErr      string
		wantRetained string
	}{
		{
			name: "retains provider contract member",
			mutate: func(_ *testing.T, document map[string]any) {
				contract := document["source_contract"].(map[string]any)
				contract["undocumented"] = true
			},
			wantRetained: `"undocumented":true`,
		},
		{
			name: "retains provider component member",
			mutate: func(t *testing.T, document map[string]any) {
				contract := document["source_contract"].(map[string]any)
				components, ok := contract["components"].(map[string]any)
				if !ok {
					t.Fatal("fixture source contract components are not an object")
				}
				components["undocumented"] = map[string]any{}
			},
			wantRetained: `"undocumented":{}`,
		},
		{
			name: "rejects non-object source contract",
			mutate: func(_ *testing.T, document map[string]any) {
				document["source_contract"] = []any{}
			},
			wantErr: "source_contract must be an object",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, _ := sourceImportSourceOperationCompatibilityFixture(t)
			var document map[string]any
			if err := json.Unmarshal(raw, &document); err != nil {
				t.Fatalf("decode source lock fixture: %v", err)
			}
			test.mutate(t, document)
			raw, err := json.Marshal(document)
			if err != nil {
				t.Fatalf("encode mutated source lock fixture: %v", err)
			}

			lock, err := parseSourceImportLock(raw, "alpha")
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("parse source lock: %v", err)
				}
				if lock.SourceContract == nil || !strings.Contains(string(lock.SourceContract.Raw), test.wantRetained) {
					t.Fatalf("source contract fragment does not retain %s", test.wantRetained)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("parse source lock error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func TestParseSourceImportLockProviderFragmentsKeepRootClosed(t *testing.T) {
	raw, _ := sourceImportSourceOperationCompatibilityFixture(t)
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode source lock fixture: %v", err)
	}
	document["undocumented"] = true
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("encode source lock fixture: %v", err)
	}

	_, err = parseSourceImportLock(raw, "alpha")
	if err == nil || !strings.Contains(err.Error(), `unknown field "undocumented"`) {
		t.Fatalf("parse source lock error = %v, want strict root rejection", err)
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
