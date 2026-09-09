package app

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// TestAsanaEventTransportMaterializesBootstrapResumeAndTombstone is a local
// provider-double proof for the already-declared tasks event source. It keeps
// the source-locked Asana origin in every request and redirects only TLS dial
// to httptest, so it proves app transport routing and DuckDB materialization
// without treating a local base_url override as provider execution evidence.
func TestAsanaEventTransportMaterializesBootstrapResumeAndTombstone(t *testing.T) {
	fixture := newAsanaEventTransportHTTPFixture(t)
	fixture.assertBootstrapResumeAndTombstoneThroughDuckDB(t)
}

const (
	asanaEventTransportFixtureToken = "fixture-asana-event-access-token"
)

type asanaEventTransportHTTPCall struct {
	Path  string
	Query string
}

type asanaEventTransportHTTPFixture struct {
	t      *testing.T
	server *httptest.Server

	mu    sync.Mutex
	calls []asanaEventTransportHTTPCall
}

func newAsanaEventTransportHTTPFixture(t *testing.T) *asanaEventTransportHTTPFixture {
	t.Helper()
	fixture := &asanaEventTransportHTTPFixture{t: t}
	fixture.server = httptest.NewTLSServer(http.HandlerFunc(fixture.serveHTTP))
	t.Cleanup(fixture.server.Close)
	fixture.installSourceBoundOriginRedirect(t)
	return fixture
}

func (f *asanaEventTransportHTTPFixture) installSourceBoundOriginRedirect(t *testing.T) {
	t.Helper()
	serverTransport, ok := f.server.Client().Transport.(*http.Transport)
	if !ok {
		t.Fatal("Asana event httptest client does not expose an HTTP transport")
	}
	transport := serverTransport.Clone()
	transport.Proxy = nil
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	} else {
		transport.TLSClientConfig = transport.TLSClientConfig.Clone()
	}
	// The source-bound URL must remain app.asana.com. This test-only loopback
	// dial reaches the ephemeral httptest certificate, not that production host.
	transport.TLSClientConfig.InsecureSkipVerify = true //nolint:gosec // local httptest-only redirect
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address == "app.asana.com:443" {
			return (&net.Dialer{}).DialContext(ctx, network, f.server.Listener.Addr().String())
		}
		return nil, fmt.Errorf("Asana event fixture blocked unexpected outbound dial to %q", address)
	}
	previous := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = previous })
}

