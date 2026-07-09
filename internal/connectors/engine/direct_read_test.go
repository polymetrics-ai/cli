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

func TestDirectReadExecutesFixedGETOperation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/repos/octo/hello/contents/docs/README.md" {
			t.Fatalf("path = %s, want contents path", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"README.md","type":"file"}`))
	}))
	defer srv.Close()

	result, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/repos/{owner}/{repo}/contents/{path}"), connectors.DirectReadRequest{
		Method: http.MethodGet,
		Path:   "/repos/{owner}/{repo}/contents/{path}",
		Config: connectors.RuntimeConfig{Config: map[string]string{
			"owner": "octo",
			"repo":  "hello",
		}},
		PathParams:   map[string]string{"path": "docs/README.md"},
		MaxBytes:     1024,
		OutputPolicy: "github_contents_file_metadata",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if result.Status != http.StatusOK {
		t.Fatalf("status = %d, want 200", result.Status)
	}
	body, ok := result.Body.(map[string]any)
	if !ok {
		t.Fatalf("body type = %T, want map", result.Body)
	}
	if body["name"] != "README.md" {
		t.Fatalf("body name = %v, want README.md", body["name"])
	}
}

func TestDirectReadResolvesPathWithConfigDefaults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/octo/hello/contents/README.md" {
			t.Fatalf("path = %s, want defaulted owner/repo path", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"README.md","type":"file"}`))
	}))
	defer srv.Close()

	b := directReadBundle(srv.URL, http.MethodGet, "/repos/{owner}/{repo}/contents/{path}")
	spec, err := CompileSchema(json.RawMessage(`{
		"type": "object",
		"properties": {
			"owner": {"type": "string", "default": "octo"},
			"repo": {"type": "string", "default": "hello"}
		}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	b.Spec = spec

	_, err = DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/repos/{owner}/{repo}/contents/{path}",
		Config:       connectors.RuntimeConfig{Config: map[string]string{}},
		PathParams:   map[string]string{"path": "README.md"},
		OutputPolicy: "github_contents_file_metadata",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
}

func TestDirectReadRejectsAbsoluteURL(t *testing.T) {
	_, err := DirectRead(context.Background(), directReadBundle("https://api.github.test", http.MethodGet, "https://evil.example.test/repos/{owner}/{repo}"), connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "https://evil.example.test/repos/{owner}/{repo}",
		Config:       connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}},
		OutputPolicy: "github_contents_file_metadata",
	}, nil)
	if err == nil {
		t.Fatal("DirectRead error = nil, want absolute URL rejection")
	}
	if !strings.Contains(err.Error(), "absolute URL") {
		t.Fatalf("DirectRead error = %q, want absolute URL", err.Error())
	}
}

func TestDirectReadRejectsMutationMethod(t *testing.T) {
	_, err := DirectRead(context.Background(), directReadBundle("https://api.github.test", http.MethodPost, "/repos/{owner}/{repo}/contents/{path}"), connectors.DirectReadRequest{
		Method:       http.MethodPost,
		Path:         "/repos/{owner}/{repo}/contents/{path}",
		Config:       connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}},
		OutputPolicy: "github_contents_file_metadata",
	}, nil)
	if err == nil {
		t.Fatal("DirectRead error = nil, want mutation rejection")
	}
	if !strings.Contains(err.Error(), "GET") {
		t.Fatalf("DirectRead error = %q, want GET", err.Error())
	}
}

func TestDirectReadMissingPathVariableFailsBeforeNetwork(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits++
	}))
	defer srv.Close()

	_, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/repos/{owner}/{repo}/contents/{path}"), connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/repos/{owner}/{repo}/contents/{path}",
		Config:       connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}},
		OutputPolicy: "github_contents_file_metadata",
	}, nil)
	if err == nil {
		t.Fatal("DirectRead error = nil, want missing path variable")
	}
	if hits != 0 {
		t.Fatalf("server hits = %d, want 0", hits)
	}
}

func TestDirectReadRejectsPathTraversalBeforeNetwork(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits++
	}))
	defer srv.Close()

	_, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/repos/{owner}/{repo}/contents/{path}"), connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/repos/{owner}/{repo}/contents/{path}",
		Config:       connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}},
		PathParams:   map[string]string{"path": "../secret"},
		OutputPolicy: "github_contents_file_metadata",
	}, nil)
	if err == nil {
		t.Fatal("DirectRead error = nil, want path traversal rejection")
	}
	if hits != 0 {
		t.Fatalf("server hits = %d, want 0", hits)
	}
}

func TestDirectReadRejectsOversizedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"README.md","content":"too large"}`))
	}))
	defer srv.Close()

	_, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/repos/{owner}/{repo}/contents/{path}"), connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/repos/{owner}/{repo}/contents/{path}",
		Config:       connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}},
		PathParams:   map[string]string{"path": "README.md"},
		MaxBytes:     8,
		OutputPolicy: "github_contents_file_metadata",
	}, nil)
	if err == nil {
		t.Fatal("DirectRead error = nil, want oversized response")
	}
	if !strings.Contains(err.Error(), "response too large") {
		t.Fatalf("DirectRead error = %q, want response too large", err.Error())
	}
}

func TestDirectReadRedactsGitHubFileContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name":"README.md",
			"type":"file",
			"content":"U0VDUkVU",
			"download_url":"https://raw.example.test/octo/hello/README.md"
		}`))
	}))
	defer srv.Close()

	result, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/repos/{owner}/{repo}/contents/{path}"), connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/repos/{owner}/{repo}/contents/{path}",
		Config:       connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}},
		PathParams:   map[string]string{"path": "README.md"},
		MaxBytes:     1024,
		OutputPolicy: "github_contents_file_metadata",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	body, ok := result.Body.(map[string]any)
	if !ok {
		t.Fatalf("body type = %T, want map", result.Body)
	}
	if _, ok := body["content"]; ok {
		t.Fatalf("content was not redacted: %+v", body)
	}
	if _, ok := body["download_url"]; ok {
		t.Fatalf("download_url was not redacted: %+v", body)
	}
	if body["content_redacted"] != true || body["download_url_redacted"] != true {
		t.Fatalf("redaction markers = %+v, want content/download_url redacted", body)
	}
}

func TestDirectReadRejectsSensitiveRepositoryPathBeforeNetwork(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits++
	}))
	defer srv.Close()

	_, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/repos/{owner}/{repo}/contents/{path}"), connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/repos/{owner}/{repo}/contents/{path}",
		Config:       connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}},
		PathParams:   map[string]string{"path": ".env"},
		OutputPolicy: "github_contents_file_metadata",
	}, nil)
	if err == nil {
		t.Fatal("DirectRead error = nil, want sensitive path rejection")
	}
	if hits != 0 {
		t.Fatalf("server hits = %d, want 0", hits)
	}
	if !strings.Contains(err.Error(), "blocked") {
		t.Fatalf("DirectRead error = %q, want blocked", err.Error())
	}
}

func TestDirectReadDirectoryPolicyRejectsFileResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"README.md","type":"file","content":"U0VDUkVU"}`))
	}))
	defer srv.Close()

	_, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/repos/{owner}/{repo}/contents/{path}"), connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/repos/{owner}/{repo}/contents/{path}",
		Config:       connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}},
		PathParams:   map[string]string{"path": "README.md"},
		OutputPolicy: "github_contents_directory",
	}, nil)
	if err == nil {
		t.Fatal("DirectRead error = nil, want directory policy shape rejection")
	}
	if !strings.Contains(err.Error(), "directory listing array") {
		t.Fatalf("DirectRead error = %q, want directory listing array", err.Error())
	}
}

func TestDirectReadJSONRedactedPolicyRedactsSensitiveFieldsRecursively(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"call-1",
			"downloadMediaUrl":"https://media.example.test/download/call-1",
			"content":"raw sensitive content",
			"nested":{"apiToken":"secret-token","safe":"ok"},
			"items":[{"password":"secret","name":"visible"}]
		}`))
	}))
	defer srv.Close()

	result, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/calls/{id}"), connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/calls/{id}",
		PathParams:   map[string]string{"id": "call-1"},
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	body, ok := result.Body.(map[string]any)
	if !ok {
		t.Fatalf("body type = %T, want object", result.Body)
	}
	for _, key := range []string{"downloadMediaUrl", "content"} {
		if _, ok := body[key]; ok {
			t.Fatalf("%s was not redacted: %+v", key, body)
		}
	}
	if body["downloadMediaUrl_redacted"] != true || body["content_redacted"] != true {
		t.Fatalf("top-level redaction markers missing: %+v", body)
	}
	nested := body["nested"].(map[string]any)
	if _, ok := nested["apiToken"]; ok || nested["apiToken_redacted"] != true || nested["safe"] != "ok" {
		t.Fatalf("nested redaction = %+v, want apiToken redacted and safe preserved", nested)
	}
	items := body["items"].([]any)
	item := items[0].(map[string]any)
	if _, ok := item["password"]; ok || item["password_redacted"] != true || item["name"] != "visible" {
		t.Fatalf("array redaction = %+v, want password redacted and name preserved", item)
	}
}

func TestDirectReadAvoidsDoubleVersionPrefixWhenBaseURLAlreadyContainsVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/calls/call-1" {
			t.Fatalf("path = %s, want /v2/calls/call-1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"call-1"}`))
	}))
	defer srv.Close()

	_, err := DirectRead(context.Background(), directReadBundle(srv.URL+"/v2", http.MethodGet, "/v2/calls/{id}"), connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/v2/calls/{id}",
		PathParams:   map[string]string{"id": "call-1"},
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
}

func directReadBundle(baseURL, method, endpointPath string) Bundle {
	return Bundle{
		Name: "github",
		HTTP: HTTPBase{URL: baseURL},
		Surface: &APISurface{
			OperationLedgerVersion: 1,
			Endpoints: []SurfaceEndpoint{
				{
					Method: method,
					Path:   endpointPath,
					CoveredBy: &SurfaceCoverage{
						DirectRead: "repo read-file",
					},
				},
			},
		},
	}
}
