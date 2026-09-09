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

func TestSourceFactsCoverageConfidence(t *testing.T) {
	row, doc := retainedFactFixture(t, "vercel", "vercel.rest.readSessionFile")
	for _, unavailable := range []bool{false, true} {
		name := "machine snapshot"
		if unavailable {
			name = "unavailable"
		}
		t.Run(name, func(t *testing.T) {
			current := row
			current.Observed = !unavailable
			facts := normalizeSourceFacts(current, doc, nil)
			raw, err := json.Marshal(facts)
			if err != nil {
				t.Fatal(err)
			}
			var got struct {
				Confidence string   `json:"coverage_confidence"`
				Limits     []string `json:"completeness_limits"`
			}
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatal(err)
			}
			want := "machine_readable_snapshot"
			if unavailable {
				want = "partial"
			}
			if got.Confidence != want {
				t.Errorf("coverage=%q want=%q", got.Confidence, want)
			}
			if len(got.Limits) == 0 {
				t.Error("retained snapshot has no explicit completeness limit")
			}
			if !unavailable && facts.Refs["source_operation"].DocumentID != doc.ID {
				t.Error("snapshot lost source basis")
			}
		})
	}
	markup := []byte("<h2 id=download>Download file</h2><p>Download the raw bytes.</p>")
	payload, err := json.Marshal(string(markup))
	if err != nil {
		t.Fatal(err)
	}
	html := retainedSourceDocument{ID: "fixture:html", ContentType: "text/html", Payload: payload, RetainedFileSHA256: sourceBytesHash(markup), Bytes: int64(len(markup))}
	node := json.RawMessage(`{"protocol":"rest","method":"get","path":"/file"}`)
	facts := normalizeSourceFacts(retainedSourceOperation{Observed: true, Node: node, SourceLocation: "#download"}, retainedSourceDocument{ID: "fixture:lock", Payload: json.RawMessage(`{"rest":{}}`)}, &html)
	raw, err := json.Marshal(facts)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Confidence string   `json:"coverage_confidence"`
		Limits     []string `json:"completeness_limits"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Confidence != "rendered_reference" || len(got.Limits) == 0 || facts.Refs["rendered_reference"].Section != "#download" {
		t.Errorf("rendered coverage=%+v status=%s refs=%+v", got, facts.Status, facts.Refs)
	}
}

func TestSourceFacts118DuplicateParameterScopes(t *testing.T) {
	for _, scope := range []string{"path", "operation"} {
		for _, same := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/equal=%t", scope, same), func(t *testing.T) {
				first := `{"name":"limit","in":"query","required":true,"schema":{"type":"integer"}}`
				second := first
				if !same {
					second = `{"name":"limit","in":"query","required":false,"schema":{"type":"string"}}`
				}
				params := `[` + first + `,` + second + `]`
				pathParams, opParams := `[]`, `[]`
				prefix := "/paths/~1items/parameters"
				if scope == "path" {
					pathParams = params
				} else {
					opParams = params
					prefix = "/paths/~1items/get/parameters"
				}
				raw := retainedSourceDocument{ID: "raw", Payload: json.RawMessage(`{"paths":{"/items":{"parameters":` + pathParams + `,"get":{"parameters":` + opParams + `,"responses":{"200":{"description":"ok"}}}}}}`)}
				node := json.RawMessage(`{"id":"read","method":"GET","protocol":"rest","path":"/items"}`)
				facts := normalizeSourceFacts(retainedSourceOperation{Observed: true, Node: node}, retainedSourceDocument{ID: "archive", Payload: json.RawMessage(`{"rest":{}}`)}, &raw)
				if facts.Status != "available" || len(facts.Parameters) != 1 || len(facts.Groups["source_operation"]) == 0 {
					t.Fatalf("normalizer phase not reached: %+v", facts)
				}
				want := "source_parameter_duplicate:" + prefix + "/0:" + prefix + "/1"
				found := false
				for _, d := range facts.Diagnostics {
					found = found || d == want
				}
				if !found {
					t.Errorf("reached actual effective parameter reduction; duplicate must retain both occurrence pointers: want %s; got %v", want, facts.Diagnostics)
				}
				if facts.Parameters[0].Ref.Pointer != prefix+"/0" || !facts.Parameters[0].Required {
					t.Errorf("ambiguous scope must retain first known occurrence, not silently replace it: %+v", facts.Parameters)
				}
				var originals []json.RawMessage
				group := "parameters"
				if scope == "path" {
					group = "path_parameters"
				}
				if err := json.Unmarshal(facts.Groups[group], &originals); err != nil || len(originals) != 2 {
					t.Fatalf("original conflicting facts lost: %s %v", facts.Groups[group], err)
				}
			})
		}
	}
}

func source118ComponentsFixture(t *testing.T, rootRole bool, marker string) ([]retainedSourceOperation, retainedSourceDocument, *retainedSourceDocument, int) {
	t.Helper()
	components := `{"securitySchemes":{"auth":{"type":"http","scheme":"bearer"}},"schemas":{"Unused":{"type":"object","description":"` + marker + `","properties":{"id":{"type":"string"}}}}}`
	operation := `{"summary":"List items","responses":{"200":{"description":"ok","content":{"application/json":{"schema":{"type":"array","items":{"type":"object","properties":{"id":{"type":"string"}}}}}}}}}`
	nodes := []json.RawMessage{}
	rows := []retainedSourceOperation{}
	paths := map[string]json.RawMessage{}
	for i := 0; i < 3; i++ {
		op := operation
		if i == 1 {
			op = `{"security":[],` + operation[1:]
		}
		id := fmt.Sprintf("read.%d", i)
		path := fmt.Sprintf("/items%d", i)
		node := fmt.Sprintf(`{"id":%q,"method":"GET","protocol":"rest","path":%q`, id, path)
		if !rootRole {
			node += `,"source_operation":` + op
		}
		node += `}`
		nodes = append(nodes, json.RawMessage(node))
		paths[path] = json.RawMessage(`{"get":` + op + `}`)
		rows = append(rows, retainedSourceOperation{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: id}, Observed: true, Node: json.RawMessage(node), DocumentID: "archive:" + marker, Pointer: fmt.Sprintf("/rest/operations/%d", i)})
	}
	operations, err := json.Marshal(nodes)
	if err != nil {
		t.Fatal(err)
	}
	doc := retainedSourceDocument{ID: "archive:" + marker, Payload: json.RawMessage(`{"source_contract":{"security":[{"auth":[]}],"components":` + components + `},"rest":{"operations":` + string(operations) + `}}`)}
	if !rootRole {
		return rows, doc, nil, len(components)
	}
	encodedPaths, err := json.Marshal(paths)
	if err != nil {
		t.Fatal(err)
	}
	raw := retainedSourceDocument{ID: "raw:" + marker, Payload: json.RawMessage(`{"security":[{"auth":[]}],"components":` + components + `,"paths":` + string(encodedPaths) + `}`)}
	return rows, doc, &raw, len(components)
}

func TestSourceFacts118ComponentDecodeWork(t *testing.T) {
	for _, prepared := range []bool{false, true} {
		for _, rootRole := range []bool{false, true} {
			t.Run(fmt.Sprintf("prepared=%t/root=%t", prepared, rootRole), func(t *testing.T) {
				for _, marker := range []string{"first", "second"} {
					rows, doc, raw, componentBytes := source118ComponentsFixture(t, rootRole, marker)
					if prepared {
						var err error
						doc, err = prepareSourceDocument(doc)
						if err != nil {
							t.Fatal(err)
						}
						if raw != nil {
							p, err := prepareSourceDocument(*raw)
							if err != nil {
								t.Fatal(err)
							}
							raw = &p
						}
					}
					count, total := 0, 0
					observe := func(n int) { count++; total += n }
					for i, row := range rows {
						facts := normalizeSourceFactsObserved(row, doc, raw, observe)
						if facts.Status != "available" || len(facts.Diagnostics) != 0 {
							t.Fatalf("normalization not reached: %+v", facts.Diagnostics)
						}
						wantSecurity := `[{"auth":[]}]`
						if i == 1 {
							wantSecurity = `[]`
						}
						if string(facts.Groups["security"]) != wantSecurity {
							t.Fatalf("operation override changed: %s", facts.Groups["security"])
						}
						wantSchemes := `{"auth":{"scheme":"bearer","type":"http"}}`
						canonical, err := canonicalSourceJSON(facts.Groups["security_schemes"])
						if err != nil || string(canonical) != wantSchemes {
							t.Fatalf("known source schemes changed: %s %v", canonical, err)
						}
						wantID, wantPointer := doc.ID, "/source_contract/components/securitySchemes"
						if raw != nil {
							wantID, wantPointer = raw.ID, "/components/securitySchemes"
						}
						ref := facts.Refs["security_schemes"]
						if ref.DocumentID != wantID || ref.Pointer != wantPointer || ref.ValueSHA256 != sourceBytesHash([]byte(wantSchemes)) {
							t.Fatalf("literal citation changed: %+v", ref)
						}
						cells := classifySourceLanes(row.Key, facts, nil)
						expected := []string{"applicable", "not_applicable", "not_applicable", "not_applicable", "applicable", "not_applicable", "not_applicable"}
						if len(cells) != 7 {
							t.Fatalf("lane membership=%d", len(cells))
						}
						for j, c := range cells {
							if c.Lane != sourceLaneNames()[j] || c.Applicability != expected[j] || len(c.References) != 0 || len(c.ProofRefs) != 0 {
								t.Fatalf("lane %d not literal expected: %+v", j, c)
							}
						}
						// Returned groups are caller-owned; the next row must still see source bytes.
						for j := range facts.Groups["security_schemes"] {
							facts.Groups["security_schemes"][j] = ' '
						}
					}
					want := len(rows)
					if prepared {
						want = 1
					}
					if count != want || total != want*componentBytes {
						t.Errorf("completed strict component decode work=%d/%dbytes; want %d/%dbytes for %d operations and one document role", count, total, want, want*componentBytes, len(rows))
					}
				}
			})
		}
	}
}

func TestSourceFacts118ComponentPreservation(t *testing.T) {
	for _, rootRole := range []bool{false, true} {
		for _, component := range []string{"absent", "null", "{}", "[]", `"invalid"`, "42"} {
			t.Run(fmt.Sprintf("root=%t/components=%s", rootRole, component), func(t *testing.T) {
				rows, doc, raw, _ := source118ComponentsFixture(t, rootRole, "preservation")
				target := &doc
				if raw != nil {
					target = raw
				}
				var document map[string]json.RawMessage
				if err := json.Unmarshal(target.Payload, &document); err != nil {
					t.Fatal(err)
				}
				contract := document
				if !rootRole {
					if err := json.Unmarshal(document["source_contract"], &contract); err != nil {
						t.Fatal(err)
					}
				}
				if component == "absent" {
					delete(contract, "components")
				} else {
					contract["components"] = json.RawMessage(component)
				}
				if !rootRole {
					value, err := json.Marshal(contract)
					if err != nil {
						t.Fatal(err)
					}
					document["source_contract"] = value
				}
				target.Payload, _ = json.Marshal(document)
				originalDoc := doc
				var originalRaw *retainedSourceDocument
				if raw != nil {
					copy := *raw
					originalRaw = &copy
				}
				var err error
				doc, err = prepareSourceDocument(doc)
				if err != nil {
					t.Fatal(err)
				}
				if raw != nil {
					prepared, err := prepareSourceDocument(*raw)
					if err != nil {
						t.Fatal(err)
					}
					raw = &prepared
				}
				count := 0
				invalid := component == "[]" || component == `"invalid"` || component == "42"
				for _, row := range rows {
					got := normalizeSourceFactsObserved(row, doc, raw, func(int) { count++ })
					baseline := normalizeSourceFacts(row, originalDoc, originalRaw)
					a, _ := json.Marshal(got)
					b, _ := json.Marshal(baseline)
					if string(a) != string(b) {
						t.Fatalf("prepared/unprepared behavior differs: %s\n%s", a, b)
					}
					if got.Status != "available" || string(got.Groups["security_schemes"]) != "null" {
						t.Fatalf("source state changed: %+v", got)
					}
					expected := []string{}
					if invalid {
						expected = []string{"source_components_invalid"}
					}
					actual, _ := json.Marshal(got.Diagnostics)
					wanted, _ := json.Marshal(expected)
					if string(actual) != string(wanted) {
						t.Fatalf("exact old diagnostic frontier/order=%s want=%s", actual, wanted)
					}
				}
				expectedCount := 1
				if component == "absent" {
					expectedCount = 0
				}
				if count != expectedCount {
					t.Fatalf("actual component decode count=%d want=%d", count, expectedCount)
				}
			})
		}
		t.Run(fmt.Sprintf("root=%t/identity-and-early-return", rootRole), func(t *testing.T) {
			rows, doc, raw, _ := source118ComponentsFixture(t, rootRole, "identity")
			var err error
			doc, err = prepareSourceDocument(doc)
			if err != nil {
				t.Fatal(err)
			}
			if raw != nil {
				p, err := prepareSourceDocument(*raw)
				if err != nil {
					t.Fatal(err)
				}
				raw = &p
			}
			count := 0
			observe := func(int) { count++ }
			missing := rows[0]
			missing.Observed = false
			got := normalizeSourceFactsObserved(missing, doc, raw, observe)
			if count != 0 || len(got.Diagnostics) != 1 || got.Diagnostics[0] != "source_unavailable" {
				t.Fatalf("early return decoded components: %d %v", count, got.Diagnostics)
			}
			got = normalizeSourceFactsObserved(rows[0], doc, raw, observe)
			if got.Status != "available" || count != 1 {
				t.Fatalf("valid preparation path absent: %d %v", count, got.Diagnostics)
			}
			for _, mutation := range []string{"same-length", "whitespace"} {
				changed := doc
				var changedRaw *retainedSourceDocument
				if raw != nil {
					copy := *raw
					changedRaw = &copy
				}
				target := &changed
				if changedRaw != nil {
					target = changedRaw
				}
				target.Payload = append(json.RawMessage(nil), target.Payload...)
				if mutation == "same-length" {
					target.Payload = json.RawMessage(strings.Replace(string(target.Payload), "identity", "modified", 1))
				} else {
					target.Payload = append(target.Payload, ' ')
				}
				before := count
				got := normalizeSourceFactsObserved(rows[0], changed, changedRaw, observe)
				want := "source_document_invalid"
				if rootRole {
					want = "raw_document_invalid"
				}
				if count != before || len(got.Diagnostics) != 1 || got.Diagnostics[0] != want {
					t.Fatalf("payload mutation reused stale prepared selection: %s count=%d diagnostics=%v", mutation, count, got.Diagnostics)
				}
			}
			// Re-preparing an independent document with the same ID creates new custody.
			rows2, doc2, raw2, _ := source118ComponentsFixture(t, rootRole, "identity")
			doc2, err = prepareSourceDocument(doc2)
			if err != nil {
				t.Fatal(err)
			}
			if raw2 != nil {
				p, err := prepareSourceDocument(*raw2)
				if err != nil {
					t.Fatal(err)
				}
				raw2 = &p
			}
			normalizeSourceFactsObserved(rows2[0], doc2, raw2, observe)
			if count != 2 {
				t.Fatalf("fresh invocation skipped its real decode: %d", count)
			}
		})
	}
	for _, payload := range []string{`{"components":{"same":1,"same":2}}`, `{"components":{"value":01}}`, `{"components":{"value":NaN}}`} {
		if _, err := prepareSourceDocument(retainedSourceDocument{Payload: json.RawMessage(payload)}); err == nil {
			t.Fatalf("strict source decode relaxed: %s", payload)
		}
	}
}

func TestSourceFacts118LargeNumberPreserved(t *testing.T) {
	document, err := prepareSourceDocument(retainedSourceDocument{Payload: json.RawMessage(`{"components":{"value":1e1000}}`)})
	if err != nil {
		t.Fatal(err)
	}
	actual, err := sourceDocumentPointer(document, "/components/value")
	if err != nil || string(actual) != "1e1000" {
		t.Fatalf("retained arbitrary precision number changed: %s %v", actual, err)
	}
}

func TestSourceFacts118ComponentsEarlyFrontiers(t *testing.T) {
	archive := retainedSourceDocument{ID: "archive", Payload: json.RawMessage(`{"source_contract":{"components":[]},"rest":{}}`)}
	prepared, err := prepareSourceDocument(archive)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, node, want string
		raw              *retainedSourceDocument
	}{
		{"not retained", `{"id":"x","method":"GET","protocol":"rest","path":"/missing"}`, "source_operation_not_retained", nil},
		{"invalid operation", `{"id":"x","method":"GET","protocol":"rest","path":"/missing","source_operation":[]}`, "source_operation_invalid", nil},
		{"raw operation missing", `{"id":"x","method":"GET","protocol":"rest","path":"/missing"}`, "raw_operation_missing", &retainedSourceDocument{ID: "raw", Payload: json.RawMessage(`{"components":[],"paths":{}}`)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			count := 0
			got := normalizeSourceFactsObserved(retainedSourceOperation{Observed: true, Node: json.RawMessage(tc.node)}, prepared, tc.raw, func(int) { count++ })
			if count != 0 || len(got.Diagnostics) != 1 || got.Diagnostics[0] != tc.want {
				t.Fatalf("early frontier gained component work/diagnostic: count=%d got=%v want=%s", count, got.Diagnostics, tc.want)
			}
		})
	}
}
