package engine

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"

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
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"token":"server-token","plainkey":"fixture-plaintext-key","nested":{"token":"nested-server-token"}}`))
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
				RedactFields: []string{"plainkey"},
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
		Body: map[string]any{"id": "t3_abc", "dir": 1},
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
	if _, exposed := body["token"]; exposed {
		t.Fatal("json_redacted direct-write result exposed a generic token-shaped field")
	}
	if body["token_redacted"] != true {
		t.Fatalf("json_redacted direct-write result lacks generic token redaction marker: %#v", body)
	}
	if _, exposed := body["plainkey"]; exposed {
		t.Fatal("json_redacted direct-write result exposed an operation-declared sensitive field")
	}
	if body["plainkey_redacted"] != true {
		t.Fatalf("json_redacted direct-write result lacks declared-field redaction marker: %#v", body)
	}
	nested, ok := body["nested"].(map[string]any)
	if !ok || nested["token_redacted"] != true {
		t.Fatalf("json_redacted direct-write result lacks nested generic redaction marker: %#v", body["nested"])
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

func TestPreflightOperationDirectWriteValidatesDeclaredContract(t *testing.T) {
	bundle := Bundle{
		Name: "acme",
		Operations: []OperationSpec{{
			ID:            "acme.vote",
			Kind:          "rest_write",
			Summary:       "Vote on a post",
			Risk:          "high",
			Approval:      "plan-preview-confirm-execute",
			OutputPolicy:  "json_redacted",
			MutationClass: "update",
			REST: &RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "/api/vote",
				ContentType: "application/json",
				MaxBytes:    1024,
				BodySchema:  json.RawMessage(`{"type":"object","additionalProperties":false}`),
			},
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
			Method: http.MethodPost,
			Path:   "/api/vote",
			CoveredBy: &SurfaceCoverage{
				DirectWrite: "vote",
			},
		}}},
	}

	tests := []struct {
		name    string
		method  string
		path    string
		policy  string
		wantErr string
	}{
		{name: "exact declared binding", method: http.MethodPost, path: "/api/vote", policy: "json_redacted"},
		{name: "wrong method", method: http.MethodPatch, path: "/api/vote", policy: "json_redacted", wantErr: "method PATCH does not match"},
		{name: "wrong path", method: http.MethodPost, path: "/api/other", policy: "json_redacted", wantErr: "path \"/api/other\" does not match"},
		{name: "wrong output policy", method: http.MethodPost, path: "/api/vote", policy: "unsupported", wantErr: "not supported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PreflightOperationDirectWrite(bundle, "acme.vote", tt.method, tt.path, tt.policy)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("PreflightOperationDirectWrite: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("PreflightOperationDirectWrite error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

// TestOperationDirectWriteUsesDeclaredOperationOriginAndAuth proves a
// credential-sensitive operation can bind its customer-hosted origin and
// bearer credential together, without reusing the connector's ordinary API
// bearer token. The operation declaration is deliberately loaded from JSON
// so the meta-schema and runtime must accept the same narrow contract.
func TestOperationDirectWriteUsesDeclaredOperationOriginAndAuth(t *testing.T) {
	var ordinaryAPICalls, keyConnectorCalls int
	ordinaryAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ordinaryAPICalls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ordinaryAPI.Close()

	keyConnectorJWT := "fixture-key-connector-jwt"
	keyConnector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keyConnectorCalls++
		if r.Method != http.MethodPost || r.URL.Path != "/kms/archival/decrypt" {
			t.Fatalf("key connector request = %s %s, want POST /kms/archival/decrypt", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer "+keyConnectorJWT {
			t.Fatal("key connector request did not receive the operation-declared bearer credential")
		}
		if r.Header.Get("X-Ordinary-Token") != "" {
			t.Fatal("key connector request inherited an ordinary API secret header")
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read no-body direct-write request: %v", err)
		}
		if len(body) != 0 {
			t.Fatalf("no-body direct-write payload = %q, want no payload", body)
		}
		if r.Header.Get("Content-Type") != "" {
			t.Fatalf("no-body direct-write Content-Type = %q, want empty", r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer keyConnector.Close()

	fsys := fullValidBundleFS("acme")
	fsys["acme/spec.json"].Data = []byte(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"base_url": {"type": "string"},
			"key_connector_base_url": {"type": "string"},
			"access_token": {"type": "string", "x-secret": true},
			"key_connector_jwt": {"type": "string", "x-secret": true}
		}
	}`)
	fsys["acme/streams.json"].Data = []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"auth": [{"mode": "bearer", "token": "{{ secrets.access_token }}"}],
			"headers": {"X-Ordinary-Token": "{{ secrets.access_token }}"},
			"pagination": {"type": "none"}
		},
		"streams": []
	}`)
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(`{
		"operations": [{
			"id": "acme.decrypt",
			"kind": "rest_write",
			"summary": "Decrypt one archival key",
			"risk": "high",
			"approval": "plan-preview-confirm-execute",
			"output_policy": "json_redacted",
			"mutation_class": "create",
			"rest": {
				"method": "POST",
				"path": "/kms/archival/decrypt",
				"base_url": "{{ config.key_connector_base_url }}",
				"auth": [{"mode": "bearer", "token": "{{ secrets.key_connector_jwt }}"}],
				"max_bytes": 1024,
				"body_schema": {"type": "object", "additionalProperties": false}
			}
		}]
	}`)}
	delete(fsys, "acme/api_surface.json")

	bundle, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load declared per-operation origin/auth bundle: %v", err)
	}
	req := connectors.OperationDirectWriteRequest{
		Operation: "acme.decrypt",
		Config: connectors.RuntimeConfig{
			Config: map[string]string{
				"base_url":               ordinaryAPI.URL,
				"key_connector_base_url": keyConnector.URL,
			},
			Secrets: map[string]string{
				"access_token":      "fixture-ordinary-access-token",
				"key_connector_jwt": keyConnectorJWT,
			},
			CredentialRevision:  "fixture-credential-revision",
			ConfigurationDigest: "fixture-configuration-digest",
			WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
		},
	}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest
	req.Approval = approvedEvidenceForPreview(t, preview)
	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err != nil {
		t.Fatalf("OperationDirectWrite: %v", err)
	}
	if ordinaryAPICalls != 0 {
		t.Fatalf("ordinary API received %d request(s), want no OAuth bearer sent to the key-connector route", ordinaryAPICalls)
	}
	if keyConnectorCalls != 1 {
		t.Fatalf("key connector requests = %d, want 1", keyConnectorCalls)
	}
}

// TestOperationDirectWriteBase64PathUploadIsPreviewBound captures the typed
// operation-level upload needed by Zoom Clips temporary image files. The
// locally supplied path must be replaced by a bounded canonical base64 value
// before preview and before network dispatch; there is no generic raw body or
// caller-selected encoder involved.
func TestOperationDirectWriteBase64PathUploadIsPreviewBound(t *testing.T) {
	decoded, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9WlCkZYAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatalf("decode fixture png: %v", err)
	}
	dir := t.TempDir()
	sourcePath := dir + "/fixture.png"
	if err := os.WriteFile(sourcePath, decoded, 0o600); err != nil {
		t.Fatalf("write fixture image: %v", err)
	}
	want := base64.StdEncoding.EncodeToString(decoded)
	digest := sha256.Sum256(decoded)
	calls := 0
	gotFile := ""
	leakedPath := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode temporary-file request: %v", err)
		}
		gotFile, _ = body["file"].(string)
		_, leakedPath = body["file_path"]
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"file_id":"fixture-temporary-file"}`))
	}))
	defer srv.Close()

	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID:              "acme.clips.temporary_file",
			Kind:            "rest_write",
			Summary:         "Upload one bounded temporary image",
			Risk:            "high",
			Approval:        "plan-preview-confirm-execute",
			OutputPolicy:    "json_redacted",
			MutationClass:   "create",
			SensitivePolicy: &SensitivePolicySpec{RedactFields: []string{"file", "file_id", "file_path"}},
			REST: &RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "/clips/files/tmp",
				ContentType: "application/json",
				MaxBytes:    1024,
				BodySchema:  json.RawMessage(`{"type":"object","additionalProperties":false,"required":["file_path"],"properties":{"file_path":{"type":"string"}}}`),
				Base64Upload: &Base64UploadSpec{
					Source:                "path",
					SourceField:           "file_path",
					ContentField:          "file",
					MaxDecodedBytes:       2 << 20,
					MaxEncodedBytes:       int64(base64.StdEncoding.EncodedLen(2 << 20)),
					AllowedMediaTypes:     []string{"image/png", "image/jpeg", "image/gif"},
					AllowedFileExtensions: []string{".png", ".jpg", ".jpeg", ".gif"},
				},
			},
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
			Method:    http.MethodPost,
			Path:      "/clips/files/tmp",
			Operation: &SurfaceOperation{Model: "sensitive_reverse_etl", Status: "blocked", Risk: "high", BlockedByDefault: true, Reason: "operation metadata is bound by the executor"},
		}}},
	}
	req := connectors.OperationDirectWriteRequest{
		Operation: "acme.clips.temporary_file",
		Config: connectors.RuntimeConfig{
			ProjectDir:            dir,
			CredentialRevision:    "fixture-credential-revision",
			ConfigurationDigest:   "fixture-configuration-digest",
			WriteApprovalScope:    connectors.WriteApprovalScopeFixture,
			ApprovedPayloadSHA256: map[string]string{connectors.PayloadApprovalKey(0, "file_path"): hex.EncodeToString(digest[:])},
		},
		Body: map[string]any{"file_path": sourcePath},
	}
	metadata, err := OperationDirectWriteMetadata(bundle, req.Operation)
	if err != nil {
		t.Fatalf("OperationDirectWriteMetadata temporary image: %v", err)
	}
	if len(metadata.PayloadFileFields) != 1 || metadata.PayloadFileFields[0] != "file_path" {
		t.Fatalf("temporary-image payload file fields = %#v, want file_path", metadata.PayloadFileFields)
	}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite temporary image: %v", err)
	}
	encodedPreview, err := json.Marshal(preview)
	if err != nil {
		t.Fatalf("marshal preview: %v", err)
	}
	if strings.Contains(string(encodedPreview), sourcePath) {
		t.Fatal("temporary-file preview exposed local source path")
	}
	if strings.Contains(string(encodedPreview), want) {
		t.Fatal("temporary-file preview exposed local base64 payload")
	}
	req.PreviewDigest = preview.Digest
	req.Approval = approvedEvidenceForPreview(t, preview)
	if err := os.WriteFile(sourcePath, append(append([]byte(nil), decoded...), 0), 0o600); err != nil {
		t.Fatalf("change fixture image after preview: %v", err)
	}
	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil || !strings.Contains(err.Error(), "does not match its approved digest") {
		t.Fatalf("OperationDirectWrite changed temporary image = %v, want approved-digest rejection", err)
	}
	if calls != 0 {
		t.Fatalf("changed temporary-file image reached network; calls = %d, want 0", calls)
	}
	if err := os.WriteFile(sourcePath, decoded, 0o600); err != nil {
		t.Fatalf("restore fixture image: %v", err)
	}
	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err != nil {
		t.Fatalf("OperationDirectWrite temporary image: %v", err)
	}
	if calls != 1 {
		t.Fatalf("temporary-file calls = %d, want one", calls)
	}
	if gotFile != want {
		t.Fatalf("temporary-file base64 = %q, want canonical local image payload", gotFile)
	}
	if leakedPath {
		t.Fatal("temporary-file request exposed local source path")
	}
}

