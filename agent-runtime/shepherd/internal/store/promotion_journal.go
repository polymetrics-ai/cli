package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type PromotionJournalState string

const (
	PromotionJournalCreated          PromotionJournalState = "journal_created"
	PromotionJournalStateStaged      PromotionJournalState = "state_staged"
	PromotionJournalGitPromoting     PromotionJournalState = "git_promoting"
	PromotionJournalGitPromoted      PromotionJournalState = "git_promoted"
	PromotionJournalStateSwapStarted PromotionJournalState = "state_swap_started"
	PromotionJournalStateInstalled   PromotionJournalState = "state_installed"
	PromotionJournalComplete         PromotionJournalState = "complete"
	PromotionJournalBlocked          PromotionJournalState = "blocked"
)

type PromotionJournal struct {
	JournalID                 string
	DeliveryID                string
	Generation                int64
	UnitID                    string
	Attempt                   int64
	BaseHead                  string
	CandidateHead             string
	ValidatedHead             string
	ProofID                   string
	EvidenceHash              string
	ValidatorSessionID        string
	AttestationRepository     string
	AttestationPR             int
	AttestationBaseBranch     string
	AttestationContractHash   string
	AttestationCreatedAt      time.Time
	AttestationValidator      string
	AttestationThinking       string
	AttestationVerdict        string
	AttestationLocalGates     bool
	AttestationUAT            bool
	AttestationMilestoneValid bool
	GovernanceStateVersion    int64
	AttestationExpiresAt      time.Time
	ManifestJSON              string
	ManifestHash              string
	BackupManifestJSON        string
	BackupManifestHash        string
	StagePath                 string
	BackupPath                string
	CanonicalPath             string
	State                     PromotionJournalState
	BlockedReason             string
	CleanupComplete           bool
	DecisionsResolved         bool
	ControllerOwner           string
	ControllerEpoch           int64
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

func (j PromotionJournal) AttemptKey() AttemptWorktreeKey {
	return AttemptWorktreeKey{DeliveryID: j.DeliveryID, Generation: j.Generation, UnitID: j.UnitID, Attempt: j.Attempt}
}

func (s *Store) CreatePromotionJournal(ctx context.Context, lease Lease, journal PromotionJournal) error {
	if journal.ManifestJSON != "" || journal.ManifestHash != "" || journal.BackupManifestJSON != "" || journal.BackupManifestHash != "" {
		return errors.New("promotion journal creation accepts staging intent only")
	}
	if err := validatePromotionJournal(journal); err != nil {
		return err
	}
	if lease.RunID != journal.DeliveryID || lease.Owner != journal.ControllerOwner || lease.Epoch != journal.ControllerEpoch {
		return errors.New("promotion journal controller does not match lease")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), "ROLLBACK")
		}
	}()
	now := time.Now().UTC()
	if err := checkAttemptLeaseConn(ctx, conn, journal.DeliveryID, journal.ControllerOwner, journal.ControllerEpoch, now); err != nil {
		return err
	}
	if existing, err := getPromotionJournalConn(ctx, conn, journal.JournalID); err == nil {
		if !samePromotionIdentity(existing, journal) {
			return errors.New("promotion journal identity cannot be rebound")
		}
		if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
			return err
		}
		committed = true
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err := verifyPromotionAuthorityConn(ctx, conn, journal, now, true); err != nil {
		return err
	}
	timestamp := now.UnixNano()
	_, err = conn.ExecContext(ctx, `INSERT INTO promotion_journals
		(journal_id, delivery_id, generation, unit_id, attempt, base_head, candidate_head,
		 validated_head, proof_id, evidence_hash, validator_session_id, attestation_repository,
		 attestation_pr, attestation_base_branch, attestation_contract_hash, attestation_created_at, attestation_validator,
		 attestation_thinking, attestation_verdict, attestation_local_gates, attestation_uat,
		 attestation_milestone_valid, governance_state_version, attestation_expires_at, manifest_json, manifest_hash, backup_manifest_json,
		 backup_manifest_hash, stage_path, backup_path, canonical_path, state, blocked_reason,
		 controller_owner, controller_epoch, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', ?, ?, ?, ?)`,
		journal.JournalID, journal.DeliveryID, journal.Generation, journal.UnitID, journal.Attempt,
		journal.BaseHead, journal.CandidateHead, journal.ValidatedHead, journal.ProofID,
		journal.EvidenceHash, journal.ValidatorSessionID, journal.AttestationRepository, journal.AttestationPR,
		journal.AttestationBaseBranch, journal.AttestationContractHash, journal.AttestationCreatedAt.UnixNano(), journal.AttestationValidator,
		journal.AttestationThinking, journal.AttestationVerdict, promotionBool(journal.AttestationLocalGates),
		promotionBool(journal.AttestationUAT), promotionBool(journal.AttestationMilestoneValid), journal.GovernanceStateVersion,
		journal.AttestationExpiresAt.UnixNano(), journal.ManifestJSON, journal.ManifestHash,
		journal.BackupManifestJSON, journal.BackupManifestHash, journal.StagePath,
		journal.BackupPath, journal.CanonicalPath, PromotionJournalCreated,
		journal.ControllerOwner, journal.ControllerEpoch, timestamp, timestamp)
	if err != nil {
		return fmt.Errorf("create promotion journal: %w", err)
	}
	result, err := conn.ExecContext(ctx, `UPDATE attempt_worktrees SET state = ?, updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND attempt = ? AND state = ?
		AND controller_owner = ? AND controller_epoch = ?`, AttemptWorktreePromoting, timestamp,
		journal.DeliveryID, journal.Generation, journal.UnitID, journal.Attempt,
		AttemptWorktreeRatified, journal.ControllerOwner, journal.ControllerEpoch)
	if err != nil {
		return fmt.Errorf("bind promoting attempt to journal: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return errors.New("ratified attempt changed before journal creation")
	}
	if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Store) HasBlockedPromotionJournal(ctx context.Context, deliveryID string) (bool, error) {
	if strings.TrimSpace(deliveryID) == "" || strings.ContainsAny(deliveryID, "\r\n\x00") {
		return false, errors.New("delivery ID is required")
	}
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM promotion_journals WHERE delivery_id = ? AND state = ?`,
		deliveryID, PromotionJournalBlocked).Scan(&count); err != nil {
		return false, fmt.Errorf("check blocked promotion journals: %w", err)
	}
	return count > 0, nil
}

func (s *Store) AttemptHasPromotionJournal(ctx context.Context, key AttemptWorktreeKey) (bool, error) {
	if err := validateAttemptWorktreeKey(key); err != nil {
		return false, err
	}
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM promotion_journals WHERE delivery_id = ?
		AND generation = ? AND unit_id = ? AND attempt = ?`, key.DeliveryID, key.Generation,
		key.UnitID, key.Attempt).Scan(&count); err != nil {
		return false, fmt.Errorf("check attempt promotion journal: %w", err)
	}
	return count > 0, nil
}

