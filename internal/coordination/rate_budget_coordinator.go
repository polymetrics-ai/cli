package coordination

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors/connsdk"
)

const maxRateBudgetReservationPolicies = 64

// RateBudgetCoordinatorOptions bounds the run-local batch-decision seam. A
// non-positive MaxInFlight retains the historical unlimited local behavior.
// LeaseTTL is intentionally short: it releases occupancy after a lost caller
// but leaves the already-reserved budget conservatively charged.
type RateBudgetCoordinatorOptions struct {
	MaxInFlight int
	LeaseTTL    time.Duration
}

// RateBudgetCoordinator owns ephemeral, opaque admission state for one run.
// It is deliberately separate from the optional Dragonfly coordinator: this
// owner has neither persistence nor cross-host meaning.
type RateBudgetCoordinator struct {
	clock       RateLimitClock
	mu          sync.Mutex
	sets        map[rateBudgetKey]rateBudgetSet
	leases      map[connsdk.RateBudgetLease]rateBudgetLease
	maxInFlight int
	leaseTTL    time.Duration
}

type rateBudgetKey struct {
	policyFingerprint string
	scope             string
}

type rateBudgetSet struct {
	contractFingerprint string
	set                 *rateLimitSet
}

type rateBudgetLease struct {
	expiresAt time.Time
	active    bool
	sets      []*rateLimitSet
}

var _ connsdk.BudgetCoordinator = (*RateBudgetCoordinator)(nil)

// NewRateBudgetCoordinator constructs a run-local coordinator. Tests may
// inject the same deterministic clock used by the process-local limiter.
func NewRateBudgetCoordinator(clock RateLimitClock, options RateBudgetCoordinatorOptions) *RateBudgetCoordinator {
	if clock == nil {
		clock = wallRateLimitClock{}
	}
	if options.LeaseTTL <= 0 {
		options.LeaseTTL = 30 * time.Second
	}
	if options.MaxInFlight < 0 {
		options.MaxInFlight = 0
	}
	return &RateBudgetCoordinator{
		clock:       clock,
		sets:        make(map[rateBudgetKey]rateBudgetSet),
		leases:      make(map[connsdk.RateBudgetLease]rateBudgetLease),
		maxInFlight: options.MaxInFlight,
		leaseTTL:    options.LeaseTTL,
	}
}

