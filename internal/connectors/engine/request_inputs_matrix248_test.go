package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
)

// These literal expectations are independent of the input compiler. Each
// adapter must send the same effective value and return all three known rows.
func TestRequestInputPresenceMatrix248(t *testing.T) {
	for _, adapter := range []string{"exported", "connector", "direct"} {
		for _, tc := range []struct {
			name, schema, config, override, wire    string
			configPresent, overridePresent, invalid bool
		}{
			{name: "optional_absent", schema: `{"type":"string"}`},
			{name: "empty_present", schema: `{"type":"string"}`, configPresent: true, wire: "page_size="},
			{name: "false_present", schema: `{"type":"boolean","default":true}`, config: "false", configPresent: true, wire: "page_size=false"},
			{name: "zero_present", schema: `{"type":"integer","default":3}`, config: "0", configPresent: true, wire: "page_size=0"},
			{name: "empty_overrides_default", schema: `{"type":"string","default":"fallback"}`, configPresent: true, wire: "page_size="},
			{name: "default_false", schema: `{"type":"boolean","default":false}`, wire: "page_size=false"},
			{name: "default_bigint", schema: `{"type":"integer","default":9007199254740993}`, wire: "page_size=9007199254740993"},
			{name: "default_decimal", schema: `{"type":"number","default":0.10000000000000002}`, wire: "page_size=0.10000000000000002"},
			{name: "integer_min", schema: `{"type":"integer","minimum":1,"maximum":5}`, config: "1", configPresent: true, wire: "page_size=1"},
			{name: "integer_max", schema: `{"type":"integer","minimum":1,"maximum":5}`, config: "5", configPresent: true, wire: "page_size=5"},
			{name: "integer_below", schema: `{"type":"integer","minimum":1,"maximum":5}`, config: "0", configPresent: true, invalid: true},
			{name: "integer_above", schema: `{"type":"integer","minimum":1,"maximum":5}`, config: "6", configPresent: true, invalid: true},
			{name: "bigint_exact", schema: `{"type":"integer","minimum":9007199254740993,"maximum":9007199254740995}`, config: "9007199254740993", configPresent: true, wire: "page_size=9007199254740993"},
			{name: "bigint_below", schema: `{"type":"integer","minimum":9007199254740993,"maximum":9007199254740995}`, config: "9007199254740992", configPresent: true, invalid: true},
			{name: "decimal_exact", schema: `{"type":"number","minimum":0.10000000000000001,"maximum":0.10000000000000003}`, config: "0.10000000000000002", configPresent: true, wire: "page_size=0.10000000000000002"},
			{name: "decimal_above", schema: `{"type":"number","minimum":0.10000000000000001,"maximum":0.10000000000000003}`, config: "0.10000000000000004", configPresent: true, invalid: true},
			{name: "query_overrides_config", schema: `{"type":"integer","minimum":1,"maximum":5}`, config: "2", configPresent: true, override: "5", overridePresent: true, wire: "page_size=5"},
			{name: "invalid_query_overrides_config", schema: `{"type":"integer","minimum":1,"maximum":5}`, config: "2", configPresent: true, override: "6", overridePresent: true, invalid: true},
			{name: "explicit_empty_override", schema: `{"type":"string","default":"fallback"}`, config: "present", configPresent: true, overridePresent: true, wire: "page_size="},
			{name: "numeric_empty", schema: `{"type":"integer"}`, configPresent: true, invalid: true},
			{name: "truthy_refused", schema: `{"type":"boolean"}`, config: "1", configPresent: true, invalid: true},
			{name: "fraction_refused", schema: `{"type":"number"}`, config: "1/10", configPresent: true, invalid: true},
			{name: "escaped_string", schema: `{"type":"string"}`, config: "a+b c", configPresent: true, wire: "page_size=a%2Bb+c"},
			{name: "pattern_refused", schema: `{"type":"string","pattern":"^[a-z]+$"}`, config: "123", configPresent: true, invalid: true},
			{name: "enum_refused", schema: `{"type":"string","enum":["acme","other"]}`, config: "wrong", configPresent: true, invalid: true},
			{name: "date_refused", schema: `{"type":"string","format":"date"}`, config: "not-a-date", configPresent: true, invalid: true},
			{name: "date_healthy", schema: `{"type":"string","format":"date"}`, config: "2026-09-09", configPresent: true, wire: "page_size=2026-09-09"},
		} {
			t.Run(adapter+"/"+tc.name, func(t *testing.T) {
				var sends atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					sends.Add(1)
					if r.Method != "GET" || r.URL.Path != "/widgets" || r.URL.RawQuery != tc.wire {
						t.Errorf("wire %s %s; want GET /widgets?%s", r.Method, r.URL.RequestURI(), tc.wire)
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"data":[{"id":101,"updated_at":"2026-01-01"},{"id":202,"updated_at":"2026-01-02"},{"id":303,"updated_at":"2026-01-03"}]}`))
				}))
				defer server.Close()
				fsys := requestInputBundle234(t)
				var doc map[string]any
				decode := json.NewDecoder(strings.NewReader(string(fsys["acme/schemas/inputs-widgets.json"].Data)))
				decode.UseNumber()
				if err := decode.Decode(&doc); err != nil {
					t.Fatal(err)
				}
				query := doc["properties"].(map[string]any)["query"].(map[string]any)
				query["properties"].(map[string]any)["page_size"] = json.RawMessage(tc.schema)
				raw, err := json.Marshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				fsys["acme/schemas/inputs-widgets.json"].Data = raw
				doc = nil
				if err := json.Unmarshal(fsys["acme/streams.json"].Data, &doc); err != nil {
					t.Fatal(err)
				}
				doc["base"].(map[string]any)["auth"] = []any{}
				raw, err = json.Marshal(doc)
				if err != nil {
					t.Fatal(err)
				}
				fsys["acme/streams.json"].Data = raw
				if adapter == "direct" {
					var scalar struct {
						Type string `json:"type"`
					}
					if err := json.Unmarshal([]byte(tc.schema), &scalar); err != nil {
						t.Fatal(err)
					}
					contract := doc["streams"].([]any)[0].(map[string]any)["request_inputs"]
					parameter := map[string]any{"name": "page_size", "in": "query", "type": scalar.Type}
					var scalarFields map[string]json.RawMessage
					if err := json.Unmarshal([]byte(tc.schema), &scalarFields); err != nil {
						t.Fatal(err)
					}
					for _, field := range []string{"minimum", "maximum"} {
						if value, exists := scalarFields[field]; exists {
							parameter[field] = value
						}
					}
					if value, exists := scalarFields["enum"]; exists {
						parameter["values"] = value
					}
					operation := map[string]any{"id": "widgets", "kind": "rest_read", "summary": "widgets", "risk": "low", "approval": "none", "output_policy": "json_redacted", "rest": map[string]any{"method": "GET", "path": "/widgets", "max_bytes": 2048, "parameters": []any{parameter}, "request_inputs": contract, "response": map[string]any{"success_statuses": []string{"200"}}}}
					raw, err = json.Marshal(map[string]any{"operations": []any{operation}})
					if err != nil {
						t.Fatal(err)
					}
					fsys["acme/operations.json"] = &fstest.MapFile{Data: raw}
				}
				bundle, err := Load(fsys, "acme")
				if err != nil {
					t.Fatal(err)
				}
				config := map[string]string{"base_url": server.URL, "unrelated": "not-an-integer"}
				if tc.configPresent {
					config["page_size"] = tc.config
				}
				req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: config}}
				if tc.overridePresent {
					req.Query = map[string]string{"page_size": tc.override}
				}
				var ids []string
				emit := func(record connectors.Record) error { ids = append(ids, fmt.Sprint(record["id"])); return nil }
				switch adapter {
				case "direct":
					query := map[string]string{}
					if tc.configPresent {
						query["page_size"] = tc.config
					}
					if tc.overridePresent {
						query["page_size"] = tc.override
					}
					result, directErr := OperationDirectRead(t.Context(), bundle, connectors.OperationDirectReadRequest{Operation: "widgets", Config: req.Config, Query: query}, nil)
					err = directErr
					if directErr == nil {
						var decoded struct {
							Data []struct {
								ID int `json:"id"`
							} `json:"data"`
						}
						raw, marshalErr := json.Marshal(result.Body)
						if marshalErr != nil {
							t.Fatal(marshalErr)
						}
						if decodeErr := json.Unmarshal(raw, &decoded); decodeErr != nil {
							t.Fatal(decodeErr)
						}
						for _, row := range decoded.Data {
							ids = append(ids, fmt.Sprint(row.ID))
						}
					}
				case "exported":
					err = Read(t.Context(), bundle, req, nil, emit)
				default:
					err = (&Connector{bundle: bundle}).Read(t.Context(), req, emit)
				}
				if tc.invalid {
					if err == nil || sends.Load() != 0 || len(ids) != 0 {
						t.Fatalf("invalid input err=%v sends=%d ids=%v", err, sends.Load(), ids)
					}
					return
				}
				if err != nil || sends.Load() != 1 || !reflect.DeepEqual(ids, []string{"101", "202", "303"}) {
					t.Fatalf("healthy input err=%v sends=%d ids=%v", err, sends.Load(), ids)
				}
			})
		}
	}
}

func TestRequestInputScopeMatrix248(t *testing.T) {
	for _, tc := range []struct {
		name, stream, size, extra, wire string
		valid                           bool
	}{
		{"first_ignores_other_required", "widgets", "3", "", "page_size=3", true},
		{"second_rejects_first_range", "other", "3", "present", "", false},
		{"second_requires_own_member", "other", "12", "", "", false},
		{"second_healthy", "other", "12", "present", "only_b=present&page_size=12", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var sends atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sends.Add(1)
				if r.URL.Path != "/"+tc.stream || r.URL.RawQuery != tc.wire {
					t.Errorf("unexpected selected wire %s", r.URL.RequestURI())
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":[{"id":101,"updated_at":"2026-01-01"},{"id":202,"updated_at":"2026-01-02"},{"id":303,"updated_at":"2026-01-03"}]}`))
			}))
			defer server.Close()
			fsys := requestInputBundle234(t)
			var doc map[string]any
			if err := json.Unmarshal(fsys["acme/streams.json"].Data, &doc); err != nil {
				t.Fatal(err)
			}
			doc["base"].(map[string]any)["auth"] = []any{}
			streams := doc["streams"].([]any)
			second := map[string]any{"name": "other", "path": "/other", "schema": "schemas/widgets.json", "records": map[string]any{"path": "data"}, "query": map[string]any{"page_size": map[string]any{"template": "{{ config.page_size }}", "omit_when_absent": true}, "only_b": map[string]any{"template": "{{ config.only_b }}"}}, "request_inputs": map[string]any{"version": 1, "schema": "schemas/inputs-other.json", "bindings": []any{map[string]any{"config_key": "page_size", "in": "query", "name": "page_size"}, map[string]any{"config_key": "only_b", "in": "query", "name": "only_b"}}}}
			doc["streams"] = append(streams, second)
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/streams.json"].Data = raw
			fsys["acme/fixtures/streams/other/page_1.json"] = &fstest.MapFile{Data: []byte(`{"request":{"method":"GET","path":"/other","query":{}},"response":{"status":200,"body":{"data":[]}}}`)}
			fsys["acme/schemas/inputs-other.json"] = &fstest.MapFile{Data: []byte(`{"type":"object","properties":{"path":{"type":"object","properties":{},"additionalProperties":false},"query":{"type":"object","properties":{"page_size":{"type":"integer","minimum":10,"maximum":20},"only_b":{"type":"string"}},"required":["only_b"],"additionalProperties":false},"header":{"type":"object","properties":{},"additionalProperties":false}},"required":["path","query","header"],"additionalProperties":false}`)}
			bundle, err := Load(fsys, "acme")
			if err != nil {
				t.Fatal(err)
			}
			config := map[string]string{"base_url": server.URL, "page_size": tc.size}
			if tc.extra != "" {
				config["only_b"] = tc.extra
			}
			rows, err := readAll(t, t.Context(), bundle, connectors.ReadRequest{Stream: tc.stream, Config: connectors.RuntimeConfig{Config: config}}, nil)
			if !tc.valid {
				if err == nil || sends.Load() != 0 || len(rows) != 0 {
					t.Fatalf("invalid selected input err=%v sends=%d rows=%v", err, sends.Load(), rows)
				}
				return
			}
			var ids []string
			for _, row := range rows {
				ids = append(ids, fmt.Sprint(row["id"]))
			}
			if err != nil || sends.Load() != 1 || !reflect.DeepEqual(ids, []string{"101", "202", "303"}) {
				t.Fatalf("selected input err=%v sends=%d ids=%v", err, sends.Load(), ids)
			}
		})
	}
}

