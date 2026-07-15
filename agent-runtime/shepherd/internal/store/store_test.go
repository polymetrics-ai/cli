package store

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
)

func TestLeaseFencingAndOutboxIdempotency(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Unix(1_700_000_000, 0).UTC()
	first, err := db.AcquireLease(ctx, "run-1", "owner-a", now, time.Minute)
	if err != nil {
		t.Fatalf("acquire first lease: %v", err)
	}
	second, err := db.AcquireLease(ctx, "run-1", "owner-b", now.Add(2*time.Minute), time.Minute)
	if err != nil {
		t.Fatalf("acquire second lease: %v", err)
	}
	if second.Epoch <= first.Epoch {
		t.Fatalf("fencing epoch did not increase: first=%d second=%d", first.Epoch, second.Epoch)
	}
	if err := db.CheckLease(ctx, first, now.Add(2*time.Minute)); err == nil {
		t.Fatal("expected stale lease to fail")
	}

	grant, err := domain.NewGrant("run-1", "repo", 372, domain.CapabilityPRUpdate, second.Epoch)
	if err != nil {
		t.Fatalf("grant: %v", err)
	}
	if err := db.PutGrant(ctx, grant); err != nil {
		t.Fatalf("put grant: %v", err)
	}

	effect := Effect{
		Key: "pr:372:ready:abc123", RunID: "run-1", Repository: "repo", Issue: 372,
		Capability: domain.CapabilityPRUpdate, Target: "pr:380", PayloadHash: "sha256:" + strings.Repeat("d", 64), Epoch: second.Epoch,
	}
	inserted, err := db.Enqueue(ctx, second, effect, now.Add(2*time.Minute))
	if err != nil || !inserted {
		t.Fatalf("first enqueue: inserted=%v err=%v", inserted, err)
	}
	inserted, err = db.Enqueue(ctx, second, effect, now.Add(2*time.Minute))
	if err != nil || inserted {
		t.Fatalf("duplicate enqueue: inserted=%v err=%v", inserted, err)
	}
	claimed, err := db.ClaimEffect(ctx, second, effect.Key, now.Add(2*time.Minute))
	if err != nil || claimed != effect {
		t.Fatalf("claim=%+v err=%v", claimed, err)
	}
	if _, err := db.ClaimEffect(ctx, second, effect.Key, now.Add(2*time.Minute)); err == nil {
		t.Fatal("claimed effect was claimed twice")
	}
	if err := db.MarkEffectSent(ctx, second, effect.Key, now.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	effect.PayloadHash = "sha256:" + strings.Repeat("e", 64)
	if _, err := db.Enqueue(ctx, second, effect, now.Add(2*time.Minute)); err == nil {
		t.Fatal("expected conflicting idempotency key to fail")
	}
}

func TestConcurrentLeaseAcquisitionHasOneWinner(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "shepherd.db")
	first, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	now := time.Unix(1_700_000_000, 0).UTC()
	type result struct{ err error }
	start := make(chan struct{})
	results := make(chan result, 2)
	for owner, database := range map[string]*Store{"a": first, "b": second} {
		go func(owner string, database *Store) {
			<-start
			_, err := database.AcquireLease(ctx, "run", owner, now, time.Minute)
			results <- result{err: err}
		}(owner, database)
	}
	close(start)
	winners := 0
	for range 2 {
		if (<-results).err == nil {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("lease winners=%d want 1", winners)
	}
}

func TestReleasedLeaseCanBeReacquiredWithHigherEpoch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(1_700_000_000, 0).UTC()
	first, err := db.AcquireLease(ctx, "issue-372", "owner-a", now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.ReleaseLease(ctx, first); err != nil {
		t.Fatal(err)
	}
	second, err := db.AcquireLease(ctx, "issue-372", "owner-b", now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if second.Epoch <= first.Epoch {
		t.Fatalf("epoch did not fence released owner: %d <= %d", second.Epoch, first.Epoch)
	}
}

func TestDeliveryBindingIsImmutable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	binding := testDelivery("issue-372", 372)
	if err := db.EnsureDelivery(ctx, binding); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if err := db.BindMilestone(ctx, binding.ID, "M001"); err != nil {
		t.Fatalf("bind milestone: %v", err)
	}
	if err := db.BindMilestone(ctx, binding.ID, "M002"); err == nil {
		t.Fatal("expected milestone rebind to fail")
	}
	binding.ContextHash = "sha256:" + strings.Repeat("b", 64)
	if err := db.EnsureDelivery(ctx, binding); err == nil {
		t.Fatal("expected context rebind to fail")
	}
}

func TestIssueProjectIdentityIsExclusiveAndImmutable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	root := filepath.Join(t.TempDir(), "wt-issue-389")
	binding := Delivery{
		ID: "issue-389", Issue: 389, ParentIssue: 372, WorkDir: root,
		ContextHash: "sha256:" + strings.Repeat("a", 64), Branch: "feat/389-autonomous-shepherd",
		BaseBranch: "feat/372-gsd-pi-go-shepherd", GSDProjectRoot: root,
		InitialHead: strings.Repeat("b", 40), GSDVersion: "1.11.0",
	}
	if err := db.EnsureDelivery(ctx, binding); err != nil {
		t.Fatal(err)
	}
	if err := db.EnsureDelivery(ctx, binding); err != nil {
		t.Fatalf("exact restart was not idempotent: %v", err)
	}

	other := binding
	other.ID, other.Issue = "issue-390", 390
	if err := db.EnsureDelivery(ctx, other); err == nil {
		t.Fatal("two issues shared one canonical GSD project root")
	}

	drifted := binding
	drifted.Branch = "feat/unrelated"
	if err := db.EnsureDelivery(ctx, drifted); err == nil {
		t.Fatal("same issue was rebound to a different branch")
	}
}

