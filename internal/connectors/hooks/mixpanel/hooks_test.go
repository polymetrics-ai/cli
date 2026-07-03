package mixpanel

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

// authHeader applies auth to a fresh *http.Request and returns the resulting
// Authorization header, so a connsdk.Basic-built Authenticator (an opaque,
// unexported type) can be inspected without a type assertion.
func authHeader(t *testing.T, auth connsdk.Authenticator) string {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return req.Header.Get("Authorization")
}

func wantBasic(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

func TestAuthenticator_ConfigUsernameAndSecretPasswordWin(t *testing.T) {
	h := New().(*Hooks)
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"username": "config-user"},
		Secrets: map[string]string{"username_secret": "secret-user", "password": "secret-pass", "api_secret": "legacy-secret"},
	}
	auth, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	if got, want := authHeader(t, auth), wantBasic("config-user", "secret-pass"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}

func TestAuthenticator_FallsBackToSecretUsernameWhenConfigUnset(t *testing.T) {
	h := New().(*Hooks)
	cfg := connectors.RuntimeConfig{
		Secrets: map[string]string{"username_secret": "secret-user", "password": "secret-pass"},
	}
	auth, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	if got, want := authHeader(t, auth), wantBasic("secret-user", "secret-pass"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}

func TestAuthenticator_FallsBackToApiSecretWhenPasswordUnset(t *testing.T) {
	h := New().(*Hooks)
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"username": "config-user"},
		Secrets: map[string]string{"api_secret": "legacy-secret"},
	}
	auth, err := h.Authenticator(context.Background(), cfg, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	if got, want := authHeader(t, auth), wantBasic("config-user", "legacy-secret"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}

func TestAuthenticator_BothCredentialsMissingErrors(t *testing.T) {
	h := New().(*Hooks)
	if _, err := h.Authenticator(context.Background(), connectors.RuntimeConfig{}, engine.AuthSpec{}); err == nil {
		t.Fatal("Authenticator: want error when no username/password resolves, got nil")
	}
}

func TestMapRecord_EngageResolvesMultiSourceFallbacks(t *testing.T) {
	h := New().(*Hooks)
	raw := connsdk.Record{
		"$distinct_id": "fixture-1",
		"$properties": map[string]any{
			"$email":   "fixture1@example.com",
			"$created": "2026-01-01T00:00:00Z",
		},
	}
	mapped, keep, err := h.MapRecord("engage", raw, connsdk.Record{})
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if !keep {
		t.Fatal("MapRecord keep = false, want true")
	}
	if mapped["distinct_id"] != "fixture-1" || mapped["email"] != "fixture1@example.com" || mapped["created"] != "2026-01-01T00:00:00Z" {
		t.Fatalf("mapped = %+v", mapped)
	}
}

func TestMapRecord_EngageFallsBackToTopLevelFieldsWhenPropertiesAbsent(t *testing.T) {
	h := New().(*Hooks)
	raw := connsdk.Record{
		"distinct_id": "fixture-2",
		"email":       "top-level@example.com",
		"created":     "2026-01-02T00:00:00Z",
	}
	mapped, _, err := h.MapRecord("engage", raw, connsdk.Record{})
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if mapped["distinct_id"] != "fixture-2" || mapped["email"] != "top-level@example.com" || mapped["created"] != "2026-01-02T00:00:00Z" {
		t.Fatalf("mapped = %+v", mapped)
	}
}

func TestMapRecord_NonEngageStreamPassesThroughProjectedUnchanged(t *testing.T) {
	h := New().(*Hooks)
	projected := connsdk.Record{"id": 11, "name": "Fixture Cohort", "count": 5}
	mapped, keep, err := h.MapRecord("cohorts", connsdk.Record{}, projected)
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if !keep {
		t.Fatal("MapRecord keep = false, want true")
	}
	if mapped["id"] != 11 || mapped["name"] != "Fixture Cohort" {
		t.Fatalf("mapped = %+v, want projected unchanged", mapped)
	}
}
