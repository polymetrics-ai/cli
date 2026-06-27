package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"polymetrics.ai/internal/rlm"
)

func runRLM(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return usageErrorf("rlm: missing subcommand (try: rlm run)")
	}
	switch args[0] {
	case "run":
		return runRLMRun(ctx, root, args[1:], stdout, jsonOut)
	default:
		return usageErrorf("rlm: unknown subcommand %q", args[0])
	}
}

func runRLMRun(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
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
		return usageErrorf("rlm run: --mode is required (deterministic|fixture|model)")
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
	switch mode {
	case "deterministic":
		analyzer = &rlm.DeterministicAnalyzer{}
	case "fixture":
		analyzer = &rlm.FixtureAnalyzer{}
	case "model":
		analyzer = &rlm.ModelAnalyzer{}
	default:
		return usageErrorf("rlm run: unknown mode %q (want deterministic|fixture|model)", mode)
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
