//go:build coordinationintegration

package github_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
	githubhooks "polymetrics.ai/internal/connectors/hooks/github"
	"polymetrics.ai/internal/coordination"
)

const (
	githubAppRateAdmissionHelperEnv = "POLYMETRICS_GITHUB_APP_RATE_ADMISSION_HELPER"
	githubAppRateAdmissionURL       = "POLYMETRICS_GITHUB_APP_RATE_ADMISSION_URL"
	githubAppRateAdmissionScope     = "POLYMETRICS_GITHUB_APP_RATE_ADMISSION_SCOPE"
)

// TestGitHubAppAuthRateAdmissionSharedBudgetAcrossProcesses proves the auth
// token request reaches the real shared-admission path: two independently
// started test processes share one Dragonfly-backed request budget. The local
// HTTP fixture is only the provider endpoint; it never substitutes for the
// coordinator.
func TestGitHubAppAuthRateAdmissionSharedBudgetAcrossProcesses(t *testing.T) {
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

	var mints atomic.Int32
	var routeMu sync.Mutex
	var route string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		mints.Add(1)
		routeMu.Lock()
		route = request.URL.Path
		routeMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"token":"synthetic-installation-token"}`))
	}))
	t.Cleanup(server.Close)

	// The scope makes the coordinator key unique to this test run without
	// passing a credential to either child process. Both children receive the
	// same public installation identifier and therefore contend for one key.
	scope := fmt.Sprintf("live-admission-%d", time.Now().UnixNano())
	results := make(chan string, 2)
	for range 2 {
		go func() {
			child := exec.Command(os.Args[0], "-test.run=^TestGitHubAppAuthRateAdmissionHelperProcess$", "-test.v")
			child.Env = []string{
				githubAppRateAdmissionHelperEnv + "=1",
				"POLYMETRICS_DRAGONFLY_ADDR=" + addr,
				githubAppRateAdmissionURL + "=" + server.URL,
				githubAppRateAdmissionScope + "=" + scope,
			}
			output, err := child.CombinedOutput()
			if err != nil {
				results <- "child-error"
				return
			}
			results <- string(output)
		}()
	}

	minted, blocked := 0, 0
	for range 2 {
		result := <-results
		switch {
		case strings.Contains(result, "github-app-rate-admission-result=minted"):
			minted++
		case strings.Contains(result, "github-app-rate-admission-result=blocked"):
			blocked++
		default:
			t.Fatal("GitHub App admission helper did not report a valid shared-budget outcome")
		}
	}
	if minted != 1 || blocked != 1 {
		t.Fatalf("shared GitHub App token budget outcomes minted=%d blocked=%d, want one of each", minted, blocked)
	}
	physicalMints := mints.Load()
	if physicalMints != 1 {
		t.Fatalf("physical GitHub App token sends = %d, want 1 after the shared budget tightened", physicalMints)
	}
	routeMu.Lock()
	gotRoute := route
	routeMu.Unlock()
	if want := "/app/installations/" + scope + "/access_tokens"; gotRoute != want {
		t.Fatalf("minted token route = %q, want %q", gotRoute, want)
	}
	t.Logf("shared GitHub App admission: processes=2 minted=%d blocked=%d physical_token_mints=%d route=%q", minted, blocked, physicalMints, gotRoute)
}

func TestGitHubAppAuthRateAdmissionHelperProcess(t *testing.T) {
	if os.Getenv(githubAppRateAdmissionHelperEnv) != "1" {
		return
	}
	addr := os.Getenv("POLYMETRICS_DRAGONFLY_ADDR")
	baseURL := os.Getenv(githubAppRateAdmissionURL)
	scope := os.Getenv(githubAppRateAdmissionScope)
	if addr == "" || baseURL == "" || scope == "" {
		t.Fatal("GitHub App admission helper is missing non-secret integration configuration")
	}

	shared := coordination.OpenSharedRateLimitRegistry(addr)
	defer func() { _ = shared.Close() }()
	engine.ConfigureSharedRateLimitRegistry(shared)
	defer engine.ConfigureSharedRateLimitRegistry(nil)

	bundle := requireSharedGitHubAppBundle(t)
	for i := range bundle.RateLimits.Policies {
		if bundle.RateLimits.Policies[i].ID != "app-installation" {
			continue
		}
		limit, windowSeconds := 1, 60
		bundle.RateLimits.Policies[i].Budgets = []connsdk.RateLimitBudget{{
			Model:         connsdk.RateLimitBudgetFixedWindow,
			Dimension:     connsdk.RateLimitBudgetSustained,
			Unit:          connsdk.RateLimitBudgetRequests,
			Limit:         &limit,
			WindowSeconds: &windowSeconds,
		}}
	}
	cfg := githubAppAuthAdmissionConfig(t)
	cfg.Config["base_url"] = baseURL
	cfg.Config["installation_id"] = scope

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	_, err := engine.NewRuntime(ctx, bundle, cfg, githubhooks.New())
	switch {
	case err == nil:
		fmt.Fprintln(os.Stdout, "github-app-rate-admission-result=minted")
	case errors.Is(err, context.DeadlineExceeded):
		fmt.Fprintln(os.Stdout, "github-app-rate-admission-result=blocked")
	default:
		t.Fatalf("GitHub App token admission: %v", err)
	}
}
