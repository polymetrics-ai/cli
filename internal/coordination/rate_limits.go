// Package coordination holds ephemeral, opaque coordination state shared by
// connector executions. It never accepts credential material or raw subjects.
package coordination

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// RateLimitClock makes budget waits deterministic in tests and guarantees the
// registry's only wait path observes the caller's context.
type RateLimitClock interface {
	Now() time.Time
	Sleep(context.Context, time.Duration) error
}

type wallRateLimitClock struct{}

func (wallRateLimitClock) Now() time.Time { return time.Now() }

func (wallRateLimitClock) Sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// RateLimitKey is the only registry key shape. Scope is the opaque salted
// projection returned by connectors.CoordinationIdentity; no raw binding,
// credential, or subject is accepted or stored here.
type RateLimitKey struct {
	Connector string
	PolicyID  string
	Scope     connectors.RateLimitScopeKey
}

const (
	maxRateBudgetReservationPolicies    = 64
	defaultCompletionObservationHorizon = 2 * time.Minute
)

// RateLimitRegistry coordinates process-local policy budgets. It is
// intentionally ephemeral: cross-process coordination remains a separate
// capability and no local state claims account-wide protection.
type RateLimitRegistry struct {
	clock                        RateLimitClock
	mu                           sync.Mutex
	sets                         map[RateLimitKey]*rateLimitSet
	budgetSets                   map[rateBudgetKey]rateBudgetSet
	leases                       map[connsdk.RateBudgetLease]rateBudgetLease
	maxInFlight                  int
	leaseTTL                     time.Duration
	completionObservationHorizon time.Duration
}

// rateBudgetKey is private owner state for the batch seam. It contains only a
// policy fingerprint and the opaque #3863 scope key, never connector/policy
// names or any raw credential/account material.
type rateBudgetKey struct {
	policyFingerprint string
	scope             string
}

type rateBudgetSet struct {
	contractFingerprint string
	set                 *rateLimitSet
}

type rateBudgetLease struct {
	grantedAt time.Time
	expiresAt time.Time
	active    bool
	sets      []*rateLimitSet
}

// RateBudgetCoordinatorOptions bounds only the optional batch-decision seam.
// A non-positive MaxInFlight leaves the process-local registry's historic
// unlimited-concurrency behavior intact. LeaseTTL is deliberately short: a
// crashed or disconnected caller remains conservatively charged, but cannot
// hold the in-flight lease forever.
type RateBudgetCoordinatorOptions struct {
	MaxInFlight                  int
	LeaseTTL                     time.Duration
	CompletionObservationHorizon time.Duration
}

// RateBudgetCoordinator is the injectable decision/finish seam used by the
// Unix owner and focused tests. The ordinary local registry remains usable via
// Limiter without any socket or external dependency.
type RateBudgetCoordinator struct {
	registry *RateLimitRegistry
}

var _ connsdk.BudgetCoordinator = (*RateLimitRegistry)(nil)
var _ connsdk.BudgetCoordinator = (*RateBudgetCoordinator)(nil)

// NewRateLimitRegistry creates a local registry. A nil clock uses a
// context-aware wall clock; tests should inject a deterministic clock.
func NewRateLimitRegistry(clock RateLimitClock) *RateLimitRegistry {
	return newRateLimitRegistry(clock, RateBudgetCoordinatorOptions{})
}

func newRateLimitRegistry(clock RateLimitClock, options RateBudgetCoordinatorOptions) *RateLimitRegistry {
	if clock == nil {
		clock = wallRateLimitClock{}
	}
	if options.LeaseTTL <= 0 {
		options.LeaseTTL = 30 * time.Second
	}
	if options.CompletionObservationHorizon <= 0 {
		options.CompletionObservationHorizon = defaultCompletionObservationHorizon
	}
	if options.MaxInFlight < 0 {
		options.MaxInFlight = 0
	}
	return &RateLimitRegistry{
		clock:                        clock,
		sets:                         make(map[RateLimitKey]*rateLimitSet),
		budgetSets:                   make(map[rateBudgetKey]rateBudgetSet),
		leases:                       make(map[connsdk.RateBudgetLease]rateBudgetLease),
		maxInFlight:                  options.MaxInFlight,
		leaseTTL:                     options.LeaseTTL,
		completionObservationHorizon: options.CompletionObservationHorizon,
	}
}

