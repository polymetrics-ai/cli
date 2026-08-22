package synctransport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"

	"polymetrics.ai/internal/certificationcatalog"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
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

func TestPreflightReturnsTypedSourceStreamIneligibleErrorBeforeExecutorAccess(t *testing.T) {
	for _, stream := range []string{"not-declared", "ISSUES"} {
		t.Run(stream, func(t *testing.T) {
			pair := newTestTransportPair("api", "database")
			registry := NewRegistry(pair.verifier)
			registerTransportPair(t, registry, pair)
			stage := &testWarehouseStage{}
			commits := 0

			_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
				Source:      pair.source,
				Destination: pair.destination,
				Stream:      stream,
				Mode:        synccontract.ModeFullAppend,
				BatchSize:   10,
				Stage:       stage,
				Commit:      func(synccontract.CheckpointEnvelope) error { commits++; return nil },
			})
			var ineligible *SourceStreamIneligibleError
			if !errors.As(err, &ineligible) {
				t.Fatalf("Preflight() error = %v, want SourceStreamIneligibleError", err)
			}
			if got, want := ineligible.Connector, pair.source.Name(); got != want {
				t.Fatalf("ineligible connector = %q, want %q", got, want)
			}
			if got, want := ineligible.Stream, stream; got != want {
				t.Fatalf("ineligible stream = %q, want %q", got, want)
			}
			if pair.sourceExecutor.readCalls != 0 || stage.calls != 0 || pair.destinationExecutor.planCalls != 0 || pair.destinationExecutor.applyCalls != 0 || commits != 0 {
				t.Fatalf("allowlist refusal effects source/stage/plan/apply/checkpoint = %d/%d/%d/%d/%d, want zero", pair.sourceExecutor.readCalls, stage.calls, pair.destinationExecutor.planCalls, pair.destinationExecutor.applyCalls, commits)
			}
		})
	}
}

func TestPreflightReturnsTypedDestinationSourceIneligibleErrorBeforeExecutorAccess(t *testing.T) {
	pair := newTestTransportPair("api", "database")
	pair.destination.descriptor.Destination.SourceBindings = []connectors.DestinationSourceBinding{{
		Executor:        connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeAPI, ID: "other_declared_source"},
		EligibleStreams: []string{"records"},
		RecordMapping: connectors.SourceRecordMapping{
			Kind:        connectors.SourceRecordMappingKindConfigMatch,
			ConfigKey:   "transport_source_id",
			RecordField: "id",
		},
	}}
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)
	stage := &testWarehouseStage{}
	commits := 0

	_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		Source:      pair.source,
		Destination: pair.destination,
		Stream:      "records",
		Mode:        synccontract.ModeFullAppend,
		BatchSize:   10,
		Stage:       stage,
		Commit:      func(synccontract.CheckpointEnvelope) error { commits++; return nil },
	})
	var ineligible *DestinationSourceIneligibleError
	if !errors.As(err, &ineligible) {
		t.Fatalf("Preflight() error = %v, want DestinationSourceIneligibleError", err)
	}
	if got, want := ineligible.SourceExecutor, pair.sourceExecutor.reference; got != want {
		t.Fatalf("ineligible source executor = %+v, want %+v", got, want)
	}
	if got, want := ineligible.Destination, pair.destination.Name(); got != want {
		t.Fatalf("ineligible destination = %q, want %q", got, want)
	}
	if pair.sourceExecutor.readCalls != 0 || stage.calls != 0 || pair.destinationExecutor.planCalls != 0 || pair.destinationExecutor.applyCalls != 0 || commits != 0 {
		t.Fatalf("destination admission refusal effects source/stage/plan/apply/checkpoint = %d/%d/%d/%d/%d, want zero", pair.sourceExecutor.readCalls, stage.calls, pair.destinationExecutor.planCalls, pair.destinationExecutor.applyCalls, commits)
	}
}

func TestOrchestratorRejectsInvalidSiblingDescriptorBeforeProviderAccess(t *testing.T) {
	for _, tt := range []struct {
		name    string
		prepare func(*testTransportPair)
		want    string
	}{
		{
			name: "source connector invalid destination sibling",
			prepare: func(pair *testTransportPair) {
				pair.source.descriptor.Destination = &connectors.DestinationTransportDescriptor{}
			},
			want: "source transport descriptor",
		},
		{
			name: "destination connector invalid source sibling",
			prepare: func(pair *testTransportPair) {
				pair.destination.descriptor.Source = &connectors.SourceTransportDescriptor{}
			},
			want: "destination transport descriptor",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			pair := newTestTransportPair("api", "database")
			tt.prepare(pair)
			registry := NewRegistry(pair.verifier)
			registerTransportPair(t, registry, pair)

			_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
				Source:      pair.source,
				Destination: pair.destination,
				Stream:      "records",
				Mode:        synccontract.ModeFullAppend,
				BatchSize:   10,
				Stage:       &testWarehouseStage{},
				Commit:      func(synccontract.CheckpointEnvelope) error { return nil },
			})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Run() error = %v, want %q", err, tt.want)
			}
			if pair.sourceExecutor.readCalls != 0 || pair.destinationExecutor.planCalls != 0 || pair.destinationExecutor.applyCalls != 0 {
				t.Fatalf("provider calls source/plan/apply = %d/%d/%d, want zero before descriptor rejection", pair.sourceExecutor.readCalls, pair.destinationExecutor.planCalls, pair.destinationExecutor.applyCalls)
			}
		})
	}
}

func TestNewRegistryFailsClosedForTypedNilConformanceVerifier(t *testing.T) {
	var verifier *typedNilConformanceVerifier
	pair := newTestTransportPair("api", "database")
	registry := NewRegistry(verifier)
	registerTransportPair(t, registry, pair)

	_, err := registry.Preflight(PreflightRequest{
		Source:      pair.source,
		Destination: pair.destination,
		Stream:      "records",
		Mode:        synccontract.ModeFullAppend,
	})
	if err == nil || !strings.Contains(err.Error(), "external transport conformance verification is unavailable") {
		t.Fatalf("Preflight() error = %v, want unavailable verifier", err)
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
	pairs := map[string]struct {
		name            string
		sourceType      string
		destinationType string
	}{
		"api_to_api":           {name: "api_to_api", sourceType: "api", destinationType: "api"},
		"api_to_database":      {name: "api_to_database", sourceType: "api", destinationType: "database"},
		"database_to_api":      {name: "database_to_api", sourceType: "database", destinationType: "api"},
		"database_to_database": {name: "database_to_database", sourceType: "database", destinationType: "database"},
	}

	flowKinds := certificationcatalog.FlowKinds()
	if len(flowKinds) != len(pairs) {
		t.Fatalf("generated flow kinds = %d, want %d conformance pairs", len(flowKinds), len(pairs))
	}
	for _, flowKind := range flowKinds {
		tt, ok := pairs[flowKind.ID]
		if !ok {
			t.Fatalf("generated flow kind %q has no orchestrator conformance pair", flowKind.ID)
		}
		if got, want := tt.sourceType+"_source", flowKind.SourceRole; got != want {
			t.Fatalf("flow kind %q source role = %q, want %q", flowKind.ID, got, want)
		}
		if got, want := tt.destinationType+"_destination", flowKind.DestinationRole; got != want {
			t.Fatalf("flow kind %q destination role = %q, want %q", flowKind.ID, got, want)
		}
		t.Run(tt.name, func(t *testing.T) {
			pair := newTestTransportPair(tt.sourceType, tt.destinationType)
			registry := NewRegistry(pair.verifier)
			registerTransportPair(t, registry, pair)
			events := make([]string, 0, 7)
			pair.sourceExecutor.events = &events
			pair.destinationExecutor.events = &events
			stage := &testWarehouseStage{events: &events}
			commits := 0

			result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
				ConnectionID: "test-" + tt.name,
				Generation:   1,
				Source:       pair.source,
				Destination:  pair.destination,
				Stream:       "records",
				Mode:         synccontract.ModeFullAppend,
				BatchSize:    10,
				Stage:        stage,
				Commit: func(synccontract.CheckpointEnvelope) error {
					commits++
					events = append(events, "checkpoint")
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
			if got, want := stage.lastRequest.ConnectionID, "test-"+tt.name; got != want {
				t.Fatalf("warehouse source owner = %q, want connection-owned %q", got, want)
			}
			if got, want := stage.lastRequest.SourceName, pair.source.Name(); got != want {
				t.Fatalf("warehouse source name = %q, want %q", got, want)
			}
			if got, want := stage.lastRequest.DestinationName, pair.destination.Name(); got != want {
				t.Fatalf("warehouse destination name = %q, want %q", got, want)
			}
			if got, want := pair.destinationExecutor.lastApply.ConnectionID, stage.lastRequest.ConnectionID; got != want {
				t.Fatalf("destination apply owner = %q, want staged owner %q", got, want)
			}
			if got, want := pair.destinationExecutor.lastApply.Receipt, stage.lastReceipt; got != want {
				t.Fatalf("destination receipt = %#v, want sealed staged receipt %#v", got, want)
			}
			if got, want := pair.destinationExecutor.lastApply.Workset.ID, stage.lastReceipt.ID; got != want {
				t.Fatalf("destination workset = %q, want reopened receipt workset %q", got, want)
			}
			if got, want := pair.destinationExecutor.lastReadBack.Workset.ID, stage.lastReceipt.ID; got != want {
				t.Fatalf("destination read-back workset = %q, want reopened receipt workset %q", got, want)
			}
			wantEvents := []string{"destination-plan", "source-read", "warehouse-stage", "warehouse-reopen", "destination-apply", "destination-readback", "checkpoint"}
			if !reflect.DeepEqual(events, wantEvents) {
				t.Fatalf("warehouse-mediated execution order = %v, want %v", events, wantEvents)
			}
		})
	}
}

func TestOrchestratorFullOverwritePublishesAllBoundedPagesBeforeOneCheckpoint(t *testing.T) {
	pair := newTestTransportPair("database", "database")
	pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{
		Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "stage_replace",
	}}
	pair.destination.descriptor.Destination.EligibleActions = []string{"stage_replace"}
	first := testCheckpoint(pair.source.Name())
	second := testCheckpoint(pair.source.Name())
	second.Position.Primary = synccontract.OpaqueToken("2")
	second.Position.TieBreaker = synccontract.OpaqueToken("2")
	pair.sourceExecutor.pages = []SourcePage{
		{Records: []connectors.Record{{"id": "one"}}, CandidateCheckpoint: first},
		{Records: []connectors.Record{{"id": "two"}}, CandidateCheckpoint: second},
	}
	publishOutput := json.RawMessage(`{"provider_id":"occurrence-9007199254740993"}`)
	fullOverwrite := &testFullOverwriteRun{sink: pair.destination.Name(), publishOutput: publishOutput}
	pair.destinationExecutor.fullOverwrite = fullOverwrite
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)
	commits := 0
	var checkpoint synccontract.CheckpointEnvelope

	result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullOverwrite,
		BatchSize: 1, Stage: &testWarehouseStage{}, Commit: func(value synccontract.CheckpointEnvelope) error {
			commits++
			checkpoint = value.Clone()
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Run() = %v", err)
	}
	if got, want := fullOverwrite.ids, []string{"one", "two"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("full-overwrite staged IDs = %v, want every source page %v", got, want)
	}
	if fullOverwrite.publishCalls != 1 || fullOverwrite.readBackCalls != 1 || fullOverwrite.abortCalls != 0 {
		t.Fatalf("full-overwrite publish/read-back/abort calls = %d/%d/%d, want 1/1/0", fullOverwrite.publishCalls, fullOverwrite.readBackCalls, fullOverwrite.abortCalls)
	}
	if pair.destinationExecutor.applyCalls != 0 || commits != 1 || result.Pages != 2 || result.CommittedCheckpoint == nil || string(checkpoint.Position.Primary) != "2" {
		t.Fatalf("legacy applies/checkpoints/result = %d/%d/%+v checkpoint=%+v, want no legacy apply and one final checkpoint", pair.destinationExecutor.applyCalls, commits, result, checkpoint)
	}
	if !reflect.DeepEqual(result.DestinationResults, []json.RawMessage{publishOutput}) {
		t.Fatalf("full-overwrite destination results = %s, want publication output", result.DestinationResults)
	}
}

