package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestHelpScoutFullSurfaceLedgerAndSafety(t *testing.T) {
	api := loadHelpScoutJSON[struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string         `json:"method"`
			Path      string         `json:"path"`
			CoveredBy map[string]any `json:"covered_by"`
			Operation map[string]any `json:"operation"`
		} `json:"endpoints"`
	}](t, "../../internal/connectors/defs/help-scout/api_surface.json")
	streams := loadHelpScoutJSON[struct {
		Streams []struct {
			Name string `json:"name"`
		} `json:"streams"`
	}](t, "../../internal/connectors/defs/help-scout/streams.json")
	writes := loadHelpScoutJSON[struct {
		Actions []struct {
			Name         string   `json:"name"`
			Kind         string   `json:"kind"`
			Method       string   `json:"method"`
			PathFields   []string `json:"path_fields"`
			RedactFields []string `json:"redact_fields"`
			Confirm      string   `json:"confirm"`
		} `json:"actions"`
	}](t, "../../internal/connectors/defs/help-scout/writes.json")
	ops := loadHelpScoutJSON[struct {
		Operations []struct {
			ID              string `json:"id"`
			Kind            string `json:"kind"`
			Risk            string `json:"risk"`
			Approval        string `json:"approval"`
			OutputPolicy    string `json:"output_policy"`
			MutationClass   string `json:"mutation_class"`
			Destructive     bool   `json:"destructive"`
			SecretSensitive bool   `json:"secret_sensitive"`
		} `json:"operations"`
	}](t, "../../internal/connectors/defs/help-scout/operations.json")
	cli := loadHelpScoutJSON[struct {
		Commands []struct {
			Path         string `json:"path"`
			Intent       string `json:"intent"`
			Availability string `json:"availability"`
			Stream       string `json:"stream"`
			Write        string `json:"write"`
			Operation    string `json:"operation"`
			Risk         string `json:"risk"`
			Approval     string `json:"approval"`
		} `json:"commands"`
	}](t, "../../internal/connectors/defs/help-scout/cli_surface.json")

	if api.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", api.OperationLedgerVersion)
	}
	if got, want := len(api.Endpoints), 144; got != want {
		t.Fatalf("api_surface endpoints = %d, want %d", got, want)
	}
	if got, want := len(streams.Streams), 45; got != want {
		t.Fatalf("streams = %d, want %d", got, want)
	}
	if got, want := len(writes.Actions), 65; got != want {
		t.Fatalf("write actions = %d, want %d", got, want)
	}
	if got, want := len(ops.Operations), 34; got != want {
		t.Fatalf("blocked/planned operations = %d, want %d", got, want)
	}

	coverage := map[string]int{}
	seenEndpoint := map[string]bool{}
	for _, ep := range api.Endpoints {
		key := strings.ToUpper(ep.Method) + " " + ep.Path
		if seenEndpoint[key] {
			t.Fatalf("duplicate api_surface endpoint %s", key)
		}
		seenEndpoint[key] = true
		if ep.CoveredBy != nil {
			for k := range ep.CoveredBy {
				coverage[k]++
			}
		}
		if ep.Operation != nil {
			coverage["operation"]++
		}
	}
	wantCoverage := map[string]int{"stream": 45, "write": 65, "operation": 34}
	for key, want := range wantCoverage {
		if got := coverage[key]; got != want {
			t.Fatalf("coverage[%s] = %d, want %d (all coverage: %+v)", key, got, want, coverage)
		}
	}

	for _, action := range writes.Actions {
		if strings.EqualFold(action.Method, "DELETE") || action.Kind == "delete" || strings.Contains(strings.ToLower(action.Name), "delete") || strings.Contains(strings.ToLower(action.Name), "remove") {
			if action.Confirm != "destructive" {
				t.Fatalf("destructive action %q confirm = %q, want destructive", action.Name, action.Confirm)
			}
			if len(action.PathFields) > 0 && len(action.RedactFields) == 0 {
				t.Fatalf("destructive action %q has path_fields but no redact_fields", action.Name)
			}
		}
	}

	implemented := map[string]int{}
	for _, cmd := range cli.Commands {
		if cmd.Availability == "implemented" {
			implemented[cmd.Intent]++
		}
		if cmd.Intent == "reverse_etl" && (cmd.Availability == "implemented" || cmd.Availability == "partial") {
			if strings.TrimSpace(cmd.Risk) == "" || strings.TrimSpace(cmd.Approval) == "" {
				t.Fatalf("reverse ETL command %q lacks risk/approval text", cmd.Path)
			}
		}
	}
	if implemented["etl"] != 45 {
		t.Fatalf("implemented ETL commands = %d, want 45", implemented["etl"])
	}
	if implemented["reverse_etl"] != 65 {
		t.Fatalf("implemented reverse ETL commands = %d, want 65", implemented["reverse_etl"])
	}
}

func loadHelpScoutJSON[T any](t *testing.T, path string) T {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return out
}
