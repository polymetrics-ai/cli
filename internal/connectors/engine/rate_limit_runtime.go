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

var processSharedRateLimitRegistry = struct {
	mu       sync.RWMutex
	registry *coordination.SharedRateLimitRegistry
}{}

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

// ConfigureSharedRateLimitRegistry makes an optional coordinator available to
// explicitly require-shared policies. Local policies never consult it, so a
// configured runtime endpoint cannot silently upgrade their enforcement.
func ConfigureSharedRateLimitRegistry(registry *coordination.SharedRateLimitRegistry) {
	processSharedRateLimitRegistry.mu.Lock()
	processSharedRateLimitRegistry.registry = registry
	processSharedRateLimitRegistry.mu.Unlock()
}

func currentSharedRateLimitRegistry() *coordination.SharedRateLimitRegistry {
	processSharedRateLimitRegistry.mu.RLock()
	defer processSharedRateLimitRegistry.mu.RUnlock()
	return processSharedRateLimitRegistry.registry
}

func replaceSharedRateLimitRegistryForTest(registry *coordination.SharedRateLimitRegistry) func() {
	processSharedRateLimitRegistry.mu.Lock()
	previous := processSharedRateLimitRegistry.registry
	processSharedRateLimitRegistry.registry = registry
	processSharedRateLimitRegistry.mu.Unlock()
	return func() {
		processSharedRateLimitRegistry.mu.Lock()
		processSharedRateLimitRegistry.registry = previous
		processSharedRateLimitRegistry.mu.Unlock()
	}
}

// rateLimitResolver is built once per engine runtime and resolves concrete
// declared request paths to opaque, shared policy limiters. It keeps only the
// connector configuration and CoordinationIdentity: never Secrets or a
// CredentialRevision. Raw scope subjects are passed straight into
// CoordinationIdentity and never stored in the registry.
type rateLimitResolver struct {
	ctx                  context.Context
	connector            string
	config               map[string]string
	coordinationIdentity connectors.CoordinationIdentity
	policies             []connsdk.RateLimitPolicy
	registry             *coordination.RateLimitRegistry
	sharedRegistry       *coordination.SharedRateLimitRegistry
}

func newRateLimitResolver(b Bundle, cfg connectors.RuntimeConfig) *rateLimitResolver {
	return newRateLimitResolverWithContext(context.Background(), b, cfg)
}

func newRateLimitResolverWithContext(ctx context.Context, b Bundle, cfg connectors.RuntimeConfig) *rateLimitResolver {
	if b.RateLimits == nil || b.RateLimits.State != connsdk.RateLimitStateDeclared || len(b.RateLimits.Policies) == 0 {
		return nil
	}
	return &rateLimitResolver{
		ctx:                  ctx,
		connector:            b.Name,
		config:               cfg.Config,
		coordinationIdentity: cfg.CoordinationIdentity,
		policies:             b.RateLimits.Policies,
		registry:             currentRateLimitRegistry(),
		sharedRegistry:       currentSharedRateLimitRegistry(),
	}
}

func (r *rateLimitResolver) requesterFor(base *connsdk.Requester, _, _ string) (*connsdk.Requester, error) {
	return r.defaultRequester(base)
}

func (r *rateLimitResolver) defaultRequester(base *connsdk.Requester) (*connsdk.Requester, error) {
	if r == nil {
		return base, nil
	}
	defaultPolicies, endpointPolicies := r.partitionPolicies()
	matched, costHeader, err := r.resolvePolicies(r.ctx, "", "", defaultPolicies)
	if err != nil {
		return nil, err
	}
	if len(matched) == 0 && len(endpointPolicies) == 0 {
		return base, nil
	}
	clone := *base
	if len(matched) > 0 {
		clone.Admission = resolvedRateLimitAdmission(matched)
		clone.Observer = resolvedRateLimitObserver(matched)
		clone.RateLimitCostHeader = costHeader
	}
	if len(endpointPolicies) > 0 {
		clone.RouteRateLimits = &endpointRateLimitResolver{resolver: r, policies: endpointPolicies}
	}
	return &clone, nil
}

func (r *rateLimitResolver) partitionPolicies() ([]connsdk.RateLimitPolicy, []connsdk.RateLimitPolicy) {
	defaultPolicies := make([]connsdk.RateLimitPolicy, 0, len(r.policies))
	endpointPolicies := make([]connsdk.RateLimitPolicy, 0, len(r.policies))
	for _, policy := range r.policies {
		if len(policy.Selector.Endpoints) != 0 || len(policy.Selector.ExcludeEndpoints) != 0 {
			endpointPolicies = append(endpointPolicies, policy)
			continue
		}
		defaultPolicies = append(defaultPolicies, policy)
	}
	return defaultPolicies, endpointPolicies
}

func (r *rateLimitResolver) resolvePolicies(ctx context.Context, method, path string, policies []connsdk.RateLimitPolicy) ([]resolvedRateLimitPolicy, string, error) {
	matched := make([]resolvedRateLimitPolicy, 0, len(policies))
	for _, policy := range policies {
		if !rateLimitSelectorMatches(policy.Selector, method, path, r.config) {
			continue
		}
		resolved, err := r.resolve(ctx, policy)
		if err != nil {
			return nil, "", err
		}
		matched = append(matched, resolved)
	}
	costHeader := ""
	for _, match := range matched {
		if match.costHeader == "" {
			continue
		}
		if costHeader != "" && !strings.EqualFold(costHeader, match.costHeader) {
			return nil, "", fmt.Errorf("matching rate-limit policies use different actual-cost headers")
		}
		costHeader = match.costHeader
	}
	return matched, costHeader, nil
}

