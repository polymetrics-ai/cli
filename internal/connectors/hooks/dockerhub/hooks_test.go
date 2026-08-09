// Package dockerhub implements the dockerhub AuthHook: a username+PAT-to-
// session-JWT exchange connsdk.Authenticator, per Docker Hub's documented
// POST /v2/users/login flow. This test file mirrors
// hooks/akeneo/hooks_test.go's structure.
package dockerhub

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

// fakeJWT builds a syntactically valid (unsigned) JWT whose payload carries
// only an `exp` claim `ttl` from now, so jwtTTL's decode path is exercised
// against a real three-segment token rather than a hand-rolled string.
func fakeJWT(t *testing.T, now time.Time, ttl time.Duration) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, err := json.Marshal(map[string]int64{"exp": now.Add(ttl).Unix()})
	if err != nil {
		t.Fatalf("marshal jwt payload: %v", err)
	}
	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".sig"
}

func baseCfg() connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"docker_username": "fixture-user",
		},
		Secrets: map[string]string{
			"docker_pat": "fixture-pat-value",
		},
	}
}

func baseSpec(tokenURL string) engine.AuthSpec {
	return engine.AuthSpec{
		Mode:     "custom",
		Hook:     "dockerhub",
		TokenURL: tokenURL,
		Username: "{{ config.docker_username }}",
		Password: "{{ secrets.docker_pat }}",
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
	return doAuthenticatedRequestToPath(t, auth, "/x")
}

func doAuthenticatedRequestToPath(t *testing.T, auth interface {
	Apply(ctx context.Context, req *http.Request) error
}, path string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "https://example.invalid"+path, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return req
}

// httpsTestServer wraps httptest.NewTLSServer with a client that trusts its
// certificate, since validateHTTPSURL requires https:// token URLs.
func httpsTestServer(t *testing.T, handler http.Handler) (*httptest.Server, *http.Client) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return srv, srv.Client()
}

func loginHTTPSServer(t *testing.T, respond func(body map[string]string) (int, map[string]any)) (*httptest.Server, *http.Client, *int32) {
	t.Helper()
	var hits int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if r.Method != http.MethodPost {
			t.Fatalf("login server: method = %s, want POST", r.Method)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("login server: decode body: %v", err)
		}
		status, respBody := respond(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if respBody == nil {
			respBody = map[string]any{"message": "error"}
		}
		_ = json.NewEncoder(w).Encode(respBody)
	})
	srv, client := httpsTestServer(t, handler)
	return srv, client, &hits
}

// --- registration -------------------------------------------------------

func TestHooksRegisteredUnderDockerhub(t *testing.T) {
	h := engine.HooksFor("dockerhub")
	if h == nil {
		t.Fatal(`engine.HooksFor("dockerhub") = nil, want registered hooks (hooks/dockerhub's init() must call engine.RegisterHooks)`)
	}
	if h.ConnectorName() != "dockerhub" {
		t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), "dockerhub")
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("registered dockerhub hooks does not implement engine.AuthHook")
	}
}

// --- login request shape -------------------------------------------------

func TestAuthenticator_LoginRequestShape(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var gotBody map[string]string
	srv, client, hits := loginHTTPSServer(t, func(body map[string]string) (int, map[string]any) {
		gotBody = body
		return http.StatusOK, map[string]any{"token": fakeJWT(t, now, time.Hour)}
	})

	h := newTestHooks(func() time.Time { return now }, client)
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec(srv.URL))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)

	if *hits != 1 {
		t.Fatalf("login endpoint hits = %d, want 1", *hits)
	}
	if gotBody["username"] != "fixture-user" {
		t.Fatalf("username = %q, want %q", gotBody["username"], "fixture-user")
	}
	if gotBody["password"] != "fixture-pat-value" {
		t.Fatalf("password = %q, want %q", gotBody["password"], "fixture-pat-value")
	}
	if len(gotBody) != 2 {
		t.Fatalf("login body = %v, want exactly {username,password}", gotBody)
	}
	wantToken := fakeJWT(t, now, time.Hour)
	if got := req.Header.Get("Authorization"); got != "Bearer "+wantToken {
		t.Fatalf("resource Authorization header = %q, want %q", got, "Bearer "+wantToken)
	}
}

// --- caching / expiry (from the JWT's own exp claim) ----------------------

