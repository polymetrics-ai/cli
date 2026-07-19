package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/rlm"
	"polymetrics.ai/internal/rlm/router"
	"polymetrics.ai/internal/safety"
)

type extractQueryFunc func(context.Context, string, string, int) ([]connectors.Record, error)

type extractCommandRuntime struct {
	prepare  func(string) error
	query    extractQueryFunc
	analyzer rlmAnalyzerFactory
}

type extractFlags struct {
	Requests    []string
	SQLs        []string
	Limits      []string
	Providers   []string
	Models      []string
	LLMBaseURLs []string
	Inputs      []string
	Outputs     []string
	SpecNames   []string
}

func defaultExtractCommandRuntime() extractCommandRuntime {
	return extractCommandRuntime{
		prepare: func(root string) error {
			return withApp(root, func(*app.App) error { return nil })
		},
		query: func(ctx context.Context, root, sql string, limit int) ([]connectors.Record, error) {
			var rows []connectors.Record
			err := withApp(root, func(a *app.App) error {
				var err error
				rows, err = a.QuerySQL(ctx, sql, limit)
				return err
			})
			return rows, err
		},
		analyzer: defaultRLMAnalyzer,
	}
}

func newExtractCobraCommand(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	return newExtractCobraCommandWithRuntime(ctx, cfg, root, stdout, jsonOut, defaultExtractCommandRuntime())
}

func newExtractCobraCommandWithRuntime(ctx context.Context, cfg config.Config, root string, stdout io.Writer, jsonOut bool, runtime extractCommandRuntime) *cobra.Command {
	var flags extractFlags
	cmd := &cobra.Command{
		Use:           "extract",
		Hidden:        true,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 && flags.empty() {
				return markCobraLegacyError(writeManual("extract", stdout, jsonOut))
			}
			if len(args) > 0 && isHelpArg(args[0]) {
				return markCobraLegacyError(writeManual("extract", stdout, jsonOut))
			}
			return markCobraLegacyError(runExtract(ctx, cfg, root, flags, stdout, jsonOut, runtime))
		},
	}
	setManualHelp(cmd, "extract", stdout, jsonOut)
	addExtractStringArrayFlag(cmd, &flags.Requests, "request", "natural-language routing request")
	addExtractStringArrayFlag(cmd, &flags.SQLs, "sql", "read-only SQL for a simple-query route")
	addExtractStringArrayFlag(cmd, &flags.Limits, "limit", "maximum query rows")
	addExtractStringArrayFlag(cmd, &flags.Providers, "provider", "optional LLM classifier provider")
	addExtractStringArrayFlag(cmd, &flags.Models, "model", "optional LLM classifier model")
	addExtractStringArrayFlag(cmd, &flags.LLMBaseURLs, "llm-base-url", "optional LLM classifier base URL")
	addExtractStringArrayFlag(cmd, &flags.Inputs, "in", "source warehouse table for RLM analysis")
	addExtractStringArrayFlag(cmd, &flags.Outputs, "out", "destination warehouse table for RLM analysis")
	addExtractStringArrayFlag(cmd, &flags.SpecNames, "spec-name", "RLM result specification name")
	cmd.AddCommand(newExtractHelpCobraCommand(stdout, jsonOut))
	return cmd
}

func newExtractHelpCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
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
			return markCobraLegacyError(writeManual("extract", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "extract", stdout, jsonOut)
	return cmd
}

func addExtractStringArrayFlag(cmd *cobra.Command, target *[]string, name, usage string) {
	cmd.Flags().StringArrayVar(target, name, nil, usage)
	if flag := cmd.Flags().Lookup(name); flag != nil {
		flag.NoOptDefVal = "true"
	}
}

func (f extractFlags) empty() bool {
	return len(f.Requests) == 0 && len(f.SQLs) == 0 && len(f.Limits) == 0 &&
		len(f.Providers) == 0 && len(f.Models) == 0 && len(f.LLMBaseURLs) == 0 &&
		len(f.Inputs) == 0 && len(f.Outputs) == 0 && len(f.SpecNames) == 0
}

func (f extractFlags) last(values []string) string {
	return lastString(values)
}

