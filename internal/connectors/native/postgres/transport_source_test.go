package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgtype"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

func TestPostgresDefinitionDeclaresResumablePollingTransportSource(t *testing.T) {
	connector := New()
	descriptor, ok := connectors.SourceTransportDescriptorOf(connector)
	if !ok {
		t.Fatal("PostgreSQL definition has no declared source transport")
	}
	wantReference := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_polling_watermark"}
	if descriptor.Executor != wantReference {
		t.Fatalf("PostgreSQL source executor = %#v, want %#v", descriptor.Executor, wantReference)
	}
	if got, want := descriptor.EligibleStreams, []string{"*"}; !sameStrings(got, want) {
		t.Fatalf("PostgreSQL source streams = %#v, want %#v", got, want)
	}
	if got, want := descriptor.Modes, []synccontract.Mode{
		synccontract.ModeFullOverwrite,
		synccontract.ModeFullAppend,
		synccontract.ModeIncrementalAppend,
		synccontract.ModeIncrementalUpsert,
		synccontract.ModeIncrementalDedupe,
		synccontract.ModeIncrementalDedupeHistory,
	}; !sameModes(got, want) {
		t.Fatalf("PostgreSQL source modes = %#v, want %#v", got, want)
	}
	if descriptor.Delivery.Deletes != connectors.DeliveryDeletesUnavailable {
		t.Fatalf("PostgreSQL polling source delete delivery = %q, want hard-delete-invisible", descriptor.Delivery.Deletes)
	}
	destination, ok := connectors.DestinationTransportDescriptorOf(connector)
	if !ok {
		t.Fatal("PostgreSQL definition has no declared managed destination transport")
	}
	wantDestination := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_managed_target"}
	if destination.Executor != wantDestination {
		t.Fatalf("PostgreSQL destination executor = %#v, want %#v", destination.Executor, wantDestination)
	}
	if got, want := destination.Modes, []synccontract.Mode{
		synccontract.ModeFullOverwrite,
		synccontract.ModeFullAppend,
		synccontract.ModeIncrementalAppend,
		synccontract.ModeIncrementalUpsert,
		synccontract.ModeIncrementalDedupe,
	}; !sameModes(got, want) {
		t.Fatalf("PostgreSQL destination modes = %#v, want %#v", got, want)
	}
	for _, forbidden := range []synccontract.Mode{synccontract.ModeIncrementalDedupeHistory, synccontract.ModeChangeCapture} {
		for _, advertised := range destination.Modes {
			if advertised == forbidden {
				t.Fatalf("PostgreSQL advertised non-executable destination mode %q", forbidden)
			}
		}
	}
	factories := connector.SyncTransportDefinitionFactories()
	foundSource := false
	foundDestination := false
	for _, factory := range factories {
		if factory.Reference == wantReference && factory.BuildSource != nil {
			foundSource = true
		}
		if factory.Reference == wantDestination && factory.BuildDestination != nil {
			foundDestination = true
		}
	}
	if !foundSource {
		t.Fatal("PostgreSQL polling source has no production definition factory")
	}
	if !foundDestination {
		t.Fatal("PostgreSQL managed destination has no production definition factory")
	}
}

func TestBootstrapTransportCommitterPublishesTransactionsOnlyAtDurableCommit(t *testing.T) {
	source := postgresCDCSource{
		identity:   synccontract.SourceIdentity{Engine: "postgres", AccountOrCluster: "system-one:analytics", ObjectScope: "public.events"},
		generation: synccontract.OpaqueToken("7\npm_events"), publication: "pm_events", schemaFingerprint: "schema-one",
		system: pglogrepl.IdentifySystemResult{SystemID: "system-one", DBName: "analytics", Timeline: 7},
	}
	barrierLSN := pglogrepl.LSN(0x16B6C50)
	barrier, err := newPostgresBootstrapBarrier(source, barrierLSN)
	if err != nil {
		t.Fatal(err)
	}
	source.bootstrap = &barrier
	nativeInitial := postgresCDCCheckpointForLSNs(source, barrierLSN, barrierLSN, barrierLSN, barrierLSN)
	resume := synccontract.ResumeExpectation{
		Source:           synccontract.SourceIdentity{Engine: "postgres", AccountOrCluster: "credential-id", ObjectScope: "public.events"},
		SourceGeneration: synccontract.OpaqueToken("app-generation"),
	}
	translatedInitial := bootstrapTransportCheckpoint(nativeInitial, resume)
	var pages []synctransport.SourcePage
	committer := &bootstrapTransportCommitter{
		resume: resume, primaryKey: []string{"id"}, finalSnapshot: &translatedInitial,
		emit: func(page synctransport.SourcePage) error {
			pages = append(pages, page)
			return nil
		},
	}
	if err := committer.CommitDurableChangefeedCheckpoint(context.Background(), nativeInitial); err != nil {
		t.Fatalf("initial barrier commit = %v", err)
	}
	if len(pages) != 0 {
		t.Fatalf("initial barrier emitted a second snapshot page: %#v", pages)
	}
	if err := committer.emitChange(connectors.CDCEvent{Operation: "insert", Record: connectors.Record{"id": int64(1), "value": "inserted"}}); err != nil {
		t.Fatal(err)
	}
	if err := committer.emitChange(connectors.CDCEvent{Operation: "delete", Record: connectors.Record{"id": int64(2)}, State: connectors.Record{"lsn": "0/16B6D00"}}); err != nil {
		t.Fatal(err)
	}
	if len(pages) != 0 {
		t.Fatal("CDC events reached the transport before their transaction checkpoint commit")
	}
	nativeNext := postgresCDCCheckpointForLSNs(source, barrierLSN, barrierLSN, pglogrepl.LSN(0x16B6C80), pglogrepl.LSN(0x16B6D00))
	if err := committer.CommitDurableChangefeedCheckpoint(context.Background(), nativeNext); err != nil {
		t.Fatalf("transaction checkpoint commit = %v", err)
	}
	if len(pages) != 1 || len(pages[0].Records) != 1 || len(pages[0].Tombstones) != 1 || pages[0].CandidateCheckpoint.Source != resume.Source || string(pages[0].CandidateCheckpoint.SourceGeneration) != string(resume.SourceGeneration) {
		t.Fatalf("committed CDC transport page = %#v, want one record/tombstone bound to App resume identity", pages)
	}
}

