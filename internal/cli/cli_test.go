package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/cli"
)

func TestHelpIncludesManPageStyleSections(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"help", "credentials"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(help) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"NAME", "SYNOPSIS", "DESCRIPTION", "SECURITY", "EXIT STATUS"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help output missing %q:\n%s", want, out)
		}
	}
}

func TestRootHelpAliasesShowManual(t *testing.T) {
	tests := [][]string{
		{"--help"},
		{"-h"},
		{"help"},
	}
	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(args, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Run(%v) code = %d stderr = %s", args, code, stderr.String())
			}
			out := stdout.String()
			for _, want := range []string{"NAME", "SYNOPSIS", "COMMANDS", "AGENT CONTRACT", "EXIT STATUS"} {
				if !strings.Contains(out, want) {
					t.Fatalf("root help missing %q:\n%s", want, out)
				}
			}
		})
	}
}

func TestRootHelpJSONIsAgentReadable(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"--json", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(--json --help) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "CommandManual"`, `"command": "pm"`, `"manual":`} {
		if !strings.Contains(out, want) {
			t.Fatalf("root json manual missing %q:\n%s", want, out)
		}
	}
}

func TestVersionCommandReportsBuildMetadata(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(version) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"pm dev", "commit: none", "built: unknown"} {
		if !strings.Contains(out, want) {
			t.Fatalf("version output missing %q:\n%s", want, out)
		}
	}
}

func TestVersionCommandJSONIsAgentReadable(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"version", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(version --json) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "Version"`, `"version": "dev"`, `"commit": "none"`, `"date": "unknown"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("version json missing %q:\n%s", want, out)
		}
	}
}

func TestBareCommandShowsManualInsteadOfUsageError(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{args: []string{"connectors"}, want: "pm connectors - inspect connector definitions, streams, and write actions"},
		{args: []string{"etl"}, want: "SYNC MODES"},
		{args: []string{"credentials"}, want: "pm credentials - manage encrypted connector credentials"},
		{args: []string{"connections"}, want: "pm connections - configure source-to-destination sync connections"},
		{args: []string{"reverse"}, want: "pm reverse - plan, preview, approve, and execute reverse ETL"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(tt.args, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Run(%v) code = %d stderr = %s", tt.args, code, stderr.String())
			}
			out := stdout.String()
			if strings.Contains(out, "invalid usage") || strings.Contains(stderr.String(), "invalid usage") {
				t.Fatalf("bare command returned usage error; stdout=%q stderr=%q", out, stderr.String())
			}
			for _, section := range []string{"NAME", "SYNOPSIS", "DESCRIPTION", "EXIT STATUS"} {
				if !strings.Contains(out, section) {
					t.Fatalf("manual missing section %q:\n%s", section, out)
				}
			}
			if !strings.Contains(out, tt.want) {
				t.Fatalf("manual missing %q:\n%s", tt.want, out)
			}
		})
	}
}

func TestBareCommandJSONShowsManualForAgents(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors --json) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "CommandManual"`, `"command": "connectors"`, `"manual":`} {
		if !strings.Contains(out, want) {
			t.Fatalf("json manual missing %q:\n%s", want, out)
		}
	}
}

func TestJiraConnectorCommandSurfaceHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "help flag", args: []string{"jira", "--help"}},
		{name: "help subcommand", args: []string{"jira", "help"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(tt.args, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Run(%v) code = %d stderr = %s", tt.args, code, stderr.String())
			}
			out := stdout.String()
			for _, want := range []string{
				"COMMAND SURFACE",
				"Usage: pm jira <command> [flags]",
				"Issue Commands",
				"issue list - List Jira issues [intent=etl availability=implemented stream=issues]",
				"issue delete - Delete an issue [intent=direct_write availability=unsafe_or_disallowed]",
				"permission scheme list - List permission schemes [intent=direct_read availability=planned]",
				"Help topics:",
			} {
				if !strings.Contains(out, want) {
					t.Fatalf("jira help missing %q:\nstdout=%s\nstderr=%s", want, out, stderr.String())
				}
			}
			if strings.Contains(out, "invalid usage") || strings.Contains(stderr.String(), "invalid usage") {
				t.Fatalf("jira help returned usage error; stdout=%q stderr=%q", out, stderr.String())
			}
		})
	}
}

func TestBareJiraConnectorCommandShowsHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"jira"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(jira) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"NAME", "SYNOPSIS", "COMMAND SURFACE", "project list", "user list"} {
		if !strings.Contains(out, want) {
			t.Fatalf("bare jira manual missing %q:\nstdout=%s\nstderr=%s", want, out, stderr.String())
		}
	}
}

func TestJiraConnectorCommandSurfaceHelpJSON(t *testing.T) {
	tests := [][]string{
		{"--json", "jira", "--help"},
		{"--json", "help", "jira"},
	}
	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(args, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Run(%v) code = %d stderr = %s", args, code, stderr.String())
			}
			out := stdout.String()
			for _, want := range []string{`"kind": "CommandManual"`, `"command": "jira"`, `"manual":`, "COMMAND SURFACE"} {
				if !strings.Contains(out, want) {
					t.Fatalf("jira json help missing %q:\n%s", want, out)
				}
			}
		})
	}
}

func TestConnectorsManualDocumentsConnectorArchitectureAndGithubExamples(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"declarative JSON bundles",
		"write=true/false",
		"REVERSE ETL WRITE ACTIONS",
		"pm connectors catalog --capability write --json",
		"GITHUB AUTHENTICATION",
		"public",
		"token",
		"github_app",
		"GITHUB ETL STREAMS",
		"issues",
		"pull_requests",
		"create_pull_request",
		"merge_pull_request",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("connectors manual missing %q:\n%s", want, out)
		}
	}
}

func TestConnectorInspectHumanShowsManualNotRawJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "inspect", "github"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors inspect github) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	if strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Fatalf("human connector inspect returned raw JSON:\n%s", out)
	}
	for _, want := range []string{"NAME", "SYNOPSIS", "AUTHENTICATION", "ETL STREAMS", "REVERSE ETL ACTIONS", "AGENT WORKFLOW"} {
		if !strings.Contains(out, want) {
			t.Fatalf("human connector manual missing %q:\n%s", want, out)
		}
	}
}

func TestDocsGenerateAndValidateConnectorDocs(t *testing.T) {
	dir := t.TempDir()
	cliDir := dir + "/cli"
	connectorsDir := dir + "/connectors"
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"docs", "generate", "--dir", cliDir, "--connectors-dir", connectorsDir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("docs generate code = %d stderr = %s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{"docs", "validate", "--connectors-dir", connectorsDir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("docs validate code = %d stderr = %s stdout = %s", code, stderr.String(), stdout.String())
	}
}

func TestConnectorListJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "list", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors list) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "ConnectorList"`, `"name": "sample"`, `"name": "warehouse"`, `"name": "akeneo"`, `"name": "github"`, `"name": "postgres"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("json output missing %q:\n%s", want, out)
		}
	}
	for _, forbidden := range []string{`"name": "source-github"`, `"name": "destination-postgres"`} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("json output contains legacy slug %q:\n%s", forbidden, out)
		}
	}
}

func TestPerfCompareJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"perf", "compare", "--iterations", "1", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(perf compare) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "PerformanceComparison"`, `"mode": "dependency-free"`, `"records": 3`} {
		if !strings.Contains(out, want) {
			t.Fatalf("json output missing %q:\n%s", want, out)
		}
	}
}

func TestPerfSyncModesJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"perf", "sync-modes", "--records", "20", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(perf sync-modes) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "SyncModeBenchmark"`, `"full_refresh_append"`, `"incremental_append_deduped"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("json output missing %q:\n%s", want, out)
		}
	}
}

func TestETLHelpListsAllSyncModes(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"help", "etl"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(help etl) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"full_refresh_append",
		"full_refresh_overwrite",
		"full_refresh_overwrite_deduped",
		"incremental_append",
		"incremental_append_deduped",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("etl help missing %q:\n%s", want, out)
		}
	}
}

func TestETLRejectsLegacyPrefixedConnectorCommands(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"init", "--root", root, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("init code = %d stderr = %s", code, stderr.String())
	}

	tests := [][]string{
		{"etl", "check", "--connector", "source-strava", "--root", root, "--json"},
		{"etl", "catalog", "--connector", "source-strava", "--root", root, "--json"},
		{"etl", "read", "--connector", "source-strava", "--stream", "records", "--limit", "1", "--root", root, "--json"},
	}
	for _, args := range tests {
		t.Run(strings.Join(args[:3], " "), func(t *testing.T) {
			stdout.Reset()
			stderr.Reset()
			code := cli.Run(args, &stdout, &stderr)
			if code == 0 {
				t.Fatalf("Run(%v) code = 0, want planned connector rejection; stdout = %s", args, stdout.String())
			}
			if !strings.Contains(stderr.String()+stdout.String(), `connector "source-strava" uses a legacy source-/destination- prefix; use bare connector name "strava"`) {
				t.Fatalf("Run(%v) did not explain legacy prefix migration; stdout=%s stderr=%s", args, stdout.String(), stderr.String())
			}
		})
	}
}

func TestRuntimeDoctorJSONDoesNotLeakPostgresPassword(t *testing.T) {
	t.Setenv("POLYMETRICS_POSTGRES_URL", "postgres://user:secret@127.0.0.1:1/db?sslmode=disable")
	t.Setenv("POLYMETRICS_DRAGONFLY_ADDR", "127.0.0.1:1")
	t.Setenv("POLYMETRICS_TEMPORAL_ADDR", "127.0.0.1:1")

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"runtime", "doctor", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(runtime doctor) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	if strings.Contains(out, "secret") {
		t.Fatalf("runtime doctor leaked password:\n%s", out)
	}
	if !strings.Contains(out, `"kind": "RuntimeDoctor"`) {
		t.Fatalf("missing RuntimeDoctor kind:\n%s", out)
	}
}

func TestGitHubCommandSurfaceRunsStreamBackedIssueList(t *testing.T) {
	var gotPath, gotState string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotState = r.URL.Query().Get("state")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"id": 101,
				"node_id": "I_kwDOAA",
				"number": 101,
				"state": "closed",
				"title": "closed issue",
				"user": {"login": "octocat", "id": 1},
				"updated_at": "2026-07-06T00:00:00Z"
			}
		]`))
	}))
	t.Cleanup(srv.Close)

	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	runCLI(t, []string{
		"credentials", "add", "github-local",
		"--connector", "github",
		"--config", "owner=octocat",
		"--config", "repo=hello-world",
		"--config", "base_url=" + srv.URL,
		"--config", "public_access=true",
		"--root", root,
		"--json",
	})

	stdout, _ := runCLI(t, []string{
		"github", "issue", "list",
		"--credential", "github-local",
		"--state", "closed",
		"--limit", "1",
		"--root", root,
		"--json",
	})
	if gotPath != "/repos/octocat/hello-world/issues" {
		t.Fatalf("request path = %q, want /repos/octocat/hello-world/issues", gotPath)
	}
	if gotState != "closed" {
		t.Fatalf("request state = %q, want closed", gotState)
	}

	var env struct {
		Kind    string `json:"kind"`
		Command string `json:"command"`
		Stream  string `json:"stream"`
		Count   int    `json:"count"`
		Records []struct {
			NodeID     string `json:"node_id"`
			State      string `json:"state"`
			Repository string `json:"repository"`
			UserLogin  string `json:"user_login"`
		} `json:"records"`
	}
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout)
	}
	if env.Kind != "ConnectorCommandRead" || env.Command != "issue list" || env.Stream != "issues" || env.Count != 1 {
		t.Fatalf("envelope = %+v, want kind ConnectorCommandRead command issue list stream issues count 1", env)
	}
	if len(env.Records) != 1 || env.Records[0].State != "closed" || env.Records[0].Repository != "octocat/hello-world" || env.Records[0].UserLogin != "octocat" {
		t.Fatalf("records = %+v, want projected GitHub issue record", env.Records)
	}
}

