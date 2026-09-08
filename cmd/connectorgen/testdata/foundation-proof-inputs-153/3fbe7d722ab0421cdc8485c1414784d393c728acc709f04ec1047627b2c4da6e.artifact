package engine

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
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
		OutputPolicy: "repository_contents_file_metadata",
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

func TestDirectReadAllowsSlashBearingRefPathVariables(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/octo/hello/git/ref/heads/main" {
			t.Fatalf("path = %s, want slash-bearing Git ref path", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ref":"heads/main"}`))
	}))
	defer srv.Close()

	result, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/repos/{owner}/{repo}/git/ref/{ref}"), connectors.DirectReadRequest{
		Method: http.MethodGet,
		Path:   "/repos/{owner}/{repo}/git/ref/{ref}",
		Config: connectors.RuntimeConfig{Config: map[string]string{
			"owner": "octo",
			"repo":  "hello",
		}},
		PathParams:   map[string]string{"ref": "heads/main"},
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if result.Status != http.StatusOK {
		t.Fatalf("status = %d, want 200", result.Status)
	}
}

func TestDirectReadExecutesHyphenatedPathVariable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/enterprises/example/teams/team-alpha/organizations" {
			t.Fatalf("path = %s, want hyphenated path variable resolved", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"organization":"example"}`))
	}))
	defer srv.Close()

	result, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/enterprises/{enterprise}/teams/{enterprise-team}/organizations"), connectors.DirectReadRequest{
		Method: http.MethodGet,
		Path:   "/enterprises/{enterprise}/teams/{enterprise-team}/organizations",
		PathParams: map[string]string{
			"enterprise":      "example",
			"enterprise-team": "team-alpha",
		},
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	if result.Status != http.StatusOK {
		t.Fatalf("status = %d, want 200", result.Status)
	}
}

func TestDirectReadHTTPErrorKeepsProviderQueryAndBodyPrivate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Query().Get("trace"), "direct-read-fixture"; got != want {
			t.Fatalf("trace = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"diagnostic":"direct-read-fixture-body"}`))
	}))
	t.Cleanup(srv.Close)

	_, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/items/{id}"), connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/items/{id}",
		PathParams:   map[string]string{"id": "fixture-item"},
		Query:        map[string]string{"trace": "direct-read-fixture"},
		OutputPolicy: "json_redacted",
	}, nil)
	if err == nil {
		t.Fatal("DirectRead error = nil, want HTTP failure")
	}
	for _, secret := range []string{"trace=direct-read-fixture", "direct-read-fixture-body"} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("DirectRead error leaked %q: %q", secret, err.Error())
		}
	}
	if !strings.Contains(err.Error(), "http 400") || !strings.Contains(err.Error(), "/items/{id}") {
		t.Fatalf("DirectRead error = %q, want safe status and declaration identity", err.Error())
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
		OutputPolicy: "repository_contents_file_metadata",
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
		OutputPolicy: "repository_contents_file_metadata",
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
		OutputPolicy: "repository_contents_file_metadata",
	}, nil)
	if err == nil {
		t.Fatal("DirectRead error = nil, want mutation rejection")
	}
	if !strings.Contains(err.Error(), "GET") {
		t.Fatalf("DirectRead error = %q, want GET", err.Error())
	}
}

func TestDirectReadRejectsOperationOnlyResponsePoliciesBeforeNetwork(t *testing.T) {
	for _, policy := range []string{"none", "text"} {
		t.Run(policy, func(t *testing.T) {
			var hits int
			srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				hits++
			}))
			defer srv.Close()

			_, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/items"), connectors.DirectReadRequest{
				Method:       http.MethodGet,
				Path:         "/items",
				OutputPolicy: policy,
			}, nil)
			if err == nil || !strings.Contains(err.Error(), "operation-backed") {
				t.Fatalf("DirectRead error = %v, want operation-backed policy refusal", err)
			}
			if hits != 0 {
				t.Fatalf("server hits = %d, want 0", hits)
			}
		})
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
		OutputPolicy: "repository_contents_file_metadata",
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
		OutputPolicy: "repository_contents_file_metadata",
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
		OutputPolicy: "repository_contents_file_metadata",
	}, nil)
	if err == nil {
		t.Fatal("DirectRead error = nil, want oversized response")
	}
	if !strings.Contains(err.Error(), "response too large") {
		t.Fatalf("DirectRead error = %q, want response too large", err.Error())
	}
}

