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

// wantModuleOperationCommandCount is the running total of implemented
// direct_read/direct_write operation commands across all landed modules
// (Wave 2+, one Zoom provider module at a time; see issue #3915 and
// .planning/phases/cli-zoom-parity-wave2-qss-r1/). Bump this FIRST when
// starting a new module's red/green cycle -- that bump is what makes
// TestCoveredStreamsHaveReachableCommands fail red before the module's
// operations.json/cli_surface.json entries exist.
//
// Landed modules: qss (3), ai-companion (1), my-notes (2), healthcare reads (2),
// quality-management reads (5), Cobrowse SDK reads (4).
const wantModuleOperationCommandCount = 17

// wantModuleWriteCommandCount is the running total of implemented reverse_etl
// write commands across the landed provider modules. Bump it first for each
// module so the command runner proves the typed action is promotable before a
// production bundle change can turn its ledger row into covered_by.write.
//
// Landed modules: healthcare (1, update_clinical_note), quality-management
// (1, create_quality_management_interaction).
const wantModuleWriteCommandCount = 2

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
	if got := covered; got != 22 {
		t.Errorf("executable rows = %d, want 22", got)
	}
	if got := implementableNow; got != 1820 {
		t.Errorf("operations awaiting Zoom-local contracts = %d, want 1820", got)
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

	for _, want := range wants {
		var found *connectors.CommandSurfaceCommand
		for i := range surface.Commands {
			if surface.Commands[i].Path == want.path {
				found = &surface.Commands[i]
				break
			}
		}
		if found == nil {
			t.Errorf("missing reachable command %q", want.path)
			continue
		}
		command := *found
		var apiMethod, apiPath string
		if len(command.APISurface) == 1 {
			apiMethod, apiPath = command.APISurface[0].Method, command.APISurface[0].Path
		}
		userIDFlag := false
		for _, flag := range command.Flags {
			if flag.Name == "user-id" && flag.MapsTo == "config.user_id" && !flag.Required {
				userIDFlag = true
			}
		}
		if command.Intent != "etl" || command.Availability != "implemented" || command.Stream != want.stream {
			t.Errorf("command %q = intent=%q availability=%q stream=%q, want implemented ETL stream %q", want.path, command.Intent, command.Availability, command.Stream, want.stream)
		}
		if apiMethod != http.MethodGet || apiPath != want.apiPath {
			t.Errorf("command %q api_surface = %s %s, want GET %s", want.path, apiMethod, apiPath, want.apiPath)
		}
		if command.SourceURL != want.sourceURL {
			t.Errorf("command %q source_url = %q, want %q", want.path, command.SourceURL, want.sourceURL)
		}
		if userIDFlag != want.userScoped {
			t.Errorf("command %q optional --user-id config override = %t, want %t", want.path, userIDFlag, want.userScoped)
		}
		if err := commandrunner.Preflight(connector, strings.Fields(want.path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want nil", want.path, err)
		}
	}

	verifyOperationCommands(t, bundle, connector, surface, wantModuleOperationCommandCount)
	verifyWriteCommands(t, bundle, connector, surface, wantModuleWriteCommandCount)
}

// verifyOperationCommands generically verifies every reachable operation
// command (direct_read today; direct_write once zoom declares writes.json)
// against its own operations.json entry, rather than hand-duplicating each
// command's expected method/path/flags in Go -- that duplication does not
// scale to zoom's 1,913-operation surface. It still proves real per-command
// correctness: every operation command must resolve to a declared
// operations.json entry whose rest.method/path match cli_surface.json's
// api_surface exactly, every required path-mapped flag must name a real
// {placeholder} in that path, output_policy must be a supported direct-read
// policy, and the real command-runner Preflight must pass -- the same bar
// TestEveryImplementedCommandPassesRuntimePreflight enforces repo-wide.
func verifyOperationCommands(t *testing.T, bundle engine.Bundle, connector connectors.Connector, surface *connectors.CommandSurface, wantCount int) {
	t.Helper()
	ops := make(map[string]engine.OperationSpec, len(bundle.Operations))
	for _, op := range bundle.Operations {
		ops[op.ID] = op
	}

	got := 0
	for _, command := range surface.Commands {
		if command.Intent != "direct_read" {
			continue
		}
		got++
		if command.Availability != "implemented" {
			t.Errorf("operation command %q availability = %q, want implemented", command.Path, command.Availability)
			continue
		}
		op, ok := ops[command.Operation]
		if !ok {
			t.Errorf("operation command %q references operation %q, not found in operations.json", command.Path, command.Operation)
			continue
		}
		if op.REST == nil {
			t.Errorf("operation %q has no rest spec", op.ID)
			continue
		}
		if len(command.APISurface) != 1 || command.APISurface[0].Method != op.REST.Method || command.APISurface[0].Path != op.REST.Path {
			t.Errorf("operation command %q api_surface = %+v, want exactly [%s %s] matching operations.json", command.Path, command.APISurface, op.REST.Method, op.REST.Path)
		}
		if command.OutputPolicy != op.OutputPolicy {
			t.Errorf("operation command %q output_policy = %q, want %q (from operations.json)", command.Path, command.OutputPolicy, op.OutputPolicy)
		}
		if !supportedDirectReadOutputPolicies[command.OutputPolicy] {
			t.Errorf("operation command %q output_policy = %q, not a supported direct-read policy", command.Path, command.OutputPolicy)
		}
		for _, flag := range command.Flags {
			if !strings.HasPrefix(flag.MapsTo, "path.") || !flag.Required {
				continue
			}
			placeholder := "{" + strings.TrimPrefix(flag.MapsTo, "path.") + "}"
			if !strings.Contains(op.REST.Path, placeholder) {
				t.Errorf("operation command %q required flag --%s maps_to %q, but %q is not present in path %q", command.Path, flag.Name, flag.MapsTo, placeholder, op.REST.Path)
			}
		}
		if err := commandrunner.Preflight(connector, strings.Fields(command.Path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want nil", command.Path, err)
		}
	}
	if got != wantCount {
		t.Errorf("reachable direct_read operation commands = %d, want %d", got, wantCount)
	}
}

// verifyWriteCommands verifies every implemented reverse_etl route through the
// same commandrunner preflight the CLI uses. Unlike direct reads, write paths
// deliberately stay inside the plan -> preview -> approval -> execute flow;
// preflight proves the typed declaration is safe to promote without issuing a
// mutation.
func verifyWriteCommands(t *testing.T, bundle engine.Bundle, connector connectors.Connector, surface *connectors.CommandSurface, wantCount int) {
	t.Helper()
	writes := make(map[string]struct{}, len(bundle.Writes))
	for _, write := range bundle.Writes {
		writes[write.Name] = struct{}{}
	}

	got := 0
	for _, command := range surface.Commands {
		if command.Intent != "reverse_etl" {
			continue
		}
		got++
		if command.Availability != "implemented" {
			t.Errorf("write command %q availability = %q, want implemented", command.Path, command.Availability)
			continue
		}
		if command.Write == "" {
			t.Errorf("write command %q has no writes.json action", command.Path)
			continue
		}
		if _, ok := writes[command.Write]; !ok {
			t.Errorf("write command %q references action %q, not found in writes.json", command.Path, command.Write)
		}
		for _, flag := range command.Flags {
			if !strings.HasPrefix(flag.MapsTo, "record.") {
				t.Errorf("write command %q flag --%s maps_to %q, want a typed record.* field", command.Path, flag.Name, flag.MapsTo)
			}
		}
		if err := commandrunner.Preflight(connector, strings.Fields(command.Path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want nil", command.Path, err)
		}
	}
	if got != wantCount {
		t.Errorf("reachable reverse_etl write commands = %d, want %d", got, wantCount)
	}
}

// supportedDirectReadOutputPolicies mirrors commandrunner's closed policy set
// (internal/connectors/commandrunner/runner.go) so this test fails loudly if
// a zoom command ever declares an output_policy the runtime cannot execute --
// duplicated here deliberately (small, closed, rarely-changing set) rather
// than exporting the unexported runner map across package boundaries.
var supportedDirectReadOutputPolicies = map[string]bool{
	"repository_contents_file_metadata": true,
	"repository_contents_directory":     true,
	"json_redacted":                     true,
	"clinical_json_redacted":            true,
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

// TestHealthcareClinicalNoteCommandsExecuteWithFixtures proves that the two
// sensitive direct reads reach their fixed Zoom paths with only documented
// request inputs, and that the single PATCH action remains approval-gated at
// the command layer while accepting a provider-successful 204 in the isolated
// write executor.
func TestHealthcareClinicalNoteCommandsExecuteWithFixtures(t *testing.T) {
	readTests := []struct {
		name        string
		path        []string
		flags       map[string][]string
		requestPath string
		query       map[string]string
		fixture     string
		pageToken   bool
	}{
		{
			name:        "list",
			path:        []string{"healthcare", "clinical-notes", "list"},
			flags:       map[string][]string{"note-owner-user-id": {"fixture-owner"}, "meeting-id": {"fixture-meeting"}},
			requestPath: "/v2/clinical_notes/notes",
			query:       map[string]string{"note_owner_user_id": "fixture-owner", "meeting_id": "fixture-meeting"},
			fixture:     "list_clinical_notes.json",
			pageToken:   true,
		},
		{
			name:        "get",
			path:        []string{"healthcare", "clinical-notes", "get"},
			flags:       map[string][]string{"note-id": {"fixture-note"}},
			requestPath: "/v2/clinical_notes/notes/fixture-note",
			fixture:     "get_clinical_note.json",
		},
	}

	for _, test := range readTests {
		t.Run(test.name, func(t *testing.T) {
			wantStatus, wantBody := zoomDirectReadFixture(t, test.fixture)
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				requests++
				if request.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", request.Method)
				}
				if request.URL.Path != test.requestPath {
					http.NotFound(w, request)
					return
				}
				for key, want := range test.query {
					if got := request.URL.Query().Get(key); got != want {
						t.Errorf("query %s = %q, want %q", key, got, want)
					}
				}
				for _, responseOnly := range []string{"from", "to", "page_size", "next_page_token"} {
					if got := request.URL.Query().Get(responseOnly); got != "" {
						t.Errorf("response-only field %s was sent as query %q", responseOnly, got)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(wantStatus)
				_, _ = w.Write(wantBody)
			}))
			defer server.Close()

			bundle := loadZoomBundle(t)
			connector := engine.New(bundle, engine.HooksFor(bundle.Name))
			result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
				Path:   test.path,
				Flags:  test.flags,
				Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL + "/v2"}, Secrets: map[string]string{"access_token": "synthetic-test-token"}},
			}, func(connectors.Record) error {
				t.Fatal("emit called for a direct_read command")
				return nil
			})
			if err != nil {
				t.Fatalf("Run(%q) = %v", strings.Join(test.path, " "), err)
			}
			if result.DirectRead == nil || result.DirectRead.Status != wantStatus {
				t.Fatalf("Run(%q) direct read = %#v, want status %d", strings.Join(test.path, " "), result.DirectRead, wantStatus)
			}
			assertClinicalResponseRedacted(t, result.DirectRead.Body, test.pageToken)
			if requests != 1 {
				t.Errorf("requests = %d, want 1", requests)
			}
		})
	}

	t.Run("update plans before mutation and accepts no content", func(t *testing.T) {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			requests++
			if request.Method != http.MethodPatch {
				t.Errorf("method = %s, want PATCH", request.Method)
			}
			if request.URL.Path != "/v2/clinical_notes/notes/fixture-note" {
				t.Errorf("path = %s, want clinical-note path", request.URL.Path)
			}
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Errorf("decode PATCH body: %v", err)
			}
			if got, want := body["is_note_completed"], true; got != want {
				t.Errorf("is_note_completed = %#v, want %t", got, want)
			}
			if _, present := body["note_id"]; present {
				t.Error("note_id was sent in the PATCH body as well as the declared path")
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		bundle := loadZoomBundle(t)
		connector := engine.New(bundle, engine.HooksFor(bundle.Name))
		config := connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL + "/v2"}, Secrets: map[string]string{"access_token": "synthetic-test-token"}}
		planned, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{
			Path:    []string{"healthcare", "clinical-notes", "update"},
			Flags:   map[string][]string{"note-id": {"fixture-note"}, "is-note-completed": {"true"}},
			Config:  config,
			Preview: true,
		})
		if err != nil {
			t.Fatalf("BuildWriteCommand = %v", err)
		}
		if planned.Write != "update_clinical_note" || planned.Preview == nil {
			t.Fatalf("planned Healthcare write = %#v, want typed action with preview", planned)
		}
		if requests != 0 {
			t.Fatalf("preview reached network; requests = %d, want 0", requests)
		}

		result, err := connector.Write(context.Background(), connectors.WriteRequest{Action: "update_clinical_note", Config: config}, []connectors.Record{{"note_id": "fixture-note", "is_note_completed": true}})
		if err != nil {
			t.Fatalf("Write(update_clinical_note) = %v", err)
		}
		if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
			t.Fatalf("Write result = %#v, want one successful 204 action", result)
		}
		if requests != 1 {
			t.Fatalf("PATCH requests = %d, want 1", requests)
		}
	})
}

