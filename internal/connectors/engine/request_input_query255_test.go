package engine

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
)

func queryFixture255(t *testing.T, mode string, limit func(map[string]any)) fstest.MapFS {
	t.Helper()
	fsys := requestInputBundle234(t)
	contract := map[string]any{"version": 1, "schema": "schemas/inputs-widgets.json", "bindings": []any{map[string]any{"config_key": "filter", "in": "query", "name": "filter"}}}
	encoding := map[string]any{"version": 1, "max_depth": 8, "max_members": 32, "max_items": 32, "max_pairs": 32, "max_bytes": 1024, "fields": map[string]any{"filter": map[string]any{"mode": mode, "empty_array": "omit", "null": "reject"}}}
	if limit != nil {
		limit(encoding)
	}
	contract["query_encoding"] = encoding
	schema := json.RawMessage(`{"type":"array","minItems":1,"items":{"type":"string"}}`)
	if mode == "brackets" {
		schema = json.RawMessage(`{"type":"object","required":["owner"],"properties":{"owner":{"type":"string"}},"additionalProperties":false}`)
	}
	input := map[string]any{"type": "object", "additionalProperties": false, "required": []string{"path", "query", "header"}, "properties": map[string]any{}}
	properties := input["properties"].(map[string]any)
	for _, key := range []string{"path", "query", "header"} {
		properties[key] = map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{}}
	}
	query := properties["query"].(map[string]any)
	query["required"] = []string{"filter"}
	query["properties"] = map[string]any{"filter": schema}
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
	delete(streams["base"].(map[string]any), "pagination")
	stream := streams["streams"].([]any)[0].(map[string]any)
	delete(stream, "pagination")
	stream["request_inputs"] = contract
	stream["query"] = map[string]any{"filter": map[string]any{"template": "{{ config.filter }}", "omit_when_absent": false}}
	raw, err = json.Marshal(streams)
	if err != nil {
		t.Fatal(err)
	}
	fsys["acme/streams.json"].Data = raw
	operation := map[string]any{"id": "widgets", "kind": "rest_read", "summary": "widgets", "risk": "low", "approval": "none", "output_policy": "json_redacted", "rest": map[string]any{"method": "GET", "path": "/widgets", "max_bytes": 2048, "parameters": []any{}, "request_inputs": contract, "response": map[string]any{"success_statuses": []string{"200"}}}}
	raw, err = json.Marshal(map[string]any{"operations": []any{operation}})
	if err != nil {
		t.Fatal(err)
	}
	fsys["acme/operations.json"] = &fstest.MapFile{Data: raw}
	return fsys
}

