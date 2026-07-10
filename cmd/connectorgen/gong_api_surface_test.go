package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestGongAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/gong/api_surface.json")
	if err != nil {
		t.Fatalf("read gong api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string                `json:"method"`
			Path      string                `json:"path"`
			CoveredBy map[string]any        `json:"covered_by"`
			Excluded  map[string]any        `json:"excluded"`
			Operation *gongSurfaceOperation `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal gong api_surface.json: %v", err)
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
			if ep.Operation.Reason == "" {
				t.Fatalf("endpoint %d operation is missing reason: %+v", i, ep.Operation)
			}
			if gongRequiresSourceOrNotes(ep.Operation.Model) && ep.Operation.SourceURL == "" && ep.Operation.Notes == "" {
				t.Fatalf("endpoint %d operation %q is missing source_url or notes", i, ep.Operation.Model)
			}
		}
	}

	if len(surface.Endpoints) != 67 {
		t.Fatalf("endpoints = %d, want 67", len(surface.Endpoints))
	}
	if covered != 57 {
		t.Fatalf("covered endpoints = %d, want 57", covered)
	}
	if operations != 10 {
		t.Fatalf("operation endpoints = %d, want 10", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertGongStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"DELETE": 3,
		"GET":    28,
		"PATCH":  1,
		"POST":   27,
		"PUT":    8,
	})
	assertGongStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"DELETE": 3,
		"GET":    28,
		"PATCH":  1,
		"POST":   17,
		"PUT":    8,
	})
	assertGongStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{
		"POST": 10,
	})
	assertGongStringIntMap(t, "models", models, map[string]int{
		"direct_read": 10,
	})

	for _, key := range []string{
		"POST /v2/calls/extensive",
		"POST /v2/calls/transcript",
		"POST /v2/stats/interaction",
	} {
		if !seen[key] {
			t.Fatalf("expected official POST read-query endpoint %q", key)
		}
	}
	for _, key := range []string{
		"GET /v2/calls/extensive",
		"GET /v2/calls/transcript",
		"GET /v2/stats/interaction",
		"GET /v2/stats/activity/trackers",
		"GET /v2/settings/webhooks",
	} {
		if seen[key] {
			t.Fatalf("stale or wrong-method endpoint %q should not be present", key)
		}
	}
}

type gongSurfaceOperation struct {
	Model            string `json:"model"`
	Status           string `json:"status"`
	Risk             string `json:"risk"`
	BlockedByDefault bool   `json:"blocked_by_default"`
	Reason           string `json:"reason"`
	SourceURL        string `json:"source_url"`
	Notes            string `json:"notes"`
}

func gongRequiresSourceOrNotes(model string) bool {
	switch model {
	case "sensitive_reverse_etl", "admin_reverse_etl", "destructive_action", "disallowed":
		return true
	default:
		return false
	}
}

func assertGongStringIntMap(t *testing.T, name string, got, want map[string]int) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}