func TestRequestInputDefaultAdmissionMatrix248(t *testing.T) {
	for _, tc := range []struct {
		name, schema string
		valid        bool
	}{
		{"integer_boundary", `{"type":"integer","minimum":2,"maximum":5,"default":2}`, true},
		{"integer_below", `{"type":"integer","minimum":2,"maximum":5,"default":1}`, false},
		{"integer_above", `{"type":"integer","minimum":2,"maximum":5,"default":6}`, false},
		{"enum_member", `{"type":"string","enum":["acme"],"default":"acme"}`, true},
		{"enum_absent", `{"type":"string","enum":["acme"],"default":"other"}`, false},
		{"pattern", `{"type":"string","pattern":"^[a-z]+$","default":"123"}`, false},
		{"type", `{"type":"boolean","default":"false"}`, false},
		{"date", `{"type":"string","format":"date","default":"not-a-date"}`, false},
		{"decimal_exact", `{"type":"number","minimum":0.10000000000000001,"maximum":0.10000000000000003,"default":0.10000000000000002}`, true},
		{"decimal_outside", `{"type":"number","minimum":0.10000000000000001,"maximum":0.10000000000000003,"default":0.10000000000000004}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fsys := requestInputBundle234(t)
			var doc map[string]any
			if err := json.Unmarshal(fsys["acme/schemas/inputs-widgets.json"].Data, &doc); err != nil {
				t.Fatal(err)
			}
			doc["properties"].(map[string]any)["query"].(map[string]any)["properties"].(map[string]any)["page_size"] = json.RawMessage(tc.schema)
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/schemas/inputs-widgets.json"].Data = raw
			_, err = Load(fsys, "acme")
			if (err == nil) != tc.valid {
				t.Fatalf("default admission valid=%t err=%v", tc.valid, err)
			}
		})
	}
}

// Unbound fixed and templated values are still request inputs. An input
// envelope that omits them cannot certify the request emitted by its stream.
func TestRequestInputWireClosureMatrix248(t *testing.T) {
	for _, name := range []string{"healthy", "extra_query", "extra_header", "wrong_path_alias", "header_case_collision", "query_default_conflict"} {
		t.Run(name, func(t *testing.T) {
			fsys := requestInputBundle234(t)
			var doc map[string]any
			if err := json.Unmarshal(fsys["acme/streams.json"].Data, &doc); err != nil {
				t.Fatal(err)
			}
			stream := doc["streams"].([]any)[0].(map[string]any)
			switch name {
			case "extra_query":
				stream["query"].(map[string]any)["unbound"] = map[string]any{"template": "{{ config.unbound }}"}
			case "extra_header":
				stream["headers"] = map[string]any{"X-Unbound": "{{ config.unbound }}"}
			case "wrong_path_alias":
				stream["path"] = "/widgets/{{ config.unbound }}"
			case "header_case_collision":
				stream["headers"] = map[string]any{"X-Unbound": "a", "x-unbound": "b"}
			case "query_default_conflict":
				stream["query"].(map[string]any)["page_size"].(map[string]any)["default"] = "999"
			}
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/streams.json"].Data = raw
			_, err = Load(fsys, "acme")
			if name == "healthy" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Load admitted request input discrepancy %s", name)
			}
		})
	}
}

func TestRequestInputAliasMatrix248(t *testing.T) {
	for _, name := range []string{"healthy", "empty_path", "missing_path", "invalid_header", "unknown_query", "cold_reload"} {
		t.Run(name, func(t *testing.T) {
			var sends atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sends.Add(1)
				if r.Method != "GET" || r.URL.EscapedPath() != "/widgets/acme" || r.URL.RawQuery != "provider_flag=false" || r.Header.Get("X-Count") != "9007199254740993" {
					t.Errorf("alias wire %s %s header=%s", r.Method, r.URL.RequestURI(), r.Header.Get("X-Count"))
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":[{"id":101,"updated_at":"2026-01-01"},{"id":202,"updated_at":"2026-01-02"},{"id":303,"updated_at":"2026-01-03"}]}`))
			}))
			defer server.Close()
			fsys := fullValidBundleFS("acme")
			var doc map[string]any
			if err := json.Unmarshal(fsys["acme/streams.json"].Data, &doc); err != nil {
				t.Fatal(err)
			}
			doc["base"].(map[string]any)["auth"] = []any{}
			stream := doc["streams"].([]any)[0].(map[string]any)
			stream["path"] = "/widgets/{{ config.local_path }}"
			stream["query"] = map[string]any{"provider_flag": map[string]any{"template": "{{ config.local_flag }}", "omit_when_absent": true}}
			stream["headers"] = map[string]any{"X-Count": "{{ config.local_count }}"}
			stream["request_inputs"] = map[string]any{"version": 1, "schema": "schemas/alias-inputs.json", "bindings": []any{map[string]any{"config_key": "local_path", "in": "path", "name": "provider_id"}, map[string]any{"config_key": "local_flag", "in": "query", "name": "provider_flag"}, map[string]any{"config_key": "local_count", "in": "header", "name": "X-Count"}}}
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/streams.json"].Data = raw
			fsys["acme/schemas/alias-inputs.json"] = &fstest.MapFile{Data: []byte(`{"type":"object","additionalProperties":false,"required":["path","query","header"],"properties":{"path":{"type":"object","additionalProperties":false,"required":["provider_id"],"properties":{"provider_id":{"type":"string"}}},"query":{"type":"object","additionalProperties":false,"properties":{"provider_flag":{"type":"boolean","default":true}}},"header":{"type":"object","additionalProperties":false,"required":["X-Count"],"properties":{"X-Count":{"type":"integer","minimum":9007199254740993,"maximum":9007199254740995}}}}}`)}
			bundle, err := Load(fsys, "acme")
			if err != nil {
				t.Fatal(err)
			}
			if name == "cold_reload" {
				detached := fstest.MapFS{}
				for path, file := range fsys {
					detached[path] = &fstest.MapFile{Data: append([]byte(nil), file.Data...)}
				}
				bundle, err = Load(detached, "acme")
				if err != nil {
					t.Fatal(err)
				}
			}
			config := map[string]string{"base_url": server.URL, "local_path": "acme", "local_flag": "false", "local_count": "9007199254740993"}
			switch name {
			case "empty_path":
				config["local_path"] = ""
			case "missing_path":
				delete(config, "local_path")
			case "invalid_header":
				config["local_count"] = "9007199254740992"
			}
			req := connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: config}}
			if name == "unknown_query" {
				req.Query = map[string]string{"unknown": "value"}
			}
			before := map[string]string{}
			for key, value := range config {
				before[key] = value
			}
			rows, err := readAll(t, t.Context(), bundle, req, nil)
			if !reflect.DeepEqual(config, before) {
				t.Fatal("caller configuration mutated")
			}
			if name != "healthy" && name != "cold_reload" {
				if err == nil || sends.Load() != 0 || len(rows) != 0 {
					t.Fatalf("alias refusal err=%v sends=%d rows=%v", err, sends.Load(), rows)
				}
				return
			}
			var ids []string
			for _, row := range rows {
				ids = append(ids, fmt.Sprint(row["id"]))
			}
			if err != nil || sends.Load() != 1 || !reflect.DeepEqual(ids, []string{"101", "202", "303"}) {
				t.Fatalf("alias execution err=%v sends=%d ids=%v", err, sends.Load(), ids)
			}
		})
	}
}

