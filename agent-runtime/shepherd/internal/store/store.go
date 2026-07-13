package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply supervisor migration: %w", err)
		}
	}
	return nil
}

func (s *Store) AcquireLease(ctx context.Context, runID, owner string, now time.Time, ttl time.Duration) (Lease, error) {
	if runID == "" || owner == "" || ttl <= 0 {
		return Lease{}, errors.New("run, owner, and positive ttl are required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Lease{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var currentOwner string
	var epoch, expiresAt int64
	err = tx.QueryRowContext(ctx, `SELECT owner, epoch, expires_at FROM leases WHERE run_id = ?`, runID).
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
	if _, err := tx.ExecContext(ctx, `INSERT INTO leases(run_id, owner, epoch, expires_at)
        VALUES (?, ?, ?, ?) ON CONFLICT(run_id) DO UPDATE SET owner=excluded.owner,
        epoch=excluded.epoch, expires_at=excluded.expires_at`, runID, owner, epoch, expires.UnixNano()); err != nil {
		return Lease{}, fmt.Errorf("write lease: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Lease{}, err
	}
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

func (s *Store) PutGrant(ctx context.Context, grant domain.Grant) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO grants
        (run_id, repository, issue, capability, epoch) VALUES (?, ?, ?, ?, ?)`,
		grant.RunID, grant.Repository, grant.Issue, grant.Capability, grant.Epoch)
	if err != nil {
		return fmt.Errorf("put grant: %w", err)
	}
	return nil
}

func (s *Store) Enqueue(ctx context.Context, lease Lease, effect Effect, now time.Time) (bool, error) {
	if effect.Key == "" || effect.RunID != lease.RunID || effect.Repository == "" || effect.Issue <= 0 ||
		effect.Target == "" || effect.PayloadHash == "" || effect.Epoch != lease.Epoch ||
		!domain.IsGrantableCapability(effect.Capability) {
		return false, errors.New("effect identity does not match fenced lease")
	}
	if err := s.CheckLease(ctx, lease, now); err != nil {
		return false, err
	}
	var grantCount int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM grants WHERE run_id = ? AND repository = ?
		AND issue = ? AND capability = ? AND epoch = ?`, effect.RunID, effect.Repository, effect.Issue,
		effect.Capability, effect.Epoch).Scan(&grantCount); err != nil {
		return false, fmt.Errorf("check effect grant: %w", err)
	}
	if grantCount != 1 {
		return false, errors.New("no matching capability grant for effect")
	}
	result, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO outbox
		(effect_key, run_id, repository, issue, capability, target, payload_hash, epoch, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, effect.Key, effect.RunID, effect.Repository, effect.Issue,
		effect.Capability, effect.Target, effect.PayloadHash, effect.Epoch, now.UTC().UnixNano())
	if err != nil {
		return false, fmt.Errorf("enqueue effect: %w", err)
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}