func TestAuthenticator_CachesTokenAcrossRequests(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	srv, client, hits := loginHTTPSServer(t, func(map[string]string) (int, map[string]any) {
		return http.StatusOK, map[string]any{"token": fakeJWT(t, now, time.Hour)}
	})

	h := newTestHooks(func() time.Time { return now }, client)
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec(srv.URL))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	req1 := doAuthenticatedRequest(t, auth)
	req2 := doAuthenticatedRequest(t, auth)

	if *hits != 1 {
		t.Fatalf("login endpoint hits = %d, want 1 (second Apply should reuse the cached token)", *hits)
	}
	if req1.Header.Get("Authorization") != req2.Header.Get("Authorization") {
		t.Fatalf("Authorization headers differ across cached requests: %q vs %q", req1.Header.Get("Authorization"), req2.Header.Get("Authorization"))
	}
}

func TestAuthenticator_RefreshesWithin60sOfJWTExpiry(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	call := 0
	current := now
	srv, client, hits := loginHTTPSServer(t, func(map[string]string) (int, map[string]any) {
		call++
		return http.StatusOK, map[string]any{"token": fakeJWT(t, current, time.Hour)}
	})

	h := newTestHooks(func() time.Time { return current }, client)
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec(srv.URL))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	_ = doAuthenticatedRequest(t, auth) // primes the cache, JWT expires at now+3600s

	current = now.Add(3539 * time.Second)
	_ = doAuthenticatedRequest(t, auth)
	if *hits != 1 {
		t.Fatalf("login endpoint hits = %d after t+3539s, want 1 (still within cache window)", *hits)
	}

	current = now.Add(3541 * time.Second)
	_ = doAuthenticatedRequest(t, auth)
	if *hits != 2 {
		t.Fatalf("login endpoint hits = %d after t+3541s, want 2 (60s-early refresh must trigger)", *hits)
	}
}

func TestAuthenticator_FallsBackToFixedTTLWhenExpUnparseable(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	current := now
	srv, client, hits := loginHTTPSServer(t, func(map[string]string) (int, map[string]any) {
		return http.StatusOK, map[string]any{"token": "not-a-jwt"}
	})

	h := newTestHooks(func() time.Time { return current }, client)
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec(srv.URL))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	_ = doAuthenticatedRequest(t, auth)
	if *hits != 1 {
		t.Fatalf("login endpoint hits = %d, want 1", *hits)
	}

	// Still within the 4-minute fallback TTL (minus the 60s refresh window).
	current = now.Add(2 * time.Minute)
	_ = doAuthenticatedRequest(t, auth)
	if *hits != 1 {
		t.Fatalf("login endpoint hits = %d at t+2m, want 1 (within fallback TTL)", *hits)
	}

	// Past the fallback TTL's refresh window.
	current = now.Add(4 * time.Minute)
	_ = doAuthenticatedRequest(t, auth)
	if *hits != 2 {
		t.Fatalf("login endpoint hits = %d at t+4m, want 2 (fallback TTL exceeded)", *hits)
	}
}

// --- error paths ----------------------------------------------------------

func TestAuthenticator_NonSuccessLoginResponseIsError(t *testing.T) {
	srv, client, _ := loginHTTPSServer(t, func(map[string]string) (int, map[string]any) {
		return http.StatusUnauthorized, map[string]any{"message": "invalid credentials"}
	})

	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec(srv.URL))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "https://example.invalid/x", nil)
	err = auth.Apply(context.Background(), req)
	if err == nil {
		t.Fatal("Apply() error = nil, want an error for a non-2xx login endpoint response")
	}
	if req.Header.Get("Authorization") != "" {
		t.Fatalf("Authorization header set = %q after a failed login exchange, want empty (no silent unauthenticated fallback)", req.Header.Get("Authorization"))
	}
}

func TestAuthenticator_MissingTokenInResponseIsError(t *testing.T) {
	srv, client, _ := loginHTTPSServer(t, func(map[string]string) (int, map[string]any) {
		return http.StatusOK, map[string]any{}
	})

	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec(srv.URL))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "https://example.invalid/x", nil)
	if err := auth.Apply(context.Background(), req); err == nil {
		t.Fatal("Apply() error = nil, want an error for a login response missing token")
	}
}

