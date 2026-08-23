package engine

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

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

var processRateLimitEventSinks = struct {
	mu    sync.RWMutex
	sinks map[string]connsdk.RateLimitEventSink
}{sinks: map[string]connsdk.RateLimitEventSink{}}

var processRateLimitAdmissionTimeouts = struct {
	mu       sync.RWMutex
	timeouts map[string]time.Duration
}{timeouts: map[string]time.Duration{}}

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

// ConfigureRateLimitEventSink attaches a bounded rate-limit event sink to one
// project-local runtime. The project directory is an engine RuntimeConfig
// boundary, so concurrent certification runs cannot collect each other's
// events. The returned cleanup restores the prior registration.
func ConfigureRateLimitEventSink(projectDir string, sink connsdk.RateLimitEventSink) func() {
	if projectDir == "" || sink == nil {
		return func() {}
	}
	processRateLimitEventSinks.mu.Lock()
	previous, hadPrevious := processRateLimitEventSinks.sinks[projectDir]
	processRateLimitEventSinks.sinks[projectDir] = sink
	processRateLimitEventSinks.mu.Unlock()
	return func() {
		processRateLimitEventSinks.mu.Lock()
		if hadPrevious {
			processRateLimitEventSinks.sinks[projectDir] = previous
		} else {
			delete(processRateLimitEventSinks.sinks, projectDir)
		}
		processRateLimitEventSinks.mu.Unlock()
	}
}

func rateLimitEventSinkFor(projectDir string) connsdk.RateLimitEventSink {
	processRateLimitEventSinks.mu.RLock()
	defer processRateLimitEventSinks.mu.RUnlock()
	return processRateLimitEventSinks.sinks[projectDir]
}

// ConfigureRateLimitAdmissionTimeout bounds individual waits for a project
// runtime. It is deliberately narrower than a run deadline: normal
// certification work may continue, but an exhausted provider budget cannot
// hold the run indefinitely.
func ConfigureRateLimitAdmissionTimeout(projectDir string, timeout time.Duration) func() {
	if projectDir == "" || timeout <= 0 {
		return func() {}
	}
	processRateLimitAdmissionTimeouts.mu.Lock()
	previous, hadPrevious := processRateLimitAdmissionTimeouts.timeouts[projectDir]
	processRateLimitAdmissionTimeouts.timeouts[projectDir] = timeout
	processRateLimitAdmissionTimeouts.mu.Unlock()
	return func() {
		processRateLimitAdmissionTimeouts.mu.Lock()
		if hadPrevious {
			processRateLimitAdmissionTimeouts.timeouts[projectDir] = previous
		} else {
			delete(processRateLimitAdmissionTimeouts.timeouts, projectDir)
		}
		processRateLimitAdmissionTimeouts.mu.Unlock()
	}
}

func rateLimitAdmissionTimeoutFor(projectDir string) time.Duration {
	processRateLimitAdmissionTimeouts.mu.RLock()
	defer processRateLimitAdmissionTimeouts.mu.RUnlock()
	return processRateLimitAdmissionTimeouts.timeouts[projectDir]
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
	parking              connectors.RateParkingAdmission
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
		parking:              cfg.RateParkingAdmission,
	}
}

func (r *rateLimitResolver) requesterFor(base *connsdk.Requester, method, path string) (*connsdk.Requester, error) {
	requester, err := r.declaredRequester(base)
	if err != nil || r == nil {
		return requester, err
	}
	_, endpointPolicies := r.partitionPolicies()
	if len(endpointPolicies) == 0 {
		return requester, nil
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	path = strings.TrimSpace(path)
	if method == "" || !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("rate-limit requester requires a declared method and path")
	}
	// A hook must opt into this path with its bundle declaration before it can
	// make an endpoint-sensitive request. The normal resolver remains at the
	// physical send boundary so redirects and the final escaped path receive
	// their own admission/observation lifecycle.
	clone := *requester
	clone.RouteRateLimits = &endpointRateLimitResolver{resolver: r, policies: endpointPolicies}
	return &clone, nil
}

