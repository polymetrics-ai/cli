package coordination

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/synccontract"
)

func TestRateParkingCoordinator_PersistsAcrossRestartAndResumesOnlyAfterReset(t *testing.T) {
	now := time.Date(2026, time.August, 15, 10, 0, 0, 0, time.UTC)
	resetAt := now.Add(5 * time.Minute)
	store := NewMemoryRateParkingStore()
	checkpoint := testParkedCheckpoint(now)
	scope := connectors.RateLimitScopeKey("scope-a")

	firstScheduler := newRateParkingTestScheduler()
	first := NewRateParkingCoordinator(RateParkingCoordinatorOptions{
		Store:     store,
		Scheduler: firstScheduler,
		Now:       func() time.Time { return now },
		Resume: func(context.Context, ParkedRateLimitRun) error {
			t.Fatal("original coordinator resumed before restart")
			return nil
		},
	})
	if err := first.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	parked, err := first.Park(context.Background(), RateParkingRequest{
		RunID:      "run-3867-a",
		Scope:      scope,
		Checkpoint: checkpoint,
		ResetAt:    resetAt,
		Reason:     connsdk.RateLimitObservationSourceRetryAfter,
	})
	if err != nil {
		t.Fatalf("Park() error = %v", err)
	}
	if parked.Outcome != RateParkingOutcomeParkedRateLimit {
		t.Fatalf("parked outcome = %q, want %q", parked.Outcome, RateParkingOutcomeParkedRateLimit)
	}
	if !parked.ResetAt.Equal(resetAt) || parked.Reason != connsdk.RateLimitObservationSourceRetryAfter {
		t.Fatalf("parked evidence = %#v, want retry_after reset at %s", parked, resetAt)
	}
	if err := first.Admit(scope); !errors.Is(err, ErrRateLimitParked) {
		t.Fatalf("Admit(parked scope) error = %v, want ErrRateLimitParked", err)
	}
	if err := first.Admit(connectors.RateLimitScopeKey("scope-b")); err != nil {
		t.Fatalf("Admit(unrelated scope) error = %v", err)
	}
	first.Close()

	// The checkpoint is a completed downstream acknowledgement. Resumption must
	// receive this exact value rather than re-running the acknowledged apply.
	acknowledgedApplies := 1
	var resumes int
	var resumed ParkedRateLimitRun
	secondScheduler := newRateParkingTestScheduler()
	second := NewRateParkingCoordinator(RateParkingCoordinatorOptions{
		Store:     store,
		Scheduler: secondScheduler,
		Now:       func() time.Time { return now },
		Resume: func(_ context.Context, got ParkedRateLimitRun) error {
			resumes++
			resumed = got
			return nil
		},
	})
	if err := second.Start(context.Background()); err != nil {
		t.Fatalf("restarted Start() error = %v", err)
	}
	now = resetAt.Add(-time.Nanosecond)
	secondScheduler.RunThrough(now)
	if resumes != 0 {
		t.Fatalf("resumes before reset = %d, want 0", resumes)
	}
	if acknowledgedApplies != 1 {
		t.Fatalf("acknowledged destination applies before reset = %d, want 1", acknowledgedApplies)
	}
	now = resetAt
	secondScheduler.RunThrough(now)
	if resumes != 1 {
		t.Fatalf("resumes at reset = %d, want 1", resumes)
	}
	if acknowledgedApplies != 1 {
		t.Fatalf("acknowledged destination applies after resume = %d, want 1 (no replay)", acknowledgedApplies)
	}
	if !bytes.Equal(resumed.Checkpoint.Position.Primary, checkpoint.Position.Primary) ||
		!bytes.Equal(resumed.Checkpoint.Position.TieBreaker, checkpoint.Position.TieBreaker) ||
		resumed.Checkpoint.CommittedAt == nil || !resumed.Checkpoint.CommittedAt.Equal(*checkpoint.CommittedAt) {
		t.Fatalf("resumed checkpoint = %#v, want exact committed checkpoint %#v", resumed.Checkpoint, checkpoint)
	}
	if err := second.Admit(scope); err != nil {
		t.Fatalf("Admit(resumed scope) error = %v", err)
	}
	if records, err := store.List(); err != nil || len(records) != 0 {
		t.Fatalf("store after successful resume = %#v, %v; want no parked records", records, err)
	}
}

