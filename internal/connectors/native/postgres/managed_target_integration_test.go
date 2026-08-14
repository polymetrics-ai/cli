//go:build databaseintegration

package postgres_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/connectors/native/dbtest"
	native "polymetrics.ai/internal/connectors/native/postgres"
	"polymetrics.ai/internal/warehouse"
)

const postgresManagedTargetRestrictedUser = "pm_managed_target_no_control"

// TestPostgresManagedTargetDriverLiveControlAssertions proves PostgreSQL's
// native identity and control-record observer against a server that it did not
// initialize. Every refusal compares durable state before and after the call,
// so a returned error alone can never satisfy this regression.
func TestPostgresManagedTargetDriverLiveControlAssertions(t *testing.T) {
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
	}()

	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatal("PostgreSQL database container did not start")
	}
	waitForPostgresCatalog(t, ctx, native.New(), postgresCatalogConfig(t, endpoint, postgresCatalogAlphaSchema))
	source := openPostgresCatalogSource(t, ctx, endpoint)
	defer func() { _ = source.Close(ctx) }()

	driver, err := native.NewDatabaseDriver(source)
	if err != nil {
		t.Fatal("could not construct PostgreSQL managed target driver")
	}
	assertPostgresManagedTargetDurability(t, ctx, source, driver)

	fixture := newPostgresManagedTargetFixture(t, ctx, driver)
	assertPostgresManagedTargetMappingFence(t, ctx, driver, fixture)
	seedPostgresManagedTargetControl(t, ctx, source, fixture)
	provisioner, err := database.NewManagedTargetProvisioner(driver)
	if err != nil {
		t.Fatal("could not construct managed target provisioner")
	}

	beforeExact := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	control, err := provisioner.CreateOrAssert(ctx, fixture.plan)
	if err != nil {
		t.Fatalf("exact PostgreSQL managed target assertion failed: %v", err)
	}
	if control.NativeIdentity().Value != beforeExact.relationOID || control.Schema() != fixture.schema {
		t.Fatal("exact PostgreSQL managed target assertion did not return the observed control identity")
	}
	assertPostgresManagedTargetState(t, ctx, source, fixture.target, beforeExact)

	ledgerKey, err := database.NewManagedTargetDeliveryLedgerKey(control)
	if err != nil {
		t.Fatal("could not derive PostgreSQL managed target ledger key")
	}
	if _, found, err := driver.LoadManagedTargetDelivery(ctx, ledgerKey); err != nil || found {
		t.Fatal("new PostgreSQL managed target ledger was not observably empty")
	}
	receipt, err := database.NewManagedTargetDeliveryRecord("delivery-managed-target-1")
	if err != nil {
		t.Fatal("could not create PostgreSQL managed target delivery record")
	}
	if err := driver.StoreManagedTargetDelivery(ctx, ledgerKey, receipt); err != nil {
		t.Fatal("could not persist PostgreSQL managed target delivery record")
	}
	stored, found, err := driver.LoadManagedTargetDelivery(ctx, ledgerKey)
	if err != nil || !found || stored.DeliveryID() != receipt.DeliveryID() {
		t.Fatal("PostgreSQL managed target ledger did not return the committed delivery record")
	}
	afterLedger := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	if afterLedger.ledgerRows != beforeExact.ledgerRows+1 {
		t.Fatal("PostgreSQL managed target ledger store did not create exactly one durable row")
	}

	assertPostgresManagedTargetSchemaDriftRefusal(t, ctx, source, provisioner, fixture)
	assertPostgresManagedTargetForeignOwnerRefusal(t, ctx, source, provisioner, fixture)
	assertPostgresManagedTargetControlCollisionRefusal(t, ctx, source, provisioner, fixture)
	assertPostgresManagedTargetPermissionRefusal(t, ctx, endpoint, source, fixture)
	assertPostgresManagedTargetOIDReplacementRefusal(t, ctx, source, provisioner, fixture)
}

