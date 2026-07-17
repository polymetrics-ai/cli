//go:build integration

package integration_test

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/contract"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/workspace"
	_ "modernc.org/sqlite"
)

type fixture struct {
	moduleRoot      string
	repo            string
	stateDir        string
	gsdHome         string
	attempts        string
	context         string
	config          string
	shepherd        string
	helper          string
	piHelper        string
	successHelper   string
	successPiHelper string
	githubState     string
	binDir          string
	baseHead        string
	baseBranch      string
	baseGSD         string
	baseGSDHash     string
	baseGitStatus   string
}

type commandResult struct {
	stdout []byte
	stderr []byte
	err    error
}

func TestSuperviseFakeRuntimeSuccess(t *testing.T) {
	fixture := newFixture(t, "success")
	result := fixture.supervise(t)
	if result.err != nil {
		t.Fatalf("supervise failed: %v\nstdout:\n%s\nstderr:\n%s", result.err, result.stdout, result.stderr)
	}
	terminal := finalStatus(t, result.stdout)
	if terminal.Status != "final_human_gate" || terminal.NextAction != "stop" {
		t.Fatalf("terminal status=%+v\nstdout:\n%s", terminal, result.stdout)
	}
	if !bytes.Contains(result.stdout, []byte(`"terminal":"success"`)) {
		t.Fatalf("successful execution record missing: %s", result.stdout)
	}
	candidateHead := git(t, fixture.repo, "rev-parse", "HEAD")
	if candidateHead == fixture.baseHead {
		t.Fatal("canonical Git did not promote the candidate")
	}
	if got := readFile(t, filepath.Join(fixture.repo, ".gsd", "STATE.md")); got != "process-boundary completed state\n" {
		t.Fatalf("canonical GSD state=%q", got)
	}
	if got := readFile(t, filepath.Join(fixture.repo, "agent-runtime", "shepherd", "integration-artifact.txt")); got != "process-boundary artifact\n" {
		t.Fatalf("promoted artifact=%q", got)
	}
	authority := filepath.Join(fixture.stateDir, "authority.db")
	assertCount(t, authority, `SELECT COUNT(*) FROM artifact_proofs
		WHERE candidate_head = ? AND validated_head = ? AND validator = 'openai-codex/gpt-5.6-sol'
		AND thinking = 'high' AND ratified = 1`, 1, candidateHead, candidateHead)
	assertCount(t, authority, `SELECT COUNT(*) FROM attestations
		WHERE candidate_head = ? AND observed_head = ? AND head_sha = ?
		AND validator = 'openai-codex/gpt-5.6-sol' AND thinking = 'high' AND verdict = 'PROCEED'`, 1,
		candidateHead, candidateHead, candidateHead)
	assertCount(t, authority, `SELECT COUNT(*) FROM promotion_journals
		WHERE candidate_head = ? AND validated_head = ? AND state = 'complete'
		AND cleanup_complete = 1 AND decisions_resolved = 1`, 1, candidateHead, candidateHead)
	assertCount(t, authority, `SELECT COUNT(*) FROM unit_attempt_identities i
		JOIN artifact_proofs p ON p.delivery_id = i.delivery_id AND p.generation = i.generation
		AND p.unit_id = i.unit_id AND p.attempt = i.attempt AND p.start_head = i.head_sha
		JOIN promotion_journals j ON j.proof_id = p.proof_id AND j.candidate_head = p.candidate_head
		WHERE p.candidate_head = ? AND i.model = 'openai-codex/gpt-5.5' AND i.thinking = 'high'
		AND length(i.session_id) = 36 AND length(i.session_fingerprint) = 71`, 1, candidateHead)
	assertCount(t, authority, `SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'cleanup_complete'`, 1)
	assertCount(t, authority, `SELECT COUNT(*) FROM delivery_runs WHERE state = 'human_gate'`, 1)
	assertProofMatchesCanonical(t, authority, fixture.repo, candidateHead, "execute-task",
		[]string{"execution"}, []string{"gsd_task_complete", "gsd_exec", "gsd_exec_search", "gsd_resume", "gsd_capture_thought"})
	observations := readFile(t, filepath.Join(fixture.stateDir, "runtime", "gsd-state", "integration-observations.jsonl"))
	if !strings.Contains(observations, `"model":"openai-codex/gpt-5.5"`) ||
		!strings.Contains(observations, `"thinking":"high"`) {
		t.Fatalf("implementation routing observation=%q", observations)
	}
	attemptPaths := queryStrings(t, authority, `SELECT path FROM attempt_worktrees`)
	resolvedAttemptRoot, err := filepath.EvalSymlinks(fixture.attempts)
	if err != nil {
		t.Fatal(err)
	}
	if len(attemptPaths) != 1 || !strings.Contains(observations, `"work_dir":"`+attemptPaths[0]+`"`) ||
		!strings.HasPrefix(attemptPaths[0], resolvedAttemptRoot+string(os.PathSeparator)) {
		t.Fatalf("implementation did not run in the registered disposable attempt: paths=%v root=%s observations=%q",
			attemptPaths, resolvedAttemptRoot, observations)
	}
	if strings.Contains(string(result.stdout)+string(result.stderr)+observations, "merge.main") ||
		strings.Contains(string(result.stdout)+string(result.stderr)+observations, "pr.merge") {
		t.Fatal("forbidden merge capability was observed")
	}
}

func TestSuperviseRoutesPlanningMetadataToSolHigh(t *testing.T) {
	fixture := newFixture(t, "planning-route")
	result := fixture.supervise(t)
	if result.err != nil {
		t.Fatalf("planning route failed: %v\nstdout:\n%s\nstderr:\n%s", result.err, result.stdout, result.stderr)
	}
	if terminal := finalStatus(t, result.stdout); terminal.Status != "final_human_gate" {
		t.Fatalf("terminal=%+v", terminal)
	}
	authority := filepath.Join(fixture.stateDir, "authority.db")
	assertCount(t, authority, `SELECT COUNT(*) FROM unit_attempt_identities i
		JOIN artifact_proofs p ON p.delivery_id = i.delivery_id AND p.generation = i.generation
		AND p.unit_id = i.unit_id AND p.attempt = i.attempt AND p.start_head = i.head_sha
		JOIN promotion_journals j ON j.proof_id = p.proof_id
		WHERE i.unit_id = 'plan-milestone/M001' AND i.model = 'openai-codex/gpt-5.6-sol'
		AND i.thinking = 'high' AND length(i.session_id) = 36`, 1)
	assertCount(t, authority, `SELECT COUNT(*) FROM attestations
		WHERE unit_id = 'plan-milestone/M001' AND validator = 'openai-codex/gpt-5.6-sol' AND thinking = 'high'
		AND validator_session_id = '019f7000-2222-7222-8222-222222222222'`, 1)
	candidateHead := git(t, fixture.repo, "rev-parse", "HEAD")
	assertProofMatchesCanonical(t, authority, fixture.repo, candidateHead, "plan-milestone",
		[]string{"planning"}, []string{"gsd_milestone_status", "gsd_plan_milestone", "gsd_plan_slice", "gsd_plan_task"})
}