func TestRateParkingCoordinator_RearmsClaimedRunWithLatestCheckpoint(t *testing.T) {
	for _, tt := range []struct {
		name  string
		store func(*testing.T) RateParkingStore
	}{
		{name: "memory", store: func(t *testing.T) RateParkingStore {
			t.Helper()
			return NewMemoryRateParkingStore()
		}},
		{name: "file", store: func(t *testing.T) RateParkingStore {
			t.Helper()
			store, err := OpenFileRateParkingStore(filepath.Join(t.TempDir(), "rate-parking.json"))
			if err != nil {
				t.Fatal(err)
			}
			return store
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Date(2026, time.August, 20, 11, 0, 0, 0, time.UTC)
			firstReset := now.Add(time.Minute)
			secondReset := now.Add(2 * time.Minute)
			checkpoint := testParkedCheckpoint(now)
			latestCheckpoint := checkpoint.Clone()
			latestCheckpoint.Position.Primary = synccontract.OpaqueToken("latest")
			latestCheckpoint.Position.TieBreaker = synccontract.OpaqueToken("latest-tiebreaker")
			latestCommittedAt := now.Add(30 * time.Second)
			latestCheckpoint.CommittedAt = &latestCommittedAt
			scope := connectors.RateLimitScopeKey("scope-rearm")
			scheduler := newRateParkingTestScheduler()
			var coordinator *RateParkingCoordinator
			resumes := 0
			coordinator = NewRateParkingCoordinator(RateParkingCoordinatorOptions{
				Store:     tt.store(t),
				Scheduler: scheduler,
				Now:       func() time.Time { return now },
				Resume: func(ctx context.Context, parked ParkedRateLimitRun) error {
					resumes++
					if resumes != 1 {
						return nil
					}
					if _, err := coordinator.Rearm(ctx, RateParkingRequest{
						RunID:      parked.RunID,
						Scope:      parked.Scope,
						Checkpoint: latestCheckpoint,
						ResetAt:    secondReset,
						Reason:     connsdk.RateLimitObservationSourceHeaders,
					}); err != nil {
						return err
					}
					return ErrRateLimitRearmed
				},
			})
			if err := coordinator.Start(context.Background()); err != nil {
				t.Fatalf("Start() error = %v", err)
			}
			if _, err := coordinator.Park(context.Background(), RateParkingRequest{
				RunID:      "run-rearm",
				Scope:      scope,
				Checkpoint: checkpoint,
				ResetAt:    firstReset,
				Reason:     connsdk.RateLimitObservationSourceRetryAfter,
			}); err != nil {
				t.Fatalf("Park() error = %v", err)
			}

			now = firstReset
			scheduler.RunThrough(now)
			if resumes != 1 {
				t.Fatalf("resume attempts after first reset = %d, want 1", resumes)
			}
			records, err := coordinator.store.List()
			if err != nil || len(records) != 1 {
				t.Fatalf("records after rearm = %#v, %v; want one record", records, err)
			}
			if !records[0].ResetAt.Equal(secondReset) || !bytes.Equal(records[0].Checkpoint.Position.Primary, latestCheckpoint.Position.Primary) || !bytes.Equal(records[0].Checkpoint.Position.TieBreaker, latestCheckpoint.Position.TieBreaker) || records[0].Checkpoint.CommittedAt == nil || !records[0].Checkpoint.CommittedAt.Equal(latestCommittedAt) {
				t.Fatalf("rearmed record = %#v, want latest checkpoint and reset", records[0])
			}
			if err := coordinator.Admit(scope); !errors.Is(err, ErrRateLimitParked) {
				t.Fatalf("Admit(rearmed scope) error = %v, want ErrRateLimitParked", err)
			}
			if scheduler.Scheduled() != 1 {
				t.Fatalf("scheduled rearm callbacks = %d, want 1", scheduler.Scheduled())
			}

			now = secondReset
			scheduler.RunThrough(now)
			if resumes != 2 {
				t.Fatalf("resume attempts after second reset = %d, want 2", resumes)
			}
			if records, err := coordinator.store.List(); err != nil || len(records) != 0 {
				t.Fatalf("records after successful rearmed resume = %#v, %v; want none", records, err)
			}
			if err := coordinator.Admit(scope); err != nil {
				t.Fatalf("Admit(resumed scope) error = %v", err)
			}
		})
	}
}