func TestOrchestratorFullOverwriteBudgetStopNeverPublishes(t *testing.T) {
	pair := newTestTransportPair("database", "database")
	pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{
		Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "stage_replace",
	}}
	pair.destination.descriptor.Destination.EligibleActions = []string{"stage_replace"}
	fullOverwrite := &testFullOverwriteRun{sink: pair.destination.Name()}
	pair.destinationExecutor.fullOverwrite = fullOverwrite
	source := &budgetStoppedTestSource{
		testSourceExecutor: pair.sourceExecutor,
		continuation: synccontract.SourceContinuation{
			Kind:  "engine_pagination_v1",
			Token: synccontract.OpaqueToken("opaque-engine-owned-continuation"),
		},
	}
	registry := NewRegistry(pair.verifier)
	if err := registry.RegisterSource(source); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterDestination(pair.destinationExecutor); err != nil {
		t.Fatal(err)
	}
	commits := 0

	_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullOverwrite,
		BatchSize: 1, Stage: &testWarehouseStage{}, Commit: func(synccontract.CheckpointEnvelope) error {
			commits++
			return nil
		},
	})
	var stopped *SourceBudgetStoppedError
	if !errors.As(err, &stopped) {
		t.Fatalf("Run() error = %v, want typed budget stop", err)
	}
	if fullOverwrite.publishCalls != 0 || fullOverwrite.readBackCalls != 0 || commits != 0 {
		t.Fatalf("publish/read-back/checkpoints = %d/%d/%d, want 0/0/0 after capped source", fullOverwrite.publishCalls, fullOverwrite.readBackCalls, commits)
	}
	if fullOverwrite.abortCalls != 1 {
		t.Fatalf("abort calls = %d, want private shadow abort", fullOverwrite.abortCalls)
	}
}

func TestFullOverwrite_PublishSuccessReadbackFailureDoesNotAbort(t *testing.T) {
	t.Run("standard", func(t *testing.T) {
		pair := newTestTransportPair("database", "database")
		pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
		pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
		pair.destination.descriptor.Destination.EligibleActions = []string{"stage_replace"}
		pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "stage_replace"}}
		run := &testFullOverwriteRun{sink: pair.destination.Name(), readBackErr: errors.New("provider read-back unavailable")}
		pair.destinationExecutor.fullOverwrite = run
		registry := NewRegistry(pair.verifier)
		registerTransportPair(t, registry, pair)
		_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullOverwrite, BatchSize: 1, Stage: &testWarehouseStage{}, Commit: func(synccontract.CheckpointEnvelope) error { return nil }})
		if err == nil || run.publishCalls != 1 || run.readBackCalls != 1 || run.abortCalls != 0 {
			t.Fatalf("err/publish/readback/abort = %v/%d/%d/%d, want read-back error after one publication and no abort", err, run.publishCalls, run.readBackCalls, run.abortCalls)
		}
	})

	for _, pipelineDepth := range []int{1, 2} {
		t.Run(fmt.Sprintf("arrow-depth-%d", pipelineDepth), func(t *testing.T) {
			pair := newTestTransportPair("database", "database")
			pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
			pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
			pair.destination.descriptor.Destination.EligibleActions = []string{"stage_replace"}
			pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "stage_replace"}}
			if pipelineDepth > 1 {
				pair.source.descriptor.Source.OrderedPipeline = true
				pair.destination.descriptor.Destination.OrderedPipeline = true
			}
			checkpoint := testCheckpoint(pair.source.Name())
			record := testArrowRecord(t, []int64{1}, []string{"open"})
			source := &testArrowSource{testSourceExecutor: &testSourceExecutor{reference: pair.sourceExecutor.reference}, batches: []ArrowSourceBatch{{Record: record, SourceLogicalBytes: 32, SourceRows: 1, CandidateCheckpoint: checkpoint}}}
			run := &testArrowFullOverwriteRun{sink: pair.destination.Name(), readBackErr: errors.New("provider read-back unavailable")}
			destination := &testArrowDestination{testDestinationExecutor: &testDestinationExecutor{reference: pair.destinationExecutor.reference, sink: pair.destination.Name()}, run: run}
			registry := NewRegistry(pair.verifier)
			if err := registry.RegisterSource(source); err != nil {
				t.Fatal(err)
			}
			if err := registry.RegisterDestination(destination); err != nil {
				t.Fatal(err)
			}
			plan := `{"version":1,"select":[{"source":"id","target":"id","type":"int64"}]}`
			hash, err := databaseTransformHash(plan)
			if err != nil {
				t.Fatal(err)
			}
			_, err = NewOrchestrator(registry).Run(context.Background(), RunRequest{Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullOverwrite, BatchSize: 1, MaxInFlightBatches: pipelineDepth, TransformPlanJSON: plan, TransformPlanHash: hash, FastSegments: &testFastSegmentStore{}, ByteCreditCapacity: 64, Commit: func(synccontract.CheckpointEnvelope) error { return nil }})
			if err == nil || run.publishCalls != 1 || run.readBackCalls != 1 || run.abortCalls != 0 {
				t.Fatalf("err/publish/readback/abort = %v/%d/%d/%d, want read-back error after one publication and no abort", err, run.publishCalls, run.readBackCalls, run.abortCalls)
			}
		})
	}
}

func TestTransport_ReadBackGetsIndependentUnitDeadline(t *testing.T) {
	const (
		unitDeadline  = 50 * time.Millisecond
		applyDelay    = 40 * time.Millisecond
		readBackDelay = 20 * time.Millisecond
	)

	delay := func(duration time.Duration) func(context.Context, int) error {
		return func(ctx context.Context, _ int) error {
			select {
			case <-time.After(duration):
				return ctx.Err()
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	configureFullOverwrite := func(pair *testTransportPair) {
		pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
		pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
		pair.destination.descriptor.Destination.EligibleActions = []string{"stage_replace"}
		pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{
			Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "stage_replace",
		}}
	}
	assertIndependentMetrics := func(t *testing.T, apply, readBack time.Duration) {
		t.Helper()
		if apply < applyDelay || apply >= unitDeadline {
			t.Fatalf("effect elapsed = %s, want its own %s bounded phase", apply, unitDeadline)
		}
		if readBack < readBackDelay || readBack >= unitDeadline {
			t.Fatalf("read-back elapsed = %s, want its own %s bounded phase", readBack, unitDeadline)
		}
	}

	t.Run("ordinary", func(t *testing.T) {
		pair := newTestTransportPair("api", "database")
		pair.destinationExecutor.applyContext = delay(applyDelay)
		pair.destinationExecutor.readBackContext = delay(readBackDelay)
		registry := NewRegistry(pair.verifier)
		registerTransportPair(t, registry, pair)

		result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
			Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
			BatchSize: 1, UnitDeadline: unitDeadline, Stage: &testWarehouseStage{}, Commit: func(synccontract.CheckpointEnvelope) error { return nil },
		})
		if err != nil {
			t.Fatalf("Run() error = %v, want independent apply/read-back unit deadlines", err)
		}
		assertIndependentMetrics(t, result.ApplyElapsed, result.ReadBackElapsed)
	})

	t.Run("full-overwrite", func(t *testing.T) {
		pair := newTestTransportPair("database", "database")
		configureFullOverwrite(pair)
		run := &testFullOverwriteRun{sink: pair.destination.Name(), publishContext: delay(applyDelay), readBackContext: delay(readBackDelay)}
		pair.destinationExecutor.fullOverwrite = run
		registry := NewRegistry(pair.verifier)
		registerTransportPair(t, registry, pair)

		result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
			Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullOverwrite,
			BatchSize: 1, UnitDeadline: unitDeadline, Stage: &testWarehouseStage{}, Commit: func(synccontract.CheckpointEnvelope) error { return nil },
		})
		if err != nil {
			t.Fatalf("Run() error = %v, want independent publication/read-back unit deadlines", err)
		}
		// Non-Arrow full overwrite retains its legacy apply bucket for the
		// externally visible publication; read-back must still be separate.
		assertIndependentMetrics(t, result.ApplyElapsed, result.ReadBackElapsed)
	})

	for _, maxInFlight := range []int{1, 2} {
		t.Run(fmt.Sprintf("arrow-max-in-flight-%d", maxInFlight), func(t *testing.T) {
			pair := newTestTransportPair("database", "database")
			configureFullOverwrite(pair)
			if maxInFlight > 1 {
				pair.source.descriptor.Source.OrderedPipeline = true
				pair.destination.descriptor.Destination.OrderedPipeline = true
			}
			source := &testArrowSource{testSourceExecutor: &testSourceExecutor{reference: pair.sourceExecutor.reference}, batches: []ArrowSourceBatch{{
				Record: testArrowRecord(t, []int64{1}, []string{"open"}), SourceLogicalBytes: 32, SourceRows: 1, CandidateCheckpoint: testCheckpoint(pair.source.Name()),
			}}}
			run := &testArrowFullOverwriteRun{sink: pair.destination.Name(), publishContext: delay(applyDelay), readBackContext: delay(readBackDelay)}
			destination := &testArrowDestination{testDestinationExecutor: &testDestinationExecutor{reference: pair.destinationExecutor.reference, sink: pair.destination.Name()}, run: run}
			registry := NewRegistry(pair.verifier)
			if err := registry.RegisterSource(source); err != nil {
				t.Fatal(err)
			}
			if err := registry.RegisterDestination(destination); err != nil {
				t.Fatal(err)
			}
			plan := `{"version":1,"select":[{"source":"id","target":"id","type":"int64"}]}`
			hash, err := databaseTransformHash(plan)
			if err != nil {
				t.Fatal(err)
			}

			result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
				Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullOverwrite,
				BatchSize: 1, MaxInFlightBatches: maxInFlight, UnitDeadline: unitDeadline, TransformPlanJSON: plan, TransformPlanHash: hash,
				FastSegments: &testFastSegmentStore{}, ByteCreditCapacity: 64, Commit: func(synccontract.CheckpointEnvelope) error { return nil },
			})
			if err != nil {
				t.Fatalf("Run() error = %v, want independent Arrow publication/read-back unit deadlines", err)
			}
			assertIndependentMetrics(t, result.PublishElapsed, result.ReadBackElapsed)
		})
	}

	t.Run("apply-unit-deadline-remains-strict", func(t *testing.T) {
		pair := newTestTransportPair("api", "database")
		pair.destinationExecutor.applyContext = delay(unitDeadline + 10*time.Millisecond)
		registry := NewRegistry(pair.verifier)
		registerTransportPair(t, registry, pair)
		commits := 0

		_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
			Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
			BatchSize: 1, UnitDeadline: unitDeadline, Stage: &testWarehouseStage{}, Commit: func(synccontract.CheckpointEnvelope) error { commits++; return nil },
		})
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Run() error = %v, want strict apply unit deadline", err)
		}
		if pair.destinationExecutor.readBackCalls != 0 || commits != 0 {
			t.Fatalf("read-back/checkpoints = %d/%d, want 0/0 after expired apply", pair.destinationExecutor.readBackCalls, commits)
		}
	})

	t.Run("parent-cancellation-reaches-read-back", func(t *testing.T) {
		pair := newTestTransportPair("api", "database")
		readBackStarted := make(chan struct{}, 1)
		pair.destinationExecutor.readBackContext = func(ctx context.Context, _ int) error {
			readBackStarted <- struct{}{}
			<-ctx.Done()
			return ctx.Err()
		}
		registry := NewRegistry(pair.verifier)
		registerTransportPair(t, registry, pair)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan error, 1)
		go func() {
			_, err := NewOrchestrator(registry).Run(ctx, RunRequest{
				Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
				BatchSize: 1, UnitDeadline: unitDeadline, Stage: &testWarehouseStage{}, Commit: func(synccontract.CheckpointEnvelope) error { return nil },
			})
			done <- err
		}()
		<-readBackStarted
		cancel()
		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Fatalf("Run() error = %v, want parent cancellation during read-back", err)
		}
	})
}

