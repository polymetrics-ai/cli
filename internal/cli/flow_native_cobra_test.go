package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/config"
)

// This compile-time reference is intentional TDD evidence: the focused test
// checkpoint must fail until flow has a native Cobra constructor.
var _ = newFlowCobraCommand

func TestFlowCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	flowCmd := findCobraCommand(root, "flow")
	if flowCmd == nil {
		t.Fatal("missing flow command")
	}
	if flowCmd.DisableFlagParsing {
		t.Fatal("flow command must use native Cobra flag parsing")
	}
	assertFlowNoFileCompletion(t, flowCmd)

	flagsByAction := map[string]map[string]string{
		"plan":    {"file": "stringArray"},
		"preview": {"file": "stringArray"},
		"run":     {"file": "stringArray", "flows-dir": "stringArray", "force": "bool"},
		"status":  {"flows-dir": "stringArray"},
		"list":    {"flows-dir": "stringArray"},
	}
	for actionName, expectedFlags := range flagsByAction {
		t.Run(actionName, func(t *testing.T) {
			action := findCobraCommand(flowCmd, actionName)
			if action == nil {
				t.Fatalf("missing flow %s command", actionName)
			}
			if action.DisableFlagParsing {
				t.Fatalf("flow %s must use native Cobra flag parsing", actionName)
			}
			if !action.FParseErrWhitelist.UnknownFlags {
				t.Fatalf("flow %s must preserve unknown-flag tolerance", actionName)
			}
			assertFlowNoFileCompletion(t, action)
			for flagName, flagType := range expectedFlags {
				flag := action.Flags().Lookup(flagName)
				if flag == nil {
					t.Fatalf("flow %s missing --%s", actionName, flagName)
				}
				if got := flag.Value.Type(); got != flagType {
					t.Fatalf("flow %s --%s type = %q, want %q", actionName, flagName, got, flagType)
				}
				if got := flag.NoOptDefVal; got != "true" {
					t.Fatalf("flow %s --%s NoOptDefVal = %q, want true", actionName, flagName, got)
				}
			}
		})
	}

	help := findCobraCommand(flowCmd, "help")
	if help == nil || !help.Hidden {
		t.Fatal("flow must preserve hidden positional help until Phase 19")
	}
	for _, spec := range cobraLegacyCommands(config.Config{}) {
		if spec.name == "flow" {
			t.Fatal("flow remains registered as a legacy Cobra wrapper")
		}
	}
}

