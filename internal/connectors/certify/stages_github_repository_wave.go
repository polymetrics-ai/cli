package certify

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors/engine"
)

// githubRepositoryWaveActions is the captain-approved, reversible first live
// wave. It deliberately names actions rather than paths: the same endpoint
// can be safe for a tagged, run-owned record and unsafe for an inherited
// repository resource. Every item is sent through reverse plan, preview, and
// run; no action in this list is a synthetic request or a DryRunWrite claim.
var githubRepositoryWaveActions = []string{
	"create_issue", "update_issue", "comment_issue", "update_issue_comment", "delete_issue_comment",
	"lock_issue", "unlock_issue", "set_issue_labels", "add_issue_labels", "remove_issue_label",
	"close_issue", "reopen_issue",
	"create_label", "update_label", "delete_label",
	"create_milestone", "update_milestone", "delete_milestone",
	"create_release", "update_release", "delete_release",
	"create_commit_comment", "update_commit_comment", "delete_commit_comment",
	"create_ref", "update_ref", "delete_ref", "replace_repo_topics",
}

var githubCommitCommentActions = []string{
	"create_commit_comment", "update_commit_comment", "delete_commit_comment",
}

const githubCommitCommentItemReadBlockedReason = "GitHub direct read GET /repos/{owner}/{repo}/comments/{comment_id} is blocked: the disposable fine-grained PAT was refused with \"Resource not accessible by personal access token\". GitHub's picker names the required permission \"Metadata\" repository permissions (read); do not count the mutation as certified until that item read succeeds."

const (
	githubRepositoryWaveFixtureOwner = "Polymetrics-Cert"
	githubRepositoryWaveFixtureRepo  = "pm-cert-3993-20260810-wz0fru"
)

type githubRepositoryWave struct {
	rc     *runContext
	rep    *Report
	ledger *Ledger
	runID  string

	// declarations lets every outcome retain its exact provider path and risk,
	// including a failure before a mutation reaches GitHub.
	declarations map[string]writeActionInventoryItem
	completed    map[string]bool
	readSequence int

	// GitHub can acknowledge a mutation before a fresh collection read sees
	// it. Keep that eventual-consistency wait bounded and testable; a failed
	// read-back remains a certification failure and a reported potential leak.
	readBackAttempts int
	waitForReadBack  func(attempt int) error
}

const (
	defaultGitHubReadBackAttempts = 5
	githubReadBackRetryBase       = time.Second
)

// stageGithubRepositoryWriteWave executes the small, self-contained GitHub
// repository wave after the generic pairing proves the baseline lifecycle.
// It is intentionally not part of a --full-only sweep: --write is the opt-in
// for a bounded live mutation, while --full-parity later refuses to claim
// success for every declared action that remains non-live.
func stageGithubRepositoryWriteWave(rc *runContext, rep *Report) error {
	if !rc.opts.Write || rc.opts.Connector != "github" {
		return nil
	}
	if strings.TrimSpace(rc.opts.Config["owner"]) != githubRepositoryWaveFixtureOwner || strings.TrimSpace(rc.opts.Config["repo"]) != githubRepositoryWaveFixtureRepo {
		recordStage(rc, rep, "write_repository_wave", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("write_repository_wave: GitHub live repository certification is restricted to the captain-approved disposable fixture %s/%s", githubRepositoryWaveFixtureOwner, githubRepositoryWaveFixtureRepo)
		})
		return nil
	}

	inventory, err := writeActionInventoryFor("github")
	if err != nil {
		recordStage(rc, rep, "write_repository_wave", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("write_repository_wave: inventory: %v", err)
		})
		return nil
	}
	declarations := make(map[string]writeActionInventoryItem, len(inventory))
	for _, item := range inventory {
		declarations[item.Action] = item
	}
	for _, action := range githubRepositoryWaveActions {
		item, ok := declarations[action]
		if !ok || item.Classification != writeClassificationRepositoryWaveReady {
			recordStage(rc, rep, "write_repository_wave", 2, func() (bool, CLIStageInfo, string) {
				return false, CLIStageInfo{}, fmt.Sprintf("write_repository_wave: action %q is not an approved repository-wave declaration", action)
			})
			return nil
		}
	}

	ledgerRoot := rc.opts.LedgerRoot
	if ledgerRoot == "" {
		ledgerRoot = rc.root
	}
	ledger, err := NewLedger(ledgerRoot)
	if err != nil {
		recordStage(rc, rep, "write_repository_wave", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("write_repository_wave: create ledger: %v", err)
		})
		return nil
	}

	wave := &githubRepositoryWave{
		rc:           rc,
		rep:          rep,
		ledger:       ledger,
		runID:        NewRunID8(),
		declarations: declarations,
		completed:    map[string]bool{},
	}
	if recovered, recoverErr := wave.recoverUncleanedReleaseScenarios(); recoverErr != nil || !recovered {
		reason := "write_repository_reconcile: tagged release recovery failed"
		if recoverErr != nil {
			reason += ": " + recoverErr.Error()
		}
		recordStage(rc, rep, "write_repository_reconcile", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, reason
		})
		return nil
	}
	if recovered, recoverErr := wave.recoverUncleanedCommitCommentScenarios(); recoverErr != nil || !recovered {
		reason := "write_repository_reconcile: tagged commit-comment recovery failed"
		if recoverErr != nil {
			reason += ": " + recoverErr.Error()
		}
		recordStage(rc, rep, "write_repository_reconcile", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, reason
		})
		return nil
	}
	if recovered, recoverErr := wave.recoverUncleanedIssueScenarios(); recoverErr != nil || !recovered {
		reason := "write_repository_reconcile: tagged issue recovery failed"
		if recoverErr != nil {
			reason += ": " + recoverErr.Error()
		}
		recordStage(rc, rep, "write_repository_reconcile", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, reason
		})
		return nil
	}
	if recovered, recoverErr := wave.recoverUncleanedNamedResourceScenarios(); recoverErr != nil || !recovered {
		reason := "write_repository_reconcile: tagged label, milestone, or branch recovery failed"
		if recoverErr != nil {
			reason += ": " + recoverErr.Error()
		}
		recordStage(rc, rep, "write_repository_reconcile", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, reason
		})
		return nil
	}
	if recovered, recoverErr := wave.recoverUncleanedTopicScenarios(); recoverErr != nil || !recovered {
		reason := "write_repository_reconcile: repository-topic baseline recovery failed"
		if recoverErr != nil {
			reason += ": " + recoverErr.Error()
		}
		recordStage(rc, rep, "write_repository_reconcile", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, reason
		})
		return nil
	}

	// Keep independent resources going after one scenario fails. That makes a
	// deliberately broken action visible as a failed certification while each
	// later scenario still proves or disproves its own production path. A
	// selector is intentionally available for bounded resumable live windows:
	// each selected scenario owns and cleans its resource before returning, so
	// a rate-limit restart never needs to replay a create from a prior window.
	ok, selectionErr := wave.runSelected(configValue(rc.opts.Config, "certification_write_wave", ""))
	if selectionErr != nil {
		recordStage(rc, rep, "write_repository_wave", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, "write_repository_wave: " + selectionErr.Error()
		})
		return nil
	}

	recordStage(rc, rep, "write_repository_wave", 2, func() (bool, CLIStageInfo, string) {
		if !ok {
			return false, CLIStageInfo{}, "write_repository_wave: one or more tagged repository scenarios failed; inspect write_wave_* stages and write_actions"
		}
		return true, CLIStageInfo{}, ""
	})
	return nil
}

func (w *githubRepositoryWave) runSelected(selection string) (bool, error) {
	all := []struct {
		name string
		run  func() bool
	}{
		{"issue_lifecycle", w.runIssueLifecycleScenario},
		{"issue_comments", w.runIssueCommentScenario},
		{"issue_lock", w.runIssueLockScenario},
		{"issue_labels", w.runIssueLabelScenario},
		{"label_lifecycle", w.runLabelLifecycleScenario},
		{"milestone", w.runMilestoneScenario},
		{"release", w.runReleaseScenario},
		{"commit_comment", w.runCommitCommentScenario},
		{"ref", w.runRefScenario},
		{"topics", w.runTopicsScenario},
	}
	if selection != "" && selection != "all" {
		if selection == "recover" {
			return true, nil
		}
		for _, candidate := range all {
			if selection == candidate.name {
				return candidate.run(), nil
			}
		}
		return false, fmt.Errorf("unknown certification_write_wave %q (want recover, issue_lifecycle, issue_comments, issue_lock, issue_labels, label_lifecycle, milestone, release, commit_comment, ref, topics, or all)", selection)
	}
	ok := true
	for _, candidate := range all {
		if !candidate.run() {
			ok = false
		}
	}
	return ok, nil
}

