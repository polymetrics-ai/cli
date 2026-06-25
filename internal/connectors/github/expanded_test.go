package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestGithubReadExpandedRepositoryStreams(t *testing.T) {
	paths := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/repos/acme/widgets":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": 1, "node_id": "R_1", "name": "widgets", "full_name": "acme/widgets",
				"private": false, "html_url": "https://github.test/acme/widgets",
				"default_branch": "main", "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z",
			})
		case "/repos/acme/widgets/branches":
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"name": "main", "protected": true, "commit": map[string]any{"sha": "abc123", "url": "https://api.github.test/commits/abc123"},
			}})
		case "/repos/acme/widgets/actions/runs":
			_ = json.NewEncoder(w).Encode(map[string]any{"workflow_runs": []map[string]any{{
				"id": 99, "name": "ci", "head_branch": "main", "head_sha": "abc123", "status": "completed", "conclusion": "success",
				"event": "push", "html_url": "https://github.test/acme/widgets/actions/runs/99", "created_at": "2026-01-03T00:00:00Z", "updated_at": "2026-01-03T00:10:00Z",
			}}})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"repository": "acme/widgets", "base_url": server.URL}}
	tests := []struct {
		stream string
		want   connectors.Record
	}{
		{"repository", connectors.Record{"full_name": "acme/widgets", "default_branch": "main"}},
		{"branches", connectors.Record{"name": "main", "commit_sha": "abc123"}},
		{"workflow_runs", connectors.Record{"name": "ci", "status": "completed", "conclusion": "success"}},
	}
	for _, tt := range tests {
		t.Run(tt.stream, func(t *testing.T) {
			var records []connectors.Record
			err := (Connector{}).Read(context.Background(), connectors.ReadRequest{Stream: tt.stream, Config: cfg}, func(record connectors.Record) error {
				records = append(records, record)
				return nil
			})
			if err != nil {
				t.Fatalf("Read(%s) error = %v", tt.stream, err)
			}
			if len(records) != 1 {
				t.Fatalf("records len = %d, want 1: %+v", len(records), records)
			}
			for key, want := range tt.want {
				if records[0][key] != want {
					t.Fatalf("record[%s] = %v, want %v; record=%+v", key, records[0][key], want, records[0])
				}
			}
		})
	}
	if got := strings.Join(paths, ","); !strings.Contains(got, "/repos/acme/widgets/actions/runs") {
		t.Fatalf("workflow runs endpoint not requested: %v", paths)
	}
}

func TestGithubReadAllAdvertisedStreamsResolveAPI(t *testing.T) {
	requested := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested[r.URL.Path]++
		switch r.URL.Path {
		case "/repos/acme/widgets":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": 1, "node_id": "R_1", "name": "widgets", "full_name": "acme/widgets",
				"default_branch": "main", "updated_at": "2026-01-02T00:00:00Z",
			})
		case "/repos/acme/widgets/issues":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 10, "node_id": "I_1", "number": 1, "title": "issue", "state": "open"}})
		case "/repos/acme/widgets/pulls":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 11, "node_id": "PR_1", "number": 2, "title": "pr", "state": "open"}})
		case "/repos/acme/widgets/branches":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"name": "main", "commit": map[string]any{"sha": "abc123"}}})
		case "/repos/acme/widgets/commits":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"sha": "abc123", "commit": map[string]any{"message": "init", "committer": map[string]any{"date": "2026-01-02T00:00:00Z"}}}})
		case "/repos/acme/widgets/tags":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"name": "v1.0.0", "commit": map[string]any{"sha": "abc123"}}})
		case "/repos/acme/widgets/releases":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 12, "tag_name": "v1.0.0", "name": "v1", "published_at": "2026-01-03T00:00:00Z"}})
		case "/repos/acme/widgets/labels":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 13, "name": "agentic", "color": "5319e7"}})
		case "/repos/acme/widgets/milestones":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 14, "number": 3, "title": "v2", "state": "open"}})
		case "/repos/acme/widgets/issues/comments":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 15, "body": "comment"}})
		case "/repos/acme/widgets/pulls/comments":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 16, "path": "main.go", "body": "review"}})
		case "/repos/acme/widgets/collaborators":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 17, "login": "ada", "role_name": "admin"}})
		case "/repos/acme/widgets/contributors":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 18, "login": "grace", "contributions": 42}})
		case "/repos/acme/widgets/stargazers":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 19, "login": "linus"}})
		case "/repos/acme/widgets/subscribers":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 20, "login": "margaret"}})
		case "/repos/acme/widgets/actions/workflows":
			_ = json.NewEncoder(w).Encode(map[string]any{"workflows": []map[string]any{{"id": 21, "name": "ci", "path": ".github/workflows/ci.yml"}}})
		case "/repos/acme/widgets/actions/runs":
			_ = json.NewEncoder(w).Encode(map[string]any{"workflow_runs": []map[string]any{{"id": 22, "name": "ci", "status": "completed"}}})
		case "/repos/acme/widgets/actions/artifacts":
			_ = json.NewEncoder(w).Encode(map[string]any{"artifacts": []map[string]any{{"id": 23, "name": "dist"}}})
		case "/repos/acme/widgets/deployments":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 24, "sha": "abc123", "environment": "production"}})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	cfg := connectors.RuntimeConfig{Config: map[string]string{"repository": "acme/widgets", "base_url": server.URL}}
	for _, stream := range githubStreams() {
		t.Run(stream.Name, func(t *testing.T) {
			var records []connectors.Record
			err := (Connector{}).Read(context.Background(), connectors.ReadRequest{Stream: stream.Name, Config: cfg}, func(record connectors.Record) error {
				records = append(records, record)
				return nil
			})
			if err != nil {
				t.Fatalf("Read(%s) error = %v", stream.Name, err)
			}
			if len(records) != 1 {
				t.Fatalf("Read(%s) emitted %d records, want 1", stream.Name, len(records))
			}
		})
	}
	for _, path := range []string{
		"/repos/acme/widgets",
		"/repos/acme/widgets/issues",
		"/repos/acme/widgets/pulls",
		"/repos/acme/widgets/branches",
		"/repos/acme/widgets/commits",
		"/repos/acme/widgets/tags",
		"/repos/acme/widgets/releases",
		"/repos/acme/widgets/labels",
		"/repos/acme/widgets/milestones",
		"/repos/acme/widgets/issues/comments",
		"/repos/acme/widgets/pulls/comments",
		"/repos/acme/widgets/collaborators",
		"/repos/acme/widgets/contributors",
		"/repos/acme/widgets/stargazers",
		"/repos/acme/widgets/subscribers",
		"/repos/acme/widgets/actions/workflows",
		"/repos/acme/widgets/actions/runs",
		"/repos/acme/widgets/actions/artifacts",
		"/repos/acme/widgets/deployments",
	} {
		if requested[path] == 0 {
			t.Fatalf("path %s was not requested", path)
		}
	}
}