func assertPostgresManagedTargetMappingFence(t *testing.T, ctx context.Context, driver *native.DatabaseDriver, fixture postgresManagedTargetFixture) {
	t.Helper()
	provisioner, err := database.NewManagedTargetProvisioner(driver)
	if err != nil {
		t.Fatal("could not construct PostgreSQL managed target provisioner for mapping fence")
	}
	if _, err := provisioner.CreateOrAssert(ctx, fixture.plan); !errors.Is(err, database.ErrManagedTargetUnverifiable) {
		t.Fatalf("PostgreSQL incomplete mapping create error = %v, want ErrManagedTargetUnverifiable", err)
	}
	after, err := driver.ObserveManagedTarget(ctx, fixture.target)
	if err != nil {
		t.Fatal("could not observe PostgreSQL target after incomplete mapping refusal")
	}
	if after.NamespacePresent || after.RelationPresent || after.ControlState != database.ManagedTargetControlAbsent {
		t.Fatal("PostgreSQL incomplete mapping refusal created namespace, relation, or control state")
	}
}

type postgresManagedTargetFixture struct {
	owner          database.TargetOwner
	target         database.ManagedTargetRef
	targetDatabase database.TargetDatabaseIdentity
	schema         database.ManagedTargetSchema
	plan           database.ManagedTargetProvisioningPlan
}

func newPostgresManagedTargetFixture(t *testing.T, ctx context.Context, driver *native.DatabaseDriver) postgresManagedTargetFixture {
	t.Helper()
	identity := database.ConnectionIdentity{
		WorkspaceID:  "managed-target-workspace",
		ConnectorID:  "postgres",
		ConnectionID: "managed-target-connection",
	}
	artifact, err := warehouse.NewArtifactRef(identity, "managed_target_orders")
	if err != nil {
		t.Fatal("could not create managed target source artifact")
	}
	owner, err := database.NewTargetOwner(identity)
	if err != nil {
		t.Fatal("could not create managed target owner")
	}
	target, err := database.NewManagedTargetRef(owner, artifact, "managed-target-stream")
	if err != nil {
		t.Fatal("could not create managed target reference")
	}
	absent, err := driver.ObserveManagedTarget(ctx, target)
	if err != nil {
		t.Fatal("could not observe absent PostgreSQL managed target")
	}
	if absent.NamespacePresent || absent.RelationPresent || absent.TargetDatabase.Value() == "" {
		t.Fatal("PostgreSQL absent managed target observation did not expose only the live database identity")
	}
	fingerprint := sha256.Sum256([]byte("postgres-managed-target-live-schema-v1"))
	schema, err := database.NewManagedTargetSchema(1, database.SchemaFingerprint(fingerprint))
	if err != nil {
		t.Fatal("could not create managed target schema assertion")
	}
	plan, err := database.NewManagedTargetProvisioningPlan(owner, target, absent.TargetDatabase, schema)
	if err != nil {
		t.Fatal("could not create managed target provisioning plan")
	}
	return postgresManagedTargetFixture{
		owner:          owner,
		target:         target,
		targetDatabase: absent.TargetDatabase,
		schema:         schema,
		plan:           plan,
	}
}

