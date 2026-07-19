package cli_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
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

func TestTelemetryRejectsUnsafeAmbientOTLPTracesEndpointBeforeExporter(t *testing.T) {
	const marker = "pm_ambient_traces_endpoint_marker"
	hitCh := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case hitCh <- struct{}{}:
		default:
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	root := t.TempDir()
	t.Setenv("PM_TELEMETRY", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", server.URL+"/v1/traces?token="+marker)
	var stdout, stderr bytes.Buffer
	var code int

	processStderr := captureProcessStderr(t, func() {
		code = cli.Run([]string{"--root", root, "version", "--json"}, &stdout, &stderr)
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s processStderr=%s", code, stdout.String(), stderr.String(), processStderr)
	}
	if !strings.Contains(stdout.String(), `"kind": "Version"`) {
		t.Fatalf("stdout missing Version envelope: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "warning: telemetry:") || !strings.Contains(stderr.String(), "invalid OTLP endpoint") {
		t.Fatalf("project stderr missing redacted OTLP endpoint warning: %q", stderr.String())
	}
	if processStderr != "" {
		t.Fatalf("process stderr = %q, want empty", processStderr)
	}
	for _, forbidden := range []string{marker, "token=", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"} {
		if strings.Contains(stderr.String(), forbidden) || strings.Contains(processStderr, forbidden) {
			t.Fatalf("stderr leaked %q: project=%q process=%q", forbidden, stderr.String(), processStderr)
		}
	}
	select {
	case <-hitCh:
		t.Fatal("unsafe ambient traces endpoint was used as collector")
	default:
	}
}

func TestTelemetryNeutralizesUnsupportedAmbientOTLPHeaders(t *testing.T) {
	const marker = "pm_ambient_header_marker"
	authCh := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case authCh <- r.Header.Get("Authorization"):
		default:
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	root := t.TempDir()
	t.Setenv("PM_TELEMETRY", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization=Bearer%20"+marker+",X-Bad=%zz"+marker)
	var stdout, stderr bytes.Buffer
	var code int

	processStderr := captureProcessStderr(t, func() {
		code = cli.Run([]string{"--root", root, "version", "--json"}, &stdout, &stderr)
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s processStderr=%s", code, stdout.String(), stderr.String(), processStderr)
	}
	if !strings.Contains(stdout.String(), `"kind": "Version"`) {
		t.Fatalf("stdout missing Version envelope: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "warning: telemetry:") || !strings.Contains(stderr.String(), "OTEL_EXPORTER_OTLP_HEADERS") {
		t.Fatalf("project stderr missing unsupported headers warning: %q", stderr.String())
	}
	if processStderr != "" {
		t.Fatalf("process stderr = %q, want empty", processStderr)
	}
	for _, forbidden := range []string{marker, "Authorization", "Bearer", "%zz"} {
		if strings.Contains(stderr.String(), forbidden) || strings.Contains(processStderr, forbidden) {
			t.Fatalf("stderr leaked %q: project=%q process=%q", forbidden, stderr.String(), processStderr)
		}
	}
	select {
	case auth := <-authCh:
		if auth != "" {
			t.Fatalf("collector received ambient Authorization header %q", auth)
		}
	default:
		t.Fatal("collector was not called for safe OTLP endpoint")
	}
}

func TestTelemetrySanitizesSDKResourceEnvForFileExporter(t *testing.T) {
	const (
		marker        = "pm_sdk_resource_marker_file"
		serviceMarker = "pm_sdk_service_marker_file"
	)
	root := t.TempDir()

	stdout, stderr, code := runTelemetryVersionSubprocess(
		t,
		root,
		"PM_TELEMETRY=file",
		"OTEL_RESOURCE_ATTRIBUTES=api_key="+marker,
		"OTEL_SERVICE_NAME="+serviceMarker,
	)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, `"kind": "Version"`) {
		t.Fatalf("stdout missing Version envelope: %s", stdout)
	}
	if !strings.Contains(stderr, "warning: telemetry:") || !strings.Contains(stderr, "OTEL_RESOURCE_ATTRIBUTES") || !strings.Contains(stderr, "OTEL_SERVICE_NAME") {
		t.Fatalf("stderr missing SDK env warnings by name: %q", stderr)
	}
	for _, forbidden := range []string{marker, serviceMarker, "api_key="} {
		if strings.Contains(stderr, forbidden) {
			t.Fatalf("stderr leaked %q: %q", forbidden, stderr)
		}
	}

	data := readCLITelemetry(t, filepath.Join(root, ".polymetrics", "telemetry"))
	assertCLIContains(t, data, "pm.command")
	for _, forbidden := range []string{marker, serviceMarker, "api_key"} {
		assertCLINotContains(t, data, forbidden)
	}
}

func TestTelemetrySanitizesSDKResourceEnvForOTLPExporter(t *testing.T) {
	const marker = "pm_sdk_resource_marker_otlp"
	bodyCh := make(chan []byte, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		select {
		case bodyCh <- body:
		default:
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	root := t.TempDir()

	stdout, stderr, code := runTelemetryVersionSubprocess(
		t,
		root,
		"PM_TELEMETRY=otlp",
		"OTEL_EXPORTER_OTLP_ENDPOINT="+server.URL,
		"OTEL_RESOURCE_ATTRIBUTES=api_key="+marker,
	)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, `"kind": "Version"`) {
		t.Fatalf("stdout missing Version envelope: %s", stdout)
	}
	if !strings.Contains(stderr, "warning: telemetry:") || !strings.Contains(stderr, "OTEL_RESOURCE_ATTRIBUTES") {
		t.Fatalf("stderr missing resource env warning by name: %q", stderr)
	}
	if strings.Contains(stderr, marker) || strings.Contains(stderr, "api_key=") {
		t.Fatalf("stderr leaked resource env value: %q", stderr)
	}

	select {
	case body := <-bodyCh:
		if bytes.Contains(body, []byte(marker)) || bytes.Contains(body, []byte("api_key")) {
			t.Fatalf("OTLP payload leaked ambient resource attr: %q", string(body))
		}
	default:
		t.Fatal("collector did not receive OTLP export")
	}
}

func TestTelemetrySanitizesInvalidSamplerEnvBeforeProvider(t *testing.T) {
	const marker = "pm_invalid_sampler_marker"
	root := t.TempDir()
	t.Setenv("PM_TELEMETRY", "file")
	t.Setenv("OTEL_TRACES_SAMPLER", "traceidratio")
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "invalid-"+marker)
	var stdout, stderr bytes.Buffer
	var code int

	processStderr := captureProcessStderr(t, func() {
		code = cli.Run([]string{"--root", root, "version", "--json"}, &stdout, &stderr)
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s processStderr=%s", code, stdout.String(), stderr.String(), processStderr)
	}
	if !strings.Contains(stdout.String(), `"kind": "Version"`) {
		t.Fatalf("stdout missing Version envelope: %s", stdout.String())
	}
	if processStderr != "" {
		t.Fatalf("process stderr = %q, want empty", processStderr)
	}
	if !strings.Contains(stderr.String(), "warning: telemetry:") || !strings.Contains(stderr.String(), "OTEL_TRACES_SAMPLER") || !strings.Contains(stderr.String(), "OTEL_TRACES_SAMPLER_ARG") {
		t.Fatalf("project stderr missing sampler env warnings by name: %q", stderr.String())
	}
	for _, forbidden := range []string{marker, "invalid-"} {
		if strings.Contains(stderr.String(), forbidden) || strings.Contains(processStderr, forbidden) {
			t.Fatalf("stderr leaked %q: project=%q process=%q", forbidden, stderr.String(), processStderr)
		}
	}
}

func TestTelemetryOTEL_GO_XObservabilityEnvWarningIsProjectOnly(t *testing.T) {
	assertTelemetryGoXObservabilityEnvProjectOnly(t, "OTEL_GO_X_OBSERVABILITY")
}

func TestTelemetryOTEL_GO_XSelfObservabilityEnvWarningIsProjectOnly(t *testing.T) {
	assertTelemetryGoXObservabilityEnvProjectOnly(t, "OTEL_GO_X_SELF_OBSERVABILITY")
}

func assertTelemetryGoXObservabilityEnvProjectOnly(t *testing.T, envName string) {
	t.Helper()
	const marker = "pm_go_x_observability_marker"
	root := t.TempDir()
	t.Setenv("PM_TELEMETRY", "file")
	t.Setenv(envName, "true")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "api_key="+marker)
	var stdout, stderr bytes.Buffer
	var code int

	processStderr := captureProcessStderr(t, func() {
		code = cli.Run([]string{"--root", root, "version", "--json"}, &stdout, &stderr)
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s processStderr=%s", code, stdout.String(), stderr.String(), processStderr)
	}
	if !strings.Contains(stdout.String(), `"kind": "Version"`) {
		t.Fatalf("stdout missing Version envelope: %s", stdout.String())
	}
	if processStderr != "" {
		t.Fatalf("process stderr = %q, want empty", processStderr)
	}
	if !strings.Contains(stderr.String(), "warning: telemetry:") || !strings.Contains(stderr.String(), envName) {
		t.Fatalf("project stderr missing %s warning by env name: %q", envName, stderr.String())
	}
	for _, forbidden := range []string{marker, "api_key=", "=true"} {
		if strings.Contains(stderr.String(), forbidden) || strings.Contains(processStderr, forbidden) {
			t.Fatalf("stderr leaked %q: project=%q process=%q", forbidden, stderr.String(), processStderr)
		}
	}

	data := readCLITelemetry(t, filepath.Join(root, ".polymetrics", "telemetry"))
	assertCLIContains(t, data, "pm.command")
	for _, forbidden := range []string{
		marker,
		"api_key",
		envName,
		"self_observability",
		"self-observability",
		"otel.sdk",
		"sdk.span",
		"go.opentelemetry.io/otel/sdk/trace",
	} {
		assertCLINotContains(t, data, forbidden)
	}
}

func TestTelemetrySelfObservabilityEnvWarnsBeforeFileExporterFailure(t *testing.T) {
	const marker = "pm_file_exporter_selfobs_marker"
	root := t.TempDir()
	notDir := filepath.Join(root, "not-dir")
	if err := os.WriteFile(notDir, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write not-dir: %v", err)
	}
	t.Setenv("PM_TELEMETRY", "file")
	t.Setenv("PM_TELEMETRY_DIR", filepath.Join("not-dir", "telemetry"))
	t.Setenv("OTEL_GO_X_SELF_OBSERVABILITY", marker)
	var stdout, stderr bytes.Buffer
	var code int

	processStderr := captureProcessStderr(t, func() {
		code = cli.Run([]string{"--root", root, "version", "--json"}, &stdout, &stderr)
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s processStderr=%s", code, stdout.String(), stderr.String(), processStderr)
	}
	if !strings.Contains(stdout.String(), `"kind": "Version"`) {
		t.Fatalf("stdout missing Version envelope: %s", stdout.String())
	}
	if processStderr != "" {
		t.Fatalf("process stderr = %q, want empty", processStderr)
	}
	if !strings.Contains(stderr.String(), "warning: telemetry:") || !strings.Contains(stderr.String(), "OTEL_GO_X_SELF_OBSERVABILITY") {
		t.Fatalf("project stderr missing self-observability warning before file exporter failure: %q", stderr.String())
	}
	for _, forbidden := range []string{marker, "selfobs_marker", "=true"} {
		if strings.Contains(stderr.String(), forbidden) || strings.Contains(processStderr, forbidden) || strings.Contains(stdout.String(), forbidden) {
			t.Fatalf("output leaked %q: stdout=%q project=%q process=%q", forbidden, stdout.String(), stderr.String(), processStderr)
		}
	}
}

func TestTelemetrySelfObservabilityEnvWarnsBeforeOTLPExporterFailure(t *testing.T) {
	const marker = "pm_otlp_exporter_selfobs_marker"
	root := t.TempDir()
	t.Setenv("PM_TELEMETRY", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://collector.example.test/v1/traces?token="+marker)
	t.Setenv("OTEL_GO_X_OBSERVABILITY", marker)
	var stdout, stderr bytes.Buffer
	var code int

	processStderr := captureProcessStderr(t, func() {
		code = cli.Run([]string{"--root", root, "version", "--json"}, &stdout, &stderr)
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s processStderr=%s", code, stdout.String(), stderr.String(), processStderr)
	}
	if !strings.Contains(stdout.String(), `"kind": "Version"`) {
		t.Fatalf("stdout missing Version envelope: %s", stdout.String())
	}
	if processStderr != "" {
		t.Fatalf("process stderr = %q, want empty", processStderr)
	}
	if !strings.Contains(stderr.String(), "warning: telemetry:") || !strings.Contains(stderr.String(), "OTEL_GO_X_OBSERVABILITY") {
		t.Fatalf("project stderr missing self-observability warning before OTLP exporter failure: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "invalid OTLP endpoint") {
		t.Fatalf("project stderr missing invalid endpoint warning: %q", stderr.String())
	}
	for _, forbidden := range []string{marker, "token=", "selfobs_marker", "=true"} {
		if strings.Contains(stderr.String(), forbidden) || strings.Contains(processStderr, forbidden) || strings.Contains(stdout.String(), forbidden) {
			t.Fatalf("output leaked %q: stdout=%q project=%q process=%q", forbidden, stdout.String(), stderr.String(), processStderr)
		}
	}
}

func TestTelemetryOTLPSelfObservabilityEnvWarningIsProjectOnly(t *testing.T) {
	for _, envName := range []string{"OTEL_GO_X_OBSERVABILITY", "OTEL_GO_X_SELF_OBSERVABILITY"} {
		t.Run(envName, func(t *testing.T) {
			const marker = "pm_otlp_self_observability_marker"
			bodyCh := make(chan []byte, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				select {
				case bodyCh <- body:
				default:
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()
			root := t.TempDir()
			t.Setenv("PM_TELEMETRY", "otlp")
			t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)
			t.Setenv(envName, marker)
			t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "api_key="+marker)
			var stdout, stderr bytes.Buffer
			var code int

			processStderr := captureProcessStderr(t, func() {
				code = cli.Run([]string{"--root", root, "version", "--json"}, &stdout, &stderr)
			})

			if code != 0 {
				t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s processStderr=%s", code, stdout.String(), stderr.String(), processStderr)
			}
			if !strings.Contains(stdout.String(), `"kind": "Version"`) {
				t.Fatalf("stdout missing Version envelope: %s", stdout.String())
			}
			if processStderr != "" {
				t.Fatalf("process stderr = %q, want empty", processStderr)
			}
			if !strings.Contains(stderr.String(), "warning: telemetry:") || !strings.Contains(stderr.String(), envName) {
				t.Fatalf("project stderr missing %s warning by env name: %q", envName, stderr.String())
			}
			for _, forbidden := range []string{marker, "api_key=", "=true"} {
				if strings.Contains(stderr.String(), forbidden) || strings.Contains(processStderr, forbidden) || strings.Contains(stdout.String(), forbidden) {
					t.Fatalf("output leaked %q: stdout=%q project=%q process=%q", forbidden, stdout.String(), stderr.String(), processStderr)
				}
			}

			select {
			case body := <-bodyCh:
				for _, forbidden := range []string{
					marker,
					"api_key",
					envName,
					"self_observability",
					"self-observability",
					"otel.sdk",
					"sdk.span",
					"go.opentelemetry.io/otel/sdk/trace",
				} {
					if bytes.Contains(body, []byte(forbidden)) {
						t.Fatalf("OTLP payload leaked %q: %q", forbidden, string(body))
					}
				}
			default:
				t.Fatal("collector did not receive OTLP export")
			}
		})
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

func TestTelemetryCertifyInvalidOptionsPreserveSingleSpanAndConnectorValidationPrecedence(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PM_TELEMETRY", "file")

	stdout, stderr, code := certifyRun(t, root, "connectors", "certify", "../sample", "--from-env=bad", "--json")
	if code != 3 {
		t.Fatalf("exit code = %d, want connector validation exit 3; stdout=%s stderr=%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, `"category": "validation"`) {
		t.Fatalf("stdout missing validation envelope: %s", stdout)
	}
	data := readCLITelemetry(t, filepath.Join(root, ".polymetrics", "telemetry"))
	const spanName = `"Name":"pm.certify.connector"`
	if got := bytes.Count(data, []byte(spanName)); got != 1 {
		t.Fatalf("certify connector span count = %d, want 1:\n%s", got, data)
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

func TestTelemetrySubprocessHelper(t *testing.T) {
	if os.Getenv("PM_TELEMETRY_TEST_HELPER") != "1" {
		return
	}
	root := os.Getenv("PM_TELEMETRY_TEST_ROOT")
	if root == "" {
		_, _ = os.Stderr.WriteString("missing PM_TELEMETRY_TEST_ROOT")
		os.Exit(2)
	}
	code := cli.Run([]string{"--root", root, "version", "--json"}, os.Stdout, os.Stderr)
	os.Exit(code)
}

func runTelemetryVersionSubprocess(t *testing.T, root string, env ...string) (string, string, int) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("resolve test executable: %v", err)
	}
	cmd := exec.Command(exe, "-test.run=TestTelemetrySubprocessHelper")
	cmd.Env = filteredTelemetrySubprocessEnv(append([]string{
		"PM_TELEMETRY_TEST_HELPER=1",
		"PM_TELEMETRY_TEST_ROOT=" + root,
	}, env...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	code := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("run telemetry helper: %v", err)
		}
		code = exitErr.ExitCode()
	}
	return stdout.String(), stderr.String(), code
}

func filteredTelemetrySubprocessEnv(extra ...string) []string {
	env := []string{}
	for _, item := range os.Environ() {
		name, _, _ := strings.Cut(item, "=")
		if strings.HasPrefix(name, "OTEL_") || strings.HasPrefix(name, "PM_TELEMETRY") || strings.HasPrefix(name, "POLYMETRICS_TELEMETRY") || name == "POLYMETRICS_OTEL_EXPORTER_OTLP_ENDPOINT" {
			continue
		}
		env = append(env, item)
	}
	return append(env, extra...)
}

func captureProcessStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}
	savedFD, err := syscall.Dup(int(old.Fd()))
	if err != nil {
		_ = r.Close()
		_ = w.Close()
		t.Fatalf("dup stderr: %v", err)
	}
	if err := syscall.Dup2(int(w.Fd()), int(old.Fd())); err != nil {
		_ = syscall.Close(savedFD)
		_ = r.Close()
		_ = w.Close()
		t.Fatalf("redirect stderr: %v", err)
	}
	os.Stderr = w
	outCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outCh <- buf.String()
	}()
	restored := false
	restore := func() {
		if restored {
			return
		}
		restored = true
		_ = syscall.Dup2(savedFD, int(old.Fd()))
		_ = syscall.Close(savedFD)
		os.Stderr = old
		_ = w.Close()
		_ = r.Close()
	}
	defer restore()

	fn()
	restore()
	return <-outCh
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
