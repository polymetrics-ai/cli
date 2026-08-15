package coordination

import (
	"context"
	"errors"
	"sync"

	"polymetrics.ai/internal/connectors"
)

var (
	errAuthCohortHealthStoreUnavailable = errors.New("authentication cohort health store is unavailable")
	// ErrAuthCohortFenced means a verified invalid-authentication result has
	// stopped the cohort. Callers must not send work when this is returned.
	ErrAuthCohortFenced = errors.New("authentication cohort is fenced")
	// ErrAuthCohortEpochMismatch means a member belongs to an older healthy
	// epoch and cannot resume after repair or a restarted coordinator reload.
	ErrAuthCohortEpochMismatch = errors.New("authentication cohort epoch is stale")
	// ErrAuthCohortAdmissionClosed means the caller released its member before
	// asking the coordinator to admit a send.
	ErrAuthCohortAdmissionClosed = errors.New("authentication cohort admission is closed")
)

// AuthenticationOutcome is a closed, provider-neutral classification of an
// authentication attempt. It contains no provider status/body/header data.
// Only AuthenticationOutcomeVerifiedInvalid changes cohort health to fenced.
type AuthenticationOutcome uint8

const (
	AuthenticationOutcomeUnknown AuthenticationOutcome = iota
	AuthenticationOutcomeUnverifiedInvalid
	AuthenticationOutcomeTransportFailure
	AuthenticationOutcomeTimeout
	AuthenticationOutcomeProviderFailure
	AuthenticationOutcomeVerifiedInvalid
	AuthenticationOutcomeVerifiedHealthy
)

// AuthCohortEpoch identifies one healthy lifetime of a cohort. It is local
// coordination evidence, never a user-facing credential or provider value.
type AuthCohortEpoch uint64

// AuthCohortHealth is the minimal opaque state that must survive a coordinator
// restart. The key is supplied separately to the store and is an already
// secret-free connectors.AuthCohortKey.
type AuthCohortHealth struct {
	Epoch           AuthCohortEpoch
	Fenced          bool
	LastFencedEpoch AuthCohortEpoch
}

// AuthCohortHealthStore persists only an opaque cohort health record. A
// coordinator owner must serialize concurrent live coordinators that share a
// store; individual coordinator operations are safe for concurrent callers.
type AuthCohortHealthStore interface {
	Load(connectors.AuthCohortKey) (AuthCohortHealth, bool, error)
	Store(connectors.AuthCohortKey, AuthCohortHealth) error
}

// MemoryAuthCohortHealthStore is a race-safe store for one process. It exists
// both as the default coordinator store and as the deterministic restart seam
// for tests; durable application-state wiring belongs to a later integration
// slice.
type MemoryAuthCohortHealthStore struct {
	mu     sync.RWMutex
	health map[connectors.AuthCohortKey]AuthCohortHealth
}

// NewMemoryAuthCohortHealthStore returns an empty, usable health store.
func NewMemoryAuthCohortHealthStore() *MemoryAuthCohortHealthStore {
	return &MemoryAuthCohortHealthStore{health: make(map[connectors.AuthCohortKey]AuthCohortHealth)}
}

// Load returns the stored health and whether the opaque cohort has been seen.
func (s *MemoryAuthCohortHealthStore) Load(cohort connectors.AuthCohortKey) (AuthCohortHealth, bool, error) {
	if s == nil {
		return AuthCohortHealth{}, false, errAuthCohortHealthStoreUnavailable
	}
	s.mu.RLock()
	health, ok := s.health[cohort]
	s.mu.RUnlock()
	return health, ok, nil
}

// Store replaces one opaque cohort health record.
func (s *MemoryAuthCohortHealthStore) Store(cohort connectors.AuthCohortKey, health AuthCohortHealth) error {
	if s == nil {
		return errAuthCohortHealthStoreUnavailable
	}
	s.mu.Lock()
	if s.health == nil {
		s.health = make(map[connectors.AuthCohortKey]AuthCohortHealth)
	}
	s.health[cohort] = health
	s.mu.Unlock()
	return nil
}

// AuthCohortCoordinator owns active members for one process and persists only
// their opaque health. It is deliberately connector-neutral: it neither sends
// requests nor classifies provider responses.
type AuthCohortCoordinator struct {
	mu      sync.Mutex
	store   AuthCohortHealthStore
	nextID  uint64
	members map[connectors.AuthCohortKey]map[uint64]*AuthCohortAdmission
}