func seedPostgresManagedTargetControl(t *testing.T, ctx context.Context, source *pgx.Conn, fixture postgresManagedTargetFixture) {
	t.Helper()
	namespace := postgresManagedTargetQualifiedName(fixture.target.Namespace())
	ownerTable := postgresManagedTargetQualifiedName(fixture.target.Namespace(), "__polymetrics_namespace_owner")
	controlTable := postgresManagedTargetQualifiedName(fixture.target.Namespace(), "__polymetrics_target_control")
	ledgerTable := postgresManagedTargetQualifiedName(fixture.target.Namespace(), "__polymetrics_delivery_ledger")
	relation := postgresManagedTargetQualifiedName(fixture.target.Namespace(), fixture.target.Relation())
	statements := []string{
		"CREATE SCHEMA " + namespace,
		"CREATE TABLE " + ownerTable + " (workspace_id text NOT NULL, connector_id text NOT NULL, connection_id text NOT NULL, target_database_oid text NOT NULL, namespace_oid text NOT NULL)",
		"CREATE TABLE " + controlTable + " (workspace_id text NOT NULL, connector_id text NOT NULL, connection_id text NOT NULL, stream_id text NOT NULL, relation_name text NOT NULL PRIMARY KEY, target_database_oid text NOT NULL, relation_oid text NOT NULL, schema_version bigint NOT NULL, schema_fingerprint bytea NOT NULL)",
		"CREATE TABLE " + ledgerTable + " (stream_id text NOT NULL PRIMARY KEY, relation_name text NOT NULL, target_database_oid text NOT NULL, delivery_id text NOT NULL)",
		"CREATE TABLE " + relation + " (id bigint NOT NULL PRIMARY KEY)",
	}
	for _, statement := range statements {
		if _, err := source.Exec(ctx, statement); err != nil {
			t.Fatal("could not seed independent PostgreSQL managed target control state")
		}
	}
	var namespaceOID, relationOID string
	if err := source.QueryRow(ctx, "SELECT oid::text FROM pg_catalog.pg_namespace WHERE nspname = $1", fixture.target.Namespace()).Scan(&namespaceOID); err != nil {
		t.Fatal("could not observe seeded PostgreSQL namespace identity")
	}
	if err := source.QueryRow(ctx, "SELECT c.oid::text FROM pg_catalog.pg_class AS c JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace WHERE n.nspname = $1 AND c.relname = $2", fixture.target.Namespace(), fixture.target.Relation()).Scan(&relationOID); err != nil {
		t.Fatal("could not observe seeded PostgreSQL relation identity")
	}
	identity := fixture.owner.Identity()
	if _, err := source.Exec(ctx, "INSERT INTO "+ownerTable+" (workspace_id, connector_id, connection_id, target_database_oid, namespace_oid) VALUES ($1, $2, $3, $4, $5)", identity.WorkspaceID, identity.ConnectorID, identity.ConnectionID, fixture.targetDatabase.Value(), namespaceOID); err != nil {
		t.Fatal("could not seed PostgreSQL namespace owner record")
	}
	fingerprint := fixture.schema.Fingerprint()
	if _, err := source.Exec(ctx, "INSERT INTO "+controlTable+" (workspace_id, connector_id, connection_id, stream_id, relation_name, target_database_oid, relation_oid, schema_version, schema_fingerprint) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", identity.WorkspaceID, identity.ConnectorID, identity.ConnectionID, fixture.target.StreamID(), fixture.target.Relation(), fixture.targetDatabase.Value(), relationOID, fixture.schema.Version(), fingerprint[:]); err != nil {
		t.Fatal("could not seed PostgreSQL managed target control record")
	}
}

type postgresManagedTargetState struct {
	namespaceOID      string
	relationOID       string
	ownerRows         int
	controlRows       int
	ledgerRows        int
	ownerConnectionID string
	streamID          string
	fingerprint       string
}

