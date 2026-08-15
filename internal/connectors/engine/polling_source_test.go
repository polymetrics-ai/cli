package engine

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

func TestPollingSourceExecutorResumesStrictTupleAcrossInterruptedPages(t *testing.T) {
	state := pollingSourceRuntimeState()
	firstSource := &pollingSourceFake{
		reference: pollingSourceReference(),
		evidence:  RequiredPollingWatermarkConformanceEvidence(),
		state:     state,
		pages: []PollingSourcePage{{
			Items: []PollingSourceItem{
				pollingSourceItem("a", "2026-08-06T10:00:00.000000001Z", "a"),
				pollingSourceItem("b", "2026-08-06T10:00:00.000000001Z", "b"),
			},
			More:       true,
			ObservedAt: time.Date(2026, time.August, 6, 10, 0, 1, 0, time.UTC),
		}},
	}
	first := newPollingSourceExecutor(t, firstSource)
	firstSink := &pollingSourceDurableSink{}
	firstStore := &pollingSourceCheckpointStore{}

	err := first.ReadTransport(context.Background(), pollingSourceRequest(nil), firstSink.emit(firstStore))
	if err == nil || !errors.Is(err, errPollingSourceInterrupted) {
		t.Fatalf("first ReadTransport() error = %v, want interrupted after one committed page", err)
	}
	if got, want := firstSource.requests, []synccontract.CheckpointPosition{{}, pollingSourcePosition("2026-08-06T10:00:00.000000001Z", "b")}; !reflect.DeepEqual(got, want) {
		t.Fatalf("first source tuple requests = %#v, want first request plus the durable page-boundary tuple %#v", got, want)
	}
	if got, want := firstSink.identities, []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("first destination identities = %v, want %v", got, want)
	}
	if got := len(firstStore.committed); got != 1 {
		t.Fatalf("first checkpoint commits = %d, want 1 after the delivered page", got)
	}
	firstCheckpoint := firstStore.committed[0]
	if got, want := firstCheckpoint.Position, pollingSourcePosition("2026-08-06T10:00:00.000000001Z", "b"); !reflect.DeepEqual(got, want) {
		t.Fatalf("first committed checkpoint position = %#v, want %#v", got, want)
	}

	secondSource := &pollingSourceFake{
		reference: pollingSourceReference(),
		evidence:  RequiredPollingWatermarkConformanceEvidence(),
		state:     state,
		pages: []PollingSourcePage{{
			Items:      []PollingSourceItem{pollingSourceItem("c", "2026-08-06T10:00:00.000000001Z", "c")},
			ObservedAt: time.Date(2026, time.August, 6, 10, 0, 2, 0, time.UTC),
		}},
	}
	second := newPollingSourceExecutor(t, secondSource)
	secondSink := &pollingSourceDurableSink{}
	secondStore := &pollingSourceCheckpointStore{}
	if err := second.ReadTransport(context.Background(), pollingSourceRequest(&firstCheckpoint), secondSink.emit(secondStore)); err != nil {
		t.Fatalf("resume ReadTransport(): %v", err)
	}
	if got, want := secondSource.requests, []synccontract.CheckpointPosition{pollingSourcePosition("2026-08-06T10:00:00.000000001Z", "b")}; !reflect.DeepEqual(got, want) {
		t.Fatalf("resume source tuple requests = %#v, want %#v", got, want)
	}
	combined := append(append([]string(nil), firstSink.identities...), secondSink.identities...)
	if got, want := combined, []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("combined delivered identities = %v, want exactly once %v", got, want)
	}
}

