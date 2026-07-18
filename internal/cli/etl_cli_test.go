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

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
)

// This compile-time reference is intentional TDD evidence: the focused test
// checkpoint must fail until ETL has a native Cobra constructor.
var _ = newETLCobraCommand

func TestETLCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	etl := findCobraCommand(root, "etl")
	if etl == nil {
		t.Fatal("missing etl command")
	}
	if etl.DisableFlagParsing {
		t.Fatal("etl command must use native Cobra flag parsing")
	}
	if etl.ValidArgsFunction == nil {
		t.Fatal("etl command must suppress file completion until Phase 15")
	}
	values, directive := etl.ValidArgsFunction(etl, nil, "")
	if len(values) != 0 || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("etl completion = (%v, %v), want no values and NoFileComp", values, directive)
	}

	flagsByAction := map[string][]string{
		"check":   {"connector", "config"},
		"catalog": {"connector", "config"},
		"read":    {"connector", "stream", "limit", "config"},
		"run":     {"connection", "stream", "batch-size", "runtime"},
		"status":  nil,
	}
	for actionName, flagNames := range flagsByAction {
		t.Run(actionName, func(t *testing.T) {
			action := findCobraCommand(etl, actionName)
			if action == nil {
				t.Fatalf("missing etl %s command", actionName)
			}
			if action.DisableFlagParsing {
				t.Fatalf("etl %s must use native Cobra flag parsing", actionName)
			}
			if !action.FParseErrWhitelist.UnknownFlags {
				t.Fatalf("etl %s must preserve unknown-flag tolerance", actionName)
			}
			if action.ValidArgsFunction == nil {
				t.Fatalf("etl %s missing no-file completion seam", actionName)
			}
			for _, flagName := range flagNames {
				flag := action.Flags().Lookup(flagName)
				if flag == nil {
					t.Fatalf("etl %s missing --%s", actionName, flagName)
				}
				if got, want := flag.Value.Type(), "stringArray"; got != want {
					t.Fatalf("etl %s --%s type = %q, want %q", actionName, flagName, got, want)
				}
				if got, want := flag.NoOptDefVal, "true"; got != want {
					t.Fatalf("etl %s --%s NoOptDefVal = %q, want %q", actionName, flagName, got, want)
				}
			}
		})
	}

	help := findCobraCommand(etl, "help")
	if help == nil || !help.Hidden {
		t.Fatal("etl must preserve hidden positional help until Phase 19")
	}
	for _, spec := range cobraLegacyCommands(config.Config{}) {
		if spec.name == "etl" {
			t.Fatal("etl remains registered as a legacy Cobra wrapper")
		}
	}
}

