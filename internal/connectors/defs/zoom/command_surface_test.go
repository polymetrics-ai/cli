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
	covered, coveredWrites, implementableNow, providerRestricted, deprecated := 0, 0, 0, 0, 0

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
			if endpoint.CoveredBy.Stream == "" && len(endpoint.CoveredBy.WriteTargets()) == 0 {
				t.Errorf("executable %s is not bound to a stream or named write action", key)
			}
			coveredWrites += len(endpoint.CoveredBy.WriteTargets())
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
	if got := covered; got != 5 {
		t.Errorf("executable provider rows = %d, want 5 (three ETL streams and two reverse-ETL actions)", got)
	}
	if got := coveredWrites; got != 2 {
		t.Errorf("executable write-backed rows = %d, want 2", got)
	}
	if got := implementableNow; got != 1837 {
		t.Errorf("operations awaiting Zoom-local contracts = %d, want 1837", got)
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

// TestSourceBackedOperationInventoryKeepsEveryContractVisible proves that the
// provider source is not silently reduced to the small executable cohort. The
// 34 multipart uploads are deliberately absent: the current file_upload
// executor cannot carry their provider contract, and the committed rejection
// ledger records each one as a recoverable foundation gap. The 70 direct reads
// remain owned by the already-delivered #4267 sub-PR and the two warehouse
// actions live in writes.json, so neither is duplicated here. The api_surface
// ledger remains the dispatch gate: these declarations have no command until
// their individual policy, fixture, and certification work is complete.
func TestSourceBackedOperationInventoryKeepsEveryContractVisible(t *testing.T) {
	bundle := loadZoomBundle(t)

	if got := len(bundle.Operations); got != 1748 {
		t.Fatalf("source-backed operation contracts = %d, want 1748 non-deprecated exact matches excluding 34 multipart foundation gaps, 70 #4267 reads, and two writes.json actions", got)
	}

	counts := map[string]int{}
	deletes := 0
	for _, operation := range bundle.Operations {
		counts[operation.Kind]++
		if operation.Kind == "rest_write" && operation.MutationClass == "delete" {
			deletes++
			if !operation.Destructive || operation.Confirmation == nil || operation.Confirmation.Kind != "destructive" {
				t.Errorf("delete operation %q lacks the destructive confirmation contract", operation.ID)
			}
		}
	}

	if got := counts["rest_read"]; got != 776 {
		t.Errorf("rest_read source contracts = %d, want 776", got)
	}
	if got := counts["rest_write"]; got != 971 {
		t.Errorf("rest_write source contracts = %d, want 971", got)
	}
	if got := counts["binary_download"]; got != 1 {
		t.Errorf("binary_download source contracts = %d, want 1", got)
	}
	if got := deletes; got != 311 {
		t.Errorf("delete write contracts = %d, want 311", got)
	}
}

// TestPinnedSourceCrosswalkAccountsForEveryIdentity checks the committed
// evidence directly. It deliberately needs no network or provider credential:
// the source lock preserves the public-document digests while the crosswalk
// preserves every normalized operation fact used by the declaration inventory.
func TestPinnedSourceCrosswalkAccountsForEveryIdentity(t *testing.T) {
	var lock struct {
		Connector string `json:"connector"`
		Rest      struct {
			SHA256          string `json:"sha256"`
			Bytes           int64  `json:"bytes"`
			SourceDocuments []struct {
				SHA256 string `json:"sha256"`
				Bytes  int64  `json:"bytes"`
			} `json:"source_documents"`
			Operations []struct {
				ID             string `json:"id"`
				Method         string `json:"method"`
				Path           string `json:"path"`
				SourceLocation string `json:"source_location"`
			} `json:"operations"`
		} `json:"rest"`
	}
	lockRaw, err := os.ReadFile(filepath.Join("sources", "zoom-operation-source-lock.json"))
	if err != nil {
		t.Fatalf("read source lock: %v", err)
	}
	if err := json.Unmarshal(lockRaw, &lock); err != nil {
		t.Fatalf("decode source lock: %v", err)
	}
	if lock.Connector != zoomBundleName || len(lock.Rest.SHA256) != 64 || lock.Rest.Bytes != 12127228 {
		t.Fatalf("source lock identity = connector:%q digest:%q bytes:%d", lock.Connector, lock.Rest.SHA256, lock.Rest.Bytes)
	}
	if got := len(lock.Rest.SourceDocuments); got != 35 {
		t.Errorf("source lock documents = %d, want 35", got)
	}
	if got := len(lock.Rest.Operations); got != 1937 {
		t.Errorf("source lock operations = %d, want 1937", got)
	}
	for _, operation := range lock.Rest.Operations {
		if operation.ID == "" || operation.Method == "" || operation.Path == "" || operation.SourceLocation == "" {
			t.Errorf("source lock has incomplete operation identity: %+v", operation)
		}
	}

	var crosswalk struct {
		Accounting struct {
			SourceOperations     int `json:"source_operations"`
			APISurfaceEndpoints  int `json:"api_surface_endpoints"`
			ExactSourceToSurface int `json:"exact_source_to_surface"`
			SourceOnly           int `json:"source_only"`
			SurfaceOnly          int `json:"surface_only"`
		} `json:"accounting"`
		SourceOperations []struct {
			ID             string `json:"id"`
			Method         string `json:"method"`
			Path           string `json:"path"`
			SourceLocation string `json:"source_location"`
			Crosswalk      struct {
				State      string `json:"state"`
				APISurface *struct {
					Method string `json:"method"`
					Path   string `json:"path"`
				} `json:"api_surface"`
			} `json:"crosswalk"`
		} `json:"source_operations"`
	}
	crosswalkRaw, err := os.ReadFile(filepath.Join("sources", "zoom-operation-crosswalk.json"))
	if err != nil {
		t.Fatalf("read source crosswalk: %v", err)
	}
	if err := json.Unmarshal(crosswalkRaw, &crosswalk); err != nil {
		t.Fatalf("decode source crosswalk: %v", err)
	}
	if got, want := crosswalk.Accounting.SourceOperations, 1937; got != want {
		t.Errorf("crosswalk source operations = %d, want %d", got, want)
	}
	if got, want := crosswalk.Accounting.APISurfaceEndpoints, 1913; got != want {
		t.Errorf("crosswalk api surface endpoints = %d, want %d", got, want)
	}
	if got, want := crosswalk.Accounting.ExactSourceToSurface, 1911; got != want {
		t.Errorf("crosswalk exact matches = %d, want %d", got, want)
	}
	if got, want := crosswalk.Accounting.SourceOnly, 26; got != want {
		t.Errorf("crosswalk source-only rows = %d, want %d", got, want)
	}
	if got, want := crosswalk.Accounting.SurfaceOnly, 2; got != want {
		t.Errorf("crosswalk surface-only rows = %d, want %d", got, want)
	}

	seenSource := map[string]bool{}
	seenExactSurface := map[string]bool{}
	exact, sourceOnly := 0, 0
	for _, operation := range crosswalk.SourceOperations {
		if operation.ID == "" || operation.Method == "" || operation.Path == "" || operation.SourceLocation == "" {
			t.Errorf("crosswalk has incomplete source identity: %+v", operation)
			continue
		}
		if seenSource[operation.ID] {
			t.Errorf("crosswalk repeats source operation %q", operation.ID)
		}
		seenSource[operation.ID] = true
		switch operation.Crosswalk.State {
		case "exact":
			exact++
			if operation.Crosswalk.APISurface == nil || operation.Crosswalk.APISurface.Method != operation.Method || operation.Crosswalk.APISurface.Path != operation.Path {
				t.Errorf("exact crosswalk %q has no exact api surface identity", operation.ID)
				continue
			}
			key := operation.Method + " " + operation.Path
			if seenExactSurface[key] {
				t.Errorf("crosswalk repeats exact api surface identity %s", key)
			}
			seenExactSurface[key] = true
		case "source_only":
			sourceOnly++
			if operation.Crosswalk.APISurface != nil {
				t.Errorf("source-only crosswalk %q unexpectedly binds an api surface row", operation.ID)
			}
		default:
			t.Errorf("crosswalk %q has unexpected state %q", operation.ID, operation.Crosswalk.State)
		}
	}
	if got := len(seenSource); got != 1937 {
		t.Errorf("unique source identities = %d, want 1937", got)
	}
	if exact != 1911 || sourceOnly != 26 {
		t.Errorf("crosswalk state counts = exact:%d source_only:%d, want exact:1911 source_only:26", exact, sourceOnly)
	}
}

// TestDeclarationDispositionAccountsForThePinnedSourceAndLedger keeps the
// declare-or-disable audit executable without adding connector-specific
// generator code outside this bundle. A source-contract inventory declaration
// is intentionally distinct from a terminal command: every disabled row must
// still give an operator a fixed-vocabulary reason, provider evidence, and a
// recovery answer.
func TestDeclarationDispositionAccountsForThePinnedSourceAndLedger(t *testing.T) {
	type rejection struct {
		Reason      string `json:"reason"`
		Recoverable bool   `json:"recoverable"`
		Detail      string `json:"detail"`
		Evidence    any    `json:"evidence"`
	}
	type disposition struct {
		Method      string     `json:"method"`
		Path        string     `json:"path"`
		State       string     `json:"state"`
		Rejection   *rejection `json:"rejection"`
		Declaration struct {
			ID string `json:"id"`
		} `json:"declaration"`
	}
	var report struct {
		Summary struct {
			APISurfaceRows                   int `json:"api_surface_rows"`
			SourceOnlyRows                   int `json:"source_only_rows"`
			DeclaredCurrentBranch            int `json:"declared_current_branch"`
			DeclaredPendingParentIntegration int `json:"declared_pending_parent_integration"`
			DisabledAPISurfaceRows           int `json:"disabled_api_surface_rows"`
			DisabledSourceOnlyRows           int `json:"disabled_source_only_rows"`
			OperationInventoryEntries        int `json:"operation_inventory_entries"`
			DeleteOperationInventoryEntries  int `json:"delete_operation_inventory_entries"`
		} `json:"summary"`
		LedgerDispositions     []disposition `json:"ledger_dispositions"`
		SourceOnlyDispositions []disposition `json:"source_only_dispositions"`
	}
	raw, err := os.ReadFile(filepath.Join("sources", "zoom-declaration-disposition.json"))
	if err != nil {
		t.Fatalf("read declaration disposition: %v", err)
	}
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatalf("decode declaration disposition: %v", err)
	}

	if got := report.Summary.APISurfaceRows; got != 1913 {
		t.Errorf("api_surface disposition rows = %d, want 1913", got)
	}
	if got := len(report.LedgerDispositions); got != 1913 {
		t.Errorf("ledger dispositions = %d, want 1913", got)
	}
	if got := report.Summary.SourceOnlyRows; got != 26 {
		t.Errorf("source-only dispositions = %d, want 26", got)
	}
	if got := len(report.SourceOnlyDispositions); got != 26 {
		t.Errorf("source-only records = %d, want 26", got)
	}
	if got := report.Summary.DeclaredCurrentBranch; got != 5 {
		t.Errorf("current declared rows = %d, want 5", got)
	}
	if got := report.Summary.DeclaredPendingParentIntegration; got != 70 {
		t.Errorf("parent-pending declarations = %d, want 70", got)
	}
	if got := report.Summary.DisabledAPISurfaceRows; got != 1838 {
		t.Errorf("disabled api surface rows = %d, want 1838", got)
	}
	if got := report.Summary.DisabledSourceOnlyRows; got != 26 {
		t.Errorf("disabled source-only rows = %d, want 26", got)
	}
	if got := report.Summary.OperationInventoryEntries; got != 1748 {
		t.Errorf("operation inventory entries = %d, want 1748", got)
	}
	if got := report.Summary.DeleteOperationInventoryEntries; got != 311 {
		t.Errorf("delete operation inventory entries = %d, want 311", got)
	}

	allowedReasons := map[string]bool{
		"provider-does-not-expose": true,
		"requires-paid-tier":       true,
		"requires-elevated-scope":  true,
		"foundation-gap":           true,
		"schema-incompatible":      true,
		"unsafe-to-exercise":       true,
		"needs-human-decision":     true,
	}
	seen := map[string]bool{}
	operationDeclarations := 0
	check := func(label string, row disposition) {
		if row.Method == "" || row.Path == "" {
			t.Errorf("%s has incomplete method/path: %+v", label, row)
		}
		key := row.Method + " " + row.Path
		if label == "ledger" {
			if seen[key] {
				t.Errorf("ledger disposition repeats %s", key)
			}
			seen[key] = true
		}
		if row.State != "disabled" {
			return
		}
		if row.Rejection == nil {
			t.Errorf("disabled %s %s has no rejection", label, key)
			return
		}
		if !allowedReasons[row.Rejection.Reason] {
			t.Errorf("disabled %s %s uses unsupported reason %q", label, key, row.Rejection.Reason)
		}
		if strings.TrimSpace(row.Rejection.Detail) == "" || row.Rejection.Evidence == nil {
			t.Errorf("disabled %s %s lacks evidence/detail: %+v", label, key, row.Rejection)
		}
		if row.Declaration.ID != "" {
			operationDeclarations++
		}
	}
	for _, row := range report.LedgerDispositions {
		check("ledger", row)
	}
	for _, row := range report.SourceOnlyDispositions {
		check("source-only", row)
	}
	if got := len(seen); got != 1913 {
		t.Errorf("unique ledger dispositions = %d, want 1913", got)
	}
	if got := operationDeclarations; got != 1748 {
		t.Errorf("disabled source-contract operation declarations = %d, want 1748", got)
	}
}