func TestSuperviseTrackedDeletionReachesFinalGate(t *testing.T) {
	fixture := newFixture(t, "tracked-deletion")
	result := fixture.supervise(t)
	if result.err != nil {
		t.Fatalf("deletion supervise failed: %v\nstdout:\n%s\nstderr:\n%s", result.err, result.stdout, result.stderr)
	}
	if terminal := finalStatus(t, result.stdout); terminal.Status != "final_human_gate" {
		t.Fatalf("terminal=%+v", terminal)
	}
	candidateHead := git(t, fixture.repo, "rev-parse", "HEAD")
	if candidateHead == fixture.baseHead {
		t.Fatal("canonical Git did not promote deletion candidate")
	}
	if _, err := os.Lstat(filepath.Join(fixture.repo, "agent-runtime", "shepherd", "integration-artifact.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("canonical deleted content still exists: %v", err)
	}
	authority := filepath.Join(fixture.stateDir, "authority.db")
	assertCount(t, authority, `SELECT COUNT(*) FROM artifact_proofs
		WHERE candidate_head = ? AND validated_head = ? AND expected_artifact LIKE '%"deleted":true%'`,
		1, candidateHead, candidateHead)
	assertProofMatchesCanonical(t, authority, fixture.repo, candidateHead, "execute-task",
		[]string{"execution"}, []string{"gsd_task_complete", "gsd_exec", "gsd_exec_search", "gsd_resume", "gsd_capture_thought"})
}

func TestSuperviseTrackedDeletionRejectsRecreatedPathAfterValidation(t *testing.T) {
	fixture := newFixture(t, "tracked-deletion-recreated")
	result := fixture.supervise(t)
	if result.err == nil {
		t.Fatalf("recreated deletion unexpectedly promoted:\n%s", result.stdout)
	}
	fixture.assertCanonicalUnchanged(t)
	assertCount(t, filepath.Join(fixture.stateDir, "authority.db"), `SELECT COUNT(*) FROM promotion_journals`, 0)
}

func TestSuperviseAcceptsValidatedGSDStateOnlyCandidate(t *testing.T) {
	fixture := newFixture(t, "gsd-state-only")
	result := fixture.supervise(t)
	if result.err != nil {
		t.Fatalf("GSD-state-only candidate failed: %v\nstdout:\n%s\nstderr:\n%s",
			result.err, result.stdout, result.stderr)
	}
	if terminal := finalStatus(t, result.stdout); terminal.Status != "final_human_gate" {
		t.Fatalf("terminal=%+v", terminal)
	}
	candidateHead := git(t, fixture.repo, "rev-parse", "HEAD")
	if candidateHead == fixture.baseHead {
		t.Fatal("empty Git checkpoint did not advance")
	}
	if got := readFile(t, filepath.Join(fixture.repo, ".gsd", "STATE.md")); got != "process-boundary completed state\n" {
		t.Fatalf("canonical GSD state=%q", got)
	}
	if _, err := os.Stat(filepath.Join(fixture.repo, "agent-runtime", "shepherd", "integration-artifact.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected non-GSD artifact: %v", err)
	}
	authority := filepath.Join(fixture.stateDir, "authority.db")
	assertCount(t, authority, `SELECT COUNT(*) FROM artifact_proofs
		WHERE candidate_head = ? AND validated_head = ? AND expected_artifact LIKE '%.gsd/STATE.md%'`,
		1, candidateHead, candidateHead)
	assertProofMatchesCanonical(t, authority, fixture.repo, candidateHead, "execute-task",
		[]string{"execution"}, []string{"gsd_task_complete", "gsd_exec", "gsd_exec_search", "gsd_resume", "gsd_capture_thought"})
}

func TestSuperviseFakeRuntimeRejectsInvalidValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		scenario         string
		failureClass     string
		failureState     string
		failureCount     int
		expectedRunState string
	}{
		{name: "candidate does not advance", scenario: "no-candidate-change", failureClass: "artifact_missing", failureState: "cleanup_complete", failureCount: 2, expectedRunState: "awaiting_decision"},
		{name: "missing candidate artifact", scenario: "missing-candidate-artifact", failureClass: "artifact_missing", failureState: "cleanup_complete", failureCount: 2, expectedRunState: "awaiting_decision"},
		{name: "wrong implementation session model", scenario: "wrong-implementation-session-model", failureClass: "unknown", expectedRunState: "awaiting_decision"},
		{name: "wrong implementation session thinking", scenario: "wrong-implementation-session-thinking", failureClass: "unknown", expectedRunState: "awaiting_decision"},
		{name: "implementation missing turn lifecycle", scenario: "implementation-missing-turn", failureClass: "unknown", expectedRunState: "awaiting_decision"},
		{name: "implementation errored final turn", scenario: "implementation-error-stop", failureClass: "unknown", expectedRunState: "awaiting_decision"},
		{name: "implementation retrying agent end", scenario: "implementation-retrying-end", failureClass: "unknown", expectedRunState: "awaiting_decision"},
		{name: "implementation errored agent end", scenario: "implementation-agent-error", failureClass: "unknown", expectedRunState: "awaiting_decision"},
		{name: "direct GSD write without workflow transition", scenario: "implementation-no-workflow-tool", failureClass: "artifact_invalid", expectedRunState: "awaiting_decision"},
		{name: "incomplete required workflow transitions", scenario: "implementation-partial-workflow", failureClass: "artifact_invalid", expectedRunState: "awaiting_decision"},
		{name: "missing evidence", scenario: "missing-validator-evidence", failureClass: "validation_failed"},
		{name: "GPT-5.5 validator", scenario: "gpt55-validator", failureClass: "validation_failed"},
		{name: "non-high thinking", scenario: "non-high-validator", failureClass: "validation_failed"},
		{name: "validator errored agent end", scenario: "validator-agent-error", failureClass: "validation_failed"},
		{name: "stale candidate head", scenario: "stale-validator-head", failureClass: "validation_failed"},
		{name: "candidate changes during validation", scenario: "candidate-moved", failureClass: "validation_failed"},
		{name: "artifact changes during validation", scenario: "artifact-changed", failureClass: "validation_failed"},
		{name: "RETRY verdict", scenario: "validator-retry", failureClass: "ratification_failed"},
		{name: "HALT verdict", scenario: "validator-halt", failureClass: "ratification_failed"},
		{name: "missing required gate", scenario: "missing-required-gate", failureClass: "validation_failed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newFixture(t, test.scenario)
			result := fixture.supervise(t)
			if result.err == nil {
				t.Fatalf("invalid validator unexpectedly succeeded\nstdout:\n%s\nstderr:\n%s", result.stdout, result.stderr)
			}
			fixture.assertCanonicalUnchanged(t)
			authority := filepath.Join(fixture.stateDir, "authority.db")
			assertCount(t, authority, `SELECT COUNT(*) FROM artifact_proofs`, 0)
			assertCount(t, authority, `SELECT COUNT(*) FROM attestations`, 0)
			assertCount(t, authority, `SELECT COUNT(*) FROM promotion_journals`, 0)
			failureState := test.failureState
			if failureState == "" {
				failureState = "retained_for_recovery"
			}
			failureCount := test.failureCount
			if failureCount == 0 {
				failureCount = 1
			}
			assertCount(t, authority, `SELECT COUNT(*) FROM attempt_worktrees
				WHERE state = ? AND failure_class = ?`, failureCount, failureState, test.failureClass)
			runState := test.expectedRunState
			if runState == "" {
				runState = "blocked"
			}
			assertCount(t, authority, `SELECT COUNT(*) FROM delivery_runs WHERE state = ?`, 1, runState)
			effects := 0
			if runState == "awaiting_decision" {
				effects = 1
			}
			assertCount(t, filepath.Join(fixture.stateDir, "outbox.db"),
				`SELECT COUNT(*) FROM effects`, effects)
		})
	}
}

func TestSuperviseRejectsPostValidationGSDMutation(t *testing.T) {
	fixture := newFixture(t, "post-validation-mutation")
	result := fixture.supervise(t)
	if result.err == nil {
		t.Fatalf("post-validation mutation unexpectedly promoted: %s", result.stdout)
	}
	fixture.assertCanonicalUnchanged(t)
	authority := filepath.Join(fixture.stateDir, "authority.db")
	assertCount(t, authority, `SELECT COUNT(*) FROM artifact_proofs WHERE ratified = 1`, 1)
	assertCount(t, authority, `SELECT COUNT(*) FROM promotion_journals
		WHERE state = 'journal_created' AND manifest_hash = ''`, 1)
	restarted := fixture.supervise(t)
	if restarted.err == nil {
		t.Fatalf("restart promoted mutated GSD state: %s", restarted.stdout)
	}
	assertCount(t, authority, `SELECT COUNT(*) FROM promotion_journals
		WHERE state = 'journal_created' AND manifest_hash = ''`, 1)
	requestID, commentID := decisionIdentity(t, authority)
	appendHumanReply(t, fixture.githubState, commentID+1, "/shepherd decide "+requestID+" retry")
	approved := fixture.supervise(t)
	if approved.err == nil {
		t.Fatalf("approved retry promoted mutated GSD state: %s", approved.stdout)
	}
	assertCount(t, authority, `SELECT COUNT(*) FROM promotion_journals
		WHERE state = 'blocked' AND blocked_reason = 'staged_state_differs_from_validated_manifest'`, 1)
	fixture.assertCanonicalUnchanged(t)
}

func TestSuperviseFakeRuntimeOfficialRegistryAdmission(t *testing.T) {
	tests := []struct {
		name     string
		scenario string
		mutate   func(*testing.T, fixture)
	}{
		{name: "unknown unit fails closed", scenario: "unknown-unit"},
		{name: "stale registry version fails closed", scenario: "success", mutate: func(t *testing.T, f fixture) {
			updateConfig(t, f.config, func(config map[string]any) { config["gsd_version"] = "1.10.0" })
		}},
		{name: "partial runtime metadata fails closed", scenario: "success", mutate: func(t *testing.T, f fixture) {
			node, err := exec.LookPath("node")
			if err != nil {
				t.Fatal(err)
			}
			updateConfig(t, f.config, func(config map[string]any) {
				config["gsd_command"] = []string{node, f.helper}
			})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newFixture(t, test.scenario)
			if test.mutate != nil {
				test.mutate(t, fixture)
			}
			result := fixture.supervise(t)
			if result.err == nil {
				t.Fatalf("invalid registry/runtime unexpectedly succeeded: %s", result.stdout)
			}
			fixture.assertCanonicalUnchanged(t)
			if !strings.Contains(string(result.stderr), "runtime_contract_mismatch") &&
				!strings.Contains(string(result.stderr), "unknown unit") {
				t.Fatalf("registry rejection was not typed: %s", result.stderr)
			}
		})
	}
}

