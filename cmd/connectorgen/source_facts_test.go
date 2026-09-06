package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func retainedFactFixture(t *testing.T, connector, id string) (retainedSourceOperation, retainedSourceDocument) {
	t.Helper()
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("internal/connectors/defs", connector, "sources", connector+"-operation-source-lock.json")
	data, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Rest struct {
			Operations []json.RawMessage `json:"operations"`
		} `json:"rest"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	for i, node := range doc.Rest.Operations {
		var r struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(node, &r); err != nil {
			t.Fatal(err)
		}
		if r.ID == id {
			return retainedSourceOperation{Key: sourceOperationKey{Connector: connector, Inventory: "primary", ID: id}, Observed: true, Node: node, DocumentID: connector + ":primary", Pointer: fmt.Sprintf("/rest/operations/%d", i)}, retainedSourceDocument{ID: connector + ":primary", Path: path, Payload: data, RetainedFileSHA256: sourceBytesHash(data)}
		}
	}
	t.Fatalf("independent fixture source %q absent", id)
	return retainedSourceOperation{}, retainedSourceDocument{}
}

func TestSourceFactsVercelRequiredBodyAndBinary(t *testing.T) {
	row, doc := retainedFactFixture(t, "vercel", "vercel.rest.readSessionFile")
	facts := normalizeSourceFacts(row, doc, nil)
	if facts.Method != "POST" || facts.Path != "/v2/sandboxes/sessions/{sessionId}/fs/read" {
		t.Fatalf("lost provider route: %+v", facts)
	}
	var body struct {
		Required *bool `json:"required"`
		Content  map[string]struct {
			Schema struct {
				Required []string `json:"required"`
			}
		} `json:"content"`
	}
	if err := json.Unmarshal(facts.Groups["request_body"], &body); err != nil {
		t.Fatalf("lost retained request body: %v", err)
	}
	if body.Required != nil || len(body.Content["application/json"].Schema.Required) != 1 || body.Content["application/json"].Schema.Required[0] != "path" {
		t.Fatalf("outer requiredness or inner required path changed: %+v", body)
	}
	var responses map[string]struct {
		Content map[string]struct {
			Schema struct {
				Type   string `json:"type"`
				Format string `json:"format"`
			}
		} `json:"content"`
	}
	if err := json.Unmarshal(facts.Groups["responses"], &responses); err != nil {
		t.Fatal(err)
	}
	if responses["200"].Content["application/octet-stream"].Schema.Format != "binary" {
		t.Fatal("retained binary response omitted")
	}
	if facts.Refs["request_body"].DocumentID != doc.ID || facts.Refs["request_body"].ValueSHA256 == "" {
		t.Fatal("missing immutable body citation")
	}
}

func TestSourceFactsWebhookEventContract(t *testing.T) {
	row, doc := retainedFactFixture(t, "vercel", "vercel.rest.createWebhook")
	facts := normalizeSourceFacts(row, doc, nil)
	var body struct {
		Content map[string]struct {
			Schema struct {
				Required   []string `json:"required"`
				Properties map[string]struct {
					MinItems int `json:"minItems"`
					Items    struct {
						Enum []string `json:"enum"`
					} `json:"items"`
				} `json:"properties"`
			}
		} `json:"content"`
	}
	if err := json.Unmarshal(facts.Groups["request_body"], &body); err != nil {
		t.Fatalf("webhook facts missing: %v", err)
	}
	events := body.Content["application/json"].Schema.Properties["events"]
	if events.MinItems != 1 || len(events.Items.Enum) == 0 {
		t.Fatalf("webhook event contract lost: %+v", events)
	}
}

func TestSourceFactsEmptySecurityOverride(t *testing.T) {
	node := json.RawMessage(`{"id":"read","method":"GET","protocol":"rest","path":"/items","source_operation":{"security":[],"responses":{"200":{"description":"ok"}}}}`)
	doc := retainedSourceDocument{ID: "fixture", Payload: json.RawMessage(`{"source_contract":{"security":[{"bearer":[]}],"components":{"securitySchemes":{"bearer":{"type":"http","scheme":"bearer"}}}},"rest":{"operations":[` + string(node) + `]}}`)}
	row := retainedSourceOperation{Observed: true, Node: node, Pointer: "/rest/operations/0"}
	facts := normalizeSourceFacts(row, doc, nil)
	if string(facts.Groups["security"]) != "[]" {
		t.Fatalf("empty operation security overridden by root: %s", facts.Groups["security"])
	}
	if facts.Refs["security"].Pointer != "/rest/operations/0/source_operation/security" {
		t.Fatal("wrong effective auth citation")
	}
}

func TestSourceFactsYAMLIntegrityAndNumbers(t *testing.T) {
	raw, err := sourceYAMLDocument([]byte("openapi: 3.0.0\nvalue: 9007199254740993\n"))
	if err != nil {
		t.Fatal(err)
	}
	value, err := sourceJSONPointer(raw, "/value")
	if err != nil || string(value) != "9007199254740993" {
		t.Fatalf("numeric source changed: %s %v", value, err)
	}
	for _, input := range []string{"key: one\nkey: two\n", "key: one\n---\nkey: two\n", "key: &loop [*loop]\n"} {
		if _, err := sourceYAMLDocument([]byte(input)); err == nil {
			t.Fatalf("ambiguous/cyclic raw input accepted: %q", input)
		}
	}
}

func TestSourceFactsAsanaRawPath(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "internal/connectors/defs/asana/sources/artifacts/cb3b90f4e0af56035eab0c648974f625b942a28a7144aa6c2326e38ca0bb3d56.artifact"))
	if err != nil {
		t.Fatal(err)
	}
	document, err := sourceYAMLDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	source := retainedSourceDocument{ID: "asana:raw", Payload: document}
	row := retainedSourceOperation{Observed: true, Node: json.RawMessage(`{"id":"asana.rest.getAgent","method":"GET","protocol":"rest","path":"/agents/{agent_gid}","operation_id":"getAgent"}`)}
	facts := normalizeSourceFacts(row, retainedSourceDocument{ID: "asana:primary", Payload: json.RawMessage(`{"rest":{}}`)}, &source)
	if facts.Status != "available" {
		t.Fatalf("raw Asana facts unavailable: %+v", facts.Diagnostics)
	}
	if facts.Refs["responses"].DocumentID != "asana:raw" || facts.Refs["responses"].Pointer != "/paths/~1agents~1{agent_gid}/get/responses" {
		t.Fatalf("wrong source response citation: %+v", facts.Refs["responses"])
	}
	actual, err := sourceJSONPointer(document, facts.Refs["responses"].Pointer)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicalSourceJSON(actual)
	if err != nil {
		t.Fatal(err)
	}
	if sourceBytesHash(canonical) != facts.Refs["responses"].ValueSHA256 {
		t.Fatal("response fact digest is not its raw source value")
	}
}

func TestSourceFactsParameterPrecedence(t *testing.T) {
	raw := json.RawMessage(`{"paths":{"/items/{id}":{"parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"string"}},{"name":"limit","in":"query","required":true,"schema":{"type":"integer"}}],"get":{"parameters":[{"name":"limit","in":"query","required":false,"schema":{"type":"integer","maximum":5}}],"responses":{"200":{"description":"ok"}}}}}}`)
	source := retainedSourceDocument{ID: "raw", Payload: raw}
	row := retainedSourceOperation{Observed: true, Node: json.RawMessage(`{"id":"read","method":"GET","protocol":"rest","path":"/items/{id}"}`)}
	facts := normalizeSourceFacts(row, retainedSourceDocument{ID: "archive", Payload: json.RawMessage(`{"rest":{}}`)}, &source)
	if len(facts.Parameters) != 2 {
		t.Fatalf("effective parameters=%+v", facts.Parameters)
	}
	if facts.Parameters[0].In != "path" || facts.Parameters[0].Name != "id" || !facts.Parameters[0].Required {
		t.Fatalf("required path parameter lost: %+v", facts.Parameters)
	}
	query := facts.Parameters[1]
	if query.Name != "limit" || query.Required || query.Ref.Pointer != "/paths/~1items~1{id}/get/parameters/0" {
		t.Fatalf("operation override not preserved: %+v", query)
	}
	var param struct {
		Schema struct {
			Maximum int `json:"maximum"`
		} `json:"schema"`
	}
	if err := json.Unmarshal(query.Node, &param); err != nil || param.Schema.Maximum != 5 {
		t.Fatalf("effective bound lost: %s", query.Node)
	}
}
