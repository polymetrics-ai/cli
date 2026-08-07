package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// notionDocumentedOperations is the operation total re-derived on 2026-08-07
// from Notion's official OpenAPI 3.1.0 document at
// https://developers.notion.com/openapi.json: 49 HTTP operations across 34
// paths (20 GET, 17 POST, 8 PATCH, 4 DELETE), plus the 2 legacy endpoints
// documented only under the reference nav's explicit "Databases (deprecated)"
// group and absent from the current document (GET /v1/databases and
// POST /v1/databases/{database_id}/query).
//
// The provider-artifact sweep carried forward 50 from an older audit and
// classified the artifact as html_reference, which is also what kept Notion out
// of the canonical batch authoring path. Both were stale: a real machine
// readable OpenAPI document exists.
//
// The document's 31 top-level webhooks entries are webhook EVENTS, excluded per
// the sweep counting policy, and Notion exposes no webhook management
// endpoints. Counting documentation PAGES rather than unique method+path
// actions yields roughly 55-57, because four operations are each documented on
// two pages; this test pins the action count so that miscount cannot land.
const (
	notionDocumentedOperations = 51
	notionDocumentedGET        = 21
	notionDocumentedPOST       = 18
	notionDocumentedPATCH      = 8
	notionDocumentedDELETE     = 4

	// Three operations are carried on two rows each, because covered_by names a
	// single target and no bundle leaves a declared stream or write without a
	// row: POST /v1/search (object=page and object=database),
	// PATCH /v1/comments/{comment_id} (its two modelled union arms), and
	// POST /v1/pages/{page_id}/move (likewise). Rows therefore exceed
	// operations by exactly this much.
	notionDuplicateRows = 3
	notionSurfaceRows   = notionDocumentedOperations + notionDuplicateRows
)

func TestNotionAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/notion/api_surface.json")
	if err != nil {
		t.Fatalf("read notion api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string         `json:"method"`
			Path      string         `json:"path"`
			CoveredBy map[string]any `json:"covered_by"`
			Excluded  map[string]any `json:"excluded"`
			Operation *struct {
				Model            string `json:"model"`
				Status           string `json:"status"`
				SourceURL        string `json:"source_url"`
				Notes            string `json:"notes"`
				BlockedByDefault bool   `json:"blocked_by_default"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal notion api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Errorf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}
	if got := len(surface.Endpoints); got != notionSurfaceRows {
		t.Errorf("api_surface declares %d rows, want %d (%d documented operations + %d duplicate rows)",
			got, notionSurfaceRows, notionDocumentedOperations, notionDuplicateRows)
	}

	byMethod := map[string]int{}
	byTarget := map[string]int{}
	seen := map[string]bool{}
	uniqueOperations := map[string]bool{}
	var blank []string
	covered, blocked, legacyExcluded := 0, 0, 0

	for _, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			t.Errorf("duplicate endpoint row %q", key)
		}
		seen[key] = true

		// A qualified row such as "/v1/search (object=page)" is one arm of a
		// single documented operation. Strip the qualifier to count actions.
		base := ep.Path
		if idx := strings.Index(base, " ("); idx >= 0 {
			base = base[:idx]
		}
		uniqueOperations[ep.Method+" "+base] = true
		byMethod[ep.Method]++

		dispositions := 0
		if len(ep.CoveredBy) > 0 {
			dispositions++
			covered++
			for target := range ep.CoveredBy {
				byTarget[target]++
			}
		}
		if ep.Operation != nil {
			dispositions++
			blocked++
			if !ep.Operation.BlockedByDefault || ep.Operation.Status != "blocked" {
				t.Errorf("%s: operation row must be blocked_by_default with status blocked", key)
			}
			if strings.TrimSpace(ep.Operation.SourceURL) == "" {
				t.Errorf("%s: blocked row has no source citation", key)
			}
			if !strings.HasPrefix(ep.Operation.Notes, "named_dependency=") {
				t.Errorf("%s: blocked row must name its dependency", key)
			}
		}
		if len(ep.Excluded) > 0 {
			dispositions++
			legacyExcluded++
		}
		if dispositions == 0 {
			blank = append(blank, key)
		}
		if dispositions > 1 {
			t.Errorf("%s: carries %d dispositions, want exactly 1", key, dispositions)
		}
	}

	if len(blank) > 0 {
		t.Errorf("%d endpoint(s) carry no disposition, want none: %s", len(blank), strings.Join(blank, ", "))
	}
	if legacyExcluded > 0 {
		t.Errorf("%d legacy excluded row(s) remain; operation_ledger_version mode requires operation rows", legacyExcluded)
	}
	if got := len(uniqueOperations); got != notionDocumentedOperations {
		t.Errorf("api_surface covers %d unique method+path actions, want %d documented operations", got, notionDocumentedOperations)
	}
	if covered+blocked != notionSurfaceRows {
		t.Errorf("covered(%d)+blocked(%d) = %d, want %d rows", covered, blocked, covered+blocked, notionSurfaceRows)
	}

	for method, want := range map[string]int{
		"GET":    notionDocumentedGET,
		"POST":   notionDocumentedPOST,
		"PATCH":  notionDocumentedPATCH,
		"DELETE": notionDocumentedDELETE,
	} {
		unique := 0
		for key := range uniqueOperations {
			if strings.HasPrefix(key, method+" ") {
				unique++
			}
		}
		if unique != want {
			t.Errorf("%s: %d unique documented operations, want %d", method, unique, want)
		}
	}

	for target, want := range map[string]int{"stream": 6, "direct_read": 18, "write": 24} {
		if byTarget[target] != want {
			t.Errorf("covered_by %s = %d, want %d", target, byTarget[target], want)
		}
	}
}
