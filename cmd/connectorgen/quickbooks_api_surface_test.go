package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestQuickBooksAPISurfaceOperationLedgerMetrics(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/quickbooks/api_surface.json")
	if err != nil {
		t.Fatalf("read quickbooks api_surface.json: %v", err)
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
		t.Fatalf("unmarshal quickbooks api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	covered, excluded, operations := 0, 0, 0
	lanes := map[string]int{}
	coveredStreams := map[string]bool{}
	models := map[string]int{}
	for i, ep := range surface.Endpoints {
		if len(ep.CoveredBy) > 0 {
			covered++
			if stream, ok := ep.CoveredBy["stream"].(string); ok && stream != "" {
				coveredStreams[stream] = true
			}
		}
		if len(ep.Excluded) > 0 {
			excluded++
		}
		if ep.Operation != nil {
			operations++
			models[ep.Operation.Model]++
			if !ep.Operation.BlockedByDefault {
				t.Fatalf("endpoint %d operation is not blocked by default: %+v", i, ep.Operation)
			}
			if ep.Operation.Reason == "" {
				t.Fatalf("endpoint %d operation is missing reason: %+v", i, ep.Operation)
			}
			if requiresSourceOrNotes(ep.Operation.Model) && ep.Operation.SourceURL == "" && ep.Operation.Notes == "" {
				t.Fatalf("endpoint %d operation %q is missing source_url or notes", i, ep.Operation.Model)
			}
		}
		lane := laneFromQuickBooksNotes(ep.Operation)
		if lane == "" && len(ep.CoveredBy) > 0 {
			lane = "etl_read"
		}
		if lane != "" {
			lanes[lane]++
		}
	}

	if len(surface.Endpoints) != 161 {
		t.Fatalf("endpoints = %d, want 161", len(surface.Endpoints))
	}
	if covered != 5 {
		t.Fatalf("covered endpoints = %d, want 5", covered)
	}
	if operations != 156 {
		t.Fatalf("operation endpoints = %d, want 156", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertStringIntMap(t, "quickbooks lanes", lanes, map[string]int{
		"etl_read":                 43,
		"reverse_etl_write":        60,
		"direct_read_query_search": 32,
		"binary_file":              25,
		"cdc_changefeed":           1,
	})
	if !reflect.DeepEqual(coveredStreams, map[string]bool{
		"accounts":  true,
		"customers": true,
		"invoices":  true,
		"payments":  true,
		"vendors":   true,
	}) {
		t.Fatalf("covered streams = %+v", coveredStreams)
	}
	if models["destructive_action"] == 0 {
		t.Fatalf("models = %+v, want destructive_action rows for QuickBooks delete/void operations", models)
	}
}

func laneFromQuickBooksNotes(op *githubOperation) string {
	if op == nil {
		return ""
	}
	for _, part := range strings.Split(op.Notes, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "official_lane=") {
			return strings.TrimPrefix(part, "official_lane=")
		}
	}
	return ""
}
