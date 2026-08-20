package coordination

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/synccontract"
)

// RateParkingOutcome is the closed durable outcome vocabulary for a run that
// must stop after a provider supplied a resumable rate-limit reset time.
type RateParkingOutcome string

const (
	RateParkingOutcomeParkedRateLimit RateParkingOutcome = "parked_rate_limit"
)

var (
	// ErrRateLimitParked rejects a new same-scope send while parked work has not
	// successfully resumed. It intentionally does not expose an opaque scope or
	// provider response detail.
	ErrRateLimitParked  = errors.New("rate-limited work is parked")
	ErrRateLimitRearmed = errors.New("rate-limited work was rearmed")

	errRateParkingUnavailable = errors.New("rate parking coordinator is unavailable")
	errRateParkingNotStarted  = errors.New("rate parking coordinator is not started")
)

// ParkedRateLimitRun is the secret-free durable state required to resume one
// run. Scope is an already opaque projection and must remain internal to a
// parking store; events and errors never include it.
type ParkedRateLimitRun struct {
	RunID      string
	Outcome    RateParkingOutcome
	Scope      connectors.RateLimitScopeKey
	Checkpoint synccontract.CheckpointEnvelope
	ResetAt    time.Time
	Reason     connsdk.RateLimitObservationSource
}

// Clone returns a defensive copy that preserves opaque checkpoint bytes.
func (r ParkedRateLimitRun) Clone() ParkedRateLimitRun {
	clone := r
	clone.Checkpoint = r.Checkpoint.Clone()
	return clone
}

// RateParkingRequest is an in-memory request to durably park a run. Engine
// creates it only from a typed rate-limit error with parsed reset evidence.
type RateParkingRequest struct {
	RunID      string
	Scope      connectors.RateLimitScopeKey
	Checkpoint synccontract.CheckpointEnvelope
	ResetAt    time.Time
	Reason     connsdk.RateLimitObservationSource
}

// RateParkingStore persists only opaque-scope run records. A coordinator owner
// must serialize concurrent live coordinators sharing one store; operations on
// one coordinator are safe for concurrent callers.
type RateParkingStore interface {
	List() ([]ParkedRateLimitRun, error)
	Create(ParkedRateLimitRun) (ParkedRateLimitRun, bool, error)
	Rearm(ParkedRateLimitRun, string, time.Time) (ParkedRateLimitRun, error)
	HasScope(connectors.RateLimitScopeKey) (bool, error)
	Claim(runID, owner string, now, until time.Time) (ParkedRateLimitRun, bool, time.Time, error)
	RenewClaim(runID, owner string, until time.Time) (bool, error)
	ReleaseClaim(runID, owner string) error
	Complete(runID, owner string) error
	Delete(runID string, now time.Time) error
}

// MemoryRateParkingStore is a race-safe persistence seam for deterministic
// restart tests and dependency-free callers. Durable app-state wiring can
// implement RateParkingStore without changing coordinator behavior.
type MemoryRateParkingStore struct {
	mu   sync.RWMutex
	runs map[string]rateParkingFileRecord
}

// NewMemoryRateParkingStore returns an empty usable parking store.
func NewMemoryRateParkingStore() *MemoryRateParkingStore {
	return &MemoryRateParkingStore{runs: make(map[string]rateParkingFileRecord)}
}

// List returns defensive copies of all persisted parked runs.
func (s *MemoryRateParkingStore) List() ([]ParkedRateLimitRun, error) {
	if s == nil {
		return nil, errRateParkingUnavailable
	}
	s.mu.RLock()
	runs := make([]ParkedRateLimitRun, 0, len(s.runs))
	for _, record := range s.runs {
		runs = append(runs, record.Run.Clone())
	}
	s.mu.RUnlock()
	sort.Slice(runs, func(i, j int) bool { return runs[i].RunID < runs[j].RunID })
	return runs, nil
}

