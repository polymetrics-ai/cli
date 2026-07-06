package certify_test

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/certify"
)

// fakeRunnable is a certify.Runnable test double letting batch tests control
// exactly what report (or error) each connector run produces, without
// driving the real CLI (this is the same pattern stages_source_test.go uses
// for the harness, applied one level up).
type fakeRunnable struct {
	rep   certify.Report
	err   error
	delay time.Duration
	onRun func()
}

func (f *fakeRunnable) Run(ctx context.Context) (certify.Report, error) {
	if f.onRun != nil {
		f.onRun()
	}
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return certify.Report{}, ctx.Err()
		}
	}
	return f.rep, f.err
}

func passingReport(connector string) certify.Report {
	now := time.Now().UTC()
	return certify.Report{
		Kind:          "ConnectorCertification",
		SchemaVersion: 1,
		Connector:     connector,
		Mode:          "live",
		Passed:        true,
		StartedAt:     now,
		CompletedAt:   now.Add(time.Second),
		Capabilities: certify.Capabilities{
			Check:   certify.CapabilityResult{Result: "pass"},
			Catalog: certify.CapabilityResult{Result: "pass", Streams: 1},
			Read:    certify.CapabilityResult{Result: "pass", Stream: "customers", Records: 3},
			SyncModes: map[string]certify.SyncModeResult{
				"full_refresh_append": {Result: "pass", DataSource: "live"},
			},
			Resume:          certify.CapabilityResult{Result: "pass"},
			JSONContract:    certify.CapabilityResult{Result: "pass", StagesChecked: 12},
			SecretRedaction: certify.CapabilityResult{Result: "pass"},
		},
	}
}

func failingReport(connector string) certify.Report {
	rep := passingReport(connector)
	rep.Passed = false
	rep.Capabilities.Read.Result = "fail"
	return rep
}

func leakedReport(connector string) certify.Report {
	rep := passingReport(connector)
	rep.Passed = false
	rep.Leaks = []certify.Leak{{Tag: "pm-cert-" + connector + "-ab12-1234", Connector: connector, Reason: "cleanup failed"}}
	return rep
}

// TestRunBatchRunsEveryConnectorAndAggregatesExitCode proves RunBatch drives
// one Runnable per creds-file connector and rolls per-connector exit codes
// up to a batch-level exit code (certification-design §B/§A: 0 pass / 2
// certification failures / 3 leaked resources dominates).
func TestRunBatchRunsEveryConnectorAndAggregatesExitCode(t *testing.T) {
	cf := certify.CredsFile{
		Defaults: certify.CredsDefaults{Parallel: 2},
		Connectors: map[string]certify.ConnectorCredsEntry{
			"github": {},
			"stripe": {},
		},
	}

	factory := func(name string, _ certify.Options) certify.Runnable {
		switch name {
		case "github":
			return &fakeRunnable{rep: passingReport("github")}
		case "stripe":
			return &fakeRunnable{rep: passingReport("stripe")}
		}
		t.Fatalf("unexpected connector %q", name)
		return nil
	}

	batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}

	if len(batch.Results) != 2 {
		t.Fatalf("len(Results) = %d, want 2", len(batch.Results))
	}
	if batch.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0 (all pass)", batch.ExitCode)
	}
}

// TestRunBatchExitCodeReflectsWorstConnector proves a single failing
// connector forces exit 2 even when others pass.
func TestRunBatchExitCodeReflectsWorstConnector(t *testing.T) {
	cf := certify.CredsFile{
		Connectors: map[string]certify.ConnectorCredsEntry{
			"github": {},
			"stripe": {},
		},
	}
	factory := func(name string, _ certify.Options) certify.Runnable {
		if name == "stripe" {
			return &fakeRunnable{rep: failingReport("stripe")}
		}
		return &fakeRunnable{rep: passingReport(name)}
	}

	batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}
	if batch.ExitCode != 2 {
		t.Errorf("ExitCode = %d, want 2 (one connector failed)", batch.ExitCode)
	}
}

