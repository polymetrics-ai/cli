package engine

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

const validMultipartRestWrite = `{
	"method": "POST",
	"path": "/attachments",
	"content_type": "multipart/form-data",
	"max_bytes": 1024,
	"body_schema": {
		"type": "object",
		"additionalProperties": false,
		"required": ["message", "media_file_path"],
		"properties": {
			"message": {"type": "string"},
			"media_file_path": {"type": "string"}
		}
	},
	"multipart": {
		"max_bytes": 2048,
		"parts": [
			{"name": "message", "type": "field", "field": "message", "required": true},
			{
				"name": "attachment",
				"type": "file",
				"field": "media_file_path",
				"required": true,
				"max_bytes": 1024,
				"content_type": "text/plain",
				"allowed_media_types": ["text/plain"]
			}
		]
	}
}`

func multipartRestWriteBundleFS(rest, kind string) fstest.MapFS {
	fsys := fullValidBundleFS("acme")
	fsys["acme/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"endpoints": [{
			"method": "POST",
			"path": "/attachments",
			"operation": {
				"model": "destructive_action",
				"status": "blocked",
				"risk": "high",
				"blocked_by_default": true,
				"reason": "operation metadata is bound by the executor"
			}
		}]
	}`)}
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`{
		"operations": [{
			"id": "acme.attachments.create",
			"kind": %q,
			"summary": "Attach one declared local file",
			"risk": "high",
			"approval": "plan-preview-confirm-execute",
			"output_policy": "json",
			"mutation_class": "destructive",
			"confirmation": {"kind": "destructive"},
			"rest": %s
		}]
	}`, kind, rest))}
	return fsys
}

func TestBundleLoadAcceptsTypedMultipartRestWriteContract(t *testing.T) {
	_, err := Load(multipartRestWriteBundleFS(validMultipartRestWrite, "rest_write"), "acme")
	if err != nil {
		t.Fatalf("Load typed multipart rest_write: %v", err)
	}
}

func TestOperationDirectWriteMetadataRecognizesTypedMultipartRestWrite(t *testing.T) {
	bundle, err := Load(multipartRestWriteBundleFS(validMultipartRestWrite, "rest_write"), "acme")
	if err != nil {
		t.Fatalf("Load typed multipart rest_write: %v", err)
	}
	metadata, err := OperationDirectWriteMetadata(bundle, "acme.attachments.create")
	if err != nil {
		t.Fatalf("OperationDirectWriteMetadata typed multipart rest_write: %v", err)
	}
	if len(metadata.PayloadFileFields) != 1 || metadata.PayloadFileFields[0] != "media_file_path" {
		t.Fatalf("multipart payload file fields = %#v, want the declared source path", metadata.PayloadFileFields)
	}
}

// TestOperationDirectWriteMultipartValidatesDeclaredJSONFile proves that a
// provider's `.json only` upload rule is enforced as an actual file contract,
// rather than weakened to text/plain merely because MIME sniffing has no JSON
// magic signature. Zoom Tasks adopts this bounded declaration.
func TestOperationDirectWriteMultipartValidatesDeclaredJSONFile(t *testing.T) {
	rest := strings.Replace(
		validMultipartRestWrite,
		`"content_type": "text/plain",
				"allowed_media_types": ["text/plain"]`,
		`"content_type": "application/json",
				"content_validation": "json",
				"allowed_file_extensions": [".json"]`,
		1,
	)
	bundle, err := Load(multipartRestWriteBundleFS(rest, "rest_write"), "acme")
	if err != nil {
		t.Fatalf("Load declared JSON multipart contract: %v", err)
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		if err := request.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		files := request.MultipartForm.File["attachment"]
		if len(files) != 1 || files[0].Filename != "valid.json" {
			t.Fatalf("multipart JSON file = %#v, want valid.json", files)
		}
		if !strings.HasPrefix(files[0].Header.Get("Content-Type"), "application/json") {
			t.Fatalf("multipart JSON content type = %q, want application/json", files[0].Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"file_id":"fixture-json-file"}`))
	}))
	defer server.Close()
	bundle.HTTP = HTTPBase{URL: server.URL}

	dir := t.TempDir()
	validPayload := []byte(`{"records":[{"title":"fixture"}]}`)
	validPath := writeMultipartOperationSource(t, dir, "valid.json", validPayload)
	valid := multipartOperationRequest(dir, validPath, validPayload)
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, valid, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite valid JSON: %v", err)
	}
	valid.PreviewDigest = preview.Digest
	valid.Approval = approvedEvidenceForPreview(t, preview)
	if result, err := OperationDirectWrite(context.Background(), bundle, valid, nil); err != nil || result.Status != http.StatusCreated {
		t.Fatalf("OperationDirectWrite valid JSON = %#v, %v; want 201", result, err)
	}

	for _, invalid := range []struct {
		name    string
		file    string
		payload []byte
		want    string
	}{
		{name: "rejects a non-json extension", file: "valid.txt", payload: validPayload, want: "allowed file extensions"},
		{name: "rejects malformed JSON", file: "invalid.json", payload: []byte(`{"records":`), want: "valid JSON"},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			path := writeMultipartOperationSource(t, dir, invalid.file, invalid.payload)
			req := multipartOperationRequest(dir, path, invalid.payload)
			preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
			if err != nil {
				t.Fatalf("PreviewOperationDirectWrite %s: %v", invalid.name, err)
			}
			req.PreviewDigest = preview.Digest
			req.Approval = approvedEvidenceForPreview(t, preview)
			if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil || !strings.Contains(err.Error(), invalid.want) {
				t.Fatalf("OperationDirectWrite %s error = %v, want %q", invalid.name, err, invalid.want)
			}
		})
	}
	if requests != 1 {
		t.Fatalf("JSON multipart fixture requests = %d, want only valid upload", requests)
	}
}