func TestJiraCommandSurfaceRunsStreamBackedCommands(t *testing.T) {
	type capturedRequest struct {
		authOK   bool
		rawQuery string
	}
	var requestsMu sync.Mutex
	requests := map[string]capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		requestsMu.Lock()
		requests[r.URL.Path] = capturedRequest{
			authOK:   ok && user == "agent@example.invalid" && pass == "test-token",
			rawQuery: r.URL.RawQuery,
		}
		requestsMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/rest/api/3/search":
			_, _ = w.Write([]byte(`{"issues":[{"id":"10001","key":"POLY-1","fields":{"summary":"Ship Jira runner","created":"2026-07-01T00:00:00Z","updated":"2026-07-02T00:00:00Z","status":{"name":"Done"},"issuetype":{"name":"Task"},"priority":{"name":"High"},"assignee":{"displayName":"Ada"},"reporter":{"displayName":"Grace"},"project":{"key":"POLY"}}}],"startAt":0,"maxResults":50,"total":1}`))
		case "/rest/api/3/project/search":
			_, _ = w.Write([]byte(`{"values":[{"id":"10000","key":"POLY","name":"Polymetrics","projectTypeKey":"software"}],"startAt":0,"maxResults":50,"total":1}`))
		case "/rest/api/3/users/search":
			_, _ = w.Write([]byte(`[{"accountId":"abc-123","accountType":"atlassian","displayName":"Ada Lovelace","emailAddress":"ada@example.invalid","active":true}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	t.Setenv("JIRA_TEST_TOKEN", "test-token")
	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	runCLI(t, []string{
		"credentials", "add", "jira-local",
		"--connector", "jira",
		"--config", "email=agent@example.invalid",
		"--config", "base_url=" + srv.URL,
		"--from-env", "api_token=JIRA_TEST_TOKEN",
		"--root", root,
		"--json",
	})

	tests := []struct {
		name       string
		args       []string
		wantPath   string
		wantKind   string
		wantStream string
		wantCount  int
		wantQuery  map[string]string
		wantRecord map[string]string
	}{
		{
			name:       "issue list",
			args:       []string{"jira", "issue", "list", "--credential", "jira-local", "--jql", "project = POLY", "--limit", "1", "--root", root, "--json"},
			wantPath:   "/rest/api/3/search",
			wantKind:   "ConnectorCommandRead",
			wantStream: "issues",
			wantCount:  1,
			wantQuery:  map[string]string{"jql": "project = POLY"},
			wantRecord: map[string]string{"key": "POLY-1", "summary": "Ship Jira runner", "status": "Done", "project": "POLY"},
		},
		{
			name:       "project list",
			args:       []string{"jira", "project", "list", "--credential", "jira-local", "--query", "Poly", "--limit", "1", "--root", root, "--json"},
			wantPath:   "/rest/api/3/project/search",
			wantKind:   "ConnectorCommandRead",
			wantStream: "projects",
			wantCount:  1,
			wantQuery:  map[string]string{"query": "Poly"},
			wantRecord: map[string]string{"key": "POLY", "name": "Polymetrics"},
		},
		{
			name:       "user list",
			args:       []string{"jira", "user", "list", "--credential", "jira-local", "--query", "ada", "--limit", "1", "--root", root, "--json"},
			wantPath:   "/rest/api/3/users/search",
			wantKind:   "ConnectorCommandRead",
			wantStream: "users",
			wantCount:  1,
			wantQuery:  map[string]string{"query": "ada"},
			wantRecord: map[string]string{"accountId": "abc-123", "displayName": "Ada Lovelace"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _ := runCLI(t, tt.args)
			requestsMu.Lock()
			request, ok := requests[tt.wantPath]
			requestsMu.Unlock()
			if !ok {
				t.Fatalf("%s was not requested; requests=%v", tt.wantPath, requests)
			}
			if !request.authOK {
				t.Fatalf("%s did not receive expected basic auth", tt.wantPath)
			}
			query, err := url.ParseQuery(request.rawQuery)
			if err != nil {
				t.Fatalf("parse raw query %q: %v", request.rawQuery, err)
			}
			for key, want := range tt.wantQuery {
				if got := query.Get(key); got != want {
					t.Fatalf("query[%s] = %q, want %q; raw=%s", key, got, want, request.rawQuery)
				}
			}

			var env struct {
				Kind    string           `json:"kind"`
				Stream  string           `json:"stream"`
				Count   int              `json:"count"`
				Records []map[string]any `json:"records"`
			}
			if err := json.Unmarshal([]byte(stdout), &env); err != nil {
				t.Fatalf("decode json: %v\n%s", err, stdout)
			}
			if env.Kind != tt.wantKind || env.Stream != tt.wantStream || env.Count != tt.wantCount {
				t.Fatalf("envelope = %+v, want kind=%s stream=%s count=%d", env, tt.wantKind, tt.wantStream, tt.wantCount)
			}
			if len(env.Records) != 1 {
				t.Fatalf("records length = %d, want 1", len(env.Records))
			}
			for key, want := range tt.wantRecord {
				if got := fmt.Sprint(env.Records[0][key]); got != want {
					t.Fatalf("record[%s] = %q, want %q; record=%+v", key, got, want, env.Records[0])
				}
			}
		})
	}
}

func TestJiraCommandSurfaceRunsGeneratedDirectRead(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		user, pass, ok := r.BasicAuth()
		if !ok || user != "agent@example.invalid" || pass != "test-token" {
			t.Errorf("basic auth = (%q, %q, %v), want Jira test credential", user, pass, ok)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"hello","content":"must redact","nested":{"apiToken":"must redact","name":"visible"}}`))
	}))
	t.Cleanup(srv.Close)

	t.Setenv("JIRA_TEST_TOKEN", "test-token")
	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	runCLI(t, []string{
		"credentials", "add", "jira-local",
		"--connector", "jira",
		"--config", "email=agent@example.invalid",
		"--config", "base_url=" + srv.URL,
		"--from-env", "api_token=JIRA_TEST_TOKEN",
		"--root", root,
		"--json",
	})

	stdout, _ := runCLI(t, []string{
		"jira", "rest", "get-banner",
		"--credential", "jira-local",
		"--root", root,
		"--json",
	})
	if gotPath != "/rest/api/3/announcementBanner" {
		t.Fatalf("request path = %q, want announcement banner endpoint", gotPath)
	}

	var env struct {
		Kind     string         `json:"kind"`
		Command  string         `json:"command"`
		Method   string         `json:"method"`
		Path     string         `json:"path"`
		Status   int            `json:"status"`
		Response map[string]any `json:"response"`
	}
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout)
	}
	if env.Kind != "ConnectorCommandDirectRead" || env.Command != "rest get-banner" || env.Method != http.MethodGet || env.Status != http.StatusOK {
		t.Fatalf("envelope = %+v, want generated Jira direct-read result", env)
	}
	if _, ok := env.Response["content"]; ok {
		t.Fatalf("response leaked content: %+v", env.Response)
	}
	nested, ok := env.Response["nested"].(map[string]any)
	if !ok {
		t.Fatalf("nested response = %T, want map", env.Response["nested"])
	}
	if _, ok := nested["apiToken"]; ok {
		t.Fatalf("response leaked apiToken: %+v", nested)
	}
	if env.Response["content_redacted"] != true || nested["apiToken_redacted"] != true || nested["name"] != "visible" {
		t.Fatalf("response redaction = %+v nested=%+v", env.Response, nested)
	}
}

