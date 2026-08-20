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
	runID                       string
	connection                  Connection
	source                      connectors.Connector
	sourceRuntime               connectors.RuntimeConfig
	destination                 connectors.Connector
	destinationRuntime          connectors.RuntimeConfig
	sourceExpectation           synccontract.ResumeExpectation
	streamName                  string
	stream                      StreamConfig
	mode                        SyncMode
	batchSize                   int
	maxInFlightBatches          int
	destinationApproval         synctransport.DestinationApproval
	rateParkingResumeCheckpoint *synccontract.CheckpointEnvelope
}

func (a *App) dispatchETLMode(ctx context.Context, request etlModeDispatchRequest) (Run, error) {
	if request.mode.ContractMode == synccontract.ModeChangeCapture && hasImplementedChangefeedExecutor(request.source) {
		if hasDestinationApproval(request.destinationApproval) {
			return a.failRun(request.runID, fmt.Errorf("destination approval is not valid for source-only change capture into the connection warehouse"))
		}
		materializer, ok := request.destination.(connectors.LocalWarehouseMaterializer)
		if !ok || !materializer.MaterializesLocalWarehouse() {
			return a.failRun(request.runID, &synccontract.ModeNotExecutableError{
				Mode:   synccontract.ModeChangeCapture,
				Reason: "change capture requires a connection-owned local warehouse destination",
			})
		}
		result, err := a.runWarehouseChangeCapture(ctx, request)
		if err != nil {
			return a.failAcknowledgedTransportRun(request.runID, result, err)
		}
		return a.completeAcknowledgedTransportRun(request.runID, result)
	}
	transportRoute := a.shouldRunTransport(request.connection, request.streamName, request.mode, request.source, request.destination)
	if request.maxInFlightBatches > 0 && (!transportRoute || !isOrderedArrowFullOverwriteCandidate(request)) {
		return a.failRun(request.runID, &synctransport.OrderedPipelineUnsupportedError{Source: request.source.Name(), Destination: request.destination.Name()})
	}
	if !transportRoute && hasDestinationApproval(request.destinationApproval) {
		return a.failRun(request.runID, fmt.Errorf("destination approval is valid only for a closed definition-selected transport route"))
	}
	if transportRoute {
		maxInFlightBatches := request.maxInFlightBatches
		if maxInFlightBatches == 0 && isOrderedArrowFullOverwriteCandidate(request) {
			maxInFlightBatches = 2
		}
		result, err := a.runTransportETL(ctx, request.runID, request.connection, request.source, request.sourceRuntime, request.destination, request.destinationRuntime, request.sourceExpectation, request.streamName, request.stream, request.mode, request.batchSize, maxInFlightBatches, request.destinationApproval, request.rateParkingResumeCheckpoint)
		if err != nil {
			if origin, tagged := synctransport.TransportExecutionOriginOf(err); tagged && origin == synctransport.TransportExecutionOriginSource {
				if parked, handled, parkErr := a.parkRateLimitedRun(ctx, request, result, err); handled {
					return parked, parkErr
				}
			}
			return a.failAcknowledgedTransportRun(request.runID, result, err)
		}
		return a.completeAcknowledgedTransportRun(request.runID, result)
	}
	if request.mode.ContractMode == synccontract.ModeChangeCapture {
		materializer, ok := request.destination.(connectors.LocalWarehouseMaterializer)
		if !ok || !materializer.MaterializesLocalWarehouse() {
			return a.failRun(request.runID, &synccontract.ModeNotExecutableError{
				Mode:   synccontract.ModeChangeCapture,
				Reason: "change capture requires a connection-owned local warehouse destination",
			})
		}
		result, err := a.runWarehouseChangeCapture(ctx, request)
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
		if parked, handled, parkErr := a.parkRateLimitedRun(ctx, request, etlExecutionResult{}, err); handled {
			return parked, parkErr
		}
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
		if parked, handled, parkErr := a.parkRateLimitedRun(ctx, request, result, err); handled {
			return parked, parkErr
		}
		return a.failRun(request.runID, err)
	}
	return a.completeRun(request.runID, result)
}

// hasDeclaredOrderedPipeline reads only connector declarations. It does not
// infer safety from connector names, pools, or a generic Read/Write surface.
func hasDeclaredOrderedPipeline(source, destination connectors.Connector) bool {
	sourceDescriptor, sourceDeclared := connectors.SourceTransportDescriptorOf(source)
	destinationDescriptor, destinationDeclared := connectors.DestinationTransportDescriptorOf(destination)
	return sourceDeclared && destinationDeclared && sourceDescriptor.OrderedPipeline && destinationDescriptor.OrderedPipeline
}

// isOrderedArrowFullOverwriteCandidate is deliberately narrower than a
// declaration: the bounded producer/consumer implementation exists only for
// transformed Arrow full-overwrite runs. All other routes retain their prior
// serial default until they have a separately admitted implementation.
func isOrderedArrowFullOverwriteCandidate(request etlModeDispatchRequest) bool {
	return request.mode.ContractMode == synccontract.ModeFullOverwrite && request.stream.TransformPlan != "" && request.stream.TransformPlanHash != "" && hasDeclaredOrderedPipeline(request.source, request.destination)
}
