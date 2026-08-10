package engine

import (
	"context"
	"errors"
	"fmt"
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
	coordinationIdentity connectors.CoordinationIdentity
	policies             []connsdk.RateLimitPolicy
	registry             *coordination.RateLimitRegistry
	coordinator          connsdk.BudgetCoordinator
	sleep                func(context.Context, time.Duration) error
}

func newRateLimitResolver(b Bundle, cfg connectors.RuntimeConfig) *rateLimitResolver {
	if b.RateLimits == nil || b.RateLimits.State != connsdk.RateLimitStateDeclared || len(b.RateLimits.Policies) == 0 {
		return nil
	}
	resolver := &rateLimitResolver{
		connector:            b.Name,
		config:               cfg.Config,
		coordinationIdentity: cfg.CoordinationIdentity,
		policies:             b.RateLimits.Policies,
	}
	switch cfg.RateBudgetBackend {
	case "", connsdk.RateBudgetBackendProcessLocal:
		resolver.registry = currentRateLimitRegistry()
		if resolver.registry == nil {
			resolver.coordinator = unavailableRateBudgetCoordinator{}
			return resolver
		}
		resolver.coordinator = resolver.registry
		resolver.sleep = resolver.registry.Sleep
	case connsdk.RateBudgetBackendRequireShared:
		if cfg.BudgetCoordinator == nil {
			resolver.coordinator = unavailableRateBudgetCoordinator{}
		} else {
			resolver.coordinator = cfg.BudgetCoordinator
		}
		resolver.sleep = sleepRateBudget
	default:
		// An unknown mode is not an invitation to silently choose local state.
		resolver.coordinator = unavailableRateBudgetCoordinator{}
	}
	return resolver
}

