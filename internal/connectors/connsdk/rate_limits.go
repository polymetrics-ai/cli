package connsdk

import "fmt"

// RateBudgetRefusalCode is a stable, machine-readable explanation for a
// request that was refused before provider dispatch.
type RateBudgetRefusalCode string

const (
	// RateBudgetRefusalSharedCoordinatorUnavailable means a require_shared
	// policy could not obtain the shared admission decision it declared.
	RateBudgetRefusalSharedCoordinatorUnavailable RateBudgetRefusalCode = "shared_coordinator_unavailable"
	// RateBudgetRefusalReservationDenied means the run-local lifecycle
	// coordinator did not grant an opaque lease before provider dispatch.
	RateBudgetRefusalReservationDenied RateBudgetRefusalCode = "reservation_denied"
)

// RateBudgetRefusalError is the SDK-facing fail-closed contract for rate
// budget admission. It carries only safe classifications; Err remains
// reachable for callers that need a more specific internal cause.
type RateBudgetRefusalError struct {
	Code   RateBudgetRefusalCode
	Reason string
	Err    error
}

func (e *RateBudgetRefusalError) Error() string {
	if e == nil {
		return "rate budget refused"
	}
	if e.Reason == "" {
		return fmt.Sprintf("rate budget refused: %s", e.Code)
	}
	return fmt.Sprintf("rate budget refused: %s: %s", e.Code, e.Reason)
}

