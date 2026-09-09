package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func multipartOperationBundle(t *testing.T, baseURL string) Bundle {
	t.Helper()
	bundle, err := Load(multipartRestWriteBundleFS(validMultipartRestWrite, "rest_write"), "acme")
	if err != nil {
		t.Fatalf("Load multipart operation bundle: %v", err)
	}
	// The reusable loader fixture includes stream authentication solely to make
	// the bundle otherwise complete. Direct-write loopback tests intentionally
	// exercise no credentials or external endpoint.
	bundle.HTTP = HTTPBase{URL: baseURL}
	return bundle
}

func multipartOperationRequest(dir, sourcePath string, payload []byte) connectors.OperationDirectWriteRequest {
	digest := sha256.Sum256(payload)
	return connectors.OperationDirectWriteRequest{
		Operation: "acme.attachments.create",
		Config: connectors.RuntimeConfig{
			ProjectDir:            dir,
			CredentialRevision:    "fixture-credential-revision",
			ConfigurationDigest:   "fixture-configuration-digest",
			WriteApprovalScope:    connectors.WriteApprovalScopeFixture,
			ApprovedPayloadSHA256: map[string]string{connectors.PayloadApprovalKey(0, "media_file_path"): hex.EncodeToString(digest[:])},
		},
		Body: map[string]any{
			"message":         "typed multipart fixture",
			"media_file_path": sourcePath,
		},
	}
}

func writeMultipartOperationSource(t *testing.T, dir, name string, payload []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatalf("WriteFile %s: %v", name, err)
	}
	return name
}

func TestOperationDirectWriteMultipartPreviewBindsEveryApprovedPayload(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		calls++
	}))
	defer server.Close()

	dir := t.TempDir()
	payload := []byte("typed multipart payload")
	path := writeMultipartOperationSource(t, dir, "media.txt", payload)
	bundle := multipartOperationBundle(t, server.URL)
	req := multipartOperationRequest(dir, path, payload)

	missingDigest := req
	missingDigest.Config.ApprovedPayloadSHA256 = nil
	if _, err := PreviewOperationDirectWrite(context.Background(), bundle, missingDigest, nil); err == nil || !strings.Contains(err.Error(), "approved payload digest") {
		t.Fatalf("PreviewOperationDirectWrite without approved file digest = %v, want rejection", err)
	}

	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	if calls != 0 {
		t.Fatalf("preview calls = %d, want 0", calls)
	}

	changedField := req
	changedField.Body = map[string]any{"message": "a different typed field", "media_file_path": path}
	changedFieldPreview, err := PreviewOperationDirectWrite(context.Background(), bundle, changedField, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite changed field: %v", err)
	}
	if changedFieldPreview.Digest == preview.Digest {
		t.Fatalf("field change preserved preview digest %q", preview.Digest)
	}

	otherPath := writeMultipartOperationSource(t, dir, "other-media.txt", payload)
	changedPath := req
	changedPath.Body = map[string]any{"message": "typed multipart fixture", "media_file_path": otherPath}
	changedPathPreview, err := PreviewOperationDirectWrite(context.Background(), bundle, changedPath, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite changed source path: %v", err)
	}
	if changedPathPreview.Digest == preview.Digest {
		t.Fatalf("source path change preserved preview digest %q", preview.Digest)
	}

	changedField.PreviewDigest = preview.Digest
	changedField.Approval = approvedEvidenceForPreview(t, preview)
	if _, err := OperationDirectWrite(context.Background(), bundle, changedField, nil); err == nil || !strings.Contains(err.Error(), "no longer matches") {
		t.Fatalf("OperationDirectWrite with stale preview = %v, want preview mismatch", err)
	}
	if calls != 0 {
		t.Fatalf("stale preview calls = %d, want 0", calls)
	}

	changedPath.PreviewDigest = preview.Digest
	changedPath.Approval = approvedEvidenceForPreview(t, preview)
	if _, err := OperationDirectWrite(context.Background(), bundle, changedPath, nil); err == nil || !strings.Contains(err.Error(), "no longer matches") {
		t.Fatalf("OperationDirectWrite with stale source path = %v, want preview mismatch", err)
	}
	if calls != 0 {
		t.Fatalf("stale source-path calls = %d, want 0", calls)
	}
}

