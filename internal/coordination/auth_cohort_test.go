package coordination

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
)

func testAuthCohortKey(t *testing.T, binding string) connectors.AuthCohortKey {
	t.Helper()
	identity, err := connectors.NewCoordinationIdentity([]byte("auth-cohort-fencing-test-salt"), connectors.CredentialBinding{
		BindingID:      binding,
		ProviderFamily: "fixture-provider",
		AuthProfile:    "fixture-profile",
	})
	if err != nil {
		t.Fatalf("NewCoordinationIdentity: %v", err)
	}
	return identity.AuthCohortKey()
}

func TestAuthCohortCoordinator_OnlyVerifiedInvalidAuthenticationFences(t *testing.T) {
	outcomes := []struct {
		name    string
		outcome AuthenticationOutcome
	}{
		{name: "unknown", outcome: AuthenticationOutcomeUnknown},
		{name: "unverified invalid", outcome: AuthenticationOutcomeUnverifiedInvalid},
		{name: "transport failure", outcome: AuthenticationOutcomeTransportFailure},
		{name: "timeout", outcome: AuthenticationOutcomeTimeout},
		{name: "provider failure", outcome: AuthenticationOutcomeProviderFailure},
		{name: "verified healthy", outcome: AuthenticationOutcomeVerifiedHealthy},
		{name: "unrecognized", outcome: AuthenticationOutcome(255)},
	}

	for _, tt := range outcomes {
		t.Run(tt.name, func(t *testing.T) {
			coordinator := NewAuthCohortCoordinator(NewMemoryAuthCohortHealthStore())
			cohort := testAuthCohortKey(t, "non-fencing-"+strings.ReplaceAll(tt.name, " ", "-"))
			member, err := coordinator.Admit(context.Background(), cohort)
			if err != nil {
				t.Fatalf("initial admission: %v", err)
			}
			defer member.Release()
			if err := coordinator.Report(member, tt.outcome); err != nil {
				t.Fatalf("report %s: %v", tt.name, err)
			}

			later, err := coordinator.Admit(context.Background(), cohort)
			if err != nil {
				t.Fatalf("admission after %s: %v; want healthy cohort", tt.name, err)
			}
			defer later.Release()
			if err := later.Check(context.Background()); err != nil {
				t.Fatalf("post-%s member check: %v; want send admission", tt.name, err)
			}
		})
	}
}

func TestAuthCohortCoordinator_VerifiedFailureCancelsSiblingsAndRejectsNewAdmissions(t *testing.T) {
	coordinator := NewAuthCohortCoordinator(NewMemoryAuthCohortHealthStore())
	cohort := testAuthCohortKey(t, "fenced-cohort")

	failing, err := coordinator.Admit(context.Background(), cohort)
	if err != nil {
		t.Fatalf("failing member admission: %v", err)
	}
	defer failing.Release()
	sibling, err := coordinator.Admit(context.Background(), cohort)
	if err != nil {
		t.Fatalf("sibling admission: %v", err)
	}
	defer sibling.Release()

	var siblingSends atomic.Int64
	siblingStopped := make(chan struct{})
	go func() {
		defer close(siblingStopped)
		<-sibling.Context().Done()
		if sibling.Check(context.Background()) == nil {
			siblingSends.Add(1)
		}
	}()

	if err := coordinator.Report(failing, AuthenticationOutcomeVerifiedInvalid); err != nil {
		t.Fatalf("report verified invalid authentication: %v", err)
	}
	select {
	case <-siblingStopped:
	case <-time.After(time.Second):
		t.Fatal("fenced sibling context was not cancelled")
	}
	if got := siblingSends.Load(); got != 0 {
		t.Fatalf("sibling sends after verified fence = %d, want 0", got)
	}
	if err := sibling.Check(context.Background()); !errors.Is(err, ErrAuthCohortFenced) {
		t.Fatalf("sibling check after fence = %v, want ErrAuthCohortFenced", err)
	}
	if _, err := coordinator.Admit(context.Background(), cohort); !errors.Is(err, ErrAuthCohortFenced) {
		t.Fatalf("new admission after fence = %v, want ErrAuthCohortFenced", err)
	}
}

