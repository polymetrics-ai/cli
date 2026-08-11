//go:build databaseintegration

package postgres_test

import (
	"context"
	"errors"
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

	postgresCatalogIntegrationEnabledEnv = "POLYMETRICS_DATABASE_INTEGRATION"
	postgresCatalogIntegrationPodmanEnv  = "POLYMETRICS_PODMAN_ENDPOINT"
	postgresCatalogAlphaSchema           = "catalog_alpha"
	postgresCatalogBetaSchema            = "catalog_beta"
	postgresCatalogUnsupportedSchema     = "catalog_unsupported"
)

// TestPostgresDynamicTypedCatalogUsesLiveMetadata is deliberately a live
// regression: a hard-coded table/field list cannot be correct for both
// independently created schemas. The information_schema assertions are an
// independent server-side oracle rather than an expected catalog object built
// by the connector under test.
func TestPostgresDynamicTypedCatalogUsesLiveMetadata(t *testing.T) {
	if os.Getenv(postgresCatalogIntegrationEnabledEnv) != "1" {
		t.Skipf("database integration skipped: set %s=1 to run the PostgreSQL Podman proof", postgresCatalogIntegrationEnabledEnv)
	}
	containerEndpoint := strings.TrimSpace(os.Getenv(postgresCatalogIntegrationPodmanEnv))
	if containerEndpoint == "" {
		t.Skipf("database integration skipped: set %s to an explicit local Podman API endpoint", postgresCatalogIntegrationPodmanEnv)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	harness, err := dbtest.New(dbtest.Config{
		Engine:             "postgres",
		Image:              postgresCatalogIntegrationImage,
		ContainerPort:      5432,
		DataVolumePath:     "/var/lib/postgresql/data",
		ContainerEndpoint:  containerEndpoint,
		ExpectedImageBytes: postgresCatalogIntegrationImageBytes,
		ContainerArgs: []string{
			"--env", "POSTGRES_DB=" + postgresCatalogIntegrationDatabase,
			"--env", "POSTGRES_USER=" + postgresCatalogIntegrationUser,
			"--env", "POSTGRES_HOST_AUTH_METHOD=trust",
		},
	})
	if err != nil {
		t.Fatal("could not configure the PostgreSQL database test harness")
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

	unsupportedConfig := postgresCatalogConfig(t, endpoint, postgresCatalogUnsupportedSchema)
	if _, err := connector.TypedCatalog(ctx, unsupportedConfig); !errors.Is(err, native.ErrUnsupportedCatalogShape) {
		t.Fatal("PostgreSQL typed catalog did not reject an unsupported native type")
	}
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
		"CREATE SCHEMA catalog_beta",
		"CREATE TABLE catalog_beta.accounts (tenant_id integer NOT NULL, record_id integer NOT NULL, label varchar(42), occurred_at timestamp(3), PRIMARY KEY (tenant_id, record_id), UNIQUE (label))",
		"CREATE SCHEMA catalog_unsupported",
		"CREATE TYPE catalog_unsupported.mood AS ENUM ('calm', 'storm')",
		"CREATE TABLE catalog_unsupported.events (id integer PRIMARY KEY, mood catalog_unsupported.mood NOT NULL)",
	}
	for _, statement := range statements {
		if _, err := source.Exec(ctx, statement); err != nil {
			t.Fatal("could not seed deterministic PostgreSQL catalog fixtures")
		}
	}
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
SELECT table_name, column_name, ordinal_position, is_nullable
FROM information_schema.columns
WHERE table_schema = $1
ORDER BY table_name, ordinal_position`, schema)
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
		if stream.Name == name {
			return
		}
	}
	t.Fatal("compatibility PostgreSQL catalog omitted its typed live relation")
}
