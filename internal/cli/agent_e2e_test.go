package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors"
)

// TestRLMAgent_FakeRunner_EndToEnd runs `pm rlm run --mode agent` with the
// hermetic fake runner (no Temporal/podman) and asserts the OutTable carries the
// exact RLM contract, with _rlm_mode="agent".
func TestRLMAgent_FakeRunner_EndToEnd(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)
	specPath := writeTestSpec(t)
	writeTestInTable(t, dir)

	t.Setenv("POLYMETRICS_RLM_FAKE_RUNNER", "1")

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"--root", dir, "--json",
		"rlm", "run",
		"--spec", specPath,
		"--in", "contacts",
		"--out", "agent_scores",
		"--mode", "agent",
		"--request", "score contacts by engagement",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}

	var res map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("stdout not JSON: %v — %s", err, stdout.String())
	}
	if res["mode"] != "agent" {
		t.Fatalf("mode = %v, want agent", res["mode"])
	}

	outPath := filepath.Join(dir, ".polymetrics", "warehouse", "agent_scores.ndjson")
	rows := readOutRows(t, outPath)
	if len(rows) != 2 {
		t.Fatalf("want 2 output rows, got %d", len(rows))
	}
	for i, r := range rows {
		for _, field := range []string{"_rlm_score", "_rlm_mode", "_rlm_spec", "_rlm_scored_at", "_polymetrics_raw_id"} {
			if _, ok := r[field]; !ok {
				t.Errorf("row %d missing %s: %+v", i, field, r)
			}
		}
		if r["_rlm_mode"] != "agent" {
			t.Errorf("row %d _rlm_mode = %v, want agent", i, r["_rlm_mode"])
		}
	}
	// Sorted score desc.
	if s0, s1 := rows[0]["_rlm_score"].(float64), rows[1]["_rlm_score"].(float64); s0 < s1 {
		t.Errorf("rows not sorted by score desc: %v < %v", s0, s1)
	}
}

// TestRLMAgent_FakeRunner_MatchesDeterministic proves materialization conformance
// (D1/D2): the agent path and the deterministic path produce identical
// per-record scores, ids, and ordering — differing only in _rlm_mode.
func TestRLMAgent_FakeRunner_MatchesDeterministic(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)
	specPath := writeTestSpec(t)
	writeTestInTable(t, dir)

	// Deterministic run.
	var o1, e1 bytes.Buffer
	if code := Run([]string{"--root", dir, "rlm", "run", "--spec", specPath, "--in", "contacts", "--out", "det", "--mode", "deterministic"}, &o1, &e1); code != 0 {
		t.Fatalf("deterministic run failed: %s", e1.String())
	}
	// Agent run (fake runner).
	t.Setenv("POLYMETRICS_RLM_FAKE_RUNNER", "1")
	var o2, e2 bytes.Buffer
	if code := Run([]string{"--root", dir, "rlm", "run", "--spec", specPath, "--in", "contacts", "--out", "agent", "--mode", "agent"}, &o2, &e2); code != 0 {
		t.Fatalf("agent run failed: %s", e2.String())
	}

	whDir := filepath.Join(dir, ".polymetrics", "warehouse")
	det := readOutRows(t, filepath.Join(whDir, "det.ndjson"))
	agent := readOutRows(t, filepath.Join(whDir, "agent.ndjson"))
	if len(det) != len(agent) {
		t.Fatalf("row count differs: det=%d agent=%d", len(det), len(agent))
	}
	for i := range det {
		if det[i]["_polymetrics_raw_id"] != agent[i]["_polymetrics_raw_id"] {
			t.Errorf("row %d id differs: det=%v agent=%v", i, det[i]["_polymetrics_raw_id"], agent[i]["_polymetrics_raw_id"])
		}
		if det[i]["_rlm_score"] != agent[i]["_rlm_score"] {
			t.Errorf("row %d score differs: det=%v agent=%v", i, det[i]["_rlm_score"], agent[i]["_rlm_score"])
		}
	}
	if agent[0]["_rlm_mode"] != "agent" || det[0]["_rlm_mode"] != "deterministic" {
		t.Errorf("modes wrong: det=%v agent=%v", det[0]["_rlm_mode"], agent[0]["_rlm_mode"])
	}
}

func TestRLMAgent_DegradesWhenNoTemporal(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)
	specPath := writeTestSpec(t)
	writeTestInTable(t, dir)
	// No fake runner, no temporal addr → ErrRemoteUnavailable (non-zero exit).
	t.Setenv("POLYMETRICS_RLM_FAKE_RUNNER", "")
	t.Setenv("POLYMETRICS_TEMPORAL_ADDR", "")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", dir, "rlm", "run", "--spec", specPath, "--in", "contacts", "--out", "x", "--mode", "agent"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("want non-zero exit when temporal/podman unavailable")
	}
}

func TestWorkerStatus_Unavailable(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)
	t.Setenv("POLYMETRICS_TEMPORAL_ADDR", "")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", dir, "--json", "worker", "status"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if env["status"] != "unavailable" {
		t.Fatalf("status = %v, want unavailable", env["status"])
	}
}

// TestExtract_RLMExecutesWithFakeRunner verifies the extract RLM branch actually
// runs the agent when --in/--out are provided and the fake runner is enabled.
func TestExtract_RLMExecutesWithFakeRunner(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)
	writeTestInTable(t, dir)
	t.Setenv("POLYMETRICS_RLM_FAKE_RUNNER", "1")

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"--root", dir, "--json",
		"extract", "--request", "predict which contacts will convert",
		"--in", "contacts", "--out", "extract_scores",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("not JSON: %v — %s", err, stdout.String())
	}
	if env["route"] != "rlm_analysis" {
		t.Fatalf("route = %v, want rlm_analysis", env["route"])
	}
	if env["rlm"] == nil {
		t.Fatalf("expected rlm result in envelope, got: %v", env)
	}
	if _, err := os.Stat(filepath.Join(dir, ".polymetrics", "warehouse", "extract_scores.ndjson")); err != nil {
		t.Errorf("OutTable not materialized: %v", err)
	}
}

func readOutRows(t *testing.T, path string) []connectors.Record {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}
	var rows []connectors.Record
	dec := json.NewDecoder(bytes.NewReader(b))
	for dec.More() {
		var r connectors.Record
		if err := dec.Decode(&r); err != nil {
			t.Fatalf("decode: %v", err)
		}
		rows = append(rows, r)
	}
	return rows
}
