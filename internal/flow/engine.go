package flow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ETLResult is the result of a sync run.
type ETLResult struct {
	RecordsRead    int
	RecordsWritten int
}

// AppAdapter abstracts the app layer for the engine.
type AppAdapter interface {
	ETLRun(ctx context.Context, connectionID string, streams []string) (ETLResult, error)
	QuerySQL(ctx context.Context, sql string, limit int) ([]map[string]any, error)
}

// LedgerAdapter abstracts the ledger for the engine.
type LedgerAdapter interface {
	Append(ctx context.Context, record LedgerRecord) error
}

// LedgerRecord is the per-step / per-flow record written to the ledger.
type LedgerRecord struct {
	Mode      string
	Operation string
	Status    string
	Error     string
}

// RunOptions controls how a flow run behaves.
type RunOptions struct {
	DryRun        bool
	Force         bool
	JSON          bool
	ApprovalToken string // required when any KindAction step is present
	PerAction     bool   // each action step gets its own token (future)
}

// StepResult is the per-step outcome in a RunResult.
type StepResult struct {
	ID             string `json:"id"`
	Kind           string `json:"kind"`
	Status         string `json:"status"`
	RecordsRead    int    `json:"records_read"`
	RecordsWritten int    `json:"records_written"`
	RecordsFailed  int    `json:"records_failed,omitempty"`
	DurationNs     int64  `json:"duration_ns"`
	Error          string `json:"error,omitempty"`
	DLQPath        string `json:"dlq_path,omitempty"`
	SchemaDrift    bool   `json:"schema_drift,omitempty"`
}

// RunResult is the top-level outcome of Engine.Run.
type RunResult struct {
	FlowName string       `json:"flow_name"`
	Status   string       `json:"status"`
	Steps    []StepResult `json:"steps"`
}

// Engine executes a FlowManifest.
type Engine struct {
	Manifest     FlowManifest
	App          AppAdapter
	Ledger       LedgerAdapter
	Checkpoint   CheckpointStore
	LockDir      string
	ActionRunner StepActionRunner // optional; required when manifest has KindAction steps
}

func (e *Engine) lockPath() string {
	return filepath.Join(e.LockDir, "flow-"+e.Manifest.Name+".lock")
}

// acquireLease writes the lock file; returns ErrLeaseHeld if it already exists.
func (e *Engine) acquireLease() error {
	p := e.lockPath()
	// Try exclusive create.
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("%w", ErrLeaseHeld)
		}
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "%d\n", os.Getpid())
	return nil
}

func (e *Engine) releaseLease() {
	_ = os.Remove(e.lockPath())
}

func (e *Engine) appendLedger(ctx context.Context, op, status, errMsg string) {
	if e.Ledger == nil {
		return
	}
	_ = e.Ledger.Append(ctx, LedgerRecord{
		Mode:      "flow",
		Operation: op,
		Status:    status,
		Error:     errMsg,
	})
}

// Run executes the flow and returns a RunResult.
func (e *Engine) Run(ctx context.Context, opts RunOptions) (RunResult, error) {
	result := RunResult{
		FlowName: e.Manifest.Name,
	}

	// Acquire lease.
	if err := e.acquireLease(); err != nil {
		return result, err
	}
	defer e.releaseLease()

	// Force: clear checkpoints.
	if opts.Force && e.Checkpoint != nil {
		if err := e.Checkpoint.Clear(e.Manifest.Name); err != nil {
			return result, err
		}
	}

	// Pre-flight: if any action step is present and no token is provided, fail fast.
	if !opts.DryRun {
		for _, s := range e.Manifest.Steps {
			if s.Kind == KindAction && opts.ApprovalToken == "" {
				return result, ErrApprovalRequired
			}
		}
	}

	// Compute topological order.
	order, err := BuildDAG(e.Manifest)
	if err != nil {
		return result, err
	}

	// Build step index.
	stepByID := map[string]FlowStep{}
	for _, s := range e.Manifest.Steps {
		stepByID[s.ID] = s
	}

	flowOp := e.Manifest.Name

	if opts.DryRun {
		for _, id := range order {
			s := stepByID[id]
			result.Steps = append(result.Steps, StepResult{
				ID:     id,
				Kind:   string(s.Kind),
				Status: "dry_run",
			})
		}
		result.Status = "dry_run"
		e.appendLedger(ctx, flowOp, "dry_run", "")
		return result, nil
	}

	var runErr error
	for _, id := range order {
		s := stepByID[id]
		sr := StepResult{ID: id, Kind: string(s.Kind)}

		// Check checkpoint.
		if e.Checkpoint != nil {
			chk, _ := e.Checkpoint.Get(e.Manifest.Name, id)
			if chk == "success" {
				sr.Status = "skipped"
				result.Steps = append(result.Steps, sr)
				continue
			}
		}

		op := e.Manifest.Name + "/" + id
		e.appendLedger(ctx, op, "running", "")

		start := time.Now()
		var stepErr error
		var etlRes ETLResult

		switch s.Kind {
		case KindSync:
			etlRes, stepErr = e.App.ETLRun(ctx, s.Connection, s.Streams)
			sr.RecordsRead = etlRes.RecordsRead
			sr.RecordsWritten = etlRes.RecordsWritten
		case KindQuery:
			_, stepErr = e.App.QuerySQL(ctx, s.SQL, 0)
		case KindAction:
			if e.ActionRunner == nil {
				stepErr = fmt.Errorf("action step %q: no ActionRunner configured", s.ID)
				break
			}
			// Fetch source records from the warehouse for this action step.
			var sourceRecords []map[string]any
			if s.ActionCfg != nil && s.ActionCfg.SourceTable != "" {
				sourceRecords, stepErr = e.App.QuerySQL(ctx, "SELECT * FROM "+s.ActionCfg.SourceTable, 0)
				if stepErr != nil {
					break
				}
			}
			runID := fmt.Sprintf("run-%d", start.UnixNano())
			var actionRes ActionResult
			actionRes, stepErr = e.ActionRunner.ExecuteStep(ctx, s, sourceRecords, opts.ApprovalToken, runID)
			sr.RecordsRead = actionRes.RecordsAttempted
			sr.RecordsWritten = actionRes.RecordsSucceeded
			sr.RecordsFailed = actionRes.RecordsFailed
			sr.DLQPath = actionRes.DLQPath
		default:
			stepErr = fmt.Errorf("%w: %s", ErrUnknownStepKind, s.Kind)
		}

		sr.DurationNs = time.Since(start).Nanoseconds()

		if stepErr != nil {
			sr.Status = "failed"
			sr.Error = stepErr.Error()
			e.appendLedger(ctx, op, "failed", stepErr.Error())
			result.Steps = append(result.Steps, sr)
			runErr = fmt.Errorf("%w: step %s: %v", ErrStepFailed, id, stepErr)
			break
		}

		sr.Status = "ok"
		e.appendLedger(ctx, op, "success", "")
		if e.Checkpoint != nil {
			_ = e.Checkpoint.Set(e.Manifest.Name, id, "success")
		}
		result.Steps = append(result.Steps, sr)
	}

	if runErr != nil {
		result.Status = "failed"
		e.appendLedger(ctx, flowOp, "failed", runErr.Error())
		return result, runErr
	}

	result.Status = "ok"
	e.appendLedger(ctx, flowOp, "success", "")
	return result, nil
}
