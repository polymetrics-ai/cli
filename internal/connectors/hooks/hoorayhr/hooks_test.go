package hoorayhr

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestHooksRegistered(t *testing.T) {
	h := engine.HooksFor("hoorayhr")
	if h == nil {
		t.Fatal("registered hooks = nil")
	}
	if h.ConnectorName() != "hoorayhr" {
		t.Fatalf("ConnectorName() = %q", h.ConnectorName())
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("hooks do not implement AuthHook")
	}
}

func TestAuthenticatorBuildsFromConfig(t *testing.T) {
	auth, err := (&Hooks{}).Authenticator(context.Background(), connectors.RuntimeConfig{
		Config:  map[string]string{"hoorayhrusername": "user@example.test"},
		Secrets: map[string]string{"hoorayhrpassword": "password"},
	}, engine.AuthSpec{})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	if auth == nil {
		t.Fatal("Authenticator returned nil")
	}
}

func TestAuthenticatorRequiresUsernameAndPassword(t *testing.T) {
	if _, err := (&Hooks{}).Authenticator(context.Background(), connectors.RuntimeConfig{}, engine.AuthSpec{}); err == nil {
		t.Fatal("Authenticator without config/secrets: want error, got nil")
	}
}
