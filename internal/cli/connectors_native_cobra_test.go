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
	"polymetrics.ai/internal/connectors/certify"
)

// This compile-time reference is intentional RED evidence: the focused test
// checkpoint must fail until connectors and nested certify are native Cobra.
var _ = newConnectorsCobraCommand

type fakeCertifyCommandRuntime struct {
	singleCalls int
	batchCalls  int
	sweepCalls  int

	singleRoot string
	singleOpts certify.Options
	batchRoot  string
	credsPath  string
	parallel   int
	resume     bool
	olderThan  time.Duration

	singleReport certify.Report
	singleErr    error
	credsFile    certify.CredsFile
	loadCredsErr error
	batchReport  certify.BatchReport
	batchErr     error
	sweepResult  map[string]certify.SweepResult
	sweepErr     error
}

func (f *fakeCertifyCommandRuntime) RunSingle(_ context.Context, root string, opts certify.Options) (certify.Report, error) {
	f.singleCalls++
	f.singleRoot = root
	f.singleOpts = opts
	return f.singleReport, f.singleErr
}

func (f *fakeCertifyCommandRuntime) LoadCredsFile(path string) (certify.CredsFile, error) {
	f.credsPath = path
	return f.credsFile, f.loadCredsErr
}

func (f *fakeCertifyCommandRuntime) RunBatch(_ context.Context, root string, credsFile certify.CredsFile, resume bool) (certify.BatchReport, error) {
	f.batchCalls++
	f.batchRoot = root
	f.parallel = credsFile.Defaults.Parallel
	f.resume = resume
	return f.batchReport, f.batchErr
}

func (f *fakeCertifyCommandRuntime) Sweep(_ context.Context, root, credsPath string, olderThan time.Duration) (map[string]certify.SweepResult, error) {
	f.sweepCalls++
	f.batchRoot = root
	f.credsPath = credsPath
	f.olderThan = olderThan
	return f.sweepResult, f.sweepErr
}

func TestConnectorsCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	connectorsCmd := findCobraCommand(root, "connectors")
	if connectorsCmd == nil {
		t.Fatal("missing connectors command")
	}
	if connectorsCmd.DisableFlagParsing {
		t.Fatal("connectors must use native Cobra parsing")
	}
	assertNoFileCompletion(t, connectorsCmd)

	flagsByAction := map[string]map[string]string{
		"list":    {"all": "stringArray"},
		"catalog": {"capability": "stringArray", "stage": "stringArray", "type": "stringArray"},
		"inspect": {},
		"certify": {
			"credential": "stringArray", "from-env": "stringArray", "config": "stringArray",
			"stream": "stringArray", "limit": "stringArray", "modes": "stringArray",
			"skip": "stringArray", "keep-workdir": "stringArray", "write": "stringArray",
			"full": "stringArray", "all": "stringArray", "credentials-file": "stringArray",
			"parallel": "stringArray", "resume": "stringArray", "sweep": "stringArray",
			"older-than": "stringArray",
		},
	}
	for actionName, expectedFlags := range flagsByAction {
		t.Run(actionName, func(t *testing.T) {
			action := findCobraCommand(connectorsCmd, actionName)
			if action == nil {
				t.Fatalf("missing connectors %s", actionName)
			}
			if action.DisableFlagParsing {
				t.Fatalf("connectors %s must be native", actionName)
			}
			if !action.FParseErrWhitelist.UnknownFlags {
				t.Fatalf("connectors %s must preserve unknown flags", actionName)
			}
			assertNoFileCompletion(t, action)
			for flagName, flagType := range expectedFlags {
				flag := action.Flags().Lookup(flagName)
				if flag == nil {
					t.Fatalf("connectors %s missing --%s", actionName, flagName)
				}
				if got := flag.Value.Type(); got != flagType {
					t.Fatalf("connectors %s --%s type=%q want=%q", actionName, flagName, got, flagType)
				}
				if flag.NoOptDefVal != "true" {
					t.Fatalf("connectors %s --%s NoOptDefVal=%q", actionName, flagName, flag.NoOptDefVal)
				}
			}
		})
	}

	help := findCobraCommand(connectorsCmd, "help")
	if help == nil || !help.Hidden {
		t.Fatal("connectors must retain hidden positional namespace help")
	}
	for _, alias := range []string{"man", "docs"} {
		if findCobraCommand(connectorsCmd, alias) == nil {
			t.Fatalf("connectors missing existing %s connector-manual alias", alias)
		}
	}
	for _, spec := range cobraLegacyCommands(config.Config{}) {
		if spec.name == "connectors" {
			t.Fatal("connectors remains registered as a legacy wrapper")
		}
	}
}

