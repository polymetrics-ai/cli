package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
)

func TestGitHubRequiredMutationBodiesReachTheWire(t *testing.T) {
	bundle, err := Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load(defs.FS, github): %v", err)
	}

	tests := []struct {
		name     string
		command  string
		action   string
		record   connectors.Record
		want     map[string]any
		required map[string][]string
	}{
		{
			name: "blob", command: "git blobs create", action: "git_blobs",
			record:   connectors.Record{"content": "pm-cert-content", "encoding": "utf-8"},
			want:     map[string]any{"content": "pm-cert-content", "encoding": "utf-8"},
			required: map[string][]string{"content": {"string"}},
		},
		{
			name: "commit", command: "git commits create", action: "git_commits",
			record:   connectors.Record{"message": "pm-cert-commit", "tree": strings.Repeat("a", 40), "parents": []string{strings.Repeat("b", 40)}},
			want:     map[string]any{"message": "pm-cert-commit", "tree": strings.Repeat("a", 40), "parents": []any{strings.Repeat("b", 40)}},
			required: map[string][]string{"message": {"string"}, "tree": {"string"}},
		},
		{
			name: "tree", command: "git trees create", action: "git_trees",
			record:   connectors.Record{"tree": []any{map[string]any{"path": "pm-cert.txt", "mode": "100644", "type": "blob", "content": "fixture"}}},
			want:     map[string]any{"tree": []any{map[string]any{"path": "pm-cert.txt", "mode": "100644", "type": "blob", "content": "fixture"}}},
			required: map[string][]string{"tree": {"array"}},
		},
		{
			name: "check run", command: "check-runs create", action: "check_runs",
			record:   connectors.Record{"name": "pm-cert-check", "head_sha": strings.Repeat("c", 40)},
			want:     map[string]any{"name": "pm-cert-check", "head_sha": strings.Repeat("c", 40)},
			required: map[string][]string{"name": {"string"}, "head_sha": {"string"}},
		},
		{
			name: "branch protection", command: "branches protection set", action: "branches_branch_protection3",
			record: connectors.Record{
				"branch":                        "main",
				"required_status_checks":        map[string]any{"strict": false, "contexts": []any{}},
				"enforce_admins":                false,
				"required_pull_request_reviews": map[string]any{"dismiss_stale_reviews": false, "require_code_owner_reviews": false, "required_approving_review_count": 0},
				"restrictions":                  map[string]any{"users": []any{}, "teams": []any{}, "apps": []any{}},
			},
			want: map[string]any{
				"required_status_checks":        map[string]any{"strict": false, "contexts": []any{}},
				"enforce_admins":                false,
				"required_pull_request_reviews": map[string]any{"dismiss_stale_reviews": false, "require_code_owner_reviews": false, "required_approving_review_count": float64(0)},
				"restrictions":                  map[string]any{"users": []any{}, "teams": []any{}, "apps": []any{}},
			},
			required: map[string][]string{
				"branch": {"string"}, "required_status_checks": {"object", "null"}, "enforce_admins": {"boolean", "null"},
				"required_pull_request_reviews": {"object", "null"}, "restrictions": {"object", "null"},
			},
		},
		{
			name: "commit status", command: "statuses create", action: "statuses_sha",
			record:   connectors.Record{"sha": strings.Repeat("d", 40), "state": "success", "context": "pm-cert/context"},
			want:     map[string]any{"state": "success", "context": "pm-cert/context"},
			required: map[string][]string{"sha": {"string"}, "state": {"string"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, err := findWriteAction(bundle, tt.action)
			if err != nil {
				t.Fatalf("findWriteAction(%q): %v", tt.action, err)
			}
			assertGitHubWriteFields(t, bundle, action, tt.command, tt.required)
			server, captured := captureServer(t, http.StatusCreated, `{}`)
			action.Path = "/fixture"
			fixture := bundle
			fixture.HTTP.URL = server.URL
			fixture.HTTP.Auth = []AuthSpec{{Mode: "none"}}
			fixture.Writes = []WriteAction{action}

			records := []connectors.Record{tt.record}
			request := connectors.WriteRequest{Action: tt.action}
			_, err = prepareDeclarativeWrite(context.Background(), fixture, request, records, nil)
			if err != nil {
				t.Fatalf("prepareDeclarativeWrite(%q): %v", tt.action, err)
			}
			runtime, err := newRuntime(context.Background(), fixture, connectors.RuntimeConfig{}, nil)
			if err != nil {
				t.Fatalf("newRuntime(%q): %v", tt.action, err)
			}
			if err := executeWriteRecord(context.Background(), fixture, action, tt.record, 0, connectors.RuntimeConfig{}, runtime); err != nil {
				t.Fatalf("executeWriteRecord(%q): %v", tt.action, err)
			}
			var got map[string]any
			if err := json.Unmarshal(captured.body, &got); err != nil {
				t.Fatalf("decode captured body %q: %v", captured.body, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("%s body = %#v, want %#v", tt.name, got, tt.want)
			}
		})
	}
}

func assertGitHubWriteFields(t *testing.T, bundle Bundle, action WriteAction, commandPath string, required map[string][]string) {
	t.Helper()
	var schema struct {
		Required   []string `json:"required"`
		Properties map[string]struct {
			Type json.RawMessage `json:"type"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(action.RecordSchema, &schema); err != nil {
		t.Fatalf("decode %s record schema: %v", action.Name, err)
	}
	requiredSet := make(map[string]bool, len(schema.Required))
	for _, field := range schema.Required {
		requiredSet[field] = true
	}
	command, ok := findGitHubCLICommand(bundle, commandPath)
	if !ok {
		t.Fatalf("GitHub command %q is missing", commandPath)
	}
	for field, wantTypes := range required {
		property, ok := schema.Properties[field]
		gotTypes, err := githubSchemaTypes(property.Type)
		if err != nil {
			t.Errorf("action %q field %q has invalid type %s: %v", action.Name, field, property.Type, err)
		} else if !ok || !reflect.DeepEqual(gotTypes, wantTypes) || !requiredSet[field] {
			t.Errorf("action %q field %q types=%#v required=%t, want types %#v and required", action.Name, field, gotTypes, requiredSet[field], wantTypes)
		}
		flagName := strings.ReplaceAll(field, "_", "-")
		found := false
		for _, flag := range command.Flags {
			if flag.Name == flagName && flag.MapsTo == "record."+field && flag.Required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("command %q has no required --%s mapping to record.%s", commandPath, flagName, field)
		}
	}
}

func githubSchemaTypes(raw json.RawMessage) ([]string, error) {
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return []string{single}, nil
	}
	var multiple []string
	if err := json.Unmarshal(raw, &multiple); err != nil {
		return nil, err
	}
	return multiple, nil
}

func TestGitHubCorrectedCommandDeclarations(t *testing.T) {
	bundle, err := Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load(defs.FS, github): %v", err)
	}

	wantActionsAliases := map[string]struct {
		write string
		path  string
	}{
		"agents set-selected-repos-for-org-secret": {
			write: "actions_set_selected_repos_for_org_secret",
			path:  "/orgs/{org}/actions/secrets/{secret_name}/repositories",
		},
		"agents set-selected-repos-for-org-variable": {
			write: "actions_set_selected_repos_for_org_variable",
			path:  "/orgs/{org}/actions/variables/{name}/repositories",
		},
		"agents update-org-variable": {
			write: "actions_update_org_variable",
			path:  "/orgs/{org}/actions/variables/{name}",
		},
	}
	for path, want := range wantActionsAliases {
		command, ok := findGitHubCLICommand(bundle, path)
		if !ok {
			t.Fatalf("GitHub command %q is missing", path)
		}
		if command.Write != want.write {
			t.Errorf("command %q write = %q, want %q", path, command.Write, want.write)
		}
		if len(command.APISurface) != 1 || command.APISurface[0].Path != want.path {
			t.Errorf("command %q api_surface = %#v, want %q", path, command.APISurface, want.path)
		}
	}

	command, ok := findGitHubCLICommand(bundle, "projects create-draft-item-for-authenticated-user")
	if !ok {
		t.Fatal("GitHub user draft command is missing")
	}
	if command.Intent != "direct_write" || command.Operation != "github.graphql.mutation.add-project-v2-draft-issue" || command.Write != "" {
		t.Fatalf("user draft binding = intent %q operation %q write %q, want fixed GraphQL direct write", command.Intent, command.Operation, command.Write)
	}
	if len(command.APISurface) != 1 || command.APISurface[0].Method != "POST" || command.APISurface[0].Path != "/graphql" {
		t.Fatalf("user draft api_surface = %#v, want POST /graphql", command.APISurface)
	}
	if len(command.Flags) != 1 || command.Flags[0].Name != "input" || command.Flags[0].Type != "json" || command.Flags[0].MapsTo != "body.input" || !command.Flags[0].Required {
		t.Fatalf("user draft flags = %#v, want one required closed --input JSON variable", command.Flags)
	}
	for _, action := range bundle.Writes {
		if action.Name == "projects_create_draft_item_for_authenticated_user" {
			t.Fatal("user draft still exposes the nonexistent REST write action")
		}
	}
}

func findGitHubCLICommand(bundle Bundle, path string) (CLICommand, bool) {
	if bundle.CLISurface == nil {
		return CLICommand{}, false
	}
	for _, command := range bundle.CLISurface.Commands {
		if command.Path == path {
			return command, true
		}
	}
	return CLICommand{}, false
}
