package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
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

func TestOperationDirectWriteStructuredRESTBodyStopsAtDeclaredBoundsBeforeIO(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		calls++
	}))
	t.Cleanup(server.Close)

	request := structuredRESTBodyRequest()
	request.Body["targets"] = make([]any, maxStructuredRESTBodyItems+1)
	_, err := OperationDirectWrite(context.Background(), structuredRESTBodyBundle(server.URL), request, nil)
	if err == nil || !strings.Contains(err.Error(), "maxItems") {
		t.Fatalf("OperationDirectWrite error = %v, want declared array bound", err)
	}
	if calls != 0 {
		t.Fatalf("over-bound body reached transport; calls = %d, want 0", calls)
	}
}

func TestOperationDirectWriteBindsResolvedHeadersBeforeApproval(t *testing.T) {
	const secret = "header-secret-canary"
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer "+secret {
			t.Fatal("request did not receive the preview-bound declared header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	bundle := structuredRESTBodyBundle(server.URL)
	bundle.HTTP.Headers = map[string]string{"Authorization": "Bearer {{ secrets.token }}"}
	request := structuredRESTBodyRequest()
	request.Config.Secrets = map[string]string{"token": secret}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	prepared, err := prepareOperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("prepareOperationDirectWrite: %v", err)
	}
	raw, err := json.Marshal(prepared.prepared.Requests[0])
	if err != nil {
		t.Fatalf("marshal prepared request: %v", err)
	}
	if strings.Contains(string(raw), secret) {
		t.Fatal("prepared request serialization exposed a resolved header secret")
	}

	request.Approval = approvedEvidenceForPreview(t, preview)
	request.PreviewDigest = preview.Digest
	bundle.HTTP.Headers = map[string]string{"Authorization": "Bearer {{ secrets.token | unix_seconds }}"}
	if _, err := OperationDirectWrite(context.Background(), bundle, request, nil); err == nil {
		t.Fatal("invalid declared header was accepted after preview")
	}
	if calls != 0 {
		t.Fatalf("invalid header consumed approval or reached transport; calls = %d", calls)
	}

	bundle.HTTP.Headers = map[string]string{"Authorization": "Bearer {{ secrets.token }}"}
	request.Config.Secrets["token"] = "header-secret-after-preview"
	if _, err := OperationDirectWrite(context.Background(), bundle, request, nil); err == nil || !strings.Contains(strings.ToLower(err.Error()), "preview") {
		t.Fatalf("changed header OperationDirectWrite error = %v, want preview rejection", err)
	}
	if calls != 0 {
		t.Fatalf("changed header reached transport; calls = %d", calls)
	}

	request.Config.Secrets["token"] = secret
	if _, err := OperationDirectWrite(context.Background(), bundle, request, nil); err != nil {
		t.Fatalf("OperationDirectWrite with original bound header: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want one bound request", calls)
	}
}

func TestOperationDirectWriteBindsStaticHTTPMutationsBeforeApproval(t *testing.T) {
	const token = "api-key-canary"
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if got := r.Header.Get("X-API-Key"); got != token {
			t.Errorf("X-API-Key = %q, want preview-bound value", got)
		}
		if got := r.Header.Get("User-Agent"); got != "acme-direct-write/1.0" {
			t.Errorf("User-Agent = %q, want preview-bound value", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	bundle := structuredRESTBodyBundle(server.URL)
	bundle.HTTP.UserAgent = "acme-direct-write/1.0"
	bundle.HTTP.Auth = []AuthSpec{{Mode: "api_key_header", Header: "X-API-Key", Value: "{{ secrets.api_key }}"}}
	request := structuredRESTBodyRequest()
	request.Config.Secrets = map[string]string{"api_key": token}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	prepared, err := prepareOperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("prepareOperationDirectWrite: %v", err)
	}
	if got := prepared.prepared.Requests[0].Headers["X-Api-Key"]; got != token {
		t.Fatalf("prepared API key header = %q, want exact bound value", got)
	}
	if got := prepared.prepared.Requests[0].Headers["User-Agent"]; got != "acme-direct-write/1.0" {
		t.Fatalf("prepared user agent = %q, want exact bound value", got)
	}
	request.Approval = approvedEvidenceForPreview(t, preview)
	request.PreviewDigest = preview.Digest
	if _, err := OperationDirectWrite(context.Background(), bundle, request, nil); err != nil {
		t.Fatalf("OperationDirectWrite: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want one bound request", calls)
	}
}

func TestOperationDirectWriteBindsStaticQueryAuthBeforeApproval(t *testing.T) {
	const token = "query-key-canary"
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if got := r.URL.Query().Get("api_key"); got != token {
			t.Error("api_key did not match the preview-bound authentication value")
		}
		if got := r.Header.Get("X-Auth"); got != token {
			t.Error("header did not use the preview-bound authentication query value")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)
	bundle := structuredRESTBodyBundle(server.URL)
	bundle.Operations[0].REST.Parameters = append(bundle.Operations[0].REST.Parameters, OperationParameter{Name: "api_key", In: "query", Type: "string", Required: true})
	bundle.HTTP.Headers["X-Auth"] = "{{ query.api_key }}"
	bundle.HTTP.Auth = []AuthSpec{{Mode: "api_key_query", Param: "api_key", Value: "{{ secrets.api_key }}"}}
	request := structuredRESTBodyRequest()
	request.Config.Secrets = map[string]string{"api_key": token}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	prepared, err := prepareOperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("prepareOperationDirectWrite: %v", err)
	}
	if got := prepared.query.Get("api_key"); got != token {
		t.Fatalf("prepared api_key = %q, want exact bound value", got)
	}
	if got := prepared.prepared.Requests[0].Headers["X-Auth"]; got != token {
		t.Fatal("prepared header did not bind the final authentication query value")
	}
	request.Approval = approvedEvidenceForPreview(t, preview)
	request.PreviewDigest = preview.Digest
	if _, err := OperationDirectWrite(context.Background(), bundle, request, nil); err != nil {
		t.Fatalf("OperationDirectWrite: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want one bound request", calls)
	}
	collision := structuredRESTBodyRequest()
	collision.Config.Secrets = map[string]string{"api_key": token}
	collision.Query["api_key"] = "caller-value"
	if _, err := PreviewOperationDirectWrite(context.Background(), bundle, collision, nil); err == nil || !strings.Contains(err.Error(), "collides") {
		t.Fatal("caller API key query override did not fail before transport")
	}
	if calls != 1 {
		t.Fatalf("caller API key collision reached provider; calls = %d", calls)
	}
}

