package commandrunner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

type fakeConnector struct {
	surface                    *connectors.CommandSurface
	manifest                   connectors.Manifest
	readReq                    connectors.ReadRequest
	directReadReq              connectors.DirectReadRequest
	operationDirectReadReq     connectors.OperationDirectReadRequest
	operationReadPreflight     operationDirectReadPreflightCall
	operationReadPreflightErr  error
	operationJSONVariable      operationStructuredJSONVariablePreflightCall
	operationJSONVariableErr   error
	operationDirectWriteReq    connectors.OperationDirectWriteRequest
	operationWritePreflight    operationDirectWritePreflightCall
	operationWritePreflightErr error
	directWriteMetadata        connectors.OperationDirectWriteMetadata
	binaryDownloadReq          connectors.OperationBinaryDownloadRequest
	directReadErr              error
	ignoresPageNavigation      bool
	operationDirectReadErr     error
	binaryDownloadErr          error
	validateReq                connectors.WriteRequest
	dryRunReq                  connectors.WriteRequest
	writeReq                   connectors.WriteRequest
	writeRecords               []connectors.Record
	validateErr                error
	dryRunErr                  error
	readErr                    error
	writeErr                   error
	readRecords                []connectors.Record
	preview                    connectors.WritePreview
	writeResult                connectors.WriteResult
}

type operationDirectReadPreflightCall struct {
	operation    string
	method       string
	path         string
	maxBytes     int
	outputPolicy string
}

type operationStructuredJSONVariablePreflightCall struct {
	operation string
	variable  string
}

type operationDirectWritePreflightCall struct {
	operation    string
	method       string
	path         string
	outputPolicy string
}

type preflightFakeConnector struct {
	*fakeConnector
	preflightErr error
}

func (f *preflightFakeConnector) PreflightWriteAction(string) error {
	return f.preflightErr
}

func (f *fakeConnector) Name() string { return "github" }
func (f *fakeConnector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "github", DisplayName: "GitHub"}
}
func (f *fakeConnector) Check(context.Context, connectors.RuntimeConfig) error { return nil }
func (f *fakeConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{}, nil
}
func (f *fakeConnector) Read(_ context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	f.readReq = req
	if f.readErr != nil {
		return f.readErr
	}
	if len(f.readRecords) > 0 {
		for _, record := range f.readRecords {
			if err := emit(record); err != nil {
				return err
			}
		}
		return nil
	}
	return emit(connectors.Record{"number": 101, "state": req.Query["state"]})
}
func (f *fakeConnector) DirectRead(_ context.Context, req connectors.DirectReadRequest) (connectors.DirectReadResult, error) {
	f.directReadReq = req
	if f.directReadErr != nil {
		return connectors.DirectReadResult{}, f.directReadErr
	}
	return connectors.DirectReadResult{
		Connector: "github",
		Method:    req.Method,
		Path:      "/repos/octo/hello/contents/README.md",
		Status:    200,
		Body: map[string]any{
			"name": "README.md",
			"type": "file",
		},
		Page: f.directReadPage(),
	}, nil
}

// directReadPage stands in for the page context a real direct-read executor
// reports. ignoresPageNavigation models the executor that does not: it accepts
// Page/PageCursor and hands back a zero page, which is what the runner refuses.
func (f *fakeConnector) directReadPage() connectors.DirectReadPage {
	if f.ignoresPageNavigation {
		return connectors.DirectReadPage{}
	}
	return connectors.DirectReadPage{Strategy: "page_number", Complete: true}
}

func (f *fakeConnector) OperationDirectRead(_ context.Context, req connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	f.operationDirectReadReq = req
	if f.operationDirectReadErr != nil {
		return connectors.DirectReadResult{}, f.operationDirectReadErr
	}
	return connectors.DirectReadResult{
		Connector: "gong",
		Method:    "POST",
		Path:      "/v2/meetings/integration/status",
		Status:    200,
		Body:      map[string]any{"ok": true},
		Page:      f.directReadPage(),
	}, nil
}
func (f *fakeConnector) PreflightOperationDirectRead(operation, method, path string, maxBytes int, outputPolicy string) error {
	f.operationReadPreflight = operationDirectReadPreflightCall{
		operation:    operation,
		method:       method,
		path:         path,
		maxBytes:     maxBytes,
		outputPolicy: outputPolicy,
	}
	return f.operationReadPreflightErr
}
func (f *fakeConnector) PreflightOperationStructuredJSONVariable(operation, variable string) error {
	f.operationJSONVariable = operationStructuredJSONVariablePreflightCall{operation: operation, variable: variable}
	return f.operationJSONVariableErr
}
func (f *fakeConnector) PreflightOperationDirectWrite(operation, method, path, outputPolicy string) error {
	f.operationWritePreflight = operationDirectWritePreflightCall{
		operation:    operation,
		method:       method,
		path:         path,
		outputPolicy: outputPolicy,
	}
	return f.operationWritePreflightErr
}
func (f *fakeConnector) PreviewOperationDirectWrite(_ context.Context, req connectors.OperationDirectWriteRequest) (connectors.WritePreview, error) {
	f.operationDirectWriteReq = req
	return connectors.WritePreview{Action: req.Operation, RecordsStaged: 1}, nil
}
func (f *fakeConnector) OperationDirectWrite(_ context.Context, req connectors.OperationDirectWriteRequest) (connectors.OperationDirectWriteResult, error) {
	f.operationDirectWriteReq = req
	return connectors.OperationDirectWriteResult{Connector: "github", Operation: req.Operation, Method: http.MethodPost, Path: "/api/vote", Status: http.StatusOK}, nil
}
func (f *fakeConnector) OperationDirectWriteMetadata(operation string) (connectors.OperationDirectWriteMetadata, error) {
	metadata := f.directWriteMetadata
	if metadata.Operation == "" {
		metadata.Operation = operation
	}
	if metadata.OutputPolicy == "" {
		metadata.OutputPolicy = "json_redacted"
	}
	return metadata, nil
}
func (f *fakeConnector) OperationBinaryDownload(_ context.Context, req connectors.OperationBinaryDownloadRequest) (connectors.OperationBinaryDownloadResult, error) {
	f.binaryDownloadReq = req
	if f.binaryDownloadErr != nil {
		return connectors.OperationBinaryDownloadResult{}, f.binaryDownloadErr
	}
	return connectors.OperationBinaryDownloadResult{
		Connector: "github",
		Operation: req.Operation,
		Record:    connectors.Record{"file_path": "out/artifact", "file_size_bytes": 12},
	}, nil
}
func (f *fakeConnector) Write(_ context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	f.writeReq = req
	f.writeRecords = append([]connectors.Record(nil), records...)
	if f.writeErr != nil {
		return connectors.WriteResult{}, f.writeErr
	}
	if f.writeResult != (connectors.WriteResult{}) {
		return f.writeResult, nil
	}
	return connectors.WriteResult{RecordsWritten: len(records)}, nil
}
func (f *fakeConnector) CommandSurface() *connectors.CommandSurface { return f.surface }
func (f *fakeConnector) Manifest() connectors.Manifest              { return f.manifest }
func (f *fakeConnector) ValidateWrite(_ context.Context, req connectors.WriteRequest, _ []connectors.Record) error {
	f.validateReq = req
	return f.validateErr
}
func (f *fakeConnector) DryRunWrite(_ context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	f.dryRunReq = req
	if f.dryRunErr != nil {
		return connectors.WritePreview{}, f.dryRunErr
	}
	if f.preview.Action != "" || f.preview.RecordsStaged != 0 || len(f.preview.Warnings) > 0 {
		return f.preview, nil
	}
	return connectors.WritePreview{Action: req.Action, RecordsStaged: len(records)}, nil
}

func TestRunImplementedStreamCommand(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "issue list",
				Intent:       "etl",
				Availability: "implemented",
				Stream:       "issues",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "state", Type: "enum", Values: []string{"open", "closed", "all"}, MapsTo: "query.state"},
				},
			},
		},
	}}

	var records []connectors.Record
	result, err := Run(context.Background(), connector, Request{
		Path:   []string{"issue", "list"},
		Flags:  map[string][]string{"state": []string{"closed"}},
		Config: connectors.RuntimeConfig{Config: map[string]string{"owner": "octocat", "repo": "hello-world"}},
		Limit:  1,
	}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Command != "issue list" || result.Stream != "issues" || result.Count != 1 {
		t.Fatalf("result = %+v, want command issue list stream issues count 1", result)
	}
	if connector.readReq.Stream != "issues" {
		t.Fatalf("read stream = %q, want issues", connector.readReq.Stream)
	}
	if got := connector.readReq.Query["state"]; got != "closed" {
		t.Fatalf("read query state = %q, want closed", got)
	}
	if len(records) != 1 || records[0]["state"] != "closed" {
		t.Fatalf("records = %+v, want one closed record", records)
	}
}

func TestRunImplementedStreamCommandConfigFlagOverridesConfig(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "agents view",
				Intent:       "etl",
				Availability: "implemented",
				Stream:       "agent_details",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "agent-id", Type: "string", MapsTo: "config.agent_id"},
					{Name: "include", Type: "string", MapsTo: "query.include"},
				},
			},
		},
	}}
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"agent_id": "connection-agent", "base_url": "https://api.example.test"},
		Secrets: map[string]string{"api_key": "credential-secret"},
	}

	_, err := Run(context.Background(), connector, Request{
		Path: []string{"agents", "view"},
		Flags: map[string][]string{
			"agent-id": {"flag-agent"},
			"include":  {"availability"},
		},
		Config: cfg,
		Limit:  1,
	}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := connector.readReq.Config.Config["agent_id"]; got != "flag-agent" {
		t.Fatalf("read config agent_id = %q, want flag-agent", got)
	}
	if got := connector.readReq.Config.Config["base_url"]; got != "https://api.example.test" {
		t.Fatalf("read config base_url = %q, want preserved base_url", got)
	}
	if got := connector.readReq.Query["include"]; got != "availability" {
		t.Fatalf("read query include = %q, want availability", got)
	}
	if got := cfg.Config["agent_id"]; got != "connection-agent" {
		t.Fatalf("input config was mutated: agent_id = %q", got)
	}
}

func TestRunCoreStreamMappings(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{Path: "pr list", Intent: "etl", Availability: "implemented", Stream: "pull_requests"},
			{Path: "release list", Intent: "etl", Availability: "implemented", Stream: "releases"},
			{Path: "workflow list", Intent: "etl", Availability: "implemented", Stream: "workflows"},
		},
	}}

	tests := []struct {
		name   string
		path   []string
		stream string
	}{
		{name: "pull requests", path: []string{"pr", "list"}, stream: "pull_requests"},
		{name: "releases", path: []string{"release", "list"}, stream: "releases"},
		{name: "workflows", path: []string{"workflow", "list"}, stream: "workflows"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var records []connectors.Record
			result, err := Run(context.Background(), connector, Request{Path: tt.path, Limit: 1}, func(record connectors.Record) error {
				records = append(records, record)
				return nil
			})
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if result.Stream != tt.stream {
				t.Fatalf("stream = %q, want %q", result.Stream, tt.stream)
			}
			if connector.readReq.Stream != tt.stream {
				t.Fatalf("read stream = %q, want %q", connector.readReq.Stream, tt.stream)
			}
			if len(records) != 1 {
				t.Fatalf("records = %d, want 1", len(records))
			}
		})
	}
}

func TestCLICommandValidationDefinitionsGeneric(t *testing.T) {
	allowEmpty := false
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "widget list",
				Intent:       "etl",
				Availability: "implemented",
				Stream:       "widgets",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "start", Type: "string", MapsTo: "query.started_after", Format: "date-time", AllowEmpty: &allowEmpty},
					{Name: "end", Type: "string", MapsTo: "query.started_before", Format: "date-time", AllowEmpty: &allowEmpty},
				},
				Constraints: []connectors.CommandSurfaceConstraint{
					{
						Kind:         "order",
						Left:         "query.started_after",
						LeftFallback: "config.default_start",
						Op:           "lt",
						Right:        "query.started_before",
						ValueType:    "date-time",
						Message:      "window start must be before end",
					},
				},
			},
		},
	}}

	tests := []struct {
		name    string
		flags   map[string][]string
		config  map[string]string
		wantErr string
	}{
		{
			name: "valid explicit bounds",
			flags: map[string][]string{
				"start": {"2026-07-01T00:00:00Z"},
				"end":   {"2026-07-02T00:00:00Z"},
			},
		},
		{
			name:  "missing right side skips order constraint",
			flags: map[string][]string{"start": {"2026-07-01T00:00:00Z"}},
		},
		{
			name:   "valid config fallback",
			flags:  map[string][]string{"end": {"2026-07-02T00:00:00Z"}},
			config: map[string]string{"default_start": "2026-07-01T00:00:00Z"},
		},
		{
			name:    "invalid timestamp",
			flags:   map[string][]string{"start": {"2026-07-01"}},
			wantErr: "invalid --start",
		},
		{
			name:    "blank timestamp",
			flags:   map[string][]string{"start": {""}},
			wantErr: "invalid --start",
		},
		{
			name:    "invalid config fallback",
			flags:   map[string][]string{"end": {"2026-07-02T00:00:00Z"}},
			config:  map[string]string{"default_start": "not-a-time"},
			wantErr: "invalid config.default_start",
		},
		{
			name: "invalid explicit order",
			flags: map[string][]string{
				"start": {"2026-07-02T00:00:00Z"},
				"end":   {"2026-07-02T00:00:00Z"},
			},
			wantErr: "window start must be before end",
		},
		{
			name:    "invalid fallback order",
			flags:   map[string][]string{"end": {"2026-07-02T00:00:00Z"}},
			config:  map[string]string{"default_start": "2026-07-03T00:00:00Z"},
			wantErr: "window start must be before end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Run(context.Background(), connector, Request{
				Path:   []string{"widget", "list"},
				Flags:  tt.flags,
				Config: connectors.RuntimeConfig{Config: tt.config},
				Limit:  1,
			}, func(connectors.Record) error { return nil })
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Run: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Run error = nil, want %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Run error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestGongCallsListDateFlagsMapToQuery(t *testing.T) {
	connector := gongCommandRunnerTestConnector(t)

	tests := []struct {
		name  string
		flags map[string][]string
		want  map[string]string
	}{
		{
			name:  "from only with offset",
			flags: map[string][]string{"from": {"2026-07-01T00:00:00-07:00"}},
			want:  map[string]string{"fromDateTime": "2026-07-01T00:00:00-07:00", "toDateTime": ""},
		},
		{
			name:  "to only",
			flags: map[string][]string{"to": {"2026-07-02T00:00:00Z"}},
			want:  map[string]string{"fromDateTime": "", "toDateTime": "2026-07-02T00:00:00Z"},
		},
		{
			name: "both",
			flags: map[string][]string{
				"from": {"2026-07-01T00:00:00Z"},
				"to":   {"2026-07-02T00:00:00Z"},
			},
			want: map[string]string{"fromDateTime": "2026-07-01T00:00:00Z", "toDateTime": "2026-07-02T00:00:00Z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotQuery url.Values
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/v2/calls" {
					t.Fatalf("request = %s %s, want GET /v2/calls", r.Method, r.URL.Path)
				}
				gotQuery = r.URL.Query()
				writeGongCallsPage(w, []int{1}, "")
			}))
			defer server.Close()

			var records []connectors.Record
			result, err := Run(context.Background(), connector, Request{
				Path:   []string{"calls", "list"},
				Flags:  tt.flags,
				Config: gongCommandRunnerTestConfig(server.URL, map[string]string{"page_size": "2"}),
				Limit:  1,
			}, func(record connectors.Record) error {
				records = append(records, record)
				return nil
			})
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if result.Count != 1 || len(records) != 1 {
				t.Fatalf("count/result records = %d/%d, want 1/1", result.Count, len(records))
			}
			for key, want := range tt.want {
				if got := gotQuery.Get(key); got != want {
					t.Fatalf("query[%s] = %q, want %q (full query %v)", key, got, want, gotQuery)
				}
			}
			if got := gotQuery.Get("limit"); got != "2" {
				t.Fatalf("query[limit] = %q, want compatibility page_size 2", got)
			}
		})
	}
}

