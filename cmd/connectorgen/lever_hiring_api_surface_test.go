package main

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// leverHiringSurfaceRows is Lever Hiring's documented operation count, re-derived 2026-08-07 from
// the provider's own reference at https://hire.lever.co/developer/documentation.
//
// THE PROVIDER-ARTIFACT LEDGER'S 67 IS WRONG, and not merely stale: no point in time supports it.
// A 2020 Wayback capture of the same page parsed with identical methodology yields ~129, and the
// current page yields 106. Three independent derivations converged on 106:
//
//  1. the sweep's HTML derivation                                        -> 106
//  2. this lane's own normalised extraction of the same page (108 raw,
//     minus 2 prose "e.g." examples that are not endpoint sections:
//     POST /postings/{postingId}/apply and GET /profile_forms)           -> 106
//  3. the preserved lever-hiring lane's bundle, once its 2 query-string
//     duplicates of GET /postings are collapsed and the one endpoint it
//     missed (GET /v1/eeo/responses, whose docs omit a leading slash)
//     is restored                                                        -> 106
//
// Counting is normalised on method+path with `:param` unified to `{param}` and example query
// strings dropped, because Lever's docs write the same operation several ways.
const leverHiringSurfaceRows = 106

// leverHiringMethodSplit is the distribution of those 106 operations.
var leverHiringMethodSplit = map[string]int{
	"GET":    55,
	"POST":   26,
	"PUT":    14,
	"DELETE": 11,
}

func TestLeverHiringAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/lever-hiring/api_surface.json")
	if err != nil {
		t.Fatalf("read lever-hiring api_surface.json: %v", err)
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
				BlockedByDefault bool   `json:"blocked_by_default"`
				Reason           string `json:"reason"`
				SourceURL        string `json:"source_url"`
				Notes            string `json:"notes"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal lever-hiring api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Errorf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	totalByMethod := map[string]int{}
	seen := map[string]bool{}
	var blank []string
	covered, blocked, legacyExcluded := 0, 0, 0

	for _, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			t.Errorf("duplicate endpoint %q", key)
		}
		seen[key] = true
		totalByMethod[ep.Method]++

		// Webhook EVENTS are excluded from the operation total by the counting policy, so they
		// must not appear as api_surface rows at all. Webhook MANAGEMENT endpoints are ordinary
		// operations and DO belong here.
		if ep.Method == "WEBHOOK" {
			t.Errorf("%q is a webhook EVENT row; events are excluded from the operation surface", key)
		}
		// The same operation documented with different example query strings is one operation.
		if strings.Contains(ep.Path, "?") {
			t.Errorf("%q carries an example query string; collapse it to the bare operation", key)
		}

		dispositions := 0
		if len(ep.CoveredBy) > 0 {
			dispositions++
			covered++
		}
		if ep.Operation != nil {
			dispositions++
			blocked++
			if strings.TrimSpace(ep.Operation.Reason) == "" {
				t.Errorf("%s: blocked row has no reason", key)
			}
			if strings.TrimSpace(ep.Operation.SourceURL) == "" && strings.TrimSpace(ep.Operation.Notes) == "" {
				t.Errorf("%s: blocked row has no source citation", key)
			}
			named := strings.Contains(ep.Operation.Reason, "#") ||
				strings.Contains(ep.Operation.Notes, "#") ||
				strings.Contains(ep.Operation.Reason, "Named dependency:") ||
				strings.Contains(ep.Operation.Notes, "Named dependency:")
			if !named {
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

	sort.Strings(blank)
	if len(blank) > 0 {
		t.Errorf("%d endpoint(s) carry no disposition: %s", len(blank), strings.Join(blank, ", "))
	}
	if legacyExcluded != 0 {
		t.Errorf("%d legacy excluded row(s) remain, want 0", legacyExcluded)
	}
	if len(surface.Endpoints) != leverHiringSurfaceRows {
		t.Errorf("endpoints = %d, want %d documented operations", len(surface.Endpoints), leverHiringSurfaceRows)
	}
	if covered+blocked != leverHiringSurfaceRows {
		t.Errorf("covered(%d)+blocked(%d) = %d, want %d", covered, blocked, covered+blocked, leverHiringSurfaceRows)
	}
	if !reflect.DeepEqual(totalByMethod, leverHiringMethodSplit) {
		t.Errorf("totalByMethod = %+v, want %+v", totalByMethod, leverHiringMethodSplit)
	}

	// GET /v1/eeo/responses is the endpoint the preserved lane dropped because Lever's docs omit
	// its leading slash. Pin it so that normalisation bug cannot silently return.
	if !seen["GET /v1/eeo/responses"] {
		t.Errorf("expected GET /v1/eeo/responses (docs omit its leading slash; easily lost)")
	}
}