func TestOperationDirectWritePlanTransformsCanonicalizeTypedStringArrays(t *testing.T) {
	bundle := structuredRESTBodyBundle("https://example.invalid")
	rest := *bundle.Operations[0].REST
	rest.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"properties":{"tags":{"type":"array","maxItems":2,"items":{"type":"string"}}}
	}`)
	bundle.Operations[0].REST = &rest
	body := map[string]any{"tags": []string{"typed-array-canary"}}

	withheld, fields, err := WithholdOperationDirectWriteBodyFields(bundle, bundle.Operations[0].ID, body, []string{"body.tags.0"})
	if err != nil {
		t.Fatalf("WithholdOperationDirectWriteBodyFields: %v", err)
	}
	if !reflect.DeepEqual(fields, []string{"tags.0"}) {
		t.Fatalf("withheld fields = %#v, want [tags.0]", fields)
	}
	withheldTags, ok := withheld["tags"].([]any)
	if !ok || len(withheldTags) != 1 || withheldTags[0] != nil {
		t.Fatalf("withheld typed array = %#v, want canonical array with withheld element", withheld["tags"])
	}

	redacted, err := RedactOperationDirectWriteBodyFields(bundle, bundle.Operations[0].ID, body, []string{"body.tags.0"})
	if err != nil {
		t.Fatalf("RedactOperationDirectWriteBodyFields: %v", err)
	}
	redactedTags, ok := redacted["tags"].([]any)
	if !ok || !reflect.DeepEqual(redactedTags, []any{"redacted"}) {
		t.Fatalf("redacted typed array = %#v, want canonical redacted array", redacted["tags"])
	}
}

func TestPreflightOperationDirectWriteRejectsConditionalAPIKeyQueryOwnership(t *testing.T) {
	bundle := structuredRESTBodyBundle("https://example.invalid")
	rest := *bundle.Operations[0].REST
	rest.Parameters = append(rest.Parameters, OperationParameter{Name: "api_key", In: "query", Type: "string", Required: true})
	bundle.Operations[0].REST = &rest
	bundle.HTTP.Auth = []AuthSpec{
		{Mode: "api_key_query", Param: "api_key", Value: "key", When: "{{ config.auth_type == 'api' }}"},
		{Mode: "bearer", Token: "token", When: "{{ config.auth_type == 'bearer' }}"},
		{Mode: "none"},
	}
	operation := bundle.Operations[0]
	if err := PreflightOperationDirectWrite(bundle, operation.ID, operation.REST.Method, operation.REST.Path, operation.OutputPolicy); err == nil || !strings.Contains(err.Error(), "conditionally supplied") {
		t.Fatalf("conditional API key ownership error = %v, want declaration rejection", err)
	}

	bundle.HTTP.Auth = []AuthSpec{
		{Mode: "api_key_query", Param: "api_key", Value: "first", When: "{{ config.auth_type == 'first' }}"},
		{Mode: "api_key_query", Param: "api_key", Value: "second"},
	}
	if err := PreflightOperationDirectWrite(bundle, operation.ID, operation.REST.Method, operation.REST.Path, operation.OutputPolicy); err != nil {
		t.Fatalf("all-auth-path API key ownership: %v", err)
	}
	if err := PreflightOperationDirectWrite(bundle, operation.ID, operation.REST.Method, operation.REST.Path, operation.OutputPolicy, "api_key"); err == nil || !strings.Contains(err.Error(), "owned") {
		t.Fatalf("caller API key collision error = %v, want declaration rejection", err)
	}
}

func TestOperationDirectWriteRedactsNestedStructuredBodyValuesInSystemErrors(t *testing.T) {
	const token = "nested-sensitive-canary"
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal("decode request body")
		}
		targets, ok := body["targets"].([]any)
		if !ok || len(targets) != 1 {
			t.Error("provider did not receive the declared nested target")
		} else if target, ok := targets[0].(map[string]any); !ok || target["token"] != token {
			t.Error("provider did not receive the reconstituted nested token")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": token})
	}))
	t.Cleanup(server.Close)
	bundle := structuredRESTBodyBundle(server.URL)
	bundle.Operations[0].REST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["targets"],
		"properties":{"targets":{"type":"array","minItems":1,"maxItems":1,"items":{"type":"object","additionalProperties":false,"required":["id","token"],"properties":{"id":{"type":"string"},"token":{"type":"string"}}}}}
	}`)
	bundle.Operations[0].SensitivePolicy = &SensitivePolicySpec{RedactFields: []string{"body.targets.0.token"}}
	request := structuredRESTBodyRequest()
	request.Body = map[string]any{"targets": []any{map[string]any{"id": "target-1", "token": token}}}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	request.Approval = approvedEvidenceForPreview(t, preview)
	request.PreviewDigest = preview.Digest
	_, err = OperationDirectWrite(context.Background(), bundle, request, nil)
	if err == nil {
		t.Fatal("OperationDirectWrite error = nil, want provider error")
	}
	if strings.Contains(err.Error(), token) {
		t.Fatal("system-generated direct-write error leaked the nested sensitive value")
	}
	var httpErr *connsdk.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatal("provider response cause was not retained")
	}
	if !strings.Contains(httpErr.Body, token) {
		t.Fatal("provider response cause lost the echoed provider body")
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want one provider request", calls)
	}
}

