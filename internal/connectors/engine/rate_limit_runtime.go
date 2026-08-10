package engine

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/coordination"
)

var processRateLimitRegistry = struct {
	mu       sync.RWMutex
	registry *coordination.RateLimitRegistry
}{registry: coordination.NewRateLimitRegistry(nil)}

func currentRateLimitRegistry() *coordination.RateLimitRegistry {
	processRateLimitRegistry.mu.RLock()
	defer processRateLimitRegistry.mu.RUnlock()
	return processRateLimitRegistry.registry
}

// replaceRateLimitRegistryForTest swaps the process-local registry so engine
// tests can inject a clock without sleeping. It is intentionally package
// private: runtime callers must never select a different registry ad hoc.
func replaceRateLimitRegistryForTest(registry *coordination.RateLimitRegistry) func() {
	processRateLimitRegistry.mu.Lock()
	previous := processRateLimitRegistry.registry
	processRateLimitRegistry.registry = registry
	processRateLimitRegistry.mu.Unlock()
	return func() {
		processRateLimitRegistry.mu.Lock()
		processRateLimitRegistry.registry = previous
		processRateLimitRegistry.mu.Unlock()
	}
}

// rateLimitResolver is built once per engine runtime and resolves concrete
// declared request paths to opaque, shared policy limiters. It keeps only the
// connector configuration and CoordinationIdentity: never Secrets or a
// CredentialRevision. Raw scope subjects are passed straight into
// CoordinationIdentity and never stored in the registry.
type rateLimitResolver struct {
	connector            string
	config               map[string]string
	host                 string
	coordinationIdentity connectors.CoordinationIdentity
	policies             []connsdk.RateLimitPolicy
	registry             *coordination.RateLimitRegistry
}

func newRateLimitResolver(b Bundle, cfg connectors.RuntimeConfig, resolvedBaseURLs ...string) *rateLimitResolver {
	if b.RateLimits == nil || b.RateLimits.State != connsdk.RateLimitStateDeclared || len(b.RateLimits.Policies) == 0 {
		return nil
	}
	baseURL := cfg.Config["base_url"]
	if len(resolvedBaseURLs) > 0 {
		baseURL = resolvedBaseURLs[0]
	}
	return &rateLimitResolver{
		connector:            b.Name,
		config:               cfg.Config,
		host:                 rateLimitRequestHost(baseURL),
		coordinationIdentity: cfg.CoordinationIdentity,
		policies:             b.RateLimits.Policies,
		registry:             currentRateLimitRegistry(),
	}
}

func (r *rateLimitResolver) requesterFor(base *connsdk.Requester, method, path string) (*connsdk.Requester, error) {
	if r == nil {
		return base, nil
	}
	matched := make([]resolvedRateLimitPolicy, 0, len(r.policies))
	for _, policy := range r.policies {
		if !rateLimitSelectorMatches(policy.Selector, method, path, r.host, r.config) {
			continue
		}
		resolved, err := r.resolve(policy)
		if err != nil {
			return nil, err
		}
		matched = append(matched, resolved)
	}
	if len(matched) == 0 {
		return base, nil
	}

	costHeader := ""
	for _, match := range matched {
		if match.costHeader == "" {
			continue
		}
		if costHeader != "" && !strings.EqualFold(costHeader, match.costHeader) {
			return nil, fmt.Errorf("rate-limit policies matching %s %s use different actual-cost headers", method, path)
		}
		costHeader = match.costHeader
	}

	clone := *base
	clone.Admission = resolvedRateLimitAdmission(matched)
	clone.Observer = resolvedRateLimitObserver(matched)
	clone.RateLimitCostHeader = costHeader
	return &clone, nil
}

// defaultRequester attaches policies that do not need endpoint matching. Hooks
// receive Runtime.Requester directly, so whole-connector and tier/auth-only
// policies must be present there as well as on the concrete declarative paths.
func (r *rateLimitResolver) defaultRequester(base *connsdk.Requester) (*connsdk.Requester, error) {
	if r == nil {
		return base, nil
	}
	matched := make([]resolvedRateLimitPolicy, 0, len(r.policies))
	for _, policy := range r.policies {
		if len(policy.Selector.Endpoints) != 0 || len(policy.Selector.ExcludeEndpoints) != 0 || !rateLimitSelectorMatches(policy.Selector, "", "", r.host, r.config) {
			continue
		}
		resolved, err := r.resolve(policy)
		if err != nil {
			return nil, err
		}
		matched = append(matched, resolved)
	}
	if len(matched) == 0 {
		return base, nil
	}
	costHeader := ""
	for _, match := range matched {
		if match.costHeader == "" {
			continue
		}
		if costHeader != "" && !strings.EqualFold(costHeader, match.costHeader) {
			return nil, fmt.Errorf("whole-connector rate-limit policies use different actual-cost headers")
		}
		costHeader = match.costHeader
	}
	clone := *base
	clone.Admission = resolvedRateLimitAdmission(matched)
	clone.Observer = resolvedRateLimitObserver(matched)
	clone.RateLimitCostHeader = costHeader
	return &clone, nil
}

type resolvedRateLimitPolicy struct {
	limiter    *coordination.RateLimiter
	costHeader string
}