// recoverUncleanedReleaseScenarios cleans an interrupted, run-owned release
// before any later wave. A normal published release is rediscovered by its
// tag. The one legacy draft created before tag read-back was introduced needs
// a caller-supplied id, but it is deleted only after the production direct
// read proves that the id still belongs to this ledger's exact ownership tag.
// No raw provider write is used for recovery.
func (w *githubRepositoryWave) recoverUncleanedReleaseScenarios() (bool, error) {
	entries, err := LoadLedger(filepath.Dir(w.ledger.Path()))
	if err != nil {
		return false, fmt.Errorf("load durable write ledger: %w", err)
	}
	pending := make([]LedgerStatus, 0)
	for _, status := range entries.Uncleaned() {
		if status.Connector != "github" || status.Scenario != "github_repository_release" || !strings.HasPrefix(status.Tag, "pm-cert-github-release-") {
			continue
		}
		pending = append(pending, status)
	}
	// The first draft-release experiment predated a durable CLI ledger. Its
	// report records the ownership tag but cannot be rediscovered through the
	// normal ledger. Accept an explicit one-time migration pair only when the
	// direct read below proves that id and tag still name the same run-owned
	// release; it cannot become a generic delete-by-id escape hatch.
	if legacyTag := strings.TrimSpace(configValue(w.rc.opts.Config, "certification_recovery_release_tag", "")); legacyTag != "" {
		if !strings.HasPrefix(legacyTag, "pm-cert-github-release-") {
			return false, fmt.Errorf("certification_recovery_release_tag %q is not a GitHub certification ownership tag", legacyTag)
		}
		if strings.TrimSpace(configValue(w.rc.opts.Config, "certification_recovery_release_id", "")) == "" {
			return false, fmt.Errorf("certification_recovery_release_tag requires certification_recovery_release_id")
		}
		found := false
		for _, status := range pending {
			if status.Tag == legacyTag {
				found = true
				break
			}
		}
		if !found {
			pending = append(pending, LedgerStatus{Tag: legacyTag, Connector: "github", Scenario: "github_repository_release"})
		}
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].Tag < pending[j].Tag })
	for _, status := range pending {
		releaseID, err := w.releaseRecoveryID(status.Tag)
		if err != nil {
			w.recordLeak(status.Tag, []string{"create_release", "update_release", "delete_release"}, "tagged release recovery could not identify the exact run-owned release: "+err.Error())
			return false, err
		}
		if !w.runAction("github_repository_release_recovery", "delete_release_recovery", "delete_release", status.Tag,
			map[string]any{"release_id": releaseID}, fmt.Sprintf("release:%d", releaseID),
			func() (string, error) { return w.verifyReleaseAbsent(releaseID) }) {
			return false, fmt.Errorf("delete recovered tagged release %d", releaseID)
		}
		rows, err := w.readStream("tags")
		if err != nil {
			return false, fmt.Errorf("read tag left by recovered release %d: %w", releaseID, err)
		}
		for _, row := range rows {
			if !valuesEqual(row["name"], status.Tag) {
				continue
			}
			if !w.runAction("github_repository_release_recovery", "delete_release_tag_recovery", "delete_ref", status.Tag,
				map[string]any{"ref": "tags/" + status.Tag}, "tag:"+status.Tag,
				func() (string, error) { return w.verifyStreamAbsent("tags", "name", status.Tag) }) {
				return false, fmt.Errorf("delete recovered tagged release ref %q", status.Tag)
			}
			break
		}
		w.markScenarioPassed(status.Tag, []string{"create_release", "update_release", "delete_release", "delete_ref"}, "interrupted tagged release removed with independent GitHub read-back")
	}
	return true, nil
}

func (w *githubRepositoryWave) releaseRecoveryID(tag string) (int, error) {
	if raw := strings.TrimSpace(configValue(w.rc.opts.Config, "certification_recovery_release_id", "")); raw != "" {
		id, err := strconv.Atoi(raw)
		if err != nil || id <= 0 {
			return 0, fmt.Errorf("certification_recovery_release_id %q must be a positive integer", raw)
		}
		response, err := w.readReleaseResponseByID(id)
		if err != nil {
			return 0, fmt.Errorf("direct read supplied release id %d: %w", id, err)
		}
		actualTag, _ := response["tag_name"].(string)
		if actualTag != tag {
			return 0, fmt.Errorf("direct read supplied release id %d has tag %q, want durable ownership tag %q", id, actualTag, tag)
		}
		return id, nil
	}
	return w.verifyReleaseTag(tag, tag)
}

// recoverUncleanedCommitCommentScenarios deliberately does not use the
// inaccessible item endpoint. The permitted collection stream establishes
// whether a tagged comment still exists; if it does, cleanup goes through the
// same production reverse-plan path and the collection confirms its absence.
// The original mutation remains blocked, never pass, because the item-level
// read-back required for its certification was refused.
func (w *githubRepositoryWave) recoverUncleanedCommitCommentScenarios() (bool, error) {
	entries, err := LoadLedger(filepath.Dir(w.ledger.Path()))
	if err != nil {
		return false, fmt.Errorf("load durable write ledger: %w", err)
	}
	for _, status := range entries.Uncleaned() {
		if status.Connector != "github" || status.Scenario != "github_repository_commit_comment" || !strings.HasPrefix(status.Tag, "pm-cert-github-commit-comment-") {
			continue
		}
		commentID, err := commitCommentIDFromLedger(status)
		if err != nil {
			w.recordLeak(status.Tag, githubCommitCommentActions, "tagged commit-comment recovery could not determine its provider id: "+err.Error())
			return false, err
		}
		present, err := w.commitCommentPresent(commentID)
		if err != nil {
			w.recordLeak(status.Tag, githubCommitCommentActions, "tagged commit-comment cleanup could not be independently checked through the permitted collection stream: "+err.Error())
			return false, err
		}
		if present {
			if !w.runAction("github_repository_commit_comment_recovery", "delete_commit_comment_recovery", "delete_commit_comment", status.Tag,
				map[string]any{"comment_id": commentID}, fmt.Sprintf("commit_comment:%d", commentID),
				func() (string, error) { return w.verifyStreamAbsent("commit_comments", "id", commentID) }) {
				return false, fmt.Errorf("delete recovered tagged commit comment %d", commentID)
			}
		}
		if err := w.ledger.RecordReadBack(status.Tag, fmt.Sprintf("commit_comment:%d:absent", commentID)); err != nil {
			return false, fmt.Errorf("checkpoint recovered commit comment %d absence: %w", commentID, err)
		}
		if err := w.ledger.RecordCleaned(status.Tag); err != nil {
			return false, fmt.Errorf("record cleanup for recovered commit comment %d: %w", commentID, err)
		}
		w.markBlocked(githubCommitCommentActions, status.Tag, githubCommitCommentItemReadBlockedReason, "verified")
	}
	return true, nil
}

func commitCommentIDFromLedger(status LedgerStatus) (int, error) {
	for _, raw := range []string{status.ResourceID, status.EntityHint} {
		idText, ok := strings.CutPrefix(raw, "commit_comment:")
		if !ok {
			continue
		}
		id, err := strconv.Atoi(idText)
		if err == nil && id > 0 {
			return id, nil
		}
	}
	return 0, fmt.Errorf("ledger resource %q / entity hint %q is not a positive commit_comment id", status.ResourceID, status.EntityHint)
}

func (w *githubRepositoryWave) commitCommentPresent(commentID int) (bool, error) {
	rows, err := w.readStream("commit_comments")
	if err != nil {
		return false, err
	}
	for _, row := range rows {
		if valuesEqual(row["id"], commentID) {
			return true, nil
		}
	}
	return false, nil
}

func (w *githubRepositoryWave) tag(scenario string) string {
	return NewTag("github-"+scenario, w.runID)
}

// recoverUncleanedIssueScenarios reconciles an interrupted issue scenario
// before a later wave starts. The durable ledger names only resources with
// our pm-cert ownership marker; recovery reads the repository independently,
// removes any tagged comment, and closes the tagged issue through the same
// reverse-plan path. It never guesses at an untagged or third-party issue.
func (w *githubRepositoryWave) recoverUncleanedIssueScenarios() (bool, error) {
	entries, err := LoadLedger(filepath.Dir(w.ledger.Path()))
	if err != nil {
		return false, fmt.Errorf("load durable write ledger: %w", err)
	}
	pending := make([]LedgerStatus, 0)
	for _, status := range entries.Uncleaned() {
		if status.Connector == "github" && strings.HasPrefix(status.Tag, "pm-cert-github-") && isRecoverableIssueScenario(status.Scenario) {
			pending = append(pending, status)
		}
	}
	if len(pending) == 0 {
		return true, nil
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].Tag < pending[j].Tag })
	rows, err := w.readStream("issues")
	if err != nil {
		return false, fmt.Errorf("read tagged issues for recovery: %w", err)
	}
	needsCommentRead := false
	for _, status := range pending {
		if status.Scenario == "github_repository_issue_comments" {
			needsCommentRead = true
			break
		}
	}
	var comments []map[string]any
	if needsCommentRead {
		comments, err = w.readStream("issue_comments")
		if err != nil {
			return false, fmt.Errorf("read tagged issue comments for recovery: %w", err)
		}
	}
	for _, status := range pending {
		var found map[string]any
		for _, row := range rows {
			if title, _ := row["title"].(string); strings.HasPrefix(title, status.Tag) {
				found = row
				break
			}
		}
		if found == nil {
			reason := "durable create_issue entry was not visible in independent GitHub read-back; refusing to guess at cleanup"
			w.recordLeak(status.Tag, []string{"create_issue"}, reason)
			return false, fmt.Errorf("%s: %s", status.Tag, reason)
		}
		number, err := integerField(found, "number")
		if err != nil {
			w.recordLeak(status.Tag, []string{"create_issue"}, "tagged issue recovery could not determine its GitHub number: "+err.Error())
			return false, err
		}
		resourceID := fmt.Sprintf("issue:%d", number)
		if status.Scenario == "github_repository_issue_comments" {
			for _, comment := range comments {
				body, _ := comment["body"].(string)
				if !strings.HasPrefix(body, status.Tag) {
					continue
				}
				commentID, err := integerField(comment, "id")
				if err != nil {
					w.recordLeak(status.Tag, []string{"comment_issue"}, "tagged issue-comment recovery could not determine its GitHub comment id: "+err.Error())
					return false, err
				}
				if !w.runAction("github_repository_issue_recovery", "delete_issue_comment_recovery", "delete_issue_comment", status.Tag,
					map[string]any{"comment_id": commentID}, fmt.Sprintf("issue_comment:%d", commentID),
					func() (string, error) { return w.verifyIssueCommentAbsent(commentID) }) {
					return false, fmt.Errorf("delete recovered tagged issue comment %d", commentID)
				}
			}
		}
		state, _ := found["state"].(string)
		if state == "closed" {
			if err := w.ledger.RecordReadBack(status.Tag, resourceID); err != nil {
				return false, fmt.Errorf("checkpoint already-closed tagged issue %d: %w", number, err)
			}
			if err := w.ledger.RecordCleaned(status.Tag); err != nil {
				return false, fmt.Errorf("record cleanup for already-closed tagged issue %d: %w", number, err)
			}
			continue
		}
		if !w.runAction("github_repository_issue_recovery", "close_issue_recovery", "close_issue", status.Tag,
			map[string]any{"issue_number": number}, resourceID,
			func() (string, error) { return w.verifyIssueState(number, "closed") }) {
			return false, fmt.Errorf("close recovered tagged issue %d", number)
		}
		w.markScenarioPassed(status.Tag, []string{"close_issue"}, "interrupted tagged issue closed with independent GitHub read-back")
	}
	return true, nil
}

func isRecoverableIssueScenario(scenario string) bool {
	switch scenario {
	case "github_repository_issue_lifecycle", "github_repository_issue_comments", "github_repository_issue_lock", "github_repository_issue_labels":
		return true
	default:
		return false
	}
}

