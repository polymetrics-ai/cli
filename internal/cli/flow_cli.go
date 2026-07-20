package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/events"
	"polymetrics.ai/internal/flow"
	"polymetrics.ai/internal/rlm"
	"polymetrics.ai/internal/safety"
	pmui "polymetrics.ai/internal/ui"
	rundashboard "polymetrics.ai/internal/ui/run"
)

type flowFileFlags struct {
	Files []string
}

type flowRunFlags struct {
	Files     []string
	FlowsDirs []string
	Force     bool
}

type flowDirFlags struct {
	FlowsDirs []string
}

func newFlowCobraCommand(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "flow",
		Args:              cobra.ArbitraryArgs,
		SilenceErrors:     true,
		SilenceUsage:      true,
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return markCobraLegacyError(usageErrorf("flow: unknown subcommand %q", args[0]))
			}
			return markCobraLegacyError(writeManual("flow", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "flow", stdout, jsonOut)
	cmd.AddCommand(newFlowPlanCobraCommand(ctx, root, stdout, jsonOut, false))
	cmd.AddCommand(newFlowPlanCobraCommand(ctx, root, stdout, jsonOut, true))
	cmd.AddCommand(newFlowRunCobraCommand(ctx, cfg, root, stdout, jsonOut))
	cmd.AddCommand(newFlowStatusCobraCommand(root, stdout, jsonOut))
	cmd.AddCommand(newFlowListCobraCommand(root, stdout, jsonOut))
	cmd.AddCommand(newFlowHelpCobraCommand(stdout, jsonOut))
	return cmd
}

func newFlowPlanCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut, dryRun bool) *cobra.Command {
	var flags flowFileFlags
	name := "plan"
	if dryRun {
		name = "preview"
	}
	cmd := newFlowActionCobraCommand(name, func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(withApp(root, func(_ *app.App) error {
			return flowPlan(ctx, lastString(flags.Files), stdout, jsonOut, dryRun)
		}))
	})
	setManualHelp(cmd, "flow", stdout, jsonOut)
	addFlowStringArrayFlag(cmd, &flags.Files, "file", "flow manifest path")
	return cmd
}

