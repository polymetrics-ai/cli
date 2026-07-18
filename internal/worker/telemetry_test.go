package worker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	temporalworker "go.temporal.io/sdk/worker"

	"polymetrics.ai/internal/telemetry"
)

func TestTemporalClientOptionsTelemetryGated(t *testing.T) {
	disabled := temporalClientOptions(context.Background(), "localhost:7233", testTemporalLogger())
	assertTemporalTelemetryDisabled(t, disabled)

	root := t.TempDir()
	ctx, handle := telemetry.Init(context.Background(), telemetry.Config{Exporter: telemetry.ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "temporal-options"}, func(string) {})
	defer telemetry.Shutdown(context.Background(), handle, func(string) {})
	enabled := temporalClientOptions(ctx, "localhost:7233", testTemporalLogger())
	if len(enabled.Interceptors) == 0 {
		t.Fatal("telemetry-enabled Temporal options missing tracing interceptor")
	}
	if enabled.MetricsHandler == nil {
		t.Fatal("telemetry-enabled Temporal options missing metrics handler")
	}
}

func TestTemporalWorkerOptionsTelemetryGated(t *testing.T) {
	disabled := temporalWorkerOptions(context.Background(), testTemporalLogger())
	assertTemporalWorkerTelemetryDisabled(t, disabled)

	root := t.TempDir()
	ctx, handle := telemetry.Init(context.Background(), telemetry.Config{Exporter: telemetry.ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "temporal-worker-options"}, func(string) {})
	defer telemetry.Shutdown(context.Background(), handle, func(string) {})
	enabled := temporalWorkerOptions(ctx, testTemporalLogger())
	if len(enabled.Interceptors) == 0 {
		t.Fatal("telemetry-enabled Temporal worker options missing tracing interceptor")
	}
}

func TestTemporalWorkerConstructorsUseTelemetryOptions(t *testing.T) {
	for _, path := range []string{"serve.go", "submit.go"} {
		t.Run(path, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			src := string(data)
			if strings.Contains(src, "worker.Options{}") {
				t.Fatalf("%s still constructs Temporal worker with empty worker.Options{}", path)
			}
			if !strings.Contains(src, "temporalWorkerOptions(") {
				t.Fatalf("%s does not wire temporalWorkerOptions into worker.New", path)
			}
		})
	}
}

func TestTemporalMetricsOnErrorWarnsWithoutPanic(t *testing.T) {
	root := t.TempDir()
	ctx, handle := telemetry.Init(context.Background(), telemetry.Config{Exporter: telemetry.ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "temporal-metrics-error"}, func(string) {})
	defer telemetry.Shutdown(context.Background(), handle, func(string) {})
	logger := &captureTemporalLogger{}
	opts := temporalClientOptions(ctx, "localhost:7233", logger)
	if opts.MetricsHandler == nil {
		t.Fatal("telemetry-enabled Temporal options missing metrics handler")
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Temporal metrics OnError panicked: %v", r)
			}
		}()
		opts.MetricsHandler.Counter("").Inc(1)
	}()

	joined := strings.Join(logger.warns, "\n")
	if !strings.Contains(joined, "temporal telemetry metrics handler error") {
		t.Fatalf("warnings %q missing Temporal metrics handler warning", joined)
	}
}

func testTemporalLogger() tlog.Logger {
	return tlog.NewStructuredLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

type captureTemporalLogger struct {
	warns []string
}

func (l *captureTemporalLogger) Debug(string, ...interface{}) {}
func (l *captureTemporalLogger) Info(string, ...interface{})  {}
func (l *captureTemporalLogger) Error(string, ...interface{}) {}
func (l *captureTemporalLogger) Warn(msg string, keyvals ...interface{}) {
	l.warns = append(l.warns, fmt.Sprint(append([]interface{}{msg}, keyvals...)...))
}

func assertTemporalTelemetryDisabled(t *testing.T, opts client.Options) {
	t.Helper()
	if len(opts.Interceptors) != 0 {
		t.Fatalf("disabled Temporal options have %d interceptors, want 0", len(opts.Interceptors))
	}
	if opts.MetricsHandler != nil {
		t.Fatal("disabled Temporal options have metrics handler, want nil")
	}
}

func assertTemporalWorkerTelemetryDisabled(t *testing.T, opts temporalworker.Options) {
	t.Helper()
	if len(opts.Interceptors) != 0 {
		t.Fatalf("disabled Temporal worker options have %d interceptors, want 0", len(opts.Interceptors))
	}
}
