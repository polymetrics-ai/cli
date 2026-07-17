package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/rlm"
	"polymetrics.ai/internal/temporalprobe"
	"polymetrics.ai/internal/worker"
)

func runRLM(ctx context.Context, cfg config.Config, root string, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return usageErrorf("rlm: missing subcommand (try: rlm run)")
	}
	switch args[0] {
	case "run":
		return runRLMRun(ctx, cfg, root, args[1:], stdout, jsonOut)
	default:
		return usageErrorf("rlm: unknown subcommand %q", args[0])
	}
}

func runRLMRun(ctx context.Context, cfg config.Config, root string, args []string, stdout io.Writer, jsonOut bool) error {
	flags := parseFlags(args)

	specPath := flags.first("spec")
	inTable := flags.first("in")
	outTable := flags.first("out")
	mode := flags.first("mode")
	dryRun := flags.first("dry-run") == "true"

	if specPath == "" {
		return usageErrorf("rlm run: --spec is required")
	}
	if outTable == "" {
		return usageErrorf("rlm run: --out is required")
	}
	if mode == "" {
		return usageErrorf("rlm run: --mode is required (deterministic|fixture|model|agent)")
	}

	specData, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("rlm: read spec %q: %w", specPath, err)
	}

	spec, err := rlm.ParseSpec(specData)
	if err != nil {
		return fmt.Errorf("rlm: parse spec: %w", err)
	}

	warehouseDir := filepath.Join(root, ".polymetrics", "warehouse")

	req := rlm.RunRequest{
		Spec:         spec,
		InTable:      inTable,
		OutTable:     outTable,
		WarehouseDir: warehouseDir,
		DryRun:       dryRun,
	}

	var analyzer rlm.Analyzer
	var closer func() error
	switch mode {
	case "deterministic":
		analyzer = &rlm.DeterministicAnalyzer{}
	case "fixture":
		analyzer = &rlm.FixtureAnalyzer{}
	case "model":
		analyzer = &rlm.ModelAnalyzer{}
	case "agent":
		a, c, err := buildAgentAnalyzer(ctx, cfg, flags.first("request"))
		if err != nil {
			return err
		}
		analyzer, closer = a, c
	default:
		return usageErrorf("rlm run: unknown mode %q (want deterministic|fixture|model|agent)", mode)
	}
	if closer != nil {
		defer closer()
	}

	result, err := analyzer.Run(ctx, req)
	if err != nil {
		return fmt.Errorf("rlm: run: %w", err)
	}

	if jsonOut {
		enc := json.NewEncoder(stdout)
		return enc.Encode(result)
	}

	fmt.Fprintf(stdout, "mode=%s in=%s out=%s records_read=%d records_scored=%d dry_run=%v duration=%s\n",
		result.Mode, result.InTable, result.OutTable,
		result.RecordsRead, result.RecordsScored, result.DryRun, result.Duration)
	return nil
}

// buildAgentAnalyzer constructs the RLM agent backend. When rlm.fake_runner is
// set it runs fully offline (no Temporal/podman) — the hermetic dev/test path.
// Otherwise it wires the real Temporal submitter (daemon by default; embedded
// with rlm.embedded_worker=true) and probe.
func buildAgentAnalyzer(ctx context.Context, cfg config.Config, request string) (rlm.Analyzer, func() error, error) {
	agentCfg := agentConfigFromConfig(cfg)

	if cfg.RLM.FakeRunner {
		a := &rlm.AgentAnalyzer{
			Cfg:      rlm.AgentConfig{TemporalAddr: "fake", PodmanBin: "fake", Image: agentCfg.Image, MaxIter: agentCfg.MaxIter},
			Probe:    func(context.Context, string) bool { return true },
			LookPath: func(string) (string, error) { return "fake", nil },
			Submit:   rlm.NewFakeRunnerSubmit(),
			Request:  request,
		}
		return a, nil, nil
	}

	if agentCfg.TemporalAddr == "" {
		return nil, nil, rlm.ErrRemoteUnavailable
	}
	embedded := cfg.RLM.EmbeddedWorker
	submit, closer, err := worker.SubmitterForActivitiesContext(ctx, agentCfg.TemporalAddr, embedded, worker.NewPodmanActivities(agentCfg.PodmanBin, agentCfg.Image))
	if err != nil {
		return nil, nil, fmt.Errorf("rlm: %w (%v)", rlm.ErrRemoteUnavailable, err)
	}
	a := &rlm.AgentAnalyzer{
		Cfg:     agentCfg,
		Probe:   temporalprobe.Probe,
		Submit:  submit,
		Request: request,
	}
	return a, closer, nil
}

func agentConfigFromConfig(cfg config.Config) rlm.AgentConfig {
	agentCfg := rlm.AgentConfig{
		TemporalAddr: explicitTemporalAddr(cfg),
		Image:        cfg.RLM.Image,
		PodmanBin:    cfg.RLM.PodmanBin,
		MaxIter:      4,
	}
	if agentCfg.PodmanBin == "" {
		agentCfg.PodmanBin = "podman"
	}
	if agentCfg.Image == "" {
		agentCfg.Image = "ghcr.io/polymetrics/rlm-agent:latest"
	}
	return agentCfg
}