func (r *rateLimitResolver) defaultRequester(base *connsdk.Requester) (*connsdk.Requester, error) {
	if r == nil {
		return base, nil
	}
	defaultPolicies, endpointPolicies := r.partitionPolicies()
	if len(endpointPolicies) == 0 {
		return r.declaredRequester(base)
	}
	// Validate any whole-connector policy during runtime construction exactly as
	// before, but do not attach it to the default requester. connsdk invokes a
	// requester's generic Admission before its route resolver; attaching it here
	// would consume a default-policy lease before the endpoint guard could
	// reject an undeclared route in a mixed policy set.
	if _, _, err := r.resolvePolicies(r.ctx, "", "", defaultPolicies); err != nil {
		return nil, err
	}
	clone := *base
	// Runtime.Requester is retained for backwards-compatible whole-connector
	// policy callers, but it must not become an untracked endpoint-policy
	// escape hatch for a hook. Only Runtime.RequesterFor replaces this guard
	// after it receives the bundle-declared method/path.
	clone.RouteRateLimits = &undeclaredRouteRateLimitGuard{resolver: r, policies: endpointPolicies}
	return &clone, nil
}

// declaredRequester attaches whole-connector policies to a requester that has
// already identified its declaration through Runtime.RequesterFor. Endpoint
// policies remain at the physical send boundary so redirects and escaped paths
// receive their own admission and observation lifecycle.
func (r *rateLimitResolver) declaredRequester(base *connsdk.Requester) (*connsdk.Requester, error) {
	if r == nil {
		return base, nil
	}
	defaultPolicies, _ := r.partitionPolicies()
	matched, costHeader, err := r.resolvePolicies(r.ctx, "", "", defaultPolicies)
	if err != nil {
		return nil, err
	}
	if len(matched) == 0 {
		return base, nil
	}
	clone := *base
	clone.Admission = resolvedRateLimitAdmission(matched)
	clone.Observer = resolvedRateLimitObserver(matched)
	clone.RateLimitCostHeader = costHeader
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
	if len(matched) > 0 {
		scopes := make([]connectors.RateLimitScopeKey, 0, len(matched))
		for _, match := range matched {
			scopes = append(scopes, match.scope)
		}
		group, err := connectors.GroupRateLimitScopeKeys(scopes...)
		if err != nil {
			return nil, "", err
		}
		for index := range matched {
			matched[index].parkingScope = group
		}
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
	id           string
	admission    connsdk.RateLimitAdmission
	observer     connsdk.RateLimitObserver
	costHeader   string
	budgets      []connsdk.RateLimitBudget
	scope        connectors.RateLimitScopeKey
	parkingScope connectors.RateLimitScopeKey
	parking      connectors.RateParkingAdmission
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
		if err := r.sharedRegistry.EnsureAvailable(ctx); err != nil {
			return resolvedRateLimitPolicy{}, rateBudgetRefusal(err)
		}
		limiter := r.sharedRegistry.Limiter(key, policy.Budgets)
		return resolvedRateLimitPolicy{id: policy.ID, admission: limiter, observer: limiter, costHeader: costHeader, budgets: policy.Budgets, scope: scope, parking: r.parking}, nil
	}
	if r.registry == nil {
		return resolvedRateLimitPolicy{}, fmt.Errorf("rate-limit policy %q local registry is unavailable", policy.ID)
	}
	limiter := r.registry.Limiter(key, policy.Budgets)
	return resolvedRateLimitPolicy{id: policy.ID, admission: limiter, observer: limiter, costHeader: costHeader, budgets: policy.Budgets, scope: scope, parking: r.parking}, nil
}

// reservationBatch derives the coordinator input from the same selected,
// declaration-owned policies that the requester will admit at its physical
// send boundary. Scope keys and policy fingerprints are opaque; neither raw
// request material nor credentials cross this seam.
func (r *rateLimitResolver) reservationBatch(ctx context.Context, method, path string) (connsdk.ReservationBatch, error) {
	if r == nil {
		return connsdk.ReservationBatch{}, nil
	}
	matched, _, err := r.resolvePolicies(ctx, strings.ToUpper(strings.TrimSpace(method)), path, r.policies)
	if err != nil {
		return connsdk.ReservationBatch{}, err
	}
	batch := connsdk.ReservationBatch{Policies: make([]connsdk.ReservationPolicy, 0, len(matched))}
	for _, policy := range matched {
		fingerprint, err := coordination.RateBudgetPolicyFingerprint(r.connector, policy.id, policy.budgets)
		if err != nil {
			return connsdk.ReservationBatch{}, fmt.Errorf("rate-limit policy %q reservation fingerprint: %w", policy.id, err)
		}
		batch.Policies = append(batch.Policies, connsdk.ReservationPolicy{
			Key: connsdk.ReservationKey{
				PolicyFingerprint: fingerprint,
				Scope:             string(policy.scope),
			},
			Budgets: policy.budgets,
		})
	}
	return batch, nil
}

