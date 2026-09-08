package coordination

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"polymetrics.ai/internal/connectors/connsdk"
)

// SharedRateLimitUnavailableReason is a stable, non-sensitive reason for a
// require-shared policy refusal.
type SharedRateLimitUnavailableReason string

const (
	SharedRateLimitCoordinatorNotConfigured SharedRateLimitUnavailableReason = "coordinator_not_configured"
	SharedRateLimitCoordinatorUnreachable   SharedRateLimitUnavailableReason = "coordinator_unreachable"
	SharedRateLimitRouteUnresolved          SharedRateLimitUnavailableReason = "route_unresolved"
)

// SharedRateLimitWindowErrorReason distinguishes declaration errors that must
// be rejected before the limiter can contact the shared coordinator.
type SharedRateLimitWindowErrorReason string

const (
	SharedRateLimitWindowNonPositive SharedRateLimitWindowErrorReason = "non_positive"
	SharedRateLimitWindowTooLarge    SharedRateLimitWindowErrorReason = "too_large"
)

// A fixed/sliding window needs one additional second of cache lifetime so a
// coordinator can retain the just-ended window while it evaluates the next
// request. This is the largest whole-second value for which both the duration
// and that TTL remain representable by time.Duration.
const maxSharedRateLimitWindowSeconds int64 = 9_223_372_035

// SharedRateLimitWindowError is the typed outcome for an invalid declared
// window. Window declarations are non-secret connector metadata, so the value
// is retained to make a bundle-authoring error actionable.
type SharedRateLimitWindowError struct {
	Seconds int64
	Reason  SharedRateLimitWindowErrorReason
}

func (e *SharedRateLimitWindowError) Error() string {
	if e == nil {
		return "shared rate-limit window is invalid"
	}
	return fmt.Sprintf("shared rate-limit window_seconds=%d is invalid: %s", e.Seconds, e.Reason)
}

// SharedRateLimitUnavailableError means an explicitly require-shared policy
// cannot be enforced. It deliberately carries only the missing component and
// reason, never a Redis address, scope, subject, credential, or raw error.
type SharedRateLimitUnavailableError struct {
	Component string
	Reason    SharedRateLimitUnavailableReason
}

func (e *SharedRateLimitUnavailableError) Error() string {
	if e == nil {
		return "shared rate-limit coordinator is unavailable"
	}
	if e.Reason == SharedRateLimitRouteUnresolved {
		return "shared rate-limit route is unresolved"
	}
	return fmt.Sprintf("shared rate-limit coordinator %s is unavailable: %s", e.Component, e.Reason)
}

var errSharedRateLimitPolicy = errors.New("shared rate-limit policy cannot admit one request")

// SharedRateLimitRegistry coordinates explicitly opted-in policies through a
// Redis-compatible server. Its state is ephemeral and contains only opaque
// registry keys and declared budget counters.
type SharedRateLimitRegistry struct {
	dragonfly *Dragonfly
}

// NewSharedRateLimitRegistry builds a registry around an optional Dragonfly
// client. A nil client is useful for an honest require-shared refusal.
func NewSharedRateLimitRegistry(dragonfly *Dragonfly) *SharedRateLimitRegistry {
	return &SharedRateLimitRegistry{dragonfly: dragonfly}
}

// OpenSharedRateLimitRegistry creates an optional Redis-compatible registry.
// It does not make a network call; only a require-shared policy verifies the
// coordinator before sending its connector request.
func OpenSharedRateLimitRegistry(addr string) *SharedRateLimitRegistry {
	if addr == "" {
		return NewSharedRateLimitRegistry(nil)
	}
	return NewSharedRateLimitRegistry(OpenDragonfly(addr))
}

// Close releases the optional client. It is safe for an unconfigured registry.
func (r *SharedRateLimitRegistry) Close() error {
	if r == nil || r.dragonfly == nil {
		return nil
	}
	return r.dragonfly.Close()
}

