//go:build databaseintegration

package postgres_test

import (
	"bytes"
	"context"
	"encoding/json"
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
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
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
	assertPostgresRegisteredSnapshotTransport(t, ctx, connector, endpoint)
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

// assertPostgresRegisteredSnapshotTransport exercises the definition-selected
// source through registry preflight before it reaches the live PostgreSQL 16
// container. Its logs are delivery evidence: rows, source identity, typed
// schema fingerprint, barrier, and candidate checkpoint all come from the
// registered executor rather than a connector.Read compatibility path.
func assertPostgresRegisteredSnapshotTransport(t *testing.T, ctx context.Context, connector native.Connector, endpoint dbtest.Endpoint) {
	t.Helper()
	registry := synctransport.NewRegistry(postgresSnapshotTransportVerifier{})
	if err := native.RegisterPollingTransportSource(registry, connector); err != nil {
		t.Fatalf("register PostgreSQL polling source: %v", err)
	}
	destination := postgresSnapshotTransportDestination{}
	if err := registry.RegisterDestination(postgresSnapshotTransportDestinationExecutor{}); err != nil {
		t.Fatalf("register PostgreSQL snapshot destination: %v", err)
	}
	identity := synccontract.SourceIdentity{
		Engine:           "postgres",
		AccountOrCluster: "postgres-catalog-integration",
		ObjectScope:      postgresCatalogIntegrationDatabase + "." + postgresCatalogAlphaSchema + "." + postgresCatalogReadTable,
	}
	for _, mode := range []synccontract.Mode{synccontract.ModeFullAppend} {
		resolved, err := registry.Preflight(synctransport.PreflightRequest{
			Source:      connector,
			Destination: destination,
			Stream:      "snapshot",
			Mode:        mode,
		})
		if err != nil {
			t.Fatalf("preflight registered PostgreSQL snapshot source for %s: %v", mode, err)
		}

		pages := make([]synctransport.SourcePage, 0, 3)
		if err := resolved.Source.ReadTransport(ctx, synctransport.SourceRequest{
			Connector:   connector,
			Runtime:     postgresCatalogConfig(t, endpoint, postgresCatalogAlphaSchema),
			Stream:      postgresCatalogReadTable,
			CursorField: "sequence",
			PrimaryKey:  []string{"id"},
			Mode:        mode,
			BatchSize:   2,
			Resume: synccontract.ResumeExpectation{
				Source:           identity,
				SourceGeneration: synccontract.OpaqueToken("postgres-catalog-integration-v1"),
			},
		}, func(page synctransport.SourcePage) error {
			pages = append(pages, page)
			return nil
		}); err != nil {
			t.Fatalf("read registered PostgreSQL snapshot source for %s: %v", mode, err)
		}
		if got, want := len(pages), 3; got != want {
			t.Fatalf("registered PostgreSQL snapshot %s pages = %d, want %d", mode, got, want)
		}
		records := make([]connectors.Record, 0, 5)
		for index, page := range pages {
			wantRecords := 2
			if index == len(pages)-1 {
				wantRecords = 1
			}
			if got := len(page.Records); got != wantRecords {
				t.Fatalf("registered PostgreSQL snapshot %s page %d records = %d, want %d", mode, index, got, wantRecords)
			}
			if err := page.CandidateCheckpoint.Validate(); err != nil {
				t.Fatalf("registered PostgreSQL snapshot %s page %d checkpoint: %v", mode, index, err)
			}
			if page.CandidateCheckpoint.Source != identity || page.CandidateCheckpoint.Mechanism != "polling_watermark" || page.CandidateCheckpoint.SchemaVersion == "" || page.CandidateCheckpoint.SnapshotBarrier == nil || len(page.CandidateCheckpoint.SnapshotBarrier.Token) == 0 || page.CandidateCheckpoint.PositionObserved == nil || !*page.CandidateCheckpoint.PositionObserved {
				t.Fatalf("registered PostgreSQL polling %s page %d omitted its resumable identity/schema/checkpoint", mode, index)
			}
			records = append(records, page.Records...)
		}
		assertLiveReadIDs(t, records, []string{"1", "2", "3", "4", "5"})
		assertPostgresSnapshotTypedValues(t, records)
		checkpoint := pages[len(pages)-1].CandidateCheckpoint
		t.Logf(
			"live registered PostgreSQL snapshot mode=%s: rows=%s pages=%d identity=%s/%s/%s schema=%s checkpoint={mechanism=%s barrier_kind=%s barrier=%s dedupe=%x} emitted={event_uuid=%s civil_timestamp=%s json_number=%s json_null=%s}",
			mode, liveReadIDs(records), len(pages), checkpoint.Source.Engine, checkpoint.Source.AccountOrCluster, checkpoint.Source.ObjectScope,
			checkpoint.SchemaVersion, checkpoint.Mechanism, checkpoint.SnapshotBarrier.Kind, checkpoint.SnapshotBarrier.Token, checkpoint.Dedupe.Value,
			postgresSnapshotRecordString(t, records, "1", "event_uuid"), postgresSnapshotRecordString(t, records, "1", "civil_timestamp"),
			postgresSnapshotJSONNumber(t, records, "1", "body", "id"), postgresSnapshotRawJSON(t, records, "2", "body"),
		)
	}
	// Single-column timestamp/UUID keys are intentionally excluded here: the
	// shared source requires a declared, distinct unique tie-breaker. The
	// PostgreSQL polling fixture above uses the real sequence/id tuple that the
	// resumable transport contract persists.
}

func assertPostgresSnapshotTypedValues(t *testing.T, records []connectors.Record) {
	t.Helper()
	if got, want := postgresSnapshotRecordString(t, records, "1", "event_uuid"), "00000000-0000-0000-0000-000000000001"; got != want {
		t.Fatalf("registered PostgreSQL snapshot event_uuid = %q, want %q", got, want)
	}
	if got, want := postgresSnapshotRecordString(t, records, "1", "civil_timestamp"), "2026-08-14T12:34:56.789"; got != want {
		t.Fatalf("registered PostgreSQL snapshot civil timestamp = %q, want zone-less %q", got, want)
	}
	if got, want := postgresSnapshotJSONNumber(t, records, "1", "body", "id"), "9007199254740993"; got != want {
		t.Fatalf("registered PostgreSQL snapshot JSON number = %q, want exact %q", got, want)
	}
	if got, want := postgresSnapshotRawJSON(t, records, "2", "body"), "null"; got != want {
		t.Fatalf("registered PostgreSQL snapshot JSON null = %q, want %q", got, want)
	}
}

func assertPostgresRegisteredSnapshotPaginationEdges(t *testing.T, ctx context.Context, connector native.Connector, endpoint dbtest.Endpoint, registry *synctransport.Registry, destination postgresSnapshotTransportDestination) {
	t.Helper()
	for _, test := range []struct {
		relation string
		field    string
		want     string
	}{
		{relation: "timestamp_key_events", field: "occurred_at", want: "-infinity,2026-08-14T12:34:56.789"},
		{relation: "uuid_key_events", field: "event_id", want: "00000000-0000-0000-0000-000000000101,00000000-0000-0000-0000-000000000102"},
	} {
		t.Run(test.relation, func(t *testing.T) {
			identity := postgresSnapshotIdentity(test.relation)
			pages, records := readPostgresRegisteredSnapshot(t, ctx, connector, endpoint, registry, destination, identity, 1)
			if got, want := len(pages), 3; got != want {
				t.Fatalf("registered PostgreSQL snapshot %s pages = %d, want %d for batch size 1", test.relation, got, want)
			}
			if got, want := postgresSnapshotRecordFieldValues(records, test.field), test.want; got != want {
				t.Fatalf("registered PostgreSQL snapshot %s %s values = %q, want %q", test.relation, test.field, got, want)
			}
			checkpoint := pages[len(pages)-1].CandidateCheckpoint
			t.Logf(
				"live registered PostgreSQL snapshot edge relation=%s rows=%s pages=%d identity=%s/%s/%s schema=%s checkpoint={mechanism=%s barrier_kind=%s barrier=%s dedupe=%x}",
				test.relation, postgresSnapshotRecordFieldValues(records, test.field), len(pages), checkpoint.Source.Engine, checkpoint.Source.AccountOrCluster, checkpoint.Source.ObjectScope,
				checkpoint.SchemaVersion, checkpoint.Mechanism, checkpoint.SnapshotBarrier.Kind, checkpoint.SnapshotBarrier.Token, checkpoint.Dedupe.Value,
			)
		})
	}
}

func readPostgresRegisteredSnapshot(t *testing.T, ctx context.Context, connector native.Connector, endpoint dbtest.Endpoint, registry *synctransport.Registry, destination postgresSnapshotTransportDestination, identity synccontract.SourceIdentity, batchSize int) ([]synctransport.SourcePage, []connectors.Record) {
	t.Helper()
	resolved, err := registry.Preflight(synctransport.PreflightRequest{
		Source:      connector,
		Destination: destination,
		Stream:      "snapshot",
		Mode:        synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatalf("preflight registered PostgreSQL snapshot source for %s: %v", identity.ObjectScope, err)
	}
	pages := make([]synctransport.SourcePage, 0, 3)
	if err := resolved.Source.ReadTransport(ctx, synctransport.SourceRequest{
		Connector: connector,
		Runtime:   postgresCatalogConfig(t, endpoint, postgresCatalogAlphaSchema),
		Stream:    "snapshot",
		Mode:      synccontract.ModeFullAppend,
		BatchSize: batchSize,
		Resume: synccontract.ResumeExpectation{
			Source:           identity,
			SourceGeneration: synccontract.OpaqueToken("postgres-catalog-integration-v1"),
		},
	}, func(page synctransport.SourcePage) error {
		pages = append(pages, page)
		return nil
	}); err != nil {
		t.Fatalf("read registered PostgreSQL snapshot source for %s: %v", identity.ObjectScope, err)
	}
	records := make([]connectors.Record, 0, len(pages)*batchSize)
	for index, page := range pages {
		if err := page.CandidateCheckpoint.Validate(); err != nil {
			t.Fatalf("registered PostgreSQL snapshot %s page %d checkpoint: %v", identity.ObjectScope, index, err)
		}
		if page.CandidateCheckpoint.Source != identity || page.CandidateCheckpoint.SchemaVersion == "" || page.CandidateCheckpoint.SnapshotBarrier == nil || len(page.CandidateCheckpoint.SnapshotBarrier.Token) == 0 || page.CandidateCheckpoint.PositionObserved == nil || *page.CandidateCheckpoint.PositionObserved {
			t.Fatalf("registered PostgreSQL snapshot %s page %d omitted its full-snapshot identity/schema/checkpoint", identity.ObjectScope, index)
		}
		records = append(records, page.Records...)
	}
	return pages, records
}

func postgresSnapshotIdentity(relation string) synccontract.SourceIdentity {
	return synccontract.SourceIdentity{
		Engine:           "postgres",
		AccountOrCluster: "postgres-catalog-integration",
		ObjectScope:      postgresCatalogIntegrationDatabase + "." + postgresCatalogAlphaSchema + "." + relation,
	}
}

func postgresSnapshotRecordString(t *testing.T, records []connectors.Record, id, field string) string {
	t.Helper()
	for _, record := range records {
		if fmt.Sprint(record["id"]) == id {
			return fmt.Sprint(record[field])
		}
	}
	t.Fatalf("registered PostgreSQL snapshot omitted row id %s", id)
	return ""
}

func postgresSnapshotRawJSON(t *testing.T, records []connectors.Record, id, field string) string {
	t.Helper()
	for _, record := range records {
		if fmt.Sprint(record["id"]) != id {
			continue
		}
		raw, ok := record[field].(json.RawMessage)
		if !ok {
			t.Fatalf("registered PostgreSQL snapshot %s row %s = %T, want json.RawMessage", field, id, record[field])
		}
		return string(raw)
	}
	t.Fatalf("registered PostgreSQL snapshot omitted row id %s", id)
	return ""
}

func postgresSnapshotJSONNumber(t *testing.T, records []connectors.Record, id, field, key string) string {
	t.Helper()
	raw := postgresSnapshotRawJSON(t, records, id, field)
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.UseNumber()
	value := map[string]json.Number{}
	if err := decoder.Decode(&value); err != nil {
		t.Fatalf("decode registered PostgreSQL snapshot JSON %s row %s: %v", field, id, err)
	}
	return value[key].String()
}

func postgresSnapshotRecordFieldValues(records []connectors.Record, field string) string {
	values := make([]string, 0, len(records))
	for _, record := range records {
		values = append(values, fmt.Sprint(record[field]))
	}
	return strings.Join(values, ",")
}

var postgresSnapshotTransportDestinationReference = connectors.TransportExecutorReference{
	Family: connectors.TransportExecutorFamilyNativeDatabase,
	ID:     "postgres_snapshot_integration_destination",
}

type postgresSnapshotTransportVerifier struct{}

func (postgresSnapshotTransportVerifier) VerifyTransportConformance(synctransport.ConformanceVerification) error {
	return nil
}

// postgresSnapshotTransportDestination is deliberately inert: it exists only
// to satisfy source-registry preflight in the live source test. It is never
// invoked as an apply path.
type postgresSnapshotTransportDestination struct{}

func (postgresSnapshotTransportDestination) Name() string {
	return "postgres-snapshot-integration-destination"
}

func (postgresSnapshotTransportDestination) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "postgres-snapshot-integration-destination", IntegrationType: "database"}
}

