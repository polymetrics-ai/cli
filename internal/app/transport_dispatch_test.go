package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	statestore "polymetrics.ai/internal/state"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

func TestRunETLCanonicalFullAppendUsesRegisteredTransports(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	sourceRef := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeAPI, ID: "fake_api_source"}
	destinationRef := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "fake_database_destination"}
	sourceEvidence := connectors.ConformanceEvidenceReference{Suite: "external_transport_test", RunID: "source_run"}
	destinationEvidence := connectors.ConformanceEvidenceReference{Suite: "external_transport_test", RunID: "destination_run"}
	sourceRecord := connectors.Record{"id": "1", "provider": "untouched"}
	source := &appTransportConnector{
		meta: connectors.Metadata{Name: "canonical_api_source", DisplayName: "Canonical API Source", IntegrationType: "api", Capabilities: connectors.Capabilities{Read: true, Catalog: true}},
		descriptor: &connectors.SyncTransportDescriptor{Source: &connectors.SourceTransportDescriptor{
			Executor: sourceRef, EligibleStreams: []string{"records"}, Modes: []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery: appTransportDelivery(), Conformance: sourceEvidence,
		}},
	}
	destination := &appTransportConnector{
		meta: connectors.Metadata{Name: "canonical_database_destination", DisplayName: "Canonical Database Destination", IntegrationType: "database", Capabilities: connectors.Capabilities{Write: true}},
		descriptor: &connectors.SyncTransportDescriptor{Destination: &connectors.DestinationTransportDescriptor{
			Executor: destinationRef, EligibleActions: []string{"stage_append"}, Modes: []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery: appTransportDelivery(), Conformance: destinationEvidence,
			Acknowledgement: connectors.TransportAcknowledgementDurableWarehouse,
			ApplyStrategies: []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullAppend, Strategy: connectors.ApplyStrategyAppend, Action: "stage_append"}},
		}},
	}
	registry := connectors.NewEmptyRegistry()
	registry.Register(source)
	registry.Register(destination)
	a.registry = registry

	sourceCredential, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "destination", Connector: destination.Name()}); err != nil {
		t.Fatal(err)
	}
	candidate := appTransportCheckpoint(source, sourceCredential, "records")
	sourceExecutor := &appTransportSourceExecutor{reference: sourceRef, page: synctransport.SourcePage{
		Records:             []connectors.Record{sourceRecord},
		CandidateCheckpoint: candidate,
	}}
	destinationExecutor := &appTransportDestinationExecutor{reference: destinationRef, sink: destination.Name()}
	verifier := appTransportVerifier{accepted: map[appTransportConformanceKey]struct{}{
		{role: connectors.TransportRoleSource, reference: sourceRef, evidence: sourceEvidence}:                {},
		{role: connectors.TransportRoleDestination, reference: destinationRef, evidence: destinationEvidence}: {},
	}}
	a.transports = synctransport.NewRegistry(verifier)
	if err := a.transports.RegisterSource(sourceExecutor); err != nil {
		t.Fatal(err)
	}
	if err := a.transports.RegisterDestination(destinationExecutor); err != nil {
		t.Fatal(err)
	}
	stage := &appTransportStage{}
	a.transportStage = stage

	if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "canonical_source_to_destination",
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: destination.Name(), Credential: "destination"},
		Streams: map[string]StreamConfig{
			"records": {SyncMode: string(synccontract.ModeFullAppend), DestinationTable: "records"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	run, err := a.RunETL(ctx, RunETLRequest{Connection: "canonical_source_to_destination", Stream: "records", BatchSize: 1})
	if err != nil {
		t.Fatalf("RunETL() = %v, want canonical mode dispatched through registered transports", err)
	}
	if sourceExecutor.readCalls != 1 || stage.calls != 1 || destinationExecutor.planCalls != 1 || destinationExecutor.applyCalls != 1 {
		t.Fatalf("transport calls source=%d stage=%d plan=%d apply=%d, want one each", sourceExecutor.readCalls, stage.calls, destinationExecutor.planCalls, destinationExecutor.applyCalls)
	}
	if source.legacyReadCalls != 0 || destination.legacyWriteCalls != 0 {
		t.Fatalf("legacy calls read=%d write=%d, want closed transport dispatch only", source.legacyReadCalls, destination.legacyWriteCalls)
	}
	if destinationExecutor.plan.ApplyStrategy.Action != "stage_append" || destinationExecutor.plan.ApplyStrategy.Strategy != connectors.ApplyStrategyAppend {
		t.Fatalf("destination plan strategy = %+v, want descriptor-declared stage_append/append", destinationExecutor.plan.ApplyStrategy)
	}
	if _, found := sourceRecord["_polymetrics_run_id"]; found {
		t.Fatalf("source provider record was mutated: %#v", sourceRecord)
	}
	if stage.lastPage.Records[0]["provider"] != "untouched" {
		t.Fatalf("stage provider record = %#v, want untouched source payload", stage.lastPage.Records[0])
	}
	if state := a.state.StreamStates[streamStateKey("canonical_source_to_destination", "records")]; state.Checkpoint == nil || state.Checkpoint.CommittedAt == nil {
		t.Fatalf("run did not persist a durable checkpoint: %#v", state)
	}
	if run.RecordsRead != 1 || run.RecordsLoaded != 1 || run.BatchCount != 1 {
		t.Fatalf("run counts = %+v, want one staged/apply page", run)
	}
}

var (
	errTransportSourceAfterAcknowledgement = errors.New("source failed after acknowledged page")
	errTransportFinalStateSave             = errors.New("final transport state save failed")
	errTransportCheckpointState            = errors.New("checkpoint state directory sync failed")
)

func TestRunETLTransportPersistsActiveCheckpointBeforeSourceFailureForAllModes(t *testing.T) {
	for _, contractMode := range synccontract.AllModes() {
		t.Run(string(contractMode), func(t *testing.T) {
			fixture := setupAppTransportFixture(t, contractMode)
			stateKey := streamStateKey(fixture.connection, "records")
			previous := fixture.sourceExecutor.page.CandidateCheckpoint.Clone()
			previous.Position.Primary = synccontract.OpaqueToken("previous")
			previous.Position.TieBreaker = synccontract.OpaqueToken("previous")
			previousCommittedAt := previous.ObservedAt.Add(time.Second)
			previous.CommittedAt = &previousCommittedAt
			fixture.sourceExecutor.page.CandidateCheckpoint.Position.Primary = synccontract.OpaqueToken("current")
			fixture.sourceExecutor.page.CandidateCheckpoint.Position.TieBreaker = synccontract.OpaqueToken("current")
			unrelatedUpdatedAt := time.Unix(1, 0).UTC()

			if _, err := fixture.app.updateState(func(current state) (state, error) {
				if current.StreamStates == nil {
					current.StreamStates = map[string]StreamState{}
				}
				current.StreamStates[stateKey] = StreamState{
					Connection:          fixture.connection,
					Stream:              "records",
					Checkpoint:          &previous,
					GenerationID:        9,
					LastSuccessfulRunID: "previous_run",
					RecordsLoaded:       99,
					UpdatedAt:           unrelatedUpdatedAt,
				}
				current.StreamStates["unrelated:records"] = StreamState{
					Connection:          "unrelated",
					Stream:              "records",
					GenerationID:        17,
					LastSuccessfulRunID: "unrelated_run",
					RecordsLoaded:       7,
					UpdatedAt:           unrelatedUpdatedAt,
				}
				return current, nil
			}); err != nil {
				t.Fatal(err)
			}
			fixture.sourceExecutor.errAfterPage = errTransportSourceAfterAcknowledgement

			run, err := fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
			if !errors.Is(err, errTransportSourceAfterAcknowledgement) {
				t.Fatalf("RunETL() error = %v, want source failure after acknowledgement", err)
			}
			if run.Status != "failed" {
				t.Fatalf("run status = %q, want failed after source error", run.Status)
			}

			wantGeneration := int64(9)
			if parsed, err := ParseSyncMode(string(contractMode)); err != nil {
				t.Fatal(err)
			} else if parsed.IsOverwrite() {
				wantGeneration++
			}
			assertInterimTransportState(t, fixture.app, stateKey, "current", wantGeneration)
			unrelated := fixture.app.state.StreamStates["unrelated:records"]
			if unrelated.GenerationID != 17 || unrelated.LastSuccessfulRunID != "unrelated_run" || unrelated.RecordsLoaded != 7 || !unrelated.UpdatedAt.Equal(unrelatedUpdatedAt) {
				t.Fatalf("unrelated stream state changed during checkpoint commit: %#v", unrelated)
			}

			reopened, err := Open(fixture.app.root)
			if err != nil {
				t.Fatal(err)
			}
			assertInterimTransportState(t, reopened, stateKey, "current", wantGeneration)
		})
	}
}

func TestRunETLTransportCommitsAcknowledgedPageBeforeCancellation(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	ctx, cancel := context.WithCancel(context.Background())
	fixture.destinationExecutor.afterApply = cancel

	run, err := fixture.app.RunETL(ctx, RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunETL() error = %v, want cancellation", err)
	}
	if run.Status != "failed" {
		t.Fatalf("run status = %q, want failed after cancellation", run.Status)
	}
	assertInterimTransportState(t, fixture.app, streamStateKey(fixture.connection, "records"), "1", 1)
}

func TestRunETLTransportRetainsInterimCheckpointWhenFinalStateSaveFails(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	fixture.app.store.Locker = &appTransportFailAtLockLocker{failAt: 3, err: errTransportFinalStateSave}

	_, err := fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
	if !errors.Is(err, errTransportFinalStateSave) {
		t.Fatalf("RunETL() error = %v, want final state save failure", err)
	}
	stateKey := streamStateKey(fixture.connection, "records")
	assertInterimTransportState(t, fixture.app, stateKey, "1", 1)
	if len(fixture.app.state.Runs) != 1 || fixture.app.state.Runs[0].Status != "running" || !fixture.app.state.Runs[0].CompletedAt.IsZero() {
		t.Fatalf("run after final state save failure = %#v, want still running without completion", fixture.app.state.Runs)
	}

	reopened, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	assertInterimTransportState(t, reopened, stateKey, "1", 1)
}

func TestRunETLTransportTreatsIndeterminateCheckpointPersistenceAsFailure(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	fixture.destinationExecutor.afterApply = func() {
		fixture.app.store.SyncDirectory = func(string) error {
			return errTransportCheckpointState
		}
	}

	_, err := fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
	var outcome *statestore.CommitOutcomeError
	if !errors.As(err, &outcome) || outcome.Outcome != statestore.CommitOutcomeIndeterminate {
		t.Fatalf("RunETL() error = %T %v, want indeterminate checkpoint persistence", err, err)
	}
	stateKey := streamStateKey(fixture.connection, "records")
	assertInterimTransportState(t, fixture.app, stateKey, "1", 1)
	if len(fixture.app.state.Runs) != 1 || fixture.app.state.Runs[0].Status == "completed" {
		t.Fatalf("run after indeterminate checkpoint persistence = %#v, must not complete", fixture.app.state.Runs)
	}

	reopened, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	assertInterimTransportState(t, reopened, stateKey, "1", 1)
}

func TestRunETLTransportPreflightRejectsMissingExecutorBeforeSourceRead(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	sourceRef := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeAPI, ID: "unregistered_source"}
	destinationRef := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "registered_destination"}
	source := &appTransportConnector{
		meta: connectors.Metadata{Name: "unregistered_api_source", IntegrationType: "api", Capabilities: connectors.Capabilities{Read: true, Catalog: true}},
		descriptor: &connectors.SyncTransportDescriptor{Source: &connectors.SourceTransportDescriptor{
			Executor: sourceRef, EligibleStreams: []string{"records"}, Modes: []synccontract.Mode{synccontract.ModeFullAppend}, Delivery: appTransportDelivery(),
			Conformance: connectors.ConformanceEvidenceReference{Suite: "external_transport_test", RunID: "source_run"},
		}},
	}
	destination := &appTransportConnector{
		meta: connectors.Metadata{Name: "registered_database_destination", IntegrationType: "database", Capabilities: connectors.Capabilities{Write: true}},
		descriptor: &connectors.SyncTransportDescriptor{Destination: &connectors.DestinationTransportDescriptor{
			Executor: destinationRef, EligibleActions: []string{"stage_append"}, Modes: []synccontract.Mode{synccontract.ModeFullAppend}, Delivery: appTransportDelivery(),
			Conformance: connectors.ConformanceEvidenceReference{Suite: "external_transport_test", RunID: "destination_run"}, Acknowledgement: connectors.TransportAcknowledgementDurableWarehouse,
			ApplyStrategies: []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullAppend, Strategy: connectors.ApplyStrategyAppend, Action: "stage_append"}},
		}},
	}
	registry := connectors.NewEmptyRegistry()
	registry.Register(source)
	registry.Register(destination)
	a.registry = registry
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "destination", Connector: destination.Name()}); err != nil {
		t.Fatal(err)
	}
	a.transports = synctransport.NewRegistry(appTransportVerifier{})
	if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name: "missing_executor_connection", Source: EndpointConfig{Connector: source.Name(), Credential: "source"}, Destination: EndpointConfig{Connector: destination.Name(), Credential: "destination"},
		Streams: map[string]StreamConfig{"records": {SyncMode: string(synccontract.ModeFullAppend)}},
	}); err != nil {
		t.Fatal(err)
	}

	_, err = a.RunETL(ctx, RunETLRequest{Connection: "missing_executor_connection", Stream: "records", BatchSize: 1})
	if err == nil || !strings.Contains(err.Error(), "source transport executor") {
		t.Fatalf("RunETL() error = %v, want missing source executor preflight failure", err)
	}
	if source.legacyReadCalls != 0 {
		t.Fatalf("legacy source Read calls = %d, want preflight rejection before source read", source.legacyReadCalls)
	}
}

