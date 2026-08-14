package app

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

// etlModeDispatchRequest groups values already resolved by RunETL before it
// chooses a closed transport, warehouse materializer, or connector executor.
// It is package-local plumbing only; no request reaches a public boundary in
// this shape.
type etlModeDispatchRequest struct {
	runID               string
	connection          Connection
	source              connectors.Connector
	sourceRuntime       connectors.RuntimeConfig
	destination         connectors.Connector
	destinationRuntime  connectors.RuntimeConfig
	sourceExpectation   synccontract.ResumeExpectation
	streamName          string
	stream              StreamConfig
	mode                SyncMode
	batchSize           int
	destinationApproval synctransport.DestinationApproval
}

func (a *App) dispatchETLMode(ctx context.Context, request etlModeDispatchRequest) (Run, error) {
	transportRoute := a.shouldRunTransport(request.connection, request.streamName, request.mode, request.source, request.destination)
	if !transportRoute && hasDestinationApproval(request.destinationApproval) {
		return a.failRun(request.runID, fmt.Errorf("destination approval is valid only for the closed issue-label transport route"))
	}
	if transportRoute {
		result, err := a.runTransportETL(ctx, request.runID, request.connection, request.source, request.sourceRuntime, request.destination, request.destinationRuntime, request.sourceExpectation, request.streamName, request.mode, request.batchSize, request.destinationApproval)
		if err != nil {
			return a.failAcknowledgedTransportRun(request.runID, result, err)
		}
		return a.completeAcknowledgedTransportRun(request.runID, result)
	}
	if request.mode.IsContractMode() {
		return a.failRun(request.runID, &synccontract.ModeNotExecutableError{
			Mode:   request.mode.ContractMode,
			Reason: "no matching closed source/destination transport has completed externally verified conformance",
		})
	}
	catalog, err := a.catalogForEndpoint(ctx, request.source, request.sourceRuntime, false)
	if err != nil {
		return a.failRun(request.runID, err)
	}
	request.sourceRuntime.ResolvedCatalog = &catalog
	var result etlExecutionResult
	if materializer, ok := request.destination.(connectors.LocalWarehouseMaterializer); ok && materializer.MaterializesLocalWarehouse() {
		result, err = a.runWarehouseETL(ctx, request.runID, request.connection, request.source, request.sourceRuntime, request.destination, request.destinationRuntime, request.sourceExpectation, request.streamName, request.stream, request.mode, request.batchSize)
	} else {
		result, err = a.runConnectorETL(ctx, request.runID, request.connection, request.source, request.sourceRuntime, request.destination, request.destinationRuntime, request.sourceExpectation, request.streamName, request.stream, request.mode, request.batchSize)
	}
	if err != nil {
		return a.failRun(request.runID, err)
	}
	return a.completeRun(request.runID, result)
}
