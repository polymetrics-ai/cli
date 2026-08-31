package cli_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors/engine"
)

func TestAsanaDeclaredETLStreamsMaterializeThroughDuckDB(t *testing.T) {
	fixture := newAsanaETLFixture(t)
	fixture.assertEveryDeclaredStreamMaterializes(t)
}

func TestAsanaDeclaredETLStreamNonSuccessDoesNotMaterialize(t *testing.T) {
	fixture := newAsanaETLFixture(t)
	fixture.assertDeclaredNonSuccessStopsBeforeWarehouseWrite(t)
}

const asanaETLFixtureToken = "fixture-asana-access-token"

// asanaETLProviderBaseURL is deliberately the source-locked production
// origin. The fixture redirects only its TLS dial to the local httptest server
// so the witness cannot make a source-bound base_url override executable.
const asanaETLProviderBaseURL = "https://app.asana.com/api/1.0"

var asanaDeclaredETLStreamNames = []string{
	"custom_fields",
	"project_statuses",
	"projects",
	"sections",
	"stories",
	"tags",
	"tasks",
	"team_memberships",
	"teams",
	"users",
	"workspace_memberships",
	"workspaces",
}

type asanaETLHTTPCall struct {
	Path  string
	Query string
}

type asanaETLFixture struct {
	t        *testing.T
	server   *httptest.Server
	failPath string

	mu    sync.Mutex
	calls []asanaETLHTTPCall
}

func newAsanaETLFixture(t *testing.T) *asanaETLFixture {
	t.Helper()
	fixture := &asanaETLFixture{t: t}
	fixture.server = httptest.NewTLSServer(http.HandlerFunc(fixture.serveHTTP))
	t.Cleanup(fixture.server.Close)
	fixture.installSourceBoundOriginRedirect(t)
	return fixture
}

func (f *asanaETLFixture) installSourceBoundOriginRedirect(t *testing.T) {
	t.Helper()
	serverTransport, ok := f.server.Client().Transport.(*http.Transport)
	if !ok {
		t.Fatal("Asana httptest client does not expose an HTTP transport")
	}
	transport := serverTransport.Clone()
	transport.Proxy = nil
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	} else {
		transport.TLSClientConfig = transport.TLSClientConfig.Clone()
	}
	// The request URL intentionally keeps app.asana.com for the source-bound
	// origin guard, while this test-only loopback dial uses httptest's ephemeral
	// certificate rather than a certificate for that production hostname.
	transport.TLSClientConfig.InsecureSkipVerify = true //nolint:gosec // local httptest-only redirect
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address == "app.asana.com:443" {
			return (&net.Dialer{}).DialContext(ctx, network, f.server.Listener.Addr().String())
		}
		return nil, fmt.Errorf("Asana ETL fixture blocked unexpected outbound dial to %q", address)
	}
	previous := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = previous })
}

func (f *asanaETLFixture) assertEveryDeclaredStreamMaterializes(t *testing.T) {
	t.Helper()
	if got := asanaBundleETLStreamNames(t); !reflect.DeepEqual(got, asanaDeclaredETLStreamNames) {
		t.Fatalf("Asana declared ETL streams = %v, want exact existing 12-stream witness cohort %v", got, asanaDeclaredETLStreamNames)
	}

	application := f.openApplication(t)
	ctx := context.Background()
	for _, stream := range asanaDeclaredETLStreamNames {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			connectionName := "asana_" + stream + "_to_warehouse"
			tableName := "asana_" + stream + "_witness"
			connection, err := application.CreateConnection(ctx, app.CreateConnectionRequest{
				Name:        connectionName,
				Source:      app.EndpointConfig{Connector: "asana", Credential: "asana-local"},
				Destination: app.EndpointConfig{Connector: "warehouse", Credential: "warehouse-local"},
				Streams: map[string]app.StreamConfig{stream: {
					// These source-lock-backed collection reads remain on the
					// established full-refresh compatibility route. They are not
					// evidence that the closed full_append transport is admitted.
					SyncMode:         "full_refresh_append",
					DestinationTable: tableName,
				}},
			})
			if err != nil {
				t.Fatalf("create %s full-refresh connection: %v", stream, err)
			}

			f.resetCalls()
			run, err := application.RunETL(ctx, app.RunETLRequest{Connection: connection.Name, Stream: stream, BatchSize: 1})
			if err != nil {
				t.Fatalf("RunETL(%s) error = %v", stream, err)
			}
			wantRows := asanaETLExpectedRows(stream)
			if run.Status != "completed" || run.RecordsRead != wantRows || run.RecordsLoaded != wantRows {
				t.Fatalf("RunETL(%s) = %+v, want completed %d-record warehouse flow", stream, run, wantRows)
			}

			rows, err := application.QueryTable(ctx, app.QueryTableRequest{Connection: connection.Name, Table: tableName, Limit: 20})
			if err != nil {
				t.Fatalf("query %s DuckDB table: %v", stream, err)
			}
			if len(rows) != wantRows {
				t.Fatalf("%s DuckDB rows = %#v, want %d source records", stream, rows, wantRows)
			}
			for _, row := range rows {
				if gid, _ := row["gid"].(string); gid == "" {
					t.Fatalf("%s DuckDB row lacks projected source gid: %#v", stream, row)
				}
			}
			f.assertPaginatedSourceRoute(t, stream, wantRows)
		})
	}
}

