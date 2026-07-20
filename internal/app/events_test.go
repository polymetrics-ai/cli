package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/events"
	pmlogging "polymetrics.ai/internal/logging"
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

func TestRunETLEmitsFailedTerminalEvent(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	source := &streamingSource{total: 1}
	dest := &failingDestination{}
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
	_, err = a.RunETL(events.WithEmitter(ctx, collector), RunETLRequest{Connection: "source_to_dest", Stream: "records", BatchSize: 1})
	if err == nil {
		t.Fatal("RunETL() error = nil, want destination failure")
	}

	got := appEventSequence(collector.Events())
	want := []string{
		"etl:records:started:running",
		"etl:records:failed:failed",
	}
	assertAppEventSequence(t, got, want)
}

func TestRunETLFailureRedactsRegisteredValuesAcrossStateEventsAndLogs(t *testing.T) {
	const marker = "pm-test-app-redaction-marker-404"
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	source := &streamingSource{total: 1}
	dest := &secretFailingDestination{}
	registry := connectors.NewRegistry()
	registry.Register(source)
	registry.Register(dest)
	a.registry = registry

	valueRegistry := pmlogging.NewValueRegistry()
	var stderr bytes.Buffer
	logger, closeLogs := pmlogging.NewLogger(filepath.Join(root, ".polymetrics"), &stderr, pmlogging.LoggerOptions{Registry: valueRegistry})
	defer func() { _ = closeLogs() }()
	collector := events.NewCollector()
	ctx := pmlogging.WithRegistry(context.Background(), valueRegistry)
	ctx = pmlogging.WithLogger(ctx, logger)
	ctx = events.WithEmitter(ctx, collector)

	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "dest", Connector: dest.Name(), Secrets: map[string]string{"token": marker}, Config: map[string]string{"path": filepath.Join(root, "out")}}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "source_to_secret_failing_dest",
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: dest.Name(), Credential: "dest"},
		Streams: map[string]StreamConfig{
			"records": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, DestinationTable: "records"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	run, err := a.RunETL(ctx, RunETLRequest{Connection: "source_to_secret_failing_dest", Stream: "records", BatchSize: 1})
	if err == nil {
		t.Fatal("RunETL() error = nil, want destination failure")
	}
	if !bytes.Contains([]byte("dirty "+marker), []byte(marker)) {
		t.Fatalf("synthetic scanner failed to detect dirty fixture")
	}
	assertNoMarker(t, marker, "run error", []byte(run.Error))
	stateBytes, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	assertNoMarker(t, marker, "state", stateBytes)
	eventBytes, err := json.Marshal(collector.Events())
	if err != nil {
		t.Fatalf("marshal events: %v", err)
	}
	assertNoMarker(t, marker, "events", eventBytes)
	logBytes := readAllRunLogs(t, filepath.Join(root, ".polymetrics", "logs"))
	assertNoMarker(t, marker, "logs", logBytes)
	assertNoMarker(t, marker, "stderr", stderr.Bytes())
}

type secretFailingDestination struct{}

func (d *secretFailingDestination) Name() string { return "secret_failing_destination" }

func (d *secretFailingDestination) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:         d.Name(),
		DisplayName:  "Secret Failing Destination",
		Description:  "Test destination that fails with a registered value in the error.",
		Capabilities: connectors.Capabilities{Check: true, Catalog: true, Write: true},
	}
}

func (d *secretFailingDestination) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return ctx.Err()
}

func (d *secretFailingDestination) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: d.Name()}, nil
}

func (d *secretFailingDestination) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	return connectors.ErrUnsupportedOperation
}

func (d *secretFailingDestination) Write(_ context.Context, req connectors.WriteRequest, _ []connectors.Record) (connectors.WriteResult, error) {
	secret := req.Config.Secrets["token"]
	return connectors.WriteResult{}, fmt.Errorf("remote write rejected %s\nfor https://api.example.test/items?token=%s", secret, secret)
}

type failingDestination struct{}

func (d *failingDestination) Name() string { return "failing_destination" }

func (d *failingDestination) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:         d.Name(),
		DisplayName:  "Failing Destination",
		Description:  "Test destination that fails writes.",
		Capabilities: connectors.Capabilities{Check: true, Catalog: true, Write: true},
	}
}

func (d *failingDestination) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return ctx.Err()
}

func (d *failingDestination) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: d.Name()}, nil
}

func (d *failingDestination) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	return connectors.ErrUnsupportedOperation
}

func (d *failingDestination) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, errFailingDestination
}

var errFailingDestination = errString("destination failed")

type errString string

func (e errString) Error() string { return string(e) }

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

func readAllRunLogs(t *testing.T, dir string) []byte {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read log dir: %v", err)
	}
	var out bytes.Buffer
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("read log file: %v", err)
		}
		out.Write(data)
	}
	if out.Len() == 0 {
		t.Fatalf("expected run logs")
	}
	return out.Bytes()
}

func assertNoMarker(t *testing.T, marker, label string, data []byte) {
	t.Helper()
	if bytes.Contains(data, []byte(marker)) {
		t.Fatalf("%s contained synthetic marker", label)
	}
}
