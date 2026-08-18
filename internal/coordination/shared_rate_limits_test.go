package coordination

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

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

func TestSharedRateLimitReservePolicyErrorIsNotCoordinatorUnavailable(t *testing.T) {
	err := sharedRateLimitReserveError(errors.New("ERR shared rate-limit request cost exceeds declared capacity"))
	if !errors.Is(err, errSharedRateLimitPolicy) {
		t.Fatalf("shared reserve policy error = %v, want policy error", err)
	}
	var unavailable *SharedRateLimitUnavailableError
	if errors.As(err, &unavailable) {
		t.Fatalf("shared reserve policy error returned unavailable reason %q", unavailable.Reason)
	}
	if strings.Contains(err.Error(), "ERR") {
		t.Fatalf("shared reserve policy error exposed Redis detail %q", err)
	}
}

func TestSharedRateLimitBudgetSpecsRejectCostAboveCapacity(t *testing.T) {
	capacity, restore, cost := 1, 1.0, 2.0
	_, err := sharedRateLimitBudgetSpecs([]connsdk.RateLimitBudget{{
		Model:            connsdk.RateLimitBudgetTokenBucket,
		Dimension:        connsdk.RateLimitBudgetSustained,
		Unit:             connsdk.RateLimitBudgetPoints,
		Capacity:         &capacity,
		RestorePerSecond: &restore,
		Cost:             &connsdk.RateLimitCost{DefaultCost: &cost},
	}})
	if !errors.Is(err, errSharedRateLimitPolicy) {
		t.Fatalf("shared budget specs error = %v, want policy error", err)
	}
}

func TestSharedRateLimitWindowBoundaryAcceptsMaximum(t *testing.T) {
	maximum := int(maxSharedRateLimitWindowSeconds)
	specs, err := sharedRateLimitBudgetSpecs([]connsdk.RateLimitBudget{fixedRequestBudget(1, maximum)})
	if err != nil {
		t.Fatalf("maximum shared window specs: %v", err)
	}
	if got, want := specs[0].Window, maxSharedRateLimitWindowSeconds*int64(time.Second/time.Millisecond); got != want {
		t.Fatalf("maximum shared window milliseconds = %d, want %d", got, want)
	}
	ttl, err := sharedRateLimitTTL(specs)
	if err != nil {
		t.Fatalf("maximum shared window TTL: %v", err)
	}
	if got, want := ttl, time.Duration(maxSharedRateLimitWindowSeconds)*time.Second+time.Second; got != want {
		t.Fatalf("maximum shared window TTL = %s, want %s", got, want)
	}
}

func TestSharedRateLimitWindowBoundaryRejectsBadInputBeforeCoordinatorIO(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for coordinator probe: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	var connections atomic.Int32
	go func() {
		for {
			conn, acceptErr := listener.Accept()
			if acceptErr != nil {
				return
			}
			connections.Add(1)
			_ = conn.Close()
		}
	}()

	for _, tt := range []struct {
		name    string
		seconds int
		reason  SharedRateLimitWindowErrorReason
	}{
		{name: "negative", seconds: -1, reason: SharedRateLimitWindowNonPositive},
		{name: "one past maximum", seconds: int(maxSharedRateLimitWindowSeconds) + 1, reason: SharedRateLimitWindowTooLarge},
	} {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewSharedRateLimitRegistry(&Dragonfly{client: redis.NewClient(&redis.Options{Addr: listener.Addr().String()})}).Limiter(testRateLimitKey(), []connsdk.RateLimitBudget{fixedRequestBudget(1, tt.seconds)})
			err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1})
			var windowErr *SharedRateLimitWindowError
			if !errors.As(err, &windowErr) {
				t.Fatalf("shared window %d admission error = %T %v, want SharedRateLimitWindowError", tt.seconds, err, err)
			}
			if got, want := windowErr.Reason, tt.reason; got != want {
				t.Fatalf("shared window %d refusal reason = %q, want %q", tt.seconds, got, want)
			}
		})
	}
	if got := connections.Load(); got != 0 {
		t.Fatalf("invalid shared windows opened %d coordinator connections, want 0", got)
	}
}

func TestSharedRateLimitWindowBoundaryRejectsZeroAndDurationOverflow(t *testing.T) {
	for _, tt := range []struct {
		name    string
		seconds int
		reason  SharedRateLimitWindowErrorReason
	}{
		{name: "zero", seconds: 0, reason: SharedRateLimitWindowNonPositive},
		{name: "duration overflow", seconds: int(^uint(0) >> 1), reason: SharedRateLimitWindowTooLarge},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sharedRateLimitBudgetSpecs([]connsdk.RateLimitBudget{fixedRequestBudget(1, tt.seconds)})
			var windowErr *SharedRateLimitWindowError
			if !errors.As(err, &windowErr) {
				t.Fatalf("shared window %d specs error = %T %v, want SharedRateLimitWindowError", tt.seconds, err, err)
			}
			if got, want := windowErr.Reason, tt.reason; got != want {
				t.Fatalf("shared window %d refusal reason = %q, want %q", tt.seconds, got, want)
			}
		})
	}
}

func TestSharedRateLimitObserveScriptArgsMatchScriptContract(t *testing.T) {
	specs, err := sharedRateLimitBudgetSpecs([]connsdk.RateLimitBudget{fixedRequestBudget(3, 60)})
	if err != nil {
		t.Fatalf("sharedRateLimitBudgetSpecs: %v", err)
	}
	observation := sharedRateLimitObservation{
		BlockFor:           60_000,
		AbsoluteResetAt:    1_786_000_000_000,
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

func TestSharedRateLimitObservationUsesRelativeBlockDuration(t *testing.T) {
	observedAt := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	observation := sharedRateLimitObservationOf(connsdk.RateLimitObservation{
		Attempted:  true,
		Status:     http.StatusTooManyRequests,
		HasReset:   true,
		ResetAt:    observedAt.Add(time.Minute),
		ObservedAt: observedAt,
	})
	if got, want := observation.BlockFor, int64(time.Minute/time.Millisecond); got != want {
		t.Fatalf("shared block duration = %dms, want %dms", got, want)
	}
	encoded, err := json.Marshal(observation)
	if err != nil {
		t.Fatalf("marshal shared observation: %v", err)
	}
	if strings.Contains(string(encoded), "blocked_until") {
		t.Fatalf("shared observation carried a client-clock deadline: %s", encoded)
	}
}

func TestSharedRateLimitObservationPreservesAbsoluteReset(t *testing.T) {
	resetAt := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	observation := sharedRateLimitObservationOf(connsdk.RateLimitObservation{
		Attempted:       true,
		Status:          http.StatusTooManyRequests,
		HasReset:        true,
		ResetAt:         resetAt,
		ResetAtAbsolute: true,
		ObservedAt:      resetAt.Add(time.Minute),
	})
	if got, want := observation.AbsoluteResetAt, resetAt.UnixMilli(); got != want {
		t.Fatalf("shared absolute reset = %d, want %d", got, want)
	}
	if observation.BlockFor != 0 {
		t.Fatalf("shared absolute reset also carried a relative block = %dms", observation.BlockFor)
	}
	if !observation.relevant() {
		t.Fatal("shared absolute reset was not relevant")
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
