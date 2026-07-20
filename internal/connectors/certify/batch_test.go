package certify_test

import (
	"context"
	"errors"
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
	stageNames := []string{
		"init", "preflight", "manual_json", "credentials_add", "warehouse_credentials_add",
		"credentials_test", "connection_create", "catalog", "etl_full_refresh_append",
		"etl_full_refresh_overwrite", "etl_full_refresh_overwrite_deduped", "etl_incremental_append",
		"resume", "etl_incremental_append_deduped", "query_contract",
	}
	stages := make([]certify.StageResult, 0, len(stageNames))
	for _, name := range stageNames {
		stages = append(stages, certify.StageResult{Name: name, Tier: 2, Passed: true})
	}
	return certify.Report{
		Kind:          "ConnectorCertification",
		SchemaVersion: 1,
		Connector:     connector,
		Mode:          "live",
		Passed:        true,
		StartedAt:     now.Add(-time.Second),
		CompletedAt:   now,
		Capabilities: certify.Capabilities{
			Check:   certify.CapabilityResult{Result: "pass"},
			Catalog: certify.CapabilityResult{Result: "pass", Streams: 1},
			Read:    certify.CapabilityResult{Result: "pass", Stream: "customers", Records: 3},
			SyncModes: map[string]certify.SyncModeResult{
				"full_refresh_append":            {Result: "pass", DataSource: "live"},
				"full_refresh_overwrite":         {Result: "pass", DataSource: "capture"},
				"full_refresh_overwrite_deduped": {Result: "pass", DataSource: "capture"},
				"incremental_append":             {Result: "pass", DataSource: "live", CursorAdvanced: true},
				"incremental_append_deduped":     {Result: "pass", DataSource: "capture"},
			},
			Resume:          certify.CapabilityResult{Result: "pass"},
			JSONContract:    certify.CapabilityResult{Result: "pass", StagesChecked: len(stages)},
			SecretRedaction: certify.CapabilityResult{Result: "pass"},
		},
		Stages: stages,
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

func TestRunBatchEffectRecorderRejectsUnsupportedCredentialsConstraintsBeforeRunner(t *testing.T) {
	tests := []struct {
		name  string
		entry certify.ConnectorCredsEntry
		defs  certify.CredsDefaults
	}{
		{name: "write without sandbox", entry: certify.ConnectorCredsEntry{Write: true}},
		{name: "credential exec", entry: certify.ConnectorCredsEntry{Credential: certify.CredentialRef{Exec: map[string][]string{"token": {"must-not-run"}}}}},
		{name: "rate limit", entry: certify.ConnectorCredsEntry{RateLimitRPS: 2}},
		{name: "budget", entry: certify.ConnectorCredsEntry{BudgetCalls: 50}},
		{name: "limit", entry: certify.ConnectorCredsEntry{Limit: 10}},
		{name: "default rate limit", defs: certify.CredsDefaults{RateLimitRPS: 2}},
		{name: "default budget", defs: certify.CredsDefaults{BudgetCalls: 50}},
		{name: "default limit", defs: certify.CredsDefaults{Limit: 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var effects []string
			_, err := certify.RunBatch(context.Background(), certify.BatchOptions{
				CredsFile: certify.CredsFile{
					Defaults: tt.defs,
					Connectors: map[string]certify.ConnectorCredsEntry{
						"sample": tt.entry,
					},
				},
				RunnerFactory: func(name string, _ certify.Options) certify.Runnable {
					effects = append(effects, "runner:"+name)
					return &fakeRunnable{rep: passingReport(name)}
				},
				BatchDir: t.TempDir(),
			})
			if err == nil {
				t.Fatal("RunBatch() error = nil, want fail-closed unsupported constraint")
			}
			if len(effects) != 0 {
				t.Fatalf("runner effects=%v, want none", effects)
			}
		})
	}
}

func TestRunBatchAllowsSandboxGatedWrites(t *testing.T) {
	var got certify.Options
	_, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile: certify.CredsFile{Connectors: map[string]certify.ConnectorCredsEntry{
			"sample": {Write: true, Sandbox: true},
		}},
		RunnerFactory: func(name string, opts certify.Options) certify.Runnable {
			got = opts
			return &fakeRunnable{rep: passingReport(name)}
		},
		BatchDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}
	if !got.Write {
		t.Fatal("sandbox-gated write setting was not propagated to runner options")
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

func TestRunBatchRunnerErrorWithLeakedReportKeepsExit3(t *testing.T) {
	cf := certify.CredsFile{
		Defaults: certify.CredsDefaults{Parallel: 1},
		Connectors: map[string]certify.ConnectorCredsEntry{
			"stripe": {},
		},
	}
	batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile: cf,
		RunnerFactory: func(name string, _ certify.Options) certify.Runnable {
			return &fakeRunnable{rep: leakedReport(name), err: errors.New("runner returned report plus ancillary error")}
		},
		BatchDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v, want per-connector error recorded in batch result", err)
	}
	if batch.ExitCode != 3 {
		t.Fatalf("Batch ExitCode = %d, want 3 when returned report contains leaks", batch.ExitCode)
	}
	if len(batch.Results) != 1 || batch.Results[0].ExitCode != 3 {
		t.Fatalf("Batch results = %+v, want connector exit 3", batch.Results)
	}
	if len(batch.Leaks()) != 1 {
		t.Fatalf("Batch leaks = %+v, want returned leaked report preserved", batch.Leaks())
	}
	if batch.Results[0].Error == "" {
		t.Fatal("runner ancillary error was not recorded")
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
// pool actually overlaps runs up to Defaults.Parallel. It uses a release
// barrier and active-worker accounting instead of wall-clock timing, so a
// loaded runner cannot turn correct parallelism into a false failure.
func TestRunBatchRunsConnectorsConcurrentlyUpToParallelLimit(t *testing.T) {
	cf := certify.CredsFile{
		Defaults: certify.CredsDefaults{Parallel: 3},
		Connectors: map[string]certify.ConnectorCredsEntry{
			"a": {}, "b": {}, "c": {}, "d": {},
		},
	}
	wantParallel := cf.Defaults.Parallel

	started := make(chan string, len(cf.Connectors))
	violations := make(chan string, 1)
	release := make(chan struct{})
	var releaseOnce sync.Once
	closeRelease := func() { releaseOnce.Do(func() { close(release) }) }

	type runBatchOutcome struct {
		batch certify.BatchReport
		err   error
	}
	outcomes := make(chan runBatchOutcome, 1)
	done := make(chan struct{})
	t.Cleanup(func() {
		closeRelease()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Errorf("timed out waiting for RunBatch() cleanup after releasing worker barrier")
		}
	})

	var mu sync.Mutex
	active := 0
	maxActive := 0
	factory := func(name string, _ certify.Options) certify.Runnable {
		return &fakeRunnable{
			rep: passingReport(name),
			onRun: func() {
				mu.Lock()
				active++
				if active > maxActive {
					maxActive = active
				}
				if active > wantParallel {
					select {
					case violations <- name:
					default:
					}
				}
				mu.Unlock()

				started <- name
				<-release

				mu.Lock()
				active--
				mu.Unlock()
			},
		}
	}

	go func() {
		batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
			CredsFile:     cf,
			RunnerFactory: factory,
			BatchDir:      t.TempDir(),
		})
		outcomes <- runBatchOutcome{batch: batch, err: err}
		close(done)
	}()

	seen := map[string]bool{}
	for range wantParallel {
		select {
		case name := <-started:
			seen[name] = true
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for %d parallel workers to start; started=%v", wantParallel, seen)
		}
	}

	mu.Lock()
	if maxActive != wantParallel {
		t.Errorf("maxActive = %d, want %d workers active before release", maxActive, wantParallel)
	}
	if active != wantParallel {
		t.Errorf("active = %d, want %d workers blocked on release barrier", active, wantParallel)
	}
	mu.Unlock()

	closeRelease()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for RunBatch() after releasing worker barrier")
	}
	outcome := <-outcomes
	if outcome.err != nil {
		t.Fatalf("RunBatch() error = %v", outcome.err)
	}
	for len(started) > 0 {
		seen[<-started] = true
	}
	if len(seen) != len(cf.Connectors) {
		t.Fatalf("started connectors = %v, want all %d connectors", seen, len(cf.Connectors))
	}
	if len(outcome.batch.Results) != len(cf.Connectors) {
		t.Fatalf("len(Results) = %d, want %d", len(outcome.batch.Results), len(cf.Connectors))
	}

	select {
	case name := <-violations:
		t.Fatalf("connector %s started while %d workers were already active", name, wantParallel)
	default:
	}
	mu.Lock()
	defer mu.Unlock()
	if active != 0 {
		t.Errorf("active = %d after RunBatch returned, want 0", active)
	}
	if maxActive != wantParallel {
		t.Errorf("maxActive = %d after RunBatch returned, want %d", maxActive, wantParallel)
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

// TestRunBatchResumeReusesCompletedReportAcrossOrdinaryRuns proves a normal
// second --resume invocation reuses the completed report persisted by the
// first invocation. No artificial timestamp newer than the second batch start
// is required, and the resumed result retains the prior report and exit code.
func TestRunBatchResumeReusesCompletedReportAcrossOrdinaryRuns(t *testing.T) {
	cf := certify.CredsFile{
		Connectors: map[string]certify.ConnectorCredsEntry{"github": {}},
	}
	batchDir := t.TempDir()
	var effects []string
	factory := func(name string, _ certify.Options) certify.Runnable {
		effects = append(effects, "run:"+name)
		return &fakeRunnable{rep: passingReport(name)}
	}

	first, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      batchDir,
	})
	if err != nil {
		t.Fatalf("first RunBatch() error = %v", err)
	}
	if findResult(t, first, "github").Resumed {
		t.Fatal("first run unexpectedly resumed")
	}

	second, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      batchDir,
		Resume:        true,
	})
	if err != nil {
		t.Fatalf("second RunBatch() error = %v", err)
	}
	if len(effects) != 1 {
		t.Fatalf("runner effects=%v, want one effect across two runs", effects)
	}
	result := findResult(t, second, "github")
	if !result.Resumed {
		t.Fatal("second run result.Resumed = false, want true")
	}
	if result.Report.Connector != "github" || !result.Report.Passed || result.ExitCode != 0 {
		t.Fatalf("resumed result did not reuse completed report: %+v", result)
	}
}

