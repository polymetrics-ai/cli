// Package coordination holds ephemeral, opaque coordination state shared by
// connector executions. It never accepts credential material or raw subjects.
package coordination

import (
	"context"
	"errors"
	"math"
	"net/http"
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

// RateLimitRegistry coordinates process-local policy budgets. It is
// intentionally ephemeral: cross-process coordination remains a separate
// capability and no local state claims account-wide protection.
type RateLimitRegistry struct {
	clock RateLimitClock
	mu    sync.Mutex
	sets  map[RateLimitKey]*rateLimitSet
}

// RateLimitCoordinationMode describes the enforcement boundary without
// exposing a scope or account identity.
type RateLimitCoordinationMode string

const RateLimitCoordinationProcessLocal RateLimitCoordinationMode = "process_local"

// RateLimitCoordinationStatus is safe for ordinary output. It intentionally
// contains no registry key, subject, binding, endpoint, or credential data.
type RateLimitCoordinationStatus struct {
	Mode    RateLimitCoordinationMode
	Message string
}

// NewRateLimitRegistry creates a local registry. A nil clock uses a
// context-aware wall clock; tests should inject a deterministic clock.
func NewRateLimitRegistry(clock RateLimitClock) *RateLimitRegistry {
	if clock == nil {
		clock = wallRateLimitClock{}
	}
	return &RateLimitRegistry{clock: clock, sets: make(map[RateLimitKey]*rateLimitSet)}
}

// Status states the exact coordination boundary of the dependency-free
// registry. Callers must not describe this as account-wide protection.
func (r *RateLimitRegistry) Status() RateLimitCoordinationStatus {
	return RateLimitCoordinationStatus{
		Mode:    RateLimitCoordinationProcessLocal,
		Message: "process-local rate-limit protection; not shared across processes",
	}
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
	for _, budget := range s.budgets {
		budget.consume(now, budget.defaultCost())
	}
	return 0, nil
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
