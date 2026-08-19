package app

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

type streamingSource struct {
	total int
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
	batches          []int
	acknowledgements int
}

type dynamicStreamingSource struct {
	catalogCalls  int
	resolvedReads int
}

func (s *dynamicStreamingSource) Name() string { return "dynamic_streaming_source" }

func (s *dynamicStreamingSource) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:         s.Name(),
		DisplayName:  "Dynamic Streaming Source",
		Description:  "Test source with provider-derived streams.",
		Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

func (s *dynamicStreamingSource) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return ctx.Err()
}

func (s *dynamicStreamingSource) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	s.catalogCalls++
	return connectors.Catalog{
		Connector: s.Name(),
		Streams:   []connectors.Stream{{Name: "records", PrimaryKey: []string{"id"}}},
		Discovery: &connectors.DiscoveryStatus{Complete: true, ExpiresAt: time.Now().UTC().Add(time.Hour)},
	}, nil
}

func (s *dynamicStreamingSource) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if req.Config.ResolvedCatalog == nil {
		return errors.New("dynamic reader did not receive its durable catalog")
	}
	s.resolvedReads++
	return emit(connectors.Record{"id": s.resolvedReads})
}

func (s *dynamicStreamingSource) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
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

func (d *batchDestination) AcknowledgeETLDurability(_ context.Context, _ string) (synccontract.DownstreamAcknowledgement, error) {
	d.acknowledgements++
	return synccontract.NewDurableDownstreamAcknowledgement(d.Name(), time.Now().UTC())
}

func TestRunETLWritesBoundedBatches(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	source := &streamingSource{total: 5}
	dest := &batchDestination{}
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

	run, err := a.RunETL(ctx, RunETLRequest{Connection: "source_to_dest", Stream: "records", BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := dest.batches, []int{2, 2, 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("destination batches = %v, want %v", got, want)
	}
	if dest.acknowledgements != 1 {
		t.Fatalf("durability acknowledgements = %d, want 1", dest.acknowledgements)
	}
	if run.RecordsRead != 5 || run.RecordsLoaded != 5 || run.BatchCount != 3 {
		t.Fatalf("unexpected run counts: %+v", run)
	}
	if run.Checkpoint["records_read"] != "5" || run.Checkpoint["batches"] != "3" {
		t.Fatalf("missing checkpoint metadata: %+v", run.Checkpoint)
	}
}

func TestRunETLReusesDurableAccountCatalogAcrossProcesses(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	source := &dynamicStreamingSource{}
	destination := &batchDestination{}
	registry := connectors.NewRegistry()
	registry.Register(source)
	registry.Register(destination)

	first, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	first.registry = registry
	if _, err := first.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()}); err != nil {
		t.Fatal(err)
	}
	if _, err := first.AddCredential(ctx, AddCredentialRequest{Name: "destination", Connector: destination.Name(), Config: map[string]string{"path": filepath.Join(root, "out")}}); err != nil {
		t.Fatal(err)
	}
	if _, err := first.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "dynamic_to_destination",
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: destination.Name(), Credential: "destination"},
		Streams: map[string]StreamConfig{
			"records": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, DestinationTable: "records"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if got := source.catalogCalls; got != 1 {
		t.Fatalf("initial discovery calls = %d, want 1", got)
	}
	if _, err := first.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "same_account_to_destination",
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: destination.Name(), Credential: "destination"},
		Streams: map[string]StreamConfig{
			"records": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, DestinationTable: "records_second_connection"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if got := source.catalogCalls; got != 1 {
		t.Fatalf("same account catalog discovered twice: %d calls", got)
	}
	if got := len(first.state.Catalogs); got != 1 {
		t.Fatalf("same account catalog references = %d, want 1", got)
	}
	if _, err := first.RefreshCatalog(ctx, "dynamic_to_destination"); err != nil {
		t.Fatalf("explicit RefreshCatalog: %v", err)
	}
	if got := source.catalogCalls; got != 2 {
		t.Fatalf("explicit refresh calls = %d, want 2", got)
	}

	second, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	second.registry = registry
	if _, err := second.RunETL(ctx, RunETLRequest{Connection: "dynamic_to_destination", Stream: "records"}); err != nil {
		t.Fatalf("RunETL with durable catalog: %v", err)
	}
	if got := source.catalogCalls; got != 2 {
		t.Fatalf("discovery reran after reopening project: %d calls", got)
	}
	if got := source.resolvedReads; got != 1 {
		t.Fatalf("reads supplied with durable catalog = %d, want 1", got)
	}

	reference := second.state.Catalogs[0]
	stored, err := second.catalogs.read(reference)
	if err != nil {
		t.Fatalf("read durable catalog: %v", err)
	}
	stored.Catalog.Discovery.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	if err := second.catalogs.write(reference, stored.Catalog, stored.UpdatedAt); err != nil {
		t.Fatalf("expire durable catalog: %v", err)
	}
	third, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	third.registry = registry
	if _, err := third.RunETL(ctx, RunETLRequest{Connection: "dynamic_to_destination", Stream: "records"}); !errors.Is(err, errCatalogStale) {
		t.Fatalf("RunETL with stale catalog error = %v, want stale catalog error", err)
	}
	if got := source.catalogCalls; got != 2 {
		t.Fatalf("stale catalog silently rediscovered: %d calls", got)
	}
}
