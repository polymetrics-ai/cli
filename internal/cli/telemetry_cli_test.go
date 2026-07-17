package cli_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

func TestTelemetryDisabledByDefaultCreatesNoDirectory(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer

	code := cli.Run([]string{"--root", root, "version", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".polymetrics", "telemetry")); !os.IsNotExist(err) {
		t.Fatalf("telemetry dir stat err = %v, want not exist", err)
	}
}

func TestTelemetryFileExporterCommandSpanAndEnvelopeOnlyStdout(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PM_TELEMETRY", "file")
	var stdout, stderr bytes.Buffer

	code := cli.Run([]string{"--root", root, "version", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"kind": "Version"`) {
		t.Fatalf("stdout missing Version envelope: %s", stdout.String())
	}
	if strings.Contains(stdout.String(), "pm.command") {
		t.Fatalf("stdout contains telemetry span data: %s", stdout.String())
	}
	data := readCLITelemetry(t, filepath.Join(root, ".polymetrics", "telemetry"))
	assertCLIContains(t, data, "pm.command")
	assertCLIContains(t, data, "pm.command.name")
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestTelemetryFileExporterInitFailureIsExitCodeNeutral(t *testing.T) {
	root := t.TempDir()
	notDir := filepath.Join(root, "not-dir")
	if err := os.WriteFile(notDir, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write not-dir: %v", err)
	}
	t.Setenv("PM_TELEMETRY", "file")
	t.Setenv("PM_TELEMETRY_DIR", filepath.Join(notDir, "telemetry"))
	var stdout, stderr bytes.Buffer

	code := cli.Run([]string{"--root", root, "version", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"kind": "Version"`) {
		t.Fatalf("stdout missing Version envelope: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "warning: telemetry:") {
		t.Fatalf("stderr missing telemetry warning: %q", stderr.String())
	}
	if strings.Contains(stdout.String(), `"kind": "Error"`) {
		t.Fatalf("stdout contains Error envelope despite neutral telemetry failure: %s", stdout.String())
	}
}

func TestTelemetryFailedCommandSpanDoesNotExportRawError(t *testing.T) {
	const marker = "pm_command_failure_marker"
	root := t.TempDir()
	t.Setenv("PM_TELEMETRY", "file")
	var stdout, stderr bytes.Buffer

	code := cli.Run([]string{"--root", root, "--json", "help", marker}, &stdout, &stderr)

	if code == 0 {
		t.Fatalf("exit code = 0, want failure; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	data := readCLITelemetry(t, filepath.Join(root, ".polymetrics", "telemetry"))
	assertCLIContains(t, data, "pm.command")
	assertCLIContains(t, data, "pm.error.type")
	assertCLIContains(t, data, "pm.error.code")
	assertCLIContains(t, data, "internal_error")
	assertCLINotContains(t, data, "exception.")
	assertCLINotContains(t, data, marker)
}

func TestTelemetryConfigSourcedOTLPRejectedAndEnvOptInAccepted(t *testing.T) {
	const marker = "pm_config_otlp_marker"
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".polymetrics"), 0o700); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	configPath := filepath.Join(root, ".polymetrics", "config.yaml")
	configBody := "telemetry:\n  exporter: otlp\n  endpoint: https://user:" + marker + "@collector.example.test/v1/traces?token=" + marker + "\n"
	if err := os.WriteFile(configPath, []byte(configBody), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	var stdout, stderr bytes.Buffer

	code := cli.Run([]string{"--root", root, "version", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"kind": "Version"`) {
		t.Fatalf("stdout missing Version envelope: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "warning: telemetry:") || !strings.Contains(stderr.String(), "config-sourced OTLP") {
		t.Fatalf("stderr missing config-sourced OTLP warning: %q", stderr.String())
	}
	if strings.Contains(stderr.String(), marker) || strings.Contains(stderr.String(), "token=") || strings.Contains(stderr.String(), "user:") {
		t.Fatalf("stderr leaked config endpoint detail: %q", stderr.String())
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	stdout.Reset()
	stderr.Reset()
	root = t.TempDir()
	t.Setenv("PM_TELEMETRY", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)

	code = cli.Run([]string{"--root", root, "version", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("env opt-in exit code = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if strings.Contains(stderr.String(), "config-sourced OTLP") {
		t.Fatalf("env opt-in stderr reported config-sourced rejection: %q", stderr.String())
	}
}

func TestTelemetryOTLPExportFailureUsesProjectWarningAndKeepsStdout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "collector failed", http.StatusInternalServerError)
	}))
	defer server.Close()
	root := t.TempDir()
	t.Setenv("PM_TELEMETRY", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)
	var stdout, stderr bytes.Buffer

	code := cli.Run([]string{"--root", root, "version", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"kind": "Version"`) {
		t.Fatalf("stdout missing Version envelope: %s", stdout.String())
	}
	if strings.Contains(stdout.String(), "telemetry") || strings.Contains(stdout.String(), `"kind": "Error"`) {
		t.Fatalf("stdout corrupted by telemetry failure: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "warning: telemetry:") {
		t.Fatalf("stderr missing telemetry warning: %q", stderr.String())
	}
}

func TestTelemetryCertifyConnectorSpan(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PM_TELEMETRY", "file")
	t.Setenv("PM_CERT_SAMPLE_TOKEN", "sample-cli-token")

	stdout, stderr, code := certifyRun(t, root, "connectors", "certify", "sample", "--from-env", "token=PM_CERT_SAMPLE_TOKEN", "--json")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s", code, stdout, stderr)
	}
	data := readCLITelemetry(t, filepath.Join(root, ".polymetrics", "telemetry"))
	assertCLIContains(t, data, "pm.command")
	assertCLIContains(t, data, "pm.certify.connector")
	assertCLINotContains(t, data, "sample-cli-token")
	assertCLINotContains(t, data, "PM_CERT_SAMPLE_TOKEN")
}

func readCLITelemetry(t *testing.T, dir string) []byte {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read telemetry dir %s: %v", dir, err)
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

func assertCLIContains(t *testing.T, data []byte, needle string) {
	t.Helper()
	if !bytes.Contains(data, []byte(needle)) {
		t.Fatalf("telemetry output missing %q:\n%s", needle, data)
	}
}

func assertCLINotContains(t *testing.T, data []byte, needle string) {
	t.Helper()
	if bytes.Contains(data, []byte(needle)) {
		t.Fatalf("telemetry output contains forbidden %q:\n%s", needle, data)
	}
}