func TestTransportPreflight_EnforcesDeliveryGuaranteeCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		mode     synccontract.Mode
		strategy connectors.ApplyStrategy
		mutate   func(*testTransportPair)
	}{
		{name: "replayable source idempotency", mode: synccontract.ModeFullAppend, strategy: connectors.ApplyStrategyAppend, mutate: func(pair *testTransportPair) {
			pair.source.descriptor.Source.Delivery.Idempotency = connectors.DeliveryIdempotencyNone
		}},
		{name: "replayable destination idempotency", mode: synccontract.ModeIncrementalUpsert, strategy: connectors.ApplyStrategyMerge, mutate: func(pair *testTransportPair) {
			pair.destination.descriptor.Destination.Delivery.Idempotency = connectors.DeliveryIdempotencyAtLeastOnce
		}},
		{name: "dedupe ordering", mode: synccontract.ModeIncrementalDedupe, strategy: connectors.ApplyStrategyDedupe, mutate: func(pair *testTransportPair) {
			pair.destination.descriptor.Destination.Delivery.Ordering = connectors.DeliveryOrderingUnordered
		}},
		{name: "change capture tombstones", mode: synccontract.ModeChangeCapture, strategy: connectors.ApplyStrategyChangeApply, mutate: func(pair *testTransportPair) {
			pair.destination.descriptor.Destination.Delivery.Deletes = connectors.DeliveryDeletesUnavailable
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			pair := newTestTransportPair("database", "database")
			pair.source.descriptor.Source.Modes = []synccontract.Mode{testCase.mode}
			pair.destination.descriptor.Destination.Modes = []synccontract.Mode{testCase.mode}
			pair.destination.descriptor.Destination.EligibleActions = []string{"apply"}
			pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{Mode: testCase.mode, Strategy: testCase.strategy, Action: "apply"}}
			testCase.mutate(pair)
			registry := NewRegistry(pair.verifier)
			registerTransportPair(t, registry, pair)
			if _, err := registry.Preflight(PreflightRequest{Source: pair.source, Destination: pair.destination, Stream: "records", Mode: testCase.mode, DestinationAction: "apply"}); err == nil {
				t.Fatal("incompatible declared delivery guarantees passed preflight")
			}
		})
	}
}

func TestTransportRuntime_RejectsUndeclaredTombstoneBeforeIO(t *testing.T) {
	pair := newTestTransportPair("database", "database")
	pair.source.descriptor.Source.Delivery.Deletes = connectors.DeliveryDeletesTombstone
	pair.destination.descriptor.Destination.Delivery.Deletes = connectors.DeliveryDeletesUnavailable
	page := pair.sourceExecutor.pages[0]
	page.Tombstones = []synccontract.Tombstone{{Operation: synccontract.OperationDelete, EventID: synccontract.OpaqueToken("event-1"), Key: json.RawMessage(`{"id":"1"}`), DeleteImage: synccontract.DeleteImageKeyOnly, Position: page.CandidateCheckpoint.Position.Clone()}}
	pair.sourceExecutor.pages = []SourcePage{page}
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)
	stage := &testWarehouseStage{}
	_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend, BatchSize: 1, Stage: stage, Commit: func(synccontract.CheckpointEnvelope) error { return nil }})
	if err == nil || stage.calls != 0 || pair.destinationExecutor.applyCalls != 0 {
		t.Fatalf("err/stage/apply = %v/%d/%d, want tombstone refusal before stage and provider I/O", err, stage.calls, pair.destinationExecutor.applyCalls)
	}
}

func TestOrchestratorSourceCheckpointFollowsRefreshSemantics(t *testing.T) {
	tests := []struct {
		mode           synccontract.Mode
		strategy       connectors.ApplyStrategy
		wantCheckpoint bool
	}{
		{mode: synccontract.ModeFullAppend, strategy: connectors.ApplyStrategyAppend},
		{mode: synccontract.ModeFullOverwrite, strategy: connectors.ApplyStrategyReplace},
		{mode: synccontract.ModeIncrementalAppend, strategy: connectors.ApplyStrategyAppend, wantCheckpoint: true},
		{mode: synccontract.ModeIncrementalDedupe, strategy: connectors.ApplyStrategyDedupe, wantCheckpoint: true},
		{mode: synccontract.ModeIncrementalDedupeHistory, strategy: connectors.ApplyStrategyDedupeHistory, wantCheckpoint: true},
		{mode: synccontract.ModeIncrementalUpsert, strategy: connectors.ApplyStrategyMerge, wantCheckpoint: true},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			pair := newTestTransportPair("database", "database")
			action := "stage_" + string(tt.strategy)
			pair.source.descriptor.Source.Modes = []synccontract.Mode{tt.mode}
			pair.destination.descriptor.Destination.Modes = []synccontract.Mode{tt.mode}
			pair.destination.descriptor.Destination.EligibleActions = []string{action}
			pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{
				Mode: tt.mode, Strategy: tt.strategy, Action: action,
			}}
			if tt.mode == synccontract.ModeFullOverwrite {
				pair.destinationExecutor.fullOverwrite = &testFullOverwriteRun{sink: pair.destination.Name()}
			}
			registry := NewRegistry(pair.verifier)
			registerTransportPair(t, registry, pair)
			prior := testCheckpoint(pair.source.Name())

			_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
				Source: pair.source, Destination: pair.destination, Stream: "records", Mode: tt.mode,
				BatchSize: 1, Stage: &testWarehouseStage{}, Checkpoint: &prior,
				Commit: func(synccontract.CheckpointEnvelope) error { return nil },
			})
			if err != nil {
				t.Fatalf("Run() = %v", err)
			}
			got := pair.sourceExecutor.lastRequest.Checkpoint
			if tt.wantCheckpoint {
				if !reflect.DeepEqual(got, &prior) {
					t.Fatalf("source checkpoint = %+v, want prior checkpoint %+v for incremental mode", got, prior)
				}
				return
			}
			if got != nil {
				t.Fatalf("source checkpoint = %+v, want nil for full-refresh mode", got)
			}
		})
	}
}

func TestOrchestratorArrowFullOverwriteTransformsDurablyBulkAppliesThenCheckpoints(t *testing.T) {
	pair := newTestTransportPair("database", "database")
	pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "stage_replace"}}
	pair.destination.descriptor.Destination.EligibleActions = []string{"stage_replace"}
	checkpoint := testCheckpoint(pair.source.Name())
	record := testArrowRecord(t, []int64{1, 2}, []string{"open", "ignored"})
	source := &testArrowSource{testSourceExecutor: &testSourceExecutor{reference: pair.sourceExecutor.reference}, batches: []ArrowSourceBatch{{Record: record, SourceLogicalBytes: 64, SourceRows: 2, CandidateCheckpoint: checkpoint}}}
	publishOutput := json.RawMessage(`{"provider_id":"arrow-occurrence-9007199254740993"}`)
	destination := &testArrowDestination{testDestinationExecutor: &testDestinationExecutor{reference: pair.destinationExecutor.reference, sink: pair.destination.Name()}, run: &testArrowFullOverwriteRun{sink: pair.destination.Name(), publishOutput: publishOutput}}
	registry := NewRegistry(pair.verifier)
	if err := registry.RegisterSource(source); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterDestination(destination); err != nil {
		t.Fatal(err)
	}
	plan := `{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"},{"expr":{"upper":"status"},"target":"status","type":"string"}],"where":{"not_equal":[{"mod":["id",2]},0]}}`
	hash, err := databaseTransformHash(plan)
	if err != nil {
		t.Fatal(err)
	}
	commits := 0
	result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		ConnectionID: "test-connection-owner", Generation: 1, Source: pair.source, Destination: pair.destination,
		DestinationBinding: DestinationBinding{WorkspaceID: "workspace", SourceConnectorID: pair.source.Name(), ConnectionID: "test-connection-owner", StreamID: "stream", PrimaryKey: []string{"id"}},
		Stream:             "records", Mode: synccontract.ModeFullOverwrite, BatchSize: 2, TransformPlanJSON: plan, TransformPlanHash: hash,
		FastSegments: &testFastSegmentStore{}, ByteCreditCapacity: 128,
		Commit: func(synccontract.CheckpointEnvelope) error { commits++; return nil },
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got, want := destination.run.ids, []int64{1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Arrow bulk applied IDs = %v, want transformed/filter result %v", got, want)
	}
	if destination.run.publishCalls != 1 || destination.run.readBackCalls != 1 || destination.run.abortCalls != 0 || commits != 1 {
		t.Fatalf("publish/readback/abort/checkpoint = %d/%d/%d/%d, want 1/1/0/1", destination.run.publishCalls, destination.run.readBackCalls, destination.run.abortCalls, commits)
	}
	if result.RecordsRead != 2 || result.RecordsApplied != 1 || result.SourceLogicalBytes != 64 || result.ParquetBytes == 0 || result.PeakCreditBytes != 64 || result.CommittedCheckpoint == nil {
		t.Fatalf("Arrow result = %#v, want source/transform/Parquet/COPY/checkpoint counters", result)
	}
	if !reflect.DeepEqual(result.DestinationResults, []json.RawMessage{publishOutput}) {
		t.Fatalf("Arrow destination results = %s, want publication output", result.DestinationResults)
	}
}

func TestOrchestratorArrowFullOverwriteRevalidatesAuthorizationBeforeEveryApply(t *testing.T) {
	for _, maxInFlight := range []int{1, 2} {
		t.Run(fmt.Sprintf("max-in-flight-%d", maxInFlight), func(t *testing.T) {
			pair := newTestTransportPair("database", "database")
			pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
			pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
			pair.destination.descriptor.Destination.EligibleActions = []string{"stage_replace"}
			pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "stage_replace"}}
			if maxInFlight > 1 {
				pair.source.descriptor.Source.OrderedPipeline = true
				pair.destination.descriptor.Destination.OrderedPipeline = true
			}
			record := testArrowRecord(t, []int64{1}, []string{"open"})
			source := &testArrowSource{testSourceExecutor: &testSourceExecutor{reference: pair.sourceExecutor.reference}, batches: []ArrowSourceBatch{{Record: record, SourceLogicalBytes: 32, SourceRows: 1, CandidateCheckpoint: testCheckpoint(pair.source.Name())}}}
			run := &testArrowFullOverwriteRun{sink: pair.destination.Name()}
			destination := &testArrowDestination{testDestinationExecutor: &testDestinationExecutor{reference: pair.destinationExecutor.reference, sink: pair.destination.Name()}, run: run}
			registry := NewRegistry(pair.verifier)
			if err := registry.RegisterSource(source); err != nil {
				t.Fatal(err)
			}
			if err := registry.RegisterDestination(destination); err != nil {
				t.Fatal(err)
			}
			plan := `{"version":1,"select":[{"source":"id","target":"id","type":"int64"}]}`
			hash, err := databaseTransformHash(plan)
			if err != nil {
				t.Fatal(err)
			}
			errAuthorizationRevoked := errors.New("authorization revoked after Arrow segment staging")
			revoked := false
			_, err = NewOrchestrator(registry).Run(context.Background(), RunRequest{
				Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullOverwrite,
				BatchSize: 1, MaxInFlightBatches: maxInFlight, TransformPlanJSON: plan, TransformPlanHash: hash,
				FastSegments: &testFastSegmentStore{afterStore: func() { revoked = true }}, ByteCreditCapacity: 64,
				Approval: DestinationApproval{AuthorizeNextUnit: func(context.Context) error {
					if revoked {
						return errAuthorizationRevoked
					}
					return nil
				}},
				Commit: func(synccontract.CheckpointEnvelope) error { return nil },
			})
			if !errors.Is(err, errAuthorizationRevoked) {
				t.Fatalf("Run() error = %T %v, want authorization revocation after Arrow staging", err, err)
			}
			if run.applyCalls != 0 || run.publishCalls != 0 {
				t.Fatalf("Arrow apply/publication calls = %d/%d, want 0 after revocation", run.applyCalls, run.publishCalls)
			}
		})
	}
}

