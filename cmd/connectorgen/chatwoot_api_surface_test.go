package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestChatwootAPISurfaceOperationLedgerMetrics(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/chatwoot/api_surface.json")
	if err != nil {
		t.Fatalf("read chatwoot api_surface.json: %v", err)
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
		t.Fatalf("unmarshal chatwoot api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	totalByMethod := map[string]int{}
	coveredByMethod := map[string]int{}
	operationByMethod := map[string]int{}
	models := map[string]int{}
	risks := map[string]int{}
	statuses := map[string]int{}
	uniquePaths := map[string]struct{}{}
	covered, excluded, operations := 0, 0, 0
	streamCovered, writeCovered, directReadCovered := 0, 0, 0
	messageWriteCoversJSON := false
	profileUpdateCovered := false

	for i, ep := range surface.Endpoints {
		totalByMethod[ep.Method]++
		uniquePaths[ep.Path] = struct{}{}
		hasCovered := len(ep.CoveredBy) > 0
		hasOperation := ep.Operation != nil
		if hasCovered == hasOperation {
			t.Fatalf("endpoint %d %s %s classification covered=%t operation=%t, want exactly one", i, ep.Method, ep.Path, hasCovered, hasOperation)
		}
		if len(ep.Excluded) > 0 {
			excluded++
		}
		if hasCovered {
			covered++
			coveredByMethod[ep.Method]++
			if _, ok := ep.CoveredBy["stream"]; ok {
				streamCovered++
			}
			if write, ok := ep.CoveredBy["write"].(string); ok {
				writeCovered++
				if ep.Method == "POST" && ep.Path == "/api/v1/accounts/{account_id}/conversations/{conversation_id}/messages" && write == "send_message" {
					messageWriteCoversJSON = true
				}
				if ep.Method == "PUT" && ep.Path == "/api/v1/profile" && write == "update_profile" {
					profileUpdateCovered = true
				}
			}
			if _, ok := ep.CoveredBy["direct_read"]; ok {
				directReadCovered++
			}
		}
		if ep.Operation != nil {
			operations++
			operationByMethod[ep.Method]++
			models[ep.Operation.Model]++
			risks[ep.Operation.Risk]++
			statuses[ep.Operation.Status]++
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

	if len(surface.Endpoints) != 144 {
		t.Fatalf("endpoints = %d, want 144", len(surface.Endpoints))
	}
	if len(uniquePaths) != 89 {
		t.Fatalf("unique paths = %d, want 89", len(uniquePaths))
	}
	if covered != 139 {
		t.Fatalf("covered endpoints = %d, want 139", covered)
	}
	if streamCovered != 7 {
		t.Fatalf("stream covered endpoints = %d, want 7", streamCovered)
	}
	if writeCovered != 79 {
		t.Fatalf("write covered endpoints = %d, want 79", writeCovered)
	}
	if directReadCovered != 53 {
		t.Fatalf("direct read covered endpoints = %d, want 53", directReadCovered)
	}
	if operations != 5 {
		t.Fatalf("operation endpoints = %d, want 5", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	if !messageWriteCoversJSON {
		t.Fatal("send_message JSON write coverage is missing from POST conversation messages")
	}
	if !profileUpdateCovered {
		t.Fatal("PUT /api/v1/profile must be covered by the typed update_profile write action")
	}
	assertStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"DELETE": 18,
		"GET":    62,
		"PATCH":  21,
		"POST":   41,
		"PUT":    2,
	})
	assertStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"DELETE": 18,
		"GET":    60,
		"PATCH":  21,
		"POST":   38,
		"PUT":    2,
	})
	assertStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{
		"GET":  2,
		"POST": 3,
	})
	assertStringIntMap(t, "models", models, map[string]int{
		"disallowed": 4,
		"duplicate":  1,
	})
	assertStringIntMap(t, "risks", risks, map[string]int{
		"low": 5,
	})
	assertStringIntMap(t, "statuses", statuses, map[string]int{
		"blocked": 5,
	})
}
