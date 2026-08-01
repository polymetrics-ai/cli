package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestXeroAPISurfaceOperationLedgerMetrics(t *testing.T) {
	root := filepath.Join("..", "..", "internal", "connectors", "defs", "xero")
	raw, err := os.ReadFile(filepath.Join(root, "api_surface.json"))
	if err != nil {
		t.Fatalf("read xero api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string            `json:"method"`
			Path      string            `json:"path"`
			CoveredBy map[string]string `json:"covered_by"`
			Excluded  map[string]any    `json:"excluded"`
			Operation *struct {
				Model            string `json:"model"`
				Status           string `json:"status"`
				BlockedByDefault bool   `json:"blocked_by_default"`
				Reason           string `json:"reason"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal xero api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}
	if len(surface.Endpoints) != 235 {
		t.Fatalf("endpoints = %d, want 235", len(surface.Endpoints))
	}

	methods := map[string]int{}
	lanes := map[string]int{}
	covered, operations, excluded := 0, 0, 0
	for i, ep := range surface.Endpoints {
		methods[ep.Method]++
		if strings.HasPrefix(ep.Path, "/api.xro/2.0") {
			t.Fatalf("endpoint %d retains stale /api.xro/2.0 prefix: %s", i, ep.Path)
		}
		if len(ep.Excluded) > 0 {
			excluded++
		}
		if len(ep.CoveredBy) > 0 {
			covered++
		}
		if ep.Operation != nil {
			operations++
			if ep.Operation.Status != "blocked" || !ep.Operation.BlockedByDefault || strings.TrimSpace(ep.Operation.Reason) == "" {
				t.Fatalf("endpoint %d has invalid blocked operation metadata: %+v", i, ep.Operation)
			}
		}
		switch {
		case isXeroBinaryPath(ep.Path):
			lanes["binary_file"]++
		case ep.Method == "GET" && strings.HasPrefix(ep.Path, "/Reports"):
			lanes["direct_read_query_search"]++
		case ep.Method == "GET":
			lanes["etl_read"]++
		default:
			lanes["reverse_etl_write"]++
		}
	}

	assertStringIntMap(t, "xero methods", methods, map[string]int{"DELETE": 10, "GET": 126, "POST": 46, "PUT": 53})
	assertStringIntMap(t, "xero lanes", lanes, map[string]int{
		"binary_file":              59,
		"direct_read_query_search": 11,
		"etl_read":                 78,
		"reverse_etl_write":        87,
	})
	if covered != 187 || operations != 48 || excluded != 0 {
		t.Fatalf("coverage counts covered/operations/excluded = %d/%d/%d, want 187/48/0", covered, operations, excluded)
	}
}

func TestXeroExecutableFixtureCoverage(t *testing.T) {
	root := filepath.Join("..", "..", "internal", "connectors", "defs", "xero")

	var streams struct {
		Streams []struct {
			Name string `json:"name"`
		} `json:"streams"`
	}
	readJSONFile(t, filepath.Join(root, "streams.json"), &streams)
	if len(streams.Streams) != 100 {
		t.Fatalf("streams = %d, want 100", len(streams.Streams))
	}
	for _, stream := range streams.Streams {
		if _, err := os.Stat(filepath.Join(root, "fixtures", "streams", stream.Name, "page_1.json")); err != nil {
			t.Fatalf("stream %q missing page_1 fixture: %v", stream.Name, err)
		}
	}

	var writes struct {
		Actions []struct {
			Name         string         `json:"name"`
			Path         string         `json:"path"`
			BodyType     string         `json:"body_type"`
			RecordSchema map[string]any `json:"record_schema"`
			Confirm      string         `json:"confirm"`
			Kind         string         `json:"kind"`
		} `json:"actions"`
	}
	readJSONFile(t, filepath.Join(root, "writes.json"), &writes)
	if len(writes.Actions) != 87 {
		t.Fatalf("write actions = %d, want 87", len(writes.Actions))
	}
	historyWrites := 0
	for _, action := range writes.Actions {
		if _, err := os.Stat(filepath.Join(root, "fixtures", "writes", action.Name+".json")); err != nil {
			t.Fatalf("write %q missing fixture: %v", action.Name, err)
		}
		if got := action.RecordSchema["additionalProperties"]; !reflect.DeepEqual(got, false) {
			t.Fatalf("write %q record_schema additionalProperties = %v, want false", action.Name, got)
		}
		if strings.HasSuffix(action.Path, "/History") {
			historyWrites++
			if action.BodyType != "json" {
				t.Fatalf("history write %q body_type = %q, want json", action.Name, action.BodyType)
			}
			if !schemaRequiredIncludes(action.RecordSchema, "HistoryRecords") {
				t.Fatalf("history write %q record_schema required missing HistoryRecords", action.Name)
			}
			if !schemaPropertiesIncludes(action.RecordSchema, "HistoryRecords") {
				t.Fatalf("history write %q record_schema properties missing HistoryRecords", action.Name)
			}
		}
		if action.Kind == "delete" && action.Confirm != "destructive" {
			t.Fatalf("delete write %q confirm = %q, want destructive", action.Name, action.Confirm)
		}
	}
	if historyWrites != 16 {
		t.Fatalf("history writes = %d, want 16", historyWrites)
	}
}

func TestXeroOperationsLedgerMetrics(t *testing.T) {
	root := filepath.Join("..", "..", "internal", "connectors", "defs", "xero")
	var ledger struct {
		Operations []struct {
			ID   string `json:"id"`
			Kind string `json:"kind"`
		} `json:"operations"`
	}
	readJSONFile(t, filepath.Join(root, "operations.json"), &ledger)
	if len(ledger.Operations) != 70 {
		t.Fatalf("operations = %d, want 70", len(ledger.Operations))
	}
	kinds := map[string]int{}
	seen := map[string]bool{}
	for _, op := range ledger.Operations {
		if seen[op.ID] {
			t.Fatalf("duplicate operation id %q", op.ID)
		}
		seen[op.ID] = true
		kinds[op.Kind]++
	}
	assertStringIntMap(t, "xero operation kinds", kinds, map[string]int{"binary_download": 26, "file_upload": 22, "rest_read": 22})
}

func schemaRequiredIncludes(schema map[string]any, field string) bool {
	required, ok := schema["required"].([]any)
	if !ok {
		return false
	}
	for _, item := range required {
		name, ok := item.(string)
		if ok && name == field {
			return true
		}
	}
	return false
}

func schemaPropertiesIncludes(schema map[string]any, field string) bool {
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return false
	}
	_, ok = properties[field]
	return ok
}

func isXeroBinaryPath(path string) bool {
	return strings.Contains(path, "Attachments") || strings.HasSuffix(strings.ToLower(path), "/pdf")
}

func readJSONFile(t *testing.T, path string, out any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
}
