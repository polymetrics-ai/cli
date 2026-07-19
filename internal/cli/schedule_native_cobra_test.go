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
	"polymetrics.ai/internal/schedule"
)

// This compile-time reference is intentional TDD evidence: the focused test
// checkpoint must fail until schedule has a native Cobra constructor.
var _ = newScheduleCobraCommand

func TestScheduleCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	scheduleCmd := findCobraCommand(root, "schedule")
	if scheduleCmd == nil {
		t.Fatal("missing schedule command")
	}
	if scheduleCmd.DisableFlagParsing {
		t.Fatal("schedule command must use native Cobra flag parsing")
	}
	assertScheduleNoFileCompletion(t, scheduleCmd)

	flagsByAction := map[string]map[string]string{
		"create":  {"name": "stringArray", "cron": "stringArray", "flow": "stringArray"},
		"list":    {},
		"install": {"crontab": "bool"},
		"remove":  {"crontab": "bool"},
	}
	for actionName, expectedFlags := range flagsByAction {
		t.Run(actionName, func(t *testing.T) {
			action := findCobraCommand(scheduleCmd, actionName)
			if action == nil {
				t.Fatalf("missing schedule %s command", actionName)
			}
			if action.DisableFlagParsing {
				t.Fatalf("schedule %s must use native Cobra flag parsing", actionName)
			}
			if !action.FParseErrWhitelist.UnknownFlags {
				t.Fatalf("schedule %s must preserve unknown-flag tolerance", actionName)
			}
			assertScheduleNoFileCompletion(t, action)
			for flagName, flagType := range expectedFlags {
				flag := action.Flags().Lookup(flagName)
				if flag == nil {
					t.Fatalf("schedule %s missing --%s", actionName, flagName)
				}
				if got := flag.Value.Type(); got != flagType {
					t.Fatalf("schedule %s --%s type = %q, want %q", actionName, flagName, got, flagType)
				}
				if got := flag.NoOptDefVal; got != "true" {
					t.Fatalf("schedule %s --%s NoOptDefVal = %q, want true", actionName, flagName, got)
				}
			}
		})
	}

	help := findCobraCommand(scheduleCmd, "help")
	if help == nil || !help.Hidden {
		t.Fatal("schedule must preserve hidden positional help until Phase 19")
	}
	for _, name := range []string{"uninstall", "run", "history"} {
		if findCobraCommand(scheduleCmd, name) != nil {
			t.Fatalf("out-of-scope schedule action %q was registered", name)
		}
	}
	for _, spec := range cobraLegacyCommands(config.Config{}) {
		if spec.name == "schedule" {
			t.Fatal("schedule remains registered as a legacy Cobra wrapper")
		}
	}
}

