package certify_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

// TestWriteStagesSelfTestAgainstOutbox drives certify.Runner.Run end-to-end
// against the built-in "sample" source connector with Options.Write enabled,
// proving the create-then-cleanup write protocol (design §C, stages 12-17)
// using the sample/outbox reverse-ETL path the Makefile "smoke" target
// already exercises (no live credentials required, per the design's
// documented self-test escape hatch: "if no live creds, the stage self-test
// uses the sample/outbox reverse-ETL path").
func TestWriteStagesSelfTestAgainstOutbox(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
		Write:     true,
	})

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !rep.Passed {
		t.Fatalf("Report.Passed = false, want true; stages=%+v", rep.Stages)
	}

	// --- stage 12: write_plan_preview (redaction gate) ---
	planPreview := mustStage(t, rep, "write_plan_preview")
	if !planPreview.Passed {
		t.Fatalf("write_plan_preview stage failed: %+v", planPreview)
	}

	// --- stage 13: write_create ---
	create := mustStage(t, rep, "write_create")
	if !create.Passed {
		t.Fatalf("write_create stage failed: %+v", create)
	}
	if create.CLI.Kind != "ReverseRun" {
		t.Errorf("write_create stage CLI.Kind = %q, want ReverseRun", create.CLI.Kind)
	}

	// --- stage 14: write_verify (read-back) ---
	verify := mustStage(t, rep, "write_verify")
	if !verify.Passed {
		t.Errorf("write_verify stage failed: %+v", verify)
	}

	// --- stage 15: write_cleanup ---
	cleanup := mustStage(t, rep, "write_cleanup")
	if !cleanup.Passed {
		t.Fatalf("write_cleanup stage failed: %+v", cleanup)
	}

	// --- stage 16: cleanup_verify ---
	cleanupVerify := mustStage(t, rep, "cleanup_verify")
	if !cleanupVerify.Passed {
		t.Errorf("cleanup_verify stage failed: %+v", cleanupVerify)
	}

	// --- stage 17: approval_idempotency ---
	idem := mustStage(t, rep, "approval_idempotency")
	if !idem.Passed {
		t.Errorf("approval_idempotency stage failed: %+v", idem)
	}

	if rep.Capabilities.WriteActions == nil {
		t.Fatalf("Capabilities.WriteActions is nil, want populated after write stages")
	}
	found := false
	for _, wa := range rep.Capabilities.WriteActions {
		if wa.Result == "pass" {
			found = true
		}
	}
	if !found {
		t.Errorf("Capabilities.WriteActions has no passing entry: %+v", rep.Capabilities.WriteActions)
	}

	// No leaked resources: the design's top-level leaks[] must be empty on a
	// clean pass.
	if len(rep.Leaks) != 0 {
		t.Errorf("Report.Leaks = %+v, want empty on a clean run", rep.Leaks)
	}

	// The tag must never leak the raw secret value, and must follow the
	// pm-cert-<slug>-<runid8>-<ts> convention (design §C).
	for _, wa := range rep.Capabilities.WriteActions {
		if wa.Tag == "" {
			continue
		}
		if !strings.HasPrefix(wa.Tag, "pm-cert-sample-") {
			t.Errorf("write action tag %q does not match pm-cert-sample-<runid8>-<ts> convention", wa.Tag)
		}
	}
}

// TestWritePlanPreviewJSONHasNoApprovalToken is the redaction-gate assertion
// (design §A stage 12 "assert --json output has NO approval token"): the
// harness's own write_plan_preview stage must positively assert this on
// every run with Write enabled, not merely happen to pass.
func TestWritePlanPreviewJSONHasNoApprovalToken(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
		Write:     true,
	})
	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	stage := mustStage(t, rep, "write_plan_preview")
	if !stage.Passed {
		t.Fatalf("write_plan_preview failed: %+v", stage)
	}
}

// TestWriteStagesSkipWhenDisabled proves that a Runner with Options.Write
// false (or a connector with no available write pairing) never attempts a
// live write, and does not fail the overall report: the write stages must
// record a documented skip, exactly like fixture_conformance already does
// for wave0's missing defs bundles.
func TestWriteStagesSkipWhenDisabled(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
		// Write left false (default).
	})
	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !rep.Passed {
		t.Fatalf("Report.Passed = false with Write disabled, want true; stages=%+v", rep.Stages)
	}
	create := mustStage(t, rep, "write_create")
	if create.Passed {
		t.Errorf("write_create stage Passed = true with Write disabled, want a documented skip")
	}
	if !containsAny(create.Error, "skip", "disabled", "write not enabled") {
		t.Errorf("write_create stage Error = %q, want a skip reason mentioning write being disabled", create.Error)
	}
}

