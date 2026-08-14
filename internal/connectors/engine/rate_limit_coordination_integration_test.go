//go:build coordinationintegration

package engine

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/coordination"
)

func TestRequireSharedRateLimitPolicyUsesAvailableCoordinator(t *testing.T) {
	if os.Getenv("POLYMETRICS_COORDINATION_INTEGRATION") != "1" {
		t.Skip("set POLYMETRICS_COORDINATION_INTEGRATION=1 with a local Dragonfly endpoint to run")
	}
	addr := os.Getenv("POLYMETRICS_DRAGONFLY_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	shared := coordination.OpenSharedRateLimitRegistry(addr)
	t.Cleanup(func() { _ = shared.Close() })
	if err := shared.EnsureAvailable(context.Background()); err != nil {
		t.Fatalf("required local Dragonfly coordinator is unavailable: %v", err)
	}
	restore := replaceSharedRateLimitRegistryForTest(shared)
	t.Cleanup(restore)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	t.Cleanup(server.Close)
	bundle := withAllRateLimit(Bundle{Name: "shared-required-integration", HTTP: HTTPBase{URL: server.URL}})
	bundle.RateLimits.Policies[0].Coordination = connsdk.RateLimitCoordinationRequireShared
	runtime, err := newRuntime(context.Background(), bundle, rateLimitTestConfig(t), nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	requester, err := runtime.RequesterFor(http.MethodGet, "/widgets")
	if err != nil {
		t.Fatalf("RequesterFor: %v", err)
	}
	admission, ok := requester.Admission.(resolvedRateLimitAdmission)
	if !ok || len(admission) != 1 {
		t.Fatalf("require_shared admission = %T, want one resolved shared policy", requester.Admission)
	}
	if _, ok := admission[0].admission.(*coordination.SharedRateLimiter); !ok {
		t.Fatalf("require_shared limiter = %T, want *coordination.SharedRateLimiter", admission[0].admission)
	}
	if _, err := requester.Do(context.Background(), http.MethodGet, "/widgets", nil, nil); err != nil {
		t.Fatalf("require_shared request: %v", err)
	}
	t.Log("require_shared result=shared coordinator admitted the request")
}

func TestSharedRateLimitReservationRetainsObservedBlock(t *testing.T) {
	if os.Getenv("POLYMETRICS_COORDINATION_INTEGRATION") != "1" {
		t.Skip("set POLYMETRICS_COORDINATION_INTEGRATION=1 with a local Dragonfly endpoint to run")
	}
	addr := os.Getenv("POLYMETRICS_DRAGONFLY_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	shared := coordination.OpenSharedRateLimitRegistry(addr)
	t.Cleanup(func() { _ = shared.Close() })
	if err := shared.EnsureAvailable(context.Background()); err != nil {
		t.Fatalf("required local Dragonfly coordinator is unavailable: %v", err)
	}
	limit, seconds := 1, 1
	limiter := shared.Limiter(coordination.RateLimitKey{
		Connector: "shared-rate-limit-test",
		PolicyID:  "observed-block-retention",
		Scope:     connectors.RateLimitScopeKey("integration-opaque-reservation-block"),
	}, []connsdk.RateLimitBudget{{
		Model:         connsdk.RateLimitBudgetFixedWindow,
		Dimension:     connsdk.RateLimitBudgetSustained,
		Unit:          connsdk.RateLimitBudgetRequests,
		Limit:         &limit,
		WindowSeconds: &seconds,
	}})
	observedAt := time.Now()
	limiter.Observe(context.Background(), connsdk.RateLimitObservation{
		Attempted:  true,
		Status:     http.StatusTooManyRequests,
		HasReset:   true,
		ResetAt:    observedAt.Add(4 * time.Second),
		ObservedAt: observedAt,
	})
	firstCtx, firstCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer firstCancel()
	if err := limiter.Admit(firstCtx, connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("initial blocked admission = %v, want context deadline", err)
	}
	time.Sleep(2200 * time.Millisecond)
	secondCtx, secondCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer secondCancel()
	if err := limiter.Admit(secondCtx, connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 2}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("admission after normal budget TTL = %v, want observed block to remain", err)
	}
}
