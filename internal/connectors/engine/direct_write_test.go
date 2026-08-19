package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

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
		if got := r.Header.Get("X-Change-Reason"); got != "correctness" {
			t.Fatalf("X-Change-Reason = %q, want declaration-owned value", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Write-Receipt", "receipt-42")
		w.Header().Add("Set-Cookie", "transport-secret")
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
				Parameters: []OperationParameter{{
					Name: "X-Change-Reason", In: "header", Type: "string", Required: true,
					Schema: json.RawMessage(`{"type":"string","enum":["correctness","moderation"]}`), MaxBytes: 32,
				}},
				BodySchema: json.RawMessage(`{
					"type": "object",
					"required": ["id", "dir"],
					"properties": {
						"id": {"type": "string"},
						"dir": {"type": "integer"}
					}
				}`),
				Response: &OperationResponseSpec{Headers: []OperationResponseHeaderSpec{
					{Name: "X-Write-Receipt", MaxBytes: 64},
					{Name: "Set-Cookie", MaxBytes: 64},
				}},
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
		Headers:      map[string]string{"X-Change-Reason": "correctness"},
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
	changedHeader := req
	changedHeader.Headers = map[string]string{"X-Change-Reason": "moderation"}
	changedPreview, err := PreviewOperationDirectWrite(context.Background(), bundle, changedHeader, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite changed header: %v", err)
	}
	if changedPreview.Digest == preview.Digest {
		t.Fatalf("header change preserved preview digest %q", preview.Digest)
	}
	changedHeader.PreviewDigest = preview.Digest
	changedHeader.Approval = approvedEvidenceForPreview(t, preview)
	if _, err := OperationDirectWrite(context.Background(), bundle, changedHeader, nil); err == nil || !strings.Contains(err.Error(), "no longer matches") {
		t.Fatalf("OperationDirectWrite with changed header = %v, want preview mismatch", err)
	}
	if calls != 0 {
		t.Fatalf("changed header reached network; calls = %d, want 0", calls)
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
	if got := result.Headers["X-Write-Receipt"].Values; len(got) != 1 || got[0] != "receipt-42" {
		t.Fatalf("write receipt = %#v, want declared provider metadata", got)
	}
	if cookie, ok := result.Headers["Set-Cookie"]; !ok || !cookie.Redacted {
		t.Fatalf("Set-Cookie = %#v, want explicit redaction marker", cookie)
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

func TestOperationDirectWriteContentTypeAllowsClosedJSONFamily(t *testing.T) {
	op := OperationSpec{ID: "acme.scim.create", REST: &RESTOperationSpec{ContentType: "application/scim+json"}}
	contentType, format, err := operationDirectWriteContentType(op, map[string]any{"userName": "ada"})
	if err != nil {
		t.Fatalf("operationDirectWriteContentType: %v", err)
	}
	if contentType != "application/scim+json" || format != "json" {
		t.Fatalf("content type/format = %q/%q, want application/scim+json/json", contentType, format)
	}
}

func TestOperationDirectWriteContentTypeRejectsUnknownBeforeIO(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	t.Cleanup(srv.Close)
	bundle := Bundle{Name: "acme", HTTP: HTTPBase{URL: srv.URL}, Operations: []OperationSpec{{
		ID: "acme.bad.create", Kind: "rest_write", Summary: "Bad content", Risk: "medium", Approval: "none", OutputPolicy: "none", MutationClass: "create",
		REST: &RESTOperationSpec{Method: http.MethodPost, Path: "/widgets", ContentType: "application/x-unsafe", MaxBytes: 1024},
	}}, Surface: &APISurface{Endpoints: []SurfaceEndpoint{{Method: http.MethodPost, Path: "/widgets", Operation: &SurfaceOperation{Model: "write"}}}}}
	_, err := PreviewOperationDirectWrite(context.Background(), bundle, connectors.OperationDirectWriteRequest{Operation: "acme.bad.create", Body: map[string]any{"value": "x"}}, nil)
	if err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("PreviewOperationDirectWrite error = %v, want closed-content-type refusal", err)
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want pre-I/O refusal", requests)
	}
}

func TestOperationDirectWriteStoresReturnedSecretAndRetainsRuntimeResponse(t *testing.T) {
	const canary = "returned-credential-canary"
	store := newRecordingSecretStore()
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"credential":"` + canary + `"}`))
	}))
	t.Cleanup(srv.Close)
	bundle := Bundle{Name: "acme", HTTP: HTTPBase{URL: srv.URL}, Operations: []OperationSpec{{
		ID: "acme.credentials.create", Kind: "rest_write", Summary: "Create credential", Risk: "high", Approval: "typed", OutputPolicy: directWritePolicySecretStored,
		MutationClass: "secret", SecretSensitive: true,
		SensitivePolicy: &SensitivePolicySpec{InputMode: "env", ApprovalMode: "typed_confirmation", ResponseSecretField: "credential", ResponseSecretStoreKey: "generated_credential"},
		REST:            &RESTOperationSpec{Method: http.MethodPost, Path: "/v2/credentials", MaxBytes: 1024},
	}}, Surface: &APISurface{Endpoints: []SurfaceEndpoint{{Method: http.MethodPost, Path: "/v2/credentials", Operation: &SurfaceOperation{Model: "write"}}}}}
	req := connectors.OperationDirectWriteRequest{Operation: "acme.credentials.create", Config: connectors.RuntimeConfig{SecretStore: store, CredentialRevision: "fixture-credential-revision", ConfigurationDigest: "fixture-configuration-digest", WriteApprovalScope: connectors.WriteApprovalScopeFixture}}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest
	req.Approval = approvedEvidenceForPreview(t, preview)
	result, err := OperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("OperationDirectWrite: %v", err)
	}
	if got, ok := store.written("generated_credential"); !ok || got != canary {
		t.Fatal("credential response did not reach the declared encrypted secret store")
	}
	body, ok := result.Body.(map[string]any)
	if !ok || body["credential"] != canary {
		t.Fatal("runtime result did not retain the complete provider response")
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}
}

func TestOperationDirectWriteRejectsSecretResponseWithoutEncryptedStoreBeforeIO(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	t.Cleanup(srv.Close)
	bundle := Bundle{Name: "acme", HTTP: HTTPBase{URL: srv.URL}, Operations: []OperationSpec{{
		ID: "acme.credentials.create", Kind: "rest_write", Summary: "Create credential", Risk: "high", Approval: "typed", OutputPolicy: "json",
		MutationClass: "secret", SecretSensitive: true,
		SensitivePolicy: &SensitivePolicySpec{InputMode: "env", ApprovalMode: "typed_confirmation", ResponseSecretField: "credential", ResponseSecretStoreKey: "generated_credential"},
		REST:            &RESTOperationSpec{Method: http.MethodPost, Path: "/v2/credentials", MaxBytes: 1024},
	}}, Surface: &APISurface{Endpoints: []SurfaceEndpoint{{Method: http.MethodPost, Path: "/v2/credentials", Operation: &SurfaceOperation{Model: "write"}}}}}
	_, err := PreviewOperationDirectWrite(context.Background(), bundle, connectors.OperationDirectWriteRequest{Operation: "acme.credentials.create", Config: connectors.RuntimeConfig{SecretStore: newRecordingSecretStore()}}, nil)
	if err == nil || !strings.Contains(err.Error(), "secret response") {
		t.Fatalf("PreviewOperationDirectWrite error = %v, want pre-I/O encrypted-store refusal", err)
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want pre-I/O refusal", requests)
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
