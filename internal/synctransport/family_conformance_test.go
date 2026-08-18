package synctransport

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

func TestTransportFamilyHalfPathConformance(t *testing.T) {
	families := transportFamilyConformanceCases()
	requiredFamilies := map[string]struct{}{
		"api_source_to_warehouse":           {},
		"database_source_to_warehouse":      {},
		"warehouse_to_api_destination":      {},
		"warehouse_to_database_destination": {},
	}
	for _, family := range families {
		if _, required := requiredFamilies[family.name]; !required {
			t.Fatalf("unexpected or duplicate transport family case %q", family.name)
		}
		delete(requiredFamilies, family.name)
	}
	if len(requiredFamilies) != 0 {
		t.Fatalf("missing transport family cases: %v", requiredFamilies)
	}

	for _, family := range families {
		for _, mode := range synccontract.AllModes() {
			family := family
			mode := mode
			t.Run(family.name+"/"+string(mode), func(t *testing.T) {
				pair := newTestTransportPair(family.sourceType, family.destinationType)
				if got, want := pair.source.Metadata().IntegrationType, family.expectedSourceType; got != want {
					t.Fatalf("source family = %q, want %q", got, want)
				}
				if got, want := pair.destination.Metadata().IntegrationType, family.expectedDestinationType; got != want {
					t.Fatalf("destination family = %q, want %q", got, want)
				}
				strategy, err := configureTransportFamilyMode(pair, mode)
				if err != nil {
					t.Fatal(err)
				}
				recordID := family.name + "-" + string(mode)
				pair.sourceExecutor.pages = []SourcePage{{
					Records: []connectors.Record{{
						"id":            recordID,
						"source_family": family.sourceType,
					}},
					CandidateCheckpoint: testCheckpoint(pair.source.Name()),
				}}

				registry := NewRegistry(pair.verifier)
				registerTransportPair(t, registry, pair)
				stage := &testWarehouseStage{}
				commits := make([]synccontract.CheckpointEnvelope, 0, 1)
				result, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
					Source:      pair.source,
					Destination: pair.destination,
					Stream:      "records",
					Mode:        mode,
					BatchSize:   1,
					Stage:       stage,
					Commit: func(checkpoint synccontract.CheckpointEnvelope) error {
						commits = append(commits, checkpoint.Clone())
						return nil
					},
				})
				if err != nil {
					t.Fatalf("Run() = %v", err)
				}
				if got, want := result.RecordsRead, 1; got != want {
					t.Fatalf("records read = %d, want %d", got, want)
				}
				if got, want := result.RecordsStaged, 1; got != want {
					t.Fatalf("records staged = %d, want %d", got, want)
				}
				if got, want := result.RecordsApplied, 1; got != want {
					t.Fatalf("records applied = %d, want %d", got, want)
				}
				if got, want := result.Pages, 1; got != want {
					t.Fatalf("pages = %d, want %d", got, want)
				}
				if got, want := stage.lastPage.Records[0]["id"], recordID; got != want {
					t.Fatalf("warehouse-staged record ID = %v, want %q", got, want)
				}
				if got, want := pair.destinationExecutor.lastPlan.ApplyStrategy, strategy; got != want {
					t.Fatalf("destination plan strategy = %#v, want %#v", got, want)
				}
				if got, want := len(commits), 1; got != want {
					t.Fatalf("checkpoint commits = %d, want %d after destination acknowledgement", got, want)
				}
				if got, want := commits[0].Position.Primary, pair.sourceExecutor.pages[0].CandidateCheckpoint.Position.Primary; !bytes.Equal(got, want) {
					t.Fatalf("committed checkpoint primary = %x, want %x", got, want)
				}
				if result.CommittedCheckpoint == nil {
					t.Fatal("result omitted committed checkpoint after durable acknowledgement")
				}

				switch family.focus {
				case transportFamilySource:
					if got, want := pair.sourceExecutor.readCalls, 1; got != want {
						t.Fatalf("%s source reads = %d, want %d", family.name, got, want)
					}
					if got, want := stage.calls, 1; got != want {
						t.Fatalf("%s warehouse stages = %d, want %d", family.name, got, want)
					}
				case transportFamilyDestination:
					assertTransportFamilyDestinationValue(t, pair, mode, recordID)
				default:
					t.Fatalf("family %q has unknown focus %q", family.name, family.focus)
				}
			})
		}
	}
}

