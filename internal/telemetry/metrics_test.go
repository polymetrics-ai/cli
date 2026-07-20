package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestPRDMetricContract(t *testing.T) {
	const marker = "pm_metric_attribute_secret_marker"
	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics", "telemetry")
	ctx, handle := Init(context.Background(), Config{
		Exporter:    ExporterFile,
		ProjectRoot: root,
		Directory:   filepath.Join(".polymetrics", "telemetry"),
		RunID:       "prd-metric-contract",
	}, func(string) {})
	if !handle.Enabled() {
		t.Fatal("file telemetry disabled, want PRD metric contract")
	}

	counters := NewRunCounters(ctx)
	counters.RecordRead()
	counters.RecordTransformed()
	counters.RecordLoaded(1)
	counters.RecordFailed(1)
	counters.RecordBatchCreated()
	counters.RecordBatchRetried()
	counters.RecordBatchSkipped()
	counters.RecordBatchFlushed()
	counters.Flush(ctx)
	RecordAPICall(ctx, "GET", 7)
	RecordAPIRetry(ctx, "GET")
	RecordRateLimitWait(ctx, "GET", 5*time.Millisecond)
	RecordConnectorOperation(ctx, "GET", 10*time.Millisecond, 11)
	RecordStageDuration(ctx, "etl", 20*time.Millisecond)
	RecordAPICall(ctx, marker, 1)
	RecordStageDuration(ctx, marker, time.Millisecond)
	Shutdown(context.Background(), handle, func(string) {})

	data := readTelemetryDir(t, dir)
	if bytes.Contains(data, []byte(marker)) {
		t.Fatalf("metric attributes leaked unbounded marker: %s", data)
	}
	for _, bounded := range []string{"pm.operation", "OTHER", "pm.stage", "other"} {
		if !bytes.Contains(data, []byte(bounded)) {
			t.Errorf("metric output missing bounded attribute %q: %s", bounded, data)
		}
	}
	if got := jsonObjectCount(t, data); got != 1 {
		t.Fatalf("file metric snapshots = %d, want one cumulative shutdown snapshot", got)
	}
	got := exportedMetricKinds(t, data)
	want := map[string]string{
		"pm.records.read":                 "counter",
		"pm.records.transformed":          "counter",
		"pm.records.loaded":               "counter",
		"pm.records.failed":               "counter",
		"pm.batches.created":              "counter",
		"pm.batches.retried":              "counter",
		"pm.batches.skipped":              "counter",
		"pm.batches.flushed":              "counter",
		"pm.api.calls":                    "counter",
		"pm.api.retries":                  "counter",
		"pm.api.rate_limit_waits":         "counter",
		"pm.bytes.read":                   "counter",
		"pm.bytes.written":                "counter",
		"pm.connector.operation.duration": "histogram",
		"pm.api.rate_limit_wait.duration": "histogram",
		"pm.stage.duration":               "histogram",
	}
	for name, kind := range want {
		if got[name] != kind {
			t.Errorf("metric %s kind = %q, want %q; exported=%v", name, got[name], kind, got)
		}
	}
}

func TestRunCountersKeepHotPathLocalUntilFlush(t *testing.T) {
	root := t.TempDir()
	ctx, handle := Init(context.Background(), Config{Exporter: ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "counter-locality"}, func(string) {})
	defer Shutdown(context.Background(), handle, func(string) {})
	if !handle.Enabled() {
		t.Fatal("file telemetry disabled, want enabled counters")
	}

	counters := NewRunCounters(ctx)
	if !counters.enabled {
		t.Fatal("NewRunCounters returned disabled counters for enabled telemetry")
	}

	counters.RecordRead()
	counters.RecordTransformed()
	counters.RecordLoaded(2)
	counters.RecordFailed(1)
	counters.RecordBatchCreated()
	counters.RecordBatchRetried()
	counters.RecordBatchSkipped()
	counters.RecordBatchFlushed()
	if counters.flushedRead != 0 || counters.flushedTransformed != 0 || counters.flushedLoaded != 0 || counters.flushedFailed != 0 || counters.flushedBatchesCreated != 0 || counters.flushedBatchesRetried != 0 || counters.flushedBatchesSkipped != 0 || counters.flushedBatches != 0 {
		t.Fatalf("record methods updated flushed counters before batch flush: %+v", counters)
	}

	counters.Flush(ctx)
	if counters.flushedRead != counters.recordsRead || counters.flushedTransformed != counters.recordsTransformed || counters.flushedLoaded != counters.recordsLoaded || counters.flushedFailed != counters.recordsFailed || counters.flushedBatchesCreated != counters.batchesCreated || counters.flushedBatchesRetried != counters.batchesRetried || counters.flushedBatchesSkipped != counters.batchesSkipped || counters.flushedBatches != counters.batchesFlushed {
		t.Fatalf("flush did not reconcile local/flushed counters: %+v", counters)
	}
}

func TestRunCounterEnabledHotPathAllocations(t *testing.T) {
	root := t.TempDir()
	ctx, handle := Init(context.Background(), Config{Exporter: ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "counter-allocs"}, func(string) {})
	defer Shutdown(context.Background(), handle, func(string) {})
	counters := NewRunCounters(ctx)
	if !counters.enabled {
		t.Fatal("NewRunCounters returned disabled counters for enabled telemetry")
	}

	allocs := testing.AllocsPerRun(1000, func() {
		counters.RecordRead()
		counters.RecordTransformed()
		counters.RecordLoaded(1)
	})
	if allocs != 0 {
		t.Fatalf("enabled hot-path counter increments allocated %.2f times per run, want 0", allocs)
	}
}

