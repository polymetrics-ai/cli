package certify_test

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors/certify"
)

// certifyRealCLIInvocationBudget is a deterministic test-only ceiling. The
// retained real-surface proof in this package is
// TestSampleOutboxWriteLifecycleAgainstRealCLI (sample/outbox write
// lifecycle); the full source/read/flow/schedule proof is coalesced with the
// real CLI route in internal/cli. Every other runner scenario must use the
// scripted harness driver. The count also includes focused Harness and Sweeper
// coverage because they deliberately exercise the same in-process cli.Run seam.
//
// The pre-refactor suite made 782 calls. A cold -count=1 GREEN run makes 25,
// so the ceiling intentionally has no slack: new real CLI work must be an
// explicit change to this test contract rather than accidental duplication.
const certifyRealCLIInvocationBudget = 25

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
	_, _ = fmt.Fprintf(os.Stderr, "certify real CLI invocations: %d (budget %d)\n", got, certifyRealCLIInvocationBudget)
	if err := certifyInvocationBudgetError(got); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	os.Exit(code)
}

func certifyInvocationBudgetError(got int64) error {
	if got <= certifyRealCLIInvocationBudget {
		return nil
	}
	return fmt.Errorf("certify real CLI invocation budget exceeded: got %d, allowed %d; retain the write-lifecycle proof, keep the full sweep coalesced in the CLI route, and script duplicate stage cases", got, certifyRealCLIInvocationBudget)
}

func TestCertifyRealCLIInvocationBudgetRejectsControlledDuplicate(t *testing.T) {
	err := certifyInvocationBudgetError(certifyRealCLIInvocationBudget + 1)
	if err == nil {
		t.Fatal("certifyInvocationBudgetError() error = nil, want duplicate invocation failure")
	}
	if got, want := err.Error(), fmt.Sprintf("got %d, allowed %d", certifyRealCLIInvocationBudget+1, certifyRealCLIInvocationBudget); !strings.Contains(got, want) {
		t.Errorf("certifyInvocationBudgetError() = %q, want %q", got, want)
	}
}
