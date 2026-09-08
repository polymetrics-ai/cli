package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

func TestSourceProjection206SourceOnlyPublishedRead(t *testing.T) {
	sourceProjectionPublishedRead206(t, "", false)
}

func TestSourceProjection206ArchivedOperationIdentity(t *testing.T) {
	t.Run("envelope_identity", func(t *testing.T) { sourceProjectionPublishedRead206(t, "envelope_identity", false) })
	t.Run("contradictory_identity", func(t *testing.T) { sourceProjectionPublishedRead206(t, "contradictory_identity", true) })
}

func TestSourceProjection206RetainedRawDocument(t *testing.T) {
	sourceProjectionPublishedRead206(t, "raw_schema3", false)
}

func TestSourceProjection206DirectInputs(t *testing.T) {
	for _, variant := range []string{"inherited_path", "optional_query_absent", "optional_query_present"} {
		t.Run(variant, func(t *testing.T) { sourceProjectionPublishedRead206(t, "direct_"+variant, false) })
	}
}

func TestSourceProjection206SourceInputs(t *testing.T) {
	for _, variant := range []string{"inherited_path", "operation_override", "optional_query_absent", "optional_query_present", "missing_path"} {
		t.Run(variant, func(t *testing.T) { sourceProjectionPublishedRead206(t, variant, false) })
	}
}

func TestSourceProjection206UnsupportedSourceSemantics(t *testing.T) {
	for _, variant := range []string{"protocol", "request_body", "response_media", "response_status"} {
		t.Run(variant, func(t *testing.T) { sourceProjectionPublishedRead206(t, variant, true) })
	}
}

