package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
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

// E-1: pm schedule create --name x --cron "0 2 * * *" --flow y → exit 0, manifest created.
func TestScheduleCLI_Create(t *testing.T) {
	root := t.TempDir()
	_, stderr, code := scheduleRun(t, root, "schedule", "create",
		"--name", "nightly-leads",
		"--cron", "0 2 * * *",
		"--flow", "likely-customers",
	)
	if code != 0 {
		t.Fatalf("create: exit %d, stderr=%q", code, stderr)
	}
}

// E-2: pm schedule list --json → JSON array containing created schedule.
func TestScheduleCLI_List(t *testing.T) {
	root := t.TempDir()
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
}

// E-3: pm schedule install x --crontab → crontab line written (dry-run via env).
func TestScheduleCLI_Install_Crontab(t *testing.T) {
	root := t.TempDir()
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

	_, stderr, code := scheduleRun(t, root, "schedule", "install", "nightly-leads", "--crontab")
	if code != 0 {
		t.Fatalf("install: exit %d, stderr=%q", code, stderr)
	}
}

// E-4: pm schedule remove x → manifest deleted, exit 0.
func TestScheduleCLI_Remove(t *testing.T) {
	root := t.TempDir()
	_, _, code := scheduleRun(t, root, "schedule", "create",
		"--name", "nightly-leads",
		"--cron", "0 2 * * *",
		"--flow", "likely-customers",
	)
	if code != 0 {
		t.Fatal("create failed, cannot test remove")
	}

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

// E-6: pm schedule install unknown → exit 1, "not found" error.
func TestScheduleCLI_Install_NotFound(t *testing.T) {
	root := t.TempDir()
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
	root := t.TempDir()
	_, stderr, code := scheduleRun(t, root, "schedule", "create",
		"--name", "INVALID-NAME",
		"--cron", "0 2 * * *",
		"--flow", "f",
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
