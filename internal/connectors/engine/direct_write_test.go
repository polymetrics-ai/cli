package engine

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

func TestOperationDirectWriteNoResponseRetainsAttemptIdentity(t *testing.T) {
	bundle := graphQLOperationBundle("http://127.0.0.1:1", "graphql_mutation")
	request := connectors.OperationDirectWriteRequest{
		Operation: "acme.widgets.mutation",
		Config: connectors.RuntimeConfig{
			CredentialRevision:  "fixture-credential-revision",
			ConfigurationDigest: "fixture-configuration-digest",
			WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
		},
		Body: map[string]any{"id": "widget-1"},
	}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	request.PreviewDigest = preview.Digest
	request.Approval = approvedEvidenceForPreview(t, preview)

	result, err := OperationDirectWrite(context.Background(), bundle, request, nil)
	if err == nil {
		t.Fatal("OperationDirectWrite() error = nil, want transport failure")
	}
	if result.Connector != "acme" || result.Operation != request.Operation || result.Method != http.MethodPost || result.Path != "/graphql" || result.ResponseReceived {
		t.Fatalf("attempt receipt = %#v, want sealed identity without a response", result)
	}

	invalid := request
	invalid.Body = map[string]any{}
	if preflight, preflightErr := OperationDirectWrite(context.Background(), bundle, invalid, nil); preflightErr == nil || preflight.Operation != "" {
		t.Fatalf("preflight result = %#v error = %v, want no attempt receipt", preflight, preflightErr)
	}
}

func TestSensitiveTransformUnimplementedFailsBeforePreview(t *testing.T) {
	bundle := graphQLOperationBundle("http://127.0.0.1:1", "graphql_mutation")
	bundle.Operations[0].SensitivePolicy = &SensitivePolicySpec{
		InputMode:    "env",
		Transform:    "github_secret_encryption",
		ApprovalMode: "typed_confirmation",
	}
	request := connectors.OperationDirectWriteRequest{
		Operation: "acme.widgets.mutation",
		Config: connectors.RuntimeConfig{
			CredentialRevision:  "fixture-credential-revision",
			ConfigurationDigest: "fixture-configuration-digest",
			WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
		},
		Body: map[string]any{"id": "widget-1"},
	}
	if _, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil); err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Fatalf("PreviewOperationDirectWrite error = %v, want pre-preview transform refusal", err)
	}
}