type resolvedRateLimitPolicy struct {
	admission  connsdk.RateLimitAdmission
	observer   connsdk.RateLimitObserver
	costHeader string
}

func (r *rateLimitResolver) resolve(ctx context.Context, policy connsdk.RateLimitPolicy) (resolvedRateLimitPolicy, error) {
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
	costHeader, err := rateLimitCostHeader(policy)
	if err != nil {
		return resolvedRateLimitPolicy{}, fmt.Errorf("rate-limit policy %q: %w", policy.ID, err)
	}
	key := coordination.RateLimitKey{Connector: r.connector, PolicyID: policy.ID, Scope: scope}
	if policy.Coordination == connsdk.RateLimitCoordinationRequireShared {
		if r.sharedRegistry == nil {
			return resolvedRateLimitPolicy{}, &coordination.SharedRateLimitUnavailableError{
				Component: "dragonfly",
				Reason:    coordination.SharedRateLimitCoordinatorNotConfigured,
			}
		}
		if err := r.sharedRegistry.EnsureAvailable(ctx); err != nil {
			return resolvedRateLimitPolicy{}, err
		}
		limiter := r.sharedRegistry.Limiter(key, policy.Budgets)
		return resolvedRateLimitPolicy{admission: limiter, observer: limiter, costHeader: costHeader}, nil
	}
	if r.registry == nil {
		return resolvedRateLimitPolicy{}, fmt.Errorf("rate-limit policy %q local registry is unavailable", policy.ID)
	}
	limiter := r.registry.Limiter(key, policy.Budgets)
	return resolvedRateLimitPolicy{admission: limiter, observer: limiter, costHeader: costHeader}, nil
}

type endpointRateLimitResolver struct {
	resolver *rateLimitResolver
	policies []connsdk.RateLimitPolicy
}

var _ connsdk.RateLimitRouteResolver = (*endpointRateLimitResolver)(nil)

func (r *endpointRateLimitResolver) AdmitRoute(ctx context.Context, route connsdk.RateLimitRoute) (string, error) {
	if r == nil || r.resolver == nil {
		return "", fmt.Errorf("rate-limit route resolver is unavailable")
	}
	matched, costHeader, err := r.resolver.resolvePolicies(ctx, strings.ToUpper(strings.TrimSpace(route.Method)), route.Path, r.policies)
	if err != nil {
		return "", err
	}
	if err := resolvedRateLimitAdmission(matched).Admit(ctx, connsdk.RateLimitRequest{Method: route.Method, Attempt: route.Attempt}); err != nil {
		return "", err
	}
	return costHeader, nil
}

func (r *endpointRateLimitResolver) ObserveRoute(ctx context.Context, route connsdk.RateLimitRoute, observation connsdk.RateLimitObservation) {
	if r == nil || r.resolver == nil {
		return
	}
	matched, _, err := r.resolver.resolvePolicies(ctx, strings.ToUpper(strings.TrimSpace(route.Method)), route.Path, r.policies)
	if err != nil {
		return
	}
	resolvedRateLimitObserver(matched).Observe(ctx, observation)
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

func rateLimitSelectorMatches(selector connsdk.RateLimitSelector, method, path string, cfg map[string]string) bool {
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
	if len(selector.Tiers) > 0 && !rateLimitSelectorValueMatches(selector.Tiers, cfg["tier"]) {
		return false
	}
	return len(selector.AuthTypes) == 0 || rateLimitSelectorValueMatches(selector.AuthTypes, cfg["auth_type"])
}

func rateLimitEndpointMatches(endpoints []connsdk.RateLimitEndpointSelector, method, path string) bool {
	method = strings.ToUpper(strings.TrimSpace(method))
	path = rateLimitSelectorPath(path)
	for _, endpoint := range endpoints {
		if endpoint.Method == method && rateLimitEndpointPathMatches(endpoint.Path, path) {
			return true
		}
	}
	return false
}

func rateLimitSelectorPath(path string) string {
	parsed, err := url.Parse(path)
	if err == nil {
		return parsed.Path
	}
	path, _, _ = strings.Cut(path, "?")
	return path
}

func rateLimitEndpointPathMatches(pattern, path string) bool {
	if pattern == path {
		return true
	}
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")
	if len(patternParts) != len(pathParts) {
		return false
	}
	for i, patternPart := range patternParts {
		if rateLimitEndpointPathParameter(patternPart) {
			if pathParts[i] != "" {
				continue
			}
			return false
		}
		if patternPart != pathParts[i] {
			return false
		}
	}
	return true
}

func rateLimitEndpointPathParameter(part string) bool {
	return len(part) > 2 && strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") && !strings.ContainsAny(part[1:len(part)-1], "{}")
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
		if err := policy.admission.Admit(ctx, request); err != nil {
			return err
		}
	}
	return nil
}

type resolvedRateLimitObserver []resolvedRateLimitPolicy

func (o resolvedRateLimitObserver) Observe(ctx context.Context, observation connsdk.RateLimitObservation) {
	for _, policy := range o {
		policy.observer.Observe(ctx, observation)
	}
}

// RequesterFor returns a requester governed by declared rate-limit policies.
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
