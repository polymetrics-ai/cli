package main

import (
	"encoding/json"
	"os"
	"reflect"
	"regexp"
	"strings"
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

	if len(surface.Endpoints) != 1604 {
		t.Fatalf("endpoints = %d, want 1604 (1596 official rows plus 8 connector conformance coverage rows)", len(surface.Endpoints))
	}
	if covered != 440 {
		t.Fatalf("covered endpoints = %d, want 440", covered)
	}
	if operations != 1164 {
		t.Fatalf("operation endpoints = %d, want 1164 blocked/planned rows", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"DELETE":  187,
		"GET":     634,
		"GRAPHQL": 309,
		"PATCH":   73,
		"POST":    193,
		"PUT":     133,
		"WEBHOOK": 75,
	})
	assertStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"DELETE":  67,
		"GET":     205,
		"GRAPHQL": 4,
		"PATCH":   34,
		"POST":    85,
		"PUT":     45,
	})
	assertStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{
		"DELETE":  120,
		"GET":     429,
		"GRAPHQL": 305,
		"PATCH":   39,
		"POST":    108,
		"PUT":     88,
		"WEBHOOK": 75,
	})
	assertStringIntMap(t, "models", models, map[string]int{
		"admin_reverse_etl":     402,
		"destructive_action":    170,
		"direct_read":           466,
		"disallowed":            1,
		"duplicate":             53,
		"deprecated":            30,
		"sensitive_reverse_etl": 42,
	})
	assertStringIntMap(t, "risks", risks, map[string]int{
		"critical": 56,
		"high":     618,
		"low":      413,
		"medium":   77,
	})
	assertStringIntMap(t, "statuses", statuses, map[string]int{
		"blocked": 1164,
	})
}

func TestGitHubDestructiveMetadataUsesTypedConfirmation(t *testing.T) {
	writesRaw, err := os.ReadFile("../../internal/connectors/defs/github/writes.json")
	if err != nil {
		t.Fatalf("read github writes.json: %v", err)
	}
	var writes struct {
		Actions []githubWriteAction `json:"actions"`
	}
	if err := json.Unmarshal(writesRaw, &writes); err != nil {
		t.Fatalf("unmarshal github writes.json: %v", err)
	}
	destructiveWrites := map[string]bool{}
	for _, action := range writes.Actions {
		if githubWritePathHasLiteralPlaceholder(action.Path) {
			t.Fatalf("write action %q path uses single-brace placeholders: %q", action.Name, action.Path)
		}
		if githubActionRequiresTypedDestructiveConfirmation(action) && action.Confirm != "destructive" {
			t.Fatalf("write action %q confirm = %q, want destructive", action.Name, action.Confirm)
		}
		if action.Confirm == "destructive" {
			destructiveWrites[action.Name] = true
		}
		if action.Delete != nil {
			t.Fatalf("write action %q declares missing_ok_status without GitHub already-absent proof: %+v", action.Name, action.Delete)
		}
	}

	cliRaw, err := os.ReadFile("../../internal/connectors/defs/github/cli_surface.json")
	if err != nil {
		t.Fatalf("read github cli_surface.json: %v", err)
	}
	var cli struct {
		Commands []githubCLICommand `json:"commands"`
	}
	if err := json.Unmarshal(cliRaw, &cli); err != nil {
		t.Fatalf("unmarshal github cli_surface.json: %v", err)
	}
	unsafe := map[string]bool{}
	for _, cmd := range cli.Commands {
		if cmd.Availability == "unsafe_or_disallowed" {
			unsafe[cmd.Path] = true
		}
		if destructiveWrites[cmd.Write] && !strings.Contains(cmd.Approval, "destructive") {
			t.Fatalf("command %q maps destructive write %q but approval omits typed destructive confirmation: %q", cmd.Path, cmd.Write, cmd.Approval)
		}
	}
	assertStringBoolMap(t, "unsafe_or_disallowed commands", unsafe, map[string]bool{
		"api":        true,
		"auth token": true,
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

type githubWriteAction struct {
	Name    string              `json:"name"`
	Kind    string              `json:"kind"`
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Risk    string              `json:"risk"`
	Confirm string              `json:"confirm"`
	Delete  *githubDeletePolicy `json:"delete"`
}

type githubDeletePolicy struct {
	Idempotent      bool  `json:"idempotent"`
	MissingOKStatus []int `json:"missing_ok_status"`
}

type githubCLICommand struct {
	Path         string `json:"path"`
	Availability string `json:"availability"`
	Write        string `json:"write"`
	Approval     string `json:"approval"`
}

var githubSingleBracePathParamRE = regexp.MustCompile(`(^|[^{])\{[A-Za-z0-9_]+\}([^}]|$)`)

var githubDestructiveRiskPhrases = []string{
	"irreversibly",
	"repository history",
	"writes a commit",
	"ci/cd",
	"deploy access",
	"protection rules",
	"replaces every",
	"clearing its approval",
	"merge commit",
	"discarding history",
	"suppress a real",
	"deployment automation",
	"ruleset",
	"grants a github user access",
}

func githubWritePathHasLiteralPlaceholder(path string) bool {
	return githubSingleBracePathParamRE.MatchString(path)
}

func githubActionRequiresTypedDestructiveConfirmation(action githubWriteAction) bool {
	if action.Method == "DELETE" || action.Kind == "delete" {
		return true
	}
	risk := strings.ToLower(action.Risk)
	if risk == "critical" || risk == "high" {
		return true
	}
	for _, phrase := range githubDestructiveRiskPhrases {
		if strings.Contains(risk, phrase) {
			return true
		}
	}
	return false
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

func assertStringBoolMap(t *testing.T, name string, got, want map[string]bool) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}
