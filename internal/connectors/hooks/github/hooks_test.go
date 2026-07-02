package github_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
	githubhooks "polymetrics.ai/internal/connectors/hooks/github"
)

// testPrivateKeyPEM returns a freshly generated (test-only) RSA private key
// PEM, matching the PKCS1 shape legacy auth.go's githubParsePrivateKey
// accepts (x509.ParsePKCS1PrivateKey).
func testPrivateKeyPEM(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	der := x509.MarshalPKCS1PrivateKey(key)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}
	return string(pem.EncodeToMemory(block))
}

func newRuntimeConfig(baseURL string, cfgExtra map[string]string, secrets map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"base_url": baseURL, "owner": "octocat", "repo": "hello-world"}
	for k, v := range cfgExtra {
		cfg[k] = v
	}
	return connectors.RuntimeConfig{Config: cfg, Secrets: secrets}
}


func TestAuthenticatorGithubApp_MintsInstallationTokenAndSetsBearer(t *testing.T) {
	pemKey := testPrivateKeyPEM(t)
	const wantToken = "ghs_installation_fixture_token"

	var gotPath, gotMethod, gotAuthPrefix string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotAuthPrefix = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"token": wantToken})
	}))
	defer srv.Close()

	h := githubhooks.New()
	cfg := newRuntimeConfig(srv.URL, map[string]string{
		"app_id":          "12345",
		"installation_id": "67890",
	}, map[string]string{"private_key": pemKey})

	spec := engine.AuthSpec{Mode: "custom", Hook: "github"}
	authenticator, err := h.Authenticator(context.Background(), cfg, spec)
	if err != nil {
		t.Fatalf("Authenticator() error = %v", err)
	}
	if authenticator == nil {
		t.Fatal("Authenticator() = nil, want a non-nil connsdk.Authenticator")
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("installation token request method = %q, want POST", gotMethod)
	}
	if gotPath != "/app/installations/67890/access_tokens" {
		t.Fatalf("installation token request path = %q, want /app/installations/67890/access_tokens", gotPath)
	}
	if !strings.HasPrefix(gotAuthPrefix, "Bearer ") {
		t.Fatalf("installation token request Authorization = %q, want a Bearer-prefixed JWT", gotAuthPrefix)
	}

	// Apply the returned Authenticator to an outbound request and assert it
	// sets Authorization: Bearer <installation token> (not the JWT).
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/octocat/hello-world", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := authenticator.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	want := "Bearer " + wantToken
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("Authorization header = %q, want %q", got, want)
	}
}

func TestAuthenticatorGithubApp_MissingAppIDErrors(t *testing.T) {
	h := githubhooks.New()
	cfg := newRuntimeConfig("https://example.invalid", map[string]string{"installation_id": "67890"}, map[string]string{"private_key": testPrivateKeyPEM(t)})
	_, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{Mode: "custom", Hook: "github"})
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing app_id")
	}
}

func TestAuthenticatorGithubApp_MissingInstallationIDErrors(t *testing.T) {
	h := githubhooks.New()
	cfg := newRuntimeConfig("https://example.invalid", map[string]string{"app_id": "12345"}, map[string]string{"private_key": testPrivateKeyPEM(t)})
	_, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{Mode: "custom", Hook: "github"})
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing installation_id")
	}
}

func TestAuthenticatorGithubApp_MissingPrivateKeyErrors(t *testing.T) {
	h := githubhooks.New()
	cfg := newRuntimeConfig("https://example.invalid", map[string]string{"app_id": "12345", "installation_id": "67890"}, nil)
	_, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{Mode: "custom", Hook: "github"})
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing private key")
	}
}

func TestAuthenticatorGithubApp_PrivateKeyBase64Variant(t *testing.T) {
	pemKey := testPrivateKeyPEM(t)
	encoded := base64.StdEncoding.EncodeToString([]byte(pemKey))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "ghs_from_base64_key"})
	}))
	defer srv.Close()

	h := githubhooks.New()
	cfg := newRuntimeConfig(srv.URL, map[string]string{
		"app_id":          "12345",
		"installation_id": "67890",
	}, map[string]string{"private_key_base64": encoded})

	authenticator, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{Mode: "custom", Hook: "github"})
	if err != nil {
		t.Fatalf("Authenticator() error = %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.github.com/", nil)
	if err := authenticator.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer ghs_from_base64_key" {
		t.Fatalf("Authorization header = %q, want %q", got, "Bearer ghs_from_base64_key")
	}
}

func TestAuthenticatorGithubApp_HonorsContextCancellation(t *testing.T) {
	pemKey := testPrivateKeyPEM(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "unused"})
	}))
	defer srv.Close()

	h := githubhooks.New()
	cfg := newRuntimeConfig(srv.URL, map[string]string{"app_id": "1", "installation_id": "2"}, map[string]string{"private_key": pemKey})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := h.Authenticator(ctx, cfg, engine.AuthSpec{Mode: "custom", Hook: "github"})
	if err == nil {
		t.Fatal("Authenticator() error = nil for an already-cancelled context, want an error (F8: ctx must be honored, not context.Background())")
	}
}

