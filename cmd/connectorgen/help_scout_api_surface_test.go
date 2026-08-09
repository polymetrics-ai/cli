package main

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// helpScoutSurfaceRows is Help Scout Mailbox's documented operation count, re-derived 2026-08-07
// from developer.helpscout.com.
//
// THREE NUMBERS ARE IN CIRCULATION AND ONLY ONE SURVIVES THE COUNTING POLICY:
//
//	146  the provider-artifact ledger  — the raw count of documentation PAGES
//	145  the sweep derivation          — deduped by (method, LITERAL example path)
//	144  this test                     — deduped by (method, TEMPLATED path), the stated policy
//
// Help Scout publishes no machine-readable spec (checked: no .json/.yaml/openapi/swagger linked
// from the docs, no spec repo in the helpscout GitHub org), so the surface is the 146 endpoint
// pages listed in the shared left-nav of https://developer.helpscout.com/mailbox-api/. Each page
// renders exactly one "METHOD path" request line — 146 pages, 146 request lines, no page with
// zero or two. The pages are not, however, 146 operations:
//
//  1. GET .../threads/{threadId}/original-source is documented on TWO pages, one per Accept
//     header (application/json vs message/rfc822). Content negotiation on one endpoint. The
//     literal request lines are byte-identical, so the sweep derivation already caught this:
//     146 -> 145.
//
//  2. DELETE /v2/customers/{customerId} is documented on TWO pages, "Delete Customer" and
//     "Delete Customer Asynchronously". Their literal request lines DIFFER — /v2/customers/100
//     versus /v2/customers/100?async=true — so deduping on the literal path misses it, which is
//     why the sweep derivation stopped at 145. Both pages declare the SAME templated path in
//     their own "Path Parameters" block: /v2/customers/{customerId}. async is a query parameter
//     that switches the response (202 Accepted instead of 204 No Content), not a second
//     endpoint. Under this sweep's policy — one operation is one unique method+path — that is
//     one operation: 145 -> 144.
//
// (2) is the exact defect class the captain flagged on lever-hiring: double-counting
// query-string variants of one endpoint. The counting is therefore done on the templated path
// Help Scout itself publishes, never on the example URL, and this test forbids a bare "?" in any
// row so the variant cannot creep back in.
const helpScoutSurfaceRows = 144

// helpScoutMethodSplit is the distribution of those 144 operations. The sweep derivation's
// DELETE 19 becomes 18 for the reason above; every other method is unchanged.
var helpScoutMethodSplit = map[string]int{
	"GET":    79,
	"POST":   21,
	"PUT":    20,
	"PATCH":  6,
	"DELETE": 18,
}

// helpScoutOutOfBaseRows are the five Mailbox API v3 operations. This bundle's base_url is
// https://api.helpscout.net/v2, so a /v3 path is outside it: the runtime operation endpoint
// ledger requires operation.rest.path to equal an api_surface row verbatim, and a base-relative
// /v3 form would build .../v2/v3/... They are recorded at their documented host-root path.
var helpScoutOutOfBaseRows = []string{
	"GET /v3/conversations/{conversationId}",
	"GET /v3/conversations/{conversationId}/threads",
	"GET /v3/customers",
	"GET /v3/system-users",
	"GET /v3/system-users/{systemUserId}",
}

func TestHelpScoutAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/help-scout/api_surface.json")
	if err != nil {
		t.Fatalf("read help-scout api_surface.json: %v", err)
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
		t.Fatalf("unmarshal help-scout api_surface.json: %v", err)
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

		// Webhook EVENTS are excluded from the operation surface by the counting policy. Help
		// Scout documents 26 of them at developer.helpscout.com/webhooks/, a section outside the
		// Mailbox API nav. Its five webhook MANAGEMENT endpoints on /v2/webhooks are ordinary
		// operations and DO belong here.
		if ep.Method == "WEBHOOK" {
			t.Errorf("%q is a webhook EVENT row; events are excluded from the operation surface", key)
		}
		// The same operation documented with different example query strings is one operation.
		// This is what separates 144 from the derivation's 145.
		if strings.Contains(ep.Path, "?") {
			t.Errorf("%q carries an example query string; collapse it to the bare operation", key)
		}
		// A wildcard is a family of operations, not an operation. The shipped bundle carried
		// "GET /v2/reports/*" as one row standing for 33 report endpoints.
		if strings.Contains(ep.Path, "*") {
			t.Errorf("%q is a wildcard, not an operation; enumerate the endpoints it stands for", key)
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
		t.Errorf("%d legacy excluded row(s) remain, want 0", legacyExcluded)
	}
	if len(surface.Endpoints) != helpScoutSurfaceRows {
		t.Errorf("endpoints = %d, want %d documented operations", len(surface.Endpoints), helpScoutSurfaceRows)
	}
	if covered+blocked != helpScoutSurfaceRows {
		t.Errorf("covered(%d)+blocked(%d) = %d, want %d", covered, blocked, covered+blocked, helpScoutSurfaceRows)
	}
	if !reflect.DeepEqual(totalByMethod, helpScoutMethodSplit) {
		t.Errorf("totalByMethod = %+v, want %+v", totalByMethod, helpScoutMethodSplit)
	}

	// The async delete must be folded into the one DELETE row, not carried as a second one.
	if !seen["DELETE /v2/customers/{customerId}"] {
		t.Errorf("expected DELETE /v2/customers/{customerId} (the sync and async deletes are one " +
			"operation; async is a query parameter)")
	}
	// Binary: exactly one download in this connector. /file streams the attachment bytes
	// (Content-Disposition + image/png); /data returns base64 inside JSON and is an ordinary
	// read. Getting that pair backwards is the binary-detection trap here.
	if !seen["GET /v2/conversations/{conversationId}/attachments/{attachmentId}/file"] {
		t.Errorf("expected the attachment file download endpoint")
	}
	if !seen["GET /v2/conversations/{conversationId}/attachments/{attachmentId}/data"] {
		t.Errorf("expected the attachment data endpoint (base64 in JSON, NOT a binary download)")
	}
	for _, want := range helpScoutOutOfBaseRows {
		if !seen[want] {
			t.Errorf("expected out-of-base v3 row %q", want)
		}
	}
}