func TestETLDirectActionsPreserveCurrentFlagForms(t *testing.T) {
	root := initETLProject(t)

	stdout, stderr, code := runETLCLI(
		"etl", "check", "extra-positional",
		"--connector", "warehouse", "--connector=sample",
		"--config", "mode=first", "--config=mode=fixture",
		"--unknown", "ignored", "--root", root, "--json",
	)
	if code != 0 || stderr != "" {
		t.Fatalf("etl check code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	assertJSONKind(t, stdout, "ETLCheck")
	if !strings.Contains(stdout, `"connector": "sample"`) {
		t.Fatalf("etl check did not use repeated last connector: %s", stdout)
	}

	stdout, stderr, code = runETLCLI(
		"etl", "catalog", "--connector=warehouse", "--connector", "sample",
		"--config=mode=first", "--config", "mode=fixture",
		"--root="+root, "--json",
	)
	if code != 0 || stderr != "" {
		t.Fatalf("etl catalog code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	assertJSONKind(t, stdout, "ETLCatalog")
	if !strings.Contains(stdout, `"name": "customers"`) {
		t.Fatalf("etl catalog missing sample customers stream: %s", stdout)
	}

	stdout, stderr, code = runETLCLI(
		"etl", "read", "--",
		"--connector", "warehouse", "--connector=sample",
		"--stream", "events", "--stream=customers",
		"--limit", "2", "--limit=1",
		"--config", "mode=fixture", "--unknown=ignored",
		"--root", root, "--json",
	)
	if code != 0 || stderr != "" {
		t.Fatalf("etl read code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	var readEnv struct {
		Kind      string `json:"kind"`
		Connector string `json:"connector"`
		Stream    string `json:"stream"`
		Count     int    `json:"count"`
	}
	decodeOneJSON(t, stdout, &readEnv)
	if readEnv.Kind != "ETLRead" || readEnv.Connector != "sample" || readEnv.Stream != "customers" || readEnv.Count != 1 {
		t.Fatalf("etl read envelope = %+v", readEnv)
	}
}

func TestETLRunStatusBatchBoundsAndRepeatedFlags(t *testing.T) {
	root := setupETLProject(t, "full_refresh_overwrite")

	stdout, stderr, code := runETLCLI(
		"etl", "run", "extra-positional",
		"--connection", "missing", "--connection=sample-to-warehouse",
		"--stream", "events", "--stream=customers",
		"--batch-size", "1", "--batch-size=2",
		"--runtime", "true", "--runtime=false",
		"--unknown", "ignored", "--root", root, "--json",
	)
	if code != 0 || stderr != "" {
		t.Fatalf("etl run code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	var runEnv struct {
		Kind            string  `json:"kind"`
		Run             app.Run `json:"run"`
		RuntimeRecorded bool    `json:"runtime_recorded"`
	}
	decodeOneJSON(t, stdout, &runEnv)
	if runEnv.Kind != "ETLRun" || runEnv.Run.Status != "completed" || runEnv.Run.RecordsRead != 3 || runEnv.Run.BatchCount != 2 || runEnv.RuntimeRecorded {
		t.Fatalf("explicit bounded ETL run = %+v", runEnv)
	}

	stdout, stderr, code = runETLCLI("etl", "status", runEnv.Run.ID, "ignored-positional", "--unknown=ignored", "--root", root, "--json")
	if code != 0 || stderr != "" {
		t.Fatalf("etl status code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	var statusEnv struct {
		Kind string  `json:"kind"`
		Run  app.Run `json:"run"`
	}
	decodeOneJSON(t, stdout, &statusEnv)
	if statusEnv.Kind != "ETLRun" || statusEnv.Run.ID != runEnv.Run.ID || statusEnv.Run.BatchCount != 2 {
		t.Fatalf("etl status envelope = %+v", statusEnv)
	}

	stdout, stderr, code = runETLCLI("etl", "run", "--connection", "sample-to-warehouse", "--stream", "customers", "--root", root, "--json")
	if code != 0 || stderr != "" {
		t.Fatalf("default batch ETL run code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	decodeOneJSON(t, stdout, &runEnv)
	if runEnv.Run.RecordsRead != 3 || runEnv.Run.BatchCount != 1 {
		t.Fatalf("default batch ETL run = %+v, want one bounded default batch", runEnv.Run)
	}
}

func TestETLBatchSizeAndConfiguredSyncValidation(t *testing.T) {
	root := setupETLProject(t, "full_refresh_overwrite")

	stdout, stderr, code := runETLCLI("etl", "run", "--connection", "sample-to-warehouse", "--stream", "customers", "--batch-size", "not-an-int", "--root", root, "--json")
	assertCLIError(t, code, stdout, stderr, 3, "validation", "invalid --batch-size")

	stdout, stderr, code = runETLCLI("etl", "run", "--connection", "missing", "--stream", "customers", "--batch-size", "--runtime", "--root", root, "--json")
	assertCLIError(t, code, stdout, stderr, 3, "validation", `invalid --batch-size "true"`)

	setETLStreamConfig(t, root, app.StreamConfig{
		SyncMode:         "incremental_append_deduped",
		DestinationTable: "customers",
	})
	stdout, stderr, code = runETLCLI("etl", "run", "--connection", "sample-to-warehouse", "--stream", "customers", "--root", root, "--json")
	if code == 0 || !strings.Contains(stdout+stderr, "requires a cursor field") {
		t.Fatalf("configured sync validation code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
}

func TestETLHelpRoutesAndLegacyActionTailCompatibility(t *testing.T) {
	var canonical string
	for _, tt := range []struct {
		name string
		args []string
	}{
		{name: "help topic", args: []string{"help", "etl"}},
		{name: "bare", args: []string{"etl"}},
		{name: "long", args: []string{"etl", "--help"}},
		{name: "short", args: []string{"etl", "-h"}},
		{name: "positional", args: []string{"etl", "help"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, code := runETLCLI(tt.args...)
			if code != 0 || stderr != "" || !strings.Contains(stdout, "pm etl - run local ETL syncs") {
				t.Fatalf("help %s code=%d stdout=%s stderr=%s", tt.name, code, stdout, stderr)
			}
			if canonical == "" {
				canonical = stdout
			} else if stdout != canonical {
				t.Fatalf("help %s differs from canonical manual", tt.name)
			}
		})
	}

	stdout, stderr, code := runETLCLI("etl", "--json")
	if code != 0 || stderr != "" {
		t.Fatalf("JSON manual code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	var manual struct {
		Kind    string `json:"kind"`
		Command string `json:"command"`
		Manual  string `json:"manual"`
	}
	decodeOneJSON(t, stdout, &manual)
	if manual.Kind != "CommandManual" || manual.Command != "etl" || manual.Manual != canonical {
		t.Fatalf("JSON manual = kind %q command %q", manual.Kind, manual.Command)
	}

	root := initETLProject(t)
	for _, args := range [][]string{
		{"etl", "run", "--help", "--root", root, "--json"},
		{"etl", "run", "-h", "--root", root, "--json"},
		{"etl", "run", "--", "--root", root, "--json"},
	} {
		stdout, stderr, code = runETLCLI(args...)
		assertCLIError(t, code, stdout, stderr, 1, "internal", `connection "" not found`)
	}
}

func TestETLUnknownInvalidActionsGlobalsAndNoDiscoveryBypass(t *testing.T) {
	for _, args := range [][]string{
		{"etl", "bogus", "run", "--json"},
		{"etl", "--unknown", "run", "--connection", "missing", "--stream", "customers", "--json"},
		{"etl", "--", "run", "--connection", "missing", "--stream", "customers", "--json"},
	} {
		stdout, stderr, code := runETLCLI(args...)
		assertCLIError(t, code, stdout, stderr, 2, "usage", "")
		if strings.Contains(stdout, `connection "missing" not found`) {
			t.Fatalf("invalid action discovered and executed later run: args=%v stdout=%s", args, stdout)
		}
	}

	stdout, stderr, code := runETLCLI("--json", "--json=maybe", "etl")
	assertCLIError(t, code, stdout, stderr, 3, "validation", "invalid --json")

	stdout, stderr, code = runETLCLI("--json=false", "--plain=true", "--no-input=on", "etl")
	if code != 0 || stderr != "" || !strings.HasPrefix(stdout, "NAME\n  pm etl") {
		t.Fatalf("valid assigned global booleans code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
}

func TestETLBareFlagCompatibility(t *testing.T) {
	root := initETLProject(t)
	for _, tt := range []struct {
		name       string
		args       []string
		code       int
		category   string
		messageSub string
	}{
		{name: "connector", args: []string{"etl", "check", "--connector", "--root", root, "--json"}, code: 1, category: "internal", messageSub: `connector "true" not found`},
		{name: "read limit", args: []string{"etl", "read", "--connector=sample", "--limit", "--root", root, "--json"}, code: 3, category: "validation", messageSub: `invalid --limit "true"`},
		{name: "run connection", args: []string{"etl", "run", "--connection", "--stream=customers", "--root", root, "--json"}, code: 1, category: "internal", messageSub: `connection "true" not found`},
		{name: "run runtime", args: []string{"etl", "run", "--connection=missing", "--stream=customers", "--runtime", "--root", root, "--json"}, code: 1, category: "internal", messageSub: `connection "missing" not found`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, code := runETLCLI(tt.args...)
			assertCLIError(t, code, stdout, stderr, tt.code, tt.category, tt.messageSub)
		})
	}
}

func TestETLCancellationPropagatesThroughNativeCommand(t *testing.T) {
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout bytes.Buffer
	cmd := newRootCmd(ctx, config.Config{Root: root}, &stdout, io.Discard)
	err := executeRootCmd(cmd, []string{"etl", "check", "--connector", "sample"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("native ETL cancellation error = %v, want context.Canceled", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("canceled ETL wrote stdout: %q", stdout.String())
	}
}

func TestETLProgressPreservesStdoutStderrAndEnvelopeSemantics(t *testing.T) {
	root := setupETLProject(t, "full_refresh_overwrite")
	stdout, stderr, code := runETLCLI(
		"--root", root, "--json", "--progress", "ndjson",
		"etl", "run", "--connection", "sample-to-warehouse", "--stream", "customers", "--batch-size", "2",
	)
	if code != 0 {
		t.Fatalf("progress ETL code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	var env struct {
		Kind string  `json:"kind"`
		Run  app.Run `json:"run"`
	}
	decodeOneJSON(t, stdout, &env)
	if env.Kind != "ETLRun" || env.Run.BatchCount != 2 {
		t.Fatalf("progress ETL final envelope = %+v", env)
	}
	if strings.Contains(stdout, `"scope":"etl"`) || strings.Contains(stdout, `"kind":"progress"`) {
		t.Fatalf("stdout contains ETL progress event: %s", stdout)
	}
	var sawStarted, sawProgress, sawCompleted bool
	for i, line := range strings.Split(strings.TrimSpace(stderr), "\n") {
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("stderr line %d is not NDJSON: %v; line=%q", i, err, line)
		}
		if event["scope"] != "etl" {
			t.Fatalf("stderr line %d scope = %v, want etl", i, event["scope"])
		}
		switch event["kind"] {
		case "started":
			sawStarted = true
		case "progress":
			sawProgress = true
		case "completed":
			sawCompleted = true
		}
	}
	if !sawStarted || !sawProgress || !sawCompleted {
		t.Fatalf("ETL progress lifecycle started=%t progress=%t completed=%t stderr=%s", sawStarted, sawProgress, sawCompleted, stderr)
	}
}

func initETLProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	stdout, stderr, code := runETLCLI("init", "--root", root, "--json")
	if code != 0 {
		t.Fatalf("init ETL project code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	return root
}

func setupETLProject(t *testing.T, syncMode string) string {
	t.Helper()
	root := initETLProject(t)
	for _, args := range [][]string{
		{"credentials", "add", "sample-local", "--connector", "sample", "--root", root, "--json"},
		{"credentials", "add", "warehouse-local", "--connector", "warehouse", "--config", "path=.polymetrics/warehouse", "--root", root, "--json"},
		{"connections", "create", "sample-to-warehouse", "--source", "sample:sample-local", "--destination", "warehouse:warehouse-local", "--stream", "customers", "--sync-mode", syncMode, "--table", "customers", "--root", root, "--json"},
	} {
		stdout, stderr, code := runETLCLI(args...)
		if code != 0 {
			t.Fatalf("setup ETL Run(%v) code=%d stdout=%s stderr=%s", args, code, stdout, stderr)
		}
	}
	return root
}

func setETLStreamConfig(t *testing.T, root string, stream app.StreamConfig) {
	t.Helper()
	path := filepath.Join(root, ".polymetrics", "state", "state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	connections, ok := state["connections"].([]any)
	if !ok || len(connections) != 1 {
		t.Fatalf("unexpected state connections: %#v", state["connections"])
	}
	connection := connections[0].(map[string]any)
	streams := connection["streams"].(map[string]any)
	value, err := json.Marshal(stream)
	if err != nil {
		t.Fatal(err)
	}
	var streamMap map[string]any
	if err := json.Unmarshal(value, &streamMap); err != nil {
		t.Fatal(err)
	}
	streams["customers"] = streamMap
	data, err = json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func runETLCLI(args ...string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func decodeOneJSON(t *testing.T, stdout string, target any) {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(stdout))
	if err := dec.Decode(target); err != nil {
		t.Fatalf("decode JSON: %v; stdout=%s", err, stdout)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		t.Fatalf("stdout contains more than one JSON value: err=%v extra=%v stdout=%s", err, extra, stdout)
	}
}

func assertJSONKind(t *testing.T, stdout, want string) {
	t.Helper()
	var env struct {
		Kind string `json:"kind"`
	}
	decodeOneJSON(t, stdout, &env)
	if env.Kind != want {
		t.Fatalf("JSON kind = %q, want %q; stdout=%s", env.Kind, want, stdout)
	}
}

func assertCLIError(t *testing.T, code int, stdout, stderr string, wantCode int, category, messageSub string) {
	t.Helper()
	if code != wantCode {
		t.Fatalf("error code=%d, want %d; stdout=%s stderr=%s", code, wantCode, stdout, stderr)
	}
	var env struct {
		Kind  string `json:"kind"`
		Error struct {
			Category string `json:"category"`
			Message  string `json:"message"`
		} `json:"error"`
	}
	decodeOneJSON(t, stdout, &env)
	if env.Kind != "Error" || env.Error.Category != category {
		t.Fatalf("error envelope=%+v, want category %q; stdout=%s", env, category, stdout)
	}
	if messageSub != "" && (!strings.Contains(env.Error.Message, messageSub) || !strings.Contains(stderr, messageSub)) {
		t.Fatalf("error missing %q; stdout=%s stderr=%s", messageSub, stdout, stderr)
	}
}
