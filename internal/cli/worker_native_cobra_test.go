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

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/worker"
)

// This compile-time reference is intentional TDD evidence: the focused test
// checkpoint must fail until worker has a native Cobra constructor.
var _ = newWorkerCobraCommand

func TestWorkerCommandIsHiddenNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), workerTestConfig(".", false, ""), io.Discard, io.Discard)
	workerCmd := findCobraCommand(root, "worker")
	if workerCmd == nil {
		t.Fatal("missing worker command")
	}
	if !workerCmd.Hidden {
		t.Fatal("worker command must remain hidden")
	}
	if workerCmd.DisableFlagParsing {
		t.Fatal("worker command must use native Cobra flag parsing")
	}
	assertWorkerNoFileCompletion(t, workerCmd)
	if workerCmd.LocalNonPersistentFlags().HasFlags() {
		t.Fatal("worker unexpectedly declares local flags")
	}

	for _, actionName := range []string{"status", "serve"} {
		action := findCobraCommand(workerCmd, actionName)
		if action == nil {
			t.Fatalf("missing worker %s command", actionName)
		}
		if action.DisableFlagParsing {
			t.Fatalf("worker %s must use native Cobra flag parsing", actionName)
		}
		if !action.FParseErrWhitelist.UnknownFlags {
			t.Fatalf("worker %s must preserve unknown-flag tolerance", actionName)
		}
		if action.LocalNonPersistentFlags().HasFlags() {
			t.Fatalf("worker %s unexpectedly declares local flags", actionName)
		}
		assertWorkerNoFileCompletion(t, action)
	}
	help := findCobraCommand(workerCmd, "help")
	if help == nil || !help.Hidden {
		t.Fatal("worker must provide hidden positional help")
	}
	for _, name := range []string{"exec", "run", "shell", "workflow", "container", "listener", "dashboard"} {
		if findCobraCommand(workerCmd, name) != nil {
			t.Fatalf("generic or out-of-scope worker action %q was registered", name)
		}
	}
	for _, spec := range cobraLegacyCommands(config.Config{}) {
		if spec.name == "worker" {
			t.Fatal("worker remains registered as a legacy Cobra wrapper")
		}
	}
}

func TestWorkerStatusAndServeUseInjectedRuntime(t *testing.T) {
	ctxKey := workerNativeContextKey{}
	ctx := context.WithValue(context.Background(), ctxKey, "kept")
	runtime := newFakeWorkerRuntime()
	runtime.reachable = true
	cfg := workerTestConfig(t.TempDir(), true, "temporal.example:7233")
	cfg.RLM.PodmanBin = "fake-podman-bin"
	cfg.RLM.Image = "example.invalid/typed-rlm-worker:test"

	stdout, stderr, err := executeNativeWorker(ctx, cfg, runtime, "worker", "status", "ignored", "--unknown", "ignored", "--", "-h")
	if err != nil || stderr != "" {
		t.Fatalf("status: err=%v stderr=%q", err, stderr)
	}
	var status map[string]any
	decodeOneJSON(t, stdout, &status)
	if status["kind"] != "WorkerStatus" || status["status"] != "ok" || status["reachable"] != true || status["addr"] != "temporal.example:7233" || status["task_queue"] != worker.TaskQueue {
		t.Fatalf("status envelope = %#v", status)
	}
	if runtime.statusCalls != 1 || runtime.statusAddr != "temporal.example:7233" {
		t.Fatalf("status calls=%d addr=%q", runtime.statusCalls, runtime.statusAddr)
	}
	if runtime.statusContextValue != "kept" || !runtime.statusHadDeadline {
		t.Fatalf("status context value=%v deadline=%v", runtime.statusContextValue, runtime.statusHadDeadline)
	}

	stdout, stderr, err = executeNativeWorker(ctx, cfg, runtime, "worker", "serve", "ignored", "--unknown=ignored", "--", "operand")
	if err != nil || stderr != "" {
		t.Fatalf("serve: err=%v stderr=%q", err, stderr)
	}
	var served map[string]any
	decodeOneJSON(t, stdout, &served)
	if served["kind"] != "WorkerServe" || served["status"] != "ready" || served["addr"] != "temporal.example:7233" || served["task_queue"] != worker.TaskQueue {
		t.Fatalf("serve envelope = %#v", served)
	}
	if runtime.serveCalls != 1 || runtime.serveAddr != "temporal.example:7233" || runtime.readyCalls != 1 {
		t.Fatalf("serve calls=%d addr=%q ready=%d", runtime.serveCalls, runtime.serveAddr, runtime.readyCalls)
	}
	if runtime.serveContextValue != "kept" {
		t.Fatalf("serve context value=%v", runtime.serveContextValue)
	}
	if runtime.activities == nil || runtime.activities.PodmanBin != "fake-podman-bin" || runtime.activities.Image != "example.invalid/typed-rlm-worker:test" {
		t.Fatalf("serve activities = %#v", runtime.activities)
	}
}