func TestDirectReadRedactsRepositoryFileContent(t *testing.T) {
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
		OutputPolicy: "repository_contents_file_metadata",
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

func TestDirectReadRepositoryContentsPolicyIsConnectorNeutral(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/acme/repository/files/src/main.go" {
			t.Fatalf("path = %s, want generic repository file endpoint", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"main.go","type":"file","content":"U0VDUkVU","download_url":"https://files.example.test/acme/main.go"}`))
	}))
	defer srv.Close()

	b := directReadBundle(srv.URL, http.MethodGet, "/projects/{project}/repository/files/{path}")
	b.Name = "artifact-host"

	result, err := DirectRead(context.Background(), b, connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/projects/{project}/repository/files/{path}",
		PathParams:   map[string]string{"project": "acme", "path": "src/main.go"},
		MaxBytes:     1024,
		OutputPolicy: "repository_contents_file_metadata",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	body, ok := result.Body.(map[string]any)
	if !ok {
		t.Fatalf("body type = %T, want map", result.Body)
	}
	if _, ok := body["content"]; ok || body["content_redacted"] != true {
		t.Fatalf("content was not redacted generically: %+v", body)
	}
	if _, ok := body["download_url"]; ok || body["download_url_redacted"] != true {
		t.Fatalf("download_url was not redacted generically: %+v", body)
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
		OutputPolicy: "repository_contents_file_metadata",
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

func TestDirectReadRejectsSensitiveRepositoryConfigPathBeforeNetwork(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits++
	}))
	defer srv.Close()

	_, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/repos/{owner}/{repo}/contents/{path}"), connectors.DirectReadRequest{
		Method: http.MethodGet,
		Path:   "/repos/{owner}/{repo}/contents/{path}",
		Config: connectors.RuntimeConfig{Config: map[string]string{
			"owner": "octo",
			"repo":  "hello",
			"path":  ".env",
		}},
		OutputPolicy: "repository_contents_file_metadata",
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

func TestDirectReadRepositoryPolicyRequiresPathVariableBeforeNetwork(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits++
	}))
	defer srv.Close()

	_, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/repos/{owner}/{repo}/contents/{file_path}"), connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/repos/{owner}/{repo}/contents/{file_path}",
		Config:       connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}},
		PathParams:   map[string]string{"file_path": ".env"},
		OutputPolicy: "repository_contents_file_metadata",
	}, nil)
	if err == nil {
		t.Fatal("DirectRead error = nil, want repository path variable rejection")
	}
	if hits != 0 {
		t.Fatalf("server hits = %d, want 0", hits)
	}
	if !strings.Contains(err.Error(), "{path}") {
		t.Fatalf("DirectRead error = %q, want {path}", err.Error())
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
		OutputPolicy: "repository_contents_directory",
	}, nil)
	if err == nil {
		t.Fatal("DirectRead error = nil, want directory policy shape rejection")
	}
	if !strings.Contains(err.Error(), "directory listing array") {
		t.Fatalf("DirectRead error = %q, want directory listing array", err.Error())
	}
}

func TestJSONOutputRedactionPreservesProviderValuesEqualToConfiguredCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"call-1",
			"downloadMediaUrl":"https://media.example.test/download/call-1",
			"content":"raw sensitive content",
			"nested":{"apiToken":"ordinary-occurrence-token","echo":"secret-token","safe":"ok"},
			"items":[{"password":"secret","name":"visible"}]
		}`))
	}))
	defer srv.Close()

	result, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/calls/{id}"), connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/calls/{id}",
		PathParams:   map[string]string{"id": "call-1"},
		Config:       connectors.RuntimeConfig{Secrets: map[string]string{"api_token": "secret-token"}},
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	body, ok := result.Body.(map[string]any)
	if !ok {
		t.Fatalf("body type = %T, want object", result.Body)
	}
	if body["downloadMediaUrl"] == nil || body["content"] != "raw sensitive content" {
		t.Fatalf("ordinary provider fields were removed: %+v", body)
	}
	nested := body["nested"].(map[string]any)
	if nested["apiToken"] != "ordinary-occurrence-token" || nested["echo"] != "secret-token" || nested["safe"] != "ok" {
		t.Fatalf("nested provider output = %+v, want undeclared values preserved", nested)
	}
	items := body["items"].([]any)
	item := items[0].(map[string]any)
	if item["password"] != "secret" || item["name"] != "visible" {
		t.Fatalf("array provider truth changed: %+v", item)
	}
}

func TestDirectReadClinicalJSONRedactedRequiresExplicitFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"record-1","value":"kept without explicit policy","patient":{"uuid":"kept"},"apiToken":"secret-token"}`))
	}))
	defer srv.Close()

	result, err := DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/records/{id}"), connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/records/{id}",
		PathParams:   map[string]string{"id": "record-1"},
		OutputPolicy: "clinical_json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead: %v", err)
	}
	body := result.Body.(map[string]any)
	if body["value"] != "kept without explicit policy" {
		t.Fatalf("value = %#v, want unchanged without explicit policy", body["value"])
	}
	if _, ok := body["patient"].(map[string]any); !ok {
		t.Fatalf("patient = %#v, want unchanged object without explicit policy", body["patient"])
	}
	if body["apiToken"] != "secret-token" {
		t.Fatalf("apiToken = %+v, want ordinary unclassified provider value preserved", body)
	}

	result, err = DirectRead(context.Background(), directReadBundle(srv.URL, http.MethodGet, "/records/{id}"), connectors.DirectReadRequest{
		Method:       http.MethodGet,
		Path:         "/records/{id}",
		PathParams:   map[string]string{"id": "record-1"},
		OutputPolicy: "clinical_json_redacted",
		RedactFields: []string{"value", "patient"},
	}, nil)
	if err != nil {
		t.Fatalf("DirectRead with explicit redact fields: %v", err)
	}
	body = result.Body.(map[string]any)
	if _, ok := body["value"]; ok || body["value_redacted"] != true {
		t.Fatalf("value redaction = %+v, want explicit redaction", body)
	}
	if _, ok := body["patient"]; ok || body["patient_redacted"] != true {
		t.Fatalf("patient redaction = %+v, want explicit redaction", body)
	}
}

func TestOperationDirectReadAppliesDeclaredSensitiveRedactFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if got := r.URL.Query().Get("q"); got != "synthetic" {
			t.Fatalf("q = %q, want synthetic", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"results":[{
				"uuid":"patient-opaque-ref",
				"identifier":"SYN-12345",
				"addressFieldValue":"Synthetic Village",
				"person":{"display":"Synthetic Person","birthdate":"1990-01-02"}
			}]
		}`))
	}))
	defer srv.Close()

	b := Bundle{
		Name: "bahmni",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID:              "bahmni.patient_search",
			Kind:            "rest_read",
			Summary:         "Search patients",
			Risk:            "high",
			Approval:        "none",
			OutputPolicy:    "json_redacted",
			SensitivePolicy: &SensitivePolicySpec{RedactFields: []string{"identifier", "addressFieldValue", "display", "birthdate"}},
			REST: &RESTOperationSpec{
				Method:     http.MethodGet,
				Path:       "/ws/rest/v1/bahmni/search/patient",
				MaxBytes:   1024,
				Parameters: []OperationParameter{{Name: "q", In: "query", Type: "string"}},
			},
		}},
	}

	result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "bahmni.patient_search",
		Query:     map[string]string{"q": "synthetic"},
		MaxBytes:  1024,
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	body := result.Body.(map[string]any)
	results := body["results"].([]any)
	patient := results[0].(map[string]any)
	for _, key := range []string{"identifier", "addressFieldValue"} {
		if _, ok := patient[key]; ok || patient[key+"_redacted"] != true {
			t.Fatalf("patient search body did not redact %s: %+v", key, patient)
		}
	}
	person := patient["person"].(map[string]any)
	for _, key := range []string{"display", "birthdate"} {
		if _, ok := person[key]; ok || person[key+"_redacted"] != true {
			t.Fatalf("patient person body did not redact %s: %+v", key, person)
		}
	}
	if patient["uuid"] != "patient-opaque-ref" {
		t.Fatalf("opaque patient uuid = %v, want retained", patient["uuid"])
	}
}

