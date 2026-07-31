package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	if len(expectedHidden) != len(cobraLegacyCommands(config.Config{})) {
		t.Fatalf("expectedHidden covers %d commands, cobraLegacyCommands registers %d", len(expectedHidden), len(cobraLegacyCommands(config.Config{})))
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

func TestCobraRouterShellPreservesLegacyHelpInterceptionForFallback(t *testing.T) {
	var stdout, stderr bytes.Buffer
	args := []string{"nosuch", "--help", "--json"}
	code := Run(args, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("Run(%v) code = %d, want 1; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
	}
	if want := `"message": "help topic \"nosuch\" not found"`; !strings.Contains(stdout.String(), want) {
		t.Fatalf("stdout missing %q:\n%s", want, stdout.String())
	}
	if strings.Contains(stderr.String(), "unknown command") || strings.Contains(stderr.String(), "missing connector command path") {
		t.Fatalf("fallback help was routed as command execution: stderr=%s", stderr.String())
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

func TestCobraRouterDynamicLucidELDUsesConfigDateWindowAndRejectsProviderPaginationFlags(t *testing.T) {
	var historyQuery url.Values
	var latestDriverQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v2/vehicle-location-history/vehicle_fixture_001":
			historyQuery = r.URL.Query()
			_, _ = w.Write([]byte(`{
				"data": [{"id": "history_fixture_001"}],
				"description": "ok",
				"next_page_token": "",
				"page": 1,
				"size": 1,
				"status_code": 200,
				"total_elements": 1,
				"total_pages": 1
			}`))
		case "/v2/latest-driver-status":
			latestDriverQuery = r.URL.Query()
			_, _ = w.Write([]byte(`{"data": [{"id": "driver_status_fixture_001"}], "status_code": 200}`))
		case "/v2/drivers", "/v2/vehicles":
			_, _ = w.Write([]byte(`{"data": [], "status_code": 200}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	root := t.TempDir()
	t.Setenv("PM_TEST_LUCID_PROVIDER_KEY", "fixture_provider_key")
	t.Setenv("PM_TEST_LUCID_COMPANY_KEY", "fixture_company_key")
	runCobraRouterCLI(t, []string{"init", "--root", root, "--json"})
	runCobraRouterCLI(t, []string{
		"credentials", "add", "lucid-router",
		"--connector", "lucid-eld",
		"--config", "base_url=" + srv.URL,
		"--config", "vehicle_id=vehicle_fixture_001",
		"--config", "start_date=01-01-2026",
		"--config", "end_date=01-02-2026",
		"--from-env", "provider_api_key=PM_TEST_LUCID_PROVIDER_KEY",
		"--from-env", "company_api_key=PM_TEST_LUCID_COMPANY_KEY",
		"--root", root,
		"--json",
	})

	stdout, _ := runCobraRouterCLI(t, []string{
		"lucid-eld", "vehicle-location-history", "list",
		"--credential", "lucid-router",
		"--limit", "1",
		"--root", root,
		"--json",
	})
	if got := historyQuery.Get("start_date"); got != "01-01-2026" {
		t.Fatalf("history start_date = %q, want config-backed date window", got)
	}
	if got := historyQuery.Get("end_date"); got != "01-02-2026" {
		t.Fatalf("history end_date = %q, want config-backed date window", got)
	}
	if got := historyQuery.Get("limit"); got != "100" {
		t.Fatalf("history provider limit = %q, want fixed Tier-1 default 100", got)
	}
	var env struct {
		Kind    string `json:"kind"`
		Command string `json:"command"`
		Stream  string `json:"stream"`
		Count   int    `json:"count"`
	}
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("decode history json: %v\n%s", err, stdout)
	}
	if env.Kind != "ConnectorCommandRead" || env.Command != "vehicle-location-history list" || env.Stream != "vehicle_location_history" || env.Count != 1 {
		t.Fatalf("history envelope = %+v, want dynamic connector read with one client-limited record", env)
	}

	runCobraRouterCLI(t, []string{
		"lucid-eld", "latest", "driver", "statuses", "list",
		"--credential", "lucid-router",
		"--driver-id", "driver_fixture_001",
		"--limit", "25",
		"--root", root,
		"--json",
	})
	if got := latestDriverQuery.Get("driver_id"); got != "driver_fixture_001" {
		t.Fatalf("latest driver_id = %q, want command query filter", got)
	}
	if got := latestDriverQuery.Get("limit"); got != "100" {
		t.Fatalf("latest driver provider limit = %q, want fixed operation default 100 despite global --limit", got)
	}
	if got := latestDriverQuery.Get("page"); got != "" {
		t.Fatalf("latest driver page = %q, want no provider page override in Tier-1 pilot", got)
	}

	unsupported := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "history start-date flag",
			args: []string{"lucid-eld", "vehicle-location-history", "list", "--credential", "lucid-router", "--start-date", "02-01-2026", "--root", root, "--json"},
			want: "unknown flag --start-date",
		},
		{
			name: "drivers page flag",
			args: []string{"lucid-eld", "drivers", "list", "--credential", "lucid-router", "--page", "2", "--root", root, "--json"},
			want: "unknown flag --page",
		},
		{
			name: "vehicles status flag",
			args: []string{"lucid-eld", "vehicles", "list", "--credential", "lucid-router", "--status", "active", "--root", root, "--json"},
			want: "unknown flag --status",
		},
		{
			name: "latest page flag",
			args: []string{"lucid-eld", "latest", "driver", "statuses", "list", "--credential", "lucid-router", "--page", "2", "--root", root, "--json"},
			want: "unknown flag --page",
		},
	}
	for _, tt := range unsupported {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer
			code := Run(tt.args, &outBuf, &errBuf)
			if code == 0 {
				t.Fatalf("Run(%v) code = 0, want rejection for removed provider flag; stdout=%s stderr=%s", tt.args, outBuf.String(), errBuf.String())
			}
			if got := outBuf.String() + errBuf.String(); !strings.Contains(got, tt.want) {
				t.Fatalf("Run(%v) output missing %q:\nstdout=%s\nstderr=%s", tt.args, tt.want, outBuf.String(), errBuf.String())
			}
		})
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
