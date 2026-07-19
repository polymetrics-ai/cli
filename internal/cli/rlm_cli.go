package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/rlm"
	"polymetrics.ai/internal/temporalprobe"
	"polymetrics.ai/internal/worker"
)

var temporalProbe = temporalprobe.Probe
var workerSubmitterForActivities = worker.SubmitterForActivitiesContext

type rlmAnalyzerFactory func(context.Context, config.Config, string, string) (rlm.Analyzer, func() error, error)

type rlmCommandRuntime struct {
	analyzer rlmAnalyzerFactory
}

type rlmRunFlags struct {
	Specs    []string
	Inputs   []string
	Outputs  []string
	Modes    []string
	DryRuns  []string
	Requests []string
}

func defaultRLMCommandRuntime() rlmCommandRuntime {
	return rlmCommandRuntime{analyzer: defaultRLMAnalyzer}
}

func defaultRLMAnalyzer(ctx context.Context, cfg config.Config, mode, request string) (rlm.Analyzer, func() error, error) {
	switch mode {
	case "deterministic":
		return &rlm.DeterministicAnalyzer{}, nil, nil
	case "fixture":
		return &rlm.FixtureAnalyzer{}, nil, nil
	case "model":
		return &rlm.ModelAnalyzer{}, nil, nil
	case "agent":
		return buildAgentAnalyzer(ctx, cfg, request)
	default:
		return nil, nil, usageErrorf("rlm run: unknown mode %q (want deterministic|fixture|model|agent)", mode)
	}
}

func newRLMCobraCommand(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	return newRLMCobraCommandWithRuntime(ctx, cfg, root, stdout, jsonOut, defaultRLMCommandRuntime())
}

func newRLMCobraCommandWithRuntime(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool, runtime rlmCommandRuntime) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "rlm",
		Args:              cobra.ArbitraryArgs,
		SilenceErrors:     true,
		SilenceUsage:      true,
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return markCobraLegacyError(usageErrorf("rlm: unknown subcommand %q", args[0]))
			}
			return markCobraLegacyError(writeManual("rlm", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "rlm", stdout, jsonOut)
	cmd.AddCommand(newRLMRunCobraCommand(ctx, cfg, root, stdout, jsonOut, runtime))
	cmd.AddCommand(newRLMHelpCobraCommand(stdout, jsonOut))
	return cmd
}

func newRLMRunCobraCommand(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool, runtime rlmCommandRuntime) *cobra.Command {
	var flags rlmRunFlags
	cmd := &cobra.Command{
		Use:           "run",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(runRLMRun(ctx, cfg, root, flags, stdout, jsonOut, runtime))
		},
	}
	setManualHelp(cmd, "rlm", stdout, jsonOut)
	addRLMStringArrayFlag(cmd, &flags.Specs, "spec", "RLM scoring spec path")
	addRLMStringArrayFlag(cmd, &flags.Inputs, "in", "source warehouse table")
	addRLMStringArrayFlag(cmd, &flags.Outputs, "out", "destination warehouse table")
	addRLMStringArrayFlag(cmd, &flags.Modes, "mode", "RLM mode: deterministic, fixture, model, or agent")
	addRLMStringArrayFlag(cmd, &flags.DryRuns, "dry-run", "score records without writing the destination table")
	addRLMStringArrayFlag(cmd, &flags.Requests, "request", "bounded request for agent mode")
	return cmd
}

func newRLMHelpCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "help",
		Hidden:        true,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, _ []string) error {
			return markCobraLegacyError(writeManual("rlm", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "rlm", stdout, jsonOut)
	return cmd
}

func addRLMStringArrayFlag(cmd *cobra.Command, target *[]string, name, usage string) {
	cmd.Flags().StringArrayVar(target, name, nil, usage)
	if flag := cmd.Flags().Lookup(name); flag != nil {
		flag.NoOptDefVal = "true"
	}
}

func runRLMRun(ctx context.Context, cfg config.Config, root string, flags rlmRunFlags, stdout io.Writer, jsonOut bool, runtime rlmCommandRuntime) error {
	specPath := lastString(flags.Specs)
	inTable := lastString(flags.Inputs)
	outTable := lastString(flags.Outputs)
	mode := lastString(flags.Modes)
	dryRun := lastString(flags.DryRuns) == "true"

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
	if !isRLMMode(mode) {
		return usageErrorf("rlm run: unknown mode %q (want deterministic|fixture|model|agent)", mode)
	}
	if runtime.analyzer == nil {
		return errors.New("rlm: analyzer factory is not configured")
	}

	factoryRequest := ""
	if mode == "agent" {
		factoryRequest = lastString(flags.Requests)
	}
	analyzer, closer, err := runtime.analyzer(ctx, cfg, mode, factoryRequest)
	if err != nil {
		return err
	}
	if analyzer == nil {
		return errors.New("rlm: analyzer factory returned nil")
	}
	if closer != nil {
		defer func() { _ = closer() }()
	}

	req := rlm.RunRequest{
		Spec:         spec,
		InTable:      inTable,
		OutTable:     outTable,
		WarehouseDir: filepath.Join(root, ".polymetrics", "warehouse"),
		DryRun:       dryRun,
	}
	result, err := analyzer.Run(ctx, req)
	if err != nil {
		return fmt.Errorf("rlm: run: %w", err)
	}
	if jsonOut {
		return json.NewEncoder(stdout).Encode(result)
	}
	fmt.Fprintf(stdout, "mode=%s in=%s out=%s records_read=%d records_scored=%d dry_run=%v duration=%s\n",
		result.Mode, result.InTable, result.OutTable,
		result.RecordsRead, result.RecordsScored, result.DryRun, result.Duration)
	return nil
}

func isRLMMode(mode string) bool {
	switch mode {
	case "deterministic", "fixture", "model", "agent":
		return true
	default:
		return false
	}
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
	probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if !temporalProbe(probeCtx, agentCfg.TemporalAddr) {
		return nil, nil, rlm.ErrRemoteUnavailable
	}
	embedded := cfg.RLM.EmbeddedWorker
	submit, closer, err := workerSubmitterForActivities(ctx, agentCfg.TemporalAddr, embedded, worker.NewPodmanActivities(agentCfg.PodmanBin, agentCfg.Image))
	if err != nil {
		return nil, nil, fmt.Errorf("rlm: %w (%v)", rlm.ErrRemoteUnavailable, err)
	}
	a := &rlm.AgentAnalyzer{
		Cfg:     agentCfg,
		Probe:   temporalProbe,
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