func (postgresSnapshotTransportDestination) Check(context.Context, connectors.RuntimeConfig) error {
	return nil
}

func (postgresSnapshotTransportDestination) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{}, nil
}

func (postgresSnapshotTransportDestination) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	return nil
}

func (postgresSnapshotTransportDestination) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, nil
}

func (postgresSnapshotTransportDestination) Definition() connectors.Definition {
	return connectors.Definition{
		Name:            "postgres-snapshot-integration-destination",
		DisplayName:     "PostgreSQL snapshot integration destination",
		IntegrationType: "database",
		SyncTransport: &connectors.SyncTransportDescriptor{Destination: &connectors.DestinationTransportDescriptor{
			Executor:        postgresSnapshotTransportDestinationReference,
			EligibleActions: []string{"stage_snapshot", "replace_snapshot"},
			Modes:           []synccontract.Mode{synccontract.ModeFullAppend, synccontract.ModeFullOverwrite},
			Delivery: connectors.DeliveryGuarantees{
				Idempotency: connectors.DeliveryIdempotencyAtLeastOnce,
				Ordering:    connectors.DeliveryOrderingSource,
				Deletes:     connectors.DeliveryDeletesUnavailable,
			},
			Conformance:     connectors.ConformanceEvidenceReference{Suite: "postgres_snapshot_integration", RunID: "source_read_v1"},
			Acknowledgement: connectors.TransportAcknowledgementDurableWarehouse,
			ApplyStrategies: []connectors.DestinationApplyStrategy{{
				Mode:     synccontract.ModeFullAppend,
				Strategy: connectors.ApplyStrategyAppend,
				Action:   "stage_snapshot",
			}, {
				Mode:     synccontract.ModeFullOverwrite,
				Strategy: connectors.ApplyStrategyReplace,
				Action:   "replace_snapshot",
			}},
		}},
	}
}

