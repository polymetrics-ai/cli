package synctransport

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

func TestPreflightRejectsClosedAdmissionFailuresBeforeSourceRead(t *testing.T) {
	for _, tt := range []struct {
		name    string
		prepare func(*testTransportPair, *Registry)
		want    string
	}{
		{
			name: "missing source executor",
			prepare: func(pair *testTransportPair, registry *Registry) {
				if err := registry.RegisterDestination(pair.destinationExecutor); err != nil {
					t.Fatal(err)
				}
			},
			want: "source transport executor",
		},
		{
			name: "integration type executor family mismatch",
			prepare: func(pair *testTransportPair, registry *Registry) {
				pair.source.meta.IntegrationType = "database"
				registerTransportPair(t, registry, pair)
			},
			want: "integration type",
		},
		{
			name: "unsupported canonical mode",
			prepare: func(pair *testTransportPair, registry *Registry) {
				pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
				registerTransportPair(t, registry, pair)
			},
			want: "does not support sync mode",
		},
		{
			name: "unsafe acknowledgement",
			prepare: func(pair *testTransportPair, registry *Registry) {
				pair.destination.descriptor.Destination.Acknowledgement = connectors.TransportAcknowledgementNone
				registerTransportPair(t, registry, pair)
			},
			want: "durable warehouse acknowledgement",
		},
		{
			name: "missing declared apply strategy",
			prepare: func(pair *testTransportPair, registry *Registry) {
				pair.destination.descriptor.Destination.ApplyStrategies = nil
				registerTransportPair(t, registry, pair)
			},
			want: "missing declared apply strategy",
		},
		{
			name: "external conformance unavailable",
			prepare: func(pair *testTransportPair, registry *Registry) {
				registerTransportPair(t, registry, pair)
				registry.verifier = rejectingConformanceVerifier{err: errors.New("external transport conformance verification is unavailable")}
			},
			want: "conformance",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			pair := newTestTransportPair("api", "database")
			registry := NewRegistry(pair.verifier)
			tt.prepare(pair, registry)

			_, err := registry.Preflight(PreflightRequest{
				Source:      pair.source,
				Destination: pair.destination,
				Stream:      "records",
				Mode:        synccontract.ModeFullAppend,
			})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Preflight() error = %v, want %q", err, tt.want)
			}
			if pair.sourceExecutor.readCalls != 0 {
				t.Fatalf("source ReadTransport calls = %d, want preflight rejection before read", pair.sourceExecutor.readCalls)
			}
		})
	}
}

func TestRegistryPreflightIsRaceSafeDuringRegistration(t *testing.T) {
	pair := newTestTransportPair("api", "database")
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)

	const readers = 32
	start := make(chan struct{})
	errs := make(chan error, readers+1)
	var group sync.WaitGroup
	for range readers {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			_, err := registry.Preflight(PreflightRequest{
				Source:      pair.source,
				Destination: pair.destination,
				Stream:      "records",
				Mode:        synccontract.ModeFullAppend,
			})
			errs <- err
		}()
	}
	group.Add(1)
	go func() {
		defer group.Done()
		<-start
		errs <- registry.RegisterSource(&testSourceExecutor{
			reference: connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeAPI, ID: "extra_api_source"},
		})
	}()
	close(start)
	group.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent registry operation = %v", err)
		}
	}
}