func TestGongCallsListFromFlagOverridesStartDateConfig(t *testing.T) {
	connector := gongCommandRunnerTestConnector(t)

	tests := []struct {
		name      string
		config    map[string]string
		flags     map[string][]string
		wantFrom  string
		wantLimit string
	}{
		{
			name:      "start date config remains compatible",
			config:    map[string]string{"start_date": "2026-06-01T00:00:00Z", "page_size": "4"},
			wantFrom:  "2026-06-01T00:00:00Z",
			wantLimit: "4",
		},
		{
			name:      "explicit from wins over configured start date",
			config:    map[string]string{"start_date": "2026-06-01T00:00:00Z", "page_size": "4"},
			flags:     map[string][]string{"from": {"2026-07-01T00:00:00Z"}},
			wantFrom:  "2026-07-01T00:00:00Z",
			wantLimit: "4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotQuery url.Values
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotQuery = r.URL.Query()
				writeGongCallsPage(w, []int{1}, "")
			}))
			defer server.Close()

			_, err := Run(context.Background(), connector, Request{
				Path:   []string{"calls", "list"},
				Flags:  tt.flags,
				Config: gongCommandRunnerTestConfig(server.URL, tt.config),
				Limit:  1,
			}, func(connectors.Record) error { return nil })
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if got := gotQuery.Get("fromDateTime"); got != tt.wantFrom {
				t.Fatalf("query[fromDateTime] = %q, want %q (full query %v)", got, tt.wantFrom, gotQuery)
			}
			if got := gotQuery.Get("limit"); got != tt.wantLimit {
				t.Fatalf("query[limit] = %q, want %q", got, tt.wantLimit)
			}
		})
	}
}

func TestGongCallsListRejectsInvalidDateFlagsBeforeHTTP(t *testing.T) {
	connector := gongCommandRunnerTestConnector(t)

	tests := []struct {
		name   string
		flags  map[string][]string
		config map[string]string
		want   string
	}{
		{name: "invalid from", flags: map[string][]string{"from": {"2026-07-01"}}, want: "invalid --from"},
		{name: "invalid to", flags: map[string][]string{"to": {"tomorrow"}}, want: "invalid --to"},
		{name: "blank from", flags: map[string][]string{"from": {""}}, want: "invalid --from"},
		{name: "blank to", flags: map[string][]string{"to": {" "}}, want: "invalid --to"},
		{
			name: "equal range",
			flags: map[string][]string{
				"from": {"2026-07-01T00:00:00Z"},
				"to":   {"2026-07-01T00:00:00Z"},
			},
			want: "invalid Gong calls list date range",
		},
		{
			name: "reversed range",
			flags: map[string][]string{
				"from": {"2026-07-02T00:00:00Z"},
				"to":   {"2026-07-01T00:00:00Z"},
			},
			want: "invalid Gong calls list date range",
		},
		{
			name:   "configured start date after explicit to",
			flags:  map[string][]string{"to": {"2026-07-01T00:00:00Z"}},
			config: map[string]string{"start_date": "2026-07-02T00:00:00Z"},
			want:   "invalid Gong calls list date range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hits := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				hits++
				writeGongCallsPage(w, []int{1}, "")
			}))
			defer server.Close()

			_, err := Run(context.Background(), connector, Request{
				Path:   []string{"calls", "list"},
				Flags:  tt.flags,
				Config: gongCommandRunnerTestConfig(server.URL, mergeCommandRunnerConfig(map[string]string{"page_size": "2"}, tt.config)),
				Limit:  1,
			}, func(connectors.Record) error { return nil })
			if err == nil {
				t.Fatal("Run error = nil, want date validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Run error = %q, want to contain %q", err.Error(), tt.want)
			}
			if hits != 0 {
				t.Fatalf("server hits = %d, want 0; invalid date flags must be rejected before HTTP", hits)
			}
		})
	}
}

func mergeCommandRunnerConfig(base, overrides map[string]string) map[string]string {
	merged := map[string]string{}
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range overrides {
		merged[key] = value
	}
	return merged
}

func TestGongCallsListLimitCapsEmittedRecordsAcrossCursorPages(t *testing.T) {
	connector := gongCommandRunnerTestConnector(t)

	tests := []struct {
		name         string
		limit        int
		wantIDs      []string
		wantRequests int
	}{
		{name: "one", limit: 1, wantIDs: []string{"call-1"}, wantRequests: 1},
		{name: "below page boundary", limit: 2, wantIDs: []string{"call-1", "call-2"}, wantRequests: 1},
		{name: "at page boundary", limit: 3, wantIDs: []string{"call-1", "call-2", "call-3"}, wantRequests: 1},
		{name: "across cursor pages", limit: 5, wantIDs: []string{"call-1", "call-2", "call-3", "call-4", "call-5"}, wantRequests: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requests []url.Values
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query()
				requests = append(requests, query)
				switch cursor := query.Get("cursor"); cursor {
				case "":
					writeGongCallsPage(w, []int{1, 2, 3}, "page-2")
				case "page-2":
					writeGongCallsPage(w, []int{4, 5, 6}, "")
				default:
					t.Fatalf("unexpected cursor %q", cursor)
				}
			}))
			defer server.Close()

			var records []connectors.Record
			result, err := Run(context.Background(), connector, Request{
				Path:   []string{"calls", "list"},
				Config: gongCommandRunnerTestConfig(server.URL, map[string]string{"page_size": "3"}),
				Limit:  tt.limit,
			}, func(record connectors.Record) error {
				records = append(records, record)
				return nil
			})
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if result.Count != len(tt.wantIDs) || len(records) != len(tt.wantIDs) {
				t.Fatalf("count/result records = %d/%d, want %d", result.Count, len(records), len(tt.wantIDs))
			}
			if got := commandRunnerRecordIDs(records); fmt.Sprint(got) != fmt.Sprint(tt.wantIDs) {
				t.Fatalf("ids = %v, want %v", got, tt.wantIDs)
			}
			if len(requests) != tt.wantRequests {
				t.Fatalf("requests = %d, want %d (queries=%v)", len(requests), tt.wantRequests, requests)
			}
			for i, query := range requests {
				if got := query.Get("limit"); got != "3" {
					t.Fatalf("request %d query[limit] = %q, want compatibility page_size 3", i+1, got)
				}
			}
		})
	}
}

func gongCommandRunnerTestConnector(t *testing.T) connectors.Connector {
	t.Helper()
	connector, ok := bundleregistry.New().Get("gong")
	if !ok {
		t.Fatal("gong connector not found")
	}
	return connector
}

func gongCommandRunnerTestConfig(serverURL string, overrides map[string]string) connectors.RuntimeConfig {
	config := map[string]string{
		"base_url":  strings.TrimRight(serverURL, "/") + "/v2",
		"page_size": "2",
	}
	for key, value := range overrides {
		config[key] = value
	}
	return connectors.RuntimeConfig{
		Config: config,
		Secrets: map[string]string{
			"access_key":        "synthetic-access-key",
			"access_key_secret": "synthetic-access-key-secret",
		},
	}
}

func writeGongCallsPage(w http.ResponseWriter, ids []int, cursor string) {
	calls := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		calls = append(calls, map[string]any{
			"id":        fmt.Sprintf("call-%d", id),
			"started":   fmt.Sprintf("2026-07-%02dT00:00:00Z", id),
			"isPrivate": false,
		})
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"calls": calls,
		"records": map[string]any{
			"cursor": cursor,
		},
	})
}

func commandRunnerRecordIDs(records []connectors.Record) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, fmt.Sprint(record["id"]))
	}
	return ids
}

func TestRunBlocksNonStreamCommands(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "repo clone",
				Intent:       "local_workflow",
				Availability: "unsupported_local",
				Notes:        "local git clone workflow",
			},
		},
	}}

	tests := []struct {
		name string
		path []string
		want string
	}{
		{name: "local_workflow", path: []string{"repo", "clone"}, want: "unsupported_local"},
		{name: "unknown", path: []string{"issue", "frobnicate"}, want: "unknown command"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Run(context.Background(), connector, Request{Path: tt.path}, func(connectors.Record) error {
				t.Fatal("emit called for blocked command")
				return nil
			})
			if err == nil {
				t.Fatal("Run error = nil, want blocker")
			}
			var blocked *BlockedCommandError
			if !errors.As(err, &blocked) {
				t.Fatalf("Run error type = %T, want BlockedCommandError", err)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Run error = %q, want to contain %q", err.Error(), tt.want)
			}
		})
	}
}

func TestPreflightRejectsEngineReportedUnpromotableWrite(t *testing.T) {
	connector := &preflightFakeConnector{
		fakeConnector: reverseETLFakeConnector(),
		preflightErr:  errors.New("record_schema admits only an empty object ({})"),
	}

	err := Preflight(connector, []string{"issue", "create"})
	if err == nil {
		t.Fatal("Preflight error = nil, want unpromotable write rejection")
	}
	var blocked *BlockedCommandError
	if !errors.As(err, &blocked) {
		t.Fatalf("Preflight error type = %T, want BlockedCommandError", err)
	}
	if !strings.Contains(blocked.Reason, "not promotable") || !strings.Contains(blocked.Reason, "empty object") {
		t.Fatalf("Preflight reason = %q, want hollow record-schema detail", blocked.Reason)
	}
}

// Regression for the 2026-08-04 hollow-command incident. This drives the real
// declarative connector through the same commandrunner preflight that the CLI
// calls, rather than duplicating the engine's promotion rule in a validator.
func TestPreflightRejectsHollowDeclarativeRecordSchemas(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		want   string
	}{
		{
			name: "union root with substantive arms",
			schema: `{"oneOf":[
				{"type":"object","required":["ticket"],"properties":{"ticket":{"type":"object","properties":{"subject":{"type":"string"}}}}},
				{"type":"object","required":["tickets"],"properties":{"tickets":{"type":"array","maxItems":100,"items":{"type":"object","properties":{"id":{"type":"integer"}}}}}}
			]}`,
			want: "separate named write action",
		},
		{
			name:   "collapsed empty record",
			schema: `{"type":"object","properties":{},"additionalProperties":false}`,
			want:   "only an empty object",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := engine.New(engine.Bundle{
				Name: "widgets",
				Writes: []engine.WriteAction{{
					Name:         "update_widgets",
					Kind:         "update",
					Method:       "PUT",
					Path:         "/widgets/update_many",
					RecordSchema: json.RawMessage(tt.schema),
				}},
				CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
					Path: "widgets update-many", Intent: "reverse_etl", Availability: "implemented", Write: "update_widgets",
				}}},
			}, nil)

			err := Preflight(connector, []string{"widgets", "update-many"})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Preflight error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestBuildWriteCommandPlansReopenAndPRSharedCommands(t *testing.T) {
	connector := reverseETLFakeConnector()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}}

	tests := []struct {
		path     []string
		flags    map[string][]string
		write    string
		resource string
		record   connectors.Record
	}{
		{[]string{"issue", "reopen"}, map[string][]string{"issue-number": {"7"}}, "reopen_issue", "issue", connectors.Record{"issue_number": 7}},
		{[]string{"pr", "reopen"}, map[string][]string{"pull-number": {"9"}}, "reopen_pull_request", "pr", connectors.Record{"pull_number": 9}},
		{[]string{"pr", "comment"}, map[string][]string{"pull-number": {"9"}, "body": {"ship it"}}, "comment_issue", "pr", connectors.Record{"issue_number": 9, "body": "ship it"}},
		{[]string{"pr", "lock"}, map[string][]string{"pull-number": {"9"}}, "lock_issue", "pr", connectors.Record{"issue_number": 9}},
		{[]string{"pr", "unlock"}, map[string][]string{"pull-number": {"9"}}, "unlock_issue", "pr", connectors.Record{"issue_number": 9}},
	}

	for _, tt := range tests {
		result, err := BuildWriteCommand(context.Background(), connector, Request{Path: tt.path, Flags: tt.flags, Config: cfg})
		if err != nil {
			t.Fatalf("BuildWriteCommand(%v): %v", tt.path, err)
		}
		if result.Write != tt.write {
			t.Fatalf("%v: Write = %q, want %q", tt.path, result.Write, tt.write)
		}
		if result.TargetResource != tt.resource {
			t.Fatalf("%v: TargetResource = %q, want %q", tt.path, result.TargetResource, tt.resource)
		}
		if !result.ApprovalRequired {
			t.Fatalf("%v: ApprovalRequired = false, want true", tt.path)
		}
		if got := len(result.Record); got != len(tt.record) {
			t.Fatalf("%v: record len = %d, want %d (%+v)", tt.path, got, len(tt.record), result.Record)
		}
		for k, v := range tt.record {
			if result.Record[k] != v {
				t.Fatalf("%v: record[%s] = %v, want %v", tt.path, k, result.Record[k], v)
			}
		}
	}
}

func TestBuildWriteCommandPlansWithoutExecuting(t *testing.T) {
	connector := reverseETLFakeConnector()

	result, err := BuildWriteCommand(context.Background(), connector, Request{
		Path: []string{"issue", "create"},
		Flags: map[string][]string{
			"title": []string{"Ship connector commands"},
			"body":  []string{"Plan first"},
		},
		Config: connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}},
	})
	if err != nil {
		t.Fatalf("BuildWriteCommand: %v", err)
	}
	if result.Connector != "github" || result.Command != "issue create" || result.Write != "create_issue" {
		t.Fatalf("write command identity = %+v, want github issue create create_issue", result)
	}
	if result.MutationClass != "create" || result.TargetResource != "issue" {
		t.Fatalf("mutation target = %+v, want create issue", result)
	}
	if result.ApprovalRequired != true {
		t.Fatalf("ApprovalRequired = false, want true")
	}
	if connector.validateReq.Action != "create_issue" {
		t.Fatalf("ValidateWrite action = %q, want create_issue", connector.validateReq.Action)
	}
	if connector.dryRunReq.Action != "" {
		t.Fatalf("DryRunWrite action = %q, want not called", connector.dryRunReq.Action)
	}
	if connector.writeReq.Action != "" {
		t.Fatalf("Write action = %q, want not called", connector.writeReq.Action)
	}
	if got := result.Record["title"]; got != "Ship connector commands" {
		t.Fatalf("plan record title = %#v, want title", got)
	}
}

