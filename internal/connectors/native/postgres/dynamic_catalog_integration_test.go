//go:build databaseintegration

package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/connectors/native/dbtest"
	native "polymetrics.ai/internal/connectors/native/postgres"
)

const (
	postgresCatalogIntegrationImage    = "docker.io/library/postgres:16.10"
	postgresCatalogIntegrationDatabase = "pm_catalog"
	postgresCatalogIntegrationUser     = "pm_catalog"
	// postgresCatalogIntegrationImageBytes is a conservative approximate
	// on-disk footprint for the pinned image. dbtest uses it only to prove
	// image-store headroom before a pull.
	postgresCatalogIntegrationImageBytes = 420 << 20

	postgresCatalogIntegrationEnabledEnv  = "POLYMETRICS_DATABASE_INTEGRATION"
	postgresCatalogIntegrationRuntimeEnv  = "POLYMETRICS_CONTAINER_RUNTIME"
	postgresCatalogIntegrationEndpointEnv = "POLYMETRICS_CONTAINER_ENDPOINT"
	postgresCatalogAlphaSchema            = "catalog_alpha"
	postgresCatalogBetaSchema             = "catalog_beta"
	postgresCatalogUnsupportedSchema      = "catalog_unsupported"
	postgresCatalogEmptySchema            = "catalog_empty"
	postgresCatalogPrivilegesSchema       = "catalog_privileges"
	postgresCatalogLimitedUser            = "pm_catalog_limited"
	postgresCatalogNoUsageSchema          = "catalog_no_usage"
	postgresCatalogNoUsageUser            = "pm_catalog_no_usage"
	postgresCatalogSystemSchemaError      = "postgres catalog schema is reserved for PostgreSQL system objects"
	postgresCatalogReadTable              = "read_events"
	postgresCatalogAlternateReadTable     = "alternate_events"
	postgresCatalogNullableReadTable      = "nullable_cursor_events"
)

var errPostgresCatalogContainerRuntime = errors.New("database integration requires POLYMETRICS_CONTAINER_RUNTIME=docker or podman and POLYMETRICS_CONTAINER_ENDPOINT=unix:///absolute/path/to/socket; no usable explicit local container runtime is configured")

