package jamfpro_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	jamfprohooks "polymetrics.ai/internal/connectors/hooks/jamf-pro"
)

// TestAuthenticator_ExchangesBasicForBearerAndCaches mirrors legacy's
// TestReadPaginatesAndAuthenticates auth assertions: the token endpoint sees
// HTTP Basic credentials and no request body, the returned token is used as
// Authorization: Bearer on subsequent requests, and a second Apply call
// within the cached TTL does not re-hit the token endpoint.
func TestAuthenticator_ExchangesBasicForBearerAndCaches(t *testing.T) {
	var sawTokenBasicAuth, sawTokenMethod string
	var sawTokenBody []byte
	tokenCalls := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/token" {
			http.NotFound(w, r)
			return
		}
		tokenCalls++
		sawTokenMethod = r.Method
		sawTokenBasicAuth = r.Header.Get("Authorization")
		buf := make([]byte, 1)
		n, _ := r.Body.Read(buf)
		sawTokenBody = buf[:n]
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"jamf-token-xyz","expires":"2099-01-01T00:00:00.000Z"}`))
	}))
	defer srv.Close()

	h := jamfprohooks.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "apiuser"},
		Secrets: map[string]string{"password": "s3cret"},
	}

	authr, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	// Basic base64("apiuser:s3cret") == "YXBpdXNlcjpzM2NyZXQ="
	req1, _ := http.NewRequest(http.MethodGet, "https://example.invalid/v1/buildings", nil)
	if err := authr.Apply(context.Background(), req1); err != nil {
		t.Fatalf("Apply (1st): %v", err)
	}
	if sawTokenMethod != http.MethodPost {
		t.Fatalf("token request method = %q, want POST", sawTokenMethod)
	}
	if sawTokenBasicAuth != "Basic YXBpdXNlcjpzM2NyZXQ=" {
		t.Fatalf("token Authorization = %q, want Basic credentials", sawTokenBasicAuth)
	}
	if len(sawTokenBody) != 0 {
		t.Fatalf("token request body = %q, want empty (no request body)", sawTokenBody)
	}
	if got := req1.Header.Get("Authorization"); got != "Bearer jamf-token-xyz" {
		t.Fatalf("data Authorization = %q, want Bearer jamf-token-xyz", got)
	}

	// A second Apply call within the cached TTL must not re-hit the token
	// endpoint (mirrors legacy's tokenCalls != 1 assertion across pages).
	req2, _ := http.NewRequest(http.MethodGet, "https://example.invalid/v1/buildings", nil)
	if err := authr.Apply(context.Background(), req2); err != nil {
		t.Fatalf("Apply (2nd): %v", err)
	}
	if tokenCalls != 1 {
		t.Fatalf("token endpoint called %d times, want 1 (token cached/reused)", tokenCalls)
	}
	if got := req2.Header.Get("Authorization"); got != "Bearer jamf-token-xyz" {
		t.Fatalf("data Authorization (2nd) = %q, want Bearer jamf-token-xyz", got)
	}
}

func TestAuthenticator_RequiresBaseURLUsernamePassword(t *testing.T) {
	h := jamfprohooks.New()

	cases := []struct {
		name string
		cfg  connectors.RuntimeConfig
	}{
		{"missing base_url", connectors.RuntimeConfig{Config: map[string]string{"username": "u"}, Secrets: map[string]string{"password": "p"}}},
		{"missing username", connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://x.example"}, Secrets: map[string]string{"password": "p"}}},
		{"missing password", connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://x.example", "username": "u"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := h.Authenticator(context.Background(), tc.cfg, engine.AuthSpec{}); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestConnectorNameAndRegistration(t *testing.T) {
	h := jamfprohooks.New()
	if h.ConnectorName() != "jamf-pro" {
		t.Fatalf("ConnectorName() = %q, want jamf-pro", h.ConnectorName())
	}
	if hooks := engine.HooksFor("jamf-pro"); hooks == nil {
		t.Fatal("engine.HooksFor(\"jamf-pro\") = nil, want registered hooks (init side effect)")
	}
}
