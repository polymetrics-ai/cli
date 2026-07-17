package flow

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/telemetry"
)

func TestEngineRunEmitsFlowAndStepTelemetrySpans(t *testing.T) {
	dir := t.TempDir()
	ctx, handle := telemetry.Init(context.Background(), telemetry.Config{Exporter: telemetry.ExporterFile, Directory: dir, RunID: "flow-span"}, func(string) {})
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
	assertFlowTelemetryContains(t, data, "sync")
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