// TestPostgresDynamicTypedCatalogUsesLiveMetadata is deliberately a live
// regression: a hard-coded table/field list cannot be correct for both
// independently created schemas. The information_schema assertions are an
// independent server-side oracle rather than an expected catalog object built
// by the connector under test.
func TestPostgresDynamicTypedCatalogUsesLiveMetadata(t *testing.T) {
	if os.Getenv(postgresCatalogIntegrationEnabledEnv) != "1" {
		t.Skipf("database integration skipped: set %s=1 to run the PostgreSQL Docker or Podman proof", postgresCatalogIntegrationEnabledEnv)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	harness, err := newPostgresCatalogContainerHarness(
		dbtest.Runtime(strings.TrimSpace(os.Getenv(postgresCatalogIntegrationRuntimeEnv))),
		strings.TrimSpace(os.Getenv(postgresCatalogIntegrationEndpointEnv)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Error("PostgreSQL database test cleanup failed")
		}
		report := harness.Report()
		t.Logf("PostgreSQL database test target image-store free bytes: before=%d after=%d", report.DiskFreeBefore, report.DiskFreeAfter)
	}()

	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatal("PostgreSQL database container did not start")
	}
	connector := native.New()
	alphaConfig := postgresCatalogConfig(t, endpoint, postgresCatalogAlphaSchema)
	waitForPostgresCatalog(t, ctx, connector, alphaConfig)

	source := openPostgresCatalogSource(t, ctx, endpoint)
	defer func() { _ = source.Close(ctx) }()
	seedPostgresCatalogs(t, ctx, source)
	assertPostgresLiveReads(t, ctx, connector, endpoint)
	assertPostgresSystemSchemasAreRejected(t, ctx, connector, endpoint)

	alpha, err := connector.TypedCatalog(ctx, alphaConfig)
	if err != nil {
		t.Fatal("PostgreSQL typed catalog discovery failed for the alpha schema")
	}
	betaConfig := postgresCatalogConfig(t, endpoint, postgresCatalogBetaSchema)
	beta, err := connector.TypedCatalog(ctx, betaConfig)
	if err != nil {
		t.Fatal("PostgreSQL typed catalog discovery failed for the beta schema")
	}
	if alpha.Fingerprint() == beta.Fingerprint() {
		t.Fatal("materially different live PostgreSQL schemas produced one catalog fingerprint")
	}

	assertCatalogMatchesInformationSchema(t, ctx, source, alpha, postgresCatalogAlphaSchema)
	assertCatalogMatchesInformationSchema(t, ctx, source, beta, postgresCatalogBetaSchema)
	assertCatalogKeysMatchInformationSchema(t, ctx, source, alpha, postgresCatalogAlphaSchema)
	assertCatalogKeysMatchInformationSchema(t, ctx, source, beta, postgresCatalogBetaSchema)
	assertAlphaTypedCatalog(t, alpha)
	assertBetaTypedCatalog(t, beta)

	legacy, err := connector.Catalog(ctx, alphaConfig)
	if err != nil {
		t.Fatal("PostgreSQL compatibility catalog failed after typed discovery")
	}
	assertLegacyStream(t, legacy, postgresCatalogAlphaSchema+".accounts")

	emptyConfig := postgresCatalogConfig(t, endpoint, postgresCatalogEmptySchema)
	if _, err := connector.TypedCatalog(ctx, emptyConfig); !errors.Is(err, native.ErrUnsupportedCatalogShape) {
		t.Fatal("PostgreSQL typed catalog did not fail closed for an eligible zero-column relation")
	}
	if _, err := connector.Catalog(ctx, emptyConfig); !errors.Is(err, native.ErrUnsupportedCatalogShape) {
		t.Fatal("PostgreSQL compatibility catalog silently omitted an eligible zero-column relation")
	}

	limitedConfig := postgresCatalogLimitedConfig(t, endpoint, postgresCatalogPrivilegesSchema)
	limited, err := connector.TypedCatalog(ctx, limitedConfig)
	if err != nil {
		t.Fatal("PostgreSQL typed catalog did not honor least-privilege discovery")
	}
	if len(limited.Relations()) != 2 {
		t.Fatal("PostgreSQL typed catalog exposed inaccessible relations")
	}
	limitedVisible := catalogRelation(t, limited, postgresCatalogPrivilegesSchema, "visible")
	assertTypedColumn(t, limitedVisible, "id", 1, false, "int4", nil, database.LogicalSignedInteger, 32, 0, 0, false)
	limitedColumnGranted := catalogRelation(t, limited, postgresCatalogPrivilegesSchema, "column_granted")
	assertTypedColumn(t, limitedColumnGranted, "id", 1, false, "int4", nil, database.LogicalSignedInteger, 32, 0, 0, false)
	assertTypedColumn(t, limitedColumnGranted, "label", 2, false, "text", nil, database.LogicalString, 0, 0, 0, false)
	assertCatalogOmitsRelation(t, limited, postgresCatalogPrivilegesSchema, "hidden")
	for _, stream := range []string{"visible", "column_granted"} {
		if err := connector.Read(ctx, connectors.ReadRequest{Stream: stream, Config: limitedConfig}, func(connectors.Record) error { return nil }); err != nil {
			t.Fatalf("PostgreSQL reader could not execute least-privilege SELECT * for %s: %v", stream, err)
		}
	}

	noUsageConfig := postgresCatalogConfig(t, endpoint, postgresCatalogNoUsageSchema)
	noUsageConfig.Config["username"] = postgresCatalogNoUsageUser
	if _, err := connector.TypedCatalog(ctx, noUsageConfig); !errors.Is(err, native.ErrNoSupportedRelations) {
		t.Fatal("PostgreSQL typed catalog exposed a relation without schema USAGE")
	}
	if err := connector.Read(ctx, connectors.ReadRequest{Stream: "blocked", Config: noUsageConfig}, func(connectors.Record) error { return nil }); err == nil {
		t.Fatal("PostgreSQL reader unexpectedly accessed a relation without schema USAGE")
	}

	unsupportedConfig := postgresCatalogConfig(t, endpoint, postgresCatalogUnsupportedSchema)
	if _, err := connector.TypedCatalog(ctx, unsupportedConfig); !errors.Is(err, native.ErrUnsupportedCatalogShape) {
		t.Fatal("PostgreSQL typed catalog did not reject an unsupported native type")
	}
}

