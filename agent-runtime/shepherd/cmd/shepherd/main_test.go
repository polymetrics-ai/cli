package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	decisionlog "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/decision"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
	shepherdgithub "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/github"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/store"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/validation"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/workspace"
	_ "modernc.org/sqlite"
)

func initializedTestRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "test@example.invalid"}, {"config", "user.name", "Test"}} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "seed.txt"), []byte("seed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "seed.txt"}, {"commit", "-qm", "seed"}} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	return root
}

func TestExplicitDepthApprovalIsNarrowlyScoped(t *testing.T) {
	t.Parallel()

	question := gsd.Question{ID: "depth_verification_M001-n6ms9v_confirm", Method: "select", Title: "Confirm the printed depth verification is sufficient", Options: []string{"Confirm (Recommended)", "Reject depth"}}
	response, approved := approveDepthQuestion(question)
	if !approved || response.Value != question.Options[0] || response.Cancelled {
		t.Fatalf("response=%+v approved=%t", response, approved)
	}
	variant := question
	variant.Options = []string{"Depth verified (Recommended)", "Needs revision", "None of the above"}
	response, approved = approveDepthQuestion(variant)
	if !approved || response.Value != variant.Options[0] || response.Cancelled {
		t.Fatalf("variant response=%+v approved=%t", response, approved)
	}
	exact := question
	exact.ID = "depth_verification_M001_confirm"
	response, approved = approveDepthQuestion(exact)
	if !approved || response.Value != exact.Options[0] || response.Cancelled {
		t.Fatalf("exact response=%+v approved=%t", response, approved)
	}
	for _, other := range []gsd.Question{
		{Method: "confirm", Title: question.Title, Options: []string{"Yes", "No"}},
		{ID: "approve_dependency", Method: "select", Title: "Approve dependency addition", Options: question.Options},
		{ID: "depth_verification_M001-n6ms9v_review", Method: "select", Title: question.Title, Options: question.Options},
		{ID: "depth_verification_M001-n6ms9v_confirm", Method: "select", Title: question.Title, Options: []string{"Proceed", "Stop"}},
	} {
		if response, approved := approveDepthQuestion(other); approved || !response.Cancelled {
			t.Fatalf("unrelated gate was approved: question=%+v response=%+v", other, response)
		}
	}
}

