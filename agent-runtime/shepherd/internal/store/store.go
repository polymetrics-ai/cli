package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/recovery"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/sensitive"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type Lease struct {
	RunID     string
	Owner     string
	Epoch     int64
	ExpiresAt time.Time
}

type Delivery struct {
	ID             string
	Issue          int
	ParentIssue    int
	Repository     string
	PullRequest    int
	WorkDir        string
	ContextHash    string
	MilestoneID    string
	Branch         string
	BaseBranch     string
	GSDProjectRoot string
	InitialHead    string
	GSDVersion     string
}

type DeliveryRun struct {
	DeliveryID string
	State      domain.RunState
	Generation int64
	Attempt    int64
	Owner      string
}

var (
	ErrRetryBudgetExhausted    = errors.New("unit retry budget exhausted")
	ErrRecoveryClaimFenced     = errors.New("recovery attempt claim is fenced")
	ErrRecoveryBackoffPending  = errors.New("recovery backoff is pending")
	ErrRecoveryDispatchClaimed = errors.New("recovery dispatch is already claimed")
	ErrRecoveryDecisionPending = errors.New("recovery decision is incomplete")
	ErrRecoveryTerminal        = errors.New("recovery decision blocks dispatch")
	ErrLegacyExternalEffects   = errors.New("legacy external effects require human reconciliation")
	ErrLegacyDeliveryTarget    = errors.New("legacy delivery target requires human reconciliation")
	ErrLegacyDecisionReply     = errors.New("legacy decision reply requires human reconciliation")
)

type UnitAttemptKey struct {
	DeliveryID string
	Generation int64
	UnitID     string
	HeadSHA    string
}

type UnitAttempt struct {
	UnitAttemptKey
	Attempts  int64
	Remaining int64
	Status    string
	Failure   string
}

type UnitAttemptIdentity struct {
	UnitAttemptKey
	Attempt            int64
	Model              string
	Thinking           string
	SessionID          string
	SessionFingerprint string
	StartedAt          time.Time
	EndedAt            time.Time
}

type DecisionRequestStatus string

const (
	DecisionRequestOpen      DecisionRequestStatus = "open"
	DecisionRequestPublished DecisionRequestStatus = "published"
	DecisionRequestAnswered  DecisionRequestStatus = "answered"
	DecisionRequestConsumed  DecisionRequestStatus = "consumed"
	DecisionRequestExpired   DecisionRequestStatus = "expired"
	DecisionRequestCancelled DecisionRequestStatus = "cancelled"
	DecisionRequestRejected  DecisionRequestStatus = "rejected"
)

type DecisionRequest struct {
	RequestID         string
	DeliveryID        string
	Repository        string
	Issue             int
	PullRequest       int
	UnitID            string
	Generation        int64
	HeadSHA           string
	Kind              string
	Evidence          string
	Options           []string
	RecommendedOption string
	SafeDefault       string
	ExpiresAt         time.Time
	GitHubCommentID   int64
	Status            DecisionRequestStatus
	AcceptedAnswer    string
	AcceptedBy        string
	AcceptedActorID   int64
	AcceptedCommentID int64
	AcceptedAt        time.Time
	ConsumedAt        time.Time
}

type RecoveryBudgetKey struct {
	DeliveryID   string
	Generation   int64
	UnitID       string
	HeadSHA      string
	FailureClass string
}

type RecoveryBudget struct {
	RecoveryBudgetKey
	PolicyVersion             int
	Attempts                  int64
	MaxAttempts               int64
	BaseBackoff               time.Duration
	MaxBackoff                time.Duration
	Backoff                   time.Duration
	LastFailureHash           string
	LastDiagnostic            string
	PlannerEvidenceHash       string
	PlannerSessionID          string
	PlannerSessionFingerprint string
	ObservedModel             string
	Thinking                  string
	SelectedAction            recovery.Action
	BoundedPlan               []recovery.PlanStep
	NextRetryAt               time.Time
	ExhaustedAt               time.Time
	Status                    string
}

type RecoveryReservation struct {
	RecoveryBudgetKey
	UnitAttempt     int64
	ClaimToken      string
	ControllerOwner string
	ControllerEpoch int64
	PolicyVersion   int
	MaxAttempts     int64
	BaseBackoff     time.Duration
	MaxBackoff      time.Duration
	FailureHash     string
	Diagnostic      string
	Reversible      bool
	ExhaustedAction recovery.Action
	Now             time.Time
}

type RecoveryOutcome struct {
	RecoveryBudgetKey
	UnitAttempt               int64
	ClaimToken                string
	ControllerOwner           string
	ControllerEpoch           int64
	PlannerRequestNonce       string
	EvidenceHash              string
	AuthorityScopeHash        string
	PlannerEvidenceHash       string
	PlannerSessionID          string
	PlannerSessionFingerprint string
	ObservedModel             string
	Thinking                  string
	SelectedAction            recovery.Action
	BoundedPlan               []recovery.PlanStep
	IssuedAt                  time.Time
	ExpiresAt                 time.Time
}

type RecoveryAttempt struct {
	RecoveryOutcome
	ReservationIndex int64
	Sequence         int64
	FailureHash      string
	Diagnostic       string
	Reversible       bool
	Backoff          time.Duration
	NextRetryAt      time.Time
	Status           string
	RejectedReason   string
	DispatchedAt     time.Time
	DispatchClaim    string
	DispatchEpoch    int64
}

type ArtifactProof struct {
	ProofID          string
	DeliveryID       string
	Generation       int64
	UnitID           string
	Attempt          int64
	StartHead        string
	CandidateHead    string
	ValidatedHead    string
	ExpectedArtifact string
	ArtifactHash     string
	Validator        string
	Thinking         string
	Ratified         bool
}

type AttestationRecord struct {
	Repository         string
	PR                 int
	BaseBranch         string
	BaseHead           string
	CandidateHead      string
	ObservedHead       string
	RunID              string
	Generation         int64
	UnitID             string
	Attempt            int64
	StateVersion       int64
	ContractHash       string
	EvidenceHash       string
	ValidatorSessionID string
	HeadSHA            string
	Validator          string
	Thinking           string
	Verdict            string
	LocalGates         bool
	UAT                bool
	MilestoneValid     bool
	CreatedAt          time.Time
	ExpiresAt          time.Time
}

