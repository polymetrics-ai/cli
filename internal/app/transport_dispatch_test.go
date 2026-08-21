package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/coordination"
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

func TestRunETLTransportPersistsPartialPhaseMeasurementBeforeFailureCleanup(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	first := fixture.sourceExecutor.page
	second := fixture.sourceExecutor.page
	second.Records = []connectors.Record{{"id": "2", "provider": "second"}}
	second.CandidateCheckpoint.Position.Primary = synccontract.OpaqueToken("2")
	second.CandidateCheckpoint.Position.TieBreaker = synccontract.OpaqueToken("2")
	fixture.sourceExecutor.pages = []synctransport.SourcePage{first, second}
	fixture.destinationExecutor.failApplyAt = 2
	fixture.destinationExecutor.applyErr = errors.New("injected destination apply failure")

	run, err := fixture.app.RunETL(context.Background(), RunETLRequest{
		Connection: fixture.connection, Stream: "records", BatchSize: 1,
	})
	if !errors.Is(err, fixture.destinationExecutor.applyErr) {
		t.Fatalf("RunETL() error = %T %v, want injected destination failure", err, err)
	}
	if run.Status != "failed" || run.TransportPhaseMeasurement == nil {
		t.Fatalf("failed run = %#v, want persisted transport phase measurement", run)
	}
	measurement := run.TransportPhaseMeasurement
	if measurement.ExtractedRecords != 2 || measurement.WarehouseParquetRecords != 2 || measurement.PostgreSQLAppliedRecords != 1 {
		t.Fatalf("failed measurement counts = %#v, want extracted/warehouse/postgres = 2/2/1", measurement)
	}
	if measurement.ExtractElapsedNanos <= 0 || measurement.WarehouseElapsedNanos <= 0 || measurement.PostgreSQLElapsedNanos <= 0 {
		t.Fatalf("failed measurement durations = %#v, want all three phase durations before deferred caller cleanup", measurement)
	}
	reopened, reopenErr := Open(filepath.Dir(fixture.app.ProjectDir()))
	if reopenErr != nil {
		t.Fatalf("Open(reopened project) error = %v", reopenErr)
	}
	persisted, persistedErr := reopened.GetRun(run.ID)
	if persistedErr != nil || !reflect.DeepEqual(persisted.TransportPhaseMeasurement, measurement) {
		t.Fatalf("persisted failed run measurement = (%#v, %v), want %#v before project cleanup", persisted.TransportPhaseMeasurement, persistedErr, measurement)
	}
}

func TestRunETLTransportParksSourceRateLimitWithDestinationResults(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	t.Cleanup(fixture.app.rateParking.Close)
	fixture.source.rateLimitScope = "transport-source-rate-limit"
	wantOutput := json.RawMessage(`{"provider_receipt":"first"}`)
	fixture.destinationExecutor.output = wantOutput
	fixture.sourceExecutor.errAfterPage = &connsdk.RateLimitError{
		HTTPError: &connsdk.HTTPError{Status: 429, URL: "https://provider.invalid/records"},
		Source:    connsdk.RateLimitObservationSourceRetryAfter,
		HasReset:  true,
		ResetAt:   time.Now().UTC().Add(time.Hour),
	}

	run, err := fixture.app.RunETL(context.Background(), RunETLRequest{
		Connection: fixture.connection, Stream: "records", BatchSize: 1,
	})
	if !errors.Is(err, coordination.ErrRateLimitParked) {
		t.Fatalf("RunETL() error = %v, want rate-limit parking", err)
	}
	if run.Status != string(coordination.RateParkingOutcomeParkedRateLimit) || !reflect.DeepEqual(run.DestinationResults, []json.RawMessage{wantOutput}) {
		t.Fatalf("parked run = %#v, want persisted destination output", run)
	}
	persisted, err := fixture.app.GetRun(run.ID)
	if err != nil || !reflect.DeepEqual(persisted.DestinationResults, []json.RawMessage{wantOutput}) {
		t.Fatalf("persisted parked run = (%#v, %v), want destination output", persisted, err)
	}
}

func TestParkedFullAppendRateResumePreservesCheckpointAndBatchSize(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	t.Cleanup(fixture.app.rateParking.Close)
	fixture.source.rateLimitScope = "transport-source-rate-limit"
	first := fixture.sourceExecutor.page
	second := fixture.sourceExecutor.page
	second.Records = []connectors.Record{{"id": "2", "provider": "resumed"}}
	second.CandidateCheckpoint.Position.Primary = synccontract.OpaqueToken("2")
	second.CandidateCheckpoint.Position.TieBreaker = synccontract.OpaqueToken("2")
	rateLimited := &connsdk.RateLimitError{
		HTTPError: &connsdk.HTTPError{Status: http.StatusTooManyRequests, URL: "https://provider.invalid/records"},
		Source:    connsdk.RateLimitObservationSourceRetryAfter,
		HasReset:  true,
		ResetAt:   time.Now().UTC().Add(time.Hour),
	}
	var parkedCheckpoint *synccontract.CheckpointEnvelope
	fixture.sourceExecutor.read = func(_ context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
		switch fixture.sourceExecutor.readCalls {
		case 1:
			if request.Checkpoint != nil {
				return errors.New("initial full-append request unexpectedly used a checkpoint")
			}
			if request.BatchSize != 1 {
				return fmt.Errorf("initial batch size = %d, want 1", request.BatchSize)
			}
			if err := emit(first); err != nil {
				return err
			}
			return rateLimited
		case 2:
			if request.BatchSize != 1 {
				return fmt.Errorf("resumed batch size = %d, want 1", request.BatchSize)
			}
			if !transportCheckpointEqual(request.Checkpoint, parkedCheckpoint) {
				return errors.New("resumed full-append request did not use the committed checkpoint")
			}
			return emit(second)
		default:
			return fmt.Errorf("unexpected source read %d", fixture.sourceExecutor.readCalls)
		}
	}

	parkedRun, err := fixture.app.RunETL(context.Background(), RunETLRequest{
		Connection: fixture.connection, Stream: "records", BatchSize: 1,
	})
	if !errors.Is(err, coordination.ErrRateLimitParked) {
		t.Fatalf("initial RunETL() error = %v, want rate-limit parking", err)
	}
	if parkedRun.BatchSize != 1 {
		t.Fatalf("parked run batch size = %d, want 1", parkedRun.BatchSize)
	}
	checkpoint := fixture.app.state.StreamStates[streamStateKey(fixture.connection, "records")].Checkpoint.Clone()
	if checkpoint.CommittedAt == nil || !bytes.Equal(checkpoint.Position.Primary, first.CandidateCheckpoint.Position.Primary) {
		t.Fatalf("parked checkpoint = %#v, want committed first acknowledged position", checkpoint)
	}
	parkedCheckpoint = &checkpoint
	if err := fixture.app.resumeParkedRateLimitRun(context.Background(), coordination.ParkedRateLimitRun{RunID: parkedRun.ID, Checkpoint: checkpoint}); err != nil {
		t.Fatalf("resume parked run: %v", err)
	}
	if fixture.sourceExecutor.readCalls != 2 || fixture.destinationExecutor.applyCalls != 2 {
		t.Fatalf("resume calls source=%d destination=%d, want one acknowledged action before and after parking", fixture.sourceExecutor.readCalls, fixture.destinationExecutor.applyCalls)
	}
	original, found := fixture.app.runByID(parkedRun.ID)
	if !found || original.Status != "resumed" {
		t.Fatalf("original parked run = %#v, want resumed", original)
	}
}

func TestParkedFullAppendRateResumeRearmsLatestCheckpoint(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	fixture.app.rateParking.Close()
	wantOutput := json.RawMessage(`{"provider_receipt":"rearmed"}`)
	fixture.destinationExecutor.output = wantOutput
	now := time.Date(2026, time.August, 20, 12, 0, 0, 0, time.UTC)
	firstReset := now.Add(time.Minute)
	secondReset := now.Add(2 * time.Minute)
	scheduler := &appRateParkingTestScheduler{}
	parkingStore := coordination.NewMemoryRateParkingStore()
	fixture.app.rateParking = coordination.NewRateParkingCoordinator(coordination.RateParkingCoordinatorOptions{
		Store:     parkingStore,
		Scheduler: scheduler,
		Now:       func() time.Time { return now },
		Resume:    fixture.app.resumeParkedRateLimitRun,
	})
	if err := fixture.app.rateParking.Start(context.Background()); err != nil {
		t.Fatalf("start test rate parking: %v", err)
	}
	t.Cleanup(fixture.app.rateParking.Close)
	fixture.source.rateLimitScope = "transport-source-rate-limit"
	first := fixture.sourceExecutor.page
	second := fixture.sourceExecutor.page
	second.Records = []connectors.Record{{"id": "2", "provider": "second"}}
	second.CandidateCheckpoint.Position.Primary = synccontract.OpaqueToken("2")
	second.CandidateCheckpoint.Position.TieBreaker = synccontract.OpaqueToken("2")
	third := fixture.sourceExecutor.page
	third.Records = []connectors.Record{{"id": "3", "provider": "third"}}
	third.CandidateCheckpoint.Position.Primary = synccontract.OpaqueToken("3")
	third.CandidateCheckpoint.Position.TieBreaker = synccontract.OpaqueToken("3")
	rateOne := &connsdk.RateLimitError{
		HTTPError: &connsdk.HTTPError{Status: http.StatusTooManyRequests, URL: "https://provider.invalid/records"},
		Source:    connsdk.RateLimitObservationSourceRetryAfter,
		HasReset:  true,
		ResetAt:   firstReset,
	}
	rateTwo := &connsdk.RateLimitError{
		HTTPError: &connsdk.HTTPError{Status: http.StatusTooManyRequests, URL: "https://provider.invalid/records"},
		Source:    connsdk.RateLimitObservationSourceHeaders,
		HasReset:  true,
		ResetAt:   secondReset,
	}
	var firstCheckpoint, secondCheckpoint *synccontract.CheckpointEnvelope
	fixture.sourceExecutor.read = func(_ context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
		switch fixture.sourceExecutor.readCalls {
		case 1:
			if request.Checkpoint != nil {
				return errors.New("initial full-append request unexpectedly used a checkpoint")
			}
			if err := emit(first); err != nil {
				return err
			}
			return rateOne
		case 2:
			if !transportCheckpointEqual(request.Checkpoint, firstCheckpoint) {
				return fmt.Errorf("first resumed checkpoint = %#v, want %#v", request.Checkpoint, firstCheckpoint)
			}
			if err := emit(second); err != nil {
				return err
			}
			return rateTwo
		case 3:
			if !transportCheckpointEqual(request.Checkpoint, secondCheckpoint) {
				return fmt.Errorf("second resumed checkpoint = %#v, want %#v", request.Checkpoint, secondCheckpoint)
			}
			return emit(third)
		default:
			return fmt.Errorf("unexpected source read %d", fixture.sourceExecutor.readCalls)
		}
	}

	parked, err := fixture.app.RunETL(context.Background(), RunETLRequest{
		Connection: fixture.connection, Stream: "records", BatchSize: 1,
	})
	if !errors.Is(err, coordination.ErrRateLimitParked) {
		t.Fatalf("initial RunETL() error = %v, want rate-limit parking", err)
	}
	checkpoint := fixture.app.state.StreamStates[streamStateKey(fixture.connection, "records")].Checkpoint.Clone()
	firstCheckpoint = &checkpoint

	now = firstReset
	scheduler.RunThrough(now)
	checkpoint = fixture.app.state.StreamStates[streamStateKey(fixture.connection, "records")].Checkpoint.Clone()
	secondCheckpoint = &checkpoint
	records, err := parkingStore.List()
	if err != nil || len(records) != 1 || records[0].RunID != parked.ID {
		t.Fatalf("rearmed parking records = %#v, %v; want original run", records, err)
	}
	if !records[0].ResetAt.Equal(secondReset) || records[0].Reason != connsdk.RateLimitObservationSourceHeaders || !transportCheckpointEqual(&records[0].Checkpoint, secondCheckpoint) {
		t.Fatalf("rearmed parking record = %#v, want latest checkpoint and second reset", records[0])
	}
	if scheduler.Scheduled() != 1 {
		t.Fatalf("scheduled rearmed callbacks = %d, want 1", scheduler.Scheduled())
	}
	var rearmedAttempt Run
	parkedRuns := 0
	for _, run := range fixture.app.state.Runs {
		if run.Status == string(coordination.RateParkingOutcomeParkedRateLimit) {
			parkedRuns++
			if run.ID != parked.ID {
				t.Fatalf("parked retry attempt = %#v, want original run %q only", run, parked.ID)
			}
		}
		if run.ID != parked.ID {
			rearmedAttempt = run
		}
	}
	if parkedRuns != 1 {
		t.Fatalf("parked runs = %d, want original run only", parkedRuns)
	}
	if rearmedAttempt.ID == "" || rearmedAttempt.Status != "failed" || rearmedAttempt.Error != coordination.ErrRateLimitRearmed.Error() || rearmedAttempt.CompletedAt.IsZero() {
		t.Fatalf("rearmed retry attempt = %#v, want terminally recorded failure", rearmedAttempt)
	}
	if rearmedAttempt.RecordsRead != 1 || rearmedAttempt.RecordsLoaded != 1 || rearmedAttempt.BatchCount != 1 || !reflect.DeepEqual(rearmedAttempt.DestinationResults, []json.RawMessage{wantOutput}) {
		t.Fatalf("rearmed retry result = %#v, want retained acknowledged output", rearmedAttempt)
	}

	now = secondReset
	scheduler.RunThrough(now)
	if fixture.sourceExecutor.readCalls != 3 || fixture.destinationExecutor.applyCalls != 3 {
		t.Fatalf("source/apply calls = %d/%d, want one apply for each acknowledged page", fixture.sourceExecutor.readCalls, fixture.destinationExecutor.applyCalls)
	}
	if records, err := parkingStore.List(); err != nil || len(records) != 0 {
		t.Fatalf("parking records after second resume = %#v, %v; want none", records, err)
	}
	original, found := fixture.app.runByID(parked.ID)
	if !found || original.Status != "resumed" {
		t.Fatalf("original parked run = %#v, want resumed", original)
	}
}