// recoverUncleanedNamedResourceScenarios reconciles the wave resources whose
// provider lists carry a stable, run-owned name. It never broadens a delete:
// every candidate must begin with the exact ledger tag and is removed through
// the normal reverse plan/preview/run path before its absence is independently
// read back. A recovered interrupted scenario is deliberately left not_live;
// cleanup is not retrospective proof of the mutation that was interrupted.
func (w *githubRepositoryWave) recoverUncleanedNamedResourceScenarios() (bool, error) {
	entries, err := LoadLedger(filepath.Dir(w.ledger.Path()))
	if err != nil {
		return false, fmt.Errorf("load durable write ledger: %w", err)
	}
	pending := make([]LedgerStatus, 0)
	for _, status := range entries.Uncleaned() {
		if status.Connector != "github" || !strings.HasPrefix(status.Tag, "pm-cert-github-") {
			continue
		}
		switch status.Scenario {
		case "github_repository_label_lifecycle", "github_repository_milestone", "github_repository_ref":
			pending = append(pending, status)
		}
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].Tag < pending[j].Tag })
	for _, status := range pending {
		var actions []string
		var recoveryErr error
		switch status.Scenario {
		case "github_repository_label_lifecycle":
			actions, recoveryErr = w.recoverTaggedLabels(status)
		case "github_repository_milestone":
			actions, recoveryErr = w.recoverTaggedMilestones(status)
		case "github_repository_ref":
			actions, recoveryErr = w.recoverTaggedRefs(status)
		}
		if recoveryErr != nil {
			w.recordLeak(status.Tag, actions, recoveryErr.Error())
			return false, recoveryErr
		}
		if err := w.ledger.RecordReadBack(status.Tag, "recovery:"+safeWaveName(status.Scenario)+":absent"); err != nil {
			return false, fmt.Errorf("checkpoint recovered %s %q absence: %w", status.Scenario, status.Tag, err)
		}
		if err := w.ledger.RecordCleaned(status.Tag); err != nil {
			return false, fmt.Errorf("record cleanup for recovered %s %q: %w", status.Scenario, status.Tag, err)
		}
		w.markRecoveryUncertified(actions, status.Tag, "interrupted tagged resource was removed with independent GitHub read-back; the interrupted mutation is not certified")
	}
	return true, nil
}

func (w *githubRepositoryWave) recoverTaggedLabels(status LedgerStatus) ([]string, error) {
	actions := []string{"create_label", "update_label", "delete_label"}
	rows, err := w.readStream("labels")
	if err != nil {
		return actions, fmt.Errorf("read tagged labels for recovery: %w", err)
	}
	for _, row := range rows {
		name, _ := row["name"].(string)
		if !strings.HasPrefix(name, status.Tag) {
			continue
		}
		if !w.runAction("github_repository_label_recovery", "delete_label_recovery", "delete_label", status.Tag,
			map[string]any{"name": name}, "label:"+name,
			func() (string, error) { return w.verifyStreamAbsent("labels", "name", name) }) {
			return actions, fmt.Errorf("delete recovered tagged label %q", name)
		}
	}
	return actions, nil
}

func (w *githubRepositoryWave) recoverTaggedMilestones(status LedgerStatus) ([]string, error) {
	actions := []string{"create_milestone", "update_milestone", "delete_milestone"}
	rows, err := w.readStream("milestones")
	if err != nil {
		return actions, fmt.Errorf("read tagged milestones for recovery: %w", err)
	}
	for _, row := range rows {
		title, _ := row["title"].(string)
		if !strings.HasPrefix(title, status.Tag) {
			continue
		}
		number, err := integerField(row, "number")
		if err != nil {
			return actions, fmt.Errorf("tagged milestone %q has no number: %w", title, err)
		}
		if !w.runAction("github_repository_milestone_recovery", "delete_milestone_recovery", "delete_milestone", status.Tag,
			map[string]any{"milestone_number": number}, fmt.Sprintf("milestone:%d", number),
			func() (string, error) { return w.verifyStreamAbsent("milestones", "number", number) }) {
			return actions, fmt.Errorf("delete recovered tagged milestone %d", number)
		}
	}
	return actions, nil
}

func (w *githubRepositoryWave) recoverTaggedRefs(status LedgerStatus) ([]string, error) {
	actions := []string{"create_ref", "update_ref", "delete_ref"}
	rows, err := w.readStream("branches")
	if err != nil {
		return actions, fmt.Errorf("read tagged branches for recovery: %w", err)
	}
	for _, row := range rows {
		name, _ := row["name"].(string)
		if !strings.HasPrefix(name, status.Tag) {
			continue
		}
		if !w.runAction("github_repository_ref_recovery", "delete_ref_recovery", "delete_ref", status.Tag,
			map[string]any{"ref": "heads/" + name}, "ref:heads/"+name,
			func() (string, error) { return w.verifyStreamAbsent("branches", "name", name) }) {
			return actions, fmt.Errorf("delete recovered tagged branch %q", name)
		}
	}
	return actions, nil
}

type topicRecoveryBaseline struct {
	Baseline []string `json:"baseline"`
}

func (w *githubRepositoryWave) recoverUncleanedTopicScenarios() (bool, error) {
	entries, err := LoadLedger(filepath.Dir(w.ledger.Path()))
	if err != nil {
		return false, fmt.Errorf("load durable write ledger: %w", err)
	}
	for _, status := range entries.Uncleaned() {
		if status.Connector != "github" || status.Scenario != "github_repository_topics" || !strings.HasPrefix(status.Tag, "pm-cert-github-topics-") {
			continue
		}
		var recovery topicRecoveryBaseline
		if len(status.Recovery) == 0 || json.Unmarshal(status.Recovery, &recovery) != nil {
			return false, fmt.Errorf("tagged topic recovery %q has no valid captured baseline", status.Tag)
		}
		for _, topic := range recovery.Baseline {
			if strings.TrimSpace(topic) == "" {
				return false, fmt.Errorf("tagged topic recovery %q has an invalid empty baseline topic", status.Tag)
			}
		}
		names, err := w.readTopics()
		if err != nil {
			return false, fmt.Errorf("read repository topics for recovery %q: %w", status.Tag, err)
		}
		if sameStrings(names, recovery.Baseline) {
			if err := w.ledger.RecordReadBack(status.Tag, "topics:baseline_restored"); err != nil {
				return false, fmt.Errorf("checkpoint unchanged topics baseline %q: %w", status.Tag, err)
			}
			if err := w.ledger.RecordCleaned(status.Tag); err != nil {
				return false, fmt.Errorf("record unchanged topics cleanup %q: %w", status.Tag, err)
			}
			w.markRecoveryUncertified([]string{"replace_repo_topics"}, status.Tag, "interrupted topic scenario had no visible mutation; it is not certified")
			continue
		}
		if !sameStrings(names, []string{status.Tag}) {
			reason := fmt.Sprintf("tagged topic recovery %q found topics other than its exact tag or captured baseline; refusing to overwrite an unknown provider state", status.Tag)
			w.recordLeak(status.Tag, []string{"replace_repo_topics"}, reason)
			return false, fmt.Errorf("%s", reason)
		}
		if !w.runAction("github_repository_topics_recovery", "replace_repo_topics_restore_recovery", "replace_repo_topics", status.Tag,
			map[string]any{"names": recovery.Baseline}, "topics:"+status.Tag,
			func() (string, error) {
				current, err := w.readTopics()
				if err != nil {
					return "", err
				}
				if !sameStrings(current, recovery.Baseline) {
					return "", fmt.Errorf("repository topics do not equal the captured baseline after recovery restore")
				}
				return "topics:baseline_restored", nil
			}) {
			return false, fmt.Errorf("restore recovered repository topics %q", status.Tag)
		}
		if err := w.ledger.RecordCleaned(status.Tag); err != nil {
			return false, fmt.Errorf("record recovered topics cleanup %q: %w", status.Tag, err)
		}
		w.markRecoveryUncertified([]string{"replace_repo_topics"}, status.Tag, "interrupted tagged topic mutation was restored with independent GitHub read-back; it is not certified")
	}
	return true, nil
}

func (w *githubRepositoryWave) recordTopicBaseline(scenario, tag string, baseline []string) error {
	recovery, err := json.Marshal(topicRecoveryBaseline{Baseline: append([]string(nil), baseline...)})
	if err != nil {
		return fmt.Errorf("encode captured topic baseline: %w", err)
	}
	return w.ledger.RecordPlanned(LedgerEntry{
		Action:     "replace_repo_topics",
		Scenario:   scenario,
		Tag:        tag,
		Connector:  "github",
		EntityHint: "topics:" + tag,
		ResourceID: "topics:" + tag,
		Recovery:   recovery,
	})
}

func (w *githubRepositoryWave) runIssueLifecycleScenario() bool {
	tag := w.tag("issue-lifecycle")
	const scenario = "github_repository_issue_lifecycle"
	actions := []string{"create_issue", "update_issue", "close_issue", "reopen_issue"}
	updated := tag + " updated"
	var number int
	created := false
	cleanup := func() bool {
		if !created {
			return true
		}
		clean := w.runAction(scenario, "close_issue_cleanup", "close_issue", tag,
			map[string]any{"issue_number": number}, fmt.Sprintf("issue:%d", number),
			func() (string, error) { return w.verifyIssueState(number, "closed") })
		if !clean {
			w.recordLeak(tag, actions, "tagged issue could not be closed for cleanup")
			return false
		}
		w.markScenarioPassed(tag, actions, "tagged issue closed with independent GitHub read-back")
		return true
	}
	if !w.runAction(scenario, "create_issue", "create_issue", tag,
		map[string]any{"title": tag, "body": tag}, "issue:"+tag,
		func() (string, error) {
			row, _, err := w.verifyStreamPresent("issues", "title", tag, "title", tag)
			if err != nil {
				return "", err
			}
			number, err = integerField(row, "number")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("issue:%d", number), nil
		}) {
		return cleanup() && false
	}
	created = true
	if !w.runAction(scenario, "update_issue", "update_issue", tag,
		map[string]any{"issue_number": number, "title": updated, "body": updated}, fmt.Sprintf("issue:%d", number),
		func() (string, error) {
			_, id, err := w.verifyStreamPresent("issues", "number", number, "title", updated)
			return id, err
		}) {
		return cleanup() && false
	}
	if !w.runAction(scenario, "close_issue", "close_issue", tag,
		map[string]any{"issue_number": number}, fmt.Sprintf("issue:%d", number),
		func() (string, error) { return w.verifyIssueState(number, "closed") }) {
		return cleanup() && false
	}
	if !w.runAction(scenario, "reopen_issue", "reopen_issue", tag,
		map[string]any{"issue_number": number}, fmt.Sprintf("issue:%d", number),
		func() (string, error) { return w.verifyIssueState(number, "open") }) {
		return cleanup() && false
	}
	return cleanup()
}

