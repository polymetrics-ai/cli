package main

import (
	"encoding/json"
	"os"
	"testing"
)

type gitlabOperation struct {
	Model            string `json:"model"`
	BlockedByDefault bool   `json:"blocked_by_default"`
	Reason           string `json:"reason"`
}

func TestGitLabAPISurfaceFullOperationParityMetrics(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/gitlab/api_surface.json")
	if err != nil {
		t.Fatalf("read gitlab api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string           `json:"method"`
			CoveredBy map[string]any   `json:"covered_by"`
			Excluded  map[string]any   `json:"excluded"`
			Operation *gitlabOperation `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal gitlab api_surface.json: %v", err)
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
		}
	}

	if len(surface.Endpoints) != 1145 {
		t.Fatalf("endpoints = %d, want 1145 (1,144 official OpenAPI operations plus /users compatibility stream)", len(surface.Endpoints))
	}
	if covered != 1142 {
		t.Fatalf("covered endpoints = %d, want 1142 (all non-deprecated GitLab operations plus /users compatibility stream)", covered)
	}
	if operations != 3 {
		t.Fatalf("operation endpoints = %d, want 3 deprecated operations left blocked", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertGitLabStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"DELETE": 125,
		"GET":    504,
		"HEAD":   3,
		"PATCH":  9,
		"POST":   241,
		"PUT":    263,
	})
	assertGitLabStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"DELETE": 125,
		"GET":    502,
		"HEAD":   3,
		"PATCH":  9,
		"POST":   240,
		"PUT":    263,
	})
	assertGitLabStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{
		"GET":  2,
		"POST": 1,
	})
	assertGitLabStringIntMap(t, "models", models, map[string]int{
		"deprecated": 3,
	})
}

func TestGitLabFullOperationParityCommandAndWriteMetrics(t *testing.T) {
	cliRaw, err := os.ReadFile("../../internal/connectors/defs/gitlab/cli_surface.json")
	if err != nil {
		t.Fatalf("read gitlab cli_surface.json: %v", err)
	}
	var cli struct {
		Commands []struct {
			Availability string `json:"availability"`
			Intent       string `json:"intent"`
			Stream       string `json:"stream"`
			Write        string `json:"write"`
			Operation    string `json:"operation"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(cliRaw, &cli); err != nil {
		t.Fatalf("unmarshal gitlab cli_surface.json: %v", err)
	}

	implemented, streamCommands, writeCommands, operationCommands := 0, 0, 0, 0
	intents := map[string]int{}
	for _, cmd := range cli.Commands {
		if cmd.Availability != "implemented" {
			continue
		}
		implemented++
		intents[cmd.Intent]++
		if cmd.Stream != "" {
			streamCommands++
		}
		if cmd.Write != "" {
			writeCommands++
		}
		if cmd.Operation != "" {
			operationCommands++
		}
	}
	if implemented != 1142 {
		t.Fatalf("implemented GitLab commands = %d, want 1142", implemented)
	}
	if streamCommands != 4 {
		t.Fatalf("stream commands = %d, want 4", streamCommands)
	}
	if writeCommands != 637 {
		t.Fatalf("write commands = %d, want 637 non-deprecated mutating operations", writeCommands)
	}
	if operationCommands != 497 {
		t.Fatalf("operation commands = %d, want 497 operation-backed read/binary/HEAD commands", operationCommands)
	}
	assertGitLabStringIntMap(t, "implemented intents", intents, map[string]int{
		"direct_read": 501,
		"etl":         4,
		"reverse_etl": 637,
	})

	writesRaw, err := os.ReadFile("../../internal/connectors/defs/gitlab/writes.json")
	if err != nil {
		t.Fatalf("read gitlab writes.json: %v", err)
	}
	var writes struct {
		Actions []struct {
			Name   string   `json:"name"`
			Method string   `json:"method"`
			Path   string   `json:"path"`
			Fields []string `json:"path_fields"`
		} `json:"actions"`
	}
	if err := json.Unmarshal(writesRaw, &writes); err != nil {
		t.Fatalf("unmarshal gitlab writes.json: %v", err)
	}
	if len(writes.Actions) != 637 {
		t.Fatalf("write actions = %d, want 637", len(writes.Actions))
	}
	for _, action := range writes.Actions {
		if action.Name == "" || action.Method == "" || action.Path == "" {
			t.Fatalf("write action has empty identity fields: %+v", action)
		}
	}
}

func assertGitLabStringIntMap(t *testing.T, name string, got, want map[string]int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Fatalf("%s = %+v, want %+v", name, got, want)
		}
	}
}