func TestParkedFullAppendRateResumeReconcilesInterruptedRearmAttempt(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	fixture.app.rateParking.Close()
	now := time.Date(2026, time.August, 20, 13, 0, 0, 0, time.UTC)
	firstReset := now.Add(time.Minute)
	secondReset := now.Add(2 * time.Minute)
	scheduler := &appRateParkingTestScheduler{}
	parkingStore := coordination.NewMemoryRateParkingStore()
	fixture.app.rateParking = coordination.NewRateParkingCoordinator(coordination.RateParkingCoordinatorOptions{
		Store:     parkingStore,
		Scheduler: scheduler,
		Now:       func() time.Time { return now },
		Resume:    fixture.app.resumeParkedRateLimitRun,
	})
	if err := fixture.app.rateParking.Start(context.Background()); err != nil {
		t.Fatalf("start test rate parking: %v", err)
	}
	t.Cleanup(fixture.app.rateParking.Close)
	fixture.source.rateLimitScope = "transport-source-rate-limit"
	first := fixture.sourceExecutor.page
	completed := fixture.sourceExecutor.page
	completed.Records = []connectors.Record{{"id": "3", "provider": "completed"}}
	completed.CandidateCheckpoint.Position.Primary = synccontract.OpaqueToken("3")
	completed.CandidateCheckpoint.Position.TieBreaker = synccontract.OpaqueToken("3")
	rateOne := &connsdk.RateLimitError{
		HTTPError: &connsdk.HTTPError{Status: http.StatusTooManyRequests, URL: "https://provider.invalid/records"},
		Source:    connsdk.RateLimitObservationSourceRetryAfter,
		HasReset:  true,
		ResetAt:   firstReset,
	}
	rateTwo := &connsdk.RateLimitError{
		HTTPError: &connsdk.HTTPError{Status: http.StatusTooManyRequests, URL: "https://provider.invalid/records"},
		Source:    connsdk.RateLimitObservationSourceHeaders,
		HasReset:  true,
		ResetAt:   secondReset,
	}
	var resumedCheckpoint *synccontract.CheckpointEnvelope
	fixture.sourceExecutor.read = func(_ context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
		switch fixture.sourceExecutor.readCalls {
		case 1:
			if request.Checkpoint != nil {
				return errors.New("initial full-append request unexpectedly used a checkpoint")
			}
			if err := emit(first); err != nil {
				return err
			}
			return rateOne
		case 2:
			if !transportCheckpointEqual(request.Checkpoint, resumedCheckpoint) {
				return fmt.Errorf("reconciled resume checkpoint = %#v, want %#v", request.Checkpoint, resumedCheckpoint)
			}
			return rateTwo
		case 3:
			if !transportCheckpointEqual(request.Checkpoint, resumedCheckpoint) {
				return fmt.Errorf("rearmed resume checkpoint = %#v, want %#v", request.Checkpoint, resumedCheckpoint)
			}
			return emit(completed)
		default:
			return fmt.Errorf("unexpected source read %d", fixture.sourceExecutor.readCalls)
		}
	}

	parked, err := fixture.app.RunETL(context.Background(), RunETLRequest{
		Connection: fixture.connection, Stream: "records", BatchSize: 1,
	})
	if !errors.Is(err, coordination.ErrRateLimitParked) {
		t.Fatalf("initial RunETL() error = %v, want rate-limit parking", err)
	}
	interruptedCheckpoint := fixture.app.state.StreamStates[streamStateKey(fixture.connection, "records")].Checkpoint.Clone()
	interruptedCheckpoint.Position.Primary = synccontract.OpaqueToken("2")
	interruptedCheckpoint.Position.TieBreaker = synccontract.OpaqueToken("2")
	interruptedCheckpoint.ObservedAt = interruptedCheckpoint.ObservedAt.Add(time.Second)
	interruptedAt := interruptedCheckpoint.ObservedAt.Add(time.Second)
	interruptedCheckpoint.CommittedAt = &interruptedAt
	resumed := interruptedCheckpoint.Clone()
	resumedCheckpoint = &resumed
	staleAttemptID, err := prefixedID("run")
	if err != nil {
		t.Fatal(err)
	}
	if err := fixture.app.persistRateParkingRearmAttemptLink(parked.ID, staleAttemptID); err != nil {
		t.Fatalf("persist interrupted rearm link: %v", err)
	}
	if _, err := fixture.app.beginRun(Run{
		ID: staleAttemptID, Type: "etl", Connection: fixture.connection, Stream: "records", Status: "running", BatchSize: 1, StartedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("begin interrupted rearm attempt: %v", err)
	}
	if _, err := fixture.app.updateState(func(current state) (state, error) {
		stream := current.StreamStates[streamStateKey(fixture.connection, "records")]
		stream.Checkpoint = &interruptedCheckpoint
		stream.LastSuccessfulRunID = staleAttemptID
		stream.RecordsLoaded = 1
		stream.UpdatedAt = interruptedAt
		current.StreamStates[streamStateKey(fixture.connection, "records")] = stream
		return current, nil
	}); err != nil {
		t.Fatalf("persist interrupted checkpoint: %v", err)
	}

	now = firstReset
	scheduler.RunThrough(now)
	records, err := parkingStore.List()
	if err != nil || len(records) != 1 || records[0].RunID != parked.ID {
		t.Fatalf("rearmed parking records = %#v, %v; want original run", records, err)
	}
	if !records[0].ResetAt.Equal(secondReset) || records[0].Reason != connsdk.RateLimitObservationSourceHeaders || !transportCheckpointEqual(&records[0].Checkpoint, resumedCheckpoint) {
		t.Fatalf("rearmed parking record = %#v, want interrupted checkpoint and second reset", records[0])
	}
	staleAttempt, found := fixture.app.runByID(staleAttemptID)
	if !found || staleAttempt.Status != "failed" || staleAttempt.Error != rateParkingRearmInterruptedError || staleAttempt.CompletedAt.IsZero() {
		t.Fatalf("interrupted rearm attempt = %#v, want terminal interruption", staleAttempt)
	}
	original, found := fixture.app.runByID(parked.ID)
	if !found || original.Status != string(coordination.RateParkingOutcomeParkedRateLimit) || original.RateParkingRearmAttemptRunID == "" || original.RateParkingRearmAttemptRunID == staleAttemptID {
		t.Fatalf("original parked run = %#v, want linked rearmed attempt", original)
	}
	rearmedAttempt, found := fixture.app.runByID(original.RateParkingRearmAttemptRunID)
	if !found || rearmedAttempt.Status != "failed" || rearmedAttempt.Error != coordination.ErrRateLimitRearmed.Error() || rearmedAttempt.CompletedAt.IsZero() {
		t.Fatalf("rearmed attempt = %#v, want terminal rearm", rearmedAttempt)
	}

	now = secondReset
	scheduler.RunThrough(now)
	if fixture.sourceExecutor.readCalls != 3 || fixture.destinationExecutor.applyCalls != 2 {
		t.Fatalf("source/apply calls = %d/%d, want 3/2", fixture.sourceExecutor.readCalls, fixture.destinationExecutor.applyCalls)
	}
	if records, err := parkingStore.List(); err != nil || len(records) != 0 {
		t.Fatalf("parking records after rearmed recovery = %#v, %v; want none", records, err)
	}
	original, found = fixture.app.runByID(parked.ID)
	if !found || original.Status != "resumed" || original.RateParkingRearmAttemptRunID != "" {
		t.Fatalf("original parked run = %#v, want resumed with cleared rearm link", original)
	}
}

func TestParkRateLimitedRunPersistsOnlyDeclarativeTypedDestinationPlanID(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	t.Cleanup(fixture.app.rateParking.Close)
	fixture.source.rateLimitScope = "transport-source-rate-limit"
	if _, err := fixture.app.RunETL(context.Background(), RunETLRequest{
		Connection: fixture.connection, Stream: "records", BatchSize: 1,
	}); err != nil {
		t.Fatalf("initial RunETL() = %v", err)
	}
	connection, found := fixture.app.findConnection(fixture.connection)
	if !found {
		t.Fatal("fixture connection is unavailable")
	}
	runID, err := prefixedID("run")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.app.beginRun(Run{
		ID: runID, Type: "etl", Connection: connection.Name, Stream: "records", Status: "running", StartedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("begin parked run: %v", err)
	}
	const planID = "rplan_declarative_rate_limit"
	const approvalToken = "must-never-persist"
	typedDestination := &appTransportConnector{
		meta: connectors.Metadata{Name: "typed_rate_limit_destination", DisplayName: "Typed Rate Limit Destination", IntegrationType: "api"},
		descriptor: &connectors.SyncTransportDescriptor{Destination: &connectors.DestinationTransportDescriptor{
			Executor: declarativeTypedDestinationReference,
		}},
	}
	parked, handled, err := fixture.app.parkRateLimitedRun(context.Background(), etlModeDispatchRequest{
		runID: runID, connection: connection, source: fixture.source, destination: typedDestination, streamName: "records",
		destinationApproval: synctransport.DestinationApproval{PlanID: planID, ApprovalToken: approvalToken},
	}, etlExecutionResult{}, &connsdk.RateLimitError{
		HTTPError: &connsdk.HTTPError{Status: http.StatusTooManyRequests, URL: "https://provider.invalid/records"},
		Source:    connsdk.RateLimitObservationSourceRetryAfter,
		HasReset:  true,
		ResetAt:   time.Now().UTC().Add(time.Hour),
	})
	if !handled || !errors.Is(err, coordination.ErrRateLimitParked) {
		t.Fatalf("parkRateLimitedRun() = (%#v, %t, %v), want parked typed run", parked, handled, err)
	}
	if parked.DeclarativeTypedDestinationPlanID != planID {
		t.Fatalf("parked plan ID = %q, want %q", parked.DeclarativeTypedDestinationPlanID, planID)
	}
	persisted, err := fixture.app.GetRun(runID)
	if err != nil || persisted.DeclarativeTypedDestinationPlanID != planID {
		t.Fatalf("persisted parked run = (%#v, %v), want plan ID %q", persisted, err, planID)
	}
	serialized, err := json.Marshal(persisted)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(serialized, []byte(approvalToken)) {
		t.Fatal("parked run persisted an approval token")
	}
}

func TestParkedRateLimitRunETLRequestUsesOnlyDeclarativePlanID(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	checkpoint := fixture.sourceExecutor.page.CandidateCheckpoint
	request, err := parkedRateLimitRunETLRequest(Run{
		Connection: "typed_connection", Stream: "widgets", BatchSize: 17, DeclarativeTypedDestinationPlanID: "rplan_typed_resume",
	}, checkpoint)
	if err != nil {
		t.Fatalf("parked resume request: %v", err)
	}
	if request.Connection != "typed_connection" || request.Stream != "widgets" || request.BatchSize != 17 || request.DestinationApproval.PlanID != "rplan_typed_resume" {
		t.Fatalf("parked resume request = %#v, want connection, stream, batch size, and plan ID", request)
	}
	if !transportCheckpointEqual(request.rateParkingResumeCheckpoint, &checkpoint) {
		t.Fatalf("parked resume request checkpoint = %#v, want %#v", request.rateParkingResumeCheckpoint, checkpoint)
	}
	if request.DestinationApproval.ApprovalToken != "" || request.DestinationApproval.Confirmation.Kind != "" || request.DestinationApproval.Evidence != nil {
		t.Fatalf("parked resume request reconstructed approval material: %#v", request.DestinationApproval)
	}
	if _, err := parkedRateLimitRunETLRequest(Run{Connection: "typed_connection", Stream: "widgets"}, checkpoint); err == nil {
		t.Fatal("parked resume request accepted an unavailable effective batch size")
	}
}

func TestRunETLTransportDoesNotParkDestinationRateLimit(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	t.Cleanup(fixture.app.rateParking.Close)
	fixture.source.rateLimitScope = "transport-source-rate-limit"
	if _, err := fixture.app.RunETL(context.Background(), RunETLRequest{
		Connection: fixture.connection, Stream: "records", BatchSize: 1,
	}); err != nil {
		t.Fatalf("initial RunETL(): %v", err)
	}
	wantOutput := json.RawMessage(`{"provider_status":429}`)
	fixture.destinationExecutor.failApplyAt = fixture.destinationExecutor.applyCalls + 1
	fixture.destinationExecutor.applyErr = synctransport.NewDestinationApplyOutputError(&connsdk.RateLimitError{
		HTTPError: &connsdk.HTTPError{Status: 429, URL: "https://provider.invalid/records"},
		Source:    connsdk.RateLimitObservationSourceRetryAfter,
		HasReset:  true,
		ResetAt:   time.Now().UTC().Add(time.Hour),
	}, wantOutput)

	run, err := fixture.app.RunETL(context.Background(), RunETLRequest{
		Connection: fixture.connection, Stream: "records", BatchSize: 1,
	})
	if errors.Is(err, coordination.ErrRateLimitParked) {
		t.Fatalf("RunETL() error = %v, destination rate limit must not park", err)
	}
	if err == nil || run.Status != "failed" || !reflect.DeepEqual(run.DestinationResults, []json.RawMessage{wantOutput}) {
		t.Fatalf("destination rate-limit run = (%#v, %v), want failed run with provider output", run, err)
	}
}

func TestRunETLTransportPersistsZeroPhaseMeasurementWhenRefusedBeforeSourceIO(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	fixture.app.transports = nil

	run, err := fixture.app.RunETL(context.Background(), RunETLRequest{
		Connection: fixture.connection, Stream: "records", BatchSize: 1,
	})
	if err == nil || !strings.Contains(err.Error(), "closed transport registry is unavailable") {
		t.Fatalf("RunETL() error = %v, want closed transport refusal before source I/O", err)
	}
	if fixture.sourceExecutor.readCalls != 0 || fixture.destinationExecutor.applyCalls != 0 {
		t.Fatalf("source/apply calls = %d/%d, want no I/O after pre-source refusal", fixture.sourceExecutor.readCalls, fixture.destinationExecutor.applyCalls)
	}
	if run.Status != "failed" || run.TransportPhaseMeasurement == nil {
		t.Fatalf("pre-source refused run = %#v, want terminal zero phase measurement", run)
	}
	measurement := run.TransportPhaseMeasurement
	if measurement.ExtractedRecords != 0 || measurement.WarehouseParquetRecords != 0 || measurement.PostgreSQLAppliedRecords != 0 || measurement.ExtractElapsedNanos != 0 || measurement.WarehouseElapsedNanos != 0 || measurement.PostgreSQLElapsedNanos != 0 {
		t.Fatalf("pre-source refused measurement = %#v, want explicit zero counts and durations", measurement)
	}
}

func TestRunETLTransportReconcilesCommittedStagesBeforeSourceIO(t *testing.T) {
	t.Run("reconciles before the source read", func(t *testing.T) {
		fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
		stage := &reconcilingAppTransportStage{}
		fixture.app.transportStage = stage

		if _, err := fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1}); err != nil {
			t.Fatalf("RunETL() = %v", err)
		}
		if stage.reconciliations != 1 || fixture.sourceExecutor.readCalls != 1 {
			t.Fatalf("reconciliations/source reads = %d/%d, want 1/1", stage.reconciliations, fixture.sourceExecutor.readCalls)
		}
	})

	t.Run("reconciliation refusal prevents source io", func(t *testing.T) {
		fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
		wantErr := errors.New("synthetic transient reconciliation failure")
		stage := &reconcilingAppTransportStage{err: wantErr}
		fixture.app.transportStage = stage

		_, err := fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
		if !errors.Is(err, wantErr) {
			t.Fatalf("RunETL() error = %v, want %v", err, wantErr)
		}
		if stage.reconciliations != 1 || fixture.sourceExecutor.readCalls != 0 || fixture.destinationExecutor.applyCalls != 0 {
			t.Fatalf("reconciliations/source reads/destination applies = %d/%d/%d, want 1/0/0", stage.reconciliations, fixture.sourceExecutor.readCalls, fixture.destinationExecutor.applyCalls)
		}
	})
}

func TestRunETLTransportRefusesDeclaredChangeCaptureDestinationBeforeIO(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeChangeCapture)

	_, err := fixture.app.RunETL(context.Background(), RunETLRequest{
		Connection: fixture.connection,
		Stream:     "records",
		BatchSize:  1,
	})
	if err == nil || !strings.Contains(err.Error(), "does not support sync mode \"change_capture\"") {
		t.Fatalf("RunETL() error = %v, want destination change_capture refusal", err)
	}
	stage, ok := fixture.app.transportStage.(*appTransportStage)
	if !ok {
		t.Fatalf("transport stage = %T, want app transport stage", fixture.app.transportStage)
	}
	if fixture.sourceExecutor.readCalls != 0 || stage.calls != 0 || fixture.destinationExecutor.planCalls != 0 || fixture.destinationExecutor.applyCalls != 0 {
		t.Fatalf("change_capture refusal source/stage/plan/apply = %d/%d/%d/%d, want zero before I/O", fixture.sourceExecutor.readCalls, stage.calls, fixture.destinationExecutor.planCalls, fixture.destinationExecutor.applyCalls)
	}
	if fixture.source.legacyReadCalls != 0 || fixture.destination.legacyWriteCalls != 0 {
		t.Fatalf("change_capture legacy calls read=%d write=%d, want zero before refusal", fixture.source.legacyReadCalls, fixture.destination.legacyWriteCalls)
	}
}

// A DestinationApproval is meaningful only to the exact closed GitHub
// transport route. App callers must not be able to attach one to an ordinary
// ETL request and have it silently ignored while legacy execution proceeds.
func TestRunETLRejectsDestinationApprovalOutsideClosedTransportRoute(t *testing.T) {
	source := newScriptedSyncSource("approval_guard_source", []connectors.Record{{"id": "1"}})
	a, connection := setupSyncModeApp(t, source, "full_refresh_overwrite")

	_, err := a.RunETL(context.Background(), RunETLRequest{
		Connection: connection,
		Stream:     "records",
		BatchSize:  1,
		DestinationApproval: synctransport.DestinationApproval{
			PlanID:        "rplan_closed_transport_only",
			ApprovalToken: "test-approval-must-not-reach-legacy-etl",
			Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
		},
	})
	if err == nil {
		t.Fatal("RunETL() accepted destination approval material on an ordinary ETL route")
	}
	if len(source.requests) != 0 {
		t.Fatalf("ordinary source Read calls = %d, want approval rejected before legacy ETL I/O", len(source.requests))
	}
}

var (
	errTransportSourceAfterAcknowledgement = errors.New("source failed after acknowledged page")
	errTransportFinalStateSave             = errors.New("final transport state save failed")
	errTransportCheckpointState            = errors.New("checkpoint state directory sync failed")
	errTransportFinalizationStateSync      = errors.New("finalization state directory sync failed")
)

func TestRunETLTransportRejectsAcknowledgedCheckpointWithIncompatibleResume(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*synccontract.CheckpointEnvelope)
		wantOutcome synccontract.RecoveryOutcome
	}{
		{
			name: "another credential",
			mutate: func(checkpoint *synccontract.CheckpointEnvelope) {
				checkpoint.Source.AccountOrCluster = "another-credential"
			},
			wantOutcome: synccontract.RecoveryOutcomeSourceIdentityIncompatible,
		},
		{
			name: "another stream",
			mutate: func(checkpoint *synccontract.CheckpointEnvelope) {
				checkpoint.Source.ObjectScope = "another-stream"
			},
			wantOutcome: synccontract.RecoveryOutcomeSourceIdentityIncompatible,
		},
		{
			name: "generation mismatch",
			mutate: func(checkpoint *synccontract.CheckpointEnvelope) {
				checkpoint.SourceGeneration = synccontract.OpaqueToken{0xff, 0x00}
			},
			wantOutcome: synccontract.RecoveryOutcomeSourceGenerationChanged,
		},
		{
			name: "matched control",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
			if tt.mutate != nil {
				tt.mutate(&fixture.sourceExecutor.page.CandidateCheckpoint)
			}

			run, err := fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
			stateKey := streamStateKey(fixture.connection, "records")
			if tt.wantOutcome == "" {
				if err != nil {
					t.Fatalf("RunETL() = %v, want matched checkpoint accepted", err)
				}
				if run.Status != "completed" {
					t.Fatalf("run status = %q, want completed", run.Status)
				}
				if state := fixture.app.state.StreamStates[stateKey]; state.Checkpoint == nil || state.Checkpoint.CommittedAt == nil {
					t.Fatalf("matched checkpoint was not persisted: %#v", state)
				}
				return
			}

			var recovery *synccontract.RebootstrapRequiredError
			if !errors.As(err, &recovery) || recovery.Outcome != tt.wantOutcome {
				t.Fatalf("RunETL() error = %T %v, want rebootstrap outcome %q", err, err, tt.wantOutcome)
			}
			if run.Status != "failed" {
				t.Fatalf("run status = %q, want failed", run.Status)
			}
			if _, exists := fixture.app.state.StreamStates[stateKey]; exists {
				t.Fatalf("incompatible checkpoint persisted as stream state: %#v", fixture.app.state.StreamStates[stateKey])
			}
			if fixture.destinationExecutor.applyCalls != 1 {
				t.Fatalf("destination apply calls = %d, want durable acknowledgement attempted once", fixture.destinationExecutor.applyCalls)
			}
		})
	}
}

func TestRunETLTransportPersistsActiveCheckpointBeforeSourceFailureForPerPageAcknowledgementModes(t *testing.T) {
	for _, contractMode := range appTransportPerPageAcknowledgementModes() {
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
			assertInterimTransportStateWithMetadata(t, fixture.app, stateKey, []byte("current"), wantGeneration, "previous_run", 99)
			unrelated := fixture.app.state.StreamStates["unrelated:records"]
			if unrelated.GenerationID != 17 || unrelated.LastSuccessfulRunID != "unrelated_run" || unrelated.RecordsLoaded != 7 || !unrelated.UpdatedAt.Equal(unrelatedUpdatedAt) {
				t.Fatalf("unrelated stream state changed during checkpoint commit: %#v", unrelated)
			}

			reopened, err := Open(fixture.app.root)
			if err != nil {
				t.Fatal(err)
			}
			assertInterimTransportStateWithMetadata(t, reopened, stateKey, []byte("current"), wantGeneration, "previous_run", 99)
		})
	}
}

func TestRunETLTransportFullOverwriteSourceFailureAbortsWithoutCheckpoint(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullOverwrite)
	assertFullOverwriteStopsBeforePublish(t, fixture, context.Background(), func() {
		fixture.sourceExecutor.errAfterPage = errTransportSourceAfterAcknowledgement
	}, errTransportSourceAfterAcknowledgement)
}

func TestRunETLTransportFullOverwriteCancellationBeforePublishAbortsWithoutCheckpoint(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullOverwrite)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	assertFullOverwriteStopsBeforePublish(t, fixture, ctx, func() {
		fixture.sourceExecutor.afterEmit = cancel
	}, context.Canceled)
}

