package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/rlm"
)

// This compile-time reference is intentional TDD evidence: the focused test
// checkpoint must fail until extract has a native Cobra constructor.
var _ = newExtractCobraCommand

func TestExtractCommandIsHiddenNativeCobraCommand(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	extract := findCobraCommand(root, "extract")
	if extract == nil {
		t.Fatal("missing extract command")
	}
	if !extract.Hidden {
		t.Fatal("extract command must remain hidden")
	}
	if extract.DisableFlagParsing {
		t.Fatal("extract command must use native Cobra flag parsing")
	}
	assertExtractNoFileCompletion(t, extract)
	if !extract.FParseErrWhitelist.UnknownFlags {
		t.Fatal("extract must preserve unknown-flag tolerance")
	}
	for _, name := range []string{"request", "sql", "limit", "provider", "model", "llm-base-url", "in", "out", "spec-name"} {
		flag := extract.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("extract missing --%s", name)
		}
		if got, want := flag.Value.Type(), "stringArray"; got != want {
			t.Fatalf("extract --%s type = %q, want %q", name, got, want)
		}
		if got, want := flag.NoOptDefVal, "true"; got != want {
			t.Fatalf("extract --%s NoOptDefVal = %q, want %q", name, got, want)
		}
	}
	help := findCobraCommand(extract, "help")
	if help == nil || !help.Hidden {
		t.Fatal("extract must provide hidden positional help")
	}
	for _, name := range []string{"run", "query", "rlm", "shell", "write", "http", "sql"} {
		if findCobraCommand(extract, name) != nil {
			t.Fatalf("unapproved extract action %q was registered", name)
		}
	}
	for _, spec := range cobraLegacyCommands(config.Config{}) {
		if spec.name == "extract" {
			t.Fatal("extract remains registered as a legacy Cobra wrapper")
		}
	}
}

func TestExtractRoutesCurrentFlagsAndOperandsThroughInjectedRuntime(t *testing.T) {
	ctxKey := extractNativeContextKey{}
	ctx := context.WithValue(context.Background(), ctxKey, "kept")
	runtime := newFakeExtractRuntime()
	cfg := testRouterConfig(t.TempDir(), true)

	stdout, stderr, err := executeNativeExtract(ctx, cfg, runtime,
		"extract", "ignored-before-flags",
		"--request", "ignored request", "--request=predict which contacts will convert",
		"--in", "ignored_input", "--in=contacts",
		"--out", "ignored_output", "--out=scores",
		"--spec-name", "ignored_spec", "--spec-name=native-extract",
		"--unknown", "ignored", "--unknown=last", "ignored-after-flags",
	)
	if err != nil || stderr != "" {
		t.Fatalf("RLM route: err=%v stderr=%q stdout=%q", err, stderr, stdout)
	}
	if runtime.queryCalls != 0 || runtime.analyzerFactoryCalls != 1 || runtime.analyzerCalls != 1 {
		t.Fatalf("runtime calls: query=%d factory=%d analyzer=%d", runtime.queryCalls, runtime.analyzerFactoryCalls, runtime.analyzerCalls)
	}
	if runtime.factoryRequest != "predict which contacts will convert" || runtime.factoryContextValue != "kept" {
		t.Fatalf("factory request/context = %q/%v", runtime.factoryRequest, runtime.factoryContextValue)
	}
	if runtime.request.InTable != "contacts" || runtime.request.OutTable != "scores" || runtime.request.Spec == nil || runtime.request.Spec.Name != "native-extract" {
		t.Fatalf("RLM request = %#v", runtime.request)
	}
	if runtime.analyzerContextValue != "kept" || runtime.closeCalls != 1 {
		t.Fatalf("analyzer context=%v close calls=%d", runtime.analyzerContextValue, runtime.closeCalls)
	}
	var env map[string]any
	decodeOneJSON(t, stdout, &env)
	if env["kind"] != "ExtractResult" || env["route"] != "rlm_analysis" || env["task_type"] != "ml" || env["rlm"] == nil {
		t.Fatalf("RLM envelope = %#v", env)
	}

	runtime = newFakeExtractRuntime()
	stdout, stderr, err = executeNativeExtract(ctx, cfg, runtime,
		"extract", "--request=show active customers", "--sql", "ignored", "--sql=SELECT * FROM contacts", "--limit", "5", "trailing-operand",
	)
	if err != nil || stderr != "" {
		t.Fatalf("query route: err=%v stderr=%q stdout=%q", err, stderr, stdout)
	}
	if runtime.queryCalls != 1 || runtime.querySQL != "SELECT * FROM contacts" || runtime.queryLimit != 5 || runtime.queryContextValue != "kept" {
		t.Fatalf("query call = count:%d sql:%q limit:%d ctx:%v", runtime.queryCalls, runtime.querySQL, runtime.queryLimit, runtime.queryContextValue)
	}
	decodeOneJSON(t, stdout, &env)
	if env["route"] != "simple_query" || env["count"] != float64(1) || env["sql"] != "SELECT * FROM contacts" {
		t.Fatalf("query envelope = %#v", env)
	}
}