func TestRunBatchResumeAcceptsCleanupFailureAbsenceProof(t *testing.T) {
	cf := certify.CredsFile{
		Connectors: map[string]certify.ConnectorCredsEntry{"github": {}},
	}
	batchDir := t.TempDir()
	rep := cleanupFailureAbsenceProofReport("github")
	var effects []string
	factory := func(name string, _ certify.Options) certify.Runnable {
		effects = append(effects, "run:"+name)
		if len(effects) == 1 {
			return &fakeRunnable{rep: rep}
		}
		return &fakeRunnable{rep: passingReport(name)}
	}
	first, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      batchDir,
	})
	if err != nil {
		t.Fatalf("first RunBatch() error = %v", err)
	}
	if findResult(t, first, "github").Resumed {
		t.Fatal("first run unexpectedly resumed")
	}

	batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      batchDir,
		Resume:        true,
	})
	if err != nil {
		t.Fatalf("second RunBatch() error = %v", err)
	}
	if len(effects) != 1 {
		t.Fatalf("absence-proven cleanup failure reran effects=%v, want only first run", effects)
	}
	result := findResult(t, batch, "github")
	if !result.Resumed {
		t.Fatalf("result.Resumed=false, want true for absence-proven cleanup failure: %+v", result)
	}
	if result.ExitCode != 2 || result.Report.Passed || len(result.Report.Leaks) != 0 {
		t.Fatalf("resumed result = %+v, want failed non-leaked report with exit 2", result)
	}
}