func TestBuildWriteCommandInfersDestructiveConfirmationFromDeleteMethod(t *testing.T) {
	connector := &fakeConnector{
		surface: &connectors.CommandSurface{
			Commands: []connectors.CommandSurfaceCommand{
				{
					Path:         "repo delete-2",
					Intent:       "reverse_etl",
					Availability: "implemented",
					Write:        "repo",
					Risk:         "critical",
				},
			},
		},
		manifest: connectors.Manifest{
			WriteActions: []connectors.WriteActionSpec{
				{Name: "repo", Method: "DELETE", Path: "/repos/{owner}/{repo}", Risk: "critical"},
			},
		},
	}

	result, err := BuildWriteCommand(context.Background(), connector, Request{Path: []string{"repo", "delete-2"}})
	if err != nil {
		t.Fatalf("BuildWriteCommand: %v", err)
	}
	if result.ConfirmationChallenge != "destructive" {
		t.Fatalf("ConfirmationChallenge = %q, want destructive", result.ConfirmationChallenge)
	}
}

func TestBuildWriteCommandAllowsEmptyRecordWhenConnectorValidatorAcceptsIt(t *testing.T) {
	connector := &fakeConnector{
		surface: &connectors.CommandSurface{
			Commands: []connectors.CommandSurfaceCommand{
				{
					Path:         "repo fork",
					Intent:       "reverse_etl",
					Availability: "implemented",
					Write:        "create_fork",
					Risk:         "creates fork",
					Approval:     "approval required",
				},
			},
		},
		manifest: connectors.Manifest{
			WriteActions: []connectors.WriteActionSpec{
				{Name: "create_fork", Method: "POST", Path: "/repos/{owner}/{repo}/forks", Risk: "creates fork"},
			},
		},
	}

	result, err := BuildWriteCommand(context.Background(), connector, Request{
		Path: []string{"repo", "fork"},
	})
	if err != nil {
		t.Fatalf("BuildWriteCommand: %v", err)
	}
	if result.Write != "create_fork" || len(result.Record) != 0 {
		t.Fatalf("result = %+v, want empty create_fork record", result)
	}
	if connector.validateReq.Action != "create_fork" {
		t.Fatalf("ValidateWrite action = %q, want create_fork", connector.validateReq.Action)
	}
}

func TestBuildWriteCommandPreviewDryRunsAndPreservesDeclaredFields(t *testing.T) {
	connector := reverseETLFakeConnector()
	connector.preview = connectors.WritePreview{
		Action:        "create_deploy_key",
		RecordsStaged: 1,
		Warnings:      []string{"resolved request: POST https://api.github.test/repos/octo/hello/keys"},
	}

	result, err := BuildWriteCommand(context.Background(), connector, Request{
		Path: []string{"repo", "deploy-key", "add"},
		Flags: map[string][]string{
			"title": []string{"deploy"},
			"key":   []string{"ssh-rsa AAAA-sensitive"},
		},
		Config:  connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}},
		Preview: true,
	})
	if err != nil {
		t.Fatalf("BuildWriteCommand: %v", err)
	}
	if result.Preview == nil {
		t.Fatalf("result = %+v, want plan and preview", result)
	}
	if connector.dryRunReq.Action != "create_deploy_key" {
		t.Fatalf("DryRunWrite action = %q, want create_deploy_key", connector.dryRunReq.Action)
	}
	if connector.writeReq.Action != "" {
		t.Fatalf("Write action = %q, want not called", connector.writeReq.Action)
	}
	if got := result.RedactedRecord["key"]; got != "ssh-rsa AAAA-sensitive" {
		t.Fatalf("plan record key = %#v, want complete input", got)
	}
}

func TestRunETLCommandPreservesDeclaredAndHeuristicSensitiveFields(t *testing.T) {
	record := connectors.Record{
		"downloadMediaUrl": "https://media.example.test/call.mp4",
		"content":          "complete body",
		"media_file_path":  "fixtures/call.mp4",
		"data_file_path":   "fixtures/crm.csv",
		"token":            "complete-token",
		"patientUuid":      "patient-uuid",
		"nested":           map[string]any{"content": "nested complete body", "key": "nested-key"},
	}
	connector := &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
		Path:         "records list",
		Intent:       "etl",
		Availability: "implemented",
		Stream:       "records",
		RedactFields: []string{"content", "token", "nested"},
	}}}, readRecords: []connectors.Record{record}}
	var records []connectors.Record
	_, err := Run(context.Background(), connector, Request{Path: []string{"records", "list"}, Limit: 1}, func(got connectors.Record) error {
		records = append(records, got)
		return nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	encoded, err := json.Marshal(records)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	for _, raw := range []string{"https://media.example.test/call.mp4", "complete body", "fixtures/call.mp4", "fixtures/crm.csv", "complete-token", "patient-uuid", "nested complete body", "nested-key"} {
		if !strings.Contains(string(encoded), raw) {
			t.Fatalf("connector command output lost %q: %s", raw, encoded)
		}
	}
	if strings.Contains(string(encoded), "***") {
		t.Fatalf("connector command output was masked: %s", encoded)
	}
}

func TestRunETLCommandPreservesDeclaredNestedFields(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		redactFields    []string
		record          connectors.Record
		preservedValues []string
	}{
		{
			name:         "patient",
			path:         "patients list",
			redactFields: []string{"uuid", "display", "identifier", "person", "givenName", "familyName", "birthdate", "gender"},
			record: connectors.Record{
				"uuid":       "patient-uuid-raw",
				"display":    "SYN-HEN-RAW Raw Name",
				"identifier": "SYN-HEN-RAW",
				"person":     map[string]any{"givenName": "Raw", "familyName": "Name", "gender": "O"},
			},
			preservedValues: []string{"patient-uuid-raw", "SYN-HEN-RAW", "Raw", "Name"},
		},
		{
			name:         "encounter",
			path:         "encounters list",
			redactFields: []string{"uuid", "display", "encounterDatetime", "patient", "visit", "obs"},
			record: connectors.Record{
				"uuid":              "encounter-uuid-raw",
				"encounterDatetime": "2026-01-01T00:00:00Z",
				"patient":           map[string]any{"uuid": "patient-uuid-raw", "display": "SYN-HEN-RAW"},
				"obs":               []any{map[string]any{"value": "fever raw"}},
			},
			preservedValues: []string{"encounter-uuid-raw", "patient-uuid-raw", "SYN-HEN-RAW", "fever raw"},
		},
		{
			name:         "observation",
			path:         "observations list",
			redactFields: []string{"uuid", "display", "person", "concept", "value", "obsDatetime", "comment"},
			record: connectors.Record{
				"uuid":        "obs-uuid-raw",
				"concept":     map[string]any{"display": "Temperature"},
				"value":       "38.6 raw",
				"obsDatetime": "2026-01-01T00:00:00Z",
				"person":      map[string]any{"uuid": "patient-uuid-raw"},
			},
			preservedValues: []string{"obs-uuid-raw", "Temperature", "38.6 raw", "patient-uuid-raw"},
		},
		{
			name:            "diagnosis",
			path:            "diagnoses list",
			redactFields:    []string{"display", "certainty", "codedAnswer", "diagnosisDateTime", "existingObs", "order"},
			record:          connectors.Record{"display": "raw diagnosis", "certainty": "CONFIRMED", "codedAnswer": map[string]any{"display": "Cold raw"}, "order": "PRIMARY"},
			preservedValues: []string{"raw diagnosis", "CONFIRMED", "Cold raw", "PRIMARY"},
		},
		{
			name:            "lab",
			path:            "lab_results list",
			redactFields:    []string{"uuid", "display", "patient", "concept", "value", "obsDatetime"},
			record:          connectors.Record{"uuid": "lab-uuid-raw", "concept": map[string]any{"display": "Serum glucose"}, "value": 120, "patient": map[string]any{"display": "SYN-HEN-RAW"}},
			preservedValues: []string{"lab-uuid-raw", "Serum glucose", "SYN-HEN-RAW", "120"},
		},
		{
			name:            "appointment",
			path:            "appointments list",
			redactFields:    []string{"uuid", "patient", "providers", "startDateTime", "endDateTime", "comments", "status"},
			record:          connectors.Record{"uuid": "appt-uuid-raw", "patient": map[string]any{"identifier": "SYN-HEN-RAW"}, "providers": []any{map[string]any{"name": "Provider Raw"}}, "comments": "raw appointment note", "status": "Scheduled"},
			preservedValues: []string{"appt-uuid-raw", "SYN-HEN-RAW", "Provider Raw", "raw appointment note", "Scheduled"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			connector := &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
				Path:         tc.path,
				Intent:       "etl",
				Availability: "implemented",
				Stream:       strings.ReplaceAll(strings.Split(tc.path, " ")[0], "-", "_"),
				RedactFields: tc.redactFields,
			}}}, readRecords: []connectors.Record{tc.record}}
			var records []connectors.Record
			_, err := Run(context.Background(), connector, Request{Path: strings.Split(tc.path, " "), Limit: 1}, func(record connectors.Record) error {
				records = append(records, record)
				return nil
			})
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			encoded, err := json.Marshal(records)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			out := string(encoded)
			for _, raw := range tc.preservedValues {
				if !strings.Contains(out, raw) {
					t.Fatalf("connector command output lost %q: %s", raw, out)
				}
			}
			if strings.Contains(out, "***") {
				t.Fatalf("connector command output was masked: %s", out)
			}
		})
	}
}

func TestBuildWriteCommandPreservesClinicalNotePreviewRecord(t *testing.T) {
	connector := &fakeConnector{
		surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
			Path:         "notes create",
			Intent:       "reverse_etl",
			Availability: "implemented",
			Write:        "create_note",
			Flags: []connectors.CommandSurfaceFlag{
				{Name: "note-type-name", Type: "string", MapsTo: "record.notes.0.noteTypeName"},
				{Name: "note-text", Type: "string", MapsTo: "record.notes.0.noteText"},
				{Name: "note-date", Type: "string", MapsTo: "record.notes.0.noteDate"},
			},
			RedactFields: []string{"notes"},
		}}},
		manifest: connectors.Manifest{WriteActions: []connectors.WriteActionSpec{{Name: "create_note", Method: http.MethodPost, Path: "/notes"}}},
	}
	cmd, err := BuildWriteCommand(context.Background(), connector, Request{
		Path:    []string{"notes", "create"},
		Preview: true,
		Flags: map[string][]string{
			"note-type-name": {"Consultation"},
			"note-text":      {"raw synthetic note text"},
			"note-date":      {"2026-01-01T00:00:00Z"},
		},
	})
	if err != nil {
		t.Fatalf("BuildWriteCommand: %v", err)
	}
	encoded, err := json.Marshal(cmd.RedactedRecord)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	out := string(encoded)
	for _, raw := range []string{"raw synthetic note text", "Consultation", "2026-01-01T00:00:00Z"} {
		if !strings.Contains(out, raw) {
			t.Fatalf("note preview lost raw note payload %q: %s", raw, out)
		}
	}
	if strings.Contains(out, "***") {
		t.Fatalf("note preview was masked: %s", out)
	}
}

func TestRunETLCommandPreservesClinicalValuesInErrors(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
		Path:         "observations list",
		Intent:       "etl",
		Availability: "implemented",
		Stream:       "observations",
		Flags: []connectors.CommandSurfaceFlag{
			{Name: "value", Type: "string", MapsTo: "query.value"},
		},
		RedactFields: []string{"patientUuid", "value"},
	}}}, readErr: errors.New("upstream rejected patient patient-uuid-raw with observation value fever raw")}
	_, err := Run(context.Background(), connector, Request{
		Path:   []string{"observations", "list"},
		Flags:  map[string][]string{"value": {"fever raw"}},
		Config: connectors.RuntimeConfig{Config: map[string]string{"patient_uuid": "patient-uuid-raw"}},
	}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Run error = nil, want upstream error")
	}
	msg := err.Error()
	for _, raw := range []string{"patient-uuid-raw", "fever raw"} {
		if !strings.Contains(msg, raw) {
			t.Fatalf("error lost %q: %s", raw, msg)
		}
	}
	if strings.Contains(msg, "***") {
		t.Fatalf("error was masked: %q", msg)
	}
}

func TestCoerceFlagValueRejectsGenericJSONFlagType(t *testing.T) {
	_, err := coerceFlagValue(connectors.CommandSurfaceFlag{Name: "diagnosis", Type: "json"}, []string{`{"nonCoded":"Synthetic"}`})
	if err == nil {
		t.Fatal("coerceFlagValue accepted generic JSON flag type; want unsupported type error")
	}
	if !strings.Contains(err.Error(), `unsupported type "json"`) {
		t.Fatalf("coerceFlagValue error = %q, want unsupported json type", err.Error())
	}
}

// The CLI metadata has long enumerated a `json` flag type, but the runner
// deliberately refused it rather than let it become an untyped body escape
// hatch. GitHub's documented oneOf arms need it for named object/array record
// fields. This contract describes the narrow change: the value is structured,
// the action schema remains the authority, and a scalar-shaped declaration is
// rejected before a command can be planned.
func TestBuildWriteCommandSupportsOnlyDeclaredStructuredJSONRecordFlags(t *testing.T) {
	newConnector := func(flagType, target string) connectors.Connector {
		return engine.New(engine.Bundle{
			Name: "widgets",
			Writes: []engine.WriteAction{{
				Name:   "create_widget",
				Kind:   "create",
				Method: http.MethodPost,
				Path:   "/widgets",
				RecordSchema: json.RawMessage(`{
					"type":"object","additionalProperties":false,"required":["payload"],
					"properties":{"payload":{"type":"object","additionalProperties":false,"required":["kind"],"properties":{"kind":{"type":"string"}}}}
				}`),
			}},
			CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
				Path: "widgets create", Intent: "reverse_etl", Availability: "implemented", Write: "create_widget",
				Flags: []engine.CLIFlag{{Name: "payload", Type: flagType, MapsTo: target, Required: true}},
			}}},
		}, nil)
	}

	connector := newConnector("json", "record.payload")
	if err := Preflight(connector, []string{"widgets", "create"}); err != nil {
		t.Fatalf("Preflight: %v", err)
	}

	tests := []struct {
		name    string
		value   string
		wantErr string
	}{
		{name: "valid object becomes typed planned record", value: `{"kind":"fixture"}`},
		{name: "malformed JSON never produces a plan", value: `{`, wantErr: "invalid JSON"},
		{name: "array cannot satisfy object field", value: `[]`, wantErr: "does not match type"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := BuildWriteCommand(context.Background(), connector, Request{
				Path: []string{"widgets", "create"}, Flags: map[string][]string{"payload": {tt.value}},
			})
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("BuildWriteCommand error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("BuildWriteCommand: %v", err)
			}
			payload, ok := command.Record["payload"].(map[string]any)
			if !ok || payload["kind"] != "fixture" {
				t.Fatalf("planned payload = %#v, want parsed object", command.Record["payload"])
			}
		})
	}

	// A structured JSON flag must never turn into a generic scalar setter or
	// body escape hatch just because a bundle says `json`: its target schema
	// must explicitly admit one top-level object/array record field before
	// preflight makes it reachable.
	for _, target := range []string{"record.payload.kind", "body.payload"} {
		invalid := newConnector("json", target)
		if err := Preflight(invalid, []string{"widgets", "create"}); err == nil || !strings.Contains(err.Error(), "structured JSON") {
			t.Fatalf("Preflight structured-json mapping %q error = %v, want declared object/array record rejection", target, err)
		}
	}
}