func TestOrchestratorArrowFullOverwriteOverlapsNextExtractionWithPreviousCopy(t *testing.T) {
	pair := newTestTransportPair("database", "database")
	pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.source.descriptor.Source.OrderedPipeline = true
	pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.OrderedPipeline = true
	pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "stage_replace"}}
	pair.destination.descriptor.Destination.EligibleActions = []string{"stage_replace"}
	first := testCheckpoint(pair.source.Name())
	second := testCheckpoint(pair.source.Name())
	third := testCheckpoint(pair.source.Name())
	second.Position.Primary = synccontract.OpaqueToken("2")
	second.Position.TieBreaker = synccontract.OpaqueToken("2")
	third.Position.Primary = synccontract.OpaqueToken("3")
	third.Position.TieBreaker = synccontract.OpaqueToken("3")
	secondExtracted := make(chan struct{})
	thirdExtracted := make(chan struct{})
	source := &testArrowSource{
		testSourceExecutor: &testSourceExecutor{reference: pair.sourceExecutor.reference},
		batches: []ArrowSourceBatch{
			{Record: testArrowRecord(t, []int64{1}, []string{"first"}), SourceLogicalBytes: 32, SourceRows: 1, CandidateCheckpoint: first},
			{Record: testArrowRecord(t, []int64{2}, []string{"second"}), SourceLogicalBytes: 32, SourceRows: 1, CandidateCheckpoint: second},
			{Record: testArrowRecord(t, []int64{3}, []string{"third"}), SourceLogicalBytes: 32, SourceRows: 1, CandidateCheckpoint: third},
		},
		onEmit: func(index int) {
			switch index {
			case 1:
				close(secondExtracted)
			case 2:
				close(thirdExtracted)
			}
		},
	}
	applyStarted := make(chan struct{}, 2)
	releaseApply := make(chan struct{})
	destination := &testArrowDestination{
		testDestinationExecutor: &testDestinationExecutor{reference: pair.destinationExecutor.reference, sink: pair.destination.Name()},
		run:                     &testArrowFullOverwriteRun{sink: pair.destination.Name(), applyStarted: applyStarted, releaseApply: releaseApply},
	}
	registry := NewRegistry(pair.verifier)
	if err := registry.RegisterSource(source); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterDestination(destination); err != nil {
		t.Fatal(err)
	}
	const plan = `{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"}]}`
	hash, err := databaseTransformHash(plan)
	if err != nil {
		t.Fatal(err)
	}
	type pipelineOutcome struct {
		result Result
		err    error
	}
	runDone := make(chan pipelineOutcome, 1)
	go func() {
		result, runErr := NewOrchestrator(registry).Run(context.Background(), RunRequest{
			ConnectionID: "test-connection-owner", Generation: 1, Source: pair.source, Destination: pair.destination,
			DestinationBinding: DestinationBinding{WorkspaceID: "workspace", SourceConnectorID: pair.source.Name(), ConnectionID: "test-connection-owner", StreamID: "stream", PrimaryKey: []string{"id"}},
			Stream:             "records", Mode: synccontract.ModeFullOverwrite, BatchSize: 1, MaxInFlightBatches: 2,
			TransformPlanJSON: plan, TransformPlanHash: hash, FastSegments: &testFastSegmentStore{}, ByteCreditCapacity: 64,
			Commit: func(synccontract.CheckpointEnvelope) error { return nil },
		})
		runDone <- pipelineOutcome{result: result, err: runErr}
	}()
	select {
	case <-applyStarted:
	case <-time.After(time.Second):
		t.Fatal("first COPY did not begin")
	}
	select {
	case <-secondExtracted:
	case <-time.After(time.Second):
		t.Fatal("source did not extract page two while page one COPY was blocked")
	}
	select {
	case <-thirdExtracted:
		t.Fatal("depth two extracted page three while page one COPY and page two admission were still in flight")
	case <-time.After(75 * time.Millisecond):
	}
	close(releaseApply)
	received := <-runDone
	if received.err != nil {
		t.Fatalf("Run() error = %v", received.err)
	}
	if got, want := destination.run.ids, []int64{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ordered pipeline COPY IDs = %v, want %v", got, want)
	}
	if received.result.PeakCreditBytes != 64 {
		t.Fatalf("ordered pipeline peak credit = %d, want two 32-byte admitted batches", received.result.PeakCreditBytes)
	}
}

func TestOrchestratorArrowFullOverwriteDepthOneKeepsSerialEmitBehavior(t *testing.T) {
	pair := newTestTransportPair("database", "database")
	pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "stage_replace"}}
	pair.destination.descriptor.Destination.EligibleActions = []string{"stage_replace"}
	first, second := testCheckpoint(pair.source.Name()), testCheckpoint(pair.source.Name())
	second.Position.Primary, second.Position.TieBreaker = synccontract.OpaqueToken("2"), synccontract.OpaqueToken("2")
	secondExtracted := make(chan struct{})
	source := &testArrowSource{testSourceExecutor: &testSourceExecutor{reference: pair.sourceExecutor.reference}, batches: []ArrowSourceBatch{
		{Record: testArrowRecord(t, []int64{1}, []string{"first"}), SourceLogicalBytes: 32, SourceRows: 1, CandidateCheckpoint: first},
		{Record: testArrowRecord(t, []int64{2}, []string{"second"}), SourceLogicalBytes: 32, SourceRows: 1, CandidateCheckpoint: second},
	}, onEmit: func(index int) {
		if index == 1 {
			close(secondExtracted)
		}
	}}
	applyStarted, releaseApply := make(chan struct{}, 2), make(chan struct{})
	destination := &testArrowDestination{testDestinationExecutor: &testDestinationExecutor{reference: pair.destinationExecutor.reference, sink: pair.destination.Name()}, run: &testArrowFullOverwriteRun{sink: pair.destination.Name(), applyStarted: applyStarted, releaseApply: releaseApply}}
	registry := NewRegistry(pair.verifier)
	if err := registry.RegisterSource(source); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterDestination(destination); err != nil {
		t.Fatal(err)
	}
	const plan = `{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"}]}`
	hash, err := databaseTransformHash(plan)
	if err != nil {
		t.Fatal(err)
	}
	runDone := make(chan error, 1)
	go func() {
		_, runErr := NewOrchestrator(registry).Run(context.Background(), RunRequest{
			ConnectionID: "test-connection-owner", Generation: 1, Source: pair.source, Destination: pair.destination,
			DestinationBinding: DestinationBinding{WorkspaceID: "workspace", SourceConnectorID: pair.source.Name(), ConnectionID: "test-connection-owner", StreamID: "stream", PrimaryKey: []string{"id"}},
			Stream:             "records", Mode: synccontract.ModeFullOverwrite, BatchSize: 1, MaxInFlightBatches: 1,
			TransformPlanJSON: plan, TransformPlanHash: hash, FastSegments: &testFastSegmentStore{}, ByteCreditCapacity: 64,
			Commit: func(synccontract.CheckpointEnvelope) error { return nil },
		})
		runDone <- runErr
	}()
	select {
	case <-applyStarted:
	case <-time.After(time.Second):
		t.Fatal("first COPY did not begin")
	}
	select {
	case <-secondExtracted:
		t.Fatal("depth one extracted page two before page one COPY completed")
	case <-time.After(75 * time.Millisecond):
	}
	close(releaseApply)
	if err := <-runDone; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if destination.run.publishCalls != 1 || destination.run.readBackCalls != 1 || destination.run.abortCalls != 0 {
		t.Fatalf("depth one publish/readback/abort = %d/%d/%d, want 1/1/0", destination.run.publishCalls, destination.run.readBackCalls, destination.run.abortCalls)
	}
}

func TestOrchestratorArrowFullOverwriteRefusesUndeclaredPipelineBeforeIO(t *testing.T) {
	pair := newTestTransportPair("database", "database")
	pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "stage_replace"}}
	pair.destination.descriptor.Destination.EligibleActions = []string{"stage_replace"}
	source := &testArrowSource{testSourceExecutor: &testSourceExecutor{reference: pair.sourceExecutor.reference}}
	destination := &testArrowDestination{testDestinationExecutor: &testDestinationExecutor{reference: pair.destinationExecutor.reference, sink: pair.destination.Name()}, run: &testArrowFullOverwriteRun{sink: pair.destination.Name()}}
	registry := NewRegistry(pair.verifier)
	if err := registry.RegisterSource(source); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterDestination(destination); err != nil {
		t.Fatal(err)
	}
	const plan = `{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"}]}`
	hash, err := databaseTransformHash(plan)
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewOrchestrator(registry).Run(context.Background(), RunRequest{
		ConnectionID: "test-connection-owner", Generation: 1, Source: pair.source, Destination: pair.destination,
		DestinationBinding: DestinationBinding{WorkspaceID: "workspace", SourceConnectorID: pair.source.Name(), ConnectionID: "test-connection-owner", StreamID: "stream", PrimaryKey: []string{"id"}},
		Stream:             "records", Mode: synccontract.ModeFullOverwrite, BatchSize: 1, MaxInFlightBatches: 2,
		TransformPlanJSON: plan, TransformPlanHash: hash, FastSegments: &testFastSegmentStore{}, Commit: func(synccontract.CheckpointEnvelope) error { return nil },
	})
	var unsupported *OrderedPipelineUnsupportedError
	if !errors.As(err, &unsupported) {
		t.Fatalf("Run() error = %T %v, want OrderedPipelineUnsupportedError", err, err)
	}
	if source.extractCalls != 0 || destination.planCalls != 0 || destination.beginCalls != 0 {
		t.Fatalf("undeclared pipeline reached extractor/plan/begin = %d/%d/%d, want 0/0/0", source.extractCalls, destination.planCalls, destination.beginCalls)
	}
}

func TestOrchestratorArrowFullOverwritePipelineFailureAbortsBeforePublicationOrCheckpoint(t *testing.T) {
	pair := newTestTransportPair("database", "database")
	pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.source.descriptor.Source.OrderedPipeline = true
	pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.OrderedPipeline = true
	pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "stage_replace"}}
	pair.destination.descriptor.Destination.EligibleActions = []string{"stage_replace"}
	first, second := testCheckpoint(pair.source.Name()), testCheckpoint(pair.source.Name())
	second.Position.Primary, second.Position.TieBreaker = synccontract.OpaqueToken("2"), synccontract.OpaqueToken("2")
	secondExtracted := make(chan struct{})
	source := &testArrowSource{testSourceExecutor: &testSourceExecutor{reference: pair.sourceExecutor.reference}, batches: []ArrowSourceBatch{
		{Record: testArrowRecord(t, []int64{1}, []string{"first"}), SourceLogicalBytes: 32, SourceRows: 1, CandidateCheckpoint: first},
		{Record: testArrowRecord(t, []int64{2}, []string{"second"}), SourceLogicalBytes: 32, SourceRows: 1, CandidateCheckpoint: second},
	}, onEmit: func(index int) {
		if index == 1 {
			close(secondExtracted)
		}
	}}
	applyStarted, releaseApply := make(chan struct{}, 2), make(chan struct{})
	destination := &testArrowDestination{testDestinationExecutor: &testDestinationExecutor{reference: pair.destinationExecutor.reference, sink: pair.destination.Name()}, run: &testArrowFullOverwriteRun{sink: pair.destination.Name(), applyStarted: applyStarted, releaseApply: releaseApply, failApplyAt: 2, applyErr: errors.New("injected pipeline COPY failure")}}
	registry := NewRegistry(pair.verifier)
	if err := registry.RegisterSource(source); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterDestination(destination); err != nil {
		t.Fatal(err)
	}
	const plan = `{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"}]}`
	hash, err := databaseTransformHash(plan)
	if err != nil {
		t.Fatal(err)
	}
	commits := 0
	runDone := make(chan error, 1)
	go func() {
		_, runErr := NewOrchestrator(registry).Run(context.Background(), RunRequest{
			ConnectionID: "test-connection-owner", Generation: 1, Source: pair.source, Destination: pair.destination,
			DestinationBinding: DestinationBinding{WorkspaceID: "workspace", SourceConnectorID: pair.source.Name(), ConnectionID: "test-connection-owner", StreamID: "stream", PrimaryKey: []string{"id"}},
			Stream:             "records", Mode: synccontract.ModeFullOverwrite, BatchSize: 1, MaxInFlightBatches: 2,
			TransformPlanJSON: plan, TransformPlanHash: hash, FastSegments: &testFastSegmentStore{}, ByteCreditCapacity: 64,
			Commit: func(synccontract.CheckpointEnvelope) error { commits++; return nil },
		})
		runDone <- runErr
	}()
	select {
	case <-applyStarted:
	case <-time.After(time.Second):
		t.Fatal("first COPY did not begin")
	}
	select {
	case <-secondExtracted:
	case <-time.After(time.Second):
		t.Fatal("page two was not queued before injected failure")
	}
	close(releaseApply)
	if err := <-runDone; err == nil || !strings.Contains(err.Error(), "injected pipeline COPY failure") {
		t.Fatalf("Run() error = %v, want injected COPY failure", err)
	}
	if destination.run.publishCalls != 0 || destination.run.readBackCalls != 0 || destination.run.abortCalls != 1 || commits != 0 {
		t.Fatalf("failed pipeline publish/readback/abort/checkpoints = %d/%d/%d/%d, want 0/0/1/0", destination.run.publishCalls, destination.run.readBackCalls, destination.run.abortCalls, commits)
	}
}

