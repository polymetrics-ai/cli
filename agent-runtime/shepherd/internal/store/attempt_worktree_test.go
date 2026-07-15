package store

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
)

func TestAttemptWorktreeLifecycleStatesPersistAcrossReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "shepherd.db")
	db := openAttemptTestStore(t, path)
	states := []AttemptWorktreeState{
		AttemptWorktreeCreated, AttemptWorktreePrepared, AttemptWorktreeRunning,
		AttemptWorktreeValidated, AttemptWorktreeRatified, AttemptWorktreePromoting,
		AttemptWorktreePromoted, AttemptWorktreeRetainedForRecovery, AttemptWorktreeCleanupPending,
		AttemptWorktreeCleanupComplete, AttemptWorktreeCleanupBlocked,
	}
	for i, target := range states {
		record := testAttemptWorktreeRecord(int64(i + 1))
		if err := db.CreateAttemptWorktree(ctx, record); err != nil {
			t.Fatalf("create %s: %v", target, err)
		}
		confirmAttemptResourcesForTest(t, ctx, db, record)
		advanceAttemptToState(t, ctx, db, record, target)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttemptTestStore(t, path)
	defer closeStoreForTest(t, db)
	for i, want := range states {
		expected := testAttemptWorktreeRecord(int64(i + 1))
		record, err := db.GetAttemptWorktree(ctx, expected.Key())
		if err != nil || record.State != want || record.CreatedAt.IsZero() || record.UpdatedAt.IsZero() ||
			record.Branch != expected.Branch || record.Path != expected.Path || record.BaseHead != expected.BaseHead ||
			record.ControllerOwner != expected.ControllerOwner || record.ControllerEpoch != expected.ControllerEpoch || !record.ResourcesCreated {
			t.Fatalf("attempt %d=%+v err=%v want=%s", i+1, record, err, want)
		}
	}
}

func TestAttemptWorktreeIdentityCannotRebind(t *testing.T) {
	ctx := context.Background()
	db := openAttemptTestStore(t, filepath.Join(t.TempDir(), "shepherd.db"))
	defer closeStoreForTest(t, db)
	record := testAttemptWorktreeRecord(1)
	if err := db.CreateAttemptWorktree(ctx, record); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateAttemptWorktree(ctx, record); err != nil {
		t.Fatalf("exact duplicate not idempotent: %v", err)
	}
	for name, mutate := range map[string]func(*AttemptWorktreeRecord){
		"branch": func(r *AttemptWorktreeRecord) { r.Branch += "-other" },
		"path":   func(r *AttemptWorktreeRecord) { r.Path += "-other" },
		"base":   func(r *AttemptWorktreeRecord) { r.BaseHead = strings.Repeat("f", 40) },
	} {
		t.Run(name, func(t *testing.T) {
			drifted := record
			mutate(&drifted)
			if err := db.CreateAttemptWorktree(ctx, drifted); err == nil {
				t.Fatalf("%s rebind accepted", name)
			}
		})
	}
}

