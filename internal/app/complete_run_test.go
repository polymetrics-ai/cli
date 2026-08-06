package app

import (
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestCompleteRunReturnsCompletedRunOnPersistenceFailure(t *testing.T) {
	a := &App{state: state{
		Runs:        []Run{{ID: "run_persistence_failure", Type: "etl", Status: "running"}},
		Checkpoints: map[string]map[string]string{},
	}}
	result := etlExecutionResult{
		RecordsRead:        2,
		RecordsTransformed: 2,
		RecordsLoaded:      2,
		BatchCount:         1,
		Checkpoint:         map[string]string{"cursor": "unbounded-cursor"},
		PendingStreamState: &pendingStreamState{Key: "source_to_warehouse/records", State: StreamState{}},
		RateLimit: connectors.RateLimitSummary{Connectors: []connectors.RateLimitConnectorSummary{{
			Connector:   "sample",
			Declaration: connectors.RateLimitDeclarationUndeclared,
		}}},
	}
	run, err := a.completeRun("run_persistence_failure", result)
	if !IsCompletedRunPersistenceError(err) {
		t.Fatalf("completeRun error = %v, want persistence error", err)
	}
	if run.ID != "run_persistence_failure" || run.Status != "completed" || run.RecordsRead != 2 || run.RecordsLoaded != 2 || len(run.RateLimit.Connectors) != 1 {
		t.Fatalf("completeRun returned run = %+v", run)
	}
}
