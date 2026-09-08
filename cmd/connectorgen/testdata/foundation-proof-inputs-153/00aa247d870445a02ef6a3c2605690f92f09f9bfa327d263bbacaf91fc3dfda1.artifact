package engine

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/credential"
)

// recordingSecretStore is an in-memory connectors.SecretStore for asserting
// what the engine hands to the caller's credential store. The real
// implementation is internal/app's vault-backed one.
type recordingSecretStore struct {
	mu     sync.Mutex
	writes map[string]string
	err    error
}

func newRecordingSecretStore() *recordingSecretStore {
	return &recordingSecretStore{writes: map[string]string{}}
}

func (s *recordingSecretStore) PutSecret(_ context.Context, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	s.writes[key] = value
	return nil
}

func (s *recordingSecretStore) written(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.writes[key]
	return v, ok
}

func refreshTokenServer(t *testing.T, rotate bool) (*httptest.Server, func() int, func() []string) {
	t.Helper()
	var (
		mu     sync.Mutex
		calls  int
		grants []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		mu.Lock()
		calls++
		n := calls
		grants = append(grants, r.PostForm.Get("refresh_token"))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if rotate {
			_, _ = fmt.Fprintf(w, `{"access_token":"AT-%d","refresh_token":"rt-rotated-%d","expires_in":3600}`, n, n)
			return
		}
		_, _ = fmt.Fprintf(w, `{"access_token":"AT-%d","expires_in":3600}`, n)
	}))
	t.Cleanup(srv.Close)
	count := func() int {
		mu.Lock()
		defer mu.Unlock()
		return calls
	}
	presented := func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), grants...)
	}
	return srv, count, presented
}

// TestSelectAuthOAuth2RefreshTokenMode is the end-to-end proof for issue #3703:
// a bundle declaring the mode gets a working bearer credential from a real
// token exchange, and the grant carries every declared field.
func TestSelectAuthOAuth2RefreshTokenMode(t *testing.T) {
	var gotForm map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		gotForm = map[string]string{}
		for k := range r.PostForm {
			gotForm[k] = r.PostForm.Get(k)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"AT-user","token_type":"bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	specs := []AuthSpec{{
		Mode:         "oauth2_refresh_token",
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ config.client_id }}",
		ClientSecret: "{{ secrets.client_secret }}",
		RefreshToken: "{{ secrets.refresh_token }}",
		Scopes:       "identity read",
		ExtraParams:  map[string]string{"duration": "{{ config.duration }}"},
	}}
	cfg := cfgWith(
		map[string]string{"token_url": srv.URL, "client_id": "cid", "duration": "permanent"},
		map[string]string{"client_secret": "csecret", "refresh_token": "rt-original"},
	)

	auth, err := selectAuth(context.Background(), cfg, specs, nil)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
	applyToRequest(t, auth, req)

	if got := req.Header.Get("Authorization"); got != "Bearer AT-user" {
		t.Fatalf("Authorization = %q, want %q", got, "Bearer AT-user")
	}
	for k, want := range map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": "rt-original",
		"client_id":     "cid",
		"client_secret": "csecret",
		"scope":         "identity read",
		"duration":      "permanent",
	} {
		if gotForm[k] != want {
			t.Fatalf("token request %s = %q, want %q (full form %+v)", k, gotForm[k], want, gotForm)
		}
	}
}

