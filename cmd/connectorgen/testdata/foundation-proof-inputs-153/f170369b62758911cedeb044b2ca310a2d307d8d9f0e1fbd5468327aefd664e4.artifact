package synctransport

import (
	"context"
	"fmt"
	"strings"

	"polymetrics.ai/internal/synccontract"
)

func (o *Orchestrator) ReadBackEmptyFullOverwrite(ctx context.Context, request EmptyPublicationReadBackRequest) error {
	if o == nil || o.registry == nil {
		return fmt.Errorf("transport orchestrator registry is required")
	}
	if err := request.validate(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	resolved, err := o.registry.Preflight(PreflightRequest{
		Source: request.Source, Destination: request.Destination, Stream: request.Stream,
		Mode: synccontract.ModeFullOverwrite, DestinationAction: request.DestinationAction,
	})
	if err != nil {
		return err
	}
	destination, ok := resolved.Destination.(EmptyPublicationReadBackDestination)
	if !ok {
		return fmt.Errorf("destination transport does not implement empty full-overwrite read-back")
	}
	readBackCtx, cancel := transportUnitContext(ctx, defaultTransportUnitDeadline)
	defer cancel()
	if err := destination.ReadBackEmptyFullOverwrite(readBackCtx, cloneEmptyPublicationReadBackRequest(request)); err != nil {
		return fmt.Errorf("read back empty full-overwrite publication: %w", tagTransportExecutionError(TransportExecutionOriginDestination, err))
	}
	return nil
}

func (r EmptyPublicationReadBackRequest) validate() error {
	if isNilInterface(r.Source) || isNilInterface(r.Destination) || strings.TrimSpace(r.Stream) == "" {
		return fmt.Errorf("empty full-overwrite read-back request is invalid")
	}
	if (strings.TrimSpace(r.TransformPlanJSON) == "") != (strings.TrimSpace(r.TransformPlanHash) == "") {
		return fmt.Errorf("empty full-overwrite read-back transform is invalid")
	}
	if err := r.Receipt.Validate(); err != nil {
		return fmt.Errorf("empty full-overwrite read-back receipt: %w", err)
	}
	if r.Receipt.Witness.Sink != r.Destination.Name() {
		return fmt.Errorf("empty full-overwrite read-back receipt sink %q does not match destination %q", r.Receipt.Witness.Sink, r.Destination.Name())
	}
	return nil
}

func cloneEmptyPublicationReadBackRequest(request EmptyPublicationReadBackRequest) EmptyPublicationReadBackRequest {
	clone := request
	clone.Runtime = cloneRuntimeConfig(request.Runtime)
	clone.SourceRuntime = cloneRuntimeConfig(request.SourceRuntime)
	clone.Binding.PrimaryKey = append([]string(nil), request.Binding.PrimaryKey...)
	clone.Receipt = request.Receipt.Clone()
	return clone
}
