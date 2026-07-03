package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

func cfgWith(config, secrets map[string]string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{Config: config, Secrets: secrets}
}

func applyToRequest(t *testing.T, auth connsdk.Authenticator, req *http.Request) {
	t.Helper()
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
}

func TestSelectAuthBearerWhenMatches(t *testing.T) {
	specs := []AuthSpec{
		{Mode: "bearer", Token: "{{ secrets.token }}", When: "{{ config.auth_type in ['auto','token'] }}"},
	}
	cfg := cfgWith(map[string]string{"auth_type": "auto"}, map[string]string{"token": "sekret"})

	auth, err := selectAuth(context.Background(), cfg, specs, nil)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
	applyToRequest(t, auth, req)
	if got := req.Header.Get("Authorization"); got != "Bearer sekret" {
		t.Fatalf("Authorization header = %q, want %q", got, "Bearer sekret")
	}
}

func TestSelectAuthNoneWhenPublic(t *testing.T) {
	specs := []AuthSpec{
		{Mode: "bearer", Token: "{{ secrets.token }}", When: "{{ config.auth_type in ['auto','token'] }}"},
		{Mode: "none", When: "{{ config.auth_type == 'public' }}"},
	}
	cfg := cfgWith(map[string]string{"auth_type": "public"}, nil)

	auth, err := selectAuth(context.Background(), cfg, specs, nil)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
	applyToRequest(t, auth, req)
	if got := req.Header.Get("Authorization"); got != "" {
		t.Fatalf("Authorization header = %q, want empty for none mode", got)
	}
}

func TestSelectAuthBasic(t *testing.T) {
	specs := []AuthSpec{
		{Mode: "basic", Username: "{{ config.user }}", Password: "{{ secrets.pass }}"},
	}
	cfg := cfgWith(map[string]string{"user": "alice"}, map[string]string{"pass": "hunter2"})

	auth, err := selectAuth(context.Background(), cfg, specs, nil)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
	applyToRequest(t, auth, req)
	wantUser, wantPass, ok := req.BasicAuth()
	if !ok || wantUser != "alice" || wantPass != "hunter2" {
		t.Fatalf("BasicAuth() = (%q, %q, %v), want (alice, hunter2, true)", wantUser, wantPass, ok)
	}
}

func TestSelectAuthAPIKeyHeaderWithPrefix(t *testing.T) {
	specs := []AuthSpec{
		{Mode: "api_key_header", Header: "X-API-Key", Prefix: "Token ", Value: "{{ secrets.key }}"},
	}
	cfg := cfgWith(nil, map[string]string{"key": "abc123"})

	auth, err := selectAuth(context.Background(), cfg, specs, nil)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
	applyToRequest(t, auth, req)
	if got := req.Header.Get("X-API-Key"); got != "Token abc123" {
		t.Fatalf("X-API-Key header = %q, want %q", got, "Token abc123")
	}
}

func TestSelectAuthAPIKeyQuery(t *testing.T) {
	specs := []AuthSpec{
		{Mode: "api_key_query", Param: "api_key", Value: "{{ secrets.key }}"},
	}
	cfg := cfgWith(nil, map[string]string{"key": "abc123"})

	auth, err := selectAuth(context.Background(), cfg, specs, nil)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
	applyToRequest(t, auth, req)
	if got := req.URL.Query().Get("api_key"); got != "abc123" {
		t.Fatalf("api_key query param = %q, want abc123", got)
	}
}

func TestSelectAuthOAuth2ClientCredentialsFetchesAndCachesToken(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok-` + http.MethodGet + `","token_type":"bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	specs := []AuthSpec{
		{
			Mode:         "oauth2_client_credentials",
			TokenURL:     "{{ config.token_url }}",
			ClientID:     "{{ config.client_id }}",
			ClientSecret: "{{ secrets.client_secret }}",
		},
	}
	cfg := cfgWith(
		map[string]string{"token_url": srv.URL, "client_id": "cid"},
		map[string]string{"client_secret": "csecret"},
	)

	auth, err := selectAuth(context.Background(), cfg, specs, nil)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
	applyToRequest(t, auth, req)
	if got := req.Header.Get("Authorization"); got == "" {
		t.Fatalf("Authorization header empty after oauth2 token fetch")
	}

	// Second Apply within the token's validity window must not refetch.
	req2, _ := http.NewRequest(http.MethodGet, "https://api.example.com/y", nil)
	applyToRequest(t, auth, req2)
	if calls != 1 {
		t.Fatalf("token endpoint called %d times, want 1 (cached token reused)", calls)
	}
}

