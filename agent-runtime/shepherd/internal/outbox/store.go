package outbox

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrFenced         = errors.New("outbox operation is fenced")
	ErrTerminal       = errors.New("outbox effect is terminal")
	ErrStaleRevision  = errors.New("outbox effect revision is stale")
	ErrRetryExhausted = errors.New("outbox pre-send retry budget is exhausted")
	ErrStaleUncertain = errors.New("stale outbox effect has uncertain external completion")
)

type Result struct {
	Code          ResultCode
	ExternalID    int64
	ExternalActor string
}

type EffectRecord struct {
	EffectID       string
	IdempotencyKey string
	Kind           Kind
	DeliveryID     string
	Target         Target
	Generation     int64
	HeadSHA        string
	SourceID       string
	Revision       int64
	Payload        []byte
	PayloadHash    string
	State          State
	RetryCount     int64
	ErrorCode      ErrorCode
	Result         Result
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ClaimedEffect struct {
	EffectRecord
	ClaimID          string
	GrantID          string
	ControllerOwner  string
	ControllerEpoch  int64
	ClaimedAt        time.Time
	ExpiresAt        time.Time
	ExecutionStarted time.Time
}

type Event struct {
	ID        string
	Sequence  int64
	EffectID  string
	ClaimID   string
	GrantID   string
	Kind      EventKind
	State     State
	ErrorCode ErrorCode
	At        time.Time
}

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, path string) (*Store, error) {
	if !filepath.IsAbs(path) || filepath.Ext(path) != ".db" {
		return nil, errors.New("absolute outbox database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create outbox directory: %w", err)
	}
	if err := os.Chmod(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("secure outbox directory: %w", err)
	}
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open outbox sqlite: %w", err)
	}
	database.SetMaxOpenConns(1)
	store := &Store{db: database}
	if err := store.migrate(ctx); err != nil {
		_ = database.Close()
		return nil, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("secure outbox database: %w", err)
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
		`CREATE TABLE IF NOT EXISTS effects (
			effect_id TEXT PRIMARY KEY, idempotency_key TEXT NOT NULL UNIQUE, kind TEXT NOT NULL,
			delivery_id TEXT NOT NULL, repository TEXT NOT NULL, issue INTEGER NOT NULL,
			pull_request INTEGER NOT NULL, generation INTEGER NOT NULL, head_sha TEXT NOT NULL,
			source_id TEXT NOT NULL, revision INTEGER NOT NULL, payload BLOB NOT NULL,
			payload_hash TEXT NOT NULL, state TEXT NOT NULL, retry_count INTEGER NOT NULL DEFAULT 0,
			error_code TEXT NOT NULL DEFAULT '', created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS authorizations (
			grant_id TEXT PRIMARY KEY, effect_id TEXT NOT NULL REFERENCES effects(effect_id),
			capability TEXT NOT NULL, controller_owner TEXT NOT NULL, controller_epoch INTEGER NOT NULL,
			issued_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS claims (
			claim_id TEXT PRIMARY KEY, effect_id TEXT NOT NULL REFERENCES effects(effect_id),
			grant_id TEXT NOT NULL REFERENCES authorizations(grant_id), controller_owner TEXT NOT NULL,
			controller_epoch INTEGER NOT NULL, claimed_at INTEGER NOT NULL, expires_at INTEGER NOT NULL,
			execution_started_at INTEGER NOT NULL DEFAULT 0, closed_at INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS claims_one_open ON claims(effect_id) WHERE closed_at = 0`,
		`CREATE TABLE IF NOT EXISTS results (
			effect_id TEXT PRIMARY KEY REFERENCES effects(effect_id), claim_id TEXT NOT NULL,
			code TEXT NOT NULL, external_id INTEGER NOT NULL, external_actor TEXT NOT NULL,
			created_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS effect_events (
			event_id TEXT PRIMARY KEY, effect_id TEXT NOT NULL REFERENCES effects(effect_id),
			sequence INTEGER NOT NULL, claim_id TEXT NOT NULL DEFAULT '', grant_id TEXT NOT NULL DEFAULT '',
			kind TEXT NOT NULL, state TEXT NOT NULL, error_code TEXT NOT NULL DEFAULT '', created_at INTEGER NOT NULL,
			UNIQUE(effect_id, sequence)
		)`,
		`CREATE INDEX IF NOT EXISTS effects_delivery_state ON effects(delivery_id, state, created_at)`,
		`CREATE INDEX IF NOT EXISTS effect_events_effect ON effect_events(effect_id, created_at, event_id)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply outbox migration: %w", err)
		}
	}
	return nil
}

func recordFromIntent(intent Intent) EffectRecord {
	return EffectRecord{
		EffectID: intent.effectID, IdempotencyKey: intent.idempotencyKey, Kind: intent.kind,
		DeliveryID: intent.deliveryID, Target: intent.target, Generation: intent.generation,
		HeadSHA: intent.headSHA, SourceID: intent.sourceID, Revision: intent.revision,
		Payload: append([]byte(nil), intent.payload...), PayloadHash: intent.payloadHash, State: StatePending,
	}
}

func (s *Store) Enqueue(ctx context.Context, authorization Authorization, now time.Time) (EffectRecord, bool, error) {
	if err := authorization.validate(); err != nil || now.IsZero() {
		if err == nil {
			err = errors.New("enqueue time is required")
		}
		return EffectRecord{}, false, err
	}
	conn, err := s.beginImmediate(ctx)
	if err != nil {
		return EffectRecord{}, false, err
	}
	defer rollback(conn)
	intent := authorization.intent
	if intent.kind == KindDecisionSummary {
		var existing, active int
		if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM effects WHERE effect_id = ?`,
			intent.effectID).Scan(&existing); err != nil {
			return EffectRecord{}, false, err
		}
		if existing == 0 {
			if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM effects WHERE delivery_id = ?
				AND repository = ? AND issue = ? AND pull_request = ? AND kind = ? AND source_id = ?
				AND state IN (?, ?)`, intent.deliveryID, intent.target.Repository, intent.target.Issue,
				intent.target.PullRequest, intent.kind, intent.sourceID, StateClaimed,
				StateUncertain).Scan(&active); err != nil {
				return EffectRecord{}, false, err
			}
			if active > 0 {
				return EffectRecord{}, false, ErrFenced
			}
		}
	}
	timestamp := now.UTC().UnixNano()
	result, err := conn.ExecContext(ctx, `INSERT OR IGNORE INTO effects
		(effect_id, idempotency_key, kind, delivery_id, repository, issue, pull_request, generation,
		 head_sha, source_id, revision, payload, payload_hash, state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, intent.effectID,
		intent.idempotencyKey, intent.kind, intent.deliveryID, intent.target.Repository, intent.target.Issue,
		intent.target.PullRequest, intent.generation, intent.headSHA, intent.sourceID, intent.revision,
		intent.payload, intent.payloadHash, StatePending, timestamp, timestamp)
	if err != nil {
		return EffectRecord{}, false, fmt.Errorf("insert immutable effect: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return EffectRecord{}, false, err
	}
	inserted := rows == 1
	record, err := getEffect(ctx, conn, intent.effectID)
	if err != nil {
		return EffectRecord{}, false, err
	}
	if err := matchIntent(record, intent); err != nil {
		return EffectRecord{}, false, err
	}
	grant := authorization.grant
	if _, err := conn.ExecContext(ctx, `INSERT OR IGNORE INTO authorizations
		(grant_id, effect_id, capability, controller_owner, controller_epoch, issued_at)
		VALUES (?, ?, ?, ?, ?, ?)`, grant.id, grant.effectID, grant.capability, grant.owner, grant.epoch,
		grant.issuedAt.UTC().UnixNano()); err != nil {
		return EffectRecord{}, false, fmt.Errorf("insert immutable authorization: %w", err)
	}
	if err := validateStoredGrant(ctx, conn, grant); err != nil {
		return EffectRecord{}, false, err
	}
	if inserted {
		if err := insertEvent(ctx, conn, Event{ID: intent.effectID + ":requested", EffectID: intent.effectID,
			Kind: EventRequested, State: StatePending, At: now}); err != nil {
			return EffectRecord{}, false, err
		}
	}
	if err := insertEvent(ctx, conn, Event{ID: grant.id + ":authorized", EffectID: intent.effectID,
		GrantID: grant.id, Kind: EventAuthorized, State: record.State, At: now}); err != nil {
		return EffectRecord{}, false, err
	}
	if inserted {
		if err := insertEvent(ctx, conn, Event{ID: intent.effectID + ":enqueued", EffectID: intent.effectID,
			GrantID: grant.id, Kind: EventEnqueued, State: StatePending, At: now}); err != nil {
			return EffectRecord{}, false, err
		}
	}
	if err := commit(conn); err != nil {
		return EffectRecord{}, false, err
	}
	committedRecord, err := s.Get(ctx, intent.effectID)
	return committedRecord, inserted, err
}

func (s *Store) Get(ctx context.Context, effectID string) (EffectRecord, error) {
	if effectID == "" {
		return EffectRecord{}, errors.New("effect id is required")
	}
	return getEffect(ctx, s.db, effectID)
}

func (s *Store) HasHigherSummaryRevision(ctx context.Context, record EffectRecord) (bool, error) {
	if record.Kind != KindDecisionSummary || record.EffectID == "" || record.Revision <= 0 {
		return false, errors.New("complete summary effect identity is required")
	}
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM effects WHERE delivery_id = ?
		AND repository = ? AND issue = ? AND pull_request = ? AND kind = ? AND source_id = ?
		AND generation = ? AND head_sha = ? AND revision > ? AND state <> ?`, record.DeliveryID,
		record.Target.Repository, record.Target.Issue, record.Target.PullRequest, record.Kind,
		record.SourceID, record.Generation, record.HeadSHA, record.Revision, StateCancelled).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) FindSentSummary(ctx context.Context, deliveryID string, target Target, revision int64,
	ledgerHash string) (EffectRecord, bool, error) {
	if !safeDeliveryID(deliveryID) || validateTarget(target) != nil || revision <= 0 || !validSHA256(ledgerHash) {
		return EffectRecord{}, false, errors.New("complete summary projection identity is required")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT effect_id FROM effects WHERE delivery_id = ?
		AND repository = ? AND issue = ? AND pull_request = ? AND kind = ? AND revision = ? AND state = ?
		ORDER BY created_at, effect_id`, deliveryID, target.Repository, target.Issue, target.PullRequest,
		KindDecisionSummary, revision, StateSent)
	if err != nil {
		return EffectRecord{}, false, err
	}
	defer func() { _ = rows.Close() }()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return EffectRecord{}, false, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return EffectRecord{}, false, err
	}
	var matched EffectRecord
	var matchedSummary string
	for _, id := range ids {
		record, err := s.Get(ctx, id)
		if err != nil {
			return EffectRecord{}, false, err
		}
		var payload SummaryPayload
		if err := decodeStrict(record.Payload, &payload); err != nil {
			return EffectRecord{}, false, err
		}
		if payload.LedgerHash != ledgerHash {
			continue
		}
		if matched.EffectID != "" && (matched.Result != record.Result || matchedSummary != payload.Summary) {
			return EffectRecord{}, false, errors.New("sent summary projection identity is ambiguous")
		}
		matched, matchedSummary = record, payload.Summary
	}
	if matched.EffectID == "" {
		return EffectRecord{}, false, nil
	}
	return matched, true, nil
}