// TestSelectAuthOAuth2RefreshTokenWhenGated proves the mode participates in the
// ordinary first-match-wins "when" selection like every other mode.
func TestSelectAuthOAuth2RefreshTokenWhenGated(t *testing.T) {
	srv, calls, _ := refreshTokenServer(t, false)

	specs := []AuthSpec{
		{Mode: "bearer", Token: "{{ secrets.token }}", When: "{{ config.auth_type == 'token' }}"},
		{
			Mode:         "oauth2_refresh_token",
			TokenURL:     "{{ config.token_url }}",
			RefreshToken: "{{ secrets.refresh_token }}",
			When:         "{{ config.auth_type == 'oauth' }}",
		},
	}

	t.Run("bearer branch does not exchange", func(t *testing.T) {
		cfg := cfgWith(
			map[string]string{"auth_type": "token", "token_url": srv.URL},
			map[string]string{"token": "static", "refresh_token": "rt"},
		)
		auth, err := selectAuth(context.Background(), cfg, specs, nil)
		if err != nil {
			t.Fatalf("selectAuth() error = %v", err)
		}
		req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
		applyToRequest(t, auth, req)
		if got := req.Header.Get("Authorization"); got != "Bearer static" {
			t.Fatalf("Authorization = %q", got)
		}
		if calls() != 0 {
			t.Fatalf("token endpoint called %d times on the bearer branch, want 0", calls())
		}
	})

	t.Run("oauth branch exchanges", func(t *testing.T) {
		cfg := cfgWith(
			map[string]string{"auth_type": "oauth", "token_url": srv.URL},
			map[string]string{"token": "static", "refresh_token": "rt"},
		)
		auth, err := selectAuth(context.Background(), cfg, specs, nil)
		if err != nil {
			t.Fatalf("selectAuth() error = %v", err)
		}
		req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
		applyToRequest(t, auth, req)
		if got := req.Header.Get("Authorization"); got != "Bearer AT-1" {
			t.Fatalf("Authorization = %q, want Bearer AT-1", got)
		}
	})
}

// TestBuildOAuth2RefreshTokenUnresolvedKeysError proves every templated field
// hard-errors on an unresolved config/secrets key, exactly as client_id and
// client_secret already do — never silently dropped.
func TestBuildOAuth2RefreshTokenUnresolvedKeysError(t *testing.T) {
	base := AuthSpec{
		Mode:         "oauth2_refresh_token",
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ config.client_id }}",
		ClientSecret: "{{ secrets.client_secret }}",
		RefreshToken: "{{ secrets.refresh_token }}",
	}
	config := map[string]string{"token_url": "https://example.invalid/token", "client_id": "cid"}
	secrets := map[string]string{"client_secret": "csecret", "refresh_token": "rt"}

	for _, field := range []string{"token_url", "client_id", "client_secret", "refresh_token", "scopes", "extra_params"} {
		t.Run(field, func(t *testing.T) {
			spec := base
			switch field {
			case "token_url":
				spec.TokenURL = "{{ config.missing }}"
			case "client_id":
				spec.ClientID = "{{ config.missing }}"
			case "client_secret":
				spec.ClientSecret = "{{ secrets.missing }}"
			case "refresh_token":
				spec.RefreshToken = "{{ secrets.missing }}"
			case "scopes":
				spec.Scopes = "{{ config.missing }}"
			case "extra_params":
				spec.ExtraParams = map[string]string{"audience": "{{ config.missing }}"}
			}
			cfg := cfgWith(config, secrets)
			_, err := buildOAuth2RefreshToken(cfg, spec, authVars(cfg))
			if err == nil {
				t.Fatalf("buildOAuth2RefreshToken() error = nil, want an error for unresolved %s", field)
			}
			if strings.Contains(err.Error(), "csecret") || strings.Contains(err.Error(), "rt") {
				t.Fatalf("error leaked a credential: %v", err)
			}
		})
	}
}

