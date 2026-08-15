//go:build databaseintegration

package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/native/dbtest"
	native "polymetrics.ai/internal/connectors/native/postgres"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

const postgresPollingIntegrationSchema = "polling_transport"

// TestPostgresSharedPollingTransportLive is supporting native evidence for the
// production-composition assertion in internal/app. It exercises the outward
// PostgreSQL transport source so the real call crosses its definition binder,
// PollingPreflight, engine.PollingSourceExecutor, and native keyset runner.
func TestPostgresSharedPollingTransportLive(t *testing.T) {
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
			t.Error("PostgreSQL polling test cleanup failed")
		}
	}()
	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatal("PostgreSQL database container did not start")
	}
	connector := native.New()
	runtime := postgresCatalogConfig(t, endpoint, postgresPollingIntegrationSchema)
	waitForPostgresCatalog(t, ctx, connector, runtime)
	databaseConnection := openPostgresCatalogSource(t, ctx, endpoint)
	defer func() { _ = databaseConnection.Close(context.WithoutCancel(ctx)) }()
	if _, err := databaseConnection.Exec(ctx, `
CREATE SCHEMA polling_transport;
CREATE TABLE polling_transport.events (
  tenant_id bigint NOT NULL,
  event_id bigint NOT NULL,
  sequence bigint NOT NULL,
  payload text NOT NULL,
  PRIMARY KEY (tenant_id, event_id)
);
INSERT INTO polling_transport.events (tenant_id, event_id, sequence, payload) VALUES
  (2, 2, 42, 'fifth'),
  (1, 2, 41, 'second'),
  (1, 1, 41, 'first'),
  (2, 1, 42, 'fourth'),
  (1, 3, 41, 'third');`); err != nil {
		t.Fatal("could not seed PostgreSQL polling relation")
	}

	source := native.NewSnapshotTransportSource(connector)
	request := postgresPollingIntegrationRequest(connector, runtime)
	var pages []synctransport.SourcePage
	if err := source.ReadTransport(ctx, request, func(page synctransport.SourcePage) error {
		pages = append(pages, page)
		return nil
	}); err != nil {
		t.Fatalf("PostgreSQL shared polling source failed: %v", err)
	}
	if got, want := postgresPollingDeliveredKeys(pages), []string{"1/1", "1/2", "1/3", "2/1", "2/2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("delivered key order = %#v, want duplicate-watermark-safe %#v", got, want)
	}
	if len(pages) != 3 || len(pages[0].Records) != 2 || len(pages[1].Records) != 2 || len(pages[2].Records) != 1 {
		t.Fatalf("bounded polling page sizes = %v, want [2 2 1]", postgresPollingPageSizes(pages))
	}

	resumed := request
	checkpoint := postgresPollingCommitCheckpoint(t, pages[0].CandidateCheckpoint)
	resumed.Checkpoint = &checkpoint
	var resumedPages []synctransport.SourcePage
	if err := source.ReadTransport(ctx, resumed, func(page synctransport.SourcePage) error {
		resumedPages = append(resumedPages, page)
		return nil
	}); err != nil {
		t.Fatalf("PostgreSQL polling resume failed: %v", err)
	}
	if got, want := postgresPollingDeliveredKeys(resumedPages), []string{"1/3", "2/1", "2/2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("resumed keys = %#v, want no skip/duplicate %#v", got, want)
	}

	interrupted := errors.New("simulated process death after fetch before acknowledgement")
	var attempted synccontract.CheckpointPosition
	if err := source.ReadTransport(ctx, request, func(page synctransport.SourcePage) error {
		attempted = page.CandidateCheckpoint.Position.Clone()
		return interrupted
	}); !errors.Is(err, interrupted) {
		t.Fatalf("interrupted polling error = %v, want simulated process death", err)
	}
	var replayed synccontract.CheckpointPosition
	if err := source.ReadTransport(ctx, request, func(page synctransport.SourcePage) error {
		replayed = page.CandidateCheckpoint.Position.Clone()
		return interrupted
	}); !errors.Is(err, interrupted) {
		t.Fatalf("replayed polling error = %v, want simulated process death", err)
	}
	if !reflect.DeepEqual(attempted, replayed) {
		t.Fatal("unacknowledged page did not replay from the prior durable tuple")
	}

	if _, err := databaseConnection.Exec(ctx, `ALTER TABLE polling_transport.events ADD COLUMN schema_drift text`); err != nil {
		t.Fatal("could not apply polling schema drift")
	}
	var driftPages int
	err = source.ReadTransport(ctx, resumed, func(synctransport.SourcePage) error { driftPages++; return nil })
	var recovery *synccontract.RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeInvalidCheckpoint {
		t.Fatalf("schema-drift error = %v, want typed rebootstrap", err)
	}
	if driftPages != 0 {
		t.Fatalf("schema drift emitted %d pages, want zero before downstream effects", driftPages)
	}

	canceled, cancelNow := context.WithCancel(ctx)
	cancelNow()
	var canceledPages int
	if err := source.ReadTransport(canceled, request, func(synctransport.SourcePage) error { canceledPages++; return nil }); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled polling error = %v, want context.Canceled", err)
	}
	if canceledPages != 0 {
		t.Fatalf("canceled polling emitted %d pages, want zero", canceledPages)
	}

	if _, err := databaseConnection.Exec(ctx, `TRUNCATE polling_transport.events`); err != nil {
		t.Fatal("could not empty polling relation")
	}
	var emptyPages int
	if err := source.ReadTransport(ctx, request, func(synctransport.SourcePage) error { emptyPages++; return nil }); err != nil {
		t.Fatalf("empty polling source failed: %v", err)
	}
	if emptyPages != 0 {
		t.Fatalf("empty polling source emitted %d pages, want zero sends and no checkpoint", emptyPages)
	}
	if _, err := databaseConnection.Exec(ctx, `INSERT INTO polling_transport.events (tenant_id, event_id, sequence, payload) VALUES (9, 9, 99, 'single')`); err != nil {
		t.Fatal("could not seed single polling row")
	}
	var single []synctransport.SourcePage
	if err := source.ReadTransport(ctx, request, func(page synctransport.SourcePage) error { single = append(single, page); return nil }); err != nil {
		t.Fatalf("single-row polling source failed: %v", err)
	}
	if got := postgresPollingDeliveredKeys(single); !reflect.DeepEqual(got, []string{"9/9"}) {
		t.Fatalf("single-row polling keys = %#v, want one row", got)
	}
}