func assertFullOverwriteStopsBeforePublish(t *testing.T, fixture appTransportFixture, ctx context.Context, configure func(), wantErr error) {
	t.Helper()
	stateKey := streamStateKey(fixture.connection, "records")
	previous := fixture.sourceExecutor.page.CandidateCheckpoint.Clone()
	previous.Position.Primary = synccontract.OpaqueToken("previous")
	previous.Position.TieBreaker = synccontract.OpaqueToken("previous")
	previousCommittedAt := previous.ObservedAt.Add(time.Second)
	previous.CommittedAt = &previousCommittedAt
	if _, err := fixture.app.updateState(func(current state) (state, error) {
		if current.StreamStates == nil {
			current.StreamStates = map[string]StreamState{}
		}
		current.StreamStates[stateKey] = StreamState{
			Connection: fixture.connection, Stream: "records", Checkpoint: &previous,
			GenerationID: 9, LastSuccessfulRunID: "previous_run", RecordsLoaded: 99,
			UpdatedAt: previousCommittedAt,
		}
		return current, nil
	}); err != nil {
		t.Fatal(err)
	}
	configure()

	run, err := fixture.app.RunETL(ctx, RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
	if !errors.Is(err, wantErr) {
		t.Fatalf("RunETL() error = %v, want %v", err, wantErr)
	}
	if run.Status != "failed" {
		t.Fatalf("run status = %q, want failed before full-overwrite publication", run.Status)
	}
	assertInterimTransportStateWithMetadata(t, fixture.app, stateKey, []byte("previous"), 9, "previous_run", 99)
	if fixture.destinationExecutor.applyCalls != 1 || fixture.destinationExecutor.abortCalls != 1 || fixture.destinationExecutor.publishCalls != 0 || fixture.destinationExecutor.readBackCalls != 0 {
		t.Fatalf("full-overwrite lifecycle apply/abort/publish/read-back = %d/%d/%d/%d, want 1/1/0/0", fixture.destinationExecutor.applyCalls, fixture.destinationExecutor.abortCalls, fixture.destinationExecutor.publishCalls, fixture.destinationExecutor.readBackCalls)
	}

	reopened, reopenErr := Open(fixture.app.root)
	if reopenErr != nil {
		t.Fatal(reopenErr)
	}
	assertInterimTransportStateWithMetadata(t, reopened, stateKey, []byte("previous"), 9, "previous_run", 99)
}

func TestRunETLTransportAdvancesInterimCheckpointAcrossPages(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	stateKey := streamStateKey(fixture.connection, "records")
	prior := fixture.sourceExecutor.page.CandidateCheckpoint.Clone()
	prior.Position.Primary = synccontract.OpaqueToken{0x00, 0xff}
	prior.Position.TieBreaker = synccontract.OpaqueToken{0x00, 0xff}
	priorCommittedAt := prior.ObservedAt.Add(time.Second)
	prior.CommittedAt = &priorCommittedAt
	if _, err := fixture.app.updateState(func(current state) (state, error) {
		if current.StreamStates == nil {
			current.StreamStates = map[string]StreamState{}
		}
		current.StreamStates[stateKey] = StreamState{
			Connection:          fixture.connection,
			Stream:              "records",
			Checkpoint:          &prior,
			GenerationID:        3,
			LastSuccessfulRunID: "previous_run",
			RecordsLoaded:       42,
			UpdatedAt:           priorCommittedAt,
		}
		return current, nil
	}); err != nil {
		t.Fatal(err)
	}

	first := fixture.sourceExecutor.page.CandidateCheckpoint.Clone()
	first.Position.Primary = synccontract.OpaqueToken{0xff}
	first.Position.TieBreaker = synccontract.OpaqueToken{0xff}
	first.CommittedAt = nil
	second := first.Clone()
	second.Position.Primary = synccontract.OpaqueToken{0xff, 0x00}
	second.Position.TieBreaker = synccontract.OpaqueToken{0xff, 0x00}
	fixture.sourceExecutor.pages = []synctransport.SourcePage{
		{Records: []connectors.Record{{"id": "first"}}, CandidateCheckpoint: first},
		{Records: []connectors.Record{{"id": "second"}}, CandidateCheckpoint: second},
	}
	fixture.sourceExecutor.errAfterPage = errTransportSourceAfterAcknowledgement

	run, err := fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
	if !errors.Is(err, errTransportSourceAfterAcknowledgement) {
		t.Fatalf("RunETL() error = %v, want source failure after two acknowledgements", err)
	}
	if run.Status != "failed" {
		t.Fatalf("run status = %q, want failed", run.Status)
	}
	if fixture.destinationExecutor.applyCalls != 2 {
		t.Fatalf("destination apply calls = %d, want two", fixture.destinationExecutor.applyCalls)
	}
	assertInterimTransportStateWithMetadata(t, fixture.app, stateKey, []byte{0xff, 0x00}, 3, "previous_run", 42)
}

func TestRunETLTransportPreservesUnrelatedStateDuringInterimCheckpointCommit(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	acknowledgementReached := make(chan struct{})
	releaseAcknowledgement := make(chan struct{})
	fixture.destinationExecutor.afterApply = func() {
		close(acknowledgementReached)
		<-releaseAcknowledgement
	}

	done := make(chan struct{})
	var run Run
	var runErr error
	go func() {
		run, runErr = fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
		close(done)
	}()
	waitForTransportSignal(t, acknowledgementReached)

	writer, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	unrelatedUpdatedAt := time.Unix(11, 0).UTC()
	if _, err := writer.updateState(func(current state) (state, error) {
		if current.StreamStates == nil {
			current.StreamStates = map[string]StreamState{}
		}
		current.StreamStates["unrelated:records"] = StreamState{
			Connection:          "unrelated",
			Stream:              "records",
			GenerationID:        8,
			LastSuccessfulRunID: "unrelated_run",
			RecordsLoaded:       13,
			UpdatedAt:           unrelatedUpdatedAt,
		}
		if current.Checkpoints == nil {
			current.Checkpoints = map[string]map[string]string{}
		}
		current.Checkpoints["unrelated_run"] = map[string]string{"cursor": "preserved"}
		return current, nil
	}); err != nil {
		t.Fatal(err)
	}
	close(releaseAcknowledgement)
	waitForTransportSignal(t, done)

	if runErr != nil {
		t.Fatalf("RunETL() = %v, want target-stream checkpoint commit", runErr)
	}
	if run.Status != "completed" {
		t.Fatalf("run status = %q, want completed", run.Status)
	}
	unrelated := fixture.app.state.StreamStates["unrelated:records"]
	if unrelated.GenerationID != 8 || unrelated.LastSuccessfulRunID != "unrelated_run" || unrelated.RecordsLoaded != 13 || !unrelated.UpdatedAt.Equal(unrelatedUpdatedAt) {
		t.Fatalf("unrelated stream state changed: %#v", unrelated)
	}
	if got := fixture.app.state.Checkpoints["unrelated_run"]["cursor"]; got != "preserved" {
		t.Fatalf("unrelated project checkpoint = %q, want preserved", got)
	}
}

func TestRunETLTransportRejectsStaleCheckpointWriter(t *testing.T) {
	assertRunETLTransportStaleWriterFinalization(t, synccontract.ModeFullAppend, false)
}

func TestRateParkingResumeClassifiesExpiredAndRevokedAuthorizationAsTerminal(t *testing.T) {
	for _, want := range []error{
		&AuthorizationExpiredError{Reference: "expired-authorization"},
		&AuthorizationRevokedError{Reference: "revoked-authorization"},
	} {
		got := classifyRateParkingResumeError(want)
		var terminal *coordination.NeedsReauthorizationError
		if !errors.As(got, &terminal) {
			t.Fatalf("classifyRateParkingResumeError(%T) = %T %v, want terminal rate-parking error", want, got, got)
		}
		if !errors.Is(got, want) {
			t.Fatalf("terminal error %v did not preserve %T", got, want)
		}
	}
	ordinary := errors.New("ordinary provider failure")
	if got := classifyRateParkingResumeError(ordinary); got != ordinary {
		t.Fatalf("classifyRateParkingResumeError(ordinary) = %T %v, want original retryable error", got, got)
	}
}

func TestTransportTwoAppsFenceBeforeAnySideEffect(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	claimed := make(chan struct{})
	releaseSource := make(chan struct{})
	var firstRead sync.Once
	fixture.sourceExecutor.beforeRead = func() {
		firstRead.Do(func() {
			close(claimed)
			<-releaseSource
		})
	}

	firstDone := make(chan struct{})
	var firstErr error
	go func() {
		_, firstErr = fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
		close(firstDone)
	}()
	waitForTransportSignal(t, claimed)

	contender, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	credential, ok := contender.findCredential("source")
	if !ok {
		t.Fatal("contender has no source credential")
	}
	contenderSource := &appTransportSourceExecutor{reference: fixture.sourceExecutor.reference, page: synctransport.SourcePage{
		Records:             []connectors.Record{{"id": "contender"}},
		CandidateCheckpoint: appTransportCheckpoint(fixture.source, credential, "records"),
	}}
	contenderDestination := &appTransportDestinationExecutor{reference: fixture.destinationExecutor.reference, sink: fixture.destination.Name()}
	fixture.configureRuntime(t, contender, contenderSource, contenderDestination)
	contenderStage := &appTransportStage{}
	contender.transportStage = contenderStage

	_, contenderErr := contender.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
	contenderEffects := []int{contenderSource.readCalls, contenderStage.calls, contenderDestination.applyCalls, contenderDestination.publishCalls}

	close(releaseSource)
	waitForTransportSignal(t, firstDone)
	if firstErr != nil {
		t.Fatalf("first RunETL() error = %v", firstErr)
	}
	if contenderErr == nil || !strings.Contains(contenderErr.Error(), "transport stream work") {
		t.Fatalf("contender RunETL() error = %v, want durable work-fence refusal", contenderErr)
	}
	if contenderEffects[0] != 0 || contenderEffects[1] != 0 || contenderEffects[2] != 0 || contenderEffects[3] != 0 {
		t.Fatalf("contender side effects source/stage/apply/publish = %d/%d/%d/%d, want all zero", contenderEffects[0], contenderEffects[1], contenderEffects[2], contenderEffects[3])
	}
}

func TestRunETLTransportStaleWriterFinalizesLosingRun(t *testing.T) {
	assertRunETLTransportStaleWriterFinalization(t, synccontract.ModeFullAppend, false)
}

func TestRunETLTransportStaleWriterDoesNotReportUncommittedFinalization(t *testing.T) {
	// The previous late-CAS scenario intentionally let both writers apply.
	// B26 supersedes it with a pre-I/O durable fence, so this variant must use
	// the same no-contender-effect proof rather than recreate unsafe effects.
	assertRunETLTransportStaleWriterFinalization(t, synccontract.ModeFullAppend, false)
	return

	/* Historical late-CAS fixture retained only while this uncommitted recovery
	set is being reconciled; it intentionally models the unsafe behavior B26
	replaces and is not executable evidence. */
	/*
		fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
		fixture.sourceExecutor.page.CandidateCheckpoint.Position.Primary = synccontract.OpaqueToken{0xff}
		fixture.sourceExecutor.page.CandidateCheckpoint.Position.TieBreaker = synccontract.OpaqueToken{0xff}
		acknowledgementReached := make(chan struct{})
		releaseAcknowledgement := make(chan struct{})
		fixture.destinationExecutor.afterApply = func() {
			close(acknowledgementReached)
			<-releaseAcknowledgement
		}

		stateDir := filepath.Dir(fixture.app.statePath)
		stateDirInfo, err := os.Stat(stateDir)
		if err != nil {
			t.Fatal(err)
		}
		locker := &appTransportPreRenamePersistenceFailureLocker{
			directory:   stateDir,
			restoreMode: stateDirInfo.Mode().Perm(),
			failAt:      3,
		}
		fixture.app.store.Locker = locker
		t.Cleanup(func() {
			if err := os.Chmod(stateDir, stateDirInfo.Mode().Perm()); err != nil {
				t.Errorf("restore state directory mode: %v", err)
			}
		})

		done := make(chan struct{})
		var losingRun Run
		var losingErr error
		go func() {
			losingRun, losingErr = fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
			close(done)
		}()
		waitForTransportSignal(t, acknowledgementReached)

		var losingRunID string
		for _, run := range fixture.app.state.Runs {
			if run.Type == "etl" && run.Connection == fixture.connection && run.Stream == "records" {
				losingRunID = run.ID
				break
			}
		}
		if losingRunID == "" {
			t.Fatal("losing run was not persisted before acknowledgement")
		}

		winner, err := Open(fixture.app.root)
		if err != nil {
			t.Fatal(err)
		}
		unrelatedUpdatedAt := time.Unix(11, 0).UTC()
		if _, err := winner.updateState(func(current state) (state, error) {
			if current.StreamStates == nil {
				current.StreamStates = map[string]StreamState{}
			}
			current.StreamStates["unrelated:records"] = StreamState{
				Connection:          "unrelated",
				Stream:              "records",
				GenerationID:        8,
				LastSuccessfulRunID: "unrelated_run",
				RecordsLoaded:       13,
				UpdatedAt:           unrelatedUpdatedAt,
			}
			if current.Checkpoints == nil {
				current.Checkpoints = map[string]map[string]string{}
			}
			current.Checkpoints["unrelated_run"] = map[string]string{"cursor": "preserved"}
			current.Runs = append(current.Runs, Run{
				ID:          "unrelated_run",
				Type:        "etl",
				Connection:  "unrelated",
				Stream:      "records",
				Status:      "completed",
				StartedAt:   unrelatedUpdatedAt.Add(-time.Second),
				CompletedAt: unrelatedUpdatedAt,
			})
			return current, nil
		}); err != nil {
			t.Fatal(err)
		}
		credential, ok := winner.findCredential("source")
		if !ok {
			t.Fatal("winner app has no source credential")
		}
		winnerCheckpoint := appTransportCheckpoint(fixture.source, credential, "records")
		winnerCheckpoint.Position.Primary = synccontract.OpaqueToken{0xff, 0x00}
		winnerCheckpoint.Position.TieBreaker = synccontract.OpaqueToken{0xff, 0x00}
		winnerSource := &appTransportSourceExecutor{
			reference: fixture.sourceExecutor.reference,
			page: synctransport.SourcePage{
				Records:             []connectors.Record{{"id": "winner"}},
				CandidateCheckpoint: winnerCheckpoint,
			},
		}
		winnerDestination := &appTransportDestinationExecutor{reference: fixture.destinationExecutor.reference, sink: fixture.destination.Name()}
		fixture.configureRuntime(t, winner, winnerSource, winnerDestination)
		winnerRun, err := winner.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
		if err != nil {
			t.Fatalf("winner RunETL() = %v", err)
		}
		if winnerRun.Status != "completed" {
			t.Fatalf("winner run status = %q, want completed", winnerRun.Status)
		}

		close(releaseAcknowledgement)
		waitForTransportSignal(t, done)
		if locker.calls != 3 {
			t.Fatalf("state lock calls = %d, want three calls through finalization", locker.calls)
		}
		if !errors.Is(losingErr, errTransportStreamStateConflict) {
			t.Fatalf("losing RunETL() error = %v, want typed stale stream-state rejection", losingErr)
		}
		if outcome := statestore.CommitOutcomeForError(losingErr); outcome != statestore.CommitOutcomeNotCommitted {
			t.Fatalf("finalization commit outcome = %s, want not committed", outcome)
		}
		if !strings.Contains(losingErr.Error(), "create temporary state file") {
			t.Fatalf("losing RunETL() error = %v, want pre-rename state persistence failure", losingErr)
		}
		if !reflect.DeepEqual(losingRun, Run{}) {
			t.Fatalf("RunETL() returned %#v, want zero Run after uncommitted finalization", losingRun)
		}

		reopened, err := Open(fixture.app.root)
		if err != nil {
			t.Fatal(err)
		}
		stateKey := streamStateKey(fixture.connection, "records")
		streamState := reopened.state.StreamStates[stateKey]
		if streamState.Checkpoint == nil || !bytes.Equal(streamState.Checkpoint.Position.Primary, []byte{0xff, 0x00}) {
			t.Fatalf("winner checkpoint was overwritten: %#v", streamState.Checkpoint)
		}
		if streamState.LastSuccessfulRunID != winnerRun.ID {
			t.Fatalf("winner run identity = %q, want %q", streamState.LastSuccessfulRunID, winnerRun.ID)
		}
		unrelated := reopened.state.StreamStates["unrelated:records"]
		if unrelated.GenerationID != 8 || unrelated.LastSuccessfulRunID != "unrelated_run" || unrelated.RecordsLoaded != 13 || !unrelated.UpdatedAt.Equal(unrelatedUpdatedAt) {
			t.Fatalf("unrelated stream state changed: %#v", unrelated)
		}
		if got := reopened.state.Checkpoints["unrelated_run"]["cursor"]; got != "preserved" {
			t.Fatalf("unrelated project checkpoint = %q, want preserved", got)
		}
		var durableLoser, unrelatedRun Run
		for _, run := range reopened.state.Runs {
			switch run.ID {
			case losingRunID:
				durableLoser = run
			case "unrelated_run":
				unrelatedRun = run
			}
		}
		if durableLoser.ID != losingRunID || durableLoser.Status != "running" || !durableLoser.CompletedAt.IsZero() {
			t.Fatalf("durable losing run = %#v, want running unfinalized run", durableLoser)
		}
		if unrelatedRun.Status != "completed" || !unrelatedRun.CompletedAt.Equal(unrelatedUpdatedAt) {
			t.Fatalf("unrelated run changed: %#v", unrelatedRun)
		}
	*/
}

// assertTransportContenderIsFencedBeforeEffects replaces the historical
// late-CAS race fixture. The owner pauses immediately after the durable claim
// and before its source reads; a second App must fail without source, stage,
// apply, or publish I/O for every ordinary and full-overwrite mode.
func assertTransportContenderIsFencedBeforeEffects(t *testing.T, mode synccontract.Mode, configureSource ...func(*appTransportSourceExecutor)) {
	t.Helper()
	fixture := setupAppTransportFixture(t, mode)
	fixture.sourceExecutor.page.CandidateCheckpoint.Position.Primary = synccontract.OpaqueToken{0xff}
	fixture.sourceExecutor.page.CandidateCheckpoint.Position.TieBreaker = synccontract.OpaqueToken{0xff}
	for _, configure := range configureSource {
		configure(fixture.sourceExecutor)
	}
	claimed := make(chan struct{})
	release := make(chan struct{})
	var sourceOnce sync.Once
	fixture.sourceExecutor.beforeRead = func() {
		sourceOnce.Do(func() {
			close(claimed)
			<-release
		})
	}
	firstDone := make(chan struct{})
	var firstRun Run
	var firstErr error
	go func() {
		firstRun, firstErr = fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
		close(firstDone)
	}()
	waitForTransportSignal(t, claimed)

	contender, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	credential, ok := contender.findCredential("source")
	if !ok {
		t.Fatal("contender app has no source credential")
	}
	contenderSource := &appTransportSourceExecutor{reference: fixture.sourceExecutor.reference, page: synctransport.SourcePage{
		Records:             []connectors.Record{{"id": "contender"}},
		CandidateCheckpoint: appTransportCheckpoint(fixture.source, credential, "records"),
	}}
	contenderDestination := &appTransportDestinationExecutor{reference: fixture.destinationExecutor.reference, sink: fixture.destination.Name()}
	fixture.configureRuntime(t, contender, contenderSource, contenderDestination)
	contenderStage := &appTransportStage{}
	contender.transportStage = contenderStage
	contenderRun, contenderErr := contender.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
	contenderEffects := []int{contenderSource.readCalls, contenderStage.calls, contenderDestination.applyCalls, contenderDestination.publishCalls}

	close(release)
	waitForTransportSignal(t, firstDone)
	if firstErr != nil || firstRun.Status != "completed" {
		t.Fatalf("owner RunETL() = %#v, %v; want completed owner", firstRun, firstErr)
	}
	if !errors.Is(contenderErr, errTransportStreamWorkInProgress) || contenderRun.Status == "completed" {
		t.Fatalf("contender RunETL() = %#v, %v; want failed pre-I/O work-fence refusal", contenderRun, contenderErr)
	}
	if contenderEffects[0] != 0 || contenderEffects[1] != 0 || contenderEffects[2] != 0 || contenderEffects[3] != 0 {
		t.Fatalf("contender side effects source/stage/apply/publish = %d/%d/%d/%d, want all zero", contenderEffects[0], contenderEffects[1], contenderEffects[2], contenderEffects[3])
	}
	reopened, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	streamState := reopened.state.StreamStates[streamStateKey(fixture.connection, "records")]
	if streamState.Checkpoint == nil || (!bytes.Equal(streamState.Checkpoint.Position.Primary, []byte{0xff}) && len(fixture.sourceExecutor.pages) == 0) {
		t.Fatalf("owner checkpoint = %#v, want the owner checkpoint", streamState.Checkpoint)
	}
	if streamState.LastSuccessfulRunID != firstRun.ID || streamState.ActiveWorkID != "" || streamState.ActiveWorkLeaseUntil != nil {
		t.Fatalf("terminal owner stream state = %#v, want cleared work lease for run %q", streamState, firstRun.ID)
	}
}

func assertRunETLTransportStaleWriterFinalization(t *testing.T, mode synccontract.Mode, cancelAfterAcknowledgement bool, configureSource ...func(*appTransportSourceExecutor)) {
	t.Helper()
	_ = cancelAfterAcknowledgement
	assertTransportContenderIsFencedBeforeEffects(t, mode, configureSource...)
	return

	/* Historical late-CAS fixture retained only while this uncommitted recovery
	set is being reconciled; the pre-I/O fence above is its replacement. */
	/*
		fixture := setupAppTransportFixture(t, mode)
		fixture.sourceExecutor.page.CandidateCheckpoint.Position.Primary = synccontract.OpaqueToken{0xff}
		fixture.sourceExecutor.page.CandidateCheckpoint.Position.TieBreaker = synccontract.OpaqueToken{0xff}
		for _, configure := range configureSource {
			configure(fixture.sourceExecutor)
		}
		acknowledgementReached := make(chan struct{})
		releaseAcknowledgement := make(chan struct{})
		pauseAfterAcknowledgement := func() {
			close(acknowledgementReached)
			<-releaseAcknowledgement
		}
		if mode == synccontract.ModeFullOverwrite {
			// A full-overwrite receipt exists only after publication and read-back.
			fixture.destinationExecutor.afterReadBack = pauseAfterAcknowledgement
		} else {
			fixture.destinationExecutor.afterApply = pauseAfterAcknowledgement
		}

		done := make(chan struct{})
		var losingRun Run
		var losingErr error
		losingCtx := context.Background()
		var cancel context.CancelFunc
		if cancelAfterAcknowledgement {
			losingCtx, cancel = context.WithCancel(context.Background())
			defer cancel()
		}
		go func() {
			losingRun, losingErr = fixture.app.RunETL(losingCtx, RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
			close(done)
		}()
		waitForTransportSignal(t, acknowledgementReached)

		winner, err := Open(fixture.app.root)
		if err != nil {
			t.Fatal(err)
		}
		unrelatedUpdatedAt := time.Unix(11, 0).UTC()
		if _, err := winner.updateState(func(current state) (state, error) {
			if current.StreamStates == nil {
				current.StreamStates = map[string]StreamState{}
			}
			current.StreamStates["unrelated:records"] = StreamState{
				Connection:          "unrelated",
				Stream:              "records",
				GenerationID:        8,
				LastSuccessfulRunID: "unrelated_run",
				RecordsLoaded:       13,
				UpdatedAt:           unrelatedUpdatedAt,
			}
			if current.Checkpoints == nil {
				current.Checkpoints = map[string]map[string]string{}
			}
			current.Checkpoints["unrelated_run"] = map[string]string{"cursor": "preserved"}
			current.Runs = append(current.Runs, Run{
				ID:          "unrelated_run",
				Type:        "etl",
				Connection:  "unrelated",
				Stream:      "records",
				Status:      "completed",
				StartedAt:   unrelatedUpdatedAt.Add(-time.Second),
				CompletedAt: unrelatedUpdatedAt,
			})
			return current, nil
		}); err != nil {
			t.Fatal(err)
		}
		credential, ok := winner.findCredential("source")
		if !ok {
			t.Fatal("winner app has no source credential")
		}
		winnerCheckpoint := appTransportCheckpoint(fixture.source, credential, "records")
		winnerCheckpoint.Position.Primary = synccontract.OpaqueToken{0xff, 0x00}
		winnerCheckpoint.Position.TieBreaker = synccontract.OpaqueToken{0xff, 0x00}
		winnerSource := &appTransportSourceExecutor{
			reference: fixture.sourceExecutor.reference,
			page: synctransport.SourcePage{
				Records:             []connectors.Record{{"id": "winner"}},
				CandidateCheckpoint: winnerCheckpoint,
			},
		}
		winnerDestination := &appTransportDestinationExecutor{reference: fixture.destinationExecutor.reference, sink: fixture.destination.Name()}
		fixture.configureRuntime(t, winner, winnerSource, winnerDestination)
		winnerRun, err := winner.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
		if err != nil {
			t.Fatalf("winner RunETL() = %v", err)
		}
		if winnerRun.Status != "completed" {
			t.Fatalf("winner run status = %q, want completed", winnerRun.Status)
		}
		if mode == synccontract.ModeFullOverwrite {
			wantApplies := 1
			if len(fixture.sourceExecutor.pages) != 0 {
				wantApplies = len(fixture.sourceExecutor.pages)
			}
			if fixture.destinationExecutor.applyCalls != wantApplies || fixture.destinationExecutor.publishCalls != 1 || fixture.destinationExecutor.readBackCalls != 1 || fixture.destinationExecutor.abortCalls != 0 {
				t.Fatalf("loser full-overwrite lifecycle apply/publish/read-back/abort = %d/%d/%d/%d, want %d/1/1/0", fixture.destinationExecutor.applyCalls, fixture.destinationExecutor.publishCalls, fixture.destinationExecutor.readBackCalls, fixture.destinationExecutor.abortCalls, wantApplies)
			}
			if winnerDestination.applyCalls != 1 || winnerDestination.publishCalls != 1 || winnerDestination.readBackCalls != 1 || winnerDestination.abortCalls != 0 {
				t.Fatalf("winner full-overwrite lifecycle apply/publish/read-back/abort = %d/%d/%d/%d, want 1/1/1/0", winnerDestination.applyCalls, winnerDestination.publishCalls, winnerDestination.readBackCalls, winnerDestination.abortCalls)
			}
		}

		if cancel != nil {
			cancel()
			if !errors.Is(losingCtx.Err(), context.Canceled) {
				t.Fatalf("losing context error = %v, want cancellation after acknowledgement", losingCtx.Err())
			}
		}
		close(releaseAcknowledgement)
		waitForTransportSignal(t, done)
		if !errors.Is(losingErr, errTransportStreamStateConflict) {
			t.Fatalf("losing RunETL() error = %v, want typed stale stream-state rejection", losingErr)
		}

		reopened, err := Open(fixture.app.root)
		if err != nil {
			t.Fatal(err)
		}
		streamState := reopened.state.StreamStates[streamStateKey(fixture.connection, "records")]
		if streamState.Checkpoint == nil || !bytes.Equal(streamState.Checkpoint.Position.Primary, []byte{0xff, 0x00}) {
			t.Fatalf("winner checkpoint was overwritten: %#v", streamState.Checkpoint)
		}
		if streamState.LastSuccessfulRunID != winnerRun.ID {
			t.Fatalf("winner run identity = %q, want %q", streamState.LastSuccessfulRunID, winnerRun.ID)
		}
		unrelated := reopened.state.StreamStates["unrelated:records"]
		if unrelated.GenerationID != 8 || unrelated.LastSuccessfulRunID != "unrelated_run" || unrelated.RecordsLoaded != 13 || !unrelated.UpdatedAt.Equal(unrelatedUpdatedAt) {
			t.Fatalf("unrelated stream state changed: %#v", unrelated)
		}
		if got := reopened.state.Checkpoints["unrelated_run"]["cursor"]; got != "preserved" {
			t.Fatalf("unrelated project checkpoint = %q, want preserved", got)
		}
		var unrelatedRun Run
		for _, run := range reopened.state.Runs {
			if run.ID == "unrelated_run" {
				unrelatedRun = run
				break
			}
		}
		if unrelatedRun.Status != "completed" || !unrelatedRun.CompletedAt.Equal(unrelatedUpdatedAt) {
			t.Fatalf("unrelated run changed: %#v", unrelatedRun)
		}

		var durableLoser Run
		loserCount := 0
		for _, run := range reopened.state.Runs {
			if run.ID == winnerRun.ID || run.Type != "etl" || run.Connection != fixture.connection || run.Stream != "records" {
				continue
			}
			durableLoser = run
			loserCount++
		}
		var symptoms []string
		if losingRun.ID == "" {
			symptoms = append(symptoms, "RunETL returned zero losing Run")
		}
		if losingRun.Status != "failed" {
			symptoms = append(symptoms, fmt.Sprintf("returned loser status=%q", losingRun.Status))
		}
		if loserCount != 1 {
			symptoms = append(symptoms, fmt.Sprintf("durable loser count=%d", loserCount))
		} else {
			if durableLoser.Status != "failed" {
				symptoms = append(symptoms, fmt.Sprintf("durable loser status=%q", durableLoser.Status))
			}
			if durableLoser.CompletedAt.IsZero() {
				symptoms = append(symptoms, "durable loser completion timestamp is zero")
			}
			if losingRun.ID != durableLoser.ID {
				symptoms = append(symptoms, fmt.Sprintf("returned loser ID=%q, durable loser ID=%q", losingRun.ID, durableLoser.ID))
			}
		}
		if len(symptoms) > 0 {
			t.Fatalf("stale writer finalization leak: %s; durable loser=%+v", strings.Join(symptoms, "; "), durableLoser)
		}
	}

	func TestRunETLTransportStaleWriterFailureSurvivesReopen(t *testing.T) {
		assertRunETLTransportStaleWriterFinalization(t, synccontract.ModeFullAppend, false)
	}

	func TestRunETLTransportStaleWriterFinalizesAfterCancellation(t *testing.T) {
		assertRunETLTransportStaleWriterFinalization(t, synccontract.ModeFullAppend, true)
	}

	func TestRunETLTransportStaleWriterFinalizesLosingRunForPerPageAcknowledgementModes(t *testing.T) {
		for _, mode := range appTransportPerPageAcknowledgementModes() {
			t.Run(string(mode), func(t *testing.T) {
				assertRunETLTransportStaleWriterFinalization(t, mode, false)
			})
		}
	}

	func TestRunETLTransportFullOverwriteStaleWriterAfterReceiptReadBackFinalizesLosingRun(t *testing.T) {
		assertRunETLTransportStaleWriterFinalization(t, synccontract.ModeFullOverwrite, false)
	}

	func TestRunETLTransportFullOverwriteReceiptReadBackThenStaleFinalCheckpointFinalizesLosingRun(t *testing.T) {
		assertRunETLTransportStaleWriterFinalization(t, synccontract.ModeFullOverwrite, false, func(source *appTransportSourceExecutor) {
			first := source.page.CandidateCheckpoint.Clone()
			second := first.Clone()
			second.Position.Primary = synccontract.OpaqueToken{0xff, 0x01}
			second.Position.TieBreaker = synccontract.OpaqueToken{0xff, 0x01}
			source.pages = []synctransport.SourcePage{
				{Records: []connectors.Record{{"id": "loser-page-one"}}, CandidateCheckpoint: first},
				{Records: []connectors.Record{{"id": "loser-page-two"}}, CandidateCheckpoint: second},
			}
		})
	}

	func TestRunETLTransportAcknowledgedPageThenStaleSecondPageFinalizesLosingRunForPerPageAcknowledgementModes(t *testing.T) {
		for _, mode := range appTransportPerPageAcknowledgementModes() {
			t.Run(string(mode), func(t *testing.T) {
				assertRunETLTransportAcknowledgedPageThenStaleSecondPageFinalization(t, mode)
			})
		}
	}

	func assertRunETLTransportAcknowledgedPageThenStaleSecondPageFinalization(t *testing.T, mode synccontract.Mode) {
		t.Helper()
		fixture := setupAppTransportFixture(t, mode)
		first := fixture.sourceExecutor.page.CandidateCheckpoint.Clone()
		first.Position.Primary = synccontract.OpaqueToken{0xff}
		first.Position.TieBreaker = synccontract.OpaqueToken{0xff}
		second := first.Clone()
		second.Position.Primary = synccontract.OpaqueToken{0xff, 0x01}
		second.Position.TieBreaker = synccontract.OpaqueToken{0xff, 0x01}
		fixture.sourceExecutor.pages = []synctransport.SourcePage{
			{Records: []connectors.Record{{"id": "loser-page-one"}}, CandidateCheckpoint: first},
			{Records: []connectors.Record{{"id": "loser-page-two"}}, CandidateCheckpoint: second},
		}

		pageOneAcknowledged := make(chan struct{})
		releaseSecondPage := make(chan struct{})
		emittedPages := 0
		fixture.sourceExecutor.afterEmit = func() {
			emittedPages++
			if emittedPages != 1 {
				return
			}
			close(pageOneAcknowledged)
			<-releaseSecondPage
		}

		var loserRun Run
		var loserErr error
		done := make(chan struct{})
		go func() {
			loserRun, loserErr = fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
			close(done)
		}()
		waitForTransportSignal(t, pageOneAcknowledged)

		stateKey := streamStateKey(fixture.connection, "records")
		acknowledged, present := fixture.app.state.StreamStates[stateKey]
		if !present || acknowledged.Checkpoint == nil || !bytes.Equal(acknowledged.Checkpoint.Position.Primary, []byte{0xff}) {
			t.Fatalf("page-one acknowledgement = %#v, want durable first checkpoint", acknowledged)
		}
		var loserRunID string
		for _, run := range fixture.app.state.Runs {
			if run.Type == "etl" && run.Connection == fixture.connection && run.Stream == "records" {
				loserRunID = run.ID
				break
			}
		}
		if loserRunID == "" {
			t.Fatal("losing transport run is missing after page-one acknowledgement")
		}

		winner, err := Open(fixture.app.root)
		if err != nil {
			t.Fatal(err)
		}
		unrelatedUpdatedAt := time.Unix(57, 0).UTC()
		unrelatedState := StreamState{
			Connection:          "unrelated",
			Stream:              "records",
			GenerationID:        23,
			LastSuccessfulRunID: "unrelated_stale_page_run",
			RecordsLoaded:       31,
			UpdatedAt:           unrelatedUpdatedAt,
		}
		unrelatedCheckpoint := map[string]string{"cursor": "unrelated-stale-page-preserved"}
		unrelatedRun := Run{
			ID:          "unrelated_stale_page_run",
			Type:        "etl",
			Connection:  "unrelated",
			Stream:      "records",
			Status:      "completed",
			StartedAt:   unrelatedUpdatedAt.Add(-time.Second),
			CompletedAt: unrelatedUpdatedAt,
		}
		if _, err := winner.updateState(func(current state) (state, error) {
			if current.StreamStates == nil {
				current.StreamStates = map[string]StreamState{}
			}
			current.StreamStates["unrelated:records"] = cloneStreamState(unrelatedState)
			if current.Checkpoints == nil {
				current.Checkpoints = map[string]map[string]string{}
			}
			current.Checkpoints[unrelatedRun.ID] = cloneStringMap(unrelatedCheckpoint)
			current.Runs = append(current.Runs, unrelatedRun)
			return current, nil
		}); err != nil {
			t.Fatal(err)
		}

		credential, ok := winner.findCredential("source")
		if !ok {
			t.Fatal("winner app has no source credential")
		}
		winnerCheckpoint := appTransportCheckpoint(fixture.source, credential, "records")
		winnerCheckpoint.Position.Primary = synccontract.OpaqueToken{0xff, 0x00}
		winnerCheckpoint.Position.TieBreaker = synccontract.OpaqueToken{0xff, 0x00}
		winnerSource := &appTransportSourceExecutor{
			reference: fixture.sourceExecutor.reference,
			page: synctransport.SourcePage{
				Records:             []connectors.Record{{"id": "winner-page-two"}},
				CandidateCheckpoint: winnerCheckpoint,
			},
		}
		winnerDestination := &appTransportDestinationExecutor{reference: fixture.destinationExecutor.reference, sink: fixture.destination.Name()}
		fixture.configureRuntime(t, winner, winnerSource, winnerDestination)
		winnerRun, err := winner.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
		if err != nil {
			t.Fatalf("winner RunETL() = %v", err)
		}
		if winnerRun.Status != "completed" || winnerDestination.applyCalls != 1 {
			t.Fatalf("winner result=%#v applies=%d, want completed winner with one apply", winnerRun, winnerDestination.applyCalls)
		}
		winningState := cloneStreamState(winner.state.StreamStates[stateKey])
		if winningState.Checkpoint == nil || !bytes.Equal(winningState.Checkpoint.Position.Primary, []byte{0xff, 0x00}) || winningState.LastSuccessfulRunID != winnerRun.ID {
			t.Fatalf("winner stream state = %#v, want second-page winner checkpoint for run %q", winningState, winnerRun.ID)
		}

		close(releaseSecondPage)
		waitForTransportSignal(t, done)
		if !errors.Is(loserErr, errTransportStreamStateConflict) {
			t.Fatalf("loser RunETL() error = %v, want typed second-page stream-state conflict", loserErr)
		}
		if fixture.destinationExecutor.applyCalls != 2 {
			t.Fatalf("loser destination applies = %d, want page one and page two exactly once", fixture.destinationExecutor.applyCalls)
		}

		reopened, err := Open(fixture.app.root)
		if err != nil {
			t.Fatal(err)
		}
		if current, present := reopened.state.StreamStates[stateKey]; !present || !transportStreamStateEqual(current, winningState) {
			t.Fatalf("winner stream state changed during loser finalization: got %#v, want %#v", current, winningState)
		}
		if current, present := reopened.state.StreamStates["unrelated:records"]; !present || !transportStreamStateEqual(current, unrelatedState) {
			t.Fatalf("unrelated stream state changed during loser finalization: got %#v, want %#v", current, unrelatedState)
		}
		if got := reopened.state.Checkpoints[unrelatedRun.ID]; !reflect.DeepEqual(got, unrelatedCheckpoint) {
			t.Fatalf("unrelated checkpoint = %#v, want %#v", got, unrelatedCheckpoint)
		}

		var durableLoser, durableUnrelated Run
		for _, run := range reopened.state.Runs {
			switch run.ID {
			case loserRunID:
				durableLoser = run
			case unrelatedRun.ID:
				durableUnrelated = run
			}
		}
		if !reflect.DeepEqual(durableUnrelated, unrelatedRun) {
			t.Fatalf("unrelated run changed during loser finalization: got %#v, want %#v", durableUnrelated, unrelatedRun)
		}

		var symptoms []string
		if loserRun.ID != loserRunID {
			symptoms = append(symptoms, fmt.Sprintf("returned loser ID=%q, want %q", loserRun.ID, loserRunID))
		}
		if loserRun.Status != "failed" || loserRun.CompletedAt.IsZero() {
			symptoms = append(symptoms, fmt.Sprintf("returned loser=%+v, want failed terminal run", loserRun))
		}
		if durableLoser.ID != loserRunID {
			symptoms = append(symptoms, fmt.Sprintf("durable loser ID=%q, want %q", durableLoser.ID, loserRunID))
		} else {
			if durableLoser.Status != "failed" || durableLoser.CompletedAt.IsZero() || durableLoser.Error == "" {
				symptoms = append(symptoms, fmt.Sprintf("durable loser=%+v, want failed terminal run", durableLoser))
			}
			if loserRun.ID != durableLoser.ID || loserRun.Status != durableLoser.Status || !loserRun.CompletedAt.Equal(durableLoser.CompletedAt) {
				symptoms = append(symptoms, fmt.Sprintf("returned loser=%+v, durable loser=%+v", loserRun, durableLoser))
			}
		}
		if len(symptoms) > 0 {
			t.Fatalf("acknowledged page-one stale second-page finalization leak: %s", strings.Join(symptoms, "; "))
		}
	}

	func TestFailRunTransportConflictPreservesLatestConcurrentState(t *testing.T) {
		fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
		loser := Run{
			ID:         "run_loser",
			Type:       "etl",
			Connection: fixture.connection,
			Stream:     "records",
			Status:     "running",
			StartedAt:  time.Unix(10, 0).UTC(),
		}
		if _, err := fixture.app.beginRun(loser); err != nil {
			t.Fatal(err)
		}

		writer, err := Open(fixture.app.root)
		if err != nil {
			t.Fatal(err)
		}
		credential, ok := writer.findCredential("source")
		if !ok {
			t.Fatal("writer app has no source credential")
		}
		winnerCheckpoint := appTransportCheckpoint(fixture.source, credential, "records")
		winnerCheckpoint.Position.Primary = synccontract.OpaqueToken{0xff, 0x00}
		winnerCheckpoint.Position.TieBreaker = synccontract.OpaqueToken{0xff, 0x00}
		stateKey := streamStateKey(fixture.connection, "records")
		unrelatedUpdatedAt := time.Unix(11, 0).UTC()
		if _, err := writer.updateState(func(current state) (state, error) {
			if current.StreamStates == nil {
				current.StreamStates = map[string]StreamState{}
			}
			current.StreamStates[stateKey] = StreamState{
				Connection:          fixture.connection,
				Stream:              "records",
				Checkpoint:          &winnerCheckpoint,
				GenerationID:        1,
				LastSuccessfulRunID: "run_winner",
				RecordsLoaded:       1,
				UpdatedAt:           unrelatedUpdatedAt,
			}
			return current, nil
		}); err != nil {
			t.Fatal(err)
		}

		_, conflictErr := fixture.app.updateState(func(current state) (state, error) {
			if _, present := current.StreamStates[stateKey]; present {
				return current, errTransportStreamStateConflict
			}
			return current, fmt.Errorf("stale conflict fixture did not observe winner stream state")
		})
		if !errors.Is(conflictErr, errTransportStreamStateConflict) {
			t.Fatalf("stale checkpoint update error = %v, want typed transport state conflict", conflictErr)
		}

		if _, err := writer.updateState(func(current state) (state, error) {
			if current.StreamStates == nil {
				current.StreamStates = map[string]StreamState{}
			}
			current.StreamStates["unrelated:records"] = StreamState{
				Connection:          "unrelated",
				Stream:              "records",
				GenerationID:        8,
				LastSuccessfulRunID: "unrelated_run",
				RecordsLoaded:       13,
				UpdatedAt:           unrelatedUpdatedAt,
			}
			if current.Checkpoints == nil {
				current.Checkpoints = map[string]map[string]string{}
			}
			current.Checkpoints["unrelated_run"] = map[string]string{"cursor": "preserved"}
			current.Runs = append(current.Runs, Run{
				ID:          "unrelated_run",
				Type:        "etl",
				Connection:  "unrelated",
				Stream:      "records",
				Status:      "completed",
				StartedAt:   unrelatedUpdatedAt.Add(-time.Second),
				CompletedAt: unrelatedUpdatedAt,
			})
			return current, nil
		}); err != nil {
			t.Fatal(err)
		}

		returned, err := fixture.app.failRun(loser.ID, conflictErr)
		if !errors.Is(err, errTransportStreamStateConflict) {
			t.Fatalf("failRun() error = %v, want typed transport state conflict", err)
		}
		if returned.ID != loser.ID || returned.Status != "failed" || returned.CompletedAt.IsZero() {
			t.Fatalf("failRun() returned %#v, want terminal loser", returned)
		}

		reopened, err := Open(fixture.app.root)
		if err != nil {
			t.Fatal(err)
		}
		winner := reopened.state.StreamStates[stateKey]
		if winner.Checkpoint == nil || !bytes.Equal(winner.Checkpoint.Position.Primary, []byte{0xff, 0x00}) || winner.LastSuccessfulRunID != "run_winner" {
			t.Fatalf("winner state changed during typed finalization: %#v", winner)
		}
		unrelated := reopened.state.StreamStates["unrelated:records"]
		if unrelated.GenerationID != 8 || unrelated.LastSuccessfulRunID != "unrelated_run" || unrelated.RecordsLoaded != 13 || !unrelated.UpdatedAt.Equal(unrelatedUpdatedAt) {
			t.Fatalf("unrelated stream state changed: %#v", unrelated)
		}
		if got := reopened.state.Checkpoints["unrelated_run"]["cursor"]; got != "preserved" {
			t.Fatalf("unrelated checkpoint = %q, want preserved", got)
		}
		var durableLoser, unrelatedRun Run
		for _, run := range reopened.state.Runs {
			switch run.ID {
			case loser.ID:
				durableLoser = run
			case "unrelated_run":
				unrelatedRun = run
			}
		}
		if durableLoser.ID != loser.ID || durableLoser.Status != "failed" || durableLoser.CompletedAt.IsZero() {
			t.Fatalf("durable loser = %#v, want failed terminal run", durableLoser)
		}
		if unrelatedRun.Status != "completed" || !unrelatedRun.CompletedAt.Equal(unrelatedUpdatedAt) {
			t.Fatalf("unrelated run changed: %#v", unrelatedRun)
		}
	*/
}

func TestFailRunRetainsRevisionGuardWithoutTransportConflict(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	loser := Run{
		ID:         "run_loser",
		Type:       "etl",
		Connection: fixture.connection,
		Stream:     "records",
		Status:     "running",
		StartedAt:  time.Unix(10, 0).UTC(),
	}
	if _, err := fixture.app.beginRun(loser); err != nil {
		t.Fatal(err)
	}
	writer, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.updateState(func(current state) (state, error) {
		if current.Checkpoints == nil {
			current.Checkpoints = map[string]map[string]string{}
		}
		current.Checkpoints["unrelated_run"] = map[string]string{"cursor": "preserved"}
		return current, nil
	}); err != nil {
		t.Fatal(err)
	}

	returned, err := fixture.app.failRun(loser.ID, errors.New("ordinary ETL failure"))
	if !errors.Is(err, errStateRevisionConflict) {
		t.Fatalf("failRun() error = %v, want revision conflict for ordinary failure", err)
	}
	if returned.ID != "" {
		t.Fatalf("failRun() returned %#v, want no rebase for ordinary failure", returned)
	}
	reopened, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	if got := reopened.state.Checkpoints["unrelated_run"]["cursor"]; got != "preserved" {
		t.Fatalf("ordinary failure overwrote unrelated checkpoint = %q", got)
	}
	for _, run := range reopened.state.Runs {
		if run.ID == loser.ID {
			if run.Status != "running" || !run.CompletedAt.IsZero() {
				t.Fatalf("ordinary failure bypassed revision guard: %#v", run)
			}
			return
		}
	}
	t.Fatalf("ordinary failure lost original run %q", loser.ID)
}

func TestFailRunTransportConflictRequiresRunningTarget(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	terminal := Run{
		ID:          "run_terminal",
		Type:        "etl",
		Connection:  fixture.connection,
		Stream:      "records",
		Status:      "completed",
		StartedAt:   time.Unix(10, 0).UTC(),
		CompletedAt: time.Unix(11, 0).UTC(),
	}
	if _, err := fixture.app.beginRun(terminal); err != nil {
		t.Fatal(err)
	}

	returned, err := fixture.app.failRun(terminal.ID, fmt.Errorf("stale checkpoint: %w", errTransportStreamStateConflict))
	if !errors.Is(err, errTransportStreamStateConflict) {
		t.Fatalf("failRun() error = %v, want typed transport state conflict", err)
	}
	if returned.ID != terminal.ID || returned.Status != "completed" || !returned.CompletedAt.Equal(terminal.CompletedAt) {
		t.Fatalf("failRun() replaced terminal run: %#v", returned)
	}
	reopened, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	for _, run := range reopened.state.Runs {
		if run.ID == terminal.ID {
			if run.Status != "completed" || !run.CompletedAt.Equal(terminal.CompletedAt) {
				t.Fatalf("durable terminal run was replaced: %#v", run)
			}
			return
		}
	}
	t.Fatalf("terminal run %q was not retained", terminal.ID)
}

func TestFailRunTransportConflictReturnsMayHaveCommittedFinalization(t *testing.T) {
	tests := []struct {
		name        string
		configure   func(*App)
		wantOutcome statestore.CommitOutcome
	}{
		{
			name: "committed unlock failure",
			configure: func(a *App) {
				a.store.Locker = &postCommitUnlockFailureLocker{failAt: 1}
			},
			wantOutcome: statestore.CommitOutcomeCommitted,
		},
		{
			name: "indeterminate directory sync failure",
			configure: func(a *App) {
				a.store.SyncDirectory = func(string) error {
					return errTransportFinalizationStateSync
				}
			},
			wantOutcome: statestore.CommitOutcomeIndeterminate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
			loser := Run{
				ID:         "run_loser",
				Type:       "etl",
				Connection: fixture.connection,
				Stream:     "records",
				Status:     "running",
				StartedAt:  time.Unix(10, 0).UTC(),
			}
			if _, err := fixture.app.beginRun(loser); err != nil {
				t.Fatal(err)
			}
			tt.configure(fixture.app)

			returned, err := fixture.app.failRun(loser.ID, fmt.Errorf("stale checkpoint: %w", errTransportStreamStateConflict))
			if !errors.Is(err, errTransportStreamStateConflict) {
				t.Fatalf("failRun() error = %v, want typed transport state conflict", err)
			}
			var outcome *statestore.CommitOutcomeError
			if !errors.As(err, &outcome) || outcome.Outcome != tt.wantOutcome || !outcome.Outcome.MayHaveCommitted() {
				t.Fatalf("failRun() commit outcome = %#v, want %s", outcome, tt.wantOutcome)
			}
			if returned.ID != loser.ID || returned.Status != "failed" || returned.CompletedAt.IsZero() {
				t.Fatalf("failRun() returned %#v, want terminal loser", returned)
			}

			reopened, err := Open(fixture.app.root)
			if err != nil {
				t.Fatal(err)
			}
			for _, run := range reopened.state.Runs {
				if run.ID != loser.ID {
					continue
				}
				if run.Status != "failed" || run.CompletedAt.IsZero() {
					t.Fatalf("durable loser = %#v, want terminal run", run)
				}
				return
			}
			t.Fatalf("durable loser %q was not retained", loser.ID)
		})
	}
}

