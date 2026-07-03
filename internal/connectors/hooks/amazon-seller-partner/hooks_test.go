package amazonsellerpartner_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	amazonsellerpartnerhooks "polymetrics.ai/internal/connectors/hooks/amazon-seller-partner"
)

func newRuntimeConfig(tokenURL string, secrets map[string]string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config:  map[string]string{"lwa_token_url": tokenURL},
		Secrets: secrets,
	}
}

// TestAuthenticator_ExchangesRefreshTokenAndSetsHeader is the red-first test:
// it asserts the LWA refresh_token grant form body and that the returned
// Authenticator sets x-amz-access-token (never Authorization).
func TestAuthenticator_ExchangesRefreshTokenAndSetsHeader(t *testing.T) {
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

	h := amazonsellerpartnerhooks.New()
	cfg := newRuntimeConfig(srv.URL+"/auth/o2/token", map[string]string{
		"lwa_app_id":        "amzn1.application-oa2-client.abc",
		"lwa_client_secret": "shhh",
		"refresh_token":     "Atzr|refresh-xyz",
	})

	authenticator, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{Mode: "custom", Hook: "amazon-seller-partner"})
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
		t.Fatalf("client_id = %q, want the configured lwa_app_id", sawClientID)
	}
	if sawClientSecret != "shhh" {
		t.Fatalf("client_secret = %q, want the configured lwa_client_secret", sawClientSecret)
	}

	req, err := http.NewRequest(http.MethodGet, "https://sellingpartnerapi-na.amazon.com/orders/v0/orders", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := authenticator.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := req.Header.Get("x-amz-access-token"); got != wantAccessToken {
		t.Fatalf("x-amz-access-token = %q, want %q", got, wantAccessToken)
	}
	if got := req.Header.Get("Authorization"); got != "" {
		t.Fatalf("Authorization header = %q, want empty (SP-API uses x-amz-access-token, not Bearer)", got)
	}
}

func TestAuthenticator_MissingAppIDErrors(t *testing.T) {
	h := amazonsellerpartnerhooks.New()
	cfg := newRuntimeConfig("https://example.com/token", map[string]string{
		"lwa_client_secret": "secret",
		"refresh_token":     "rt",
	})
	if _, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{}); err == nil {
		t.Fatal("Authenticator() with missing lwa_app_id should error")
	}
}

func TestAuthenticator_MissingClientSecretErrors(t *testing.T) {
	h := amazonsellerpartnerhooks.New()
	cfg := newRuntimeConfig("https://example.com/token", map[string]string{
		"lwa_app_id":    "id",
		"refresh_token": "rt",
	})
	if _, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{}); err == nil {
		t.Fatal("Authenticator() with missing lwa_client_secret should error")
	}
}

func TestAuthenticator_MissingRefreshTokenErrors(t *testing.T) {
	h := amazonsellerpartnerhooks.New()
	cfg := newRuntimeConfig("https://example.com/token", map[string]string{
		"lwa_app_id":        "id",
		"lwa_client_secret": "secret",
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

	h := amazonsellerpartnerhooks.New()
	cfg := newRuntimeConfig(srv.URL, map[string]string{
		"lwa_app_id":        "id",
		"lwa_client_secret": "secret",
		"refresh_token":     "bad",
	})
	if _, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{}); err == nil {
		t.Fatal("Authenticator() with a non-2xx token endpoint response should error")
	}
}

func TestAuthenticator_HonorsContextCancellation(t *testing.T) {
	h := amazonsellerpartnerhooks.New()
	cfg := newRuntimeConfig("https://example.com/token", map[string]string{
		"lwa_app_id":        "id",
		"lwa_client_secret": "secret",
		"refresh_token":     "rt",
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := h.Authenticator(ctx, cfg, engine.AuthSpec{}); err == nil {
		t.Fatal("Authenticator() with a cancelled context should error")
	}
}

func TestConnectorName(t *testing.T) {
	h := amazonsellerpartnerhooks.New()
	if h.ConnectorName() != "amazon-seller-partner" {
		t.Fatalf("ConnectorName() = %q, want amazon-seller-partner", h.ConnectorName())
	}
}