func TestFlowPlanPreviewRunListStatusPreserveFlagsOperandsAndDirectories(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)
	flowDir := filepath.Join(root, ".polymetrics", "flows")
	if err := os.MkdirAll(flowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeNativeFlowFixture(t, flowDir, "alpha")
	writeNativeFlowFixture(t, flowDir, "ignored-first")
	writeNativeFlowFixture(t, flowDir, "zeta")

	alphaPath := filepath.Join(flowDir, "alpha.json")
	stdout, stderr, code := runNativeFlowCLI("flow", "plan", "ignored", "--file", filepath.Join(flowDir, "missing.json"), "--file", alphaPath, "--unknown", "ignored", "--root", root, "--json")
	if code != 0 || stderr != "" {
		t.Fatalf("plan failed: code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	var plan struct {
		Status string   `json:"status"`
		Flow   string   `json:"flow"`
		Order  []string `json:"order"`
	}
	decodeOneJSON(t, stdout, &plan)
	if plan.Status != "ok" || plan.Flow != "alpha" || len(plan.Order) != 1 || plan.Order[0] != "score" {
		t.Fatalf("plan output = %#v", plan)
	}

	stdout, stderr, code = runNativeFlowCLI("flow", "preview", "--file", alphaPath, "--help", "--", "-h", "--unknown=ignored", "--root", root, "--json")
	if code != 0 || stderr != "" {
		t.Fatalf("preview with ignored trailing controls failed: code=%d stderr=%s", code, stderr)
	}
	var preview struct {
		Status string `json:"status"`
	}
	decodeOneJSON(t, stdout, &preview)
	if preview.Status != "dry_run" {
		t.Fatalf("preview status = %q, want dry_run", preview.Status)
	}

	stdout, stderr, code = runNativeFlowCLI("flow", "run", "ignored-first", "alpha", "--flows-dir", filepath.Join(root, "missing"), "--flows-dir", flowDir, "--force", "--force", "--unknown=ignored", "--root", root, "--json", "--progress", "ndjson")
	if code != 0 {
		t.Fatalf("named run failed: code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	var run struct {
		FlowName string `json:"flow_name"`
		Status   string `json:"status"`
	}
	decodeOneJSON(t, stdout, &run)
	if run.FlowName != "ignored-first" || run.Status != "ok" {
		t.Fatalf("run output = %#v, want first positional flow", run)
	}
	assertFlowNDJSONSequence(t, stderr, []string{"started:running", "started:running", "completed:success", "completed:success"})

	stateDir := filepath.Join(root, ".polymetrics")
	checkpointData, err := os.ReadFile(filepath.Join(stateDir, "flow-checkpoints.json"))
	if err != nil {
		t.Fatalf("read flow checkpoint: %v", err)
	}
	if !bytes.Contains(checkpointData, []byte(`"ignored-first/score":"success"`)) {
		t.Fatalf("checkpoint did not preserve named flow status: %s", checkpointData)
	}
	statusManifest := strings.ReplaceAll(nativeFlowManifest("ignored-first"), `"name": "ignored-first"`, `"name": "ignored-first"`)
	if err := os.WriteFile(filepath.Join(stateDir, "ignored-first.json"), []byte(statusManifest), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code = runNativeFlowCLI("flow", "status", "ignored-first", "later-name", "--flows-dir", t.TempDir(), "--flows-dir", stateDir, "--unknown", "ignored", "--root", root, "--json")
	if code != 0 || stderr != "" {
		t.Fatalf("status failed: code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	if !strings.Contains(stdout, `"status": "success"`) || strings.Contains(stdout, "later-name") {
		t.Fatalf("status did not preserve first operand/checkpoint: %s", stdout)
	}

	stdout, stderr, code = runNativeFlowCLI("flow", "list", "ignored", "--flows-dir", t.TempDir(), "--flows-dir", flowDir, "--=x", "---x", "--unknown", "ignored", "--root", root, "--json")
	if code != 0 || stderr != "" {
		t.Fatalf("list failed: code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	var list struct {
		Flows []string `json:"flows"`
	}
	decodeOneJSON(t, stdout, &list)
	if got, want := strings.Join(list.Flows, ","), "alpha,fixture-spec,ignored-first,zeta"; got != want {
		t.Fatalf("deterministic flow list = %q, want %q", got, want)
	}
}

func TestFlowHelpMalformedUnknownActionsAndGlobalBooleans(t *testing.T) {
	var canonical string
	for _, tt := range []struct {
		name string
		args []string
	}{
		{name: "help topic", args: []string{"help", "flow"}},
		{name: "bare", args: []string{"flow"}},
		{name: "long", args: []string{"flow", "--help"}},
		{name: "short", args: []string{"flow", "-h"}},
		{name: "positional", args: []string{"flow", "help"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, code := runNativeFlowCLI(tt.args...)
			if code != 0 || stderr != "" || !strings.Contains(stdout, "pm flow - plan, preview, run, list, and inspect multi-step flows") {
				t.Fatalf("help mismatch: code=%d stderr=%s stdout=%s", code, stderr, stdout)
			}
			if canonical == "" {
				canonical = stdout
			} else if stdout != canonical {
				t.Fatalf("%s help differs from canonical", tt.name)
			}
		})
	}

	stdout, stderr, code := runNativeFlowCLI("flow", "--json")
	if code != 0 || stderr != "" {
		t.Fatalf("JSON manual failed: code=%d stderr=%s", code, stderr)
	}
	var manual struct {
		Kind    string `json:"kind"`
		Command string `json:"command"`
		Manual  string `json:"manual"`
	}
	decodeOneJSON(t, stdout, &manual)
	if manual.Kind != "CommandManual" || manual.Command != "flow" || manual.Manual != canonical {
		t.Fatalf("JSON manual mismatch: %#v", manual)
	}

	root := t.TempDir()
	initProject(t, root)
	for _, args := range [][]string{
		{"flow", "bogus", "plan", "--root", root, "--json"},
		{"flow", "bogus", "--help", "--root", root, "--json"},
		{"flow", "--unknown", "plan", "--root", root, "--json"},
		{"flow", "--", "plan", "--root", root, "--json"},
		{"flow", "--=x", "plan", "--root", root, "--json"},
		{"flow", "---x", "plan", "--root", root, "--json"},
	} {
		stdout, stderr, code = runNativeFlowCLI(args...)
		assertCLIError(t, code, stdout, stderr, 2, "usage", "")
		if strings.Contains(stdout, `"status": "ok"`) {
			t.Fatalf("invalid action discovered later plan: %v", args)
		}
	}

	stdout, stderr, code = runNativeFlowCLI("--json", "--json=maybe", "flow")
	assertCLIError(t, code, stdout, stderr, 3, "validation", "invalid --json")
	stdout, stderr, code = runNativeFlowCLI("--json=false", "--plain=true", "--no-input=on", "flow")
	if code != 0 || stderr != "" || !strings.HasPrefix(stdout, "NAME\n  pm flow") {
		t.Fatalf("assigned global booleans mismatch: code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
}

func TestFlowExactErrorTaxonomyAndDeterministicTextOutput(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)
	manifestDir := t.TempDir()
	writeNativeFlowFixture(t, manifestDir, "stable")
	path := filepath.Join(manifestDir, "stable.json")

	stdout, stderr, code := runNativeFlowCLI("flow", "plan", "--root", root, "--json")
	assertCLIError(t, code, stdout, stderr, 2, "usage", "--file")

	malformed := filepath.Join(manifestDir, "malformed.json")
	if err := os.WriteFile(malformed, []byte(`{"version":`), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code = runNativeFlowCLI("flow", "plan", "--file", malformed, "--root", root, "--json")
	assertCLIError(t, code, stdout, stderr, 1, "internal", "manifest parse")

	firstOut, firstErr, firstCode := runNativeFlowCLI("flow", "plan", "--file", path, "--root", root)
	secondOut, secondErr, secondCode := runNativeFlowCLI("flow", "plan", "--file", path, "--root", root)
	if firstCode != 0 || secondCode != 0 || firstErr != "" || secondErr != "" || firstOut != secondOut {
		t.Fatalf("text plan is not deterministic: first=(%d,%q,%q) second=(%d,%q,%q)", firstCode, firstOut, firstErr, secondCode, secondOut, secondErr)
	}
	if firstOut != "Flow: stable  status=ok\n  1. score\n" {
		t.Fatalf("unexpected plan text: %q", firstOut)
	}
}

func assertFlowNoFileCompletion(t *testing.T, cmd *cobra.Command) {
	t.Helper()
	if cmd.ValidArgsFunction == nil {
		t.Fatalf("%s missing completion seam", cmd.CommandPath())
	}
	values, directive := cmd.ValidArgsFunction(cmd, nil, "")
	if len(values) != 0 || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("%s completion = (%v, %v), want no values and NoFileComp", cmd.CommandPath(), values, directive)
	}
}

func runNativeFlowCLI(args ...string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func writeNativeFlowFixture(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".json"), []byte(nativeFlowManifest(name)), 0o600); err != nil {
		t.Fatal(err)
	}
	spec := `{"name":"fixture-spec","features":[{"name":"email","weight":1,"score_if_set":1}]}`
	if err := os.WriteFile(filepath.Join(dir, "fixture-spec.json"), []byte(spec), 0o600); err != nil {
		t.Fatal(err)
	}
}

func nativeFlowManifest(name string) string {
	return `{
  "version": 1,
  "name": "` + name + `",
  "steps": [
    {
      "id": "score",
      "kind": "rlm",
      "spec": "fixture-spec.json",
      "mode": "fixture",
      "in": [],
      "out": ["` + name + `_scores"]
    }
  ]
}`
}

func assertFlowNDJSONSequence(t *testing.T, stderr string, want []string) {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(stderr), "\n")
	if len(lines) != len(want) {
		t.Fatalf("event line count=%d, want=%d: %s", len(lines), len(want), stderr)
	}
	for i, line := range lines {
		var event struct {
			Kind   string `json:"kind"`
			Scope  string `json:"scope"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("decode event %d: %v: %s", i, err, line)
		}
		if event.Scope != "flow" || event.Kind+":"+event.Status != want[i] {
			t.Fatalf("event[%d]=%#v, want %q", i, event, want[i])
		}
	}
}