func TestPollingSourceExecutorReplaysAfterAcknowledgedCheckpointFailure(t *testing.T) {
	state := pollingSourceRuntimeState()
	prior := pollingSourceCommittedCheckpoint(t, state, pollingSourcePosition("2026-08-06T10:00:00Z", "a"))
	firstSource := &pollingSourceFake{
		reference: pollingSourceReference(), evidence: RequiredPollingWatermarkConformanceEvidence(), state: state,
		pages: []PollingSourcePage{{
			Items:      []PollingSourceItem{pollingSourceItem("b", "2026-08-06T10:00:00Z", "b")},
			ObservedAt: time.Date(2026, time.August, 6, 10, 1, 0, 0, time.UTC),
		}},
	}
	first := newPollingSourceExecutor(t, firstSource)
	firstSink := &pollingSourceDurableSink{}
	failingStore := &pollingSourceCheckpointStore{failAt: 1}
	err := first.ReadTransport(context.Background(), pollingSourceRequest(&prior), firstSink.emit(failingStore))
	if err == nil || !errors.Is(err, errPollingSourceCheckpointStore) {
		t.Fatalf("first ReadTransport() error = %v, want checkpoint persistence failure", err)
	}
	if got, want := firstSink.identities, []string{"b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("destination delivery before persistence failure = %v, want %v", got, want)
	}
	if got := len(failingStore.committed); got != 0 {
		t.Fatalf("checkpoint commits after failed persistence = %d, want 0", got)
	}

	secondSource := &pollingSourceFake{
		reference: pollingSourceReference(), evidence: RequiredPollingWatermarkConformanceEvidence(), state: state,
		pages: []PollingSourcePage{{
			Items:      []PollingSourceItem{pollingSourceItem("b", "2026-08-06T10:00:00Z", "b")},
			ObservedAt: time.Date(2026, time.August, 6, 10, 1, 1, 0, time.UTC),
		}},
	}
	second := newPollingSourceExecutor(t, secondSource)
	secondSink := &pollingSourceDurableSink{}
	secondStore := &pollingSourceCheckpointStore{}
	if err := second.ReadTransport(context.Background(), pollingSourceRequest(&prior), secondSink.emit(secondStore)); err != nil {
		t.Fatalf("replay ReadTransport(): %v", err)
	}
	if got, want := secondSource.requests, []synccontract.CheckpointPosition{prior.Position}; !reflect.DeepEqual(got, want) {
		t.Fatalf("replay source tuple requests = %#v, want prior committed %#v", got, want)
	}
	if got, want := append(firstSink.identities, secondSink.identities...), []string{"b", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("delivery attempts = %v, want at-least-once replay %v", got, want)
	}
	if got, want := secondStore.committed[0].Position, pollingSourcePosition("2026-08-06T10:00:00Z", "b"); !reflect.DeepEqual(got, want) {
		t.Fatalf("replayed committed position = %#v, want %#v", got, want)
	}
}

func TestPollingSourceExecutorCommitsEmptyPageWithoutInventingCursor(t *testing.T) {
	source := &pollingSourceFake{
		reference: pollingSourceReference(), evidence: RequiredPollingWatermarkConformanceEvidence(), state: pollingSourceRuntimeState(),
		pages: []PollingSourcePage{{ObservedAt: time.Date(2026, time.August, 16, 1, 0, 0, 0, time.UTC)}},
	}
	executor := newPollingSourceExecutor(t, source)
	sink := &pollingSourceDurableSink{}
	store := &pollingSourceCheckpointStore{}
	if err := executor.ReadTransport(context.Background(), pollingSourceRequest(nil), sink.emit(store)); err != nil {
		t.Fatalf("ReadTransport() on empty page: %v", err)
	}
	if got, want := source.fetches, 1; got != want {
		t.Fatalf("source fetches = %d, want one bounded empty-page read", got)
	}
	if got := len(sink.identities); got != 0 {
		t.Fatalf("destination deliveries = %v, want none for an empty page", sink.identities)
	}
	if got, want := len(store.committed), 1; got != want {
		t.Fatalf("empty page checkpoint commits = %d, want %d", got, want)
	}
	checkpoint := store.committed[0]
	if checkpoint.PositionObserved == nil || *checkpoint.PositionObserved {
		t.Fatalf("empty page position observation = %#v, want explicit false", checkpoint.PositionObserved)
	}
	if len(checkpoint.Position.Primary) != 0 || len(checkpoint.Position.TieBreaker) != 0 {
		t.Fatalf("empty page checkpoint invented cursor position %#v", checkpoint.Position)
	}
}

