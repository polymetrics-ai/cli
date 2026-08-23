package coordination

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

func TestRateParkingStores_RejectInvalidMutationWithoutStateChange(t *testing.T) {
	now := time.Date(2026, 8, 21, 8, 0, 0, 0, time.UTC)
	valid := ParkedRateLimitRun{RunID: "stable", Outcome: RateParkingOutcomeParkedRateLimit, Scope: "stable-scope", Checkpoint: testParkedCheckpoint(now), ResetAt: now, Reason: connsdk.RateLimitObservationSourceHeaders}

	t.Run("memory", func(t *testing.T) {
		store := NewMemoryRateParkingStore()
		if _, _, err := store.Create(valid); err != nil {
			t.Fatal(err)
		}
		before := make(map[string]rateParkingFileRecord, len(store.runs))
		for key, record := range store.runs {
			before[key] = rateParkingFileRecord{Run: record.Run.Clone(), ClaimOwner: record.ClaimOwner, ClaimUntil: record.ClaimUntil}
		}
		invalid := valid.Clone()
		invalid.RunID = ""
		if _, _, err := store.Create(invalid); err == nil {
			t.Fatal("memory store accepted invalid create")
		}
		if _, _, _, err := store.Claim(valid.RunID, "", now, now.Add(time.Minute)); err == nil {
			t.Fatal("memory store accepted blank claim owner")
		}
		if _, _, _, err := store.Claim(valid.RunID, "owner", now, now); err == nil {
			t.Fatal("memory store accepted non-forward claim deadline")
		}
		if !reflect.DeepEqual(store.runs, before) {
			t.Fatalf("rejected memory mutations changed state: before=%#v after=%#v", before, store.runs)
		}
	})

	t.Run("file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "stable.json")
		store, err := OpenFileRateParkingStore(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := store.Create(valid); err != nil {
			t.Fatal(err)
		}
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		invalid := valid.Clone()
		invalid.RunID = ""
		if _, _, err := store.Create(invalid); err == nil {
			t.Fatal("file store accepted invalid create")
		}
		if _, _, _, err := store.Claim(valid.RunID, "", now, now.Add(time.Minute)); err == nil {
			t.Fatal("file store accepted blank claim owner")
		}
		if _, _, _, err := store.Claim(valid.RunID, "owner", now, now); err == nil {
			t.Fatal("file store accepted non-forward claim deadline")
		}
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, after) {
			t.Fatal("rejected file mutations changed durable bytes")
		}
		reopened, err := OpenFileRateParkingStore(path)
		if err != nil {
			t.Fatal(err)
		}
		if runs, err := reopened.List(); err != nil || len(runs) != 1 || !parkedRateLimitRunEqual(runs[0], valid) {
			t.Fatalf("reopened state = %#v, %v; want original run", runs, err)
		}
	})
}

func TestFileRateParkingStoreEmptySingleLargeDuplicateAndOutOfOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "parking.json")
	store, err := OpenFileRateParkingStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if runs, err := store.List(); err != nil || len(runs) != 0 {
		t.Fatalf("empty store = %#v, %v", runs, err)
	}
	now := time.Date(2026, 8, 15, 1, 0, 0, 0, time.UTC)
	first := ParkedRateLimitRun{RunID: "single", Outcome: RateParkingOutcomeParkedRateLimit,
		Scope: "scope-single", Checkpoint: testParkedCheckpoint(now), ResetAt: now.Add(time.Minute), Reason: connsdk.RateLimitObservationSourceRetryAfter}
	if _, created, err := store.Create(first); err != nil || !created {
		t.Fatalf("single create = created %t, %v", created, err)
	}
	if _, created, err := store.Create(first); err != nil || created {
		t.Fatalf("duplicate create = created %t, %v; want idempotent", created, err)
	}
	conflict := first
	conflict.ResetAt = first.ResetAt.Add(-time.Second)
	if _, _, err := store.Create(conflict); !errors.Is(err, ErrRateParkingConflict) {
		t.Fatalf("out-of-order duplicate = %v, want ErrRateParkingConflict", err)
	}
	if runs, err := store.List(); err != nil || len(runs) != 1 || !runs[0].ResetAt.Equal(first.ResetAt) {
		t.Fatalf("conflicting duplicate mutated store = %#v, %v", runs, err)
	}
	for index := 0; index < 128; index++ {
		run := first.Clone()
		run.RunID = fmt.Sprintf("large-%03d", index)
		run.Scope = connectors.RateLimitScopeKey(fmt.Sprintf("scope-%03d", index))
		if _, created, err := store.Create(run); err != nil || !created {
			t.Fatalf("large create %d = created %t, %v", index, created, err)
		}
	}
	reopened, err := OpenFileRateParkingStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if runs, err := reopened.List(); err != nil || len(runs) != 129 {
		t.Fatalf("large store after reopen count = %d, %v; want 129", len(runs), err)
	}
}

