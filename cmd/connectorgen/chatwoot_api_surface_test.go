package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

// chatwootDocumentedOperations is the operation total re-derived on 2026-08-07
// from Chatwoot's own bundled OpenAPI 3.1.0 document at
// https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/swagger.json
// (488785 bytes, info.version 1.1.0): 148 operations across 92 paths (GET 64,
// POST 42, PATCH 22, DELETE 18, PUT 2).
//
// The provider-artifact ledger recorded 146 and cited the tag_groups directory
// listing as its source, which is not itself a fetchable artifact; the repo's
// own build bundles those tag_groups/*.yml fragments into the single
// swagger.json used here. Both the ledger's total and its artifact URL were
// stale by this measure: re-derivation adds 2.
//
// Two of the 148 are distinguished ONLY by a trailing slash on an otherwise
// identical path: GET /api/v2/accounts/{account_id}/reports/conversations
// (get-account-conversation-metrics) and GET /api/v2/accounts/{account_id}/
// reports/conversations/ (get-agent-conversation-metrics). This test asserts
// both survive as distinct rows so that class of collapse (cf. DynamoDB's
// X-Amz-Target normalization defect) cannot land silently here.
const (
	chatwootDocumentedOperations = 148
	chatwootDocumentedGET        = 64
	chatwootDocumentedPOST       = 42
	chatwootDocumentedPATCH      = 22
	chatwootDocumentedDELETE     = 18
	chatwootDocumentedPUT        = 2
)

func TestChatwootAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/chatwoot/api_surface.json")
	if err != nil {
		t.Fatalf("read chatwoot api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string                    `json:"method"`
			Path      string                    `json:"path"`
			CoveredBy map[string]any            `json:"covered_by"`
			Excluded  map[string]any            `json:"excluded"`
			Operation *chatwootSurfaceOperation `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal chatwoot api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}
	if len(surface.Endpoints) != chatwootDocumentedOperations {
		t.Fatalf("endpoints = %d, want %d", len(surface.Endpoints), chatwootDocumentedOperations)
	}

	totalByMethod := map[string]int{}
	coveredByMethod := map[string]int{}
	operationByMethod := map[string]int{}
	byTarget := map[string]int{}
	models := map[string]int{}
	seen := map[string]bool{}
	var blank []string
	covered, blocked, legacyExcluded := 0, 0, 0

	for _, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			t.Errorf("duplicate endpoint row %q", key)
		}
		seen[key] = true
		totalByMethod[ep.Method]++

		dispositions := 0
		if len(ep.CoveredBy) > 0 {
			dispositions++
			covered++
			coveredByMethod[ep.Method]++
			for target := range ep.CoveredBy {
				byTarget[target]++
			}
		}
		if ep.Operation != nil {
			dispositions++
			blocked++
			operationByMethod[ep.Method]++
			models[ep.Operation.Model]++
			if !ep.Operation.BlockedByDefault || ep.Operation.Status != "blocked" {
				t.Errorf("%s: operation row must be blocked_by_default with status blocked, got %+v", key, ep.Operation)
			}
			if strings.TrimSpace(ep.Operation.Reason) == "" {
				t.Errorf("%s: blocked row has no reason", key)
			}
			if strings.TrimSpace(ep.Operation.SourceURL) == "" {
				t.Errorf("%s: blocked row has no source citation", key)
			}
			if !strings.HasPrefix(ep.Operation.Notes, "named_dependency=") {
				t.Errorf("%s: blocked row must name its dependency (notes must start \"named_dependency=\")", key)
			}
			if ep.Operation.Model == "duplicate" && strings.TrimSpace(ep.Operation.DuplicateOf) == "" {
				t.Errorf("%s: model=duplicate row requires duplicate_of", key)
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
	if legacyExcluded != 0 {
		t.Errorf("%d legacy excluded row(s) remain, want 0: operation_ledger_version mode requires operation rows", legacyExcluded)
	}
	if covered != 101 {
		t.Errorf("covered endpoints = %d, want 101", covered)
	}
	if blocked != 47 {
		t.Errorf("blocked endpoints = %d, want 47", blocked)
	}
	if covered+blocked != chatwootDocumentedOperations {
		t.Errorf("covered(%d)+blocked(%d) = %d, want %d", covered, blocked, covered+blocked, chatwootDocumentedOperations)
	}

	assertChatwootStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"GET":    chatwootDocumentedGET,
		"POST":   chatwootDocumentedPOST,
		"PATCH":  chatwootDocumentedPATCH,
		"DELETE": chatwootDocumentedDELETE,
		"PUT":    chatwootDocumentedPUT,
	})
	assertChatwootStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"GET":    39,
		"POST":   31,
		"PATCH":  16,
		"DELETE": 14,
		"PUT":    1,
	})
	assertChatwootStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{
		"GET":    25,
		"POST":   11,
		"PATCH":  6,
		"DELETE": 4,
		"PUT":    1,
	})
	assertChatwootStringIntMap(t, "byTarget", byTarget, map[string]int{
		"stream":      7,
		"write":       60,
		"direct_read": 34,
	})
	assertChatwootStringIntMap(t, "models", models, map[string]int{
		"direct_read":           23,
		"admin_reverse_etl":     13,
		"sensitive_reverse_etl": 9,
		"disallowed":            1,
		"duplicate":             1,
	})

	// Hazard: two documented operations share a path and are distinguished ONLY
	// by a trailing slash. If api_surface.json keying normalises or strips
	// trailing slashes, these collapse into one row and the count silently
	// drops to 147. Both must survive as distinct keys.
	const (
		accountConversationMetrics = "GET /api/v2/accounts/{account_id}/reports/conversations"
		agentConversationMetrics   = "GET /api/v2/accounts/{account_id}/reports/conversations/"
	)
	if !seen[accountConversationMetrics] {
		t.Errorf("expected trailing-slash pair member %q (get-account-conversation-metrics) to be present", accountConversationMetrics)
	}
	if !seen[agentConversationMetrics] {
		t.Errorf("expected trailing-slash pair member %q (get-agent-conversation-metrics) to be present", agentConversationMetrics)
	}
	if accountConversationMetrics == agentConversationMetrics {
		t.Fatalf("test bug: trailing-slash pair constants are identical strings")
	}

	// A representative sample of newly-covered rows that must be present and
	// executable, spanning stream/write/direct_read targets.
	for _, key := range []string{
		"GET /api/v1/accounts/{account_id}/conversations",
		"GET /api/v1/accounts/{account_id}/contacts/{id}",
		"DELETE /api/v1/accounts/{account_id}/contacts/{id}",
		"GET /api/v1/accounts/{account_id}/agent_bots",
		"POST /api/v1/accounts/{account_id}/agent_bots",
	} {
		if !seen[key] {
			t.Errorf("expected documented endpoint %q to be present", key)
		}
	}

	// A representative sample of rows that must stay blocked, not executable,
	// because they live outside this bundle's single account-scoped base URL
	// or are permanently disallowed.
	for _, key := range []string{
		"GET /platform/api/v1/users/{id}/login",
		"POST /platform/api/v1/accounts",
		"GET /api/v2/accounts/{account_id}/reports/summary",
		"GET /api/v1/profile",
		"GET /survey/responses/{conversation_uuid}",
		"GET /public/api/v1/inboxes/{inbox_identifier}",
	} {
		if !seen[key] {
			t.Errorf("expected documented (but blocked) endpoint %q to be present", key)
		}
	}
}

type chatwootSurfaceOperation struct {
	Model            string `json:"model"`
	Status           string `json:"status"`
	Risk             string `json:"risk"`
	BlockedByDefault bool   `json:"blocked_by_default"`
	Reason           string `json:"reason"`
	SourceURL        string `json:"source_url"`
	Notes            string `json:"notes"`
	DuplicateOf      string `json:"duplicate_of"`
}

func assertChatwootStringIntMap(t *testing.T, name string, got, want map[string]int) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s = %+v, want %+v", name, got, want)
	}
}
