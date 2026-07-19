package certify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRereviewCleanupVerificationFailureRemainsSweepEligible(t *testing.T) {
	oldRun := cliRun
	oldRunContext := cliRunContext
	t.Cleanup(func() {
		cliRun = oldRun
		cliRunContext = oldRunContext
	})
	cliRun = func(args []string, stdout, _ io.Writer) int {
		switch {
		case hasArgs(args, "reverse", "plan"):
			_, _ = fmt.Fprint(stdout, "Created reverse plan cleanup-plan\nApproval token: approval-marker\n")
		case hasArgs(args, "reverse", "preview"):
			_, _ = fmt.Fprint(stdout, `{"kind":"ReversePlanPreview","plan":{}}`)
		case hasArgs(args, "reverse", "run"):
			_, _ = fmt.Fprint(stdout, `{"kind":"ReverseRun","run":{"records_succeeded":1,"records_failed":0}}`)
		}
		return 0
	}
	cliRunContext = func(_ context.Context, args []string, stdout, stderr io.Writer, _ CLIInvocationOptions) int {
		return cliRun(args, stdout, stderr)
	}

	ledgerRoot := t.TempDir()
	ledger, err := NewLedger(ledgerRoot)
	if err != nil {
		t.Fatal(err)
	}
	tag := "pm-cert-sample-12345678-1700000000"
	if err := ledger.RecordPlanned(LedgerEntry{
		Connector: "sample", Action: "create", CleanupAction: "delete", Tag: tag,
		RunID: "12345678", EntityHint: "outbox_record", PlannedAt: time.Now().Add(-48 * time.Hour),
	}); err != nil {
		t.Fatal(err)
	}
	rc := &runContext{
		ctx: context.Background(), harness: NewHarness(t.TempDir()), root: t.TempDir(),
		opts: Options{Connector: "sample"}, cleanupVerifySabotage: true,
		write: &writeContext{
			pairing: WritePairing{Create: "create", Cleanup: "delete"}, connector: "sample",
			tag: tag, runID8: "12345678", selfTest: true, createPassed: true, ledger: ledger,
		},
	}
	var rep Report
	if err := stageWriteCleanup(rc, &rep); err != nil {
		t.Fatal(err)
	}
	if err := stageCleanupVerify(rc, &rep); err != nil {
		t.Fatal(err)
	}
	entries, err := LoadLedger(ledgerRoot)
	if err != nil {
		t.Fatal(err)
	}
	status, ok := entries.StatusFor(tag)
	if !ok {
		t.Fatal("planned ledger status missing")
	}
	if status.Cleaned || len(entries.Uncleaned()) != 1 {
		t.Fatalf("verification failure made cleanup non-retryable: cleaned=%v uncleaned=%d", status.Cleaned, len(entries.Uncleaned()))
	}
}