func TestAttemptWorktreeTransitionsRejectIllegalAndStaleOwner(t *testing.T) {
	ctx := context.Background()
	db := openAttemptTestStore(t, filepath.Join(t.TempDir(), "shepherd.db"))
	defer closeStoreForTest(t, db)
	record := testAttemptWorktreeRecord(1)
	if err := db.CreateAttemptWorktree(ctx, record); err != nil {
		t.Fatal(err)
	}
	confirmAttemptResourcesForTest(t, ctx, db, record)
	if _, err := db.TransitionAttemptWorktree(ctx, record.Key(), "other", record.ControllerEpoch, AttemptWorktreePrepared, AttemptWorktreeUpdate{}); err == nil {
		t.Fatal("stale owner accepted")
	}
	if _, err := db.TransitionAttemptWorktree(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, AttemptWorktreeRunning, AttemptWorktreeUpdate{}); err == nil {
		t.Fatal("created -> running accepted")
	}
	backwardRecord := testAttemptWorktreeRecord(3)
	if err := db.CreateAttemptWorktree(ctx, backwardRecord); err != nil {
		t.Fatal(err)
	}
	confirmAttemptResourcesForTest(t, ctx, db, backwardRecord)
	transitionAttempt(t, ctx, db, backwardRecord, AttemptWorktreePrepared, AttemptWorktreeUpdate{})
	transitionAttempt(t, ctx, db, backwardRecord, AttemptWorktreeRunning, AttemptWorktreeUpdate{})
	if _, err := db.TransitionAttemptWorktree(ctx, backwardRecord.Key(), backwardRecord.ControllerOwner, backwardRecord.ControllerEpoch, AttemptWorktreePrepared, AttemptWorktreeUpdate{}); err == nil {
		t.Fatal("backward running -> prepared accepted")
	}
	staleRecord := testAttemptWorktreeRecord(2)
	if err := db.CreateAttemptWorktree(ctx, staleRecord); err != nil {
		t.Fatal(err)
	}
	confirmAttemptResourcesForTest(t, ctx, db, staleRecord)
	advanceAttemptToState(t, ctx, db, record, AttemptWorktreeCleanupComplete)
	if _, err := db.TransitionAttemptWorktree(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, AttemptWorktreeCleanupPending, AttemptWorktreeUpdate{}); err == nil {
		t.Fatal("terminal transition accepted")
	}
	if err := db.ReleaseLease(ctx, Lease{RunID: record.DeliveryID, Owner: record.ControllerOwner, Epoch: record.ControllerEpoch}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AcquireLease(ctx, record.DeliveryID, "new-owner", time.Now().UTC(), time.Hour); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransitionAttemptWorktree(ctx, staleRecord.Key(), staleRecord.ControllerOwner, staleRecord.ControllerEpoch, AttemptWorktreePrepared, AttemptWorktreeUpdate{}); err == nil {
		t.Fatal("superseded lease transitioned attempt")
	}
}

func TestAttemptWorktreeCleanupClaimRejectsUnconfirmedResources(t *testing.T) {
	ctx := context.Background()
	db := openAttemptTestStore(t, filepath.Join(t.TempDir(), "shepherd.db"))
	defer closeStoreForTest(t, db)
	record := testAttemptWorktreeRecord(1)
	if err := db.CreateAttemptWorktree(ctx, record); err != nil {
		t.Fatal(err)
	}
	transitionAttempt(t, ctx, db, record, AttemptWorktreeRetainedForRecovery, AttemptWorktreeUpdate{FailureClass: "worktree_creation_failure"})
	if err := db.ReleaseLease(ctx, Lease{RunID: record.DeliveryID, Owner: record.ControllerOwner, Epoch: record.ControllerEpoch}); err != nil {
		t.Fatal(err)
	}
	lease, err := db.AcquireLease(ctx, record.DeliveryID, "reconciler", time.Now().UTC(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimAttemptWorktreeCleanup(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, lease.Owner, lease.Epoch); err == nil {
		t.Fatal("unconfirmed Git resources became cleanup-eligible")
	}
}

func TestAttemptWorktreeCleanupClaimBlocksUnknownLiveAndPromotingStates(t *testing.T) {
	ctx := context.Background()
	db := openAttemptTestStore(t, filepath.Join(t.TempDir(), "shepherd.db"))
	defer closeStoreForTest(t, db)
	running := testAttemptWorktreeRecord(1)
	promoting := testAttemptWorktreeRecord(2)
	for _, record := range []AttemptWorktreeRecord{running, promoting} {
		if err := db.CreateAttemptWorktree(ctx, record); err != nil {
			t.Fatal(err)
		}
		confirmAttemptResourcesForTest(t, ctx, db, record)
		transitionAttempt(t, ctx, db, record, AttemptWorktreePrepared, AttemptWorktreeUpdate{})
		transitionAttempt(t, ctx, db, record, AttemptWorktreeRunning, AttemptWorktreeUpdate{})
	}
	candidate := strings.Repeat("c", 40)
	for _, record := range []AttemptWorktreeRecord{running, promoting} {
		if _, err := db.RecordAttemptWorktreeCandidate(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, candidate); err != nil {
			t.Fatal(err)
		}
	}
	transitionAttempt(t, ctx, db, promoting, AttemptWorktreeValidated, AttemptWorktreeUpdate{ValidatedHead: candidate})
	ratifyAttemptForTest(t, ctx, db, promoting)
	transitionAttempt(t, ctx, db, promoting, AttemptWorktreePromoting, AttemptWorktreeUpdate{})
	if _, err := db.TransitionAttemptWorktree(ctx, promoting.Key(), promoting.ControllerOwner, promoting.ControllerEpoch, AttemptWorktreeRetainedForRecovery, AttemptWorktreeUpdate{}); err == nil {
		t.Fatal("ambiguous promoting state became cleanup-eligible")
	}
	if err := db.ReleaseLease(ctx, Lease{RunID: running.DeliveryID, Owner: running.ControllerOwner, Epoch: running.ControllerEpoch}); err != nil {
		t.Fatal(err)
	}
	lease, err := db.AcquireLease(ctx, running.DeliveryID, "reconciler", time.Now().UTC(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range []AttemptWorktreeRecord{running, promoting} {
		if _, err := db.ClaimAttemptWorktreeCleanup(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, lease.Owner, lease.Epoch); err == nil {
			t.Fatalf("unsafe %+v cleanup claim accepted", record.Key())
		}
	}
}

func TestAttemptWorktreePersistsHeadsFailuresAndCleanupRecovery(t *testing.T) {
	ctx := context.Background()
	db := openAttemptTestStore(t, filepath.Join(t.TempDir(), "shepherd.db"))
	defer closeStoreForTest(t, db)
	record := testAttemptWorktreeRecord(1)
	if err := db.CreateAttemptWorktree(ctx, record); err != nil {
		t.Fatal(err)
	}
	confirmAttemptResourcesForTest(t, ctx, db, record)
	transitionAttempt(t, ctx, db, record, AttemptWorktreePrepared, AttemptWorktreeUpdate{})
	transitionAttempt(t, ctx, db, record, AttemptWorktreeRunning, AttemptWorktreeUpdate{})
	candidate := strings.Repeat("c", 40)
	candidateRecord, err := db.RecordAttemptWorktreeCandidate(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, candidate)
	if err != nil || candidateRecord.CandidateHead != candidate || candidateRecord.State != AttemptWorktreeRunning {
		t.Fatalf("candidate record=%+v err=%v", candidateRecord, err)
	}
	transitionAttempt(t, ctx, db, record, AttemptWorktreeValidated, AttemptWorktreeUpdate{ValidatedHead: candidate})
	ratifyAttemptForTest(t, ctx, db, record)
	transitionAttempt(t, ctx, db, record, AttemptWorktreeRetainedForRecovery, AttemptWorktreeUpdate{FailureClass: "runtime_failure"})
	transitionAttempt(t, ctx, db, record, AttemptWorktreeCleanupPending, AttemptWorktreeUpdate{})
	transitionAttempt(t, ctx, db, record, AttemptWorktreeCleanupBlocked, AttemptWorktreeUpdate{CleanupError: strings.Repeat("x", 5000)})
	loaded, err := db.GetAttemptWorktree(ctx, record.Key())
	if err != nil {
		t.Fatal(err)
	}
	if loaded.CandidateHead != candidate || loaded.ValidatedHead != candidate || loaded.FailureClass == "" || len(loaded.CleanupError) > 512 {
		t.Fatalf("loaded=%+v", loaded)
	}
	if err := db.ReleaseLease(ctx, Lease{RunID: record.DeliveryID, Owner: record.ControllerOwner, Epoch: record.ControllerEpoch}); err != nil {
		t.Fatal(err)
	}
	lease, err := db.AcquireLease(ctx, record.DeliveryID, "reconciler", time.Now().UTC(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := db.ClaimAttemptWorktreeCleanup(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, lease.Owner, lease.Epoch)
	if err != nil || claimed.State != AttemptWorktreeCleanupPending || claimed.ControllerOwner != "reconciler" {
		t.Fatalf("claimed=%+v err=%v", claimed, err)
	}
}

func TestRatifiedAttemptEvidenceIsAtomicAndLeaseFenced(t *testing.T) {
	ctx := context.Background()
	db := openAttemptTestStore(t, filepath.Join(t.TempDir(), "shepherd.db"))
	defer closeStoreForTest(t, db)
	record := testAttemptWorktreeRecord(1)
	if err := db.CreateAttemptWorktree(ctx, record); err != nil {
		t.Fatal(err)
	}
	confirmAttemptResourcesForTest(t, ctx, db, record)
	transitionAttempt(t, ctx, db, record, AttemptWorktreePrepared, AttemptWorktreeUpdate{})
	transitionAttempt(t, ctx, db, record, AttemptWorktreeRunning, AttemptWorktreeUpdate{})
	candidate := strings.Repeat("c", 40)
	if _, err := db.RecordAttemptWorktreeCandidate(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, candidate); err != nil {
		t.Fatal(err)
	}
	transitionAttempt(t, ctx, db, record, AttemptWorktreeValidated, AttemptWorktreeUpdate{ValidatedHead: candidate})
	proof := ArtifactProof{ProofID: "atomic-proof", DeliveryID: record.DeliveryID, Generation: record.Generation,
		UnitID: record.UnitID, Attempt: record.Attempt, StartHead: record.BaseHead, CandidateHead: candidate,
		ValidatedHead: candidate, ExpectedArtifact: "artifact", ArtifactHash: "sha256:" + strings.Repeat("d", 64),
		Validator: "openai-codex/gpt-5.6-sol", Thinking: "high", Ratified: true}
	attestation := testAttestationRecord(candidate)
	attestation.RunID, attestation.Generation, attestation.UnitID, attestation.Attempt = record.DeliveryID, record.Generation, record.UnitID, record.Attempt
	attestation.BaseHead = record.BaseHead
	attestation.Verdict = "RETRY"
	if _, err := db.RatifyAttemptWorktree(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, proof, attestation); err == nil {
		t.Fatal("invalid attestation committed")
	}
	if _, err := db.GetArtifactProof(ctx, proof.ProofID); err == nil {
		t.Fatal("proof survived rolled-back ratification")
	}
	loaded, err := db.GetAttemptWorktree(ctx, record.Key())
	if err != nil || loaded.State != AttemptWorktreeValidated {
		t.Fatalf("rolled-back state=%+v err=%v", loaded, err)
	}
	attestation.Verdict = "PROCEED"
	loaded, err = db.RatifyAttemptWorktree(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, proof, attestation)
	if err != nil || loaded.State != AttemptWorktreeRatified {
		t.Fatalf("ratified state=%+v err=%v", loaded, err)
	}
	if _, err := db.GetArtifactProof(ctx, proof.ProofID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetAttestation(ctx, record.DeliveryID, candidate); err != nil {
		t.Fatal(err)
	}
}

func TestReconciliationLeaseRecoversInterruptedDeliveryRun(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "shepherd.db")
	db := openAttemptTestStore(t, path)
	run, err := db.BeginAttempt(ctx, "issue-389", "crashed-owner")
	if err != nil {
		t.Fatal(err)
	}
	unitKey := UnitAttemptKey{DeliveryID: "issue-389", Generation: run.Generation, UnitID: "execute-task/M001/S01/T01", HeadSHA: strings.Repeat("a", 40)}
	if _, err := db.BeginUnitAttempt(ctx, unitKey, 3); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer closeStoreForTest(t, db)
	lease, err := db.AcquireReconciliationLease(ctx, "issue-389", "restart-owner", time.Now().UTC(), time.Minute)
	if err != nil || lease.Epoch <= 1 {
		t.Fatalf("reconciliation lease=%+v err=%v", lease, err)
	}
	if err := db.ReconcileInterruptedDelivery(ctx, lease, domain.RunReady); err != nil {
		t.Fatal(err)
	}
	var status, failure string
	if err := db.db.QueryRowContext(ctx, `SELECT status, last_failure FROM unit_attempts WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ?`, unitKey.DeliveryID, unitKey.Generation, unitKey.UnitID, unitKey.HeadSHA).Scan(&status, &failure); err != nil {
		t.Fatal(err)
	}
	if status != "terminal" || failure != "controller_interrupted" {
		t.Fatalf("interrupted unit status=%q failure=%q", status, failure)
	}
	if _, err := db.BeginAttempt(ctx, "issue-389", "fresh-owner"); err != nil {
		t.Fatalf("reconciled delivery did not restart: %v", err)
	}
}

func TestAttemptWorktreeMigrationPreservesExistingRecords(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "shepherd.db")
	db := openAttemptTestStore(t, path)
	proof := ArtifactProof{ProofID: "proof-1", DeliveryID: "issue-389", Generation: 1, UnitID: "unit", Attempt: 1, StartHead: strings.Repeat("a", 40), CandidateHead: strings.Repeat("b", 40), ValidatedHead: strings.Repeat("b", 40), ExpectedArtifact: "x", ArtifactHash: "sha256:" + strings.Repeat("c", 64), Validator: "openai-codex/gpt-5.6-sol", Thinking: "high", Ratified: true}
	if err := db.PutArtifactProof(ctx, proof); err != nil {
		t.Fatal(err)
	}
	attestation := testAttestationRecord(strings.Repeat("e", 40))
	attestation.RunID = proof.DeliveryID
	if err := db.PutAttestation(ctx, attestation); err != nil {
		t.Fatal(err)
	}
	if _, err := db.db.ExecContext(ctx, `DROP TABLE attempt_worktrees`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttemptTestStore(t, path)
	defer closeStoreForTest(t, db)
	loaded, err := db.GetArtifactProof(ctx, proof.ProofID)
	if err != nil || loaded.ProofID != proof.ProofID {
		t.Fatalf("proof=%+v err=%v", loaded, err)
	}
	if _, err := db.ListAttemptWorktrees(ctx, proof.DeliveryID); err != nil {
		t.Fatalf("new lifecycle table unavailable: %v", err)
	}
	if _, err := db.GetAttestation(ctx, attestation.RunID, attestation.HeadSHA); err != nil {
		t.Fatalf("attestation lost during migration: %v", err)
	}
	var runCount int
	if err := db.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM delivery_runs WHERE delivery_id = ?`, proof.DeliveryID).Scan(&runCount); err != nil || runCount != 1 {
		t.Fatalf("delivery run migration count=%d err=%v", runCount, err)
	}
}

func openAttemptTestStore(t *testing.T, path string) *Store {
	t.Helper()
	db, err := Open(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.EnsureDelivery(context.Background(), testDelivery("issue-389", 389)); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if _, err := db.AcquireLease(context.Background(), "issue-389", "owner-1", time.Now().UTC(), time.Hour); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	return db
}

func testAttemptWorktreeRecord(attempt int64) AttemptWorktreeRecord {
	return AttemptWorktreeRecord{DeliveryID: "issue-389", Generation: 1, UnitID: "execute-task/M001/S01/T01", Attempt: attempt, Branch: "shepherd/issue-389/unit/a" + string(rune('0'+attempt)), Path: filepath.Join("/tmp", "attempts", "a"+string(rune('0'+attempt))), BaseHead: strings.Repeat("a", 40), State: AttemptWorktreeCreated, ControllerOwner: "owner-1", ControllerEpoch: 1}
}

func advanceAttemptToState(t *testing.T, ctx context.Context, db *Store, record AttemptWorktreeRecord, target AttemptWorktreeState) {
	t.Helper()
	if target == AttemptWorktreeCreated {
		return
	}
	path := []AttemptWorktreeState{AttemptWorktreePrepared, AttemptWorktreeRunning, AttemptWorktreeValidated, AttemptWorktreeRatified, AttemptWorktreePromoting, AttemptWorktreePromoted, AttemptWorktreeCleanupPending, AttemptWorktreeCleanupComplete}
	if target == AttemptWorktreeRetainedForRecovery || target == AttemptWorktreeCleanupBlocked {
		path = []AttemptWorktreeState{AttemptWorktreeRetainedForRecovery}
		if target == AttemptWorktreeCleanupBlocked {
			path = append(path, AttemptWorktreeCleanupPending, AttemptWorktreeCleanupBlocked)
		}
	}
	for _, state := range path {
		if state == AttemptWorktreeRatified {
			ratifyAttemptForTest(t, ctx, db, record)
			if state == target {
				return
			}
			continue
		}
		update := AttemptWorktreeUpdate{}
		if state == AttemptWorktreeValidated || state == AttemptWorktreeRatified || state == AttemptWorktreePromoting || state == AttemptWorktreePromoted {
			update.CandidateHead, update.ValidatedHead = strings.Repeat("c", 40), strings.Repeat("c", 40)
		}
		transitionAttempt(t, ctx, db, record, state, update)
		if state == target {
			return
		}
	}
}

func confirmAttemptResourcesForTest(t *testing.T, ctx context.Context, db *Store, record AttemptWorktreeRecord) {
	t.Helper()
	if _, err := db.ConfirmAttemptWorktreeResources(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch); err != nil {
		t.Fatalf("confirm attempt resources: %v", err)
	}
}

func ratifyAttemptForTest(t *testing.T, ctx context.Context, db *Store, record AttemptWorktreeRecord) {
	t.Helper()
	current, err := db.GetAttemptWorktree(ctx, record.Key())
	if err != nil {
		t.Fatal(err)
	}
	proof := ArtifactProof{ProofID: fmt.Sprintf("proof-%d", record.Attempt), DeliveryID: record.DeliveryID,
		Generation: record.Generation, UnitID: record.UnitID, Attempt: record.Attempt, StartHead: record.BaseHead,
		CandidateHead: current.CandidateHead, ValidatedHead: current.ValidatedHead, ExpectedArtifact: "artifact",
		ArtifactHash: "sha256:" + strings.Repeat("d", 64), Validator: "openai-codex/gpt-5.6-sol", Thinking: "high", Ratified: true}
	attestation := testAttestationRecord(current.CandidateHead)
	attestation.RunID, attestation.Generation, attestation.UnitID, attestation.Attempt = record.DeliveryID, record.Generation, record.UnitID, record.Attempt
	attestation.BaseHead = record.BaseHead
	if _, err := db.RatifyAttemptWorktree(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, proof, attestation); err != nil {
		t.Fatalf("ratify attempt: %v", err)
	}
}

func transitionAttempt(t *testing.T, ctx context.Context, db *Store, record AttemptWorktreeRecord, next AttemptWorktreeState, update AttemptWorktreeUpdate) {
	t.Helper()
	if _, err := db.TransitionAttemptWorktree(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, next, update); err != nil {
		t.Fatalf("transition to %s: %v", next, err)
	}
}