func observePostgresManagedTargetState(t *testing.T, ctx context.Context, source *pgx.Conn, target database.ManagedTargetRef) postgresManagedTargetState {
	t.Helper()
	ownerTable := postgresManagedTargetQualifiedName(target.Namespace(), "__polymetrics_namespace_owner")
	controlTable := postgresManagedTargetQualifiedName(target.Namespace(), "__polymetrics_target_control")
	ledgerTable := postgresManagedTargetQualifiedName(target.Namespace(), "__polymetrics_delivery_ledger")
	var state postgresManagedTargetState
	if err := source.QueryRow(ctx, "SELECT oid::text FROM pg_catalog.pg_namespace WHERE nspname = $1", target.Namespace()).Scan(&state.namespaceOID); err != nil {
		t.Fatal("could not observe PostgreSQL managed target namespace state")
	}
	if err := source.QueryRow(ctx, "SELECT c.oid::text FROM pg_catalog.pg_class AS c JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace WHERE n.nspname = $1 AND c.relname = $2", target.Namespace(), target.Relation()).Scan(&state.relationOID); err != nil {
		t.Fatal("could not observe PostgreSQL managed target relation state")
	}
	if err := source.QueryRow(ctx, "SELECT count(*) FROM "+ownerTable).Scan(&state.ownerRows); err != nil {
		t.Fatal("could not observe PostgreSQL managed target owner state")
	}
	if err := source.QueryRow(ctx, "SELECT count(*) FROM "+controlTable).Scan(&state.controlRows); err != nil {
		t.Fatal("could not observe PostgreSQL managed target control count")
	}
	if err := source.QueryRow(ctx, "SELECT count(*) FROM "+ledgerTable).Scan(&state.ledgerRows); err != nil {
		t.Fatal("could not observe PostgreSQL managed target ledger count")
	}
	if err := source.QueryRow(ctx, "SELECT connection_id FROM "+ownerTable).Scan(&state.ownerConnectionID); err != nil {
		t.Fatal("could not observe PostgreSQL managed target owner row")
	}
	if err := source.QueryRow(ctx, "SELECT stream_id, encode(schema_fingerprint, 'hex') FROM "+controlTable+" WHERE relation_name = $1", target.Relation()).Scan(&state.streamID, &state.fingerprint); err != nil {
		t.Fatal("could not observe PostgreSQL managed target control row")
	}
	return state
}

func assertPostgresManagedTargetState(t *testing.T, ctx context.Context, source *pgx.Conn, target database.ManagedTargetRef, want postgresManagedTargetState) {
	t.Helper()
	if got := observePostgresManagedTargetState(t, ctx, source, target); got != want {
		t.Fatalf("PostgreSQL managed target durable state changed during refusal: got=%+v want=%+v", got, want)
	}
}

func assertPostgresManagedTargetOIDReplacementRefusal(t *testing.T, ctx context.Context, source *pgx.Conn, provisioner *database.ManagedTargetProvisioner, fixture postgresManagedTargetFixture) {
	t.Helper()
	relation := postgresManagedTargetQualifiedName(fixture.target.Namespace(), fixture.target.Relation())
	if _, err := source.Exec(ctx, "DROP TABLE "+relation); err != nil {
		t.Fatal("could not replace the PostgreSQL managed target relation")
	}
	if _, err := source.Exec(ctx, "CREATE TABLE "+relation+" (id bigint NOT NULL PRIMARY KEY)"); err != nil {
		t.Fatal("could not create same-named PostgreSQL replacement relation")
	}
	before := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	if _, err := provisioner.CreateOrAssert(ctx, fixture.plan); !errors.Is(err, database.ErrManagedTargetReplaced) {
		t.Fatalf("PostgreSQL OID replacement error = %v, want ErrManagedTargetReplaced", err)
	}
	assertPostgresManagedTargetState(t, ctx, source, fixture.target, before)
}

func assertPostgresManagedTargetSchemaDriftRefusal(t *testing.T, ctx context.Context, source *pgx.Conn, provisioner *database.ManagedTargetProvisioner, fixture postgresManagedTargetFixture) {
	t.Helper()
	controlTable := postgresManagedTargetQualifiedName(fixture.target.Namespace(), "__polymetrics_target_control")
	fingerprint := sha256.Sum256([]byte("postgres-managed-target-live-schema-v2"))
	if _, err := source.Exec(ctx, "UPDATE "+controlTable+" SET schema_fingerprint = $1 WHERE relation_name = $2", fingerprint[:], fixture.target.Relation()); err != nil {
		t.Fatal("could not create PostgreSQL managed target schema drift")
	}
	before := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	if _, err := provisioner.CreateOrAssert(ctx, fixture.plan); !errors.Is(err, database.ErrManagedTargetSchemaDrift) {
		t.Fatalf("PostgreSQL schema drift error = %v, want ErrManagedTargetSchemaDrift", err)
	}
	assertPostgresManagedTargetState(t, ctx, source, fixture.target, before)
	original := fixture.schema.Fingerprint()
	if _, err := source.Exec(ctx, "UPDATE "+controlTable+" SET schema_fingerprint = $1 WHERE relation_name = $2", original[:], fixture.target.Relation()); err != nil {
		t.Fatal("could not restore PostgreSQL managed target schema assertion")
	}
}