func TestHasDeclaredSyncTransportRoutesInvalidDescriptorToPreflight(t *testing.T) {
	source := &appTransportConnector{
		meta:       connectors.Metadata{Name: "invalid_source", IntegrationType: "api"},
		descriptor: &connectors.SyncTransportDescriptor{},
	}
	destination := &appTransportConnector{meta: connectors.Metadata{Name: "destination", IntegrationType: "database"}}
	if !hasDeclaredSyncTransport(source, destination) {
		t.Fatal("empty authored transport descriptor was treated as absent instead of being routed to preflight")
	}
}

type appTransportConnector struct {
	meta             connectors.Metadata
	descriptor       *connectors.SyncTransportDescriptor
	legacyReadCalls  int
	legacyWriteCalls int
}

func (c *appTransportConnector) Name() string                  { return c.meta.Name }
func (c *appTransportConnector) Metadata() connectors.Metadata { return c.meta }
func (c *appTransportConnector) Definition() connectors.Definition {
	return connectors.Definition{Name: c.meta.Name, DisplayName: c.meta.DisplayName, IntegrationType: c.meta.IntegrationType, Capabilities: c.meta.Capabilities, SyncTransport: c.descriptor}
}
func (*appTransportConnector) Check(context.Context, connectors.RuntimeConfig) error { return nil }
func (c *appTransportConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{{Name: "records", PrimaryKey: []string{"id"}}}}, nil
}
func (c *appTransportConnector) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	c.legacyReadCalls++
	return errors.New("legacy Read must not be called by closed transport dispatch")
}
func (c *appTransportConnector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	c.legacyWriteCalls++
	return connectors.WriteResult{}, errors.New("legacy Write must not be called by closed transport dispatch")
}