// EnsureAvailable verifies the coordinator without exposing transport details
// in the error path. Engine and the limiter call this before a require-shared
// request can be sent, so a missing coordinator cannot silently become local admission.
func (r *SharedRateLimitRegistry) EnsureAvailable(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r == nil || r.dragonfly == nil || r.dragonfly.client == nil {
		return sharedRateLimitUnavailable(SharedRateLimitCoordinatorNotConfigured)
	}
	if err := r.dragonfly.Ping(ctx); err != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
		return sharedRateLimitUnavailable(SharedRateLimitCoordinatorUnreachable)
	}
	return nil
}

func sharedRateLimitUnavailable(reason SharedRateLimitUnavailableReason) error {
	return &SharedRateLimitUnavailableError{Component: "dragonfly", Reason: reason}
}

func sharedRateLimitReserveError(err error) error {
	if strings.HasPrefix(strings.TrimSpace(err.Error()), "ERR shared rate-limit ") {
		return errSharedRateLimitPolicy
	}
	return sharedRateLimitUnavailable(SharedRateLimitCoordinatorUnreachable)
}

// Limiter returns a shared limiter for one already-opaque key. The limiter
// performs availability checks itself as a race-safe second line of defence.
func (r *SharedRateLimitRegistry) Limiter(key RateLimitKey, budgets []connsdk.RateLimitBudget) *SharedRateLimiter {
	return &SharedRateLimiter{registry: r, key: key, budgets: append([]connsdk.RateLimitBudget(nil), budgets...)}
}

// SharedRateLimiter implements the same requester seams as the local limiter
// while keeping mutable budget state in one atomic Redis key.
type SharedRateLimiter struct {
	registry *SharedRateLimitRegistry
	key      RateLimitKey
	budgets  []connsdk.RateLimitBudget
}

var _ connsdk.RateLimitAdmission = (*SharedRateLimiter)(nil)
var _ connsdk.RateLimitObserver = (*SharedRateLimiter)(nil)

