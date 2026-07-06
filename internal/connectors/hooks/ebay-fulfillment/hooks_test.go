package ebayfulfillment

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestHooksRegistered(t *testing.T) {
	h := engine.HooksFor("ebay-fulfillment")
	if h == nil {
		t.Fatal("registered hooks = nil")
	}
	if h.ConnectorName() != "ebay-fulfillment" {
		t.Fatalf("ConnectorName() = %q", h.ConnectorName())
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("hooks do not implement AuthHook")
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("hooks do not implement StreamHook")
	}
}

func TestAuthenticatorBuildsFromTemplates(t *testing.T) {
	auth, err := (&Hooks{}).Authenticator(context.Background(), connectors.RuntimeConfig{
		Config:  map[string]string{"token_url": "https://auth.example.test/token"},
		Secrets: map[string]string{"refresh_token": "refresh"},
	}, engine.AuthSpec{
		TokenURL: "{{ config.token_url }}",
		Token:    "{{ secrets.refresh_token }}",
	})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	if auth == nil {
		t.Fatal("Authenticator returned nil")
	}
}
