package synctransport

import (
	"context"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

func TestRegisterDeclaredTransportsRegistersTwoDefinitionOwnedPairs(t *testing.T) {
	first := newTestTransportPair("api", "database")
	second := newTestTransportPair("api", "database")
	second.source.meta.Name = "second_api_source"
	second.sourceExecutor.reference.ID = "second_api_source_executor"
	second.source.descriptor.Source.Executor = second.sourceExecutor.reference
	second.destination.meta.Name = "second_database_destination"
	second.destinationExecutor.reference.ID = "second_database_destination_executor"
	second.destination.descriptor.Destination.Executor = second.destinationExecutor.reference

	connectorsRegistry := connectors.NewEmptyRegistry()
	connectorsRegistry.Register(first.source)
	connectorsRegistry.Register(first.destination)
	connectorsRegistry.Register(second.source)
	connectorsRegistry.Register(second.destination)

	var sourceBuilds, destinationBuilds int
	factories := []DefinitionFactory{
		countingSourceFactory(first.sourceExecutor.reference, first.sourceExecutor, first.source, &sourceBuilds),
		countingDestinationFactory(first.destinationExecutor.reference, first.destinationExecutor, first.destination, &destinationBuilds),
		countingSourceFactory(second.sourceExecutor.reference, second.sourceExecutor, second.source, &sourceBuilds),
		countingDestinationFactory(second.destinationExecutor.reference, second.destinationExecutor, second.destination, &destinationBuilds),
	}
	verifier, err := NewDefinitionConformanceVerifier(factories)
	if err != nil {
		t.Fatal(err)
	}
	transports := NewRegistry(verifier)
	if err := RegisterDeclaredTransports(connectorsRegistry, transports, factories); err != nil {
		t.Fatalf("RegisterDeclaredTransports() = %v", err)
	}
	if sourceBuilds != 2 || destinationBuilds != 2 {
		t.Fatalf("factory builds source/destination = %d/%d, want two definition-owned registrations each", sourceBuilds, destinationBuilds)
	}
	if len(transports.sources) != 2 || len(transports.destinations) != 2 {
		t.Fatalf("registered source/destination executors = %d/%d, want two exact pairs", len(transports.sources), len(transports.destinations))
	}
	if _, err := transports.Preflight(PreflightRequest{
		Source: first.source, Destination: first.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
	}); err != nil {
		t.Fatalf("Preflight() = %v, want the first declared source/destination pair registered without App dispatch edits", err)
	}
}

func TestRegisterDeclaredTransportsRefusesBeforeAnyRegistration(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testTransportPair)
	}{
		{
			name: "unknown source executor",
			mutate: func(pair *testTransportPair) {
				pair.source.descriptor.Source.Executor.ID = "unknown_source_executor"
			},
		},
		{
			name: "destination change capture",
			mutate: func(pair *testTransportPair) {
				pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeChangeCapture}
				pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{
					Mode:     synccontract.ModeChangeCapture,
					Strategy: connectors.ApplyStrategyChangeApply,
					Action:   "stage_append",
				}}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair := newTestTransportPair("api", "database")
			tt.mutate(pair)
			connectorsRegistry := connectors.NewEmptyRegistry()
			connectorsRegistry.Register(pair.source)
			connectorsRegistry.Register(pair.destination)

			var sourceBuilds, destinationBuilds int
			factories := []DefinitionFactory{
				countingSourceFactory(pair.sourceExecutor.reference, pair.sourceExecutor, pair.source, &sourceBuilds),
				countingDestinationFactory(pair.destinationExecutor.reference, pair.destinationExecutor, pair.destination, &destinationBuilds),
			}
			verifier, err := NewDefinitionConformanceVerifier(factories)
			if err != nil {
				t.Fatal(err)
			}
			transports := NewRegistry(verifier)
			err = RegisterDeclaredTransports(connectorsRegistry, transports, factories)
			if err == nil {
				t.Fatal("RegisterDeclaredTransports() error = nil, want fail-closed declaration refusal")
			}
			if sourceBuilds != 0 || destinationBuilds != 0 || len(transports.sources) != 0 || len(transports.destinations) != 0 || pair.sourceExecutor.readCalls != 0 || pair.destinationExecutor.planCalls != 0 || pair.destinationExecutor.applyCalls != 0 {
				t.Fatalf("refused declaration side effects builds=%d/%d registered=%d/%d reads/plans/applies=%d/%d/%d, want zero", sourceBuilds, destinationBuilds, len(transports.sources), len(transports.destinations), pair.sourceExecutor.readCalls, pair.destinationExecutor.planCalls, pair.destinationExecutor.applyCalls)
			}
		})
	}
}

func TestDefinitionConformanceVerifierRefusesAlteredEvidenceBeforeSourceIO(t *testing.T) {
	pair := newTestTransportPair("api", "database")
	connectorsRegistry := connectors.NewEmptyRegistry()
	connectorsRegistry.Register(pair.source)
	connectorsRegistry.Register(pair.destination)
	factories := []DefinitionFactory{
		countingSourceFactory(pair.sourceExecutor.reference, pair.sourceExecutor, pair.source, new(int)),
		countingDestinationFactory(pair.destinationExecutor.reference, pair.destinationExecutor, pair.destination, new(int)),
	}
	verifier, err := NewDefinitionConformanceVerifier(factories)
	if err != nil {
		t.Fatal(err)
	}
	transports := NewRegistry(verifier)
	if err := RegisterDeclaredTransports(connectorsRegistry, transports, factories); err != nil {
		t.Fatal(err)
	}
	pair.source.descriptor.Source.Conformance.RunID = "self_reported_altered_evidence"

	_, err = transports.Preflight(PreflightRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
	})
	if err == nil || !strings.Contains(err.Error(), "conformance") {
		t.Fatalf("Preflight() error = %v, want external evidence refusal", err)
	}
	if pair.sourceExecutor.readCalls != 0 {
		t.Fatalf("source reads = %d, want zero before altered evidence refusal", pair.sourceExecutor.readCalls)
	}
}

func countingSourceFactory(reference connectors.TransportExecutorReference, executor SourceExecutor, expected connectors.Connector, builds *int) DefinitionFactory {
	descriptor, ok := connectors.SourceTransportDescriptorOf(expected)
	if !ok {
		panic("test source connector has no transport descriptor")
	}
	return DefinitionFactory{
		Reference:      reference,
		SourceEvidence: descriptor.Conformance,
		BuildSource: func(connector connectors.Connector) (SourceExecutor, error) {
			*builds++
			if connector != expected {
				return nil, context.Canceled
			}
			return executor, nil
		},
	}
}

func countingDestinationFactory(reference connectors.TransportExecutorReference, executor DestinationExecutor, expected connectors.Connector, builds *int) DefinitionFactory {
	descriptor, ok := connectors.DestinationTransportDescriptorOf(expected)
	if !ok {
		panic("test destination connector has no transport descriptor")
	}
	return DefinitionFactory{
		Reference:           reference,
		DestinationEvidence: descriptor.Conformance,
		BuildDestination: func(connector connectors.Connector) (DestinationExecutor, error) {
			*builds++
			if connector != expected {
				return nil, context.Canceled
			}
			return executor, nil
		},
	}
}
