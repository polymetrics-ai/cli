package chift_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	chifthooks "polymetrics.ai/internal/connectors/hooks/chift"
)

func newRuntimeConfig(baseURL string, secrets map[string]string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": baseURL},
		Secrets: secrets,
	}
}

// TestAuthenticator_MintsSessionTokenAndSetsBearer asserts the JSON POST to
// /token carries clientId/clientSecret/accountId and that the resulting
// Authenticator sets a Bearer header from the response's access_token.
func TestAuthenticator_MintsSessionTokenAndSetsBearer(t *testing.T) {
	var gotPath, gotMethod, gotContentType string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		defer func() { _ = r.Body.Close() }()
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok_fixture_abc","expires_in":3600}`))
	}))
	defer srv.Close()

	h := chifthooks.New()
	cfg := newRuntimeConfig(srv.URL, map[string]string{
		"client_id":     "cid",
		"client_secret": "csecret",
		"account_id":    "acct_1",
	})
	authenticator, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.invalid/consumers", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := authenticator.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("token exchange method = %q, want POST", gotMethod)
	}
	if gotPath != "/token" {
		t.Fatalf("token exchange path = %q, want /token", gotPath)
	}
	if gotContentType != "application/json" {
		t.Fatalf("token exchange Content-Type = %q, want application/json", gotContentType)
	}
	if gotBody["clientId"] != "cid" || gotBody["clientSecret"] != "csecret" || gotBody["accountId"] != "acct_1" {
		t.Fatalf("token request body = %+v, want clientId/clientSecret/accountId set", gotBody)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer tok_fixture_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_fixture_abc", got)
	}
}

// TestAuthenticator_MissingSecretsErrors asserts a clear error, not a silent
// unauthenticated request, when a required secret is absent.
func TestAuthenticator_MissingSecretsErrors(t *testing.T) {
	h := chifthooks.New()
	cfg := newRuntimeConfig("https://api.chift.eu", map[string]string{"client_id": "cid"})
	_, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{})
	if err == nil {
		t.Fatal("expected error for missing client_secret/account_id")
	}
}

// TestConnectorNameAndRegistration asserts the hook set self-registers under
// "chift" and reports the matching ConnectorName.
func TestConnectorNameAndRegistration(t *testing.T) {
	h := chifthooks.New()
	if h.ConnectorName() != "chift" {
		t.Fatalf("ConnectorName() = %q, want chift", h.ConnectorName())
	}
	if hooks := engine.HooksFor("chift"); hooks == nil {
		t.Fatal("engine.HooksFor(\"chift\") returned nil; hook did not self-register")
	}
}