func TestRecordOverridesBuildsExplicitNestedScalarFields(t *testing.T) {
	record, err := recordOverrides(connectors.CommandSurfaceCommand{
		Path:         "diagnoses create",
		Intent:       "reverse_etl",
		Availability: "implemented",
		Flags: []connectors.CommandSurfaceFlag{
			{Name: "non-coded-diagnosis", Type: "string", MapsTo: "record.diagnosis.nonCoded"},
			{Name: "rank", Type: "integer", MapsTo: "record.rank"},
		},
	}, map[string][]string{
		"non-coded-diagnosis": {"Synthetic condition"},
		"rank":                {"1"},
	})
	if err != nil {
		t.Fatalf("recordOverrides: %v", err)
	}
	diagnosis, ok := record["diagnosis"].(map[string]any)
	if !ok {
		t.Fatalf("record diagnosis = %#v, want nested object", record["diagnosis"])
	}
	if diagnosis["nonCoded"] != "Synthetic condition" || record["rank"] != 1 {
		t.Fatalf("record = %+v, want explicit nested scalar diagnosis and integer rank", record)
	}
}

func TestRecordOverridesBuildsExplicitNestedArrayObjectFields(t *testing.T) {
	record, err := recordOverrides(connectors.CommandSurfaceCommand{
		Path:         "patients create",
		Intent:       "reverse_etl",
		Availability: "implemented",
		Flags: []connectors.CommandSurfaceFlag{
			{Name: "identifier", Type: "string", MapsTo: "record.identifiers.0.identifier"},
			{Name: "identifier-type", Type: "string", MapsTo: "record.identifiers.0.identifierType"},
			{Name: "identifier-location", Type: "string", MapsTo: "record.identifiers.0.location"},
			{Name: "identifier-preferred", Type: "boolean", MapsTo: "record.identifiers.0.preferred"},
			{Name: "given-name", Type: "string", MapsTo: "record.person.names.0.givenName"},
			{Name: "family-name", Type: "string", MapsTo: "record.person.names.0.familyName"},
			{Name: "gender", Type: "enum", Values: []string{"M", "F", "O"}, MapsTo: "record.person.gender"},
			{Name: "birthdate", Type: "string", MapsTo: "record.person.birthdate"},
		},
	}, map[string][]string{
		"identifier":           {"SYN-LIVE-001"},
		"identifier-type":      {"patient-id-type"},
		"identifier-location":  {"location-uuid"},
		"identifier-preferred": {"true"},
		"given-name":           {"Synthetic"},
		"family-name":          {"Connectorcase"},
		"gender":               {"O"},
		"birthdate":            {"2000-01-02"},
	})
	if err != nil {
		t.Fatalf("recordOverrides: %v", err)
	}
	identifiers, ok := record["identifiers"].([]any)
	if !ok || len(identifiers) != 1 {
		t.Fatalf("identifiers = %#v, want one-element array", record["identifiers"])
	}
	identifier, ok := identifiers[0].(map[string]any)
	if !ok || identifier["identifier"] != "SYN-LIVE-001" || identifier["preferred"] != true {
		t.Fatalf("identifier = %#v, want typed nested identifier", identifiers[0])
	}
	person, ok := record["person"].(map[string]any)
	if !ok {
		t.Fatalf("person = %#v, want object", record["person"])
	}
	names, ok := person["names"].([]any)
	if !ok || len(names) != 1 {
		t.Fatalf("names = %#v, want one-element array", person["names"])
	}
	name, ok := names[0].(map[string]any)
	if !ok || name["givenName"] != "Synthetic" || name["familyName"] != "Connectorcase" {
		t.Fatalf("name = %#v, want typed nested name", names[0])
	}
	if person["gender"] != "O" || person["birthdate"] != "2000-01-02" {
		t.Fatalf("person = %#v, want scalar demographic fields", person)
	}
}

func TestRecordOverridesRejectsSparseHugeAndConflictingNestedPaths(t *testing.T) {
	tests := []struct {
		name    string
		command connectors.CommandSurfaceCommand
		flags   map[string][]string
		want    string
	}{
		{
			name: "sparse array index",
			command: connectors.CommandSurfaceCommand{Path: "patients create", Intent: "reverse_etl", Availability: "implemented", Flags: []connectors.CommandSurfaceFlag{
				{Name: "family-name", Type: "string", MapsTo: "record.person.names.1.familyName"},
			}},
			flags: map[string][]string{"family-name": {"Connectorcase"}},
			want:  "sparse array index 1",
		},
		{
			name: "huge array index",
			command: connectors.CommandSurfaceCommand{Path: "patients create", Intent: "reverse_etl", Availability: "implemented", Flags: []connectors.CommandSurfaceFlag{
				{Name: "family-name", Type: "string", MapsTo: "record.person.names.129.familyName"},
			}},
			flags: map[string][]string{"family-name": {"Connectorcase"}},
			want:  "exceeds max 128",
		},
		{
			name: "leading zero array index",
			command: connectors.CommandSurfaceCommand{Path: "patients create", Intent: "reverse_etl", Availability: "implemented", Flags: []connectors.CommandSurfaceFlag{
				{Name: "family-name", Type: "string", MapsTo: "record.person.names.01.familyName"},
			}},
			flags: map[string][]string{"family-name": {"Connectorcase"}},
			want:  "must not have leading zeroes",
		},
		{
			name: "conflicting object and child",
			command: connectors.CommandSurfaceCommand{Path: "patients create", Intent: "reverse_etl", Availability: "implemented", Flags: []connectors.CommandSurfaceFlag{
				{Name: "person", Type: "string", MapsTo: "record.person"},
				{Name: "given-name", Type: "string", MapsTo: "record.person.names.0.givenName"},
			}},
			flags: map[string][]string{"person": {"not-json"}, "given-name": {"Synthetic"}},
			want:  "conflicting record mappings",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := recordOverrides(tc.command, tc.flags)
			if err == nil {
				t.Fatal("recordOverrides succeeded, want error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %q, want %q", err.Error(), tc.want)
			}
		})
	}
}

func TestRunReverseETLCommandRemainsNonExecutableInGenericRunner(t *testing.T) {
	connector := reverseETLFakeConnector()

	_, err := Run(context.Background(), connector, Request{
		Path:  []string{"issue", "close"},
		Flags: map[string][]string{"issue-number": []string{"101"}},
	}, func(connectors.Record) error {
		t.Fatal("emit called for reverse ETL command")
		return nil
	})
	if err == nil {
		t.Fatal("Run error = nil, want blocked generic runner")
	}
	var blocked *BlockedCommandError
	if !errors.As(err, &blocked) {
		t.Fatalf("Run error type = %T, want BlockedCommandError", err)
	}
	if !strings.Contains(err.Error(), "reverse_etl") {
		t.Fatalf("Run error = %q, want reverse_etl", err.Error())
	}
}

func TestRunReverseETLRejectsMissingWriteAndUnsupportedFlagMapping(t *testing.T) {
	tests := []struct {
		name    string
		command connectors.CommandSurfaceCommand
		flags   map[string][]string
		want    string
	}{
		{
			name: "missing write",
			command: connectors.CommandSurfaceCommand{
				Path:         "issue create",
				Intent:       "reverse_etl",
				Availability: "implemented",
				Risk:         "creates issue",
				Approval:     "approval required",
			},
			want: "must reference write action",
		},
		{
			name: "unsupported flag mapping",
			command: connectors.CommandSurfaceCommand{
				Path:         "issue create",
				Intent:       "reverse_etl",
				Availability: "implemented",
				Write:        "create_issue",
				Risk:         "creates issue",
				Approval:     "approval required",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "state", Type: "string", MapsTo: "query.state"},
				},
			},
			flags: map[string][]string{"state": []string{"open"}},
			want:  "unsupported target",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &fakeConnector{
				surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{tt.command}},
				manifest: connectors.Manifest{WriteActions: []connectors.WriteActionSpec{
					{Name: "create_issue", Method: "POST", Path: "/issues"},
					{Name: "create_fork", Method: "POST", Path: "/forks"},
				}},
			}
			_, err := BuildWriteCommand(context.Background(), connector, Request{
				Path:  strings.Fields(tt.command.Path),
				Flags: tt.flags,
			})
			if err == nil {
				t.Fatal("Run error = nil, want rejection")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Run error = %q, want %q", err.Error(), tt.want)
			}
		})
	}
}

func TestRunImplementedOperationCommandRequiresTypedMetadata(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "project list",
				Intent:       "direct_read",
				Availability: "implemented",
				Operation:    "github.projects.list",
			},
		},
	}}

	_, err := Run(context.Background(), connector, Request{Path: []string{"project", "list"}}, func(connectors.Record) error {
		t.Fatal("emit called for feature-gated operation command")
		return nil
	})
	if err == nil {
		t.Fatal("Run error = nil, want feature gate")
	}
	var blocked *BlockedCommandError
	if !errors.As(err, &blocked) {
		t.Fatalf("Run error type = %T, want BlockedCommandError", err)
	}
	if !strings.Contains(err.Error(), "operation direct_read commands require exactly one api_surface endpoint") {
		t.Fatalf("Run error = %q, want typed operation metadata gate", err.Error())
	}
}

func reverseETLFakeConnector() *fakeConnector {
	return &fakeConnector{
		surface: &connectors.CommandSurface{
			Commands: []connectors.CommandSurfaceCommand{
				{
					Path:         "issue create",
					Intent:       "reverse_etl",
					Availability: "implemented",
					Write:        "create_issue",
					Risk:         "creates a visible issue",
					Approval:     "approval required",
					Flags: []connectors.CommandSurfaceFlag{
						{Name: "title", Type: "string", MapsTo: "record.title"},
						{Name: "body", Type: "string", MapsTo: "record.body"},
					},
				},
				{
					Path:         "issue close",
					Intent:       "reverse_etl",
					Availability: "implemented",
					Write:        "close_issue",
					Risk:         "closes an issue",
					Approval:     "approval required",
					Flags: []connectors.CommandSurfaceFlag{
						{Name: "issue-number", Type: "integer", MapsTo: "record.issue_number"},
					},
				},
				{
					Path:         "repo deploy-key add",
					Intent:       "reverse_etl",
					Availability: "implemented",
					Write:        "create_deploy_key",
					Risk:         "adds deploy key",
					Approval:     "approval required",
					Flags: []connectors.CommandSurfaceFlag{
						{Name: "title", Type: "string", MapsTo: "record.title"},
						{Name: "key", Type: "string", MapsTo: "record.key"},
					},
				},
				{
					Path:         "issue reopen",
					Intent:       "reverse_etl",
					Availability: "implemented",
					Write:        "reopen_issue",
					Risk:         "reopens an issue",
					Approval:     "approval required",
					Flags: []connectors.CommandSurfaceFlag{
						{Name: "issue-number", Type: "integer", MapsTo: "record.issue_number"},
					},
				},
				{
					Path:         "pr reopen",
					Intent:       "reverse_etl",
					Availability: "implemented",
					Write:        "reopen_pull_request",
					Risk:         "reopens a pull request",
					Approval:     "approval required",
					Flags: []connectors.CommandSurfaceFlag{
						{Name: "pull-number", Type: "integer", MapsTo: "record.pull_number"},
					},
				},
				{
					Path:         "pr comment",
					Intent:       "reverse_etl",
					Availability: "implemented",
					Write:        "comment_issue",
					Risk:         "comments on a pull request",
					Approval:     "approval required",
					Flags: []connectors.CommandSurfaceFlag{
						{Name: "pull-number", Type: "integer", MapsTo: "record.issue_number"},
						{Name: "body", Type: "string", MapsTo: "record.body"},
					},
				},
				{
					Path:         "pr lock",
					Intent:       "reverse_etl",
					Availability: "implemented",
					Write:        "lock_issue",
					Risk:         "locks a pull request",
					Approval:     "approval required",
					Flags: []connectors.CommandSurfaceFlag{
						{Name: "pull-number", Type: "integer", MapsTo: "record.issue_number"},
					},
				},
				{
					Path:         "pr unlock",
					Intent:       "reverse_etl",
					Availability: "implemented",
					Write:        "unlock_issue",
					Risk:         "unlocks a pull request",
					Approval:     "approval required",
					Flags: []connectors.CommandSurfaceFlag{
						{Name: "pull-number", Type: "integer", MapsTo: "record.issue_number"},
					},
				},
			},
		},
		manifest: connectors.Manifest{
			WriteActions: []connectors.WriteActionSpec{
				{Name: "create_issue", Method: "POST", Path: "/repos/{owner}/{repo}/issues", Risk: "creates issue"},
				{Name: "close_issue", Method: "PATCH", Path: "/repos/{owner}/{repo}/issues/{issue_number}", Risk: "closes issue"},
				{Name: "reopen_issue", Method: "PATCH", Path: "/repos/{owner}/{repo}/issues/{issue_number}", Risk: "reopens issue"},
				{Name: "reopen_pull_request", Method: "PATCH", Path: "/repos/{owner}/{repo}/pulls/{pull_number}", Risk: "reopens pull request"},
				{Name: "comment_issue", Method: "POST", Path: "/repos/{owner}/{repo}/issues/{issue_number}/comments", Risk: "comments on issue"},
				{Name: "lock_issue", Method: "PUT", Path: "/repos/{owner}/{repo}/issues/{issue_number}/lock", Risk: "locks issue"},
				{Name: "unlock_issue", Method: "DELETE", Path: "/repos/{owner}/{repo}/issues/{issue_number}/lock", Risk: "unlocks issue"},
				{Name: "create_deploy_key", Method: "POST", Path: "/repos/{owner}/{repo}/keys", Risk: "adds deploy key"},
			},
		},
	}
}

