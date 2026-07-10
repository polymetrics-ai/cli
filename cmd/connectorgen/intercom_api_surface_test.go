package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestIntercomAPISurfaceFullCoverage(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/intercom/api_surface.json")
	if err != nil {
		t.Fatalf("read intercom api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string             `json:"method"`
			CoveredBy map[string]string  `json:"covered_by"`
			Excluded  map[string]any     `json:"excluded"`
			Operation *intercomOperation `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal intercom api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	totalByMethod := map[string]int{}
	coveredByMethod := map[string]int{}
	coverageModels := map[string]int{}
	operationByMethod := map[string]int{}
	models := map[string]int{}
	coveredStreams := map[string]int{}
	covered, excluded, operations := 0, 0, 0

	for i, ep := range surface.Endpoints {
		totalByMethod[ep.Method]++
		if len(ep.CoveredBy) > 0 {
			covered++
			coveredByMethod[ep.Method]++
			for model, value := range ep.CoveredBy {
				if value == "" {
					continue
				}
				coverageModels[model]++
			}
			if stream := ep.CoveredBy["stream"]; stream != "" {
				coveredStreams[stream]++
			}
		} else {
			t.Fatalf("endpoint %d %s is not covered by stream, direct read, write, binary policy, or explicit block", i, ep.Method)
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
			if requiresSourceOrNotes(ep.Operation.Model) && ep.Operation.SourceURL == "" && ep.Operation.Notes == "" {
				t.Fatalf("endpoint %d operation %q is missing source_url or notes", i, ep.Operation.Model)
			}
			if ep.Operation.Model == "duplicate" && ep.Operation.DuplicateOf == "" {
				t.Fatalf("endpoint %d duplicate operation is missing duplicate_of", i)
			}
		}
	}

	if len(surface.Endpoints) != 149 {
		t.Fatalf("endpoints = %d, want 149", len(surface.Endpoints))
	}
	if covered != 149 {
		t.Fatalf("covered endpoints = %d, want 149", covered)
	}
	if operations != 0 {
		t.Fatalf("operation endpoints = %d, want 0 blocked metadata-only rows; models=%+v byMethod=%+v", operations, models, operationByMethod)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"DELETE": 19,
		"GET":    67,
		"POST":   47,
		"PUT":    16,
	})
	assertStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"DELETE": 19,
		"GET":    67,
		"POST":   47,
		"PUT":    16,
	})
	for _, model := range []string{"stream", "direct_read", "direct_reads", "write"} {
		if coverageModels[model] == 0 {
			t.Fatalf("coverage model %q was not represented; coverage=%+v", model, coverageModels)
		}
	}
	for _, stream := range []string{"admins", "companies", "contacts", "conversations", "tags"} {
		if coveredStreams[stream] == 0 {
			t.Fatalf("required stream %q was not represented; streams=%+v", stream, coveredStreams)
		}
	}
}

type intercomOperation struct {
	Model            string `json:"model"`
	Status           string `json:"status"`
	Risk             string `json:"risk"`
	BlockedByDefault bool   `json:"blocked_by_default"`
	Reason           string `json:"reason"`
	SourceURL        string `json:"source_url"`
	Notes            string `json:"notes"`
	DuplicateOf      string `json:"duplicate_of"`
}