func TestOperationDirectWriteRejectsStaticHTTPMutationConflictsBeforeIO(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*Bundle)
		want   string
	}{
		{
			name: "declared header collision",
			mutate: func(bundle *Bundle) {
				bundle.HTTP.Headers = map[string]string{"X-API-Key": "fixed"}
				bundle.HTTP.Auth = []AuthSpec{{Mode: "api_key_header", Header: "X-API-Key", Value: "key"}}
			},
			want: "collides",
		},
		{
			name: "user agent collision",
			mutate: func(bundle *Bundle) {
				bundle.HTTP.UserAgent = "acme/1.0"
				bundle.HTTP.Headers = map[string]string{"User-Agent": "provider-agent"}
			},
			want: "collides",
		},
		{
			name: "query collision",
			mutate: func(bundle *Bundle) {
				bundle.Operations[0].REST.Parameters = append(bundle.Operations[0].REST.Parameters, OperationParameter{Name: "api_key", In: "query", Type: "string"})
				bundle.Operations[0].REST.Query = map[string]string{"api_key": "fixed"}
				bundle.HTTP.Auth = []AuthSpec{{Mode: "api_key_query", Param: "api_key", Value: "key"}}
			},
			want: "collides",
		},
		{
			name: "invalid auth header",
			mutate: func(bundle *Bundle) {
				bundle.HTTP.Auth = []AuthSpec{{Mode: "api_key_header", Header: "X Invalid", Value: "key"}}
			},
			want: "invalid header",
		},
		{
			name: "invalid auth reference",
			mutate: func(bundle *Bundle) {
				bundle.HTTP.Auth = []AuthSpec{{Mode: "api_key_header", Header: "X-API-Key", Value: "{{ secrets.api_key.extra }}"}}
			},
			want: "malformed reference",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
			t.Cleanup(server.Close)
			bundle := structuredRESTBodyBundle(server.URL)
			test.mutate(&bundle)
			if _, err := PreviewOperationDirectWrite(context.Background(), bundle, structuredRESTBodyRequest(), nil); err == nil || !strings.Contains(strings.ToLower(err.Error()), test.want) {
				t.Fatalf("PreviewOperationDirectWrite error = %v, want %q", err, test.want)
			}
			if calls != 0 {
				t.Fatalf("invalid static HTTP mutation reached provider; calls = %d", calls)
			}
		})
	}
}

func TestOperationDirectWriteRejectsMalformedBaseURLTemplateBeforeIO(t *testing.T) {
	for _, template := range []string{"{{ config.base_url.extra }}", "{{ query.forbidden }}"} {
		t.Run(template, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
			t.Cleanup(server.Close)
			bundle := structuredRESTBodyBundle(server.URL + "/" + template)
			if _, err := PreviewOperationDirectWrite(context.Background(), bundle, structuredRESTBodyRequest(), nil); err == nil || !strings.Contains(err.Error(), "declared base URL") {
				t.Fatalf("PreviewOperationDirectWrite error = %v, want strict base URL template rejection", err)
			}
			if calls != 0 {
				t.Fatalf("malformed base URL reached provider; calls = %d", calls)
			}
		})
	}
}

func TestOperationDirectWriteRejectsOversizeErrorBodyWithoutPresentingPrefix(t *testing.T) {
	const canary = "oversize-provider-body-canary"
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("X-Provider-Trace", "oversize")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(canary + strings.Repeat("x", 2048)))
	}))
	t.Cleanup(server.Close)
	bundle := structuredRESTBodyBundle(server.URL)
	request := structuredRESTBodyRequest()
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	request.Approval = approvedEvidenceForPreview(t, preview)
	request.PreviewDigest = preview.Digest
	_, err = OperationDirectWrite(context.Background(), bundle, request, nil)
	if err == nil || !strings.Contains(err.Error(), "exceeds declared limit") {
		t.Fatalf("OperationDirectWrite error = %v, want explicit over-limit failure", err)
	}
	if strings.Contains(err.Error(), canary) {
		t.Fatal("oversize provider response prefix was presented")
	}
	var httpErr *connsdk.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("OperationDirectWrite cause = %T, want HTTPError", err)
	}
	if httpErr.Body != "" || httpErr.Status != http.StatusInternalServerError || httpErr.Header.Get("X-Provider-Trace") != "oversize" {
		t.Fatalf("oversize HTTP cause = %#v, want status and headers without a truncated body", httpErr)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want one provider response", calls)
	}
}

func TestOperationDirectWriteErrorsDoNotExposeResolvedURLValues(t *testing.T) {
	const secret = "url-secret-canary"
	request := structuredRESTBodyRequest()
	request.Config.Secrets = map[string]string{"token": secret}

	invalid := structuredRESTBodyBundle("http://%zz{{ secrets.token }}")
	if _, err := PreviewOperationDirectWrite(context.Background(), invalid, request, nil); err == nil {
		t.Fatal("invalid resolved URL was accepted")
	} else if strings.Contains(err.Error(), secret) {
		t.Fatal("invalid URL error exposed a resolved secret")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("provider failure detail"))
	}))
	t.Cleanup(server.Close)
	bundle := structuredRESTBodyBundle(server.URL + "/{{ secrets.token }}")
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	request.Approval = approvedEvidenceForPreview(t, preview)
	request.PreviewDigest = preview.Digest
	_, err = OperationDirectWrite(context.Background(), bundle, request, nil)
	if err == nil {
		t.Fatal("provider HTTP failure was accepted")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal("HTTP failure exposed a resolved secret")
	}
	if !strings.Contains(err.Error(), "provider failure detail") {
		t.Fatalf("HTTP failure = %v, want provider response body", err)
	}
}

func TestValidateOperationDirectWriteCLIFlagsUsesCanonicalStructuredSchema(t *testing.T) {
	base := structuredRESTBodyBundle("https://example.invalid").Operations[0]
	valid := []CLIFlag{
		{Name: "label", Type: "string", MapsTo: "body.label", Required: true},
		{Name: "attributes", Type: "json", MapsTo: "body.attributes", Required: true},
		{Name: "targets", Type: "json", MapsTo: "body.targets", Required: true},
	}
	if err := ValidateOperationDirectWriteCLIFlags(base, valid); err != nil {
		t.Fatalf("ValidateOperationDirectWriteCLIFlags valid schema: %v", err)
	}

	wrongType := append([]CLIFlag(nil), valid...)
	wrongType[1].Type = "string"
	if err := ValidateOperationDirectWriteCLIFlags(base, wrongType); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("ValidateOperationDirectWriteCLIFlags wrong object type = %v, want mismatch", err)
	}

	inferred := base
	inferredREST := *base.REST
	inferredREST.BodySchema = json.RawMessage(`{
		"additionalProperties": false,
		"required": ["settings"],
		"properties": {
			"settings": {"additionalProperties": false, "required": ["enabled"], "properties": {"enabled": {"type": "boolean"}}}
		}
	}`)
	inferred.REST = &inferredREST
	if err := ValidateOperationDirectWriteCLIFlags(inferred, []CLIFlag{{Name: "settings", Type: "json", MapsTo: "body.settings", Required: true}}); err != nil {
		t.Fatalf("ValidateOperationDirectWriteCLIFlags inferred object: %v", err)
	}

	prefix := base
	prefixREST := *base.REST
	prefixREST.BodySchema = json.RawMessage(`{
		"type": "object",
		"additionalProperties": false,
		"required": ["values"],
		"properties": {
			"values": {"maxItems": 2, "prefixItems": [{"type": "string"}, {"type": "integer"}], "items": {"type": "string"}}
		}
	}`)
	prefix.REST = &prefixREST
	if err := ValidateOperationDirectWriteCLIFlags(prefix, []CLIFlag{{Name: "values", Type: "string_array", MapsTo: "body.values", Required: true}}); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("ValidateOperationDirectWriteCLIFlags prefixItems mismatch = %v, want mismatch", err)
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
		{name: "scalar dotted traversal", op: first, field: "attributes.owner", want: "must be an object or array"},
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