func TestPollingSourceExecutorRefusesAnUnbudgetedProviderReadBeforeDelivery(t *testing.T) {
	source := &pollingSourceFake{
		reference: pollingSourceReference(), evidence: RequiredPollingWatermarkConformanceEvidence(), state: pollingSourceRuntimeState(), skipBudget: true,
		pages: []PollingSourcePage{{
			Items:      []PollingSourceItem{pollingSourceItem("a", "2026-08-06T10:00:00Z", "a")},
			ObservedAt: time.Date(2026, time.August, 6, 10, 4, 0, 0, time.UTC),
		}},
	}
	executor := newPollingSourceExecutor(t, source)
	sink := &pollingSourceDurableSink{}
	store := &pollingSourceCheckpointStore{}
	if err := executor.ReadTransport(context.Background(), pollingSourceRequest(nil), sink.emit(store)); err == nil {
		t.Fatal("ReadTransport() error = nil, want unbudgeted provider-read refusal")
	}
	if got, want := source.fetches, 1; got != want {
		t.Fatalf("unbudgeted source fetches = %d, want the one observable attempted provider read", got)
	}
	if got := len(sink.identities); got != 0 {
		t.Fatalf("destination deliveries = %v, want no delivery from an unbudgeted read", sink.identities)
	}
	if got := len(store.committed); got != 0 {
		t.Fatalf("checkpoint commits = %d, want no mutation from an unbudgeted read", got)
	}
}

func TestPollingSourceExecutorRefusesUnsafeStateAndPageBeforeCheckpointMutation(t *testing.T) {
	state := pollingSourceRuntimeState()
	prior := pollingSourceCommittedCheckpoint(t, state, pollingSourcePosition("2026-08-06T10:00:00Z", "a"))

	tests := []struct {
		name       string
		checkpoint *synccontract.CheckpointEnvelope
		page       PollingSourcePage
		wantFetch  int
		traversal  error
	}{
		{
			name: "schema mismatch", checkpoint: checkpointWithSchema(prior, "other-schema"), wantFetch: 0,
		},
		{
			name: "source generation mismatch", checkpoint: checkpointWithGeneration(prior, "other-generation"), wantFetch: 0,
		},
		{
			name: "null cursor", checkpoint: &prior,
			page:      PollingSourcePage{Items: []PollingSourceItem{pollingSourceItem("b", "", "b")}, ObservedAt: time.Date(2026, time.August, 6, 10, 2, 0, 0, time.UTC)},
			wantFetch: 1,
		},
		{
			name: "non advancing tuple", checkpoint: &prior,
			page:      PollingSourcePage{Items: []PollingSourceItem{pollingSourceItem("a", "2026-08-06T10:00:00Z", "a")}, More: true, ObservedAt: time.Date(2026, time.August, 6, 10, 2, 0, 0, time.UTC)},
			wantFetch: 1,
		},
		{
			name: "native traversal rejects a lower opaque tie breaker", checkpoint: &prior,
			page:      PollingSourcePage{Items: []PollingSourceItem{pollingSourceItem("lower", "2026-08-06T10:00:00Z", "0")}, ObservedAt: time.Date(2026, time.August, 6, 10, 2, 0, 0, time.UTC)},
			wantFetch: 1,
			traversal: errors.New("native keyset traversal did not advance"),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			source := &pollingSourceFake{
				reference: pollingSourceReference(), evidence: RequiredPollingWatermarkConformanceEvidence(), state: state,
				pages: []PollingSourcePage{testCase.page}, traversalErr: testCase.traversal,
			}
			executor := newPollingSourceExecutor(t, source)
			sink := &pollingSourceDurableSink{}
			store := &pollingSourceCheckpointStore{}
			err := executor.ReadTransport(context.Background(), pollingSourceRequest(testCase.checkpoint), sink.emit(store))
			if err == nil {
				t.Fatal("ReadTransport() error = nil, want fail-closed refusal")
			}
			if got := source.fetches; got != testCase.wantFetch {
				t.Fatalf("source fetches = %d, want %d", got, testCase.wantFetch)
			}
			if got := source.traversalChecks; got != testCase.wantFetch {
				t.Fatalf("native traversal checks = %d, want %d after each fetched page", got, testCase.wantFetch)
			}
			if got := len(sink.identities); got != 0 {
				t.Fatalf("destination deliveries = %v, want none after refusal", sink.identities)
			}
			if got := len(store.committed); got != 0 {
				t.Fatalf("checkpoint mutations = %d, want 0 after refusal", got)
			}
		})
	}
}

