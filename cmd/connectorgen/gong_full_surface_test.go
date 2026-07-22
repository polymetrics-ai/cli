package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGongFullSurfaceCommandAndOperationCoverage(t *testing.T) {
	api := loadGongJSON[struct {
		Endpoints []struct {
			Method    string         `json:"method"`
			Path      string         `json:"path"`
			CoveredBy map[string]any `json:"covered_by"`
			Operation map[string]any `json:"operation"`
		} `json:"endpoints"`
	}](t, "../../internal/connectors/defs/gong/api_surface.json")
	cli := loadGongJSON[struct {
		Commands []struct {
			Path         string `json:"path"`
			Intent       string `json:"intent"`
			Availability string `json:"availability"`
			Stream       string `json:"stream"`
			Write        string `json:"write"`
			Operation    string `json:"operation"`
			OutputPolicy string `json:"output_policy"`
			Flags        []struct {
				Name   string `json:"name"`
				MapsTo string `json:"maps_to"`
			} `json:"flags"`
		} `json:"commands"`
	}](t, "../../internal/connectors/defs/gong/cli_surface.json")
	writes := loadGongJSON[struct {
		Actions []struct {
			Name    string `json:"name"`
			Kind    string `json:"kind"`
			Method  string `json:"method"`
			Path    string `json:"path"`
			Risk    string `json:"risk"`
			Confirm string `json:"confirm"`
		} `json:"actions"`
	}](t, "../../internal/connectors/defs/gong/writes.json")
	ops := loadGongJSON[struct {
		Operations []struct {
			ID              string          `json:"id"`
			Kind            string          `json:"kind"`
			Risk            string          `json:"risk"`
			Approval        string          `json:"approval"`
			OutputPolicy    string          `json:"output_policy"`
			MutationClass   string          `json:"mutation_class"`
			SecretSensitive bool            `json:"secret_sensitive"`
			REST            json.RawMessage `json:"rest"`
			SensitivePolicy json.RawMessage `json:"sensitive_policy"`
		} `json:"operations"`
	}](t, "../../internal/connectors/defs/gong/operations.json")

	if got, want := len(writes.Actions), 26; got != want {
		t.Fatalf("write actions = %d, want %d", got, want)
	}
	if got, want := len(ops.Operations), 16; got != want {
		t.Fatalf("operations = %d, want %d", got, want)
	}

	coverage := map[string]int{}
	for _, ep := range api.Endpoints {
		if ep.CoveredBy != nil {
			for key := range ep.CoveredBy {
				coverage[key]++
			}
		}
		if ep.Operation != nil {
			coverage["operation"]++
		}
	}
	wantCoverage := map[string]int{"stream": 12, "direct_read": 29, "write": 26}
	for key, want := range wantCoverage {
		if got := coverage[key]; got != want {
			t.Fatalf("coverage[%s] = %d, want %d (all coverage: %+v)", key, got, want, coverage)
		}
	}

	commandsByPath := map[string]struct {
		intent, availability, stream, write, operation, outputPolicy string
	}{}
	flagsByPath := map[string]map[string]string{}
	for _, cmd := range cli.Commands {
		commandsByPath[cmd.Path] = struct {
			intent, availability, stream, write, operation, outputPolicy string
		}{cmd.Intent, cmd.Availability, cmd.Stream, cmd.Write, cmd.Operation, cmd.OutputPolicy}
		flagsByPath[cmd.Path] = map[string]string{}
		for _, flag := range cmd.Flags {
			flagsByPath[cmd.Path][flag.Name] = flag.MapsTo
		}
		if cmd.Intent == "direct_read" && cmd.Availability != "implemented" {
			t.Fatalf("direct read command %q availability = %q, want implemented", cmd.Path, cmd.Availability)
		}
	}
	for _, tc := range []struct {
		path, intent, availability, target string
	}{
		{path: "calls list", intent: "etl", availability: "implemented", target: "calls"},
		{path: "workspaces list", intent: "etl", availability: "implemented", target: "workspaces"},
		{path: "calls get", intent: "direct_read", availability: "implemented", target: "json_redacted"},
		{path: "users get", intent: "direct_read", availability: "implemented", target: "json_redacted"},
		{path: "calls create", intent: "reverse_etl", availability: "partial", target: "add_call"},
		{path: "privacy erase-phone", intent: "reverse_etl", availability: "partial", target: "purge_phone_number"},
		{path: "calls extensive", intent: "direct_read", availability: "implemented", target: "json_redacted"},
		{path: "calls transcript", intent: "direct_read", availability: "implemented", target: "json_redacted"},
		{path: "meetings integration-status", intent: "direct_read", availability: "implemented", target: "json_redacted"},
		{path: "crm upload-entities", intent: "reverse_etl", availability: "implemented", target: "upload_crm_entities"},
	} {
		cmd, ok := commandsByPath[tc.path]
		if !ok {
			t.Fatalf("missing cli command %q", tc.path)
		}
		if cmd.intent != tc.intent || cmd.availability != tc.availability {
			t.Fatalf("command %q intent/availability = %s/%s, want %s/%s", tc.path, cmd.intent, cmd.availability, tc.intent, tc.availability)
		}
		switch tc.intent {
		case "etl":
			if cmd.stream != tc.target {
				t.Fatalf("command %q stream = %q, want %q", tc.path, cmd.stream, tc.target)
			}
		case "direct_read":
			if tc.availability == "implemented" && cmd.outputPolicy != tc.target {
				t.Fatalf("command %q output_policy = %q, want %q", tc.path, cmd.outputPolicy, tc.target)
			}
			if tc.availability != "implemented" && cmd.operation != tc.target {
				t.Fatalf("command %q operation = %q, want %q", tc.path, cmd.operation, tc.target)
			}
		case "reverse_etl":
			if (tc.availability == "partial" || tc.availability == "implemented") && cmd.write != tc.target {
				t.Fatalf("command %q write = %q, want %q", tc.path, cmd.write, tc.target)
			}
			if tc.availability == "planned" && cmd.operation != tc.target {
				t.Fatalf("command %q operation = %q, want %q", tc.path, cmd.operation, tc.target)
			}
		}
	}

	writesByName := map[string]struct{ kind, method, path, risk, confirm string }{}
	for _, action := range writes.Actions {
		writesByName[action.Name] = struct{ kind, method, path, risk, confirm string }{action.Kind, action.Method, action.Path, action.Risk, action.Confirm}
	}
	if got := flagsByPath["calls transcript"]["call-id"]; got != "body.filter.callIds" {
		t.Fatalf("calls transcript --call-id maps_to = %q, want body.filter.callIds", got)
	}
	if _, exists := flagsByPath["calls transcript"]["body"]; exists {
		t.Fatal("calls transcript must not expose a raw body flag")
	}

	for _, name := range []string{"add_call", "update_permission_profile", "delete_meeting", "integration_settings", "purge_phone_number", "update_task", "upload_call_media", "upload_crm_entities", "upload_crm_entity_schema"} {
		if _, ok := writesByName[name]; !ok {
			t.Fatalf("missing write action %q", name)
		}
	}
	if writesByName["delete_meeting"].confirm != "destructive" || writesByName["purge_phone_number"].confirm != "destructive" {
		t.Fatalf("destructive Gong writes must require destructive confirmation: %+v %+v", writesByName["delete_meeting"], writesByName["purge_phone_number"])
	}

	opsByID := map[string]struct {
		kind, risk, approval, outputPolicy, mutationClass string
		secretSensitive                                   bool
		rest, sensitivePolicy                             json.RawMessage
	}{}
	for _, op := range ops.Operations {
		opsByID[op.ID] = struct {
			kind, risk, approval, outputPolicy, mutationClass string
			secretSensitive                                   bool
			rest, sensitivePolicy                             json.RawMessage
		}{op.Kind, op.Risk, op.Approval, op.OutputPolicy, op.MutationClass, op.SecretSensitive, op.REST, op.SensitivePolicy}
	}
	for _, id := range []string{"gong.calls_extensive", "gong.stats_interaction", "gong.calls_transcript", "gong.calls_media_upload", "gong.crm_upload_entities"} {
		if _, ok := opsByID[id]; !ok {
			t.Fatalf("missing operation %q", id)
		}
	}
	if opsByID["gong.calls_extensive"].kind != "rest_read" || !json.Valid(opsByID["gong.calls_extensive"].rest) {
		t.Fatalf("calls extensive operation = %+v, want typed rest_read", opsByID["gong.calls_extensive"])
	}
	if opsByID["gong.calls_media_upload"].mutationClass == "" || len(opsByID["gong.calls_media_upload"].sensitivePolicy) == 0 {
		t.Fatalf("media upload operation missing mutation class or sensitive policy: %+v", opsByID["gong.calls_media_upload"])
	}
}

func TestGongMetadataEnablesWriteCapability(t *testing.T) {
	metadata := loadGongJSON[struct {
		Capabilities struct {
			Read  bool `json:"read"`
			Write bool `json:"write"`
		} `json:"capabilities"`
	}](t, "../../internal/connectors/defs/gong/metadata.json")
	if !metadata.Capabilities.Read || !metadata.Capabilities.Write {
		t.Fatalf("Gong capabilities read/write = %t/%t, want true/true", metadata.Capabilities.Read, metadata.Capabilities.Write)
	}
}

func loadGongJSON[T any](t *testing.T, path string) T {
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
