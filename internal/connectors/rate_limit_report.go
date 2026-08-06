package connectors

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors/connsdk"
)

// RateLimitDeclaration describes what the connector bundle declares about
// provider limits. Undeclared deliberately means no rate_limits.json exists;
// it never means the provider permits unlimited traffic.
type RateLimitDeclaration string

const (
	RateLimitDeclarationDeclared      RateLimitDeclaration = "declared"
	RateLimitDeclarationUnknown       RateLimitDeclaration = "unknown"
	RateLimitDeclarationNotApplicable RateLimitDeclaration = "not_applicable"
	RateLimitDeclarationUndeclared    RateLimitDeclaration = "undeclared"
)

// RateLimitSummary is the bounded, secret-free rate-limit result attached to
// an ETL run. It contains declaration identifiers and typed provider facts
// only; it never accepts credentials, bindings, scope keys, or runtime subject
// values.
type RateLimitSummary struct {
	Connectors []RateLimitConnectorSummary `json:"connectors"`
}

// RateLimitConnectorSummary separates the three causes of a slow run:
// local pacing, provider 429 retry waits, and ordinary request latency.
type RateLimitConnectorSummary struct {
	Connector           string                   `json:"connector,omitempty"`
	Declaration         RateLimitDeclaration     `json:"declaration"`
	Policies            []RateLimitPolicySummary `json:"policies"`
	PoliciesOmitted     int                      `json:"policies_omitted,omitempty"`
	PacingWaitMS        int64                    `json:"pacing_wait_ms"`
	Provider429Observed int                      `json:"provider_429_observed"`
	Provider429Honored  int                      `json:"provider_429_honored"`
	ProviderWaitMS      int64                    `json:"provider_wait_ms"`
	RequestLatencyMS    int64                    `json:"request_latency_ms"`
	RequestCount        int                      `json:"request_count"`
}

// RateLimitPolicySummary contains only a declared policy ID, a declared
// subject kind, a structural selector reason, and typed provider budget facts.
// In particular, SubjectKind is not its runtime subject value.
type RateLimitPolicySummary struct {
	ID                string     `json:"id"`
	SubjectKind       string     `json:"subject_kind"`
	SelectionReason   string     `json:"selection_reason"`
	ProviderLimit     *int64     `json:"provider_limit,omitempty"`
	ProviderRemaining *int64     `json:"provider_remaining,omitempty"`
	ProviderResetAt   *time.Time `json:"provider_reset_at,omitempty"`
}

const maxRateLimitReportPolicies = 16

// RateLimitReport collects one bounded summary for an execution. It is safe
// for concurrent requester callbacks and intentionally exposes narrow typed
// recorders instead of a generic event payload.
type RateLimitReport struct {
	mu         sync.Mutex
	connectors map[string]*rateLimitReportConnector
}

type rateLimitReportConnector struct {
	declaration         RateLimitDeclaration
	policies            map[string]*rateLimitReportPolicy
	omittedPolicyIDs    map[string]struct{}
	policiesOmitted     int
	pacingWaitMS        int64
	provider429Observed int
	provider429Honored  int
	providerWaitMS      int64
	requestLatencyMS    int64
	requestCount        int
}

type rateLimitReportPolicy struct {
	id                string
	subjectKind       string
	selectionReason   string
	providerLimit     *int64
	providerRemaining *int64
	providerResetAt   *time.Time
}

// NewRateLimitReport creates an empty run report. Connectors are registered by
// the engine as their bundle is entered so a missing declaration can be
// reported honestly as undeclared.
func NewRateLimitReport() *RateLimitReport {
	return &RateLimitReport{connectors: make(map[string]*rateLimitReportConnector)}
}

// Declare records one bundle's declaration state. Unknown input is treated as
// undeclared rather than suggesting a permissive policy.
func (r *RateLimitReport) Declare(connector string, declaration RateLimitDeclaration) {
	if r == nil || connector == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connector(connector).declaration = normalizeRateLimitDeclaration(declaration)
}