// TestRunBatchLeakDominatesExitCode proves ANY leak forces exit 3
// regardless of other connectors' outcomes (certification-design §B:
// "any leak row printed first and forces exit 3").
func TestRunBatchLeakDominatesExitCode(t *testing.T) {
	cf := certify.CredsFile{
		Connectors: map[string]certify.ConnectorCredsEntry{
			"github": {},
			"stripe": {},
		},
	}
	factory := func(name string, _ certify.Options) certify.Runnable {
		if name == "stripe" {
			return &fakeRunnable{rep: leakedReport("stripe")}
		}
		return &fakeRunnable{rep: passingReport(name)}
	}

	batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}
	if batch.ExitCode != 3 {
		t.Errorf("ExitCode = %d, want 3 (a connector leaked)", batch.ExitCode)
	}
	if len(batch.Leaks()) != 1 {
		t.Fatalf("len(Leaks()) = %d, want 1", len(batch.Leaks()))
	}
}

// TestRunBatchSkipsConfiguredConnectors proves a creds-file entry with
// skip:true never invokes the RunnerFactory and records a skip reason
// (certification-design §B example: "salesforce: skip: true, reason: ...").
func TestRunBatchSkipsConfiguredConnectors(t *testing.T) {
	cf := certify.CredsFile{
		Connectors: map[string]certify.ConnectorCredsEntry{
			"github":     {},
			"salesforce": {Skip: true, Reason: "no sandbox tenant yet"},
		},
	}
	invoked := map[string]bool{}
	var mu sync.Mutex
	factory := func(name string, _ certify.Options) certify.Runnable {
		mu.Lock()
		invoked[name] = true
		mu.Unlock()
		return &fakeRunnable{rep: passingReport(name)}
	}

	batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}
	if invoked["salesforce"] {
		t.Errorf("RunnerFactory invoked for skipped connector salesforce")
	}
	result := findResult(t, batch, "salesforce")
	if !result.Skipped {
		t.Errorf("salesforce result.Skipped = false, want true")
	}
	if result.SkipReason != "no sandbox tenant yet" {
		t.Errorf("salesforce result.SkipReason = %q, want %q", result.SkipReason, "no sandbox tenant yet")
	}
	if batch.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0 (skip is not a failure)", batch.ExitCode)
	}
}

// TestRunBatchRunsConnectorsConcurrentlyUpToParallelLimit proves the worker
// pool actually overlaps runs (bounded by Defaults.Parallel), not a
// disguised serial loop, using a barrier: with parallel=3 and delay=50ms for
// 3 connectors, wall time should be well under 3x50ms.
func TestRunBatchRunsConnectorsConcurrentlyUpToParallelLimit(t *testing.T) {
	cf := certify.CredsFile{
		Defaults: certify.CredsDefaults{Parallel: 3},
		Connectors: map[string]certify.ConnectorCredsEntry{
			"a": {}, "b": {}, "c": {},
		},
	}
	var mu sync.Mutex
	inFlight := 0
	maxInFlight := 0
	factory := func(name string, _ certify.Options) certify.Runnable {
		return &fakeRunnable{
			rep:   passingReport(name),
			delay: 80 * time.Millisecond,
			onRun: func() {
				mu.Lock()
				inFlight++
				if inFlight > maxInFlight {
					maxInFlight = inFlight
				}
				mu.Unlock()
			},
		}
	}

	start := time.Now()
	_, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      t.TempDir(),
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}
	if elapsed > 200*time.Millisecond {
		t.Errorf("elapsed = %v, want well under 3x80ms serial time (parallelism not happening)", elapsed)
	}
	mu.Lock()
	defer mu.Unlock()
	if maxInFlight < 2 {
		t.Errorf("maxInFlight = %d, want >= 2 (connectors should overlap)", maxInFlight)
	}
}

