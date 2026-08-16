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

type repositoryWriteWave struct {
	rc        *runContext
	rep       *Report
	ledger    *Ledger
	runID     string
	connector string
	profile   *engine.CertificationWriteWaveSpec

	// declarations lets every outcome retain its exact provider path and risk,
	// including a failure before a mutation reaches the provider.
	declarations map[string]writeActionInventoryItem
	completed    map[string]bool
	readSequence int

	// A provider can acknowledge a mutation before a fresh collection read sees
	// it. Keep that eventual-consistency wait bounded and testable; a failed
	// read-back remains a certification failure and a reported potential leak.
	readBackAttempts int
	waitForReadBack  func(attempt int) error
}

const (
	defaultReadBackAttempts = 5
	readBackRetryBase       = time.Second
)

func validateRepositoryWriteWave(profile *engine.CertificationWriteWaveSpec) error {
	if profile == nil {
		return fmt.Errorf("profile is missing")
	}
	if len(profile.Fixture.Config) == 0 || strings.TrimSpace(profile.Fixture.Description) == "" {
		return fmt.Errorf("fixture config and description are required")
	}
	if strings.TrimSpace(profile.TagPrefix) == "" || strings.TrimSpace(profile.TagSubjectPrefix) == "" {
		return fmt.Errorf("tag_prefix and tag_subject_prefix are required")
	}
	declaredActions := make(map[string]struct{}, len(profile.Actions))
	for _, action := range profile.Actions {
		if strings.TrimSpace(action) == "" {
			return fmt.Errorf("wave declares an empty action")
		}
		if _, duplicate := declaredActions[action]; duplicate {
			return fmt.Errorf("wave declares action %q more than once", action)
		}
		declaredActions[action] = struct{}{}
	}
	if len(declaredActions) == 0 {
		return fmt.Errorf("wave declares no actions")
	}
	for key, action := range profile.ActionBindings {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(action) == "" {
			return fmt.Errorf("wave action binding has an empty key or action")
		}
		if _, ok := declaredActions[action]; !ok {
			return fmt.Errorf("wave action binding %q references undeclared action %q", key, action)
		}
	}
	if len(profile.ActionBindings) == 0 {
		return fmt.Errorf("wave declares no action bindings")
	}
	scenarios := make(map[string]struct{}, len(profile.Scenarios))
	for _, scenario := range profile.Scenarios {
		if strings.TrimSpace(scenario.Name) == "" || strings.TrimSpace(scenario.LedgerName) == "" || strings.TrimSpace(scenario.TagName) == "" {
			return fmt.Errorf("wave scenario has an empty name, ledger_name, or tag_name")
		}
		if _, duplicate := scenarios[scenario.Name]; duplicate {
			return fmt.Errorf("wave declares scenario %q more than once", scenario.Name)
		}
		scenarios[scenario.Name] = struct{}{}
		if len(scenario.Actions) == 0 {
			return fmt.Errorf("wave scenario %q declares no actions", scenario.Name)
		}
		for _, action := range scenario.Actions {
			if _, ok := declaredActions[action]; !ok {
				return fmt.Errorf("wave scenario %q references undeclared action %q", scenario.Name, action)
			}
		}
	}
	if len(scenarios) == 0 {
		return fmt.Errorf("wave declares no scenarios")
	}
	for _, blocked := range profile.BlockedActions {
		if strings.TrimSpace(blocked.Name) == "" || strings.TrimSpace(blocked.Reason) == "" || len(blocked.Actions) == 0 {
			return fmt.Errorf("blocked action declaration is incomplete")
		}
		for _, action := range blocked.Actions {
			if _, ok := declaredActions[action]; !ok {
				return fmt.Errorf("blocked action %q references undeclared action %q", blocked.Name, action)
			}
		}
	}
	return nil
}

func (w *repositoryWriteWave) scenario(name string) engine.CertificationWriteWaveScenario {
	if w == nil || w.profile == nil {
		return engine.CertificationWriteWaveScenario{}
	}
	for _, scenario := range w.profile.Scenarios {
		if scenario.Name == name {
			return scenario
		}
	}
	return engine.CertificationWriteWaveScenario{}
}

func (w *repositoryWriteWave) scenarioActions(name string) []string {
	return append([]string(nil), w.scenario(name).Actions...)
}

func (w *repositoryWriteWave) action(binding string) string {
	if w == nil || w.profile == nil {
		return ""
	}
	return w.profile.ActionBindings[binding]
}

func (w *repositoryWriteWave) blockedAction(name string) engine.CertificationWriteWaveBlockedAction {
	if w == nil || w.profile == nil {
		return engine.CertificationWriteWaveBlockedAction{}
	}
	for _, blocked := range w.profile.BlockedActions {
		if blocked.Name == name {
			return blocked
		}
	}
	return engine.CertificationWriteWaveBlockedAction{}
}

func (w *repositoryWriteWave) tag(name string) string {
	scenario := w.scenario(name)
	return NewTag(w.profile.TagSubjectPrefix+"-"+scenario.TagName, w.runID)
}

func (w *repositoryWriteWave) scenarioTagPrefix(name string) string {
	return w.profile.TagPrefix + w.scenario(name).TagName + "-"
}

func (w *repositoryWriteWave) recoveryScenario(name string) string {
	return w.scenario(name).LedgerName + "_recovery"
}