func TestOperationDirectReadPOSTJSONBodyValidatesAndPreservesUndeclaredProviderValue(t *testing.T) {
	var sawBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v2/meetings/integration/status" {
			t.Fatalf("path = %s, want /v2/meetings/integration/status", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&sawBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_, _ = w.Write([]byte(`{"ok":true,"apiToken":"secret-token"}`))
	}))
	defer srv.Close()

	b := Bundle{
		Name: "gong",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID:           "gong.meetings_integration_status",
			Kind:         "rest_read",
			Summary:      "Validate meeting integration",
			Risk:         "medium",
			Approval:     "none",
			OutputPolicy: "json_redacted",
			REST: &RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "/v2/meetings/integration/status",
				ContentType: "application/json",
				MaxBytes:    1024,
				BodySchema:  json.RawMessage(`{"type":"object","required":["emails"],"properties":{"emails":{"type":"array","items":{"type":"string"},"maxItems":100}},"additionalProperties":false}`),
			},
		}},
	}

	result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation:    "gong.meetings_integration_status",
		Config:       connectors.RuntimeConfig{Secrets: map[string]string{"api_token": "secret-token"}},
		Body:         map[string]any{"emails": []any{"ada@example.com"}},
		MaxBytes:     1024,
		OutputPolicy: "json_redacted",
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if emails, ok := sawBody["emails"].([]any); !ok || len(emails) != 1 || emails[0] != "ada@example.com" {
		t.Fatalf("request body = %+v, want emails array", sawBody)
	}
	body := result.Body.(map[string]any)
	if body["apiToken"] != "secret-token" {
		t.Fatalf("response body = %+v, want undeclared provider value preserved", body)
	}
}

func TestOperationDirectReadSupportsBoundedStatusAndTextResponses(t *testing.T) {
	t.Run("status only accepts an empty success body", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || r.URL.Path != "/membership" {
				t.Fatalf("request = %s %s, want GET /membership", r.Method, r.URL.Path)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		result, err := OperationDirectRead(context.Background(), statusTextReadBundle(srv.URL, "acme.membership", http.MethodGet, "/membership", "none", "", 1024), connectors.OperationDirectReadRequest{
			Operation: "acme.membership",
			MaxBytes:  1024,
		}, nil)
		if err != nil {
			t.Fatalf("OperationDirectRead: %v", err)
		}
		if result.Status != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", result.Status, http.StatusNoContent)
		}
		if result.Body != nil {
			t.Fatalf("body = %#v, want nil for a status-only response", result.Body)
		}
	})

	t.Run("status only rejects a nonempty success body", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("unexpected"))
		}))
		defer srv.Close()

		_, err := OperationDirectRead(context.Background(), statusTextReadBundle(srv.URL, "acme.membership", http.MethodGet, "/membership", "none", "", 1024), connectors.OperationDirectReadRequest{
			Operation: "acme.membership",
			MaxBytes:  1024,
		}, nil)
		if err == nil || !strings.Contains(err.Error(), "status-only response must be empty") {
			t.Fatalf("OperationDirectRead error = %v, want nonempty status-only response rejection", err)
		}
	})

	t.Run("text response is bounded and valid UTF-8", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/zen" {
				t.Fatalf("path = %q, want /zen", r.URL.Path)
			}
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("keep it simple"))
		}))
		defer srv.Close()

		result, err := OperationDirectRead(context.Background(), statusTextReadBundle(srv.URL, "acme.zen", http.MethodGet, "/zen", "text", "", 1024), connectors.OperationDirectReadRequest{
			Operation: "acme.zen",
			MaxBytes:  1024,
		}, nil)
		if err != nil {
			t.Fatalf("OperationDirectRead: %v", err)
		}
		if got, want := result.Body, any("keep it simple"); got != want {
			t.Fatalf("body = %#v, want %#v", got, want)
		}
	})

	t.Run("text response rejects invalid UTF-8", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte{0xff, 0xfe})
		}))
		defer srv.Close()

		_, err := OperationDirectRead(context.Background(), statusTextReadBundle(srv.URL, "acme.octets", http.MethodGet, "/octets", "text", "", 1024), connectors.OperationDirectReadRequest{
			Operation: "acme.octets",
			MaxBytes:  1024,
		}, nil)
		if err == nil || !strings.Contains(err.Error(), "valid UTF-8") {
			t.Fatalf("OperationDirectRead error = %v, want UTF-8 rejection", err)
		}
	})

	t.Run("text response retains its declared cap", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("five!"))
		}))
		defer srv.Close()

		_, err := OperationDirectRead(context.Background(), statusTextReadBundle(srv.URL, "acme.short", http.MethodGet, "/short", "text", "", 4), connectors.OperationDirectReadRequest{
			Operation: "acme.short",
			MaxBytes:  4,
		}, nil)
		if err == nil || !strings.Contains(err.Error(), "response too large") {
			t.Fatalf("OperationDirectRead error = %v, want response cap rejection", err)
		}
	})
}

func statusTextReadBundle(baseURL, id, method, endpointPath, outputPolicy, contentType string, maxBytes int) Bundle {
	return Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: baseURL},
		Operations: []OperationSpec{{
			ID:           id,
			Kind:         "rest_read",
			Summary:      "bounded fixture read",
			Risk:         "low",
			Approval:     "none",
			OutputPolicy: outputPolicy,
			REST: &RESTOperationSpec{
				Method:      method,
				Path:        endpointPath,
				ContentType: contentType,
				MaxBytes:    maxBytes,
			},
		}},
	}
}