func queryBundle255(t *testing.T, mode string, limit func(map[string]any)) Bundle {
	t.Helper()
	bundle, err := Load(queryFixture255(t, mode, limit), "acme")
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func TestTypedQueryConsumer255(t *testing.T) {
	for _, kind := range []string{"repeated", "brackets"} {
		for _, variant := range []string{"healthy", "unknown", "wrong_type", "conflicting_channels", "byte_budget", "pair_budget", "item_budget", "member_budget", "compiled_identity"} {
			label := variant
			if kind == "brackets" && variant == "pair_budget" {
				label = "undeclared_map_member"
			}
			if kind == "brackets" && variant == "item_budget" {
				label = "nested_wrong_type"
			}
			if kind == "repeated" && variant == "member_budget" {
				label = "nested_non_scalar_item"
			}
			t.Run(kind+"/"+label, func(t *testing.T) {
				var sends atomic.Int32
				expected := "filter=a&filter=b&filter=a"
				var value any = []any{"a", "b", "a"}
				if kind == "brackets" {
					expected = "filter%5Bowner%5D=a%2Bb+c"
					value = map[string]any{"owner": "a+b c"}
				}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					sends.Add(1)
					if r.URL.RawQuery != expected {
						t.Errorf("query=%s want%s", r.URL.RawQuery, expected)
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"data":[{"id":101},{"id":202},{"id":303}]}`))
				}))
				defer server.Close()
				baseline := queryBundle255(t, kind, nil)
				cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL}}
				healthy := connectors.OperationDirectReadRequest{Operation: "widgets", Config: cfg, QueryValues: map[string]any{"filter": value}}
				result, err := OperationDirectRead(t.Context(), baseline, healthy, nil)
				if err != nil || sends.Load() != 1 {
					t.Fatalf("healthy typed consumer: %v sends%d", err, sends.Load())
				}
				body, err := json.Marshal(result.Body)
				if err != nil || string(body) != `{"data":[{"id":101},{"id":202},{"id":303}]}` {
					t.Fatalf("returned rows: %s %v", body, err)
				}
				candidate := baseline
				request := healthy
				request.QueryValues = map[string]any{"filter": value}
				switch variant {
				case "unknown":
					request.QueryValues["other"] = value
				case "wrong_type":
					request.QueryValues["filter"] = "json-looking-value"
				case "conflicting_channels":
					request.Query = map[string]string{"filter": "[]"}
				case "byte_budget":
					candidate = queryBundle255(t, kind, func(e map[string]any) { e["max_bytes"] = 4 })
				case "pair_budget":
					if kind == "repeated" {
						candidate = queryBundle255(t, kind, func(e map[string]any) { e["max_pairs"] = 1 })
					}
					if kind == "brackets" {
						request.QueryValues["filter"] = map[string]any{"owner": "a", "extra": "b"}
					}
				case "item_budget":
					if kind == "repeated" {
						candidate = queryBundle255(t, kind, func(e map[string]any) { e["max_items"] = 1 })
					}
					if kind == "brackets" {
						request.QueryValues["filter"] = map[string]any{"owner": []any{"a", "b"}}
					}
				case "member_budget":
					if kind == "brackets" {
						candidate = queryBundle255(t, kind, func(e map[string]any) { e["max_members"] = 1 })
					}
					if kind == "repeated" {
						request.QueryValues["filter"] = []any{map[string]any{"a": "b", "c": "d"}}
					}
				case "compiled_identity":
					candidate.Operations[0].REST.RequestInputs.QueryEncoding.Fields["filter"] = FormFieldEncoding{Mode: "json"}
				}
				before, _ := json.Marshal(request.QueryValues)
				result, err = OperationDirectRead(t.Context(), candidate, request, nil)
				after, _ := json.Marshal(request.QueryValues)
				if !reflect.DeepEqual(before, after) {
					t.Fatal("caller input mutated")
				}
				wantSuccess := variant == "healthy" || variant == "compiled_identity"
				if wantSuccess {
					if err != nil || sends.Load() != 2 {
						t.Fatalf("stable selected input: %v sends%d", err, sends.Load())
					}
				} else if err == nil || sends.Load() != 1 {
					t.Fatalf("invalid %s error%v sends%d", variant, err, sends.Load())
				}
				if !wantSuccess && strings.HasSuffix(label, "budget") && !strings.Contains(err.Error(), "limit") {
					t.Fatalf("resource case did not reach its limit: %s: %v", label, err)
				}
			})
		}
	}
}

func TestTypedQueryGraphQLRefusal255(t *testing.T) {
	var sends atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sends.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"widget":{"id":"one"}}}`))
	}))
	defer server.Close()
	bundle := graphQLOperationBundle(server.URL, "graphql_query")
	bundle.Operations[0].GraphQL.Pagination = nil
	req := connectors.OperationDirectReadRequest{Operation: bundle.Operations[0].ID, Body: map[string]any{"id": "one"}}
	if _, err := OperationDirectRead(t.Context(), bundle, req, nil); err != nil {
		t.Fatal(err)
	}
	if sends.Load() != 1 {
		t.Fatal("healthy GraphQL did not send")
	}
	req.QueryValues = map[string]any{"filter": []any{"unowned"}}
	if _, err := OperationDirectRead(t.Context(), bundle, req, nil); err == nil || sends.Load() != 1 {
		t.Fatalf("GraphQL ignored structured query: %v sends%d", err, sends.Load())
	}
}

func TestSourceStructuredInputBudget255(t *testing.T) {
	// These values are passed through the actual selected saved read consumer.
	for _, bad := range []string{strings.Repeat("[", 34) + `"a"` + strings.Repeat("]", 34), `["` + strings.Repeat("x", (1<<20)+1) + `"]`} {
		b := queryBundle255(t, "repeated", nil)
		sends := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sends++
			t.Error("invalid structured input reached HTTP")
		}))
		rows := 0
		err := Read(t.Context(), b, connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL, "filter": bad}}}, nil, func(connectors.Record) error { rows++; return nil })
		server.Close()
		if err == nil || sends != 0 || rows != 0 {
			t.Fatalf("structured resource refusal: %v sends%d rows%d", err, sends, rows)
		}
	}
}

