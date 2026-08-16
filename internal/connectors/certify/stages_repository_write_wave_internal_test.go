package certify

import (
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func testRepositoryWriteWaveProfile(t *testing.T) (string, *engine.CertificationWriteWaveSpec) {
	t.Helper()
	entries, err := defs.FS.ReadDir(".")
	if err != nil {
		t.Fatalf("read definition root: %v", err)
	}
	var connector string
	var profile *engine.CertificationWriteWaveSpec
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate, ok := certificationWriteWaveFor(entry.Name())
		if !ok {
			continue
		}
		if profile != nil {
			t.Fatalf("repository write-wave test requires one definition-owned profile, found %q and %q", connector, entry.Name())
		}
		connector, profile = entry.Name(), candidate
	}
	if profile == nil {
		t.Fatal("definition-owned repository write wave is missing")
	}
	return connector, profile
}

func TestRepositoryWaveMatchesDefinitionInventory(t *testing.T) {
	connector, profile := testRepositoryWriteWaveProfile(t)
	seen := make(map[string]bool, len(profile.Actions))
	for _, action := range profile.Actions {
		if seen[action] {
			t.Fatalf("repository wave repeats action %q", action)
		}
		seen[action] = true
	}
	inventory, err := writeActionInventoryFor(connector)
	if err != nil {
		t.Fatalf("writeActionInventoryFor() = %v", err)
	}
	ready := map[string]bool{}
	for _, item := range inventory {
		if item.Classification == writeClassificationRepositoryWaveReady {
			ready[item.Action] = true
		}
	}
	if got, want := len(seen), len(ready); got != want {
		t.Fatalf("repository wave has %d actions, approved inventory has %d", got, want)
	}
	for action := range ready {
		if !seen[action] {
			t.Fatalf("approved repository action %q is missing from the production wave", action)
		}
	}
}

func TestRepositoryWaveFixtureIsDefinitionOwned(t *testing.T) {
	_, profile := testRepositoryWriteWaveProfile(t)
	if profile.Fixture.Config["owner"] != "Polymetrics-Cert" || profile.Fixture.Config["repo"] != "pm-cert-3993-20260810-wz0fru" {
		t.Fatalf("repository wave fixture = %#v, want the captain-approved disposable fixture", profile.Fixture)
	}
}

func TestRepositoryWaveRejectsAnyOtherRepositoryBeforeWriteSetup(t *testing.T) {
	connector, profile := testRepositoryWriteWaveProfile(t)
	rc := &runContext{opts: Options{
		Connector: connector,
		Write:     true,
		Config:    map[string]string{"owner": "third-party", "repo": "not-authorized"},
	}}
	rep := &Report{}
	if err := stageRepositoryWriteWave(rc, rep); err != nil {
		t.Fatalf("stageRepositoryWriteWave() error = %v", err)
	}
	if len(rep.Stages) != 1 || rep.Stages[0].Passed || !strings.Contains(rep.Stages[0].Error, profile.Fixture.Description) {
		t.Fatalf("non-fixture stage = %+v, want pre-write disposable-fixture refusal", rep.Stages)
	}
}

func TestRepositoryWaveFailureNeverRemainsPass(t *testing.T) {
	_, profile := testRepositoryWriteWaveProfile(t)
	action := profile.ActionBindings["issue_update"]
	w := &repositoryWriteWave{
		rep: &Report{Capabilities: Capabilities{WriteActions: map[string]WriteActionResult{
			action: {Result: "pass"},
		}}},
		declarations: map[string]writeActionInventoryItem{
			action: {Action: action, Path: "/repos/{owner}/{repo}/issues/{issue_number}", Risk: "mutates issue metadata"},
		},
	}
	w.markFailed(action, profile.TagPrefix+"test", "production reverse write returned a provider error")
	got := w.rep.Capabilities.WriteActions[action]
	if got.Result != "fail" {
		t.Fatalf("broken update_issue result = %+v, want fail rather than inherited pass", got)
	}
	if got.Verify != "" {
		t.Fatalf("broken update_issue verify = %q, want no read-back claim", got.Verify)
	}
}

func TestRepositoryWaveRejectsUnknownBoundedWindow(t *testing.T) {
	connector, profile := testRepositoryWriteWaveProfile(t)
	w := &repositoryWriteWave{connector: connector, profile: profile}
	if _, err := w.runSelected("every_repository"); err == nil {
		t.Fatal("runSelected accepted an undeclared bounded window")
	}
}