func TestRunRejectsUnknownFlagAndInvalidEnum(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "issue list",
				Intent:       "etl",
				Availability: "implemented",
				Stream:       "issues",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "state", Type: "enum", Values: []string{"open", "closed", "all"}, MapsTo: "query.state"},
				},
			},
		},
	}}

	tests := []struct {
		name  string
		flags map[string][]string
		want  string
	}{
		{name: "unknown flag", flags: map[string][]string{"author": []string{"octocat"}}, want: "unknown flag"},
		{name: "invalid enum", flags: map[string][]string{"state": []string{"merged"}}, want: "invalid --state"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Run(context.Background(), connector, Request{
				Path:  []string{"issue", "list"},
				Flags: tt.flags,
			}, func(connectors.Record) error {
				t.Fatal("emit called for invalid flags")
				return nil
			})
			if err == nil {
				t.Fatal("Run error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Run error = %q, want to contain %q", err.Error(), tt.want)
			}
		})
	}
}

func TestRunImplementedDirectReadCommand(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "repo read-file",
				Intent:       "direct_read",
				Availability: "implemented",
				APISurface: []connectors.CommandSurfaceEndpointRef{
					{Method: "GET", Path: "/repos/{owner}/{repo}/contents/{path}"},
				},
				OutputPolicy: "repository_contents_file_metadata",
				RedactFields: []string{"path"},
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "path", Type: "string", MapsTo: "path.path"},
				},
			},
		},
	}}

	result, err := Run(context.Background(), connector, Request{
		Path:  []string{"repo", "read-file"},
		Flags: map[string][]string{"path": []string{"README.md"}},
		Config: connectors.RuntimeConfig{Config: map[string]string{
			"owner": "octo",
			"repo":  "hello",
		}},
	}, func(connectors.Record) error {
		t.Fatal("emit called for direct-read command")
		return nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.DirectRead == nil {
		t.Fatalf("DirectRead = nil, want result")
	}
	if connector.directReadReq.Method != "GET" {
		t.Fatalf("direct read method = %q, want GET", connector.directReadReq.Method)
	}
	if connector.directReadReq.PathParams["path"] != "README.md" {
		t.Fatalf("direct read path param = %q, want README.md", connector.directReadReq.PathParams["path"])
	}
	if connector.directReadReq.OutputPolicy != "repository_contents_file_metadata" {
		t.Fatalf("direct read output policy = %q, want repository_contents_file_metadata", connector.directReadReq.OutputPolicy)
	}
	if len(connector.directReadReq.RedactFields) != 0 {
		t.Fatalf("direct read RedactFields = %#v, want empty", connector.directReadReq.RedactFields)
	}
}

func TestRunImplementedOperationDirectReadCommand(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "meetings integration-status",
				Intent:       "direct_read",
				Availability: "implemented",
				Operation:    "gong.meetings_integration_status",
				APISurface: []connectors.CommandSurfaceEndpointRef{
					{Method: "POST", Path: "/v2/meetings/integration/status"},
				},
				OutputPolicy: "json_redacted",
				RedactFields: []string{"email"},
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "email", Type: "string_array", MapsTo: "body.emails"},
					{Name: "header-x-request-mode", Type: "enum", Values: []string{"safe", "full"}, MapsTo: "header.X-Request-Mode"},
				},
			},
		},
	}}

	result, err := Run(context.Background(), connector, Request{
		Path:  []string{"meetings", "integration-status"},
		Flags: map[string][]string{"email": {"ada@example.com", "grace@example.com"}, "header-x-request-mode": {"safe"}},
	}, func(connectors.Record) error {
		t.Fatal("emit called for operation direct-read command")
		return nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.DirectRead == nil {
		t.Fatalf("DirectRead = nil, want result")
	}
	if connector.operationDirectReadReq.Operation != "gong.meetings_integration_status" {
		t.Fatalf("operation = %q", connector.operationDirectReadReq.Operation)
	}
	if connector.operationDirectReadReq.MaxBytes != MaxOperationDirectReadBytes {
		t.Fatalf("operation max bytes = %d, want %d", connector.operationDirectReadReq.MaxBytes, MaxOperationDirectReadBytes)
	}
	emails, ok := connector.operationDirectReadReq.Body["emails"].([]string)
	if !ok || len(emails) != 2 || emails[0] != "ada@example.com" || emails[1] != "grace@example.com" {
		t.Fatalf("operation body = %#v, want typed emails", connector.operationDirectReadReq.Body)
	}
	if len(connector.operationDirectReadReq.RedactFields) != 0 {
		t.Fatalf("operation direct read RedactFields = %#v, want empty", connector.operationDirectReadReq.RedactFields)
	}
	if got := connector.operationDirectReadReq.Headers["X-Request-Mode"]; got != "safe" {
		t.Fatalf("operation request header = %q, want exact declared value", got)
	}
}

func TestRunOperationDirectReadRejectsRepeatedHeaderFlagsBeforeDispatch(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
		Path: "widgets list", Intent: "direct_read", Availability: "implemented", Operation: "acme.widgets.list", OutputPolicy: "json_redacted",
		APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodGet, Path: "/widgets"}},
		Flags:      []connectors.CommandSurfaceFlag{{Name: "header-x-mode", Type: "enum", Values: []string{"safe", "full"}, MapsTo: "header.X-Mode"}},
	}}}}
	_, err := Run(context.Background(), connector, Request{Path: []string{"widgets", "list"}, Flags: map[string][]string{"header-x-mode": {"invalid", "safe"}}}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "accept exactly one value") {
		t.Fatalf("Run error = %v, want repeated header refusal", err)
	}
	if connector.operationDirectReadReq.Operation != "" {
		t.Fatalf("duplicate header reached operation dispatch: %#v", connector.operationDirectReadReq)
	}
}

func TestRunOperationDirectReadPreservesDeclaredRepeatableHeaderValues(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
		Path: "widgets list", Intent: "direct_read", Availability: "implemented", Operation: "acme.widgets.list", OutputPolicy: "json_redacted",
		APISurface: []connectors.CommandSurfaceEndpointRef{{Method: http.MethodGet, Path: "/widgets"}},
		Flags:      []connectors.CommandSurfaceFlag{{Name: "header-x-mode", Type: "enum", Values: []string{"safe", "full"}, Repeatable: true, MapsTo: "header.X-Mode"}},
	}}}}
	_, err := Run(context.Background(), connector, Request{Path: []string{"widgets", "list"}, Flags: map[string][]string{"header-x-mode": {"invalid", "safe"}}}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "want one of") {
		t.Fatalf("Run error = %v, want every repeated value validated", err)
	}
	if connector.operationDirectReadReq.Operation != "" {
		t.Fatalf("invalid repeated header reached operation dispatch: %#v", connector.operationDirectReadReq)
	}
	_, err = Run(context.Background(), connector, Request{Path: []string{"widgets", "list"}, Flags: map[string][]string{"header-x-mode": {"safe", "full"}}}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Run repeatable header: %v", err)
	}
	if got := connector.operationDirectReadReq.HeaderValues["X-Mode"]; len(got) != 2 || got[0] != "safe" || got[1] != "full" {
		t.Fatalf("repeatable header values = %#v, want ordered values", connector.operationDirectReadReq.HeaderValues)
	}
}

func TestRunOperationDirectReadAdmitsOnlyPreflightedStructuredGraphQLVariable(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
		Path:         "graphql query widgets",
		Intent:       "direct_read",
		Availability: "implemented",
		Operation:    "github.graphql.query.widgets",
		APISurface:   []connectors.CommandSurfaceEndpointRef{{Method: http.MethodPost, Path: "/graphql"}},
		OutputPolicy: "json_redacted",
		Flags: []connectors.CommandSurfaceFlag{
			{Name: "input", Type: "json", Required: true, MapsTo: "body.input"},
		},
	}}}}

	_, err := Run(context.Background(), connector, Request{
		Path:  []string{"graphql", "query", "widgets"},
		Flags: map[string][]string{"input": {`{"ids":["widget-1"]}`}},
	}, func(connectors.Record) error {
		t.Fatal("emit called for operation direct-read command")
		return nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got, want := connector.operationJSONVariable, (operationStructuredJSONVariablePreflightCall{operation: "github.graphql.query.widgets", variable: "input"}); got != want {
		t.Fatalf("structured JSON preflight = %#v, want %#v", got, want)
	}
	input, ok := connector.operationDirectReadReq.Body["input"].(map[string]any)
	if !ok {
		t.Fatalf("typed GraphQL input = %#v, want object", connector.operationDirectReadReq.Body["input"])
	}
	ids, ok := input["ids"].([]any)
	if !ok || len(ids) != 1 || ids[0] != "widget-1" {
		t.Fatalf("typed GraphQL ids = %#v, want one declared item", input["ids"])
	}

	connector.operationJSONVariableErr = errors.New("variable is not a closed object or array")
	_, err = Run(context.Background(), connector, Request{
		Path:  []string{"graphql", "query", "widgets"},
		Flags: map[string][]string{"input": {`{"ids":["widget-2"]}`}},
	}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "closed object or array") {
		t.Fatalf("Run rejected GraphQL variable = %v, want declaration-owned preflight rejection", err)
	}
}

func TestRunOperationDirectReadPassesDeclaredPlainTextBody(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "markdown raw",
				Intent:       "direct_read",
				Availability: "implemented",
				Operation:    "github.render_raw_markdown",
				APISurface: []connectors.CommandSurfaceEndpointRef{
					{Method: http.MethodPost, Path: "/markdown/raw"},
				},
				OutputPolicy: "text",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "text", Type: "string", Required: true, MapsTo: "body"},
				},
			},
		},
	}}

	const source = "# hello\n"
	_, err := Run(context.Background(), connector, Request{
		Path:  []string{"markdown", "raw"},
		Flags: map[string][]string{"text": {source}},
	}, func(connectors.Record) error {
		t.Fatal("emit called for operation direct-read command")
		return nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if connector.operationDirectReadReq.RawBody == nil || *connector.operationDirectReadReq.RawBody != source {
		t.Fatalf("RawBody = %#v, want literal source", connector.operationDirectReadReq.RawBody)
	}
	if len(connector.operationDirectReadReq.Body) != 0 {
		t.Fatalf("Body = %#v, want no JSON fields for raw input", connector.operationDirectReadReq.Body)
	}
}

func TestRunOperationDirectReadRejectsMixedRawAndJSONBodyMappings(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "markdown raw",
				Intent:       "direct_read",
				Availability: "implemented",
				Operation:    "github.render_raw_markdown",
				APISurface: []connectors.CommandSurfaceEndpointRef{
					{Method: http.MethodPost, Path: "/markdown/raw"},
				},
				OutputPolicy: "text",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "text", Type: "string", Required: true, MapsTo: "body"},
					{Name: "context", Type: "string", MapsTo: "body.context"},
				},
			},
		},
	}}

	_, err := Run(context.Background(), connector, Request{
		Path:  []string{"markdown", "raw"},
		Flags: map[string][]string{"text": {"# hello"}, "context": {"octo/example"}},
	}, func(connectors.Record) error {
		t.Fatal("emit called for rejected operation direct-read command")
		return nil
	})
	if err == nil {
		t.Fatal("Run error = nil, want mixed body rejection")
	}
	if !strings.Contains(err.Error(), "raw body cannot mix with JSON body fields") {
		t.Fatalf("Run error = %q, want mixed body rejection", err)
	}
	if connector.operationDirectReadReq.RawBody != nil {
		t.Fatalf("OperationDirectRead dispatched with raw body %#v", connector.operationDirectReadReq.RawBody)
	}
}

func TestRunOperationDirectReadPlainTextBodyOnlyAdmitsDocumentWhitespace(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "markdown raw",
				Intent:       "direct_read",
				Availability: "implemented",
				Operation:    "github.render_raw_markdown",
				APISurface: []connectors.CommandSurfaceEndpointRef{
					{Method: http.MethodPost, Path: "/markdown/raw"},
				},
				OutputPolicy: "text",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "text", Type: "string", Required: true, MapsTo: "body"},
				},
			},
		},
	}}

	_, err := Run(context.Background(), connector, Request{
		Path:  []string{"markdown", "raw"},
		Flags: map[string][]string{"text": {"# hello\x00"}},
	}, func(connectors.Record) error {
		t.Fatal("emit called for rejected operation direct-read command")
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported control character") {
		t.Fatalf("Run error = %v, want raw control-character rejection", err)
	}
	if connector.operationDirectReadReq.RawBody != nil {
		t.Fatalf("OperationDirectRead dispatched with raw body %#v", connector.operationDirectReadReq.RawBody)
	}
}

func TestPreflightOperationDirectReadRejectsNonExecutableOperationMetadata(t *testing.T) {
	connector := &fakeConnector{
		operationReadPreflightErr: errors.New("operation direct read requires rest_read or provider_search operation, got \"graphql_query\""),
		surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
			Path:         "meetings integration-status",
			Intent:       "direct_read",
			Availability: "implemented",
			Operation:    "gong.meetings_integration_status",
			APISurface: []connectors.CommandSurfaceEndpointRef{
				{Method: http.MethodPost, Path: "/v2/meetings/integration/status"},
			},
			OutputPolicy: "json_redacted",
		}}},
	}

	err := Preflight(connector, []string{"meetings", "integration-status"})
	if err == nil {
		t.Fatal("Preflight error = nil, want loaded operation rejection")
	}
	if !strings.Contains(err.Error(), "operation direct read metadata is not executable") {
		t.Fatalf("Preflight error = %q, want executable metadata rejection", err)
	}
	if got, want := connector.operationReadPreflight, (operationDirectReadPreflightCall{
		operation:    "gong.meetings_integration_status",
		method:       http.MethodPost,
		path:         "/v2/meetings/integration/status",
		maxBytes:     MaxOperationDirectReadBytes,
		outputPolicy: "json_redacted",
	}); got != want {
		t.Fatalf("operation preflight = %#v, want %#v", got, want)
	}
	if connector.operationDirectReadReq.Operation != "" {
		t.Fatalf("OperationDirectRead dispatched during preflight: %+v", connector.operationDirectReadReq)
	}
}