func (s *Store) ListDelivery(ctx context.Context, deliveryID string, states ...State) ([]EffectRecord, error) {
	if !safeDeliveryID(deliveryID) {
		return nil, errors.New("delivery identity is required")
	}
	query := `SELECT effect_id FROM effects WHERE delivery_id = ?`
	arguments := []any{deliveryID}
	if len(states) > 0 {
		query += ` AND state IN (`
		for index, state := range states {
			if !validState(state) {
				return nil, fmt.Errorf("unknown effect state %q", state)
			}
			if index > 0 {
				query += `,`
			}
			query += `?`
			arguments = append(arguments, state)
		}
		query += `)`
	}
	query += ` ORDER BY created_at, effect_id`
	rows, err := s.db.QueryContext(ctx, query, arguments...)
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
	records := make([]EffectRecord, 0, len(ids))
	for _, id := range ids {
		record, err := s.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (s *Store) Claim(ctx context.Context, authorization Authorization, owner string, epoch int64, now time.Time, ttl time.Duration) (ClaimedEffect, error) {
	if err := authorization.validate(); err != nil {
		return ClaimedEffect{}, err
	}
	if owner != authorization.grant.owner || epoch != authorization.grant.epoch || ttl <= 0 || now.IsZero() {
		return ClaimedEffect{}, ErrFenced
	}
	conn, err := s.beginImmediate(ctx)
	if err != nil {
		return ClaimedEffect{}, err
	}
	defer rollback(conn)
	record, err := getEffect(ctx, conn, authorization.intent.effectID)
	if err != nil {
		return ClaimedEffect{}, err
	}
	if err := matchIntent(record, authorization.intent); err != nil {
		return ClaimedEffect{}, err
	}
	if record.State == StateSent || record.State == StateUncertain || record.State == StateBlocked || record.State == StateCancelled {
		return ClaimedEffect{}, ErrTerminal
	}
	if record.State != StatePending {
		return ClaimedEffect{}, fmt.Errorf("effect in state %q cannot be claimed", record.State)
	}
	if err := validateStoredGrant(ctx, conn, authorization.grant); err != nil {
		return ClaimedEffect{}, err
	}
	if record.Kind == KindDecisionSummary {
		var active, newer int
		if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM effects
			WHERE delivery_id = ? AND repository = ? AND issue = ? AND pull_request = ? AND kind = ?
			AND source_id = ? AND generation = ? AND head_sha = ? AND revision > ? AND state <> ?`,
			record.DeliveryID, record.Target.Repository, record.Target.Issue, record.Target.PullRequest,
			record.Kind, record.SourceID, record.Generation, record.HeadSHA, record.Revision,
			StateCancelled).Scan(&newer); err != nil {
			return ClaimedEffect{}, err
		}
		if newer > 0 {
			if _, err := conn.ExecContext(ctx, `UPDATE effects SET state = ?, error_code = ?, updated_at = ?
				WHERE effect_id = ? AND state = ?`, StateCancelled, ErrorStaleRevision,
				now.UTC().UnixNano(), record.EffectID, StatePending); err != nil {
				return ClaimedEffect{}, err
			}
			if err := insertEvent(ctx, conn, Event{ID: record.EffectID + ":superseded",
				EffectID: record.EffectID, GrantID: authorization.grant.id, Kind: EventCancelled,
				State: StateCancelled, ErrorCode: ErrorStaleRevision, At: now}); err != nil {
				return ClaimedEffect{}, err
			}
			if err := commit(conn); err != nil {
				return ClaimedEffect{}, err
			}
			return ClaimedEffect{}, ErrStaleRevision
		}
		if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM effects
			WHERE delivery_id = ? AND repository = ? AND issue = ? AND pull_request = ? AND kind = ?
			AND source_id = ? AND state IN (?, ?) AND effect_id <> ?`, record.DeliveryID,
			record.Target.Repository, record.Target.Issue, record.Target.PullRequest, record.Kind,
			record.SourceID, StateClaimed, StateUncertain, record.EffectID).Scan(&active); err != nil {
			return ClaimedEffect{}, err
		}
		if active > 0 {
			return ClaimedEffect{}, ErrFenced
		}
	}
	claimID, err := randomID("claim-")
	if err != nil {
		return ClaimedEffect{}, err
	}
	expiresAt := now.UTC().Add(ttl)
	if _, err := conn.ExecContext(ctx, `INSERT INTO claims
		(claim_id, effect_id, grant_id, controller_owner, controller_epoch, claimed_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, claimID, record.EffectID, authorization.grant.id, owner, epoch,
		now.UTC().UnixNano(), expiresAt.UnixNano()); err != nil {
		return ClaimedEffect{}, fmt.Errorf("insert effect claim: %w", err)
	}
	updated, err := conn.ExecContext(ctx, `UPDATE effects SET state = ?, error_code = '', updated_at = ?
		WHERE effect_id = ? AND state = ?`, StateClaimed, now.UTC().UnixNano(), record.EffectID, StatePending)
	if err != nil {
		return ClaimedEffect{}, err
	}
	if rows, _ := updated.RowsAffected(); rows != 1 {
		return ClaimedEffect{}, ErrFenced
	}
	if err := insertEvent(ctx, conn, Event{ID: claimID + ":claimed", EffectID: record.EffectID,
		ClaimID: claimID, GrantID: authorization.grant.id, Kind: EventClaimed, State: StateClaimed, At: now}); err != nil {
		return ClaimedEffect{}, err
	}
	if err := commit(conn); err != nil {
		return ClaimedEffect{}, err
	}
	record.State = StateClaimed
	record.UpdatedAt = now.UTC()
	return ClaimedEffect{EffectRecord: record, ClaimID: claimID, GrantID: authorization.grant.id,
		ControllerOwner: owner, ControllerEpoch: epoch, ClaimedAt: now.UTC(), ExpiresAt: expiresAt}, nil
}

func (s *Store) StartExecution(ctx context.Context, claim ClaimedEffect, now time.Time) error {
	if err := validateClaim(claim); err != nil || now.IsZero() {
		if err == nil {
			err = errors.New("execution start time is required")
		}
		return err
	}
	conn, err := s.beginImmediate(ctx)
	if err != nil {
		return err
	}
	defer rollback(conn)
	if !now.Before(claim.ExpiresAt) {
		return ErrFenced
	}
	if claim.Kind == KindDecisionSummary {
		var newer int
		if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM effects
			WHERE delivery_id = ? AND repository = ? AND issue = ? AND pull_request = ? AND kind = ?
			AND source_id = ? AND generation = ? AND head_sha = ? AND revision > ? AND state <> ?`,
			claim.DeliveryID, claim.Target.Repository, claim.Target.Issue, claim.Target.PullRequest,
			claim.Kind, claim.SourceID, claim.Generation, claim.HeadSHA, claim.Revision,
			StateCancelled).Scan(&newer); err != nil {
			return err
		}
		if newer > 0 {
			return ErrStaleRevision
		}
	}
	result, err := conn.ExecContext(ctx, `UPDATE claims SET execution_started_at = ?
		WHERE claim_id = ? AND effect_id = ? AND grant_id = ? AND controller_owner = ? AND controller_epoch = ?
		AND execution_started_at = 0 AND closed_at = 0`, now.UTC().UnixNano(), claim.ClaimID, claim.EffectID,
		claim.GrantID, claim.ControllerOwner, claim.ControllerEpoch)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return ErrFenced
	}
	if err := insertEvent(ctx, conn, Event{ID: claim.ClaimID + ":execution", EffectID: claim.EffectID,
		ClaimID: claim.ClaimID, GrantID: claim.GrantID, Kind: EventExecutionStarted, State: StateClaimed, At: now}); err != nil {
		return err
	}
	return commit(conn)
}