func (f *asanaEventTransportHTTPFixture) assertBootstrapResumeAndTombstoneThroughDuckDB(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatalf("initialize Asana event witness project: %v", err)
	}
	application, err := Open(root)
	if err != nil {
		t.Fatalf("open Asana event witness project: %v", err)
	}
	if _, err := application.AddCredential(ctx, AddCredentialRequest{
		Name:      "asana-event-local",
		Connector: "asana",
		// project_id is intentionally the only event-scope selector: the
		// source contract refuses workspace_id and assignee for this route.
		Config:  map[string]string{"project_id": "project-one"},
		Secrets: map[string]string{"access_token": asanaEventTransportFixtureToken},
	}); err != nil {
		t.Fatalf("add local Asana event credential: %v", err)
	}
	if _, err := application.AddCredential(ctx, AddCredentialRequest{
		Name:      "warehouse-local",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatalf("add local event warehouse credential: %v", err)
	}
	connection, err := application.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "asana_project_tasks_events_to_warehouse",
		Source:      EndpointConfig{Connector: "asana", Credential: "asana-event-local"},
		Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse-local"},
		Streams: map[string]StreamConfig{"tasks": {
			SyncMode:         "incremental_append",
			PrimaryKey:       []string{"gid"},
			DestinationTable: "asana_project_task_events_witness",
		}},
	})
	if err != nil {
		t.Fatalf("create Asana task event connection: %v", err)
	}

	bootstrap, err := application.RunETL(ctx, RunETLRequest{Connection: connection.Name, Stream: "tasks", BatchSize: 1})
	if err != nil {
		t.Fatalf("bootstrap Asana task event RunETL: %v", err)
	}
	f.assertCompletedRun(t, "bootstrap", bootstrap, 1)
	f.assertCheckpointToken(t, application, connection.Name, "bootstrap-token")

	// Reopen before the token window so this assertion verifies durable state,
	// not merely an in-memory cursor.
	application, err = Open(root)
	if err != nil {
		t.Fatalf("reopen after Asana event bootstrap: %v", err)
	}
	resumed, err := application.RunETL(ctx, RunETLRequest{Connection: connection.Name, Stream: "tasks", BatchSize: 1})
	if err != nil {
		t.Fatalf("resume Asana task event RunETL: %v", err)
	}
	// The changed task is the single loaded record; the deleted task is carried
	// as a separate tombstone and is asserted from DuckDB below.
	f.assertCompletedRun(t, "resumed token window", resumed, 1)
	f.assertCheckpointToken(t, application, connection.Name, "final-token")

	// A successfully exhausted empty window may advance only the provider
	// token. It must not materialize another record or synthesize a delete.
	application, err = Open(root)
	if err != nil {
		t.Fatalf("reopen before Asana event replay: %v", err)
	}
	replayed, err := application.RunETL(ctx, RunETLRequest{Connection: connection.Name, Stream: "tasks", BatchSize: 1})
	if err != nil {
		t.Fatalf("replay Asana task event RunETL: %v", err)
	}
	f.assertCompletedRun(t, "empty replay token window", replayed, 0)
	f.assertCheckpointToken(t, application, connection.Name, "event-replay-token")

	rows, err := application.QueryTable(ctx, QueryTableRequest{
		Connection: connection.Name,
		Table:      "asana_project_task_events_witness",
		Limit:      20,
	})
	if err != nil {
		t.Fatalf("query Asana event DuckDB table: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("Asana event DuckDB rows = %#v, want bootstrap task, changed task, and deleted tombstone", rows)
	}
	byGID := make(map[string]map[string]any, len(rows))
	for _, row := range rows {
		gid, _ := row["gid"].(string)
		if gid == "" {
			t.Fatalf("Asana event DuckDB row lacks gid: %#v", row)
		}
		byGID[gid] = map[string]any(row)
	}
	if got := byGID["task-bootstrap"]["name"]; got != "bootstrap snapshot task" {
		t.Fatalf("bootstrap DuckDB record = %#v, want source snapshot payload", byGID["task-bootstrap"])
	}
	if got := byGID["task-live"]["name"]; got != "hydrated current task" {
		t.Fatalf("resumed DuckDB record = %#v, want hydrated changed task", byGID["task-live"])
	}
	if deleted, _ := byGID["task-deleted"]["_polymetrics_deleted"].(bool); !deleted {
		t.Fatalf("deleted Asana event row = %#v, want durable tombstone marker", byGID["task-deleted"])
	}
	if _, found := byGID["task-removed"]; found {
		t.Fatalf("relationship-removed event became a resource tombstone: %#v", byGID["task-removed"])
	}

	f.assertEventTokenCalls(t, []string{"", "bootstrap-token", "middle-token", "final-token"})
	f.assertSourcePaths(t, []string{
		"/events", "/tasks", "/events", "/events", "/tasks/task-deleted", "/tasks/task-live", "/tasks/task-removed", "/events",
	})
}

func (f *asanaEventTransportHTTPFixture) assertCompletedRun(t *testing.T, phase string, run Run, wantLoaded int) {
	t.Helper()
	if run.Status != "completed" || run.RecordsLoaded != wantLoaded {
		t.Fatalf("%s Asana event run = %#v, want completed with %d materialized records", phase, run, wantLoaded)
	}
}

func (f *asanaEventTransportHTTPFixture) assertCheckpointToken(t *testing.T, application *App, connection, want string) {
	t.Helper()
	state := application.state.StreamStates[streamStateKey(connection, "tasks")]
	if state.Checkpoint == nil || state.Checkpoint.CommittedAt == nil || string(state.Checkpoint.Position.Primary) != want {
		t.Fatalf("Asana event stream state = %#v, want durable %q sync token", state, want)
	}
}