func TestOperationDirectReadUsesOnlyDeclaredTypedHeaders(t *testing.T) {
	type connectorFixture struct {
		name   string
		header string
		path   string
	}
	fixtures := []connectorFixture{
		{name: "synthetic-alpha", header: "X-Alpha-Tenant", path: "/v1/alpha/{id}"},
		{name: "synthetic-beta", header: "X-Beta-Tenant", path: "/v1/beta/{id}"},
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			var issued int
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				issued++
				if got, want := r.Header.Get(fixture.header), "tenant-42"; got != want {
					t.Fatalf("%s = %q, want %q", fixture.header, got, want)
				}
				if got := r.Header.Get("X-Other-Operation"); got != "" {
					t.Fatalf("cross-operation header reached wire: %q", got)
				}
				if got, want := r.URL.Query().Get("query_only"), "query-value"; got != want {
					t.Fatalf("query_only = %q, want %q", got, want)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"provider_field":"preserved"}`))
			}))
			defer srv.Close()

			b := statusTextReadBundle(srv.URL, fixture.name+".read", http.MethodGet, fixture.path, "json_redacted", "", 1024)
			b.Name = fixture.name
			b.Operations[0].REST.Parameters = []OperationParameter{
				{Name: "id", In: "path", Type: "string", Required: true},
				{Name: "query_only", In: "query", Type: "string"},
				{
					Name:     fixture.header,
					In:       "header",
					Type:     "string",
					Required: true,
					Schema:   json.RawMessage(`{"type":"string","pattern":"^[a-z0-9-]+$","minLength":1,"maxLength":16}`),
					MaxBytes: 16,
				},
			}

			result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
				Operation:  fixture.name + ".read",
				PathParams: map[string]string{"id": "fixed-path"},
				Query:      map[string]string{"query_only": "query-value"},
				Headers:    map[string]string{fixture.header: "tenant-42"},
				MaxBytes:   1024,
			}, nil)
			if err != nil {
				t.Fatalf("OperationDirectRead: %v", err)
			}
			body, ok := result.Body.(map[string]any)
			if !ok || body["provider_field"] != "preserved" {
				t.Fatalf("body = %#v, want ordinary provider fields preserved", result.Body)
			}
			if issued != 1 {
				t.Fatalf("requests = %d, want 1", issued)
			}
		})
	}
}

func TestOperationDirectReadQueryMergesDeclaredAndRequestedValues(t *testing.T) {
	op := OperationSpec{
		ID:   "acme.query",
		Kind: "rest_read",
		REST: &RESTOperationSpec{
			Query: map[string]string{"fixed": "source-value"},
			Parameters: []OperationParameter{{
				Name: "requested",
				In:   "query",
				Type: "string",
			}},
		},
	}

	got, err := operationDirectReadQuery(op, map[string]string{"requested": "caller-value"}, nil)
	if err != nil {
		t.Fatalf("operationDirectReadQuery: %v", err)
	}
	want := map[string]string{"fixed": "source-value", "requested": "caller-value"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("query = %#v, want %#v", got, want)
	}
}

func TestOperationDirectReadRejectsHeaderEscapeHatchesBeforeNetwork(t *testing.T) {
	var issued int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		issued++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	b := statusTextReadBundle(srv.URL, "synthetic.header_read", http.MethodGet, "/v1/resource", "json_redacted", "", 1024)
	b.Operations[0].REST.Parameters = []OperationParameter{{
		Name:     "X-Declared-Tenant",
		In:       "header",
		Type:     "string",
		Required: true,
		Values:   []string{"alpha"},
		Schema:   json.RawMessage(`{"type":"string","pattern":"^[a-z]+$","minLength":1,"maxLength":8}`),
		MaxBytes: 8,
	}}
	b.Operations = append(b.Operations, OperationSpec{
		ID: "synthetic.other_operation", Kind: "rest_read", Summary: "other header namespace", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
		REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/v1/other", MaxBytes: 1024, Parameters: []OperationParameter{{Name: "X-Other-Operation", In: "header", Type: "string", Schema: json.RawMessage(`{"type":"string"}`), MaxBytes: 8}}},
	})
	b.HTTP.Headers = map[string]string{"X-Runtime-Header": "fixed-runtime-value"}

	tests := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{name: "unknown", headers: map[string]string{"X-Unknown": "alpha"}, want: "unknown declared request header"},
		{name: "cross operation", headers: map[string]string{"X-Other-Operation": "alpha"}, want: "unknown declared request header"},
		{name: "malformed name", headers: map[string]string{"X Declared Tenant": "alpha"}, want: "malformed"},
		{name: "duplicate case variant", headers: map[string]string{"X-Declared-Tenant": "alpha", "x-declared-tenant": "alpha"}, want: "duplicate"},
		{name: "protected authorization case variant", headers: map[string]string{"aUtHoRiZaTiOn": "Bearer forbidden"}, want: "protected"},
		{name: "protected proxy authorization case variant", headers: map[string]string{"PrOxY-AuThOrIzAtIoN": "forbidden"}, want: "protected"},
		{name: "protected cookie case variant", headers: map[string]string{"cOoKiE": "forbidden"}, want: "protected"},
		{name: "protected set cookie case variant", headers: map[string]string{"sEt-CoOkIe": "forbidden"}, want: "protected"},
		{name: "protected host case variant", headers: map[string]string{"hOsT": "forbidden.test"}, want: "protected"},
		{name: "protected connection case variant", headers: map[string]string{"cOnNeCtIoN": "close"}, want: "protected"},
		{name: "protected forwarding case variant", headers: map[string]string{"X-FoRwArDeD-FoR": "203.0.113.1"}, want: "protected"},
		{name: "protected API key underscore variant", headers: map[string]string{"X_Api_Key": "forbidden"}, want: "protected"},
		{name: "protected auth token underscore variant", headers: map[string]string{"x_auth_token": "forbidden"}, want: "protected"},
		{name: "protected forwarding underscore variant", headers: map[string]string{"x_forwarded_host": "forbidden.test"}, want: "protected"},
		{name: "protected method override underscore variant", headers: map[string]string{"X_HTTP_Method_Override": "DELETE"}, want: "protected"},
		{name: "runtime configured header", headers: map[string]string{"x-runtime-header": "caller-override"}, want: "protected"},
		{name: "CRLF injection", headers: map[string]string{"X-Declared-Tenant": "alpha\r\nX-Escaped: true"}, want: "control character"},
		{name: "over declared byte cap", headers: map[string]string{"X-Declared-Tenant": "abcdefghx"}, want: "exceeds declared byte cap"},
		{name: "schema mismatch", headers: map[string]string{"X-Declared-Tenant": "ALPHA"}, want: "does not satisfy declared schema"},
		{name: "enum mismatch", headers: map[string]string{"X-Declared-Tenant": "beta"}, want: "not one of the declared values"},
		{name: "missing required", headers: nil, want: "requires declared request header"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
				Operation: "synthetic.header_read", Headers: tt.headers, MaxBytes: 1024,
			}, nil)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("OperationDirectRead error = %v, want %q", err, tt.want)
			}
		})
	}
	if issued != 0 {
		t.Fatalf("requests = %d, want header rejections before I/O", issued)
	}
}

func TestOperationDirectReadSendsOnlyDeclaredRepeatableHeaderValues(t *testing.T) {
	var issued int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issued++
		if got := r.Header.Values("X-Mode"); len(got) != 2 || got[0] != "safe" || got[1] != "full" {
			t.Fatalf("X-Mode values = %#v, want ordered declared values", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer srv.Close()
	b := statusTextReadBundle(srv.URL, "synthetic.repeatable_header", http.MethodGet, "/v1/resource", "json_redacted", "", 1024)
	b.Operations[0].REST.Parameters = []OperationParameter{{
		Name: "X-Mode", In: "header", Type: "string", Repeatable: true, Values: []string{"safe", "full"}, Schema: json.RawMessage(`{"type":"string","enum":["safe","full"]}`), MaxBytes: 8,
	}}
	if _, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{Operation: "synthetic.repeatable_header", HeaderValues: map[string][]string{"X-Mode": {"invalid", "safe"}}, MaxBytes: 1024}, nil); err == nil || !strings.Contains(err.Error(), "does not satisfy declared schema") {
		t.Fatalf("invalid repeatable header error = %v", err)
	}
	if issued != 0 {
		t.Fatalf("invalid repeatable header reached provider %d times", issued)
	}
	if _, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{Operation: "synthetic.repeatable_header", HeaderValues: map[string][]string{"X-Mode": {"safe", "full"}}, MaxBytes: 1024}, nil); err != nil {
		t.Fatalf("OperationDirectRead repeatable header: %v", err)
	}
	if issued != 1 {
		t.Fatalf("requests = %d, want 1", issued)
	}
	b.Operations[0].REST.Parameters[0].Repeatable = false
	if _, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{Operation: "synthetic.repeatable_header", HeaderValues: map[string][]string{"X-Mode": {"safe", "full"}}, MaxBytes: 1024}, nil); err == nil || !strings.Contains(err.Error(), "exactly one value") {
		t.Fatalf("singleton repeated header error = %v", err)
	}
	if issued != 1 {
		t.Fatalf("singleton rejection reached provider %d times", issued)
	}
}

func TestOperationDirectReadPreservesDeclaredResponseFieldsAndMasksKnownSecrets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", "req-123")
		w.Header().Set("X-Provider-Tier", "rare-paid-feature")
		w.Header().Set("X-Provider-Key", "echoed-credential")
		w.Header().Set("X-Undeclared-Transport-Metadata", "not-an-output-channel")
		w.Header().Add("Set-Cookie", "session=transport-secret")
		_, _ = w.Write([]byte(`{"ordinary":"retained","rare":"retained","paid_tier":"retained","destructive_state":"retained"}`))
	}))
	t.Cleanup(srv.Close)
	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL, Auth: []AuthSpec{{Mode: "api_key_header", Header: "X-Provider-Key", Value: "fixture-key"}}},
		Operations: []OperationSpec{{
			ID: "acme.result", Kind: "rest_read", Summary: "Declared result", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
			REST: &RESTOperationSpec{
				Method: http.MethodGet, Path: "/v1/result", MaxBytes: 1024,
				Response: &OperationResponseSpec{Headers: []OperationResponseHeaderSpec{
					{Name: "X-Request-ID", MaxBytes: 64},
					{Name: "X-Provider-Tier", MaxBytes: 64},
					{Name: "X-Provider-Key", MaxBytes: 64},
					{Name: "Set-Cookie", MaxBytes: 128},
				}},
			},
		}},
	}

	result, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{Operation: "acme.result"}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	body, ok := result.Body.(map[string]any)
	if !ok || body["ordinary"] != "retained" || body["rare"] != "retained" || body["paid_tier"] != "retained" || body["destructive_state"] != "retained" {
		t.Fatalf("body = %#v, want every declared ordinary response field", result.Body)
	}
	if got := result.Headers["X-Request-ID"].Values; len(got) != 1 || got[0] != "req-123" {
		t.Fatalf("request id header = %#v, want complete declared metadata", got)
	}
	if got := result.Headers["X-Provider-Tier"].Values; len(got) != 1 || got[0] != "rare-paid-feature" {
		t.Fatalf("provider tier header = %#v, want complete declared metadata", got)
	}
	if cookie, ok := result.Headers["Set-Cookie"]; !ok || !reflect.DeepEqual(cookie.Values, []string{"session=transport-secret"}) {
		t.Fatalf("Set-Cookie = %#v, want exact internal provider metadata", cookie)
	}
	if providerKey, ok := result.Headers["X-Provider-Key"]; !ok || !reflect.DeepEqual(providerKey.Values, []string{"echoed-credential"}) {
		t.Fatalf("X-Provider-Key = %#v, want exact internal provider metadata", providerKey)
	}
	if _, present := result.Headers["X-Undeclared-Transport-Metadata"]; present {
		t.Fatalf("headers = %#v, undeclared provider metadata became an output channel", result.Headers)
	}
}

func TestOperationDirectReadPreservesConfiguredEqualResponseHeaderValues(t *testing.T) {
	const credential = "configured-header-material"
	const occurrenceID = "occurrence-9007199254740993"
	const unconfiguredToken = "ghp_unconfigured_provider_token"
	encodedCredential := base64.StdEncoding.EncodeToString([]byte(credential))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", occurrenceID)
		w.Header().Set("X-Provider-Token", unconfiguredToken)
		w.Header().Set("X-Configured-Secret", credential)
		w.Header().Set("X-Configured-Encoding", encodedCredential)
		_, _ = w.Write([]byte(`{"ordinary":"retained"}`))
	}))
	t.Cleanup(srv.Close)
	bundle := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID: "acme.result", Kind: "rest_read", Summary: "Declared result", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
			REST: &RESTOperationSpec{
				Method: http.MethodGet, Path: "/v1/result", MaxBytes: 1024,
				Response: &OperationResponseSpec{Headers: []OperationResponseHeaderSpec{
					{Name: "X-Request-ID", MaxBytes: 64},
					{Name: "X-Provider-Token", MaxBytes: 64},
					{Name: "X-Configured-Secret", MaxBytes: 64},
					{Name: "X-Configured-Encoding", MaxBytes: 64},
				}},
			},
		}},
	}

	result, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
		Operation: "acme.result",
		Config:    connectors.RuntimeConfig{Secrets: map[string]string{"credential": credential}},
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	for name, want := range map[string][]string{
		"X-Request-ID":          {occurrenceID},
		"X-Provider-Token":      {unconfiguredToken},
		"X-Configured-Secret":   {credential},
		"X-Configured-Encoding": {encodedCredential},
	} {
		header, ok := result.Headers[name]
		if !ok || header.Redacted || !reflect.DeepEqual(header.Values, want) {
			t.Fatalf("header %q = %#v, want values %#v without header-name redaction", name, header, want)
		}
	}
	public, err := json.Marshal(result.Headers)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(public, []byte(credential)) || !bytes.Contains(public, []byte(encodedCredential)) {
		t.Fatal("public direct-read headers did not preserve provider-returned configured-equal material")
	}
}

func TestOperationDirectReadSendsDeclaredPlainTextBody(t *testing.T) {
	source := "# rendered from a declared plain-text body"
	var issued int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issued++
		if r.Method != http.MethodPost || r.URL.Path != "/markdown/raw" {
			t.Fatalf("request = %s %s, want POST /markdown/raw", r.Method, r.URL.Path)
		}
		if got, want := r.Header.Get("Content-Type"), "text/plain"; got != want {
			t.Fatalf("Content-Type = %q, want %q", got, want)
		}
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if got := string(raw); got != source {
			t.Fatalf("request body = %q, want literal source %q", got, source)
		}
		_, _ = w.Write([]byte("<h1>rendered</h1>"))
	}))
	defer srv.Close()

	b := statusTextReadBundle(srv.URL, "github.markdown_raw", http.MethodPost, "/markdown/raw", "text", "text/plain", 1024)
	b.Operations[0].REST.BodySchema = json.RawMessage(`{"type":"string"}`)
	result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "github.markdown_raw",
		RawBody:   &source,
		MaxBytes:  1024,
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if got, want := result.Body, any("<h1>rendered</h1>"); got != want {
		t.Fatalf("body = %#v, want %#v", got, want)
	}
	if issued != 1 {
		t.Fatalf("requests = %d, want 1", issued)
	}
}

func TestOperationDirectReadRejectsUndeclaredOrInvalidPlainTextBodiesBeforeNetwork(t *testing.T) {
	var issued int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		issued++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	plainBundle := func() Bundle {
		b := statusTextReadBundle(srv.URL, "github.markdown_raw", http.MethodPost, "/markdown/raw", "text", "text/plain", 4)
		b.Operations[0].REST.BodySchema = json.RawMessage(`{"type":"string"}`)
		return b
	}
	tests := []struct {
		name   string
		bundle func() Bundle
		req    connectors.OperationDirectReadRequest
		want   string
	}{
		{
			name:   "missing raw body",
			bundle: plainBundle,
			req:    connectors.OperationDirectReadRequest{Operation: "github.markdown_raw", MaxBytes: 4},
			want:   "requires a raw body",
		},
		{
			name:   "mixed raw and JSON fields",
			bundle: plainBundle,
			req: func() connectors.OperationDirectReadRequest {
				raw := "text"
				return connectors.OperationDirectReadRequest{Operation: "github.markdown_raw", Body: map[string]any{"context": "octo/example"}, RawBody: &raw, MaxBytes: 4}
			}(),
			want: "cannot mix a raw body with JSON body fields",
		},
		{
			name:   "body exceeds declared cap",
			bundle: plainBundle,
			req: func() connectors.OperationDirectReadRequest {
				raw := "five!"
				return connectors.OperationDirectReadRequest{Operation: "github.markdown_raw", RawBody: &raw, MaxBytes: 4}
			}(),
			want: "request body too large",
		},
		{
			name: "static JSON body is not silently discarded",
			bundle: func() Bundle {
				b := plainBundle()
				b.Operations[0].REST.Body = map[string]any{"context": "octo/example"}
				return b
			},
			req: func() connectors.OperationDirectReadRequest {
				raw := "text"
				return connectors.OperationDirectReadRequest{Operation: "github.markdown_raw", RawBody: &raw, MaxBytes: 4}
			}(),
			want: "must not declare rest.body",
		},
		{
			name: "JSON operation cannot opt into raw input",
			bundle: func() Bundle {
				b := statusTextReadBundle(srv.URL, "github.markdown", http.MethodPost, "/markdown", "text", "application/json", 4)
				b.Operations[0].REST.BodySchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{}}`)
				return b
			},
			req: func() connectors.OperationDirectReadRequest {
				raw := "text"
				return connectors.OperationDirectReadRequest{Operation: "github.markdown", RawBody: &raw, MaxBytes: 4}
			}(),
			want: "requires declared text/plain",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := issued
			_, err := OperationDirectRead(context.Background(), tt.bundle(), tt.req, nil)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("OperationDirectRead error = %v, want %q", err, tt.want)
			}
			if issued != before {
				t.Fatalf("requests = %d, want no request for rejected body", issued)
			}
		})
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
		Name: "code-host",
		HTTP: HTTPBase{URL: baseURL},
		CLISurface: &CLISurface{Commands: []CLICommand{{
			Path: "fixture direct read", Intent: "direct_read", Availability: "implemented",
			APISurface: []CLISurfaceEndpointRef{{Method: method, Path: endpointPath}},
		}}},
	}
}