func TestOrchestratorArrowFullOverwritePipelineCancellationDrainsBeforeAbort(t *testing.T) {
	pair := newTestTransportPair("database", "database")
	pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.source.descriptor.Source.OrderedPipeline = true
	pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.OrderedPipeline = true
	pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "stage_replace"}}
	pair.destination.descriptor.Destination.EligibleActions = []string{"stage_replace"}
	first, second := testCheckpoint(pair.source.Name()), testCheckpoint(pair.source.Name())
	second.Position.Primary, second.Position.TieBreaker = synccontract.OpaqueToken("2"), synccontract.OpaqueToken("2")
	secondExtracted := make(chan struct{})
	source := &testArrowSource{testSourceExecutor: &testSourceExecutor{reference: pair.sourceExecutor.reference}, batches: []ArrowSourceBatch{
		{Record: testArrowRecord(t, []int64{1}, []string{"first"}), SourceLogicalBytes: 32, SourceRows: 1, CandidateCheckpoint: first},
		{Record: testArrowRecord(t, []int64{2}, []string{"second"}), SourceLogicalBytes: 32, SourceRows: 1, CandidateCheckpoint: second},
	}, onEmit: func(index int) {
		if index == 1 {
			close(secondExtracted)
		}
	}}
	applyStarted, releaseApply := make(chan struct{}, 2), make(chan struct{})
	destination := &testArrowDestination{testDestinationExecutor: &testDestinationExecutor{reference: pair.destinationExecutor.reference, sink: pair.destination.Name()}, run: &testArrowFullOverwriteRun{sink: pair.destination.Name(), applyStarted: applyStarted, releaseApply: releaseApply}}
	registry := NewRegistry(pair.verifier)
	if err := registry.RegisterSource(source); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterDestination(destination); err != nil {
		t.Fatal(err)
	}
	const plan = `{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"}]}`
	hash, err := databaseTransformHash(plan)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runDone := make(chan error, 1)
	go func() {
		_, runErr := NewOrchestrator(registry).Run(ctx, RunRequest{
			ConnectionID: "test-connection-owner", Generation: 1, Source: pair.source, Destination: pair.destination,
			DestinationBinding: DestinationBinding{WorkspaceID: "workspace", SourceConnectorID: pair.source.Name(), ConnectionID: "test-connection-owner", StreamID: "stream", PrimaryKey: []string{"id"}},
			Stream:             "records", Mode: synccontract.ModeFullOverwrite, BatchSize: 1, MaxInFlightBatches: 2,
			TransformPlanJSON: plan, TransformPlanHash: hash, FastSegments: &testFastSegmentStore{}, ByteCreditCapacity: 64,
			Commit: func(synccontract.CheckpointEnvelope) error { return nil },
		})
		runDone <- runErr
	}()
	select {
	case <-applyStarted:
	case <-time.After(time.Second):
		t.Fatal("first COPY did not begin")
	}
	select {
	case <-secondExtracted:
	case <-time.After(time.Second):
		t.Fatal("page two was not admitted before cancellation")
	}
	cancel()
	close(releaseApply)
	if err := <-runDone; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %T %v, want context.Canceled", err, err)
	}
	if destination.run.publishCalls != 0 || destination.run.readBackCalls != 0 || destination.run.abortCalls != 1 {
		t.Fatalf("canceled pipeline publish/readback/abort = %d/%d/%d, want 0/0/1", destination.run.publishCalls, destination.run.readBackCalls, destination.run.abortCalls)
	}
}

func TestOrchestratorArrowFullOverwriteRefusesMissingSegmentStoreBeforeExtractorIO(t *testing.T) {
	pair := newTestTransportPair("database", "database")
	pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "stage_replace"}}
	pair.destination.descriptor.Destination.EligibleActions = []string{"stage_replace"}
	source := &testArrowSource{testSourceExecutor: &testSourceExecutor{reference: pair.sourceExecutor.reference}}
	destination := &testArrowDestination{testDestinationExecutor: &testDestinationExecutor{reference: pair.destinationExecutor.reference, sink: pair.destination.Name()}, run: &testArrowFullOverwriteRun{sink: pair.destination.Name()}}
	registry := NewRegistry(pair.verifier)
	if err := registry.RegisterSource(source); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterDestination(destination); err != nil {
		t.Fatal(err)
	}
	_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullOverwrite, BatchSize: 1,
		TransformPlanJSON: `{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"}]}`,
		TransformPlanHash: "0c4d306bca2e827717e1f4ef641ce2bc39d7e05ac418bbf332d94e907e1d1fe1",
		Commit:            func(synccontract.CheckpointEnvelope) error { return nil },
	})
	if !errors.Is(err, ErrArrowFastPathInvalid) {
		t.Fatalf("Run() error = %T %v, want ErrArrowFastPathInvalid", err, err)
	}
	if source.extractCalls != 0 || destination.beginCalls != 0 {
		t.Fatalf("missing store reached extractor/begin = %d/%d, want 0/0", source.extractCalls, destination.beginCalls)
	}
}

func TestOrchestratorArrowFullOverwritePublishesZeroSourceWithNoCreditLeak(t *testing.T) {
	pair := newTestTransportPair("database", "database")
	pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "stage_replace"}}
	pair.destination.descriptor.Destination.EligibleActions = []string{"stage_replace"}
	source := &testArrowSource{testSourceExecutor: &testSourceExecutor{reference: pair.sourceExecutor.reference}}
	destination := &testArrowDestination{testDestinationExecutor: &testDestinationExecutor{reference: pair.destinationExecutor.reference, sink: pair.destination.Name()}, run: &testArrowFullOverwriteRun{sink: pair.destination.Name()}}
	registry := NewRegistry(pair.verifier)
	if err := registry.RegisterSource(source); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterDestination(destination); err != nil {
		t.Fatal(err)
	}
	const plan = `{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"}]}`
	hash, err := databaseTransformHash(plan)
	if err != nil {
		t.Fatal(err)
	}
	result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullOverwrite, BatchSize: 1,
		TransformPlanJSON: plan, TransformPlanHash: hash, FastSegments: &testFastSegmentStore{},
		Commit: func(synccontract.CheckpointEnvelope) error { return nil },
	})
	if err != nil {
		t.Fatalf("Run(zero source) error = %v", err)
	}
	if destination.run.publishCalls != 1 || result.PeakCreditBytes != 0 || result.RecordsApplied != 0 || result.CommittedCheckpoint != nil {
		t.Fatalf("zero-source Arrow result = %#v, publish calls=%d", result, destination.run.publishCalls)
	}
}

func TestOrchestratorTimesOutOnlyTheSecondDestinationUnitAndRetainsDurablePhaseCounts(t *testing.T) {
	pair := newTestTransportPair("api", "database")
	first := testCheckpoint("api")
	second := testCheckpoint("api")
	second.Position.Primary = synccontract.OpaqueToken("2")
	second.Position.TieBreaker = synccontract.OpaqueToken("2")
	pair.sourceExecutor.pages = []SourcePage{
		{Records: []connectors.Record{{"id": "one"}}, CandidateCheckpoint: first},
		{Records: []connectors.Record{{"id": "two"}}, CandidateCheckpoint: second},
	}
	pair.destinationExecutor.applyContext = func(ctx context.Context, calls int) error {
		if calls != 2 {
			return nil
		}
		<-ctx.Done()
		return ctx.Err()
	}
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)
	commits := 0

	result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
		BatchSize: 1, UnitDeadline: 20 * time.Millisecond, Stage: &testWarehouseStage{},
		Commit: func(synccontract.CheckpointEnvelope) error { commits++; return nil },
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run() error = %T %v, want a timed-out second unit", err, err)
	}
	if pair.destinationExecutor.applyCalls != 2 || commits != 1 {
		t.Fatalf("destination applies/checkpoints = %d/%d, want first unit durable and only the second unit timed out", pair.destinationExecutor.applyCalls, commits)
	}
	if result.RecordsRead != 2 || result.RecordsStaged != 2 || result.RecordsApplied != 1 || result.Pages != 1 {
		t.Fatalf("partial result = %+v, want extracted/staged/applied/pages = 2/2/1/1", result)
	}
	if result.ExtractElapsed <= 0 || result.StageElapsed <= 0 || result.ApplyElapsed <= 0 {
		t.Fatalf("phase timing = extract:%s stage:%s apply:%s, want all phase durations persisted before failure", result.ExtractElapsed, result.StageElapsed, result.ApplyElapsed)
	}
}

func TestOrchestratorRechecksAuthorizationImmediatelyBeforeEachApplyAndRefusesSecondEffect(t *testing.T) {
	pair := newTestTransportPair("api", "database")
	first := testCheckpoint("api")
	second := testCheckpoint("api")
	second.Position.Primary = synccontract.OpaqueToken("2")
	second.Position.TieBreaker = synccontract.OpaqueToken("2")
	pair.sourceExecutor.pages = []SourcePage{
		{Records: []connectors.Record{{"id": "one"}}, CandidateCheckpoint: first},
		{Records: []connectors.Record{{"id": "two"}}, CandidateCheckpoint: second},
	}
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)
	stage := &testWarehouseStage{}
	errAuthorizationRevoked := errors.New("authorization revoked")
	authorizationChecks := 0

	result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
		BatchSize: 1, Stage: stage,
		Approval: DestinationApproval{AuthorizeNextUnit: func(context.Context) error {
			authorizationChecks++
			if authorizationChecks == 2 {
				return errAuthorizationRevoked
			}
			return nil
		}},
		Commit: func(synccontract.CheckpointEnvelope) error { return nil },
	})
	if !errors.Is(err, errAuthorizationRevoked) {
		t.Fatalf("Run() error = %T %v, want second-unit authorization revocation", err, err)
	}
	if authorizationChecks != 2 || stage.calls != 2 || pair.destinationExecutor.applyCalls != 1 {
		t.Fatalf("authorization/stage/apply calls = %d/%d/%d, want 2/2/1 so revocation stops immediately before the second destination mutation", authorizationChecks, stage.calls, pair.destinationExecutor.applyCalls)
	}
	if result.RecordsRead != 2 || result.RecordsStaged != 2 || result.RecordsApplied != 1 || result.Pages != 1 {
		t.Fatalf("partial result = %+v, want extracted/staged/applied/pages = 2/2/1/1", result)
	}
}

func TestOrchestratorRevalidatesDestinationAuthorizationImmediatelyBeforeApply(t *testing.T) {
	pair := newTestTransportPair("api", "database")
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)
	stage := &testWarehouseStage{}
	errAuthorizationRevoked := errors.New("authorization revoked after staging")
	revoked := false
	stage.afterStage = func() { revoked = true }

	_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
		BatchSize: 10, Stage: stage,
		Approval: DestinationApproval{AuthorizeNextUnit: func(context.Context) error {
			if revoked {
				return errAuthorizationRevoked
			}
			return nil
		}},
		Commit: func(synccontract.CheckpointEnvelope) error { return nil },
	})
	if !errors.Is(err, errAuthorizationRevoked) {
		t.Fatalf("Run() error = %T %v, want authorization revocation after reopen", err, err)
	}
	if stage.calls != 1 || pair.destinationExecutor.applyCalls != 0 {
		t.Fatalf("stage/apply calls = %d/%d, want staged work and no destination mutation after revocation", stage.calls, pair.destinationExecutor.applyCalls)
	}
}

func TestOrchestratorFullOverwriteRevalidatesDestinationAuthorizationImmediatelyBeforeApply(t *testing.T) {
	pair := newTestTransportPair("api", "database")
	pair.source.descriptor.Source.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.Modes = []synccontract.Mode{synccontract.ModeFullOverwrite}
	pair.destination.descriptor.Destination.EligibleActions = []string{"replace"}
	pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{Mode: synccontract.ModeFullOverwrite, Strategy: connectors.ApplyStrategyReplace, Action: "replace"}}
	pair.destinationExecutor.fullOverwrite = &testFullOverwriteRun{sink: pair.destination.Name()}
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)
	stage := &testWarehouseStage{}
	errAuthorizationRevoked := errors.New("authorization revoked after full-overwrite staging")
	revoked := false
	stage.afterStage = func() { revoked = true }

	_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullOverwrite,
		BatchSize: 10, Stage: stage,
		Approval: DestinationApproval{AuthorizeNextUnit: func(context.Context) error {
			if revoked {
				return errAuthorizationRevoked
			}
			return nil
		}},
		Commit: func(synccontract.CheckpointEnvelope) error { return nil },
	})
	if !errors.Is(err, errAuthorizationRevoked) {
		t.Fatalf("Run() error = %T %v, want authorization revocation after full-overwrite staging", err, err)
	}
	if got := len(pair.destinationExecutor.fullOverwrite.ids); got != 0 {
		t.Fatalf("full-overwrite destination mutations = %d, want 0 after revocation", got)
	}
	if pair.destinationExecutor.fullOverwrite.publishCalls != 0 {
		t.Fatalf("full-overwrite publication calls = %d, want 0 after revocation", pair.destinationExecutor.fullOverwrite.publishCalls)
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

func TestOrchestratorRetiresOwnedStageOnlyAfterCheckpointCommit(t *testing.T) {
	tests := []struct {
		name        string
		commit      func(synccontract.CheckpointEnvelope) error
		wantErr     bool
		wantRetires int
	}{
		{
			name: "checkpoint committed",
			commit: func(synccontract.CheckpointEnvelope) error {
				return nil
			},
			wantRetires: 1,
		},
		{
			name: "checkpoint persistence fails",
			commit: func(synccontract.CheckpointEnvelope) error {
				return errors.New("synthetic checkpoint persistence failure")
			},
			wantErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			pair := newTestTransportPair("api", "database")
			registry := NewRegistry(pair.verifier)
			registerTransportPair(t, registry, pair)
			stage := &retiringTestWarehouseStage{}

			_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
				ConnectionID: "test-connection-owner", Generation: 1,
				Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
				BatchSize: 10, Stage: stage, Commit: testCase.commit,
			})
			if (err != nil) != testCase.wantErr {
				t.Fatalf("Run() error = %v, want error=%t", err, testCase.wantErr)
			}
			if got := len(stage.retired); got != testCase.wantRetires {
				t.Fatalf("retired receipts = %d, want %d; stage artifacts must survive an uncommitted checkpoint", got, testCase.wantRetires)
			}
			if stage.calls != 1 || pair.destinationExecutor.applyCalls != 1 {
				t.Fatalf("stage/apply calls = %d/%d, want one accepted effect before checkpoint disposition", stage.calls, pair.destinationExecutor.applyCalls)
			}
		})
	}
}