func TestExtractHelpRoutesAreContextualHiddenAndSideEffectFree(t *testing.T) {
	manual := docs["extract"]
	if manual == "" || !strings.Contains(manual, "narrow natural-language") || !strings.Contains(manual, "bare warehouse table") {
		t.Fatal("extract manual must document its narrow route and bounded table names")
	}
	tests := []struct {
		name string
		args []string
		json bool
	}{
		{name: "topic", args: []string{"help", "extract"}},
		{name: "bare", args: []string{"extract"}},
		{name: "long", args: []string{"extract", "--help"}},
		{name: "short", args: []string{"extract", "-h"}},
		{name: "positional", args: []string{"extract", "help"}},
		{name: "trailing long", args: []string{"extract", "--request=show customers", "--help"}},
		{name: "trailing short", args: []string{"extract", "--request=show customers", "-h"}},
		{name: "JSON bare", args: []string{"extract"}, json: true},
		{name: "JSON trailing", args: []string{"extract", "--request=show customers", "--help"}, json: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime := newFakeExtractRuntime()
			cfg := testRouterConfig(t.TempDir(), tt.json)
			stdout, stderr, err := executeNativeExtract(context.Background(), cfg, runtime, tt.args...)
			if err != nil || stderr != "" {
				t.Fatalf("help: err=%v stderr=%q stdout=%q", err, stderr, stdout)
			}
			if tt.json {
				var envelope map[string]any
				decodeOneJSON(t, stdout, &envelope)
				if envelope["kind"] != "CommandManual" || envelope["command"] != "extract" || envelope["manual"] != manual {
					t.Fatalf("JSON manual = %#v", envelope)
				}
			} else if stdout != manual {
				t.Fatal("manual differs from canonical extract text")
			}
			assertExtractRuntimeNotCalled(t, runtime)
		})
	}

	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	for _, cmd := range root.Commands() {
		if cmd.Name() == "extract" && !cmd.Hidden {
			t.Fatal("extract leaked into root discovery")
		}
	}
}

func TestExtractLiteralUnknownOperandsGlobalsAndOutputCompatibility(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "literal does not terminate legacy parsing", args: []string{"extract", "--", "--request=show customers"}},
		{name: "legal unknown", args: []string{"extract", "--unknown", "ignored", "--request=show customers"}},
		{name: "malformed assigned unknown", args: []string{"extract", "--=ignored", "---bad=ignored", "--request=show customers"}},
		{name: "ignored operands", args: []string{"extract", "operand", "--request=show customers", "tail"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime := newFakeExtractRuntime()
			stdout, stderr, err := executeNativeExtract(context.Background(), testRouterConfig(t.TempDir(), true), runtime, tt.args...)
			if err != nil || stderr != "" {
				t.Fatalf("err=%v stderr=%q stdout=%q", err, stderr, stdout)
			}
			var env map[string]any
			decodeOneJSON(t, stdout, &env)
			if env["kind"] != "ExtractResult" || env["route"] != "simple_query" {
				t.Fatalf("envelope = %#v", env)
			}
			assertExtractRuntimeNotCalled(t, runtime)
		})
	}

	runtime := newFakeExtractRuntime()
	stdout, _, err := executeNativeExtract(context.Background(), testRouterConfig(t.TempDir(), true), runtime, "extract", "show customers")
	if err == nil || classifyError(mapCobraErr(err)).category != categoryUsage || stdout != "" {
		t.Fatalf("operand-only invocation: err=%v stdout=%q", err, stdout)
	}
	assertExtractRuntimeNotCalled(t, runtime)

	for _, args := range [][]string{
		{"--unknown", "extract", "--request=show customers"},
		{"--=bad", "extract", "--request=show customers"},
		{"operand", "extract", "--request=show customers"},
	} {
		var out, stderr bytes.Buffer
		code := Run(args, &out, &stderr)
		if code == 0 || strings.Contains(out.String(), "ExtractResult") {
			t.Fatalf("command-head discovery succeeded for %q: code=%d stdout=%q stderr=%q", args, code, out.String(), stderr.String())
		}
	}

	for _, args := range [][]string{
		{"--root", t.TempDir(), "--json", "extract", "--request=show customers"},
		{"extract", "--request=show customers", "--json=true", "--plain=true", "--no-input=true"},
	} {
		var out, stderr bytes.Buffer
		if code := Run(args, &out, &stderr); code != 0 || stderr.Len() != 0 {
			t.Fatalf("global form %q: code=%d stdout=%q stderr=%q", args, code, out.String(), stderr.String())
		}
		var env map[string]any
		decodeOneJSON(t, out.String(), &env)
		if env["kind"] != "ExtractResult" || env["api_version"] != apiVersion {
			t.Fatalf("global output = %#v", env)
		}
	}
}

