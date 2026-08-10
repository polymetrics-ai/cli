package coordination

import (
	"context"
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
		Key: connsdk.ReservationKey{
			Connector: "paced",
			PolicyID:  policyID,
			Scope:     scope,
		},
		Fingerprint: fingerprint,
		Budgets:     []connsdk.RateLimitBudget{budget},
	}
}

func testReservationBatch(policies ...connsdk.ReservationPolicy) connsdk.ReservationBatch {
	return connsdk.ReservationBatch{Policies: policies}
}

func TestRateBudgetReserveBatchAllOrNothing(t *testing.T) {
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)}
	coordinator := NewRateBudgetCoordinator(clock, RateBudgetCoordinatorOptions{
		MaxInFlight: 1,
		LeaseTTL:    time.Minute,
	})
	ctx := context.Background()
	first := testReservationPolicy(t, "first", "opaque-scope-a", 1)
	blocked := testReservationPolicy(t, "blocked", "opaque-scope-a", 1)

	exhausted, err := coordinator.Decide(ctx, testReservationBatch(blocked))
	if err != nil {
		t.Fatalf("exhaust blocked policy: %v", err)
	}
	if !exhausted.Granted {
		t.Fatal("initial blocked-policy reservation was not granted")
	}
	if err := coordinator.Finish(ctx, exhausted.Lease, connsdk.CompletionObservation{Attempted: true}); err != nil {
		t.Fatalf("finish initial reservation: %v", err)
	}

	decision, err := coordinator.Decide(ctx, testReservationBatch(first, blocked))
	if err != nil {
		t.Fatalf("batch decision: %v", err)
	}
	if decision.Granted || decision.NotBefore.IsZero() {
		t.Fatalf("blocked batch = %+v, want typed non-grant with not_before", decision)
	}

	firstOnly, err := coordinator.Decide(ctx, testReservationBatch(first))
	if err != nil {
		t.Fatalf("first policy after blocked batch: %v", err)
	}
	if !firstOnly.Granted {
		t.Fatalf("blocked batch consumed first policy or lease: %+v", firstOnly)
	}
}