func TestTypedQueryContinuationIdentity255(t *testing.T) {
	for _, variant := range []string{"healthy", "changed", "omitted", "reordered", "extra_member", "foreign"} {
		t.Run(variant, func(t *testing.T) {
			var sends, foreignSends atomic.Int32
			foreign := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { foreignSends.Add(1) }))
			defer foreign.Close()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sends.Add(1)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":[{"id":101},{"id":202},{"id":303}]}`))
			}))
			defer server.Close()
			bundle := queryBundle255(t, "repeated", nil)
			bundle.Operations[0].REST.PaginationParameters = []OperationParameter{{Name: "page", In: "query", Type: "integer"}, {Name: "filter", In: "query", Type: "string"}}
			bundle.Operations[0].REST.Pagination = &PaginationSpec{Type: "next_url", NextURLPath: "next", PageParam: "page", NextURLQuery: &NextURLQuerySpec{Allowed: []string{"page", "filter"}}}
			req := connectors.OperationDirectReadRequest{Operation: "widgets", Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL}}, QueryValues: map[string]any{"filter": []any{"a", "b", "a"}}}
			result, err := OperationDirectRead(t.Context(), bundle, req, nil)
			if err != nil || sends.Load() != 1 {
				t.Fatalf("initial consumer: %v sends%d", err, sends.Load())
			}
			raw, _ := json.Marshal(result.Body)
			if string(raw) != `{"data":[{"id":101},{"id":202},{"id":303}]}` {
				t.Fatalf("initial rows: %s", raw)
			}
			suffix := "filter=a&filter=b&filter=a&page=2"
			switch variant {
			case "changed":
				suffix = "filter=other&page=2"
			case "omitted":
				suffix = "page=2"
			case "reordered":
				suffix = "filter=b&filter=a&filter=a&page=2"
			case "extra_member":
				suffix += "&filter%5Bextra%5D=x"
			}
			req.PageCursor = server.URL + "/widgets?" + suffix
			if variant == "foreign" {
				req.PageCursor = foreign.URL + "/widgets?" + suffix
			}
			result, err = OperationDirectRead(t.Context(), bundle, req, nil)
			if variant == "healthy" {
				if err != nil || sends.Load() != 2 {
					t.Fatalf("continuation: %v sends%d", err, sends.Load())
				}
				raw, _ := json.Marshal(result.Body)
				if string(raw) != `{"data":[{"id":101},{"id":202},{"id":303}]}` {
					t.Fatalf("continued rows: %s", raw)
				}
			} else if err == nil || sends.Load() != 1 {
				t.Fatalf("changed structured identity %s: %v sends%d", variant, err, sends.Load())
			}
			if foreignSends.Load() != 0 {
				t.Fatal("foreign request sent")
			}
		})
	}
}

func TestTypedQuerySavedContinuation255(t *testing.T) {
	for _, variant := range []string{"healthy", "changed", "omitted", "reordered"} {
		t.Run(variant, func(t *testing.T) {
			var sends atomic.Int32
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := sends.Add(1)
				w.Header().Set("Content-Type", "application/json")
				if n == 1 {
					suffix := "filter=a&filter=b&filter=a&page=2"
					switch variant {
					case "changed":
						suffix = "filter=x&page=2"
					case "omitted":
						suffix = "page=2"
					case "reordered":
						suffix = "filter=b&filter=a&filter=a&page=2"
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{map[string]any{"id": 101}, map[string]any{"id": 202}, map[string]any{"id": 303}}, "next": server.URL + "/widgets?" + suffix})
				} else {
					_, _ = w.Write([]byte(`{"data":[{"id":404},{"id":505}],"next":null}`))
				}
			}))
			defer server.Close()
			fsys := queryFixture255(t, "repeated", nil)
			var inputs, streams map[string]any
			if err := json.Unmarshal(fsys["acme/schemas/inputs-widgets.json"].Data, &inputs); err != nil {
				t.Fatal(err)
			}
			inputs["properties"].(map[string]any)["query"].(map[string]any)["properties"].(map[string]any)["page"] = map[string]any{"type": "integer", "minimum": 1, "maximum": 2}
			raw, err := json.Marshal(inputs)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/schemas/inputs-widgets.json"].Data = raw
			if err := json.Unmarshal(fsys["acme/streams.json"].Data, &streams); err != nil {
				t.Fatal(err)
			}
			stream := streams["streams"].([]any)[0].(map[string]any)
			contract := stream["request_inputs"].(map[string]any)
			stream["query"].(map[string]any)["page"] = map[string]any{"template": "{{ config.page }}", "omit_when_absent": true}
			contract["bindings"] = append(contract["bindings"].([]any), map[string]any{"config_key": "page", "in": "query", "name": "page"})
			contract["query_encoding"].(map[string]any)["fields"].(map[string]any)["page"] = map[string]any{"mode": "scalar"}
			stream["pagination"] = map[string]any{"type": "next_url", "next_url_path": "next", "next_url_query": map[string]any{"allowed": []string{"filter", "page"}}}
			raw, err = json.Marshal(streams)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/streams.json"].Data = raw
			delete(fsys, "acme/operations.json")
			bundle, err := Load(fsys, "acme")
			if err != nil {
				t.Fatal(err)
			}
			var ids []any
			err = Read(t.Context(), bundle, connectors.ReadRequest{Stream: bundle.Streams[0].Name, Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL, "filter": `["a","b","a"]`}}}, nil, func(row connectors.Record) error { ids = append(ids, row["id"]); return nil })
			raw, _ = json.Marshal(ids)
			if variant == "healthy" {
				if err != nil || sends.Load() != 2 || string(raw) != `[101,202,303,404,505]` {
					t.Fatalf("healthy saved continuation: %v sends%d ids%s", err, sends.Load(), raw)
				}
			} else if err == nil || sends.Load() != 1 || string(raw) != `[101,202,303]` {
				t.Fatalf("changed saved query: %v sends%d ids%s", err, sends.Load(), raw)
			}
		})
	}
}