// RecordPolicySelection records a selected declared policy once. Its arguments
// are declaration metadata only: callers must never pass a runtime subject,
// opaque scope, binding, credential, or approval revision.
func (r *RateLimitReport) RecordPolicySelection(connector, policyID, subjectKind, selectionReason string) {
	if r == nil || connector == "" || policyID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	c := r.connector(connector)
	if _, ok := c.policies[policyID]; ok {
		return
	}
	if len(c.policies) >= maxRateLimitReportPolicies {
		c.omittedPolicyIDs[policyID] = struct{}{}
		c.policiesOmitted = len(c.omittedPolicyIDs)
		return
	}
	c.policies[policyID] = &rateLimitReportPolicy{
		id:              policyID,
		subjectKind:     subjectKind,
		selectionReason: selectionReason,
	}
}

// RecordPacingWait adds time spent in the local declared-policy limiter.
func (r *RateLimitReport) RecordPacingWait(connector string, wait time.Duration) {
	if r == nil || connector == "" || wait <= 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connector(connector).pacingWaitMS += wait.Milliseconds()
}

// RecordProviderObservation retains only typed provider budget facts for a
// policy already selected in this execution.
func (r *RateLimitReport) RecordProviderObservation(connector, policyID string, observation connsdk.RateLimitObservation) {
	if r == nil || connector == "" || policyID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	policy := r.connector(connector).policies[policyID]
	if policy == nil {
		return
	}
	if observation.HasLimit {
		policy.providerLimit = int64Pointer(observation.Limit)
	}
	if observation.HasRemaining {
		policy.providerRemaining = int64Pointer(observation.Remaining)
	}
	if observation.HasReset {
		resetAt := observation.ResetAt.UTC()
		policy.providerResetAt = &resetAt
	}
}

// RecordProvider429Observed notes a provider response; no response body,
// header value, URL, or credential-derived data is retained.
func (r *RateLimitReport) RecordProvider429Observed(connector string) {
	if r == nil || connector == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connector(connector).provider429Observed++
}