func TestOrchestratorDispatchesFourClosedPairingsWithoutPairBranches(t *testing.T) {
	pairs := []struct {
		name            string
		sourceType      string
		destinationType string
	}{
		{name: "api_to_api", sourceType: "api", destinationType: "api"},
		{name: "api_to_database", sourceType: "api", destinationType: "database"},
		{name: "database_to_api", sourceType: "database", destinationType: "api"},
		{name: "database_to_database", sourceType: "database", destinationType: "database"},
	}

	for _, tt := range pairs {
		t.Run(tt.name, func(t *testing.T) {
			pair := newTestTransportPair(tt.sourceType, tt.destinationType)
			registry := NewRegistry(pair.verifier)
			registerTransportPair(t, registry, pair)
			stage := &testWarehouseStage{}
			commits := 0

			result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
				Source:      pair.source,
				Destination: pair.destination,
				Stream:      "records",
				Mode:        synccontract.ModeFullAppend,
				BatchSize:   10,
				Stage:       stage,
				Commit: func(synccontract.CheckpointEnvelope) error {
					commits++
					return nil
				},
			})
			if err != nil {
				t.Fatalf("Run() = %v", err)
			}
			if pair.sourceExecutor.readCalls != 1 || stage.calls != 1 || pair.destinationExecutor.planCalls != 1 || pair.destinationExecutor.applyCalls != 1 {
				t.Fatalf("dispatch calls source=%d stage=%d plan=%d apply=%d, want 1 each", pair.sourceExecutor.readCalls, stage.calls, pair.destinationExecutor.planCalls, pair.destinationExecutor.applyCalls)
			}
			if commits != 1 || result.CommittedCheckpoint == nil {
				t.Fatalf("commits/checkpoint = %d/%+v, want one durable checkpoint", commits, result.CommittedCheckpoint)
			}
			if pair.destinationExecutor.lastPlan.ApplyStrategy.Action != "stage_append" {
				t.Fatalf("destination strategy = %+v, want descriptor-declared stage_append", pair.destinationExecutor.lastPlan.ApplyStrategy)
			}
			if got := stage.lastPage.Records[0]; !reflect.DeepEqual(got, connectors.Record{"id": "1", "provider": "untouched"}) {
				t.Fatalf("warehouse stage provider record = %#v, want unchanged provider payload", got)
			}
		})
	}
}

func TestOrchestratorCommitsOnlyAfterDurableAcknowledgement(t *testing.T) {
	pair := newTestTransportPair("api", "database")
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)
	stage := &testWarehouseStage{}
	commits := 0

	pair.destinationExecutor.acknowledgement = synccontract.DownstreamAcknowledgement{}
	pair.destinationExecutor.acknowledgementSet = true
	_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
		BatchSize: 10, Stage: stage,
		Commit: func(synccontract.CheckpointEnvelope) error {
			commits++
			return nil
		},
	})
	if !errors.Is(err, synccontract.ErrDownstreamAcknowledgementRequired) {
		t.Fatalf("Run() error = %v, want durable acknowledgement failure", err)
	}
	if commits != 0 {
		t.Fatalf("checkpoint commits = %d, want zero before durable acknowledgement", commits)
	}
}

func TestOrchestratorStagesDeepCopyOfProviderRecords(t *testing.T) {
	pair := newTestTransportPair("api", "database")
	originalNested := map[string]any{"provider": "untouched"}
	pair.sourceExecutor.pages[0].Records[0]["nested"] = originalNested
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)
	stage := &testWarehouseStage{mutateNestedPayload: true}

	_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
		BatchSize: 10, Stage: stage, Commit: func(synccontract.CheckpointEnvelope) error { return nil },
	})
	if err != nil {
		t.Fatalf("Run() = %v", err)
	}
	if _, mutated := originalNested["staged"]; mutated {
		t.Fatalf("stage mutated the source provider payload through an aliased nested value: %#v", originalNested)
	}
}

func TestOrchestratorStopsOnCancellationBetweenWarehouseStageAndApply(t *testing.T) {
	pair := newTestTransportPair("api", "database")
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)
	ctx, cancel := context.WithCancel(context.Background())
	stage := &testWarehouseStage{afterStage: cancel}
	commits := 0

	_, err := NewOrchestrator(registry).Run(ctx, RunRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
		BatchSize: 10, Stage: stage,
		Commit: func(synccontract.CheckpointEnvelope) error {
			commits++
			return nil
		},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want cancellation", err)
	}
	if pair.destinationExecutor.applyCalls != 0 || commits != 0 {
		t.Fatalf("apply/commits = %d/%d, want zero after cancellation", pair.destinationExecutor.applyCalls, commits)
	}
}

type testTransportPair struct {
	source              *testConnector
	destination         *testConnector
	sourceExecutor      *testSourceExecutor
	destinationExecutor *testDestinationExecutor
	verifier            acceptingConformanceVerifier
}

