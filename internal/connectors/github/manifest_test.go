package github_test

import (
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/github"
)

func TestGithubManifestAdvertisesAllSyncModesAndPRDefaults(t *testing.T) {
	manifest := connectors.ManifestOf(github.New())
	for _, want := range []string{"public", "token", "github_app"} {
		if !containsAuthMode(manifest.AuthModes, want) {
			t.Fatalf("github manifest auth modes missing %q: %+v", want, manifest.AuthModes)
		}
	}
	for _, want := range []string{
		"full_refresh_append",
		"full_refresh_overwrite",
		"full_refresh_overwrite_deduped",
		"incremental_append",
		"incremental_append_deduped",
	} {
		if !containsString(manifest.SyncModes, want) {
			t.Fatalf("github manifest sync modes missing %q: %+v", want, manifest.SyncModes)
		}
	}
	if !containsString(manifest.SourceSyncModes, "full_refresh") || !containsString(manifest.SourceSyncModes, "incremental") {
		t.Fatalf("github manifest source modes = %+v", manifest.SourceSyncModes)
	}
	for _, want := range []string{
		"repository",
		"issues",
		"pull_requests",
		"branches",
		"commits",
		"tags",
		"releases",
		"labels",
		"milestones",
		"issue_comments",
		"pull_request_review_comments",
		"collaborators",
		"stargazers",
		"workflows",
		"workflow_runs",
		"workflow_artifacts",
	} {
		if !containsStream(manifest.Streams, want) {
			t.Fatalf("github manifest streams missing %q: %+v", want, manifest.Streams)
		}
	}
	for _, want := range []string{
		"create_issue",
		"create_pull_request",
		"comment_issue",
		"merge_pull_request",
		"create_label",
		"update_label",
		"delete_label",
		"create_milestone",
		"update_milestone",
		"delete_milestone",
		"create_release",
		"update_release",
		"delete_release",
		"dispatch_workflow",
		"rerun_workflow_run",
		"cancel_workflow_run",
		"delete_workflow_run",
		"create_pull_request_review",
		"create_or_update_file",
		"delete_file",
	} {
		if !containsWriteAction(manifest.WriteActions, want) {
			t.Fatalf("github manifest write actions missing %q: %+v", want, manifest.WriteActions)
		}
	}
	var prs connectors.Stream
	for _, stream := range manifest.Streams {
		if stream.Name == "pull_requests" {
			prs = stream
			break
		}
	}
	if prs.Name == "" {
		t.Fatal("github manifest missing pull_requests stream")
	}
	if !containsString(prs.PrimaryKey, "node_id") {
		t.Fatalf("pull_requests primary key = %+v, want node_id", prs.PrimaryKey)
	}
	if !containsString(prs.CursorFields, "updated_at") {
		t.Fatalf("pull_requests cursor fields = %+v, want updated_at", prs.CursorFields)
	}
}

func TestGithubGuideIncludesAuthStreamsActionsLinksAndExamples(t *testing.T) {
	manual := connectors.RenderConnectorManual(github.New())
	for _, want := range []string{
		"AUTHENTICATION",
		"github_app",
		"--value-stdin private_key",
		"ETL STREAMS",
		"pull_requests",
		"REVERSE ETL ACTIONS",
		"create_pull_request",
		"merge_pull_request",
		"SEE ALSO",
		"https://docs.github.com/en/rest/authentication/authenticating-to-the-rest-api",
		"pm reverse preview <plan-id> --json",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("github guide missing %q:\n%s", want, manual)
		}
	}
}

func containsStream(values []connectors.Stream, want string) bool {
	for _, value := range values {
		if value.Name == want {
			return true
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsWriteAction(values []connectors.WriteActionSpec, want string) bool {
	for _, value := range values {
		if value.Name == want {
			return true
		}
	}
	return false
}

func containsAuthMode(values []connectors.AuthModeSpec, want string) bool {
	for _, value := range values {
		if value.Name == want {
			return true
		}
	}
	return false
}