func TestNativeBootstrapTransportCheckpointRestoresSealedPostgresIdentity(t *testing.T) {
	source := postgresCDCSource{
		identity:   synccontract.SourceIdentity{Engine: "postgres", AccountOrCluster: "system-one:analytics", ObjectScope: "public.events"},
		generation: synccontract.OpaqueToken("7\npm_events"), publication: "pm_events", schemaFingerprint: "schema-one",
		system: pglogrepl.IdentifySystemResult{SystemID: "system-one", DBName: "analytics", Timeline: 7},
	}
	lsn := pglogrepl.LSN(0x16B6C50)
	barrier, err := newPostgresBootstrapBarrier(source, lsn)
	if err != nil {
		t.Fatal(err)
	}
	source.bootstrap = &barrier
	native := postgresCDCCheckpointForLSNs(source, lsn, lsn, lsn, lsn)
	appResume := synccontract.ResumeExpectation{
		Source:           synccontract.SourceIdentity{Engine: "postgres", AccountOrCluster: "credential-id", ObjectScope: "public.events"},
		SourceGeneration: synccontract.OpaqueToken("app-generation"),
	}
	persisted := bootstrapTransportCheckpoint(native, appResume)
	restored, err := nativeBootstrapTransportCheckpoint(persisted, connectors.RuntimeConfig{
		Config:  map[string]string{"host": "localhost", "database": "analytics", "username": "pm", "sslmode": "disable"},
		Secrets: map[string]string{"password": "fixture"},
	})
	if err != nil {
		t.Fatalf("nativeBootstrapTransportCheckpoint() = %v", err)
	}
	if restored.Source != source.identity || string(restored.SourceGeneration) != string(source.generation) || string(restored.Position.Primary) != string(native.Position.Primary) {
		t.Fatalf("restored native checkpoint = %#v, want source/generation/LSN from sealed barrier", restored)
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
			name:     "missing descriptor",
			source:   newPreflightSpyConnector(nil),
			register: false,
			wantErr:  "has no declared source transport",
		},
		{
			name: "wrong connector family",
			source: newPreflightSpyConnector(&connectors.SourceTransportDescriptor{
				Executor:        connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeAPI, ID: "postgres_polling_watermark"},
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
			wantErr:  "incompatible with transport executor family",
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

func TestRegisterPostgresPollingTransportSourceMakesDefinitionSelectedSourceReachable(t *testing.T) {
	registry := synctransport.NewRegistry(postgresTransportTestVerifier{})
	connector := New()
	if err := RegisterPollingTransportSource(registry, connector); err != nil {
		t.Fatalf("RegisterSnapshotTransportSource() error = %v", err)
	}
	destination := newPreflightDestinationConnector()
	if err := registry.RegisterDestination(&preflightSpyDestinationExecutor{reference: destination.destination.Executor}); err != nil {
		t.Fatalf("RegisterDestination() error = %v", err)
	}

	resolved, err := registry.Preflight(synctransport.PreflightRequest{
		Source:      connector,
		Destination: destination,
		Stream:      "snapshot",
		Mode:        synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatalf("Preflight() error = %v", err)
	}
	if got := resolved.Source.TransportExecutorReference(); got != (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_polling_watermark"}) {
		t.Fatalf("resolved source executor = %#v", got)
	}
}

// Happy path: this is the source half used by the shipped transport route,
// not a hand-assembled engine runner. A durable first fixture page resumes at
// its complete tuple and returns only the remaining row.
func TestPostgresPollingTransportResumesFixtureCursor(t *testing.T) {
	connector := New()
	source := NewPollingTransportSource(connector)
	request := postgresPollingFixtureRequest()

	var first synccontract.CheckpointEnvelope
	errInterrupted := errors.New("stop after durable first page")
	err := source.ReadTransport(context.Background(), request, func(page synctransport.SourcePage) error {
		if got, want := len(page.Records), 2; got != want {
			t.Fatalf("first polling page records = %d, want %d", got, want)
		}
		if got, want := page.CandidateCheckpoint.Mechanism, "polling_watermark"; got != want {
			t.Fatalf("first polling checkpoint mechanism = %q, want %q", got, want)
		}
		ack, ackErr := synccontract.NewDurableDownstreamAcknowledgement("fixture-warehouse", page.CandidateCheckpoint.ObservedAt.Add(time.Second))
		if ackErr != nil {
			return ackErr
		}
		return errors.Join(
			synccontract.CommitAfterDownstreamAcknowledgement(page.CandidateCheckpoint, ack, func(checkpoint synccontract.CheckpointEnvelope) error {
				first = checkpoint.Clone()
				return nil
			}),
			errInterrupted,
		)
	})
	if !errors.Is(err, errInterrupted) {
		t.Fatalf("first polling read error = %v, want durable interruption", err)
	}

	request.Checkpoint = &first
	var resumed []connectors.Record
	if err := source.ReadTransport(context.Background(), request, func(page synctransport.SourcePage) error {
		resumed = append(resumed, page.Records...)
		return nil
	}); err != nil {
		t.Fatalf("resumed polling read = %v", err)
	}
	if got, want := len(resumed), 1; got != want {
		t.Fatalf("resumed polling records = %#v, want exactly the uncommitted tail", resumed)
	}
	if got, want := resumed[0]["id"], 3; got != want {
		t.Fatalf("resumed polling row id = %#v, want %#v", got, want)
	}
}

// Bad path: a cursor-required PostgreSQL mode is rejected before any runtime
// configuration or provider operation is attempted, so it cannot silently
// treat an existing checkpoint as an unfiltered full scan.
func TestPostgresPollingTransportRefusesMissingPerStreamCursorBeforeIO(t *testing.T) {
	source := NewPollingTransportSource(New())
	request := postgresPollingFixtureRequest()
	request.CursorField = ""
	request.Runtime = connectors.RuntimeConfig{}

	err := source.ReadTransport(context.Background(), request, func(synctransport.SourcePage) error { return nil })
	if !errors.Is(err, ErrPollingCursorFieldRequired) {
		t.Fatalf("missing per-stream cursor error = %v, want %v", err, ErrPollingCursorFieldRequired)
	}
}

// Edge case: nullable columns cannot form a complete resumable cursor tuple.
// The catalog planner must refuse them rather than using SQL's NULL ordering
// to omit those source rows.
func TestPostgresPollingReadPlanRefusesNullableCursor(t *testing.T) {
	logical, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatal(err)
	}
	relation := database.RelationRef{Schema: database.SchemaRef{Catalog: database.CatalogRef{Name: "analytics"}, Name: "public"}, Name: "events"}
	id := database.ColumnRef{Relation: relation, Name: "id"}
	updatedAt := database.ColumnRef{Relation: relation, Name: "updated_at"}
	catalog, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, []database.Relation{{
		Ref:            relation,
		NativeIdentity: database.NativeRelationIdentity{Kind: "oid", Value: "10001"},
		Columns: []database.Column{
			{Ref: id, Type: logical, Nullable: false, Ordinal: 1},
			{Ref: updatedAt, Type: logical, Nullable: true, Ordinal: 2},
		},
		Keys: []database.Key{{Name: "events_pkey", Kind: database.KeyPrimary, Columns: []database.ColumnRef{id}}},
	}})
	if err != nil {
		t.Fatal(err)
	}

	_, err = newPostgresPollingReadPlan(catalog, relation, "updated_at", []string{"id"}, 2)
	if !errors.Is(err, ErrPollingCursorNullable) {
		t.Fatalf("nullable cursor plan error = %v, want %v", err, ErrPollingCursorNullable)
	}
}

// Edge case: a persisted checkpoint from a private snapshot protocol must be
// rejected with the shared typed recovery outcome, never cleared and retried
// as a new full read.
func TestPostgresPollingTransportRefusesInvalidCheckpointWithoutRestart(t *testing.T) {
	source := NewPollingTransportSource(New())
	request := postgresPollingFixtureRequest()
	// An invalid checkpoint is rejected before runtime configuration, pool
	// setup, catalog discovery, or a polling page. The empty config makes that
	// ordering observable without a live server.
	request.Runtime = connectors.RuntimeConfig{}
	observed := true
	request.Checkpoint = &synccontract.CheckpointEnvelope{
		StateVersion:     synccontract.StateVersion,
		Source:           request.Resume.Source,
		Mechanism:        "postgres_bounded_full_snapshot",
		SnapshotBarrier:  &synccontract.SnapshotBarrier{Kind: "none", Token: synccontract.OpaqueToken("postgres-polling-v1")},
		Position:         synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("2000"), TieBreaker: synccontract.OpaqueToken("2")},
		PositionObserved: &observed,
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: request.Resume.SourceGeneration,
		SchemaVersion:    "fixture-postgres-users-v1",
		ProtocolVersion:  "postgres_snapshot_v1",
		Dedupe:           synccontract.DedupeIdentity{Kind: "postgres_polling_tuple", Value: synccontract.OpaqueToken("fixture")},
		DedupeWindow:     synccontract.DedupeWindow{Kind: "postgres_polling_overlap", Start: synccontract.OpaqueToken("fixture"), End: synccontract.OpaqueToken("fixture")},
		ObservedAt:       time.Date(2026, time.August, 16, 0, 0, 0, 0, time.UTC),
	}
	committed := request.Checkpoint.ObservedAt.Add(time.Second)
	request.Checkpoint.CommittedAt = &committed

	var emitted int
	err := source.ReadTransport(context.Background(), request, func(synctransport.SourcePage) error {
		emitted++
		return nil
	})
	var recovery *synccontract.RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeInvalidCheckpoint {
		t.Fatalf("invalid checkpoint error = %v, want invalid-checkpoint rebootstrap", err)
	}
	if emitted != 0 {
		t.Fatalf("invalid checkpoint emitted %d restarted pages, want 0", emitted)
	}
}

// Edge case: a checkpoint whose persisted catalog fingerprint no longer
// matches the selected stream must not restart at page one. The fixture runner
// makes the assertion hermetic while retaining shared-executor validation.
func TestPostgresPollingTransportRefusesStaleSchemaCheckpointBeforePageRead(t *testing.T) {
	source := NewPollingTransportSource(New())
	request := postgresPollingFixtureRequest()
	checkpoint := validPostgresPollingFixtureCheckpoint(t, source, request)
	checkpoint.SchemaVersion = "stale-postgres-schema"
	request.Checkpoint = &checkpoint

	emitted := 0
	err := source.ReadTransport(context.Background(), request, func(synctransport.SourcePage) error {
		emitted++
		return nil
	})
	var recovery *synccontract.RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeInvalidCheckpoint || !strings.Contains(err.Error(), "schema fingerprint") {
		t.Fatalf("stale schema checkpoint error = %v, want named invalid-checkpoint rebootstrap", err)
	}
	if emitted != 0 {
		t.Fatalf("stale schema checkpoint emitted %d restarted pages, want 0", emitted)
	}
}

func validPostgresPollingFixtureCheckpoint(t *testing.T, source *PollingTransportSource, request synctransport.SourceRequest) synccontract.CheckpointEnvelope {
	t.Helper()
	var committed synccontract.CheckpointEnvelope
	errStop := errors.New("stop after first fixture checkpoint")
	err := source.ReadTransport(context.Background(), request, func(page synctransport.SourcePage) error {
		ack, err := synccontract.NewDurableDownstreamAcknowledgement("fixture-warehouse", page.CandidateCheckpoint.ObservedAt.Add(time.Second))
		if err != nil {
			return err
		}
		if err := synccontract.CommitAfterDownstreamAcknowledgement(page.CandidateCheckpoint, ack, func(checkpoint synccontract.CheckpointEnvelope) error {
			committed = checkpoint.Clone()
			return nil
		}); err != nil {
			return err
		}
		return errStop
	})
	if !errors.Is(err, errStop) {
		t.Fatalf("produce valid fixture polling checkpoint: %v", err)
	}
	return committed
}

func postgresPollingFixtureRequest() synctransport.SourceRequest {
	return synctransport.SourceRequest{
		Connector:   New(),
		Runtime:     connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "host": "fixture.internal", "database": "analytics", "username": "reader", "sslmode": "require"}},
		Stream:      "public.users",
		CursorField: "updated_at",
		PrimaryKey:  []string{"id"},
		Mode:        synccontract.ModeIncrementalAppend,
		BatchSize:   2,
		Resume: synccontract.ResumeExpectation{
			Source:           synccontract.SourceIdentity{Engine: "postgres", AccountOrCluster: "fixture-credential", ObjectScope: "public.users"},
			SourceGeneration: synccontract.OpaqueToken("fixture-generation"),
		},
	}
}

