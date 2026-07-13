package app

import (
	"context"
	"errors"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

type scriptedSyncSource struct {
	name      string
	records   []connectors.Record
	failAfter int
	requests  []connectors.ReadRequest
}

func newScriptedSyncSource(name string, records []connectors.Record) *scriptedSyncSource {
	return &scriptedSyncSource{name: name, records: records, failAfter: -1}
}

func (s *scriptedSyncSource) Name() string { return s.name }

func (s *scriptedSyncSource) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:         s.Name(),
		DisplayName:  "Scripted Sync Source",
		Description:  "Scripted source for sync-mode tests.",
		Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

func (s *scriptedSyncSource) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return ctx.Err()
}

func (s *scriptedSyncSource) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: s.Name(), Streams: []connectors.Stream{{
		Name:         "records",
		Description:  "Scripted records.",
		PrimaryKey:   []string{"id"},
		CursorFields: []string{"updated_at"},
	}}}, ctx.Err()
}

func (s *scriptedSyncSource) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	s.requests = append(s.requests, req)
	for i, record := range s.records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(cloneRecord(record)); err != nil {
			return err
		}
		if s.failAfter >= 0 && i+1 >= s.failAfter {
			return errors.New("scripted source failure")
		}
	}
	return nil
}

func (s *scriptedSyncSource) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func setupSyncModeApp(t *testing.T, source *scriptedSyncSource, mode string) (*App, string) {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	registry := connectors.NewRegistry()
	registry.Register(source)
	a.registry = registry
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{
		Name:      "warehouse",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "records_to_warehouse",
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
		Streams: map[string]StreamConfig{
			"records": {
				SyncMode:         mode,
				CursorField:      "updated_at",
				PrimaryKey:       []string{"id"},
				DestinationTable: "records",
			},
		},
	}); err != nil {
		t.Fatal(err)
	}
	return a, "records_to_warehouse"
}

func rowsByID(rows []connectors.Record) map[string]connectors.Record {
	out := map[string]connectors.Record{}
	for _, row := range rows {
		out[toComparableString(row["id"])] = row
	}
	return out
}

func TestRunETLLimitCapsWarehouseRead(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("scripted_limited_warehouse", []connectors.Record{
		{"id": "a", "name": "Ada", "updated_at": "2026-01-01T00:00:00Z"},
		{"id": "g", "name": "Grace", "updated_at": "2026-01-02T00:00:00Z"},
		{"id": "k", "name": "Katherine", "updated_at": "2026-01-03T00:00:00Z"},
		{"id": "m", "name": "Margaret", "updated_at": "2026-01-04T00:00:00Z"},
	})
	a, connection := setupSyncModeApp(t, source, "full_refresh_append")

	run, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 10, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(source.requests) != 1 {
		t.Fatalf("source read requests = %d, want 1", len(source.requests))
	}
	if source.requests[0].Limit != 2 {
		t.Fatalf("source ReadRequest.Limit = %d, want 2", source.requests[0].Limit)
	}
	if run.RecordsRead != 2 || run.RecordsLoaded != 2 || run.BatchCount != 1 {
		t.Fatalf("unexpected capped warehouse run counts: %+v", run)
	}
	rows, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("warehouse rows = %d, want 2", len(rows))
	}
	if got := rowsByID(rows); got["a"] == nil || got["g"] == nil || got["k"] != nil {
		t.Fatalf("warehouse rows by id = %+v, want only a and g", got)
	}
}

func TestParseSyncModeMatrix(t *testing.T) {
	tests := []struct {
		raw            string
		source         SourceSyncMode
		destination    DestinationSyncMode
		requiresCursor bool
		requiresPK     bool
	}{
		{"full_refresh_append", SourceSyncFullRefresh, DestinationSyncAppend, false, false},
		{"full_refresh_overwrite", SourceSyncFullRefresh, DestinationSyncOverwrite, false, false},
		{"full_refresh_overwrite_deduped", SourceSyncFullRefresh, DestinationSyncOverwriteDeduped, true, true},
		{"incremental_append", SourceSyncIncremental, DestinationSyncAppend, true, false},
		{"incremental_append_deduped", SourceSyncIncremental, DestinationSyncAppendDeduped, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			mode, err := ParseSyncMode(tt.raw)
			if err != nil {
				t.Fatalf("ParseSyncMode() error = %v", err)
			}
			if mode.Source != tt.source || mode.Destination != tt.destination {
				t.Fatalf("mode = %+v, want source=%s destination=%s", mode, tt.source, tt.destination)
			}
			if mode.RequiresCursor() != tt.requiresCursor || mode.RequiresPrimaryKey() != tt.requiresPK {
				t.Fatalf("requirements cursor=%v pk=%v, want cursor=%v pk=%v", mode.RequiresCursor(), mode.RequiresPrimaryKey(), tt.requiresCursor, tt.requiresPK)
			}
		})
	}
	if _, err := ParseSyncMode("full_refresh_replace"); err == nil {
		t.Fatal("ParseSyncMode(invalid) error = nil")
	}
}