func TestBuildOperationDirectWriteCommandUsesTypedInputsAndPlanLifecycle(t *testing.T) {
	connector := &fakeConnector{
		directWriteMetadata: connectors.OperationDirectWriteMetadata{
			Operation:             "acme.vote",
			MutationClass:         "destructive",
			Risk:                  "high",
			Approval:              "plan-preview-confirm-execute",
			ConfirmationChallenge: "destructive",
			OutputPolicy:          "json_redacted",
			Batchable:             false,
		},
		surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
			Path:         "vote",
			Intent:       "direct_write",
			Availability: "implemented",
			Operation:    "acme.vote",
			APISurface:   []connectors.CommandSurfaceEndpointRef{{Method: http.MethodPost, Path: "/api/vote"}},
			OutputPolicy: "json_redacted",
			RedactFields: []string{"id"},
			Flags: []connectors.CommandSurfaceFlag{
				{Name: "id", Type: "string", MapsTo: "body.id", Required: true},
				{Name: "dir", Type: "integer", MapsTo: "body.dir", Required: true},
			},
		}}},
	}

	command, err := BuildWriteCommand(context.Background(), connector, Request{
		Path:  []string{"vote"},
		Flags: map[string][]string{"id": {"t3_abc"}, "dir": {"1"}},
	})
	if err != nil {
		t.Fatalf("BuildWriteCommand: %v", err)
	}
	if command.Operation != "acme.vote" || command.Write != "acme.vote" {
		t.Fatalf("operation/write = %q/%q, want acme.vote", command.Operation, command.Write)
	}
	if command.Batchable {
		t.Fatal("direct write command made batchable:false operation batchable")
	}
	if got := command.Record["dir"]; got != 1 {
		t.Fatalf("typed body dir = %#v (%T), want integer 1", got, got)
	}
	if got := command.RedactedRecord["id"]; got != "t3_abc" {
		t.Fatalf("direct-write preview record id = %#v, want complete input", got)
	}
	if command.ConfirmationChallenge != "destructive" {
		t.Fatalf("confirmation = %q, want destructive", command.ConfirmationChallenge)
	}

	_, err = Run(context.Background(), connector, Request{Path: []string{"vote"}}, func(connectors.Record) error {
		t.Fatal("direct_write bypassed the plan lifecycle")
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "plan, preview, approval, execute") {
		t.Fatalf("Run(direct_write) error = %v, want plan lifecycle block", err)
	}
}

func TestGitHubUserDraftCommandBuildsFixedGraphQLMutation(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load GitHub bundle: %v", err)
	}
	connector := engine.New(bundle, nil)

	command, err := BuildWriteCommand(context.Background(), connector, Request{
		Path: []string{"projects", "create-draft-item-for-authenticated-user"},
		Flags: map[string][]string{
			"input": {`{"projectId":"PVT_kwDOBigProject","title":"pm-cert-draft","body":"fixed GraphQL route"}`},
		},
	})
	if err != nil {
		t.Fatalf("BuildWriteCommand: %v", err)
	}
	if got, want := command.Operation, "github.graphql.mutation.add-project-v2-draft-issue"; got != want {
		t.Fatalf("operation = %q, want %q", got, want)
	}
	input, ok := command.Record["input"].(map[string]any)
	if !ok {
		t.Fatalf("record input = %#v, want closed GraphQL input object", command.Record["input"])
	}
	if input["projectId"] != "PVT_kwDOBigProject" || input["title"] != "pm-cert-draft" || input["body"] != "fixed GraphQL route" {
		t.Fatalf("GraphQL input = %#v, want exact projectId/title/body", input)
	}
}

func TestBuildOperationDirectWriteCommandKeepsCompleteInputError(t *testing.T) {
	connector := &fakeConnector{
		directWriteMetadata: connectors.OperationDirectWriteMetadata{
			Operation:    "acme.vote",
			OutputPolicy: "json_redacted",
		},
		surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
			Path:         "vote",
			Intent:       "direct_write",
			Availability: "implemented",
			Operation:    "acme.vote",
			APISurface:   []connectors.CommandSurfaceEndpointRef{{Method: http.MethodPost, Path: "/api/vote"}},
			OutputPolicy: "json_redacted",
			RedactFields: []string{"dir"},
			Flags: []connectors.CommandSurfaceFlag{
				{Name: "dir", Type: "integer", MapsTo: "body.dir"},
			},
		}}},
	}

	_, err := BuildWriteCommand(context.Background(), connector, Request{
		Path:  []string{"vote"},
		Flags: map[string][]string{"dir": {"complete-input"}},
	})
	if err == nil {
		t.Fatal("BuildWriteCommand error = nil, want invalid integer")
	}
	if !strings.Contains(err.Error(), "complete-input") {
		t.Fatalf("BuildWriteCommand error = %q, want complete input value", err)
	}
}

func TestPreflightOperationDirectWriteRejectsMismatchedOperationPolicy(t *testing.T) {
	connector := &fakeConnector{
		directWriteMetadata: connectors.OperationDirectWriteMetadata{
			Operation:    "acme.vote",
			OutputPolicy: "write_result_redacted",
		},
		surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
			Path:         "vote",
			Intent:       "direct_write",
			Availability: "implemented",
			Operation:    "acme.vote",
			APISurface:   []connectors.CommandSurfaceEndpointRef{{Method: http.MethodPost, Path: "/api/vote"}},
			OutputPolicy: "json_redacted",
		}}},
	}

	err := Preflight(connector, []string{"vote"})
	if err == nil || !strings.Contains(err.Error(), "output_policy") {
		t.Fatalf("Preflight(direct_write) error = %v, want mismatched output_policy rejection", err)
	}
}

func TestPreflightOperationDirectWriteRequiresRuntimeBinding(t *testing.T) {
	connector := &fakeConnector{
		operationWritePreflightErr: errors.New("operation direct write path \"/graphql/raw\" does not match declared operation path \"/graphql\""),
		directWriteMetadata: connectors.OperationDirectWriteMetadata{
			Operation:    "github.delete_issue",
			OutputPolicy: "json_redacted",
		},
		surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
			Path:         "issue delete",
			Intent:       "direct_write",
			Availability: "implemented",
			Operation:    "github.delete_issue",
			APISurface:   []connectors.CommandSurfaceEndpointRef{{Method: http.MethodPost, Path: "/graphql/raw"}},
			OutputPolicy: "json_redacted",
		}}},
	}

	err := Preflight(connector, []string{"issue", "delete"})
	if err == nil || !strings.Contains(err.Error(), "operation direct write metadata is not executable") {
		t.Fatalf("Preflight(direct_write) error = %v, want runtime binding rejection", err)
	}
	if got, want := connector.operationWritePreflight, (operationDirectWritePreflightCall{
		operation:    "github.delete_issue",
		method:       http.MethodPost,
		path:         "/graphql/raw",
		outputPolicy: "json_redacted",
	}); got != want {
		t.Fatalf("operation write preflight = %#v, want %#v", got, want)
	}
}

func TestRunOperationDirectReadRequiredQueryFlags(t *testing.T) {
	command := connectors.CommandSurfaceCommand{
		Path:         "customers invoices list",
		Intent:       "direct_read",
		Availability: "implemented",
		Operation:    "google_ads.customers.invoices.list",
		APISurface: []connectors.CommandSurfaceEndpointRef{
			{Method: "GET", Path: "/v22/customers/{customer_id}/invoices"},
		},
		OutputPolicy: "json_redacted",
		Flags: []connectors.CommandSurfaceFlag{
			{Name: "billing-setup", Type: "string", MapsTo: "query.billingSetup", Required: true},
			{Name: "issue-month", Type: "enum", Values: []string{"JANUARY", "FEBRUARY"}, MapsTo: "query.issueMonth", Required: true},
			{Name: "issue-year", Type: "string", MapsTo: "query.issueYear", Required: true},
		},
	}

	t.Run("missing required flag", func(t *testing.T) {
		connector := &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{command}}}
		_, err := Run(context.Background(), connector, Request{
			Path: []string{"customers", "invoices", "list"},
			Flags: map[string][]string{
				"issue-month": {"JANUARY"},
				"issue-year":  {"2026"},
			},
		}, func(connectors.Record) error {
			t.Fatal("emit called for rejected operation direct-read command")
			return nil
		})
		if err == nil || !strings.Contains(err.Error(), "missing required flag --billing-setup") {
			t.Fatalf("Run error = %v, want missing billing setup", err)
		}
		var missingRequired *MissingRequiredFlagError
		if !errors.As(err, &missingRequired) {
			t.Fatalf("Run error = %T %v, want MissingRequiredFlagError", err, err)
		}
		if missingRequired.Command != command.Path || missingRequired.Flag != "billing-setup" {
			t.Fatalf("MissingRequiredFlagError = %+v, want command %q flag billing-setup", missingRequired, command.Path)
		}
		if connector.operationDirectReadReq.Operation != "" {
			t.Fatalf("operation direct read executed: %+v", connector.operationDirectReadReq)
		}
	})

	t.Run("maps required query flags", func(t *testing.T) {
		connector := &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{command}}}
		_, err := Run(context.Background(), connector, Request{
			Path: []string{"customers", "invoices", "list"},
			Flags: map[string][]string{
				"billing-setup": {"customers/123/billingSetups/456"},
				"issue-month":   {"JANUARY"},
				"issue-year":    {"2026"},
			},
		}, func(connectors.Record) error {
			t.Fatal("emit called for operation direct-read command")
			return nil
		})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		query := connector.operationDirectReadReq.Query
		if query["billingSetup"] != "customers/123/billingSetups/456" || query["issueMonth"] != "JANUARY" || query["issueYear"] != "2026" {
			t.Fatalf("operation query = %#v", query)
		}
	})
}

func TestRunGongTypedPOSTDirectReadsIncludingTranscript(t *testing.T) {
	tests := []struct {
		name         string
		path         []string
		endpointPath string
		flags        map[string][]string
	}{
		{name: "calls extensive", path: []string{"calls", "extensive"}, endpointPath: "/v2/calls/extensive", flags: map[string][]string{"call-id": {"call-1"}}},
		{name: "call access", path: []string{"calls", "users-access", "get"}, endpointPath: "/v2/calls/users-access", flags: map[string][]string{"call-id": {"call-1"}}},
		{name: "users extensive", path: []string{"users", "extensive"}, endpointPath: "/v2/users/extensive"},
		{name: "tasks", path: []string{"tasks", "list"}, endpointPath: "/v2/tasks", flags: map[string][]string{"user-id": {"user-1"}, "status": {"OPEN"}, "task-action": {"CALL"}, "task-type": {"FLOW"}}},
		{name: "interaction stats", path: []string{"stats", "interaction"}, endpointPath: "/v2/stats/interaction", flags: map[string][]string{"from-date": {"2026-01-01"}, "to-date": {"2026-01-02"}}},
		{name: "scorecards", path: []string{"stats", "activity-scorecards"}, endpointPath: "/v2/stats/activity/scorecards", flags: map[string][]string{"scorecard-id": {"scorecard-1"}}},
		{name: "day by day", path: []string{"stats", "activity-day-by-day"}, endpointPath: "/v2/stats/activity/day-by-day", flags: map[string][]string{"from-date": {"2026-01-01"}, "to-date": {"2026-01-02"}}},
		{name: "aggregate", path: []string{"stats", "activity-aggregate"}, endpointPath: "/v2/stats/activity/aggregate", flags: map[string][]string{"from-date": {"2026-01-01"}, "to-date": {"2026-01-02"}}},
		{name: "aggregate by period", path: []string{"stats", "activity-aggregate-by-period"}, endpointPath: "/v2/stats/activity/aggregate-by-period", flags: map[string][]string{"from-date": {"2026-01-01"}, "to-date": {"2026-01-02"}, "aggregation-period": {"DAY"}}},
		{name: "transcript", path: []string{"calls", "transcript"}, endpointPath: "/v2/calls/transcript", flags: map[string][]string{"call-id": {"call-1"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != tt.endpointPath {
					t.Errorf("request = %s %s, want POST %s", r.Method, r.URL.Path, tt.endpointPath)
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode request body: %v", err)
				}
				_, _ = w.Write([]byte(`{"ok":true}`))
			}))
			defer server.Close()

			bundle, err := engine.Load(defs.FS, "gong")
			if err != nil {
				t.Fatalf("load Gong bundle: %v", err)
			}
			bundle.HTTP.URL = server.URL
			connector := engine.New(bundle, nil)
			result, err := Run(context.Background(), connector, Request{
				Path:  tt.path,
				Flags: tt.flags,
				Config: connectors.RuntimeConfig{Secrets: map[string]string{
					"access_key":        "fixture_access_key",
					"access_key_secret": "fixture_access_key_secret",
				}},
			}, func(connectors.Record) error {
				t.Fatal("emit called for direct read")
				return nil
			})
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if result.DirectRead == nil || body == nil {
				t.Fatalf("result/body = %+v / %+v, want direct read and typed body", result, body)
			}
			if tt.name == "transcript" {
				filter, ok := body["filter"].(map[string]any)
				if !ok {
					t.Fatalf("transcript filter = %#v, want object", body["filter"])
				}
				callIDs, ok := filter["callIds"].([]any)
				if !ok || len(callIDs) != 1 || callIDs[0] != "call-1" {
					t.Fatalf("transcript callIds = %#v, want [call-1]", filter["callIds"])
				}
			}
		})
	}
}

func TestRunOperationDirectReadRejectsRawBodyAndMissingPolicy(t *testing.T) {
	tests := []struct {
		name    string
		command connectors.CommandSurfaceCommand
		flags   map[string][]string
		want    string
	}{
		{
			name: "unknown raw body flag",
			command: connectors.CommandSurfaceCommand{
				Path:         "meetings integration-status",
				Intent:       "direct_read",
				Availability: "implemented",
				Operation:    "gong.meetings_integration_status",
				APISurface:   []connectors.CommandSurfaceEndpointRef{{Method: "POST", Path: "/v2/meetings/integration/status"}},
				OutputPolicy: "json_redacted",
				Flags:        []connectors.CommandSurfaceFlag{{Name: "email", Type: "string_array", MapsTo: "body.emails"}},
			},
			flags: map[string][]string{"body": {`{"emails":["ada@example.com"]}`}},
			want:  "unknown flag",
		},
		{
			name: "missing output policy keeps operation blocked",
			command: connectors.CommandSurfaceCommand{
				Path:         "meetings integration-status",
				Intent:       "direct_read",
				Availability: "implemented",
				Operation:    "gong.meetings_integration_status",
				APISurface:   []connectors.CommandSurfaceEndpointRef{{Method: "POST", Path: "/v2/meetings/integration/status"}},
				Flags:        []connectors.CommandSurfaceFlag{{Name: "email", Type: "string_array", MapsTo: "body.emails"}},
			},
			flags: map[string][]string{"email": {"ada@example.com"}},
			want:  "output_policy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{tt.command}}}
			_, err := Run(context.Background(), connector, Request{Path: strings.Fields(tt.command.Path), Flags: tt.flags}, func(connectors.Record) error {
				t.Fatal("emit called for rejected operation direct-read command")
				return nil
			})
			if err == nil {
				t.Fatal("Run error = nil, want rejection")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Run error = %q, want %q", err.Error(), tt.want)
			}
		})
	}
}

func TestRunDirectReadRejectsUnsafeEndpointMetadata(t *testing.T) {
	tests := []struct {
		name     string
		endpoint connectors.CommandSurfaceEndpointRef
		want     string
	}{
		{
			name:     "mutation method",
			endpoint: connectors.CommandSurfaceEndpointRef{Method: "POST", Path: "/repos/{owner}/{repo}/contents/{path}"},
			want:     "GET",
		},
		{
			name:     "absolute url",
			endpoint: connectors.CommandSurfaceEndpointRef{Method: "GET", Path: "https://evil.example.test/repos/{owner}/{repo}"},
			want:     "absolute URL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &fakeConnector{surface: &connectors.CommandSurface{
				Commands: []connectors.CommandSurfaceCommand{
					{
						Path:         "repo read-file",
						Intent:       "direct_read",
						Availability: "implemented",
						APISurface:   []connectors.CommandSurfaceEndpointRef{tt.endpoint},
						OutputPolicy: "repository_contents_file_metadata",
						Flags: []connectors.CommandSurfaceFlag{
							{Name: "path", Type: "string", MapsTo: "path.path"},
						},
					},
				},
			}}

			_, err := Run(context.Background(), connector, Request{
				Path:  []string{"repo", "read-file"},
				Flags: map[string][]string{"path": []string{"README.md"}},
			}, func(connectors.Record) error {
				t.Fatal("emit called for rejected direct-read command")
				return nil
			})
			if err == nil {
				t.Fatal("Run error = nil, want rejection")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Run error = %q, want to contain %q", err.Error(), tt.want)
			}
		})
	}
}