func TestRunETLTransportDistinguishesMissingAndPresentStreamState(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	acknowledgementReached := make(chan struct{})
	releaseAcknowledgement := make(chan struct{})
	fixture.destinationExecutor.afterApply = func() {
		close(acknowledgementReached)
		<-releaseAcknowledgement
	}

	done := make(chan struct{})
	var run Run
	var runErr error
	go func() {
		run, runErr = fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
		close(done)
	}()
	waitForTransportSignal(t, acknowledgementReached)

	writer, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	stateKey := streamStateKey(fixture.connection, "records")
	if _, err := writer.updateState(func(current state) (state, error) {
		if current.StreamStates == nil {
			current.StreamStates = map[string]StreamState{}
		}
		current.StreamStates[stateKey] = StreamState{}
		return current, nil
	}); err != nil {
		t.Fatal(err)
	}
	close(releaseAcknowledgement)
	waitForTransportSignal(t, done)

	if runErr == nil || !strings.Contains(runErr.Error(), "transport stream state changed") {
		t.Fatalf("RunETL() error = %v, want present-state conflict", runErr)
	}
	if run.Status == "completed" {
		t.Fatalf("run status = %q, must not complete", run.Status)
	}
	reopened, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	state, present := reopened.state.StreamStates[stateKey]
	if !present || state.Checkpoint != nil || state.GenerationID != 0 || state.LastSuccessfulRunID != "" || state.RecordsLoaded != 0 || !state.UpdatedAt.IsZero() {
		t.Fatalf("present zero stream state was overwritten: %#v", state)
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
	// The Group-6 durable work fence deliberately commits claim, source
	// admission, destination admission, and acknowledgement before this final
	// run-status write. Keep the failure pinned to that final write rather than
	// weakening the test into an earlier pre-I/O failure.
	fixture.app.store.Locker = &appTransportFailAtLockLocker{failAt: 6, err: errTransportFinalStateSave}

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

func TestRunETLTransportAcknowledgedCompletionRebasesUnrelatedStateForPerPageAcknowledgementModes(t *testing.T) {
	for _, mode := range appTransportPerPageAcknowledgementModes() {
		t.Run(string(mode), func(t *testing.T) {
			assertRunETLTransportAcknowledgedCompletionRebasesUnrelatedState(t, mode)
		})
	}
}

func TestRunETLTransportFullOverwriteCompletionRebasesUnrelatedStateAfterFinalCheckpoint(t *testing.T) {
	assertRunETLTransportAcknowledgedCompletionRebasesUnrelatedState(t, synccontract.ModeFullOverwrite)
}

func assertRunETLTransportAcknowledgedCompletionRebasesUnrelatedState(t *testing.T, mode synccontract.Mode) {
	t.Helper()
	paused := startPausedAcknowledgedTransportCompletion(t, mode, context.Background())
	writer, err := Open(paused.fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	unrelatedUpdatedAt := time.Unix(11, 0).UTC()
	if _, err := writer.updateState(func(current state) (state, error) {
		if current.StreamStates == nil {
			current.StreamStates = map[string]StreamState{}
		}
		current.StreamStates["unrelated:records"] = StreamState{
			Connection:          "unrelated",
			Stream:              "records",
			GenerationID:        8,
			LastSuccessfulRunID: "unrelated_run",
			RecordsLoaded:       13,
			UpdatedAt:           unrelatedUpdatedAt,
		}
		if current.Checkpoints == nil {
			current.Checkpoints = map[string]map[string]string{}
		}
		current.Checkpoints["unrelated_run"] = map[string]string{"cursor": "preserved"}
		current.Runs = append(current.Runs, Run{
			ID:          "unrelated_run",
			Type:        "etl",
			Connection:  "unrelated",
			Stream:      "records",
			Status:      "completed",
			StartedAt:   unrelatedUpdatedAt.Add(-time.Second),
			CompletedAt: unrelatedUpdatedAt,
		})
		return current, nil
	}); err != nil {
		t.Fatal(err)
	}

	paused.releaseAndWait(t)
	if paused.err != nil {
		t.Fatalf("RunETL() = %v, want acknowledged completion after unrelated post-checkpoint write", paused.err)
	}

	reopened, err := Open(paused.fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	completedState := reopened.state.StreamStates[paused.stateKey]
	if completedState.Connection != paused.acknowledged.Connection ||
		completedState.Stream != paused.acknowledged.Stream ||
		completedState.GenerationID != paused.acknowledged.GenerationID ||
		!completedState.UpdatedAt.Equal(paused.acknowledged.UpdatedAt) ||
		!transportCheckpointEqual(completedState.Checkpoint, paused.acknowledged.Checkpoint) {
		t.Fatalf("acknowledged checkpoint changed during terminal completion: got %#v, want checkpoint-bearing state %#v", completedState, paused.acknowledged)
	}
	if completedState.LastSuccessfulRunID != paused.runID || completedState.RecordsLoaded != 1 {
		t.Fatalf("completed stream metadata = %#v, want run %q with one loaded record", completedState, paused.runID)
	}
	unrelated := reopened.state.StreamStates["unrelated:records"]
	if unrelated.GenerationID != 8 || unrelated.LastSuccessfulRunID != "unrelated_run" || unrelated.RecordsLoaded != 13 || !unrelated.UpdatedAt.Equal(unrelatedUpdatedAt) {
		t.Fatalf("unrelated stream state changed: %#v", unrelated)
	}
	if got := reopened.state.Checkpoints["unrelated_run"]["cursor"]; got != "preserved" {
		t.Fatalf("unrelated checkpoint = %q, want preserved", got)
	}

	var durable, unrelatedRun Run
	for _, run := range reopened.state.Runs {
		switch run.ID {
		case paused.runID:
			durable = run
		case "unrelated_run":
			unrelatedRun = run
		}
	}
	if unrelatedRun.Status != "completed" || !unrelatedRun.CompletedAt.Equal(unrelatedUpdatedAt) {
		t.Fatalf("unrelated run changed: %#v", unrelatedRun)
	}

	var symptoms []string
	if paused.returned.ID == "" {
		symptoms = append(symptoms, "RunETL returned zero run")
	}
	if paused.returned.Status != "completed" {
		symptoms = append(symptoms, fmt.Sprintf("returned run status=%q", paused.returned.Status))
	}
	if durable.ID != paused.runID {
		symptoms = append(symptoms, fmt.Sprintf("durable run ID=%q, want %q", durable.ID, paused.runID))
	} else {
		if durable.Status != "completed" {
			symptoms = append(symptoms, fmt.Sprintf("durable run status=%q", durable.Status))
		}
		if durable.CompletedAt.IsZero() {
			symptoms = append(symptoms, "durable run completion timestamp is zero")
		}
		if paused.returned.ID != durable.ID {
			symptoms = append(symptoms, fmt.Sprintf("returned run ID=%q, durable run ID=%q", paused.returned.ID, durable.ID))
		}
	}
	if len(symptoms) > 0 {
		t.Fatalf("acknowledged completion leak: %s; durable run=%+v", strings.Join(symptoms, "; "), durable)
	}
}

type pausedAcknowledgedTransportCompletion struct {
	fixture      appTransportFixture
	stateKey     string
	acknowledged StreamState
	runID        string
	release      chan struct{}
	done         chan struct{}
	returned     Run
	err          error
}

func startPausedAcknowledgedTransportCompletion(t *testing.T, mode synccontract.Mode, ctx context.Context) *pausedAcknowledgedTransportCompletion {
	t.Helper()
	paused := &pausedAcknowledgedTransportCompletion{
		fixture: setupAppTransportFixture(t, mode),
		release: make(chan struct{}),
		done:    make(chan struct{}),
	}
	checkpointAcknowledged := make(chan struct{})
	if mode == synccontract.ModeFullOverwrite {
		// The wrapper releases the durable state lock before it pauses. A second
		// App can therefore write a real intervening revision, while the first
		// App cannot enter terminal completion until the test releases it.
		paused.fixture.app.store.Locker = &appTransportPostFinalCheckpointLocker{
			lock:    statestore.FileLock{Path: paused.fixture.app.store.Path + ".lock"},
			// Claim plus the exact Group-6 source/begin/destination/publish fences
			// precede the final checkpoint. Pause only after that checkpoint has
			// been atomically persisted and its lock released.
			pauseAt: 7,
			reached: checkpointAcknowledged,
			release: paused.release,
		}
		go func() {
			paused.returned, paused.err = paused.fixture.app.RunETL(ctx, RunETLRequest{Connection: paused.fixture.connection, Stream: "records", BatchSize: 1})
			close(paused.done)
		}()
		waitForTransportSignal(t, checkpointAcknowledged)

		observed, err := Open(paused.fixture.app.root)
		if err != nil {
			t.Fatal(err)
		}
		paused.stateKey = streamStateKey(paused.fixture.connection, "records")
		acknowledged, present := observed.state.StreamStates[paused.stateKey]
		if !present || acknowledged.Checkpoint == nil || acknowledged.Checkpoint.CommittedAt == nil {
			t.Fatalf("full-overwrite state = %#v, want durable final checkpoint before completion", acknowledged)
		}
		paused.acknowledged = cloneStreamState(acknowledged)
		for _, run := range observed.state.Runs {
			if run.Type == "etl" && run.Connection == paused.fixture.connection && run.Stream == "records" {
				paused.runID = run.ID
				break
			}
		}
		if paused.runID == "" {
			t.Fatal("full-overwrite run is missing after final checkpoint")
		}
		if paused.fixture.destinationExecutor.applyCalls != 1 || paused.fixture.destinationExecutor.publishCalls != 1 || paused.fixture.destinationExecutor.readBackCalls != 1 || paused.fixture.destinationExecutor.abortCalls != 0 {
			t.Fatalf("full-overwrite lifecycle apply/publish/read-back/abort = %d/%d/%d/%d, want 1/1/1/0", paused.fixture.destinationExecutor.applyCalls, paused.fixture.destinationExecutor.publishCalls, paused.fixture.destinationExecutor.readBackCalls, paused.fixture.destinationExecutor.abortCalls)
		}
		return paused
	}
	paused.fixture.sourceExecutor.afterEmit = func() {
		close(checkpointAcknowledged)
		<-paused.release
	}
	go func() {
		paused.returned, paused.err = paused.fixture.app.RunETL(ctx, RunETLRequest{Connection: paused.fixture.connection, Stream: "records", BatchSize: 1})
		close(paused.done)
	}()
	waitForTransportSignal(t, checkpointAcknowledged)

	paused.stateKey = streamStateKey(paused.fixture.connection, "records")
	acknowledged, present := paused.fixture.app.state.StreamStates[paused.stateKey]
	if !present || acknowledged.Checkpoint == nil || acknowledged.Checkpoint.CommittedAt == nil {
		t.Fatalf("acknowledged transport state = %#v, want durable checkpoint before completion", acknowledged)
	}
	paused.acknowledged = cloneStreamState(acknowledged)
	for _, run := range paused.fixture.app.state.Runs {
		if run.Type == "etl" && run.Connection == paused.fixture.connection && run.Stream == "records" {
			paused.runID = run.ID
			break
		}
	}
	if paused.runID == "" {
		t.Fatal("acknowledged transport run is missing")
	}
	return paused
}

func (p *pausedAcknowledgedTransportCompletion) releaseAndWait(t *testing.T) {
	t.Helper()
	close(p.release)
	waitForTransportSignal(t, p.done)
}

// appTransportPostFinalCheckpointLocker pauses only after JSONStore has
// renamed and directory-synced the replacement state file, then released its
// file lock. That creates the App terminal-completion race without a test hook
// in production code or a pause at the earlier receipt boundary.
type appTransportPostFinalCheckpointLocker struct {
	lock    statestore.FileLock
	pauseAt int
	calls   int
	reached chan<- struct{}
	release <-chan struct{}
}

func (l *appTransportPostFinalCheckpointLocker) Lock() (func() error, error) {
	unlock, err := l.lock.Lock()
	if err != nil {
		return nil, err
	}
	l.calls++
	pause := l.calls == l.pauseAt
	return func() error {
		if err := unlock(); err != nil {
			return err
		}
		if pause {
			close(l.reached)
			<-l.release
		}
		return nil
	}, nil
}

func TestRunETLTransportAcknowledgedCompletionFailsClosedWhenTargetChanges(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*state, *pausedAcknowledgedTransportCompletion)
		assert func(*testing.T, state, *pausedAcknowledgedTransportCompletion)
	}{
		{
			name: "target stream changed",
			mutate: func(current *state, paused *pausedAcknowledgedTransportCompletion) {
				changed := cloneStreamState(paused.acknowledged)
				changed.LastSuccessfulRunID = "concurrent_winner"
				changed.RecordsLoaded = 42
				current.StreamStates[paused.stateKey] = changed
			},
			assert: func(t *testing.T, reopened state, paused *pausedAcknowledgedTransportCompletion) {
				t.Helper()
				got := reopened.StreamStates[paused.stateKey]
				if got.LastSuccessfulRunID != "concurrent_winner" || got.RecordsLoaded != 42 || transportStreamStateEqual(got, paused.acknowledged) {
					t.Fatalf("changed target stream was overwritten: %#v", got)
				}
			},
		},
		{
			name: "target stream removed",
			mutate: func(current *state, paused *pausedAcknowledgedTransportCompletion) {
				delete(current.StreamStates, paused.stateKey)
			},
			assert: func(t *testing.T, reopened state, paused *pausedAcknowledgedTransportCompletion) {
				t.Helper()
				if _, present := reopened.StreamStates[paused.stateKey]; present {
					t.Fatalf("removed target stream was recreated: %#v", reopened.StreamStates[paused.stateKey])
				}
			},
		},
		{
			name: "target run terminal",
			mutate: func(current *state, paused *pausedAcknowledgedTransportCompletion) {
				for i := range current.Runs {
					if current.Runs[i].ID == paused.runID {
						current.Runs[i].Status = "completed"
						current.Runs[i].CompletedAt = time.Unix(12, 0).UTC()
						return
					}
				}
			},
			assert: func(t *testing.T, reopened state, paused *pausedAcknowledgedTransportCompletion) {
				t.Helper()
				for _, run := range reopened.Runs {
					if run.ID == paused.runID {
						if run.Status != "completed" || !run.CompletedAt.Equal(time.Unix(12, 0).UTC()) {
							t.Fatalf("terminal target run was overwritten: %#v", run)
						}
						return
					}
				}
				t.Fatalf("terminal target run %q disappeared", paused.runID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paused := startPausedAcknowledgedTransportCompletion(t, synccontract.ModeFullAppend, context.Background())
			writer, err := Open(paused.fixture.app.root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := writer.updateState(func(current state) (state, error) {
				tt.mutate(&current, paused)
				return current, nil
			}); err != nil {
				t.Fatal(err)
			}

			paused.releaseAndWait(t)
			if !errors.Is(paused.err, errStateRevisionConflict) {
				t.Fatalf("RunETL() error = %v, want detectable revision conflict", paused.err)
			}
			if !reflect.DeepEqual(paused.returned, Run{}) {
				t.Fatalf("RunETL() returned %#v, want zero run after fail-closed completion", paused.returned)
			}

			reopened, err := Open(paused.fixture.app.root)
			if err != nil {
				t.Fatal(err)
			}
			tt.assert(t, reopened.state, paused)
		})
	}
}

// full_overwrite has no context observation after its sole final checkpoint,
// so cancellation-after-acknowledgement is intentionally inapplicable here.
// Its real contract is exercised by TestRunETLTransportFullOverwriteCancellationBeforePublishAbortsWithoutCheckpoint.
func TestRunETLTransportCancellationAfterAcknowledgedCheckpointForPerPageAcknowledgementModes(t *testing.T) {
	for _, mode := range appTransportPerPageAcknowledgementModes() {
		t.Run(string(mode), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			paused := startPausedAcknowledgedTransportCompletion(t, mode, ctx)
			cancel()
			if !errors.Is(ctx.Err(), context.Canceled) {
				t.Fatalf("context error = %v, want cancellation after acknowledged checkpoint", ctx.Err())
			}
			paused.releaseAndWait(t)
			if !errors.Is(paused.err, context.Canceled) {
				t.Fatalf("RunETL() error = %v, want cancellation after acknowledged checkpoint", paused.err)
			}
			if paused.returned.ID != paused.runID || paused.returned.Status != "failed" || paused.returned.CompletedAt.IsZero() {
				t.Fatalf("RunETL() returned %#v, want failed terminal acknowledged run", paused.returned)
			}

			reopened, err := Open(paused.fixture.app.root)
			if err != nil {
				t.Fatal(err)
			}
			if current := reopened.state.StreamStates[paused.stateKey]; !transportStreamStateEqual(current, paused.acknowledged) {
				t.Fatalf("acknowledged checkpoint changed after cancellation: got %#v, want %#v", current, paused.acknowledged)
			}
			for _, run := range reopened.state.Runs {
				if run.ID != paused.runID {
					continue
				}
				if run.Status != "failed" || run.CompletedAt.IsZero() || run.Error == "" {
					t.Fatalf("durable cancelled run = %#v, want failed terminal record", run)
				}
				return
			}
			t.Fatalf("durable cancelled run %q was not retained", paused.runID)
		})
	}
}

// The same inapplicability applies to an end-to-end acknowledged failure for
// full_overwrite: manufacturing a post-final-checkpoint context check would
// decide the captain's pending product behavior rather than test it.
func TestRunETLTransportAcknowledgedFailureAfterUnrelatedRevisionForPerPageAcknowledgementModes(t *testing.T) {
	for _, mode := range appTransportPerPageAcknowledgementModes() {
		t.Run(string(mode), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			paused := startPausedAcknowledgedTransportCompletion(t, mode, ctx)
			unrelated := persistUnrelatedAcknowledgedFailureWrite(t, paused)
			cancel()
			if !errors.Is(ctx.Err(), context.Canceled) {
				t.Fatalf("context error = %v, want cancellation after unrelated post-acknowledgement revision", ctx.Err())
			}
			paused.releaseAndWait(t)
			assertAcknowledgedFailureAfterUnrelatedRevision(t, paused, unrelated, context.Canceled)
		})
	}
}

func TestRunETLTransportAcknowledgedFailurePreservesSourceError(t *testing.T) {
	paused := startPausedAcknowledgedTransportCompletion(t, synccontract.ModeFullAppend, context.Background())
	paused.fixture.sourceExecutor.errAfterPage = errTransportSourceAfterAcknowledgement
	unrelated := persistUnrelatedAcknowledgedFailureWrite(t, paused)
	paused.releaseAndWait(t)
	assertAcknowledgedFailureAfterUnrelatedRevision(t, paused, unrelated, errTransportSourceAfterAcknowledgement)
}

func TestRunETLTransportAcknowledgedFailureFailsClosedWhenTargetChanges(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*state, *pausedAcknowledgedTransportCompletion)
	}{
		{
			name: "target stream changed",
			mutate: func(current *state, paused *pausedAcknowledgedTransportCompletion) {
				changed := cloneStreamState(paused.acknowledged)
				changed.LastSuccessfulRunID = "concurrent_winner"
				changed.RecordsLoaded = 42
				current.StreamStates[paused.stateKey] = changed
			},
		},
		{
			name: "target stream removed",
			mutate: func(current *state, paused *pausedAcknowledgedTransportCompletion) {
				delete(current.StreamStates, paused.stateKey)
			},
		},
		{
			name: "target run terminal",
			mutate: func(current *state, paused *pausedAcknowledgedTransportCompletion) {
				for i := range current.Runs {
					if current.Runs[i].ID == paused.runID {
						current.Runs[i].Status = "completed"
						current.Runs[i].CompletedAt = time.Unix(43, 0).UTC()
						return
					}
				}
			},
		},
		{
			name: "target run removed",
			mutate: func(current *state, paused *pausedAcknowledgedTransportCompletion) {
				runs := make([]Run, 0, len(current.Runs))
				for _, run := range current.Runs {
					if run.ID != paused.runID {
						runs = append(runs, run)
					}
				}
				current.Runs = runs
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paused := startPausedAcknowledgedTransportCompletion(t, synccontract.ModeFullAppend, context.Background())
			paused.fixture.sourceExecutor.errAfterPage = errTransportSourceAfterAcknowledgement
			writer, err := Open(paused.fixture.app.root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := writer.updateState(func(current state) (state, error) {
				tt.mutate(&current, paused)
				return current, nil
			}); err != nil {
				t.Fatal(err)
			}
			expected := writer.state

			paused.releaseAndWait(t)
			if paused.fixture.destinationExecutor.applyCalls != 1 {
				t.Fatalf("destination apply calls = %d, want exactly one acknowledged apply", paused.fixture.destinationExecutor.applyCalls)
			}
			if !errors.Is(paused.err, errTransportSourceAfterAcknowledgement) || !errors.Is(paused.err, errStateRevisionConflict) {
				t.Fatalf("RunETL() error = %v, want original source error and typed revision conflict", paused.err)
			}
			if !reflect.DeepEqual(paused.returned, Run{}) {
				t.Fatalf("RunETL() returned %#v, want zero run after fail-closed acknowledged failure", paused.returned)
			}

			reopened, err := Open(paused.fixture.app.root)
			if err != nil {
				t.Fatal(err)
			}
			if len(reopened.state.Checkpoints) != 0 || len(expected.Checkpoints) != 0 {
				t.Fatalf("fail-closed acknowledged failure changed checkpoints: got %#v, want %#v", reopened.state.Checkpoints, expected.Checkpoints)
			}
			actual := reopened.state
			actual.Checkpoints = nil
			expected.Checkpoints = nil
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("fail-closed acknowledged failure mutated latest state: got %#v, want %#v", actual, expected)
			}
		})
	}
}

