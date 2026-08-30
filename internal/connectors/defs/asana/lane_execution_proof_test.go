package asana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// TestAsanaTrackCImplementedCommandLanesReachEmbeddedRegistryAndCredentialBoundary
// proves that the claimed command lanes are loaded from the production embed,
// registered normally, and rejected at the credential boundary before any
// provider request can begin. It deliberately uses no credential and no live
// provider endpoint: Track C is execution proof, not live certification.
func TestAsanaTrackCImplementedCommandLanesReachEmbeddedRegistryAndCredentialBoundary(t *testing.T) {
	embedded, err := engine.Load(defs.FS, bundleName)
	if err != nil {
		t.Fatalf("load embedded Asana bundle: %v", err)
	}
	if embedded.Name != bundleName {
		t.Fatalf("embedded bundle name = %q, want %q", embedded.Name, bundleName)
	}

	registered, found := bundleregistry.New().Get(bundleName)
	if !found {
		t.Fatal("embedded Asana bundle is absent from the normal registry")
	}
	surfaceProvider, ok := registered.(connectors.CommandSurfaceProvider)
	if !ok || surfaceProvider.CommandSurface() == nil {
		t.Fatalf("registered Asana connector = %T, want command surface provider", registered)
	}
	assertAsanaTrackCCommandSurface(t, surfaceProvider.CommandSurface())

	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() = %v", err)
	}
	attachmentPath := filepath.Join(root, "track-c-attachment.txt")
	if err := os.WriteFile(attachmentPath, []byte("Track C bounded attachment fixture"), 0o600); err != nil {
		t.Fatalf("write attachment fixture: %v", err)
	}

	spy := &asanaTrackCNoProviderTransport{}
	previousTransport := http.DefaultTransport
	http.DefaultTransport = spy
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	for _, test := range []struct {
		name string
		args []string
	}{
		{
			name: "direct read",
			args: []string{"asana", "agents", "get-agents-for-workspace", "--workspace-gid", "workspace-track-c", "--root", root},
		},
		{
			name: "direct write",
			args: []string{"asana", "tasks", "create", "--name", "Track C proof", "--workspace", "workspace-track-c", "--root", root},
		},
		{
			name: "binary upload",
			args: []string{"asana", "attachments", "binary-upload-attachment", "--parent", "task-track-c", "--file-path", "track-c-attachment.txt", "--root", root},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(test.args, &stdout, &stderr)
			if code == 0 {
				t.Fatalf("Run(%v) succeeded; stdout=%s stderr=%s", test.args, stdout.String(), stderr.String())
			}
			if got := strings.TrimSpace(stdout.String() + stderr.String()); got != "error: missing --credential" {
				t.Fatalf("Run(%v) output = %q, want credential boundary", test.args, got)
			}
			if got := spy.requests.Load(); got != 0 {
				t.Fatalf("Run(%v) provider requests = %d, want zero before credential resolution", test.args, got)
			}
		})
	}
}

func assertAsanaTrackCCommandSurface(t *testing.T, surface *connectors.CommandSurface) {
	t.Helper()
	commands := map[string]connectors.CommandSurfaceCommand{}
	for _, command := range surface.Commands {
		commands[command.Path] = command
	}
	for _, want := range []struct {
		path            string
		intent          string
		sourceID        string
		stream          string
		write           string
		method          string
		providerPath    string
		rejectSourceCLI bool
	}{
		{path: "agents get-agents-for-workspace", intent: "direct_read", sourceID: "asana.rest.getAgentsForWorkspace", method: http.MethodGet, providerPath: "/workspaces/{workspace_gid}/agents"},
		{path: "tasks list", intent: "direct_read", sourceID: "asana.rest.getTasks", stream: "tasks", method: http.MethodGet, providerPath: "/tasks"},
		// Direct writes bind through the closed typed write/API-surface pair.
		// The current CLI surface intentionally does not invent a redundant
		// source_operation property for this path.
		{path: "tasks create", intent: "direct_write", write: "create_task", method: http.MethodPost, providerPath: "/tasks", rejectSourceCLI: true},
		{path: "attachments binary-upload-attachment", intent: "binary_upload", write: "upload_attachment_file", method: http.MethodPost, providerPath: "/attachments", rejectSourceCLI: true},
	} {
		command, ok := commands[want.path]
		if !ok {
			t.Fatalf("embedded command surface lacks %q", want.path)
		}
		if command.Intent != want.intent || command.Availability != "implemented" || command.SourceOperation != want.sourceID || command.Stream != want.stream || command.Write != want.write {
			t.Fatalf("embedded command %q = %+v, want implemented intent=%q source=%q stream=%q write=%q", want.path, command, want.intent, want.sourceID, want.stream, want.write)
		}
		if want.rejectSourceCLI && command.SourceOperation != "" {
			t.Fatalf("typed write command %q invents source_operation %q instead of using its write/API-surface binding", want.path, command.SourceOperation)
		}
		if len(command.APISurface) != 1 || !strings.EqualFold(command.APISurface[0].Method, want.method) || command.APISurface[0].Path != want.providerPath {
			t.Fatalf("embedded command %q API surface = %+v, want %s %s", want.path, command.APISurface, want.method, want.providerPath)
		}
	}
}

