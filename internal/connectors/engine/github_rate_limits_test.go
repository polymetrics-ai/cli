package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/coordination"
)

const githubRateLimitSource = "https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api"

func TestGitHubDeclaredRateLimits(t *testing.T) {
	bundle, err := Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load(defs.FS, github): %v", err)
	}
	if bundle.RateLimits == nil {
		t.Fatal("GitHub has no rate_limits.json declaration")
	}
	if got, want := bundle.RateLimits.State, connsdk.RateLimitStateDeclared; got != want {
		t.Fatalf("GitHub rate-limit state = %q, want %q", got, want)
	}

	type policyExpectation struct {
		id           string
		authType     string
		scopeConfig  string
		scopeValue   string
		scopeKind    connsdk.RateLimitScopeSubjectKind
		primaryLimit int
	}
	expectations := []policyExpectation{
		{
			id:           "authenticated-user",
			authType:     "token",
			scopeConfig:  "rate_limit_account",
			scopeValue:   "octocat",
			scopeKind:    connsdk.RateLimitScopeAccount,
			primaryLimit: 5000,
		},
		{
			id:           "app-installation",
			authType:     "github_app",
			scopeConfig:  "installation_id",
			scopeValue:   "12345",
			scopeKind:    connsdk.RateLimitScopeInstallation,
			primaryLimit: 5000,
		},
		{
			id:           "actions-token",
			authType:     "github_token",
			scopeConfig:  "rate_limit_repository",
			scopeValue:   "octocat/example",
			scopeKind:    connsdk.RateLimitScopeEndpoint,
			primaryLimit: 1000,
		},
		{
			id:           "unauthenticated",
			authType:     "public",
			scopeConfig:  "rate_limit_ip",
			scopeValue:   "203.0.113.7",
			scopeKind:    connsdk.RateLimitScopeIP,
			primaryLimit: 60,
		},
	}

	policies := make(map[string]connsdk.RateLimitPolicy, len(bundle.RateLimits.Policies))
	for _, policy := range bundle.RateLimits.Policies {
		if got, want := policy.Source.URL, githubRateLimitSource; got != want {
			t.Fatalf("policy %q source URL = %q, want %q", policy.ID, got, want)
		}
		if got, want := policy.Source.RetrievedAt, "2026-08-08"; got != want {
			t.Fatalf("policy %q retrieved_at = %q, want %q", policy.ID, got, want)
		}
		if policy.Scope.SubjectConfig == "token" {
			t.Fatalf("policy %q scopes a limiter by credential material", policy.ID)
		}
		policies[policy.ID] = policy
	}

	for _, want := range expectations {
		want := want
		t.Run(want.id, func(t *testing.T) {
			policy, ok := policies[want.id]
			if !ok {
				t.Fatalf("GitHub policy %q is absent", want.id)
			}
			if policy.Scope.SubjectKind != want.scopeKind || policy.Scope.SubjectConfig != want.scopeConfig {
				t.Fatalf("policy %q scope = %+v, want %q/%q", want.id, policy.Scope, want.scopeKind, want.scopeConfig)
			}
			if !containsRateLimitName(policy.Selector.AuthTypes, want.authType) {
				t.Fatalf("policy %q auth types = %v, want %q", want.id, policy.Selector.AuthTypes, want.authType)
			}
			if !rateLimitEndpointMatches(policy.Selector.ExcludeEndpoints, http.MethodPost, "/graphql") {
				t.Fatalf("policy %q does not exclude GitHub GraphQL traffic", want.id)
			}
			if !hasFixedRequestBudget(policy, want.primaryLimit, 3600) {
				t.Fatalf("policy %q does not declare a %d-request/hour primary budget", want.id, want.primaryLimit)
			}
			if !hasSlidingPointBudget(policy, 900, 60, 5) {
				t.Fatalf("policy %q does not declare the conservative 900-point/minute secondary budget", want.id)
			}

			cfg := githubRateLimitConfig(t, want.authType, want.scopeConfig, want.scopeValue)
			resolver := newRateLimitResolver(bundle, cfg)
			runtime := &Runtime{
				baseRequester: &connsdk.Requester{},
				rateLimits:    resolver,
			}
			requester, err := runtime.RequesterFor(http.MethodGet, "/repos/{owner}/{repo}")
			if err != nil {
				t.Fatalf("RequesterFor: %v", err)
			}
			if requester.Admission != nil || requester.Observer != nil || requester.RouteRateLimits == nil {
				t.Fatalf("policy %q did not attach a path-aware rate-limit resolver", want.id)
			}
			graphqlRequester, err := runtime.RequesterFor(http.MethodPost, "/graphql")
			if err != nil {
				t.Fatalf("RequesterFor GraphQL: %v", err)
			}
			if graphqlRequester.Admission != nil || graphqlRequester.Observer != nil {
				t.Fatalf("policy %q applied REST rate-limit accounting to GitHub GraphQL", want.id)
			}
			defaultRequester, err := resolver.defaultRequester(&connsdk.Requester{})
			if err != nil {
				t.Fatalf("defaultRequester: %v", err)
			}
			if defaultRequester.Admission != nil || defaultRequester.Observer != nil {
				t.Fatalf("policy %q attached path-aware REST accounting to the hook runtime requester", want.id)
			}
		})
	}

	unmatched := &Runtime{
		baseRequester: &connsdk.Requester{},
		rateLimits: newRateLimitResolver(bundle, githubRateLimitConfig(t,
			"unmatched", "rate_limit_account", "octocat")),
	}
	requester, err := unmatched.RequesterFor(http.MethodGet, "/repos/{owner}/{repo}")
	if err != nil {
		t.Fatalf("RequesterFor unmatched auth type: %v", err)
	}
	if requester.Admission != nil || requester.Observer != nil {
		t.Fatal("unmatched GitHub auth type acquired a rate-limit policy")
	}
}

