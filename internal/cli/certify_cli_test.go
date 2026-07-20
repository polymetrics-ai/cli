package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors/certify"
)

// TestMain wires the real cli.Run entrypoint into certify's in-process CLI
// driver exactly once for this test binary (mirroring cmd/pm/main.go),
// since `pm connectors certify` drives certify.Runner/RunBatch which in turn
// drive cli.Run recursively via certify.Harness (see
// internal/connectors/certify/cliharness.go SetCLIRunFunc).
func TestMain(m *testing.M) {
	certify.SetCLIRunFunc(cli.Run)
	certify.SetCLIRunContextFunc(func(ctx context.Context, args []string, stdout, stderr io.Writer, opts certify.CLIInvocationOptions) int {
		return cli.RunWithContext(ctx, args, stdout, stderr, cli.RunOptions{Mode: cli.ModePlain, ScheduleCrontabFile: opts.CrontabFile})
	})
	os.Exit(m.Run())
}

func certifyRun(t *testing.T, root string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	allArgs := append([]string{"--root", root}, args...)
	code = cli.Run(allArgs, &outBuf, &errBuf)
	return outBuf.String(), errBuf.String(), code
}

// TestCertifyCLISingleConnectorPassExitsZero drives `pm connectors certify
// sample` against the real CLI end-to-end and proves a passing run exits 0
// with a ConnectorCertification JSON envelope (certification design §A).
func TestCertifyCLISingleConnectorPassExitsZero(t *testing.T) {
	t.Setenv("PM_CERT_SAMPLE_TOKEN", "sample-cli-token")
	root := t.TempDir()

	stdout, stderr, code := certifyRun(t, root, "connectors", "certify", "sample",
		"--from-env", "token=PM_CERT_SAMPLE_TOKEN", "--json")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, `"kind": "ConnectorCertification"`) {
		t.Errorf("stdout missing ConnectorCertification envelope: %s", stdout)
	}
	if !strings.Contains(stdout, `"connector": "sample"`) {
		t.Errorf("stdout missing connector=sample: %s", stdout)
	}
}

// TestCertifyCLISingleConnectorTextMode proves the non-JSON rendering path
// also works and reports PASS.
func TestCertifyCLISingleConnectorTextMode(t *testing.T) {
	t.Setenv("PM_CERT_SAMPLE_TOKEN", "sample-cli-token")
	root := t.TempDir()

	stdout, stderr, code := certifyRun(t, root, "connectors", "certify", "sample",
		"--from-env", "token=PM_CERT_SAMPLE_TOKEN")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "Certification: sample [PASS]") {
		t.Errorf("stdout missing human-readable PASS summary: %s", stdout)
	}
}

// TestCertifyCLISingleConnectorSavesReport proves the CLI persists the
// report under <root>/.polymetrics/certifications/<connector>.json
// (certification design §A report artifact path).
func TestCertifyCLISingleConnectorSavesReport(t *testing.T) {
	t.Setenv("PM_CERT_SAMPLE_TOKEN", "sample-cli-token")
	root := t.TempDir()

	_, _, code := certifyRun(t, root, "connectors", "certify", "sample",
		"--from-env", "token=PM_CERT_SAMPLE_TOKEN", "--json")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	path := filepath.Join(root, ".polymetrics", "certifications", "sample.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("report not saved at %s: %v", path, err)
	}
}

// TestCertifyCLIMissingConnectorArgIsUsageError proves `pm connectors
// certify` with no connector name and no --all/--sweep is a usage error, not
// a panic or a certify-mode exit code.
func TestCertifyCLIMissingConnectorArgIsUsageError(t *testing.T) {
	root := t.TempDir()

	_, stderr, code := certifyRun(t, root, "connectors", "certify")

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero for missing connector argument")
	}
	if !strings.Contains(stderr, "error:") {
		t.Errorf("stderr missing error message: %s", stderr)
	}
}

// TestCertifyCLIUnknownConnectorFails proves a connector name not present in
// the registry surfaces as a certify failure with a non-zero exit rather
// than a panic.
func TestCertifyCLIUnknownConnectorFails(t *testing.T) {
	root := t.TempDir()

	_, _, code := certifyRun(t, root, "connectors", "certify", "definitely-not-a-real-connector", "--json")

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero for unknown connector")
	}
}

// TestCertifyCLIAllRequiresCredentialsFile proves --all without
// --credentials-file is a clear usage error (certification design §A
// command spec: "pm connectors certify --all --credentials-file creds.yaml").
func TestCertifyCLIBatchValidatesSafetyControlsBeforeCredentialLoading(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "missing-creds.yaml")
	_, readErr := os.ReadFile(missing)
	if readErr == nil {
		t.Fatal("missing credential fixture unexpectedly exists")
	}

	t.Run("valid controls preserve credential read error bytes", func(t *testing.T) {
		wantStderr := fmt.Sprintf("error: certify: read creds file %s: %v\n", missing, readErr)
		stdout, stderr, code := certifyRun(t, root, "connectors", "certify", "--all", "--credentials-file", missing)
		if code != 1 || stdout != "" || stderr != wantStderr {
			t.Fatalf("exit=%d stdout=%q stderr=%q, want exit 1, empty stdout, stderr %q", code, stdout, stderr, wantStderr)
		}
	})

	t.Run("invalid parallel wins before credential loading", func(t *testing.T) {
		stdout, stderr, code := certifyRun(t, root, "connectors", "certify", "--all", "--credentials-file", missing, "--parallel", "invalid")
		if code != 3 || stdout != "" || stderr != "error: invalid --parallel \"invalid\", want integer\n" {
			t.Fatalf("exit=%d stdout=%q stderr=%q, want validation before credential read", code, stdout, stderr)
		}
	})
}