func (r *rateLimitResolver) resolve(policy connsdk.RateLimitPolicy) (resolvedRateLimitPolicy, error) {
	kind, err := coordinationRateScopeKind(policy.Scope.SubjectKind)
	if err != nil {
		return resolvedRateLimitPolicy{}, err
	}
	subject := r.config[policy.Scope.SubjectConfig]
	if strings.TrimSpace(subject) == "" {
		return resolvedRateLimitPolicy{}, fmt.Errorf("rate-limit policy %q requires non-secret config %q for its declared scope", policy.ID, policy.Scope.SubjectConfig)
	}
	scope, err := r.coordinationIdentity.RateScopeKey(connectors.RateLimitScope{
		PolicyID: policy.ID,
		Kind:     kind,
		Subject:  subject,
	})
	if err != nil {
		return resolvedRateLimitPolicy{}, fmt.Errorf("rate-limit policy %q scope identity: %w", policy.ID, err)
	}
	if r.registry == nil {
		return resolvedRateLimitPolicy{}, fmt.Errorf("rate-limit policy %q local registry is unavailable", policy.ID)
	}
	costHeader, err := rateLimitCostHeader(policy)
	if err != nil {
		return resolvedRateLimitPolicy{}, fmt.Errorf("rate-limit policy %q: %w", policy.ID, err)
	}
	return resolvedRateLimitPolicy{
		limiter:    r.registry.Limiter(coordination.RateLimitKey{Connector: r.connector, PolicyID: policy.ID, Scope: scope}, policy.Budgets),
		costHeader: costHeader,
	}, nil
}

func coordinationRateScopeKind(subject connsdk.RateLimitScopeSubjectKind) (connectors.RateScopeKind, error) {
	switch subject {
	case connsdk.RateLimitScopeAccount:
		return connectors.RateScopeKindAccount, nil
	case connsdk.RateLimitScopeInstallation:
		return connectors.RateScopeKindInstallation, nil
	case connsdk.RateLimitScopeApplication:
		return connectors.RateScopeKindApplication, nil
	case connsdk.RateLimitScopeEndpoint:
		return connectors.RateScopeKindEndpointResource, nil
	case connsdk.RateLimitScopeIP:
		return connectors.RateScopeKindIP, nil
	default:
		return "", fmt.Errorf("rate-limit scope kind %q is unsupported", subject)
	}
}

func rateLimitSelectorMatches(selector connsdk.RateLimitSelector, method, path, host string, cfg map[string]string) bool {
	if selector.All {
		return true
	}
	if len(selector.Endpoints) > 0 {
		if !rateLimitEndpointMatches(selector.Endpoints, method, path) {
			return false
		}
	}
	if rateLimitEndpointMatches(selector.ExcludeEndpoints, method, path) {
		return false
	}
	if len(selector.Hosts) > 0 && !rateLimitSelectorHostMatches(selector.Hosts, host) {
		return false
	}
	if len(selector.Tiers) > 0 && !rateLimitSelectorValueMatches(selector.Tiers, cfg["tier"]) {
		return false
	}
	return len(selector.AuthTypes) == 0 || rateLimitSelectorValueMatches(selector.AuthTypes, cfg["auth_type"])
}

func rateLimitEndpointMatches(endpoints []connsdk.RateLimitEndpointSelector, method, path string) bool {
	for _, endpoint := range endpoints {
		if endpoint.Method == method && endpoint.Path == path {
			return true
		}
	}
	return false
}

func rateLimitRequestHost(baseURL string) string {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
}

func rateLimitSelectorHostMatches(hosts []string, host string) bool {
	for _, candidate := range hosts {
		if strings.EqualFold(candidate, host) {
			return true
		}
	}
	return false
}

func rateLimitSelectorValueMatches(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

type resolvedRateLimitAdmission []resolvedRateLimitPolicy

func (a resolvedRateLimitAdmission) Admit(ctx context.Context, request connsdk.RateLimitRequest) error {
	for _, policy := range a {
		if err := policy.limiter.Admit(ctx, request); err != nil {
			return err
		}
	}
	return nil
}

type resolvedRateLimitObserver []resolvedRateLimitPolicy

func (o resolvedRateLimitObserver) Observe(ctx context.Context, observation connsdk.RateLimitObservation) {
	for _, policy := range o {
		policy.limiter.Observe(ctx, observation)
	}
}

// RequesterFor returns a requester whose next logical requests are governed by
// every policy matching the declaration-level method and path. The input path
// is a bundle declaration, never a resolved runtime URL, so no runtime subject
// is added to the requester admission seam.
func (rt *Runtime) RequesterFor(method, path string) (*connsdk.Requester, error) {
	if rt == nil {
		return nil, fmt.Errorf("engine runtime requester is unavailable")
	}
	base := rt.baseRequester
	if base == nil {
		// Unit-tested hook/legacy callers can construct a Runtime literal with
		// only Requester. It has no resolver, so preserving that requester is
		// the exact pre-rate-limit behavior.
		base = rt.Requester
	}
	if base == nil {
		return nil, fmt.Errorf("engine runtime requester is unavailable")
	}
	return rt.rateLimits.requesterFor(base, strings.ToUpper(strings.TrimSpace(method)), path)
}

// requesterFor permits legacy internal call sites to preserve the old base
// requester when no resolver exists, while making missed policy attachment a
// testable error path rather than an accidental direct send.
func (rt *Runtime) requesterFor(method, path string) (*connsdk.Requester, error) {
	return rt.RequesterFor(method, path)
}
