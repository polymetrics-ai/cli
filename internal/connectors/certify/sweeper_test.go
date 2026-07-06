package certify_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/certify"
)

// TestSweeperCleansUnledgeredAgedEntries proves --sweep's core behavior
// (design §C "Orphan sweeper: ledger entries without cleaned_at ... cleanup
// through the same plan/approve/run path"): an aged, uncleaned ledger entry
// gets cleaned and RecordCleaned-ed by the sweeper.
func TestSweeperCleansUnledgeredAgedEntries(t *testing.T) {
	root := t.TempDir()
	if err := initSweeperProject(t, root); err != nil {
		t.Fatalf("init sweeper project: %v", err)
	}

	ledger, err := certify.NewLedger(root)
	if err != nil {
		t.Fatalf("NewLedger() error = %v", err)
	}
	agedTag := "pm-cert-sample-aged0001-1700000000"
	if err := ledger.RecordPlanned(certify.LedgerEntry{
		Action:     "create",
		Tag:        agedTag,
		Connector:  "sample",
		EntityHint: "outbox_record",
		PlannedAt:  time.Now().UTC().Add(-48 * time.Hour),
	}); err != nil {
		t.Fatalf("RecordPlanned() error = %v", err)
	}

	sweeper := certify.NewSweeper(certify.SweeperOptions{
		Root:      root,
		OlderThan: 24 * time.Hour,
	})
	result, err := sweeper.Sweep(context.Background())
	if err != nil {
		t.Fatalf("Sweep() error = %v", err)
	}
	if result.Scanned == 0 {
		t.Errorf("SweepResult.Scanned = 0, want >0")
	}
	found := false
	for _, swept := range result.Cleaned {
		if swept == agedTag {
			found = true
		}
	}
	if !found {
		t.Errorf("SweepResult.Cleaned = %v, want to include aged tag %q", result.Cleaned, agedTag)
	}

	entries, err := certify.LoadLedger(root)
	if err != nil {
		t.Fatalf("LoadLedger() error = %v", err)
	}
	status, ok := entries.StatusFor(agedTag)
	if !ok || !status.Cleaned {
		t.Errorf("ledger StatusFor(%q) = %+v, ok=%v, want Cleaned=true after sweep", agedTag, status, ok)
	}
}

// TestSweeperSkipsRecentEntries proves the --older-than threshold: a
// recently-planned, uncleaned entry is left alone (it may still be mid-run).
func TestSweeperSkipsRecentEntries(t *testing.T) {
	root := t.TempDir()
	if err := initSweeperProject(t, root); err != nil {
		t.Fatalf("init sweeper project: %v", err)
	}

	ledger, err := certify.NewLedger(root)
	if err != nil {
		t.Fatalf("NewLedger() error = %v", err)
	}
	recentTag := "pm-cert-sample-recent01-1751450000"
	if err := ledger.RecordPlanned(certify.LedgerEntry{
		Action:    "create",
		Tag:       recentTag,
		Connector: "sample",
		PlannedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("RecordPlanned() error = %v", err)
	}

	sweeper := certify.NewSweeper(certify.SweeperOptions{
		Root:      root,
		OlderThan: 24 * time.Hour,
	})
	result, err := sweeper.Sweep(context.Background())
	if err != nil {
		t.Fatalf("Sweep() error = %v", err)
	}
	for _, swept := range result.Cleaned {
		if swept == recentTag {
			t.Errorf("Sweep() cleaned recent tag %q, want it left alone (not yet aged)", recentTag)
		}
	}

	entries, err := certify.LoadLedger(root)
	if err != nil {
		t.Fatalf("LoadLedger() error = %v", err)
	}
	status, ok := entries.StatusFor(recentTag)
	if !ok {
		t.Fatalf("StatusFor(%q) not found", recentTag)
	}
	if status.Cleaned {
		t.Errorf("StatusFor(%q).Cleaned = true, want false (not aged past threshold)", recentTag)
	}
}

// TestSweeperNoOpOnCleanLedger proves a ledger with only cleaned entries
// (or no entries at all) is a pure no-op.
func TestSweeperNoOpOnCleanLedger(t *testing.T) {
	root := t.TempDir()
	if err := initSweeperProject(t, root); err != nil {
		t.Fatalf("init sweeper project: %v", err)
	}

	sweeper := certify.NewSweeper(certify.SweeperOptions{Root: root, OlderThan: time.Hour})
	result, err := sweeper.Sweep(context.Background())
	if err != nil {
		t.Fatalf("Sweep() error = %v", err)
	}
	if len(result.Cleaned) != 0 {
		t.Errorf("Sweep() on empty ledger Cleaned = %v, want empty", result.Cleaned)
	}
}

// initSweeperProject initializes a minimal pm project (via `pm init`,
// through the CLI harness itself, mirroring how certify.Runner sets up its
// own ephemeral root) at root so the sweeper has a valid --root to operate
// against for cleanup CLI calls.
func initSweeperProject(t *testing.T, root string) error {
	t.Helper()
	h := certify.NewHarness(root)
	res := h.Run("init", "--json")
	if res.ExitCode != 0 {
		return os.ErrInvalid
	}
	_ = filepath.Join(root, ".polymetrics")
	return nil
}