// RecordProvider429Wait records a completed provider-directed retry wait. The
// supplied duration is the exact scheduled wait, which requester retry logic
// honours without capping when the provider provides a reset.
func (r *RateLimitReport) RecordProvider429Wait(connector string, wait time.Duration, honored bool) {
	if r == nil || connector == "" || !honored {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	c := r.connector(connector)
	c.provider429Honored++
	if wait > 0 {
		c.providerWaitMS += wait.Milliseconds()
	}
}

// RecordRequestLatency adds response latency separately from rate-limit waits.
func (r *RateLimitReport) RecordRequestLatency(connector string, latency time.Duration) {
	if r == nil || connector == "" || latency < 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	c := r.connector(connector)
	c.requestCount++
	c.requestLatencyMS += latency.Milliseconds()
}

// Snapshot returns a stable copy suitable for human or structured output.
func (r *RateLimitReport) Snapshot() RateLimitSummary {
	if r == nil {
		return RateLimitSummary{Connectors: []RateLimitConnectorSummary{}}
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	names := make([]string, 0, len(r.connectors))
	for name := range r.connectors {
		names = append(names, name)
	}
	sort.Strings(names)
	summary := RateLimitSummary{Connectors: make([]RateLimitConnectorSummary, 0, len(names))}
	for _, name := range names {
		c := r.connectors[name]
		policyIDs := make([]string, 0, len(c.policies))
		for id := range c.policies {
			policyIDs = append(policyIDs, id)
		}
		sort.Strings(policyIDs)
		policies := make([]RateLimitPolicySummary, 0, len(policyIDs))
		for _, id := range policyIDs {
			p := c.policies[id]
			policies = append(policies, RateLimitPolicySummary{
				ID:                p.id,
				SubjectKind:       p.subjectKind,
				SelectionReason:   p.selectionReason,
				ProviderLimit:     cloneInt64Pointer(p.providerLimit),
				ProviderRemaining: cloneInt64Pointer(p.providerRemaining),
				ProviderResetAt:   cloneTimePointer(p.providerResetAt),
			})
		}
		summary.Connectors = append(summary.Connectors, RateLimitConnectorSummary{
			Connector:           name,
			Declaration:         c.declaration,
			Policies:            policies,
			PoliciesOmitted:     c.policiesOmitted,
			PacingWaitMS:        c.pacingWaitMS,
			Provider429Observed: c.provider429Observed,
			Provider429Honored:  c.provider429Honored,
			ProviderWaitMS:      c.providerWaitMS,
			RequestLatencyMS:    c.requestLatencyMS,
			RequestCount:        c.requestCount,
		})
	}
	return summary
}

func (s RateLimitSummary) Normalized() RateLimitSummary {
	if len(s.Connectors) == 0 {
		return RateLimitSummary{Connectors: []RateLimitConnectorSummary{{
			Declaration: RateLimitDeclarationUndeclared,
			Policies:    []RateLimitPolicySummary{},
		}}}
	}
	connectors := make([]RateLimitConnectorSummary, len(s.Connectors))
	for i, connector := range s.Connectors {
		connector.Declaration = normalizeRateLimitDeclaration(connector.Declaration)
		if connector.Policies == nil {
			connector.Policies = []RateLimitPolicySummary{}
		}
		connectors[i] = connector
	}
	return RateLimitSummary{Connectors: connectors}
}

// HumanLines renders the same bounded, secret-free values as Snapshot for the
// CLI's text surface. It does not accept arbitrary request data, and therefore
// cannot introduce credentials, bindings, runtime subjects, scope values, or
// credential revisions into operator output.
func (s RateLimitSummary) HumanLines() []string {
	s = s.Normalized()
	lines := make([]string, 0, len(s.Connectors))
	for _, connector := range s.Connectors {
		connectorName := ""
		if connector.Connector != "" {
			connectorName = " connector=" + connector.Connector
		}
		lines = append(lines, fmt.Sprintf("Rate limits:%s declaration=%s policies=%s local_pacing_wait=%dms provider_429_observed=%d provider_429_honored=%d provider_429_wait=%dms request_latency=%dms requests=%d",
			connectorName,
			connector.Declaration,
			formatRateLimitPolicies(connector),
			connector.PacingWaitMS,
			connector.Provider429Observed,
			connector.Provider429Honored,
			connector.ProviderWaitMS,
			connector.RequestLatencyMS,
			connector.RequestCount,
		))
	}
	return lines
}

func formatRateLimitPolicies(connector RateLimitConnectorSummary) string {
	if len(connector.Policies) == 0 {
		return "none"
	}
	policies := make([]string, 0, len(connector.Policies)+1)
	for _, policy := range connector.Policies {
		parts := []string{
			policy.ID,
			"subject_kind=" + policy.SubjectKind,
			"selected_by=" + policy.SelectionReason,
		}
		if policy.ProviderLimit != nil {
			parts = append(parts, fmt.Sprintf("provider_limit=%d", *policy.ProviderLimit))
		}
		if policy.ProviderRemaining != nil {
			parts = append(parts, fmt.Sprintf("provider_remaining=%d", *policy.ProviderRemaining))
		}
		if policy.ProviderResetAt != nil {
			parts = append(parts, "provider_reset_at="+policy.ProviderResetAt.UTC().Format(time.RFC3339))
		}
		policies = append(policies, "("+strings.Join(parts, ",")+")")
	}
	if connector.PoliciesOmitted > 0 {
		policies = append(policies, fmt.Sprintf("(+%d policies omitted)", connector.PoliciesOmitted))
	}
	return strings.Join(policies, ";")
}

func (r *RateLimitReport) connector(name string) *rateLimitReportConnector {
	c := r.connectors[name]
	if c == nil {
		c = &rateLimitReportConnector{
			declaration:      RateLimitDeclarationUndeclared,
			policies:         make(map[string]*rateLimitReportPolicy),
			omittedPolicyIDs: make(map[string]struct{}),
		}
		r.connectors[name] = c
	}
	return c
}

func normalizeRateLimitDeclaration(declaration RateLimitDeclaration) RateLimitDeclaration {
	switch declaration {
	case RateLimitDeclarationDeclared, RateLimitDeclarationUnknown, RateLimitDeclarationNotApplicable, RateLimitDeclarationUndeclared:
		return declaration
	default:
		return RateLimitDeclarationUndeclared
	}
}

func int64Pointer(value int64) *int64 {
	return &value
}

func cloneInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}
	return int64Pointer(*value)
}

func cloneTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