// Admit atomically reserves every budget from server-time state. A retry waits
// with the caller's context; no goroutine or timer survives caller cancellation.
func (l *SharedRateLimiter) Admit(ctx context.Context, _ connsdk.RateLimitRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if l == nil || l.registry == nil {
		return sharedRateLimitUnavailable(SharedRateLimitCoordinatorNotConfigured)
	}
	specs, err := sharedRateLimitBudgetSpecs(l.budgets)
	if err != nil {
		return err
	}
	ttl, err := sharedRateLimitTTL(specs)
	if err != nil {
		return err
	}
	if err := l.registry.EnsureAvailable(ctx); err != nil {
		return err
	}
	encoded, err := json.Marshal(specs)
	if err != nil {
		return fmt.Errorf("encode shared rate-limit budget: %w", err)
	}
	for {
		result, err := sharedRateLimitReserveScript.Run(ctx, l.registry.dragonfly.client, []string{sharedRateLimitRedisKey(l.key)}, string(encoded), ttl.Milliseconds()).Result()
		if err != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
			return sharedRateLimitReserveError(err)
		}
		wait, err := sharedRateLimitWait(result)
		if err != nil {
			return err
		}
		if wait <= 0 {
			return nil
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (l *SharedRateLimiter) Observe(ctx context.Context, observation connsdk.RateLimitObservation) {
	if l == nil || l.registry == nil || !observation.Attempted {
		return
	}
	sharedObservation := sharedRateLimitObservationOf(observation)
	if !sharedObservation.relevant() {
		return
	}
	specs, err := sharedRateLimitBudgetSpecs(l.budgets)
	if err != nil {
		return
	}
	ttl, err := sharedRateLimitTTL(specs)
	if err != nil {
		return
	}
	if l.registry.EnsureAvailable(ctx) != nil {
		return
	}
	args, err := sharedRateLimitObserveScriptArgs(specs, sharedObservation, ttl)
	if err != nil {
		return
	}
	_, _ = sharedRateLimitObserveScript.Run(ctx, l.registry.dragonfly.client, []string{sharedRateLimitRedisKey(l.key)}, args...).Result()
}

type sharedRateLimitObservation struct {
	BlockFor           int64   `json:"block_for_ms"`
	AbsoluteResetAt    int64   `json:"absolute_reset_at_ms"`
	Limit              float64 `json:"limit"`
	HasLimit           bool    `json:"has_limit"`
	Remaining          float64 `json:"remaining"`
	HasRemaining       bool    `json:"has_remaining"`
	ForceRemainingZero bool    `json:"force_remaining_zero"`
	Cost               float64 `json:"cost"`
	HasCost            bool    `json:"has_cost"`
	CostSource         string  `json:"cost_source"`
}

func sharedRateLimitObservationOf(observation connsdk.RateLimitObservation) sharedRateLimitObservation {
	observedAt := observation.ObservedAt
	if observedAt.IsZero() {
		observedAt = time.Now()
	}
	return sharedRateLimitObservationOfAt(observation, observedAt)
}

func sharedRateLimitObservationOfAt(observation connsdk.RateLimitObservation, observedAt time.Time) sharedRateLimitObservation {
	shared := sharedRateLimitObservation{
		Limit:              float64(observation.Limit),
		HasLimit:           observation.HasLimit,
		Remaining:          float64(observation.Remaining),
		HasRemaining:       observation.HasRemaining,
		ForceRemainingZero: observation.Status == 429 && !observation.HasReset,
		Cost:               observation.Cost,
		HasCost:            observation.HasCost,
		CostSource:         string(observation.CostSource),
	}
	if rateLimitObservationBlocksUntilReset(observation) && !observation.ResetAt.IsZero() {
		if observation.ResetAtAbsolute {
			shared.AbsoluteResetAt = observation.ResetAt.UTC().UnixMilli()
		} else if blockFor := observation.ResetAt.Sub(observedAt); blockFor > 0 {
			shared.BlockFor = blockFor.Milliseconds()
			if shared.BlockFor == 0 {
				shared.BlockFor = 1
			}
		}
	}
	return shared
}

func (o sharedRateLimitObservation) relevant() bool {
	return o.BlockFor > 0 || o.AbsoluteResetAt > 0 || o.HasLimit || o.HasRemaining || o.ForceRemainingZero || o.HasCost
}

type sharedRateLimitBudget struct {
	Model      string  `json:"model"`
	Unit       string  `json:"unit"`
	Limit      float64 `json:"limit"`
	Window     int64   `json:"window_ms"`
	Capacity   float64 `json:"capacity"`
	Restore    float64 `json:"restore_per_ms"`
	Cost       float64 `json:"cost"`
	CostSource string  `json:"cost_source"`
}

func sharedRateLimitBudgetSpecs(budgets []connsdk.RateLimitBudget) ([]sharedRateLimitBudget, error) {
	if len(budgets) == 0 {
		return nil, errors.New("shared rate-limit policy has no budgets")
	}
	out := make([]sharedRateLimitBudget, 0, len(budgets))
	for _, budget := range budgets {
		cost := 1.0
		if budget.Unit == connsdk.RateLimitBudgetPoints && budget.Cost != nil && budget.Cost.DefaultCost != nil {
			cost = *budget.Cost.DefaultCost
		}
		if cost <= 0 {
			return nil, errors.New("shared rate-limit request cost must be positive")
		}
		spec := sharedRateLimitBudget{Model: string(budget.Model), Unit: string(budget.Unit), Cost: cost}
		if budget.Cost != nil {
			switch {
			case budget.Cost.ResponseHeader != "":
				spec.CostSource = string(connsdk.RateLimitCostSourceResponseHeader)
			case budget.Cost.ResponseBody != "":
				spec.CostSource = budget.Cost.ResponseBody
			}
		}
		switch budget.Model {
		case connsdk.RateLimitBudgetFixedWindow, connsdk.RateLimitBudgetSlidingWindow:
			if budget.Limit == nil || budget.WindowSeconds == nil || *budget.Limit <= 0 {
				return nil, errors.New("shared rate-limit window budget is invalid")
			}
			if err := validateSharedRateLimitWindowSeconds(*budget.WindowSeconds); err != nil {
				return nil, err
			}
			spec.Limit = float64(*budget.Limit)
			spec.Window = int64(*budget.WindowSeconds) * int64(time.Second/time.Millisecond)
			if spec.Cost > spec.Limit {
				return nil, errSharedRateLimitPolicy
			}
		case connsdk.RateLimitBudgetTokenBucket, connsdk.RateLimitBudgetLeakyBucket:
			if budget.Capacity == nil || budget.RestorePerSecond == nil || *budget.Capacity <= 0 || *budget.RestorePerSecond <= 0 {
				return nil, errors.New("shared rate-limit bucket budget is invalid")
			}
			spec.Capacity = float64(*budget.Capacity)
			spec.Restore = *budget.RestorePerSecond / float64(time.Second/time.Millisecond)
			if spec.Cost > spec.Capacity {
				return nil, errSharedRateLimitPolicy
			}
		default:
			return nil, fmt.Errorf("shared rate-limit model %q is unsupported", budget.Model)
		}
		out = append(out, spec)
	}
	return out, nil
}

func validateSharedRateLimitWindowSeconds(seconds int) error {
	value := int64(seconds)
	if value <= 0 {
		return &SharedRateLimitWindowError{Seconds: value, Reason: SharedRateLimitWindowNonPositive}
	}
	if value > maxSharedRateLimitWindowSeconds {
		return &SharedRateLimitWindowError{Seconds: value, Reason: SharedRateLimitWindowTooLarge}
	}
	return nil
}

func sharedRateLimitTTL(specs []sharedRateLimitBudget) (time.Duration, error) {
	ttl := time.Second
	for _, spec := range specs {
		if spec.Window < 0 || spec.Window > maxSharedRateLimitWindowSeconds*int64(time.Second/time.Millisecond) {
			return 0, &SharedRateLimitWindowError{Seconds: spec.Window / int64(time.Second/time.Millisecond), Reason: SharedRateLimitWindowTooLarge}
		}
		candidate := time.Duration(spec.Window) * time.Millisecond
		if spec.Restore > 0 {
			candidate = time.Duration(spec.Capacity/spec.Restore) * time.Millisecond
		}
		if candidate > ttl {
			ttl = candidate
		}
	}
	if ttl > time.Duration(1<<63-1)-time.Second {
		return 0, &SharedRateLimitWindowError{Seconds: maxSharedRateLimitWindowSeconds + 1, Reason: SharedRateLimitWindowTooLarge}
	}
	return ttl + time.Second, nil
}

func sharedRateLimitObserveScriptArgs(specs []sharedRateLimitBudget, observation sharedRateLimitObservation, ttl time.Duration) ([]any, error) {
	encodedSpecs, err := json.Marshal(specs)
	if err != nil {
		return nil, err
	}
	encodedObservation, err := json.Marshal(observation)
	if err != nil {
		return nil, err
	}
	return []any{string(encodedSpecs), string(encodedObservation), ttl.Milliseconds()}, nil
}

func sharedRateLimitWait(result any) (time.Duration, error) {
	switch value := result.(type) {
	case int64:
		return time.Duration(value) * time.Millisecond, nil
	case string:
		milliseconds, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse shared rate-limit wait: %w", err)
		}
		return time.Duration(milliseconds) * time.Millisecond, nil
	default:
		return 0, fmt.Errorf("shared rate-limit coordinator returned invalid wait type %T", result)
	}
}

// sharedRateLimitRedisKey contains only names already validated in a bundle
// plus the opaque #3863 scope projection. No raw subject or credential exists
// at this layer.
func sharedRateLimitRedisKey(key RateLimitKey) string {
	return "polymetrics:rate-limit:v1:" + key.Connector + ":" + key.PolicyID + ":" + string(key.Scope)
}

var sharedRateLimitReserveScript = redis.NewScript(`
local nowParts = redis.call('TIME')
local now = tonumber(nowParts[1]) * 1000 + math.floor(tonumber(nowParts[2]) / 1000)
local specs = cjson.decode(ARGV[1])
local ttl = tonumber(ARGV[2])
local raw = redis.call('GET', KEYS[1])
local state = raw and cjson.decode(raw) or {}
state.budgets = state.budgets or {}
local wait = 0

local function preserveDebtTTL()
  for i, spec in ipairs(specs) do
    if (spec.model == 'token_bucket' or spec.model == 'leaky_bucket') and spec.restore_per_ms > 0 then
      local entry = state.budgets[tostring(i)]
      if entry and entry.tokens and entry.tokens < spec.cost then
        ttl = math.max(ttl, math.ceil((spec.cost - entry.tokens) / spec.restore_per_ms))
      end
    end
  end
end

for i, spec in ipairs(specs) do
  local entry = state.budgets[tostring(i)] or {}
  state.budgets[tostring(i)] = entry
  if spec.model == 'fixed_window' then
    local limit = spec.limit
    if entry.observed_limit and entry.observed_limit < limit then limit = entry.observed_limit end
    if (not entry.start) or now >= entry.start + spec.window_ms then entry.start = now; entry.used = 0 end
    entry.used = entry.used or 0
    if spec.cost > limit then return redis.error_reply('ERR shared rate-limit request cost exceeds declared capacity') end
    if entry.used + spec.cost > limit then wait = math.max(wait, entry.start + spec.window_ms - now) end
  elseif spec.model == 'sliding_window' then
    local limit = spec.limit
    if entry.observed_limit and entry.observed_limit < limit then limit = entry.observed_limit end
    entry.uses = entry.uses or {}
    local kept, used = {}, 0
    for _, use in ipairs(entry.uses) do
      if now < use.at + spec.window_ms then table.insert(kept, use); used = used + use.cost end
    end
    entry.uses = kept
    if spec.cost > limit then return redis.error_reply('ERR shared rate-limit request cost exceeds declared capacity') end
    if used + spec.cost > limit then
      local remaining = used
      for _, use in ipairs(kept) do
        remaining = remaining - use.cost
        if remaining + spec.cost <= limit then wait = math.max(wait, use.at + spec.window_ms - now); break end
      end
    end
  elseif spec.model == 'token_bucket' or spec.model == 'leaky_bucket' then
    local capacity = spec.capacity
    if entry.observed_limit and entry.observed_limit < capacity then capacity = entry.observed_limit end
    if not entry.updated then entry.updated = now; entry.tokens = capacity end
    entry.tokens = math.min(capacity, entry.tokens + math.max(0, now - entry.updated) * spec.restore_per_ms)
    entry.updated = now
    if spec.cost > capacity then return redis.error_reply('ERR shared rate-limit request cost exceeds declared capacity') end
    if entry.tokens < spec.cost then wait = math.max(wait, math.ceil((spec.cost - entry.tokens) / spec.restore_per_ms)) end
  else
    return redis.error_reply('ERR shared rate-limit model is unsupported')
  end
end

preserveDebtTTL()
if state.blocked_until and state.blocked_until > now then
	local blockedTTL = state.blocked_until - now
	wait = math.max(wait, blockedTTL)
  ttl = math.max(ttl, blockedTTL)
end
if wait > 0 then redis.call('PSETEX', KEYS[1], ttl, cjson.encode(state)); return wait end

for i, spec in ipairs(specs) do
  local entry = state.budgets[tostring(i)]
  if spec.model == 'fixed_window' then entry.used = entry.used + spec.cost
  elseif spec.model == 'sliding_window' then table.insert(entry.uses, {at = now, cost = spec.cost})
  else entry.tokens = entry.tokens - spec.cost end
end
preserveDebtTTL()
redis.call('PSETEX', KEYS[1], ttl, cjson.encode(state))
return 0
`)

var sharedRateLimitObserveScript = redis.NewScript(`
local nowParts = redis.call('TIME')
local now = tonumber(nowParts[1]) * 1000 + math.floor(tonumber(nowParts[2]) / 1000)
local specs = cjson.decode(ARGV[1])
local observation = cjson.decode(ARGV[2])
local ttl = tonumber(ARGV[3])
local raw = redis.call('GET', KEYS[1])
local state = raw and cjson.decode(raw) or {budgets = {}}
state.budgets = state.budgets or {}

local function preserveDebtTTL()
  for i, spec in ipairs(specs) do
    if (spec.model == 'token_bucket' or spec.model == 'leaky_bucket') and spec.restore_per_ms > 0 then
      local entry = state.budgets[tostring(i)]
      if entry and entry.tokens and entry.tokens < spec.cost then
        ttl = math.max(ttl, math.ceil((spec.cost - entry.tokens) / spec.restore_per_ms))
      end
    end
  end
end

for i, spec in ipairs(specs) do
  local entry = state.budgets[tostring(i)] or {}
  state.budgets[tostring(i)] = entry
  local declared = spec.limit
  if spec.model == 'token_bucket' or spec.model == 'leaky_bucket' then declared = spec.capacity end
  if observation.has_limit and observation.limit >= 0 and observation.limit < declared and ((not entry.observed_limit) or observation.limit < entry.observed_limit) then entry.observed_limit = observation.limit end
  local limit = declared
  if entry.observed_limit and entry.observed_limit < limit then limit = entry.observed_limit end
  local remaining = nil
  if observation.force_remaining_zero then remaining = 0 elseif observation.has_remaining then remaining = observation.remaining end

  if spec.model == 'fixed_window' then
    if (not entry.start) or now >= entry.start + spec.window_ms then entry.start = now; entry.used = 0 end
    entry.used = entry.used or 0
    if remaining and limit - remaining > entry.used then entry.used = limit - remaining end
    if observation.has_cost and spec.unit == 'points' and (not observation.cost_source or observation.cost_source == '' or spec.cost_source == observation.cost_source) and observation.cost > spec.cost then entry.used = entry.used + observation.cost - spec.cost end
  elseif spec.model == 'sliding_window' then
    entry.uses = entry.uses or {}
    local kept, used = {}, 0
    for _, use in ipairs(entry.uses) do
      if now < use.at + spec.window_ms then table.insert(kept, use); used = used + use.cost end
    end
    entry.uses = kept
    if remaining and limit - remaining > used then
      table.insert(entry.uses, {at = now, cost = limit - remaining - used})
      used = limit - remaining
    end
    if observation.has_cost and spec.unit == 'points' and (not observation.cost_source or observation.cost_source == '' or spec.cost_source == observation.cost_source) and observation.cost > spec.cost then table.insert(entry.uses, {at = now, cost = observation.cost - spec.cost}) end
  elseif spec.model == 'token_bucket' or spec.model == 'leaky_bucket' then
    if not entry.updated then entry.updated = now; entry.tokens = limit end
    entry.tokens = math.min(limit, entry.tokens + math.max(0, now - entry.updated) * spec.restore_per_ms)
    entry.updated = now
    if remaining and remaining < entry.tokens then entry.tokens = remaining end
    if observation.has_cost and spec.unit == 'points' and (not observation.cost_source or observation.cost_source == '' or spec.cost_source == observation.cost_source) and observation.cost > spec.cost then entry.tokens = entry.tokens - (observation.cost - spec.cost) end
  end
end

preserveDebtTTL()
local blockedUntil = nil
if observation.block_for_ms > 0 then
  blockedUntil = now + observation.block_for_ms
elseif observation.absolute_reset_at_ms and observation.absolute_reset_at_ms > now then
  blockedUntil = observation.absolute_reset_at_ms
end
if blockedUntil and ((not state.blocked_until) or blockedUntil > state.blocked_until) then
  state.blocked_until = blockedUntil
end
if state.blocked_until and state.blocked_until > now then ttl = math.max(ttl, state.blocked_until - now) end
redis.call('PSETEX', KEYS[1], ttl, cjson.encode(state))
return 1
`)
