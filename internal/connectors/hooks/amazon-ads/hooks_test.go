package amazonads_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	amazonadshooks "polymetrics.ai/internal/connectors/hooks/amazon-ads"
	"polymetrics.ai/internal/credential"
)

func newRuntimeConfig(tokenURL, profileID string, secrets map[string]string) connectors.RuntimeConfig {
	cfg := map[string]string{"token_url": tokenURL}
	if profileID != "" {
		cfg["profile_id"] = profileID
	}
	return connectors.RuntimeConfig{Config: cfg, Secrets: secrets}
}

// TestAuthenticator_ExchangesRefreshTokenAndSetsBearer is the red-first test:
// it asserts the LWA refresh_token grant form body and that the returned
// Authenticator sets Authorization: Bearer <access_token> (not a custom
// header — unlike amazon-seller-partner's x-amz-access-token).
func TestAuthenticator_ExchangesRefreshTokenAndSetsBearer(t *testing.T) {
	var sawGrantType, sawRefresh, sawClientID, sawClientSecret string
	const wantAccessToken = "Atza|access-fixture-123"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/o2/token" {
			http.NotFound(w, r)
			return
		}
		_ = r.ParseForm()
		sawGrantType = r.PostForm.Get("grant_type")
		sawRefresh = r.PostForm.Get("refresh_token")
		sawClientID = r.PostForm.Get("client_id")
		sawClientSecret = r.PostForm.Get("client_secret")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"` + wantAccessToken + `","token_type":"bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	h := amazonadshooks.New()
	cfg := newRuntimeConfig(srv.URL+"/auth/o2/token", "987654321", map[string]string{
		"client_id":     "amzn1.application-oa2-client.abc",
		"client_secret": "shhh",
		"refresh_token": "Atzr|refresh-xyz",
	})

	authenticator, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{Mode: "custom", Hook: "amazon-ads"})
	if err != nil {
		t.Fatalf("Authenticator() error = %v", err)
	}
	if authenticator == nil {
		t.Fatal("Authenticator() = nil, want a non-nil connsdk.Authenticator")
	}

	if sawGrantType != "refresh_token" {
		t.Fatalf("grant_type = %q, want refresh_token", sawGrantType)
	}
	if sawRefresh != "Atzr|refresh-xyz" {
		t.Fatalf("refresh_token = %q, want the configured refresh token", sawRefresh)
	}
	if sawClientID != "amzn1.application-oa2-client.abc" {
		t.Fatalf("client_id = %q, want the configured client_id", sawClientID)
	}
	if sawClientSecret != "shhh" {
		t.Fatalf("client_secret = %q, want the configured client_secret", sawClientSecret)
	}

	req, err := http.NewRequest(http.MethodGet, "https://advertising-api.amazon.com/v2/sp/campaigns", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := authenticator.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer "+wantAccessToken {
		t.Fatalf("Authorization = %q, want Bearer %s", got, wantAccessToken)
	}
	if got := req.Header.Get("Amazon-Advertising-API-Scope"); got != "987654321" {
		t.Fatalf("Amazon-Advertising-API-Scope = %q, want 987654321", got)
	}
}

