package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestSourceImportV3BatchActionInventoryResolvesClosedProviderContract(t *testing.T) {
	t.Parallel()
	raw, artifact := sourceImportBatchActionInventoryFixture(t)
	lock, err := parseSourceImportLock(raw, "fixture")
	if err != nil {
		t.Fatalf("parse batch-action inventory: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(_ context.Context, sourceURL string) ([]byte, error) {
		if sourceURL != "https://fixtures.polymetrics.invalid/primary.openapi.json" {
			return nil, fmt.Errorf("unexpected source URL %q", sourceURL)
		}
		return artifact, nil
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("resolve batch-action inventory: %v", err)
	}
	if len(result.Operations) != 1 {
		t.Fatalf("operation count = %d, want 1", len(result.Operations))
	}
	encoded, err := json.Marshal(result.Operations[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"batch_action"`, `"max_actions":10`, `"request_actions_field":"actions"`, `"response_status_field":"status_code"`} {
		if !strings.Contains(string(encoded), want) {
			t.Errorf("batch source descriptor = %s, want %s", encoded, want)
		}
	}
	if _, err := parseDeclarationAdmissionSourceLock(raw, "fixture"); err != nil {
		t.Fatalf("mapping reader rejected closed batch-action inventory: %v", err)
	}
}

func TestSourceImportV3BatchActionInventoryRejectsUnboundOrOpenContracts(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name   string
		mutate func(map[string]any)
		want   string
	}{
		{
			name: "unknown source operation",
			mutate: func(inventory map[string]any) {
				inventory["source_operation"] = "fixture.rest.missing"
			},
			want: "unknown source operation",
		},
		{
			name: "noncanonical schema selector",
			mutate: func(inventory map[string]any) {
				inventory["action_schema"].(map[string]any)["source_location"] = `paths["/batch"]`
			},
			want: "canonical schema source location",
		},
		{
			name: "open raw provider schema",
			mutate: func(inventory map[string]any) {
				inventory["provider_schema"] = map[string]any{"type": "object"}
			},
			want: `unknown field "provider_schema"`,
		},
		{
			name: "unproved max actions",
			mutate: func(inventory map[string]any) {
				inventory["max_actions"] = 11
			},
			want: "max actions evidence",
		},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			raw, artifact := sourceImportBatchActionInventoryFixture(t)
			var document map[string]any
			if err := json.Unmarshal(raw, &document); err != nil {
				t.Fatal(err)
			}
			inventory := document["rest"].(map[string]any)["batch_action_inventory"].(map[string]any)
			testCase.mutate(inventory)
			mutated, err := json.Marshal(document)
			if err != nil {
				t.Fatal(err)
			}
			lock, err := parseSourceImportLock(mutated, "fixture")
			if err == nil {
				_, err = importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) {
					return artifact, nil
				}), defaultSourceImportLimits())
			}
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("batch-action inventory error = %v, want %q", err, testCase.want)
			}
		})
	}
}

func sourceImportBatchActionInventoryFixture(t *testing.T) ([]byte, []byte) {
	t.Helper()
	artifact := []byte(`{
  "openapi":"3.0.3","info":{"title":"batch","version":"1"},
  "tags":[{"name":"Batch API","description":"The maximum number of actions allowed in a single batch request is 10."}],
  "components":{"schemas":{
    "BatchRequest":{"type":"object","properties":{"actions":{"type":"array","items":{"$ref":"#/components/schemas/BatchRequestAction"}}}},
    "BatchRequestAction":{"type":"object","required":["relative_path","method"],"properties":{"relative_path":{"type":"string"},"method":{"type":"string","enum":["get","post","put","delete","patch","head"]},"data":{"type":"object"}}},
    "BatchResponse":{"type":"object","properties":{"status_code":{"type":"integer"},"body":{"type":"object"}}}
  }},
  "paths":{"/batch":{"post":{"operationId":"createBatchRequest","requestBody":{"required":true,"content":{"application/json":{"schema":{"type":"object","properties":{"data":{"$ref":"#/components/schemas/BatchRequest"}}}}}},"responses":{"200":{"description":"ok","content":{"application/json":{"schema":{"type":"object","properties":{"data":{"type":"array","items":{"$ref":"#/components/schemas/BatchResponse"}}}}}}}}}}}
}`)
	raw := sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{{
		ID: "primary", Path: "/batch", Method: "POST", OperationID: "createBatchRequest", Artifact: artifact,
	}})
	var lock map[string]any
	if err := json.Unmarshal(raw, &lock); err != nil {
		t.Fatal(err)
	}
	lock["rest"].(map[string]any)["batch_action_inventory"] = map[string]any{
		"source_document":             "primary",
		"source_operation":            "fixture.rest.primary.shared",
		"request_schema":              sourceImportEventSchemaSelectorWire("BatchRequest"),
		"action_schema":               sourceImportEventSchemaSelectorWire("BatchRequestAction"),
		"response_schema":             sourceImportEventSchemaSelectorWire("BatchResponse"),
		"max_actions":                 10,
		"max_actions_source_location": `tags["Batch API"].description`,
		"provider_methods":            []any{"get", "post", "put", "delete", "patch", "head"},
		"request_envelope_field":      "data",
		"request_actions_field":       "actions",
		"action_method_field":         "method",
		"action_path_field":           "relative_path",
		"action_data_field":           "data",
		"response_envelope_field":     "data",
		"response_status_field":       "status_code",
	}
	encoded, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	return encoded, artifact
}