func TestRequestInputResourceMatrix248(t *testing.T) {
	for _, tc := range []struct {
		name, kind, value string
		valid             bool
	}{
		{"healthy", "string", "safe", true},
		{"value_over_budget", "string", strings.Repeat("a", (1<<20)+1), false},
		{"exponent_over_budget", "number", "1e1048577", false},
		{"invalid_utf8", "string", string([]byte{0xff}), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fsys := requestInputBundle234(t)
			var doc map[string]any
			if err := json.Unmarshal(fsys["acme/schemas/inputs-widgets.json"].Data, &doc); err != nil {
				t.Fatal(err)
			}
			doc["properties"].(map[string]any)["query"].(map[string]any)["properties"].(map[string]any)["page_size"] = map[string]any{"type": tc.kind}
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/schemas/inputs-widgets.json"].Data = raw
			bundle, err := Load(fsys, "acme")
			if err != nil {
				t.Fatal(err)
			}
			err = New(bundle, nil).ValidateReadInputs(t.Context(), connectors.ReadInputValidationRequest{Stream: "widgets", Config: map[string]string{"page_size": tc.value}})
			if (err == nil) != tc.valid {
				t.Fatalf("resource admission valid=%t err=%v", tc.valid, err)
			}
		})
	}
}

func TestRequestInputFlagCorrespondenceMatrix248(t *testing.T) {
	for _, name := range []string{"healthy_enum", "changed_enum", "missing_enum", "wrong_format", "wrong_required", "wrong_repeatability"} {
		t.Run(name, func(t *testing.T) {
			fsys := requestInputBundle234(t)
			var doc map[string]any
			if err := json.Unmarshal(fsys["acme/schemas/inputs-widgets.json"].Data, &doc); err != nil {
				t.Fatal(err)
			}
			doc["properties"].(map[string]any)["query"].(map[string]any)["properties"].(map[string]any)["page_size"] = map[string]any{"type": "string", "enum": []string{"acme", "other"}, "default": "acme"}
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/schemas/inputs-widgets.json"].Data = raw
			flag := map[string]any{"name": "page-size", "type": "string", "maps_to": "config.page_size", "input_codec": "source_scalar_v1", "values": []string{"acme", "other"}}
			switch name {
			case "changed_enum":
				flag["values"] = []string{"acme", "wrong"}
			case "missing_enum":
				delete(flag, "values")
			case "wrong_format":
				flag["format"] = "date"
			case "wrong_required":
				flag["required"] = true
			case "wrong_repeatability":
				flag["repeatable"] = true
			}
			raw, err = json.Marshal(map[string]any{"tagline": "widgets", "usage": "pm acme widgets", "commands": []any{map[string]any{"path": "widgets", "summary": "widgets", "intent": "etl", "availability": "implemented", "stream": "widgets", "flags": []any{flag}}}})
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/cli_surface.json"] = &fstest.MapFile{Data: raw}
			_, err = Load(fsys, "acme")
			if name == "healthy_enum" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatalf("admitted same-source metadata discrepancy %s", name)
			}
		})
	}
}