func newTestTransportPair(sourceType, destinationType string) *testTransportPair {
	sourceRef := connectors.TransportExecutorReference{Family: familyForType(sourceType), ID: "fake_" + sourceType + "_source"}
	destinationRef := connectors.TransportExecutorReference{Family: familyForType(destinationType), ID: "fake_" + destinationType + "_destination"}
	sourceEvidence := connectors.ConformanceEvidenceReference{Suite: "external_transport_test", RunID: "source_run"}
	destinationEvidence := connectors.ConformanceEvidenceReference{Suite: "external_transport_test", RunID: "destination_run"}

	source := &testConnector{
		meta: connectors.Metadata{Name: "source_" + sourceType, DisplayName: "Source", IntegrationType: sourceType, Capabilities: connectors.Capabilities{Read: true}},
		descriptor: &connectors.SyncTransportDescriptor{Source: &connectors.SourceTransportDescriptor{
			Executor: sourceRef, EligibleStreams: []string{"records"}, Modes: []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery: testDeliveryGuarantees(), Conformance: sourceEvidence,
		}},
	}
	destination := &testConnector{
		meta: connectors.Metadata{Name: "destination_" + destinationType, DisplayName: "Destination", IntegrationType: destinationType, Capabilities: connectors.Capabilities{Write: true}},
		descriptor: &connectors.SyncTransportDescriptor{Destination: &connectors.DestinationTransportDescriptor{
			Executor: destinationRef, EligibleActions: []string{"stage_append"}, Modes: []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery: testDeliveryGuarantees(), Conformance: destinationEvidence,
			Acknowledgement: connectors.TransportAcknowledgementDurableWarehouse,
			ApplyStrategies: []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullAppend, Strategy: connectors.ApplyStrategyAppend, Action: "stage_append"}},
		}},
	}
	candidate := testCheckpoint(source.meta.Name)
	sourceExecutor := &testSourceExecutor{reference: sourceRef, pages: []SourcePage{{
		Records:             []connectors.Record{{"id": "1", "provider": "untouched"}},
		CandidateCheckpoint: candidate,
	}}}
	destinationExecutor := &testDestinationExecutor{reference: destinationRef, sink: destination.meta.Name}
	verifier := acceptingConformanceVerifier{accepted: map[conformanceKey]struct{}{
		{role: connectors.TransportRoleSource, reference: sourceRef, evidence: sourceEvidence}:                {},
		{role: connectors.TransportRoleDestination, reference: destinationRef, evidence: destinationEvidence}: {},
	}}
	return &testTransportPair{source: source, destination: destination, sourceExecutor: sourceExecutor, destinationExecutor: destinationExecutor, verifier: verifier}
}

func registerTransportPair(t *testing.T, registry *Registry, pair *testTransportPair) {
	t.Helper()
	if err := registry.RegisterSource(pair.sourceExecutor); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterDestination(pair.destinationExecutor); err != nil {
		t.Fatal(err)
	}
}

func familyForType(integrationType string) connectors.TransportExecutorFamily {
	if integrationType == "database" {
		return connectors.TransportExecutorFamilyNativeDatabase
	}
	return connectors.TransportExecutorFamilyNativeAPI
}

func testDeliveryGuarantees() connectors.DeliveryGuarantees {
	return connectors.DeliveryGuarantees{Idempotency: connectors.DeliveryIdempotencyKeyed, Ordering: connectors.DeliveryOrderingSource, Deletes: connectors.DeliveryDeletesTombstone}
}

type testConnector struct {
	meta       connectors.Metadata
	descriptor *connectors.SyncTransportDescriptor
}

func (c *testConnector) Name() string                  { return c.meta.Name }
func (c *testConnector) Metadata() connectors.Metadata { return c.meta }
func (c *testConnector) Definition() connectors.Definition {
	return connectors.Definition{Name: c.meta.Name, DisplayName: c.meta.DisplayName, IntegrationType: c.meta.IntegrationType, SyncTransport: c.descriptor}
}
func (*testConnector) Check(context.Context, connectors.RuntimeConfig) error { return nil }
func (c *testConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{{Name: "records"}}}, nil
}
func (*testConnector) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	return errors.New("legacy connector Read must not be called by transport dispatch")
}
func (*testConnector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, errors.New("legacy connector Write must not be called by transport dispatch")
}