type endpointRateLimitResolver struct {
	resolver *rateLimitResolver
	policies []connsdk.RateLimitPolicy
}

// undeclaredRouteRateLimitGuard fails closed when code attempts an
// endpoint-sensitive request through Runtime.Requester directly. It does not
// resolve or admit any policy: doing either would let a mixed policy set
// partially consume a lease before the hook identified its declaration.
type undeclaredRouteRateLimitGuard struct {
	resolver *rateLimitResolver
	policies []connsdk.RateLimitPolicy
}

type responseFormatError struct {
	message string
	cause   error
}

func (e *responseFormatError) Error() string { return e.message }

func (e *responseFormatError) Is(target error) bool {
	if e == nil {
		return false
	}
	switch target {
	case context.Canceled, context.DeadlineExceeded:
		return errors.Is(e.cause, target)
	default:
		return false
	}
}

func (e *responseFormatError) As(target any) bool {
	if e == nil {
		return false
	}
	switch typed := target.(type) {
	case **connsdk.ProviderResponseError:
		var httpErr *connsdk.HTTPError
		if !errors.As(e.cause, &httpErr) {
			return false
		}
		*typed = &connsdk.ProviderResponseError{Status: httpErr.Status}
		return true
	case **connsdk.CredentialRejectedError:
		var httpErr *connsdk.HTTPError
		if !errors.As(e.cause, &httpErr) || httpErr.Status != http.StatusUnauthorized {
			return false
		}
		*typed = &connsdk.CredentialRejectedError{}
		return true
	case **coordination.SharedRateLimitUnavailableError:
		var cause *coordination.SharedRateLimitUnavailableError
		if !errors.As(e.cause, &cause) {
			return false
		}
		*typed = cause
		return true
	case **connsdk.RateBudgetRefusalError:
		var cause *connsdk.RateBudgetRefusalError
		if !errors.As(e.cause, &cause) {
			return false
		}
		*typed = cause
		return true
	default:
		return false
	}
}

func formatResponseError(message string, cause error) error {
	return &responseFormatError{message: message, cause: cause}
}

var _ connsdk.RateLimitRouteResolver = (*endpointRateLimitResolver)(nil)
var _ connsdk.RateLimitRouteResolver = (*undeclaredRouteRateLimitGuard)(nil)

