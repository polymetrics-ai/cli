package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestGitHubAPISurfaceOperationLedgerMetrics(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/github/api_surface.json")
	if err != nil {
		t.Fatalf("read github api_surface.json: %v", err)
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
		t.Fatalf("unmarshal github api_surface.json: %v", err)
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

	// These are a snapshot of derived truth, not a budget. They moved when the
	// bundle stopped enumerating only /repos/{owner}/{repo}/… and recorded the
	// whole documented surface: 1220 REST operations plus the 4 fixed GraphQL
	// rows, which are counted separately and never folded into the REST total.
	// Every structural assertion below is unchanged; only the counts the surface
	// now actually holds were re-derived.
	if len(surface.Endpoints) != 1224 {
		t.Fatalf("endpoints = %d, want 1224", len(surface.Endpoints))
	}
	if covered != 1126 {
		t.Fatalf("covered endpoints = %d, want 1126", covered)
	}
	if operations != 98 {
		t.Fatalf("operation endpoints = %d, want 98 (duplicate/deprecated, plus the operations no runtime component can execute)", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"DELETE":  187,
		"GET":     636,
		"GRAPHQL": 4,
		"PATCH":   70,
		"POST":    193,
		"PUT":     134,
	})
	assertStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"DELETE":  179,
		"GET":     571,
		"GRAPHQL": 4,
		"PATCH":   65,
		"POST":    175,
		"PUT":     132,
	})
	assertStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{
		"DELETE": 8,
		"GET":    65,
		"PATCH":  5,
		"POST":   18,
		"PUT":    2,
	})
	assertStringIntMap(t, "models", models, map[string]int{
		"disallowed": 19,
		"duplicate":  67,
		"deprecated": 1,
		// GETs GitHub documents with no JSON success body: 9 boolean 204 status
		// checks plus /zen and /octocat, which return text/plain and
		// application/octocat-stream. engine.decodeDirectReadBody json-decodes
		// every direct-read body, so these are blocked with that dependency
		// named rather than shipped as commands that always fail.
		"direct_read": 11,
	})
	assertStringIntMap(t, "risks", risks, map[string]int{
		"low": 98,
	})
	assertStringIntMap(t, "statuses", statuses, map[string]int{
		"blocked": 98,
	})
}

type githubOperation struct {
	Model            string `json:"model"`
	Status           string `json:"status"`
	Risk             string `json:"risk"`
	BlockedByDefault bool   `json:"blocked_by_default"`
	Reason           string `json:"reason"`
	SourceURL        string `json:"source_url"`
	Notes            string `json:"notes"`
	DuplicateOf      string `json:"duplicate_of"`
}

func requiresSourceOrNotes(model string) bool {
	switch model {
	case "sensitive_reverse_etl", "admin_reverse_etl", "destructive_action", "disallowed":
		return true
	default:
		return false
	}
}

func assertStringIntMap(t *testing.T, name string, got, want map[string]int) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}
