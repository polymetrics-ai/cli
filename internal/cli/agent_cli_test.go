package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/config"
)

type agentImageRuntimeCall struct {
	bin  string
	args []string
}

type fakeAgentImageRuntime struct {
	lookPathErr error
	fileErr     error
	runErrs     []error
	lookups     []string
	files       []string
	calls       []agentImageRuntimeCall
	ctxKey      any
	ctxValue    any
}

func (f *fakeAgentImageRuntime) LookPath(bin string) (string, error) {
	f.lookups = append(f.lookups, bin)
	if f.lookPathErr != nil {
		return "", f.lookPathErr
	}
	return "/fake/" + filepath.Base(bin), nil
}

func (f *fakeAgentImageRuntime) FileExists(path string) error {
	f.files = append(f.files, path)
	return f.fileErr
}

func (f *fakeAgentImageRuntime) Run(ctx context.Context, bin string, args []string, stdout, stderr io.Writer) error {
	f.calls = append(f.calls, agentImageRuntimeCall{bin: bin, args: append([]string(nil), args...)})
	if f.ctxKey != nil && ctx.Value(f.ctxKey) != f.ctxValue {
		return fmt.Errorf("context value not propagated")
	}
	if len(f.runErrs) == 0 {
		return nil
	}
	err := f.runErrs[0]
	f.runErrs = f.runErrs[1:]
	return err
}

func TestAgentCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), agentTestConfig(".", false), io.Discard, io.Discard)
	agent := findCobraCommand(root, "agent")
	if agent == nil {
		t.Fatal("missing agent command")
	}
	if agent.DisableFlagParsing {
		t.Fatal("agent command must use native Cobra flag parsing")
	}

	plan := findCobraCommand(agent, "plan")
	if plan == nil || plan.DisableFlagParsing {
		t.Fatal("agent plan must be a native Cobra command")
	}
	if !plan.FParseErrWhitelist.UnknownFlags {
		t.Fatal("agent plan must preserve legacy unknown-flag tolerance")
	}
	if plan.ValidArgsFunction == nil {
		t.Fatal("agent plan must suppress file completion fallback until Phase 15")
	}
	completions, directive := plan.ValidArgsFunction(plan, nil, "")
	if len(completions) != 0 || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("agent plan completion = %v/%v, want no values/NoFileComp", completions, directive)
	}
	request := plan.Flags().Lookup("request")
	if request == nil {
		t.Fatal("agent plan missing native --request flag")
	}
	if got, want := request.Value.Type(), "stringArray"; got != want {
		t.Fatalf("agent plan --request type = %q, want %q", got, want)
	}
	if got, want := request.NoOptDefVal, "true"; got != want {
		t.Fatalf("agent plan --request NoOptDefVal = %q, want %q", got, want)
	}

	image := findCobraCommand(agent, "image")
	if image == nil || image.DisableFlagParsing {
		t.Fatal("agent image must be a native Cobra command")
	}
	for _, action := range []string{"build", "pull", "ensure"} {
		cmd := findCobraCommand(image, action)
		if cmd == nil || cmd.DisableFlagParsing {
			t.Fatalf("agent image %s must be a native Cobra command", action)
		}
		if !cmd.FParseErrWhitelist.UnknownFlags {
			t.Fatalf("agent image %s must preserve legacy unknown-flag tolerance", action)
		}
		if cmd.ValidArgsFunction == nil {
			t.Fatalf("agent image %s must suppress file completion fallback", action)
		}
	}

	help := findCobraCommand(agent, "help")
	if help == nil || !help.Hidden {
		t.Fatal("agent must preserve a hidden positional help alias until Phase 19")
	}
}

