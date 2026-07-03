// Package blogger implements the blogger AuthHook: an OAuth 2.0
// refresh-token-grant connsdk.Authenticator, ported from legacy
// internal/connectors/blogger/blogger.go's refreshTokenAuth.
package blogger

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
	"polymetrics.ai/internal/connectors/engine"
)

// --- test helpers -----------------------------------------------------

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
		Config: map[string]string{"token_url": tokenURL, "blog_id": "fixture_blog_1"},
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
		Hook:         "blogger",
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

func TestHooksRegisteredUnderBlogger(t *testing.T) {
	h := engine.HooksFor("blogger")
	if h == nil {
		t.Fatal("engine.HooksFor(\"blogger\") = nil, want registered hooks (hooks/blogger's init() must call engine.RegisterHooks)")
	}
	if h.ConnectorName() != "blogger" {
		t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), "blogger")
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("registered blogger hooks does not implement engine.AuthHook")
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

// --- caching / expiry ----------------------------------------------------

func TestAuthenticator_CachesTokenAcrossRequests(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	srv, client, hits := tokenServer(t, func(form url.Values) (int, map[string]any) {
		return http.StatusOK, map[string]any{"access_token": "tok_1", "expires_in": 3600}
	})

	h := newTestHooks(func() time.Time { return now }, client)
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

// TestAuthenticator_RefreshesWithin60sOfExpiry: the cache must be treated as
// stale starting 60s before the declared expiry (legacy blogger.go:338-341).
func TestAuthenticator_RefreshesWithin60sOfExpiry(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	call := 0
	srv, client, hits := tokenServer(t, func(form url.Values) (int, map[string]any) {
		call++
		return http.StatusOK, map[string]any{"access_token": "tok_" + itoa(call), "expires_in": 3600}
	})

	current := now
	h := newTestHooks(func() time.Time { return current }, client)
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
	srv, client, _ := tokenServer(t, func(form url.Values) (int, map[string]any) {
		return http.StatusUnauthorized, map[string]any{"error": "invalid_grant"}
	})

	h := newClientHooks(client)
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

func TestAuthenticator_MissingRefreshTokenIsError(t *testing.T) {
	cfg := baseCfg("https://oauth2.googleapis.com/token")
	delete(cfg.Secrets, "client_refresh_token")

	h := New().(*Hooks)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing refresh token")
	}
	if !strings.Contains(err.Error(), "refresh") {
		t.Fatalf("error = %q, want it to name the missing refresh_token field", err.Error())
	}
}

func TestAuthenticator_MissingClientIDIsError(t *testing.T) {
	cfg := baseCfg("https://oauth2.googleapis.com/token")
	delete(cfg.Secrets, "client_id")

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
	cfg := baseCfg("https://oauth2.googleapis.com/token")
	delete(cfg.Secrets, "client_secret")

	h := New().(*Hooks)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing client_secret")
	}
	if !strings.Contains(err.Error(), "client_secret") {
		t.Fatalf("error = %q, want it to name client_secret", err.Error())
	}
}

// TestAuthenticator_TokenURLMustBeHTTPS is the fail-closed SSRF guard: a
// non-https token_url override must fail rather than send secrets to an
// arbitrary endpoint.
func TestAuthenticator_TokenURLMustBeHTTPS(t *testing.T) {
	cfg := baseCfg("http://insecure.example.invalid/token")

	h := newClientHooks(nil)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("Authenticator() error = nil, want a fail-closed error for a non-https token_url")
	}
	if !strings.Contains(err.Error(), "https") {
		t.Fatalf("error = %q, want it to mention the https requirement", err.Error())
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
	srv, client, _ := tokenServer(t, func(form url.Values) (int, map[string]any) {
		return http.StatusOK, map[string]any{"access_token": "tok_abc", "expires_in": 3600}
	})

	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err := auth.Apply(ctx, req); err == nil {
		t.Fatal("Apply(cancelled ctx) error = nil, want a cancellation error")
	}
}

// --- secret redaction ------------------------------------------------------

func TestAuthenticator_ErrorsNeverContainSecretText(t *testing.T) {
	const (
		secretMarkerClientSecret = "client-secret-fixture"
		secretMarkerRefreshToken = "refresh-token-fixture"
		secretMarkerAccessTok    = "tok_super_secret_access_value"
	)

	srv, client, _ := tokenServer(t, func(form url.Values) (int, map[string]any) {
		return http.StatusUnauthorized, map[string]any{"error": "invalid_grant", "access_token": secretMarkerAccessTok}
	})

	h := newClientHooks(client)
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
	for _, marker := range []string{secretMarkerClientSecret, secretMarkerRefreshToken} {
		if strings.Contains(msg, marker) {
			t.Fatalf("error text contains secret marker %q: %s", marker, msg)
		}
	}
}
