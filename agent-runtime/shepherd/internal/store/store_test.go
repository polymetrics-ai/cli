package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
)

func TestLeaseFencingAndOutboxIdempotency(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Unix(1_700_000_000, 0).UTC()
	first, err := db.AcquireLease(ctx, "run-1", "owner-a", now, time.Minute)
	if err != nil {
		t.Fatalf("acquire first lease: %v", err)
	}
	second, err := db.AcquireLease(ctx, "run-1", "owner-b", now.Add(2*time.Minute), time.Minute)
	if err != nil {
		t.Fatalf("acquire second lease: %v", err)
	}
	if second.Epoch <= first.Epoch {
		t.Fatalf("fencing epoch did not increase: first=%d second=%d", first.Epoch, second.Epoch)
	}
	if err := db.CheckLease(ctx, first, now.Add(2*time.Minute)); err == nil {
		t.Fatal("expected stale lease to fail")
	}

	grant, err := domain.NewGrant("run-1", "repo", 372, domain.CapabilityPRUpdate, second.Epoch)
	if err != nil {
		t.Fatalf("grant: %v", err)
	}
	if err := db.PutGrant(ctx, grant); err != nil {
		t.Fatalf("put grant: %v", err)
	}

	effect := Effect{
		Key: "pr:372:ready:abc123", RunID: "run-1", Repository: "repo", Issue: 372,
		Capability: domain.CapabilityPRUpdate, Target: "pr:380", PayloadHash: "sha256:deadbeef", Epoch: second.Epoch,
	}
	inserted, err := db.Enqueue(ctx, second, effect, now.Add(2*time.Minute))
	if err != nil || !inserted {
		t.Fatalf("first enqueue: inserted=%v err=%v", inserted, err)
	}
	inserted, err = db.Enqueue(ctx, second, effect, now.Add(2*time.Minute))
	if err != nil || inserted {
		t.Fatalf("duplicate enqueue: inserted=%v err=%v", inserted, err)
	}
}