func TestOTLPMetricsExportBeforeShutdown(t *testing.T) {
	requests := make(chan string, 8)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, handle := Init(context.Background(), Config{
		Exporter:             ExporterOTLP,
		Endpoint:             server.URL + "/collector",
		MetricExportInterval: 10 * time.Millisecond,
		ShutdownTimeout:      time.Second,
	}, func(string) {})
	if !handle.Enabled() {
		t.Fatal("OTLP telemetry disabled")
	}
	counters := NewRunCounters(ctx)
	counters.RecordRead()
	counters.Flush(ctx)

	select {
	case path := <-requests:
		if path != "/collector/v1/metrics" {
			t.Fatalf("live metrics path = %q, want /collector/v1/metrics", path)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("OTLP metrics were not exported before Shutdown")
	}
	Shutdown(context.Background(), handle, func(string) {})
}

func TestOTLPMetricEndpointPathSemantics(t *testing.T) {
	tests := []struct {
		name        string
		configEnv   string
		configPath  string
		metricsPath string
		wantPath    string
	}{
		{name: "generic endpoint appends signal path", configEnv: "OTEL_EXPORTER_OTLP_ENDPOINT", configPath: "/prefix", wantPath: "/prefix/v1/metrics"},
		{name: "trace endpoint rewrites signal path", configEnv: "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", configPath: "/prefix/v1/traces", wantPath: "/prefix/v1/metrics"},
		{name: "metrics endpoint remains exact", configEnv: "OTEL_EXPORTER_OTLP_ENDPOINT", configPath: "/ignored", metricsPath: "/custom", wantPath: "/custom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := make(chan string, 8)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests <- r.URL.Path
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			t.Setenv(tt.configEnv, server.URL+tt.configPath)
			if tt.metricsPath != "" {
				t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", server.URL+tt.metricsPath)
			} else {
				t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "")
			}

			ctx, handle := Init(context.Background(), Config{
				Exporter:             ExporterOTLP,
				Endpoint:             os.Getenv(tt.configEnv),
				MetricExportInterval: 10 * time.Millisecond,
				ShutdownTimeout:      time.Second,
			}, func(string) {})
			if !handle.Enabled() {
				t.Fatal("OTLP telemetry disabled")
			}
			counters := NewRunCounters(ctx)
			counters.RecordRead()
			counters.Flush(ctx)

			select {
			case path := <-requests:
				if path != tt.wantPath {
					t.Fatalf("metrics path = %q, want %q", path, tt.wantPath)
				}
			case <-time.After(2 * time.Second):
				t.Fatalf("no metrics request received at %s", tt.wantPath)
			}
			Shutdown(context.Background(), handle, func(string) {})
		})
	}
}

func TestDisabledMetricsStartsNoExporter(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, handle := Init(context.Background(), Config{
		Exporter:             ExporterNone,
		Endpoint:             server.URL,
		MetricExportInterval: time.Millisecond,
	}, func(string) {})
	NewRunCounters(ctx).RecordRead()
	time.Sleep(25 * time.Millisecond)
	Shutdown(context.Background(), handle, func(string) {})
	if got := requests.Load(); got != 0 {
		t.Fatalf("disabled telemetry exporter requests = %d, want 0", got)
	}
}

func jsonObjectCount(t *testing.T, data []byte) int {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	count := 0
	for {
		var obj map[string]any
		err := dec.Decode(&obj)
		if errors.Is(err, io.EOF) {
			return count
		}
		if err != nil {
			t.Fatalf("decode telemetry JSONL: %v", err)
		}
		count++
	}
}

func exportedMetricKinds(t *testing.T, data []byte) map[string]string {
	t.Helper()
	got := map[string]string{}
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		var obj map[string]any
		if err := dec.Decode(&obj); err != nil {
			if errors.Is(err, io.EOF) {
				return got
			}
			t.Fatalf("decode telemetry JSONL: %v", err)
		}
		scopeMetrics, _ := obj["ScopeMetrics"].([]any)
		for _, rawScope := range scopeMetrics {
			scope, _ := rawScope.(map[string]any)
			metrics, _ := scope["Metrics"].([]any)
			for _, rawMetric := range metrics {
				metric, _ := rawMetric.(map[string]any)
				name, _ := metric["Name"].(string)
				metricData, _ := metric["Data"].(map[string]any)
				points, _ := metricData["DataPoints"].([]any)
				if name == "" || len(points) == 0 {
					continue
				}
				point, _ := points[0].(map[string]any)
				switch {
				case point["Value"] != nil:
					got[name] = "counter"
				case point["BucketCounts"] != nil:
					got[name] = "histogram"
				default:
					got[name] = "unknown:" + strings.Join(sortedKeys(point), ",")
				}
			}
		}
	}
}

func sortedKeys(in map[string]any) []string {
	keys := make([]string, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}
	// The helper only supports deterministic failure output; production metric
	// attributes never flow through this test-only key list.
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}
