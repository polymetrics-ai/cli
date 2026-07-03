// Package akeneo implements the akeneo AuthHook: an OAuth2 password-grant
// connsdk.Authenticator, ported from legacy
// internal/connectors/akeneo/akeneo.go's passwordGrantAuth. This test file
// mirrors hooks/gmail/hooks_test.go's structure for the akeneo bundle's
// substitute-hook-test-in-place-of-dynamic-conformance-coverage marker
// (metadata.json's skip_dynamic reason names this test file explicitly).
package akeneo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

// --- test helpers -----------------------------------------------------

// tokenServer builds an httptest.NewServer (legacy's own akeneoBaseURL
// tolerates plain http, and legacy's own test suite exercises a plain-http
// httptest.Server for both the token and resource endpoints — unlike
// gmail's hook, this one does not require TLS) standing in for Akeneo's
// OAuth token endpoint. respond is invoked per request; it receives the
// decoded JSON body and the Authorization header, returning (status,
// body-or-nil). A nil body writes {} on a non-2xx status.
func tokenServer(t *testing.T, respond func(basicAuth string, body map[string]string) (int, map[string]any)) (*httptest.Server, *int32) {
	t.Helper()
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("token server: decode body: %v", err)
		}
		status, respBody := respond(r.Header.Get("Authorization"), body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if respBody == nil {
			respBody = map[string]any{"error": "server_error"}
		}
		_ = json.NewEncoder(w).Encode(respBody)
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

func baseCfg(tokenURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":     tokenURL,
			"client_id":    "client-id-fixture",
			"api_username": "api-user-fixture",
		},
		Secrets: map[string]string{
			"secret":   "client-secret-fixture",
			"password": "user-password-fixture",
		},
	}
}

func baseSpec() engine.AuthSpec {
	return engine.AuthSpec{
		Mode:         "custom",
		Hook:         "akeneo",
		TokenURL:     "{{ config.base_url }}/api/oauth/v1/token",
		ClientID:     "{{ config.client_id }}",
		ClientSecret: "{{ secrets.secret }}",
		Username:     "{{ config.api_username }}",
		Password:     "{{ secrets.password }}",
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

func TestHooksRegisteredUnderAkeneo(t *testing.T) {
	h := engine.HooksFor("akeneo")
	if h == nil {
		t.Fatal("engine.HooksFor(\"akeneo\") = nil, want registered hooks (hooks/akeneo's init() must call engine.RegisterHooks)")
	}
	if h.ConnectorName() != "akeneo" {
		t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), "akeneo")
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("registered akeneo hooks does not implement engine.AuthHook")
	}
}

// --- password-grant request shape ---------------------------------------

func TestAuthenticator_PasswordGrantRequestShape(t *testing.T) {
	var gotBasic string
	var gotBody map[string]string
	srv, hits := tokenServer(t, func(basicAuth string, body map[string]string) (int, map[string]any) {
		gotBasic = basicAuth
		gotBody = body
		return http.StatusOK, map[string]any{"access_token": "tok_abc", "token_type": "bearer", "expires_in": 3600}
	})

	h := newClientHooks(nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)

	if *hits != 1 {
		t.Fatalf("token endpoint hits = %d, want 1", *hits)
	}
	wantBasic := "Basic " + base64.StdEncoding.EncodeToString([]byte("client-id-fixture:client-secret-fixture"))
	if gotBasic != wantBasic {
		t.Fatalf("token Authorization = %q, want %q", gotBasic, wantBasic)
	}
	if gotBody["grant_type"] != "password" {
		t.Fatalf("grant_type = %q, want %q", gotBody["grant_type"], "password")
	}
	if gotBody["username"] != "api-user-fixture" {
		t.Fatalf("username = %q, want %q", gotBody["username"], "api-user-fixture")
	}
	if gotBody["password"] != "user-password-fixture" {
		t.Fatalf("password = %q, want %q", gotBody["password"], "user-password-fixture")
	}
	if got := req.Header.Get("Authorization"); got != "Bearer tok_abc" {
		t.Fatalf("resource Authorization header = %q, want %q", got, "Bearer tok_abc")
	}
}

