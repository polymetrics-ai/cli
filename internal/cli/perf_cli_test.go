package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPerfCompareFlagFormsPreserveLegacySemantics(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	args := []string{
		"perf", "compare", "extra-positional",
		"--iterations", "1",
		"--iterations=2",
		"--runtime", "true",
		"--runtime=false",
		"--unknown", "value",
		"--root", root, "--json",
	}
	code := Run(args, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(%v) code = %d, want 0; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String()+stderr.String(), "unknown flag") {
		t.Fatalf("Run(%v) rejected legacy-tolerated unknown flag: stdout=%s stderr=%s", args, stdout.String(), stderr.String())
	}

	var env struct {
		Kind       string `json:"kind"`
		Comparison struct {
			DependencyFree struct {
				Iterations int `json:"iterations"`
				Records    int `json:"records"`
			} `json:"dependency_free"`
			RuntimeBacked map[string]any `json:"runtime_backed"`
			RuntimeReport map[string]any `json:"runtime_report"`
		} `json:"comparison"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("decode perf compare json: %v\n%s", err, stdout.String())
	}
	if env.Kind != "PerformanceComparison" {
		t.Fatalf("kind = %q, want PerformanceComparison", env.Kind)
	}
	if env.Comparison.DependencyFree.Iterations != 2 {
		t.Fatalf("iterations = %d, want repeated scalar last-wins value 2", env.Comparison.DependencyFree.Iterations)
	}
	if env.Comparison.DependencyFree.Records != 6 {
		t.Fatalf("records = %d, want 6 from two dependency-free iterations", env.Comparison.DependencyFree.Records)
	}
	if env.Comparison.RuntimeBacked != nil || env.Comparison.RuntimeReport != nil {
		t.Fatalf("--runtime=false should disable runtime comparison, got runtime_backed=%v runtime_report=%v", env.Comparison.RuntimeBacked, env.Comparison.RuntimeReport)
	}
}

func TestPerfCompareBareRuntimeFlagUsesConfig(t *testing.T) {
	root := writeMigrationConfig(t, `runtime:
  postgres_url: postgres://127.0.0.1:1/polymetrics?sslmode=disable
  dragonfly_addr: 127.0.0.1:2
  temporal_addr: 127.0.0.1:3
`)

	var stdout, stderr bytes.Buffer
	args := []string{"--root", root, "--json", "perf", "compare", "--iterations", "1", "--runtime"}
	code := Run(args, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(%v) code = %d, want 0; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
	}

	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("decode perf compare json: %v\n%s", err, stdout.String())
	}
	comparison, ok := env["comparison"].(map[string]any)
	if !ok {
		t.Fatalf("missing comparison object: %v", env)
	}
	if _, ok := comparison["runtime_backed"].(map[string]any); !ok {
		t.Fatalf("bare --runtime did not request runtime-backed comparison: %v", comparison)
	}
	report, ok := comparison["runtime_report"].(map[string]any)
	if !ok {
		t.Fatalf("bare --runtime did not include runtime report: %v", comparison)
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

func TestPerfCompareBareIterationsValidation(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"perf", "compare", "--iterations", "--json"}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("Run(perf compare --iterations --json) code = %d, want 3; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"category": "validation"`) || !strings.Contains(stdout.String(), `invalid --iterations`) {
		t.Fatalf("bare --iterations did not preserve validation sentinel: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}

func TestPerfSyncModesFlagFormsPreserveLegacySemantics(t *testing.T) {
	var stdout, stderr bytes.Buffer
	args := []string{
		"perf", "sync-modes", "extra-positional",
		"--records", "10",
		"--records=20",
		"--unknown", "value",
		"--json",
	}
	code := Run(args, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(%v) code = %d, want 0; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String()+stderr.String(), "unknown flag") {
		t.Fatalf("Run(%v) rejected legacy-tolerated unknown flag: stdout=%s stderr=%s", args, stdout.String(), stderr.String())
	}

	var env struct {
		Kind      string `json:"kind"`
		Benchmark struct {
			Records int `json:"records"`
			Results []struct {
				Mode    string `json:"mode"`
				Records int    `json:"records"`
			} `json:"results"`
		} `json:"benchmark"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("decode sync-modes json: %v\n%s", err, stdout.String())
	}
	if env.Kind != "SyncModeBenchmark" {
		t.Fatalf("kind = %q, want SyncModeBenchmark", env.Kind)
	}
	if env.Benchmark.Records != 20 {
		t.Fatalf("records = %d, want repeated scalar last-wins value 20", env.Benchmark.Records)
	}
	if len(env.Benchmark.Results) == 0 {
		t.Fatalf("benchmark results empty: %+v", env.Benchmark)
	}
	for _, result := range env.Benchmark.Results {
		if result.Records != 20 {
			t.Fatalf("result %s records = %d, want 20", result.Mode, result.Records)
		}
	}
}

func TestPerfBareAndInvalidActionSemantics(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"perf"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(perf) code = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "pm perf - compare dependency-free") {
		t.Fatalf("bare perf did not render contextual help: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"perf", "bogus", "--json"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("Run(perf bogus --json) code = %d, want 2; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"category": "usage"`) {
		t.Fatalf("invalid perf action did not produce usage error: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String(), `"kind": "CommandManual"`) {
		t.Fatalf("invalid perf action rendered contextual help instead of usage error: stdout=%s", stdout.String())
	}
}
