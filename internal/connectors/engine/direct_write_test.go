package engine

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestOperationDirectWriteAuthModeNoneSkipsAuthenticatorConstruction(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("Authorization = %q, want empty", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	hooks := &fakeHooks{name: "acme", authErr: errors.New("custom authentication should not be constructed")}
	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{
			URL:  srv.URL,
			Auth: []AuthSpec{{Mode: "custom", Hook: "acme"}},
		},
		Operations: []OperationSpec{{
			ID:            "acme.auth-login",
			Kind:          "rest_write",
			Summary:       "Exchange credentials",
			Risk:          "low",
			Approval:      "none",
			OutputPolicy:  "json",
			MutationClass: "create",
			REST: &RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "/auth/login",
				ContentType: "application/json",
				AuthMode:    "none",
				MaxBytes:    1024,
				BodySchema:  json.RawMessage(`{"type":"object","required":["username","password"],"properties":{"username":{"type":"string"},"password":{"type":"string"}}}`),
			},
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
			Method:    http.MethodPost,
			Path:      "/auth/login",
			Operation: &SurfaceOperation{Model: "write_action"},
		}}},
	}
	req := connectors.OperationDirectWriteRequest{
		Operation: "acme.auth-login",
		Body:      map[string]any{"username": "fixture-user", "password": "fixture-password"},
	}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, hooks)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest

	result, err := OperationDirectWrite(context.Background(), bundle, req, hooks)
	if err != nil {
		t.Fatalf("OperationDirectWrite: %v", err)
	}
	if result.Status != http.StatusOK {
		t.Fatalf("status = %d, want %d", result.Status, http.StatusOK)
	}
	if hooks.authCalls != 0 {
		t.Fatalf("custom authenticator calls = %d, want 0", hooks.authCalls)
	}
}

func TestOperationDirectWriteResponseSensitiveBaseURLRequiresHTTPS(t *testing.T) {
	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: "https://data.example.invalid/v2"},
		Operations: []OperationSpec{{
			ID:                "acme.login",
			Kind:              "rest_write",
			Summary:           "Exchange credentials",
			Risk:              "low",
			Approval:          "none",
			OutputPolicy:      "json",
			MutationClass:     "create",
			ResponseSensitive: true,
			REST: &RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "/v2/users/login",
				BaseURL:     "{{ config.auth_url }}",
				ContentType: "application/json",
				AuthMode:    "none",
				MaxBytes:    1024,
				BodySchema:  json.RawMessage(`{"type":"object","required":["username","password"],"properties":{"username":{"type":"string"},"password":{"type":"string"}}}`),
			},
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
			Method:    http.MethodPost,
			Path:      "/v2/users/login",
			Operation: &SurfaceOperation{Model: "write_action"},
		}}},
	}
	for _, tt := range []struct {
		name      string
		authURL   string
		wantError string
	}{
		{name: "rejects cleartext config overlay", authURL: "http://auth.example.invalid/v2", wantError: "requires an HTTPS base URL"},
		{name: "rejects query config overlay", authURL: "https://auth.example.invalid/v2?tenant=fixture", wantError: "without a query or fragment"},
		{name: "rejects fragment config overlay", authURL: "https://auth.example.invalid/v2#tenant", wantError: "without a query or fragment"},
		{name: "accepts HTTPS config overlay", authURL: "https://auth.example.invalid/v2"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PreviewOperationDirectWrite(context.Background(), bundle, connectors.OperationDirectWriteRequest{
				Operation: "acme.login",
				Config:    connectors.RuntimeConfig{Config: map[string]string{"auth_url": tt.authURL}},
				Body:      map[string]any{"username": "fixture-user", "password": "fixture-password"},
			}, nil)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("PreviewOperationDirectWrite error = %v, want %q", err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("PreviewOperationDirectWrite: %v", err)
			}
		})
	}
}