func (f *asanaETLFixture) assertDeclaredNonSuccessStopsBeforeWarehouseWrite(t *testing.T) {
	t.Helper()
	application := f.openApplication(t)
	ctx := context.Background()
	connection, err := application.CreateConnection(ctx, app.CreateConnectionRequest{
		Name:        "asana_tasks_non_success_to_warehouse",
		Source:      app.EndpointConfig{Connector: "asana", Credential: "asana-local"},
		Destination: app.EndpointConfig{Connector: "warehouse", Credential: "warehouse-local"},
		Streams: map[string]app.StreamConfig{"tasks": {
			SyncMode:         "full_refresh_append",
			DestinationTable: "asana_tasks_non_success_witness",
		}},
	})
	if err != nil {
		t.Fatalf("create failed Asana tasks connection: %v", err)
	}

	f.failPath = "/tasks"
	f.resetCalls()
	run, err := application.RunETL(ctx, app.RunETLRequest{Connection: connection.Name, Stream: "tasks", BatchSize: 1})
	if err == nil || !strings.Contains(err.Error(), "502") {
		t.Fatalf("RunETL(tasks) non-success error = %v, want declared provider 502", err)
	}
	if run.Status != "failed" || run.RecordsRead != 0 || run.RecordsLoaded != 0 {
		t.Fatalf("failed Asana tasks run = %+v, want terminal failure before any warehouse record", run)
	}
	calls := f.snapshotCalls()
	if len(calls) == 0 {
		t.Fatal("failed Asana source made no provider request")
	}
	for _, call := range calls {
		if call.Path != "/tasks" || strings.Contains(call.Query, "offset=") {
			t.Fatalf("failed Asana source calls = %+v, want only first-page tasks retries", calls)
		}
	}
	rows, queryErr := application.QueryTable(ctx, app.QueryTableRequest{
		Connection: connection.Name,
		Table:      "asana_tasks_non_success_witness",
		Limit:      20,
	})
	if queryErr == nil || len(rows) != 0 {
		t.Fatalf("failed Asana tasks run materialized rows=%#v err=%v, want no warehouse result", rows, queryErr)
	}
}

func (f *asanaETLFixture) openApplication(t *testing.T) *app.App {
	t.Helper()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("initialize Asana ETL witness project: %v", err)
	}
	application, err := app.Open(root)
	if err != nil {
		t.Fatalf("open Asana ETL witness project: %v", err)
	}
	ctx := context.Background()
	if _, err := application.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "asana-local",
		Connector: "asana",
		Config: map[string]string{
			"workspace_id": "workspace-one",
			"project_id":   "project-one",
			"team_id":      "team-one",
		},
		Secrets: map[string]string{"access_token": asanaETLFixtureToken},
	}); err != nil {
		t.Fatalf("add local Asana credential: %v", err)
	}
	if _, err := application.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "warehouse-local",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatalf("add local warehouse credential: %v", err)
	}
	return application
}

func asanaBundleETLStreamNames(t *testing.T) []string {
	t.Helper()
	bundle, err := engine.Load(os.DirFS("../connectors/defs"), "asana")
	if err != nil {
		t.Fatalf("load Asana declared stream bundle: %v", err)
	}
	names := make([]string, 0, len(bundle.Streams))
	for _, stream := range bundle.Streams {
		names = append(names, stream.Name)
	}
	sort.Strings(names)
	return names
}

func asanaETLExpectedRows(stream string) int {
	switch stream {
	case "sections", "stories", "project_statuses":
		return 4
	default:
		return 2
	}
}

