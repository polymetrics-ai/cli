package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
	if cfg["postgres_url"] != "postgres://127.0.0.1:1/polymetrics?sslmode=disable" {
		t.Fatalf("postgres_url = %v, want config-file value", cfg["postgres_url"])
	}
	if cfg["dragonfly_addr"] != "127.0.0.1:1" || cfg["temporal_addr"] != "127.0.0.1:1" {
		t.Fatalf("runtime config = %v, want configured endpoints", cfg)
	}
}

func TestWorkerStatusUsesExplicitConfigFileTemporalAddr(t *testing.T) {
	root := writeMigrationConfig(t, `runtime:
  temporal_addr: 127.0.0.1:1
`)

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", root, "--json", "worker", "status"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not JSON: %v — %s", err, stdout.String())
	}
	if env["addr"] != "127.0.0.1:1" {
		t.Fatalf("addr = %v, want explicit config-file temporal addr", env["addr"])
	}
	if env["status"] != "unavailable" {
		t.Fatalf("status = %v, want unavailable for closed local port", env["status"])
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
