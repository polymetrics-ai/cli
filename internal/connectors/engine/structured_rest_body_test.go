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
					Parameters: []OperationParameter{{
						Name: "dry_run",
						In:   "query",
						Type: "boolean",
					}},
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
			wantErr: "additional property",
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

func TestOperationStructuredJSONBodyPreflightNormalizesUnambiguousTypes(t *testing.T) {
	base := structuredRESTBodyBundle("https://example.invalid")
	first := base.Operations[0]
	tests := []struct {
		name       string
		bodySchema json.RawMessage
		wantErr    string
	}{
		{
			name:       "explicit object type",
			bodySchema: structuredRESTBodySchemaWithTarget(`{"type":"object","additionalProperties":false,"properties":{}}`),
		},
		{
			name: "omitted root and nested object types",
			bodySchema: json.RawMessage(`{
				"additionalProperties": false,
				"properties": {
					"targets": {
						"additionalProperties": false,
						"properties": {
							"settings": {"additionalProperties": false, "properties": {}}
						}
					}
				}
			}`),
		},
		{
			name:       "omitted array type through items",
			bodySchema: structuredRESTBodySchemaWithTarget(`{"maxItems":1,"items":{"type":"string"}}`),
		},
		{
			name:       "omitted array type through prefix items",
			bodySchema: structuredRESTBodySchemaWithTarget(`{"maxItems":1,"prefixItems":[{"type":"string"}]}`),
		},
		{
			name:       "conflicting explicit type",
			bodySchema: structuredRESTBodySchemaWithTarget(`{"type":"array","maxItems":1,"items":{"type":"string"},"properties":{}}`),
			wantErr:    "foundation gap",
		},
		{
			name:       "ambiguous omitted type",
			bodySchema: structuredRESTBodySchemaWithTarget(`{"additionalProperties":false,"properties":{},"maxItems":1,"items":{"type":"string"}}`),
			wantErr:    "foundation gap",
		},
		{
			name:       "unconstrained prefix item tail",
			bodySchema: structuredRESTBodySchemaWithTarget(`{"maxItems":2,"prefixItems":[{"type":"string"}]}`),
			wantErr:    "foundation gap",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := first
			rest := *first.REST
			rest.BodySchema = test.bodySchema
			op.REST = &rest

			err := ValidateOperationStructuredJSONBodyField(op, "targets")
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateOperationStructuredJSONBodyField: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("ValidateOperationStructuredJSONBodyField error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func TestOperationDirectWriteStructuredRESTBodyMaterializesInferredTypes(t *testing.T) {
	tests := []struct {
		name         string
		targetSchema string
	}{
		{
			name: "items",
			targetSchema: `{
				"maxItems": 1,
				"items": {
					"additionalProperties": false,
					"required": ["settings"],
					"properties": {
						"settings": {"additionalProperties": false, "properties": {}}
					}
				}
			}`,
		},
		{
			name: "prefix items",
			targetSchema: `{
				"maxItems": 1,
				"prefixItems": [{
					"additionalProperties": false,
					"required": ["settings"],
					"properties": {
						"settings": {"additionalProperties": false, "properties": {}}
					}
				}]
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true}`))
			}))
			t.Cleanup(server.Close)

			bundle := structuredRESTBodyBundle(server.URL)
			rest := *bundle.Operations[0].REST
			rest.BodySchema = structuredRESTBodySchemaWithTarget(test.targetSchema)
			bundle.Operations[0].REST = &rest
			request := structuredRESTBodyRequest()
			request.Body = map[string]any{
				"targets": []any{map[string]any{"settings": map[string]any{}}},
			}

			preview, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil)
			if err != nil {
				t.Fatalf("PreviewOperationDirectWrite: %v", err)
			}
			request.Approval = approvedEvidenceForPreview(t, preview)
			request.PreviewDigest = preview.Digest
			if _, err := OperationDirectWrite(context.Background(), bundle, request, nil); err != nil {
				t.Fatalf("OperationDirectWrite: %v", err)
			}
			if calls != 1 {
				t.Fatalf("valid request calls = %d, want 1", calls)
			}

			for _, invalid := range []struct {
				name  string
				value any
				want  string
			}{
				{
					name:  "wrong nested object type",
					value: []any{map[string]any{"settings": []any{}}},
					want:  "does not match type",
				},
				{
					name:  "undeclared nested object field",
					value: []any{map[string]any{"settings": map[string]any{"undeclared": true}}},
					want:  "additional property",
				},
			} {
				t.Run(invalid.name, func(t *testing.T) {
					invalidRequest := structuredRESTBodyRequest()
					invalidRequest.Body = map[string]any{"targets": invalid.value}
					_, err := OperationDirectWrite(context.Background(), bundle, invalidRequest, nil)
					if err == nil || !strings.Contains(err.Error(), invalid.want) {
						t.Fatalf("OperationDirectWrite error = %v, want %q", err, invalid.want)
					}
					if calls != 1 {
						t.Fatalf("invalid request reached transport; calls = %d, want 1", calls)
					}
				})
			}
		})
	}
}

