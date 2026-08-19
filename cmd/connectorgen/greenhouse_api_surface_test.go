package main

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// greenhouseSurfaceRows is Greenhouse Harvest's documented operation count, re-derived 2026-08-07
// from the provider's own reference at https://developers.greenhouse.io/harvest.html
// (1,636,662 bytes, HTTP 200).
//
// Greenhouse publishes no OpenAPI/Swagger document for Harvest; the reference is a single
// Slate-generated static HTML page. The count is taken from the canonical
// "<h3>HTTP Request</h3><p><code>METHOD https://harvest.greenhouse.io/vN/path</code></p>"
// declaration that follows every endpoint section - one declaration per operation, independent of
// heading or prose wording. That yields 138 declarations with zero duplicates, reconciling exactly
// with the provider-artifact ledger's 138.
//
// Two extraction hazards are pinned below because both silently under-count:
//
//  1. DELETE /v1/tags/candidate/{tag id} carries markup damage in Greenhouse's own docs - a stray
//     unescaped &#39; before the URL and a path placeholder with a literal space. A naive regex
//     drops it, and the shipped bundle had in fact dropped it (129 rows, not 130 v1 rows).
//  2. Three <h2> sections each document TWO versioned operations under one heading (a deprecated
//     v1 and its v2 replacement). Counting headings rather than declarations loses three rows.
const greenhouseSurfaceRows = 138

// greenhouseMethodSplit is the distribution of those 138 operations.
var greenhouseMethodSplit = map[string]int{
	"GET":    69,
	"POST":   28,
	"PUT":    8,
	"PATCH":  19,
	"DELETE": 14,
}

// greenhouseOutOfBaseRows are the eight Harvest v2 operations. This bundle binds exactly one
// HTTPBase (https://harvest.greenhouse.io/v1), so a /v2 path cannot be both correctly constructed
// at request time and present in the runtime operation endpoint ledger, which requires
// operation.rest.path to equal an api_surface row verbatim. They are recorded at their documented
// host-root path so they cannot collide with a base-relative v1 row.
var greenhouseOutOfBaseRows = []string{
	"DELETE /v2/jobs/{job_id}/openings",
	"PATCH /v2/job_posts/{job_post_id}",
	"PATCH /v2/job_posts/{job_post_id}/status",
	"PATCH /v2/scheduled_interviews/{scheduled_interview_id}",
	"PATCH /v2/users/",
	"PATCH /v2/users/disable",
	"PATCH /v2/users/enable",
	"POST /v2/scheduled_interviews",
}

func TestGreenhouseAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/greenhouse/api_surface.json")
	if err != nil {
		t.Fatalf("read greenhouse api_surface.json: %v", err)
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
		t.Fatalf("unmarshal greenhouse api_surface.json: %v", err)
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

		// Webhook EVENTS are excluded from the operation total by the counting policy, so they must
		// not appear as api_surface rows at all. Greenhouse documents its webhook events on a
		// separate page (developers.greenhouse.io/webhooks.html) and exposes no webhook management
		// endpoints on Harvest, so this bundle must carry zero webhook rows of either kind.
		if ep.Method == "WEBHOOK" {
			t.Errorf("%q is a webhook EVENT row; events are excluded from the operation surface", key)
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
			// "Blocked" with no named, checkable dependency is a shrug. Every blocked row must say
			// what specifically it waits on.
			if !strings.Contains(ep.Operation.Notes, "Named dependency:") &&
				!strings.Contains(ep.Operation.Reason, "Named dependency:") {
				t.Errorf("%s: blocked row must carry a 'Named dependency:' marker", key)
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
		t.Errorf("%d legacy excluded row(s) remain, want 0 (deprecated operations still count and "+
			"must carry an operation disposition, not an excluded stub)", legacyExcluded)
	}
	if len(surface.Endpoints) != greenhouseSurfaceRows {
		t.Errorf("endpoints = %d, want %d documented operations", len(surface.Endpoints), greenhouseSurfaceRows)
	}
	if covered+blocked != greenhouseSurfaceRows {
		t.Errorf("covered(%d)+blocked(%d) = %d, want %d", covered, blocked, covered+blocked, greenhouseSurfaceRows)
	}
	if !reflect.DeepEqual(totalByMethod, greenhouseMethodSplit) {
		t.Errorf("totalByMethod = %+v, want %+v", totalByMethod, greenhouseMethodSplit)
	}

	// Hazard 1: the markup-damaged declaration the shipped bundle dropped.
	if !seen["DELETE /tags/candidate/{tag_id}"] {
		t.Errorf("expected DELETE /tags/candidate/{tag_id} (its docs markup carries a stray &#39; " +
			"and a placeholder with a literal space; a naive extraction drops it)")
	}
	// Hazard 2: the v2 operations that share an <h2> with a deprecated v1 sibling.
	for _, want := range greenhouseOutOfBaseRows {
		if !seen[want] {
			t.Errorf("expected out-of-base row %q", want)
		}
	}
}
