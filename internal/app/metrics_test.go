package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/telemetry"
)

func TestRunETLMetricsAccumulateAndFlushByBatch(t *testing.T) {
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

	ctx := context.Background()
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

	dir := filepath.Join(root, ".polymetrics", "telemetry")
	ctx, handle := telemetry.Init(ctx, telemetry.Config{Exporter: telemetry.ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "etl-metrics"}, func(string) {})
	run, err := a.RunETL(ctx, RunETLRequest{Connection: "source_to_dest", Stream: "records", BatchSize: 2})
	if err != nil {
		t.Fatalf("RunETL: %v", err)
	}
	telemetry.Shutdown(context.Background(), handle, func(string) {})

	data := readMetricTelemetry(t, dir)
	assertMetricSum(t, data, "pm.records.read", run.RecordsRead)
	assertMetricSum(t, data, "pm.records.transformed", run.RecordsTransformed)
	assertMetricSum(t, data, "pm.records.loaded", run.RecordsLoaded)
	assertMetricSum(t, data, "pm.records.failed", run.RecordsFailed)
	assertMetricSum(t, data, "pm.batches.flushed", run.BatchCount)
	if got := intMetricSum(t, data, "pm.batches.flushed"); got != 3 {
		t.Fatalf("pm.batches.flushed = %d, want 3 flushes for 5 records with batch size 2", got)
	}
}

func TestRunCounterHotPathAllocations(t *testing.T) {
	counters := telemetry.NewRunCounters(context.Background())
	allocs := testing.AllocsPerRun(1000, func() {
		counters.RecordRead()
		counters.RecordTransformed()
		counters.RecordLoaded(1)
	})
	if allocs != 0 {
		t.Fatalf("hot-path counter increments allocated %.2f times per run, want 0", allocs)
	}
}

func BenchmarkEmit(b *testing.B) {
	counters := telemetry.NewRunCounters(context.Background())
	b.ReportAllocs()
	for b.Loop() {
		counters.RecordRead()
		counters.RecordTransformed()
		counters.RecordLoaded(1)
	}
}

func readMetricTelemetry(t *testing.T, dir string) []byte {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read telemetry dir: %v", err)
	}
	var out bytes.Buffer
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("read telemetry file: %v", err)
		}
		out.Write(data)
	}
	if out.Len() == 0 {
		t.Fatalf("no telemetry JSONL data under %s", dir)
	}
	return out.Bytes()
}

func assertMetricSum(t *testing.T, data []byte, name string, want int) {
	t.Helper()
	if got := intMetricSum(t, data, name); got != want {
		t.Fatalf("metric %s sum = %d, want %d\ntelemetry:\n%s", name, got, want, data)
	}
}

func intMetricSum(t *testing.T, data []byte, name string) int {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	total := 0
	for {
		var obj map[string]any
		if err := dec.Decode(&obj); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatalf("decode telemetry JSONL: %v", err)
		}
		scopeMetrics, _ := obj["ScopeMetrics"].([]any)
		for _, scopeMetric := range scopeMetrics {
			sm, _ := scopeMetric.(map[string]any)
			metrics, _ := sm["Metrics"].([]any)
			for _, metric := range metrics {
				m, _ := metric.(map[string]any)
				if m["Name"] != name {
					continue
				}
				dataObj, _ := m["Data"].(map[string]any)
				points, _ := dataObj["DataPoints"].([]any)
				for _, point := range points {
					dp, _ := point.(map[string]any)
					total += int(dp["Value"].(float64))
				}
			}
		}
	}
	return total
}