func TestCertifyCLIAllRequiresCredentialsFile(t *testing.T) {
	root := t.TempDir()

	_, stderr, code := certifyRun(t, root, "connectors", "certify", "--all")

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero (missing --credentials-file)")
	}
	if !strings.Contains(stderr, "credentials-file") {
		t.Errorf("stderr should mention --credentials-file: %s", stderr)
	}
}

// TestCertifyCLIBatchModeRunsCredsFileConnectors drives `pm connectors
// certify --all --credentials-file` end-to-end against the real CLI with
// "sample" as the only creds-file connector, and proves the batch JSON
// envelope + summary matrix are produced with exit 0.
func TestCertifyCLIBatchModeRunsCredsFileConnectors(t *testing.T) {
	t.Setenv("PM_CERT_SAMPLE_TOKEN", "sample-cli-token")
	root := t.TempDir()

	credsPath := filepath.Join(root, "creds.yaml")
	credsYAML := `
version: 1
connectors:
  sample:
    credential:
      from_env: {token: PM_CERT_SAMPLE_TOKEN}
`
	if err := os.WriteFile(credsPath, []byte(credsYAML), 0o600); err != nil {
		t.Fatalf("write creds file: %v", err)
	}

	stdout, stderr, code := certifyRun(t, root, "connectors", "certify", "--all",
		"--credentials-file", credsPath, "--json")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, `"kind": "ConnectorCertificationBatch"`) {
		t.Errorf("stdout missing ConnectorCertificationBatch envelope: %s", stdout)
	}
}

// TestCertifyCLIBatchModeTextRendersMatrix proves the non-JSON batch
// rendering includes the summary matrix header row (certification design
// §B columns).
func TestCertifyCLIBatchModeTextRendersMatrix(t *testing.T) {
	t.Setenv("PM_CERT_SAMPLE_TOKEN", "sample-cli-token")
	root := t.TempDir()

	credsPath := filepath.Join(root, "creds.yaml")
	credsYAML := `
version: 1
connectors:
  sample:
    credential:
      from_env: {token: PM_CERT_SAMPLE_TOKEN}
`
	if err := os.WriteFile(credsPath, []byte(credsYAML), 0o600); err != nil {
		t.Fatalf("write creds file: %v", err)
	}

	stdout, _, code := certifyRun(t, root, "connectors", "certify", "--all", "--credentials-file", credsPath)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "connector\tcheck\tcatalog") {
		t.Errorf("stdout missing matrix header: %s", stdout)
	}
	if !strings.Contains(stdout, "exit_code: 0") {
		t.Errorf("stdout missing exit_code summary line: %s", stdout)
	}
}

// TestCertifyCLISweepWithoutTargetsIsUsageError proves --sweep with nothing
// to sweep (no --credentials-file, no prior ledger) fails clearly instead of
// silently reporting success.
func TestCertifyCLISweepWithoutTargetsIsUsageError(t *testing.T) {
	root := t.TempDir()

	_, stderr, code := certifyRun(t, root, "connectors", "certify", "--sweep")

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero (nothing to sweep)")
	}
	if stderr == "" {
		t.Errorf("stderr empty, want an explanatory error")
	}
}

// TestCertifyCLISweepInvalidOlderThanIsUsageError proves a malformed
// --older-than duration is rejected before any sweeping is attempted.
func TestCertifyCLISweepInvalidOlderThanIsUsageError(t *testing.T) {
	root := t.TempDir()
	credsPath := filepath.Join(root, "creds.yaml")
	if err := os.WriteFile(credsPath, []byte("version: 1\nconnectors: {}\n"), 0o600); err != nil {
		t.Fatalf("write creds file: %v", err)
	}

	_, _, code := certifyRun(t, root, "connectors", "certify", "--sweep",
		"--credentials-file", credsPath, "--older-than", "not-a-duration")

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero for invalid --older-than")
	}
}

// TestCertifyCLIDoesNotBreakExistingConnectorsSubcommands is a regression
// guard: adding the certify dispatch case must not change behavior of any
// pre-existing `pm connectors` subcommand.
func TestCertifyCLIDoesNotBreakExistingConnectorsSubcommands(t *testing.T) {
	root := t.TempDir()

	stdout, stderr, code := certifyRun(t, root, "connectors", "list", "--json")
	if code != 0 {
		t.Fatalf("connectors list --json: exit %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, `"kind": "ConnectorList"`) {
		t.Errorf("connectors list output missing ConnectorList kind: %s", stdout)
	}

	stdout2, stderr2, code2 := certifyRun(t, root, "connectors", "inspect", "sample", "--json")
	if code2 != 0 {
		t.Fatalf("connectors inspect sample --json: exit %d, stderr=%s", code2, stderr2)
	}
	if !strings.Contains(stdout2, `"kind": "Connector"`) {
		t.Errorf("connectors inspect output missing Connector kind: %s", stdout2)
	}
}