func TestBuildOAuth2RefreshTokenRejectsTransportOnlyDeclaredMaterial(t *testing.T) {
	srv, calls, _ := refreshTokenServer(t, false)
	base := AuthSpec{
		Mode:         "oauth2_refresh_token",
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ config.client_id }}",
		ClientSecret: "{{ secrets.client_secret }}",
		RefreshToken: "{{ secrets.refresh_token }}",
	}
	fields := []struct {
		name       string
		credential string
		wantField  string
	}{
		{name: "refresh token", credential: "refresh_token", wantField: "OAuth2 refresh token"},
		{name: "client ID", credential: "client_id", wantField: "OAuth2 client ID"},
		{name: "declared client secret", credential: "client_secret", wantField: "OAuth2 client secret"},
	}
	values := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "LF only", value: "\n"},
		{name: "CRLF only", value: "\r\n"},
	}

	for _, field := range fields {
		for _, value := range values {
			t.Run(field.name+" "+value.name, func(t *testing.T) {
				config := map[string]string{
					"token_url": srv.URL,
					"client_id": "oauth-client-id-redaction-canary",
				}
				secrets := map[string]string{
					"client_secret": "oauth-client-secret-redaction-canary",
					"refresh_token": "oauth-refresh-token-redaction-canary",
				}
				switch field.credential {
				case "client_id":
					config["client_id"] = value.value
				default:
					secrets[field.credential] = value.value
				}
				cfg := cfgWith(config, secrets)
				_, err := selectAuth(context.Background(), cfg, []AuthSpec{base}, nil)
				var empty *credential.EmptySecretError
				if !errors.As(err, &empty) {
					t.Fatalf("selectAuth() error type = %T, want EmptySecretError", err)
				}
				if empty.Field != field.wantField {
					t.Fatalf("empty credential field = %q, want %q", empty.Field, field.wantField)
				}
				for _, canary := range []string{"oauth-client-id-redaction-canary", "oauth-client-secret-redaction-canary", "oauth-refresh-token-redaction-canary"} {
					if strings.Contains(err.Error(), canary) {
						t.Fatal("credential validation error exposed a credential value")
					}
				}
				if got := calls(); got != 0 {
					t.Fatalf("token endpoint calls = %d, want 0", got)
				}
			})
		}
	}
}

func TestBuildOAuth2RefreshTokenOmitsUndeclaredPublicClientSecret(t *testing.T) {
	const (
		clientID     = "oauth-public-client-id-canary\n"
		refreshToken = "oauth-public-refresh-token-canary\n"
	)
	var gotForm map[string][]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm() error = %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		gotForm = r.PostForm
		_, _ = w.Write([]byte(`{"access_token":"public-client-access-token","expires_in":3600}`))
	}))
	defer srv.Close()

	spec := AuthSpec{
		Mode:         "oauth2_refresh_token",
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ config.client_id }}",
		RefreshToken: "{{ secrets.refresh_token }}",
	}
	cfg := cfgWith(
		map[string]string{"token_url": srv.URL, "client_id": clientID},
		map[string]string{"refresh_token": refreshToken},
	)
	auth, err := selectAuth(context.Background(), cfg, []AuthSpec{spec}, nil)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/public", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	applyToRequest(t, auth, req)
	if _, present := gotForm["client_secret"]; present {
		t.Fatal("public-client token form emitted an undeclared client secret")
	}
	for key, want := range map[string]string{"client_id": clientID, "refresh_token": refreshToken} {
		got := gotForm[key]
		if len(got) != 1 {
			t.Fatalf("token form %s count = %d, want 1", key, len(got))
		}
		if gotLength, wantLength := len(got[0]), len(want); gotLength != wantLength {
			t.Fatalf("token form %s length = %d, want %d", key, gotLength, wantLength)
		}
		if gotHash, wantHash := sha256.Sum256([]byte(got[0])), sha256.Sum256([]byte(want)); gotHash != wantHash {
			t.Fatalf("token form %s SHA-256 = %x, want %x", key, gotHash, wantHash)
		}
	}
}