func TestRateParkingCoordinator_IsolatesScopesAndMakesDuplicateCancellationAndFailureObservable(t *testing.T) {
	now := time.Date(2026, time.August, 15, 11, 0, 0, 0, time.UTC)
	resetAt := now.Add(time.Minute)
	store := NewMemoryRateParkingStore()
	scheduler := newRateParkingTestScheduler()
	events := &rateParkingEventRecorder{}
	var resumes int
	coordinator := NewRateParkingCoordinator(RateParkingCoordinatorOptions{
		Store:     store,
		Scheduler: scheduler,
		Now:       func() time.Time { return now },
		Events:    events,
		Resume: func(_ context.Context, parked ParkedRateLimitRun) error {
			resumes++
			if parked.RunID == "run-fails" {
				return errors.New("resume fixture failure")
			}
			return nil
		},
	})
	if err := coordinator.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	request := RateParkingRequest{
		RunID:      "run-cancelled",
		Scope:      connectors.RateLimitScopeKey("scope-a"),
		Checkpoint: testParkedCheckpoint(now),
		ResetAt:    resetAt,
		Reason:     connsdk.RateLimitObservationSourceHeaders,
	}
	if _, err := coordinator.Park(context.Background(), request); err != nil {
		t.Fatalf("first Park() error = %v", err)
	}
	if _, err := coordinator.Park(context.Background(), request); err != nil {
		t.Fatalf("duplicate Park() error = %v", err)
	}
	if scheduler.Scheduled() != 1 {
		t.Fatalf("duplicate park schedules = %d, want 1", scheduler.Scheduled())
	}
	if err := coordinator.Cancel(request.RunID); err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	now = resetAt
	scheduler.RunThrough(now)
	if resumes != 0 {
		t.Fatalf("resumes after cancellation = %d, want 0", resumes)
	}
	if records, err := store.List(); err != nil || len(records) != 0 {
		t.Fatalf("store after cancellation = %#v, %v; want empty", records, err)
	}

	failing := request
	failing.RunID = "run-fails"
	failing.Scope = connectors.RateLimitScopeKey("scope-c")
	if _, err := coordinator.Park(context.Background(), failing); err != nil {
		t.Fatalf("Park(failing) error = %v", err)
	}
	if err := coordinator.Admit(connectors.RateLimitScopeKey("scope-d")); err != nil {
		t.Fatalf("Admit(unrelated scope) error = %v", err)
	}
	now = resetAt
	scheduler.RunThrough(resetAt)
	if resumes != 1 {
		t.Fatalf("failed callback resume attempts = %d, want 1", resumes)
	}
	records, err := store.List()
	if err != nil || len(records) != 1 || records[0].RunID != failing.RunID {
		t.Fatalf("failed callback persisted records = %#v, %v; want %q retained", records, err, failing.RunID)
	}
	if err := coordinator.Admit(failing.Scope); !errors.Is(err, ErrRateLimitParked) {
		t.Fatalf("Admit(failed parked scope) error = %v, want ErrRateLimitParked", err)
	}

	event := events.Event(RateLimitEventParked)
	if event.Reason != string(connsdk.RateLimitObservationSourceHeaders) || !event.ResetAt.Equal(resetAt) {
		t.Fatalf("park event = %#v, want headers reason and reset %s", event, resetAt)
	}
}

func TestRateParkingCoordinator_ConcurrentSameScopeAdmissionHasZeroSends(t *testing.T) {
	now := time.Date(2026, time.August, 15, 13, 0, 0, 0, time.UTC)
	store := NewMemoryRateParkingStore()
	coordinator := NewRateParkingCoordinator(RateParkingCoordinatorOptions{
		Store:     store,
		Scheduler: newRateParkingTestScheduler(),
		Now:       func() time.Time { return now },
		Resume:    func(context.Context, ParkedRateLimitRun) error { return nil },
	})
	if err := coordinator.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	scope := connectors.RateLimitScopeKey("scope-race")
	if _, err := coordinator.Park(context.Background(), RateParkingRequest{
		RunID:      "run-race",
		Scope:      scope,
		Checkpoint: testParkedCheckpoint(now),
		ResetAt:    now.Add(time.Minute),
		Reason:     connsdk.RateLimitObservationSourceHTTP429,
	}); err != nil {
		t.Fatalf("Park() error = %v", err)
	}

	var sends int
	var sendsMu sync.Mutex
	var group sync.WaitGroup
	for range 64 {
		group.Add(1)
		go func() {
			defer group.Done()
			if err := coordinator.Admit(scope); err == nil {
				sendsMu.Lock()
				sends++
				sendsMu.Unlock()
			}
		}()
	}
	group.Wait()
	if sends != 0 {
		t.Fatalf("same-scope sends during parked state = %d, want 0", sends)
	}
}