func TestOperationDirectReadBodySchemaMinItems(t *testing.T) {
	issued := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		issued = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	b := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID:           "acme.search",
			Kind:         "rest_read",
			Summary:      "Search",
			Risk:         "medium",
			Approval:     "none",
			OutputPolicy: "json_redacted",
			REST: &RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "/v1/search",
				ContentType: "application/json",
				MaxBytes:    1024,
				BodySchema: json.RawMessage(`{
					"type": "object", "additionalProperties": false,
					"required": ["ids"],
					"properties": {"ids": {"type": "array", "minItems": 1, "maxItems": 10, "items": {"type": "string"}}}
				}`),
			},
		}},
	}

	_, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "acme.search",
		Body:      map[string]any{"ids": []any{}},
	}, nil)
	if err == nil {
		t.Fatal("empty documented array: want body_schema error, got nil")
	}
	if !strings.Contains(err.Error(), "minItems") {
		t.Fatalf("error should name minItems, got %v", err)
	}
	if issued {
		t.Fatal("request must not be issued for an invalid body")
	}

	if _, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "acme.search",
		Body:      map[string]any{"ids": []any{"a"}},
	}, nil); err != nil {
		t.Fatalf("non-empty array: unexpected error: %v", err)
	}
}

func TestOperationDirectReadHTTPErrorKeepsProviderQueryAndBodyPrivate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Query().Get("trace"), "operation-read-fixture"; got != want {
			t.Fatalf("trace = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"diagnostic":"operation-read-fixture-body"}`))
	}))
	t.Cleanup(srv.Close)
	b := Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID: "acme.lookup", Kind: "rest_read", Summary: "lookup", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
			REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/items", MaxBytes: 1024, Parameters: []OperationParameter{{Name: "trace", In: "query", Type: "string"}}},
		}},
	}

	_, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "acme.lookup",
		Query:     map[string]string{"trace": "operation-read-fixture"},
	}, nil)
	if err == nil {
		t.Fatal("OperationDirectRead error = nil, want HTTP failure")
	}
	for _, secret := range []string{"trace=operation-read-fixture", "operation-read-fixture-body"} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("OperationDirectRead error leaked %q: %q", secret, err.Error())
		}
	}
	if !strings.Contains(err.Error(), "http 400") || !strings.Contains(err.Error(), "/items") {
		t.Fatalf("OperationDirectRead error = %q, want safe status and declaration identity", err.Error())
	}
}