func sourceProjectionPublishedRead206(t *testing.T, variant string, wantRefusal bool) {
	t.Helper()
	direct := strings.HasPrefix(variant, "direct_")
	variant = strings.TrimPrefix(variant, "direct_")
	var requests []string
	var requestMu sync.Mutex
	requestSnapshot := func() []string {
		requestMu.Lock()
		defer requestMu.Unlock()
		return append([]string(nil), requests...)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMu.Lock()
		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		requestMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"alpha"},{"id":"bravo"},{"id":"charlie"}]}`))
	}))
	defer server.Close()
	root := t.TempDir()
	connector := filepath.Join(root, "acme")
	if err := os.MkdirAll(filepath.Join(connector, "sources"), 0o700); err != nil {
		t.Fatal(err)
	}
	source := []byte(`{"schema_version":2,"connector":"acme","rest":{"operations":[{"id":"fixture.widgets","method":"GET","path":"/widgets","protocol":"rest","source_operation":{"operationId":"widgets","responses":{"200":{"description":"Page","content":{"application/json":{"schema":{"type":"object","required":["data"],"properties":{"data":{"type":"array","items":{"type":"object","required":["id"],"properties":{"id":{"type":"string"}}}}}}}}}}}}]},"counts":{"total":1}}`)
	switch variant {
	case "envelope_identity":
		source = bytes.Replace(source, []byte(`"protocol":"rest"`), []byte(`"protocol":"rest","operation_id":"widgets"`), 1)
		source = bytes.Replace(source, []byte(`"operationId":"widgets",`), nil, 1)
	case "contradictory_identity":
		source = bytes.Replace(source, []byte(`"protocol":"rest"`), []byte(`"protocol":"rest","operation_id":"different"`), 1)
	case "protocol":
		source = bytes.Replace(source, []byte(`"protocol":"rest"`), []byte(`"protocol":"graphql"`), 1)
	case "request_body":
		source = bytes.Replace(source, []byte(`"operationId":"widgets"`), []byte(`"operationId":"widgets","requestBody":{"required":true,"content":{"application/json":{"schema":{"type":"object","required":["filter"],"properties":{"filter":{"type":"string"}}}}}}`), 1)
	case "response_media":
		source = bytes.Replace(source, []byte(`"application/json"`), []byte(`"application/xml"`), 1)
	case "response_status":
		source = bytes.Replace(source, []byte(`"200"`), []byte(`"201"`), 1)
	}
	inputVariant := variant == "inherited_path" || variant == "operation_override" || variant == "optional_query_absent" || variant == "optional_query_present" || variant == "missing_path"
	expectedPath, expectedWire := "/widgets", "GET /widgets"
	runtimeConfig := map[string]string{}
	if inputVariant {
		var envelope map[string]any
		if err := json.Unmarshal(source, &envelope); err != nil {
			t.Fatal(err)
		}
		row := envelope["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)
		operation := row["source_operation"].(map[string]any)
		pathParameter := func(name string) map[string]any {
			return map[string]any{"in": "path", "name": name, "required": true, "schema": map[string]any{"type": "string"}}
		}
		if variant == "inherited_path" || variant == "operation_override" || variant == "missing_path" {
			row["path"] = "/namespaces/{namespace}/repositories/{repository}/widgets"
			operation["path_parameters"] = []any{pathParameter("namespace"), pathParameter("repository")}
			expectedPath = "/namespaces/{{ config.namespace }}/repositories/{{ config.repository }}/widgets"
			expectedWire = "GET /namespaces/acme/repositories/warehouse/widgets"
			runtimeConfig["namespace"], runtimeConfig["repository"] = "acme", "warehouse"
			if variant == "missing_path" {
				delete(runtimeConfig, "repository")
			}
			if variant == "operation_override" {
				override := pathParameter("namespace")
				override["schema"] = map[string]any{"type": "string", "enum": []any{"acme"}}
				operation["parameters"] = []any{override}
			}
		} else {
			operation["parameters"] = []any{map[string]any{"in": "query", "name": "owner-id", "required": false, "schema": map[string]any{"type": "string"}}}
			if variant == "optional_query_present" {
				runtimeConfig["owner-id"] = "a+b c"
				expectedWire = "GET /widgets?owner-id=a%2Bb+c"
			}
		}
		var err error
		source, err = json.Marshal(envelope)
		if err != nil {
			t.Fatal(err)
		}
	}
	if !json.Valid(source) {
		t.Fatal("invalid retained test fixture")
	}
	if variant == "raw_schema3" {
		var envelope map[string]any
		if err := json.Unmarshal(source, &envelope); err != nil {
			t.Fatal(err)
		}
		row := envelope["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)
		operation := row["source_operation"]
		delete(row, "source_operation")
		artifact, err := json.Marshal(map[string]any{"openapi": "3.0.0", "info": map[string]string{"title": "Independent fixture", "version": "1"}, "paths": map[string]any{"/widgets": map[string]any{"get": operation}}})
		if err != nil {
			t.Fatal(err)
		}
		artifactDir := filepath.Join(connector, "sources", "artifacts")
		if err := os.Mkdir(artifactDir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(artifactDir, sourceBytesHash(artifact)+".artifact"), artifact, 0o600); err != nil {
			t.Fatal(err)
		}
		envelope["schema_version"] = 3
		envelope["rest"] = map[string]any{"source_documents": []any{map[string]any{"id": "raw", "content_type": "application/json", "artifact": map[string]any{"sha256": sourceBytesHash(artifact), "bytes": len(artifact), "openapi": "3.0.0"}, "operations": []any{row}}}}
		source, err = json.Marshal(envelope)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(connector, "sources", "operations.json"), source, 0o600); err != nil {
		t.Fatal(err)
	}
	lock := minimalVNextLockForTest()
	lock.Operations, lock.Schemas = nil, nil
	if direct {
		lock.Lanes["direct_read"] = "implemented"
	}
	lock.HTTP = json.RawMessage(strings.Replace(string(lock.HTTP), "https://api.acme.example", server.URL, 1))
	pointer := "/data"
	lock.SourceProjection = &vNextSourceProjection{
		Version:     1,
		Inventories: []vNextSourceProjectionInventory{{ID: "primary", Path: "sources/operations.json", SHA256: sourceBytesHash(source), Bytes: int64(len(source))}},
		Semantics: []vNextSourceProjectionSemantic{{Source: vNextSourceProjectionKey{Inventory: "primary", ID: "fixture.widgets"}, Collection: &vNextSourceProjectionCollection{
			Records: vNextSourceCoordinate{Response: &vNextSourceResponseCoordinate{Status: "200", Media: "application/json", Pointer: &pointer}}, PrimaryKey: []string{"/id"},
		}}},
	}
	if variant == "response_media" {
		lock.SourceProjection.Semantics[0].Collection.Records.Response.Media = "application/xml"
	}
	if variant == "response_status" {
		lock.SourceProjection.Semantics[0].Collection.Records.Response.Status = "201"
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(connector, "source.lock.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := runLockRender([]string{"lock-render", "acme", "--defs", root}, &stdout, &stderr)
	if wantRefusal {
		if code == 0 {
			t.Fatal("unsupported retained source semantics published as implemented execution")
		}
		if len(requestSnapshot()) != 0 {
			t.Fatal("source refusal sent provider request")
		}
		if _, err := os.Lstat(filepath.Join(connector, vNextPublicationCurrentFile)); !os.IsNotExist(err) {
			t.Fatalf("source refusal changed CURRENT: %v", err)
		}
		return
	}
	if code != 0 {
		t.Fatalf("source-only lock-render refused before admitted execution: code=%d error=%s", code, stderr.String())
	}
	if len(requestSnapshot()) != 0 {
		t.Fatal("authoring contacted provider")
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
	execution := map[string][]byte{}
	for _, name := range handle.Files() {
		switch name {
		case "manifest.json", "provenance.json", "atlas.json", "index.json", "proof.json":
			continue
		}
		execution[name], err = handle.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
	}
	bundle, err := engine.Load(newVNextExecutionFS("acme", execution), "acme")
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Streams) != 1 || bundle.Streams[0].Name != "widgets" || bundle.Streams[0].Path != expectedPath {
		t.Fatalf("source route/identity not preserved: %+v", bundle.Streams)
	}
	if variant == "operation_override" {
		var schema map[string]any
		if err := json.Unmarshal(execution["spec.json"], &schema); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(schema["properties"].(map[string]any)["namespace"].(map[string]any)["enum"], []any{"acme"}) {
			t.Fatal("operation override constraint lost")
		}
	}
	var ids []string
	readErr := engine.Read(context.Background(), bundle, connectors.ReadRequest{Stream: "widgets", Config: connectors.RuntimeConfig{Config: runtimeConfig}}, nil, func(record connectors.Record) error {
		id, ok := record["id"].(string)
		if !ok {
			t.Error("returned record lost string identity")
		}
		ids = append(ids, id)
		return nil
	})
	if variant == "missing_path" {
		if readErr == nil || len(ids) != 0 || len(requestSnapshot()) != 0 {
			t.Fatalf("missing path input sent or emitted: error=%v ids=%v requests=%v", readErr, ids, requestSnapshot())
		}
		return
	}
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !reflect.DeepEqual(ids, []string{"alpha", "bravo", "charlie"}) || !reflect.DeepEqual(requestSnapshot(), []string{expectedWire}) {
		t.Fatalf("actual returned identities=%v requests=%v", ids, requestSnapshot())
	}
	if direct {
		consumer := engine.New(bundle, nil)
		if err := commandrunner.Preflight(consumer, []string{"api", "widgets"}); err != nil {
			t.Fatalf("generated direct preflight: %v", err)
		}
		flags := map[string][]string{}
		for key, value := range runtimeConfig {
			flags[key] = []string{value}
		}
		result, err := commandrunner.Run(t.Context(), consumer, commandrunner.Request{Path: []string{"api", "widgets"}, Flags: flags}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result.DirectRead == nil {
			t.Fatal("generated direct command returned no direct result")
		}
		body, err := json.Marshal(result.DirectRead.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != `{"data":[{"id":"alpha"},{"id":"bravo"},{"id":"charlie"}]}` || !reflect.DeepEqual(requestSnapshot(), []string{expectedWire, expectedWire}) {
			t.Fatalf("direct body=%s requests=%v", body, requestSnapshot())
		}
	}
}