func postgresPollingIntegrationRequest(connector native.Connector, runtime connectors.RuntimeConfig) synctransport.SourceRequest {
	return synctransport.SourceRequest{
		Connector:   connector,
		Runtime:     runtime,
		Stream:      postgresPollingIntegrationSchema + ".events",
		CursorField: "sequence",
		Mode:        synccontract.ModeIncrementalUpsert,
		BatchSize:   2,
		PrimaryKey:  []string{"tenant_id", "event_id"},
		Resume: synccontract.ResumeExpectation{
			Source: synccontract.SourceIdentity{
				Engine: "postgres", AccountOrCluster: "live_cluster", ObjectScope: postgresPollingIntegrationSchema + ".events",
			},
			SourceGeneration: synccontract.OpaqueToken("live-polling-generation-v1"),
		},
	}
}

func postgresPollingCommitCheckpoint(t *testing.T, candidate synccontract.CheckpointEnvelope) synccontract.CheckpointEnvelope {
	t.Helper()
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement("postgres-polling-integration", candidate.ObservedAt.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	var committed synccontract.CheckpointEnvelope
	if err := synccontract.CommitAfterDownstreamAcknowledgement(candidate, acknowledgement, func(checkpoint synccontract.CheckpointEnvelope) error {
		committed = checkpoint
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return committed
}

func postgresPollingDeliveredKeys(pages []synctransport.SourcePage) []string {
	keys := make([]string, 0)
	for _, page := range pages {
		for _, record := range page.Records {
			keys = append(keys, toIntegrationInt(record["tenant_id"])+"/"+toIntegrationInt(record["event_id"]))
		}
	}
	return keys
}

func postgresPollingPageSizes(pages []synctransport.SourcePage) []int {
	sizes := make([]int, len(pages))
	for index := range pages {
		sizes[index] = len(pages[index].Records)
	}
	return sizes
}

func toIntegrationInt(value any) string {
	switch typed := value.(type) {
	case int64:
		return strconv.FormatInt(typed, 10)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	default:
		return fmt.Sprint(value)
	}
}
