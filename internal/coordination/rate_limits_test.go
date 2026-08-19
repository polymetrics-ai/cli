package coordination

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

type fakeRateLimitClock struct {
	now   time.Time
	waits []time.Duration
}

func (c *fakeRateLimitClock) Now() time.Time { return c.now }

func (c *fakeRateLimitClock) Sleep(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.waits = append(c.waits, d)
	c.now = c.now.Add(d)
	return nil
}

func testRateLimitKey() RateLimitKey {
	return RateLimitKey{
		Connector: "paced",
		PolicyID:  "core",
		Scope:     connectors.RateLimitScopeKey("opaque-scope-projection"),
	}
}

func fixedRequestBudget(limit, seconds int) connsdk.RateLimitBudget {
	return connsdk.RateLimitBudget{
		Model:         connsdk.RateLimitBudgetFixedWindow,
		Dimension:     connsdk.RateLimitBudgetSustained,
		Unit:          connsdk.RateLimitBudgetRequests,
		Limit:         &limit,
		WindowSeconds: &seconds,
	}
}

func bucketPointBudget(model connsdk.RateLimitBudgetModel, capacity int, rate, cost float64) connsdk.RateLimitBudget {
	return connsdk.RateLimitBudget{
		Model:            model,
		Dimension:        connsdk.RateLimitBudgetBurst,
		Unit:             connsdk.RateLimitBudgetPoints,
		Capacity:         &capacity,
		RestorePerSecond: &rate,
		Cost:             &connsdk.RateLimitCost{DefaultCost: &cost, ResponseHeader: "X-Actual-Cost"},
	}
}

func TestRateLimitRegistrySharesOpaqueScopeAcrossLimiters(t *testing.T) {
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	registry := NewRateLimitRegistry(clock)
	budget := fixedRequestBudget(1, 60)

	first := registry.Limiter(testRateLimitKey(), []connsdk.RateLimitBudget{budget})
	second := registry.Limiter(testRateLimitKey(), []connsdk.RateLimitBudget{budget})
	if err := first.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1}); err != nil {
		t.Fatalf("first Admit: %v", err)
	}
	if err := second.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1}); err != nil {
		t.Fatalf("second Admit: %v", err)
	}
	if got, want := clock.waits, []time.Duration{time.Minute}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("waits = %v, want %v", got, want)
	}
}

func TestRateLimitRegistryWaitHonorsContextCancellation(t *testing.T) {
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	registry := NewRateLimitRegistry(clock)
	limiter := registry.Limiter(testRateLimitKey(), []connsdk.RateLimitBudget{fixedRequestBudget(1, 60)})
	if err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1}); err != nil {
		t.Fatalf("initial Admit: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := limiter.Admit(ctx, connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 2}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Admit error = %v, want context.Canceled", err)
	}
	if len(clock.waits) != 0 {
		t.Fatalf("cancelled Admit slept: %v", clock.waits)
	}
}

type cancellingRateLimitClock struct {
	fakeRateLimitClock
	cancel context.CancelFunc
}

func (c *cancellingRateLimitClock) Sleep(ctx context.Context, d time.Duration) error {
	c.waits = append(c.waits, d)
	c.cancel()
	return ctx.Err()
}

func TestRateLimitRegistryCancelsWhileWaiting(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	clock := &cancellingRateLimitClock{
		fakeRateLimitClock: fakeRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)},
		cancel:             cancel,
	}
	registry := NewRateLimitRegistry(clock)
	limiter := registry.Limiter(testRateLimitKey(), []connsdk.RateLimitBudget{fixedRequestBudget(1, 60)})
	if err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1}); err != nil {
		t.Fatalf("initial Admit: %v", err)
	}
	if err := limiter.Admit(ctx, connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 2}); !errors.Is(err, context.Canceled) {
		t.Fatalf("waiting Admit error = %v, want context.Canceled", err)
	}
	if got, want := clock.waits, []time.Duration{time.Minute}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("waits = %v, want %v", got, want)
	}
}

func TestRateLimitRegistryAppliesActualPointCostOnlyToTightenBudget(t *testing.T) {
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	registry := NewRateLimitRegistry(clock)
	limiter := registry.Limiter(testRateLimitKey(), []connsdk.RateLimitBudget{
		bucketPointBudget(connsdk.RateLimitBudgetTokenBucket, 2, 1, 1),
	})
	if err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodPost, Attempt: 1}); err != nil {
		t.Fatalf("initial Admit: %v", err)
	}
	limiter.Observe(context.Background(), connsdk.RateLimitObservation{Attempted: true, Cost: 2, HasCost: true})
	if err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodPost, Attempt: 2}); err != nil {
		t.Fatalf("second Admit: %v", err)
	}
	if got, want := clock.waits, []time.Duration{time.Second}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("waits = %v, want %v", got, want)
	}
}