type postgresSnapshotTransportDestinationExecutor struct{}

func (postgresSnapshotTransportDestinationExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return postgresSnapshotTransportDestinationReference
}

func (postgresSnapshotTransportDestinationExecutor) PlanDestination(context.Context, synctransport.DestinationPlanRequest) (synctransport.DestinationPlan, error) {
	return synctransport.DestinationPlan{}, errors.New("PostgreSQL snapshot integration destination must not plan")
}

func (postgresSnapshotTransportDestinationExecutor) ApplyDestination(context.Context, synctransport.DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	return synccontract.DownstreamAcknowledgement{}, errors.New("PostgreSQL snapshot integration destination must not apply")
}

func (postgresSnapshotTransportDestinationExecutor) ReadBackDestination(context.Context, synctransport.DestinationReadBackRequest) error {
	return errors.New("PostgreSQL snapshot integration destination must not read back")
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
		"CREATE TABLE catalog_alpha.read_events (id bigint NOT NULL, sequence bigint NOT NULL, event_uuid uuid NOT NULL, civil_timestamp timestamp(3) NOT NULL, body jsonb NOT NULL, label text NOT NULL, PRIMARY KEY (id))",
		"INSERT INTO catalog_alpha.read_events (id, sequence, event_uuid, civil_timestamp, body, label) VALUES (1, 10, '00000000-0000-0000-0000-000000000001', '2026-08-14 12:34:56.789', '{\"id\":9007199254740993}', 'alpha'), (2, 10, '00000000-0000-0000-0000-000000000002', '2026-08-14 12:34:57.789', 'null', 'bravo'), (3, 11, '00000000-0000-0000-0000-000000000003', '2026-08-14 12:34:58.789', '{\"kind\":\"event\"}', 'charlie'), (4, 12, '00000000-0000-0000-0000-000000000004', '2026-08-14 12:34:59.789', '{\"kind\":\"event\"}', 'delta'), (5, 12, '00000000-0000-0000-0000-000000000005', '2026-08-14 12:35:00.789', '{\"kind\":\"event\"}', 'echo')",
		"CREATE TABLE catalog_alpha.timestamp_key_events (occurred_at timestamp NOT NULL PRIMARY KEY, label text NOT NULL)",
		"INSERT INTO catalog_alpha.timestamp_key_events (occurred_at, label) VALUES ('-infinity', 'negative-infinity'), ('2026-08-14 12:34:56.789', 'finite')",
		"CREATE TABLE catalog_alpha.uuid_key_events (event_id uuid NOT NULL PRIMARY KEY, label text NOT NULL)",
		"INSERT INTO catalog_alpha.uuid_key_events (event_id, label) VALUES ('00000000-0000-0000-0000-000000000101', 'first'), ('00000000-0000-0000-0000-000000000102', 'second')",
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

// assertPostgresLiveReads proves the settled cursor contract through the
// definition-selected shared polling source. It intentionally does not use
// Connector.Read's legacy compatibility reader: #3976 is about the source
// path that owns resumable ETL checkpoints.
func assertPostgresLiveReads(t *testing.T, ctx context.Context, connector native.Connector, endpoint dbtest.Endpoint) {
	t.Helper()
	config := postgresCatalogConfig(t, endpoint, postgresCatalogAlphaSchema)

	catalog, err := connector.Catalog(ctx, config)
	if err != nil {
		t.Fatal("PostgreSQL live read catalog discovery failed")
	}
	assertLiveReadCatalog(t, catalog, postgresCatalogReadTable)

	full := collectPostgresPollingTransport(t, ctx, connector, config, postgresCatalogReadTable, "sequence", "id", nil)
	assertLiveReadIDs(t, full, []string{"1", "2", "3", "4", "5"})
	t.Logf("live PostgreSQL polling full read %s: ids=%s labels=%s", postgresCatalogReadTable, liveReadIDs(full), liveReadLabels(full))

	missingCursor := postgresPollingLiveRequest(connector, config, postgresCatalogReadTable, "", "id", nil)
	if err := native.NewPollingTransportSource(connector).ReadTransport(ctx, missingCursor, func(synctransport.SourcePage) error { return nil }); !errors.Is(err, native.ErrPollingCursorFieldRequired) {
		t.Fatalf("live PostgreSQL missing stream cursor error = %v, want typed pre-I/O refusal", err)
	}

	missingColumn := postgresPollingLiveRequest(connector, config, postgresCatalogReadTable, "missing_sequence", "id", nil)
	if err := native.NewPollingTransportSource(connector).ReadTransport(ctx, missingColumn, func(synctransport.SourcePage) error { return nil }); !errors.Is(err, native.ErrPollingCursorFieldNotFound) {
		t.Fatalf("live PostgreSQL nonexistent stream cursor error = %v, want typed catalog refusal", err)
	}

	nullable := postgresPollingLiveRequest(connector, config, postgresCatalogNullableReadTable, "cursor_value", "id", nil)
	if err := native.NewPollingTransportSource(connector).ReadTransport(ctx, nullable, func(synctransport.SourcePage) error { return nil }); !errors.Is(err, native.ErrPollingCursorNullable) {
		t.Fatalf("live PostgreSQL nullable cursor error = %v, want typed no-omission refusal", err)
	}

	alternate := collectPostgresPollingTransport(t, ctx, connector, config, postgresCatalogAlternateReadTable, "alternate_cursor", "id", nil)
	assertLiveReadIDs(t, alternate, []string{"11"})
	t.Log("live PostgreSQL polling uses cursor fields independently per stream; no connection-level cursor leaks across relations")
}

func collectPostgresPollingTransport(t *testing.T, ctx context.Context, connector native.Connector, config connectors.RuntimeConfig, stream, cursor, primaryKey string, checkpoint *synccontract.CheckpointEnvelope) []connectors.Record {
	t.Helper()
	request := postgresPollingLiveRequest(connector, config, stream, cursor, primaryKey, checkpoint)
	records := make([]connectors.Record, 0)
	if err := native.NewPollingTransportSource(connector).ReadTransport(ctx, request, func(page synctransport.SourcePage) error {
		records = append(records, page.Records...)
		return nil
	}); err != nil {
		t.Fatalf("PostgreSQL live polling source read failed: %v", err)
	}
	return records
}

func postgresPollingLiveRequest(connector connectors.Connector, config connectors.RuntimeConfig, stream, cursor, primaryKey string, checkpoint *synccontract.CheckpointEnvelope) synctransport.SourceRequest {
	return synctransport.SourceRequest{
		Connector: connector, Runtime: config, Stream: stream, CursorField: cursor, PrimaryKey: []string{primaryKey},
		Mode: synccontract.ModeIncrementalDedupe, BatchSize: 10, Checkpoint: checkpoint,
		Resume: synccontract.ResumeExpectation{Source: synccontract.SourceIdentity{Engine: "postgres", AccountOrCluster: "postgres-catalog-integration", ObjectScope: stream}, SourceGeneration: synccontract.OpaqueToken("postgres-catalog-integration-v1")},
	}
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
		if strings.Join(stream.PrimaryKey, ",") != "id" || len(stream.Fields) != 6 || strings.Join([]string{stream.Fields[0].Type, stream.Fields[1].Type, stream.Fields[2].Type, stream.Fields[3].Type, stream.Fields[4].Type, stream.Fields[5].Type}, ",") != "integer,integer,string,timestamp,object,string" {
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
