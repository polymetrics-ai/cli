package engine

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"polymetrics.ai/internal/connectors"
	"sync/atomic"
	"testing"
	"testing/fstest"
)

func requestInputBundle234(t *testing.T) fstest.MapFS {
	t.Helper()
	fsys := fullValidBundleFS("acme")
	var doc map[string]any
	if err := json.Unmarshal(fsys["acme/streams.json"].Data, &doc); err != nil {
		t.Fatal(err)
	}
	stream := doc["streams"].([]any)[0].(map[string]any)
	stream["query"] = map[string]any{"page_size": map[string]any{"template": "{{ config.page_size }}", "omit_when_absent": true}}
	stream["request_inputs"] = map[string]any{"version": 1, "schema": "schemas/inputs-widgets.json", "bindings": []any{map[string]any{"config_key": "page_size", "in": "query", "name": "page_size"}}}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	fsys["acme/streams.json"].Data = raw
	fsys["acme/schemas/inputs-widgets.json"] = &fstest.MapFile{Data: []byte(`{"type":"object","properties":{"path":{"type":"object","properties":{},"additionalProperties":false},"query":{"type":"object","properties":{"page_size":{"type":"integer","minimum":2,"maximum":5}},"additionalProperties":false},"header":{"type":"object","properties":{},"additionalProperties":false}},"required":["path","query","header"],"additionalProperties":false}`)}
	return fsys
}

func TestLoadRequestInputContract234(t *testing.T) {
	for _, name := range []string{"healthy", "missing_schema", "outside_schema", "missing_member", "wrong_container", "missing_binding", "duplicate_binding", "wrong_version", "wrong_wire_alias", "missing_wire_query"} {
		t.Run(name, func(t *testing.T) {
			fsys := requestInputBundle234(t)
			var doc map[string]any
			if err := json.Unmarshal(fsys["acme/streams.json"].Data, &doc); err != nil {
				t.Fatal(err)
			}
			contract := doc["streams"].([]any)[0].(map[string]any)["request_inputs"].(map[string]any)
			bindings := contract["bindings"].([]any)
			switch name {
			case "missing_schema":
				delete(fsys, "acme/schemas/inputs-widgets.json")
			case "outside_schema":
				contract["schema"] = "../inputs.json"
			case "missing_member":
				bindings[0].(map[string]any)["name"] = "absent"
			case "wrong_container":
				fsys["acme/schemas/inputs-widgets.json"].Data = []byte(`{"type":"string"}`)
			case "missing_binding":
				contract["bindings"] = []any{}
			case "duplicate_binding":
				contract["bindings"] = append(bindings, bindings[0])
			case "wrong_wire_alias":
				bindings[0].(map[string]any)["config_key"] = "other"
			case "missing_wire_query":
				delete(doc["streams"].([]any)[0].(map[string]any), "query")
			case "wrong_version":
				contract["version"] = 2
			}
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/streams.json"].Data = raw
			b, err := Load(fsys, "acme")
			if name == "healthy" {
				if err != nil {
					t.Fatal(err)
				}
				if len(b.Streams) != 1 {
					t.Fatal("lost selected stream")
				}
				return
			}
			if err == nil {
				t.Fatalf("Load admitted invalid %s request input contract", name)
			}
		})
	}
}

