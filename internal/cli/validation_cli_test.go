package cli_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/cli"
)

func TestMalformedConfigFlagIsValidationError(t *testing.T) {
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"credentials", "add", "warehouse-local", "--connector", "warehouse", "--config", "path", "--root", root, "--json"}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("code = %d, want validation error; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"category": "validation"`) {
		t.Fatalf("stdout missing validation category: %s", stdout.String())
	}
}

func TestWarehouseCredentialExternalPathRequiresOptIn(t *testing.T) {
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	external := filepath.Join(t.TempDir(), "warehouse")

	var deniedStdout, deniedStderr bytes.Buffer
	code := cli.Run([]string{"credentials", "add", "warehouse-local", "--connector", "warehouse", "--config", "path=" + external, "--root", root, "--json"}, &deniedStdout, &deniedStderr)
	if code != 3 {
		t.Fatalf("code = %d, want validation error; stdout=%s stderr=%s", code, deniedStdout.String(), deniedStderr.String())
	}

	var allowedStdout, allowedStderr bytes.Buffer
	code = cli.Run([]string{"credentials", "add", "warehouse-local", "--connector", "warehouse", "--config", "path=" + external, "--config", "allow_external_path=true", "--root", root, "--json"}, &allowedStdout, &allowedStderr)
	if code != 0 {
		t.Fatalf("opt-in code = %d, want success; stdout=%s stderr=%s", code, allowedStdout.String(), allowedStderr.String())
	}
}

func TestFileCredentialAllowsExternalReadPath(t *testing.T) {
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	external := filepath.Join(t.TempDir(), "source.jsonl")

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"credentials", "add", "file-local", "--connector", "file", "--config", "path=" + external, "--root", root, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code = %d, want external file source allowed; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestMalformedMapFlagIsValidationError(t *testing.T) {
	root := setupReverseCLIProject(t)

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"reverse", "plan", "bad_plan", "--source-table", "sample_customers", "--destination", "outbox:outbox-local", "--map", "id", "--root", root, "--json"}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("code = %d, want validation error; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"category": "validation"`) {
		t.Fatalf("stdout missing validation category: %s", stdout.String())
	}
}

func TestMalformedIntegerFlagIsValidationError(t *testing.T) {
	root := setupReverseCLIProject(t)

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"query", "run", "--table", "sample_customers", "--limit", "not-a-number", "--root", root, "--json"}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("code = %d, want validation error; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"category": "validation"`) {
		t.Fatalf("stdout missing validation category: %s", stdout.String())
	}
}
