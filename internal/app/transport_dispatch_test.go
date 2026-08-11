package app

import (
	"bytes"
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
	fixture := setupAppTransportFixture(t, synccontract.ModeFullAppend)
	fixture.sourceExecutor.page.CandidateCheckpoint.Position.Primary = synccontract.OpaqueToken{0xff}
	fixture.sourceExecutor.page.CandidateCheckpoint.Position.TieBreaker = synccontract.OpaqueToken{0xff}
	acknowledgementReached := make(chan struct{})
	releaseAcknowledgement := make(chan struct{})
	fixture.destinationExecutor.afterApply = func() {
		close(acknowledgementReached)
		<-releaseAcknowledgement
	}

	done := make(chan struct{})
	var losingRun Run
	var losingErr error
	go func() {
		losingRun, losingErr = fixture.app.RunETL(context.Background(), RunETLRequest{Connection: fixture.connection, Stream: "records", BatchSize: 1})
		close(done)
	}()
	waitForTransportSignal(t, acknowledgementReached)

	winner, err := Open(fixture.app.root)
	if err != nil {
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
	if losingErr == nil || !strings.Contains(losingErr.Error(), "transport stream state changed") {
		t.Fatalf("losing RunETL() error = %v, want stale stream-state rejection", losingErr)
	}
	if losingRun.Status == "completed" {
		t.Fatalf("losing run status = %q, must not complete", losingRun.Status)
	}

	reopened, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	state := reopened.state.StreamStates[streamStateKey(fixture.connection, "records")]
	if state.Checkpoint == nil || !bytes.Equal(state.Checkpoint.Position.Primary, []byte{0xff, 0x00}) {
		t.Fatalf("winner checkpoint was overwritten: %#v", state.Checkpoint)
	}
	if state.LastSuccessfulRunID != winnerRun.ID {
		t.Fatalf("winner run identity = %q, want %q", state.LastSuccessfulRunID, winnerRun.ID)
	}
}

func TestRunETLTransportStaleWriterFinalizesLosingRun(t *testing.T) {
	assertRunETLTransportStaleWriterFinalization(t, synccontract.ModeFullAppend, false)
}

func assertRunETLTransportStaleWriterFinalization(t *testing.T, mode synccontract.Mode, cancelAfterAcknowledgement bool) {
	t.Helper()
	fixture := setupAppTransportFixture(t, mode)
	fixture.sourceExecutor.page.CandidateCheckpoint.Position.Primary = synccontract.OpaqueToken{0xff}
	fixture.sourceExecutor.page.CandidateCheckpoint.Position.TieBreaker = synccontract.OpaqueToken{0xff}
	acknowledgementReached := make(chan struct{})
	releaseAcknowledgement := make(chan struct{})
	fixture.destinationExecutor.afterApply = func() {
		close(acknowledgementReached)
		<-releaseAcknowledgement
	}

	done := make(chan struct{})
	var losingRun Run
	var losingErr error
	var losingCtx context.Context = context.Background()
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

func TestRunETLTransportStaleWriterFinalizesLosingRunForAllModes(t *testing.T) {
	for _, mode := range []synccontract.Mode{
		synccontract.ModeFullOverwrite,
		synccontract.ModeFullAppend,
		synccontract.ModeIncrementalAppend,
		synccontract.ModeIncrementalUpsert,
		synccontract.ModeIncrementalDedupe,
		synccontract.ModeIncrementalDedupeHistory,
		synccontract.ModeChangeCapture,
	} {
		t.Run(string(mode), func(t *testing.T) {
			assertRunETLTransportStaleWriterFinalization(t, mode, false)
		})
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
	pages        []synctransport.SourcePage
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
	pages := e.pages
	if len(pages) == 0 {
		pages = []synctransport.SourcePage{e.page}
	}
	for _, page := range pages {
		if err := emit(page); err != nil {
			return err
		}
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
