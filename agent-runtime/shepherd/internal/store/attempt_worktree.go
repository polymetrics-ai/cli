package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
)

// AttemptWorktreeState is the durable lifecycle of one isolated unit attempt.
type AttemptWorktreeState string

const (
	AttemptWorktreeCreated             AttemptWorktreeState = "created"
	AttemptWorktreePrepared            AttemptWorktreeState = "prepared"
	AttemptWorktreeRunning             AttemptWorktreeState = "running"
	AttemptWorktreeValidated           AttemptWorktreeState = "validated"
	AttemptWorktreeRatified            AttemptWorktreeState = "ratified"
	AttemptWorktreePromoting           AttemptWorktreeState = "promoting"
	AttemptWorktreePromoted            AttemptWorktreeState = "promoted"
	AttemptWorktreeRetainedForRecovery AttemptWorktreeState = "retained_for_recovery"
	AttemptWorktreeCleanupPending      AttemptWorktreeState = "cleanup_pending"
	AttemptWorktreeCleanupComplete     AttemptWorktreeState = "cleanup_complete"
	AttemptWorktreeCleanupBlocked      AttemptWorktreeState = "cleanup_blocked"
)

type AttemptWorktreeKey struct {
	DeliveryID string
	Generation int64
	UnitID     string
	Attempt    int64
}

type AttemptWorktreeRecord struct {
	DeliveryID       string
	Generation       int64
	UnitID           string
	Attempt          int64
	Branch           string
	Path             string
	BaseHead         string
	CandidateHead    string
	ValidatedHead    string
	State            AttemptWorktreeState
	ControllerOwner  string
	ControllerEpoch  int64
	ResourcesCreated bool
	FailureClass     string
	CleanupError     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (r AttemptWorktreeRecord) Key() AttemptWorktreeKey {
	return AttemptWorktreeKey{DeliveryID: r.DeliveryID, Generation: r.Generation, UnitID: r.UnitID, Attempt: r.Attempt}
}

type AttemptWorktreeUpdate struct {
	CandidateHead string
	ValidatedHead string
	FailureClass  string
	CleanupError  string
}

func (s *Store) CreateAttemptWorktree(ctx context.Context, record AttemptWorktreeRecord) error {
	if err := validateAttemptWorktreeRecord(record); err != nil {
		return err
	}
	now := time.Now().UTC().UnixNano()
	var leaseCount int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM leases WHERE run_id = ? AND owner = ?
		AND epoch = ? AND expires_at > ?`, record.DeliveryID, record.ControllerOwner,
		record.ControllerEpoch, now).Scan(&leaseCount); err != nil || leaseCount != 1 {
		return errors.New("attempt controller lease is stale, expired, or fenced")
	}
	_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO attempt_worktrees
		(delivery_id, generation, unit_id, attempt, branch, path, base_head, candidate_head,
		 validated_head, state, controller_owner, controller_epoch, resources_created, failure_class, cleanup_error,
		 created_at, updated_at) SELECT ?, ?, ?, ?, ?, ?, ?, '', '', ?, ?, ?, 0, '', '', ?, ?
		 WHERE EXISTS (SELECT 1 FROM leases WHERE run_id = ? AND owner = ? AND epoch = ? AND expires_at > ?)`,
		record.DeliveryID, record.Generation, record.UnitID, record.Attempt, record.Branch,
		record.Path, record.BaseHead, AttemptWorktreeCreated, record.ControllerOwner,
		record.ControllerEpoch, now, now, record.DeliveryID, record.ControllerOwner,
		record.ControllerEpoch, now)
	if err != nil {
		return fmt.Errorf("create attempt worktree record: %w", err)
	}
	loaded, err := s.GetAttemptWorktree(ctx, record.Key())
	if err != nil {
		return err
	}
	if loaded.Branch != record.Branch || loaded.Path != filepath.Clean(record.Path) ||
		loaded.BaseHead != record.BaseHead || loaded.ControllerOwner != record.ControllerOwner ||
		loaded.ControllerEpoch != record.ControllerEpoch {
		return errors.New("attempt worktree identity or ownership cannot be rebound")
	}
	return nil
}