func (r *undeclaredRouteRateLimitGuard) AdmitRoute(ctx context.Context, route connsdk.RateLimitRoute) (string, error) {
	if r == nil || r.resolver == nil {
		return "", fmt.Errorf("rate-limit route resolver is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	for _, policy := range r.policies {
		if rateLimitSelectorRequiresDeclaredRoute(policy.Selector, r.resolver.config) {
			return "", fmt.Errorf("rate-limit policy %q requires a declared method and path", policy.ID)
		}
	}
	return "", nil
}

// rateLimitSelectorRequiresDeclaredRoute reports whether selector needs the
// request method/path to decide whether it applies. Direct Runtime.Requester
// callers cannot safely make that choice, even if their physical path happens
// to look like a declared endpoint; Runtime.RequesterFor owns that boundary.
func rateLimitSelectorRequiresDeclaredRoute(selector connsdk.RateLimitSelector, cfg map[string]string) bool {
	if len(selector.Endpoints) == 0 && len(selector.ExcludeEndpoints) == 0 {
		return false
	}
	if len(selector.Tiers) > 0 && !rateLimitSelectorValueMatches(selector.Tiers, cfg["tier"]) {
		return false
	}
	return len(selector.AuthTypes) == 0 || rateLimitSelectorValueMatches(selector.AuthTypes, cfg["auth_type"])
}

func (*undeclaredRouteRateLimitGuard) ObserveRoute(context.Context, connsdk.RateLimitRoute, connsdk.RateLimitObservation) {
	// No route was admitted, so no observation may be attributed to a policy.
}

func (r *endpointRateLimitResolver) AdmitRoute(ctx context.Context, route connsdk.RateLimitRoute) (string, error) {
	if r == nil || r.resolver == nil {
		return "", fmt.Errorf("rate-limit route resolver is unavailable")
	}
	matched, costHeader, err := r.resolver.resolvePolicies(ctx, strings.ToUpper(strings.TrimSpace(route.Method)), route.Path, r.policies)
	if err != nil {
		return "", err
	}
	if len(matched) == 0 {
		for _, policy := range r.policies {
			if policy.Coordination != connsdk.RateLimitCoordinationRequireShared || !rateLimitSelectorMatchesNonRoute(policy.Selector, r.resolver.config) {
				continue
			}
			if err := ctx.Err(); err != nil {
				return "", err
			}
			return "", rateBudgetRefusal(&coordination.SharedRateLimitUnavailableError{Reason: coordination.SharedRateLimitRouteUnresolved})
		}
	}
	if err := resolvedRateLimitAdmission(matched).Admit(ctx, connsdk.RateLimitRequest{Method: route.Method, Attempt: route.Attempt}); err != nil {
		return "", rateBudgetRefusal(err)
	}
	return costHeader, nil
}

// rateLimitSelectorMatchesNonRoute distinguishes a require-shared policy that
// genuinely selected this credential/tier from one that belongs to another
// traffic family. A selected policy with no matching endpoint must fail closed;
// an unrelated GitHub certification overlay must not change ordinary traffic.
func rateLimitSelectorMatchesNonRoute(selector connsdk.RateLimitSelector, cfg map[string]string) bool {
	selector.Endpoints = nil
	selector.ExcludeEndpoints = nil
	return rateLimitSelectorMatches(selector, "", "", cfg)
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
		if escapedPath := parsed.EscapedPath(); escapedPath != "" {
			return escapedPath
		}
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
	admittedParking := make(map[connectors.RateLimitScopeKey]struct{}, len(a))
	for _, policy := range a {
		if policy.parking != nil {
			if _, done := admittedParking[policy.parkingScope]; !done {
				if err := policy.parking.Admit(policy.parkingScope); err != nil {
					return err
				}
				admittedParking[policy.parkingScope] = struct{}{}
			}
		}
		if err := policy.admission.Admit(ctx, request); err != nil {
			return rateBudgetRefusal(err)
		}
	}
	return nil
}

func rateBudgetRefusal(err error) error {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	var refusal *connsdk.RateBudgetRefusalError
	if errors.As(err, &refusal) {
		return err
	}
	var unavailable *coordination.SharedRateLimitUnavailableError
	if !errors.As(err, &unavailable) {
		return err
	}
	return &connsdk.RateBudgetRefusalError{
		Code:   connsdk.RateBudgetRefusalSharedCoordinatorUnavailable,
		Reason: string(unavailable.Reason),
		Err:    err,
	}
}

type resolvedRateLimitObserver []resolvedRateLimitPolicy

func (o resolvedRateLimitObserver) Observe(ctx context.Context, observation connsdk.RateLimitObservation) {
	for _, policy := range o {
		if !rateLimitPolicyAcceptsObservation(policy, observation) {
			continue
		}
		policy.observer.Observe(ctx, observation)
	}
}

// rateLimitPolicyAcceptsObservation keeps a GraphQL response's primary
// rateLimit selection out of the independently declared secondary point
// budget. GitHub's response body describes the primary cost, remaining, and
// reset time; applying any of those fields to the secondary policy would
// invent a provider signal and incorrectly block otherwise available traffic.
// Header observations retain their foundation behavior because no body family
// is involved.
func rateLimitPolicyAcceptsObservation(policy resolvedRateLimitPolicy, observation connsdk.RateLimitObservation) bool {
	if observation.CostSource != connsdk.RateLimitCostSourceGraphQLRateLimit {
		return true
	}
	for _, budget := range policy.budgets {
		if budget.Cost != nil && budget.Cost.ResponseBody == string(connsdk.RateLimitCostSourceGraphQLRateLimit) {
			return true
		}
	}
	return false
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
