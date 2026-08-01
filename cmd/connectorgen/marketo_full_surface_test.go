package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestMarketoFullSurfaceCommandAndOperationCoverage(t *testing.T) {
	api := loadMarketoJSON[struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string         `json:"method"`
			Path      string         `json:"path"`
			CoveredBy map[string]any `json:"covered_by"`
			Excluded  map[string]any `json:"excluded"`
			Operation map[string]any `json:"operation"`
		} `json:"endpoints"`
	}](t, "../../internal/connectors/defs/marketo/api_surface.json")
	streams := loadMarketoJSON[struct {
		Streams []struct {
			Name string `json:"name"`
		} `json:"streams"`
	}](t, "../../internal/connectors/defs/marketo/streams.json")
	writes := loadMarketoJSON[struct {
		Actions []struct {
			Name    string `json:"name"`
			Kind    string `json:"kind"`
			Method  string `json:"method"`
			Confirm string `json:"confirm"`
		} `json:"actions"`
	}](t, "../../internal/connectors/defs/marketo/writes.json")
	ops := loadMarketoJSON[struct {
		Operations []struct {
			ID           string          `json:"id"`
			Kind         string          `json:"kind"`
			OutputPolicy string          `json:"output_policy"`
			REST         json.RawMessage `json:"rest"`
		} `json:"operations"`
	}](t, "../../internal/connectors/defs/marketo/operations.json")
	cli := loadMarketoJSON[struct {
		Commands []struct {
			Path         string `json:"path"`
			Intent       string `json:"intent"`
			Availability string `json:"availability"`
			Stream       string `json:"stream"`
			Write        string `json:"write"`
			Operation    string `json:"operation"`
			OutputPolicy string `json:"output_policy"`
		} `json:"commands"`
	}](t, "../../internal/connectors/defs/marketo/cli_surface.json")

	if api.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", api.OperationLedgerVersion)
	}
	if got, want := len(api.Endpoints), 327; got != want {
		t.Fatalf("api_surface endpoints = %d, want %d", got, want)
	}
	if got, want := len(streams.Streams), 117; got != want {
		t.Fatalf("streams = %d, want %d", got, want)
	}
	if got, want := len(writes.Actions), 158; got != want {
		t.Fatalf("write actions = %d, want %d", got, want)
	}
	if got, want := len(ops.Operations), 28; got != want {
		t.Fatalf("operations = %d, want %d", got, want)
	}
	if got, want := len(cli.Commands), 303; got != want {
		t.Fatalf("cli commands = %d, want %d", got, want)
	}

	coverage := map[string]int{}
	for i, ep := range api.Endpoints {
		if len(ep.Excluded) > 0 {
			t.Fatalf("endpoint %d %s %s uses legacy excluded row in operation-ledger mode", i, ep.Method, ep.Path)
		}
		if len(ep.CoveredBy) > 0 {
			for key := range ep.CoveredBy {
				coverage[key]++
			}
		}
		if len(ep.Operation) > 0 {
			coverage["operation"]++
			if ep.Operation["blocked_by_default"] != true {
				t.Fatalf("endpoint %d operation is not blocked by default: %+v", i, ep.Operation)
			}
			if ep.Operation["reason"] == "" {
				t.Fatalf("endpoint %d operation has empty reason: %+v", i, ep.Operation)
			}
		}
	}
	wantCoverage := map[string]int{"stream": 117, "direct_read": 28, "write": 158, "operation": 24}
	for key, want := range wantCoverage {
		if got := coverage[key]; got != want {
			t.Fatalf("coverage[%s] = %d, want %d (all coverage: %+v)", key, got, want, coverage)
		}
	}

	commandIntents := map[string]int{}
	for _, cmd := range cli.Commands {
		if cmd.Availability != "implemented" {
			t.Fatalf("command %q availability = %q, want implemented", cmd.Path, cmd.Availability)
		}
		commandIntents[cmd.Intent]++
		if cmd.Intent == "direct_read" && cmd.OutputPolicy != "json_redacted" {
			t.Fatalf("direct command %q output_policy = %q, want json_redacted", cmd.Path, cmd.OutputPolicy)
		}
	}
	wantCommandIntents := map[string]int{"etl": 117, "direct_read": 28, "reverse_etl": 158}
	for key, want := range wantCommandIntents {
		if got := commandIntents[key]; got != want {
			t.Fatalf("commands[%s] = %d, want %d (all commands: %+v)", key, got, want, commandIntents)
		}
	}

	for _, op := range ops.Operations {
		if op.Kind != "rest_read" || op.OutputPolicy != "json_redacted" || !json.Valid(op.REST) {
			t.Fatalf("operation %q = kind %q policy %q rest valid %t, want rest_read/json_redacted/valid", op.ID, op.Kind, op.OutputPolicy, json.Valid(op.REST))
		}
	}
}