func newPostgresCatalogContainerHarness(runtime dbtest.Runtime, endpoint string) (*dbtest.Harness, error) {
	if runtime == "" || endpoint == "" {
		return nil, errPostgresCatalogContainerRuntime
	}
	harness, err := dbtest.New(dbtest.Config{
		Engine:                   "postgres",
		ContainerRuntime:         runtime,
		Image:                    postgresCatalogIntegrationImage,
		ContainerPort:            5432,
		DataVolumePath:           "/var/lib/postgresql/data",
		ContainerEndpoint:        endpoint,
		ExpectedImageBytes:       postgresCatalogIntegrationImageBytes,
		DockerCapacityProbeImage: "docker.io/library/busybox:1.37.0",
		ContainerArgs: []string{
			"--env", "POSTGRES_DB=" + postgresCatalogIntegrationDatabase,
			"--env", "POSTGRES_USER=" + postgresCatalogIntegrationUser,
			"--env", "POSTGRES_HOST_AUTH_METHOD=trust",
		},
	})
	if err != nil {
		return nil, errPostgresCatalogContainerRuntime
	}
	return harness, nil
}

func assertPostgresSystemSchemasAreRejected(t *testing.T, ctx context.Context, connector native.Connector, endpoint dbtest.Endpoint) {
	t.Helper()

	temporarySource := openPostgresCatalogSource(t, ctx, endpoint)
	defer func() { _ = temporarySource.Close(ctx) }()
	if _, err := temporarySource.Exec(ctx, "CREATE TEMPORARY TABLE catalog_scope_temp_probe_4070 (probe_id integer PRIMARY KEY, marker text NOT NULL)"); err != nil {
		t.Fatal("could not create the held PostgreSQL temporary-table scope probe")
	}
	var temporarySchema string
	if err := temporarySource.QueryRow(ctx, "SELECT nspname FROM pg_catalog.pg_namespace WHERE oid = pg_my_temp_schema()").Scan(&temporarySchema); err != nil {
		t.Fatal("could not identify the held PostgreSQL temporary-table schema")
	}
	if !strings.HasPrefix(temporarySchema, "pg_temp_") {
		t.Fatal("PostgreSQL temporary-table probe did not use a physical pg_temp_N schema")
	}

	schemas := []string{
		"pg_catalog",
		"information_schema",
		"pg_toast",
		"pg_toast_4070",
		temporarySchema,
	}
	for _, schema := range schemas {
		config := postgresCatalogConfig(t, endpoint, schema)
		if _, err := connector.TypedCatalog(ctx, config); !errors.Is(err, native.ErrSystemCatalogSchema) || !strings.Contains(err.Error(), postgresCatalogSystemSchemaError) {
			t.Fatal("typed PostgreSQL catalog did not reject a system-owned schema before discovery")
		}
		if _, err := connector.Catalog(ctx, config); !errors.Is(err, native.ErrSystemCatalogSchema) || !strings.Contains(err.Error(), postgresCatalogSystemSchemaError) {
			t.Fatal("legacy PostgreSQL catalog did not preserve the typed system-schema rejection")
		}
	}
}

func postgresCatalogLimitedConfig(t *testing.T, endpoint dbtest.Endpoint, schema string) connectors.RuntimeConfig {
	t.Helper()
	config := postgresCatalogConfig(t, endpoint, schema)
	config.Config["username"] = postgresCatalogLimitedUser
	return config
}

func postgresCatalogConfig(t *testing.T, endpoint dbtest.Endpoint, schema string) connectors.RuntimeConfig {
	t.Helper()
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"host":     endpoint.Host,
			"port":     strconv.Itoa(endpoint.Port),
			"database": postgresCatalogIntegrationDatabase,
			"username": postgresCatalogIntegrationUser,
			"schema":   schema,
			"sslmode":  "disable",
		},
		// PostgreSQL trust authentication ignores this generated non-secret value;
		// it exists only because live connector configuration requires a nonempty
		// password field before it opens a pool.
		Secrets: map[string]string{"password": t.Name()},
	}
}