func (f *asanaEventTransportHTTPFixture) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if got := r.Header.Get("Authorization"); got != "Bearer "+asanaEventTransportFixtureToken {
		f.failf("Asana event authorization = %q, want fixture bearer token", got)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodGet {
		f.failf("Asana event request method = %s, want GET", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/api/1.0/") {
		f.failf("Asana event path = %q, want source-locked /api/1.0 prefix", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/1.0")
	f.mu.Lock()
	f.calls = append(f.calls, asanaEventTransportHTTPCall{Path: path, Query: r.URL.RawQuery})
	f.mu.Unlock()

	switch path {
	case "/events":
		if got := r.URL.Query().Get("resource"); got != "project-one" {
			f.failf("Asana events resource = %q, want exact project-one scope", got)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch r.URL.Query().Get("sync") {
		case "":
			f.writeJSON(w, http.StatusPreconditionFailed, map[string]any{"sync": "bootstrap-token"})
		case "bootstrap-token":
			f.writeJSON(w, http.StatusOK, map[string]any{
				"data": []any{
					map[string]any{"action": "changed", "resource": map[string]any{"gid": "task-live", "resource_type": "task"}},
					map[string]any{"action": "deleted", "resource": map[string]any{"gid": "task-deleted", "resource_type": "task"}},
					map[string]any{"action": "removed", "resource": map[string]any{"gid": "task-removed", "resource_type": "task"}},
				},
				"sync": "middle-token", "has_more": true,
			})
		case "middle-token":
			f.writeJSON(w, http.StatusOK, map[string]any{
				"data": []any{map[string]any{"action": "changed", "resource": map[string]any{"gid": "task-live", "resource_type": "task"}}},
				"sync": "final-token", "has_more": false,
			})
		case "final-token":
			f.writeJSON(w, http.StatusOK, map[string]any{"data": []any{}, "sync": "event-replay-token", "has_more": false})
		default:
			f.failf("Asana events sync token = %q, want fixture bootstrap/resume token", r.URL.Query().Get("sync"))
			w.WriteHeader(http.StatusBadRequest)
		}
	case "/tasks":
		if got := r.URL.Query().Get("project"); got != "project-one" {
			f.failf("Asana bootstrap task project = %q, want project-one", got)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		f.writeJSON(w, http.StatusOK, map[string]any{"data": []any{map[string]any{
			"gid": "task-bootstrap", "name": "bootstrap snapshot task", "resource_type": "task",
		}}})
	case "/tasks/task-live":
		f.writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
			"gid": "task-live", "name": "hydrated current task", "resource_type": "task",
		}})
	case "/tasks/task-deleted", "/tasks/task-removed":
		f.writeJSON(w, http.StatusNotFound, map[string]any{"errors": []any{map[string]any{"message": "fixture task no longer available"}}})
	default:
		f.failf("unexpected Asana event source path %q", path)
		w.WriteHeader(http.StatusNotFound)
	}
}

func (f *asanaEventTransportHTTPFixture) writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		f.failf("encode Asana event fixture response: %v", err)
	}
}

func (f *asanaEventTransportHTTPFixture) assertEventTokenCalls(t *testing.T, want []string) {
	t.Helper()
	calls := f.snapshotCalls()
	got := make([]string, 0, len(calls))
	for _, call := range calls {
		if call.Path != "/events" {
			continue
		}
		token := ""
		for _, pair := range strings.Split(call.Query, "&") {
			if strings.HasPrefix(pair, "sync=") {
				token = strings.TrimPrefix(pair, "sync=")
				break
			}
		}
		got = append(got, token)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Asana Events sync calls = %v, want bootstrap then durable resume/replay tokens %v", got, want)
	}
}

func (f *asanaEventTransportHTTPFixture) assertSourcePaths(t *testing.T, want []string) {
	t.Helper()
	calls := f.snapshotCalls()
	got := make([]string, 0, len(calls))
	for _, call := range calls {
		got = append(got, call.Path)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Asana event source paths = %v, want exact bootstrap, hydration, and replay sequence %v", got, want)
	}
}

func (f *asanaEventTransportHTTPFixture) snapshotCalls() []asanaEventTransportHTTPCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]asanaEventTransportHTTPCall(nil), f.calls...)
}

func (f *asanaEventTransportHTTPFixture) failf(format string, args ...any) {
	// httptest handlers must let the response close before they mark the parent
	// test failed; Fatalf from the handler goroutine is not safe.
	f.t.Errorf(format, args...)
}
