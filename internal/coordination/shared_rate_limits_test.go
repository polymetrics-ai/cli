package coordination

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

func TestRateLimitRegistryStatusIsExplicitlyProcessLocal(t *testing.T) {
	registry := NewRateLimitRegistry(nil)
	status := registry.Status()
	if got, want := status.Mode, RateLimitCoordinationProcessLocal; got != want {
		t.Fatalf("local registry mode = %q, want %q", got, want)
	}
	if !strings.Contains(status.Message, "process-local") || strings.Contains(status.Message, "cross-process") {
		t.Fatalf("local registry message = %q, want an honest process-local statement", status.Message)
	}
	t.Logf("rate-limit coordination: %s", status.Message)
}

func TestSharedRateLimitRegistryRefusesWhenCoordinatorIsMissing(t *testing.T) {
	shared := NewSharedRateLimitRegistry(nil)
	limiter := shared.Limiter(testRateLimitKey(), []connsdk.RateLimitBudget{fixedRequestBudget(1, 60)})
	err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1})
	var unavailable *SharedRateLimitUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("shared admission error = %T %v, want SharedRateLimitUnavailableError", err, err)
	}
	if got, want := unavailable.Reason, SharedRateLimitCoordinatorNotConfigured; got != want {
		t.Fatalf("unavailable reason = %q, want %q", got, want)
	}
	if strings.Contains(err.Error(), string(testRateLimitKey().Scope)) {
		t.Fatalf("shared unavailable error exposed an opaque scope: %v", err)
	}
	t.Logf("require_shared result=refused reason=%s", unavailable.Reason)
}

func TestSharedRateLimitRedisKeyUsesOnlyOpaqueScope(t *testing.T) {
	identity, err := connectors.NewCoordinationIdentity([]byte("coordination-test-salt"), connectors.CredentialBinding{
		BindingID:      "binding-test-only",
		ProviderFamily: "provider-test-only",
		AuthProfile:    "profile-test-only",
	})
	if err != nil {
		t.Fatalf("NewCoordinationIdentity: %v", err)
	}
	scope, err := identity.RateScopeKey(connectors.RateLimitScope{
		PolicyID: "core",
		Kind:     connectors.RateScopeKindAccount,
		Subject:  "public-account-id",
	})
	if err != nil {
		t.Fatalf("RateScopeKey: %v", err)
	}
	key := RateLimitKey{
		Connector: "paced",
		PolicyID:  "core",
		Scope:     scope,
	}
	stored := sharedRateLimitRedisKey(key)
	for _, forbidden := range []string{"public-account-id", "binding-test-only", "provider-test-only", "profile-test-only"} {
		if strings.Contains(stored, forbidden) {
			t.Fatalf("shared registry key %q exposed protected material %q", stored, forbidden)
		}
	}
	if !strings.Contains(stored, string(key.Scope)) {
		t.Fatalf("shared registry key %q does not carry the opaque scope", stored)
	}
}
