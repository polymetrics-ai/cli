package rlm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"polymetrics.ai/internal/connectors"
)

// ErrRemoteUnavailable is returned when the RLM agent backend cannot run because
// its opt-in dependencies are missing (no Temporal address, Temporal unreachable,
// or podman absent). Callers should fall back to a deterministic/fixture run.
var ErrRemoteUnavailable = errors.New("rlm: remote agent backend unavailable (set POLYMETRICS_TEMPORAL_ADDR and install podman, or use --mode deterministic)")

// AgentRequest is the JSON-safe payload handed to the Temporal worker. It carries
// only a reference to the staged JobDir and metadata — never warehouse row data,
// so rows do not enter Temporal's persisted history.
type AgentRequest struct {
	Fingerprint string `json:"fingerprint"`
	JobDir      string `json:"job_dir"`
	Image       string `json:"image"`
	MaxIter     int    `json:"max_iter"`
	Request     string `json:"request"`
	SpecName    string `json:"spec_name"`
}

// AgentResult is what the worker returns. It also carries only counts and the
// JobDir reference; the scored rows are read from disk by the analyzer.
type AgentResult struct {
	JobDir        string `json:"job_dir"`
	RecordsRead   int    `json:"records_read"`
	RecordsScored int    `json:"records_scored"`
	RecordsFailed int    `json:"records_failed"`
}

// SubmitFunc is the Temporal seam: it runs the WEVC workflow (podman pi-mono
// agent) for req and returns when the job is done. Injected by the CLI so the rlm
// package never imports go.temporal.io.
type SubmitFunc func(ctx context.Context, req AgentRequest) (AgentResult, error)

// AgentConfig configures the RLM agent backend.
type AgentConfig struct {
	TemporalAddr string
	Image        string
	MaxIter      int
	PodmanBin    string // default "podman"
	JobRoot      string // staging root; default os.TempDir()
}

// AgentConfigFromEnv reads the agent config from the environment. TemporalAddr
// defaults to empty (disabled) so --mode agent is never silently chosen.
func AgentConfigFromEnv(getenv func(string) string) AgentConfig {
	if getenv == nil {
		getenv = func(string) string { return "" }
	}
	cfg := AgentConfig{
		TemporalAddr: getenv("POLYMETRICS_TEMPORAL_ADDR"),
		Image:        getenv("POLYMETRICS_RLM_IMAGE"),
		PodmanBin:    getenv("POLYMETRICS_PODMAN_BIN"),
	}
	if cfg.PodmanBin == "" {
		cfg.PodmanBin = "podman"
	}
	if cfg.Image == "" {
		cfg.Image = "ghcr.io/polymetrics/rlm-agent:latest"
	}
	cfg.MaxIter = 4
	return cfg
}

// AgentAnalyzer runs scoring via the containerized PI mono RLM agent, orchestrated
// over Temporal. Probe and Submit are injected so this package stays Temporal-free.
type AgentAnalyzer struct {
	Cfg      AgentConfig
	Probe    func(ctx context.Context, addr string) bool
	Submit   SubmitFunc
	LookPath func(string) (string, error) // defaults to exec.LookPath

	// Request is the natural-language analysis request handed to the agent.
	Request string
}

// Mode returns the backend identifier.
func (a *AgentAnalyzer) Mode() string { return "agent" }

// available reports whether the opt-in dependencies are present. It fails closed:
// a missing Probe, empty Temporal address, unreachable Temporal, or absent podman
// all return false.
func (a *AgentAnalyzer) available(ctx context.Context) bool {
	if a.Cfg.TemporalAddr == "" {
		return false
	}
	lookPath := a.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	if _, err := lookPath(a.Cfg.PodmanBin); err != nil {
		return false
	}
	if a.Probe == nil {
		return false
	}
	return a.Probe(ctx, a.Cfg.TemporalAddr)
}