func waitForPostgresCatalog(t *testing.T, ctx context.Context, connector native.Connector, config connectors.RuntimeConfig) {
	t.Helper()
	for {
		if err := connector.Check(ctx, config); err == nil {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("PostgreSQL engine was not reachable before the integration deadline")
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func openPostgresCatalogSource(t *testing.T, ctx context.Context, endpoint dbtest.Endpoint) *pgx.Conn {
	t.Helper()
	config, err := pgx.ParseConfig("sslmode=disable")
	if err != nil {
		t.Fatal("could not configure the isolated PostgreSQL test source")
	}
	config.Host = endpoint.Host
	config.Port = uint16(endpoint.Port)
	config.Database = postgresCatalogIntegrationDatabase
	config.User = postgresCatalogIntegrationUser
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		t.Fatal("could not open the isolated PostgreSQL test source")
	}
	return conn
}

func seedPostgresCatalogs(t *testing.T, ctx context.Context, source *pgx.Conn) {
	t.Helper()
	statements := []string{
		"CREATE SCHEMA catalog_alpha",
		"CREATE TABLE catalog_alpha.accounts (account_id bigint NOT NULL, total numeric(12,2) NOT NULL, occurred_at timestamptz(3) NOT NULL, body jsonb, PRIMARY KEY (account_id))",
		"CREATE TABLE catalog_alpha.audit (event_id uuid NOT NULL, body jsonb NOT NULL, PRIMARY KEY (event_id))",
		"CREATE TABLE catalog_alpha.read_events (id bigint NOT NULL, sequence bigint NOT NULL, label text NOT NULL, PRIMARY KEY (id))",
		"INSERT INTO catalog_alpha.read_events (id, sequence, label) VALUES (1, 10, 'alpha'), (2, 10, 'bravo'), (3, 11, 'charlie'), (4, 12, 'delta'), (5, 12, 'echo')",
		"CREATE TABLE catalog_alpha.alternate_events (id bigint NOT NULL, alternate_cursor bigint NOT NULL, label text NOT NULL, PRIMARY KEY (id))",
		"INSERT INTO catalog_alpha.alternate_events (id, alternate_cursor, label) VALUES (11, 100, 'other')",
		"CREATE TABLE catalog_alpha.nullable_cursor_events (id bigint NOT NULL, cursor_value bigint, label text NOT NULL, PRIMARY KEY (id))",
		"INSERT INTO catalog_alpha.nullable_cursor_events (id, cursor_value, label) VALUES (21, NULL, 'null'), (22, 1, 'one'), (23, 2, 'two')",
		"CREATE VIEW catalog_alpha.accounts_view AS SELECT account_id FROM catalog_alpha.accounts",
		"CREATE SCHEMA catalog_beta",
		"CREATE TABLE catalog_beta.accounts (tenant_id integer NOT NULL, record_id integer NOT NULL, label varchar(42), occurred_at timestamp(3), PRIMARY KEY (tenant_id, record_id), UNIQUE (label))",
		"CREATE SCHEMA catalog_unsupported",
		"CREATE TYPE catalog_unsupported.mood AS ENUM ('calm', 'storm')",
		"CREATE TABLE catalog_unsupported.events (id integer PRIMARY KEY, mood catalog_unsupported.mood NOT NULL)",
		"CREATE SCHEMA catalog_empty",
		"CREATE TABLE catalog_empty.visible (id integer PRIMARY KEY)",
		"CREATE TABLE catalog_empty.empty ()",
		"CREATE SCHEMA catalog_privileges",
		"CREATE TABLE catalog_privileges.visible (id integer PRIMARY KEY)",
		"CREATE TABLE catalog_privileges.column_granted (id integer PRIMARY KEY, label text NOT NULL)",
		"CREATE TYPE catalog_privileges.hidden_kind AS ENUM ('hidden')",
		"CREATE TABLE catalog_privileges.hidden (id integer PRIMARY KEY, kind catalog_privileges.hidden_kind NOT NULL)",
		"CREATE ROLE pm_catalog_limited LOGIN",
		"GRANT USAGE ON SCHEMA catalog_privileges TO pm_catalog_limited",
		"GRANT SELECT ON TABLE catalog_privileges.visible TO pm_catalog_limited",
		"GRANT SELECT (id, label) ON TABLE catalog_privileges.column_granted TO pm_catalog_limited",
		"CREATE SCHEMA catalog_no_usage",
		"REVOKE ALL ON SCHEMA catalog_no_usage FROM PUBLIC",
		"CREATE TABLE catalog_no_usage.blocked (id integer PRIMARY KEY)",
		"CREATE ROLE pm_catalog_no_usage LOGIN",
		"GRANT SELECT ON TABLE catalog_no_usage.blocked TO pm_catalog_no_usage",
	}
	for _, statement := range statements {
		if _, err := source.Exec(ctx, statement); err != nil {
			t.Fatal("could not seed deterministic PostgreSQL catalog fixtures")
		}
	}
}

// assertPostgresLiveReads proves records observed through Connector.Read,
// rather than considering a container's exit status evidence. It also locks
// today's cursor-field behavior for the captain's separate cursor-contract
// decision without changing that user-facing connection option here.
func assertPostgresLiveReads(t *testing.T, ctx context.Context, connector native.Connector, endpoint dbtest.Endpoint) {
	t.Helper()
	config := postgresCatalogConfig(t, endpoint, postgresCatalogAlphaSchema)
	config.Config["cursor_field"] = "sequence"
	config.Config["read_limit"] = "10"

	catalog, err := connector.Catalog(ctx, config)
	if err != nil {
		t.Fatal("PostgreSQL live read catalog discovery failed")
	}
	assertLiveReadCatalog(t, catalog, postgresCatalogReadTable)

	full := collectPostgresRead(t, ctx, connector, connectors.ReadRequest{
		Stream: postgresCatalogReadTable,
		Config: config,
	})
	assertLiveReadIDs(t, full, []string{"1", "2", "3", "4", "5"})
	t.Logf("live PostgreSQL full read %s: ids=%s labels=%s", postgresCatalogReadTable, liveReadIDs(full), liveReadLabels(full))

	incremental := collectPostgresRead(t, ctx, connector, connectors.ReadRequest{
		Stream: postgresCatalogReadTable,
		Config: config,
		State:  map[string]string{"cursor": "10"},
	})
	assertLiveReadIDs(t, incremental, []string{"3", "4", "5"})
	t.Logf("live PostgreSQL cursor read %s after=10: ids=%s labels=%s", postgresCatalogReadTable, liveReadIDs(incremental), liveReadLabels(incremental))

	noCursorConfig := postgresCatalogConfig(t, endpoint, postgresCatalogAlphaSchema)
	noCursorConfig.Config["read_limit"] = "10"
	withoutCursor := collectPostgresRead(t, ctx, connector, connectors.ReadRequest{
		Stream: postgresCatalogReadTable,
		Config: noCursorConfig,
		State:  map[string]string{"cursor": "12"},
	})
	assertLiveReadIDSet(t, withoutCursor, []string{"1", "2", "3", "4", "5"})
	t.Logf("live PostgreSQL cursor_field absent with stored cursor=12: ids=%s", liveReadIDs(withoutCursor))

	missingCursorConfig := config
	missingCursorConfig.Config = clonePostgresCatalogConfig(config.Config)
	missingCursorConfig.Config["cursor_field"] = "missing_sequence"
	if err := connector.Read(ctx, connectors.ReadRequest{Stream: postgresCatalogReadTable, Config: missingCursorConfig}, func(connectors.Record) error { return nil }); err == nil {
		t.Fatal("PostgreSQL read accepted a nonexistent cursor column")
	}
	t.Log("live PostgreSQL nonexistent cursor column: read rejected")

	nullableConfig := postgresCatalogConfig(t, endpoint, postgresCatalogAlphaSchema)
	nullableConfig.Config["cursor_field"] = "cursor_value"
	nullableConfig.Config["read_limit"] = "10"
	nullable := collectPostgresRead(t, ctx, connector, connectors.ReadRequest{
		Stream: postgresCatalogNullableReadTable,
		Config: nullableConfig,
		State:  map[string]string{"cursor": "1"},
	})
	assertLiveReadIDs(t, nullable, []string{"23"})
	t.Logf("live PostgreSQL nullable cursor rows after=1: ids=%s; null cursor row omitted", liveReadIDs(nullable))

	if err := connector.Read(ctx, connectors.ReadRequest{Stream: postgresCatalogAlternateReadTable, Config: config}, func(connectors.Record) error { return nil }); err == nil {
		t.Fatal("PostgreSQL connection-level cursor_field unexpectedly worked for a table with a different cursor column")
	}
	t.Log("live PostgreSQL connection-level cursor_field=sequence: alternate_events rejected because it requires alternate_cursor")
}

func collectPostgresRead(t *testing.T, ctx context.Context, connector native.Connector, request connectors.ReadRequest) []connectors.Record {
	t.Helper()
	records := make([]connectors.Record, 0)
	if err := connector.Read(ctx, request, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatal("PostgreSQL live source read failed")
	}
	return records
}

func clonePostgresCatalogConfig(config map[string]string) map[string]string {
	clone := make(map[string]string, len(config))
	for key, value := range config {
		clone[key] = value
	}
	return clone
}

func assertLiveReadCatalog(t *testing.T, catalog connectors.Catalog, name string) {
	t.Helper()
	for _, stream := range catalog.Streams {
		if stream.Name != postgresCatalogAlphaSchema+"."+name {
			continue
		}
		if strings.Join(stream.PrimaryKey, ",") != "id" || len(stream.Fields) != 3 || strings.Join([]string{stream.Fields[0].Type, stream.Fields[1].Type, stream.Fields[2].Type}, ",") != "integer,integer,string" {
			t.Fatal("PostgreSQL live catalog did not report the seeded read table's primary key and types")
		}
		return
	}
	t.Fatal("PostgreSQL live catalog omitted the seeded read table")
}

func assertLiveReadIDs(t *testing.T, records []connectors.Record, want []string) {
	t.Helper()
	if got := liveReadIDs(records); got != strings.Join(want, ",") {
		t.Fatalf("PostgreSQL live read ids = %s, want %s", got, strings.Join(want, ","))
	}
}

func assertLiveReadIDSet(t *testing.T, records []connectors.Record, want []string) {
	t.Helper()
	got := strings.Split(liveReadIDs(records), ",")
	if len(got) != len(want) {
		t.Fatalf("PostgreSQL live read returned %d records, want %d", len(got), len(want))
	}
	for _, expected := range want {
		found := false
		for _, actual := range got {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("PostgreSQL live read ids = %s, missing %s", liveReadIDs(records), expected)
		}
	}
}

func liveReadIDs(records []connectors.Record) string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, fmt.Sprint(record["id"]))
	}
	return strings.Join(ids, ",")
}

