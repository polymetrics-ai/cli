package certify

import (
	"errors"
	"strings"
	"testing"
)

func TestGithubRepositoryWaveMatchesApprovedRepositoryInventory(t *testing.T) {
	seen := make(map[string]bool, len(githubRepositoryWaveActions))
	for _, action := range githubRepositoryWaveActions {
		if seen[action] {
			t.Fatalf("repository wave repeats action %q", action)
		}
		seen[action] = true
		if _, ok := githubRepositoryWaveReadyActions[action]; !ok {
			t.Fatalf("repository wave action %q is not in the approved inventory", action)
		}
	}
	if got, want := len(seen), len(githubRepositoryWaveReadyActions); got != want {
		t.Fatalf("repository wave has %d actions, approved inventory has %d", got, want)
	}
	for action := range githubRepositoryWaveReadyActions {
		if !seen[action] {
			t.Fatalf("approved repository action %q is missing from the production wave", action)
		}
	}
}

func TestGithubRepositoryWaveFixtureIsCaptainApprovedDisposableRepository(t *testing.T) {
	if githubRepositoryWaveFixtureOwner != "Polymetrics-Cert" || githubRepositoryWaveFixtureRepo != "pm-cert-3993-20260810-wz0fru" {
		t.Fatalf("repository wave fixture = %s/%s, want the captain-approved disposable fixture", githubRepositoryWaveFixtureOwner, githubRepositoryWaveFixtureRepo)
	}
}

func TestGithubRepositoryWaveRejectsAnyOtherRepositoryBeforeWriteSetup(t *testing.T) {
	rc := &runContext{opts: Options{
		Connector: "github",
		Write:     true,
		Config:    map[string]string{"owner": "third-party", "repo": "not-authorized"},
	}}
	rep := &Report{}
	if err := stageGithubRepositoryWriteWave(rc, rep); err != nil {
		t.Fatalf("stageGithubRepositoryWriteWave() error = %v", err)
	}
	if len(rep.Stages) != 1 || rep.Stages[0].Passed || !strings.Contains(rep.Stages[0].Error, githubRepositoryWaveFixtureOwner+"/"+githubRepositoryWaveFixtureRepo) {
		t.Fatalf("non-fixture stage = %+v, want pre-write disposable-fixture refusal", rep.Stages)
	}
}

func TestGithubRepositoryWaveFailureNeverRemainsPass(t *testing.T) {
	w := &githubRepositoryWave{
		rep: &Report{Capabilities: Capabilities{WriteActions: map[string]WriteActionResult{
			"update_issue": {Result: "pass"},
		}}},
		declarations: map[string]writeActionInventoryItem{
			"update_issue": {Action: "update_issue", Path: "/repos/{owner}/{repo}/issues/{issue_number}", Risk: "mutates issue metadata"},
		},
	}
	w.markFailed("update_issue", "pm-cert-github-test", "production reverse write returned a provider error")
	got := w.rep.Capabilities.WriteActions["update_issue"]
	if got.Result != "fail" {
		t.Fatalf("broken update_issue result = %+v, want fail rather than inherited pass", got)
	}
	if got.Verify != "" {
		t.Fatalf("broken update_issue verify = %q, want no read-back claim", got.Verify)
	}
}

func TestGithubRepositoryWaveRejectsUnknownBoundedWindow(t *testing.T) {
	w := &githubRepositoryWave{}
	if _, err := w.runSelected("every_repository"); err == nil {
		t.Fatal("runSelected accepted an undeclared bounded window")
	}
}

func TestGithubRepositoryWaveInitializesWriteActionsAfterVerifiedCleanup(t *testing.T) {
	ledger, err := NewLedger(t.TempDir())
	if err != nil {
		t.Fatalf("NewLedger() = %v", err)
	}
	w := &githubRepositoryWave{
		rep:       &Report{},
		ledger:    ledger,
		completed: map[string]bool{"create_milestone": true},
		declarations: map[string]writeActionInventoryItem{
			"create_milestone": {Action: "create_milestone", Path: "/repos/{owner}/{repo}/milestones", Risk: "creates planning metadata"},
		},
	}
	w.markScenarioPassed("pm-cert-github-test", []string{"create_milestone"}, "cleaned")
	got := w.rep.Capabilities.WriteActions["create_milestone"]
	if got.Result != "pass" || got.Verify != "read_back" {
		t.Fatalf("verified cleanup result = %+v, want pass with read_back", got)
	}
}

func TestGithubRepositoryWaveRetriesIndependentReadBack(t *testing.T) {
	attempts := 0
	w := &githubRepositoryWave{
		readBackAttempts: 3,
		waitForReadBack:  func(_ int) error { return nil },
	}
	id, err := w.readBackWithRetry(func() (string, error) {
		attempts++
		if attempts < 3 {
			return "", errors.New("provider projection has not caught up")
		}
		return "issue:42", nil
	})
	if err != nil {
		t.Fatalf("readBackWithRetry() = %v", err)
	}
	if id != "issue:42" {
		t.Fatalf("readBackWithRetry() id = %q, want issue:42", id)
	}
	if attempts != 3 {
		t.Fatalf("read-back attempts = %d, want 3", attempts)
	}
}

