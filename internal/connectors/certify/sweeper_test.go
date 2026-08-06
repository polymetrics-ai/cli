package certify_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/cli"
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

	// The sample/outbox plan-approve-run lifecycle is covered end-to-end by
	// TestWriteStagesSelfTestAgainstOutbox. Keep this test focused on the
	// sweeper's ledger transition and its exact cleanup orchestration, without
	// repeatedly loading every connector bundle through the CLI.
	var calls []string
	certify.SetCLIRunFunc(func(args []string, stdout, _ io.Writer) int {
		call := strings.Join(args, " ")
		calls = append(calls, call)

		switch {
		case strings.HasPrefix(call, "reverse plan "):
			if _, err := io.WriteString(stdout, "Created reverse plan sweep-plan\nApproval token: sweep-approval\n"); err != nil {
				t.Errorf("write fake reverse plan result: %v", err)
				return 1
			}
		case strings.HasPrefix(call, "reverse run "):
			if _, err := io.WriteString(stdout, `{"kind":"ReverseRun"}`); err != nil {
				t.Errorf("write fake reverse run result: %v", err)
				return 1
			}
		}
		return 0
	})
	t.Cleanup(func() { certify.SetCLIRunFunc(cli.Run) })

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
	for i, want := range []string{
		"credentials add cert-outbox",
		"credentials add cert-sweep-warehouse",
		"credentials add cert-sweep-seed-file",
		"connections create cert_sweep_seed_conn",
		"etl run",
		"reverse plan cert_write_selftest",
		"reverse run sweep-plan",
	} {
		if i >= len(calls) || !strings.HasPrefix(calls[i], want) {
			t.Errorf("cleanup call %d = %q, want prefix %q", i, callAt(calls, i), want)
		}
	}
	if len(calls) != 7 {
		t.Errorf("cleanup call count = %d, want 7 (%v)", len(calls), calls)
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

func callAt(calls []string, index int) string {
	if index >= len(calls) {
		return ""
	}
	return calls[index]
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

func TestSweeperUnknownConnectorDoesNotInventCleanup(t *testing.T) {
	root := t.TempDir()
	ledger, err := certify.NewLedger(root)
	if err != nil {
		t.Fatalf("NewLedger() error = %v", err)
	}
	tag := "pm-cert-unknown-aged0001-1700000000"
	if err := ledger.RecordPlanned(certify.LedgerEntry{
		Action:    "create_widget",
		Tag:       tag,
		Connector: "unknown-certification-connector",
		PlannedAt: time.Now().UTC().Add(-48 * time.Hour),
	}); err != nil {
		t.Fatalf("RecordPlanned() error = %v", err)
	}

	result, err := certify.NewSweeper(certify.SweeperOptions{Root: root, OlderThan: 24 * time.Hour}).Sweep(context.Background())
	if err != nil {
		t.Fatalf("Sweep() error = %v", err)
	}
	if len(result.Cleaned) != 0 {
		t.Fatalf("Cleaned = %v, want no cleanup for unknown connector", result.Cleaned)
	}
	if got := result.Failed[tag]; !strings.Contains(got, "no known cleanup mechanism") {
		t.Fatalf("Failed[%q] = %q, want no known cleanup mechanism", tag, got)
	}

	entries, err := certify.LoadLedger(root)
	if err != nil {
		t.Fatalf("LoadLedger() error = %v", err)
	}
	status, ok := entries.StatusFor(tag)
	if !ok || status.Cleaned {
		t.Fatalf("StatusFor(%q) = %+v ok=%v, want uncleaned", tag, status, ok)
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