// TestQualityManagementCommandsExecuteWithFixtures pins the live provider's
// five GET routes to fixed paths with no invented response-pagination inputs,
// then proves the one typed POST request builds the entire closed nested body
// and accepts Zoom's documented 201 success status.
func TestQualityManagementCommandsExecuteWithFixtures(t *testing.T) {
	readTests := []struct {
		name        string
		path        []string
		flags       map[string][]string
		requestPath string
		fixture     string
		raw         []string
		fields      []string
	}{
		{
			name:        "list automated evaluations",
			path:        []string{"quality-management", "automated-evaluations", "list"},
			requestPath: "/v2/qm/automated_evaluations",
			fixture:     "list_quality_management_automated_evaluations.json",
			raw:         []string{"fixture-account", "fixture-agent-name", "fixture-agent-user", "fixture-page-token"},
			fields:      []string{"account_id", "display_name", "user_id", "next_page_token"},
		},
		{
			name:        "list evaluations",
			path:        []string{"quality-management", "evaluations", "list"},
			requestPath: "/v2/qm/evaluation",
			fixture:     "list_quality_management_evaluations.json",
			raw:         []string{"fixture-account", "fixture-creator", "fixture-creator-name", "fixture-creator-user", "fixture-page-token"},
			fields:      []string{"account_id", "creator_id", "display_name", "user_id", "next_page_token"},
		},
		{
			name:        "get evaluation",
			path:        []string{"quality-management", "evaluations", "get"},
			flags:       map[string][]string{"evaluation-id": {"fixture-evaluation"}},
			requestPath: "/v2/qm/evaluation/fixture-evaluation",
			fixture:     "get_quality_management_evaluation.json",
			raw:         []string{"fixture-account", "fixture-evaluator-name", "fixture-evaluator-user"},
			fields:      []string{"account_id", "display_name", "user_id"},
		},
		{
			name:        "list interactions",
			path:        []string{"quality-management", "interactions", "list"},
			requestPath: "/v2/qm/interactions",
			fixture:     "list_quality_management_interactions.json",
			raw:         []string{"fixture-account", "fixture-agent@example.invalid", "fixture-consumer", "+15550000001", "+15550000002", "fixture-page-token"},
			fields:      []string{"account_id", "agent_email", "consumer_name", "from", "to", "next_page_token"},
		},
		{
			name:        "get interaction",
			path:        []string{"quality-management", "interactions", "get"},
			flags:       map[string][]string{"interaction-id": {"fixture-interaction"}},
			requestPath: "/v2/qm/interactions/fixture-interaction",
			fixture:     "get_quality_management_interaction.json",
			raw:         []string{"fixture-account", "fixture-agent@example.invalid", "fixture-consumer", "+15550000001", "+15550000002"},
			fields:      []string{"account_id", "agent_email", "consumer_name", "from", "to"},
		},
	}

	for _, test := range readTests {
		t.Run(test.name, func(t *testing.T) {
			wantStatus, wantBody := zoomDirectReadFixture(t, test.fixture)
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				requests++
				if request.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", request.Method)
				}
				if request.URL.Path != test.requestPath {
					http.NotFound(w, request)
					return
				}
				for _, responseOnly := range []string{"from", "to", "page_size", "next_page_token", "limit"} {
					if got := request.URL.Query().Get(responseOnly); got != "" {
						t.Errorf("response-only field %s was sent as query %q", responseOnly, got)
					}
				}
				if got := request.URL.Query().Encode(); got != "" {
					t.Errorf("Quality Management GET query = %q, want no undocumented parameters", got)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(wantStatus)
				_, _ = w.Write(wantBody)
			}))
			defer server.Close()

			bundle := loadZoomBundle(t)
			connector := engine.New(bundle, engine.HooksFor(bundle.Name))
			result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
				Path:   test.path,
				Flags:  test.flags,
				Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL + "/v2"}, Secrets: map[string]string{"access_token": "synthetic-test-token"}},
			}, func(connectors.Record) error {
				t.Fatal("emit called for a direct_read command")
				return nil
			})
			if err != nil {
				t.Fatalf("Run(%q) = %v", strings.Join(test.path, " "), err)
			}
			if result.DirectRead == nil || result.DirectRead.Status != wantStatus {
				t.Fatalf("Run(%q) direct read = %#v, want status %d", strings.Join(test.path, " "), result.DirectRead, wantStatus)
			}
			assertQualityManagementResponseRedacted(t, result.DirectRead.Body, test.raw, test.fields)
			if requests != 1 {
				t.Errorf("requests = %d, want 1", requests)
			}
		})
	}

	t.Run("create interaction plans before mutation and accepts created", func(t *testing.T) {
		wantStatus, wantResponse := zoomWriteFixture(t, "create_quality_management_interaction.json")
		wantRequestBody := map[string]any{
			"download_url": "https://files.example.invalid/fixture-interaction.mp3",
			"direction":    "inbound",
			"disposition":  "fixture-disposition",
			"interaction_info": map[string]any{
				"channel_type":  "voice",
				"agent_email":   "fixture-agent@example.invalid",
				"agent_id":      "fixture-agent-id",
				"consumer_name": "fixture-consumer",
				"from":          "+15550000001",
				"to":            "+15550000002",
			},
			"primary_language": "en-US",
			"queue_id":         "fixture-queue",
			"start_time":       "2026-08-08T09:00:00Z",
		}
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			requests++
			if request.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", request.Method)
			}
			if request.URL.Path != "/v2/qm/interactions" {
				t.Errorf("path = %s, want Quality Management interaction path", request.URL.Path)
			}
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Errorf("decode POST body: %v", err)
			}
			if !reflect.DeepEqual(body, wantRequestBody) {
				t.Errorf("POST body = %#v, want %#v", body, wantRequestBody)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(wantStatus)
			_, _ = w.Write(wantResponse)
		}))
		defer server.Close()

		bundle := loadZoomBundle(t)
		connector := engine.New(bundle, engine.HooksFor(bundle.Name))
		config := connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL + "/v2"}, Secrets: map[string]string{"access_token": "synthetic-test-token"}}
		flags := map[string][]string{
			"download-url":              {"https://files.example.invalid/fixture-interaction.mp3"},
			"direction":                 {"inbound"},
			"disposition":               {"fixture-disposition"},
			"interaction-channel-type":  {"voice"},
			"interaction-agent-email":   {"fixture-agent@example.invalid"},
			"interaction-agent-id":      {"fixture-agent-id"},
			"interaction-consumer-name": {"fixture-consumer"},
			"interaction-from":          {"+15550000001"},
			"interaction-to":            {"+15550000002"},
			"primary-language":          {"en-US"},
			"queue-id":                  {"fixture-queue"},
			"start-time":                {"2026-08-08T09:00:00Z"},
		}
		planned, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{
			Path:    []string{"quality-management", "interactions", "create"},
			Flags:   flags,
			Config:  config,
			Preview: true,
		})
		if err != nil {
			t.Fatalf("BuildWriteCommand = %v", err)
		}
		if planned.Write != "create_quality_management_interaction" || planned.Preview == nil {
			t.Fatalf("planned Quality Management write = %#v, want typed action with preview", planned)
		}
		if requests != 0 {
			t.Fatalf("preview reached network; requests = %d, want 0", requests)
		}

		result, err := connector.Write(context.Background(), connectors.WriteRequest{Action: "create_quality_management_interaction", Config: config}, []connectors.Record{{
			"download_url": "https://files.example.invalid/fixture-interaction.mp3",
			"direction":    "inbound",
			"disposition":  "fixture-disposition",
			"interaction_info": map[string]any{
				"channel_type":  "voice",
				"agent_email":   "fixture-agent@example.invalid",
				"agent_id":      "fixture-agent-id",
				"consumer_name": "fixture-consumer",
				"from":          "+15550000001",
				"to":            "+15550000002",
			},
			"primary_language": "en-US",
			"queue_id":         "fixture-queue",
			"start_time":       "2026-08-08T09:00:00Z",
		}})
		if err != nil {
			t.Fatalf("Write(create_quality_management_interaction) = %v", err)
		}
		if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
			t.Fatalf("Write result = %#v, want one successful 201 action", result)
		}
		if requests != 1 {
			t.Fatalf("POST requests = %d, want 1", requests)
		}
	})
}

