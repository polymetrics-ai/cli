package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/rlm"
	"polymetrics.ai/internal/rlm/router"
)

// runExtract is the natural-language entry point an LLM caller (terminal /
// Claude Code) uses: `pm extract --request "..."`. It routes the request to a
// simple SQL query or the RLM agent, then returns one ExtractResult envelope.
//
// Routing is provider-agnostic: when an LLM credential is resolvable the router
// uses it (Tier 2) and otherwise falls back to the offline keyword heuristic
// (Tier 1). The RLM execution branch is wired in a later slice; until then a
// data_analysis/ml route returns its decision so the caller can escalate.
func runExtract(ctx context.Context, a *app.App, root string, args []string, stdout io.Writer, jsonOut bool) error {
	flags := parseFlags(args)
	request := flags.first("request")
	if request == "" {
		return usageErrorf("extract: --request is required")
	}
	limit, err := parseIntFlag("limit", valueOr(flags.first("limit"), "100"), 100)
	if err != nil {
		return err
	}

	r := &router.Router{LLM: extractLLM(flags)} // nil LLM → heuristic-only
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
		// Simple path: run the SQL the router suggested, or one the caller passed
		// explicitly via --sql. Always validated by the query engine.
		sql := valueOr(flags.first("sql"), decision.SuggestedSQL)
		if sql != "" {
			rows, err := a.QuerySQL(ctx, sql, limit)
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
		// RLM path: execute the agent when both tables are provided and the
		// backend (fake runner, or Temporal+podman) is available; otherwise
		// return the decision so the caller can supply tables or escalate.
		inTable, outTable := flags.first("in"), flags.first("out")
		if inTable != "" && outTable != "" {
			analyzer, closer, err := buildAgentAnalyzer(request)
			if err != nil {
				env["note"] = fmt.Sprintf("rlm_analysis route; agent backend unavailable: %v", err)
			} else {
				if closer != nil {
					defer closer()
				}
				rlmReq := rlm.RunRequest{
					Spec:         &rlm.Spec{Name: valueOr(flags.first("spec-name"), "extract")},
					InTable:      inTable,
					OutTable:     outTable,
					WarehouseDir: filepath.Join(root, ".polymetrics", "warehouse"),
				}
				res, runErr := analyzer.Run(ctx, rlmReq)
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

// extractLLM resolves an optional Tier-2 classifier from flags/env. It returns
// nil (heuristic-only) when no LLM provider is resolvable, so routing always
// works offline. --provider/--model/--llm-base-url override the env config.
func extractLLM(flags parsedFlags) router.LLMClassifier {
	cfg := rlm.LLMConfigFromEnv(os.Getenv)
	if p := flags.first("provider"); p != "" {
		cfg.Provider = p
	}
	if m := flags.first("model"); m != "" {
		cfg.Model = m
	}
	if b := flags.first("llm-base-url"); b != "" {
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
