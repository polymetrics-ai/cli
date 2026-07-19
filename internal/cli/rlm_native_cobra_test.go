package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/rlm"
)

// This compile-time reference is intentional TDD evidence: the focused test
// checkpoint must fail until RLM has a native Cobra constructor.
var _ = newRLMCobraCommand

func TestRLMCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	rlmCmd := findCobraCommand(root, "rlm")
	if rlmCmd == nil {
		t.Fatal("missing rlm command")
	}
	if rlmCmd.DisableFlagParsing {
		t.Fatal("rlm command must use native Cobra flag parsing")
	}
	assertRLMNoFileCompletion(t, rlmCmd)

	run := findCobraCommand(rlmCmd, "run")
	if run == nil {
		t.Fatal("missing rlm run command")
	}
	if run.DisableFlagParsing {
		t.Fatal("rlm run must use native Cobra flag parsing")
	}
	if !run.FParseErrWhitelist.UnknownFlags {
		t.Fatal("rlm run must preserve unknown-flag tolerance")
	}
	assertRLMNoFileCompletion(t, run)
	for _, name := range []string{"spec", "in", "out", "mode", "dry-run", "request"} {
		flag := run.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("rlm run missing --%s", name)
		}
		if got, want := flag.Value.Type(), "stringArray"; got != want {
			t.Fatalf("rlm run --%s type = %q, want %q", name, got, want)
		}
		if got, want := flag.NoOptDefVal, "true"; got != want {
			t.Fatalf("rlm run --%s NoOptDefVal = %q, want %q", name, got, want)
		}
	}

	help := findCobraCommand(rlmCmd, "help")
	if help == nil || !help.Hidden {
		t.Fatal("rlm must preserve hidden positional help until Phase 19")
	}
	for _, name := range []string{"exec", "shell", "model", "viewer", "dashboard"} {
		if findCobraCommand(rlmCmd, name) != nil {
			t.Fatalf("generic or out-of-scope RLM action %q was registered", name)
		}
	}
	for _, spec := range cobraLegacyCommands(config.Config{}) {
		if spec.name == "rlm" {
			t.Fatal("rlm remains registered as a legacy Cobra wrapper")
		}
	}
}

func TestRLMRunRoutesModesThroughInjectedAnalyzers(t *testing.T) {
	for _, mode := range []string{"deterministic", "fixture", "model", "agent"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			specPath := writeRLMNativeSpec(t)
			ctxKey := rlmNativeContextKey{}
			ctx := context.WithValue(context.Background(), ctxKey, "kept")
			harness := newRLMNativeHarness(root)
			harness.ctx = ctx

			stdout, err := harness.execute(true,
				"rlm", "run", "ignored-operand",
				"--spec", "ignored", "--spec="+specPath,
				"--in", "ignored-in", "--in=contacts",
				"--out", "ignored-out", "--out=scores",
				"--mode", "ignored-mode", "--mode="+mode,
				"--request", "ignored request", "--request=private request",
				"--dry-run=false", "--dry-run",
				"--unknown", "ignored", "--unknown=ignored", "--=x", "---x",
				"--help", "help", "--", "ignored-after-literal", "-h",
			)
			if err != nil {
				t.Fatalf("execute: %v; stdout=%s", err, stdout)
			}
			if len(harness.factoryCalls) != 1 {
				t.Fatalf("factory calls = %#v", harness.factoryCalls)
			}
			call := harness.factoryCalls[0]
			if call.mode != mode || call.request != "private request" {
				t.Fatalf("factory call = %#v, want mode=%q request routed", call, mode)
			}
			if call.contextValue != "kept" {
				t.Fatalf("factory context value = %v, want kept", call.contextValue)
			}
			if len(harness.analyzer.calls) != 1 {
				t.Fatalf("analyzer calls = %d, want 1", len(harness.analyzer.calls))
			}
			req := harness.analyzer.calls[0]
			if req.Spec == nil || req.Spec.Name != "native-test" || req.InTable != "contacts" || req.OutTable != "scores" {
				t.Fatalf("run request = %#v", req)
			}
			if req.WarehouseDir != filepath.Join(root, ".polymetrics", "warehouse") || !req.DryRun {
				t.Fatalf("run request warehouse/dry-run = %#v", req)
			}
			if harness.analyzer.contextValues[0] != "kept" {
				t.Fatalf("analyzer context = %v, want kept", harness.analyzer.contextValues[0])
			}
			if harness.closeCalls != 1 {
				t.Fatalf("close calls = %d, want 1", harness.closeCalls)
			}
			if strings.Contains(stdout, "private request") {
				t.Fatalf("request leaked to output: %s", stdout)
			}
			var result rlm.RunResult
			decodeOneJSON(t, stdout, &result)
			if result.Mode != mode || result.InTable != "contacts" || result.OutTable != "scores" || !result.DryRun {
				t.Fatalf("result = %#v", result)
			}
		})
	}
}

