package postgres

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/synccontract"
)

func TestPostgresPollingPlanBindsCompleteCompositeKeyAndStrictResume(t *testing.T) {
	catalog, relation := postgresPollingTestCatalog(t, false)
	plan, err := newPostgresPollingReadPlan(catalog, relation, "sequence", 100)
	if err != nil {
		t.Fatalf("newPostgresPollingReadPlan() error = %v", err)
	}
	resume := synccontract.ResumeExpectation{
		Source:           synccontract.SourceIdentity{Engine: "postgres", AccountOrCluster: "cred_cluster", ObjectScope: "analytics.public.events"},
		SourceGeneration: synccontract.OpaqueToken("generation-1"),
	}
	template := New().Definition().PollingWatermark
	bound, err := bindPostgresPollingDeclaration(template, resume, plan)
	if err != nil {
		t.Fatalf("bindPostgresPollingDeclaration() error = %v", err)
	}
	if got, want := bound.Source.Executor, postgresPollingSourceReference; got != want {
		t.Fatalf("bound source reference = %+v, want %+v", got, want)
	}
	if got, want := bound.Source.Ordering.Watermark.CatalogField, "sequence"; got != want {
		t.Fatalf("bound cursor = %q, want %q", got, want)
	}
	if got, want := bound.Source.Ordering.TieBreaker.CatalogColumns(), []string{"tenant_id", "event_id"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("bound descriptor tie tuple = %#v, want %#v", got, want)
	}
	if got, want := bound.Target.StableKeyMapping, []string{"tenant_id", "event_id"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("bound stable key = %#v, want %#v", got, want)
	}
	if got, want := plan.query(true, 3), `SELECT "tenant_id", "event_id", "sequence", "payload" FROM "public"."events" WHERE ("sequence", "tenant_id", "event_id") > ($1, $2, $3) ORDER BY "sequence" ASC, "tenant_id" ASC, "event_id" ASC LIMIT $4`; got != want {
		t.Fatalf("polling keyset SQL = %q, want %q", got, want)
	}

	first := postgresPollingTestPosition(t, plan, 41, 7, 9)
	second := postgresPollingTestPosition(t, plan, 41, 7, 10)
	third := postgresPollingTestPosition(t, plan, 42, 1, 1)
	arguments, err := plan.afterArguments(&second)
	if err != nil {
		t.Fatalf("afterArguments() error = %v", err)
	}
	if got, want := arguments, []any{int64(41), int64(7), int64(10)}; !reflect.DeepEqual(got, want) {
		t.Fatalf("resume arguments = %#v, want complete tuple %#v", got, want)
	}
	runner := &postgresPollingSourceRunner{plan: plan}
	page := engine.PollingSourcePage{Items: []engine.PollingSourceItem{{Record: connectors.Record{"id": 10}, Position: second}, {Record: connectors.Record{"id": 11}, Position: third}}}
	if err := runner.ValidatePollingSourcePageTraversal(context.Background(), &first, page); err != nil {
		t.Fatalf("strict traversal rejected an advancing duplicate-cursor/composite-key page: %v", err)
	}
	for name, invalid := range map[string]engine.PollingSourcePage{
		"duplicate":    {Items: []engine.PollingSourceItem{{Record: connectors.Record{"id": 9}, Position: first}}},
		"out_of_order": {Items: []engine.PollingSourceItem{{Record: connectors.Record{"id": 8}, Position: postgresPollingTestPosition(t, plan, 40, 9, 9)}}},
	} {
		t.Run(name, func(t *testing.T) {
			err := runner.ValidatePollingSourcePageTraversal(context.Background(), &first, invalid)
			var refusal *engine.PollingWatermarkNonAdvancingError
			if !errors.As(err, &refusal) || refusal.Reason != engine.PollingWatermarkNonAdvancingReasonCursor {
				t.Fatalf("ValidatePollingSourcePageTraversal() error = %T %v, want typed pre-delivery order refusal", err, err)
			}
		})
	}
}