// fakeAuthHook is a minimal Hooks+AuthHook fake for the custom-mode test.
type fakeAuthHook struct {
	name string
	auth connsdk.Authenticator
	err  error
}

func (f *fakeAuthHook) ConnectorName() string { return f.name }
func (f *fakeAuthHook) Authenticator(_ context.Context, _ connectors.RuntimeConfig, _ AuthSpec) (connsdk.Authenticator, error) {
	return f.auth, f.err
}

func TestSelectAuthCustomResolvesAuthHook(t *testing.T) {
	want := connsdk.Bearer("hook-token")
	hooks := &fakeAuthHook{name: "acme", auth: want}

	specs := []AuthSpec{
		{Mode: "custom", Hook: "acme"},
	}
	cfg := cfgWith(nil, nil)

	auth, err := selectAuth(context.Background(), cfg, specs, hooks)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}
	if auth != want {
		t.Fatalf("selectAuth() did not return the AuthHook's authenticator")
	}
}

// --- F8 (REVIEW.md flag): AuthHook.Authenticator must be called with the
// CALLER's ctx, not context.Background() — a github_app JWT->installation-
// token exchange (network call) needs to honor cancellation/deadlines. ---

type ctxProbeKey struct{}

// ctxProbeAuthHook is a fake AuthHook that stashes the ctx it actually
// received (via Value lookup on ctxProbeKey) into probe, so the test can
// assert selectAuth threaded the CALLER's context through rather than
// substituting context.Background().
type ctxProbeAuthHook struct {
	name  string
	probe *string
}

func (f *ctxProbeAuthHook) ConnectorName() string { return f.name }
func (f *ctxProbeAuthHook) Authenticator(ctx context.Context, _ connectors.RuntimeConfig, _ AuthSpec) (connsdk.Authenticator, error) {
	if v, ok := ctx.Value(ctxProbeKey{}).(string); ok {
		*f.probe = v
	} else {
		*f.probe = ""
	}
	return connsdk.Bearer("hook-token"), nil
}

func TestSelectAuthCustomThreadsCallerContext(t *testing.T) {
	var got string
	hooks := &ctxProbeAuthHook{name: "acme", probe: &got}

	specs := []AuthSpec{{Mode: "custom", Hook: "acme"}}
	cfg := cfgWith(nil, nil)

	ctx := context.WithValue(context.Background(), ctxProbeKey{}, "caller-marker")
	_, err := selectAuth(ctx, cfg, specs, hooks)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}
	if got != "caller-marker" {
		t.Fatalf("AuthHook.Authenticator received ctx value = %q, want %q (selectAuth must thread the caller's ctx, not context.Background())", got, "caller-marker")
	}
}

func TestSelectAuthCustomMissingHookIsTypedError(t *testing.T) {
	specs := []AuthSpec{
		{Mode: "custom", Hook: "acme"},
	}
	cfg := cfgWith(nil, nil)

	// nil Hooks: no hook set registered at all.
	_, err := selectAuth(context.Background(), cfg, specs, nil)
	if err == nil {
		t.Fatalf("selectAuth() error = nil, want error naming missing hook %q", "acme")
	}

	// Hooks present but does not implement AuthHook.
	_, err2 := selectAuth(context.Background(), cfg, specs, plainHooks{name: "acme"})
	if err2 == nil {
		t.Fatalf("selectAuth() error = nil, want error when Hooks does not implement AuthHook")
	}
}

// plainHooks implements only Hooks (no AuthHook), for the missing-capability
// error-path test.
type plainHooks struct{ name string }

func (p plainHooks) ConnectorName() string { return p.name }

func TestSelectAuthNoRuleMatchesIsTypedError(t *testing.T) {
	specs := []AuthSpec{
		{Mode: "bearer", Token: "{{ secrets.token }}", When: "{{ config.auth_type == 'token' }}"},
	}
	cfg := cfgWith(map[string]string{"auth_type": "public"}, nil)

	_, err := selectAuth(context.Background(), cfg, specs, nil)
	if err == nil {
		t.Fatalf("selectAuth() error = nil, want typed error naming auth_type when no rule matches")
	}
}

func TestSelectAuthEmptySpecsIsError(t *testing.T) {
	_, err := selectAuth(context.Background(), cfgWith(nil, nil), nil, nil)
	if err == nil {
		t.Fatalf("selectAuth() error = nil, want error for empty spec list")
	}
}

