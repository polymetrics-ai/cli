package app_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/warehouse"
)

// toComparable renders a scanned value as a stable string so an assertion does
// not depend on which numeric width the engine chose for a column.
func toComparable(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case float64:
		if typed == float64(int64(typed)) {
			return fmt.Sprintf("%d", int64(typed))
		}
		return fmt.Sprintf("%v", typed)
	default:
		return fmt.Sprintf("%v", typed)
	}
}

// readOutboxRecords reads the JSONL the outbox destination writes. The outbox
// is not a warehouse table, so it is unaffected by the table format switch.
func readOutboxRecords(t *testing.T, path string) []connectors.Record {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open outbox %s: %v", path, err)
	}
	defer func() { _ = file.Close() }()
	out := make([]connectors.Record, 0)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var record connectors.Record
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("decode outbox record: %v", err)
		}
		out = append(out, record)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan outbox: %v", err)
	}
	return out
}

// TestQuerySQLAggregatesOverParquetTables proves DuckDB is the engine for
// real. The shipped binary's ceiling was SELECT * FROM <table> [LIMIT n]:
// every query below was unanswerable, not slow. Each asserts on the records
// returned, so a query that silently answered from the wrong table or returned
// nothing would fail rather than pass on a zero exit status.
func TestQuerySQLAggregatesOverParquetTables(t *testing.T) {
	ctx := context.Background()
	a := setupParquetQueryApp(t, ctx)

	if got := a.QueryEngineName(); got != "duckdb" {
		t.Fatalf("QueryEngineName() = %q, want %q", got, "duckdb")
	}

	grouped, err := a.QuerySQL(ctx, app.QuerySQLRequest{SQL: "SELECT status, count(*) AS n FROM records GROUP BY status ORDER BY status"})
	if err != nil {
		t.Fatalf("QuerySQL(GROUP BY) error = %v", err)
	}
	gotGroups := make([]string, 0, len(grouped))
	for _, row := range grouped {
		gotGroups = append(gotGroups, toComparable(row["status"])+"="+toComparable(row["n"]))
	}
	if want := []string{"active=2", "churned=1"}; !reflect.DeepEqual(gotGroups, want) {
		t.Fatalf("QuerySQL(GROUP BY) = %v, want %v", gotGroups, want)
	}

	filtered, err := a.QuerySQL(ctx, app.QuerySQLRequest{SQL: "SELECT id FROM records WHERE status = 'active' ORDER BY id"})
	if err != nil {
		t.Fatalf("QuerySQL(WHERE) error = %v", err)
	}
	gotIDs := make([]string, 0, len(filtered))
	for _, row := range filtered {
		gotIDs = append(gotIDs, toComparable(row["id"]))
	}
	if want := []string{"r1", "r3"}; !reflect.DeepEqual(gotIDs, want) {
		t.Fatalf("QuerySQL(WHERE) = %v, want %v", gotIDs, want)
	}

	aggregated, err := a.QuerySQL(ctx, app.QuerySQLRequest{SQL: "SELECT sum(amount) AS total, max(amount) AS peak FROM records"})
	if err != nil {
		t.Fatalf("QuerySQL(aggregate) error = %v", err)
	}
	if len(aggregated) != 1 {
		t.Fatalf("QuerySQL(aggregate) rows = %d, want 1", len(aggregated))
	}
	if got := toComparable(aggregated[0]["total"]); got != "60" {
		t.Fatalf("sum(amount) = %q, want %q", got, "60")
	}
	if got := toComparable(aggregated[0]["peak"]); got != "30" {
		t.Fatalf("max(amount) = %q, want %q", got, "30")
	}

	// A write must still be refused, whatever the engine can express.
	if _, err := a.QuerySQL(ctx, app.QuerySQLRequest{SQL: "DELETE FROM records"}); err == nil {
		t.Fatal("QuerySQL(DELETE) succeeded; only read-only queries are allowed")
	}
}