func (s *Store) GetAttemptWorktree(ctx context.Context, key AttemptWorktreeKey) (AttemptWorktreeRecord, error) {
	if err := validateAttemptWorktreeKey(key); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	return scanAttemptWorktree(s.db.QueryRowContext(ctx, `SELECT delivery_id, generation, unit_id,
		attempt, branch, path, base_head, candidate_head, validated_head, state, controller_owner,
		controller_epoch, resources_created, failure_class, cleanup_error, created_at, updated_at
		FROM attempt_worktrees WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND attempt = ?`,
		key.DeliveryID, key.Generation, key.UnitID, key.Attempt))
}

func (s *Store) ListAttemptWorktrees(ctx context.Context, deliveryID string) ([]AttemptWorktreeRecord, error) {
	if strings.TrimSpace(deliveryID) == "" || strings.ContainsAny(deliveryID, "\r\n\x00") {
		return nil, errors.New("delivery ID is required")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT delivery_id, generation, unit_id, attempt, branch,
		path, base_head, candidate_head, validated_head, state, controller_owner, controller_epoch,
		resources_created, failure_class, cleanup_error, created_at, updated_at FROM attempt_worktrees
		WHERE delivery_id = ? ORDER BY generation, unit_id, attempt LIMIT 1025`, deliveryID)
	if err != nil {
		return nil, fmt.Errorf("list attempt worktrees: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var records []AttemptWorktreeRecord
	for rows.Next() {
		record, err := scanAttemptWorktree(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
		if len(records) > 1024 {
			return nil, errors.New("attempt worktree reconciliation set exceeds the governed limit")
		}
	}
	return records, rows.Err()
}

func (s *Store) ConfirmAttemptWorktreeResources(ctx context.Context, key AttemptWorktreeKey, owner string, epoch int64) (AttemptWorktreeRecord, error) {
	if err := validateAttemptWorktreeKey(key); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	now := time.Now().UTC().UnixNano()
	result, err := s.db.ExecContext(ctx, `UPDATE attempt_worktrees SET resources_created = 1, updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND attempt = ? AND state = ?
		AND controller_owner = ? AND controller_epoch = ? AND resources_created = 0
		AND EXISTS (SELECT 1 FROM leases WHERE run_id = ? AND owner = ? AND epoch = ? AND expires_at > ?)`,
		now, key.DeliveryID, key.Generation, key.UnitID, key.Attempt, AttemptWorktreeCreated,
		owner, epoch, key.DeliveryID, owner, epoch, now)
	if err != nil {
		return AttemptWorktreeRecord{}, fmt.Errorf("confirm attempt worktree resources: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return AttemptWorktreeRecord{}, errors.New("attempt worktree resource confirmation is stale or invalid")
	}
	return s.GetAttemptWorktree(ctx, key)
}

func (s *Store) TransitionAttemptWorktree(ctx context.Context, key AttemptWorktreeKey, owner string, epoch int64, next AttemptWorktreeState, update AttemptWorktreeUpdate) (AttemptWorktreeRecord, error) {
	if err := validateAttemptWorktreeKey(key); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if owner == "" || strings.ContainsAny(owner, "\r\n\x00") || epoch <= 0 {
		return AttemptWorktreeRecord{}, errors.New("attempt controller ownership is required")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return AttemptWorktreeRecord{}, err
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	current, err := getAttemptWorktreeConn(ctx, conn, key)
	if err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if current.ControllerOwner != owner || current.ControllerEpoch != epoch {
		return AttemptWorktreeRecord{}, errors.New("attempt worktree transition is fenced by controller ownership")
	}
	if err := checkAttemptLeaseConn(ctx, conn, key.DeliveryID, owner, epoch, time.Now().UTC()); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if next == AttemptWorktreePrepared && !current.ResourcesCreated {
		return AttemptWorktreeRecord{}, errors.New("attempt resources are not confirmed as Shepherd-created")
	}
	if !legalAttemptTransition(current.State, next) {
		return AttemptWorktreeRecord{}, fmt.Errorf("illegal attempt worktree transition %s -> %s", current.State, next)
	}
	candidate, validated, err := transitionHeads(current, update, next)
	if err != nil {
		return AttemptWorktreeRecord{}, err
	}
	failure := current.FailureClass
	if update.FailureClass != "" {
		failure = boundedAttemptDiagnostic(update.FailureClass)
	}
	cleanup := current.CleanupError
	if update.CleanupError != "" {
		cleanup = boundedAttemptDiagnostic(update.CleanupError)
	}
	now := time.Now().UTC().UnixNano()
	result, err := conn.ExecContext(ctx, `UPDATE attempt_worktrees SET state = ?, candidate_head = ?,
		validated_head = ?, failure_class = ?, cleanup_error = ?, updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND attempt = ?
		AND state = ? AND controller_owner = ? AND controller_epoch = ?`, next, candidate, validated,
		failure, cleanup, now, key.DeliveryID, key.Generation, key.UnitID, key.Attempt,
		current.State, owner, epoch)
	if err != nil {
		return AttemptWorktreeRecord{}, fmt.Errorf("transition attempt worktree: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return AttemptWorktreeRecord{}, errors.New("attempt worktree transition was concurrently fenced")
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	committed = true
	return getAttemptWorktreeConn(ctx, conn, key)
}

func (s *Store) ClaimAttemptWorktreeCleanup(ctx context.Context, key AttemptWorktreeKey, expectedOwner string, expectedEpoch int64, newOwner string, newEpoch int64) (AttemptWorktreeRecord, error) {
	if err := validateAttemptWorktreeKey(key); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if expectedOwner == "" || newOwner == "" || strings.ContainsAny(expectedOwner+newOwner, "\r\n\x00") || expectedEpoch <= 0 || newEpoch <= expectedEpoch {
		return AttemptWorktreeRecord{}, errors.New("strictly newer cleanup controller ownership is required")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return AttemptWorktreeRecord{}, err
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	if err := checkAttemptLeaseConn(ctx, conn, key.DeliveryID, newOwner, newEpoch, time.Now().UTC()); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	current, err := getAttemptWorktreeConn(ctx, conn, key)
	if err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if current.ControllerOwner != expectedOwner || current.ControllerEpoch != expectedEpoch || !cleanupClaimAllowed(current) {
		return AttemptWorktreeRecord{}, errors.New("attempt cleanup claim is stale, live, ambiguous, or complete")
	}
	now := time.Now().UTC().UnixNano()
	result, err := conn.ExecContext(ctx, `UPDATE attempt_worktrees SET state = ?, controller_owner = ?,
		controller_epoch = ?, failure_class = CASE WHEN failure_class = '' THEN 'restart_reconciliation' ELSE failure_class END,
		updated_at = ? WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND attempt = ?
		AND controller_owner = ? AND controller_epoch = ? AND state = ?`, AttemptWorktreeCleanupPending,
		newOwner, newEpoch, now, key.DeliveryID, key.Generation, key.UnitID, key.Attempt,
		expectedOwner, expectedEpoch, current.State)
	if err != nil {
		return AttemptWorktreeRecord{}, fmt.Errorf("claim attempt cleanup: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return AttemptWorktreeRecord{}, errors.New("attempt cleanup claim was concurrently fenced")
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	committed = true
	return getAttemptWorktreeConn(ctx, conn, key)
}

func (s *Store) RecordAttemptWorktreeCandidate(ctx context.Context, key AttemptWorktreeKey, owner string, epoch int64, candidateHead string) (AttemptWorktreeRecord, error) {
	if err := validateAttemptWorktreeKey(key); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if owner == "" || epoch <= 0 || !validGitSHA(candidateHead) {
		return AttemptWorktreeRecord{}, errors.New("owned full candidate head is required")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return AttemptWorktreeRecord{}, err
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	current, err := getAttemptWorktreeConn(ctx, conn, key)
	if err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if current.ControllerOwner != owner || current.ControllerEpoch != epoch || current.State != AttemptWorktreeRunning {
		return AttemptWorktreeRecord{}, errors.New("attempt candidate record is stale or not running")
	}
	if err := checkAttemptLeaseConn(ctx, conn, key.DeliveryID, owner, epoch, time.Now().UTC()); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if current.CandidateHead != "" && current.CandidateHead != candidateHead {
		return AttemptWorktreeRecord{}, errors.New("attempt candidate head cannot be rebound")
	}
	result, err := conn.ExecContext(ctx, `UPDATE attempt_worktrees SET candidate_head = ?, updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND attempt = ? AND state = ?
		AND controller_owner = ? AND controller_epoch = ? AND (candidate_head = '' OR candidate_head = ?)`,
		candidateHead, time.Now().UTC().UnixNano(), key.DeliveryID, key.Generation, key.UnitID, key.Attempt,
		AttemptWorktreeRunning, owner, epoch, candidateHead)
	if err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return AttemptWorktreeRecord{}, errors.New("attempt candidate record was concurrently fenced")
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	committed = true
	return getAttemptWorktreeConn(ctx, conn, key)
}

func (s *Store) RatifyAttemptWorktree(ctx context.Context, key AttemptWorktreeKey, owner string, epoch int64, proof ArtifactProof, attestation AttestationRecord) (AttemptWorktreeRecord, error) {
	if err := validateAttemptWorktreeKey(key); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if owner == "" || epoch <= 0 || proof.DeliveryID != key.DeliveryID || proof.Generation != key.Generation ||
		proof.UnitID != key.UnitID || proof.Attempt != key.Attempt || attestation.RunID != key.DeliveryID ||
		attestation.Generation != key.Generation || attestation.UnitID != key.UnitID || attestation.Attempt != key.Attempt {
		return AttemptWorktreeRecord{}, errors.New("ratified evidence does not match attempt identity")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return AttemptWorktreeRecord{}, err
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	current, err := getAttemptWorktreeConn(ctx, conn, key)
	if err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if current.ControllerOwner != owner || current.ControllerEpoch != epoch || current.State != AttemptWorktreeValidated {
		return AttemptWorktreeRecord{}, errors.New("attempt ratification is stale or not validated")
	}
	if err := checkAttemptLeaseConn(ctx, conn, key.DeliveryID, owner, epoch, time.Now().UTC()); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if current.CandidateHead == "" || current.ValidatedHead != current.CandidateHead ||
		proof.StartHead != current.BaseHead || proof.CandidateHead != current.CandidateHead ||
		proof.ValidatedHead != current.ValidatedHead || attestation.BaseHead != current.BaseHead ||
		attestation.CandidateHead != current.CandidateHead || attestation.HeadSHA != current.ValidatedHead {
		return AttemptWorktreeRecord{}, errors.New("ratified evidence heads do not match durable attempt heads")
	}
	if err := putArtifactProof(ctx, conn, proof); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if err := putAttestation(ctx, conn, attestation); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	result, err := conn.ExecContext(ctx, `UPDATE attempt_worktrees SET state = ?, updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND attempt = ? AND state = ?
		AND controller_owner = ? AND controller_epoch = ?`, AttemptWorktreeRatified,
		time.Now().UTC().UnixNano(), key.DeliveryID, key.Generation, key.UnitID, key.Attempt,
		AttemptWorktreeValidated, owner, epoch)
	if err != nil {
		return AttemptWorktreeRecord{}, err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return AttemptWorktreeRecord{}, errors.New("attempt ratification was concurrently fenced")
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return AttemptWorktreeRecord{}, err
	}
	committed = true
	return getAttemptWorktreeConn(ctx, conn, key)
}

func (s *Store) ResolveHumanClearedAttempt(ctx context.Context, key AttemptWorktreeKey, expectedOwner string, expectedEpoch int64, lease Lease) error {
	if err := validateAttemptWorktreeKey(key); err != nil {
		return err
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	if err := checkAttemptLeaseConn(ctx, conn, key.DeliveryID, lease.Owner, lease.Epoch, time.Now().UTC()); err != nil {
		return err
	}
	current, err := getAttemptWorktreeConn(ctx, conn, key)
	if err != nil {
		return err
	}
	ambiguous := current.State == AttemptWorktreeRunning || current.State == AttemptWorktreePromoting || !current.ResourcesCreated
	if current.ControllerOwner != expectedOwner || current.ControllerEpoch != expectedEpoch || !ambiguous {
		return errors.New("attempt is not an exact human-cleared ambiguous record")
	}
	result, err := conn.ExecContext(ctx, `UPDATE attempt_worktrees SET state = ?, controller_owner = ?,
		controller_epoch = ?, cleanup_error = 'human_confirmed_resources_absent', updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND attempt = ?
		AND state = ? AND controller_owner = ? AND controller_epoch = ?`, AttemptWorktreeCleanupComplete,
		lease.Owner, lease.Epoch, time.Now().UTC().UnixNano(), key.DeliveryID, key.Generation, key.UnitID,
		key.Attempt, current.State, expectedOwner, expectedEpoch)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return errors.New("human-cleared attempt resolution was concurrently fenced")
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Store) ReconcileInterruptedDelivery(ctx context.Context, lease Lease, target domain.RunState) error {
	if target != domain.RunReady && target != domain.RunAwaitingDecision {
		return errors.New("interrupted delivery must become ready or awaiting_decision")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	if err := checkAttemptLeaseConn(ctx, conn, lease.RunID, lease.Owner, lease.Epoch, time.Now().UTC()); err != nil {
		return err
	}
	var state domain.RunState
	var generation int64
	var owner string
	if err := conn.QueryRowContext(ctx, `SELECT state, generation, owner FROM delivery_runs WHERE delivery_id = ?`, lease.RunID).Scan(&state, &generation, &owner); err != nil {
		return err
	}
	if state != domain.RunRunning || owner == lease.Owner {
		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			return err
		}
		committed = true
		return nil
	}
	if err := domain.ValidateRunTransition(state, target); err != nil {
		return err
	}
	now := time.Now().UTC().UnixNano()
	result, err := conn.ExecContext(ctx, `UPDATE delivery_runs SET state = ?, owner = ?, updated_at = ?
		WHERE delivery_id = ? AND state = ? AND owner = ?`, target, lease.Owner, now,
		lease.RunID, domain.RunRunning, owner)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return errors.New("interrupted delivery reconciliation was concurrently fenced")
	}
	if _, err := conn.ExecContext(ctx, `UPDATE unit_attempts SET status = 'terminal',
		last_failure = 'controller_interrupted', updated_at = ? WHERE delivery_id = ?
		AND generation = ? AND status = 'running'`, now, lease.RunID, generation); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, `UPDATE recovery_attempts SET status = 'used_failed', updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND status = 'dispatched' AND dispatch_claim = ?
		AND dispatch_epoch > 0`, now, lease.RunID, generation, owner); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return err
	}
	committed = true
	return nil
}

func cleanupClaimAllowed(record AttemptWorktreeRecord) bool {
	if !record.ResourcesCreated {
		return false
	}
	switch record.State {
	case AttemptWorktreeCreated, AttemptWorktreePrepared, AttemptWorktreeValidated,
		AttemptWorktreeRatified, AttemptWorktreePromoted, AttemptWorktreeRetainedForRecovery,
		AttemptWorktreeCleanupPending, AttemptWorktreeCleanupBlocked:
		return true
	default:
		// Running can still own a live validator, and promoting is ambiguous until
		// Slice C adds a promotion journal. Both remain fail-closed for human recovery.
		return false
	}
}

func legalAttemptTransition(current, next AttemptWorktreeState) bool {
	allowed := map[AttemptWorktreeState]map[AttemptWorktreeState]bool{
		AttemptWorktreeCreated:             {AttemptWorktreePrepared: true, AttemptWorktreeRetainedForRecovery: true},
		AttemptWorktreePrepared:            {AttemptWorktreeRunning: true, AttemptWorktreeRetainedForRecovery: true},
		AttemptWorktreeRunning:             {AttemptWorktreeValidated: true, AttemptWorktreeRetainedForRecovery: true},
		AttemptWorktreeValidated:           {AttemptWorktreeRetainedForRecovery: true},
		AttemptWorktreeRatified:            {AttemptWorktreePromoting: true, AttemptWorktreeRetainedForRecovery: true},
		AttemptWorktreePromoting:           {AttemptWorktreePromoted: true},
		AttemptWorktreePromoted:            {AttemptWorktreeCleanupPending: true},
		AttemptWorktreeRetainedForRecovery: {AttemptWorktreeCleanupPending: true},
		AttemptWorktreeCleanupPending:      {AttemptWorktreeCleanupComplete: true, AttemptWorktreeCleanupBlocked: true},
	}
	return allowed[current][next]
}

func transitionHeads(current AttemptWorktreeRecord, update AttemptWorktreeUpdate, next AttemptWorktreeState) (string, string, error) {
	candidate, validated := current.CandidateHead, current.ValidatedHead
	if update.CandidateHead != "" {
		candidate = update.CandidateHead
	}
	if update.ValidatedHead != "" {
		validated = update.ValidatedHead
	}
	if candidate != "" && !validGitSHA(candidate) || validated != "" && !validGitSHA(validated) {
		return "", "", errors.New("attempt candidate and validated heads must be full Git SHAs")
	}
	if current.CandidateHead != "" && candidate != current.CandidateHead || current.ValidatedHead != "" && validated != current.ValidatedHead {
		return "", "", errors.New("attempt candidate or validated head cannot be rebound")
	}
	if next == AttemptWorktreeValidated || next == AttemptWorktreeRatified || next == AttemptWorktreePromoting || next == AttemptWorktreePromoted {
		if candidate == "" || validated == "" || candidate != validated {
			return "", "", errors.New("validated lifecycle states require one exact candidate head")
		}
	}
	return candidate, validated, nil
}

func validateAttemptWorktreeRecord(record AttemptWorktreeRecord) error {
	if err := validateAttemptWorktreeKey(record.Key()); err != nil {
		return err
	}
	if record.State != AttemptWorktreeCreated || record.Branch == "" || strings.HasPrefix(record.Branch, "-") ||
		strings.ContainsAny(record.Branch, "\r\n\x00") || !filepath.IsAbs(record.Path) ||
		filepath.Clean(record.Path) != record.Path || !validGitSHA(record.BaseHead) || record.ControllerOwner == "" ||
		strings.ContainsAny(record.ControllerOwner, "\r\n\x00") || record.ControllerEpoch <= 0 {
		return errors.New("complete immutable attempt worktree identity is required")
	}
	return nil
}

func validateAttemptWorktreeKey(key AttemptWorktreeKey) error {
	if key.DeliveryID == "" || key.Generation <= 0 || strings.TrimSpace(key.UnitID) == "" || key.Attempt <= 0 ||
		strings.ContainsAny(key.DeliveryID+key.UnitID, "\r\n\x00") {
		return errors.New("complete attempt worktree key is required")
	}
	return nil
}

type attemptWorktreeScanner interface {
	Scan(dest ...any) error
}

func scanAttemptWorktree(scanner attemptWorktreeScanner) (AttemptWorktreeRecord, error) {
	var record AttemptWorktreeRecord
	var resourcesCreated int
	var created, updated int64
	err := scanner.Scan(&record.DeliveryID, &record.Generation, &record.UnitID, &record.Attempt,
		&record.Branch, &record.Path, &record.BaseHead, &record.CandidateHead, &record.ValidatedHead,
		&record.State, &record.ControllerOwner, &record.ControllerEpoch, &resourcesCreated,
		&record.FailureClass, &record.CleanupError, &created, &updated)
	if err != nil {
		return AttemptWorktreeRecord{}, fmt.Errorf("read attempt worktree: %w", err)
	}
	record.ResourcesCreated = resourcesCreated == 1
	record.CreatedAt, record.UpdatedAt = time.Unix(0, created).UTC(), time.Unix(0, updated).UTC()
	return record, nil
}

func checkAttemptLeaseConn(ctx context.Context, conn *sql.Conn, deliveryID, owner string, epoch int64, now time.Time) error {
	var count int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM leases WHERE run_id = ? AND owner = ?
		AND epoch = ? AND expires_at > ?`, deliveryID, owner, epoch, now.UnixNano()).Scan(&count); err != nil {
		return fmt.Errorf("verify attempt controller lease: %w", err)
	}
	if count != 1 {
		return errors.New("attempt controller lease is stale, expired, or fenced")
	}
	return nil
}

func getAttemptWorktreeConn(ctx context.Context, conn *sql.Conn, key AttemptWorktreeKey) (AttemptWorktreeRecord, error) {
	return scanAttemptWorktree(conn.QueryRowContext(ctx, `SELECT delivery_id, generation, unit_id,
		attempt, branch, path, base_head, candidate_head, validated_head, state, controller_owner,
		controller_epoch, resources_created, failure_class, cleanup_error, created_at, updated_at FROM attempt_worktrees
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND attempt = ?`,
		key.DeliveryID, key.Generation, key.UnitID, key.Attempt))
}

func boundedAttemptDiagnostic(value string) string {
	value = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, strings.TrimSpace(value))
	for len(value) > 512 {
		_, size := utf8.DecodeLastRuneInString(value)
		value = value[:len(value)-size]
	}
	return value
}
