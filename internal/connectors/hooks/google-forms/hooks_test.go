package googleforms

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

// --- test helpers -----------------------------------------------------

// tokenServer builds an httptest.NewTLSServer standing in for Google's OAuth
// token endpoint (the hook requires token_url to be https). respond is
// invoked per request; it receives the decoded form body and returns
// (status, body-or-nil). The returned *http.Client trusts the server's
// self-signed test certificate and must be set on Hooks.Client.
func tokenServer(t *testing.T, respond func(form url.Values) (int, map[string]any)) (*httptest.Server, *http.Client, *int32) {
	t.Helper()
	var hits int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if err := r.ParseForm(); err != nil {
			t.Fatalf("token server: parse form: %v", err)
		}
		status, body := respond(r.PostForm)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body == nil {
			body = map[string]any{"error": "server_error"}
		}
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	return srv, srv.Client(), &hits
}

func baseCfg(tokenURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{"token_url": tokenURL, "form_id": "abc123"},
		Secrets: map[string]string{
			"client_id":            "client-id-fixture",
			"client_secret":        "client-secret-fixture",
			"client_refresh_token": "refresh-token-fixture",
		},
	}
}

func baseSpec() engine.AuthSpec {
	return engine.AuthSpec{
		Mode:         "custom",
		Hook:         "google-forms",
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ secrets.client_id }}",
		ClientSecret: "{{ secrets.client_secret }}",
		Token:        "{{ secrets.client_refresh_token }}",
	}
}

func newTestHooks(now func() time.Time, client *http.Client) *Hooks {
	h := New().(*Hooks)
	h.Now = now
	h.Client = client
	return h
}

func newClientHooks(client *http.Client) *Hooks {
	h := New().(*Hooks)
	h.Client = client
	return h
}

func doAuthenticatedRequest(t *testing.T, auth interface {
	Apply(ctx context.Context, req *http.Request) error
}) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return req
}

// --- registration -------------------------------------------------------

func TestHooksRegisteredUnderGoogleForms(t *testing.T) {
	h := engine.HooksFor("google-forms")
	if h == nil {
		t.Fatal(`engine.HooksFor("google-forms") = nil, want registered hooks (init() must call engine.RegisterHooks)`)
	}
	if h.ConnectorName() != "google-forms" {
		t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), "google-forms")
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("registered google-forms hooks does not implement engine.AuthHook")
	}
	if _, ok := h.(engine.CheckHook); !ok {
		t.Fatal("registered google-forms hooks does not implement engine.CheckHook")
	}
}

// --- refresh-grant form shape -------------------------------------------

func TestAuthenticator_RefreshGrantFormShape(t *testing.T) {
	var gotForm url.Values
	srv, client, hits := tokenServer(t, func(form url.Values) (int, map[string]any) {
		gotForm = form
		return http.StatusOK, map[string]any{"access_token": "tok_abc", "token_type": "Bearer", "expires_in": 3600}
	})

	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)

	if *hits != 1 {
		t.Fatalf("token endpoint hits = %d, want 1", *hits)
	}
	if got := gotForm.Get("grant_type"); got != "refresh_token" {
		t.Fatalf("grant_type = %q, want %q", got, "refresh_token")
	}
	if got := gotForm.Get("refresh_token"); got != "refresh-token-fixture" {
		t.Fatalf("refresh_token = %q, want %q", got, "refresh-token-fixture")
	}
	if got := gotForm.Get("client_id"); got != "client-id-fixture" {
		t.Fatalf("client_id = %q, want %q", got, "client-id-fixture")
	}
	if got := gotForm.Get("client_secret"); got != "client-secret-fixture" {
		t.Fatalf("client_secret = %q, want %q", got, "client-secret-fixture")
	}
	if got := req.Header.Get("Authorization"); got != "Bearer tok_abc" {
		t.Fatalf("Authorization header = %q, want %q", got, "Bearer tok_abc")
	}
}

// TestAuthenticator_ClientSecretOmittedWhenUnset mirrors legacy's
// `if a.clientSecret != ""` guard: a caller that never sets the
// client_secret secret must not send an empty client_secret= form field at
// all.
func TestAuthenticator_ClientSecretOmittedWhenUnset(t *testing.T) {
	var gotForm url.Values
	srv, client, _ := tokenServer(t, func(form url.Values) (int, map[string]any) {
		gotForm = form
		return http.StatusOK, map[string]any{"access_token": "tok_abc", "expires_in": 3600}
	})

	cfg := baseCfg(srv.URL)
	delete(cfg.Secrets, "client_secret")

	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	_ = doAuthenticatedRequest(t, auth)

	if _, ok := gotForm["client_secret"]; ok {
		t.Fatalf("client_secret form field present = %v, want absent when unset", gotForm["client_secret"])
	}
}

