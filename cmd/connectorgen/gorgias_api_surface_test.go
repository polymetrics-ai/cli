package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGorgiasAPISurfaceOperationLedgerMetrics(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/gorgias/api_surface.json")
	if err != nil {
		t.Fatalf("read gorgias api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string           `json:"method"`
			CoveredBy map[string]any   `json:"covered_by"`
			Excluded  map[string]any   `json:"excluded"`
			Operation *githubOperation `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal gorgias api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	totalByMethod := map[string]int{}
	coveredByMethod := map[string]int{}
	operationByMethod := map[string]int{}
	models := map[string]int{}
	covered, excluded, operations := 0, 0, 0

	for i, ep := range surface.Endpoints {
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
			if ep.Operation.SourceURL == "" && ep.Operation.Notes == "" {
				t.Fatalf("endpoint %d operation %q is missing source_url or notes", i, ep.Operation.Model)
			}
		}
	}

	if len(surface.Endpoints) != 114 {
		t.Fatalf("endpoints = %d, want 114", len(surface.Endpoints))
	}
	if covered != 4 {
		t.Fatalf("covered endpoints = %d, want 4", covered)
	}
	if operations != 110 {
		t.Fatalf("operation endpoints = %d, want 110", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"DELETE": 18,
		"GET":    46,
		"POST":   23,
		"PUT":    27,
	})
	assertStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"GET": 4,
	})
	assertStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{
		"DELETE": 18,
		"GET":    42,
		"POST":   23,
		"PUT":    27,
	})
	assertStringIntMap(t, "models", models, map[string]int{
		"admin_reverse_etl":     27,
		"binary_read":           5,
		"destructive_action":    20,
		"direct_read":           42,
		"disallowed":            1,
		"sensitive_reverse_etl": 15,
	})
}