func TestScheduleCreateAndListPreserveFlagsValidationAndDeterministicOutput(t *testing.T) {
	root := t.TempDir()
	fixed := time.Date(2026, 7, 19, 4, 48, 19, 0, time.UTC)
	harness := newScheduleHarness(root, fixed)

	stdout, err := harness.execute(true,
		"schedule", "create", "ignored-positional",
		"--name", "ignored", "--name=alpha",
		"--cron", "bad", "--cron=0 2 * * *",
		"--flow", "ignored", "--flow=nightly-leads",
		"--unknown", "ignored", "--=x", "---x",
	)
	if err != nil {
		t.Fatalf("create failed: %v; stdout=%s", err, stdout)
	}
	var created struct {
		Kind     string            `json:"kind"`
		OK       bool              `json:"ok"`
		Schedule schedule.Manifest `json:"schedule"`
	}
	decodeOneJSON(t, stdout, &created)
	if created.Kind != "Schedule" || !created.OK || created.Schedule.Name != "alpha" || created.Schedule.Cron != "0 2 * * *" || created.Schedule.Flow != "nightly-leads" {
		t.Fatalf("created schedule = %#v", created)
	}
	if !created.Schedule.CreatedAt.Equal(fixed) || !created.Schedule.UpdatedAt.Equal(fixed) {
		t.Fatalf("timestamps = (%s, %s), want fixed %s", created.Schedule.CreatedAt, created.Schedule.UpdatedAt, fixed)
	}
	if len(harness.installBackend.installs) != 0 || len(harness.installBackend.removes) != 0 {
		t.Fatal("create reached scheduler backend")
	}

	if _, err := harness.execute(true, "schedule", "create", "--name=zeta", "--cron=0 3 * * *", "--flow=other"); err != nil {
		t.Fatalf("create zeta: %v", err)
	}
	first, err := harness.execute(false, "schedule", "list", "ignored", "--help", "--", "-h", "--unknown", "ignored", "--=x", "---x")
	if err != nil {
		t.Fatalf("text list failed: %v", err)
	}
	second, err := harness.execute(false, "schedule", "list")
	if err != nil {
		t.Fatalf("second text list failed: %v", err)
	}
	wantText := "alpha\t0 2 * * *\tnightly-leads\nzeta\t0 3 * * *\tother\n"
	if first != wantText || second != first {
		t.Fatalf("list output is not deterministic: first=%q second=%q", first, second)
	}

	listJSON, err := harness.execute(true, "schedule", "list")
	if err != nil {
		t.Fatalf("JSON list failed: %v", err)
	}
	var listed struct {
		Kind      string              `json:"kind"`
		Schedules []schedule.Manifest `json:"schedules"`
	}
	decodeOneJSON(t, listJSON, &listed)
	if listed.Kind != "ScheduleList" || len(listed.Schedules) != 2 || listed.Schedules[0].Name != "alpha" || listed.Schedules[1].Name != "zeta" {
		t.Fatalf("ordered schedule list = %#v", listed)
	}

	for _, tt := range []struct {
		name     string
		args     []string
		category errorCategory
		message  string
	}{
		{name: "missing", args: []string{"schedule", "create"}, category: categoryUsage, message: "requires --name, --cron, --flow"},
		{name: "bare cron", args: []string{"schedule", "create", "--name=bare-cron", "--cron", "--flow=f"}, category: categoryValidation, message: "expected 5 fields"},
		{name: "invalid cron", args: []string{"schedule", "create", "--name=invalid-cron", "--cron=nope", "--flow=f"}, category: categoryValidation, message: "invalid --cron"},
		{name: "invalid name remains internal", args: []string{"schedule", "create", "--name=INVALID", "--cron=0 2 * * *", "--flow=f"}, category: categoryInternal, message: "invalid schedule name"},
		{name: "conflict", args: []string{"schedule", "create", "--name=alpha", "--cron=0 2 * * *", "--flow=f"}, category: categoryValidation, message: "already exists"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stdout, err := harness.execute(true, tt.args...)
			if err == nil {
				t.Fatalf("expected error; stdout=%s", stdout)
			}
			classified := classifyError(mapCobraErr(err))
			if classified.category != tt.category || !strings.Contains(classified.message, tt.message) {
				t.Fatalf("error = (%s, %q), want (%s, contains %q)", classified.category, classified.message, tt.category, tt.message)
			}
		})
	}
}