func TestRateLimitRegistryHonorsEachReplenishmentModel(t *testing.T) {
	for _, model := range []connsdk.RateLimitBudgetModel{
		connsdk.RateLimitBudgetTokenBucket,
		connsdk.RateLimitBudgetLeakyBucket,
	} {
		t.Run(string(model), func(t *testing.T) {
			clock := &fakeRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
			registry := NewRateLimitRegistry(clock)
			limiter := registry.Limiter(testRateLimitKey(), []connsdk.RateLimitBudget{bucketPointBudget(model, 1, 2, 1)})
			if err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1}); err != nil {
				t.Fatalf("initial Admit: %v", err)
			}
			if err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 2}); err != nil {
				t.Fatalf("second Admit: %v", err)
			}
			if got, want := clock.waits, []time.Duration{500 * time.Millisecond}; len(got) != len(want) || got[0] != want[0] {
				t.Fatalf("waits = %v, want %v", got, want)
			}
		})
	}
}

func TestRateLimitRegistryHonorsSlidingWindow(t *testing.T) {
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	registry := NewRateLimitRegistry(clock)
	limit, window := 1, 60
	limiter := registry.Limiter(testRateLimitKey(), []connsdk.RateLimitBudget{{
		Model: connsdk.RateLimitBudgetSlidingWindow, Dimension: connsdk.RateLimitBudgetSustained,
		Unit: connsdk.RateLimitBudgetRequests, Limit: &limit, WindowSeconds: &window,
	}})
	if err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1}); err != nil {
		t.Fatalf("initial Admit: %v", err)
	}
	if err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 2}); err != nil {
		t.Fatalf("second Admit: %v", err)
	}
	if got, want := clock.waits, []time.Duration{time.Minute}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("waits = %v, want %v", got, want)
	}
}

func TestRateLimitRegistryHonorsProviderResetObservation(t *testing.T) {
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	registry := NewRateLimitRegistry(clock)
	limiter := registry.Limiter(testRateLimitKey(), []connsdk.RateLimitBudget{fixedRequestBudget(5, 60)})
	limiter.Observe(context.Background(), connsdk.RateLimitObservation{Attempted: true, HasReset: true, ResetAt: clock.now.Add(45 * time.Second), Status: http.StatusTooManyRequests})
	if err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1}); err != nil {
		t.Fatalf("Admit after reset observation: %v", err)
	}
	if got, want := clock.waits, []time.Duration{45 * time.Second}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("waits = %v, want %v", got, want)
	}
}

func TestRateLimitRegistryDoesNotBlockForNonExhaustedResetObservation(t *testing.T) {
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	registry := NewRateLimitRegistry(clock)
	limiter := registry.Limiter(testRateLimitKey(), []connsdk.RateLimitBudget{fixedRequestBudget(100, 60)})
	limiter.Observe(context.Background(), connsdk.RateLimitObservation{
		Attempted:    true,
		Status:       http.StatusOK,
		HasRemaining: true,
		Remaining:    99,
		HasReset:     true,
		ResetAt:      clock.now.Add(time.Minute),
	})
	if err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1}); err != nil {
		t.Fatalf("Admit after non-exhausted observation: %v", err)
	}
	if len(clock.waits) != 0 {
		t.Fatalf("non-exhausted reset blocked admission: %v", clock.waits)
	}
}

func TestRateLimitRegistryRejectsProviderLimitBelowOneRequestCost(t *testing.T) {
	clock := &fakeRateLimitClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	registry := NewRateLimitRegistry(clock)
	limiter := registry.Limiter(testRateLimitKey(), []connsdk.RateLimitBudget{fixedRequestBudget(1, 60)})
	limiter.Observe(context.Background(), connsdk.RateLimitObservation{Attempted: true, HasLimit: true, Limit: 0})
	if err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1}); err == nil {
		t.Fatal("Admit accepted a provider limit below one request cost")
	}
	if len(clock.waits) != 0 {
		t.Fatalf("impossible observed limit waited instead of refusing: %v", clock.waits)
	}
}
