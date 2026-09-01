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
		userScoped bool
	}
	wants := []commandWant{
		{
			path:    "users list",
			stream:  "users",
			apiPath: "/v2/users",
		},
		{
			path:       "meetings list",
			stream:     "meetings",
			apiPath:    "/v2/users/{userId}/meetings",
			userScoped: true,
		},
		{
			path:       "webinars list",
			stream:     "webinars",
			apiPath:    "/v2/users/{userId}/webinars",
			userScoped: true,
		},
	}
	if got := len(surface.Commands); got != len(wants) {
		t.Fatalf("Zoom cli_surface commands = %d, want exactly %d in Wave 1", got, len(wants))
	}

	commands := make(map[string]struct {
		stream     string
		intent     string
		available  string
		apiMethod  string
		apiPath    string
		userIDFlag bool
	}, len(surface.Commands))
	for _, command := range surface.Commands {
		entry := struct {
			stream     string
			intent     string
			available  string
			apiMethod  string
			apiPath    string
			userIDFlag bool
		}{
			stream:    command.Stream,
			intent:    command.Intent,
			available: command.Availability,
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
		if command.userIDFlag != want.userScoped {
			t.Errorf("command %q optional --user-id config override = %t, want %t", want.path, command.userIDFlag, want.userScoped)
		}
		if err := commandrunner.Preflight(connector, strings.Fields(want.path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want nil", want.path, err)
		}
	}

	for path := range commands {
		if _, ok := map[string]struct{}{
			"users list":    {},
			"meetings list": {},
			"webinars list": {},
		}[path]; !ok {
			t.Errorf("Wave 1 must not promote additional Zoom operations; found %q", path)
		}
	}

}

// TestCoveredStreamCommandsExecuteWithLocalServer runs each command through
// commandrunner against a local HTTP server. It proves
// that command-specific config overrides reach the stream and that --limit
// prevents the users cursor from fetching a second response page once enough
// records have been emitted.
func TestCoveredStreamCommandsExecuteWithLocalServer(t *testing.T) {
	responses := map[string]json.RawMessage{
		"users-page-1":    json.RawMessage(`{"users":[{"id":"user_one","email":"one@example.test","first_name":"One","last_name":"User","updated_at":"2026-01-01T00:00:00Z"},{"id":"user_two","email":"two@example.test","first_name":"Two","last_name":"User","updated_at":"2026-01-02T00:00:00Z"}],"next_page_token":"page_two"}`),
		"users-page-2":    json.RawMessage(`{"users":[{"id":"user_three","email":"three@example.test","first_name":"Three","last_name":"User","updated_at":"2026-01-03T00:00:00Z"}],"next_page_token":""}`),
		"meetings-page-1": json.RawMessage(`{"meetings":[{"id":"meeting_one","uuid":"meeting_uuid_one","topic":"Meeting One","updated_at":"2026-01-01T00:00:00Z"},{"id":"meeting_two","uuid":"meeting_uuid_two","topic":"Meeting Two","start_time":"2026-01-02T00:00:00Z"}],"next_page_token":""}`),
		"webinars-page-1": json.RawMessage(`{"webinars":[{"id":"webinar_one","uuid":"webinar_uuid_one","topic":"Webinar One","updated_at":"2026-01-01T00:00:00Z"},{"id":"webinar_two","uuid":"webinar_uuid_two","topic":"Webinar Two","start_time":"2026-01-02T00:00:00Z"}],"next_page_token":""}`),
	}

	var requestsMu sync.Mutex
	requests := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		responseKey := ""
		switch request.URL.Path {
		case "/users":
			if request.URL.Query().Get("next_page_token") == "page_two" {
				responseKey = "users-page-2"
			} else {
				responseKey = "users-page-1"
			}
		case "/users/local-user/meetings":
			responseKey = "meetings-page-1"
		case "/users/local-user/webinars":
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
			flags:      map[string][]string{"user-id": {"local-user"}},
			wantStream: "meetings",
		},
		{
			name:       "webinars with user override",
			path:       []string{"webinars", "list"},
			flags:      map[string][]string{"user-id": {"local-user"}},
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
				t.Fatalf("Run(%q) stream=%q count=%d emitted=%d, want stream=%q and two records", strings.Join(test.path, " "), result.Stream, result.Count, len(emitted), test.wantStream)
			}
		})
	}

	requestsMu.Lock()
	defer requestsMu.Unlock()
	for _, path := range []string{"/users", "/users/local-user/meetings", "/users/local-user/webinars"} {
		if got := requests[path]; got != 1 {
			t.Errorf("requests for %s = %d, want 1", path, got)
		}
	}
}
