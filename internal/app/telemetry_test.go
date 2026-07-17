package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/telemetry"
)

func TestRunETLEmitsTelemetrySpan(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	source := &streamingSource{total: 2}
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
	ctx, handle := telemetry.Init(ctx, telemetry.Config{Exporter: telemetry.ExporterFile, Directory: dir, RunID: "etl-span"}, func(string) {})
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: "source_to_dest", Stream: "records", BatchSize: 1}); err != nil {
		t.Fatalf("RunETL: %v", err)
	}
	telemetry.Shutdown(context.Background(), handle, func(string) {})

	data := readAppTelemetry(t, dir)
	assertAppTelemetryContains(t, data, "pm.etl.run")
	assertAppTelemetryContains(t, data, "pm.etl.connection")
	assertAppTelemetryContains(t, data, "source_to_dest")
	assertAppTelemetryContains(t, data, "pm.etl.stream")
	assertAppTelemetryContains(t, data, "records")
}

func readAppTelemetry(t *testing.T, dir string) []byte {
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

func assertAppTelemetryContains(t *testing.T, data []byte, needle string) {
	t.Helper()
	if !bytes.Contains(data, []byte(needle)) {
		t.Fatalf("telemetry output missing %q:\n%s", needle, data)
	}
}