func cleanupFailureAbsenceProofReport(connector string) certify.Report {
	rep := passingReport(connector)
	tag := "pm-cert-" + connector + "-12345678-1700000000"
	rep.Passed = false
	rep.Capabilities.WriteActions = map[string]certify.WriteActionResult{
		"create_issue": {
			Result:  "fail",
			Cleanup: "close_issue",
			Verify:  "read_back",
			Tag:     tag,
			Reason:  "write_cleanup failed, but cleanup_verify proved the entity absent",
		},
	}
	rep.Stages = append(rep.Stages,
		certify.StageResult{Name: "write_plan_preview", Tier: 2, Passed: true},
		certify.StageResult{Name: "write_create", Tier: 2, Passed: true},
		certify.StageResult{Name: "write_verify", Tier: 2, Passed: true},
		certify.StageResult{Name: "approval_idempotency", Tier: 2, Passed: true},
		certify.StageResult{Name: "write_cleanup", Tier: 2, Passed: false, Error: "write_cleanup: reverse run exit=1 stderr=fixture cleanup error"},
		certify.StageResult{Name: "cleanup_verify", Tier: 2, Passed: true},
	)
	return rep
}

func futureDatedReport(rep certify.Report) certify.Report {
	future := time.Now().UTC().Add(24 * time.Hour)
	rep.StartedAt = future
	rep.CompletedAt = future.Add(time.Minute)
	return rep
}

