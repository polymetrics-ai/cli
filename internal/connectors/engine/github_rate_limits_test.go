package engine

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/coordination"
)

const (
	githubRateLimitSource        = "https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api"
	githubGraphQLRateLimitSource = "https://docs.github.com/en/graphql/overview/rate-limits-and-query-limits-for-the-graphql-api"
)

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
	graphqlExpectations := []policyExpectation{
		{
			id:           "graphql-authenticated-user-primary",
			authType:     "token",
			scopeConfig:  "rate_limit_account",
			scopeValue:   "octocat",
			scopeKind:    connsdk.RateLimitScopeAccount,
			primaryLimit: 5000,
		},
		{
			id:           "graphql-app-installation-primary",
			authType:     "github_app",
			scopeConfig:  "installation_id",
			scopeValue:   "12345",
			scopeKind:    connsdk.RateLimitScopeInstallation,
			primaryLimit: 5000,
		},
		{
			id:           "graphql-actions-token-primary",
			authType:     "github_token",
			scopeConfig:  "rate_limit_repository",
			scopeValue:   "octocat/example",
			scopeKind:    connsdk.RateLimitScopeEndpoint,
			primaryLimit: 1000,
		},
	}

	policies := make(map[string]connsdk.RateLimitPolicy, len(bundle.RateLimits.Policies))
	for _, policy := range bundle.RateLimits.Policies {
		wantSource, wantRetrievedAt := githubRateLimitSource, "2026-08-08"
		if strings.Contains(policy.ID, "graphql") {
			wantSource, wantRetrievedAt = githubGraphQLRateLimitSource, "2026-08-14"
		}
		if got := policy.Source.URL; got != wantSource {
			t.Fatalf("policy %q source URL = %q, want %q", policy.ID, got, wantSource)
		}
		if got := policy.Source.RetrievedAt; got != wantRetrievedAt {
			want := wantRetrievedAt
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

	for _, want := range graphqlExpectations {
		want := want
		t.Run(want.id, func(t *testing.T) {
			policy, ok := policies[want.id]
			if !ok {
				t.Fatalf("GitHub GraphQL policy %q is absent", want.id)
			}
			if policy.Scope.SubjectKind != want.scopeKind || policy.Scope.SubjectConfig != want.scopeConfig {
				t.Fatalf("policy %q scope = %+v, want %q/%q", want.id, policy.Scope, want.scopeKind, want.scopeConfig)
			}
			if !containsRateLimitName(policy.Selector.AuthTypes, want.authType) || !rateLimitEndpointMatches(policy.Selector.Endpoints, http.MethodPost, "/graphql") {
				t.Fatalf("policy %q selector = %+v, want auth %q and POST /graphql", want.id, policy.Selector, want.authType)
			}
			if !hasFixedPointBudgetWithResponseBody(policy, want.primaryLimit, 3600, 1, "graphql_rate_limit") {
				t.Fatalf("policy %q does not declare a %d-point/hour GraphQL primary budget", want.id, want.primaryLimit)
			}
			secondaryID := strings.TrimSuffix(want.id, "-primary") + "-secondary"
			secondary, ok := policies[secondaryID]
			if !ok || secondary.Scope != policy.Scope || !rateLimitSelectorMatches(secondary.Selector, http.MethodPost, "/graphql", map[string]string{"auth_type": want.authType}) || !hasSlidingPointBudget(secondary, 2000, 60, 5) {
				t.Fatalf("policy %q does not pair a matching 2,000-point/minute GraphQL secondary budget", want.id)
			}

			cfg := githubRateLimitConfig(t, want.authType, want.scopeConfig, want.scopeValue)
			resolver := newRateLimitResolver(bundle, cfg)
			runtime := &Runtime{baseRequester: &connsdk.Requester{}, rateLimits: resolver}
			requester, err := runtime.RequesterFor(http.MethodPost, "/graphql")
			if err != nil {
				t.Fatalf("RequesterFor GraphQL: %v", err)
			}
			if requester.RouteRateLimits == nil {
				t.Fatalf("policy %q did not attach GraphQL path-aware admission", want.id)
			}
		})
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)
	unmatched := &Runtime{
		baseRequester: &connsdk.Requester{BaseURL: server.URL, DisableRetries: true},
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
	if _, err := requester.Do(context.Background(), http.MethodGet, "/repos/octocat/example", nil, nil); err != nil {
		t.Fatalf("unmatched local GitHub request: %v", err)
	}
	if got, want := requests, 1; got != want {
		t.Fatalf("unmatched local GitHub request count = %d, want %d", got, want)
	}
}

func TestGitHubCertificationRatePoliciesFailClosedBeforeSend(t *testing.T) {
	type policyExpectation struct {
		id          string
		method      string
		path        string
		authType    string
		scopeConfig string
		scopeValue  string
		scopeKind   connsdk.RateLimitScopeSubjectKind
	}
	expectations := []policyExpectation{
		{"certification-authenticated-user", http.MethodGet, "/repos/octocat/example", "token", "rate_limit_account", "octocat", connsdk.RateLimitScopeAccount},
		{"certification-app-installation", http.MethodGet, "/repos/octocat/example", "github_app", "installation_id", "12345", connsdk.RateLimitScopeInstallation},
		{"certification-actions-token", http.MethodGet, "/repos/octocat/example", "github_token", "rate_limit_repository", "octocat/example", connsdk.RateLimitScopeEndpoint},
		{"certification-unauthenticated", http.MethodGet, "/repos/octocat/example", "public", "rate_limit_ip", "203.0.113.7", connsdk.RateLimitScopeIP},
		{"certification-graphql-authenticated-user", http.MethodPost, "/graphql", "token", "rate_limit_account", "octocat", connsdk.RateLimitScopeAccount},
		{"certification-graphql-app-installation", http.MethodPost, "/graphql", "github_app", "installation_id", "12345", connsdk.RateLimitScopeInstallation},
		{"certification-graphql-actions-token", http.MethodPost, "/graphql", "github_token", "rate_limit_repository", "octocat/example", connsdk.RateLimitScopeEndpoint},
	}
	bundle, err := Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load(defs.FS, github): %v", err)
	}
	var sends atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		sends.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	bundle.HTTP.URL = server.URL
	bundle.HTTP.Auth = nil
	policies := make(map[string]connsdk.RateLimitPolicy, len(bundle.RateLimits.Policies))
	for _, policy := range bundle.RateLimits.Policies {
		policies[policy.ID] = policy
	}

	for _, want := range expectations {
		want := want
		t.Run(want.id, func(t *testing.T) {
			policy, ok := policies[want.id]
			if !ok {
				t.Fatalf("GitHub certification rate policy %q is absent", want.id)
			}
			if policy.Coordination != connsdk.RateLimitCoordinationRequireShared || !containsRateLimitName(policy.Selector.Tiers, "certification") {
				t.Fatalf("certification policy %q coordination = %q selector=%+v, want require_shared certification tier", want.id, policy.Coordination, policy.Selector)
			}
			if policy.Scope.SubjectKind != want.scopeKind || policy.Scope.SubjectConfig != want.scopeConfig {
				t.Fatalf("certification policy %q scope = %+v, want %q/%q", want.id, policy.Scope, want.scopeKind, want.scopeConfig)
			}
			cfg := githubRateLimitConfig(t, want.authType, want.scopeConfig, want.scopeValue)
			cfg.Config["tier"] = "certification"
			runtime, err := newRuntime(context.Background(), bundle, cfg, nil)
			if err != nil {
				t.Fatalf("newRuntime: %v", err)
			}
			requester, err := runtime.RequesterFor(want.method, want.path)
			if err != nil {
				t.Fatalf("RequesterFor: %v", err)
			}
			_, err = requester.Do(context.Background(), want.method, want.path, nil, map[string]any{"query": "fixed"})
			var unavailable *coordination.SharedRateLimitUnavailableError
			if !errors.As(err, &unavailable) || unavailable.Reason != coordination.SharedRateLimitCoordinatorNotConfigured {
				t.Fatalf("certification request error = %v, want coordinator-not-configured pre-send refusal", err)
			}
			if got := sends.Load(); got != 0 {
				t.Fatalf("refused certification policy %q sent %d provider requests, want 0", want.id, got)
			}
		})
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

func hasFixedPointBudgetWithResponseBody(policy connsdk.RateLimitPolicy, limit, seconds int, defaultCost float64, responseBody string) bool {
	for _, budget := range policy.Budgets {
		if budget.Model != connsdk.RateLimitBudgetFixedWindow || budget.Unit != connsdk.RateLimitBudgetPoints || budget.Limit == nil || budget.WindowSeconds == nil || budget.Cost == nil || budget.Cost.DefaultCost == nil {
			continue
		}
		if *budget.Limit == limit && *budget.WindowSeconds == seconds && *budget.Cost.DefaultCost == defaultCost && budget.Cost.ResponseBody == responseBody {
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