// TestReverseETLReadsAParquetSourceTable drives the whole reverse path against
// a Parquet source table — plan, approve, run — and asserts on the records that
// reached the destination, not on the run's status.
func TestReverseETLReadsAParquetSourceTable(t *testing.T) {
	ctx := context.Background()
	a := setupParquetQueryApp(t, ctx)

	// Reverse ETL reads the same materialized table ETL wrote, so the format
	// switch reaches it too. Assert the file it is about to read really is
	// Parquet, or this test would keep passing against JSONL and prove nothing.
	located, err := warehouse.FindTable(filepath.Join(a.ProjectDir(), "warehouse"), "records", "")
	if err != nil {
		t.Fatalf("FindTable(records) error = %v", err)
	}
	if filepath.Ext(located.Path) != warehouse.TableFileExt {
		t.Fatalf("reverse source table %s has extension %q, want %q", located.Path, filepath.Ext(located.Path), warehouse.TableFileExt)
	}

	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "records_to_outbox",
		SourceTable:           "records",
		DestinationConnector:  "outbox",
		DestinationCredential: "outbox-local",
		Action:                "upsert",
		Mappings:              map[string]string{"id": "external_id", "name": "full_name"},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL() error = %v", err)
	}
	if plan.RecordCount != 3 {
		t.Fatalf("plan record count = %d, want 3 rows read from the Parquet table", plan.RecordCount)
	}

	// The plan's own sample is the operator-visible proof that the rows came
	// out of the Parquet table, so it is asserted on content rather than count.
	sampled := make([]string, 0, len(plan.Sample))
	for _, record := range plan.Sample {
		sampled = append(sampled, toComparable(record["external_id"])+":"+toComparable(record["full_name"]))
	}
	sort.Strings(sampled)
	if len(sampled) == 0 {
		t.Fatal("reverse plan carries no sample rows from the Parquet source table")
	}
	for _, entry := range sampled {
		if !strings.Contains(entry, ":") || strings.HasPrefix(entry, ":") {
			t.Fatalf("reverse plan sample %v did not resolve the mapped source columns", sampled)
		}
	}

	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken})
	if err != nil {
		t.Fatalf("RunReverseETL() error = %v", err)
	}
	if run.RecordsSucceeded != 3 {
		t.Fatalf("reverse run wrote %d records, want 3", run.RecordsSucceeded)
	}

	written := readOutboxRecords(t, filepath.Join(a.ProjectDir(), "outbox", "records_to_outbox.jsonl"))
	got := make([]string, 0, len(written))
	for _, record := range written {
		got = append(got, toComparable(record["external_id"])+":"+toComparable(record["full_name"]))
	}
	sort.Strings(got)
	want := []string{"r1:Ada", "r2:Alan", "r3:Grace"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("outbox records = %v, want %v", got, want)
	}
}

// setupParquetQueryApp syncs a small scripted source into the warehouse so the
// tables under test are produced by the real materialization path.
func setupParquetQueryApp(t *testing.T, ctx context.Context) *app.App {
	t.Helper()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	source := newParquetQuerySource()
	a.Registry().Register(source)

	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{Name: "source-local", Connector: source.Name()}); err != nil {
		t.Fatalf("AddCredential(source) error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "warehouse-local",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatalf("AddCredential(warehouse) error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "outbox-local",
		Connector: "outbox",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "outbox")},
	}); err != nil {
		t.Fatalf("AddCredential(outbox) error = %v", err)
	}
	if _, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
		Name:        "records_to_warehouse",
		Source:      app.EndpointConfig{Connector: source.Name(), Credential: "source-local"},
		Destination: app.EndpointConfig{Connector: "warehouse", Credential: "warehouse-local"},
		Streams: map[string]app.StreamConfig{
			"records": {
				SyncMode:         "full_refresh_overwrite",
				CursorField:      "updated_at",
				PrimaryKey:       []string{"id"},
				DestinationTable: "records",
			},
		},
	}); err != nil {
		t.Fatalf("CreateConnection() error = %v", err)
	}
	if _, err := a.RunETL(ctx, app.RunETLRequest{Connection: "records_to_warehouse", Stream: "records"}); err != nil {
		t.Fatalf("RunETL() error = %v", err)
	}
	return a
}

type parquetQuerySource struct {
	records []connectors.Record
}

func newParquetQuerySource() *parquetQuerySource {
	return &parquetQuerySource{records: []connectors.Record{
		{"id": "r1", "name": "Ada", "status": "active", "amount": 10, "updated_at": "2026-08-06T00:00:00Z"},
		{"id": "r2", "name": "Alan", "status": "churned", "amount": 20, "updated_at": "2026-08-06T00:00:01Z"},
		{"id": "r3", "name": "Grace", "status": "active", "amount": 30, "updated_at": "2026-08-06T00:00:02Z"},
	}}
}

func (s *parquetQuerySource) Name() string { return "parquet_query_source" }

func (s *parquetQuerySource) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:         s.Name(),
		DisplayName:  "Parquet Query Source",
		Description:  "Scripted source for Parquet query tests.",
		Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

func (s *parquetQuerySource) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return ctx.Err()
}

func (s *parquetQuerySource) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: s.Name(), Streams: []connectors.Stream{{
		Name:         "records",
		Description:  "Scripted records.",
		PrimaryKey:   []string{"id"},
		CursorFields: []string{"updated_at"},
	}}}, ctx.Err()
}

func (s *parquetQuerySource) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for _, record := range s.records {
		if err := ctx.Err(); err != nil {
			return err
		}
		clone := make(connectors.Record, len(record))
		for key, value := range record {
			clone[key] = value
		}
		if err := emit(clone); err != nil {
			return err
		}
	}
	return nil
}

func (s *parquetQuerySource) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
