package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
)

func TestBlockAwaitingDecisionTransitionsToBlocked(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	binding := testDelivery("issue-389", 389)
	if err := db.EnsureDelivery(ctx, binding); err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, binding.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	if err := db.FinishAttempt(ctx, binding.ID, "owner", domain.RunAwaitingDecision); err != nil {
		t.Fatal(err)
	}
	if err := db.BlockAwaitingDecision(ctx, binding.ID, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, binding.ID, "next"); err == nil {
		t.Fatal("blocked delivery resumed after stop decision")
	}
}