func liveReadLabels(records []connectors.Record) string {
	labels := make([]string, 0, len(records))
	for _, record := range records {
		labels = append(labels, fmt.Sprint(record["label"]))
	}
	return strings.Join(labels, ",")
}

type catalogOracleColumn struct {
	relation string
	column   string
	ordinal  int
	nullable bool
}

func assertCatalogMatchesInformationSchema(t *testing.T, ctx context.Context, source *pgx.Conn, catalog database.Catalog, schema string) {
	t.Helper()
	if catalog.Ref().Name != postgresCatalogIntegrationDatabase {
		t.Fatal("typed PostgreSQL catalog did not retain the configured database identity")
	}
	rows, err := source.Query(ctx, `
	SELECT columns.table_name, columns.column_name, columns.ordinal_position, columns.is_nullable
FROM information_schema.columns AS columns
JOIN information_schema.tables AS tables
  ON tables.table_schema = columns.table_schema
 AND tables.table_name = columns.table_name
WHERE columns.table_schema = $1
  AND tables.table_type = 'BASE TABLE'
ORDER BY columns.table_name, columns.ordinal_position`, schema)
	if err != nil {
		t.Fatal("could not inspect PostgreSQL information_schema column metadata")
	}
	defer rows.Close()

	oracle := make([]catalogOracleColumn, 0)
	for rows.Next() {
		var item catalogOracleColumn
		var nullable string
		if err := rows.Scan(&item.relation, &item.column, &item.ordinal, &nullable); err != nil {
			t.Fatal("could not scan PostgreSQL information_schema column metadata")
		}
		item.nullable = nullable == "YES"
		oracle = append(oracle, item)
	}
	if err := rows.Err(); err != nil {
		t.Fatal("could not finish PostgreSQL information_schema column metadata")
	}

	discovered := make([]catalogOracleColumn, 0)
	for _, relation := range catalog.Relations() {
		if relation.Ref.Schema.Catalog.Name != postgresCatalogIntegrationDatabase || relation.Ref.Schema.Name != schema {
			t.Fatal("typed PostgreSQL catalog collapsed or escaped its configured schema identity")
		}
		for _, column := range relation.Columns {
			discovered = append(discovered, catalogOracleColumn{
				relation: relation.Ref.Name,
				column:   column.Ref.Name,
				ordinal:  column.Ordinal,
				nullable: column.Nullable,
			})
		}
	}
	if len(discovered) != len(oracle) {
		t.Fatal("typed PostgreSQL catalog column count disagrees with information_schema")
	}
	for index := range oracle {
		if discovered[index] != oracle[index] {
			t.Fatal("typed PostgreSQL catalog column metadata disagrees with information_schema")
		}
	}
}

