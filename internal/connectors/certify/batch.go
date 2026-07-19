// Batch mode (certification design §B): drives certify.Runner (or a test
// double satisfying Runnable) over every connector listed in a CredsFile,
// using a worker pool bounded by Defaults.Parallel; stages within one
// connector stay strictly serial (Runner.Run already guarantees this —
// batch mode only parallelizes ACROSS connectors, never within one). Batch
// mode never runs a Runner concurrently for the same connector name.
package certify

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"polymetrics.ai/internal/events"
)

// Runnable is the subset of *Runner's behavior batch mode depends on,
// letting tests substitute a fake without driving the real CLI. *Runner
// (stages_source.go) satisfies this interface.
type Runnable interface {
	Run(ctx context.Context) (Report, error)
}

// RunnerFactory builds a Runnable for one connector from settings that passed
// credential-file constraint validation.
// Production callers pass a factory that wraps NewRunner; tests substitute
// their own for isolation from the real CLI/network.
type RunnerFactory func(connector string, opts Options) Runnable

// BatchOptions configures RunBatch (certification design §B).
type BatchOptions struct {
	CredsFile CredsFile
	// RunnerFactory builds the Runnable for each connector. Required —
	// RunBatch has no built-in default so tests never accidentally drive
	// the real CLI/network; production callers (certify_cli.go) supply
	// DefaultRunnerFactory.
	RunnerFactory RunnerFactory
	// BatchDir is the project root under which per-connector reports
	// (Report.Save's <dir>/certifications/<connector>.json) and this
	// batch's progress.json are written.
	BatchDir string
	// Resume skips any connector whose existing
	// <BatchDir>/certifications/<connector>.json report completed after
	// this RunBatch call's start time (certification design §B: "--resume
	// skips connectors whose report is newer than batch start").
	Resume bool
}

// BatchConnectorResult is one connector's outcome within a BatchReport.
type BatchConnectorResult struct {
	Connector  string `json:"connector"`
	Report     Report `json:"report,omitempty"`
	Skipped    bool   `json:"skipped,omitempty"`
	SkipReason string `json:"skip_reason,omitempty"`
	Resumed    bool   `json:"resumed,omitempty"`
	Error      string `json:"error,omitempty"`
	ExitCode   int    `json:"exit_code"`
}

// BatchReport is RunBatch's aggregate result: one BatchConnectorResult per
// creds-file connector plus a batch-level exit code (certification design
// §A exit codes, rolled up across every connector — leaks dominate).
type BatchReport struct {
	RunID     string                 `json:"run_id"`
	StartedAt time.Time              `json:"started_at"`
	Results   []BatchConnectorResult `json:"results"`
	ExitCode  int                    `json:"exit_code"`
}

// Leaks collects every leak across every connector's report, for a batch's
// top-level "leaked resources" summary (certification design §B: "any leak
// row printed first and forces exit 3").
func (b BatchReport) Leaks() []Leak {
	var out []Leak
	for _, r := range b.Results {
		out = append(out, r.Report.Leaks...)
	}
	return out
}

// MatrixRow is one connector's row in the batch summary matrix
// (certification design §B columns: check/catalog/read/5 modes/resume/
// write/flow/schedule/redaction/leaks).
type MatrixRow struct {
	Connector                   string
	Check                       string
	Catalog                     string
	Read                        string
	FullRefreshAppend           string
	FullRefreshOverwrite        string
	FullRefreshOverwriteDeduped string
	IncrementalAppend           string
	IncrementalAppendDeduped    string
	Resume                      string
	Write                       string
	Flow                        string
	Schedule                    string
	Redaction                   string
	Leaked                      bool
}

// SummaryMatrix is the rendered certification design §B matrix: rows =
// connectors, leak rows sorted first (design: "any leak row printed first"),
// remaining rows in stable sorted-by-name order.
type SummaryMatrix struct {
	Rows []MatrixRow
}

// SummaryMatrix builds b's rendered matrix.
func (b BatchReport) SummaryMatrix() SummaryMatrix {
	rows := make([]MatrixRow, 0, len(b.Results))
	for _, r := range b.Results {
		rows = append(rows, matrixRowFor(r))
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Leaked != rows[j].Leaked {
			return rows[i].Leaked // leaked rows sort first
		}
		return rows[i].Connector < rows[j].Connector
	})
	return SummaryMatrix{Rows: rows}
}

