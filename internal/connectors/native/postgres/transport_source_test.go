package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

func TestPostgresDefinitionDeclaresBoundedSnapshotTransportSource(t *testing.T) {
	connector := New()
	descriptor, ok := connectors.SourceTransportDescriptorOf(connector)
	if !ok {
		t.Fatal("PostgreSQL definition has no declared source transport")
	}
	wantReference := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_bounded_snapshot"}
	if descriptor.Executor != wantReference {
		t.Fatalf("PostgreSQL source executor = %#v, want %#v", descriptor.Executor, wantReference)
	}
	if got, want := descriptor.EligibleStreams, []string{"snapshot"}; !sameStrings(got, want) {
		t.Fatalf("PostgreSQL source streams = %#v, want %#v", got, want)
	}
	if got, want := descriptor.Modes, []synccontract.Mode{synccontract.ModeFullAppend, synccontract.ModeFullOverwrite}; !sameModes(got, want) {
		t.Fatalf("PostgreSQL source modes = %#v, want %#v", got, want)
	}
}

func TestPostgresTransportRegistryPreflightRefusesBeforeSourceIO(t *testing.T) {
	tests := []struct {
		name       string
		source     *preflightSpyConnector
		register   bool
		wantErr    string
		wantSource bool
	}{
		{
			name:    "missing descriptor",
			source:  newPreflightSpyConnector(nil),
			register: false,
			wantErr: "has no declared source transport",
		},
		{
			name: "wrong connector family",
			source: newPreflightSpyConnector(&connectors.SourceTransportDescriptor{
				Executor: connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeAPI, ID: "postgres_bounded_snapshot"},
				EligibleStreams: []string{"snapshot"},
				Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
				Delivery: connectors.DeliveryGuarantees{
					Idempotency: connectors.DeliveryIdempotencyAtLeastOnce,
					Ordering:    connectors.DeliveryOrderingSource,
					Deletes:     connectors.DeliveryDeletesUnavailable,
				},
				Conformance: connectors.ConformanceEvidenceReference{Suite: "postgres_snapshot", RunID: "bounded_full_v1"},
			}),
			register: true,
			wantErr: "incompatible with transport executor family",
		},
		{
			name:     "unregistered declared executor",
			source:   newPreflightSpyConnector(postgresTransportTestSourceDescriptor()),
			wantErr:  "is not registered",
			register: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry := synctransport.NewRegistry(postgresTransportTestVerifier{})
			if test.register && test.source.descriptor != nil {
				if err := registry.RegisterSource(&preflightSpySourceExecutor{reference: test.source.descriptor.Executor}); err != nil {
					t.Fatalf("RegisterSource() error = %v", err)
				}
			}
			destination := newPreflightDestinationConnector()
			if err := registry.RegisterDestination(&preflightSpyDestinationExecutor{reference: destination.destination.Executor}); err != nil {
				t.Fatalf("RegisterDestination() error = %v", err)
			}

			_, err := registry.Preflight(synctransport.PreflightRequest{
				Source:      test.source,
				Destination: destination,
				Stream:      "snapshot",
				Mode:        synccontract.ModeFullAppend,
			})
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("Preflight() error = %v, want substring %q", err, test.wantErr)
			}
			if test.source.ioCalls != 0 {
				t.Fatalf("Preflight() triggered source I/O %d times", test.source.ioCalls)
			}
		})
	}
}

type preflightSpyConnector struct {
	descriptor *connectors.SourceTransportDescriptor
	ioCalls    int
}

func newPreflightSpyConnector(descriptor *connectors.SourceTransportDescriptor) *preflightSpyConnector {
	return &preflightSpyConnector{descriptor: descriptor}
}

func postgresTransportTestSourceDescriptor() *connectors.SourceTransportDescriptor {
	return &connectors.SourceTransportDescriptor{
		Executor:        connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_bounded_snapshot"},
		EligibleStreams: []string{"snapshot"},
		Modes:           []synccontract.Mode{synccontract.ModeFullAppend, synccontract.ModeFullOverwrite},
		Delivery: connectors.DeliveryGuarantees{
			Idempotency: connectors.DeliveryIdempotencyAtLeastOnce,
			Ordering:    connectors.DeliveryOrderingSource,
			Deletes:     connectors.DeliveryDeletesUnavailable,
		},
		Conformance: connectors.ConformanceEvidenceReference{Suite: "postgres_snapshot", RunID: "bounded_full_v1"},
	}
}

