// Package zoom holds connector-local reachability evidence for the Zoom
// declarative bundle. The bundle itself is JSON; this test keeps its provider
// inventory and executable command surface from drifting apart.
package zoom

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

const zoomBundleName = "zoom"

func loadZoomBundle(t *testing.T) engine.Bundle {
	t.Helper()
	bundle, err := engine.Load(os.DirFS(".."), zoomBundleName)
	if err != nil {
		t.Fatalf("load %s bundle: %v", zoomBundleName, err)
	}
	return bundle
}

// TestProviderInventoryLedgerIsComplete pins the provider-owned inventory
// rebuilt from Zoom's published OpenAPI 3.1.1 reference corpus. Any endpoint
// that is not already stream-backed must remain explicitly blocked or a cited
// justified exclusion; dropping a row is never an acceptable way to make the
// inventory appear complete.
func TestProviderInventoryLedgerIsComplete(t *testing.T) {
	bundle := loadZoomBundle(t)
	if bundle.Surface == nil {
		t.Fatal("api_surface.json did not load")
	}
	if bundle.Surface.ReviewedAt != "2026-08-05" {
		t.Fatalf("provider inventory reviewed_at = %q, want 2026-08-05", bundle.Surface.ReviewedAt)
	}
	if bundle.Surface.OperationLedgerVersion != 1 {
		t.Fatalf("provider inventory operation_ledger_version = %d, want 1", bundle.Surface.OperationLedgerVersion)
	}
	if !strings.Contains(bundle.Surface.API, "OpenAPI 3.1.1") || !strings.Contains(bundle.Surface.API, "2026-08-03T14-58-19-06-00") {
		t.Fatalf("provider inventory source = %q, want OpenAPI version and static-build provenance", bundle.Surface.API)
	}

	wantMethods := map[string]int{
		http.MethodDelete: 319,
		http.MethodGet:    881,
		http.MethodPatch:  269,
		http.MethodPost:   392,
		http.MethodPut:    52,
	}
	gotMethods := make(map[string]int, len(wantMethods))
	seen := make(map[string]struct{}, len(bundle.Surface.Endpoints))
	unclassified := make([]string, 0)
	covered, implementableNow, providerRestricted, deprecated := 0, 0, 0, 0

	for _, endpoint := range bundle.Surface.Endpoints {
		method := strings.ToUpper(strings.TrimSpace(endpoint.Method))
		path := strings.TrimSpace(endpoint.Path)
		key := method + " " + path
		if method == "" || path == "" {
			t.Errorf("provider inventory has an endpoint without method/path: %+v", endpoint)
			continue
		}
		if _, duplicate := seen[key]; duplicate {
			t.Errorf("provider inventory repeats %s", key)
			continue
		}
		seen[key] = struct{}{}
		gotMethods[method]++

		switch {
		case endpoint.CoveredBy != nil:
			if endpoint.Operation != nil || endpoint.Excluded != nil {
				t.Errorf("executable %s carries another disposition", key)
			}
			covering := endpoint.CoveredBy
			if covering.Stream == "" && covering.Write == "" && covering.DirectRead == "" && len(covering.DirectReads) == 0 {
				t.Errorf("executable %s is not bound to a stream, write, or direct_read command", key)
			}
			covered++
		case endpoint.Operation != nil:
			operation := endpoint.Operation
			if operation.Status != "blocked" || !operation.BlockedByDefault || strings.TrimSpace(operation.Reason) == "" {
				t.Errorf("blocked %s has incomplete disposition: %+v", key, operation)
			}
			if !strings.HasPrefix(operation.SourceURL, "https://developers.zoom.us/docs/api/") {
				t.Errorf("blocked %s source_url = %q, want Zoom provider citation", key, operation.SourceURL)
			}
			switch {
			case strings.Contains(operation.Notes, "classification=implementable_now"):
				implementableNow++
			case strings.Contains(operation.Notes, "classification=provider_restriction"):
				providerRestricted++
			case operation.Model == "deprecated" && strings.Contains(operation.Notes, "classification=justified_excluded"):
				deprecated++
			default:
				unclassified = append(unclassified, key+" ("+operation.Model+")")
			}
		default:
			unclassified = append(unclassified, key)
		}
	}

	if got := len(bundle.Surface.Endpoints); got != 1913 {
		t.Errorf("provider inventory endpoints = %d, want 1913", got)
	}
	if got := len(seen); got != 1913 {
		t.Errorf("provider inventory unique method/path rows = %d, want 1913", got)
	}
	for method, want := range wantMethods {
		if got := gotMethods[method]; got != want {
			t.Errorf("provider inventory %s rows = %d, want %d", method, got, want)
		}
	}
	if got := covered; got != 6 {
		t.Errorf("executable stream-backed rows = %d, want 6", got)
	}
	if got := implementableNow; got != 1836 {
		t.Errorf("operations awaiting Zoom-local contracts = %d, want 1836", got)
	}
	if got := providerRestricted; got != 17 {
		t.Errorf("provider-restricted operations = %d, want 17", got)
	}
	if got := deprecated; got != 54 {
		t.Errorf("provider-deprecated justified exclusions = %d, want 54", got)
	}
	if len(unclassified) > 0 {
		t.Errorf("provider inventory has %d rows without a recognized disposition: %s", len(unclassified), strings.Join(unclassified, ", "))
	}
}

