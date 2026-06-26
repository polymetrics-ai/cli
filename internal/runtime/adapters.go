package runtime

import (
	"context"
	"errors"

	"polymetrics.ai/internal/coordination"
	internalledger "polymetrics.ai/internal/ledger"
)

type DragonflyLeaseStore struct {
	dragonfly *coordination.Dragonfly
}

func NewDragonflyLeaseStore(dragonfly *coordination.Dragonfly) *DragonflyLeaseStore {
	return &DragonflyLeaseStore{dragonfly: dragonfly}
}

func OpenDragonflyLeaseStore(addr string) *DragonflyLeaseStore {
	return NewDragonflyLeaseStore(coordination.OpenDragonfly(addr))
}

func (s *DragonflyLeaseStore) Close() error {
	if s == nil || s.dragonfly == nil {
		return nil
	}
	return s.dragonfly.Close()
}

func (s *DragonflyLeaseStore) Ping(ctx context.Context) error {
	if s == nil || s.dragonfly == nil {
		return errors.New("runtime dragonfly lease store is not configured")
	}
	return s.dragonfly.Ping(ctx)
}

func (s *DragonflyLeaseStore) Acquire(ctx context.Context, req LeaseRequest) (Lease, error) {
	if s == nil || s.dragonfly == nil {
		return nil, errors.New("runtime dragonfly lease store is not configured")
	}
	acquired, err := s.dragonfly.AcquireLease(ctx, req.Key, req.Value, req.TTL)
	if err != nil {
		return nil, err
	}
	if !acquired {
		return nil, ErrLeaseDenied
	}
	return dragonflyLease{dragonfly: s.dragonfly, key: req.Key}, nil
}

type dragonflyLease struct {
	dragonfly *coordination.Dragonfly
	key       string
}

func (l dragonflyLease) Release(ctx context.Context) error {
	if l.dragonfly == nil {
		return errors.New("runtime dragonfly lease is not configured")
	}
	return l.dragonfly.ReleaseLease(ctx, l.key)
}

type PostgresRunLedger struct {
	postgres *internalledger.PostgresLedger
}

func NewPostgresRunLedger(postgres *internalledger.PostgresLedger) *PostgresRunLedger {
	return &PostgresRunLedger{postgres: postgres}
}

func OpenPostgresRunLedger(ctx context.Context, postgresURL string) (*PostgresRunLedger, error) {
	postgres, err := internalledger.OpenPostgres(ctx, postgresURL)
	if err != nil {
		return nil, err
	}
	return NewPostgresRunLedger(postgres), nil
}

func (l *PostgresRunLedger) Close() {
	if l != nil && l.postgres != nil {
		l.postgres.Close()
	}
}

func (l *PostgresRunLedger) Migrate(ctx context.Context) error {
	if l == nil || l.postgres == nil {
		return errors.New("runtime postgres run ledger is not configured")
	}
	return l.postgres.Migrate(ctx)
}

func (l *PostgresRunLedger) Append(ctx context.Context, record RunRecord) error {
	if l == nil || l.postgres == nil {
		return errors.New("runtime postgres run ledger is not configured")
	}
	return l.postgres.Append(ctx, internalledger.RunRecord{
		ID:             record.ID,
		Mode:           record.Mode,
		Operation:      record.Operation,
		Status:         record.Status,
		RecordsRead:    record.RecordsRead,
		RecordsWritten: record.RecordsWritten,
		Duration:       record.Duration,
		CreatedAt:      record.CreatedAt,
	})
}