func TestFileAuthCohortStoreEmptySingleAndLargeCollections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth.json")
	store, err := OpenFileAuthCohortHealthStore(path)
	if err != nil {
		t.Fatal(err)
	}
	missing := testAuthCohortKey(t, "empty")
	if health, found, err := store.Load(missing); err != nil || found || health != (AuthCohortHealth{}) {
		t.Fatalf("empty auth store = %+v found=%t err=%v", health, found, err)
	}
	for index := 0; index < 128; index++ {
		cohort := testAuthCohortKey(t, fmt.Sprintf("cohort-%03d", index))
		initial := AuthCohortHealth{Epoch: AuthCohortEpoch(index + 1)}
		if got, err := store.Initialize(cohort, initial); err != nil || got != initial {
			t.Fatalf("initialize auth cohort %d = %+v, %v", index, got, err)
		}
	}
	reopened, err := OpenFileAuthCohortHealthStore(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, index := range []int{0, 63, 127} {
		cohort := testAuthCohortKey(t, fmt.Sprintf("cohort-%03d", index))
		want := AuthCohortHealth{Epoch: AuthCohortEpoch(index + 1)}
		if got, found, err := reopened.Load(cohort); err != nil || !found || got != want {
			t.Fatalf("reopen auth cohort %d = %+v found=%t err=%v, want %+v", index, got, found, err, want)
		}
	}
}