func TestAuthenticator_MissingPasswordIsError(t *testing.T) {
	cfg := baseCfg()
	delete(cfg.Secrets, "docker_pat")

	h := New().(*Hooks)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec("https://example.invalid"))
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing docker_pat")
	}
	if !strings.Contains(err.Error(), "docker_pat") {
		t.Fatalf("error = %q, want it to name the missing docker_pat field", err.Error())
	}
}

func TestAuthenticator_MissingUsernameIsError(t *testing.T) {
	cfg := baseCfg()
	delete(cfg.Config, "docker_username")

	h := New().(*Hooks)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec("https://example.invalid"))
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error naming the missing docker_username")
	}
	if !strings.Contains(err.Error(), "docker_username") {
		t.Fatalf("error = %q, want it to name docker_username", err.Error())
	}
}

func TestAuthenticator_TokenURLRejectsPlainHTTP(t *testing.T) {
	cfg := baseCfg()
	h := newClientHooks(nil)
	_, err := h.Authenticator(context.Background(), cfg, baseSpec("http://insecure.example.invalid/users/login"))
	if err == nil {
		t.Fatal("Authenticator() error = nil, want a fail-closed error for a plain-http token_url")
	}
	if !strings.Contains(err.Error(), "https") {
		t.Fatalf("error = %q, want it to mention the https requirement", err.Error())
	}
}

func TestAuthenticator_TokenURLUnparseableIsError(t *testing.T) {
	cfg := baseCfg()
	h := New().(*Hooks)
	if _, err := h.Authenticator(context.Background(), cfg, baseSpec("://not-a-url")); err == nil {
		t.Fatal("Authenticator() error = nil, want an error for an unparseable token_url")
	}
}

// --- ctx cancellation -----------------------------------------------------

func TestAuthenticator_HonorsContextCancellation(t *testing.T) {
	srv, client, _ := loginHTTPSServer(t, func(map[string]string) (int, map[string]any) {
		return http.StatusOK, map[string]any{"token": fakeJWT(t, time.Now(), time.Hour)}
	})

	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec(srv.URL))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req, _ := http.NewRequest(http.MethodGet, "https://example.invalid/x", nil)
	if err := auth.Apply(ctx, req); err == nil {
		t.Fatal("Apply(cancelled ctx) error = nil, want a cancellation error (ctx must be honored, not context.Background())")
	}
}

func TestAuthenticator_HonorsContextCancellationBeforeAnyRequest(t *testing.T) {
	h := newClientHooks(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := h.Authenticator(ctx, baseCfg(), baseSpec("https://example.invalid")); err == nil {
		t.Fatal("Authenticator(cancelled ctx) error = nil, want a cancellation error")
	}
}

// --- secret redaction ------------------------------------------------------

func TestAuthenticator_ErrorsNeverContainSecretText(t *testing.T) {
	const secretMarkerPAT = "fixture-pat-value"

	srv, client, _ := loginHTTPSServer(t, func(map[string]string) (int, map[string]any) {
		return http.StatusUnauthorized, map[string]any{"message": "invalid credentials"}
	})

	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec(srv.URL))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "https://example.invalid/x", nil)
	err = auth.Apply(context.Background(), req)
	if err == nil {
		t.Fatal("expected an error from the 401 login response")
	}
	if strings.Contains(err.Error(), secretMarkerPAT) {
		t.Fatalf("error text contains secret marker %q: %s", secretMarkerPAT, err.Error())
	}
}

// --- jwtTTL unit coverage ---------------------------------------------------

func TestJWTTTL_ValidExpClaim(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tok := fakeJWT(t, now, 90*time.Second)
	got := jwtTTL(tok, now)
	if got < 89*time.Second || got > 90*time.Second {
		t.Fatalf("jwtTTL = %v, want ~90s", got)
	}
}

func TestJWTTTL_MalformedTokenFallsBack(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, tok := range []string{"", "a.b", "a.b.c.d", "a." + base64.RawURLEncoding.EncodeToString([]byte("not json")) + ".c"} {
		if got := jwtTTL(tok, now); got != fallbackTokenTTL {
			t.Errorf("jwtTTL(%q) = %v, want fallbackTokenTTL (%v)", tok, got, fallbackTokenTTL)
		}
	}
}