func assertPostgresManagedTargetForeignOwnerRefusal(t *testing.T, ctx context.Context, source *pgx.Conn, provisioner *database.ManagedTargetProvisioner, fixture postgresManagedTargetFixture) {
	t.Helper()
	ownerTable := postgresManagedTargetQualifiedName(fixture.target.Namespace(), "__polymetrics_namespace_owner")
	if _, err := source.Exec(ctx, "UPDATE "+ownerTable+" SET connection_id = $1", "foreign-managed-target-connection"); err != nil {
		t.Fatal("could not create a foreign PostgreSQL namespace owner record")
	}
	before := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	if _, err := provisioner.CreateOrAssert(ctx, fixture.plan); !errors.Is(err, database.ErrManagedTargetUnverifiable) {
		t.Fatalf("PostgreSQL foreign owner error = %v, want fail-closed ErrManagedTargetUnverifiable", err)
	}
	assertPostgresManagedTargetState(t, ctx, source, fixture.target, before)
	identity := fixture.owner.Identity()
	if _, err := source.Exec(ctx, "UPDATE "+ownerTable+" SET connection_id = $1", identity.ConnectionID); err != nil {
		t.Fatal("could not restore PostgreSQL managed target namespace owner")
	}
}

func assertPostgresManagedTargetControlCollisionRefusal(t *testing.T, ctx context.Context, source *pgx.Conn, provisioner *database.ManagedTargetProvisioner, fixture postgresManagedTargetFixture) {
	t.Helper()
	controlTable := postgresManagedTargetQualifiedName(fixture.target.Namespace(), "__polymetrics_target_control")
	if _, err := source.Exec(ctx, "UPDATE "+controlTable+" SET stream_id = $1 WHERE relation_name = $2", "other-managed-target-stream", fixture.target.Relation()); err != nil {
		t.Fatal("could not create a PostgreSQL managed target control collision")
	}
	before := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	if _, err := provisioner.CreateOrAssert(ctx, fixture.plan); !errors.Is(err, database.ErrManagedTargetUnverifiable) {
		t.Fatalf("PostgreSQL control collision error = %v, want fail-closed ErrManagedTargetUnverifiable", err)
	}
	assertPostgresManagedTargetState(t, ctx, source, fixture.target, before)
	if _, err := source.Exec(ctx, "UPDATE "+controlTable+" SET stream_id = $1 WHERE relation_name = $2", fixture.target.StreamID(), fixture.target.Relation()); err != nil {
		t.Fatal("could not restore PostgreSQL managed target control record")
	}
}

