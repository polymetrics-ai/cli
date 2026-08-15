package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

var errTransportStreamStateConflict = errors.New("transport stream state changed in another process")

func hasDeclaredSyncTransport(source, destination connectors.Connector) bool {
	_, sourceDeclared := connectors.SyncTransportDescriptorOf(source)
	_, destinationDeclared := connectors.SyncTransportDescriptorOf(destination)
	// A transport is a closed pair: registry preflight needs a source and a
	// destination declaration. Requiring both here prevents a newly declared
	// primitive destination from rerouting an existing legacy source before that
	// source has declared and registered its own transport role. Invalid
	// two-sided declarations still reach preflight and fail closed there.
	return sourceDeclared && destinationDeclared
}

// shouldRunTransport keeps the closed issue-label walking slice opt-in at the
// persisted connection boundary. Open installs the exact definition-owned
// composition so it can be preflighted, but that must not turn every existing
// JSON ETL connection into a transport run merely because its connector now
// advertises the closed descriptor.
//
// Other externally declared transport pairs retain the normal descriptor-led
// dispatch. The walking slice participates only when the connection itself
// satisfies its fixed issues/source-issue/target-issue/label contract. The
// definition selects the allowed destination action for each declared mode;
// non-additive modes are additionally gated by the persisted connection.
func (a *App) shouldRunTransport(conn Connection, streamName string, mode SyncMode, source, destination connectors.Connector) bool {
	sourceIssueLabel := isIssueLabelTransportConnector(source)
	destinationIssueLabel := isIssueLabelTransportConnector(destination)
	if sourceIssueLabel || destinationIssueLabel {
		// The closed composition is a same-definition source/destination pair.
		// A one-sided descriptor is an ordinary legacy ETL connection, not a
		// half-transport route; in particular, it must not divert a historical
		// declarative source into the warehouse destination transport preflight.
		if !sourceIssueLabel || !destinationIssueLabel {
			return false
		}
	} else {
		return hasDeclaredSyncTransport(source, destination)
	}
	if streamName != "issues" {
		return false
	}
	transportConn, err := a.issueLabelTransportConnection(conn.ID)
	if err != nil {
		return false
	}
	contract, err := a.issueLabelTransportContract(transportConn)
	if err != nil {
		return false
	}
	_, err = contract.actionForSyncMode(mode.ContractMode)
	return err == nil
}

func isIssueLabelTransportConnector(connector connectors.Connector) bool {
	descriptor, ok := connectors.SyncTransportDescriptorOf(connector)
	if !ok || descriptor.Source == nil || descriptor.Destination == nil {
		return false
	}
	definition, ok := connectors.DefinitionOf(connector)
	if !ok {
		return false
	}
	contract, err := issueLabelTransportContractForDefinition(definition)
	if err != nil {
		return false
	}
	if descriptor.Source.Executor != issueLabelSourceReference || descriptor.Destination.Executor != issueLabelDestinationReference {
		return false
	}
	if len(descriptor.Source.EligibleStreams) != 1 || descriptor.Source.EligibleStreams[0] != contract.stream {
		return false
	}
	wantModes := contract.modes()
	if len(descriptor.Source.Modes) != len(wantModes) || len(descriptor.Destination.Modes) != len(wantModes) {
		return false
	}
	for i, mode := range wantModes {
		if descriptor.Source.Modes[i] != mode || descriptor.Destination.Modes[i] != mode {
			return false
		}
	}
	wantActions := contract.destinationActionNames()
	if len(descriptor.Destination.EligibleActions) != len(wantActions) || len(descriptor.Destination.ApplyStrategies) != len(wantModes) {
		return false
	}
	strategies, err := contract.applyStrategies()
	if err != nil {
		return false
	}
	for i := range wantActions {
		if descriptor.Destination.EligibleActions[i] != wantActions[i] || descriptor.Destination.ApplyStrategies[i] != strategies[i] {
			return false
		}
	}
	return true
}

