package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
)

// This compile-time reference is intentional TDD evidence: the focused test
// checkpoint must fail until reverse has a native Cobra constructor.
var _ = newReverseCobraCommand

func TestReverseCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	reverse := findCobraCommand(root, "reverse")
	if reverse == nil {
		t.Fatal("missing reverse command")
	}
	if reverse.DisableFlagParsing {
		t.Fatal("reverse command must use native Cobra flag parsing")
	}
	if reverse.ValidArgsFunction == nil {
		t.Fatal("reverse command must suppress file completion until Phase 15")
	}
	values, directive := reverse.ValidArgsFunction(reverse, nil, "")
	if len(values) != 0 || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("reverse completion = (%v, %v), want no values and NoFileComp", values, directive)
	}

	flagsByAction := map[string][]string{
		"list":    nil,
		"plan":    {"source-table", "destination", "map", "action", "limit"},
		"preview": nil,
		"run":     {"approve", "confirm"},
		"status":  nil,
	}
	for actionName, flagNames := range flagsByAction {
		t.Run(actionName, func(t *testing.T) {
			action := findCobraCommand(reverse, actionName)
			if action == nil {
				t.Fatalf("missing reverse %s command", actionName)
			}
			if action.DisableFlagParsing {
				t.Fatalf("reverse %s must use native Cobra flag parsing", actionName)
			}
			if !action.FParseErrWhitelist.UnknownFlags {
				t.Fatalf("reverse %s must preserve unknown-flag tolerance", actionName)
			}
			if action.ValidArgsFunction == nil {
				t.Fatalf("reverse %s missing no-file completion seam", actionName)
			}
			for _, flagName := range flagNames {
				flag := action.Flags().Lookup(flagName)
				if flag == nil {
					t.Fatalf("reverse %s missing --%s", actionName, flagName)
				}
				if got, want := flag.Value.Type(), "stringArray"; got != want {
					t.Fatalf("reverse %s --%s type = %q, want %q", actionName, flagName, got, want)
				}
				if got, want := flag.NoOptDefVal, "true"; got != want {
					t.Fatalf("reverse %s --%s NoOptDefVal = %q, want %q", actionName, flagName, got, want)
				}
			}
		})
	}

	help := findCobraCommand(reverse, "help")
	if help == nil || !help.Hidden {
		t.Fatal("reverse must preserve hidden positional help until Phase 19")
	}
	for _, spec := range cobraLegacyCommands(config.Config{}) {
		if spec.name == "reverse" {
			t.Fatal("reverse remains registered as a legacy Cobra wrapper")
		}
	}
}

