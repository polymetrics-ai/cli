package app

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
)

type streamingSource struct {
	total    int
	requests []connectors.ReadRequest
}

func (s *streamingSource) Name() string { return "streaming_source" }

func (s *streamingSource) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:         s.Name(),
		DisplayName:  "Streaming Source",
		Description:  "Test streaming source.",
		Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

func (s *streamingSource) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return ctx.Err()
}

func (s *streamingSource) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: s.Name(), Streams: []connectors.Stream{{Name: "records", PrimaryKey: []string{"id"}}}}, nil
}

func (s *streamingSource) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	s.requests = append(s.requests, req)
	for i := 0; i < s.total; i++ {
		if err := emit(connectors.Record{"id": i}); err != nil {
			return err
		}
	}
	return nil
}

func (s *streamingSource) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type batchDestination struct {
	batches []int
}

func (d *batchDestination) Name() string { return "batch_destination" }

func (d *batchDestination) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:         d.Name(),
		DisplayName:  "Batch Destination",
		Description:  "Test batch destination.",
		Capabilities: connectors.Capabilities{Check: true, Catalog: true, Write: true},
	}
}

func (d *batchDestination) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return ctx.Err()
}

func (d *batchDestination) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: d.Name()}, nil
}

func (d *batchDestination) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	return connectors.ErrUnsupportedOperation
}

func (d *batchDestination) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	d.batches = append(d.batches, len(records))
	return connectors.WriteResult{RecordsWritten: len(records)}, nil
}

func setupStreamingETLApp(t *testing.T, source *streamingSource, dest *batchDestination) (*App, string) {
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
	registry.Register(dest)
	a.registry = registry

	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "dest", Connector: dest.Name(), Config: map[string]string{"path": filepath.Join(root, "out")}}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "source_to_dest",
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: dest.Name(), Credential: "dest"},
		Streams: map[string]StreamConfig{
			"records": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, DestinationTable: "records"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	return a, "source_to_dest"
}

func TestRunETLWritesBoundedBatches(t *testing.T) {
	ctx := context.Background()
	source := &streamingSource{total: 5}
	dest := &batchDestination{}
	a, connection := setupStreamingETLApp(t, source, dest)

	run, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := dest.batches, []int{2, 2, 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("destination batches = %v, want %v", got, want)
	}
	if run.RecordsRead != 5 || run.RecordsLoaded != 5 || run.BatchCount != 3 {
		t.Fatalf("unexpected run counts: %+v", run)
	}
	if run.Checkpoint["records_read"] != "5" || run.Checkpoint["batches"] != "3" {
		t.Fatalf("missing checkpoint metadata: %+v", run.Checkpoint)
	}
}

func TestRunETLLimitCapsConnectorDestinationRead(t *testing.T) {
	ctx := context.Background()
	source := &streamingSource{total: 5}
	dest := &batchDestination{}
	a, connection := setupStreamingETLApp(t, source, dest)

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
	if got, want := dest.batches, []int{2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("destination batches = %v, want %v", got, want)
	}
	if run.RecordsRead != 2 || run.RecordsLoaded != 2 || run.BatchCount != 1 {
		t.Fatalf("unexpected capped run counts: %+v", run)
	}
	if run.Checkpoint["records_read"] != "2" || run.Checkpoint["batches"] != "1" {
		t.Fatalf("missing capped checkpoint metadata: %+v", run.Checkpoint)
	}
}