func (w *githubRepositoryWave) runIssueCommentScenario() bool {
	tag := w.tag("issue-comment")
	const scenario = "github_repository_issue_comments"
	actions := []string{"create_issue", "comment_issue", "update_issue_comment", "delete_issue_comment", "close_issue"}
	body := tag + " comment"
	updated := body + " updated"
	var number, commentID int
	var commentUpdatedAt string
	issueCreated := false
	commentAlive := false
	cleanup := func() bool {
		clean := true
		if commentAlive {
			if !w.runAction(scenario, "delete_issue_comment_cleanup", "delete_issue_comment", tag,
				map[string]any{"comment_id": commentID}, fmt.Sprintf("issue_comment:%d", commentID),
				func() (string, error) { return w.verifyIssueCommentAbsent(commentID) }) {
				clean = false
			}
		}
		if issueCreated {
			if !w.runAction(scenario, "close_issue_cleanup", "close_issue", tag,
				map[string]any{"issue_number": number}, fmt.Sprintf("issue:%d", number),
				func() (string, error) { return w.verifyIssueState(number, "closed") }) {
				clean = false
			}
		}
		if !clean {
			w.recordLeak(tag, actions, "tagged issue comment or its containing issue remained without verified cleanup")
			return false
		}
		if issueCreated {
			w.markScenarioPassed(tag, actions, "tagged issue comment removed and containing issue closed with independent GitHub read-back")
		}
		return true
	}
	if !w.runAction(scenario, "create_issue", "create_issue", tag,
		map[string]any{"title": tag, "body": tag}, "issue:"+tag,
		func() (string, error) {
			row, _, err := w.verifyStreamPresent("issues", "title", tag, "title", tag)
			if err != nil {
				return "", err
			}
			number, err = integerField(row, "number")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("issue:%d", number), nil
		}) {
		return cleanup() && false
	}
	issueCreated = true
	if !w.runAction(scenario, "comment_issue", "comment_issue", tag,
		map[string]any{"issue_number": number, "body": body}, fmt.Sprintf("issue:%d", number),
		func() (string, error) {
			row, _, err := w.verifyStreamPresent("issue_comments", "body", body, "body", body)
			if err != nil {
				return "", err
			}
			commentID, err = integerField(row, "id")
			if err != nil {
				return "", err
			}
			commentUpdatedAt, _ = row["updated_at"].(string)
			if commentUpdatedAt == "" {
				return "", fmt.Errorf("issue comment %d read-back omitted updated_at", commentID)
			}
			return fmt.Sprintf("issue_comment:%d", commentID), nil
		}) {
		return cleanup() && false
	}
	commentAlive = true
	if !w.runAction(scenario, "update_issue_comment", "update_issue_comment", tag,
		map[string]any{"comment_id": commentID, "body": updated}, fmt.Sprintf("issue_comment:%d", commentID),
		func() (string, error) {
			return w.verifyIssueCommentUpdated(commentID, commentUpdatedAt)
		}) {
		return cleanup() && false
	}
	if !w.runAction(scenario, "delete_issue_comment", "delete_issue_comment", tag,
		map[string]any{"comment_id": commentID}, fmt.Sprintf("issue_comment:%d", commentID),
		func() (string, error) { return w.verifyIssueCommentAbsent(commentID) }) {
		return cleanup() && false
	}
	commentAlive = false
	return cleanup()
}

func (w *githubRepositoryWave) runIssueLockScenario() bool {
	tag := w.tag("issue-lock")
	const scenario = "github_repository_issue_lock"
	actions := []string{"create_issue", "lock_issue", "unlock_issue", "close_issue"}
	var number int
	created := false
	locked := false
	cleanup := func() bool {
		clean := true
		if locked {
			if !w.runAction(scenario, "unlock_issue_cleanup", "unlock_issue", tag,
				map[string]any{"issue_number": number}, fmt.Sprintf("issue:%d", number),
				func() (string, error) { return w.verifyIssueLocked(number, false) }) {
				clean = false
			}
		}
		if created {
			if !w.runAction(scenario, "close_issue_cleanup", "close_issue", tag,
				map[string]any{"issue_number": number}, fmt.Sprintf("issue:%d", number),
				func() (string, error) { return w.verifyIssueState(number, "closed") }) {
				clean = false
			}
		}
		if !clean {
			w.recordLeak(tag, actions, "tagged issue lock state or issue cleanup could not be verified")
			return false
		}
		if created {
			w.markScenarioPassed(tag, actions, "tagged issue unlocked and closed with independent GitHub read-back")
		}
		return true
	}
	if !w.runAction(scenario, "create_issue", "create_issue", tag,
		map[string]any{"title": tag, "body": tag}, "issue:"+tag,
		func() (string, error) {
			row, _, err := w.verifyStreamPresent("issues", "title", tag, "title", tag)
			if err != nil {
				return "", err
			}
			number, err = integerField(row, "number")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("issue:%d", number), nil
		}) {
		return cleanup() && false
	}
	created = true
	if !w.runAction(scenario, "lock_issue", "lock_issue", tag,
		map[string]any{"issue_number": number, "lock_reason": "resolved"}, fmt.Sprintf("issue:%d", number),
		func() (string, error) { return w.verifyIssueLocked(number, true) }) {
		return cleanup() && false
	}
	locked = true
	if !w.runAction(scenario, "unlock_issue", "unlock_issue", tag,
		map[string]any{"issue_number": number}, fmt.Sprintf("issue:%d", number),
		func() (string, error) { return w.verifyIssueLocked(number, false) }) {
		return cleanup() && false
	}
	locked = false
	return cleanup()
}

func (w *githubRepositoryWave) runIssueLabelScenario() bool {
	tag := w.tag("issue-labels")
	const scenario = "github_repository_issue_labels"
	actions := []string{"create_issue", "set_issue_labels", "add_issue_labels", "remove_issue_label", "close_issue"}
	labels, err := w.fixtureLabels(2)
	if err != nil {
		w.recordScenarioFailure(scenario, actions, tag, "read two fixture-label baseline values: "+err.Error())
		return false
	}
	var number int
	created := false
	labelsTouched := false
	cleanup := func() bool {
		clean := true
		if labelsTouched {
			if !w.runAction(scenario, "set_issue_labels_restore", "set_issue_labels", tag,
				map[string]any{"issue_number": number, "labels": []string{}}, fmt.Sprintf("issue:%d", number),
				func() (string, error) { return w.verifyIssueNoLabels(number) }) {
				clean = false
			}
		}
		if created {
			if !w.runAction(scenario, "close_issue_cleanup", "close_issue", tag,
				map[string]any{"issue_number": number}, fmt.Sprintf("issue:%d", number),
				func() (string, error) { return w.verifyIssueState(number, "closed") }) {
				clean = false
			}
		}
		if !clean {
			w.recordLeak(tag, actions, "tagged issue labels were not restored before close")
			return false
		}
		if created {
			w.markScenarioPassed(tag, actions, "tagged issue labels restored to empty and issue closed with independent GitHub read-back")
		}
		return true
	}
	if !w.runAction(scenario, "create_issue", "create_issue", tag,
		map[string]any{"title": tag, "body": tag}, "issue:"+tag,
		func() (string, error) {
			row, _, err := w.verifyStreamPresent("issues", "title", tag, "title", tag)
			if err != nil {
				return "", err
			}
			number, err = integerField(row, "number")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("issue:%d", number), nil
		}) {
		return cleanup() && false
	}
	created = true
	if !w.runAction(scenario, "set_issue_labels", "set_issue_labels", tag,
		map[string]any{"issue_number": number, "labels": []string{labels[0]}}, fmt.Sprintf("issue:%d", number),
		func() (string, error) { return w.verifyIssueLabels(number, labels[0]) }) {
		return cleanup() && false
	}
	labelsTouched = true
	if !w.runAction(scenario, "add_issue_labels", "add_issue_labels", tag,
		map[string]any{"issue_number": number, "labels": []string{labels[1]}}, fmt.Sprintf("issue:%d", number),
		func() (string, error) { return w.verifyIssueLabels(number, labels[0], labels[1]) }) {
		return cleanup() && false
	}
	if !w.runAction(scenario, "remove_issue_label", "remove_issue_label", tag,
		map[string]any{"issue_number": number, "name": labels[1]}, fmt.Sprintf("issue:%d", number),
		func() (string, error) { return w.verifyIssueLabelAbsent(number, labels[1]) }) {
		return cleanup() && false
	}
	return cleanup()
}

func (w *githubRepositoryWave) runLabelLifecycleScenario() bool {
	tag := w.tag("label")
	const scenario = "github_repository_label_lifecycle"
	actions := []string{"create_label", "update_label", "delete_label"}
	first, updated, second := githubWaveLabelNames(tag)
	firstCreated := false
	secondCreated := false
	firstCurrent := first
	cleanup := func() bool {
		clean := true
		if secondCreated {
			if !w.runAction(scenario, "delete_label_second", "delete_label", tag,
				map[string]any{"name": second}, "label:"+second,
				func() (string, error) { return w.verifyStreamAbsent("labels", "name", second) }) {
				clean = false
			}
		}
		if firstCreated {
			if !w.runAction(scenario, "delete_label_first", "delete_label", tag,
				map[string]any{"name": firstCurrent}, "label:"+firstCurrent,
				func() (string, error) { return w.verifyStreamAbsent("labels", "name", firstCurrent) }) {
				clean = false
			}
		}
		if !clean {
			w.recordLeak(tag, actions, "tagged labels remained after cleanup")
			return false
		}
		if firstCreated || secondCreated {
			w.markScenarioPassed(tag, actions, "tagged labels removed with independent GitHub read-back")
		}
		return true
	}
	if !w.runAction(scenario, "create_label_first", "create_label", tag,
		map[string]any{"name": first, "color": "ededed", "description": tag}, "label:"+first,
		func() (string, error) {
			_, id, err := w.verifyStreamPresent("labels", "name", first, "name", first)
			return id, err
		}) {
		return cleanup() && false
	}
	firstCreated = true
	if !w.runAction(scenario, "create_label_second", "create_label", tag,
		map[string]any{"name": second, "color": "ededed", "description": tag}, "label:"+second,
		func() (string, error) {
			_, id, err := w.verifyStreamPresent("labels", "name", second, "name", second)
			return id, err
		}) {
		return cleanup() && false
	}
	secondCreated = true
	if !w.runAction(scenario, "update_label", "update_label", tag,
		map[string]any{"name": first, "new_name": updated, "color": "ededed", "description": updated}, "label:"+updated,
		func() (string, error) {
			_, id, err := w.verifyStreamPresent("labels", "name", updated, "name", updated)
			return id, err
		}) {
		return cleanup() && false
	}
	firstCurrent = updated
	return cleanup()
}

