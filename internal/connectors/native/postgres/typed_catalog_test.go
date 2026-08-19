package postgres

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
)

const reservedPostgresCatalogSchemaError = "postgres catalog schema is reserved for PostgreSQL system objects"

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

func TestTypedCatalogRejectsReservedConfiguredSchemasBeforeConnect(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("could not reserve a loopback port for the pre-connection scope test")
	}
	port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	if err := listener.Close(); err != nil {
		t.Fatal("could not release the loopback port for the pre-connection scope test")
	}

	for _, schema := range []string{
		"pg_catalog",
		"information_schema",
		"pg_toast",
		"pg_toast_4070",
		"pg_temp_4070",
	} {
		t.Run(schema, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			_, err := New().TypedCatalog(ctx, connectors.RuntimeConfig{
				Config: map[string]string{
					"host":     "127.0.0.1",
					"port":     port,
					"database": "analytics",
					"username": "reader",
					"schema":   schema,
					"sslmode":  "disable",
				},
				Secrets: map[string]string{"password": t.Name()},
			})
			if !errors.Is(err, ErrSystemCatalogSchema) || !strings.Contains(err.Error(), reservedPostgresCatalogSchemaError) {
				t.Fatalf("reserved PostgreSQL schema was not rejected before a connection attempt: %v", err)
			}
		})
	}
}

func TestTypedCatalogHonorsCancelledContextBeforeConfiguration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := New().TypedCatalog(ctx, connectors.RuntimeConfig{}); !errors.Is(err, context.Canceled) {
		t.Fatal("typed PostgreSQL catalog ignored a cancelled context")
	}
}

func TestTypedCatalogResourcePolicyEnforcesDeclaredBounds(t *testing.T) {
	policy := New().databaseDefinition.Resources()
	resources, err := newTypedCatalogResources(policy)
	if err != nil {
		t.Fatalf("newTypedCatalogResources() error = %v", err)
	}
	if resources.poolSize != int32(policy.Pool.Default) {
		t.Fatalf("catalog pool size = %d, want declared default %d", resources.poolSize, policy.Pool.Default)
	}

	conn, err := resolveConfig(connectors.RuntimeConfig{
		Config: map[string]string{
			"host":     "postgres.example",
			"database": "analytics",
			"username": "reader",
			"sslmode":  "disable",
		},
		Secrets: map[string]string{"password": t.Name()},
	})
	if err != nil {
		t.Fatalf("resolveConfig() error = %v", err)
	}
	poolConfig, err := conn.typedCatalogPoolConfig(resources)
	if err != nil {
		t.Fatalf("typedCatalogPoolConfig() error = %v", err)
	}
	if poolConfig.MaxConns != int32(policy.Pool.Default) {
		t.Fatalf("catalog pool max connections = %d, want %d", poolConfig.MaxConns, policy.Pool.Default)
	}
	if poolConfig.ConnConfig.ConnectTimeout != policy.ConnectTimeout {
		t.Fatalf("catalog connect timeout = %s, want %s", poolConfig.ConnConfig.ConnectTimeout, policy.ConnectTimeout)
	}

	before := time.Now()
	operationCtx, cancel, err := resources.operationContext(context.Background())
	if err != nil {
		t.Fatalf("operationContext() error = %v", err)
	}
	defer cancel()
	deadline, ok := operationCtx.Deadline()
	if !ok || deadline.Before(before.Add(policy.OperationTimeout-time.Second)) || deadline.After(before.Add(policy.OperationTimeout+time.Second)) {
		t.Fatal("catalog operation context did not retain the declared bounded timeout")
	}

	atMaximum := policy
	atMaximum.Pool.Default = atMaximum.Pool.Maximum
	maximumResources, err := newTypedCatalogResources(atMaximum)
	if err != nil {
		t.Fatalf("newTypedCatalogResources(maximum pool) error = %v", err)
	}
	maximumConfig, err := conn.typedCatalogPoolConfig(maximumResources)
	if err != nil || maximumConfig.MaxConns != int32(atMaximum.Pool.Maximum) {
		t.Fatal("catalog resource policy did not accept its declared pool boundary")
	}

	overPool := policy
	overPool.Pool.Default = overPool.Pool.Maximum + 1
	if _, err := newTypedCatalogResources(overPool); !errors.Is(err, errCatalogResourcePolicy) {
		t.Fatalf("over-bound catalog pool policy error = %v, want typed refusal", err)
	}
	overOperation := policy
	overOperation.OperationTimeout = time.Hour + time.Nanosecond
	if _, err := newTypedCatalogResources(overOperation); !errors.Is(err, errCatalogResourcePolicy) {
		t.Fatalf("over-bound catalog operation policy error = %v, want typed refusal", err)
	}
}

func TestTypedCatalogQueriesRequireExecutableReadAuthorization(t *testing.T) {
	for _, query := range []string{typedCatalogRelationsSQL, typedCatalogColumnsSQL, typedCatalogKeysSQL} {
		for _, predicate := range []string{
			"pg_catalog.has_schema_privilege(n.oid, 'USAGE')",
			"pg_catalog.has_table_privilege(c.oid, 'SELECT')",
			"OR NOT EXISTS (",
			"pg_catalog.has_column_privilege(c.oid, readable_attribute.attnum, 'SELECT')",
			"readable_attribute.attnum > 0",
			"NOT readable_attribute.attisdropped",
		} {
			if !strings.Contains(query, predicate) {
				t.Fatalf("typed catalog query did not require executable SELECT * authorization: %q", predicate)
			}
		}
	}
}

func TestTypedCatalogSnapshotIsReadOnlyRepeatableRead(t *testing.T) {
	options := typedCatalogTransactionOptions()
	if options.IsoLevel != pgx.RepeatableRead || options.AccessMode != pgx.ReadOnly {
		t.Fatal("typed catalog did not require one read-only repeatable-read snapshot")
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