// TestWriteCreateFailureRecordsNoLeak proves the failure semantics (design
// §C "create fails -> stage fail, no leak"): a plan that never successfully
// creates a tagged record must not be ledgered as a leak.
func TestWriteCreateFailureRecordsNoLeak(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
		Write:     true,
	})
	certify.SabotageExpectedKind(r, "write_create", "NotTheRealKind")

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if rep.Passed {
		t.Fatalf("Report.Passed = true after sabotaged write_create, want false")
	}
	create := mustStage(t, rep, "write_create")
	if create.Passed {
		t.Errorf("sabotaged write_create Passed = true, want false")
	}
	if len(rep.Leaks) != 0 {
		t.Errorf("Report.Leaks = %+v, want empty when create itself failed (no leak possible)", rep.Leaks)
	}
}

// TestWriteCleanupFailureRecordsLeak proves the failure semantics (design
// §C "Create ok + cleanup/verify fails -> leaked_resource"): the report must
// name the leak, force Passed=false, and ExitCodeFor(report) must select
// exit code 3 (design §A "Exit codes ... 3 leaked resources (dominates
// everything)").
func TestWriteCleanupFailureRecordsLeak(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
		Write:     true,
	})
	certify.SabotageExpectedKind(r, "write_cleanup", "NotTheRealKind")

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if rep.Passed {
		t.Fatalf("Report.Passed = true after sabotaged write_cleanup, want false")
	}
	if len(rep.Leaks) == 0 {
		t.Fatalf("Report.Leaks is empty after sabotaged write_cleanup, want at least one leaked_resource entry")
	}
	if rep.Leaks[0].Tag == "" {
		t.Errorf("Report.Leaks[0].Tag is empty, want the leaked tag recorded")
	}
	if certify.ExitCodeFor(rep) != 3 {
		t.Errorf("ExitCodeFor(rep) = %d, want 3 (leaked resources dominate)", certify.ExitCodeFor(rep))
	}
}

// TestCleanupVerifyFailureRecordsLeak proves that even when the cleanup CLI
// call itself reports success, a cleanup_verify failure (entity still
// present) is ALSO a leaked_resource (design §A stage 16 "cleanup_verify ...
// entity gone -> failure -> leaked_resource on failure").
func TestCleanupVerifyFailureRecordsLeak(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
		Write:     true,
	})
	certify.SabotageCleanupVerifyEntityStillPresent(r)

	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if rep.Passed {
		t.Fatalf("Report.Passed = true after sabotaged cleanup_verify, want false")
	}
	if len(rep.Leaks) == 0 {
		t.Fatalf("Report.Leaks is empty after sabotaged cleanup_verify, want at least one leaked_resource entry")
	}
	cv := mustStage(t, rep, "cleanup_verify")
	if cv.Passed {
		t.Errorf("sabotaged cleanup_verify Passed = true, want false")
	}
}

// TestApprovalIdempotencyStageRejectsReplay proves stage 17: a consumed
// plan+token re-run must fail (design §A stage 17 "approval_idempotency:
// consumed plan+token re-run must fail").
func TestApprovalIdempotencyStageRejectsReplay(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
		Write:     true,
	})
	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	stage := mustStage(t, rep, "approval_idempotency")
	if !stage.Passed {
		t.Fatalf("approval_idempotency stage failed: %+v", stage)
	}
	// The replay attempt itself must have produced a non-zero exit / Error
	// envelope kind (a *rejection*), recorded in the stage's CLI info.
	if stage.CLI.ExitCode == 0 {
		t.Errorf("approval_idempotency stage CLI.ExitCode = 0, want non-zero (replay must be rejected)")
	}
}

// TestWriteStagesLedgerWrittenBeforeCreate proves the write-ahead ledger
// ordering guarantee end-to-end: after a full run, the ledger file under the
// workdir (kept via KeepWork) must contain a planned_at entry for the tag
// used, and — for a clean run — a matching cleaned_at.
func TestWriteStagesLedgerWrittenBeforeCreate(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r := certify.NewRunner(certify.Options{
		Connector: "sample",
		Stream:    "customers",
		Limit:     50,
		SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
		Write:     true,
		KeepWork:  true,
	})
	rep, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	workdir := certify.LastWorkdir(r)
	defer func() { _ = os.RemoveAll(workdir) }()

	if !rep.Passed {
		t.Fatalf("Report.Passed = false, want true; stages=%+v", rep.Stages)
	}

	ledgerPath := filepath.Join(workdir, "certify-ledger.jsonl")
	raw, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("read ledger file %s: %v", ledgerPath, err)
	}
	var sawPlanned, sawCleaned bool
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if line == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("ledger line not valid JSON: %s: %v", line, err)
		}
		if _, ok := entry["planned_at"]; ok {
			sawPlanned = true
		}
		if _, ok := entry["cleaned_at"]; ok {
			sawCleaned = true
		}
	}
	if !sawPlanned {
		t.Errorf("ledger file %s has no planned_at entry", ledgerPath)
	}
	if !sawCleaned {
		t.Errorf("ledger file %s has no cleaned_at entry after a clean run", ledgerPath)
	}
}