func TestSuperviseFakeRuntimeRecoveryUsesFreshAttempt(t *testing.T) {
	fixture := newFixture(t, "recoverable-failure")
	result := fixture.supervise(t)
	if result.err != nil {
		t.Fatalf("recoverable supervise failed: %v\nstdout:\n%s\nstderr:\n%s", result.err,
			result.stdout, result.stderr)
	}
	if terminal := finalStatus(t, result.stdout); terminal.Status != "final_human_gate" {
		t.Fatalf("recovered terminal=%+v", terminal)
	}
	authority := filepath.Join(fixture.stateDir, "authority.db")
	assertCount(t, authority, `SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'cleanup_complete'
		AND failure_class = 'dead_worker'`, 1)
	assertCount(t, authority, `SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'cleanup_complete'`, 2)
	assertCount(t, authority, `SELECT COUNT(DISTINCT path) FROM attempt_worktrees`, 2)
	assertCount(t, authority, `SELECT COUNT(*) FROM recovery_attempts WHERE failure_class = 'dead_worker'
		AND status = 'consumed'`, 1)
	assertCount(t, authority, `SELECT COUNT(*) FROM recovery_budgets WHERE failure_class = 'dead_worker'
		AND attempts = 1`, 1)
}

func TestSuperviseFreshAttemptAfterValidationTimeoutUsesFreshProof(t *testing.T) {
	fixture := newFixture(t, "recoverable-validation-failure")
	updateConfig(t, fixture.config, func(config map[string]any) { config["timeout_seconds"] = 2 })
	crashed := fixture.superviseAtBoundary(t, "retry_persisted")
	var exitErr *exec.ExitError
	if !errors.As(crashed.err, &exitErr) || exitErr.ExitCode() != 96 {
		t.Fatalf("persisted-retry boundary did not hard-exit: %v\n%s\n%s", crashed.err,
			crashed.stdout, crashed.stderr)
	}
	authority := filepath.Join(fixture.stateDir, "authority.db")
	assertCount(t, authority, `SELECT COUNT(*) FROM attempt_worktrees
		WHERE state = 'retained_for_recovery' AND failure_class = 'validation_failed'
		AND candidate_head <> ''`, 1)
	result := fixture.supervise(t)
	if result.err != nil {
		t.Fatalf("fresh validation attempt failed after restart: %v\nstdout:\n%s\nstderr:\n%s",
			result.err, result.stdout, result.stderr)
	}
	if terminal := finalStatus(t, result.stdout); terminal.Status != "final_human_gate" {
		t.Fatalf("terminal=%+v", terminal)
	}
	assertCount(t, authority, `SELECT COUNT(*) FROM attempt_worktrees
		WHERE state = 'cleanup_complete' AND failure_class = 'validation_failed'`, 1)
	assertCount(t, authority, `SELECT COUNT(DISTINCT path) FROM attempt_worktrees`, 2)
	assertCount(t, authority, `SELECT COUNT(DISTINCT candidate_head) FROM attempt_worktrees
		WHERE candidate_head <> ''`, 2)
	assertCount(t, authority, `SELECT COUNT(DISTINCT session_id) FROM unit_attempt_identities`, 2)
	assertCount(t, authority, `SELECT COUNT(*) FROM artifact_proofs
		WHERE candidate_head = (SELECT candidate_head FROM attempt_worktrees WHERE state = 'cleanup_complete' AND failure_class = '')
		AND validator = 'openai-codex/gpt-5.6-sol' AND thinking = 'high'`, 1)
	assertCount(t, authority, `SELECT COUNT(*) FROM artifact_proofs
		WHERE candidate_head = (SELECT candidate_head FROM attempt_worktrees WHERE failure_class = 'validation_failed')`, 0)
}

func TestSuperviseRestartAwaitingDecisionConsumesBoundReplyOnce(t *testing.T) {
	t.Parallel()
	fixture := newFixture(t, "always-fail")
	first := fixture.supervise(t)
	if first.err == nil {
		t.Fatalf("exhausted runtime unexpectedly succeeded\nstdout:\n%s\nstderr:\n%s", first.stdout, first.stderr)
	}
	authority := filepath.Join(fixture.stateDir, "authority.db")
	assertCount(t, authority, `SELECT COUNT(*) FROM decision_requests WHERE status = 'published'`, 1)
	assertCount(t, authority, `SELECT COUNT(*) FROM delivery_runs WHERE state = 'awaiting_decision'`, 1)
	assertCount(t, filepath.Join(fixture.stateDir, "outbox.db"),
		`SELECT COUNT(*) FROM effects WHERE kind = 'github.comment.decision_question.v1' AND state = 'sent'`, 1)
	if got := botCommentCount(t, fixture.githubState); got != 1 {
		t.Fatalf("question comments=%d want 1", got)
	}
	assertQuestionBody(t, authority, fixture.githubState)
	withoutReply := fixture.supervise(t)
	if withoutReply.err == nil {
		t.Fatalf("restart without reply unexpectedly advanced: %s", withoutReply.stdout)
	}
	if got := botCommentCount(t, fixture.githubState); got != 1 {
		t.Fatalf("restart duplicated question: comments=%d", got)
	}
	requestID, questionID := decisionIdentity(t, authority)
	appendStoredReply(t, fixture.githubState, questionID+1, "/shepherd decide "+requestID+" retry",
		"mallory", 1, "User", false)
	appendStoredReply(t, fixture.githubState, questionID+2, "/shepherd decide "+requestID+" retry",
		"karthik-sivadas", 6113982, "User", true)
	appendHumanReply(t, fixture.githubState, questionID+3, "/shepherd decide stale-request retry")
	stale := fixture.supervise(t)
	if stale.err == nil {
		t.Fatalf("unauthorized, edited, or stale reply unexpectedly advanced: %s", stale.stdout)
	}
	assertCount(t, authority, `SELECT COUNT(*) FROM decision_requests WHERE status = 'published'`, 1)
	fixture.assertCanonicalUnchanged(t)
	appendHumanReply(t, fixture.githubState, questionID+4, "/shepherd decide "+requestID+" retry")
	appendHumanReply(t, fixture.githubState, questionID+5, "/shepherd decide "+requestID+" stop")
	ambiguous := fixture.supervise(t)
	if ambiguous.err == nil {
		t.Fatalf("duplicate valid replies unexpectedly advanced: %s", ambiguous.stdout)
	}
	assertCount(t, authority, `SELECT COUNT(*) FROM decision_requests WHERE status = 'published'`, 1)
	comments := readStoredComments(t, fixture.githubState)
	writeJSON(t, fixture.githubState, comments[:len(comments)-1])
	fixture.useSuccessHelper(t)
	resumed := fixture.supervise(t)
	if resumed.err != nil {
		t.Fatalf("bound reply resume failed: %v\nstdout:\n%s\nstderr:\n%s", resumed.err,
			resumed.stdout, resumed.stderr)
	}
	if terminal := finalStatus(t, resumed.stdout); terminal.Status != "final_human_gate" {
		t.Fatalf("resumed terminal=%+v", terminal)
	}
	assertCount(t, authority, `SELECT COUNT(*) FROM decision_requests WHERE status = 'consumed'
		AND accepted_actor_id = 6113982 AND accepted_comment_id > github_comment_id`, 1)
	assertCount(t, authority, `SELECT COUNT(*) FROM delivery_runs WHERE state = 'human_gate'`, 1)
	assertCount(t, filepath.Join(fixture.stateDir, "outbox.db"),
		`SELECT COUNT(*) FROM effects WHERE kind = 'github.comment.decision_question.v1'`, 1)
	if got := botCommentCount(t, fixture.githubState); got != 1 {
		t.Fatalf("resume duplicated question: bot comments=%d", got)
	}
}

