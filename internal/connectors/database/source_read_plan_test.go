package database_test

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/warehouse"
)

func TestStructuredCatalogIdentityAndReadPlanAreStable(t *testing.T) {
	definition := loadTestDefinition(t, validDefinitionJSON)
	identity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "postgres",
		ConnectionID: "connection-1",
	}
	relation := database.RelationRef{
		Schema: database.SchemaRef{
			Catalog: database.CatalogRef{Name: "analytics"},
			Name:    "public",
		},
		Name: "widgets",
	}
	idColumn := database.ColumnRef{Relation: relation, Name: "id"}
	updatedAtColumn := database.ColumnRef{Relation: relation, Name: "updated_at"}
	int64Type, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatal(err)
	}
	timestamp, err := database.NewTimestamp(6, true)
	if err != nil {
		t.Fatal(err)
	}
	catalogInput := []database.Relation{{
		Ref: relation,
		Columns: []database.Column{
			{Ref: idColumn, Type: int64Type, Ordinal: 1},
			{Ref: updatedAtColumn, Type: timestamp, Ordinal: 2},
		},
		Keys: []database.Key{{
			Name:    "widgets_pk",
			Kind:    database.KeyPrimary,
			Columns: []database.ColumnRef{idColumn},
		}},
	}}
	catalog, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, catalogInput)
	if err != nil {
		t.Fatalf("NewCatalog() error = %v", err)
	}
	catalogInput[0].Columns[0].Ref.Name = "input_mutated"
	if got := catalog.Relations()[0].Columns[0].Ref.Name; got == "input_mutated" {
		t.Fatal("NewCatalog() retained caller-owned mutable relation state")
	}

	source, err := database.NewSourceRef(identity, relation)
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := warehouse.NewArtifactRef(identity, "widgets")
	if err != nil {
		t.Fatal(err)
	}
	inbound, err := database.NewWarehouseInboundRef(source, artifact)
	if err != nil {
		t.Fatal(err)
	}
	target, err := database.NewTargetRef(identity, relation)
	if err != nil {
		t.Fatal(err)
	}
	if got := source.Identity(); got != identity {
		t.Fatalf("source identity = %#v, want %#v", got, identity)
	}
	if got := target.Identity(); got != identity {
		t.Fatalf("target identity = %#v, want %#v", got, identity)
	}
	if _, err := database.NewTargetRef(database.ConnectionIdentity{
		WorkspaceID: "workspace-1",
		ConnectorID: "postgres",
	}, relation); err == nil {
		t.Fatal("NewTargetRef() accepted an owner without a connection ID")
	}

	plan, err := database.NewReadPlan(context.Background(), database.ReadPlanRequest{
		Inbound:    inbound,
		Definition: definition,
		Catalog:    catalog,
		Relation:   relation,
		Columns:    []database.ColumnRef{idColumn, updatedAtColumn},
		Order: []database.OrderTerm{
			{Column: updatedAtColumn, Direction: database.SortAscending},
			{Column: idColumn, Direction: database.SortAscending},
		},
		PageSize: 0,
	})
	if err != nil {
		t.Fatalf("NewReadPlan() error = %v", err)
	}
	if plan.SchemaFingerprint().IsZero() {
		t.Fatal("read plan has no schema fingerprint")
	}
	if got := plan.PageSize(); got != definition.Resources().ReadPage.Default {
		t.Fatalf("read plan page size = %d, want definition default %d", got, definition.Resources().ReadPage.Default)
	}
	if got := plan.Warehouse(); got.Identity() != identity || got.Table() != "widgets" {
		t.Fatalf("read plan warehouse = %#v/%q, want source-owned widgets artifact", got.Identity(), got.Table())
	}
	relations := catalog.Relations()
	relations[0].Columns[0].Ref.Name = "mutated"
	if got := catalog.Relations()[0].Columns[0].Ref.Name; got == "mutated" {
		t.Fatal("Catalog.Relations() returned mutable internal state")
	}
	columns := plan.Columns()
	columns[0].Name = "mutated"
	if got := plan.Columns()[0].Name; got == "mutated" {
		t.Fatal("ReadPlan.Columns() returned mutable internal state")
	}
	order := plan.Order()
	order[0].Direction = database.SortDescending
	if got := plan.Order()[0].Direction; got == database.SortDescending {
		t.Fatal("ReadPlan.Order() returned mutable internal state")
	}

	permutedColumns := catalog.Relations()[0].Columns
	permutedColumns[0], permutedColumns[1] = permutedColumns[1], permutedColumns[0]
	permuted, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, []database.Relation{{
		Ref:     relation,
		Columns: permutedColumns,
		Keys: []database.Key{{
			Name:    "widgets_pk",
			Kind:    database.KeyPrimary,
			Columns: []database.ColumnRef{idColumn},
		}},
	}})
	if err != nil {
		t.Fatalf("NewCatalog(permuted columns) error = %v", err)
	}
	if got, want := permuted.Fingerprint(), catalog.Fingerprint(); got != want {
		t.Fatalf("schema fingerprint changed with column input order: got %s, want %s", got, want)
	}

	_, err = database.NewReadPlan(context.Background(), database.ReadPlanRequest{
		Inbound:    inbound,
		Definition: definition,
		Catalog:    catalog,
		Relation:   relation,
		Columns:    []database.ColumnRef{idColumn, updatedAtColumn},
		Order:      []database.OrderTerm{{Column: updatedAtColumn, Direction: database.SortAscending}},
		PageSize:   100,
	})
	if err == nil {
		t.Fatal("NewReadPlan() accepted an order without the unique-key suffix")
	}

	_, err = database.NewReadPlan(context.Background(), database.ReadPlanRequest{
		Inbound:    inbound,
		Definition: definition,
		Catalog:    catalog,
		Relation:   relation,
		Columns:    []database.ColumnRef{idColumn, updatedAtColumn},
		Order: []database.OrderTerm{
			{Column: updatedAtColumn, Direction: database.SortAscending},
			{Column: idColumn, Direction: database.SortAscending},
		},
		PageSize: definition.Resources().ReadPage.Maximum + 1,
	})
	if err == nil {
		t.Fatal("NewReadPlan() accepted a page size above the definition maximum")
	}

	emailColumn := database.ColumnRef{Relation: relation, Name: "email"}
	textType, err := database.NewString(0, "")
	if err != nil {
		t.Fatal(err)
	}
	nullableUniqueCatalog, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, []database.Relation{{
		Ref: relation,
		Columns: []database.Column{
			{Ref: idColumn, Type: int64Type, Ordinal: 1},
			{Ref: emailColumn, Type: textType, Nullable: true, Ordinal: 2},
		},
		Keys: []database.Key{{
			Name:    "widgets_email_key",
			Kind:    database.KeyUnique,
			Columns: []database.ColumnRef{emailColumn},
		}},
	}})
	if err != nil {
		t.Fatalf("NewCatalog(nullable unique key) error = %v", err)
	}
	if _, err := database.NewReadPlan(context.Background(), database.ReadPlanRequest{
		Inbound:    inbound,
		Definition: definition,
		Catalog:    nullableUniqueCatalog,
		Relation:   relation,
		Columns:    []database.ColumnRef{emailColumn},
		Order:      []database.OrderTerm{{Column: emailColumn, Direction: database.SortAscending}},
		PageSize:   100,
	}); err == nil {
		t.Fatal("NewReadPlan() accepted a nullable unique-key suffix")
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := database.NewReadPlan(cancelled, database.ReadPlanRequest{
		Inbound:    inbound,
		Definition: definition,
		Catalog:    catalog,
		Relation:   relation,
		Columns:    []database.ColumnRef{idColumn, updatedAtColumn},
		Order: []database.OrderTerm{
			{Column: updatedAtColumn, Direction: database.SortAscending},
			{Column: idColumn, Direction: database.SortAscending},
		},
		PageSize: 100,
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("NewReadPlan(cancelled) error = %v, want context.Canceled", err)
	}
}

func TestDatabaseLoadAndReadPlanHonorCancellation(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := database.Load(cancelled, fstest.MapFS{"database.json": &fstest.MapFile{Data: []byte(validDefinitionJSON)}}); !errors.Is(err, context.Canceled) {
		t.Fatalf("Load(cancelled) error = %v, want context.Canceled", err)
	}
}

func TestDatabaseLoadChecksCancellationBeforeReturningProjection(t *testing.T) {
	ctx := &cancelOnErrCallContext{cancelAt: 4, done: make(chan struct{})}
	definition, err := database.Load(ctx, fstest.MapFS{"database.json": &fstest.MapFile{Data: []byte(validDefinitionJSON)}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Load() error = %v, want context.Canceled", err)
	}
	if definition.SchemaVersion() != 0 {
		t.Fatalf("Load() definition schema version = %d, want zero projection after cancellation", definition.SchemaVersion())
	}
}

func TestReadPlanChecksCancellationBeforeReturningProjection(t *testing.T) {
	ctx := &cancelOnErrCallContext{cancelAt: 3, done: make(chan struct{})}
	plan, err := database.NewReadPlan(ctx, testReadPlanRequest(t))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("NewReadPlan() error = %v, want context.Canceled", err)
	}
	if plan.PageSize() != 0 {
		t.Fatalf("NewReadPlan() page size = %d, want zero plan after cancellation", plan.PageSize())
	}
}