func TestFileRateParkingStoreClaimExpiryPreventsDuplicateResumeAndRecoversDeadOwner(t *testing.T) {
	store, err := OpenFileRateParkingStore(filepath.Join(t.TempDir(), "parking.json"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 15, 2, 0, 0, 0, time.UTC)
	run := ParkedRateLimitRun{RunID: "claimed", Outcome: RateParkingOutcomeParkedRateLimit, Scope: "claim-scope",
		Checkpoint: testParkedCheckpoint(now), ResetAt: now, Reason: connsdk.RateLimitObservationSourceHeaders}
	if _, _, err := store.Create(run); err != nil {
		t.Fatal(err)
	}
	if _, claimed, _, err := store.Claim(run.RunID, "owner-one", now, now.Add(time.Minute)); err != nil || !claimed {
		t.Fatalf("first claim = %t, %v", claimed, err)
	}
	if _, claimed, retryAt, err := store.Claim(run.RunID, "owner-two", now.Add(time.Second), now.Add(time.Minute)); err != nil || claimed || !retryAt.Equal(now.Add(time.Minute)) {
		t.Fatalf("duplicate live claim = claimed %t retry %s err %v", claimed, retryAt, err)
	}
	if _, claimed, _, err := store.Claim(run.RunID, "owner-two", now.Add(time.Minute), now.Add(2*time.Minute)); err != nil || !claimed {
		t.Fatalf("claim after dead-owner expiry = %t, %v", claimed, err)
	}
	if err := store.Delete(run.RunID, now.Add(time.Minute)); !errors.Is(err, ErrRateParkingClaimLost) {
		t.Fatalf("cancellation during active claim = %v, want ErrRateParkingClaimLost", err)
	}
	if runs, _ := store.List(); len(runs) != 1 {
		t.Fatalf("active-claim cancellation deleted durable work: %#v", runs)
	}
	if err := store.Complete(run.RunID, "owner-one"); !errors.Is(err, ErrRateParkingClaimLost) {
		t.Fatalf("stale owner completion = %v, want ErrRateParkingClaimLost", err)
	}
	if runs, _ := store.List(); len(runs) != 1 {
		t.Fatalf("stale owner deleted durable work: %#v", runs)
	}
	if _, err := store.BeginResume(run.RunID, "owner-two"); err != nil {
		t.Fatalf("begin current owner resume: %v", err)
	}
	if _, err := store.MarkResumeCompleted(run.RunID, "owner-two"); err != nil {
		t.Fatalf("persist current owner resume completion: %v", err)
	}
	if err := store.Complete(run.RunID, "owner-two"); err != nil {
		t.Fatalf("current owner completion: %v", err)
	}
	if runs, _ := store.List(); len(runs) != 0 {
		t.Fatalf("current owner completion retained work: %#v", runs)
	}
}

func TestFileCoordinationStoresCancellationAndPermissionRefusalHaveZeroMutation(t *testing.T) {
	dir := t.TempDir()
	parkingPath := filepath.Join(dir, "parking.json")
	parkingStore, err := OpenFileRateParkingStore(parkingPath)
	if err != nil {
		t.Fatal(err)
	}
	coordinator := NewRateParkingCoordinator(RateParkingCoordinatorOptions{Store: parkingStore, Resume: func(context.Context, ParkedRateLimitRun) error { return nil }})
	if err := coordinator.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := coordinator.Park(cancelled, RateParkingRequest{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Park = %v, want context.Canceled", err)
	}
	if runs, _ := parkingStore.List(); len(runs) != 0 {
		t.Fatalf("cancelled Park wrote records: %#v", runs)
	}

	authPath := filepath.Join(dir, "auth.json")
	authStore, err := OpenFileAuthCohortHealthStore(authPath)
	if err != nil {
		t.Fatal(err)
	}
	auth := NewAuthCohortCoordinator(authStore)
	cohort := testAuthCohortKey(t, "permission-refusal")
	if _, err := auth.CurrentEpoch(context.Background(), cohort); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(authPath, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(authPath, 0o600) })
	_, openErr := OpenFileAuthCohortHealthStore(authPath)
	if openErr == nil {
		t.Skip("filesystem owner can read mode-000 files; permission refusal is unavailable")
	}
	var pathErr *os.PathError
	if !errors.As(openErr, &pathErr) {
		t.Fatalf("permission refusal type = %T %v, want *os.PathError", openErr, openErr)
	}
	if err := os.Chmod(authPath, 0o600); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("permission refusal mutated authentication health")
	}
}

func TestRateParkingResumeInterruptionRetainsCommittedCheckpoint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "parking.json")
	store, err := OpenFileRateParkingStore(path)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 15, 3, 0, 0, 0, time.UTC)
	run := ParkedRateLimitRun{RunID: "interrupted", Outcome: RateParkingOutcomeParkedRateLimit, Scope: "interrupt-scope",
		Checkpoint: testParkedCheckpoint(now), ResetAt: now, Reason: connsdk.RateLimitObservationSourceBody}
	if _, _, err := store.Create(run); err != nil {
		t.Fatal(err)
	}
	coordinator := NewRateParkingCoordinator(RateParkingCoordinatorOptions{Store: store, Now: func() time.Time { return now }, Resume: func(context.Context, ParkedRateLimitRun) error {
		return context.Canceled
	}})
	if err := coordinator.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	reopened, err := OpenFileRateParkingStore(path)
	if err != nil {
		t.Fatal(err)
	}
	runs, err := reopened.List()
	if err != nil || len(runs) != 1 || !checkpointEnvelopeEqual(runs[0].Checkpoint, run.Checkpoint) {
		t.Fatalf("interrupted resume state = %#v, %v; want original checkpoint", runs, err)
	}
	coordinator.Close()
	var resumed ParkedRateLimitRun
	restarted := NewRateParkingCoordinator(RateParkingCoordinatorOptions{Store: reopened, Now: func() time.Time { return now }, Resume: func(_ context.Context, got ParkedRateLimitRun) error {
		resumed = got.Clone()
		return nil
	}})
	if err := restarted.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(restarted.Close)
	if resumed.RunID != run.RunID || !checkpointEnvelopeEqual(resumed.Checkpoint, run.Checkpoint) {
		t.Fatalf("restart resumed %#v, want original durable run %#v", resumed, run)
	}
	if runs, err := reopened.List(); err != nil || len(runs) != 0 {
		t.Fatalf("restart after interrupted resume retained %#v, %v", runs, err)
	}
}