// TestRunBatchDefaultParallelIsAtLeastOne proves a creds file with no
// explicit defaults.parallel still runs (doesn't deadlock on a zero-sized
// worker pool).
func TestRunBatchDefaultParallelIsAtLeastOne(t *testing.T) {
	cf := certify.CredsFile{
		Connectors: map[string]certify.ConnectorCredsEntry{
			"github": {},
		},
	}
	factory := func(name string, _ certify.Options) certify.Runnable {
		return &fakeRunnable{rep: passingReport(name)}
	}

	done := make(chan struct{})
	go func() {
		_, _ = certify.RunBatch(context.Background(), certify.BatchOptions{
			CredsFile:     cf,
			RunnerFactory: factory,
			BatchDir:      t.TempDir(),
		})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("RunBatch() did not return within 5s (likely deadlocked on zero-sized pool)")
	}
}

// TestRunBatchWritesProgressFile proves batch mode persists a resumability
// marker (certification-design §B: "certifications/batch-<runid>/
// progress.json").
func TestRunBatchWritesProgressFile(t *testing.T) {
	cf := certify.CredsFile{
		Connectors: map[string]certify.ConnectorCredsEntry{"github": {}},
	}
	factory := func(name string, _ certify.Options) certify.Runnable {
		return &fakeRunnable{rep: passingReport(name)}
	}
	batchDir := t.TempDir()

	batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      batchDir,
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}
	if batch.RunID == "" {
		t.Fatalf("BatchReport.RunID is empty")
	}

	progressPath := filepath.Join(batchDir, "certifications", "batch-"+batch.RunID, "progress.json")
	if _, err := os.Stat(progressPath); err != nil {
		t.Fatalf("progress.json not written at %s: %v", progressPath, err)
	}
}

// TestRunBatchResumeSkipsConnectorsWithFreshReport proves --resume skips a
// connector whose <dir>/certifications/<name>.json report is newer than the
// batch start time (certification-design §B: "--resume skips connectors
// whose report is newer than batch start").
func TestRunBatchResumeSkipsConnectorsWithFreshReport(t *testing.T) {
	cf := certify.CredsFile{
		Connectors: map[string]certify.ConnectorCredsEntry{
			"github": {},
			"stripe": {},
		},
	}
	batchDir := t.TempDir()

	// Pre-seed a fresh report for github (as if a prior batch run already
	// certified it), timestamped after "batch start" from RunBatch's
	// perspective (we set StartedAt/CompletedAt to now, then wait long
	// enough that a fresh RunBatch call's own start time is strictly later
	// -- resume compares report CompletedAt against batch start).
	existing := passingReport("github")
	if err := existing.Save(batchDir); err != nil {
		t.Fatalf("seed report Save() error = %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	invoked := map[string]bool{}
	var mu sync.Mutex
	factory := func(name string, _ certify.Options) certify.Runnable {
		mu.Lock()
		invoked[name] = true
		mu.Unlock()
		return &fakeRunnable{rep: passingReport(name)}
	}

	batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      batchDir,
		Resume:        true,
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if invoked["github"] {
		t.Errorf("RunnerFactory invoked for github, want skipped via --resume (fresh existing report)")
	}
	if !invoked["stripe"] {
		t.Errorf("RunnerFactory not invoked for stripe, want a fresh run (no existing report)")
	}
	result := findResult(t, batch, "github")
	if !result.Resumed {
		t.Errorf("github result.Resumed = false, want true")
	}
}

// TestRunBatchWithoutResumeAlwaysReruns proves the default (Resume=false)
// behavior ignores any pre-existing report and re-runs every connector.
func TestRunBatchWithoutResumeAlwaysReruns(t *testing.T) {
	cf := certify.CredsFile{
		Connectors: map[string]certify.ConnectorCredsEntry{"github": {}},
	}
	batchDir := t.TempDir()
	existing := passingReport("github")
	if err := existing.Save(batchDir); err != nil {
		t.Fatalf("seed report Save() error = %v", err)
	}

	invoked := false
	var mu sync.Mutex
	factory := func(name string, _ certify.Options) certify.Runnable {
		mu.Lock()
		invoked = true
		mu.Unlock()
		return &fakeRunnable{rep: passingReport(name)}
	}

	_, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      batchDir,
		Resume:        false,
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if !invoked {
		t.Errorf("RunnerFactory not invoked, want a fresh run since Resume=false")
	}
}

// TestBatchReportSummaryMatrixLeaksRowFirst proves the rendered summary
// matrix places any leaked connector's row first (certification-design §B:
// "any leak row printed first").
func TestBatchReportSummaryMatrixLeaksRowFirst(t *testing.T) {
	cf := certify.CredsFile{
		Connectors: map[string]certify.ConnectorCredsEntry{
			"alpha": {}, "beta": {}, "gamma": {},
		},
	}
	factory := func(name string, _ certify.Options) certify.Runnable {
		if name == "gamma" {
			return &fakeRunnable{rep: leakedReport("gamma")}
		}
		return &fakeRunnable{rep: passingReport(name)}
	}

	batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}

	matrix := batch.SummaryMatrix()
	if len(matrix.Rows) != 3 {
		t.Fatalf("len(Rows) = %d, want 3", len(matrix.Rows))
	}
	if matrix.Rows[0].Connector != "gamma" {
		t.Errorf("Rows[0].Connector = %q, want gamma (leaked connector first)", matrix.Rows[0].Connector)
	}
	// Remaining rows still deterministic (sorted) so output is stable.
	rest := []string{matrix.Rows[1].Connector, matrix.Rows[2].Connector}
	sort.Strings(rest)
	if rest[0] != "alpha" || rest[1] != "beta" {
		t.Errorf("remaining rows = %v, want [alpha beta]", rest)
	}
}

// TestBatchReportSummaryMatrixColumns proves the matrix carries the
// design's column set (certification-design §B: "columns = check/catalog/
// read/5 modes/resume/write/flow/schedule/redaction/leaks").
func TestBatchReportSummaryMatrixColumns(t *testing.T) {
	cf := certify.CredsFile{
		Connectors: map[string]certify.ConnectorCredsEntry{"github": {}},
	}
	factory := func(name string, _ certify.Options) certify.Runnable {
		return &fakeRunnable{rep: passingReport(name)}
	}

	batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}
	row := batch.SummaryMatrix().Rows[0]
	if row.Check != "pass" {
		t.Errorf("row.Check = %q, want pass", row.Check)
	}
	if row.Catalog != "pass" {
		t.Errorf("row.Catalog = %q, want pass", row.Catalog)
	}
	if row.Read != "pass" {
		t.Errorf("row.Read = %q, want pass", row.Read)
	}
	if row.Resume != "pass" {
		t.Errorf("row.Resume = %q, want pass", row.Resume)
	}
	if row.Redaction != "pass" {
		t.Errorf("row.Redaction = %q, want pass", row.Redaction)
	}
	if row.Leaked {
		t.Errorf("row.Leaked = true, want false")
	}
}

