package coordination

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

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

func TestSharedRateLimitRegistryPreservesCallerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := NewSharedRateLimitRegistry(nil).EnsureAvailable(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled shared availability error = %v, want context.Canceled", err)
	}
	var unavailable *SharedRateLimitUnavailableError
	if errors.As(err, &unavailable) {
		t.Fatalf("cancelled shared availability returned unavailable reason %q", unavailable.Reason)
	}
}

func TestSharedRateLimitObserveScriptArgsMatchScriptContract(t *testing.T) {
	specs, err := sharedRateLimitBudgetSpecs([]connsdk.RateLimitBudget{fixedRequestBudget(3, 60)})
	if err != nil {
		t.Fatalf("sharedRateLimitBudgetSpecs: %v", err)
	}
	observation := sharedRateLimitObservation{
		BlockedUntil:       1_786_068_045_000,
		Limit:              2,
		HasLimit:           true,
		Remaining:          1,
		HasRemaining:       true,
		ForceRemainingZero: true,
		Cost:               2,
		HasCost:            true,
	}
	args, err := sharedRateLimitObserveScriptArgs(specs, observation, 2*time.Minute)
	if err != nil {
		t.Fatalf("sharedRateLimitObserveScriptArgs: %v", err)
	}
	if got, want := len(args), 3; got != want {
		t.Fatalf("observe script arg count = %d, want %d", got, want)
	}
	encodedSpecs, ok := args[0].(string)
	if !ok {
		t.Fatalf("observe script specs argument = %T, want string", args[0])
	}
	var gotSpecs []sharedRateLimitBudget
	if err := json.Unmarshal([]byte(encodedSpecs), &gotSpecs); err != nil {
		t.Fatalf("decode observe script specs: %v", err)
	}
	if len(gotSpecs) != len(specs) || gotSpecs[0] != specs[0] {
		t.Fatalf("observe script specs = %+v, want %+v", gotSpecs, specs)
	}
	encodedObservation, ok := args[1].(string)
	if !ok {
		t.Fatalf("observe script observation argument = %T, want string", args[1])
	}
	var gotObservation sharedRateLimitObservation
	if err := json.Unmarshal([]byte(encodedObservation), &gotObservation); err != nil {
		t.Fatalf("decode observe script observation: %v", err)
	}
	if gotObservation != observation {
		t.Fatalf("observe script observation = %+v, want %+v", gotObservation, observation)
	}
	if got, ok := args[2].(int64); !ok || got != (2*time.Minute).Milliseconds() {
		t.Fatalf("observe script TTL argument = %T(%v), want milliseconds", args[2], args[2])
	}
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