// runExtract is the narrow natural-language entry point. It routes a request
// to a read-only SQL query or to the typed RLM analyzer boundary.
func runExtract(ctx context.Context, cfg config.Config, root string, flags extractFlags, stdout io.Writer, jsonOut bool, runtime extractCommandRuntime) error {
	if runtime.prepare != nil {
		if err := runtime.prepare(root); err != nil {
			return err
		}
	}
	request := flags.last(flags.Requests)
	if request == "" {
		return usageErrorf("extract: --request is required")
	}
	limit, err := parseIntFlag("limit", valueOr(flags.last(flags.Limits), "100"), 100)
	if err != nil {
		return err
	}

	r := &router.Router{LLM: extractLLM(cfg, flags)} // nil LLM → heuristic-only
	decision := r.Classify(ctx, request, "")

	route := "simple_query"
	if decision.Task.IsRLM() {
		route = "rlm_analysis"
	}

	env := envelope{
		"kind":      "ExtractResult",
		"route":     route,
		"task_type": string(decision.Task),
		"decision":  decision,
	}

	if !decision.Task.IsRLM() {
		sql := valueOr(flags.last(flags.SQLs), decision.SuggestedSQL)
		if sql != "" {
			if runtime.query == nil {
				return errors.New("extract: query runtime is not configured")
			}
			rows, err := runtime.query(ctx, root, sql, limit)
			if err != nil {
				return fmt.Errorf("extract: query: %w", err)
			}
			env["rows"] = rows
			env["count"] = len(rows)
			env["sql"] = sql
		} else {
			env["note"] = "simple_query route, but no SQL was generated; pass --sql or configure an LLM provider for SQL synthesis"
		}
	} else {
		inTable, outTable := flags.last(flags.Inputs), flags.last(flags.Outputs)
		if inTable != "" && outTable != "" {
			if err := validateExtractTable(inTable, "input table"); err != nil {
				return err
			}
			if err := validateExtractTable(outTable, "output table"); err != nil {
				return err
			}
			warehouseDir := filepath.Join(root, ".polymetrics", "warehouse")
			if err := safety.ValidateLocalWritePath(root, warehouseDir, "extract warehouse path", false); err != nil {
				return validationErrorf("extract: %v", err)
			}
			if runtime.analyzer == nil {
				return errors.New("extract: analyzer runtime is not configured")
			}
			analyzer, closer, err := runtime.analyzer(ctx, cfg, "agent", request)
			if err != nil {
				env["note"] = fmt.Sprintf("rlm_analysis route; agent backend unavailable: %v", err)
			} else {
				if closer != nil {
					defer func() { _ = closer() }()
				}
				if analyzer == nil {
					return errors.New("extract: analyzer runtime returned nil")
				}
				runRequest := rlm.RunRequest{
					Spec:         &rlm.Spec{Name: valueOr(flags.last(flags.SpecNames), "extract")},
					InTable:      inTable,
					OutTable:     outTable,
					WarehouseDir: warehouseDir,
				}
				res, runErr := analyzer.Run(ctx, runRequest)
				if runErr != nil {
					return fmt.Errorf("extract: rlm agent: %w", runErr)
				}
				env["rlm"] = res
			}
		} else {
			env["note"] = "rlm_analysis route; provide --in and --out (and run `pm worker serve` or set POLYMETRICS_RLM_FAKE_RUNNER) to execute"
		}
	}

	if jsonOut {
		return writeJSON(stdout, env)
	}
	b, _ := json.MarshalIndent(env, "", "  ")
	fmt.Fprintln(stdout, string(b))
	return nil
}

func validateExtractTable(table, field string) error {
	if err := safety.ValidateIdentifier(table, field); err != nil {
		return validationErrorf("extract: %v", err)
	}
	return nil
}

// extractLLM resolves an optional Tier-2 classifier from typed flags/config. It
// returns nil when no provider is resolvable, preserving offline routing.
func extractLLM(invocation config.Config, flags extractFlags) router.LLMClassifier {
	cfg := rlm.LLMConfigFromSettings(invocation.RLM.LLM.Provider, invocation.RLM.LLM.BaseURL, invocation.RLM.LLM.Model, os.Getenv)
	if p := flags.last(flags.Providers); p != "" {
		cfg.Provider = p
	}
	if m := flags.last(flags.Models); m != "" {
		cfg.Model = m
	}
	if b := flags.last(flags.LLMBaseURLs); b != "" {
		cfg.BaseURL = b
	}
	if !cfg.Resolvable() || cfg.Model == "" {
		return nil
	}
	client, err := rlm.NewLLM(cfg)
	if err != nil {
		return nil
	}
	return client
}