func TestAuthCohortCoordinator_IsolatesCohortsAndRepairCreatesHealthyEpoch(t *testing.T) {
	store := NewMemoryAuthCohortHealthStore()
	coordinator := NewAuthCohortCoordinator(store)
	fencedCohort := testAuthCohortKey(t, "fenced")
	healthyCohort := testAuthCohortKey(t, "healthy")

	stale, err := coordinator.Admit(context.Background(), fencedCohort)
	if err != nil {
		t.Fatalf("fenced-cohort initial admission: %v", err)
	}
	defer stale.Release()
	if err := coordinator.Report(stale, AuthenticationOutcomeVerifiedInvalid); err != nil {
		t.Fatalf("fence cohort: %v", err)
	}

	healthy, err := coordinator.Admit(context.Background(), healthyCohort)
	if err != nil {
		t.Fatalf("unrelated cohort admission: %v", err)
	}
	defer healthy.Release()
	var unrelatedSends atomic.Int64
	if err := healthy.Check(context.Background()); err != nil {
		t.Fatalf("unrelated cohort check: %v", err)
	}
	unrelatedSends.Add(1)
	if got := unrelatedSends.Load(); got != 1 {
		t.Fatalf("unrelated cohort sends = %d, want 1", got)
	}

	newEpoch, err := coordinator.Repair(fencedCohort, AuthenticationOutcomeVerifiedHealthy)
	if err != nil {
		t.Fatalf("verified repair: %v", err)
	}
	if newEpoch <= stale.Epoch() {
		t.Fatalf("repair epoch = %d, stale member epoch = %d; want observable bump", newEpoch, stale.Epoch())
	}
	health, found, err := store.Load(fencedCohort)
	if err != nil || !found {
		t.Fatalf("load repaired cohort health = %+v, found=%t, err=%v", health, found, err)
	}
	if health.LastFencedEpoch != stale.Epoch() {
		t.Fatalf("repair lost fenced epoch evidence: got %d, want %d", health.LastFencedEpoch, stale.Epoch())
	}
	if health.Fenced {
		t.Fatal("verified repair left the new epoch fenced")
	}
	if err := stale.Check(context.Background()); !errors.Is(err, ErrAuthCohortEpochMismatch) {
		t.Fatalf("stale member check after repair = %v, want ErrAuthCohortEpochMismatch", err)
	}

	fresh, err := coordinator.Admit(context.Background(), fencedCohort)
	if err != nil {
		t.Fatalf("admission after verified repair: %v", err)
	}
	defer fresh.Release()
	if fresh.Epoch() != newEpoch {
		t.Fatalf("fresh member epoch = %d, want repaired epoch %d", fresh.Epoch(), newEpoch)
	}
	var freshSends atomic.Int64
	if err := fresh.Check(context.Background()); err != nil {
		t.Fatalf("fresh member check: %v", err)
	}
	freshSends.Add(1)
	if got := freshSends.Load(); got != 1 {
		t.Fatalf("fresh epoch sends = %d, want 1", got)
	}
}

func TestAuthCohortCoordinator_RepairRequiresVerifiedHealthyOutcome(t *testing.T) {
	coordinator := NewAuthCohortCoordinator(NewMemoryAuthCohortHealthStore())
	cohort := testAuthCohortKey(t, "repair-verification")
	member, err := coordinator.Admit(context.Background(), cohort)
	if err != nil {
		t.Fatalf("initial admission: %v", err)
	}
	defer member.Release()
	if err := coordinator.Report(member, AuthenticationOutcomeVerifiedInvalid); err != nil {
		t.Fatalf("verified fence: %v", err)
	}
	if _, err := coordinator.Repair(cohort, AuthenticationOutcomeUnverifiedInvalid); err == nil {
		t.Fatal("unverified repair outcome reopened a fenced cohort")
	}
	if _, err := coordinator.Admit(context.Background(), cohort); !errors.Is(err, ErrAuthCohortFenced) {
		t.Fatalf("admission after unverified repair = %v, want ErrAuthCohortFenced", err)
	}
}