// stageRepositoryWriteWave executes a definition-owned, self-contained
// repository wave after the generic pairing proves the baseline lifecycle.
// It is intentionally not part of a --full-only sweep: --write is the opt-in
// for a bounded live mutation, while --full-parity later refuses to claim
// success for every declared action that remains non-live.
func stageRepositoryWriteWave(rc *runContext, rep *Report) error {
	if !rc.opts.Write {
		return nil
	}
	profile, ok := certificationWriteWaveFor(rc.opts.Connector)
	if !ok {
		return nil
	}
	if err := validateRepositoryWriteWave(profile); err != nil {
		recordStage(rc, rep, "write_repository_wave", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, "write_repository_wave: invalid definition-owned profile: " + err.Error()
		})
		return nil
	}
	for key, expected := range profile.Fixture.Config {
		if strings.TrimSpace(rc.opts.Config[key]) == expected {
			continue
		}
		recordStage(rc, rep, "write_repository_wave", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("write_repository_wave: live repository certification is restricted to %s", profile.Fixture.Description)
		})
		return nil
	}

	inventory, err := writeActionInventoryFor(rc.opts.Connector)
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
	for _, action := range profile.Actions {
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

	wave := &repositoryWriteWave{
		rc:           rc,
		rep:          rep,
		ledger:       ledger,
		runID:        NewRunID8(),
		connector:    rc.opts.Connector,
		profile:      profile,
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

func (w *repositoryWriteWave) runSelected(selection string) (bool, error) {
	runners := map[string]func() bool{
		"issue_lifecycle": w.runIssueLifecycleScenario,
		"issue_comments":  w.runIssueCommentScenario,
		"issue_lock":      w.runIssueLockScenario,
		"issue_labels":    w.runIssueLabelScenario,
		"label_lifecycle": w.runLabelLifecycleScenario,
		"milestone":       w.runMilestoneScenario,
		"release":         w.runReleaseScenario,
		"commit_comment":  w.runCommitCommentScenario,
		"ref":             w.runRefScenario,
		"topics":          w.runTopicsScenario,
	}
	if selection != "" && selection != "all" {
		if selection == "recover" {
			return true, nil
		}
		if _, declared := w.scenario(selection).Actions, w.scenario(selection).Name; declared != "" {
			if run := runners[selection]; run != nil {
				return run(), nil
			}
			return false, fmt.Errorf("declared certification_write_wave %q has no shared repository-wave runner", selection)
		}
		return false, fmt.Errorf("certification_write_wave %q is not declared by the connector profile", selection)
	}
	ok := true
	for _, scenario := range w.profile.Scenarios {
		run := runners[scenario.Name]
		if run == nil {
			return false, fmt.Errorf("declared certification_write_wave %q has no shared repository-wave runner", scenario.Name)
		}
		if !run() {
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
func (w *repositoryWriteWave) recoverUncleanedReleaseScenarios() (bool, error) {
	entries, err := LoadLedger(filepath.Dir(w.ledger.Path()))
	if err != nil {
		return false, fmt.Errorf("load durable write ledger: %w", err)
	}
	pending := make([]LedgerStatus, 0)
	for _, status := range entries.Uncleaned() {
		if status.Connector != w.connector || status.Scenario != w.scenario("release").LedgerName || !strings.HasPrefix(status.Tag, w.scenarioTagPrefix("release")) {
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
		if !strings.HasPrefix(legacyTag, w.scenarioTagPrefix("release")) {
			return false, fmt.Errorf("certification_recovery_release_tag %q is not this connector profile's certification ownership tag", legacyTag)
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
			pending = append(pending, LedgerStatus{Tag: legacyTag, Connector: w.connector, Scenario: w.scenario("release").LedgerName})
		}
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].Tag < pending[j].Tag })
	for _, status := range pending {
		releaseID, err := w.releaseRecoveryID(status.Tag)
		if err != nil {
			w.recordLeak(status.Tag, []string{w.action("release_create"), w.action("release_update"), w.action("release_delete")}, "tagged release recovery could not identify the exact run-owned release: "+err.Error())
			return false, err
		}
		if !w.runAction(w.recoveryScenario("release"), "delete_release_recovery", w.action("release_delete"), status.Tag,
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
			if !w.runAction(w.recoveryScenario("release"), "delete_release_tag_recovery", w.action("ref_delete"), status.Tag,
				map[string]any{"ref": "tags/" + status.Tag}, "tag:"+status.Tag,
				func() (string, error) { return w.verifyStreamAbsent("tags", "name", status.Tag) }) {
				return false, fmt.Errorf("delete recovered tagged release ref %q", status.Tag)
			}
			break
		}
		w.markScenarioPassed(status.Tag, w.scenarioActions("release"), "interrupted tagged release removed with independent provider read-back")
	}
	return true, nil
}

func (w *repositoryWriteWave) releaseRecoveryID(tag string) (int, error) {
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
func (w *repositoryWriteWave) recoverUncleanedCommitCommentScenarios() (bool, error) {
	entries, err := LoadLedger(filepath.Dir(w.ledger.Path()))
	if err != nil {
		return false, fmt.Errorf("load durable write ledger: %w", err)
	}
	for _, status := range entries.Uncleaned() {
		if status.Connector != w.connector || status.Scenario != w.scenario("commit_comment").LedgerName || !strings.HasPrefix(status.Tag, w.scenarioTagPrefix("commit_comment")) {
			continue
		}
		commentID, err := commitCommentIDFromLedger(status)
		if err != nil {
			w.recordLeak(status.Tag, w.blockedAction("commit_comment_item_read").Actions, "tagged commit-comment recovery could not determine its provider id: "+err.Error())
			return false, err
		}
		present, err := w.commitCommentPresent(commentID)
		if err != nil {
			w.recordLeak(status.Tag, w.blockedAction("commit_comment_item_read").Actions, "tagged commit-comment cleanup could not be independently checked through the permitted collection stream: "+err.Error())
			return false, err
		}
		if present {
			if !w.runAction(w.recoveryScenario("commit_comment"), "delete_commit_comment_recovery", w.action("comment_delete"), status.Tag,
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
		blocked := w.blockedAction("commit_comment_item_read")
		w.markBlocked(blocked.Actions, status.Tag, blocked.Reason, "verified")
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

func (w *repositoryWriteWave) commitCommentPresent(commentID int) (bool, error) {
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

// recoverUncleanedIssueScenarios reconciles an interrupted issue scenario
// before a later wave starts. The durable ledger names only resources with
// our pm-cert ownership marker; recovery reads the repository independently,
// removes any tagged comment, and closes the tagged issue through the same
// reverse-plan path. It never guesses at an untagged or third-party issue.
func (w *repositoryWriteWave) recoverUncleanedIssueScenarios() (bool, error) {
	entries, err := LoadLedger(filepath.Dir(w.ledger.Path()))
	if err != nil {
		return false, fmt.Errorf("load durable write ledger: %w", err)
	}
	pending := make([]LedgerStatus, 0)
	for _, status := range entries.Uncleaned() {
		if status.Connector == w.connector && strings.HasPrefix(status.Tag, w.profile.TagPrefix) && w.isRecoverableIssueScenario(status.Scenario) {
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
		if status.Scenario == w.scenario("issue_comments").LedgerName {
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
			reason := "durable create entry was not visible in independent provider read-back; refusing to guess at cleanup"
			w.recordLeak(status.Tag, []string{w.action("issue_create")}, reason)
			return false, fmt.Errorf("%s: %s", status.Tag, reason)
		}
		number, err := integerField(found, "number")
		if err != nil {
			w.recordLeak(status.Tag, []string{w.action("issue_create")}, "tagged issue recovery could not determine its provider number: "+err.Error())
			return false, err
		}
		resourceID := fmt.Sprintf("issue:%d", number)
		if status.Scenario == w.scenario("issue_comments").LedgerName {
			for _, comment := range comments {
				body, _ := comment["body"].(string)
				if !strings.HasPrefix(body, status.Tag) {
					continue
				}
				commentID, err := integerField(comment, "id")
				if err != nil {
					w.recordLeak(status.Tag, []string{w.action("issue_comment_create")}, "tagged issue-comment recovery could not determine its provider comment id: "+err.Error())
					return false, err
				}
				if !w.runAction(w.recoveryScenario("issue_comments"), "delete_issue_comment_recovery", w.action("issue_comment_delete"), status.Tag,
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
		if !w.runAction(w.recoveryScenario("issue_lifecycle"), "close_issue_recovery", w.action("issue_close"), status.Tag,
			map[string]any{"issue_number": number}, resourceID,
			func() (string, error) { return w.verifyIssueState(number, "closed") }) {
			return false, fmt.Errorf("close recovered tagged issue %d", number)
		}
		w.markScenarioPassed(status.Tag, []string{w.action("issue_close")}, "interrupted tagged issue closed with independent provider read-back")
	}
	return true, nil
}

func (w *repositoryWriteWave) isRecoverableIssueScenario(scenario string) bool {
	switch scenario {
	case w.scenario("issue_lifecycle").LedgerName, w.scenario("issue_comments").LedgerName, w.scenario("issue_lock").LedgerName, w.scenario("issue_labels").LedgerName:
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
func (w *repositoryWriteWave) recoverUncleanedNamedResourceScenarios() (bool, error) {
	entries, err := LoadLedger(filepath.Dir(w.ledger.Path()))
	if err != nil {
		return false, fmt.Errorf("load durable write ledger: %w", err)
	}
	pending := make([]LedgerStatus, 0)
	for _, status := range entries.Uncleaned() {
		if status.Connector != w.connector || !strings.HasPrefix(status.Tag, w.profile.TagPrefix) {
			continue
		}
		switch status.Scenario {
		case w.scenario("label_lifecycle").LedgerName, w.scenario("milestone").LedgerName, w.scenario("ref").LedgerName:
			pending = append(pending, status)
		}
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].Tag < pending[j].Tag })
	for _, status := range pending {
		var actions []string
		var recoveryErr error
		switch status.Scenario {
		case w.scenario("label_lifecycle").LedgerName:
			actions, recoveryErr = w.recoverTaggedLabels(status)
		case w.scenario("milestone").LedgerName:
			actions, recoveryErr = w.recoverTaggedMilestones(status)
		case w.scenario("ref").LedgerName:
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
		w.markRecoveryUncertified(actions, status.Tag, "interrupted tagged resource was removed with independent provider read-back; the interrupted mutation is not certified")
	}
	return true, nil
}

func (w *repositoryWriteWave) recoverTaggedLabels(status LedgerStatus) ([]string, error) {
	actions := w.scenarioActions("label_lifecycle")
	rows, err := w.readStream("labels")
	if err != nil {
		return actions, fmt.Errorf("read tagged labels for recovery: %w", err)
	}
	for _, row := range rows {
		name, _ := row["name"].(string)
		if !strings.HasPrefix(name, status.Tag) {
			continue
		}
		if !w.runAction(w.recoveryScenario("label_lifecycle"), "delete_label_recovery", w.action("label_delete"), status.Tag,
			map[string]any{"name": name}, "label:"+name,
			func() (string, error) { return w.verifyStreamAbsent("labels", "name", name) }) {
			return actions, fmt.Errorf("delete recovered tagged label %q", name)
		}
	}
	return actions, nil
}

func (w *repositoryWriteWave) recoverTaggedMilestones(status LedgerStatus) ([]string, error) {
	actions := w.scenarioActions("milestone")
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
		if !w.runAction(w.recoveryScenario("milestone"), "delete_milestone_recovery", w.action("milestone_delete"), status.Tag,
			map[string]any{"milestone_number": number}, fmt.Sprintf("milestone:%d", number),
			func() (string, error) { return w.verifyStreamAbsent("milestones", "number", number) }) {
			return actions, fmt.Errorf("delete recovered tagged milestone %d", number)
		}
	}
	return actions, nil
}

func (w *repositoryWriteWave) recoverTaggedRefs(status LedgerStatus) ([]string, error) {
	actions := w.scenarioActions("ref")
	rows, err := w.readStream("branches")
	if err != nil {
		return actions, fmt.Errorf("read tagged branches for recovery: %w", err)
	}
	for _, row := range rows {
		name, _ := row["name"].(string)
		if !strings.HasPrefix(name, status.Tag) {
			continue
		}
		if !w.runAction(w.recoveryScenario("ref"), "delete_ref_recovery", w.action("ref_delete"), status.Tag,
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

func (w *repositoryWriteWave) recoverUncleanedTopicScenarios() (bool, error) {
	entries, err := LoadLedger(filepath.Dir(w.ledger.Path()))
	if err != nil {
		return false, fmt.Errorf("load durable write ledger: %w", err)
	}
	for _, status := range entries.Uncleaned() {
		if status.Connector != w.connector || status.Scenario != w.scenario("topics").LedgerName || !strings.HasPrefix(status.Tag, w.scenarioTagPrefix("topics")) {
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
			w.markRecoveryUncertified([]string{w.action("topics_replace")}, status.Tag, "interrupted topic scenario had no visible mutation; it is not certified")
			continue
		}
		if !sameStrings(names, []string{status.Tag}) {
			reason := fmt.Sprintf("tagged topic recovery %q found topics other than its exact tag or captured baseline; refusing to overwrite an unknown provider state", status.Tag)
			w.recordLeak(status.Tag, []string{w.action("topics_replace")}, reason)
			return false, fmt.Errorf("%s", reason)
		}
		if !w.runAction(w.recoveryScenario("topics"), "replace_repo_topics_restore_recovery", w.action("topics_replace"), status.Tag,
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
		w.markRecoveryUncertified([]string{w.action("topics_replace")}, status.Tag, "interrupted tagged topic mutation was restored with independent provider read-back; it is not certified")
	}
	return true, nil
}

func (w *repositoryWriteWave) recordTopicBaseline(scenario, tag string, baseline []string) error {
	recovery, err := json.Marshal(topicRecoveryBaseline{Baseline: append([]string(nil), baseline...)})
	if err != nil {
		return fmt.Errorf("encode captured topic baseline: %w", err)
	}
	return w.ledger.RecordPlanned(LedgerEntry{
		Action:     w.action("topics_replace"),
		Scenario:   scenario,
		Tag:        tag,
		Connector:  w.connector,
		EntityHint: "topics:" + tag,
		ResourceID: "topics:" + tag,
		Recovery:   recovery,
	})
}

func (w *repositoryWriteWave) runIssueLifecycleScenario() bool {
	tag := w.tag("issue_lifecycle")
	scenario := w.scenario("issue_lifecycle").LedgerName
	actions := w.scenarioActions("issue_lifecycle")
	updated := tag + " updated"
	var number int
	created := false
	cleanup := func() bool {
		if !created {
			return true
		}
		clean := w.runAction(scenario, "close_issue_cleanup", w.action("issue_close"), tag,
			map[string]any{"issue_number": number}, fmt.Sprintf("issue:%d", number),
			func() (string, error) { return w.verifyIssueState(number, "closed") })
		if !clean {
			w.recordLeak(tag, actions, "tagged issue could not be closed for cleanup")
			return false
		}
		w.markScenarioPassed(tag, actions, "tagged issue closed with independent provider read-back")
		return true
	}
	if !w.runAction(scenario, w.action("issue_create"), w.action("issue_create"), tag,
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
	if !w.runAction(scenario, w.action("issue_update"), w.action("issue_update"), tag,
		map[string]any{"issue_number": number, "title": updated, "body": updated}, fmt.Sprintf("issue:%d", number),
		func() (string, error) {
			_, id, err := w.verifyStreamPresent("issues", "number", number, "title", updated)
			return id, err
		}) {
		return cleanup() && false
	}
	if !w.runAction(scenario, w.action("issue_close"), w.action("issue_close"), tag,
		map[string]any{"issue_number": number}, fmt.Sprintf("issue:%d", number),
		func() (string, error) { return w.verifyIssueState(number, "closed") }) {
		return cleanup() && false
	}
	if !w.runAction(scenario, w.action("issue_reopen"), w.action("issue_reopen"), tag,
		map[string]any{"issue_number": number}, fmt.Sprintf("issue:%d", number),
		func() (string, error) { return w.verifyIssueState(number, "open") }) {
		return cleanup() && false
	}
	return cleanup()
}

func (w *repositoryWriteWave) runIssueCommentScenario() bool {
	tag := w.tag("issue_comments")
	scenario := w.scenario("issue_comments").LedgerName
	actions := w.scenarioActions("issue_comments")
	body := tag + " comment"
	updated := body + " updated"
	var number, commentID int
	var commentUpdatedAt string
	issueCreated := false
	commentAlive := false
	cleanup := func() bool {
		clean := true
		if commentAlive {
			if !w.runAction(scenario, "delete_issue_comment_cleanup", w.action("issue_comment_delete"), tag,
				map[string]any{"comment_id": commentID}, fmt.Sprintf("issue_comment:%d", commentID),
				func() (string, error) { return w.verifyIssueCommentAbsent(commentID) }) {
				clean = false
			}
		}
		if issueCreated {
			if !w.runAction(scenario, "close_issue_cleanup", w.action("issue_close"), tag,
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
			w.markScenarioPassed(tag, actions, "tagged issue comment removed and containing issue closed with independent provider read-back")
		}
		return true
	}
	if !w.runAction(scenario, w.action("issue_create"), w.action("issue_create"), tag,
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
	if !w.runAction(scenario, w.action("issue_comment_create"), w.action("issue_comment_create"), tag,
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
	if !w.runAction(scenario, w.action("issue_comment_update"), w.action("issue_comment_update"), tag,
		map[string]any{"comment_id": commentID, "body": updated}, fmt.Sprintf("issue_comment:%d", commentID),
		func() (string, error) {
			return w.verifyIssueCommentUpdated(commentID, commentUpdatedAt)
		}) {
		return cleanup() && false
	}
	if !w.runAction(scenario, w.action("issue_comment_delete"), w.action("issue_comment_delete"), tag,
		map[string]any{"comment_id": commentID}, fmt.Sprintf("issue_comment:%d", commentID),
		func() (string, error) { return w.verifyIssueCommentAbsent(commentID) }) {
		return cleanup() && false
	}
	commentAlive = false
	return cleanup()
}

func (w *repositoryWriteWave) runIssueLockScenario() bool {
	tag := w.tag("issue_lock")
	scenario := w.scenario("issue_lock").LedgerName
	actions := w.scenarioActions("issue_lock")
	var number int
	created := false
	locked := false
	cleanup := func() bool {
		clean := true
		if locked {
			if !w.runAction(scenario, "unlock_issue_cleanup", w.action("issue_unlock"), tag,
				map[string]any{"issue_number": number}, fmt.Sprintf("issue:%d", number),
				func() (string, error) { return w.verifyIssueLocked(number, false) }) {
				clean = false
			}
		}
		if created {
			if !w.runAction(scenario, "close_issue_cleanup", w.action("issue_close"), tag,
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
			w.markScenarioPassed(tag, actions, "tagged issue unlocked and closed with independent provider read-back")
		}
		return true
	}
	if !w.runAction(scenario, w.action("issue_create"), w.action("issue_create"), tag,
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
	if !w.runAction(scenario, w.action("issue_lock"), w.action("issue_lock"), tag,
		map[string]any{"issue_number": number, "lock_reason": "resolved"}, fmt.Sprintf("issue:%d", number),
		func() (string, error) { return w.verifyIssueLocked(number, true) }) {
		return cleanup() && false
	}
	locked = true
	if !w.runAction(scenario, w.action("issue_unlock"), w.action("issue_unlock"), tag,
		map[string]any{"issue_number": number}, fmt.Sprintf("issue:%d", number),
		func() (string, error) { return w.verifyIssueLocked(number, false) }) {
		return cleanup() && false
	}
	locked = false
	return cleanup()
}

func (w *repositoryWriteWave) runIssueLabelScenario() bool {
	tag := w.tag("issue_labels")
	scenario := w.scenario("issue_labels").LedgerName
	actions := w.scenarioActions("issue_labels")
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
			if !w.runAction(scenario, "set_issue_labels_restore", w.action("issue_labels_set"), tag,
				map[string]any{"issue_number": number, "labels": []string{}}, fmt.Sprintf("issue:%d", number),
				func() (string, error) { return w.verifyIssueNoLabels(number) }) {
				clean = false
			}
		}
		if created {
			if !w.runAction(scenario, "close_issue_cleanup", w.action("issue_close"), tag,
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
			w.markScenarioPassed(tag, actions, "tagged issue labels restored to empty and issue closed with independent provider read-back")
		}
		return true
	}
	if !w.runAction(scenario, w.action("issue_create"), w.action("issue_create"), tag,
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
	if !w.runAction(scenario, w.action("issue_labels_set"), w.action("issue_labels_set"), tag,
		map[string]any{"issue_number": number, "labels": []string{labels[0]}}, fmt.Sprintf("issue:%d", number),
		func() (string, error) { return w.verifyIssueLabels(number, labels[0]) }) {
		return cleanup() && false
	}
	labelsTouched = true
	if !w.runAction(scenario, w.action("issue_labels_add"), w.action("issue_labels_add"), tag,
		map[string]any{"issue_number": number, "labels": []string{labels[1]}}, fmt.Sprintf("issue:%d", number),
		func() (string, error) { return w.verifyIssueLabels(number, labels[0], labels[1]) }) {
		return cleanup() && false
	}
	if !w.runAction(scenario, w.action("issue_label_remove"), w.action("issue_label_remove"), tag,
		map[string]any{"issue_number": number, "name": labels[1]}, fmt.Sprintf("issue:%d", number),
		func() (string, error) { return w.verifyIssueLabelAbsent(number, labels[1]) }) {
		return cleanup() && false
	}
	return cleanup()
}

func (w *repositoryWriteWave) runLabelLifecycleScenario() bool {
	tag := w.tag("label_lifecycle")
	scenario := w.scenario("label_lifecycle").LedgerName
	actions := w.scenarioActions("label_lifecycle")
	first, updated, second := waveLabelNames(tag)
	firstCreated := false
	secondCreated := false
	firstCurrent := first
	cleanup := func() bool {
		clean := true
		if secondCreated {
			if !w.runAction(scenario, "delete_label_second", w.action("label_delete"), tag,
				map[string]any{"name": second}, "label:"+second,
				func() (string, error) { return w.verifyStreamAbsent("labels", "name", second) }) {
				clean = false
			}
		}
		if firstCreated {
			if !w.runAction(scenario, "delete_label_first", w.action("label_delete"), tag,
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
			w.markScenarioPassed(tag, actions, "tagged labels removed with independent provider read-back")
		}
		return true
	}
	if !w.runAction(scenario, "create_label_first", w.action("label_create"), tag,
		map[string]any{"name": first, "color": "ededed", "description": tag}, "label:"+first,
		func() (string, error) {
			_, id, err := w.verifyStreamPresent("labels", "name", first, "name", first)
			return id, err
		}) {
		return cleanup() && false
	}
	firstCreated = true
	if !w.runAction(scenario, "create_label_second", w.action("label_create"), tag,
		map[string]any{"name": second, "color": "ededed", "description": tag}, "label:"+second,
		func() (string, error) {
			_, id, err := w.verifyStreamPresent("labels", "name", second, "name", second)
			return id, err
		}) {
		return cleanup() && false
	}
	secondCreated = true
	if !w.runAction(scenario, w.action("label_update"), w.action("label_update"), tag,
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

// waveLabelNames leaves room under the provider's label-name
// ceiling. NewTag already carries the full run ownership marker, so compact
// distinct suffixes preserve cleanup safety without turning an update into a
// provider-side 422.
func waveLabelNames(tag string) (first, updated, second string) {
	return tag + "-a", tag + "-b", tag + "-c"
}

func (w *repositoryWriteWave) runMilestoneScenario() bool {
	tag := w.tag("milestone")
	scenario := w.scenario("milestone").LedgerName
	actions := w.scenarioActions("milestone")
	updated := tag + " updated"
	var number int
	created := false
	cleanup := func() bool {
		if !created {
			return true
		}
		clean := w.runAction(scenario, w.action("milestone_delete"), w.action("milestone_delete"), tag,
			map[string]any{"milestone_number": number}, fmt.Sprintf("milestone:%d", number),
			func() (string, error) { return w.verifyStreamAbsent("milestones", "number", number) })
		if !clean {
			w.recordLeak(tag, actions, "tagged milestone remained after cleanup")
			return false
		}
		w.markScenarioPassed(tag, actions, "tagged milestone removed with independent provider read-back")
		return true
	}
	if !w.runAction(scenario, w.action("milestone_create"), w.action("milestone_create"), tag,
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
	if !w.runAction(scenario, w.action("milestone_update"), w.action("milestone_update"), tag,
		map[string]any{"milestone_number": number, "title": updated, "description": tag + " updated"}, fmt.Sprintf("milestone:%d", number),
		func() (string, error) {
			_, id, err := w.verifyStreamPresent("milestones", "number", number, "title", updated)
			return id, err
		}) {
		return cleanup() && false
	}
	return cleanup()
}

func (w *repositoryWriteWave) runReleaseScenario() bool {
	tag := w.tag("release")
	scenario := w.scenario("release").LedgerName
	actions := w.scenarioActions("release")
	updated := tag + " updated"
	var releaseID int
	created := false
	deletedRelease := false
	cleanup := func() bool {
		clean := true
		if created && !deletedRelease {
			if !w.runAction(scenario, "delete_release_cleanup", w.action("release_delete"), tag,
				map[string]any{"release_id": releaseID}, fmt.Sprintf("release:%d", releaseID),
				func() (string, error) { return w.verifyStreamAbsent("releases", "id", releaseID) }) {
				clean = false
			}
		}
		// The provider's delete-release endpoint deliberately leaves its tag behind.
		// Deleting that run-owned tag via the declared ref-delete action is part
		// of cleanup, not an untracked shell/API side effect.
		if created {
			if !w.runAction(scenario, "delete_ref_release_tag", w.action("ref_delete"), tag,
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
			w.markScenarioPassed(tag, actions, "tagged release deleted and its release-created tag removed with independent provider read-back")
		}
		return true
	}
	if !w.runAction(scenario, w.action("release_create"), w.action("release_create"), tag,
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
	if !w.runAction(scenario, w.action("release_update"), w.action("release_update"), tag,
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
	if !w.runAction(scenario, w.action("release_delete"), w.action("release_delete"), tag,
		map[string]any{"release_id": releaseID}, fmt.Sprintf("release:%d", releaseID),
		func() (string, error) { return w.verifyReleaseAbsent(releaseID) }) {
		return cleanup() && false
	}
	deletedRelease = true
	return cleanup()
}

func (w *repositoryWriteWave) runCommitCommentScenario() bool {
	tag := w.tag("commit_comment")
	scenario := w.scenario("commit_comment").LedgerName
	blocked := w.blockedAction("commit_comment_item_read")
	actions := blocked.Actions
	// A create/update/delete response cannot stand in for the item-level
	// read-back. The current disposable token has been observed refusing that
	// read, so do not keep creating a comment that cannot be certified. After a
	// browser permission repair the caller opts in to re-attempting this bounded
	// scenario; a provider refusal still leaves it blocked and non-pass.
	if configValue(w.rc.opts.Config, "certification_commit_comment_item_read", "") != "enabled" {
		w.markBlocked(actions, tag, blocked.Reason, "not_attempted")
		recordStage(w.rc, w.rep, "write_wave_commit_comment", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, blocked.Reason
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
		clean := w.runAction(scenario, w.action("comment_delete"), w.action("comment_delete"), tag,
			map[string]any{"comment_id": commentID}, fmt.Sprintf("commit_comment:%d", commentID),
			func() (string, error) { return w.verifyCommitCommentAbsent(commentID) })
		if !clean {
			w.recordLeak(tag, actions, "tagged commit comment remained after cleanup")
			return false
		}
		w.markScenarioPassed(tag, actions, "tagged commit comment removed with independent provider read-back")
		return true
	}
	if !w.runAction(scenario, w.action("comment_create"), w.action("comment_create"), tag,
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
	if !w.runAction(scenario, w.action("comment_update"), w.action("comment_update"), tag,
		map[string]any{"comment_id": commentID, "body": updated}, fmt.Sprintf("commit_comment:%d", commentID),
		func() (string, error) {
			return w.verifyCommitCommentUpdated(commentID, commentUpdatedAt)
		}) {
		return cleanup() && false
	}
	if !w.runAction(scenario, w.action("comment_delete"), w.action("comment_delete"), tag,
		map[string]any{"comment_id": commentID}, fmt.Sprintf("commit_comment:%d", commentID),
		func() (string, error) { return w.verifyCommitCommentAbsent(commentID) }) {
		return cleanup() && false
	}
	commentAlive = false
	w.markScenarioPassed(tag, actions, "tagged commit comment removed with independent provider read-back")
	return true
}

func (w *repositoryWriteWave) runRefScenario() bool {
	tag := w.tag("ref")
	scenario := w.scenario("ref").LedgerName
	actions := w.scenarioActions("ref")
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
		clean := w.runAction(scenario, w.action("ref_delete"), w.action("ref_delete"), tag,
			map[string]any{"ref": "heads/" + branch}, "ref:heads/"+branch,
			func() (string, error) { return w.verifyStreamAbsent("branches", "name", branch) })
		if !clean {
			w.recordLeak(tag, actions, "tagged branch ref remained after cleanup")
			return false
		}
		w.markScenarioPassed(tag, actions, "tagged branch ref removed with independent provider read-back")
		return true
	}
	if !w.runAction(scenario, w.action("ref_create"), w.action("ref_create"), tag,
		map[string]any{"ref": "refs/heads/" + branch, "sha": sha}, "ref:heads/"+branch,
		func() (string, error) {
			_, id, err := w.verifyStreamPresent("branches", "name", branch, "commit_sha", sha)
			return id, err
		}) {
		return cleanup() && false
	}
	created = true
	if !w.runAction(scenario, w.action("ref_update"), w.action("ref_update"), tag,
		map[string]any{"ref": "heads/" + branch, "sha": sha, "force": false}, "ref:heads/"+branch,
		func() (string, error) {
			_, id, err := w.verifyStreamPresent("branches", "name", branch, "commit_sha", sha)
			return id, err
		}) {
		return cleanup() && false
	}
	return cleanup()
}

func (w *repositoryWriteWave) runTopicsScenario() bool {
	tag := w.tag("topics")
	scenario := w.scenario("topics").LedgerName
	actions := w.scenarioActions("topics")
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
		clean := w.runAction(scenario, "replace_repo_topics_restore", w.action("topics_replace"), tag,
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
		w.markScenarioPassed(tag, actions, "repository topics restored to the captured baseline with independent provider direct-read")
		return true
	}
	if !w.runAction(scenario, w.action("topics_replace"), w.action("topics_replace"), tag,
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
func (w *repositoryWriteWave) runAction(scenario, stageID, action, tag string, overrides map[string]any, resourceHint string, verify func() (string, error)) bool {
	stageName := "write_wave_" + stageID
	stage := recordStage(w.rc, w.rep, stageName, 2, func() (bool, CLIStageInfo, string) {
		schema, err := writeActionRecordSchema(w.connector, action)
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
			Connector:  w.connector,
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

// readBackWithRetry allows the provider collection projection a small, explicit
// window to observe a just-completed mutation. It is not a success shortcut:
// once the bound is exhausted, the action remains failed and leaked_resource.
func (w *repositoryWriteWave) readBackWithRetry(verify func() (string, error)) (string, error) {
	attempts := w.readBackAttempts
	if attempts <= 0 {
		attempts = defaultReadBackAttempts
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

func (w *repositoryWriteWave) waitForReadBackAttempt(attempt int) error {
	if w.waitForReadBack != nil {
		return w.waitForReadBack(attempt)
	}
	ctx := context.Background()
	if w.rc != nil && w.rc.ctx != nil {
		ctx = w.rc.ctx
	}
	delay := readBackRetryBase << (attempt - 1)
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (w *repositoryWriteWave) executeProductionWrite(scenario, stageID, action string, record map[string]any) (CLIStageInfo, error) {
	seedTable := "cert_write_wave_" + safeWaveName(scenario) + "_" + safeWaveName(stageID)
	seedRecord := make(map[string]any, len(record)+1)
	for key, value := range record {
		seedRecord[key] = value
	}
	// A synthetic source-table primary key is deliberately not mapped into the
	// provider record. Some declared write schemas (notably topics.names) have
	// no scalar identity but still must use the production source-table route.
	seedRecord["certification_seed_id"] = NewTag(w.profile.TagSubjectPrefix+"-seed", w.runID)
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
		"--destination", w.connector + ":" + sourceCredentialName,
		"--action", action,
	}, mapArgs...)...)
	if planRes.ExitCode != 0 {
		return cliInfoFrom(planRes), fmt.Errorf("reverse plan exit=%d stderr=%s", planRes.ExitCode, planRes.Stderr)
	}
	planID := firstMatch(planIDLinePattern, planRes.Stdout)
	if planID == "" {
		return cliInfoFrom(planRes), fmt.Errorf("reverse plan did not report a plan id")
	}
	destructive := w.writeRequiresConfirmation(action)
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

func (w *repositoryWriteWave) writeRequiresConfirmation(actionName string) bool {
	profile := certificationProfileFor(w.connector)
	for _, action := range profile.writes {
		if action.Name == actionName {
			return engine.DestructiveTargetForWrite(w.connector, action).RequiresApproval()
		}
	}
	// An unknown action must be stopped before dispatch by
	// writeActionRecordSchema. Returning true preserves a fail-closed posture
	// should this helper ever be called directly first.
	return true
}

func (w *repositoryWriteWave) verifyStreamPresent(stream, matchField string, matchValue any, expectField string, expectValue any) (map[string]any, string, error) {
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

func (w *repositoryWriteWave) verifyStreamAbsent(stream, matchField string, matchValue any) (string, error) {
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

func (w *repositoryWriteWave) verifyIssueState(number int, want string) (string, error) {
	_, id, err := w.verifyStreamPresent("issues", "number", number, "state", want)
	return id, err
}

func (w *repositoryWriteWave) verifyIssueLocked(number int, want bool) (string, error) {
	_, id, err := w.verifyStreamPresent("issues", "number", number, "locked", want)
	return id, err
}

// verifyIssueCommentUpdated uses the declaration-owned direct-read command
// rather than the collection stream. Its json_redacted output policy omits
// comment bodies deliberately, so the independent proof binds the exact
// comment id and requires the provider's non-sensitive updated_at to advance from
// the creation read-back.
func (w *repositoryWriteWave) verifyIssueCommentUpdated(commentID int, previousUpdatedAt string) (string, error) {
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

func (w *repositoryWriteWave) readIssueCommentResponse(commentID int) (map[string]any, error) {
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

func (w *repositoryWriteWave) verifyIssueCommentAbsent(commentID int) (string, error) {
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
func (w *repositoryWriteWave) verifyReleaseTag(tag, wantName string) (int, error) {
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

func (w *repositoryWriteWave) verifyReleaseByID(releaseID int, wantName string) (int, error) {
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

func (w *repositoryWriteWave) verifyReleaseAbsent(releaseID int) (string, error) {
	res := w.readReleaseByID(releaseID)
	if res.Kind == "Error" && res.ExitCode != 0 && strings.Contains(strings.ToLower(res.Stdout+"\n"+res.Stderr), "404") {
		return fmt.Sprintf("release:%d:absent", releaseID), nil
	}
	if passed, msg := assertKind(w.rc, "write_repository_wave", res, "ConnectorCommandDirectRead", 0); !passed {
		return "", fmt.Errorf("direct read deleted release %d: %s", releaseID, msg)
	}
	return "", fmt.Errorf("release %d remains present after delete", releaseID)
}

func (w *repositoryWriteWave) readReleaseResponseByTag(tag string) (map[string]any, error) {
	return w.releaseResponse(w.readReleaseByTag(tag), fmt.Sprintf("tag %q", tag))
}

func (w *repositoryWriteWave) readReleaseResponseByID(releaseID int) (map[string]any, error) {
	return w.releaseResponse(w.readReleaseByID(releaseID), fmt.Sprintf("id %d", releaseID))
}

func (w *repositoryWriteWave) releaseResponse(res CLIResult, description string) (map[string]any, error) {
	if passed, msg := assertKind(w.rc, "write_repository_wave", res, "ConnectorCommandDirectRead", 0); !passed {
		return nil, fmt.Errorf("direct read release %s: %s", description, msg)
	}
	response, _ := res.Envelope["response"].(map[string]any)
	if response == nil {
		return nil, fmt.Errorf("direct read release %s omitted response object", description)
	}
	return response, nil
}

func (w *repositoryWriteWave) readReleaseByTag(tag string) CLIResult {
	return w.rc.run(w.connector, "releases", "tags", "view",
		"--credential", sourceCredentialName,
		"--tag", tag,
		"--json")
}

func (w *repositoryWriteWave) readReleaseByID(releaseID int) CLIResult {
	return w.rc.run(w.connector, "releases", "view",
		"--credential", sourceCredentialName,
		"--release-id", strconv.Itoa(releaseID),
		"--json")
}

func (w *repositoryWriteWave) readIssueComment(commentID int) CLIResult {
	return w.rc.run(w.connector, "issues", "comments", "view",
		"--credential", sourceCredentialName,
		"--comment-id", strconv.Itoa(commentID),
		"--json")
}

// Commit-comment direct reads have the same deliberately redacted body policy
// as issue comments. Bind the verified update to the provider-assigned id and
// provider updated_at rather than pretending a redacted body was observable.
func (w *repositoryWriteWave) verifyCommitCommentUpdated(commentID int, previousUpdatedAt string) (string, error) {
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

func (w *repositoryWriteWave) verifyCommitCommentAbsent(commentID int) (string, error) {
	res := w.readCommitComment(commentID)
	if res.Kind == "Error" && res.ExitCode != 0 && strings.Contains(strings.ToLower(res.Stdout+"\n"+res.Stderr), "404") {
		return fmt.Sprintf("commit_comment:%d:absent", commentID), nil
	}
	if passed, msg := assertKind(w.rc, "write_repository_wave", res, "ConnectorCommandDirectRead", 0); !passed {
		return "", fmt.Errorf("direct read deleted commit comment %d: %s", commentID, msg)
	}
	return "", fmt.Errorf("commit comment %d remains present after delete", commentID)
}

func (w *repositoryWriteWave) readCommitCommentResponse(commentID int) (map[string]any, error) {
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

func (w *repositoryWriteWave) readCommitComment(commentID int) CLIResult {
	return w.rc.run(w.connector, "comments", "view",
		"--credential", sourceCredentialName,
		"--comment-id", strconv.Itoa(commentID),
		"--json")
}

func (w *repositoryWriteWave) verifyIssueLabels(number int, want ...string) (string, error) {
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

func (w *repositoryWriteWave) verifyIssueLabelAbsent(number int, absent string) (string, error) {
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

func (w *repositoryWriteWave) verifyIssueNoLabels(number int) (string, error) {
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

func (w *repositoryWriteWave) fixtureLabels(count int) ([]string, error) {
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

func (w *repositoryWriteWave) firstCommitSHA() (string, error) {
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

func (w *repositoryWriteWave) readStream(stream string) ([]map[string]any, error) {
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
		"--source", w.connector+":"+sourceCredentialName,
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

func (w *repositoryWriteWave) readTopics() ([]string, error) {
	res := w.rc.run(w.connector, "topics", "view", "--credential", sourceCredentialName, "--json")
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

func (w *repositoryWriteWave) markScenarioPassed(tag string, actions []string, reason string) {
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

func (w *repositoryWriteWave) markFailed(action, tag, reason string) {
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
func (w *repositoryWriteWave) markBlocked(actions []string, tag, reason, cleanup string) {
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
func (w *repositoryWriteWave) markRecoveryUncertified(actions []string, tag, reason string) {
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

func (w *repositoryWriteWave) recordScenarioFailure(scenario string, actions []string, tag, reason string) {
	for _, action := range actions {
		if !w.completed[action] {
			w.markFailed(action, tag, scenario+": "+reason)
		}
	}
}

func (w *repositoryWriteWave) recordLeak(tag string, actions []string, reason string) {
	if w.rep.Capabilities.WriteActions == nil {
		w.rep.Capabilities.WriteActions = map[string]WriteActionResult{}
	}
	w.rep.Leaks = append(w.rep.Leaks, Leak{Tag: tag, Connector: w.connector, Action: strings.Join(actions, ","), Reason: reason})
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
