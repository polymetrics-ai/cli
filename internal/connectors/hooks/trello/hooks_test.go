package trello

import (
	"context"
	"net/http"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestAuthenticatorAddsKeyAndTokenQuery(t *testing.T) {
	auth, err := New().Authenticator(context.Background(), connectors.RuntimeConfig{Secrets: map[string]string{"key": "fixture-key", "token": "fixture-token"}}, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, "https://api.trello.test/1/cards/abc?fields=name", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	query := req.URL.Query()
	if query.Get("key") != "fixture-key" {
		t.Fatalf("key query = %q", query.Get("key"))
	}
	if query.Get("token") != "fixture-token" {
		t.Fatalf("token query = %q", query.Get("token"))
	}
	if query.Get("fields") != "name" {
		t.Fatalf("fields query = %q", query.Get("fields"))
	}
}

func TestAuthenticatorRequiresSecrets(t *testing.T) {
	for _, secrets := range []map[string]string{{"token": "fixture-token"}, {"key": "fixture-key"}} {
		if _, err := New().Authenticator(context.Background(), connectors.RuntimeConfig{Secrets: secrets}, engine.AuthSpec{}); err == nil {
			t.Fatalf("Authenticator(%v) succeeded, want error", secrets)
		}
	}
}