// --- caching / expiry ----------------------------------------------------

func TestAuthenticator_CachesTokenAcrossRequests(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	srv, hits := tokenServer(t, func(string, map[string]string) (int, map[string]any) {
		return http.StatusOK, map[string]any{"access_token": "tok_1", "expires_in": 3600}
	})

	h := newTestHooks(func() time.Time { return now }, nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	req1 := doAuthenticatedRequest(t, auth)
	req2 := doAuthenticatedRequest(t, auth)

	if *hits != 1 {
		t.Fatalf("token endpoint hits = %d, want 1 (second Apply should reuse the cached token)", *hits)
	}
	if req1.Header.Get("Authorization") != req2.Header.Get("Authorization") {
		t.Fatalf("Authorization headers differ across cached requests: %q vs %q", req1.Header.Get("Authorization"), req2.Header.Get("Authorization"))
	}
}

func TestAuthenticator_RefreshesWithin60sOfExpiry(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	call := 0
	srv, hits := tokenServer(t, func(string, map[string]string) (int, map[string]any) {
		call++
		return http.StatusOK, map[string]any{"access_token": "tok_" + itoa(call), "expires_in": 3600}
	})

	current := now
	h := newTestHooks(func() time.Time { return current }, nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	_ = doAuthenticatedRequest(t, auth) // primes the cache, expires at now+3600s

	current = now.Add(3539 * time.Second)
	_ = doAuthenticatedRequest(t, auth)
	if *hits != 1 {
		t.Fatalf("token endpoint hits = %d after t+3539s, want 1 (still within cache window)", *hits)
	}

	current = now.Add(3541 * time.Second)
	_ = doAuthenticatedRequest(t, auth)
	if *hits != 2 {
		t.Fatalf("token endpoint hits = %d after t+3541s, want 2 (60s-early refresh must trigger)", *hits)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// --- error paths ----------------------------------------------------------

func TestAuthenticator_NonSuccessTokenResponseIsError(t *testing.T) {
	srv, _ := tokenServer(t, func(string, map[string]string) (int, map[string]any) {
		return http.StatusUnauthorized, map[string]any{"error": "invalid_grant"}
	})

	h := newClientHooks(nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	err = auth.Apply(context.Background(), req)
	if err == nil {
		t.Fatal("Apply() error = nil, want an error for a non-2xx token endpoint response")
	}
	if req.Header.Get("Authorization") != "" {
		t.Fatalf("Authorization header set = %q after a failed token exchange, want empty (no silent unauthenticated fallback)", req.Header.Get("Authorization"))
	}
}

func TestAuthenticator_MissingPasswordIsError(t *testing.T) {
	cfg := baseCfg("http://example.invalid")
	delete(cfg.Secrets, "password")

	h := New().(*Hooks)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing password")
	}
	if !strings.Contains(err.Error(), "password") {
		t.Fatalf("error = %q, want it to name the missing password field", err.Error())
	}
}

func TestAuthenticator_MissingClientIDIsError(t *testing.T) {
	cfg := baseCfg("http://example.invalid")
	delete(cfg.Config, "client_id")

	h := New().(*Hooks)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing client_id")
	}
	if !strings.Contains(err.Error(), "client_id") {
		t.Fatalf("error = %q, want it to name client_id", err.Error())
	}
}

func TestAuthenticator_MissingClientSecretIsError(t *testing.T) {
	cfg := baseCfg("http://example.invalid")
	delete(cfg.Secrets, "secret")

	h := New().(*Hooks)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing client secret")
	}
}

func TestAuthenticator_MissingUsernameIsError(t *testing.T) {
	cfg := baseCfg("http://example.invalid")
	delete(cfg.Config, "api_username")

	h := New().(*Hooks)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing api_username")
	}
	if !strings.Contains(err.Error(), "api_username") {
		t.Fatalf("error = %q, want it to name api_username", err.Error())
	}
}

// TestAuthenticator_TokenURLAcceptsPlainHTTP asserts the hook does NOT
// narrow legacy's http-or-https tolerance (akeneoBaseURL, akeneo.go:399)
// the way gmail's hook does for the unrelated Google OAuth endpoint —
// Akeneo deployments are not guaranteed to sit behind a well-known
// https-only domain, and legacy's own test suite drives a plain-http
// httptest.Server.
func TestAuthenticator_TokenURLAcceptsPlainHTTP(t *testing.T) {
	srv, _ := tokenServer(t, func(string, map[string]string) (int, map[string]any) {
		return http.StatusOK, map[string]any{"access_token": "tok_abc", "expires_in": 3600}
	})
	if !strings.HasPrefix(srv.URL, "http://") {
		t.Fatalf("test server URL = %q, want a plain http:// URL (test setup bug)", srv.URL)
	}

	h := newClientHooks(nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator() error = %v, want nil for a plain http token_url", err)
	}
	req := doAuthenticatedRequest(t, auth)
	if req.Header.Get("Authorization") != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", req.Header.Get("Authorization"))
	}
}

