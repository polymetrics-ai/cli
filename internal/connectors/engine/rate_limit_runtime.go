package engine

import (
	"context"
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
	report               *connectors.RateLimitReport
}

func newRateLimitResolver(b Bundle, cfg connectors.RuntimeConfig) *rateLimitResolver {
	if cfg.RateLimitReport != nil {
		cfg.RateLimitReport.Declare(b.Name, rateLimitDeclaration(b.RateLimits))
	}
	if b.RateLimits == nil || b.RateLimits.State != connsdk.RateLimitStateDeclared || len(b.RateLimits.Policies) == 0 {
		return nil
	}
	return &rateLimitResolver{
		connector:            b.Name,
		config:               cfg.Config,
		coordinationIdentity: cfg.CoordinationIdentity,
		policies:             b.RateLimits.Policies,
		registry:             currentRateLimitRegistry(),
		report:               cfg.RateLimitReport,
	}
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
		if len(policy.Selector.Endpoints) != 0 || !rateLimitSelectorMatches(policy.Selector, "", "", r.config) {
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
	limiter         *coordination.RateLimiter
	costHeader      string
	id              string
	subjectKind     string
	selectionReason string
	report          *connectors.RateLimitReport
	connector       string
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
		limiter:         r.registry.Limiter(coordination.RateLimitKey{Connector: r.connector, PolicyID: policy.ID, Scope: scope}, policy.Budgets),
		costHeader:      costHeader,
		id:              policy.ID,
		subjectKind:     string(policy.Scope.SubjectKind),
		selectionReason: rateLimitSelectionReason(policy.Selector),
		report:          r.report,
		connector:       r.connector,
	}, nil
}

func rateLimitDeclaration(limits *connsdk.RateLimits) connectors.RateLimitDeclaration {
	if limits == nil {
		return connectors.RateLimitDeclarationUndeclared
	}
	switch limits.State {
	case connsdk.RateLimitStateDeclared:
		return connectors.RateLimitDeclarationDeclared
	case connsdk.RateLimitStateUnknown:
		return connectors.RateLimitDeclarationUnknown
	case connsdk.RateLimitStateNotApplicable:
		return connectors.RateLimitDeclarationNotApplicable
	default:
		return connectors.RateLimitDeclarationUndeclared
	}
}

func rateLimitSelectionReason(selector connsdk.RateLimitSelector) string {
	if selector.All {
		return "all"
	}
	reasons := make([]string, 0, 3)
	if len(selector.Endpoints) > 0 {
		reasons = append(reasons, "endpoint")
	}
	if len(selector.Tiers) > 0 {
		reasons = append(reasons, "tier")
	}
	if len(selector.AuthTypes) > 0 {
		reasons = append(reasons, "auth_type")
	}
	if len(reasons) == 0 {
		return "default"
	}
	return strings.Join(reasons, "+")
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
		endpointMatch := false
		for _, endpoint := range selector.Endpoints {
			if endpoint.Method == method && endpoint.Path == path {
				endpointMatch = true
				break
			}
		}
		if !endpointMatch {
			return false
		}
	}
	if len(selector.Tiers) > 0 && !rateLimitSelectorValueMatches(selector.Tiers, cfg["tier"]) {
		return false
	}
	return len(selector.AuthTypes) == 0 || rateLimitSelectorValueMatches(selector.AuthTypes, cfg["auth_type"])
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
		policy.report.RecordPolicySelection(policy.connector, policy.id, policy.subjectKind, policy.selectionReason)
		if err := policy.limiter.AdmitWithWaitObserver(ctx, request, func(wait time.Duration) {
			policy.report.RecordPacingWait(policy.connector, wait)
		}); err != nil {
			return err
		}
	}
	return nil
}

type resolvedRateLimitObserver []resolvedRateLimitPolicy

func (o resolvedRateLimitObserver) Observe(ctx context.Context, observation connsdk.RateLimitObservation) {
	for _, policy := range o {
		policy.limiter.Observe(ctx, observation)
		policy.report.RecordProviderObservation(policy.connector, policy.id, observation)
	}
}

func AttachRateLimitActivityObserver(requester *connsdk.Requester, cfg connectors.RuntimeConfig, connector string) {
	if requester == nil || cfg.RateLimitReport == nil || connector == "" {
		return
	}
	requester.ActivityObserver = rateLimitActivityReporter{report: cfg.RateLimitReport, connector: connector}
}

type rateLimitActivityReporter struct {
	report    *connectors.RateLimitReport
	connector string
}

var _ connsdk.RateLimitActivityObserver = rateLimitActivityReporter{}

func (r rateLimitActivityReporter) ObserveProviderRateLimit(_ context.Context, observation connsdk.RateLimitObservation) {
	if observation.Status == 429 {
		r.report.RecordProvider429Observed(r.connector)
	}
}

func (r rateLimitActivityReporter) ObserveProviderRateLimitWait(_ context.Context, observation connsdk.RateLimitObservation, wait time.Duration, honored bool) {
	if observation.Status == 429 {
		r.report.RecordProvider429Wait(r.connector, wait, honored)
	}
}

func (r rateLimitActivityReporter) ObserveRequestLatency(_ context.Context, latency time.Duration) {
	r.report.RecordRequestLatency(r.connector, latency)
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