// TestAuthenticator_CachesTokenUntil60sBeforeExpiry verifies the token cache
// mirrors legacy's 60s-before-expiry refresh guard.
func TestAuthenticator_CachesTokenUntil60sBeforeExpiry(t *testing.T) {
	var hits int32
	srv, client, _ := tokenServer(t, func(form url.Values) (int, map[string]any) {
		n := atomic.AddInt32(&hits, 1)
		return http.StatusOK, map[string]any{"access_token": "tok_" + strings.Repeat("x", int(n)), "expires_in": 3600}
	})

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	h := newTestHooks(clock, client)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	req1 := doAuthenticatedRequest(t, auth)
	req2 := doAuthenticatedRequest(t, auth)
	if hits != 1 {
		t.Fatalf("token endpoint hits = %d, want 1 (second call should hit cache)", hits)
	}
	if req1.Header.Get("Authorization") != req2.Header.Get("Authorization") {
		t.Fatalf("cached token changed between calls: %q vs %q", req1.Header.Get("Authorization"), req2.Header.Get("Authorization"))
	}

	// Advance past expiry minus the 60s guard window.
	now = now.Add(3600*time.Second - 30*time.Second)
	_ = doAuthenticatedRequest(t, auth)
	if hits != 2 {
		t.Fatalf("token endpoint hits = %d, want 2 (should refresh within 60s of expiry)", hits)
	}
}

// TestAuthenticator_MissingRefreshTokenErrors verifies a required-field
// guard.
func TestAuthenticator_MissingRefreshTokenErrors(t *testing.T) {
	cfg := baseCfg("https://oauth2.googleapis.com/token")
	delete(cfg.Secrets, "client_refresh_token")

	h := newClientHooks(nil)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("Authenticator: want error for missing refresh token, got nil")
	}
}

// TestAuthenticator_TokenURLMustBeHTTPS verifies the SSRF-hardening guard:
// a non-https token_url override fails closed.
func TestAuthenticator_TokenURLMustBeHTTPS(t *testing.T) {
	cfg := baseCfg("http://insecure.example.com/token")
	h := newClientHooks(nil)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil || !strings.Contains(err.Error(), "https") {
		t.Fatalf("Authenticator: want https-required error, got %v", err)
	}
}

// TestAuthenticator_TokenEndpointErrorPropagates verifies a non-2xx token
// response surfaces as an error, never a silently-empty token.
func TestAuthenticator_TokenEndpointErrorPropagates(t *testing.T) {
	srv, client, _ := tokenServer(t, func(form url.Values) (int, map[string]any) {
		return http.StatusUnauthorized, map[string]any{"error": "invalid_grant"}
	})
	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err := auth.Apply(context.Background(), req); err == nil {
		t.Fatal("Apply: want error for a non-2xx token response, got nil")
	}
}

// --- CheckHook ------------------------------------------------------------

func TestCheck_ReadsFirstConfiguredFormID(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"formId":"abc123"}`))
	}))
	defer srv.Close()

	rt := &engine.Runtime{Requester: &connsdk.Requester{BaseURL: srv.URL}}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"form_id": "abc123, def456"}}

	h := &Hooks{}
	handled, err := h.Check(context.Background(), cfg, rt)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if sawPath != "/forms/abc123" {
		t.Fatalf("path = %q, want /forms/abc123 (first configured form_id)", sawPath)
	}
}

func TestCheck_MissingFormIDErrors(t *testing.T) {
	rt := &engine.Runtime{Requester: &connsdk.Requester{BaseURL: "http://example.invalid"}}
	cfg := connectors.RuntimeConfig{Config: map[string]string{}}

	h := &Hooks{}
	handled, err := h.Check(context.Background(), cfg, rt)
	if err == nil {
		t.Fatal("Check: want error for missing form_id, got nil")
	}
	if !handled {
		t.Fatal("handled = false, want true (Check always handles, even on error)")
	}
}

func TestCheck_UpstreamErrorPropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	rt := &engine.Runtime{Requester: &connsdk.Requester{BaseURL: srv.URL}}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"form_id": "missing"}}

	h := &Hooks{}
	_, err := h.Check(context.Background(), cfg, rt)
	if err == nil {
		t.Fatal("Check: want error for upstream 404, got nil")
	}
}

// --- splitFormIDs ----------------------------------------------------------

func TestSplitFormIDs(t *testing.T) {
	cases := []struct {
		raw  string
		want []string
	}{
		{"abc123", []string{"abc123"}},
		{"abc123,def456", []string{"abc123", "def456"}},
		{"abc123, def456 ,ghi789", []string{"abc123", "def456", "ghi789"}},
		{"abc123\ndef456", []string{"abc123", "def456"}},
		{"", nil},
		{"   ", nil},
	}
	for _, c := range cases {
		got := splitFormIDs(c.raw)
		if len(got) != len(c.want) {
			t.Errorf("splitFormIDs(%q) = %v, want %v", c.raw, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitFormIDs(%q) = %v, want %v", c.raw, got, c.want)
				break
			}
		}
	}
}