func TestPollingSourceExecutorForwardsDeclaredSoftDeleteAsTombstone(t *testing.T) {
	state := pollingSourceRuntimeState()
	source := &pollingSourceFake{
		reference: pollingSourceReference(), evidence: RequiredPollingWatermarkConformanceEvidence(), state: state,
		pages: []PollingSourcePage{{
			Items: []PollingSourceItem{{
				Position: pollingSourcePosition("2026-08-06T10:00:00Z", "deleted"),
				Tombstone: &synccontract.Tombstone{
					Operation:   synccontract.OperationDelete,
					EventID:     synccontract.OpaqueToken("deleted"),
					Key:         []byte(`{"id":"deleted"}`),
					DeleteImage: synccontract.DeleteImageKeyOnly,
					Position:    pollingSourcePosition("2026-08-06T10:00:00Z", "deleted"),
				},
			}},
			ObservedAt: time.Date(2026, time.August, 6, 10, 3, 0, 0, time.UTC),
		}},
	}
	executor := newPollingSourceExecutor(t, source, withPollingDeleteVisibility())
	sink := &pollingSourceDurableSink{}
	store := &pollingSourceCheckpointStore{}
	if err := executor.ReadTransport(context.Background(), pollingSourceRequest(nil), sink.emit(store)); err != nil {
		t.Fatalf("ReadTransport(): %v", err)
	}
	if got, want := sink.tombstoneIDs, []string{"deleted"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("delivered tombstone IDs = %v, want %v", got, want)
	}
	if got := len(store.committed); got != 1 {
		t.Fatalf("checkpoint commits = %d, want 1 after tombstone delivery", got)
	}
}

var _ synctransport.SourceExecutor = (*PollingSourceExecutor)(nil)

var (
	errPollingSourceInterrupted     = errors.New("polling source interrupted")
	errPollingSourceCheckpointStore = errors.New("polling checkpoint store failed")
)

type pollingSourceFake struct {
	reference       connectors.TransportExecutorReference
	evidence        PollingWatermarkConformanceEvidence
	state           PollingSourceRuntimeState
	pages           []PollingSourcePage
	requests        []synccontract.CheckpointPosition
	fetches         int
	skipBudget      bool
	traversalChecks int
	traversalErr    error
}

func (s *pollingSourceFake) PollingSourceExecutorReference() connectors.TransportExecutorReference {
	return s.reference
}

func (s *pollingSourceFake) PollingSourceConformanceEvidence() PollingWatermarkConformanceEvidence {
	return s.evidence
}

func (s *pollingSourceFake) PollingSourceRuntimeState(context.Context, connectors.PollingCatalogObject) (PollingSourceRuntimeState, error) {
	return s.state.Clone(), nil
}

func (s *pollingSourceFake) FetchPollingSourcePage(ctx context.Context, request PollingSourcePageRequest) (PollingSourcePage, error) {
	if !s.skipBudget {
		if err := request.RequestBudget.Consume(ctx); err != nil {
			return PollingSourcePage{}, err
		}
	}
	s.fetches++
	if request.After == nil {
		s.requests = append(s.requests, synccontract.CheckpointPosition{})
	} else {
		s.requests = append(s.requests, request.After.Clone())
	}
	if s.fetches > len(s.pages) {
		return PollingSourcePage{}, errPollingSourceInterrupted
	}
	return s.pages[s.fetches-1].Clone(), nil
}

