package github_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	githubhooks "polymetrics.ai/internal/connectors/hooks/github"
	"polymetrics.ai/internal/coordination"
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

// githubAppAuthRecordingTransport proves whether a physical token exchange
// reached the provider boundary. It deliberately records no headers or body:
// both contain credential material for the GitHub App flow.
type githubAppAuthRecordingTransport struct {
	sends atomic.Int32
}

func (t *githubAppAuthRecordingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	t.sends.Add(1)
	return &http.Response{
		StatusCode: http.StatusCreated,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"token":"synthetic-installation-token"}`)),
	}, nil
}

func requireSharedGitHubAppBundle(t *testing.T) engine.Bundle {
	t.Helper()
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load(defs.FS, github): %v", err)
	}
	for i := range bundle.RateLimits.Policies {
		if bundle.RateLimits.Policies[i].ID == "app-installation" {
			bundle.RateLimits.Policies[i].Coordination = connsdk.RateLimitCoordinationRequireShared
			return bundle
		}
	}
	t.Fatal("GitHub app-installation rate policy is absent")
	return engine.Bundle{}
}

func githubAppAuthAdmissionConfig(t *testing.T) connectors.RuntimeConfig {
	t.Helper()
	identity, err := connectors.NewCoordinationIdentity([]byte("github-app-auth-admission-test-salt"), connectors.CredentialBinding{
		BindingID:      "github-app-auth-admission-test-binding",
		ProviderFamily: "github",
		AuthProfile:    "github_app",
	})
	if err != nil {
		t.Fatalf("NewCoordinationIdentity: %v", err)
	}
	cfg := newRuntimeConfig("https://github-app-auth-rate.test", map[string]string{
		"app_id":          "4072",
		"installation_id": "admission-test-installation",
		"auth_type":       "github_app",
	}, map[string]string{"private_key": testPrivateKeyPEM(t)})
	cfg.CoordinationIdentity = identity
	return cfg
}

// githubAppDeclaredRouteRecorder proves the hook asks the engine for a
// declaration-aware route without retaining JWT/header values in test state.
type githubAppDeclaredRouteRecorder struct {
	method       string
	declaredPath string
	path         string
	hasBearerJWT bool
	responseBody []byte
}

func (r *githubAppDeclaredRouteRecorder) DoJSON(_ context.Context, request engine.DeclaredRouteRequest) (*connsdk.Response, error) {
	r.method = request.Method
	r.declaredPath = request.DeclaredPath
	r.path = request.Path
	r.hasBearerJWT = strings.HasPrefix(request.Headers["Authorization"], "Bearer ")
	return &connsdk.Response{Status: http.StatusCreated, Body: r.responseBody}, nil
}

func TestGitHubAppAuthRateAdmissionRequireSharedRefusesBeforeTokenSend(t *testing.T) {
	recordingTransport := &githubAppAuthRecordingTransport{}
	previousDefaultTransport := http.DefaultTransport
	http.DefaultTransport = recordingTransport
	t.Cleanup(func() { http.DefaultTransport = previousDefaultTransport })

	engine.ConfigureSharedRateLimitRegistry(coordination.NewSharedRateLimitRegistry(nil))
	t.Cleanup(func() { engine.ConfigureSharedRateLimitRegistry(nil) })

	_, err := engine.NewRuntime(context.Background(), requireSharedGitHubAppBundle(t), githubAppAuthAdmissionConfig(t), githubhooks.New())
	if got := recordingTransport.sends.Load(); got != 0 {
		t.Fatalf("physical GitHub App token sends = %d, want 0 before shared admission refusal (NewRuntime error = %v)", got, err)
	}
	var unavailable *coordination.SharedRateLimitUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("NewRuntime error = %T %v, want typed shared coordinator refusal", err, err)
	}
}

func TestAuthenticatorGithubApp_MintsInstallationTokenAndSetsBearer(t *testing.T) {
	pemKey := testPrivateKeyPEM(t)
	const wantToken = "ghs_installation_fixture_token"

	h := githubhooks.New()
	cfg := newRuntimeConfig("https://example.invalid", map[string]string{
		"app_id":          "12345",
		"installation_id": "67890",
	}, map[string]string{"private_key": pemKey})
	route := &githubAppDeclaredRouteRecorder{responseBody: []byte(`{"token":"` + wantToken + `"}`)}

	spec := engine.AuthSpec{Mode: "custom", Hook: "github"}
	authenticator, err := h.AuthenticatorWithDeclaredRoute(context.Background(), cfg, spec, route)
	if err != nil {
		t.Fatalf("AuthenticatorWithDeclaredRoute() error = %v", err)
	}
	if authenticator == nil {
		t.Fatal("AuthenticatorWithDeclaredRoute() = nil, want a non-nil connsdk.Authenticator")
	}

	if got, want := route.method, http.MethodPost; got != want {
		t.Fatalf("installation token request method = %q, want %q", got, want)
	}
	if got, want := route.declaredPath, "/app/installations/{installation_id}/access_tokens"; got != want {
		t.Fatalf("installation token declaration path = %q, want %q", got, want)
	}
	if got, want := route.path, "/app/installations/67890/access_tokens"; got != want {
		t.Fatalf("installation token request path = %q, want %q", got, want)
	}
	if !route.hasBearerJWT {
		t.Fatal("installation token request did not provide a Bearer JWT to the engine-owned route capability")
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
	_, err := h.AuthenticatorWithDeclaredRoute(context.Background(), cfg, engine.AuthSpec{Mode: "custom", Hook: "github"}, &githubAppDeclaredRouteRecorder{})
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing app_id")
	}
}

func TestAuthenticatorGithubApp_MissingInstallationIDErrors(t *testing.T) {
	h := githubhooks.New()
	cfg := newRuntimeConfig("https://example.invalid", map[string]string{"app_id": "12345"}, map[string]string{"private_key": testPrivateKeyPEM(t)})
	_, err := h.AuthenticatorWithDeclaredRoute(context.Background(), cfg, engine.AuthSpec{Mode: "custom", Hook: "github"}, &githubAppDeclaredRouteRecorder{})
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing installation_id")
	}
}

func TestAuthenticatorGithubApp_MissingPrivateKeyErrors(t *testing.T) {
	h := githubhooks.New()
	cfg := newRuntimeConfig("https://example.invalid", map[string]string{"app_id": "12345", "installation_id": "67890"}, nil)
	_, err := h.AuthenticatorWithDeclaredRoute(context.Background(), cfg, engine.AuthSpec{Mode: "custom", Hook: "github"}, &githubAppDeclaredRouteRecorder{})
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing private key")
	}
}

func TestAuthenticatorGithubApp_PrivateKeyBase64Variant(t *testing.T) {
	pemKey := testPrivateKeyPEM(t)
	encoded := base64.StdEncoding.EncodeToString([]byte(pemKey))

	h := githubhooks.New()
	cfg := newRuntimeConfig("https://example.invalid", map[string]string{
		"app_id":          "12345",
		"installation_id": "67890",
	}, map[string]string{"private_key_base64": encoded})
	route := &githubAppDeclaredRouteRecorder{responseBody: []byte(`{"token":"ghs_from_base64_key"}`)}

	authenticator, err := h.AuthenticatorWithDeclaredRoute(context.Background(), cfg, engine.AuthSpec{Mode: "custom", Hook: "github"}, route)
	if err != nil {
		t.Fatalf("AuthenticatorWithDeclaredRoute() error = %v", err)
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
	h := githubhooks.New()
	cfg := newRuntimeConfig("https://example.invalid", map[string]string{"app_id": "1", "installation_id": "2"}, map[string]string{"private_key": pemKey})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := h.AuthenticatorWithDeclaredRoute(ctx, cfg, engine.AuthSpec{Mode: "custom", Hook: "github"}, &githubAppDeclaredRouteRecorder{})
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

func TestExecuteWrite_ReopenIssue(t *testing.T) {
	srv, reqs := newWriteCaptureServer(t, nil)
	h := githubhooks.New()
	cfg := newRuntimeConfig(srv.URL, nil, nil)
	rt := newTestRuntime(srv.URL, cfg)

	action := engine.WriteAction{Name: "reopen_issue", Method: "PATCH", Path: "/repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}"}
	rec := connectors.Record{"issue_number": 101}

	handled, err := h.ExecuteWrite(context.Background(), action, rec, rt)
	if err != nil {
		t.Fatalf("ExecuteWrite() error = %v", err)
	}
	if !handled {
		t.Fatal("ExecuteWrite() handled = false, want true for reopen_issue (compound)")
	}
	if len(*reqs) != 1 {
		t.Fatalf("requests = %d, want 1 (state PATCH only), got %+v", len(*reqs), *reqs)
	}
	patch := (*reqs)[0]
	if patch.Method != http.MethodPatch || patch.Path != "/repos/octocat/hello-world/issues/101" {
		t.Fatalf("reopen request = %+v, want PATCH /repos/octocat/hello-world/issues/101", patch)
	}
	if patch.Body["state"] != "open" {
		t.Fatalf("reopen body = %+v, want state=open", patch.Body)
	}
	if _, ok := patch.Body["state_reason"]; ok {
		t.Fatalf("reopen body has state_reason, want none for reopen: %+v", patch.Body)
	}
}

func TestExecuteWrite_ReopenPullRequest(t *testing.T) {
	srv, reqs := newWriteCaptureServer(t, nil)
	h := githubhooks.New()
	cfg := newRuntimeConfig(srv.URL, nil, nil)
	rt := newTestRuntime(srv.URL, cfg)

	action := engine.WriteAction{Name: "reopen_pull_request", Method: "PATCH", Path: "/repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}"}
	rec := connectors.Record{"pull_number": 301}

	handled, err := h.ExecuteWrite(context.Background(), action, rec, rt)
	if err != nil {
		t.Fatalf("ExecuteWrite() error = %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true for reopen_pull_request")
	}
	if len(*reqs) != 1 {
		t.Fatalf("requests = %d, want 1, got %+v", len(*reqs), *reqs)
	}
	if (*reqs)[0].Method != http.MethodPatch || (*reqs)[0].Path != "/repos/octocat/hello-world/pulls/301" || (*reqs)[0].Body["state"] != "open" {
		t.Fatalf("reopen pr request = %+v, want PATCH pulls/301 state=open", (*reqs)[0])
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

func TestExecuteWrite_CreateLabelStripsLeadingHash(t *testing.T) {
	srv, reqs := newWriteCaptureServer(t, nil)
	h := githubhooks.New()
	cfg := newRuntimeConfig(srv.URL, nil, nil)
	rt := newTestRuntime(srv.URL, cfg)

	action := engine.WriteAction{Name: "create_label"}
	rec := connectors.Record{"name": "bug", "color": "#ff0000", "description": "Fixture label"}

	handled, err := h.ExecuteWrite(context.Background(), action, rec, rt)
	if err != nil {
		t.Fatalf("ExecuteWrite() error = %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true for create_label")
	}
	if len(*reqs) != 1 {
		t.Fatalf("requests = %d, want 1", len(*reqs))
	}
	got := (*reqs)[0]
	if got.Method != http.MethodPost || got.Path != "/repos/octocat/hello-world/labels" {
		t.Fatalf("request = %+v, want POST /repos/octocat/hello-world/labels", got)
	}
	if got.Body["color"] != "ff0000" {
		t.Fatalf("body color = %#v, want %q (leading # stripped)", got.Body["color"], "ff0000")
	}
	if got.Body["description"] != "Fixture label" {
		t.Fatalf("body description = %#v, want %q", got.Body["description"], "Fixture label")
	}
}

func TestExecuteWrite_CreateLabelMissingColorErrors(t *testing.T) {
	srv, _ := newWriteCaptureServer(t, nil)
	h := githubhooks.New()
	rt := newTestRuntime(srv.URL, newRuntimeConfig(srv.URL, nil, nil))

	action := engine.WriteAction{Name: "create_label"}
	rec := connectors.Record{"name": "bug"}

	handled, err := h.ExecuteWrite(context.Background(), action, rec, rt)
	if !handled {
		t.Fatal("handled = false, want true (create_label is always hook-handled)")
	}
	if err == nil {
		t.Fatal("ExecuteWrite() error = nil, want an error for a missing required color field")
	}
}

func TestExecuteWrite_UpdateLabelStripsLeadingHashWhenColorPresent(t *testing.T) {
	srv, reqs := newWriteCaptureServer(t, nil)
	h := githubhooks.New()
	rt := newTestRuntime(srv.URL, newRuntimeConfig(srv.URL, nil, nil))

	action := engine.WriteAction{Name: "update_label"}
	rec := connectors.Record{"name": "bug", "color": "#00ff00"}

	handled, err := h.ExecuteWrite(context.Background(), action, rec, rt)
	if err != nil {
		t.Fatalf("ExecuteWrite() error = %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true for update_label")
	}
	got := (*reqs)[0]
	if got.Method != http.MethodPatch || got.Path != "/repos/octocat/hello-world/labels/bug" {
		t.Fatalf("request = %+v, want PATCH /repos/octocat/hello-world/labels/bug", got)
	}
	if got.Body["color"] != "00ff00" {
		t.Fatalf("body color = %#v, want %q (leading # stripped)", got.Body["color"], "00ff00")
	}
	if _, ok := got.Body["new_name"]; ok {
		t.Fatalf("body = %+v, want no new_name key (field was absent on record)", got.Body)
	}
}

func TestExecuteWrite_UpdateLabelMissingNameErrors(t *testing.T) {
	srv, _ := newWriteCaptureServer(t, nil)
	h := githubhooks.New()
	rt := newTestRuntime(srv.URL, newRuntimeConfig(srv.URL, nil, nil))

	action := engine.WriteAction{Name: "update_label"}
	rec := connectors.Record{"color": "#00ff00"}

	handled, err := h.ExecuteWrite(context.Background(), action, rec, rt)
	if !handled {
		t.Fatal("handled = false, want true (update_label is always hook-handled)")
	}
	if err == nil {
		t.Fatal("ExecuteWrite() error = nil, want an error for a missing required name field")
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

// archive_repo and unarchive_repo share PATCH /repos/{owner}/{repo} with the
// generic repo update, so the only thing separating them is the body the hook
// pins. These two tests are what stop them collapsing back into "PATCH
// whatever the caller sent" — a `repo unarchive` that forwards an empty body
// silently does nothing while reporting success.
//
// They pin through MapWriteRecord rather than ExecuteWrite so both stay on the
// declarative path: a hook that overrides execution builds a request the
// preview never saw, so the digest an operator approves would not be the
// request that runs.
func TestMapWriteRecord_ArchiveRepoPinsArchivedTrue(t *testing.T) {
	h := githubhooks.New()
	action := engine.WriteAction{Name: "archive_repo", Method: "PATCH", Path: "/repos/{owner}/{repo}"}

	pinned, handled, err := h.MapWriteRecord(action, connectors.Record{})
	if err != nil {
		t.Fatalf("MapWriteRecord() error = %v", err)
	}
	if !handled {
		t.Fatal("MapWriteRecord() handled = false, want true for archive_repo")
	}
	if pinned["archived"] != true {
		t.Fatalf("archive record = %+v, want archived=true", pinned)
	}
	if handled, err := h.ExecuteWrite(context.Background(), action, connectors.Record{}, nil); handled || err != nil {
		t.Fatalf("ExecuteWrite() = (%v, %v), want (false, nil): the pinned-body action must stay declarative", handled, err)
	}
	if h.HandlesWriteAction(action) {
		t.Fatal("HandlesWriteAction(archive_repo) = true; the engine reads this to route the action away from the declarative path")
	}
}

func TestMapWriteRecord_UnarchiveRepoPinsArchivedFalse(t *testing.T) {
	h := githubhooks.New()
	action := engine.WriteAction{Name: "unarchive_repo", Method: "PATCH", Path: "/repos/{owner}/{repo}"}

	// A caller-supplied archived=true must not survive: the command name is the
	// instruction, not the record.
	pinned, handled, err := h.MapWriteRecord(action, connectors.Record{"archived": true})
	if err != nil {
		t.Fatalf("MapWriteRecord() error = %v", err)
	}
	if !handled {
		t.Fatal("MapWriteRecord() handled = false, want true for unarchive_repo")
	}
	if pinned["archived"] != false {
		t.Fatalf("unarchive record = %+v, want archived=false", pinned)
	}
	if handled, err := h.ExecuteWrite(context.Background(), action, connectors.Record{}, nil); handled || err != nil {
		t.Fatalf("ExecuteWrite() = (%v, %v), want (false, nil): the pinned-body action must stay declarative", handled, err)
	}
	if h.HandlesWriteAction(action) {
		t.Fatal("HandlesWriteAction(unarchive_repo) = true; the engine reads this to route the action away from the declarative path")
	}
}

// The pinned record must not leak back into the caller's record: bulk reverse
// ETL reuses the staged rows across preview and execution.
func TestMapWriteRecord_DoesNotMutateCallerRecord(t *testing.T) {
	h := githubhooks.New()
	action := engine.WriteAction{Name: "archive_repo", Method: "PATCH", Path: "/repos/{owner}/{repo}"}

	original := connectors.Record{}
	if _, _, err := h.MapWriteRecord(action, original); err != nil {
		t.Fatalf("MapWriteRecord() error = %v", err)
	}
	if _, ok := original["archived"]; ok {
		t.Fatalf("caller record = %+v, want untouched", original)
	}
}
