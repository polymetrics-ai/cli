package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
)

func TestPostgresColumnTypeRetainsSupportedNativeDetails(t *testing.T) {
	definition := New().databaseDefinition
	for _, tc := range []struct {
		name          string
		column        postgresCatalogColumn
		wantNative    string
		wantModifiers []string
		wantKind      database.LogicalKind
		wantBits      uint8
		wantPrecision uint16
		wantScale     uint16
		wantTimezone  bool
	}{
		{
			name:       "defined signed integer",
			column:     postgresCatalogColumn{nativeName: "int8", typeKind: "b", typeModifier: -1},
			wantNative: "int8", wantKind: database.LogicalSignedInteger, wantBits: 64,
		},
		{
			name:          "bounded numeric",
			column:        postgresCatalogColumn{nativeName: "numeric", typeKind: "b", typeModifier: 4 + (12 << 16) + 2},
			wantNative:    "numeric",
			wantModifiers: []string{"precision-12", "scale-2"},
			wantKind:      database.LogicalDecimal,
			wantPrecision: 12,
			wantScale:     2,
		},
		{
			name:          "bounded varchar",
			column:        postgresCatalogColumn{nativeName: "varchar", typeKind: "b", typeModifier: 4 + 42, collation: "default"},
			wantNative:    "varchar",
			wantModifiers: []string{"length-42"},
			wantKind:      database.LogicalString,
		},
		{
			name:          "timestamp with zone",
			column:        postgresCatalogColumn{nativeName: "timestamptz", typeKind: "b", typeModifier: 3},
			wantNative:    "timestamptz",
			wantModifiers: []string{"precision-3"},
			wantKind:      database.LogicalTimestamp,
			wantPrecision: 3,
			wantTimezone:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			native, logical, err := postgresColumnType(definition, tc.column)
			if err != nil {
				t.Fatal("supported PostgreSQL type was rejected")
			}
			if native.Name != tc.wantNative || strings.Join(native.Modifiers, ",") != strings.Join(tc.wantModifiers, ",") {
				t.Fatal("PostgreSQL native type details were not retained")
			}
			if logical.Kind() != tc.wantKind || logical.BitWidth() != tc.wantBits || logical.Precision() != tc.wantPrecision || logical.Scale() != tc.wantScale || logical.WithTimezone() != tc.wantTimezone {
				t.Fatal("PostgreSQL logical type mapping was incorrect")
			}
		})
	}
}

func TestPostgresColumnTypeRejectsUnsupportedShapes(t *testing.T) {
	definition := New().databaseDefinition
	for _, column := range []postgresCatalogColumn{
		{nativeName: "mood", typeKind: "e", typeModifier: -1},
		{nativeName: "_int4", typeKind: "b", elementOID: 23, typeModifier: -1},
		{nativeName: "numeric", typeKind: "b", typeModifier: -1},
		{nativeName: "timestamptz", typeKind: "b", typeModifier: 10},
	} {
		if _, _, err := postgresColumnType(definition, column); !errors.Is(err, ErrUnsupportedCatalogShape) {
			t.Fatal("unsupported PostgreSQL catalog shape was not rejected explicitly")
		}
	}
}

func TestTypedCatalogRejectsUnsafeConfiguredSchemaBeforeConnect(t *testing.T) {
	_, err := New().TypedCatalog(context.Background(), connectors.RuntimeConfig{
		Config: map[string]string{
			"host":     "db.internal",
			"database": "analytics",
			"username": "reader",
			"schema":   "quoted schema",
			"sslmode":  "disable",
		},
		Secrets: map[string]string{"password": t.Name()},
	})
	if !errors.Is(err, ErrUnsupportedCatalogShape) {
		t.Fatal("unsafe PostgreSQL schema was not rejected before a connection attempt")
	}
}

func TestTypedCatalogHonorsCancelledContextBeforeConfiguration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := New().TypedCatalog(ctx, connectors.RuntimeConfig{}); !errors.Is(err, context.Canceled) {
		t.Fatal("typed PostgreSQL catalog ignored a cancelled context")
	}
}

func TestLegacyStreamsFromTypedCatalogIsProjectionOnly(t *testing.T) {
	integer, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatal("could not create typed test integer")
	}
	text, err := database.NewString(0, "")
	if err != nil {
		t.Fatal("could not create typed test text")
	}
	relationRef := database.RelationRef{
		Schema: database.SchemaRef{Catalog: database.CatalogRef{Name: "analytics"}, Name: "reporting"},
		Name:   "events",
	}
	typed, err := database.NewCatalog(database.CatalogRef{Name: "analytics"}, []database.Relation{{
		Ref:            relationRef,
		NativeIdentity: database.NativeRelationIdentity{Kind: "oid", Value: "42"},
		Columns: []database.Column{
			{Ref: database.ColumnRef{Relation: relationRef, Name: "event_id"}, Type: integer, Ordinal: 1, Native: &database.NativeType{Name: "int8"}},
			{Ref: database.ColumnRef{Relation: relationRef, Name: "body"}, Type: text, Nullable: true, Ordinal: 2, Native: &database.NativeType{Name: "text"}},
		},
		Keys: []database.Key{{
			Name: "events_pkey", Kind: database.KeyPrimary,
			Columns: []database.ColumnRef{{Relation: relationRef, Name: "event_id"}},
		}},
	}})
	if err != nil {
		t.Fatal("could not create typed test catalog")
	}

	streams := legacyStreamsFromTypedCatalog(typed)
	if len(streams) != 1 || streams[0].Name != "reporting.events" || strings.Join(streams[0].PrimaryKey, ",") != "event_id" || len(streams[0].Fields) != 2 || streams[0].Fields[0].Type != "integer" || streams[0].Fields[1].Type != "string" {
		t.Fatal("legacy PostgreSQL catalog was not derived from the typed catalog")
	}
}