func TestRunETLTransportAcknowledgedCompletionMissingRunIsTypedConflictForPerPageAcknowledgementModes(t *testing.T) {
	for _, mode := range appTransportPerPageAcknowledgementModes() {
		t.Run(string(mode), func(t *testing.T) {
			assertRunETLTransportAcknowledgedCompletionMissingRunIsTypedConflict(t, mode)
		})
	}
}

func TestRunETLTransportFullOverwriteCompletionMissingRunIsTypedConflictAfterFinalCheckpoint(t *testing.T) {
	assertRunETLTransportAcknowledgedCompletionMissingRunIsTypedConflict(t, synccontract.ModeFullOverwrite)
}

func assertRunETLTransportAcknowledgedCompletionMissingRunIsTypedConflict(t *testing.T, mode synccontract.Mode) {
	t.Helper()
	paused := startPausedAcknowledgedTransportCompletion(t, mode, context.Background())
	writer, err := Open(paused.fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	removed := false
	if _, err := writer.updateState(func(current state) (state, error) {
		runs := make([]Run, 0, len(current.Runs))
		for _, run := range current.Runs {
			if run.ID == paused.runID {
				removed = true
				continue
			}
			runs = append(runs, run)
		}
		current.Runs = runs
		return current, nil
	}); err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Fatalf("target run %q was not removed before acknowledged completion", paused.runID)
	}
	expected := writer.state

	paused.releaseAndWait(t)
	if paused.fixture.destinationExecutor.applyCalls != 1 {
		t.Fatalf("destination apply calls = %d, want exactly one acknowledged apply", paused.fixture.destinationExecutor.applyCalls)
	}
	if mode == synccontract.ModeFullOverwrite && (paused.fixture.destinationExecutor.publishCalls != 1 || paused.fixture.destinationExecutor.readBackCalls != 1) {
		t.Fatalf("full-overwrite publish/read-back calls = %d/%d, want 1/1 before final completion", paused.fixture.destinationExecutor.publishCalls, paused.fixture.destinationExecutor.readBackCalls)
	}
	reopened, err := Open(paused.fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	if len(reopened.state.Checkpoints) != 0 || len(expected.Checkpoints) != 0 {
		t.Fatalf("missing target run changed checkpoints: got %#v, want %#v", reopened.state.Checkpoints, expected.Checkpoints)
	}
	actual := reopened.state
	actual.Checkpoints = nil
	expected.Checkpoints = nil
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("missing target run mutated reopened state: got %#v, want %#v", actual, expected)
	}
	if !errors.Is(paused.err, errStateRevisionConflict) {
		t.Fatalf("RunETL() error = %v, want typed acknowledged revision conflict after missing target run", paused.err)
	}
	if !reflect.DeepEqual(paused.returned, Run{}) {
		t.Fatalf("RunETL() returned %#v, want zero run after missing target run", paused.returned)
	}
}