// TestTransport_PostCheckpointBookkeepingFailureRemainsDelivered is the
// transport-core half of the delivered-reconciliation contract.  The provider
// effect, read-back, and checkpoint have all completed before Retire is
// attempted, so a retiring-stage failure must retain that completed evidence
// rather than turning the request into a replayable failed delivery.
func TestTransport_PostCheckpointBookkeepingFailureRemainsDelivered(t *testing.T) {
	pair := newTestTransportPair("api", "database")
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)
	retireErr := errors.New("transient receipt retirement failure")
	stage := &failingRetiringTestWarehouseStage{retireErr: retireErr}
	commits := 0

	result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		ConnectionID: "test-connection-owner",
		Generation:   1,
		Source:       pair.source,
		Destination:  pair.destination,
		Stream:       "records",
		Mode:         synccontract.ModeFullAppend,
		BatchSize:    10,
		Stage:        stage,
		Commit: func(synccontract.CheckpointEnvelope) error {
			commits++
			return nil
		},
	})

	var reconciliation *DeliveredReconciliationRequiredError
	if !errors.As(err, &reconciliation) || !errors.Is(err, retireErr) {
		t.Fatalf("Run() error = %T %v, want delivered reconciliation wrapping %v", err, err, retireErr)
	}
	if !result.DeliveredReconciliationRequired || result.CommittedCheckpoint == nil || result.Pages != 1 || result.RecordsApplied != 1 {
		t.Fatalf("Run() result = %#v, want durable checkpoint plus reconciliation-required outcome", result)
	}
	if commits != 1 || pair.destinationExecutor.applyCalls != 1 || pair.destinationExecutor.readBackCalls != 1 || stage.retireCalls != 1 {
		t.Fatalf("post-checkpoint effects commit/apply/read-back/retire = %d/%d/%d/%d, want 1/1/1/1", commits, pair.destinationExecutor.applyCalls, pair.destinationExecutor.readBackCalls, stage.retireCalls)
	}
}

// TestTransport_DeferredCheckpointRetirementFailureRemainsDelivered covers
// the bootstrap/deferred branch: both bounded pages may have been applied,
// but only the final acknowledged checkpoint permits receipt retirement.
func TestTransport_DeferredCheckpointRetirementFailureRemainsDelivered(t *testing.T) {
	pair := newTestTransportPair("database", "database")
	first := testCheckpoint(pair.source.Name())
	second := testCheckpoint(pair.source.Name())
	second.Position.Primary = synccontract.OpaqueToken("second")
	second.Dedupe.Value = synccontract.OpaqueToken("second")
	second.DedupeWindow.End = synccontract.OpaqueToken("second")
	pair.sourceExecutor.pages = []SourcePage{
		{Records: []connectors.Record{{"id": "1"}}, CandidateCheckpoint: first, DeferCheckpoint: true},
		{Records: []connectors.Record{{"id": "2"}}, CandidateCheckpoint: second},
	}
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)
	retireErr := errors.New("deferred receipt retirement failure")
	stage := &failingRetiringTestWarehouseStage{retireErr: retireErr}
	commits := 0

	result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		ConnectionID: "test-connection-owner", Generation: 1,
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
		BatchSize: 10, Stage: stage, Commit: func(synccontract.CheckpointEnvelope) error { commits++; return nil },
	})
	var reconciliation *DeliveredReconciliationRequiredError
	if !errors.As(err, &reconciliation) || !errors.Is(err, retireErr) {
		t.Fatalf("Run() error = %T %v, want deferred delivered reconciliation wrapping %v", err, err, retireErr)
	}
	if !result.DeliveredReconciliationRequired || result.CommittedCheckpoint == nil || string(result.CommittedCheckpoint.Position.Primary) != "second" || result.Pages != 2 || result.RecordsApplied != 2 {
		t.Fatalf("Run() result = %#v, want both delivered pages and final durable checkpoint", result)
	}
	if commits != 1 || pair.destinationExecutor.applyCalls != 2 || pair.destinationExecutor.readBackCalls != 2 || stage.retireCalls != 1 {
		t.Fatalf("deferred post-checkpoint effects commit/apply/read-back/retire = %d/%d/%d/%d, want 1/2/2/1", commits, pair.destinationExecutor.applyCalls, pair.destinationExecutor.readBackCalls, stage.retireCalls)
	}
}

func TestOrchestratorAdmitsEmptyResultOnlyFromExplicitSourceMarker(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		marked  bool
		wantErr bool
	}{
		{name: "unmarked source remains refused", wantErr: true},
		{name: "explicit empty-result source succeeds", marked: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			pair := newTestTransportPair("database", "database")
			pair.sourceExecutor.pages = nil
			registry := NewRegistry(pair.verifier)
			if testCase.marked {
				if err := registry.RegisterSource(&emptyResultTestSource{testSourceExecutor: pair.sourceExecutor}); err != nil {
					t.Fatal(err)
				}
			} else if err := registry.RegisterSource(pair.sourceExecutor); err != nil {
				t.Fatal(err)
			}
			if err := registry.RegisterDestination(pair.destinationExecutor); err != nil {
				t.Fatal(err)
			}
			stage := &testWarehouseStage{}
			commits := 0
			result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
				Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
				BatchSize: 10, Stage: stage, Commit: func(synccontract.CheckpointEnvelope) error { commits++; return nil },
			})
			if (err != nil) != testCase.wantErr {
				t.Fatalf("empty source Run() error = %v, wantErr=%t", err, testCase.wantErr)
			}
			if stage.calls != 0 || pair.destinationExecutor.applyCalls != 0 || commits != 0 || result.Pages != 0 || result.CommittedCheckpoint != nil {
				t.Fatalf("empty source effects = stage:%d apply:%d commits:%d result:%+v, want no effects", stage.calls, pair.destinationExecutor.applyCalls, commits, result)
			}
		})
	}
}

func TestOrchestratorAppliesDeferredBootstrapPagesWithoutAdvancingCheckpoint(t *testing.T) {
	pair := newTestTransportPair("database", "database")
	first := testCheckpoint(pair.source.Name())
	second := testCheckpoint(pair.source.Name())
	second.Position.Primary = synccontract.OpaqueToken("second")
	second.Dedupe.Value = synccontract.OpaqueToken("second")
	second.DedupeWindow.End = synccontract.OpaqueToken("second")
	pair.sourceExecutor.pages = []SourcePage{
		{Records: []connectors.Record{{"id": "1"}}, CandidateCheckpoint: first, DeferCheckpoint: true},
		{Records: []connectors.Record{{"id": "2"}}, CandidateCheckpoint: second},
	}
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)
	stage := &testWarehouseStage{}
	commits := 0
	var committed synccontract.CheckpointEnvelope
	var committedReceipts []WarehouseReceipt

	result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
		BatchSize: 10, Stage: stage, CommitWorksets: func(checkpoint synccontract.CheckpointEnvelope, receipts []WarehouseReceipt) error {
			commits++
			committed = checkpoint.Clone()
			committedReceipts = append([]WarehouseReceipt(nil), receipts...)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Run() = %v", err)
	}
	if pair.destinationExecutor.applyCalls != 2 || pair.destinationExecutor.readBackCalls != 2 || result.Pages != 2 || result.RecordsApplied != 2 {
		t.Fatalf("deferred bootstrap apply/read-back/result = %d/%d/%+v, want both pages durable", pair.destinationExecutor.applyCalls, pair.destinationExecutor.readBackCalls, result)
	}
	if commits != 1 || string(committed.Position.Primary) != "second" || result.CommittedCheckpoint == nil || string(result.CommittedCheckpoint.Position.Primary) != "second" {
		t.Fatalf("deferred bootstrap checkpoint = commits %d committed=%+v result=%+v, want only final page", commits, committed, result.CommittedCheckpoint)
	}
	if len(committedReceipts) != 2 || committedReceipts[0].ID != "stage-1" || committedReceipts[1].ID != "stage-2" {
		t.Fatalf("deferred bootstrap committed receipts = %#v, want both pending worksets", committedReceipts)
	}
}

func TestCloneRecordCopiesBinaryValuesAtEveryNestingLevel(t *testing.T) {
	scalar := []byte{0x01}
	nested := []byte{0x02}
	list := []byte{0x03}
	original := connectors.Record{
		"scalar": scalar,
		"nested": map[string]any{"binary": nested},
		"list":   []any{list},
	}

	clone, err := cloneRecord(original)
	if err != nil {
		t.Fatalf("cloneRecord() = %v", err)
	}
	clone["scalar"].([]byte)[0] = 0x11
	clone["nested"].(map[string]any)["binary"].([]byte)[0] = 0x12
	clone["list"].([]any)[0].([]byte)[0] = 0x13

	if scalar[0] != 0x01 || nested[0] != 0x02 || list[0] != 0x03 {
		t.Fatalf("clone mutated provider binary values: scalar=%x nested=%x list=%x", scalar, nested, list)
	}
}

func TestCloneRecordCopiesRawMessageAndStringMapValuesAtEveryNestingLevel(t *testing.T) {
	raw := json.RawMessage(`{"owner":"source"}`)
	labels := map[string]string{"owner": "source"}
	nestedRaw := json.RawMessage(`{"level":"nested"}`)
	nestedLabels := map[string]string{"level": "nested"}
	listRaw := json.RawMessage(`{"level":"list"}`)
	listLabels := map[string]string{"level": "list"}
	recordRaw := json.RawMessage(`{"level":"record"}`)
	recordLabels := map[string]string{"level": "record"}
	original := connectors.Record{
		"raw":    raw,
		"labels": labels,
		"nested": map[string]any{
			"raw":    nestedRaw,
			"labels": nestedLabels,
		},
		"list": []any{listRaw, listLabels},
		"records": []connectors.Record{{
			"raw":    recordRaw,
			"labels": recordLabels,
		}},
	}

	clone, err := cloneRecord(original)
	if err != nil {
		t.Fatalf("cloneRecord() = %v", err)
	}
	clone["raw"].(json.RawMessage)[0] = '['
	clone["labels"].(map[string]string)["owner"] = "clone"
	clone["nested"].(map[string]any)["raw"].(json.RawMessage)[0] = '['
	clone["nested"].(map[string]any)["labels"].(map[string]string)["level"] = "clone"
	clone["list"].([]any)[0].(json.RawMessage)[0] = '['
	clone["list"].([]any)[1].(map[string]string)["level"] = "clone"
	clone["records"].([]connectors.Record)[0]["raw"].(json.RawMessage)[0] = '['
	clone["records"].([]connectors.Record)[0]["labels"].(map[string]string)["level"] = "clone"

	if got := string(raw); got != `{"owner":"source"}` {
		t.Fatalf("direct raw message clone mutated source storage: %q", got)
	}
	if got := labels["owner"]; got != "source" {
		t.Fatalf("direct string map clone mutated source storage: %q", got)
	}
	if got := string(nestedRaw); got != `{"level":"nested"}` {
		t.Fatalf("nested raw message clone mutated source storage: %q", got)
	}
	if got := nestedLabels["level"]; got != "nested" {
		t.Fatalf("nested string map clone mutated source storage: %q", got)
	}
	if got := string(listRaw); got != `{"level":"list"}` {
		t.Fatalf("list raw message clone mutated source storage: %q", got)
	}
	if got := listLabels["level"]; got != "list" {
		t.Fatalf("list string map clone mutated source storage: %q", got)
	}
	if got := string(recordRaw); got != `{"level":"record"}` {
		t.Fatalf("record raw message clone mutated source storage: %q", got)
	}
	if got := recordLabels["level"]; got != "record" {
		t.Fatalf("record string map clone mutated source storage: %q", got)
	}
}

