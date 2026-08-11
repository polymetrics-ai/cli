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
	return testReservationPolicyWithBudget(t, policyID, scope, fixedRequestBudget(limit, 60))
}

func testReservationPolicyWithBudget(t *testing.T, policyID, scope string, budget connsdk.RateLimitBudget) connsdk.ReservationPolicy {
	t.Helper()
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

func testRateBudgetLeaseCount(t *testing.T, coordinator *RateBudgetCoordinator) int {
	t.Helper()
	coordinator.registry.mu.Lock()
	defer coordinator.registry.mu.Unlock()
	return len(coordinator.registry.leases)
}

func newTestRateBudgetCoordinator(t *testing.T, clock RateLimitClock, options RateBudgetCoordinatorOptions) *RateBudgetCoordinator {
	t.Helper()
	coordinator, err := NewRateBudgetCoordinator(clock, options)
	if err != nil {
		t.Fatalf("NewRateBudgetCoordinator: %v", err)
	}
	return coordinator
}

func TestRateBudgetReserveBatchAllOrNothing(t *testing.T) {
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)}
	coordinator := newTestRateBudgetCoordinator(t, clock, RateBudgetCoordinatorOptions{
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
	coordinator := newTestRateBudgetCoordinator(t, clock, RateBudgetCoordinatorOptions{
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
	coordinator := newTestRateBudgetCoordinator(t, nil, RateBudgetCoordinatorOptions{LeaseTTL: time.Minute})
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
	coordinator := newTestRateBudgetCoordinator(t, clock, RateBudgetCoordinatorOptions{
		MaxInFlight:                  1,
		LeaseTTL:                     time.Second,
		CompletionObservationHorizon: 2 * time.Minute,
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

func TestRateBudgetObservationCleanupNeverDeletesActiveLease(t *testing.T) {
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)}
	coordinator := newTestRateBudgetCoordinator(t, clock, RateBudgetCoordinatorOptions{
		MaxInFlight:                  1,
		LeaseTTL:                     3 * time.Minute,
		CompletionObservationHorizon: 4 * time.Minute,
	})
	batch := testReservationBatch(testReservationPolicy(t, "core", "opaque-scope-a", 10))
	first, err := coordinator.Decide(context.Background(), batch)
	if err != nil || !first.Granted {
		t.Fatalf("first decision = %+v, %v", first, err)
	}
	coordinator.registry.mu.Lock()
	coordinator.registry.completionObservationHorizon = 2 * time.Minute
	coordinator.registry.mu.Unlock()
	clock.now = clock.now.Add(2 * time.Minute)
	blocked, err := coordinator.Decide(context.Background(), batch)
	if err != nil {
		t.Fatalf("decision at invalid cleanup horizon: %v", err)
	}
	if blocked.Granted || blocked.Wait != time.Minute {
		t.Fatalf("decision at invalid cleanup horizon = %+v, want one-minute concurrency refusal", blocked)
	}
	if got := testRateBudgetLeaseCount(t, coordinator); got != 1 {
		t.Fatalf("lease records at invalid cleanup horizon = %d, want active lease retained", got)
	}
}

func TestRateBudgetCoordinatorRejectsHorizonNotGreaterThanTTL(t *testing.T) {
	for _, test := range []struct {
		name    string
		options RateBudgetCoordinatorOptions
		wantErr bool
	}{
		{name: "defaulted horizon", options: RateBudgetCoordinatorOptions{LeaseTTL: 3 * time.Minute}, wantErr: true},
		{name: "shorter horizon", options: RateBudgetCoordinatorOptions{LeaseTTL: 3 * time.Minute, CompletionObservationHorizon: 2 * time.Minute}, wantErr: true},
		{name: "equal horizon", options: RateBudgetCoordinatorOptions{LeaseTTL: 3 * time.Minute, CompletionObservationHorizon: 3 * time.Minute}, wantErr: true},
		{name: "longer horizon", options: RateBudgetCoordinatorOptions{LeaseTTL: 3 * time.Minute, CompletionObservationHorizon: 4 * time.Minute}},
	} {
		t.Run(test.name, func(t *testing.T) {
			coordinator, err := NewRateBudgetCoordinator(nil, test.options)
			if test.wantErr {
				if err == nil || coordinator != nil {
					t.Fatalf("NewRateBudgetCoordinator = (%v, %v), want nil coordinator and error", coordinator, err)
				}
				return
			}
			if err != nil || coordinator == nil {
				t.Fatalf("NewRateBudgetCoordinator = (%v, %v), want configured coordinator", coordinator, err)
			}
			if got, want := coordinator.registry.leaseTTL, 3*time.Minute; got != want {
				t.Fatalf("lease ttl = %s, want %s", got, want)
			}
			if got, want := coordinator.registry.completionObservationHorizon, 4*time.Minute; got != want {
				t.Fatalf("completion observation horizon = %s, want %s", got, want)
			}
		})
	}
}

func TestRateBudgetCompletionObservationHorizonNormalizesAtOwner(t *testing.T) {
	for _, test := range []struct {
		name       string
		configured time.Duration
		want       time.Duration
	}{
		{name: "zero", want: defaultCompletionObservationHorizon},
		{name: "negative", configured: -time.Second, want: defaultCompletionObservationHorizon},
		{name: "explicit", configured: 3 * time.Minute, want: 3 * time.Minute},
	} {
		t.Run(test.name, func(t *testing.T) {
			coordinator := newTestRateBudgetCoordinator(t, nil, RateBudgetCoordinatorOptions{
				CompletionObservationHorizon: test.configured,
			})
			if got := coordinator.registry.completionObservationHorizon; got != test.want {
				t.Fatalf("completion observation horizon = %s, want %s", got, test.want)
			}
		})
	}
}

func TestRateBudgetCompletionObservationHorizonAppliesLateObservationBeforeBoundary(t *testing.T) {
	const horizon = 2 * time.Minute
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)}
	coordinator := newTestRateBudgetCoordinator(t, clock, RateBudgetCoordinatorOptions{
		MaxInFlight:                  1,
		LeaseTTL:                     time.Second,
		CompletionObservationHorizon: horizon,
	})
	batch := testReservationBatch(testReservationPolicy(t, "core", "opaque-scope-a", 10))
	first, err := coordinator.Decide(context.Background(), batch)
	if err != nil || !first.Granted {
		t.Fatalf("first decision = %+v, %v", first, err)
	}

	clock.now = clock.now.Add(horizon - time.Nanosecond)
	second, err := coordinator.Decide(context.Background(), batch)
	if err != nil || !second.Granted {
		t.Fatalf("decision before completion horizon = %+v, %v", second, err)
	}
	if got := testRateBudgetLeaseCount(t, coordinator); got != 2 {
		t.Fatalf("lease records before completion horizon = %d, want 2", got)
	}
	if err := coordinator.Finish(context.Background(), first.Lease, connsdk.CompletionObservation{
		Attempted:    true,
		Status:       http.StatusTooManyRequests,
		HasReset:     true,
		ResetAt:      clock.now.Add(time.Minute),
		HasRemaining: true,
		Remaining:    0,
	}); err != nil {
		t.Fatalf("late finish before completion horizon: %v", err)
	}

	decision, err := coordinator.Decide(context.Background(), batch)
	if err != nil {
		t.Fatalf("decision after late observation: %v", err)
	}
	if decision.Granted || decision.Wait != time.Minute {
		t.Fatalf("decision after late observation = %+v, want reset-window refusal", decision)
	}
}

func TestRateBudgetCompletionObservationHorizonDropsAtBoundaryWithoutRefund(t *testing.T) {
	const horizon = 2 * time.Minute
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)}
	coordinator := newTestRateBudgetCoordinator(t, clock, RateBudgetCoordinatorOptions{
		LeaseTTL:                     time.Second,
		CompletionObservationHorizon: horizon,
	})
	policy := testReservationPolicyWithBudget(t, "core", "opaque-scope-a", fixedRequestBudget(2, 600))
	batch := testReservationBatch(policy)
	first, err := coordinator.Decide(context.Background(), batch)
	if err != nil || !first.Granted {
		t.Fatalf("first decision = %+v, %v", first, err)
	}

	clock.now = clock.now.Add(horizon)
	observation := connsdk.CompletionObservation{
		Attempted:    true,
		Status:       http.StatusTooManyRequests,
		HasReset:     true,
		ResetAt:      clock.now.Add(time.Minute),
		HasRemaining: true,
		Remaining:    0,
	}
	if err := coordinator.Finish(context.Background(), first.Lease, observation); err != nil {
		t.Fatalf("finish at completion horizon: %v", err)
	}
	if err := coordinator.Finish(context.Background(), first.Lease, observation); err != nil {
		t.Fatalf("repeat finish at completion horizon: %v", err)
	}
	if got := testRateBudgetLeaseCount(t, coordinator); got != 0 {
		t.Fatalf("lease records at completion horizon = %d, want 0", got)
	}

	second, err := coordinator.Decide(context.Background(), batch)
	if err != nil || !second.Granted {
		t.Fatalf("decision after expired finish = %+v, %v", second, err)
	}
	third, err := coordinator.Decide(context.Background(), batch)
	if err != nil {
		t.Fatalf("third decision after expired finish: %v", err)
	}
	if third.Granted || third.Wait != 8*time.Minute {
		t.Fatalf("third decision after expired finish = %+v, want charged fixed-window refusal", third)
	}
}