func assertQualityManagementResponseRedacted(t *testing.T, body any, raw, fields []string) {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal Quality Management response: %v", err)
	}
	got := string(encoded)
	for _, value := range raw {
		if strings.Contains(got, value) {
			t.Errorf("Quality Management response exposed %q: %s", value, got)
		}
	}
	for _, field := range fields {
		if !strings.Contains(got, "\""+field+"_redacted\":true") {
			t.Errorf("Quality Management response is missing %s_redacted marker: %s", field, got)
		}
	}
}

// TestCobrowseSDKCommandsExecuteWithFixtures fixes the provider's four Cobrowse
// SDK GET paths, explicitly tests only the two prose-declared report dates,
// and proves session pins, IDs, names, connection IDs, IP addresses, and
// response pagination tokens are redacted before output.
func TestCobrowseSDKCommandsExecuteWithFixtures(t *testing.T) {
	readTests := []struct {
		name        string
		path        []string
		flags       map[string][]string
		requestPath string
		query       map[string]string
		fixture     string
		raw         []string
		fields      []string
	}{
		{
			name:        "list live sessions",
			path:        []string{"cobrowse-sdk", "live-sessions", "list"},
			flags:       map[string][]string{"from": {"2026-08-01"}, "to": {"2026-08-08"}},
			requestPath: "/v2/cobrowsesdk/live_sessions",
			query:       map[string]string{"from": "2026-08-01", "to": "2026-08-08"},
			fixture:     "list_cobrowse_live_sessions.json",
			raw:         []string{"fixture-live-session", "fixture-session-pin", "fixture-live-user", "fixture-live-user-name", "fixture-page-token"},
			fields:      []string{"session_id", "session_pin", "user_id", "user_name", "next_page_token"},
		},
		{
			name:        "list past sessions",
			path:        []string{"cobrowse-sdk", "past-sessions", "list"},
			flags:       map[string][]string{"from": {"2026-08-01"}, "to": {"2026-08-08"}},
			requestPath: "/v2/cobrowsesdk/past_sessions",
			query:       map[string]string{"from": "2026-08-01", "to": "2026-08-08"},
			fixture:     "list_cobrowse_past_sessions.json",
			raw:         []string{"fixture-past-session", "fixture-session-pin", "fixture-past-user", "fixture-past-user-name", "fixture-page-token"},
			fields:      []string{"session_id", "session_pin", "user_id", "user_name", "next_page_token"},
		},
		{
			name:        "get session",
			path:        []string{"cobrowse-sdk", "sessions", "get"},
			flags:       map[string][]string{"session-id": {"fixture-session"}},
			requestPath: "/v2/cobrowsesdk/sessions/fixture-session",
			fixture:     "get_cobrowse_session.json",
			raw:         []string{"fixture-session", "fixture-session-pin"},
			fields:      []string{"session_id", "session_pin"},
		},
		{
			name:        "list session users",
			path:        []string{"cobrowse-sdk", "sessions", "users", "list"},
			flags:       map[string][]string{"session-id": {"fixture-session"}},
			requestPath: "/v2/cobrowsesdk/sessions/fixture-session/users",
			fixture:     "list_cobrowse_session_users.json",
			raw:         []string{"fixture-user-connection", "fixture-session-user", "fixture-session-user-name", "192.0.2.1", "fixture-page-token"},
			fields:      []string{"user_connection_id", "user_id", "user_name", "ip_address", "next_page_token"},
		},
	}

	for _, test := range readTests {
		t.Run(test.name, func(t *testing.T) {
			wantStatus, wantBody := zoomDirectReadFixture(t, test.fixture)
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				requests++
				if request.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", request.Method)
				}
				if request.URL.Path != test.requestPath {
					http.NotFound(w, request)
					return
				}
				for key, want := range test.query {
					if got := request.URL.Query().Get(key); got != want {
						t.Errorf("query %s = %q, want %q", key, got, want)
					}
				}
				if got, want := len(request.URL.Query()), len(test.query); got != want {
					t.Errorf("query field count = %d, want %d (%v)", got, want, test.query)
				}
				for _, responseOnly := range []string{"page", "per_page", "limit", "page_size", "next_page_token"} {
					if got := request.URL.Query().Get(responseOnly); got != "" {
						t.Errorf("response-only field %s was sent as query %q", responseOnly, got)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(wantStatus)
				_, _ = w.Write(wantBody)
			}))
			defer server.Close()

			bundle := loadZoomBundle(t)
			connector := engine.New(bundle, engine.HooksFor(bundle.Name))
			result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
				Path:   test.path,
				Flags:  test.flags,
				Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL + "/v2"}, Secrets: map[string]string{"access_token": "synthetic-test-token"}},
			}, func(connectors.Record) error {
				t.Fatal("emit called for a direct_read command")
				return nil
			})
			if err != nil {
				t.Fatalf("Run(%q) = %v", strings.Join(test.path, " "), err)
			}
			if result.DirectRead == nil || result.DirectRead.Status != wantStatus {
				t.Fatalf("Run(%q) direct read = %#v, want status %d", strings.Join(test.path, " "), result.DirectRead, wantStatus)
			}
			assertCobrowseSDKResponseRedacted(t, result.DirectRead.Body, test.raw, test.fields)
			if requests != 1 {
				t.Errorf("requests = %d, want 1", requests)
			}
		})
	}
}

