package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestHelpScoutAllOperationsCovered(t *testing.T) {
	surfaceRaw, err := os.ReadFile("../../internal/connectors/defs/help-scout/api_surface.json")
	if err != nil {
		t.Fatalf("read help-scout api_surface.json: %v", err)
	}
	writesRaw, err := os.ReadFile("../../internal/connectors/defs/help-scout/writes.json")
	if err != nil {
		t.Fatalf("read help-scout writes.json: %v", err)
	}
	cliRaw, err := os.ReadFile("../../internal/connectors/defs/help-scout/cli_surface.json")
	if err != nil {
		t.Fatalf("read help-scout cli_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string                 `json:"method"`
			Path      string                 `json:"path"`
			CoveredBy map[string]any         `json:"covered_by"`
			Operation map[string]interface{} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(surfaceRaw, &surface); err != nil {
		t.Fatalf("unmarshal help-scout api_surface.json: %v", err)
	}
	var writes struct {
		Actions []struct {
			Name string `json:"name"`
		} `json:"actions"`
	}
	if err := json.Unmarshal(writesRaw, &writes); err != nil {
		t.Fatalf("unmarshal help-scout writes.json: %v", err)
	}
	var cli struct {
		Commands []struct {
			Path         string `json:"path"`
			Intent       string `json:"intent"`
			Availability string `json:"availability"`
			Write        string `json:"write"`
			Stream       string `json:"stream"`
			Operation    string `json:"operation"`
			OutputPolicy string `json:"output_policy"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(cliRaw, &cli); err != nil {
		t.Fatalf("unmarshal help-scout cli_surface.json: %v", err)
	}

	writeActions := map[string]bool{}
	for _, action := range writes.Actions {
		writeActions[action.Name] = true
	}

	implementedDirectReads := 0
	implementedWrites := 0
	binaryOperations := 0
	covered := 0
	for i, ep := range surface.Endpoints {
		if len(ep.CoveredBy) > 0 {
			covered++
		}
		if ep.Method == "GET" {
			if ep.Operation != nil {
				if ep.Operation["model"] == "binary_read" {
					binaryOperations++
					continue
				}
				t.Fatalf("GET endpoint %d %s has blocked operation %+v; want stream/direct_read/binary coverage", i, ep.Path, ep.Operation)
			}
			continue
		}
		writeName, _ := ep.CoveredBy["write"].(string)
		if writeName == "" {
			t.Fatalf("mutation endpoint %d %s %s missing covered_by.write", i, ep.Method, ep.Path)
		}
		if !writeActions[writeName] {
			t.Fatalf("mutation endpoint %d %s %s references missing write action %q", i, ep.Method, ep.Path, writeName)
		}
	}

	for _, cmd := range cli.Commands {
		switch {
		case cmd.Availability == "implemented" && cmd.Intent == "direct_read":
			implementedDirectReads++
			if cmd.OutputPolicy != "json" {
				t.Fatalf("direct_read command %q output_policy = %q, want json", cmd.Path, cmd.OutputPolicy)
			}
		case cmd.Availability == "implemented" && cmd.Intent == "reverse_etl" && cmd.Write != "":
			implementedWrites++
			if !writeActions[cmd.Write] {
				t.Fatalf("write command %q references missing write action %q", cmd.Path, cmd.Write)
			}
		}
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}
	if len(surface.Endpoints) != 145 {
		t.Fatalf("endpoints = %d, want 145", len(surface.Endpoints))
	}
	if len(writes.Actions) != 66 {
		t.Fatalf("write actions = %d, want 66", len(writes.Actions))
	}
	if implementedDirectReads != 73 {
		t.Fatalf("implemented direct-read commands = %d, want 73", implementedDirectReads)
	}
	if implementedWrites != 66 {
		t.Fatalf("implemented write commands = %d, want 66", implementedWrites)
	}
	if binaryOperations != 2 {
		t.Fatalf("binary operation endpoints = %d, want 2", binaryOperations)
	}
	if covered != 143 {
		t.Fatalf("covered endpoints = %d, want 143", covered)
	}
}