// TestOperationDirectWriteMultipartAcceptsDeclaredCSVFile captures the source
// format contract used by Zoom Workforce Management imports. A filename suffix
// alone is not an honest CSV guarantee; the bounded multipart policy must
// recognize a declaration-owned CSV validator before the connector can expose
// either published upload.
func TestOperationDirectWriteMultipartAcceptsDeclaredCSVFile(t *testing.T) {
	rest := strings.Replace(
		validMultipartRestWrite,
		`"content_type": "text/plain",
				"allowed_media_types": ["text/plain"]`,
		`"content_type": "text/csv",
				"content_validation": "csv",
				"allowed_file_extensions": [".csv"]`,
		1,
	)
	if _, err := Load(multipartRestWriteBundleFS(rest, "rest_write"), "acme"); err != nil {
		t.Fatalf("Load declared CSV multipart contract: %v", err)
	}
}

// TestOperationDirectWriteMultipartFollowsDeclaredRedirect captures the
// provider-required exception to ordinary direct-write redirect refusal. The
// declaration is closed: only the declared localhost suffix may receive the
// rebuilt snapshot-bound body, and the operation bearer must be re-applied at
// that approved host. Zoom Tasks adopts the same shape for its fileapi upload.
func TestOperationDirectWriteMultipartFollowsDeclaredRedirect(t *testing.T) {
	const accessToken = "fixture-declared-redirect-token"
	var redirectURL string
	var initialCalls, redirectedCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("declared multipart redirect did not retain the operation bearer")
		}
		switch request.URL.Path {
		case "/attachments":
			initialCalls++
			http.Redirect(w, request, redirectURL, http.StatusTemporaryRedirect)
		case "/redirected":
			redirectedCalls++
			if !strings.HasPrefix(request.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
				t.Fatal("redirected multipart request did not retain its declared body format")
			}
			if err := request.ParseMultipartForm(1 << 20); err != nil {
				t.Fatalf("parse redirected multipart form: %v", err)
			}
			if got := request.FormValue("message"); got != "typed multipart fixture" {
				t.Fatalf("redirected multipart field = %q, want declared value", got)
			}
			files := request.MultipartForm.File["attachment"]
			if len(files) != 1 || files[0].Filename != "media.txt" {
				t.Fatalf("redirected multipart file = %#v, want the declared snapshot", files)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			t.Fatalf("unexpected declared redirect path %q", request.URL.Path)
		}
	}))
	defer server.Close()
	baseURL := strings.Replace(server.URL, "127.0.0.1", "localhost", 1)
	redirectURL = baseURL + "/redirected"

	rest := strings.Replace(validMultipartRestWrite, `"path": "/attachments",`, fmt.Sprintf(`"path": "/attachments", "base_url": %q, "auth": [{"mode": "bearer", "token": "{{ secrets.fixture_token }}"}], "redirect": {"allowed_host_suffixes": ["localhost"], "max_hops": 1},`, baseURL), 1)
	bundle, err := Load(multipartRestWriteBundleFS(rest, "rest_write"), "acme")
	if err != nil {
		t.Fatalf("Load declared multipart redirect contract: %v", err)
	}

	dir := t.TempDir()
	payload := []byte("typed multipart payload")
	path := writeMultipartOperationSource(t, dir, "media.txt", payload)
	req := multipartOperationRequest(dir, path, payload)
	req.Config.Secrets = map[string]string{"fixture_token": accessToken}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite declared redirect: %v", err)
	}
	req.PreviewDigest = preview.Digest
	req.Approval = approvedEvidenceForPreview(t, preview)
	result, err := OperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("OperationDirectWrite declared redirect: %v", err)
	}
	if result.Status != http.StatusOK || initialCalls != 1 || redirectedCalls != 1 {
		t.Fatalf("declared redirect result/status/calls = %#v/%d/%d, want one admitted redirect", result, initialCalls, redirectedCalls)
	}
}

