package coordination

import (
	"context"
	"net/http"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/connsdk"
)

func testReservationPolicy(t *testing.T, policyID, scope string, limit int) connsdk.ReservationPolicy {
	t.Helper()
	budget := fixedRequestBudget(limit, 60)
	fingerprint, err := RateBudgetPolicyFingerprint("paced", policyID, []connsdk.RateLimitBudget{budget})
	if err != nil {
		t.Fatalf("RateBudgetPolicyFingerprint: %v", err)
	}
	return connsdk.ReservationPolicy{
		Key:     connsdk.ReservationKey{PolicyFingerprint: fingerprint, Scope: scope},
		Budgets: []connsdk.RateLimitBudget{budget},
	}
}

func TestRateBudgetLeaseTTLFreesConcurrencyWithoutDroppingLateObservation(t *testing.T) {
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)}
	coordinator := NewRateBudgetCoordinator(clock, RateBudgetCoordinatorOptions{MaxInFlight: 1, LeaseTTL: time.Second})
	batch := connsdk.ReservationBatch{Policies: []connsdk.ReservationPolicy{testReservationPolicy(t, "core", "opaque-scope-a", 10)}}

	first, err := coordinator.Decide(context.Background(), batch)
	if err != nil || !first.Granted {
		t.Fatalf("first decision = %+v, %v", first, err)
	}

	clock.now = clock.now.Add(3 * time.Second)
	second, err := coordinator.Decide(context.Background(), batch)
	if err != nil || !second.Granted {
		t.Fatalf("decision after lease TTL = %+v, %v; want freed concurrency", second, err)
	}

	if err := coordinator.Finish(context.Background(), first.Lease, connsdk.CompletionObservation{
		Attempted:    true,
		Status:       http.StatusTooManyRequests,
		HasReset:     true,
		ResetAt:      clock.now.Add(time.Minute),
		HasRemaining: true,
		Remaining:    0,
	}); err != nil {
		t.Fatalf("late finish: %v", err)
	}

	decision, err := coordinator.Decide(context.Background(), batch)
	if err != nil {
		t.Fatalf("decision after late observation: %v", err)
	}
	if decision.Granted || decision.Wait != time.Minute {
		t.Fatalf("decision after late observation = %+v, want reset-window refusal", decision)
	}
}