func TestGithubRepositoryWaveRecoveryOnlyAcceptsOwnedIssueScenarios(t *testing.T) {
	for _, scenario := range []string{
		"github_repository_issue_lifecycle",
		"github_repository_issue_comments",
		"github_repository_issue_lock",
		"github_repository_issue_labels",
	} {
		if !isRecoverableIssueScenario(scenario) {
			t.Fatalf("isRecoverableIssueScenario(%q) = false, want true", scenario)
		}
	}
	if isRecoverableIssueScenario("github_repository_topics") {
		t.Fatal("topic replacement recovery must not guess an unknown baseline")
	}
}

func TestGithubRepositoryWaveLeakInitializesWriteActions(t *testing.T) {
	w := &githubRepositoryWave{
		rep: &Report{},
		declarations: map[string]writeActionInventoryItem{
			"create_issue": {Action: "create_issue", Path: "/repos/{owner}/{repo}/issues", Risk: "creates issue"},
		},
	}
	w.recordLeak("pm-cert-github-test", []string{"create_issue"}, "read-back never observed the tagged issue")
	got := w.rep.Capabilities.WriteActions["create_issue"]
	if got.Result != "leaked_resource" {
		t.Fatalf("leaked write action = %+v, want leaked_resource", got)
	}
}

func TestGithubRepositoryWaveBlockedCommitCommentNeverClaimsCoverage(t *testing.T) {
	w := &githubRepositoryWave{
		rep: &Report{},
		declarations: map[string]writeActionInventoryItem{
			"create_commit_comment": {Action: "create_commit_comment", Path: "/repos/{owner}/{repo}/commits/{commit_sha}/comments", Risk: "creates comment"},
			"update_commit_comment": {Action: "update_commit_comment", Path: "/repos/{owner}/{repo}/comments/{comment_id}", Risk: "updates comment"},
			"delete_commit_comment": {Action: "delete_commit_comment", Path: "/repos/{owner}/{repo}/comments/{comment_id}", Risk: "deletes comment"},
		},
	}
	w.markBlocked(githubCommitCommentActions, "pm-cert-github-commit-comment-test", githubCommitCommentItemReadBlockedReason, "verified")
	if w.rep.Passed {
		t.Fatal("blocked item read left report passed")
	}
	for _, action := range githubCommitCommentActions {
		got := w.rep.Capabilities.WriteActions[action]
		if got.Result != "blocked" || got.Verify != "unavailable" || got.Cleanup != "verified" {
			t.Fatalf("blocked %s = %+v, want non-pass unavailable verification with verified cleanup", action, got)
		}
		if got.Reason != githubCommitCommentItemReadBlockedReason {
			t.Fatalf("blocked %s reason = %q, want exact provider refusal", action, got.Reason)
		}
	}
}

func TestBlockedWriteActionFailsReportAggregateButNotLiveDoesNot(t *testing.T) {
	if !hasBlockingWriteAction(map[string]WriteActionResult{"commit": {Result: "blocked"}}) {
		t.Fatal("blocked write action did not fail the report aggregate")
	}
	if hasBlockingWriteAction(map[string]WriteActionResult{"deferred": {Result: "not_live"}}) {
		t.Fatal("not_live write action unexpectedly became a generic report failure")
	}
	if !hasIncompleteWriteAction(map[string]WriteActionResult{"deferred": {Result: "not_live"}}) {
		t.Fatal("not_live write action did not mark the mode partial_live")
	}
}

func TestRecoveredInterruptedWriteFailsReportAggregate(t *testing.T) {
	if !hasBlockingWriteAction(map[string]WriteActionResult{"topics": {Result: "recovered_unverified"}}) {
		t.Fatal("recovered-only write action did not fail the report aggregate")
	}
}

func TestCommitCommentIDFromLedger(t *testing.T) {
	id, err := commitCommentIDFromLedger(LedgerStatus{ResourceID: "commit_comment:42"})
	if err != nil || id != 42 {
		t.Fatalf("commitCommentIDFromLedger() = %d, %v; want 42, nil", id, err)
	}
	if _, err := commitCommentIDFromLedger(LedgerStatus{EntityHint: "commit_comment:not-a-number"}); err == nil {
		t.Fatal("commitCommentIDFromLedger accepted a non-numeric resource id")
	}
}

func TestGithubRepositoryWaveLabelNamesFitGitHubLimit(t *testing.T) {
	first, updated, second := githubWaveLabelNames("pm-cert-github-label-12345678-1234567890")
	for _, name := range []string{first, updated, second} {
		if len(name) > 50 {
			t.Fatalf("generated label %q has length %d, GitHub permits at most 50", name, len(name))
		}
	}
	if first == updated || first == second || updated == second {
		t.Fatalf("generated label names must be distinct: %q, %q, %q", first, updated, second)
	}
}