func TestScheduleInstallRemoveUseInjectedBackendsAndPreserveOperands(t *testing.T) {
	root := t.TempDir()
	fixed := time.Date(2026, 7, 19, 4, 48, 19, 0, time.UTC)
	harness := newScheduleHarness(root, fixed)
	ctxKey := scheduleContextKey{}
	harness.ctx = context.WithValue(context.Background(), ctxKey, "kept")
	if _, err := harness.execute(false, "schedule", "create", "--name=alpha", "--cron=0 2 * * *", "--flow=nightly-leads"); err != nil {
		t.Fatalf("seed create: %v", err)
	}

	harness.selectKind = schedule.KindCrontab
	stdout, err := harness.execute(false,
		"schedule", "install", "alpha", "ignored-later",
		"--crontab=false", "--crontab", "--unknown=ignored", "--help", "--", "-h",
	)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if stdout != "Installed schedule alpha via crontab\n" {
		t.Fatalf("install text = %q", stdout)
	}
	if len(harness.selections) != 1 || !harness.selections[0].forceCrontab || harness.selections[0].cfg.CrontabFile != harness.crontabFile {
		t.Fatalf("backend selection = %#v", harness.selections)
	}
	if len(harness.installBackend.installs) != 1 {
		t.Fatalf("install calls = %d, want 1", len(harness.installBackend.installs))
	}
	install := harness.installBackend.installs[0]
	if install.manifest.Name != "alpha" || install.manifest.Root != root || install.pmBin != "/tmp/fake-pm" || install.contextValue != "kept" {
		t.Fatalf("install call = %#v", install)
	}

	for _, tt := range []struct {
		name       string
		args       []string
		wantForce  bool
		wantUsage  bool
		wantNeedle string
	}{
		{name: "space false", args: []string{"schedule", "install", "alpha", "--crontab", "false"}, wantForce: false},
		{name: "assigned arbitrary false", args: []string{"schedule", "install", "alpha", "--crontab=anything"}, wantForce: false},
		{name: "last assigned wins", args: []string{"schedule", "install", "alpha", "--crontab=true", "--crontab=false"}, wantForce: false},
		{name: "flag consumes would-be operand", args: []string{"schedule", "install", "--crontab", "alpha"}, wantUsage: true, wantNeedle: "schedule install <name>"},
		{name: "unknown consumes would-be operand", args: []string{"schedule", "install", "--unknown", "alpha"}, wantUsage: true, wantNeedle: "schedule install <name>"},
		{name: "literal consumes would-be operand", args: []string{"schedule", "install", "--", "alpha"}, wantUsage: true, wantNeedle: "schedule install <name>"},
		{name: "short first operand", args: []string{"schedule", "install", "-h", "alpha"}, wantNeedle: `schedule "-h" not found`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			beforeSelections := len(harness.selections)
			beforeInstalls := len(harness.installBackend.installs)
			stdout, err := harness.execute(true, tt.args...)
			if tt.wantUsage || tt.wantNeedle != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantNeedle) {
					t.Fatalf("error=%v stdout=%s, want contains %q", err, stdout, tt.wantNeedle)
				}
				if len(harness.installBackend.installs) != beforeInstalls {
					t.Fatal("invalid install reached backend")
				}
				return
			}
			if err != nil {
				t.Fatalf("install form failed: %v", err)
			}
			if len(harness.selections) != beforeSelections+1 || harness.selections[len(harness.selections)-1].forceCrontab != tt.wantForce {
				t.Fatalf("force selection mismatch: %#v", harness.selections)
			}
		})
	}

	harness.selectKind = schedule.KindLaunchd
	removeJSON, err := harness.execute(true, "schedule", "remove", "alpha", "ignored-later", "--crontab=false", "--unknown", "ignored")
	if err != nil {
		t.Fatalf("remove failed: %v", err)
	}
	var removed struct {
		Kind string `json:"kind"`
		OK   bool   `json:"ok"`
		Name string `json:"name"`
	}
	decodeOneJSON(t, removeJSON, &removed)
	if removed.Kind != "ScheduleRemove" || !removed.OK || removed.Name != "alpha" {
		t.Fatalf("remove envelope = %#v", removed)
	}
	if len(harness.installBackend.removes) != 1 || harness.installBackend.removes[0].name != "alpha" || harness.installBackend.removes[0].contextValue != "kept" {
		t.Fatalf("selected backend removals = %#v", harness.installBackend.removes)
	}
	if len(harness.crontabBackend.removes) != 1 || harness.crontabBackend.removes[0].name != "alpha" {
		t.Fatalf("crontab fallback removals = %#v", harness.crontabBackend.removes)
	}
	if _, err := schedule.Load(root, "alpha"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("removed manifest still loads: %v", err)
	}
}