type testSourceExecutor struct {
	reference connectors.TransportExecutorReference
	pages     []SourcePage
	readCalls int
}

func (e *testSourceExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return e.reference
}
func (e *testSourceExecutor) ReadTransport(ctx context.Context, _ SourceRequest, emit func(SourcePage) error) error {
	e.readCalls++
	for _, page := range e.pages {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(page); err != nil {
			return err
		}
	}
	return nil
}

type testDestinationExecutor struct {
	reference          connectors.TransportExecutorReference
	sink               string
	acknowledgement    synccontract.DownstreamAcknowledgement
	acknowledgementSet bool
	planCalls          int
	applyCalls         int
	lastPlan           DestinationPlanRequest
}

func (e *testDestinationExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return e.reference
}
func (e *testDestinationExecutor) PlanDestination(_ context.Context, request DestinationPlanRequest) (DestinationPlan, error) {
	e.planCalls++
	e.lastPlan = request
	return DestinationPlan{ApplyStrategy: request.ApplyStrategy}, nil
}
func (e *testDestinationExecutor) ApplyDestination(_ context.Context, _ DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	e.applyCalls++
	if e.acknowledgementSet {
		return e.acknowledgement, nil
	}
	return synccontract.NewDurableDownstreamAcknowledgement(e.sink, time.Now().UTC())
}

type testWarehouseStage struct {
	calls               int
	lastPage            SourcePage
	afterStage          func()
	mutateNestedPayload bool
}

func (s *testWarehouseStage) Stage(_ context.Context, request WarehouseStageRequest) (WarehouseWorkset, error) {
	s.calls++
	s.lastPage = request.Page
	if s.mutateNestedPayload {
		nested, ok := request.Page.Records[0]["nested"].(map[string]any)
		if !ok {
			return WarehouseWorkset{}, fmt.Errorf("test stage expected nested provider map")
		}
		nested["staged"] = true
	}
	if s.afterStage != nil {
		s.afterStage()
	}
	return WarehouseWorkset{ID: fmt.Sprintf("stage-%d", s.calls), Records: request.Page.Records, Tombstones: request.Page.Tombstones, CandidateCheckpoint: request.Page.CandidateCheckpoint}, nil
}

type conformanceKey struct {
	role      connectors.TransportRole
	reference connectors.TransportExecutorReference
	evidence  connectors.ConformanceEvidenceReference
}

type acceptingConformanceVerifier struct {
	accepted map[conformanceKey]struct{}
}

func (v acceptingConformanceVerifier) VerifyTransportConformance(request ConformanceVerification) error {
	if _, ok := v.accepted[conformanceKey{role: request.Role, reference: request.Executor, evidence: request.Evidence}]; !ok {
		return fmt.Errorf("external conformance verifier has no accepted result for %s %s", request.Role, request.Executor.ID)
	}
	return nil
}

type rejectingConformanceVerifier struct{ err error }

func (v rejectingConformanceVerifier) VerifyTransportConformance(ConformanceVerification) error {
	return v.err
}

func testCheckpoint(engine string) synccontract.CheckpointEnvelope {
	observed := time.Now().UTC().Add(-time.Minute)
	positionObserved := true
	return synccontract.CheckpointEnvelope{
		StateVersion:     synccontract.StateVersion,
		Source:           synccontract.SourceIdentity{Engine: engine, AccountOrCluster: "test-account", ObjectScope: "records"},
		Mechanism:        "fake_transport",
		SnapshotBarrier:  &synccontract.SnapshotBarrier{Kind: "fake", Token: synccontract.OpaqueToken("barrier")},
		Position:         synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("1"), TieBreaker: synccontract.OpaqueToken("1")},
		PositionObserved: &positionObserved,
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: synccontract.OpaqueToken("generation"),
		SchemaVersion:    "test-v1",
		ProtocolVersion:  "fake-v1",
		Dedupe:           synccontract.DedupeIdentity{Kind: "test", Value: synccontract.OpaqueToken("record-1")},
		DedupeWindow:     synccontract.DedupeWindow{Kind: "test", Start: synccontract.OpaqueToken("start"), End: synccontract.OpaqueToken("end")},
		ObservedAt:       observed,
	}
}