func TestAuthenticator_TokenURLRejectsBadScheme(t *testing.T) {
	cfg := baseCfg("ftp://insecure.example.invalid")

	h := newClientHooks(nil)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("Authenticator() error = nil, want a fail-closed error for a non-http(s) token_url")
	}
	if !strings.Contains(err.Error(), "http") {
		t.Fatalf("error = %q, want it to mention the http/https requirement", err.Error())
	}
}

func TestAuthenticator_TokenURLUnparseableIsError(t *testing.T) {
	cfg := baseCfg("://not-a-url")

	h := New().(*Hooks)
	if _, err := h.Authenticator(context.Background(), cfg, baseSpec()); err == nil {
		t.Fatal("Authenticator() error = nil, want an error for an unparseable token_url")
	}
}

// --- ctx cancellation -----------------------------------------------------

func TestAuthenticator_HonorsContextCancellation(t *testing.T) {
	srv, _ := tokenServer(t, func(string, map[string]string) (int, map[string]any) {
		return http.StatusOK, map[string]any{"access_token": "tok_abc", "expires_in": 3600}
	})

	h := newClientHooks(nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err := auth.Apply(ctx, req); err == nil {
		t.Fatal("Apply(cancelled ctx) error = nil, want a cancellation error (ctx must be honored, not context.Background())")
	}
}

func TestAuthenticator_HonorsContextCancellationBeforeAnyRequest(t *testing.T) {
	h := newClientHooks(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := h.Authenticator(ctx, baseCfg("http://example.invalid"), baseSpec()); err == nil {
		t.Fatal("Authenticator(cancelled ctx) error = nil, want a cancellation error")
	}
}

// --- secret redaction ------------------------------------------------------

func TestAuthenticator_ErrorsNeverContainSecretText(t *testing.T) {
	const (
		secretMarkerClientSecret = "client-secret-fixture"
		secretMarkerPassword     = "user-password-fixture"
		secretMarkerAccessTok    = "tok_super_secret_access_value"
	)

	srv, _ := tokenServer(t, func(string, map[string]string) (int, map[string]any) {
		return http.StatusUnauthorized, map[string]any{"error": "invalid_grant", "access_token": secretMarkerAccessTok}
	})

	h := newClientHooks(nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	err = auth.Apply(context.Background(), req)
	if err == nil {
		t.Fatal("expected an error from the 401 token response")
	}
	msg := err.Error()
	for _, marker := range []string{secretMarkerClientSecret, secretMarkerPassword} {
		if strings.Contains(msg, marker) {
			t.Fatalf("error text contains secret marker %q: %s", marker, msg)
		}
	}
}
