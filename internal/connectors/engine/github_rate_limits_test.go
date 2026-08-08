package engine

import (
	"net/http"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/defs"
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
			if !hasFixedRequestBudget(policy, want.primaryLimit, 3600) {
				t.Fatalf("policy %q does not declare a %d-request/hour primary budget", want.id, want.primaryLimit)
			}

			cfg := githubRateLimitConfig(t, want.authType, want.scopeConfig, want.scopeValue)
			runtime := &Runtime{
				baseRequester: &connsdk.Requester{},
				rateLimits:    newRateLimitResolver(bundle, cfg),
			}
			requester, err := runtime.RequesterFor(http.MethodGet, "/repos/{owner}/{repo}")
			if err != nil {
				t.Fatalf("RequesterFor: %v", err)
			}
			if requester.Admission == nil || requester.Observer == nil {
				t.Fatalf("policy %q did not attach admission and observation hooks", want.id)
			}
		})
	}
}

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