func TestWorkerHelpRoutesAreContextualHiddenAndSideEffectFree(t *testing.T) {
	manual := docs["worker"]
	if manual == "" || !strings.Contains(manual, "typed RLM Temporal workflow") || !strings.Contains(manual, "not a generic") {
		t.Fatal("worker manual must document the typed non-generic boundary")
	}

	tests := []struct {
		name string
		args []string
		json bool
	}{
		{name: "topic", args: []string{"help", "worker"}},
		{name: "bare", args: []string{"worker"}},
		{name: "long", args: []string{"worker", "--help"}},
		{name: "short", args: []string{"worker", "-h"}},
		{name: "positional", args: []string{"worker", "help"}},
		{name: "status long", args: []string{"worker", "status", "--help"}},
		{name: "status positional", args: []string{"worker", "status", "help"}},
		{name: "serve short", args: []string{"worker", "serve", "-h"}},
		{name: "serve positional", args: []string{"worker", "serve", "help"}},
		{name: "JSON bare", args: []string{"worker"}, json: true},
		{name: "JSON trailing", args: []string{"worker", "status", "--help"}, json: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime := newFakeWorkerRuntime()
			cfg := workerTestConfig(t.TempDir(), tt.json, "temporal.invalid:7233")
			stdout, stderr, err := executeNativeWorker(context.Background(), cfg, runtime, tt.args...)
			if err != nil || stderr != "" {
				t.Fatalf("help: err=%v stderr=%q stdout=%q", err, stderr, stdout)
			}
			if tt.json {
				var envelope map[string]any
				decodeOneJSON(t, stdout, &envelope)
				if envelope["kind"] != "CommandManual" || envelope["command"] != "worker" || envelope["manual"] != manual {
					t.Fatalf("JSON manual = %#v", envelope)
				}
			} else if stdout != manual {
				t.Fatalf("manual differs from canonical worker text")
			}
			assertWorkerRuntimeNotCalled(t, runtime)
		})
	}

	root := newRootCmd(context.Background(), workerTestConfig(".", false, ""), io.Discard, io.Discard)
	for _, cmd := range root.Commands() {
		if cmd.Name() == "worker" && !cmd.Hidden {
			t.Fatal("worker leaked into root discovery")
		}
	}
}

func TestWorkerInvalidActionsLiteralUnknownAndNoActionDiscovery(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "unknown action", args: []string{"worker", "bogus", "status"}},
		{name: "generic run", args: []string{"worker", "run", "serve"}},
		{name: "unknown flag head", args: []string{"worker", "--unknown", "status"}},
		{name: "assigned unknown head", args: []string{"worker", "--unknown=value", "serve"}},
		{name: "malformed empty flag", args: []string{"worker", "--=x", "status"}},
		{name: "malformed triple dash", args: []string{"worker", "---x", "serve"}},
		{name: "short flag head", args: []string{"worker", "-x", "status"}},
		{name: "literal head", args: []string{"worker", "--", "serve"}},
		{name: "operand head", args: []string{"worker", "operand", "status"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime := newFakeWorkerRuntime()
			cfg := workerTestConfig(t.TempDir(), true, "temporal.invalid:7233")
			stdout, _, err := executeNativeWorker(context.Background(), cfg, runtime, tt.args...)
			if err == nil {
				t.Fatalf("invalid action succeeded; stdout=%q", stdout)
			}
			if got := exitCodeFor(classifyError(mapCobraErr(err))); got != 2 {
				t.Fatalf("exit code = %d, want usage 2; error=%v", got, err)
			}
			assertWorkerRuntimeNotCalled(t, runtime)
			if strings.Contains(stdout, "WorkerStatus") || strings.Contains(stdout, "WorkerServe") {
				t.Fatal("invalid action emitted a worker operation envelope")
			}
		})
	}
}