func TestSensitiveTransformRegisteredExecutionBindsDigest(t *testing.T) {
	bundle := graphQLOperationBundle("http://127.0.0.1:1", "graphql_mutation")
	bundle.Operations[0].SensitivePolicy = &SensitivePolicySpec{InputMode: "env", Transform: "none", ApprovalMode: "typed_confirmation"}
	request := connectors.OperationDirectWriteRequest{
		Operation: "acme.widgets.mutation",
		Config: connectors.RuntimeConfig{
			CredentialRevision:  "fixture-credential-revision",
			ConfigurationDigest: "fixture-configuration-digest",
			WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
		},
		Body: map[string]any{"id": "widget-1"},
	}
	prepared, err := prepareOperationDirectWrite(context.Background(), bundle, request, nil)
	if err != nil {
		t.Fatalf("prepareOperationDirectWrite: %v", err)
	}
	definition, ok := prepared.prepared.Definition.(map[string]any)
	if !ok {
		t.Fatalf("prepared definition = %#v", prepared.prepared.Definition)
	}
	transform, ok := definition["sensitive_transform"].(map[string]any)
	if !ok || transform["name"] != "none" || transform["version"] != "v1" {
		t.Fatalf("prepared transform definition = %#v", definition["sensitive_transform"])
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
	if cookie, ok := result.Headers["Set-Cookie"]; !ok || len(cookie.Values) != 1 || cookie.Values[0] != "transport-secret" {
		t.Fatalf("Set-Cookie = %#v, want exact internal provider metadata", cookie)
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

// TestOperationDirectWriteRetainsProviderReceiptWhenSecretStoreFails protects
// the terminal-failure contract: the printable error remains secret-safe, but
// the engine result and wrapped cause retain the exact provider receipt for
// the App's bounded persistence projection.
func TestOperationDirectWriteRetainsProviderReceiptWhenSecretStoreFails(t *testing.T) {
	const providerCredential = "returned-credential-store-failure-canary"
	store := newRecordingSecretStore()
	store.err = errors.New("fixture secret store failure")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Provider-Trace", "secret-store-failure")
		_, _ = w.Write([]byte(`{"credential":"` + providerCredential + `"}`))
	}))
	t.Cleanup(server.Close)
	bundle := Bundle{Name: "acme", HTTP: HTTPBase{URL: server.URL}, Operations: []OperationSpec{{
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
	if err == nil {
		t.Fatal("OperationDirectWrite accepted secret-store failure")
	}
	if strings.Contains(err.Error(), providerCredential) {
		t.Fatalf("secret-store error leaked provider credential: %v", err)
	}
	if result.Status != http.StatusOK || result.Headers["X-Provider-Trace"].Values[0] != "secret-store-failure" {
		t.Fatalf("result = %#v, want complete provider receipt", result)
	}
	var providerErr *connsdk.HTTPError
	if !errors.As(err, &providerErr) || providerErr.Body != `{"credential":"`+providerCredential+`"}` {
		t.Fatalf("error = %T %v, want exact retained provider receipt", err, err)
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
			response, err := operationDirectWriteResponseBody(policy, raw, 1024, nil)
			if err != nil {
				t.Fatalf("operationDirectWriteResponseBody: %v", err)
			}
			decoded, ok := response.body.(map[string]any)
			if !ok {
				t.Fatalf("body type = %T, want map", response.body)
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
		name      string
		policy    string
		wantBody  bool
		bodyless  bool
		response  string
		wantErr   string
		wantNull  bool
		emptyJSON bool
	}{
		{name: "json returns complete decoded body", policy: directWritePolicyJSON, wantBody: true},
		{name: "none retains complete response body", policy: directWritePolicyNone, wantBody: true},
		{name: "none accepts bodyless response", policy: directWritePolicyNone, bodyless: true},
		{name: "none accepts empty declared JSON as absent body", policy: directWritePolicyNone, emptyJSON: true},
		{name: "none preserves literal JSON null", policy: directWritePolicyNone, response: `null`, wantNull: true},
		{name: "json rejects trailing response content", policy: directWritePolicyJSON, response: `{"created":true} trailing`, wantErr: "not JSON"},
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
				if tt.bodyless {
					w.Header().Set("X-Provider-Receipt", "receipt-204")
					w.WriteHeader(http.StatusNoContent)
					return
				}
				if tt.emptyJSON {
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("X-Provider-Receipt", "receipt-empty-json")
					w.WriteHeader(http.StatusOK)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				response := tt.response
				if response == "" {
					response = `{"created":true,"id":"widget-42","nested":{"state":"complete"}}`
				}
				_, _ = w.Write([]byte(response))
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
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("OperationDirectWrite error = %v, want %q", err, tt.wantErr)
				}
				if !result.ResponseReceived || result.BodyPresent != (len(tt.response) > 0) || result.BodyRaw != tt.response {
					t.Fatalf("trailing response result = %#v, want complete received raw response", result)
				}
				if tt.emptyJSON {
					if result.Status != http.StatusOK || result.BodyBytes != 0 || result.BodyRawEncoding != "" {
						t.Fatalf("empty declared JSON result = %#v, want complete zero-value provider response", result)
					}
					if receipt := result.Headers["X-Provider-Receipt"].Values; len(receipt) != 1 || receipt[0] != "receipt-empty-json" {
						t.Fatalf("empty declared JSON receipt = %#v, want preserved provider receipt", receipt)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("OperationDirectWrite: %v", err)
			}
			if calls != 1 {
				t.Fatalf("request calls = %d, want 1", calls)
			}
			if !result.ResponseReceived {
				t.Fatal("result did not retain the successful provider response")
			}
			if tt.bodyless {
				if result.Status != http.StatusNoContent {
					t.Fatalf("bodyless response status = %d, want %d", result.Status, http.StatusNoContent)
				}
				if receipt := result.Headers["X-Provider-Receipt"].Values; len(receipt) != 1 || receipt[0] != "receipt-204" {
					t.Fatalf("bodyless response receipt header = %#v, want receipt-204", receipt)
				}
				if result.BodyPresent || result.BodyBytes != 0 || result.BodyRaw != "" || result.Body != nil {
					t.Fatalf("bodyless none result = %#v, want a distinct empty-body response", result)
				}
				return
			}
			if tt.emptyJSON {
				if result.Status != http.StatusOK || result.BodyPresent || result.BodyBytes != 0 || result.BodyRaw != "" || result.Body != nil {
					t.Fatalf("empty declared JSON result = %#v, want an absent transport body", result)
				}
				return
			}
			if tt.wantNull {
				if !result.BodyPresent || result.BodyRaw != "null" || result.BodyBytes != len("null") || result.Body != nil {
					t.Fatalf("literal JSON null result = %#v, want a present null body", result)
				}
				return
			}
			if !tt.wantBody || !result.BodyPresent || result.BodyRaw == "" {
				t.Fatalf("direct-write result = %#v, want a present provider body", result)
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

func TestOperationDirectWriteRetainsRESTDecodeFailureResponse(t *testing.T) {
	const rawResponse = `{"result":"provider-response-canary"`
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("X-Provider-Trace", "trace-123")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(rawResponse))
	}))
	t.Cleanup(srv.Close)

	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID:            "acme.widgets.create",
			Kind:          "rest_write",
			Summary:       "Create one widget",
			Risk:          "medium",
			Approval:      "none",
			OutputPolicy:  directWritePolicyJSON,
			MutationClass: "create",
			REST: &RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "/widgets",
				ContentType: "application/json",
				MaxBytes:    1024,
				BodySchema:  json.RawMessage(`{"type":"object","required":["name"],"properties":{"name":{"type":"string"}}}`),
			},
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{Method: http.MethodPost, Path: "/widgets", Operation: &SurfaceOperation{Model: "write"}}}},
	}
	req := connectors.OperationDirectWriteRequest{Operation: "acme.widgets.create", Body: map[string]any{"name": "widget"}}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest
	_, err = OperationDirectWrite(context.Background(), bundle, req, nil)
	if err == nil {
		t.Fatal("OperationDirectWrite error = nil, want decode failure")
	}
	if strings.Contains(err.Error(), "provider-response-canary") {
		t.Fatalf("OperationDirectWrite printable error exposed provider response: %v", err)
	}
	var providerResponse *connsdk.HTTPError
	if !errors.As(err, &providerResponse) {
		t.Fatalf("OperationDirectWrite error = %v, want retained provider response cause", err)
	}
	if providerResponse.Status != http.StatusOK {
		t.Fatalf("provider response status = %d, want %d", providerResponse.Status, http.StatusOK)
	}
	if providerResponse.Header.Get("X-Provider-Trace") != "trace-123" {
		t.Fatalf("provider response headers = %#v, want provider trace", providerResponse.Header)
	}
	if providerResponse.Body != rawResponse {
		t.Fatalf("provider response body = %q, want %q", providerResponse.Body, rawResponse)
	}
	if calls != 1 {
		t.Fatalf("provider calls = %d, want 1", calls)
	}
}