func TestReverseLocalWorkflowPreservesFlagsRedactionConfirmationAndOrder(t *testing.T) {
	root := setupNativeReverseProject(t)
	outboxPath := filepath.Join(root, ".polymetrics", "outbox", "native_reverse.jsonl")

	planStdout, planStderr, code := runNativeReverseCLI(
		"reverse", "plan", "native_reverse", "ignored-positional",
		"--source-table", "missing", "--source-table=sample_customers",
		"--destination=outbox:missing", "--destination", "outbox:outbox-local",
		"--action", "ignored", "--action=upsert",
		"--map", "id:wrong", "--map=id:external_id",
		"--map", "name:full_name", "--map=email:email",
		"--limit", "1", "--limit=3", "--unknown", "ignored",
		"--root", root,
	)
	if code != 0 || planStderr != "" {
		t.Fatalf("local reverse plan failed: code=%d stderr=%s", code, planStderr)
	}
	planID := extractSensitiveReverseField(t, planStdout, `Created reverse plan (\S+)`)
	approval := extractSensitiveReverseField(t, planStdout, `Approval token: (\S+)`)
	if _, err := os.Stat(outboxPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("plan performed a local write: err=%v", err)
	}

	setReversePlanConfirmation(t, root, planID, "destructive")
	previewStdout, previewStderr, code := runNativeReverseCLI(
		"reverse", "preview", planID, "ignored-positional", "--unknown", "ignored", "--root", root, "--json",
	)
	if code != 0 || previewStderr != "" {
		t.Fatalf("local reverse preview failed: code=%d stderr=%s", code, previewStderr)
	}
	var preview struct {
		Kind string          `json:"kind"`
		Plan app.ReversePlan `json:"plan"`
	}
	decodeOneJSON(t, previewStdout, &preview)
	if preview.Kind != "ReversePlanPreview" || preview.Plan.ID != planID || preview.Plan.RecordCount != 3 || preview.Plan.ConfirmationChallenge != "destructive" {
		t.Fatalf("preview metadata mismatch: kind=%q id-match=%t records=%d confirmation=%q", preview.Kind, preview.Plan.ID == planID, preview.Plan.RecordCount, preview.Plan.ConfirmationChallenge)
	}
	assertApprovalAbsent(t, approval, previewStdout, previewStderr)
	if preview.Plan.ApprovalToken != "" || preview.Plan.ApprovalTokenHash != "" {
		t.Fatal("preview JSON retained approval material")
	}
	if _, err := os.Stat(outboxPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("preview performed a local write: err=%v", err)
	}

	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{name: "missing approval", args: []string{"reverse", "run", planID, "--root", root, "--json"}, want: "approval token is invalid"},
		{name: "wrong approval", args: []string{"reverse", "run", planID, "--approve", "not-the-approval", "--root", root, "--json"}, want: "approval token is invalid"},
		{name: "missing confirmation", args: []string{"reverse", "run", planID, "--approve", approval, "--root", root, "--json"}, want: "requires typed confirmation"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, gotCode := runNativeReverseCLI(tt.args...)
			if gotCode != 1 || !strings.Contains(stdout+stderr, tt.want) {
				t.Fatalf("gate rejection mismatch: code=%d want-message=%q", gotCode, tt.want)
			}
			assertApprovalAbsent(t, approval, stdout, stderr)
			if _, err := os.Stat(outboxPath); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("write occurred before all gates passed: err=%v", err)
			}
		})
	}

	runStdout, runStderr, code := runNativeReverseCLI(
		"reverse", "run", planID,
		"--approve=not-the-approval", "--approve", approval,
		"--confirm", "wrong", "--confirm=destructive",
		"--unknown=ignored", "--root", root, "--json",
	)
	if code != 0 || runStderr != "" {
		t.Fatalf("confirmed local reverse run failed: code=%d stderr=%s", code, runStderr)
	}
	var runEnv struct {
		Kind string         `json:"kind"`
		Run  app.ReverseRun `json:"run"`
	}
	decodeOneJSON(t, runStdout, &runEnv)
	if runEnv.Kind != "ReverseRun" || runEnv.Run.Status != "completed" || runEnv.Run.RecordsSucceeded != 3 {
		t.Fatalf("local run result mismatch: kind=%q status=%q succeeded=%d", runEnv.Kind, runEnv.Run.Status, runEnv.Run.RecordsSucceeded)
	}
	assertApprovalAbsent(t, approval, runStdout, runStderr)
	data, err := os.ReadFile(outboxPath)
	if err != nil {
		t.Fatalf("read local fake outbox: %v", err)
	}
	if got := len(strings.Split(strings.TrimSpace(string(data)), "\n")); got != 3 {
		t.Fatalf("local fake write count=%d, want 3", got)
	}

	statusStdout, statusStderr, code := runNativeReverseCLI("reverse", "status", runEnv.Run.ID, "ignored", "--root", root, "--json")
	if code != 0 || statusStderr != "" {
		t.Fatalf("local reverse status failed: code=%d stderr=%s", code, statusStderr)
	}
	assertJSONKind(t, statusStdout, "ReverseRun")
	assertApprovalAbsent(t, approval, statusStdout, statusStderr)

	listStdout, listStderr, code := runNativeReverseCLI("reverse", "list", "ignored", "--unknown", "ignored", "--root", root, "--json")
	if code != 0 || listStderr != "" {
		t.Fatalf("local reverse list failed: code=%d stderr=%s", code, listStderr)
	}
	assertJSONKind(t, listStdout, "ReversePlanList")
	assertApprovalAbsent(t, approval, listStdout, listStderr)

	replayStdout, replayStderr, code := runNativeReverseCLI("reverse", "run", planID, "--approve", approval, "--confirm", "destructive", "--root", root, "--json")
	if code != 1 || !strings.Contains(replayStdout+replayStderr, "already") {
		t.Fatalf("approval replay mismatch: code=%d", code)
	}
	assertApprovalAbsent(t, approval, replayStdout, replayStderr)
	assertApprovalAbsentFromFiles(t, root, approval)
}