func TestCloneRuntimeConfigDefensivelyCopiesCatalogNestedState(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"}}}`)
	runtime := connectors.RuntimeConfig{
		Config:                map[string]string{"endpoint": "source"},
		Secrets:               map[string]string{"token": "source"},
		ApprovedPayloadSHA256: map[string]string{"payload": "source"},
		ResolvedCatalog: &connectors.Catalog{
			Connector: "source",
			Streams: []connectors.Stream{{
				Name: "records", Fields: []connectors.Field{{Name: "id", Type: "string"}},
				PrimaryKey: []string{"id"}, CursorFields: []string{"id"}, Schema: schema,
			}},
			Discovery: &connectors.DiscoveryStatus{Failures: []connectors.DiscoveryFailure{{Object: "records", Stage: "source", Attempts: 1}}},
		},
	}

	clone := cloneRuntimeConfig(runtime)
	clone.Config["endpoint"] = "clone"
	clone.Secrets["token"] = "clone"
	clone.ApprovedPayloadSHA256["payload"] = "clone"
	clone.ResolvedCatalog.Streams[0].Fields[0].Name = "clone"
	clone.ResolvedCatalog.Streams[0].PrimaryKey[0] = "clone"
	clone.ResolvedCatalog.Streams[0].CursorFields[0] = "clone"
	clone.ResolvedCatalog.Streams[0].Schema[0] = '['
	clone.ResolvedCatalog.Discovery.Failures[0].Stage = "clone"

	stream := runtime.ResolvedCatalog.Streams[0]
	if runtime.Config["endpoint"] != "source" || runtime.Secrets["token"] != "source" || runtime.ApprovedPayloadSHA256["payload"] != "source" {
		t.Fatalf("runtime maps were mutated through clone: %#v", runtime)
	}
	if stream.Fields[0].Name != "id" || stream.PrimaryKey[0] != "id" || stream.CursorFields[0] != "id" || string(stream.Schema) != `{"type":"object","properties":{"id":{"type":"string"}}}` {
		t.Fatalf("catalog stream was mutated through clone: %#v", stream)
	}
	if got := runtime.ResolvedCatalog.Discovery.Failures[0].Stage; got != "source" {
		t.Fatalf("catalog discovery failure was mutated through clone: %q", got)
	}

	begin := cloneArrowFullOverwriteRunRequest(ArrowFullOverwriteRunRequest{
		Runtime: runtime, SourceRuntime: runtime, Binding: DestinationBinding{PrimaryKey: []string{"id"}},
	})
	firstSegment := cloneArrowBulkApplyRequest(ArrowBulkApplyRequest{
		Runtime: runtime, SourceRuntime: runtime, Binding: DestinationBinding{PrimaryKey: []string{"id"}},
	})
	secondSegment := cloneArrowBulkApplyRequest(ArrowBulkApplyRequest{
		Runtime: runtime, SourceRuntime: runtime, Binding: DestinationBinding{PrimaryKey: []string{"id"}},
	})
	begin.Runtime.ResolvedCatalog.Streams[0].Fields[0].Name = "begin"
	begin.SourceRuntime.ResolvedCatalog.Discovery.Failures[0].Stage = "begin"
	begin.Binding.PrimaryKey[0] = "begin"
	firstSegment.Runtime.ResolvedCatalog.Streams[0].PrimaryKey[0] = "first"
	firstSegment.SourceRuntime.ResolvedCatalog.Streams[0].CursorFields[0] = "first"
	firstSegment.Binding.PrimaryKey[0] = "first"
	if got := secondSegment.Runtime.ResolvedCatalog.Streams[0].PrimaryKey[0]; got != "id" {
		t.Fatalf("first Arrow segment changed later segment runtime primary key: %q", got)
	}
	if got := secondSegment.SourceRuntime.ResolvedCatalog.Streams[0].CursorFields[0]; got != "id" {
		t.Fatalf("first Arrow segment changed later segment source cursor: %q", got)
	}
	if got := secondSegment.Binding.PrimaryKey[0]; got != "id" {
		t.Fatalf("first Arrow segment changed later segment binding: %q", got)
	}
	if got := runtime.ResolvedCatalog.Streams[0].Fields[0].Name; got != "id" {
		t.Fatalf("Arrow request changed caller catalog field: %q", got)
	}
	if got := runtime.ResolvedCatalog.Discovery.Failures[0].Stage; got != "source" {
		t.Fatalf("Arrow request changed caller discovery failure: %q", got)
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

func TestOrchestratorProtectsRawMessageAndStringMapFromStageAndDestinationMutation(t *testing.T) {
	for _, tt := range []struct {
		name                    string
		mutateRawMessageInStage bool
		mutateStringMapInDest   bool
	}{
		{name: "warehouse stage mutates raw message", mutateRawMessageInStage: true},
		{name: "destination mutates string map", mutateStringMapInDest: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			pair := newTestTransportPair("api", "database")
			raw := json.RawMessage(`{"owner":"source"}`)
			labels := map[string]string{"owner": "source"}
			pair.sourceExecutor.pages[0].Records[0]["raw"] = raw
			pair.sourceExecutor.pages[0].Records[0]["labels"] = labels
			registry := NewRegistry(pair.verifier)
			registerTransportPair(t, registry, pair)
			stage := &testWarehouseStage{mutateRawMessage: tt.mutateRawMessageInStage}
			pair.destinationExecutor.mutateStringMap = tt.mutateStringMapInDest

			_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
				Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
				BatchSize: 10, Stage: stage, Commit: func(synccontract.CheckpointEnvelope) error { return nil },
			})
			if err != nil {
				t.Fatalf("Run() = %v", err)
			}
			if got := string(raw); got != `{"owner":"source"}` {
				t.Fatalf("source raw message after downstream mutation = %q, want untouched", got)
			}
			if got := labels["owner"]; got != "source" {
				t.Fatalf("source string map after downstream mutation = %q, want untouched", got)
			}
		})
	}
}

func TestOrchestratorRejectsUnsupportedMutableValuesBeforeBoundaryCrossing(t *testing.T) {
	for _, tt := range []struct {
		name                 string
		configure            func(*testTransportPair, *testWarehouseStage)
		wantStageCalls       int
		wantDestinationCalls int
	}{
		{
			name: "source record",
			configure: func(pair *testTransportPair, _ *testWarehouseStage) {
				pair.sourceExecutor.pages[0].Records[0]["unsupported"] = map[string]int{"source": 1}
			},
			wantStageCalls:       0,
			wantDestinationCalls: 0,
		},
		{
			name: "warehouse workset",
			configure: func(_ *testTransportPair, stage *testWarehouseStage) {
				stage.injectUnsupportedRecord = true
			},
			wantStageCalls:       1,
			wantDestinationCalls: 0,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			pair := newTestTransportPair("api", "database")
			registry := NewRegistry(pair.verifier)
			registerTransportPair(t, registry, pair)
			stage := &testWarehouseStage{}
			tt.configure(pair, stage)

			_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
				Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
				BatchSize: 10, Stage: stage, Commit: func(synccontract.CheckpointEnvelope) error { return nil },
			})
			if !errors.Is(err, errUnsupportedTransportRecordValue) {
				t.Fatalf("Run() error = %v, want unsupported mutable value rejection", err)
			}
			if stage.calls != tt.wantStageCalls {
				t.Fatalf("warehouse stage calls = %d, want %d", stage.calls, tt.wantStageCalls)
			}
			if pair.destinationExecutor.applyCalls != tt.wantDestinationCalls {
				t.Fatalf("destination apply calls = %d, want %d", pair.destinationExecutor.applyCalls, tt.wantDestinationCalls)
			}
		})
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

func TestOrchestratorCommitsAcknowledgedPageBeforeReturningCancellation(t *testing.T) {
	pair := newTestTransportPair("api", "database")
	registry := NewRegistry(pair.verifier)
	registerTransportPair(t, registry, pair)
	ctx, cancel := context.WithCancel(context.Background())
	pair.destinationExecutor.afterApply = cancel
	commits := 0

	_, err := NewOrchestrator(registry).Run(ctx, RunRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
		BatchSize: 10, Stage: &testWarehouseStage{},
		Commit: func(synccontract.CheckpointEnvelope) error {
			commits++
			return nil
		},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want cancellation", err)
	}
	if pair.destinationExecutor.applyCalls != 1 || commits != 1 {
		t.Fatalf("apply/commits = %d/%d, want acknowledged page committed before cancellation", pair.destinationExecutor.applyCalls, commits)
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
	reference   connectors.TransportExecutorReference
	pages       []SourcePage
	readCalls   int
	events      *[]string
	lastRequest SourceRequest
}

// budgetStoppedTestSource models the closed error-only compatibility path
// used by full overwrite. The engine has already staged the bounded prefix,
// but it must never let that prefix reach replacement publication.
type budgetStoppedTestSource struct {
	*testSourceExecutor
	continuation synccontract.SourceContinuation
}

func (e *budgetStoppedTestSource) ReadTransport(ctx context.Context, request SourceRequest, emit func(SourcePage) error) error {
	if err := e.testSourceExecutor.ReadTransport(ctx, request, emit); err != nil {
		return err
	}
	return &SourceBudgetStoppedError{Continuation: *e.continuation.Clone()}
}

type emptyResultTestSource struct {
	*testSourceExecutor
}

func (*emptyResultTestSource) AllowEmptySourceResult() {}

func (e *testSourceExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return e.reference
}
func (e *testSourceExecutor) ReadTransport(ctx context.Context, request SourceRequest, emit func(SourcePage) error) error {
	e.readCalls++
	e.lastRequest = cloneSourceRequest(request)
	if e.events != nil {
		*e.events = append(*e.events, "source-read")
	}
	for _, page := range e.pages {
		if err := ctx.Err(); err != nil {
			return err
		}
		if request.RecordExtraction != nil {
			request.RecordExtraction(time.Nanosecond)
		}
		if err := emit(page); err != nil {
			return err
		}
	}
	return nil
}

type testDestinationExecutor struct {
	reference           connectors.TransportExecutorReference
	sink                string
	acknowledgement     synccontract.DownstreamAcknowledgement
	acknowledgementSet  bool
	afterApply          func()
	applyContext        func(context.Context, int) error
	afterReadBack       func()
	readBackContext     func(context.Context, int) error
	readBackErr         error
	mutateStringMap     bool
	planCalls           int
	applyCalls          int
	readBackCalls       int
	lastPlan            DestinationPlanRequest
	lastApply           DestinationApplyRequest
	lastReadBack        DestinationReadBackRequest
	events              *[]string
	fullOverwrite       *testFullOverwriteRun
	fullOverwriteBegins int
}

func (e *testDestinationExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return e.reference
}
func (e *testDestinationExecutor) PlanDestination(_ context.Context, request DestinationPlanRequest) (DestinationPlan, error) {
	e.planCalls++
	e.lastPlan = request
	if e.events != nil {
		*e.events = append(*e.events, "destination-plan")
	}
	return DestinationPlan{ApplyStrategy: request.ApplyStrategy}, nil
}
func (e *testDestinationExecutor) ApplyDestination(ctx context.Context, request DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	e.applyCalls++
	e.lastApply = request
	if e.events != nil {
		*e.events = append(*e.events, "destination-apply")
	}
	if e.applyContext != nil {
		if err := e.applyContext(ctx, e.applyCalls); err != nil {
			return synccontract.DownstreamAcknowledgement{}, err
		}
	}
	if e.mutateStringMap {
		labels, ok := request.Workset.Records[0]["labels"].(map[string]string)
		if !ok {
			return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("test destination expected string map")
		}
		labels["owner"] = "destination"
	}
	var acknowledgement synccontract.DownstreamAcknowledgement
	if e.acknowledgementSet {
		acknowledgement = e.acknowledgement
	} else {
		var err error
		acknowledgement, err = synccontract.NewDurableDownstreamAcknowledgement(e.sink, time.Now().UTC())
		if err != nil {
			return synccontract.DownstreamAcknowledgement{}, err
		}
	}
	if e.afterApply != nil {
		e.afterApply()
	}
	return acknowledgement, nil
}

func (e *testDestinationExecutor) ReadBackDestination(ctx context.Context, request DestinationReadBackRequest) error {
	e.readBackCalls++
	e.lastReadBack = request
	if e.events != nil {
		*e.events = append(*e.events, "destination-readback")
	}
	if e.readBackContext != nil {
		if err := e.readBackContext(ctx, e.readBackCalls); err != nil {
			return err
		}
	}
	if e.afterReadBack != nil {
		e.afterReadBack()
	}
	return e.readBackErr
}

func (e *testDestinationExecutor) BeginFullOverwrite(_ context.Context, request FullOverwriteRunRequest) (FullOverwriteRun, error) {
	e.fullOverwriteBegins++
	if request.Mode != synccontract.ModeFullOverwrite || e.fullOverwrite == nil {
		return nil, fmt.Errorf("test full-overwrite run is unavailable")
	}
	return e.fullOverwrite, nil
}

type testFullOverwriteRun struct {
	sink            string
	ids             []string
	applyCalls      int
	publishCalls    int
	readBackCalls   int
	abortCalls      int
	applyContext    func(context.Context, int) error
	publishContext  func(context.Context, int) error
	readBackContext func(context.Context, int) error
	publishOutput   json.RawMessage
	readBackErr     error
}

func (r *testFullOverwriteRun) ApplyFullOverwrite(ctx context.Context, request DestinationApplyRequest) error {
	r.applyCalls++
	if r.applyContext != nil {
		if err := r.applyContext(ctx, r.applyCalls); err != nil {
			return err
		}
	}
	for _, record := range request.Workset.Records {
		id, ok := record["id"].(string)
		if !ok {
			return fmt.Errorf("test full-overwrite record has no string ID")
		}
		r.ids = append(r.ids, id)
	}
	return nil
}

func (r *testFullOverwriteRun) PublishFullOverwrite(ctx context.Context, request FullOverwritePublicationRequest) (synccontract.DownstreamAcknowledgement, error) {
	r.publishCalls++
	if r.publishContext != nil {
		if err := r.publishContext(ctx, r.publishCalls); err != nil {
			return synccontract.DownstreamAcknowledgement{}, err
		}
	}
	if request.Pages != len(r.ids) || request.Records != len(r.ids) || request.LastCheckpoint == nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("test full-overwrite publication is incomplete")
	}
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement(r.sink, time.Now().UTC())
	acknowledgement.Output = append(json.RawMessage(nil), r.publishOutput...)
	return acknowledgement, err
}

