// Package zoom holds connector-local reachability evidence for the Zoom
// declarative bundle. The bundle itself is JSON; this test keeps its provider
// inventory and executable command surface from drifting apart.
package zoom

import (
	"bytes"
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

	"polymetrics.ai/internal/cli"
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
	covered, coveredDirectReads, coveredWrites, implementableNow, providerRestricted, deprecated := 0, 0, 0, 0, 0, 0

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
			if endpoint.CoveredBy.Stream == "" && len(endpoint.CoveredBy.WriteTargets()) == 0 && endpoint.CoveredBy.DirectRead == "" && len(endpoint.CoveredBy.DirectReads) == 0 && len(endpoint.CoveredBy.OperationTargets()) == 0 {
				t.Errorf("executable %s is not bound to a stream, direct read, operation, or named write action", key)
			}
			coveredDirectReads += len(endpoint.CoveredBy.DirectReads)
			if endpoint.CoveredBy.DirectRead != "" {
				coveredDirectReads++
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
	if got := covered; got != 712 {
		t.Errorf("executable provider rows = %d, want 712", got)
	}
	if got := coveredDirectReads; got != 505 {
		t.Errorf("executable direct-read-backed rows = %d, want 505", got)
	}
	if got := coveredWrites; got != 204 {
		t.Errorf("executable write-backed rows = %d, want 204", got)
	}
	if got := implementableNow; got != 1134 {
		t.Errorf("operations still blocked after declared runnable coverage = %d, want 1134", got)
	}
	if got := providerRestricted; got != 13 {
		t.Errorf("provider-restricted operations = %d, want 13", got)
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
			ImplementedPendingCertification  int `json:"implemented_pending_certification"`
			RunnableCLICommands              int `json:"runnable_cli_commands"`
			RunnableWriteActions             int `json:"runnable_write_actions"`
			RunnableDeleteActions            int `json:"runnable_delete_actions"`
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
	if got := report.Summary.DisabledAPISurfaceRows; got != 1131 {
		t.Errorf("disabled api surface rows = %d, want 1131", got)
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
	if got := report.Summary.ImplementedPendingCertification; got != 707 {
		t.Errorf("implemented pending certification rows = %d, want 707", got)
	}
	if got := report.Summary.RunnableCLICommands; got != 712 {
		t.Errorf("runnable CLI commands = %d, want 712", got)
	}
	if got := report.Summary.RunnableWriteActions; got != 204 {
		t.Errorf("runnable write actions = %d, want 204", got)
	}
	if got := report.Summary.RunnableDeleteActions; got != 185 {
		t.Errorf("runnable delete actions = %d, want 185", got)
	}

	allowedReasons := map[string]bool{
		"provider-does-not-expose": true,
		"requires-paid-tier":       true,
		"foundation-gap":           true,
		"schema-incompatible":      true,
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
		if row.Declaration.ID != "" {
			operationDeclarations++
		}
		switch row.State {
		case "disabled":
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
		case "declared", "declared-pending-certification", "declared-pending-parent-integration", "implemented-pending-certification":
			if row.Rejection != nil {
				t.Errorf("enabled %s %s unexpectedly has rejection %+v", label, key, row.Rejection)
			}
		default:
			t.Errorf("%s %s has unsupported state %q", label, key, row.State)
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
		t.Errorf("source-contract operation declarations = %d, want 1748", got)
	}
}

// TestSourceBackedReverseETLActionsUseSanitizedFixtures proves the two actual
// warehouse destination actions through the real command surface and the
// engine's HTTP writer against a loopback fixture server. It deliberately
// never invokes Zoom or resolves a real credential.
func TestSourceBackedReverseETLActionsUseSanitizedFixtures(t *testing.T) {
	bundle := loadZoomBundle(t)
	if got := len(bundle.Writes); got != 204 {
		t.Fatalf("Zoom typed write actions = %d, want 204", got)
	}
	for _, name := range []string{"update_clinical_note", "create_quality_management_interaction"} {
		if _, ok := findZoomWriteAction(bundle.Writes, name); !ok {
			t.Fatalf("Zoom source-backed fixture action %q is missing", name)
		}
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

func findZoomWriteAction(actions []engine.WriteAction, name string) (engine.WriteAction, bool) {
	for _, action := range actions {
		if action.Name == name {
			return action, true
		}
	}
	return engine.WriteAction{}, false
}

// TestEveryTypedZoomActionHasReverseETLCommandAndCandidate preserves the
// connector-owned inputs that a future generic typed destination can select.
// It deliberately declares neither a destination transport nor a transport
// binding: until #4303 supplies the connector-neutral executor and schema,
// either claim would be a false declaration.
func TestEveryTypedZoomActionHasReverseETLCommandAndCandidate(t *testing.T) {
	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))

	actions := make(map[string]engine.WriteAction, len(bundle.Writes))
	for _, action := range bundle.Writes {
		if action.Name == "" || action.Kind == "" || action.Method == "" || action.Path == "" {
			t.Errorf("typed write action is incomplete: %+v", action)
			continue
		}
		if _, duplicate := actions[action.Name]; duplicate {
			t.Errorf("typed write action %q is duplicated", action.Name)
		}
		actions[action.Name] = action
	}
	if got := len(actions); got != 204 {
		t.Fatalf("typed Zoom write actions = %d, want 204", got)
	}

	commands := make(map[string]connectors.CommandSurfaceCommand, len(actions))
	for _, command := range connector.CommandSurface().Commands {
		if command.Intent != "reverse_etl" {
			continue
		}
		if command.Availability != "implemented" || command.Write == "" {
			t.Errorf("reverse-ETL command %q = availability:%q write:%q, want implemented named action", command.Path, command.Availability, command.Write)
			continue
		}
		if _, duplicate := commands[command.Write]; duplicate {
			t.Errorf("typed action %q is selected by more than one reverse-ETL command", command.Write)
		}
		commands[command.Write] = command
	}
	if got := len(commands); got != len(actions) {
		t.Errorf("reverse-ETL commands with distinct actions = %d, want %d", got, len(actions))
	}

	var candidates struct {
		MutationCandidates []struct {
			Command     string `json:"command"`
			Intent      string `json:"intent"`
			Declaration struct {
				Kind     string `json:"kind"`
				ID       string `json:"id"`
				Executor string `json:"executor"`
			} `json:"declaration"`
			Classification struct {
				Code   string `json:"code"`
				Family string `json:"family"`
			} `json:"classification"`
		} `json:"mutation_candidates"`
	}
	raw, err := os.ReadFile("certification-mutation-candidates.json")
	if err != nil {
		t.Fatalf("read mutation candidates: %v", err)
	}
	if err := json.Unmarshal(raw, &candidates); err != nil {
		t.Fatalf("decode mutation candidates: %v", err)
	}
	candidatesByAction := make(map[string]struct {
		command  string
		intent   string
		kind     string
		executor string
		code     string
		family   string
	}, len(candidates.MutationCandidates))
	for _, candidate := range candidates.MutationCandidates {
		if candidate.Declaration.ID == "" {
			t.Errorf("mutation candidate %q has no typed action ID", candidate.Command)
			continue
		}
		if _, duplicate := candidatesByAction[candidate.Declaration.ID]; duplicate {
			t.Errorf("typed action %q has more than one mutation candidate", candidate.Declaration.ID)
		}
		candidatesByAction[candidate.Declaration.ID] = struct {
			command  string
			intent   string
			kind     string
			executor string
			code     string
			family   string
		}{candidate.Command, candidate.Intent, candidate.Declaration.Kind, candidate.Declaration.Executor, candidate.Classification.Code, candidate.Classification.Family}
	}
	if got := len(candidatesByAction); got != len(actions) {
		t.Errorf("mutation candidates with distinct typed actions = %d, want %d", got, len(actions))
	}

	for name, action := range actions {
		command, ok := commands[name]
		if !ok {
			t.Errorf("typed action %q has no reverse-ETL command", name)
			continue
		}
		if len(command.APISurface) != 1 || command.APISurface[0].Method != action.Method {
			t.Errorf("typed action %q command %q API surface = %+v, want one %s endpoint", name, command.Path, command.APISurface, action.Method)
		}
		candidate, ok := candidatesByAction[name]
		if !ok {
			t.Errorf("typed action %q has no mutation candidate", name)
			continue
		}
		if candidate.command != command.Path || candidate.intent != "reverse_etl" || candidate.kind != "write_action" || candidate.executor != "reverse_plan" {
			t.Errorf("typed action %q candidate = %+v, want command:%q reverse_etl write_action/reverse_plan", name, candidate, command.Path)
		}
		if candidate.code != "unassessed" || candidate.family != "generic_typed_destination_executor_deferred" {
			t.Errorf("typed action %q candidate classification = code:%q family:%q, want deferred generic typed destination", name, candidate.code, candidate.family)
		}
	}
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

}

// TestSourceTransportDeclaresEveryExecutableZoomStream keeps the merged
// definition-owned transport adapter honest: the concrete allowlist is the
// complete stream surface, not an optimistic wildcard. The existing per-stream
// fixture execution test below supplies the corresponding source proof.
func TestSourceTransportDeclaresEveryExecutableZoomStream(t *testing.T) {
	bundle := loadZoomBundle(t)
	if bundle.SyncTransport == nil || bundle.SyncTransport.Source == nil {
		t.Fatal("Zoom must declare its source sync transport")
	}
	if bundle.SyncTransport.Destination != nil {
		t.Fatal("Zoom has no connector-neutral typed destination executor or action binding to declare")
	}
	source := bundle.SyncTransport.Source
	if got, want := source.Executor.Family, connectors.TransportExecutorFamilyDeclarativeAPI; got != want {
		t.Errorf("source transport family = %q, want %q", got, want)
	}
	if got, want := source.Executor.ID, "declarative_stream_source"; got != want {
		t.Errorf("source transport executor = %q, want %q", got, want)
	}
	if got, want := source.Conformance.Suite, "declarative_stream_transport"; got != want {
		t.Errorf("source transport conformance suite = %q, want %q", got, want)
	}
	if got, want := source.Conformance.RunID, "zoom_users_meetings_webinars_v1"; got != want {
		t.Errorf("source transport conformance run = %q, want %q", got, want)
	}
	if got, want := source.EligibleStreams, []string{"users", "meetings", "webinars"}; !reflect.DeepEqual(got, want) {
		t.Errorf("source transport eligible streams = %#v, want %#v", got, want)
	}
	if got, want := source.Delivery.Idempotency, connectors.DeliveryIdempotencyKeyed; got != want {
		t.Errorf("source transport idempotency = %q, want %q", got, want)
	}
	if got, want := source.Delivery.Ordering, connectors.DeliveryOrderingSource; got != want {
		t.Errorf("source transport ordering = %q, want %q", got, want)
	}
	if got, want := source.Delivery.Deletes, connectors.DeliveryDeletesUnavailable; got != want {
		t.Errorf("source transport deletes = %q, want %q", got, want)
	}
}

// TestPinnedCreationCursorsAreProjected proves the cursors used by certification
// are not invented. The SHA-pinned public OpenAPI modules identify users'
// user_created_at/created_at and meetings/webinars' created_at fields as
// date-time creation timestamps; streams.json projects those exact fields into
// the one common, declared cursor field.
func TestPinnedCreationCursorsAreProjected(t *testing.T) {
	type streamDoc struct {
		Streams []struct {
			Name           string            `json:"name"`
			ComputedFields map[string]string `json:"computed_fields"`
			Schema         string            `json:"schema"`
		} `json:"streams"`
	}
	type schemaDoc struct {
		Cursor     string                     `json:"x-cursor-field"`
		Properties map[string]json.RawMessage `json:"properties"`
	}

	raw, err := os.ReadFile("streams.json")
	if err != nil {
		t.Fatalf("read streams: %v", err)
	}
	var streams streamDoc
	if err := json.Unmarshal(raw, &streams); err != nil {
		t.Fatalf("decode streams: %v", err)
	}
	want := map[string]string{
		"users":    "{{ coalesce record.user_created_at record.created_at }}",
		"meetings": "{{ record.created_at }}",
		"webinars": "{{ record.created_at }}",
	}
	for _, stream := range streams.Streams {
		computed, expected := want[stream.Name]
		if !expected {
			continue
		}
		if got := stream.ComputedFields["created_at"]; got != computed {
			t.Errorf("stream %q created_at projection = %q, want %q", stream.Name, got, computed)
		}
		schemaRaw, err := os.ReadFile(stream.Schema)
		if err != nil {
			t.Errorf("read schema %q: %v", stream.Schema, err)
			continue
		}
		var schema schemaDoc
		if err := json.Unmarshal(schemaRaw, &schema); err != nil {
			t.Errorf("decode schema %q: %v", stream.Schema, err)
			continue
		}
		if schema.Cursor != "created_at" {
			t.Errorf("schema %q cursor = %q, want created_at", stream.Schema, schema.Cursor)
		}
		property, ok := schema.Properties["created_at"]
		if !ok || len(property) == 0 {
			t.Errorf("schema %q must retain the projected created_at property", stream.Schema)
		}
	}
}

// TestCertificationCandidatesDescribeOneBoundedReadAndDeferWrites proves the
// definition can generate the one safe live read candidate without pretending
// that reverse-ETL actions have a connector-neutral destination executor.
func TestCertificationCandidatesDescribeOneBoundedReadAndDeferWrites(t *testing.T) {
	bundle := loadZoomBundle(t)
	if bundle.Certification == nil {
		t.Fatal("certification.json did not load")
	}
	if got := bundle.Certification.Source.DefaultStream; got != "users" {
		t.Errorf("certification default stream = %q, want users", got)
	}
	generation := bundle.Certification.DirectReadGeneration
	if generation == nil || len(generation.Cohorts) != 1 {
		t.Fatalf("direct-read candidate generation = %+v, want one bounded cohort", generation)
	}
	cohort := generation.Cohorts[0]
	if cohort.Name != "authenticated_self_read" || cohort.CommandCount != 1 || !reflect.DeepEqual(cohort.Commands, []string{"api users user"}) {
		t.Errorf("direct-read candidate cohort = %+v, want the authenticated self read", cohort)
	}
	if generation.RequiredFlagDefaults["user-id"] != "me" || generation.RequiredFlagDefaults["userId"] != "me" {
		t.Error("self-read candidate must supply both declared user path flag forms as me")
	}
	mutations := bundle.Certification.MutationGeneration
	if mutations == nil || mutations.Cohort.CommandCount != 204 || !reflect.DeepEqual(mutations.Cohort.Intents, []string{"reverse_etl"}) {
		t.Fatalf("mutation candidate generation = %+v, want all 204 reverse-ETL actions", mutations)
	}
	if len(mutations.Families) != 1 || mutations.Families[0].ID != "generic_typed_destination_executor_deferred" || mutations.Families[0].Classification.Code != "unassessed" {
		t.Errorf("mutation candidate containment = %+v, want one explicitly deferred generic-destination family", mutations.Families)
	}
}

// TestAcceptedLiveReadProofDoesNotOverstateCertification locks the distinction
// between an accepted live observation and a certified cell. The matrix has no
// operation-specific fixture projection yet, so fixture_tested must remain
// false even though the bounded REST read has accepted live evidence.
func TestAcceptedLiveReadProofDoesNotOverstateCertification(t *testing.T) {
	raw, err := os.ReadFile("certification-matrix.json")
	if err != nil {
		t.Fatalf("read generated certification matrix: %v", err)
	}
	var matrix struct {
		Connector struct {
			Cells []struct {
				FunctionKind  string `json:"function_kind"`
				FixtureTested bool   `json:"fixture_tested"`
				LiveTested    bool   `json:"live_tested"`
				LiveEvidence  []any  `json:"live_evidence"`
			} `json:"cells"`
		} `json:"connector"`
	}
	if err := json.Unmarshal(raw, &matrix); err != nil {
		t.Fatalf("decode generated certification matrix: %v", err)
	}
	for _, cell := range matrix.Connector.Cells {
		if cell.FunctionKind != "operation:rest_read" {
			continue
		}
		if !cell.LiveTested || len(cell.LiveEvidence) == 0 {
			t.Fatal("operation:rest_read must retain its accepted bounded live proof")
		}
		if cell.FixtureTested {
			t.Fatal("operation:rest_read must remain uncertified until an operation-specific fixture projection exists")
		}
		return
	}
	t.Fatal("generated certification matrix has no operation:rest_read cell")
}

// TestRunnableOperationContractsHaveCommands is the corrected parity gate: an
// operation contract is useful only when a user can reach it through an
// implemented command. It deliberately limits the first promoted write cohort
// to operations with no request body. A content type alone is not a typed root
// payload schema, so an application/json write cannot be made runnable by
// inventing an --input wrapper.
func TestRunnableOperationContractsHaveCommands(t *testing.T) {
	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	if bundle.Surface == nil || connector.CommandSurface() == nil {
		t.Fatal("Zoom has no API or CLI surface")
	}

	type disposition struct {
		Method string `json:"method"`
		Path   string `json:"path"`
		State  string `json:"state"`
		Source struct {
			Prerequisites any `json:"prerequisites"`
		} `json:"source"`
		Rejection *struct {
			Reason string `json:"reason"`
		} `json:"rejection"`
		Declaration struct {
			ID string `json:"id"`
		} `json:"declaration"`
	}
	var report struct {
		LedgerDispositions []disposition `json:"ledger_dispositions"`
	}
	raw, err := os.ReadFile(filepath.Join("sources", "zoom-declaration-disposition.json"))
	if err != nil {
		t.Fatalf("read declaration disposition: %v", err)
	}
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatalf("decode declaration disposition: %v", err)
	}

	paid := make(map[string]bool, len(report.LedgerDispositions))
	runnable := make(map[string]bool, len(report.LedgerDispositions))
	for _, row := range report.LedgerDispositions {
		if row.Rejection != nil && (row.Rejection.Reason == "requires-elevated-scope" || row.Rejection.Reason == "unsafe-to-exercise") {
			t.Errorf("operation %q remains disabled with disallowed reason %q", row.Declaration.ID, row.Rejection.Reason)
		}
		if row.Declaration.ID == "" {
			continue
		}
		prerequisites := strings.ToLower(strings.TrimSpace(stringifyZoomPrerequisites(row.Source.Prerequisites)))
		paid[row.Declaration.ID] = row.Rejection != nil && row.Rejection.Reason == "requires-paid-tier" || zoomPrerequisitesRequirePaidTier(prerequisites)
		runnable[row.Declaration.ID] = row.State == "implemented-pending-certification"
	}

	commands := make(map[string]connectors.CommandSurfaceCommand, len(connector.CommandSurface().Commands))
	for _, command := range connector.CommandSurface().Commands {
		if command.Intent == "direct_read" && command.Operation != "" {
			commands[command.Operation] = command
		}
		if command.Intent == "reverse_etl" && strings.HasPrefix(command.Write, "zoom_") {
			commands[command.Write] = command
		}
	}
	actions := make(map[string]engine.WriteAction, len(bundle.Writes))
	for _, action := range bundle.Writes {
		actions[action.Name] = action
	}

	wantReads, wantWrites, wantDeletes := 0, 0, 0
	for _, operation := range bundle.Operations {
		if paid[operation.ID] {
			if command, ok := commands[operation.ID]; ok && command.Availability == "implemented" {
				t.Errorf("paid-tier operation %q must not have an implemented command %q", operation.ID, command.Path)
			}
			continue
		}
		if !runnable[operation.ID] {
			continue
		}
		switch {
		case operation.Kind == "rest_read" && zoomOperationParametersRunnable(operation):
			wantReads++
			assertRunnableZoomOperation(t, connector, commands, actions, operation, "direct_read")
		case operation.Kind == "rest_write" && operation.REST != nil && operation.REST.ContentType == "" && zoomOperationParametersRunnable(operation):
			wantWrites++
			if operation.MutationClass == "delete" {
				wantDeletes++
			}
			assertRunnableZoomOperation(t, connector, commands, actions, operation, "reverse_etl")
		}
	}

	if wantReads != 505 || wantWrites != 202 || wantDeletes != 185 {
		t.Fatalf("runnable cohort reads=%d writes=%d deletes=%d, want 505/202/185", wantReads, wantWrites, wantDeletes)
	}
	if got := len(connector.CommandSurface().Commands); got != 712 {
		t.Fatalf("Zoom runnable CLI commands = %d, want 712", got)
	}
}

// TestEveryImplementedZoomCommandStopsAtCredentialBoundary proves each generated
// command reaches the real CLI dispatcher and refuses before provider I/O when
// the initialized project has no Zoom credential. Required fixture values only
// exercise local parsing; they are never stored or sent to Zoom.
func TestEveryImplementedZoomCommandStopsAtCredentialBoundary(t *testing.T) {
	bundle := loadZoomBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	surface := connector.CommandSurface()
	if surface == nil {
		t.Fatal("Zoom has no CLI surface")
	}

	root := t.TempDir()
	var initStdout, initStderr bytes.Buffer
	if code := cli.Run([]string{"init", "--root", root, "--json"}, &initStdout, &initStderr); code != 0 {
		t.Fatalf("initialize no-credential project: code=%d stdout=%s stderr=%s", code, initStdout.String(), initStderr.String())
	}

	for _, command := range surface.Commands {
		args := []string{"--root", root, "zoom"}
		args = append(args, strings.Fields(command.Path)...)
		for _, flag := range command.Flags {
			if flag.Required {
				args = append(args, "--"+flag.Name, zoomFixtureFlagValue(flag))
			}
		}
		if command.Intent == "reverse_etl" && strings.HasPrefix(command.Write, "zoom_") {
			args = append(args, "--confirm", "destructive")
		}
		args = append(args, "--json")

		var stdout, stderr bytes.Buffer
		if code := cli.Run(args, &stdout, &stderr); code != 1 || !strings.Contains(stdout.String()+stderr.String(), "missing --credential") {
			t.Errorf("CLI %q = code=%d stdout=%s stderr=%s, want missing --credential before provider I/O", command.Path, code, stdout.String(), stderr.String())
		}
	}
}

func zoomFixtureFlagValue(flag connectors.CommandSurfaceFlag) string {
	if len(flag.Values) > 0 {
		return flag.Values[0]
	}
	switch flag.Type {
	case "boolean":
		return "true"
	case "integer", "number":
		return "1"
	default:
		return "fixture-value"
	}
}

func assertRunnableZoomOperation(t *testing.T, connector connectors.Connector, commands map[string]connectors.CommandSurfaceCommand, actions map[string]engine.WriteAction, operation engine.OperationSpec, wantIntent string) {
	t.Helper()
	commandKey := operation.ID
	if wantIntent == "reverse_etl" {
		commandKey = zoomWriteActionName(operation.ID)
	}
	command, ok := commands[commandKey]
	if !ok {
		t.Errorf("runnable operation %q has no CLI command", operation.ID)
		return
	}
	if command.Intent != wantIntent || command.Availability != "implemented" {
		t.Errorf("operation %q command %q = intent=%q availability=%q, want implemented %s", operation.ID, command.Path, command.Intent, command.Availability, wantIntent)
	}
	if len(command.APISurface) != 1 || operation.REST == nil || command.APISurface[0].Method != operation.REST.Method || command.APISurface[0].Path != operation.REST.Path {
		t.Errorf("operation %q command %q does not bind its exact API surface", operation.ID, command.Path)
	}
	if wantIntent == "reverse_etl" {
		action, ok := actions[command.Write]
		if !ok {
			t.Errorf("operation %q command %q references missing write action %q", operation.ID, command.Path, command.Write)
		} else {
			if action.Kind != operation.MutationClass || action.Method != operation.REST.Method {
				t.Errorf("operation %q action %q = kind=%q method=%q, want kind=%q method=%q", operation.ID, action.Name, action.Kind, action.Method, operation.MutationClass, operation.REST.Method)
			}
			if operation.MutationClass == "delete" && (action.Confirmation == nil || action.Confirmation.Kind != "destructive") {
				t.Errorf("delete operation %q action %q lacks destructive confirmation", operation.ID, action.Name)
			}
		}
	}
	if err := commandrunner.Preflight(connector, strings.Fields(command.Path)); err != nil {
		t.Errorf("Preflight(%q) = %v", command.Path, err)
	}
}

func zoomOperationParametersRunnable(operation engine.OperationSpec) bool {
	if operation.REST == nil {
		return false
	}
	for _, parameter := range operation.REST.Parameters {
		switch parameter.Type {
		case "string", "integer", "number", "boolean":
		default:
			return false
		}
	}
	return true
}

func zoomWriteActionName(operationID string) string {
	name := strings.TrimPrefix(operationID, "zoom.")
	name = strings.NewReplacer(".", "_", "-", "_").Replace(name)
	return "zoom_" + name
}

func stringifyZoomPrerequisites(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			parts = append(parts, stringifyZoomPrerequisites(item))
		}
		return strings.Join(parts, " ")
	default:
		return ""
	}
}

func zoomPrerequisitesRequirePaidTier(prerequisites string) bool {
	for _, marker := range []string{
		"paid", "pro or a higher", "pro or higher", "pro plan", "business", "education", "api plan", "webinar plan", "webinar add-on", "webinar add on", "zoom room license", "zoom rooms license", "licensed user", "licensed or an on-prem license", "video sdk account", "subscription",
	} {
		if strings.Contains(prerequisites, marker) {
			return true
		}
	}
	return false
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