func newFlowRunCobraCommand(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags flowRunFlags
	cmd := newFlowActionCobraCommand("run", func(cmd *cobra.Command, args []string) error {
		if operand, ok := flowOperand(cmd); ok {
			args = []string{operand}
		}
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return flowRun(ctx, cfg, a, flags, args, stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "flow", stdout, jsonOut)
	addFlowStringArrayFlag(cmd, &flags.Files, "file", "flow manifest path")
	addFlowStringArrayFlag(cmd, &flags.FlowsDirs, "flows-dir", "directory containing named flow manifests")
	cmd.Flags().BoolVar(&flags.Force, "force", false, "clear checkpoints before running")
	return cmd
}

func newFlowStatusCobraCommand(root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags flowDirFlags
	cmd := newFlowActionCobraCommand("status", func(cmd *cobra.Command, args []string) error {
		if operand, ok := flowOperand(cmd); ok {
			args = []string{operand}
		}
		return markCobraLegacyError(withApp(root, func(_ *app.App) error {
			return flowStatus(lastString(flags.FlowsDirs), args, stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "flow", stdout, jsonOut)
	addFlowStringArrayFlag(cmd, &flags.FlowsDirs, "flows-dir", "directory containing the flow manifest and checkpoints")
	return cmd
}

func newFlowListCobraCommand(root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags flowDirFlags
	cmd := newFlowActionCobraCommand("list", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(withApp(root, func(_ *app.App) error {
			return flowList(lastString(flags.FlowsDirs), stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "flow", stdout, jsonOut)
	addFlowStringArrayFlag(cmd, &flags.FlowsDirs, "flows-dir", "directory containing flow manifests")
	return cmd
}

func newFlowHelpCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newFlowActionCobraCommand("help", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(writeManual("flow", stdout, jsonOut))
	})
	cmd.Hidden = true
	setManualHelp(cmd, "flow", stdout, jsonOut)
	return cmd
}

func newFlowActionCobraCommand(use string, run func(*cobra.Command, []string) error) *cobra.Command {
	return &cobra.Command{
		Use:           use,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE:              run,
	}
}

func addFlowStringArrayFlag(cmd *cobra.Command, target *[]string, name, usage string) {
	cmd.Flags().StringArrayVar(target, name, nil, usage)
	if flag := cmd.Flags().Lookup(name); flag != nil {
		flag.NoOptDefVal = "true"
	}
}

func flowOperand(cmd *cobra.Command) (string, bool) {
	state, ok := cmd.Context().Value(flowCommandStateKey{}).(flowCommandState)
	if !ok || !state.operandSet {
		return "", false
	}
	return state.operand, true
}

func readManifestFile(path string) (flow.FlowManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return flow.FlowManifest{}, fmt.Errorf("flow: cannot read manifest %q: %w", path, err)
	}
	m, err := flow.ParseManifest(data)
	if err != nil {
		return flow.FlowManifest{}, fmt.Errorf("flow: manifest parse error: %w", err)
	}
	resolveManifestPaths(filepath.Dir(path), &m)
	if errs := flow.ValidateManifest(m); len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = e.Error()
		}
		return flow.FlowManifest{}, fmt.Errorf("flow: manifest validation failed: %s", strings.Join(msgs, "; "))
	}
	return m, nil
}

func resolveManifestPaths(baseDir string, m *flow.FlowManifest) {
	if baseDir == "" {
		baseDir = "."
	}
	for i := range m.Steps {
		step := &m.Steps[i]
		if step.Kind != flow.KindRLM || step.Spec == "" || filepath.IsAbs(step.Spec) {
			continue
		}
		step.Spec = filepath.Clean(filepath.Join(baseDir, step.Spec))
	}
}

// flowPlan parses + validates the manifest and prints the DAG order.
// dryRun=true makes it behave as "preview".
func flowPlan(_ context.Context, file string, stdout io.Writer, jsonOut bool, dryRun bool) error {
	if file == "" {
		return usageErrorf("flow plan: --file <path> is required")
	}

	m, err := readManifestFile(file)
	if err != nil {
		return err
	}

	order, err := flow.BuildDAG(m)
	if err != nil {
		return err
	}

	status := "ok"
	if dryRun {
		status = "dry_run"
	}

	if jsonOut {
		return writeJSON(stdout, envelope{
			"status": status,
			"flow":   m.Name,
			"order":  order,
		})
	}
	fmt.Fprintf(stdout, "Flow: %s  status=%s\n", humanFlowField(m.Name), humanFlowField(status))
	for i, id := range order {
		fmt.Fprintf(stdout, "  %d. %s\n", i+1, humanFlowField(id))
	}
	return nil
}

// flowRun executes the flow.
func flowRun(ctx context.Context, cfg config.Config, a *app.App, flags flowRunFlags, positional []string, stdout io.Writer, jsonOut bool) error {
	file := lastString(flags.Files)
	flowsDir := lastString(flags.FlowsDirs)
	if file == "" {
		if len(positional) == 0 {
			return usageErrorf("flow run: --file <path> or <flow-name> is required")
		}
		if flowsDir == "" {
			if a != nil {
				flowsDir = filepath.Join(a.ProjectDir(), "flows")
			} else {
				flowsDir = filepath.Join(".polymetrics", "flows")
			}
		}
		file = filepath.Join(flowsDir, positional[0])
		if filepath.Ext(file) == "" {
			file += ".json"
		}
	}

	m, err := readManifestFile(file)
	if err != nil {
		return err
	}

	// Build a no-op adapter when app is nil (testing without real app).
	var adapter flow.AppAdapter
	if a != nil {
		adapter = &appFlowAdapter{app: a, cfg: cfg}
	} else {
		adapter = &noopAppAdapter{}
	}

	dir := os.TempDir()
	if a != nil {
		dir = a.ProjectDir()
	}
	cs := &flow.FileCheckpointStore{Dir: dir}

	e := &flow.Engine{
		Manifest:   m,
		App:        adapter,
		Checkpoint: cs,
		LockDir:    dir,
	}

	if uiDetectionFromContext(ctx).Mode == pmui.ModeTUI {
		return flowRunDashboard(ctx, e, m, stdout, flags.Force)
	}

	result, err := e.Run(ctx, flow.RunOptions{Force: flags.Force})
	if err != nil {
		return err
	}

	if jsonOut {
		return writeJSON(stdout, result)
	}
	fmt.Fprintf(stdout, "Flow %s: %s\n", humanFlowField(result.FlowName), humanFlowField(result.Status))
	return nil
}

func flowRunDashboard(ctx context.Context, engine *flow.Engine, manifest flow.FlowManifest, stdout io.Writer, force bool) error {
	detection := uiDetectionFromContext(ctx)
	width, height := dashboardDimensions(stdout)
	session := rundashboard.NewSession(ctx, rundashboard.SessionOptions{
		Config: rundashboard.Config{
			Title:         "Flow",
			Name:          manifest.Name,
			Steps:         flowDashboardSteps(manifest),
			Width:         width,
			Height:        height,
			ASCII:         detection.ASCII,
			NoColor:       !detection.Color,
			StartedAt:     time.Now(),
			ResumeCommand: fmt.Sprintf("pm flow run %s", humanFlowField(manifest.Name)),
		},
		Upstream: events.FromContext(ctx),
		Interval: 100 * time.Millisecond,
		Input:    uiInputFromContext(ctx),
		Output:   stdout,
	})
	err := session.Execute(func(runCtx context.Context) error {
		_, runErr := engine.Run(runCtx, flow.RunOptions{Force: force})
		return runErr
	})
	return err
}

func flowDashboardSteps(manifest flow.FlowManifest) []rundashboard.Step {
	steps := make([]rundashboard.Step, 0, len(manifest.Steps))
	for _, step := range manifest.Steps {
		detail := ""
		switch step.Kind {
		case flow.KindSync:
			detail = strings.Join(step.Streams, ", ")
		case flow.KindQuery:
			detail = "query"
		case flow.KindRLM:
			detail = filepath.Base(step.Spec)
		case flow.KindAction:
			if step.ActionCfg != nil {
				detail = step.ActionCfg.SourceTable
			}
		}
		steps = append(steps, rundashboard.Step{ID: step.ID, Kind: string(step.Kind), Detail: detail})
	}
	return steps
}

// flowStatus returns last checkpoint info for a named flow.
func flowStatus(flowsDir string, positional []string, stdout io.Writer, jsonOut bool) error {
	if len(positional) == 0 {
		return usageErrorf("flow status: flow name required")
	}
	name := positional[0]

	if flowsDir == "" {
		flowsDir = os.TempDir()
	}

	cs := &flow.FileCheckpointStore{Dir: flowsDir}
	// There's no "list all steps" without the manifest; return a not-found error.
	// Try to read a manifest from the flows dir to enumerate steps.
	manifestPath := filepath.Join(flowsDir, name+".json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("flow status: flow %q not found in %s", name, flowsDir)
	}

	m, err := flow.ParseManifest(data)
	if err != nil {
		return err
	}

	type stepStatus struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	statuses := []stepStatus{}
	for _, s := range m.Steps {
		st, _ := cs.Get(name, s.ID)
		statuses = append(statuses, stepStatus{ID: s.ID, Status: st})
	}

	if jsonOut {
		return writeJSON(stdout, envelope{"flow": name, "steps": statuses})
	}
	fmt.Fprintf(stdout, "Flow: %s\n", humanFlowField(name))
	for _, s := range statuses {
		fmt.Fprintf(stdout, "  %s: %s\n", humanFlowField(s.ID), humanFlowField(s.Status))
	}
	return nil
}

// flowList lists flow manifest files in the flows directory.
func flowList(flowsDir string, stdout io.Writer, jsonOut bool) error {
	if flowsDir == "" {
		flowsDir = ".polymetrics/flows"
	}

	entries, err := os.ReadDir(flowsDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var flows []string
	for _, e := range entries {
		if !e.IsDir() && (strings.HasSuffix(e.Name(), ".json") || strings.HasSuffix(e.Name(), ".yaml")) {
			flows = append(flows, strings.TrimSuffix(strings.TrimSuffix(e.Name(), ".json"), ".yaml"))
		}
	}
	if flows == nil {
		flows = []string{}
	}

	if jsonOut {
		data, _ := json.Marshal(envelope{"flows": flows})
		_, err = stdout.Write(data)
		return err
	}
	if len(flows) == 0 {
		fmt.Fprintln(stdout, "(no flows)")
		return nil
	}
	for _, f := range flows {
		fmt.Fprintln(stdout, humanFlowField(f))
	}
	return nil
}

func humanFlowField(value string) string {
	return safety.SanitizeTerminalLine(value)
}

// noopAppAdapter satisfies AppAdapter for plan/preview (no real app needed).
type noopAppAdapter struct{}

func (n *noopAppAdapter) ETLRun(_ context.Context, _ string, _ []string) (flow.ETLResult, error) {
	return flow.ETLResult{}, nil
}
func (n *noopAppAdapter) QuerySQL(_ context.Context, _ string, _ int) ([]map[string]any, error) {
	return nil, nil
}
func (n *noopAppAdapter) RLMRun(_ context.Context, _ flow.RLMRunRequest) (flow.RLMResult, error) {
	return flow.RLMResult{}, nil
}

// appFlowAdapter wraps *app.App to satisfy AppAdapter.
type appFlowAdapter struct {
	app *app.App
	cfg config.Config
}

func (a *appFlowAdapter) ETLRun(ctx context.Context, connectionID string, streams []string) (flow.ETLResult, error) {
	result := flow.ETLResult{}
	for _, stream := range streams {
		run, err := a.app.RunETL(ctx, app.RunETLRequest{Connection: connectionID, Stream: stream})
		if err != nil {
			return result, err
		}
		result.RecordsRead += run.RecordsRead
		result.RecordsWritten += run.RecordsLoaded
	}
	return result, nil
}

func (a *appFlowAdapter) QuerySQL(ctx context.Context, sql string, limit int) ([]map[string]any, error) {
	records, err := a.app.QuerySQL(ctx, sql, limit)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, len(records))
	for i, r := range records {
		out[i] = map[string]any(r)
	}
	return out, nil
}

func (a *appFlowAdapter) RLMRun(ctx context.Context, req flow.RLMRunRequest) (flow.RLMResult, error) {
	specData, err := os.ReadFile(req.Spec)
	if err != nil {
		return flow.RLMResult{}, fmt.Errorf("flow rlm: read spec %q: %w", req.Spec, err)
	}
	spec, err := rlm.ParseSpec(specData)
	if err != nil {
		return flow.RLMResult{}, fmt.Errorf("flow rlm: parse spec: %w", err)
	}

	analyzer, closer, err := flowRLMAnalyzer(ctx, a.cfg, req.Mode)
	if err != nil {
		return flow.RLMResult{}, err
	}
	if closer != nil {
		defer closer()
	}

	result, err := analyzer.Run(ctx, rlm.RunRequest{
		Spec:         spec,
		InTable:      req.InTable,
		OutTable:     req.OutTable,
		WarehouseDir: filepath.Join(a.app.ProjectDir(), "warehouse"),
		DryRun:       req.DryRun,
	})
	if err != nil {
		return flow.RLMResult{}, fmt.Errorf("flow rlm: run: %w", err)
	}
	return flow.RLMResult{
		RecordsRead:   result.RecordsRead,
		RecordsScored: result.RecordsScored,
		RecordsFailed: result.RecordsFailed,
	}, nil
}

func flowRLMAnalyzer(ctx context.Context, cfg config.Config, mode string) (rlm.Analyzer, func() error, error) {
	if mode == "" {
		mode = "deterministic"
	}
	switch mode {
	case "deterministic":
		return &rlm.DeterministicAnalyzer{}, nil, nil
	case "fixture":
		return &rlm.FixtureAnalyzer{}, nil, nil
	case "model":
		return &rlm.ModelAnalyzer{}, nil, nil
	case "agent":
		return buildAgentAnalyzer(ctx, cfg, "")
	default:
		return nil, nil, usageErrorf("flow rlm: unknown mode %q (want deterministic|fixture|model|agent)", mode)
	}
}