func TestDeliveryAttemptStateRequiresExplicitResume(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	binding := testDelivery("issue-372", 372)
	if err := db.EnsureDelivery(ctx, binding); err != nil {
		t.Fatal(err)
	}
	state, err := db.BeginAttempt(ctx, binding.ID, "owner-1")
	if err != nil || state.Attempt != 1 || state.Generation != 1 {
		t.Fatalf("begin=%+v err=%v", state, err)
	}
	if err := db.FinishAttempt(ctx, binding.ID, "owner-1", domain.RunBlocked); err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, binding.ID, "owner-2"); err == nil {
		t.Fatal("blocked delivery resumed without decision")
	}
	decision := domain.HumanDecision{RunID: binding.ID, Generation: 1, ActorKind: domain.ActorHuman, Approved: true}
	if err := db.ResumeDelivery(ctx, decision); err != nil {
		t.Fatal(err)
	}
	state, err = db.BeginAttempt(ctx, binding.ID, "owner-2")
	if err != nil || state.Generation != 2 || state.Attempt != 2 {
		t.Fatalf("resumed=%+v err=%v", state, err)
	}
}

func TestFailedBoundDeliveryRequiresExplicitResume(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	binding := testDelivery("issue-380", 380)
	if err := db.EnsureDelivery(ctx, binding); err != nil {
		t.Fatal(err)
	}
	if err := db.BindMilestone(ctx, binding.ID, "M001"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, binding.ID, "owner-1"); err != nil {
		t.Fatal(err)
	}
	if err := db.FinishAttempt(ctx, binding.ID, "owner-1", domain.RunFailed); err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, binding.ID, "owner-2"); err == nil {
		t.Fatal("failed delivery retried without a human decision")
	}
	decision := domain.HumanDecision{RunID: binding.ID, Generation: 1, ActorKind: domain.ActorHuman, Approved: true}
	if err := db.ResumeDelivery(ctx, decision); err != nil {
		t.Fatal(err)
	}
	run, err := db.BeginAttempt(ctx, binding.ID, "owner-2")
	if err != nil || run.Generation != 2 || run.Attempt != 2 {
		t.Fatalf("resumed failed delivery=%+v err=%v", run, err)
	}
}

func TestRetryFailedIntakeRequiresUnboundMilestone(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	binding := testDelivery("issue-380", 380)
	if err := db.EnsureDelivery(ctx, binding); err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, binding.ID, "owner-1"); err != nil {
		t.Fatal(err)
	}
	if err := db.FinishAttempt(ctx, binding.ID, "owner-1", domain.RunFailed); err != nil {
		t.Fatal(err)
	}
	if err := db.RetryFailedIntake(ctx, binding.ID); err != nil {
		t.Fatal(err)
	}
	state, err := db.BeginAttempt(ctx, binding.ID, "owner-2")
	if err != nil || state.Generation != 2 || state.Attempt != 2 {
		t.Fatalf("retry=%+v err=%v", state, err)
	}
	if err := db.FinishAttempt(ctx, binding.ID, "owner-2", domain.RunFailed); err != nil {
		t.Fatal(err)
	}
	if err := db.BindMilestone(ctx, binding.ID, "M001"); err != nil {
		t.Fatal(err)
	}
	if err := db.RetryFailedIntake(ctx, binding.ID); err == nil {
		t.Fatal("failed delivery with a bound milestone must not be reset")
	}
}

