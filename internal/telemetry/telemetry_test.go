package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pmlogging "polymetrics.ai/internal/logging"
)

func TestInitDisabledCreatesNoSDKOrTelemetryDir(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics", "telemetry")
	var warnings []string

	ctx, handle := Init(context.Background(), Config{Exporter: ExporterNone, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry")}, func(msg string) {
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
	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics", "telemetry")
	var warnings []string

	ctx, handle := Init(context.Background(), Config{Exporter: ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "test-run"}, func(msg string) {
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
	assertTelemetryPathPerms(t, dir, filepath.Join(dir, "test-run.jsonl"))
	data := readTelemetryDir(t, dir)
	assertContains(t, data, "pm.command")
	assertContains(t, data, "pm.command.name")
	assertContains(t, data, "pm.http.path")
	assertNotContains(t, data, marker)
	assertNotContains(t, data, "url.full")
	assertNotContains(t, data, "request.body")
	assertNotContains(t, data, "token=")
}

func TestFileExporterRejectsUnsafePaths(t *testing.T) {
	root := t.TempDir()
	for _, tt := range []struct {
		name string
		dir  string
	}{
		{name: "absolute", dir: filepath.Join(root, "absolute-telemetry")},
		{name: "escape", dir: filepath.Join("..", "escape-telemetry")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var warnings []string
			_, handle := Init(context.Background(), Config{Exporter: ExporterFile, ProjectRoot: root, Directory: tt.dir, RunID: tt.name}, func(msg string) {
				warnings = append(warnings, msg)
			})
			if handle.Enabled() {
				t.Fatalf("unsafe directory %q Enabled() = true, want false", tt.dir)
			}
			if len(warnings) == 0 {
				t.Fatal("warnings empty, want unsafe path warning")
			}
		})
	}
}

func TestFileExporterRejectsSymlinkedDirectoryAndFile(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	polyDir := filepath.Join(root, ".polymetrics")
	if err := os.Mkdir(polyDir, 0o700); err != nil {
		t.Fatalf("mkdir .polymetrics: %v", err)
	}
	symlinkDir := filepath.Join(polyDir, "telemetry")
	if err := os.Symlink(target, symlinkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	var warnings []string
	_, handle := Init(context.Background(), Config{Exporter: ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "symlink-dir"}, func(msg string) {
		warnings = append(warnings, msg)
	})
	if handle.Enabled() {
		t.Fatal("symlinked telemetry directory Enabled() = true, want false")
	}
	if len(warnings) == 0 {
		t.Fatal("warnings empty, want symlink directory warning")
	}

	root = t.TempDir()
	dir := filepath.Join(root, ".polymetrics", "telemetry")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir telemetry dir: %v", err)
	}
	targetFile := filepath.Join(root, "target.jsonl")
	if err := os.WriteFile(targetFile, []byte("target"), 0o600); err != nil {
		t.Fatalf("write target file: %v", err)
	}
	if err := os.Symlink(targetFile, filepath.Join(dir, "symlink-file.jsonl")); err != nil {
		t.Skipf("symlink file unavailable: %v", err)
	}
	warnings = nil
	_, handle = Init(context.Background(), Config{Exporter: ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "symlink-file"}, func(msg string) {
		warnings = append(warnings, msg)
	})
	if handle.Enabled() {
		t.Fatal("symlinked telemetry file Enabled() = true, want false")
	}
	if len(warnings) == 0 {
		t.Fatal("warnings empty, want symlink file warning")
	}
}

func TestRecordErrorEmitsOnlySafeMetadata(t *testing.T) {
	const marker = "pm_registered_marker_command_failure"
	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics", "telemetry")
	registry := pmlogging.NewValueRegistry()
	registry.Register(marker)
	ctx := pmlogging.WithRegistry(context.Background(), registry)
	ctx, handle := Init(ctx, Config{Exporter: ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "record-error"}, func(string) {})

	_, span := StartSpan(ctx, "pm.command", StringAttr("pm.command.name", "help"))
	span.RecordError(fmt.Errorf("command failed with %s and response body contained api_key=%s", marker, marker))
	span.End()
	Shutdown(context.Background(), handle, func(string) {})

	data := readTelemetryDir(t, dir)
	assertContains(t, data, "pm.error.type")
	assertNotContains(t, data, "exception.")
	assertNotContains(t, data, marker)
	assertNotContains(t, data, "response body")
	assertNotContains(t, data, "api_key=")
}

func TestRecordErrorMinimalCaptureSuppressesMessages(t *testing.T) {
	const marker = "pm_registered_marker_minimal_failure"
	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics", "telemetry")
	registry := pmlogging.NewValueRegistry()
	registry.Register(marker)
	ctx := pmlogging.WithRegistry(context.Background(), registry)
	ctx, handle := Init(ctx, Config{Exporter: ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), Capture: captureMinimal, RunID: "record-error-minimal"}, func(string) {})

	_, span := StartSpan(ctx, "pm.command", StringAttr("pm.command.name", "help"))
	span.RecordError(fmt.Errorf("minimal capture must not export %s", marker))
	span.End()
	Shutdown(context.Background(), handle, func(string) {})

	data := readTelemetryDir(t, dir)
	assertNotContains(t, data, "exception.")
	assertNotContains(t, data, marker)
	assertNotContains(t, data, "minimal capture must not export")
}

func TestAddEventKeepsAllowlistedAttrsOnEvent(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics", "telemetry")
	ctx, handle := Init(context.Background(), Config{Exporter: ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry"), RunID: "event-attrs"}, func(string) {})

	_, span := StartSpan(ctx, "pm.connector.http", StringAttr("pm.http.method", "GET"))
	span.AddEvent("pm.connector.http.retry", IntAttr("pm.http.attempt", 2), IntAttr("pm.http.status_code", 503), BoolAttr("pm.http.retry", true), StringAttr("request.body", "forbidden"))
	span.End()
	Shutdown(context.Background(), handle, func(string) {})

	data := readTelemetryDir(t, dir)
	if !spanEventHasAttr(t, data, "pm.connector.http.retry", "pm.http.attempt") {
		t.Fatalf("retry event missing pm.http.attempt attr:\n%s", data)
	}
	if !spanEventHasAttr(t, data, "pm.connector.http.retry", "pm.http.status_code") {
		t.Fatalf("retry event missing pm.http.status_code attr:\n%s", data)
	}
	assertNotContains(t, data, "request.body")
	assertNotContains(t, data, "forbidden")
}

func TestOTLPEndpointValidationRejectsUnsafeURLs(t *testing.T) {
	const marker = "pm_endpoint_marker"
	for _, endpoint := range []string{
		"https://user:" + marker + "@collector.example.test/v1/traces",
		"https://collector.example.test/v1/traces?token=" + marker,
		"https://collector.example.test/v1/traces#" + marker,
		"ftp://collector.example.test/v1/traces",
	} {
		t.Run(endpoint, func(t *testing.T) {
			var warnings []string
			_, handle := Init(context.Background(), Config{Exporter: ExporterOTLP, Endpoint: endpoint}, func(msg string) {
				warnings = append(warnings, msg)
			})
			if handle.Enabled() {
				t.Fatal("unsafe OTLP endpoint Enabled() = true, want false")
			}
			joined := strings.Join(warnings, "\n")
			if !strings.Contains(joined, "invalid OTLP endpoint") {
				t.Fatalf("warnings = %q, want invalid OTLP endpoint", joined)
			}
			if strings.Contains(joined, marker) || strings.Contains(joined, "token=") || strings.Contains(joined, "user:") {
				t.Fatalf("warning leaked endpoint detail: %q", joined)
			}
		})
	}
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
	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics", "telemetry")
	var warnings []string

	_, handle := Init(context.Background(), Config{Exporter: ExporterFile, ProjectRoot: root, Directory: filepath.Join(".polymetrics", "telemetry")}, func(msg string) {
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

func spanEventHasAttr(t *testing.T, data []byte, eventName, attrKey string) bool {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		var span map[string]any
		err := dec.Decode(&span)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("decode span JSON: %v\n%s", err, data)
		}
		events, _ := span["Events"].([]any)
		for _, rawEvent := range events {
			event, _ := rawEvent.(map[string]any)
			if event["Name"] != eventName {
				continue
			}
			attrs, _ := event["Attributes"].([]any)
			for _, rawAttr := range attrs {
				attr, _ := rawAttr.(map[string]any)
				if attr["Key"] == attrKey {
					return true
				}
			}
		}
	}
	return false
}

func assertJoinedContains(t *testing.T, warnings []string, needle string) {
	t.Helper()
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, needle) {
		t.Fatalf("warnings %q missing %q", joined, needle)
	}
}

func assertTelemetryPathPerms(t *testing.T, dir, file string) {
	t.Helper()
	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat telemetry dir: %v", err)
	}
	if mode := dirInfo.Mode().Perm(); mode != 0o700 {
		t.Fatalf("telemetry dir mode = %v, want 0700", mode)
	}
	fileInfo, err := os.Stat(file)
	if err != nil {
		t.Fatalf("stat telemetry file: %v", err)
	}
	if mode := fileInfo.Mode().Perm(); mode != 0o600 {
		t.Fatalf("telemetry file mode = %v, want 0600", mode)
	}
}