type unrelatedAcknowledgedFailureWrite struct {
	stateKey   string
	runID      string
	stream     StreamState
	checkpoint map[string]string
	run        Run
}

func persistUnrelatedAcknowledgedFailureWrite(t *testing.T, paused *pausedAcknowledgedTransportCompletion) unrelatedAcknowledgedFailureWrite {
	t.Helper()
	updatedAt := time.Unix(41, 0).UTC()
	unrelated := unrelatedAcknowledgedFailureWrite{
		stateKey: "unrelated:records",
		runID:    "unrelated_failure_run",
		stream: StreamState{
			Connection:          "unrelated",
			Stream:              "records",
			GenerationID:        17,
			LastSuccessfulRunID: "unrelated_failure_run",
			RecordsLoaded:       29,
			UpdatedAt:           updatedAt,
		},
		checkpoint: map[string]string{"cursor": "unrelated-preserved"},
		run: Run{
			ID:          "unrelated_failure_run",
			Type:        "etl",
			Connection:  "unrelated",
			Stream:      "records",
			Status:      "completed",
			StartedAt:   updatedAt.Add(-time.Second),
			CompletedAt: updatedAt,
		},
	}
	writer, err := Open(paused.fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.updateState(func(current state) (state, error) {
		if current.StreamStates == nil {
			current.StreamStates = map[string]StreamState{}
		}
		current.StreamStates[unrelated.stateKey] = cloneStreamState(unrelated.stream)
		if current.Checkpoints == nil {
			current.Checkpoints = map[string]map[string]string{}
		}
		current.Checkpoints[unrelated.runID] = cloneStringMap(unrelated.checkpoint)
		current.Runs = append(current.Runs, unrelated.run)
		return current, nil
	}); err != nil {
		t.Fatal(err)
	}
	return unrelated
}

func assertAcknowledgedFailureAfterUnrelatedRevision(t *testing.T, paused *pausedAcknowledgedTransportCompletion, unrelated unrelatedAcknowledgedFailureWrite, wantErr error) {
	t.Helper()
	if !errors.Is(paused.err, wantErr) {
		t.Fatalf("RunETL() error = %v, want original post-acknowledgement error %v", paused.err, wantErr)
	}
	if paused.fixture.destinationExecutor.applyCalls != 1 {
		t.Fatalf("destination apply calls = %d, want exactly one acknowledged apply", paused.fixture.destinationExecutor.applyCalls)
	}

	reopened, err := Open(paused.fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	if current, present := reopened.state.StreamStates[paused.stateKey]; !present || !transportStreamStateEqual(current, paused.acknowledged) {
		t.Fatalf("acknowledged checkpoint changed during post-acknowledgement error: got %#v, want %#v", current, paused.acknowledged)
	}
	if current, present := reopened.state.StreamStates[unrelated.stateKey]; !present || !transportStreamStateEqual(current, unrelated.stream) {
		t.Fatalf("unrelated stream state changed during post-acknowledgement error: got %#v, want %#v", current, unrelated.stream)
	}
	if got := reopened.state.Checkpoints[unrelated.runID]; !reflect.DeepEqual(got, unrelated.checkpoint) {
		t.Fatalf("unrelated checkpoint changed during post-acknowledgement error: got %#v, want %#v", got, unrelated.checkpoint)
	}

	var durable, durableUnrelated Run
	for _, run := range reopened.state.Runs {
		switch run.ID {
		case paused.runID:
			durable = run
		case unrelated.runID:
			durableUnrelated = run
		}
	}
	if !reflect.DeepEqual(durableUnrelated, unrelated.run) {
		t.Fatalf("unrelated run changed during post-acknowledgement error: got %#v, want %#v", durableUnrelated, unrelated.run)
	}

	var symptoms []string
	if paused.returned.ID != paused.runID {
		symptoms = append(symptoms, fmt.Sprintf("returned run ID=%q, want %q", paused.returned.ID, paused.runID))
	}
	if paused.returned.Status != "failed" || paused.returned.CompletedAt.IsZero() {
		symptoms = append(symptoms, fmt.Sprintf("returned run=%+v, want failed terminal run", paused.returned))
	}
	if durable.ID != paused.runID {
		symptoms = append(symptoms, fmt.Sprintf("durable run ID=%q, want %q", durable.ID, paused.runID))
	} else {
		if durable.Status != "failed" || durable.CompletedAt.IsZero() || durable.Error == "" {
			symptoms = append(symptoms, fmt.Sprintf("durable run=%+v, want failed terminal run", durable))
		}
		if paused.returned.ID != durable.ID || paused.returned.Status != durable.Status || !paused.returned.CompletedAt.Equal(durable.CompletedAt) {
			symptoms = append(symptoms, fmt.Sprintf("returned run=%+v, durable run=%+v", paused.returned, durable))
		}
	}
	if len(symptoms) > 0 {
		t.Fatalf("acknowledged failure finalization leak: %s", strings.Join(symptoms, "; "))
	}
}

func TestRunETLTransportAcknowledgedFailureReturnsTruthfulPersistenceOutcome(t *testing.T) {
	tests := []struct {
		name         string
		configure    func(*App)
		wantOutcome  statestore.CommitOutcome
		wantTerminal bool
	}{
		{
			name: "definite pre-commit failure",
			configure: func(a *App) {
				a.store.Locker = &appTransportFailAtLockLocker{failAt: 1, err: errTransportFinalStateSave}
			},
			wantOutcome: statestore.CommitOutcomeNotCommitted,
		},
		{
			name: "committed unlock failure",
			configure: func(a *App) {
				a.store.Locker = &postCommitUnlockFailureLocker{failAt: 1}
			},
			wantOutcome:  statestore.CommitOutcomeCommitted,
			wantTerminal: true,
		},
		{
			name: "indeterminate directory sync failure",
			configure: func(a *App) {
				a.store.SyncDirectory = func(string) error {
					return errTransportFinalizationStateSync
				}
			},
			wantOutcome:  statestore.CommitOutcomeIndeterminate,
			wantTerminal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paused := startPausedAcknowledgedTransportCompletion(t, synccontract.ModeFullAppend, context.Background())
			paused.fixture.sourceExecutor.errAfterPage = errTransportSourceAfterAcknowledgement
			unrelated := persistUnrelatedAcknowledgedFailureWrite(t, paused)
			expected, err := Open(paused.fixture.app.root)
			if err != nil {
				t.Fatal(err)
			}
			tt.configure(paused.fixture.app)
			paused.releaseAndWait(t)

			if !errors.Is(paused.err, errTransportSourceAfterAcknowledgement) {
				t.Fatalf("RunETL() error = %v, want original post-acknowledgement source error", paused.err)
			}
			if tt.wantOutcome == statestore.CommitOutcomeNotCommitted {
				if !errors.Is(paused.err, errTransportFinalStateSave) {
					t.Fatalf("RunETL() error = %v, want definite failure-finalization persistence error", paused.err)
				}
				if !reflect.DeepEqual(paused.returned, Run{}) {
					t.Fatalf("RunETL() returned %#v, want zero run after definite non-commit", paused.returned)
				}
			} else {
				var outcome *statestore.CommitOutcomeError
				if !errors.As(paused.err, &outcome) || outcome.Outcome != tt.wantOutcome || !outcome.Outcome.MayHaveCommitted() {
					t.Fatalf("RunETL() commit outcome = %#v, want %s", outcome, tt.wantOutcome)
				}
				if paused.returned.ID != paused.runID || paused.returned.Status != "failed" || paused.returned.CompletedAt.IsZero() {
					t.Fatalf("RunETL() returned %#v, want durable-consistent failed run", paused.returned)
				}
			}

			reopened, err := Open(paused.fixture.app.root)
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantTerminal {
				assertAcknowledgedFailureAfterUnrelatedRevision(t, paused, unrelated, errTransportSourceAfterAcknowledgement)
				return
			}
			if !reflect.DeepEqual(reopened.state, expected.state) {
				t.Fatalf("definite failure-finalization non-commit changed durable state: got %#v, want %#v", reopened.state, expected.state)
			}
		})
	}
}

func TestRunETLTransportAcknowledgedCompletionReturnsTruthfulPersistenceOutcome(t *testing.T) {
	tests := []struct {
		name         string
		configure    func(*App)
		wantOutcome  statestore.CommitOutcome
		wantTerminal bool
	}{
		{
			name: "definite pre-commit failure",
			configure: func(a *App) {
				a.store.Locker = &appTransportFailAtLockLocker{failAt: 1, err: errTransportFinalStateSave}
			},
			wantOutcome: statestore.CommitOutcomeNotCommitted,
		},
		{
			name: "committed unlock failure",
			configure: func(a *App) {
				a.store.Locker = &postCommitUnlockFailureLocker{failAt: 1}
			},
			wantOutcome:  statestore.CommitOutcomeCommitted,
			wantTerminal: true,
		},
		{
			name: "indeterminate directory sync failure",
			configure: func(a *App) {
				a.store.SyncDirectory = func(string) error {
					return errTransportFinalizationStateSync
				}
			},
			wantOutcome:  statestore.CommitOutcomeIndeterminate,
			wantTerminal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paused := startPausedAcknowledgedTransportCompletion(t, synccontract.ModeFullAppend, context.Background())
			persistUnrelatedCompletionWrite(t, paused)
			tt.configure(paused.fixture.app)
			paused.releaseAndWait(t)

			if tt.wantOutcome == statestore.CommitOutcomeNotCommitted {
				if !errors.Is(paused.err, errTransportFinalStateSave) {
					t.Fatalf("RunETL() error = %v, want definite completion persistence failure", paused.err)
				}
				if !reflect.DeepEqual(paused.returned, Run{}) {
					t.Fatalf("RunETL() returned %#v, want zero run after definite non-commit", paused.returned)
				}
			} else {
				var outcome *statestore.CommitOutcomeError
				if !errors.As(paused.err, &outcome) || outcome.Outcome != tt.wantOutcome || !outcome.Outcome.MayHaveCommitted() {
					t.Fatalf("RunETL() commit outcome = %#v, want %s", outcome, tt.wantOutcome)
				}
				if paused.returned.ID != paused.runID || paused.returned.Status != "completed" || paused.returned.CompletedAt.IsZero() {
					t.Fatalf("RunETL() returned %#v, want durable-consistent completed run", paused.returned)
				}
			}

			reopened, err := Open(paused.fixture.app.root)
			if err != nil {
				t.Fatal(err)
			}
			if got := reopened.state.Checkpoints["unrelated_run"]["cursor"]; got != "preserved" {
				t.Fatalf("unrelated checkpoint = %q, want preserved", got)
			}
			streamState := reopened.state.StreamStates[paused.stateKey]
			var durable Run
			for _, run := range reopened.state.Runs {
				if run.ID == paused.runID {
					durable = run
					break
				}
			}
			if tt.wantTerminal {
				if durable.Status != "completed" || durable.CompletedAt.IsZero() || paused.returned.ID != durable.ID {
					t.Fatalf("durable completed run = %#v, want matching terminal result", durable)
				}
				if streamState.LastSuccessfulRunID != paused.runID || streamState.RecordsLoaded != 1 || !transportCheckpointEqual(streamState.Checkpoint, paused.acknowledged.Checkpoint) {
					t.Fatalf("durable completed stream state = %#v, want only final metadata advanced", streamState)
				}
			} else {
				if durable.Status != "running" || !durable.CompletedAt.IsZero() {
					t.Fatalf("durable run after definite non-commit = %#v, want running", durable)
				}
				if !transportStreamStateEqual(streamState, paused.acknowledged) {
					t.Fatalf("acknowledged stream state changed after definite non-commit: got %#v, want %#v", streamState, paused.acknowledged)
				}
			}
		})
	}
}

