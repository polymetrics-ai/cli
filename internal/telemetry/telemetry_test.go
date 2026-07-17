package telemetry

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitDisabledCreatesNoSDKOrTelemetryDir(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics", "telemetry")
	var warnings []string

	ctx, handle := Init(context.Background(), Config{Exporter: ExporterNone, Directory: dir}, func(msg string) {
		warnings = append(warnings, msg)
	})

	if ctx == nil {
		t.Fatal("Init returned nil context")
	}
	if handle.Enabled() {
		t.Fatal("disabled telemetry handle Enabled() = true, want false")
	}
	if handle.provider != nil {
		t.Fatal("disabled telemetry constructed SDK provider")
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("telemetry dir stat err = %v, want not exist", err)
	}
}

func TestFileExporterWritesAllowlistedSecretSafeSpans(t *testing.T) {
	const marker = "pm_test_secret_token_file_exporter"
	dir := t.TempDir()
	var warnings []string

	ctx, handle := Init(context.Background(), Config{Exporter: ExporterFile, Directory: dir, RunID: "test-run"}, func(msg string) {
		warnings = append(warnings, msg)
	})
	if !handle.Enabled() {
		t.Fatalf("file telemetry disabled; warnings=%v", warnings)
	}

	ctx, span := StartSpan(ctx, "pm.command",
		StringAttr("pm.command.name", "version"),
		StringAttr("url.full", "https://api.example.test/v1/accounts?token="+marker),
		StringAttr("pm.http.path", "/v1/accounts"),
	)
	span.AddEvent("attempt", StringAttr("pm.http.attempt", "1"), StringAttr("request.body", marker))
	span.End()
	Shutdown(context.Background(), handle, func(msg string) { warnings = append(warnings, msg) })

	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}
	data := readTelemetryDir(t, dir)
	assertContains(t, data, "pm.command")
	assertContains(t, data, "pm.command.name")
	assertContains(t, data, "pm.http.path")
	assertNotContains(t, data, marker)
	assertNotContains(t, data, "url.full")
	assertNotContains(t, data, "request.body")
	assertNotContains(t, data, "token=")
}

func TestInitFileExporterFailureWarnsAndDisables(t *testing.T) {
	root := t.TempDir()
	notDir := filepath.Join(root, "not-dir")
	if err := os.WriteFile(notDir, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write not-dir: %v", err)
	}
	var warnings []string

	_, handle := Init(context.Background(), Config{Exporter: ExporterFile, Directory: filepath.Join(notDir, "telemetry")}, func(msg string) {
		warnings = append(warnings, msg)
	})

	if handle.Enabled() {
		t.Fatal("failed file exporter Enabled() = true, want false")
	}
	if len(warnings) == 0 {
		t.Fatal("warnings empty, want file exporter warning")
	}
	assertJoinedContains(t, warnings, "telemetry")
}

func TestShutdownFailureWarnsWithoutReturningError(t *testing.T) {
	boom := errors.New("boom")
	handle := &Handle{enabled: true, shutdown: func(context.Context) error { return boom }}
	var warnings []string

	Shutdown(context.Background(), handle, func(msg string) {
		warnings = append(warnings, msg)
	})

	if len(warnings) != 1 {
		t.Fatalf("warnings = %v, want one shutdown warning", warnings)
	}
	assertJoinedContains(t, warnings, "shutdown")
	assertJoinedContains(t, warnings, "boom")
}

func TestOTELSDKDisabledOverridesFileExporter(t *testing.T) {
	t.Setenv("OTEL_SDK_DISABLED", "true")
	dir := filepath.Join(t.TempDir(), "telemetry")
	var warnings []string

	_, handle := Init(context.Background(), Config{Exporter: ExporterFile, Directory: dir}, func(msg string) {
		warnings = append(warnings, msg)
	})

	if handle.Enabled() {
		t.Fatal("OTEL_SDK_DISABLED=true handle Enabled() = true, want false")
	}
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("telemetry dir stat err = %v, want not exist", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}
}

func readTelemetryDir(t *testing.T, dir string) []byte {
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
			t.Fatalf("read telemetry file %s: %v", entry.Name(), err)
		}
		out.Write(data)
	}
	if out.Len() == 0 {
		t.Fatalf("no telemetry JSONL data under %s", dir)
	}
	return out.Bytes()
}

func assertContains(t *testing.T, data []byte, needle string) {
	t.Helper()
	if !bytes.Contains(data, []byte(needle)) {
		t.Fatalf("telemetry output missing %q:\n%s", needle, data)
	}
}

func assertNotContains(t *testing.T, data []byte, needle string) {
	t.Helper()
	if bytes.Contains(data, []byte(needle)) {
		t.Fatalf("telemetry output contains forbidden %q:\n%s", needle, data)
	}
}

func assertJoinedContains(t *testing.T, warnings []string, needle string) {
	t.Helper()
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, needle) {
		t.Fatalf("warnings %q missing %q", joined, needle)
	}
}