// TestBuildOAuth2RefreshTokenRotationWiredToSecretStore proves the engine wires
// the declared store key to RuntimeConfig.SecretStore (issue #3705).
func TestBuildOAuth2RefreshTokenRotationWiredToSecretStore(t *testing.T) {
	srv, _, _ := refreshTokenServer(t, true)
	store := newRecordingSecretStore()

	spec := AuthSpec{
		Mode:                 "oauth2_refresh_token",
		TokenURL:             "{{ config.token_url }}",
		RefreshToken:         "{{ secrets.refresh_token }}",
		RefreshTokenStoreKey: "refresh_token",
	}
	cfg := connectors.RuntimeConfig{
		Config:      map[string]string{"token_url": srv.URL},
		Secrets:     map[string]string{"refresh_token": "rt-original"},
		SecretStore: store,
	}

	auth, err := selectAuth(context.Background(), cfg, []AuthSpec{spec}, nil)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
	applyToRequest(t, auth, req)

	got, ok := store.written("refresh_token")
	if !ok {
		t.Fatalf("nothing persisted under refresh_token; writes = %+v", store.writes)
	}
	if got != "rt-rotated-1" {
		t.Fatalf("persisted refresh token = %q, want rt-rotated-1", got)
	}
}

// TestBuildOAuth2RefreshTokenNoRotationWithoutDeclaredKey proves the engine
// never guesses where to write: a bundle that does not declare a store key
// persists nothing, even against a provider that rotates.
func TestBuildOAuth2RefreshTokenNoRotationWithoutDeclaredKey(t *testing.T) {
	srv, _, presented := refreshTokenServer(t, true)
	store := newRecordingSecretStore()

	spec := AuthSpec{
		Mode:         "oauth2_refresh_token",
		TokenURL:     "{{ config.token_url }}",
		RefreshToken: "{{ secrets.refresh_token }}",
	}
	cfg := connectors.RuntimeConfig{
		Config:      map[string]string{"token_url": srv.URL},
		Secrets:     map[string]string{"refresh_token": "rt-original"},
		SecretStore: store,
	}

	auth, err := selectAuth(context.Background(), cfg, []AuthSpec{spec}, nil)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
	applyToRequest(t, auth, req)

	store.mu.Lock()
	writes := len(store.writes)
	store.mu.Unlock()
	if writes != 0 {
		t.Fatalf("persisted %d secrets without a declared store key, want 0", writes)
	}
	// Rotation is still honoured in memory for this process.
	if got := presented(); len(got) != 1 || got[0] != "rt-original" {
		t.Fatalf("presented grants = %v", got)
	}
}

