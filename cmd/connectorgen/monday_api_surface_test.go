package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMondayAPISurfaceOperationLedger(t *testing.T) {
	root := filepath.Join("..", "..", "internal", "connectors", "defs", "monday")

	var metadata struct {
		Capabilities struct {
			Write bool `json:"write"`
		} `json:"capabilities"`
	}
	readJSONFile(t, filepath.Join(root, "metadata.json"), &metadata)
	if !metadata.Capabilities.Write {
		t.Fatalf("metadata capabilities.write = false, want true for typed Monday reverse ETL actions")
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string `json:"method"`
			Path      string `json:"path"`
			CoveredBy *struct {
				Stream     string `json:"stream"`
				Write      string `json:"write"`
				DirectRead string `json:"direct_read"`
			} `json:"covered_by"`
			Excluded  any `json:"excluded"`
			Operation *struct {
				Model            string `json:"model"`
				Status           string `json:"status"`
				BlockedByDefault bool   `json:"blocked_by_default"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	readJSONFile(t, filepath.Join(root, "api_surface.json"), &surface)
	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}
	if len(surface.Endpoints) != 254 {
		t.Fatalf("endpoints = %d, want 254 official docs GraphQL operations", len(surface.Endpoints))
	}
	seen := map[string]bool{}
	coveredStreams := map[string]bool{}
	coveredWrites := map[string]bool{}
	coveredDirect := 0
	blocked := 0
	for _, ep := range surface.Endpoints {
		if ep.Excluded != nil {
			t.Fatalf("%s %s uses legacy excluded in operation ledger mode", ep.Method, ep.Path)
		}
		key := strings.ToUpper(ep.Method) + " " + ep.Path
		if seen[key] {
			t.Fatalf("duplicate api_surface endpoint %s", key)
		}
		seen[key] = true
		if ep.CoveredBy != nil {
			if ep.CoveredBy.Stream != "" {
				coveredStreams[ep.CoveredBy.Stream] = true
			}
			if ep.CoveredBy.Write != "" {
				coveredWrites[ep.CoveredBy.Write] = true
			}
			if ep.CoveredBy.DirectRead != "" {
				coveredDirect++
			}
		}
		if ep.Operation != nil {
			blocked++
			if ep.Operation.Status != "blocked" || !ep.Operation.BlockedByDefault {
				t.Fatalf("operation row %s must remain blocked_by_default", key)
			}
		}
	}
	for _, stream := range []string{"boards", "items", "users", "teams", "tags"} {
		if !coveredStreams[stream] {
			t.Fatalf("stream %q is not covered in api_surface.json", stream)
		}
	}
	if coveredDirect == 0 {
		t.Fatal("no typed direct/provider query operations are covered")
	}
	if blocked == 0 {
		t.Fatal("expected provider/source/complex-shape operation rows to remain explicitly blocked")
	}

	var operations struct {
		Operations []struct {
			ID            string `json:"id"`
			Kind          string `json:"kind"`
			MutationClass string `json:"mutation_class"`
			Destructive   bool   `json:"destructive"`
		} `json:"operations"`
	}
	readJSONFile(t, filepath.Join(root, "operations.json"), &operations)
	if len(operations.Operations) != 254 {
		t.Fatalf("operations = %d, want 254", len(operations.Operations))
	}
	counts := map[string]int{}
	for _, op := range operations.Operations {
		counts[op.Kind]++
	}
	if counts["rest_read"] != 66 {
		t.Fatalf("rest_read operations = %d, want 66", counts["rest_read"])
	}
	if counts["graphql_mutation"] != 188 {
		t.Fatalf("graphql_mutation operations = %d, want 188", counts["graphql_mutation"])
	}

	var writes struct {
		Actions []struct {
			Name    string `json:"name"`
			Confirm string `json:"confirm"`
			GraphQL *struct {
				Document string `json:"document"`
			} `json:"graphql"`
		} `json:"actions"`
	}
	readJSONFile(t, filepath.Join(root, "writes.json"), &writes)
	if len(writes.Actions) == 0 {
		t.Fatal("writes.json has zero actions")
	}
	var sawDeleteBoard bool
	for _, action := range writes.Actions {
		if !coveredWrites[action.Name] {
			t.Fatalf("write action %q has no api_surface covered_by.write row", action.Name)
		}
		if action.GraphQL == nil || !strings.Contains(action.GraphQL.Document, action.Name) {
			t.Fatalf("write action %q must use a fixed GraphQL document", action.Name)
		}
		if action.Name == "delete_board" {
			sawDeleteBoard = true
			if action.Confirm != "destructive" {
				t.Fatalf("delete_board confirm = %q, want destructive", action.Confirm)
			}
		}
	}
	if !sawDeleteBoard {
		t.Fatal("delete_board write action missing")
	}

	var cli struct {
		Commands []struct {
			Path         string `json:"path"`
			Intent       string `json:"intent"`
			Availability string `json:"availability"`
			Write        string `json:"write"`
		} `json:"commands"`
	}
	readJSONFile(t, filepath.Join(root, "cli_surface.json"), &cli)
	var sawDeleteCommand bool
	for _, cmd := range cli.Commands {
		if cmd.Path == "reverse delete-board" && cmd.Intent == "reverse_etl" && cmd.Availability == "implemented" && cmd.Write == "delete_board" {
			sawDeleteCommand = true
		}
	}
	if !sawDeleteCommand {
		t.Fatal("implemented CLI metadata for reverse delete-board missing")
	}
}

func readJSONFile(t *testing.T, path string, v any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
}
