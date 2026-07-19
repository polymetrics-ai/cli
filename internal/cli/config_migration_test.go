package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeDoctorUsesConfigFile(t *testing.T) {
	root := writeMigrationConfig(t, `runtime:
  postgres_url: postgres://127.0.0.1:1/polymetrics?sslmode=disable
  dragonfly_addr: 127.0.0.1:1
  temporal_addr: 127.0.0.1:1
`)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", root, "--json", "runtime", "doctor"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not JSON: %v — %s", err, stdout.String())
	}
	cfg, ok := env["config"].(map[string]any)
	if !ok {
		t.Fatalf("missing config object: %v", env)
	}
	if cfg["postgres_url"] != "postgres://127.0.0.1:1/polymetrics" {
		t.Fatalf("postgres_url = %v, want redacted config-file value", cfg["postgres_url"])
	}
	if cfg["dragonfly_addr"] != "127.0.0.1:1" || cfg["temporal_addr"] != "127.0.0.1:1" {
		t.Fatalf("runtime config = %v, want configured endpoints", cfg)
	}
}

func TestWorkerStatusUsesExplicitConfigFileTemporalAddr(t *testing.T) {
	root := writeMigrationConfig(t, `runtime:
  temporal_addr: 127.0.0.1:1
`)

	runtime := newFakeWorkerRuntime()
	stdout, stderr, code, cfg := runWorkerInvocation(t, runtime, "--root", root, "--json", "worker", "status")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
	if runtime.statusCalls != 1 {
		t.Fatalf("injected status calls = %d, want 1", runtime.statusCalls)
	}
	if runtime.statusAddr != "127.0.0.1:1" || cfg.Source("runtime.temporal_addr") != "config" {
		t.Fatalf("status addr/source = %q/%q, want config-file address", runtime.statusAddr, cfg.Source("runtime.temporal_addr"))
	}
	var env map[string]any
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("stdout not JSON: %v — %s", err, stdout)
	}
	if env["addr"] != "127.0.0.1:1" {
		t.Fatalf("addr = %v, want explicit config-file temporal addr", env["addr"])
	}
	if env["status"] != "unavailable" {
		t.Fatalf("status = %v, want unavailable for closed local port", env["status"])
	}
}

func TestWorkerServeThreadsConfigFileActivities(t *testing.T) {
	root := writeMigrationConfig(t, `runtime:
  temporal_addr: 127.0.0.1:7233
rlm:
  image: ghcr.io/example/custom-rlm-agent:test
  podman_bin: /tmp/custom-podman-from-config
`)

	t.Setenv("POLYMETRICS_TEMPORAL_ADDR", "")
	t.Setenv("PM_TEMPORAL_ADDR", "")
	runtime := newFakeWorkerRuntime()
	stdout, stderr, code, _ := runWorkerInvocation(t, runtime, "--root", root, "--json", "worker", "serve")
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr)
	}
	if runtime.serveAddr != "127.0.0.1:7233" {
		t.Fatalf("addr = %q, want config-file temporal addr", runtime.serveAddr)
	}
	if runtime.activities == nil {
		t.Fatal("worker serve did not receive activities")
	}
	if runtime.activities.PodmanBin != "/tmp/custom-podman-from-config" {
		t.Fatalf("PodmanBin = %q, want config-file podman bin", runtime.activities.PodmanBin)
	}
	if runtime.activities.Image != "ghcr.io/example/custom-rlm-agent:test" {
		t.Fatalf("Image = %q, want config-file image", runtime.activities.Image)
	}
	if !strings.Contains(stdout, `"kind": "WorkerServe"`) {
		t.Fatalf("worker serve output = %q", stdout)
	}
}