func persistUnrelatedCompletionWrite(t *testing.T, paused *pausedAcknowledgedTransportCompletion) {
	t.Helper()
	writer, err := Open(paused.fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.updateState(func(current state) (state, error) {
		if current.Checkpoints == nil {
			current.Checkpoints = map[string]map[string]string{}
		}
		current.Checkpoints["unrelated_run"] = map[string]string{"cursor": "preserved"}
		return current, nil
	}); err != nil {
		t.Fatal(err)
	}
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

func TestETLRouteSelection_PropagatesDeclaredRoutePreflightErrors(t *testing.T) {
	TestRunETLTransportPreflightRejectsMissingExecutorBeforeSourceRead(t)
}

// TestETLRouteSelection_DeclarativeSourcePreservesDeclaredDestinationPreflightError
// prevents route selection from relabeling a declared but unregistered
// destination as an absent route. The typed registry refusal is the contract;
// treating it as an ordinary ETL fallback would make a declared operation
// unreachable and hide its exact failure before I/O.
func TestETLRouteSelection_DeclarativeSourcePreservesDeclaredDestinationPreflightError(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	sourceEvidence := connectors.ConformanceEvidenceReference{Suite: "declared-route", RunID: "source"}
	destinationEvidence := connectors.ConformanceEvidenceReference{Suite: "declared-route", RunID: "destination"}
	destinationRef := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyDeclarativeAPI, ID: "unregistered_declared_destination"}
	source := &appTransportConnector{
		meta:              connectors.Metadata{Name: "declared_source", IntegrationType: "api", Capabilities: connectors.Capabilities{Read: true}},
		definitionStreams: []connectors.StreamSummary{{Name: "records"}},
		descriptor: &connectors.SyncTransportDescriptor{Source: &connectors.SourceTransportDescriptor{
			Executor: declarativeStreamSourceReference, EligibleStreams: []string{"records"}, Modes: []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery: appTransportDelivery(), Conformance: sourceEvidence,
		}},
	}
	destination := &appTransportConnector{
		meta: connectors.Metadata{Name: "declared_destination", IntegrationType: "api", Capabilities: connectors.Capabilities{Write: true}},
		descriptor: &connectors.SyncTransportDescriptor{Destination: &connectors.DestinationTransportDescriptor{
			Executor: destinationRef, EligibleActions: []string{"apply_records"}, Modes: []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery: appTransportDelivery(), Conformance: destinationEvidence, Acknowledgement: connectors.TransportAcknowledgementDurableWarehouse,
			ApplyStrategies: []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullAppend, Strategy: connectors.ApplyStrategyAppend, Action: "apply_records"}},
		}},
	}
	sourceExecutor := &appTransportSourceExecutor{reference: declarativeStreamSourceReference}
	a.transports = synctransport.NewRegistry(appTransportVerifier{})
	if err := a.transports.RegisterSource(sourceExecutor); err != nil {
		t.Fatal(err)
	}

	selected, reason, err := a.selectTransportRoute(Connection{Streams: map[string]StreamConfig{"records": {DestinationAction: "apply_records"}}}, "records", SyncMode{ContractMode: synccontract.ModeFullAppend}, source, destination)
	var unregistered *synctransport.DestinationExecutorUnregisteredError
	if selected || reason != transportRouteDeclared || !errors.As(err, &unregistered) || unregistered.Executor != destinationRef {
		t.Fatalf("selectTransportRoute() = selected=%t reason=%q err=%v, want declared typed missing-destination preflight error", selected, reason, err)
	}
	if sourceExecutor.readCalls != 0 || source.legacyReadCalls != 0 || destination.legacyWriteCalls != 0 {
		t.Fatalf("declared route preflight source/legacy reads/writes = %d/%d/%d, want zero before I/O", sourceExecutor.readCalls, source.legacyReadCalls, destination.legacyWriteCalls)
	}
}

func TestETLRouteSelection_DeclarativeSourceRejectsUnmarkedResolvedDestination(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	sourceEvidence := connectors.ConformanceEvidenceReference{Suite: "declared-route", RunID: "source"}
	destinationEvidence := connectors.ConformanceEvidenceReference{Suite: "declared-route", RunID: "destination"}
	destinationRef := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyDeclarativeAPI, ID: "unmarked_declared_destination"}
	source := &appTransportConnector{
		meta:              connectors.Metadata{Name: "declared_marker_source", IntegrationType: "api", Capabilities: connectors.Capabilities{Read: true}},
		definitionStreams: []connectors.StreamSummary{{Name: "records"}},
		descriptor: &connectors.SyncTransportDescriptor{Source: &connectors.SourceTransportDescriptor{
			Executor: declarativeStreamSourceReference, EligibleStreams: []string{"records"}, Modes: []synccontract.Mode{synccontract.ModeFullAppend}, Delivery: appTransportDelivery(), Conformance: sourceEvidence,
		}},
	}
	destination := &appTransportConnector{
		meta: connectors.Metadata{Name: "declared_marker_destination", IntegrationType: "api", Capabilities: connectors.Capabilities{Write: true}},
		descriptor: &connectors.SyncTransportDescriptor{Destination: &connectors.DestinationTransportDescriptor{
			Executor: destinationRef, EligibleActions: []string{"apply_records"}, Modes: []synccontract.Mode{synccontract.ModeFullAppend}, Delivery: appTransportDelivery(), Conformance: destinationEvidence,
			Acknowledgement: connectors.TransportAcknowledgementDurableWarehouse, ApplyStrategies: []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullAppend, Strategy: connectors.ApplyStrategyAppend, Action: "apply_records"}},
		}},
	}
	verifier := appTransportVerifier{accepted: map[appTransportConformanceKey]struct{}{
		{role: connectors.TransportRoleSource, reference: declarativeStreamSourceReference, evidence: sourceEvidence}: {},
		{role: connectors.TransportRoleDestination, reference: destinationRef, evidence: destinationEvidence}:                {},
	}}
	a.transports = synctransport.NewRegistry(verifier)
	sourceExecutor := &appTransportSourceExecutor{reference: declarativeStreamSourceReference}
	destinationExecutor := &appTransportDestinationExecutor{reference: destinationRef, sink: destination.Name()}
	if err := a.transports.RegisterSource(sourceExecutor); err != nil {
		t.Fatal(err)
	}
	if err := a.transports.RegisterDestination(destinationExecutor); err != nil {
		t.Fatal(err)
	}

	selected, reason, err := a.selectTransportRoute(Connection{Streams: map[string]StreamConfig{"records": {DestinationAction: "apply_records"}}}, "records", SyncMode{ContractMode: synccontract.ModeFullAppend}, source, destination)
	var route *synctransport.DeclaredDestinationRouteError
	if selected || reason != transportRouteDeclared || !errors.As(err, &route) || route.Executor != destinationRef {
		t.Fatalf("selectTransportRoute() = selected=%t reason=%q err=%v, want typed unmarked declared-route refusal", selected, reason, err)
	}
	if sourceExecutor.readCalls != 0 || destinationExecutor.applyCalls != 0 || destinationExecutor.readBackCalls != 0 {
		t.Fatalf("unmarked declared route source/apply/read-back = %d/%d/%d, want zero before I/O", sourceExecutor.readCalls, destinationExecutor.applyCalls, destinationExecutor.readBackCalls)
	}
}

// TestETLRouteSelection_DeclarativeSourceKeepsBoundedLocalWarehouseLegacyModes
// protects the established declaration-owned local warehouse representation.
// Only its two dedupe modes use the closed transport executor; its remaining
// executable modes stay on the bounded ordinary warehouse route.  They must
// not be mistaken for an unmarked destination or forced through a source mode
// the transport declaration did not select.
func TestETLRouteSelection_DeclarativeSourceKeepsBoundedLocalWarehouseLegacyModes(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	source, ok := a.registry.Get("github")
	if !ok || !isDeclarativeStreamTransportConnector(source) {
		t.Fatal("GitHub declaration-owned stream source is unavailable")
	}
	destination, ok := a.registry.Get("warehouse")
	if !ok || !isLocalWarehouseDestination(destination) {
		t.Fatal("closed local warehouse destination representation is unavailable")
	}
	materializer, ok := destination.(connectors.LocalWarehouseMaterializer)
	if !ok || !materializer.MaterializesLocalWarehouse() {
		t.Fatal("closed local warehouse destination is not a materializer")
	}

	for _, mode := range []synccontract.Mode{
		synccontract.ModeFullAppend,
		synccontract.ModeFullOverwrite,
		synccontract.ModeIncrementalAppend,
	} {
		t.Run(string(mode), func(t *testing.T) {
			selected, reason, err := a.selectTransportRoute(
				Connection{Streams: map[string]StreamConfig{"pull_requests": {}}},
				"pull_requests",
				SyncMode{ContractMode: mode},
				source,
				destination,
			)
			if selected || reason != transportRouteDeclarationAbsent || err != nil {
				t.Fatalf("selectTransportRoute(%q) = selected=%t reason=%q err=%v, want bounded ordinary local-warehouse route", mode, selected, reason, err)
			}
		})
	}
}

// TestRunETLTransportPostCheckpointBookkeepingPersistsDeliveredReconciliationAndRepairsWithoutReplay
// proves the App boundary preserves the durable checkpoint and provider
// receipt when only local stage retirement fails. A restart repairs from the
// recorded state and must not invoke the source or destination again.
func TestRunETLTransportPostCheckpointBookkeepingPersistsDeliveredReconciliationAndRepairsWithoutReplay(t *testing.T) {
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	stage := &deliveredReconciliationAppTransportStage{retireErr: errors.New("transient retirement failure")}
	fixture.app.transportStage = stage
	fixture.destinationExecutor.output = json.RawMessage(`{"occurrence_id":"provider-occurred-once"}`)

	first, err := fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
	var reconciliation *synctransport.DeliveredReconciliationRequiredError
	if !errors.As(err, &reconciliation) {
		t.Fatalf("first RunETL() error = %T %v, want delivered reconciliation", err, err)
	}
	if first.Status != ETLRunStatusDeliveredReconciliationRequired || first.DeliveryReconciliation == nil || !first.DeliveryReconciliation.StageRetirement {
		t.Fatalf("first RunETL() = %#v, want persisted delivered-reconciliation terminal run", first)
	}
	if len(first.DestinationResults) != 1 || string(first.DestinationResults[0]) != `{"occurrence_id":"provider-occurred-once"}` {
		t.Fatalf("first provider results = %s, want exact retained receipt", first.DestinationResults)
	}
	if stage.retireCalls != 1 || fixture.sourceExecutor.readCalls != 1 || fixture.destinationExecutor.applyCalls != 1 || fixture.destinationExecutor.readBackCalls != 1 {
		t.Fatalf("first effects retire/source/apply/read-back = %d/%d/%d/%d, want 1/1/1/1", stage.retireCalls, fixture.sourceExecutor.readCalls, fixture.destinationExecutor.applyCalls, fixture.destinationExecutor.readBackCalls)
	}

	reopened, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	fixture.configureRuntime(t, reopened, fixture.sourceExecutor, fixture.destinationExecutor)
	reopened.transportStage = stage
	repaired, err := reopened.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
	if err != nil {
		t.Fatalf("restart RunETL() = %v, want reconciliation repair", err)
	}
	if repaired.ID != first.ID || repaired.Status != "completed" || repaired.DeliveryReconciliation != nil {
		t.Fatalf("repaired RunETL() = %#v, want original completed run with reconciliation cleared", repaired)
	}
	if stage.reconciliations < 2 || fixture.sourceExecutor.readCalls != 1 || fixture.destinationExecutor.applyCalls != 1 || fixture.destinationExecutor.readBackCalls != 1 {
		t.Fatalf("restart reconciliation/source/apply/read-back = %d/%d/%d/%d, want repair with no replay", stage.reconciliations, fixture.sourceExecutor.readCalls, fixture.destinationExecutor.applyCalls, fixture.destinationExecutor.readBackCalls)
	}
}

// TestDeliveredReconciliationApprovalMarkersRepairOrFailClosedWithoutReplay
// covers the two declaration-owned post-checkpoint marker stores. The
// persisted run is built with the same acknowledged stream state and provider
// receipt the transport path returns after an effect. Recovery may mark the
// exact plan, but it never re-enters source or destination I/O; an unknown
// marker remains durably reconciliation-required instead of being downgraded
// into an ordinary ETL fallback.
func TestDeliveredReconciliationApprovalMarkersRepairOrFailClosedWithoutReplay(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		plan       *ReversePlan
		reconcile  func(string) *DeliveryReconciliation
		wantRepair bool
	}{
		{
			name: "managed target marker",
			plan: &ReversePlan{ID: "rplan_managed_marker", Mode: reversePlanModePostgresManagedTarget, Status: reversePlanStatusApprovalConsumptionUncertain},
			reconcile: func(planID string) *DeliveryReconciliation {
				return &DeliveryReconciliation{State: ETLRunStatusDeliveredReconciliationRequired, PostgresManagedTargetPlanID: planID}
			},
			wantRepair: true,
		},
		{
			name: "declarative typed destination marker",
			plan: &ReversePlan{ID: "rplan_declarative_marker", Mode: reversePlanModeDeclarativeTypedDestinationTransport, Status: reversePlanStatusApprovalConsumptionUncertain},
			reconcile: func(planID string) *DeliveryReconciliation {
				return &DeliveryReconciliation{State: ETLRunStatusDeliveredReconciliationRequired, DeclarativeTypedDestinationPlanID: planID}
			},
			wantRepair: true,
		},
		{
			name: "unknown marker stays terminal",
			reconcile: func(string) *DeliveryReconciliation {
				return &DeliveryReconciliation{State: ETLRunStatusDeliveredReconciliationRequired, DeclarativeTypedDestinationPlanID: "rplan_missing_marker"}
			},
		},
		{
			name: "corrupt reconciliation state stays terminal",
			reconcile: func(string) *DeliveryReconciliation {
				return &DeliveryReconciliation{State: "corrupt", StageRetirement: true}
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
			if testCase.plan != nil {
				if _, err := fixture.app.updateState(func(current state) (state, error) {
					current.ReversePlans = append(current.ReversePlans, *testCase.plan)
					return current, nil
				}); err != nil {
					t.Fatal(err)
				}
			}
			planID := ""
			if testCase.plan != nil {
				planID = testCase.plan.ID
			}
			runID := "run_marker_reconciliation"
			if _, err := fixture.app.beginRun(Run{ID: runID, Type: "etl", Connection: fixture.connection, Stream: "records", Status: "running", StartedAt: time.Now().UTC()}); err != nil {
				t.Fatal(err)
			}
			stateKey := streamStateKey(fixture.connection, "records")
			checkpoint := fixture.sourceExecutor.page.CandidateCheckpoint.Clone()
			leaseUntil := time.Now().UTC().Add(time.Minute)
			acknowledged := StreamState{Connection: fixture.connection, Stream: "records", GenerationID: 1, Checkpoint: &checkpoint, ActiveWorkID: runID, ActiveWorkFence: 1, ActiveWorkLeaseUntil: &leaseUntil}
			if _, err := fixture.app.updateState(func(current state) (state, error) {
				if current.StreamStates == nil {
					current.StreamStates = map[string]StreamState{}
				}
				current.StreamStates[stateKey] = cloneStreamState(acknowledged)
				return current, nil
			}); err != nil {
				t.Fatal(err)
			}
			pending := cloneStreamState(acknowledged)
			pending.ActiveWorkID = ""
			pending.ActiveWorkLeaseUntil = nil
			result := etlExecutionResult{
				Checkpoint:                map[string]string{"mode": string(synccontract.ModeFullAppend)},
				DestinationResults:        []json.RawMessage{json.RawMessage(`{"occurrence_id":"marker-provider-receipt"}`)},
				DeliveryReconciliation:    testCase.reconcile(planID),
				TransportPhaseMeasurement: &TransportPhaseMeasurement{},
				PendingStreamState:        &pendingStreamState{Key: stateKey, State: pending},
			}
			markerErr := synctransport.NewDeliveredReconciliationRequiredError(errors.New("post-checkpoint approval marker write failed"))
			delivered, err := fixture.app.failAcknowledgedTransportRun(runID, result, markerErr)
			if !errors.Is(err, markerErr) || delivered.Status != ETLRunStatusDeliveredReconciliationRequired || delivered.DeliveryReconciliation == nil || !reflect.DeepEqual(delivered.DestinationResults, result.DestinationResults) {
				t.Fatalf("marker terminal persistence run/error = %#v / %v, want durable receipt plus reconciliation", delivered, err)
			}

			repaired, repairErr := fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
			if testCase.wantRepair {
				if repairErr != nil || repaired.ID != runID || repaired.Status != "completed" || repaired.DeliveryReconciliation != nil {
					t.Fatalf("marker repair run/error = %#v / %v, want completed exact run", repaired, repairErr)
				}
				plan, err := fixture.app.GetReversePlan(planID)
				if err != nil || plan.Status != "executed" {
					t.Fatalf("marker repair plan/error = %#v / %v, want executed exact plan", plan, err)
				}
			} else {
				var reconciliation *synctransport.DeliveredReconciliationRequiredError
				if !errors.As(repairErr, &reconciliation) || repaired.ID != runID || repaired.Status != ETLRunStatusDeliveredReconciliationRequired {
					t.Fatalf("unknown marker repair run/error = %#v / %v, want retained terminal reconciliation", repaired, repairErr)
				}
			}
			if fixture.sourceExecutor.readCalls != 0 || fixture.destinationExecutor.applyCalls != 0 || fixture.destinationExecutor.readBackCalls != 0 {
				t.Fatalf("marker recovery source/apply/read-back = %d/%d/%d, want zero replay I/O", fixture.sourceExecutor.readCalls, fixture.destinationExecutor.applyCalls, fixture.destinationExecutor.readBackCalls)
			}
		})
	}
}

func TestHasDeclaredSyncTransportRequiresBothEndpoints(t *testing.T) {
	source := &appTransportConnector{
		meta:       connectors.Metadata{Name: "invalid_source", IntegrationType: "api"},
		descriptor: &connectors.SyncTransportDescriptor{},
	}
	destination := &appTransportConnector{meta: connectors.Metadata{Name: "destination", IntegrationType: "database"}}
	if hasDeclaredSyncTransport(source, destination) {
		t.Fatal("one-sided transport declaration diverted a legacy route to preflight")
	}
	destination.descriptor = &connectors.SyncTransportDescriptor{}
	if !hasDeclaredSyncTransport(source, destination) {
		t.Fatal("two-sided malformed transport declaration was treated as a legacy route instead of being routed to preflight")
	}
}