func (r *testFullOverwriteRun) ReadBackFullOverwrite(ctx context.Context, acknowledgement synccontract.DownstreamAcknowledgement) error {
	r.readBackCalls++
	if r.readBackContext != nil {
		if err := r.readBackContext(ctx, r.readBackCalls); err != nil {
			return err
		}
	}
	if acknowledgement.Sink != r.sink || acknowledgement.AcknowledgedAt.IsZero() {
		return fmt.Errorf("test full-overwrite receipt is not durable")
	}
	return r.readBackErr
}

func (r *testFullOverwriteRun) AbortFullOverwrite(context.Context) error {
	r.abortCalls++
	return nil
}

type testArrowSource struct {
	*testSourceExecutor
	batches      []ArrowSourceBatch
	extractCalls int
	onEmit       func(int)
}

func (*testArrowSource) AllowEmptySourceResult() {}

func (s *testArrowSource) ExtractArrowRanges(ctx context.Context, _ ArrowExtractRequest, emit func(ArrowSourceBatch) error) error {
	s.extractCalls++
	for index, batch := range s.batches {
		if err := ctx.Err(); err != nil {
			return err
		}
		if s.onEmit != nil {
			s.onEmit(index)
		}
		if err := emit(batch); err != nil {
			return err
		}
		batch.Record.Release()
	}
	return nil
}

type testArrowDestination struct {
	*testDestinationExecutor
	run        *testArrowFullOverwriteRun
	beginCalls int
}

func (d *testArrowDestination) PlanDestination(_ context.Context, request DestinationPlanRequest) (DestinationPlan, error) {
	d.planCalls++
	d.lastPlan = request
	return DestinationPlan{ApplyStrategy: request.ApplyStrategy, TransformPlanHash: request.TransformPlanHash}, nil
}

func (d *testArrowDestination) BeginArrowFullOverwrite(_ context.Context, request ArrowFullOverwriteRunRequest) (ArrowFullOverwriteRun, error) {
	d.beginCalls++
	if request.Plan.TransformPlanHash != request.TransformPlanHash || d.run == nil {
		return nil, ErrArrowFastPathInvalid
	}
	return d.run, nil
}

type testArrowFullOverwriteRun struct {
	sink            string
	ids             []int64
	publishCalls    int
	readBackCalls   int
	abortCalls      int
	applyStarted    chan<- struct{}
	releaseApply    <-chan struct{}
	applyCalls      int
	failApplyAt     int
	applyErr        error
	publishContext  func(context.Context, int) error
	readBackContext func(context.Context, int) error
	publishOutput   json.RawMessage
	readBackErr     error
}

func (r *testArrowFullOverwriteRun) ApplyArrowSegment(_ context.Context, request ArrowBulkApplyRequest) error {
	r.applyCalls++
	if r.applyStarted != nil {
		r.applyStarted <- struct{}{}
	}
	if r.releaseApply != nil {
		<-r.releaseApply
	}
	if r.failApplyAt == r.applyCalls && r.applyErr != nil {
		return r.applyErr
	}
	ids, ok := request.Record.Column(0).(*array.Int64)
	if !ok || request.Segment.TransformedRows != request.Record.NumRows() {
		return ErrArrowFastPathInvalid
	}
	for index := 0; index < ids.Len(); index++ {
		r.ids = append(r.ids, ids.Value(index))
	}
	return nil
}

func (r *testArrowFullOverwriteRun) PublishArrowFullOverwrite(ctx context.Context, request ArrowFullOverwritePublicationRequest) (synccontract.DownstreamAcknowledgement, error) {
	r.publishCalls++
	if r.publishContext != nil {
		if err := r.publishContext(ctx, r.publishCalls); err != nil {
			return synccontract.DownstreamAcknowledgement{}, err
		}
	}
	if request.TransformedRows != int64(len(r.ids)) || request.SourceRows < request.TransformedRows {
		return synccontract.DownstreamAcknowledgement{}, ErrArrowFastPathInvalid
	}
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement(r.sink, time.Now().UTC())
	acknowledgement.Output = append(json.RawMessage(nil), r.publishOutput...)
	return acknowledgement, err
}

func (r *testArrowFullOverwriteRun) ReadBackArrowFullOverwrite(ctx context.Context, acknowledgement synccontract.DownstreamAcknowledgement) error {
	r.readBackCalls++
	if r.readBackContext != nil {
		if err := r.readBackContext(ctx, r.readBackCalls); err != nil {
			return err
		}
	}
	if acknowledgement.Sink != r.sink || acknowledgement.AcknowledgedAt.IsZero() {
		return ErrArrowFastPathInvalid
	}
	return r.readBackErr
}

func (r *testArrowFullOverwriteRun) AbortArrowFullOverwrite(context.Context) error {
	r.abortCalls++
	return nil
}

type testFastSegmentStore struct {
	calls      int
	afterStore func()
}

func (s *testFastSegmentStore) StoreArrowSegment(_ context.Context, request FastSegmentWriteRequest) (FastSegmentReceipt, error) {
	s.calls++
	if request.Record == nil || request.SegmentID == "" {
		return FastSegmentReceipt{}, ErrArrowFastPathInvalid
	}
	if s.afterStore != nil {
		s.afterStore()
	}
	return FastSegmentReceipt{
		ID: request.SegmentID, SchemaHash: "schema", TransformPlanHash: request.TransformPlanHash,
		ContentSHA256: "content", ParquetSHA256: "parquet", SourceLogicalBytes: request.SourceLogicalBytes,
		SourceRows: request.SourceRows, TransformedRows: request.Record.NumRows(), TransformedBytes: 16, ParquetBytes: 16,
	}, nil
}

func (*testFastSegmentStore) DiscardArrowSegment(context.Context, FastSegmentReceipt) error {
	return nil
}

func testArrowRecord(t *testing.T, ids []int64, statuses []string) arrow.Record {
	t.Helper()
	if len(ids) != len(statuses) {
		t.Fatalf("Arrow input ids/statuses length = %d/%d", len(ids), len(statuses))
	}
	schema := arrow.NewSchema([]arrow.Field{{Name: "id", Type: arrow.PrimitiveTypes.Int64}, {Name: "status", Type: arrow.BinaryTypes.String}}, nil)
	builder := array.NewRecordBuilder(memory.DefaultAllocator, schema)
	builder.Field(0).(*array.Int64Builder).AppendValues(ids, nil)
	builder.Field(1).(*array.StringBuilder).AppendValues(statuses, nil)
	record := builder.NewRecord()
	builder.Release()
	return record
}

func databaseTransformHash(raw string) (string, error) {
	plan, err := database.ParseTransformPlanV1([]byte(raw))
	if err != nil {
		return "", err
	}
	return plan.Hash(), nil
}

type testWarehouseStage struct {
	calls                   int
	lastPage                SourcePage
	lastRequest             WarehouseStageRequest
	lastReceipt             WarehouseReceipt
	events                  *[]string
	afterStage              func()
	mutateNestedPayload     bool
	mutateRawMessage        bool
	injectUnsupportedRecord bool
	worksets                map[string]WarehouseWorkset
}

type retiringTestWarehouseStage struct {
	testWarehouseStage
	retired []WarehouseReceipt
}

func (s *retiringTestWarehouseStage) Retire(_ context.Context, receipt WarehouseReceipt) error {
	s.retired = append(s.retired, receipt)
	return nil
}

type failingRetiringTestWarehouseStage struct {
	testWarehouseStage
	retireCalls int
	retireErr   error
}

func (s *failingRetiringTestWarehouseStage) Retire(_ context.Context, _ WarehouseReceipt) error {
	s.retireCalls++
	return s.retireErr
}

func (s *testWarehouseStage) Stage(_ context.Context, request WarehouseStageRequest) (WarehouseReceipt, error) {
	s.calls++
	s.lastPage = request.Page
	s.lastRequest = request
	if s.events != nil {
		*s.events = append(*s.events, "warehouse-stage")
	}
	if s.mutateNestedPayload {
		nested, ok := request.Page.Records[0]["nested"].(map[string]any)
		if !ok {
			return WarehouseReceipt{}, fmt.Errorf("test stage expected nested provider map")
		}
		nested["staged"] = true
	}
	if s.mutateRawMessage {
		raw, ok := request.Page.Records[0]["raw"].(json.RawMessage)
		if !ok {
			return WarehouseReceipt{}, fmt.Errorf("test stage expected raw message")
		}
		raw[0] = '['
	}
	if s.afterStage != nil {
		s.afterStage()
	}
	workset := WarehouseWorkset{ID: fmt.Sprintf("stage-%d", s.calls), Records: request.Page.Records, Tombstones: request.Page.Tombstones, CandidateCheckpoint: request.Page.CandidateCheckpoint}
	if s.injectUnsupportedRecord {
		workset.Records = []connectors.Record{{"unsupported": map[string]int{"warehouse": 1}}}
	}
	if s.worksets == nil {
		s.worksets = make(map[string]WarehouseWorkset)
	}
	s.worksets[workset.ID] = workset
	receipt := WarehouseReceipt{
		ID:               workset.ID,
		Owner:            "test-connection-owner",
		Generation:       1,
		Stream:           request.Stream,
		Mode:             request.Mode,
		CheckpointSHA256: "test-checkpoint",
		TombstonesSHA256: "test-tombstones",
		ManifestSHA256:   "test-manifest",
		ContentSHA256:    "test-content",
		ParquetSHA256:    "test-parquet",
		Records:          len(workset.Records),
		Tombstones:       len(workset.Tombstones),
	}
	s.lastReceipt = receipt
	return receipt, nil
}

func (s *testWarehouseStage) Reopen(_ context.Context, receipt WarehouseReceipt) (WarehouseWorkset, error) {
	if err := receipt.Validate(); err != nil {
		return WarehouseWorkset{}, err
	}
	if s.events != nil {
		*s.events = append(*s.events, "warehouse-reopen")
	}
	workset, ok := s.worksets[receipt.ID]
	if !ok {
		return WarehouseWorkset{}, fmt.Errorf("test stage receipt %q is unavailable", receipt.ID)
	}
	return workset, nil
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

type typedNilConformanceVerifier struct{}

func (*typedNilConformanceVerifier) VerifyTransportConformance(ConformanceVerification) error {
	panic("typed nil conformance verifier must not be invoked")
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