func TestMarketoWriteAndCLISafetyInvariants(t *testing.T) {
	writes := loadMarketoJSON[struct {
		Actions []struct {
			Name         string         `json:"name"`
			Kind         string         `json:"kind"`
			Path         string         `json:"path"`
			PathFields   []string       `json:"path_fields"`
			RedactFields []string       `json:"redact_fields"`
			Confirm      string         `json:"confirm"`
			RecordSchema map[string]any `json:"record_schema"`
		} `json:"actions"`
	}](t, "../../internal/connectors/defs/marketo/writes.json")
	cli := loadMarketoJSON[struct {
		Commands []struct {
			Path       string   `json:"path"`
			Intent     string   `json:"intent"`
			Notes      string   `json:"notes"`
			Examples   []string `json:"examples"`
			APISurface []struct {
				Path string `json:"path"`
			} `json:"api_surface"`
			Flags []struct {
				Name   string `json:"name"`
				MapsTo string `json:"maps_to"`
			} `json:"flags"`
		} `json:"commands"`
	}](t, "../../internal/connectors/defs/marketo/cli_surface.json")

	writeByName := map[string]map[string]any{}
	for _, action := range writes.Actions {
		if _, exists := writeByName[action.Name]; exists {
			t.Fatalf("duplicate write action %q", action.Name)
		}
		writeByName[action.Name] = action.RecordSchema

		assertMarketoQueryTemplatesAreEncodedAndRedacted(t, action.Name, action.Path, action.RecordSchema, action.RedactFields)
		assertMarketoWriteSchemaHasNoResponseOnlyFields(t, action.Name, action.RecordSchema, "record")
		if action.Kind == "delete" || action.Confirm == "destructive" {
			if action.Confirm != "destructive" {
				t.Fatalf("destructive write %q confirm = %q, want destructive", action.Name, action.Confirm)
			}
			required := marketoRequiredSet(action.RecordSchema)
			if len(required) == 0 {
				t.Fatalf("destructive write %q has no required top-level target fields", action.Name)
			}
			for _, field := range action.PathFields {
				if strings.Contains(field, ".") {
					continue
				}
				if !required[field] {
					t.Fatalf("destructive write %q path field %q is not required", action.Name, field)
				}
			}
			for _, arrayField := range []string{"input", "attributes"} {
				if raw, hasArray := marketoProperties(action.RecordSchema)[arrayField]; hasArray {
					if !required[arrayField] {
						t.Fatalf("destructive write %q has %s target body but does not require it", action.Name, arrayField)
					}
					if itemRequired := marketoArrayItemRequiredSet(raw); len(itemRequired) == 0 {
						t.Fatalf("destructive write %q array field %q has no required item selector fields", action.Name, arrayField)
					}
				}
			}
		}
	}
	assertMarketoRequiredFields(t, writeByName, "delete_token_by_name", "id", "folderType", "name", "type")
	assertMarketoRequiredFields(t, writeByName, "delete_companies", "deleteBy", "input")

	for _, cmd := range cli.Commands {
		seen := map[string]string{}
		for _, flag := range cmd.Flags {
			if flag.Name == "" {
				continue
			}
			if flag.Name != strings.ToLower(flag.Name) || strings.ContainsAny(flag.Name, " _") {
				t.Fatalf("command %q has non-kebab flag %q", cmd.Path, flag.Name)
			}
			if prev, exists := seen[flag.Name]; exists {
				t.Fatalf("command %q duplicates flag %q for %s and %s", cmd.Path, flag.Name, prev, flag.MapsTo)
			}
			seen[flag.Name] = flag.MapsTo
		}
		if cmd.Intent == "etl" || cmd.Intent == "direct_read" {
			assertMarketoPathParameterHelp(t, cmd.Path, cmd.Notes, cmd.Examples, cmd.APISurface)
		}
		if cmd.Intent == "direct_read" {
			for _, flag := range cmd.Flags {
				if strings.HasPrefix(flag.MapsTo, "path.") {
					needle := "--" + flag.Name
					if !strings.Contains(cmd.Notes, needle) {
						t.Fatalf("direct command %q path flag %q missing from notes", cmd.Path, needle)
					}
					found := false
					for _, example := range cmd.Examples {
						if strings.Contains(example, needle) {
							found = true
							break
						}
					}
					if !found {
						t.Fatalf("direct command %q path flag %q missing from examples", cmd.Path, needle)
					}
				}
			}
		}
	}
}

