package rlm

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func okProbe(context.Context, string) bool   { return true }
func failProbe(context.Context, string) bool { return false }
func foundPath(string) (string, error)       { return "/usr/bin/podman", nil }
func missingPath(string) (string, error)     { return "", errors.New("not found") }

func baseAgent() *AgentAnalyzer {
	return &AgentAnalyzer{
		Cfg:      AgentConfig{TemporalAddr: "localhost:7233", PodmanBin: "podman", Image: "img", MaxIter: 4},
		Probe:    okProbe,
		LookPath: foundPath,
		Submit:   func(context.Context, AgentRequest) (AgentResult, error) { return AgentResult{}, nil },
	}
}

func writeInTable(t *testing.T, whDir string) {
	t.Helper()
	if err := os.MkdirAll(whDir, 0o755); err != nil {
		t.Fatal(err)
	}
	rows := `{"_polymetrics_raw_id":"r1","record":{"email":"a@b.com"}}
{"_polymetrics_raw_id":"r2","record":{"email":"c@d.com"}}
`
	if err := os.WriteFile(filepath.Join(whDir, "contacts.ndjson"), []byte(rows), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAgent_Mode(t *testing.T) {
	if (&AgentAnalyzer{}).Mode() != "agent" {
		t.Fatal("mode should be agent")
	}
}

func TestAgent_Gate_NoTemporalAddr(t *testing.T) {
	a := baseAgent()
	a.Cfg.TemporalAddr = ""
	_, err := a.Run(context.Background(), RunRequest{OutTable: "out", WarehouseDir: t.TempDir()})
	if !errors.Is(err, ErrRemoteUnavailable) {
		t.Fatalf("want ErrRemoteUnavailable, got %v", err)
	}
}

func TestAgent_Gate_PodmanAbsent(t *testing.T) {
	a := baseAgent()
	a.LookPath = missingPath
	_, err := a.Run(context.Background(), RunRequest{OutTable: "out", WarehouseDir: t.TempDir()})
	if !errors.Is(err, ErrRemoteUnavailable) {
		t.Fatalf("want ErrRemoteUnavailable, got %v", err)
	}
}

func TestAgent_Gate_ProbeFails(t *testing.T) {
	a := baseAgent()
	a.Probe = failProbe
	_, err := a.Run(context.Background(), RunRequest{OutTable: "out", WarehouseDir: t.TempDir()})
	if !errors.Is(err, ErrRemoteUnavailable) {
		t.Fatalf("want ErrRemoteUnavailable, got %v", err)
	}
}

func TestAgent_Gate_NilProbeFailsClosed(t *testing.T) {
	a := baseAgent()
	a.Probe = nil
	_, err := a.Run(context.Background(), RunRequest{OutTable: "out", WarehouseDir: t.TempDir()})
	if !errors.Is(err, ErrRemoteUnavailable) {
		t.Fatalf("nil probe should fail closed, got %v", err)
	}
}

func TestAgent_RejectsBadOutTable(t *testing.T) {
	a := baseAgent()
	_, err := a.Run(context.Background(), RunRequest{OutTable: "../escape", WarehouseDir: t.TempDir()})
	if err == nil || errors.Is(err, ErrRemoteUnavailable) {
		t.Fatalf("want OutTable validation error, got %v", err)
	}
}

func TestAgent_Stage_WritesJobDir(t *testing.T) {
	whDir := t.TempDir()
	writeInTable(t, whDir)

	var captured AgentRequest
	a := baseAgent()
	a.Request = "score leads by engagement"
	a.Cfg.JobRoot = t.TempDir()
	a.Submit = func(_ context.Context, req AgentRequest) (AgentResult, error) {
		captured = req
		// Simulate the container writing output + manifest.
		out := []connectors.Record{
			{"_polymetrics_raw_id": "r1", "_rlm_score": 0.9},
			{"_polymetrics_raw_id": "r2", "_rlm_score": 0.4},
		}
		writeAgentFixtureOutput(t, req.JobDir, out)
		return AgentResult{JobDir: req.JobDir, RecordsScored: 2}, nil
	}

	res, err := a.Run(context.Background(), RunRequest{
		Spec:         &Spec{Name: "s"},
		InTable:      "contacts",
		OutTable:     "scored",
		WarehouseDir: whDir,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Staging wrote the input + request descriptor into the JobDir.
	if _, err := os.Stat(filepath.Join(captured.JobDir, "in", "input.ndjson")); err != nil {
		t.Errorf("input.ndjson not staged: %v", err)
	}
	rb, err := os.ReadFile(filepath.Join(captured.JobDir, "in", "request.json"))
	if err != nil {
		t.Fatalf("request.json not staged: %v", err)
	}
	var desc map[string]any
	if err := json.Unmarshal(rb, &desc); err != nil || desc["request"] != "score leads by engagement" {
		t.Fatalf("request descriptor wrong: %v (%v)", desc, err)
	}
	if captured.Fingerprint == "" {
		t.Error("fingerprint should be set")
	}
	if res.RecordsScored != 2 {
		t.Errorf("RecordsScored = %d, want 2", res.RecordsScored)
	}

	// OutTable materialized with the contract fields, sorted score desc.
	out := readPlainNDJSON(t, filepath.Join(whDir, "scored.ndjson"))
	if len(out) != 2 || out[0]["_polymetrics_raw_id"] != "r1" {
		t.Fatalf("OutTable wrong: %+v", out)
	}
	if out[0]["_rlm_mode"] != "agent" {
		t.Errorf("_rlm_mode = %v, want agent", out[0]["_rlm_mode"])
	}
}

func TestAgent_DetectsTruncatedOutput(t *testing.T) {
	whDir := t.TempDir()
	writeInTable(t, whDir)
	a := baseAgent()
	a.Cfg.JobRoot = t.TempDir()
	a.Submit = func(_ context.Context, req AgentRequest) (AgentResult, error) {
		// Write one row but a manifest expecting two → truncation.
		writeAgentFixtureOutputWithManifest(t, req.JobDir,
			[]connectors.Record{{"_polymetrics_raw_id": "r1", "_rlm_score": 0.9}}, 2)
		return AgentResult{JobDir: req.JobDir}, nil
	}
	_, err := a.Run(context.Background(), RunRequest{
		Spec: &Spec{Name: "s"}, InTable: "contacts", OutTable: "scored", WarehouseDir: whDir,
	})
	if err == nil {
		t.Fatal("want truncation error, got nil")
	}
}

func writeAgentFixtureOutput(t *testing.T, jobDir string, rows []connectors.Record) {
	t.Helper()
	writeAgentFixtureOutputWithManifest(t, jobDir, rows, len(rows))
}

func writeAgentFixtureOutputWithManifest(t *testing.T, jobDir string, rows []connectors.Record, expected int) {
	t.Helper()
	outDir := filepath.Join(jobDir, "out")
	if err := os.MkdirAll(outDir, 0o700); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(outDir, "output.ndjson"))
	if err != nil {
		t.Fatal(err)
	}
	enc := json.NewEncoder(f)
	for _, r := range rows {
		if err := enc.Encode(r); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	mb, _ := json.Marshal(agentManifest{ExpectedCount: expected, RecordsRead: expected})
	if err := os.WriteFile(filepath.Join(outDir, "manifest.json"), mb, 0o600); err != nil {
		t.Fatal(err)
	}
}
