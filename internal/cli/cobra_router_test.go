package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/connectors"
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
	if len(expectedHidden) != len(cobraLegacyCommands(config.Config{}, defaultAppOpeners())) {
		t.Fatalf("expectedHidden covers %d commands, cobraLegacyCommands registers %d", len(expectedHidden), len(cobraLegacyCommands(config.Config{}, defaultAppOpeners())))
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
				if !got.DisableFlagParsing {
					t.Fatalf("%s wrapper must keep DisableFlagParsing", name)
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
	cmd.SetArgs([]string{"--definitely-unknown"})

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
}

func TestCobraRouterShellRejectsUnknownHelpWithUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	args := []string{"nosuch", "--help", "--json"}
	code := Run(args, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("Run(%v) code = %d, want 2; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
	}
	if want := `"message": "unknown command \"nosuch\""`; !strings.Contains(stdout.String(), want) {
		t.Fatalf("stdout missing %q:\n%s", want, stdout.String())
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("stderr missing usage error: %s", stderr.String())
	}
}

func TestCobraRouterShellPreservesDynamicConnectorHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	args := []string{"github", "help", "--json"}
	code := Run(args, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(%v) code = %d, want 0; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
	}
	for _, want := range []string{`"kind": "CommandManual"`, `"command": "github"`, `pm github`} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout missing %q:\n%s", want, stdout.String())
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
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

type legacyHelpCase struct {
	path []string
}

func TestEveryLegacyLeafHelpIsExecutable(t *testing.T) {
	cases := legacyHelpCases(t)
	for _, project := range []struct {
		name        string
		initialized bool
	}{
		{name: "outside a project"},
		{name: "inside an initialized project", initialized: true},
	} {
		t.Run(project.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			if project.initialized {
				var stdout, stderr bytes.Buffer
				if code := Run([]string{"init", "--root", root}, &stdout, &stderr); code != 0 {
					t.Fatalf("initialize project: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
				}
			}

			for _, helpFlag := range []string{"--help", "-h"} {
				for _, tc := range cases {
					t.Run(strings.Join(append(append([]string(nil), tc.path...), helpFlag), " "), func(t *testing.T) {
						args := append(append([]string(nil), tc.path...), helpFlag)
						var stdout, stderr bytes.Buffer
						if code := Run(args, &stdout, &stderr); code != 0 {
							t.Fatalf("Run(%v) code=%d stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
						}
						if stderr.Len() != 0 {
							t.Fatalf("Run(%v) wrote stderr while rendering help: %s", args, stderr.String())
						}
						want := "pm " + strings.Join(tc.path, " ")
						if !strings.Contains(stdout.String(), want) {
							t.Fatalf("Run(%v) missing its documented invocation %q:\n%s", args, want, stdout.String())
						}
					})
				}
			}
		})
	}
}

func TestEveryDynamicConnectorLeafHelpRendersWithoutDispatch(t *testing.T) {
	registry := appRegistry()
	type helpCase struct {
		connectorName string
		connector     connectors.Connector
		surface       *connectors.CommandSurface
		commandPath   string
		path          []string
		helpFlag      string
	}
	var cases []helpCase
	for _, meta := range registry.List() {
		connector, ok := registry.Get(meta.Name)
		if !ok {
			t.Fatalf("registered connector %q is missing", meta.Name)
		}
		provider, ok := connector.(connectors.CommandSurfaceProvider)
		if !ok || provider.CommandSurface() == nil {
			continue
		}
		surface := provider.CommandSurface()
		for _, command := range surface.Commands {
			path := strings.Fields(command.Path)
			if len(path) == 0 {
				t.Errorf("connector %q declares a command with an empty path", meta.Name)
				continue
			}
			for _, helpFlag := range []string{"--help", "-h"} {
				cases = append(cases, helpCase{
					connectorName: meta.Name,
					connector:     connector,
					surface:       surface,
					commandPath:   command.Path,
					path:          append([]string(nil), path...),
					helpFlag:      helpFlag,
				})
			}
		}
	}
	if len(cases) == 0 {
		t.Fatal("no dynamic connector commands were checked")
	}
	for _, tc := range cases {
		tc := tc
		t.Run(strings.Join([]string{tc.connectorName, tc.commandPath, tc.helpFlag}, " "), func(t *testing.T) {
			t.Parallel()
			args := append(append([]string(nil), tc.path...), tc.helpFlag)
			if !connectorHelpRequested(args, tc.surface) {
				t.Fatalf("connector %q command %q did not resolve %s before dispatch", tc.connectorName, tc.commandPath, tc.helpFlag)
			}
			gotCommand, manual := renderConnectorCommandManual(tc.connectorName, tc.connector, tc.surface, args)
			wantCommand := tc.connectorName + " " + tc.commandPath
			if gotCommand != wantCommand || !strings.Contains(manual, "NAME\n") {
				t.Fatalf("connector %q command %q %s manual = %q / %q, want %q with NAME section", tc.connectorName, tc.commandPath, tc.helpFlag, gotCommand, manual, wantCommand)
			}
		})
	}
	t.Logf("checked %d dynamic connector command help variants", len(cases))
}

func TestLeafHelpDoesNotMaskInvalidCommandsOrApprovalCarrierSyntax(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
	}{
		{name: "unknown top-level command", args: []string{"not-a-command", "--help"}},
		{name: "unknown leaf command", args: []string{"credentials", "not-a-command", "--help"}},
		{name: "approval carrier with value", args: []string{"reverse", "run", "--approval-token-stdin=token", "--help"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run(tt.args, &stdout, &stderr); code != 2 {
				t.Fatalf("Run(%v) code=%d want usage exit 2; stdout=%s stderr=%s", tt.args, code, stdout.String(), stderr.String())
			}
			if strings.Contains(stdout.String(), "NAME\n") {
				t.Fatalf("Run(%v) rendered help instead of refusing invalid input: %s", tt.args, stdout.String())
			}
		})
	}
}

func TestLeafHelpWithOtherFlagsRendersBeforeRequiredFlags(t *testing.T) {
	for _, args := range [][]string{
		{"schedule", "create", "--name", "nightly", "--help"},
		{"schedule", "inspect", "--help"},
	} {
		var stdout, stderr bytes.Buffer
		if code := Run(args, &stdout, &stderr); code != 0 {
			t.Fatalf("Run(%v) code=%d stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
		}
		if !strings.Contains(stdout.String(), "pm schedule create --name nightly") {
			t.Fatalf("Run(%v) did not render the schedule manual:\n%s", args, stdout.String())
		}
		if stderr.Len() != 0 {
			t.Fatalf("Run(%v) wrote stderr while rendering help: %s", args, stderr.String())
		}
	}
}

func legacyHelpCases(t *testing.T) []legacyHelpCase {
	t.Helper()
	seen := map[string]legacyHelpCase{}
	for _, spec := range cobraLegacyCommands(config.Config{}, defaultAppOpeners()) {
		manual, ok := docs[spec.name]
		if !ok {
			t.Errorf("legacy wrapper %q has no manual; every executable wrapper must be discoverable", spec.name)
			continue
		}
		found := 0
		for _, path := range manualCommandPaths(manual) {
			if len(path) == 0 || path[0] != spec.name {
				continue
			}
			found++
			key := strings.Join(path, "\x00")
			seen[key] = legacyHelpCase{path: path}
		}
		if found == 0 {
			t.Errorf("legacy wrapper %q manual has no parseable pm invocation", spec.name)
		}
	}
	if t.Failed() {
		t.FailNow()
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	cases := make([]legacyHelpCase, 0, len(keys))
	for _, key := range keys {
		cases = append(cases, seen[key])
	}
	return cases
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