func assertMarketoPathParameterHelp(t *testing.T, commandPath, notes string, examples []string, apiSurface []struct {
	Path string `json:"path"`
}) {
	t.Helper()
	params := []string{}
	seen := map[string]bool{}
	for _, api := range apiSurface {
		for _, match := range marketoPathParam.FindAllStringSubmatch(api.Path, -1) {
			param := marketoConfigKeyForPathParam(match[1])
			if !seen[param] {
				params = append(params, param)
				seen[param] = true
			}
		}
	}
	if len(params) == 0 {
		return
	}
	for _, param := range params {
		if !strings.Contains(notes, param) {
			t.Fatalf("command %q has path param %q but notes do not mention the required config override", commandPath, param)
		}
		requiredExample := "--config " + param + "="
		found := false
		for _, example := range examples {
			if strings.Contains(example, requiredExample) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("command %q has path param %q but examples do not include %q", commandPath, param, requiredExample)
		}
	}
}

var marketoPathParam = regexp.MustCompile(`\{([^}]+)\}`)

func marketoConfigKeyForPathParam(param string) string {
	var b strings.Builder
	for i, r := range param {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

func assertMarketoQueryTemplatesAreEncodedAndRedacted(t *testing.T, actionName, path string, _ map[string]any, _ []string) {
	t.Helper()
	if strings.Contains(path, "?") {
		t.Fatalf("write action %q embeds query parameters in action.path; block it until writes support a structured typed query map: %s", actionName, path)
	}
}

func assertMarketoWriteSchemaHasNoResponseOnlyFields(t *testing.T, actionName string, schema map[string]any, path string) {
	t.Helper()
	for name, raw := range marketoProperties(schema) {
		child, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		desc := strings.ToLower(fmt.Sprint(child["description"]))
		if name == "errors" || name == "reasons" || name == "seq" || strings.Contains(desc, "should only be part of responses") || strings.Contains(desc, "should not be submitted") || strings.Contains(desc, "status of the operation") || strings.Contains(desc, "reason for the status of the operation") {
			t.Fatalf("write action %q includes response-only field %s/%s", actionName, path, name)
		}
		assertMarketoWriteSchemaHasNoResponseOnlyFields(t, actionName, child, path+"/"+name)
		if items, ok := child["items"].(map[string]any); ok {
			assertMarketoWriteSchemaHasNoResponseOnlyFields(t, actionName, items, path+"/"+name+"[]")
		}
	}
}

func assertMarketoRequiredFields(t *testing.T, writeByName map[string]map[string]any, action string, fields ...string) {
	t.Helper()
	required := marketoRequiredSet(writeByName[action])
	if required == nil {
		t.Fatalf("write action %q not found", action)
	}
	for _, field := range fields {
		if !required[field] {
			t.Fatalf("write action %q does not require %q", action, field)
		}
	}
}

func marketoRequiredSet(schema map[string]any) map[string]bool {
	if schema == nil {
		return nil
	}
	out := map[string]bool{}
	for _, item := range marketoStringSlice(schema["required"]) {
		out[item] = true
	}
	return out
}

func marketoProperties(schema map[string]any) map[string]any {
	if schema == nil {
		return nil
	}
	props, _ := schema["properties"].(map[string]any)
	return props
}

func marketoArrayItemRequiredSet(raw any) map[string]bool {
	arraySchema, _ := raw.(map[string]any)
	items, _ := arraySchema["items"].(map[string]any)
	return marketoRequiredSet(items)
}

func marketoStringSlice(raw any) []string {
	items, _ := raw.([]any)
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func TestMarketoMetadataEnablesReadWriteCapabilities(t *testing.T) {
	metadata := loadMarketoJSON[struct {
		Capabilities struct {
			Read  bool `json:"read"`
			Write bool `json:"write"`
			Query bool `json:"query"`
		} `json:"capabilities"`
	}](t, "../../internal/connectors/defs/marketo/metadata.json")
	if !metadata.Capabilities.Read || !metadata.Capabilities.Write {
		t.Fatalf("Marketo capabilities read/write = %t/%t, want true/true", metadata.Capabilities.Read, metadata.Capabilities.Write)
	}
	if metadata.Capabilities.Query {
		t.Fatal("Marketo must not flip generic query capability for provider direct reads")
	}
}

func loadMarketoJSON[T any](t *testing.T, path string) T {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return out
}