// runTransportETL is the bounded bridge from persisted connection state to
// the transport-neutral orchestrator. It does not call legacy Connector.Read
// or Connector.Write, inject provider metadata, or select a generic `upsert`
// action. Real durable warehouse/apply adapters remain separate foundations;
// this seam therefore fails closed unless a stage and externally verified
// transports are registered.
func (a *App) runTransportETL(ctx context.Context, runID string, conn Connection, source connectors.Connector, sourceRuntime connectors.RuntimeConfig, destination connectors.Connector, destRuntime connectors.RuntimeConfig, sourceExpectation synccontract.ResumeExpectation, streamName string, mode SyncMode, batchSize int, approval synctransport.DestinationApproval) (etlExecutionResult, error) {
	if a.transports == nil {
		return etlExecutionResult{}, fmt.Errorf("closed transport registry is unavailable")
	}
	stateKey := streamStateKey(conn.Name, streamName)
	prior, priorPresent := a.state.StreamStates[stateKey]
	prior = cloneStreamState(prior)
	if prior.Checkpoint != nil {
		if err := validateStreamStateResume(prior, sourceExpectation); err != nil {
			return etlExecutionResult{}, err
		}
	}
	generationID := prior.GenerationID
	if generationID == 0 || mode.IsOverwrite() {
		generationID++
	}

	expectedState := cloneStreamState(prior)
	expectedPresent := priorPresent
	var committed *synccontract.CheckpointEnvelope
	transportResult, err := synctransport.NewOrchestrator(a.transports).Run(ctx, synctransport.RunRequest{
		ConnectionID:       conn.ID,
		Generation:         generationID,
		Source:             source,
		SourceRuntime:      sourceRuntime,
		Destination:        destination,
		DestinationRuntime: destRuntime,
		Stream:             streamName,
		Mode:               mode.ContractMode,
		BatchSize:          batchSize,
		Resume:             sourceExpectation,
		Checkpoint:         prior.Checkpoint,
		Approval:           approval,
		Stage:              a.transportStage,
		Commit: func(checkpoint synccontract.CheckpointEnvelope) error {
			interim := checkpoint.Clone()
			if interim.CommittedAt == nil {
				return fmt.Errorf("closed transport committed checkpoint is missing its acknowledgement timestamp")
			}
			if err := interim.ValidateResume(sourceExpectation); err != nil {
				return err
			}
			interimState := cloneStreamState(expectedState)
			interimState.Connection = conn.Name
			interimState.Stream = streamName
			interimState.Checkpoint = &interim
			interimState.GenerationID = generationID
			interimState.UpdatedAt = *interim.CommittedAt
			if _, err := a.updateState(func(current state) (state, error) {
				currentState, currentPresent := current.StreamStates[stateKey]
				if currentPresent != expectedPresent || (currentPresent && !transportStreamStateEqual(currentState, expectedState)) {
					return current, errTransportStreamStateConflict
				}
				if current.StreamStates == nil {
					current.StreamStates = map[string]StreamState{}
				}
				current.StreamStates[stateKey] = cloneStreamState(interimState)
				return current, nil
			}); err != nil {
				return fmt.Errorf("persist acknowledged transport checkpoint: %w", err)
			}
			expectedState = cloneStreamState(interimState)
			expectedPresent = true
			copy := interim.Clone()
			committed = &copy
			return nil
		},
	})
	if committed == nil || committed.CommittedAt == nil {
		if err != nil {
			return etlExecutionResult{}, err
		}
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
	if err != nil {
		return result, err
	}
	return result, nil
}

func transportStreamStateEqual(left, right StreamState) bool {
	if left.Connection != right.Connection ||
		left.Stream != right.Stream ||
		left.GenerationID != right.GenerationID ||
		left.LastSuccessfulRunID != right.LastSuccessfulRunID ||
		left.RecordsLoaded != right.RecordsLoaded ||
		!left.UpdatedAt.Equal(right.UpdatedAt) {
		return false
	}
	return transportCheckpointEqual(left.Checkpoint, right.Checkpoint)
}

func transportCheckpointEqual(left, right *synccontract.CheckpointEnvelope) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	if left.StateVersion != right.StateVersion ||
		left.Source.Engine != right.Source.Engine ||
		left.Source.AccountOrCluster != right.Source.AccountOrCluster ||
		left.Source.ObjectScope != right.Source.ObjectScope ||
		left.Mechanism != right.Mechanism ||
		!transportSnapshotBarrierEqual(left.SnapshotBarrier, right.SnapshotBarrier) ||
		!transportCheckpointPositionEqual(left.Position, right.Position) ||
		!transportOptionalBoolEqual(left.PositionObserved, right.PositionObserved) ||
		!transportOpaqueTokenEqual(left.SourceGeneration, right.SourceGeneration) ||
		left.SchemaVersion != right.SchemaVersion ||
		left.ProtocolVersion != right.ProtocolVersion ||
		left.Dedupe.Kind != right.Dedupe.Kind ||
		!transportOpaqueTokenEqual(left.Dedupe.Value, right.Dedupe.Value) ||
		left.DedupeWindow.Kind != right.DedupeWindow.Kind ||
		!transportOpaqueTokenEqual(left.DedupeWindow.Start, right.DedupeWindow.Start) ||
		!transportOpaqueTokenEqual(left.DedupeWindow.End, right.DedupeWindow.End) ||
		!left.ObservedAt.Equal(right.ObservedAt) ||
		!transportOptionalTimeEqual(left.CommittedAt, right.CommittedAt) ||
		(left.Partitions == nil) != (right.Partitions == nil) ||
		len(left.Partitions) != len(right.Partitions) {
		return false
	}
	for index := range left.Partitions {
		if !transportOpaqueTokenEqual(left.Partitions[index].Partition, right.Partitions[index].Partition) ||
			!transportCheckpointPositionEqual(left.Partitions[index].Position, right.Partitions[index].Position) {
			return false
		}
	}
	return true
}

func transportSnapshotBarrierEqual(left, right *synccontract.SnapshotBarrier) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Kind == right.Kind && transportOpaqueTokenEqual(left.Token, right.Token)
}

func transportCheckpointPositionEqual(left, right synccontract.CheckpointPosition) bool {
	return transportOpaqueTokenEqual(left.Primary, right.Primary) && transportOpaqueTokenEqual(left.TieBreaker, right.TieBreaker)
}

func transportOpaqueTokenEqual(left, right synccontract.OpaqueToken) bool {
	return (left == nil) == (right == nil) && bytes.Equal(left, right)
}

func transportOptionalBoolEqual(left, right *bool) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func transportOptionalTimeEqual(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}
