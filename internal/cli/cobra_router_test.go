package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/config"
)

func TestCobraRouterShellBuildsFreshHiddenWrapperTree(t *testing.T) {
	first := newRootCmd(context.Background(), testRouterConfig("/first-root", false), io.Discard, io.Discard)
	second := newRootCmd(context.Background(), testRouterConfig("/second-root", true), io.Discard, io.Discard)
	if first == second {
		t.Fatal("newRootCmd returned the same command tree instance")
	}

	expectedHidden := map[string]bool{
		"init":        false,
		"help":        false,
		"man":         false,
		"connectors":  false,
		"credentials": false,
		"connections": false,
		"catalog":     false,
		"etl":         false,
		"query":       false,
		"reverse":     false,
		"agent":       false,
		"runtime":     false,
		"flow":        false,
		"extract":     true,
		"perf":        false,
		"docs":        false,
		"skills":      false,
		"version":     false,
		"rlm":         false,
		"schedule":    false,
		"worker":      true,
	}
	legacyCommands := map[string]struct{}{}
	for _, spec := range cobraLegacyCommands(config.Config{}) {
		legacyCommands[spec.name] = struct{}{}
	}
	nativeCommands := map[string]struct{}{"agent": {}, "catalog": {}, "connections": {}, "credentials": {}, "docs": {}, "etl": {}, "extract": {}, "flow": {}, "perf": {}, "query": {}, "reverse": {}, "rlm": {}, "runtime": {}, "schedule": {}, "skills": {}, "version": {}, "worker": {}}
	if len(expectedHidden) != len(legacyCommands)+len(nativeCommands) {
		t.Fatalf("expectedHidden covers %d commands, legacy commands plus native commands registers %d", len(expectedHidden), len(legacyCommands)+len(nativeCommands))
	}

	for _, root := range []*cobra.Command{first, second} {
		t.Run(root.CommandPath(), func(t *testing.T) {
			if !root.DisableFlagParsing {
				t.Fatal("root command must keep legacy global parsing and connector flag passthrough")
			}
			if !root.SilenceErrors || !root.SilenceUsage {
				t.Fatal("cobra errors/usages must be silenced so writeError remains the sole reporter")
			}
			for name, hidden := range expectedHidden {
				got := findCobraCommand(root, name)
				if got == nil {
					t.Fatalf("missing top-level cobra wrapper %q", name)
				}
				if got.Hidden != hidden {
					t.Fatalf("%s hidden = %t, want %t", name, got.Hidden, hidden)
				}
				_, legacy := legacyCommands[name]
				_, native := nativeCommands[name]
				if legacy && !got.DisableFlagParsing {
					t.Fatalf("%s wrapper must keep DisableFlagParsing", name)
				}
				if native && got.DisableFlagParsing {
					t.Fatalf("%s command must use native Cobra flag parsing", name)
				}
			}
			for _, cmd := range root.Commands() {
				if _, ok := expectedHidden[cmd.Name()]; !ok {
					t.Fatalf("unexpected top-level cobra wrapper %q", cmd.Name())
				}
			}
		})
	}
}

func TestCatalogCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	catalog := findCobraCommand(root, "catalog")
	if catalog == nil {
		t.Fatal("missing catalog command")
	}
	if catalog.DisableFlagParsing {
		t.Fatal("catalog command must use native Cobra flag parsing")
	}

	for _, name := range []string{"refresh", "show"} {
		t.Run(name, func(t *testing.T) {
			action := findCobraCommand(catalog, name)
			if action == nil {
				t.Fatalf("missing catalog %s subcommand", name)
			}
			if action.DisableFlagParsing {
				t.Fatalf("catalog %s must use native Cobra flag parsing", name)
			}
			flag := action.Flags().Lookup("connection")
			if flag == nil {
				t.Fatalf("catalog %s missing native --connection flag", name)
			}
			if got, want := flag.Value.Type(), "stringArray"; got != want {
				t.Fatalf("catalog %s --connection flag type = %q, want %q", name, got, want)
			}
			if got, want := flag.NoOptDefVal, "true"; got != want {
				t.Fatalf("catalog %s --connection NoOptDefVal = %q, want %q", name, got, want)
			}
			if !action.FParseErrWhitelist.UnknownFlags {
				t.Fatalf("catalog %s must preserve legacy unknown-flag tolerance", name)
			}
		})
	}
}

func TestQueryCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	query := findCobraCommand(root, "query")
	if query == nil {
		t.Fatal("missing query command")
	}
	if query.DisableFlagParsing {
		t.Fatal("query command must use native Cobra flag parsing")
	}

	run := findCobraCommand(query, "run")
	if run == nil {
		t.Fatal("missing query run subcommand")
	}
	if run.DisableFlagParsing {
		t.Fatal("query run must use native Cobra flag parsing")
	}
	if !run.FParseErrWhitelist.UnknownFlags {
		t.Fatal("query run must preserve legacy unknown-flag tolerance")
	}
	if run.ValidArgsFunction == nil {
		t.Fatal("query run must suppress file completion fallback until Phase 15 completions")
	}
	completions, directive := run.ValidArgsFunction(run, nil, "")
	if len(completions) != 0 {
		t.Fatalf("query run completion seam returned %v, want no Phase 15 completions", completions)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("query run completion directive = %v, want NoFileComp", directive)
	}
	for _, name := range []string{"table", "sql", "limit", "fields", "agent-mode", "sample"} {
		t.Run("run flag "+name, func(t *testing.T) {
			flag := run.Flags().Lookup(name)
			if flag == nil {
				t.Fatalf("query run missing native --%s flag", name)
			}
			if got, want := flag.Value.Type(), "stringArray"; got != want {
				t.Fatalf("query run --%s flag type = %q, want %q", name, got, want)
			}
			if got, want := flag.NoOptDefVal, "true"; got != want {
				t.Fatalf("query run --%s NoOptDefVal = %q, want %q", name, got, want)
			}
		})
	}
}

func TestRuntimeCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	runtimeCmd := findCobraCommand(root, "runtime")
	if runtimeCmd == nil {
		t.Fatal("missing runtime command")
	}
	if runtimeCmd.DisableFlagParsing {
		t.Fatal("runtime command must use native Cobra flag parsing")
	}

	doctor := findCobraCommand(runtimeCmd, "doctor")
	if doctor == nil {
		t.Fatal("missing runtime doctor subcommand")
	}
	if doctor.DisableFlagParsing {
		t.Fatal("runtime doctor must use native Cobra flag parsing")
	}
	if !doctor.FParseErrWhitelist.UnknownFlags {
		t.Fatal("runtime doctor must preserve legacy unknown-flag tolerance")
	}
	if doctor.ValidArgsFunction == nil {
		t.Fatal("runtime doctor must suppress file completion fallback until Phase 15 completions")
	}
	completions, directive := doctor.ValidArgsFunction(doctor, nil, "")
	if len(completions) != 0 {
		t.Fatalf("runtime doctor completion seam returned %v, want no Phase 15 completions", completions)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("runtime doctor completion directive = %v, want NoFileComp", directive)
	}
}

func TestDocsCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	docsCmd := findCobraCommand(root, "docs")
	if docsCmd == nil {
		t.Fatal("missing docs command")
	}
	if docsCmd.DisableFlagParsing {
		t.Fatal("docs command must use native Cobra flag parsing")
	}

	for _, tt := range []struct {
		name  string
		flags []string
	}{
		{name: "generate", flags: []string{"dir", "connectors-dir"}},
		{name: "validate", flags: []string{"dir", "connectors-dir"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			action := findCobraCommand(docsCmd, tt.name)
			if action == nil {
				t.Fatalf("missing docs %s subcommand", tt.name)
			}
			if action.DisableFlagParsing {
				t.Fatalf("docs %s must use native Cobra flag parsing", tt.name)
			}
			if !action.FParseErrWhitelist.UnknownFlags {
				t.Fatalf("docs %s must preserve legacy unknown-flag tolerance", tt.name)
			}
			if action.ValidArgsFunction == nil {
				t.Fatalf("docs %s must suppress file completion fallback until Phase 15 completions", tt.name)
			}
			completions, directive := action.ValidArgsFunction(action, nil, "")
			if len(completions) != 0 {
				t.Fatalf("docs %s completion seam returned %v, want no Phase 15 completions", tt.name, completions)
			}
			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Fatalf("docs %s completion directive = %v, want NoFileComp", tt.name, directive)
			}
			for _, name := range tt.flags {
				flag := action.Flags().Lookup(name)
				if flag == nil {
					t.Fatalf("docs %s missing native --%s flag", tt.name, name)
				}
				if got, want := flag.Value.Type(), "stringArray"; got != want {
					t.Fatalf("docs %s --%s flag type = %q, want %q", tt.name, name, got, want)
				}
				if got, want := flag.NoOptDefVal, "true"; got != want {
					t.Fatalf("docs %s --%s NoOptDefVal = %q, want %q", tt.name, name, got, want)
				}
			}
		})
	}

	help := findCobraCommand(docsCmd, "help")
	if help == nil || !help.Hidden {
		t.Fatal("docs must preserve a hidden positional help alias until Phase 19")
	}
}

func TestSkillsCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	skills := findCobraCommand(root, "skills")
	if skills == nil {
		t.Fatal("missing skills command")
	}
	if skills.DisableFlagParsing {
		t.Fatal("skills command must use native Cobra flag parsing")
	}

	generate := findCobraCommand(skills, "generate")
	if generate == nil {
		t.Fatal("missing skills generate subcommand")
	}
	if generate.DisableFlagParsing {
		t.Fatal("skills generate must use native Cobra flag parsing")
	}
	if !generate.FParseErrWhitelist.UnknownFlags {
		t.Fatal("skills generate must preserve legacy unknown-flag tolerance")
	}
	if generate.ValidArgsFunction == nil {
		t.Fatal("skills generate must suppress file completion fallback until Phase 15 completions")
	}
	completions, directive := generate.ValidArgsFunction(generate, nil, "")
	if len(completions) != 0 {
		t.Fatalf("skills generate completion seam returned %v, want no Phase 15 completions", completions)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("skills generate completion directive = %v, want NoFileComp", directive)
	}
	flag := generate.Flags().Lookup("dir")
	if flag == nil {
		t.Fatal("skills generate missing native --dir flag")
	}
	if got, want := flag.Value.Type(), "stringArray"; got != want {
		t.Fatalf("skills generate --dir flag type = %q, want %q", got, want)
	}
	if got, want := flag.NoOptDefVal, "true"; got != want {
		t.Fatalf("skills generate --dir NoOptDefVal = %q, want %q", got, want)
	}

	help := findCobraCommand(skills, "help")
	if help == nil || !help.Hidden {
		t.Fatal("skills must preserve a hidden positional help alias until Phase 19")
	}
}

func TestVersionCommandIsNativeCobraLeaf(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	versionCmd := findCobraCommand(root, "version")
	if versionCmd == nil {
		t.Fatal("missing version command")
	}
	if versionCmd.DisableFlagParsing {
		t.Fatal("version command must use native Cobra flag parsing")
	}
	if versionCmd.FParseErrWhitelist.UnknownFlags {
		t.Fatal("version command must preserve legacy rejection of unknown flags")
	}
	if versionCmd.ValidArgsFunction == nil {
		t.Fatal("version command must suppress file completion fallback until Phase 15 completions")
	}
	completions, directive := versionCmd.ValidArgsFunction(versionCmd, nil, "")
	if len(completions) != 0 {
		t.Fatalf("version completion seam returned %v, want no Phase 15 completions", completions)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("version completion directive = %v, want NoFileComp", directive)
	}

	help := findCobraCommand(versionCmd, "help")
	if help == nil {
		t.Fatal("version command must preserve positional help compatibility")
	}
	if !help.Hidden {
		t.Fatal("version positional help alias must stay hidden until Phase 19 help-tree work")
	}
}

func TestPerfCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	perf := findCobraCommand(root, "perf")
	if perf == nil {
		t.Fatal("missing perf command")
	}
	if perf.DisableFlagParsing {
		t.Fatal("perf command must use native Cobra flag parsing")
	}

	compare := findCobraCommand(perf, "compare")
	if compare == nil {
		t.Fatal("missing perf compare subcommand")
	}
	if compare.DisableFlagParsing {
		t.Fatal("perf compare must use native Cobra flag parsing")
	}
	if !compare.FParseErrWhitelist.UnknownFlags {
		t.Fatal("perf compare must preserve legacy unknown-flag tolerance")
	}
	if compare.ValidArgsFunction == nil {
		t.Fatal("perf compare must suppress file completion fallback until Phase 15 completions")
	}
	completions, directive := compare.ValidArgsFunction(compare, nil, "")
	if len(completions) != 0 {
		t.Fatalf("perf compare completion seam returned %v, want no Phase 15 completions", completions)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("perf compare completion directive = %v, want NoFileComp", directive)
	}
	for _, name := range []string{"iterations", "runtime"} {
		t.Run("compare flag "+name, func(t *testing.T) {
			flag := compare.Flags().Lookup(name)
			if flag == nil {
				t.Fatalf("perf compare missing native --%s flag", name)
			}
			if got, want := flag.Value.Type(), "stringArray"; got != want {
				t.Fatalf("perf compare --%s flag type = %q, want %q", name, got, want)
			}
			if got, want := flag.NoOptDefVal, "true"; got != want {
				t.Fatalf("perf compare --%s NoOptDefVal = %q, want %q", name, got, want)
			}
		})
	}

	syncModes := findCobraCommand(perf, "sync-modes")
	if syncModes == nil {
		t.Fatal("missing perf sync-modes subcommand")
	}
	if syncModes.DisableFlagParsing {
		t.Fatal("perf sync-modes must use native Cobra flag parsing")
	}
	if !syncModes.FParseErrWhitelist.UnknownFlags {
		t.Fatal("perf sync-modes must preserve legacy unknown-flag tolerance")
	}
	if syncModes.ValidArgsFunction == nil {
		t.Fatal("perf sync-modes must suppress file completion fallback until Phase 15 completions")
	}
	completions, directive = syncModes.ValidArgsFunction(syncModes, nil, "")
	if len(completions) != 0 {
		t.Fatalf("perf sync-modes completion seam returned %v, want no Phase 15 completions", completions)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("perf sync-modes completion directive = %v, want NoFileComp", directive)
	}
	flag := syncModes.Flags().Lookup("records")
	if flag == nil {
		t.Fatal("perf sync-modes missing native --records flag")
	}
	if got, want := flag.Value.Type(), "stringArray"; got != want {
		t.Fatalf("perf sync-modes --records flag type = %q, want %q", got, want)
	}
	if got, want := flag.NoOptDefVal, "true"; got != want {
		t.Fatalf("perf sync-modes --records NoOptDefVal = %q, want %q", got, want)
	}
}

func TestConnectionsCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	connections := findCobraCommand(root, "connections")
	if connections == nil {
		t.Fatal("missing connections command")
	}
	if connections.DisableFlagParsing {
		t.Fatal("connections command must use native Cobra flag parsing")
	}
	if connections.ValidArgsFunction == nil {
		t.Fatal("connections command must preserve connection-name completion compatibility seam")
	}
	completions, directive := connections.ValidArgsFunction(connections, nil, "")
	if len(completions) != 0 {
		t.Fatalf("connection completion seam returned %v, want no Phase 15 completions", completions)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("connection completion directive = %v, want NoFileComp", directive)
	}

	create := findCobraCommand(connections, "create")
	if create == nil {
		t.Fatal("missing connections create subcommand")
	}
	if create.DisableFlagParsing {
		t.Fatal("connections create must use native Cobra flag parsing")
	}
	if !create.FParseErrWhitelist.UnknownFlags {
		t.Fatal("connections create must preserve legacy unknown-flag tolerance")
	}
	for _, name := range []string{"source", "destination", "stream", "sync-mode", "cursor", "primary-key", "table", "source-config", "destination-config"} {
		t.Run("create flag "+name, func(t *testing.T) {
			flag := create.Flags().Lookup(name)
			if flag == nil {
				t.Fatalf("connections create missing native --%s flag", name)
			}
			if got, want := flag.Value.Type(), "stringArray"; got != want {
				t.Fatalf("connections create --%s flag type = %q, want %q", name, got, want)
			}
			if got, want := flag.NoOptDefVal, "true"; got != want {
				t.Fatalf("connections create --%s NoOptDefVal = %q, want %q", name, got, want)
			}
		})
	}

	list := findCobraCommand(connections, "list")
	if list == nil {
		t.Fatal("missing connections list subcommand")
	}
	if list.DisableFlagParsing {
		t.Fatal("connections list must use native Cobra flag parsing")
	}
	if !list.FParseErrWhitelist.UnknownFlags {
		t.Fatal("connections list must preserve legacy unknown-flag tolerance")
	}
}

