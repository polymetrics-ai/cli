package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func structuredRESTBodyBundle(baseURL string) Bundle {
	batchable := false
	return Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: baseURL, Headers: map[string]string{"X-Provider-Version": "2026-08-20"}},
		Operations: []OperationSpec{
			{
				ID:            "acme.workspaces.create_widget",
				Kind:          "rest_write",
				Summary:       "Create a widget",
				Risk:          "high",
				Approval:      "plan-preview-confirm-execute",
				OutputPolicy:  "json",
				MutationClass: "create",
				Confirmation:  &ConfirmationSpec{Kind: connectors.ConfirmationKindDestructive},
				Batchable:     &batchable,
				REST: &RESTOperationSpec{
					Method:      http.MethodPost,
					Path:        "/workspaces/{workspace_id}/widgets",
					ContentType: "application/json",
					MaxBytes:    1024,
					BodySchema: json.RawMessage(`{
						"type": "object",
						"additionalProperties": false,
						"required": ["label", "attributes", "targets"],
						"properties": {
							"label": {"type": "string"},
							"attributes": {
								"type": "object",
								"additionalProperties": false,
								"required": ["owner", "active"],
								"properties": {
									"owner": {"type": "string"},
									"active": {"type": "boolean"}
								}
							},
							"targets": {
								"type": "array",
								"minItems": 1,
								"maxItems": 2,
								"items": {
									"type": "object",
									"additionalProperties": false,
									"required": ["id"],
									"properties": {"id": {"type": "string"}}
								}
							}
						}
					}`),
				},
			},
			{
				ID:            "acme.workspaces.archive_widget",
				Kind:          "rest_write",
				Summary:       "Archive a widget",
				Risk:          "high",
				Approval:      "plan-preview-confirm-execute",
				OutputPolicy:  "json",
				MutationClass: "update",
				Confirmation:  &ConfirmationSpec{Kind: connectors.ConfirmationKindDestructive},
				Batchable:     &batchable,
				REST: &RESTOperationSpec{
					Method:      http.MethodPost,
					Path:        "/workspaces/{workspace_id}/widgets/archive",
					ContentType: "application/json",
					MaxBytes:    1024,
					BodySchema: json.RawMessage(`{
						"type": "object",
						"additionalProperties": false,
						"required": ["archive"],
						"properties": {
							"archive": {
								"type": "object",
								"additionalProperties": false,
								"required": ["reason"],
								"properties": {"reason": {"type": "string"}}
							}
						}
					}`),
				},
			},
		},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{
			{Method: http.MethodPost, Path: "/workspaces/{workspace_id}/widgets", Operation: &SurfaceOperation{Model: "write"}},
			{Method: http.MethodPost, Path: "/workspaces/{workspace_id}/widgets/archive", Operation: &SurfaceOperation{Model: "write"}},
		}},
	}
}

func structuredRESTBodyRequest() connectors.OperationDirectWriteRequest {
	return connectors.OperationDirectWriteRequest{
		Operation: "acme.workspaces.create_widget",
		Config: connectors.RuntimeConfig{
			CredentialRevision:  "fixture-credential-revision",
			ConfigurationDigest: "fixture-configuration-digest",
			WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
		},
		PathParams:   map[string]string{"workspace_id": "workspace-1"},
		Query:        map[string]string{"dry_run": "true"},
		OutputPolicy: "json",
		Body: map[string]any{
			"label": "fixture widget",
			"attributes": map[string]any{
				"owner":  "owner-1",
				"active": true,
			},
			"targets": []any{map[string]any{"id": "target-1"}},
		},
	}
}

