package strava

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestHooksRegistered(t *testing.T) {
	h := engine.HooksFor("strava")
	if h == nil {
		t.Fatal("registered hooks = nil")
	}
	if h.ConnectorName() != "strava" {
		t.Fatalf("ConnectorName() = %q", h.ConnectorName())
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("hooks do not implement AuthHook")
	}
}

func TestAuthenticatorBuildsFromTemplates(t *testing.T) {
	auth, err := (&Hooks{}).Authenticator(context.Background(), connectors.RuntimeConfig{
		Config: map[string]string{"token_url": "https://auth.example.test/token"},
		Secrets: map[string]string{
			"client_id":     "client-id",
			"client_secret": "client-secret",
			"refresh_token": "refresh",
		},
	}, engine.AuthSpec{
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ secrets.client_id }}",
		ClientSecret: "{{ secrets.client_secret }}",
		Token:        "{{ secrets.refresh_token }}",
	})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	if auth == nil {
		t.Fatal("Authenticator returned nil")
	}
}