// RateBudgetPolicyFingerprint commits to a declared consumptive shape without
// carrying its raw provider subject, credential, request, URL, or headers.
func RateBudgetPolicyFingerprint(connector, policyID string, budgets []connsdk.RateLimitBudget) (string, error) {
	if strings.TrimSpace(connector) == "" || strings.TrimSpace(policyID) == "" || len(budgets) == 0 {
		return "", errors.New("rate-budget policy fingerprint requires a declared policy")
	}
	payload, err := json.Marshal(struct {
		Version   int                       `json:"version"`
		Connector string                    `json:"connector"`
		PolicyID  string                    `json:"policy_id"`
		Budgets   []connsdk.RateLimitBudget `json:"budgets"`
	}{Version: 1, Connector: connector, PolicyID: policyID, Budgets: budgets})
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

// Decide atomically checks each selected budget and one in-flight lease. A
// non-grant consumes neither a budget nor an occupancy slot.
func (c *RateBudgetCoordinator) Decide(ctx context.Context, batch connsdk.ReservationBatch) (connsdk.AdmissionDecision, error) {
	if err := ctx.Err(); err != nil {
		return connsdk.AdmissionDecision{}, err
	}
	if c == nil || c.clock == nil {
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

	c.mu.Lock()
	defer c.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return connsdk.AdmissionDecision{}, err
	}
	now := c.clock.Now()
	c.expireLeasesLocked(now)

	entries := make([]entry, 0, len(batch.Policies))
	seen := make(map[rateBudgetKey]struct{}, len(batch.Policies))
	for _, policy := range batch.Policies {
		if strings.TrimSpace(policy.Key.PolicyFingerprint) == "" || strings.TrimSpace(policy.Key.Scope) == "" || len(policy.Budgets) == 0 {
			return connsdk.AdmissionDecision{}, errors.New("rate-budget reservation policy is invalid")
		}
		key := rateBudgetKey{policyFingerprint: policy.Key.PolicyFingerprint, scope: policy.Key.Scope}
		if _, duplicate := seen[key]; duplicate {
			return connsdk.AdmissionDecision{}, errors.New("rate-budget reservation batch repeats a policy")
		}
		seen[key] = struct{}{}
		contractFingerprint, err := rateBudgetContractFingerprint(policy.Budgets)
		if err != nil {
			return connsdk.AdmissionDecision{}, err
		}
		existing, exists := c.sets[key]
		if exists && existing.contractFingerprint != contractFingerprint {
			return connsdk.AdmissionDecision{}, errors.New("rate-budget policy fingerprint does not match registered contract")
		}
		entries = append(entries, entry{key: key, contractFingerprint: contractFingerprint, budgets: policy.Budgets, set: existing.set, new: !exists})
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

	wait := c.inFlightWaitLocked(now)
	for i := range entries {
		candidate, err := entries[i].set.waitLocked(now)
		if err != nil {
			for j := len(entries) - 1; j >= 0; j-- {
				entries[j].set.mu.Unlock()
			}
			return connsdk.AdmissionDecision{}, err
		}
		if candidate > wait {
			wait = candidate
		}
	}
	if wait > 0 {
		for i := len(entries) - 1; i >= 0; i-- {
			entries[i].set.mu.Unlock()
		}
		return connsdk.AdmissionDecision{NotBefore: now.Add(wait), Wait: wait}, nil
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
			c.sets[entries[i].key] = rateBudgetSet{contractFingerprint: entries[i].contractFingerprint, set: entries[i].set}
		}
		entries[i].set.consumeLocked(now)
		sets = append(sets, entries[i].set)
	}
	for i := len(entries) - 1; i >= 0; i-- {
		entries[i].set.mu.Unlock()
	}
	c.leases[lease] = rateBudgetLease{expiresAt: now.Add(c.leaseTTL), active: true, sets: sets}
	return connsdk.AdmissionDecision{Granted: true, Lease: lease}, nil
}

// Finish releases occupancy exactly once and applies parsed stricter response
// facts. Lease expiry deliberately only flips active: a late provider response
// must still be allowed to tighten the budget it already consumed.
func (c *RateBudgetCoordinator) Finish(_ context.Context, lease connsdk.RateBudgetLease, observation connsdk.CompletionObservation) error {
	if c == nil || c.clock == nil {
		return errors.New("rate-budget coordinator is unavailable")
	}
	if lease == "" {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.expireLeasesLocked(c.clock.Now())
	record, ok := c.leases[lease]
	if !ok {
		return nil
	}
	delete(c.leases, lease)
	if !observation.Attempted {
		return nil
	}
	for _, set := range record.sets {
		set.observe(c.clock.Now(), observation)
	}
	return nil
}

func (c *RateBudgetCoordinator) expireLeasesLocked(now time.Time) {
	for lease, record := range c.leases {
		if record.active && !record.expiresAt.After(now) {
			record.active = false
			c.leases[lease] = record
		}
	}
}

func (c *RateBudgetCoordinator) inFlightWaitLocked(now time.Time) time.Duration {
	if c.maxInFlight <= 0 {
		return 0
	}
	active := 0
	var earliest time.Time
	for _, record := range c.leases {
		if !record.active {
			continue
		}
		active++
		if earliest.IsZero() || record.expiresAt.Before(earliest) {
			earliest = record.expiresAt
		}
	}
	if active < c.maxInFlight || !earliest.After(now) {
		return 0
	}
	return earliest.Sub(now)
}

func newRateBudgetLease() (connsdk.RateBudgetLease, error) {
	var bytes [24]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", errors.New("rate-budget lease could not be created")
	}
	return connsdk.RateBudgetLease(hex.EncodeToString(bytes[:])), nil
}