func (s *MemoryRateParkingStore) Create(run ParkedRateLimitRun) (ParkedRateLimitRun, bool, error) {
	if s == nil {
		return ParkedRateLimitRun{}, false, errRateParkingUnavailable
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.runs == nil {
		s.runs = make(map[string]rateParkingFileRecord)
	}
	if existing, found := s.runs[run.RunID]; found {
		if !parkedRateLimitRunEqual(existing.Run, run) {
			return ParkedRateLimitRun{}, false, ErrRateParkingConflict
		}
		return existing.Run.Clone(), false, nil
	}
	s.runs[run.RunID] = rateParkingFileRecord{Run: run.Clone()}
	return run.Clone(), true, nil
}

func (s *MemoryRateParkingStore) Rearm(run ParkedRateLimitRun, owner string, until time.Time) (ParkedRateLimitRun, error) {
	if s == nil {
		return ParkedRateLimitRun{}, errRateParkingUnavailable
	}
	if err := validateParkedRateLimitRun(run); err != nil {
		return ParkedRateLimitRun{}, err
	}
	if until.IsZero() {
		return ParkedRateLimitRun{}, errors.New("rate parking claim deadline is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record, found := s.runs[run.RunID]
	if !found || record.ClaimOwner != owner {
		return ParkedRateLimitRun{}, ErrRateParkingClaimLost
	}
	if record.Run.Scope != run.Scope {
		return ParkedRateLimitRun{}, ErrRateParkingConflict
	}
	record.Run = run.Clone()
	record.ClaimUntil = until.UTC()
	s.runs[run.RunID] = record
	return run.Clone(), nil
}

func (s *MemoryRateParkingStore) HasScope(scope connectors.RateLimitScopeKey) (bool, error) {
	if s == nil {
		return false, errRateParkingUnavailable
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, record := range s.runs {
		if record.Run.Scope == scope {
			return true, nil
		}
	}
	return false, nil
}

func (s *MemoryRateParkingStore) Claim(runID, owner string, now, until time.Time) (ParkedRateLimitRun, bool, time.Time, error) {
	if s == nil {
		return ParkedRateLimitRun{}, false, time.Time{}, errRateParkingUnavailable
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record, found := s.runs[runID]
	if !found {
		return ParkedRateLimitRun{}, false, time.Time{}, ErrRateParkingClaimLost
	}
	if record.ClaimOwner != "" && record.ClaimOwner != owner && record.ClaimUntil.After(now) {
		return record.Run.Clone(), false, record.ClaimUntil, nil
	}
	record.ClaimOwner = owner
	record.ClaimUntil = until.UTC()
	s.runs[runID] = record
	return record.Run.Clone(), true, time.Time{}, nil
}

func (s *MemoryRateParkingStore) RenewClaim(runID, owner string, until time.Time) (bool, error) {
	if s == nil {
		return false, errRateParkingUnavailable
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record, found := s.runs[runID]
	if !found || record.ClaimOwner != owner {
		return false, nil
	}
	record.ClaimUntil = until.UTC()
	s.runs[runID] = record
	return true, nil
}

func (s *MemoryRateParkingStore) ReleaseClaim(runID, owner string) error {
	if s == nil {
		return errRateParkingUnavailable
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record, found := s.runs[runID]
	if !found {
		return nil
	}
	if record.ClaimOwner != owner {
		return ErrRateParkingClaimLost
	}
	record.ClaimOwner = ""
	record.ClaimUntil = time.Time{}
	s.runs[runID] = record
	return nil
}

func (s *MemoryRateParkingStore) Complete(runID, owner string) error {
	if s == nil {
		return errRateParkingUnavailable
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record, found := s.runs[runID]
	if !found || record.ClaimOwner != owner {
		return ErrRateParkingClaimLost
	}
	delete(s.runs, runID)
	return nil
}

// Delete removes a parked run after cancellation or successful resume.
func (s *MemoryRateParkingStore) Delete(runID string, now time.Time) error {
	if s == nil {
		return errRateParkingUnavailable
	}
	s.mu.Lock()
	if record, found := s.runs[runID]; found && record.ClaimOwner != "" && record.ClaimUntil.After(now) {
		s.mu.Unlock()
		return ErrRateParkingClaimLost
	}
	delete(s.runs, runID)
	s.mu.Unlock()
	return nil
}

// RateParkingTimer owns one scheduled resumption. Stop is idempotent in the
// coordinator even when an underlying timer has already fired.
type RateParkingTimer interface {
	Stop() bool
}

// RateParkingScheduler is the narrow scheduling seam used to make reset
// boundaries deterministic in tests and to keep no timer alive after Close.
type RateParkingScheduler interface {
	Schedule(time.Time, func()) RateParkingTimer
}

type wallRateParkingScheduler struct{}

func (wallRateParkingScheduler) Schedule(at time.Time, callback func()) RateParkingTimer {
	return time.AfterFunc(time.Until(at), callback)
}

// RateParkingEventType is a closed, secret-free state-transition vocabulary.
type RateParkingEventType string

const (
	RateLimitEventParked  RateParkingEventType = "parked"
	RateLimitEventResumed RateParkingEventType = "resumed"
)

// RateParkingEvent is safe for operator output. It deliberately excludes run
// IDs, scope keys, provider URLs, headers, bodies, and credential material.
type RateParkingEvent struct {
	Type    RateParkingEventType
	ResetAt time.Time
	Reason  string
}

// RateParkingEventSink records a parking state transition synchronously. It is
// a reporting seam only and cannot affect admission, persistence, or resume.
type RateParkingEventSink interface {
	RecordRateParkingEvent(RateParkingEvent)
}

// RateParkingResumeFunc restarts source execution from the supplied committed
// checkpoint. It must not replay an already acknowledged destination apply.
type RateParkingResumeFunc func(context.Context, ParkedRateLimitRun) error

// RateParkingCoordinatorOptions configures a connector-neutral parking owner.
type RateParkingCoordinatorOptions struct {
	Store     RateParkingStore
	Scheduler RateParkingScheduler
	Now       func() time.Time
	Resume    RateParkingResumeFunc
	Events    RateParkingEventSink
	ClaimTTL  time.Duration
}

// RateParkingCoordinator blocks same-scope admission while a persisted run is
// parked, then automatically invokes its resumer at or after the stored reset.
// It never interprets provider data or performs a connector request itself.
type RateParkingCoordinator struct {
	mu        sync.Mutex
	store     RateParkingStore
	scheduler RateParkingScheduler
	now       func() time.Time
	resume    RateParkingResumeFunc
	events    RateParkingEventSink
	owner     string
	claimTTL  time.Duration

	ctx          context.Context
	cancel       context.CancelFunc
	started      bool
	runs         map[string]ParkedRateLimitRun
	timers       map[string]RateParkingTimer
	resuming     map[string]uint64
	rearmPending map[string]uint64
	nextLease    uint64
}

// NewRateParkingCoordinator constructs a coordinator. A nil store/scheduler
// uses dependency-free in-memory/wall-clock defaults; a nil resume function is
// refused when Start is called so parked work cannot disappear silently.
func NewRateParkingCoordinator(options RateParkingCoordinatorOptions) *RateParkingCoordinator {
	if options.Store == nil {
		options.Store = NewMemoryRateParkingStore()
	}
	if options.Scheduler == nil {
		options.Scheduler = wallRateParkingScheduler{}
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.ClaimTTL <= 0 {
		options.ClaimTTL = 30 * time.Second
	}
	return &RateParkingCoordinator{
		store:        options.Store,
		scheduler:    options.Scheduler,
		now:          options.Now,
		resume:       options.Resume,
		events:       options.Events,
		owner:        newRateParkingOwner(),
		claimTTL:     options.ClaimTTL,
		runs:         make(map[string]ParkedRateLimitRun),
		timers:       make(map[string]RateParkingTimer),
		resuming:     make(map[string]uint64),
		rearmPending: make(map[string]uint64),
	}
}

// Start reloads durable parked records and re-arms their automatic resumption.
// The caller owns ctx: cancelling it pauses in-process scheduling without
// deleting persisted records, so a later restart can resume safely.
func (c *RateParkingCoordinator) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if c == nil || c.store == nil || c.scheduler == nil || c.now == nil || c.resume == nil {
		return errRateParkingUnavailable
	}
	runs, err := c.store.List()
	if err != nil {
		return errRateParkingUnavailable
	}
	c.mu.Lock()
	if c.started {
		c.mu.Unlock()
		return nil
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.started = true
	c.runs = make(map[string]ParkedRateLimitRun)
	c.timers = make(map[string]RateParkingTimer)
	c.resuming = make(map[string]uint64)
	c.rearmPending = make(map[string]uint64)
	for _, run := range runs {
		if err := validateParkedRateLimitRun(run); err != nil {
			c.cancel()
			c.runs = make(map[string]ParkedRateLimitRun)
			c.started = false
			c.mu.Unlock()
			return err
		}
		c.runs[run.RunID] = run.Clone()
	}
	due := make([]string, 0, len(c.runs))
	for runID, run := range c.runs {
		if !c.now().Before(run.ResetAt) {
			due = append(due, runID)
			continue
		}
		c.scheduleLocked(runID)
	}
	c.mu.Unlock()
	// Due work is claimed synchronously during production composition. A
	// short-lived CLI command therefore cannot exit before crash recovery has
	// had an opportunity to observe the durable record.
	for _, runID := range due {
		c.resumeDue(runID)
	}
	return nil
}

// Close stops in-process timers but preserves parked records for a later
// Start/restart. It is safe to call more than once.
func (c *RateParkingCoordinator) Close() {
	if c == nil {
		return
	}
	c.mu.Lock()
	for _, timer := range c.timers {
		if timer != nil {
			timer.Stop()
		}
	}
	c.timers = make(map[string]RateParkingTimer)
	c.resuming = make(map[string]uint64)
	c.rearmPending = make(map[string]uint64)
	if c.cancel != nil {
		c.cancel()
	}
	c.cancel = nil
	c.ctx = nil
	c.started = false
	c.mu.Unlock()
}

// Park persists one typed rate-limit outcome before same-scope admission can
// observe it, then schedules automatic resumption. Duplicate identical parks
// are idempotent and never create another scheduled callback.
func (c *RateParkingCoordinator) Park(ctx context.Context, request RateParkingRequest) (ParkedRateLimitRun, error) {
	if err := ctx.Err(); err != nil {
		return ParkedRateLimitRun{}, err
	}
	if c == nil || c.store == nil {
		return ParkedRateLimitRun{}, errRateParkingUnavailable
	}
	run, err := parkedRateLimitRunFromRequest(request)
	if err != nil {
		return ParkedRateLimitRun{}, err
	}

	c.mu.Lock()
	if !c.started {
		c.mu.Unlock()
		return ParkedRateLimitRun{}, errRateParkingNotStarted
	}
	if existing, found := c.runs[run.RunID]; found {
		if parkedRateLimitRunEqual(existing, run) {
			c.mu.Unlock()
			return existing.Clone(), nil
		}
		c.mu.Unlock()
		return ParkedRateLimitRun{}, errors.New("rate-limited run is already parked with different evidence")
	}
	persisted, _, err := c.store.Create(run)
	if err != nil {
		c.mu.Unlock()
		return ParkedRateLimitRun{}, err
	}
	c.runs[run.RunID] = persisted.Clone()
	c.scheduleLocked(run.RunID)
	c.mu.Unlock()
	c.recordEvent(RateParkingEvent{Type: RateLimitEventParked, ResetAt: run.ResetAt, Reason: string(run.Reason)})
	return run.Clone(), nil
}

func (c *RateParkingCoordinator) Rearm(ctx context.Context, request RateParkingRequest) (ParkedRateLimitRun, error) {
	if err := ctx.Err(); err != nil {
		return ParkedRateLimitRun{}, err
	}
	if c == nil || c.store == nil {
		return ParkedRateLimitRun{}, errRateParkingUnavailable
	}
	run, err := parkedRateLimitRunFromRequest(request)
	if err != nil {
		return ParkedRateLimitRun{}, err
	}
	c.mu.Lock()
	if !c.started {
		c.mu.Unlock()
		return ParkedRateLimitRun{}, errRateParkingNotStarted
	}
	existing, found := c.runs[run.RunID]
	if !found {
		c.mu.Unlock()
		return ParkedRateLimitRun{}, ErrRateParkingClaimLost
	}
	if existing.Scope != run.Scope {
		c.mu.Unlock()
		return ParkedRateLimitRun{}, ErrRateParkingConflict
	}
	persisted, err := c.store.Rearm(run, c.owner, c.now().Add(c.claimTTL))
	if err != nil {
		c.mu.Unlock()
		return ParkedRateLimitRun{}, err
	}
	if timer := c.timers[run.RunID]; timer != nil {
		timer.Stop()
	}
	delete(c.timers, run.RunID)
	c.runs[run.RunID] = persisted.Clone()
	if lease, resuming := c.resuming[run.RunID]; resuming {
		c.rearmPending[run.RunID] = lease
	} else {
		c.scheduleLocked(run.RunID)
	}
	c.mu.Unlock()
	c.recordEvent(RateParkingEvent{Type: RateLimitEventParked, ResetAt: persisted.ResetAt, Reason: string(persisted.Reason)})
	return persisted.Clone(), nil
}

// Admit refuses a same-scope send while any parked run awaits a successful
// resume. Different opaque scopes remain independently admissible.
func (c *RateParkingCoordinator) Admit(scope connectors.RateLimitScopeKey) error {
	if c == nil || c.store == nil {
		return errRateParkingUnavailable
	}
	if scope == "" {
		return errors.New("rate-limit scope is unavailable")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.started {
		return errRateParkingNotStarted
	}
	parked, err := c.store.HasScope(scope)
	if err != nil {
		return errRateParkingUnavailable
	}
	if parked {
		return ErrRateLimitParked
	}
	return nil
}

// Cancel removes a parked run and stops its callback. A racing callback checks
// the in-memory record again before emitting a resumed event or deleting state.
func (c *RateParkingCoordinator) Cancel(runID string) error {
	if c == nil || c.store == nil {
		return errRateParkingUnavailable
	}
	c.mu.Lock()
	if _, found := c.runs[runID]; !found {
		c.mu.Unlock()
		return errors.New("parked rate-limit run is unavailable")
	}
	if timer := c.timers[runID]; timer != nil {
		timer.Stop()
	}
	delete(c.timers, runID)
	if err := c.store.Delete(runID, c.now()); err != nil {
		c.mu.Unlock()
		return errRateParkingUnavailable
	}
	delete(c.runs, runID)
	delete(c.resuming, runID)
	delete(c.rearmPending, runID)
	c.mu.Unlock()
	return nil
}

func (c *RateParkingCoordinator) scheduleLocked(runID string) {
	if _, exists := c.timers[runID]; exists || !c.started {
		return
	}
	run, exists := c.runs[runID]
	if !exists {
		return
	}
	c.timers[runID] = c.scheduler.Schedule(run.ResetAt, func() { c.resumeDue(runID) })
}

func (c *RateParkingCoordinator) resumeDue(runID string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	if !c.started || c.ctx == nil || c.ctx.Err() != nil {
		c.mu.Unlock()
		return
	}
	run, exists := c.runs[runID]
	if !exists {
		c.mu.Unlock()
		return
	}
	if _, resuming := c.resuming[runID]; resuming {
		c.mu.Unlock()
		return
	}
	delete(c.timers, runID)
	if c.now().Before(run.ResetAt) {
		c.scheduleLocked(runID)
		c.mu.Unlock()
		return
	}
	c.nextLease++
	if c.nextLease == 0 {
		c.nextLease++
	}
	lease := c.nextLease
	c.resuming[runID] = lease
	resumeCtx := c.ctx
	c.mu.Unlock()

	claimedRun, claimed, retryAt, err := c.store.Claim(runID, c.owner, c.now(), c.now().Add(c.claimTTL))
	if err != nil {
		c.finishResumeLease(runID, lease)
		return
	}
	if !claimed {
		c.mu.Lock()
		owned, rearmed := c.finishResumeLeaseLocked(runID, lease)
		if owned && c.started && !rearmed {
			c.timers[runID] = c.scheduler.Schedule(retryAt, func() { c.resumeDue(runID) })
		}
		c.mu.Unlock()
		return
	}
	run = claimedRun
	operationCtx, cancelOperation := context.WithCancel(resumeCtx)
	renewDone := make(chan struct{})
	renewResult := make(chan bool, 1)
	go c.renewClaim(operationCtx, runID, renewDone, renewResult, cancelOperation)
	resumeErr := c.resume(operationCtx, run.Clone())
	close(renewDone)
	claimLost := <-renewResult
	cancelOperation()
	if errors.Is(resumeErr, ErrRateLimitRearmed) {
		c.finishResumeLease(runID, lease)
		return
	}
	if claimLost {
		c.finishResumeLease(runID, lease)
		return
	}
	if resumeErr != nil {
		_ = c.store.ReleaseClaim(runID, c.owner)
		c.finishResumeLease(runID, lease)
		return
	}

	c.mu.Lock()
	current, exists := c.runs[runID]
	if !exists || !parkedRateLimitRunEqual(current, run) {
		c.finishResumeLeaseLocked(runID, lease)
		c.mu.Unlock()
		return
	}
	if err := c.store.Complete(runID, c.owner); err != nil {
		c.finishResumeLeaseLocked(runID, lease)
		c.mu.Unlock()
		return
	}
	delete(c.runs, runID)
	c.finishResumeLeaseLocked(runID, lease)
	c.mu.Unlock()
	c.recordEvent(RateParkingEvent{Type: RateLimitEventResumed, ResetAt: run.ResetAt, Reason: string(run.Reason)})
}

func (c *RateParkingCoordinator) finishResumeLease(runID string, lease uint64) {
	c.mu.Lock()
	c.finishResumeLeaseLocked(runID, lease)
	c.mu.Unlock()
}

func (c *RateParkingCoordinator) finishResumeLeaseLocked(runID string, lease uint64) (bool, bool) {
	if c.resuming[runID] != lease {
		return false, false
	}
	delete(c.resuming, runID)
	if c.rearmPending[runID] != lease {
		return true, false
	}
	delete(c.rearmPending, runID)
	c.scheduleLocked(runID)
	return true, true
}

func (c *RateParkingCoordinator) renewClaim(ctx context.Context, runID string, done <-chan struct{}, result chan<- bool, cancel context.CancelFunc) {
	interval := c.claimTTL / 3
	if interval <= 0 {
		interval = time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			result <- false
			return
		case <-ctx.Done():
			result <- false
			return
		case <-ticker.C:
			renewed, err := c.store.RenewClaim(runID, c.owner, c.now().Add(c.claimTTL))
			if err != nil || !renewed {
				cancel()
				result <- true
				return
			}
		}
	}
}

func newRateParkingOwner() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err == nil {
		return hex.EncodeToString(value[:])
	}
	return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
}

func (c *RateParkingCoordinator) recordEvent(event RateParkingEvent) {
	if c != nil && c.events != nil {
		c.events.RecordRateParkingEvent(event)
	}
}

func validateParkedRateLimitRun(run ParkedRateLimitRun) error {
	if run.RunID == "" || len(run.RunID) > 256 {
		return errors.New("parked rate-limit run identifier is invalid")
	}
	if run.Scope == "" {
		return errors.New("parked rate-limit scope is unavailable")
	}
	if run.Outcome != RateParkingOutcomeParkedRateLimit {
		return errors.New("parked rate-limit outcome is invalid")
	}
	if run.ResetAt.IsZero() {
		return errors.New("parked rate-limit reset time is unavailable")
	}
	if !rateParkingReasonValid(run.Reason) {
		return errors.New("parked rate-limit reason is invalid")
	}
	if run.Checkpoint.CommittedAt == nil {
		return errors.New("parked rate-limit checkpoint is not committed")
	}
	if err := run.Checkpoint.Validate(); err != nil {
		return fmt.Errorf("parked rate-limit checkpoint: %w", err)
	}
	return nil
}

func parkedRateLimitRunFromRequest(request RateParkingRequest) (ParkedRateLimitRun, error) {
	run := ParkedRateLimitRun{
		RunID:      request.RunID,
		Outcome:    RateParkingOutcomeParkedRateLimit,
		Scope:      request.Scope,
		Checkpoint: request.Checkpoint.Clone(),
		ResetAt:    request.ResetAt.UTC(),
		Reason:     request.Reason,
	}
	if err := validateParkedRateLimitRun(run); err != nil {
		return ParkedRateLimitRun{}, err
	}
	return run, nil
}

func rateParkingReasonValid(reason connsdk.RateLimitObservationSource) bool {
	switch reason {
	case connsdk.RateLimitObservationSourceRetryAfter,
		connsdk.RateLimitObservationSourceHeaders,
		connsdk.RateLimitObservationSourceBody,
		connsdk.RateLimitObservationSourceHTTP429:
		return true
	default:
		return false
	}
}

func parkedRateLimitRunEqual(left, right ParkedRateLimitRun) bool {
	if left.RunID != right.RunID || left.Outcome != right.Outcome || left.Scope != right.Scope ||
		!left.ResetAt.Equal(right.ResetAt) || left.Reason != right.Reason {
		return false
	}
	return checkpointEnvelopeEqual(left.Checkpoint, right.Checkpoint)
}

func checkpointEnvelopeEqual(left, right synccontract.CheckpointEnvelope) bool {
	return reflect.DeepEqual(left.Clone(), right.Clone())
}