func TestAuthCohortCoordinator_RestartAndRaceNeverAdmitAFencedCohort(t *testing.T) {
	store := NewMemoryAuthCohortHealthStore()
	cohort := testAuthCohortKey(t, "restart-race")
	firstCoordinator := NewAuthCohortCoordinator(store)
	member, err := firstCoordinator.Admit(context.Background(), cohort)
	if err != nil {
		t.Fatalf("initial admission: %v", err)
	}
	defer member.Release()
	if err := firstCoordinator.Report(member, AuthenticationOutcomeVerifiedInvalid); err != nil {
		t.Fatalf("verified fence: %v", err)
	}

	// A new coordinator models restart after the health transition was stored.
	restarted := NewAuthCohortCoordinator(store)
	var sends atomic.Int64
	var wg sync.WaitGroup
	for range 64 {
		wg.Go(func() {
			admission, err := restarted.Admit(context.Background(), cohort)
			if err != nil {
				if !errors.Is(err, ErrAuthCohortFenced) {
					t.Errorf("post-restart admission error = %v, want ErrAuthCohortFenced", err)
				}
				return
			}
			defer admission.Release()
			if err := admission.Check(context.Background()); err == nil {
				sends.Add(1)
			}
		})
	}
	wg.Wait()
	if got := sends.Load(); got != 0 {
		t.Fatalf("post-fence admissions that reached send boundary = %d, want 0", got)
	}

	newEpoch, err := restarted.Repair(cohort, AuthenticationOutcomeVerifiedHealthy)
	if err != nil {
		t.Fatalf("verified repair after restart: %v", err)
	}
	if newEpoch <= member.Epoch() {
		t.Fatalf("restarted repair epoch = %d, stale epoch = %d; want bump", newEpoch, member.Epoch())
	}
	if err := member.Check(context.Background()); !errors.Is(err, ErrAuthCohortEpochMismatch) {
		t.Fatalf("stale pre-restart member = %v, want ErrAuthCohortEpochMismatch", err)
	}
}

func TestAuthCohortRuntime_PublishesCurrentOwnershipAndRefusesStaleGeneration(t *testing.T) {
	store := NewMemoryAuthCohortHealthStore()
	cohort := testAuthCohortKey(t, "ownership-generation")
	first := NewAuthCohortCoordinator(store)
	staleRuntime, err := NewAuthCohortRuntime(context.Background(), first, cohort)
	if err != nil {
		t.Fatalf("resolve original ownership: %v", err)
	}
	member, err := first.Admit(context.Background(), cohort)
	if err != nil {
		t.Fatalf("admit original member: %v", err)
	}
	defer member.Release()
	if err := first.Report(member, AuthenticationOutcomeVerifiedInvalid); err != nil {
		t.Fatalf("persist original fence: %v", err)
	}

	restarted := NewAuthCohortCoordinator(store)
	if _, err := NewAuthCohortRuntime(context.Background(), restarted, cohort); !errors.Is(err, ErrAuthCohortFenced) {
		t.Fatalf("runtime before repair = %v, want ErrAuthCohortFenced", err)
	}
	if _, err := restarted.Repair(cohort, AuthenticationOutcomeVerifiedHealthy); err != nil {
		t.Fatalf("verified repair: %v", err)
	}
	currentRuntime, err := NewAuthCohortRuntime(context.Background(), restarted, cohort)
	if err != nil {
		t.Fatalf("resolve repaired ownership: %v", err)
	}
	var writes atomic.Int64
	if err := currentRuntime.Execute(context.Background(), func(context.Context) error {
		writes.Add(1)
		return nil
	}); err != nil {
		t.Fatalf("current generation execute: %v", err)
	}
	if got := writes.Load(); got != 1 {
		t.Fatalf("current generation writes = %d, want 1", got)
	}
	if err := staleRuntime.Execute(context.Background(), func(context.Context) error {
		writes.Add(1)
		return nil
	}); !errors.Is(err, ErrAuthCohortEpochMismatch) {
		t.Fatalf("stale generation execute = %v, want ErrAuthCohortEpochMismatch", err)
	}
	if got := writes.Load(); got != 1 {
		t.Fatalf("stale generation changed writes = %d, want 1", got)
	}
}

func TestAuthCohortRuntime_CancelledAdmissionHasZeroWritesAndNoHealthMutation(t *testing.T) {
	store := NewMemoryAuthCohortHealthStore()
	cohort := testAuthCohortKey(t, "cancelled-runtime")
	coordinator := NewAuthCohortCoordinator(store)
	runtime, err := NewAuthCohortRuntime(context.Background(), coordinator, cohort)
	if err != nil {
		t.Fatal(err)
	}
	before, found, err := store.Load(cohort)
	if err != nil || !found {
		t.Fatalf("load pre-cancellation health = %+v found=%t err=%v", before, found, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var writes atomic.Int64
	if err := runtime.Execute(ctx, func(context.Context) error {
		writes.Add(1)
		return nil
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled runtime = %v, want context.Canceled", err)
	}
	if got := writes.Load(); got != 0 {
		t.Fatalf("cancelled runtime writes = %d, want zero", got)
	}
	if after, found, err := store.Load(cohort); err != nil || !found || after != before {
		t.Fatalf("cancelled runtime health = %+v found=%t err=%v, want unchanged %+v", after, found, err, before)
	}
}