func TestOperationDirectWritePreviewsApprovesAndExecutesSingleFormRequest(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/vote" {
			t.Fatalf("path = %s, want /api/vote", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("id"); got != "t3_abc" {
			t.Fatalf("form id = %q, want t3_abc", got)
		}
		if got := r.Form.Get("dir"); got != "1" {
			t.Fatalf("form dir = %q, want 1", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"token":"server-token","nested":{"token":"nested-server-token"}}`))
	}))
	defer srv.Close()

	batchable := false
	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID:            "acme.vote",
			Kind:          "rest_write",
			Summary:       "Vote on a post",
			Risk:          "high",
			Approval:      "plan-preview-confirm-execute",
			OutputPolicy:  "json_redacted",
			MutationClass: "destructive",
			Confirmation:  &ConfirmationSpec{Kind: "destructive"},
			SensitivePolicy: &SensitivePolicySpec{
				RedactFields: []string{"nested.token"},
			},
			Batchable: &batchable,
			REST: &RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "/api/vote",
				ContentType: "application/x-www-form-urlencoded",
				MaxBytes:    1024,
				BodySchema: json.RawMessage(`{
					"type": "object",
					"required": ["id", "dir"],
					"properties": {
						"id": {"type": "string"},
						"dir": {"type": "integer"}
					}
				}`),
			},
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
			Method: http.MethodPost,
			Path:   "/api/vote",
			Operation: &SurfaceOperation{
				Model:            "destructive_action",
				Status:           "blocked",
				Risk:             "high",
				BlockedByDefault: true,
				Reason:           "operation metadata is bound by the executor",
			},
		}}},
	}
	req := connectors.OperationDirectWriteRequest{
		Operation: "acme.vote",
		Config: connectors.RuntimeConfig{
			CredentialRevision:  "fixture-credential-revision",
			ConfigurationDigest: "fixture-configuration-digest",
			WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
		},
		Body:         map[string]any{"id": "t3_abc", "dir": 1},
		RedactFields: []string{"token"},
	}

	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	if calls != 0 {
		t.Fatalf("preview reached the network; calls = %d, want 0", calls)
	}
	if preview.ApprovalTarget.Batchable {
		t.Fatal("preview made a batchable:false operation batchable")
	}

	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil {
		t.Fatal("OperationDirectWrite dispatched a destructive request without approval")
	}
	if calls != 0 {
		t.Fatalf("unapproved write reached the network; calls = %d, want 0", calls)
	}

	req.Approval = approvedEvidenceForPreview(t, preview)
	req.PreviewDigest = preview.Digest
	result, err := OperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("OperationDirectWrite: %v", err)
	}
	if calls != 1 {
		t.Fatalf("approved write calls = %d, want 1", calls)
	}
	if result.Status != http.StatusOK {
		t.Fatalf("status = %d, want %d", result.Status, http.StatusOK)
	}
	body, ok := result.Body.(map[string]any)
	if !ok {
		t.Fatalf("body type = %T, want map", result.Body)
	}
	if got := body["token"]; got != "server-token" {
		t.Fatalf("result token = %#v, want complete server token", got)
	}
	nested, ok := body["nested"].(map[string]any)
	if !ok || nested["token"] != "nested-server-token" {
		t.Fatalf("result nested body = %#v, want complete nested token", body["nested"])
	}
	if _, ok := body["token_redacted"]; ok {
		t.Fatalf("result body marked token redacted: %#v", body)
	}

	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil {
		t.Fatal("OperationDirectWrite accepted replayed approval evidence")
	} else if !strings.Contains(strings.ToLower(err.Error()), "approval") {
		t.Fatalf("replayed approval error = %v, want approval rejection", err)
	}
	if calls != 1 {
		t.Fatalf("replayed approval reached the network; calls = %d, want 1", calls)
	}
}

func TestOperationDirectWriteRedactingPoliciesKeepResponseBody(t *testing.T) {
	raw := []byte(`{"ok":true,"token":"server-token","nested":{"value":"visible"}}`)
	for _, policy := range []string{
		directWritePolicyJSONRedacted,
		directWritePolicyWriteResultRedacted,
		directWritePolicyGongBoundedInputRedacted,
	} {
		t.Run(policy, func(t *testing.T) {
			body, err := operationDirectWriteResponseBody(policy, raw, 1024)
			if err != nil {
				t.Fatalf("operationDirectWriteResponseBody: %v", err)
			}
			decoded, ok := body.(map[string]any)
			if !ok {
				t.Fatalf("body type = %T, want map", body)
			}
			if got := decoded["token"]; got != "server-token" {
				t.Fatalf("token = %#v, want complete response value", got)
			}
			nested, ok := decoded["nested"].(map[string]any)
			if !ok || nested["value"] != "visible" {
				t.Fatalf("nested = %#v, want complete response content", decoded["nested"])
			}
			if _, redacted := decoded["token_redacted"]; redacted {
				t.Fatalf("response was redacted: %#v", decoded)
			}
		})
	}
}

func TestOperationDirectWriteHonorsDeclaredJSONAndNoneResponsePolicies(t *testing.T) {
	for _, tt := range []struct {
		name     string
		policy   string
		wantBody bool
	}{
		{name: "json returns complete decoded body", policy: directWritePolicyJSON, wantBody: true},
		{name: "none intentionally suppresses response body", policy: directWritePolicyNone},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodPost {
					t.Fatalf("method = %s, want POST", r.Method)
				}
				if r.URL.Path != "/widgets" {
					t.Fatalf("path = %s, want /widgets", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"created":true,"id":"widget-42","nested":{"state":"complete"}}`))
			}))
			defer srv.Close()

			bundle := Bundle{
				Name: "acme",
				HTTP: HTTPBase{URL: srv.URL},
				Operations: []OperationSpec{{
					ID:            "acme.widgets.create",
					Kind:          "rest_write",
					Summary:       "Create one widget",
					Risk:          "medium",
					Approval:      "none",
					OutputPolicy:  tt.policy,
					MutationClass: "create",
					REST: &RESTOperationSpec{
						Method:      http.MethodPost,
						Path:        "/widgets",
						ContentType: "application/json",
						MaxBytes:    1024,
						BodySchema:  json.RawMessage(`{"type":"object","required":["name"],"properties":{"name":{"type":"string"}}}`),
					},
				}},
				Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
					Method: http.MethodPost,
					Path:   "/widgets",
					Operation: &SurfaceOperation{
						Model:            "write_action",
						Status:           "blocked",
						Risk:             "medium",
						BlockedByDefault: true,
						Reason:           "operation metadata is bound by the executor",
					},
				}}},
			}
			req := connectors.OperationDirectWriteRequest{
				Operation: "acme.widgets.create",
				Body:      map[string]any{"name": "widget"},
			}
			preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
			if err != nil {
				t.Fatalf("PreviewOperationDirectWrite: %v", err)
			}
			req.PreviewDigest = preview.Digest

			result, err := OperationDirectWrite(context.Background(), bundle, req, nil)
			if err != nil {
				t.Fatalf("OperationDirectWrite: %v", err)
			}
			if calls != 1 {
				t.Fatalf("request calls = %d, want 1", calls)
			}
			if !tt.wantBody {
				if result.Body != nil {
					t.Fatalf("none policy body = %#v, want nil", result.Body)
				}
				t.Logf("direct-write policy=%q status=%d response=<none>", tt.policy, result.Status)
				return
			}
			body, ok := result.Body.(map[string]any)
			if !ok {
				t.Fatalf("json policy body type = %T, want map", result.Body)
			}
			if body["id"] != "widget-42" || body["created"] != true {
				t.Fatalf("json policy body = %#v, want complete response fields", body)
			}
			nested, ok := body["nested"].(map[string]any)
			if !ok || nested["state"] != "complete" {
				t.Fatalf("json policy nested body = %#v, want complete nested response", body["nested"])
			}
			encoded, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("marshal json policy response: %v", err)
			}
			t.Logf("direct-write policy=%q status=%d response=%s", tt.policy, result.Status, encoded)
		})
	}
}