func TestScheduleHelpInvalidActionsGlobalsAndNoBackendDiscovery(t *testing.T) {
	var canonical string
	for _, tt := range []struct {
		name string
		args []string
	}{
		{name: "help topic", args: []string{"help", "schedule"}},
		{name: "bare", args: []string{"schedule"}},
		{name: "long", args: []string{"schedule", "--help"}},
		{name: "short", args: []string{"schedule", "-h"}},
		{name: "positional", args: []string{"schedule", "help"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, code := runNativeScheduleCLI(tt.args...)
			if code != 0 || stderr != "" || !strings.Contains(stdout, "pm schedule - create, list, install, and remove flow schedules") {
				t.Fatalf("help mismatch: code=%d stderr=%s stdout=%s", code, stderr, stdout)
			}
			if canonical == "" {
				canonical = stdout
			} else if stdout != canonical {
				t.Fatalf("%s help differs from canonical", tt.name)
			}
		})
	}

	stdout, stderr, code := runNativeScheduleCLI("schedule", "--json")
	if code != 0 || stderr != "" {
		t.Fatalf("JSON manual failed: code=%d stderr=%s", code, stderr)
	}
	var manual struct {
		Kind    string `json:"kind"`
		Command string `json:"command"`
		Manual  string `json:"manual"`
	}
	decodeOneJSON(t, stdout, &manual)
	if manual.Kind != "CommandManual" || manual.Command != "schedule" || manual.Manual != canonical {
		t.Fatalf("JSON manual mismatch: %#v", manual)
	}

	root := t.TempDir()
	harness := newScheduleHarness(root, time.Date(2026, 7, 19, 4, 48, 19, 0, time.UTC))
	for _, args := range [][]string{
		{"schedule", "uninstall", "alpha"},
		{"schedule", "run", "alpha"},
		{"schedule", "history", "alpha"},
		{"schedule", "bogus", "create", "--name=alpha", "--cron=0 2 * * *", "--flow=f"},
		{"schedule", "bogus", "--help", "create"},
		{"schedule", "--unknown", "create"},
		{"schedule", "--", "create"},
		{"schedule", "--=x", "create"},
		{"schedule", "---x", "create"},
	} {
		stdout, err := harness.execute(true, args...)
		if err == nil || exitCodeFor(classifyError(mapCobraErr(err))) != 2 {
			t.Fatalf("invalid action %v: err=%v stdout=%s", args, err, stdout)
		}
		if len(harness.selections) != 0 || len(harness.installBackend.installs) != 0 || len(harness.installBackend.removes) != 0 || len(harness.crontabBackend.removes) != 0 {
			t.Fatalf("invalid action reached backend: args=%v", args)
		}
	}

	stdout, stderr, code = runNativeScheduleCLI("--json", "--json=maybe", "schedule")
	assertCLIError(t, code, stdout, stderr, 3, "validation", "invalid --json")
	stdout, stderr, code = runNativeScheduleCLI("--json=false", "--plain=true", "--no-input=on", "schedule")
	if code != 0 || stderr != "" || !strings.HasPrefix(stdout, "NAME\n  pm schedule") {
		t.Fatalf("assigned global booleans mismatch: code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
}

func TestScheduleBackendFailureAndNotFoundPreserveErrorsWithoutExternalEffects(t *testing.T) {
	root := t.TempDir()
	harness := newScheduleHarness(root, time.Date(2026, 7, 19, 4, 48, 19, 0, time.UTC))

	stdout, err := harness.execute(true, "schedule", "install", "missing", "--crontab")
	if err == nil || classifyError(mapCobraErr(err)).category != categoryValidation || !strings.Contains(err.Error(), `schedule "missing" not found`) {
		t.Fatalf("missing install: err=%v stdout=%s", err, stdout)
	}
	if len(harness.selections) != 0 || len(harness.installBackend.installs) != 0 {
		t.Fatal("missing schedule reached backend selection")
	}

	if _, err := harness.execute(false, "schedule", "create", "--name=alpha", "--cron=0 2 * * *", "--flow=f"); err != nil {
		t.Fatalf("seed create: %v", err)
	}
	harness.installBackend.installErr = errors.New("fake scheduler rejected install")
	stdout, err = harness.execute(true, "schedule", "install", "alpha", "--crontab")
	if err == nil || classifyError(mapCobraErr(err)).category != categoryInternal || !strings.Contains(err.Error(), "install failed: fake scheduler rejected install") {
		t.Fatalf("install failure: err=%v stdout=%s", err, stdout)
	}
}

func assertScheduleNoFileCompletion(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	if cmd.ValidArgsFunction == nil {
		t.Fatalf("%s missing completion seam", cmd.CommandPath())
	}
	values, directive := cmd.ValidArgsFunction(cmd, nil, "")
	if len(values) != 0 || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("%s completion = (%v, %v), want no values and NoFileComp", cmd.CommandPath(), values, directive)
	}
}

func runNativeScheduleCLI(args ...string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

type scheduleContextKey struct{}

type scheduleBackendInstall struct {
	manifest     schedule.Manifest
	pmBin        string
	contextValue string
}

type scheduleBackendRemove struct {
	name         string
	contextValue string
}

type fakeScheduleBackend struct {
	kind       schedule.BackendKind
	installs   []scheduleBackendInstall
	removes    []scheduleBackendRemove
	installErr error
	removeErr  error
}

func (b *fakeScheduleBackend) Install(ctx context.Context, manifest schedule.Manifest, pmBin string) error {
	value, _ := ctx.Value(scheduleContextKey{}).(string)
	b.installs = append(b.installs, scheduleBackendInstall{manifest: manifest, pmBin: pmBin, contextValue: value})
	return b.installErr
}

func (b *fakeScheduleBackend) Remove(ctx context.Context, name string) error {
	value, _ := ctx.Value(scheduleContextKey{}).(string)
	b.removes = append(b.removes, scheduleBackendRemove{name: name, contextValue: value})
	return b.removeErr
}

func (b *fakeScheduleBackend) Kind() schedule.BackendKind { return b.kind }

type scheduleSelection struct {
	forceCrontab bool
	cfg          schedule.BackendConfig
}

type scheduleHarness struct {
	root           string
	fixed          time.Time
	ctx            context.Context
	crontabFile    string
	selectKind     schedule.BackendKind
	selections     []scheduleSelection
	installBackend *fakeScheduleBackend
	crontabBackend *fakeScheduleBackend
}

func newScheduleHarness(root string, fixed time.Time) *scheduleHarness {
	return &scheduleHarness{
		root:           root,
		fixed:          fixed,
		ctx:            context.Background(),
		crontabFile:    filepath.Join(root, "fake-crontab"),
		selectKind:     schedule.KindCrontab,
		installBackend: &fakeScheduleBackend{kind: schedule.KindCrontab},
		crontabBackend: &fakeScheduleBackend{kind: schedule.KindCrontab},
	}
}

func (h *scheduleHarness) execute(jsonOut bool, args ...string) (string, error) {
	h.installBackend.kind = h.selectKind
	cfg := testRouterConfig(h.root, jsonOut)
	cfg.Schedule.CrontabFile = h.crontabFile
	var stdout bytes.Buffer
	runtime := scheduleCommandRuntime{
		now:        func() time.Time { return h.fixed },
		executable: func() (string, error) { return "/tmp/fake-pm", nil },
		selectBackend: func(_ context.Context, forceCrontab bool, cfg schedule.BackendConfig) schedule.Backend {
			h.selections = append(h.selections, scheduleSelection{forceCrontab: forceCrontab, cfg: cfg})
			return h.installBackend
		},
		crontabBackend: func(_ string) schedule.Backend { return h.crontabBackend },
	}
	cmd := newRootCmdWithScheduleRuntime(h.ctx, cfg, &stdout, io.Discard, runtime)
	err := executeRootCmd(cmd, args)
	return stdout.String(), err
}

func decodeScheduleEnvelope(t *testing.T, stdout string) map[string]any {
	t.Helper()
	var env map[string]any
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("decode schedule envelope: %v; stdout=%s", err, stdout)
	}
	return env
}
