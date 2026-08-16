package certify

import (
	"encoding/json"
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors/defs"
)

type writesFile struct {
	Actions []writeActionDecl `json:"actions"`
}

type writeActionDecl struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Path string `json:"path"`
	Risk string `json:"risk"`
}

type writeActionInventoryItem struct {
	Action         string
	Path           string
	Risk           string
	Pairing        WritePairing
	Classification string
	Reason         string
}

const (
	writeClassificationRepositoryWaveReady                = "repository_wave_ready"
	writeClassificationRepositoryFixturePending           = "repository_fixture_pending"
	writeClassificationGistFixturePending                 = "gist_fixture_pending"
	writeClassificationOrgFixtureAndPermissionPending     = "org_fixture_and_permission_pending"
	writeClassificationAppOrOAuthPending                  = "app_or_oauth_pending"
	writeClassificationEnterpriseTrialAndTokenPending     = "enterprise_trial_and_token_pending"
	writeClassificationPrimaryUserFixtureAndPermission    = "primary_user_fixture_and_permission_pending"
	writeClassificationSecondaryUserFixtureAndPermission  = "secondary_user_fixture_and_permission_pending"
	writeClassificationNotificationTokenAndFixturePending = "notification_token_and_fixture_pending"
	writeClassificationSacrificialCredentialPending       = "sacrificial_credential_pending"
)

// githubRepositoryWaveReadyActions is the captain-approved initial wave. The
// set is explicit because a path alone cannot distinguish a reversible
// run-owned mutation from a repository control-plane action that needs a
// separately constructed fixture/baseline.
var githubRepositoryWaveReadyActions = map[string]struct{}{
	"create_issue": {}, "update_issue": {}, "comment_issue": {}, "close_issue": {}, "reopen_issue": {},
	"create_label": {}, "update_label": {}, "delete_label": {},
	"create_milestone": {}, "update_milestone": {}, "delete_milestone": {},
	"create_release": {}, "update_release": {}, "delete_release": {},
	"create_commit_comment": {}, "update_commit_comment": {}, "delete_commit_comment": {},
	"update_issue_comment": {}, "delete_issue_comment": {}, "lock_issue": {}, "unlock_issue": {},
	"set_issue_labels": {}, "add_issue_labels": {}, "remove_issue_label": {},
	"create_ref": {}, "update_ref": {}, "delete_ref": {}, "replace_repo_topics": {},
}

func declaredWriteActions(connector string) ([]writeActionDecl, error) {
	raw, err := defs.FS.ReadFile(connector + "/writes.json")
	if err != nil {
		return nil, fmt.Errorf("read %s writes: %w", connector, err)
	}
	var file writesFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("parse %s writes: %w", connector, err)
	}
	out := make([]writeActionDecl, 0, len(file.Actions))
	for _, action := range file.Actions {
		name := action.Name
		if name == "" {
			name = action.ID
		}
		if name != "" {
			action.Name = name
			out = append(out, action)
		}
	}
	return out, nil
}

func writeActionInventoryFor(connector string) ([]writeActionInventoryItem, error) {
	declared, err := declaredWriteActions(connector)
	if err != nil {
		return nil, err
	}
	curated := map[string]WritePairing{}
	names := make([]string, 0, len(declared))
	for _, action := range declared {
		names = append(names, action.Name)
	}
	for _, pairing := range PairingsFor(connector) {
		curated[pairing.Create] = pairing
		if pairing.Cleanup != "" {
			curated[pairing.Cleanup] = WritePairing{}
		}
	}

	out := make([]writeActionInventoryItem, 0, len(names))
	for _, declaredAction := range declared {
		name := declaredAction.Name
		item := writeActionInventoryItem{Action: name, Path: declaredAction.Path, Risk: declaredAction.Risk}
		if connector == "github" {
			classification, reason, err := classifyGitHubWriteAction(name, declaredAction.Path)
			if err != nil {
				return nil, err
			}
			item.Classification = classification
			item.Reason = reason
		}
		if pairing, ok := curated[name]; ok && pairing.Create != "" {
			item.Pairing = pairing
			out = append(out, item)
			continue
		}
		if pairing, ok := InferPairing(name, names); ok {
			item.Pairing = pairing
			if item.Reason == "" {
				item.Reason = "inferred create/cleanup pair exists but no curated verify stream/id field is certified for live execution yet"
			}
			out = append(out, item)
			continue
		}
		if item.Reason == "" {
			item.Reason = "not a safe standalone create action with a certified cleanup lifecycle"
		}
		out = append(out, item)
	}
	return out, nil
}

// classifyGitHubWriteAction makes every non-live certification row actionable.
// It is an inventory/precondition classifier, not evidence of a mutation: a
// row is only marked pass after stages_write verifies a production mutation,
// independent read-back, and cleanup. The concrete path remains in the report
// beside this reason and the declaration's own risk text.
func classifyGitHubWriteAction(action, path string) (classification, reason string, err error) {
	if _, ok := githubRepositoryWaveReadyActions[action]; ok {
		return writeClassificationRepositoryWaveReady,
			"the captain-approved repository scenario is not yet implemented for this action", nil
	}

	switch {
	case strings.HasPrefix(path, "/repos/"):
		return writeClassificationRepositoryFixturePending,
			"the run-owned repository fixture addressed by " + path + " has not been created and baseline-captured", nil
	case path == "/gists" || strings.HasPrefix(path, "/gists/"):
		return writeClassificationGistFixturePending,
			"the purpose-made private Gist fixture addressed by " + path + " has not been created", nil
	case strings.HasPrefix(path, "/orgs/"), strings.HasPrefix(path, "/organizations/"), strings.HasPrefix(path, "/teams/"):
		return writeClassificationOrgFixtureAndPermissionPending,
			"Polymetrics-Cert fixture addressed by " + path + " has not been created; the disposable classic test PAT also needs admin:org", nil
	case strings.HasPrefix(path, "/app/"), strings.HasPrefix(path, "/applications/"), strings.HasPrefix(path, "/app-manifests/"), strings.HasPrefix(path, "/installation/"), strings.HasPrefix(path, "/agents/"):
		return writeClassificationAppOrOAuthPending,
			"the dedicated Polymetrics-Cert GitHub App/OAuth fixture required by " + path + " has not been browser-provisioned", nil
	case strings.HasPrefix(path, "/enterprises/"):
		return writeClassificationEnterpriseTrialAndTokenPending,
			"the Polymetrics-Cert Enterprise Cloud trial/slug and fixture addressed by " + path + " are not ready; the disposable classic test PAT needs admin:enterprise", nil
	case path == "/user" || strings.HasPrefix(path, "/user/"):
		return writeClassificationPrimaryUserFixtureAndPermission,
			"the disposable primary-user fixture addressed by " + path + " has not been created with its required classic-PAT scope", nil
	case strings.HasPrefix(path, "/users/"):
		return writeClassificationSecondaryUserFixtureAndPermission,
			"the second disposable-user fixture and credential addressed by " + path + " have not been created", nil
	case strings.HasPrefix(path, "/notifications/") || path == "/notifications":
		return writeClassificationNotificationTokenAndFixturePending,
			"the tagged disposable inbox fixture addressed by " + path + " is absent; this endpoint requires a classic PAT with notifications scope", nil
	case path == "/credentials/revoke":
		return writeClassificationSacrificialCredentialPending,
			"a separate sacrificial disposable credential has not been supplied for the final revocation scenario", nil
	default:
		return "", "", fmt.Errorf("github write action %q path %q has no disposable-boundary classification", action, path)
	}
}