func TestPostgresSnapshotReadPlanAndCheckpointUseTypedStableIdentity(t *testing.T) {
	logical, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatalf("NewSignedInteger() error = %v", err)
	}
	relationRef := database.RelationRef{
		Schema: database.SchemaRef{Catalog: database.CatalogRef{Name: "analytics"}, Name: "public"},
		Name:   "events",
	}
	id := database.ColumnRef{Relation: relationRef, Name: "id"}
	sequence := database.ColumnRef{Relation: relationRef, Name: "sequence"}
	label := database.ColumnRef{Relation: relationRef, Name: "label"}
	catalog, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, []database.Relation{{
		Ref: relationRef,
		NativeIdentity: database.NativeRelationIdentity{
			Kind:  "oid",
			Value: "10001",
		},
		Columns: []database.Column{
			{Ref: label, Type: newTransportTestString(t, 0, ""), Nullable: false, Ordinal: 3},
			{Ref: sequence, Type: logical, Nullable: false, Ordinal: 2},
			{Ref: id, Type: logical, Nullable: false, Ordinal: 1},
		},
		Keys: []database.Key{{Name: "events_pkey", Kind: database.KeyPrimary, Columns: []database.ColumnRef{id}}, {
			Name: "events_sequence_key", Kind: database.KeyUnique, Columns: []database.ColumnRef{sequence},
		}},
	}})
	if err != nil {
		t.Fatalf("NewCatalog() error = %v", err)
	}

	plan, err := newPostgresSnapshotReadPlan(catalog, relationRef, 2)
	if err != nil {
		t.Fatalf("newPostgresSnapshotReadPlan() error = %v", err)
	}
	if got, want := []string{plan.columns[0].Ref.Name, plan.columns[1].Ref.Name, plan.columns[2].Ref.Name}, []string{"id", "sequence", "label"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("selected catalog columns = %#v, want %#v", got, want)
	}
	if got, want := plan.order, []database.ColumnRef{id}; !reflect.DeepEqual(got, want) {
		t.Fatalf("stable read order = %#v, want %#v", got, want)
	}
	if got, want := plan.query([]any{int64(2)}), `SELECT "id", "sequence", "label" FROM "public"."events" WHERE ("id") > ($1) ORDER BY "id" ASC LIMIT $2`; got != want {
		t.Fatalf("bounded SQL = %q, want %q", got, want)
	}
	if got, want := plan.queryArguments([]any{int64(2)}), []any{int64(2), 2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("bounded SQL arguments = %#v, want %#v", got, want)
	}

	resume := synccontract.ResumeExpectation{
		Source:           synccontract.SourceIdentity{Engine: "postgres", AccountOrCluster: "cluster-alpha", ObjectScope: "analytics.public.events"},
		SourceGeneration: synccontract.OpaqueToken("generation-1"),
	}
	first, err := postgresSnapshotCheckpoint(resume, "742:742:", catalog.Fingerprint(), 0)
	if err != nil {
		t.Fatalf("postgresSnapshotCheckpoint() error = %v", err)
	}
	second, err := postgresSnapshotCheckpoint(resume, "742:742:", catalog.Fingerprint(), 0)
	if err != nil {
		t.Fatalf("postgresSnapshotCheckpoint() second error = %v", err)
	}
	if err := first.Validate(); err != nil {
		t.Fatalf("checkpoint Validate() error = %v", err)
	}
	if first.Source != resume.Source || first.SchemaVersion != catalog.Fingerprint().String() || first.SnapshotBarrier == nil || string(first.SnapshotBarrier.Token) != "742:742:" || first.PositionObserved == nil || *first.PositionObserved {
		t.Fatalf("checkpoint = %#v, want exact full-snapshot identity/schema/barrier", first)
	}
	if !reflect.DeepEqual(first.Dedupe, second.Dedupe) || string(first.DedupeWindow.Start) != "742:742:" || string(first.DedupeWindow.End) != "742:742:" {
		t.Fatalf("checkpoint dedupe boundary was not deterministic: %#v %#v", first.Dedupe, first.DedupeWindow)
	}
}