func TestDirectReadCompleteReceiptOnSuccessAndProviderError(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusBadRequest} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Add("X-Provider-Receipt", "first")
				w.Header().Add("X-Provider-Receipt", "second")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"id":"occurrence-9007199254740993","token":"ordinary-token"}`))
			}))
			t.Cleanup(srv.Close)
			b := Bundle{
				Name: "acme", HTTP: HTTPBase{URL: srv.URL},
				Operations: []OperationSpec{{ID: "acme.lookup", Kind: "rest_read", Summary: "lookup", Risk: "low", Approval: "none", OutputPolicy: "json_redacted", REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/items", MaxBytes: 1024}}},
			}
			result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{Operation: "acme.lookup"}, nil)
			if status >= 400 && err == nil {
				t.Fatal("provider error = nil")
			}
			if status < 400 && err != nil {
				t.Fatalf("success: %v", err)
			}
			if result.Operation != "acme.lookup" || result.Receipt == nil || !result.Receipt.ResponseReceived || result.Receipt.Status != status || !result.Receipt.BodyPresent {
				t.Fatalf("complete receipt = %#v", result)
			}
			if got := result.Receipt.Headers["X-Provider-Receipt"].Values; !reflect.DeepEqual(got, []string{"first", "second"}) {
				t.Fatalf("duplicate receipt headers = %#v", got)
			}
			if !strings.Contains(result.Receipt.BodyRaw, "occurrence-9007199254740993") {
				t.Fatalf("receipt lost provider occurrence ID: %#v", result.Receipt)
			}
			if status >= 400 {
				var providerErr *connsdk.ProviderResponseError
				if !errors.As(err, &providerErr) || providerErr.Status != status {
					t.Fatalf("typed provider cause unavailable: %v", err)
				}
			}
		})
	}
}

func TestDirectReadReceiptPreservesAbsentAndInvalidBodiesWithConfiguredEqualValues(t *testing.T) {
	t.Run("absent body", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		t.Cleanup(srv.Close)
		b := Bundle{
			Name: "acme", HTTP: HTTPBase{URL: srv.URL},
			Operations: []OperationSpec{{ID: "acme.lookup", Kind: "rest_read", Summary: "lookup", Risk: "low", Approval: "none", OutputPolicy: "json_redacted", REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/items", MaxBytes: 1024}}},
		}
		result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{Operation: "acme.lookup"}, nil)
		if err == nil {
			t.Fatal("empty JSON response error = nil")
		}
		if result.Receipt == nil || result.Receipt.BodyPresent || result.Receipt.BodyBytes != 0 || result.Receipt.BodyRaw != "" {
			t.Fatalf("absent body receipt = %#v", result.Receipt)
		}
	})

	t.Run("invalid UTF-8", func(t *testing.T) {
		secret := "configured-secret-123"
		raw := append([]byte{0xff}, []byte(secret)...)
		raw = append(raw, 0xfe)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = w.Write(raw)
		}))
		t.Cleanup(srv.Close)
		b := Bundle{
			Name: "acme", HTTP: HTTPBase{URL: srv.URL},
			Operations: []OperationSpec{{ID: "acme.lookup", Kind: "rest_read", Summary: "lookup", Risk: "low", Approval: "none", OutputPolicy: "json_redacted", REST: &RESTOperationSpec{Method: http.MethodGet, Path: "/items", MaxBytes: 1024}}},
		}
		result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
			Operation: "acme.lookup",
			Config:    connectors.RuntimeConfig{Secrets: map[string]string{"api_token": secret}},
		}, nil)
		if err == nil {
			t.Fatal("invalid UTF-8 response error = nil")
		}
		if result.Receipt == nil || result.Receipt.BodyRawEncoding != "base64" || result.Receipt.BodyBytes != int64(len(raw)) {
			t.Fatalf("invalid UTF-8 receipt = %#v", result.Receipt)
		}
		// The engine receipt is immutable. Public serialization preserves a
		// provider-returned byte sequence even when it happens to contain a
		// configured credential value, retaining its binary encoding.
		public := connectors.SanitizeProviderResponseReceiptForOutput(*result.Receipt, map[string]string{"api_token": secret})
		decoded, decodeErr := base64.StdEncoding.DecodeString(public.BodyRaw)
		if decodeErr != nil || !bytes.Equal(decoded, raw) {
			t.Fatalf("preserved invalid UTF-8 receipt = %q, decode err %v", decoded, decodeErr)
		}
	})
}

func TestOperationDirectReadPreservesConfiguredEqualCursorAtThePublicBoundary(t *testing.T) {
	const configuredValue = "fixture-cursor-material"
	issued := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		issued++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"occurrence_id":"occurrence-9007199254740993","token_type":"ordinary"}],"next_cursor":"fixture-cursor-material"}`))
	}))
	t.Cleanup(srv.Close)
	b := paginatedOperationBundle(srv.URL, &PaginationSpec{Type: "cursor", CursorParam: "cursor", TokenPath: "next_cursor"}, "/v1/results")
	result, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "acme.list",
		Config:    connectors.RuntimeConfig{Secrets: map[string]string{"credential": configuredValue}},
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if issued != 1 {
		t.Fatalf("requests = %d, want 1", issued)
	}
	public, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(public, []byte(configuredValue)) {
		t.Fatal("public direct-read result did not preserve provider cursor material")
	}
	body := result.Body.(map[string]any)
	row := body["results"].([]any)[0].(map[string]any)
	if row["occurrence_id"] != "occurrence-9007199254740993" || row["token_type"] != "ordinary" {
		t.Fatal("cursor masking changed ordinary provider output")
	}
}