func TestValidateSyncModeRequirements(t *testing.T) {
	if err := ValidateStreamSyncConfig(StreamConfig{SyncMode: "incremental_append"}); err == nil {
		t.Fatal("ValidateStreamSyncConfig(incremental without cursor) error = nil")
	}
	if err := ValidateStreamSyncConfig(StreamConfig{SyncMode: "incremental_append_deduped", CursorField: "updated_at"}); err == nil {
		t.Fatal("ValidateStreamSyncConfig(deduped without primary key) error = nil")
	}
	if err := ValidateStreamSyncConfig(StreamConfig{SyncMode: "incremental_append_deduped", CursorField: "updated_at", PrimaryKey: []string{"id"}}); err != nil {
		t.Fatalf("ValidateStreamSyncConfig(valid) error = %v", err)
	}
}

func TestFullRefreshAppendDuplicatesAcrossRuns(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("scripted_append", []connectors.Record{
		{"id": "a", "name": "Ada", "updated_at": "2026-01-01T00:00:00Z"},
		{"id": "g", "name": "Grace", "updated_at": "2026-01-02T00:00:00Z"},
	})
	a, connection := setupSyncModeApp(t, source, "full_refresh_append")

	for i := 0; i < 2; i++ {
		run, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1})
		if err != nil {
			t.Fatalf("RunETL(%d) error = %v", i, err)
		}
		if run.Checkpoint["sync_mode"] != "full_refresh_append" {
			t.Fatalf("sync_mode checkpoint = %q", run.Checkpoint["sync_mode"])
		}
	}
	rows, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 {
		t.Fatalf("rows len = %d, want 4 append duplicates", len(rows))
	}
}

func TestFullRefreshOverwriteUsesFailureSafeFinalSwap(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("scripted_overwrite", []connectors.Record{
		{"id": "a", "name": "Ada", "updated_at": "2026-01-01T00:00:00Z"},
		{"id": "g", "name": "Grace", "updated_at": "2026-01-02T00:00:00Z"},
	})
	a, connection := setupSyncModeApp(t, source, "full_refresh_overwrite")

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatal(err)
	}
	source.records = []connectors.Record{{"id": "k", "name": "Katherine", "updated_at": "2026-01-03T00:00:00Z"}}
	source.failAfter = 1
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err == nil {
		t.Fatal("RunETL(failing overwrite) error = nil")
	}
	rows, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if got := sortedNames(rows); got != "Ada,Grace" {
		t.Fatalf("rows after failed overwrite = %s, want Ada,Grace", got)
	}

	source.failAfter = -1
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatal(err)
	}
	rows, err = a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if got := sortedNames(rows); got != "Katherine" {
		t.Fatalf("rows after successful overwrite = %s, want Katherine", got)
	}
}

func TestIncrementalAppendCommitsCursorOnlyAfterSuccess(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("scripted_incremental", []connectors.Record{
		{"id": "a", "name": "Ada", "updated_at": "2026-01-01T00:00:00Z"},
		{"id": "g", "name": "Grace", "updated_at": "2026-01-02T00:00:00Z"},
	})
	a, connection := setupSyncModeApp(t, source, "incremental_append")

	run, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if run.Checkpoint["cursor"] != "2026-01-02T00:00:00Z" {
		t.Fatalf("cursor = %q", run.Checkpoint["cursor"])
	}

	source.records = []connectors.Record{
		{"id": "g", "name": "Grace resent", "updated_at": "2026-01-02T00:00:00Z"},
		{"id": "k", "name": "Katherine", "updated_at": "2026-01-03T00:00:00Z"},
	}
	run, err = a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if run.Checkpoint["cursor"] != "2026-01-03T00:00:00Z" {
		t.Fatalf("cursor after second run = %q", run.Checkpoint["cursor"])
	}
	rows, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 {
		t.Fatalf("rows len = %d, want 4 including inclusive resent cursor row", len(rows))
	}

	source.records = []connectors.Record{{"id": "m", "name": "Margaret", "updated_at": "2026-01-04T00:00:00Z"}}
	source.failAfter = 1
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err == nil {
		t.Fatal("RunETL(failing incremental) error = nil")
	}
	state := a.state.StreamStates[streamStateKey(connection, "records")]
	if state.Cursor != "2026-01-03T00:00:00Z" {
		t.Fatalf("cursor advanced after failed run = %q", state.Cursor)
	}
}

