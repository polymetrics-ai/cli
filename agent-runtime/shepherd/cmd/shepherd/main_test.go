package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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

func TestFinalUnitRunStateBlocksWhenRetryBudgetExhausted(t *testing.T) {
	t.Parallel()
	result := gsd.Result{Terminal: gsd.TerminalError, Err: errors.New("artifact missing: M001-ROADMAP.md")}
	if got := finalUnitRunState(&result, "executing", 0); got != domain.RunBlocked {
		t.Fatalf("state=%s want blocked", got)
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

func TestLoadConfigRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.json")
	raw := `{"gsd_command":["gsd"],"work_dir":"/tmp/work","gsd_home":"/tmp/home","state_dir":"/tmp/state","unexpected":true}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(path); err == nil {
		t.Fatal("expected unknown config field to fail")
	}
}

func TestLoadConfigRequiresCompleteDecisionPRBinding(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.json")
	raw := `{"gsd_command":["gsd"],"work_dir":"/tmp/work","gsd_home":"/tmp/home","state_dir":"/tmp/state","repository":"polymetrics-ai/cli"}`
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
	raw := `{"gsd_command":["gsd"],"work_dir":"/tmp/work","gsd_home":"/tmp/home","state_dir":"/tmp/state","repository":"polymetrics-ai/cli","pull_request":388}`
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

func TestModelForCommandKeepsImplementationSeparateFromShepherd(t *testing.T) {
	t.Parallel()
	config := fileConfig{CoordinatorModel: "openai-codex/gpt-5.6-sol", ImplementationModel: "openai-codex/gpt-5.5"}
	if got := modelForCommand(config, "execute-task"); got != config.ImplementationModel {
		t.Fatalf("execute-task model=%q", got)
	}
	for _, command := range []string{"research-slice", "plan-slice", "complete-slice", "validate-milestone", "complete-milestone"} {
		if got := modelForCommand(config, command); got != config.CoordinatorModel {
			t.Fatalf("%s model=%q", command, got)
		}
	}
}

func TestLaunchModelForCommandUsesImplementationOnlyForDirectExecution(t *testing.T) {
	t.Parallel()
	config := fileConfig{CoordinatorModel: "openai-codex/gpt-5.6-sol", ImplementationModel: "openai-codex/gpt-5.5"}
	if got := launchModelForCommand(config, "execute-task"); got != config.ImplementationModel {
		t.Fatalf("execute-task launch model=%q, want implementation %q", got, config.ImplementationModel)
	}
	for _, command := range []string{"plan-slice", "validate-milestone"} {
		if got := launchModelForCommand(config, command); got != config.CoordinatorModel {
			t.Fatalf("%s launch model=%q, want coordinator %q", command, got, config.CoordinatorModel)
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
