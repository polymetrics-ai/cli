package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestFreshchatAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/freshchat/api_surface.json")
	if err != nil {
		t.Fatalf("read freshchat api_surface.json: %v", err)
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
		t.Fatalf("unmarshal freshchat api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	streams, writes, directReads, excluded, operations := 0, 0, 0, 0, 0
	models := map[string]int{}
	for i, ep := range surface.Endpoints {
		if _, ok := ep.CoveredBy["stream"]; ok {
			streams++
		}
		if _, ok := ep.CoveredBy["write"]; ok {
			writes++
		}
		if _, ok := ep.CoveredBy["direct_read"]; ok {
			directReads++
		}
		if len(ep.Excluded) > 0 {
			excluded++
		}
		if ep.Operation != nil {
			operations++
			models[ep.Operation.Model]++
			if ep.Operation.Status != "blocked" {
				t.Fatalf("endpoint %d (%s %s) operation status = %q, want blocked", i, ep.Method, ep.Path, ep.Operation.Status)
			}
			if !ep.Operation.BlockedByDefault {
				t.Fatalf("endpoint %d (%s %s) operation is not blocked by default", i, ep.Method, ep.Path)
			}
			if ep.Operation.Reason == "" {
				t.Fatalf("endpoint %d (%s %s) operation is missing reason", i, ep.Method, ep.Path)
			}
			if requiresSourceOrNotes(ep.Operation.Model) && ep.Operation.SourceURL == "" && ep.Operation.Notes == "" {
				t.Fatalf("endpoint %d (%s %s) operation %q is missing source_url or notes", i, ep.Method, ep.Path, ep.Operation.Model)
			}
		}
	}

	if len(surface.Endpoints) != 34 {
		t.Fatalf("endpoints = %d, want 34", len(surface.Endpoints))
	}
	if streams != 18 {
		t.Fatalf("stream endpoints = %d, want 18", streams)
	}
	if writes != 13 {
		t.Fatalf("write endpoints = %d, want 13", writes)
	}
	if directReads != 1 {
		t.Fatalf("direct-read endpoints = %d, want 1", directReads)
	}
	if operations != 2 {
		t.Fatalf("blocked operation endpoints = %d, want 2", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertStringIntMap(t, "models", models, map[string]int{
		"disallowed": 2,
	})
}