func TestExtractRejectsUnsafeTablesBeforeAnalyzer(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   string
		out  string
	}{
		{name: "input traversal", in: "../outside", out: "scores"},
		{name: "input absolute", in: "/tmp/outside", out: "scores"},
		{name: "output traversal", in: "contacts", out: "../outside"},
		{name: "output separator", in: "contacts", out: "nested/scores"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			runtime := newFakeExtractRuntime()
			stdout, _, err := executeNativeExtract(context.Background(), testRouterConfig(t.TempDir(), true), runtime,
				"extract", "--request=predict churn", "--in="+tt.in, "--out="+tt.out,
			)
			if err == nil {
				t.Fatalf("unsafe tables succeeded: stdout=%q", stdout)
			}
			if got := classifyError(mapCobraErr(err)).category; got != categoryValidation {
				t.Fatalf("error category = %s, want validation: %v", got, err)
			}
			assertExtractRuntimeNotCalled(t, runtime)
		})
	}
}

type extractNativeContextKey struct{}

type fakeExtractRuntime struct {
	queryCalls           int
	querySQL             string
	queryLimit           int
	queryContextValue    any
	queryErr             error
	analyzerFactoryCalls int
	factoryRequest       string
	factoryContextValue  any
	factoryErr           error
	analyzerCalls        int
	analyzerContextValue any
	analyzerErr          error
	request              rlm.RunRequest
	closeCalls           int
	closeErr             error
}

func newFakeExtractRuntime() *fakeExtractRuntime { return &fakeExtractRuntime{} }

func (f *fakeExtractRuntime) runtime() extractCommandRuntime {
	return extractCommandRuntime{
		query: func(ctx context.Context, _ string, sql string, limit int) ([]connectors.Record, error) {
			f.queryCalls++
			f.querySQL = sql
			f.queryLimit = limit
			f.queryContextValue = ctx.Value(extractNativeContextKey{})
			if f.queryErr != nil {
				return nil, f.queryErr
			}
			return []connectors.Record{{"id": "synthetic"}}, nil
		},
		analyzer: func(ctx context.Context, _ config.Config, _ string, request string) (rlm.Analyzer, func() error, error) {
			f.analyzerFactoryCalls++
			f.factoryRequest = request
			f.factoryContextValue = ctx.Value(extractNativeContextKey{})
			if f.factoryErr != nil {
				return nil, nil, f.factoryErr
			}
			return fakeExtractAnalyzer{parent: f}, func() error {
				f.closeCalls++
				return f.closeErr
			}, nil
		},
	}
}

type fakeExtractAnalyzer struct{ parent *fakeExtractRuntime }

func (a fakeExtractAnalyzer) Run(ctx context.Context, req rlm.RunRequest) (rlm.RunResult, error) {
	a.parent.analyzerCalls++
	a.parent.analyzerContextValue = ctx.Value(extractNativeContextKey{})
	a.parent.request = req
	if a.parent.analyzerErr != nil {
		return rlm.RunResult{}, a.parent.analyzerErr
	}
	return rlm.RunResult{Mode: "agent", InTable: req.InTable, OutTable: req.OutTable, RecordsRead: 1, RecordsScored: 1, Duration: time.Nanosecond}, nil
}

func (a fakeExtractAnalyzer) Mode() string { return "agent" }

func executeNativeExtract(ctx context.Context, cfg config.Config, runtime *fakeExtractRuntime, args ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := newRootCmdWithExtractRuntime(ctx, cfg, &stdout, &stderr, runtime.runtime())
	err := executeRootCmd(cmd, args)
	return stdout.String(), stderr.String(), err
}

func assertExtractRuntimeNotCalled(t *testing.T, runtime *fakeExtractRuntime) {
	t.Helper()
	if runtime.queryCalls != 0 || runtime.analyzerFactoryCalls != 0 || runtime.analyzerCalls != 0 || runtime.closeCalls != 0 {
		t.Fatalf("extract runtime called: query=%d factory=%d analyzer=%d close=%d", runtime.queryCalls, runtime.analyzerFactoryCalls, runtime.analyzerCalls, runtime.closeCalls)
	}
}

func assertExtractNoFileCompletion(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	if cmd.ValidArgsFunction == nil {
		t.Fatalf("%s missing completion seam", cmd.CommandPath())
	}
	values, directive := cmd.ValidArgsFunction(cmd, nil, "")
	if len(values) != 0 || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("%s completion = (%v, %v), want empty NoFileComp", cmd.CommandPath(), values, directive)
	}
}

func TestExtractInjectedErrorsPreserveWrapping(t *testing.T) {
	runtime := newFakeExtractRuntime()
	runtime.queryErr = errors.New("synthetic query failure")
	_, _, err := executeNativeExtract(context.Background(), testRouterConfig(t.TempDir(), false), runtime,
		"extract", "--request=show customers", "--sql=SELECT * FROM contacts",
	)
	if err == nil || !strings.Contains(err.Error(), "extract: query: synthetic query failure") {
		t.Fatalf("query error = %v", err)
	}
}
