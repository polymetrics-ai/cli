package cli_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
)

func TestGitLabCommandSurfaceAdvertisesOnlyCitedReadCommands(t *testing.T) {
	registry := bundleregistry.New()
	connector, ok := registry.Get("gitlab")
	if !ok {
		t.Fatal("GitLab connector is not registered")
	}

	manifest := connectors.ManifestOf(connector)
	if !manifest.Metadata.Capabilities.Read || manifest.Metadata.Capabilities.Write || len(manifest.WriteActions) != 0 {
		t.Fatalf("GitLab executable capabilities = %+v with %d write actions, want read-only", manifest.Metadata.Capabilities, len(manifest.WriteActions))
	}

	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		t.Fatal("GitLab connector has no command surface")
	}
	want := map[string]struct {
		stream string
		method string
		path   string
		url    string
	}{
		"projects list": {stream: "projects", method: http.MethodGet, path: "/projects", url: "https://docs.gitlab.com/api/projects/"},
		"groups list":   {stream: "groups", method: http.MethodGet, path: "/groups", url: "https://docs.gitlab.com/api/groups/"},
		"users list":    {stream: "users", method: http.MethodGet, path: "/users", url: "https://docs.gitlab.com/api/users/"},
		"issues list":   {stream: "issues", method: http.MethodGet, path: "/issues", url: "https://docs.gitlab.com/api/issues/"},
	}

	surface := provider.CommandSurface()
	if len(surface.Commands) != len(want) {
		t.Fatalf("GitLab command count = %d, want %d", len(surface.Commands), len(want))
	}
	for _, command := range surface.Commands {
		expectation, ok := want[command.Path]
		if !ok {
			t.Fatalf("unexpected GitLab command %q", command.Path)
		}
		if command.Intent != "etl" || command.Availability != "implemented" || command.Stream != expectation.stream || command.SourceCLIPath != expectation.method+" "+expectation.path || command.SourceURL != expectation.url {
			t.Fatalf("command %q = %+v, want implemented %s stream with provider citation %s", command.Path, command, expectation.stream, expectation.url)
		}
		if len(command.APISurface) != 1 || command.APISurface[0].Method != expectation.method || command.APISurface[0].Path != expectation.path {
			t.Fatalf("command %q API surface = %+v, want %s %s", command.Path, command.APISurface, expectation.method, expectation.path)
		}
	}
}

func TestGitLabCommandSurfaceRunsAllDeclaredStreams(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		command []string
		stream  string
	}{
		{name: "projects", path: "/projects", command: []string{"projects", "list"}, stream: "projects"},
		{name: "groups", path: "/groups", command: []string{"groups", "list"}, stream: "groups"},
		{name: "users", path: "/users", command: []string{"users", "list"}, stream: "users"},
		{name: "issues", path: "/issues", command: []string{"issues", "list"}, stream: "issues"},
	}

	requests := make(map[string]int, len(tests))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		requests[r.URL.Path]++
		if r.URL.Query().Get("per_page") != "50" {
			t.Errorf("per_page = %q, want 50", r.URL.Query().Get("per_page"))
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Error("request did not carry bearer authentication")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1}]`))
	}))
	t.Cleanup(server.Close)

	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	t.Setenv("PM_TEST_GITLAB_ACCESS_TOKEN", "fixture_credential_placeholder")
	runCLI(t, []string{
		"credentials", "add", "gitlab-local",
		"--connector", "gitlab",
		"--from-env", "access_token=PM_TEST_GITLAB_ACCESS_TOKEN",
		"--config", "base_url=" + server.URL,
		"--root", root,
		"--json",
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _ := runCLI(t, append([]string{
				"gitlab",
			}, append(tt.command,
				"--credential", "gitlab-local",
				"--limit", "1",
				"--root", root,
				"--json",
			)...))

			var envelope struct {
				Kind    string           `json:"kind"`
				Command string           `json:"command"`
				Stream  string           `json:"stream"`
				Count   int              `json:"count"`
				Records []map[string]any `json:"records"`
			}
			if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
				t.Fatalf("decode command output: %v\n%s", err, stdout)
			}
			if envelope.Kind != "ConnectorCommandRead" || envelope.Command != strings.Join(tt.command, " ") || envelope.Stream != tt.stream || envelope.Count != 1 {
				t.Fatalf("envelope = %+v, want one %s ConnectorCommandRead", envelope, tt.stream)
			}
			if len(envelope.Records) != 1 || envelope.Records[0]["id"] == nil {
				t.Fatalf("records = %+v, want one projected record with id", envelope.Records)
			}
		})
	}

	for _, tt := range tests {
		if requests[tt.path] != 1 {
			t.Fatalf("requests to %s = %d, want 1", tt.path, requests[tt.path])
		}
	}
}
