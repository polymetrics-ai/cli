package main

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// gmailSurfaceRows is Gmail's documented operation count, re-derived 2026-08-07 from Google's
// official Discovery document at https://gmail.googleapis.com/$discovery/rest?version=v1
// (kind discovery#restDescription, id gmail:v1, revision 20260803, 217,687 bytes, public/no auth).
//
// Counted by walking resources.*.methods.* RECURSIVELY: Gmail nests resources
// (users.messages.attachments, users.settings.sendAs.smimeInfo), so a flat one-level walk
// undercounts. There are zero top-level methods outside resources.
//
// The provider-artifact ledger's carried-forward 79 reconciles exactly.
const gmailSurfaceRows = 79

// gmailMethodSplit is the httpMethod distribution of those 79 methods.
var gmailMethodSplit = map[string]int{
	"GET":    30,
	"POST":   28,
	"DELETE": 10,
	"PUT":    8,
	"PATCH":  3,
}

// gmailPromotedEndpoints are operations the carried-forward bundle marked `excluded` that this
// sweep's bar requires to be reachable. Each is adjudicated in
// .planning/phases/gmail-parity-sweep-r1/PLAN.md.
//
//   - watch/stop register and cancel a Pub/Sub push SUBSCRIPTION. Webhook *management* stays in
//     scope; only webhook *events* are deferred, so "non_data_endpoint" was the wrong call.
//   - The smimeInfo rows were excluded for needing the https://mail.google.com/ full-mailbox scope,
//     which gmail's own spec.json already declares — the stated dependency does not exist.
//   - The single-resource detail GETs were excluded as duplicates of a list stream, which conflates
//     ETL coverage with command reachability. Direct read is its own required scope.
var gmailPromotedEndpoints = []string{
	"POST /users/{userId}/watch",
	"POST /users/{userId}/stop",
	"GET /users/{userId}/settings/sendAs/{sendAsEmail}/smimeInfo",
	"GET /users/{userId}/messages/{id}",
	"GET /users/{userId}/threads/{id}",
	"GET /users/{userId}/settings/vacation",
}

func TestGmailAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/gmail/api_surface.json")
	if err != nil {
		t.Fatalf("read gmail api_surface.json: %v", err)
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
				Dependency       string `json:"dependency"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal gmail api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Errorf("operation_ledger_version = %d, want 1 (the v2 provenance ledger is required)", surface.OperationLedgerVersion)
	}

	totalByMethod := map[string]int{}
	seen := map[string]bool{}
	var blank []string
	covered, blocked, legacyExcluded := 0, 0, 0

	for _, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			t.Fatalf("duplicate endpoint %q", key)
		}
		seen[key] = true
		totalByMethod[ep.Method]++

		dispositions := 0
		if len(ep.CoveredBy) > 0 {
			dispositions++
			covered++
		}
		if ep.Operation != nil {
			dispositions++
			blocked++
			// A blocked row must name what it is waiting on. A blank or generic block is exactly
			// the defect this programme exists to remove.
			if strings.TrimSpace(ep.Operation.Reason) == "" {
				t.Errorf("%s: blocked row has no reason", key)
			}
			if strings.TrimSpace(ep.Operation.SourceURL) == "" && strings.TrimSpace(ep.Operation.Notes) == "" {
				t.Errorf("%s: blocked row has no source citation", key)
			}
			// A dependency is "named" when it identifies the specific thing being waited on:
			// an issue reference, an explicit "Named dependency:" marker, or a concrete engine
			// symbol. Prose alone ("not supported yet") is not enough.
			named := strings.TrimSpace(ep.Operation.Dependency) != "" ||
				strings.Contains(ep.Operation.Reason, "#") ||
				strings.Contains(ep.Operation.Notes, "#") ||
				strings.Contains(ep.Operation.Notes, "Named dependency:")
			if !named {
				t.Errorf("%s: blocked row must name its dependency (issue ref or explicit dependency field)", key)
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
		t.Errorf("%d endpoint(s) carry no disposition, want none: %s", len(blank), strings.Join(blank, ", "))
	}
	// The legacy `excluded` category is not one of this sweep's three accepted dispositions
	// (executable, blocked-with-named-dependency, unsupported-with-source-citation).
	if legacyExcluded != 0 {
		t.Errorf("%d legacy excluded row(s) remain, want 0; re-disposition them as covered or blocked-with-named-dependency", legacyExcluded)
	}

	if len(surface.Endpoints) != gmailSurfaceRows {
		t.Errorf("endpoints = %d, want %d documented operations", len(surface.Endpoints), gmailSurfaceRows)
	}
	if covered+blocked != gmailSurfaceRows {
		t.Errorf("covered(%d)+blocked(%d) = %d, want %d", covered, blocked, covered+blocked, gmailSurfaceRows)
	}
	if !reflect.DeepEqual(totalByMethod, gmailMethodSplit) {
		t.Errorf("totalByMethod = %+v, want %+v", totalByMethod, gmailMethodSplit)
	}

	for _, key := range gmailPromotedEndpoints {
		if !seen[key] {
			t.Errorf("expected documented endpoint %q to be present", key)
		}
	}
}