func TestSuperviseWaitsForDecisionAndConsumesReplyInSameProcess(t *testing.T) {
	t.Parallel()
	fixture := newFixture(t, "decision-resume")
	cmd, stdout, stderr := fixture.superviseCommand(t, "")
	cmd.Env = withoutEnvironment(cmd.Env, "SHEPHERD_INTEGRATION_EXIT_AWAITING_DECISION")
	resumeFile := filepath.Join(fixture.stateDir, "runtime", "gsd-state", "integration-decision-resume")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	authority := filepath.Join(fixture.stateDir, "authority.db")
	waitForSQLiteCount(t, authority, `SELECT COUNT(*) FROM decision_requests WHERE status = 'published'`, 1)
	waitForSQLiteCount(t, authority, `SELECT COUNT(*) FROM delivery_runs WHERE state = 'awaiting_decision'`, 1)
	time.Sleep(1200 * time.Millisecond)
	requestID, questionID := decisionIdentity(t, authority)
	if err := os.WriteFile(resumeFile, []byte("resume\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	appendHumanReply(t, fixture.githubState, questionID+1, "/shepherd decide "+requestID+" retry")
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("same-process decision resume failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
		}
	case <-time.After(60 * time.Second):
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("same-process decision resume did not finish\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if terminal := finalStatus(t, stdout.Bytes()); terminal.Status != "final_human_gate" {
		t.Fatalf("terminal=%+v", terminal)
	}
	if heartbeats := bytes.Count(stdout.Bytes(), []byte(`"status":"blocked"`)); heartbeats < 2 {
		t.Fatalf("awaiting-decision heartbeats=%d want at least 2; stdout=%s", heartbeats, stdout)
	}
	assertCount(t, authority, `SELECT COUNT(*) FROM decision_requests WHERE status = 'consumed'`, 1)
	assertCount(t, authority, `SELECT COUNT(*) FROM delivery_runs WHERE state = 'human_gate'`, 1)
	if got := botCommentCount(t, fixture.githubState); got != 1 {
		t.Fatalf("same-process resume duplicated question: comments=%d", got)
	}
}

func TestSuperviseExpiresUnansweredDecisionInSameProcess(t *testing.T) {
	t.Parallel()
	fixture := newFixture(t, "always-fail")
	cmd, stdout, stderr := fixture.superviseCommand(t, "")
	cmd.Env = withoutEnvironment(cmd.Env, "SHEPHERD_INTEGRATION_EXIT_AWAITING_DECISION")
	cmd.Env = append(cmd.Env, "SHEPHERD_INTEGRATION_SHORT_DECISION_TTL=1")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 10 {
			t.Fatalf("expired decision exit=%v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
		}
	case <-time.After(45 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("same-process decision expiry did not terminate")
	}
	authority := filepath.Join(fixture.stateDir, "authority.db")
	statuses := queryStrings(t, authority, `SELECT status FROM decision_requests ORDER BY request_id`)
	if len(statuses) != 1 || statuses[0] != "expired" {
		t.Fatalf("decision statuses=%v want [expired]\nstdout:\n%s\nstderr:\n%s", statuses, stdout, stderr)
	}
	assertCount(t, authority, `SELECT COUNT(*) FROM delivery_runs WHERE state = 'blocked'`, 1)
	fixture.assertCanonicalUnchanged(t)
}

func TestSuperviseRestartOutboxDispatchesPendingEffectOnce(t *testing.T) {
	fixture := newFixture(t, "always-fail")
	crashed := fixture.superviseWithEnv(t, map[string]string{
		"SHEPHERD_INTEGRATION_OUTBOX_CRASH": "post_enqueue",
	})
	var exitErr *exec.ExitError
	if !errors.As(crashed.err, &exitErr) || exitErr.ExitCode() != 99 {
		t.Fatalf("outbox post-enqueue boundary did not hard-exit: %v", crashed.err)
	}
	outboxPath := filepath.Join(fixture.stateDir, "outbox.db")
	assertCount(t, outboxPath, `SELECT COUNT(*) FROM effects WHERE state = 'pending'`, 1)
	if _, err := os.Stat(fixture.githubState); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("pending effect executed before restart: %v", err)
	}
	restarted := fixture.supervise(t)
	if restarted.err == nil {
		t.Fatalf("awaiting-decision restart unexpectedly succeeded: %s", restarted.stdout)
	}
	assertCount(t, outboxPath, `SELECT COUNT(*) FROM effects WHERE state = 'sent'`, 1)
	if got := botCommentCount(t, fixture.githubState); got != 1 {
		t.Fatalf("pending recovery comments=%d want 1", got)
	}
	if again := fixture.supervise(t); again.err == nil {
		t.Fatalf("second awaiting-decision restart unexpectedly succeeded: %s", again.stdout)
	}
	if got := botCommentCount(t, fixture.githubState); got != 1 {
		t.Fatalf("second pending recovery duplicated comment: comments=%d", got)
	}
	fixture.assertCanonicalUnchanged(t)
}

func TestSuperviseRestartOutboxRecoversPreExecutionClaimOnce(t *testing.T) {
	fixture := newFixture(t, "always-fail")
	crashed := fixture.superviseWithEnv(t, map[string]string{
		"SHEPHERD_INTEGRATION_OUTBOX_CRASH":     "post_claim",
		"SHEPHERD_INTEGRATION_SHORT_EFFECT_TTL": "1",
	})
	var exitErr *exec.ExitError
	if !errors.As(crashed.err, &exitErr) || exitErr.ExitCode() != 97 {
		t.Fatalf("outbox post-claim boundary did not hard-exit: %v", crashed.err)
	}
	outboxPath := filepath.Join(fixture.stateDir, "outbox.db")
	assertCount(t, outboxPath, `SELECT COUNT(*) FROM effects e JOIN claims c ON c.effect_id = e.effect_id
		WHERE e.state = 'claimed' AND c.closed_at = 0 AND c.execution_started_at = 0`, 1)
	waitForClaimExpiry(t, outboxPath)
	restarted := fixture.superviseWithEnv(t, map[string]string{
		"SHEPHERD_INTEGRATION_SHORT_EFFECT_TTL": "1",
	})
	if restarted.err == nil {
		t.Fatalf("awaiting-decision restart unexpectedly succeeded: %s", restarted.stdout)
	}
	assertCount(t, outboxPath, `SELECT COUNT(*) FROM effects WHERE state = 'sent'`, 1)
	if got := botCommentCount(t, fixture.githubState); got != 1 {
		t.Fatalf("pre-execution claim recovery comments=%d want 1", got)
	}
	fixture.assertCanonicalUnchanged(t)
}

func TestSuperviseRestartOutboxDoesNotSendExecutionStartedClaim(t *testing.T) {
	fixture := newFixture(t, "always-fail")
	crashed := fixture.superviseWithEnv(t, map[string]string{
		"SHEPHERD_INTEGRATION_OUTBOX_CRASH":     "post_execution_start",
		"SHEPHERD_INTEGRATION_SHORT_EFFECT_TTL": "1",
	})
	var exitErr *exec.ExitError
	if !errors.As(crashed.err, &exitErr) || exitErr.ExitCode() != 96 {
		t.Fatalf("outbox post-execution-start boundary did not hard-exit: %v", crashed.err)
	}
	outboxPath := filepath.Join(fixture.stateDir, "outbox.db")
	assertCount(t, outboxPath, `SELECT COUNT(*) FROM effects e JOIN claims c ON c.effect_id = e.effect_id
		WHERE e.state = 'claimed' AND c.closed_at = 0 AND c.execution_started_at > 0`, 1)
	waitForClaimExpiry(t, outboxPath)
	restarted := fixture.superviseWithEnv(t, map[string]string{
		"SHEPHERD_INTEGRATION_SHORT_EFFECT_TTL": "1",
	})
	if restarted.err == nil {
		t.Fatalf("execution-started uncertainty unexpectedly succeeded: %s", restarted.stdout)
	}
	assertCount(t, outboxPath, `SELECT COUNT(*) FROM effects WHERE state = 'uncertain'
		AND error_code = 'claim_expired_after_execution'`, 1)
	if _, err := os.Stat(fixture.githubState); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("execution-started recovery created GitHub state: %v", err)
	}
	fixture.assertCanonicalUnchanged(t)
}

func TestSuperviseRestartOutboxReconcilesPostSendCrash(t *testing.T) {
	fixture := newFixture(t, "always-fail")
	crashed := fixture.superviseWithEnv(t, map[string]string{
		"SHEPHERD_INTEGRATION_OUTBOX_CRASH":     "post_send",
		"SHEPHERD_INTEGRATION_SHORT_EFFECT_TTL": "1",
	})
	var exitErr *exec.ExitError
	if !errors.As(crashed.err, &exitErr) || exitErr.ExitCode() != 98 {
		t.Fatalf("outbox post-send boundary did not hard-exit: err=%v\nstdout:\n%s\nstderr:\n%s",
			crashed.err, crashed.stdout, crashed.stderr)
	}
	if got := botCommentCount(t, fixture.githubState); got != 1 {
		t.Fatalf("post-send crash comments=%d want 1", got)
	}
	outboxPath := filepath.Join(fixture.stateDir, "outbox.db")
	assertCount(t, outboxPath, `SELECT COUNT(*) FROM effects WHERE state = 'claimed'`, 1)
	waitForClaimExpiry(t, outboxPath)
	restarted := fixture.superviseWithEnv(t, map[string]string{
		"SHEPHERD_INTEGRATION_SHORT_EFFECT_TTL": "1",
	})
	if restarted.err == nil {
		t.Fatalf("awaiting-decision restart unexpectedly succeeded: %s", restarted.stdout)
	}
	if got := botCommentCount(t, fixture.githubState); got != 1 {
		t.Fatalf("reconciliation duplicated comment: comments=%d", got)
	}
	assertCount(t, outboxPath, `SELECT COUNT(*) FROM effects WHERE state = 'sent'
		AND error_code = ''`, 1)
	assertCount(t, outboxPath, `SELECT COUNT(*) FROM results WHERE code = 'reconciled'`, 1)
	assertCount(t, filepath.Join(fixture.stateDir, "authority.db"),
		`SELECT COUNT(*) FROM decision_requests WHERE status = 'published' AND github_comment_id > 0`, 1)
	again := fixture.superviseWithEnv(t, map[string]string{
		"SHEPHERD_INTEGRATION_SHORT_EFFECT_TTL": "1",
	})
	if again.err == nil {
		t.Fatalf("second awaiting-decision restart unexpectedly succeeded: %s", again.stdout)
	}
	if got := botCommentCount(t, fixture.githubState); got != 1 {
		t.Fatalf("second restart duplicated comment: comments=%d", got)
	}
	fixture.assertCanonicalUnchanged(t)
}

func TestSuperviseRestartOutboxDuplicateMarkerCollisionBlocks(t *testing.T) {
	fixture := newFixture(t, "always-fail")
	crashed := fixture.superviseWithEnv(t, map[string]string{
		"SHEPHERD_INTEGRATION_OUTBOX_CRASH":     "post_send",
		"SHEPHERD_INTEGRATION_SHORT_EFFECT_TTL": "1",
	})
	var exitErr *exec.ExitError
	if !errors.As(crashed.err, &exitErr) || exitErr.ExitCode() != 98 {
		t.Fatalf("outbox post-send boundary did not hard-exit: %v", crashed.err)
	}
	comments := readStoredComments(t, fixture.githubState)
	if len(comments) != 1 {
		t.Fatalf("initial marker comments=%d want 1", len(comments))
	}
	duplicate := comments[0]
	duplicate.ID++
	comments = append(comments, duplicate)
	writeJSON(t, fixture.githubState, comments)
	outboxPath := filepath.Join(fixture.stateDir, "outbox.db")
	waitForClaimExpiry(t, outboxPath)
	restarted := fixture.superviseWithEnv(t, map[string]string{
		"SHEPHERD_INTEGRATION_SHORT_EFFECT_TTL": "1",
	})
	if restarted.err == nil {
		t.Fatalf("duplicate marker collision unexpectedly reconciled: %s", restarted.stdout)
	}
	assertCount(t, outboxPath, `SELECT COUNT(*) FROM effects WHERE state = 'uncertain'`, 1)
	if got := botCommentCount(t, fixture.githubState); got != 2 {
		t.Fatalf("collision recovery wrote a new comment: comments=%d", got)
	}
	fixture.assertCanonicalUnchanged(t)
}

func TestSuperviseRestartOutboxAmbiguityDoesNotBlindlyReplay(t *testing.T) {
	fixture := newFixture(t, "always-fail")
	crashed := fixture.superviseWithEnv(t, map[string]string{
		"SHEPHERD_INTEGRATION_OUTBOX_CRASH":     "post_send",
		"SHEPHERD_INTEGRATION_SHORT_EFFECT_TTL": "1",
	})
	var exitErr *exec.ExitError
	if !errors.As(crashed.err, &exitErr) || exitErr.ExitCode() != 98 {
		t.Fatalf("outbox post-send boundary did not hard-exit: %v", crashed.err)
	}
	writeJSON(t, fixture.githubState, []storedGitHubComment{})
	outboxPath := filepath.Join(fixture.stateDir, "outbox.db")
	waitForClaimExpiry(t, outboxPath)
	restarted := fixture.superviseWithEnv(t, map[string]string{
		"SHEPHERD_INTEGRATION_SHORT_EFFECT_TTL": "1",
	})
	if restarted.err == nil {
		t.Fatalf("markerless ambiguous effect unexpectedly succeeded: %s", restarted.stdout)
	}
	assertCount(t, outboxPath, `SELECT COUNT(*) FROM effects WHERE state = 'uncertain'
		AND error_code = 'claim_expired_after_execution'`, 1)
	if got := botCommentCount(t, fixture.githubState); got != 0 {
		t.Fatalf("ambiguous recovery replayed write: comments=%d", got)
	}
	again := fixture.superviseWithEnv(t, map[string]string{
		"SHEPHERD_INTEGRATION_SHORT_EFFECT_TTL": "1",
	})
	if again.err == nil {
		t.Fatalf("second markerless recovery unexpectedly succeeded: %s", again.stdout)
	}
	if got := botCommentCount(t, fixture.githubState); got != 0 {
		t.Fatalf("second ambiguous recovery replayed write: comments=%d", got)
	}
	fixture.assertCanonicalUnchanged(t)
}

func TestSuperviseRestartLegacyProofCompatibilityIsPostGitOnly(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name        string
		boundary    string
		wantSuccess bool
	}{
		{name: "pre Git blocks", boundary: "before_git_promotion"},
		{name: "post Git completes forward", boundary: "after_git_promotion", wantSuccess: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newFixture(t, "success")
			crashed := fixture.superviseAtBoundary(t, test.boundary)
			var exitErr *exec.ExitError
			if !errors.As(crashed.err, &exitErr) || exitErr.ExitCode() != 97 {
				t.Fatalf("boundary %s did not hard-exit: %v", test.boundary, crashed.err)
			}
			downgradeProofDatabaseToLegacy(t, filepath.Join(fixture.stateDir, "authority.db"))
			restarted := fixture.supervise(t)
			if test.wantSuccess {
				if restarted.err != nil || finalStatus(t, restarted.stdout).Status != "final_human_gate" {
					t.Fatalf("legacy post-Git recovery failed: %v\n%s\n%s", restarted.err,
						restarted.stdout, restarted.stderr)
				}
				return
			}
			if restarted.err == nil {
				t.Fatalf("legacy pre-Git proof promoted: %s", restarted.stdout)
			}
			fixture.assertCanonicalUnchanged(t)
			assertCount(t, filepath.Join(fixture.stateDir, "authority.db"), `SELECT COUNT(*)
				FROM promotion_journals WHERE state = 'blocked'
				AND blocked_reason = 'legacy_pre_git_manifest_requires_human_reconciliation'`, 1)
		})
	}
}

