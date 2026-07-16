package app

import (
	"context"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/events"
)

func TestRunETLEmitsConnectorFlushEvents(t *testing.T) {
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

	collector := events.NewCollector()
	run, err := a.RunETL(events.WithEmitter(ctx, collector), RunETLRequest{Connection: "source_to_dest", Stream: "records", BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if run.ID == "" {
		t.Fatal("run ID is empty")
	}

	got := appEventSequence(collector.Events())
	want := []string{
		"etl:records:started:running",
		"etl:records:progress:batch",
		"etl:records:progress:batch",
		"etl:records:progress:batch",
		"etl:records:completed:success",
	}
	assertAppEventSequence(t, got, want)
	assertBatchCounters(t, collector.Events(), []int64{2, 4, 5})
}

func TestRunWarehouseETLEmitsFlushEvents(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	source := &streamingSource{total: 3}
	a.registry.Register(source)

	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "warehouse", Connector: "warehouse", Config: map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")}}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "source_to_warehouse",
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
		Streams: map[string]StreamConfig{
			"records": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, DestinationTable: "records"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	collector := events.NewCollector()
	_, err = a.RunETL(events.WithEmitter(ctx, collector), RunETLRequest{Connection: "source_to_warehouse", Stream: "records", BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}

	got := appEventSequence(collector.Events())
	want := []string{
		"etl:records:started:running",
		"etl:records:progress:batch",
		"etl:records:progress:batch",
		"etl:records:completed:success",
	}
	assertAppEventSequence(t, got, want)
	assertBatchCounters(t, collector.Events(), []int64{2, 3})
}

func appEventSequence(in []events.Event) []string {
	out := make([]string, 0, len(in))
	for _, ev := range in {
		out = append(out, string(ev.Scope)+":"+ev.StepID+":"+string(ev.Kind)+":"+ev.Status)
	}
	return out
}

func assertAppEventSequence(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("sequence length = %d, want %d\ngot  %#v\nwant %#v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sequence[%d] = %q, want %q\ngot  %#v\nwant %#v", i, got[i], want[i], got, want)
		}
	}
}

func assertBatchCounters(t *testing.T, in []events.Event, loaded []int64) {
	t.Helper()
	var got []int64
	for _, ev := range in {
		if ev.Kind == events.KindProgress {
			got = append(got, ev.Counters.RecordsWritten)
		}
	}
	if len(got) != len(loaded) {
		t.Fatalf("progress counter count = %d, want %d: %v", len(got), len(loaded), got)
	}
	for i := range loaded {
		if got[i] != loaded[i] {
			t.Fatalf("progress[%d].RecordsWritten = %d, want %d (all=%v)", i, got[i], loaded[i], got)
		}
	}
}
