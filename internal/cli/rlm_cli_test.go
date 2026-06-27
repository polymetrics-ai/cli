package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestRLMRunFixture tests the fixture backend via CLI.
func TestRLMRunFixture(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)

	specPath := writeTestSpec(t)

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"--root", dir,
		"rlm", "run",
		"--spec", specPath,
		"--in", "contacts",
		"--out", "lead_scores",
		"--mode", "fixture",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr.String())
	}

	warehouseDir := filepath.Join(dir, ".polymetrics", "warehouse")
	outPath := filepath.Join(warehouseDir, "lead_scores.ndjson")
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("OutTable not written: %v", err)
	}
}

// TestRLMRunDeterministic tests the deterministic backend via CLI with an InTable present.
func TestRLMRunDeterministic(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)

	specPath := writeTestSpec(t)
	writeTestInTable(t, dir)

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"--root", dir,
		"--json",
		"rlm", "run",
		"--spec", specPath,
		"--in", "contacts",
		"--out", "lead_scores",
		"--mode", "deterministic",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %s, stdout = %s", code, stderr.String(), stdout.String())
	}

	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout is not valid JSON: %v — output: %s", err, stdout.String())
	}
	scored, _ := env["records_scored"].(float64)
	if scored <= 0 {
		t.Errorf("records_scored = %v, want > 0", env["records_scored"])
	}
}

// TestRLMRunModelStub tests that the model backend exits non-zero with "not implemented".
func TestRLMRunModelStub(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)

	specPath := writeTestSpec(t)

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"--root", dir,
		"rlm", "run",
		"--spec", specPath,
		"--in", "contacts",
		"--out", "lead_scores",
		"--mode", "model",
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit for model mode (not implemented)")
	}
	combined := stdout.String() + stderr.String()
	if combined == "" {
		t.Error("expected error output mentioning 'not implemented'")
	}
}

// TestRLMRunMissingSpec tests that a missing spec file causes exit 1.
func TestRLMRunMissingSpec(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"--root", dir,
		"rlm", "run",
		"--spec", "/nonexistent/path/spec.json",
		"--in", "contacts",
		"--out", "lead_scores",
		"--mode", "fixture",
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit for missing spec file")
	}
}

// TestRLMRunMissingOutFlag tests that missing --out flag causes exit 1.
func TestRLMRunMissingOutFlag(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)

	specPath := writeTestSpec(t)

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"--root", dir,
		"rlm", "run",
		"--spec", specPath,
		"--in", "contacts",
		"--mode", "fixture",
		// --out missing
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit when --out is missing")
	}
}

// TestRLMRunDryRun tests that --dry-run skips OutTable creation.
func TestRLMRunDryRun(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)

	specPath := writeTestSpec(t)

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"--root", dir,
		"rlm", "run",
		"--spec", specPath,
		"--in", "contacts",
		"--out", "lead_scores",
		"--mode", "fixture",
		"--dry-run",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr.String())
	}

	warehouseDir := filepath.Join(dir, ".polymetrics", "warehouse")
	outPath := filepath.Join(warehouseDir, "lead_scores.ndjson")
	if _, err := os.Stat(outPath); err == nil {
		t.Error("OutTable should NOT be written when --dry-run is set")
	}
}

// TestRLMRunJSONOutput tests that --json produces a valid JSON object on stdout.
func TestRLMRunJSONOutput(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)

	specPath := writeTestSpec(t)

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"--root", dir,
		"--json",
		"rlm", "run",
		"--spec", specPath,
		"--in", "contacts",
		"--out", "lead_scores",
		"--mode", "fixture",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr.String())
	}

	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout is not valid JSON: %v — output: %s", err, stdout.String())
	}
}

// --- helpers ---

// initProject creates a minimal .polymetrics project directory.
func initProject(t *testing.T, root string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", root, "init"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("init failed: %s", stderr.String())
	}
}

// writeTestSpec writes a minimal JSON spec to a temp file and returns its path.
func writeTestSpec(t *testing.T) string {
	t.Helper()
	spec := `{
		"name": "test-spec",
		"features": [
			{"name": "email", "weight": 0.5, "score_if_set": 1.0},
			{"name": "company", "weight": 0.5, "score_if_set": 1.0}
		]
	}`
	f, err := os.CreateTemp(t.TempDir(), "spec-*.json")
	if err != nil {
		t.Fatalf("create spec file: %v", err)
	}
	if _, err := f.WriteString(spec); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	f.Close()
	return f.Name()
}

// writeTestInTable writes a minimal NDJSON InTable for deterministic mode.
func writeTestInTable(t *testing.T, root string) {
	t.Helper()
	warehouseDir := filepath.Join(root, ".polymetrics", "warehouse")
	if err := os.MkdirAll(warehouseDir, 0o755); err != nil {
		t.Fatalf("mkdir warehouse: %v", err)
	}
	inPath := filepath.Join(warehouseDir, "contacts.ndjson")
	rows := `{"_polymetrics_raw_id":"r1","record":{"email":"a@b.com","company":"Acme"}}
{"_polymetrics_raw_id":"r2","record":{"email":"b@c.com"}}
`
	if err := os.WriteFile(inPath, []byte(rows), 0o644); err != nil {
		t.Fatalf("write InTable: %v", err)
	}
}