func TestOperationDirectWritePreservesExplicitNonJSONResponses(t *testing.T) {
	for _, tt := range []struct {
		name        string
		contentType string
		body        []byte
	}{
		{name: "text begins t", contentType: "text/plain", body: []byte("thanks")},
		{name: "text begins f", contentType: "text/plain", body: []byte("false")},
		{name: "text begins n", contentType: "text/plain", body: []byte("null")},
		{name: "text begins digit", contentType: "text/plain", body: []byte("123")},
		{name: "text begins brace", contentType: "text/plain", body: []byte(`{"provider":"text"}`)},
		{name: "text begins bracket", contentType: "text/plain", body: []byte(`["provider","text"]`)},
		{name: "binary", contentType: "application/octet-stream", body: []byte{0xff, 0x00, 0x7f}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				_, _ = w.Write(tt.body)
			}))
			defer server.Close()

			bundle := Bundle{
				Name: "acme",
				HTTP: HTTPBase{URL: server.URL},
				Operations: []OperationSpec{{
					ID: "acme.widgets.create", Kind: "rest_write", Summary: "Create one widget", Risk: "medium", Approval: "none", OutputPolicy: directWritePolicyJSON, MutationClass: "create",
					REST: &RESTOperationSpec{Method: http.MethodPost, Path: "/widgets", ContentType: "application/json", MaxBytes: 1024, BodySchema: json.RawMessage(`{"type":"object","required":["name"],"properties":{"name":{"type":"string"}}}`)},
				}},
				Surface: &APISurface{Endpoints: []SurfaceEndpoint{{Method: http.MethodPost, Path: "/widgets", Operation: &SurfaceOperation{Model: "write_action", Status: "blocked", Risk: "medium", BlockedByDefault: true, Reason: "operation metadata is bound by the executor"}}}},
			}
			req := connectors.OperationDirectWriteRequest{Operation: "acme.widgets.create", Body: map[string]any{"name": "widget"}}
			preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
			if err != nil {
				t.Fatalf("PreviewOperationDirectWrite: %v", err)
			}
			req.PreviewDigest = preview.Digest
			result, err := OperationDirectWrite(context.Background(), bundle, req, nil)
			if err != nil {
				t.Fatalf("OperationDirectWrite: %v", err)
			}
			if !result.ResponseReceived || !result.BodyPresent || result.Body != nil || result.BodyBytes != len(tt.body) {
				t.Fatalf("non-JSON result = %#v, want raw provider response", result)
			}
			if tt.contentType == "text/plain" {
				if result.BodyRawEncoding != "text" || result.BodyRaw != string(tt.body) {
					t.Fatalf("text provider response = %#v, want %q", result, tt.body)
				}
				return
			}
			if result.BodyRawEncoding != "base64" || result.BodyRaw != base64.StdEncoding.EncodeToString(tt.body) {
				t.Fatalf("binary provider response = %#v, want byte-exact base64", result)
			}
		})
	}
}