func TestBundleLoadRejectsUnsafeDeclaredMultipartRedirect(t *testing.T) {
	const (
		baseURL = "https://files.example.test/v2"
		bearer  = `[{"mode": "bearer", "token": "{{ secrets.fixture_token }}"}]`
		basic   = `[{"mode": "basic", "username": "fixture", "password": "{{ secrets.fixture_token }}"}]`
		valid   = `{"allowed_host_suffixes": ["example.test"], "max_hops": 1}`
	)
	withRedirect := func(replacement string) string {
		return strings.Replace(validMultipartRestWrite, `"path": "/attachments",`, replacement, 1)
	}
	cases := []struct {
		name string
		rest string
		want string
	}{
		{
			name: "requires fixed operation base and auth",
			rest: withRedirect(`"path": "/attachments", "redirect": {"allowed_host_suffixes": ["example.test"], "max_hops": 1},`),
			want: "base_url",
		},
		{
			name: "refuses configurable base",
			rest: withRedirect(`"path": "/attachments", "base_url": "{{ config.upload_base_url }}", "auth": ` + bearer + `, "redirect": ` + valid + `,`),
			want: "fixed literal",
		},
		{
			name: "requires bearer auth",
			rest: withRedirect(`"path": "/attachments", "base_url": "` + baseURL + `", "auth": ` + basic + `, "redirect": ` + valid + `,`),
			want: "bearer",
		},
		{
			name: "rejects wildcard host suffix",
			rest: withRedirect(`"path": "/attachments", "base_url": "` + baseURL + `", "auth": ` + bearer + `, "redirect": {"allowed_host_suffixes": ["*.example.test"], "max_hops": 1},`),
			want: "host suffix",
		},
		{
			name: "requires the operation base to be inside the provider boundary",
			rest: withRedirect(`"path": "/attachments", "base_url": "https://files.unapproved.test/v2", "auth": ` + bearer + `, "redirect": ` + valid + `,`),
			want: "base_url host",
		},
		{
			name: "rejects an IP address as a provider suffix",
			rest: withRedirect(`"path": "/attachments", "base_url": "` + baseURL + `", "auth": ` + bearer + `, "redirect": {"allowed_host_suffixes": ["127.0.0.1"], "max_hops": 1},`),
			want: "host suffix",
		},
		{
			name: "requires finite positive hop cap",
			rest: withRedirect(`"path": "/attachments", "base_url": "` + baseURL + `", "auth": ` + bearer + `, "redirect": {"allowed_host_suffixes": ["example.test"], "max_hops": 4},`),
			want: "max_hops",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(multipartRestWriteBundleFS(tt.rest, "rest_write"), "acme")
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.want)) {
				t.Fatalf("Load declared redirect error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestOperationDirectWriteMultipartRedirectFailsClosed(t *testing.T) {
	const accessToken = "fixture-redirect-boundary-token"
	tests := []struct {
		name           string
		status         int
		locationHost   string
		maxHops        int
		secondRedirect bool
		want           string
	}{
		{
			name:         "unapproved host",
			status:       http.StatusTemporaryRedirect,
			locationHost: "127.0.0.1",
			maxHops:      1,
			want:         "not approved",
		},
		{
			name:         "method changing status",
			status:       http.StatusFound,
			locationHost: "localhost",
			maxHops:      1,
			want:         "method-preserving",
		},
		{
			name:           "hop cap",
			status:         http.StatusTemporaryRedirect,
			locationHost:   "localhost",
			maxHops:        1,
			secondRedirect: true,
			want:           "hop cap",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var redirectURL string
			var initialCalls, redirectedCalls, finalCalls int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				if request.Header.Get("Authorization") != "Bearer "+accessToken {
					t.Fatal("redirect boundary fixture did not receive the declared bearer")
				}
				switch request.URL.Path {
				case "/attachments":
					initialCalls++
					http.Redirect(w, request, redirectURL, tt.status)
				case "/redirected":
					redirectedCalls++
					if !tt.secondRedirect {
						finalCalls++
						return
					}
					http.Redirect(w, request, redirectURL+"?fixture-signed-redirect-value", http.StatusTemporaryRedirect)
				default:
					t.Fatalf("unexpected redirect boundary request path %q", request.URL.Path)
				}
			}))
			defer server.Close()
			baseURL := strings.Replace(server.URL, "127.0.0.1", "localhost", 1)
			redirectURL = strings.Replace(baseURL, "localhost", tt.locationHost, 1) + "/redirected"

			rest := strings.Replace(validMultipartRestWrite, `"path": "/attachments",`, fmt.Sprintf(`"path": "/attachments", "base_url": %q, "auth": [{"mode": "bearer", "token": "{{ secrets.fixture_token }}"}], "redirect": {"allowed_host_suffixes": ["localhost"], "max_hops": %d},`, baseURL, tt.maxHops), 1)
			bundle, err := Load(multipartRestWriteBundleFS(rest, "rest_write"), "acme")
			if err != nil {
				t.Fatalf("Load redirect boundary contract: %v", err)
			}

			dir := t.TempDir()
			payload := []byte("typed multipart payload")
			path := writeMultipartOperationSource(t, dir, "media.txt", payload)
			req := multipartOperationRequest(dir, path, payload)
			req.Config.Secrets = map[string]string{"fixture_token": accessToken}
			preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
			if err != nil {
				t.Fatalf("PreviewOperationDirectWrite redirect boundary: %v", err)
			}
			req.PreviewDigest = preview.Digest
			req.Approval = approvedEvidenceForPreview(t, preview)
			_, err = OperationDirectWrite(context.Background(), bundle, req, nil)
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.want)) {
				t.Fatalf("OperationDirectWrite redirect boundary error = %v, want %q", err, tt.want)
			}
			if strings.Contains(err.Error(), "fixture-signed-redirect-value") {
				t.Fatal("declared redirect error exposed a signed target value")
			}
			if initialCalls != 1 || finalCalls != 0 {
				t.Fatalf("redirect boundary initial/final calls = %d/%d, want 1/0", initialCalls, finalCalls)
			}
			if tt.secondRedirect && redirectedCalls != 1 {
				t.Fatalf("hop-cap fixture redirected calls = %d, want 1", redirectedCalls)
			}
			if !tt.secondRedirect && redirectedCalls != 0 {
				t.Fatalf("rejected redirect target received %d unexpected calls", redirectedCalls)
			}
		})
	}
}

