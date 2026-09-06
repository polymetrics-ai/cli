package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestSourceFactsRetainedRenderedSupplements(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	lock, err := os.ReadFile(filepath.Join(root, "internal/connectors/defs/gitlab/sources/gitlab-binary-operation-source-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	var retained struct {
		Rest struct {
			Documents []struct {
				Operations []json.RawMessage `json:"operations"`
			} `json:"source_documents"`
		} `json:"rest"`
	}
	if err := json.Unmarshal(lock, &retained); err != nil {
		t.Fatal(err)
	}
	for i, tc := range []struct{ id, sha, section, title, clause string }{
		{"gitlab.docs.generic_packages.upload_file", "f59c93194c095d0e925a5751a08eb7a2176a26c6b5f38bda52f805154219d0f0", "#publish-a-single-file", "Publish a single file", "--upload-file path/to/file.txt"},
		{"gitlab.docs.repository_files.raw_download", "53244a720b8509536290e0058c946a246817c775c797df36f4c9aa1225fdf0a4", "#retrieve-a-raw-file-from-a-repository", "Retrieve a raw file from a repository", "Retrieves the raw file contents"},
	} {
		t.Run(tc.id, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(root, "internal/connectors/defs/gitlab/sources/artifacts", tc.sha+".artifact"))
			if err != nil {
				t.Fatal(err)
			}
			if sourceBytesHash(raw) != tc.sha {
				t.Fatal("retained rendered bytes drifted")
			}
			htmlJSON, err := json.Marshal(string(raw))
			if err != nil {
				t.Fatal(err)
			}
			row := retainedSourceOperation{Key: sourceOperationKey{Connector: "gitlab", Inventory: "binary-docs", ID: tc.id}, Observed: true, Node: retained.Rest.Documents[i].Operations[0], Pointer: fmt.Sprintf("/rest/source_documents/%d/operations/0", i), SourceLocation: tc.section}
			document := retainedSourceDocument{ID: "gitlab:binary-docs", Payload: lock}
			artifact := retainedSourceDocument{ID: "gitlab:binary-docs:raw:" + tc.sha, ContentType: "text/html", Payload: htmlJSON, RetainedFileSHA256: tc.sha, Bytes: int64(len(raw)), UpstreamBytesVerified: true}
			facts := normalizeSourceFacts(row, document, &artifact)
			if facts.Status != "partial" || sourceFactText(facts, "summary") != tc.title {
				t.Fatalf("retained section facts unavailable: status=%s summary=%q diagnostics=%v", facts.Status, sourceFactText(facts, "summary"), facts.Diagnostics)
			}
			if !strings.Contains(sourceFactText(facts, "rendered_reference"), tc.clause) {
				t.Fatalf("section lost independently expected clause %q", tc.clause)
			}
			expectedJSON, _ := json.Marshal(tc.title)
			ref := facts.Refs["summary"]
			if ref.DocumentID != artifact.ID || ref.Section != tc.section || ref.Part != "heading" || ref.ValueSHA256 != sourceBytesHash(expectedJSON) {
				t.Fatalf("section citation mismatch: %+v", ref)
			}
			if len(facts.Groups["request_body"]) != 0 || len(facts.Groups["responses"]) != 0 {
				t.Fatal("rendered section fabricated machine-readable request/response schemas")
			}
		})
	}
}