func TestOperationDirectWriteMultipartRequiresApprovalAndDispatchesOnce(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/attachments" {
			t.Fatalf("path = %s, want /attachments", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
			t.Fatalf("Content-Type = %q, want multipart boundary", r.Header.Get("Content-Type"))
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		if got := r.MultipartForm.Value["message"]; len(got) != 1 || got[0] != "typed multipart fixture" {
			t.Fatalf("message = %#v, want declared field value", got)
		}
		files := r.MultipartForm.File["attachment"]
		if len(files) != 1 {
			t.Fatalf("attachment files = %d, want 1", len(files))
		}
		if files[0].Filename != "media.txt" {
			t.Fatalf("attachment filename = %q, want media.txt", files[0].Filename)
		}
		if !strings.HasPrefix(files[0].Header.Get("Content-Type"), "text/plain") {
			t.Fatalf("attachment content type = %q, want text/plain", files[0].Header.Get("Content-Type"))
		}
		file, err := files[0].Open()
		if err != nil {
			t.Fatalf("Open attachment: %v", err)
		}
		defer func() { _ = file.Close() }()
		got, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("ReadAll attachment: %v", err)
		}
		if string(got) != "typed multipart payload" {
			t.Fatalf("attachment bytes = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"server_value":"complete"}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	payload := []byte("typed multipart payload")
	path := writeMultipartOperationSource(t, dir, "media.txt", payload)
	bundle := multipartOperationBundle(t, server.URL)
	req := multipartOperationRequest(dir, path, payload)
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	if calls != 0 {
		t.Fatalf("preview calls = %d, want 0", calls)
	}

	req.PreviewDigest = preview.Digest
	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil || !strings.Contains(err.Error(), "approval") {
		t.Fatalf("OperationDirectWrite without approval = %v, want approval rejection", err)
	}
	if calls != 0 {
		t.Fatalf("unapproved calls = %d, want 0", calls)
	}

	req.Approval = approvedEvidenceForPreview(t, preview)
	result, err := OperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("OperationDirectWrite approved multipart: %v", err)
	}
	if calls != 1 {
		t.Fatalf("approved calls = %d, want exactly 1", calls)
	}
	if result.Status != http.StatusOK {
		t.Fatalf("status = %d, want %d", result.Status, http.StatusOK)
	}
	body, ok := result.Body.(map[string]any)
	if !ok || body["server_value"] != "complete" {
		t.Fatalf("complete response body = %#v", result.Body)
	}

	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil || !strings.Contains(err.Error(), "approval") {
		t.Fatalf("OperationDirectWrite replayed approval = %v, want approval rejection", err)
	}
	if calls != 1 {
		t.Fatalf("replayed approval calls = %d, want 1", calls)
	}
}

func TestOperationDirectWriteMultipartRejectsUnsafeSourcesBeforeNetwork(t *testing.T) {
	tests := []struct {
		name          string
		beforePreview func(t *testing.T, dir string, req *connectors.OperationDirectWriteRequest)
		afterPreview  func(t *testing.T, dir string, req *connectors.OperationDirectWriteRequest)
		wantErr       string
	}{
		{
			name: "file changed after preview",
			afterPreview: func(t *testing.T, dir string, req *connectors.OperationDirectWriteRequest) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(dir, "media.txt"), []byte("changed multipart payload"), 0o600); err != nil {
					t.Fatalf("WriteFile changed source: %v", err)
				}
			},
			wantErr: "changed since approval",
		},
		{
			name: "source removed after preview",
			afterPreview: func(t *testing.T, dir string, req *connectors.OperationDirectWriteRequest) {
				t.Helper()
				if err := os.Remove(filepath.Join(dir, "media.txt")); err != nil {
					t.Fatalf("Remove source: %v", err)
				}
			},
			wantErr: "no such file",
		},
		{
			name: "source escapes project root",
			beforePreview: func(t *testing.T, dir string, req *connectors.OperationDirectWriteRequest) {
				t.Helper()
				outside := filepath.Join(filepath.Dir(dir), "outside-media.txt")
				payload := []byte("typed multipart payload")
				if err := os.WriteFile(outside, payload, 0o600); err != nil {
					t.Fatalf("WriteFile outside source: %v", err)
				}
				t.Cleanup(func() { _ = os.Remove(outside) })
				req.Body["media_file_path"] = outside
				digest := sha256.Sum256(payload)
				req.Config.ApprovedPayloadSHA256[connectors.PayloadApprovalKey(0, "media_file_path")] = hex.EncodeToString(digest[:])
			},
			wantErr: "outside the project root",
		},
		{
			name: "per file cap is enforced",
			beforePreview: func(t *testing.T, dir string, req *connectors.OperationDirectWriteRequest) {
				t.Helper()
				payload := []byte(strings.Repeat("x", 2048))
				if err := os.WriteFile(filepath.Join(dir, "media.txt"), payload, 0o600); err != nil {
					t.Fatalf("WriteFile oversized source: %v", err)
				}
				digest := sha256.Sum256(payload)
				req.Config.ApprovedPayloadSHA256[connectors.PayloadApprovalKey(0, "media_file_path")] = hex.EncodeToString(digest[:])
			},
			wantErr: "exceeds limit",
		},
		{
			name: "media type allow list is enforced",
			beforePreview: func(t *testing.T, dir string, req *connectors.OperationDirectWriteRequest) {
				t.Helper()
				payload := []byte{0, 1, 2, 3, 4, 5}
				if err := os.WriteFile(filepath.Join(dir, "media.txt"), payload, 0o600); err != nil {
					t.Fatalf("WriteFile binary source: %v", err)
				}
				digest := sha256.Sum256(payload)
				req.Config.ApprovedPayloadSHA256[connectors.PayloadApprovalKey(0, "media_file_path")] = hex.EncodeToString(digest[:])
			},
			wantErr: "allowed media types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls int
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				calls++
			}))
			defer server.Close()

			dir := t.TempDir()
			payload := []byte("typed multipart payload")
			path := writeMultipartOperationSource(t, dir, "media.txt", payload)
			bundle := multipartOperationBundle(t, server.URL)
			req := multipartOperationRequest(dir, path, payload)
			if tt.beforePreview != nil {
				tt.beforePreview(t, dir, &req)
			}
			preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
			if err != nil {
				t.Fatalf("PreviewOperationDirectWrite: %v", err)
			}
			if tt.afterPreview != nil {
				tt.afterPreview(t, dir, &req)
			}
			req.PreviewDigest = preview.Digest
			req.Approval = approvedEvidenceForPreview(t, preview)
			if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantErr)) {
				t.Fatalf("OperationDirectWrite unsafe source error = %v, want %q", err, tt.wantErr)
			}
			if calls != 0 {
				t.Fatalf("unsafe source calls = %d, want 0", calls)
			}
		})
	}
}