func TestTransportFamilyHalfPathConformanceRefusesUnboundSourceBeforeIO(t *testing.T) {
	pair := newTestTransportPair("api", "database")
	pair.destination.descriptor.Destination.SourceBindings = []connectors.DestinationSourceBinding{{
		Executor:        connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeAPI, ID: "unbound_source"},
		EligibleStreams: []string{"records"},
		RecordMapping: connectors.SourceRecordMapping{
			Kind:        connectors.SourceRecordMappingKindConfigMatch,
			ConfigKey:   "fixture_source",
			RecordField: "source_family",
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
		BatchSize:   1,
		Stage:       stage,
		Commit: func(synccontract.CheckpointEnvelope) error {
			commits++
			return nil
		},
	})
	var ineligible *DestinationSourceIneligibleError
	if !errors.As(err, &ineligible) {
		t.Fatalf("Run() error = %T %v, want DestinationSourceIneligibleError", err, err)
	}
	if got, want := ineligible.SourceExecutor, pair.sourceExecutor.reference; got != want {
		t.Fatalf("refused source = %#v, want %#v", got, want)
	}
	if pair.sourceExecutor.readCalls != 0 || stage.calls != 0 || pair.destinationExecutor.planCalls != 0 || pair.destinationExecutor.applyCalls != 0 || commits != 0 {
		t.Fatalf("unbound-source refusal source/stage/plan/apply/commit = %d/%d/%d/%d/%d, want zero before I/O", pair.sourceExecutor.readCalls, stage.calls, pair.destinationExecutor.planCalls, pair.destinationExecutor.applyCalls, commits)
	}
}

