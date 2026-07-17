package cli_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

const syntheticLogCanary = "pm-test-cli-log-redaction-value-404"

func TestRedactedRunLogsSmoke(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PM_SYNTHETIC_LOG_CANARY", syntheticLogCanary)

	runCLIForLogSmoke(t, root, "init", "--json")
	runCLIForLogSmoke(t, root, "credentials", "add", "sample-cred", "--connector", "sample", "--from-env", "token=PM_SYNTHETIC_LOG_CANARY", "--json")
	runCLIForLogSmoke(t, root, "credentials", "add", "warehouse-local", "--connector", "warehouse", "--config", "path=.polymetrics/warehouse", "--json")
	runCLIForLogSmoke(t, root, "connections", "create", "sample-to-warehouse", "--source", "sample:sample-cred", "--destination", "warehouse:warehouse-local", "--stream", "customers", "--sync-mode", "full_refresh_overwrite", "--table", "customers", "--json")

	stdout, stderr := runCLIForLogSmoke(t, root, "etl", "run", "--connection", "sample-to-warehouse", "--stream", "customers", "--batch-size", "2", "--json")
	assertSingleJSONEnvelopeKind(t, stdout, "ETLRun")
	assertCleanOfSyntheticCanary(t, stdout)
	assertCleanOfSyntheticCanary(t, stderr)

	logs := readSmokeLogs(t, root)
	if len(logs) == 0 {
		t.Fatalf("run log is empty")
	}
	if !syntheticCanaryScanner([]byte("dirty " + syntheticLogCanary)) {
		t.Fatalf("synthetic canary scanner failed to detect dirty fixture")
	}
	if syntheticCanaryScanner(logs) {
		t.Fatalf("run log contained synthetic canary")
	}
}

func runCLIForLogSmoke(t *testing.T, root string, args ...string) (string, string) {
	t.Helper()
	allArgs := append([]string{"--root", root}, args...)
	var stdout, stderr bytes.Buffer
	code := cli.Run(allArgs, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("cli %v exit %d; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
	}
	return stdout.String(), stderr.String()
}

func assertSingleJSONEnvelopeKind(t *testing.T, stdout string, wantKind string) {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(stdout))
	var env map[string]any
	if err := dec.Decode(&env); err != nil {
		t.Fatalf("stdout is not one JSON envelope: %v", err)
	}
	if got, _ := env["api_version"].(string); got != "polymetrics.ai/v1" {
		t.Fatalf("api_version = %q", got)
	}
	if got, _ := env["kind"].(string); got != wantKind {
		t.Fatalf("kind = %q, want %q", got, wantKind)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		t.Fatalf("stdout has extra JSON/text after first envelope")
	}
}

func readSmokeLogs(t *testing.T, root string) []byte {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(root, ".polymetrics", "logs", "*.jsonl"))
	if err != nil {
		t.Fatalf("glob logs: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("expected at least one run log")
	}
	var combined bytes.Buffer
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read log: %v", err)
		}
		combined.Write(data)
	}
	return combined.Bytes()
}

func syntheticCanaryScanner(data []byte) bool {
	return bytes.Contains(data, []byte(syntheticLogCanary))
}

func assertCleanOfSyntheticCanary(t *testing.T, text string) {
	t.Helper()
	if strings.Contains(text, syntheticLogCanary) {
		t.Fatalf("output contained synthetic canary")
	}
}