func TestOperationDirectWriteLegacyRedactingPolicyNamesKeepResponseBody(t *testing.T) {
	raw := []byte(`{"ok":true,"token":"server-token","nested":{"value":"visible"}}`)
	for _, policy := range []string{
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

func TestOperationDirectWriteJSONRedactedErrorsHideDeclaredRequestAndResponseFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"key_id":"fixture-key-id","plainkey":"fixture-plaintext-key","token":"server-token"}`))
	}))
	defer srv.Close()

	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID:            "acme.decrypt",
			Kind:          "rest_write",
			Summary:       "Decrypt one fixture key",
			Risk:          "high",
			Approval:      "plan-preview-approve-execute",
			OutputPolicy:  "json_redacted",
			MutationClass: "update",
			SensitivePolicy: &SensitivePolicySpec{
				RedactFields: []string{"key_id", "message_id", "plainkey"},
			},
			REST: &RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "/api/decrypt/{message_id}",
				ContentType: "application/json",
				MaxBytes:    1024,
				BodySchema:  json.RawMessage(`{"type":"object","required":["key_id"],"properties":{"key_id":{"type":"string"}}}`),
			},
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
			Method: http.MethodPost,
			Path:   "/api/decrypt/{message_id}",
			Operation: &SurfaceOperation{
				Model:            "sensitive_reverse_etl",
				Status:           "blocked",
				Risk:             "high",
				BlockedByDefault: true,
				Reason:           "fixture direct write",
			},
		}}},
	}
	req := connectors.OperationDirectWriteRequest{
		Operation:  "acme.decrypt",
		PathParams: map[string]string{"message_id": "fixture-message-id"},
		Body:       map[string]any{"key_id": "fixture-key-id"},
	}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest
	_, err = OperationDirectWrite(context.Background(), bundle, req, nil)
	if err == nil {
		t.Fatal("OperationDirectWrite error = nil, want provider rejection")
	}
	for _, sensitive := range []string{"fixture-key-id", "fixture-message-id", "fixture-plaintext-key", "server-token"} {
		if strings.Contains(err.Error(), sensitive) {
			t.Fatal("json_redacted direct-write error exposed sensitive request or response content")
		}
	}
}

func TestSecretSensitiveDirectWriteRequiresTypedConfirmation(t *testing.T) {
	op := OperationSpec{
		ID:              "acme.decrypt",
		Kind:            "rest_write",
		Summary:         "Decrypt one fixture key",
		Risk:            "high",
		Approval:        "plan-preview-confirm-execute",
		OutputPolicy:    "json_redacted",
		MutationClass:   "secret",
		SecretSensitive: true,
		SensitivePolicy: &SensitivePolicySpec{
			InputMode:    "env_or_stdin",
			Transform:    "none",
			ApprovalMode: "typed_confirmation",
		},
		REST: &RESTOperationSpec{Method: http.MethodPost, Path: "/api/decrypt", ContentType: "application/json", MaxBytes: 1024, BodySchema: json.RawMessage(`{"type":"object"}`)},
	}
	target := DestructiveTargetForOperation("acme", op)
	if !target.RequiresApproval() || target.Confirmation != connectors.ConfirmationKindDestructive {
		t.Fatalf("secret-sensitive direct-write target = %#v, want typed destructive confirmation", target)
	}
	bundle := Bundle{Name: "acme", HTTP: HTTPBase{URL: "https://example.invalid"}, Operations: []OperationSpec{op}, Surface: &APISurface{Endpoints: []SurfaceEndpoint{{Method: http.MethodPost, Path: "/api/decrypt", Operation: &SurfaceOperation{Model: "sensitive_reverse_etl", Status: "blocked", Risk: "high", BlockedByDefault: true, Reason: "fixture direct write"}}}}}
	metadata, err := OperationDirectWriteMetadata(bundle, op.ID)
	if err != nil {
		t.Fatalf("OperationDirectWriteMetadata: %v", err)
	}
	if metadata.ConfirmationChallenge != string(connectors.ConfirmationKindDestructive) {
		t.Fatalf("secret-sensitive direct-write confirmation = %q, want destructive", metadata.ConfirmationChallenge)
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
	} else if strings.Contains(err.Error(), "server-token") || !strings.Contains(err.Error(), "http 500") || !strings.Contains(err.Error(), "[redacted]") {
		t.Fatalf("OperationDirectWrite error = %q, want a status-bearing redacted response error", err)
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