func TestSuperviseRestartRejectsCanonicalGSDDriftAfterFinalGate(t *testing.T) {
	fixture := newFixture(t, "success")
	first := fixture.supervise(t)
	if first.err != nil {
		t.Fatalf("initial supervise failed: %v", first.err)
	}
	head := git(t, fixture.repo, "rev-parse", "HEAD")
	writeFile(t, filepath.Join(fixture.repo, ".gsd", "unratified-state.txt"), "drift\n", 0o600)
	restarted := fixture.supervise(t)
	if restarted.err == nil {
		t.Fatalf("canonical GSD drift re-emitted final gate: %s", restarted.stdout)
	}
	if strings.Contains(string(restarted.stdout), `"status":"final_human_gate"`) {
		t.Fatalf("drift emitted final gate status: %s", restarted.stdout)
	}
	if got := git(t, fixture.repo, "rev-parse", "HEAD"); got != head {
		t.Fatalf("drift check changed Git head: %s != %s", got, head)
	}
	assertCount(t, filepath.Join(fixture.stateDir, "authority.db"),
		`SELECT COUNT(*) FROM delivery_runs WHERE state = 'human_gate'`, 1)
}

func TestSuperviseRestartPromotionBoundaries(t *testing.T) {
	t.Parallel()
	boundaries := []string{
		"after_journal_created", "after_state_staged", "before_git_promotion",
		"after_git_promotion", "after_git_promoted_journaled", "before_backup_rename",
		"after_backup_rename", "after_state_install", "after_state_installed_journaled",
		"after_journal_complete", "final_gate_projected",
	}
	for _, boundary := range boundaries {
		t.Run(boundary, func(t *testing.T) {
			fixture := newFixture(t, "success")
			crashed := fixture.superviseAtBoundary(t, boundary)
			var exitErr *exec.ExitError
			if !errors.As(crashed.err, &exitErr) || exitErr.ExitCode() != 97 {
				t.Fatalf("boundary %s did not hard-exit at durable seam: err=%v\nstdout:\n%s\nstderr:\n%s",
					boundary, crashed.err, crashed.stdout, crashed.stderr)
			}
			restarted := fixture.supervise(t)
			if restarted.err != nil {
				t.Fatalf("boundary %s restart failed: %v\nstdout:\n%s\nstderr:\n%s",
					boundary, restarted.err, restarted.stdout, restarted.stderr)
			}
			terminal := finalStatus(t, restarted.stdout)
			if terminal.Status != "final_human_gate" {
				t.Fatalf("boundary %s terminal=%+v", boundary, terminal)
			}
			authority := filepath.Join(fixture.stateDir, "authority.db")
			assertCount(t, authority, `SELECT COUNT(*) FROM promotion_journals
				WHERE state = 'complete' AND cleanup_complete = 1 AND decisions_resolved = 1`, 1)
			assertCount(t, authority, `SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'cleanup_complete'`, 1)
			if got := readFile(t, filepath.Join(fixture.repo, ".gsd", "STATE.md")); got != "process-boundary completed state\n" {
				t.Fatalf("boundary %s left mixed GSD state: %q", boundary, got)
			}
			if git(t, fixture.repo, "rev-parse", "HEAD") == fixture.baseHead {
				t.Fatalf("boundary %s did not converge Git forward", boundary)
			}
			observations := readFile(t, filepath.Join(fixture.stateDir, "runtime", "gsd-state",
				"integration-observations.jsonl"))
			if lines := strings.Count(strings.TrimSpace(observations), "\n") + 1; lines != 1 {
				t.Fatalf("boundary %s duplicated implementation: observations=%q", boundary, observations)
			}
		})
	}
}