func TestRLMRunPreservesFlagAndOutputForms(t *testing.T) {
	root := t.TempDir()
	specPath := writeRLMNativeSpec(t)
	harness := newRLMNativeHarness(root)

	text, err := harness.execute(false,
		"rlm", "run",
		"--spec="+specPath,
		"--in=contacts",
		"--out=scores",
		"--mode=fixture",
		"--dry-run=anything",
		"--request",
	)
	if err != nil {
		t.Fatalf("text run: %v", err)
	}
	want := "mode=fixture in=contacts out=scores records_read=3 records_scored=2 dry_run=false duration=7ns\n"
	if text != want {
		t.Fatalf("text output = %q, want %q", text, want)
	}
	if harness.factoryCalls[0].request != "true" {
		t.Fatalf("bare --request = %q, want true", harness.factoryCalls[0].request)
	}

	harness = newRLMNativeHarness(root)
	jsonOut, err := harness.execute(true,
		"rlm", "run",
		"--spec", specPath,
		"--in", "contacts",
		"--out", "scores",
		"--mode", "fixture",
		"--dry-run", "false",
	)
	if err != nil {
		t.Fatalf("JSON run: %v", err)
	}
	var got rlm.RunResult
	decodeOneJSON(t, jsonOut, &got)
	if got.Mode != "fixture" || got.DryRun {
		t.Fatalf("JSON result = %#v", got)
	}
}

