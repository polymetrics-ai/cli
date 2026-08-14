//go:build coordinationintegration

package coordination

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const sharedRateLimitHelperEnv = "POLYMETRICS_COORDINATION_HELPER"

func TestSharedRateLimitCoordinatesSeparateProcesses(t *testing.T) {
	if os.Getenv("POLYMETRICS_COORDINATION_INTEGRATION") != "1" {
		t.Skip("set POLYMETRICS_COORDINATION_INTEGRATION=1 with a local Dragonfly endpoint to run")
	}
	addr := os.Getenv("POLYMETRICS_DRAGONFLY_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	dragonfly := OpenDragonfly(addr)
	t.Cleanup(func() { _ = dragonfly.Close() })
	if err := dragonfly.Ping(context.Background()); err != nil {
		t.Fatalf("required local Dragonfly coordinator is unavailable: %v", err)
	}
	key := RateLimitKey{Connector: "paced", PolicyID: "shared-window", Scope: connectors.RateLimitScopeKey("integration-opaque-scope")}
	if err := dragonfly.client.Del(context.Background(), sharedRateLimitRedisKey(key)).Err(); err != nil {
		t.Fatalf("clear integration rate-limit key: %v", err)
	}
	t.Cleanup(func() { _ = dragonfly.client.Del(context.Background(), sharedRateLimitRedisKey(key)).Err() })

	outputs := make(chan string, 2)
	for range 2 {
		go func() {
			child := exec.Command(os.Args[0], "-test.run=^TestSharedRateLimitHelperProcess$", "-test.v")
			// The helper gets only the opt-in test switch and a non-secret local
			// coordinator address. It cannot inherit any calling credential env.
			child.Env = []string{
				sharedRateLimitHelperEnv + "=1",
				"POLYMETRICS_DRAGONFLY_ADDR=" + addr,
			}
			result, err := child.CombinedOutput()
			if err != nil {
				outputs <- fmt.Sprintf("child error: %v: %s", err, result)
				return
			}
			outputs <- string(result)
		}()
	}

	granted, blocked := 0, 0
	for range 2 {
		result := <-outputs
		switch {
		case strings.Contains(result, "shared-rate-limit-result=granted"):
			granted++
		case strings.Contains(result, "shared-rate-limit-result=blocked"):
			blocked++
		default:
			t.Fatalf("helper returned an unexpected result: %s", result)
		}
	}
	if granted != 1 || blocked != 1 {
		t.Fatalf("separate-process shared budget grants=%d blocked=%d, want one of each", granted, blocked)
	}
}

func TestSharedRateLimitHonorsEveryDeclaredBudgetModel(t *testing.T) {
	if os.Getenv("POLYMETRICS_COORDINATION_INTEGRATION") != "1" {
		t.Skip("set POLYMETRICS_COORDINATION_INTEGRATION=1 with a local Dragonfly endpoint to run")
	}
	addr := os.Getenv("POLYMETRICS_DRAGONFLY_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	dragonfly := OpenDragonfly(addr)
	t.Cleanup(func() { _ = dragonfly.Close() })
	if err := dragonfly.Ping(context.Background()); err != nil {
		t.Fatalf("required local Dragonfly coordinator is unavailable: %v", err)
	}
	for _, model := range []connsdk.RateLimitBudgetModel{
		connsdk.RateLimitBudgetFixedWindow,
		connsdk.RateLimitBudgetSlidingWindow,
		connsdk.RateLimitBudgetTokenBucket,
		connsdk.RateLimitBudgetLeakyBucket,
	} {
		t.Run(string(model), func(t *testing.T) {
			key := RateLimitKey{Connector: "paced", PolicyID: "shared-" + string(model), Scope: connectors.RateLimitScopeKey("integration-opaque-scope")}
			if err := dragonfly.client.Del(context.Background(), sharedRateLimitRedisKey(key)).Err(); err != nil {
				t.Fatalf("clear integration rate-limit key: %v", err)
			}
			t.Cleanup(func() { _ = dragonfly.client.Del(context.Background(), sharedRateLimitRedisKey(key)).Err() })
			limiter := NewSharedRateLimitRegistry(dragonfly).Limiter(key, []connsdk.RateLimitBudget{integrationBudget(model)})
			if err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: "GET", Attempt: 1}); err != nil {
				t.Fatalf("initial %s admission: %v", model, err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			if err := limiter.Admit(ctx, connsdk.RateLimitRequest{Method: "GET", Attempt: 2}); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("second %s admission = %v, want context deadline while shared budget is exhausted", model, err)
			}
		})
	}
}

func TestSharedRateLimitObservationDoesNotBlockNonExhaustedReset(t *testing.T) {
	if os.Getenv("POLYMETRICS_COORDINATION_INTEGRATION") != "1" {
		t.Skip("set POLYMETRICS_COORDINATION_INTEGRATION=1 with a local Dragonfly endpoint to run")
	}
	addr := os.Getenv("POLYMETRICS_DRAGONFLY_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	dragonfly := OpenDragonfly(addr)
	t.Cleanup(func() { _ = dragonfly.Close() })
	if err := dragonfly.Ping(context.Background()); err != nil {
		t.Fatalf("required local Dragonfly coordinator is unavailable: %v", err)
	}
	key := RateLimitKey{Connector: "paced", PolicyID: "shared-nonexhausted-reset", Scope: connectors.RateLimitScopeKey("integration-opaque-scope")}
	if err := dragonfly.client.Del(context.Background(), sharedRateLimitRedisKey(key)).Err(); err != nil {
		t.Fatalf("clear integration rate-limit key: %v", err)
	}
	t.Cleanup(func() { _ = dragonfly.client.Del(context.Background(), sharedRateLimitRedisKey(key)).Err() })
	limiter := NewSharedRateLimitRegistry(dragonfly).Limiter(key, []connsdk.RateLimitBudget{fixedRequestBudget(100, 60)})
	limiter.Observe(context.Background(), connsdk.RateLimitObservation{
		Attempted:    true,
		Status:       http.StatusOK,
		HasRemaining: true,
		Remaining:    99,
		HasReset:     true,
		ResetAt:      time.Now().Add(time.Minute),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := limiter.Admit(ctx, connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1}); err != nil {
		t.Fatalf("non-exhausted reset admission: %v", err)
	}
}

func TestSharedRateLimitUsesCoordinatorTimeForAbsoluteReset(t *testing.T) {
	if os.Getenv("POLYMETRICS_COORDINATION_INTEGRATION") != "1" {
		t.Skip("set POLYMETRICS_COORDINATION_INTEGRATION=1 with a local Dragonfly endpoint to run")
	}
	addr := os.Getenv("POLYMETRICS_DRAGONFLY_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	dragonfly := OpenDragonfly(addr)
	t.Cleanup(func() { _ = dragonfly.Close() })
	if err := dragonfly.Ping(context.Background()); err != nil {
		t.Fatalf("required local Dragonfly coordinator is unavailable: %v", err)
	}
	key := RateLimitKey{Connector: "paced", PolicyID: "shared-absolute-reset", Scope: connectors.RateLimitScopeKey("integration-opaque-scope")}
	if err := dragonfly.client.Del(context.Background(), sharedRateLimitRedisKey(key)).Err(); err != nil {
		t.Fatalf("clear integration rate-limit key: %v", err)
	}
	t.Cleanup(func() { _ = dragonfly.client.Del(context.Background(), sharedRateLimitRedisKey(key)).Err() })
	limiter := NewSharedRateLimitRegistry(dragonfly).Limiter(key, []connsdk.RateLimitBudget{fixedRequestBudget(1, 60)})
	resetAt := time.Now().Add(4 * time.Second).UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", resetAt.Format(http.TimeFormat))
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(server.Close)
	requester := &connsdk.Requester{
		BaseURL:        server.URL,
		DisableRetries: true,
		Now:            func() time.Time { return resetAt.Add(time.Minute) },
		Observer:       limiter,
	}
	if _, err := requester.Do(context.Background(), http.MethodGet, "/limited", nil, nil); err == nil {
		t.Fatal("rate-limited requester did not return an error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := limiter.Admit(ctx, connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("admission under absolute reset = %v, want context deadline", err)
	}
}

func TestSharedRateLimitObservationsTightenProviderState(t *testing.T) {
	if os.Getenv("POLYMETRICS_COORDINATION_INTEGRATION") != "1" {
		t.Skip("set POLYMETRICS_COORDINATION_INTEGRATION=1 with a local Dragonfly endpoint to run")
	}
	addr := os.Getenv("POLYMETRICS_DRAGONFLY_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	dragonfly := OpenDragonfly(addr)
	t.Cleanup(func() { _ = dragonfly.Close() })
	if err := dragonfly.Ping(context.Background()); err != nil {
		t.Fatalf("required local Dragonfly coordinator is unavailable: %v", err)
	}

	for _, tt := range []struct {
		name        string
		budget      connsdk.RateLimitBudget
		before      bool
		observation connsdk.RateLimitObservation
	}{
		{
			name:        "remaining",
			budget:      fixedRequestBudget(10, 60),
			observation: connsdk.RateLimitObservation{Attempted: true, HasRemaining: true, Remaining: 0},
		},
		{
			name:        "limit",
			budget:      fixedRequestBudget(10, 60),
			before:      true,
			observation: connsdk.RateLimitObservation{Attempted: true, HasLimit: true, Limit: 1},
		},
		{
			name:        "actual point cost",
			budget:      bucketPointBudget(connsdk.RateLimitBudgetTokenBucket, 10, 0.001, 1),
			before:      true,
			observation: connsdk.RateLimitObservation{Attempted: true, HasCost: true, Cost: 10},
		},
		{
			name:        "429 without reset",
			budget:      fixedRequestBudget(10, 60),
			observation: connsdk.RateLimitObservation{Attempted: true, Status: http.StatusTooManyRequests},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			key := RateLimitKey{Connector: "paced", PolicyID: "shared-observation-" + strings.ReplaceAll(tt.name, " ", "-"), Scope: connectors.RateLimitScopeKey("integration-opaque-scope")}
			if err := dragonfly.client.Del(context.Background(), sharedRateLimitRedisKey(key)).Err(); err != nil {
				t.Fatalf("clear integration rate-limit key: %v", err)
			}
			t.Cleanup(func() { _ = dragonfly.client.Del(context.Background(), sharedRateLimitRedisKey(key)).Err() })
			limiter := NewSharedRateLimitRegistry(dragonfly).Limiter(key, []connsdk.RateLimitBudget{tt.budget})
			if tt.before {
				if err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1}); err != nil {
					t.Fatalf("initial admission: %v", err)
				}
			}
			limiter.Observe(context.Background(), tt.observation)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			if err := limiter.Admit(ctx, connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 2}); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("admission after %s observation = %v, want context deadline", tt.name, err)
			}
		})
	}
}

func TestSharedRateLimitReservationRetainsObservedPointDebt(t *testing.T) {
	if os.Getenv("POLYMETRICS_COORDINATION_INTEGRATION") != "1" {
		t.Skip("set POLYMETRICS_COORDINATION_INTEGRATION=1 with a local Dragonfly endpoint to run")
	}
	addr := os.Getenv("POLYMETRICS_DRAGONFLY_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	dragonfly := OpenDragonfly(addr)
	t.Cleanup(func() { _ = dragonfly.Close() })
	if err := dragonfly.Ping(context.Background()); err != nil {
		t.Fatalf("required local Dragonfly coordinator is unavailable: %v", err)
	}
	key := RateLimitKey{Connector: "paced", PolicyID: "shared-observed-point-debt", Scope: connectors.RateLimitScopeKey("integration-opaque-scope")}
	if err := dragonfly.client.Del(context.Background(), sharedRateLimitRedisKey(key)).Err(); err != nil {
		t.Fatalf("clear integration rate-limit key: %v", err)
	}
	t.Cleanup(func() { _ = dragonfly.client.Del(context.Background(), sharedRateLimitRedisKey(key)).Err() })
	capacity, restore, defaultCost := 1, 1000.0, 1.0
	limiter := NewSharedRateLimitRegistry(dragonfly).Limiter(key, []connsdk.RateLimitBudget{{
		Model:            connsdk.RateLimitBudgetTokenBucket,
		Dimension:        connsdk.RateLimitBudgetSustained,
		Unit:             connsdk.RateLimitBudgetPoints,
		Capacity:         &capacity,
		RestorePerSecond: &restore,
		Cost:             &connsdk.RateLimitCost{DefaultCost: &defaultCost},
	}})
	if err := limiter.Admit(context.Background(), connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 1}); err != nil {
		t.Fatalf("initial point-budget admission: %v", err)
	}
	limiter.Observe(context.Background(), connsdk.RateLimitObservation{Attempted: true, HasCost: true, Cost: 5000})
	firstCtx, firstCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer firstCancel()
	if err := limiter.Admit(firstCtx, connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 2}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("initial debt admission = %v, want context deadline", err)
	}
	time.Sleep(2500 * time.Millisecond)
	secondCtx, secondCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer secondCancel()
	if err := limiter.Admit(secondCtx, connsdk.RateLimitRequest{Method: http.MethodGet, Attempt: 3}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("admission after normal budget TTL = %v, want observed debt to remain", err)
	}
}

