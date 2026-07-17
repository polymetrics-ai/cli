package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
	pmlogging "polymetrics.ai/internal/logging"
)

func TestRedactedErrorOutputSingleLineSmoke(t *testing.T) {
	pmlogging.RegisterValue(syntheticLogCanary)
	root := t.TempDir()
	stdout, stderr, code := runCLIRawForLogSmoke(root, "help", syntheticLogCanary+"\nforged-line", "--json")
	if code == 0 {
		t.Fatalf("help with unknown topic exit code = 0, want failure")
	}
	assertSingleJSONEnvelopeKind(t, stdout, "Error")
	if !syntheticCanaryScanner([]byte("dirty " + syntheticLogCanary)) {
		t.Fatalf("synthetic canary scanner failed to detect dirty fixture")
	}
	assertCleanOfSyntheticCanary(t, stdout)
	assertCleanOfSyntheticCanary(t, stderr)
	if lines := strings.Split(strings.TrimSuffix(stderr, "\n"), "\n"); len(lines) != 1 {
		t.Fatalf("stderr diagnostic was not single-line: %q", stderr)
	}
	if !strings.Contains(stdout, `\n`) {
		t.Fatalf("JSON error did not preserve escaped newline")
	}
}

func runCLIRawForLogSmoke(root string, args ...string) (string, string, int) {
	allArgs := append([]string{"--root", root}, args...)
	var stdout, stderr bytes.Buffer
	code := cli.Run(allArgs, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}
