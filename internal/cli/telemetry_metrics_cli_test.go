package cli_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

const syntheticMetricCanary = "pm-test-cli-metric-redaction-value-415"

func TestTelemetryFileMetricsReconcileWithETLRunEnvelope(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PM_TELEMETRY", "file")
	t.Setenv("PM_SYNTHETIC_METRIC_CANARY", syntheticMetricCanary)

	runCLIForLogSmoke(t, root, "init", "--json")
	runCLIForLogSmoke(t, root, "credentials", "add", "sample-cred", "--connector", "sample", "--from-env", "token=PM_SYNTHETIC_METRIC_CANARY", "--json")
	runCLIForLogSmoke(t, root, "credentials", "add", "warehouse-local", "--connector", "warehouse", "--config", "path=.polymetrics/warehouse", "--json")
	runCLIForLogSmoke(t, root, "connections", "create", "sample-to-warehouse", "--source", "sample:sample-cred", "--destination", "warehouse:warehouse-local", "--stream", "customers", "--sync-mode", "full_refresh_overwrite", "--table", "customers", "--json")

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"--root", root, "etl", "run", "--connection", "sample-to-warehouse", "--stream", "customers", "--batch-size", "2", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("etl run exit %d; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	env := decodeETLRunEnvelope(t, stdout.Bytes())
	data := readCLITelemetry(t, filepath.Join(root, ".polymetrics", "telemetry"))
	assertCLIMetricSum(t, data, "pm.records.read", env.Run.RecordsRead)
	assertCLIMetricSum(t, data, "pm.records.transformed", env.Run.RecordsTransformed)
	assertCLIMetricSum(t, data, "pm.records.loaded", env.Run.RecordsLoaded)
	assertCLIMetricSum(t, data, "pm.records.failed", env.Run.RecordsFailed)
	assertCLIMetricSum(t, data, "pm.batches.flushed", env.Run.BatchCount)
	for _, forbidden := range []string{syntheticMetricCanary, "token=", "Authorization", "request.body", "http.request.header", "url.full", "argv"} {
		assertCLINotContains(t, data, forbidden)
	}
}

type etlRunEnvelope struct {
	Kind string `json:"kind"`
	Run  struct {
		RecordsRead        int `json:"records_read"`
		RecordsTransformed int `json:"records_transformed"`
		RecordsLoaded      int `json:"records_loaded"`
		RecordsFailed      int `json:"records_failed"`
		BatchCount         int `json:"batch_count"`
	} `json:"run"`
}

func decodeETLRunEnvelope(t *testing.T, data []byte) etlRunEnvelope {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	var env etlRunEnvelope
	if err := dec.Decode(&env); err != nil {
		t.Fatalf("decode ETLRun envelope: %v", err)
	}
	if env.Kind != "ETLRun" {
		t.Fatalf("kind = %q, want ETLRun", env.Kind)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		t.Fatalf("stdout has extra JSON/text after first envelope")
	}
	return env
}

func assertCLIMetricSum(t *testing.T, data []byte, name string, want int) {
	t.Helper()
	if got := cliMetricSum(t, data, name); got != want {
		t.Fatalf("metric %s sum = %d, want %d\ntelemetry:\n%s", name, got, want, data)
	}
}

func cliMetricSum(t *testing.T, data []byte, name string) int {
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

func TestTelemetryRejectsUnsafeAmbientOTLPMetricsEndpointBeforeExporter(t *testing.T) {
	const marker = "pm_ambient_metrics_endpoint_marker"
	root := t.TempDir()
	t.Setenv("PM_TELEMETRY", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "https://user:"+marker+"@collector.example.test/v1/metrics?token="+marker)
	var stdout, stderr bytes.Buffer

	code := cli.Run([]string{"--root", root, "version", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"kind": "Version"`) {
		t.Fatalf("stdout missing Version envelope: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "warning: telemetry:") || !strings.Contains(stderr.String(), "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT") {
		t.Fatalf("stderr missing redacted OTLP metrics endpoint warning: %q", stderr.String())
	}
	if strings.Contains(stderr.String(), marker) || strings.Contains(stderr.String(), "token=") || strings.Contains(stderr.String(), "user:") {
		t.Fatalf("stderr leaked metrics endpoint detail: %q", stderr.String())
	}
}