// TestSourceBackedReverseETLActionsUseSanitizedFixtures proves the two actual
// warehouse destination actions through the real command surface and the
// engine's HTTP writer against a loopback fixture server. It deliberately
// never invokes Zoom or resolves a real credential.
func TestSourceBackedReverseETLActionsUseSanitizedFixtures(t *testing.T) {
	bundle := loadZoomBundle(t)
	if got := len(bundle.Writes); got != 2 {
		t.Fatalf("Zoom warehouse destination actions = %d, want 2", got)
	}
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))

	tests := []struct {
		name    string
		action  string
		command []string
		flags   map[string][]string
	}{
		{
			name:    "update clinical note",
			action:  "update_clinical_note",
			command: []string{"healthcare", "clinical-notes", "update"},
			flags: map[string][]string{
				"note-id":           {"fixture-note"},
				"is-note-completed": {"true"},
			},
		},
		{
			name:    "create quality management interaction",
			action:  "create_quality_management_interaction",
			command: []string{"quality-management", "interactions", "create"},
			flags: map[string][]string{
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
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := loadZoomWriteFixture(t, test.action)
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				requests++
				if got, want := request.Method, fixture.Expect.Method; got != want {
					t.Errorf("method = %s, want %s", got, want)
				}
				if got, want := request.URL.Path, "/v2"+fixture.Expect.Path; got != want {
					t.Errorf("path = %s, want %s", got, want)
				}
				if len(fixture.Expect.Body) > 0 {
					var body map[string]any
					if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
						t.Errorf("decode request body: %v", err)
					} else if !reflect.DeepEqual(body, fixture.Expect.Body) {
						t.Errorf("request body = %#v, want %#v", body, fixture.Expect.Body)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(fixture.Status())
				if len(fixture.Response.Body) > 0 {
					_, _ = w.Write(fixture.Response.Body)
				}
			}))
			defer server.Close()

			config := connectors.RuntimeConfig{
				Config:  map[string]string{"base_url": server.URL + "/v2"},
				Secrets: map[string]string{"access_token": "synthetic-test-token"},
			}
			if err := commandrunner.Preflight(connector, test.command); err != nil {
				t.Fatalf("Preflight(%q) = %v", strings.Join(test.command, " "), err)
			}
			plan, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{
				Path: test.command, Flags: test.flags, Config: config, Preview: true,
			})
			if err != nil {
				t.Fatalf("BuildWriteCommand(%q) = %v", strings.Join(test.command, " "), err)
			}
			if !plan.ApprovalRequired || plan.Preview == nil || plan.Preview.RecordsStaged != 1 {
				t.Fatalf("BuildWriteCommand(%q) = %#v, want one staged, approval-gated record", strings.Join(test.command, " "), plan)
			}
			if requests != 0 {
				t.Fatalf("plan/preview made %d fixture requests, want 0", requests)
			}

			result, err := connector.Write(context.Background(), connectors.WriteRequest{Action: test.action, Config: config}, []connectors.Record{fixture.Record})
			if err != nil {
				t.Fatalf("Write(%q) = %v", test.action, err)
			}
			if result.RecordsWritten != 1 || result.RecordsFailed != 0 || requests != 1 {
				t.Fatalf("Write(%q) result=%#v fixture requests=%d, want one successful request", test.action, result, requests)
			}
		})
	}
}

