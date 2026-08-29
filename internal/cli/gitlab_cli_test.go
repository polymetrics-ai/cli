package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/commandrunner"
)

func TestGitLabCommandSurfaceAdvertisesSourceLockedLanes(t *testing.T) {
	registry := bundleregistry.New()
	connector, ok := registry.Get("gitlab")
	if !ok {
		t.Fatal("GitLab connector is not registered")
	}

	manifest := connectors.ManifestOf(connector)
	if !manifest.Metadata.Capabilities.Read || !manifest.Metadata.Capabilities.Write || len(manifest.WriteActions) != 147 {
		t.Fatalf("GitLab executable capabilities = %+v with %d write actions, want source-locked read/write lanes", manifest.Metadata.Capabilities, len(manifest.WriteActions))
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
	if !strings.Contains(surface.Tagline, "582 source-bound direct reads") || !strings.Contains(surface.Tagline, "147 approval-gated reverse-ETL actions") {
		t.Fatalf("GitLab command tagline = %q, want current source-locked lanes", surface.Tagline)
	}
	writeActions := make(map[string]struct{}, len(manifest.WriteActions))
	for _, action := range manifest.WriteActions {
		writeActions[action.Name] = struct{}{}
	}
	counts := map[string]int{}
	for _, command := range surface.Commands {
		switch command.Intent {
		case "etl":
			counts["etl"]++
			expectation, known := want[command.Path]
			if !known {
				t.Fatalf("unexpected GitLab ETL command %q", command.Path)
			}
			if command.Availability != "implemented" || command.Stream != expectation.stream || command.SourceCLIPath != expectation.method+" "+expectation.path || command.SourceURL != expectation.url {
				t.Fatalf("command %q = %+v, want implemented %s stream with provider citation %s", command.Path, command, expectation.stream, expectation.url)
			}
			if len(command.APISurface) != 1 || command.APISurface[0].Method != expectation.method || command.APISurface[0].Path != expectation.path {
				t.Fatalf("command %q API surface = %+v, want %s %s", command.Path, command.APISurface, expectation.method, expectation.path)
			}
		case "direct_read":
			counts["direct_read"]++
			if command.Availability != "implemented" || command.Operation == "" || command.SourceOperation == "" || len(command.APISurface) != 1 {
				t.Fatalf("GitLab direct-read command %q = %+v, want source-bound implemented command", command.Path, command)
			}
		case "reverse_etl":
			counts["reverse_etl"]++
			if command.Availability != "implemented" || command.Write == "" || command.SourceOperation == "" || len(command.APISurface) != 1 || !strings.Contains(command.Approval, "plan, preview, approval, execute") {
				t.Fatalf("GitLab reverse-ETL command %q = %+v, want source-bound approval-gated action", command.Path, command)
			}
			if _, found := writeActions[command.Write]; !found {
				t.Fatalf("GitLab reverse-ETL command %q references unknown write %q", command.Path, command.Write)
			}
		default:
			t.Fatalf("unexpected GitLab command intent %q for %q", command.Intent, command.Path)
		}
	}
	if counts["direct_read"] != 582 || counts["etl"] != len(want) || counts["reverse_etl"] != len(manifest.WriteActions) {
		t.Fatalf("GitLab command lane counts = %+v, want 582 direct reads, %d ETL streams, and %d reverse-ETL actions", counts, len(want), len(manifest.WriteActions))
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

func TestGitLabGeneratedDirectReadReachesCredentialBoundary(t *testing.T) {
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("init project: %v", err)
	}

	spy := &gitLabNoNetworkTransportSpy{}
	oldTransport := http.DefaultTransport
	http.DefaultTransport = spy
	t.Cleanup(func() { http.DefaultTransport = oldTransport })

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"gitlab", "api", "op-474554202f6170692f76342f61646d696e2f6163746976655f636f6e746578742f636f6e6e656374696f6e73",
		"--root", root,
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("generated GitLab direct read unexpectedly succeeded; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if got := strings.TrimSpace(stdout.String() + stderr.String()); got != "error: missing --credential" {
		t.Fatalf("generated GitLab direct read = %q, want credential boundary", got)
	}
	if got := spy.requests.Load(); got != 0 {
		t.Fatalf("generated GitLab direct read provider requests = %d, want zero", got)
	}
}

func TestGitLabSourceLockedCommandsPassRuntimePreflight(t *testing.T) {
	registry := bundleregistry.New()
	connector, ok := registry.Get("gitlab")
	if !ok {
		t.Fatal("GitLab connector is not registered")
	}
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		t.Fatal("GitLab connector has no command surface")
	}

	counts := map[string]int{}
	for _, command := range provider.CommandSurface().Commands {
		if command.Availability != "implemented" {
			continue
		}
		if err := commandrunner.Preflight(connector, strings.Fields(command.Path)); err != nil {
			t.Fatalf("GitLab implemented %s command %q fails runtime preflight: %v", command.Intent, command.Path, err)
		}
		counts[command.Intent]++
	}
	if counts["direct_read"] != 582 || counts["etl"] != 4 || counts["reverse_etl"] != 147 {
		t.Fatalf("runtime-preflight lane counts = %+v, want direct_read=582 etl=4 reverse_etl=147", counts)
	}
}

type gitLabNoNetworkTransportSpy struct {
	requests atomic.Int64
}

func (spy *gitLabNoNetworkTransportSpy) RoundTrip(*http.Request) (*http.Response, error) {
	spy.requests.Add(1)
	return nil, fmt.Errorf("unexpected GitLab provider I/O")
}