func (s *Store) GetFinalGatePromotionJournal(ctx context.Context, deliveryID string, generation int64,
	headSHA string,
) (PromotionJournal, error) {
	if strings.TrimSpace(deliveryID) == "" || generation <= 0 || !validGitSHA(headSHA) {
		return PromotionJournal{}, errors.New("final-gate promotion identity is required")
	}
	rows, err := s.db.QueryContext(ctx, promotionJournalSelect+` WHERE delivery_id = ?
		AND candidate_head = ? AND validated_head = ? AND state = ? AND cleanup_complete = 1
		AND (generation = ? OR (generation + 1 = ? AND EXISTS (
			SELECT 1 FROM human_decisions h WHERE h.delivery_id = promotion_journals.delivery_id
			AND h.generation = promotion_journals.generation AND h.approved = 1))) LIMIT 2`,
		deliveryID, headSHA, headSHA, PromotionJournalComplete, generation, generation)
	if err != nil {
		return PromotionJournal{}, err
	}
	defer func() { _ = rows.Close() }()
	var journal PromotionJournal
	count := 0
	for rows.Next() {
		journal, err = scanPromotionJournal(rows)
		if err != nil {
			return PromotionJournal{}, err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return PromotionJournal{}, err
	}
	if count != 1 {
		return PromotionJournal{}, errors.New("exactly one resolved promotion must bind the final gate")
	}
	return journal, nil
}

func (s *Store) GetPromotionJournal(ctx context.Context, journalID string) (PromotionJournal, error) {
	if err := validatePromotionJournalID(journalID); err != nil {
		return PromotionJournal{}, err
	}
	return scanPromotionJournal(s.db.QueryRowContext(ctx, promotionJournalSelect+" WHERE journal_id = ?", journalID))
}

func (s *Store) ListIncompletePromotionJournals(ctx context.Context, deliveryID string) ([]PromotionJournal, error) {
	if strings.TrimSpace(deliveryID) == "" || strings.ContainsAny(deliveryID, "\r\n\x00") {
		return nil, errors.New("delivery ID is required")
	}
	rows, err := s.db.QueryContext(ctx, promotionJournalSelect+` WHERE delivery_id = ?
		AND state <> ? AND (state <> ? OR cleanup_complete = 0 OR decisions_resolved = 0)
		ORDER BY created_at, journal_id LIMIT 65`, deliveryID, PromotionJournalBlocked,
		PromotionJournalComplete)
	if err != nil {
		return nil, fmt.Errorf("list promotion journals: %w", err)
	}
	defer func() { _ = rows.Close() }()
	journals := make([]PromotionJournal, 0, 4)
	for rows.Next() {
		journal, err := scanPromotionJournal(rows)
		if err != nil {
			return nil, err
		}
		journals = append(journals, journal)
		if len(journals) > 64 {
			return nil, errors.New("promotion journal recovery set exceeds governed limit")
		}
	}
	return journals, rows.Err()
}

func (s *Store) FinalizePromotionJournalStage(ctx context.Context, lease Lease, journalID, manifestJSON, manifestHash, backupManifestJSON, backupManifestHash string) (PromotionJournal, error) {
	candidate := PromotionJournal{ManifestJSON: manifestJSON, ManifestHash: manifestHash,
		BackupManifestJSON: backupManifestJSON, BackupManifestHash: backupManifestHash}
	if !validSHA256Digest(candidate.ManifestHash) || !validSHA256Digest(candidate.BackupManifestHash) {
		return PromotionJournal{}, errors.New("promotion manifest hashes must be sha256 digests")
	}
	if err := validateStoredManifest(candidate.ManifestJSON); err != nil || digestManifestJSON(candidate.ManifestJSON) != candidate.ManifestHash {
		return PromotionJournal{}, errors.New("staged manifest JSON and hash do not match")
	}
	if err := validateStoredManifest(candidate.BackupManifestJSON); err != nil || digestManifestJSON(candidate.BackupManifestJSON) != candidate.BackupManifestHash {
		return PromotionJournal{}, errors.New("backup manifest JSON and hash do not match")
	}
	current, err := s.GetPromotionJournal(ctx, journalID)
	if err != nil {
		return PromotionJournal{}, err
	}
	if current.DeliveryID != lease.RunID || current.ControllerOwner != lease.Owner || current.ControllerEpoch != lease.Epoch {
		return PromotionJournal{}, errors.New("promotion journal controller is stale or fenced")
	}
	if current.State == PromotionJournalStateStaged && current.ManifestJSON == manifestJSON &&
		current.ManifestHash == manifestHash && current.BackupManifestJSON == backupManifestJSON &&
		current.BackupManifestHash == backupManifestHash {
		return current, nil
	}
	if current.State != PromotionJournalCreated || current.ManifestJSON != "" || current.ManifestHash != "" ||
		current.BackupManifestJSON != "" || current.BackupManifestHash != "" {
		return PromotionJournal{}, errors.New("promotion stage identity cannot be rebound")
	}
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx, `UPDATE promotion_journals SET manifest_json = ?, manifest_hash = ?,
		backup_manifest_json = ?, backup_manifest_hash = ?, state = ?, updated_at = ?
		WHERE journal_id = ? AND state = ? AND manifest_json = '' AND manifest_hash = ''
		AND backup_manifest_json = '' AND backup_manifest_hash = '' AND controller_owner = ?
		AND controller_epoch = ? AND EXISTS (SELECT 1 FROM leases WHERE run_id = ? AND owner = ?
		AND epoch = ? AND expires_at > ?)`, manifestJSON, manifestHash, backupManifestJSON,
		backupManifestHash, PromotionJournalStateStaged, now.UnixNano(), journalID,
		PromotionJournalCreated, lease.Owner, lease.Epoch, lease.RunID, lease.Owner, lease.Epoch, now.UnixNano())
	if err != nil {
		return PromotionJournal{}, fmt.Errorf("finalize promotion stage: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return PromotionJournal{}, errors.New("promotion stage finalization is stale or fenced")
	}
	return s.GetPromotionJournal(ctx, journalID)
}

func (s *Store) TransitionPromotionJournal(ctx context.Context, lease Lease, journalID string, next PromotionJournalState, blockedReason string) (PromotionJournal, error) {
	if err := validatePromotionJournalID(journalID); err != nil {
		return PromotionJournal{}, err
	}
	if next == PromotionJournalBlocked {
		blockedReason = boundedAttemptDiagnostic(blockedReason)
		if blockedReason == "" {
			return PromotionJournal{}, errors.New("blocked promotion journal requires a reason")
		}
	} else if blockedReason != "" {
		return PromotionJournal{}, errors.New("blocked reason is only valid for blocked journals")
	}
	current, err := s.GetPromotionJournal(ctx, journalID)
	if err != nil {
		return PromotionJournal{}, err
	}
	if current.ControllerOwner != lease.Owner || current.ControllerEpoch != lease.Epoch || current.DeliveryID != lease.RunID {
		return PromotionJournal{}, errors.New("promotion journal controller is stale or fenced")
	}
	if current.State == next && current.BlockedReason == blockedReason {
		return current, nil
	}
	if !legalPromotionTransition(current.State, next) {
		return PromotionJournal{}, fmt.Errorf("illegal promotion journal transition %s -> %s", current.State, next)
	}
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx, `UPDATE promotion_journals SET state = ?, blocked_reason = ?, updated_at = ?
		WHERE journal_id = ? AND state = ? AND controller_owner = ? AND controller_epoch = ?
		AND EXISTS (SELECT 1 FROM leases WHERE run_id = ? AND owner = ? AND epoch = ? AND expires_at > ?)`,
		next, blockedReason, now.UnixNano(), journalID, current.State, lease.Owner, lease.Epoch,
		lease.RunID, lease.Owner, lease.Epoch, now.UnixNano())
	if err != nil {
		return PromotionJournal{}, fmt.Errorf("transition promotion journal: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return PromotionJournal{}, errors.New("promotion journal transition is stale or fenced")
	}
	return s.GetPromotionJournal(ctx, journalID)
}

func (s *Store) ClaimPromotionJournalRecovery(ctx context.Context, journalID string, lease Lease) (PromotionJournal, error) {
	if err := validatePromotionJournalID(journalID); err != nil {
		return PromotionJournal{}, err
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return PromotionJournal{}, err
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return PromotionJournal{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), "ROLLBACK")
		}
	}()
	now := time.Now().UTC()
	if err := checkAttemptLeaseConn(ctx, conn, lease.RunID, lease.Owner, lease.Epoch, now); err != nil {
		return PromotionJournal{}, err
	}
	journal, err := getPromotionJournalConn(ctx, conn, journalID)
	if err != nil {
		return PromotionJournal{}, err
	}
	if journal.DeliveryID != lease.RunID || journal.State == PromotionJournalBlocked {
		return PromotionJournal{}, errors.New("promotion journal is not recoverable by this delivery")
	}
	result, err := conn.ExecContext(ctx, `UPDATE promotion_journals SET controller_owner = ?, controller_epoch = ?, updated_at = ?
		WHERE journal_id = ? AND state <> ?`, lease.Owner, lease.Epoch, now.UnixNano(), journalID,
		PromotionJournalBlocked)
	if err != nil {
		return PromotionJournal{}, err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return PromotionJournal{}, errors.New("promotion journal recovery claim failed")
	}
	result, err = conn.ExecContext(ctx, `UPDATE attempt_worktrees SET controller_owner = ?, controller_epoch = ?, updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND attempt = ? AND state IN (?, ?)`,
		lease.Owner, lease.Epoch, now.UnixNano(), journal.DeliveryID, journal.Generation,
		journal.UnitID, journal.Attempt, AttemptWorktreePromoting, AttemptWorktreePromoted)
	if err != nil {
		return PromotionJournal{}, err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return PromotionJournal{}, errors.New("promoting attempt recovery claim failed")
	}
	if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
		return PromotionJournal{}, err
	}
	committed = true
	return getPromotionJournalConn(ctx, conn, journalID)
}