func testParkedCheckpoint(observedAt time.Time) synccontract.CheckpointEnvelope {
	committedAt := observedAt.Add(time.Second)
	positionObserved := true
	return synccontract.CheckpointEnvelope{
		StateVersion: synccontract.StateVersion,
		Source: synccontract.SourceIdentity{
			Engine:           "fixture-source",
			AccountOrCluster: "fixture-account",
			ObjectScope:      "records",
		},
		Mechanism:       "fixture-parking",
		SnapshotBarrier: &synccontract.SnapshotBarrier{Kind: "fixture", Token: synccontract.OpaqueToken("barrier")},
		Position: synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken{0x01, 0xff},
			TieBreaker: synccontract.OpaqueToken{0x02, 0x00},
		},
		PositionObserved: &positionObserved,
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: synccontract.OpaqueToken("generation"),
		SchemaVersion:    "v1",
		ProtocolVersion:  "v1",
		Dedupe:           synccontract.DedupeIdentity{Kind: "fixture", Value: synccontract.OpaqueToken("identity")},
		DedupeWindow:     synccontract.DedupeWindow{Kind: "fixture", Start: synccontract.OpaqueToken("start"), End: synccontract.OpaqueToken("end")},
		ObservedAt:       observedAt,
		CommittedAt:      &committedAt,
	}
}

type rateParkingEventRecorder struct {
	mu     sync.Mutex
	events []RateParkingEvent
}

func (r *rateParkingEventRecorder) RecordRateParkingEvent(event RateParkingEvent) {
	r.mu.Lock()
	r.events = append(r.events, event)
	r.mu.Unlock()
}

func (r *rateParkingEventRecorder) Event(kind RateParkingEventType) RateParkingEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, event := range r.events {
		if event.Type == kind {
			return event
		}
	}
	return RateParkingEvent{}
}

type rateParkingTestScheduler struct {
	mu    sync.Mutex
	next  uint64
	tasks map[uint64]*rateParkingTestTask
}

type rateParkingTestTask struct {
	at       time.Time
	callback func()
	stopped  bool
}

func newRateParkingTestScheduler() *rateParkingTestScheduler {
	return &rateParkingTestScheduler{tasks: make(map[uint64]*rateParkingTestTask)}
}

func (s *rateParkingTestScheduler) Schedule(at time.Time, callback func()) RateParkingTimer {
	s.mu.Lock()
	s.next++
	id := s.next
	s.tasks[id] = &rateParkingTestTask{at: at, callback: callback}
	s.mu.Unlock()
	return rateParkingTestTimer{scheduler: s, id: id}
}

func (s *rateParkingTestScheduler) Scheduled() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, task := range s.tasks {
		if !task.stopped {
			count++
		}
	}
	return count
}

func (s *rateParkingTestScheduler) RunThrough(now time.Time) {
	for {
		s.mu.Lock()
		var task *rateParkingTestTask
		var id uint64
		for candidateID, candidate := range s.tasks {
			if candidate.stopped || candidate.at.After(now) {
				continue
			}
			if task == nil || candidate.at.Before(task.at) || (candidate.at.Equal(task.at) && candidateID < id) {
				task = candidate
				id = candidateID
			}
		}
		if task != nil {
			delete(s.tasks, id)
		}
		s.mu.Unlock()
		if task == nil {
			return
		}
		task.callback()
	}
}

type rateParkingTestTimer struct {
	scheduler *rateParkingTestScheduler
	id        uint64
}

func (t rateParkingTestTimer) Stop() bool {
	if t.scheduler == nil {
		return false
	}
	t.scheduler.mu.Lock()
	defer t.scheduler.mu.Unlock()
	task, ok := t.scheduler.tasks[t.id]
	if !ok || task.stopped {
		return false
	}
	task.stopped = true
	return true
}
