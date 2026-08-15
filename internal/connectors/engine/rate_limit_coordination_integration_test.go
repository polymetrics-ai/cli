//go:build coordinationintegration

package engine

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
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/coordination"
)

const githubCertificationRateLimitHelperEnv = "POLYMETRICS_GITHUB_CERTIFICATION_RATE_LIMIT_HELPER"

func TestGitHubCertificationSharedBudgetCoordinatesSeparateProcesses(t *testing.T) {
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
	// Each parent run creates a fresh, non-secret declared account subject. The
	// shared key therefore has no inherited capacity; its bounded 60-second TTL
	// removes the test state without a raw-key cleanup escape hatch.
	subject := fmt.Sprintf("github-certification-multiprocess-%d", time.Now().UnixNano())

	var sends int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		sends++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	runChild := func() string {
		child := exec.Command(os.Args[0], "-test.run=^TestGitHubCertificationSharedBudgetHelperProcess$", "-test.v")
		child.Env = []string{
			githubCertificationRateLimitHelperEnv + "=1",
			"POLYMETRICS_DRAGONFLY_ADDR=" + addr,
			"POLYMETRICS_GITHUB_RATE_LIMIT_SERVER=" + server.URL,
			"POLYMETRICS_GITHUB_RATE_LIMIT_SUBJECT=" + subject,
		}
		output, err := child.CombinedOutput()
		if err != nil {
			t.Fatalf("GitHub certification helper: %v: %s", err, output)
		}
		return string(output)
	}

	first := runChild()
	if !strings.Contains(first, "github-certification-rate-limit-result=sent") {
		t.Fatalf("first process result = %s, want sent", first)
	}
	second := runChild()
	if !strings.Contains(second, "github-certification-rate-limit-result=blocked") {
		t.Fatalf("second process result = %s, want shared-budget block", second)
	}
	if got, want := sends, 1; got != want {
		t.Fatalf("second certification process reached provider after first consumed shared budget: sends = %d, want %d", got, want)
	}
}

func TestGitHubCertificationSharedBudgetHelperProcess(t *testing.T) {
	if os.Getenv(githubCertificationRateLimitHelperEnv) != "1" {
		return
	}
	addr := os.Getenv("POLYMETRICS_DRAGONFLY_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	shared := coordination.OpenSharedRateLimitRegistry(addr)
	defer func() { _ = shared.Close() }()
	ConfigureSharedRateLimitRegistry(shared)
	defer ConfigureSharedRateLimitRegistry(nil)
	bundle, err := Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load GitHub bundle: %v", err)
	}
	bundle.HTTP.URL = os.Getenv("POLYMETRICS_GITHUB_RATE_LIMIT_SERVER")
	bundle.HTTP.Auth = nil
	for i := range bundle.RateLimits.Policies {
		policy := &bundle.RateLimits.Policies[i]
		if policy.ID != "certification-authenticated-user" {
			continue
		}
		limit, seconds := 1, 60
		policy.Budgets = []connsdk.RateLimitBudget{{
			Model:         connsdk.RateLimitBudgetFixedWindow,
			Dimension:     connsdk.RateLimitBudgetSustained,
			Unit:          connsdk.RateLimitBudgetRequests,
			Limit:         &limit,
			WindowSeconds: &seconds,
		}}
	}
	subject := os.Getenv("POLYMETRICS_GITHUB_RATE_LIMIT_SUBJECT")
	if subject == "" {
		t.Fatal("GitHub certification helper requires its declared rate-limit subject")
	}
	config := githubRateLimitConfig(t, "token", "rate_limit_account", subject)
	config.Config["tier"] = "certification"
	runtime, err := newRuntime(context.Background(), bundle, config, nil)
	if err != nil {
		t.Fatalf("newRuntime: %v", err)
	}
	requester, err := runtime.RequesterFor(http.MethodGet, "/repos/octocat/example")
	if err != nil {
		t.Fatalf("RequesterFor: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	_, err = requester.Do(ctx, http.MethodGet, "/repos/octocat/example", nil, nil)
	switch {
	case err == nil:
		fmt.Fprintln(os.Stdout, "github-certification-rate-limit-result=sent")
	case errors.Is(err, context.DeadlineExceeded):
		fmt.Fprintln(os.Stdout, "github-certification-rate-limit-result=blocked")
	default:
		t.Fatalf("GitHub certification helper request: %v", err)
	}
}

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