func TestPerfCompareRuntimeUsesConfigFileEndpoints(t *testing.T) {
	root := writeMigrationConfig(t, `runtime:
  postgres_url: postgres://127.0.0.1:1/polymetrics?sslmode=disable
  dragonfly_addr: 127.0.0.1:2
  temporal_addr: 127.0.0.1:3
`)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", root, "--json", "perf", "compare", "--iterations", "1", "--runtime"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not JSON: %v — %s", err, stdout.String())
	}
	comparison, ok := env["comparison"].(map[string]any)
	if !ok {
		t.Fatalf("missing comparison object: %v", env)
	}
	report, ok := comparison["runtime_report"].(map[string]any)
	if !ok {
		t.Fatalf("missing runtime report: %v", comparison)
	}
	checks, ok := report["checks"].([]any)
	if !ok {
		t.Fatalf("missing runtime checks: %v", report)
	}
	endpoints := map[string]string{}
	for _, item := range checks {
		check, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("check has unexpected shape: %v", item)
		}
		name, _ := check["name"].(string)
		endpoint, _ := check["endpoint"].(string)
		endpoints[name] = endpoint
	}
	if endpoints["postgres"] != "postgres://127.0.0.1:1/polymetrics" {
		t.Fatalf("postgres endpoint = %q, want redacted config-file endpoint", endpoints["postgres"])
	}
	if endpoints["dragonfly"] != "127.0.0.1:2" {
		t.Fatalf("dragonfly endpoint = %q, want config-file endpoint", endpoints["dragonfly"])
	}
	if endpoints["temporal"] != "127.0.0.1:3" {
		t.Fatalf("temporal endpoint = %q, want config-file endpoint", endpoints["temporal"])
	}
}

func TestScheduleInstallUsesConfigFileCrontabPath(t *testing.T) {
	crontabFile := t.TempDir() + "/crontab"
	root := writeMigrationConfig(t, "schedule:\n  crontab_file: "+crontabFile+"\n")
	initProject(t, root)

	_, stderr, code := scheduleRun(t, root, "schedule", "create",
		"--name", "nightly-leads",
		"--cron", "0 2 * * *",
		"--flow", "likely-customers",
	)
	if code != 0 {
		t.Fatalf("create: exit %d, stderr=%q", code, stderr)
	}
	_, stderr, code = scheduleRun(t, root, "schedule", "install", "nightly-leads", "--crontab")
	if code != 0 {
		t.Fatalf("install: exit %d, stderr=%q", code, stderr)
	}
	data, err := os.ReadFile(crontabFile)
	if err != nil {
		t.Fatalf("read configured crontab: %v", err)
	}
	if !bytes.Contains(data, []byte("pm-schedule-nightly-leads")) {
		t.Fatalf("configured crontab missing sentinel: %q", string(data))
	}
}

func TestAgentImageUsesConfigFilePodmanBin(t *testing.T) {
	root := writeMigrationConfig(t, "rlm:\n  podman_bin: definitely-not-a-real-binary-from-config\n")
	initProject(t, root)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", root, "agent", "image", "build"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("want non-zero exit when configured podman bin is absent")
	}
	if !bytes.Contains(stderr.Bytes(), []byte("definitely-not-a-real-binary-from-config")) {
		t.Fatalf("stderr = %q, want configured podman bin", stderr.String())
	}
}

func TestRLMAgentFakeRunnerUsesConfigFile(t *testing.T) {
	root := writeMigrationConfig(t, `rlm:
  fake_runner: true
`)
	initProject(t, root)
	specPath := writeTestSpec(t)
	writeTestInTable(t, root)

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"--root", root, "--json",
		"rlm", "run",
		"--spec", specPath,
		"--in", "contacts",
		"--out", "config_agent_scores",
		"--mode", "agent",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".polymetrics", "warehouse", "config_agent_scores.ndjson")); err != nil {
		t.Fatalf("agent fake-runner output not materialized: %v", err)
	}
}

func writeMigrationConfig(t *testing.T, yaml string) string {
	t.Helper()
	for _, name := range []string{
		"POLYMETRICS_POSTGRES_URL", "PM_POSTGRES_URL",
		"POLYMETRICS_DRAGONFLY_ADDR", "PM_DRAGONFLY_ADDR",
		"POLYMETRICS_TEMPORAL_ADDR", "PM_TEMPORAL_ADDR",
		"POLYMETRICS_RLM_IMAGE", "PM_RLM_IMAGE",
		"POLYMETRICS_PODMAN_BIN", "PM_PODMAN_BIN",
		"POLYMETRICS_RLM_FAKE_RUNNER", "PM_RLM_FAKE_RUNNER",
		"POLYMETRICS_RLM_EMBEDDED_WORKER", "PM_RLM_EMBEDDED_WORKER",
		"POLYMETRICS_CRONTAB_FILE", "PM_CRONTAB_FILE",
	} {
		t.Setenv(name, "")
	}
	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	return root
}
