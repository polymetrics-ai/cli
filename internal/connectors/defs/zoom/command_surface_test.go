// Package zoom holds connector-local reachability evidence for the Zoom
// declarative bundle. The bundle itself is JSON; this test keeps its provider
// inventory and executable command surface from drifting apart.
package zoom

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
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
// quality-management reads (5), Cobrowse SDK reads (4), SCIM2 reads (4),
// Virtual Agent reads (9), Auto Dialer reads (8), Tasks reads (6), and Workforce
// Management reads (11). Clips adds six JSON direct reads; its binary download
// has a separate typed executor and is asserted in TestClipsOperationCommandsAreReachable.
// CRC adds nine bounded direct reads. AI Services adds twelve REST direct
// reads; its documented Live Scribe WebSocket has a separate fixed-session
// executor and is asserted in TestAIServicesOperationCommandsAreReachable.
// Chatbot, SCIM2, Virtual Agent, Auto Dialer, Tasks, Workforce Management, and
// Clips mutations are tracked by wantModuleDirectWriteCommandCount.
const wantModuleOperationCommandCount = 82

// wantModuleDirectWriteCommandCount is the running total of implemented
// operations.json rest_write commands. It is distinct from writes.json
// reverse-ETL actions because direct writes are executable only through the
// typed plan lifecycle.
// Clips has fourteen documented write endpoints and sixteen concrete typed
// commands: transfer and multipart-event oneOf contracts are represented by
// separate named request arms rather than a generic union body. CRC adds its
// eleven provider-defined mutations, including seven status-only actions. AI
// Services adds six typed JSON mutations and three destructive status-only
// cancellation actions.
const wantModuleDirectWriteCommandCount = 78

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
			if covering.Stream == "" && covering.Write == "" && covering.DirectRead == "" && len(covering.DirectReads) == 0 && covering.DirectWrite == "" && len(covering.DirectWrites) == 0 && covering.WebSocketSession == "" {
				t.Errorf("executable %s is not bound to a stream, write, direct_read, direct_write, or websocket_session command", key)
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
	if got := covered; got != 165 {
		t.Errorf("executable rows = %d, want 165", got)
	}
	if got := implementableNow; got != 1677 {
		t.Errorf("operations awaiting Zoom-local contracts = %d, want 1677", got)
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

// TestVirtualAgentOperationCommandsAreReachable is the provider-category RED
// surface contract. It enumerates every action in Zoom's published Virtual
// Agent artifact through the real commandrunner preflight, so an apparently
// complete JSON bundle cannot claim that a compiled `pm zoom virtual-agent …`
// route is executable while it remains an unknown command.
func TestVirtualAgentOperationCommandsAreReachable(t *testing.T) {
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
		{path: "virtual-agent knowledge-bases articles list", operation: "zoom.list_virtual_agent_articles", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/km/kbs/{kbId}/articles", policy: "json_redacted"},
		{path: "virtual-agent knowledge-bases articles create", operation: "zoom.create_virtual_agent_article", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/km/kbs/{kbId}/articles", policy: "json_redacted"},
		{path: "virtual-agent knowledge-bases articles get", operation: "zoom.get_virtual_agent_article", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/km/kbs/{kbId}/articles/{articleId}", policy: "json_redacted"},
		{path: "virtual-agent knowledge-bases articles update", operation: "zoom.update_virtual_agent_article", intent: "direct_write", method: http.MethodPut, apiPath: "/v2/km/kbs/{kbId}/articles/{articleId}", policy: "json_redacted"},
		{path: "virtual-agent knowledge-bases articles delete", operation: "zoom.delete_virtual_agent_article", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/km/kbs/{kbId}/articles/{articleId}", policy: "none"},
		{path: "virtual-agent knowledge-bases sync create", operation: "zoom.create_virtual_agent_sync_request", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/km/kbs/{kbId}/sync", policy: "json_redacted"},
		{path: "virtual-agent knowledge-bases sync get", operation: "zoom.get_virtual_agent_sync", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/km/kbs/{kbId}/sync/{syncId}", policy: "json_redacted"},
		{path: "virtual-agent reports engagements list", operation: "zoom.list_virtual_agent_engagements", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/virtual_agent/report/engagements", policy: "json_redacted"},
		{path: "virtual-agent reports engagements query-details list", operation: "zoom.list_virtual_agent_engagement_query_details", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/virtual_agent/report/engagements/query_details", policy: "json_redacted"},
		{path: "virtual-agent reports engagements variable-details list", operation: "zoom.list_virtual_agent_engagement_variable_details", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/virtual_agent/report/engagements/variables", policy: "json_redacted"},
		{path: "virtual-agent reports surveys list", operation: "zoom.list_virtual_agent_surveys", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/virtual_agent/report/surveys", policy: "json_redacted"},
		{path: "virtual-agent reports transcripts list", operation: "zoom.list_virtual_agent_transcripts", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/virtual_agent/report/transcripts", policy: "json_redacted"},
		{path: "virtual-agent reports operation-logs list", operation: "zoom.list_virtual_agent_operation_logs", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/ai_studio/reports/operation_logs", policy: "json_redacted"},
	}
	for _, want := range wants {
		if err := commandrunner.Preflight(connector, strings.Fields(want.path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want declared executable Virtual Agent action", want.path, err)
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
			t.Errorf("Virtual Agent command %q does not retain its declared operation/endpoint/output contract", want.path)
			continue
		}
		for _, flag := range found.Flags {
			if flag.Name == "page" || flag.Name == "per-page" || flag.Name == "limit" || flag.Name == "page-size" || flag.Name == "next-page-token" {
				t.Errorf("Virtual Agent command %q invents paging flag --%s", want.path, flag.Name)
			}
		}
	}
}

// TestAutoDialerOperationCommandsAreReachable is the provider-category RED
// surface contract. It enumerates every action in Zoom's published Auto Dialer
// artifact through the real commandrunner preflight, so an apparently complete
// JSON bundle cannot claim a compiled `pm zoom auto-dialer …` route is
// executable while it remains an unknown command.
func TestAutoDialerOperationCommandsAreReachable(t *testing.T) {
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
		{path: "auto-dialer call-histories get", operation: "zoom.get_auto_dialer_call_history", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/dialer/call-histories/{callHistoryId}", policy: "json_redacted"},
		{path: "auto-dialer call-history list", operation: "zoom.list_auto_dialer_call_history", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/dialer/call-history", policy: "json_redacted"},
		{path: "auto-dialer reports call-history list", operation: "zoom.list_auto_dialer_report_call_history", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/dialer/reports/call-history", policy: "json_redacted"},
		{path: "auto-dialer reports seller-productivity get", operation: "zoom.get_auto_dialer_seller_productivity_report", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/dialer/reports/seller-productivity", policy: "json_redacted"},
		{path: "auto-dialer call-lists list", operation: "zoom.list_auto_dialer_call_lists", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/dialer/call-lists", policy: "json_redacted"},
		{path: "auto-dialer call-lists create", operation: "zoom.create_auto_dialer_call_list", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/dialer/call-lists", policy: "json_redacted"},
		{path: "auto-dialer call-lists get", operation: "zoom.get_auto_dialer_call_list", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/dialer/call-lists/{callListId}", policy: "json_redacted"},
		{path: "auto-dialer call-lists delete", operation: "zoom.delete_auto_dialer_call_list", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/dialer/call-lists/{callListId}", policy: "none"},
		{path: "auto-dialer call-lists update", operation: "zoom.update_auto_dialer_call_list", intent: "direct_write", method: http.MethodPatch, apiPath: "/v2/dialer/call-lists/{callListId}", policy: "none"},
		{path: "auto-dialer call-lists prospects list", operation: "zoom.list_auto_dialer_call_list_prospects", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/dialer/call-lists/{callListId}/prospects", policy: "json_redacted"},
		{path: "auto-dialer call-lists prospects create", operation: "zoom.create_auto_dialer_prospect", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/dialer/call-lists/{callListId}/prospects", policy: "json_redacted"},
		{path: "auto-dialer call-lists prospects update-batch", operation: "zoom.update_auto_dialer_prospects_batch", intent: "direct_write", method: http.MethodPatch, apiPath: "/v2/dialer/call-lists/{callListId}/prospects", policy: "json_redacted"},
		{path: "auto-dialer call-lists prospects create-batch", operation: "zoom.create_auto_dialer_prospects_batch", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/dialer/call-lists/{callListId}/prospects/batch", policy: "json_redacted"},
		{path: "auto-dialer call-lists prospects delete", operation: "zoom.delete_auto_dialer_prospect", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/dialer/call-lists/{callListId}/prospects/{prospectId}", policy: "none"},
		{path: "auto-dialer call-lists prospects update", operation: "zoom.update_auto_dialer_prospect", intent: "direct_write", method: http.MethodPatch, apiPath: "/v2/dialer/call-lists/{callListId}/prospects/{prospectId}", policy: "none"},
		{path: "auto-dialer prospects get", operation: "zoom.get_auto_dialer_prospect", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/dialer/prospects/{prospectId}", policy: "json_redacted"},
	}
	for _, want := range wants {
		if err := commandrunner.Preflight(connector, strings.Fields(want.path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want declared executable Auto Dialer action", want.path, err)
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
			t.Errorf("Auto Dialer command %q does not retain its declared operation/endpoint/output contract", want.path)
			continue
		}
		for _, flag := range found.Flags {
			if flag.Name == "page" || flag.Name == "per-page" || flag.Name == "limit" || flag.Name == "page-size" || flag.Name == "next-page-token" {
				t.Errorf("Auto Dialer command %q invents paging flag --%s", want.path, flag.Name)
			}
		}
	}
}

// TestTasksOperationCommandsAreReachable is the provider-category RED surface
// contract. It enumerates each action in Zoom's published Tasks artifact
// through the real commandrunner preflight, including the bounded multipart
// file upload, so a declaration cannot claim implementation while the binary
// would still report an unknown route.
func TestTasksOperationCommandsAreReachable(t *testing.T) {
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
		{path: "tasks assignees list", operation: "zoom.list_task_assignees", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/tasks/items/{taskId}/assignees", policy: "json_redacted"},
		{path: "tasks assignees add", operation: "zoom.add_task_assignees", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/tasks/items/{taskId}/assignees", policy: "json_redacted"},
		{path: "tasks assignees remove", operation: "zoom.remove_task_assignee", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/tasks/items/{taskId}/assignees/{userId}", policy: "none"},
		{path: "tasks collaborators list", operation: "zoom.list_task_collaborators", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/tasks/items/{taskId}/collaborators", policy: "json_redacted"},
		{path: "tasks collaborators add", operation: "zoom.add_task_collaborators", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/tasks/items/{taskId}/collaborators", policy: "json_redacted"},
		{path: "tasks collaborators remove", operation: "zoom.remove_task_collaborator", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/tasks/items/{taskId}/collaborators/{userId}", policy: "none"},
		{path: "tasks comments list", operation: "zoom.list_task_comments", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/tasks/items/{taskId}/comments", policy: "json_redacted"},
		{path: "tasks comments add", operation: "zoom.add_task_comment", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/tasks/items/{taskId}/comments", policy: "json_redacted"},
		{path: "tasks comments delete", operation: "zoom.delete_task_comment", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/tasks/items/{taskId}/comments/{commentId}", policy: "none"},
		{path: "tasks files upload", operation: "zoom.upload_task_file", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/tasks/files", policy: "json_redacted"},
		{path: "tasks imports submit", operation: "zoom.submit_task_import", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/tasks/imports", policy: "json_redacted"},
		{path: "tasks imports get", operation: "zoom.get_task_import", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/tasks/imports/{importId}", policy: "json_redacted"},
		{path: "tasks items list", operation: "zoom.list_tasks", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/tasks/items", policy: "json_redacted"},
		{path: "tasks items create", operation: "zoom.create_task", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/tasks/items", policy: "json_redacted"},
		{path: "tasks items get", operation: "zoom.get_task", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/tasks/items/{taskId}", policy: "json_redacted"},
		{path: "tasks items delete", operation: "zoom.delete_task", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/tasks/items/{taskId}", policy: "none"},
		{path: "tasks items update", operation: "zoom.update_task", intent: "direct_write", method: http.MethodPatch, apiPath: "/v2/tasks/items/{taskId}", policy: "none"},
	}
	for _, want := range wants {
		if err := commandrunner.Preflight(connector, strings.Fields(want.path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want declared executable Tasks action", want.path, err)
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
			t.Errorf("Tasks command %q does not retain its declared operation/endpoint/output contract", want.path)
			continue
		}
		for _, flag := range found.Flags {
			if flag.Name == "page" || flag.Name == "per-page" || flag.Name == "limit" || flag.Name == "page-size" || flag.Name == "next-page-token" {
				t.Errorf("Tasks command %q invents paging flag --%s", want.path, flag.Name)
			}
		}
	}
}

// TestWorkforceManagementOperationCommandsAreReachable is the provider-category
// RED surface contract. It enumerates every documented action in Zoom's
// Workforce Management artifact through real commandrunner preflight, including
// its declared CSV imports, so a bundle cannot claim implementation while the
// executable binary would still report an unknown command.
func TestWorkforceManagementOperationCommandsAreReachable(t *testing.T) {
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
		{path: "workforce-management filter-groups list", operation: "zoom.list_workforce_filter_groups", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/workforce-management/filter-groups", policy: "json_redacted"},
		{path: "workforce-management forecasts list", operation: "zoom.list_workforce_forecasts", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/workforce-management/forecasts", policy: "json_redacted"},
		{path: "workforce-management forecasts scheduling-groups get", operation: "zoom.get_workforce_forecast_scheduling_group", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/workforce-management/forecasts/{forecastId}/scheduling-groups/{schedulingGroupId}", policy: "json_redacted"},
		{path: "workforce-management imports historical-agent-status upload", operation: "zoom.upload_workforce_historical_agent_status", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/workforce-management/imports/historical-agent-status", policy: "json_redacted"},
		{path: "workforce-management imports historical-agent-status delete", operation: "zoom.delete_workforce_historical_agent_status", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/workforce-management/imports/historical-agent-status", policy: "none"},
		{path: "workforce-management imports historical-queue-metrics upload", operation: "zoom.upload_workforce_historical_queue_metrics", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/workforce-management/imports/historical-queue-metrics", policy: "json_redacted"},
		{path: "workforce-management imports staffing upload", operation: "zoom.upload_workforce_forecast_staffing", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/workforce-management/imports/staffing", policy: "json_redacted"},
		{path: "workforce-management imports historical-queue-metrics get", operation: "zoom.get_workforce_historical_queue_metrics_import", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/workforce-management/imports/{importId}/historical-queue-metrics", policy: "json_redacted"},
		{path: "workforce-management organizational-groups list", operation: "zoom.list_workforce_organizational_groups", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/workforce-management/organizational-groups", policy: "json_redacted"},
		{path: "workforce-management organizational-groups create", operation: "zoom.create_workforce_organizational_group", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/workforce-management/organizational-groups", policy: "json_redacted"},
		{path: "workforce-management organizational-groups get", operation: "zoom.get_workforce_organizational_group", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/workforce-management/organizational-groups/{organizationalGroupId}", policy: "json_redacted"},
		{path: "workforce-management organizational-groups delete", operation: "zoom.delete_workforce_organizational_group", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/workforce-management/organizational-groups/{organizationalGroupId}", policy: "none"},
		{path: "workforce-management organizational-groups update", operation: "zoom.update_workforce_organizational_group", intent: "direct_write", method: http.MethodPatch, apiPath: "/v2/workforce-management/organizational-groups/{organizationalGroupId}", policy: "json_redacted"},
		{path: "workforce-management reports adherence agents list", operation: "zoom.list_workforce_adherence_agents", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/workforce-management/reports/adherence/agents", policy: "json_redacted"},
		{path: "workforce-management reports schedules agents list", operation: "zoom.list_workforce_report_schedule_agents", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/workforce-management/reports/schedules/agents", policy: "json_redacted"},
		{path: "workforce-management schedules agents list", operation: "zoom.list_workforce_schedule_agents", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/workforce-management/schedules/agents", policy: "json_redacted"},
		{path: "workforce-management scheduling-groups list", operation: "zoom.list_workforce_scheduling_groups", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/workforce-management/scheduling-groups", policy: "json_redacted"},
		{path: "workforce-management users list", operation: "zoom.list_workforce_users", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/workforce-management/users", policy: "json_redacted"},
	}
	for _, want := range wants {
		if err := commandrunner.Preflight(connector, strings.Fields(want.path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want declared executable Workforce Management action", want.path, err)
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
			t.Errorf("Workforce Management command %q does not retain its declared operation/endpoint/output contract", want.path)
			continue
		}
		for _, flag := range found.Flags {
			if flag.Name == "page" || flag.Name == "per-page" || flag.Name == "limit" || flag.Name == "page-size" || flag.Name == "next-page-token" {
				t.Errorf("Workforce Management command %q invents paging flag --%s", want.path, flag.Name)
			}
		}
	}
}

// TestClipsOperationCommandsAreReachable is the provider-category RED surface
// contract. It enumerates every Clips endpoint through real commandrunner
// preflight. The source's two discriminated payloads intentionally have two
// concrete command contracts each, rather than a generic oneOf/raw-body path.
func TestClipsOperationCommandsAreReachable(t *testing.T) {
	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	wants := []struct {
		path      string
		operation string
		intent    string
		method    string
		apiPath   string
		policy    string
		opPolicy  string
		minItems  int
	}{
		{path: "clips list", operation: "zoom.list_clips", intent: "direct_read", method: http.MethodGet, apiPath: "/clips", policy: "json_redacted"},
		{path: "clips starred set", operation: "zoom.set_clips_starred", intent: "direct_write", method: http.MethodPost, apiPath: "/clips/starred", policy: "none"},
		{path: "clips collaborators list", operation: "zoom.list_clip_collaborators", intent: "direct_read", method: http.MethodGet, apiPath: "/clips/{clipId}/collaborators", policy: "json_redacted"},
		{path: "clips collaborators share", operation: "zoom.share_clip_collaborators", intent: "direct_write", method: http.MethodPost, apiPath: "/clips/{clipId}/collaborators", policy: "json_redacted"},
		{path: "clips collaborators remove", operation: "zoom.remove_clip_collaborator", intent: "direct_write", method: http.MethodDelete, apiPath: "/clips/{clipId}/collaborators", policy: "none"},
		{path: "clips comments list", operation: "zoom.list_clip_comments", intent: "direct_read", method: http.MethodGet, apiPath: "/clips/{clipId}/comments", policy: "json_redacted"},
		{path: "clips comments delete", operation: "zoom.delete_clip_comment", intent: "direct_write", method: http.MethodDelete, apiPath: "/clips/{clipId}/comments/{commentId}", policy: "none"},
		{path: "clips download", operation: "zoom.download_clip", intent: "binary_download", method: http.MethodGet, apiPath: "/clips/{clipId}/download", policy: "", opPolicy: "binary_file_bounded"},
		{path: "clips get", operation: "zoom.get_clip", intent: "direct_read", method: http.MethodGet, apiPath: "/clips/{clipId}", policy: "json_redacted"},
		{path: "clips delete", operation: "zoom.delete_clip", intent: "direct_write", method: http.MethodDelete, apiPath: "/clips/{clipId}", policy: "none"},
		{path: "clips update", operation: "zoom.update_clip", intent: "direct_write", method: http.MethodPatch, apiPath: "/clips/{clipId}", policy: "none"},
		{path: "clips chapters get", operation: "zoom.get_clip_chapters", intent: "direct_read", method: http.MethodGet, apiPath: "/clips/{clipId}/chapters", policy: "json_redacted"},
		{path: "clips chapters create", operation: "zoom.create_clip_chapters", intent: "direct_write", method: http.MethodPost, apiPath: "/clips/{clipId}/chapters", policy: "json_redacted"},
		{path: "clips duplicate", operation: "zoom.duplicate_clip", intent: "direct_write", method: http.MethodPost, apiPath: "/clips/{clipId}/duplicate", policy: "json_redacted"},
		{path: "clips share-settings update", operation: "zoom.update_clip_share_settings", intent: "direct_write", method: http.MethodPatch, apiPath: "/clips/{clipId}/share_settings", policy: "none"},
		{path: "clips transfers partial", operation: "zoom.transfer_clips_partial", intent: "direct_write", method: http.MethodPost, apiPath: "/clips/transfers", policy: "json_redacted", minItems: 1},
		{path: "clips transfers full", operation: "zoom.transfer_clips_full", intent: "direct_write", method: http.MethodPost, apiPath: "/clips/transfers", policy: "json_redacted"},
		{path: "clips transfers get", operation: "zoom.get_clip_transfer", intent: "direct_read", method: http.MethodGet, apiPath: "/clips/transfers/{taskId}", policy: "json_redacted"},
		{path: "clips files upload", operation: "zoom.upload_clip_file", intent: "direct_write", method: http.MethodPost, apiPath: "/clips/files", policy: "json_redacted"},
		{path: "clips files multipart upload", operation: "zoom.upload_clip_multipart_part", intent: "direct_write", method: http.MethodPost, apiPath: "/clips/files/multipart", policy: "json_redacted"},
		{path: "clips files multipart initiate", operation: "zoom.initiate_clip_multipart_upload", intent: "direct_write", method: http.MethodPost, apiPath: "/clips/files/multipart/upload_events", policy: "json_redacted"},
		{path: "clips files multipart complete", operation: "zoom.complete_clip_multipart_upload", intent: "direct_write", method: http.MethodPost, apiPath: "/clips/files/multipart/upload_events", policy: "json_redacted"},
		{path: "clips files temporary upload", operation: "zoom.upload_clip_temporary_file", intent: "direct_write", method: http.MethodPost, apiPath: "/clips/files/tmp", policy: "json_redacted"},
	}
	for _, want := range wants {
		if err := commandrunner.Preflight(connector, strings.Fields(want.path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want declared executable Clips action", want.path, err)
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
			t.Errorf("Clips command %q does not retain its declared operation/endpoint/output contract", want.path)
			continue
		}
		if want.opPolicy != "" {
			var operation *engine.OperationSpec
			for i := range bundle.Operations {
				if bundle.Operations[i].ID == want.operation {
					operation = &bundle.Operations[i]
					break
				}
			}
			gotPolicy := ""
			if operation != nil {
				gotPolicy = operation.OutputPolicy
			}
			if gotPolicy != want.opPolicy {
				t.Errorf("Clips binary command %q operation output_policy = %q, want %q", want.path, gotPolicy, want.opPolicy)
			}
		}
		hasListUserID := false
		gotMinItems := 0
		for _, flag := range found.Flags {
			if flag.Name == "page" || flag.Name == "per-page" || flag.Name == "limit" || flag.Name == "page-size" || flag.Name == "next-page-token" {
				t.Errorf("Clips command %q invents paging flag --%s", want.path, flag.Name)
			}
			if want.path == "clips list" && flag.Name == "user-id" && flag.MapsTo == "query.user_id" && !flag.Required && flag.AllowEmpty != nil && !*flag.AllowEmpty {
				hasListUserID = true
			}
			if want.path == "clips transfers partial" && flag.Name == "clip-id-list" {
				gotMinItems = flag.MinItems
			}
		}
		if want.path == "clips list" && !hasListUserID {
			t.Errorf("Clips command %q does not retain the documented optional --user-id query selector", want.path)
		}
		if want.minItems > 0 {
			if gotMinItems != want.minItems {
				t.Errorf("Clips command %q clip-id-list min_items = %d, want %d", want.path, gotMinItems, want.minItems)
			}
			var operation *engine.OperationSpec
			for i := range bundle.Operations {
				if bundle.Operations[i].ID == want.operation {
					operation = &bundle.Operations[i]
					break
				}
			}
			if operation == nil || operation.REST == nil {
				t.Errorf("Clips command %q does not retain its REST operation schema", want.path)
				continue
			}
			var schema struct {
				Properties map[string]struct {
					MinItems int `json:"minItems"`
				} `json:"properties"`
			}
			if err := json.Unmarshal(operation.REST.BodySchema, &schema); err != nil {
				t.Fatalf("decode Clips command %q body schema: %v", want.path, err)
			}
			if schema.Properties["clip_id_list"].MinItems != want.minItems {
				t.Errorf("Clips operation %q clip_id_list minItems = %d, want %d", want.operation, schema.Properties["clip_id_list"].MinItems, want.minItems)
			}
		}
	}
}

// TestCRCOperationCommandsAreReachable is the provider-category RED surface
// contract. It enumerates every CRC operation from Zoom's own artifact through
// real commandrunner preflight, including the declared private-key redaction
// and status-only mutation policies.
func TestCRCOperationCommandsAreReachable(t *testing.T) {
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
		{path: "crc managed-rooms account-setting get", operation: "zoom.get_crc_managed_room_account_setting", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/crc/managed_rooms/account_setting", policy: "json_redacted"},
		{path: "crc managed-rooms account-setting update", operation: "zoom.update_crc_managed_room_account_setting", intent: "direct_write", method: http.MethodPatch, apiPath: "/v2/crc/managed_rooms/account_setting", policy: "none"},
		{path: "crc api-connectors list", operation: "zoom.list_crc_api_connectors", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/crc/api_connectors", policy: "json_redacted"},
		{path: "crc api-connectors create", operation: "zoom.create_crc_api_connector", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/crc/api_connectors", policy: "json_redacted"},
		{path: "crc api-connectors get", operation: "zoom.get_crc_api_connector", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/crc/api_connectors/{connectorId}", policy: "json_redacted"},
		{path: "crc api-connectors delete", operation: "zoom.delete_crc_api_connector", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/crc/api_connectors/{connectorId}", policy: "none"},
		{path: "crc api-connectors update", operation: "zoom.update_crc_api_connector", intent: "direct_write", method: http.MethodPatch, apiPath: "/v2/crc/api_connectors/{connectorId}", policy: "none"},
		{path: "crc api-connectors private-key get", operation: "zoom.get_crc_api_connector_private_key", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/crc/api_connectors/{connectorId}/private_key", policy: "json_redacted"},
		{path: "crc api-connectors private-key update", operation: "zoom.update_crc_api_connector_private_key", intent: "direct_write", method: http.MethodPatch, apiPath: "/v2/crc/api_connectors/{connectorId}/private_key", policy: "json_redacted"},
		{path: "crc managed-rooms list", operation: "zoom.list_crc_managed_rooms", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/crc/managed_rooms", policy: "json_redacted"},
		{path: "crc managed-rooms create", operation: "zoom.create_crc_managed_room", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/crc/managed_rooms", policy: "json_redacted"},
		{path: "crc managed-rooms get", operation: "zoom.get_crc_managed_room", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/crc/managed_rooms/{deviceId}", policy: "json_redacted"},
		{path: "crc managed-rooms delete", operation: "zoom.delete_crc_managed_room", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/crc/managed_rooms/{deviceId}", policy: "none"},
		{path: "crc managed-rooms update", operation: "zoom.update_crc_managed_room", intent: "direct_write", method: http.MethodPatch, apiPath: "/v2/crc/managed_rooms/{deviceId}", policy: "none"},
		{path: "crc participant-identifier-code get", operation: "zoom.get_crc_participant_identifier_code", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/crc/participant_identifier_code", policy: "json_redacted"},
		{path: "crc room-templates list", operation: "zoom.list_crc_room_templates", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/crc/room_templates", policy: "json_redacted"},
		{path: "crc room-templates create", operation: "zoom.create_crc_room_template", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/crc/room_templates", policy: "json_redacted"},
		{path: "crc room-templates get", operation: "zoom.get_crc_room_template", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/crc/room_templates/{templateId}", policy: "json_redacted"},
		{path: "crc room-templates delete", operation: "zoom.delete_crc_room_template", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/crc/room_templates/{templateId}", policy: "none"},
		{path: "crc room-templates update", operation: "zoom.update_crc_room_template", intent: "direct_write", method: http.MethodPatch, apiPath: "/v2/crc/room_templates/{templateId}", policy: "none"},
	}
	for _, want := range wants {
		if err := commandrunner.Preflight(connector, strings.Fields(want.path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want declared executable CRC action", want.path, err)
			continue
		}
		var found *connectors.CommandSurfaceCommand
		for i := range connector.CommandSurface().Commands {
			candidate := &connector.CommandSurface().Commands[i]
			if candidate.Path == want.path {
				found = candidate
				break
			}
		}
		if found == nil || found.Intent != want.intent || found.Availability != "implemented" || found.Operation != want.operation || len(found.APISurface) != 1 || found.APISurface[0].Method != want.method || found.APISurface[0].Path != want.apiPath || found.OutputPolicy != want.policy {
			t.Errorf("CRC command %q does not retain its declared operation/endpoint/output contract", want.path)
			continue
		}
		for _, flag := range found.Flags {
			if flag.Name == "page" || flag.Name == "per-page" || flag.Name == "limit" || flag.Name == "page-size" || flag.Name == "next-page-token" {
				t.Errorf("CRC command %q invents paging flag --%s", want.path, flag.Name)
			}
		}
	}
}

// TestAIServicesOperationCommandsAreReachable is the provider-category RED
// surface contract. It names all 22 operations from Zoom's own AI Services
// artifact, including the live 101 WebSocket protocol. A documented WebSocket
// operation is not an inbound event or a reason to omit the route.
func TestAIServicesOperationCommandsAreReachable(t *testing.T) {
	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	const sourceURL = "https://developers.zoom.us/docs/api/ai-services.md"
	wants := []struct {
		path        string
		operation   string
		intent      string
		method      string
		apiPath     string
		policy      string
		destructive bool
	}{
		{path: "ai-services scribe jobs list", operation: "zoom.list_ai_services_scribe_jobs", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/aiservices/scribe/jobs", policy: "json_redacted"},
		{path: "ai-services scribe jobs submit", operation: "zoom.submit_ai_services_scribe_job", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/aiservices/scribe/jobs", policy: "json_redacted"},
		{path: "ai-services scribe jobs get", operation: "zoom.get_ai_services_scribe_job", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/aiservices/scribe/jobs/{jobId}", policy: "json_redacted"},
		{path: "ai-services scribe jobs cancel", operation: "zoom.cancel_ai_services_scribe_job", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/aiservices/scribe/jobs/{jobId}", policy: "none", destructive: true},
		{path: "ai-services scribe jobs files list", operation: "zoom.list_ai_services_scribe_job_files", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/aiservices/scribe/jobs/{jobId}/files", policy: "json_redacted"},
		{path: "ai-services scribe jobs files get", operation: "zoom.get_ai_services_scribe_job_file", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/aiservices/scribe/jobs/{jobId}/files/{fileId}", policy: "json_redacted"},
		{path: "ai-services scribe live", operation: "zoom.live_ai_services_scribe", intent: "websocket_session", method: http.MethodGet, apiPath: "/v2/aiservices/scribe/live", policy: "json_redacted"},
		{path: "ai-services scribe transcribe", operation: "zoom.scribe_ai_services_transcribe", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/aiservices/scribe/transcribe", policy: "json_redacted"},
		{path: "ai-services summarizer jobs list", operation: "zoom.list_ai_services_summarizer_jobs", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/aiservices/summarizer/jobs", policy: "json_redacted"},
		{path: "ai-services summarizer jobs submit", operation: "zoom.submit_ai_services_summarizer_job", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/aiservices/summarizer/jobs", policy: "json_redacted"},
		{path: "ai-services summarizer jobs get", operation: "zoom.get_ai_services_summarizer_job", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/aiservices/summarizer/jobs/{jobId}", policy: "json_redacted"},
		{path: "ai-services summarizer jobs cancel", operation: "zoom.cancel_ai_services_summarizer_job", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/aiservices/summarizer/jobs/{jobId}", policy: "none", destructive: true},
		{path: "ai-services summarizer jobs files list", operation: "zoom.list_ai_services_summarizer_job_files", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/aiservices/summarizer/jobs/{jobId}/files", policy: "json_redacted"},
		{path: "ai-services summarizer jobs files get", operation: "zoom.get_ai_services_summarizer_job_file", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/aiservices/summarizer/jobs/{jobId}/files/{fileId}", policy: "json_redacted"},
		{path: "ai-services summarizer summarize", operation: "zoom.summarize_ai_services", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/aiservices/summarizer/summarize", policy: "json_redacted"},
		{path: "ai-services translator jobs list", operation: "zoom.list_ai_services_translator_jobs", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/aiservices/translator/jobs", policy: "json_redacted"},
		{path: "ai-services translator jobs submit", operation: "zoom.submit_ai_services_translator_job", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/aiservices/translator/jobs", policy: "json_redacted"},
		{path: "ai-services translator jobs get", operation: "zoom.get_ai_services_translator_job", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/aiservices/translator/jobs/{jobId}", policy: "json_redacted"},
		{path: "ai-services translator jobs cancel", operation: "zoom.cancel_ai_services_translator_job", intent: "direct_write", method: http.MethodDelete, apiPath: "/v2/aiservices/translator/jobs/{jobId}", policy: "none", destructive: true},
		{path: "ai-services translator jobs files list", operation: "zoom.list_ai_services_translator_job_files", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/aiservices/translator/jobs/{jobId}/files", policy: "json_redacted"},
		{path: "ai-services translator jobs files get", operation: "zoom.get_ai_services_translator_job_file", intent: "direct_read", method: http.MethodGet, apiPath: "/v2/aiservices/translator/jobs/{jobId}/files/{fileId}", policy: "json_redacted"},
		{path: "ai-services translator translate", operation: "zoom.translate_ai_services", intent: "direct_write", method: http.MethodPost, apiPath: "/v2/aiservices/translator/translate", policy: "json_redacted"},
	}

	for _, want := range wants {
		if err := commandrunner.Preflight(connector, strings.Fields(want.path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want declared executable AI Services action", want.path, err)
			continue
		}
		var command *connectors.CommandSurfaceCommand
		for i := range connector.CommandSurface().Commands {
			candidate := &connector.CommandSurface().Commands[i]
			if candidate.Path == want.path {
				command = candidate
				break
			}
		}
		if command == nil || command.Intent != want.intent || command.Availability != "implemented" || command.Operation != want.operation || command.SourceURL != sourceURL || len(command.APISurface) != 1 || command.APISurface[0].Method != want.method || command.APISurface[0].Path != want.apiPath || command.OutputPolicy != want.policy {
			t.Errorf("AI Services command %q does not retain its declared operation/endpoint/output contract", want.path)
			continue
		}
		for _, flag := range command.Flags {
			if flag.Name == "page" || flag.Name == "per-page" || flag.Name == "limit" || flag.Name == "page-size" || flag.Name == "next-page-token" {
				t.Errorf("AI Services command %q invents paging flag --%s", want.path, flag.Name)
			}
		}
		if !want.destructive {
			continue
		}
		var operation *engine.OperationSpec
		for i := range bundle.Operations {
			if bundle.Operations[i].ID == want.operation {
				operation = &bundle.Operations[i]
				break
			}
		}
		if operation == nil || operation.Confirmation == nil || operation.Confirmation.Kind != connectors.ConfirmationKindDestructive {
			t.Errorf("AI Services command %q does not retain typed destructive confirmation", want.path)
		}
	}
}

// TestAIServicesDirectReadCommandsExecuteWithFixtures runs every documented
// AI Services REST GET through the command runner. It proves the imported
// provider parameters, fixed path binding, ordinary Zoom bearer transport,
// bounded response handling, and redacted token output without contacting
// Zoom.
func TestAIServicesDirectReadCommandsExecuteWithFixtures(t *testing.T) {
	const accessToken = "fixture-ai-services-read-access-token"
	type readAction struct {
		name          string
		path          []string
		flags         map[string][]string
		requestPath   string
		expectedQuery map[string]string
	}
	actions := []readAction{
		{name: "list Scribe jobs", path: []string{"ai-services", "scribe", "jobs", "list"}, flags: map[string][]string{"state": {"SUCCEEDED"}}, requestPath: "/v2/aiservices/scribe/jobs", expectedQuery: map[string]string{"state": "SUCCEEDED"}},
		{name: "get Scribe job", path: []string{"ai-services", "scribe", "jobs", "get"}, flags: map[string][]string{"jobId": {"fixture-scribe-job"}}, requestPath: "/v2/aiservices/scribe/jobs/fixture-scribe-job"},
		{name: "list Scribe job files", path: []string{"ai-services", "scribe", "jobs", "files", "list"}, flags: map[string][]string{"jobId": {"fixture-scribe-job"}}, requestPath: "/v2/aiservices/scribe/jobs/fixture-scribe-job/files"},
		{name: "get Scribe job file", path: []string{"ai-services", "scribe", "jobs", "files", "get"}, flags: map[string][]string{"jobId": {"fixture-scribe-job"}, "fileId": {"fixture-scribe-file"}}, requestPath: "/v2/aiservices/scribe/jobs/fixture-scribe-job/files/fixture-scribe-file"},
		{name: "list Summarizer jobs", path: []string{"ai-services", "summarizer", "jobs", "list"}, flags: map[string][]string{"state": {"SUCCEEDED"}}, requestPath: "/v2/aiservices/summarizer/jobs", expectedQuery: map[string]string{"state": "SUCCEEDED"}},
		{name: "get Summarizer job", path: []string{"ai-services", "summarizer", "jobs", "get"}, flags: map[string][]string{"jobId": {"fixture-summarizer-job"}}, requestPath: "/v2/aiservices/summarizer/jobs/fixture-summarizer-job"},
		{name: "list Summarizer job files", path: []string{"ai-services", "summarizer", "jobs", "files", "list"}, flags: map[string][]string{"jobId": {"fixture-summarizer-job"}, "state": {"FILE_SUCCEEDED"}}, requestPath: "/v2/aiservices/summarizer/jobs/fixture-summarizer-job/files", expectedQuery: map[string]string{"state": "FILE_SUCCEEDED"}},
		{name: "get Summarizer job file", path: []string{"ai-services", "summarizer", "jobs", "files", "get"}, flags: map[string][]string{"jobId": {"fixture-summarizer-job"}, "fileId": {"fixture-summarizer-file"}}, requestPath: "/v2/aiservices/summarizer/jobs/fixture-summarizer-job/files/fixture-summarizer-file"},
		{name: "list Translator jobs", path: []string{"ai-services", "translator", "jobs", "list"}, flags: map[string][]string{"state": {"SUCCEEDED"}}, requestPath: "/v2/aiservices/translator/jobs", expectedQuery: map[string]string{"state": "SUCCEEDED"}},
		{name: "get Translator job", path: []string{"ai-services", "translator", "jobs", "get"}, flags: map[string][]string{"jobId": {"fixture-translator-job"}}, requestPath: "/v2/aiservices/translator/jobs/fixture-translator-job"},
		{name: "list Translator job files", path: []string{"ai-services", "translator", "jobs", "files", "list"}, flags: map[string][]string{"jobId": {"fixture-translator-job"}, "state": {"FILE_SUCCEEDED"}}, requestPath: "/v2/aiservices/translator/jobs/fixture-translator-job/files", expectedQuery: map[string]string{"state": "FILE_SUCCEEDED"}},
		{name: "get Translator job file", path: []string{"ai-services", "translator", "jobs", "files", "get"}, flags: map[string][]string{"jobId": {"fixture-translator-job"}, "fileId": {"fixture-translator-file"}}, requestPath: "/v2/aiservices/translator/jobs/fixture-translator-job/files/fixture-translator-file"},
	}
	byRequest := make(map[string]*readAction, len(actions))
	for i := range actions {
		byRequest[http.MethodGet+" "+actions[i].requestPath] = &actions[i]
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("AI Services read fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("AI Services read fixture did not receive the Zoom bearer credential")
		}
		query := request.URL.Query()
		if len(query) != len(action.expectedQuery) {
			t.Fatalf("AI Services read query = %v, want %v", query, action.expectedQuery)
		}
		for name, want := range action.expectedQuery {
			if got := query.Get(name); got != want {
				t.Fatalf("AI Services read query %s = %q, want %q", name, got, want)
			}
		}
		for _, forbidden := range []string{"page_size", "next_page_token", "page", "per_page", "limit"} {
			if got := query.Get(forbidden); got != "" {
				t.Fatalf("AI Services read sent raw paging parameter %s=%q", forbidden, got)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"state":"SUCCEEDED","next_page_token":"fixture-ai-services-read-token"}`))
	}))
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	config := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": server.URL + "/v2"},
		Secrets: map[string]string{"access_token": accessToken},
	}
	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
				Path:   action.path,
				Flags:  action.flags,
				Config: config,
			}, func(connectors.Record) error {
				t.Fatal("emit called for an AI Services direct_read command")
				return nil
			})
			if err != nil {
				t.Fatalf("Run(%q): %v", strings.Join(action.path, " "), err)
			}
			if result.DirectRead == nil || result.DirectRead.Status != http.StatusOK {
				t.Fatalf("Run(%q) direct read = %#v, want HTTP %d", strings.Join(action.path, " "), result.DirectRead, http.StatusOK)
			}
			encoded, err := json.Marshal(result.DirectRead.Body)
			if err != nil {
				t.Fatalf("marshal AI Services read response: %v", err)
			}
			if strings.Contains(string(encoded), "fixture-ai-services-read-token") || !strings.Contains(string(encoded), `"next_page_token_redacted":true`) {
				t.Fatal("AI Services read response did not redact the provider token field")
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("AI Services read fixtures received %d requests, want %d", requests, len(actions))
	}
}

// TestAIServicesDirectWriteCommandsExecuteWithFixtures exercises all nine
// documented JSON mutations through plan, no-network preview, single-use
// approval, typed destructive confirmation for cancellation, and execution.
// The three 204 cancellation responses remain completed status-only actions.
func TestAIServicesDirectWriteCommandsExecuteWithFixtures(t *testing.T) {
	const (
		credentialName = "zoom-ai-services-write-fixture"
		accessToken    = "fixture-ai-services-write-access-token"
	)
	type writeAction struct {
		name         string
		path         []string
		flags        map[string][]string
		method       string
		requestPath  string
		expectedBody json.RawMessage
		status       int
		destructive  bool
	}
	batchBody := func() json.RawMessage {
		return json.RawMessage(`{"input":{"source":"S3","uri":"s3://fixture-ai-services/input"},"output":{"destination":"S3","uri":"s3://fixture-ai-services/output"}}`)
	}
	actions := []writeAction{
		{name: "submit Scribe job", path: []string{"ai-services", "scribe", "jobs", "submit"}, flags: map[string][]string{"scribe-job": {`{"config":{"language":"en-US"},"input":{"source":"S3","uri":"s3://fixture-ai-services/input"},"output":{"destination":"S3","uri":"s3://fixture-ai-services/output"}}`}}, method: http.MethodPost, requestPath: "/v2/aiservices/scribe/jobs", expectedBody: json.RawMessage(`{"config":{"language":"en-US"},"input":{"source":"S3","uri":"s3://fixture-ai-services/input"},"output":{"destination":"S3","uri":"s3://fixture-ai-services/output"}}`), status: http.StatusOK},
		{name: "cancel Scribe job", path: []string{"ai-services", "scribe", "jobs", "cancel"}, flags: map[string][]string{"job-id": {"fixture-scribe-job"}}, method: http.MethodDelete, requestPath: "/v2/aiservices/scribe/jobs/fixture-scribe-job", status: http.StatusNoContent, destructive: true},
		{name: "scribe synchronous", path: []string{"ai-services", "scribe", "transcribe"}, flags: map[string][]string{"transcription-request": {`{"config":{"language":"en-US"},"file":"ZmFrZS1hdWRpby1ieXRlcw=="}`}}, method: http.MethodPost, requestPath: "/v2/aiservices/scribe/transcribe", expectedBody: json.RawMessage(`{"config":{"language":"en-US"},"file":"ZmFrZS1hdWRpby1ieXRlcw=="}`), status: http.StatusOK},
		{name: "submit Summarizer job", path: []string{"ai-services", "summarizer", "jobs", "submit"}, flags: map[string][]string{"summarizer-job": {string(batchBody())}}, method: http.MethodPost, requestPath: "/v2/aiservices/summarizer/jobs", expectedBody: batchBody(), status: http.StatusOK},
		{name: "cancel Summarizer job", path: []string{"ai-services", "summarizer", "jobs", "cancel"}, flags: map[string][]string{"job-id": {"fixture-summarizer-job"}}, method: http.MethodDelete, requestPath: "/v2/aiservices/summarizer/jobs/fixture-summarizer-job", status: http.StatusNoContent, destructive: true},
		{name: "summarize synchronous", path: []string{"ai-services", "summarizer", "summarize"}, flags: map[string][]string{"summary-request": {`{"input":{"text":"Synthetic AI Services test input."},"config":{"task":"summary"}}`}}, method: http.MethodPost, requestPath: "/v2/aiservices/summarizer/summarize", expectedBody: json.RawMessage(`{"input":{"text":"Synthetic AI Services test input."},"config":{"task":"summary"}}`), status: http.StatusOK},
		{name: "submit Translator job", path: []string{"ai-services", "translator", "jobs", "submit"}, flags: map[string][]string{"translator-job": {string(batchBody())}}, method: http.MethodPost, requestPath: "/v2/aiservices/translator/jobs", expectedBody: batchBody(), status: http.StatusOK},
		{name: "cancel Translator job", path: []string{"ai-services", "translator", "jobs", "cancel"}, flags: map[string][]string{"job-id": {"fixture-translator-job"}}, method: http.MethodDelete, requestPath: "/v2/aiservices/translator/jobs/fixture-translator-job", status: http.StatusNoContent, destructive: true},
		{name: "translate synchronous", path: []string{"ai-services", "translator", "translate"}, flags: map[string][]string{"translation-request": {`{"text":"Synthetic AI Services test input.","config":{"source_language":"en-US","target_languages":["es-ES"]}}`}}, method: http.MethodPost, requestPath: "/v2/aiservices/translator/translate", expectedBody: json.RawMessage(`{"text":"Synthetic AI Services test input.","config":{"source_language":"en-US","target_languages":["es-ES"]}}`), status: http.StatusOK},
	}
	byRequest := make(map[string]*writeAction, len(actions))
	for i := range actions {
		byRequest[actions[i].method+" "+actions[i].requestPath] = &actions[i]
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("AI Services write fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("AI Services write fixture did not receive the Zoom bearer credential")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("AI Services write fixture received undeclared query or paging input")
		}
		if len(action.expectedBody) == 0 {
			body, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatalf("read AI Services status-only request: %v", err)
			}
			if len(body) != 0 {
				t.Fatal("AI Services status-only action sent an undeclared request body")
			}
		} else {
			if !strings.HasPrefix(request.Header.Get("Content-Type"), "application/json") {
				t.Fatal("AI Services JSON action did not declare JSON content")
			}
			var got, want any
			if err := json.NewDecoder(request.Body).Decode(&got); err != nil {
				t.Fatalf("decode AI Services request body: %v", err)
			}
			if err := json.Unmarshal(action.expectedBody, &want); err != nil {
				t.Fatalf("decode expected AI Services request body: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("AI Services request body did not contain exactly the declared documented object")
			}
		}
		if action.status != http.StatusNoContent {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(action.status)
		if action.status != http.StatusNoContent {
			_, _ = w.Write([]byte(`{"token":"fixture-ai-services-write-response-token"}`))
		}
	}))
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	for _, action := range actions {
		if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{Path: action.path, Flags: action.flags}); err != nil {
			t.Fatalf("BuildWriteCommand(%s): %v", action.name, err)
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
		Config:    map[string]string{"base_url": server.URL + "/v2"},
		Secrets:   map[string]string{"access_token": accessToken},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			beforeRequests := requests
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
				t.Fatal("AI Services write plan did not produce a no-network preview and single-use approval")
			}
			if requests != beforeRequests {
				t.Fatal("AI Services write plan or preview reached the fixture endpoint")
			}

			runRequest := app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}
			if action.destructive {
				if _, err := application.RunReverseETL(context.Background(), runRequest); err == nil {
					t.Fatal("AI Services destructive action bypassed its typed confirmation gate")
				}
				if requests != beforeRequests {
					t.Fatal("unconfirmed AI Services destructive action reached the fixture endpoint")
				}
				runRequest.Confirmation = connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
			}
			run, err := application.RunReverseETL(context.Background(), runRequest)
			if err != nil {
				t.Fatalf("RunReverseETL(%s): %v", action.name, err)
			}
			if run.Status != "completed" || run.OperationDirectWrite == nil || run.OperationDirectWrite.Status != action.status {
				t.Fatalf("AI Services write run = %#v, want completed declared action status %d", run, action.status)
			}
			if action.status == http.StatusNoContent {
				if run.OperationDirectWrite.Body != nil {
					t.Fatal("AI Services status-only action returned an invented response body")
				}
				return
			}
			encoded, err := json.Marshal(run.OperationDirectWrite.Body)
			if err != nil {
				t.Fatalf("marshal AI Services write response: %v", err)
			}
			if strings.Contains(string(encoded), "fixture-ai-services-write-response-token") || !strings.Contains(string(encoded), `"token_redacted":true`) {
				t.Fatal("AI Services write response did not redact the provider token field")
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("AI Services write fixtures received %d requests, want %d", requests, len(actions))
	}
}

// TestClipsStatusOnlyDirectWritesExecuteWithFixtures exercises every documented
// Clips 204 mutation through the real plan, preview, typed confirmation, and
// execution lifecycle. A no-content provider response must remain a completed
// action with its declared status, never become a fake body or a read route.
func TestClipsStatusOnlyDirectWritesExecuteWithFixtures(t *testing.T) {
	const (
		credentialName = "zoom-clips-status-fixture"
		accessToken    = "fixture-clips-status-access-token"
	)
	type clipsStatusAction struct {
		name         string
		path         []string
		flags        map[string][]string
		method       string
		requestPath  string
		expectedBody json.RawMessage
		destructive  bool
	}
	actions := []clipsStatusAction{
		{
			name:         "set starred",
			path:         []string{"clips", "starred", "set"},
			flags:        map[string][]string{"clip-ids": {"fixture-clip"}, "starred": {"true"}},
			method:       http.MethodPost,
			requestPath:  "/v2/clips/starred",
			expectedBody: json.RawMessage(`{"clip_ids":["fixture-clip"],"starred":true}`),
		},
		{
			name:        "remove collaborator",
			path:        []string{"clips", "collaborators", "remove"},
			flags:       map[string][]string{"clip-id": {"fixture-clip"}},
			method:      http.MethodDelete,
			requestPath: "/v2/clips/fixture-clip/collaborators",
			destructive: true,
		},
		{
			name:        "delete comment",
			path:        []string{"clips", "comments", "delete"},
			flags:       map[string][]string{"clip-id": {"fixture-clip"}, "comment-id": {"fixture-comment"}},
			method:      http.MethodDelete,
			requestPath: "/v2/clips/fixture-clip/comments/fixture-comment",
			destructive: true,
		},
		{
			name:        "delete clip",
			path:        []string{"clips", "delete"},
			flags:       map[string][]string{"clip-id": {"fixture-clip"}},
			method:      http.MethodDelete,
			requestPath: "/v2/clips/fixture-clip",
			destructive: true,
		},
		{
			name:         "update clip",
			path:         []string{"clips", "update"},
			flags:        map[string][]string{"clip-id": {"fixture-clip"}, "clip": {`{"title":"Fixture Clip title"}`}},
			method:       http.MethodPatch,
			requestPath:  "/v2/clips/fixture-clip",
			expectedBody: json.RawMessage(`{"title":"Fixture Clip title"}`),
		},
		{
			name:         "update share settings",
			path:         []string{"clips", "share-settings", "update"},
			flags:        map[string][]string{"clip-id": {"fixture-clip"}, "share-scope": {"PRIVATE"}},
			method:       http.MethodPatch,
			requestPath:  "/v2/clips/fixture-clip/share_settings",
			expectedBody: json.RawMessage(`{"share_scope":"PRIVATE"}`),
		},
	}
	byRequest := make(map[string]*clipsStatusAction, len(actions))
	for i := range actions {
		byRequest[actions[i].method+" "+actions[i].requestPath] = &actions[i]
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("Clips status fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("Clips status fixture did not receive the Zoom bearer credential")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("Clips status fixture received undeclared query input")
		}
		if len(action.expectedBody) == 0 {
			body, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatalf("read Clips status-only request body: %v", err)
			}
			if len(body) != 0 {
				t.Fatal("Clips status-only action sent an undeclared request body")
			}
		} else {
			if !strings.HasPrefix(request.Header.Get("Content-Type"), "application/json") {
				t.Fatal("Clips status action body did not declare JSON content")
			}
			var got, want any
			if err := json.NewDecoder(request.Body).Decode(&got); err != nil {
				t.Fatalf("decode Clips status action body: %v", err)
			}
			if err := json.Unmarshal(action.expectedBody, &want); err != nil {
				t.Fatalf("decode expected Clips status action body: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("Clips status action body did not contain exactly the declared fields")
			}
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	for _, action := range actions {
		if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{Path: action.path, Flags: action.flags}); err != nil {
			t.Fatalf("BuildWriteCommand(%s): %v", action.name, err)
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
		Config:    map[string]string{"base_url": server.URL + "/v2"},
		Secrets:   map[string]string{"access_token": accessToken},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			beforeRequests := requests
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
				t.Fatal("Clips status action plan did not produce a no-network preview and single-use approval")
			}
			if action.destructive != (plan.ConfirmationChallenge == string(connectors.ConfirmationKindDestructive)) {
				t.Fatal("Clips status action did not retain its declared destructive confirmation policy")
			}
			if requests != beforeRequests {
				t.Fatal("Clips status action plan or preview reached the fixture endpoint")
			}

			runRequest := app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}
			if action.destructive {
				if _, err := application.RunReverseETL(context.Background(), runRequest); err == nil {
					t.Fatal("Clips destructive status action bypassed typed destructive confirmation")
				}
				if requests != beforeRequests {
					t.Fatal("unconfirmed Clips destructive status action reached the fixture endpoint")
				}
				runRequest.Confirmation = connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
			}
			run, err := application.RunReverseETL(context.Background(), runRequest)
			if err != nil {
				t.Fatalf("RunReverseETL(%s): %v", action.name, err)
			}
			if run.Status != "completed" || run.OperationDirectWrite == nil || run.OperationDirectWrite.Status != http.StatusNoContent {
				t.Fatalf("Clips status action run = %#v, want completed HTTP %d", run, http.StatusNoContent)
			}
			if run.OperationDirectWrite.Body != nil {
				t.Fatal("Clips status-only action returned an invented response body")
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("Clips status fixtures received %d requests, want %d", requests, len(actions))
	}
}

// TestCRCDirectReadCommandsExecuteWithFixtures runs every documented CRC GET
// through the real command runner. The loopback server proves each provider
// path, bearer transport, derived camelCase path binding, and declared output
// redaction without ever contacting Zoom.
func TestCRCDirectReadCommandsExecuteWithFixtures(t *testing.T) {
	const accessToken = "fixture-crc-read-access-token"
	type crcReadAction struct {
		name         string
		path         []string
		flags        map[string][]string
		requestPath  string
		redactFields []string
		response     json.RawMessage
	}
	actions := []crcReadAction{
		{
			name:        "get managed-room account setting",
			path:        []string{"crc", "managed-rooms", "account-setting", "get"},
			requestPath: "/v2/crc/managed_rooms/account_setting",
			redactFields: []string{
				"account_id", "connector_id", "connector_log_level", "firmware_key", "ip_address", "time_server", "time_zone",
			},
		},
		{
			name:        "list API connectors",
			path:        []string{"crc", "api-connectors", "list"},
			requestPath: "/v2/crc/api_connectors",
			redactFields: []string{
				"connector_id", "description", "managed_rooms", "networks", "private_key",
			},
		},
		{
			name:         "get API connector",
			path:         []string{"crc", "api-connectors", "get"},
			flags:        map[string][]string{"connector-id": {"fixture-crc-connector"}},
			requestPath:  "/v2/crc/api_connectors/fixture-crc-connector",
			redactFields: []string{"connector_id", "description", "managed_rooms", "networks", "private_key"},
		},
		{
			name:         "get API connector private key",
			path:         []string{"crc", "api-connectors", "private-key", "get"},
			flags:        map[string][]string{"connector-id": {"fixture-crc-connector"}},
			requestPath:  "/v2/crc/api_connectors/fixture-crc-connector/private_key",
			redactFields: []string{"connector_id", "private_key"},
		},
		{
			name:        "list managed rooms",
			path:        []string{"crc", "managed-rooms", "list"},
			requestPath: "/v2/crc/managed_rooms",
			redactFields: []string{
				"auth_password", "auth_username", "connector_id", "device_id", "device_password", "device_username", "firmware_key", "ip_address", "name", "room_host_key", "serial_number", "template_id",
			},
		},
		{
			name:         "get managed room",
			path:         []string{"crc", "managed-rooms", "get"},
			flags:        map[string][]string{"device-id": {"fixture-crc-device"}},
			requestPath:  "/v2/crc/managed_rooms/fixture-crc-device",
			redactFields: []string{"auth_password", "auth_username", "connector_id", "device_id", "device_password", "device_username", "firmware_key", "ip_address", "name", "room_host_key", "serial_number", "template_id"},
		},
		{
			name:         "get participant identifier code",
			path:         []string{"crc", "participant-identifier-code", "get"},
			requestPath:  "/v2/crc/participant_identifier_code",
			redactFields: []string{"meeting_id", "participant_identifier_code", "participant_id"},
		},
		{
			name:         "list room templates",
			path:         []string{"crc", "room-templates", "list"},
			requestPath:  "/v2/crc/room_templates",
			redactFields: []string{"device_type", "name", "template_id", "time_server", "time_zone"},
		},
		{
			name:         "get room template",
			path:         []string{"crc", "room-templates", "get"},
			flags:        map[string][]string{"template-id": {"fixture-crc-template"}},
			requestPath:  "/v2/crc/room_templates/fixture-crc-template",
			redactFields: []string{"device_type", "name", "template_id", "time_server", "time_zone"},
		},
	}
	for i := range actions {
		fixture := map[string]any{"visible": "fixture-crc-visible"}
		for _, field := range actions[i].redactFields {
			fixture[field] = "fixture-crc-sensitive-value"
		}
		response, err := json.Marshal(fixture)
		if err != nil {
			t.Fatalf("marshal CRC read fixture %s: %v", actions[i].name, err)
		}
		actions[i].response = response
	}
	byRequest := make(map[string]*crcReadAction, len(actions))
	for i := range actions {
		byRequest[http.MethodGet+" "+actions[i].requestPath] = &actions[i]
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("CRC read fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("CRC read fixture did not receive the Zoom bearer credential")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("CRC read fixture received undeclared query or paging input")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(action.response)
	}))
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	config := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": server.URL + "/v2"},
		Secrets: map[string]string{"access_token": accessToken},
	}
	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
				Path:   action.path,
				Flags:  action.flags,
				Config: config,
			}, func(connectors.Record) error {
				t.Fatal("emit called for a CRC direct_read command")
				return nil
			})
			if err != nil {
				t.Fatalf("Run(%q): %v", strings.Join(action.path, " "), err)
			}
			if result.DirectRead == nil || result.DirectRead.Status != http.StatusOK {
				t.Fatalf("Run(%q) direct read = %#v, want HTTP %d", strings.Join(action.path, " "), result.DirectRead, http.StatusOK)
			}
			encoded, err := json.Marshal(result.DirectRead.Body)
			if err != nil {
				t.Fatalf("marshal CRC read response: %v", err)
			}
			if strings.Contains(string(encoded), "fixture-crc-sensitive-value") {
				t.Fatal("CRC read response exposed a declared sensitive field")
			}
			for _, field := range action.redactFields {
				if !strings.Contains(string(encoded), "\""+field+"_redacted\":true") {
					t.Fatalf("CRC read response is missing %s_redacted marker", field)
				}
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("CRC read fixtures received %d requests, want %d", requests, len(actions))
	}
}

// TestCRCDirectWriteCommandsExecuteWithFixtures sends every documented CRC
// mutation through plan, no-network preview, single-use approval, and execute.
// It keeps 204 actions status-only, requires typed confirmation for deletes
// and the secret-returning key regeneration, and proves redacted responses.
func TestCRCDirectWriteCommandsExecuteWithFixtures(t *testing.T) {
	const (
		credentialName = "zoom-crc-write-fixture"
		accessToken    = "fixture-crc-write-access-token"
	)
	type crcWriteAction struct {
		name              string
		path              []string
		flags             map[string][]string
		method            string
		requestPath       string
		expectedBody      json.RawMessage
		multipartFields   map[string]string
		status            int
		response          json.RawMessage
		destructive       bool
		secretSensitive   bool
		statusOnly        bool
		responseSensitive []string
		responseMarkers   []string
	}
	actions := []crcWriteAction{
		{
			name:            "update managed-room account setting",
			path:            []string{"crc", "managed-rooms", "account-setting", "update"},
			flags:           map[string][]string{"account-settings": {`{"enable_1080p":true}`}},
			method:          http.MethodPatch,
			requestPath:     "/v2/crc/managed_rooms/account_setting",
			multipartFields: map[string]string{"enable_1080p": "true"},
			status:          http.StatusNoContent,
			statusOnly:      true,
		},
		{
			name:              "create API connector",
			path:              []string{"crc", "api-connectors", "create"},
			flags:             map[string][]string{"api-connector": {`{"description":"Fixture CRC API connector","log_level":"INFO"}`}},
			method:            http.MethodPost,
			requestPath:       "/v2/crc/api_connectors",
			expectedBody:      json.RawMessage(`{"description":"Fixture CRC API connector","log_level":"INFO"}`),
			status:            http.StatusCreated,
			response:          json.RawMessage(`{"connector_id":"fixture-crc-sensitive-response"}`),
			responseSensitive: []string{"fixture-crc-sensitive-response"},
			responseMarkers:   []string{"connector_id"},
		},
		{
			name:        "delete API connector",
			path:        []string{"crc", "api-connectors", "delete"},
			flags:       map[string][]string{"connector-id": {"fixture-crc-connector"}},
			method:      http.MethodDelete,
			requestPath: "/v2/crc/api_connectors/fixture-crc-connector",
			status:      http.StatusNoContent,
			destructive: true,
			statusOnly:  true,
		},
		{
			name:         "update API connector",
			path:         []string{"crc", "api-connectors", "update"},
			flags:        map[string][]string{"connector-id": {"fixture-crc-connector"}, "api-connector": {`{"log_level":"INFO"}`}},
			method:       http.MethodPatch,
			requestPath:  "/v2/crc/api_connectors/fixture-crc-connector",
			expectedBody: json.RawMessage(`{"log_level":"INFO"}`),
			status:       http.StatusNoContent,
			statusOnly:   true,
		},
		{
			name:              "regenerate API connector private key",
			path:              []string{"crc", "api-connectors", "private-key", "update"},
			flags:             map[string][]string{"connector-id": {"fixture-crc-connector"}},
			method:            http.MethodPatch,
			requestPath:       "/v2/crc/api_connectors/fixture-crc-connector/private_key",
			status:            http.StatusOK,
			response:          json.RawMessage(`{"connector_id":"fixture-crc-sensitive-response","private_key":"fixture-crc-private-key-response"}`),
			secretSensitive:   true,
			responseSensitive: []string{"fixture-crc-sensitive-response", "fixture-crc-private-key-response"},
			responseMarkers:   []string{"connector_id", "private_key"},
		},
		{
			name:              "create managed room",
			path:              []string{"crc", "managed-rooms", "create"},
			flags:             map[string][]string{"managed-room": {`{"device_type":"Cisco Series","name":"Fixture CRC managed room"}`}},
			method:            http.MethodPost,
			requestPath:       "/v2/crc/managed_rooms",
			expectedBody:      json.RawMessage(`{"device_type":"Cisco Series","name":"Fixture CRC managed room"}`),
			status:            http.StatusCreated,
			response:          json.RawMessage(`{"device_id":"fixture-crc-sensitive-response","auth_username":"fixture-crc-sensitive-response"}`),
			responseSensitive: []string{"fixture-crc-sensitive-response"},
			responseMarkers:   []string{"device_id", "auth_username"},
		},
		{
			name:        "delete managed room",
			path:        []string{"crc", "managed-rooms", "delete"},
			flags:       map[string][]string{"device-id": {"fixture-crc-device"}},
			method:      http.MethodDelete,
			requestPath: "/v2/crc/managed_rooms/fixture-crc-device",
			status:      http.StatusNoContent,
			destructive: true,
			statusOnly:  true,
		},
		{
			name:         "update managed room",
			path:         []string{"crc", "managed-rooms", "update"},
			flags:        map[string][]string{"device-id": {"fixture-crc-device"}, "managed-room": {`{"enable_1080p":true}`}},
			method:       http.MethodPatch,
			requestPath:  "/v2/crc/managed_rooms/fixture-crc-device",
			expectedBody: json.RawMessage(`{"enable_1080p":true}`),
			status:       http.StatusNoContent,
			statusOnly:   true,
		},
		{
			name:              "create room template",
			path:              []string{"crc", "room-templates", "create"},
			flags:             map[string][]string{"room-template": {`{"device_type":"Cisco Series","name":"Fixture CRC room template"}`}},
			method:            http.MethodPost,
			requestPath:       "/v2/crc/room_templates",
			expectedBody:      json.RawMessage(`{"device_type":"Cisco Series","name":"Fixture CRC room template"}`),
			status:            http.StatusCreated,
			response:          json.RawMessage(`{"template_id":"fixture-crc-sensitive-response"}`),
			responseSensitive: []string{"fixture-crc-sensitive-response"},
			responseMarkers:   []string{"template_id"},
		},
		{
			name:        "delete room template",
			path:        []string{"crc", "room-templates", "delete"},
			flags:       map[string][]string{"template-id": {"fixture-crc-template"}},
			method:      http.MethodDelete,
			requestPath: "/v2/crc/room_templates/fixture-crc-template",
			status:      http.StatusNoContent,
			destructive: true,
			statusOnly:  true,
		},
		{
			name:         "update room template",
			path:         []string{"crc", "room-templates", "update"},
			flags:        map[string][]string{"template-id": {"fixture-crc-template"}, "room-template": {`{"name":"Fixture CRC room template"}`}},
			method:       http.MethodPatch,
			requestPath:  "/v2/crc/room_templates/fixture-crc-template",
			expectedBody: json.RawMessage(`{"name":"Fixture CRC room template"}`),
			status:       http.StatusNoContent,
			statusOnly:   true,
		},
	}
	byRequest := make(map[string]*crcWriteAction, len(actions))
	for i := range actions {
		byRequest[actions[i].method+" "+actions[i].requestPath] = &actions[i]
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("CRC write fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("CRC write fixture did not receive the Zoom bearer credential")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("CRC write fixture received undeclared query or paging input")
		}
		switch {
		case len(action.multipartFields) > 0:
			assertCRCMultipartFields(t, request, action.multipartFields)
		case len(action.expectedBody) == 0:
			body, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatalf("read CRC no-body request: %v", err)
			}
			if len(body) != 0 {
				t.Fatal("CRC no-body action sent an undeclared request body")
			}
		default:
			if !strings.HasPrefix(request.Header.Get("Content-Type"), "application/json") {
				t.Fatal("CRC body action did not declare JSON content")
			}
			var got, want any
			if err := json.NewDecoder(request.Body).Decode(&got); err != nil {
				t.Fatalf("decode CRC action body: %v", err)
			}
			if err := json.Unmarshal(action.expectedBody, &want); err != nil {
				t.Fatalf("decode expected CRC action body: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("CRC action body did not contain exactly the declared documented fields")
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
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	for _, action := range actions {
		if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{Path: action.path, Flags: action.flags}); err != nil {
			t.Fatalf("BuildWriteCommand(%s): %v", action.name, err)
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
		Config:    map[string]string{"base_url": server.URL + "/v2"},
		Secrets:   map[string]string{"access_token": accessToken},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			beforeRequests := requests
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
				t.Fatal("CRC write plan did not produce a no-network preview and single-use approval")
			}
			requiresTypedConfirmation := action.destructive || action.secretSensitive
			if requiresTypedConfirmation != (plan.ConfirmationChallenge == string(connectors.ConfirmationKindDestructive)) {
				t.Fatal("CRC write action did not retain its declared typed-confirmation policy")
			}
			if requests != beforeRequests {
				t.Fatal("CRC write plan or preview reached the fixture endpoint")
			}

			runRequest := app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}
			if requiresTypedConfirmation {
				if _, err := application.RunReverseETL(context.Background(), runRequest); err == nil {
					t.Fatal("CRC typed-confirmation action bypassed its confirmation gate")
				}
				if requests != beforeRequests {
					t.Fatal("unconfirmed CRC action reached the fixture endpoint")
				}
				runRequest.Confirmation = connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
			}
			run, err := application.RunReverseETL(context.Background(), runRequest)
			if err != nil {
				t.Fatalf("RunReverseETL(%s): %v", action.name, err)
			}
			if run.Status != "completed" || run.OperationDirectWrite == nil || run.OperationDirectWrite.Status != action.status {
				t.Fatalf("CRC write run = %#v, want completed declared action status %d", run, action.status)
			}
			if action.statusOnly {
				if run.OperationDirectWrite.Body != nil {
					t.Fatal("CRC 204 status-only action returned an invented response body")
				}
				return
			}
			encoded, err := json.Marshal(run.OperationDirectWrite.Body)
			if err != nil {
				t.Fatalf("marshal CRC write response: %v", err)
			}
			for _, raw := range action.responseSensitive {
				if strings.Contains(string(encoded), raw) {
					t.Fatal("CRC write response exposed a declared sensitive field")
				}
			}
			for _, field := range action.responseMarkers {
				if !strings.Contains(string(encoded), "\""+field+"_redacted\":true") {
					t.Fatalf("CRC write response is missing %s_redacted marker", field)
				}
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("CRC write fixtures received %d requests, want %d", requests, len(actions))
	}
}

func assertCRCMultipartFields(t *testing.T, request *http.Request, wantFields map[string]string) {
	t.Helper()
	mediaType, params, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" || params["boundary"] == "" {
		t.Fatalf("CRC multipart content type = %q, want multipart/form-data with boundary", request.Header.Get("Content-Type"))
	}
	reader := multipart.NewReader(request.Body, params["boundary"])
	gotFields := make(map[string]string, len(wantFields))
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read CRC multipart part: %v", err)
		}
		if part.FileName() != "" {
			_ = part.Close()
			t.Fatal("CRC account-setting multipart request contains an undeclared file part")
		}
		content, err := io.ReadAll(part)
		_ = part.Close()
		if err != nil {
			t.Fatalf("read CRC multipart field: %v", err)
		}
		gotFields[part.FormName()] = string(content)
	}
	if !reflect.DeepEqual(gotFields, wantFields) {
		t.Fatalf("CRC multipart fields = %#v, want %#v", gotFields, wantFields)
	}
}

// TestAutoDialerDirectReadCommandsExecuteWithFixtures runs every published Auto
// Dialer GET through the real command runner. It proves the ordinary Zoom
// bearer, each exact fixed endpoint, no undeclared query or paging input, and
// named sensitive-field redaction without reaching the provider.
func TestAutoDialerDirectReadCommandsExecuteWithFixtures(t *testing.T) {
	const accessToken = "fixture-auto-dialer-access-token"
	type autoDialerReadAction struct {
		name              string
		path              []string
		flags             map[string][]string
		requestPath       string
		fixture           string
		status            int
		response          json.RawMessage
		responseSensitive []string
		responseMarkers   []string
	}
	actions := []autoDialerReadAction{
		{
			name:              "get call history",
			path:              []string{"auto-dialer", "call-histories", "get"},
			flags:             map[string][]string{"call-history-id": {"fixture-ad-call-history"}},
			requestPath:       "/v2/dialer/call-histories/fixture-ad-call-history",
			fixture:           "get_auto_dialer_call_history.json",
			responseSensitive: []string{"fixture-ad-call-history", "fixture-ad-call", "fixture-ad-prospect", "fixture Auto Dialer private call note", "fixture.ad@example.invalid", "fixture-ad-response-token"},
			responseMarkers:   []string{"call_history_id", "call_id", "prospect_id", "call_note", "email_addresses", "token"},
		},
		{
			name:              "list call history",
			path:              []string{"auto-dialer", "call-history", "list"},
			requestPath:       "/v2/dialer/call-history",
			fixture:           "list_auto_dialer_call_history.json",
			responseSensitive: []string{"fixture-ad-call-history", "fixture-ad-call", "fixture Auto Dialer history note", "fixture-ad-page-token", "fixture-ad-response-token"},
			responseMarkers:   []string{"call_history", "next_page_token", "token"},
		},
		{
			name:              "list report call history",
			path:              []string{"auto-dialer", "reports", "call-history", "list"},
			requestPath:       "/v2/dialer/reports/call-history",
			fixture:           "list_auto_dialer_report_call_history.json",
			responseSensitive: []string{"fixture-ad-report-call-history", "Fixture Auto Dialer seller", "+15550000000", "fixture-ad-page-token", "fixture-ad-response-token"},
			responseMarkers:   []string{"call_history", "next_page_token", "token"},
		},
		{
			name:              "get seller productivity report",
			path:              []string{"auto-dialer", "reports", "seller-productivity", "get"},
			requestPath:       "/v2/dialer/reports/seller-productivity",
			fixture:           "get_auto_dialer_seller_productivity_report.json",
			responseSensitive: []string{"fixture-ad-seller", "Fixture Auto Dialer seller", "fixture.seller@example.invalid", "fixture-ad-response-token"},
			responseMarkers:   []string{"seller_id", "seller_name", "email", "token"},
		},
		{
			name:              "list call lists",
			path:              []string{"auto-dialer", "call-lists", "list"},
			requestPath:       "/v2/dialer/call-lists",
			fixture:           "list_auto_dialer_call_lists.json",
			responseSensitive: []string{"fixture-ad-call-list", "fixture-ad-user", "fixture.user@example.invalid", "Fixture Auto Dialer call list", "fixture call-list description", "fixture-ad-page-token", "fixture-ad-response-token"},
			responseMarkers:   []string{"call_lists", "next_page_token", "token"},
		},
		{
			name:              "get call list",
			path:              []string{"auto-dialer", "call-lists", "get"},
			flags:             map[string][]string{"call-list-id": {"fixture-ad-call-list"}},
			requestPath:       "/v2/dialer/call-lists/fixture-ad-call-list",
			fixture:           "get_auto_dialer_call_list.json",
			responseSensitive: []string{"fixture-ad-call-list", "fixture-ad-user", "fixture.user@example.invalid", "Fixture Auto Dialer call list", "fixture call-list description", "fixture-ad-response-token"},
			responseMarkers:   []string{"call_list_id", "assigned_to_user_id", "assigned_to_user_email", "name", "description", "token"},
		},
		{
			name:              "list call-list prospects",
			path:              []string{"auto-dialer", "call-lists", "prospects", "list"},
			flags:             map[string][]string{"call-list-id": {"fixture-ad-call-list"}},
			requestPath:       "/v2/dialer/call-lists/fixture-ad-call-list/prospects",
			fixture:           "list_auto_dialer_call_list_prospects.json",
			responseSensitive: []string{"fixture-ad-prospect", "Fixture Auto Dialer prospect", "+15550000000", "fixture-ad-page-token", "fixture-ad-response-token"},
			responseMarkers:   []string{"prospects", "next_page_token", "token"},
		},
		{
			name:              "get prospect",
			path:              []string{"auto-dialer", "prospects", "get"},
			flags:             map[string][]string{"prospect-id": {"fixture-ad-prospect"}},
			requestPath:       "/v2/dialer/prospects/fixture-ad-prospect",
			fixture:           "get_auto_dialer_prospect.json",
			responseSensitive: []string{"fixture-ad-prospect", "fixture-ad-user", "fixture-ad-call-list", "Fixture Auto Dialer prospect", "Fixture Auto Dialer company", "fixture.prospect@example.invalid", "+15550000000", "fixture-ad-response-token"},
			responseMarkers:   []string{"prospect_id", "assignee_user_id", "call_list_id", "primary_name", "company", "email_addresses", "phone_numbers", "token"},
		},
	}
	byRequest := make(map[string]*autoDialerReadAction, len(actions))
	for i := range actions {
		actions[i].status, actions[i].response = zoomDirectReadFixture(t, actions[i].fixture)
		byRequest[http.MethodGet+" "+actions[i].requestPath] = &actions[i]
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("Auto Dialer read fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("Auto Dialer read fixture did not receive the Zoom bearer credential")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("Auto Dialer read fixture received undeclared query or paging input")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(action.status)
		_, _ = w.Write(action.response)
	}))
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	config := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": server.URL + "/v2"},
		Secrets: map[string]string{"access_token": accessToken},
	}
	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
				Path:   action.path,
				Flags:  action.flags,
				Config: config,
			}, func(connectors.Record) error {
				t.Fatal("emit called for an Auto Dialer direct_read command")
				return nil
			})
			if err != nil {
				t.Fatalf("Run(%q): %v", strings.Join(action.path, " "), err)
			}
			if result.DirectRead == nil || result.DirectRead.Status != action.status {
				t.Fatalf("Run(%q) result = %#v, want status %d", strings.Join(action.path, " "), result.DirectRead, action.status)
			}
			encoded, err := json.Marshal(result.DirectRead.Body)
			if err != nil {
				t.Fatalf("marshal Auto Dialer read response: %v", err)
			}
			for index, raw := range action.responseSensitive {
				if strings.Contains(string(encoded), raw) {
					t.Fatalf("Auto Dialer read response exposed declared or generic sensitive field %d", index)
				}
			}
			for _, field := range action.responseMarkers {
				if !strings.Contains(string(encoded), "\""+field+"_redacted\":true") {
					t.Fatalf("Auto Dialer read response is missing %s_redacted marker", field)
				}
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("Auto Dialer read fixtures received %d requests, want %d", requests, len(actions))
	}
}

// TestAutoDialerDirectWriteCommandsExecuteWithFixtures runs every published
// Auto Dialer mutation through plan, no-network preview, single-use approval,
// and execute. It proves exact fixed methods/body schemas/auth/status, typed
// destructive confirmation, and 204 status-only semantics without Zoom.
func TestAutoDialerDirectWriteCommandsExecuteWithFixtures(t *testing.T) {
	const (
		credentialName = "zoom-auto-dialer-fixture"
		accessToken    = "fixture-auto-dialer-access-token"
	)
	type autoDialerWriteAction struct {
		name              string
		path              []string
		flags             map[string][]string
		rootFlag          string
		fixture           string
		method            string
		requestPath       string
		expectedBody      json.RawMessage
		status            int
		response          json.RawMessage
		destructive       bool
		inputSensitive    []string
		responseSensitive []string
		responseMarkers   []string
	}
	actions := []autoDialerWriteAction{
		{
			name: "create call list",
			path: []string{"auto-dialer", "call-lists", "create"},
			flags: map[string][]string{
				"assigned-to-user-id": {"fixture-ad-user"},
				"name":                {"Fixture Auto Dialer call list"},
				"prospect-type":       {"CONTACT"},
				"description":         {"fixture call-list description"},
			},
			fixture:           "create_auto_dialer_call_list.json",
			inputSensitive:    []string{"fixture-ad-user", "Fixture Auto Dialer call list", "fixture call-list description"},
			responseSensitive: []string{"fixture-ad-call-list", "fixture-ad-user", "Fixture Auto Dialer call list", "fixture-ad-response-token"},
			responseMarkers:   []string{"call_list_id", "assigned_to_user_id", "name", "token"},
		},
		{
			name:        "delete call list",
			path:        []string{"auto-dialer", "call-lists", "delete"},
			flags:       map[string][]string{"call-list-id": {"fixture-ad-call-list"}},
			fixture:     "delete_auto_dialer_call_list.json",
			destructive: true,
		},
		{
			name:           "update call list",
			path:           []string{"auto-dialer", "call-lists", "update"},
			flags:          map[string][]string{"call-list-id": {"fixture-ad-call-list"}},
			rootFlag:       "call-list",
			fixture:        "update_auto_dialer_call_list.json",
			inputSensitive: []string{"Updated Fixture Auto Dialer call list", "updated fixture call-list description"},
		},
		{
			name:              "create prospect",
			path:              []string{"auto-dialer", "call-lists", "prospects", "create"},
			flags:             map[string][]string{"call-list-id": {"fixture-ad-call-list"}},
			rootFlag:          "prospect",
			fixture:           "create_auto_dialer_prospect.json",
			inputSensitive:    []string{"Fixture Auto Dialer prospect", "Fixture Auto Dialer company", "fixture custom field", "+15550000000", "fixture.prospect@example.invalid", "fixture-ad-external-id"},
			responseSensitive: []string{"fixture-ad-prospect", "Fixture Auto Dialer prospect", "+15550000000", "fixture-ad-response-token"},
			responseMarkers:   []string{"prospect_id", "primary_name", "primary_phone", "token"},
		},
		{
			name:              "update prospects batch",
			path:              []string{"auto-dialer", "call-lists", "prospects", "update-batch"},
			flags:             map[string][]string{"call-list-id": {"fixture-ad-call-list"}},
			rootFlag:          "request",
			fixture:           "update_auto_dialer_prospects_batch.json",
			inputSensitive:    []string{"fixture-ad-prospect", "Updated Fixture Auto Dialer prospect", "+15550000001", "updated.fixture.prospect@example.invalid", "fixture-ad-other-call-list"},
			responseSensitive: []string{"fixture-ad-prospect", "fixture-ad-failed-prospect", "fixture validation detail", "fixture-ad-user", "fixture.user@example.invalid", "fixture-ad-response-token"},
			responseMarkers:   []string{"prospect_id", "error_message", "assigned_to_user_id", "assigned_to_user_email", "token"},
		},
		{
			name:              "create prospects batch",
			path:              []string{"auto-dialer", "call-lists", "prospects", "create-batch"},
			flags:             map[string][]string{"call-list-id": {"fixture-ad-call-list"}},
			rootFlag:          "request",
			fixture:           "create_auto_dialer_prospects_batch.json",
			inputSensitive:    []string{"Fixture batch Auto Dialer prospect", "Fixture Auto Dialer company", "+15550000002", "batch.fixture.prospect@example.invalid"},
			responseSensitive: []string{"fixture-ad-batch-prospect", "Fixture batch Auto Dialer prospect", "+15550000002", "fixture-ad-user", "fixture.user@example.invalid", "fixture-ad-response-token"},
			responseMarkers:   []string{"prospect_id", "name", "phone", "assigned_to_user_id", "assigned_to_user_email", "token"},
		},
		{
			name:        "delete prospect",
			path:        []string{"auto-dialer", "call-lists", "prospects", "delete"},
			flags:       map[string][]string{"call-list-id": {"fixture-ad-call-list"}, "prospect-id": {"fixture-ad-prospect"}},
			fixture:     "delete_auto_dialer_prospect.json",
			destructive: true,
		},
		{
			name:           "update prospect",
			path:           []string{"auto-dialer", "call-lists", "prospects", "update"},
			flags:          map[string][]string{"call-list-id": {"fixture-ad-call-list"}, "prospect-id": {"fixture-ad-prospect"}},
			rootFlag:       "prospect",
			fixture:        "update_auto_dialer_prospect.json",
			inputSensitive: []string{"Updated Fixture Auto Dialer prospect", "Fixture sales title"},
		},
	}
	byRequest := make(map[string]*autoDialerWriteAction, len(actions))
	for i := range actions {
		fixture := zoomSCIM2WriteFixture(t, actions[i].fixture)
		actions[i].method = fixture.Expect.Method
		actions[i].requestPath = fixture.Expect.Path
		actions[i].expectedBody = fixture.Expect.Body
		actions[i].status = fixture.Response.Status
		actions[i].response = fixture.Response.Body
		if actions[i].rootFlag != "" {
			if actions[i].flags == nil {
				actions[i].flags = map[string][]string{}
			}
			var record any
			if err := json.Unmarshal(fixture.Record, &record); err != nil {
				t.Fatalf("decode Auto Dialer fixture record %s: %v", actions[i].fixture, err)
			}
			compactRecord, err := json.Marshal(record)
			if err != nil {
				t.Fatalf("compact Auto Dialer fixture record %s: %v", actions[i].fixture, err)
			}
			actions[i].flags[actions[i].rootFlag] = []string{string(compactRecord)}
		}
		byRequest[actions[i].method+" "+actions[i].requestPath] = &actions[i]
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("Auto Dialer write fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("Auto Dialer write fixture did not receive the Zoom bearer credential")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("Auto Dialer write fixture received undeclared query or paging input")
		}
		if len(action.expectedBody) == 0 {
			var unexpected any
			if err := json.NewDecoder(request.Body).Decode(&unexpected); err == nil {
				t.Fatal("Auto Dialer no-body action unexpectedly sent a request body")
			}
		} else {
			if !strings.HasPrefix(request.Header.Get("Content-Type"), "application/json") {
				t.Fatal("Auto Dialer body action did not declare JSON content")
			}
			var got, want any
			if err := json.NewDecoder(request.Body).Decode(&got); err != nil {
				t.Fatalf("decode Auto Dialer action body: %v", err)
			}
			if err := json.Unmarshal(action.expectedBody, &want); err != nil {
				t.Fatalf("decode Auto Dialer fixture body: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("Auto Dialer action body did not contain exactly the declared documented fields")
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
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	for _, action := range actions {
		if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{Path: action.path, Flags: action.flags}); err != nil {
			t.Fatalf("BuildWriteCommand(%s): %v", action.name, err)
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
		Config:    map[string]string{"base_url": server.URL + "/v2"},
		Secrets:   map[string]string{"access_token": accessToken},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			beforeRequests := requests
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
				t.Fatal("Auto Dialer plan did not produce a no-network preview and single-use approval")
			}
			if action.destructive != (plan.ConfirmationChallenge == string(connectors.ConfirmationKindDestructive)) {
				t.Fatal("Auto Dialer plan did not retain the declared destructive confirmation policy")
			}
			encodedPlan, err := json.Marshal(plan.Sample)
			if err != nil {
				t.Fatalf("marshal Auto Dialer plan sample: %v", err)
			}
			for index, raw := range action.inputSensitive {
				if strings.Contains(string(encodedPlan), raw) {
					t.Fatalf("Auto Dialer plan sample exposed declared sensitive input %d", index)
				}
			}
			if requests != beforeRequests {
				t.Fatal("Auto Dialer plan or preview reached the fixture endpoint")
			}

			runRequest := app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}
			if action.destructive {
				if _, err := application.RunReverseETL(context.Background(), runRequest); err == nil {
					t.Fatal("Auto Dialer DELETE execution bypassed typed destructive confirmation")
				}
				if requests != beforeRequests {
					t.Fatal("unconfirmed Auto Dialer DELETE reached the fixture endpoint")
				}
				runRequest.Confirmation = connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
			}
			run, err := application.RunReverseETL(context.Background(), runRequest)
			if err != nil {
				t.Fatalf("RunReverseETL(%s): %v", action.name, err)
			}
			if run.Status != "completed" || run.OperationDirectWrite == nil || run.OperationDirectWrite.Status != action.status {
				t.Fatalf("Auto Dialer run = %#v, want completed declared action status %d", run, action.status)
			}
			if len(action.response) == 0 {
				if run.OperationDirectWrite.Body != nil {
					t.Fatal("Auto Dialer status-only action returned an invented response body")
				}
				return
			}
			encoded, err := json.Marshal(run.OperationDirectWrite.Body)
			if err != nil {
				t.Fatalf("marshal Auto Dialer action response: %v", err)
			}
			for index, raw := range action.responseSensitive {
				if strings.Contains(string(encoded), raw) {
					t.Fatalf("Auto Dialer action response exposed declared or generic sensitive field %d", index)
				}
			}
			for _, field := range action.responseMarkers {
				if !strings.Contains(string(encoded), "\""+field+"_redacted\":true") {
					t.Fatalf("Auto Dialer action response is missing %s_redacted marker", field)
				}
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("Auto Dialer write fixtures received %d requests, want %d", requests, len(actions))
	}
}

// TestTasksDirectReadCommandsExecuteWithFixtures runs each documented Tasks
// GET through the real command runner. It locks exact paths and bearer use,
// forbids synthesized paging/filter flags, and proves named response redaction
// without reaching Zoom.
func TestTasksDirectReadCommandsExecuteWithFixtures(t *testing.T) {
	const accessToken = "fixture-tasks-access-token"
	type taskReadAction struct {
		name              string
		path              []string
		flags             map[string][]string
		requestPath       string
		fixture           string
		status            int
		response          json.RawMessage
		responseSensitive []string
		responseMarkers   []string
	}
	actions := []taskReadAction{
		{
			name:              "list assignees",
			path:              []string{"tasks", "assignees", "list"},
			flags:             map[string][]string{"task-id": {"fixture-task-assignees"}},
			requestPath:       "/v2/tasks/items/fixture-task-assignees/assignees",
			fixture:           "get_task_assignees.json",
			responseSensitive: []string{"fixture-task-assignees", "fixture-task-assignee-user", "Fixture Task Assignee", "fixture.task.assignee@example.invalid"},
			responseMarkers:   []string{"task_id", "assignees"},
		},
		{
			name:              "list collaborators",
			path:              []string{"tasks", "collaborators", "list"},
			flags:             map[string][]string{"task-id": {"fixture-task-collaborators"}},
			requestPath:       "/v2/tasks/items/fixture-task-collaborators/collaborators",
			fixture:           "get_task_collaborators.json",
			responseSensitive: []string{"fixture-task-collaborators", "fixture-task-collaborator-user", "Fixture Task Collaborator", "fixture.task.collaborator@example.invalid"},
			responseMarkers:   []string{"task_id", "collaborators"},
		},
		{
			name:              "list comments",
			path:              []string{"tasks", "comments", "list"},
			flags:             map[string][]string{"task-id": {"fixture-task-comment"}},
			requestPath:       "/v2/tasks/items/fixture-task-comment/comments",
			fixture:           "list_task_comments.json",
			responseSensitive: []string{"fixture-task-comment", "Fixture Task private comment", "fixture.task.mention@example.invalid", "fixture-task-comments-page-token"},
			responseMarkers:   []string{"comments", "next_page_token"},
		},
		{
			name:              "get import",
			path:              []string{"tasks", "imports", "get"},
			flags:             map[string][]string{"import-id": {"fixture-task-import"}},
			requestPath:       "/v2/tasks/imports/fixture-task-import",
			fixture:           "get_task_import.json",
			responseSensitive: []string{"fixture-task-import", "Fixture Task Project", "fixture-task-source-record", "fixture task import error"},
			responseMarkers:   []string{"import_id", "project_name", "failed_task_details", "error"},
		},
		{
			name:              "list items",
			path:              []string{"tasks", "items", "list"},
			requestPath:       "/v2/tasks/items",
			fixture:           "list_tasks.json",
			responseSensitive: []string{"fixture-task-list-item", "Fixture Task list title", "fixture.task.list@example.invalid", "fixture-task-items-page-token"},
			responseMarkers:   []string{"tasks", "next_page_token"},
		},
		{
			name:              "get item",
			path:              []string{"tasks", "items", "get"},
			flags:             map[string][]string{"task-id": {"fixture-task-detail"}},
			requestPath:       "/v2/tasks/items/fixture-task-detail",
			fixture:           "get_task.json",
			responseSensitive: []string{"fixture-task-detail", "Fixture Task detail title", "fixture.task.detail@example.invalid", "fixture-task-attachment", "fixture-task-meeting-uuid"},
			responseMarkers:   []string{"task_id", "title", "tasks_attachments", "associated_meeting"},
		},
	}
	byRequest := make(map[string]*taskReadAction, len(actions))
	for i := range actions {
		actions[i].status, actions[i].response = zoomDirectReadFixture(t, actions[i].fixture)
		byRequest[http.MethodGet+" "+actions[i].requestPath] = &actions[i]
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("Tasks read fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("Tasks read fixture did not receive the Zoom bearer credential")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("Tasks read fixture received undeclared query or paging input")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(action.status)
		_, _ = w.Write(action.response)
	}))
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	config := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": server.URL + "/v2"},
		Secrets: map[string]string{"access_token": accessToken},
	}
	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
				Path:   action.path,
				Flags:  action.flags,
				Config: config,
			}, func(connectors.Record) error {
				t.Fatal("emit called for a Tasks direct_read command")
				return nil
			})
			if err != nil {
				t.Fatalf("Run(%q): %v", strings.Join(action.path, " "), err)
			}
			if result.DirectRead == nil || result.DirectRead.Status != action.status {
				t.Fatalf("Run(%q) result = %#v, want status %d", strings.Join(action.path, " "), result.DirectRead, action.status)
			}
			encoded, err := json.Marshal(result.DirectRead.Body)
			if err != nil {
				t.Fatalf("marshal Tasks read response: %v", err)
			}
			for index, raw := range action.responseSensitive {
				if strings.Contains(string(encoded), raw) {
					t.Fatalf("Tasks read response exposed declared or generic sensitive field %d", index)
				}
			}
			for _, field := range action.responseMarkers {
				if !strings.Contains(string(encoded), "\""+field+"_redacted\":true") {
					t.Fatalf("Tasks read response is missing %s_redacted marker", field)
				}
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("Tasks read fixtures received %d requests, want %d", requests, len(actions))
	}
}

// TestTasksDirectWriteCommandsExecuteWithFixtures runs every non-upload Tasks
// mutation through plan, no-network preview, approval, and execute. The upload
// has its own test below because its provider-declared fileapi redirect must be
// exercised through the fixed origin rather than the normal Zoom base URL.
func TestTasksDirectWriteCommandsExecuteWithFixtures(t *testing.T) {
	const (
		credentialName = "zoom-tasks-fixture"
		accessToken    = "fixture-tasks-access-token"
	)
	type taskWriteAction struct {
		name              string
		path              []string
		flags             map[string][]string
		rootFlag          string
		fixture           string
		method            string
		requestPath       string
		expectedBody      json.RawMessage
		status            int
		response          json.RawMessage
		destructive       bool
		inputSensitive    []string
		responseSensitive []string
		responseMarkers   []string
	}
	actions := []taskWriteAction{
		{
			name:              "add assignees",
			path:              []string{"tasks", "assignees", "add"},
			flags:             map[string][]string{"task-id": {"fixture-task-assignees"}},
			rootFlag:          "request",
			fixture:           "add_task_assignees.json",
			inputSensitive:    []string{"fixture.task.assignee@example.invalid", "Fixture Task assignee invitation"},
			responseSensitive: []string{"fixture-task-assignees", "fixture-task-assignee-user", "Fixture Task Assignee", "fixture.task.assignee@example.invalid"},
			responseMarkers:   []string{"task_id", "assignees"},
		},
		{
			name:        "remove assignee",
			path:        []string{"tasks", "assignees", "remove"},
			flags:       map[string][]string{"task-id": {"fixture-task-assignees"}, "user-id": {"fixture-task-assignee-user"}},
			fixture:     "remove_task_assignee.json",
			destructive: true,
		},
		{
			name:              "add collaborators",
			path:              []string{"tasks", "collaborators", "add"},
			flags:             map[string][]string{"task-id": {"fixture-task-collaborators"}},
			rootFlag:          "request",
			fixture:           "add_task_collaborators.json",
			inputSensitive:    []string{"fixture.task.collaborator@example.invalid", "Fixture Task collaborator invitation"},
			responseSensitive: []string{"fixture-task-collaborators", "fixture-task-collaborator-user", "Fixture Task Collaborator", "fixture.task.collaborator@example.invalid"},
			responseMarkers:   []string{"task_id", "collaborators"},
		},
		{
			name:        "remove collaborator",
			path:        []string{"tasks", "collaborators", "remove"},
			flags:       map[string][]string{"task-id": {"fixture-task-collaborators"}, "user-id": {"fixture-task-collaborator-user"}},
			fixture:     "remove_task_collaborator.json",
			destructive: true,
		},
		{
			name:              "add comment",
			path:              []string{"tasks", "comments", "add"},
			flags:             map[string][]string{"task-id": {"fixture-task-comment"}},
			rootFlag:          "comment",
			fixture:           "add_task_comment.json",
			inputSensitive:    []string{"Fixture Task comment text", "https://example.invalid/fixture-task-comment"},
			responseSensitive: []string{"fixture-task-comment", "Fixture Task comment text"},
			responseMarkers:   []string{"comment_id", "content", "content_text"},
		},
		{
			name:        "delete comment",
			path:        []string{"tasks", "comments", "delete"},
			flags:       map[string][]string{"task-id": {"fixture-task-comment"}, "comment-id": {"fixture-task-comment"}},
			fixture:     "delete_task_comment.json",
			destructive: true,
		},
		{
			name: "submit import",
			path: []string{"tasks", "imports", "submit"},
			flags: map[string][]string{
				"file-id":        {"fixture-task-upload-file"},
				"source":         {"slack_lists"},
				"status-mapping": {`{"Not started":"To do"}`},
				"user-mapping":   {`{"fixture-task-source-user":"fixture.task.user@example.invalid"}`},
			},
			fixture:           "submit_task_import.json",
			inputSensitive:    []string{"fixture-task-upload-file", "Not started", "fixture-task-source-user", "fixture.task.user@example.invalid"},
			responseSensitive: []string{"fixture-task-import"},
			responseMarkers:   []string{"import_id"},
		},
		{
			name:              "create item",
			path:              []string{"tasks", "items", "create"},
			rootFlag:          "task",
			fixture:           "create_task.json",
			inputSensitive:    []string{"Fixture Task create title", "Fixture Task create description", "fixture.task.create.assignee@example.invalid", "fixture.task.create.collaborator@example.invalid"},
			responseSensitive: []string{"fixture-task-created", "Fixture Task create title", "Fixture Task create description", "fixture.task.create.assignee@example.invalid", "FIXTURE-TASK-3"},
			responseMarkers:   []string{"task_id", "title", "description", "assignees", "collaborators", "friendly_task_id", "link"},
		},
		{
			name:        "delete item",
			path:        []string{"tasks", "items", "delete"},
			flags:       map[string][]string{"task-id": {"fixture-task-delete"}},
			fixture:     "delete_task.json",
			destructive: true,
		},
		{
			name:           "update item",
			path:           []string{"tasks", "items", "update"},
			flags:          map[string][]string{"task-id": {"fixture-task-update"}},
			rootFlag:       "task",
			fixture:        "update_task.json",
			inputSensitive: []string{"Fixture Task updated title", "Fixture Task updated description"},
		},
	}
	byRequest := make(map[string]*taskWriteAction, len(actions))
	for i := range actions {
		fixture := zoomSCIM2WriteFixture(t, actions[i].fixture)
		actions[i].method = fixture.Expect.Method
		actions[i].requestPath = fixture.Expect.Path
		actions[i].expectedBody = fixture.Expect.Body
		actions[i].status = fixture.Response.Status
		actions[i].response = fixture.Response.Body
		if actions[i].rootFlag != "" {
			if actions[i].flags == nil {
				actions[i].flags = map[string][]string{}
			}
			var record any
			if err := json.Unmarshal(fixture.Record, &record); err != nil {
				t.Fatalf("decode Tasks fixture record %s: %v", actions[i].fixture, err)
			}
			compactRecord, err := json.Marshal(record)
			if err != nil {
				t.Fatalf("compact Tasks fixture record %s: %v", actions[i].fixture, err)
			}
			actions[i].flags[actions[i].rootFlag] = []string{string(compactRecord)}
		}
		byRequest[actions[i].method+" "+actions[i].requestPath] = &actions[i]
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("Tasks write fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("Tasks write fixture did not receive the Zoom bearer credential")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("Tasks write fixture received undeclared query or paging input")
		}
		if len(action.expectedBody) == 0 {
			var unexpected any
			if err := json.NewDecoder(request.Body).Decode(&unexpected); err == nil {
				t.Fatal("Tasks no-body action unexpectedly sent a request body")
			}
		} else {
			if !strings.HasPrefix(request.Header.Get("Content-Type"), "application/json") {
				t.Fatal("Tasks body action did not declare JSON content")
			}
			var got, want any
			if err := json.NewDecoder(request.Body).Decode(&got); err != nil {
				t.Fatalf("decode Tasks action body: %v", err)
			}
			if err := json.Unmarshal(action.expectedBody, &want); err != nil {
				t.Fatalf("decode Tasks fixture body: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("Tasks action body did not contain exactly the declared documented fields")
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
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	for _, action := range actions {
		if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{Path: action.path, Flags: action.flags}); err != nil {
			t.Fatalf("BuildWriteCommand(%s): %v", action.name, err)
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
		Config:    map[string]string{"base_url": server.URL + "/v2"},
		Secrets:   map[string]string{"access_token": accessToken},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			beforeRequests := requests
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
				t.Fatal("Tasks plan did not produce a no-network preview and single-use approval")
			}
			if action.destructive != (plan.ConfirmationChallenge == string(connectors.ConfirmationKindDestructive)) {
				t.Fatal("Tasks plan did not retain the declared destructive confirmation policy")
			}
			encodedPlan, err := json.Marshal(plan.Sample)
			if err != nil {
				t.Fatalf("marshal Tasks plan sample: %v", err)
			}
			for index, raw := range action.inputSensitive {
				if strings.Contains(string(encodedPlan), raw) {
					t.Fatalf("Tasks plan sample exposed declared sensitive input %d", index)
				}
			}
			if requests != beforeRequests {
				t.Fatal("Tasks plan or preview reached the fixture endpoint")
			}

			runRequest := app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}
			if action.destructive {
				if _, err := application.RunReverseETL(context.Background(), runRequest); err == nil {
					t.Fatal("Tasks DELETE execution bypassed typed destructive confirmation")
				}
				if requests != beforeRequests {
					t.Fatal("unconfirmed Tasks DELETE reached the fixture endpoint")
				}
				runRequest.Confirmation = connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
			}
			run, err := application.RunReverseETL(context.Background(), runRequest)
			if err != nil {
				t.Fatalf("RunReverseETL(%s): %v", action.name, err)
			}
			if run.Status != "completed" || run.OperationDirectWrite == nil || run.OperationDirectWrite.Status != action.status {
				t.Fatalf("Tasks run = %#v, want completed declared action status %d", run, action.status)
			}
			if len(action.response) == 0 {
				if run.OperationDirectWrite.Body != nil {
					t.Fatal("Tasks status-only action returned an invented response body")
				}
				return
			}
			encoded, err := json.Marshal(run.OperationDirectWrite.Body)
			if err != nil {
				t.Fatalf("marshal Tasks action response: %v", err)
			}
			for index, raw := range action.responseSensitive {
				if strings.Contains(string(encoded), raw) {
					t.Fatalf("Tasks action response exposed declared or generic sensitive field %d", index)
				}
			}
			for _, field := range action.responseMarkers {
				if !strings.Contains(string(encoded), "\""+field+"_redacted\":true") {
					t.Fatalf("Tasks action response is missing %s_redacted marker", field)
				}
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("Tasks write fixtures received %d requests, want %d", requests, len(actions))
	}
}

// TestTasksUploadDirectWriteExecutesWithFixture exercises the provider's
// unusual fixed fileapi origin and documented bearer-preserving 30x replay.
// The test transport owns both Zoom-host names, so the real command lifecycle
// is exercised without a live request or any caller-selected redirect target.
func TestTasksUploadDirectWriteExecutesWithFixture(t *testing.T) {
	const (
		credentialName = "zoom-tasks-upload-fixture"
		accessToken    = "fixture-tasks-upload-access-token"
	)
	fixture := zoomSCIM2WriteFixture(t, "upload_task_file.json")
	var record struct {
		File string `json:"file"`
	}
	if err := json.Unmarshal(fixture.Record, &record); err != nil {
		t.Fatalf("decode Tasks upload fixture record: %v", err)
	}
	if record.File == "" {
		t.Fatal("Tasks upload fixture record lacks file")
	}

	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	sourcePath := filepath.Join(root, ".polymetrics", filepath.FromSlash(record.File))
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		t.Fatalf("MkdirAll Tasks upload source: %v", err)
	}
	sourceContent := []byte(`{"items":["fixture-task-upload"]}`)
	if err := os.WriteFile(sourcePath, sourceContent, 0o600); err != nil {
		t.Fatalf("WriteFile Tasks upload source: %v", err)
	}

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	flags := map[string][]string{"file-path": {record.File}}
	if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{Path: []string{"tasks", "files", "upload"}, Flags: flags}); err != nil {
		t.Fatalf("BuildWriteCommand(tasks files upload): %v", err)
	}

	initialCalls, redirectCalls := 0, 0
	originalTransport := http.DefaultTransport
	http.DefaultTransport = zoomRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		defer func() { _ = request.Body.Close() }()
		if request.Method != fixture.Expect.Method || request.URL.Path != fixture.Expect.Path {
			t.Fatalf("Tasks upload transport request = %s %s, want %s %s", request.Method, request.URL.Path, fixture.Expect.Method, fixture.Expect.Path)
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("Tasks upload transport did not receive the declared bearer credential")
		}
		switch request.URL.Host {
		case "fileapi.zoom.us":
			initialCalls++
			if request.URL.RawQuery != "" {
				t.Fatal("Tasks upload initial request carried undeclared query input")
			}
			assertTasksUploadMultipart(t, request, sourceContent, filepath.Base(record.File))
			return &http.Response{
				StatusCode: http.StatusTemporaryRedirect,
				Header:     http.Header{"Location": []string{"https://uploads.zoom.us/v2/tasks/files?fixture-signed-redirect=opaque"}},
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    request,
			}, nil
		case "uploads.zoom.us":
			redirectCalls++
			if request.URL.RawQuery != "fixture-signed-redirect=opaque" {
				t.Fatal("Tasks upload redirect request did not preserve the provider-supplied signed query")
			}
			assertTasksUploadMultipart(t, request, sourceContent, filepath.Base(record.File))
			return &http.Response{
				StatusCode: fixture.Response.Status,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(string(fixture.Response.Body))),
				Request:    request,
			}, nil
		default:
			t.Fatalf("Tasks upload transport received unexpected host %q", request.URL.Host)
			return nil, nil
		}
	})
	t.Cleanup(func() { http.DefaultTransport = originalTransport })

	application, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := application.AddCredential(context.Background(), app.AddCredentialRequest{
		Name:      credentialName,
		Connector: zoomBundleName,
		Config:    map[string]string{"base_url": "https://api.zoom.us/v2"},
		Secrets:   map[string]string{"access_token": accessToken},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	plan, preview, err := application.PlanConnectorCommand(context.Background(), app.PlanConnectorCommandRequest{
		Connector:  zoomBundleName,
		Credential: credentialName,
		Path:       []string{"tasks", "files", "upload"},
		Flags:      flags,
		Preview:    true,
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand(tasks files upload): %v", err)
	}
	if preview == nil || preview.Digest == "" || plan.ApprovalToken == "" {
		t.Fatal("Tasks upload plan did not produce a no-network preview and single-use approval")
	}
	encodedPlan, err := json.Marshal(plan.Sample)
	if err != nil {
		t.Fatalf("marshal Tasks upload plan sample: %v", err)
	}
	if strings.Contains(string(encodedPlan), record.File) {
		t.Fatal("Tasks upload plan sample exposed the declared source path")
	}
	if initialCalls != 0 || redirectCalls != 0 {
		t.Fatal("Tasks upload plan or preview reached the fixture transport")
	}

	run, err := application.RunReverseETL(context.Background(), app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken})
	if err != nil {
		t.Fatalf("RunReverseETL(tasks files upload): %v", err)
	}
	if run.Status != "completed" || run.OperationDirectWrite == nil || run.OperationDirectWrite.Status != fixture.Response.Status {
		t.Fatalf("Tasks upload run = %#v, want completed declared action status %d", run, fixture.Response.Status)
	}
	encodedResponse, err := json.Marshal(run.OperationDirectWrite.Body)
	if err != nil {
		t.Fatalf("marshal Tasks upload response: %v", err)
	}
	if strings.Contains(string(encodedResponse), "fixture-task-upload-file") || !strings.Contains(string(encodedResponse), "\"file_id_redacted\":true") {
		t.Fatal("Tasks upload response did not redact the provider file identifier")
	}
	if initialCalls != 1 || redirectCalls != 1 {
		t.Fatalf("Tasks upload transport calls initial=%d redirected=%d, want one each", initialCalls, redirectCalls)
	}
}

func assertTasksUploadMultipart(t *testing.T, request *http.Request, want []byte, wantName string) {
	t.Helper()
	mediaType, params, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" || params["boundary"] == "" {
		t.Fatalf("Tasks upload content type = %q, want multipart/form-data with boundary", request.Header.Get("Content-Type"))
	}
	reader := multipart.NewReader(request.Body, params["boundary"])
	part, err := reader.NextPart()
	if err != nil {
		t.Fatalf("read Tasks upload multipart part: %v", err)
	}
	defer func() { _ = part.Close() }()
	if part.FormName() != "file" || part.FileName() != wantName {
		t.Fatalf("Tasks upload part = field %q file %q, want file %q", part.FormName(), part.FileName(), wantName)
	}
	if part.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("Tasks upload part content type = %q, want application/json", part.Header.Get("Content-Type"))
	}
	got, err := io.ReadAll(part)
	if err != nil {
		t.Fatalf("read Tasks upload multipart content: %v", err)
	}
	if string(got) != string(want) {
		t.Fatal("Tasks upload multipart content does not match the approved source snapshot")
	}
	if _, err := reader.NextPart(); err != io.EOF {
		t.Fatalf("Tasks upload multipart has an undeclared extra part: %v", err)
	}
}

// TestWorkforceManagementDirectReadCommandsExecuteWithFixtures sends every
// documented Workforce Management GET through the real runner. The loopback
// server proves each fixed path and ordinary Zoom bearer without sending a
// request to the provider, while the fixture assertions prove output
// redaction and reject synthesized query/paging input.
func TestWorkforceManagementDirectReadCommandsExecuteWithFixtures(t *testing.T) {
	const accessToken = "fixture-wfm-read-access-token"
	type workforceReadAction struct {
		name              string
		path              []string
		flags             map[string][]string
		requestPath       string
		fixture           string
		status            int
		response          json.RawMessage
		responseSensitive []string
		responseMarkers   []string
	}
	actions := []workforceReadAction{
		{
			name:              "list filter groups",
			path:              []string{"workforce-management", "filter-groups", "list"},
			requestPath:       "/v2/workforce-management/filter-groups",
			fixture:           "list_workforce_filter_groups.json",
			responseSensitive: []string{"fixture-wfm-filter-group", "Fixture Workforce Management filter group", "fixture-wfm-filter-page"},
			responseMarkers:   []string{"filter_groups", "next_page_token", "total_records"},
		},
		{
			name:              "list forecasts",
			path:              []string{"workforce-management", "forecasts", "list"},
			requestPath:       "/v2/workforce-management/forecasts",
			fixture:           "list_workforce_forecasts.json",
			responseSensitive: []string{"fixture-wfm-forecast", "Fixture Workforce Management forecast", "fixture-wfm-scheduling-group", "fixture-wfm-forecast-page"},
			responseMarkers:   []string{"forecasts", "next_page_token", "total_records"},
		},
		{
			name:              "get forecast scheduling group",
			path:              []string{"workforce-management", "forecasts", "scheduling-groups", "get"},
			flags:             map[string][]string{"forecast-id": {"fixture-wfm-forecast"}, "scheduling-group-id": {"fixture-wfm-scheduling-group"}},
			requestPath:       "/v2/workforce-management/forecasts/fixture-wfm-forecast/scheduling-groups/fixture-wfm-scheduling-group",
			fixture:           "get_workforce_forecast_scheduling_group.json",
			responseSensitive: []string{"fixture-wfm-forecast", "fixture-wfm-scheduling-group", "2026-01-01T00:00:00Z", "2026-01-01T00:15:00Z"},
			responseMarkers:   []string{"forecast_id", "scheduling_group_id", "forecast_intervals"},
		},
		{
			name:              "get historical queue metrics import",
			path:              []string{"workforce-management", "imports", "historical-queue-metrics", "get"},
			flags:             map[string][]string{"import-id": {"fixture-wfm-queue-import"}},
			requestPath:       "/v2/workforce-management/imports/fixture-wfm-queue-import/historical-queue-metrics",
			fixture:           "get_workforce_historical_queue_metrics_import.json",
			responseSensitive: []string{"fixture-wfm-queue-import", "fixture-wfm-queue.csv", "fixture-wfm-user", "2026-01-01T00:00:00Z", "UTC"},
			responseMarkers:   []string{"import_id", "file_name", "imported_by", "create_time", "timezone"},
		},
		{
			name:              "list organizational groups",
			path:              []string{"workforce-management", "organizational-groups", "list"},
			requestPath:       "/v2/workforce-management/organizational-groups",
			fixture:           "list_workforce_organizational_groups.json",
			responseSensitive: []string{"fixture-wfm-org-group", "Fixture Workforce Management organizational group", "fixture-wfm-org-page"},
			responseMarkers:   []string{"organizational_groups", "next_page_token", "total_records"},
		},
		{
			name:              "get organizational group",
			path:              []string{"workforce-management", "organizational-groups", "get"},
			flags:             map[string][]string{"organizational-group-id": {"fixture-wfm-org-group"}},
			requestPath:       "/v2/workforce-management/organizational-groups/fixture-wfm-org-group",
			fixture:           "get_workforce_organizational_group.json",
			responseSensitive: []string{"fixture-wfm-org-group", "Fixture Workforce Management organizational group", "fixture-wfm-scheduling-group", "2026-01-01T00:00:00Z"},
			responseMarkers:   []string{"org_group_id", "name", "description", "scheduling_group_ids"},
		},
		{
			name:              "list adherence agents",
			path:              []string{"workforce-management", "reports", "adherence", "agents", "list"},
			requestPath:       "/v2/workforce-management/reports/adherence/agents",
			fixture:           "list_workforce_adherence_agents.json",
			responseSensitive: []string{"fixture-wfm-adherence-user", "Fixture Workforce Management adherence agent", "fixture-wfm-activity", "fixture-wfm-adherence-page"},
			responseMarkers:   []string{"agents", "next_page_token"},
		},
		{
			name:              "list report schedule agents",
			path:              []string{"workforce-management", "reports", "schedules", "agents", "list"},
			requestPath:       "/v2/workforce-management/reports/schedules/agents",
			fixture:           "list_workforce_report_schedule_agents.json",
			responseSensitive: []string{"fixture-wfm-report-user", "fixture.wfm.report@example.invalid", "Fixture Workforce Management report schedule agent", "fixture-wfm-report-page"},
			responseMarkers:   []string{"user_id", "user_email", "display_name", "activities", "next_page_token"},
		},
		{
			name:              "list schedule agents",
			path:              []string{"workforce-management", "schedules", "agents", "list"},
			requestPath:       "/v2/workforce-management/schedules/agents",
			fixture:           "list_workforce_schedule_agents.json",
			responseSensitive: []string{"fixture-wfm-schedule-user", "fixture.wfm.schedule@example.invalid", "Fixture Workforce Management schedule agent", "fixture-wfm-schedule-page"},
			responseMarkers:   []string{"user_id", "user_email", "display_name", "activities", "next_page_token"},
		},
		{
			name:              "list scheduling groups",
			path:              []string{"workforce-management", "scheduling-groups", "list"},
			requestPath:       "/v2/workforce-management/scheduling-groups",
			fixture:           "list_workforce_scheduling_groups.json",
			responseSensitive: []string{"fixture-wfm-scheduling-group", "Fixture Workforce Management scheduling group", "fixture-wfm-scheduling-page"},
			responseMarkers:   []string{"scheduling_groups", "next_page_token", "total_records"},
		},
		{
			name:              "list users",
			path:              []string{"workforce-management", "users", "list"},
			requestPath:       "/v2/workforce-management/users",
			fixture:           "list_workforce_users.json",
			responseSensitive: []string{"fixture-wfm-user", "Fixture Workforce Management user", "fixture.wfm.user@example.invalid", "fixture-wfm-users-page"},
			responseMarkers:   []string{"users", "next_page_token", "total_records"},
		},
	}
	byRequest := make(map[string]*workforceReadAction, len(actions))
	for i := range actions {
		actions[i].status, actions[i].response = zoomDirectReadFixture(t, actions[i].fixture)
		byRequest[http.MethodGet+" "+actions[i].requestPath] = &actions[i]
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("Workforce Management read fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("Workforce Management read fixture did not receive the Zoom bearer credential")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("Workforce Management read fixture received undeclared query or paging input")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(action.status)
		_, _ = w.Write(action.response)
	}))
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	config := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": server.URL + "/v2"},
		Secrets: map[string]string{"access_token": accessToken},
	}
	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
				Path:   action.path,
				Flags:  action.flags,
				Config: config,
			}, func(connectors.Record) error {
				t.Fatal("emit called for a Workforce Management direct_read command")
				return nil
			})
			if err != nil {
				t.Fatalf("Run(%q): %v", strings.Join(action.path, " "), err)
			}
			if result.DirectRead == nil || result.DirectRead.Status != action.status {
				t.Fatalf("Run(%q) result = %#v, want status %d", strings.Join(action.path, " "), result.DirectRead, action.status)
			}
			encoded, err := json.Marshal(result.DirectRead.Body)
			if err != nil {
				t.Fatalf("marshal Workforce Management read response: %v", err)
			}
			for index, raw := range action.responseSensitive {
				if strings.Contains(string(encoded), raw) {
					t.Fatalf("Workforce Management read response exposed declared or generic sensitive field %d", index)
				}
			}
			for _, field := range action.responseMarkers {
				if !strings.Contains(string(encoded), "\""+field+"_redacted\":true") {
					t.Fatalf("Workforce Management read response is missing %s_redacted marker", field)
				}
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("Workforce Management read fixtures received %d requests, want %d", requests, len(actions))
	}
}

// TestWorkforceManagementJSONDirectWriteCommandsExecuteWithFixtures exercises
// the five documented JSON mutations through the typed plan, preview,
// approval, and execution lifecycle. It makes 204 deletes status-only and
// proves destructive confirmation is required before either delete can reach
// the loopback endpoint.
func TestWorkforceManagementJSONDirectWriteCommandsExecuteWithFixtures(t *testing.T) {
	const (
		credentialName = "zoom-wfm-json-fixture"
		accessToken    = "fixture-wfm-json-access-token"
	)
	type workforceWriteAction struct {
		name              string
		path              []string
		flags             map[string][]string
		rootFlag          string
		fixture           string
		method            string
		requestPath       string
		expectedBody      json.RawMessage
		status            int
		response          json.RawMessage
		destructive       bool
		inputSensitive    []string
		responseSensitive []string
		responseMarkers   []string
	}
	actions := []workforceWriteAction{
		{
			name:            "upload historical agent status",
			path:            []string{"workforce-management", "imports", "historical-agent-status", "upload"},
			rootFlag:        "status-import",
			fixture:         "upload_workforce_historical_agent_status.json",
			inputSensitive:  []string{"fixture-wfm-status-user", "offline", "Available"},
			responseMarkers: []string{"imported_rows", "created_statuses"},
		},
		{
			name:           "delete historical agent status",
			path:           []string{"workforce-management", "imports", "historical-agent-status", "delete"},
			rootFlag:       "status-deletion",
			fixture:        "delete_workforce_historical_agent_status.json",
			destructive:    true,
			inputSensitive: []string{"fixture-wfm-status-user", "2026-01-01T00:00:00Z"},
		},
		{
			name:              "create organizational group",
			path:              []string{"workforce-management", "organizational-groups", "create"},
			rootFlag:          "organizational-group",
			fixture:           "create_workforce_organizational_group.json",
			inputSensitive:    []string{"Fixture Workforce Management organizational group", "fixture-wfm-scheduling-group"},
			responseSensitive: []string{"fixture-wfm-org-group", "Fixture Workforce Management organizational group", "fixture-wfm-scheduling-group"},
			responseMarkers:   []string{"org_group_id", "name", "description", "scheduling_group_ids"},
		},
		{
			name:        "delete organizational group",
			path:        []string{"workforce-management", "organizational-groups", "delete"},
			flags:       map[string][]string{"organizational-group-id": {"fixture-wfm-org-group"}},
			fixture:     "delete_workforce_organizational_group.json",
			destructive: true,
		},
		{
			name:              "update organizational group",
			path:              []string{"workforce-management", "organizational-groups", "update"},
			flags:             map[string][]string{"organizational-group-id": {"fixture-wfm-org-group"}},
			rootFlag:          "organizational-group",
			fixture:           "update_workforce_organizational_group.json",
			inputSensitive:    []string{"Fixture Workforce Management updated organizational group description"},
			responseSensitive: []string{"fixture-wfm-org-group", "Fixture Workforce Management updated organizational group description"},
			responseMarkers:   []string{"org_group_id", "description"},
		},
	}
	byRequest := make(map[string]*workforceWriteAction, len(actions))
	for i := range actions {
		fixture := zoomSCIM2WriteFixture(t, actions[i].fixture)
		actions[i].method = fixture.Expect.Method
		actions[i].requestPath = fixture.Expect.Path
		actions[i].expectedBody = fixture.Expect.Body
		actions[i].status = fixture.Response.Status
		actions[i].response = fixture.Response.Body
		if actions[i].rootFlag != "" {
			if actions[i].flags == nil {
				actions[i].flags = map[string][]string{}
			}
			var record any
			if err := json.Unmarshal(fixture.Record, &record); err != nil {
				t.Fatalf("decode Workforce Management fixture record %s: %v", actions[i].fixture, err)
			}
			compactRecord, err := json.Marshal(record)
			if err != nil {
				t.Fatalf("compact Workforce Management fixture record %s: %v", actions[i].fixture, err)
			}
			actions[i].flags[actions[i].rootFlag] = []string{string(compactRecord)}
		}
		byRequest[actions[i].method+" "+actions[i].requestPath] = &actions[i]
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("Workforce Management JSON write fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("Workforce Management JSON write fixture did not receive the Zoom bearer credential")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("Workforce Management JSON write fixture received undeclared query or paging input")
		}
		if len(action.expectedBody) == 0 {
			var unexpected any
			if err := json.NewDecoder(request.Body).Decode(&unexpected); err == nil {
				t.Fatal("Workforce Management no-body action unexpectedly sent a request body")
			}
		} else {
			if !strings.HasPrefix(request.Header.Get("Content-Type"), "application/json") {
				t.Fatal("Workforce Management body action did not declare JSON content")
			}
			var got, want any
			if err := json.NewDecoder(request.Body).Decode(&got); err != nil {
				t.Fatalf("decode Workforce Management action body: %v", err)
			}
			if err := json.Unmarshal(action.expectedBody, &want); err != nil {
				t.Fatalf("decode Workforce Management fixture body: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("Workforce Management action body did not contain exactly the declared documented fields")
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
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	for _, action := range actions {
		if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{Path: action.path, Flags: action.flags}); err != nil {
			t.Fatalf("BuildWriteCommand(%s): %v", action.name, err)
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
		Config:    map[string]string{"base_url": server.URL + "/v2"},
		Secrets:   map[string]string{"access_token": accessToken},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			beforeRequests := requests
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
				t.Fatal("Workforce Management JSON plan did not produce a no-network preview and single-use approval")
			}
			if action.destructive != (plan.ConfirmationChallenge == string(connectors.ConfirmationKindDestructive)) {
				t.Fatal("Workforce Management JSON plan did not retain the declared destructive confirmation policy")
			}
			encodedPlan, err := json.Marshal(plan.Sample)
			if err != nil {
				t.Fatalf("marshal Workforce Management JSON plan sample: %v", err)
			}
			for index, raw := range action.inputSensitive {
				if strings.Contains(string(encodedPlan), raw) {
					t.Fatalf("Workforce Management JSON plan sample exposed declared sensitive input %d", index)
				}
			}
			if requests != beforeRequests {
				t.Fatal("Workforce Management JSON plan or preview reached the fixture endpoint")
			}

			runRequest := app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}
			if action.destructive {
				if _, err := application.RunReverseETL(context.Background(), runRequest); err == nil {
					t.Fatal("Workforce Management DELETE execution bypassed typed destructive confirmation")
				}
				if requests != beforeRequests {
					t.Fatal("unconfirmed Workforce Management DELETE reached the fixture endpoint")
				}
				runRequest.Confirmation = connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
			}
			run, err := application.RunReverseETL(context.Background(), runRequest)
			if err != nil {
				t.Fatalf("RunReverseETL(%s): %v", action.name, err)
			}
			if run.Status != "completed" || run.OperationDirectWrite == nil || run.OperationDirectWrite.Status != action.status {
				t.Fatalf("Workforce Management JSON run = %#v, want completed declared action status %d", run, action.status)
			}
			if len(action.response) == 0 {
				if run.OperationDirectWrite.Body != nil {
					t.Fatal("Workforce Management status-only action returned an invented response body")
				}
				return
			}
			encoded, err := json.Marshal(run.OperationDirectWrite.Body)
			if err != nil {
				t.Fatalf("marshal Workforce Management JSON action response: %v", err)
			}
			for index, raw := range action.responseSensitive {
				if strings.Contains(string(encoded), raw) {
					t.Fatalf("Workforce Management JSON action response exposed declared or generic sensitive field %d", index)
				}
			}
			for _, field := range action.responseMarkers {
				if !strings.Contains(string(encoded), "\""+field+"_redacted\":true") {
					t.Fatalf("Workforce Management JSON action response is missing %s_redacted marker", field)
				}
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("Workforce Management JSON write fixtures received %d requests, want %d", requests, len(actions))
	}
}

// TestWorkforceManagementCSVDirectWritesExecuteWithFixtures exercises both
// provider-declared CSV imports through the typed plan lifecycle. It proves
// the approved bounded snapshot, fixed text/csv part header, exact declared
// non-file parts, source duration range, response redaction, and the absence
// of caller-controlled URL/query/parser behavior.
func TestWorkforceManagementCSVDirectWritesExecuteWithFixtures(t *testing.T) {
	const (
		credentialName = "zoom-wfm-csv-fixture"
		accessToken    = "fixture-wfm-csv-access-token"
	)
	type workforceCSVFixture struct {
		Record struct {
			File string `json:"file"`
		} `json:"record"`
		SourceCSV string `json:"source_csv"`
		Expect    struct {
			Method string            `json:"method"`
			Path   string            `json:"path"`
			Fields map[string]string `json:"fields"`
		} `json:"expect"`
		Response struct {
			Status int             `json:"status"`
			Body   json.RawMessage `json:"body"`
		} `json:"response"`
	}
	type workforceCSVAction struct {
		name              string
		path              []string
		flags             map[string][]string
		fixture           string
		file              string
		sourceCSV         []byte
		method            string
		requestPath       string
		expectedFields    map[string]string
		status            int
		response          json.RawMessage
		responseSensitive []string
		responseMarkers   []string
	}
	actions := []workforceCSVAction{
		{
			name:              "upload historical queue metrics",
			path:              []string{"workforce-management", "imports", "historical-queue-metrics", "upload"},
			flags:             map[string][]string{"allow-overwrite": {"true"}, "timezone": {"UTC"}},
			fixture:           "upload_workforce_historical_queue_metrics.json",
			responseSensitive: []string{"fixture-wfm-queue-import"},
			responseMarkers:   []string{"import_id"},
		},
		{
			name: "upload forecast staffing",
			path: []string{"workforce-management", "imports", "staffing", "upload"},
			flags: map[string][]string{
				"forecast-duration-weeks": {"2"},
				"forecast-name":           {"Fixture Workforce Management staffing forecast"},
				"forecast-start-date":     {"2026-01-01"},
				"has-interval-offsets":    {"false"},
				"timezone":                {"UTC"},
			},
			fixture:           "upload_workforce_forecast_staffing.json",
			responseSensitive: []string{"fixture-wfm-staffing-forecast", "Fixture Workforce Management staffing forecast"},
			responseMarkers:   []string{"forecasts"},
		},
	}
	for i := range actions {
		raw, err := os.ReadFile(filepath.Join("fixtures", "writes", actions[i].fixture))
		if err != nil {
			t.Fatalf("read Workforce Management CSV fixture %s: %v", actions[i].fixture, err)
		}
		var fixture workforceCSVFixture
		if err := json.Unmarshal(raw, &fixture); err != nil {
			t.Fatalf("decode Workforce Management CSV fixture %s: %v", actions[i].fixture, err)
		}
		if fixture.Record.File == "" || fixture.SourceCSV == "" || fixture.Expect.Method == "" || fixture.Expect.Path == "" || fixture.Response.Status == 0 {
			t.Fatalf("Workforce Management CSV fixture %s lacks a documented lifecycle value", actions[i].fixture)
		}
		actions[i].file = fixture.Record.File
		actions[i].sourceCSV = []byte(fixture.SourceCSV)
		actions[i].method = fixture.Expect.Method
		actions[i].requestPath = fixture.Expect.Path
		actions[i].expectedFields = fixture.Expect.Fields
		actions[i].status = fixture.Response.Status
		actions[i].response = fixture.Response.Body
		actions[i].flags["file-path"] = []string{fixture.Record.File}
	}

	byRequest := make(map[string]*workforceCSVAction, len(actions))
	for i := range actions {
		byRequest[actions[i].method+" "+actions[i].requestPath] = &actions[i]
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("Workforce Management CSV fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("Workforce Management CSV fixture did not receive the Zoom bearer credential")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("Workforce Management CSV fixture received undeclared query or paging input")
		}
		assertWorkforceManagementCSVMultipart(t, request, action.sourceCSV, filepath.Base(action.file), action.expectedFields)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(action.status)
		_, _ = w.Write(action.response)
	}))
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	for _, action := range actions {
		if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{Path: action.path, Flags: action.flags}); err != nil {
			t.Fatalf("BuildWriteCommand(%s): %v", action.name, err)
		}
	}
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	for _, action := range actions {
		sourcePath := filepath.Join(root, ".polymetrics", filepath.FromSlash(action.file))
		if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
			t.Fatalf("MkdirAll Workforce Management CSV source: %v", err)
		}
		if err := os.WriteFile(sourcePath, action.sourceCSV, 0o600); err != nil {
			t.Fatalf("WriteFile Workforce Management CSV source: %v", err)
		}
	}
	application, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := application.AddCredential(context.Background(), app.AddCredentialRequest{
		Name:      credentialName,
		Connector: zoomBundleName,
		Config:    map[string]string{"base_url": server.URL + "/v2"},
		Secrets:   map[string]string{"access_token": accessToken},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}
	invalidStaffingFlags := map[string][]string{
		"file-path":               {"imports/fixture-wfm-staffing.csv"},
		"forecast-duration-weeks": {"4.01"},
		"forecast-name":           {"Fixture Workforce Management staffing forecast"},
		"forecast-start-date":     {"2026-01-01"},
		"has-interval-offsets":    {"false"},
		"timezone":                {"UTC"},
	}
	if _, _, err := application.PlanConnectorCommand(context.Background(), app.PlanConnectorCommandRequest{
		Connector:  zoomBundleName,
		Credential: credentialName,
		Path:       []string{"workforce-management", "imports", "staffing", "upload"},
		Flags:      invalidStaffingFlags,
		Preview:    true,
	}); err == nil || !strings.Contains(err.Error(), "maximum") {
		t.Fatalf("PlanConnectorCommand(staffing above source maximum) = %v, want numeric range rejection", err)
	}
	if requests != 0 {
		t.Fatal("Workforce Management invalid staffing plan reached the fixture endpoint")
	}

	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			beforeRequests := requests
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
				t.Fatal("Workforce Management CSV plan did not produce a no-network preview and single-use approval")
			}
			encodedPlan, err := json.Marshal(plan.Sample)
			if err != nil {
				t.Fatalf("marshal Workforce Management CSV plan sample: %v", err)
			}
			if strings.Contains(string(encodedPlan), action.file) || strings.Contains(string(encodedPlan), "UTC") {
				t.Fatal("Workforce Management CSV plan sample exposed a declared source path or field value")
			}
			if requests != beforeRequests {
				t.Fatal("Workforce Management CSV plan or preview reached the fixture endpoint")
			}

			run, err := application.RunReverseETL(context.Background(), app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken})
			if err != nil {
				t.Fatalf("RunReverseETL(%s): %v", action.name, err)
			}
			if run.Status != "completed" || run.OperationDirectWrite == nil || run.OperationDirectWrite.Status != action.status {
				t.Fatalf("Workforce Management CSV run = %#v, want completed declared action status %d", run, action.status)
			}
			encoded, err := json.Marshal(run.OperationDirectWrite.Body)
			if err != nil {
				t.Fatalf("marshal Workforce Management CSV action response: %v", err)
			}
			for index, raw := range action.responseSensitive {
				if strings.Contains(string(encoded), raw) {
					t.Fatalf("Workforce Management CSV action response exposed declared sensitive field %d", index)
				}
			}
			for _, field := range action.responseMarkers {
				if !strings.Contains(string(encoded), "\""+field+"_redacted\":true") {
					t.Fatalf("Workforce Management CSV action response is missing %s_redacted marker", field)
				}
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("Workforce Management CSV fixtures received %d requests, want %d", requests, len(actions))
	}
}

func assertWorkforceManagementCSVMultipart(t *testing.T, request *http.Request, wantFile []byte, wantName string, wantFields map[string]string) {
	t.Helper()
	mediaType, params, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" || params["boundary"] == "" {
		t.Fatalf("Workforce Management CSV content type = %q, want multipart/form-data with boundary", request.Header.Get("Content-Type"))
	}
	reader := multipart.NewReader(request.Body, params["boundary"])
	gotFields := make(map[string]string, len(wantFields))
	gotFile := false
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read Workforce Management CSV multipart part: %v", err)
		}
		content, err := io.ReadAll(part)
		if err != nil {
			_ = part.Close()
			t.Fatalf("read Workforce Management CSV multipart content: %v", err)
		}
		if part.FormName() == "file" {
			if gotFile || part.FileName() != wantName {
				_ = part.Close()
				t.Fatalf("Workforce Management CSV file part = already=%t name=%q, want one %q", gotFile, part.FileName(), wantName)
			}
			if part.Header.Get("Content-Type") != "text/csv" {
				_ = part.Close()
				t.Fatalf("Workforce Management CSV file content type = %q, want text/csv", part.Header.Get("Content-Type"))
			}
			if !reflect.DeepEqual(content, wantFile) {
				_ = part.Close()
				t.Fatal("Workforce Management CSV file content does not match the approved source snapshot")
			}
			gotFile = true
		} else {
			if part.FileName() != "" {
				_ = part.Close()
				t.Fatalf("Workforce Management CSV unexpected file field %q", part.FormName())
			}
			gotFields[part.FormName()] = string(content)
		}
		_ = part.Close()
	}
	if !gotFile {
		t.Fatal("Workforce Management CSV multipart request is missing the declared file part")
	}
	if !reflect.DeepEqual(gotFields, wantFields) {
		t.Fatalf("Workforce Management CSV fields = %#v, want %#v", gotFields, wantFields)
	}
}

type zoomRoundTripperFunc func(*http.Request) (*http.Response, error)

func (fn zoomRoundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
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
			case mapsTo == "body":
				if flag.Type != "json_object" && flag.Type != "json_array" {
					t.Errorf("direct-write command %q root body flag --%s type = %q, want json_object or json_array", command.Path, flag.Name, flag.Type)
				}
			case strings.HasPrefix(mapsTo, "body."), strings.HasPrefix(mapsTo, "query."):
				// Required and optional provider-defined body/query members are
				// checked by the operation body schema and connectorgen's exact
				// mapping validator. Requiring every member here would make a
				// documented optional field impossible to expose.
			default:
				t.Errorf("direct-write command %q flag --%s maps_to %q, want typed path.*, query.*, body.*, or named root body binding", command.Path, flag.Name, mapsTo)
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

// TestVirtualAgentDirectReadCommandsExecuteWithFixtures runs each documented
// Virtual Agent GET through its fixed endpoint and ordinary Zoom bearer
// transport. It proves response-only paging values are never sent as query
// input and that sensitive report/content data is redacted before output.
func TestVirtualAgentDirectReadCommandsExecuteWithFixtures(t *testing.T) {
	const accessToken = "fixture-virtual-agent-access-token"
	type virtualAgentReadAction struct {
		name              string
		path              []string
		flags             map[string][]string
		requestPath       string
		fixture           string
		status            int
		response          json.RawMessage
		responseSensitive []string
		responseMarkers   []string
	}
	actions := []virtualAgentReadAction{
		{
			name:              "list articles",
			path:              []string{"virtual-agent", "knowledge-bases", "articles", "list"},
			flags:             map[string][]string{"kb-id": {"fixture-va-kb"}},
			requestPath:       "/v2/km/kbs/fixture-va-kb/articles",
			fixture:           "list_virtual_agent_articles.json",
			responseSensitive: []string{"fixture-va-article", "fixture-va-kb", "fixture category", "Fixture Virtual Agent article", "https://fixture.invalid/virtual-agent/article", "fixture-va-external-id", "fixture Virtual Agent article content", "fixture-va-page-token", "fixture-va-response-token"},
			responseMarkers:   []string{"id", "kb_id", "category", "title", "url", "external_id", "content", "next_page_token", "token"},
		},
		{
			name:              "get article",
			path:              []string{"virtual-agent", "knowledge-bases", "articles", "get"},
			flags:             map[string][]string{"kb-id": {"fixture-va-kb"}, "article-id": {"fixture-va-article"}},
			requestPath:       "/v2/km/kbs/fixture-va-kb/articles/fixture-va-article",
			fixture:           "get_virtual_agent_article.json",
			responseSensitive: []string{"fixture-va-article", "fixture-va-kb", "fixture category", "Fixture Virtual Agent article", "https://fixture.invalid/virtual-agent/article", "fixture-va-external-id", "fixture Virtual Agent article content", "fixture-va-response-token"},
			responseMarkers:   []string{"id", "kb_id", "category", "title", "url", "external_id", "content", "token"},
		},
		{
			name:              "get sync",
			path:              []string{"virtual-agent", "knowledge-bases", "sync", "get"},
			flags:             map[string][]string{"kb-id": {"fixture-va-kb"}, "sync-id": {"fixture-va-sync"}},
			requestPath:       "/v2/km/kbs/fixture-va-kb/sync/fixture-va-sync",
			fixture:           "get_virtual_agent_sync.json",
			responseSensitive: []string{"fixture-va-sync", "fixture-va-kb", "fixture Virtual Agent sync detail", "fixture-va-response-token"},
			responseMarkers:   []string{"sync_id", "kb_id", "error_message", "token"},
		},
		{
			name:              "list engagements",
			path:              []string{"virtual-agent", "reports", "engagements", "list"},
			requestPath:       "/v2/virtual_agent/report/engagements",
			fixture:           "list_virtual_agent_engagements.json",
			responseSensitive: []string{"fixture-va-engagement", "fixture-va-consumer", "fixture-va-agent", "Fixture Virtual Agent", "fixture-va-article", "Fixture Virtual Agent article", "fixture article answer", "https://fixture.invalid/virtual-agent/article", "fixture engagement query summary", "fixture engagement intent summary", "+10000000000", "fixture-va-page-token", "fixture-va-response-token"},
			responseMarkers:   []string{"engagement_id", "consumer_id", "agents", "articles", "query_summary", "intent_summary", "user_phone_number", "next_page_token", "token"},
		},
		{
			name:              "list engagement query details",
			path:              []string{"virtual-agent", "reports", "engagements", "query-details", "list"},
			requestPath:       "/v2/virtual_agent/report/engagements/query_details",
			fixture:           "list_virtual_agent_engagement_query_details.json",
			responseSensitive: []string{"fixture-va-engagement", "fixture-va-query", "fixture query text", "fixture-va-agent", "Fixture Virtual Agent", "fixture-va-agent-session", "fixture-va-bot", "Fixture Virtual Agent bot", "fixture-va-bot-session", "fixture-va-intent", "Fixture intent", "fixture-va-article", "Fixture Virtual Agent article", "fixture article answer", "https://fixture.invalid/virtual-agent/article", "fixture-va-page-token", "fixture-va-response-token"},
			responseMarkers:   []string{"engagement_id", "query_details", "next_page_token", "token"},
		},
		{
			name:              "list engagement variable details",
			path:              []string{"virtual-agent", "reports", "engagements", "variable-details", "list"},
			requestPath:       "/v2/virtual_agent/report/engagements/variables",
			fixture:           "list_virtual_agent_engagement_variable_details.json",
			responseSensitive: []string{"fixture-va-engagement", "fixture-va-variable", "fixture variable name", "fixture variable value", "fixture-va-group", "Fixture variable group", "fixture-va-page-token", "fixture-va-response-token"},
			responseMarkers:   []string{"engagement_id", "variable_details", "next_page_token", "token"},
		},
		{
			name:              "list surveys",
			path:              []string{"virtual-agent", "reports", "surveys", "list"},
			requestPath:       "/v2/virtual_agent/report/surveys",
			fixture:           "list_virtual_agent_surveys.json",
			responseSensitive: []string{"fixture-va-survey", "Fixture survey", "fixture-va-engagement", "fixture survey consumer", "Fixture Virtual Agent", "fixture survey question", "fixture survey answer", "fixture-va-page-token", "fixture-va-response-token"},
			responseMarkers:   []string{"survey_id", "survey_name", "engagement_id", "consumer", "virtual_agent", "results", "next_page_token", "token"},
		},
		{
			name:              "list transcripts",
			path:              []string{"virtual-agent", "reports", "transcripts", "list"},
			requestPath:       "/v2/virtual_agent/report/transcripts",
			fixture:           "list_virtual_agent_transcripts.json",
			responseSensitive: []string{"fixture-va-engagement", "fixture transcript text", "Fixture Virtual Agent article", "https://fixture.invalid/virtual-agent/article", "fixture-va-page-token", "fixture-va-response-token"},
			responseMarkers:   []string{"engagement_id", "messages", "next_page_token", "token"},
		},
		{
			name:              "list operation logs",
			path:              []string{"virtual-agent", "reports", "operation-logs", "list"},
			requestPath:       "/v2/ai_studio/reports/operation_logs",
			fixture:           "list_virtual_agent_operation_logs.json",
			responseSensitive: []string{"fixture-va-account", "fixture action filter", "fixture-va-operation-log", "fixture-va-business", "fixture category filter", "fixture operation detail", "fixture-va-operator", "fixture operator info", "fixture-va-resource", "fixture subaction", "fixture-va-page-token", "fixture-va-response-token"},
			responseMarkers:   []string{"account_id", "action_filter_key", "ai_studio_operation_id", "business_id", "category_filter_key", "detail", "operator_id", "operator_info", "resource_id", "subaction", "next_page_token", "token"},
		},
	}
	byRequest := make(map[string]*virtualAgentReadAction, len(actions))
	for i := range actions {
		actions[i].status, actions[i].response = zoomDirectReadFixture(t, actions[i].fixture)
		byRequest[http.MethodGet+" "+actions[i].requestPath] = &actions[i]
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("Virtual Agent read fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("Virtual Agent read fixture did not receive the Zoom bearer credential")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("Virtual Agent read fixture received undeclared query or paging input")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(action.status)
		_, _ = w.Write(action.response)
	}))
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	config := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": server.URL + "/v2"},
		Secrets: map[string]string{"access_token": accessToken},
	}
	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
				Path:   action.path,
				Flags:  action.flags,
				Config: config,
			}, func(connectors.Record) error {
				t.Fatal("emit called for a Virtual Agent direct_read command")
				return nil
			})
			if err != nil {
				t.Fatalf("Run(%q): %v", strings.Join(action.path, " "), err)
			}
			if result.DirectRead == nil || result.DirectRead.Status != action.status {
				t.Fatalf("Run(%q) result = %#v, want status %d", strings.Join(action.path, " "), result.DirectRead, action.status)
			}
			encoded, err := json.Marshal(result.DirectRead.Body)
			if err != nil {
				t.Fatalf("marshal Virtual Agent read response: %v", err)
			}
			for index, raw := range action.responseSensitive {
				if strings.Contains(string(encoded), raw) {
					t.Fatalf("Virtual Agent read response exposed declared or generic sensitive field %d", index)
				}
			}
			for _, field := range action.responseMarkers {
				if !strings.Contains(string(encoded), "\""+field+"_redacted\":true") {
					t.Fatalf("Virtual Agent read response is missing %s_redacted marker", field)
				}
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("Virtual Agent read fixtures received %d requests, want %d", requests, len(actions))
	}
}

// TestVirtualAgentDirectWriteCommandsExecuteWithFixtures exercises every
// documented Virtual Agent mutation through the shared plan/preview/approval
// lifecycle. It proves typed article input, no-body sync creation, destructive
// article deletion, exact endpoint/auth/status contracts, and redacted output.
func TestVirtualAgentDirectWriteCommandsExecuteWithFixtures(t *testing.T) {
	const (
		credentialName = "zoom-virtual-agent-fixture"
		accessToken    = "fixture-virtual-agent-access-token"
	)
	type virtualAgentWriteAction struct {
		name              string
		path              []string
		flags             map[string][]string
		fixture           string
		method            string
		requestPath       string
		expectedBody      json.RawMessage
		destructive       bool
		inputSensitive    []string
		responseSensitive []string
		responseMarkers   []string
		status            int
		response          json.RawMessage
	}
	actions := []virtualAgentWriteAction{
		{
			name: "create article",
			path: []string{"virtual-agent", "knowledge-bases", "articles", "create"},
			flags: map[string][]string{
				"kb-id":       {"fixture-va-kb"},
				"content":     {"fixture Virtual Agent article content"},
				"exclude":     {"false"},
				"title":       {"Fixture Virtual Agent article"},
				"category":    {"fixture category"},
				"external-id": {"fixture-va-external-id"},
				"language":    {"en-US"},
				"url":         {"https://fixture.invalid/virtual-agent/article"},
			},
			fixture:           "create_virtual_agent_article.json",
			inputSensitive:    []string{"fixture Virtual Agent article content", "Fixture Virtual Agent article", "fixture category", "fixture-va-external-id", "https://fixture.invalid/virtual-agent/article"},
			responseSensitive: []string{"fixture-va-article", "fixture-va-kb", "fixture Virtual Agent article content", "Fixture Virtual Agent article", "fixture-va-response-token"},
			responseMarkers:   []string{"id", "kb_id", "content", "title", "token"},
		},
		{
			name: "update article",
			path: []string{"virtual-agent", "knowledge-bases", "articles", "update"},
			flags: map[string][]string{
				"kb-id":       {"fixture-va-kb"},
				"article-id":  {"fixture-va-article"},
				"content":     {"fixture updated Virtual Agent article content"},
				"exclude":     {"true"},
				"title":       {"Fixture updated Virtual Agent article"},
				"category":    {"fixture updated category"},
				"external-id": {"fixture-va-external-id"},
				"language":    {"en-US"},
				"url":         {"https://fixture.invalid/virtual-agent/article-updated"},
			},
			fixture:           "update_virtual_agent_article.json",
			inputSensitive:    []string{"fixture updated Virtual Agent article content", "Fixture updated Virtual Agent article", "fixture updated category", "fixture-va-external-id", "https://fixture.invalid/virtual-agent/article-updated"},
			responseSensitive: []string{"fixture-va-article", "fixture-va-kb", "fixture updated Virtual Agent article content", "Fixture updated Virtual Agent article", "fixture-va-response-token"},
			responseMarkers:   []string{"id", "kb_id", "content", "title", "token"},
		},
		{
			name:        "delete article",
			path:        []string{"virtual-agent", "knowledge-bases", "articles", "delete"},
			flags:       map[string][]string{"kb-id": {"fixture-va-kb"}, "article-id": {"fixture-va-article"}},
			fixture:     "delete_virtual_agent_article.json",
			destructive: true,
		},
		{
			name:              "create sync request",
			path:              []string{"virtual-agent", "knowledge-bases", "sync", "create"},
			flags:             map[string][]string{"kb-id": {"fixture-va-kb"}},
			fixture:           "create_virtual_agent_sync_request.json",
			responseSensitive: []string{"fixture-va-sync", "fixture-va-kb", "fixture Virtual Agent sync detail", "fixture-va-response-token"},
			responseMarkers:   []string{"sync_id", "kb_id", "error_message", "token"},
		},
	}
	byRequest := make(map[string]*virtualAgentWriteAction, len(actions))
	for i := range actions {
		fixture := zoomSCIM2WriteFixture(t, actions[i].fixture)
		actions[i].method = fixture.Expect.Method
		actions[i].requestPath = fixture.Expect.Path
		actions[i].expectedBody = fixture.Expect.Body
		actions[i].status = fixture.Response.Status
		actions[i].response = fixture.Response.Body
		byRequest[actions[i].method+" "+actions[i].requestPath] = &actions[i]
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("Virtual Agent write fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("Virtual Agent write fixture did not receive the Zoom bearer credential")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("Virtual Agent write fixture received undeclared query or paging input")
		}
		if len(action.expectedBody) == 0 {
			var unexpected any
			if err := json.NewDecoder(request.Body).Decode(&unexpected); err == nil {
				t.Fatal("Virtual Agent no-body action unexpectedly sent a request body")
			}
		} else {
			if !strings.HasPrefix(request.Header.Get("Content-Type"), "application/json") {
				t.Fatal("Virtual Agent body action did not declare JSON content")
			}
			var got, want any
			if err := json.NewDecoder(request.Body).Decode(&got); err != nil {
				t.Fatalf("decode Virtual Agent action body: %v", err)
			}
			if err := json.Unmarshal(action.expectedBody, &want); err != nil {
				t.Fatalf("decode Virtual Agent fixture body: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("Virtual Agent action body did not contain exactly the declared documented fields")
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
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	for _, action := range actions {
		if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{Path: action.path, Flags: action.flags}); err != nil {
			t.Fatalf("BuildWriteCommand(%s): %v", action.name, err)
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
		Config:    map[string]string{"base_url": server.URL + "/v2"},
		Secrets:   map[string]string{"access_token": accessToken},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			beforeRequests := requests
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
				t.Fatal("Virtual Agent plan did not produce a no-network preview and single-use approval")
			}
			if action.destructive != (plan.ConfirmationChallenge == string(connectors.ConfirmationKindDestructive)) {
				t.Fatal("Virtual Agent plan did not retain the declared destructive confirmation policy")
			}
			encodedPlan, err := json.Marshal(plan.Sample)
			if err != nil {
				t.Fatalf("marshal Virtual Agent plan sample: %v", err)
			}
			for index, raw := range action.inputSensitive {
				if strings.Contains(string(encodedPlan), raw) {
					t.Fatalf("Virtual Agent plan sample exposed declared sensitive input %d", index)
				}
			}
			if requests != beforeRequests {
				t.Fatal("Virtual Agent plan or preview reached the fixture endpoint")
			}

			if action.destructive {
				if _, err := application.RunReverseETL(context.Background(), app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}); err == nil {
					t.Fatal("Virtual Agent DELETE execution bypassed typed destructive confirmation")
				}
				if requests != beforeRequests {
					t.Fatal("unconfirmed Virtual Agent DELETE reached the fixture endpoint")
				}
			}

			run, err := application.RunReverseETL(context.Background(), app.RunReverseETLRequest{
				PlanID:        plan.ID,
				ApprovalToken: plan.ApprovalToken,
				Confirmation: connectors.WriteConfirmation{
					Kind: connectors.ConfirmationKindDestructive,
				},
			})
			if err != nil {
				t.Fatalf("RunReverseETL(%s): %v", action.name, err)
			}
			if run.Status != "completed" || run.OperationDirectWrite == nil || run.OperationDirectWrite.Status != action.status {
				t.Fatalf("Virtual Agent run = %#v, want completed declared action status %d", run, action.status)
			}
			if len(action.response) == 0 {
				if run.OperationDirectWrite.Body != nil {
					t.Fatal("Virtual Agent status-only action returned an invented response body")
				}
				return
			}
			encoded, err := json.Marshal(run.OperationDirectWrite.Body)
			if err != nil {
				t.Fatalf("marshal Virtual Agent action response: %v", err)
			}
			for index, raw := range action.responseSensitive {
				if strings.Contains(string(encoded), raw) {
					t.Fatalf("Virtual Agent action response exposed declared or generic sensitive field %d", index)
				}
			}
			for _, field := range action.responseMarkers {
				if !strings.Contains(string(encoded), "\""+field+"_redacted\":true") {
					t.Fatalf("Virtual Agent action response is missing %s_redacted marker", field)
				}
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("Virtual Agent write fixtures received %d requests, want %d", requests, len(actions))
	}
}

// TestSCIM2DirectReadCommandsExecuteWithFixtures runs every documented SCIM2
// read through the real command runner. The loopback server proves that the
// operation-scoped SCIM root/auth transport is used instead of the ordinary
// Zoom /v2 base, and that response-only SCIM paging fields never become
// hand-authored CLI flags.
func TestSCIM2DirectReadCommandsExecuteWithFixtures(t *testing.T) {
	const accessToken = "fixture-scim2-access-token"
	type scim2ReadAction struct {
		name              string
		path              []string
		flags             map[string][]string
		requestPath       string
		fixture           string
		status            int
		response          json.RawMessage
		responseSensitive []string
		responseMarkers   []string
	}
	actions := []scim2ReadAction{
		{
			name:              "list groups",
			path:              []string{"scim2", "groups", "list"},
			requestPath:       "/scim2/Groups",
			fixture:           "list_scim2_groups.json",
			responseSensitive: []string{"fixture-scim-group", "fixture SCIM group", "fixture-group-member", "Fixture group member", "fixture-group-alias", "fixture-scim-response-token"},
			responseMarkers:   []string{"id", "displayName", "members", "urn:ietf:params:scim:schemas:extension:zoom:2.0:Group", "token"},
		},
		{
			name:              "get group",
			path:              []string{"scim2", "groups", "get"},
			flags:             map[string][]string{"group-id": {"fixture-scim-group"}},
			requestPath:       "/scim2/Groups/fixture-scim-group",
			fixture:           "get_scim2_group.json",
			responseSensitive: []string{"fixture-scim-group", "fixture SCIM group", "fixture-group-member", "Fixture group member", "fixture-group-alias", "fixture-scim-response-token"},
			responseMarkers:   []string{"id", "displayName", "members", "urn:ietf:params:scim:schemas:extension:zoom:2.0:Group", "token"},
		},
		{
			name:              "list users",
			path:              []string{"scim2", "users", "list"},
			requestPath:       "/scim2/Users",
			fixture:           "list_scim2_users.json",
			responseSensitive: []string{"fixture-scim-user", "fixture.user@example.invalid", "Fixture SCIM user", "fixture-department", "fixture-custom-attribute", "fixture-scim-response-token"},
			responseMarkers:   []string{"id", "userName", "displayName", "emails", "name", "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User", "urn:ietf:params:scim:schemas:extension:zoom:2.0:User", "token"},
		},
		{
			name:              "get user",
			path:              []string{"scim2", "users", "get"},
			flags:             map[string][]string{"user-id": {"fixture-scim-user"}},
			requestPath:       "/scim2/Users/fixture-scim-user",
			fixture:           "get_scim2_user.json",
			responseSensitive: []string{"fixture-scim-user", "fixture.user@example.invalid", "Fixture SCIM user", "fixture-department", "fixture-custom-attribute", "fixture-scim-response-token"},
			responseMarkers:   []string{"id", "userName", "displayName", "emails", "name", "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User", "urn:ietf:params:scim:schemas:extension:zoom:2.0:User", "token"},
		},
	}
	for i := range actions {
		actions[i].status, actions[i].response = zoomDirectReadFixture(t, actions[i].fixture)
	}

	byRequest := make(map[string]*scim2ReadAction, len(actions))
	for i := range actions {
		byRequest[http.MethodGet+" "+actions[i].requestPath] = &actions[i]
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("SCIM2 read fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("SCIM2 read fixture did not receive the declared operation bearer")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("SCIM2 read fixture received undeclared query or paging input")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(action.status)
		_, _ = w.Write(action.response)
	}))
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	config := connectors.RuntimeConfig{
		Config: map[string]string{
			// Deliberately unrelated: every SCIM2 request must use the declared
			// operation root rather than the ordinary Zoom API base.
			"base_url":       "https://ordinary.zoom.invalid/v2",
			"scim2_base_url": server.URL,
		},
		Secrets: map[string]string{"access_token": accessToken},
	}

	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
				Path:   action.path,
				Flags:  action.flags,
				Config: config,
			}, func(connectors.Record) error {
				t.Fatal("emit called for a SCIM2 direct_read command")
				return nil
			})
			if err != nil {
				t.Fatalf("Run(%q): %v", strings.Join(action.path, " "), err)
			}
			if result.DirectRead == nil || result.DirectRead.Status != action.status {
				t.Fatalf("Run(%q) result = %#v, want status %d", strings.Join(action.path, " "), result, action.status)
			}
			encoded, err := json.Marshal(result.DirectRead.Body)
			if err != nil {
				t.Fatalf("marshal SCIM2 read response: %v", err)
			}
			for _, raw := range action.responseSensitive {
				if strings.Contains(string(encoded), raw) {
					t.Fatal("SCIM2 read response exposed a declared or generic sensitive field")
				}
			}
			for _, field := range action.responseMarkers {
				if !strings.Contains(string(encoded), `"`+field+`_redacted":true`) {
					t.Fatalf("SCIM2 read response is missing %s_redacted marker", field)
				}
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("SCIM2 read fixtures received %d requests, want %d", requests, len(actions))
	}
}

// TestSCIM2DirectWriteCommandsExecuteWithFixtures exercises every documented
// SCIM2 mutation through the shared plan/preview/approval lifecycle. It
// proves fixed methods, paths, root JSON-object body shaping, the separate
// SCIM origin/auth declaration, destructive deletion confirmation, response
// redaction, and 204 status-only semantics without reaching Zoom.
func TestSCIM2DirectWriteCommandsExecuteWithFixtures(t *testing.T) {
	const (
		credentialName = "zoom-scim2-fixture"
		accessToken    = "fixture-scim2-access-token"
	)
	type scim2WriteAction struct {
		name              string
		path              []string
		flags             map[string][]string
		rootFlag          string
		fixture           string
		method            string
		requestPath       string
		expectedBody      json.RawMessage
		status            int
		response          json.RawMessage
		destructive       bool
		inputSensitive    []string
		responseSensitive []string
		responseMarkers   []string
	}
	actions := []scim2WriteAction{
		{
			name:              "create group",
			path:              []string{"scim2", "groups", "create"},
			rootFlag:          "resource",
			fixture:           "create_scim2_group.json",
			inputSensitive:    []string{"fixture SCIM group", "fixture-group-alias"},
			responseSensitive: []string{"fixture-scim-group", "fixture SCIM group", "fixture-group-alias", "fixture-scim-response-token"},
			responseMarkers:   []string{"id", "displayName", "urn:ietf:params:scim:schemas:extension:zoom:2.0:Group", "token"},
		},
		{
			name:              "delete group",
			path:              []string{"scim2", "groups", "delete"},
			flags:             map[string][]string{"group-id": {"fixture-scim-group"}},
			fixture:           "delete_scim2_group.json",
			destructive:       true,
			responseSensitive: nil,
		},
		{
			name:           "update group",
			path:           []string{"scim2", "groups", "update"},
			flags:          map[string][]string{"group-id": {"fixture-scim-group"}},
			rootFlag:       "patch",
			fixture:        "update_scim2_group.json",
			inputSensitive: []string{"fixture updated SCIM group"},
		},
		{
			name:              "create user",
			path:              []string{"scim2", "users", "create"},
			rootFlag:          "resource",
			fixture:           "create_scim2_user.json",
			inputSensitive:    []string{"fixture.user@example.invalid", "Fixture SCIM user", "fixture-custom-attribute"},
			responseSensitive: []string{"fixture-scim-user", "fixture.user@example.invalid", "Fixture SCIM user", "fixture-scim-response-token"},
			responseMarkers:   []string{"id", "userName", "displayName", "emails", "token"},
		},
		{
			name:              "update user",
			path:              []string{"scim2", "users", "update"},
			flags:             map[string][]string{"user-id": {"fixture-scim-user"}},
			rootFlag:          "resource",
			fixture:           "update_scim2_user.json",
			inputSensitive:    []string{"fixture.user@example.invalid", "Fixture updated SCIM user"},
			responseSensitive: []string{"fixture-scim-user", "fixture.user@example.invalid", "Fixture updated SCIM user", "fixture-scim-response-token"},
			responseMarkers:   []string{"id", "userName", "displayName", "name", "token"},
		},
		{
			name:        "delete user",
			path:        []string{"scim2", "users", "delete"},
			flags:       map[string][]string{"user-id": {"fixture-scim-user"}},
			fixture:     "delete_scim2_user.json",
			destructive: true,
		},
		{
			name:              "deactivate user",
			path:              []string{"scim2", "users", "deactivate"},
			flags:             map[string][]string{"user-id": {"fixture-scim-user"}},
			rootFlag:          "patch",
			fixture:           "deactivate_scim2_user.json",
			responseSensitive: []string{"fixture-scim-user", "fixture.user@example.invalid", "fixture-scim-response-token"},
			responseMarkers:   []string{"id", "userName", "active", "token"},
		},
	}
	byRequest := make(map[string]*scim2WriteAction, len(actions))
	for i := range actions {
		fixture := zoomSCIM2WriteFixture(t, actions[i].fixture)
		actions[i].method = fixture.Expect.Method
		actions[i].requestPath = fixture.Expect.Path
		actions[i].expectedBody = fixture.Expect.Body
		actions[i].status = fixture.Response.Status
		actions[i].response = fixture.Response.Body
		if actions[i].rootFlag != "" {
			if actions[i].flags == nil {
				actions[i].flags = map[string][]string{}
			}
			var record any
			if err := json.Unmarshal(fixture.Record, &record); err != nil {
				t.Fatalf("decode SCIM2 fixture record %s: %v", actions[i].fixture, err)
			}
			compactRecord, err := json.Marshal(record)
			if err != nil {
				t.Fatalf("compact SCIM2 fixture record %s: %v", actions[i].fixture, err)
			}
			actions[i].flags[actions[i].rootFlag] = []string{string(compactRecord)}
		}
		byRequest[actions[i].method+" "+actions[i].requestPath] = &actions[i]
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests++
		action := byRequest[request.Method+" "+request.URL.Path]
		if action == nil {
			t.Fatal("SCIM2 write fixture received an undeclared method/path")
		}
		if request.Header.Get("Authorization") != "Bearer "+accessToken {
			t.Fatal("SCIM2 write fixture did not receive the declared operation bearer")
		}
		if len(request.URL.Query()) != 0 {
			t.Fatal("SCIM2 write fixture received undeclared query or paging input")
		}
		if len(action.expectedBody) == 0 {
			var unexpected any
			if err := json.NewDecoder(request.Body).Decode(&unexpected); err == nil {
				t.Fatal("SCIM2 no-body action unexpectedly sent a request body")
			}
		} else {
			if !strings.HasPrefix(request.Header.Get("Content-Type"), "application/json") {
				t.Fatal("SCIM2 body action did not declare JSON content")
			}
			var got, want any
			if err := json.NewDecoder(request.Body).Decode(&got); err != nil {
				t.Fatalf("decode SCIM2 action body: %v", err)
			}
			if err := json.Unmarshal(action.expectedBody, &want); err != nil {
				t.Fatalf("decode SCIM2 fixture body: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("SCIM2 action body did not contain exactly the declared documented object")
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
	defer server.Close()

	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	for _, action := range actions {
		if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{Path: action.path, Flags: action.flags}); err != nil {
			t.Fatalf("BuildWriteCommand(%s): %v", action.name, err)
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
			"base_url":       "https://ordinary.zoom.invalid/v2",
			"scim2_base_url": server.URL,
		},
		Secrets: map[string]string{"access_token": accessToken},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			beforeRequests := requests
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
				t.Fatal("SCIM2 plan did not produce a no-network preview and single-use approval")
			}
			if action.destructive != (plan.ConfirmationChallenge == string(connectors.ConfirmationKindDestructive)) {
				t.Fatal("SCIM2 plan did not retain the declared destructive confirmation policy")
			}
			encodedPlan, err := json.Marshal(plan.Sample)
			if err != nil {
				t.Fatalf("marshal SCIM2 plan sample: %v", err)
			}
			for index, raw := range action.inputSensitive {
				if strings.Contains(string(encodedPlan), raw) {
					t.Fatalf("SCIM2 plan sample exposed declared sensitive input %d", index)
				}
			}
			if requests != beforeRequests {
				t.Fatal("SCIM2 plan or preview reached the fixture endpoint")
			}

			runRequest := app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}
			if action.destructive {
				if _, err := application.RunReverseETL(context.Background(), runRequest); err == nil {
					t.Fatal("SCIM2 DELETE execution bypassed typed destructive confirmation")
				}
				if requests != beforeRequests {
					t.Fatal("unconfirmed SCIM2 DELETE reached the fixture endpoint")
				}
				runRequest.Confirmation = connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
			}
			run, err := application.RunReverseETL(context.Background(), runRequest)
			if err != nil {
				t.Fatalf("RunReverseETL(%s): %v", action.name, err)
			}
			if run.Status != "completed" || run.OperationDirectWrite == nil || run.OperationDirectWrite.Status != action.status {
				t.Fatal("SCIM2 action did not complete with its declared response status")
			}
			if len(action.response) == 0 {
				if run.OperationDirectWrite.Body != nil {
					t.Fatal("SCIM2 status-only action returned an invented response body")
				}
				return
			}
			encoded, err := json.Marshal(run.OperationDirectWrite.Body)
			if err != nil {
				t.Fatalf("marshal SCIM2 action response: %v", err)
			}
			for _, raw := range action.responseSensitive {
				if strings.Contains(string(encoded), raw) {
					t.Fatal("SCIM2 action response exposed a declared or generic sensitive field")
				}
			}
			for _, field := range action.responseMarkers {
				if !strings.Contains(string(encoded), `"`+field+`_redacted":true`) {
					t.Fatalf("SCIM2 action response is missing %s_redacted marker", field)
				}
			}
		})
	}
	if requests != len(actions) {
		t.Fatalf("SCIM2 write fixtures received %d requests, want %d", requests, len(actions))
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

type zoomSCIM2WriteFixtureSpec struct {
	Record json.RawMessage `json:"record"`
	Expect struct {
		Method string          `json:"method"`
		Path   string          `json:"path"`
		Body   json.RawMessage `json:"body"`
	} `json:"expect"`
	Response struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	} `json:"response"`
}

func zoomSCIM2WriteFixture(t *testing.T, file string) zoomSCIM2WriteFixtureSpec {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("fixtures", "writes", file))
	if err != nil {
		t.Fatalf("read SCIM2 write fixture %s: %v", file, err)
	}
	var fixture zoomSCIM2WriteFixtureSpec
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode SCIM2 write fixture %s: %v", file, err)
	}
	if fixture.Expect.Method == "" || fixture.Expect.Path == "" || fixture.Response.Status == 0 {
		t.Fatalf("SCIM2 write fixture %s lacks expected method/path/response status", file)
	}
	return fixture
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