func TestPrepareAdoptedDeliveryResetsFailedBoundRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	binding := testDelivery("issue-380", 380)
	if err := db.EnsureDelivery(ctx, binding); err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, binding.ID, "owner-1"); err != nil {
		t.Fatal(err)
	}
	if err := db.BindMilestone(ctx, binding.ID, "M001"); err != nil {
		t.Fatal(err)
	}
	if err := db.FinishAttempt(ctx, binding.ID, "owner-1", domain.RunFailed); err != nil {
		t.Fatal(err)
	}
	if err := db.PrepareAdoptedDelivery(ctx, binding.ID, "M001"); err != nil {
		t.Fatal(err)
	}
	run, err := db.BeginAttempt(ctx, binding.ID, "owner-2")
	if err != nil || run.State != domain.RunRunning || run.Generation != 2 || run.Attempt != 2 {
		t.Fatalf("adopted run=%+v err=%v", run, err)
	}
}

func TestUnitAttemptIdentityPersistsAndBindsExactAttempt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "shepherd.db")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	key := UnitAttemptKey{DeliveryID: "identity-run", Generation: 1, UnitID: "execute-task/M001/S01/T01", HeadSHA: strings.Repeat("a", 40)}
	if err := db.EnsureDelivery(ctx, testDelivery(key.DeliveryID, 389)); err != nil {
		t.Fatal(err)
	}
	attempt, err := db.BeginUnitAttempt(ctx, key, 2)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now().UTC().Add(-time.Second)
	identity := UnitAttemptIdentity{UnitAttemptKey: key, Attempt: attempt.Attempts, Model: "openai-codex/gpt-5.5", Thinking: "high",
		SessionID: "019f5d4a-9fb4-7852-b640-d6fdf71bd3d9", SessionFingerprint: "sha256:" + strings.Repeat("b", 64), StartedAt: started, EndedAt: started.Add(time.Second)}
	if err := db.RecordUnitAttemptIdentity(ctx, identity); err != nil {
		t.Fatal(err)
	}
	if err := db.RecordUnitAttemptIdentity(ctx, identity); err == nil {
		t.Fatal("duplicate attempt identity was accepted")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	observed, err := db.GetUnitAttemptIdentity(ctx, key, attempt.Attempts)
	if err != nil {
		t.Fatal(err)
	}
	if observed.Model != identity.Model || observed.Thinking != "high" || observed.SessionID != identity.SessionID || observed.SessionFingerprint != identity.SessionFingerprint || !observed.StartedAt.Equal(identity.StartedAt) || !observed.EndedAt.Equal(identity.EndedAt) {
		t.Fatalf("persisted identity=%+v", observed)
	}
}

func TestUnitAttemptBudgetSurvivesStoreRestart(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "shepherd.db")
	key := UnitAttemptKey{
		DeliveryID: "issue-389", Generation: 1, UnitID: "plan-milestone/M001",
		HeadSHA: strings.Repeat("a", 40),
	}

	first, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.EnsureDelivery(ctx, testDelivery(key.DeliveryID, 389)); err != nil {
		t.Fatalf("ensure delivery: %v", err)
	}
	for want := int64(1); want <= 2; want++ {
		attempt, err := first.BeginUnitAttempt(ctx, key, 3)
		if err != nil || attempt.Attempts != want || attempt.Remaining != 3-want {
			t.Fatalf("attempt=%+v err=%v", attempt, err)
		}
		if err := first.FinishUnitAttempt(ctx, key, "interrupted"); err != nil {
			t.Fatal(err)
		}
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	attempt, err := second.BeginUnitAttempt(ctx, key, 3)
	if err != nil || attempt.Attempts != 3 || attempt.Remaining != 0 {
		t.Fatalf("reopened attempt=%+v err=%v", attempt, err)
	}
	if _, err := second.BeginUnitAttempt(ctx, key, 3); !errors.Is(err, ErrRetryBudgetExhausted) {
		t.Fatalf("error=%v, want durable retry exhaustion", err)
	}
}

func testDelivery(id string, issue int) Delivery {
	root := filepath.Join("/tmp", id)
	return Delivery{
		ID: id, Issue: issue, ParentIssue: 372, WorkDir: root,
		ContextHash: "sha256:" + strings.Repeat("a", 64), Branch: "feat/" + id,
		BaseBranch: "feat/372-gsd-pi-go-shepherd", GSDProjectRoot: root,
		InitialHead: strings.Repeat("b", 40), GSDVersion: "1.11.0",
	}
}