func TestWorkerConfigPrecedenceGlobalsAndNondisclosure(t *testing.T) {
	const (
		fileAddr     = "file-temporal.example:7233"
		aliasAddr    = "alias-temporal.example:7233"
		primaryAddr  = "primary-temporal.example:7233"
		imageCanary  = "example.invalid/config-image-canary:test"
		podmanCanary = "config-podman-canary"
	)
	root := writeWorkerConfig(t, "runtime:\n  temporal_addr: "+fileAddr+"\nrlm:\n  image: "+imageCanary+"\n  podman_bin: "+podmanCanary+"\n")

	tests := []struct {
		name       string
		primary    string
		alias      string
		args       []string
		wantAddr   string
		wantJSON   bool
		wantSource string
	}{
		{name: "file", args: []string{"--root", root, "worker", "status"}, wantAddr: fileAddr, wantSource: "config"},
		{name: "legacy alias", alias: aliasAddr, args: []string{"worker", "status", "--root=" + root, "--json"}, wantAddr: aliasAddr, wantJSON: true, wantSource: "env"},
		{name: "primary over alias", primary: primaryAddr, alias: aliasAddr, args: []string{"--json=false", "--plain=true", "worker", "status", "--json", "--root", root}, wantAddr: primaryAddr, wantJSON: true, wantSource: "env"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("POLYMETRICS_TEMPORAL_ADDR", tt.primary)
			t.Setenv("PM_TEMPORAL_ADDR", tt.alias)
			runtime := newFakeWorkerRuntime()
			stdout, stderr, code, cfg := runWorkerInvocation(t, runtime, tt.args...)
			if code != 0 || stderr != "" {
				t.Fatalf("code=%d stderr=%q", code, stderr)
			}
			if runtime.statusAddr != tt.wantAddr || cfg.Source("runtime.temporal_addr") != tt.wantSource {
				t.Fatalf("addr=%q source=%q, want %q/%q", runtime.statusAddr, cfg.Source("runtime.temporal_addr"), tt.wantAddr, tt.wantSource)
			}
			if tt.wantJSON {
				var status map[string]any
				decodeOneJSON(t, stdout, &status)
				if status["addr"] != tt.wantAddr {
					t.Fatalf("JSON addr=%v, want %q", status["addr"], tt.wantAddr)
				}
			} else if !strings.Contains(stdout, "temporal="+tt.wantAddr) {
				t.Fatalf("text status missing selected address: %q", stdout)
			}
			if strings.Contains(stdout+stderr, imageCanary) || strings.Contains(stdout+stderr, podmanCanary) {
				t.Fatal("unrelated worker configuration leaked")
			}
			if tt.primary != "" && tt.alias != "" && strings.Contains(stdout+stderr, tt.alias) {
				t.Fatal("lower-precedence worker configuration leaked")
			}
		})
	}

	t.Setenv("POLYMETRICS_TEMPORAL_ADDR", "")
	t.Setenv("PM_TEMPORAL_ADDR", "")
	runtime := newFakeWorkerRuntime()
	stdout, stderr, code, _ := runWorkerInvocation(t, runtime, "--root", t.TempDir(), "--json", "worker", "serve")
	if code != 3 || !strings.Contains(stdout, `"code": "validation_error"`) || !strings.Contains(stderr, "POLYMETRICS_TEMPORAL_ADDR is not set") {
		t.Fatalf("missing explicit config: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	assertWorkerRuntimeNotCalled(t, runtime)

	const malformedCanary = "malformed-config-canary"
	malformedRoot := writeWorkerConfig(t, "runtime:\n  temporal_addr: ["+malformedCanary+"\n")
	runtime = newFakeWorkerRuntime()
	stdout, stderr, code, _ = runWorkerInvocation(t, runtime, "--root", malformedRoot, "--json", "worker", "status")
	if code != 3 {
		t.Fatalf("malformed config code=%d, want 3", code)
	}
	if strings.Contains(stdout+stderr, malformedCanary) {
		t.Fatal("malformed configuration value leaked")
	}
	assertWorkerRuntimeNotCalled(t, runtime)

	runtime = newFakeWorkerRuntime()
	_, _, code, _ = runWorkerInvocation(t, runtime, "--json=maybe", "worker", "status")
	if code != 3 {
		t.Fatalf("invalid global code=%d, want 3", code)
	}
	assertWorkerRuntimeNotCalled(t, runtime)
}

func TestWorkerContextCancellationPropagatesToInjectedRuntime(t *testing.T) {
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), workerNativeContextKey{}, "kept"))
	cancel()
	cfg := workerTestConfig(t.TempDir(), false, "temporal.invalid:7233")

	statusRuntime := newFakeWorkerRuntime()
	_, _, err := executeNativeWorker(ctx, cfg, statusRuntime, "worker", "status")
	if err != nil {
		t.Fatalf("status cancellation: %v", err)
	}
	if statusRuntime.statusCalls != 1 || !errors.Is(statusRuntime.statusContextErr, context.Canceled) {
		t.Fatalf("status calls=%d context error=%v", statusRuntime.statusCalls, statusRuntime.statusContextErr)
	}

	serveRuntime := newFakeWorkerRuntime()
	serveRuntime.serveErr = context.Canceled
	_, _, err = executeNativeWorker(ctx, cfg, serveRuntime, "worker", "serve")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("serve error=%v, want context canceled", err)
	}
	if serveRuntime.serveCalls != 1 || !errors.Is(serveRuntime.serveContextErr, context.Canceled) {
		t.Fatalf("serve calls=%d context error=%v", serveRuntime.serveCalls, serveRuntime.serveContextErr)
	}
}