func TestCobraRouterShellRootPersistentFlagsArePerFreshCommand(t *testing.T) {
	first := newRootCmd(context.Background(), testRouterConfig("/first-root", false), io.Discard, io.Discard)
	second := newRootCmd(context.Background(), testRouterConfig("/second-root", true), io.Discard, io.Discard)

	firstRoot := first.PersistentFlags().Lookup("root")
	secondRoot := second.PersistentFlags().Lookup("root")
	if firstRoot == nil || secondRoot == nil {
		t.Fatalf("fresh roots must define persistent --root flags: first=%v second=%v", firstRoot, secondRoot)
	}
	firstJSON := first.PersistentFlags().Lookup("json")
	secondJSON := second.PersistentFlags().Lookup("json")
	if firstJSON == nil || secondJSON == nil {
		t.Fatalf("fresh roots must define persistent --json flags: first=%v second=%v", firstJSON, secondJSON)
	}

	if got, want := firstRoot.Value.String(), "/first-root"; got != want {
		t.Fatalf("first --root value = %q, want %q", got, want)
	}
	if got, want := secondRoot.Value.String(), "/second-root"; got != want {
		t.Fatalf("second --root value = %q, want %q", got, want)
	}
	if got, want := firstJSON.Value.String(), "false"; got != want {
		t.Fatalf("first --json value = %q, want %q", got, want)
	}
	if got, want := secondJSON.Value.String(), "true"; got != want {
		t.Fatalf("second --json value = %q, want %q", got, want)
	}

	if err := firstRoot.Value.Set("/mutated-root"); err != nil {
		t.Fatalf("mutate first root flag: %v", err)
	}
	if err := secondJSON.Value.Set("false"); err != nil {
		t.Fatalf("mutate second json flag: %v", err)
	}
	if got, want := secondRoot.Value.String(), "/second-root"; got != want {
		t.Fatalf("second root flag shared state after first mutation: got %q want %q", got, want)
	}
	if got, want := firstJSON.Value.String(), "false"; got != want {
		t.Fatalf("first json flag shared state after second mutation: got %q want %q", got, want)
	}
}