func TestAgentPlanPreservesFlagsAndDeterministicOutput(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "default", args: []string{"agent", "plan"}, want: "pm connectors list --json\npm help etl\n"},
		{name: "spaced", args: []string{"agent", "plan", "--request", "sample customers"}, want: sampleAgentPlanText()},
		{name: "assigned", args: []string{"agent", "plan", "--request=sample customers"}, want: sampleAgentPlanText()},
		{name: "repeated last wins", args: []string{"agent", "plan", "--request", "ignored", "--request", "sample customers"}, want: sampleAgentPlanText()},
		{name: "bare request", args: []string{"agent", "plan", "--request"}, want: "pm connectors list --json\npm help etl\n"},
		{name: "unknown and positional", args: []string{"agent", "plan", "extra", "--unknown", "value", "--request", "sample customers"}, want: sampleAgentPlanText()},
		{name: "literal separator continues", args: []string{"agent", "plan", "--request", "ignored", "--", "--request", "sample customers"}, want: sampleAgentPlanText()},
		{name: "trailing long help ignored", args: []string{"agent", "plan", "--request", "sample customers", "--help"}, want: sampleAgentPlanText()},
		{name: "trailing short help ignored", args: []string{"agent", "plan", "--request", "sample customers", "-h"}, want: sampleAgentPlanText()},
		{name: "positional action help ignored", args: []string{"agent", "plan", "help", "--request", "sample customers"}, want: sampleAgentPlanText()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first, _ := runAgentCLI(t, tt.args)
			second, _ := runAgentCLI(t, tt.args)
			if first != tt.want {
				t.Fatalf("stdout = %q, want %q", first, tt.want)
			}
			if second != first {
				t.Fatalf("nondeterministic output: first=%q second=%q", first, second)
			}
		})
	}

	stdout, _ := runAgentCLI(t, []string{"agent", "plan", "--request", "sample customers", "--json"})
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if got["kind"] != "AgentPlan" || got["risk"] != "read_local" || got["api_version"] != apiVersion {
		t.Fatalf("agent JSON envelope = %#v", got)
	}
	steps, ok := got["steps"].([]any)
	if !ok || len(steps) != 4 {
		t.Fatalf("agent JSON steps = %#v, want four deterministic steps", got["steps"])
	}
}

func TestAgentHelpInvalidActionsAndAssignedGlobals(t *testing.T) {
	manualArgs := [][]string{
		{"agent"},
		{"help", "agent"},
		{"agent", "--help"},
		{"agent", "-h"},
		{"agent", "help"},
	}
	for _, args := range manualArgs {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			stdout, _ := runAgentCLI(t, args)
			if stdout != agentHelp {
				t.Fatalf("manual bytes differ for %v", args)
			}
		})
	}

	for _, args := range [][]string{
		{"agent", "bogus"},
		{"agent", "bogus", "--help"},
		{"agent", "image"},
	} {
		var stdout, stderr bytes.Buffer
		code := Run(args, &stdout, &stderr)
		if code != 2 {
			t.Fatalf("Run(%v) exit = %d, want usage 2; stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
		}
	}

	root := t.TempDir()
	initProject(t, root)
	configPath := filepath.Join(root, ".polymetrics", "config.yaml")
	if err := writeTestFile(configPath, "json: true\n"); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"agent", "plan", "--root=" + root, "--json=false", "--plain=false", "--no-input=true"}, &stdout, &stderr)
	if code != 0 || strings.HasPrefix(strings.TrimSpace(stdout.String()), "{") {
		t.Fatalf("assigned globals exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"agent", "plan", "--json=not-bool"}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("invalid assigned boolean exit=%d, want validation 3", code)
	}
}

func TestAgentRejectsUnsafeRequestBeforePlanning(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"agent", "plan", "--request", "sample\x1bcustomers", "--json"}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("unsafe request exit=%d, want validation 3; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "control characters") {
		t.Fatalf("unsafe request JSON = %q", stdout.String())
	}
}

func TestAgentImageActionsUseInjectedRuntime(t *testing.T) {
	root := t.TempDir()
	containerfile := filepath.Join(root, "build", "agent", "Containerfile")
	ctxKey := struct{}{}
	ctx := context.WithValue(context.Background(), ctxKey, "kept")

	tests := []struct {
		name       string
		action     string
		jsonOut    bool
		runErrs    []error
		wantCalls  []agentImageRuntimeCall
		wantFiles  []string
		wantOutput string
	}{
		{
			name:      "build text",
			action:    "build",
			wantCalls: []agentImageRuntimeCall{{bin: "fake-podman", args: []string{"build", "-f", containerfile, "-t", "ghcr.io/example/agent:test", filepath.Dir(containerfile)}}},
			wantFiles: []string{containerfile}, wantOutput: "AgentImageBuild ok: ghcr.io/example/agent:test\n",
		},
		{
			name:       "pull json",
			action:     "pull",
			jsonOut:    true,
			wantCalls:  []agentImageRuntimeCall{{bin: "fake-podman", args: []string{"pull", "ghcr.io/example/agent:test"}}},
			wantOutput: "{\n  \"api_version\": \"polymetrics.ai/v1\",\n  \"image\": \"ghcr.io/example/agent:test\",\n  \"kind\": \"AgentImagePull\",\n  \"status\": \"ok\"\n}\n",
		},
		{
			name:       "ensure present text",
			action:     "ensure",
			wantCalls:  []agentImageRuntimeCall{{bin: "fake-podman", args: []string{"image", "exists", "ghcr.io/example/agent:test"}}},
			wantOutput: "agent image present: ghcr.io/example/agent:test\n",
		},
		{
			name:    "ensure absent pulls json",
			action:  "ensure",
			jsonOut: true,
			runErrs: []error{errors.New("absent"), nil},
			wantCalls: []agentImageRuntimeCall{
				{bin: "fake-podman", args: []string{"image", "exists", "ghcr.io/example/agent:test"}},
				{bin: "fake-podman", args: []string{"pull", "ghcr.io/example/agent:test"}},
			},
			wantOutput: "{\n  \"api_version\": \"polymetrics.ai/v1\",\n  \"image\": \"ghcr.io/example/agent:test\",\n  \"kind\": \"AgentImageEnsure\",\n  \"status\": \"ok\"\n}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeAgentImageRuntime{runErrs: append([]error(nil), tt.runErrs...), ctxKey: ctxKey, ctxValue: "kept"}
			var stdout bytes.Buffer
			err := runAgentImageAction(ctx, agentTestConfig(root, tt.jsonOut), root, tt.action, &stdout, tt.jsonOut, fake)
			if err != nil {
				t.Fatalf("runAgentImageAction: %v", err)
			}
			if got := stdout.String(); got != tt.wantOutput {
				t.Fatalf("stdout = %q, want %q", got, tt.wantOutput)
			}
			if !reflect.DeepEqual(fake.lookups, []string{"fake-podman"}) {
				t.Fatalf("lookups = %#v", fake.lookups)
			}
			if !reflect.DeepEqual(fake.files, tt.wantFiles) {
				t.Fatalf("file checks = %#v, want %#v", fake.files, tt.wantFiles)
			}
			if !reflect.DeepEqual(fake.calls, tt.wantCalls) {
				t.Fatalf("calls = %#v, want %#v", fake.calls, tt.wantCalls)
			}
		})
	}
}