func Open(ctx context.Context, path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create supervisor directory: %w", err)
	}
	if err := os.Chmod(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("secure supervisor directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	store := &Store{db: db}
	if err := store.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("secure supervisor database: %w", err)
	}
	return store, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate(ctx context.Context) error {
	statements := []string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA foreign_keys=ON`,
		`PRAGMA synchronous=FULL`,
		`PRAGMA busy_timeout=5000`,
		`CREATE TABLE IF NOT EXISTS authority_state (
            state_id TEXT PRIMARY KEY, version INTEGER NOT NULL, updated_at INTEGER NOT NULL
        )`,
		`INSERT OR IGNORE INTO authority_state(state_id, version, updated_at) VALUES ('governance', 1, strftime('%s','now') * 1000000000)`,
		`CREATE TABLE IF NOT EXISTS leases (
            run_id TEXT PRIMARY KEY, owner TEXT NOT NULL, epoch INTEGER NOT NULL,
            expires_at INTEGER NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS deliveries (
            delivery_id TEXT PRIMARY KEY, issue INTEGER NOT NULL, work_dir TEXT NOT NULL,
			context_hash TEXT NOT NULL, milestone_id TEXT NOT NULL DEFAULT '',
			parent_issue INTEGER NOT NULL DEFAULT 0, repository TEXT NOT NULL DEFAULT '',
			pull_request INTEGER NOT NULL DEFAULT 0, branch TEXT NOT NULL DEFAULT '',
			base_branch TEXT NOT NULL DEFAULT '', gsd_project_root TEXT NOT NULL DEFAULT '',
			initial_head TEXT NOT NULL DEFAULT '', gsd_version TEXT NOT NULL DEFAULT '',
            created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS delivery_runs (
            delivery_id TEXT PRIMARY KEY REFERENCES deliveries(delivery_id),
            state TEXT NOT NULL, generation INTEGER NOT NULL, attempt INTEGER NOT NULL,
			owner TEXT NOT NULL DEFAULT '', start_head TEXT NOT NULL DEFAULT '', end_head TEXT NOT NULL DEFAULT '',
			updated_at INTEGER NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS human_decisions (
            delivery_id TEXT NOT NULL, generation INTEGER NOT NULL, approved INTEGER NOT NULL,
            consumed_at INTEGER NOT NULL, PRIMARY KEY (delivery_id, generation)
        )`,
		`CREATE TABLE IF NOT EXISTS attestations (
            run_id TEXT NOT NULL, head_sha TEXT NOT NULL, validator TEXT NOT NULL,
            thinking TEXT NOT NULL, verdict TEXT NOT NULL, created_at INTEGER NOT NULL,
            repository TEXT NOT NULL DEFAULT '', pr INTEGER NOT NULL DEFAULT 0, base_branch TEXT NOT NULL DEFAULT '',
            base_head TEXT NOT NULL DEFAULT '', candidate_head TEXT NOT NULL DEFAULT '', observed_head TEXT NOT NULL DEFAULT '',
            generation INTEGER NOT NULL DEFAULT 0, unit_id TEXT NOT NULL DEFAULT '', attempt INTEGER NOT NULL DEFAULT 0,
            state_version INTEGER NOT NULL DEFAULT 0, contract_hash TEXT NOT NULL DEFAULT '', evidence_hash TEXT NOT NULL DEFAULT '',
            validator_session_id TEXT NOT NULL DEFAULT '', local_gates INTEGER NOT NULL DEFAULT 0,
            uat INTEGER NOT NULL DEFAULT 0, milestone_valid INTEGER NOT NULL DEFAULT 0, expires_at INTEGER NOT NULL DEFAULT 0,
            PRIMARY KEY (run_id, head_sha)
        )`,
		`CREATE TABLE IF NOT EXISTS unit_attempts (
			delivery_id TEXT NOT NULL REFERENCES deliveries(delivery_id), generation INTEGER NOT NULL,
			unit_id TEXT NOT NULL, head_sha TEXT NOT NULL, attempts INTEGER NOT NULL,
			max_attempts INTEGER NOT NULL, status TEXT NOT NULL, last_failure TEXT NOT NULL DEFAULT '',
			updated_at INTEGER NOT NULL,
			PRIMARY KEY (delivery_id, generation, unit_id, head_sha)
		)`,
		`CREATE TABLE IF NOT EXISTS unit_attempt_identities (
			delivery_id TEXT NOT NULL REFERENCES deliveries(delivery_id), generation INTEGER NOT NULL,
			unit_id TEXT NOT NULL, head_sha TEXT NOT NULL, attempt INTEGER NOT NULL,
			model TEXT NOT NULL, thinking TEXT NOT NULL, session_id TEXT NOT NULL,
			session_fingerprint TEXT NOT NULL, started_at INTEGER NOT NULL, ended_at INTEGER NOT NULL,
			PRIMARY KEY (delivery_id, generation, unit_id, head_sha, attempt)
		)`,
		`CREATE TABLE IF NOT EXISTS decision_requests (
			request_id TEXT PRIMARY KEY, delivery_id TEXT NOT NULL REFERENCES deliveries(delivery_id),
			repository TEXT NOT NULL, issue INTEGER NOT NULL, pull_request INTEGER NOT NULL, unit_id TEXT NOT NULL,
			generation INTEGER NOT NULL, head_sha TEXT NOT NULL, kind TEXT NOT NULL,
			evidence TEXT NOT NULL, options_json TEXT NOT NULL, recommended_option TEXT NOT NULL DEFAULT '',
			safe_default TEXT NOT NULL DEFAULT '', expires_at INTEGER NOT NULL,
			github_comment_id INTEGER NOT NULL DEFAULT 0, status TEXT NOT NULL,
			accepted_answer TEXT NOT NULL DEFAULT '', accepted_by TEXT NOT NULL DEFAULT '',
			accepted_actor_id INTEGER NOT NULL DEFAULT 0, accepted_comment_id INTEGER NOT NULL DEFAULT 0,
			accepted_at INTEGER NOT NULL DEFAULT 0,
			consumed_at INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS recovery_budgets (
			delivery_id TEXT NOT NULL REFERENCES deliveries(delivery_id), generation INTEGER NOT NULL,
			unit_id TEXT NOT NULL, head_sha TEXT NOT NULL, failure_class TEXT NOT NULL,
			attempts INTEGER NOT NULL, max_attempts INTEGER NOT NULL, backoff_ms INTEGER NOT NULL,
			last_failure TEXT NOT NULL DEFAULT '', recovery_plan TEXT NOT NULL DEFAULT '',
			next_retry_at INTEGER NOT NULL DEFAULT 0, exhausted_at INTEGER NOT NULL DEFAULT 0,
			policy_version INTEGER NOT NULL DEFAULT 0, base_backoff_ms INTEGER NOT NULL DEFAULT 0,
			max_backoff_ms INTEGER NOT NULL DEFAULT 0, last_failure_hash TEXT NOT NULL DEFAULT '',
			last_diagnostic TEXT NOT NULL DEFAULT '', planner_evidence_hash TEXT NOT NULL DEFAULT '',
			planner_session_id TEXT NOT NULL DEFAULT '', planner_session_fingerprint TEXT NOT NULL DEFAULT '',
			observed_model TEXT NOT NULL DEFAULT '', thinking TEXT NOT NULL DEFAULT '',
			selected_action TEXT NOT NULL DEFAULT '', bounded_plan_json TEXT NOT NULL DEFAULT '[]',
			status TEXT NOT NULL DEFAULT '', updated_at INTEGER NOT NULL,
			PRIMARY KEY (delivery_id, generation, unit_id, head_sha, failure_class)
		)`,
		`CREATE TABLE IF NOT EXISTS recovery_attempts (
			delivery_id TEXT NOT NULL REFERENCES deliveries(delivery_id), generation INTEGER NOT NULL,
			unit_id TEXT NOT NULL, head_sha TEXT NOT NULL, failure_class TEXT NOT NULL,
			unit_attempt INTEGER NOT NULL, claim_token TEXT NOT NULL, controller_owner TEXT NOT NULL,
			controller_epoch INTEGER NOT NULL, reservation_index INTEGER NOT NULL, sequence INTEGER NOT NULL, failure_hash TEXT NOT NULL,
			diagnostic TEXT NOT NULL, reversible INTEGER NOT NULL, status TEXT NOT NULL,
			planner_request_nonce TEXT NOT NULL DEFAULT '', evidence_hash TEXT NOT NULL DEFAULT '',
			authority_scope_hash TEXT NOT NULL DEFAULT '', planner_evidence_hash TEXT NOT NULL DEFAULT '',
			planner_session_id TEXT NOT NULL DEFAULT '', planner_session_fingerprint TEXT NOT NULL DEFAULT '',
			observed_model TEXT NOT NULL DEFAULT '', thinking TEXT NOT NULL DEFAULT '',
			selected_action TEXT NOT NULL DEFAULT '', bounded_plan_json TEXT NOT NULL DEFAULT '[]',
			backoff_ms INTEGER NOT NULL, next_retry_at INTEGER NOT NULL,
			issued_at INTEGER NOT NULL DEFAULT 0, expires_at INTEGER NOT NULL DEFAULT 0,
			rejected_reason TEXT NOT NULL DEFAULT '', dispatch_claim TEXT NOT NULL DEFAULT '',
			dispatch_epoch INTEGER NOT NULL DEFAULT 0, dispatched_at INTEGER NOT NULL DEFAULT 0, created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL,
			PRIMARY KEY (delivery_id, generation, unit_id, head_sha, failure_class, unit_attempt)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS recovery_attempts_nonce_unique
			ON recovery_attempts(planner_request_nonce) WHERE planner_request_nonce <> ''`,
		`CREATE UNIQUE INDEX IF NOT EXISTS recovery_attempts_evidence_unique
			ON recovery_attempts(planner_evidence_hash) WHERE planner_evidence_hash <> ''`,
		`CREATE TABLE IF NOT EXISTS artifact_proofs (
			proof_id TEXT PRIMARY KEY, delivery_id TEXT NOT NULL REFERENCES deliveries(delivery_id),
			generation INTEGER NOT NULL, unit_id TEXT NOT NULL, attempt INTEGER NOT NULL,
			start_head TEXT NOT NULL, candidate_head TEXT NOT NULL, validated_head TEXT NOT NULL,
			expected_artifact TEXT NOT NULL, artifact_hash TEXT NOT NULL, validator TEXT NOT NULL,
			thinking TEXT NOT NULL, ratified INTEGER NOT NULL, created_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS attempt_worktrees (
			delivery_id TEXT NOT NULL REFERENCES deliveries(delivery_id), generation INTEGER NOT NULL,
			unit_id TEXT NOT NULL, attempt INTEGER NOT NULL, branch TEXT NOT NULL UNIQUE,
			path TEXT NOT NULL UNIQUE, base_head TEXT NOT NULL, candidate_head TEXT NOT NULL DEFAULT '',
			validated_head TEXT NOT NULL DEFAULT '', state TEXT NOT NULL,
			controller_owner TEXT NOT NULL, controller_epoch INTEGER NOT NULL,
			resources_created INTEGER NOT NULL DEFAULT 0,
			failure_class TEXT NOT NULL DEFAULT '', cleanup_error TEXT NOT NULL DEFAULT '',
			created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL,
			PRIMARY KEY (delivery_id, generation, unit_id, attempt)
		)`,
		`CREATE TABLE IF NOT EXISTS promotion_journals (
			journal_id TEXT PRIMARY KEY, delivery_id TEXT NOT NULL, generation INTEGER NOT NULL,
			unit_id TEXT NOT NULL, attempt INTEGER NOT NULL, base_head TEXT NOT NULL,
			candidate_head TEXT NOT NULL, validated_head TEXT NOT NULL, proof_id TEXT NOT NULL,
			evidence_hash TEXT NOT NULL, validator_session_id TEXT NOT NULL,
			attestation_repository TEXT NOT NULL, attestation_pr INTEGER NOT NULL,
			attestation_base_branch TEXT NOT NULL, attestation_contract_hash TEXT NOT NULL,
			attestation_created_at INTEGER NOT NULL, attestation_validator TEXT NOT NULL, attestation_thinking TEXT NOT NULL,
			attestation_verdict TEXT NOT NULL, attestation_local_gates INTEGER NOT NULL,
			attestation_uat INTEGER NOT NULL, attestation_milestone_valid INTEGER NOT NULL,
			governance_state_version INTEGER NOT NULL, attestation_expires_at INTEGER NOT NULL,
			manifest_json TEXT NOT NULL, manifest_hash TEXT NOT NULL,
			backup_manifest_json TEXT NOT NULL, backup_manifest_hash TEXT NOT NULL,
			stage_path TEXT NOT NULL UNIQUE, backup_path TEXT NOT NULL UNIQUE, canonical_path TEXT NOT NULL,
			state TEXT NOT NULL, blocked_reason TEXT NOT NULL DEFAULT '', cleanup_complete INTEGER NOT NULL DEFAULT 0,
			decisions_resolved INTEGER NOT NULL DEFAULT 0, controller_owner TEXT NOT NULL, controller_epoch INTEGER NOT NULL, created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL,
			FOREIGN KEY (delivery_id, generation, unit_id, attempt)
				REFERENCES attempt_worktrees(delivery_id, generation, unit_id, attempt)
		)`,
		`CREATE INDEX IF NOT EXISTS promotion_journals_delivery_state
			ON promotion_journals(delivery_id, state, created_at)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply supervisor migration: %w", err)
		}
	}
	var legacyOutboxTables int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'table' AND name = 'outbox'`).Scan(&legacyOutboxTables); err != nil {
		return fmt.Errorf("inspect legacy outbox schema: %w", err)
	}
	if legacyOutboxTables == 1 {
		if _, err := s.db.ExecContext(ctx, `UPDATE outbox SET status = 'blocked' WHERE status <> 'sent'`); err != nil {
			return fmt.Errorf("quarantine legacy outbox effects: %w", err)
		}
	}
	columns := map[string]string{
		"parent_issue": "INTEGER NOT NULL DEFAULT 0", "repository": "TEXT NOT NULL DEFAULT ''",
		"pull_request": "INTEGER NOT NULL DEFAULT 0", "branch": "TEXT NOT NULL DEFAULT ''",
		"base_branch": "TEXT NOT NULL DEFAULT ''", "gsd_project_root": "TEXT NOT NULL DEFAULT ''",
		"initial_head": "TEXT NOT NULL DEFAULT ''", "gsd_version": "TEXT NOT NULL DEFAULT ''",
	}
	for name, definition := range columns {
		if err := s.ensureColumn(ctx, "deliveries", name, definition); err != nil {
			return err
		}
	}
	if err := s.ensureColumn(ctx, "decision_requests", "repository", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := s.ensureColumn(ctx, "decision_requests", "accepted_comment_id", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := s.ensureColumn(ctx, "decision_requests", "accepted_actor_id", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := s.ensureColumn(ctx, "promotion_journals", "decisions_resolved", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	for name, definition := range map[string]string{"start_head": "TEXT NOT NULL DEFAULT ''", "end_head": "TEXT NOT NULL DEFAULT ''"} {
		if err := s.ensureColumn(ctx, "delivery_runs", name, definition); err != nil {
			return err
		}
	}
	for name, definition := range map[string]string{
		"policy_version": "INTEGER NOT NULL DEFAULT 0", "base_backoff_ms": "INTEGER NOT NULL DEFAULT 0",
		"max_backoff_ms": "INTEGER NOT NULL DEFAULT 0", "last_failure_hash": "TEXT NOT NULL DEFAULT ''",
		"last_diagnostic": "TEXT NOT NULL DEFAULT ''", "planner_evidence_hash": "TEXT NOT NULL DEFAULT ''",
		"planner_session_id": "TEXT NOT NULL DEFAULT ''", "planner_session_fingerprint": "TEXT NOT NULL DEFAULT ''",
		"observed_model": "TEXT NOT NULL DEFAULT ''", "thinking": "TEXT NOT NULL DEFAULT ''",
		"selected_action": "TEXT NOT NULL DEFAULT ''", "bounded_plan_json": "TEXT NOT NULL DEFAULT '[]'",
		"status": "TEXT NOT NULL DEFAULT ''",
	} {
		if err := s.ensureColumn(ctx, "recovery_budgets", name, definition); err != nil {
			return err
		}
	}
	legacyRecoveryStamp := time.Now().UTC().UnixNano()
	if _, err := s.db.ExecContext(ctx, `UPDATE recovery_budgets SET policy_version = ?, max_attempts = attempts,
		selected_action = ?, status = 'exhausted', exhausted_at = CASE WHEN exhausted_at = 0 THEN ? ELSE exhausted_at END,
		next_retry_at = 0, bounded_plan_json = '[]', updated_at = ? WHERE policy_version = 0`,
		recovery.PolicyVersion, recovery.ActionAwaitDecision, legacyRecoveryStamp, legacyRecoveryStamp); err != nil {
		return fmt.Errorf("exhaust legacy recovery budgets: %w", err)
	}
	for name, definition := range map[string]string{
		"repository": "TEXT NOT NULL DEFAULT ''", "pr": "INTEGER NOT NULL DEFAULT 0", "base_branch": "TEXT NOT NULL DEFAULT ''",
		"base_head": "TEXT NOT NULL DEFAULT ''", "candidate_head": "TEXT NOT NULL DEFAULT ''", "observed_head": "TEXT NOT NULL DEFAULT ''",
		"generation": "INTEGER NOT NULL DEFAULT 0", "unit_id": "TEXT NOT NULL DEFAULT ''", "attempt": "INTEGER NOT NULL DEFAULT 0",
		"state_version": "INTEGER NOT NULL DEFAULT 0", "contract_hash": "TEXT NOT NULL DEFAULT ''", "evidence_hash": "TEXT NOT NULL DEFAULT ''",
		"validator_session_id": "TEXT NOT NULL DEFAULT ''", "local_gates": "INTEGER NOT NULL DEFAULT 0", "uat": "INTEGER NOT NULL DEFAULT 0",
		"milestone_valid": "INTEGER NOT NULL DEFAULT 0", "expires_at": "INTEGER NOT NULL DEFAULT 0",
	} {
		if err := s.ensureColumn(ctx, "attestations", name, definition); err != nil {
			return err
		}
	}
	if err := s.verifyDurability(ctx); err != nil {
		return err
	}
	for _, statement := range []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS deliveries_issue_unique ON deliveries(issue)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS deliveries_project_root_unique ON deliveries(gsd_project_root) WHERE gsd_project_root <> ''`,
	} {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply supervisor identity index: %w", err)
		}
	}
	return nil
}

func (s *Store) GovernanceStateVersion(ctx context.Context) (int64, error) {
	var version int64
	if err := s.db.QueryRowContext(ctx, `SELECT version FROM authority_state WHERE state_id = 'governance'`).Scan(&version); err != nil {
		return 0, fmt.Errorf("read governance state version: %w", err)
	}
	if version <= 0 {
		return 0, errors.New("governance state version is invalid")
	}
	return version, nil
}

func (s *Store) BeginUnitAttempt(ctx context.Context, key UnitAttemptKey, maxAttempts int64) (UnitAttempt, error) {
	return s.beginUnitAttempt(ctx, key, maxAttempts, "", 0)
}

func (s *Store) BeginUnitAttemptFenced(ctx context.Context, key UnitAttemptKey, maxAttempts int64, controllerOwner string, controllerEpoch int64) (UnitAttempt, error) {
	if strings.TrimSpace(controllerOwner) == "" || strings.ContainsAny(controllerOwner, "\r\n\x00") || controllerEpoch <= 0 {
		return UnitAttempt{}, errors.New("controller owner and lease epoch are required")
	}
	return s.beginUnitAttempt(ctx, key, maxAttempts, controllerOwner, controllerEpoch)
}

func (s *Store) beginUnitAttempt(ctx context.Context, key UnitAttemptKey, maxAttempts int64, controllerOwner string, controllerEpoch int64) (UnitAttempt, error) {
	if key.DeliveryID == "" || key.Generation <= 0 || strings.TrimSpace(key.UnitID) == "" ||
		strings.ContainsAny(key.UnitID, "\r\n\x00") || !validGitSHA(key.HeadSHA) || maxAttempts <= 0 {
		return UnitAttempt{}, errors.New("complete bounded unit attempt identity is required")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return UnitAttempt{}, err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return UnitAttempt{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	now := time.Now().UTC().UnixNano()
	if controllerOwner != "" {
		var fenced int
		if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM delivery_runs WHERE delivery_id = ?
			AND generation = ? AND state = ? AND owner = ? AND EXISTS (SELECT 1 FROM leases WHERE run_id = ?
			AND owner = ? AND epoch = ? AND expires_at > ?)`, key.DeliveryID, key.Generation,
			domain.RunRunning, controllerOwner, key.DeliveryID, controllerOwner, controllerEpoch, now).Scan(&fenced); err != nil || fenced != 1 {
			return UnitAttempt{}, ErrRecoveryClaimFenced
		}
	}
	if _, err := conn.ExecContext(ctx, `INSERT OR IGNORE INTO unit_attempts
		(delivery_id, generation, unit_id, head_sha, attempts, max_attempts, status, updated_at)
		VALUES (?, ?, ?, ?, 0, ?, 'pending', ?)`, key.DeliveryID, key.Generation, key.UnitID,
		key.HeadSHA, maxAttempts, now); err != nil {
		return UnitAttempt{}, fmt.Errorf("initialize unit attempt: %w", err)
	}
	var attempt UnitAttempt
	var configuredMax int64
	attempt.UnitAttemptKey = key
	if err := conn.QueryRowContext(ctx, `SELECT attempts, max_attempts, status, last_failure
		FROM unit_attempts WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ?`,
		key.DeliveryID, key.Generation, key.UnitID, key.HeadSHA).Scan(&attempt.Attempts,
		&configuredMax, &attempt.Status, &attempt.Failure); err != nil {
		return UnitAttempt{}, err
	}
	if configuredMax > maxAttempts {
		return UnitAttempt{}, errors.New("unit attempt budget cannot shrink within a generation")
	}
	if configuredMax < maxAttempts {
		if _, err := conn.ExecContext(ctx, `UPDATE unit_attempts SET max_attempts = ?, updated_at = ?
			WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ? AND max_attempts = ?`,
			maxAttempts, now, key.DeliveryID, key.Generation, key.UnitID, key.HeadSHA, configuredMax); err != nil {
			return UnitAttempt{}, err
		}
		configuredMax = maxAttempts
	}
	if attempt.Attempts >= configuredMax {
		return UnitAttempt{}, ErrRetryBudgetExhausted
	}
	attempt.Attempts++
	attempt.Remaining = configuredMax - attempt.Attempts
	attempt.Status, attempt.Failure = "running", ""
	if _, err := conn.ExecContext(ctx, `UPDATE unit_attempts SET attempts = ?, status = ?,
		last_failure = '', updated_at = ? WHERE delivery_id = ? AND generation = ? AND unit_id = ?
		AND head_sha = ?`, attempt.Attempts, attempt.Status, now, key.DeliveryID, key.Generation,
		key.UnitID, key.HeadSHA); err != nil {
		return UnitAttempt{}, err
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return UnitAttempt{}, err
	}
	committed = true
	return attempt, nil
}

func (s *Store) FinishUnitAttemptFenced(ctx context.Context, key UnitAttemptKey, attempt int64, outcome, controllerOwner string, controllerEpoch int64) error {
	if attempt <= 0 || strings.TrimSpace(controllerOwner) == "" || strings.ContainsAny(controllerOwner, "\r\n\x00") || controllerEpoch <= 0 {
		return errors.New("unit attempt, controller owner, and lease epoch are required")
	}
	return s.finishUnitAttempt(ctx, key, attempt, outcome, controllerOwner, controllerEpoch)
}

func (s *Store) finishUnitAttempt(ctx context.Context, key UnitAttemptKey, attempt int64, outcome, controllerOwner string, controllerEpoch int64) error {
	outcome = strings.TrimSpace(outcome)
	if outcome == "" || strings.ContainsAny(outcome, "\r\n\x00") {
		return errors.New("bounded unit attempt outcome is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	stamp := time.Now().UTC().UnixNano()
	var result sql.Result
	if controllerOwner == "" {
		result, err = tx.ExecContext(ctx, `UPDATE unit_attempts SET status = 'terminal',
			last_failure = ?, updated_at = ? WHERE delivery_id = ? AND generation = ? AND unit_id = ?
			AND head_sha = ? AND status = 'running'`, outcome, stamp, key.DeliveryID, key.Generation,
			key.UnitID, key.HeadSHA)
	} else {
		result, err = tx.ExecContext(ctx, `UPDATE unit_attempts SET status = 'terminal',
			last_failure = ?, updated_at = ? WHERE delivery_id = ? AND generation = ? AND unit_id = ?
			AND head_sha = ? AND attempts = ? AND status = 'running' AND EXISTS (SELECT 1 FROM delivery_runs
			WHERE delivery_id = ? AND generation = ? AND state = ? AND owner = ?) AND EXISTS
			(SELECT 1 FROM leases WHERE run_id = ? AND owner = ? AND epoch = ? AND expires_at > ?)`, outcome,
			stamp, key.DeliveryID, key.Generation, key.UnitID, key.HeadSHA, attempt, key.DeliveryID,
			key.Generation, domain.RunRunning, controllerOwner, key.DeliveryID, controllerOwner,
			controllerEpoch, stamp)
	}
	if err != nil {
		return fmt.Errorf("finish unit attempt: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return errors.New("unit attempt is not running or is fenced")
	}
	if controllerOwner != "" {
		dispatchStatus := "used_failed"
		if outcome == "success" || outcome == "mutating_skip" {
			dispatchStatus = "consumed"
		}
		var dispatched int64
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM recovery_attempts WHERE delivery_id = ?
			AND generation = ? AND unit_id = ? AND head_sha = ? AND status = 'dispatched'`, key.DeliveryID,
			key.Generation, key.UnitID, key.HeadSHA).Scan(&dispatched); err != nil {
			return err
		}
		updated, err := tx.ExecContext(ctx, `UPDATE recovery_attempts SET status = ?, updated_at = ?
			WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ? AND status = 'dispatched'
			AND dispatch_claim = ? AND dispatch_epoch = ? AND EXISTS (SELECT 1 FROM delivery_runs WHERE delivery_id = ?
			AND generation = ? AND state = ? AND owner = ?) AND EXISTS (SELECT 1 FROM leases WHERE run_id = ?
			AND owner = ? AND epoch = ? AND expires_at > ?)`, dispatchStatus, stamp, key.DeliveryID,
			key.Generation, key.UnitID, key.HeadSHA, controllerOwner, controllerEpoch, key.DeliveryID, key.Generation,
			domain.RunRunning, controllerOwner, key.DeliveryID, controllerOwner, controllerEpoch, stamp)
		if err != nil {
			return err
		}
		rows, _ := updated.RowsAffected()
		if rows != dispatched {
			return errors.New("recovery dispatch disposition is fenced")
		}
	}
	return tx.Commit()
}

func (s *Store) RecordUnitAttemptIdentity(ctx context.Context, identity UnitAttemptIdentity) error {
	if identity.DeliveryID == "" || identity.Generation <= 0 || identity.UnitID == "" || !validGitSHA(identity.HeadSHA) || identity.Attempt <= 0 ||
		(identity.Model != "openai-codex/gpt-5.6-sol" && identity.Model != "openai-codex/gpt-5.5") || identity.Thinking != "high" ||
		len(identity.SessionID) != 36 || strings.ContainsAny(identity.SessionID, "\r\n\x00") || !validSHA256(identity.SessionFingerprint) ||
		identity.StartedAt.IsZero() || identity.EndedAt.Before(identity.StartedAt) {
		return errors.New("complete bounded unit attempt identity is required")
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO unit_attempt_identities
		(delivery_id, generation, unit_id, head_sha, attempt, model, thinking, session_id,
		 session_fingerprint, started_at, ended_at)
		SELECT delivery_id, generation, unit_id, head_sha, ?, ?, ?, ?, ?, ?, ? FROM unit_attempts
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ?
		AND attempts = ? AND status = 'running'`, identity.Attempt, identity.Model, identity.Thinking,
		identity.SessionID, identity.SessionFingerprint, identity.StartedAt.UTC().UnixNano(), identity.EndedAt.UTC().UnixNano(),
		identity.DeliveryID, identity.Generation, identity.UnitID, identity.HeadSHA, identity.Attempt)
	if err != nil {
		return fmt.Errorf("record unit attempt identity: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return errors.New("unit attempt identity is not bound to the running attempt")
	}
	return nil
}

func (s *Store) GetUnitAttemptIdentity(ctx context.Context, key UnitAttemptKey, attempt int64) (UnitAttemptIdentity, error) {
	identity := UnitAttemptIdentity{UnitAttemptKey: key, Attempt: attempt}
	var startedAt, endedAt int64
	err := s.db.QueryRowContext(ctx, `SELECT model, thinking, session_id, session_fingerprint, started_at, ended_at
		FROM unit_attempt_identities WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ? AND attempt = ?`,
		key.DeliveryID, key.Generation, key.UnitID, key.HeadSHA, attempt).Scan(&identity.Model, &identity.Thinking,
		&identity.SessionID, &identity.SessionFingerprint, &startedAt, &endedAt)
	if err != nil {
		return UnitAttemptIdentity{}, err
	}
	identity.StartedAt = time.Unix(0, startedAt).UTC()
	identity.EndedAt = time.Unix(0, endedAt).UTC()
	return identity, nil
}

func (s *Store) verifyDurability(ctx context.Context) error {
	var journalMode string
	if err := s.db.QueryRowContext(ctx, `PRAGMA journal_mode`).Scan(&journalMode); err != nil {
		return fmt.Errorf("verify supervisor journal mode: %w", err)
	}
	if !strings.EqualFold(journalMode, "wal") {
		return fmt.Errorf("supervisor database did not enter WAL mode: %s", journalMode)
	}
	return nil
}

func (s *Store) ensureColumn(ctx context.Context, table, column, definition string) error {
	rows, err := s.db.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return fmt.Errorf("inspect supervisor schema: %w", err)
	}
	found := false
	for rows.Next() {
		var position int
		var name, dataType string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&position, &name, &dataType, &notNull, &defaultValue, &primaryKey); err != nil {
			_ = rows.Close()
			return fmt.Errorf("read supervisor schema: %w", err)
		}
		if name == column {
			found = true
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if found {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE `+table+` ADD COLUMN `+column+` `+definition); err != nil {
		return fmt.Errorf("add supervisor schema column %s: %w", column, err)
	}
	return nil
}

func (s *Store) RecordAttemptHeads(ctx context.Context, deliveryID, owner, startHead, endHead string) error {
	if deliveryID == "" || owner == "" || len(startHead) != 40 || len(endHead) != 40 {
		return errors.New("delivery owner and exact attempt heads are required")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE delivery_runs SET start_head = ?, end_head = ?, updated_at = ?
        WHERE delivery_id = ? AND state = ? AND owner = ?`, startHead, endHead, time.Now().UTC().UnixNano(),
		deliveryID, domain.RunRunning, owner)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return errors.New("cannot bind heads to a stale delivery attempt")
	}
	return nil
}

func (s *Store) EnsureDelivery(ctx context.Context, delivery Delivery) error {
	if err := validateDelivery(delivery); err != nil {
		return err
	}
	now := time.Now().UTC().UnixNano()
	result, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO deliveries
		(delivery_id, issue, parent_issue, repository, pull_request, work_dir, context_hash, milestone_id,
		 branch, base_branch, gsd_project_root, initial_head, gsd_version, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, delivery.ID, delivery.Issue,
		delivery.ParentIssue, delivery.Repository, delivery.PullRequest, delivery.WorkDir, delivery.ContextHash,
		delivery.MilestoneID, delivery.Branch, delivery.BaseBranch, delivery.GSDProjectRoot,
		delivery.InitialHead, delivery.GSDVersion, now, now)
	if err != nil {
		return fmt.Errorf("ensure delivery: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows == 1 {
		_, err := s.db.ExecContext(ctx, `INSERT INTO delivery_runs
            (delivery_id, state, generation, attempt, owner, updated_at) VALUES (?, ?, 1, 0, '', ?)`,
			delivery.ID, domain.RunPlanned, now)
		if err != nil {
			return fmt.Errorf("initialize delivery run: %w", err)
		}
		return nil
	}
	// Existing pre-identity controller databases are upgraded once, but only
	// when every newly introduced identity column is still empty.
	if _, err := s.db.ExecContext(ctx, `UPDATE deliveries SET parent_issue = ?, branch = ?, base_branch = ?,
		gsd_project_root = ?, initial_head = ?, gsd_version = ?, updated_at = ?
		WHERE delivery_id = ? AND issue = ? AND work_dir = ? AND context_hash = ?
		AND parent_issue = 0 AND branch = '' AND base_branch = '' AND gsd_project_root = ''
		AND initial_head = '' AND gsd_version = ''`, delivery.ParentIssue, delivery.Branch,
		delivery.BaseBranch, delivery.GSDProjectRoot, delivery.InitialHead, delivery.GSDVersion, now,
		delivery.ID, delivery.Issue, delivery.WorkDir, delivery.ContextHash); err != nil {
		return fmt.Errorf("upgrade delivery identity: %w", err)
	}
	current, err := s.GetDelivery(ctx, delivery.ID)
	if err != nil {
		return fmt.Errorf("read delivery: %w", err)
	}
	if current.Repository == "" || current.PullRequest == 0 {
		return ErrLegacyDeliveryTarget
	}
	if current.Issue != delivery.Issue || current.ParentIssue != delivery.ParentIssue ||
		current.Repository != delivery.Repository || current.PullRequest != delivery.PullRequest ||
		current.WorkDir != delivery.WorkDir || current.ContextHash != delivery.ContextHash ||
		current.Branch != delivery.Branch || current.BaseBranch != delivery.BaseBranch ||
		current.GSDProjectRoot != delivery.GSDProjectRoot || current.InitialHead != delivery.InitialHead ||
		current.GSDVersion != delivery.GSDVersion ||
		(delivery.MilestoneID != "" && current.MilestoneID != "" && current.MilestoneID != delivery.MilestoneID) {
		return errors.New("delivery identity is already bound to different canonical inputs")
	}
	return nil
}

func validateDelivery(delivery Delivery) error {
	if delivery.ID == "" || delivery.Issue <= 0 || delivery.ParentIssue <= 0 ||
		!validRepository(delivery.Repository) || delivery.PullRequest <= 0 || !filepath.IsAbs(delivery.WorkDir) || !filepath.IsAbs(delivery.GSDProjectRoot) ||
		filepath.Clean(delivery.WorkDir) != filepath.Clean(delivery.GSDProjectRoot) ||
		!validSHA256(delivery.ContextHash) || strings.TrimSpace(delivery.Branch) == "" ||
		strings.TrimSpace(delivery.BaseBranch) == "" || !validGitSHA(delivery.InitialHead) ||
		strings.TrimSpace(delivery.GSDVersion) == "" {
		return errors.New("complete canonical issue, branch, GSD project, head, version, and context identity are required")
	}
	return nil
}

func (s *Store) BeginAttempt(ctx context.Context, deliveryID, owner string) (DeliveryRun, error) {
	if deliveryID == "" || owner == "" {
		return DeliveryRun{}, errors.New("delivery and owner are required")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return DeliveryRun{}, err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return DeliveryRun{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var run DeliveryRun
	err = conn.QueryRowContext(ctx, `SELECT delivery_id, state, generation, attempt, owner
        FROM delivery_runs WHERE delivery_id = ?`, deliveryID).Scan(&run.DeliveryID, &run.State,
		&run.Generation, &run.Attempt, &run.Owner)
	if err != nil {
		return DeliveryRun{}, fmt.Errorf("read delivery run: %w", err)
	}
	if err := domain.ValidateRunTransition(run.State, domain.RunRunning); err != nil {
		return DeliveryRun{}, err
	}
	run.State, run.Attempt, run.Owner = domain.RunRunning, run.Attempt+1, owner
	if _, err := conn.ExecContext(ctx, `UPDATE delivery_runs SET state = ?, attempt = ?, owner = ?, updated_at = ?
        WHERE delivery_id = ?`, run.State, run.Attempt, owner, time.Now().UTC().UnixNano(), deliveryID); err != nil {
		return DeliveryRun{}, err
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return DeliveryRun{}, err
	}
	committed = true
	return run, nil
}

// RetryFailedIntake resets only an unbound delivery whose milestone intake failed. It never resets
// a blocked delivery or a delivery already bound to canonical GSD state.
func (s *Store) RetryFailedIntake(ctx context.Context, deliveryID string) error {
	if deliveryID == "" {
		return errors.New("delivery is required")
	}
	now := time.Now().UTC().UnixNano()
	result, err := s.db.ExecContext(ctx, `UPDATE delivery_runs
        SET state = ?, generation = generation + 1, owner = '', updated_at = ?
        WHERE delivery_id = ? AND state = ? AND EXISTS (
            SELECT 1 FROM deliveries WHERE delivery_id = ? AND milestone_id = ''
        )`, domain.RunPlanned, now, deliveryID, domain.RunFailed, deliveryID)
	if err != nil {
		return fmt.Errorf("retry failed intake: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows == 1 {
		return nil
	}
	var state domain.RunState
	var milestoneID string
	if err := s.db.QueryRowContext(ctx, `SELECT r.state, d.milestone_id
        FROM delivery_runs r JOIN deliveries d ON d.delivery_id = r.delivery_id
        WHERE r.delivery_id = ?`, deliveryID).Scan(&state, &milestoneID); err != nil {
		return fmt.Errorf("read intake retry state: %w", err)
	}
	if state == domain.RunPlanned && milestoneID == "" {
		return nil
	}
	return errors.New("only failed pre-milestone intake may be retried")
}

// PrepareAdoptedDelivery makes a validated, canonically bound milestone runnable. Adoption may
// recover a failed controller attempt because the caller has independently queried and verified
// the existing GSD milestone. Blocked and human-gated deliveries still require explicit resume.
func (s *Store) PrepareAdoptedDelivery(ctx context.Context, deliveryID, milestoneID string) error {
	if deliveryID == "" || milestoneID == "" {
		return errors.New("delivery and milestone are required")
	}
	now := time.Now().UTC().UnixNano()
	result, err := s.db.ExecContext(ctx, `UPDATE delivery_runs
        SET state = ?, generation = CASE WHEN state = ? THEN generation + 1 ELSE generation END,
            owner = '', updated_at = ?
        WHERE delivery_id = ? AND state IN (?, ?) AND EXISTS (
            SELECT 1 FROM deliveries WHERE delivery_id = ? AND milestone_id = ?
        )`, domain.RunReady, domain.RunFailed, now, deliveryID, domain.RunPlanned, domain.RunFailed,
		deliveryID, milestoneID)
	if err != nil {
		return fmt.Errorf("prepare adopted delivery: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows == 1 {
		return nil
	}
	var state domain.RunState
	var bound string
	if err := s.db.QueryRowContext(ctx, `SELECT r.state, d.milestone_id
        FROM delivery_runs r JOIN deliveries d ON d.delivery_id = r.delivery_id
        WHERE r.delivery_id = ?`, deliveryID).Scan(&state, &bound); err != nil {
		return fmt.Errorf("read adopted delivery: %w", err)
	}
	if state == domain.RunReady && bound == milestoneID {
		return nil
	}
	return fmt.Errorf("delivery in state %s cannot adopt milestone %s", state, milestoneID)
}

func (s *Store) FinishAttempt(ctx context.Context, deliveryID, owner string, target domain.RunState) error {
	if target == domain.RunHumanGate {
		return errors.New("final human gate requires exact promotion proof projection")
	}
	if err := domain.ValidateRunTransition(domain.RunRunning, target); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE delivery_runs SET state = ?, owner = '', updated_at = ?
        WHERE delivery_id = ? AND state = ? AND owner = ?`, target, time.Now().UTC().UnixNano(),
		deliveryID, domain.RunRunning, owner)
	if err != nil {
		return fmt.Errorf("finish delivery attempt: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return errors.New("delivery attempt is stale or owned by another controller")
	}
	return nil
}

func (s *Store) IsRecoveredPromotionDecision(ctx context.Context, request DecisionRequest) (bool, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM promotion_journals WHERE delivery_id = ?
		AND generation = ? AND unit_id = ? AND state = ? AND cleanup_complete = 1
		AND decisions_resolved = 0`, request.DeliveryID, request.Generation, request.UnitID,
		PromotionJournalComplete).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) CancelUnpublishedRecoveredPromotionDecision(ctx context.Context, lease Lease,
	request DecisionRequest, now time.Time) error {
	if lease.RunID != request.DeliveryID || request.Status != DecisionRequestOpen || now.IsZero() {
		return errors.New("fenced unpublished recovered promotion decision is required")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE decision_requests SET status = ?, updated_at = ?
		WHERE request_id = ? AND delivery_id = ? AND generation = ? AND unit_id = ? AND status = ?
		AND EXISTS (SELECT 1 FROM leases WHERE run_id = ? AND owner = ? AND epoch = ? AND expires_at > ?)
		AND EXISTS (SELECT 1 FROM promotion_journals WHERE delivery_id = ? AND generation = ?
			AND unit_id = ? AND state = ? AND cleanup_complete = 1 AND decisions_resolved = 0)`,
		DecisionRequestCancelled, now.UTC().UnixNano(), request.RequestID, request.DeliveryID,
		request.Generation, request.UnitID, DecisionRequestOpen, lease.RunID, lease.Owner, lease.Epoch,
		now.UTC().UnixNano(), request.DeliveryID, request.Generation, request.UnitID, PromotionJournalComplete)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return errors.New("recovered promotion decision cancellation is stale")
	}
	return nil
}

func (s *Store) ResolveRecoveredPromotionDecisions(ctx context.Context, lease Lease, now time.Time) error {
	if lease.RunID == "" || now.IsZero() {
		return errors.New("fenced recovered promotion decision identity is required")
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
	var owner string
	var epoch, expiresAt int64
	if err := conn.QueryRowContext(ctx, `SELECT owner, epoch, expires_at FROM leases WHERE run_id = ?`,
		lease.RunID).Scan(&owner, &epoch, &expiresAt); err != nil {
		return err
	}
	if owner != lease.Owner || epoch != lease.Epoch || now.UTC().UnixNano() >= expiresAt {
		return errors.New("recovered promotion decision lease is stale")
	}
	if _, err := conn.ExecContext(ctx, `UPDATE decision_requests SET status = ?, updated_at = ?
		WHERE delivery_id = ? AND status IN (?, ?) AND EXISTS (
			SELECT 1 FROM promotion_journals p WHERE p.delivery_id = decision_requests.delivery_id
			AND p.generation = decision_requests.generation AND p.unit_id = decision_requests.unit_id
			AND p.state = ? AND p.cleanup_complete = 1 AND p.decisions_resolved = 0
		)`, DecisionRequestCancelled, now.UTC().UnixNano(), lease.RunID,
		DecisionRequestOpen, DecisionRequestPublished, PromotionJournalComplete); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, `UPDATE promotion_journals SET decisions_resolved = 1, updated_at = ?
		WHERE delivery_id = ? AND state = ? AND cleanup_complete = 1 AND decisions_resolved = 0
		AND NOT EXISTS (SELECT 1 FROM decision_requests q WHERE q.delivery_id = promotion_journals.delivery_id
			AND q.generation = promotion_journals.generation AND q.unit_id = promotion_journals.unit_id
			AND q.status = ?)`, now.UTC().UnixNano(), lease.RunID, PromotionJournalComplete,
		DecisionRequestAnswered); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Store) ProjectFinalHumanGate(ctx context.Context, lease Lease, generation int64,
	headSHA string, now time.Time) error {
	if lease.RunID == "" || generation <= 0 || !validGitSHA(headSHA) || now.IsZero() {
		return errors.New("fenced final human gate identity is required")
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
	var owner string
	var epoch, expiresAt int64
	if err := conn.QueryRowContext(ctx, `SELECT owner, epoch, expires_at FROM leases WHERE run_id = ?`,
		lease.RunID).Scan(&owner, &epoch, &expiresAt); err != nil {
		return err
	}
	if owner != lease.Owner || epoch != lease.Epoch || now.UTC().UnixNano() >= expiresAt {
		return errors.New("final human gate projection lease is stale")
	}
	result, err := conn.ExecContext(ctx, `UPDATE delivery_runs SET state = ?, owner = '', updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND state IN (?, ?, ?, ?, ?, ?, ?) AND
		(state <> ? OR owner = ?) AND NOT EXISTS (
			SELECT 1 FROM decision_requests q WHERE q.delivery_id = delivery_runs.delivery_id
			AND q.generation = delivery_runs.generation AND (q.status IN (?, ?, ?, ?) OR
				(q.status = ? AND q.accepted_answer NOT IN (?, ?)))
		) AND EXISTS (
			SELECT 1 FROM promotion_journals p WHERE p.delivery_id = delivery_runs.delivery_id
			AND p.candidate_head = ? AND p.validated_head = ? AND p.state = ? AND p.cleanup_complete = 1
			AND p.decisions_resolved = 1 AND (
				p.generation = delivery_runs.generation OR (
					p.generation + 1 = delivery_runs.generation AND EXISTS (
						SELECT 1 FROM human_decisions h WHERE h.delivery_id = p.delivery_id
						AND h.generation = p.generation AND h.approved = 1
					)
				)
			)
		)`, domain.RunHumanGate, now.UTC().UnixNano(), lease.RunID, generation,
		domain.RunPlanned, domain.RunReady, domain.RunFailed, domain.RunBlocked, domain.RunAwaitingDecision,
		domain.RunHumanGate, domain.RunRunning, domain.RunRunning, lease.Owner,
		DecisionRequestOpen, DecisionRequestPublished, DecisionRequestAnswered, DecisionRequestExpired,
		DecisionRequestConsumed, "retry", "continue", headSHA, headSHA, PromotionJournalComplete)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return errors.New("delivery is not eligible for final human gate projection")
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Store) BlockAwaitingDecision(ctx context.Context, deliveryID string, generation int64) error {
	if deliveryID == "" || generation <= 0 {
		return errors.New("delivery and generation are required")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE delivery_runs SET state = ?, owner = '', updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND state = ?`, domain.RunBlocked,
		time.Now().UTC().UnixNano(), deliveryID, generation, domain.RunAwaitingDecision)
	if err != nil {
		return fmt.Errorf("block awaiting decision: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return errors.New("awaiting decision delivery is stale or already advanced")
	}
	return nil
}

func (s *Store) ResumeDelivery(ctx context.Context, decision domain.HumanDecision) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var state domain.RunState
	var generation int64
	if err := conn.QueryRowContext(ctx, `SELECT state, generation FROM delivery_runs WHERE delivery_id = ?`,
		decision.RunID).Scan(&state, &generation); err != nil {
		return fmt.Errorf("read blocked delivery: %w", err)
	}
	next, err := domain.ResumeStopped(decision.RunID, generation, state, decision)
	if err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, `INSERT INTO human_decisions(delivery_id, generation, approved, consumed_at)
        VALUES (?, ?, 1, ?)`, decision.RunID, generation, time.Now().UTC().UnixNano()); err != nil {
		return errors.New("human decision was already consumed")
	}
	if _, err := conn.ExecContext(ctx, `UPDATE delivery_runs SET state = ?, generation = ?, owner = '', updated_at = ?
        WHERE delivery_id = ?`, domain.RunReady, next, time.Now().UTC().UnixNano(), decision.RunID); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Store) GetDelivery(ctx context.Context, id string) (Delivery, error) {
	var delivery Delivery
	err := s.db.QueryRowContext(ctx, `SELECT delivery_id, issue, parent_issue, repository, pull_request,
		work_dir, context_hash, milestone_id, branch, base_branch, gsd_project_root, initial_head, gsd_version
		FROM deliveries WHERE delivery_id = ?`, id).Scan(&delivery.ID, &delivery.Issue,
		&delivery.ParentIssue, &delivery.Repository, &delivery.PullRequest, &delivery.WorkDir,
		&delivery.ContextHash, &delivery.MilestoneID,
		&delivery.Branch, &delivery.BaseBranch, &delivery.GSDProjectRoot, &delivery.InitialHead,
		&delivery.GSDVersion)
	if err != nil {
		return Delivery{}, fmt.Errorf("read delivery: %w", err)
	}
	return delivery, nil
}

func (s *Store) HasLegacyExternalEffects(ctx context.Context) (bool, error) {
	var tables int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'table' AND name = 'outbox'`).Scan(&tables); err != nil {
		return false, err
	}
	if tables == 0 {
		return false, nil
	}
	var effects int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox`).Scan(&effects); err != nil {
		return false, err
	}
	return effects > 0, nil
}

func (s *Store) GetDeliveryRun(ctx context.Context, deliveryID string) (DeliveryRun, error) {
	if deliveryID == "" {
		return DeliveryRun{}, errors.New("delivery identity is required")
	}
	var run DeliveryRun
	err := s.db.QueryRowContext(ctx, `SELECT delivery_id, state, generation, attempt, owner
		FROM delivery_runs WHERE delivery_id = ?`, deliveryID).Scan(&run.DeliveryID, &run.State,
		&run.Generation, &run.Attempt, &run.Owner)
	if err != nil {
		return DeliveryRun{}, fmt.Errorf("read delivery run: %w", err)
	}
	return run, nil
}

func (s *Store) BindMilestone(ctx context.Context, deliveryID, milestoneID string) error {
	if deliveryID == "" || milestoneID == "" {
		return errors.New("delivery and milestone are required")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE deliveries SET milestone_id = ?, updated_at = ?
        WHERE delivery_id = ? AND (milestone_id = '' OR milestone_id = ?)`, milestoneID,
		time.Now().UTC().UnixNano(), deliveryID, milestoneID)
	if err != nil {
		return fmt.Errorf("bind milestone: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return errors.New("delivery milestone is missing or already bound differently")
	}
	return nil
}

func (s *Store) UpsertDecisionRequest(ctx context.Context, request DecisionRequest) (DecisionRequest, error) {
	if request.Status == "" {
		request.Status = DecisionRequestOpen
	}
	if err := validateDecisionRequest(request); err != nil {
		return DecisionRequest{}, err
	}
	delivery, err := s.GetDelivery(ctx, request.DeliveryID)
	if err != nil {
		return DecisionRequest{}, err
	}
	if delivery.Repository != request.Repository || delivery.Issue != request.Issue ||
		delivery.PullRequest != request.PullRequest {
		return DecisionRequest{}, errors.New("decision request target does not match immutable delivery authority")
	}
	optionsRaw, err := jsonMarshalStrings(request.Options)
	if err != nil {
		return DecisionRequest{}, err
	}
	now := time.Now().UTC().UnixNano()
	_, err = s.db.ExecContext(ctx, `INSERT INTO decision_requests
		(request_id, delivery_id, repository, issue, pull_request, unit_id, generation, head_sha, kind,
		 evidence, options_json, recommended_option, safe_default, expires_at, status,
		 created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(request_id) DO NOTHING`, request.RequestID, request.DeliveryID, request.Repository,
		request.Issue, request.PullRequest, request.UnitID, request.Generation, request.HeadSHA, request.Kind,
		request.Evidence, optionsRaw, request.RecommendedOption, request.SafeDefault,
		request.ExpiresAt.UTC().UnixNano(), request.Status, now, now)
	if err != nil {
		return DecisionRequest{}, fmt.Errorf("upsert decision request: %w", err)
	}
	existing, err := s.GetDecisionRequest(ctx, request.RequestID)
	if err != nil {
		return DecisionRequest{}, err
	}
	if existing.DeliveryID != request.DeliveryID || existing.Repository != request.Repository ||
		existing.Issue != request.Issue || existing.PullRequest != request.PullRequest ||
		existing.Generation != request.Generation || existing.HeadSHA != request.HeadSHA ||
		existing.UnitID != request.UnitID || existing.Kind != request.Kind ||
		strings.Join(existing.Options, "\x00") != strings.Join(request.Options, "\x00") {
		return DecisionRequest{}, errors.New("decision request id collides with different identity")
	}
	return existing, nil
}

func (s *Store) CancelStaleDecisionRequests(ctx context.Context, lease Lease, deliveryID, repository string,
	issue, pullRequest int, generation int64, headSHA string, now time.Time) error {
	if lease.RunID != deliveryID || !validRepository(repository) || issue <= 0 || pullRequest <= 0 ||
		generation <= 0 || !validGitSHA(headSHA) || now.IsZero() {
		return errors.New("complete fenced current decision-request identity is required")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var owner string
	var epoch, expiresAt int64
	if err := conn.QueryRowContext(ctx, `SELECT owner, epoch, expires_at FROM leases WHERE run_id = ?`,
		deliveryID).Scan(&owner, &epoch, &expiresAt); err != nil {
		return err
	}
	if owner != lease.Owner || epoch != lease.Epoch || now.UTC().UnixNano() >= expiresAt {
		return errors.New("stale decision-request cancellation lease")
	}
	if _, err := conn.ExecContext(ctx, `UPDATE decision_requests SET status = ?, updated_at = ?
		WHERE delivery_id = ? AND status IN (?, ?) AND (repository <> ? OR issue <> ? OR pull_request <> ?
		OR generation <> ? OR head_sha <> ?) AND NOT EXISTS (
			SELECT 1 FROM promotion_journals p WHERE p.delivery_id = decision_requests.delivery_id
			AND p.generation = decision_requests.generation AND p.unit_id = decision_requests.unit_id
			AND p.state = ? AND p.cleanup_complete = 1 AND p.decisions_resolved = 0
		)`, DecisionRequestCancelled, now.UTC().UnixNano(), deliveryID, DecisionRequestOpen,
		DecisionRequestPublished, repository, issue, pullRequest, generation, headSHA,
		PromotionJournalComplete); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Store) CancelUnansweredPromotionDecisions(ctx context.Context, lease Lease, now time.Time) error {
	if lease.RunID == "" || now.IsZero() {
		return errors.New("fenced promotion decision cancellation is required")
	}
	_, err := s.db.ExecContext(ctx, `UPDATE decision_requests SET status = ?, updated_at = ?
		WHERE delivery_id = ? AND status IN (?, ?) AND EXISTS (
			SELECT 1 FROM leases WHERE run_id = ? AND owner = ? AND epoch = ? AND expires_at > ?
		) AND EXISTS (
			SELECT 1 FROM promotion_journals p WHERE p.delivery_id = decision_requests.delivery_id
			AND p.generation = decision_requests.generation AND p.unit_id = decision_requests.unit_id
			AND p.state <> ? AND p.decisions_resolved = 0
		) AND (EXISTS (
			SELECT 1 FROM decision_requests approved WHERE approved.delivery_id = decision_requests.delivery_id
			AND approved.generation = decision_requests.generation AND approved.unit_id = decision_requests.unit_id
			AND approved.status = ? AND approved.accepted_answer IN (?, ?)
		) OR EXISTS (
			SELECT 1 FROM promotion_journals completed WHERE completed.delivery_id = decision_requests.delivery_id
			AND completed.generation = decision_requests.generation AND completed.unit_id = decision_requests.unit_id
			AND completed.state = ? AND completed.cleanup_complete = 1
		))`, DecisionRequestCancelled, now.UTC().UnixNano(), lease.RunID, DecisionRequestOpen,
		DecisionRequestPublished, lease.RunID, lease.Owner, lease.Epoch, now.UTC().UnixNano(),
		PromotionJournalBlocked, DecisionRequestConsumed, "retry", "continue", PromotionJournalComplete)
	return err
}

func (s *Store) ResumeDeliveryFenced(ctx context.Context, lease Lease, decision domain.HumanDecision,
	now time.Time,
) error {
	if lease.RunID == "" || decision.RunID != lease.RunID || decision.Generation <= 0 || now.IsZero() {
		return errors.New("fenced resume decision is required")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var leaseOwner string
	var leaseEpoch, leaseExpires int64
	if err := conn.QueryRowContext(ctx, `SELECT owner, epoch, expires_at FROM leases WHERE run_id = ?`,
		lease.RunID).Scan(&leaseOwner, &leaseEpoch, &leaseExpires); err != nil {
		return err
	}
	if leaseOwner != lease.Owner || leaseEpoch != lease.Epoch || now.UTC().UnixNano() >= leaseExpires {
		return errors.New("resume controller lease is stale")
	}
	var state domain.RunState
	var generation int64
	if err := conn.QueryRowContext(ctx, `SELECT state, generation FROM delivery_runs WHERE delivery_id = ?`,
		decision.RunID).Scan(&state, &generation); err != nil {
		return err
	}
	next, err := domain.ResumeStopped(decision.RunID, generation, state, decision)
	if err != nil {
		return err
	}
	stamp := now.UTC().UnixNano()
	if _, err := conn.ExecContext(ctx, `UPDATE decision_requests SET status = ?, updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND status IN (?, ?)`, DecisionRequestCancelled,
		stamp, decision.RunID, generation, DecisionRequestOpen, DecisionRequestPublished); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, `INSERT INTO human_decisions(delivery_id, generation, approved, consumed_at)
		VALUES (?, ?, 1, ?)`, decision.RunID, generation, stamp); err != nil {
		return errors.New("human decision was already consumed")
	}
	result, err := conn.ExecContext(ctx, `UPDATE delivery_runs SET state = ?, generation = ?, owner = '', updated_at = ?
		WHERE delivery_id = ? AND state = ? AND generation = ?`, domain.RunReady, next, stamp,
		decision.RunID, state, generation)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return errors.New("delivery changed during fenced resume")
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Store) HasBlockingDecisionDisposition(ctx context.Context, deliveryID string,
	generation int64) (bool, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM decision_requests WHERE delivery_id = ?
		AND generation = ? AND (status = ? OR (status = ? AND accepted_answer NOT IN (?, ?)))`,
		deliveryID, generation, DecisionRequestExpired, DecisionRequestConsumed, "retry", "continue").Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) HasOutstandingDecisionRequest(ctx context.Context, deliveryID, repository string, issue,
	pullRequest int, generation int64, headSHA string) (bool, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM decision_requests WHERE delivery_id = ?
		AND repository = ? AND issue = ? AND pull_request = ? AND generation = ? AND head_sha = ?
		AND status IN (?, ?, ?, ?)`, deliveryID, repository, issue, pullRequest, generation, headSHA,
		DecisionRequestOpen, DecisionRequestPublished, DecisionRequestAnswered, DecisionRequestConsumed).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) ListOpenDecisionRequests(ctx context.Context, deliveryID string) ([]DecisionRequest, error) {
	return s.listDecisionRequestsByStatus(ctx, deliveryID, DecisionRequestOpen, DecisionRequestPublished)
}

func (s *Store) ListDecisionAnswerRecovery(ctx context.Context, deliveryID string, generation int64,
	headSHA string) ([]DecisionRequest, error) {
	if deliveryID == "" || generation <= 0 || !validGitSHA(headSHA) {
		return nil, errors.New("current decision answer recovery identity is required")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT request_id FROM decision_requests WHERE delivery_id = ?
		AND status IN (?, ?, ?) AND ((generation = ? AND head_sha = ?) OR EXISTS (
			SELECT 1 FROM promotion_journals p WHERE p.delivery_id = decision_requests.delivery_id
			AND p.generation = decision_requests.generation AND p.unit_id = decision_requests.unit_id
			AND p.state = ? AND p.cleanup_complete = 1 AND p.decisions_resolved = 0
		)) ORDER BY created_at`, deliveryID, DecisionRequestPublished, DecisionRequestAnswered,
		DecisionRequestConsumed, generation, headSHA, PromotionJournalComplete)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	requests := make([]DecisionRequest, 0, len(ids))
	for _, id := range ids {
		request, err := s.GetDecisionRequest(ctx, id)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func (s *Store) listDecisionRequestsByStatus(ctx context.Context, deliveryID string,
	statuses ...DecisionRequestStatus) ([]DecisionRequest, error) {
	if deliveryID == "" || len(statuses) == 0 {
		return nil, errors.New("delivery and decision request statuses are required")
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(statuses)), ",")
	arguments := make([]any, 0, len(statuses)+1)
	arguments = append(arguments, deliveryID)
	for _, status := range statuses {
		arguments = append(arguments, status)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT request_id FROM decision_requests WHERE delivery_id = ? AND status IN (`+
		placeholders+`) ORDER BY created_at`, arguments...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	requests := make([]DecisionRequest, 0, len(ids))
	for _, id := range ids {
		request, err := s.GetDecisionRequest(ctx, id)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func (s *Store) GetDecisionRequest(ctx context.Context, requestID string) (DecisionRequest, error) {
	var request DecisionRequest
	var optionsRaw string
	var expiresAt, acceptedAt, consumedAt int64
	err := s.db.QueryRowContext(ctx, `SELECT request_id, delivery_id, repository, issue, pull_request, unit_id,
		generation, head_sha, kind, evidence, options_json, recommended_option, safe_default,
		expires_at, github_comment_id, status, accepted_answer, accepted_by, accepted_actor_id,
		accepted_comment_id, accepted_at, consumed_at FROM decision_requests WHERE request_id = ?`, requestID).Scan(&request.RequestID, &request.DeliveryID,
		&request.Repository, &request.Issue, &request.PullRequest, &request.UnitID, &request.Generation, &request.HeadSHA,
		&request.Kind, &request.Evidence, &optionsRaw, &request.RecommendedOption, &request.SafeDefault,
		&expiresAt, &request.GitHubCommentID, &request.Status, &request.AcceptedAnswer, &request.AcceptedBy,
		&request.AcceptedActorID, &request.AcceptedCommentID, &acceptedAt, &consumedAt)
	if err != nil {
		return DecisionRequest{}, fmt.Errorf("read decision request: %w", err)
	}
	options, err := jsonUnmarshalStrings(optionsRaw)
	if err != nil {
		return DecisionRequest{}, err
	}
	request.Options = options
	request.ExpiresAt = time.Unix(0, expiresAt).UTC()
	if acceptedAt > 0 {
		request.AcceptedAt = time.Unix(0, acceptedAt).UTC()
	}
	if consumedAt > 0 {
		request.ConsumedAt = time.Unix(0, consumedAt).UTC()
	}
	return request, nil
}

func (s *Store) MarkDecisionRequestPublishedFenced(ctx context.Context, lease Lease,
	request DecisionRequest, commentID int64, now time.Time) error {
	if request.RequestID == "" || commentID <= 0 || now.IsZero() || lease.RunID != request.DeliveryID {
		return errors.New("fenced decision request and comment identity are required")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var owner string
	var epoch, expiresAt int64
	if err := conn.QueryRowContext(ctx, `SELECT owner, epoch, expires_at FROM leases WHERE run_id = ?`,
		lease.RunID).Scan(&owner, &epoch, &expiresAt); err != nil {
		return err
	}
	if owner != lease.Owner || epoch != lease.Epoch || now.UTC().UnixNano() >= expiresAt {
		return errors.New("decision publication projection lease is stale")
	}
	result, err := conn.ExecContext(ctx, `UPDATE decision_requests SET status = ?, github_comment_id = ?, updated_at = ?
		WHERE request_id = ? AND delivery_id = ? AND repository = ? AND issue = ? AND pull_request = ?
		AND generation = ? AND head_sha = ? AND status IN (?, ?)
		AND (github_comment_id = 0 OR github_comment_id = ?)`, DecisionRequestPublished, commentID,
		now.UTC().UnixNano(), request.RequestID, request.DeliveryID, request.Repository, request.Issue,
		request.PullRequest, request.Generation, request.HeadSHA,
		DecisionRequestOpen, DecisionRequestPublished, commentID)
	if err != nil {
		return fmt.Errorf("mark decision request published: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return errors.New("decision request publication projection is stale or rebound")
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Store) ExpireDecisionRequestAndBlock(ctx context.Context, lease Lease, requestID string, now time.Time) error {
	if requestID == "" || now.IsZero() {
		return errors.New("decision request and expiry time are required")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var leaseOwner string
	var leaseEpoch, leaseExpires int64
	if err := conn.QueryRowContext(ctx, `SELECT owner, epoch, expires_at FROM leases WHERE run_id = ?`,
		lease.RunID).Scan(&leaseOwner, &leaseEpoch, &leaseExpires); err != nil {
		return err
	}
	if leaseOwner != lease.Owner || leaseEpoch != lease.Epoch || now.UTC().UnixNano() >= leaseExpires {
		return errors.New("decision expiry controller lease is stale")
	}
	var deliveryID, safeDefault string
	var generation, expiresAt int64
	var status DecisionRequestStatus
	if err := conn.QueryRowContext(ctx, `SELECT delivery_id, generation, safe_default, expires_at, status
		FROM decision_requests WHERE request_id = ?`, requestID).Scan(&deliveryID, &generation,
		&safeDefault, &expiresAt, &status); err != nil {
		return err
	}
	if deliveryID != lease.RunID || safeDefault != "stop" || now.UTC().UnixNano() < expiresAt {
		return errors.New("decision request is not eligible for safe-stop expiry")
	}
	if status != DecisionRequestExpired {
		if status != DecisionRequestOpen && status != DecisionRequestPublished {
			return errors.New("decision request is not open for expiry")
		}
		if _, err := conn.ExecContext(ctx, `UPDATE decision_requests SET status = ?, updated_at = ?
			WHERE request_id = ? AND status = ?`, DecisionRequestExpired, now.UTC().UnixNano(), requestID,
			status); err != nil {
			return err
		}
	}
	var runState domain.RunState
	var runGeneration int64
	if err := conn.QueryRowContext(ctx, `SELECT state, generation FROM delivery_runs WHERE delivery_id = ?`,
		deliveryID).Scan(&runState, &runGeneration); err != nil {
		return err
	}
	if runGeneration != generation {
		return errors.New("expired decision request generation is stale")
	}
	if runState != domain.RunBlocked {
		if runState != domain.RunReady && runState != domain.RunRunning && runState != domain.RunAwaitingDecision {
			return errors.New("delivery state cannot accept safe-stop decision expiry")
		}
		if _, err := conn.ExecContext(ctx, `UPDATE delivery_runs SET state = ?, owner = '', updated_at = ?
			WHERE delivery_id = ? AND generation = ? AND state = ?`, domain.RunBlocked,
			now.UTC().UnixNano(), deliveryID, generation, runState); err != nil {
			return err
		}
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Store) ApplyDecisionRequestAnswer(ctx context.Context, lease Lease, requestID, answer, actor string,
	actorID, commentID, generation int64, headSHA string, replyAt, now time.Time) error {
	answer, actor = strings.TrimSpace(answer), strings.TrimSpace(actor)
	if requestID == "" || answer == "" || actor == "" || actorID <= 0 || commentID <= 0 || generation <= 0 ||
		!validGitSHA(headSHA) || replyAt.IsZero() || now.IsZero() || replyAt.After(now) {
		return errors.New("complete decision answer identity is required")
	}
	request, err := s.GetDecisionRequest(ctx, requestID)
	if err != nil {
		return err
	}
	allowed := false
	for _, option := range request.Options {
		if answer == option {
			allowed = true
			break
		}
	}
	if !allowed || lease.RunID != request.DeliveryID {
		return errors.New("decision answer is not one of the bounded options or delivery authority is stale")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var leaseOwner string
	var leaseEpoch, leaseExpires int64
	if err := conn.QueryRowContext(ctx, `SELECT owner, epoch, expires_at FROM leases WHERE run_id = ?`,
		lease.RunID).Scan(&leaseOwner, &leaseEpoch, &leaseExpires); err != nil {
		return err
	}
	if leaseOwner != lease.Owner || leaseEpoch != lease.Epoch || now.UTC().UnixNano() >= leaseExpires {
		return errors.New("decision answer controller lease is stale")
	}
	var status DecisionRequestStatus
	var storedAnswer, storedActor string
	var storedActorID, questionCommentID, acceptedCommentID, expiresAt int64
	if err := conn.QueryRowContext(ctx, `SELECT status, accepted_answer, accepted_by, accepted_actor_id,
		github_comment_id, accepted_comment_id, expires_at FROM decision_requests WHERE request_id = ? AND delivery_id = ?
		AND generation = ? AND head_sha = ?`, requestID, lease.RunID, generation, headSHA).Scan(&status,
		&storedAnswer, &storedActor, &storedActorID, &questionCommentID, &acceptedCommentID, &expiresAt); err != nil {
		return err
	}
	switch status {
	case DecisionRequestPublished:
		if questionCommentID <= 0 || commentID <= questionCommentID || replyAt.UTC().UnixNano() >= expiresAt {
			return errors.New("decision reply was not created after the question and before expiry")
		}
	case DecisionRequestAnswered:
		if storedAnswer != answer || storedActor != actor || storedActorID != actorID ||
			acceptedCommentID != commentID {
			return errors.New("persisted decision reply identity cannot be rebound")
		}
	case DecisionRequestConsumed:
		if storedAnswer != answer || storedActor != actor || storedActorID != actorID ||
			acceptedCommentID != commentID {
			return errors.New("persisted decision reply identity cannot be rebound")
		}
		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			return err
		}
		committed = true
		return nil
	default:
		return errors.New("decision request is not eligible for answer application")
	}
	var runState domain.RunState
	var runGeneration int64
	if err := conn.QueryRowContext(ctx, `SELECT state, generation FROM delivery_runs WHERE delivery_id = ?`,
		lease.RunID).Scan(&runState, &runGeneration); err != nil {
		return err
	}
	approved := answer == "retry" || answer == "continue"
	var promotionDecision int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM promotion_journals WHERE delivery_id = ?
		AND generation = ? AND unit_id = ? AND state <> ?`, lease.RunID, generation, request.UnitID,
		PromotionJournalBlocked).Scan(&promotionDecision); err != nil {
		return err
	}
	if promotionDecision > 0 {
		if runGeneration != generation {
			return errors.New("recovered promotion reply generation is stale")
		}
		if approved {
			if runState != domain.RunReady && runState != domain.RunFailed &&
				runState != domain.RunAwaitingDecision && runState != domain.RunBlocked {
				return errors.New("recovered promotion reply cannot resume the current delivery state")
			}
			if _, err := conn.ExecContext(ctx, `INSERT OR IGNORE INTO human_decisions
				(delivery_id, generation, approved, consumed_at) VALUES (?, ?, 1, ?)`, lease.RunID,
				generation, now.UTC().UnixNano()); err != nil {
				return err
			}
			if _, err := conn.ExecContext(ctx, `UPDATE delivery_runs SET state = ?, owner = '', updated_at = ?
				WHERE delivery_id = ? AND generation = ?`, domain.RunReady, now.UTC().UnixNano(),
				lease.RunID, generation); err != nil {
				return err
			}
		} else {
			if runState != domain.RunReady && runState != domain.RunFailed &&
				runState != domain.RunAwaitingDecision && runState != domain.RunBlocked {
				return errors.New("recovered promotion reply cannot block the current delivery state")
			}
			if _, err := conn.ExecContext(ctx, `UPDATE delivery_runs SET state = ?, owner = '', updated_at = ?
				WHERE delivery_id = ? AND generation = ?`, domain.RunBlocked, now.UTC().UnixNano(),
				lease.RunID, generation); err != nil {
				return err
			}
		}
	} else if approved {
		switch {
		case runGeneration == generation && (runState == domain.RunFailed ||
			runState == domain.RunAwaitingDecision || runState == domain.RunBlocked):
			next, err := domain.ResumeStopped(lease.RunID, runGeneration, runState, domain.HumanDecision{
				RunID: lease.RunID, Generation: generation, ActorKind: domain.ActorHuman, Approved: true,
			})
			if err != nil {
				return err
			}
			if _, err := conn.ExecContext(ctx, `INSERT INTO human_decisions(delivery_id, generation, approved, consumed_at)
				VALUES (?, ?, 1, ?)`, lease.RunID, generation, now.UTC().UnixNano()); err != nil {
				return errors.New("human decision was already consumed")
			}
			if _, err := conn.ExecContext(ctx, `UPDATE delivery_runs SET state = ?, generation = ?, owner = '', updated_at = ?
				WHERE delivery_id = ? AND state = ? AND generation = ?`, domain.RunReady, next,
				now.UTC().UnixNano(), lease.RunID, runState, generation); err != nil {
				return err
			}
		case runGeneration == generation+1 && runState == domain.RunReady:
			// The same durable answer was already applied before restart.
		default:
			return errors.New("decision reply cannot resume the current delivery generation")
		}
	} else {
		switch {
		case runGeneration == generation && runState == domain.RunAwaitingDecision:
			if _, err := conn.ExecContext(ctx, `UPDATE delivery_runs SET state = ?, owner = '', updated_at = ?
				WHERE delivery_id = ? AND generation = ? AND state = ?`, domain.RunBlocked,
				now.UTC().UnixNano(), lease.RunID, generation, domain.RunAwaitingDecision); err != nil {
				return err
			}
		case runGeneration == generation && runState == domain.RunBlocked:
			// The same safe-stop answer was already applied before restart.
		default:
			return errors.New("decision reply cannot block the current delivery generation")
		}
	}
	if _, err := conn.ExecContext(ctx, `UPDATE decision_requests SET status = ?, accepted_answer = ?,
		accepted_by = ?, accepted_actor_id = ?, accepted_comment_id = ?,
		accepted_at = CASE WHEN accepted_at = 0 THEN ? ELSE accepted_at END,
		consumed_at = ?, updated_at = ? WHERE request_id = ? AND delivery_id = ?`, DecisionRequestConsumed,
		answer, actor, actorID, commentID, replyAt.UTC().UnixNano(), now.UTC().UnixNano(), now.UTC().UnixNano(),
		requestID, lease.RunID); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Store) BeginRecoveryAttempt(context.Context, RecoveryBudgetKey, int64, time.Duration, string, string, time.Time) (RecoveryBudget, error) {
	return RecoveryBudget{}, errors.New("static recovery plans are not accepted; use structured fenced recovery evidence")
}

func (s *Store) ReserveRecoveryAttempt(ctx context.Context, request RecoveryReservation) (RecoveryBudget, error) {
	if err := validateRecoveryReservation(request); err != nil {
		return RecoveryBudget{}, err
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return RecoveryBudget{}, err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return RecoveryBudget{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var runState domain.RunState
	var runGeneration int64
	var runOwner string
	if err := conn.QueryRowContext(ctx, `SELECT state, generation, owner FROM delivery_runs WHERE delivery_id = ?`,
		request.DeliveryID).Scan(&runState, &runGeneration, &runOwner); err != nil {
		return RecoveryBudget{}, err
	}
	if runState != domain.RunRunning || runGeneration != request.Generation || runOwner != request.ControllerOwner {
		return RecoveryBudget{}, ErrRecoveryClaimFenced
	}
	var leaseOwner string
	var leaseEpoch, leaseExpiresAt int64
	if err := conn.QueryRowContext(ctx, `SELECT owner, epoch, expires_at FROM leases WHERE run_id = ?`,
		request.DeliveryID).Scan(&leaseOwner, &leaseEpoch, &leaseExpiresAt); err != nil || leaseOwner != request.ControllerOwner ||
		leaseEpoch != request.ControllerEpoch || time.Now().UTC().UnixNano() >= leaseExpiresAt {
		return RecoveryBudget{}, ErrRecoveryClaimFenced
	}
	stamp := request.Now.UTC().UnixNano()
	if _, err := conn.ExecContext(ctx, `UPDATE recovery_attempts SET status = 'rejected', selected_action = ?,
		rejected_reason = 'controller ownership changed before planner completion', updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ? AND failure_class = ?
		AND status = 'reserved' AND (controller_owner <> ? OR controller_epoch <> ?)`, recovery.ActionBlock, stamp,
		request.DeliveryID, request.Generation, request.UnitID, request.HeadSHA, request.FailureClass,
		request.ControllerOwner, request.ControllerEpoch); err != nil {
		return RecoveryBudget{}, err
	}
	_, err = conn.ExecContext(ctx, `INSERT OR IGNORE INTO recovery_budgets
		(delivery_id, generation, unit_id, head_sha, failure_class, attempts, max_attempts,
		 backoff_ms, policy_version, base_backoff_ms, max_backoff_ms, bounded_plan_json, status, updated_at)
		VALUES (?, ?, ?, ?, ?, 0, ?, 0, ?, ?, ?, '[]', 'ready', ?)`,
		request.DeliveryID, request.Generation, request.UnitID, request.HeadSHA, request.FailureClass,
		request.MaxAttempts, request.PolicyVersion, request.BaseBackoff.Milliseconds(),
		request.MaxBackoff.Milliseconds(), stamp)
	if err != nil {
		return RecoveryBudget{}, fmt.Errorf("initialize recovery budget: %w", err)
	}
	budget, err := getRecoveryBudgetWith(ctx, conn, request.RecoveryBudgetKey)
	if err != nil {
		return RecoveryBudget{}, err
	}
	if budget.Status == "exhausted" {
		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			return RecoveryBudget{}, err
		}
		committed = true
		return budget, ErrRetryBudgetExhausted
	}
	if budget.PolicyVersion != request.PolicyVersion || budget.MaxAttempts != request.MaxAttempts ||
		budget.BaseBackoff != request.BaseBackoff || budget.MaxBackoff != request.MaxBackoff {
		return RecoveryBudget{}, errors.New("recovery budget policy cannot change within its key")
	}
	var existingClaim, existingOwner string
	var existingEpoch int64
	existingErr := conn.QueryRowContext(ctx, `SELECT claim_token, controller_owner, controller_epoch FROM recovery_attempts
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ? AND failure_class = ? AND unit_attempt = ?`,
		request.DeliveryID, request.Generation, request.UnitID, request.HeadSHA, request.FailureClass,
		request.UnitAttempt).Scan(&existingClaim, &existingOwner, &existingEpoch)
	if existingErr == nil {
		if existingClaim != request.ClaimToken || existingOwner != request.ControllerOwner || existingEpoch != request.ControllerEpoch {
			return RecoveryBudget{}, ErrRecoveryClaimFenced
		}
		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			return RecoveryBudget{}, err
		}
		committed = true
		return budget, nil
	}
	if !errors.Is(existingErr, sql.ErrNoRows) {
		return RecoveryBudget{}, existingErr
	}
	var active int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM recovery_attempts
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ? AND failure_class = ?
		AND status = 'reserved'`, request.DeliveryID, request.Generation, request.UnitID, request.HeadSHA,
		request.FailureClass).Scan(&active); err != nil {
		return RecoveryBudget{}, err
	}
	if active != 0 {
		return RecoveryBudget{}, ErrRecoveryClaimFenced
	}
	var sequence int64
	if err := conn.QueryRowContext(ctx, `SELECT COALESCE(MAX(sequence), 0) + 1 FROM recovery_attempts
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ?`, request.DeliveryID,
		request.Generation, request.UnitID, request.HeadSHA).Scan(&sequence); err != nil {
		return RecoveryBudget{}, err
	}
	if budget.Attempts >= budget.MaxAttempts {
		budget.ExhaustedAt = request.Now.UTC()
		budget.SelectedAction = request.ExhaustedAction
		budget.Status = "exhausted"
		_, err := conn.ExecContext(ctx, `UPDATE recovery_budgets SET exhausted_at = ?, selected_action = ?,
			status = 'exhausted', updated_at = ? WHERE delivery_id = ? AND generation = ? AND unit_id = ?
			AND head_sha = ? AND failure_class = ?`, stamp, request.ExhaustedAction, stamp,
			request.DeliveryID, request.Generation, request.UnitID, request.HeadSHA, request.FailureClass)
		if err != nil {
			return RecoveryBudget{}, err
		}
		_, err = conn.ExecContext(ctx, `INSERT INTO recovery_attempts
			(delivery_id, generation, unit_id, head_sha, failure_class, unit_attempt, claim_token, controller_owner,
			 controller_epoch, reservation_index, sequence, failure_hash, diagnostic, reversible, status, selected_action, bounded_plan_json,
			 backoff_ms, next_retry_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'exhausted', ?, '[]', 0, 0, ?, ?)`,
			request.DeliveryID, request.Generation, request.UnitID, request.HeadSHA, request.FailureClass,
			request.UnitAttempt, request.ClaimToken, request.ControllerOwner, request.ControllerEpoch, budget.Attempts, sequence,
			request.FailureHash, request.Diagnostic, request.Reversible, request.ExhaustedAction, stamp, stamp)
		if err != nil {
			return RecoveryBudget{}, err
		}
		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			return RecoveryBudget{}, err
		}
		committed = true
		return budget, ErrRetryBudgetExhausted
	}
	budget.Attempts++
	budget.Backoff, err = recovery.Backoff(request.BaseBackoff, request.MaxBackoff, budget.Attempts)
	if err != nil {
		return RecoveryBudget{}, err
	}
	budget.NextRetryAt = request.Now.Add(budget.Backoff).UTC()
	budget.LastFailureHash = request.FailureHash
	budget.LastDiagnostic = request.Diagnostic
	budget.Status = "reserved"
	_, err = conn.ExecContext(ctx, `INSERT INTO recovery_attempts
		(delivery_id, generation, unit_id, head_sha, failure_class, unit_attempt, claim_token, controller_owner,
		 controller_epoch, reservation_index, sequence, failure_hash, diagnostic, reversible, status, bounded_plan_json, backoff_ms,
		 next_retry_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'reserved', '[]', ?, ?, ?, ?)`,
		request.DeliveryID, request.Generation, request.UnitID, request.HeadSHA, request.FailureClass,
		request.UnitAttempt, request.ClaimToken, request.ControllerOwner, request.ControllerEpoch, budget.Attempts, sequence,
		request.FailureHash, request.Diagnostic, request.Reversible, budget.Backoff.Milliseconds(), budget.NextRetryAt.UnixNano(), stamp, stamp)
	if err != nil {
		return RecoveryBudget{}, fmt.Errorf("reserve recovery attempt: %w", err)
	}
	_, err = conn.ExecContext(ctx, `UPDATE recovery_budgets SET attempts = ?, backoff_ms = ?,
		last_failure_hash = ?, last_diagnostic = ?, next_retry_at = ?, status = 'reserved', updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ? AND failure_class = ?`,
		budget.Attempts, budget.Backoff.Milliseconds(), budget.LastFailureHash, budget.LastDiagnostic,
		budget.NextRetryAt.UnixNano(), stamp, request.DeliveryID, request.Generation, request.UnitID,
		request.HeadSHA, request.FailureClass)
	if err != nil {
		return RecoveryBudget{}, err
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return RecoveryBudget{}, err
	}
	committed = true
	return budget, nil
}

func (s *Store) CompleteRecoveryAttempt(ctx context.Context, outcome RecoveryOutcome) error {
	if err := validateRecoveryOutcome(outcome); err != nil {
		return err
	}
	planRaw, err := json.Marshal(outcome.BoundedPlan)
	if err != nil {
		return err
	}
	status := "terminal"
	if outcome.SelectedAction == recovery.ActionRetrySameUnit || outcome.SelectedAction == recovery.ActionRetryAfterBackoff || outcome.SelectedAction == recovery.ActionRunRecoveryPlan {
		status = "planned"
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var reservedHash string
	var reversible bool
	if err := conn.QueryRowContext(ctx, `SELECT failure_hash, reversible FROM recovery_attempts WHERE delivery_id = ?
		AND generation = ? AND unit_id = ? AND head_sha = ? AND failure_class = ? AND unit_attempt = ?
		AND claim_token = ? AND controller_owner = ? AND controller_epoch = ? AND status = 'reserved'`, outcome.DeliveryID,
		outcome.Generation, outcome.UnitID, outcome.HeadSHA, outcome.FailureClass, outcome.UnitAttempt,
		outcome.ClaimToken, outcome.ControllerOwner, outcome.ControllerEpoch).Scan(&reservedHash, &reversible); err != nil {
		return ErrRecoveryClaimFenced
	}
	if outcome.EvidenceHash != reservedHash {
		return errors.New("recovery planner evidence does not match the reserved failure")
	}
	class, err := recovery.ParseFailureClass(outcome.FailureClass)
	if err != nil {
		return err
	}
	policy, err := recovery.PolicyFor(recovery.Failure{Class: class, Reversible: reversible})
	if err != nil {
		return err
	}
	if err := policy.ValidateAction(outcome.SelectedAction); err != nil {
		return err
	}
	stamp := time.Now().UTC().UnixNano()
	result, err := conn.ExecContext(ctx, `UPDATE recovery_attempts SET status = ?, planner_request_nonce = ?,
		evidence_hash = ?, authority_scope_hash = ?, planner_evidence_hash = ?, planner_session_id = ?,
		planner_session_fingerprint = ?, observed_model = ?, thinking = ?, selected_action = ?, bounded_plan_json = ?, issued_at = ?,
		expires_at = ?, updated_at = ? WHERE delivery_id = ? AND generation = ? AND unit_id = ?
		AND head_sha = ? AND failure_class = ? AND unit_attempt = ? AND claim_token = ? AND controller_owner = ?
		AND controller_epoch = ? AND status = 'reserved' AND EXISTS (SELECT 1 FROM delivery_runs
		WHERE delivery_id = ? AND generation = ? AND state = ? AND owner = ?)
		AND EXISTS (SELECT 1 FROM leases WHERE run_id = ? AND owner = ? AND epoch = ? AND expires_at > ?)`,
		status, outcome.PlannerRequestNonce, outcome.EvidenceHash, outcome.AuthorityScopeHash,
		outcome.PlannerEvidenceHash, outcome.PlannerSessionID, outcome.PlannerSessionFingerprint,
		outcome.ObservedModel, outcome.Thinking, outcome.SelectedAction,
		string(planRaw), outcome.IssuedAt.UTC().UnixNano(), outcome.ExpiresAt.UTC().UnixNano(), stamp,
		outcome.DeliveryID, outcome.Generation, outcome.UnitID, outcome.HeadSHA, outcome.FailureClass,
		outcome.UnitAttempt, outcome.ClaimToken, outcome.ControllerOwner, outcome.ControllerEpoch,
		outcome.DeliveryID, outcome.Generation, domain.RunRunning, outcome.ControllerOwner,
		outcome.DeliveryID, outcome.ControllerOwner, outcome.ControllerEpoch, stamp)
	if err != nil {
		return fmt.Errorf("complete recovery attempt: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return ErrRecoveryClaimFenced
	}
	_, err = conn.ExecContext(ctx, `UPDATE recovery_budgets SET planner_evidence_hash = ?,
		planner_session_id = ?, planner_session_fingerprint = ?, observed_model = ?, thinking = ?,
		selected_action = ?, bounded_plan_json = ?, status = ?, updated_at = ? WHERE delivery_id = ?
		AND generation = ? AND unit_id = ? AND head_sha = ? AND failure_class = ?`,
		outcome.PlannerEvidenceHash, outcome.PlannerSessionID, outcome.PlannerSessionFingerprint,
		outcome.ObservedModel, outcome.Thinking, outcome.SelectedAction, string(planRaw), status, stamp,
		outcome.DeliveryID, outcome.Generation, outcome.UnitID, outcome.HeadSHA, outcome.FailureClass)
	if err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Store) CompleteRecoveryDecision(ctx context.Context, key RecoveryBudgetKey, unitAttempt int64, claimToken, controllerOwner string, controllerEpoch int64, action recovery.Action, now time.Time) error {
	if err := validateRecoveryBudgetKey(key); err != nil || unitAttempt <= 0 || !validRecoveryToken(claimToken) || strings.TrimSpace(controllerOwner) == "" ||
		(action != recovery.ActionBlock && action != recovery.ActionAwaitDecision && action != recovery.ActionFinalHumanGate) || now.IsZero() {
		return errors.New("complete bounded controller recovery decision is required")
	}
	return s.finishRecoveryAttempt(ctx, key, unitAttempt, claimToken, controllerOwner, controllerEpoch, action, "terminal", "", now)
}

func (s *Store) RejectRecoveryAttempt(ctx context.Context, key RecoveryBudgetKey, unitAttempt int64, claimToken, controllerOwner string, controllerEpoch int64, action recovery.Action, reason string, now time.Time) error {
	if err := validateRecoveryBudgetKey(key); err != nil || unitAttempt <= 0 || !validRecoveryToken(claimToken) || strings.TrimSpace(controllerOwner) == "" ||
		(action != recovery.ActionBlock && action != recovery.ActionAwaitDecision) || now.IsZero() {
		return errors.New("complete bounded recovery rejection is required")
	}
	return s.finishRecoveryAttempt(ctx, key, unitAttempt, claimToken, controllerOwner, controllerEpoch, action, "rejected", boundedStoreString(reason, 256), now)
}

func (s *Store) finishRecoveryAttempt(ctx context.Context, key RecoveryBudgetKey, unitAttempt int64, claimToken, controllerOwner string, controllerEpoch int64, action recovery.Action, status, reason string, now time.Time) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var reversible bool
	if err := conn.QueryRowContext(ctx, `SELECT reversible FROM recovery_attempts WHERE delivery_id = ?
		AND generation = ? AND unit_id = ? AND head_sha = ? AND failure_class = ? AND unit_attempt = ?
		AND claim_token = ? AND controller_owner = ? AND controller_epoch = ? AND status = 'reserved'`, key.DeliveryID,
		key.Generation, key.UnitID, key.HeadSHA, key.FailureClass, unitAttempt, claimToken, controllerOwner,
		controllerEpoch).Scan(&reversible); err != nil {
		return ErrRecoveryClaimFenced
	}
	class, err := recovery.ParseFailureClass(key.FailureClass)
	if err != nil {
		return err
	}
	policy, err := recovery.PolicyFor(recovery.Failure{Class: class, Reversible: reversible})
	if err != nil {
		return err
	}
	if err := policy.ValidateAction(action); err != nil {
		return err
	}
	stamp := time.Now().UTC().UnixNano()
	result, err := conn.ExecContext(ctx, `UPDATE recovery_attempts SET status = ?, selected_action = ?,
		rejected_reason = ?, updated_at = ? WHERE delivery_id = ? AND generation = ? AND unit_id = ?
		AND head_sha = ? AND failure_class = ? AND unit_attempt = ? AND claim_token = ? AND controller_owner = ?
		AND controller_epoch = ? AND status = 'reserved' AND EXISTS (SELECT 1 FROM delivery_runs
		WHERE delivery_id = ? AND generation = ? AND state = ? AND owner = ?)
		AND EXISTS (SELECT 1 FROM leases WHERE run_id = ? AND owner = ? AND epoch = ? AND expires_at > ?)`,
		status, action, reason, stamp, key.DeliveryID, key.Generation, key.UnitID, key.HeadSHA,
		key.FailureClass, unitAttempt, claimToken, controllerOwner, controllerEpoch, key.DeliveryID,
		key.Generation, domain.RunRunning, controllerOwner, key.DeliveryID, controllerOwner, controllerEpoch, stamp)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return ErrRecoveryClaimFenced
	}
	if _, err := conn.ExecContext(ctx, `UPDATE recovery_budgets SET selected_action = ?, status = ?,
		updated_at = ? WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ?
		AND failure_class = ?`, action, status, stamp, key.DeliveryID, key.Generation, key.UnitID,
		key.HeadSHA, key.FailureClass); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *Store) GetRecoveryBudget(ctx context.Context, key RecoveryBudgetKey) (RecoveryBudget, error) {
	if err := validateRecoveryBudgetKey(key); err != nil {
		return RecoveryBudget{}, err
	}
	return getRecoveryBudgetWith(ctx, s.db, key)
}

func getRecoveryBudgetWith(ctx context.Context, query interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, key RecoveryBudgetKey) (RecoveryBudget, error) {
	budget := RecoveryBudget{RecoveryBudgetKey: key}
	var baseMS, maxMS, backoffMS, nextRetryAt, exhaustedAt int64
	var action, planJSON string
	err := query.QueryRowContext(ctx, `SELECT policy_version, attempts, max_attempts, base_backoff_ms,
		max_backoff_ms, backoff_ms, last_failure_hash, last_diagnostic, planner_evidence_hash,
		planner_session_id, planner_session_fingerprint, observed_model, thinking, selected_action,
		bounded_plan_json, next_retry_at, exhausted_at, status FROM recovery_budgets WHERE delivery_id = ?
		AND generation = ? AND unit_id = ? AND head_sha = ? AND failure_class = ?`, key.DeliveryID,
		key.Generation, key.UnitID, key.HeadSHA, key.FailureClass).Scan(&budget.PolicyVersion,
		&budget.Attempts, &budget.MaxAttempts, &baseMS, &maxMS, &backoffMS, &budget.LastFailureHash,
		&budget.LastDiagnostic, &budget.PlannerEvidenceHash, &budget.PlannerSessionID,
		&budget.PlannerSessionFingerprint, &budget.ObservedModel, &budget.Thinking, &action, &planJSON,
		&nextRetryAt, &exhaustedAt, &budget.Status)
	if err != nil {
		return RecoveryBudget{}, err
	}
	budget.BaseBackoff = time.Duration(baseMS) * time.Millisecond
	budget.MaxBackoff = time.Duration(maxMS) * time.Millisecond
	budget.Backoff = time.Duration(backoffMS) * time.Millisecond
	budget.SelectedAction = recovery.Action(action)
	if nextRetryAt != 0 {
		budget.NextRetryAt = time.Unix(0, nextRetryAt).UTC()
	}
	if exhaustedAt != 0 {
		budget.ExhaustedAt = time.Unix(0, exhaustedAt).UTC()
	}
	if err := json.Unmarshal([]byte(planJSON), &budget.BoundedPlan); err != nil {
		return RecoveryBudget{}, errors.New("stored recovery plan is malformed")
	}
	return budget, nil
}

func (s *Store) GetRecoveryAttempt(ctx context.Context, key RecoveryBudgetKey, unitAttempt int64) (RecoveryAttempt, error) {
	if err := validateRecoveryBudgetKey(key); err != nil || unitAttempt <= 0 {
		return RecoveryAttempt{}, errors.New("complete recovery attempt identity is required")
	}
	attempt := RecoveryAttempt{RecoveryOutcome: RecoveryOutcome{RecoveryBudgetKey: key, UnitAttempt: unitAttempt}}
	var planJSON, action string
	var backoffMS, nextRetryAt, issuedAt, expiresAt, dispatchedAt int64
	err := s.db.QueryRowContext(ctx, `SELECT claim_token, controller_owner, controller_epoch, reservation_index, sequence, failure_hash, diagnostic,
		reversible, status, planner_request_nonce, evidence_hash, authority_scope_hash, planner_evidence_hash,
		planner_session_id, planner_session_fingerprint, observed_model, thinking, selected_action, bounded_plan_json,
		backoff_ms, next_retry_at, issued_at, expires_at, rejected_reason, dispatch_claim, dispatch_epoch, dispatched_at
		FROM recovery_attempts WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ?
		AND failure_class = ? AND unit_attempt = ?`, key.DeliveryID, key.Generation, key.UnitID, key.HeadSHA,
		key.FailureClass, unitAttempt).Scan(&attempt.ClaimToken, &attempt.ControllerOwner, &attempt.ControllerEpoch, &attempt.ReservationIndex,
		&attempt.Sequence, &attempt.FailureHash, &attempt.Diagnostic, &attempt.Reversible, &attempt.Status, &attempt.PlannerRequestNonce,
		&attempt.EvidenceHash, &attempt.AuthorityScopeHash, &attempt.PlannerEvidenceHash,
		&attempt.PlannerSessionID, &attempt.PlannerSessionFingerprint,
		&attempt.ObservedModel, &attempt.Thinking, &action, &planJSON, &backoffMS, &nextRetryAt,
		&issuedAt, &expiresAt, &attempt.RejectedReason, &attempt.DispatchClaim, &attempt.DispatchEpoch, &dispatchedAt)
	if err != nil {
		return RecoveryAttempt{}, err
	}
	attempt.SelectedAction = recovery.Action(action)
	attempt.Backoff = time.Duration(backoffMS) * time.Millisecond
	if nextRetryAt != 0 {
		attempt.NextRetryAt = time.Unix(0, nextRetryAt).UTC()
	}
	if issuedAt != 0 {
		attempt.IssuedAt = time.Unix(0, issuedAt).UTC()
	}
	if expiresAt != 0 {
		attempt.ExpiresAt = time.Unix(0, expiresAt).UTC()
	}
	if dispatchedAt != 0 {
		attempt.DispatchedAt = time.Unix(0, dispatchedAt).UTC()
	}
	if err := json.Unmarshal([]byte(planJSON), &attempt.BoundedPlan); err != nil {
		return RecoveryAttempt{}, errors.New("stored recovery attempt plan is malformed")
	}
	return attempt, nil
}

func (s *Store) FailRecoveryDispatch(ctx context.Context, attempt RecoveryAttempt, controllerOwner string, controllerEpoch int64) error {
	if attempt.SelectedAction == "" {
		return nil
	}
	if err := validateRecoveryBudgetKey(attempt.RecoveryBudgetKey); err != nil || attempt.UnitAttempt <= 0 ||
		strings.TrimSpace(controllerOwner) == "" || controllerEpoch <= 0 {
		return errors.New("complete fenced recovery dispatch identity is required")
	}
	stamp := time.Now().UTC().UnixNano()
	result, err := s.db.ExecContext(ctx, `UPDATE recovery_attempts SET status = 'used_failed', updated_at = ?
		WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ? AND failure_class = ?
		AND unit_attempt = ? AND status = 'dispatched' AND dispatch_claim = ? AND dispatch_epoch = ?
		AND EXISTS (SELECT 1 FROM delivery_runs WHERE delivery_id = ? AND generation = ? AND state = ?
		AND owner = ?) AND EXISTS (SELECT 1 FROM leases WHERE run_id = ? AND owner = ? AND epoch = ?
		AND expires_at > ?)`, stamp, attempt.DeliveryID, attempt.Generation, attempt.UnitID, attempt.HeadSHA,
		attempt.FailureClass, attempt.UnitAttempt, controllerOwner, controllerEpoch, attempt.DeliveryID,
		attempt.Generation, domain.RunRunning, controllerOwner, attempt.DeliveryID, controllerOwner,
		controllerEpoch, stamp)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return errors.New("recovery dispatch failure disposition is fenced")
	}
	return nil
}

func (s *Store) ClaimRecoveryDispatch(ctx context.Context, deliveryID string, generation int64, unitID, headSHA, executionID string, controllerEpoch int64, now time.Time) (RecoveryAttempt, error) {
	if deliveryID == "" || generation <= 0 || unitID == "" || !validGitSHA(headSHA) || executionID == "" || controllerEpoch <= 0 || now.IsZero() {
		return RecoveryAttempt{}, errors.New("complete recovery dispatch identity is required")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return RecoveryAttempt{}, err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return RecoveryAttempt{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var class, status, dispatchOwner, selectedAction string
	var unitAttempt, nextRetryAt, expiresAt, dispatchEpoch int64
	err = conn.QueryRowContext(ctx, `SELECT failure_class, unit_attempt, status, next_retry_at, expires_at,
		dispatch_claim, dispatch_epoch, selected_action FROM recovery_attempts WHERE delivery_id = ? AND generation = ?
		AND unit_id = ? AND head_sha = ? ORDER BY sequence DESC LIMIT 1`, deliveryID, generation, unitID, headSHA).
		Scan(&class, &unitAttempt, &status, &nextRetryAt, &expiresAt, &dispatchOwner, &dispatchEpoch, &selectedAction)
	if errors.Is(err, sql.ErrNoRows) {
		var exhausted, failedWithoutDecision int
		if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM recovery_budgets WHERE delivery_id = ?
			AND generation = ? AND unit_id = ? AND head_sha = ? AND status = 'exhausted'`, deliveryID,
			generation, unitID, headSHA).Scan(&exhausted); err != nil {
			return RecoveryAttempt{}, err
		}
		if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM unit_attempts WHERE delivery_id = ?
			AND generation = ? AND unit_id = ? AND head_sha = ? AND status = 'terminal'
			AND last_failure NOT IN ('success', 'mutating_skip')`, deliveryID, generation, unitID, headSHA).Scan(&failedWithoutDecision); err != nil {
			return RecoveryAttempt{}, err
		}
		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			return RecoveryAttempt{}, err
		}
		committed = true
		if exhausted > 0 {
			return RecoveryAttempt{}, ErrRetryBudgetExhausted
		}
		if failedWithoutDecision > 0 {
			return RecoveryAttempt{}, ErrRecoveryDecisionPending
		}
		return RecoveryAttempt{}, nil
	}
	if err != nil {
		return RecoveryAttempt{}, err
	}
	var runState domain.RunState
	var runGeneration int64
	var runOwner string
	if err := conn.QueryRowContext(ctx, `SELECT state, generation, owner FROM delivery_runs WHERE delivery_id = ?`, deliveryID).
		Scan(&runState, &runGeneration, &runOwner); err != nil {
		return RecoveryAttempt{}, err
	}
	if runState != domain.RunRunning || runGeneration != generation || runOwner != executionID {
		return RecoveryAttempt{}, ErrRecoveryDispatchClaimed
	}
	var leaseCount int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM leases WHERE run_id = ? AND owner = ?
		AND epoch = ? AND expires_at > ?`, deliveryID, executionID, controllerEpoch, time.Now().UTC().UnixNano()).Scan(&leaseCount); err != nil || leaseCount != 1 {
		return RecoveryAttempt{}, ErrRecoveryDispatchClaimed
	}
	key := RecoveryBudgetKey{DeliveryID: deliveryID, Generation: generation, UnitID: unitID, HeadSHA: headSHA, FailureClass: class}
	switch status {
	case "consumed":
		var currentAttempts int64
		var unitStatus, unitOutcome string
		unitErr := conn.QueryRowContext(ctx, `SELECT attempts, status, last_failure FROM unit_attempts
			WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ?`, deliveryID,
			generation, unitID, headSHA).Scan(&currentAttempts, &unitStatus, &unitOutcome)
		if unitErr != nil && !errors.Is(unitErr, sql.ErrNoRows) {
			return RecoveryAttempt{}, unitErr
		}
		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			return RecoveryAttempt{}, err
		}
		committed = true
		if unitErr == nil && currentAttempts >= unitAttempt+1 && unitStatus == "terminal" &&
			unitOutcome != "success" && unitOutcome != "mutating_skip" {
			return RecoveryAttempt{}, ErrRecoveryDecisionPending
		}
		return RecoveryAttempt{}, nil
	case "reserved", "used_failed":
		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			return RecoveryAttempt{}, err
		}
		committed = true
		return RecoveryAttempt{}, ErrRecoveryDecisionPending
	case "exhausted":
		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			return RecoveryAttempt{}, err
		}
		committed = true
		return RecoveryAttempt{}, ErrRetryBudgetExhausted
	case "terminal", "rejected":
		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			return RecoveryAttempt{}, err
		}
		committed = true
		if recovery.Action(selectedAction) == recovery.ActionAwaitDecision || recovery.Action(selectedAction) == recovery.ActionFinalHumanGate {
			return RecoveryAttempt{}, ErrRetryBudgetExhausted
		}
		return RecoveryAttempt{}, ErrRecoveryTerminal
	case "planned", "dispatched":
		if action := recovery.Action(selectedAction); action != recovery.ActionRetrySameUnit && action != recovery.ActionRetryAfterBackoff && action != recovery.ActionRunRecoveryPlan {
			return RecoveryAttempt{}, ErrRecoveryTerminal
		}
	default:
		return RecoveryAttempt{}, ErrRecoveryTerminal
	}
	if now.Before(time.Unix(0, nextRetryAt)) {
		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			return RecoveryAttempt{}, err
		}
		committed = true
		if err := conn.Close(); err != nil {
			return RecoveryAttempt{}, err
		}
		attempt, getErr := s.GetRecoveryAttempt(ctx, key, unitAttempt)
		if getErr != nil {
			return RecoveryAttempt{}, getErr
		}
		return attempt, ErrRecoveryBackoffPending
	}
	stamp := now.UTC().UnixNano()
	if expiresAt == 0 || !now.Before(time.Unix(0, expiresAt)) {
		if _, err := conn.ExecContext(ctx, `UPDATE recovery_attempts SET status = 'terminal', selected_action = ?,
			rejected_reason = 'planner evidence expired before dispatch', updated_at = ? WHERE delivery_id = ?
			AND generation = ? AND unit_id = ? AND head_sha = ? AND failure_class = ? AND unit_attempt = ?`,
			recovery.ActionAwaitDecision, stamp, deliveryID, generation, unitID, headSHA, class, unitAttempt); err != nil {
			return RecoveryAttempt{}, err
		}
		if _, err := conn.ExecContext(ctx, `UPDATE recovery_budgets SET status = 'exhausted', selected_action = ?,
			exhausted_at = ?, next_retry_at = 0, updated_at = ? WHERE delivery_id = ? AND generation = ?
			AND unit_id = ? AND head_sha = ? AND failure_class = ?`, recovery.ActionAwaitDecision, stamp, stamp,
			deliveryID, generation, unitID, headSHA, class); err != nil {
			return RecoveryAttempt{}, err
		}
		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			return RecoveryAttempt{}, err
		}
		committed = true
		return RecoveryAttempt{}, ErrRetryBudgetExhausted
	}
	if status == "dispatched" && dispatchOwner == executionID && dispatchEpoch == controllerEpoch {
		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			return RecoveryAttempt{}, err
		}
		committed = true
		if err := conn.Close(); err != nil {
			return RecoveryAttempt{}, err
		}
		return s.GetRecoveryAttempt(ctx, key, unitAttempt)
	}
	result, err := conn.ExecContext(ctx, `UPDATE recovery_attempts SET status = 'dispatched',
		dispatch_claim = ?, dispatch_epoch = ?, dispatched_at = ?, updated_at = ? WHERE delivery_id = ? AND generation = ?
		AND unit_id = ? AND head_sha = ? AND failure_class = ? AND unit_attempt = ?
		AND status IN ('planned', 'dispatched') AND dispatch_claim = ? AND dispatch_epoch = ?`,
		executionID, controllerEpoch, stamp, stamp, deliveryID, generation, unitID, headSHA, class,
		unitAttempt, dispatchOwner, dispatchEpoch)
	if err != nil {
		return RecoveryAttempt{}, err
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return RecoveryAttempt{}, ErrRecoveryDispatchClaimed
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return RecoveryAttempt{}, err
	}
	committed = true
	if err := conn.Close(); err != nil {
		return RecoveryAttempt{}, err
	}
	return s.GetRecoveryAttempt(ctx, key, unitAttempt)
}

func validateRecoveryReservation(request RecoveryReservation) error {
	if err := validateRecoveryBudgetKey(request.RecoveryBudgetKey); err != nil {
		return err
	}
	if request.UnitAttempt <= 0 || !validRecoveryToken(request.ClaimToken) || strings.TrimSpace(request.ControllerOwner) == "" ||
		strings.ContainsAny(request.ControllerOwner, "\r\n\x00") || request.ControllerEpoch <= 0 || request.PolicyVersion != recovery.PolicyVersion ||
		request.MaxAttempts <= 0 || request.BaseBackoff < 0 || request.MaxBackoff < request.BaseBackoff ||
		request.MaxBackoff <= 0 || !validSHA256(request.FailureHash) || request.Now.IsZero() ||
		strings.TrimSpace(request.Diagnostic) == "" || len(request.Diagnostic) > 256 ||
		strings.ContainsAny(request.Diagnostic, "\r\n\x00") {
		return errors.New("complete bounded recovery reservation is required")
	}
	class, err := recovery.ParseFailureClass(request.FailureClass)
	if err != nil {
		return err
	}
	policy, err := recovery.PolicyFor(recovery.Failure{Class: class, Reversible: request.Reversible})
	if err != nil {
		return err
	}
	return policy.ValidateAction(request.ExhaustedAction)
}

func validateRecoveryOutcome(outcome RecoveryOutcome) error {
	if err := validateRecoveryBudgetKey(outcome.RecoveryBudgetKey); err != nil {
		return err
	}
	validIdentity := outcome.UnitAttempt > 0 && validRecoveryToken(outcome.ClaimToken) &&
		strings.TrimSpace(outcome.ControllerOwner) != "" && !strings.ContainsAny(outcome.ControllerOwner, "\r\n\x00") && outcome.ControllerEpoch > 0 &&
		validRecoveryToken(outcome.PlannerRequestNonce) && validSHA256(outcome.EvidenceHash) &&
		validSHA256(outcome.AuthorityScopeHash) && validSHA256(outcome.PlannerEvidenceHash) &&
		len(outcome.PlannerSessionID) == 36 && validSHA256(outcome.PlannerSessionFingerprint) &&
		outcome.ObservedModel == recovery.RequiredModel && outcome.Thinking == recovery.RequiredThinking
	if !validIdentity || outcome.IssuedAt.IsZero() || outcome.ExpiresAt.Before(outcome.IssuedAt) ||
		outcome.ExpiresAt.Sub(outcome.IssuedAt) > 10*time.Minute {
		return errors.New("complete bounded recovery planner evidence is required")
	}
	if _, err := recovery.ParseAction(string(outcome.SelectedAction)); err != nil {
		return err
	}
	return recovery.ValidatePlan(outcome.SelectedAction, outcome.BoundedPlan)
}

func validRecoveryToken(value string) bool {
	if len(value) != 32 {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

type contextExecer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func (s *Store) PutArtifactProof(ctx context.Context, proof ArtifactProof) error {
	return putArtifactProof(ctx, s.db, proof)
}

func putArtifactProof(ctx context.Context, executor contextExecer, proof ArtifactProof) error {
	if proof.ProofID == "" || proof.DeliveryID == "" || proof.Generation <= 0 || strings.TrimSpace(proof.UnitID) == "" ||
		proof.Attempt <= 0 || !validGitSHA(proof.StartHead) || !validGitSHA(proof.CandidateHead) ||
		!validGitSHA(proof.ValidatedHead) || proof.CandidateHead != proof.ValidatedHead ||
		strings.TrimSpace(proof.ExpectedArtifact) == "" || !validSHA256(proof.ArtifactHash) ||
		proof.Validator != "openai-codex/gpt-5.6-sol" || proof.Thinking != "high" || !proof.Ratified {
		return errors.New("complete exact-head artifact proof is required")
	}
	ratified := 0
	if proof.Ratified {
		ratified = 1
	}
	_, err := executor.ExecContext(ctx, `INSERT INTO artifact_proofs
		(proof_id, delivery_id, generation, unit_id, attempt, start_head, candidate_head, validated_head,
		 expected_artifact, artifact_hash, validator, thinking, ratified, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, proof.ProofID, proof.DeliveryID,
		proof.Generation, proof.UnitID, proof.Attempt, proof.StartHead, proof.CandidateHead,
		proof.ValidatedHead, proof.ExpectedArtifact, proof.ArtifactHash, proof.Validator, proof.Thinking,
		ratified, time.Now().UTC().UnixNano())
	if err != nil {
		return fmt.Errorf("put artifact proof: %w", err)
	}
	return nil
}

func (s *Store) GetArtifactProof(ctx context.Context, proofID string) (ArtifactProof, error) {
	var proof ArtifactProof
	var ratified int
	err := s.db.QueryRowContext(ctx, `SELECT proof_id, delivery_id, generation, unit_id, attempt,
		start_head, candidate_head, validated_head, expected_artifact, artifact_hash, validator, thinking, ratified
		FROM artifact_proofs WHERE proof_id = ?`, proofID).Scan(&proof.ProofID, &proof.DeliveryID,
		&proof.Generation, &proof.UnitID, &proof.Attempt, &proof.StartHead, &proof.CandidateHead,
		&proof.ValidatedHead, &proof.ExpectedArtifact, &proof.ArtifactHash, &proof.Validator,
		&proof.Thinking, &ratified)
	if err != nil {
		return ArtifactProof{}, fmt.Errorf("read artifact proof: %w", err)
	}
	proof.Ratified = ratified == 1
	return proof, nil
}

func (s *Store) PutAttestation(ctx context.Context, attestation AttestationRecord) error {
	return putAttestation(ctx, s.db, attestation)
}

func putAttestation(ctx context.Context, executor contextExecer, attestation AttestationRecord) error {
	if attestation.Repository == "" || attestation.PR <= 0 || attestation.BaseBranch == "" || !validGitSHA(attestation.BaseHead) ||
		!validGitSHA(attestation.CandidateHead) || !validGitSHA(attestation.ObservedHead) || attestation.RunID == "" ||
		attestation.Generation <= 0 || strings.TrimSpace(attestation.UnitID) == "" || attestation.Attempt <= 0 ||
		attestation.StateVersion <= 0 || !validSHA256(attestation.ContractHash) || !validSHA256(attestation.EvidenceHash) ||
		strings.TrimSpace(attestation.ValidatorSessionID) == "" || !validGitSHA(attestation.HeadSHA) ||
		attestation.HeadSHA != attestation.CandidateHead || attestation.ObservedHead != attestation.CandidateHead ||
		attestation.Validator != "openai-codex/gpt-5.6-sol" || attestation.Thinking != "high" || attestation.Verdict != "PROCEED" ||
		attestation.CreatedAt.IsZero() || attestation.ExpiresAt.IsZero() || !attestation.ExpiresAt.After(attestation.CreatedAt) {
		return errors.New("complete attestation identity is required")
	}
	created := attestation.CreatedAt.UTC()
	expires := attestation.ExpiresAt.UTC()
	localGates, uat, milestoneValid := 0, 0, 0
	if attestation.LocalGates {
		localGates = 1
	}
	if attestation.UAT {
		uat = 1
	}
	if attestation.MilestoneValid {
		milestoneValid = 1
	}
	result, err := executor.ExecContext(ctx, `INSERT INTO attestations(
		run_id, head_sha, validator, thinking, verdict, created_at, repository, pr, base_branch,
		base_head, candidate_head, observed_head, generation, unit_id, attempt, state_version,
		contract_hash, evidence_hash, validator_session_id, local_gates, uat, milestone_valid, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(run_id, head_sha) DO UPDATE SET head_sha = excluded.head_sha
		WHERE validator = excluded.validator AND thinking = excluded.thinking AND verdict = excluded.verdict
		AND created_at = excluded.created_at AND repository = excluded.repository AND pr = excluded.pr
		AND base_branch = excluded.base_branch AND base_head = excluded.base_head
		AND candidate_head = excluded.candidate_head AND observed_head = excluded.observed_head
		AND generation = excluded.generation AND unit_id = excluded.unit_id AND attempt = excluded.attempt
		AND state_version = excluded.state_version AND contract_hash = excluded.contract_hash
		AND evidence_hash = excluded.evidence_hash AND validator_session_id = excluded.validator_session_id
		AND local_gates = excluded.local_gates AND uat = excluded.uat
		AND milestone_valid = excluded.milestone_valid AND expires_at = excluded.expires_at`, attestation.RunID, attestation.HeadSHA, attestation.Validator,
		attestation.Thinking, attestation.Verdict, created.UnixNano(), attestation.Repository, attestation.PR,
		attestation.BaseBranch, attestation.BaseHead, attestation.CandidateHead, attestation.ObservedHead,
		attestation.Generation, attestation.UnitID, attestation.Attempt, attestation.StateVersion,
		attestation.ContractHash, attestation.EvidenceHash, attestation.ValidatorSessionID,
		localGates, uat, milestoneValid, expires.UnixNano())
	if err != nil {
		return fmt.Errorf("put attestation: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return errors.New("attestation identity is immutable and cannot be rebound")
	}
	return nil
}

func (s *Store) GetAttestation(ctx context.Context, runID, headSHA string) (AttestationRecord, error) {
	var attestation AttestationRecord
	var created, expires int64
	var localGates, uat, milestoneValid int
	err := s.db.QueryRowContext(ctx, `SELECT run_id, head_sha, validator, thinking, verdict, created_at,
		repository, pr, base_branch, base_head, candidate_head, observed_head, generation, unit_id,
		attempt, state_version, contract_hash, evidence_hash, validator_session_id, local_gates,
		uat, milestone_valid, expires_at FROM attestations WHERE run_id = ? AND head_sha = ?`, runID, headSHA).Scan(&attestation.RunID,
		&attestation.HeadSHA, &attestation.Validator, &attestation.Thinking, &attestation.Verdict, &created,
		&attestation.Repository, &attestation.PR, &attestation.BaseBranch, &attestation.BaseHead,
		&attestation.CandidateHead, &attestation.ObservedHead, &attestation.Generation, &attestation.UnitID,
		&attestation.Attempt, &attestation.StateVersion, &attestation.ContractHash, &attestation.EvidenceHash,
		&attestation.ValidatorSessionID, &localGates, &uat, &milestoneValid, &expires)
	if err != nil {
		return AttestationRecord{}, fmt.Errorf("read attestation: %w", err)
	}
	attestation.CreatedAt = time.Unix(0, created).UTC()
	attestation.ExpiresAt = time.Unix(0, expires).UTC()
	attestation.LocalGates = localGates == 1
	attestation.UAT = uat == 1
	attestation.MilestoneValid = milestoneValid == 1
	return attestation, nil
}

func validateDecisionRequest(request DecisionRequest) error {
	if request.RequestID == "" || request.DeliveryID == "" || !validRepository(request.Repository) ||
		request.Issue <= 0 || request.PullRequest <= 0 ||
		strings.TrimSpace(request.UnitID) == "" || request.Generation <= 0 || !validGitSHA(request.HeadSHA) ||
		strings.TrimSpace(request.Kind) == "" || strings.TrimSpace(request.Evidence) == "" || len(request.Options) == 0 ||
		request.ExpiresAt.IsZero() {
		return errors.New("complete durable decision request identity is required")
	}
	if request.Status == "" {
		return errors.New("decision request status is required")
	}
	if err := sensitive.ValidatePublicIdentifier(request.UnitID); err != nil {
		return fmt.Errorf("decision request unit: %w", err)
	}
	if err := sensitive.ValidatePublicIdentifier(request.Kind); err != nil {
		return fmt.Errorf("decision request kind: %w", err)
	}
	if err := sensitive.ValidatePublicText(request.Evidence); err != nil {
		return fmt.Errorf("decision request evidence: %w", err)
	}
	allowedOptions := make(map[string]struct{}, len(request.Options))
	for _, option := range request.Options {
		if strings.TrimSpace(option) == "" || strings.ContainsAny(option, "\r\n\x00") ||
			sensitive.ValidatePublicIdentifier(option) != nil {
			return errors.New("decision request options must be bounded single-line values")
		}
		normalized := strings.ToLower(option)
		if _, duplicate := allowedOptions[normalized]; duplicate {
			return errors.New("decision request options contain a duplicate")
		}
		allowedOptions[normalized] = struct{}{}
	}
	for _, option := range []string{request.RecommendedOption, request.SafeDefault} {
		if option != "" {
			if sensitive.ValidatePublicIdentifier(option) != nil {
				return errors.New("decision request policy option is unsafe")
			}
			if _, allowed := allowedOptions[strings.ToLower(option)]; !allowed {
				return errors.New("decision request policy option is not allowlisted")
			}
		}
	}
	return nil
}

func validRepository(repository string) bool {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return false
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return false
		}
		for _, character := range part {
			if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
				(character >= '0' && character <= '9') || strings.ContainsRune("._-", character) {
				continue
			}
			return false
		}
	}
	return true
}

func validateRecoveryBudgetKey(key RecoveryBudgetKey) error {
	if key.DeliveryID == "" || key.Generation <= 0 || strings.TrimSpace(key.UnitID) == "" ||
		!validGitSHA(key.HeadSHA) || strings.TrimSpace(key.FailureClass) == "" || strings.ContainsAny(key.FailureClass, "\r\n\x00") {
		return errors.New("complete recovery budget identity is required")
	}
	_, err := recovery.ParseFailureClass(key.FailureClass)
	return err
}

func boundedStoreString(value string, limit int) string {
	value = strings.TrimSpace(strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, value))
	if len(value) > limit {
		return value[:limit]
	}
	return value
}

func jsonMarshalStrings(values []string) (string, error) {
	raw, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func jsonUnmarshalStrings(raw string) ([]string, error) {
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, fmt.Errorf("decode decision request options: %w", err)
	}
	return values, nil
}

func validSHA256(value string) bool {
	if len(value) != len("sha256:")+64 || value[:len("sha256:")] != "sha256:" {
		return false
	}
	for _, char := range value[len("sha256:"):] {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}

func validGitSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}

func (s *Store) AcquireLease(ctx context.Context, runID, owner string, now time.Time, ttl time.Duration) (Lease, error) {
	if runID == "" || owner == "" || ttl <= 0 {
		return Lease{}, errors.New("run, owner, and positive ttl are required")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return Lease{}, err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return Lease{}, fmt.Errorf("begin lease transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()

	var currentOwner string
	var epoch, expiresAt int64
	err = conn.QueryRowContext(ctx, `SELECT owner, epoch, expires_at FROM leases WHERE run_id = ?`, runID).
		Scan(&currentOwner, &epoch, &expiresAt)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		epoch = 1
	case err != nil:
		return Lease{}, fmt.Errorf("read lease: %w", err)
	case now.UnixNano() < expiresAt && currentOwner != owner:
		return Lease{}, errors.New("lease is held by another live owner")
	default:
		epoch++
	}
	expires := now.Add(ttl).UTC()
	if _, err := conn.ExecContext(ctx, `INSERT INTO leases(run_id, owner, epoch, expires_at)
        VALUES (?, ?, ?, ?) ON CONFLICT(run_id) DO UPDATE SET owner=excluded.owner,
        epoch=excluded.epoch, expires_at=excluded.expires_at`, runID, owner, epoch, expires.UnixNano()); err != nil {
		return Lease{}, fmt.Errorf("write lease: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return Lease{}, err
	}
	committed = true
	return Lease{RunID: runID, Owner: owner, Epoch: epoch, ExpiresAt: expires}, nil
}

// AcquireReconciliationLease immediately fences the prior controller. Callers
// must hold the repository-global controller lock, which proves the previous
// Shepherd process no longer owns the repository mutation boundary.
func (s *Store) AcquireReconciliationLease(ctx context.Context, runID, owner string, now time.Time, ttl time.Duration) (Lease, error) {
	if runID == "" || owner == "" || ttl <= 0 {
		return Lease{}, errors.New("run, owner, and positive reconciliation TTL are required")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return Lease{}, err
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return Lease{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var epoch int64
	err = conn.QueryRowContext(ctx, `SELECT epoch FROM leases WHERE run_id = ?`, runID).Scan(&epoch)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Lease{}, err
	}
	epoch++
	expires := now.Add(ttl).UTC()
	if _, err := conn.ExecContext(ctx, `INSERT INTO leases(run_id, owner, epoch, expires_at)
		VALUES (?, ?, ?, ?) ON CONFLICT(run_id) DO UPDATE SET owner=excluded.owner,
		epoch=excluded.epoch, expires_at=excluded.expires_at`, runID, owner, epoch, expires.UnixNano()); err != nil {
		return Lease{}, err
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return Lease{}, err
	}
	committed = true
	return Lease{RunID: runID, Owner: owner, Epoch: epoch, ExpiresAt: expires}, nil
}

func (s *Store) CheckLease(ctx context.Context, lease Lease, now time.Time) error {
	var owner string
	var epoch, expiresAt int64
	if err := s.db.QueryRowContext(ctx, `SELECT owner, epoch, expires_at FROM leases WHERE run_id = ?`, lease.RunID).
		Scan(&owner, &epoch, &expiresAt); err != nil {
		return fmt.Errorf("read lease: %w", err)
	}
	if owner != lease.Owner || epoch != lease.Epoch || now.UnixNano() >= expiresAt {
		return errors.New("lease is stale, expired, or fenced")
	}
	return nil
}

func (s *Store) ReleaseLease(ctx context.Context, lease Lease) error {
	result, err := s.db.ExecContext(ctx, `UPDATE leases SET expires_at = 0
        WHERE run_id = ? AND owner = ? AND epoch = ?`, lease.RunID, lease.Owner, lease.Epoch)
	if err != nil {
		return fmt.Errorf("release lease: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return errors.New("cannot release a stale or fenced lease")
	}
	return nil
}