func matrixRowFor(r BatchConnectorResult) MatrixRow {
	row := MatrixRow{Connector: r.Connector, Leaked: len(r.Report.Leaks) != 0}
	if r.Skipped {
		row.Check = "skip"
		row.Catalog = "skip"
		row.Read = "skip"
		row.Resume = "skip"
		row.Write = "skip"
		row.Flow = "skip"
		row.Schedule = "skip"
		row.Redaction = "skip"
		return row
	}
	rep := r.Report
	row.Check = resultOrUncert(rep.Capabilities.Check.Result)
	row.Catalog = resultOrUncert(rep.Capabilities.Catalog.Result)
	row.Read = resultOrUncert(rep.Capabilities.Read.Result)
	row.FullRefreshAppend = syncModeResult(rep, "full_refresh_append")
	row.FullRefreshOverwrite = syncModeResult(rep, "full_refresh_overwrite")
	row.FullRefreshOverwriteDeduped = syncModeResult(rep, "full_refresh_overwrite_deduped")
	row.IncrementalAppend = syncModeResult(rep, "incremental_append")
	row.IncrementalAppendDeduped = syncModeResult(rep, "incremental_append_deduped")
	row.Resume = resultOrUncert(rep.Capabilities.Resume.Result)
	row.Redaction = resultOrUncert(rep.Capabilities.SecretRedaction.Result)
	if rep.Capabilities.Flow != nil {
		row.Flow = resultOrUncert(rep.Capabilities.Flow.Result)
	} else {
		row.Flow = "uncert"
	}
	if rep.Capabilities.Schedule != nil {
		row.Schedule = resultOrUncert(rep.Capabilities.Schedule.Result)
	} else {
		row.Schedule = "uncert"
	}
	row.Write = writeActionsSummary(rep.Capabilities.WriteActions)
	return row
}

func resultOrUncert(result string) string {
	if result == "" {
		return "uncert"
	}
	return result
}

func syncModeResult(rep Report, mode string) string {
	if r, ok := rep.Capabilities.SyncModes[mode]; ok {
		return resultOrUncert(r.Result)
	}
	return "uncert"
}

func writeActionsSummary(actions map[string]WriteActionResult) string {
	if len(actions) == 0 {
		return "uncert"
	}
	names := make([]string, 0, len(actions))
	for name := range actions {
		names = append(names, name)
	}
	sort.Strings(names)
	worst := "pass"
	for _, name := range names {
		r := actions[name].Result
		if r != "pass" && r != "skipped" {
			worst = r
		}
	}
	return worst
}

// RunBatch drives one Runnable per opts.CredsFile.Connectors entry
// (certification design §B), honoring skip:true entries and (with
// opts.Resume) fresh existing reports, using a worker pool bounded by
// opts.CredsFile.Defaults.Parallel (minimum 1). A per-connector Runnable
// error is captured into that connector's BatchConnectorResult.Error rather
// than aborting the whole batch — RunBatch itself only returns an error for
// unrecoverable setup failures (e.g. an unwritable BatchDir).
func RunBatch(ctx context.Context, opts BatchOptions) (BatchReport, error) {
	if opts.RunnerFactory == nil {
		return BatchReport{}, fmt.Errorf("certify: BatchOptions.RunnerFactory is required")
	}
	if err := ValidateBatchConstraints(opts.CredsFile); err != nil {
		return BatchReport{}, err
	}

	runID, err := newRunID()
	if err != nil {
		return BatchReport{}, fmt.Errorf("certify: generate batch run id: %w", err)
	}

	batch := BatchReport{RunID: runID, StartedAt: time.Now().UTC()}

	names := opts.CredsFile.ConnectorNames()
	parallel := opts.CredsFile.Defaults.Parallel
	if parallel < 1 {
		parallel = 1
	}
	emitBatchEvent(ctx, events.KindStarted, runID, "", "running", len(names), 0, "")

	results := make([]BatchConnectorResult, len(names))
	var completed atomic.Int64
	type batchJob struct {
		index int
		name  string
		entry ConnectorCredsEntry
	}
	jobs := make([]batchJob, 0, len(names))

	for i, name := range names {
		entry := opts.CredsFile.Connectors[name]

		if entry.Skip {
			results[i] = BatchConnectorResult{Connector: name, Skipped: true, SkipReason: entry.Reason}
			done := int(completed.Add(1))
			emitBatchEvent(ctx, events.KindSkipped, runID, name, "skipped", len(names), done, entry.Reason)
			continue
		}

		if opts.Resume && hasFreshReport(opts.BatchDir, name, batch.StartedAt) {
			results[i] = BatchConnectorResult{Connector: name, Resumed: true}
			done := int(completed.Add(1))
			emitBatchEvent(ctx, events.KindResumed, runID, name, "resumed", len(names), done, "")
			continue
		}

		emitBatchEvent(ctx, events.KindQueued, runID, name, "queued", len(names), int(completed.Load()), "")
		jobs = append(jobs, batchJob{index: i, name: name, entry: entry})
	}

	jobCh := make(chan batchJob)
	var wg sync.WaitGroup
	for workerID := 0; workerID < parallel; workerID++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				emitBatchEvent(ctx, events.KindStarted, runID, job.name, "running", len(names), int(completed.Load()), "")
				runnerOpts := optionsFromCredsEntry(job.name, opts.CredsFile, job.entry)
				runner := opts.RunnerFactory(job.name, runnerOpts)
				rep, runErr := runner.Run(ctx)

				result := BatchConnectorResult{Connector: job.name, Report: rep}
				if runErr != nil {
					result.Error = runErr.Error()
					result.ExitCode = 2
					results[job.index] = result
					done := int(completed.Add(1))
					emitBatchEvent(ctx, events.KindFailed, runID, job.name, "failed", len(names), done, result.Error)
					continue
				}
				result.ExitCode = ExitCodeFor(rep)
				if opts.BatchDir != "" && rep.Connector != "" {
					_ = rep.Save(opts.BatchDir)
				}
				results[job.index] = result
				done := int(completed.Add(1))
				if result.ExitCode == 0 {
					emitBatchEvent(ctx, events.KindCompleted, runID, job.name, "success", len(names), done, "")
				} else {
					emitBatchEvent(ctx, events.KindFailed, runID, job.name, "failed", len(names), done, result.Error)
				}
			}
		}()
	}
	for _, job := range jobs {
		jobCh <- job
	}
	close(jobCh)
	wg.Wait()

	batch.Results = results
	batch.ExitCode = aggregateExitCode(results)

	if opts.BatchDir != "" {
		if err := writeBatchProgress(opts.BatchDir, batch); err != nil {
			emitBatchEvent(ctx, events.KindFailed, runID, "", "failed", len(names), int(completed.Load()), err.Error())
			return batch, err
		}
	}

	if batch.ExitCode == 0 {
		emitBatchEvent(ctx, events.KindCompleted, runID, "", "success", len(names), int(completed.Load()), "")
	} else {
		emitBatchEvent(ctx, events.KindFailed, runID, "", "failed", len(names), int(completed.Load()), "")
	}
	return batch, nil
}

