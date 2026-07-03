package cli

import (
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

// TestExitForReportPassIsNil proves a passing report yields a nil error
// (cli.Run's default 0-exit path), matching every other successful command.
func TestExitForReportPassIsNil(t *testing.T) {
	rep := certify.Report{Connector: "sample", Passed: true}
	if err := exitForReport(rep); err != nil {
		t.Errorf("exitForReport(passed) = %v, want nil", err)
	}
}

// TestExitForReportFailureIsExit2 proves a failing (non-leaked) report maps
// to exit code 2, per certification design §A.
func TestExitForReportFailureIsExit2(t *testing.T) {
	rep := certify.Report{Connector: "sample", Passed: false}
	err := exitForReport(rep)
	if err == nil {
		t.Fatalf("exitForReport(failed) = nil, want an error")
	}
	if code := exitCodeFor(classifyError(err)); code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
}

// TestExitForReportLeakIsExit3 proves ANY leak forces exit 3 even when
// Passed happens to be false too (certification design §A: "3 leaked
// resources (dominates everything)").
func TestExitForReportLeakIsExit3(t *testing.T) {
	rep := certify.Report{
		Connector: "sample",
		Passed:    false,
		Leaks:     []certify.Leak{{Tag: "pm-cert-sample-ab12-1234", Connector: "sample", Reason: "cleanup failed"}},
	}
	err := exitForReport(rep)
	if err == nil {
		t.Fatalf("exitForReport(leaked) = nil, want an error")
	}
	if code := exitCodeFor(classifyError(err)); code != 3 {
		t.Errorf("exit code = %d, want 3", code)
	}
}

// TestExitForReportDoesNotDoubleReportError proves the certify exit-code
// error is marked alreadyReported so writeError doesn't emit a second,
// conflicting JSON envelope on top of the report the caller already wrote
// (cli.Run's one-envelope-per-invocation contract).
func TestExitForReportDoesNotDoubleReportError(t *testing.T) {
	rep := certify.Report{Connector: "sample", Passed: false}
	err := exitForReport(rep)
	ce := classifyError(err)
	if ce == nil {
		t.Fatalf("classifyError(err) = nil")
	}
	if !ce.alreadyReported {
		t.Errorf("alreadyReported = false, want true (certify writes its own report)")
	}
}

// TestExitForBatchZeroIsNil proves a fully-passing batch yields a nil error.
func TestExitForBatchZeroIsNil(t *testing.T) {
	batch := certify.BatchReport{ExitCode: 0}
	if err := exitForBatch(batch); err != nil {
		t.Errorf("exitForBatch(0) = %v, want nil", err)
	}
}

// TestExitForBatchPropagatesWorstCode proves exitForBatch's returned error
// maps back to exactly batch.ExitCode.
func TestExitForBatchPropagatesWorstCode(t *testing.T) {
	for _, code := range []int{1, 2, 3} {
		batch := certify.BatchReport{ExitCode: code}
		err := exitForBatch(batch)
		if err == nil {
			t.Fatalf("exitForBatch(%d) = nil, want an error", code)
		}
		if got := exitCodeFor(classifyError(err)); got != code {
			t.Errorf("exitForBatch(%d) exit code = %d, want %d", code, got, code)
		}
	}
}

// TestExitForSweepNoFailuresIsNil proves a sweep with nothing failed exits 0.
func TestExitForSweepNoFailuresIsNil(t *testing.T) {
	results := map[string]certify.SweepResult{
		"github": {Scanned: 2, Cleaned: []string{"pm-cert-github-a"}},
	}
	if err := exitForSweep(results); err != nil {
		t.Errorf("exitForSweep(no failures) = %v, want nil", err)
	}
}

// TestExitForSweepFailureIsExit3 proves any sweep cleanup failure maps to
// exit 3 (a leak that could not be resolved even by the orphan sweeper).
func TestExitForSweepFailureIsExit3(t *testing.T) {
	results := map[string]certify.SweepResult{
		"github": {Scanned: 1, Failed: map[string]string{"pm-cert-github-a": "cleanup api error"}},
	}
	err := exitForSweep(results)
	if err == nil {
		t.Fatalf("exitForSweep(failure) = nil, want an error")
	}
	if code := exitCodeFor(classifyError(err)); code != 3 {
		t.Errorf("exit code = %d, want 3", code)
	}
}
