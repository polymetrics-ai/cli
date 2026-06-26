package runtime

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrLeaseDenied = errors.New("runtime lease denied")

type LeaseRequest struct {
	Key   string
	Value string
	TTL   time.Duration
}

type RunRecord struct {
	ID             string
	Mode           string
	Operation      string
	Status         string
	RecordsRead    int
	RecordsWritten int
	Duration       int64
	CreatedAt      time.Time
}

type LeaseStore interface {
	Acquire(ctx context.Context, req LeaseRequest) (Lease, error)
}

type Lease interface {
	Release(ctx context.Context) error
}

type RunLedger interface {
	Append(ctx context.Context, record RunRecord) error
}

type Module struct {
	Leases LeaseStore
	Ledger RunLedger
}

func (m Module) RecordRunWithLease(ctx context.Context, leaseReq LeaseRequest, record RunRecord) error {
	if m.Leases == nil {
		return errors.New("runtime leases are not configured")
	}
	if m.Ledger == nil {
		return errors.New("runtime ledger is not configured")
	}
	lease, err := m.Leases.Acquire(ctx, leaseReq)
	if err != nil {
		return fmt.Errorf("acquire runtime lease: %w", err)
	}
	if lease == nil {
		return ErrLeaseDenied
	}

	appendErr := m.Ledger.Append(ctx, record)
	releaseErr := lease.Release(ctx)
	if appendErr != nil {
		if releaseErr != nil {
			return errors.Join(
				fmt.Errorf("append runtime run: %w", appendErr),
				fmt.Errorf("release runtime lease: %w", releaseErr),
			)
		}
		return fmt.Errorf("append runtime run: %w", appendErr)
	}
	if releaseErr != nil {
		return fmt.Errorf("release runtime lease: %w", releaseErr)
	}
	return nil
}