func TestRunBatchResumeRerunsFutureDatedReports(t *testing.T) {
	for _, tc := range []struct {
		name   string
		report func(string) certify.Report
	}{
		{name: "ordinary completed report", report: passingReport},
		{name: "cleanup failure absence proof", report: cleanupFailureAbsenceProofReport},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cf := certify.CredsFile{Connectors: map[string]certify.ConnectorCredsEntry{"github": {}}}
			batchDir := t.TempDir()
			var effects []string
			factory := func(name string, _ certify.Options) certify.Runnable {
				effects = append(effects, "run:"+name)
				if len(effects) == 1 {
					return &fakeRunnable{rep: futureDatedReport(tc.report(name))}
				}
				return &fakeRunnable{rep: passingReport(name)}
			}

			if _, err := certify.RunBatch(context.Background(), certify.BatchOptions{
				CredsFile:     cf,
				RunnerFactory: factory,
				BatchDir:      batchDir,
			}); err != nil {
				t.Fatalf("seed RunBatch() error = %v", err)
			}

			batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
				CredsFile:     cf,
				RunnerFactory: factory,
				BatchDir:      batchDir,
				Resume:        true,
			})
			if err != nil {
				t.Fatalf("resume RunBatch() error = %v", err)
			}
			if len(effects) != 2 {
				t.Fatalf("future-dated report effects=%v, want rerun instead of resume", effects)
			}
			if findResult(t, batch, "github").Resumed {
				t.Fatalf("future-dated report was marked resumed: %+v", batch.Results)
			}
		})
	}
}