func TestOperationDirectWriteMultipartRejectsAggregateOverflowBeforeNetwork(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		calls++
	}))
	defer server.Close()

	dir := t.TempDir()
	firstPayload := []byte("1234")
	secondPayload := []byte("5678")
	firstPath := writeMultipartOperationSource(t, dir, "first.txt", firstPayload)
	secondPath := writeMultipartOperationSource(t, dir, "second.txt", secondPayload)
	bundle := multipartOperationBundle(t, server.URL)
	op := &bundle.Operations[0]
	op.REST.BodySchema = json.RawMessage(`{
		"type": "object",
		"additionalProperties": false,
		"required": ["message", "media_file_path", "second_file_path"],
		"properties": {
			"message": {"type": "string"},
			"media_file_path": {"type": "string"},
			"second_file_path": {"type": "string"}
		}
	}`)
	op.REST.Multipart = &MultipartSpec{MaxBytes: 5, Parts: []MultipartPartSpec{
		{Name: "message", Type: "field", Field: "message", Required: true},
		{Name: "attachment", Type: "file", Field: "media_file_path", Required: true, MaxBytes: 4, ContentType: "text/plain", AllowedMediaTypes: []string{"text/plain"}},
		{Name: "second_attachment", Type: "file", Field: "second_file_path", Required: true, MaxBytes: 4, ContentType: "text/plain", AllowedMediaTypes: []string{"text/plain"}},
	}}
	req := multipartOperationRequest(dir, firstPath, firstPayload)
	secondDigest := sha256.Sum256(secondPayload)
	req.Body["second_file_path"] = secondPath
	req.Config.ApprovedPayloadSHA256[connectors.PayloadApprovalKey(0, "second_file_path")] = hex.EncodeToString(secondDigest[:])
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest
	req.Approval = approvedEvidenceForPreview(t, preview)
	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil || !strings.Contains(err.Error(), "multipart payload too large") {
		t.Fatalf("OperationDirectWrite aggregate overflow error = %v, want aggregate rejection", err)
	}
	if calls != 0 {
		t.Fatalf("aggregate overflow calls = %d, want 0", calls)
	}
}

