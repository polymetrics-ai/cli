package dockerhub_test

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/connsdk"
	connectorDefs "polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestDockerHubRegistryPullRateLimitsAreEmbedded(t *testing.T) {
	bundle, err := engine.Load(connectorDefs.FS, "dockerhub")
	if err != nil {
		t.Fatalf("load Docker Hub bundle: %v", err)
	}
	if bundle.RateLimits == nil {
		t.Fatal("Docker Hub bundle has no embedded provider-cited rate_limits.json")
	}
	if got, want := bundle.RateLimits.State, connsdk.RateLimitStateDeclared; got != want {
		t.Fatalf("rate limit state = %q, want %q", got, want)
	}
	if !strings.Contains(bundle.RateLimits.Reason, "Retry-After") {
		t.Fatalf("rate limit reason = %q, want documented unbudgeted abuse-limit handling", bundle.RateLimits.Reason)
	}
	policies := make(map[string]connsdk.RateLimitPolicy, len(bundle.RateLimits.Policies))
	for _, policy := range bundle.RateLimits.Policies {
		policies[policy.ID] = policy
	}
	assertDockerHubPullPolicy(t, policies, "registry-pull-unauthenticated", 100, "ip", "registry_client_ip", "unauthenticated", "")
	assertDockerHubPullPolicy(t, policies, "registry-pull-authenticated-free", 200, "account", "docker_username", "authenticated", "free")
	if len(policies) != 2 {
		t.Fatalf("policy count = %d, want exactly the two documented Registry pull policies", len(policies))
	}
}

func assertDockerHubPullPolicy(t *testing.T, policies map[string]connsdk.RateLimitPolicy, id string, limit int, subjectKind, subjectConfig, authType, tier string) {
	t.Helper()
	policy, ok := policies[id]
	if !ok {
		t.Fatalf("missing policy %q", id)
	}
	if got, want := policy.Source.URL, "https://docs.docker.com/docker-hub/download-rate-limit/"; got != want {
		t.Fatalf("%s source = %q, want %q", id, got, want)
	}
	if got, want := policy.Source.RetrievedAt, "2026-08-08"; got != want {
		t.Fatalf("%s retrieved_at = %q, want %q", id, got, want)
	}
	if got, want := policy.Selector.Hosts, []string{"registry-1.docker.io"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("%s selector hosts = %v, want %v", id, got, want)
	}
	if got, want := policy.Selector.AuthTypes, []string{authType}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("%s selector auth_types = %v, want %v", id, got, want)
	}
	if tier == "" {
		if len(policy.Selector.Tiers) != 0 {
			t.Fatalf("%s selector tiers = %v, want none", id, policy.Selector.Tiers)
		}
	} else if got := policy.Selector.Tiers; len(got) != 1 || got[0] != tier {
		t.Fatalf("%s selector tiers = %v, want [%s]", id, got, tier)
	}
	if got, want := string(policy.Scope.SubjectKind), subjectKind; got != want {
		t.Fatalf("%s subject kind = %q, want %q", id, got, want)
	}
	if got, want := policy.Scope.SubjectConfig, subjectConfig; got != want {
		t.Fatalf("%s subject config = %q, want %q", id, got, want)
	}
	if got := policy.Budgets; len(got) != 1 || got[0].Limit == nil || *got[0].Limit != limit || got[0].WindowSeconds == nil || *got[0].WindowSeconds != 21600 {
		t.Fatalf("%s budgets = %+v, want %d fixed-window requests per 21600 seconds", id, got, limit)
	}
}