func TestOperationDirectWriteNeverRetriesNonIdempotentFailure(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Add("X-Provider-Receipt", "receipt-one")
		w.Header().Add("X-Provider-Receipt", "receipt-two")
		w.Header().Set("Content-Type", "application/json")
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

	result, err := OperationDirectWrite(context.Background(), bundle, req, nil)
	if err == nil {
		t.Fatal("OperationDirectWrite error = nil, want HTTP 500")
	} else if strings.Contains(err.Error(), "server-token") {
		t.Fatal("OperationDirectWrite error leaked a provider body")
	}
	if !result.ResponseReceived || result.Status != http.StatusInternalServerError || result.BodyRaw != `{"error":"failed","token":"server-token"}` {
		t.Fatal("terminal direct-write result did not retain the provider response")
	}
	if got := result.Headers["X-Provider-Receipt"].Values; !reflect.DeepEqual(got, []string{"receipt-one", "receipt-two"}) {
		t.Fatal("terminal direct-write receipt did not retain both provider values")
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

func TestOperationDirectWriteRejectsUndeclaredRequestBindingsBeforeIO(t *testing.T) {
	for _, tc := range []struct {
		name    string
		bundle  func(string) Bundle
		request func() connectors.OperationDirectWriteRequest
		wantErr string
	}{
		{
			name:   "undeclared path parameter",
			bundle: structuredRESTBodyBundle,
			request: func() connectors.OperationDirectWriteRequest {
				req := structuredRESTBodyRequest()
				req.PathParams["undeclared"] = "override"
				return req
			},
			wantErr: "not declared by rest.path",
		},
		{
			name: "open scalar nesting",
			bundle: func(baseURL string) Bundle {
				bundle := structuredRESTBodyBundle(baseURL)
				rest := *bundle.Operations[0].REST
				rest.BodySchema = json.RawMessage(`{"type":"object","properties":{"payload":{}}}`)
				bundle.Operations[0].REST = &rest
				return bundle
			},
			request: func() connectors.OperationDirectWriteRequest {
				req := structuredRESTBodyRequest()
				req.Body = map[string]any{"payload": map[string]any{"undeclared": "value"}}
				return req
			},
			wantErr: "does not permit a nested object",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				calls++
			}))
			t.Cleanup(server.Close)

			_, err := OperationDirectWrite(context.Background(), tc.bundle(server.URL), tc.request(), nil)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("OperationDirectWrite error = %v, want %q", err, tc.wantErr)
			}
			if calls != 0 {
				t.Fatalf("rejected binding reached provider; calls = %d", calls)
			}
		})
	}
}

func TestPreviewOperationDirectWriteDoesNotExposeSecretFilterValues(t *testing.T) {
	const secret = "secret-canary-preview"
	bundle := structuredRESTBodyBundle("https://example.invalid/{{ secrets.token | unix_seconds }}")
	request := structuredRESTBodyRequest()
	request.Config.Secrets = map[string]string{"token": secret}
	_, err := PreviewOperationDirectWrite(context.Background(), bundle, request, nil)
	if err == nil {
		t.Fatal("PreviewOperationDirectWrite error = nil, want secret filter rejection")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("PreviewOperationDirectWrite error exposed secret value: %v", err)
	}
	if !strings.Contains(err.Error(), "secrets.token") {
		t.Fatalf("PreviewOperationDirectWrite error = %v, want secret reference", err)
	}
}