func TestSuperviseRestartRunningTerminationPreservesAmbiguousAttempt(t *testing.T) {
	fixture := newFixture(t, "running-termination")
	cmd, stdout, stderr := fixture.superviseCommand(t, "")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	stateDir := filepath.Join(fixture.stateDir, "runtime", "gsd-state")
	waitForPath(t, filepath.Join(stateDir, "integration-running-ready"))
	activityPath := filepath.Join(fixture.stateDir, "activity", "segments", "activity-000001.jsonl")
	waitForHeartbeatCadence(t, activityPath)
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	fifo, err := os.OpenFile(filepath.Join(stateDir, "integration-running-release"), os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fifo.Write([]byte{'x'}); err != nil {
		_ = fifo.Close()
		t.Fatal(err)
	}
	if err := fifo.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err == nil {
		t.Fatalf("killed supervise unexpectedly succeeded\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	activity := readFile(t, activityPath)
	if len(activity) > 1024*1024 || strings.Contains(strings.ToLower(activity), "authorization:") {
		t.Fatalf("activity telemetry was unbounded or sensitive: bytes=%d", len(activity))
	}
	unknownWorktree := filepath.Join(fixture.attempts, "unknown-external-worktree")
	git(t, fixture.repo, "worktree", "add", "-q", "-b", "integration-unknown", unknownWorktree,
		fixture.baseHead)
	restarted := fixture.supervise(t)
	if restarted.err == nil {
		t.Fatalf("ambiguous running attempt restarted automatically\nstdout:\n%s\nstderr:\n%s",
			restarted.stdout, restarted.stderr)
	}
	fixture.assertCanonicalUnchanged(t)
	authority := filepath.Join(fixture.stateDir, "authority.db")
	assertCount(t, authority, `SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'running'`, 1)
	assertCount(t, authority, `SELECT COUNT(*) FROM delivery_runs WHERE state = 'awaiting_decision'`, 1)
	if _, err := os.Stat(unknownWorktree); err != nil {
		t.Fatalf("unknown worktree was not preserved: %v", err)
	}
	if got := git(t, fixture.repo, "branch", "--list", "integration-unknown"); got == "" {
		t.Fatal("unknown branch was not preserved")
	}
}

func TestSuperviseSIGINTTerminatesRunningChildAndConvergesOnRestart(t *testing.T) {
	t.Parallel()
	fixture := newFixture(t, "running-termination")
	cmd, stdout, stderr := fixture.superviseCommand(t, "")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	stateDir := filepath.Join(fixture.stateDir, "runtime", "gsd-state")
	waitForPath(t, filepath.Join(stateDir, "integration-running-ready"))
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatalf("interrupted supervise unexpectedly succeeded\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
		}
	case <-time.After(15 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("SIGINT did not terminate supervise and its running child")
	}
	fixture.assertCanonicalUnchanged(t)
	authority := filepath.Join(fixture.stateDir, "authority.db")
	assertCount(t, authority, `SELECT COUNT(*) FROM attempt_worktrees WHERE state = 'running'`, 0)
	restarted := fixture.supervise(t)
	if restarted.err == nil {
		t.Fatalf("interrupted attempt restarted without a human decision: %s", restarted.stdout)
	}
	assertCount(t, authority, `SELECT COUNT(*) FROM delivery_runs WHERE state = 'awaiting_decision'`, 1)
}

func TestSuperviseRestartIdempotentFinalGate(t *testing.T) {
	fixture := newFixture(t, "success")
	if result := fixture.supervise(t); result.err != nil {
		t.Fatalf("initial process-boundary run failed: %v\n%s", result.err, result.stderr)
	}
	if result := fixture.supervise(t); result.err != nil {
		t.Fatalf("idempotent restart failed: %v\nstdout:\n%s\nstderr:\n%s", result.err,
			result.stdout, result.stderr)
	}
	assertCount(t, filepath.Join(fixture.stateDir, "authority.db"),
		`SELECT COUNT(*) FROM promotion_journals WHERE state = 'complete'`, 1)
}

func newFixture(t *testing.T, scenario string) fixture {
	t.Helper()
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Fatalf("integration tag requires the qualified darwin/arm64 host, got %s/%s",
			runtime.GOOS, runtime.GOARCH)
	}
	moduleRoot := moduleRoot(t)
	binDir := t.TempDir()
	helperSource := filepath.Join(binDir, "fake-helper")
	build(t, moduleRoot, helperSource, "./integration/testhelper")
	helper := filepath.Join(binDir, "fake-gsd-"+scenario)
	piHelper := filepath.Join(binDir, "fake-pi-"+scenario)
	copyExecutable(t, helperSource, helper)
	copyExecutable(t, helperSource, piHelper)
	successHelper := filepath.Join(binDir, "fake-gsd-success")
	successPiHelper := filepath.Join(binDir, "fake-pi-success")
	if successHelper != helper {
		copyExecutable(t, helperSource, successHelper)
	}
	if successPiHelper != piHelper {
		copyExecutable(t, helperSource, successPiHelper)
	}
	copyExecutable(t, helperSource, filepath.Join(binDir, "gh"))
	shepherd := filepath.Join(binDir, "shepherd")
	build(t, moduleRoot, shepherd, "./cmd/shepherd")

	repo := realTempDir(t)
	git(t, repo, "init", "-q")
	git(t, repo, "config", "user.email", "integration@example.invalid")
	git(t, repo, "config", "user.name", "Integration")
	writeFile(t, filepath.Join(repo, "seed.txt"), "seed\n", 0o600)
	git(t, repo, "add", "seed.txt")
	git(t, repo, "commit", "-qm", "seed")
	branch := git(t, repo, "symbolic-ref", "--quiet", "--short", "HEAD")
	contextPath := filepath.Join(repo, "issue-context.json")
	contextRaw := fmt.Sprintf(`{
  "issue":389,
  "parent_issue":372,
  "objective":"process-boundary supervise integration",
  "scope":["real command boundary"],
  "non_goals":["live GitHub"],
  "acceptance_criteria":["final human gate"],
  "dependencies":[],
  "write_scope":["agent-runtime/shepherd/**"],
  "required_reading":["AGENTS.md"],
  "required_skills":["golang-how-to"],
  "tdd":{"red":"real command fake is unavailable","green":"supervise reaches final gate","refactor":"keep process fakes bounded"},
  "verification":["go test -tags=integration ./integration/..."],
  "safety":["no secrets or network"],
  "human_gates":["parent merge"],
  "branch":%q,
  "pr_base":"main",
  "review_route":"local",
  "sources":["issue #389"]
}`, branch)
	writeFile(t, contextPath, contextRaw, 0o600)
	writeFile(t, filepath.Join(repo, ".gsd", "STATE.md"), "process-boundary canonical state\n", 0o600)
	if strings.HasPrefix(scenario, "tracked-deletion") {
		writeFile(t, filepath.Join(repo, "agent-runtime", "shepherd", "integration-artifact.txt"), "canonical artifact to delete\n", 0o600)
		git(t, repo, "add", "agent-runtime/shepherd/integration-artifact.txt")
	}
	git(t, repo, "add", "issue-context.json")
	git(t, repo, "commit", "-qm", "test: add governed context")
	baseHead := git(t, repo, "rev-parse", "HEAD")
	issueContext, protectedRaw, err := contract.DecodeIssueContext(bytes.NewReader([]byte(contextRaw)), 389)
	if err != nil {
		t.Fatal(err)
	}
	contextDigest := sha256.Sum256(protectedRaw)
	if err := gsd.BootstrapIssueProject(repo, gsd.IssueProjectIdentity{
		DeliveryID: "issue-389", Issue: 389, ParentIssue: 372, Branch: branch, BaseBranch: "main",
		ProjectRoot: repo, InitialHead: baseHead,
		ContextHash: "sha256:" + hex.EncodeToString(contextDigest[:]), GSDVersion: "1.11.0",
	}, issueContext); err != nil {
		t.Fatal(err)
	}
	baseGitStatus := git(t, repo, "status", "--porcelain=v2", "--untracked-files=all")
	baseManifest, err := workspace.BuildGSDManifest(context.Background(), filepath.Join(repo, ".gsd"))
	if err != nil {
		t.Fatal(err)
	}

	root := realTempDir(t)
	stateDir := filepath.Join(root, "state")
	gsdHome := filepath.Join(root, "gsd-home")
	attempts := filepath.Join(root, "attempts")
	writeFile(t, filepath.Join(gsdHome, "agent", "settings.json"),
		"{\"defaultProvider\":\"openai-codex\",\"defaultModel\":\"gpt-5.6-sol\",\"defaultThinkingLevel\":\"high\"}\n", 0o600)
	writeFile(t, filepath.Join(gsdHome, "PREFERENCES.md"), modelPreferences, 0o600)
	loader := officialLoader(t, moduleRoot)
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	node, err = filepath.Abs(node)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "shepherd.json")
	config := map[string]any{
		"gsd_command": []string{node, loader}, "pi_command": []string{piHelper},
		"work_dir": repo, "gsd_home": gsdHome, "state_dir": stateDir, "attempt_root": attempts,
		"coordinator_model": "openai-codex/gpt-5.6-sol", "implementation_model": "openai-codex/gpt-5.5",
		"gsd_version": "1.11.0", "timeout_seconds": 30, "heartbeat_seconds": 1,
		"max_event_bytes": 8 * 1024 * 1024, "max_unit_attempts": 2,
		"repository": "polymetrics-ai/cli", "pull_request": 391, "runtime": "host",
	}
	writeJSON(t, configPath, config)
	return fixture{moduleRoot: moduleRoot, repo: repo, stateDir: stateDir, gsdHome: gsdHome,
		attempts: attempts, context: contextPath, config: configPath, shepherd: shepherd, helper: helper,
		piHelper: piHelper, successHelper: successHelper, successPiHelper: successPiHelper,
		githubState: filepath.Join(root, "fake-github-comments.json"),
		binDir:      binDir, baseHead: baseHead, baseBranch: branch, baseGSD: "process-boundary canonical state\n",
		baseGSDHash: baseManifest.Hash, baseGitStatus: baseGitStatus}
}

func (f fixture) assertCanonicalUnchanged(t *testing.T) {
	t.Helper()
	if got := git(t, f.repo, "rev-parse", "HEAD"); got != f.baseHead {
		t.Fatalf("canonical head changed: %s != %s", got, f.baseHead)
	}
	if branch := git(t, f.repo, "symbolic-ref", "--quiet", "--short", "HEAD"); branch != f.baseBranch {
		t.Fatalf("canonical branch changed: %s != %s", branch, f.baseBranch)
	}
	if status := git(t, f.repo, "status", "--porcelain=v2", "--untracked-files=all"); status != f.baseGitStatus {
		t.Fatalf("canonical worktree status changed: %q != %q", status, f.baseGitStatus)
	}
	manifest, err := workspace.BuildGSDManifest(context.Background(), filepath.Join(f.repo, ".gsd"))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Hash != f.baseGSDHash {
		t.Fatalf("canonical GSD manifest changed: %s != %s", manifest.Hash, f.baseGSDHash)
	}
}

func (f *fixture) useSuccessHelper(t *testing.T) {
	t.Helper()
	f.helper = f.successHelper
	f.piHelper = f.successPiHelper
	updateConfig(t, f.config, func(config map[string]any) {
		config["pi_command"] = []string{f.successPiHelper}
	})
}

func (f fixture) supervise(t *testing.T) commandResult {
	t.Helper()
	return f.runSupervise(t, "")
}

func (f fixture) superviseAtBoundary(t *testing.T, boundary string) commandResult {
	t.Helper()
	return f.runSupervise(t, boundary)
}

func (f fixture) superviseWithEnv(t *testing.T, values map[string]string) commandResult {
	t.Helper()
	cmd, stdout, stderr := f.superviseCommand(t, "")
	for name, value := range values {
		cmd.Env = append(cmd.Env, name+"="+value)
	}
	err := cmd.Run()
	return commandResult{stdout: stdout.Bytes(), stderr: stderr.Bytes(), err: err}
}

func (f fixture) runSupervise(t *testing.T, boundary string) commandResult {
	t.Helper()
	cmd, stdout, stderr := f.superviseCommand(t, boundary)
	err := cmd.Run()
	return commandResult{stdout: stdout.Bytes(), stderr: stderr.Bytes(), err: err}
}

func (f fixture) superviseCommand(t *testing.T, boundary string) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	cmd := exec.Command(f.shepherd, "supervise", "--config", f.config, "--issue", "389",
		"--context", filepath.Base(f.context))
	cmd.Dir = f.repo
	processHome := filepath.Join(f.stateDir, "integration-process-home")
	if err := os.MkdirAll(processHome, 0o700); err != nil {
		t.Fatal(err)
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	gitPath, err = filepath.Abs(gitPath)
	if err != nil {
		t.Fatal(err)
	}
	cmd.Env = []string{
		"HOME=" + processHome,
		"XDG_CONFIG_HOME=" + filepath.Join(processHome, ".config"),
		"PATH=" + strings.Join([]string{f.binDir, filepath.Dir(gitPath), "/usr/bin", "/bin"},
			string(os.PathListSeparator)),
		"TMPDIR=" + os.TempDir(),
		"LANG=C.UTF-8", "LC_ALL=C.UTF-8",
		"GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0", "GH_PROMPT_DISABLED=1",
		"SHEPHERD_INTEGRATION_GSD_EXECUTOR=" + f.helper,
		"SHEPHERD_INTEGRATION_FAKE_GITHUB_STATE=" + f.githubState,
		"SHEPHERD_INTEGRATION_EXIT_AWAITING_DECISION=1",
	}
	if strings.Contains(filepath.Base(f.helper), "post-validation-mutation") {
		cmd.Env = append(cmd.Env, "SHEPHERD_INTEGRATION_MUTATE_AFTER_VALIDATION=1")
	}
	if boundary != "" {
		cmd.Env = append(cmd.Env, "SHEPHERD_INTEGRATION_CRASH_BOUNDARY="+boundary)
	}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	return cmd, stdout, stderr
}

type statusRecord struct {
	Status     string `json:"status"`
	NextAction string `json:"next_action"`
}

func finalStatus(t *testing.T, stdout []byte) statusRecord {
	t.Helper()
	var result statusRecord
	scanner := bufio.NewScanner(bytes.NewReader(stdout))
	for scanner.Scan() {
		var candidate statusRecord
		if json.Unmarshal(scanner.Bytes(), &candidate) == nil && candidate.Status != "" {
			result = candidate
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return result
}

func realTempDir(t *testing.T) string {
	t.Helper()
	path := t.TempDir()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Dir(cwd)
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("resolve module root from %s: %v", cwd, err)
	}
	return root
}

func officialLoader(t *testing.T, moduleRoot string) string {
	t.Helper()
	if configured := os.Getenv("GSD_OFFICIAL_LOADER"); configured != "" {
		path, err := filepath.Abs(configured)
		if err != nil {
			t.Fatal(err)
		}
		return path
	}
	repoRoot := filepath.Dir(filepath.Dir(moduleRoot))
	path := filepath.Join(filepath.Dir(repoRoot), ".tools", "gsd-pi-1.11.0", "node_modules",
		"@opengsd", "gsd-pi", "dist", "loader.js")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("official GSD loader is unavailable at %s; set GSD_OFFICIAL_LOADER: %v", path, err)
	}
	return path
}

func build(t *testing.T, moduleRoot, output, target string) {
	t.Helper()
	args := []string{"build", "-tags=integration"}
	if raceEnabled {
		args = append(args, "-race")
	}
	args = append(args, "-o", output, target)
	cmd := exec.Command("go", args...)
	cmd.Dir = moduleRoot
	if raw, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build %s: %v: %s", target, err, raw)
	}
}

