package salesloft

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func baseCfg(tokenURL string, extraSecrets map[string]string) connectors.RuntimeConfig {
	secrets := map[string]string{
		"client_id":     "client-id-fixture",
		"client_secret": "client-secret-fixture",
		"refresh_token": "refresh-token-fixture",
	}
	for k, v := range extraSecrets {
		secrets[k] = v
	}
	return connectors.RuntimeConfig{
		Config:  map[string]string{"token_url": tokenURL},
		Secrets: secrets,
	}
}

func baseSpec() engine.AuthSpec {
	return engine.AuthSpec{
		Mode:         "custom",
		Hook:         "salesloft",
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ secrets.client_id }}",
		ClientSecret: "{{ secrets.client_secret }}",
		Token:        "{{ secrets.refresh_token }}",
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

// --- registration ---

func TestHooksRegisteredUnderSalesloft(t *testing.T) {
	h := engine.HooksFor("salesloft")
	if h == nil {
		t.Fatal(`engine.HooksFor("salesloft") = nil, want registered hooks (hooks/salesloft's init() must call engine.RegisterHooks)`)
	}
	if h.ConnectorName() != "salesloft" {
		t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), "salesloft")
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("registered salesloft hooks does not implement engine.AuthHook")
	}
}

// --- refresh-grant form shape (form-encoded, NOT HTTP Basic) ---

func TestAuthenticator_RefreshGrantFormShape(t *testing.T) {
	var gotForm url.Values
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_ = r.ParseForm()
		gotForm = r.Form
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok_abc","token_type":"bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	h := newClientHooks(nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL, nil), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)

	if hits != 1 {
		t.Fatalf("token endpoint hits = %d, want 1", hits)
	}
	if got := gotForm.Get("grant_type"); got != "refresh_token" {
		t.Fatalf("grant_type = %q, want refresh_token", got)
	}
	if got := gotForm.Get("client_id"); got != "client-id-fixture" {
		t.Fatalf("client_id = %q, want client-id-fixture", got)
	}
	if got := gotForm.Get("client_secret"); got != "client-secret-fixture" {
		t.Fatalf("client_secret = %q, want client-secret-fixture", got)
	}
	if got := gotForm.Get("refresh_token"); got != "refresh-token-fixture" {
		t.Fatalf("refresh_token = %q, want refresh-token-fixture", got)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer tok_abc" {
		t.Fatalf("data request Authorization = %q, want Bearer tok_abc", got)
	}
}

// TestAuthenticator_SeedTokenHonoredBeforeRefresh asserts a config-provided
// access_token is used exactly once, without hitting the token endpoint at
// all, matching legacy's seedToken behavior (auth.go:63-69).
func TestAuthenticator_SeedTokenHonoredBeforeRefresh(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok_refreshed","token_type":"bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	h := newClientHooks(nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL, map[string]string{"access_token": "tok_seed"}), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)

	if hits != 0 {
		t.Fatalf("token endpoint hits = %d, want 0 (seed token honored without a refresh call)", hits)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer tok_seed" {
		t.Fatalf("Authorization = %q, want Bearer tok_seed", got)
	}
}

// TestAuthenticator_RefreshTokenRotates asserts a rotated refresh_token in
// the token response is used on a SUBSEQUENT exchange (legacy auth.go:113-116).
func TestAuthenticator_RefreshTokenRotates(t *testing.T) {
	var seenRefreshTokens []string
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		seenRefreshTokens = append(seenRefreshTokens, r.Form.Get("refresh_token"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok_x","refresh_token":"rotated-refresh-token","token_type":"bearer","expires_in":60}`))
	}))
	defer srv.Close()

	h := newTestHooks(func() time.Time { return now }, nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL, nil), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	doAuthenticatedRequest(t, auth)
	// Advance past expiry to force a second exchange.
	now = now.Add(120 * time.Second)
	doAuthenticatedRequest(t, auth)

	if len(seenRefreshTokens) != 2 {
		t.Fatalf("token endpoint hits = %d, want 2", len(seenRefreshTokens))
	}
	if seenRefreshTokens[0] != "refresh-token-fixture" {
		t.Fatalf("first exchange refresh_token = %q, want refresh-token-fixture", seenRefreshTokens[0])
	}
	if seenRefreshTokens[1] != "rotated-refresh-token" {
		t.Fatalf("second exchange refresh_token = %q, want rotated-refresh-token (rotation)", seenRefreshTokens[1])
	}
}

// TestAuthenticator_CachesTokenUntilNearExpiry asserts a cached token is
// reused across Apply calls and refreshed within 60s of expiry.
func TestAuthenticator_CachesTokenUntilNearExpiry(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok_cached","token_type":"bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	h := newTestHooks(func() time.Time { return now }, nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL, nil), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	doAuthenticatedRequest(t, auth)
	doAuthenticatedRequest(t, auth)
	if hits != 1 {
		t.Fatalf("token endpoint hits = %d, want 1 (cached token reused)", hits)
	}

	now = now.Add(3600*time.Second - 30*time.Second)
	doAuthenticatedRequest(t, auth)
	if hits != 2 {
		t.Fatalf("token endpoint hits = %d, want 2 (refreshed near expiry)", hits)
	}
}

// TestAuthenticator_MissingAccessTokenErrors asserts a malformed token
// response (no access_token) surfaces a clear, secret-free error.
func TestAuthenticator_MissingAccessTokenErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token_type":"bearer"}`))
	}))
	defer srv.Close()

	h := newClientHooks(nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL, nil), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err := auth.Apply(context.Background(), req); err == nil {
		t.Fatal("Apply: expected error for missing access_token, got nil")
	}
}

// TestAuthenticator_TokenEndpointErrorSurfaces asserts a non-2xx token
// response surfaces an error rather than silently proceeding unauthenticated.
func TestAuthenticator_TokenEndpointErrorSurfaces(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	h := newClientHooks(nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL, nil), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err := auth.Apply(context.Background(), req); err == nil {
		t.Fatal("Apply: expected error for 401 token endpoint response, got nil")
	}
}

// TestAuthenticator_RejectsNonHTTPTokenURL asserts a malformed/non-http(s)
// token_url fails closed.
func TestAuthenticator_RejectsNonHTTPTokenURL(t *testing.T) {
	h := newClientHooks(nil)
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"token_url": "ftp://example.com/token"},
		Secrets: map[string]string{
			"client_id":     "cid",
			"client_secret": "csecret",
			"refresh_token": "rtoken",
		},
	}
	if _, err := h.Authenticator(context.Background(), cfg, baseSpec()); err == nil {
		t.Fatal("Authenticator: expected error for non-http(s) token_url, got nil")
	}
}

// TestAuthenticator_RequiresRefreshToken asserts a missing refresh token
// secret is a clear configuration error.
func TestAuthenticator_RequiresRefreshToken(t *testing.T) {
	h := newClientHooks(nil)
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"token_url": "http://example.invalid/token"},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csecret"},
	}
	if _, err := h.Authenticator(context.Background(), cfg, baseSpec()); err == nil {
		t.Fatal("Authenticator: expected error for missing refresh_token, got nil")
	}
}
