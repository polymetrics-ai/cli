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
	if variant == "optional_body" {
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
	if len(requests) != 1 || requests[0] != `POST /widgets application/json {"data":{"name":"widget-a"}}` {
		t.Fatalf("actual typed source request=%v", requests)
	}
}