// TestRunBatchConnectorRunnableErrorRecordedAsFailure proves a Runnable that
// returns a Go error (e.g. an unrecoverable setup failure, per Runner.Run's
// contract) is captured as a failed batch result rather than aborting the
// whole batch.
func TestRunBatchConnectorRunnableErrorRecordedAsFailure(t *testing.T) {
	cf := certify.CredsFile{
		Connectors: map[string]certify.ConnectorCredsEntry{
			"github": {},
			"stripe": {},
		},
	}
	factory := func(name string, _ certify.Options) certify.Runnable {
		if name == "github" {
			return &fakeRunnable{err: context.DeadlineExceeded}
		}
		return &fakeRunnable{rep: passingReport(name)}
	}

	batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v, want nil (per-connector errors must not abort the batch)", err)
	}
	if batch.ExitCode != 2 {
		t.Errorf("ExitCode = %d, want 2 (github's runner errored)", batch.ExitCode)
	}
	result := findResult(t, batch, "github")
	if result.Error == "" {
		t.Errorf("github result.Error is empty, want the runner error message recorded")
	}
}

func findResult(t *testing.T, batch certify.BatchReport, connector string) certify.BatchConnectorResult {
	t.Helper()
	for _, r := range batch.Results {
		if r.Connector == connector {
			return r
		}
	}
	t.Fatalf("no result for connector %q in %+v", connector, batch.Results)
	return certify.BatchConnectorResult{}
}