func TestIncrementalAppendDedupedMaterializesLatestRows(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("scripted_incremental_deduped", []connectors.Record{
		{"id": "a", "name": "Ada", "updated_at": "2026-01-01T00:00:00Z"},
		{"id": "g", "name": "Grace", "updated_at": "2026-01-02T00:00:00Z"},
	})
	a, connection := setupSyncModeApp(t, source, "incremental_append_deduped")

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 2}); err != nil {
		t.Fatal(err)
	}
	source.records = []connectors.Record{
		{"id": "a", "name": "Ada latest", "updated_at": "2026-01-03T00:00:00Z"},
		{"id": "g", "name": "Grace resent", "updated_at": "2026-01-02T00:00:00Z"},
		{"id": "k", "name": "Katherine", "updated_at": "2026-01-04T00:00:00Z"},
		{"id": "m", "name": "Margaret deleted", "updated_at": "2026-01-05T00:00:00Z", "_polymetrics_deleted": true},
	}
	run, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if run.Checkpoint["cursor"] != "2026-01-05T00:00:00Z" {
		t.Fatalf("cursor = %q", run.Checkpoint["cursor"])
	}
	rows, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	byID := rowsByID(rows)
	if len(byID) != 3 {
		t.Fatalf("deduped rows len = %d, want 3: %+v", len(byID), rows)
	}
	if byID["a"]["name"] != "Ada latest" || byID["g"]["name"] != "Grace resent" || byID["k"]["name"] != "Katherine" {
		t.Fatalf("unexpected deduped rows: %+v", byID)
	}
	if _, ok := byID["m"]; ok {
		t.Fatalf("delete/tombstone row remained in final output: %+v", byID["m"])
	}
}

func TestFullRefreshOverwriteDedupedReplacesFinalWithCurrentGeneration(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("scripted_full_deduped", []connectors.Record{
		{"id": "a", "name": "Ada old", "updated_at": "2026-01-01T00:00:00Z"},
		{"id": "a", "name": "Ada latest", "updated_at": "2026-01-03T00:00:00Z"},
		{"id": "g", "name": "Grace", "updated_at": "2026-01-02T00:00:00Z"},
	})
	a, connection := setupSyncModeApp(t, source, "full_refresh_overwrite_deduped")

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 2}); err != nil {
		t.Fatal(err)
	}
	rows, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	byID := rowsByID(rows)
	if len(byID) != 2 || byID["a"]["name"] != "Ada latest" || byID["g"]["name"] != "Grace" {
		t.Fatalf("unexpected first deduped final: %+v", byID)
	}

	source.records = []connectors.Record{{"id": "k", "name": "Katherine", "updated_at": "2026-01-04T00:00:00Z"}}
	source.failAfter = 1
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err == nil {
		t.Fatal("RunETL(failing full refresh dedupe) error = nil")
	}
	rows, err = a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if got := sortedNames(rows); got != "Ada latest,Grace" {
		t.Fatalf("rows after failed full refresh dedupe = %s", got)
	}

	source.failAfter = -1
	source.records = []connectors.Record{
		{"id": "a", "name": "Ada current", "updated_at": "2026-01-05T00:00:00Z"},
		{"id": "a", "name": "Ada stale", "updated_at": "2026-01-04T00:00:00Z"},
	}
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatal(err)
	}
	rows, err = a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	byID = rowsByID(rows)
	if len(byID) != 1 || byID["a"]["name"] != "Ada current" {
		t.Fatalf("unexpected current generation final: %+v", byID)
	}
}

func sortedNames(rows []connectors.Record) string {
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		if value, ok := row["name"].(string); ok {
			names = append(names, value)
		}
	}
	sort.Strings(names)
	return strings.Join(names, ",")
}
