package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestTrelloAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/trello/api_surface.json")
	if err != nil {
		t.Fatalf("read trello api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string                  `json:"method"`
			Path      string                  `json:"path"`
			CoveredBy map[string]any          `json:"covered_by"`
			Excluded  map[string]any          `json:"excluded"`
			Operation *trelloSurfaceOperation `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal trello api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	totalByMethod := map[string]int{}
	coveredByMethod := map[string]int{}
	operationByMethod := map[string]int{}
	models := map[string]int{}
	covered, excluded, operations := 0, 0, 0
	seen := map[string]bool{}

	for i, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			t.Fatalf("duplicate endpoint %q", key)
		}
		seen[key] = true
		totalByMethod[ep.Method]++
		if len(ep.CoveredBy) > 0 {
			covered++
			coveredByMethod[ep.Method]++
		}
		if len(ep.Excluded) > 0 {
			excluded++
		}
		if ep.Operation != nil {
			operations++
			operationByMethod[ep.Method]++
			models[ep.Operation.Model]++
			if !ep.Operation.BlockedByDefault {
				t.Fatalf("endpoint %d operation is not blocked by default: %+v", i, ep.Operation)
			}
			if ep.Operation.Status != "blocked" || ep.Operation.Reason == "" {
				t.Fatalf("endpoint %d operation has incomplete blocked evidence: %+v", i, ep.Operation)
			}
			if ep.Operation.SourceURL == "" && ep.Operation.Notes == "" {
				t.Fatalf("endpoint %d operation is missing source_url or notes", i)
			}
		}
	}

	if len(surface.Endpoints) != 261 {
		t.Fatalf("endpoints = %d, want 261", len(surface.Endpoints))
	}
	if covered != 219 {
		t.Fatalf("covered endpoints = %d, want 219", covered)
	}
	if operations != 42 {
		t.Fatalf("blocked operation endpoints = %d, want 42", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertTrelloStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"DELETE": 37,
		"GET":    128,
		"POST":   45,
		"PUT":    51,
	})
	assertTrelloStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"DELETE": 33,
		"GET":    98,
		"POST":   43,
		"PUT":    45,
	})
	assertTrelloStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{
		"DELETE": 4,
		"GET":    30,
		"POST":   2,
		"PUT":    6,
	})
	assertTrelloStringIntMap(t, "models", models, map[string]int{
		"admin_reverse_etl": 12,
		"direct_read":       19,
		"disallowed":        1,
		"duplicate":         10,
	})

	for _, key := range []string{
		"GET /members/{id}/boards",
		"GET /boards/{id}/cards",
		"POST /cards",
		"POST /cards/{id}/actions/comments",
		"GET /cards/{id}/attachments",
		"POST /cards/{id}/attachments",
		"GET /search",
		"POST /webhooks/",
		"DELETE /webhooks/{id}",
	} {
		if !seen[key] {
			t.Fatalf("expected Trello official endpoint %q", key)
		}
	}
}

type trelloSurfaceOperation struct {
	Model            string `json:"model"`
	Status           string `json:"status"`
	Risk             string `json:"risk"`
	BlockedByDefault bool   `json:"blocked_by_default"`
	Reason           string `json:"reason"`
	SourceURL        string `json:"source_url"`
	Notes            string `json:"notes"`
}

func assertTrelloStringIntMap(t *testing.T, name string, got, want map[string]int) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}