func emitBatchEvent(ctx context.Context, kind events.Kind, runID, connector, status string, total, completed int, message string) {
	events.Emit(ctx, events.Event{
		Kind:    kind,
		Scope:   events.ScopeCertify,
		RunID:   runID,
		StepID:  connector,
		Status:  status,
		Message: message,
		Counters: events.Counters{
			Completed: int64(completed),
			Total:     int64(total),
		},
	})
}

// ValidateBatchConstraints rejects credential-file settings that the current
// Runner cannot enforce. Sandbox is enforced here as a mandatory gate for
// writes; rate, budget, and limit settings fail closed before any runner is
// constructed instead of being silently discarded.
func ValidateBatchConstraints(cf CredsFile) error {
	for _, name := range cf.ConnectorNames() {
		entry := cf.Connectors[name]
		if entry.Skip {
			continue
		}
		effective := cf.EffectiveOptions(name)
		switch {
		case effective.Write && !effective.Sandbox:
			return fmt.Errorf("certify: connector %q write=true without sandbox is not supported", name)
		case effective.RateLimitRPS != 0:
			return fmt.Errorf("certify: connector %q rate_limit_rps is not supported by the certification runner", name)
		case effective.BudgetCalls != 0:
			return fmt.Errorf("certify: connector %q budget_calls is not supported by the certification runner", name)
		case effective.Limit != 0:
			return fmt.Errorf("certify: connector %q limit is not supported by the certification runner", name)
		}
	}
	return nil
}

// optionsFromCredsEntry derives only settings the current Runner enforces.
// ValidateBatchConstraints must run before this conversion.
func optionsFromCredsEntry(name string, _ CredsFile, entry ConnectorCredsEntry) Options {
	return Options{
		Connector: name,
		Config:    entry.Credential.Config,
		SecretEnv: entry.Credential.FromEnv,
		Write:     entry.Write,
	}
}

// aggregateExitCode rolls per-connector exit codes up to a single batch exit
// code: any leak (exit 3) dominates, else any failure (exit 2), else 0.
func aggregateExitCode(results []BatchConnectorResult) int {
	worst := 0
	for _, r := range results {
		if r.ExitCode > worst {
			worst = r.ExitCode
		}
	}
	return worst
}

// hasFreshReport reports whether <dir>/certifications/<connector>.json
// exists and its CompletedAt is at or after since (certification design §B
// --resume semantics).
func hasFreshReport(dir, connector string, since time.Time) bool {
	if dir == "" {
		return false
	}
	path := filepath.Join(dir, certificationsDirName, connector+".json")
	rep, err := LoadReport(path)
	if err != nil {
		return false
	}
	return !rep.CompletedAt.Before(since)
}

// writeBatchProgress persists <dir>/certifications/batch-<runid>/
// progress.json (certification design §B resumability marker).
func writeBatchProgress(dir string, batch BatchReport) error {
	progressDir := filepath.Join(dir, certificationsDirName, "batch-"+batch.RunID)
	if err := os.MkdirAll(progressDir, 0o755); err != nil {
		return fmt.Errorf("certify: create batch progress dir: %w", err)
	}
	raw, err := json.MarshalIndent(batch, "", "  ")
	if err != nil {
		return fmt.Errorf("certify: marshal batch progress: %w", err)
	}
	path := filepath.Join(progressDir, "progress.json")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return fmt.Errorf("certify: write %s: %w", path, err)
	}
	return nil
}

// newRunID generates a short random hex id for a batch run
// (certification design §B: "certifications/batch-<runid>/progress.json").
func newRunID() (string, error) {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// DefaultRunnerFactory adapts NewRunner to the RunnerFactory signature, for
// production callers (certify_cli.go) that want the real CLI-driving
// Runner.
func DefaultRunnerFactory(_ string, opts Options) Runnable {
	return NewRunner(opts)
}