func (s *Store) MarkSent(ctx context.Context, claim ClaimedEffect, result Result, now time.Time) error {
	if result.ExternalID <= 0 || !safeID(result.ExternalActor) ||
		(result.Code != ResultSent && result.Code != ResultReconciled) {
		return errors.New("typed sent result and external identity are required")
	}
	return s.finishClaim(ctx, claim, StateSent, ErrorNone, result, now)
}

func (s *Store) MarkFailed(ctx context.Context, claim ClaimedEffect, code ErrorCode, now time.Time) error {
	if code != ErrorPreSend {
		return errors.New("only a definite pre-send failure may enter failed")
	}
	if claim.RetryCount >= 1 {
		return s.finishClaim(ctx, claim, StateBlocked, ErrorRetryExhausted, Result{}, now)
	}
	return s.finishClaim(ctx, claim, StateFailed, code, Result{}, now)
}

func (s *Store) MarkUncertain(ctx context.Context, claim ClaimedEffect, code ErrorCode, now time.Time) error {
	if code != ErrorPostSendAmbiguous {
		return errors.New("execution uncertainty requires a post-send ambiguous error")
	}
	return s.finishClaim(ctx, claim, StateUncertain, code, Result{}, now)
}

func (s *Store) MarkBlocked(ctx context.Context, claim ClaimedEffect, code ErrorCode, now time.Time) error {
	if code == ErrorNone {
		return errors.New("blocked effect requires a typed reason")
	}
	return s.finishClaim(ctx, claim, StateBlocked, code, Result{}, now)
}

