package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// gorgiasDocumentedOperations is the operation total re-derived on 2026-08-07
// from Gorgias's own stable OpenAPI 3.1.0 document (gorgias-rest-api.json,
// bound to ReadMe docs version 1.5.1), fetched via
// https://dash.readme.com/api/v1/api-registry/1qfhqbgmshn434r: 61 paths, 114
// unique (METHOD, path) operations (46 GET, 23 POST, 27 PUT, 18 DELETE), 1
// deprecated (still counted).
//
// The provider-artifact sweep ledger recorded this artifact's kind as
// html_reference, which is wrong: developers.gorgias.com/reference is a
// ReadMe.com-hosted docs site whose STABLE version (1.5.1) publishes exactly
// one registered OpenAPI definition. Fetched as JSON, its operation count
// reconciles EXACTLY with the ledger's carried-forward 114 -- delta 0.
//
// The document has no top-level webhooks block and Gorgias publishes no
// webhook management endpoints, so there are no duplicate/qualified rows the
// way Notion's split search arms required: every row below is exactly one
// unique method+path action.
const (
	gorgiasDocumentedOperations = 114
	gorgiasDocumentedGET        = 46
	gorgiasDocumentedPOST       = 23
	gorgiasDocumentedPUT        = 27
	gorgiasDocumentedDELETE     = 18
)

func TestGorgiasAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/gorgias/api_surface.json")
	if err != nil {
		t.Fatalf("read gorgias api_surface.json: %v", err)
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
				Risk             string `json:"risk"`
				BlockedByDefault bool   `json:"blocked_by_default"`
				Reason           string `json:"reason"`
				SourceURL        string `json:"source_url"`
				Notes            string `json:"notes"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal gorgias api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Errorf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}
	if got := len(surface.Endpoints); got != gorgiasDocumentedOperations {
		t.Errorf("api_surface declares %d rows, want %d documented operations", got, gorgiasDocumentedOperations)
	}

	byMethod := map[string]int{}
	seen := map[string]bool{}
	var blank []string
	covered, blocked, legacyExcluded := 0, 0, 0

	for _, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			t.Errorf("duplicate endpoint row %q", key)
		}
		seen[key] = true
		byMethod[ep.Method]++

		dispositions := 0
		if len(ep.CoveredBy) > 0 {
			dispositions++
			covered++
		}
		if ep.Operation != nil {
			dispositions++
			blocked++
			if !ep.Operation.BlockedByDefault || ep.Operation.Status != "blocked" {
				t.Errorf("%s: operation row must be blocked_by_default with status blocked", key)
			}
			if strings.TrimSpace(ep.Operation.Reason) == "" {
				t.Errorf("%s: blocked row has no reason", key)
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
	if covered+blocked != gorgiasDocumentedOperations {
		t.Errorf("covered(%d)+blocked(%d) = %d, want %d rows", covered, blocked, covered+blocked, gorgiasDocumentedOperations)
	}

	for method, want := range map[string]int{
		"GET":    gorgiasDocumentedGET,
		"POST":   gorgiasDocumentedPOST,
		"PUT":    gorgiasDocumentedPUT,
		"DELETE": gorgiasDocumentedDELETE,
	} {
		if byMethod[method] != want {
			t.Errorf("%s: %d rows, want %d", method, byMethod[method], want)
		}
	}

	for _, key := range []string{
		"GET /api/tickets",
		"GET /api/customers",
		"GET /api/messages",
		"GET /api/satisfaction-surveys",
		"POST /api/upload",
		"GET /api/{file_type}/download/{domain_hash}/{resource_name}",
	} {
		if !seen[key] {
			t.Errorf("expected documented endpoint %q", key)
		}
	}
}