func TestNativeConnectorsActionsPreserveFlagsOperandsAndOutput(t *testing.T) {
	stdout, stderr, code := runNativeConnectorsCLI(nil, true, "connectors", "list", "ignored", "--all=true", "--all=false", "--unknown=ignored")
	if code != 0 || stderr != "" {
		t.Fatalf("list: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	assertJSONKind(t, stdout, "ConnectorCatalog")

	stdout, stderr, code = runNativeConnectorsCLI(nil, true, "connectors", "catalog", "ignored", "--capability=read", "--capability=write", "--stage=missing-stage", "--unknown", "ignored")
	if code != 0 || stderr != "" {
		t.Fatalf("catalog: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	var catalog struct {
		Kind       string `json:"kind"`
		Count      int    `json:"count"`
		Connectors []any  `json:"connectors"`
	}
	decodeOneJSON(t, stdout, &catalog)
	if catalog.Kind != "ConnectorCatalog" || catalog.Count != 0 || len(catalog.Connectors) != 0 {
		t.Fatalf("catalog output=%+v", catalog)
	}

	stdout, stderr, code = runNativeConnectorsCLI(nil, true, "connectors", "inspect", "sample", "ignored", "--unknown", "ignored")
	if code != 0 || stderr != "" {
		t.Fatalf("inspect: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	assertJSONKind(t, stdout, "Connector")

	for _, action := range []string{"man", "docs"} {
		stdout, stderr, code = runNativeConnectorsCLI(nil, false, "connectors", action, "sample", "ignored")
		if code != 0 || stderr != "" || !strings.Contains(stdout, "pm connectors inspect sample") {
			t.Fatalf("%s alias: code=%d stderr=%q stdout=%q", action, code, stderr, stdout)
		}
	}
}

func TestNativeConnectorsAndCertifyHelpDiscoveryGlobalsAndMalformedInputs(t *testing.T) {
	canonical := docs["connectors"]
	for _, tt := range []struct {
		name string
		args []string
		json bool
	}{
		{name: "topic", args: []string{"help", "connectors"}},
		{name: "bare", args: []string{"connectors"}},
		{name: "long", args: []string{"connectors", "--help"}},
		{name: "short", args: []string{"connectors", "-h"}},
		{name: "positional", args: []string{"connectors", "help"}},
		{name: "list trailing long", args: []string{"connectors", "list", "--help"}},
		{name: "catalog trailing positional", args: []string{"connectors", "catalog", "help"}},
		{name: "inspect direct long", args: []string{"connectors", "inspect", "--help"}},
		{name: "inspect direct positional", args: []string{"connectors", "inspect", "help"}},
		{name: "inspect trailing long", args: []string{"connectors", "inspect", "sample", "--help"}},
		{name: "certify trailing positional", args: []string{"connectors", "certify", "sample", "help"}},
		{name: "JSON trailing", args: []string{"connectors", "certify", "sample", "--help"}, json: true},
		{name: "JSON long", args: []string{"connectors", "--help"}, json: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr string
			var code int
			if tt.name == "topic" {
				var out bytes.Buffer
				err := runHelp([]string{"connectors"}, &out)
				stdout = out.String()
				if err != nil {
					code = 1
					stderr = err.Error()
				}
			} else {
				stdout, stderr, code = runNativeConnectorsCLI(nil, tt.json, tt.args...)
			}
			if code != 0 || stderr != "" {
				t.Fatalf("help: code=%d stderr=%q stdout=%q", code, stderr, stdout)
			}
			if tt.json {
				var env struct{ Kind, Command, Manual string }
				decodeOneJSON(t, stdout, &env)
				if env.Kind != "CommandManual" || env.Command != "connectors" || env.Manual != canonical {
					t.Fatalf("manual=%+v", env)
				}
			} else if stdout != canonical {
				t.Fatal("text manual differs from canonical")
			}
		})
	}

	for _, args := range [][]string{
		{"connectors", "bogus", "list"},
		{"connectors", "--", "list"},
		{"connectors", "--unknown", "list"},
		{"connectors", "--=x", "list"},
		{"connectors", "---x", "list"},
		{"connectors", "certify", "--", "sample"},
		{"connectors", "certify", "sample", "tail"},
	} {
		stdout, stderr, code := runNativeConnectorsCLI(nil, true, args...)
		assertCLIError(t, code, stdout, stderr, 2, "usage", "")
		if strings.Contains(stdout, "ConnectorList") || strings.Contains(stdout, "ConnectorCertification") {
			t.Fatalf("invalid input discovered later action: %v", args)
		}
	}

	stdout, stderr, code := runNativeConnectorsCLI(nil, true, "connectors", "inspect", "--unknown", "sample")
	assertCLIError(t, code, stdout, stderr, 1, "internal", `connector "--unknown" not found`)

	stdout, stderr, code = runNativeConnectorsCLI(nil, true, "connectors", "catalog", "--type=source")
	assertCLIError(t, code, stdout, stderr, 3, "validation", "legacy --type")
	stdout, stderr, code = runNativeConnectorsCLI(nil, true, "connectors", "inspect", "../sample")
	assertCLIError(t, code, stdout, stderr, 3, "validation", "")
}

func TestNativeCertifyRejectsUnsupportedSafetyAndModeControls(t *testing.T) {
	unsupported := []string{
		"record",
		"replay",
		"allow-production-writes",
		"rate-limit",
		"budget",
		"live-all-modes",
	}

	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	connectorsCmd := findCobraCommand(root, "connectors")
	if connectorsCmd == nil {
		t.Fatal("missing connectors command")
	}
	certifyCmd := findCobraCommand(connectorsCmd, "certify")
	if certifyCmd == nil {
		t.Fatal("missing connectors certify command")
	}
	for _, name := range unsupported {
		flag := certifyCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("unsupported --%s must remain an explicit fail-closed parser control", name)
		}
		if !flag.Hidden {
			t.Errorf("unsupported --%s is advertised in Cobra help", name)
		}
	}

	for _, tt := range []struct {
		name string
		arg  string
	}{
		{name: "record", arg: "--record"},
		{name: "replay", arg: "--replay"},
		{name: "replay false", arg: "--replay=false"},
		{name: "production write allow", arg: "--allow-production-writes"},
		{name: "production write allow false", arg: "--allow-production-writes=false"},
		{name: "rate limit", arg: "--rate-limit=2"},
		{name: "budget", arg: "--budget=50"},
		{name: "live all modes", arg: "--live-all-modes"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			runtime := &fakeCertifyCommandRuntime{singleReport: passingCLIReport("sample")}
			stdout, stderr, code := runNativeConnectorsCLI(runtime, true, "connectors", "certify", "sample", tt.arg)
			assertCLIError(t, code, stdout, stderr, 2, "usage", "not supported")
			if runtime.singleCalls != 0 || runtime.batchCalls != 0 || runtime.sweepCalls != 0 {
				t.Fatalf("unsupported control invoked runtime: single=%d batch=%d sweep=%d", runtime.singleCalls, runtime.batchCalls, runtime.sweepCalls)
			}
		})
	}
}

func TestNativeConnectorsOnlyExactHelpFormsRenderManual(t *testing.T) {
	for _, arg := range []string{"--help=false", "--help=true", "--help=malformed", "-xh", "-hx"} {
		t.Run(arg, func(t *testing.T) {
			stdout, stderr, code := runNativeConnectorsCLI(nil, true, "connectors", "list", arg)
			if code != 0 || stderr != "" {
				t.Fatalf("code=%d stderr=%q stdout=%q", code, stderr, stdout)
			}
			assertJSONKind(t, stdout, "ConnectorList")
			if strings.Contains(stdout, `"kind": "CommandManual"`) {
				t.Fatalf("non-exact help form rendered manual: %s", stdout)
			}
		})
	}
}

func TestConnectorsManualSeparatesCLIAndCompletedCertificationExits(t *testing.T) {
	manual := docs["connectors"]
	for _, required := range []string{
		"Before a certification report is completed",
		"usage errors exit 2",
		"validation errors exit 3",
		"A completed certification report exits 0 for pass, 2 for certification",
		"failure, or 3 for leaked resources",
	} {
		if !strings.Contains(manual, required) {
			t.Errorf("connectors manual missing %q", required)
		}
	}
	if strings.Contains(manual, "1 usage/internal") {
		t.Error("connectors manual still conflates CLI usage and internal exits")
	}
}

func TestNativeCertifyModesCurrentFlagsAndExitContract(t *testing.T) {
	secretValue := "planted-native-certify-secret"
	t.Setenv("PM_CERT_NATIVE_TOKEN", secretValue)
	passing := passingCLIReport("sample")
	runtime := &fakeCertifyCommandRuntime{singleReport: passing}

	stdout, stderr, code := runNativeConnectorsCLI(runtime, true,
		"connectors", "certify", "sample",
		"--stream=ignored", "--stream=customers", "--limit=4", "--limit=7",
		"--modes=full_refresh_append", "--modes=incremental_append,incremental_append_deduped",
		"--skip=flow", "--skip=schedule,write", "--from-env=token=PM_CERT_NATIVE_TOKEN",
		"--config=base_url=https://example.invalid", "--config=account=fixture",
		"--keep-workdir", "--write", "--full", "--credential=ignored", "--unknown=ignored",
	)
	if code != 0 || stderr != "" {
		t.Fatalf("single: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	assertJSONKind(t, stdout, "ConnectorCertification")
	if runtime.singleCalls != 1 || runtime.batchCalls != 0 || runtime.sweepCalls != 0 {
		t.Fatalf("mode calls single=%d batch=%d sweep=%d", runtime.singleCalls, runtime.batchCalls, runtime.sweepCalls)
	}
	if got := runtime.singleOpts; got.Connector != "sample" || got.Stream != "customers" || got.Limit != 7 || !got.KeepWork || got.Write || !got.Full {
		t.Fatalf("single options=%+v", got)
	}
	if strings.Join(runtime.singleOpts.Modes, ",") != "full_refresh_append,incremental_append,incremental_append_deduped" {
		t.Fatalf("modes=%v", runtime.singleOpts.Modes)
	}
	if runtime.singleOpts.SecretEnv["token"] != "PM_CERT_NATIVE_TOKEN" || runtime.singleOpts.Config["account"] != "fixture" {
		t.Fatalf("secret/config refs=%+v/%+v", runtime.singleOpts.SecretEnv, runtime.singleOpts.Config)
	}
	if strings.Contains(stdout+stderr, secretValue) {
		t.Fatal("credential value leaked to CLI output")
	}

	runtime = &fakeCertifyCommandRuntime{batchReport: certify.BatchReport{ExitCode: 2}}
	stdout, stderr, code = runNativeConnectorsCLI(runtime, true, "connectors", "certify", "--all", "--credentials-file=fixture.yaml", "--parallel=2", "--parallel=4", "--resume")
	if code != 2 || stderr != "" {
		t.Fatalf("batch exit 2: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	assertJSONKind(t, stdout, "ConnectorCertificationBatch")
	if runtime.batchCalls != 1 || runtime.credsPath != "fixture.yaml" || runtime.parallel != 4 || !runtime.resume {
		t.Fatalf("batch call=%+v", runtime)
	}

	runtime = &fakeCertifyCommandRuntime{batchErr: errors.New("fixture batch failure")}
	stdout, stderr, code = runNativeConnectorsCLI(runtime, true, "connectors", "certify", "--all", "--credentials-file=fixture.yaml")
	assertCLIError(t, code, stdout, stderr, 1, "internal", "certify: batch run failed: fixture batch failure")

	runtime = &fakeCertifyCommandRuntime{sweepResult: map[string]certify.SweepResult{"sample": {Failed: map[string]string{"fixture-tag": "cleanup failed"}}}}
	stdout, stderr, code = runNativeConnectorsCLI(runtime, true, "connectors", "certify", "--sweep", "--credentials-file=fixture.yaml", "--older-than=2h")
	if code != 3 || stderr != "" {
		t.Fatalf("sweep exit 3: code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	assertJSONKind(t, stdout, "ConnectorCertificationSweep")
	if runtime.sweepCalls != 1 || runtime.olderThan != 2*time.Hour {
		t.Fatalf("sweep call=%+v", runtime)
	}

	runtime = &fakeCertifyCommandRuntime{singleErr: errors.New("fixture internal failure")}
	stdout, stderr, code = runNativeConnectorsCLI(runtime, true, "connectors", "certify", "sample")
	assertCLIError(t, code, stdout, stderr, 1, "internal", "fixture internal failure")
}

func TestNativeCertifyFreshTreesPreserveContextCancellationAndNoCredentialOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	runtime := &fakeCertifyCommandRuntime{singleErr: ctx.Err()}
	stdout, stderr, code := executeNativeConnectors(ctx, testRouterConfig(t.TempDir(), true), runtime, "connectors", "certify", "sample")
	assertCLIError(t, code, stdout, stderr, 1, "internal", "context canceled")

	for i := 0; i < 2; i++ {
		runtime = &fakeCertifyCommandRuntime{singleReport: passingCLIReport("sample")}
		stdout, stderr, code = executeNativeConnectors(context.Background(), testRouterConfig(t.TempDir(), true), runtime, "connectors", "certify", "sample", "--limit="+string(rune('1'+i)))
		if code != 0 || stderr != "" || runtime.singleCalls != 1 {
			t.Fatalf("fresh invocation %d: code=%d stderr=%q stdout=%q calls=%d", i, code, stderr, stdout, runtime.singleCalls)
		}
		assertJSONKind(t, stdout, "ConnectorCertification")
	}
}

func passingCLIReport(connector string) certify.Report {
	return certify.Report{
		Kind:          "ConnectorCertification",
		SchemaVersion: 1,
		Connector:     connector,
		Passed:        true,
		Capabilities: certify.Capabilities{
			Check:           certify.CapabilityResult{Result: "pass"},
			Catalog:         certify.CapabilityResult{Result: "pass"},
			Read:            certify.CapabilityResult{Result: "pass"},
			Resume:          certify.CapabilityResult{Result: "pass"},
			JSONContract:    certify.CapabilityResult{Result: "pass"},
			SecretRedaction: certify.CapabilityResult{Result: "pass"},
		},
	}
}

func runNativeConnectorsCLI(runtime *fakeCertifyCommandRuntime, jsonOut bool, args ...string) (string, string, int) {
	return executeNativeConnectors(context.Background(), testRouterConfig(".", jsonOut), runtime, args...)
}

func executeNativeConnectors(ctx context.Context, cfg config.Config, runtime *fakeCertifyCommandRuntime, args ...string) (string, string, int) {
	if runtime == nil {
		runtime = &fakeCertifyCommandRuntime{singleReport: passingCLIReport("sample")}
	}
	var stdout, stderr bytes.Buffer
	root := &cobra.Command{
		Use:                "pm",
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		SilenceErrors:      true,
		SilenceUsage:       true,
		RunE:               func(_ *cobra.Command, _ []string) error { return errUsage },
	}
	root.SetContext(ctx)
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.AddCommand(newConnectorsCobraCommandWithRuntime(ctx, cfg.Root, &stdout, cfg.JSON, runtime))
	err := executeRootCmd(root, args)
	code := writeError(ctx, &stdout, &stderr, mapCobraErr(err), cfg.JSON)
	return stdout.String(), stderr.String(), code
}

func assertNoFileCompletion(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	if cmd.ValidArgsFunction == nil {
		t.Fatalf("%s missing completion seam", cmd.CommandPath())
	}
	values, directive := cmd.ValidArgsFunction(cmd, nil, "")
	if len(values) != 0 || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("%s completion=%v/%v", cmd.CommandPath(), values, directive)
	}
}