func TestRunBatchResumeRerunsIncompleteReport(t *testing.T) {
	cf := certify.CredsFile{
		Connectors: map[string]certify.ConnectorCredsEntry{"github": {}},
	}
	batchDir := t.TempDir()
	incomplete := passingReport("github")
	incomplete.CompletedAt = time.Time{}
	if err := incomplete.Save(batchDir); err != nil {
		t.Fatalf("seed incomplete report: %v", err)
	}

	invoked := false
	batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
		CredsFile: cf,
		RunnerFactory: func(name string, _ certify.Options) certify.Runnable {
			invoked = true
			return &fakeRunnable{rep: passingReport(name)}
		},
		BatchDir: batchDir,
		Resume:   true,
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}
	if !invoked {
		t.Fatal("RunnerFactory not invoked for incomplete prior report")
	}
	if findResult(t, batch, "github").Resumed {
		t.Fatal("incomplete prior report was marked resumed")
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

func TestRunBatchResumeRerunsMinimalAndEditedReports(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*certify.Report)
	}{
		{
			name: "minimal",
			mutate: func(rep *certify.Report) {
				rep.Stages = nil
				rep.Capabilities = certify.Capabilities{}
			},
		},
		{
			name: "edited passed outcome",
			mutate: func(rep *certify.Report) {
				rep.Passed = true
				rep.Stages[8].Passed = false
				rep.Stages[8].Error = "fixture stage failure"
			},
		},
		{
			name: "edited leaks",
			mutate: func(rep *certify.Report) {
				rep.Passed = true
				rep.Leaks = nil
				rep.Capabilities.WriteActions = map[string]certify.WriteActionResult{
					"create_label": {
						Result: "leaked_resource", Cleanup: "delete_label", Verify: "read_back",
						Tag: "pm-cert-github-12345678-1700000000", Reason: "cleanup verification failed",
					},
				}
			},
		},
		{
			name: "duplicate required stage",
			mutate: func(rep *certify.Report) {
				rep.Stages = append(rep.Stages, rep.Stages[0])
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cf := certify.CredsFile{Connectors: map[string]certify.ConnectorCredsEntry{"github": {}}}
			batchDir := t.TempDir()
			first, err := certify.RunBatch(context.Background(), certify.BatchOptions{
				CredsFile: cf,
				RunnerFactory: func(name string, _ certify.Options) certify.Runnable {
					return &fakeRunnable{rep: passingReport(name)}
				},
				BatchDir: batchDir,
			})
			if err != nil {
				t.Fatal(err)
			}
			seed := findResult(t, first, "github").Report
			tc.mutate(&seed)
			if err := seed.Save(batchDir); err != nil {
				t.Fatal(err)
			}

			invoked := false
			batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
				CredsFile: cf,
				RunnerFactory: func(name string, _ certify.Options) certify.Runnable {
					invoked = true
					return &fakeRunnable{rep: passingReport(name)}
				},
				BatchDir: batchDir,
				Resume:   true,
			})
			if err != nil {
				t.Fatal(err)
			}
			if !invoked || findResult(t, batch, "github").Resumed {
				t.Fatal("invalid or edited report was trusted instead of rerun")
			}
		})
	}
}

func TestRunBatchReportPersistenceFailureIsSurfacedWithLeakPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name   string
		report func(string) certify.Report
		code   int
	}{
		{name: "passing report", report: passingReport, code: 1},
		{name: "leaked report", report: leakedReport, code: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			batchDir := t.TempDir()
			certDir := filepath.Join(batchDir, "certifications")
			if err := os.MkdirAll(certDir, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(t.TempDir(), "outside.json"), filepath.Join(certDir, "github.json")); err != nil {
				t.Fatal(err)
			}
			batch, err := certify.RunBatch(context.Background(), certify.BatchOptions{
				CredsFile: certify.CredsFile{Connectors: map[string]certify.ConnectorCredsEntry{"github": {}}},
				RunnerFactory: func(name string, _ certify.Options) certify.Runnable {
					return &fakeRunnable{rep: tc.report(name)}
				},
				BatchDir: batchDir,
			})
			if err != nil {
				t.Fatalf("RunBatch persistence failure should be represented in the batch result: %v", err)
			}
			if batch.ExitCode != tc.code {
				t.Fatalf("batch exit=%d, want %d after persistence failure", batch.ExitCode, tc.code)
			}
			result := findResult(t, batch, "github")
			if result.Error == "" {
				t.Fatal("batch result discarded report persistence failure")
			}
			if tc.code == 3 {
				found := false
				for _, stage := range result.Report.Stages {
					found = found || stage.Name == "report_persistence"
				}
				if !found {
					t.Fatal("leak-dominant batch result did not record report persistence failure")
				}
			}
		})
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