func TestRateBudgetCompletionObservationHorizonBoundsAbandonedLeaseRecordsAndScans(t *testing.T) {
	const horizon = 2 * time.Minute
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)}
	coordinator := newTestRateBudgetCoordinator(t, clock, RateBudgetCoordinatorOptions{
		MaxInFlight:                  1,
		LeaseTTL:                     time.Second,
		CompletionObservationHorizon: horizon,
	})
	batch := testReservationBatch(testReservationPolicy(t, "core", "opaque-scope-a", 100))
	for attempt := 0; attempt < 8; attempt++ {
		decision, err := coordinator.Decide(context.Background(), batch)
		if err != nil || !decision.Granted {
			t.Fatalf("abandoned lease decision %d = %+v, %v", attempt, decision, err)
		}
		if got, want := testRateBudgetLeaseCount(t, coordinator), attempt+1; got != want {
			t.Fatalf("abandoned lease records after decision %d = %d, want %d", attempt, got, want)
		}
		clock.now = clock.now.Add(2 * time.Second)
	}
	clock.now = clock.now.Add(horizon)
	decision, err := coordinator.Decide(context.Background(), batch)
	if err != nil || !decision.Granted {
		t.Fatalf("decision after abandoned lease horizon = %+v, %v", decision, err)
	}
	if got := testRateBudgetLeaseCount(t, coordinator); got != 1 {
		t.Fatalf("abandoned lease records after cleanup = %d, want one bounded scan record", got)
	}
}