// --- required_query any-of groups ------------------------------------------
//
// Airtable's ledger blocks GET /v0/meta/enterpriseAccounts/{id}/users until the
// engine can "require at least one documented email[] or id[] query value
// without claiming an unfiltered executable stream". Nothing in the rule below
// mentions email, id, or Airtable.

func requiredQueryBundle(srv *httptest.Server, issued *bool) Bundle {
	return Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: srv.URL},
		Operations: []OperationSpec{{
			ID:           "acme.list_users",
			Kind:         "rest_read",
			Summary:      "List users by filter",
			Risk:         "medium",
			Approval:     "none",
			OutputPolicy: "json_redacted",
			REST: &RESTOperationSpec{
				Method:   http.MethodGet,
				Path:     "/v1/users",
				MaxBytes: 1024,
				Parameters: []OperationParameter{
					{Name: "email", In: "query", Type: "string"},
					{Name: "id", In: "query", Type: "string"},
					{Name: "since", In: "query", Type: "string"},
				},
				RequiredQuery: []RequiredQueryGroup{{AnyOf: []string{"email", "id"}}},
			},
		}},
	}
}

func TestOperationDirectReadRequiredQueryAnyOf(t *testing.T) {
	issued := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		issued = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"users":[]}`))
	}))
	defer srv.Close()
	b := requiredQueryBundle(srv, &issued)

	_, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "acme.list_users",
	}, nil)
	if err == nil {
		t.Fatal("unfiltered request: want error, got nil")
	}
	if !strings.Contains(err.Error(), "email") || !strings.Contains(err.Error(), "id") {
		t.Fatalf("error should name the group's parameters, got %v", err)
	}
	if issued {
		t.Fatal("the unfiltered request must never reach the provider")
	}

	if _, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "acme.list_users",
		Query:     map[string]string{"id": "usr123"},
	}, nil); err != nil {
		t.Fatalf("one member supplied: unexpected error: %v", err)
	}
	if !issued {
		t.Fatal("a satisfied request must be issued")
	}
}

func TestOperationDirectReadRequiredQueryRejectsBlankValue(t *testing.T) {
	issued := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		issued = true
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	// A present-but-empty parameter is exactly the unfiltered request the
	// constraint exists to prevent, so presence alone must not satisfy it.
	_, err := OperationDirectRead(context.Background(), requiredQueryBundle(srv, &issued), connectors.OperationDirectReadRequest{
		Operation: "acme.list_users",
		Query:     map[string]string{"email": "   "},
	}, nil)
	if err == nil {
		t.Fatal("blank value: want error, got nil")
	}
	if issued {
		t.Fatal("blank value must not reach the provider")
	}
}

func TestOperationDirectReadRequiredQuerySatisfiedByDeclaredQuery(t *testing.T) {
	// The constraint is about the request that goes on the wire, not about who
	// supplied the value: a value hardcoded in the operation's own rest.query
	// satisfies the group.
	issued := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issued = true
		if got := r.URL.Query().Get("email"); got != "ops@example.com" {
			t.Fatalf("email = %q, want the declared value", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	b := requiredQueryBundle(srv, &issued)
	b.Operations[0].REST.Query = map[string]string{"email": "ops@example.com"}

	if _, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "acme.list_users",
	}, nil); err != nil {
		t.Fatalf("declared query value: unexpected error: %v", err)
	}
	if !issued {
		t.Fatal("request should have been issued")
	}
}

func TestOperationDirectReadRequiredQueryEveryGroupMustBeSatisfied(t *testing.T) {
	issued := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		issued = true
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	b := requiredQueryBundle(srv, &issued)
	b.Operations[0].REST.RequiredQuery = []RequiredQueryGroup{
		{AnyOf: []string{"email", "id"}},
		{AnyOf: []string{"since"}},
	}

	_, err := OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{
		Operation: "acme.list_users",
		Query:     map[string]string{"id": "usr123"},
	}, nil)
	if err == nil {
		t.Fatal("second group unsatisfied: want error, got nil")
	}
	if !strings.Contains(err.Error(), "since") {
		t.Fatalf("error should name the unsatisfied group, got %v", err)
	}
	if issued {
		t.Fatal("request must not be issued while a group is unsatisfied")
	}
}