func TestRequestInputEffectivePagingMatrix248(t *testing.T) {
	for _, name := range []string{"healthy", "invalid_initial_size", "invalid_later_offset"} {
		t.Run(name, func(t *testing.T) {
			var sends atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := sends.Add(1)
				want := "offset=0&page_size=2"
				response := `{"data":[{"id":101,"updated_at":"2026-01-01"},{"id":202,"updated_at":"2026-01-02"}]}`
				if n == 2 {
					want = "offset=2&page_size=2"
					response = `{"data":[{"id":303,"updated_at":"2026-01-03"}]}`
				}
				if n > 2 || r.URL.RawQuery != want {
					t.Errorf("paging request %d query=%s want=%s", n, r.URL.RawQuery, want)
					w.WriteHeader(400)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(response))
			}))
			defer server.Close()
			fsys := requestInputBundle234(t)
			var doc map[string]any
			if err := json.Unmarshal(fsys["acme/streams.json"].Data, &doc); err != nil {
				t.Fatal(err)
			}
			doc["base"].(map[string]any)["auth"] = []any{}
			stream := doc["streams"].([]any)[0].(map[string]any)
			stream["pagination"] = map[string]any{"type": "offset_limit", "offset_param": "offset", "limit_param": "page_size", "page_size": 2}
			stream["query"].(map[string]any)["offset"] = map[string]any{"template": "{{ config.offset }}", "omit_when_absent": true}
			stream["request_inputs"].(map[string]any)["bindings"] = []any{map[string]any{"in": "query", "name": "page_size", "config_key": "page_size"}, map[string]any{"in": "query", "name": "offset", "config_key": "offset"}}
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/streams.json"].Data = raw
			doc = nil
			if err := json.Unmarshal(fsys["acme/schemas/inputs-widgets.json"].Data, &doc); err != nil {
				t.Fatal(err)
			}
			maximum := 2
			if name == "invalid_later_offset" {
				maximum = 1
			}
			props := doc["properties"].(map[string]any)["query"].(map[string]any)["properties"].(map[string]any)
			props["offset"] = map[string]any{"type": "integer", "minimum": 0, "maximum": maximum}
			if name == "invalid_initial_size" {
				props["page_size"] = map[string]any{"type": "integer", "maximum": 1}
			}
			raw, err = json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/schemas/inputs-widgets.json"].Data = raw
			bundle, err := Load(fsys, "acme")
			if err != nil {
				t.Fatal(err)
			}
			var ids []string
			err = Read(t.Context(), bundle, connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL}}}, nil, func(row connectors.Record) error { ids = append(ids, fmt.Sprint(row["id"])); return nil })
			switch name {
			case "invalid_initial_size":
				if err == nil || sends.Load() != 0 || len(ids) != 0 {
					t.Fatalf("initial invalid err=%v sends=%d ids=%v", err, sends.Load(), ids)
				}
			case "invalid_later_offset":
				if err == nil || sends.Load() != 1 || !reflect.DeepEqual(ids, []string{"101", "202"}) {
					t.Fatalf("later invalid err=%v sends=%d priorIDs=%v", err, sends.Load(), ids)
				}
			default:
				if err != nil || sends.Load() != 2 || !reflect.DeepEqual(ids, []string{"101", "202", "303"}) {
					t.Fatalf("effective paging err=%v sends=%d ids=%v", err, sends.Load(), ids)
				}
			}
		})
	}
}