func (r *rateLimitResolver) requesterFor(base *connsdk.Requester, method, path string) (*connsdk.Requester, error) {
	if r == nil {
		return base, nil
	}
	matched := make([]resolvedRateLimitPolicy, 0, len(r.policies))
	for _, policy := range r.policies {
		if !rateLimitSelectorMatches(policy.Selector, method, path, r.config) {
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

	return r.withResolvedPolicies(base, matched, costHeader), nil
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
		if len(policy.Selector.Endpoints) != 0 || len(policy.Selector.ExcludeEndpoints) != 0 || !rateLimitSelectorMatches(policy.Selector, "", "", r.config) {
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
	return r.withResolvedPolicies(base, matched, costHeader), nil
}

func (r *rateLimitResolver) withResolvedPolicies(base *connsdk.Requester, matched []resolvedRateLimitPolicy, costHeader string) *connsdk.Requester {
	clone := *base
	clone.LeaseAdmission = resolvedRateBudgetAdmission{
		coordinator: r.coordinator,
		batch: connsdk.ReservationBatch{
			Policies: resolvedReservationPolicies(matched),
		},
		sleep: r.sleep,
	}
	// Preserve the existing local seam for callers and tests that inspect it.
	// Requester gives LeaseAdmission precedence, so these compatibility hooks do
	// not double-charge a batch or apply observations twice.
	if r.registry != nil {
		clone.Admission = resolvedRateLimitAdmission(matched)
		clone.Observer = resolvedRateLimitObserver(matched)
	}
	clone.RateLimitCostHeader = costHeader
	return &clone
}

type resolvedRateLimitPolicy struct {
	limiter     *coordination.RateLimiter
	reservation connsdk.ReservationPolicy
	costHeader  string
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
	costHeader, err := rateLimitCostHeader(policy)
	if err != nil {
		return resolvedRateLimitPolicy{}, fmt.Errorf("rate-limit policy %q: %w", policy.ID, err)
	}
	fingerprint, err := coordination.RateBudgetPolicyFingerprint(r.connector, policy.ID, policy.Budgets)
	if err != nil {
		return resolvedRateLimitPolicy{}, fmt.Errorf("rate-limit policy %q fingerprint: %w", policy.ID, err)
	}
	resolved := resolvedRateLimitPolicy{
		reservation: connsdk.ReservationPolicy{
			Key: connsdk.ReservationKey{
				PolicyFingerprint: fingerprint,
				Scope:             string(scope),
			},
			Budgets: policy.Budgets,
		},
		costHeader: costHeader,
	}
	if r.registry != nil {
		resolved.limiter = r.registry.Limiter(coordination.RateLimitKey{Connector: r.connector, PolicyID: policy.ID, Scope: scope}, policy.Budgets)
	}
	return resolved, nil
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
	for _, endpoint := range endpoints {
		if endpoint.Method == method && endpoint.Path == path {
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

func resolvedReservationPolicies(policies []resolvedRateLimitPolicy) []connsdk.ReservationPolicy {
	reservations := make([]connsdk.ReservationPolicy, 0, len(policies))
	for _, policy := range policies {
		reservations = append(reservations, policy.reservation)
	}
	return reservations
}

// resolvedRateBudgetAdmission turns one declaration-level policy match into a
// complete atomic coordinator batch. It is intentionally the only requester
// route that calls Decide, so a matching policy can never be charged ahead of
// another policy in the same logical send.
type resolvedRateBudgetAdmission struct {
	coordinator connsdk.BudgetCoordinator
	batch       connsdk.ReservationBatch
	sleep       func(context.Context, time.Duration) error
}

var _ connsdk.RateLimitLeaseAdmission = resolvedRateBudgetAdmission{}

func (a resolvedRateBudgetAdmission) Admit(ctx context.Context, _ connsdk.RateLimitRequest) (connsdk.RateBudgetLease, error) {
	if a.coordinator == nil {
		return "", &connsdk.RateBudgetRefusalError{Reason: connsdk.RateBudgetRefusalSharedCoordinatorUnavailable}
	}
	for {
		decision, err := a.coordinator.Decide(ctx, a.batch)
		if err != nil {
			return "", normalizeRateBudgetCoordinatorError(err)
		}
		if decision.Granted {
			if decision.Lease == "" {
				return "", &connsdk.RateBudgetRefusalError{Reason: connsdk.RateBudgetRefusalSharedCoordinatorUnavailable}
			}
			return decision.Lease, nil
		}
		wait := decision.Wait
		if wait <= 0 && !decision.NotBefore.IsZero() {
			wait = time.Until(decision.NotBefore)
		}
		if wait <= 0 {
			return "", &connsdk.RateBudgetRefusalError{
				Reason:    connsdk.RateBudgetRefusalNotBefore,
				NotBefore: decision.NotBefore,
			}
		}
		if deadline, ok := ctx.Deadline(); ok && !deadline.After(time.Now().Add(wait)) {
			return "", &connsdk.RateBudgetRefusalError{
				Reason:    connsdk.RateBudgetRefusalDeadlineTooShort,
				NotBefore: decision.NotBefore,
			}
		}
		sleep := a.sleep
		if sleep == nil {
			sleep = sleepRateBudget
		}
		if err := sleep(ctx, wait); err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return "", &connsdk.RateBudgetRefusalError{
					Reason:    connsdk.RateBudgetRefusalDeadlineTooShort,
					NotBefore: decision.NotBefore,
				}
			}
			return "", err
		}
	}
}

func (a resolvedRateBudgetAdmission) Finish(ctx context.Context, lease connsdk.RateBudgetLease, observation connsdk.CompletionObservation) error {
	if a.coordinator == nil {
		return &connsdk.RateBudgetRefusalError{Reason: connsdk.RateBudgetRefusalSharedCoordinatorUnavailable}
	}
	return normalizeRateBudgetCoordinatorError(a.coordinator.Finish(ctx, lease, observation))
}

type unavailableRateBudgetCoordinator struct{}

func (unavailableRateBudgetCoordinator) Decide(context.Context, connsdk.ReservationBatch) (connsdk.AdmissionDecision, error) {
	return connsdk.AdmissionDecision{}, &connsdk.RateBudgetRefusalError{Reason: connsdk.RateBudgetRefusalSharedCoordinatorUnavailable}
}

func (unavailableRateBudgetCoordinator) Finish(context.Context, connsdk.RateBudgetLease, connsdk.CompletionObservation) error {
	return &connsdk.RateBudgetRefusalError{Reason: connsdk.RateBudgetRefusalSharedCoordinatorUnavailable}
}

func normalizeRateBudgetCoordinatorError(err error) error {
	if err == nil {
		return nil
	}
	var refusal *connsdk.RateBudgetRefusalError
	if errors.As(err, &refusal) {
		return err
	}
	if coordination.IsSharedCoordinatorUnavailable(err) || coordination.IsSharedCoordinatorEpochMismatch(err) {
		return &connsdk.RateBudgetRefusalError{Reason: connsdk.RateBudgetRefusalSharedCoordinatorUnavailable}
	}
	return err
}

func sleepRateBudget(ctx context.Context, wait time.Duration) error {
	if wait <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
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