func TestCobraRouterShellDoesNotReclassifyLegacyHandlerErrors(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{name: "unknown flag text", message: `legacy connector handler failed: unknown flag --private`},
		{name: "unknown command text", message: `legacy connector handler failed: unknown command "private"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &cobra.Command{Use: "pm", DisableFlagParsing: true, SilenceErrors: true, SilenceUsage: true}
			root.AddCommand(newLegacyCobraCommand(context.Background(), ".", io.Discard, false, cobraLegacyCommand{
				name: "legacy",
				handler: func(context.Context, string, []string, io.Writer, bool) error {
					return errors.New(tt.message)
				},
			}))

			err := executeRootCmd(root, []string{"legacy", "run"})
			if err == nil {
				t.Fatal("executeRootCmd returned nil, want legacy handler error")
			}
			classified := classifyError(mapCobraErr(err))
			if classified.category != categoryInternal {
				t.Fatalf("category = %s, want %s for %q", classified.category, categoryInternal, tt.message)
			}
			if code := exitCodeFor(classified); code != 1 {
				t.Fatalf("exit code = %d, want 1 for preserved legacy classification", code)
			}
			if classified.Error() != tt.message {
				t.Fatalf("message = %q, want %q", classified.Error(), tt.message)
			}
		})
	}
}

func TestCobraRouterShellMapsGenuineCobraParseErrorsToUsage(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		configure func(*cobra.Command)
	}{
		{name: "unknown flag", args: []string{"--definitely-unknown"}},
		{
			name: "missing known flag value",
			args: []string{"--root"},
			configure: func(cmd *cobra.Command) {
				cmd.Flags().String("root", ".", "project root")
			},
		},
		{
			name: "invalid bool value",
			args: []string{"--json=maybe"},
			configure: func(cmd *cobra.Command) {
				cmd.Flags().Bool("json", false, "write JSON")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:           "pm",
				SilenceUsage:  true,
				SilenceErrors: true,
				RunE: func(*cobra.Command, []string) error {
					return nil
				},
			}
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			if tt.configure != nil {
				tt.configure(cmd)
			}
			cmd.SetArgs(tt.args)

			_, err := cmd.ExecuteC()
			if err == nil {
				t.Fatal("ExecuteC returned nil, want Cobra parse error")
			}
			classified := classifyError(mapCobraErr(err))
			if classified.category != categoryUsage {
				t.Fatalf("category = %s, want %s for genuine Cobra parse error %q", classified.category, categoryUsage, err.Error())
			}
			if code := exitCodeFor(classified); code != 2 {
				t.Fatalf("exit code = %d, want 2 for Cobra parse error", code)
			}
		})
	}
}

func TestCobraRouterShellPreservesLegacyHelpInterceptionForFallback(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "unknown command help", args: []string{"nosuch", "--help", "--json"}, want: `"message": "help topic \"nosuch\" not found"`},
		{name: "dynamic connector help", args: []string{"github", "help", "--json"}, want: `"message": "help topic \"github\" not found"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(tt.args, &stdout, &stderr)
			if code != 1 {
				t.Fatalf("Run(%v) code = %d, want 1; stdout=%s stderr=%s", tt.args, code, stdout.String(), stderr.String())
			}
			if !strings.Contains(stdout.String(), tt.want) {
				t.Fatalf("stdout missing %q:\n%s", tt.want, stdout.String())
			}
			if strings.Contains(stderr.String(), "unknown command") || strings.Contains(stderr.String(), "missing connector command path") {
				t.Fatalf("fallback help was routed as command execution: stderr=%s", stderr.String())
			}
		})
	}
}

func TestCobraRouterShellPreservesDynamicConnectorPassthroughWithLateGlobals(t *testing.T) {
	var gotPath, gotState string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotState = r.URL.Query().Get("state")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"id": 101,
				"node_id": "I_kwDOAA",
				"number": 101,
				"state": "closed",
				"title": "closed issue",
				"user": {"login": "octocat", "id": 1},
				"updated_at": "2026-07-06T00:00:00Z"
			}
		]`))
	}))
	t.Cleanup(srv.Close)

	root := t.TempDir()
	runCobraRouterCLI(t, []string{"init", "--root", root, "--json"})
	runCobraRouterCLI(t, []string{
		"credentials", "add", "github-router",
		"--connector", "github",
		"--config", "owner=octocat",
		"--config", "repo=hello-world",
		"--config", "base_url=" + srv.URL,
		"--config", "public_access=true",
		"--root", root,
		"--json",
	})

	stdout, _ := runCobraRouterCLI(t, []string{
		"github", "issue", "list",
		"--credential", "github-router",
		"--state", "closed",
		"--limit", "1",
		"--root", root,
		"--json",
	})
	if gotPath != "/repos/octocat/hello-world/issues" {
		t.Fatalf("request path = %q, want /repos/octocat/hello-world/issues", gotPath)
	}
	if gotState != "closed" {
		t.Fatalf("request state = %q, want connector flag passthrough", gotState)
	}

	var env struct {
		Kind    string `json:"kind"`
		Command string `json:"command"`
		Stream  string `json:"stream"`
		Count   int    `json:"count"`
	}
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout)
	}
	if env.Kind != "ConnectorCommandRead" || env.Command != "issue list" || env.Stream != "issues" || env.Count != 1 {
		t.Fatalf("envelope = %+v, want dynamic connector read result", env)
	}
}

func runCobraRouterCLI(t *testing.T, args []string) (stdout string, stderr string) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	code := Run(args, &outBuf, &errBuf)
	if code != 0 {
		t.Fatalf("Run(%v) code = %d stderr=%s stdout=%s", args, code, errBuf.String(), outBuf.String())
	}
	return outBuf.String(), errBuf.String()
}

func testRouterConfig(root string, jsonOut bool) config.Config {
	return config.Config{Root: root, JSON: jsonOut}
}

func findCobraCommand(root *cobra.Command, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