func (s *pollingSourceFake) ValidatePollingSourcePageTraversal(_ context.Context, _ *synccontract.CheckpointPosition, _ PollingSourcePage) error {
	s.traversalChecks++
	return s.traversalErr
}

type pollingSourceDurableSink struct {
	identities   []string
	tombstoneIDs []string
}

func (s *pollingSourceDurableSink) emit(store *pollingSourceCheckpointStore) func(synctransport.SourcePage) error {
	return func(page synctransport.SourcePage) error {
		for _, record := range page.Records {
			s.identities = append(s.identities, record["id"].(string))
		}
		for _, tombstone := range page.Tombstones {
			s.tombstoneIDs = append(s.tombstoneIDs, string(tombstone.EventID))
		}
		acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement("fixture-destination", page.CandidateCheckpoint.ObservedAt.Add(time.Second))
		if err != nil {
			return err
		}
		return synccontract.CommitAfterDownstreamAcknowledgement(page.CandidateCheckpoint, acknowledgement, store.Commit)
	}
}

type pollingSourceCheckpointStore struct {
	committed []synccontract.CheckpointEnvelope
	failAt    int
	calls     int
}

func (s *pollingSourceCheckpointStore) Commit(checkpoint synccontract.CheckpointEnvelope) error {
	s.calls++
	if s.failAt == s.calls {
		return errPollingSourceCheckpointStore
	}
	s.committed = append(s.committed, checkpoint.Clone())
	return nil
}

func newPollingSourceExecutor(t *testing.T, source *pollingSourceFake, options ...func(*connectors.PollingWatermarkDescriptor)) *PollingSourceExecutor {
	t.Helper()
	fixture := newPollingPreflightFixture(t)
	for _, option := range options {
		option(fixture.declaration)
	}
	if fixture.declaration.Source.DeleteVisibility == connectors.PollingDeleteVisibilityTombstone {
		fixture.object.Columns = append(fixture.object.Columns, fixture.declaration.Source.SoftDeleteField)
	}
	registry := NewPollingPreflightRegistry()
	if err := registry.RegisterSource(source); err != nil {
		t.Fatalf("RegisterSource: %v", err)
	}
	if err := registry.RegisterApply(fixture.apply); err != nil {
		t.Fatalf("RegisterApply: %v", err)
	}
	resolved, err := PollingPreflight(context.Background(), registry, fixture.declaration, fixture.object, synccontract.ModeIncrementalUpsert)
	if err != nil {
		t.Fatalf("PollingPreflight: %v", err)
	}
	executor, err := NewPollingSourceExecutor(resolved)
	if err != nil {
		t.Fatalf("NewPollingSourceExecutor: %v", err)
	}
	return executor
}

func withPollingDeleteVisibility() func(*connectors.PollingWatermarkDescriptor) {
	return func(declaration *connectors.PollingWatermarkDescriptor) {
		declaration.Source.DeleteVisibility = connectors.PollingDeleteVisibilityTombstone
		declaration.Source.SoftDeleteField = "deleted_at"
		declaration.Source.SoftDeleteAdvancesCursor = true
	}
}

func pollingSourceReference() connectors.TransportExecutorReference {
	return connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "fixture-polling-source-v1"}
}

func pollingSourceRequest(checkpoint *synccontract.CheckpointEnvelope) synctransport.SourceRequest {
	request := synctransport.SourceRequest{
		Mode:      synccontract.ModeIncrementalUpsert,
		BatchSize: 2,
		Resume: synccontract.ResumeExpectation{Source: synccontract.SourceIdentity{
			Engine: "fixture-native", AccountOrCluster: "fixture-account", ObjectScope: "widgets",
		}, SourceGeneration: synccontract.OpaqueToken("generation-1")},
	}
	if checkpoint != nil {
		copy := checkpoint.Clone()
		request.Checkpoint = &copy
	}
	return request
}

