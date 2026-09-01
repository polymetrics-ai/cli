package coordination

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// TestSharedTransportCoordinationContract keeps the shared authenticated
// scheduling and rate-parking contracts executable using local deterministic
// fakes.
func TestSharedTransportCoordinationContract(t *testing.T) {
	t.Run("verified authentication fences one cohort before a later fake send", func(t *testing.T) {
		coordinator := NewAuthCohortCoordinator(NewMemoryAuthCohortHealthStore())
		cohort := testAuthCohortKey(t, "transport-family-matrix")
		admission, err := coordinator.Admit(context.Background(), cohort)
		if err != nil {
			t.Fatalf("initial Admit() = %v", err)
		}
		defer admission.Release()

		fakeSends := 0
		if err := coordinator.Report(admission, AuthenticationOutcomeUnverifiedInvalid); err != nil {
			t.Fatalf("unverified Report() = %v", err)
		}
		if err := admission.Check(context.Background()); err != nil {
			t.Fatalf("unverified invalid outcome blocked fake send: %v", err)
		}
		fakeSends++

		if err := coordinator.Report(admission, AuthenticationOutcomeVerifiedInvalid); err != nil {
			t.Fatalf("verified-invalid Report() = %v", err)
		}
		if !errors.Is(admission.Check(context.Background()), ErrAuthCohortFenced) {
			t.Fatalf("fenced admission Check() = %v, want ErrAuthCohortFenced", admission.Check(context.Background()))
		}
		if _, err := coordinator.Admit(context.Background(), cohort); !errors.Is(err, ErrAuthCohortFenced) {
			t.Fatalf("post-fence Admit() = %T %v, want ErrAuthCohortFenced", err, err)
		}
		if got, want := fakeSends, 1; got != want {
			t.Fatalf("fake sends after verified fence = %d, want only the pre-fence unverified send", got)
		}
	})

	t.Run("parked scope resumes exact checkpoint without replay after restart", func(t *testing.T) {
		now := time.Date(2026, time.August, 17, 9, 0, 0, 0, time.UTC)
		resetAt := now.Add(time.Minute)
		store := NewMemoryRateParkingStore()
		scope := connectors.RateLimitScopeKey("transport-family-matrix-scope")
		checkpoint := testParkedCheckpoint(now)
		firstScheduler := newRateParkingTestScheduler()
		first := NewRateParkingCoordinator(RateParkingCoordinatorOptions{
			Store:     store,
			Scheduler: firstScheduler,
			Now:       func() time.Time { return now },
			Resume: func(context.Context, ParkedRateLimitRun) error {
				t.Fatal("pre-restart coordinator resumed unexpectedly")
				return nil
			},
		})
		if err := first.Start(context.Background()); err != nil {
			t.Fatalf("first Start() = %v", err)
		}
		request := RateParkingRequest{
			RunID:      "transport-family-matrix-run",
			Scope:      scope,
			Checkpoint: checkpoint,
			ResetAt:    resetAt,
			Reason:     connsdk.RateLimitObservationSourceHTTP429,
		}
		parked, err := first.Park(context.Background(), request)
		if err != nil {
			t.Fatalf("Park() = %v", err)
		}
		if got, want := parked.Outcome, RateParkingOutcomeParkedRateLimit; got != want {
			t.Fatalf("parked outcome = %q, want %q", got, want)
		}
		if err := first.Admit(scope); !errors.Is(err, ErrRateLimitParked) {
			t.Fatalf("parked Admit() = %T %v, want ErrRateLimitParked", err, err)
		}
		conflict := request
		conflict.ResetAt = resetAt.Add(time.Minute)
		if _, _, err := store.Create(ParkedRateLimitRun{
			RunID:      conflict.RunID,
			Outcome:    RateParkingOutcomeParkedRateLimit,
			Scope:      conflict.Scope,
			Checkpoint: conflict.Checkpoint,
			ResetAt:    conflict.ResetAt,
			Reason:     conflict.Reason,
		}); !errors.Is(err, ErrRateParkingConflict) {
			t.Fatalf("durable conflicting park = %T %v, want ErrRateParkingConflict", err, err)
		}
		if got, want := firstScheduler.Scheduled(), 1; got != want {
			t.Fatalf("conflicting park schedules = %d, want %d", got, want)
		}
		first.Close()

		resumeCalls := 0
		acknowledgedDestinationApplies := 1
		var resumed ParkedRateLimitRun
		secondScheduler := newRateParkingTestScheduler()
		second := NewRateParkingCoordinator(RateParkingCoordinatorOptions{
			Store:     store,
			Scheduler: secondScheduler,
			Now:       func() time.Time { return now },
			Resume: func(_ context.Context, run ParkedRateLimitRun) error {
				resumeCalls++
				resumed = run
				return nil
			},
		})
		if err := second.Start(context.Background()); err != nil {
			t.Fatalf("restart Start() = %v", err)
		}
		t.Cleanup(second.Close)

		now = resetAt.Add(-time.Nanosecond)
		secondScheduler.RunThrough(now)
		if resumeCalls != 0 || acknowledgedDestinationApplies != 1 {
			t.Fatalf("before-reset resume/applies = %d/%d, want 0/1", resumeCalls, acknowledgedDestinationApplies)
		}
		now = resetAt
		secondScheduler.RunThrough(now)
		if resumeCalls != 1 || acknowledgedDestinationApplies != 1 {
			t.Fatalf("after-reset resume/applies = %d/%d, want 1/1 without acknowledged replay", resumeCalls, acknowledgedDestinationApplies)
		}
		if !bytes.Equal(resumed.Checkpoint.Position.Primary, checkpoint.Position.Primary) || !bytes.Equal(resumed.Checkpoint.Position.TieBreaker, checkpoint.Position.TieBreaker) {
			t.Fatalf("resumed checkpoint = %#v, want original checkpoint = %#v", resumed.Checkpoint, checkpoint)
		}
		if err := second.Admit(scope); err != nil {
			t.Fatalf("resumed Admit() = %v", err)
		}
	})

	t.Run("cancellation removes a parked fake run before it can resume", func(t *testing.T) {
		now := time.Date(2026, time.August, 17, 10, 0, 0, 0, time.UTC)
		scheduler := newRateParkingTestScheduler()
		resumes := 0
		coordinator := NewRateParkingCoordinator(RateParkingCoordinatorOptions{
			Store:     NewMemoryRateParkingStore(),
			Scheduler: scheduler,
			Now:       func() time.Time { return now },
			Resume: func(context.Context, ParkedRateLimitRun) error {
				resumes++
				return nil
			},
		})
		if err := coordinator.Start(context.Background()); err != nil {
			t.Fatalf("Start() = %v", err)
		}
		t.Cleanup(coordinator.Close)
		request := RateParkingRequest{
			RunID:      "transport-family-cancelled",
			Scope:      connectors.RateLimitScopeKey("transport-family-cancelled-scope"),
			Checkpoint: testParkedCheckpoint(now),
			ResetAt:    now.Add(time.Minute),
			Reason:     connsdk.RateLimitObservationSourceHeaders,
		}
		if _, err := coordinator.Park(context.Background(), request); err != nil {
			t.Fatalf("Park() = %v", err)
		}
		if err := coordinator.Cancel(request.RunID); err != nil {
			t.Fatalf("Cancel() = %v", err)
		}
		now = request.ResetAt
		scheduler.RunThrough(now)
		if got, want := resumes, 0; got != want {
			t.Fatalf("cancelled run resumes = %d, want %d", got, want)
		}
	})
}