// TestBuildOAuth2RefreshTokenNoStoreConfiguredStillWorks proves a nil
// SecretStore (fixture tests, tests, any caller with no credential
// store) degrades to in-memory rotation, never to a plaintext write.
func TestBuildOAuth2RefreshTokenNoStoreConfiguredStillWorks(t *testing.T) {
	srv, _, _ := refreshTokenServer(t, true)

	spec := AuthSpec{
		Mode:                 "oauth2_refresh_token",
		TokenURL:             "{{ config.token_url }}",
		RefreshToken:         "{{ secrets.refresh_token }}",
		RefreshTokenStoreKey: "refresh_token",
	}
	cfg := cfgWith(map[string]string{"token_url": srv.URL}, map[string]string{"refresh_token": "rt-original"})

	auth, err := selectAuth(context.Background(), cfg, []AuthSpec{spec}, nil)
	if err != nil {
		t.Fatalf("selectAuth() error = %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
	applyToRequest(t, auth, req)
	if got := req.Header.Get("Authorization"); got != "Bearer AT-1" {
		t.Fatalf("Authorization = %q, want Bearer AT-1", got)
	}
}

// TestBuildOAuth2RefreshTokenRejectsInvalidStoreKey proves the declared store
// key is validated as an identifier before anything can be written with it.
func TestBuildOAuth2RefreshTokenRejectsInvalidStoreKey(t *testing.T) {
	for _, key := range []string{"../escape", "has space", "with/slash", "nul\x00byte"} {
		t.Run(key, func(t *testing.T) {
			spec := AuthSpec{
				Mode:                 "oauth2_refresh_token",
				TokenURL:             "{{ config.token_url }}",
				RefreshToken:         "{{ secrets.refresh_token }}",
				RefreshTokenStoreKey: key,
			}
			cfg := cfgWith(
				map[string]string{"token_url": "https://example.invalid/token"},
				map[string]string{"refresh_token": "rt"},
			)
			if _, err := buildOAuth2RefreshToken(cfg, spec, authVars(cfg)); err == nil {
				t.Fatalf("buildOAuth2RefreshToken() error = nil, want rejection of store key %q", key)
			}
		})
	}
}

// TestBuildOAuth2RefreshTokenReturnsRefresherAuthenticator proves the engine
// hands back an authenticator that Requester can refresh on a 401 — the mode is
// useless for unattended runs without it.
func TestBuildOAuth2RefreshTokenReturnsRefresherAuthenticator(t *testing.T) {
	spec := AuthSpec{
		Mode:         "oauth2_refresh_token",
		TokenURL:     "{{ config.token_url }}",
		RefreshToken: "{{ secrets.refresh_token }}",
	}
	cfg := cfgWith(
		map[string]string{"token_url": "https://example.invalid/token"},
		map[string]string{"refresh_token": "rt"},
	)
	auth, err := buildOAuth2RefreshToken(cfg, spec, authVars(cfg))
	if err != nil {
		t.Fatalf("buildOAuth2RefreshToken() error = %v", err)
	}
	if _, ok := auth.(connsdk.AuthRefresher); !ok {
		t.Fatalf("built authenticator %T does not implement connsdk.AuthRefresher", auth)
	}
}

// TestExistingAuthModesAreNotRefreshers is the additive/opt-in proof at the
// engine layer: no mode that predates this change gains 401-refresh behaviour.
func TestExistingAuthModesAreNotRefreshers(t *testing.T) {
	cfg := cfgWith(
		map[string]string{"token_url": "https://example.invalid/token", "client_id": "cid"},
		map[string]string{"token": "t", "key": "k", "client_secret": "cs"},
	)
	specs := map[string]AuthSpec{
		"none":                      {Mode: "none"},
		"bearer":                    {Mode: "bearer", Token: "{{ secrets.token }}"},
		"basic":                     {Mode: "basic", Username: "u", Password: "{{ secrets.token }}"},
		"api_key_header":            {Mode: "api_key_header", Header: "X-Key", Value: "{{ secrets.key }}"},
		"api_key_query":             {Mode: "api_key_query", Param: "key", Value: "{{ secrets.key }}"},
		"oauth2_client_credentials": {Mode: "oauth2_client_credentials", TokenURL: "{{ config.token_url }}", ClientID: "{{ config.client_id }}", ClientSecret: "{{ secrets.client_secret }}"},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			auth, err := selectAuth(context.Background(), cfg, []AuthSpec{spec}, nil)
			if err != nil {
				t.Fatalf("selectAuth(%s) error = %v", name, err)
			}
			if _, ok := auth.(connsdk.AuthRefresher); ok {
				t.Fatalf("mode %q unexpectedly implements connsdk.AuthRefresher; this change must be additive", name)
			}
		})
	}
}

// TestBundleLoadAcceptsOAuth2RefreshTokenAuthSpec proves a real bundle can
// DECLARE the mode: the meta-schema's auth item block is
// additionalProperties:false, and the loader strict-decodes streams.json
// independently of it, so both layers must know the two new keys or a
// connector could never adopt this.
func TestBundleLoadAcceptsOAuth2RefreshTokenAuthSpec(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"auth": [
				{
					"mode": "oauth2_refresh_token",
					"token_url": "https://example.invalid/token",
					"client_id": "{{ config.client_id }}",
					"client_secret": "{{ secrets.client_secret }}",
					"refresh_token": "{{ secrets.refresh_token }}",
					"refresh_token_store_key": "refresh_token",
					"scopes": "identity read",
					"extra_params": { "duration": "permanent" }
				}
			]
		},
		"streams": [
			{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json" }
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v, want an oauth2_refresh_token auth spec to load cleanly", err)
	}
	if len(b.HTTP.Auth) != 1 {
		t.Fatalf("HTTP.Auth = %+v, want one spec", b.HTTP.Auth)
	}
	spec := b.HTTP.Auth[0]
	if spec.Mode != "oauth2_refresh_token" {
		t.Fatalf("Mode = %q", spec.Mode)
	}
	if spec.RefreshToken != "{{ secrets.refresh_token }}" {
		t.Fatalf("RefreshToken = %q", spec.RefreshToken)
	}
	if spec.RefreshTokenStoreKey != "refresh_token" {
		t.Fatalf("RefreshTokenStoreKey = %q", spec.RefreshTokenStoreKey)
	}
}

