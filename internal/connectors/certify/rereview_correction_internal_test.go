package certify

import (
	"bytes"
	"context"
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

func TestRereviewApprovalReplaySuccessIsCoveredByFinalCleanup(t *testing.T) {
	oldRun := cliRun
	oldRunContext := cliRunContext
	t.Cleanup(func() {
		cliRun = oldRun
		cliRunContext = oldRunContext
	})

	var originalToken string
	var createApproved bool
	var replaySucceeded atomic.Bool
	runOriginal := func(ctx context.Context, args []string, stdout, stderr io.Writer, opts CLIInvocationOptions) int {
		if oldRunContext != nil {
			return oldRunContext(ctx, args, stdout, stderr, opts)
		}
		return oldRun(args, stdout, stderr)
	}
	runIntercepted := func(ctx context.Context, args []string, stdout, stderr io.Writer, opts CLIInvocationOptions) int {
		if hasArgs(args, "reverse", "plan") && flagValue(args, "--action") == "create" && originalToken == "" {
			var captured bytes.Buffer
			code := runOriginal(ctx, args, &captured, stderr, opts)
			out := captured.String()
			_, _ = io.WriteString(stdout, out)
			originalToken = firstMatch(approvalTokenLinePattern, out)
			return code
		}
		if hasArgs(args, "reverse", "run") && originalToken != "" && flagValue(args, "--approve") == originalToken {
			if !createApproved {
				createApproved = true
				return runOriginal(ctx, args, stdout, stderr, opts)
			}
			root := flagValue(args, "--root")
			tag, err := firstOutboxTag(root)
			if err != nil || tag == "" {
				_, _ = fmt.Fprintf(stderr, "find outbox tag: %v", err)
				return 1
			}
			if err := appendOutboxAction(root, tag, "create"); err != nil {
				_, _ = fmt.Fprintf(stderr, "append replay create: %v", err)
				return 1
			}
			replaySucceeded.Store(true)
			_, _ = fmt.Fprint(stdout, `{"kind":"ReverseRun","run":{"records_succeeded":1,"records_failed":0}}`)
			return 0
		}
		return runOriginal(ctx, args, stdout, stderr, opts)
	}
	cliRun = func(args []string, stdout, stderr io.Writer) int {
		return runIntercepted(context.Background(), args, stdout, stderr, CLIInvocationOptions{})
	}
	cliRunContext = func(ctx context.Context, args []string, stdout, stderr io.Writer, opts CLIInvocationOptions) int {
		return runIntercepted(ctx, args, stdout, stderr, opts)
	}

	r := NewRunner(Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
		Write:     true,
		KeepWork:  true,
	})
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")
	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	workdir := LastWorkdir(r)
	t.Cleanup(func() { _ = os.RemoveAll(workdir) })
	if !replaySucceeded.Load() {
		t.Fatal("test did not exercise unexpected approval replay success")
	}
	tag := rep.Capabilities.WriteActions["create"].Tag
	lastAction, err := outboxLastActionForTag(workdir, tag)
	if err != nil {
		t.Fatal(err)
	}
	if lastAction != "delete" {
		t.Fatalf("approval replay success left final outbox action %q, want final cleanup action delete", lastAction)
	}
	idem := lastStage(rep.Stages, "approval_idempotency")
	if idem == nil || idem.Passed {
		t.Fatalf("approval_idempotency stage = %+v, want failed evidence for unexpected replay success", idem)
	}
}

func TestRereviewCleanupFailureThenAbsenceProofClearsStaleLeak(t *testing.T) {
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
			return 0
		case hasArgs(args, "reverse", "preview"):
			_, _ = fmt.Fprint(stdout, `{"kind":"ReversePlanPreview","plan":{}}`)
			return 0
		case hasArgs(args, "reverse", "run"):
			_, _ = fmt.Fprint(stdout, `{"kind":"Error","error":{"message":"cleanup failed"}}`)
			return 1
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
	tag := "pm-cert-sample-33333333-1700000000"
	if err := ledger.RecordPlanned(LedgerEntry{
		Connector: "sample", Action: "create", CleanupAction: "delete", Tag: tag,
		RunID: "33333333", EntityHint: "outbox_record", PlannedAt: time.Now().Add(-48 * time.Hour),
	}); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	rc := &runContext{
		ctx: context.Background(), harness: NewHarness(root), root: root,
		opts: Options{Connector: "sample"},
		write: &writeContext{
			pairing: WritePairing{Create: "create", Cleanup: "delete"}, connector: "sample",
			tag: tag, runID8: "33333333", selfTest: true, createPassed: true, ledger: ledger,
		},
	}
	var rep Report
	if err := stageWriteCleanup(rc, &rep); err != nil {
		t.Fatal(err)
	}
	if len(rep.Leaks) == 0 {
		t.Fatal("cleanup failure did not record initial leak; test setup invalid")
	}
	if err := appendOutboxAction(root, tag, "delete"); err != nil {
		t.Fatal(err)
	}
	if err := stageCleanupVerify(rc, &rep); err != nil {
		t.Fatal(err)
	}
	if len(rep.Leaks) != 0 {
		t.Fatalf("cleanup absence proof left stale top-level leaks: %+v", rep.Leaks)
	}
	entry := rep.Capabilities.WriteActions["create"]
	if entry.Result == "pass" || entry.Result == "leaked_resource" {
		t.Fatalf("write action result = %q, want honest non-leaked failure after cleanup call failed", entry.Result)
	}
	cleanup := lastStage(rep.Stages, "write_cleanup")
	if cleanup == nil || cleanup.Passed {
		t.Fatalf("write_cleanup stage = %+v, want original cleanup failure preserved", cleanup)
	}
}

func firstOutboxTag(root string) (string, error) {
	rows, err := readOutboxJSONL(root)
	if err != nil {
		return "", err
	}
	for _, row := range rows {
		if tag, _ := row["tag"].(string); tag != "" {
			return tag, nil
		}
	}
	return "", nil
}

func appendOutboxAction(root, tag, action string) error {
	outboxDir := filepath.Join(root, ".polymetrics", "outbox")
	if err := os.MkdirAll(outboxDir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(outboxDir, writeReverseSelfTestName+".jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, `{"external_id":%q,"tag":%q,"_outbox_action":%q}`+"\n", tag, tag, action)
	return err
}

func lastStage(stages []StageResult, name string) *StageResult {
	for i := len(stages) - 1; i >= 0; i-- {
		if stages[i].Name == name {
			return &stages[i]
		}
	}
	return nil
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
		for i := 0; i < 10001; i++ {
			raw.WriteString("{\"tag\":\"duplicate\"}\n")
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
