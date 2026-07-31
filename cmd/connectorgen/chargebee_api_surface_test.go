package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestChargebeeAPISurfaceOperationLedgerMetrics(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/chargebee/api_surface.json")
	if err != nil {
		t.Fatalf("read chargebee api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string              `json:"method"`
			Path      string              `json:"path"`
			CoveredBy map[string]any      `json:"covered_by"`
			Excluded  map[string]any      `json:"excluded"`
			Operation *chargebeeOperation `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal chargebee api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	totalByMethod := map[string]int{}
	coveredByMethod := map[string]int{}
	operationByMethod := map[string]int{}
	coveredTargets := map[string]int{}
	models := map[string]int{}
	risks := map[string]int{}
	lanes := map[string]int{}
	covered, excluded, operations := 0, 0, 0

	for i, ep := range surface.Endpoints {
		totalByMethod[ep.Method]++
		if len(ep.CoveredBy) > 0 {
			covered++
			coveredByMethod[ep.Method]++
			for target := range ep.CoveredBy {
				coveredTargets[target]++
			}
		}
		if len(ep.Excluded) > 0 {
			excluded++
		}
		if ep.Operation != nil {
			operations++
			operationByMethod[ep.Method]++
			models[ep.Operation.Model]++
			risks[ep.Operation.Risk]++
			if !ep.Operation.BlockedByDefault {
				t.Fatalf("endpoint %d operation is not blocked by default: %+v", i, ep.Operation)
			}
			if ep.Operation.Reason == "" {
				t.Fatalf("endpoint %d operation is missing reason: %+v", i, ep.Operation)
			}
			lane := officialLaneFromNotes(ep.Operation.Notes)
			if lane == "" {
				t.Fatalf("endpoint %d (%s %s) operation is missing official_lane notes: %+v", i, ep.Method, ep.Path, ep.Operation)
			}
			lanes[lane]++
		}
	}

	if len(surface.Endpoints) != 655 {
		t.Fatalf("endpoints = %d, want 655", len(surface.Endpoints))
	}
	if covered != 389 {
		t.Fatalf("covered endpoints = %d, want 389", covered)
	}
	if operations != 266 {
		t.Fatalf("operation endpoints = %d, want 266", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertChargebeeStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"GET":     134,
		"POST":    298,
		"WEBHOOK": 223,
	})
	assertChargebeeStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"GET":  125,
		"POST": 264,
	})
	assertChargebeeStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{
		"GET":     9,
		"POST":    34,
		"WEBHOOK": 223,
	})
	assertChargebeeStringIntMap(t, "coveredTargets", coveredTargets, map[string]int{
		"stream": 125,
		"write":  264,
	})
	assertChargebeeStringIntMap(t, "models", models, map[string]int{
		"binary_read": 14,
		"direct_read": 252,
	})
	assertChargebeeStringIntMap(t, "risks", risks, map[string]int{
		"medium": 266,
	})
	assertChargebeeStringIntMap(t, "lanes", lanes, map[string]int{
		"binary_file":              14,
		"cdc_changefeed":           234,
		"direct_read_query_search": 18,
	})
}

type chargebeeOperation struct {
	Model            string `json:"model"`
	Status           string `json:"status"`
	Risk             string `json:"risk"`
	BlockedByDefault bool   `json:"blocked_by_default"`
	Reason           string `json:"reason"`
	SourceURL        string `json:"source_url"`
	Notes            string `json:"notes"`
}

func officialLaneFromNotes(notes string) string {
	const prefix = "official_lane="
	idx := strings.Index(notes, prefix)
	if idx < 0 {
		return ""
	}
	lane := notes[idx+len(prefix):]
	if cut := strings.Index(lane, ";"); cut >= 0 {
		lane = lane[:cut]
	}
	return strings.TrimSpace(lane)
}

func assertChargebeeStringIntMap(t *testing.T, name string, got, want map[string]int) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}