func TestAgentLeadingInvalidActionTokensCannotReachImageRuntime(t *testing.T) {
	root := t.TempDir()
	levels := []struct {
		name   string
		prefix []string
	}{
		{name: "agent", prefix: []string{"agent"}},
		{name: "image", prefix: []string{"agent", "image"}},
	}
	leadingTokens := []struct {
		name string
		args []string
	}{
		{name: "assigned unknown", args: []string{"--unknown=x"}},
		{name: "bare unknown", args: []string{"--unknown"}},
		{name: "short unknown", args: []string{"-x"}},
		{name: "assigned help-like", args: []string{"--help=false"}},
		{name: "literal boundary", args: []string{"--"}},
	}

	for _, level := range levels {
		for _, leading := range leadingTokens {
			for _, action := range []string{"build", "pull", "ensure"} {
				name := strings.Join([]string{level.name, leading.name, action}, "/")
				t.Run(name, func(t *testing.T) {
					args := append([]string(nil), level.prefix...)
					args = append(args, leading.args...)
					if level.name == "agent" {
						args = append(args, "image")
					}
					args = append(args, action)

					fake := &fakeAgentImageRuntime{}
					var stdout, stderr bytes.Buffer
					cmd := newRootCmdWithAgentImageRuntime(context.Background(), agentTestConfig(root, false), &stdout, &stderr, fake)
					err := executeRootCmd(cmd, args)
					if err == nil || exitCodeFor(classifyError(mapCobraErr(err))) != 2 {
						t.Fatalf("executeRootCmd(%v) err=%v, want usage error", args, err)
					}
					if len(fake.lookups) != 0 || len(fake.files) != 0 || len(fake.calls) != 0 {
						t.Fatalf("invalid action head reached runtime: lookups=%v files=%v calls=%v", fake.lookups, fake.files, fake.calls)
					}
					if stdout.Len() != 0 {
						t.Fatalf("invalid action head leaked output: %q", stdout.String())
					}
				})
			}
		}
	}
}

func TestAgentImageNativeActionsPreserveUnknownFlagsHelpAndSeparator(t *testing.T) {
	root := t.TempDir()
	for _, args := range [][]string{
		{"agent", "image", "pull", "--unknown", "value"},
		{"agent", "image", "pull", "--help"},
		{"agent", "image", "pull", "-h"},
		{"agent", "image", "pull", "--", "--unknown", "value"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			fake := &fakeAgentImageRuntime{}
			var stdout, stderr bytes.Buffer
			cmd := newRootCmdWithAgentImageRuntime(context.Background(), agentTestConfig(root, false), &stdout, &stderr, fake)
			if err := executeRootCmd(cmd, args); err != nil {
				t.Fatalf("executeRootCmd(%v): %v", args, err)
			}
			if len(fake.calls) != 1 || !reflect.DeepEqual(fake.calls[0].args, []string{"pull", "ghcr.io/example/agent:test"}) {
				t.Fatalf("calls for %v = %#v", args, fake.calls)
			}
		})
	}
}