type workerNativeContextKey struct{}

type fakeWorkerRuntime struct {
	reachable          bool
	serveErr           error
	statusCalls        int
	serveCalls         int
	readyCalls         int
	statusAddr         string
	serveAddr          string
	statusContextValue any
	serveContextValue  any
	statusContextErr   error
	serveContextErr    error
	statusHadDeadline  bool
	activities         *worker.PodmanActivities
}

func newFakeWorkerRuntime() *fakeWorkerRuntime {
	return &fakeWorkerRuntime{}
}

func (f *fakeWorkerRuntime) runtime() workerCommandRuntime {
	return workerCommandRuntime{
		status: func(ctx context.Context, addr string) bool {
			f.statusCalls++
			f.statusAddr = addr
			f.statusContextValue = ctx.Value(workerNativeContextKey{})
			f.statusContextErr = ctx.Err()
			_, f.statusHadDeadline = ctx.Deadline()
			return f.reachable
		},
		serve: func(ctx context.Context, addr string, activities *worker.PodmanActivities, ready func()) error {
			f.serveCalls++
			f.serveAddr = addr
			f.serveContextValue = ctx.Value(workerNativeContextKey{})
			f.serveContextErr = ctx.Err()
			f.activities = activities
			if f.serveErr == nil {
				ready()
				f.readyCalls++
			}
			return f.serveErr
		},
	}
}

func workerTestConfig(root string, jsonOut bool, temporalAddr string) config.Config {
	cfg := testRouterConfig(root, jsonOut)
	cfg.Runtime.TemporalAddr = temporalAddr
	if temporalAddr != "" {
		cfg.ExplicitKeys = map[string]bool{"runtime.temporal_addr": true}
		cfg.ValueSources = map[string]string{"runtime.temporal_addr": "file"}
	}
	return cfg
}

func executeNativeWorker(ctx context.Context, cfg config.Config, runtime *fakeWorkerRuntime, args ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := newRootCmdWithWorkerRuntime(ctx, cfg, &stdout, &stderr, runtime.runtime())
	err := executeRootCmd(cmd, args)
	return stdout.String(), stderr.String(), err
}

func runWorkerInvocation(t *testing.T, runtime *fakeWorkerRuntime, args ...string) (string, string, int, config.Config) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	globals, err := parseGlobal(args)
	if err != nil {
		code := writeError(context.Background(), &stdout, &stderr, err, globals.jsonOut)
		return stdout.String(), stderr.String(), code, config.Config{}
	}
	opts := config.Options{Root: globals.root, Flags: globalConfigFlags(args, globals.root, globals.jsonOut)}
	bootstrap, err := config.ResolveBootstrap(opts)
	if err != nil {
		code := writeError(context.Background(), &stdout, &stderr, validationErrorf("%v", err), bootstrap.JSON)
		return stdout.String(), stderr.String(), code, config.Config{}
	}
	cfg, err := config.Load(opts)
	if err != nil {
		code := writeError(context.Background(), &stdout, &stderr, validationErrorf("%v", err), bootstrap.JSON)
		return stdout.String(), stderr.String(), code, config.Config{}
	}
	cmd := newRootCmdWithWorkerRuntime(context.Background(), cfg, &stdout, &stderr, runtime.runtime())
	if err := executeRootCmd(cmd, globals.clean); err != nil {
		code := writeError(context.Background(), &stdout, &stderr, mapCobraErr(err), cfg.JSON)
		return stdout.String(), stderr.String(), code, cfg
	}
	return stdout.String(), stderr.String(), 0, cfg
}

func writeWorkerConfig(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return root
}

func assertWorkerNoFileCompletion(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	if cmd.ValidArgsFunction == nil {
		t.Fatalf("%s missing completion seam", cmd.CommandPath())
	}
	values, directive := cmd.ValidArgsFunction(cmd, nil, "")
	if len(values) != 0 || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("%s completion values=%v directive=%v", cmd.CommandPath(), values, directive)
	}
}

func assertWorkerRuntimeNotCalled(t *testing.T, runtime *fakeWorkerRuntime) {
	t.Helper()
	if runtime.statusCalls != 0 || runtime.serveCalls != 0 || runtime.readyCalls != 0 {
		t.Fatalf("worker runtime called: status=%d serve=%d ready=%d", runtime.statusCalls, runtime.serveCalls, runtime.readyCalls)
	}
}

func decodeWorkerEnvelope(t *testing.T, stdout string) map[string]any {
	t.Helper()
	var env map[string]any
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("decode worker envelope: %v", err)
	}
	return env
}