// TestAsanaTrackCETLThroughDuckDBUsesSourceBoundLocalFixture proves the one
// source-backed ETL witness through the full app path: embedded Asana stream,
// local provider fixture, local warehouse materialization, and owner-scoped
// read-back. The source matrix keeps all other candidate cells separate.
func TestAsanaTrackCETLThroughDuckDBUsesSourceBoundLocalFixture(t *testing.T) {
	assertAsanaTrackCMatrixWitnesses(t)

	fixture := &asanaTrackCTrustedFixtureTransport{}
	previousTransport := http.DefaultTransport
	http.DefaultTransport = fixture
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() = %v", err)
	}
	application, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	if _, err := application.AddCredential(ctx, app.AddCredentialRequest{
		Name: "asana-track-c", Connector: bundleName,
		Config:  map[string]string{"workspace_id": "workspace-track-c"},
		Secrets: map[string]string{"access_token": "asana-track-c-token"},
	}); err != nil {
		t.Fatalf("AddCredential(asana) = %v", err)
	}
	if _, err := application.AddCredential(ctx, app.AddCredentialRequest{
		Name: "warehouse-track-c", Connector: "warehouse",
		Config: map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatalf("AddCredential(warehouse) = %v", err)
	}
	connection, err := application.CreateConnection(ctx, app.CreateConnectionRequest{
		Name:        "asana_track_c_to_warehouse",
		Source:      app.EndpointConfig{Connector: bundleName, Credential: "asana-track-c"},
		Destination: app.EndpointConfig{Connector: "warehouse", Credential: "warehouse-track-c"},
		Streams: map[string]app.StreamConfig{
			"tasks": {
				SyncMode:         "full_refresh_overwrite",
				PrimaryKey:       []string{"gid"},
				DestinationTable: "asana_track_c_tasks",
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateConnection() = %v", err)
	}

	run, err := application.RunETL(ctx, app.RunETLRequest{Connection: connection.Name, Stream: "tasks", BatchSize: 1})
	if err != nil {
		t.Fatalf("RunETL() = %v", err)
	}
	if run.Status != "completed" || run.RecordsRead != 1 || run.RecordsLoaded != 1 || run.BatchCount != 1 {
		t.Fatalf("RunETL() = %+v, want one completed materialized record", run)
	}
	if run.Checkpoint["records_read"] != "1" || run.Checkpoint["batches"] != "1" {
		t.Fatalf("RunETL checkpoint = %+v, want acknowledged one-record/batch checkpoint", run.Checkpoint)
	}
	if got := fixture.requests.Load(); got != 1 {
		t.Fatalf("Asana fixture requests = %d, want exactly one bounded stream request", got)
	}

	rows, err := application.QueryTable(ctx, app.QueryTableRequest{Table: "asana_track_c_tasks", Connection: connection.Name, Limit: 10})
	if err != nil {
		t.Fatalf("QueryTable() = %v", err)
	}
	if len(rows) != 1 || fmt.Sprint(rows[0]["gid"]) != "task-track-c" {
		t.Fatalf("materialized DuckDB rows = %#v, want one source task", rows)
	}
}

// TestAsanaTrackCETLRejectsCredentialBaseURLOverride proves the local fixture
// route is test-only transport interception of the declared origin, rather
// than a configuration escape hatch that would weaken source-bound admission.
func TestAsanaTrackCETLRejectsCredentialBaseURLOverride(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() = %v", err)
	}
	application, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	if _, err := application.AddCredential(ctx, app.AddCredentialRequest{
		Name: "asana-track-c-rejected", Connector: bundleName,
		Config:  map[string]string{"base_url": "https://invalid.example", "workspace_id": "workspace-track-c"},
		Secrets: map[string]string{"access_token": "asana-track-c-token"},
	}); err != nil {
		t.Fatalf("AddCredential(asana) = %v", err)
	}
	if _, err := application.AddCredential(ctx, app.AddCredentialRequest{
		Name: "warehouse-track-c-rejected", Connector: "warehouse",
		Config: map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatalf("AddCredential(warehouse) = %v", err)
	}
	connection, err := application.CreateConnection(ctx, app.CreateConnectionRequest{
		Name:        "asana_track_c_rejected_origin",
		Source:      app.EndpointConfig{Connector: bundleName, Credential: "asana-track-c-rejected"},
		Destination: app.EndpointConfig{Connector: "warehouse", Credential: "warehouse-track-c-rejected"},
		Streams: map[string]app.StreamConfig{
			"tasks": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"gid"}, DestinationTable: "asana_track_c_rejected_tasks"},
		},
	})
	if err != nil {
		t.Fatalf("CreateConnection() = %v", err)
	}
	if _, err := application.RunETL(ctx, app.RunETLRequest{Connection: connection.Name, Stream: "tasks", BatchSize: 1}); err == nil || !strings.Contains(err.Error(), `source-bound provider operation "tasks" rejects configured base_url override`) {
		t.Fatalf("RunETL(base_url override) error = %v, want source-bound origin refusal", err)
	}
}

func assertAsanaTrackCMatrixWitnesses(t *testing.T) {
	t.Helper()
	matrix := loadAsanaSourceLaneMatrix(t)
	cells := map[string]map[string]string{}
	for _, row := range matrix.SourceOperations {
		cells[row.SourceID] = map[string]string{}
		for lane, cell := range row.Lanes {
			cells[row.SourceID][lane] = cell.Disposition
		}
	}
	for _, want := range []struct {
		sourceID    string
		lane        string
		disposition string
	}{
		{sourceID: "asana.rest.getTasks", lane: "direct_read", disposition: "implemented"},
		{sourceID: "asana.rest.getTasks", lane: "etl", disposition: "implemented"},
		{sourceID: "asana.rest.getAgentsForWorkspace", lane: "etl", disposition: "mapped_unproven"},
		{sourceID: "asana.rest.createTask", lane: "direct_write", disposition: "implemented"},
		{sourceID: "asana.rest.createTask", lane: "reverse_etl", disposition: "implemented"},
		{sourceID: "asana.rest.createAttachmentForObject", lane: "binary_upload", disposition: "implemented"},
		{sourceID: "asana.rest.createAttachmentForObject", lane: "binary_download", disposition: "not_applicable"},
		{sourceID: "asana.rest.getEvents", lane: "sync_transport", disposition: "implemented"},
	} {
		if got := cells[want.sourceID][want.lane]; got != want.disposition {
			t.Fatalf("matrix %s/%s = %q, want %q", want.sourceID, want.lane, got, want.disposition)
		}
	}
}

type asanaTrackCNoProviderTransport struct {
	requests atomic.Int64
}

func (spy *asanaTrackCNoProviderTransport) RoundTrip(*http.Request) (*http.Response, error) {
	spy.requests.Add(1)
	return nil, fmt.Errorf("unexpected provider I/O before credential resolution")
}

// asanaTrackCTrustedFixtureTransport preserves the bundle's real declared
// origin in the request and supplies only a local synthetic response. It never
// dials the network, so the proof cannot become a live-provider test.
type asanaTrackCTrustedFixtureTransport struct {
	requests atomic.Int64
}

func (fixture *asanaTrackCTrustedFixtureTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	fixture.requests.Add(1)
	if request.Method != http.MethodGet || request.URL.Scheme != "https" || request.URL.Host != "app.asana.com" || request.URL.Path != "/api/1.0/tasks" {
		return nil, fmt.Errorf("unexpected source-bound provider route %s %s", request.Method, request.URL.String())
	}
	if got := request.Header.Get("Authorization"); got != "Bearer asana-track-c-token" {
		return nil, fmt.Errorf("authorization = %q, want synthetic bearer credential", got)
	}
	query := request.URL.Query()
	if query.Get("limit") != "100" || query.Get("workspace") != "workspace-track-c" || query.Get("opt_fields") == "" {
		return nil, fmt.Errorf("tasks query = %q, want declared limit/workspace/opt_fields", request.URL.RawQuery)
	}
	body, err := json.Marshal(map[string]any{"data": []map[string]any{{
		"gid":           "task-track-c",
		"name":          "Track C source-bound task",
		"resource_type": "task",
		"created_at":    "2026-08-30T00:00:00Z",
		"modified_at":   "2026-08-30T00:00:00Z",
	}}})
	if err != nil {
		return nil, fmt.Errorf("encode Asana fixture response: %w", err)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(body)),
		Request:    request,
	}, nil
}