type structuredRESTBodyMarshaler struct{}

func (structuredRESTBodyMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(`{"unknown":"value"}`), nil
}

func TestOperationDirectWriteStructuredRESTBodyRejectsUnsupportedTypedContainersBeforeIO(t *testing.T) {
	for _, tc := range []struct {
		name    string
		value   func() any
		wantErr string
	}{
		{
			name: "pointer map",
			value: func() any {
				value := map[string]any{"owner": "owner-1", "active": true, "undeclared": "value"}
				return &value
			},
			wantErr: "additional property",
		},
		{name: "record map alias", value: func() any { return connectors.Record{"owner": "owner-1", "active": true, "undeclared": "value"} }, wantErr: "additional property"},
		{name: "custom marshaler", value: func() any { return structuredRESTBodyMarshaler{} }, wantErr: "custom JSON marshalers"},
		{name: "struct", value: func() any { return struct{ Value string }{Value: "value"} }, wantErr: "struct values"},
		{
			name: "cyclic map",
			value: func() any {
				value := map[string]any{}
				value["self"] = value
				return value
			},
			wantErr: "depth limit",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				calls++
			}))
			t.Cleanup(server.Close)

			request := structuredRESTBodyRequest()
			request.Body = cloneAnyMap(request.Body)
			request.Body["attributes"] = tc.value()
			_, err := OperationDirectWrite(context.Background(), structuredRESTBodyBundle(server.URL), request, nil)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("OperationDirectWrite error = %v, want %q", err, tc.wantErr)
			}
			if calls != 0 {
				t.Fatalf("rejected typed container reached provider; calls = %d", calls)
			}
		})
	}
}

func TestOperationDirectWriteValidatesPathAndBodyBindingsAgainstDeclaration(t *testing.T) {
	op := structuredRESTBodyBundle("https://example.invalid").Operations[0]
	for _, tc := range []struct {
		name       string
		pathFields []string
		bodyFields []string
		wantErr    string
	}{
		{name: "declared path and nested body", pathFields: []string{"workspace_id"}, bodyFields: []string{"attributes.owner", "targets.0.id"}},
		{name: "undeclared path", pathFields: []string{"override"}, wantErr: "not declared by rest.path"},
		{name: "undeclared body", bodyFields: []string{"undeclared"}, wantErr: "additional property"},
		{name: "scalar descent", bodyFields: []string{"label.value"}, wantErr: "descends into scalar"},
		{name: "array index exceeds declared bound", bodyFields: []string{"targets.2.id"}, wantErr: "exceeds declared maxItems"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOperationDirectWriteMappings(op, tc.pathFields, tc.bodyFields)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateOperationDirectWriteMappings: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("ValidateOperationDirectWriteMappings error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestMaterializeOperationDirectWriteBodyMappingsUsesDeclaredSchemaPaths(t *testing.T) {
	bundle := structuredRESTBodyBundle("https://example.invalid")
	op := bundle.Operations[0]
	rest := *op.REST
	rest.BodySchema = json.RawMessage(`{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"0": {"type": "string"},
			"a.b": {"type": "string"},
			"config.v1": {"type": "object", "additionalProperties": false, "properties": {"enabled": {"type": "boolean"}}},
			"targets": {
				"type": "array",
				"maxItems": 1024,
				"items": {
					"type": "object",
					"additionalProperties": false,
					"required": ["id"],
					"properties": {"id": {"type": "string"}}
				}
			}
		}
	}`)
	op.REST = &rest
	bundle.Operations[0] = op

	mappings := map[string]any{"0": "zero", "a.b": "dotted"}
	for index := 0; index < 130; index++ {
		mappings[fmt.Sprintf("targets.%d.id", index)] = fmt.Sprintf("id-%d", index)
	}
	body, err := MaterializeOperationDirectWriteBodyMappings(bundle, op.ID, mappings)
	if err != nil {
		t.Fatalf("MaterializeOperationDirectWriteBodyMappings: %v", err)
	}
	if body["0"] != "zero" || body["a.b"] != "dotted" {
		t.Fatalf("object-key mappings = %#v, want declared numeric and dotted keys", body)
	}
	if err := ValidateOperationStructuredJSONBodyField(op, "config.v1"); err != nil {
		t.Fatalf("dotted structured property preflight: %v", err)
	}
	targets, ok := body["targets"].([]any)
	if !ok || len(targets) != 130 {
		t.Fatalf("targets = %#v, want 130 declared array entries", body["targets"])
	}
	for _, index := range []int{0, 1, 10, 129} {
		entry, ok := targets[index].(map[string]any)
		if !ok || entry["id"] != fmt.Sprintf("id-%d", index) {
			t.Fatalf("targets[%d] = %#v, want declared indexed value", index, targets[index])
		}
	}

	_, err = MaterializeOperationDirectWriteBodyMappings(bundle, op.ID, map[string]any{"targets.1.id": "second"})
	if err == nil || !strings.Contains(err.Error(), "sparse array") {
		t.Fatalf("sparse array error = %v, want declared sparse-index rejection", err)
	}
}

func TestValidateOperationDirectWriteCLIFlagsProjectsRequiredArraysAndDomains(t *testing.T) {
	requiredArray := structuredRESTBodyBundle("https://example.invalid").Operations[0]
	rest := *requiredArray.REST
	rest.BodySchema = json.RawMessage(`{
		"type": "object",
		"additionalProperties": false,
		"required": ["targets"],
		"properties": {
			"targets": {
				"type": "array",
				"minItems": 1,
				"maxItems": 1,
				"items": {
					"type": "object",
					"additionalProperties": false,
					"required": ["id"],
					"properties": {"id": {"type": "string", "enum": ["safe"]}}
				}
			}
		}
	}`)
	requiredArray.REST = &rest
	validNested := []CLIFlag{{Name: "target-id", Type: "enum", Values: []string{"safe"}, MapsTo: "body.targets.0.id", Required: true}}
	if err := ValidateOperationDirectWriteCLIFlags(requiredArray, validNested); err != nil {
		t.Fatalf("ValidateOperationDirectWriteCLIFlags required nested array: %v", err)
	}
	invalidNested := append([]CLIFlag(nil), validNested...)
	invalidNested[0].Values = []string{"unsafe"}
	if err := ValidateOperationDirectWriteCLIFlags(requiredArray, invalidNested); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("invalid nested enum error = %v, want declared-domain rejection", err)
	}

	stringArray := requiredArray
	stringArrayREST := *requiredArray.REST
	stringArrayREST.BodySchema = json.RawMessage(`{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"tags": {"type": "array", "maxItems": 1, "items": {"type": "string", "enum": ["safe"]}}
		}
	}`)
	stringArray.REST = &stringArrayREST
	for _, test := range []struct {
		name string
		flag CLIFlag
		want bool
	}{
		{name: "bounded declared domain", flag: CLIFlag{Name: "tags", Type: "string_array", Values: []string{"safe"}, MapsTo: "body.tags", MaxItems: 1}, want: true},
		{name: "undeclared enum value", flag: CLIFlag{Name: "tags", Type: "string_array", Values: []string{"unsafe"}, MapsTo: "body.tags", MaxItems: 1}},
		{name: "overdeclared cardinality", flag: CLIFlag{Name: "tags", Type: "string_array", Values: []string{"safe"}, MapsTo: "body.tags", MaxItems: 2}},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateOperationDirectWriteCLIFlags(stringArray, []CLIFlag{test.flag})
			if test.want {
				if err != nil {
					t.Fatalf("ValidateOperationDirectWriteCLIFlags: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "does not match") {
				t.Fatalf("ValidateOperationDirectWriteCLIFlags error = %v, want declared-domain rejection", err)
			}
		})
	}
}

func TestValidateOperationDirectWriteCLIFlagsProjectsContainerLowerBounds(t *testing.T) {
	base := structuredRESTBodyBundle("https://example.invalid").Operations[0]
	tags := base
	tagsREST := *base.REST
	tagsREST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["tags"],
		"properties":{"tags":{"type":"array","minItems":1,"maxItems":1,"items":{"type":"string"}}}
	}`)
	tags.REST = &tagsREST
	if err := ValidateOperationDirectWriteCLIFlags(tags, []CLIFlag{{Name: "tag", Type: "string", MapsTo: "body.tags.0", Required: true}}); err != nil {
		t.Fatalf("ValidateOperationDirectWriteCLIFlags scalar lower-bound array: %v", err)
	}
	if err := ValidateOperationDirectWriteCLIFlags(tags, nil); err == nil || !strings.Contains(err.Error(), "body.tags.0") {
		t.Fatalf("scalar lower-bound array error = %v, want required indexed mapping", err)
	}

	properties := base
	propertiesREST := *base.REST
	propertiesREST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"minProperties":2,
		"properties":{"one":{"type":"string"},"two":{"type":"string"}}
	}`)
	properties.REST = &propertiesREST
	if err := ValidateOperationDirectWriteCLIFlags(properties, []CLIFlag{
		{Name: "one", Type: "string", MapsTo: "body.one", Required: true},
		{Name: "two", Type: "string", MapsTo: "body.two", Required: true},
	}); err != nil {
		t.Fatalf("ValidateOperationDirectWriteCLIFlags minProperties coverage: %v", err)
	}
	if err := ValidateOperationDirectWriteCLIFlags(properties, nil); err == nil || !strings.Contains(err.Error(), "minProperties") {
		t.Fatalf("minProperties coverage error = %v, want declaration-bound lower-bound rejection", err)
	}
}