// githubWaveLabelNames leaves room under GitHub's 50-character label-name
// ceiling. NewTag already carries the full run ownership marker, so compact
// distinct suffixes preserve cleanup safety without turning an update into a
// provider-side 422.
func githubWaveLabelNames(tag string) (first, updated, second string) {
	return tag + "-a", tag + "-b", tag + "-c"
}

func (w *githubRepositoryWave) runMilestoneScenario() bool {
	tag := w.tag("milestone")
	const scenario = "github_repository_milestone"
	actions := []string{"create_milestone", "update_milestone", "delete_milestone"}
	updated := tag + " updated"
	var number int
	created := false
	cleanup := func() bool {
		if !created {
			return true
		}
		clean := w.runAction(scenario, "delete_milestone", "delete_milestone", tag,
			map[string]any{"milestone_number": number}, fmt.Sprintf("milestone:%d", number),
			func() (string, error) { return w.verifyStreamAbsent("milestones", "number", number) })
		if !clean {
			w.recordLeak(tag, actions, "tagged milestone remained after cleanup")
			return false
		}
		w.markScenarioPassed(tag, actions, "tagged milestone removed with independent GitHub read-back")
		return true
	}
	if !w.runAction(scenario, "create_milestone", "create_milestone", tag,
		map[string]any{"title": tag, "description": tag}, "milestone:"+tag,
		func() (string, error) {
			row, _, err := w.verifyStreamPresent("milestones", "title", tag, "title", tag)
			if err != nil {
				return "", err
			}
			number, err = integerField(row, "number")
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("milestone:%d", number), nil
		}) {
		return cleanup() && false
	}
	created = true
	if !w.runAction(scenario, "update_milestone", "update_milestone", tag,
		map[string]any{"milestone_number": number, "title": updated, "description": tag + " updated"}, fmt.Sprintf("milestone:%d", number),
		func() (string, error) {
			_, id, err := w.verifyStreamPresent("milestones", "number", number, "title", updated)
			return id, err
		}) {
		return cleanup() && false
	}
	return cleanup()
}

func (w *githubRepositoryWave) runReleaseScenario() bool {
	tag := w.tag("release")
	const scenario = "github_repository_release"
	actions := []string{"create_release", "update_release", "delete_release", "delete_ref"}
	updated := tag + " updated"
	var releaseID int
	created := false
	deletedRelease := false
	cleanup := func() bool {
		clean := true
		if created && !deletedRelease {
			if !w.runAction(scenario, "delete_release_cleanup", "delete_release", tag,
				map[string]any{"release_id": releaseID}, fmt.Sprintf("release:%d", releaseID),
				func() (string, error) { return w.verifyStreamAbsent("releases", "id", releaseID) }) {
				clean = false
			}
		}
		// GitHub's delete-release endpoint deliberately leaves its tag behind.
		// Deleting that run-owned tag via the declared delete_ref action is part
		// of cleanup, not an untracked shell/API side effect.
		if created {
			if !w.runAction(scenario, "delete_ref_release_tag", "delete_ref", tag,
				map[string]any{"ref": "tags/" + tag}, "tag:"+tag,
				func() (string, error) { return w.verifyStreamAbsent("tags", "name", tag) }) {
				clean = false
			}
		}
		if !clean {
			w.recordLeak(tag, actions, "tagged release or its release-created tag remained after cleanup")
			return false
		}
		if created {
			w.markScenarioPassed(tag, actions, "tagged release deleted and its release-created tag removed with independent GitHub read-back")
		}
		return true
	}
	if !w.runAction(scenario, "create_release", "create_release", tag,
		map[string]any{"tag_name": tag, "name": tag, "body": tag, "draft": false}, "release:"+tag,
		func() (string, error) {
			var err error
			releaseID, err = w.verifyReleaseTag(tag, tag)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("release:%d", releaseID), nil
		}) {
		return cleanup() && false
	}
	created = true
	if !w.runAction(scenario, "update_release", "update_release", tag,
		map[string]any{"release_id": releaseID, "name": updated, "body": tag + " updated", "draft": false}, fmt.Sprintf("release:%d", releaseID),
		func() (string, error) {
			id, err := w.verifyReleaseByID(releaseID, updated)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("release:%d", id), nil
		}) {
		return cleanup() && false
	}
	if !w.runAction(scenario, "delete_release", "delete_release", tag,
		map[string]any{"release_id": releaseID}, fmt.Sprintf("release:%d", releaseID),
		func() (string, error) { return w.verifyReleaseAbsent(releaseID) }) {
		return cleanup() && false
	}
	deletedRelease = true
	return cleanup()
}

func (w *githubRepositoryWave) runCommitCommentScenario() bool {
	tag := w.tag("commit-comment")
	const scenario = "github_repository_commit_comment"
	actions := githubCommitCommentActions
	// A create/update/delete response cannot stand in for the item-level
	// read-back. The current disposable token has been observed refusing that
	// read, so do not keep creating a comment that cannot be certified. After a
	// browser permission repair the caller opts in to re-attempting this bounded
	// scenario; a provider refusal still leaves it blocked and non-pass.
	if configValue(w.rc.opts.Config, "certification_commit_comment_item_read", "") != "enabled" {
		w.markBlocked(actions, tag, githubCommitCommentItemReadBlockedReason, "not_attempted")
		recordStage(w.rc, w.rep, "write_wave_commit_comment", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, githubCommitCommentItemReadBlockedReason
		})
		return false
	}
	sha, err := w.firstCommitSHA()
	if err != nil {
		w.recordScenarioFailure(scenario, actions, tag, "read fixture commit for commit-comment scenario: "+err.Error())
		return false
	}
	body := tag + " comment"
	updated := body + " updated"
	var commentID int
	var commentUpdatedAt string
	created := false
	commentAlive := false
	cleanup := func() bool {
		if !created || !commentAlive {
			return true
		}
		clean := w.runAction(scenario, "delete_commit_comment", "delete_commit_comment", tag,
			map[string]any{"comment_id": commentID}, fmt.Sprintf("commit_comment:%d", commentID),
			func() (string, error) { return w.verifyCommitCommentAbsent(commentID) })
		if !clean {
			w.recordLeak(tag, actions, "tagged commit comment remained after cleanup")
			return false
		}
		w.markScenarioPassed(tag, actions, "tagged commit comment removed with independent GitHub read-back")
		return true
	}
	if !w.runAction(scenario, "create_commit_comment", "create_commit_comment", tag,
		map[string]any{"commit_sha": sha, "body": body}, "commit_comment:"+tag,
		func() (string, error) {
			row, _, err := w.verifyStreamPresent("commit_comments", "body", body, "body", body)
			if err != nil {
				return "", err
			}
			commentID, err = integerField(row, "id")
			if err != nil {
				return "", err
			}
			commentUpdatedAt, _ = row["updated_at"].(string)
			if commentUpdatedAt == "" {
				return "", fmt.Errorf("commit comment %d read-back omitted updated_at", commentID)
			}
			return fmt.Sprintf("commit_comment:%d", commentID), nil
		}) {
		return cleanup() && false
	}
	created = true
	commentAlive = true
	if !w.runAction(scenario, "update_commit_comment", "update_commit_comment", tag,
		map[string]any{"comment_id": commentID, "body": updated}, fmt.Sprintf("commit_comment:%d", commentID),
		func() (string, error) {
			return w.verifyCommitCommentUpdated(commentID, commentUpdatedAt)
		}) {
		return cleanup() && false
	}
	if !w.runAction(scenario, "delete_commit_comment", "delete_commit_comment", tag,
		map[string]any{"comment_id": commentID}, fmt.Sprintf("commit_comment:%d", commentID),
		func() (string, error) { return w.verifyCommitCommentAbsent(commentID) }) {
		return cleanup() && false
	}
	commentAlive = false
	w.markScenarioPassed(tag, actions, "tagged commit comment removed with independent GitHub read-back")
	return true
}

func (w *githubRepositoryWave) runRefScenario() bool {
	tag := w.tag("ref")
	const scenario = "github_repository_ref"
	actions := []string{"create_ref", "update_ref", "delete_ref"}
	sha, err := w.firstCommitSHA()
	if err != nil {
		w.recordScenarioFailure(scenario, actions, tag, "read fixture commit for ref scenario: "+err.Error())
		return false
	}
	branch := tag + "-branch"
	created := false
	cleanup := func() bool {
		if !created {
			return true
		}
		clean := w.runAction(scenario, "delete_ref", "delete_ref", tag,
			map[string]any{"ref": "heads/" + branch}, "ref:heads/"+branch,
			func() (string, error) { return w.verifyStreamAbsent("branches", "name", branch) })
		if !clean {
			w.recordLeak(tag, actions, "tagged branch ref remained after cleanup")
			return false
		}
		w.markScenarioPassed(tag, actions, "tagged branch ref removed with independent GitHub read-back")
		return true
	}
	if !w.runAction(scenario, "create_ref", "create_ref", tag,
		map[string]any{"ref": "refs/heads/" + branch, "sha": sha}, "ref:heads/"+branch,
		func() (string, error) {
			_, id, err := w.verifyStreamPresent("branches", "name", branch, "commit_sha", sha)
			return id, err
		}) {
		return cleanup() && false
	}
	created = true
	if !w.runAction(scenario, "update_ref", "update_ref", tag,
		map[string]any{"ref": "heads/" + branch, "sha": sha, "force": false}, "ref:heads/"+branch,
		func() (string, error) {
			_, id, err := w.verifyStreamPresent("branches", "name", branch, "commit_sha", sha)
			return id, err
		}) {
		return cleanup() && false
	}
	return cleanup()
}

