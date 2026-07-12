package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestFreshdeskAPISurfaceOperationCoverageMetrics(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/freshdesk/api_surface.json")
	if err != nil {
		t.Fatalf("read freshdesk api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string           `json:"method"`
			Path      string           `json:"path"`
			CoveredBy map[string]any   `json:"covered_by"`
			Excluded  map[string]any   `json:"excluded"`
			Operation *githubOperation `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal freshdesk api_surface.json: %v", err)
	}
	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	totalByMethod := map[string]int{}
	coveredByMethod := map[string]int{}
	covered, operations, excluded := 0, 0, 0
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
			switch {
			case ep.Method == "POST" && ep.Path == "/contacts/imports":
				if ep.Operation.Model != "sensitive_reverse_etl" {
					t.Fatalf("contact import blocker model = %q, want sensitive_reverse_etl", ep.Operation.Model)
				}
			case ep.Method == "GET" && ep.Path == "/custom_objects/schemas/{schema_id}/records?{field_name}{operator}{value}":
				if ep.Operation.Model != "direct_read" {
					t.Fatalf("custom object filter blocker model = %q, want direct_read", ep.Operation.Model)
				}
			default:
				t.Fatalf("endpoint %d unexpected blocked operation %s %s", i, ep.Method, ep.Path)
			}
			if !ep.Operation.BlockedByDefault {
				t.Fatalf("blocked operation is not blocked_by_default: %+v", ep.Operation)
			}
		}
	}

	if len(surface.Endpoints) != 170 {
		t.Fatalf("endpoints = %d, want 170", len(surface.Endpoints))
	}
	if covered != 168 {
		t.Fatalf("covered endpoints = %d, want 168", covered)
	}
	if operations != 2 {
		t.Fatalf("operation endpoints = %d, want 2 safe blockers", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"DELETE": 33,
		"GET":    117,
		"POST":   10,
		"PUT":    10,
	})
	assertStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"DELETE": 33,
		"GET":    116,
		"POST":   9,
		"PUT":    10,
	})
}

func TestFreshdeskCLISurfaceAndWriteActionCounts(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/freshdesk/cli_surface.json")
	if err != nil {
		t.Fatalf("read freshdesk cli_surface.json: %v", err)
	}
	var cli struct {
		Commands []struct {
			Intent       string `json:"intent"`
			Availability string `json:"availability"`
			OutputPolicy string `json:"output_policy"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(raw, &cli); err != nil {
		t.Fatalf("unmarshal freshdesk cli_surface.json: %v", err)
	}
	implementedDirectReads := 0
	for i, cmd := range cli.Commands {
		if cmd.Intent == "direct_read" && cmd.Availability == "implemented" {
			implementedDirectReads++
			if cmd.OutputPolicy != "json" {
				t.Fatalf("direct_read command %d output_policy = %q, want json", i, cmd.OutputPolicy)
			}
		}
	}
	if implementedDirectReads != 109 {
		t.Fatalf("implemented direct_read commands = %d, want 109", implementedDirectReads)
	}

	raw, err = os.ReadFile("../../internal/connectors/defs/freshdesk/writes.json")
	if err != nil {
		t.Fatalf("read freshdesk writes.json: %v", err)
	}
	var writes struct {
		Actions []struct {
			Name    string `json:"name"`
			Confirm string `json:"confirm"`
		} `json:"actions"`
	}
	if err := json.Unmarshal(raw, &writes); err != nil {
		t.Fatalf("unmarshal freshdesk writes.json: %v", err)
	}
	if len(writes.Actions) != 50 {
		t.Fatalf("write actions = %d, want 50", len(writes.Actions))
	}
}