// TestAuthenticator_ProfilesEndpointHasNoScopeHeader verifies the returned
// Authenticator omits Amazon-Advertising-API-Scope for the unscoped profiles
// endpoint (matches legacy's streamEndpoint.scoped == false for profiles).
func TestAuthenticator_ProfilesEndpointHasNoScopeHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"atok","expires_in":3600}`))
	}))
	defer srv.Close()

	h := amazonadshooks.New()
	cfg := newRuntimeConfig(srv.URL, "987654321", map[string]string{
		"client_id":     "cid",
		"client_secret": "csec",
		"refresh_token": "rtok",
	})
	authenticator, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator() error = %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://advertising-api.amazon.com/v2/profiles", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := authenticator.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := req.Header.Get("Amazon-Advertising-API-Scope"); got != "" {
		t.Fatalf("Amazon-Advertising-API-Scope = %q, want empty for the unscoped profiles endpoint", got)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer atok" {
		t.Fatalf("Authorization = %q, want Bearer atok", got)
	}
}

// TestAuthenticator_ScopedEndpointMissingProfileIDErrors verifies a
// profile-scoped stream request errors when profile_id is unset, matching
// legacy's requester(cfg, scoped=true) hard-error gate rather than silently
// sending an unscoped request.
func TestAuthenticator_ScopedEndpointMissingProfileIDErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"atok","expires_in":3600}`))
	}))
	defer srv.Close()

	h := amazonadshooks.New()
	cfg := newRuntimeConfig(srv.URL, "", map[string]string{
		"client_id":     "cid",
		"client_secret": "csec",
		"refresh_token": "rtok",
	})
	authenticator, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator() error = %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://advertising-api.amazon.com/v2/sp/campaigns", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := authenticator.Apply(context.Background(), req); err == nil {
		t.Fatal("Apply() on a profile-scoped request with no profile_id should error")
	}
}

func TestAuthenticator_MissingClientIDErrors(t *testing.T) {
	h := amazonadshooks.New()
	cfg := newRuntimeConfig("https://example.com/token", "1", map[string]string{
		"client_secret": "secret",
		"refresh_token": "rt",
	})
	if _, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{}); err == nil {
		t.Fatal("Authenticator() with missing client_id should error")
	}
}

func TestAuthenticator_MissingClientSecretErrors(t *testing.T) {
	h := amazonadshooks.New()
	cfg := newRuntimeConfig("https://example.com/token", "1", map[string]string{
		"client_id":     "id",
		"refresh_token": "rt",
	})
	if _, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{}); err == nil {
		t.Fatal("Authenticator() with missing client_secret should error")
	}
}

func TestAuthenticator_MissingRefreshTokenErrors(t *testing.T) {
	h := amazonadshooks.New()
	cfg := newRuntimeConfig("https://example.com/token", "1", map[string]string{
		"client_id":     "id",
		"client_secret": "secret",
	})
	if _, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{}); err == nil {
		t.Fatal("Authenticator() with missing refresh_token should error")
	}
}

func TestAuthenticator_TokenEndpointErrorStatusErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer srv.Close()

	h := amazonadshooks.New()
	cfg := newRuntimeConfig(srv.URL, "1", map[string]string{
		"client_id":     "id",
		"client_secret": "secret",
		"refresh_token": "bad",
	})
	if _, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{}); err == nil {
		t.Fatal("Authenticator() with a non-2xx token endpoint response should error")
	}
}

func TestAuthenticator_HonorsContextCancellation(t *testing.T) {
	h := amazonadshooks.New()
	cfg := newRuntimeConfig("https://example.com/token", "1", map[string]string{
		"client_id":     "id",
		"client_secret": "secret",
		"refresh_token": "rt",
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := h.Authenticator(ctx, cfg, engine.AuthSpec{}); err == nil {
		t.Fatal("Authenticator() with a cancelled context should error")
	}
}

func TestAuthenticatorPreservesCredentialFormBytes(t *testing.T) {
	base := " " + strings.Repeat("r", 8192) + " "
	clientID := " " + strings.Repeat("i", 1024) + " "
	clientSecret := " " + strings.Repeat("s", 1024) + " "
	tests := []struct {
		name      string
		delimiter string
		retained  string
		escaped   string
	}{
		{name: "two terminal LFs", delimiter: "\n\n", retained: "\n", escaped: "%0A"},
		{name: "two terminal CRLFs", delimiter: "\r\n\r\n", retained: "\r\n", escaped: "%0D%0A"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshToken := credential.NormalizeStdin(base + tt.delimiter)
			wantRefreshToken := base + tt.retained
			assertAmazonFingerprint(t, "normalized refresh token", refreshToken, wantRefreshToken)

			var calls atomic.Int32
			var body string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				payload, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("ReadAll() error = %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				body = string(payload)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"access_token":"returned-access-token"}`))
			}))
			defer srv.Close()

			h := amazonadshooks.New()
			h.Client = srv.Client()
			_, err := h.Authenticator(context.Background(), newRuntimeConfig(srv.URL, "1", map[string]string{
				"client_id":     clientID,
				"client_secret": clientSecret,
				"refresh_token": refreshToken,
			}), engine.AuthSpec{})
			if err != nil {
				t.Fatalf("Authenticator() error = %v", err)
			}
			if calls.Load() != 1 {
				t.Fatalf("provider calls = %d, want 1", calls.Load())
			}
			wantBody := url.Values{
				"client_id":     []string{clientID},
				"client_secret": []string{clientSecret},
				"grant_type":    []string{"refresh_token"},
				"refresh_token": []string{wantRefreshToken},
			}.Encode()
			assertAmazonFingerprint(t, "form body", body, wantBody)
			if !strings.Contains(body, tt.escaped) || !strings.Contains(body, "+") {
				t.Fatal("form body did not percent-encode retained delimiter and spaces")
			}
		})
	}
}