func TestReversePlanJSONNeverDisclosesApprovalMaterial(t *testing.T) {
	root := setupNativeReverseProject(t)
	stdout, stderr, code := runNativeReverseCLI(
		"reverse", "plan", "json_redaction",
		"--source-table", "sample_customers",
		"--destination", "outbox:outbox-local",
		"--map", "id:external_id",
		"--limit", "1",
		"--root", root, "--json",
	)
	if code != 0 || stderr != "" {
		t.Fatalf("JSON plan failed: code=%d stderr=%s", code, stderr)
	}
	var env struct {
		Kind             string          `json:"kind"`
		Plan             app.ReversePlan `json:"plan"`
		ApprovalRequired bool            `json:"approval_required"`
	}
	decodeOneJSON(t, stdout, &env)
	if env.Kind != "ReversePlan" || !env.ApprovalRequired {
		t.Fatalf("JSON plan envelope mismatch: kind=%q approval-required=%t", env.Kind, env.ApprovalRequired)
	}
	if env.Plan.ApprovalToken != "" || env.Plan.ApprovalTokenHash != "" || strings.Contains(stdout, "approval_token") || strings.Contains(stdout, "approval_token_hash") {
		t.Fatal("JSON plan disclosed approval material")
	}
}

func TestReverseFirstOperandOwnershipFailsClosed(t *testing.T) {
	root := setupNativeReverseProject(t)
	planStdout, planStderr, code := runNativeReverseCLI(
		"reverse", "plan", "owned-plan", "--source-table", "sample_customers", "--destination", "outbox:outbox-local", "--map", "id:id", "--limit", "1", "--root", root,
	)
	if code != 0 || planStderr != "" {
		t.Fatalf("seed plan failed: code=%d stderr=%s", code, planStderr)
	}
	planID := extractSensitiveReverseField(t, planStdout, `Created reverse plan (\S+)`)
	approval := extractSensitiveReverseField(t, planStdout, `Approval token: (\S+)`)

	for _, tt := range []struct {
		name    string
		action  string
		operand string
		tail    []string
	}{
		{name: "preview long help", action: "preview", operand: "--help", tail: []string{planID}},
		{name: "preview short help", action: "preview", operand: "-h", tail: []string{planID}},
		{name: "preview literal separator", action: "preview", operand: "--", tail: []string{planID}},
		{name: "preview unknown flag", action: "preview", operand: "--unknown-preview-id", tail: []string{planID}},
		{name: "run long help", action: "run", operand: "--help", tail: []string{planID, "--approve", approval}},
		{name: "run carrier shaped", action: "run", operand: "--pm-internal-reverse-plan-id=" + planID, tail: []string{planID, "--approve", approval}},
		{name: "status literal separator", action: "status", operand: "--", tail: []string{"later-valid-run"}},
		{name: "status unknown flag", action: "status", operand: "--unknown-run-id", tail: []string{"later-valid-run"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"reverse", tt.action, tt.operand}
			args = append(args, tt.tail...)
			args = append(args, "--root", root, "--json")
			stdout, stderr, gotCode := runNativeReverseCLI(args...)
			if gotCode == 0 || !strings.Contains(stdout+stderr, tt.operand) {
				t.Fatalf("first operand was not owned fail-closed: action=%s operand=%q code=%d", tt.action, tt.operand, gotCode)
			}
			assertApprovalAbsent(t, approval, stdout, stderr)
		})
	}

	for _, operand := range []string{"--help", "-h", "--", "--unknown-plan-name", "--pm-internal-reverse-plan-name=owned-plan"} {
		t.Run("plan "+operand, func(t *testing.T) {
			stdout, stderr, gotCode := runNativeReverseCLI(
				"reverse", "plan", operand, "later-name",
				"--source-table", "sample_customers", "--destination", "outbox:outbox-local", "--map", "id:id", "--limit", "1", "--root", root, "--json",
			)
			if gotCode != 0 || stderr != "" {
				t.Fatalf("legacy plan-name operand rejected: operand=%q code=%d stderr=%s", operand, gotCode, stderr)
			}
			var env struct {
				Plan app.ReversePlan `json:"plan"`
			}
			decodeOneJSON(t, stdout, &env)
			if env.Plan.Name != operand {
				t.Fatalf("plan name=%q, want first operand %q", env.Plan.Name, operand)
			}
		})
	}
}