type appTransportSourceExecutor struct {
	reference    connectors.TransportExecutorReference
	page         synctransport.SourcePage
	readCalls    int
	errAfterPage error
}

func (e *appTransportSourceExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return e.reference
}
func (e *appTransportSourceExecutor) ReadTransport(ctx context.Context, _ synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	e.readCalls++
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := emit(e.page); err != nil {
		return err
	}
	return e.errAfterPage
}

type appTransportDestinationExecutor struct {
	reference  connectors.TransportExecutorReference
	sink       string
	plan       synctransport.DestinationPlanRequest
	planCalls  int
	applyCalls int
	afterApply func()
}

func (e *appTransportDestinationExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return e.reference
}
func (e *appTransportDestinationExecutor) PlanDestination(_ context.Context, request synctransport.DestinationPlanRequest) (synctransport.DestinationPlan, error) {
	e.planCalls++
	e.plan = request
	return synctransport.DestinationPlan{ApplyStrategy: request.ApplyStrategy}, nil
}
func (e *appTransportDestinationExecutor) ApplyDestination(_ context.Context, _ synctransport.DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	e.applyCalls++
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement(e.sink, time.Now().UTC())
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if e.afterApply != nil {
		e.afterApply()
	}
	return acknowledgement, nil
}