// TestCoveredStreamsHaveReachableCommands proves the executable subset through
// the real command runner. Before cli_surface.json exists this deliberately
// fails: engine.synthesizeCommandSurface returns nil and `pm zoom <command>`
// cannot resolve the existing streams.
func TestCoveredStreamsHaveReachableCommands(t *testing.T) {
	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	surface := connector.CommandSurface()
	if surface == nil {
		t.Fatal("Zoom has no cli_surface.json; its stream-backed provider operations are unreachable as pm zoom commands")
	}

	type commandWant struct {
		path       string
		stream     string
		apiPath    string
		sourceURL  string
		userScoped bool
	}
	wants := []commandWant{
		{
			path:      "users list",
			stream:    "users",
			apiPath:   "/v2/users",
			sourceURL: "https://developers.zoom.us/docs/api/users.md",
		},
		{
			path:       "meetings list",
			stream:     "meetings",
			apiPath:    "/v2/users/{userId}/meetings",
			sourceURL:  "https://developers.zoom.us/docs/api/meetings.md",
			userScoped: true,
		},
		{
			path:       "webinars list",
			stream:     "webinars",
			apiPath:    "/v2/users/{userId}/webinars",
			sourceURL:  "https://developers.zoom.us/docs/api/meetings.md",
			userScoped: true,
		},
	}

	type operationWant struct {
		path      string
		operation string
		apiMethod string
		apiPath   string
		pathFlag  string
	}
	operationWants := []operationWant{
		{
			path:      "qss meeting-participants list",
			operation: "zoom.list_meeting_participants_qos_summary",
			apiMethod: http.MethodGet,
			apiPath:   "/v2/metrics/meetings/{meetingId}/participants/qos_summary",
			pathFlag:  "meeting-id",
		},
		{
			path:      "qss webinar-participants list",
			operation: "zoom.list_webinar_participants_qos_summary",
			apiMethod: http.MethodGet,
			apiPath:   "/v2/metrics/webinars/{webinarId}/participants/qos_summary",
			pathFlag:  "webinar-id",
		},
		{
			path:      "qss session-users list",
			operation: "zoom.list_session_users_qos_summary",
			apiMethod: http.MethodGet,
			apiPath:   "/v2/videosdk/sessions/{sessionId}/users/qos_summary",
			pathFlag:  "session-id",
		},
	}

	wantCommandCount := len(wants) + len(operationWants)
	if got := len(surface.Commands); got != wantCommandCount {
		t.Fatalf("Zoom cli_surface commands = %d, want exactly %d (Wave 1 streams + Wave 2 qss operations)", got, wantCommandCount)
	}

	commands := make(map[string]struct {
		stream       string
		intent       string
		available    string
		operation    string
		apiMethod    string
		apiPath      string
		sourceURL    string
		outputPolicy string
		userIDFlag   bool
		requiredFlag string
	}, len(surface.Commands))
	for _, command := range surface.Commands {
		entry := struct {
			stream       string
			intent       string
			available    string
			operation    string
			apiMethod    string
			apiPath      string
			sourceURL    string
			outputPolicy string
			userIDFlag   bool
			requiredFlag string
		}{
			stream:       command.Stream,
			intent:       command.Intent,
			available:    command.Availability,
			operation:    command.Operation,
			sourceURL:    command.SourceURL,
			outputPolicy: command.OutputPolicy,
		}
		if len(command.APISurface) == 1 {
			entry.apiMethod = command.APISurface[0].Method
			entry.apiPath = command.APISurface[0].Path
		}
		for _, flag := range command.Flags {
			if flag.Name == "user-id" && flag.MapsTo == "config.user_id" && !flag.Required {
				entry.userIDFlag = true
			}
			if strings.HasPrefix(flag.MapsTo, "path.") && flag.Required {
				entry.requiredFlag = flag.Name
			}
		}
		commands[command.Path] = entry
	}

	for _, want := range wants {
		command, ok := commands[want.path]
		if !ok {
			t.Errorf("missing reachable command %q", want.path)
			continue
		}
		if command.intent != "etl" || command.available != "implemented" || command.stream != want.stream {
			t.Errorf("command %q = intent=%q availability=%q stream=%q, want implemented ETL stream %q", want.path, command.intent, command.available, command.stream, want.stream)
		}
		if command.apiMethod != http.MethodGet || command.apiPath != want.apiPath {
			t.Errorf("command %q api_surface = %s %s, want GET %s", want.path, command.apiMethod, command.apiPath, want.apiPath)
		}
		if command.sourceURL != want.sourceURL {
			t.Errorf("command %q source_url = %q, want %q", want.path, command.sourceURL, want.sourceURL)
		}
		if command.userIDFlag != want.userScoped {
			t.Errorf("command %q optional --user-id config override = %t, want %t", want.path, command.userIDFlag, want.userScoped)
		}
		if err := commandrunner.Preflight(connector, strings.Fields(want.path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want nil", want.path, err)
		}
	}

	for _, want := range operationWants {
		command, ok := commands[want.path]
		if !ok {
			t.Errorf("missing reachable command %q", want.path)
			continue
		}
		if command.intent != "direct_read" || command.available != "implemented" || command.operation != want.operation {
			t.Errorf("command %q = intent=%q availability=%q operation=%q, want implemented direct_read operation %q", want.path, command.intent, command.available, command.operation, want.operation)
		}
		if command.apiMethod != want.apiMethod || command.apiPath != want.apiPath {
			t.Errorf("command %q api_surface = %s %s, want %s %s", want.path, command.apiMethod, command.apiPath, want.apiMethod, want.apiPath)
		}
		if command.outputPolicy != "json_redacted" {
			t.Errorf("command %q output_policy = %q, want json_redacted", want.path, command.outputPolicy)
		}
		if command.requiredFlag != want.pathFlag {
			t.Errorf("command %q required path flag = %q, want %q", want.path, command.requiredFlag, want.pathFlag)
		}
		if err := commandrunner.Preflight(connector, strings.Fields(want.path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want nil", want.path, err)
		}
	}

	allowed := map[string]struct{}{
		"users list":                    {},
		"meetings list":                 {},
		"webinars list":                 {},
		"qss meeting-participants list": {},
		"qss webinar-participants list": {},
		"qss session-users list":        {},
	}
	for path := range commands {
		if _, ok := allowed[path]; !ok {
			t.Errorf("Wave 1/Wave 2 must not promote additional Zoom operations; found %q", path)
		}
	}
}

// TestCoveredStreamCommandsExecuteWithFixtures runs each Wave 1 command
// through commandrunner against Zoom's committed sanitized fixtures. It proves
// that command-specific config overrides reach the stream and that --limit
// prevents the users cursor from fetching a second fixture page once enough
// records have been emitted.
func TestCoveredStreamCommandsExecuteWithFixtures(t *testing.T) {
	responses := map[string]json.RawMessage{
		"users-page-1":    zoomFixtureResponseBody(t, "users", "page_1.json"),
		"users-page-2":    zoomFixtureResponseBody(t, "users", "page_2.json"),
		"meetings-page-1": zoomFixtureResponseBody(t, "meetings", "page_1.json"),
		"webinars-page-1": zoomFixtureResponseBody(t, "webinars", "page_1.json"),
	}

	var requestsMu sync.Mutex
	requests := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		responseKey := ""
		switch request.URL.Path {
		case "/users":
			if request.URL.Query().Get("next_page_token") == "fixture_token_2" {
				responseKey = "users-page-2"
			} else {
				responseKey = "users-page-1"
			}
		case "/users/fixture-user/meetings":
			responseKey = "meetings-page-1"
		case "/users/fixture-user/webinars":
			responseKey = "webinars-page-1"
		default:
			http.NotFound(w, request)
			return
		}
		if got := request.URL.Query().Get("page_size"); got != "100" {
			t.Errorf("%s page_size = %q, want 100", request.URL.Path, got)
		}
		requestsMu.Lock()
		requests[request.URL.Path]++
		requestsMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(responses[responseKey])
	}))
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	config := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  server.URL,
			"page_size": "100",
			"user_id":   "credential-config-user",
		},
		Secrets: map[string]string{"access_token": "synthetic-test-token"},
	}
	tests := []struct {
		name       string
		path       []string
		flags      map[string][]string
		wantStream string
	}{
		{
			name:       "users",
			path:       []string{"users", "list"},
			wantStream: "users",
		},
		{
			name:       "meetings with user override",
			path:       []string{"meetings", "list"},
			flags:      map[string][]string{"user-id": {"fixture-user"}},
			wantStream: "meetings",
		},
		{
			name:       "webinars with user override",
			path:       []string{"webinars", "list"},
			flags:      map[string][]string{"user-id": {"fixture-user"}},
			wantStream: "webinars",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			emitted := make([]connectors.Record, 0, 2)
			result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
				Path:   test.path,
				Flags:  test.flags,
				Config: config,
				Limit:  2,
			}, func(record connectors.Record) error {
				emitted = append(emitted, record)
				return nil
			})
			if err != nil {
				t.Fatalf("Run(%q) = %v", strings.Join(test.path, " "), err)
			}
			if result.Stream != test.wantStream || result.Count != 2 || len(emitted) != 2 {
				t.Fatalf("Run(%q) stream=%q count=%d emitted=%d, want stream=%q and two fixture records", strings.Join(test.path, " "), result.Stream, result.Count, len(emitted), test.wantStream)
			}
		})
	}

	requestsMu.Lock()
	defer requestsMu.Unlock()
	for _, path := range []string{"/users", "/users/fixture-user/meetings", "/users/fixture-user/webinars"} {
		if got := requests[path]; got != 1 {
			t.Errorf("fixture requests for %s = %d, want 1", path, got)
		}
	}
}