func copyExecutable(t *testing.T, source, target string) {
	t.Helper()
	raw, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, raw, 0o700); err != nil {
		t.Fatal(err)
	}
}

func git(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	raw, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, raw)
	}
	return strings.TrimSpace(string(raw))
}

func writeFile(t *testing.T, path, value string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value), mode); err != nil {
		t.Fatal(err)
	}
}

func updateConfig(t *testing.T, path string, mutate func(map[string]any)) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatal(err)
	}
	mutate(config)
	writeJSON(t, path, config)
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, string(raw), 0o600)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func waitForClaimExpiry(t *testing.T, path string) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	var expiresAt int64
	if err := db.QueryRow(`SELECT expires_at FROM claims WHERE closed_at = 0`).Scan(&expiresAt); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	remaining := time.Until(time.Unix(0, expiresAt).UTC())
	if remaining > 0 {
		time.Sleep(remaining + 10*time.Millisecond)
	}
}

func waitForHeartbeatCadence(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		raw, err := os.ReadFile(path)
		if err == nil {
			var observed []time.Time
			for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
				var event struct {
					Kind string    `json:"kind"`
					At   time.Time `json:"at"`
				}
				if json.Unmarshal([]byte(line), &event) == nil && event.Kind == "heartbeat" {
					observed = append(observed, event.At)
				}
			}
			if len(observed) >= 2 {
				gap := observed[len(observed)-1].Sub(observed[len(observed)-2])
				if gap <= 0 || gap > 15*time.Second {
					t.Fatalf("heartbeat gap=%s outside bounded cadence", gap)
				}
				return
			}
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for two heartbeats in %s", path)
}

func waitForPath(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for durable synchronization path %s", path)
}