func TestRunDirectReadRequiresOutputPolicy(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "repo read-file",
				Intent:       "direct_read",
				Availability: "implemented",
				APISurface: []connectors.CommandSurfaceEndpointRef{
					{Method: "GET", Path: "/repos/{owner}/{repo}/contents/{path}"},
				},
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "path", Type: "string", MapsTo: "path.path"},
				},
			},
		},
	}}

	_, err := Run(context.Background(), connector, Request{
		Path:  []string{"repo", "read-file"},
		Flags: map[string][]string{"path": []string{"README.md"}},
	}, func(connectors.Record) error {
		t.Fatal("emit called for rejected direct-read command")
		return nil
	})
	if err == nil {
		t.Fatal("Run error = nil, want output policy rejection")
	}
	if !strings.Contains(err.Error(), "output_policy") {
		t.Fatalf("Run error = %q, want output_policy", err.Error())
	}
}

func TestCLISurfaceOutputPolicyEnumMatchesRuntimePolicySets(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "engine", "schema", "cli_surface.schema.json"))
	if err != nil {
		t.Fatalf("read cli surface schema: %v", err)
	}
	type schemaNode struct {
		Properties map[string]schemaNode `json:"properties"`
		Items      *schemaNode           `json:"items"`
		Enum       []string              `json:"enum"`
	}
	var schema schemaNode
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("decode cli surface schema: %v", err)
	}
	commands, ok := schema.Properties["commands"]
	if !ok || commands.Items == nil {
		t.Fatal("cli surface schema has no commands.items")
	}
	outputPolicy, ok := commands.Items.Properties["output_policy"]
	if !ok || len(outputPolicy.Enum) == 0 {
		t.Fatal("cli surface schema has no output_policy enum")
	}

	want := make(map[string]struct{}, len(supportedDirectReadOutputPolicies)+len(supportedDirectWriteOutputPolicies)+2)
	for policy := range supportedDirectReadOutputPolicies {
		want[policy] = struct{}{}
	}
	for policy := range supportedDirectWriteOutputPolicies {
		want[policy] = struct{}{}
	}
	// This legacy compatibility value belongs to binary_download metadata, not
	// either JSON direct-read/write executor, and existing bundles still use it.
	want["binary_file_bounded"] = struct{}{}
	want["status"] = struct{}{}

	got := make(map[string]struct{}, len(outputPolicy.Enum))
	for _, policy := range outputPolicy.Enum {
		if _, duplicate := got[policy]; duplicate {
			t.Fatalf("cli surface output_policy enum repeats %q", policy)
		}
		got[policy] = struct{}{}
	}
	if missing, unexpected := outputPolicySetDifference(want, got), outputPolicySetDifference(got, want); len(missing) > 0 || len(unexpected) > 0 {
		t.Fatalf("cli surface output_policy enum diverges from runtime support: missing=%v unexpected=%v", missing, unexpected)
	}
}

func outputPolicySetDifference(left, right map[string]struct{}) []string {
	result := make([]string, 0)
	for policy := range left {
		if _, ok := right[policy]; !ok {
			result = append(result, policy)
		}
	}
	sort.Strings(result)
	return result
}

// TestCoerceFlagValueBoundsStringArrayItems pins the flag-level list bound. It
// is deliberately independent of the body schema's maxItems: the schema fires on
// the assembled body, this fires on the flag the user typed, so the error can
// name it.
func TestCoerceFlagValueBoundsStringArrayItems(t *testing.T) {
	flag := connectors.CommandSurfaceFlag{Name: "ids", Type: "string_array", MaxItems: 3, MinItems: 1}

	if _, err := coerceFlagValue(flag, []string{"a,b,c"}); err != nil {
		t.Fatalf("coerceFlagValue at the bound = %v, want accepted", err)
	}
	_, err := coerceFlagValue(flag, []string{"a,b,c,d"})
	if err == nil {
		t.Fatal("coerceFlagValue over the bound = nil, want rejection")
	}
	if !strings.Contains(err.Error(), "--ids") || !strings.Contains(err.Error(), "maximum of 3") {
		t.Fatalf("error = %q, want it to name the flag and the bound", err.Error())
	}
	if _, err := coerceFlagValue(flag, []string{","}); err == nil {
		t.Fatal("coerceFlagValue under the minimum = nil, want rejection")
	}
}

func TestValidateFlagMinimumIsOptIn(t *testing.T) {
	if err := validateFlagValue(connectors.CommandSurfaceFlag{Name: "page-number", Type: "integer"}, "0"); err != nil {
		t.Fatalf("validateFlagValue without minimum = %v, want unchanged acceptance", err)
	}

	minimum := 1.0
	flag := connectors.CommandSurfaceFlag{Name: "page-number", Type: "integer", Minimum: &minimum}
	if err := validateFlagValue(flag, "1"); err != nil {
		t.Fatalf("validateFlagValue at minimum = %v, want accepted", err)
	}
	err := validateFlagValue(flag, "0")
	var minimumErr *MinimumFlagError
	if !errors.As(err, &minimumErr) {
		t.Fatalf("validateFlagValue below minimum error = %T %v, want MinimumFlagError", err, err)
	}
	if minimumErr.Parameter != "page-number" || minimumErr.Minimum != 1 {
		t.Fatalf("MinimumFlagError = %+v, want page-number minimum 1", minimumErr)
	}
}

func TestStreamOverridesConfigMinimumIsOptIn(t *testing.T) {
	minimum := 1.0
	tests := []struct {
		name        string
		flag        connectors.CommandSurfaceFlag
		config      map[string]string
		flags       map[string][]string
		wantConfig  string
		wantMinimum bool
	}{
		{
			name:       "no declared minimum preserves config value",
			flag:       connectors.CommandSurfaceFlag{Name: "page-number", Type: "integer", MapsTo: "config.page_number"},
			config:     map[string]string{"page_number": "0"},
			wantConfig: "0",
		},
		{
			name:        "declared minimum rejects config value",
			flag:        connectors.CommandSurfaceFlag{Name: "page-number", Type: "integer", MapsTo: "config.page_number", Minimum: &minimum},
			config:      map[string]string{"page_number": "0"},
			wantMinimum: true,
		},
		{
			name:       "command flag wins over invalid config value",
			flag:       connectors.CommandSurfaceFlag{Name: "page-number", Type: "integer", MapsTo: "config.page_number", Minimum: &minimum},
			config:     map[string]string{"page_number": "0"},
			flags:      map[string][]string{"page-number": {"1"}},
			wantConfig: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := streamOverrides(connectors.CommandSurfaceCommand{
				Path:  "records list",
				Flags: []connectors.CommandSurfaceFlag{tt.flag},
			}, connectors.RuntimeConfig{Config: tt.config}, tt.flags)
			if tt.wantMinimum {
				var minimumErr *MinimumFlagError
				if !errors.As(err, &minimumErr) {
					t.Fatalf("streamOverrides error = %T %v, want MinimumFlagError", err, err)
				}
				if minimumErr.Parameter != "page-number" || minimumErr.Minimum != 1 {
					t.Fatalf("MinimumFlagError = %+v, want page-number minimum 1", minimumErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("streamOverrides error = %v, want nil", err)
			}
			if got := cfg.Config["page_number"]; got != tt.wantConfig {
				t.Fatalf("runtime config page_number = %q, want %q", got, tt.wantConfig)
			}
		})
	}
}

func TestValidateRequiredCommandFlagsPreservesStringArrayPresence(t *testing.T) {
	tests := []struct {
		name    string
		flag    connectors.CommandSurfaceFlag
		flags   map[string][]string
		wantErr string
	}{
		{
			name:    "missing key remains missing",
			flag:    connectors.CommandSurfaceFlag{Name: "items", Type: "string_array", Required: true},
			flags:   map[string][]string{},
			wantErr: "missing required flag --items",
		},
		{
			name:    "raw empty values remain missing",
			flag:    connectors.CommandSurfaceFlag{Name: "items", Type: "string_array", Required: true},
			flags:   map[string][]string{"items": {}},
			wantErr: "missing required flag --items",
		},
		{
			name:    "required scalar blank remains missing",
			flag:    connectors.CommandSurfaceFlag{Name: "title", Type: "string", Required: true},
			flags:   map[string][]string{"title": {""}},
			wantErr: "missing required flag --title",
		},
		{
			name:  "explicit blank zero minimum array is supplied",
			flag:  connectors.CommandSurfaceFlag{Name: "items", Type: "string_array", Required: true},
			flags: map[string][]string{"items": {""}},
		},
		{
			name:  "blank only csv zero minimum array is supplied",
			flag:  connectors.CommandSurfaceFlag{Name: "items", Type: "string_array", Required: true},
			flags: map[string][]string{"items": {", ,"}},
		},
		{
			name:    "minimum one still rejects materialized empty array",
			flag:    connectors.CommandSurfaceFlag{Name: "items", Type: "string_array", Required: true, MinItems: 1},
			flags:   map[string][]string{"items": {""}},
			wantErr: "below the minimum of 1",
		},
		{
			name:    "maximum remains authoritative",
			flag:    connectors.CommandSurfaceFlag{Name: "items", Type: "string_array", Required: true, MaxItems: 2},
			flags:   map[string][]string{"items": {"one,two,three"}},
			wantErr: "maximum of 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredCommandFlags(connectors.CommandSurfaceCommand{
				Path:  "widgets create",
				Flags: []connectors.CommandSurfaceFlag{tt.flag},
			}, tt.flags)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateRequiredCommandFlags error = %v, want accepted", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateRequiredCommandFlags error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestRunOperationDirectReadPreservesExplicitEmptyRequiredStringArray(t *testing.T) {
	tests := []struct {
		name     string
		flags    map[string][]string
		minItems int
		wantErr  string
	}{
		{
			name:  "explicit blank maps to literal empty body array",
			flags: map[string][]string{"items": {""}},
		},
		{
			name:    "omitted required flag does not invoke operation",
			flags:   map[string][]string{},
			wantErr: "missing required flag --items",
		},
		{
			name:    "raw empty values do not invoke operation",
			flags:   map[string][]string{"items": {}},
			wantErr: "missing required flag --items",
		},
		{
			name:     "minimum one does not invoke operation",
			flags:    map[string][]string{"items": {""}},
			minItems: 1,
			wantErr:  "below the minimum of 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
				Path:         "widgets search",
				Intent:       "direct_read",
				Availability: "implemented",
				Operation:    "acme.widgets.search",
				APISurface: []connectors.CommandSurfaceEndpointRef{
					{Method: http.MethodPost, Path: "/widgets/search"},
				},
				OutputPolicy: "json_redacted",
				Flags: []connectors.CommandSurfaceFlag{{
					Name: "items", Type: "string_array", Required: true, MinItems: tt.minItems, MapsTo: "body.items",
				}},
			}}}}

			result, err := Run(context.Background(), connector, Request{
				Path:  []string{"widgets", "search"},
				Flags: tt.flags,
			}, func(connectors.Record) error {
				t.Fatal("emit called for operation direct-read command")
				return nil
			})
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Run error = %v, want %q", err, tt.wantErr)
				}
				if connector.operationDirectReadReq.Operation != "" {
					t.Fatalf("operation direct read executed: %+v", connector.operationDirectReadReq)
				}
				return
			}
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if result.DirectRead == nil {
				t.Fatalf("result = %+v, want direct-read result", result)
			}
			items, ok := connector.operationDirectReadReq.Body["items"].([]string)
			if !ok || len(items) != 0 {
				t.Fatalf("operation body items = %#v, want literal empty []string", connector.operationDirectReadReq.Body["items"])
			}
			bodyJSON, err := json.Marshal(connector.operationDirectReadReq.Body)
			if err != nil {
				t.Fatalf("marshal operation body: %v", err)
			}
			if string(bodyJSON) != `{"items":[]}` {
				t.Fatalf("operation body JSON = %s, want literal empty array", bodyJSON)
			}
		})
	}
}

func TestBuildWriteCommandPreservesExplicitEmptyRequiredStringArray(t *testing.T) {
	tests := []struct {
		name     string
		flags    map[string][]string
		minItems int
		wantErr  string
	}{
		{
			name:  "explicit blank maps to literal empty planned record array",
			flags: map[string][]string{"items": {""}},
		},
		{
			name:    "omitted required flag does not plan or execute",
			flags:   map[string][]string{},
			wantErr: "missing required flag --items",
		},
		{
			name:    "raw empty values do not plan or execute",
			flags:   map[string][]string{"items": {}},
			wantErr: "missing required flag --items",
		},
		{
			name:     "minimum one does not plan or execute",
			flags:    map[string][]string{"items": {""}},
			minItems: 1,
			wantErr:  "below the minimum of 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &fakeConnector{
				surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
					Path:         "widgets create",
					Intent:       "reverse_etl",
					Availability: "implemented",
					Write:        "create_widgets",
					Flags: []connectors.CommandSurfaceFlag{{
						Name: "items", Type: "string_array", Required: true, MinItems: tt.minItems, MapsTo: "record.items",
					}},
				}}},
				manifest: connectors.Manifest{WriteActions: []connectors.WriteActionSpec{{
					Name: "create_widgets", Method: http.MethodPost, Path: "/widgets",
				}}},
			}

			command, err := BuildWriteCommand(context.Background(), connector, Request{
				Path:  []string{"widgets", "create"},
				Flags: tt.flags,
			})
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("BuildWriteCommand error = %v, want %q", err, tt.wantErr)
				}
				if connector.validateReq.Action != "" || connector.writeReq.Action != "" || len(connector.writeRecords) != 0 {
					t.Fatalf("rejected command reached write path: validate=%+v write=%+v records=%+v", connector.validateReq, connector.writeReq, connector.writeRecords)
				}
				return
			}
			if err != nil {
				t.Fatalf("BuildWriteCommand: %v", err)
			}
			if !command.ApprovalRequired {
				t.Fatalf("command = %+v, want approval-required reverse-ETL plan", command)
			}
			if connector.validateReq.Action != "create_widgets" || connector.writeReq.Action != "" || len(connector.writeRecords) != 0 {
				t.Fatalf("plan lifecycle = validate=%+v write=%+v records=%+v, want validation only", connector.validateReq, connector.writeReq, connector.writeRecords)
			}
			items, ok := command.Record["items"].([]string)
			if !ok || len(items) != 0 {
				t.Fatalf("planned record items = %#v, want literal empty []string", command.Record["items"])
			}
			recordJSON, err := json.Marshal(command.Record)
			if err != nil {
				t.Fatalf("marshal planned record: %v", err)
			}
			if string(recordJSON) != `{"items":[]}` {
				t.Fatalf("planned record JSON = %s, want literal empty array", recordJSON)
			}
		})
	}
}

