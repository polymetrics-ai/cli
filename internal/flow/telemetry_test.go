package flow

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/telemetry"
)

func TestEngineEmitsStageDurationMetricAndFlowSpans(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics", "telemetry")
	ctx, handle := telemetry.Init(context.Background(), telemetry.Config{Exporter: telemetry.ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "flow-span"}, func(string) {})
	engine := &Engine{
		Manifest: FlowManifest{
			Name: "nightly_test_flow",
			Steps: []FlowStep{
				{ID: "sync_accounts", Kind: KindSync, Connection: "source_to_dest", Streams: []string{"accounts"}},
			},
		},
		App:     telemetryFlowApp{},
		LockDir: t.TempDir(),
	}

	result, err := engine.Run(ctx, RunOptions{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Status != "ok" {
		t.Fatalf("Status = %q, want ok", result.Status)
	}
	telemetry.Shutdown(context.Background(), handle, func(string) {})

	data := readFlowTelemetry(t, dir)
	assertFlowTelemetryContains(t, data, "pm.flow.run")
	assertFlowTelemetryContains(t, data, "pm.flow.step")
	assertFlowTelemetryContains(t, data, "pm.flow.name")
	assertFlowTelemetryContains(t, data, "nightly_test_flow")
	assertFlowTelemetryContains(t, data, "pm.flow.step_id")
	assertFlowTelemetryContains(t, data, "sync_accounts")
	assertFlowTelemetryContains(t, data, "pm.flow.step_kind")
	assertFlowTelemetryContains(t, data, "pm.stage.duration")
	assertFlowTelemetryContains(t, data, "pm.stage")
	assertFlowTelemetryContains(t, data, "sync")
}

func TestEngineRunFailedStepTelemetryRedactsError(t *testing.T) {
	const marker = "pm_flow_response_body_marker"
	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics", "telemetry")
	ctx, handle := telemetry.Init(context.Background(), telemetry.Config{Exporter: telemetry.ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "flow-failed-span"}, func(string) {})
	engine := &Engine{
		Manifest: FlowManifest{
			Name: "failed_flow",
			Steps: []FlowStep{
				{ID: "sync_accounts", Kind: KindSync, Connection: "source_to_dest", Streams: []string{"accounts"}},
			},
		},
		App:     failingTelemetryFlowApp{marker: marker},
		LockDir: t.TempDir(),
	}

	_, err := engine.Run(ctx, RunOptions{})
	if err == nil {
		t.Fatal("Run error = nil, want failed step")
	}
	telemetry.Shutdown(context.Background(), handle, func(string) {})

	data := readFlowTelemetry(t, dir)
	assertFlowTelemetryContains(t, data, "pm.flow.step")
	assertFlowTelemetryContains(t, data, "pm.error.type")
	assertFlowTelemetryContains(t, data, "pm.error.code")
	assertFlowTelemetryNotContains(t, data, "exception.")
	assertFlowTelemetryNotContains(t, data, "errorString")
	assertFlowTelemetryNotContains(t, data, marker)
	assertFlowTelemetryNotContains(t, data, "response body")
}

type telemetryFlowApp struct{}

func (telemetryFlowApp) ETLRun(context.Context, string, []string) (ETLResult, error) {
	return ETLResult{RecordsRead: 1, RecordsWritten: 1}, nil
}

func (telemetryFlowApp) QuerySQL(context.Context, string, int) ([]map[string]any, error) {
	return nil, nil
}

func (telemetryFlowApp) RLMRun(context.Context, RLMRunRequest) (RLMResult, error) {
	return RLMResult{}, nil
}

type failingTelemetryFlowApp struct {
	marker string
}

func (f failingTelemetryFlowApp) ETLRun(context.Context, string, []string) (ETLResult, error) {
	return ETLResult{}, fmt.Errorf("upstream response body contained %s", f.marker)
}

func (f failingTelemetryFlowApp) QuerySQL(context.Context, string, int) ([]map[string]any, error) {
	return nil, nil
}

func (f failingTelemetryFlowApp) RLMRun(context.Context, RLMRunRequest) (RLMResult, error) {
	return RLMResult{}, nil
}

func readFlowTelemetry(t *testing.T, dir string) []byte {
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

func assertFlowTelemetryContains(t *testing.T, data []byte, needle string) {
	t.Helper()
	if !bytes.Contains(data, []byte(needle)) {
		t.Fatalf("telemetry output missing %q:\n%s", needle, data)
	}
}

func assertFlowTelemetryNotContains(t *testing.T, data []byte, needle string) {
	t.Helper()
	if bytes.Contains(data, []byte(needle)) {
		t.Fatalf("telemetry output contains forbidden %q:\n%s", needle, data)
	}
}