func TestOperationDirectWriteNeverRetriesNonIdempotentFailure(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"failed","token":"server-token"}`))
	}))
	defer srv.Close()

	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID:            "acme.create-widget",
			Kind:          "rest_write",
			Summary:       "Create a widget",
			Risk:          "medium",
			Approval:      "none",
			OutputPolicy:  "json_redacted",
			MutationClass: "create",
			REST: &RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "/widgets",
				ContentType: "application/json",
				MaxBytes:    1024,
				BodySchema:  json.RawMessage(`{"type":"object","required":["name"],"properties":{"name":{"type":"string"}}}`),
			},
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
			Method: http.MethodPost,
			Path:   "/widgets",
			Operation: &SurfaceOperation{
				Model:            "write_action",
				Status:           "blocked",
				Risk:             "medium",
				BlockedByDefault: true,
				Reason:           "operation metadata is bound by the executor",
			},
		}}},
	}
	req := connectors.OperationDirectWriteRequest{
		Operation: "acme.create-widget",
		Body:      map[string]any{"name": "widget"},
	}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest

	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil {
		t.Fatal("OperationDirectWrite error = nil, want HTTP 500")
	} else if !strings.Contains(err.Error(), "server-token") {
		t.Fatalf("OperationDirectWrite error = %q, want complete response error content", err)
	}
	if calls != 1 {
		t.Fatalf("non-idempotent write calls = %d, want exactly 1", calls)
	}
}

// A 307/308 redirect replays the original method and body. It is therefore a
// retry for a non-idempotent mutation, even though it happens below Requester
// in net/http. A prepared direct write must fail rather than follow it to a
// target the preview did not bind.
func TestOperationDirectWriteRefusesRedirectReplay(t *testing.T) {
	calls := 0
	redirectedCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path == "/widgets" {
			http.Redirect(w, r, "/redirected", http.StatusTemporaryRedirect)
			return
		}
		if r.URL.Path == "/redirected" {
			redirectedCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
			return
		}
		t.Fatalf("unexpected path %q", r.URL.Path)
	}))
	defer srv.Close()

	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID:            "acme.create-widget",
			Kind:          "rest_write",
			Summary:       "Create a widget",
			Risk:          "medium",
			Approval:      "none",
			OutputPolicy:  "json_redacted",
			MutationClass: "create",
			REST: &RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "/widgets",
				ContentType: "application/json",
				MaxBytes:    1024,
				BodySchema:  json.RawMessage(`{"type":"object","required":["name"],"properties":{"name":{"type":"string"}}}`),
			},
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
			Method: http.MethodPost,
			Path:   "/widgets",
			Operation: &SurfaceOperation{
				Model:            "write_action",
				Status:           "blocked",
				Risk:             "medium",
				BlockedByDefault: true,
				Reason:           "operation metadata is bound by the executor",
			},
		}}},
	}
	req := connectors.OperationDirectWriteRequest{
		Operation: "acme.create-widget",
		Body:      map[string]any{"name": "widget"},
	}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest

	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil {
		t.Fatal("OperationDirectWrite error = nil, want redirect rejection")
	}
	if calls != 1 || redirectedCalls != 0 {
		t.Fatalf("redirect calls = total %d / followed %d, want exactly 1 / 0", calls, redirectedCalls)
	}
}

func TestOperationDirectWriteProxyBasePathConsumesAPIRootOnce(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/dockerhub/v2/namespaces/acme/repositories" {
			t.Fatalf("request = %s %s, want POST proxy base plus one provider root", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	bundle := Bundle{
		Name: "dockerhub",
		HTTP: HTTPBase{URL: srv.URL + "/dockerhub/v2", APIRoot: "/v2"},
		Operations: []OperationSpec{{
			ID: "dockerhub.create-repository", Kind: "rest_write", Summary: "Create repository", Risk: "low", Approval: "none", OutputPolicy: "json", MutationClass: "create",
			REST: &RESTOperationSpec{Method: http.MethodPost, Path: "/v2/namespaces/acme/repositories", ContentType: "application/json", MaxBytes: 1024, BodySchema: json.RawMessage(`{"type":"object","required":["name"],"properties":{"name":{"type":"string"}},"additionalProperties":false}`)},
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{Method: http.MethodPost, Path: "/v2/namespaces/acme/repositories", Operation: &SurfaceOperation{Model: "write_action"}}}},
	}
	req := connectors.OperationDirectWriteRequest{Operation: "dockerhub.create-repository", Body: map[string]any{"name": "fixture"}}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest
	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err != nil {
		t.Fatalf("OperationDirectWrite: %v", err)
	}
}