func TestRequestInputAggregateBudget248(t *testing.T) {
	for _, count := range []int{16, 17} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			fsys := requestInputBundle234(t)
			var streamDoc map[string]any
			if err := json.Unmarshal(fsys["acme/streams.json"].Data, &streamDoc); err != nil {
				t.Fatal(err)
			}
			stream := streamDoc["streams"].([]any)[0].(map[string]any)
			query := map[string]any{}
			properties := map[string]any{}
			bindings := []any{}
			config := map[string]string{}
			for i := 0; i < count; i++ {
				name := fmt.Sprintf("value%d", i)
				query[name] = map[string]any{"template": "{{ config." + name + " }}", "omit_when_absent": true}
				properties[name] = map[string]any{"type": "string"}
				bindings = append(bindings, map[string]any{"config_key": name, "in": "query", "name": name})
				config[name] = strings.Repeat("a", 1<<20)
			}
			stream["query"] = query
			stream["request_inputs"].(map[string]any)["bindings"] = bindings
			raw, err := json.Marshal(streamDoc)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/streams.json"].Data = raw
			var schema map[string]any
			if err := json.Unmarshal(fsys["acme/schemas/inputs-widgets.json"].Data, &schema); err != nil {
				t.Fatal(err)
			}
			schema["properties"].(map[string]any)["query"].(map[string]any)["properties"] = properties
			raw, err = json.Marshal(schema)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/schemas/inputs-widgets.json"].Data = raw
			bundle, err := Load(fsys, "acme")
			if err != nil {
				t.Fatal(err)
			}
			err = New(bundle, nil).ValidateReadInputs(t.Context(), connectors.ReadInputValidationRequest{Stream: "widgets", Config: config})
			if (err == nil) != (count == 16) {
				t.Fatalf("aggregate scalar budget count=%d err=%v", count, err)
			}
		})
	}
}