func TestRereviewSweepVerificationFailureRemainsRetryable(t *testing.T) {
	oldRun := cliRun
	oldRunContext := cliRunContext
	t.Cleanup(func() {
		cliRun = oldRun
		cliRunContext = oldRunContext
	})
	cliRun = func(args []string, stdout, _ io.Writer) int {
		switch {
		case hasArgs(args, "init"):
			_, _ = fmt.Fprint(stdout, `{"kind":"InitResult"}`)
		case hasArgs(args, "credentials", "add"):
			_, _ = fmt.Fprint(stdout, `{"kind":"Credential"}`)
		case hasArgs(args, "reverse", "plan"):
			_, _ = fmt.Fprint(stdout, "Created reverse plan cleanup-plan\nApproval token: approval-marker\n")
		case hasArgs(args, "reverse", "preview"):
			_, _ = fmt.Fprint(stdout, `{"kind":"ReversePlanPreview","plan":{}}`)
		case hasArgs(args, "reverse", "run"):
			_, _ = fmt.Fprint(stdout, `{"kind":"ReverseRun","run":{"records_succeeded":1,"records_failed":0}}`)
		default:
			_, _ = fmt.Fprint(stdout, `{"kind":"FixtureResult"}`)
		}
		return 0
	}
	cliRunContext = func(_ context.Context, args []string, stdout, stderr io.Writer, _ CLIInvocationOptions) int {
		return cliRun(args, stdout, stderr)
	}

	ledgerRoot := t.TempDir()
	ledger, err := NewLedger(ledgerRoot)
	if err != nil {
		t.Fatal(err)
	}
	tag := "pm-cert-sample-87654321-1700000000"
	if err := ledger.RecordPlanned(LedgerEntry{
		Connector: "sample", Action: "create", CleanupAction: "delete", Tag: tag,
		RunID: "87654321", EntityHint: "outbox_record", PlannedAt: time.Now().Add(-48 * time.Hour),
	}); err != nil {
		t.Fatal(err)
	}
	result, err := NewSweeper(SweeperOptions{
		LedgerRoot: ledgerRoot, WorkspaceRoot: t.TempDir(), Connector: "sample", OlderThan: time.Hour,
	}).Sweep(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	entries, err := LoadLedger(ledgerRoot)
	if err != nil {
		t.Fatal(err)
	}
	status, ok := entries.StatusFor(tag)
	if !ok {
		t.Fatal("planned ledger status missing")
	}
	if status.Cleaned || len(entries.Uncleaned()) != 1 || len(result.Cleaned) != 0 {
		t.Fatalf("unverified sweep became non-retryable: cleaned=%v uncleaned=%d result=%v", status.Cleaned, len(entries.Uncleaned()), result.Cleaned)
	}
}

func TestRereviewForgedNumericPairingsHaveZeroEffects(t *testing.T) {
	oldRun := cliRun
	oldRunContext := cliRunContext
	var effects atomic.Int64
	cliRun = func(_ []string, _, _ io.Writer) int {
		effects.Add(1)
		return 1
	}
	cliRunContext = func(_ context.Context, args []string, stdout, stderr io.Writer, _ CLIInvocationOptions) int {
		return cliRun(args, stdout, stderr)
	}
	t.Cleanup(func() {
		cliRun = oldRun
		cliRunContext = oldRunContext
	})

	for _, tc := range []struct {
		action, cleanup, entity, verify, runID string
	}{
		{action: "create_issue", cleanup: "close_issue", entity: "issues", verify: "title", runID: "11111111"},
		{action: "create_milestone", cleanup: "delete_milestone", entity: "milestones", verify: "title", runID: "22222222"},
	} {
		t.Run(tc.action, func(t *testing.T) {
			effects.Store(0)
			status := LedgerStatus{
				Tag: "pm-cert-github-" + tc.runID + "-1700000000", Connector: "github",
				Action: tc.action, CleanupAction: tc.cleanup, RunID: tc.runID,
				EntityHint: tc.entity, VerifyField: tc.verify, PlannedAt: time.Now().Add(-48 * time.Hour),
			}
			if ok, _ := sweepCleanTag(NewHarness(t.TempDir()), status); ok {
				t.Fatal("forged numeric cleanup unexpectedly succeeded")
			}
			if effects.Load() != 0 {
				t.Fatalf("forged numeric cleanup reached %d CLI effects", effects.Load())
			}
		})
	}
}

func TestRereviewLedgerInputIsBoundedAndSecretSafe(t *testing.T) {
	t.Run("total bytes", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, ledgerFileName), bytes.Repeat([]byte{' '}, (1<<20)+1), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := LoadLedger(root)
		if err == nil || !strings.Contains(err.Error(), "exceeds") {
			t.Fatal("oversized ledger did not fail with a bounded-input error")
		}
	})

	t.Run("entry count", func(t *testing.T) {
		root := t.TempDir()
		var raw bytes.Buffer
		planned := time.Now().Add(-48 * time.Hour).UTC()
		for i := 0; i < 10001; i++ {
			entry := LedgerEntry{
				Connector: "sample", Action: "create", CleanupAction: "delete", RunID: "12345678",
				Tag: fmt.Sprintf("pm-cert-sample-12345678-%d", 1700000000+i), EntityHint: "outbox_record", PlannedAt: planned,
			}
			line, err := json.Marshal(entry)
			if err != nil {
				t.Fatal(err)
			}
			raw.Write(line)
			raw.WriteByte('\n')
		}
		if err := os.WriteFile(filepath.Join(root, ledgerFileName), raw.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadLedger(root); err == nil {
			t.Fatal("ledger with too many entries was accepted")
		}
	})

	t.Run("malformed line nondisclosure", func(t *testing.T) {
		const marker = "planted-ledger-marker"
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, ledgerFileName), []byte(`{"tag":"`+marker+`"`), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := LoadLedger(root)
		if err == nil {
			t.Fatal("malformed ledger was accepted")
		}
		if strings.Contains(err.Error(), marker) {
			t.Fatal("malformed ledger error reflected planted content")
		}
	})
}