func TestSourceFactsRenderedSectionControls(t *testing.T) {
	markup := []byte(`<h2 id="before">Other</h2><p>unrelated</p><h2 id="selected">Read <span>file</span></h2><p>up<span>load</span> bytes</p><script>do not cite script</script><style>do not cite style</style><h3>Details</h3><p>bounded text</p><h2 id="after">After</h2><p>unrelated-after</p>`)
	heading, text, err := sourceRenderedSection(markup, "#selected")
	if err != nil || heading != "Read file" || text != "Read file upload bytes Details bounded text" {
		t.Fatalf("section text mismatch: heading=%q text=%q err=%v", heading, text, err)
	}
	for _, tc := range []struct{ name, markup, section string }{
		{"missing", string(markup), "#absent"},
		{"duplicate", `<h2 id=x>One</h2><h2 id=x>Two</h2>`, "#x"},
		{"unclosed heading", `<h2 id=x>One<p>body`, "#x"},
		{"not a section", string(markup), "https://example.invalid/#selected"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := sourceRenderedSection([]byte(tc.markup), tc.section); err == nil {
				t.Fatal("invalid section accepted")
			}
		})
	}
	payload, err := json.Marshal(string(markup))
	if err != nil {
		t.Fatal(err)
	}
	document := retainedSourceDocument{ID: "fixture:html", ContentType: "text/html", Payload: payload, RetainedFileSHA256: sourceBytesHash(markup), Bytes: int64(len(markup))}
	ref := sourceFactRef{DocumentID: document.ID, Section: "#selected", Part: "heading"}
	value, err := resolveSourceFactValue(document, ref)
	if err != nil || string(value) != `"Read file"` {
		t.Fatalf("citation resolution=%s err=%v", value, err)
	}
	ref.Part = "executable-selector"
	if _, err := resolveSourceFactValue(document, ref); err == nil {
		t.Fatal("unknown selector accepted")
	}
	ref.Part = "heading"
	document.RetainedFileSHA256 = "changed"
	if _, err := resolveSourceFactValue(document, ref); err == nil {
		t.Fatal("changed retained document accepted")
	}
}

func TestSourceFactsRenderedTokenBudget(t *testing.T) {
	raw := []byte(`<h2 id=x>Bounded</h2>` + strings.Repeat("<br>", 1000000))
	if _, _, err := sourceRenderedSection(raw, "#x"); err == nil || !strings.Contains(err.Error(), "token budget") {
		t.Fatalf("token limit not reached/refused: %v", err)
	}
}

func TestSourceFactsPreparedDocument(t *testing.T) {
	row, original := retainedFactFixture(t, "vercel", "vercel.rest.readSessionFile")
	prepared, err := prepareSourceDocument(original)
	if err != nil {
		t.Fatal(err)
	}
	before := normalizeSourceFacts(row, original, nil)
	after := normalizeSourceFacts(row, prepared, nil)
	a, _ := json.Marshal(before)
	b, _ := json.Marshal(after)
	if string(a) != string(b) {
		t.Fatal("prepared document changed normalized facts")
	}
	if after.Method != "POST" || after.Path != "/v2/sandboxes/sessions/{sessionId}/fs/read" || len(after.Groups["request_body"]) == 0 || len(after.Groups["responses"]) == 0 {
		t.Fatal("prepared facts lost independent provider contract")
	}
	expected, err := sourceJSONPointer(original.Payload, row.Pointer+"/source_operation")
	if err != nil {
		t.Fatal(err)
	}
	actual, err := sourceDocumentPointer(prepared, row.Pointer+"/source_operation")
	expectedCanonical, expectedErr := canonicalSourceJSON(expected)
	actualCanonical, actualErr := canonicalSourceJSON(actual)
	if err != nil || expectedErr != nil || actualErr != nil || string(expectedCanonical) != string(actualCanonical) {
		t.Fatalf("prepared pointer changed source value: %v", err)
	}
	if _, err := sourceDocumentPointer(prepared, "/absent"); err == nil {
		t.Fatal("missing pointer resolved")
	}
	prepared.Payload = append(append([]byte{}, prepared.Payload...), ' ')
	if _, err := sourceDocumentViewFor(prepared); err == nil {
		t.Fatal("mutated document reused old view")
	}
	for _, raw := range []string{"null", "[]", `{"duplicate":1,"duplicate":2}`} {
		if _, err := prepareSourceDocument(retainedSourceDocument{Payload: json.RawMessage(raw)}); err == nil {
			t.Fatalf("invalid prepared document accepted: %s", raw)
		}
	}
}
