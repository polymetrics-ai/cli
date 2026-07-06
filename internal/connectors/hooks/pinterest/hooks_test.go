package pinterest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func baseCfg(tokenURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{"token_url": tokenURL},
		Secrets: map[string]string{
			"client_id":     "client-id-fixture",
			"client_secret": "client-secret-fixture",
			"refresh_token": "refresh-token-fixture",
		},
	}
}

func baseSpec() engine.AuthSpec {
	return engine.AuthSpec{
		Mode:         "custom",
		Hook:         "pinterest",
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

// --- registration ---------------------------------------------------------

func TestHooksRegisteredUnderPinterest(t *testing.T) {
	h := engine.HooksFor("pinterest")
	if h == nil {
		t.Fatal(`engine.HooksFor("pinterest") = nil, want registered hooks (hooks/pinterest's init() must call engine.RegisterHooks)`)
	}
	if h.ConnectorName() != "pinterest" {
		t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), "pinterest")
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("registered pinterest hooks does not implement engine.AuthHook")
	}
}

// --- refresh-grant wire shape ----------------------------------------------

// TestAuthenticator_RefreshGrantUsesBasicClientAuth asserts the token request
// authenticates the client via HTTP Basic (Pinterest's exact wire shape,
// pinterest.go's accessToken), NOT form-encoded client_id/client_secret
// fields — this is the one behavioral difference from strava/gmail's shape.
func TestAuthenticator_RefreshGrantUsesBasicClientAuth(t *testing.T) {
	var gotAuthHeader string
	var gotForm url.Values
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		gotAuthHeader = r.Header.Get("Authorization")
		_ = r.ParseForm()
		gotForm = r.Form
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok_abc","token_type":"bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	h := newClientHooks(nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)

	if hits != 1 {
		t.Fatalf("token endpoint hits = %d, want 1", hits)
	}
	username, password, ok := parseBasicAuth(gotAuthHeader)
	if !ok {
		t.Fatalf("token request Authorization = %q, want HTTP Basic", gotAuthHeader)
	}
	if username != "client-id-fixture" || password != "client-secret-fixture" {
		t.Fatalf("basic auth = %q/%q, want client-id-fixture/client-secret-fixture", username, password)
	}
	if got := gotForm.Get("grant_type"); got != "refresh_token" {
		t.Fatalf("grant_type = %q, want refresh_token", got)
	}
	if got := gotForm.Get("refresh_token"); got != "refresh-token-fixture" {
		t.Fatalf("refresh_token = %q, want refresh-token-fixture", got)
	}
	// client_id/client_secret must NOT be sent as form fields (Basic auth only).
	if gotForm.Get("client_id") != "" || gotForm.Get("client_secret") != "" {
		t.Fatalf("form leaked client_id/client_secret: %v", gotForm)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer tok_abc" {
		t.Fatalf("data request Authorization = %q, want Bearer tok_abc", got)
	}
}

func parseBasicAuth(header string) (username, password string, ok bool) {
	req := &http.Request{Header: http.Header{"Authorization": []string{header}}}
	return req.BasicAuth()
}

// TestAuthenticator_CachesTokenUntilNearExpiry asserts a cached token is
// reused across Apply calls and is not refreshed until within 60s of expiry
// (matches legacy's a.now().Add(60*time.Second).Before(a.expires) check).
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
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	doAuthenticatedRequest(t, auth)
	doAuthenticatedRequest(t, auth)
	if hits != 1 {
		t.Fatalf("token endpoint hits = %d, want 1 (cached token reused)", hits)
	}

	// Advance to within 60s of expiry: must refresh.
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
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
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
	auth, err := h.Authenticator(context.Background(), baseCfg(srv.URL), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err := auth.Apply(context.Background(), req); err == nil {
		t.Fatal("Apply: expected error for 401 token endpoint response, got nil")
	}
}

// TestAuthenticator_RequiresTokenURL asserts an empty token_url is a
// validate-shaped error rather than a request to an empty/undefined URL.
func TestAuthenticator_RequiresTokenURL(t *testing.T) {
	h := newClientHooks(nil)
	spec := baseSpec()
	cfg := baseCfg("")
	if _, err := h.Authenticator(context.Background(), cfg, spec); err == nil {
		t.Fatal("Authenticator: expected error for empty token_url, got nil")
	}
}

// TestAuthenticator_RejectsNonHTTPTokenURL asserts a malformed/non-http(s)
// token_url fails closed rather than sending credentials to an unvalidated
// destination.
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

// TestAuthenticator_ErrorMessagesNeverLeakSecrets asserts no error string
// produced by this package embeds a configured secret value verbatim.
func TestAuthenticator_ErrorMessagesNeverLeakSecrets(t *testing.T) {
	h := newClientHooks(nil)
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"token_url": "not a url"},
		Secrets: map[string]string{
			"client_id":     "super-secret-client-id",
			"client_secret": "super-secret-client-secret",
			"refresh_token": "super-secret-refresh-token",
		},
	}
	_, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err == nil {
		t.Fatal("Authenticator: expected error for invalid token_url, got nil")
	}
	msg := err.Error()
	for _, secret := range []string{"super-secret-client-id", "super-secret-client-secret", "super-secret-refresh-token"} {
		if strings.Contains(msg, secret) {
			t.Fatalf("error message leaked a secret value: %q", msg)
		}
	}
}
