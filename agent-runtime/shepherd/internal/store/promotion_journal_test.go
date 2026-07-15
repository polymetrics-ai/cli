package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPromotionJournalPersistsAcrossReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "shepherd.db")
	db, lease, journal := promotionJournalFixture(t, path)
	createFinalizedPromotionJournal(t, ctx, db, lease, journal)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer closeStoreForTest(t, db)
	loaded, err := db.GetPromotionJournal(ctx, journal.JournalID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.State != PromotionJournalStateStaged || loaded.ManifestHash != journal.ManifestHash ||
		loaded.BackupManifestHash != journal.BackupManifestHash || loaded.ProofID != journal.ProofID ||
		loaded.ValidatorSessionID != journal.ValidatorSessionID || loaded.GovernanceStateVersion != journal.GovernanceStateVersion {
		t.Fatalf("journal after reopen=%+v", loaded)
	}
}

func TestPromotionJournalIntentPersistsBeforeStageAndFinalizesOnce(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "shepherd.db")
	db, lease, journal := promotionJournalFixture(t, path)
	manifestJSON, manifestHash := journal.ManifestJSON, journal.ManifestHash
	backupJSON, backupHash := journal.BackupManifestJSON, journal.BackupManifestHash
	journal.ManifestJSON, journal.ManifestHash = "", ""
	journal.BackupManifestJSON, journal.BackupManifestHash = "", ""
	if err := db.CreatePromotionJournal(ctx, lease, journal); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer closeStoreForTest(t, db)
	loaded, err := db.GetPromotionJournal(ctx, journal.JournalID)
	if err != nil || loaded.State != PromotionJournalCreated || loaded.ManifestHash != "" {
		t.Fatalf("staging intent=%+v err=%v", loaded, err)
	}
	loaded, err = db.FinalizePromotionJournalStage(ctx, lease, journal.JournalID, manifestJSON, manifestHash, backupJSON, backupHash)
	if err != nil || loaded.State != PromotionJournalStateStaged || loaded.ManifestHash != manifestHash {
		t.Fatalf("finalized stage=%+v err=%v", loaded, err)
	}
	if _, err := db.FinalizePromotionJournalStage(ctx, lease, journal.JournalID, manifestJSON, "sha256:"+strings.Repeat("f", 64), backupJSON, backupHash); err == nil {
		t.Fatal("finalized manifest was rebound")
	}
}

func TestPromotionJournalTransitionGraphAndIdempotency(t *testing.T) {
	ctx := context.Background()
	db, lease, journal := promotionJournalFixture(t, filepath.Join(t.TempDir(), "shepherd.db"))
	defer closeStoreForTest(t, db)
	createFinalizedPromotionJournal(t, ctx, db, lease, journal)
	if _, err := db.TransitionPromotionJournal(ctx, lease, journal.JournalID, PromotionJournalGitPromoted, ""); err == nil {
		t.Fatal("journal skipped directly to git_promoted")
	}
	for _, state := range []PromotionJournalState{PromotionJournalGitPromoting, PromotionJournalGitPromoted, PromotionJournalStateSwapStarted, PromotionJournalStateInstalled, PromotionJournalComplete} {
		if _, err := db.TransitionPromotionJournal(ctx, lease, journal.JournalID, state, ""); err != nil {
			t.Fatalf("transition %s: %v", state, err)
		}
	}
	if _, err := db.CompletePromotionAttempt(ctx, lease, journal.JournalID); err != nil {
		t.Fatal(err)
	}
	if err := db.MarkPromotionCleanupComplete(ctx, lease, journal.JournalID); err != nil {
		t.Fatal(err)
	}
	loaded, err := db.GetPromotionJournal(ctx, journal.JournalID)
	if err != nil || !loaded.CleanupComplete {
		t.Fatalf("cleanup completion journal=%+v err=%v", loaded, err)
	}
	if _, err := db.TransitionPromotionJournal(ctx, lease, journal.JournalID, PromotionJournalStateInstalled, ""); err == nil {
		t.Fatal("complete journal moved backward")
	}
}

