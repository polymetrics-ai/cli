package uptick

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestHooksRegistered(t *testing.T) {
	h := engine.HooksFor("uptick")
	if h == nil {
		t.Fatal("registered hooks = nil")
	}
	if h.ConnectorName() != "uptick" {
		t.Fatalf("ConnectorName() = %q", h.ConnectorName())
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("hooks do not implement AuthHook")
	}
}

func TestAuthenticatorBuildsFromTemplates(t *testing.T) {
	auth, err := (&Hooks{}).Authenticator(context.Background(), connectors.RuntimeConfig{
		Config: map[string]string{
			"token_url": "https://auth.example.test/token",
			"username":  "user@example.test",
		},
		Secrets: map[string]string{
			"client_id":     "client-id",
			"client_secret": "client-secret",
			"password":      "password",
		},
	}, engine.AuthSpec{
		TokenURL:     "{{ config.token_url }}",
		ClientID:     "{{ secrets.client_id }}",
		ClientSecret: "{{ secrets.client_secret }}",
		Username:     "{{ config.username }}",
		Password:     "{{ secrets.password }}",
	})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	if auth == nil {
		t.Fatal("Authenticator returned nil")
	}
}