func integrationBudget(model connsdk.RateLimitBudgetModel) connsdk.RateLimitBudget {
	limit, seconds := 1, 10
	capacity, restore := 1, 0.001
	b := connsdk.RateLimitBudget{Model: model, Dimension: connsdk.RateLimitBudgetSustained, Unit: connsdk.RateLimitBudgetRequests}
	switch model {
	case connsdk.RateLimitBudgetFixedWindow, connsdk.RateLimitBudgetSlidingWindow:
		b.Limit, b.WindowSeconds = &limit, &seconds
	case connsdk.RateLimitBudgetTokenBucket, connsdk.RateLimitBudgetLeakyBucket:
		b.Capacity, b.RestorePerSecond = &capacity, &restore
	}
	return b
}

func TestSharedRateLimitHelperProcess(t *testing.T) {
	if os.Getenv(sharedRateLimitHelperEnv) != "1" {
		return
	}
	addr := os.Getenv("POLYMETRICS_DRAGONFLY_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	registry := OpenSharedRateLimitRegistry(addr)
	defer func() { _ = registry.Close() }()
	limit, seconds := 1, 10
	limiter := registry.Limiter(RateLimitKey{
		Connector: "paced",
		PolicyID:  "shared-window",
		Scope:     connectors.RateLimitScopeKey("integration-opaque-scope"),
	}, []connsdk.RateLimitBudget{{
		Model:         connsdk.RateLimitBudgetFixedWindow,
		Dimension:     connsdk.RateLimitBudgetSustained,
		Unit:          connsdk.RateLimitBudgetRequests,
		Limit:         &limit,
		WindowSeconds: &seconds,
	}})
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	err := limiter.Admit(ctx, connsdk.RateLimitRequest{Method: "GET", Attempt: 1})
	switch {
	case err == nil:
		fmt.Fprintln(os.Stdout, "shared-rate-limit-result=granted")
	case errors.Is(err, context.DeadlineExceeded):
		fmt.Fprintln(os.Stdout, "shared-rate-limit-result=blocked")
	default:
		t.Fatalf("shared helper admission: %v", err)
	}
}