func TestGitHubCommandSurfaceClampsOversizedLimit(t *testing.T) {
	const wantLimit = 10000
	var body strings.Builder
	body.WriteByte('[')
	for i := 1; i <= wantLimit+1; i++ {
		if i > 1 {
			body.WriteByte(',')
		}
		fmt.Fprintf(&body, `{
			"id": %d,
			"node_id": "I_%d",
			"number": %d,
			"state": "open",
			"title": "issue %d",
			"user": {"login": "octocat", "id": 1},
			"updated_at": "2026-07-06T00:00:00Z"
		}`, i, i, i, i)
	}
	body.WriteByte(']')

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body.String()))
	}))
	t.Cleanup(srv.Close)

	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	runCLI(t, []string{
		"credentials", "add", "github-local",
		"--connector", "github",
		"--config", "owner=octocat",
		"--config", "repo=hello-world",
		"--config", "base_url=" + srv.URL,
		"--config", "public_access=true",
		"--root", root,
		"--json",
	})

	stdout, _ := runCLI(t, []string{
		"github", "issue", "list",
		"--credential", "github-local",
		"--limit", fmt.Sprint(wantLimit + 1),
		"--root", root,
		"--json",
	})

	var env struct {
		Kind  string `json:"kind"`
		Count int    `json:"count"`
	}
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout)
	}
	if env.Kind != "ConnectorCommandRead" || env.Count != wantLimit {
		t.Fatalf("envelope = %+v, want clamped ConnectorCommandRead count %d", env, wantLimit)
	}
}