type catalogOracleKey struct {
	relation string
	name     string
	kind     database.KeyKind
	columns  []string
}

func assertCatalogKeysMatchInformationSchema(t *testing.T, ctx context.Context, source *pgx.Conn, catalog database.Catalog, schema string) {
	t.Helper()
	rows, err := source.Query(ctx, `
SELECT tc.table_name, tc.constraint_name, tc.constraint_type, kcu.column_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
  ON kcu.constraint_catalog = tc.constraint_catalog
 AND kcu.constraint_schema = tc.constraint_schema
 AND kcu.constraint_name = tc.constraint_name
WHERE tc.table_schema = $1
  AND tc.constraint_type IN ('PRIMARY KEY', 'UNIQUE')
ORDER BY tc.table_name, tc.constraint_name, kcu.ordinal_position`, schema)
	if err != nil {
		t.Fatal("could not inspect PostgreSQL information_schema key metadata")
	}
	defer rows.Close()

	oracle := make([]catalogOracleKey, 0)
	for rows.Next() {
		var relation, name, kindText, column string
		if err := rows.Scan(&relation, &name, &kindText, &column); err != nil {
			t.Fatal("could not scan PostgreSQL information_schema key metadata")
		}
		kind := database.KeyUnique
		if kindText == "PRIMARY KEY" {
			kind = database.KeyPrimary
		}
		if len(oracle) == 0 || oracle[len(oracle)-1].relation != relation || oracle[len(oracle)-1].name != name {
			oracle = append(oracle, catalogOracleKey{relation: relation, name: name, kind: kind})
		}
		oracle[len(oracle)-1].columns = append(oracle[len(oracle)-1].columns, column)
	}
	if err := rows.Err(); err != nil {
		t.Fatal("could not finish PostgreSQL information_schema key metadata")
	}

	discovered := make([]catalogOracleKey, 0)
	for _, relation := range catalog.Relations() {
		for _, key := range relation.Keys {
			item := catalogOracleKey{relation: relation.Ref.Name, name: key.Name, kind: key.Kind}
			for _, column := range key.Columns {
				item.columns = append(item.columns, column.Name)
			}
			discovered = append(discovered, item)
		}
	}
	if len(discovered) != len(oracle) {
		t.Fatal("typed PostgreSQL catalog key count disagrees with information_schema")
	}
	for index := range oracle {
		if discovered[index].relation != oracle[index].relation || discovered[index].name != oracle[index].name || discovered[index].kind != oracle[index].kind || strings.Join(discovered[index].columns, ",") != strings.Join(oracle[index].columns, ",") {
			t.Fatal("typed PostgreSQL catalog key metadata disagrees with information_schema")
		}
	}
}

