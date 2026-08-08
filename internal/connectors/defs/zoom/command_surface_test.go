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

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

const zoomBundleName = "zoom"

// wantModuleOperationCommandCount is the running total of implemented
// direct_read operation commands across all landed modules
// (Wave 2+, one Zoom provider module at a time; see issue #3915 and
// .planning/phases/cli-zoom-parity-wave2-qss-r1/). Bump this FIRST when
// starting a new module's red/green cycle -- that bump is what makes
// TestCoveredStreamsHaveReachableCommands fail red before the module's
// operations.json/cli_surface.json entries exist.
//
// Landed modules: qss (3), ai-companion (1), my-notes (2), healthcare reads (2),
// quality-management reads (5), Cobrowse SDK reads (4), and SCIM2 reads (4).
// Chatbot and SCIM2 mutations are tracked by wantModuleDirectWriteCommandCount.
const wantModuleOperationCommandCount = 21

// wantModuleDirectWriteCommandCount is the running total of implemented
// operations.json rest_write commands. It is distinct from writes.json
// reverse-ETL actions because direct writes are executable only through the
// typed plan lifecycle.
const wantModuleDirectWriteCommandCount = 12

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
			if covering.Stream == "" && covering.Write == "" && covering.DirectRead == "" && len(covering.DirectReads) == 0 && covering.DirectWrite == "" && len(covering.DirectWrites) == 0 {
				t.Errorf("executable %s is not bound to a stream, write, direct_read, or direct_write command", key)
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
	if got := covered; got != 38 {
		t.Errorf("executable rows = %d, want 38", got)
	}
	if got := implementableNow; got != 1804 {
		t.Errorf("operations awaiting Zoom-local contracts = %d, want 1804", got)
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
	verifyDirectWriteOperationCommands(t, bundle, connector, surface, wantModuleDirectWriteCommandCount)
	verifyWriteCommands(t, bundle, connector, surface, wantModuleWriteCommandCount)
}

// TestChatbotDirectWriteCommandsAreReachable is the provider-category RED
// surface test. It enumerates every documented Chatbot action through the
// real commandrunner preflight, which prevents a plausible JSON declaration
// from claiming implementation while the binary would still say unknown
// command.
func TestChatbotDirectWriteCommandsAreReachable(t *testing.T) {
	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	wants := []struct {
		path      string
		operation string
		method    string
		apiPath   string
	}{
		{path: "chatbot messages send", operation: "zoom.send_chatbot_message", method: http.MethodPost, apiPath: "/v2/im/chat/messages"},
		{path: "chatbot messages edit", operation: "zoom.edit_chatbot_message", method: http.MethodPut, apiPath: "/v2/im/chat/messages/{message_id}"},
		{path: "chatbot messages delete", operation: "zoom.delete_chatbot_message", method: http.MethodDelete, apiPath: "/v2/im/chat/messages/{message_id}"},
		{path: "chatbot link-unfurls create", operation: "zoom.create_chatbot_link_unfurl", method: http.MethodPost, apiPath: "/v2/im/chat/users/{userId}/unfurls/{triggerId}"},
	}
	for _, want := range wants {
		if err := commandrunner.Preflight(connector, strings.Fields(want.path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want declared executable Chatbot action", want.path, err)
			continue
		}
		var found *connectors.CommandSurfaceCommand
		for i := range connector.CommandSurface().Commands {
			if connector.CommandSurface().Commands[i].Path == want.path {
				found = &connector.CommandSurface().Commands[i]
				break
			}
		}
		if found == nil || found.Intent != "direct_write" || found.Availability != "implemented" || found.Operation != want.operation || len(found.APISurface) != 1 || found.APISurface[0].Method != want.method || found.APISurface[0].Path != want.apiPath {
			t.Errorf("Chatbot command %q does not retain its declared operation/endpoint contract", want.path)
		}
	}
}

// TestSCIM2OperationCommandsAreReachable is the provider-category RED surface
// contract. It enumerates every operation in Zoom's own SCIM2 category through
// real commandrunner preflight, so an apparently complete JSON bundle cannot
// claim that a compiled `pm zoom scim2 …` route is executable while it remains
// an unknown command.
func TestSCIM2OperationCommandsAreReachable(t *testing.T) {
	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	wants := []struct {
		path      string
		operation string
		intent    string
		method    string
		apiPath   string
		policy    string
	}{
		{path: "scim2 groups list", operation: "zoom.list_scim2_groups", intent: "direct_read", method: http.MethodGet, apiPath: "/scim2/Groups", policy: "json_redacted"},
		{path: "scim2 groups create", operation: "zoom.create_scim2_group", intent: "direct_write", method: http.MethodPost, apiPath: "/scim2/Groups", policy: "json_redacted"},
		{path: "scim2 groups get", operation: "zoom.get_scim2_group", intent: "direct_read", method: http.MethodGet, apiPath: "/scim2/Groups/{groupId}", policy: "json_redacted"},
		{path: "scim2 groups delete", operation: "zoom.delete_scim2_group", intent: "direct_write", method: http.MethodDelete, apiPath: "/scim2/Groups/{groupId}", policy: "none"},
		{path: "scim2 groups update", operation: "zoom.update_scim2_group", intent: "direct_write", method: http.MethodPatch, apiPath: "/scim2/Groups/{groupId}", policy: "none"},
		{path: "scim2 users list", operation: "zoom.list_scim2_users", intent: "direct_read", method: http.MethodGet, apiPath: "/scim2/Users", policy: "json_redacted"},
		{path: "scim2 users create", operation: "zoom.create_scim2_user", intent: "direct_write", method: http.MethodPost, apiPath: "/scim2/Users", policy: "json_redacted"},
		{path: "scim2 users get", operation: "zoom.get_scim2_user", intent: "direct_read", method: http.MethodGet, apiPath: "/scim2/Users/{userId}", policy: "json_redacted"},
		{path: "scim2 users update", operation: "zoom.update_scim2_user", intent: "direct_write", method: http.MethodPut, apiPath: "/scim2/Users/{userId}", policy: "json_redacted"},
		{path: "scim2 users delete", operation: "zoom.delete_scim2_user", intent: "direct_write", method: http.MethodDelete, apiPath: "/scim2/Users/{userId}", policy: "none"},
		{path: "scim2 users deactivate", operation: "zoom.deactivate_scim2_user", intent: "direct_write", method: http.MethodPatch, apiPath: "/scim2/Users/{userId}", policy: "json_redacted"},
	}
	for _, want := range wants {
		if err := commandrunner.Preflight(connector, strings.Fields(want.path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want declared executable SCIM2 action", want.path, err)
			continue
		}
		var found *connectors.CommandSurfaceCommand
		for i := range connector.CommandSurface().Commands {
			if connector.CommandSurface().Commands[i].Path == want.path {
				found = &connector.CommandSurface().Commands[i]
				break
			}
		}
		if found == nil || found.Intent != want.intent || found.Availability != "implemented" || found.Operation != want.operation || len(found.APISurface) != 1 || found.APISurface[0].Method != want.method || found.APISurface[0].Path != want.apiPath || found.OutputPolicy != want.policy {
			t.Errorf("SCIM2 command %q does not retain its declared operation/endpoint/output contract", want.path)
			continue
		}
		for _, flag := range found.Flags {
			if flag.Name == "page" || flag.Name == "per-page" || flag.Name == "limit" {
				t.Errorf("SCIM2 command %q invents paging flag --%s", want.path, flag.Name)
			}
		}
	}
}

// verifyDirectWriteOperationCommands is the direct-write counterpart of
// verifyOperationCommands. It keeps rest_write declarations inside the real
// commandrunner preflight and plan lifecycle rather than treating a POST as a
// generic HTTP escape hatch.
func verifyDirectWriteOperationCommands(t *testing.T, bundle engine.Bundle, connector connectors.Connector, surface *connectors.CommandSurface, wantCount int) {
	t.Helper()
	ops := make(map[string]engine.OperationSpec, len(bundle.Operations))
	for _, op := range bundle.Operations {
		ops[op.ID] = op
	}

	got := 0
	for _, command := range surface.Commands {
		if command.Intent != "direct_write" {
			continue
		}
		got++
		if command.Availability != "implemented" {
			t.Errorf("direct-write command %q availability = %q, want implemented", command.Path, command.Availability)
			continue
		}
		op, ok := ops[command.Operation]
		if !ok {
			t.Errorf("direct-write command %q references operation %q, not found in operations.json", command.Path, command.Operation)
			continue
		}
		if op.Kind != "rest_write" || op.REST == nil {
			t.Errorf("direct-write command %q operation = %#v, want typed rest_write", command.Path, op)
			continue
		}
		if len(command.APISurface) != 1 || command.APISurface[0].Method != op.REST.Method || command.APISurface[0].Path != op.REST.Path {
			t.Errorf("direct-write command %q api_surface = %+v, want exactly [%s %s] matching operations.json", command.Path, command.APISurface, op.REST.Method, op.REST.Path)
		}
		if command.OutputPolicy != op.OutputPolicy || !supportedDirectWriteOutputPolicies[command.OutputPolicy] {
			t.Errorf("direct-write command %q output_policy = %q, want a supported declared rest_write policy %q", command.Path, command.OutputPolicy, op.OutputPolicy)
		}
		for _, flag := range command.Flags {
			mapsTo := strings.TrimSpace(flag.MapsTo)
			switch {
			case strings.HasPrefix(mapsTo, "path."):
				if !flag.Required {
					t.Errorf("direct-write command %q path flag --%s must be required", command.Path, flag.Name)
					continue
				}
				placeholder := "{" + strings.TrimPrefix(mapsTo, "path.") + "}"
				if !strings.Contains(op.REST.Path, placeholder) {
					t.Errorf("direct-write command %q path flag --%s maps_to %q, but %q is absent from %q", command.Path, flag.Name, mapsTo, placeholder, op.REST.Path)
				}
			case strings.HasPrefix(mapsTo, "body."), strings.HasPrefix(mapsTo, "query."):
				// Required and optional provider-defined body/query members are
				// checked by the operation body schema and connectorgen's exact
				// mapping validator. Requiring every member here would make a
				// documented optional field impossible to expose.
			default:
				t.Errorf("direct-write command %q flag --%s maps_to %q, want typed path.*, query.*, or body.* binding", command.Path, flag.Name, mapsTo)
			}
		}
		if err := commandrunner.Preflight(connector, strings.Fields(command.Path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want nil", command.Path, err)
		}
	}
	if got != wantCount {
		t.Errorf("reachable direct_write operation commands = %d, want %d", got, wantCount)
	}
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

var supportedDirectWriteOutputPolicies = map[string]bool{
	"none":                        true,
	"json":                        true,
	"json_redacted":               true,
	"write_result_redacted":       true,
	"gong_bounded_input_redacted": true,
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

// TestCustomerManagedKeysHybridCommandExecutesWithFixture pins Zoom's one
// customer-hosted Key Connector archival operation. It runs the real
// connector-command plan lifecycle against a loopback key connector so the
// declared customer host, operation-scoped key-connector JWT auth, exact
// POST, no-pagination contract, typed confirmation, and redacted key response
// are all executable without contacting an operator deployment.
func TestCustomerManagedKeysHybridCommandExecutesWithFixture(t *testing.T) {
	const (
		credentialName = "zoom-cmk-fixture"
		encryptContext = "fixture-encrypt-context"
		keyID          = "fixture-key-id"
		fixtureJWT     = "synthetic-key-connector-jwt"
	)
	wantStatus, wantBody := zoomWriteFixture(t, "decrypt_customer_managed_key_archival.json")
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		if request.Method != http.MethodPost {
			t.Fatal("Customer Managed Keys Hybrid fixture did not receive the declared POST")
		}
		if request.URL.Path != "/api/v2/kms/cse/archival/datakey/decrypt" {
			t.Fatal("Customer Managed Keys Hybrid fixture did not receive the normalized declared path")
		}
		if request.Header.Get("Authorization") != "Bearer "+fixtureJWT {
			t.Fatal("Customer Managed Keys Hybrid fixture did not receive the selected key-connector bearer auth")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("Customer Managed Keys Hybrid command sent undeclared query or pagination input")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode Customer Managed Keys Hybrid request: %v", err)
		}
		if len(body) != 2 || body["encrypt_context"] != encryptContext || body["key_id"] != keyID {
			t.Fatal("Customer Managed Keys Hybrid request did not contain exactly the declared typed body")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(wantStatus)
		_, _ = w.Write(wantBody)
	}))
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{
		Path:  []string{"customer-managed-keys-hybrid", "archival-key", "decrypt"},
		Flags: map[string][]string{"key-id": {keyID}},
	}); err == nil || !strings.Contains(err.Error(), "encrypt-context") {
		t.Fatalf("BuildWriteCommand without encrypt-context = %v, want required typed flag rejection", err)
	}

	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	application, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := application.AddCredential(context.Background(), app.AddCredentialRequest{
		Name:      credentialName,
		Connector: zoomBundleName,
		Config: map[string]string{
			"key_connector_base_url": server.URL + "/api/v2",
		},
		Secrets: map[string]string{"key_connector_jwt": fixtureJWT},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	plan, preview, err := application.PlanConnectorCommand(context.Background(), app.PlanConnectorCommandRequest{
		Connector:  zoomBundleName,
		Credential: credentialName,
		Path:       []string{"customer-managed-keys-hybrid", "archival-key", "decrypt"},
		Flags: map[string][]string{
			"encrypt-context": {encryptContext},
			"key-id":          {keyID},
		},
		Preview: true,
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand: %v", err)
	}
	if preview == nil || preview.Digest == "" || plan.ApprovalToken == "" || plan.ConfirmationChallenge != string(connectors.ConfirmationKindDestructive) {
		t.Fatalf("Customer Managed Keys Hybrid plan/preview = %#v/%#v, want typed confirmation evidence", plan, preview)
	}
	if len(plan.Sample) != 1 || plan.Sample[0]["encrypt_context"] != "redacted" || plan.Sample[0]["key_id"] != "redacted" {
		t.Fatalf("Customer Managed Keys Hybrid plan sample = %#v, want redacted sensitive body fields", plan.Sample)
	}
	if requests != 0 {
		t.Fatalf("Customer Managed Keys Hybrid planning reached network; requests = %d, want 0", requests)
	}

	if _, err := application.RunReverseETL(context.Background(), app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}); err == nil {
		t.Fatal("Customer Managed Keys Hybrid execution bypassed typed confirmation")
	}
	if requests != 0 {
		t.Fatalf("unconfirmed Customer Managed Keys Hybrid operation reached network; requests = %d, want 0", requests)
	}

	run, err := application.RunReverseETL(context.Background(), app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		t.Fatalf("RunReverseETL: %v", err)
	}
	if requests != 1 || run.Status != "completed" || run.OperationDirectWrite == nil || run.OperationDirectWrite.Status != wantStatus {
		t.Fatalf("Customer Managed Keys Hybrid run = %#v; requests = %d, want one completed declared POST", run, requests)
	}
	encoded, err := json.Marshal(run.OperationDirectWrite.Body)
	if err != nil {
		t.Fatalf("marshal Customer Managed Keys Hybrid result: %v", err)
	}
	for _, raw := range []string{keyID, "fixture-plaintext-key", "fixture-response-token"} {
		if strings.Contains(string(encoded), raw) {
			t.Fatal("Customer Managed Keys Hybrid result exposed a declared or generic sensitive response field")
		}
	}
	for _, field := range []string{"key_id", "plainkey", "token"} {
		if !strings.Contains(string(encoded), "\""+field+"_redacted\":true") {
			t.Fatalf("Customer Managed Keys Hybrid result is missing %s redaction marker", field)
		}
	}
}

// TestChatbotCommandsExecuteWithFixture proves every documented Chatbot write
// reaches the real plan/preview/approval lifecycle through a dedicated Client
// Credentials token exchange. The loopback transport catches regressions that
// a declaration-only preflight cannot: Basic client authentication, Bearer
// action authentication, the provider's exact method/path/body contract,
// destructive DELETE confirmation, response redaction, and the 204 status-only
// Link Unfurls action.
func TestChatbotCommandsExecuteWithFixture(t *testing.T) {
	const (
		credentialName = "zoom-chatbot-fixture"
		clientID       = "fixture-chatbot-client-id"
		clientSecret   = "fixture-chatbot-client-secret"
		accessToken    = "fixture-chatbot-access-token"
	)

	type chatbotAction struct {
		name              string
		path              []string
		flags             map[string][]string
		method            string
		requestPath       string
		fixture           string
		expectedBody      map[string]any
		destructive       bool
		expectsBody       bool
		inputSensitive    []string
		responseSensitive []string
		responseMarkers   []string
		status            int
		response          json.RawMessage
	}
	actions := []chatbotAction{
		{
			name:        "send",
			path:        []string{"chatbot", "messages", "send"},
			method:      http.MethodPost,
			requestPath: "/v2/im/chat/messages",
			fixture:     "send_chatbot_message.json",
			flags: map[string][]string{
				"account-id":          {"fixture-chatbot-account"},
				"content":             {`{"body":[{"type":"section","text":"fixture chatbot message"}]}`},
				"robot-jid":           {"fixture-robot-jid"},
				"to-jid":              {"fixture-recipient-jid"},
				"user-jid":            {"fixture-sender-jid"},
				"is-markdown-support": {"true"},
				"reply-to":            {"fixture-reply-jid"},
				"visible-to-user":     {"true"},
			},
			expectedBody: map[string]any{
				"account_id":          "fixture-chatbot-account",
				"content":             map[string]any{"body": []any{map[string]any{"type": "section", "text": "fixture chatbot message"}}},
				"robot_jid":           "fixture-robot-jid",
				"to_jid":              "fixture-recipient-jid",
				"user_jid":            "fixture-sender-jid",
				"is_markdown_support": true,
				"reply_to":            "fixture-reply-jid",
				"visible_to_user":     "true",
			},
			expectsBody:       true,
			inputSensitive:    []string{"fixture-chatbot-account", "fixture chatbot message", "fixture-robot-jid", "fixture-recipient-jid", "fixture-sender-jid", "fixture-reply-jid"},
			responseSensitive: []string{"fixture-chatbot-account", "fixture-chatbot-message", "fixture chatbot response", "fixture-chatbot-response-token"},
			responseMarkers:   []string{"account_id", "content", "message_id", "token"},
		},
		{
			name:        "edit",
			path:        []string{"chatbot", "messages", "edit"},
			method:      http.MethodPut,
			requestPath: "/v2/im/chat/messages/fixture-chatbot-message",
			fixture:     "edit_chatbot_message.json",
			flags: map[string][]string{
				"message-id":          {"fixture-chatbot-message"},
				"account-id":          {"fixture-chatbot-account"},
				"content":             {`{"body":[{"type":"section","text":"fixture edited chatbot message"}]}`},
				"robot-jid":           {"fixture-robot-jid"},
				"is-markdown-support": {"false"},
				"user-jid":            {"fixture-sender-jid"},
			},
			expectedBody: map[string]any{
				"account_id":          "fixture-chatbot-account",
				"content":             map[string]any{"body": []any{map[string]any{"type": "section", "text": "fixture edited chatbot message"}}},
				"robot_jid":           "fixture-robot-jid",
				"is_markdown_support": false,
				"user_jid":            "fixture-sender-jid",
			},
			expectsBody:       true,
			inputSensitive:    []string{"fixture-chatbot-account", "fixture edited chatbot message", "fixture-robot-jid", "fixture-sender-jid"},
			responseSensitive: []string{"fixture-chatbot-message", "fixture-robot-jid", "fixture edited chatbot response", "fixture-chatbot-response-token"},
			responseMarkers:   []string{"content", "message_id", "robot_jid", "token"},
		},
		{
			name:              "delete",
			path:              []string{"chatbot", "messages", "delete"},
			method:            http.MethodDelete,
			requestPath:       "/v2/im/chat/messages/fixture-chatbot-message",
			fixture:           "delete_chatbot_message.json",
			flags:             map[string][]string{"message-id": {"fixture-chatbot-message"}},
			destructive:       true,
			expectsBody:       true,
			responseSensitive: []string{"fixture-chatbot-message", "fixture-chatbot-response-token"},
			responseMarkers:   []string{"message_id", "token"},
		},
		{
			name:        "link unfurl",
			path:        []string{"chatbot", "link-unfurls", "create"},
			method:      http.MethodPost,
			requestPath: "/v2/im/chat/users/fixture-chatbot-user/unfurls/fixture-trigger",
			fixture:     "create_chatbot_link_unfurl.json",
			flags: map[string][]string{
				"user-id":    {"fixture-chatbot-user"},
				"trigger-id": {"fixture-trigger"},
				"content":    {`{"type":"message","text":"fixture unfurl"}`},
			},
			expectedBody:   map[string]any{"content": `{"type":"message","text":"fixture unfurl"}`},
			inputSensitive: []string{`{"type":"message","text":"fixture unfurl"}`},
		},
	}

	byRequest := make(map[string]*chatbotAction, len(actions))
	for i := range actions {
		actions[i].status, actions[i].response = zoomWriteFixture(t, actions[i].fixture)
		byRequest[actions[i].method+" "+actions[i].requestPath] = &actions[i]
	}

	tokenRequests := 0
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		tokenRequests++
		if request.Method != http.MethodPost {
			t.Fatal("Chatbot token fixture did not receive POST")
		}
		gotID, gotSecret, ok := request.BasicAuth()
		if !ok || gotID != clientID || gotSecret != clientSecret {
			t.Fatal("Chatbot token fixture did not receive declared Basic client authentication")
		}
		if err := request.ParseForm(); err != nil {
			t.Fatalf("parse Chatbot token form: %v", err)
		}
		if request.PostForm.Get("grant_type") != "client_credentials" || request.PostForm.Get("client_id") != "" || request.PostForm.Get("client_secret") != "" {
			t.Fatal("Chatbot token fixture received credentials in the form instead of the declared Basic exchange")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"fixture-chatbot-access-token","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenServer.Close()

	apiRequests := 0
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		apiRequests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("Chatbot API fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("Chatbot API fixture did not receive the token-exchange Bearer credential")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("Chatbot action sent undeclared query or pagination input")
		}
		if action.expectedBody == nil {
			var unexpected any
			if err := json.NewDecoder(request.Body).Decode(&unexpected); err == nil {
				t.Fatal("Chatbot no-body action unexpectedly sent a request body")
			}
		} else {
			if !strings.HasPrefix(request.Header.Get("Content-Type"), "application/json") {
				t.Fatal("Chatbot body action did not declare JSON content")
			}
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode Chatbot action body: %v", err)
			}
			if !reflect.DeepEqual(body, action.expectedBody) {
				t.Fatal("Chatbot action body did not contain exactly the declared typed fields")
			}
		}
		if len(action.response) > 0 {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(action.status)
		if len(action.response) > 0 {
			_, _ = w.Write(action.response)
		}
	}))
	defer apiServer.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	for _, action := range actions {
		if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{Path: action.path, Flags: action.flags}); err != nil {
			t.Fatalf("BuildWriteCommand(%s) = %v, want declared Chatbot command", action.name, err)
		}
	}

	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	application, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := application.AddCredential(context.Background(), app.AddCredentialRequest{
		Name:      credentialName,
		Connector: zoomBundleName,
		Config: map[string]string{
			"base_url":          apiServer.URL + "/v2",
			"chatbot_token_url": tokenServer.URL,
		},
		Secrets: map[string]string{
			"chatbot_client_id":     clientID,
			"chatbot_client_secret": clientSecret,
		},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			beforeAPIRequests, beforeTokenRequests := apiRequests, tokenRequests
			plan, preview, err := application.PlanConnectorCommand(context.Background(), app.PlanConnectorCommandRequest{
				Connector:  zoomBundleName,
				Credential: credentialName,
				Path:       action.path,
				Flags:      action.flags,
				Preview:    true,
			})
			if err != nil {
				t.Fatalf("PlanConnectorCommand(%s): %v", action.name, err)
			}
			if preview == nil || preview.Digest == "" || plan.ApprovalToken == "" {
				t.Fatal("Chatbot plan did not produce a no-network preview and single-use approval")
			}
			if action.destructive != (plan.ConfirmationChallenge == string(connectors.ConfirmationKindDestructive)) {
				t.Fatal("Chatbot plan did not retain the declared destructive confirmation policy")
			}
			encodedPlan, err := json.Marshal(plan.Sample)
			if err != nil {
				t.Fatalf("marshal Chatbot plan sample: %v", err)
			}
			for _, raw := range action.inputSensitive {
				if strings.Contains(string(encodedPlan), raw) {
					t.Fatal("Chatbot plan sample exposed a declared sensitive input")
				}
			}
			if apiRequests != beforeAPIRequests || tokenRequests != beforeTokenRequests {
				t.Fatal("Chatbot plan or preview reached a fixture endpoint")
			}

			runRequest := app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}
			if action.destructive {
				if _, err := application.RunReverseETL(context.Background(), runRequest); err == nil {
					t.Fatal("Chatbot DELETE execution bypassed typed destructive confirmation")
				}
				if apiRequests != beforeAPIRequests || tokenRequests != beforeTokenRequests {
					t.Fatal("unconfirmed Chatbot DELETE reached a fixture endpoint")
				}
				runRequest.Confirmation = connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
			}
			run, err := application.RunReverseETL(context.Background(), runRequest)
			if err != nil {
				t.Fatalf("RunReverseETL(%s): %v", action.name, err)
			}
			if run.Status != "completed" || run.OperationDirectWrite == nil || run.OperationDirectWrite.Status != action.status {
				t.Fatal("Chatbot action did not complete with its declared response status")
			}
			if !action.expectsBody {
				if run.OperationDirectWrite.Body != nil {
					t.Fatal("Chatbot status-only action returned an invented response body")
				}
				return
			}
			encoded, err := json.Marshal(run.OperationDirectWrite.Body)
			if err != nil {
				t.Fatalf("marshal Chatbot action response: %v", err)
			}
			for _, raw := range action.responseSensitive {
				if strings.Contains(string(encoded), raw) {
					t.Fatal("Chatbot action response exposed a declared or generic sensitive field")
				}
			}
			for _, field := range action.responseMarkers {
				if !strings.Contains(string(encoded), `"`+field+`_redacted":true`) {
					t.Fatal("Chatbot action response is missing a required redaction marker")
				}
			}
		})
	}
	if apiRequests != len(actions) || tokenRequests != len(actions) {
		t.Fatal("Chatbot fixtures did not receive exactly one token exchange and one action request per command")
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
	if fixture.Response.Status == 0 {
		t.Fatalf("writes fixture %s requires response.status", file)
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
