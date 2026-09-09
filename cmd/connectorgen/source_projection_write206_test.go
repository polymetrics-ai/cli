package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"strings"
	"sync"
	"testing"
)

func TestSourceProjection206TypedMutation(t *testing.T) {
	for _, variant := range []string{"success", "single_attempt", "optional_body"} {
		t.Run(variant, func(t *testing.T) { sourceProjectionMutation206(t, variant) })
	}
}

func TestSourceProjectionBounds262(t *testing.T) {
	for _, variant := range []string{"bounded_integer", "bounded_nullable_integer", "bounded_nullable_null", "bounded_nullable_false", "bounded_ttl", "bounded_zero", "bounded_bigint", "invalid_nullable_null", "invalid_nullable_string", "invalid_nullable_union", "invalid_nullable_object", "invalid_nullable_keyword"} {
		t.Run(variant, func(t *testing.T) { sourceProjectionMutation206(t, variant) })
	}
}

func sourceProjectionMutation206(t *testing.T, variant string) {
	t.Helper()
	var mu sync.Mutex
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 4097))
		if err != nil {
			t.Error(err)
		}
		mu.Lock()
		requests = append(requests, r.Method+" "+r.URL.RequestURI()+" "+r.Header.Get("Content-Type")+" "+string(body))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if variant == "single_attempt" {
			w.WriteHeader(503)
		} else {
			w.WriteHeader(201)
		}
		_, _ = w.Write([]byte(`{"id":"created-1"}`))
	}))
	defer server.Close()
	count := func() int { mu.Lock(); defer mu.Unlock(); return len(requests) }
	schema := json.RawMessage(`{"type":"object","additionalProperties":false,"required":["data"],"properties":{"data":{"type":"object","additionalProperties":false,"required":["name"],"properties":{"name":{"type":"string"}}}}}`)
	bounded := strings.HasPrefix(variant, "bounded_")
	minimum, maximum, below, above := "1", "100", "0", "101"
	switch variant {
	case "bounded_ttl":
		maximum, above = "43200", "43201"
	case "bounded_zero":
		minimum, below = "0", "-1"
	case "bounded_bigint":
		minimum, maximum, below, above = "9007199254740993", "9007199254740995", "9007199254740992", "9007199254740996"
	}
	if bounded {
		schema = json.RawMessage(`{"type":"object","additionalProperties":false,"required":["data"],"properties":{"data":{"type":"object","additionalProperties":false,"required":["name"],"properties":{"name":{"type":"integer","minimum":1,"maximum":100}}}}}`)
		schema = bytes.Replace(schema, []byte(`"minimum":1,"maximum":100`), []byte(`"minimum":`+minimum+`,"maximum":`+maximum), 1)
		if variant == "bounded_ttl" || strings.HasPrefix(variant, "bounded_nullable_") {
			schema = bytes.Replace(schema, []byte(`"type":"integer"`), []byte(`"type":"integer","nullable":true`), 1)
		}
	}
	if variant == "bounded_nullable_false" {
		schema = bytes.Replace(schema, []byte(`"nullable":true`), []byte(`"nullable":false`), 1)
	}
	if strings.HasPrefix(variant, "invalid_nullable_") {
		field := `"type":"integer","nullable":true`
		switch variant {
		case "invalid_nullable_null":
			field = `"type":"integer","nullable":null`
		case "invalid_nullable_string":
			field = `"type":"integer","nullable":"true"`
		case "invalid_nullable_union":
			field = `"type":["integer","null"],"nullable":true`
		case "invalid_nullable_object":
			field = `"type":"object","nullable":true`
		case "invalid_nullable_keyword":
			field += `,"unknown_constraint":1`
		}
		schema = bytes.Replace(schema, []byte(`"type":"string"`), []byte(field), 1)
	}
	operation := map[string]any{"operationId": "create_widget", "requestBody": map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": schema}}}, "responses": map[string]any{"201": map[string]any{"description": "Created"}}}
	if variant == "optional_body" {
		operation["requestBody"].(map[string]any)["required"] = false
	}
	source, err := json.Marshal(map[string]any{"schema_version": 2, "connector": "acme", "rest": map[string]any{"operations": []any{map[string]any{"id": "fixture.create", "method": "POST", "path": "/widgets", "protocol": "rest", "source_operation": operation}}}, "counts": map[string]int{"total": 1}})
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	directory := filepath.Join(root, "acme")
	if err := os.MkdirAll(filepath.Join(directory, "sources"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "sources", "operations.json"), source, 0600); err != nil {
		t.Fatal(err)
	}
	lock := minimalVNextLockForTest()
	lock.Operations = nil
	lock.Schemas = nil
	lock.Lanes["etl"] = "unsupported"
	lock.Lanes["reverse_etl"] = "implemented"
	lock.Metadata = bytes.Replace(lock.Metadata, []byte(`"write":false`), []byte(`"write":true`), 1)
	lock.HTTP = json.RawMessage(strings.Replace(string(lock.HTTP), "https://api.acme.example", server.URL, 1))
	batchable := true
	lock.SourceProjection = &vNextSourceProjection{Version: 1, Inventories: []vNextSourceProjectionInventory{{ID: "primary", Path: "sources/operations.json", SHA256: sourceBytesHash(source), Bytes: int64(len(source))}}, Semantics: []vNextSourceProjectionSemantic{{Source: vNextSourceProjectionKey{Inventory: "primary", ID: "fixture.create"}, Effect: "mutation", Write: &vNextSourceProjectionWrite{RowDelivery: "one_request", Batchable: &batchable, Risk: "medium", Retry: "single_attempt"}}}}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "source.lock.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := runLockRender([]string{"lock-render", "acme", "--defs", root}, &stdout, &stderr)
	if variant == "optional_body" || strings.HasPrefix(variant, "invalid_nullable_") {
		if code == 0 {
			t.Fatal("optional-body source arm silently restricted to required record body")
		}
		if count() != 0 {
			t.Fatal("unaccounted source variant sent request")
		}
		if _, err := os.Lstat(filepath.Join(directory, vNextPublicationCurrentFile)); !os.IsNotExist(err) {
			t.Fatalf("unaccounted source variant changed CURRENT: %v", err)
		}
		return
	}
	if code != 0 {
		t.Fatalf("actual source mutation admission: code=%d %s", code, stderr.String())
	}
	publisher, err := newVNextGenerationPublisher(root, "acme", vNextPublicationHooks{})
	if err != nil {
		t.Fatal(err)
	}
	handle, err := publisher.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Release()
	files := map[string][]byte{}
	for _, name := range handle.Files() {
		switch name {
		case "manifest.json", "provenance.json", "atlas.json", "index.json", "proof.json":
			continue
		}
		files[name], err = handle.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
	}
	bundle, err := engine.Load(newVNextExecutionFS("acme", files), "acme")
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.New(bundle, nil).PreflightSavedWriteAction("create_widget"); err != nil {
		t.Fatal(err)
	}
	req := connectors.WriteRequest{Action: "create_widget"}
	bad := []connectors.Record{{"data": map[string]any{}}}
	if err := engine.ValidateWrite(t.Context(), bundle, req, bad); err == nil {
		t.Fatal("missing source-required name accepted")
	}
	records := []connectors.Record{{"data": map[string]any{"name": "widget-a"}}}
	expectedBody := `POST /widgets application/json {"data":{"name":"widget-a"}}`
	if bounded {
		records = []connectors.Record{{"data": map[string]any{"name": json.Number(minimum)}}}
		expectedBody = `POST /widgets application/json {"data":{"name":` + minimum + `}}`
		if variant == "bounded_nullable_null" {
			records[0]["data"].(map[string]any)["name"] = nil
			expectedBody = `POST /widgets application/json {"data":{"name":null}}`
		}
		for _, value := range []any{json.Number(minimum), json.Number(maximum)} {
			healthy := []connectors.Record{{"data": map[string]any{"name": value}}}
			if _, err := engine.DryRunWrite(t.Context(), bundle, req, healthy, nil); err != nil {
				t.Fatalf("source endpoint %v refused: %v", value, err)
			}
		}
		for _, value := range []any{json.Number(below), json.Number(above), json.Number("1.5"), "1"} {
			invalid := []connectors.Record{{"data": map[string]any{"name": value}}}
			if err := engine.ValidateWrite(t.Context(), bundle, req, invalid); err == nil {
				t.Fatalf("source-invalid bound/type %v accepted", value)
			}
			if _, err := engine.DryRunWrite(t.Context(), bundle, req, invalid, nil); err == nil {
				t.Fatalf("source-invalid bound/type %v previewed", value)
			}
			if count() != 0 {
				t.Fatal("invalid source value sent provider request")
			}
		}
		if variant == "bounded_integer" || variant == "bounded_nullable_false" || variant == "bounded_zero" || variant == "bounded_bigint" {
			if err := engine.ValidateWrite(t.Context(), bundle, req, []connectors.Record{{"data": map[string]any{"name": nil}}}); err == nil {
				t.Fatal("nonnullable source accepted null")
			}
		}
	}
	if _, err := engine.DryRunWrite(t.Context(), bundle, req, records, nil); err != nil {
		t.Fatal(err)
	}
	if count() != 0 {
		t.Fatal("admission/invalid record/preview sent provider request")
	}
	result, err := engine.Write(t.Context(), bundle, req, records, nil)
	if variant == "single_attempt" {
		if err == nil || result.RecordsWritten != 0 || count() != 1 {
			t.Fatalf("single-attempt mutation outcome=%+v error=%v sends=%d", result, err, count())
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(result.ProviderResponses) != 1 || result.ProviderResponses[0].Status != 201 || !result.ProviderResponses[0].BodyPresent || result.ProviderResponses[0].BodyRaw != `{"id":"created-1"}` {
		t.Fatalf("actual provider receipt=%+v", result.ProviderResponses)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("actual write outcomes=%+v", result)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 1 || requests[0] != expectedBody {
		t.Fatalf("actual typed source request=%v", requests)
	}
}
