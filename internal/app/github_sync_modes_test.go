package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"polymetrics/internal/connectors"
)

func TestGithubPullRequestsETLSupportsAllSyncModes(t *testing.T) {
	tests := []struct {
		name string
		mode string
		run  func(t *testing.T, a *App, connection string, setRecords func([]map[string]any))
	}{
		{
			name: "full refresh append duplicates",
			mode: "full_refresh_append",
			run: func(t *testing.T, a *App, connection string, setRecords func([]map[string]any)) {
				setRecords([]map[string]any{
					githubPRFixture("PR_1", 1, "first", "2026-01-01T00:00:00Z"),
					githubPRFixture("PR_2", 2, "second", "2026-01-02T00:00:00Z"),
				})
				runGithubETL(t, a, connection)
				runGithubETL(t, a, connection)
				assertGithubRows(t, a, 4, map[string]string{})
			},
		},
		{
			name: "full refresh overwrite replaces final",
			mode: "full_refresh_overwrite",
			run: func(t *testing.T, a *App, connection string, setRecords func([]map[string]any)) {
				setRecords([]map[string]any{
					githubPRFixture("PR_1", 1, "first", "2026-01-01T00:00:00Z"),
					githubPRFixture("PR_2", 2, "second", "2026-01-02T00:00:00Z"),
				})
				runGithubETL(t, a, connection)
				setRecords([]map[string]any{
					githubPRFixture("PR_3", 3, "third", "2026-01-03T00:00:00Z"),
				})
				runGithubETL(t, a, connection)
				assertGithubRows(t, a, 1, map[string]string{"PR_3": "third"})
			},
		},
		{
			name: "full refresh overwrite deduped keeps latest duplicate",
			mode: "full_refresh_overwrite_deduped",
			run: func(t *testing.T, a *App, connection string, setRecords func([]map[string]any)) {
				setRecords([]map[string]any{
					githubPRFixture("PR_1", 1, "first old", "2026-01-01T00:00:00Z"),
					githubPRFixture("PR_1", 1, "first latest", "2026-01-03T00:00:00Z"),
					githubPRFixture("PR_2", 2, "second", "2026-01-02T00:00:00Z"),
				})
				runGithubETL(t, a, connection)
				assertGithubRows(t, a, 2, map[string]string{"PR_1": "first latest", "PR_2": "second"})
			},
		},
		{
			name: "incremental append filters older cursor and appends inclusive cursor",
			mode: "incremental_append",
			run: func(t *testing.T, a *App, connection string, setRecords func([]map[string]any)) {
				setRecords([]map[string]any{
					githubPRFixture("PR_1", 1, "first", "2026-01-01T00:00:00Z"),
					githubPRFixture("PR_2", 2, "second", "2026-01-02T00:00:00Z"),
				})
				runGithubETL(t, a, connection)
				setRecords([]map[string]any{
					githubPRFixture("PR_1", 1, "first older", "2025-12-31T00:00:00Z"),
					githubPRFixture("PR_2", 2, "second resent", "2026-01-02T00:00:00Z"),
					githubPRFixture("PR_3", 3, "third", "2026-01-03T00:00:00Z"),
				})
				run := runGithubETL(t, a, connection)
				if run.Checkpoint["cursor"] != "2026-01-03T00:00:00Z" {
					t.Fatalf("cursor = %q, want 2026-01-03T00:00:00Z", run.Checkpoint["cursor"])
				}
				assertGithubRows(t, a, 4, map[string]string{})
			},
		},
		{
			name: "incremental append deduped materializes latest PR rows",
			mode: "incremental_append_deduped",
			run: func(t *testing.T, a *App, connection string, setRecords func([]map[string]any)) {
				setRecords([]map[string]any{
					githubPRFixture("PR_1", 1, "first", "2026-01-01T00:00:00Z"),
					githubPRFixture("PR_2", 2, "second", "2026-01-02T00:00:00Z"),
				})
				runGithubETL(t, a, connection)
				setRecords([]map[string]any{
					githubPRFixture("PR_1", 1, "first updated", "2026-01-03T00:00:00Z"),
					githubPRFixture("PR_3", 3, "third", "2026-01-04T00:00:00Z"),
				})
				runGithubETL(t, a, connection)
				assertGithubRows(t, a, 3, map[string]string{"PR_1": "first updated", "PR_2": "second", "PR_3": "third"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, connection, setRecords := setupGithubSyncModeApp(t, tt.mode)
			tt.run(t, a, connection, setRecords)
		})
	}
}

func setupGithubSyncModeApp(t *testing.T, mode string) (*App, string, func([]map[string]any)) {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	currentRecords := []map[string]any{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/acme/widgets/pulls" {
			t.Fatalf("path = %s, want /repos/acme/widgets/pulls", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(currentRecords)
	}))
	t.Cleanup(server.Close)

	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{
		Name:      "github-local",
		Connector: "github",
		Config: map[string]string{
			"repository": "acme/widgets",
			"base_url":   server.URL,
			"max_pages":  "1",
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{
		Name:      "warehouse-local",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatal(err)
	}
	conn, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "github_prs_to_warehouse",
		Source:      EndpointConfig{Connector: "github", Credential: "github-local"},
		Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse-local"},
		Streams: map[string]StreamConfig{
			"pull_requests": {
				SyncMode:         mode,
				DestinationTable: "github_prs",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	stream := conn.Streams["pull_requests"]
	if stream.CursorField != "updated_at" {
		t.Fatalf("cursor default = %q, want updated_at", stream.CursorField)
	}
	if len(stream.PrimaryKey) != 1 || stream.PrimaryKey[0] != "node_id" {
		t.Fatalf("primary key default = %+v, want [node_id]", stream.PrimaryKey)
	}
	setRecords := func(records []map[string]any) {
		currentRecords = records
	}
	return a, conn.Name, setRecords
}

func runGithubETL(t *testing.T, a *App, connection string) Run {
	t.Helper()
	run, err := a.RunETL(context.Background(), RunETLRequest{
		Connection: connection,
		Stream:     "pull_requests",
		BatchSize:  2,
	})
	if err != nil {
		t.Fatal(err)
	}
	return run
}

func assertGithubRows(t *testing.T, a *App, count int, titles map[string]string) {
	t.Helper()
	rows, err := a.QueryTable(context.Background(), QueryTableRequest{Table: "github_prs", Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != count {
		t.Fatalf("rows len = %d, want %d: %+v", len(rows), count, rows)
	}
	byNodeID := map[string]connectors.Record{}
	for _, row := range rows {
		nodeID, _ := row["node_id"].(string)
		if nodeID != "" {
			byNodeID[nodeID] = row
		}
	}
	for nodeID, title := range titles {
		row, ok := byNodeID[nodeID]
		if !ok {
			t.Fatalf("missing node_id %s in rows %+v", nodeID, rows)
		}
		if row["title"] != title {
			t.Fatalf("title for %s = %v, want %s", nodeID, row["title"], title)
		}
	}
}

func githubPRFixture(nodeID string, number int, title, updatedAt string) map[string]any {
	return map[string]any{
		"id":                 number,
		"node_id":            nodeID,
		"number":             number,
		"state":              "open",
		"title":              title,
		"body":               "",
		"html_url":           "https://github.test/acme/widgets/pull/" + toComparableString(number),
		"url":                "https://api.github.test/repos/acme/widgets/pulls/" + toComparableString(number),
		"user":               map[string]any{"login": "octocat", "id": 1},
		"author_association": "CONTRIBUTOR",
		"comments":           0,
		"locked":             false,
		"created_at":         "2025-12-01T00:00:00Z",
		"updated_at":         updatedAt,
		"closed_at":          nil,
		"merged_at":          nil,
		"draft":              false,
		"merge_commit_sha":   "",
		"base":               map[string]any{"ref": "main", "sha": "base-sha"},
		"head":               map[string]any{"ref": "feature", "sha": "head-sha"},
	}
}