// TestEveryImplementedCommandPassesRuntimePreflight is the structural guard
// against a command claiming availability "implemented" while failing at
// runtime.
//
// It asserts against the REAL runtime rather than a description of it: it walks
// every bundle registered from defs.FS and calls Preflight, the same entry
// point internal/cli calls before executing a connector command. That matters
// because the defect this test exists to prevent was precisely a validator that
// hand-copied the runtime's rules and drifted from them -- cmd/connectorgen
// exempted operation-backed direct reads from the api_surface check that
// commandrunner enforces, so 174 dead commands validated clean.
//
// Because it calls the runtime instead of restating it, every future executor
// kind is covered the day that executor lands, with no change here.
func TestEveryImplementedCommandPassesRuntimePreflight(t *testing.T) {
	registry := bundleregistry.New()

	type deadCommand struct {
		connector string
		command   string
		reason    string
	}
	var dead []deadCommand
	checked := 0

	for _, meta := range registry.List() {
		connector, ok := registry.Get(meta.Name)
		if !ok {
			t.Fatalf("registry lists %q but Get returned nothing", meta.Name)
		}
		provider, ok := connector.(connectors.CommandSurfaceProvider)
		if !ok || provider.CommandSurface() == nil {
			continue
		}
		for _, cmd := range provider.CommandSurface().Commands {
			if cmd.Availability != "implemented" {
				continue
			}
			checked++
			err := Preflight(connector, strings.Fields(cmd.Path))
			if err == nil {
				continue
			}
			reason := err.Error()
			var blocked *BlockedCommandError
			if errors.As(err, &blocked) {
				reason = blocked.Reason
			}
			dead = append(dead, deadCommand{connector: connector.Name(), command: cmd.Path, reason: reason})
		}
	}

	if checked == 0 {
		t.Fatal("no implemented commands were checked; the sweep is not reaching any bundle")
	}
	if len(dead) == 0 {
		return
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%d of %d commands marked \"implemented\" fail runtime Preflight:\n", len(dead), checked)
	for i, entry := range dead {
		if i == 25 {
			fmt.Fprintf(&b, "  ... and %d more\n", len(dead)-i)
			break
		}
		fmt.Fprintf(&b, "  %s %q: %s\n", entry.connector, entry.command, entry.reason)
	}
	b.WriteString("\nEither make the command executable or stop claiming it is implemented.")
	t.Fatal(b.String())
}

func binaryDownloadTestConnector() *fakeConnector {
	return &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "artifact download",
				Intent:       "binary_download",
				Availability: "implemented",
				Operation:    "github.artifact",
				APISurface: []connectors.CommandSurfaceEndpointRef{
					{Method: "GET", Path: "/repos/{owner}/{repo}/actions/artifacts/{artifact_id}/{archive_format}"},
				},
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "artifact-id", Type: "string", MapsTo: "path.artifact_id"},
					{Name: "archive-format", Type: "string", MapsTo: "path.archive_format"},
				},
				RedactFields: []string{"file_path"},
			},
		},
	}}
}

func TestRunBinaryDownloadCommandPassesDestinationThrough(t *testing.T) {
	connector := binaryDownloadTestConnector()

	result, err := Run(context.Background(), connector, Request{
		Path:     []string{"artifact", "download"},
		Flags:    map[string][]string{"artifact-id": {"42"}, "archive-format": {"zip"}},
		DestRoot: "out",
		FileName: "artifact.zip",
	}, func(connectors.Record) error {
		t.Fatal("emit must not be called for a binary download")
		return nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.BinaryDownload == nil {
		t.Fatal("Run result carries no binary download")
	}
	if got := connector.binaryDownloadReq.DestRoot; got != "out" {
		t.Fatalf("DestRoot = %q, want %q", got, "out")
	}
	if got := connector.binaryDownloadReq.FileName; got != "artifact.zip" {
		t.Fatalf("FileName = %q, want %q", got, "artifact.zip")
	}
	if got := connector.binaryDownloadReq.PathParams["artifact_id"]; got != "42" {
		t.Fatalf("PathParams[artifact_id] = %q, want %q", got, "42")
	}
	if len(connector.binaryDownloadReq.RedactFields) != 0 {
		t.Fatalf("binary download RedactFields = %#v, want empty", connector.binaryDownloadReq.RedactFields)
	}
	// The response never becomes records: a download is a file, not a stream.
	if result.Count != 0 || result.DirectRead != nil {
		t.Fatalf("binary download produced stream/direct-read output: %+v", result)
	}
}

// A destination is never inferred: without --dest-root the command is refused
// rather than defaulting to the working directory.
func TestRunBinaryDownloadRequiresDestinationRoot(t *testing.T) {
	connector := binaryDownloadTestConnector()

	_, err := Run(context.Background(), connector, Request{
		Path:  []string{"artifact", "download"},
		Flags: map[string][]string{"artifact-id": {"42"}, "archive-format": {"zip"}},
	}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Run error = nil, want a missing destination refusal")
	}
	if !strings.Contains(err.Error(), "dest-root") {
		t.Fatalf("Run error = %q, want it to name --dest-root", err.Error())
	}
}

func TestPreflightBlocksBinaryDownloadWithUnsafeMetadata(t *testing.T) {
	tests := []struct {
		name    string
		command connectors.CommandSurfaceCommand
		wantErr string
	}{
		{
			name: "no api_surface endpoint",
			command: connectors.CommandSurfaceCommand{
				Path: "artifact download", Intent: "binary_download",
				Availability: "implemented", Operation: "github.artifact",
			},
			wantErr: "exactly one api_surface endpoint",
		},
		{
			name: "non-GET endpoint",
			command: connectors.CommandSurfaceCommand{
				Path: "artifact download", Intent: "binary_download",
				Availability: "implemented", Operation: "github.artifact",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: "POST", Path: "/artifacts"}},
			},
			wantErr: "require GET",
		},
		{
			name: "absolute URL endpoint",
			command: connectors.CommandSurfaceCommand{
				Path: "artifact download", Intent: "binary_download",
				Availability: "implemented", Operation: "github.artifact",
				APISurface: []connectors.CommandSurfaceEndpointRef{{Method: "GET", Path: "https://evil.example.com/x"}},
			},
			wantErr: "absolute URL",
		},
		{
			name: "missing operation",
			command: connectors.CommandSurfaceCommand{
				Path: "artifact download", Intent: "binary_download",
				Availability: "implemented",
				APISurface:   []connectors.CommandSurfaceEndpointRef{{Method: "GET", Path: "/artifacts"}},
			},
			wantErr: "require operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := &fakeConnector{surface: &connectors.CommandSurface{
				Commands: []connectors.CommandSurfaceCommand{tt.command},
			}}
			err := Preflight(connector, []string{"artifact", "download"})
			if err == nil {
				t.Fatalf("Preflight error = nil, want %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Preflight error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// TestRunBinaryDownloadReachesOperationDeclaredCap crosses the ceiling that the
// CLI used to impose on every intent.
//
// It downloads a body LARGER than commandrunner.MaxOperationDirectReadBytes,
// which is exactly the case the previous verification missed: it used a
// 980-byte payload, so it never reached any ceiling and could not tell a
// working cap from a broken one. The CLI defaulted --max-bytes to the 16 MiB
// direct-read ceiling and passed it to every intent, so a github artifact
// declaring max_bytes 104857600 was silently truncated to 16 MiB and a user
// could not raise it back with the flag.
//
// The whole chain runs for real here -- commandrunner.Run, the engine executor,
// and the file that lands on disk -- because the defect lived in the plumbing
// between them, not in any one of them.
func TestRunBinaryDownloadReachesOperationDeclaredCap(t *testing.T) {
	const payloadSize = MaxOperationDirectReadBytes + (1 << 20)
	payload := bytes.Repeat([]byte("z"), payloadSize)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("request method = %s, want GET", r.Method)
		}
		if want := "/repos/acme/widgets/actions/artifacts/7/zip"; r.URL.Path != want {
			t.Errorf("request path = %s, want %s", r.URL.Path, want)
		}
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	githubConnector := func(t *testing.T) connectors.Connector {
		t.Helper()
		bundle, err := engine.Load(defs.FS, "github")
		if err != nil {
			t.Fatalf("load github bundle: %v", err)
		}
		bundle.HTTP.URL = server.URL
		return engine.New(bundle, nil)
	}
	config := connectors.RuntimeConfig{Config: map[string]string{
		"owner":         "acme",
		"repo":          "widgets",
		"base_url":      server.URL,
		"public_access": "true",
	}}
	flags := map[string][]string{"artifact-id": {"7"}, "archive-format": {"zip"}}

	t.Run("unset max bytes reaches the operation cap", func(t *testing.T) {
		dest := t.TempDir()
		result, err := Run(context.Background(), githubConnector(t), Request{
			Path:     []string{"artifact", "download"},
			Flags:    flags,
			Config:   config,
			DestRoot: dest,
			FileName: "artifact.zip",
			// MaxBytes stays zero: that is what the CLI passes when the user
			// does not type --max-bytes, and it must not lower the cap.
		}, func(connectors.Record) error {
			t.Fatal("emit must not be called for a binary download")
			return nil
		})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if result.BinaryDownload == nil {
			t.Fatal("Run result carries no binary download")
		}
		if got := result.BinaryDownload.Record["file_size_bytes"]; got != int64(payloadSize) {
			t.Fatalf("file_size_bytes = %v, want %d", got, payloadSize)
		}
		info, err := os.Stat(filepath.Join(dest, "artifact.zip"))
		if err != nil {
			t.Fatalf("stat downloaded file: %v", err)
		}
		if info.Size() != int64(payloadSize) {
			t.Fatalf("downloaded file size = %d, want %d", info.Size(), payloadSize)
		}
	})

	// The flag still only ever lowers: an explicit value below the body size
	// rejects rather than truncates.
	t.Run("explicit max bytes still lowers", func(t *testing.T) {
		dest := t.TempDir()
		_, err := Run(context.Background(), githubConnector(t), Request{
			Path:     []string{"artifact", "download"},
			Flags:    flags,
			Config:   config,
			MaxBytes: 1 << 10,
			DestRoot: dest,
			FileName: "artifact.zip",
		}, func(connectors.Record) error { return nil })
		if err == nil {
			t.Fatal("Run error = nil, want the explicit limit to reject the body")
		}
		if !strings.Contains(err.Error(), "too large") {
			t.Fatalf("Run error = %q, want it to name the size limit", err.Error())
		}
	})
}

// TestRunRejectsPageFlagsForIntentsThatCannotHonourThem is the anti-silence
// guard for the page flags themselves: only a direct_read pages, so every
// other intent must keep refusing --page instead of accepting and ignoring it.
func TestRunRejectsPageFlagsForIntentsThatCannotHonourThem(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "issue list",
				Intent:       "etl",
				Availability: "implemented",
				Stream:       "issues",
			},
		},
	}}

	for _, flag := range []string{"page", "page-cursor"} {
		t.Run(flag, func(t *testing.T) {
			_, err := Run(context.Background(), connector, Request{
				Path:  []string{"issue", "list"},
				Flags: map[string][]string{flag: {"3"}},
				Page:  3,
			}, func(connectors.Record) error { return nil })
			if err == nil {
				t.Fatalf("--%s on an etl command returned no error, want an unknown-flag refusal", flag)
			}
			if !strings.Contains(err.Error(), "unknown flag --"+flag) {
				t.Fatalf("error = %q, want an unknown flag --%s refusal", err.Error(), flag)
			}
		})
	}
}

// TestRunAcceptsPageFlagsForDirectRead is the other half: the intent that can
// honour them must not see them as command flags, and must receive them as the
// typed navigation inputs instead.
func TestRunAcceptsPageFlagsForDirectRead(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "repo read-file",
				Intent:       "direct_read",
				Availability: "implemented",
				APISurface: []connectors.CommandSurfaceEndpointRef{
					{Method: "GET", Path: "/repos/{owner}/{repo}/contents/{path}"},
				},
				OutputPolicy: "repository_contents_file_metadata",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "path", Type: "string", MapsTo: "path.path"},
				},
			},
		},
	}}

	if _, err := Run(context.Background(), connector, Request{
		Path:  []string{"repo", "read-file"},
		Flags: map[string][]string{"path": {"README.md"}, "page": {"2"}},
		Page:  2,
		Config: connectors.RuntimeConfig{Config: map[string]string{
			"owner": "octo",
			"repo":  "hello",
		}},
	}, func(connectors.Record) error { return nil }); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if connector.directReadReq.Page != 2 {
		t.Fatalf("direct read page = %d, want 2", connector.directReadReq.Page)
	}
}

// TestRunRefusesPageNavigationAReaderCannotHonour is the general guard behind
// the amazon-sqs finding. Page/PageCursor are handed to whatever direct-read
// implementation a connector supplies, and nothing in the type system makes it
// honour them; one that does not returns a zero page, so the caller got page
// one at status 200 with the request silently discarded.
func TestRunRefusesPageNavigationAReaderCannotHonour(t *testing.T) {
	surface := &connectors.CommandSurface{
		Commands: []connectors.CommandSurfaceCommand{
			{
				Path:         "repo read-file",
				Intent:       "direct_read",
				Availability: "implemented",
				APISurface: []connectors.CommandSurfaceEndpointRef{
					{Method: "GET", Path: "/repos/{owner}/{repo}/contents/{path}"},
				},
				OutputPolicy: "repository_contents_file_metadata",
				Flags: []connectors.CommandSurfaceFlag{
					{Name: "path", Type: "string", MapsTo: "path.path"},
				},
			},
		},
	}
	config := connectors.RuntimeConfig{Config: map[string]string{"owner": "octo", "repo": "hello"}}

	for _, tc := range []struct {
		name string
		req  Request
	}{
		{name: "page", req: Request{Flags: map[string][]string{"path": {"README.md"}, "page": {"3"}}, Page: 3}},
		{name: "page-cursor", req: Request{Flags: map[string][]string{"path": {"README.md"}, "page-cursor": {"abc"}}, PageCursor: "abc"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			connector := &fakeConnector{surface: surface, ignoresPageNavigation: true}
			req := tc.req
			req.Path = []string{"repo", "read-file"}
			req.Config = config
			_, err := Run(context.Background(), connector, req, func(connectors.Record) error { return nil })
			if err == nil {
				t.Fatal("a reader that ignores page navigation returned no error, want a refusal rather than page one")
			}
			if !strings.Contains(err.Error(), "reported no page context") {
				t.Fatalf("error = %q, want it to name the missing page context", err.Error())
			}
		})
	}
}