func (s *Store) finishClaim(ctx context.Context, claim ClaimedEffect, state State, code ErrorCode, result Result, now time.Time) error {
	if err := validateClaim(claim); err != nil || now.IsZero() {
		if err == nil {
			err = errors.New("terminal effect time is required")
		}
		return err
	}
	conn, err := s.beginImmediate(ctx)
	if err != nil {
		return err
	}
	defer rollback(conn)
	var current State
	var startedAt, closedAt int64
	if err := conn.QueryRowContext(ctx, `SELECT e.state, c.execution_started_at, c.closed_at
		FROM effects e JOIN claims c ON c.effect_id = e.effect_id
		WHERE e.effect_id = ? AND c.claim_id = ? AND c.grant_id = ? AND c.controller_owner = ? AND c.controller_epoch = ?`,
		claim.EffectID, claim.ClaimID, claim.GrantID, claim.ControllerOwner, claim.ControllerEpoch).Scan(&current, &startedAt, &closedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrFenced
		}
		return err
	}
	if current != StateClaimed || closedAt != 0 {
		if validTerminalState(current) {
			return ErrTerminal
		}
		return ErrFenced
	}
	if state == StateSent && !now.UTC().Before(claim.ExpiresAt) {
		return ErrFenced
	}
	if state == StateSent && result.Code == ResultSent && startedAt == 0 {
		return errors.New("a direct send cannot complete before execution starts")
	}
	if state == StateUncertain && startedAt == 0 {
		return errors.New("an unstarted effect cannot become post-send uncertain")
	}
	if (state == StateFailed || state == StateBlocked) && startedAt != 0 {
		return errors.New("an execution-started effect cannot be marked as definitely unsent")
	}
	updated, err := conn.ExecContext(ctx, `UPDATE effects SET state = ?, error_code = ?, updated_at = ?
		WHERE effect_id = ? AND state = ?`, state, code, now.UTC().UnixNano(), claim.EffectID, StateClaimed)
	if err != nil {
		return err
	}
	if rows, _ := updated.RowsAffected(); rows != 1 {
		return ErrFenced
	}
	if _, err := conn.ExecContext(ctx, `UPDATE claims SET closed_at = ? WHERE claim_id = ? AND closed_at = 0`,
		now.UTC().UnixNano(), claim.ClaimID); err != nil {
		return err
	}
	if state == StateSent {
		if _, err := conn.ExecContext(ctx, `INSERT INTO results
			(effect_id, claim_id, code, external_id, external_actor, created_at)
			VALUES (?, ?, ?, ?, ?, ?)`, claim.EffectID, claim.ClaimID, result.Code, result.ExternalID,
			result.ExternalActor, now.UTC().UnixNano()); err != nil {
			return fmt.Errorf("persist immutable effect result: %w", err)
		}
	}
	eventKind := mapStateEvent(state, result.Code)
	if err := insertEvent(ctx, conn, Event{ID: claim.ClaimID + ":" + string(eventKind), EffectID: claim.EffectID,
		ClaimID: claim.ClaimID, GrantID: claim.GrantID, Kind: eventKind, State: state, ErrorCode: code, At: now}); err != nil {
		return err
	}
	if state == StateSent && result.Code == ResultReconciled {
		if err := insertEvent(ctx, conn, Event{ID: claim.ClaimID + ":sent", EffectID: claim.EffectID,
			ClaimID: claim.ClaimID, GrantID: claim.GrantID, Kind: EventSent, State: state, At: now}); err != nil {
			return err
		}
	}
	return commit(conn)
}

