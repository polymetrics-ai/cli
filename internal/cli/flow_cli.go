package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/flow"
	"polymetrics.ai/internal/rlm"
)

// runFlow dispatches pm flow subcommands: plan | preview | run | status | list.
func runFlow(ctx context.Context, cfg config.Config, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return usageErrorf("flow: subcommand required (plan|preview|run|status|list)")
	}

	sub := args[0]
	rest := args[1:]

	switch sub {
	case "plan":
		return flowPlan(ctx, rest, stdout, jsonOut, false)
	case "preview":
		return flowPlan(ctx, rest, stdout, jsonOut, true)
	case "run":
		return flowRun(ctx, cfg, a, rest, stdout, jsonOut)
	case "status":
		return flowStatus(rest, stdout, jsonOut)
	case "list":
		return flowList(rest, stdout, jsonOut)
	default:
		return usageErrorf("flow: unknown subcommand %q", sub)
	}
}

// parseFlowFlags extracts --file, --force, --flows-dir from args.
func parseFlowFlags(args []string) (file, flowsDir string, force bool, positional []string) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--file":
			if i+1 < len(args) {
				i++
				file = args[i]
			}
		case "--flows-dir":
			if i+1 < len(args) {
				i++
				flowsDir = args[i]
			}
		case "--force":
			force = true
		default:
			if !strings.HasPrefix(args[i], "--") {
				positional = append(positional, args[i])
			}
		}
	}
	return
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
func flowPlan(_ context.Context, args []string, stdout io.Writer, jsonOut bool, dryRun bool) error {
	file, _, _, _ := parseFlowFlags(args)
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
	fmt.Fprintf(stdout, "Flow: %s  status=%s\n", m.Name, status)
	for i, id := range order {
		fmt.Fprintf(stdout, "  %d. %s\n", i+1, id)
	}
	return nil
}

// flowRun executes the flow.
func flowRun(ctx context.Context, cfg config.Config, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
	file, flowsDir, force, positional := parseFlowFlags(args)
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

	result, err := e.Run(ctx, flow.RunOptions{Force: force})
	if err != nil {
		return err
	}

	if jsonOut {
		return writeJSON(stdout, result)
	}
	fmt.Fprintf(stdout, "Flow %s: %s\n", result.FlowName, result.Status)
	return nil
}

// flowStatus returns last checkpoint info for a named flow.
func flowStatus(args []string, stdout io.Writer, jsonOut bool) error {
	_, flowsDir, _, positional := parseFlowFlags(args)
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
	fmt.Fprintf(stdout, "Flow: %s\n", name)
	for _, s := range statuses {
		fmt.Fprintf(stdout, "  %s: %s\n", s.ID, s.Status)
	}
	return nil
}

// flowList lists flow manifest files in the flows directory.
func flowList(args []string, stdout io.Writer, jsonOut bool) error {
	_, flowsDir, _, _ := parseFlowFlags(args)
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
		fmt.Fprintln(stdout, f)
	}
	return nil
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

	analyzer, closer, err := flowRLMAnalyzer(a.cfg, req.Mode)
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

func flowRLMAnalyzer(cfg config.Config, mode string) (rlm.Analyzer, func() error, error) {
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
		return buildAgentAnalyzer(cfg, "")
	default:
		return nil, nil, usageErrorf("flow rlm: unknown mode %q (want deterministic|fixture|model|agent)", mode)
	}
}