func (w *githubRepositoryWave) runTopicsScenario() bool {
	tag := w.tag("topics")
	const scenario = "github_repository_topics"
	actions := []string{"replace_repo_topics"}
	baseline, err := w.readTopics()
	if err != nil {
		w.recordScenarioFailure(scenario, actions, tag, "read repository topics baseline: "+err.Error())
		return false
	}
	if err := w.recordTopicBaseline(scenario, tag, baseline); err != nil {
		w.recordScenarioFailure(scenario, actions, tag, "durably record repository topics baseline before mutation: "+err.Error())
		return false
	}
	changed := false
	cleanup := func() bool {
		if !changed {
			return true
		}
		clean := w.runAction(scenario, "replace_repo_topics_restore", "replace_repo_topics", tag,
			map[string]any{"names": baseline}, "topics:"+tag,
			func() (string, error) {
				names, err := w.readTopics()
				if err != nil {
					return "", err
				}
				if !sameStrings(names, baseline) {
					return "", fmt.Errorf("repository topics do not equal the captured baseline after restore")
				}
				return "topics:baseline_restored", nil
			})
		if !clean {
			w.recordLeak(tag, actions, "repository topics were not restored to their captured baseline")
			return false
		}
		w.markScenarioPassed(tag, actions, "repository topics restored to the captured baseline with independent GitHub direct-read")
		return true
	}
	if !w.runAction(scenario, "replace_repo_topics", "replace_repo_topics", tag,
		map[string]any{"names": []string{tag}}, "topics:"+tag,
		func() (string, error) {
			names, err := w.readTopics()
			if err != nil {
				return "", err
			}
			if !sameStrings(names, []string{tag}) {
				return "", fmt.Errorf("repository topics do not contain only the certification tag")
			}
			return "topics:" + tag, nil
		}) {
		return cleanup() && false
	}
	changed = true
	return cleanup()
}

// runAction is the only write dispatch for this wave. It always generates a
// record from the declared post-schema record_schema, seeds it through the
// normal ETL path, creates a production reverse plan, previews it, and runs
// it with the normal approval/confirmation boundary. verify is an independent
// connector read (or the fixed direct-read command for topics), never the
// reverse-run response.
func (w *githubRepositoryWave) runAction(scenario, stageID, action, tag string, overrides map[string]any, resourceHint string, verify func() (string, error)) bool {
	stageName := "write_wave_" + stageID
	stage := recordStage(w.rc, w.rep, stageName, 2, func() (bool, CLIStageInfo, string) {
		schema, err := writeActionRecordSchema("github", action)
		if err != nil {
			return false, CLIStageInfo{}, fmt.Sprintf("%s: load declared record schema: %v", action, err)
		}
		record, err := GenerateRecordWithOverrides(schema, tag, w.runID, overrides)
		if err != nil {
			return false, CLIStageInfo{}, fmt.Sprintf("%s: generate declared record: %v", action, err)
		}
		if err := w.ledger.RecordPlanned(LedgerEntry{
			Action:     action,
			Scenario:   scenario,
			Tag:        tag,
			Connector:  "github",
			EntityHint: resourceHint,
			ResourceID: resourceHint,
		}); err != nil {
			return false, CLIStageInfo{}, fmt.Sprintf("%s: write-ahead ledger: %v", action, err)
		}
		if cli, err := w.executeProductionWrite(scenario, stageID, action, record); err != nil {
			return false, cli, fmt.Sprintf("%s: production reverse write: %v", action, err)
		}
		// A completed reverse run is the earliest point at which the ledger can
		// durably record a possible provider mutation. Never put this checkpoint
		// after read-back: a crash or stale read between the two otherwise leaves
		// a real resource looking like a merely planned action.
		if err := w.ledger.RecordMutated(tag, resourceHint); err != nil {
			w.recordLeak(tag, []string{action}, "production reverse write completed but its durable mutation checkpoint failed: "+err.Error())
			return false, CLIStageInfo{}, fmt.Sprintf("%s: ledger mutation checkpoint: %v", action, err)
		}
		resourceID, err := w.readBackWithRetry(verify)
		if err != nil {
			w.recordLeak(tag, []string{action}, "production reverse write completed but independent read-back did not verify the run-owned resource: "+err.Error())
			return false, CLIStageInfo{}, fmt.Sprintf("%s: independent read-back: %v", action, err)
		}
		if resourceID == "" {
			resourceID = resourceHint
		}
		if err := w.ledger.RecordReadBack(tag, resourceID); err != nil {
			return false, CLIStageInfo{}, fmt.Sprintf("%s: ledger read-back checkpoint: %v", action, err)
		}
		return true, CLIStageInfo{}, ""
	})
	if !stage.Passed {
		if existing, ok := w.rep.Capabilities.WriteActions[action]; !ok || existing.Result != "leaked_resource" {
			w.markFailed(action, tag, stage.Error)
		}
		return false
	}
	w.completed[action] = true
	return true
}

// readBackWithRetry allows GitHub's collection projection a small, explicit
// window to observe a just-completed mutation. It is not a success shortcut:
// once the bound is exhausted, the action remains failed and leaked_resource.
func (w *githubRepositoryWave) readBackWithRetry(verify func() (string, error)) (string, error) {
	attempts := w.readBackAttempts
	if attempts <= 0 {
		attempts = defaultGitHubReadBackAttempts
	}
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		resourceID, err := verify()
		if err == nil {
			return resourceID, nil
		}
		lastErr = err
		if attempt == attempts {
			break
		}
		if err := w.waitForReadBackAttempt(attempt); err != nil {
			return "", fmt.Errorf("wait before read-back retry %d/%d: %w", attempt+1, attempts, err)
		}
	}
	return "", fmt.Errorf("not observed after %d independent read-back attempts: %w", attempts, lastErr)
}

