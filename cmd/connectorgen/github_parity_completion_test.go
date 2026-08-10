package main

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"
)

// TestGitHubCompleteParityHasNoNonterminalCommandRows records the captain's
// completion contract at the source boundary. It intentionally reads the
// generated GitHub bundle rather than duplicating commandrunner's admission
// logic: engine/runtime tests prove execution; this test makes an unresolved
// source classification impossible to hide behind an old parity summary.
func TestGitHubCompleteParityHasNoNonterminalCommandRows(t *testing.T) {
	t.Helper()

	cliRaw, err := os.ReadFile("../../internal/connectors/defs/github/cli_surface.json")
	if err != nil {
		t.Fatalf("read github cli surface: %v", err)
	}
	var cli struct {
		Commands []struct {
			Path         string `json:"path"`
			Availability string `json:"availability"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(cliRaw, &cli); err != nil {
		t.Fatalf("unmarshal github cli surface: %v", err)
	}

	var unresolvedCommands []string
	for _, command := range cli.Commands {
		switch command.Availability {
		case "partial", "planned", "unsafe_or_disallowed":
			unresolvedCommands = append(unresolvedCommands, command.Availability+":"+command.Path)
		}
	}
	sort.Strings(unresolvedCommands)
	if len(unresolvedCommands) != 0 {
		t.Errorf("GitHub completion still has %d nonterminal command rows: %s", len(unresolvedCommands), strings.Join(unresolvedCommands, ", "))
	}

	surfaceRaw, err := os.ReadFile("../../internal/connectors/defs/github/api_surface.json")
	if err != nil {
		t.Fatalf("read github api surface: %v", err)
	}
	var surface struct {
		Endpoints []struct {
			Method    string          `json:"method"`
			Path      string          `json:"path"`
			CoveredBy json.RawMessage `json:"covered_by"`
			Operation *struct {
				Model string `json:"model"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(surfaceRaw, &surface); err != nil {
		t.Fatalf("unmarshal github api surface: %v", err)
	}

	var unresolvedEndpoints []string
	for _, endpoint := range surface.Endpoints {
		if endpoint.Method == "GRAPHQL" || len(endpoint.CoveredBy) != 0 {
			continue
		}
		if endpoint.Operation != nil && endpoint.Operation.Model == "named_dependency" {
			continue
		}
		unresolvedEndpoints = append(unresolvedEndpoints, endpoint.Method+" "+endpoint.Path)
	}
	sort.Strings(unresolvedEndpoints)
	if len(unresolvedEndpoints) != 0 {
		t.Errorf("GitHub completion still has %d REST endpoints without a fixed executable contract or named dependency: %s", len(unresolvedEndpoints), strings.Join(unresolvedEndpoints, ", "))
	}
}