func TestAuthenticatorRejectsTransportOnlyCredentialsBeforeProvider(t *testing.T) {
	for _, field := range []string{"client_id", "client_secret", "refresh_token"} {
		for _, tt := range []struct {
			name  string
			value string
		}{
			{name: "empty", value: ""},
			{name: "LF only", value: "\n"},
			{name: "CRLF only", value: "\r\n"},
		} {
			t.Run(field+"/"+tt.name, func(t *testing.T) {
				var calls atomic.Int32
				srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
					calls.Add(1)
				}))
				defer srv.Close()

				secrets := map[string]string{
					"client_id":     "client-id",
					"client_secret": "client-secret",
					"refresh_token": "refresh-token",
				}
				secrets[field] = tt.value
				h := amazonadshooks.New()
				h.Client = srv.Client()
				_, err := h.Authenticator(context.Background(), newRuntimeConfig(srv.URL, "1", secrets), engine.AuthSpec{})
				var empty *credential.EmptySecretError
				if !errors.As(err, &empty) {
					t.Fatalf("error type = %T, want EmptySecretError", err)
				}
				if calls.Load() != 0 {
					t.Fatalf("provider calls = %d, want 0", calls.Load())
				}
			})
		}
	}
}

func TestAuthenticatorRejectsHeaderControlAccessTokenWithoutLeakingIt(t *testing.T) {
	marker := "amazon-ads-" + strings.Repeat("a", 256)
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"` + marker + `\n"}`))
	}))
	defer srv.Close()

	h := amazonadshooks.New()
	h.Client = srv.Client()
	authenticator, err := h.Authenticator(context.Background(), newRuntimeConfig(srv.URL, "1", map[string]string{
		"client_id":     "client-id",
		"client_secret": "client-secret",
		"refresh_token": "refresh-token",
	}), engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator() error = %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, "https://advertising-api.amazon.com/v2/profiles", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	err = authenticator.Apply(context.Background(), req)
	var invalid *credential.InvalidSecretValueError
	if !errors.As(err, &invalid) {
		t.Fatalf("Apply() error type = %T, want InvalidSecretValueError", err)
	}
	if req.Header.Get("Authorization") != "" {
		t.Fatal("Apply() emitted an invalid authorization header")
	}
	if calls.Load() != 1 {
		t.Fatalf("token endpoint calls = %d, want 1", calls.Load())
	}
	if strings.Contains(err.Error(), marker) {
		t.Fatal("credential diagnostic exposed the canary")
	}
}

func assertAmazonFingerprint(t *testing.T, label, got, want string) {
	t.Helper()
	gotHash := sha256.Sum256([]byte(got))
	wantHash := sha256.Sum256([]byte(want))
	if len(got) != len(want) || gotHash != wantHash {
		t.Fatalf("%s length/hash = %d/%x, want %d/%x", label, len(got), gotHash, len(want), wantHash)
	}
}

func TestConnectorName(t *testing.T) {
	h := amazonadshooks.New()
	if h.ConnectorName() != "amazon-ads" {
		t.Fatalf("ConnectorName() = %q, want amazon-ads", h.ConnectorName())
	}
}