// TestQSSOperationDirectReadCommandsExecuteWithFixtures runs each Wave 2 QSS
// direct_read command through commandrunner against Zoom's committed
// sanitized fixtures, proving the required path parameter reaches the
// resolved request path. The json_redacted output policy redacts any field
// whose name contains "token" (engine/direct_read.go's shouldRedactJSONField)
// -- including the QSS response's own pagination cursor field
// next_page_token, which is not itself a secret -- so the assertion below
// expects that field replaced with next_page_token_redacted: true rather
// than the raw fixture value; every other field must survive unchanged.
func TestQSSOperationDirectReadCommandsExecuteWithFixtures(t *testing.T) {
	tests := []struct {
		name        string
		path        []string
		flags       map[string][]string
		requestPath string
		fixture     string
	}{
		{
			name:        "meeting participants",
			path:        []string{"qss", "meeting-participants", "list"},
			flags:       map[string][]string{"meeting-id": {"fixture-meeting"}},
			requestPath: "/v2/metrics/meetings/fixture-meeting/participants/qos_summary",
			fixture:     "list_meeting_participants_qos_summary.json",
		},
		{
			name:        "webinar participants",
			path:        []string{"qss", "webinar-participants", "list"},
			flags:       map[string][]string{"webinar-id": {"fixture-webinar"}},
			requestPath: "/v2/metrics/webinars/fixture-webinar/participants/qos_summary",
			fixture:     "list_webinar_participants_qos_summary.json",
		},
		{
			name:        "session users",
			path:        []string{"qss", "session-users", "list"},
			flags:       map[string][]string{"session-id": {"fixture-session"}},
			requestPath: "/v2/videosdk/sessions/fixture-session/users/qos_summary",
			fixture:     "list_session_users_qos_summary.json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wantStatus, wantBody := zoomDirectReadFixture(t, test.fixture)

			var requestsMu sync.Mutex
			requests := make(map[string]int)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				requestsMu.Lock()
				requests[request.URL.Path]++
				requestsMu.Unlock()
				if request.URL.Path != test.requestPath {
					http.NotFound(w, request)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(wantStatus)
				_, _ = w.Write(wantBody)
			}))
			defer server.Close()

			bundle := loadZoomBundle(t)
			connector := engine.New(bundle, engine.HooksFor(bundle.Name))
			config := connectors.RuntimeConfig{
				// Mirror the production base_url shape (https://api.zoom.us/v2):
				// the engine strips the base URL's own path prefix from the
				// declared /v2/... operation path before issuing the request.
				Config:  map[string]string{"base_url": server.URL + "/v2"},
				Secrets: map[string]string{"access_token": "synthetic-test-token"},
			}

			result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
				Path:   test.path,
				Flags:  test.flags,
				Config: config,
			}, func(connectors.Record) error {
				t.Fatal("emit called for a direct_read command")
				return nil
			})
			if err != nil {
				t.Fatalf("Run(%q) = %v", strings.Join(test.path, " "), err)
			}
			if result.DirectRead == nil {
				t.Fatalf("Run(%q) DirectRead = nil, want a result", strings.Join(test.path, " "))
			}
			if result.DirectRead.Status != wantStatus {
				t.Errorf("Run(%q) status = %d, want %d", strings.Join(test.path, " "), result.DirectRead.Status, wantStatus)
			}

			var wantDecoded map[string]any
			if err := json.Unmarshal(wantBody, &wantDecoded); err != nil {
				t.Fatalf("decode fixture body: %v", err)
			}
			delete(wantDecoded, "next_page_token")
			wantDecoded["next_page_token_redacted"] = true

			gotBody, err := json.Marshal(result.DirectRead.Body)
			if err != nil {
				t.Fatalf("marshal result body: %v", err)
			}
			var gotDecoded map[string]any
			if err := json.Unmarshal(gotBody, &gotDecoded); err != nil {
				t.Fatalf("decode result body: %v", err)
			}
			if !reflect.DeepEqual(gotDecoded, map[string]any(wantDecoded)) {
				t.Errorf("Run(%q) body = %s, want %v", strings.Join(test.path, " "), gotBody, wantDecoded)
			}

			requestsMu.Lock()
			defer requestsMu.Unlock()
			if got := requests[test.requestPath]; got != 1 {
				t.Errorf("fixture requests for %s = %d, want 1", test.requestPath, got)
			}
		})
	}
}

func zoomDirectReadFixture(t *testing.T, file string) (int, json.RawMessage) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("fixtures", "direct_reads", file))
	if err != nil {
		t.Fatalf("read direct_reads fixture %s: %v", file, err)
	}
	var fixture struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode direct_reads fixture %s: %v", file, err)
	}
	if len(fixture.Body) == 0 {
		t.Fatalf("direct_reads fixture %s has no body", file)
	}
	return fixture.Status, fixture.Body
}

func zoomFixtureResponseBody(t *testing.T, stream, file string) json.RawMessage {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("fixtures", "streams", stream, file))
	if err != nil {
		t.Fatalf("read %s fixture %s: %v", stream, file, err)
	}
	var fixture struct {
		Response struct {
			Body json.RawMessage `json:"body"`
		} `json:"response"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode %s fixture %s: %v", stream, file, err)
	}
	if len(fixture.Response.Body) == 0 {
		t.Fatalf("%s fixture %s has no response body", stream, file)
	}
	return fixture.Response.Body
}