func (w *githubRepositoryWave) waitForReadBackAttempt(attempt int) error {
	if w.waitForReadBack != nil {
		return w.waitForReadBack(attempt)
	}
	ctx := context.Background()
	if w.rc != nil && w.rc.ctx != nil {
		ctx = w.rc.ctx
	}
	delay := githubReadBackRetryBase << (attempt - 1)
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (w *githubRepositoryWave) executeProductionWrite(scenario, stageID, action string, record map[string]any) (CLIStageInfo, error) {
	seedTable := "cert_write_wave_" + safeWaveName(scenario) + "_" + safeWaveName(stageID)
	seedRecord := make(map[string]any, len(record)+1)
	for key, value := range record {
		seedRecord[key] = value
	}
	// A synthetic source-table primary key is deliberately not mapped into the
	// provider record. Some declared write schemas (notably topics.names) have
	// no scalar identity but still must use the production source-table route.
	seedRecord["certification_seed_id"] = NewTag("github-seed", w.runID)
	if err := seedGeneratedSourceTable(w.rc, seedTable, "certification_seed_id", seedRecord); err != nil {
		return CLIStageInfo{}, fmt.Errorf("seed declared record through ETL: %w", err)
	}

	mapArgs := make([]string, 0, len(record)*2)
	for _, field := range fieldNames(record) {
		mapArgs = append(mapArgs, "--map", field+":"+field)
	}
	planName := "cert_write_wave_" + safeWaveName(stageID)
	planRes := w.rc.run(append([]string{
		"reverse", "plan", planName,
		"--source-table", seedTable,
		"--destination", "github:" + sourceCredentialName,
		"--action", action,
	}, mapArgs...)...)
	if planRes.ExitCode != 0 {
		return cliInfoFrom(planRes), fmt.Errorf("reverse plan exit=%d stderr=%s", planRes.ExitCode, planRes.Stderr)
	}
	planID := firstMatch(planIDLinePattern, planRes.Stdout)
	if planID == "" {
		return cliInfoFrom(planRes), fmt.Errorf("reverse plan did not report a plan id")
	}
	destructive := githubWriteRequiresConfirmation(action)
	var token string
	if destructive {
		var previewRes CLIResult
		var previewErr string
		token, previewRes, previewErr = previewReversePlanApproval(w.rc, planID)
		if previewErr != "" {
			return cliInfoFrom(previewRes), fmt.Errorf("%s", previewErr)
		}
	} else {
		// Non-destructive reverse plans issue their single-use approval at plan
		// time; the JSON preview must redact it. Destructive plans intentionally
		// withhold approval until their preview declares the confirmation gate.
		token = firstMatch(approvalTokenLinePattern, planRes.Stdout)
		if token == "" {
			return cliInfoFrom(planRes), fmt.Errorf("reverse plan did not report a non-destructive approval token")
		}
		previewRes := w.rc.run("reverse", "preview", planID, "--json")
		if passed, msg := assertKind(w.rc, "write_wave_"+stageID, previewRes, "ReversePlanPreview", 0); !passed {
			return cliInfoFrom(previewRes), fmt.Errorf("%s", msg)
		}
		if passed, _, msg := checkPlanPreviewRedaction(previewRes, token); !passed {
			return cliInfoFrom(previewRes), fmt.Errorf("%s", msg)
		}
	}
	args := []string{"reverse", "run", planID, "--approval-token-stdin"}
	if destructive {
		args = append(args, "--confirm", "destructive")
	}
	args = append(args, "--json")
	runRes := w.rc.runWithStdin(token+"\n", args...)
	passed, errMsg := assertKind(w.rc, "write_wave_"+stageID, runRes, "ReverseRun", 0)
	if !passed {
		return cliInfoFrom(runRes), fmt.Errorf("%s", errMsg)
	}
	run, _ := runRes.Envelope["run"].(map[string]any)
	if succeeded, _ := run["records_succeeded"].(float64); succeeded < 1 {
		return cliInfoFrom(runRes), fmt.Errorf("records_succeeded=%v, want >=1", run["records_succeeded"])
	}
	if failed, _ := run["records_failed"].(float64); failed != 0 {
		return cliInfoFrom(runRes), fmt.Errorf("records_failed=%v, want 0", run["records_failed"])
	}
	return cliInfoFrom(runRes), nil
}

func githubWriteRequiresConfirmation(actionName string) bool {
	profile := certificationProfileFor("github")
	for _, action := range profile.writes {
		if action.Name == actionName {
			return engine.DestructiveTargetForWrite("github", action).RequiresApproval()
		}
	}
	// An unknown action must be stopped before dispatch by
	// writeActionRecordSchema. Returning true preserves a fail-closed posture
	// should this helper ever be called directly first.
	return true
}

func (w *githubRepositoryWave) verifyStreamPresent(stream, matchField string, matchValue any, expectField string, expectValue any) (map[string]any, string, error) {
	rows, err := w.readStream(stream)
	if err != nil {
		return nil, "", err
	}
	for _, row := range rows {
		if valuesEqual(row[matchField], matchValue) && valuesEqual(row[expectField], expectValue) {
			return row, streamResourceID(stream, row), nil
		}
	}
	return nil, "", fmt.Errorf("%s row with %s=%v and %s=%v was not returned", stream, matchField, matchValue, expectField, expectValue)
}

func (w *githubRepositoryWave) verifyStreamAbsent(stream, matchField string, matchValue any) (string, error) {
	rows, err := w.readStream(stream)
	if err != nil {
		return "", err
	}
	for _, row := range rows {
		if valuesEqual(row[matchField], matchValue) {
			return "", fmt.Errorf("%s row with %s=%v remains present", stream, matchField, matchValue)
		}
	}
	return stream + ":absent", nil
}

func (w *githubRepositoryWave) verifyIssueState(number int, want string) (string, error) {
	_, id, err := w.verifyStreamPresent("issues", "number", number, "state", want)
	return id, err
}

func (w *githubRepositoryWave) verifyIssueLocked(number int, want bool) (string, error) {
	_, id, err := w.verifyStreamPresent("issues", "number", number, "locked", want)
	return id, err
}

// verifyIssueCommentUpdated uses the declaration-owned direct-read command
// rather than the collection stream. Its json_redacted output policy omits
// comment bodies deliberately, so the independent proof binds the exact
// comment id and requires GitHub's non-sensitive updated_at to advance from
// the creation read-back.
func (w *githubRepositoryWave) verifyIssueCommentUpdated(commentID int, previousUpdatedAt string) (string, error) {
	response, err := w.readIssueCommentResponse(commentID)
	if err != nil {
		return "", err
	}
	updatedAt, _ := response["updated_at"].(string)
	if updatedAt == "" {
		return "", fmt.Errorf("direct read issue comment %d omitted updated_at", commentID)
	}
	if updatedAt == previousUpdatedAt {
		return "", fmt.Errorf("issue comment %d updated_at did not advance from %q", commentID, previousUpdatedAt)
	}
	return fmt.Sprintf("issue_comment:%d", commentID), nil
}

func (w *githubRepositoryWave) readIssueCommentResponse(commentID int) (map[string]any, error) {
	res := w.readIssueComment(commentID)
	if passed, msg := assertKind(w.rc, "write_repository_wave", res, "ConnectorCommandDirectRead", 0); !passed {
		return nil, fmt.Errorf("direct read issue comment %d: %s", commentID, msg)
	}
	response, _ := res.Envelope["response"].(map[string]any)
	if response == nil {
		return nil, fmt.Errorf("direct read issue comment %d omitted response object", commentID)
	}
	return response, nil
}

func (w *githubRepositoryWave) verifyIssueCommentAbsent(commentID int) (string, error) {
	res := w.readIssueComment(commentID)
	if res.Kind == "Error" && res.ExitCode != 0 && strings.Contains(strings.ToLower(res.Stdout+"\n"+res.Stderr), "404") {
		return fmt.Sprintf("issue_comment:%d:absent", commentID), nil
	}
	if passed, msg := assertKind(w.rc, "write_repository_wave", res, "ConnectorCommandDirectRead", 0); !passed {
		return "", fmt.Errorf("direct read deleted issue comment %d: %s", commentID, msg)
	}
	return "", fmt.Errorf("issue comment %d remains present after delete", commentID)
}

// verifyReleaseTag and verifyReleaseByID use declaration-owned direct reads.
// Release collection streams can lag a just-created or just-deleted release;
// the item endpoints bind the proof to the provider-assigned resource id.
func (w *githubRepositoryWave) verifyReleaseTag(tag, wantName string) (int, error) {
	response, err := w.readReleaseResponseByTag(tag)
	if err != nil {
		return 0, err
	}
	actualTag, _ := response["tag_name"].(string)
	if actualTag != tag {
		return 0, fmt.Errorf("direct read release tag %q returned tag_name %q", tag, actualTag)
	}
	actualName, _ := response["name"].(string)
	if actualName != wantName {
		return 0, fmt.Errorf("direct read release %q returned name %q, want %q", tag, actualName, wantName)
	}
	return integerField(response, "id")
}

func (w *githubRepositoryWave) verifyReleaseByID(releaseID int, wantName string) (int, error) {
	response, err := w.readReleaseResponseByID(releaseID)
	if err != nil {
		return 0, err
	}
	actualID, err := integerField(response, "id")
	if err != nil {
		return 0, err
	}
	if actualID != releaseID {
		return 0, fmt.Errorf("direct read release id %d returned id %d", releaseID, actualID)
	}
	actualName, _ := response["name"].(string)
	if actualName != wantName {
		return 0, fmt.Errorf("direct read release %d returned name %q, want %q", releaseID, actualName, wantName)
	}
	return actualID, nil
}

func (w *githubRepositoryWave) verifyReleaseAbsent(releaseID int) (string, error) {
	res := w.readReleaseByID(releaseID)
	if res.Kind == "Error" && res.ExitCode != 0 && strings.Contains(strings.ToLower(res.Stdout+"\n"+res.Stderr), "404") {
		return fmt.Sprintf("release:%d:absent", releaseID), nil
	}
	if passed, msg := assertKind(w.rc, "write_repository_wave", res, "ConnectorCommandDirectRead", 0); !passed {
		return "", fmt.Errorf("direct read deleted release %d: %s", releaseID, msg)
	}
	return "", fmt.Errorf("release %d remains present after delete", releaseID)
}

func (w *githubRepositoryWave) readReleaseResponseByTag(tag string) (map[string]any, error) {
	return w.releaseResponse(w.readReleaseByTag(tag), fmt.Sprintf("tag %q", tag))
}

func (w *githubRepositoryWave) readReleaseResponseByID(releaseID int) (map[string]any, error) {
	return w.releaseResponse(w.readReleaseByID(releaseID), fmt.Sprintf("id %d", releaseID))
}

func (w *githubRepositoryWave) releaseResponse(res CLIResult, description string) (map[string]any, error) {
	if passed, msg := assertKind(w.rc, "write_repository_wave", res, "ConnectorCommandDirectRead", 0); !passed {
		return nil, fmt.Errorf("direct read release %s: %s", description, msg)
	}
	response, _ := res.Envelope["response"].(map[string]any)
	if response == nil {
		return nil, fmt.Errorf("direct read release %s omitted response object", description)
	}
	return response, nil
}

func (w *githubRepositoryWave) readReleaseByTag(tag string) CLIResult {
	return w.rc.run("github", "releases", "tags", "view",
		"--credential", sourceCredentialName,
		"--tag", tag,
		"--json")
}

func (w *githubRepositoryWave) readReleaseByID(releaseID int) CLIResult {
	return w.rc.run("github", "releases", "view",
		"--credential", sourceCredentialName,
		"--release-id", strconv.Itoa(releaseID),
		"--json")
}

func (w *githubRepositoryWave) readIssueComment(commentID int) CLIResult {
	return w.rc.run("github", "issues", "comments", "view",
		"--credential", sourceCredentialName,
		"--comment-id", strconv.Itoa(commentID),
		"--json")
}

// Commit-comment direct reads have the same deliberately redacted body policy
// as issue comments. Bind the verified update to the provider-assigned id and
// GitHub's updated_at rather than pretending a redacted body was observable.
func (w *githubRepositoryWave) verifyCommitCommentUpdated(commentID int, previousUpdatedAt string) (string, error) {
	response, err := w.readCommitCommentResponse(commentID)
	if err != nil {
		return "", err
	}
	updatedAt, _ := response["updated_at"].(string)
	if updatedAt == "" {
		return "", fmt.Errorf("direct read commit comment %d omitted updated_at", commentID)
	}
	if updatedAt == previousUpdatedAt {
		return "", fmt.Errorf("commit comment %d updated_at did not advance from %q", commentID, previousUpdatedAt)
	}
	return fmt.Sprintf("commit_comment:%d", commentID), nil
}

func (w *githubRepositoryWave) verifyCommitCommentAbsent(commentID int) (string, error) {
	res := w.readCommitComment(commentID)
	if res.Kind == "Error" && res.ExitCode != 0 && strings.Contains(strings.ToLower(res.Stdout+"\n"+res.Stderr), "404") {
		return fmt.Sprintf("commit_comment:%d:absent", commentID), nil
	}
	if passed, msg := assertKind(w.rc, "write_repository_wave", res, "ConnectorCommandDirectRead", 0); !passed {
		return "", fmt.Errorf("direct read deleted commit comment %d: %s", commentID, msg)
	}
	return "", fmt.Errorf("commit comment %d remains present after delete", commentID)
}

func (w *githubRepositoryWave) readCommitCommentResponse(commentID int) (map[string]any, error) {
	res := w.readCommitComment(commentID)
	if passed, msg := assertKind(w.rc, "write_repository_wave", res, "ConnectorCommandDirectRead", 0); !passed {
		return nil, fmt.Errorf("direct read commit comment %d: %s", commentID, msg)
	}
	response, _ := res.Envelope["response"].(map[string]any)
	if response == nil {
		return nil, fmt.Errorf("direct read commit comment %d omitted response object", commentID)
	}
	return response, nil
}

func (w *githubRepositoryWave) readCommitComment(commentID int) CLIResult {
	return w.rc.run("github", "comments", "view",
		"--credential", sourceCredentialName,
		"--comment-id", strconv.Itoa(commentID),
		"--json")
}

func (w *githubRepositoryWave) verifyIssueLabels(number int, want ...string) (string, error) {
	rows, err := w.readStream("issues")
	if err != nil {
		return "", err
	}
	for _, row := range rows {
		if !valuesEqual(row["number"], number) {
			continue
		}
		for _, label := range want {
			if !rowHasLabel(row, label) {
				return "", fmt.Errorf("issue %d does not contain label %q", number, label)
			}
		}
		return streamResourceID("issues", row), nil
	}
	return "", fmt.Errorf("issue %d was not returned", number)
}

func (w *githubRepositoryWave) verifyIssueLabelAbsent(number int, absent string) (string, error) {
	rows, err := w.readStream("issues")
	if err != nil {
		return "", err
	}
	for _, row := range rows {
		if !valuesEqual(row["number"], number) {
			continue
		}
		if rowHasLabel(row, absent) {
			return "", fmt.Errorf("issue %d still contains label %q", number, absent)
		}
		return streamResourceID("issues", row), nil
	}
	return "", fmt.Errorf("issue %d was not returned", number)
}

func (w *githubRepositoryWave) verifyIssueNoLabels(number int) (string, error) {
	rows, err := w.readStream("issues")
	if err != nil {
		return "", err
	}
	for _, row := range rows {
		if !valuesEqual(row["number"], number) {
			continue
		}
		if labels, ok := row["labels"].([]any); ok && len(labels) != 0 {
			return "", fmt.Errorf("issue %d still has %d labels after restore", number, len(labels))
		}
		return streamResourceID("issues", row), nil
	}
	return "", fmt.Errorf("issue %d was not returned", number)
}

func (w *githubRepositoryWave) fixtureLabels(count int) ([]string, error) {
	rows, err := w.readStream("labels")
	if err != nil {
		return nil, err
	}
	labels := make([]string, 0, count)
	for _, row := range rows {
		name, _ := row["name"].(string)
		if name == "" || IsCertifyTag(name) {
			continue
		}
		labels = append(labels, name)
		if len(labels) == count {
			return labels, nil
		}
	}
	return nil, fmt.Errorf("fixture has %d usable baseline labels, need %d", len(labels), count)
}

func (w *githubRepositoryWave) firstCommitSHA() (string, error) {
	rows, err := w.readStream("commits")
	if err != nil {
		return "", err
	}
	for _, row := range rows {
		if sha, _ := row["sha"].(string); strings.TrimSpace(sha) != "" {
			return sha, nil
		}
	}
	return "", fmt.Errorf("fixture repository has no readable commit SHA")
}

func (w *githubRepositoryWave) readStream(stream string) ([]map[string]any, error) {
	primaryKey, ok := map[string]string{
		"issues":          "node_id",
		"issue_comments":  "id",
		"labels":          "name",
		"milestones":      "number",
		"releases":        "id",
		"commit_comments": "id",
		"branches":        "name",
		"tags":            "name",
		"commits":         "sha",
	}[stream]
	if !ok {
		return nil, fmt.Errorf("no repository-wave primary key declared for stream %q", stream)
	}
	w.readSequence++
	id := fmt.Sprintf("%02d_%s", w.readSequence, safeWaveName(stream))
	connection := "cert_write_wave_read_" + id
	table := "cert_write_wave_read_" + id
	connectionRes := w.rc.run("connections", "create", connection,
		"--source", "github:"+sourceCredentialName,
		"--destination", "warehouse:"+warehouseCredentialName,
		"--stream", stream,
		"--primary-key", primaryKey,
		"--sync-mode", "full_refresh_overwrite",
		"--table", table,
		"--json")
	if passed, msg := assertKind(w.rc, "write_repository_wave", connectionRes, "Connection", 0); !passed {
		return nil, fmt.Errorf("create independent read-back connection for %s: %s", stream, msg)
	}
	runRes := w.rc.run("etl", "run", "--connection", connection, "--stream", stream, "--json")
	if passed, msg := assertKind(w.rc, "write_repository_wave", runRes, "ETLRun", 0); !passed {
		return nil, fmt.Errorf("run independent read-back for %s: %s", stream, msg)
	}
	queryRes := w.rc.run("query", "run", "--table", table, "--json")
	if passed, msg := assertKind(w.rc, "write_repository_wave", queryRes, "QueryResult", 0); !passed {
		return nil, fmt.Errorf("query independent read-back for %s: %s", stream, msg)
	}
	rowsRaw, _ := queryRes.Envelope["rows"].([]any)
	rows := make([]map[string]any, 0, len(rowsRaw))
	for _, raw := range rowsRaw {
		if row, ok := raw.(map[string]any); ok {
			rows = append(rows, row)
		}
	}
	return rows, nil
}

func (w *githubRepositoryWave) readTopics() ([]string, error) {
	res := w.rc.run("github", "topics", "view", "--credential", sourceCredentialName, "--json")
	if passed, msg := assertKind(w.rc, "write_repository_wave", res, "ConnectorCommandDirectRead", 0); !passed {
		return nil, fmt.Errorf("direct read repository topics: %s", msg)
	}
	response, _ := res.Envelope["response"].(map[string]any)
	rawNames, ok := response["names"].([]any)
	if !ok {
		return nil, fmt.Errorf("direct read repository topics response omits names array")
	}
	names := make([]string, 0, len(rawNames))
	for _, raw := range rawNames {
		name, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("direct read repository topic has non-string value")
		}
		names = append(names, name)
	}
	return names, nil
}