func TestRLMRunRejectsInvalidShapeBeforeAnalyzer(t *testing.T) {
	root := t.TempDir()
	specPath := writeRLMNativeSpec(t)

	tests := []struct {
		name     string
		args     []string
		category errorCategory
		message  string
	}{
		{name: "missing spec", args: []string{"rlm", "run", "--out=o", "--mode=fixture"}, category: categoryUsage, message: "--spec is required"},
		{name: "missing out", args: []string{"rlm", "run", "--spec=" + specPath, "--mode=fixture"}, category: categoryUsage, message: "--out is required"},
		{name: "missing mode", args: []string{"rlm", "run", "--spec=" + specPath, "--out=o"}, category: categoryUsage, message: "--mode is required"},
		{name: "unknown mode", args: []string{"rlm", "run", "--spec=" + specPath, "--out=o", "--mode=remote"}, category: categoryUsage, message: `unknown mode "remote"`},
		{name: "unknown action", args: []string{"rlm", "exec", "run", "--spec=" + specPath, "--out=o", "--mode=fixture"}, category: categoryUsage, message: `unknown subcommand "exec"`},
		{name: "generic shell", args: []string{"rlm", "shell", "run", "--spec=" + specPath, "--out=o", "--mode=fixture"}, category: categoryUsage, message: `unknown subcommand "shell"`},
		{name: "phase16 viewer", args: []string{"rlm", "viewer", "run", "--spec=" + specPath, "--out=o", "--mode=fixture"}, category: categoryUsage, message: `unknown subcommand "viewer"`},
		{name: "literal action", args: []string{"rlm", "--", "run", "--spec=" + specPath, "--out=o", "--mode=fixture"}, category: categoryUsage, message: `unknown subcommand "--"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			harness := newRLMNativeHarness(root)
			stdout, err := harness.execute(true, tt.args...)
			if err == nil {
				t.Fatalf("expected error; stdout=%s", stdout)
			}
			classified := classifyError(mapCobraErr(err))
			if classified.category != tt.category || !strings.Contains(classified.message, tt.message) {
				t.Fatalf("error = (%s, %q), want (%s, contains %q)", classified.category, classified.message, tt.category, tt.message)
			}
			if len(harness.factoryCalls) != 0 || len(harness.analyzer.calls) != 0 {
				t.Fatalf("invalid shape reached analyzer: factory=%#v runs=%#v", harness.factoryCalls, harness.analyzer.calls)
			}
		})
	}
}

func TestRLMRunPropagatesFactoryAnalyzerAndCloseErrors(t *testing.T) {
	root := t.TempDir()
	specPath := writeRLMNativeSpec(t)
	baseArgs := []string{"rlm", "run", "--spec=" + specPath, "--out=scores", "--mode=agent", "--request=private request"}

	t.Run("factory", func(t *testing.T) {
		harness := newRLMNativeHarness(root)
		harness.factoryErr = errors.New("factory unavailable")
		stdout, err := harness.execute(false, baseArgs...)
		if err == nil || !strings.Contains(err.Error(), "factory unavailable") {
			t.Fatalf("error = %v; stdout=%s", err, stdout)
		}
		if strings.Contains(stdout+err.Error(), "private request") {
			t.Fatalf("request leaked: stdout=%q err=%v", stdout, err)
		}
	})

	t.Run("analyzer", func(t *testing.T) {
		harness := newRLMNativeHarness(root)
		harness.analyzer.err = errors.New("scoring failed")
		_, err := harness.execute(false, baseArgs...)
		if err == nil || !strings.Contains(err.Error(), "rlm: run: scoring failed") {
			t.Fatalf("error = %v", err)
		}
		if harness.closeCalls != 1 {
			t.Fatalf("close calls = %d, want 1", harness.closeCalls)
		}
	})

	t.Run("close remains best effort", func(t *testing.T) {
		harness := newRLMNativeHarness(root)
		harness.closeErr = errors.New("close failed")
		stdout, err := harness.execute(false, baseArgs...)
		if err != nil {
			t.Fatalf("close changed legacy success: %v", err)
		}
		if stdout == "" || harness.closeCalls != 1 {
			t.Fatalf("stdout=%q close calls=%d", stdout, harness.closeCalls)
		}
	})
}

func TestRLMHelpRoutesAndTrailingHelpCompatibility(t *testing.T) {
	manual := docs["rlm"]
	for _, tt := range []struct {
		name string
		args []string
		json bool
	}{
		{name: "help topic", args: []string{"help", "rlm"}},
		{name: "bare", args: []string{"rlm"}},
		{name: "long", args: []string{"rlm", "--help"}},
		{name: "short", args: []string{"rlm", "-h"}},
		{name: "positional", args: []string{"rlm", "help"}},
		{name: "JSON bare", args: []string{"rlm", "--json"}, json: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(tt.args, &stdout, &stderr)
			if code != 0 || stderr.Len() != 0 {
				t.Fatalf("exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
			if tt.json {
				var env struct {
					Kind    string `json:"kind"`
					Command string `json:"command"`
					Manual  string `json:"manual"`
				}
				decodeOneJSON(t, stdout.String(), &env)
				if env.Kind != "CommandManual" || env.Command != "rlm" || env.Manual != manual {
					t.Fatalf("manual envelope = %#v", env)
				}
				return
			}
			if stdout.String() != manual {
				t.Fatalf("manual mismatch\ngot: %q\nwant: %q", stdout.String(), manual)
			}
		})
	}

	root := t.TempDir()
	specPath := writeRLMNativeSpec(t)
	for _, trailing := range [][]string{{"--help"}, {"-h"}, {"help"}, {"--"}} {
		harness := newRLMNativeHarness(root)
		args := []string{"rlm", "run", "--spec=" + specPath, "--out=scores", "--mode=fixture"}
		args = append(args, trailing...)
		stdout, err := harness.execute(false, args...)
		if err != nil || !strings.HasPrefix(stdout, "mode=fixture") || len(harness.analyzer.calls) != 1 {
			t.Fatalf("trailing %q: stdout=%q err=%v calls=%d", trailing, stdout, err, len(harness.analyzer.calls))
		}
	}
}

func TestRLMRunStdoutStderrAndRequestRedaction(t *testing.T) {
	root := t.TempDir()
	request := "private-request-must-not-leak"

	t.Run("text usage", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := Run([]string{"--root", root, "rlm", "run", "--request", request}, &stdout, &stderr)
		if code != 2 || stdout.Len() != 0 || stderr.Len() == 0 {
			t.Fatalf("exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if strings.Contains(stdout.String()+stderr.String(), request) {
			t.Fatal("request leaked through text error")
		}
	})

	t.Run("JSON usage", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := Run([]string{"--root", root, "rlm", "run", "--request", request, "--json=true"}, &stdout, &stderr)
		if code != 2 || stderr.Len() != 0 {
			t.Fatalf("exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
		if strings.Contains(stdout.String(), request) {
			t.Fatal("request leaked through JSON error")
		}
		var env map[string]any
		dec := json.NewDecoder(strings.NewReader(stdout.String()))
		if err := dec.Decode(&env); err != nil {
			t.Fatalf("decode JSON error: %v — %s", err, stdout.String())
		}
		if env["kind"] != "Error" || env["category"] != "usage_error" {
			t.Fatalf("error envelope = %#v", env)
		}
		if dec.Decode(&map[string]any{}) == nil {
			t.Fatal("stdout contains more than one JSON value")
		}
	})
}

func TestRLMRunMalformedSpecUsesInternalErrorAndNoAnalyzer(t *testing.T) {
	root := t.TempDir()
	badSpec := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(badSpec, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	harness := newRLMNativeHarness(root)
	_, err := harness.execute(false, "rlm", "run", "--spec="+badSpec, "--out=scores", "--mode=fixture")
	if err == nil {
		t.Fatal("expected malformed spec error")
	}
	classified := classifyError(mapCobraErr(err))
	if classified.category != categoryInternal || !strings.Contains(classified.message, "rlm: parse spec") {
		t.Fatalf("error = (%s, %q)", classified.category, classified.message)
	}
	if len(harness.factoryCalls) != 0 || len(harness.analyzer.calls) != 0 {
		t.Fatal("malformed spec reached analyzer")
	}
}

func assertRLMNoFileCompletion(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	if cmd.ValidArgsFunction == nil {
		t.Fatalf("%s missing completion seam", cmd.CommandPath())
	}
	values, directive := cmd.ValidArgsFunction(cmd, nil, "")
	if len(values) != 0 || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("%s completion = (%v, %v), want empty NoFileComp", cmd.CommandPath(), values, directive)
	}
}

type rlmNativeContextKey struct{}

type rlmNativeFactoryCall struct {
	mode         string
	request      string
	contextValue any
}

type rlmNativeAnalyzer struct {
	mode          string
	calls         []rlm.RunRequest
	contextValues []any
	err           error
}

func (a *rlmNativeAnalyzer) Run(ctx context.Context, req rlm.RunRequest) (rlm.RunResult, error) {
	a.calls = append(a.calls, req)
	a.contextValues = append(a.contextValues, ctx.Value(rlmNativeContextKey{}))
	if a.err != nil {
		return rlm.RunResult{}, a.err
	}
	return rlm.RunResult{
		Mode:          a.mode,
		InTable:       req.InTable,
		OutTable:      req.OutTable,
		RecordsRead:   3,
		RecordsScored: 2,
		RecordsFailed: 1,
		Duration:      7 * time.Nanosecond,
		DryRun:        req.DryRun,
	}, nil
}

func (a *rlmNativeAnalyzer) Mode() string { return a.mode }

type rlmNativeHarness struct {
	root         string
	ctx          context.Context
	analyzer     *rlmNativeAnalyzer
	factoryCalls []rlmNativeFactoryCall
	factoryErr   error
	closeCalls   int
	closeErr     error
}

func newRLMNativeHarness(root string) *rlmNativeHarness {
	return &rlmNativeHarness{
		root:     root,
		ctx:      context.Background(),
		analyzer: &rlmNativeAnalyzer{},
	}
}

func (h *rlmNativeHarness) runtime() rlmCommandRuntime {
	return rlmCommandRuntime{
		analyzer: func(ctx context.Context, _ config.Config, mode, request string) (rlm.Analyzer, func() error, error) {
			h.factoryCalls = append(h.factoryCalls, rlmNativeFactoryCall{
				mode:         mode,
				request:      request,
				contextValue: ctx.Value(rlmNativeContextKey{}),
			})
			if h.factoryErr != nil {
				return nil, nil, h.factoryErr
			}
			h.analyzer.mode = mode
			return h.analyzer, func() error {
				h.closeCalls++
				return h.closeErr
			}, nil
		},
	}
}

func (h *rlmNativeHarness) execute(jsonOut bool, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cfg := testRouterConfig(h.root, jsonOut)
	cmd := newRootCmdWithRLMRuntime(h.ctx, cfg, &stdout, &stderr, h.runtime())
	err := executeRootCmd(cmd, args)
	if stderr.Len() != 0 {
		return stdout.String(), errors.New(stderr.String())
	}
	return stdout.String(), err
}

func writeRLMNativeSpec(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "spec.json")
	data := []byte(`{"name":"native-test","features":[{"name":"email","weight":1,"score_if_set":1}]}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	return path
}