func TestGitHubCommandSurfaceRunsDirectReadFile(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"README.md","type":"file","encoding":"base64","content":"SGVsbG8=","download_url":"https://raw.example.test/README.md"}`))
	}))
	t.Cleanup(srv.Close)

	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	runCLI(t, []string{
		"credentials", "add", "github-local",
		"--connector", "github",
		"--config", "owner=octocat",
		"--config", "repo=hello-world",
		"--config", "base_url=" + srv.URL,
		"--config", "public_access=true",
		"--root", root,
		"--json",
	})

	stdout, _ := runCLI(t, []string{
		"github", "repo", "read-file",
		"--credential", "github-local",
		"--path", "README.md",
		"--root", root,
		"--json",
	})
	if gotPath != "/repos/octocat/hello-world/contents/README.md" {
		t.Fatalf("request path = %q, want contents file path", gotPath)
	}

	var env struct {
		Kind     string         `json:"kind"`
		Command  string         `json:"command"`
		Method   string         `json:"method"`
		Path     string         `json:"path"`
		Status   int            `json:"status"`
		Response map[string]any `json:"response"`
	}
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout)
	}
	if env.Kind != "ConnectorCommandDirectRead" || env.Command != "repo read-file" || env.Method != "GET" || env.Status != http.StatusOK {
		t.Fatalf("envelope = %+v, want direct-read README result", env)
	}
	if env.Response["name"] != "README.md" || env.Response["type"] != "file" {
		t.Fatalf("response = %+v, want README file metadata", env.Response)
	}
	if _, ok := env.Response["content"]; ok {
		t.Fatalf("response leaked content: %+v", env.Response)
	}
	if _, ok := env.Response["download_url"]; ok {
		t.Fatalf("response leaked download_url: %+v", env.Response)
	}
	if env.Response["content_redacted"] != true || env.Response["download_url_redacted"] != true {
		t.Fatalf("response redaction markers = %+v, want content and download_url redacted", env.Response)
	}
}

func TestGitHubCommandSurfacePlansReverseETLCommand(t *testing.T) {
	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	runCLI(t, []string{
		"credentials", "add", "github-local",
		"--connector", "github",
		"--config", "owner=octocat",
		"--config", "repo=hello-world",
		"--config", "public_access=true",
		"--root", root,
		"--json",
	})

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"github", "issue", "create",
		"--title", "Ship connector command plans",
		"--credential", "github-local",
		"--root", root,
		"--json",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("issue create code = %d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "ConnectorCommandWritePlan"`, `"connector_command": "issue create"`, `"action": "create_issue"`, `"approval_required": true`} {
		if !strings.Contains(out, want) {
			t.Fatalf("planned command output missing %q:\nstdout=%s\nstderr=%s", want, out, stderr.String())
		}
	}
	if strings.Contains(out, "approval_token") || strings.Contains(out, "approval_token_hash") ||
		strings.Contains(out, "connector_command_record") {
		t.Fatalf("plan JSON leaked approval or raw command payload:\n%s", out)
	}
}

func TestGitHubCommandSurfaceBlocksOperationBeforeCredentialResolution(t *testing.T) {
	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"github", "issue", "delete",
		"--issue-number", "40",
		"--root", root,
		"--json",
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("issue delete code = 0, want policy error; stdout=%s", stdout.String())
	}
	out := stdout.String()
	for _, want := range []string{`"category": "policy"`, `"code": "connector_command_blocked"`, "issue delete", "operation github.issue.delete"} {
		if !strings.Contains(out, want) {
			t.Fatalf("blocked operation output missing %q:\nstdout=%s\nstderr=%s", want, out, stderr.String())
		}
	}
	if strings.Contains(out, "missing --credential") || strings.Contains(stderr.String(), "missing --credential") {
		t.Fatalf("operation-backed command attempted credential resolution before blocking:\nstdout=%s\nstderr=%s", out, stderr.String())
	}
}

func runCLI(t *testing.T, args []string) (stdout string, stderr string) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	code := cli.Run(args, &outBuf, &errBuf)
	if code != 0 {
		t.Fatalf("Run(%v) code = %d stderr=%s stdout=%s", args, code, errBuf.String(), outBuf.String())
	}
	return outBuf.String(), errBuf.String()
}