func TestRepositoryWaveInitializesWriteActionsAfterVerifiedCleanup(t *testing.T) {
	_, profile := testRepositoryWriteWaveProfile(t)
	action := profile.ActionBindings["milestone_create"]
	ledger, err := NewLedger(t.TempDir())
	if err != nil {
		t.Fatalf("NewLedger() = %v", err)
	}
	w := &repositoryWriteWave{
		rep:       &Report{},
		ledger:    ledger,
		completed: map[string]bool{action: true},
		declarations: map[string]writeActionInventoryItem{
			action: {Action: action, Path: "/repos/{owner}/{repo}/milestones", Risk: "creates planning metadata"},
		},
	}
	w.markScenarioPassed(profile.TagPrefix+"test", []string{action}, "cleaned")
	got := w.rep.Capabilities.WriteActions[action]
	if got.Result != "pass" || got.Verify != "read_back" {
		t.Fatalf("verified cleanup result = %+v, want pass with read_back", got)
	}
}

func TestRepositoryWaveRetriesIndependentReadBack(t *testing.T) {
	attempts := 0
	w := &repositoryWriteWave{
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

func TestRepositoryWaveRecoveryOnlyAcceptsOwnedIssueScenarios(t *testing.T) {
	connector, profile := testRepositoryWriteWaveProfile(t)
	w := &repositoryWriteWave{connector: connector, profile: profile}
	for _, scenario := range []string{
		w.scenario("issue_lifecycle").LedgerName,
		w.scenario("issue_comments").LedgerName,
		w.scenario("issue_lock").LedgerName,
		w.scenario("issue_labels").LedgerName,
	} {
		if !w.isRecoverableIssueScenario(scenario) {
			t.Fatalf("isRecoverableIssueScenario(%q) = false, want true", scenario)
		}
	}
	if w.isRecoverableIssueScenario(w.scenario("topics").LedgerName) {
		t.Fatal("topic replacement recovery must not guess an unknown baseline")
	}
}

func TestRepositoryWaveLeakInitializesWriteActions(t *testing.T) {
	_, profile := testRepositoryWriteWaveProfile(t)
	action := profile.ActionBindings["issue_create"]
	w := &repositoryWriteWave{
		rep: &Report{},
		declarations: map[string]writeActionInventoryItem{
			action: {Action: action, Path: "/repos/{owner}/{repo}/issues", Risk: "creates issue"},
		},
	}
	w.recordLeak(profile.TagPrefix+"test", []string{action}, "read-back never observed the tagged issue")
	got := w.rep.Capabilities.WriteActions[action]
	if got.Result != "leaked_resource" {
		t.Fatalf("leaked write action = %+v, want leaked_resource", got)
	}
}

func TestRepositoryWaveBlockedCommitCommentNeverClaimsCoverage(t *testing.T) {
	_, profile := testRepositoryWriteWaveProfile(t)
	blocked := profile.BlockedActions[0]
	w := &repositoryWriteWave{
		rep: &Report{},
		declarations: map[string]writeActionInventoryItem{
			"create_commit_comment": {Action: "create_commit_comment", Path: "/repos/{owner}/{repo}/commits/{commit_sha}/comments", Risk: "creates comment"},
			"update_commit_comment": {Action: "update_commit_comment", Path: "/repos/{owner}/{repo}/comments/{comment_id}", Risk: "updates comment"},
			"delete_commit_comment": {Action: "delete_commit_comment", Path: "/repos/{owner}/{repo}/comments/{comment_id}", Risk: "deletes comment"},
		},
	}
	w.markBlocked(blocked.Actions, profile.TagPrefix+"commit-comment-test", blocked.Reason, "verified")
	if w.rep.Passed {
		t.Fatal("blocked item read left report passed")
	}
	for _, action := range blocked.Actions {
		got := w.rep.Capabilities.WriteActions[action]
		if got.Result != "blocked" || got.Verify != "unavailable" || got.Cleanup != "verified" {
			t.Fatalf("blocked %s = %+v, want non-pass unavailable verification with verified cleanup", action, got)
		}
		if got.Reason != blocked.Reason {
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

func TestRepositoryWaveLabelNamesFitProviderLimit(t *testing.T) {
	_, profile := testRepositoryWriteWaveProfile(t)
	first, updated, second := waveLabelNames(profile.TagPrefix + "label-12345678-1234567890")
	for _, name := range []string{first, updated, second} {
		if len(name) > 50 {
			t.Fatalf("generated label %q has length %d, provider permits at most 50", name, len(name))
		}
	}
	if first == updated || first == second || updated == second {
		t.Fatalf("generated label names must be distinct: %q, %q, %q", first, updated, second)
	}
}