// TestBundleLoadStillRejectsUnknownAuthKey pins that widening the auth block by
// exactly two keys did not reopen it to anything else.
func TestBundleLoadStillRejectsUnknownAuthKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"auth": [
				{ "mode": "oauth2_refresh_token", "refresh_tokens": "{{ secrets.refresh_token }}" }
			]
		},
		"streams": [
			{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json" }
		]
	}`)}

	if _, err := Load(fsys, "acme"); err == nil {
		t.Fatalf("Load: expected an error for the unknown auth key %q, got nil", "refresh_tokens")
	}
}

// TestResolveCheckAuthSpecCoversRefreshTokenFields closes the same gap F9
// closed for token_url/client_id/client_secret/scopes: a typo'd template in a
// new auth field must fail STATIC validation (connectorgen validate), not only
// at runtime on a real sync.
func TestResolveCheckAuthSpecCoversRefreshTokenFields(t *testing.T) {
	specKeys := map[string]bool{"token_url": true, "client_id": true, "client_secret": true, "refresh_token": true}

	t.Run("valid spec passes", func(t *testing.T) {
		spec := AuthSpec{
			Mode:                 "oauth2_refresh_token",
			TokenURL:             "{{ config.token_url }}",
			ClientID:             "{{ config.client_id }}",
			ClientSecret:         "{{ secrets.client_secret }}",
			RefreshToken:         "{{ secrets.refresh_token }}",
			RefreshTokenStoreKey: "refresh_token",
		}
		if err := ResolveCheckAuthSpec(spec, specKeys); err != nil {
			t.Fatalf("ResolveCheckAuthSpec() error = %v, want nil", err)
		}
	})

	// The secrets.* namespace is deliberately not statically checkable against
	// specKeys (checkNamespaceRef), and refresh_token is no exception — exactly
	// like client_secret. What IS checkable is a config.* reference and the
	// filter grammar, so those are what prove the field reaches ResolveCheck at
	// all rather than being skipped like it was before.
	for _, tc := range []struct {
		name string
		tmpl string
	}{
		{"undeclared config key", "{{ config.refersh_token }}"},
		{"unknown filter", "{{ secrets.refresh_token | bogus_filter }}"},
		{"unknown namespace", "{{ vault.refresh_token }}"},
	} {
		t.Run("rejected refresh_token template: "+tc.name, func(t *testing.T) {
			spec := AuthSpec{
				Mode:         "oauth2_refresh_token",
				TokenURL:     "{{ config.token_url }}",
				RefreshToken: tc.tmpl,
			}
			err := ResolveCheckAuthSpec(spec, specKeys)
			if err == nil {
				t.Fatalf("ResolveCheckAuthSpec(%q) error = nil, want rejection", tc.tmpl)
			}
			if !strings.Contains(err.Error(), "refresh_token") {
				t.Fatalf("ResolveCheckAuthSpec() error = %v, want it to name the refresh_token field", err)
			}
		})
	}

	t.Run("invalid refresh_token_store_key is rejected statically", func(t *testing.T) {
		for _, key := range []string{"../escape", "has space", "with/slash"} {
			spec := AuthSpec{
				Mode:                 "oauth2_refresh_token",
				TokenURL:             "{{ config.token_url }}",
				RefreshToken:         "{{ secrets.refresh_token }}",
				RefreshTokenStoreKey: key,
			}
			if err := ResolveCheckAuthSpec(spec, specKeys); err == nil {
				t.Fatalf("ResolveCheckAuthSpec() error = nil for store key %q, want rejection", key)
			}
		}
	})
}
