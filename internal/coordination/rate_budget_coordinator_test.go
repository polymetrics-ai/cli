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
		Key: connsdk.ReservationKey{
			PolicyFingerprint: fingerprint,
			Scope:             scope,
		},
		Budgets: []connsdk.RateLimitBudget{budget},
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

func TestRateBudgetFinishIsIdempotentAndReleasesLease(t *testing.T) {
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)}
	coordinator := NewRateBudgetCoordinator(clock, RateBudgetCoordinatorOptions{
		MaxInFlight: 1,
		LeaseTTL:    time.Minute,
	})
	batch := testReservationBatch(testReservationPolicy(t, "core", "opaque-scope-a", 3))
	first, err := coordinator.Decide(context.Background(), batch)
	if err != nil || !first.Granted || first.Lease == "" {
		t.Fatalf("first decision = %+v, %v", first, err)
	}
	blocked, err := coordinator.Decide(context.Background(), batch)
	if err != nil {
		t.Fatalf("concurrent decision: %v", err)
	}
	if blocked.Granted || blocked.NotBefore.IsZero() {
		t.Fatalf("concurrent decision = %+v, want lease refusal", blocked)
	}
	if err := coordinator.Finish(context.Background(), first.Lease, connsdk.CompletionObservation{Attempted: true}); err != nil {
		t.Fatalf("finish first lease: %v", err)
	}
	if err := coordinator.Finish(context.Background(), first.Lease, connsdk.CompletionObservation{Attempted: true}); err != nil {
		t.Fatalf("finish first lease twice: %v", err)
	}
	next, err := coordinator.Decide(context.Background(), batch)
	if err != nil || !next.Granted {
		t.Fatalf("decision after finish = %+v, %v", next, err)
	}
}

func TestRateBudgetRejectsMismatchedRegisteredFingerprint(t *testing.T) {
	coordinator := NewRateBudgetCoordinator(nil, RateBudgetCoordinatorOptions{LeaseTTL: time.Minute})
	policy := testReservationPolicy(t, "core", "opaque-scope-a", 3)
	initial, err := coordinator.Decide(context.Background(), testReservationBatch(policy))
	if err != nil || !initial.Granted {
		t.Fatalf("initial decision = %+v, %v", initial, err)
	}
	if err := coordinator.Finish(context.Background(), initial.Lease, connsdk.CompletionObservation{Attempted: true}); err != nil {
		t.Fatalf("finish initial decision: %v", err)
	}
	mismatched := policy
	mismatched.Budgets = []connsdk.RateLimitBudget{fixedRequestBudget(4, 60)}
	if _, err := coordinator.Decide(context.Background(), testReservationBatch(mismatched)); err == nil {
		t.Fatal("mismatched registered fingerprint was accepted")
	}
}

func TestRateBudgetLeaseTTLFreesConcurrencyWithoutDroppingLateObservation(t *testing.T) {
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)}
	coordinator := NewRateBudgetCoordinator(clock, RateBudgetCoordinatorOptions{
		MaxInFlight: 1,
		LeaseTTL:    time.Second,
	})
	policy := testReservationPolicy(t, "core", "opaque-scope-a", 10)
	batch := testReservationBatch(policy)
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