func TestGovernedPathRejectsEscape(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	inside := filepath.Join(root, "context.md")
	if err := os.WriteFile(inside, []byte("context"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := governedPath(root, inside); err != nil {
		t.Fatalf("inside path rejected: %v", err)
	}
	outside := filepath.Join(t.TempDir(), "context.md")
	if err := os.WriteFile(outside, []byte("context"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := governedPath(root, outside); err == nil {
		t.Fatal("expected path escape to fail")
	}
	symlink := filepath.Join(root, "link.json")
	if err := os.Symlink(outside, symlink); err != nil {
		t.Fatal(err)
	}
	if _, err := governedPath(root, symlink); err == nil {
		t.Fatal("expected symlink escape to fail")
	}
}

func TestDeliveryIDIsStablePerIssue(t *testing.T) {
	t.Parallel()
	if got := deliveryID(372); got != "issue-372" {
		t.Fatalf("delivery id=%q", got)
	}
}

func TestTerminalDiagnosticIsSingleLineAndBounded(t *testing.T) {
	t.Parallel()
	got := terminalDiagnostic("first\n[headless] Error: session switch cancelled\x00\n")
	if got != "[headless] Error: session switch cancelled" {
		t.Fatalf("diagnostic=%q", got)
	}
	if len(terminalDiagnostic(string(make([]byte, 2000)))) > 512 {
		t.Fatal("diagnostic was not bounded")
	}
}

func TestJoinTerminalFailurePreservesPrimaryCause(t *testing.T) {
	t.Parallel()
	primary := errors.New("canonical unit did not advance")
	secondary := errors.New("worktree must be clean")
	joined := joinTerminalFailure(primary, secondary)
	if !errors.Is(joined, primary) || !errors.Is(joined, secondary) {
		t.Fatalf("joined error=%v", joined)
	}
}

func TestMonitorWriteScopeReportsForbiddenWriteAndStopsOnCancellation(t *testing.T) {
	t.Parallel()
	root := initializedTestRepository(t)
	ctx, cancel := context.WithCancel(context.Background())
	results := monitorWriteScope(ctx, root, []string{"internal/cli/**"}, 5*time.Millisecond)

	outside := filepath.Join(root, "docs", "connectors", "100ms", "MANUAL.md")
	if err := os.MkdirAll(filepath.Dir(outside), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, []byte("unexpected churn\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-results:
		if err == nil || !strings.Contains(err.Error(), "docs/connectors/100ms/MANUAL.md") {
			t.Fatalf("scope monitor error=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("scope monitor did not report forbidden write")
	}

	allowedCtx, allowedCancel := context.WithCancel(context.Background())
	allowedResults := monitorWriteScope(allowedCtx, root, []string{"docs/connectors/**"}, 5*time.Millisecond)
	allowedCancel()
	select {
	case err := <-allowedResults:
		if err != nil {
			t.Fatalf("cancelled allowed monitor error=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("scope monitor leaked after cancellation")
	}
	cancel()
}

func TestMonitorWriteScopeFailsClosedWhenRepositoryCannotBeInspected(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	results := monitorWriteScope(ctx, filepath.Join(t.TempDir(), "missing"), []string{"internal/cli/**"}, time.Millisecond)
	select {
	case err := <-results:
		if err == nil || !strings.Contains(err.Error(), "inspect live write scope") {
			t.Fatalf("scope monitor error=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("scope monitor did not fail closed")
	}
}

func TestFinalUnitRunStateAwaitsDecisionWhenRetryBudgetExhausted(t *testing.T) {
	t.Parallel()
	result := gsd.Result{Terminal: gsd.TerminalError, Err: errors.New("artifact missing: M001-ROADMAP.md")}
	if got := finalUnitRunState(&result, "executing", 0); got != domain.RunAwaitingDecision {
		t.Fatalf("state=%s want awaiting_decision", got)
	}
	if !errors.Is(result.Err, store.ErrRetryBudgetExhausted) {
		t.Fatalf("error=%v, want retry budget exhausted", result.Err)
	}
	if class := classifyUnitFailure(result); class != unitFailureRetryExhausted {
		t.Fatalf("class=%s want %s", class, unitFailureRetryExhausted)
	}
}

func TestFinalUnitRunStateRetriesWhileBudgetRemains(t *testing.T) {
	t.Parallel()
	result := gsd.Result{Terminal: gsd.TerminalError, Err: errors.New("artifact missing: M001-ROADMAP.md")}
	if got := finalUnitRunState(&result, "executing", 1); got != domain.RunReady {
		t.Fatalf("state=%s want ready", got)
	}
	if errors.Is(result.Err, store.ErrRetryBudgetExhausted) {
		t.Fatalf("unexpected exhausted marker: %v", result.Err)
	}
}

func TestClassifyUnitFailureAndRetryPolicy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		result    gsd.Result
		class     string
		retryable bool
	}{
		{name: "runtime contract", result: gsd.Result{Terminal: gsd.TerminalError, Err: gsd.ErrRuntimeContractMismatch}, class: unitFailureRuntimeContractMismatch, retryable: false},
		{name: "missing artifact", result: gsd.Result{Terminal: gsd.TerminalError, Err: errors.New("artifact missing: M001-ROADMAP.md")}, class: unitFailureArtifactMissing, retryable: true},
		{name: "interrupted", result: gsd.Result{Terminal: gsd.TerminalTimeout, Err: context.DeadlineExceeded}, class: unitFailureInterrupted, retryable: true},
		{name: "stale head", result: gsd.Result{Terminal: gsd.TerminalError, Err: errors.New("stale head before checkpoint")}, class: unitFailureStaleHead, retryable: false},
		{name: "scope breach", result: gsd.Result{Terminal: gsd.TerminalError, Err: errors.New("live write-scope breach: changed path outside the issue write scope")}, class: unitFailureScopeBreach, retryable: false},
		{name: "model drift", result: gsd.Result{Terminal: gsd.TerminalError, Err: errors.New("effective runtime identity was not observed as openai-codex/gpt-5.5/high")}, class: unitFailureModelDrift, retryable: false},
		{name: "orphan child", result: gsd.Result{Terminal: gsd.TerminalError, Err: errors.New("orphaned subagent still running")}, class: unitFailureOrphanedSubagent, retryable: false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := classifyUnitFailure(tc.result); got != tc.class {
				t.Fatalf("class=%s want %s", got, tc.class)
			}
			if got := isAutomaticallyRetryable(tc.result.Err); got != tc.retryable {
				t.Fatalf("retryable=%v want %v", got, tc.retryable)
			}
		})
	}
}

func TestMutatingSkipFenceLeavesAuthorityReady(t *testing.T) {
	t.Parallel()
	if got := targetRunState(gsd.TerminalBlocked, gsd.ErrMutatingSkip, "executing"); got != domain.RunReady {
		t.Fatalf("target state=%s want ready", got)
	}
	if got := targetRunState(gsd.TerminalBlocked, errors.New("real blocker"), "executing"); got != domain.RunBlocked {
		t.Fatalf("real blocker target state=%s", got)
	}
}

func TestIndependentValidatorFactoryUsesConfiguredPiNotGSD(t *testing.T) {
	t.Parallel()
	config := fileConfig{
		PiCommand: []string{"/configured/pi"}, GSDCommand: []string{"/pinned/gsd"},
		GSDHome: "/protected/home", StateDir: "/protected/state", TimeoutSeconds: 60,
	}
	validator, ok := independentValidatorFactory(nil, config).(validation.GSDValidator)
	if !ok {
		t.Fatalf("validator type=%T", independentValidatorFactory(nil, config))
	}
	if len(validator.Command) != 1 || validator.Command[0] != config.PiCommand[0] {
		t.Fatalf("validator command=%v want Pi %v", validator.Command, config.PiCommand)
	}
	if validator.Command[0] == config.GSDCommand[0] || validator.SessionsDir != "" {
		t.Fatalf("validator must not use GSD or shared sessions: %+v", validator)
	}
}

func TestLoadConfigRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.json")
	raw := `{"gsd_command":["gsd"],"work_dir":"/tmp/work","gsd_home":"/tmp/home","state_dir":"/tmp/state","attempt_root":"/tmp/attempts","unexpected":true}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(path); err == nil {
		t.Fatal("expected unknown config field to fail")
	}
}

func TestLoadConfigRequiresAllowlistedPiExecutable(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.json")
	raw := `{"gsd_command":["gsd"],"work_dir":"/tmp/work","gsd_home":"/tmp/home","state_dir":"/tmp/state","attempt_root":"/tmp/attempts","repository":"polymetrics-ai/cli","pull_request":388}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(path); err == nil || !strings.Contains(err.Error(), "pi_command") {
		t.Fatalf("missing pi_command err=%v", err)
	}
}

func TestLoadConfigRequiresCompleteDecisionPRBinding(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.json")
	raw := `{"gsd_command":["gsd"],"work_dir":"/tmp/work","gsd_home":"/tmp/home","state_dir":"/tmp/state","attempt_root":"/tmp/attempts","repository":"polymetrics-ai/cli"}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(path); err == nil {
		t.Fatal("partial PR binding accepted")
	}
}

func TestLoadConfigDefaultsToBoundedNestedAgentEnvelopeSize(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.json")
	raw := `{"gsd_command":["gsd"],"pi_command":["/usr/bin/true"],"work_dir":"/tmp/work","gsd_home":"/tmp/home","state_dir":"/tmp/state","attempt_root":"/tmp/attempts","repository":"polymetrics-ai/cli","pull_request":388}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if config.MaxEventBytes < 4_000_000 || config.MaxEventBytes > 8*1024*1024 {
		t.Fatalf("max event bytes=%d does not admit measured nested-agent envelopes within 8 MiB", config.MaxEventBytes)
	}
	if config.CoordinatorModel != "openai-codex/gpt-5.6-sol" || config.ImplementationModel != "openai-codex/gpt-5.5" {
		t.Fatalf("model split coordinator=%q implementation=%q", config.CoordinatorModel, config.ImplementationModel)
	}
}

func TestLoadConfigRejectsAttemptRootContainingProtectedState(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "config.json")
	raw := `{"gsd_command":["gsd"],"pi_command":["/usr/bin/true"],"work_dir":"/tmp/work","gsd_home":"/tmp/home","state_dir":"/tmp/protected-state","attempt_root":"/tmp/protected-state/attempts","repository":"polymetrics-ai/cli","pull_request":388}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(path); err == nil || !strings.Contains(err.Error(), "attempt_root") {
		t.Fatalf("nested attempt root err=%v", err)
	}
}

func TestModelForUnitTypeUsesOfficialPhaseMetadata(t *testing.T) {
	t.Parallel()
	config := fileConfig{CoordinatorModel: "openai-codex/gpt-5.6-sol", ImplementationModel: "openai-codex/gpt-5.5"}
	registry := gsd.BuiltinUnitRegistry()
	for _, unitType := range []string{"execute-task", "execute-task-simple", "reactive-execute", "quick-task"} {
		got, err := modelForUnitType(config, registry, unitType)
		if err != nil || got != config.ImplementationModel {
			t.Fatalf("%s model=%q err=%v, want implementation %q", unitType, got, err, config.ImplementationModel)
		}
	}
	for _, unitType := range []string{"research-slice", "plan-slice", "complete-slice", "validate-milestone", "run-uat", "complete-milestone"} {
		got, err := modelForUnitType(config, registry, unitType)
		if err != nil || got != config.CoordinatorModel {
			t.Fatalf("%s model=%q err=%v, want coordinator %q", unitType, got, err, config.CoordinatorModel)
		}
	}
}

func TestLaunchModelForCommandUsesOfficialPhaseMetadata(t *testing.T) {
	t.Parallel()
	config := fileConfig{CoordinatorModel: "openai-codex/gpt-5.6-sol", ImplementationModel: "openai-codex/gpt-5.5"}
	registry := gsd.BuiltinUnitRegistry()
	got, err := launchModelForCommand(config, registry, "execute-task-simple")
	if err != nil || got != config.ImplementationModel {
		t.Fatalf("execute-task-simple launch model=%q err=%v, want implementation %q", got, err, config.ImplementationModel)
	}
	for _, command := range []string{"plan-slice", "validate-milestone"} {
		got, err := launchModelForCommand(config, registry, command)
		if err != nil || got != config.CoordinatorModel {
			t.Fatalf("%s launch model=%q err=%v, want coordinator %q", command, got, err, config.CoordinatorModel)
		}
	}
}

type recordingDecisionPublisher struct {
	summaries []string
	err       error
}

func (p *recordingDecisionPublisher) SyncDecisionComment(_ context.Context, _ shepherdgithub.Target, summary string) error {
	p.summaries = append(p.summaries, summary)
	return p.err
}

func TestAppendAndPublishDecisionRetainsDurableRecordWhenPublicationFails(t *testing.T) {
	t.Parallel()

	store, err := decisionlog.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	publisher := &recordingDecisionPublisher{err: errors.New("offline")}
	question := gsd.Question{ID: "scope", Title: "What should ship?"}
	response := gsd.UIResponse{Value: "Full safe parity"}
	err = appendAndPublishDecision(context.Background(), store, publisher,
		shepherdgithub.Target{Repository: "polymetrics-ai/cli", PullRequest: 388, DeliveryID: "issue-380"},
		"issue-380", "execution-1", "discuss-milestone/M001", question, response, "shepherd", "approved issue context")
	if err == nil {
		t.Fatal("publication failure did not fail the answered gate")
	}
	records, readErr := store.Records()
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(records) != 1 || records[0].Actor != decisionlog.ActorShepherd || records[0].At.Before(time.Now().Add(-time.Minute)) {
		t.Fatalf("durable records=%+v", records)
	}
	if len(publisher.summaries) != 1 {
		t.Fatalf("publication attempts=%d", len(publisher.summaries))
	}
}

func TestRecordOperationalDecisionPublishesAttributedLedgerEntry(t *testing.T) {
	t.Parallel()

	store, err := decisionlog.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	publisher := &recordingDecisionPublisher{}
	err = recordOperationalDecision(context.Background(), store, publisher,
		shepherdgithub.Target{Repository: "polymetrics-ai/cli", PullRequest: 388, DeliveryID: "issue-380"},
		"issue-380", "reconcile/M001/S01/T04", "Amend the issue write scope?", "Add only internal/connectors/defs/defs_test.go",
		"shepherd", "approved task plan requires the production embed assertion")
	if err != nil {
		t.Fatal(err)
	}
	records, err := store.Records()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].Actor != decisionlog.ActorShepherd || records[0].Question != "Amend the issue write scope?" {
		t.Fatalf("records=%+v", records)
	}
	if len(publisher.summaries) != 1 || !strings.Contains(publisher.summaries[0], "Add only internal/connectors/defs/defs_test.go") {
		t.Fatalf("published summaries=%v", publisher.summaries)
	}
}

func TestMaterializeContainerContextCopiesPlanningFileIntoProtectedOverlay(t *testing.T) {
	t.Parallel()
	workDir := t.TempDir()
	contextPath := filepath.Join(workDir, ".planning", "phases", "issue-380", "ISSUE-CONTEXT.json")
	stateDir := t.TempDir()
	raw := []byte(`{"issue":380}`)
	config := fileConfig{Runtime: "podman", WorkDir: workDir, StateDir: stateDir}
	if err := materializeContainerContext(config, contextPath, raw); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(stateDir, "runtime", "planning", "phases", "issue-380", "ISSUE-CONTEXT.json")
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(raw) {
		t.Fatalf("protected context=%q, want %q", got, raw)
	}
}

func TestProtectedIssueContextIsImmutableAndHashBound(t *testing.T) {
	t.Parallel()
	stateDir := t.TempDir()
	raw := []byte(`{"issue":380,"parent_issue":380,"objective":"Asana parity","scope":["connector"],"non_goals":[],"acceptance_criteria":["green"],"dependencies":[],"write_scope":["internal/connectors/engine/**"],"required_reading":["AGENTS.md"],"required_skills":["golang-how-to"],"tdd":{"red":"fail","green":"pass","refactor":"clean"},"verification":["go test ./..."],"safety":["no secrets"],"human_gates":["main merge"],"branch":"feat/380-asana-cli-parity","pr_base":"main","review_route":"local","sources":["https://github.com/polymetrics-ai/cli/issues/380"]}`)
	if err := materializeProtectedIssueContext(stateDir, 380, raw); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(raw)
	context, err := loadProtectedIssueContext(stateDir, 380, "sha256:"+hex.EncodeToString(hash[:]))
	if err != nil {
		t.Fatal(err)
	}
	if len(context.WriteScope) != 1 || context.WriteScope[0] != "internal/connectors/engine/**" {
		t.Fatalf("context=%+v", context)
	}
	if err := materializeProtectedIssueContext(stateDir, 380, append(raw, ' ')); err == nil {
		t.Fatal("protected context overwrite accepted")
	}
	if _, err := loadProtectedIssueContext(stateDir, 380, "sha256:"+strings.Repeat("0", 64)); err == nil {
		t.Fatal("context hash mismatch accepted")
	}
}

func TestSuperviseFakeRuntimeToFinalHumanGate(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	repo := initializedTestRepository(t)
	branchRaw := runGitForTest(t, repo, "symbolic-ref", "--quiet", "--short", "HEAD")
	contextPath := filepath.Join(repo, "issue-context.json")
	contextRaw := `{
  "issue":389,
  "parent_issue":372,
  "objective":"fake supervise integration",
  "scope":["fake supervise"],
  "non_goals":["live github"],
  "acceptance_criteria":["final human gate"],
  "dependencies":[],
  "write_scope":["agent-runtime/shepherd/**"],
  "required_reading":["AGENTS.md"],
  "required_skills":["golang-how-to"],
  "tdd":{"red":"fake runtime fails before supervise","green":"supervise reaches final gate","refactor":"keep fake bounded"},
  "verification":["go test ./..."],
  "safety":["no secrets"],
  "human_gates":["parent merge"],
  "branch":"` + strings.TrimSpace(branchRaw) + `",
  "pr_base":"main",
  "review_route":"local",
  "sources":["issue #389"]
}`
	if err := os.WriteFile(contextPath, []byte(contextRaw), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, ".gsd"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".gsd", "STATE.md"), []byte("fake canonical state\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitForTest(t, repo, "add", "issue-context.json")
	runGitForTest(t, repo, "commit", "-m", "test: add issue context")
	stateDir := filepath.Join(t.TempDir(), "state")
	gsdHome := filepath.Join(t.TempDir(), "home")
	if err := os.MkdirAll(filepath.Join(gsdHome, "agent"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gsdHome, "agent", "settings.json"), []byte(`{"defaultProvider":"openai-codex","defaultModel":"gpt-5.6-sol","defaultThinkingLevel":"high"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	runner, err := gsd.NewRunner(gsd.Config{
		Command: []string{os.Args[0], "-test.run=TestSuperviseFakeRuntimeHelper", "--"},
		WorkDir: repo, GSDHome: gsdHome, StateDir: stateDir,
		Model: "openai-codex/gpt-5.6-sol", Thinking: "high", Timeout: 10 * time.Second,
		Environment: []string{"GO_WANT_RUNNER_HELPER=supervise"},
	})
	if err != nil {
		t.Fatal(err)
	}
	config := fileConfig{WorkDir: repo, GSDHome: gsdHome, StateDir: stateDir, AttemptRoot: filepath.Join(t.TempDir(), "attempt-worktrees"), CoordinatorModel: "openai-codex/gpt-5.6-sol", ImplementationModel: "openai-codex/gpt-5.5", GSDVersion: "1.11.0", TimeoutSeconds: 10, HeartbeatSeconds: 1, MaxEventBytes: defaultMaxEventBytes, MaxUnitAttempts: 2, Repository: "polymetrics-ai/cli", PullRequest: 391, Runtime: "host"}
	restore := installFakeIndependentValidator(t, &fakeIndependentValidator{result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED")})
	defer restore()
	if err := runSupervise(context.Background(), runner, config, gsd.BuiltinUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(repo, "agent-runtime", "shepherd", "canary.txt")); err != nil {
		t.Fatalf("promoted canary artifact missing: %v", err)
	}
	if count := countSQLiteQueryForTest(t, filepath.Join(config.StateDir, "authority.db"), "SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'cleanup_complete'"); count != 1 {
		t.Fatalf("cleanup-complete attempt records=%d want 1", count)
	}
}

func TestSuperviseRejectsInvalidIndependentValidationWithoutPromotion(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	for _, test := range []struct {
		name   string
		result validation.Result
	}{
		{name: "missing validator evidence", result: validation.Result{}},
		{name: "gpt 5.5 validator", result: validFakeValidationResult("openai-codex/gpt-5.5", "PROCEED")},
		{name: "retry verdict", result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "RETRY")},
		{name: "halt verdict", result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "HALT")},
		{name: "stale candidate head", result: func() validation.Result {
			result := validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED")
			result.ObservedHead = strings.Repeat("9", 40)
			return result
		}()},
		{name: "failed local gates", result: func() validation.Result {
			result := validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED")
			result.LocalGates = false
			return result
		}()},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
			beforeHead := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD"))
			beforeGSD := readFileForTest(t, filepath.Join(repo, ".gsd", "STATE.md"))
			validator := &fakeIndependentValidator{result: test.result}
			restore := installFakeIndependentValidator(t, validator)
			defer restore()

			err := runSupervise(context.Background(), runner, config, gsd.BuiltinUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime")
			if err == nil {
				t.Fatal("invalid validation unexpectedly promoted")
			}
			if got := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD")); got != beforeHead {
				t.Fatalf("canonical HEAD changed: got %s want %s", got, beforeHead)
			}
			if got := readFileForTest(t, filepath.Join(repo, ".gsd", "STATE.md")); got != beforeGSD {
				t.Fatalf("canonical GSD state changed: got %q want %q", got, beforeGSD)
			}
			if _, err := os.Stat(filepath.Join(repo, "agent-runtime", "shepherd", "canary.txt")); !os.IsNotExist(err) {
				t.Fatalf("canonical canary artifact should not be promoted: %v", err)
			}
			if count := countSQLiteRowsForTest(t, filepath.Join(config.StateDir, "authority.db"), "artifact_proofs"); count != 0 {
				t.Fatalf("artifact proofs persisted on rejected validation: %d", count)
			}
			if count := countSQLiteRowsForTest(t, filepath.Join(config.StateDir, "authority.db"), "attestations"); count != 0 {
				t.Fatalf("attestations persisted on rejected validation: %d", count)
			}
			if validator.calls != 1 {
				t.Fatalf("validator calls=%d want 1", validator.calls)
			}
			if count := countSQLiteQueryForTest(t, filepath.Join(config.StateDir, "authority.db"), "SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'retained_for_recovery'"); count != 1 {
				t.Fatalf("retained attempt records=%d want 1", count)
			}
		})
	}
}

func TestSupervisePersistsPreparationAndPreDispatchFailures(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	for _, test := range []struct {
		name         string
		failureClass string
		arrange      func(*testing.T, string) func()
	}{
		{name: "preparation", failureClass: "preparation_failure", arrange: func(_ *testing.T, _ string) func() {
			previous := prepareAttemptGSDState
			prepareAttemptGSDState = func(context.Context, *workspace.Manager, workspace.AttemptWorktree) error {
				return errors.New("bounded preparation failure")
			}
			return func() { prepareAttemptGSDState = previous }
		}},
		{name: "pre-dispatch query", failureClass: "pre_dispatch_query_failure", arrange: func(t *testing.T, _ string) func() {
			t.Setenv("RUNNER_HELPER_MODE", "fail-attempt-query")
			return func() {}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
			restore := test.arrange(t, repo)
			defer restore()
			err := runSupervise(context.Background(), runner, config, gsd.BuiltinUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime")
			if err == nil {
				t.Fatal("early failure unexpectedly succeeded")
			}
			query := "SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'retained_for_recovery' AND failure_class = '" + test.failureClass + "'"
			if count := countSQLiteQueryForTest(t, filepath.Join(config.StateDir, "authority.db"), query); count != 1 {
				t.Fatalf("durable %s records=%d want 1; supervise err=%v", test.failureClass, count, err)
			}
		})
	}
}

func TestSuperviseRuntimeRetryReconcilesOldAttemptAndRetainsFreshAttempt(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	t.Setenv("RUNNER_HELPER_MODE", "fail-runtime")
	_, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	err := runSupervise(context.Background(), runner, config, gsd.BuiltinUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime")
	if err == nil {
		t.Fatal("runtime failure unexpectedly succeeded")
	}
	dbPath := filepath.Join(config.StateDir, "authority.db")
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'cleanup_complete'"); count != 1 {
		t.Fatalf("reconciled attempts=%d want 1", count)
	}
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'retained_for_recovery' AND failure_class = 'runtime_failure'"); count != 1 {
		t.Fatalf("retained runtime attempts=%d want 1", count)
	}
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(DISTINCT path) FROM attempt_worktrees"); count != 2 {
		t.Fatalf("distinct retry worktree paths=%d want 2", count)
	}
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(DISTINCT branch) FROM attempt_worktrees"); count != 2 {
		t.Fatalf("distinct retry branches=%d want 2", count)
	}
}

func TestSuperviseStartupReconcilesRetainedAttemptBeforeFinalGateQuery(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	_, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	restore := installFakeIndependentValidator(t, &fakeIndependentValidator{result: validation.Result{}})
	defer restore()
	if err := runSupervise(context.Background(), runner, config, gsd.BuiltinUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("rejected validation unexpectedly succeeded")
	}
	restore()
	dbPath := filepath.Join(config.StateDir, "authority.db")
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'retained_for_recovery'"); count != 1 {
		t.Fatalf("retained attempts=%d want 1", count)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE delivery_runs SET state = 'running', owner = 'crashed-owner'`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := runSupervise(context.Background(), runner, config, gsd.BuiltinUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err != nil {
		t.Fatalf("final-gate startup reconciliation: %v", err)
	}
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'cleanup_complete'"); count != 1 {
		t.Fatalf("startup cleanup-complete attempts=%d want 1", count)
	}
	if err := runSupervise(context.Background(), runner, config, gsd.BuiltinUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err != nil {
		t.Fatalf("idempotent final-gate restart: %v", err)
	}
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'cleanup_complete'"); count != 1 {
		t.Fatalf("idempotent cleanup-complete attempts=%d want 1", count)
	}
}

func TestSuperviseStartupPreservesAmbiguousRunningAttempt(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	previous := prepareAttemptGSDState
	defer func() { prepareAttemptGSDState = previous }()
	prepareAttemptGSDState = func(context.Context, *workspace.Manager, workspace.AttemptWorktree) error {
		return errors.New("bounded preparation failure")
	}
	if err := runSupervise(context.Background(), runner, config, gsd.BuiltinUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("preparation failure unexpectedly succeeded")
	}
	prepareAttemptGSDState = previous
	dbPath := filepath.Join(config.StateDir, "authority.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE attempt_worktrees SET state = 'running', candidate_head = ''`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE delivery_runs SET state = 'running', owner = 'crashed-owner'`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	var attemptPath, attemptBranch string
	if err := db.QueryRow(`SELECT path, branch FROM attempt_worktrees`).Scan(&attemptPath, &attemptBranch); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := runSupervise(context.Background(), runner, config, gsd.BuiltinUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("ambiguous running attempt did not block startup")
	}
	if _, err := os.Stat(attemptPath); err != nil {
		t.Fatalf("ambiguous running worktree was removed: %v", err)
	}
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'running'"); count != 1 {
		t.Fatalf("ambiguous running state changed: %d", count)
	}
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM delivery_runs WHERE state = 'awaiting_decision'"); count != 1 {
		t.Fatalf("interrupted delivery was not durably blocked: %d", count)
	}
	authorityStore, err := store.Open(context.Background(), dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := resolveHumanClearedAttempts(context.Background(), authorityStore, config, "issue-389"); err == nil {
		_ = authorityStore.Close()
		t.Fatal("human resume accepted while ambiguous resources remained")
	}
	runGitForTest(t, repo, "worktree", "remove", "--force", attemptPath)
	runGitForTest(t, repo, "branch", "-D", "--", attemptBranch)
	if err := resolveHumanClearedAttempts(context.Background(), authorityStore, config, "issue-389"); err != nil {
		_ = authorityStore.Close()
		t.Fatalf("human-cleared resolution: %v", err)
	}
	if err := authorityStore.ResumeDelivery(context.Background(), domain.HumanDecision{RunID: "issue-389", Generation: 1, ActorKind: domain.ActorHuman, Approved: true}); err != nil {
		_ = authorityStore.Close()
		t.Fatal(err)
	}
	if err := authorityStore.Close(); err != nil {
		t.Fatal(err)
	}
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'cleanup_complete'"); count != 1 {
		t.Fatalf("human-cleared attempt did not converge: %d", count)
	}
}

func TestSuperviseRetainsMovedCandidateBeforePromotionIntent(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	_, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	validator := &fakeIndependentValidator{result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED"), onValidate: func(_ context.Context, request validation.Request) error {
		runGitForTest(t, request.WorkDir, "commit", "--allow-empty", "-m", "move candidate during validation")
		return nil
	}}
	restore := installFakeIndependentValidator(t, validator)
	defer restore()
	if err := runSupervise(context.Background(), runner, config, gsd.BuiltinUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("moved candidate unexpectedly promoted")
	}
	dbPath := filepath.Join(config.StateDir, "authority.db")
	var attemptPath string
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT path FROM attempt_worktrees WHERE state = 'retained_for_recovery'`).Scan(&attemptPath); err != nil {
		_ = db.Close()
		t.Fatalf("moved candidate was not durably retained: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := runSupervise(context.Background(), runner, config, gsd.BuiltinUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("moved retained candidate did not remain human-gated")
	}
	if _, err := os.Stat(attemptPath); err != nil {
		t.Fatalf("moved retained attempt was changed: %v", err)
	}
}

func TestSupervisePersistsCleanupBlockedWithoutDeletingAttempt(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	_, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	restoreValidator := installFakeIndependentValidator(t, &fakeIndependentValidator{result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED")})
	defer restoreValidator()
	previous := reconcileOwnedAttempt
	reconcileOwnedAttempt = func(context.Context, *workspace.Manager, workspace.OwnedAttempt) (workspace.ReconcileResult, error) {
		return workspace.ReconcileBlocked, errors.New("bounded cleanup failure")
	}
	defer func() { reconcileOwnedAttempt = previous }()
	if err := runSupervise(context.Background(), runner, config, gsd.BuiltinUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err == nil {
		t.Fatal("cleanup failure unexpectedly succeeded")
	}
	dbPath := filepath.Join(config.StateDir, "authority.db")
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'cleanup_blocked' AND cleanup_error <> ''"); count != 1 {
		t.Fatalf("cleanup-blocked attempts=%d want 1", count)
	}
	if count := countSQLiteQueryForTest(t, dbPath, "SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'cleanup_blocked' AND length(cleanup_error) <= 512"); count != 1 {
		t.Fatalf("cleanup diagnostics were not bounded")
	}
}

func TestSuperviseRatifiesBeforePromotingCandidate(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "" {
		return
	}
	repo, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	beforeHead := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD"))
	beforeGSD := readFileForTest(t, filepath.Join(repo, ".gsd", "STATE.md"))
	validator := &fakeIndependentValidator{result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED"), onValidate: func(_ context.Context, request validation.Request) error {
		if got := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD")); got != beforeHead {
			t.Fatalf("candidate was promoted before validation/ratification: got %s want %s", got, beforeHead)
		}
		if got := readFileForTest(t, filepath.Join(repo, ".gsd", "STATE.md")); got != beforeGSD {
			t.Fatalf("canonical GSD changed before validation/ratification: got %q want %q", got, beforeGSD)
		}
		if request.BaseHead != beforeHead || request.CandidateHead == "" || request.CandidateHead == beforeHead {
			t.Fatalf("validator request not bound to base/candidate heads: %+v", request)
		}
		return nil
	}}
	restore := installFakeIndependentValidator(t, validator)
	defer restore()

	if err := runSupervise(context.Background(), runner, config, gsd.BuiltinUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(runGitForTest(t, repo, "rev-parse", "HEAD")); got == beforeHead {
		t.Fatal("ratified candidate was not promoted")
	}
	if got := readFileForTest(t, filepath.Join(repo, ".gsd", "STATE.md")); got == beforeGSD {
		t.Fatal("ratified GSD state was not adopted")
	}
	if _, err := os.Stat(filepath.Join(repo, "agent-runtime", "shepherd", "canary.txt")); err != nil {
		t.Fatalf("promoted canary artifact missing: %v", err)
	}
	if count := countSQLiteRowsForTest(t, filepath.Join(config.StateDir, "authority.db"), "artifact_proofs"); count != 1 {
		t.Fatalf("artifact proofs=%d want 1", count)
	}
	if count := countSQLiteRowsForTest(t, filepath.Join(config.StateDir, "authority.db"), "attestations"); count != 1 {
		t.Fatalf("attestations=%d want 1", count)
	}
	if validator.calls != 1 {
		t.Fatalf("validator calls=%d want 1", validator.calls)
	}
	if count := countSQLiteQueryForTest(t, filepath.Join(config.StateDir, "authority.db"), "SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'cleanup_complete'"); count != 1 {
		t.Fatalf("cleanup-complete attempt records=%d want 1", count)
	}
}

func setupFakeSuperviseRuntime(t *testing.T) (string, string, fileConfig, *gsd.Runner) {
	t.Helper()
	repo := initializedTestRepository(t)
	branchRaw := runGitForTest(t, repo, "symbolic-ref", "--quiet", "--short", "HEAD")
	contextPath := filepath.Join(repo, "issue-context.json")
	contextRaw := `{
  "issue":389,
  "parent_issue":372,
  "objective":"fake supervise integration",
  "scope":["fake supervise"],
  "non_goals":["live github"],
  "acceptance_criteria":["final human gate"],
  "dependencies":[],
  "write_scope":["agent-runtime/shepherd/**"],
  "required_reading":["AGENTS.md"],
  "required_skills":["golang-how-to"],
  "tdd":{"red":"fake runtime fails before supervise","green":"supervise reaches final gate","refactor":"keep fake bounded"},
  "verification":["go test ./..."],
  "safety":["no secrets"],
  "human_gates":["parent merge"],
  "branch":"` + strings.TrimSpace(branchRaw) + `",
  "pr_base":"main",
  "review_route":"local",
  "sources":["issue #389"]
}`
	if err := os.WriteFile(contextPath, []byte(contextRaw), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, ".gsd"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".gsd", "STATE.md"), []byte("fake canonical state\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitForTest(t, repo, "add", "issue-context.json")
	runGitForTest(t, repo, "commit", "-m", "test: add issue context")
	stateDir := filepath.Join(t.TempDir(), "state")
	gsdHome := filepath.Join(t.TempDir(), "home")
	if err := os.MkdirAll(filepath.Join(gsdHome, "agent"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gsdHome, "agent", "settings.json"), []byte(`{"defaultProvider":"openai-codex","defaultModel":"gpt-5.6-sol","defaultThinkingLevel":"high"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	runner, err := gsd.NewRunner(gsd.Config{
		Command: []string{os.Args[0], "-test.run=TestSuperviseFakeRuntimeHelper", "--"},
		WorkDir: repo, GSDHome: gsdHome, StateDir: stateDir,
		Model: "openai-codex/gpt-5.6-sol", Thinking: "high", Timeout: 10 * time.Second,
		Environment: []string{"GO_WANT_RUNNER_HELPER=supervise"},
	})
	if err != nil {
		t.Fatal(err)
	}
	config := fileConfig{WorkDir: repo, GSDHome: gsdHome, StateDir: stateDir, AttemptRoot: filepath.Join(t.TempDir(), "attempt-worktrees"), CoordinatorModel: "openai-codex/gpt-5.6-sol", ImplementationModel: "openai-codex/gpt-5.5", GSDVersion: "1.11.0", TimeoutSeconds: 10, HeartbeatSeconds: 1, MaxEventBytes: defaultMaxEventBytes, MaxUnitAttempts: 2, Repository: "polymetrics-ai/cli", PullRequest: 391, Runtime: "host"}
	return repo, contextPath, config, runner
}

func validFakeValidationResult(model, verdict string) validation.Result {
	now := time.Now().UTC()
	return validation.Result{ObservedModel: model, Thinking: "high", SessionID: "11111111-1111-1111-1111-111111111111", Verdict: verdict, LocalGates: true, UAT: true, MilestoneValid: true, IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(10 * time.Minute)}
}

type fakeIndependentValidator struct {
	result     validation.Result
	onValidate func(context.Context, validation.Request) error
	calls      int
}

func (f *fakeIndependentValidator) Validate(ctx context.Context, request validation.Request) (validation.Result, error) {
	f.calls++
	if f.onValidate != nil {
		if err := f.onValidate(ctx, request); err != nil {
			return validation.Result{}, err
		}
	}
	result := f.result
	if result.ObservedHead == "" {
		result.ObservedHead = request.CandidateHead
	}
	if result.EvidenceHash == "" {
		result.EvidenceHash = request.EvidenceHash
	}
	return result, nil
}

func installFakeIndependentValidator(t *testing.T, validator validation.Validator) func() {
	t.Helper()
	previous := independentValidatorFactory
	independentValidatorFactory = func(*gsd.Runner, fileConfig) validation.Validator { return validator }
	return func() { independentValidatorFactory = previous }
}

func countSQLiteRowsForTest(t *testing.T, path, table string) int {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func countSQLiteQueryForTest(t *testing.T, path, query string) int {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	var count int
	if err := db.QueryRow(query).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func readFileForTest(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestSuperviseFakeRuntimeHelper(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "supervise" {
		return
	}
	args := helperArgsAfterDoubleDash(os.Args)
	if os.Getenv("RUNNER_HELPER_MODE") == "fail-attempt-query" && len(args) >= 2 && args[0] == "headless" && args[1] == "query" && strings.Contains(os.Getenv("GSD_PROJECT_ROOT"), "attempt-worktrees") {
		fmt.Fprintln(os.Stderr, "bounded fake attempt query failure")
		os.Exit(2)
	}
	if len(args) >= 2 && args[0] == "headless" && args[1] == "query" {
		statePath := filepath.Join(os.Getenv("GSD_STATE_DIR"), "fake-workflow-state")
		if _, err := os.Stat(statePath); err == nil {
			fmt.Print(`{"state":{"activeMilestone":{"id":"M001"},"phase":"complete","nextAction":"stop","blockers":[]},"next":{"action":"stop"}}`)
		} else {
			fmt.Print(`{"state":{"activeMilestone":{"id":"M001"},"activeSlice":{"id":"S01"},"activeTask":{"id":"T01"},"phase":"execution","nextAction":"dispatch","blockers":[]},"next":{"action":"dispatch","unitType":"execute-task","unitId":"M001/S01/T01"}}`)
		}
		os.Exit(0)
	}
	model := ""
	for i, arg := range args {
		if arg == "--model" && i+1 < len(args) {
			model = args[i+1]
		}
	}
	if model != "openai-codex/gpt-5.5" {
		fmt.Fprintf(os.Stderr, "unexpected model %s", model)
		os.Exit(2)
	}
	if os.Getenv("RUNNER_HELPER_MODE") == "fail-runtime" {
		fmt.Println(`{"type":"model_select","model":{"provider":"openai-codex","id":"gpt-5.5"}}`)
		fmt.Println(`{"type":"thinking_level_select","level":"high"}`)
		fmt.Fprintln(os.Stderr, "bounded fake runtime process failure")
		os.Exit(2)
	}
	workDir := os.Getenv("GSD_PROJECT_ROOT")
	artifact := filepath.Join(workDir, "agent-runtime", "shepherd", "canary.txt")
	if err := os.MkdirAll(filepath.Dir(artifact), 0o700); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := os.WriteFile(artifact, []byte("fake supervise artifact\n"), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := os.MkdirAll(os.Getenv("GSD_STATE_DIR"), 0o700); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := os.WriteFile(filepath.Join(os.Getenv("GSD_STATE_DIR"), "fake-workflow-state"), []byte("complete"), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := os.WriteFile(filepath.Join(workDir, ".gsd", "STATE.md"), []byte("fake completed state\n"), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Println(`{"type":"model_select","model":{"provider":"openai-codex","id":"gpt-5.5"}}`)
	fmt.Println(`{"type":"thinking_level_select","level":"high"}`)
	fmt.Println(`{"type":"agent_end","status":"success"}`)
	os.Exit(0)
}

func helperArgsAfterDoubleDash(args []string) []string {
	for i, arg := range args {
		if arg == "--" && i+1 < len(args) {
			return args[i+1:]
		}
	}
	return args
}

func runGitForTest(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}