func TestOperationDirectWriteMultipartBoundsResponseCapture(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_, _ = w.Write([]byte("response beyond the declared cap"))
	}))
	defer server.Close()

	dir := t.TempDir()
	payload := []byte("typed multipart payload")
	path := writeMultipartOperationSource(t, dir, "media.txt", payload)
	bundle := multipartOperationBundle(t, server.URL)
	bundle.Operations[0].REST.MaxBytes = 4
	req := multipartOperationRequest(dir, path, payload)
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest
	req.Approval = approvedEvidenceForPreview(t, preview)
	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil || !strings.Contains(err.Error(), "response too large") {
		t.Fatalf("OperationDirectWrite over-cap response error = %v, want response cap rejection", err)
	}
	if calls != 1 {
		t.Fatalf("over-cap response calls = %d, want 1", calls)
	}
}

func TestOperationDirectWriteMultipartNeverRetriesOrReplaysRedirect(t *testing.T) {
	for _, status := range []int{http.StatusTooManyRequests, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var calls int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls++
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"error":"provider error content"}`))
			}))
			defer server.Close()

			dir := t.TempDir()
			payload := []byte("typed multipart payload")
			path := writeMultipartOperationSource(t, dir, "media.txt", payload)
			bundle := multipartOperationBundle(t, server.URL)
			req := multipartOperationRequest(dir, path, payload)
			preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
			if err != nil {
				t.Fatalf("PreviewOperationDirectWrite: %v", err)
			}
			req.PreviewDigest = preview.Digest
			req.Approval = approvedEvidenceForPreview(t, preview)
			result, err := OperationDirectWrite(context.Background(), bundle, req, nil)
			if err == nil || strings.Contains(err.Error(), "provider error content") {
				t.Fatalf("OperationDirectWrite status %d error = %v, want secret-safe diagnostic", status, err)
			}
			if !result.ResponseReceived || result.Status != status || result.BodyRaw != `{"error":"provider error content"}` {
				t.Fatalf("OperationDirectWrite status %d result = %#v, want complete provider response", status, result)
			}
			if calls != 1 {
				t.Fatalf("status %d calls = %d, want 1", status, calls)
			}
		})
	}

	var calls, redirectedCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path == "/attachments" {
			http.Redirect(w, r, "/redirected", http.StatusTemporaryRedirect)
			return
		}
		if r.URL.Path == "/redirected" {
			redirectedCalls++
			w.WriteHeader(http.StatusOK)
			return
		}
		t.Fatalf("unexpected path %q", r.URL.Path)
	}))
	defer server.Close()

	dir := t.TempDir()
	payload := []byte("typed multipart payload")
	path := writeMultipartOperationSource(t, dir, "media.txt", payload)
	bundle := multipartOperationBundle(t, server.URL)
	req := multipartOperationRequest(dir, path, payload)
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest
	req.Approval = approvedEvidenceForPreview(t, preview)
	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil {
		t.Fatal("OperationDirectWrite redirect error = nil, want redirect refusal")
	}
	if calls != 1 || redirectedCalls != 0 {
		t.Fatalf("redirect calls = total %d / followed %d, want 1 / 0", calls, redirectedCalls)
	}
}