func assertAlphaTypedCatalog(t *testing.T, catalog database.Catalog) {
	t.Helper()
	accounts := catalogRelation(t, catalog, postgresCatalogAlphaSchema, "accounts")
	assertTypedColumn(t, accounts, "account_id", 1, false, "int8", nil, database.LogicalSignedInteger, 64, 0, 0, false)
	assertTypedColumn(t, accounts, "total", 2, false, "numeric", []string{"precision-12", "scale-2"}, database.LogicalDecimal, 0, 12, 2, false)
	assertTypedColumn(t, accounts, "occurred_at", 3, false, "timestamptz", []string{"precision-3"}, database.LogicalTimestamp, 0, 3, 0, true)
	assertTypedColumn(t, accounts, "body", 4, true, "jsonb", nil, database.LogicalJSON, 0, 0, 0, false)
	assertKey(t, accounts, "accounts_pkey", database.KeyPrimary, []string{"account_id"})

	audit := catalogRelation(t, catalog, postgresCatalogAlphaSchema, "audit")
	assertTypedColumn(t, audit, "event_id", 1, false, "uuid", nil, database.LogicalUUID, 0, 0, 0, false)
	assertCatalogOmitsRelation(t, catalog, postgresCatalogAlphaSchema, "accounts_view")
}