func TestAgentImageInvalidActionsPreserveLegacyHelpTreatment(t *testing.T) {
	root := t.TempDir()
	for _, args := range [][]string{
		{"agent", "image", "--help"},
		{"agent", "image", "-h"},
		{"agent", "image", "frobnicate", "--help"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			fake := &fakeAgentImageRuntime{}
			var stdout, stderr bytes.Buffer
			cmd := newRootCmdWithAgentImageRuntime(context.Background(), agentTestConfig(root, false), &stdout, &stderr, fake)
			err := executeRootCmd(cmd, args)
			if err == nil || exitCodeFor(classifyError(mapCobraErr(err))) != 2 {
				t.Fatalf("executeRootCmd(%v) err=%v, want usage error", args, err)
			}
			if !reflect.DeepEqual(fake.lookups, []string{"fake-podman"}) || len(fake.calls) != 0 {
				t.Fatalf("legacy invalid action order changed: lookups=%v calls=%v", fake.lookups, fake.calls)
			}
			if stdout.Len() != 0 {
				t.Fatalf("invalid action leaked help/output: %q", stdout.String())
			}
		})
	}
}

func TestAgentImageValidationPreventsRuntimeExecution(t *testing.T) {
	root := t.TempDir()
	tests := []struct {
		name   string
		cfg    config.Config
		root   string
		action string
	}{
		{name: "unsafe root path", cfg: agentTestConfig(root, false), root: root + "\n", action: "build"},
		{name: "unsafe podman binary", cfg: func() config.Config { c := agentTestConfig(root, false); c.RLM.PodmanBin = "podman\x1b"; return c }(), root: root, action: "pull"},
		{name: "empty image", cfg: func() config.Config { c := agentTestConfig(root, false); c.RLM.Image = ""; return c }(), root: root, action: "pull"},
		{name: "option-like image", cfg: func() config.Config { c := agentTestConfig(root, false); c.RLM.Image = "--tls-verify=false"; return c }(), root: root, action: "pull"},
		{name: "spaced image", cfg: func() config.Config { c := agentTestConfig(root, false); c.RLM.Image = "bad image:tag"; return c }(), root: root, action: "ensure"},
		{name: "traversal image", cfg: func() config.Config {
			c := agentTestConfig(root, false)
			c.RLM.Image = "registry/../agent:tag"
			return c
		}(), root: root, action: "build"},
		{name: "unicode image", cfg: func() config.Config {
			c := agentTestConfig(root, false)
			c.RLM.Image = "registry/ågent:tag"
			return c
		}(), root: root, action: "pull"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeAgentImageRuntime{}
			err := runAgentImageAction(context.Background(), tt.cfg, tt.root, tt.action, io.Discard, false, fake)
			if err == nil {
				t.Fatal("want validation error")
			}
			if got := classifyError(err).category; got != categoryValidation {
				t.Fatalf("category = %q, want validation; err=%v", got, err)
			}
			if len(fake.lookups) != 0 || len(fake.files) != 0 || len(fake.calls) != 0 {
				t.Fatalf("runtime touched before validation: lookups=%v files=%v calls=%v", fake.lookups, fake.files, fake.calls)
			}
		})
	}
}

func TestAgentImageRuntimeErrorsRemainDeterministic(t *testing.T) {
	root := t.TempDir()
	cfg := agentTestConfig(root, false)

	absent := &fakeAgentImageRuntime{lookPathErr: errors.New("missing")}
	err := runAgentImageAction(context.Background(), cfg, root, "pull", io.Discard, false, absent)
	if err == nil || !strings.Contains(err.Error(), "install podman") {
		t.Fatalf("look path error = %v", err)
	}

	failed := &fakeAgentImageRuntime{runErrs: []error{errors.New("fake pull failed")}}
	err = runAgentImageAction(context.Background(), cfg, root, "pull", io.Discard, false, failed)
	if err == nil || err.Error() != "agent image: fake pull failed" {
		t.Fatalf("run error = %v", err)
	}
}

func agentTestConfig(root string, jsonOut bool) config.Config {
	return config.Config{
		Root: root,
		JSON: jsonOut,
		RLM: config.RLMConfig{
			PodmanBin: "fake-podman",
			Image:     "ghcr.io/example/agent:test",
		},
	}
}

func runAgentCLI(t *testing.T, args []string) (string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(%v) exit=%d stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
	}
	return stdout.String(), stderr.String()
}

func sampleAgentPlanText() string {
	return strings.Join([]string{
		"pm credentials add sample-local --connector sample",
		"pm credentials add warehouse-local --connector warehouse",
		"pm connections create sample_to_warehouse --source sample:sample-local --destination warehouse:warehouse-local --stream customers --primary-key id --table sample_customers",
		"pm etl run --connection sample_to_warehouse --stream customers --json",
		"",
	}, "\n")
}

func writeTestFile(path, contents string) error {
	return os.WriteFile(path, []byte(contents), 0o600)
}