func TestSelectAuthFirstDeclaredRuleWinsOnOrdering(t *testing.T) {
	// Two rules both with no "when" (always match): first declared wins.
	specs := []AuthSpec{
		{Mode: "api_key_header", Header: "X-First", Value: "first"},
		{Mode: "api_key_header", Header: "X-Second", Value: "second"},
	}
	cfg := cfgWith(nil, nil)

	auth, err := selectAuth(context.Background(), cfg, specs, nil)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
	applyToRequest(t, auth, req)
	if got := req.Header.Get("X-First"); got != "first" {
		t.Fatalf("X-First header = %q, want %q (first declared rule should win)", got, "first")
	}
	if got := req.Header.Get("X-Second"); got != "" {
		t.Fatalf("X-Second header = %q, want empty (second rule should not apply)", got)
	}
}

func TestSelectAuthGithubStyleAutoTokenPublicTable(t *testing.T) {
	specs := []AuthSpec{
		{Mode: "bearer", Token: "{{ secrets.token }}", When: "{{ config.auth_type in ['auto','token'] }}"},
		{Mode: "custom", Hook: "github_app", When: "{{ config.auth_type == 'github_app' }}"},
		{Mode: "none", When: "{{ config.auth_type == 'public' }}"},
	}

	t.Run("auto with token set uses bearer", func(t *testing.T) {
		cfg := cfgWith(map[string]string{"auth_type": "auto"}, map[string]string{"token": "tok"})
		auth, err := selectAuth(context.Background(), cfg, specs, nil)
		if err != nil {
			t.Fatalf("selectAuth() error = %v", err)
		}
		req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
		applyToRequest(t, auth, req)
		if got := req.Header.Get("Authorization"); got != "Bearer tok" {
			t.Fatalf("Authorization = %q, want Bearer tok", got)
		}
	})

	t.Run("token explicit uses bearer", func(t *testing.T) {
		cfg := cfgWith(map[string]string{"auth_type": "token"}, map[string]string{"token": "tok2"})
		auth, err := selectAuth(context.Background(), cfg, specs, nil)
		if err != nil {
			t.Fatalf("selectAuth() error = %v", err)
		}
		req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
		applyToRequest(t, auth, req)
		if got := req.Header.Get("Authorization"); got != "Bearer tok2" {
			t.Fatalf("Authorization = %q, want Bearer tok2", got)
		}
	})

	t.Run("public uses none", func(t *testing.T) {
		cfg := cfgWith(map[string]string{"auth_type": "public"}, nil)
		auth, err := selectAuth(context.Background(), cfg, specs, nil)
		if err != nil {
			t.Fatalf("selectAuth() error = %v", err)
		}
		req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
		applyToRequest(t, auth, req)
		if got := req.Header.Get("Authorization"); got != "" {
			t.Fatalf("Authorization = %q, want empty for public", got)
		}
	})

	t.Run("github_app resolves custom hook", func(t *testing.T) {
		want := connsdk.Bearer("installation-token")
		hooks := &fakeAuthHook{name: "github_app", auth: want}
		cfg := cfgWith(map[string]string{"auth_type": "github_app"}, nil)
		auth, err := selectAuth(context.Background(), cfg, specs, hooks)
		if err != nil {
			t.Fatalf("selectAuth() error = %v", err)
		}
		if auth != want {
			t.Fatalf("selectAuth() did not resolve the github_app AuthHook")
		}
	})
}

func TestSelectAuthSecretsNeverInError(t *testing.T) {
	specs := []AuthSpec{
		{Mode: "bearer", Token: "{{ secrets.missing }}"},
	}
	cfg := cfgWith(nil, map[string]string{"other": "sk_live_shouldnotleak"})

	_, err := selectAuth(context.Background(), cfg, specs, nil)
	if err == nil {
		t.Fatalf("selectAuth() error = nil, want unresolved-key error")
	}
	if got := err.Error(); contains(got, "sk_live_shouldnotleak") {
		t.Fatalf("selectAuth() error leaked a secret value: %q", got)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		for i := 0; i+len(substr) <= len(s); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()
}

// --- S4 engine mini-wave item 4: OAuth2 extra params -------------------------

// TestBuildOAuth2ClientCredentialsExtraParamsWiredIntoTokenRequest proves
// AuthSpec.ExtraParams flows into connsdk.OAuth2ClientCredentials.ExtraParams
// (auth0's audience form param — connsdk already has an ExtraParams field;
// the gap was purely that AuthSpec had nothing to populate it from).
func TestBuildOAuth2ClientCredentialsExtraParamsWiredIntoTokenRequest(t *testing.T) {
	spec := AuthSpec{
		Mode:         "oauth2_client_credentials",
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ config.client_id }}",
		ClientSecret: "{{ secrets.client_secret }}",
		ExtraParams:  map[string]string{"audience": "{{ config.audience }}"},
	}
	vars := authVars(cfgWith(
		map[string]string{"token_url": "https://example.invalid/token", "client_id": "cid", "audience": "https://api.example.com/"},
		map[string]string{"client_secret": "csecret"},
	))

	auth, err := buildOAuth2ClientCredentials(spec, vars)
	if err != nil {
		t.Fatalf("buildOAuth2ClientCredentials() error = %v", err)
	}
	oauth2, ok := auth.(*connsdk.OAuth2ClientCredentials)
	if !ok {
		t.Fatalf("buildOAuth2ClientCredentials() returned %T, want *connsdk.OAuth2ClientCredentials", auth)
	}
	if got := oauth2.ExtraParams.Get("audience"); got != "https://api.example.com/" {
		t.Fatalf("ExtraParams[audience] = %q, want %q", got, "https://api.example.com/")
	}
}