type storedGitHubComment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
	User struct {
		Login string `json:"login"`
		ID    int64  `json:"id"`
		Type  string `json:"type"`
	} `json:"user"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func botCommentCount(t *testing.T, path string) int {
	t.Helper()
	comments := readStoredComments(t, path)
	assertRecordedGitHubCalls(t, path+".calls")
	count := 0
	for _, comment := range comments {
		if comment.User.Type == "Bot" {
			count++
		}
	}
	return count
}

func assertRecordedGitHubCalls(t *testing.T, path string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) == 0 {
		t.Fatal("fake GitHub boundary recorded no calls")
	}
	for _, line := range lines {
		var record struct {
			Args []string `json:"args"`
		}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		joined := strings.Join(record.Args, " ")
		if strings.Contains(joined, "graphql") || strings.Contains(joined, "/pulls/") ||
			strings.Contains(joined, "merge") || !strings.HasPrefix(joined, "api ") {
			t.Fatalf("forbidden GitHub invocation was recorded: %q", joined)
		}
	}
}

func appendHumanReply(t *testing.T, path string, id int64, body string) {
	t.Helper()
	appendStoredReply(t, path, id, body, "karthik-sivadas", 6113982, "User", false)
}

func appendStoredReply(t *testing.T, path string, id int64, body, login string, actorID int64,
	actorType string, edited bool,
) {
	t.Helper()
	comments := readStoredComments(t, path)
	created := time.Now().UTC()
	updated := created
	if edited {
		updated = updated.Add(time.Second)
	}
	comment := storedGitHubComment{ID: id, Body: body,
		CreatedAt: created.Format(time.RFC3339Nano), UpdatedAt: updated.Format(time.RFC3339Nano)}
	comment.User.Login, comment.User.ID, comment.User.Type = login, actorID, actorType
	comments = append(comments, comment)
	writeJSON(t, path, comments)
}

func readStoredComments(t *testing.T, path string) []storedGitHubComment {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var comments []storedGitHubComment
	if err := json.Unmarshal(raw, &comments); err != nil {
		t.Fatal(err)
	}
	return comments
}

func downgradeProofDatabaseToLegacy(t *testing.T, authorityPath string) {
	t.Helper()
	db, err := sql.Open("sqlite", authorityPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var proofID, raw string
	if err := db.QueryRow(`SELECT proof_id, expected_artifact FROM artifact_proofs`).Scan(&proofID, &raw); err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal([]byte(raw), &manifest); err != nil {
		t.Fatal(err)
	}
	delete(manifest, "gsd_manifest_hash")
	legacyRaw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(legacyRaw)
	hash := "sha256:" + hex.EncodeToString(digest[:])
	for _, update := range []struct {
		query string
		args  []any
	}{
		{`UPDATE artifact_proofs SET expected_artifact = ?, artifact_hash = ? WHERE proof_id = ?`,
			[]any{string(legacyRaw), hash, proofID}},
		{`UPDATE attestations SET evidence_hash = ?`, []any{hash}},
		{`UPDATE promotion_journals SET evidence_hash = ?`, []any{hash}},
	} {
		if _, err := db.Exec(update.query, update.args...); err != nil {
			t.Fatal(err)
		}
	}
}

func assertProofMatchesCanonical(t *testing.T, authorityPath, repo, candidateHead, expectedUnit string,
	expectedPhases, expectedTools []string,
) {
	t.Helper()
	db, err := sql.Open("sqlite", authorityPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var raw, evidenceHash, startHead, validatedHead string
	if err := db.QueryRow(`SELECT expected_artifact, artifact_hash, start_head, validated_head FROM artifact_proofs
		WHERE candidate_head = ?`, candidateHead).Scan(&raw, &evidenceHash, &startHead, &validatedHead); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte(raw))
	if got := "sha256:" + hex.EncodeToString(digest[:]); got != evidenceHash {
		t.Fatalf("proof evidence hash=%s want %s", evidenceHash, got)
	}
	if validatedHead != candidateHead {
		t.Fatalf("validated head=%s want candidate %s", validatedHead, candidateHead)
	}
	var manifest struct {
		UnitType              string   `json:"unit_type"`
		PhaseChain            []string `json:"phase_chain"`
		RequiredWorkflowTools []string `json:"required_workflow_tools"`
		ObservedWorkflowTools []string `json:"observed_workflow_tools"`
		GSDManifestHash       string   `json:"gsd_manifest_hash"`
		Artifacts             []struct {
			Path    string `json:"path"`
			Hash    string `json:"hash"`
			Deleted bool   `json:"deleted,omitempty"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal([]byte(raw), &manifest); err != nil || manifest.GSDManifestHash == "" ||
		len(manifest.Artifacts) == 0 {
		t.Fatalf("invalid proof manifest: %v", err)
	}
	if manifest.UnitType != expectedUnit || strings.Join(manifest.PhaseChain, ",") != strings.Join(expectedPhases, ",") ||
		strings.Join(manifest.RequiredWorkflowTools, ",") != strings.Join(expectedTools, ",") {
		t.Fatalf("official metadata mismatch: unit=%s phases=%v tools=%v", manifest.UnitType,
			manifest.PhaseChain, manifest.RequiredWorkflowTools)
	}
	for _, required := range expectedTools {
		observed := false
		for _, candidate := range manifest.ObservedWorkflowTools {
			if candidate == required {
				observed = true
				break
			}
		}
		if !observed {
			t.Fatalf("proof lacks required workflow transition %s: observed=%v",
				required, manifest.ObservedWorkflowTools)
		}
	}
	proofPaths := make(map[string]struct{}, len(manifest.Artifacts))
	for _, artifact := range manifest.Artifacts {
		proofPaths[artifact.Path] = struct{}{}
		if artifact.Deleted {
			if artifact.Hash != "sha256:0000000000000000000000000000000000000000000000000000000000000000" {
				t.Fatalf("deleted artifact %s hash=%s", artifact.Path, artifact.Hash)
			}
			if _, err := os.Lstat(filepath.Join(repo, filepath.FromSlash(artifact.Path))); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("deleted artifact %s exists after promotion: %v", artifact.Path, err)
			}
			continue
		}
		content, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(artifact.Path)))
		if err != nil {
			t.Fatalf("read promoted proof artifact %s: %v", artifact.Path, err)
		}
		hash := sha256.Sum256(content)
		if got := "sha256:" + hex.EncodeToString(hash[:]); got != artifact.Hash {
			t.Fatalf("artifact %s hash=%s want %s", artifact.Path, got, artifact.Hash)
		}
	}
	changed := git(t, repo, "diff", "--name-only", startHead, candidateHead, "--", ".")
	for _, path := range strings.Fields(changed) {
		if _, ok := proofPaths[path]; !ok {
			t.Fatalf("candidate diff path %s is absent from proof", path)
		}
	}
	gsdManifest, err := workspace.SnapshotGSDManifest(context.Background(), filepath.Join(repo, ".gsd"),
		filepath.Dir(authorityPath))
	if err != nil || gsdManifest.Hash != manifest.GSDManifestHash {
		t.Fatalf("promoted GSD manifest=%s want %s err=%v", gsdManifest.Hash, manifest.GSDManifestHash, err)
	}
}

func assertQuestionBody(t *testing.T, authorityPath, commentsPath string) {
	t.Helper()
	db, err := sql.Open("sqlite", authorityPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var requestID, headSHA, evidence string
	var generation int64
	if err := db.QueryRow(`SELECT request_id, generation, head_sha, evidence FROM decision_requests
		WHERE status = 'published'`).Scan(&requestID, &generation, &headSHA, &evidence); err != nil {
		t.Fatal(err)
	}
	comments := readStoredComments(t, commentsPath)
	if len(comments) == 0 {
		t.Fatal("published question has no external comment")
	}
	body := comments[0].Body
	for _, required := range []string{"@karthik-sivadas", "Request: `" + requestID + "`",
		fmt.Sprintf("Generation: `%d`", generation), "Head: `" + headSHA + "`", "Evidence: " + evidence,
		"Recommended option:", "Safe default at expiry:", "Expires:",
		"/shepherd decide " + requestID + " <option>"} {
		if !strings.Contains(body, required) {
			t.Fatalf("question body is missing %q: %s", required, body)
		}
	}
}

func decisionIdentity(t *testing.T, authorityPath string) (string, int64) {
	t.Helper()
	db, err := sql.Open("sqlite", authorityPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var requestID string
	var commentID int64
	if err := db.QueryRow(`SELECT request_id, github_comment_id FROM decision_requests
		WHERE status = 'published'`).Scan(&requestID, &commentID); err != nil {
		t.Fatal(err)
	}
	return requestID, commentID
}

func queryStrings(t *testing.T, path, query string) []string {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	rows, err := db.Query(query)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var values []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			t.Fatal(err)
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return values
}

func withoutEnvironment(environment []string, name string) []string {
	prefix := name + "="
	filtered := make([]string, 0, len(environment))
	for _, entry := range environment {
		if !strings.HasPrefix(entry, prefix) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func waitForSQLiteCount(t *testing.T, path, query string, want int) {
	t.Helper()
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		db, err := sql.Open("sqlite", path)
		if err == nil {
			var got int
			err = db.QueryRow(query).Scan(&got)
			_ = db.Close()
			if err == nil && got == want {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for query %q count=%d", query, want)
}

func assertCount(t *testing.T, path, query string, want int, args ...any) {
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
	var got int
	if err := db.QueryRow(query, args...).Scan(&got); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	if got != want {
		t.Fatalf("query %q count=%d want=%d", query, got, want)
	}
}

const modelPreferences = `---
version: 1
models:
  research: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  planning: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  discuss: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  execution: { provider: openai-codex, model: gpt-5.5, thinking: high }
  execution_simple: { provider: openai-codex, model: gpt-5.5, thinking: high }
  completion: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  validation: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  subagent: { provider: openai-codex, model: gpt-5.5, thinking: high }
  uat: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
---
`
