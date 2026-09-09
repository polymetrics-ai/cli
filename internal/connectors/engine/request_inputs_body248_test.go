package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
)

// Every healthy case enters through Load: a private partial plan cannot stand
// in for the generated input contract or its complete effective body.
func TestRequestInputLoadedBodyMatrix248(t *testing.T) {
	for _, name := range []string{"literal", "template", "named", "named_invalid", "required_missing", "extra_member", "engine_required_size", "later_invalid_cursor", "missing_plan", "corrupt_plan"} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if failure := recover(); failure != nil {
					t.Errorf("public Read panicked: %v", failure)
				}
			}()
			var sends atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := sends.Add(1)
				raw, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				want := `{"maxResults":2,"scope":"acme"}`
				response := `{"data":[{"id":101,"updated_at":"2026-01-01"},{"id":202,"updated_at":"2026-01-02"}],"nextPageToken":"next"}`
				if name == "later_invalid_cursor" {
					response = `{"data":[{"id":101,"updated_at":"2026-01-01"},{"id":202,"updated_at":"2026-01-02"}],"nextPageToken":"way-too-long"}`
				}
				if n == 2 {
					want = `{"maxResults":2,"nextPageToken":"next","scope":"acme"}`
					response = `{"data":[{"id":303,"updated_at":"2026-01-03"}],"nextPageToken":null}`
				}
				if n > 2 || r.Method != "POST" || r.URL.RequestURI() != "/widgets" || string(raw) != want {
					t.Errorf("unexpected body request %d %s %s %s", n, r.Method, r.URL.RequestURI(), raw)
					w.WriteHeader(400)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, response)
			}))
			defer server.Close()
			fsys := fullValidBundleFS("acme")
			var doc map[string]any
			if err := json.Unmarshal(fsys["acme/streams.json"].Data, &doc); err != nil {
				t.Fatal(err)
			}
			doc["base"].(map[string]any)["auth"] = []any{}
			stream := doc["streams"].([]any)[0].(map[string]any)
			stream["method"], stream["body_type"] = "POST", "json"
			delete(stream, "query")
			body := map[string]any{"scope": "acme"}
			bindings := []any{}
			switch name {
			case "template":
				body["scope"] = "{{ config.scope }}"
			case "named", "named_invalid":
				delete(body, "scope")
				bindings = append(bindings, map[string]any{"config_key": "scope", "in": "body", "pointer": "/scope"})
			case "required_missing":
				delete(body, "scope")
			case "extra_member":
				body["extra"] = "unreviewed"
			}
			stream["body"] = body
			stream["pagination"] = map[string]any{"type": "cursor", "cursor_param": "nextPageToken", "token_path": "nextPageToken", "body_cursor_field": "nextPageToken", "size_param": "maxResults", "body_limit_field": "maxResults", "page_size": 2}
			stream["request_inputs"] = map[string]any{"version": 1, "schema": "schemas/inputs-body.json", "bindings": bindings}
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			fsys["acme/streams.json"].Data = raw
			required := `"scope"`
			if name == "engine_required_size" {
				required += `,"maxResults"`
			}
			schema := `{"type":"object","additionalProperties":false,"required":["path","query","header","body"],"properties":{"path":{"type":"object","properties":{},"additionalProperties":false},"query":{"type":"object","properties":{},"additionalProperties":false},"header":{"type":"object","properties":{},"additionalProperties":false},"body":{"type":"object","additionalProperties":false,"required":[` + required + `],"properties":{"scope":{"type":"string","enum":["acme"]},"maxResults":{"type":"integer","minimum":1,"maximum":5,"default":2},"nextPageToken":{"type":"string","maxLength":4}}}}}`
			fsys["acme/schemas/inputs-body.json"] = &fstest.MapFile{Data: []byte(schema)}
			bundle, err := Load(fsys, "acme")
			if err != nil {
				t.Fatal(err)
			}
			if name == "missing_plan" {
				bundle.Streams[0].inputPlan = nil
			}
			if name == "corrupt_plan" {
				bundle.Streams[0].inputPlan = &compiledRequestInputPlan{}
			}
			config := map[string]string{"base_url": server.URL, "scope": "acme"}
			if name == "named_invalid" {
				config["scope"] = "wrong"
			}
			rows, err := readAll(t, t.Context(), bundle, connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: config}}, nil)
			invalid := name == "named_invalid" || name == "required_missing" || name == "extra_member" || name == "missing_plan" || name == "corrupt_plan"
			if invalid {
				if err == nil || sends.Load() != 0 || len(rows) != 0 {
					t.Fatalf("invalid body err=%v sends=%d rows=%v", err, sends.Load(), rows)
				}
				return
			}
			if name == "later_invalid_cursor" {
				if err == nil || sends.Load() != 1 {
					t.Fatalf("invalid continuation err=%v sends=%d rows=%v", err, sends.Load(), rows)
				}
				return
			}
			var ids []string
			for _, row := range rows {
				ids = append(ids, fmt.Sprint(row["id"]))
			}
			if err != nil || sends.Load() != 2 || !reflect.DeepEqual(ids, []string{"101", "202", "303"}) {
				t.Fatalf("loaded body err=%v sends=%d ids=%v", err, sends.Load(), ids)
			}
		})
	}
}