// NewRateBudgetCoordinator creates a local decision/finish coordinator. It
// is useful for tests and for an injected same-host owner; callers selecting
// the ordinary CLI backend continue to use NewRateLimitRegistry.
func NewRateBudgetCoordinator(clock RateLimitClock, options RateBudgetCoordinatorOptions) *RateBudgetCoordinator {
	return &RateBudgetCoordinator{registry: newRateLimitRegistry(clock, options)}
}

func (c *RateBudgetCoordinator) Decide(ctx context.Context, batch connsdk.ReservationBatch) (connsdk.AdmissionDecision, error) {
	if c == nil || c.registry == nil {
		return connsdk.AdmissionDecision{}, errors.New("rate-budget coordinator is unavailable")
	}
	return c.registry.Decide(ctx, batch)
}

func (c *RateBudgetCoordinator) Finish(ctx context.Context, lease connsdk.RateBudgetLease, observation connsdk.CompletionObservation) error {
	if c == nil || c.registry == nil {
		return errors.New("rate-budget coordinator is unavailable")
	}
	return c.registry.Finish(ctx, lease, observation)
}

// Limiter returns the stable local limiter for key. Declaration validation
// happens in engine; an empty budget slice therefore returns a no-op limiter
// instead of manufacturing a rate policy.
func (r *RateLimitRegistry) Limiter(key RateLimitKey, budgets []connsdk.RateLimitBudget) *RateLimiter {
	if r == nil || key.Connector == "" || key.PolicyID == "" || key.Scope == "" || len(budgets) == 0 {
		return &RateLimiter{}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	set := r.sets[key]
	if set == nil {
		set = newRateLimitSet(budgets)
		r.sets[key] = set
	}
	return &RateLimiter{clock: r.clock, set: set}
}

// RateBudgetPolicyFingerprint is a stable, secret-free commitment to one
// declared policy's consumptive shape. It deliberately excludes policy source,
// selector, scope preimages, credentials, URLs, headers, and request data.
func RateBudgetPolicyFingerprint(connector, policyID string, budgets []connsdk.RateLimitBudget) (string, error) {
	if strings.TrimSpace(connector) == "" || strings.TrimSpace(policyID) == "" || len(budgets) == 0 {
		return "", errors.New("rate-budget policy fingerprint requires a declared policy")
	}
	payload, err := json.Marshal(struct {
		Version   int                       `json:"version"`
		Connector string                    `json:"connector"`
		PolicyID  string                    `json:"policy_id"`
		Budgets   []connsdk.RateLimitBudget `json:"budgets"`
	}{
		Version:   1,
		Connector: connector,
		PolicyID:  policyID,
		Budgets:   budgets,
	})
	if err != nil {
		return "", errors.New("rate-budget policy fingerprint could not be encoded")
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func rateBudgetContractFingerprint(budgets []connsdk.RateLimitBudget) (string, error) {
	if len(budgets) == 0 {
		return "", errors.New("rate-budget contract fingerprint requires a budget")
	}
	payload, err := json.Marshal(struct {
		Version int                       `json:"version"`
		Budgets []connsdk.RateLimitBudget `json:"budgets"`
	}{Version: 1, Budgets: budgets})
	if err != nil {
		return "", errors.New("rate-budget contract fingerprint could not be encoded")
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

// Decide atomically evaluates every selected consumptive policy and reserves
// all of them plus one shared in-flight lease. A non-grant has no lease and
// consumes nothing; callers decide whether their own deadline can safely wait
// until NotBefore.
func (r *RateLimitRegistry) Decide(ctx context.Context, batch connsdk.ReservationBatch) (connsdk.AdmissionDecision, error) {
	if err := ctx.Err(); err != nil {
		return connsdk.AdmissionDecision{}, err
	}
	if r == nil || r.clock == nil {
		return connsdk.AdmissionDecision{}, errors.New("rate-budget coordinator is unavailable")
	}
	if len(batch.Policies) == 0 || len(batch.Policies) > maxRateBudgetReservationPolicies {
		return connsdk.AdmissionDecision{}, errors.New("rate-budget reservation batch is invalid")
	}

	type entry struct {
		key                 rateBudgetKey
		contractFingerprint string
		budgets             []connsdk.RateLimitBudget
		set                 *rateLimitSet
		new                 bool
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return connsdk.AdmissionDecision{}, err
	}
	now := r.clock.Now()
	r.expireLeasesLocked(now)

	entries := make([]entry, 0, len(batch.Policies))
	seen := make(map[rateBudgetKey]struct{}, len(batch.Policies))
	for _, policy := range batch.Policies {
		if strings.TrimSpace(policy.Key.PolicyFingerprint) == "" || strings.TrimSpace(policy.Key.Scope) == "" || len(policy.Budgets) == 0 {
			return connsdk.AdmissionDecision{}, errors.New("rate-budget reservation policy is invalid")
		}
		key := rateBudgetKey{
			policyFingerprint: policy.Key.PolicyFingerprint,
			scope:             policy.Key.Scope,
		}
		contractFingerprint, err := rateBudgetContractFingerprint(policy.Budgets)
		if err != nil {
			return connsdk.AdmissionDecision{}, err
		}
		if _, duplicate := seen[key]; duplicate {
			return connsdk.AdmissionDecision{}, errors.New("rate-budget reservation batch repeats a policy")
		}
		seen[key] = struct{}{}
		existing, exists := r.budgetSets[key]
		if exists && existing.contractFingerprint != contractFingerprint {
			return connsdk.AdmissionDecision{}, errors.New("rate-budget policy fingerprint does not match registered contract")
		}
		entries = append(entries, entry{
			key:                 key,
			contractFingerprint: contractFingerprint,
			budgets:             policy.Budgets,
			set:                 existing.set,
			new:                 !exists,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].key.policyFingerprint != entries[j].key.policyFingerprint {
			return entries[i].key.policyFingerprint < entries[j].key.policyFingerprint
		}
		return entries[i].key.scope < entries[j].key.scope
	})
	for i := range entries {
		if entries[i].new {
			entries[i].set = newRateLimitSet(entries[i].budgets)
		}
		entries[i].set.mu.Lock()
	}

	wait := r.inFlightWaitLocked(now)
	var decisionErr error
	for i := range entries {
		candidate, err := entries[i].set.waitLocked(now)
		if err != nil {
			decisionErr = err
			break
		}
		if candidate > wait {
			wait = candidate
		}
	}
	if decisionErr != nil || wait > 0 {
		for i := len(entries) - 1; i >= 0; i-- {
			entries[i].set.mu.Unlock()
		}
		if decisionErr != nil {
			return connsdk.AdmissionDecision{}, decisionErr
		}
		return connsdk.AdmissionDecision{
			NotBefore: now.Add(wait),
			Wait:      wait,
		}, nil
	}

	lease, err := newRateBudgetLease()
	if err != nil {
		for i := len(entries) - 1; i >= 0; i-- {
			entries[i].set.mu.Unlock()
		}
		return connsdk.AdmissionDecision{}, err
	}
	if err := ctx.Err(); err != nil {
		for i := len(entries) - 1; i >= 0; i-- {
			entries[i].set.mu.Unlock()
		}
		return connsdk.AdmissionDecision{}, err
	}
	sets := make([]*rateLimitSet, 0, len(entries))
	for i := range entries {
		if entries[i].new {
			r.budgetSets[entries[i].key] = rateBudgetSet{
				contractFingerprint: entries[i].contractFingerprint,
				set:                 entries[i].set,
			}
		}
		entries[i].set.consumeLocked(now)
		sets = append(sets, entries[i].set)
	}
	for i := len(entries) - 1; i >= 0; i-- {
		entries[i].set.mu.Unlock()
	}
	expiresAt := now.Add(r.leaseTTL)
	r.leases[lease] = rateBudgetLease{
		grantedAt: now,
		expiresAt: expiresAt,
		active:    true,
		sets:      sets,
	}
	return connsdk.AdmissionDecision{Granted: true, Lease: lease}, nil
}

// Finish releases the in-flight lease exactly once and applies only the
// already-parsed stricter response facts. Expired leases remain finishable
// within the completion-observation horizon while no longer occupying
// concurrency; a missing lease is an idempotent no-op and its uncertain
// consumptive reservation remains charged.
func (r *RateLimitRegistry) Finish(_ context.Context, lease connsdk.RateBudgetLease, observation connsdk.CompletionObservation) error {
	if r == nil || r.clock == nil {
		return errors.New("rate-budget coordinator is unavailable")
	}
	if lease == "" {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.clock.Now()
	r.expireLeasesLocked(now)
	record, ok := r.leases[lease]
	if !ok {
		return nil
	}
	delete(r.leases, lease)
	if !observation.Attempted {
		return nil
	}
	for _, set := range record.sets {
		set.observe(now, observation)
	}
	return nil
}

// Sleep exposes the registry's context-aware clock to the process-local
// adapter. Shared clients use their owner-provided wait with a wall-clock
// timer, while deterministic local tests retain their injected clock.
func (r *RateLimitRegistry) Sleep(ctx context.Context, wait time.Duration) error {
	if r == nil || r.clock == nil {
		return errors.New("rate-budget coordinator is unavailable")
	}
	return r.clock.Sleep(ctx, wait)
}

func (r *RateLimitRegistry) expireLeasesLocked(now time.Time) {
	for lease, record := range r.leases {
		if !record.grantedAt.Add(r.completionObservationHorizon).After(now) {
			delete(r.leases, lease)
			continue
		}
		if record.active && !record.expiresAt.After(now) {
			record.active = false
			r.leases[lease] = record
		}
	}
}

func (r *RateLimitRegistry) inFlightWaitLocked(now time.Time) time.Duration {
	if r.maxInFlight <= 0 {
		return 0
	}
	active := 0
	var earliest time.Time
	for _, record := range r.leases {
		if !record.active {
			continue
		}
		active++
		if earliest.IsZero() || record.expiresAt.Before(earliest) {
			earliest = record.expiresAt
		}
	}
	if active < r.maxInFlight {
		return 0
	}
	if earliest.After(now) {
		return earliest.Sub(now)
	}
	return 0
}

func newRateBudgetLease() (connsdk.RateBudgetLease, error) {
	var bytes [24]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", errors.New("rate-budget lease could not be created")
	}
	return connsdk.RateBudgetLease(hex.EncodeToString(bytes[:])), nil
}

// RateLimiter implements connsdk's admission and observation seams for one
// previously-resolved declared policy.
type RateLimiter struct {
	clock RateLimitClock
	set   *rateLimitSet
}

var _ connsdk.RateLimitAdmission = (*RateLimiter)(nil)
var _ connsdk.RateLimitObserver = (*RateLimiter)(nil)

// Admit reserves capacity for one logical requester send. It has no network
// operation and waits only through the injected context-aware clock.
func (l *RateLimiter) Admit(ctx context.Context, _ connsdk.RateLimitRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if l == nil || l.set == nil || l.clock == nil {
		return nil
	}
	for {
		now := l.clock.Now()
		wait, err := l.set.reserve(now)
		if err != nil {
			return err
		}
		if wait <= 0 {
			return nil
		}
		if err := l.clock.Sleep(ctx, wait); err != nil {
			return err
		}
	}
}

// Observe accepts only the typed requester observation. Provider reset and
// higher actual point costs tighten local state; no observation can expand a
// declared ceiling or replenish a budget.
func (l *RateLimiter) Observe(_ context.Context, observation connsdk.RateLimitObservation) {
	if l == nil || l.set == nil || l.clock == nil || !observation.Attempted {
		return
	}
	l.set.observe(l.clock.Now(), observation)
}

type rateLimitSet struct {
	mu           sync.Mutex
	blockedUntil time.Time
	budgets      []*localRateBudget
}

func newRateLimitSet(specs []connsdk.RateLimitBudget) *rateLimitSet {
	budgets := make([]*localRateBudget, 0, len(specs))
	for _, spec := range specs {
		budgets = append(budgets, &localRateBudget{spec: spec})
	}
	return &rateLimitSet{budgets: budgets}
}

func (s *rateLimitSet) reserve(now time.Time) (time.Duration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	wait, err := s.waitLocked(now)
	if err != nil || wait > 0 {
		return wait, err
	}
	s.consumeLocked(now)
	return 0, nil
}

func (s *rateLimitSet) waitLocked(now time.Time) (time.Duration, error) {
	var wait time.Duration
	if s.blockedUntil.After(now) {
		wait = s.blockedUntil.Sub(now)
	}
	for _, budget := range s.budgets {
		candidate, err := budget.wait(now)
		if err != nil {
			return 0, err
		}
		if candidate > wait {
			wait = candidate
		}
	}
	if wait > 0 {
		return wait, nil
	}
	return 0, nil
}

func (s *rateLimitSet) consumeLocked(now time.Time) {
	for _, budget := range s.budgets {
		budget.consume(now, budget.defaultCost())
	}
}

func (s *rateLimitSet) observe(now time.Time, observation connsdk.RateLimitObservation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rateLimitObservationBlocksUntilReset(observation) && observation.ResetAt.After(s.blockedUntil) {
		s.blockedUntil = observation.ResetAt
	}
	for _, budget := range s.budgets {
		if observation.HasLimit {
			budget.tightenLimit(now, float64(observation.Limit))
		}
		if observation.HasRemaining {
			budget.tightenRemaining(now, float64(observation.Remaining))
		}
		if observation.Status == 429 && !observation.HasReset {
			budget.tightenRemaining(now, 0)
		}
		if observation.HasCost && budget.spec.Unit == connsdk.RateLimitBudgetPoints {
			if extra := observation.Cost - budget.defaultCost(); extra > 0 {
				budget.consume(now, extra)
			}
		}
	}
}

func rateLimitObservationBlocksUntilReset(observation connsdk.RateLimitObservation) bool {
	return observation.HasReset && (observation.HasRetryAfter || observation.Status == http.StatusTooManyRequests || (observation.HasRemaining && observation.Remaining <= 0))
}

type rateLimitUse struct {
	at   time.Time
	cost float64
}

type localRateBudget struct {
	spec connsdk.RateLimitBudget

	windowStart      time.Time
	used             float64
	uses             []rateLimitUse
	tokens           float64
	updatedAt        time.Time
	observedLimit    float64
	hasObservedLimit bool
}

func (b *localRateBudget) defaultCost() float64 {
	if b.spec.Unit != connsdk.RateLimitBudgetPoints || b.spec.Cost == nil || b.spec.Cost.DefaultCost == nil {
		return 1
	}
	return *b.spec.Cost.DefaultCost
}

func (b *localRateBudget) wait(now time.Time) (time.Duration, error) {
	cost := b.defaultCost()
	switch b.spec.Model {
	case connsdk.RateLimitBudgetFixedWindow:
		limit, window := b.fixedWindow()
		if cost > limit {
			return 0, errors.New("provider observation reduced the rate-limit budget below one request cost")
		}
		if b.windowStart.IsZero() || !now.Before(b.windowStart.Add(window)) {
			b.windowStart, b.used = now, 0
		}
		if b.used+cost <= limit {
			return 0, nil
		}
		return b.windowStart.Add(window).Sub(now), nil
	case connsdk.RateLimitBudgetSlidingWindow:
		limit, window := b.fixedWindow()
		if cost > limit {
			return 0, errors.New("provider observation reduced the rate-limit budget below one request cost")
		}
		b.trimUses(now, window)
		if b.used+cost <= limit {
			return 0, nil
		}
		remaining := b.used
		for _, use := range b.uses {
			remaining -= use.cost
			if remaining+cost <= limit {
				return use.at.Add(window).Sub(now), nil
			}
		}
		return window, nil
	case connsdk.RateLimitBudgetTokenBucket, connsdk.RateLimitBudgetLeakyBucket:
		capacity, restore := b.bucket()
		if cost > capacity {
			return 0, errors.New("rate limit request cost exceeds declared capacity")
		}
		b.refill(now, capacity, restore)
		if b.tokens >= cost {
			return 0, nil
		}
		return durationForRate(cost-b.tokens, restore), nil
	default:
		return 0, errors.New("rate limit model is unsupported")
	}
}

func (b *localRateBudget) consume(now time.Time, cost float64) {
	switch b.spec.Model {
	case connsdk.RateLimitBudgetFixedWindow:
		_, window := b.fixedWindow()
		if b.windowStart.IsZero() || !now.Before(b.windowStart.Add(window)) {
			b.windowStart, b.used = now, 0
		}
		b.used += cost
	case connsdk.RateLimitBudgetSlidingWindow:
		_, window := b.fixedWindow()
		b.trimUses(now, window)
		b.uses = append(b.uses, rateLimitUse{at: now, cost: cost})
		b.used += cost
	case connsdk.RateLimitBudgetTokenBucket, connsdk.RateLimitBudgetLeakyBucket:
		capacity, restore := b.bucket()
		b.refill(now, capacity, restore)
		b.tokens -= cost
	}
}

func (b *localRateBudget) tightenRemaining(now time.Time, remaining float64) {
	switch b.spec.Model {
	case connsdk.RateLimitBudgetFixedWindow:
		limit, window := b.fixedWindow()
		if b.windowStart.IsZero() || !now.Before(b.windowStart.Add(window)) {
			b.windowStart, b.used = now, 0
		}
		if used := limit - remaining; used > b.used {
			b.used = used
		}
	case connsdk.RateLimitBudgetSlidingWindow:
		limit, window := b.fixedWindow()
		b.trimUses(now, window)
		if used := limit - remaining; used > b.used {
			b.uses = append(b.uses, rateLimitUse{at: now, cost: used - b.used})
			b.used = used
		}
	case connsdk.RateLimitBudgetTokenBucket, connsdk.RateLimitBudgetLeakyBucket:
		capacity, restore := b.bucket()
		b.refill(now, capacity, restore)
		if remaining < b.tokens {
			b.tokens = remaining
		}
	}
}

func (b *localRateBudget) tightenLimit(now time.Time, limit float64) {
	if limit < 0 || (b.hasObservedLimit && limit >= b.observedLimit) {
		return
	}
	declared := b.declaredLimit()
	if limit > declared {
		return
	}
	b.observedLimit = limit
	b.hasObservedLimit = true
	if b.spec.Model == connsdk.RateLimitBudgetTokenBucket || b.spec.Model == connsdk.RateLimitBudgetLeakyBucket {
		capacity, restore := b.bucket()
		b.refill(now, capacity, restore)
		if b.tokens > capacity {
			b.tokens = capacity
		}
	}
}

func (b *localRateBudget) fixedWindow() (float64, time.Duration) {
	limit, seconds := 1, 1
	if b.spec.Limit != nil {
		limit = *b.spec.Limit
	}
	if b.spec.WindowSeconds != nil {
		seconds = *b.spec.WindowSeconds
	}
	ceiling := float64(limit)
	if b.hasObservedLimit && b.observedLimit < ceiling {
		ceiling = b.observedLimit
	}
	return ceiling, time.Duration(seconds) * time.Second
}

func (b *localRateBudget) bucket() (float64, float64) {
	capacity, restore := 1, 1.0
	if b.spec.Capacity != nil {
		capacity = *b.spec.Capacity
	}
	if b.spec.RestorePerSecond != nil {
		restore = *b.spec.RestorePerSecond
	}
	ceiling := float64(capacity)
	if b.hasObservedLimit && b.observedLimit < ceiling {
		ceiling = b.observedLimit
	}
	return ceiling, restore
}

func (b *localRateBudget) declaredLimit() float64 {
	switch b.spec.Model {
	case connsdk.RateLimitBudgetFixedWindow, connsdk.RateLimitBudgetSlidingWindow:
		limit := 1
		if b.spec.Limit != nil {
			limit = *b.spec.Limit
		}
		return float64(limit)
	case connsdk.RateLimitBudgetTokenBucket, connsdk.RateLimitBudgetLeakyBucket:
		capacity := 1
		if b.spec.Capacity != nil {
			capacity = *b.spec.Capacity
		}
		return float64(capacity)
	default:
		return 1
	}
}

func (b *localRateBudget) refill(now time.Time, capacity, restore float64) {
	if b.updatedAt.IsZero() {
		b.updatedAt, b.tokens = now, capacity
		return
	}
	if now.After(b.updatedAt) {
		b.tokens = math.Min(capacity, b.tokens+now.Sub(b.updatedAt).Seconds()*restore)
		b.updatedAt = now
	}
}

func (b *localRateBudget) trimUses(now time.Time, window time.Duration) {
	first := 0
	for first < len(b.uses) && !b.uses[first].at.Add(window).After(now) {
		b.used -= b.uses[first].cost
		first++
	}
	if first > 0 {
		b.uses = append([]rateLimitUse(nil), b.uses[first:]...)
	}
}

func durationForRate(deficit, restore float64) time.Duration {
	if deficit <= 0 {
		return 0
	}
	seconds := deficit / restore
	return time.Duration(math.Ceil(seconds * float64(time.Second)))
}