func TestStructuredRESTBodyDeclarationSatisfiabilityAndStaticCoverage(t *testing.T) {
	base := structuredRESTBodyBundle("https://example.invalid").Operations[0]
	unsatisfiable := base
	unsatisfiableREST := *base.REST
	unsatisfiableREST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"required":["mode"],"properties":{}}`)
	unsatisfiable.REST = &unsatisfiableREST
	if err := ValidateOperationDirectWriteMappings(unsatisfiable, nil, nil); err == nil || !strings.Contains(err.Error(), "required property") {
		t.Fatalf("unsatisfiable required property error = %v, want declaration rejection", err)
	}

	invalidStatic := base
	invalidStaticREST := *base.REST
	invalidStaticREST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["settings"],
		"properties":{"settings":{"type":"object","additionalProperties":false,"properties":{"enabled":{"type":"boolean"}}}}
	}`)
	invalidStaticREST.Body = map[string]any{"settings": nil}
	invalidStatic.REST = &invalidStaticREST
	if err := ValidateOperationDirectWriteCLIFlags(invalidStatic, nil); err == nil || !strings.Contains(err.Error(), "does not match type") {
		t.Fatalf("invalid static body error = %v, want preflight schema rejection", err)
	}

	merge := base
	mergeREST := *base.REST
	mergeREST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["payload"],
		"properties":{"payload":{"type":"object","additionalProperties":false,"required":["fixed","name"],"properties":{"fixed":{"type":"string"},"name":{"type":"string"}}}}
	}`)
	mergeREST.Body = map[string]any{"payload": map[string]any{"fixed": "provider"}}
	merge.REST = &mergeREST
	bundle := structuredRESTBodyBundle("https://example.invalid")
	bundle.Operations[0] = merge
	dynamic, err := MaterializeOperationDirectWriteBodyMappings(bundle, merge.ID, map[string]any{"payload.name": "caller"})
	if err != nil {
		t.Fatalf("MaterializeOperationDirectWriteBodyMappings nested value: %v", err)
	}
	materialized, err := materializeStructuredRESTBody(merge, merge.REST.Body, dynamic)
	if err != nil {
		t.Fatalf("materializeStructuredRESTBody nested merge: %v", err)
	}
	payload, ok := materialized["payload"].(map[string]any)
	if !ok || payload["fixed"] != "provider" || payload["name"] != "caller" {
		t.Fatalf("nested static merge = %#v, want fixed provider data and caller field", materialized)
	}
	persisted, err := json.Marshal(dynamic)
	if err != nil {
		t.Fatalf("marshal dynamic body: %v", err)
	}
	var reopened map[string]any
	if err := json.Unmarshal(persisted, &reopened); err != nil {
		t.Fatalf("unmarshal dynamic body: %v", err)
	}
	reopenedPayload, ok := reopened["payload"].(map[string]any)
	if !ok {
		t.Fatalf("reopened dynamic body = %#v, want payload object", reopened)
	}
	if _, present := reopenedPayload["fixed"]; present {
		t.Fatalf("reopened dynamic body = %#v, must not persist static-body placeholders", reopened)
	}
	materialized, err = materializeStructuredRESTBody(merge, merge.REST.Body, reopened)
	if err != nil {
		t.Fatalf("materializeStructuredRESTBody persisted nested merge: %v", err)
	}
	payload, ok = materialized["payload"].(map[string]any)
	if !ok || payload["fixed"] != "provider" || payload["name"] != "caller" {
		t.Fatalf("persisted nested static merge = %#v, want fixed provider data and caller field", materialized)
	}
	if err := ValidateOperationDirectWriteCLIFlags(merge, []CLIFlag{{Name: "payload", Type: "json", MapsTo: "body.payload", Required: true}}); err != nil {
		t.Fatalf("ValidateOperationDirectWriteCLIFlags static container: %v", err)
	}
	containerDynamic, err := MaterializeOperationDirectWriteBodyMappings(bundle, merge.ID, map[string]any{"payload": map[string]any{"name": "caller"}})
	if err != nil {
		t.Fatalf("MaterializeOperationDirectWriteBodyMappings static container: %v", err)
	}
	materialized, err = materializeStructuredRESTBody(merge, merge.REST.Body, containerDynamic)
	if err != nil {
		t.Fatalf("materializeStructuredRESTBody static container: %v", err)
	}
	payload, ok = materialized["payload"].(map[string]any)
	if !ok || payload["fixed"] != "provider" || payload["name"] != "caller" {
		t.Fatalf("container static merge = %#v, want fixed provider data and caller field", materialized)
	}

	unmergeableArray := base
	unmergeableArrayREST := *base.REST
	unmergeableArrayREST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["tags"],
		"properties":{"tags":{"type":"array","minItems":2,"maxItems":2,"items":{"type":"string"}}}
	}`)
	unmergeableArrayREST.Body = map[string]any{"tags": []any{"provider"}}
	unmergeableArray.REST = &unmergeableArrayREST
	if err := ValidateOperationDirectWriteMappings(unmergeableArray, nil, []string{"tags"}); err == nil || !strings.Contains(err.Error(), "cannot merge fixed rest.body container") {
		t.Fatalf("unmergeable static array container error = %v, want declaration rejection", err)
	}

	_, err = materializeStructuredRESTBody(merge, merge.REST.Body, map[string]any{"payload": map[string]any{"fixed": "caller", "name": "caller"}})
	if err == nil || !strings.Contains(err.Error(), "cannot be caller-overridden") {
		t.Fatalf("container static leaf collision = %v, want fixed-field rejection", err)
	}
	if err := ValidateOperationDirectWriteMappings(merge, nil, []string{"payload.fixed"}); err == nil || !strings.Contains(err.Error(), "fixed rest.body") {
		t.Fatalf("static overlap error = %v, want declared fixed-field rejection", err)
	}

	arrayMerge := base
	arrayMergeREST := *base.REST
	arrayMergeREST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["targets"],
		"properties":{"targets":{"type":"array","minItems":1,"maxItems":1,"items":{"type":"object","additionalProperties":false,"required":["fixed","name"],"properties":{"fixed":{"type":"string"},"name":{"type":"string"}}}}}
	}`)
	arrayMergeREST.Body = map[string]any{"targets": []any{map[string]any{"fixed": "provider"}}}
	arrayMerge.REST = &arrayMergeREST
	if err := ValidateOperationDirectWriteMappings(arrayMerge, nil, []string{"targets.0.name"}); err != nil {
		t.Fatalf("ValidateOperationDirectWriteMappings array nested value: %v", err)
	}
	arrayBundle := structuredRESTBodyBundle("https://example.invalid")
	arrayBundle.Operations[0] = arrayMerge
	arrayDynamic, err := MaterializeOperationDirectWriteBodyMappings(arrayBundle, arrayMerge.ID, map[string]any{"targets.0.name": "caller"})
	if err != nil {
		t.Fatalf("MaterializeOperationDirectWriteBodyMappings array nested value: %v", err)
	}
	arrayMaterialized, err := materializeStructuredRESTBody(arrayMerge, arrayMerge.REST.Body, arrayDynamic)
	if err != nil {
		t.Fatalf("materializeStructuredRESTBody array nested merge: %v", err)
	}
	targets, ok := arrayMaterialized["targets"].([]any)
	if !ok || len(targets) != 1 {
		t.Fatalf("array nested static merge targets = %#v, want one target", arrayMaterialized)
	}
	target, ok := targets[0].(map[string]any)
	if !ok || target["fixed"] != "provider" || target["name"] != "caller" {
		t.Fatalf("array nested static merge = %#v, want fixed provider data and caller field", arrayMaterialized)
	}
	if err := ValidateOperationDirectWriteMappings(arrayMerge, nil, []string{"targets.0.fixed"}); err == nil || !strings.Contains(err.Error(), "fixed rest.body") {
		t.Fatalf("array static overlap error = %v, want declared fixed-field rejection", err)
	}

	arrayPrefix := base
	arrayPrefixREST := *base.REST
	arrayPrefixREST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["targets"],
		"properties":{"targets":{"type":"array","minItems":2,"maxItems":2,"prefixItems":[
			{"type":"object","additionalProperties":false,"required":["fixed"],"properties":{"fixed":{"type":"string"}}},
			{"type":"object","additionalProperties":false,"required":["name"],"properties":{"name":{"type":"string"}}}
		]}}
	}`)
	arrayPrefixREST.Body = map[string]any{"targets": []any{map[string]any{"fixed": "provider"}}}
	arrayPrefix.REST = &arrayPrefixREST
	if err := ValidateOperationDirectWriteMappings(arrayPrefix, nil, []string{"targets.1.name"}); err != nil {
		t.Fatalf("ValidateOperationDirectWriteMappings static array prefix: %v", err)
	}
	arrayPrefixBundle := structuredRESTBodyBundle("https://example.invalid")
	arrayPrefixBundle.Operations[0] = arrayPrefix
	arrayPrefixDynamic, err := MaterializeOperationDirectWriteBodyMappings(arrayPrefixBundle, arrayPrefix.ID, map[string]any{"targets.1.name": "caller"})
	if err != nil {
		t.Fatalf("MaterializeOperationDirectWriteBodyMappings static array prefix: %v", err)
	}
	arrayPrefixMaterialized, err := materializeStructuredRESTBody(arrayPrefix, arrayPrefix.REST.Body, arrayPrefixDynamic)
	if err != nil {
		t.Fatalf("materializeStructuredRESTBody static array prefix: %v", err)
	}
	arrayPrefixTargets, ok := arrayPrefixMaterialized["targets"].([]any)
	if !ok || len(arrayPrefixTargets) != 2 {
		t.Fatalf("static array prefix targets = %#v, want two targets", arrayPrefixMaterialized)
	}
	if first, ok := arrayPrefixTargets[0].(map[string]any); !ok || first["fixed"] != "provider" {
		t.Fatalf("static array prefix first target = %#v, want fixed provider value", arrayPrefixTargets[0])
	}
	if second, ok := arrayPrefixTargets[1].(map[string]any); !ok || second["name"] != "caller" {
		t.Fatalf("static array prefix second target = %#v, want caller value", arrayPrefixTargets[1])
	}
}