type appTransportStage struct {
	calls    int
	lastPage synctransport.SourcePage
}

func (s *appTransportStage) Stage(_ context.Context, request synctransport.WarehouseStageRequest) (synctransport.WarehouseWorkset, error) {
	s.calls++
	s.lastPage = request.Page
	return synctransport.WarehouseWorkset{ID: fmt.Sprintf("stage-%d", s.calls), Records: request.Page.Records, Tombstones: request.Page.Tombstones, CandidateCheckpoint: request.Page.CandidateCheckpoint}, nil
}

type appTransportConformanceKey struct {
	role      connectors.TransportRole
	reference connectors.TransportExecutorReference
	evidence  connectors.ConformanceEvidenceReference
}

type appTransportVerifier struct {
	accepted map[appTransportConformanceKey]struct{}
}

func (v appTransportVerifier) VerifyTransportConformance(request synctransport.ConformanceVerification) error {
	if _, ok := v.accepted[appTransportConformanceKey{role: request.Role, reference: request.Executor, evidence: request.Evidence}]; !ok {
		return fmt.Errorf("external conformance verification has no accepted result for %s", request.Executor.ID)
	}
	return nil
}

func appTransportDelivery() connectors.DeliveryGuarantees {
	return connectors.DeliveryGuarantees{Idempotency: connectors.DeliveryIdempotencyKeyed, Ordering: connectors.DeliveryOrderingSource, Deletes: connectors.DeliveryDeletesTombstone}
}