func TestReverseHelpTrailingLiteralUnknownActionsAndGlobals(t *testing.T) {
	var canonical string
	for _, tt := range []struct {
		name string
		args []string
	}{
		{name: "help topic", args: []string{"help", "reverse"}},
		{name: "bare", args: []string{"reverse"}},
		{name: "long", args: []string{"reverse", "--help"}},
		{name: "short", args: []string{"reverse", "-h"}},
		{name: "positional", args: []string{"reverse", "help"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, code := runNativeReverseCLI(tt.args...)
			if code != 0 || stderr != "" || !strings.Contains(stdout, "pm reverse - plan, preview, approve, and execute reverse ETL") {
				t.Fatalf("help route mismatch: code=%d stderr=%s", code, stderr)
			}
			if canonical == "" {
				canonical = stdout
			} else if stdout != canonical {
				t.Fatalf("help route %s differs from canonical manual", tt.name)
			}
		})
	}
	stdout, stderr, code := runNativeReverseCLI("reverse", "--json")
	if code != 0 || stderr != "" {
		t.Fatalf("JSON manual failed: code=%d stderr=%s", code, stderr)
	}
	var manual struct {
		Kind    string `json:"kind"`
		Command string `json:"command"`
		Manual  string `json:"manual"`
	}
	decodeOneJSON(t, stdout, &manual)
	if manual.Kind != "CommandManual" || manual.Command != "reverse" || manual.Manual != canonical {
		t.Fatalf("JSON manual mismatch: kind=%q command=%q", manual.Kind, manual.Command)
	}

	root := initNativeReverseProject(t)
	for _, args := range [][]string{
		{"reverse", "bogus", "plan", "--root", root, "--json"},
		{"reverse", "bogus", "--help", "--root", root, "--json"},
		{"reverse", "bogus", "-h", "--root", root, "--json"},
		{"reverse", "--unknown", "plan", "later", "--root", root, "--json"},
		{"reverse", "--", "plan", "later", "--root", root, "--json"},
	} {
		stdout, stderr, code = runNativeReverseCLI(args...)
		assertCLIError(t, code, stdout, stderr, 2, "usage", "")
		if strings.Contains(stdout, `"kind": "ReversePlan"`) {
			t.Fatalf("invalid action discovered and executed a later plan: args=%v", args)
		}
	}

	stdout, stderr, code = runNativeReverseCLI("reverse", "list", "--help", "--", "-h", "--unknown", "ignored", "--root", root, "--json")
	if code != 0 || stderr != "" {
		t.Fatalf("legacy action tail compatibility failed: code=%d stderr=%s", code, stderr)
	}
	assertJSONKind(t, stdout, "ReversePlanList")

	stdout, stderr, code = runNativeReverseCLI("--json", "--json=maybe", "reverse")
	assertCLIError(t, code, stdout, stderr, 3, "validation", "invalid --json")
	stdout, stderr, code = runNativeReverseCLI("--json=false", "--plain=true", "--no-input=on", "reverse")
	if code != 0 || stderr != "" || !strings.HasPrefix(stdout, "NAME\n  pm reverse") {
		t.Fatalf("assigned global booleans mismatch: code=%d stderr=%s", code, stderr)
	}
}

func TestNormalizeReverseLegacyActionArgsRewritesOnlyMalformedUnknowns(t *testing.T) {
	args := []string{
		"reverse", "run",
		"--approve=first",
		"--ordinary-unknown=value",
		"--=x",
		"---x",
		"ignored-unknown-value",
		"--confirm=typed",
		"--approve=last",
	}
	want := []string{
		"reverse", "run",
		"--approve=first",
		"--ordinary-unknown=value",
		"--pm-legacy-malformed-unknown=x",
		"--pm-legacy-malformed-unknown",
		"ignored-unknown-value",
		"--confirm=typed",
		"--approve=last",
	}
	got := normalizeReverseLegacyActionArgs(args, 2)
	if len(got) != len(want) {
		t.Fatalf("normalized args length=%d, want %d: %q", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("normalized arg[%d]=%q, want %q; all args=%q", i, got[i], want[i], got)
		}
	}
}