func TestStructuredRESTBodyRejectsUnreachableNodeMinimumAndFloatByteDrift(t *testing.T) {
	base := structuredRESTBodyBundle("https://example.invalid").Operations[0]
	properties := make(map[string]any, maxStructuredRESTBodyFields-1)
	required := make([]any, 0, maxStructuredRESTBodyFields-1)
	for index := 0; index < maxStructuredRESTBodyFields-1; index++ {
		name := fmt.Sprintf("field_%d", index)
		properties[name] = map[string]any{"type": "string"}
		required = append(required, name)
	}
	nodeSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []any{"targets"},
		"properties": map[string]any{
			"targets": map[string]any{
				"type":     "array",
				"minItems": float64(maxStructuredRESTBodyItems),
				"maxItems": float64(maxStructuredRESTBodyItems),
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             required,
					"properties":           properties,
				},
			},
		},
	}
	rawNodeSchema, err := json.Marshal(nodeSchema)
	if err != nil {
		t.Fatalf("marshal node schema: %v", err)
	}
	nodeLimited := base
	nodeLimitedREST := *base.REST
	nodeLimitedREST.BodySchema = rawNodeSchema
	nodeLimited.REST = &nodeLimitedREST
	if err := ValidateOperationDirectWriteMappings(nodeLimited, nil, nil); err == nil || !strings.Contains(err.Error(), "minimum valid body requires") {
		t.Fatalf("node-minimum error = %v, want unreachable declaration rejection", err)
	}

	floatLimited := base
	floatLimitedREST := *base.REST
	floatLimitedREST.MaxBytes = 17
	floatLimitedREST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"ratio":{"type":"number"}}}`)
	floatLimited.REST = &floatLimitedREST
	_, err = materializeStructuredRESTBody(floatLimited, nil, map[string]any{"ratio": 1e-6})
	if err == nil || !strings.Contains(err.Error(), "request body too large") || !strings.Contains(err.Error(), "body.ratio") {
		t.Fatalf("float byte-limit error = %v, want pre-copy JSON-compatible accounting", err)
	}

	byteLimited := base
	byteLimitedREST := *base.REST
	byteLimitedREST.MaxBytes = 1024
	byteLimitedREST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["mode"],
		"properties":{"mode":{"type":"string","enum":["` + strings.Repeat("x", 2048) + `"]}}
	}`)
	byteLimited.REST = &byteLimitedREST
	if err := ValidateOperationDirectWriteMappings(byteLimited, nil, nil); err == nil || !strings.Contains(err.Error(), "minimum valid body requires") || !strings.Contains(err.Error(), "bytes") {
		t.Fatalf("byte-minimum error = %v, want unreachable declaration rejection", err)
	}

	patternLimited := base
	patternLimitedREST := *base.REST
	patternLimitedREST.MaxBytes = 512
	patternLimitedREST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["code"],
		"properties":{"code":{"type":"string","pattern":"^.{1000}$"}}
	}`)
	patternLimited.REST = &patternLimitedREST
	if err := ValidateOperationDirectWriteMappings(patternLimited, nil, nil); err == nil || !strings.Contains(err.Error(), "minimum valid body requires") || !strings.Contains(err.Error(), "bytes") {
		t.Fatalf("pattern byte-minimum error = %v, want unreachable declaration rejection", err)
	}

	escapedPatternLimited := base
	escapedPatternLimitedREST := *base.REST
	escapedPatternLimitedREST.MaxBytes = 1024
	escapedPatternLimitedREST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["code"],
		"properties":{"code":{"type":"string","pattern":"^(?:\\x00){200}$"}}
	}`)
	escapedPatternLimited.REST = &escapedPatternLimitedREST
	if err := ValidateOperationDirectWriteMappings(escapedPatternLimited, nil, nil); err == nil || !strings.Contains(err.Error(), "minimum valid body requires") || !strings.Contains(err.Error(), "bytes") {
		t.Fatalf("escaped pattern byte-minimum error = %v, want unreachable declaration rejection", err)
	}

	wordBoundary := base
	wordBoundaryREST := *base.REST
	wordBoundaryREST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["code"],
		"properties":{"code":{"type":"string","pattern":"^\\b$"}}
	}`)
	wordBoundary.REST = &wordBoundaryREST
	if err := ValidateOperationDirectWriteMappings(wordBoundary, nil, nil); err == nil || !strings.Contains(err.Error(), "schema-valid string witness") {
		t.Fatalf("word-boundary satisfiability error = %v, want declaration rejection", err)
	}

	conflictingString := base
	conflictingStringREST := *base.REST
	conflictingStringREST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["endpoint"],
		"properties":{"endpoint":{"type":"string","pattern":"^x$","format":"uri"}}
	}`)
	conflictingString.REST = &conflictingStringREST
	if err := ValidateOperationDirectWriteMappings(conflictingString, nil, nil); err == nil || !strings.Contains(err.Error(), "schema-valid string witness") {
		t.Fatalf("pattern-format satisfiability error = %v, want declaration rejection", err)
	}

	compatibleString := base
	compatibleStringREST := *base.REST
	compatibleStringREST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["endpoint"],
		"properties":{"endpoint":{"type":"string","pattern":"^x://a$","format":"uri"}}
	}`)
	compatibleString.REST = &compatibleStringREST
	if err := ValidateOperationDirectWriteMappings(compatibleString, nil, nil); err != nil {
		t.Fatalf("compatible pattern-format declaration: %v", err)
	}

	completionLimited := base
	completionLimitedREST := *base.REST
	completionLimitedREST.MaxBytes = 1024
	completionLimitedREST.BodySchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["fixed","name"],
		"properties":{"fixed":{"type":"string"},"name":{"type":"string","enum":["x"]}}
	}`)
	completionLimitedREST.Body = map[string]any{"fixed": strings.Repeat("x", 1003)}
	completionLimited.REST = &completionLimitedREST
	if err := ValidateOperationDirectWriteMappings(completionLimited, nil, nil); err == nil || !strings.Contains(err.Error(), "completion") || !strings.Contains(err.Error(), "bytes") {
		t.Fatalf("merged static completion error = %v, want unreachable declaration rejection", err)
	}

	staticLimited := base
	staticLimitedREST := *base.REST
	staticLimitedREST.MaxBytes = 32
	staticLimitedREST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"fixed":{"type":"string"}}}`)
	staticLimitedREST.Body = map[string]any{"fixed": strings.Repeat("x", 64)}
	staticLimited.REST = &staticLimitedREST
	if err := ValidateOperationDirectWriteMappings(staticLimited, nil, nil); err == nil || !strings.Contains(err.Error(), "rest.body") {
		t.Fatalf("fixed body byte-limit error = %v, want declaration rejection", err)
	}
}

func TestOperationDirectWriteStrictHeadersAndProviderHeaders(t *testing.T) {
	request := structuredRESTBodyRequest()
	request.Config.Secrets = map[string]string{"token": "header-secret-canary"}
	for _, test := range []struct {
		name    string
		headers map[string]string
	}{
		{name: "reference tail", headers: map[string]string{"Authorization": "Bearer {{ secrets.token.extra }}"}},
		{name: "unclosed expression", headers: map[string]string{"Authorization": "Bearer {{ secrets.token"}},
		{name: "unbound query namespace", headers: map[string]string{"X-Trace": "{{ query.trace }}"}},
		{name: "unbound record namespace", headers: map[string]string{"X-Trace": "{{ record.trace }}"}},
		{name: "canonical duplicate", headers: map[string]string{"X-Mode": "one", "x-mode": "two"}},
		{name: "invalid name", headers: map[string]string{"X Mode": "one"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
			t.Cleanup(server.Close)
			bundle := structuredRESTBodyBundle(server.URL)
			bundle.HTTP.Headers = test.headers
			if _, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil); err == nil {
				t.Fatal("invalid declared header was accepted")
			}
			if calls != 0 {
				t.Fatalf("invalid header reached provider; calls = %d", calls)
			}
		})
	}

	t.Run("declared query namespace is bound", func(t *testing.T) {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			if got := r.Header.Get("X-Trace"); got != "trace-1" {
				t.Errorf("X-Trace = %q, want trace-1", got)
			}
			_, _ = w.Write([]byte(`{"ok":true}`))
		}))
		t.Cleanup(server.Close)
		bundle := structuredRESTBodyBundle(server.URL)
		rest := *bundle.Operations[0].REST
		rest.Parameters = append([]OperationParameter(nil), rest.Parameters...)
		rest.Parameters = append(rest.Parameters, OperationParameter{Name: "trace", In: "query", Type: "string"})
		bundle.Operations[0].REST = &rest
		bundle.HTTP.Headers = map[string]string{"X-Trace": "{{ query.trace }}"}
		request := structuredRESTBodyRequest()
		request.Query["trace"] = "trace-1"
		preview, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil)
		if err != nil {
			t.Fatalf("PreviewOperationDirectWrite: %v", err)
		}
		prepared, err := prepareOperationDirectWrite(context.Background(), bundle, request, nil)
		if err != nil {
			t.Fatalf("prepareOperationDirectWrite: %v", err)
		}
		if got := prepared.prepared.Requests[0].Headers["X-Trace"]; got != "trace-1" {
			t.Fatalf("prepared X-Trace = %q, want trace-1", got)
		}
		request.Approval = approvedEvidenceForPreview(t, preview)
		request.PreviewDigest = preview.Digest
		if _, err := OperationDirectWrite(context.Background(), bundle, request, nil); err != nil {
			t.Fatalf("OperationDirectWrite: %v", err)
		}
		if calls != 1 {
			t.Fatalf("calls = %d, want 1", calls)
		}
	})

	t.Run("successful response retains headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Add("X-Provider-Trace", "one")
			w.Header().Add("X-Provider-Trace", "two")
			_, _ = w.Write([]byte(`{"ok":true}`))
		}))
		t.Cleanup(server.Close)
		bundle := structuredRESTBodyBundle(server.URL)
		request := structuredRESTBodyRequest()
		preview, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil)
		if err != nil {
			t.Fatalf("PreviewOperationDirectWrite: %v", err)
		}
		request.Approval = approvedEvidenceForPreview(t, preview)
		request.PreviewDigest = preview.Digest
		result, err := OperationDirectWrite(context.Background(), bundle, request, nil)
		if err != nil {
			t.Fatalf("OperationDirectWrite: %v", err)
		}
		if got, want := result.Headers["X-Provider-Trace"], []string{"one", "two"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("response headers = %#v, want %#v", result.Headers, want)
		}
	})

	t.Run("failure retains headers with safe identity", func(t *testing.T) {
		const secret = "url-secret-canary"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Add("X-Provider-Trace", "failed")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(" provider failure body "))
		}))
		t.Cleanup(server.Close)
		bundle := structuredRESTBodyBundle(server.URL + "/{{ secrets.token }}")
		request := structuredRESTBodyRequest()
		request.Config.Secrets = map[string]string{"token": secret}
		preview, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil)
		if err != nil {
			t.Fatalf("PreviewOperationDirectWrite: %v", err)
		}
		request.Approval = approvedEvidenceForPreview(t, preview)
		request.PreviewDigest = preview.Digest
		_, err = OperationDirectWrite(context.Background(), bundle, request, nil)
		if err == nil {
			t.Fatal("provider failure was accepted")
		}
		var httpErr *connsdk.HTTPError
		if !errors.As(err, &httpErr) {
			t.Fatalf("error = %T %v, want retained HTTP error", err, err)
		}
		if strings.Contains(httpErr.URL, secret) || strings.Contains(err.Error(), secret) {
			t.Fatalf("provider failure exposed resolved URL secret: %v", err)
		}
		if got, want := httpErr.Header.Values("X-Provider-Trace"), []string{"failed"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("error headers = %#v, want %#v", httpErr.Header, want)
		}
		if httpErr.Body != " provider failure body " {
			t.Fatalf("error body = %q, want exact provider response", httpErr.Body)
		}
	})
}
