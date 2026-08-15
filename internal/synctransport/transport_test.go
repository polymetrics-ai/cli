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
	commits := 0
	var committed synccontract.CheckpointEnvelope

	result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
		Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
		BatchSize: 10, Stage: &testWarehouseStage{}, Commit: func(checkpoint synccontract.CheckpointEnvelope) error {
			commits++
			committed = checkpoint.Clone()
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
	reference connectors.TransportExecutorReference
	pages     []SourcePage
	readCalls int
}

type emptyResultTestSource struct {
	*testSourceExecutor
}

func (*emptyResultTestSource) AllowEmptySourceResult() {}

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
	afterApply         func()
	afterReadBack      func()
	readBackErr        error
	mutateStringMap    bool
	planCalls          int
	applyCalls         int
	readBackCalls      int
	lastPlan           DestinationPlanRequest
	lastReadBack       DestinationReadBackRequest
}

func (e *testDestinationExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return e.reference
}
func (e *testDestinationExecutor) PlanDestination(_ context.Context, request DestinationPlanRequest) (DestinationPlan, error) {
	e.planCalls++
	e.lastPlan = request
	return DestinationPlan{ApplyStrategy: request.ApplyStrategy}, nil
}
func (e *testDestinationExecutor) ApplyDestination(_ context.Context, request DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	e.applyCalls++
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

func (e *testDestinationExecutor) ReadBackDestination(_ context.Context, request DestinationReadBackRequest) error {
	e.readBackCalls++
	e.lastReadBack = request
	if e.afterReadBack != nil {
		e.afterReadBack()
	}
	return e.readBackErr
}

type testWarehouseStage struct {
	calls                   int
	lastPage                SourcePage
	afterStage              func()
	mutateNestedPayload     bool
	mutateRawMessage        bool
	injectUnsupportedRecord bool
	worksets                map[string]WarehouseWorkset
}

func (s *testWarehouseStage) Stage(_ context.Context, request WarehouseStageRequest) (WarehouseReceipt, error) {
	s.calls++
	s.lastPage = request.Page
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
	return WarehouseReceipt{
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
	}, nil
}

func (s *testWarehouseStage) Reopen(_ context.Context, receipt WarehouseReceipt) (WarehouseWorkset, error) {
	if err := receipt.Validate(); err != nil {
		return WarehouseWorkset{}, err
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