func TestPromotionJournalRejectsExpiredOrMismatchedAuthority(t *testing.T) {
	ctx := context.Background()
	for _, test := range []struct {
		name   string
		mutate func(*PromotionJournal)
	}{
		{name: "wrong proof", mutate: func(j *PromotionJournal) { j.ProofID = "wrong-proof" }},
		{name: "wrong candidate", mutate: func(j *PromotionJournal) { j.CandidateHead = strings.Repeat("f", 40) }},
		{name: "wrong evidence", mutate: func(j *PromotionJournal) { j.EvidenceHash = "sha256:" + strings.Repeat("f", 64) }},
		{name: "wrong governance version", mutate: func(j *PromotionJournal) { j.GovernanceStateVersion++ }},
		{name: "mismatched backup ownership", mutate: func(j *PromotionJournal) {
			j.BackupPath = filepath.Join(filepath.Dir(j.BackupPath), ".other.shepherd-gsd-j1.backup")
		}},
		{name: "expired", mutate: func(j *PromotionJournal) { j.AttestationExpiresAt = time.Now().UTC().Add(-time.Minute) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, lease, journal := promotionJournalFixture(t, filepath.Join(t.TempDir(), "shepherd.db"))
			defer closeStoreForTest(t, db)
			test.mutate(&journal)
			journal.ManifestJSON, journal.ManifestHash = "", ""
			journal.BackupManifestJSON, journal.BackupManifestHash = "", ""
			if err := db.CreatePromotionJournal(ctx, lease, journal); err == nil {
				t.Fatal("invalid promotion authority accepted")
			}
		})
	}
}

func TestPromotionJournalRejectsCrossRecordProofAndAttestationMismatch(t *testing.T) {
	ctx := context.Background()
	db, lease, journal := promotionJournalFixture(t, filepath.Join(t.TempDir(), "shepherd.db"))
	defer closeStoreForTest(t, db)
	if _, err := db.db.ExecContext(ctx, `UPDATE artifact_proofs SET artifact_hash = ? WHERE proof_id = ?`,
		"sha256:"+strings.Repeat("f", 64), journal.ProofID); err != nil {
		t.Fatal(err)
	}
	journal.ManifestJSON, journal.ManifestHash = "", ""
	journal.BackupManifestJSON, journal.BackupManifestHash = "", ""
	if err := db.CreatePromotionJournal(ctx, lease, journal); err == nil {
		t.Fatal("cross-record proof/attestation mismatch accepted")
	}
}