func TestOperationDirectWriteStructuredRESTBodyCanonicalizesTypedAliases(t *testing.T) {
	type recordSlice []connectors.Record
	type recordTuple [1]connectors.Record

	bundle := structuredRESTBodyBundle("https://example.invalid")
	rest := *bundle.Operations[0].REST
	rest.BodySchema = json.RawMessage(`{
		"additionalProperties": false,
		"required": ["attributes", "targets"],
		"properties": {
			"attributes": {
				"additionalProperties": false,
				"required": ["owner", "settings"],
				"properties": {
					"owner": {"type": "string"},
					"settings": {
						"additionalProperties": false,
						"required": ["active"],
						"properties": {"active": {"type": "boolean"}}
					}
				}
			},
			"targets": {
				"maxItems": 1,
				"items": {
					"additionalProperties": false,
					"required": ["id"],
					"properties": {"id": {"type": "string"}}
				}
			}
		}
	}`)
	bundle.Operations[0].REST = &rest

	valid := structuredRESTBodyRequest()
	valid.Body = map[string]any{
		"attributes": connectors.Record{
			"owner": "owner-1",
			"settings": connectors.Record{
				"active": true,
			},
		},
		"targets": recordSlice{{"id": "target-1"}},
	}
	prepared, err := prepareOperationDirectWrite(context.Background(), bundle, valid, nil)
	if err != nil {
		t.Fatalf("prepareOperationDirectWrite typed aliases: %v", err)
	}
	attributes, ok := prepared.body["attributes"].(map[string]any)
	if !ok || attributes["owner"] != "owner-1" {
		t.Fatalf("canonical attributes = %#v, want ordinary declared object", prepared.body["attributes"])
	}
	targets, ok := prepared.body["targets"].([]any)
	if !ok || len(targets) != 1 {
		t.Fatalf("canonical targets = %#v, want one ordinary array item", prepared.body["targets"])
	}

	for _, test := range []struct {
		name string
		body map[string]any
		want string
	}{
		{
			name: "top level alias undeclared property",
			body: map[string]any{
				"attributes": connectors.Record{"owner": "owner-1", "settings": connectors.Record{"active": true}, "undeclared": true},
				"targets":    recordSlice{{"id": "target-1"}},
			},
			want: "additional property",
		},
		{
			name: "nested alias undeclared property",
			body: map[string]any{
				"attributes": connectors.Record{"owner": "owner-1", "settings": connectors.Record{"active": true, "undeclared": true}},
				"targets":    recordSlice{{"id": "target-1"}},
			},
			want: "additional property",
		},
		{
			name: "scalar where inferred object is required",
			body: map[string]any{
				"attributes": "not an object",
				"targets":    recordSlice{{"id": "target-1"}},
			},
			want: "does not match type",
		},
		{
			name: "typed array alias undeclared item property",
			body: map[string]any{
				"attributes": connectors.Record{"owner": "owner-1", "settings": connectors.Record{"active": true}},
				"targets":    recordTuple{{"id": "target-1", "undeclared": true}},
			},
			want: "additional property",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := structuredRESTBodyRequest()
			req.Body = test.body
			_, err := prepareOperationDirectWrite(context.Background(), bundle, req, nil)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("prepareOperationDirectWrite error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestOperationDirectWriteQueryOwnership(t *testing.T) {
	bundle := structuredRESTBodyBundle("https://example.invalid")
	rest := *bundle.Operations[0].REST
	rest.Query = map[string]string{"fixed": "provider"}
	rest.Parameters = []OperationParameter{
		{Name: "scope", In: "query", Type: "string", Required: true},
		{Name: "dry_run", In: "query", Type: "boolean"},
		{Name: "mode", In: "query", Type: "string", Values: []string{"safe"}},
	}
	bundle.Operations[0].REST = &rest

	request := structuredRESTBodyRequest()
	request.Query = map[string]string{"scope": "workspace-1", "mode": "safe"}
	prepared, err := prepareOperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("prepareOperationDirectWrite: %v", err)
	}
	if got := prepared.query.Get("scope"); got != "workspace-1" {
		t.Fatalf("query scope = %q, want workspace-1", got)
	}
	if got := prepared.query.Get("mode"); got != "safe" {
		t.Fatalf("query mode = %q, want safe", got)
	}
	if got := prepared.query.Get("fixed"); got != "provider" {
		t.Fatalf("provider fixed query = %q, want provider", got)
	}
	if _, present := prepared.query["dry_run"]; present {
		t.Fatalf("optional absent query = %#v, want dry_run omitted", prepared.query)
	}

	for _, test := range []struct {
		name   string
		mutate func(*Bundle, *connectors.OperationDirectWriteRequest)
		want   string
	}{
		{
			name: "unknown caller query",
			mutate: func(_ *Bundle, req *connectors.OperationDirectWriteRequest) {
				req.Query["unknown"] = "value"
			},
			want: "not source-declared",
		},
		{
			name: "fixed caller query overlap",
			mutate: func(_ *Bundle, req *connectors.OperationDirectWriteRequest) {
				req.Query["fixed"] = "caller"
			},
			want: "fixed by rest.query",
		},
		{
			name: "malformed boolean",
			mutate: func(_ *Bundle, req *connectors.OperationDirectWriteRequest) {
				req.Query["dry_run"] = "not-a-bool"
			},
			want: "must be boolean",
		},
		{
			name: "malformed fixed source query",
			mutate: func(bundle *Bundle, _ *connectors.OperationDirectWriteRequest) {
				bundle.Operations[0].REST.Query["dry_run"] = "not-a-bool"
			},
			want: "must be boolean",
		},
		{
			name: "invalid enum",
			mutate: func(_ *Bundle, req *connectors.OperationDirectWriteRequest) {
				req.Query["mode"] = "unsafe"
			},
			want: "must be one of",
		},
		{
			name: "required query absent",
			mutate: func(_ *Bundle, req *connectors.OperationDirectWriteRequest) {
				req.Query = map[string]string{"mode": "safe"}
			},
			want: "requires query parameter",
		},
		{
			name: "duplicate source declaration",
			mutate: func(bundle *Bundle, _ *connectors.OperationDirectWriteRequest) {
				rest := *bundle.Operations[0].REST
				rest.Parameters = append(rest.Parameters, OperationParameter{Name: "scope", In: "query", Type: "string"})
				bundle.Operations[0].REST = &rest
			},
			want: "duplicates query parameter",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			testBundle := bundle
			testBundle.Operations = append([]OperationSpec(nil), bundle.Operations...)
			testRest := *bundle.Operations[0].REST
			testRest.Parameters = append([]OperationParameter(nil), testRest.Parameters...)
			testRest.Query = map[string]string{}
			for name, value := range bundle.Operations[0].REST.Query {
				testRest.Query[name] = value
			}
			testBundle.Operations[0].REST = &testRest
			testRequest := structuredRESTBodyRequest()
			testRequest.Query = map[string]string{}
			for name, value := range request.Query {
				testRequest.Query[name] = value
			}
			test.mutate(&testBundle, &testRequest)
			_, err := prepareOperationDirectWrite(context.Background(), testBundle, testRequest, nil)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("prepareOperationDirectWrite error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestStructuredRESTBodyPrefixItemsRemainPositionBound(t *testing.T) {
	bundle := structuredRESTBodyBundle("https://example.invalid")
	op := bundle.Operations[0]
	rest := *op.REST
	rest.BodySchema = json.RawMessage(`{
		"additionalProperties": false,
		"properties": {
			"targets": {
				"maxItems": 2,
				"prefixItems": [{"type": "string"}],
				"items": {"type": "boolean"}
			}
		}
	}`)
	op.REST = &rest
	compiled, err := compileStructuredRESTBodySchema(op)
	if err != nil {
		t.Fatalf("compileStructuredRESTBodySchema: %v", err)
	}
	for _, test := range []struct {
		name    string
		value   any
		wantErr string
	}{
		{name: "valid prefix and tail", value: map[string]any{"targets": []any{"head", true}}},
		{name: "wrong prefix type", value: map[string]any{"targets": []any{true, true}}, wantErr: "does not match type"},
		{name: "wrong tail type", value: map[string]any{"targets": []any{"head", "tail"}}, wantErr: "does not match type"},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := compiled.schema.Validate(test.value)
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("Validate error = %v, want %q", err, test.wantErr)
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