func assertBetaTypedCatalog(t *testing.T, catalog database.Catalog) {
	t.Helper()
	accounts := catalogRelation(t, catalog, postgresCatalogBetaSchema, "accounts")
	assertTypedColumn(t, accounts, "tenant_id", 1, false, "int4", nil, database.LogicalSignedInteger, 32, 0, 0, false)
	assertTypedColumn(t, accounts, "record_id", 2, false, "int4", nil, database.LogicalSignedInteger, 32, 0, 0, false)
	assertTypedColumn(t, accounts, "label", 3, true, "varchar", []string{"length-42"}, database.LogicalString, 0, 0, 0, false)
	assertTypedColumn(t, accounts, "occurred_at", 4, true, "timestamp", []string{"precision-3"}, database.LogicalTimestamp, 0, 3, 0, false)
	assertKey(t, accounts, "accounts_pkey", database.KeyPrimary, []string{"tenant_id", "record_id"})
	assertKey(t, accounts, "accounts_label_key", database.KeyUnique, []string{"label"})
}

func catalogRelation(t *testing.T, catalog database.Catalog, schema, name string) database.Relation {
	t.Helper()
	for _, relation := range catalog.Relations() {
		if relation.Ref.Schema.Name == schema && relation.Ref.Name == name {
			return relation
		}
	}
	t.Fatal("typed PostgreSQL catalog omitted an expected live relation")
	return database.Relation{}
}

func assertCatalogOmitsRelation(t *testing.T, catalog database.Catalog, schema, name string) {
	t.Helper()
	for _, relation := range catalog.Relations() {
		if relation.Ref.Schema.Name == schema && relation.Ref.Name == name {
			t.Fatal("typed PostgreSQL catalog included a non-base relation")
		}
	}
}

func assertTypedColumn(t *testing.T, relation database.Relation, name string, ordinal int, nullable bool, nativeName string, modifiers []string, kind database.LogicalKind, bits uint8, precision, scale uint16, withTimezone bool) {
	t.Helper()
	for _, column := range relation.Columns {
		if column.Ref.Name != name {
			continue
		}
		if column.Ordinal != ordinal || column.Nullable != nullable || column.Type.Kind() != kind || column.Type.BitWidth() != bits || column.Type.Precision() != precision || column.Type.Scale() != scale || column.Type.WithTimezone() != withTimezone || column.Native == nil || column.Native.Name != nativeName || strings.Join(column.Native.Modifiers, ",") != strings.Join(modifiers, ",") {
			t.Fatal("typed PostgreSQL column metadata is incorrect")
		}
		return
	}
	t.Fatal("typed PostgreSQL catalog omitted an expected column")
}

func assertKey(t *testing.T, relation database.Relation, name string, kind database.KeyKind, columns []string) {
	t.Helper()
	for _, key := range relation.Keys {
		if key.Name != name {
			continue
		}
		got := make([]string, 0, len(key.Columns))
		for _, column := range key.Columns {
			got = append(got, column.Name)
		}
		if key.Kind != kind || strings.Join(got, ",") != strings.Join(columns, ",") {
			t.Fatal("typed PostgreSQL key metadata is incorrect")
		}
		return
	}
	t.Fatal("typed PostgreSQL catalog omitted an expected key")
}

func assertLegacyStream(t *testing.T, catalog connectors.Catalog, name string) {
	t.Helper()
	for _, stream := range catalog.Streams {
		if stream.Name != name {
			continue
		}
		if strings.Join(stream.PrimaryKey, ",") != "account_id" || len(stream.Fields) != 4 || strings.Join([]string{stream.Fields[0].Type, stream.Fields[1].Type, stream.Fields[2].Type, stream.Fields[3].Type}, ",") != "integer,number,timestamp,object" {
			t.Fatal("compatibility PostgreSQL catalog did not project the typed live relation")
		}
		return
	}
	t.Fatal("compatibility PostgreSQL catalog omitted its typed live relation")
}