func TestJWTTTL_ExpiredClaimFallsBack(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tok := fakeJWT(t, now, -time.Hour) // already-expired exp
	if got := jwtTTL(tok, now); got != fallbackTokenTTL {
		t.Fatalf("jwtTTL(expired) = %v, want fallbackTokenTTL (%v)", got, fallbackTokenTTL)
	}
}

func TestJWTTTL_NonNumericExpFallsBack(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"exp":"not-a-number"}`))
	tok := "h." + payload + ".sig"
	if got := jwtTTL(tok, now); got != fallbackTokenTTL {
		t.Fatalf("jwtTTL(non-numeric exp) = %v, want fallbackTokenTTL (%v)", got, fallbackTokenTTL)
	}
}

// --- dualAuth: SCIM path routing to a second, independent credential -------

func scimCfg() connectors.RuntimeConfig {
	cfg := baseCfg()
	cfg.Secrets["scim_bearer_token"] = "fixture-scim-token"
	return cfg
}

func TestAuthenticator_SCIMPathUsesScimBearerToken(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var loginHits int32
	srv, client, _ := loginHTTPSServer(t, func(map[string]string) (int, map[string]any) {
		loginHits++
		return http.StatusOK, map[string]any{"token": fakeJWT(t, now, time.Hour)}
	})

	h := newTestHooks(func() time.Time { return now }, client)
	auth, err := h.Authenticator(context.Background(), scimCfg(), baseSpec(srv.URL))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	req := doAuthenticatedRequestToPath(t, auth, "/v2/scim/2.0/Users")
	if got := req.Header.Get("Authorization"); got != "Bearer fixture-scim-token" {
		t.Fatalf("SCIM request Authorization = %q, want the static SCIM bearer token, not the session JWT", got)
	}
	if loginHits != 0 {
		t.Fatalf("login endpoint hits = %d, want 0 (a SCIM-path request must never trigger the session login exchange)", loginHits)
	}
}

func TestAuthenticator_NonSCIMPathUsesSessionJWTEvenWithScimTokenConfigured(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	srv, client, hits := loginHTTPSServer(t, func(map[string]string) (int, map[string]any) {
		return http.StatusOK, map[string]any{"token": fakeJWT(t, now, time.Hour)}
	})

	h := newTestHooks(func() time.Time { return now }, client)
	auth, err := h.Authenticator(context.Background(), scimCfg(), baseSpec(srv.URL))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	req := doAuthenticatedRequestToPath(t, auth, "/v2/access-tokens")
	wantToken := fakeJWT(t, now, time.Hour)
	if got := req.Header.Get("Authorization"); got != "Bearer "+wantToken {
		t.Fatalf("non-SCIM request Authorization = %q, want the session JWT even though a SCIM token is configured", got)
	}
	if *hits != 1 {
		t.Fatalf("login endpoint hits = %d, want 1", *hits)
	}
}

func TestAuthenticator_SCIMPathWithNoScimTokenFailsClosed(t *testing.T) {
	h := newClientHooks(nil)
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec("https://example.invalid"))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.invalid/v2/scim/2.0/Users", nil)
	err = auth.Apply(context.Background(), req)
	if err == nil {
		t.Fatal("Apply() error = nil, want an error for a SCIM request with no scim_bearer_token configured")
	}
	if !strings.Contains(err.Error(), "scim_bearer_token") {
		t.Fatalf("error = %q, want it to name scim_bearer_token", err.Error())
	}
	if req.Header.Get("Authorization") != "" {
		t.Fatalf("Authorization header set = %q for a failed SCIM auth resolution, want empty (no silent fallback to the session JWT)", req.Header.Get("Authorization"))
	}
}

func TestAuthenticator_SCIMPathNeverFallsBackToSessionJWT(t *testing.T) {
	// Even if a session JWT were somehow already cached, a SCIM request must
	// never receive it — the two credentials are never interchangeable.
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	srv, client, _ := loginHTTPSServer(t, func(map[string]string) (int, map[string]any) {
		return http.StatusOK, map[string]any{"token": fakeJWT(t, now, time.Hour)}
	})

	h := newTestHooks(func() time.Time { return now }, client)
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec(srv.URL))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	// Prime the session cache via a non-SCIM request first.
	_ = doAuthenticatedRequestToPath(t, auth, "/v2/access-tokens")

	// A SCIM request with no scim_bearer_token configured must still fail
	// closed, not reuse the now-cached session JWT.
	req, _ := http.NewRequest(http.MethodGet, "https://example.invalid/v2/scim/2.0/Users", nil)
	err = auth.Apply(context.Background(), req)
	if err == nil {
		t.Fatal("Apply() error = nil for a SCIM request with no scim_bearer_token, want an error even with a cached session JWT available")
	}
	if req.Header.Get("Authorization") != "" {
		t.Fatalf("Authorization = %q, want empty (must not fall back to the cached session JWT)", req.Header.Get("Authorization"))
	}
}

// --- SCIM-only connections (scim_bearer_token configured, docker_pat absent) --
//
// streams.json declares a second custom-auth spec gated on
// secrets.scim_bearer_token that carries no login fields, because the `when`
// grammar has no OR operator. These fixtures pin what that spec resolves to.

// scimOnlyCfg is a connection configured with ONLY the SCIM credential: no
// docker_pat secret at all.
func scimOnlyCfg() connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config:  map[string]string{"docker_username": "fixture-user"},
		Secrets: map[string]string{"scim_bearer_token": "fixture-scim-token"},
	}
}

// scimOnlySpec mirrors streams.json's scim_bearer_token-gated auth spec: a
// custom spec with no token_url/username/password.
func scimOnlySpec() engine.AuthSpec {
	return engine.AuthSpec{Mode: "custom", Hook: "dockerhub"}
}

func TestAuthenticator_SCIMOnlyConnectionAuthenticatesSCIMRequests(t *testing.T) {
	h := newClientHooks(nil)
	auth, err := h.Authenticator(context.Background(), scimOnlyCfg(), scimOnlySpec())
	if err != nil {
		t.Fatalf("Authenticator with a SCIM-only connection: %v", err)
	}

	req := doAuthenticatedRequestToPath(t, auth, "/v2/scim/2.0/Users")
	if got := req.Header.Get("Authorization"); got != "Bearer fixture-scim-token" {
		t.Fatalf("SCIM request Authorization = %q, want the SCIM bearer token with no docker_pat configured", got)
	}
}

func TestAuthenticator_SCIMOnlyConnectionFailsClosedOnNonSCIMPath(t *testing.T) {
	h := newClientHooks(nil)
	auth, err := h.Authenticator(context.Background(), scimOnlyCfg(), scimOnlySpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.invalid/v2/access-tokens", nil)
	err = auth.Apply(context.Background(), req)
	if err == nil {
		t.Fatal("Apply() error = nil for a non-SCIM request on a SCIM-only connection, want a closed failure naming docker_pat")
	}
	if !strings.Contains(err.Error(), "docker_pat") {
		t.Fatalf("error = %q, want it to name docker_pat", err.Error())
	}
	if req.Header.Get("Authorization") != "" {
		t.Fatalf("Authorization = %q, want empty (the SCIM token must never sign a bearerAuth endpoint)", req.Header.Get("Authorization"))
	}
}

func TestAuthenticator_NoCredentialConfiguredIsError(t *testing.T) {
	cfg := connectors.RuntimeConfig{Config: map[string]string{"docker_username": "fixture-user"}}
	h := newClientHooks(nil)
	if _, err := h.Authenticator(context.Background(), cfg, scimOnlySpec()); err == nil {
		t.Fatal("Authenticator error = nil with neither docker_pat nor scim_bearer_token configured, want an error")
	}
}

// --- SCIM routing under a proxy base_url path prefix ------------------------

func TestAuthenticator_SCIMRoutingHonorsProxyBaseURLPathPrefix(t *testing.T) {
	cfg := scimOnlyCfg()
	cfg.Config["base_url"] = "https://proxy.internal/dockerhub/v2"

	h := newClientHooks(nil)
	auth, err := h.Authenticator(context.Background(), cfg, scimOnlySpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	req := doAuthenticatedRequestToPath(t, auth, "/dockerhub/v2/scim/2.0/Users")
	if got := req.Header.Get("Authorization"); got != "Bearer fixture-scim-token" {
		t.Fatalf("proxy SCIM Authorization = %q, want the SCIM bearer token", got)
	}
	doubled, err := http.NewRequest(http.MethodGet, "https://example.invalid/dockerhub/v2/v2/scim/2.0/Users", nil)
	if err != nil {
		t.Fatalf("build doubled proxy request: %v", err)
	}
	if err := auth.Apply(context.Background(), doubled); err == nil {
		t.Fatal("doubled proxy path authenticated as SCIM, want a closed non-SCIM failure")
	}
}

func TestAuthenticator_SCIMRoutingNormalizesProxyAPIRoot(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	srv, client, hits := loginHTTPSServer(t, func(map[string]string) (int, map[string]any) {
		return http.StatusOK, map[string]any{"token": fakeJWT(t, now, time.Hour)}
	})
	for _, tt := range []struct {
		name string
		cfg  connectors.RuntimeConfig
		spec engine.AuthSpec
	}{
		{name: "SCIM-only", cfg: scimOnlyCfg(), spec: scimOnlySpec()},
		{name: "dual credential", cfg: scimCfg(), spec: baseSpec(srv.URL)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.cfg.Config["base_url"] = "https://proxy.internal/dockerhub"
			h := newTestHooks(func() time.Time { return now }, client)
			auth, err := h.Authenticator(context.Background(), tt.cfg, tt.spec)
			if err != nil {
				t.Fatalf("Authenticator: %v", err)
			}
			req := doAuthenticatedRequestToPath(t, auth, "/dockerhub/v2/scim/2.0/Users")
			if got := req.Header.Get("Authorization"); got != "Bearer fixture-scim-token" {
				t.Fatalf("proxy SCIM Authorization = %q, want SCIM bearer token", got)
			}
		})
	}
	if *hits != 0 {
		t.Fatalf("login endpoint hits = %d, want 0 for normalized SCIM routing", *hits)
	}
}

func TestAuthenticator_LoginRedirectDoesNotReplayPAT(t *testing.T) {
	var destinationHits int32
	server, client := httpsTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect-target" {
			atomic.AddInt32(&destinationHits, 1)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/redirect-target", http.StatusTemporaryRedirect)
	}))
	h := newClientHooks(client)
	auth, err := h.Authenticator(context.Background(), baseCfg(), baseSpec(server.URL))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, "https://example.invalid/x", nil)
	if err != nil {
		t.Fatalf("build resource request: %v", err)
	}
	err = auth.Apply(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "307") {
		t.Fatalf("redirect login error = %v, want rejected 307", err)
	}
	if got := atomic.LoadInt32(&destinationHits); got != 0 {
		t.Fatalf("redirect target hits = %d, want 0 so the PAT body is never replayed", got)
	}
}

func TestAuthenticator_ProxyBaseURLNonSCIMPathStillUsesSessionJWT(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	srv, client, _ := loginHTTPSServer(t, func(map[string]string) (int, map[string]any) {
		return http.StatusOK, map[string]any{"token": fakeJWT(t, now, time.Hour)}
	})

	cfg := scimCfg()
	cfg.Config["base_url"] = "https://proxy.internal/dockerhub/v2"

	h := newTestHooks(func() time.Time { return now }, client)
	auth, err := h.Authenticator(context.Background(), cfg, baseSpec(srv.URL))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	req := doAuthenticatedRequestToPath(t, auth, "/dockerhub/v2/access-tokens")
	if got := req.Header.Get("Authorization"); got != "Bearer "+fakeJWT(t, now, time.Hour) {
		t.Fatalf("proxy non-SCIM Authorization = %q, want the session JWT", got)
	}
}

func TestAuthenticator_RepositoryPathNamedSCIMUsesSessionJWT(t *testing.T) {
	// Namespace "scim" + repository "2.0" produces a path CONTAINING
	// "/scim/2.0/" that is not a SCIM endpoint: routing is anchored on the
	// base path, so this must keep using the session JWT.
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	srv, client, _ := loginHTTPSServer(t, func(map[string]string) (int, map[string]any) {
		return http.StatusOK, map[string]any{"token": fakeJWT(t, now, time.Hour)}
	})

	cfg := scimCfg()
	cfg.Config["base_url"] = "https://hub.docker.com/v2"

	h := newTestHooks(func() time.Time { return now }, client)
	auth, err := h.Authenticator(context.Background(), cfg, baseSpec(srv.URL))
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	req := doAuthenticatedRequestToPath(t, auth, "/v2/repositories/scim/2.0/tags")
	if got := req.Header.Get("Authorization"); got != "Bearer "+fakeJWT(t, now, time.Hour) {
		t.Fatalf("repository path Authorization = %q, want the session JWT (the SCIM token must not leak to a bearerAuth endpoint)", got)
	}
}
