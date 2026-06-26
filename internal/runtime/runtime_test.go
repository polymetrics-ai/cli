package runtime

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestRecordRunWithLeaseAppendsAndReleases(t *testing.T) {
	ctx := context.Background()
	req := LeaseRequest{Key: "polymetrics:test:run_1", Value: "recording", TTL: time.Second}
	record := RunRecord{
		ID:             "run_1",
		Mode:           "runtime-backed",
		Operation:      "etl",
		Status:         "completed",
		RecordsRead:    10,
		RecordsWritten: 10,
		Duration:       int64(time.Second),
		CreatedAt:      time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC),
	}
	lease := &fakeLease{}
	leases := &fakeLeaseStore{lease: lease}
	ledger := &fakeRunLedger{}
	module := Module{Leases: leases, Ledger: ledger}

	if err := module.RecordRunWithLease(ctx, req, record); err != nil {
		t.Fatalf("RecordRunWithLease() error = %v", err)
	}
	if !reflect.DeepEqual(leases.requests, []LeaseRequest{req}) {
		t.Fatalf("lease requests = %#v, want %#v", leases.requests, []LeaseRequest{req})
	}
	if !reflect.DeepEqual(ledger.records, []RunRecord{record}) {
		t.Fatalf("ledger records = %#v, want %#v", ledger.records, []RunRecord{record})
	}
	if lease.releases != 1 {
		t.Fatalf("lease releases = %d, want 1", lease.releases)
	}
}

func TestRecordRunWithLeaseDeniedLease(t *testing.T) {
	lease := &fakeLease{}
	module := Module{
		Leases: &fakeLeaseStore{lease: lease, err: ErrLeaseDenied},
		Ledger: &fakeRunLedger{},
	}

	err := module.RecordRunWithLease(context.Background(), LeaseRequest{Key: "busy", Value: "recording", TTL: time.Second}, RunRecord{ID: "run_busy"})
	if !errors.Is(err, ErrLeaseDenied) {
		t.Fatalf("RecordRunWithLease() error = %v, want ErrLeaseDenied", err)
	}
	ledger := module.Ledger.(*fakeRunLedger)
	if len(ledger.records) != 0 {
		t.Fatalf("ledger records = %#v, want none", ledger.records)
	}
	if lease.releases != 0 {
		t.Fatalf("lease releases = %d, want 0", lease.releases)
	}
}

func TestRecordRunWithLeaseReleasesAfterAppendError(t *testing.T) {
	appendErr := errors.New("append failed")
	lease := &fakeLease{}
	module := Module{
		Leases: &fakeLeaseStore{lease: lease},
		Ledger: &fakeRunLedger{err: appendErr},
	}

	err := module.RecordRunWithLease(context.Background(), LeaseRequest{Key: "lease", Value: "recording", TTL: time.Second}, RunRecord{ID: "run_error"})
	if !errors.Is(err, appendErr) {
		t.Fatalf("RecordRunWithLease() error = %v, want append error", err)
	}
	if lease.releases != 1 {
		t.Fatalf("lease releases = %d, want 1", lease.releases)
	}
}

type fakeLeaseStore struct {
	lease    Lease
	err      error
	requests []LeaseRequest
}

func (s *fakeLeaseStore) Acquire(ctx context.Context, req LeaseRequest) (Lease, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.requests = append(s.requests, req)
	return s.lease, s.err
}

type fakeLease struct {
	err      error
	releases int
}

func (l *fakeLease) Release(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	l.releases++
	return l.err
}

type fakeRunLedger struct {
	err     error
	records []RunRecord
}

func (l *fakeRunLedger) Append(ctx context.Context, record RunRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if l.err != nil {
		return l.err
	}
	l.records = append(l.records, record)
	return nil
}