func (s *Store) MarkReconciled(ctx context.Context, authorization Authorization, result Result, now time.Time) error {
	if err := authorization.validate(); err != nil || result.Code != ResultReconciled ||
		result.ExternalID <= 0 || !safeID(result.ExternalActor) || now.IsZero() {
		if err == nil {
			err = errors.New("complete reconciliation authorization, result, and time are required")
		}
		return err
	}
	conn, err := s.beginImmediate(ctx)
	if err != nil {
		return err
	}
	defer rollback(conn)
	record, err := getEffect(ctx, conn, authorization.EffectID())
	if err != nil {
		return err
	}
	if err := matchIntent(record, authorization.intent); err != nil {
		return err
	}
	if record.State != StateUncertain {
		if validTerminalState(record.State) {
			return ErrTerminal
		}
		return errors.New("only an uncertain effect may be reconciled")
	}
	if err := validateStoredGrant(ctx, conn, authorization.grant); err != nil {
		return err
	}
	claimID, err := randomID("reconcile-")
	if err != nil {
		return err
	}
	timestamp := now.UTC().UnixNano()
	if _, err := conn.ExecContext(ctx, `INSERT INTO claims
		(claim_id, effect_id, grant_id, controller_owner, controller_epoch, claimed_at, expires_at,
		 execution_started_at, closed_at) VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?)`, claimID, record.EffectID,
		authorization.GrantID(), authorization.Owner(), authorization.Epoch(), timestamp, timestamp, timestamp); err != nil {
		return fmt.Errorf("insert reconciliation claim: %w", err)
	}
	updated, err := conn.ExecContext(ctx, `UPDATE effects SET state = ?, error_code = '', updated_at = ?
		WHERE effect_id = ? AND state = ?`, StateSent, timestamp, record.EffectID, StateUncertain)
	if err != nil {
		return err
	}
	if rows, _ := updated.RowsAffected(); rows != 1 {
		return ErrFenced
	}
	if _, err := conn.ExecContext(ctx, `INSERT INTO results
		(effect_id, claim_id, code, external_id, external_actor, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`, record.EffectID, claimID, result.Code, result.ExternalID,
		result.ExternalActor, timestamp); err != nil {
		return err
	}
	for _, event := range []Event{
		{ID: claimID + ":claimed", EffectID: record.EffectID, ClaimID: claimID, GrantID: authorization.GrantID(), Kind: EventClaimed, State: StateUncertain, At: now},
		{ID: claimID + ":reconciled", EffectID: record.EffectID, ClaimID: claimID, GrantID: authorization.GrantID(), Kind: EventReconciled, State: StateSent, At: now},
		{ID: claimID + ":sent", EffectID: record.EffectID, ClaimID: claimID, GrantID: authorization.GrantID(), Kind: EventSent, State: StateSent, At: now},
	} {
		if err := insertEvent(ctx, conn, event); err != nil {
			return err
		}
	}
	return commit(conn)
}