func TestGithubWriteExpandedReverseActions(t *testing.T) {
	tests := []struct {
		name       string
		action     string
		record     connectors.Record
		wantMethod string
		wantPath   string
	}{
		{
			name:       "create issue",
			action:     "create_issue",
			record:     connectors.Record{"title": "Bug", "body": "Created by pm"},
			wantMethod: http.MethodPost,
			wantPath:   "/repos/acme/widgets/issues",
		},
		{
			name:       "update issue",
			action:     "update_issue",
			record:     connectors.Record{"issue_number": 4, "title": "Updated"},
			wantMethod: http.MethodPatch,
			wantPath:   "/repos/acme/widgets/issues/4",
		},
		{
			name:       "comment issue",
			action:     "comment_issue",
			record:     connectors.Record{"issue_number": 4, "body": "Looks good"},
			wantMethod: http.MethodPost,
			wantPath:   "/repos/acme/widgets/issues/4/comments",
		},
		{
			name:       "close issue",
			action:     "close_issue",
			record:     connectors.Record{"issue_number": 4},
			wantMethod: http.MethodPatch,
			wantPath:   "/repos/acme/widgets/issues/4",
		},
		{
			name:       "create pull request",
			action:     "create_pull_request",
			record:     connectors.Record{"title": "Ship", "head": "feature", "base": "main", "body": "Ready"},
			wantMethod: http.MethodPost,
			wantPath:   "/repos/acme/widgets/pulls",
		},
		{
			name:       "update pull request",
			action:     "update_pull_request",
			record:     connectors.Record{"pull_number": 7, "title": "Ship v2"},
			wantMethod: http.MethodPatch,
			wantPath:   "/repos/acme/widgets/pulls/7",
		},
		{
			name:       "close pull request",
			action:     "close_pull_request",
			record:     connectors.Record{"pull_number": 7},
			wantMethod: http.MethodPatch,
			wantPath:   "/repos/acme/widgets/pulls/7",
		},
		{
			name:       "request reviewers",
			action:     "request_reviewers",
			record:     connectors.Record{"pull_number": 7, "reviewers": "ada,grace"},
			wantMethod: http.MethodPost,
			wantPath:   "/repos/acme/widgets/pulls/7/requested_reviewers",
		},
		{
			name:       "merge pull request",
			action:     "merge_pull_request",
			record:     connectors.Record{"pull_number": 7, "merge_method": "squash"},
			wantMethod: http.MethodPut,
			wantPath:   "/repos/acme/widgets/pulls/7/merge",
		},
		{
			name:       "create label",
			action:     "create_label",
			record:     connectors.Record{"name": "agentic", "color": "5319e7", "description": "Created by pm"},
			wantMethod: http.MethodPost,
			wantPath:   "/repos/acme/widgets/labels",
		},
		{
			name:       "update label",
			action:     "update_label",
			record:     connectors.Record{"name": "agentic", "new_name": "pm", "color": "#0e8a16"},
			wantMethod: http.MethodPatch,
			wantPath:   "/repos/acme/widgets/labels/agentic",
		},
		{
			name:       "delete label",
			action:     "delete_label",
			record:     connectors.Record{"name": "pm"},
			wantMethod: http.MethodDelete,
			wantPath:   "/repos/acme/widgets/labels/pm",
		},
		{
			name:       "create milestone",
			action:     "create_milestone",
			record:     connectors.Record{"title": "v2", "state": "open"},
			wantMethod: http.MethodPost,
			wantPath:   "/repos/acme/widgets/milestones",
		},
		{
			name:       "update milestone",
			action:     "update_milestone",
			record:     connectors.Record{"milestone_number": 2, "title": "v2", "state": "open"},
			wantMethod: http.MethodPatch,
			wantPath:   "/repos/acme/widgets/milestones/2",
		},
		{
			name:       "delete milestone",
			action:     "delete_milestone",
			record:     connectors.Record{"milestone_number": 2},
			wantMethod: http.MethodDelete,
			wantPath:   "/repos/acme/widgets/milestones/2",
		},
		{
			name:       "create release",
			action:     "create_release",
			record:     connectors.Record{"tag_name": "v2.0.0", "name": "v2", "draft": true},
			wantMethod: http.MethodPost,
			wantPath:   "/repos/acme/widgets/releases",
		},
		{
			name:       "update release",
			action:     "update_release",
			record:     connectors.Record{"release_id": 44, "name": "v2.0.1"},
			wantMethod: http.MethodPatch,
			wantPath:   "/repos/acme/widgets/releases/44",
		},
		{
			name:       "delete release",
			action:     "delete_release",
			record:     connectors.Record{"release_id": 44},
			wantMethod: http.MethodDelete,
			wantPath:   "/repos/acme/widgets/releases/44",
		},
		{
			name:       "dispatch workflow",
			action:     "dispatch_workflow",
			record:     connectors.Record{"workflow_id": "ci.yml", "ref": "main", "inputs": `{"env":"test"}`},
			wantMethod: http.MethodPost,
			wantPath:   "/repos/acme/widgets/actions/workflows/ci.yml/dispatches",
		},
		{
			name:       "rerun workflow run",
			action:     "rerun_workflow_run",
			record:     connectors.Record{"run_id": 55},
			wantMethod: http.MethodPost,
			wantPath:   "/repos/acme/widgets/actions/runs/55/rerun",
		},
		{
			name:       "cancel workflow run",
			action:     "cancel_workflow_run",
			record:     connectors.Record{"run_id": 55},
			wantMethod: http.MethodPost,
			wantPath:   "/repos/acme/widgets/actions/runs/55/cancel",
		},
		{
			name:       "delete workflow run",
			action:     "delete_workflow_run",
			record:     connectors.Record{"run_id": 55},
			wantMethod: http.MethodDelete,
			wantPath:   "/repos/acme/widgets/actions/runs/55",
		},
		{
			name:       "create pull request review",
			action:     "create_pull_request_review",
			record:     connectors.Record{"pull_number": 7, "event": "approve", "body": "Approved"},
			wantMethod: http.MethodPost,
			wantPath:   "/repos/acme/widgets/pulls/7/reviews",
		},
		{
			name:       "create file",
			action:     "create_or_update_file",
			record:     connectors.Record{"path": "docs/generated.md", "message": "Update docs", "content": "# Generated", "branch": "main"},
			wantMethod: http.MethodPut,
			wantPath:   "/repos/acme/widgets/contents/docs%2Fgenerated.md",
		},
		{
			name:       "delete file",
			action:     "delete_file",
			record:     connectors.Record{"path": "docs/generated.md", "message": "Remove docs", "sha": "abc123"},
			wantMethod: http.MethodDelete,
			wantPath:   "/repos/acme/widgets/contents/docs%2Fgenerated.md",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotBody map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.wantMethod || r.URL.EscapedPath() != tt.wantPath {
					t.Fatalf("%s %s, want %s %s", r.Method, r.URL.EscapedPath(), tt.wantMethod, tt.wantPath)
				}
				if r.Body != nil {
					_ = json.NewDecoder(r.Body).Decode(&gotBody)
				}
				if r.Method == http.MethodDelete {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				if tt.action == "create_pull_request" {
					_ = json.NewEncoder(w).Encode(map[string]any{"number": 7})
					return
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
			}))
			defer server.Close()

			cfg := connectors.RuntimeConfig{Config: map[string]string{"repository": "acme/widgets", "base_url": server.URL}, Secrets: map[string]string{"token": "secret-token"}}
			result, err := (Connector{}).Write(context.Background(), connectors.WriteRequest{Action: tt.action, Config: cfg}, []connectors.Record{tt.record})
			if err != nil {
				t.Fatalf("Write(%s) error = %v", tt.action, err)
			}
			if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
				t.Fatalf("result = %+v", result)
			}
			if tt.action == "create_or_update_file" && gotBody["content"] == "# Generated" {
				t.Fatalf("file content was not base64 encoded: %+v", gotBody)
			}
		})
	}
}