// AuthCohortAdmission is a cancellable in-flight member of one cohort epoch.
// Callers pass Context to their normal request/admission boundary and call
// Check immediately before a send. Release is idempotent.
type AuthCohortAdmission struct {
	coordinator *AuthCohortCoordinator
	cohort      connectors.AuthCohortKey
	epoch       AuthCohortEpoch
	id          uint64
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewAuthCohortCoordinator constructs a coordinator. A nil store uses an
// in-memory store so the zero-configuration path stays secret-free and usable.
func NewAuthCohortCoordinator(store AuthCohortHealthStore) *AuthCohortCoordinator {
	if store == nil {
		store = NewMemoryAuthCohortHealthStore()
	}
	return &AuthCohortCoordinator{
		store:   store,
		members: make(map[connectors.AuthCohortKey]map[uint64]*AuthCohortAdmission),
	}
}

// Admit joins an active cohort epoch after reading persisted health. Fenced or
// unavailable health refuses before a caller can reach its send boundary.
func (c *AuthCohortCoordinator) Admit(ctx context.Context, cohort connectors.AuthCohortKey) (*AuthCohortAdmission, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if c == nil || c.store == nil {
		return nil, errors.New("authentication cohort coordinator is unavailable")
	}
	if cohort == "" {
		return nil, errors.New("authentication cohort is unavailable")
	}

	c.mu.Lock()
	health, err := c.healthLocked(cohort)
	if err != nil {
		c.mu.Unlock()
		return nil, err
	}
	if health.Fenced {
		c.mu.Unlock()
		return nil, ErrAuthCohortFenced
	}
	if err := ctx.Err(); err != nil {
		c.mu.Unlock()
		return nil, err
	}
	c.nextID++
	if c.nextID == 0 {
		c.mu.Unlock()
		return nil, errors.New("authentication cohort admission identifiers are exhausted")
	}
	memberCtx, cancel := context.WithCancel(ctx)
	admission := &AuthCohortAdmission{
		coordinator: c,
		cohort:      cohort,
		epoch:       health.Epoch,
		id:          c.nextID,
		ctx:         memberCtx,
		cancel:      cancel,
	}
	if c.members[cohort] == nil {
		c.members[cohort] = make(map[uint64]*AuthCohortAdmission)
	}
	c.members[cohort][admission.id] = admission
	c.mu.Unlock()
	return admission, nil
}

// Context is cancelled when the caller cancels, a verified invalid result
// fences the cohort, or a verified repair replaces the member's epoch.
func (a *AuthCohortAdmission) Context() context.Context {
	if a == nil || a.ctx == nil {
		return context.Background()
	}
	return a.ctx
}

// Epoch returns the member's immutable healthy epoch.
func (a *AuthCohortAdmission) Epoch() AuthCohortEpoch {
	if a == nil {
		return 0
	}
	return a.epoch
}

// Check is the send-admission boundary for an active member. It reads current
// persisted health while holding the coordinator lock, so no admission begun
// after a successful verified fence can pass.
func (a *AuthCohortAdmission) Check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if a == nil || a.coordinator == nil {
		return ErrAuthCohortAdmissionClosed
	}
	c := a.coordinator
	c.mu.Lock()
	if c.members[a.cohort] == nil || c.members[a.cohort][a.id] != a {
		c.mu.Unlock()
		return ErrAuthCohortAdmissionClosed
	}
	health, err := c.healthLocked(a.cohort)
	c.mu.Unlock()
	if err != nil {
		return err
	}
	if health.Epoch != a.epoch {
		return ErrAuthCohortEpochMismatch
	}
	if health.Fenced {
		return ErrAuthCohortFenced
	}
	if err := a.Context().Err(); err != nil {
		return err
	}
	return nil
}

