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

type Effect struct {
	Key         string
	RunID       string
	Repository  string
	Issue       int
	Capability  domain.Capability
	Target      string
	PayloadHash string
	Epoch       int64
}

type Delivery struct {
	ID             string
	Issue          int
	ParentIssue    int
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

var ErrRetryBudgetExhausted = errors.New("unit retry budget exhausted")

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
	Attempts     int64
	MaxAttempts  int64
	Backoff      time.Duration
	LastFailure  string
	RecoveryPlan string
	NextRetryAt  time.Time
	ExhaustedAt  time.Time
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
	RunID     string
	HeadSHA   string
	Validator string
	Thinking  string
	Verdict   string
	CreatedAt time.Time
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
		`CREATE TABLE IF NOT EXISTS leases (
            run_id TEXT PRIMARY KEY, owner TEXT NOT NULL, epoch INTEGER NOT NULL,
            expires_at INTEGER NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS deliveries (
            delivery_id TEXT PRIMARY KEY, issue INTEGER NOT NULL, work_dir TEXT NOT NULL,
			context_hash TEXT NOT NULL, milestone_id TEXT NOT NULL DEFAULT '',
			parent_issue INTEGER NOT NULL DEFAULT 0, branch TEXT NOT NULL DEFAULT '',
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
		`CREATE TABLE IF NOT EXISTS grants (
            run_id TEXT NOT NULL, repository TEXT NOT NULL, issue INTEGER NOT NULL,
            capability TEXT NOT NULL, epoch INTEGER NOT NULL,
            PRIMARY KEY (run_id, repository, issue, capability, epoch)
        )`,
		`CREATE TABLE IF NOT EXISTS attestations (
            run_id TEXT NOT NULL, head_sha TEXT NOT NULL, validator TEXT NOT NULL,
            thinking TEXT NOT NULL, verdict TEXT NOT NULL, created_at INTEGER NOT NULL,
            PRIMARY KEY (run_id, head_sha)
        )`,
		`CREATE TABLE IF NOT EXISTS outbox (
			effect_key TEXT PRIMARY KEY, run_id TEXT NOT NULL, repository TEXT NOT NULL,
			issue INTEGER NOT NULL, capability TEXT NOT NULL, target TEXT NOT NULL,
			payload_hash TEXT NOT NULL, epoch INTEGER NOT NULL, status TEXT NOT NULL DEFAULT 'pending',
            created_at INTEGER NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS unit_attempts (
			delivery_id TEXT NOT NULL REFERENCES deliveries(delivery_id), generation INTEGER NOT NULL,
			unit_id TEXT NOT NULL, head_sha TEXT NOT NULL, attempts INTEGER NOT NULL,
			max_attempts INTEGER NOT NULL, status TEXT NOT NULL, last_failure TEXT NOT NULL DEFAULT '',
			updated_at INTEGER NOT NULL,
			PRIMARY KEY (delivery_id, generation, unit_id, head_sha)
		)`,
		`CREATE TABLE IF NOT EXISTS decision_requests (
			request_id TEXT PRIMARY KEY, delivery_id TEXT NOT NULL REFERENCES deliveries(delivery_id),
			issue INTEGER NOT NULL, pull_request INTEGER NOT NULL, unit_id TEXT NOT NULL,
			generation INTEGER NOT NULL, head_sha TEXT NOT NULL, kind TEXT NOT NULL,
			evidence TEXT NOT NULL, options_json TEXT NOT NULL, recommended_option TEXT NOT NULL DEFAULT '',
			safe_default TEXT NOT NULL DEFAULT '', expires_at INTEGER NOT NULL,
			github_comment_id INTEGER NOT NULL DEFAULT 0, status TEXT NOT NULL,
			accepted_answer TEXT NOT NULL DEFAULT '', accepted_by TEXT NOT NULL DEFAULT '',
			accepted_at INTEGER NOT NULL DEFAULT 0, consumed_at INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS recovery_budgets (
			delivery_id TEXT NOT NULL REFERENCES deliveries(delivery_id), generation INTEGER NOT NULL,
			unit_id TEXT NOT NULL, head_sha TEXT NOT NULL, failure_class TEXT NOT NULL,
			attempts INTEGER NOT NULL, max_attempts INTEGER NOT NULL, backoff_ms INTEGER NOT NULL,
			last_failure TEXT NOT NULL DEFAULT '', recovery_plan TEXT NOT NULL DEFAULT '',
			next_retry_at INTEGER NOT NULL DEFAULT 0, exhausted_at INTEGER NOT NULL DEFAULT 0,
			updated_at INTEGER NOT NULL,
			PRIMARY KEY (delivery_id, generation, unit_id, head_sha, failure_class)
		)`,
		`CREATE TABLE IF NOT EXISTS artifact_proofs (
			proof_id TEXT PRIMARY KEY, delivery_id TEXT NOT NULL REFERENCES deliveries(delivery_id),
			generation INTEGER NOT NULL, unit_id TEXT NOT NULL, attempt INTEGER NOT NULL,
			start_head TEXT NOT NULL, candidate_head TEXT NOT NULL, validated_head TEXT NOT NULL,
			expected_artifact TEXT NOT NULL, artifact_hash TEXT NOT NULL, validator TEXT NOT NULL,
			thinking TEXT NOT NULL, ratified INTEGER NOT NULL, created_at INTEGER NOT NULL
		)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply supervisor migration: %w", err)
		}
	}
	columns := map[string]string{
		"parent_issue": "INTEGER NOT NULL DEFAULT 0", "branch": "TEXT NOT NULL DEFAULT ''",
		"base_branch": "TEXT NOT NULL DEFAULT ''", "gsd_project_root": "TEXT NOT NULL DEFAULT ''",
		"initial_head": "TEXT NOT NULL DEFAULT ''", "gsd_version": "TEXT NOT NULL DEFAULT ''",
	}
	for name, definition := range columns {
		if err := s.ensureColumn(ctx, "deliveries", name, definition); err != nil {
			return err
		}
	}
	for name, definition := range map[string]string{"claimed_by": "TEXT NOT NULL DEFAULT ''", "claimed_at": "INTEGER NOT NULL DEFAULT 0", "sent_at": "INTEGER NOT NULL DEFAULT 0", "last_error": "TEXT NOT NULL DEFAULT ''"} {
		if err := s.ensureColumn(ctx, "outbox", name, definition); err != nil {
			return err
		}
	}
	for name, definition := range map[string]string{"start_head": "TEXT NOT NULL DEFAULT ''", "end_head": "TEXT NOT NULL DEFAULT ''"} {
		if err := s.ensureColumn(ctx, "delivery_runs", name, definition); err != nil {
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

func (s *Store) BeginUnitAttempt(ctx context.Context, key UnitAttemptKey, maxAttempts int64) (UnitAttempt, error) {
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
	if configuredMax != maxAttempts {
		return UnitAttempt{}, errors.New("unit attempt budget cannot change within a generation")
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

func (s *Store) FinishUnitAttempt(ctx context.Context, key UnitAttemptKey, outcome string) error {
	outcome = strings.TrimSpace(outcome)
	if outcome == "" || strings.ContainsAny(outcome, "\r\n\x00") {
		return errors.New("bounded unit attempt outcome is required")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE unit_attempts SET status = 'terminal',
		last_failure = ?, updated_at = ? WHERE delivery_id = ? AND generation = ? AND unit_id = ?
		AND head_sha = ? AND status = 'running'`, outcome, time.Now().UTC().UnixNano(),
		key.DeliveryID, key.Generation, key.UnitID, key.HeadSHA)
	if err != nil {
		return fmt.Errorf("finish unit attempt: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return errors.New("unit attempt is not running or is fenced")
	}
	return nil
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
		(delivery_id, issue, parent_issue, work_dir, context_hash, milestone_id, branch, base_branch,
		 gsd_project_root, initial_head, gsd_version, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, delivery.ID, delivery.Issue,
		delivery.ParentIssue, delivery.WorkDir, delivery.ContextHash, delivery.MilestoneID, delivery.Branch,
		delivery.BaseBranch, delivery.GSDProjectRoot, delivery.InitialHead, delivery.GSDVersion, now, now)
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
	if current.Issue != delivery.Issue || current.ParentIssue != delivery.ParentIssue ||
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
		!filepath.IsAbs(delivery.WorkDir) || !filepath.IsAbs(delivery.GSDProjectRoot) ||
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
	err := s.db.QueryRowContext(ctx, `SELECT delivery_id, issue, parent_issue, work_dir, context_hash,
		milestone_id, branch, base_branch, gsd_project_root, initial_head, gsd_version
		FROM deliveries WHERE delivery_id = ?`, id).Scan(&delivery.ID, &delivery.Issue,
		&delivery.ParentIssue, &delivery.WorkDir, &delivery.ContextHash, &delivery.MilestoneID,
		&delivery.Branch, &delivery.BaseBranch, &delivery.GSDProjectRoot, &delivery.InitialHead,
		&delivery.GSDVersion)
	if err != nil {
		return Delivery{}, fmt.Errorf("read delivery: %w", err)
	}
	return delivery, nil
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
	optionsRaw, err := jsonMarshalStrings(request.Options)
	if err != nil {
		return DecisionRequest{}, err
	}
	now := time.Now().UTC().UnixNano()
	_, err = s.db.ExecContext(ctx, `INSERT INTO decision_requests
		(request_id, delivery_id, issue, pull_request, unit_id, generation, head_sha, kind,
		 evidence, options_json, recommended_option, safe_default, expires_at, status,
		 created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(request_id) DO NOTHING`, request.RequestID, request.DeliveryID, request.Issue,
		request.PullRequest, request.UnitID, request.Generation, request.HeadSHA, request.Kind,
		request.Evidence, optionsRaw, request.RecommendedOption, request.SafeDefault,
		request.ExpiresAt.UTC().UnixNano(), request.Status, now, now)
	if err != nil {
		return DecisionRequest{}, fmt.Errorf("upsert decision request: %w", err)
	}
	existing, err := s.GetDecisionRequest(ctx, request.RequestID)
	if err != nil {
		return DecisionRequest{}, err
	}
	if existing.DeliveryID != request.DeliveryID || existing.Generation != request.Generation ||
		existing.HeadSHA != request.HeadSHA || existing.UnitID != request.UnitID || existing.Kind != request.Kind ||
		strings.Join(existing.Options, "\x00") != strings.Join(request.Options, "\x00") {
		return DecisionRequest{}, errors.New("decision request id collides with different identity")
	}
	return existing, nil
}

func (s *Store) ListOpenDecisionRequests(ctx context.Context, deliveryID string) ([]DecisionRequest, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT request_id FROM decision_requests WHERE delivery_id = ? AND status IN (?, ?) ORDER BY created_at`, deliveryID, DecisionRequestOpen, DecisionRequestPublished)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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
	err := s.db.QueryRowContext(ctx, `SELECT request_id, delivery_id, issue, pull_request, unit_id,
		generation, head_sha, kind, evidence, options_json, recommended_option, safe_default,
		expires_at, github_comment_id, status, accepted_answer, accepted_by, accepted_at, consumed_at
		FROM decision_requests WHERE request_id = ?`, requestID).Scan(&request.RequestID, &request.DeliveryID,
		&request.Issue, &request.PullRequest, &request.UnitID, &request.Generation, &request.HeadSHA,
		&request.Kind, &request.Evidence, &optionsRaw, &request.RecommendedOption, &request.SafeDefault,
		&expiresAt, &request.GitHubCommentID, &request.Status, &request.AcceptedAnswer, &request.AcceptedBy,
		&acceptedAt, &consumedAt)
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

func (s *Store) MarkDecisionRequestPublished(ctx context.Context, requestID string, commentID int64) error {
	if requestID == "" || commentID <= 0 {
		return errors.New("decision request and comment identity are required")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE decision_requests SET status = ?, github_comment_id = ?, updated_at = ?
		WHERE request_id = ? AND status IN (?, ?)`, DecisionRequestPublished, commentID,
		time.Now().UTC().UnixNano(), requestID, DecisionRequestOpen, DecisionRequestPublished)
	if err != nil {
		return fmt.Errorf("mark decision request published: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return errors.New("decision request cannot be marked published")
	}
	return nil
}

func (s *Store) AcceptDecisionRequestAnswer(ctx context.Context, requestID, answer, actor string, generation int64, headSHA string, now time.Time) error {
	answer = strings.TrimSpace(answer)
	if requestID == "" || answer == "" || strings.TrimSpace(actor) == "" || generation <= 0 || !validGitSHA(headSHA) {
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
	if !allowed {
		return errors.New("decision answer is not one of the bounded options")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE decision_requests SET status = ?, accepted_answer = ?,
		accepted_by = ?, accepted_at = ?, updated_at = ? WHERE request_id = ? AND generation = ? AND head_sha = ?
		AND status IN (?, ?) AND consumed_at = 0 AND expires_at > ?`, DecisionRequestAnswered,
		strings.TrimSpace(answer), strings.TrimSpace(actor), now.UTC().UnixNano(), now.UTC().UnixNano(),
		requestID, generation, headSHA, DecisionRequestOpen, DecisionRequestPublished, now.UTC().UnixNano())
	if err != nil {
		return fmt.Errorf("accept decision request answer: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return errors.New("decision answer is stale, duplicate, expired, or unauthorized for this generation/head")
	}
	return nil
}

func (s *Store) ConsumeDecisionRequest(ctx context.Context, requestID string) (DecisionRequest, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return DecisionRequest{}, err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return DecisionRequest{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var status DecisionRequestStatus
	if err := conn.QueryRowContext(ctx, `SELECT status FROM decision_requests WHERE request_id = ?`, requestID).Scan(&status); err != nil {
		return DecisionRequest{}, fmt.Errorf("read decision request status: %w", err)
	}
	if status != DecisionRequestAnswered {
		return DecisionRequest{}, errors.New("decision request has no unconsumed accepted answer")
	}
	if _, err := conn.ExecContext(ctx, `UPDATE decision_requests SET status = ?, consumed_at = ?, updated_at = ?
		WHERE request_id = ? AND status = ? AND consumed_at = 0`, DecisionRequestConsumed,
		time.Now().UTC().UnixNano(), time.Now().UTC().UnixNano(), requestID, DecisionRequestAnswered); err != nil {
		return DecisionRequest{}, err
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return DecisionRequest{}, err
	}
	committed = true
	if err := conn.Close(); err != nil {
		return DecisionRequest{}, err
	}
	return s.GetDecisionRequest(ctx, requestID)
}

func (s *Store) BeginRecoveryAttempt(ctx context.Context, key RecoveryBudgetKey, maxAttempts int64, backoff time.Duration, failure, recoveryPlan string, now time.Time) (RecoveryBudget, error) {
	if err := validateRecoveryBudgetKey(key); err != nil || maxAttempts <= 0 || backoff < 0 {
		if err != nil {
			return RecoveryBudget{}, err
		}
		return RecoveryBudget{}, errors.New("positive recovery budget is required")
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
	stamp := now.UTC().UnixNano()
	if _, err := conn.ExecContext(ctx, `INSERT OR IGNORE INTO recovery_budgets
		(delivery_id, generation, unit_id, head_sha, failure_class, attempts, max_attempts, backoff_ms, updated_at)
		VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?)`, key.DeliveryID, key.Generation, key.UnitID,
		key.HeadSHA, key.FailureClass, maxAttempts, backoff.Milliseconds(), stamp); err != nil {
		return RecoveryBudget{}, fmt.Errorf("initialize recovery budget: %w", err)
	}
	var budget RecoveryBudget
	budget.RecoveryBudgetKey = key
	var backoffMS, nextRetryAt, exhaustedAt int64
	if err := conn.QueryRowContext(ctx, `SELECT attempts, max_attempts, backoff_ms, last_failure,
		recovery_plan, next_retry_at, exhausted_at FROM recovery_budgets WHERE delivery_id = ? AND generation = ?
		AND unit_id = ? AND head_sha = ? AND failure_class = ?`, key.DeliveryID, key.Generation,
		key.UnitID, key.HeadSHA, key.FailureClass).Scan(&budget.Attempts, &budget.MaxAttempts, &backoffMS,
		&budget.LastFailure, &budget.RecoveryPlan, &nextRetryAt, &exhaustedAt); err != nil {
		return RecoveryBudget{}, err
	}
	if budget.MaxAttempts != maxAttempts || time.Duration(backoffMS)*time.Millisecond != backoff {
		return RecoveryBudget{}, errors.New("recovery budget cannot change within a generation")
	}
	if budget.Attempts >= budget.MaxAttempts {
		if exhaustedAt == 0 {
			exhaustedAt = stamp
			if _, err := conn.ExecContext(ctx, `UPDATE recovery_budgets SET exhausted_at = ?, updated_at = ? WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ? AND failure_class = ?`, exhaustedAt, stamp, key.DeliveryID, key.Generation, key.UnitID, key.HeadSHA, key.FailureClass); err != nil {
				return RecoveryBudget{}, err
			}
		}
		if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
			return RecoveryBudget{}, err
		}
		committed = true
		return RecoveryBudget{}, ErrRetryBudgetExhausted
	}
	budget.Attempts++
	budget.LastFailure = boundedStoreString(failure, 512)
	budget.RecoveryPlan = boundedStoreString(recoveryPlan, 2048)
	budget.NextRetryAt = now.Add(backoff).UTC()
	if _, err := conn.ExecContext(ctx, `UPDATE recovery_budgets SET attempts = ?, last_failure = ?, recovery_plan = ?,
		next_retry_at = ?, updated_at = ? WHERE delivery_id = ? AND generation = ? AND unit_id = ? AND head_sha = ?
		AND failure_class = ?`, budget.Attempts, budget.LastFailure, budget.RecoveryPlan,
		budget.NextRetryAt.UnixNano(), stamp, key.DeliveryID, key.Generation, key.UnitID, key.HeadSHA,
		key.FailureClass); err != nil {
		return RecoveryBudget{}, err
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return RecoveryBudget{}, err
	}
	committed = true
	budget.Backoff = backoff
	return budget, nil
}

func (s *Store) PutArtifactProof(ctx context.Context, proof ArtifactProof) error {
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
	_, err := s.db.ExecContext(ctx, `INSERT INTO artifact_proofs
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
	if attestation.RunID == "" || !validGitSHA(attestation.HeadSHA) || attestation.Validator != "openai-codex/gpt-5.6-sol" ||
		attestation.Thinking != "high" || attestation.Verdict != "PROCEED" {
		return errors.New("complete attestation identity is required")
	}
	created := attestation.CreatedAt.UTC()
	if created.IsZero() {
		created = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO attestations(run_id, head_sha, validator, thinking, verdict, created_at)
		VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(run_id, head_sha) DO UPDATE SET
		validator = excluded.validator, thinking = excluded.thinking, verdict = excluded.verdict,
		created_at = excluded.created_at`, attestation.RunID, attestation.HeadSHA, attestation.Validator,
		attestation.Thinking, attestation.Verdict, created.UnixNano())
	if err != nil {
		return fmt.Errorf("put attestation: %w", err)
	}
	return nil
}

func (s *Store) GetAttestation(ctx context.Context, runID, headSHA string) (AttestationRecord, error) {
	var attestation AttestationRecord
	var created int64
	err := s.db.QueryRowContext(ctx, `SELECT run_id, head_sha, validator, thinking, verdict, created_at
		FROM attestations WHERE run_id = ? AND head_sha = ?`, runID, headSHA).Scan(&attestation.RunID,
		&attestation.HeadSHA, &attestation.Validator, &attestation.Thinking, &attestation.Verdict, &created)
	if err != nil {
		return AttestationRecord{}, fmt.Errorf("read attestation: %w", err)
	}
	attestation.CreatedAt = time.Unix(0, created).UTC()
	return attestation, nil
}

func validateDecisionRequest(request DecisionRequest) error {
	if request.RequestID == "" || request.DeliveryID == "" || request.Issue <= 0 || request.PullRequest <= 0 ||
		strings.TrimSpace(request.UnitID) == "" || request.Generation <= 0 || !validGitSHA(request.HeadSHA) ||
		strings.TrimSpace(request.Kind) == "" || strings.TrimSpace(request.Evidence) == "" || len(request.Options) == 0 ||
		request.ExpiresAt.IsZero() {
		return errors.New("complete durable decision request identity is required")
	}
	if request.Status == "" {
		return errors.New("decision request status is required")
	}
	for _, option := range request.Options {
		if strings.TrimSpace(option) == "" || strings.ContainsAny(option, "\r\n\x00") {
			return errors.New("decision request options must be bounded single-line values")
		}
	}
	return nil
}

func validateRecoveryBudgetKey(key RecoveryBudgetKey) error {
	if key.DeliveryID == "" || key.Generation <= 0 || strings.TrimSpace(key.UnitID) == "" ||
		!validGitSHA(key.HeadSHA) || strings.TrimSpace(key.FailureClass) == "" || strings.ContainsAny(key.FailureClass, "\r\n\x00") {
		return errors.New("complete recovery budget identity is required")
	}
	return nil
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

func (s *Store) PutGrant(ctx context.Context, grant domain.Grant) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO grants
        (run_id, repository, issue, capability, epoch) VALUES (?, ?, ?, ?, ?)`,
		grant.RunID, grant.Repository, grant.Issue, grant.Capability, grant.Epoch)
	if err != nil {
		return fmt.Errorf("put grant: %w", err)
	}
	return nil
}

func (s *Store) ClaimEffect(ctx context.Context, lease Lease, key string, now time.Time) (Effect, error) {
	if key == "" {
		return Effect{}, errors.New("effect key is required")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return Effect{}, err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return Effect{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var owner string
	var epoch, expiresAt int64
	if err := conn.QueryRowContext(ctx, `SELECT owner, epoch, expires_at FROM leases WHERE run_id = ?`, lease.RunID).Scan(&owner, &epoch, &expiresAt); err != nil {
		return Effect{}, fmt.Errorf("read fenced lease: %w", err)
	}
	if owner != lease.Owner || epoch != lease.Epoch || now.UnixNano() >= expiresAt {
		return Effect{}, errors.New("lease is stale, expired, or fenced")
	}
	result, err := conn.ExecContext(ctx, `UPDATE outbox SET status = 'claimed', claimed_by = ?, claimed_at = ?, last_error = ''
		WHERE effect_key = ? AND run_id = ? AND epoch = ? AND status = 'pending'`, lease.Owner, now.UTC().UnixNano(), key, lease.RunID, lease.Epoch)
	if err != nil {
		return Effect{}, fmt.Errorf("claim effect: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return Effect{}, errors.New("effect is not pending for this lease")
	}
	var effect Effect
	if err := conn.QueryRowContext(ctx, `SELECT effect_key, run_id, repository, issue, capability, target, payload_hash, epoch
		FROM outbox WHERE effect_key = ?`, key).Scan(&effect.Key, &effect.RunID, &effect.Repository,
		&effect.Issue, &effect.Capability, &effect.Target, &effect.PayloadHash, &effect.Epoch); err != nil {
		return Effect{}, err
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return Effect{}, err
	}
	committed = true
	return effect, nil
}

func (s *Store) MarkEffectSent(ctx context.Context, lease Lease, key string, now time.Time) error {
	return s.markClaimedEffect(ctx, lease, key, "sent", "", now)
}

func (s *Store) MarkEffectFailed(ctx context.Context, lease Lease, key string, cause error, now time.Time) error {
	message := "effect failed"
	if cause != nil {
		message = boundedStoreString(cause.Error(), 512)
	}
	return s.markClaimedEffect(ctx, lease, key, "failed", message, now)
}

func (s *Store) markClaimedEffect(ctx context.Context, lease Lease, key, status, message string, now time.Time) error {
	if key == "" || (status != "sent" && status != "failed") {
		return errors.New("effect key and terminal status are required")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE outbox SET status = ?, sent_at = ?, last_error = ?
		WHERE effect_key = ? AND run_id = ? AND epoch = ? AND status = 'claimed' AND claimed_by = ?`, status, now.UTC().UnixNano(), message, key, lease.RunID, lease.Epoch, lease.Owner)
	if err != nil {
		return fmt.Errorf("mark effect %s: %w", status, err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return errors.New("claimed effect is stale or fenced")
	}
	return nil
}

func (s *Store) Enqueue(ctx context.Context, lease Lease, effect Effect, now time.Time) (bool, error) {
	if effect.Key == "" || effect.RunID != lease.RunID || effect.Repository == "" || effect.Issue <= 0 ||
		effect.Target == "" || !validSHA256(effect.PayloadHash) || effect.Epoch != lease.Epoch ||
		!domain.IsGrantableCapability(effect.Capability) {
		return false, errors.New("effect identity does not match fenced lease")
	}
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return false, fmt.Errorf("begin outbox transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
		}
	}()
	var owner string
	var epoch, expiresAt int64
	if err := conn.QueryRowContext(ctx, `SELECT owner, epoch, expires_at FROM leases WHERE run_id = ?`, lease.RunID).
		Scan(&owner, &epoch, &expiresAt); err != nil {
		return false, fmt.Errorf("read fenced lease: %w", err)
	}
	if owner != lease.Owner || epoch != lease.Epoch || now.UnixNano() >= expiresAt {
		return false, errors.New("lease is stale, expired, or fenced")
	}
	var grantCount int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM grants WHERE run_id = ? AND repository = ?
		AND issue = ? AND capability = ? AND epoch = ?`, effect.RunID, effect.Repository, effect.Issue,
		effect.Capability, effect.Epoch).Scan(&grantCount); err != nil {
		return false, fmt.Errorf("check effect grant: %w", err)
	}
	if grantCount != 1 {
		return false, errors.New("no matching capability grant for effect")
	}
	result, err := conn.ExecContext(ctx, `INSERT OR IGNORE INTO outbox
		(effect_key, run_id, repository, issue, capability, target, payload_hash, epoch, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, effect.Key, effect.RunID, effect.Repository, effect.Issue,
		effect.Capability, effect.Target, effect.PayloadHash, effect.Epoch, now.UTC().UnixNano())
	if err != nil {
		return false, fmt.Errorf("enqueue effect: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return false, err
	}
	committed = true
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows == 1 {
		return true, nil
	}
	var existing Effect
	if err := conn.QueryRowContext(ctx, `SELECT effect_key, run_id, repository, issue, capability,
        target, payload_hash, epoch FROM outbox WHERE effect_key = ?`, effect.Key).Scan(&existing.Key,
		&existing.RunID, &existing.Repository, &existing.Issue, &existing.Capability, &existing.Target,
		&existing.PayloadHash, &existing.Epoch); err != nil {
		return false, err
	}
	if existing != effect {
		return false, errors.New("idempotency key collides with a different effect")
	}
	return false, nil
}