func (e *RateBudgetRefusalError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// RateLimitState records whether a bundle has an enforceable provider policy.
// Unknown and not_applicable are deliberate, honest states; only declared
// policies are eligible for engine resolution and pacing.
type RateLimitState string

const (
	RateLimitStateDeclared      RateLimitState = "declared"
	RateLimitStateUnknown       RateLimitState = "unknown"
	RateLimitStateNotApplicable RateLimitState = "not_applicable"
)

// RateLimits is the typed form of an optional bundle rate_limits.json file.
// It contains no credential values or config templates. The engine resolves
// its declared selectors and scope subject kind without inferring scope
// identity from a secret.
type RateLimits struct {
	SchemaVersion int               `json:"schema_version"`
	State         RateLimitState    `json:"state"`
	Reason        string            `json:"reason,omitempty"`
	Policies      []RateLimitPolicy `json:"policies,omitempty"`
}

// RateLimitPolicy is one provider-cited rate-limit contract.
type RateLimitPolicy struct {
	ID           string                      `json:"id"`
	Source       RateLimitSource             `json:"source"`
	Selector     RateLimitSelector           `json:"selector"`
	Scope        RateLimitScope              `json:"scope"`
	Coordination RateLimitCoordinationPolicy `json:"coordination,omitempty"`
	Budgets      []RateLimitBudget           `json:"budgets"`
}

// RateLimitCoordinationPolicy controls where a declared policy coordinates.
// The zero value intentionally means process-local protection; shared
// coordination is never selected by endpoint configuration or inheritance.
type RateLimitCoordinationPolicy string

const (
	RateLimitCoordinationRequireShared RateLimitCoordinationPolicy = "require_shared"
)

// RateLimitSource records the provider artifact from which a policy was
// authored. RetrievedAt is mandatory so a reviewer can judge freshness even
// when the provider has no versioned documentation.
type RateLimitSource struct {
	URL         string `json:"url"`
	RetrievedAt string `json:"retrieved_at"`
	Version     string `json:"version,omitempty"`
}

// RateLimitSelector selects the traffic to which a policy applies. Selector
// dimensions compose with AND; endpoint entries are alternatives. All is an
// explicit whole-connector selector and cannot be combined with a narrower
// selector dimension.
type RateLimitSelector struct {
	All              bool                        `json:"all,omitempty"`
	Endpoints        []RateLimitEndpointSelector `json:"endpoints,omitempty"`
	ExcludeEndpoints []RateLimitEndpointSelector `json:"exclude_endpoints,omitempty"`
	Tiers            []string                    `json:"tiers,omitempty"`
	AuthTypes        []string                    `json:"auth_types,omitempty"`
}

// RateLimitEndpointSelector is a provider endpoint selector. Path is a
// connector-relative path pattern, never a complete URL or caller value.
type RateLimitEndpointSelector struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// RateLimitScopeSubjectKind names the non-secret provider subject that joins
// a policy ID and credential binding in a coordination key. It is
// intentionally a kind, not a value: raw credentials and secret-derived
// values must never enter a rate-limit key, registry, or event.
type RateLimitScopeSubjectKind string

const (
	RateLimitScopeAccount      RateLimitScopeSubjectKind = "account"
	RateLimitScopeInstallation RateLimitScopeSubjectKind = "installation"
	RateLimitScopeApplication  RateLimitScopeSubjectKind = "application"
	RateLimitScopeEndpoint     RateLimitScopeSubjectKind = "endpoint"
	RateLimitScopeIP           RateLimitScopeSubjectKind = "ip"
)

// RateLimitScope declares the one non-secret subject class for the policy.
// The engine resolves that class to an opaque registry scope key.
type RateLimitScope struct {
	SubjectKind RateLimitScopeSubjectKind `json:"subject_kind"`
	// SubjectConfig names the non-secret config property that supplies this
	// policy's runtime subject. The declaration contains only the property
	// name; the resolved value is passed directly to CoordinationIdentity and
	// is never retained by the registry or emitted in observations.
	SubjectConfig string `json:"subject_config"`
}

// RateLimitBudgetModel describes how a provider replenishes capacity.
type RateLimitBudgetModel string

const (
	RateLimitBudgetFixedWindow   RateLimitBudgetModel = "fixed_window"
	RateLimitBudgetSlidingWindow RateLimitBudgetModel = "sliding_window"
	RateLimitBudgetTokenBucket   RateLimitBudgetModel = "token_bucket"
	RateLimitBudgetLeakyBucket   RateLimitBudgetModel = "leaky_bucket"
)

// RateLimitBudgetDimension separates short burst ceilings from sustained
// provider limits. A policy can declare both without flattening them into one
// misleading requests-per-second number.
type RateLimitBudgetDimension string

const (
	RateLimitBudgetBurst     RateLimitBudgetDimension = "burst"
	RateLimitBudgetSustained RateLimitBudgetDimension = "sustained"
)

// RateLimitBudgetUnit names the budget currency. Points accommodates
// cost-weighted APIs such as GraphQL or leaky-bucket providers.
type RateLimitBudgetUnit string

const (
	RateLimitBudgetRequests RateLimitBudgetUnit = "requests"
	RateLimitBudgetPoints   RateLimitBudgetUnit = "points"
)

// RateLimitBudget is one independently enforced provider budget. Pointer
// fields preserve omitted-versus-zero information for loader validation:
// fixed/sliding windows require Limit and WindowSeconds; token/leaky buckets
// require Capacity and RestorePerSecond.
type RateLimitBudget struct {
	Model            RateLimitBudgetModel     `json:"model"`
	Dimension        RateLimitBudgetDimension `json:"dimension"`
	Unit             RateLimitBudgetUnit      `json:"unit"`
	Limit            *int                     `json:"limit,omitempty"`
	WindowSeconds    *int                     `json:"window_seconds,omitempty"`
	Capacity         *int                     `json:"capacity,omitempty"`
	RestorePerSecond *float64                 `json:"restore_per_second,omitempty"`
	Cost             *RateLimitCost           `json:"cost,omitempty"`
}

// RateLimitCost describes a cost-weighted request budget. DefaultCost applies
// when the provider gives no per-request observation; ResponseHeader names an
// optional provider header parsed into a typed scalar. ResponseBody is a
// closed vocabulary for a bounded, provider-declared response selection; it
// is never a caller-selected JSON path. Neither field carries a request value
// or credential.
type RateLimitCost struct {
	DefaultCost    *float64 `json:"default_cost,omitempty"`
	ResponseHeader string   `json:"response_header,omitempty"`
	ResponseBody   string   `json:"response_body,omitempty"`
}