type zoomWriteFixture struct {
	Record connectors.Record `json:"record"`
	Expect struct {
		Method string         `json:"method"`
		Path   string         `json:"path"`
		Status int            `json:"status"`
		Body   map[string]any `json:"body"`
	} `json:"expect"`
	Response struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	} `json:"response"`
}

func (fixture zoomWriteFixture) Status() int {
	if fixture.Expect.Status != 0 {
		return fixture.Expect.Status
	}
	return fixture.Response.Status
}

func loadZoomWriteFixture(t *testing.T, action string) zoomWriteFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("fixtures", "writes", action+".json"))
	if err != nil {
		t.Fatalf("read fixture for %q: %v", action, err)
	}
	var fixture zoomWriteFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode fixture for %q: %v", action, err)
	}
	if fixture.Expect.Method == "" || fixture.Expect.Path == "" || fixture.Status() == 0 {
		t.Fatalf("fixture for %q has an incomplete request/response contract: %#v", action, fixture)
	}
	return fixture
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
	if got := len(surface.Commands); got != len(wants)+2 {
		t.Fatalf("Zoom cli_surface commands = %d, want %d existing ETL commands plus two reverse-ETL actions", got, len(wants))
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

	for path := range commands {
		if _, ok := map[string]struct{}{
			"users list":                             {},
			"meetings list":                          {},
			"webinars list":                          {},
			"healthcare clinical-notes update":       {},
			"quality-management interactions create": {},
		}[path]; !ok {
			t.Errorf("Zoom mapping must not promote an untracked terminal operation; found %q", path)
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
