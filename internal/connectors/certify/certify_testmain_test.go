package certify_test

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors/certify"
)

// certifyRealCLIInvocationBudget is a deterministic test-only ceiling. The
// retained real-surface proof is TestFullSweepSourceStagesAgainstSample; every
// other runner scenario must use the scripted harness driver. The count also
// includes focused Harness and Sweeper coverage because they deliberately
// exercise the same in-process cli.Run seam.
//
// This deliberately generous initial ceiling is a RED checkpoint: the
// pre-refactor suite's repeated Runner.Run calls exceed it. The GREEN slice
// narrows it to the measured retained-proof count.
const certifyRealCLIInvocationBudget = 128

var certifyRealCLIInvocations atomic.Int64

func countedCertifyCLIRun(args []string, stdout, stderr io.Writer) int {
	certifyRealCLIInvocations.Add(1)
	return cli.Run(args, stdout, stderr)
}

// TestMain wires the real internal/cli.Run entrypoint into this package's
// in-process CLI driver exactly once for the whole test binary
// (certify.SetCLIRunFunc), mirroring what cmd/pm/main.go does in
// production. certify cannot import internal/cli directly (internal/cli's
// own `pm connectors certify` dispatch imports certify, and Go forbids the
// resulting cycle), so every stage/harness/runner test in this package
// depends on this registration having already happened.
func TestMain(m *testing.M) {
	certify.SetCLIRunFunc(countedCertifyCLIRun)
	code := m.Run()
	got := certifyRealCLIInvocations.Load()
	fmt.Fprintf(os.Stderr, "certify real CLI invocations: %d (budget %d)\n", got, certifyRealCLIInvocationBudget)
	if got > certifyRealCLIInvocationBudget {
		fmt.Fprintf(os.Stderr, "certify real CLI invocation budget exceeded: got %d, allowed %d; retain only the exhaustive real proof and script duplicate stage cases\n", got, certifyRealCLIInvocationBudget)
		code = 1
	}
	os.Exit(code)
}