func TestRequestInputHeaderCustody248(t *testing.T) {
	for _, name := range []string{"healthy", "protected", "base_collision", "canonical_collision", "line_break", "unbound"} {
		t.Run(name, func(t *testing.T) {
			fsys := requestInputBundle234(t)
			var doc map[string]any
			if err := json.Unmarshal(fsys["acme/streams.json"].Data, &doc); err != nil {
				t.Fatal(err)
			}
			stream := doc["streams"].([]any)[0].(map[string]any)
			header := "X-Scope"
			if name == "protected" {
				header = "Authorization"
			}
			stream["headers"] = map[string]any{header: "{{ config.scope }}"}
			contract := stream["request_inputs"].(map[string]any)
			bindings := contract["bindings"].([]any)
			if name != "unbound" {
				bindings = append(bindings, map[string]any{"in": "header", "name": header, "config_key": "scope"})
			}
			if name == "canonical_collision" {
				stream["headers"].(map[string]any)["X_Scope"] = "{{ config.other }}"
				bindings = append(bindings, map[string]any{"in": "header", "name": "X_Scope", "config_key": "other"})
			}
			contract["bindings"] = bindings
			if name == "base_collision" {
				doc["base"].(map[string]any)["headers"] = map[string]any{"X-Scope": "runtime-owned"}
			}
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/streams.json"].Data = raw
			doc = nil
			if err := json.Unmarshal(fsys["acme/schemas/inputs-widgets.json"].Data, &doc); err != nil {
				t.Fatal(err)
			}
			props := doc["properties"].(map[string]any)["header"].(map[string]any)["properties"].(map[string]any)
			if name != "unbound" {
				props[header] = map[string]any{"type": "string", "maxLength": 64}
			}
			if name == "canonical_collision" {
				props["X_Scope"] = map[string]any{"type": "string", "maxLength": 64}
			}
			raw, err = json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/schemas/inputs-widgets.json"].Data = raw
			bundle, err := Load(fsys, "acme")
			if name != "healthy" && name != "line_break" {
				if err == nil {
					t.Fatal("unsafe header admitted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			value := "acme"
			if name == "line_break" {
				value = "acme\r\nX-Injected: yes"
			}
			err = New(bundle, nil).ValidateReadInputs(t.Context(), connectors.ReadInputValidationRequest{Stream: "widgets", Config: map[string]string{"scope": value}})
			if (err == nil) != (name == "healthy") {
				t.Fatalf("header value frontier err=%v", err)
			}
		})
	}
}
