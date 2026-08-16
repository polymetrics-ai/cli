package postgres

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/arrow"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

func TestPostgresArrowRangePlanProjectsOnlyClosedTransformAndTraversalFields(t *testing.T) {
	transform, err := database.ParseTransformPlanV1([]byte(`{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"},{"expr":{"upper":"status"},"target":"status","type":"string"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	integer, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatal(err)
	}
	text, err := database.NewString(0, "")
	if err != nil {
		t.Fatal(err)
	}
	base := postgresPollingReadPlan{
		cursor: "sequence", tieBreaker: "id", pageSize: 128,
		columns: []database.Column{
			{Ref: database.ColumnRef{Name: "ignored"}, Type: text, Ordinal: 1},
			{Ref: database.ColumnRef{Name: "id"}, Type: integer, Ordinal: 2},
			{Ref: database.ColumnRef{Name: "sequence"}, Type: integer, Ordinal: 3},
			{Ref: database.ColumnRef{Name: "status"}, Type: text, Ordinal: 4},
		},
	}
	plan, err := newPostgresArrowRangePlan(base, transform)
	if err != nil {
		t.Fatalf("newPostgresArrowRangePlan() = %v", err)
	}
	if got, want := len(plan.polling.columns), 3; got != want {
		t.Fatalf("range-query columns = %d, want transform fields plus traversal tuple", got)
	}
	if plan.schema.NumFields() != 2 || plan.schema.Field(0).Name != "id" || plan.schema.Field(0).Type.ID() != arrow.INT64 || plan.schema.Field(1).Name != "status" || plan.schema.Field(1).Type.ID() != arrow.STRING {
		t.Fatalf("projected Arrow schema = %s, want typed transform fields without traversal-only sequence", plan.schema)
	}
	if got, want := plan.recordQueryIndexes, []int{0, 2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Arrow record query indexes = %v, want transform fields only", got)
	}
}

func TestPostgresArrowRangeCheckpointPreservesPreflightSourceIdentity(t *testing.T) {
	identity := synccontract.SourceIdentity{Engine: postgresSnapshotSourceEngine, AccountOrCluster: "cluster", ObjectScope: "public.events"}
	state := engine.PollingSourceRuntimeState{
		SourceGeneration: synccontract.OpaqueToken("generation"), SchemaVersion: "schema",
		SnapshotBarrier: synccontract.SnapshotBarrier{Kind: "none", Token: synccontract.OpaqueToken("postgres-polling-v1")},
		Partitions:      []synccontract.PartitionState{},
		Dedupe:          synccontract.DedupeIdentity{Kind: "tuple", Value: synccontract.OpaqueToken("dedupe")},
		DedupeWindow:    synccontract.DedupeWindow{Kind: "overlap", Start: synccontract.OpaqueToken("start"), End: synccontract.OpaqueToken("end")},
	}
	candidate, err := postgresArrowRangeCheckpoint(identity, state, nil, synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("7"), TieBreaker: synccontract.OpaqueToken("9")}, 1, time.Unix(7, 0).UTC())
	if err != nil {
		t.Fatalf("postgresArrowRangeCheckpoint() = %v", err)
	}
	if candidate.Source != identity || candidate.Mechanism != engine.PollingSourceCheckpointMechanism || string(candidate.Position.Primary) != "7" || string(candidate.Position.TieBreaker) != "9" {
		t.Fatalf("candidate = %#v, want exact source identity and resumable polling candidate", candidate)
	}
}

func TestPostgresArrowRangeCheckpointRefusesNonPostgresSourceBeforeQueryIO(t *testing.T) {
	state := engine.PollingSourceRuntimeState{
		SourceGeneration: synccontract.OpaqueToken("generation"), SchemaVersion: "schema",
		SnapshotBarrier: synccontract.SnapshotBarrier{Kind: "none", Token: synccontract.OpaqueToken("postgres-polling-v1")},
		Partitions:      []synccontract.PartitionState{},
		Dedupe:          synccontract.DedupeIdentity{Kind: "tuple", Value: synccontract.OpaqueToken("dedupe")},
		DedupeWindow:    synccontract.DedupeWindow{Kind: "overlap", Start: synccontract.OpaqueToken("start"), End: synccontract.OpaqueToken("end")},
	}
	_, err := postgresArrowRangeCheckpoint(synccontract.SourceIdentity{Engine: "mysql", AccountOrCluster: "cluster", ObjectScope: "events"}, state, nil, synccontract.CheckpointPosition{}, 0, time.Unix(7, 0).UTC())
	if !errors.Is(err, synctransport.ErrArrowFastPathInvalid) {
		t.Fatalf("postgresArrowRangeCheckpoint(non-postgres source) error = %T %v, want ErrArrowFastPathInvalid before query I/O", err, err)
	}
}