func TestPostgresSnapshotRecordValuesNormalizeTransportVocabulary(t *testing.T) {
	relation := database.RelationRef{
		Schema: database.SchemaRef{Catalog: database.CatalogRef{Name: "analytics"}, Name: "public"},
		Name:   "events",
	}
	uuid := [16]byte{0x54, 0x85, 0xb5, 0x8f, 0x40, 0x11, 0x4b, 0xd2, 0x8a, 0x9e, 0xf7, 0x80, 0x1e, 0x12, 0x7a, 0x91}
	timestamp := time.Date(2026, time.August, 14, 12, 34, 56, 789000000, time.FixedZone("UTC+5:30", 5*60*60+30*60))
	date := time.Date(2026, time.August, 14, 0, 0, 0, 0, time.UTC)
	plan := postgresSnapshotReadPlan{
		columns: []database.Column{
			{Ref: database.ColumnRef{Relation: relation, Name: "event_id"}, Type: database.NewUUID()},
			{Ref: database.ColumnRef{Relation: relation, Name: "occurred_at"}, Type: mustTransportTimestamp(t, true)},
			{Ref: database.ColumnRef{Relation: relation, Name: "recorded_at"}, Type: mustTransportTimestamp(t, false)},
			{Ref: database.ColumnRef{Relation: relation, Name: "event_date"}, Type: database.NewDate()},
			{Ref: database.ColumnRef{Relation: relation, Name: "amount"}, Type: mustTransportDecimal(t)},
			{Ref: database.ColumnRef{Relation: relation, Name: "opened_at"}, Type: mustTransportTime(t)},
		},
		order: []database.ColumnRef{{Relation: relation, Name: "event_id"}},
	}
	values := []any{
		uuid,
		timestamp,
		time.Date(2026, time.August, 14, 12, 34, 56, 789000000, time.UTC),
		date,
		pgtype.Numeric{Int: big.NewInt(12345), Exp: -2, Valid: true},
		pgtype.Time{Microseconds: 12*60*60*1_000_000 + 34*60*1_000_000 + 56*1_000_000 + 789000, Valid: true},
	}

	record, err := plan.recordForValues(values, make([][]byte, len(values)))
	if err != nil {
		t.Fatalf("recordForValues() error = %v", err)
	}
	if got, want := record["event_id"], "5485b58f-4011-4bd2-8a9e-f7801e127a91"; got != want {
		t.Fatalf("event_id = %#v, want %q", got, want)
	}
	if got, want := record["occurred_at"], "2026-08-14T12:34:56.789+05:30"; got != want {
		t.Fatalf("occurred_at = %#v, want %q", got, want)
	}
	if got, want := record["recorded_at"], "2026-08-14T12:34:56.789"; got != want {
		t.Fatalf("recorded_at = %#v, want %q", got, want)
	}
	if got, want := record["event_date"], "2026-08-14"; got != want {
		t.Fatalf("event_date = %#v, want %q", got, want)
	}
	if got, want := record["amount"], "123.45"; got != want {
		t.Fatalf("amount = %#v, want %q", got, want)
	}
	if got, want := record["opened_at"], "12:34:56.789000"; got != want {
		t.Fatalf("opened_at = %#v, want %q", got, want)
	}
	position, err := plan.orderValues(values, make([][]byte, len(values)))
	if err != nil {
		t.Fatalf("orderValues() error = %v", err)
	}
	if got, want := position, []any{pgtype.UUID{Bytes: uuid, Valid: true}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("orderValues() = %#v, want PostgreSQL-encodable key %#v", got, want)
	}
}