func (s *Store) RetryFailed(ctx context.Context, authorization Authorization, now time.Time) error {
	if err := authorization.validate(); err != nil || now.IsZero() {
		if err == nil {
			err = errors.New("retry time is required")
		}
		return err
	}
	conn, err := s.beginImmediate(ctx)
	if err != nil {
		return err
	}
	defer rollback(conn)
	record, err := getEffect(ctx, conn, authorization.EffectID())
	if err != nil {
		return err
	}
	if err := matchIntent(record, authorization.intent); err != nil {
		return err
	}
	if err := validateStoredGrant(ctx, conn, authorization.grant); err != nil {
		return err
	}
	if record.State != StateFailed {
		if validTerminalState(record.State) {
			return ErrTerminal
		}
		return errors.New("only a definitely failed effect may consume retry budget")
	}
	if record.Kind == KindDecisionSummary {
		var newer int
		if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM effects WHERE delivery_id = ?
			AND repository = ? AND issue = ? AND pull_request = ? AND kind = ? AND source_id = ?
			AND generation = ? AND head_sha = ? AND revision > ? AND state <> ?`, record.DeliveryID,
			record.Target.Repository, record.Target.Issue, record.Target.PullRequest, record.Kind,
			record.SourceID, record.Generation, record.HeadSHA, record.Revision, StateCancelled).Scan(&newer); err != nil {
			return err
		}
		if newer > 0 {
			if _, err := conn.ExecContext(ctx, `UPDATE effects SET state = ?, error_code = ?, updated_at = ?
				WHERE effect_id = ? AND state = ?`, StateCancelled, ErrorStaleRevision,
				now.UTC().UnixNano(), record.EffectID, StateFailed); err != nil {
				return err
			}
			if err := insertEvent(ctx, conn, Event{ID: record.EffectID + ":superseded",
				EffectID: record.EffectID, GrantID: authorization.GrantID(), Kind: EventCancelled,
				State: StateCancelled, ErrorCode: ErrorStaleRevision, At: now}); err != nil {
				return err
			}
			if err := commit(conn); err != nil {
				return err
			}
			return ErrStaleRevision
		}
	}
	if record.RetryCount >= 1 {
		if _, err := conn.ExecContext(ctx, `UPDATE effects SET state = ?, error_code = ?, updated_at = ?
			WHERE effect_id = ? AND state = ?`, StateBlocked, ErrorRetryExhausted, now.UTC().UnixNano(),
			record.EffectID, StateFailed); err != nil {
			return err
		}
		if err := insertEvent(ctx, conn, Event{ID: record.EffectID + ":retry-exhausted",
			EffectID: record.EffectID, GrantID: authorization.GrantID(), Kind: EventBlocked,
			State: StateBlocked, ErrorCode: ErrorRetryExhausted, At: now}); err != nil {
			return err
		}
		if err := commit(conn); err != nil {
			return err
		}
		return ErrRetryExhausted
	}
	result, err := conn.ExecContext(ctx, `UPDATE effects SET state = ?, retry_count = retry_count + 1,
		error_code = '', updated_at = ? WHERE effect_id = ? AND state = ? AND retry_count = ?`, StatePending,
		now.UTC().UnixNano(), record.EffectID, StateFailed, record.RetryCount)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return ErrFenced
	}
	return commit(conn)
}

func (s *Store) Cancel(ctx context.Context, authorization Authorization, now time.Time) error {
	if err := authorization.validate(); err != nil || now.IsZero() {
		if err == nil {
			err = errors.New("cancellation time is required")
		}
		return err
	}
	conn, err := s.beginImmediate(ctx)
	if err != nil {
		return err
	}
	defer rollback(conn)
	record, err := getEffect(ctx, conn, authorization.EffectID())
	if err != nil {
		return err
	}
	if err := matchIntent(record, authorization.intent); err != nil {
		return err
	}
	if err := validateStoredGrant(ctx, conn, authorization.grant); err != nil {
		return err
	}
	if record.State != StatePending && record.State != StateFailed {
		if validTerminalState(record.State) {
			return ErrTerminal
		}
		return errors.New("only a pending or definitely failed effect may be cancelled")
	}
	result, err := conn.ExecContext(ctx, `UPDATE effects SET state = ?, error_code = '', updated_at = ?
		WHERE effect_id = ? AND state = ?`, StateCancelled, now.UTC().UnixNano(), record.EffectID, record.State)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return ErrFenced
	}
	if err := insertEvent(ctx, conn, Event{ID: record.EffectID + ":cancelled", EffectID: record.EffectID,
		GrantID: authorization.GrantID(), Kind: EventCancelled, State: StateCancelled, At: now}); err != nil {
		return err
	}
	return commit(conn)
}

func (s *Store) CancelExpiredQuestion(ctx context.Context, controller *Controller, effectID string, now time.Time) error {
	if controller == nil || effectID == "" || now.IsZero() {
		return errors.New("controller, effect, and expiry time are required")
	}
	if err := controller.Revalidate(ctx, now); err != nil {
		return err
	}
	conn, err := s.beginImmediate(ctx)
	if err != nil {
		return err
	}
	defer rollback(conn)
	record, err := getEffect(ctx, conn, effectID)
	if err != nil {
		return err
	}
	intent, err := DecodeIntent(record.Kind, record.Payload)
	if err != nil || record.Kind != KindDecisionQuestion || controller.matches(intent) != nil {
		return errors.New("expired question effect does not match current protected authority")
	}
	var payload QuestionPayload
	if err := decodeStrict(record.Payload, &payload); err != nil {
		return err
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, payload.ExpiresAt)
	if err != nil || now.Before(expiresAt) {
		return errors.New("question effect is not expired")
	}
	switch record.State {
	case StatePending, StateFailed:
		if _, err := conn.ExecContext(ctx, `UPDATE effects SET state = ?, error_code = ?, updated_at = ?
			WHERE effect_id = ? AND state = ?`, StateCancelled, ErrorEffectExpired, now.UTC().UnixNano(),
			record.EffectID, record.State); err != nil {
			return err
		}
		if err := insertEvent(ctx, conn, Event{ID: record.EffectID + ":expired", EffectID: record.EffectID,
			Kind: EventCancelled, State: StateCancelled, ErrorCode: ErrorEffectExpired, At: now}); err != nil {
			return err
		}
	case StateClaimed:
		var claimID, grantID string
		var startedAt int64
		if err := conn.QueryRowContext(ctx, `SELECT claim_id, grant_id, execution_started_at FROM claims
			WHERE effect_id = ? AND closed_at = 0`, record.EffectID).Scan(&claimID, &grantID, &startedAt); err != nil {
			return err
		}
		state, event := StateCancelled, EventCancelled
		if startedAt != 0 {
			state, event = StateUncertain, EventUncertain
		}
		if _, err := conn.ExecContext(ctx, `UPDATE claims SET closed_at = ? WHERE claim_id = ? AND closed_at = 0`,
			now.UTC().UnixNano(), claimID); err != nil {
			return err
		}
		if _, err := conn.ExecContext(ctx, `UPDATE effects SET state = ?, error_code = ?, updated_at = ?
			WHERE effect_id = ? AND state = ?`, state, ErrorEffectExpired, now.UTC().UnixNano(),
			record.EffectID, StateClaimed); err != nil {
			return err
		}
		if err := insertEvent(ctx, conn, Event{ID: claimID + ":expired", EffectID: record.EffectID,
			ClaimID: claimID, GrantID: grantID, Kind: event, State: state,
			ErrorCode: ErrorEffectExpired, At: now}); err != nil {
			return err
		}
		if state == StateUncertain {
			if err := commit(conn); err != nil {
				return err
			}
			return ErrStaleUncertain
		}
	case StateUncertain:
		if err := commit(conn); err != nil {
			return err
		}
		return ErrStaleUncertain
	case StateSent, StateBlocked, StateCancelled:
		// Already terminal; request expiry only changes local decision authority.
	}
	return commit(conn)
}

func (s *Store) ReconcileStale(ctx context.Context, controller *Controller, now time.Time) error {
	if controller == nil || now.IsZero() {
		return errors.New("controller and stale-effect reconciliation time are required")
	}
	if err := controller.Revalidate(ctx, now); err != nil {
		return err
	}
	conn, err := s.beginImmediate(ctx)
	if err != nil {
		return err
	}
	defer rollback(conn)
	facts := controller.facts
	rows, err := conn.QueryContext(ctx, `SELECT effect_id FROM effects WHERE delivery_id = ?
		AND state IN (?, ?, ?, ?, ?) AND (repository <> ? OR issue <> ? OR pull_request <> ?
		OR generation <> ? OR head_sha <> ?) ORDER BY created_at, effect_id`, facts.DeliveryID,
		StatePending, StateFailed, StateClaimed, StateUncertain, StateBlocked, facts.Repository,
		facts.Issue, facts.PullRequest, facts.Generation, facts.HeadSHA)
	if err != nil {
		return err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	var terminalErr error
	for _, id := range ids {
		record, err := getEffect(ctx, conn, id)
		if err != nil {
			return err
		}
		switch record.State {
		case StatePending, StateFailed:
			if _, err := conn.ExecContext(ctx, `UPDATE effects SET state = ?, error_code = ?, updated_at = ?
				WHERE effect_id = ? AND state = ?`, StateCancelled, ErrorChangedTarget, now.UTC().UnixNano(),
				record.EffectID, record.State); err != nil {
				return err
			}
			if err := insertEvent(ctx, conn, Event{ID: record.EffectID + ":stale-cancelled",
				EffectID: record.EffectID, Kind: EventCancelled, State: StateCancelled,
				ErrorCode: ErrorChangedTarget, At: now}); err != nil {
				return err
			}
		case StateClaimed:
			var claimID, grantID string
			var expiresAt, startedAt int64
			if err := conn.QueryRowContext(ctx, `SELECT claim_id, grant_id, expires_at, execution_started_at
				FROM claims WHERE effect_id = ? AND closed_at = 0`, record.EffectID).Scan(
				&claimID, &grantID, &expiresAt, &startedAt); err != nil {
				return err
			}
			if now.UTC().UnixNano() < expiresAt {
				terminalErr = errors.Join(terminalErr, ErrFenced)
				continue
			}
			if _, err := conn.ExecContext(ctx, `UPDATE claims SET closed_at = ? WHERE claim_id = ? AND closed_at = 0`,
				now.UTC().UnixNano(), claimID); err != nil {
				return err
			}
			state, code, event := StateCancelled, ErrorChangedTarget, EventCancelled
			if startedAt != 0 {
				state, code, event = StateUncertain, ErrorChangedTarget, EventUncertain
				terminalErr = errors.Join(terminalErr, ErrStaleUncertain)
			}
			if _, err := conn.ExecContext(ctx, `UPDATE effects SET state = ?, error_code = ?, updated_at = ?
				WHERE effect_id = ? AND state = ?`, state, code, now.UTC().UnixNano(), record.EffectID,
				StateClaimed); err != nil {
				return err
			}
			if err := insertEvent(ctx, conn, Event{ID: claimID + ":stale-recovered", EffectID: record.EffectID,
				ClaimID: claimID, GrantID: grantID, Kind: EventClaimRecovered, State: state,
				ErrorCode: code, At: now}); err != nil {
				return err
			}
			if err := insertEvent(ctx, conn, Event{ID: claimID + ":stale-terminal", EffectID: record.EffectID,
				ClaimID: claimID, GrantID: grantID, Kind: event, State: state, ErrorCode: code, At: now}); err != nil {
				return err
			}
		case StateUncertain:
			terminalErr = errors.Join(terminalErr, ErrStaleUncertain)
		case StateBlocked:
			// A stale blocked effect is terminal and non-replayable, but a
			// definite old-target blocker does not widen or block new authority.
		}
	}
	if err := commit(conn); err != nil {
		return err
	}
	return terminalErr
}

func (s *Store) RecoverExpiredClaims(ctx context.Context, controller *Controller, now time.Time) ([]EffectRecord, error) {
	if controller == nil || now.IsZero() {
		return nil, errors.New("controller and recovery time are required")
	}
	if err := controller.Revalidate(ctx, now); err != nil {
		return nil, err
	}
	conn, err := s.beginImmediate(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(conn)
	rows, err := conn.QueryContext(ctx, `SELECT c.claim_id, c.effect_id, c.grant_id, c.controller_owner,
		c.controller_epoch, c.execution_started_at FROM claims c JOIN effects e ON e.effect_id = c.effect_id
		WHERE e.delivery_id = ? AND e.repository = ? AND e.issue = ? AND e.pull_request = ?
		AND e.generation = ? AND e.head_sha = ? AND e.state = ? AND c.closed_at = 0 AND c.expires_at <= ?
		ORDER BY c.claimed_at, c.claim_id`, controller.facts.DeliveryID, controller.facts.Repository,
		controller.facts.Issue, controller.facts.PullRequest, controller.facts.Generation,
		controller.facts.HeadSHA, StateClaimed, now.UTC().UnixNano())
	if err != nil {
		return nil, err
	}
	type expiredClaim struct {
		claimID, effectID, grantID, owner string
		epoch, startedAt                  int64
	}
	var claims []expiredClaim
	for rows.Next() {
		var claim expiredClaim
		if err := rows.Scan(&claim.claimID, &claim.effectID, &claim.grantID, &claim.owner, &claim.epoch, &claim.startedAt); err != nil {
			_ = rows.Close()
			return nil, err
		}
		claims = append(claims, claim)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	var recoveredIDs []string
	for _, claim := range claims {
		state, code := StatePending, ErrorNone
		if claim.startedAt != 0 {
			state, code = StateUncertain, ErrorClaimExpired
		}
		if _, err := conn.ExecContext(ctx, `UPDATE effects SET state = ?, error_code = ?, updated_at = ?
			WHERE effect_id = ? AND state = ?`, state, code, now.UTC().UnixNano(), claim.effectID, StateClaimed); err != nil {
			return nil, err
		}
		if _, err := conn.ExecContext(ctx, `UPDATE claims SET closed_at = ? WHERE claim_id = ? AND closed_at = 0`,
			now.UTC().UnixNano(), claim.claimID); err != nil {
			return nil, err
		}
		if err := insertEvent(ctx, conn, Event{ID: claim.claimID + ":recovered", EffectID: claim.effectID,
			ClaimID: claim.claimID, GrantID: claim.grantID, Kind: EventClaimRecovered, State: state,
			ErrorCode: code, At: now}); err != nil {
			return nil, err
		}
		if state == StateUncertain {
			if err := insertEvent(ctx, conn, Event{ID: claim.claimID + ":uncertain", EffectID: claim.effectID,
				ClaimID: claim.claimID, GrantID: claim.grantID, Kind: EventUncertain, State: state,
				ErrorCode: code, At: now}); err != nil {
				return nil, err
			}
		}
		recoveredIDs = append(recoveredIDs, claim.effectID)
	}
	if err := commit(conn); err != nil {
		return nil, err
	}
	recovered := make([]EffectRecord, 0, len(recoveredIDs))
	for _, id := range recoveredIDs {
		record, err := s.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		recovered = append(recovered, record)
	}
	return recovered, nil
}

func (s *Store) Events(ctx context.Context, effectID string) ([]Event, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT event_id, effect_id, sequence, claim_id, grant_id, kind, state,
		error_code, created_at FROM effect_events WHERE effect_id = ? ORDER BY sequence`, effectID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var events []Event
	for rows.Next() {
		var event Event
		var timestamp int64
		if err := rows.Scan(&event.ID, &event.EffectID, &event.Sequence, &event.ClaimID, &event.GrantID,
			&event.Kind, &event.State, &event.ErrorCode, &timestamp); err != nil {
			return nil, err
		}
		event.At = time.Unix(0, timestamp).UTC()
		events = append(events, event)
	}
	return events, rows.Err()
}

func getEffect(ctx context.Context, querier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, effectID string) (EffectRecord, error) {
	var record EffectRecord
	var payload []byte
	var createdAt, updatedAt int64
	err := querier.QueryRowContext(ctx, `SELECT e.effect_id, e.idempotency_key, e.kind, e.delivery_id,
		e.repository, e.issue, e.pull_request, e.generation, e.head_sha, e.source_id, e.revision,
		e.payload, e.payload_hash, e.state, e.retry_count, e.error_code, e.created_at, e.updated_at,
		COALESCE(r.code, ''), COALESCE(r.external_id, 0), COALESCE(r.external_actor, '')
		FROM effects e LEFT JOIN results r ON r.effect_id = e.effect_id WHERE e.effect_id = ?`, effectID).Scan(
		&record.EffectID, &record.IdempotencyKey, &record.Kind, &record.DeliveryID,
		&record.Target.Repository, &record.Target.Issue, &record.Target.PullRequest, &record.Generation,
		&record.HeadSHA, &record.SourceID, &record.Revision, &payload, &record.PayloadHash, &record.State,
		&record.RetryCount, &record.ErrorCode, &createdAt, &updatedAt, &record.Result.Code,
		&record.Result.ExternalID, &record.Result.ExternalActor)
	if err != nil {
		return EffectRecord{}, err
	}
	intent, err := DecodeIntent(record.Kind, payload)
	if err != nil {
		return EffectRecord{}, fmt.Errorf("decode persisted effect: %w", err)
	}
	record.Payload = append([]byte(nil), payload...)
	record.CreatedAt = time.Unix(0, createdAt).UTC()
	record.UpdatedAt = time.Unix(0, updatedAt).UTC()
	if err := matchIntent(record, intent); err != nil || !validState(record.State) {
		if err == nil {
			err = fmt.Errorf("persisted effect has unknown state %q", record.State)
		}
		return EffectRecord{}, err
	}
	return record, nil
}

func matchIntent(record EffectRecord, intent Intent) error {
	if record.EffectID != intent.effectID || record.IdempotencyKey != intent.idempotencyKey ||
		record.Kind != intent.kind || record.DeliveryID != intent.deliveryID || record.Target != intent.target ||
		record.Generation != intent.generation || record.HeadSHA != intent.headSHA || record.SourceID != intent.sourceID ||
		record.Revision != intent.revision || record.PayloadHash != intent.payloadHash ||
		string(record.Payload) != string(intent.payload) {
		return errors.New("effect identity collides with different immutable content")
	}
	return nil
}

func validateStoredGrant(ctx context.Context, querier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, grant Grant) error {
	var effectID, capability, owner string
	var epoch int64
	if err := querier.QueryRowContext(ctx, `SELECT effect_id, capability, controller_owner, controller_epoch
		FROM authorizations WHERE grant_id = ?`, grant.id).Scan(&effectID, &capability, &owner, &epoch); err != nil {
		return err
	}
	if effectID != grant.effectID || Capability(capability) != grant.capability || owner != grant.owner || epoch != grant.epoch {
		return errors.New("grant identity collides with different immutable authority")
	}
	return nil
}

func validateClaim(claim ClaimedEffect) error {
	if claim.EffectID == "" || claim.ClaimID == "" || claim.GrantID == "" || !safeID(claim.ControllerOwner) ||
		claim.ControllerEpoch <= 0 || claim.ClaimedAt.IsZero() || claim.ExpiresAt.IsZero() {
		return errors.New("complete fenced claim identity is required")
	}
	return nil
}

func insertEvent(ctx context.Context, executor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, event Event) error {
	if event.ID == "" || event.EffectID == "" || event.Kind == "" || !validState(event.State) || event.At.IsZero() {
		return errors.New("complete typed effect event is required")
	}
	if err := executor.QueryRowContext(ctx, `SELECT COALESCE(MAX(sequence), 0) + 1 FROM effect_events
		WHERE effect_id = ?`, event.EffectID).Scan(&event.Sequence); err != nil {
		return err
	}
	_, err := executor.ExecContext(ctx, `INSERT OR IGNORE INTO effect_events
		(event_id, effect_id, sequence, claim_id, grant_id, kind, state, error_code, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, event.ID, event.EffectID, event.Sequence, event.ClaimID,
		event.GrantID, event.Kind, event.State, event.ErrorCode, event.At.UTC().UnixNano())
	return err
}

func validState(state State) bool {
	switch state {
	case StatePending, StateClaimed, StateSent, StateFailed, StateUncertain, StateBlocked, StateCancelled:
		return true
	default:
		return false
	}
}

func validTerminalState(state State) bool {
	return state == StateSent || state == StateUncertain || state == StateBlocked || state == StateCancelled
}

func mapStateEvent(state State, result ResultCode) EventKind {
	switch state {
	case StateSent:
		if result == ResultReconciled {
			return EventReconciled
		}
		return EventSent
	case StateFailed:
		return EventFailed
	case StateUncertain:
		return EventUncertain
	case StateBlocked:
		return EventBlocked
	default:
		return EventBlocked
	}
}

func randomID(prefix string) (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate claim identity: %w", err)
	}
	return prefix + hex.EncodeToString(buffer), nil
}

func (s *Store) beginImmediate(ctx context.Context) (*sql.Conn, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

func rollback(conn *sql.Conn) {
	_, _ = conn.ExecContext(context.Background(), `ROLLBACK`)
	_ = conn.Close()
}

func commit(conn *sql.Conn) error {
	if _, err := conn.ExecContext(context.Background(), `COMMIT`); err != nil {
		return err
	}
	return conn.Close()
}