func TestOperationDirectReadInputContract234(t *testing.T) {
	for _, tc := range []struct {
		name, kind, raw   string
		required, invalid bool
	}{
		{"boolean_false", "boolean", "false", false, false},
		{"boolean_truthy", "boolean", "1", false, true},
		{"required_empty_string", "string", "", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var sends atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sends.Add(1)
				if values, exists := r.URL.Query()["page_size"]; !exists || len(values) != 1 || values[0] != tc.raw {
					t.Errorf("wire query: %s", r.URL.RawQuery)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":[{"id":1},{"id":2},{"id":3}]}`))
			}))
			defer server.Close()
			fsys := requestInputBundle234(t)
			var input map[string]any
			if err := json.Unmarshal(fsys["acme/schemas/inputs-widgets.json"].Data, &input); err != nil {
				t.Fatal(err)
			}
			query := input["properties"].(map[string]any)["query"].(map[string]any)
			query["properties"].(map[string]any)["page_size"] = map[string]any{"type": tc.kind}
			if tc.required {
				query["required"] = []string{"page_size"}
			}
			raw, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/schemas/inputs-widgets.json"].Data = raw
			var streams map[string]any
			if err := json.Unmarshal(fsys["acme/streams.json"].Data, &streams); err != nil {
				t.Fatal(err)
			}
			streams["base"].(map[string]any)["auth"] = []any{}
			stream := streams["streams"].([]any)[0].(map[string]any)
			stream["query"].(map[string]any)["page_size"].(map[string]any)["omit_when_absent"] = !tc.required
			raw, err = json.Marshal(streams)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/streams.json"].Data = raw
			operation := map[string]any{"id": "widgets", "kind": "rest_read", "summary": "widgets", "risk": "low", "approval": "none", "output_policy": "json_redacted", "rest": map[string]any{"method": "GET", "path": "/widgets", "max_bytes": 1024, "parameters": []any{map[string]any{"name": "page_size", "in": "query", "type": tc.kind, "required": tc.required}}, "request_inputs": stream["request_inputs"], "response": map[string]any{"success_statuses": []string{"200"}}}}
			raw, err = json.Marshal(map[string]any{"operations": []any{operation}})
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/operations.json"] = &fstest.MapFile{Data: raw}
			bundle, err := Load(fsys, "acme")
			if err != nil {
				t.Fatal(err)
			}
			result, err := OperationDirectRead(t.Context(), bundle, connectors.OperationDirectReadRequest{Operation: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL}}, Query: map[string]string{"page_size": tc.raw}}, nil)
			if tc.invalid {
				if err == nil || sends.Load() != 0 {
					t.Fatalf("invalid input: err=%v sends=%d", err, sends.Load())
				}
				return
			}
			if err != nil || sends.Load() != 1 {
				t.Fatalf("healthy input: err=%v sends=%d result=%+v", err, sends.Load(), result)
			}
			encoded, err := json.Marshal(result.Body)
			if err != nil || string(encoded) != `{"data":[{"id":1},{"id":2},{"id":3}]}` {
				t.Fatalf("returned rows: %s %v", encoded, err)
			}
		})
	}
}

func TestLoadRequestInputFlagJoin234(t *testing.T) {
	for _, name := range []string{"healthy", "missing_contract", "wrong_type", "wrong_mapping", "omitted_minimum", "omitted_codec"} {
		t.Run(name, func(t *testing.T) {
			fsys := requestInputBundle234(t)
			flag := map[string]any{"name": "page-size", "type": "integer", "maps_to": "config.page_size", "input_codec": "source_scalar_v1", "minimum": 2, "maximum": 5}
			switch name {
			case "wrong_type":
				flag["type"] = "string"
			case "wrong_mapping":
				flag["maps_to"] = "config.other"
			case "omitted_minimum":
				delete(flag, "minimum")
			case "omitted_codec":
				delete(flag, "input_codec")
			}
			if name == "missing_contract" {
				var doc map[string]any
				if err := json.Unmarshal(fsys["acme/streams.json"].Data, &doc); err != nil {
					t.Fatal(err)
				}
				delete(doc["streams"].([]any)[0].(map[string]any), "request_inputs")
				raw, err := json.Marshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				fsys["acme/streams.json"].Data = raw
			}
			raw, err := json.Marshal(map[string]any{"tagline": "widgets", "usage": "pm acme widgets", "commands": []any{map[string]any{"path": "widgets", "summary": "widgets", "intent": "etl", "availability": "implemented", "stream": "widgets", "flags": []any{flag}}}})
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/cli_surface.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(fsys, "acme")
			if name == "healthy" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Load admitted mismatched %s flag input contract", name)
			}
		})
	}
}