func TestPostgresSnapshotStableKeyPaginationValuesAreTypedOrRefused(t *testing.T) {
	relation := database.RelationRef{
		Schema: database.SchemaRef{Catalog: database.CatalogRef{Name: "analytics"}, Name: "public"},
		Name:   "events",
	}
	uuid := [16]byte{0x54, 0x85, 0xb5, 0x8f, 0x40, 0x11, 0x4b, 0xd2, 0x8a, 0x9e, 0xf7, 0x80, 0x1e, 0x12, 0x7a, 0x91}

	for _, test := range []struct {
		name    string
		column  database.Column
		value   any
		raw     []byte
		want    any
		wantErr string
	}{
		{
			name:   "uuid",
			column: database.Column{Ref: database.ColumnRef{Relation: relation, Name: "event_id"}, Type: database.NewUUID(), Nullable: false, Ordinal: 1},
			value:  uuid,
			want:   pgtype.UUID{Bytes: uuid, Valid: true},
		},
		{
			name:   "negative timestamp infinity",
			column: database.Column{Ref: database.ColumnRef{Relation: relation, Name: "occurred_at"}, Type: mustTransportTimestamp(t, false), Nullable: false, Ordinal: 1},
			value:  pgtype.NegativeInfinity,
			want:   pgtype.Timestamp{InfinityModifier: pgtype.NegativeInfinity, Valid: true},
		},
		{
			name:    "json",
			column:  database.Column{Ref: database.ColumnRef{Relation: relation, Name: "body"}, Type: database.NewJSON(), Nullable: false, Ordinal: 1},
			value:   map[string]any{"id": int64(1)},
			raw:     []byte(`{"id":9007199254740993}`),
			wantErr: "json",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			plan := postgresSnapshotReadPlan{
				columns: []database.Column{test.column},
				order:   []database.ColumnRef{test.column.Ref},
			}
			got, err := plan.orderValues([]any{test.value}, [][]byte{test.raw})
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("orderValues() error = %v, want named %q refusal", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("orderValues() error = %v", err)
			}
			if want := []any{test.want}; !reflect.DeepEqual(got, want) {
				t.Fatalf("orderValues() = %#v, want %#v", got, want)
			}
		})
	}
}