type appTransportFixture struct {
	app                 *App
	connection          string
	sourceExecutor      *appTransportSourceExecutor
	destinationExecutor *appTransportDestinationExecutor
}

func setupAppTransportFixture(t *testing.T, mode synccontract.Mode) appTransportFixture {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	sourceRef := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeAPI, ID: "fixture_api_source"}
	destinationRef := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "fixture_database_destination"}
	sourceEvidence := connectors.ConformanceEvidenceReference{Suite: "external_transport_test", RunID: "fixture_source_run"}
	destinationEvidence := connectors.ConformanceEvidenceReference{Suite: "external_transport_test", RunID: "fixture_destination_run"}
	actions, strategies := appTransportStrategies()
	source := &appTransportConnector{
		meta: connectors.Metadata{Name: "fixture_api_source", DisplayName: "Fixture API Source", IntegrationType: "api", Capabilities: connectors.Capabilities{Read: true, Catalog: true}},
		descriptor: &connectors.SyncTransportDescriptor{Source: &connectors.SourceTransportDescriptor{
			Executor: sourceRef, EligibleStreams: []string{"records"}, Modes: synccontract.AllModes(),
			Delivery: appTransportDelivery(), Conformance: sourceEvidence,
		}},
	}
	destination := &appTransportConnector{
		meta: connectors.Metadata{Name: "fixture_database_destination", DisplayName: "Fixture Database Destination", IntegrationType: "database", Capabilities: connectors.Capabilities{Write: true}},
		descriptor: &connectors.SyncTransportDescriptor{Destination: &connectors.DestinationTransportDescriptor{
			Executor: destinationRef, EligibleActions: actions, Modes: synccontract.AllModes(),
			Delivery: appTransportDelivery(), Conformance: destinationEvidence,
			Acknowledgement: connectors.TransportAcknowledgementDurableWarehouse,
			ApplyStrategies: strategies,
		}},
	}
	registry := connectors.NewEmptyRegistry()
	registry.Register(source)
	registry.Register(destination)
	a.registry = registry

	sourceCredential, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "destination", Connector: destination.Name()}); err != nil {
		t.Fatal(err)
	}
	sourceExecutor := &appTransportSourceExecutor{reference: sourceRef, page: synctransport.SourcePage{
		Records:             []connectors.Record{{"id": "1", "provider": "untouched"}},
		CandidateCheckpoint: appTransportCheckpoint(source, sourceCredential, "records"),
	}}
	destinationExecutor := &appTransportDestinationExecutor{reference: destinationRef, sink: destination.Name()}
	verifier := appTransportVerifier{accepted: map[appTransportConformanceKey]struct{}{
		{role: connectors.TransportRoleSource, reference: sourceRef, evidence: sourceEvidence}:                {},
		{role: connectors.TransportRoleDestination, reference: destinationRef, evidence: destinationEvidence}: {},
	}}
	a.transports = synctransport.NewRegistry(verifier)
	if err := a.transports.RegisterSource(sourceExecutor); err != nil {
		t.Fatal(err)
	}
	if err := a.transports.RegisterDestination(destinationExecutor); err != nil {
		t.Fatal(err)
	}
	a.transportStage = &appTransportStage{}
	connection := "fixture_transport_" + string(mode)
	if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name:        connection,
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: destination.Name(), Credential: "destination"},
		Streams: map[string]StreamConfig{
			"records": {SyncMode: string(mode), CursorField: "updated_at", PrimaryKey: []string{"id"}, DestinationTable: "records"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	return appTransportFixture{app: a, connection: connection, sourceExecutor: sourceExecutor, destinationExecutor: destinationExecutor}
}

func appTransportStrategies() ([]string, []connectors.DestinationApplyStrategy) {
	modes := synccontract.AllModes()
	actions := make([]string, 0, len(modes))
	strategies := make([]connectors.DestinationApplyStrategy, 0, len(modes))
	for _, mode := range modes {
		action := "stage_" + string(mode)
		actions = append(actions, action)
		strategy := connectors.ApplyStrategyAppend
		switch mode {
		case synccontract.ModeFullOverwrite:
			strategy = connectors.ApplyStrategyReplace
		case synccontract.ModeIncrementalUpsert:
			strategy = connectors.ApplyStrategyMerge
		case synccontract.ModeIncrementalDedupe:
			strategy = connectors.ApplyStrategyDedupe
		case synccontract.ModeIncrementalDedupeHistory:
			strategy = connectors.ApplyStrategyDedupeHistory
		case synccontract.ModeChangeCapture:
			strategy = connectors.ApplyStrategyChangeApply
		}
		strategies = append(strategies, connectors.DestinationApplyStrategy{Mode: mode, Strategy: strategy, Action: action})
	}
	return actions, strategies
}

func assertInterimTransportState(t *testing.T, a *App, stateKey, wantPosition string, wantGeneration int64) {
	t.Helper()
	streamState, ok := a.state.StreamStates[stateKey]
	if !ok || streamState.Checkpoint == nil || streamState.Checkpoint.CommittedAt == nil {
		t.Fatalf("interim stream state = %#v, want durable checkpoint", streamState)
	}
	if got := string(streamState.Checkpoint.Position.Primary); got != wantPosition {
		t.Fatalf("interim checkpoint position = %q, want %q", got, wantPosition)
	}
	if streamState.GenerationID != wantGeneration {
		t.Fatalf("interim generation = %d, want %d", streamState.GenerationID, wantGeneration)
	}
	if streamState.LastSuccessfulRunID != "" || streamState.RecordsLoaded != 0 || streamState.UpdatedAt.IsZero() {
		t.Fatalf("interim stream state = %#v, must not claim successful completion", streamState)
	}
}

type appTransportFailAtLockLocker struct {
	calls  int
	failAt int
	err    error
}

func (l *appTransportFailAtLockLocker) Lock() (func() error, error) {
	l.calls++
	if l.calls == l.failAt {
		return nil, l.err
	}
	return func() error { return nil }, nil
}

func appTransportCheckpoint(source connectors.Connector, credential CredentialMeta, stream string) synccontract.CheckpointEnvelope {
	observed := time.Now().UTC().Add(-time.Minute)
	positionObserved := true
	expectation := streamResumeExpectation(source, credential, connectors.RuntimeConfig{}, stream)
	return synccontract.CheckpointEnvelope{
		StateVersion:     synccontract.StateVersion,
		Source:           expectation.Source,
		Mechanism:        "fake_transport",
		SnapshotBarrier:  &synccontract.SnapshotBarrier{Kind: "fake", Token: synccontract.OpaqueToken("barrier")},
		Position:         synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("1"), TieBreaker: synccontract.OpaqueToken("1")},
		PositionObserved: &positionObserved,
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: expectation.SourceGeneration,
		SchemaVersion:    "test-v1",
		ProtocolVersion:  "fake-v1",
		Dedupe:           synccontract.DedupeIdentity{Kind: "test", Value: synccontract.OpaqueToken("record-1")},
		DedupeWindow:     synccontract.DedupeWindow{Kind: "test", Start: synccontract.OpaqueToken("start"), End: synccontract.OpaqueToken("end")},
		ObservedAt:       observed,
	}
}