func TestPromotionJournalRecoveryClaimAndBlockedState(t *testing.T) {
	ctx := context.Background()
	db, lease, journal := promotionJournalFixture(t, filepath.Join(t.TempDir(), "shepherd.db"))
	defer closeStoreForTest(t, db)
	journal.ManifestJSON, journal.ManifestHash = "", ""
	journal.BackupManifestJSON, journal.BackupManifestHash = "", ""
	if err := db.CreatePromotionJournal(ctx, lease, journal); err != nil {
		t.Fatal(err)
	}
	if err := db.ReleaseLease(ctx, lease); err != nil {
		t.Fatal(err)
	}
	recoveryLease, err := db.AcquireReconciliationLease(ctx, lease.RunID, "recovery", time.Now().UTC(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := db.ClaimPromotionJournalRecovery(ctx, journal.JournalID, recoveryLease)
	if err != nil || claimed.ControllerOwner != recoveryLease.Owner || claimed.ControllerEpoch != recoveryLease.Epoch {
		t.Fatalf("claimed=%+v err=%v", claimed, err)
	}
	blocked, err := db.TransitionPromotionJournal(ctx, recoveryLease, journal.JournalID, PromotionJournalBlocked, "stage_manifest_mismatch")
	if err != nil || blocked.State != PromotionJournalBlocked || blocked.BlockedReason == "" || len(blocked.BlockedReason) > 512 {
		t.Fatalf("blocked=%+v err=%v", blocked, err)
	}
	hasBlocked, err := db.HasBlockedPromotionJournal(ctx, journal.DeliveryID)
	if err != nil || !hasBlocked {
		t.Fatalf("blocked promotion gate=%t err=%v", hasBlocked, err)
	}
	hasJournal, err := db.AttemptHasPromotionJournal(ctx, journal.AttemptKey())
	if err != nil || !hasJournal {
		t.Fatalf("attempt journal ownership=%t err=%v", hasJournal, err)
	}
}

func createFinalizedPromotionJournal(t *testing.T, ctx context.Context, db *Store, lease Lease, journal PromotionJournal) {
	t.Helper()
	manifestJSON, manifestHash := journal.ManifestJSON, journal.ManifestHash
	backupJSON, backupHash := journal.BackupManifestJSON, journal.BackupManifestHash
	journal.ManifestJSON, journal.ManifestHash = "", ""
	journal.BackupManifestJSON, journal.BackupManifestHash = "", ""
	if err := db.CreatePromotionJournal(ctx, lease, journal); err != nil {
		t.Fatal(err)
	}
	if _, err := db.FinalizePromotionJournalStage(ctx, lease, journal.JournalID, manifestJSON, manifestHash, backupJSON, backupHash); err != nil {
		t.Fatal(err)
	}
}

func promotionJournalFixture(t *testing.T, path string) (*Store, Lease, PromotionJournal) {
	t.Helper()
	ctx := context.Background()
	db := openAttemptTestStore(t, path)
	lease := Lease{RunID: "issue-389", Owner: "owner-1", Epoch: 1}
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
	now := time.Now().UTC()
	attestation := testAttestationRecord(candidate)
	attestation.RunID, attestation.Generation, attestation.UnitID, attestation.Attempt = record.DeliveryID, record.Generation, record.UnitID, record.Attempt
	attestation.BaseHead = record.BaseHead
	attestation.StateVersion = 1
	attestation.CreatedAt, attestation.ExpiresAt = now.Add(-time.Minute), now.Add(time.Hour)
	proof := ArtifactProof{ProofID: "promotion-proof", DeliveryID: record.DeliveryID, Generation: record.Generation,
		UnitID: record.UnitID, Attempt: record.Attempt, StartHead: record.BaseHead, CandidateHead: candidate,
		ValidatedHead: candidate, ExpectedArtifact: "artifact", ArtifactHash: attestation.EvidenceHash,
		Validator: attestation.Validator, Thinking: attestation.Thinking, Ratified: true}
	if _, err := db.RatifyAttemptWorktree(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, proof, attestation); err != nil {
		t.Fatal(err)
	}
	manifestJSON := `[{"path":"STATE.md","type":"file","size":1,"hash":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}]`
	backupManifestJSON := `[{"path":"STATE.md","type":"file","size":1,"hash":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}]`
	manifestSum, backupManifestSum := sha256.Sum256([]byte(manifestJSON)), sha256.Sum256([]byte(backupManifestJSON))
	promotionRoot := t.TempDir()
	promotionPrefix := filepath.Join(promotionRoot, ".repo.shepherd-gsd-j1")
	return db, lease, PromotionJournal{
		JournalID: "promotion-journal-1", DeliveryID: record.DeliveryID, Generation: record.Generation,
		UnitID: record.UnitID, Attempt: record.Attempt, BaseHead: record.BaseHead, CandidateHead: candidate,
		ValidatedHead: candidate, ProofID: proof.ProofID, EvidenceHash: attestation.EvidenceHash,
		ValidatorSessionID: attestation.ValidatorSessionID, AttestationRepository: attestation.Repository,
		AttestationPR: attestation.PR, AttestationBaseBranch: attestation.BaseBranch,
		AttestationContractHash: attestation.ContractHash, AttestationCreatedAt: attestation.CreatedAt,
		AttestationValidator: attestation.Validator,
		AttestationThinking:  attestation.Thinking, AttestationVerdict: attestation.Verdict,
		AttestationLocalGates: attestation.LocalGates, AttestationUAT: attestation.UAT,
		AttestationMilestoneValid: attestation.MilestoneValid, GovernanceStateVersion: attestation.StateVersion,
		AttestationExpiresAt: attestation.ExpiresAt, ManifestJSON: manifestJSON,
		ManifestHash: "sha256:" + hex.EncodeToString(manifestSum[:]), BackupManifestJSON: backupManifestJSON,
		BackupManifestHash: "sha256:" + hex.EncodeToString(backupManifestSum[:]), StagePath: promotionPrefix + ".stage",
		BackupPath: promotionPrefix + ".backup", CanonicalPath: filepath.Join(promotionRoot, "repo", ".gsd"),
		State: PromotionJournalCreated, ControllerOwner: lease.Owner, ControllerEpoch: lease.Epoch,
	}
}