func TestRateParkingLongResumeRenewsOwnershipAgainstConcurrentProcess(t *testing.T) {
	store := NewMemoryRateParkingStore()
	now := time.Now().UTC()
	run := ParkedRateLimitRun{RunID: "long-resume", Outcome: RateParkingOutcomeParkedRateLimit, Scope: "long-scope",
		Checkpoint: testParkedCheckpoint(now.Add(-time.Second)), ResetAt: now.Add(-time.Millisecond), Reason: connsdk.RateLimitObservationSourceRetryAfter}
	if _, _, err := store.Create(run); err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	release := make(chan struct{})
	coordinator := NewRateParkingCoordinator(RateParkingCoordinatorOptions{Store: store, ClaimTTL: 30 * time.Millisecond, Resume: func(context.Context, ParkedRateLimitRun) error {
		close(entered)
		<-release
		return nil
	}})
	started := make(chan error, 1)
	go func() { started <- coordinator.Start(context.Background()) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("long resume did not start")
	}
	time.Sleep(80 * time.Millisecond)
	if _, claimed, retryAt, err := store.Claim(run.RunID, "intruding-process", time.Now(), time.Now().Add(time.Minute)); err != nil || claimed || retryAt.IsZero() {
		t.Fatalf("concurrent claim during renewed resume = claimed %t retry %s err %v", claimed, retryAt, err)
	}
	close(release)
	if err := <-started; err != nil {
		t.Fatal(err)
	}
	if runs, _ := store.List(); len(runs) != 0 {
		t.Fatalf("successful long resume retained records: %#v", runs)
	}
}

func TestFileAuthCohortStoreRejectsOutOfOrderEpochWithoutOverwrite(t *testing.T) {
	store, err := OpenFileAuthCohortHealthStore(filepath.Join(t.TempDir(), "auth.json"))
	if err != nil {
		t.Fatal(err)
	}
	cohort := testAuthCohortKey(t, "out-of-order-epoch")
	first, err := store.Initialize(cohort, AuthCohortHealth{Epoch: 1})
	if err != nil {
		t.Fatal(err)
	}
	current := AuthCohortHealth{Epoch: 2, LastFencedEpoch: 1}
	if swapped, err := store.CompareAndSwap(cohort, first, current); err != nil || !swapped {
		t.Fatalf("current epoch CAS = %t, %v", swapped, err)
	}
	if swapped, err := store.CompareAndSwap(cohort, first, AuthCohortHealth{Epoch: 1, Fenced: true, LastFencedEpoch: 1}); err != nil || swapped {
		t.Fatalf("out-of-order epoch CAS = %t, %v; want refused", swapped, err)
	}
	got, found, err := store.Load(cohort)
	if err != nil || !found || got != current {
		t.Fatalf("health after stale CAS = %+v, found=%t err=%v", got, found, err)
	}
}

func TestAuthCohortFenceIndeterminateCommitCancelsOldLocalEpoch(t *testing.T) {
	store, err := OpenFileAuthCohortHealthStore(filepath.Join(t.TempDir(), "auth.json"))
	if err != nil {
		t.Fatal(err)
	}
	cohort := testAuthCohortKey(t, "indeterminate-fence-cancellation")
	coordinator := NewAuthCohortCoordinator(store)
	admission, err := coordinator.Admit(context.Background(), cohort)
	if err != nil {
		t.Fatalf("Admit() error = %v", err)
	}
	defer admission.Release()
	store.store.SyncDirectory = func(string) error { return errors.New("directory sync failed") }

	err = coordinator.Fence(cohort, AuthenticationOutcomeVerifiedInvalid)
	if err == nil {
		t.Fatal("Fence() succeeded despite post-rename directory-sync failure")
	}
	select {
	case <-admission.Context().Done():
	case <-time.After(time.Second):
		t.Fatal("indeterminate durable fence left the old local epoch runnable")
	}
	health, found, loadErr := store.Load(cohort)
	if loadErr != nil || !found || !health.Fenced {
		t.Fatalf("reloaded health = %+v found=%t err=%v, want persisted fenced epoch", health, found, loadErr)
	}
}