func TestTransportFamilyHalfPathConformanceAcknowledgementAndCancellation(t *testing.T) {
	t.Run("missing durable acknowledgement refuses checkpoint commit", func(t *testing.T) {
		pair := newTestTransportPair("api", "database")
		pair.destinationExecutor.acknowledgementSet = true
		registry := NewRegistry(pair.verifier)
		registerTransportPair(t, registry, pair)
		commits := 0

		_, err := NewOrchestrator(registry).Run(context.Background(), RunRequest{
			Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
			BatchSize: 1, Stage: &testWarehouseStage{}, Commit: func(synccontract.CheckpointEnvelope) error {
				commits++
				return nil
			},
		})
		if !errors.Is(err, synccontract.ErrDownstreamAcknowledgementRequired) {
			t.Fatalf("Run() error = %T %v, want ErrDownstreamAcknowledgementRequired", err, err)
		}
		if got, want := pair.destinationExecutor.applyCalls, 1; got != want {
			t.Fatalf("destination applies before acknowledgement validation = %d, want %d", got, want)
		}
		if commits != 0 {
			t.Fatalf("checkpoint commits after missing acknowledgement = %d, want zero", commits)
		}
	})

	t.Run("cancellation after warehouse stage prevents destination I/O", func(t *testing.T) {
		pair := newTestTransportPair("database", "api")
		registry := NewRegistry(pair.verifier)
		registerTransportPair(t, registry, pair)
		ctx, cancel := context.WithCancel(context.Background())
		stage := &testWarehouseStage{afterStage: cancel}
		commits := 0

		_, err := NewOrchestrator(registry).Run(ctx, RunRequest{
			Source: pair.source, Destination: pair.destination, Stream: "records", Mode: synccontract.ModeFullAppend,
			BatchSize: 1, Stage: stage, Commit: func(synccontract.CheckpointEnvelope) error {
				commits++
				return nil
			},
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run() error = %T %v, want context.Canceled", err, err)
		}
		if pair.destinationExecutor.applyCalls != 0 || commits != 0 {
			t.Fatalf("cancellation apply/commit = %d/%d, want zero", pair.destinationExecutor.applyCalls, commits)
		}
	})
}

type transportFamilyFocus string

const (
	transportFamilySource      transportFamilyFocus = "source"
	transportFamilyDestination transportFamilyFocus = "destination"
)

type transportFamilyConformanceCase struct {
	name                    string
	sourceType              string
	destinationType         string
	expectedSourceType      string
	expectedDestinationType string
	focus                   transportFamilyFocus
}

func transportFamilyConformanceCases() []transportFamilyConformanceCase {
	return []transportFamilyConformanceCase{
		{name: "api_source_to_warehouse", sourceType: "api", destinationType: "api", expectedSourceType: "api", expectedDestinationType: "api", focus: transportFamilySource},
		{name: "database_source_to_warehouse", sourceType: "database", destinationType: "database", expectedSourceType: "database", expectedDestinationType: "database", focus: transportFamilySource},
		{name: "warehouse_to_api_destination", sourceType: "api", destinationType: "api", expectedSourceType: "api", expectedDestinationType: "api", focus: transportFamilyDestination},
		{name: "warehouse_to_database_destination", sourceType: "database", destinationType: "database", expectedSourceType: "database", expectedDestinationType: "database", focus: transportFamilyDestination},
	}
}

func configureTransportFamilyMode(pair *testTransportPair, mode synccontract.Mode) (connectors.DestinationApplyStrategy, error) {
	strategy, err := transportFamilyApplyStrategy(mode)
	if err != nil {
		return connectors.DestinationApplyStrategy{}, err
	}
	action := "fixture_" + string(mode)
	pair.source.descriptor.Source.Modes = []synccontract.Mode{mode}
	pair.destination.descriptor.Destination.Modes = []synccontract.Mode{mode}
	pair.destination.descriptor.Destination.EligibleActions = []string{action}
	pair.destination.descriptor.Destination.ApplyStrategies = []connectors.DestinationApplyStrategy{{
		Mode: mode, Strategy: strategy, Action: action,
	}}
	if mode == synccontract.ModeFullOverwrite {
		pair.destinationExecutor.fullOverwrite = &testFullOverwriteRun{sink: pair.destination.Name()}
	}
	return pair.destination.descriptor.Destination.ApplyStrategies[0], nil
}

func transportFamilyApplyStrategy(mode synccontract.Mode) (connectors.ApplyStrategy, error) {
	switch mode {
	case synccontract.ModeFullOverwrite:
		return connectors.ApplyStrategyReplace, nil
	case synccontract.ModeFullAppend, synccontract.ModeIncrementalAppend:
		return connectors.ApplyStrategyAppend, nil
	case synccontract.ModeIncrementalUpsert:
		return connectors.ApplyStrategyMerge, nil
	case synccontract.ModeIncrementalDedupe:
		return connectors.ApplyStrategyDedupe, nil
	case synccontract.ModeIncrementalDedupeHistory:
		return connectors.ApplyStrategyDedupeHistory, nil
	case synccontract.ModeChangeCapture:
		return connectors.ApplyStrategyChangeApply, nil
	default:
		return "", errors.New("unknown canonical sync mode")
	}
}

func assertTransportFamilyDestinationValue(t *testing.T, pair *testTransportPair, mode synccontract.Mode, recordID string) {
	t.Helper()
	if mode == synccontract.ModeFullOverwrite {
		if pair.destinationExecutor.fullOverwrite == nil {
			t.Fatal("full-overwrite destination run is absent")
		}
		if got, want := pair.destinationExecutor.fullOverwrite.ids, []string{recordID}; len(got) != len(want) || got[0] != want[0] {
			t.Fatalf("full-overwrite destination IDs = %v, want %v", got, want)
		}
		if got, want := pair.destinationExecutor.fullOverwrite.publishCalls, 1; got != want {
			t.Fatalf("full-overwrite publishes = %d, want %d", got, want)
		}
		return
	}
	if got, want := pair.destinationExecutor.applyCalls, 1; got != want {
		t.Fatalf("destination applies = %d, want %d", got, want)
	}
	if got, want := pair.destinationExecutor.readBackCalls, 1; got != want {
		t.Fatalf("destination read-backs = %d, want %d", got, want)
	}
	if got, want := pair.destinationExecutor.lastApply.Workset.Records[0]["id"], recordID; got != want {
		t.Fatalf("destination-applied record ID = %v, want %q", got, want)
	}
}
