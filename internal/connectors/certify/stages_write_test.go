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

func TestSampleOutboxWriteLifecycleAgainstRealCLI(t *testing.T) {
	root := t.TempDir()
	h := certify.NewHarness(root)
	mustHarnessKind(t, h, "InitResult", "init", "--json")

	tag := "pm-cert-sample-real-write"
	seedPath := filepath.Join(root, ".polymetrics", "warehouse", "cert_write_seed_sample.jsonl")
	if err := os.MkdirAll(filepath.Dir(seedPath), 0o700); err != nil {
		t.Fatalf("mkdir seed warehouse: %v", err)
	}
	seed, err := json.Marshal(map[string]string{"id": tag, "tag": tag})
	if err != nil {
		t.Fatalf("marshal seed record: %v", err)
	}
	if err := os.WriteFile(seedPath, append(seed, '\n'), 0o600); err != nil {
		t.Fatalf("write seed record: %v", err)
	}

	mustHarnessKind(t, h, "Credential", "credentials", "add", "cert-outbox", "--connector", "outbox",
		"--config", "path="+filepath.Join(root, ".polymetrics", "outbox"), "--json")

	createPlan, createApproval := planOutboxLifecycle(t, h, "create")
	preview := mustHarnessKind(t, h, "ReversePlanPreview", "reverse", "preview", createPlan, "--json")
	if hits := certify.ScanForSecrets(preview.Stdout, []string{createApproval}); len(hits) != 0 {
		t.Fatalf("reverse preview leaked approval token: %v", hits)
	}
	if plan, _ := preview.Envelope["plan"].(map[string]any); plan != nil {
		if approval, _ := plan["approval_token"].(string); approval != "" {
			t.Fatalf("reverse preview approval_token = %q, want redacted", approval)
		}
	}

	create := mustHarnessKind(t, h, "ReverseRun", "reverse", "run", createPlan, "--approve", createApproval, "--json")
	if run, _ := create.Envelope["run"].(map[string]any); run == nil || run["records_succeeded"] != float64(1) {
		t.Fatalf("create run = %+v, want one successful record", run)
	}
	if got := outboxActionForTag(t, root, tag); got != "create" {
		t.Fatalf("outbox action after create = %q, want create", got)
	}

	cleanupPlan, cleanupApproval := planOutboxLifecycle(t, h, "delete")
	mustHarnessKind(t, h, "ReverseRun", "reverse", "run", cleanupPlan, "--approve", cleanupApproval, "--json")
	if got := outboxActionForTag(t, root, tag); got != "delete" {
		t.Fatalf("outbox action after cleanup = %q, want delete", got)
	}

	replay := h.Run("reverse", "run", createPlan, "--approve", createApproval, "--json")
	if replay.ExitCode == 0 || replay.Kind != "Error" {
		t.Fatalf("replayed approval = exit %d kind %q, want rejected Error", replay.ExitCode, replay.Kind)
	}
}

func mustHarnessKind(t *testing.T, h *certify.Harness, kind string, args ...string) certify.CLIResult {
	t.Helper()
	res := h.Run(args...)
	if err := h.MustKind(res, kind, 0); err != nil {
		t.Fatalf("%s: stdout=%s stderr=%s", err, res.Stdout, res.Stderr)
	}
	return res
}

func planOutboxLifecycle(t *testing.T, h *certify.Harness, action string) (string, string) {
	t.Helper()
	res := h.Run("reverse", "plan", "cert_write_selftest",
		"--source-table", "cert_write_seed_sample",
		"--destination", "outbox:cert-outbox",
		"--map", "id:external_id",
		"--map", "tag:tag",
		"--action", action)
	if res.ExitCode != 0 {
		t.Fatalf("reverse plan %s: exit %d", action, res.ExitCode)
	}
	return reversePlanField(t, res.Stdout, "Created reverse plan "), reversePlanField(t, res.Stdout, "Approval token: ")
}

func reversePlanField(t *testing.T, stdout, prefix string) string {
	t.Helper()
	for _, line := range strings.Split(stdout, "\n") {
		if value, ok := strings.CutPrefix(line, prefix); ok && strings.TrimSpace(value) != "" {
			return strings.Fields(value)[0]
		}
	}
	t.Fatalf("reverse plan output missing %q", prefix)
	return ""
}

func outboxActionForTag(t *testing.T, root, tag string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, ".polymetrics", "outbox", "cert_write_selftest.jsonl"))
	if err != nil {
		t.Fatalf("read outbox records: %v", err)
	}
	last := ""
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatalf("decode outbox record %q: %v", line, err)
		}
		if rowTag, _ := row["tag"].(string); rowTag == tag {
			last, _ = row["_outbox_action"].(string)
		}
	}
	return last
}

// TestWritePlanPreviewJSONHasNoApprovalToken is the JSON token-omission assertion
// (design §A stage 12 "assert --json output has NO approval token"): the
// harness's own write_plan_preview stage must positively assert this on
// every run with Write enabled, not merely happen to pass.
func TestWritePlanPreviewJSONHasNoApprovalToken(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r, driver := scriptedSampleRunner(t, certify.Options{
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
	driver.assertProtocol(t)
	stage := mustStage(t, rep, "write_plan_preview")
	if !stage.Passed {
		t.Fatalf("write_plan_preview failed: %+v", stage)
	}
}

func TestWritePlanPreviewErrorsExcludeApprovalToken(t *testing.T) {
	for _, tt := range []struct {
		name   string
		mutate func(*scriptedCLI)
	}{
		{
			name: "missing plan id",
			mutate: func(driver *scriptedCLI) {
				driver.reversePlanOutput = "Approval token: scripted-approval-1\n"
			},
		},
		{
			name: "preview output",
			mutate: func(driver *scriptedCLI) {
				driver.previewOutputLeak = "scripted-approval-1"
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

			r, driver := scriptedSampleRunner(t, certify.Options{
				Connector: "sample",
				Stream:    "customers",
				Limit:     50,
				SecretEnv: map[string]string{"token": "PM_SAMPLE_TOKEN"},
				Write:     true,
			})
			tt.mutate(driver)

			rep, err := r.Run(context.Background())
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			stage := mustStage(t, rep, "write_plan_preview")
			if stage.Passed {
				t.Fatal("write_plan_preview stage Passed = true, want false")
			}
			if strings.Contains(stage.Error, "scripted-approval-1") {
				t.Error("write_plan_preview error exposed an approval token")
			}
		})
	}
}

// TestWriteStagesSkipWhenDisabled proves that a Runner with Options.Write
// false (or a connector with no available write pairing) never attempts a
// live write, and does not fail the overall report: the write stages must
// record a documented skip, exactly like fixture_conformance already does
// for wave0's missing defs bundles.
func TestWriteStagesSkipWhenDisabled(t *testing.T) {
	t.Setenv("PM_SAMPLE_TOKEN", "sample-cert-token")

	r, driver := scriptedSampleRunner(t, certify.Options{
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
	driver.assertProtocol(t)
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

	r, _ := scriptedSampleRunner(t, certify.Options{
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

	r, _ := scriptedSampleRunner(t, certify.Options{
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

	r, _ := scriptedSampleRunner(t, certify.Options{
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

	r, driver := scriptedSampleRunner(t, certify.Options{
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
	driver.assertProtocol(t)
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

	r, driver := scriptedSampleRunner(t, certify.Options{
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
	driver.assertProtocol(t)
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