func TestPostgresSnapshotRecordValuesPreserveRawJSON(t *testing.T) {
	relation := database.RelationRef{
		Schema: database.SchemaRef{Catalog: database.CatalogRef{Name: "analytics"}, Name: "public"},
		Name:   "events",
	}
	payloadRef := database.ColumnRef{Relation: relation, Name: "payload"}
	plan := postgresSnapshotReadPlan{columns: []database.Column{
		{Ref: payloadRef, Type: database.NewJSON()},
		{Ref: database.ColumnRef{Relation: relation, Name: "null_payload"}, Type: database.NewJSON()},
		{Ref: database.ColumnRef{Relation: relation, Name: "empty_payload"}, Type: database.NewJSON()},
	}, order: []database.ColumnRef{payloadRef}}
	raw := []byte(`{"id":9007199254740993}`)
	record, err := plan.recordForValues(
		[]any{map[string]any{"id": float64(9007199254740992)}, nil, nil},
		[][]byte{raw, []byte("null"), nil},
	)
	if err != nil {
		t.Fatalf("recordForValues() error = %v", err)
	}
	payload, ok := record["payload"].(json.RawMessage)
	if !ok {
		t.Fatalf("payload type = %T, want json.RawMessage", record["payload"])
	}
	if got, want := string(payload), `{"id":9007199254740993}`; got != want {
		t.Fatalf("payload = %q, want %q", got, want)
	}
	nullPayload, ok := record["null_payload"].(json.RawMessage)
	if !ok || string(nullPayload) != "null" {
		t.Fatalf("null_payload = %#v, want JSON null", record["null_payload"])
	}
	if record["empty_payload"] != nil {
		t.Fatalf("empty_payload = %#v, want SQL NULL", record["empty_payload"])
	}
	_, err = plan.orderValues(
		[]any{map[string]any{"id": float64(9007199254740992)}, nil, nil},
		[][]byte{raw, []byte("null"), nil},
	)
	if err == nil || !strings.Contains(err.Error(), "json") {
		t.Fatalf("orderValues() error = %v, want named JSON stable-key refusal", err)
	}
}