// Report applies an authentication outcome from one active member. Outcomes
// other than verified-invalid are deliberately observational: they cannot
// alter health, cancel peers, or reject future admission.
func (c *AuthCohortCoordinator) Report(admission *AuthCohortAdmission, outcome AuthenticationOutcome) error {
	if c == nil || c.store == nil {
		return errors.New("authentication cohort coordinator is unavailable")
	}
	if admission == nil || admission.coordinator != c || admission.cohort == "" {
		return ErrAuthCohortAdmissionClosed
	}

	c.mu.Lock()
	if c.members[admission.cohort] == nil || c.members[admission.cohort][admission.id] != admission {
		c.mu.Unlock()
		return ErrAuthCohortAdmissionClosed
	}
	health, err := c.healthLocked(admission.cohort)
	if err != nil {
		c.mu.Unlock()
		return err
	}
	if health.Epoch != admission.epoch {
		c.mu.Unlock()
		return ErrAuthCohortEpochMismatch
	}
	if health.Fenced {
		c.mu.Unlock()
		return ErrAuthCohortFenced
	}
	if outcome != AuthenticationOutcomeVerifiedInvalid {
		c.mu.Unlock()
		return nil
	}
	health.Fenced = true
	health.LastFencedEpoch = health.Epoch
	if err := c.store.Store(admission.cohort, health); err != nil {
		c.mu.Unlock()
		return errAuthCohortHealthStoreUnavailable
	}
	cancellations := c.cancellationsLocked(admission.cohort)
	c.mu.Unlock()
	for _, cancel := range cancellations {
		cancel()
	}
	return nil
}

// Repair starts a new healthy epoch only after a caller supplies the typed
// verified-healthy result from a dedicated credential repair or test path.
// The previous epoch's active members are cancelled and subsequently refused.
func (c *AuthCohortCoordinator) Repair(cohort connectors.AuthCohortKey, outcome AuthenticationOutcome) (AuthCohortEpoch, error) {
	if c == nil || c.store == nil {
		return 0, errors.New("authentication cohort coordinator is unavailable")
	}
	if cohort == "" {
		return 0, errors.New("authentication cohort is unavailable")
	}
	if outcome != AuthenticationOutcomeVerifiedHealthy {
		return 0, errors.New("authentication cohort repair is not verified")
	}

	c.mu.Lock()
	health, err := c.healthLocked(cohort)
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	if health.Epoch == ^AuthCohortEpoch(0) {
		c.mu.Unlock()
		return 0, errors.New("authentication cohort epochs are exhausted")
	}
	health.Epoch++
	health.Fenced = false
	if err := c.store.Store(cohort, health); err != nil {
		c.mu.Unlock()
		return 0, errAuthCohortHealthStoreUnavailable
	}
	cancellations := c.cancellationsLocked(cohort)
	c.mu.Unlock()
	for _, cancel := range cancellations {
		cancel()
	}
	return health.Epoch, nil
}

// Release removes a member from future cancellation fan-out. It is idempotent
// and cancels the member context so callers cannot accidentally reuse it.
func (a *AuthCohortAdmission) Release() {
	if a == nil || a.coordinator == nil {
		return
	}
	c := a.coordinator
	c.mu.Lock()
	if members := c.members[a.cohort]; members != nil && members[a.id] == a {
		delete(members, a.id)
		if len(members) == 0 {
			delete(c.members, a.cohort)
		}
	}
	c.mu.Unlock()
	a.cancel()
}

func (c *AuthCohortCoordinator) healthLocked(cohort connectors.AuthCohortKey) (AuthCohortHealth, error) {
	health, found, err := c.store.Load(cohort)
	if err != nil {
		return AuthCohortHealth{}, errAuthCohortHealthStoreUnavailable
	}
	if !found {
		health = AuthCohortHealth{Epoch: 1}
		if err := c.store.Store(cohort, health); err != nil {
			return AuthCohortHealth{}, errAuthCohortHealthStoreUnavailable
		}
		return health, nil
	}
	if health.Epoch == 0 || health.LastFencedEpoch > health.Epoch || (health.Fenced && health.LastFencedEpoch != health.Epoch) {
		return AuthCohortHealth{}, errors.New("authentication cohort health state is invalid")
	}
	return health, nil
}

func (c *AuthCohortCoordinator) cancellationsLocked(cohort connectors.AuthCohortKey) []context.CancelFunc {
	members := c.members[cohort]
	cancellations := make([]context.CancelFunc, 0, len(members))
	for _, admission := range members {
		cancellations = append(cancellations, admission.cancel)
	}
	return cancellations
}
