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
	profileUpdateMentionsMultipartPolicy := false

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
			if ep.Method == "PUT" && ep.Path == "/api/v1/profile" && ep.Operation.Model == "sensitive_reverse_etl" && ep.Operation.Notes == "Multipart avatar/profile updates require the later bounded binary/file policy before any file-bearing action is exposed." {
				profileUpdateMentionsMultipartPolicy = true
			}
		}
	}

	if len(surface.Endpoints) != 144 {
		t.Fatalf("endpoints = %d, want 144", len(surface.Endpoints))
	}
	if len(uniquePaths) != 89 {
		t.Fatalf("unique paths = %d, want 89", len(uniquePaths))
	}
	if covered != 16 {
		t.Fatalf("covered endpoints = %d, want 16", covered)
	}
	if streamCovered != 7 {
		t.Fatalf("stream covered endpoints = %d, want 7", streamCovered)
	}
	if writeCovered != 6 {
		t.Fatalf("write covered endpoints = %d, want 6", writeCovered)
	}
	if directReadCovered != 3 {
		t.Fatalf("direct read covered endpoints = %d, want 3", directReadCovered)
	}
	if operations != 128 {
		t.Fatalf("operation endpoints = %d, want 128", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	if !messageWriteCoversJSON {
		t.Fatal("send_message JSON write coverage is missing from POST conversation messages")
	}
	if !profileUpdateMentionsMultipartPolicy {
		t.Fatal("PUT /api/v1/profile must record the blocked multipart avatar/profile policy")
	}
	assertStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"DELETE": 18,
		"GET":    62,
		"PATCH":  21,
		"POST":   41,
		"PUT":    2,
	})
	assertStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"GET":  10,
		"POST": 5,
		"PUT":  1,
	})
	assertStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{
		"DELETE": 18,
		"GET":    52,
		"PATCH":  21,
		"POST":   36,
		"PUT":    1,
	})
	assertStringIntMap(t, "models", models, map[string]int{
		"admin_reverse_etl":     35,
		"destructive_action":    19,
		"direct_read":           50,
		"disallowed":            4,
		"duplicate":             1,
		"sensitive_reverse_etl": 19,
	})
	assertStringIntMap(t, "risks", risks, map[string]int{
		"critical": 5,
		"high":     61,
		"low":      5,
		"medium":   57,
	})
	assertStringIntMap(t, "statuses", statuses, map[string]int{
		"blocked": 128,
	})
}