// Run gates on opt-in deps, stages a JobDir, submits the Temporal workflow, then
// materializes the agent's output to OutTable via the shared writer.
func (a *AgentAnalyzer) Run(ctx context.Context, req RunRequest) (RunResult, error) {
	start := time.Now()
	result := RunResult{Mode: a.Mode(), InTable: req.InTable, OutTable: req.OutTable, DryRun: req.DryRun}

	if !a.available(ctx) {
		return result, ErrRemoteUnavailable
	}
	if a.Submit == nil {
		return result, fmt.Errorf("rlm: agent backend has no submitter wired")
	}
	if err := validateOutTable(req.OutTable); err != nil {
		return result, err
	}

	jobDir, fingerprint, recordsRead, err := a.stage(req)
	if err != nil {
		return result, err
	}
	result.RecordsRead = recordsRead

	agentReq := AgentRequest{
		Fingerprint: fingerprint,
		JobDir:      jobDir,
		Image:       a.Cfg.Image,
		MaxIter:     a.Cfg.MaxIter,
		Request:     a.Request,
		SpecName:    specName(req.Spec),
	}
	out, err := a.Submit(ctx, agentReq)
	if err != nil {
		return result, fmt.Errorf("rlm: agent run: %w", err)
	}

	records, err := readAgentOutput(jobDir)
	if err != nil {
		return result, err
	}
	result.RecordsScored = len(records)
	result.RecordsFailed = out.RecordsFailed

	if !req.DryRun {
		sortScored(records)
		now := time.Now().UTC().Format(time.RFC3339)
		if err := writeOutTable(req.WarehouseDir, req.OutTable, records, a.Mode(), specName(req.Spec), now); err != nil {
			return result, fmt.Errorf("rlm: write OutTable: %w", err)
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// stage copies the InTable into a fresh JobDir's in/ directory, writes the
// request descriptor, and returns the JobDir, a content fingerprint, and the
// input record count.
func (a *AgentAnalyzer) stage(req RunRequest) (jobDir, fingerprint string, recordsRead int, err error) {
	inPath := filepath.Join(req.WarehouseDir, req.InTable+".ndjson")
	inBytes, err := os.ReadFile(inPath)
	if err != nil {
		return "", "", 0, fmt.Errorf("rlm: read InTable %q: %w", inPath, err)
	}
	records, err := readEnvelopedRecords(inPath)
	if err != nil {
		return "", "", 0, fmt.Errorf("rlm: parse InTable %q: %w", inPath, err)
	}

	specBytes, _ := json.Marshal(req.Spec)
	fingerprint = fingerprintBytes(inBytes, []byte(a.Request), specBytes)

	root := a.Cfg.JobRoot
	if root == "" {
		root = os.TempDir()
	}
	jobDir, err = os.MkdirTemp(root, "pm-rlm-"+fingerprint+"-")
	if err != nil {
		return "", "", 0, fmt.Errorf("rlm: create job dir: %w", err)
	}
	inDir := filepath.Join(jobDir, "in")
	outDir := filepath.Join(jobDir, "out")
	for _, d := range []string{inDir, outDir} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			return "", "", 0, fmt.Errorf("rlm: create job subdir: %w", err)
		}
	}
	if err := os.WriteFile(filepath.Join(inDir, "input.ndjson"), inBytes, 0o600); err != nil {
		return "", "", 0, fmt.Errorf("rlm: stage input: %w", err)
	}
	reqDesc := map[string]any{
		"request":   a.Request,
		"spec":      req.Spec,
		"in_table":  req.InTable,
		"out_table": req.OutTable,
	}
	rb, _ := json.MarshalIndent(reqDesc, "", "  ")
	if err := os.WriteFile(filepath.Join(inDir, "request.json"), rb, 0o600); err != nil {
		return "", "", 0, fmt.Errorf("rlm: stage request: %w", err)
	}
	return jobDir, fingerprint, len(records), nil
}

// readAgentOutput reads the agent's out/output.ndjson and asserts the row count
// matches out/manifest.json (when present) so a truncated/partial write is not
// silently accepted as success.
func readAgentOutput(jobDir string) ([]connectors.Record, error) {
	outPath := filepath.Join(jobDir, "out", "output.ndjson")
	f, err := os.Open(outPath)
	if err != nil {
		return nil, fmt.Errorf("rlm: read agent output: %w", err)
	}
	defer f.Close()

	var records []connectors.Record
	dec := json.NewDecoder(f)
	for {
		var rec connectors.Record
		if err := dec.Decode(&rec); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("rlm: parse agent output: %w", err)
		}
		records = append(records, rec)
	}

	if manifest := readManifest(jobDir); manifest != nil {
		if manifest.ExpectedCount != len(records) {
			return nil, fmt.Errorf("rlm: agent output truncated: got %d rows, manifest expected %d", len(records), manifest.ExpectedCount)
		}
	}
	return records, nil
}

type agentManifest struct {
	ExpectedCount int `json:"expected_count"`
	RecordsRead   int `json:"records_read"`
}

func readManifest(jobDir string) *agentManifest {
	b, err := os.ReadFile(filepath.Join(jobDir, "out", "manifest.json"))
	if err != nil {
		return nil
	}
	var m agentManifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return &m
}

func fingerprintBytes(parts ...[]byte) string {
	h := sha256.New()
	for _, p := range parts {
		// length-prefix each part so concatenation is unambiguous
		var lp [8]byte
		n := len(p)
		for i := 0; i < 8; i++ {
			lp[i] = byte(n >> (8 * i))
		}
		h.Write(lp[:])
		h.Write(p)
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func specName(s *Spec) string {
	if s == nil {
		return ""
	}
	return s.Name
}