func (f *asanaETLFixture) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if got := r.Header.Get("Authorization"); got != "Bearer "+asanaETLFixtureToken {
		f.failf("Asana source authorization = %q, want fixture bearer token", got)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodGet {
		f.failf("Asana source method = %s, want GET", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if !strings.HasPrefix(r.URL.Path, "/api/1.0/") {
		f.failf("Asana source path = %q, want source-locked /api/1.0 prefix", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/1.0")

	f.mu.Lock()
	f.calls = append(f.calls, asanaETLHTTPCall{Path: path, Query: r.URL.RawQuery})
	f.mu.Unlock()

	if path == f.failPath && r.URL.Query().Get("offset") == "" {
		f.writeJSON(w, http.StatusBadGateway, map[string]any{"errors": []any{map[string]any{"message": "fixture source failure"}}})
		return
	}
	page := r.URL.Query().Get("offset")
	if page != "" && page != "fixture-page-two" {
		f.failf("Asana source offset = %q, want empty or fixture-page-two", page)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !asanaETLPathKnown(path) {
		f.failf("unexpected Asana ETL source path %q", path)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	record := asanaETLRecord(path, page == "fixture-page-two")
	payload := map[string]any{"data": []any{record}}
	if page == "" {
		payload["next_page"] = map[string]any{"uri": asanaETLProviderBaseURL + path + "?offset=fixture-page-two&limit=100"}
	}
	f.writeJSON(w, http.StatusOK, payload)
}

func asanaETLPathKnown(path string) bool {
	switch path {
	case "/workspaces", "/projects", "/tasks", "/users", "/tags", "/team_memberships":
		return true
	case "/workspaces/workspace-one/teams", "/workspaces/workspace-one/custom_fields", "/workspaces/workspace-one/workspace_memberships":
		return true
	case "/projects/project-one/sections", "/projects/project-two/sections", "/projects/project-one/project_statuses", "/projects/project-two/project_statuses":
		return true
	case "/tasks/task-one/stories", "/tasks/task-two/stories":
		return true
	default:
		return false
	}
}

func asanaETLRecord(path string, secondPage bool) map[string]any {
	pageSuffix := "one"
	if secondPage {
		pageSuffix = "two"
	}
	resourceType, idPrefix := "resource", "resource"
	switch {
	case path == "/workspaces":
		resourceType, idPrefix = "workspace", "workspace"
	case path == "/projects":
		resourceType, idPrefix = "project", "project"
	case path == "/tasks":
		resourceType, idPrefix = "task", "task"
	case path == "/users":
		resourceType, idPrefix = "user", "user"
	case path == "/workspaces/workspace-one/teams":
		resourceType, idPrefix = "team", "team"
	case path == "/tags":
		resourceType, idPrefix = "tag", "tag"
	case strings.Contains(path, "/sections"):
		resourceType, idPrefix = "section", "section-"+strings.TrimPrefix(strings.TrimSuffix(path, "/sections"), "/projects/")
	case strings.Contains(path, "/stories"):
		resourceType, idPrefix = "story", "story-"+strings.TrimPrefix(strings.TrimSuffix(path, "/stories"), "/tasks/")
	case path == "/workspaces/workspace-one/custom_fields":
		resourceType, idPrefix = "custom_field", "custom-field"
	case strings.Contains(path, "/project_statuses"):
		resourceType, idPrefix = "project_status", "status-"+strings.TrimPrefix(strings.TrimSuffix(path, "/project_statuses"), "/projects/")
	case path == "/team_memberships":
		resourceType, idPrefix = "team_membership", "team-membership"
	case path == "/workspaces/workspace-one/workspace_memberships":
		resourceType, idPrefix = "workspace_membership", "workspace-membership"
	}
	return map[string]any{
		"gid":           idPrefix + "-" + pageSuffix,
		"name":          idPrefix + " " + pageSuffix,
		"resource_type": resourceType,
	}
}

func (f *asanaETLFixture) writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		f.failf("encode Asana ETL fixture response: %v", err)
	}
}

func (f *asanaETLFixture) assertPaginatedSourceRoute(t *testing.T, stream string, wantRows int) {
	t.Helper()
	calls := f.snapshotCalls()
	wantCalls := 2
	if wantRows == 4 {
		wantCalls = 6
	}
	if len(calls) != wantCalls {
		t.Fatalf("%s source requests = %+v, want %d declared first/next-page requests", stream, calls, wantCalls)
	}
	secondPage := false
	for _, call := range calls {
		if strings.Contains(call.Query, "offset=fixture-page-two") {
			secondPage = true
		}
	}
	if !secondPage {
		t.Fatalf("%s source requests = %+v, want at least one next_page continuation", stream, calls)
	}
}

func (f *asanaETLFixture) resetCalls() {
	f.mu.Lock()
	f.calls = nil
	f.mu.Unlock()
}

func (f *asanaETLFixture) snapshotCalls() []asanaETLHTTPCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]asanaETLHTTPCall(nil), f.calls...)
}

func (f *asanaETLFixture) failf(format string, args ...any) {
	f.t.Errorf(format, args...)
}