func assertCobrowseSDKResponseRedacted(t *testing.T, body any, raw, fields []string) {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal Cobrowse SDK response: %v", err)
	}
	got := string(encoded)
	for _, value := range raw {
		if strings.Contains(got, value) {
			t.Errorf("Cobrowse SDK response exposed %q: %s", value, got)
		}
	}
	for _, field := range fields {
		if !strings.Contains(got, "\""+field+"_redacted\":true") {
			t.Errorf("Cobrowse SDK response is missing %s_redacted marker: %s", field, got)
		}
	}
}

func assertClinicalResponseRedacted(t *testing.T, body any, expectPageToken bool) {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal clinical response: %v", err)
	}
	got := string(encoded)
	for _, raw := range []string{"fixture-clinical-content", "fixture-patient", "fixture-provider", "fixture-appointment", "fixture-owner", "fixture-editor", "fixture-page-token"} {
		if strings.Contains(got, raw) {
			t.Errorf("clinical response exposed %q: %s", raw, got)
		}
	}
	fields := []string{"note_content", "patient_id", "provider_id", "appointment_id", "note_owner_user_id", "note_last_modified_user_id"}
	if expectPageToken {
		fields = append(fields, "next_page_token")
	}
	for _, field := range fields {
		if !strings.Contains(got, "\""+field+"_redacted\":true") {
			t.Errorf("clinical response is missing %s_redacted marker: %s", field, got)
		}
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

func zoomWriteFixture(t *testing.T, file string) (int, json.RawMessage) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("fixtures", "writes", file))
	if err != nil {
		t.Fatalf("read writes fixture %s: %v", file, err)
	}
	var fixture struct {
		Response struct {
			Status int             `json:"status"`
			Body   json.RawMessage `json:"body"`
		} `json:"response"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode writes fixture %s: %v", file, err)
	}
	if fixture.Response.Status == 0 || len(fixture.Response.Body) == 0 {
		t.Fatalf("writes fixture %s requires response.status and response.body", file)
	}
	return fixture.Response.Status, fixture.Response.Body
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