func assertPostgresManagedTargetPermissionRefusal(t *testing.T, ctx context.Context, endpoint dbtest.Endpoint, source *pgx.Conn, fixture postgresManagedTargetFixture) {
	t.Helper()
	if _, err := source.Exec(ctx, "CREATE ROLE "+postgresManagedTargetRestrictedUser+" LOGIN"); err != nil {
		t.Fatal("could not create PostgreSQL managed target restricted role")
	}
	if _, err := source.Exec(ctx, "GRANT CONNECT ON DATABASE "+postgresManagedTargetQualifiedName(postgresCatalogIntegrationDatabase)+" TO "+postgresManagedTargetQualifiedName(postgresManagedTargetRestrictedUser)); err != nil {
		t.Fatal("could not grant PostgreSQL managed target restricted database access")
	}
	if _, err := source.Exec(ctx, "GRANT USAGE ON SCHEMA "+postgresManagedTargetQualifiedName(fixture.target.Namespace())+" TO "+postgresManagedTargetQualifiedName(postgresManagedTargetRestrictedUser)); err != nil {
		t.Fatal("could not grant PostgreSQL managed target restricted schema usage")
	}
	restricted := openPostgresCatalogUser(t, ctx, endpoint, postgresManagedTargetRestrictedUser)
	defer func() { _ = restricted.Close(ctx) }()
	driver, err := native.NewDatabaseDriver(restricted)
	if err != nil {
		t.Fatal("could not construct restricted PostgreSQL managed target driver")
	}
	var databaseOID, namespaceOID string
	if err := restricted.QueryRow(ctx, "SELECT oid::text FROM pg_catalog.pg_database WHERE datname = current_database()").Scan(&databaseOID); err != nil || databaseOID == "" {
		t.Fatal("restricted PostgreSQL session could not observe its destination database identity")
	}
	if err := restricted.QueryRow(ctx, "SELECT oid::text FROM pg_catalog.pg_namespace WHERE nspname = $1", fixture.target.Namespace()).Scan(&namespaceOID); err != nil || namespaceOID == "" {
		t.Fatal("restricted PostgreSQL session could not observe the derived namespace identity")
	}
	provisioner, err := database.NewManagedTargetProvisioner(driver)
	if err != nil {
		t.Fatal("could not construct restricted PostgreSQL managed target provisioner")
	}
	observation, err := driver.ObserveManagedTarget(ctx, fixture.target)
	if err != nil {
		t.Fatal("restricted PostgreSQL managed target observation did not reach the private control boundary")
	}
	if observation.NamespaceOwnerState != database.ManagedTargetNamespaceOwnerUnreadable || observation.ControlState != database.ManagedTargetControlUnreadable {
		t.Fatalf("restricted PostgreSQL managed target observation = namespace_owner=%d control=%d, want both unreadable", observation.NamespaceOwnerState, observation.ControlState)
	}
	before := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	if _, err := provisioner.CreateOrAssert(ctx, fixture.plan); !errors.Is(err, database.ErrManagedTargetOwnerUnreadable) {
		t.Fatalf("PostgreSQL managed target permission error = %v, want ErrManagedTargetOwnerUnreadable", err)
	}
	assertPostgresManagedTargetState(t, ctx, source, fixture.target, before)
}

func assertPostgresManagedTargetDurability(t *testing.T, ctx context.Context, source *pgx.Conn, driver *native.DatabaseDriver) {
	t.Helper()
	if err := driver.PreflightDurability(ctx); err != nil {
		t.Fatalf("PostgreSQL default durability preflight failed: %v", err)
	}
	if _, err := source.Exec(ctx, "SET synchronous_commit = off"); err != nil {
		t.Fatal("could not set PostgreSQL session durability test value")
	}
	var setting string
	if err := source.QueryRow(ctx, "SELECT current_setting('synchronous_commit')").Scan(&setting); err != nil || setting != "off" {
		t.Fatal("PostgreSQL session did not apply unsafe durability test setting")
	}
	if err := driver.PreflightDurability(ctx); err == nil {
		t.Fatal("PostgreSQL durability preflight accepted synchronous_commit=off")
	}
	if _, err := source.Exec(ctx, "SET synchronous_commit = on"); err != nil {
		t.Fatal("could not restore PostgreSQL session durability test value")
	}
	if err := driver.PreflightDurability(ctx); err != nil {
		t.Fatal("PostgreSQL durability preflight did not observe restored synchronous_commit=on")
	}
}

func openPostgresCatalogUser(t *testing.T, ctx context.Context, endpoint dbtest.Endpoint, user string) *pgx.Conn {
	t.Helper()
	config, err := pgx.ParseConfig("sslmode=disable")
	if err != nil {
		t.Fatal("could not configure the isolated PostgreSQL restricted source")
	}
	config.Host = endpoint.Host
	config.Port = uint16(endpoint.Port)
	config.Database = postgresCatalogIntegrationDatabase
	config.User = user
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		t.Fatal("could not open the isolated PostgreSQL restricted source")
	}
	return conn
}

func postgresManagedTargetQualifiedName(parts ...string) string {
	quoted := make([]string, 0, len(parts))
	for _, part := range parts {
		quoted = append(quoted, `"`+strings.ReplaceAll(part, `"`, `""`)+`"`)
	}
	return strings.Join(quoted, ".")
}
