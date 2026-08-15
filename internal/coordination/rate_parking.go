package coordination

import (
	"context"
	"errors"
	"fmt"
	"reflect"
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
	ErrRateLimitParked = errors.New("rate-limited work is parked")

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
	Store(ParkedRateLimitRun) error
	Delete(runID string) error
}

// MemoryRateParkingStore is a race-safe persistence seam for deterministic
// restart tests and dependency-free callers. Durable app-state wiring can
// implement RateParkingStore without changing coordinator behavior.
type MemoryRateParkingStore struct {
	mu   sync.RWMutex
	runs map[string]ParkedRateLimitRun
}

// NewMemoryRateParkingStore returns an empty usable parking store.
func NewMemoryRateParkingStore() *MemoryRateParkingStore {
	return &MemoryRateParkingStore{runs: make(map[string]ParkedRateLimitRun)}
}

// List returns defensive copies of all persisted parked runs.
func (s *MemoryRateParkingStore) List() ([]ParkedRateLimitRun, error) {
	if s == nil {
		return nil, errRateParkingUnavailable
	}
	s.mu.RLock()
	runs := make([]ParkedRateLimitRun, 0, len(s.runs))
	for _, run := range s.runs {
		runs = append(runs, run.Clone())
	}
	s.mu.RUnlock()
	return runs, nil
}

// Store replaces one parked run with a defensive copy.
func (s *MemoryRateParkingStore) Store(run ParkedRateLimitRun) error {
	if s == nil {
		return errRateParkingUnavailable
	}
	s.mu.Lock()
	if s.runs == nil {
		s.runs = make(map[string]ParkedRateLimitRun)
	}
	s.runs[run.RunID] = run.Clone()
	s.mu.Unlock()
	return nil
}

// Delete removes a parked run after cancellation or successful resume.
func (s *MemoryRateParkingStore) Delete(runID string) error {
	if s == nil {
		return errRateParkingUnavailable
	}
	s.mu.Lock()
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

	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	runs    map[string]ParkedRateLimitRun
	timers  map[string]RateParkingTimer
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
	return &RateParkingCoordinator{
		store:     options.Store,
		scheduler: options.Scheduler,
		now:       options.Now,
		resume:    options.Resume,
		events:    options.Events,
		runs:      make(map[string]ParkedRateLimitRun),
		timers:    make(map[string]RateParkingTimer),
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
	for runID := range c.runs {
		c.scheduleLocked(runID)
	}
	c.mu.Unlock()
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
	if err := c.store.Store(run); err != nil {
		c.mu.Unlock()
		return ParkedRateLimitRun{}, errRateParkingUnavailable
	}
	c.runs[run.RunID] = run.Clone()
	c.scheduleLocked(run.RunID)
	c.mu.Unlock()
	c.recordEvent(RateParkingEvent{Type: RateLimitEventParked, ResetAt: run.ResetAt, Reason: string(run.Reason)})
	return run.Clone(), nil
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
	for _, run := range c.runs {
		if run.Scope == scope {
			return ErrRateLimitParked
		}
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
	if err := c.store.Delete(runID); err != nil {
		c.mu.Unlock()
		return errRateParkingUnavailable
	}
	delete(c.runs, runID)
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
	delete(c.timers, runID)
	if c.now().Before(run.ResetAt) {
		c.scheduleLocked(runID)
		c.mu.Unlock()
		return
	}
	resumeCtx := c.ctx
	c.mu.Unlock()

	if err := c.resume(resumeCtx, run.Clone()); err != nil {
		return
	}

	c.mu.Lock()
	current, exists := c.runs[runID]
	if !exists || !parkedRateLimitRunEqual(current, run) {
		c.mu.Unlock()
		return
	}
	if err := c.store.Delete(runID); err != nil {
		c.mu.Unlock()
		return
	}
	delete(c.runs, runID)
	c.mu.Unlock()
	c.recordEvent(RateParkingEvent{Type: RateLimitEventResumed, ResetAt: run.ResetAt, Reason: string(run.Reason)})
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