func TestReverseMalformedUnknownFlagsPreserveLegacyActionOutcomesAndNoEffects(t *testing.T) {
	root := initNativeReverseProject(t)
	statePath := filepath.Join(root, ".polymetrics", "state", "state.json")
	stateBefore, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read initial reverse state: %v", err)
	}

	actions := []struct {
		name string
		args []string
	}{
		{name: "list", args: []string{"reverse", "list"}},
		{name: "plan validation", args: []string{"reverse", "plan", "malformed-unknown-must-not-plan"}},
		{name: "preview missing plan", args: []string{"reverse", "preview", "missing-plan"}},
		{name: "run missing plan", args: []string{"reverse", "run", "missing-plan"}},
		{name: "status missing run", args: []string{"reverse", "status", "missing-run"}},
	}
	malformedUnknowns := []string{
		"--=x",
		"--=",
		"--==x",
		"---x",
		"---x=y",
		"---",
		"----x",
		"----x=y",
		"----",
		"-----x",
	}

	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			baselineArgs := append(append([]string(nil), action.args...), "--root", root, "--json")
			wantStdout, wantStderr, wantCode := runNativeReverseCLI(baselineArgs...)

			for _, malformed := range malformedUnknowns {
				t.Run(malformed, func(t *testing.T) {
					args := append(append([]string(nil), action.args...), malformed, "--root", root, "--json")
					stdout, stderr, code := runNativeReverseCLI(args...)
					if code != wantCode || stdout != wantStdout || stderr != wantStderr {
						t.Fatalf("malformed unknown changed action outcome: code=%d want=%d stdout-match=%t stderr-match=%t", code, wantCode, stdout == wantStdout, stderr == wantStderr)
					}
					if strings.Contains(stdout+stderr, "Approval token:") {
						t.Fatal("malformed unknown action disclosed approval output")
					}
					stateAfter, err := os.ReadFile(statePath)
					if err != nil {
						t.Fatalf("read reverse state after malformed unknown: %v", err)
					}
					if !bytes.Equal(stateAfter, stateBefore) {
						t.Fatal("malformed unknown action changed reverse state")
					}
					outboxEntries, err := os.ReadDir(filepath.Join(root, ".polymetrics", "outbox"))
					if err != nil && !errors.Is(err, os.ErrNotExist) {
						t.Fatalf("read local outbox after malformed unknown: %v", err)
					}
					if len(outboxEntries) != 0 {
						t.Fatalf("malformed unknown action created %d outbox effects", len(outboxEntries))
					}
				})
			}
		})
	}
}

func TestReverseExactExitTaxonomyAndBareFlags(t *testing.T) {
	root := setupNativeReverseProject(t)
	for _, tt := range []struct {
		name       string
		args       []string
		code       int
		category   string
		messageSub string
	}{
		{name: "missing plan name", args: []string{"reverse", "plan", "--root", root, "--json"}, code: 2, category: "usage", messageSub: "invalid usage"},
		{name: "malformed endpoint", args: []string{"reverse", "plan", "bad", "--destination", "bad", "--map", "id:id", "--root", root, "--json"}, code: 3, category: "validation", messageSub: "invalid endpoint"},
		{name: "malformed map", args: []string{"reverse", "plan", "bad", "--destination", "outbox:outbox-local", "--map", "id", "--root", root, "--json"}, code: 3, category: "validation", messageSub: "invalid mapping"},
		{name: "bare destination", args: []string{"reverse", "plan", "bad", "--destination", "--map=id:id", "--root", root, "--json"}, code: 3, category: "validation", messageSub: `invalid endpoint "true"`},
		{name: "bare limit", args: []string{"reverse", "plan", "bad", "--destination=outbox:outbox-local", "--map=id:id", "--limit", "--root", root, "--json"}, code: 3, category: "validation", messageSub: `invalid --limit "true"`},
		{name: "missing preview", args: []string{"reverse", "preview", "missing", "--root", root, "--json"}, code: 1, category: "internal", messageSub: `reverse plan "missing" not found`},
		{name: "bare approve", args: []string{"reverse", "run", "missing", "--approve", "--root", root, "--json"}, code: 1, category: "internal", messageSub: `reverse plan "missing" not found`},
		{name: "missing status", args: []string{"reverse", "status", "missing", "--root", root, "--json"}, code: 1, category: "internal", messageSub: `reverse run "missing" not found`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, gotCode := runNativeReverseCLI(tt.args...)
			assertCLIError(t, gotCode, stdout, stderr, tt.code, tt.category, tt.messageSub)
		})
	}
}

