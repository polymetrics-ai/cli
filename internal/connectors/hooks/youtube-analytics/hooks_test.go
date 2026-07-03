// Package youtubeanalytics implements the youtube-analytics AuthHook
// (docs/migration/quarantine.json AUTH_COMPLEX entry): an OAuth 2.0
// refresh-token-grant connsdk.Authenticator, ported from legacy
// internal/connectors/youtube-analytics/auth.go's oauthRefreshAuth. This
// test file mirrors hooks/gmail/hooks_test.go's precedent byte-for-byte,
// adjusted only for this bundle's secret names (client_id/client_secret/
// refresh_token, no client_refresh_token indirection).
package youtubeanalytics

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

// tokenServer builds an httptest.NewTLSServer standing in for Google's OAuth
// token endpoint (THREAT-MODEL.md Delta 2: the hook requires token_url to be
// https, so the replay server must actually speak TLS, not plain http — an
// httptest TLS server on loopback only). respond is invoked per request; it
// receives the decoded form body and returns (status, body-or-nil). A nil
// body writes {} on a non-2xx status. The returned *http.Client trusts the
// server's self-signed test certificate and must be set on Hooks.Client.
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
		Config: map[string]string{"token_url": tokenURL, "scopes": "https://www.googleapis.com/auth/yt-analytics.readonly"},
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
		Hook:         "youtube-analytics",
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ secrets.client_id }}",
		ClientSecret: "{{ secrets.client_secret }}",
		Token:        "{{ secrets.refresh_token }}",
		Scopes:       "{{ config.scopes }}",
	}
}

func newTestHooks(now func() time.Time, client *http.Client) *Hooks {
	h := New().(*Hooks)
	h.Now = now
	h.Client = client
	return h
}

// newClientHooks returns a Hooks wired only with client (no clock override),
// for tests that don't exercise caching/expiry.
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

func TestHooksRegisteredUnderYoutubeAnalytics(t *testing.T) {
	h := engine.HooksFor("youtube-analytics")
	if h == nil {
		t.Fatal("engine.HooksFor(\"youtube-analytics\") = nil, want registered hooks (hooks/youtube-analytics's init() must call engine.RegisterHooks)")
	}
	if h.ConnectorName() != "youtube-analytics" {
		t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), "youtube-analytics")
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("registered youtube-analytics hooks does not implement engine.AuthHook")
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
	if got := gotForm.Get("scope"); got != "https://www.googleapis.com/auth/yt-analytics.readonly" {
		t.Fatalf("scope = %q, want the configured scope", got)
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
		t.Fatalf("form has client_secret key = %v, want omitted entirely when unset", gotForm["client_secret"])
	}
}

// TestAuthenticator_ScopeOmittedWhenUnset mirrors legacy's `if a.scope !=
// ""` guard.
func TestAuthenticator_ScopeOmittedWhenUnset(t *testing.T) {
	var gotForm url.Values
	srv, client, _ := tokenServer(t, func(form url.Values) (int, map[string]any) {
		gotForm = form
		return http.StatusOK, map[string]any{"access_token": "tok_abc", "expires_in": 3600}
	})

	cfg := baseCfg(srv.URL)
	delete(cfg.Config, "scopes")

	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	_ = doAuthenticatedRequest(t, auth)

	if _, ok := gotForm["scope"]; ok {
		t.Fatalf("form has scope key = %v, want omitted entirely when unset", gotForm["scope"])
	}
}

// TestInterpolateOptional_AnyErrorResolvesToEmptyString pins
// interpolateOptional's actual behavior (matches gmail's hooks.go/
// hooks_test.go precedent exactly): a genuinely-absent key, a
// CRLF-injecting resolved value, and an unknown filter/namespace reference
// all resolve to "" rather than propagating an error.
func TestInterpolateOptional_AnyErrorResolvesToEmptyString(t *testing.T) {
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"crlf_value": "bad\r\nvalue"},
		Secrets: map[string]string{},
	}

	t.Run("genuinely absent key resolves to empty string", func(t *testing.T) {
		if got := interpolateOptional("{{ config.nope }}", cfg); got != "" {
			t.Fatalf("interpolateOptional(absent key) = %q, want \"\"", got)
		}
	})

	t.Run("CRLF injection in the resolved value ALSO resolves to empty string, not propagated", func(t *testing.T) {
		if got := interpolateOptional("{{ config.crlf_value }}", cfg); got != "" {
			t.Fatalf("interpolateOptional(CRLF value) = %q, want \"\"", got)
		}
	})

	t.Run("unknown filter ALSO resolves to empty string, not propagated", func(t *testing.T) {
		if got := interpolateOptional("{{ config.crlf_value | bogus_filter }}", cfg); got != "" {
			t.Fatalf("interpolateOptional(unknown filter) = %q, want \"\"", got)
		}
	})

	t.Run("a value that resolves cleanly is returned unchanged", func(t *testing.T) {
		clean := connectors.RuntimeConfig{Config: map[string]string{"clean": "hello"}}
		if got := interpolateOptional("{{ config.clean }}", clean); got != "hello" {
			t.Fatalf("interpolateOptional(clean value) = %q, want %q", got, "hello")
		}
	})
}

// --- caching / expiry ----------------------------------------------------

// TestAuthenticator_CachesTokenAcrossRequests: a second Apply before expiry
// must NOT hit the token endpoint again.
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
// stale starting 60s before the declared expiry (legacy auth.go).
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

	// Advance to exactly 60s before expiry: still cache-fresh per legacy's
	// "Add(60s).Before(expires)" boundary semantics (3600-60=3540 is the
	// last still-cached instant; anything at or past 3541s forces refresh).
	current = now.Add(3539 * time.Second)
	_ = doAuthenticatedRequest(t, auth)
	if *hits != 1 {
		t.Fatalf("token endpoint hits = %d after t+3539s, want 1 (still within cache window)", *hits)
	}

	// Now within 60s of expiry: must refresh.
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

// TestAuthenticator_MissingRefreshTokenIsError: templated AuthSpec fields
// are resolved (and thus validated) at Authenticator()-build time, matching
// every other engine auth mode's eager Interpolate-at-build-time behavior.
func TestAuthenticator_MissingRefreshTokenIsError(t *testing.T) {
	cfg := baseCfg("https://oauth2.googleapis.com/token")
	delete(cfg.Secrets, "refresh_token")

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

// TestAuthenticator_TokenURLMustBeHTTPS is the THREAT-MODEL.md Delta 2 guard:
// a non-https token_url override must fail closed rather than send secrets
// to an arbitrary endpoint.
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
		t.Fatal("Apply(cancelled ctx) error = nil, want a cancellation error (F8: ctx must be honored, not context.Background())")
	}
}

// --- secret redaction ------------------------------------------------------

// TestAuthenticator_ErrorsNeverContainSecretText asserts none of the error
// paths above ever leak client_secret/refresh_token/client_id values or the
// resolved access token into an error string.
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
