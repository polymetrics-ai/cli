package app

import (
	"context"
	"errors"
	"testing"

	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

func TestRunETLRefusesExplicitPipelineDepthWithoutEndpointDeclarationBeforeRead(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)

	run, err := fixture.app.RunETL(context.Background(), RunETLRequest{
		Connection: fixture.connection, Stream: "records", BatchSize: 1, MaxInFlightBatches: 1,
	})
	var unsupported *synctransport.OrderedPipelineUnsupportedError
	if !errors.As(err, &unsupported) {
		t.Fatalf("RunETL() error = %T %v, want OrderedPipelineUnsupportedError", err, err)
	}
	if fixture.sourceExecutor.readCalls != 0 || fixture.destinationExecutor.planCalls != 0 || fixture.destinationExecutor.applyCalls != 0 {
		t.Fatalf("unsupported explicit pipeline reached read/plan/apply = %d/%d/%d, want 0/0/0", fixture.sourceExecutor.readCalls, fixture.destinationExecutor.planCalls, fixture.destinationExecutor.applyCalls)
	}
	if run.Status != "failed" {
		t.Fatalf("failed pre-I/O pipeline run status = %q, want failed", run.Status)
	}
}

func TestRunETLRefusesExplicitPipelineDepthWhenRouteCannotExecuteIt(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	fixture.source.descriptor.Source.OrderedPipeline = true
	fixture.destination.descriptor.Destination.OrderedPipeline = true

	run, err := fixture.app.RunETL(context.Background(), RunETLRequest{
		Connection: fixture.connection, Stream: "records", BatchSize: 1, MaxInFlightBatches: 1,
	})
	var unsupported *synctransport.OrderedPipelineUnsupportedError
	if !errors.As(err, &unsupported) {
		t.Fatalf("RunETL() error = %T %v, want OrderedPipelineUnsupportedError", err, err)
	}
	if fixture.sourceExecutor.readCalls != 0 || fixture.destinationExecutor.planCalls != 0 || fixture.destinationExecutor.applyCalls != 0 {
		t.Fatalf("ineligible explicit pipeline reached read/plan/apply = %d/%d/%d, want 0/0/0", fixture.sourceExecutor.readCalls, fixture.destinationExecutor.planCalls, fixture.destinationExecutor.applyCalls)
	}
	if run.Status != "failed" {
		t.Fatalf("failed pre-I/O pipeline run status = %q, want failed", run.Status)
	}
}