func TestReverseCancellationPropagatesThroughNativeCommand(t *testing.T) {
	root := setupNativeReverseProject(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout bytes.Buffer
	cmd := newRootCmd(ctx, config.Config{Root: root}, &stdout, io.Discard)
	err := executeRootCmd(cmd, []string{
		"reverse", "plan", "canceled", "--source-table", "sample_customers", "--destination", "outbox:outbox-local", "--map", "id:id",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("native reverse cancellation error=%v, want context.Canceled", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("canceled reverse wrote stdout: %q", stdout.String())
	}
}

func initNativeReverseProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	stdout, stderr, code := runNativeReverseCLI("init", "--root", root, "--json")
	if code != 0 {
		t.Fatalf("init local reverse project failed: code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	return root
}

func setupNativeReverseProject(t *testing.T) string {
	t.Helper()
	root := initNativeReverseProject(t)
	commands := [][]string{
		{"credentials", "add", "sample-local", "--connector", "sample", "--root", root, "--json"},
		{"credentials", "add", "warehouse-local", "--connector", "warehouse", "--config", "path=" + filepath.Join(root, ".polymetrics", "warehouse"), "--root", root, "--json"},
		{"credentials", "add", "outbox-local", "--connector", "outbox", "--config", "path=" + filepath.Join(root, ".polymetrics", "outbox"), "--root", root, "--json"},
		{"connections", "create", "sample-to-warehouse", "--source", "sample:sample-local", "--destination", "warehouse:warehouse-local", "--stream", "customers", "--sync-mode", "full_refresh_overwrite", "--table", "sample_customers", "--root", root, "--json"},
		{"etl", "run", "--connection", "sample-to-warehouse", "--stream", "customers", "--root", root, "--json"},
	}
	for _, args := range commands {
		stdout, stderr, code := runNativeReverseCLI(args...)
		if code != 0 {
			t.Fatalf("setup local reverse command failed: args=%v code=%d stdout=%s stderr=%s", args[:2], code, stdout, stderr)
		}
	}
	return root
}

func runNativeReverseCLI(args ...string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func extractSensitiveReverseField(t *testing.T, text, pattern string) string {
	t.Helper()
	match := regexp.MustCompile(pattern).FindStringSubmatch(text)
	if len(match) != 2 || match[1] == "" {
		t.Fatal("expected reverse field was not present")
	}
	return match[1]
}

func setReversePlanConfirmation(t *testing.T, root, planID, confirmation string) {
	t.Helper()
	path := filepath.Join(root, ".polymetrics", "state", "state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read local reverse state: %v", err)
	}
	var state map[string]any
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("decode local reverse state: %v", err)
	}
	plans, ok := state["reverse_plans"].([]any)
	if !ok {
		t.Fatal("local reverse state has no plans")
	}
	found := false
	for _, raw := range plans {
		plan, ok := raw.(map[string]any)
		if !ok || plan["id"] != planID {
			continue
		}
		plan["confirmation_challenge"] = confirmation
		found = true
		break
	}
	if !found {
		t.Fatal("local reverse plan not found for confirmation setup")
	}
	data, err = json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatalf("encode local reverse state: %v", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write local reverse state: %v", err)
	}
}

func assertApprovalAbsent(t *testing.T, approval string, values ...string) {
	t.Helper()
	for _, value := range values {
		if strings.Contains(value, approval) {
			t.Fatal("approval value leaked outside the human plan output")
		}
	}
}

func assertApprovalAbsentFromFiles(t *testing.T, root, approval string) {
	t.Helper()
	logsDir := filepath.Join(root, ".polymetrics", "logs")
	err := filepath.WalkDir(logsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Contains(data, []byte(approval)) {
			t.Fatal("approval value leaked to a local log")
		}
		return nil
	})
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("scan local reverse logs: %v", err)
	}
}
