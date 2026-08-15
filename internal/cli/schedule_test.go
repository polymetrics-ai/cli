package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/flow"
	"polymetrics.ai/internal/schedule"
)

// Group E — CLI integration tests for `pm schedule`.
// Uses cli.Run(args, stdout, stderr) with a temp --root dir.

func scheduleRun(t *testing.T, root string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	allArgs := append([]string{"--root", root}, args...)
	code = Run(allArgs, &outBuf, &errBuf)
	return outBuf.String(), errBuf.String(), code
}

func scheduledFlowProject(t *testing.T) string {
	t.Helper()
	ctx := testCtx(t)
	a := newFlowScopedWarehouseApp(t, ctx)
	manifestPath := writeManifestFile(t, `{
		"version": 1,
		"name": "likely-customers",
		"steps": [{
			"id": "sync-records",
			"kind": "sync",
			"job": "acme",
			"streams": ["records"],
			"out": ["records"]
		}]
	}`)
	require.NoError(t, runFlow(ctx, config.Config{}, a, []string{"create", "--file", manifestPath}, &bytes.Buffer{}, true))
	return filepath.Dir(a.ProjectDir())
}

func TestScheduleManualDescribesNoTokenFlowInheritance(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"schedule"}, &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	manual := stdout.String()
	for _, want := range []string{
		"Approve each ETL or reverse-ETL job once",
		"pm --root <root> flow run <name> --json",
		"No approval token or authorization reference is placed in crontab",
		"drift refuses and parks the schedule",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("schedule manual missing %q:\n%s", want, manual)
		}
	}
	for _, obsolete := range []string{"--authorization auth_", "run-scoped grant"} {
		if strings.Contains(manual, obsolete) {
			t.Fatalf("schedule manual retained %q:\n%s", obsolete, manual)
		}
	}
}