// --- ExecuteWrite (WriteHook) ---

// captureServer records every request it receives (method, path, decoded
// JSON body) in order, answering each with a fixed JSON response.
type recordedRequest struct {
	Method string
	Path   string
	Body   map[string]any
}

func newWriteCaptureServer(t *testing.T, response map[string]any) (*httptest.Server, *[]recordedRequest) {
	t.Helper()
	var reqs []recordedRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		reqs = append(reqs, recordedRequest{Method: r.Method, Path: r.URL.Path, Body: body})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if response != nil {
			_ = json.NewEncoder(w).Encode(response)
		} else {
			_, _ = w.Write([]byte("{}"))
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &reqs
}

func newTestRuntime(baseURL string, cfg connectors.RuntimeConfig) *engine.Runtime {
	return &engine.Runtime{
		Requester: &connsdk.Requester{BaseURL: baseURL},
		Config:    cfg,
	}
}

func TestExecuteWrite_CloseIssueWithComment(t *testing.T) {
	srv, reqs := newWriteCaptureServer(t, nil)
	h := githubhooks.New()
	cfg := newRuntimeConfig(srv.URL, nil, nil)
	rt := newTestRuntime(srv.URL, cfg)

	action := engine.WriteAction{Name: "close_issue", Method: "PATCH", Path: "/repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}"}
	rec := connectors.Record{"issue_number": 101, "comment": "Closing via fixture", "state_reason": "completed"}

	handled, err := h.ExecuteWrite(context.Background(), action, rec, rt)
	if err != nil {
		t.Fatalf("ExecuteWrite() error = %v", err)
	}
	if !handled {
		t.Fatal("ExecuteWrite() handled = false, want true for close_issue (compound)")
	}
	if len(*reqs) != 2 {
		t.Fatalf("requests = %d, want 2 (comment POST then state PATCH), got %+v", len(*reqs), *reqs)
	}
	comment := (*reqs)[0]
	if comment.Method != http.MethodPost || comment.Path != "/repos/octocat/hello-world/issues/101/comments" {
		t.Fatalf("comment request = %+v, want POST /repos/octocat/hello-world/issues/101/comments", comment)
	}
	if comment.Body["body"] != "Closing via fixture" {
		t.Fatalf("comment body = %+v, want body=Closing via fixture", comment.Body)
	}
	patch := (*reqs)[1]
	if patch.Method != http.MethodPatch || patch.Path != "/repos/octocat/hello-world/issues/101" {
		t.Fatalf("close request = %+v, want PATCH /repos/octocat/hello-world/issues/101", patch)
	}
	if patch.Body["state"] != "closed" || patch.Body["state_reason"] != "completed" {
		t.Fatalf("close body = %+v, want state=closed state_reason=completed", patch.Body)
	}
}

func TestExecuteWrite_CloseIssueWithoutComment(t *testing.T) {
	srv, reqs := newWriteCaptureServer(t, nil)
	h := githubhooks.New()
	cfg := newRuntimeConfig(srv.URL, nil, nil)
	rt := newTestRuntime(srv.URL, cfg)

	action := engine.WriteAction{Name: "close_issue"}
	rec := connectors.Record{"issue_number": 101}

	handled, err := h.ExecuteWrite(context.Background(), action, rec, rt)
	if err != nil {
		t.Fatalf("ExecuteWrite() error = %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if len(*reqs) != 1 {
		t.Fatalf("requests = %d, want 1 (no comment configured)", len(*reqs))
	}
	if (*reqs)[0].Method != http.MethodPatch {
		t.Fatalf("method = %q, want PATCH", (*reqs)[0].Method)
	}
}

func TestExecuteWrite_CreatePullRequestWithFollowups(t *testing.T) {
	srv, reqs := newWriteCaptureServer(t, map[string]any{"number": 301})
	h := githubhooks.New()
	cfg := newRuntimeConfig(srv.URL, nil, nil)
	rt := newTestRuntime(srv.URL, cfg)

	action := engine.WriteAction{Name: "create_pull_request"}
	rec := connectors.Record{
		"head": "feature-1", "base": "main", "title": "Fixture PR",
		"labels": []string{"bug"}, "reviewers": []string{"octocat"},
	}

	handled, err := h.ExecuteWrite(context.Background(), action, rec, rt)
	if err != nil {
		t.Fatalf("ExecuteWrite() error = %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if len(*reqs) != 3 {
		t.Fatalf("requests = %d, want 3 (create PR, issue-metadata PATCH, reviewers POST), got %+v", len(*reqs), *reqs)
	}
	create := (*reqs)[0]
	if create.Method != http.MethodPost || create.Path != "/repos/octocat/hello-world/pulls" {
		t.Fatalf("create request = %+v, want POST /repos/octocat/hello-world/pulls", create)
	}
	meta := (*reqs)[1]
	if meta.Method != http.MethodPatch || meta.Path != "/repos/octocat/hello-world/issues/301" {
		t.Fatalf("metadata request = %+v, want PATCH /repos/octocat/hello-world/issues/301", meta)
	}
	reviewers := (*reqs)[2]
	if reviewers.Method != http.MethodPost || reviewers.Path != "/repos/octocat/hello-world/pulls/301/requested_reviewers" {
		t.Fatalf("reviewers request = %+v, want POST /repos/octocat/hello-world/pulls/301/requested_reviewers", reviewers)
	}
}

func TestExecuteWrite_CreatePullRequestNoFollowups(t *testing.T) {
	srv, reqs := newWriteCaptureServer(t, map[string]any{"number": 301})
	h := githubhooks.New()
	cfg := newRuntimeConfig(srv.URL, nil, nil)
	rt := newTestRuntime(srv.URL, cfg)

	action := engine.WriteAction{Name: "create_pull_request"}
	rec := connectors.Record{"head": "feature-1", "base": "main", "title": "Fixture PR"}

	handled, err := h.ExecuteWrite(context.Background(), action, rec, rt)
	if err != nil {
		t.Fatalf("ExecuteWrite() error = %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if len(*reqs) != 1 {
		t.Fatalf("requests = %d, want 1 (no labels/assignees/milestone/reviewers configured)", len(*reqs))
	}
}

func TestExecuteWrite_UpdatePullRequestWithFollowups(t *testing.T) {
	srv, reqs := newWriteCaptureServer(t, nil)
	h := githubhooks.New()
	cfg := newRuntimeConfig(srv.URL, nil, nil)
	rt := newTestRuntime(srv.URL, cfg)

	action := engine.WriteAction{Name: "update_pull_request"}
	rec := connectors.Record{"pull_number": 301, "title": "Updated", "reviewers": []string{"octocat"}}

	handled, err := h.ExecuteWrite(context.Background(), action, rec, rt)
	if err != nil {
		t.Fatalf("ExecuteWrite() error = %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if len(*reqs) != 2 {
		t.Fatalf("requests = %d, want 2 (core PATCH + reviewers POST), got %+v", len(*reqs), *reqs)
	}
	if (*reqs)[0].Path != "/repos/octocat/hello-world/pulls/301" || (*reqs)[0].Method != http.MethodPatch {
		t.Fatalf("core request = %+v", (*reqs)[0])
	}
	if (*reqs)[1].Path != "/repos/octocat/hello-world/pulls/301/requested_reviewers" {
		t.Fatalf("reviewers request = %+v", (*reqs)[1])
	}
}

func TestExecuteWrite_ClosePullRequestWithComment(t *testing.T) {
	srv, reqs := newWriteCaptureServer(t, nil)
	h := githubhooks.New()
	cfg := newRuntimeConfig(srv.URL, nil, nil)
	rt := newTestRuntime(srv.URL, cfg)

	action := engine.WriteAction{Name: "close_pull_request"}
	rec := connectors.Record{"pull_number": 301, "comment": "Closing PR"}

	handled, err := h.ExecuteWrite(context.Background(), action, rec, rt)
	if err != nil {
		t.Fatalf("ExecuteWrite() error = %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if len(*reqs) != 2 {
		t.Fatalf("requests = %d, want 2 (comment POST then close PATCH)", len(*reqs))
	}
	if (*reqs)[0].Path != "/repos/octocat/hello-world/issues/301/comments" {
		t.Fatalf("comment request path = %q", (*reqs)[0].Path)
	}
	if (*reqs)[1].Path != "/repos/octocat/hello-world/pulls/301" || (*reqs)[1].Body["state"] != "closed" {
		t.Fatalf("close request = %+v", (*reqs)[1])
	}
}

func TestExecuteWrite_NonCompoundActionFallsBackToDeclarative(t *testing.T) {
	h := githubhooks.New()
	rt := &engine.Runtime{}
	action := engine.WriteAction{Name: "create_issue"}
	rec := connectors.Record{"title": "not compound"}

	handled, err := h.ExecuteWrite(context.Background(), action, rec, rt)
	if err != nil {
		t.Fatalf("ExecuteWrite() error = %v", err)
	}
	if handled {
		t.Fatal("handled = true for a non-compound action, want false (declarative fallback)")
	}
}

func TestConnectorName(t *testing.T) {
	h := githubhooks.New()
	if got := h.ConnectorName(); got != "github" {
		t.Fatalf("ConnectorName() = %q, want %q", got, "github")
	}
}