func TestAppTransportDestinationAfterApplyRunsAfterOrdinaryAcknowledgement(t *testing.T) {
	t.Run("happy path hook observes durable acknowledgement", func(t *testing.T) {
		destination := &appTransportDestinationExecutor{sink: "warehouse"}
		observedAcknowledgements := 0
		destination.afterApply = func() {
			observedAcknowledgements = destination.acknowledgementCalls
		}

		acknowledgement, err := destination.ApplyDestination(context.Background(), synctransport.DestinationApplyRequest{})
		if err != nil {
			t.Fatalf("ApplyDestination() error = %v", err)
		}
		if acknowledgement.Sink != "warehouse" {
			t.Fatalf("acknowledgement sink = %q, want warehouse", acknowledgement.Sink)
		}
		if observedAcknowledgements != 1 {
			t.Fatalf("afterApply observed %d acknowledgements, want callback after the first acknowledgement", observedAcknowledgements)
		}
	})

	t.Run("bad path apply refusal has no acknowledgement hook", func(t *testing.T) {
		applyErr := errors.New("destination apply refused")
		destination := &appTransportDestinationExecutor{sink: "warehouse", failApplyAt: 1, applyErr: applyErr}
		hookCalls := 0
		destination.afterApply = func() { hookCalls++ }

		_, err := destination.ApplyDestination(context.Background(), synctransport.DestinationApplyRequest{})
		if !errors.Is(err, applyErr) {
			t.Fatalf("ApplyDestination() error = %v, want %v", err, applyErr)
		}
		if hookCalls != 0 || destination.acknowledgementCalls != 0 {
			t.Fatalf("refused apply hook/acknowledgements = %d/%d, want 0/0", hookCalls, destination.acknowledgementCalls)
		}
	})

	t.Run("edge repeated applies keep acknowledgement ordering", func(t *testing.T) {
		destination := &appTransportDestinationExecutor{sink: "warehouse"}
		observed := []int{}
		destination.afterApply = func() {
			observed = append(observed, destination.acknowledgementCalls)
		}

		for range 2 {
			if _, err := destination.ApplyDestination(context.Background(), synctransport.DestinationApplyRequest{}); err != nil {
				t.Fatalf("ApplyDestination() error = %v", err)
			}
		}
		if !reflect.DeepEqual(observed, []int{1, 2}) {
			t.Fatalf("afterApply acknowledgement observations = %#v, want [1 2]", observed)
		}
	})
}

func TestAppTransportFullOverwriteRunLifecycleHooks(t *testing.T) {
	t.Run("happy path exposes distinct page publish and read-back boundaries", func(t *testing.T) {
		destination := &appTransportDestinationExecutor{sink: "warehouse"}
		boundaries := []string{}
		destination.afterPageApply = func() { boundaries = append(boundaries, "page") }
		destination.afterPublish = func() { boundaries = append(boundaries, "publish") }
		destination.afterReadBack = func() { boundaries = append(boundaries, "read-back") }
		run := &appTransportFullOverwriteRun{destination: destination}

		if err := run.ApplyFullOverwrite(context.Background(), synctransport.DestinationApplyRequest{}); err != nil {
			t.Fatalf("ApplyFullOverwrite() error = %v", err)
		}
		acknowledgement, err := run.PublishFullOverwrite(context.Background(), synctransport.FullOverwritePublicationRequest{})
		if err != nil {
			t.Fatalf("PublishFullOverwrite() error = %v", err)
		}
		if err := run.ReadBackFullOverwrite(context.Background(), acknowledgement); err != nil {
			t.Fatalf("ReadBackFullOverwrite() error = %v", err)
		}
		if !reflect.DeepEqual(boundaries, []string{"page", "publish", "read-back"}) {
			t.Fatalf("full-overwrite lifecycle boundaries = %#v, want page/publish/read-back", boundaries)
		}
		if destination.applyCalls != 1 || destination.publishCalls != 1 || destination.readBackCalls != 1 || destination.abortCalls != 0 {
			t.Fatalf("full-overwrite lifecycle apply/publish/read-back/abort = %d/%d/%d/%d, want 1/1/1/0", destination.applyCalls, destination.publishCalls, destination.readBackCalls, destination.abortCalls)
		}
	})

	t.Run("bad path rejects read-back before publication and wrong acknowledgement", func(t *testing.T) {
		destination := &appTransportDestinationExecutor{sink: "warehouse"}
		run := &appTransportFullOverwriteRun{destination: destination}
		if err := run.ReadBackFullOverwrite(context.Background(), synccontract.DownstreamAcknowledgement{}); err == nil || !strings.Contains(err.Error(), "before publication") {
			t.Fatalf("ReadBackFullOverwrite() error = %v, want pre-publication refusal", err)
		}
		acknowledgement, err := run.PublishFullOverwrite(context.Background(), synctransport.FullOverwritePublicationRequest{})
		if err != nil {
			t.Fatal(err)
		}
		wrongAcknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement("other", acknowledgement.AcknowledgedAt)
		if err != nil {
			t.Fatal(err)
		}
		if err := run.ReadBackFullOverwrite(context.Background(), wrongAcknowledgement); err == nil || !strings.Contains(err.Error(), "does not match") {
			t.Fatalf("ReadBackFullOverwrite() error = %v, want acknowledgement mismatch refusal", err)
		}
		if destination.readBackCalls != 0 {
			t.Fatalf("read-back calls after refused receipts = %d, want 0", destination.readBackCalls)
		}
	})

	t.Run("edge abort is idempotent and closes pre-publication lifecycle", func(t *testing.T) {
		destination := &appTransportDestinationExecutor{sink: "warehouse"}
		run := &appTransportFullOverwriteRun{destination: destination}
		if err := run.AbortFullOverwrite(context.Background()); err != nil {
			t.Fatalf("first AbortFullOverwrite() error = %v", err)
		}
		if err := run.AbortFullOverwrite(context.Background()); err != nil {
			t.Fatalf("second AbortFullOverwrite() error = %v", err)
		}
		if _, err := run.PublishFullOverwrite(context.Background(), synctransport.FullOverwritePublicationRequest{}); err == nil || !strings.Contains(err.Error(), "after abort") {
			t.Fatalf("PublishFullOverwrite() error = %v, want post-abort refusal", err)
		}
		if destination.abortCalls != 1 || destination.publishCalls != 0 {
			t.Fatalf("abort/publish calls = %d/%d, want 1/0", destination.abortCalls, destination.publishCalls)
		}
	})
}

type appTransportConnector struct {
	meta              connectors.Metadata
	descriptor        *connectors.SyncTransportDescriptor
	definitionStreams []connectors.StreamSummary
	rateLimitScope    connectors.RateLimitScopeKey
	legacyReadCalls   int
	legacyWriteCalls  int
}

func (c *appTransportConnector) Name() string                  { return c.meta.Name }
func (c *appTransportConnector) Metadata() connectors.Metadata { return c.meta }
func (c *appTransportConnector) Definition() connectors.Definition {
	return connectors.Definition{Name: c.meta.Name, DisplayName: c.meta.DisplayName, IntegrationType: c.meta.IntegrationType, Capabilities: c.meta.Capabilities, Streams: c.definitionStreams, SyncTransport: c.descriptor}
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
func (c *appTransportConnector) RateLimitParkingScope(context.Context, connectors.RuntimeConfig, string, error) (connectors.RateLimitScopeKey, error) {
	if c.rateLimitScope == "" {
		return "", errors.New("test transport rate-limit scope is not configured")
	}
	return c.rateLimitScope, nil
}

type appTransportSourceExecutor struct {
	reference    connectors.TransportExecutorReference
	page         synctransport.SourcePage
	pages        []synctransport.SourcePage
	readCalls    int
	errAfterPage error
	beforeRead   func()
	afterEmit    func()
	read         func(context.Context, synctransport.SourceRequest, func(synctransport.SourcePage) error) error
}

func (e *appTransportSourceExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return e.reference
}
func (e *appTransportSourceExecutor) ReadTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	e.readCalls++
	if err := ctx.Err(); err != nil {
		return err
	}
	if e.beforeRead != nil {
		e.beforeRead()
	}
	if e.read != nil {
		return e.read(ctx, request, emit)
	}
	pages := e.pages
	if len(pages) == 0 {
		pages = []synctransport.SourcePage{e.page}
	}
	for _, page := range pages {
		if request.RecordExtraction != nil {
			request.RecordExtraction(time.Nanosecond)
		}
		if err := emit(page); err != nil {
			return err
		}
		if e.afterEmit != nil {
			e.afterEmit()
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return e.errAfterPage
}

type appTransportDestinationExecutor struct {
	reference            connectors.TransportExecutorReference
	sink                 string
	plan                 synctransport.DestinationPlanRequest
	planCalls            int
	applyCalls           int
	acknowledgementCalls int
	publishCalls         int
	readBackCalls        int
	abortCalls           int
	afterApply           func()
	afterPageApply       func()
	afterPublish         func()
	afterReadBack        func()
	failApplyAt          int
	applyErr             error
	output               json.RawMessage
}

func (e *appTransportDestinationExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return e.reference
}
func (e *appTransportDestinationExecutor) PlanDestination(_ context.Context, request synctransport.DestinationPlanRequest) (synctransport.DestinationPlan, error) {
	e.planCalls++
	e.plan = request
	return synctransport.DestinationPlan{ApplyStrategy: request.ApplyStrategy}, nil
}
func (e *appTransportDestinationExecutor) applyDestination() error {
	e.applyCalls++
	if e.failApplyAt == e.applyCalls && e.applyErr != nil {
		return e.applyErr
	}
	if e.afterPageApply != nil {
		e.afterPageApply()
	}
	return nil
}

func (e *appTransportDestinationExecutor) ApplyDestination(_ context.Context, _ synctransport.DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	if err := e.applyDestination(); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement(e.sink, time.Now().UTC())
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if len(e.output) != 0 {
		acknowledgement, err = acknowledgement.WithOutput(e.output)
		if err != nil {
			return synccontract.DownstreamAcknowledgement{}, err
		}
	}
	e.acknowledgementCalls++
	if e.afterApply != nil {
		e.afterApply()
	}
	return acknowledgement, nil
}

func (e *appTransportDestinationExecutor) ReadBackDestination(_ context.Context, _ synctransport.DestinationReadBackRequest) error {
	e.readBackCalls++
	return nil
}

func (e *appTransportDestinationExecutor) BeginFullOverwrite(_ context.Context, request synctransport.FullOverwriteRunRequest) (synctransport.FullOverwriteRun, error) {
	if request.Mode != synccontract.ModeFullOverwrite {
		return nil, fmt.Errorf("test full-overwrite run mode = %q, want %q", request.Mode, synccontract.ModeFullOverwrite)
	}
	return &appTransportFullOverwriteRun{destination: e}, nil
}

type appTransportFullOverwriteRun struct {
	destination     *appTransportDestinationExecutor
	published       bool
	aborted         bool
	acknowledgement synccontract.DownstreamAcknowledgement
}

func (r *appTransportFullOverwriteRun) ApplyFullOverwrite(_ context.Context, _ synctransport.DestinationApplyRequest) error {
	if r.published {
		return errors.New("test full-overwrite run was applied after publication")
	}
	if r.aborted {
		return errors.New("test full-overwrite run was applied after abort")
	}
	return r.destination.applyDestination()
}

func (r *appTransportFullOverwriteRun) PublishFullOverwrite(_ context.Context, _ synctransport.FullOverwritePublicationRequest) (synccontract.DownstreamAcknowledgement, error) {
	if r.published {
		return synccontract.DownstreamAcknowledgement{}, errors.New("test full-overwrite run was published twice")
	}
	if r.aborted {
		return synccontract.DownstreamAcknowledgement{}, errors.New("test full-overwrite run was published after abort")
	}
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement(r.destination.sink, time.Now().UTC())
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	r.destination.publishCalls++
	r.published = true
	r.acknowledgement = acknowledgement
	if r.destination.afterPublish != nil {
		r.destination.afterPublish()
	}
	return acknowledgement, nil
}

func (r *appTransportFullOverwriteRun) ReadBackFullOverwrite(_ context.Context, acknowledgement synccontract.DownstreamAcknowledgement) error {
	if !r.published {
		return errors.New("test full-overwrite read-back occurred before publication")
	}
	if !reflect.DeepEqual(acknowledgement, r.acknowledgement) {
		return errors.New("test full-overwrite read-back acknowledgement does not match publication")
	}
	r.destination.readBackCalls++
	if r.destination.afterReadBack != nil {
		r.destination.afterReadBack()
	}
	return nil
}

func (r *appTransportFullOverwriteRun) AbortFullOverwrite(context.Context) error {
	if r.aborted {
		return nil
	}
	r.aborted = true
	r.destination.abortCalls++
	return nil
}

type appTransportStage struct {
	calls    int
	lastPage synctransport.SourcePage
	worksets map[string]synctransport.WarehouseWorkset
}

type reconcilingAppTransportStage struct {
	appTransportStage
	reconciliations int
	err             error
}

type deliveredReconciliationAppTransportStage struct {
	appTransportStage
	retireErr       error
	retireCalls     int
	reconciliations int
}

func (s *deliveredReconciliationAppTransportStage) Retire(context.Context, synctransport.WarehouseReceipt) error {
	s.retireCalls++
	return s.retireErr
}

func (s *deliveredReconciliationAppTransportStage) ReconcileCommitted(context.Context) error {
	s.reconciliations++
	if s.retireCalls > 0 {
		s.retireErr = nil
	}
	return nil
}

func (s *reconcilingAppTransportStage) ReconcileCommitted(context.Context) error {
	s.reconciliations++
	return s.err
}

func (s *appTransportStage) Stage(_ context.Context, request synctransport.WarehouseStageRequest) (synctransport.WarehouseReceipt, error) {
	s.calls++
	s.lastPage = request.Page
	workset := synctransport.WarehouseWorkset{ID: fmt.Sprintf("stage-%d", s.calls), Records: request.Page.Records, Tombstones: request.Page.Tombstones, CandidateCheckpoint: request.Page.CandidateCheckpoint}
	if s.worksets == nil {
		s.worksets = make(map[string]synctransport.WarehouseWorkset)
	}
	s.worksets[workset.ID] = workset
	return synctransport.WarehouseReceipt{
		ID:               workset.ID,
		Owner:            "app-test-owner",
		Generation:       1,
		Stream:           request.Stream,
		Mode:             request.Mode,
		CheckpointSHA256: "app-test-checkpoint",
		TombstonesSHA256: "app-test-tombstones",
		ManifestSHA256:   "app-test-manifest",
		ContentSHA256:    "app-test-content",
		ParquetSHA256:    "app-test-parquet",
		Records:          len(workset.Records),
		Tombstones:       len(workset.Tombstones),
	}, nil
}

func (s *appTransportStage) Reopen(_ context.Context, receipt synctransport.WarehouseReceipt) (synctransport.WarehouseWorkset, error) {
	workset, ok := s.worksets[receipt.ID]
	if !ok {
		return synctransport.WarehouseWorkset{}, fmt.Errorf("unknown app test receipt %q", receipt.ID)
	}
	return workset, nil
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
	source              *appTransportConnector
	destination         *appTransportConnector
	verifier            appTransportVerifier
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
			Executor: destinationRef, EligibleActions: actions, Modes: appTransportDestinationModes(),
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
	return appTransportFixture{
		app:                 a,
		connection:          connection,
		source:              source,
		destination:         destination,
		verifier:            verifier,
		sourceExecutor:      sourceExecutor,
		destinationExecutor: destinationExecutor,
	}
}

func (f appTransportFixture) configureRuntime(t *testing.T, a *App, sourceExecutor *appTransportSourceExecutor, destinationExecutor *appTransportDestinationExecutor) {
	t.Helper()
	registry := connectors.NewEmptyRegistry()
	registry.Register(f.source)
	registry.Register(f.destination)
	a.registry = registry
	a.transports = synctransport.NewRegistry(f.verifier)
	if err := a.transports.RegisterSource(sourceExecutor); err != nil {
		t.Fatal(err)
	}
	if err := a.transports.RegisterDestination(destinationExecutor); err != nil {
		t.Fatal(err)
	}
	a.transportStage = &appTransportStage{}
}

func appTransportStrategies() ([]string, []connectors.DestinationApplyStrategy) {
	modes := appTransportDestinationModes()
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
		}
		strategies = append(strategies, connectors.DestinationApplyStrategy{Mode: mode, Strategy: strategy, Action: action})
	}
	return actions, strategies
}

// appTransportPerPageAcknowledgementModes is deliberately separate from
// full_overwrite: those modes mint one acknowledgement per bounded page,
// whereas full_overwrite mints exactly one receipt for its whole run.
func appTransportPerPageAcknowledgementModes() []synccontract.Mode {
	modes := make([]synccontract.Mode, 0, len(appTransportDestinationModes())-1)
	for _, mode := range appTransportDestinationModes() {
		if mode != synccontract.ModeFullOverwrite {
			modes = append(modes, mode)
		}
	}
	return modes
}

func appTransportDestinationModes() []synccontract.Mode {
	modes := make([]synccontract.Mode, 0, len(synccontract.AllModes())-1)
	for _, mode := range synccontract.AllModes() {
		if mode != synccontract.ModeChangeCapture {
			modes = append(modes, mode)
		}
	}
	return modes
}

func assertInterimTransportState(t *testing.T, a *App, stateKey, wantPosition string, wantGeneration int64) {
	assertInterimTransportStateWithMetadata(t, a, stateKey, []byte(wantPosition), wantGeneration, "", 0)
}

func assertInterimTransportStateWithMetadata(t *testing.T, a *App, stateKey string, wantPosition []byte, wantGeneration int64, wantLastSuccessfulRunID string, wantRecordsLoaded int) {
	t.Helper()
	streamState, ok := a.state.StreamStates[stateKey]
	if !ok || streamState.Checkpoint == nil || streamState.Checkpoint.CommittedAt == nil {
		t.Fatalf("interim stream state = %#v, want durable checkpoint", streamState)
	}
	if !bytes.Equal(streamState.Checkpoint.Position.Primary, wantPosition) {
		t.Fatalf("interim checkpoint position = %x, want %x", streamState.Checkpoint.Position.Primary, wantPosition)
	}
	if streamState.GenerationID != wantGeneration {
		t.Fatalf("interim generation = %d, want %d", streamState.GenerationID, wantGeneration)
	}
	if streamState.LastSuccessfulRunID != wantLastSuccessfulRunID || streamState.RecordsLoaded != wantRecordsLoaded || streamState.UpdatedAt.IsZero() {
		t.Fatalf("interim stream state = %#v, want last successful run %q and records loaded %d", streamState, wantLastSuccessfulRunID, wantRecordsLoaded)
	}
}

func waitForTransportSignal(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(20 * time.Second):
		t.Fatal("timed out waiting for transport synchronization")
	}
}

type appRateParkingTestScheduler struct {
	tasks []*appRateParkingTestTask
}

type appRateParkingTestTask struct {
	at       time.Time
	callback func()
	stopped  bool
	fired    bool
}

func (s *appRateParkingTestScheduler) Schedule(at time.Time, callback func()) coordination.RateParkingTimer {
	task := &appRateParkingTestTask{at: at, callback: callback}
	s.tasks = append(s.tasks, task)
	return task
}

func (s *appRateParkingTestScheduler) Scheduled() int {
	count := 0
	for _, task := range s.tasks {
		if !task.stopped && !task.fired {
			count++
		}
	}
	return count
}

func (s *appRateParkingTestScheduler) RunThrough(now time.Time) {
	for {
		var next *appRateParkingTestTask
		for _, task := range s.tasks {
			if task.stopped || task.fired || task.at.After(now) {
				continue
			}
			if next == nil || task.at.Before(next.at) {
				next = task
			}
		}
		if next == nil {
			return
		}
		next.fired = true
		next.callback()
	}
}

func (t *appRateParkingTestTask) Stop() bool {
	if t == nil || t.stopped || t.fired {
		return false
	}
	t.stopped = true
	return true
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

type appTransportPreRenamePersistenceFailureLocker struct {
	calls       int
	failAt      int
	directory   string
	restoreMode os.FileMode
}

func (l *appTransportPreRenamePersistenceFailureLocker) Lock() (func() error, error) {
	l.calls++
	if l.calls != l.failAt {
		return func() error { return nil }, nil
	}
	if err := os.Chmod(l.directory, 0o500); err != nil {
		return nil, err
	}
	return func() error {
		return os.Chmod(l.directory, l.restoreMode)
	}, nil
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