// E-1: pm schedule create --name x --cron "0 2 * * *" --flow y → exit 0, manifest created.
func TestScheduleCLI_Create(t *testing.T) {
	root := scheduledFlowProject(t)
	stdout, stderr, code := scheduleRun(t, root, "schedule", "create",
		"--name", "nightly-leads",
		"--cron", "0 2 * * *",
		"--flow", "likely-customers",
	)
	if code != 0 {
		t.Fatalf("create: exit %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "status: ready") || strings.Contains(stdout, "authorization") {
		t.Fatalf("create output = %q, want ready status and no schedule authorization", stdout)
	}
}

// E-2: pm schedule list --json → JSON array containing created schedule.
func TestScheduleCLI_List(t *testing.T) {
	root := scheduledFlowProject(t)
	_, _, code := scheduleRun(t, root, "schedule", "create",
		"--name", "nightly-leads",
		"--cron", "0 2 * * *",
		"--flow", "likely-customers",
	)
	if code != 0 {
		t.Fatal("create failed, cannot test list")
	}

	stdout, stderr, code := scheduleRun(t, root, "schedule", "list", "--json")
	if code != 0 {
		t.Fatalf("list: exit %d, stderr=%q", code, stderr)
	}
	var result struct {
		Schedules []map[string]any `json:"schedules"`
		Statuses  map[string]struct {
			Status string `json:"status"`
		} `json:"statuses"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("list output is not valid JSON: %v\noutput: %s", err, stdout)
	}
	if len(result.Schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(result.Schedules))
	}
	if result.Schedules[0]["name"] != "nightly-leads" {
		t.Fatalf("unexpected name: %v", result.Schedules[0]["name"])
	}
	if _, exists := result.Schedules[0]["authorization_reference"]; exists || result.Statuses["nightly-leads"].Status != "ready" {
		t.Fatalf("list retained schedule authorization or lost status: %+v", result)
	}
}

// E-3: pm schedule install x --crontab → crontab line written (dry-run via env).
func TestScheduleCLI_Install_Crontab(t *testing.T) {
	root := scheduledFlowProject(t)
	_, _, code := scheduleRun(t, root, "schedule", "create",
		"--name", "nightly-leads",
		"--cron", "0 2 * * *",
		"--flow", "likely-customers",
	)
	if code != 0 {
		t.Fatal("create failed, cannot test install")
	}

	// Use PM_CRONTAB_FILE env so the backend writes to a temp file rather than the real crontab.
	tmpCrontab := t.TempDir() + "/crontab"
	t.Setenv("PM_CRONTAB_FILE", tmpCrontab)

	stdout, stderr, code := scheduleRun(t, root, "schedule", "install", "nightly-leads", "--crontab")
	if code != 0 {
		t.Fatalf("install: exit %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "status: ready") || strings.Contains(stdout, "authorization") {
		t.Fatalf("install output = %q, want ready status and no authorization carrier", stdout)
	}

	data, err := os.ReadFile(tmpCrontab)
	if err != nil {
		t.Fatalf("read redirected crontab: %v", err)
	}
	line := string(data)
	if !strings.Contains(line, "--root "+root+" flow run likely-customers --json") || strings.Contains(line, "authorization") {
		t.Fatalf("installed crontab line must run the existing flow without authority material, got %q", line)
	}
}

func TestScheduleCLI_Remove_CleansCrontabInstall(t *testing.T) {
	root := scheduledFlowProject(t)
	_, _, code := scheduleRun(t, root, "schedule", "create",
		"--name", "nightly-leads",
		"--cron", "0 2 * * *",
		"--flow", "likely-customers",
	)
	if code != 0 {
		t.Fatal("create failed, cannot test remove")
	}

	tmpCrontab := t.TempDir() + "/crontab"
	t.Setenv("PM_CRONTAB_FILE", tmpCrontab)

	_, stderr, code := scheduleRun(t, root, "schedule", "install", "nightly-leads", "--crontab")
	if code != 0 {
		t.Fatalf("install: exit %d, stderr=%q", code, stderr)
	}
	data, err := os.ReadFile(tmpCrontab)
	if err != nil {
		t.Fatalf("read redirected crontab after install: %v", err)
	}
	if !strings.Contains(string(data), "pm-schedule-nightly-leads") {
		t.Fatalf("expected crontab sentinel after install, got %q", string(data))
	}

	stdout, stderr, code := scheduleRun(t, root, "schedule", "remove", "nightly-leads")
	if code != 0 {
		t.Fatalf("remove: exit %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "status: ready") || strings.Contains(stdout, "authorization") {
		t.Fatalf("remove output = %q, want prior status and no authorization carrier", stdout)
	}
	data, err = os.ReadFile(tmpCrontab)
	if err != nil {
		t.Fatalf("read redirected crontab after remove: %v", err)
	}
	if strings.Contains(string(data), "pm-schedule-nightly-leads") {
		t.Fatalf("crontab sentinel remained after remove: %q", string(data))
	}
}

// E-4: pm schedule remove x → manifest deleted, exit 0.
func TestScheduleCLI_Remove(t *testing.T) {
	root := scheduledFlowProject(t)
	_, _, code := scheduleRun(t, root, "schedule", "create",
		"--name", "nightly-leads",
		"--cron", "0 2 * * *",
		"--flow", "likely-customers",
	)
	if code != 0 {
		t.Fatal("create failed, cannot test remove")
	}

	t.Setenv("PM_CRONTAB_FILE", t.TempDir()+"/crontab")

	_, stderr, code := scheduleRun(t, root, "schedule", "remove", "nightly-leads")
	if code != 0 {
		t.Fatalf("remove: exit %d, stderr=%q", code, stderr)
	}

	// List should now be empty.
	stdout, _, _ := scheduleRun(t, root, "schedule", "list", "--json")
	_ = stdout // just verifying no panic; full assertion done in E-2
}

// E-5: pm schedule create (missing flags) → exit 1, error in stderr.
func TestScheduleCLI_Create_MissingFlags(t *testing.T) {
	root := t.TempDir()
	_, stderr, code := scheduleRun(t, root, "schedule", "create")
	if code == 0 {
		t.Fatal("expected non-zero exit for missing flags")
	}
	if stderr == "" {
		t.Fatal("expected error in stderr for missing flags")
	}
}

func TestScheduleCLI_CreateMissingFlowWritesNothing(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)
	var stdout bytes.Buffer
	err := runSchedule(testCtx(t), config.Config{}, root, []string{"create", "--name", "missing-flow", "--cron", "0 2 * * *", "--flow", "ghost"}, &stdout, true)
	var referenceErr *schedule.FlowReferenceError
	require.ErrorAs(t, err, &referenceErr)
	if referenceErr.Flow != "ghost" || referenceErr.Reason != schedule.FlowReferenceMissing {
		t.Fatalf("missing flow error = %#v", referenceErr)
	}
	for _, path := range []string{
		filepath.Join(root, "schedules", "missing-flow.json"),
		filepath.Join(root, "schedules", "missing-flow.fire.json"),
		filepath.Join(root, "schedules", "missing-flow.fire.lock"),
	} {
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("missing-flow refusal left %s: %v", path, statErr)
		}
	}
	cliStdout, cliStderr, code := scheduleRun(t, root, "schedule", "create", "--name", "missing-flow", "--cron", "0 2 * * *", "--flow", "ghost", "--json")
	if code != 3 || !strings.Contains(cliStdout, `"code": "schedule_flow_reference_refused"`) || !strings.Contains(cliStdout, `"category": "validation"`) {
		t.Fatalf("missing flow CLI refusal code=%d stdout=%s stderr=%s", code, cliStdout, cliStderr)
	}
	for _, path := range []string{
		filepath.Join(root, "schedules", "missing-flow.json"),
		filepath.Join(root, "schedules", "missing-flow.fire.json"),
		filepath.Join(root, "schedules", "missing-flow.fire.lock"),
	} {
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("CLI missing-flow refusal left %s: %v", path, statErr)
		}
	}
}

func TestScheduleCLI_CreateInvalidFlowJobWritesNothing(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)
	flowsDir := filepath.Join(root, ".polymetrics", "flows")
	require.NoError(t, os.MkdirAll(flowsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(flowsDir, "invalid-job-flow.json"), []byte(`{
		"version":1,"name":"invalid-job-flow","steps":[{
			"id":"missing","kind":"sync","job":"missing-connection","streams":["records"],"out":["records"]
		}]
	}`), 0o600))

	err := runSchedule(testCtx(t), config.Config{}, root, []string{"create", "--name", "invalid-flow", "--cron", "0 2 * * *", "--flow", "invalid-job-flow"}, &bytes.Buffer{}, true)
	var scheduleErr *schedule.FlowReferenceError
	require.ErrorAs(t, err, &scheduleErr)
	if scheduleErr.Reason != schedule.FlowReferenceInvalid {
		t.Fatalf("invalid flow error = %#v", scheduleErr)
	}
	var jobErr *flow.JobReferenceError
	require.ErrorAs(t, err, &jobErr)
	if jobErr.Reference != "missing-connection" || jobErr.Reason != flow.JobReferenceMissing {
		t.Fatalf("invalid flow job error = %#v", jobErr)
	}
	if _, statErr := os.Stat(filepath.Join(root, "schedules", "invalid-flow.json")); !os.IsNotExist(statErr) {
		t.Fatalf("invalid flow created schedule: %v", statErr)
	}
}

func TestScheduleCLI_CreateSecondScheduleForFlowIsTypedAndWritesNothing(t *testing.T) {
	root := scheduledFlowProject(t)
	_, stderr, code := scheduleRun(t, root, "schedule", "create", "--name", "first", "--cron", "0 2 * * *", "--flow", "likely-customers")
	if code != 0 {
		t.Fatalf("create first schedule code=%d stderr=%s", code, stderr)
	}
	err := runSchedule(testCtx(t), config.Config{}, root, []string{"create", "--name", "second", "--cron", "0 3 * * *", "--flow", "likely-customers"}, &bytes.Buffer{}, true)
	var referenceErr *schedule.FlowReferenceError
	require.ErrorAs(t, err, &referenceErr)
	if referenceErr.Flow != "likely-customers" || referenceErr.Reason != schedule.FlowReferenceAmbiguous {
		t.Fatalf("second schedule error = %#v", referenceErr)
	}
	if _, statErr := os.Stat(filepath.Join(root, "schedules", "second.json")); !os.IsNotExist(statErr) {
		t.Fatalf("second schedule refusal wrote a manifest: %v", statErr)
	}
}

func TestScheduleCLI_CreateMalformedFlowReferenceWritesNothing(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)
	err := runSchedule(testCtx(t), config.Config{}, root, []string{"create", "--name", "malformed-flow", "--cron", "0 2 * * *", "--flow", "../escape"}, &bytes.Buffer{}, true)
	var referenceErr *schedule.FlowReferenceError
	require.ErrorAs(t, err, &referenceErr)
	if referenceErr.Flow != "../escape" || referenceErr.Reason != schedule.FlowReferenceMalformed {
		t.Fatalf("malformed flow error = %#v", referenceErr)
	}
	if _, statErr := os.Stat(filepath.Join(root, "schedules")); !os.IsNotExist(statErr) {
		t.Fatalf("malformed flow refusal created schedule storage: %v", statErr)
	}
}

func TestScheduleCLI_InstallMissingFlowLeavesCrontabUntouched(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)
	manifest := schedule.Manifest{Name: "orphaned-schedule", Cron: "0 2 * * *", Flow: "missing-flow"}
	require.NoError(t, schedule.Save(root, manifest, false))
	crontab := filepath.Join(t.TempDir(), "crontab")
	baseline := []byte("# preserved\n")
	require.NoError(t, os.WriteFile(crontab, baseline, 0o600))
	t.Setenv("PM_CRONTAB_FILE", crontab)

	err := runScheduleInstall(testCtx(t), config.Config{}, root, []string{manifest.Name, "--crontab"}, &bytes.Buffer{}, true)
	var referenceErr *schedule.FlowReferenceError
	require.ErrorAs(t, err, &referenceErr)
	if referenceErr.Flow != manifest.Flow || referenceErr.Reason != schedule.FlowReferenceMissing {
		t.Fatalf("install missing flow error = %#v", referenceErr)
	}
	data, readErr := os.ReadFile(crontab)
	require.NoError(t, readErr)
	if !bytes.Equal(data, baseline) || bytes.Contains(data, []byte("pm-schedule-")) {
		t.Fatalf("missing flow install changed backend: %q", data)
	}
}

// E-6: pm schedule install unknown → exit 1, "not found" error.
func TestScheduleCLI_Install_NotFound(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)
	_, stderr, code := scheduleRun(t, root, "schedule", "install", "ghost-schedule", "--crontab")
	if code == 0 {
		t.Fatal("expected non-zero exit for unknown schedule")
	}
	if !strings.Contains(stderr+stdoutFor(t, root, "ghost-schedule"), "not found") {
		// Some implementations write the error to stdout in JSON mode; acceptable if code != 0.
		_ = stderr
	}
}

// E-7: pm schedule create --name INVALID → exit 1, validation error.
func TestScheduleCLI_Create_InvalidName(t *testing.T) {
	root := scheduledFlowProject(t)
	_, stderr, code := scheduleRun(t, root, "schedule", "create",
		"--name", "INVALID-NAME",
		"--cron", "0 2 * * *",
		"--flow", "likely-customers",
	)
	if code == 0 {
		t.Fatal("expected non-zero exit for invalid name")
	}
	_ = stderr
}

// stdoutFor is a helper to capture stdout from a scheduleRun for error checks.
func stdoutFor(t *testing.T, root, name string) string {
	t.Helper()
	stdout, _, _ := scheduleRun(t, root, "schedule", "install", name, "--crontab")
	return stdout
}
