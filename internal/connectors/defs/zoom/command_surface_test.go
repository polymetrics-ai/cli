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
	streamBacked, directReadCovered, implementableNow, providerRestricted, deprecated := 0, 0, 0, 0, 0

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
			switch {
			case endpoint.CoveredBy.Stream != "":
				streamBacked++
			case endpoint.CoveredBy.DirectRead != "" || len(endpoint.CoveredBy.DirectReads) > 0:
				directReadCovered++
			default:
				t.Errorf("executable %s is not bound to a stream or direct read", key)
			}
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
	if got := streamBacked; got != 3 {
		t.Errorf("executable stream-backed rows = %d, want 3", got)
	}
	if got := directReadCovered; got != 70 {
		t.Errorf("executable direct-read rows = %d, want 70", got)
	}
	if got := implementableNow; got != 1769 {
		t.Errorf("operations awaiting Zoom-local contracts = %d, want 1839", got)
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

// TestCoveredStreamsHaveReachableCommands preserves the three existing ETL
// commands through the real command runner while later parity waves add other
// declarative intents to the same Zoom surface.
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
	if got := len(surface.Commands); got < len(wants) {
		t.Fatalf("Zoom cli_surface commands = %d, want at least the %d existing ETL commands", got, len(wants))
	}

	commands := make(map[string]struct {
		stream     string
		intent     string
		available  string
		apiMethod  string
		apiPath    string
		sourceURL  string
		userIDFlag bool
	}, len(surface.Commands))
	for _, command := range surface.Commands {
		entry := struct {
			stream     string
			intent     string
			available  string
			apiMethod  string
			apiPath    string
			sourceURL  string
			userIDFlag bool
		}{
			stream:    command.Stream,
			intent:    command.Intent,
			available: command.Availability,
			sourceURL: command.SourceURL,
		}
		if len(command.APISurface) == 1 {
			entry.apiMethod = command.APISurface[0].Method
			entry.apiPath = command.APISurface[0].Path
		}
		for _, flag := range command.Flags {
			if flag.Name == "user-id" && flag.MapsTo == "config.user_id" && !flag.Required {
				entry.userIDFlag = true
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

	for path, command := range commands {
		if command.intent != "etl" {
			continue
		}
		if _, ok := map[string]struct{}{
			"users list":    {},
			"meetings list": {},
			"webinars list": {},
		}[path]; !ok {
			t.Errorf("unexpected Zoom ETL command %q", path)
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

// TestReviewedDirectReadSalvageCohort keeps Wave 2's reviewed declarations,
// command paths, and sanitized fixture corpus together. Preflight is the real
// runner boundary: a generated surface is not enough when a command cannot
// dispatch its named operation.
func TestReviewedDirectReadSalvageCohort(t *testing.T) {
	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))

	operations := make(map[string]struct{}, 70)
	for _, operation := range bundle.Operations {
		if operation.Kind == "rest_read" {
			operations[operation.ID] = struct{}{}
		}
	}
	if got := len(operations); got != 70 {
		t.Fatalf("reviewed rest_read operations = %d, want 70", got)
	}

	surface := connector.CommandSurface()
	if surface == nil {
		t.Fatal("Zoom direct-read cohort has no command surface")
	}
	commands := 0
	for _, command := range surface.Commands {
		if command.Intent != "direct_read" {
			continue
		}
		commands++
		if _, ok := operations[command.Operation]; !ok {
			t.Errorf("direct-read command %q references unknown operation %q", command.Path, command.Operation)
			continue
		}
		if command.Availability != "implemented" {
			t.Errorf("direct-read command %q availability = %q, want implemented", command.Path, command.Availability)
		}
		if err := commandrunner.Preflight(connector, strings.Fields(command.Path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want nil", command.Path, err)
		}
	}
	if commands != 70 {
		t.Fatalf("reviewed direct-read commands = %d, want 70", commands)
	}

	entries, err := os.ReadDir(filepath.Join("fixtures", "direct_reads"))
	if err != nil {
		t.Fatalf("read direct-read fixtures: %v", err)
	}
	fixtures := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join("fixtures", "direct_reads", entry.Name()))
		if err != nil {
			t.Fatalf("read direct-read fixture %q: %v", entry.Name(), err)
		}
		if !json.Valid(raw) {
			t.Errorf("direct-read fixture %q is not JSON", entry.Name())
		}
		fixtures++
	}
	if fixtures != 52 {
		t.Fatalf("sanitized direct-read fixtures = %d, want 52", fixtures)
	}
}

// TestReviewedDirectReadFixturesExecute exercises every salvaged fixture
// through the operation executor against a loopback Zoom provider double. The
// companion cohort test proves each CLI command reaches this executor through
// its real preflight path; this test proves the declared method, path binding,
// bearer transport, response bound, and redacting output policy run together.
func TestReviewedDirectReadFixturesExecute(t *testing.T) {
	type fixture struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	}
	type action struct {
		name       string
		operation  engine.OperationSpec
		pathParams map[string]string
		requestURL string
		response   fixture
	}

	bundle := loadZoomBundle(t)
	operations := make(map[string]engine.OperationSpec, len(bundle.Operations))
	for _, operation := range bundle.Operations {
		if operation.Kind == "rest_read" {
			operations[operation.ID] = operation
		}
	}

	entries, err := os.ReadDir(filepath.Join("fixtures", "direct_reads"))
	if err != nil {
		t.Fatalf("read direct-read fixtures: %v", err)
	}
	fixtureOperationIDs := map[string]string{
		// These fixtures preserve their original item-read names even though
		// the provider contracts are collection reads.
		"get_task_assignees":     "zoom.list_task_assignees",
		"get_task_collaborators": "zoom.list_task_collaborators",
	}
	actions := make([]action, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join("fixtures", "direct_reads", entry.Name()))
		if err != nil {
			t.Fatalf("read direct-read fixture %q: %v", entry.Name(), err)
		}
		var response fixture
		if err := json.Unmarshal(raw, &response); err != nil {
			t.Fatalf("decode direct-read fixture %q: %v", entry.Name(), err)
		}
		fixtureID := strings.TrimSuffix(entry.Name(), ".json")
		operationID := fixtureOperationIDs[fixtureID]
		if operationID == "" {
			operationID = "zoom." + fixtureID
		}
		operation, ok := operations[operationID]
		if !ok {
			t.Fatalf("direct-read fixture %q does not map to a reviewed operation", entry.Name())
		}
		params, requestURL := zoomFixturePathParams(operation.REST.Path)
		if response.Status == 0 {
			response.Status = http.StatusOK
		}
		actions = append(actions, action{
			name:       entry.Name(),
			operation:  operation,
			pathParams: params,
			requestURL: requestURL,
			response:   response,
		})
	}
	if len(actions) != 52 {
		t.Fatalf("fixture actions = %d, want 52", len(actions))

	}

	byRequest := make(map[string]action, len(actions))
	for _, action := range actions {
		key := action.operation.REST.Method + " " + action.requestURL
		if _, duplicate := byRequest[key]; duplicate {
			t.Fatalf("direct-read fixture repeats request %s", key)
		}
		byRequest[key] = action
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action, ok := byRequest[request.Method+" "+request.URL.Path]
		if !ok {
			http.Error(w, "undeclared fixture request", http.StatusNotFound)
			t.Errorf("direct-read fixture received undeclared request %s %s", request.Method, request.URL.Path)
			return
		}
		if got := request.Header.Get("Authorization"); got != "Bearer fixture-direct-read-access-token" {
			http.Error(w, "missing bearer credential", http.StatusUnauthorized)
			t.Errorf("direct-read fixture %q authorization = %q, want fixture bearer credential", action.name, got)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(action.response.Status)
		_, _ = w.Write(action.response.Body)
	}))
	defer server.Close()

	config := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": server.URL},
		Secrets: map[string]string{"access_token": "fixture-direct-read-access-token"},
	}
	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			result, err := engine.OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
				Operation:    action.operation.ID,
				Config:       config,
				PathParams:   action.pathParams,
				MaxBytes:     action.operation.REST.MaxBytes,
				OutputPolicy: action.operation.OutputPolicy,
			}, engine.HooksFor(bundle.Name))
			if err != nil {
				t.Fatalf("OperationDirectRead(%q): %v", action.operation.ID, err)
			}
			if result.Status != action.response.Status {
				t.Fatalf("OperationDirectRead(%q) status = %d, want %d", action.operation.ID, result.Status, action.response.Status)
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("direct-read fixture requests = %d, want %d", requests, len(actions))
	}
}

func zoomFixturePathParams(path string) (map[string]string, string) {
	params := make(map[string]string)
	resolved := path
	for {
		start := strings.Index(resolved, "{")
		if start < 0 {
			return params, resolved
		}
		end := strings.Index(resolved[start:], "}")
		if end < 0 {
			return params, resolved
		}
		end += start
		name := resolved[start+1 : end]
		value := "fixture-" + strings.ToLower(name)
		params[name] = value
		resolved = strings.Replace(resolved, resolved[start:end+1], value, 1)
	}
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