func TestPostgresSnapshotOrderValuesUseEncodableTemporalInfinities(t *testing.T) {
	relation := database.RelationRef{
		Schema: database.SchemaRef{Catalog: database.CatalogRef{Name: "analytics"}, Name: "public"},
		Name:   "events",
	}
	tests := []struct {
		name    string
		logical database.LogicalType
		value   pgtype.InfinityModifier
		want    any
	}{
		{
			name:    "date",
			logical: database.NewDate(),
			value:   pgtype.NegativeInfinity,
			want:    pgtype.Date{InfinityModifier: pgtype.NegativeInfinity, Valid: true},
		},
		{
			name:    "timestamp",
			logical: mustTransportTimestamp(t, false),
			value:   pgtype.Infinity,
			want:    pgtype.Timestamp{InfinityModifier: pgtype.Infinity, Valid: true},
		},
		{
			name:    "timestamptz",
			logical: mustTransportTimestamp(t, true),
			value:   pgtype.NegativeInfinity,
			want:    pgtype.Timestamptz{InfinityModifier: pgtype.NegativeInfinity, Valid: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			column := database.Column{Ref: database.ColumnRef{Relation: relation, Name: "stable_key"}, Type: test.logical}
			plan := postgresSnapshotReadPlan{columns: []database.Column{column}, order: []database.ColumnRef{column.Ref}}
			values, err := plan.orderValues([]any{test.value}, make([][]byte, 1))
			if err != nil {
				t.Fatalf("orderValues() error = %v", err)
			}
			if got, want := values, []any{test.want}; !reflect.DeepEqual(got, want) {
				t.Fatalf("orderValues() = %#v, want %#v", got, want)
			}
		})
	}
}

func mustTransportDecimal(t *testing.T) database.LogicalType {
	t.Helper()
	logical, err := database.NewDecimal(12, 2)
	if err != nil {
		t.Fatalf("NewDecimal() error = %v", err)
	}
	return logical
}

func mustTransportTime(t *testing.T) database.LogicalType {
	t.Helper()
	logical, err := database.NewTime(6, false)
	if err != nil {
		t.Fatalf("NewTime() error = %v", err)
	}
	return logical
}

func mustTransportTimestamp(t *testing.T, withTimezone bool) database.LogicalType {
	t.Helper()
	logical, err := database.NewTimestamp(3, withTimezone)
	if err != nil {
		t.Fatalf("NewTimestamp() error = %v", err)
	}
	return logical
}

func TestPostgresSnapshotReadPlanRefusesUnstableOrMissingRelation(t *testing.T) {
	logical, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatalf("NewSignedInteger() error = %v", err)
	}
	relationRef := database.RelationRef{
		Schema: database.SchemaRef{Catalog: database.CatalogRef{Name: "analytics"}, Name: "public"},
		Name:   "events",
	}
	id := database.ColumnRef{Relation: relationRef, Name: "id"}
	catalog, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, []database.Relation{{
		Ref:            relationRef,
		NativeIdentity: database.NativeRelationIdentity{Kind: "oid", Value: "10002"},
		Columns:        []database.Column{{Ref: id, Type: logical, Nullable: false, Ordinal: 1}},
	}})
	if err != nil {
		t.Fatalf("NewCatalog() error = %v", err)
	}
	if _, err := newPostgresSnapshotReadPlan(catalog, relationRef, 1); err == nil || !strings.Contains(err.Error(), "non-null primary or unique key") {
		t.Fatalf("newPostgresSnapshotReadPlan() error = %v, want stable-key refusal", err)
	}
	missing := relationRef
	missing.Name = "missing"
	if _, err := newPostgresSnapshotReadPlan(catalog, missing, 1); err == nil || !strings.Contains(err.Error(), "absent from the typed catalog") {
		t.Fatalf("newPostgresSnapshotReadPlan() missing relation error = %v", err)
	}
}