func (w *githubRepositoryWave) markScenarioPassed(tag string, actions []string, reason string) {
	if err := w.ledger.RecordCleaned(tag); err != nil {
		w.recordScenarioFailure("ledger", actions, tag, "record verified cleanup: "+err.Error())
		return
	}
	if w.rep.Capabilities.WriteActions == nil {
		w.rep.Capabilities.WriteActions = map[string]WriteActionResult{}
	}
	for _, action := range actions {
		if !w.completed[action] {
			continue
		}
		item := w.declarations[action]
		w.rep.Capabilities.WriteActions[action] = WriteActionResult{
			Result:  "pass",
			Path:    item.Path,
			Risk:    item.Risk,
			Verify:  "read_back",
			Tag:     tag,
			Reason:  reason,
			Cleanup: "verified",
		}
	}
}

func (w *githubRepositoryWave) markFailed(action, tag, reason string) {
	if w.rep.Capabilities.WriteActions == nil {
		w.rep.Capabilities.WriteActions = map[string]WriteActionResult{}
	}
	item := w.declarations[action]
	w.rep.Capabilities.WriteActions[action] = WriteActionResult{
		Result: "fail", Path: item.Path, Risk: item.Risk, Tag: tag, Reason: reason,
	}
}

// markBlocked is intentionally distinct from fail and pass. A provider may
// have accepted a mutation, but a known permission boundary can prevent the
// independent read-back that certification requires. That is not coverage and
// must not be folded into a pass roll-up, even when cleanup was later proven.
func (w *githubRepositoryWave) markBlocked(actions []string, tag, reason, cleanup string) {
	if w.rep.Capabilities.WriteActions == nil {
		w.rep.Capabilities.WriteActions = map[string]WriteActionResult{}
	}
	w.rep.Passed = false
	for _, action := range actions {
		item := w.declarations[action]
		w.rep.Capabilities.WriteActions[action] = WriteActionResult{
			Result:  "blocked",
			Path:    item.Path,
			Risk:    item.Risk,
			Cleanup: cleanup,
			Verify:  "unavailable",
			Tag:     tag,
			Reason:  reason,
		}
	}
}

// markRecoveryUncertified records a completed cleanup without laundering a
// prior interrupted mutation into coverage. A resumed `all` wave replaces
// this outcome only after it sends the complete fresh scenario itself.
func (w *githubRepositoryWave) markRecoveryUncertified(actions []string, tag, reason string) {
	if w.rep.Capabilities.WriteActions == nil {
		w.rep.Capabilities.WriteActions = map[string]WriteActionResult{}
	}
	for _, action := range actions {
		item := w.declarations[action]
		w.rep.Capabilities.WriteActions[action] = WriteActionResult{
			Result:  "recovered_unverified",
			Path:    item.Path,
			Risk:    item.Risk,
			Cleanup: "verified",
			Verify:  "recovery_only",
			Tag:     tag,
			Reason:  reason,
		}
	}
}

func (w *githubRepositoryWave) recordScenarioFailure(scenario string, actions []string, tag, reason string) {
	for _, action := range actions {
		if !w.completed[action] {
			w.markFailed(action, tag, scenario+": "+reason)
		}
	}
}

func (w *githubRepositoryWave) recordLeak(tag string, actions []string, reason string) {
	if w.rep.Capabilities.WriteActions == nil {
		w.rep.Capabilities.WriteActions = map[string]WriteActionResult{}
	}
	w.rep.Leaks = append(w.rep.Leaks, Leak{Tag: tag, Connector: "github", Action: strings.Join(actions, ","), Reason: reason})
	for _, action := range actions {
		item := w.declarations[action]
		w.rep.Capabilities.WriteActions[action] = WriteActionResult{
			Result: "leaked_resource", Path: item.Path, Risk: item.Risk, Tag: tag, Reason: reason,
		}
	}
}

func safeWaveName(s string) string {
	s = strings.ToLower(s)
	s = strings.NewReplacer("-", "_", "/", "_", " ", "_").Replace(s)
	return strings.Trim(s, "_")
}

func integerField(row map[string]any, field string) (int, error) {
	switch value := row[field].(type) {
	case float64:
		return int(value), nil
	case float32:
		return int(value), nil
	case int:
		return value, nil
	case int64:
		return int(value), nil
	case jsonNumber:
		return strconv.Atoi(string(value))
	default:
		return 0, fmt.Errorf("read-back field %q is not an integer", field)
	}
}

// jsonNumber is kept local so the helper also accepts numeric test doubles
// without exporting a certification-specific conversion API.
type jsonNumber string

func valuesEqual(got, want any) bool {
	if got == nil || want == nil {
		return got == want
	}
	return fmt.Sprint(got) == fmt.Sprint(want)
}

func streamResourceID(stream string, row map[string]any) string {
	for _, field := range []string{"id", "number", "name", "sha", "node_id"} {
		if value := fmt.Sprint(row[field]); value != "" && value != "<nil>" {
			return stream + ":" + value
		}
	}
	return stream
}

func rowHasLabel(row map[string]any, want string) bool {
	labels, ok := row["labels"].([]any)
	if !ok {
		return false
	}
	for _, raw := range labels {
		if label, ok := raw.(map[string]any); ok && valuesEqual(label["name"], want) {
			return true
		}
	}
	return false
}

func sameStrings(left, right []string) bool {
	left = append([]string(nil), left...)
	right = append([]string(nil), right...)
	sort.Strings(left)
	sort.Strings(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