func (s *Store) MarkPromotionCleanupComplete(ctx context.Context, lease Lease, journalID string) error {
	journal, err := s.GetPromotionJournal(ctx, journalID)
	if err != nil {
		return err
	}
	if journal.State != PromotionJournalComplete || journal.ControllerOwner != lease.Owner ||
		journal.ControllerEpoch != lease.Epoch || journal.DeliveryID != lease.RunID {
		return errors.New("complete controller-owned promotion journal is required for cleanup")
	}
	if journal.CleanupComplete {
		return nil
	}
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx, `UPDATE promotion_journals SET cleanup_complete = 1, updated_at = ?
		WHERE journal_id = ? AND state = ? AND cleanup_complete = 0 AND controller_owner = ?
		AND controller_epoch = ? AND EXISTS (SELECT 1 FROM leases WHERE run_id = ? AND owner = ?
		AND epoch = ? AND expires_at > ?)`, now.UnixNano(), journalID, PromotionJournalComplete,
		lease.Owner, lease.Epoch, lease.RunID, lease.Owner, lease.Epoch, now.UnixNano())
	if err != nil {
		return fmt.Errorf("mark promotion cleanup complete: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return errors.New("promotion cleanup completion is stale or fenced")
	}
	return nil
}

func (s *Store) GetRatifiedPromotionAuthority(ctx context.Context, key AttemptWorktreeKey) (ArtifactProof, AttestationRecord, error) {
	if err := validateAttemptWorktreeKey(key); err != nil {
		return ArtifactProof{}, AttestationRecord{}, err
	}
	attempt, err := s.GetAttemptWorktree(ctx, key)
	if err != nil {
		return ArtifactProof{}, AttestationRecord{}, err
	}
	if attempt.State != AttemptWorktreeRatified || attempt.CandidateHead == "" || attempt.CandidateHead != attempt.ValidatedHead {
		return ArtifactProof{}, AttestationRecord{}, errors.New("ratified attempt with one exact candidate is required")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT proof_id, delivery_id, generation, unit_id, attempt,
		start_head, candidate_head, validated_head, expected_artifact, artifact_hash, validator,
		thinking, ratified FROM artifact_proofs WHERE delivery_id = ? AND generation = ? AND unit_id = ?
		AND attempt = ? AND candidate_head = ? AND validated_head = ? AND ratified = 1 LIMIT 2`,
		key.DeliveryID, key.Generation, key.UnitID, key.Attempt, attempt.CandidateHead, attempt.ValidatedHead)
	if err != nil {
		return ArtifactProof{}, AttestationRecord{}, err
	}
	defer func() { _ = rows.Close() }()
	var proof ArtifactProof
	var count int
	for rows.Next() {
		var ratified int
		if err := rows.Scan(&proof.ProofID, &proof.DeliveryID, &proof.Generation, &proof.UnitID,
			&proof.Attempt, &proof.StartHead, &proof.CandidateHead, &proof.ValidatedHead,
			&proof.ExpectedArtifact, &proof.ArtifactHash, &proof.Validator, &proof.Thinking, &ratified); err != nil {
			return ArtifactProof{}, AttestationRecord{}, err
		}
		proof.Ratified = ratified == 1
		count++
	}
	if err := rows.Err(); err != nil {
		return ArtifactProof{}, AttestationRecord{}, err
	}
	if count != 1 || proof.StartHead != attempt.BaseHead {
		return ArtifactProof{}, AttestationRecord{}, errors.New("exactly one ratified proof must bind the attempt")
	}
	attestation, err := s.GetAttestation(ctx, key.DeliveryID, attempt.CandidateHead)
	if err != nil {
		return ArtifactProof{}, AttestationRecord{}, err
	}
	if attestation.Generation != key.Generation || attestation.UnitID != key.UnitID ||
		attestation.Attempt != key.Attempt || attestation.BaseHead != attempt.BaseHead ||
		attestation.CandidateHead != attempt.CandidateHead || attestation.ObservedHead != attempt.CandidateHead ||
		proof.ArtifactHash != attestation.EvidenceHash || proof.Validator != attestation.Validator || proof.Thinking != attestation.Thinking {
		return ArtifactProof{}, AttestationRecord{}, errors.New("attestation does not bind the ratified attempt")
	}
	return proof, attestation, nil
}

func (s *Store) ValidatePromotionAuthority(ctx context.Context, lease Lease, journalID string, now time.Time) error {
	journal, err := s.GetPromotionJournal(ctx, journalID)
	if err != nil {
		return err
	}
	if journal.ControllerOwner != lease.Owner || journal.ControllerEpoch != lease.Epoch || journal.DeliveryID != lease.RunID {
		return errors.New("promotion journal controller is stale or fenced")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	if err := checkAttemptLeaseConn(ctx, conn, lease.RunID, lease.Owner, lease.Epoch, now.UTC()); err != nil {
		return err
	}
	return verifyPromotionAuthorityConn(ctx, conn, journal, now.UTC(), false)
}

func (s *Store) CompletePromotionAttempt(ctx context.Context, lease Lease, journalID string) (AttemptWorktreeRecord, error) {
	journal, err := s.GetPromotionJournal(ctx, journalID)
	if err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if journal.State != PromotionJournalComplete || journal.ControllerOwner != lease.Owner || journal.ControllerEpoch != lease.Epoch || journal.DeliveryID != lease.RunID {
		return AttemptWorktreeRecord{}, errors.New("complete, controller-owned promotion journal is required")
	}
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx, `UPDATE attempt_worktrees SET state = ?, updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND attempt = ? AND state = ?
		AND controller_owner = ? AND controller_epoch = ?
		AND EXISTS (SELECT 1 FROM leases WHERE run_id = ? AND owner = ? AND epoch = ? AND expires_at > ?)`,
		AttemptWorktreePromoted, now.UnixNano(), journal.DeliveryID, journal.Generation, journal.UnitID,
		journal.Attempt, AttemptWorktreePromoting, lease.Owner, lease.Epoch, lease.RunID, lease.Owner,
		lease.Epoch, now.UnixNano())
	if err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		attempt, getErr := s.GetAttemptWorktree(ctx, journal.AttemptKey())
		if getErr == nil && attempt.State == AttemptWorktreePromoted && attempt.ControllerOwner == lease.Owner && attempt.ControllerEpoch == lease.Epoch {
			return attempt, nil
		}
		return AttemptWorktreeRecord{}, errors.New("complete promotion attempt transition is stale or fenced")
	}
	return s.GetAttemptWorktree(ctx, journal.AttemptKey())
}

func verifyPromotionAuthorityConn(ctx context.Context, conn *sql.Conn, journal PromotionJournal, now time.Time, requireRatified bool) error {
	var attempt AttemptWorktreeRecord
	var created, updated int64
	var resources int
	err := conn.QueryRowContext(ctx, `SELECT delivery_id, generation, unit_id, attempt, branch, path,
		base_head, candidate_head, validated_head, state, controller_owner, controller_epoch,
		resources_created, failure_class, cleanup_error, created_at, updated_at FROM attempt_worktrees
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND attempt = ?`, journal.DeliveryID,
		journal.Generation, journal.UnitID, journal.Attempt).Scan(&attempt.DeliveryID, &attempt.Generation,
		&attempt.UnitID, &attempt.Attempt, &attempt.Branch, &attempt.Path, &attempt.BaseHead,
		&attempt.CandidateHead, &attempt.ValidatedHead, &attempt.State, &attempt.ControllerOwner,
		&attempt.ControllerEpoch, &resources, &attempt.FailureClass, &attempt.CleanupError, &created, &updated)
	if err != nil {
		return fmt.Errorf("read promotion attempt: %w", err)
	}
	if requireRatified && attempt.State != AttemptWorktreeRatified || !requireRatified && attempt.State != AttemptWorktreePromoting {
		return errors.New("promotion attempt is not in the required durable state")
	}
	if attempt.BaseHead != journal.BaseHead || attempt.CandidateHead != journal.CandidateHead ||
		attempt.ValidatedHead != journal.ValidatedHead || attempt.CandidateHead != attempt.ValidatedHead ||
		attempt.ControllerOwner != journal.ControllerOwner || attempt.ControllerEpoch != journal.ControllerEpoch {
		return errors.New("promotion attempt identity or controller mismatch")
	}
	var proof ArtifactProof
	var ratified int
	err = conn.QueryRowContext(ctx, `SELECT proof_id, delivery_id, generation, unit_id, attempt,
		start_head, candidate_head, validated_head, expected_artifact, artifact_hash, validator,
		thinking, ratified FROM artifact_proofs WHERE proof_id = ?`, journal.ProofID).Scan(&proof.ProofID,
		&proof.DeliveryID, &proof.Generation, &proof.UnitID, &proof.Attempt, &proof.StartHead,
		&proof.CandidateHead, &proof.ValidatedHead, &proof.ExpectedArtifact, &proof.ArtifactHash,
		&proof.Validator, &proof.Thinking, &ratified)
	if err != nil {
		return fmt.Errorf("read promotion proof: %w", err)
	}
	if ratified != 1 || proof.DeliveryID != journal.DeliveryID || proof.Generation != journal.Generation ||
		proof.UnitID != journal.UnitID || proof.Attempt != journal.Attempt || proof.StartHead != journal.BaseHead ||
		proof.CandidateHead != journal.CandidateHead || proof.ValidatedHead != journal.ValidatedHead ||
		proof.ArtifactHash != journal.EvidenceHash || proof.Validator != journal.AttestationValidator ||
		proof.Thinking != journal.AttestationThinking {
		return errors.New("promotion proof does not bind the journal attempt")
	}
	var evidence, session, base, candidate, observed, head, unitID, validator, thinking, verdict string
	var repository, baseBranch, contractHash string
	var pr int
	var generation, attestationAttempt, stateVersion, issued, expires int64
	var localGates, uat, milestoneValid int
	err = conn.QueryRowContext(ctx, `SELECT base_head, candidate_head, observed_head, head_sha,
		generation, unit_id, attempt, state_version, evidence_hash, validator_session_id,
		repository, pr, base_branch, contract_hash, created_at, validator, thinking, verdict, local_gates, uat, milestone_valid, expires_at FROM attestations
		WHERE run_id = ? AND head_sha = ?`, journal.DeliveryID, journal.CandidateHead).Scan(&base,
		&candidate, &observed, &head, &generation, &unitID, &attestationAttempt, &stateVersion, &evidence,
		&session, &repository, &pr, &baseBranch, &contractHash, &issued, &validator, &thinking, &verdict, &localGates, &uat, &milestoneValid, &expires)
	if err != nil {
		return fmt.Errorf("read promotion attestation: %w", err)
	}
	if base != journal.BaseHead || candidate != journal.CandidateHead || observed != journal.CandidateHead ||
		head != journal.CandidateHead || generation != journal.Generation || unitID != journal.UnitID || attestationAttempt != journal.Attempt ||
		stateVersion != journal.GovernanceStateVersion || evidence != journal.EvidenceHash || session != journal.ValidatorSessionID ||
		repository != journal.AttestationRepository || pr != journal.AttestationPR || baseBranch != journal.AttestationBaseBranch ||
		contractHash != journal.AttestationContractHash || issued != journal.AttestationCreatedAt.UnixNano() || validator != journal.AttestationValidator || thinking != journal.AttestationThinking || verdict != journal.AttestationVerdict ||
		(localGates == 1) != journal.AttestationLocalGates || (uat == 1) != journal.AttestationUAT ||
		(milestoneValid == 1) != journal.AttestationMilestoneValid || expires != journal.AttestationExpiresAt.UnixNano() || expires <= now.UnixNano() {
		return errors.New("promotion attestation is expired or does not bind the journal")
	}
	var governanceVersion int64
	if err := conn.QueryRowContext(ctx, `SELECT version FROM authority_state WHERE state_id = 'governance'`).Scan(&governanceVersion); err != nil {
		return fmt.Errorf("read governance state: %w", err)
	}
	if governanceVersion != journal.GovernanceStateVersion {
		return errors.New("promotion governance state changed")
	}
	return nil
}

func validatePromotionJournal(journal PromotionJournal) error {
	if err := validatePromotionJournalID(journal.JournalID); err != nil {
		return err
	}
	if err := validateAttemptWorktreeKey(journal.AttemptKey()); err != nil {
		return err
	}
	if !validGitSHA(journal.BaseHead) || !validGitSHA(journal.CandidateHead) ||
		!validGitSHA(journal.ValidatedHead) || journal.CandidateHead != journal.ValidatedHead ||
		journal.ProofID == "" || journal.AttestationRepository == "" || journal.AttestationPR <= 0 || journal.AttestationBaseBranch == "" ||
		!validSHA256Digest(journal.AttestationContractHash) || journal.AttestationCreatedAt.IsZero() ||
		journal.AttestationValidator == "" || journal.AttestationThinking == "" || journal.AttestationVerdict == "" ||
		strings.ContainsAny(journal.ProofID+journal.EvidenceHash+journal.ValidatorSessionID+journal.AttestationValidator+journal.AttestationThinking+journal.AttestationVerdict, "\r\n\x00") ||
		journal.GovernanceStateVersion <= 0 || journal.AttestationExpiresAt.IsZero() ||
		journal.State != PromotionJournalCreated || journal.ControllerOwner == "" || journal.ControllerEpoch <= 0 {
		return errors.New("complete immutable promotion journal identity is required")
	}
	if !validSHA256Digest(journal.EvidenceHash) {
		return errors.New("promotion evidence hash must be a sha256 digest")
	}
	manifestsEmpty := journal.ManifestJSON == "" && journal.ManifestHash == "" &&
		journal.BackupManifestJSON == "" && journal.BackupManifestHash == ""
	if !manifestsEmpty {
		if !validSHA256Digest(journal.ManifestHash) || !validSHA256Digest(journal.BackupManifestHash) {
			return errors.New("promotion manifest hashes must be sha256 digests")
		}
		if err := validateStoredManifest(journal.ManifestJSON); err != nil {
			return fmt.Errorf("invalid staged manifest: %w", err)
		}
		if digestManifestJSON(journal.ManifestJSON) != journal.ManifestHash {
			return errors.New("staged manifest JSON does not match its hash")
		}
		if err := validateStoredManifest(journal.BackupManifestJSON); err != nil {
			return fmt.Errorf("invalid backup manifest: %w", err)
		}
		if digestManifestJSON(journal.BackupManifestJSON) != journal.BackupManifestHash {
			return errors.New("backup manifest JSON does not match its hash")
		}
	}
	for _, path := range []string{journal.StagePath, journal.BackupPath, journal.CanonicalPath} {
		if !filepath.IsAbs(path) || filepath.Clean(path) != path || strings.ContainsAny(path, "\r\n\x00") {
			return errors.New("promotion journal paths must be absolute and clean")
		}
	}
	if journal.StagePath == journal.BackupPath || journal.StagePath == journal.CanonicalPath || journal.BackupPath == journal.CanonicalPath || filepath.Base(journal.CanonicalPath) != ".gsd" {
		return errors.New("promotion journal paths must be distinct and canonical")
	}
	repositoryRoot := filepath.Dir(journal.CanonicalPath)
	ownerParent := filepath.Dir(repositoryRoot)
	stageBase, backupBase := filepath.Base(journal.StagePath), filepath.Base(journal.BackupPath)
	if filepath.Dir(journal.StagePath) != ownerParent || filepath.Dir(journal.BackupPath) != ownerParent ||
		!strings.Contains(stageBase, ".shepherd-gsd-") || !strings.HasSuffix(stageBase, ".stage") ||
		!strings.HasSuffix(backupBase, ".backup") || strings.TrimSuffix(stageBase, ".stage") != strings.TrimSuffix(backupBase, ".backup") {
		return errors.New("promotion stage and backup ownership does not match canonical state")
	}
	return nil
}

func validateStoredManifest(value string) error {
	if len(value) == 0 || len(value) > 2*1024*1024 {
		return errors.New("manifest JSON exceeds bound")
	}
	var entries []struct {
		Path string `json:"path"`
		Type string `json:"type"`
		Size int64  `json:"size"`
		Hash string `json:"hash"`
	}
	if err := json.Unmarshal([]byte(value), &entries); err != nil {
		return err
	}
	if len(entries) > 4096 {
		return errors.New("manifest entry count exceeds bound")
	}
	previous := ""
	for _, entry := range entries {
		if !filepath.IsLocal(filepath.FromSlash(entry.Path)) || strings.ContainsAny(entry.Path, "\r\n\x00") || entry.Path <= previous ||
			(entry.Type != "file" && entry.Type != "dir") || entry.Size < 0 || entry.Type == "file" && !validSHA256Digest(entry.Hash) {
			return errors.New("manifest contains unsafe or non-deterministic entry")
		}
		previous = entry.Path
	}
	return nil
}

func validatePromotionJournalID(value string) error {
	if strings.TrimSpace(value) == "" || len(value) > 200 || strings.ContainsAny(value, "/\\\r\n\x00") {
		return errors.New("safe promotion journal ID is required")
	}
	return nil
}

func promotionBool(value bool) int {
	if value {
		return 1
	}
	return 0
}

func digestManifestJSON(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validSHA256Digest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != 71 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func legalPromotionTransition(current, next PromotionJournalState) bool {
	if next == PromotionJournalBlocked {
		return current != PromotionJournalBlocked
	}
	allowed := map[PromotionJournalState]PromotionJournalState{
		PromotionJournalCreated:          PromotionJournalStateStaged,
		PromotionJournalStateStaged:      PromotionJournalGitPromoting,
		PromotionJournalGitPromoting:     PromotionJournalGitPromoted,
		PromotionJournalGitPromoted:      PromotionJournalStateSwapStarted,
		PromotionJournalStateSwapStarted: PromotionJournalStateInstalled,
		PromotionJournalStateInstalled:   PromotionJournalComplete,
	}
	return allowed[current] == next
}

const promotionJournalSelect = `SELECT journal_id, delivery_id, generation, unit_id, attempt,
	base_head, candidate_head, validated_head, proof_id, evidence_hash, validator_session_id,
	attestation_repository, attestation_pr, attestation_base_branch, attestation_contract_hash, attestation_created_at,
	attestation_validator, attestation_thinking, attestation_verdict, attestation_local_gates,
	attestation_uat, attestation_milestone_valid, governance_state_version, attestation_expires_at, manifest_json, manifest_hash,
	backup_manifest_json, backup_manifest_hash, stage_path, backup_path, canonical_path,
	state, blocked_reason, cleanup_complete, decisions_resolved, controller_owner, controller_epoch, created_at, updated_at
	FROM promotion_journals`

type promotionJournalScanner interface{ Scan(...any) error }

func scanPromotionJournal(scanner promotionJournalScanner) (PromotionJournal, error) {
	var journal PromotionJournal
	var expires, attestationCreated, created, updated int64
	var cleanupComplete, decisionsResolved, localGates, uat, milestoneValid int
	err := scanner.Scan(&journal.JournalID, &journal.DeliveryID, &journal.Generation, &journal.UnitID,
		&journal.Attempt, &journal.BaseHead, &journal.CandidateHead, &journal.ValidatedHead,
		&journal.ProofID, &journal.EvidenceHash, &journal.ValidatorSessionID,
		&journal.AttestationRepository, &journal.AttestationPR, &journal.AttestationBaseBranch,
		&journal.AttestationContractHash, &attestationCreated, &journal.AttestationValidator, &journal.AttestationThinking, &journal.AttestationVerdict,
		&localGates, &uat, &milestoneValid, &journal.GovernanceStateVersion, &expires, &journal.ManifestJSON, &journal.ManifestHash,
		&journal.BackupManifestJSON, &journal.BackupManifestHash, &journal.StagePath,
		&journal.BackupPath, &journal.CanonicalPath, &journal.State, &journal.BlockedReason,
		&cleanupComplete, &decisionsResolved, &journal.ControllerOwner, &journal.ControllerEpoch, &created, &updated)
	if err != nil {
		return PromotionJournal{}, err
	}
	journal.AttestationCreatedAt = time.Unix(0, attestationCreated).UTC()
	journal.AttestationExpiresAt = time.Unix(0, expires).UTC()
	journal.AttestationLocalGates, journal.AttestationUAT = localGates == 1, uat == 1
	journal.AttestationMilestoneValid = milestoneValid == 1
	journal.CleanupComplete = cleanupComplete == 1
	journal.DecisionsResolved = decisionsResolved == 1
	journal.CreatedAt, journal.UpdatedAt = time.Unix(0, created).UTC(), time.Unix(0, updated).UTC()
	return journal, nil
}

func getPromotionJournalConn(ctx context.Context, conn *sql.Conn, journalID string) (PromotionJournal, error) {
	return scanPromotionJournal(conn.QueryRowContext(ctx, promotionJournalSelect+" WHERE journal_id = ?", journalID))
}

func samePromotionIdentity(left, right PromotionJournal) bool {
	return left.JournalID == right.JournalID && left.DeliveryID == right.DeliveryID &&
		left.Generation == right.Generation && left.UnitID == right.UnitID && left.Attempt == right.Attempt &&
		left.BaseHead == right.BaseHead && left.CandidateHead == right.CandidateHead &&
		left.ValidatedHead == right.ValidatedHead && left.ProofID == right.ProofID &&
		left.EvidenceHash == right.EvidenceHash && left.ValidatorSessionID == right.ValidatorSessionID &&
		left.AttestationRepository == right.AttestationRepository && left.AttestationPR == right.AttestationPR &&
		left.AttestationBaseBranch == right.AttestationBaseBranch && left.AttestationContractHash == right.AttestationContractHash &&
		left.AttestationCreatedAt.Equal(right.AttestationCreatedAt) && left.AttestationValidator == right.AttestationValidator && left.AttestationThinking == right.AttestationThinking &&
		left.AttestationVerdict == right.AttestationVerdict && left.AttestationLocalGates == right.AttestationLocalGates &&
		left.AttestationUAT == right.AttestationUAT && left.AttestationMilestoneValid == right.AttestationMilestoneValid &&
		left.GovernanceStateVersion == right.GovernanceStateVersion &&
		left.AttestationExpiresAt.Equal(right.AttestationExpiresAt) && left.ManifestJSON == right.ManifestJSON &&
		left.ManifestHash == right.ManifestHash && left.BackupManifestJSON == right.BackupManifestJSON &&
		left.BackupManifestHash == right.BackupManifestHash && left.StagePath == right.StagePath &&
		left.BackupPath == right.BackupPath && left.CanonicalPath == right.CanonicalPath
}