func (c *preflightSpyConnector) Name() string { return "postgres" }

func (*preflightSpyConnector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "postgres", IntegrationType: "database"}
}

func (c *preflightSpyConnector) Check(context.Context, connectors.RuntimeConfig) error {
	c.ioCalls++
	return errors.New("source I/O is forbidden in transport preflight")
}

func (c *preflightSpyConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	c.ioCalls++
	return connectors.Catalog{}, errors.New("source I/O is forbidden in transport preflight")
}

func (c *preflightSpyConnector) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	c.ioCalls++
	return errors.New("source I/O is forbidden in transport preflight")
}

func (c *preflightSpyConnector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	c.ioCalls++
	return connectors.WriteResult{}, errors.New("source I/O is forbidden in transport preflight")
}

func (c *preflightSpyConnector) Definition() connectors.Definition {
	definition := connectors.Definition{
		Name:            c.Name(),
		DisplayName:     "PostgreSQL preflight probe",
		IntegrationType: "database",
		SyncTransport:   &connectors.SyncTransportDescriptor{Source: c.descriptor},
	}
	if c.descriptor == nil {
		definition.SyncTransport = nil
	}
	return definition
}

type preflightDestinationConnector struct {
	destination connectors.DestinationTransportDescriptor
}

func newPreflightDestinationConnector() *preflightDestinationConnector {
	return &preflightDestinationConnector{destination: connectors.DestinationTransportDescriptor{
		Executor:        connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_transport_test_destination"},
		EligibleActions: []string{"apply_snapshot"},
		Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
		Delivery: connectors.DeliveryGuarantees{
			Idempotency: connectors.DeliveryIdempotencyAtLeastOnce,
			Ordering:    connectors.DeliveryOrderingSource,
			Deletes:     connectors.DeliveryDeletesUnavailable,
		},
		Conformance:     connectors.ConformanceEvidenceReference{Suite: "postgres_transport_test", RunID: "destination"},
		Acknowledgement: connectors.TransportAcknowledgementDurableWarehouse,
		ApplyStrategies: []connectors.DestinationApplyStrategy{{
			Mode:     synccontract.ModeFullAppend,
			Strategy: connectors.ApplyStrategyAppend,
			Action:   "apply_snapshot",
		}},
	}}
}

func (*preflightDestinationConnector) Name() string { return "postgres-destination" }

func (*preflightDestinationConnector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "postgres-destination", IntegrationType: "database"}
}

func (*preflightDestinationConnector) Check(context.Context, connectors.RuntimeConfig) error { return nil }

func (*preflightDestinationConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{}, nil
}

func (*preflightDestinationConnector) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	return nil
}

func (*preflightDestinationConnector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, nil
}

func (c *preflightDestinationConnector) Definition() connectors.Definition {
	return connectors.Definition{
		Name:            c.Name(),
		DisplayName:     "PostgreSQL destination preflight probe",
		IntegrationType: "database",
		SyncTransport:   &connectors.SyncTransportDescriptor{Destination: &c.destination},
	}
}

type preflightSpySourceExecutor struct {
	reference connectors.TransportExecutorReference
}

func (e *preflightSpySourceExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return e.reference
}

func (*preflightSpySourceExecutor) ReadTransport(context.Context, synctransport.SourceRequest, func(synctransport.SourcePage) error) error {
	return errors.New("source executor must not run during preflight")
}

type preflightSpyDestinationExecutor struct {
	reference connectors.TransportExecutorReference
}

func (e *preflightSpyDestinationExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return e.reference
}

func (*preflightSpyDestinationExecutor) PlanDestination(context.Context, synctransport.DestinationPlanRequest) (synctransport.DestinationPlan, error) {
	return synctransport.DestinationPlan{}, errors.New("destination executor must not run during preflight")
}

func (*preflightSpyDestinationExecutor) ApplyDestination(context.Context, synctransport.DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	return synccontract.DownstreamAcknowledgement{}, errors.New("destination executor must not run during preflight")
}

func (*preflightSpyDestinationExecutor) ReadBackDestination(context.Context, synctransport.DestinationReadBackRequest) error {
	return errors.New("destination executor must not run during preflight")
}

type postgresTransportTestVerifier struct{}

func (postgresTransportTestVerifier) VerifyTransportConformance(synctransport.ConformanceVerification) error {
	return nil
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sameModes(left, right []synccontract.Mode) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