func TestPostgresSnapshotReadPlanRefusesNonEncodableStableKeyBeforeQuery(t *testing.T) {
	relation := database.RelationRef{
		Schema: database.SchemaRef{Catalog: database.CatalogRef{Name: "analytics"}, Name: "public"},
		Name:   "events",
	}
	key := database.ColumnRef{Relation: relation, Name: "body"}
	catalog, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, []database.Relation{{
		Ref:            relation,
		NativeIdentity: database.NativeRelationIdentity{Kind: "oid", Value: "10004"},
		Columns:        []database.Column{{Ref: key, Type: database.NewJSON(), Nullable: false, Ordinal: 1}},
		Keys:           []database.Key{{Name: "events_pkey", Kind: database.KeyPrimary, Columns: []database.ColumnRef{key}}},
	}})
	if err != nil {
		t.Fatalf("NewCatalog() error = %v", err)
	}
	if _, err := newPostgresSnapshotReadPlan(catalog, relation, 1); err == nil || !strings.Contains(err.Error(), "json") {
		t.Fatalf("newPostgresSnapshotReadPlan() error = %v, want named JSON stable-key refusal", err)
	}
}

func TestPostgresSnapshotCheckpointDoesNotRequireJSONEncodableKeyValues(t *testing.T) {
	logical, err := database.NewFloat(64)
	if err != nil {
		t.Fatalf("NewFloat() error = %v", err)
	}
	relationRef := database.RelationRef{
		Schema: database.SchemaRef{Catalog: database.CatalogRef{Name: "analytics"}, Name: "public"},
		Name:   "events",
	}
	id := database.ColumnRef{Relation: relationRef, Name: "id"}
	catalog, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, []database.Relation{{
		Ref:            relationRef,
		NativeIdentity: database.NativeRelationIdentity{Kind: "oid", Value: "10003"},
		Columns:        []database.Column{{Ref: id, Type: logical, Nullable: false, Ordinal: 1}},
		Keys:           []database.Key{{Name: "events_pkey", Kind: database.KeyPrimary, Columns: []database.ColumnRef{id}}},
	}})
	if err != nil {
		t.Fatalf("NewCatalog() error = %v", err)
	}
	plan, err := newPostgresSnapshotReadPlan(catalog, relationRef, 1)
	if err != nil {
		t.Fatalf("newPostgresSnapshotReadPlan() error = %v", err)
	}
	position, err := plan.orderValues([]any{math.NaN()}, make([][]byte, 1))
	if err != nil || len(position) != 1 || !math.IsNaN(position[0].(float64)) {
		t.Fatalf("orderValues() = %#v, %v; want valid NaN PostgreSQL key", position, err)
	}
	resume := synccontract.ResumeExpectation{
		Source:           synccontract.SourceIdentity{Engine: "postgres", AccountOrCluster: "cluster-alpha", ObjectScope: "analytics.public.events"},
		SourceGeneration: synccontract.OpaqueToken("generation-1"),
	}
	if _, err := postgresSnapshotCheckpoint(resume, "742:742:", catalog.Fingerprint(), 0); err != nil {
		t.Fatalf("postgresSnapshotCheckpoint() rejected a valid opaque PostgreSQL key value: %v", err)
	}
}

func TestPostgresSnapshotRelationRefAllowsTypedCatalogDatabaseName(t *testing.T) {
	relation, err := postgresSnapshotRelationRef("analytics-db.public.events", "analytics-db", "public")
	if err != nil {
		t.Fatalf("postgresSnapshotRelationRef() error = %v", err)
	}
	if relation.Schema.Catalog.Name != "analytics-db" || relation.Schema.Name != "public" || relation.Name != "events" {
		t.Fatalf("postgresSnapshotRelationRef() = %#v, want preserved catalog/schema/relation", relation)
	}
}

func newTransportTestString(t *testing.T, maxLength uint32, collation string) database.LogicalType {
	t.Helper()
	logical, err := database.NewString(maxLength, collation)
	if err != nil {
		t.Fatalf("NewString() error = %v", err)
	}
	return logical
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
		Executor:        connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_polling_watermark"},
		EligibleStreams: []string{"snapshot"},
		Modes:           []synccontract.Mode{synccontract.ModeFullAppend, synccontract.ModeFullOverwrite},
		Delivery: connectors.DeliveryGuarantees{
			Idempotency: connectors.DeliveryIdempotencyAtLeastOnce,
			Ordering:    connectors.DeliveryOrderingSource,
			Deletes:     connectors.DeliveryDeletesUnavailable,
		},
		Conformance: connectors.ConformanceEvidenceReference{Suite: "postgres_polling_watermark", RunID: "shared_source_v1"},
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

func (*preflightDestinationConnector) Check(context.Context, connectors.RuntimeConfig) error {
	return nil
}

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