func TestBundleLoadRejectsUnsafeMultipartRestWriteContracts(t *testing.T) {
	tests := []struct {
		name    string
		kind    string
		rest    string
		wantErr string
	}{
		{
			name:    "rest read cannot declare multipart",
			kind:    "rest_read",
			rest:    validMultipartRestWrite,
			wantErr: "multipart",
		},
		{
			name:    "provider search cannot declare multipart",
			kind:    "provider_search",
			rest:    validMultipartRestWrite,
			wantErr: "multipart",
		},
		{
			name: "content type must be literal multipart form data",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"content_type": "multipart/form-data"`,
				`"content_type": "multipart/form-data; boundary=caller-controlled"`,
				1,
			),
			wantErr: "multipart/form-data",
		},
		{
			name: "endpoint must be connector relative",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"path": "/attachments"`,
				`"path": "https://untrusted.example/attachments"`,
				1,
			),
			wantErr: "connector-relative",
		},
		{
			name: "response capture must be bounded separately",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"max_bytes": 1024,`,
				`"max_bytes": 0,`,
				1,
			),
			wantErr: "response",
		},
		{
			name: "aggregate upload cap is required",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"max_bytes": 2048,`,
				`"max_bytes": 0,`,
				1,
			),
			wantErr: "aggregate",
		},
		{
			name: "parts cannot be empty",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"parts": [
			{"name": "message", "type": "field", "field": "message", "required": true},
			{
				"name": "attachment",
				"type": "file",
				"field": "media_file_path",
				"required": true,
				"max_bytes": 1024,
				"content_type": "text/plain",
				"allowed_media_types": ["text/plain"]
			}
		]`,
				`"parts": []`,
				1,
			),
			wantErr: "parts",
		},
		{
			name: "body schema must be closed",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"additionalProperties": false`,
				`"additionalProperties": true`,
				1,
			),
			wantErr: "additionalProperties",
		},
		{
			name: "body schema is required",
			kind: "rest_write",
			rest: `{
				"method": "POST",
				"path": "/attachments",
				"content_type": "multipart/form-data",
				"max_bytes": 1024,
				"multipart": {
					"max_bytes": 2048,
					"parts": [{
						"name": "attachment",
						"type": "file",
						"field": "media_file_path",
						"required": true,
						"max_bytes": 1024
					}]
				}
			}`,
			wantErr: "body_schema",
		},
		{
			name: "every part names a declared body field",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"field": "message"`,
				`"field": "unknown"`,
				1,
			),
			wantErr: "declared body field",
		},
		{
			name: "file source must be a required string",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"required": ["message", "media_file_path"]`,
				`"required": ["message"]`,
				1,
			),
			wantErr: "required string",
		},
		{
			name: "inline bytes cannot substitute for a file source path",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"media_file_path": {"type": "string"}`,
				`"media_file_path": {"type": "array", "items": {"type": "integer"}}`,
				1,
			),
			wantErr: "required string",
		},
		{
			name: "file source must have a positive cap",
			kind: "rest_write",
			rest: strings.Replace(
				validMultipartRestWrite,
				`"max_bytes": 1024,
				"content_type": "text/plain"`,
				`"max_bytes": 0,
				"content_type": "text/plain"`,
				1,
			),
			wantErr: "file part",
		},
		{
			name:    "legacy file upload does not become an operation executor",
			kind:    "file_upload",
			rest:    validMultipartRestWrite,
			wantErr: "file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(multipartRestWriteBundleFS(tt.rest, tt.kind), "acme")
			if err == nil {
				t.Fatal("Load unsafe multipart declaration: error = nil")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantErr)) {
				t.Fatalf("Load unsafe multipart declaration error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