func TestPostgresPollingPlanRefusesNullableCursorAndUnstableShapesBeforeRead(t *testing.T) {
	catalog, relation := postgresPollingTestCatalog(t, true)
	if _, err := newPostgresPollingReadPlan(catalog, relation, "sequence", 100); err == nil || !strings.Contains(err.Error(), "non-null") {
		t.Fatalf("nullable cursor plan error = %v, want non-null refusal", err)
	}

	catalog, relation = postgresPollingTestCatalogWithoutPrimaryKey(t)
	if _, err := newPostgresPollingReadPlan(catalog, relation, "sequence", 100); err == nil || !strings.Contains(err.Error(), "primary key") {
		t.Fatalf("missing key plan error = %v, want primary-key refusal", err)
	}
}

func TestPostgresPollingInvalidOpaqueTupleRequiresRebootstrap(t *testing.T) {
	catalog, relation := postgresPollingTestCatalog(t, false)
	plan, err := newPostgresPollingReadPlan(catalog, relation, "sequence", 100)
	if err != nil {
		t.Fatal(err)
	}
	_, err = plan.afterArguments(&synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("not-json"), TieBreaker: synccontract.OpaqueToken("not-json")})
	var recovery *synccontract.RebootstrapRequiredError
	if !errorsAs(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeInvalidCheckpoint {
		t.Fatalf("invalid tuple error = %v, want typed invalid-checkpoint rebootstrap", err)
	}
}

func postgresPollingTestCatalog(t *testing.T, nullableCursor bool) (database.Catalog, database.RelationRef) {
	t.Helper()
	integer, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatal(err)
	}
	text, err := database.NewString(0, "")
	if err != nil {
		t.Fatal(err)
	}
	relation := database.RelationRef{Schema: database.SchemaRef{Catalog: database.CatalogRef{Name: "analytics"}, Name: "public"}, Name: "events"}
	tenant := database.ColumnRef{Relation: relation, Name: "tenant_id"}
	event := database.ColumnRef{Relation: relation, Name: "event_id"}
	sequence := database.ColumnRef{Relation: relation, Name: "sequence"}
	payload := database.ColumnRef{Relation: relation, Name: "payload"}
	catalog, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, []database.Relation{{
		Ref: relation, NativeIdentity: database.NativeRelationIdentity{Kind: "oid", Value: "42001"},
		Columns: []database.Column{
			{Ref: tenant, Type: integer, Ordinal: 1},
			{Ref: event, Type: integer, Ordinal: 2},
			{Ref: sequence, Type: integer, Nullable: nullableCursor, Ordinal: 3},
			{Ref: payload, Type: text, Ordinal: 4},
		},
		Keys: []database.Key{{Name: "events_pkey", Kind: database.KeyPrimary, Columns: []database.ColumnRef{tenant, event}}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	return catalog, relation
}

func postgresPollingTestCatalogWithoutPrimaryKey(t *testing.T) (database.Catalog, database.RelationRef) {
	t.Helper()
	integer, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatal(err)
	}
	relation := database.RelationRef{Schema: database.SchemaRef{Catalog: database.CatalogRef{Name: "analytics"}, Name: "public"}, Name: "events"}
	sequence := database.ColumnRef{Relation: relation, Name: "sequence"}
	catalog, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, []database.Relation{{
		Ref: relation, NativeIdentity: database.NativeRelationIdentity{Kind: "oid", Value: "42002"},
		Columns: []database.Column{{Ref: sequence, Type: integer, Ordinal: 1}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	return catalog, relation
}

func postgresPollingTestPosition(t *testing.T, plan postgresPollingReadPlan, cursor, tenant, event int64) synccontract.CheckpointPosition {
	t.Helper()
	primary, err := encodePostgresPollingToken([]database.Column{plan.cursor}, []any{cursor})
	if err != nil {
		t.Fatal(err)
	}
	tie, err := encodePostgresPollingToken(plan.tieColumns(), []any{tenant, event})
	if err != nil {
		t.Fatal(err)
	}
	return synccontract.CheckpointPosition{Primary: primary, TieBreaker: tie}
}

// errorsAs keeps the test focused on the contract while avoiding a second
// assertion helper dependency.
func errorsAs(err error, target any) bool {
	return errors.As(err, target)
}
