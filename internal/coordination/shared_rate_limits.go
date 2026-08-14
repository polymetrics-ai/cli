package coordination

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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
)

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
	return fmt.Sprintf("shared rate-limit coordinator %s is unavailable: %s", e.Component, e.Reason)
}

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
// in the error path. Engine calls this before building a require-shared
// requester, so a missing coordinator cannot silently become local admission.
func (r *SharedRateLimitRegistry) EnsureAvailable(ctx context.Context) error {
	if r == nil || r.dragonfly == nil || r.dragonfly.client == nil {
		return sharedRateLimitUnavailable(SharedRateLimitCoordinatorNotConfigured)
	}
	if err := r.dragonfly.Ping(ctx); err != nil {
		return sharedRateLimitUnavailable(SharedRateLimitCoordinatorUnreachable)
	}
	return nil
}

func sharedRateLimitUnavailable(reason SharedRateLimitUnavailableReason) error {
	return &SharedRateLimitUnavailableError{Component: "dragonfly", Reason: reason}
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
	if err := l.registry.EnsureAvailable(ctx); err != nil {
		return err
	}
	specs, err := sharedRateLimitBudgetSpecs(l.budgets)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(specs)
	if err != nil {
		return fmt.Errorf("encode shared rate-limit budget: %w", err)
	}
	ttl := sharedRateLimitTTL(specs)
	for {
		result, err := sharedRateLimitReserveScript.Run(ctx, l.registry.dragonfly.client, []string{sharedRateLimitRedisKey(l.key)}, string(encoded), ttl.Milliseconds()).Result()
		if err != nil {
			return sharedRateLimitUnavailable(SharedRateLimitCoordinatorUnreachable)
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

// Observe propagates only an authoritative provider reset. That fact can make
// another process slower, never faster; local parking/resume remains #3867.
func (l *SharedRateLimiter) Observe(ctx context.Context, observation connsdk.RateLimitObservation) {
	if l == nil || l.registry == nil || !observation.Attempted || !observation.HasReset || observation.ResetAt.IsZero() {
		return
	}
	if l.registry.EnsureAvailable(ctx) != nil {
		return
	}
	specs, err := sharedRateLimitBudgetSpecs(l.budgets)
	if err != nil {
		return
	}
	ttl := sharedRateLimitTTL(specs)
	_, _ = sharedRateLimitObserveScript.Run(ctx, l.registry.dragonfly.client, []string{sharedRateLimitRedisKey(l.key)}, observation.ResetAt.UTC().UnixMilli(), ttl.Milliseconds()).Result()
}

type sharedRateLimitBudget struct {
	Model    string  `json:"model"`
	Limit    float64 `json:"limit"`
	Window   int64   `json:"window_ms"`
	Capacity float64 `json:"capacity"`
	Restore  float64 `json:"restore_per_ms"`
	Cost     float64 `json:"cost"`
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
		spec := sharedRateLimitBudget{Model: string(budget.Model), Cost: cost}
		switch budget.Model {
		case connsdk.RateLimitBudgetFixedWindow, connsdk.RateLimitBudgetSlidingWindow:
			if budget.Limit == nil || budget.WindowSeconds == nil || *budget.Limit <= 0 || *budget.WindowSeconds <= 0 {
				return nil, errors.New("shared rate-limit window budget is invalid")
			}
			spec.Limit = float64(*budget.Limit)
			spec.Window = int64(*budget.WindowSeconds) * int64(time.Second/time.Millisecond)
		case connsdk.RateLimitBudgetTokenBucket, connsdk.RateLimitBudgetLeakyBucket:
			if budget.Capacity == nil || budget.RestorePerSecond == nil || *budget.Capacity <= 0 || *budget.RestorePerSecond <= 0 {
				return nil, errors.New("shared rate-limit bucket budget is invalid")
			}
			spec.Capacity = float64(*budget.Capacity)
			spec.Restore = *budget.RestorePerSecond / float64(time.Second/time.Millisecond)
		default:
			return nil, fmt.Errorf("shared rate-limit model %q is unsupported", budget.Model)
		}
		out = append(out, spec)
	}
	return out, nil
}

func sharedRateLimitTTL(specs []sharedRateLimitBudget) time.Duration {
	ttl := time.Second
	for _, spec := range specs {
		candidate := time.Duration(spec.Window) * time.Millisecond
		if spec.Restore > 0 {
			candidate = time.Duration(spec.Capacity/spec.Restore) * time.Millisecond
		}
		if candidate > ttl {
			ttl = candidate
		}
	}
	return ttl + time.Second
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

for i, spec in ipairs(specs) do
  local entry = state.budgets[tostring(i)] or {}
  state.budgets[tostring(i)] = entry
  if spec.model == 'fixed_window' then
    if (not entry.start) or now >= entry.start + spec.window_ms then entry.start = now; entry.used = 0 end
    entry.used = entry.used or 0
    if spec.cost > spec.limit then return redis.error_reply('ERR shared rate-limit request cost exceeds declared capacity') end
    if entry.used + spec.cost > spec.limit then wait = math.max(wait, entry.start + spec.window_ms - now) end
  elseif spec.model == 'sliding_window' then
    entry.uses = entry.uses or {}
    local kept, used = {}, 0
    for _, use in ipairs(entry.uses) do
      if now < use.at + spec.window_ms then table.insert(kept, use); used = used + use.cost end
    end
    entry.uses = kept
    if spec.cost > spec.limit then return redis.error_reply('ERR shared rate-limit request cost exceeds declared capacity') end
    if used + spec.cost > spec.limit then
      local remaining = used
      for _, use in ipairs(kept) do
        remaining = remaining - use.cost
        if remaining + spec.cost <= spec.limit then wait = math.max(wait, use.at + spec.window_ms - now); break end
      end
    end
  elseif spec.model == 'token_bucket' or spec.model == 'leaky_bucket' then
    if not entry.updated then entry.updated = now; entry.tokens = spec.capacity end
    entry.tokens = math.min(spec.capacity, entry.tokens + math.max(0, now - entry.updated) * spec.restore_per_ms)
    entry.updated = now
    if spec.cost > spec.capacity then return redis.error_reply('ERR shared rate-limit request cost exceeds declared capacity') end
    if entry.tokens < spec.cost then wait = math.max(wait, math.ceil((spec.cost - entry.tokens) / spec.restore_per_ms)) end
  else
    return redis.error_reply('ERR shared rate-limit model is unsupported')
  end
end

if state.blocked_until and state.blocked_until > now then wait = math.max(wait, state.blocked_until - now) end
if wait > 0 then redis.call('PSETEX', KEYS[1], ttl, cjson.encode(state)); return wait end

for i, spec in ipairs(specs) do
  local entry = state.budgets[tostring(i)]
  if spec.model == 'fixed_window' then entry.used = entry.used + spec.cost
  elseif spec.model == 'sliding_window' then table.insert(entry.uses, {at = now, cost = spec.cost})
  else entry.tokens = entry.tokens - spec.cost end
end
redis.call('PSETEX', KEYS[1], ttl, cjson.encode(state))
return 0
`)

var sharedRateLimitObserveScript = redis.NewScript(`
local nowParts = redis.call('TIME')
local now = tonumber(nowParts[1]) * 1000 + math.floor(tonumber(nowParts[2]) / 1000)
local resetAt = tonumber(ARGV[1])
local ttl = tonumber(ARGV[2])
local raw = redis.call('GET', KEYS[1])
local state = raw and cjson.decode(raw) or {budgets = {}}
if resetAt > now and ((not state.blocked_until) or resetAt > state.blocked_until) then state.blocked_until = resetAt end
if state.blocked_until and state.blocked_until > now then ttl = math.max(ttl, state.blocked_until - now) end
redis.call('PSETEX', KEYS[1], ttl, cjson.encode(state))
return 1
`)
