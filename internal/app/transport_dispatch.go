package app

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

func hasDeclaredSyncTransport(source, destination connectors.Connector) bool {
	_, sourceDeclared := connectors.SyncTransportDescriptorOf(source)
	_, destinationDeclared := connectors.SyncTransportDescriptorOf(destination)
	return sourceDeclared || destinationDeclared
}

// runTransportETL is the bounded bridge from persisted connection state to
// the transport-neutral orchestrator. It does not call legacy Connector.Read
// or Connector.Write, inject provider metadata, or select a generic `upsert`
// action. Real durable warehouse/apply adapters remain separate foundations;
// this seam therefore fails closed unless a stage and externally verified
// transports are registered.
func (a *App) runTransportETL(ctx context.Context, runID string, conn Connection, source connectors.Connector, sourceRuntime connectors.RuntimeConfig, destination connectors.Connector, destRuntime connectors.RuntimeConfig, sourceExpectation synccontract.ResumeExpectation, streamName string, mode SyncMode, batchSize int) (etlExecutionResult, error) {
	if a.transports == nil {
		return etlExecutionResult{}, fmt.Errorf("closed transport registry is unavailable")
	}
	stateKey := streamStateKey(conn.Name, streamName)
	prior := a.state.StreamStates[stateKey]
	if prior.Checkpoint != nil {
		if err := validateStreamStateResume(prior, sourceExpectation); err != nil {
			return etlExecutionResult{}, err
		}
	}
	generationID := prior.GenerationID
	if generationID == 0 || mode.IsOverwrite() {
		generationID++
	}

	var committed *synccontract.CheckpointEnvelope
	transportResult, err := synctransport.NewOrchestrator(a.transports).Run(ctx, synctransport.RunRequest{
		Source:             source,
		SourceRuntime:      sourceRuntime,
		Destination:        destination,
		DestinationRuntime: destRuntime,
		Stream:             streamName,
		Mode:               mode.ContractMode,
		BatchSize:          batchSize,
		Resume:             sourceExpectation,
		Checkpoint:         prior.Checkpoint,
		Stage:              a.transportStage,
		Commit: func(checkpoint synccontract.CheckpointEnvelope) error {
			copy := checkpoint.Clone()
			committed = &copy
			return nil
		},
	})
	if err != nil {
		return etlExecutionResult{}, err
	}
	if committed == nil || committed.CommittedAt == nil {
		return etlExecutionResult{}, fmt.Errorf("closed transport completed without a durable committed checkpoint")
	}

	updated := StreamState{
		Connection:          conn.Name,
		Stream:              streamName,
		Checkpoint:          committed,
		GenerationID:        generationID,
		LastSuccessfulRunID: runID,
		RecordsLoaded:       transportResult.RecordsApplied,
		UpdatedAt:           *committed.CommittedAt,
	}
	result := etlExecutionResult{
		RecordsRead:   transportResult.RecordsRead,
		RecordsLoaded: transportResult.RecordsApplied,
		BatchCount:    transportResult.Pages,
		PendingStreamState: &pendingStreamState{
			Key:   stateKey,
			State: updated,
		},
	}
	result.Checkpoint = checkpointForResult(result, mode, stateKey, updated, "", false)
	return result, nil
}