func structuredRESTBodySchemaWithTarget(targetSchema string) json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"targets": ` + targetSchema + `
		}
	}`)
}

func TestOperationDirectWriteStructuredRESTBodyIsExactAndPreviewBound(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost || r.URL.Path != "/workspaces/workspace-1/widgets" {
			t.Errorf("request = %s %s, want POST /workspaces/workspace-1/widgets", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("dry_run"); got != "true" {
			t.Errorf("query dry_run = %q, want true", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("content type = %q, want application/json", got)
		}
		if got := r.Header.Get("X-Provider-Version"); got != "2026-08-20" {
			t.Errorf("provider-owned header = %q, want 2026-08-20", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		want := map[string]any{
			"label": "fixture widget",
			"attributes": map[string]any{
				"owner":  "owner-1",
				"active": true,
			},
			"targets": []any{map[string]any{"id": "target-1"}},
		}
		if !reflect.DeepEqual(body, want) {
			t.Errorf("body = %#v, want %#v", body, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	bundle := structuredRESTBodyBundle(server.URL)
	request := structuredRESTBodyRequest()
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	if calls != 0 {
		t.Fatalf("preview calls = %d, want 0", calls)
	}

	request.Approval = approvedEvidenceForPreview(t, preview)
	request.PreviewDigest = preview.Digest
	mutated := structuredRESTBodyRequest()
	mutated.Approval = request.Approval
	mutated.PreviewDigest = request.PreviewDigest
	mutated.Body["attributes"].(map[string]any)["owner"] = "owner-after-preview"
	if _, err := OperationDirectWrite(context.Background(), bundle, mutated, nil); err == nil || !strings.Contains(strings.ToLower(err.Error()), "preview") {
		t.Fatalf("mutated OperationDirectWrite error = %v, want preview-bound rejection", err)
	}
	if calls != 0 {
		t.Fatalf("mutated request reached transport; calls = %d, want 0", calls)
	}

	result, err := OperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("OperationDirectWrite: %v", err)
	}
	if result.Status != http.StatusOK || calls != 1 {
		t.Fatalf("result/calls = %+v/%d, want status 200 and one request", result, calls)
	}
}

func TestOperationDirectWriteStructuredRESTBodyRejectsInvalidInputBeforeIO(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*connectors.OperationDirectWriteRequest)
		wantErr string
	}{
		{
			name: "unknown nested field",
			mutate: func(req *connectors.OperationDirectWriteRequest) {
				req.Body["attributes"].(map[string]any)["unknown"] = "nope"
			},
			wantErr: "additional property",
		},
		{
			name: "missing required nested field",
			mutate: func(req *connectors.OperationDirectWriteRequest) {
				delete(req.Body["attributes"].(map[string]any), "owner")
			},
			wantErr: "required property missing",
		},
		{
			name: "wrong object type",
			mutate: func(req *connectors.OperationDirectWriteRequest) {
				req.Body["attributes"] = []any{}
			},
			wantErr: "does not match type",
		},
		{
			name: "array exceeds declaration limit",
			mutate: func(req *connectors.OperationDirectWriteRequest) {
				req.Body["targets"] = []any{
					map[string]any{"id": "target-1"},
					map[string]any{"id": "target-2"},
					map[string]any{"id": "target-3"},
				}
			},
			wantErr: "maxItems",
		},
		{
			name: "encoded payload exceeds declaration limit",
			mutate: func(req *connectors.OperationDirectWriteRequest) {
				req.Body["label"] = strings.Repeat("x", 1024)
			},
			wantErr: "request body too large",
		},
		{
			name: "cross action field",
			mutate: func(req *connectors.OperationDirectWriteRequest) {
				req.Body["archive"] = map[string]any{"reason": "wrong action"}
			},
			wantErr: "does not declare structured field",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				calls++
			}))
			t.Cleanup(server.Close)

			request := structuredRESTBodyRequest()
			request.Body = cloneAnyMap(request.Body)
			request.Body["attributes"] = cloneAnyMap(request.Body["attributes"].(map[string]any))
			test.mutate(&request)
			_, err := OperationDirectWrite(context.Background(), structuredRESTBodyBundle(server.URL), request, nil)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("OperationDirectWrite error = %v, want %q", err, test.wantErr)
			}
			if calls != 0 {
				t.Fatalf("invalid request reached transport; calls = %d, want 0", calls)
			}
		})
	}
}

func TestOperationDirectWriteStructuredRESTBodyRejectsUnconstrainedArrayDeclarationsBeforeIO(t *testing.T) {
	tests := []struct {
		name         string
		targetSchema string
		value        any
		wantErr      string
	}{
		{
			name:         "missing item schema",
			targetSchema: `{"type":"array","maxItems":1}`,
			value: []any{map[string]any{
				"undeclared": map[string]any{"nested": []any{}},
			}},
			wantErr: "items schema",
		},
		{
			name:         "empty item schema",
			targetSchema: `{"type":"array","maxItems":1,"items":{}}`,
			value: []any{map[string]any{
				"undeclared": map[string]any{"nested": []any{}},
			}},
			wantErr: "items schema",
		},
		{
			name:         "untyped nested item descendant",
			targetSchema: `{"type":"array","maxItems":1,"items":{"type":"object","additionalProperties":false,"properties":{"payload":{}}}}`,
			value: []any{map[string]any{
				"payload": map[string]any{"undeclared": map[string]any{"nested": []any{}}},
			}},
			wantErr: "explicit type",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				calls++
			}))
			t.Cleanup(server.Close)

			bundle := structuredRESTBodyBundle(server.URL)
			rest := *bundle.Operations[0].REST
			rest.BodySchema = structuredRESTBodySchemaWithTarget(test.targetSchema)
			bundle.Operations[0].REST = &rest
			request := structuredRESTBodyRequest()
			request.Body = map[string]any{
				"targets": test.value,
			}

			_, err := OperationDirectWrite(context.Background(), bundle, request, nil)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("OperationDirectWrite error = %v, want %q", err, test.wantErr)
			}
			if calls != 0 {
				t.Fatalf("unconstrained declaration reached transport; calls = %d, want 0", calls)
			}
		})
	}
}

func TestOperationStructuredJSONBodyPreflightRejectsSchemaAndActionMismatches(t *testing.T) {
	bundle := structuredRESTBodyBundle("https://example.invalid")
	first := bundle.Operations[0]
	second := bundle.Operations[1]
	if err := ValidateOperationStructuredJSONBodyField(first, "attributes"); err != nil {
		t.Fatalf("first structured field: %v", err)
	}
	for _, test := range []struct {
		name  string
		op    OperationSpec
		field string
		want  string
	}{
		{name: "field from another action", op: first, field: "archive", want: "does not declare"},
		{name: "first field on second action", op: second, field: "attributes", want: "does not declare"},
		{name: "dotted traversal", op: first, field: "attributes.owner", want: "top-level"},
		{
			name: "open nested object",
			op: func() OperationSpec {
				changed := first
				changed.REST = &RESTOperationSpec{
					Method:      http.MethodPost,
					Path:        first.REST.Path,
					ContentType: "application/json",
					MaxBytes:    1024,
					BodySchema:  json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"attributes":{"type":"object","properties":{"owner":{"type":"string"}}}}}`),
				}
				return changed
			}(),
			field: "attributes",
			want:  "additionalProperties",
		},
		{
			name: "unbounded array",
			op: func() OperationSpec {
				changed := first
				changed.REST = &RESTOperationSpec{
					Method:      http.MethodPost,
					Path:        first.REST.Path,
					ContentType: "application/json",
					MaxBytes:    1024,
					BodySchema:  json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"targets":{"type":"array","items":{"type":"string"}}}}`),
				}
				return changed
			}(),
			field: "targets",
			want:  "without maxItems",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateOperationStructuredJSONBodyField(test.op, test.field); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ValidateOperationStructuredJSONBodyField error = %v, want %q", err, test.want)
			}
		})
	}
}