// TestBuildOAuth2ClientCredentialsExtraParamsTemplatedFromConfig proves an
// extra_params value can be a DERIVED template (not just a bare config
// reference), e.g. auth0's own audience-derived-from-base_url convention.
func TestBuildOAuth2ClientCredentialsExtraParamsTemplatedFromConfig(t *testing.T) {
	spec := AuthSpec{
		Mode:         "oauth2_client_credentials",
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ config.client_id }}",
		ClientSecret: "{{ secrets.client_secret }}",
		ExtraParams:  map[string]string{"audience": "{{ config.base_url }}/api/v2/"},
	}
	vars := authVars(cfgWith(
		map[string]string{"token_url": "https://example.invalid/token", "client_id": "cid", "base_url": "https://acme.auth0.com"},
		map[string]string{"client_secret": "csecret"},
	))

	auth, err := buildOAuth2ClientCredentials(spec, vars)
	if err != nil {
		t.Fatalf("buildOAuth2ClientCredentials() error = %v", err)
	}
	oauth2 := auth.(*connsdk.OAuth2ClientCredentials)
	if got := oauth2.ExtraParams.Get("audience"); got != "https://acme.auth0.com/api/v2/" {
		t.Fatalf("ExtraParams[audience] = %q, want %q", got, "https://acme.auth0.com/api/v2/")
	}
}

// TestBuildOAuth2ClientCredentialsExtraParamsUnresolvedKeyErrors proves an
// extra_params value referencing an undeclared/absent config key hard-errors
// exactly like ClientID/ClientSecret do — never silently dropped.
func TestBuildOAuth2ClientCredentialsExtraParamsUnresolvedKeyErrors(t *testing.T) {
	spec := AuthSpec{
		Mode:         "oauth2_client_credentials",
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ config.client_id }}",
		ClientSecret: "{{ secrets.client_secret }}",
		ExtraParams:  map[string]string{"audience": "{{ config.missing_key }}"},
	}
	vars := authVars(cfgWith(
		map[string]string{"token_url": "https://example.invalid/token", "client_id": "cid"},
		map[string]string{"client_secret": "csecret"},
	))

	_, err := buildOAuth2ClientCredentials(spec, vars)
	if err == nil {
		t.Fatalf("buildOAuth2ClientCredentials() error = nil, want error for unresolved extra_params key")
	}
}

// TestReadOAuth2ClientCredentialsExtraParamsSentOnTokenRequest is the full
// integration proof: a real token-exchange round trip against an httptest
// token endpoint, asserting the form-encoded request body carries the
// resolved audience value.
func TestReadOAuth2ClientCredentialsExtraParamsSentOnTokenRequest(t *testing.T) {
	var gotAudience string
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		gotAudience = r.PostForm.Get("audience")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok-abc","token_type":"bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	spec := AuthSpec{
		Mode:         "oauth2_client_credentials",
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ config.client_id }}",
		ClientSecret: "{{ secrets.client_secret }}",
		ExtraParams:  map[string]string{"audience": "{{ config.audience }}"},
	}
	cfg := cfgWith(
		map[string]string{"token_url": tokenSrv.URL, "client_id": "cid", "audience": "https://api.example.com/"},
		map[string]string{"client_secret": "csecret"},
	)

	auth, err := selectAuth(context.Background(), cfg, []AuthSpec{spec}, nil)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
	applyToRequest(t, auth, req)

	if gotAudience != "https://api.example.com/" {
		t.Fatalf("token request audience form param = %q, want %q", gotAudience, "https://api.example.com/")
	}
}