func pollingSourceRuntimeState() PollingSourceRuntimeState {
	return PollingSourceRuntimeState{
		SourceGeneration: synccontract.OpaqueToken("generation-1"),
		SchemaVersion:    "schema-v1",
		SnapshotBarrier:  synccontract.SnapshotBarrier{Kind: "transaction_snapshot", Token: synccontract.OpaqueToken("fixture-barrier")},
		Partitions:       []synccontract.PartitionState{},
		Dedupe:           synccontract.DedupeIdentity{Kind: "fixture_id", Value: synccontract.OpaqueToken("id")},
		DedupeWindow:     synccontract.DedupeWindow{Kind: "fixture_overlap", Start: synccontract.OpaqueToken("start"), End: synccontract.OpaqueToken("end")},
	}
}

func pollingSourceCommittedCheckpoint(t *testing.T, state PollingSourceRuntimeState, position synccontract.CheckpointPosition) synccontract.CheckpointEnvelope {
	t.Helper()
	positionObserved := true
	candidate := synccontract.CheckpointEnvelope{
		StateVersion: synccontract.StateVersion,
		Source: synccontract.SourceIdentity{
			Engine: "fixture-native", AccountOrCluster: "fixture-account", ObjectScope: "widgets",
		},
		Mechanism:        "polling_watermark",
		SnapshotBarrier:  &state.SnapshotBarrier,
		Position:         position,
		PositionObserved: &positionObserved,
		Partitions:       state.Partitions,
		SourceGeneration: state.SourceGeneration,
		SchemaVersion:    state.SchemaVersion,
		ProtocolVersion:  pollingSourceProtocolVersion,
		Dedupe:           state.Dedupe,
		DedupeWindow:     state.DedupeWindow,
		ObservedAt:       time.Date(2026, time.August, 6, 9, 59, 0, 0, time.UTC),
	}
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement("fixture-destination", candidate.ObservedAt.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	var committed synccontract.CheckpointEnvelope
	if err := synccontract.CommitAfterDownstreamAcknowledgement(candidate, acknowledgement, func(checkpoint synccontract.CheckpointEnvelope) error {
		committed = checkpoint
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return committed
}

func pollingSourceItem(id, watermark, tieBreaker string) PollingSourceItem {
	return PollingSourceItem{
		Record:   connectors.Record{"id": id},
		Position: pollingSourcePosition(watermark, tieBreaker),
	}
}

func pollingSourcePosition(watermark, tieBreaker string) synccontract.CheckpointPosition {
	return synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken(watermark), TieBreaker: synccontract.OpaqueToken(tieBreaker)}
}

func checkpointWithSchema(checkpoint synccontract.CheckpointEnvelope, schema string) *synccontract.CheckpointEnvelope {
	copy := checkpoint.Clone()
	copy.SchemaVersion = schema
	return &copy
}

func checkpointWithGeneration(checkpoint synccontract.CheckpointEnvelope, generation string) *synccontract.CheckpointEnvelope {
	copy := checkpoint.Clone()
	copy.SourceGeneration = synccontract.OpaqueToken(generation)
	return &copy
}

func TestPollingSourceExecutorPublicInputsExcludeRawProtocolFields(t *testing.T) {
	for _, value := range []any{
		PollingSourcePageRequest{},
		PollingSourceRuntimeState{},
		PollingSourceItem{},
	} {
		typ := reflect.TypeOf(value)
		for _, forbidden := range []string{"SQL", "Query", "URL", "Method", "Body", "Path", "Command", "Shell"} {
			if _, found := typ.FieldByName(forbidden); found {
				t.Fatalf("%s exposes forbidden raw protocol field %q", typ.Name(), forbidden)
			}
		}
	}
	if bytes.Equal(pollingSourcePosition("a", "b").Primary, []byte("")) {
		t.Fatal("test position unexpectedly lost its opaque primary value")
	}
}
