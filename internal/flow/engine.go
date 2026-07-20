package flow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"polymetrics.ai/internal/events"
	"polymetrics.ai/internal/telemetry"
)

// ETLResult is the result of a sync run.
type ETLResult struct {
	RecordsRead    int
	RecordsWritten int
}

// RLMResult is the result of an RLM scoring run.
type RLMResult struct {
	RecordsRead   int
	RecordsScored int
	RecordsFailed int
}

// RLMRunRequest is the engine-level RLM run request.
type RLMRunRequest struct {
	Spec     string
	Mode     string
	InTable  string
	OutTable string
	DryRun   bool
}

// AppAdapter abstracts the app layer for the engine.
type AppAdapter interface {
	ETLRun(ctx context.Context, connectionID string, streams []string) (ETLResult, error)
	QuerySQL(ctx context.Context, sql string, limit int) ([]map[string]any, error)
	RLMRun(ctx context.Context, req RLMRunRequest) (RLMResult, error)
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
func (e *Engine) Run(ctx context.Context, opts RunOptions) (result RunResult, err error) {
	result = RunResult{
		FlowName: e.Manifest.Name,
	}
	ctx, flowSpan := telemetry.StartSpan(ctx, "pm.flow.run", telemetry.StringAttr("pm.flow.name", e.Manifest.Name))
	defer func() {
		if err != nil {
			flowSpan.RecordError(err)
			flowSpan.SetAttributes(telemetry.StringAttr("pm.flow.status", "failed"))
		} else {
			flowSpan.SetAttributes(telemetry.StringAttr("pm.flow.status", result.Status))
		}
		flowSpan.End()
	}()

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
	e.emitFlowEvent(ctx, events.KindStarted, "running", "", StepResult{}, "")

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
		e.emitFlowEvent(ctx, events.KindCompleted, "dry_run", "", StepResult{}, "")
		return result, nil
	}

	var runErr error
	for _, id := range order {
		s := stepByID[id]
		sr := StepResult{ID: id, Kind: string(s.Kind)}
		stepCtx, stepSpan := telemetry.StartSpan(ctx, "pm.flow.step",
			telemetry.StringAttr("pm.flow.name", e.Manifest.Name),
			telemetry.StringAttr("pm.flow.step_id", id),
			telemetry.StringAttr("pm.flow.step_kind", string(s.Kind)),
		)

		// Check checkpoint.
		if e.Checkpoint != nil {
			chk, _ := e.Checkpoint.Get(e.Manifest.Name, id)
			if chk == "success" {
				sr.Status = "skipped"
				result.Steps = append(result.Steps, sr)
				e.emitStepEvent(ctx, events.KindSkipped, "skipped", sr, "")
				stepSpan.SetAttributes(telemetry.StringAttr("pm.flow.status", "skipped"))
				stepSpan.End()
				continue
			}
		}

		op := e.Manifest.Name + "/" + id
		e.appendLedger(stepCtx, op, "running", "")
		e.emitStepEvent(stepCtx, events.KindStarted, "running", sr, "")

		start := time.Now()
		var stepErr error
		var etlRes ETLResult

		switch s.Kind {
		case KindSync:
			etlRes, stepErr = e.App.ETLRun(stepCtx, s.Connection, s.Streams)
			sr.RecordsRead = etlRes.RecordsRead
			sr.RecordsWritten = etlRes.RecordsWritten
		case KindQuery:
			_, stepErr = e.App.QuerySQL(stepCtx, s.SQL, 0)
		case KindRLM:
			var rlmRes RLMResult
			rlmRes, stepErr = e.App.RLMRun(stepCtx, RLMRunRequest{
				Spec:     s.Spec,
				Mode:     s.Mode,
				InTable:  firstTable(s.In),
				OutTable: firstTable(s.Out),
			})
			sr.RecordsRead = rlmRes.RecordsRead
			sr.RecordsWritten = rlmRes.RecordsScored
			sr.RecordsFailed = rlmRes.RecordsFailed
		case KindAction:
			if e.ActionRunner == nil {
				stepErr = fmt.Errorf("action step %q: no ActionRunner configured", s.ID)
				break
			}
			// Fetch source records from the warehouse for this action step.
			var sourceRecords []map[string]any
			if s.ActionCfg != nil && s.ActionCfg.SourceTable != "" {
				sourceRecords, stepErr = e.App.QuerySQL(stepCtx, "SELECT * FROM "+s.ActionCfg.SourceTable, 0)
				if stepErr != nil {
					break
				}
			}
			runID := fmt.Sprintf("run-%d", start.UnixNano())
			var actionRes ActionResult
			actionRes, stepErr = e.ActionRunner.ExecuteStep(stepCtx, s, sourceRecords, opts.ApprovalToken, runID)
			sr.RecordsRead = actionRes.RecordsAttempted
			sr.RecordsWritten = actionRes.RecordsSucceeded
			sr.RecordsFailed = actionRes.RecordsFailed
			sr.DLQPath = actionRes.DLQPath
		default:
			stepErr = fmt.Errorf("%w: %s", ErrUnknownStepKind, s.Kind)
		}

		stepDuration := time.Since(start)
		sr.DurationNs = stepDuration.Nanoseconds()
		telemetry.RecordStageDuration(stepCtx, string(s.Kind), stepDuration)

		if stepErr != nil {
			sr.Status = "failed"
			sr.Error = stepErr.Error()
			e.appendLedger(stepCtx, op, "failed", stepErr.Error())
			e.emitStepEvent(stepCtx, events.KindFailed, "failed", sr, stepErr.Error())
			result.Steps = append(result.Steps, sr)
			stepSpan.RecordError(stepErr)
			stepSpan.SetAttributes(telemetry.StringAttr("pm.flow.status", "failed"))
			stepSpan.End()
			runErr = fmt.Errorf("%w: step %s: %v", ErrStepFailed, id, stepErr)
			break
		}

		sr.Status = "ok"
		e.appendLedger(stepCtx, op, "success", "")
		e.emitStepEvent(stepCtx, events.KindCompleted, "success", sr, "")
		stepSpan.SetAttributes(telemetry.StringAttr("pm.flow.status", "ok"))
		stepSpan.End()
		if e.Checkpoint != nil {
			_ = e.Checkpoint.Set(e.Manifest.Name, id, "success")
		}
		result.Steps = append(result.Steps, sr)
	}

	if runErr != nil {
		result.Status = "failed"
		e.appendLedger(ctx, flowOp, "failed", runErr.Error())
		e.emitFlowEvent(ctx, events.KindFailed, "failed", "", StepResult{}, runErr.Error())
		return result, runErr
	}

	result.Status = "ok"
	e.appendLedger(ctx, flowOp, "success", "")
	e.emitFlowEvent(ctx, events.KindCompleted, "success", "", StepResult{}, "")
	return result, nil
}

func (e *Engine) emitFlowEvent(ctx context.Context, kind events.Kind, status, stepID string, result StepResult, message string) {
	events.Emit(ctx, events.Event{
		Kind:    kind,
		Scope:   events.ScopeFlow,
		RunID:   e.Manifest.Name,
		StepID:  stepID,
		Status:  status,
		Message: message,
		Counters: events.Counters{
			RecordsRead:    int64(result.RecordsRead),
			RecordsWritten: int64(result.RecordsWritten),
			RecordsFailed:  int64(result.RecordsFailed),
		},
	})
}

func (e *Engine) emitStepEvent(ctx context.Context, kind events.Kind, status string, result StepResult, message string) {
	e.emitFlowEvent(ctx, kind, status, result.ID, result, message)
}

func firstTable(tables []string) string {
	if len(tables) == 0 {
		return ""
	}
	return tables[0]
}