func TestGitHubRateLimitAdmissionPrecedesProviderAndIsolatesScope(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Limit", "1")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(githubRateLimitProofNow.Add(time.Minute).Unix(), 10))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	bundle, err := Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load(defs.FS, github): %v", err)
	}
	bundle.HTTP.URL = server.URL
	bundle.HTTP.Auth = nil
	// Keep the provider declaration and replace only the budget window so the
	// test can deterministically cross a local boundary without waiting an
	// hour. The runtime path, scope binding, and observation remain GitHub's.
	policy := bundle.RateLimits.Policies[0]
	budgets := make([]connsdk.RateLimitBudget, 0, 1)
	for i := range policy.Budgets {
		if policy.Budgets[i].Model == connsdk.RateLimitBudgetFixedWindow {
			limit, window := 1, 60
			policy.Budgets[i].Limit = &limit
			policy.Budgets[i].WindowSeconds = &window
			budgets = append(budgets, policy.Budgets[i])
		}
	}
	policy.Budgets = budgets
	bundle.RateLimits.Policies = []connsdk.RateLimitPolicy{policy}

	clock := &engineRateLimitClock{now: githubRateLimitProofNow}
	restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
	t.Cleanup(restore)
	firstConfig := githubRateLimitConfig(t, "token", "rate_limit_account", "provider-double-account-a")
	firstRuntime, err := newRuntime(context.Background(), bundle, firstConfig, nil)
	if err != nil {
		t.Fatalf("newRuntime first scope: %v", err)
	}
	firstRequester, err := firstRuntime.RequesterFor(http.MethodGet, "/repos/{owner}/{repo}")
	if err != nil {
		t.Fatalf("RequesterFor first scope: %v", err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		if _, err := firstRequester.Do(context.Background(), http.MethodGet, "/repos/provider-double-owner/provider-double-repo", nil, nil); err != nil {
			t.Fatalf("first scope request %d: %v", attempt, err)
		}
	}
	if len(clock.waits) != 1 || clock.waits[0] != time.Minute {
		t.Fatalf("same-scope waits = %v, want one local one-minute wait", clock.waits)
	}

	secondConfig := githubRateLimitConfig(t, "token", "rate_limit_account", "provider-double-account-b")
	secondRuntime, err := newRuntime(context.Background(), bundle, secondConfig, nil)
	if err != nil {
		t.Fatalf("newRuntime independent scope: %v", err)
	}
	secondRequester, err := secondRuntime.RequesterFor(http.MethodGet, "/repos/{owner}/{repo}")
	if err != nil {
		t.Fatalf("RequesterFor independent scope: %v", err)
	}
	if _, err := secondRequester.Do(context.Background(), http.MethodGet, "/repos/provider-double-owner/provider-double-repo", nil, nil); err != nil {
		t.Fatalf("independent scope request: %v", err)
	}
	if got, want := requests.Load(), int32(3); got != want {
		t.Fatalf("provider requests = %d, want %d", got, want)
	}
	// Every call returned 200, so the provider never had to reject a request;
	// the one-minute pause came from local admission after the first response.
}

var githubRateLimitProofNow = time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)

func githubRateLimitConfig(t *testing.T, authType, scopeConfig, scopeValue string) connectors.RuntimeConfig {
	t.Helper()
	identity, err := connectors.NewCoordinationIdentity([]byte("github-rate-limit-test-salt"), connectors.CredentialBinding{
		BindingID:      "github-rate-limit-test-binding",
		ProviderFamily: "github",
		AuthProfile:    authType,
	})
	if err != nil {
		t.Fatalf("NewCoordinationIdentity: %v", err)
	}
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"auth_type": authType,
			scopeConfig: scopeValue,
		},
		CoordinationIdentity: identity,
	}
}

func containsRateLimitName(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func hasFixedRequestBudget(policy connsdk.RateLimitPolicy, limit, seconds int) bool {
	for _, budget := range policy.Budgets {
		if budget.Model != connsdk.RateLimitBudgetFixedWindow || budget.Unit != connsdk.RateLimitBudgetRequests || budget.Limit == nil || budget.WindowSeconds == nil {
			continue
		}
		if *budget.Limit == limit && *budget.WindowSeconds == seconds {
			return true
		}
	}
	return false
}

func hasSlidingPointBudget(policy connsdk.RateLimitPolicy, limit, seconds int, defaultCost float64) bool {
	for _, budget := range policy.Budgets {
		if budget.Model != connsdk.RateLimitBudgetSlidingWindow || budget.Unit != connsdk.RateLimitBudgetPoints || budget.Limit == nil || budget.WindowSeconds == nil || budget.Cost == nil || budget.Cost.DefaultCost == nil {
			continue
		}
		if *budget.Limit == limit && *budget.WindowSeconds == seconds && *budget.Cost.DefaultCost == defaultCost {
			return true
		}
	}
	return false
}
