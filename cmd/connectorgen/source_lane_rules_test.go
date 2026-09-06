package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func requireSourceLane(t *testing.T, cells []sourceLaneCell, lane, applicability string) sourceLaneCell {
	t.Helper()
	if len(cells) != 7 {
		t.Fatalf("got %d lane cells, want exactly seven", len(cells))
	}
	for _, cell := range cells {
		if cell.Lane == lane {
			if cell.Applicability != applicability {
				t.Fatalf("%s applicability=%s want %s; %+v", lane, cell.Applicability, applicability, cell)
			}
			return cell
		}
	}
	t.Fatalf("missing lane %s", lane)
	return sourceLaneCell{}
}

func TestSourceLaneVercelPostRead(t *testing.T) {
	row, doc := retainedFactFixture(t, "vercel", "vercel.rest.readSessionFile")
	facts := normalizeSourceFacts(row, doc, nil)
	var summary string
	if err := json.Unmarshal(facts.Groups["summary"], &summary); err != nil {
		t.Fatal(err)
	}
	a := sourceSemanticAnnotation{Key: row.Key, Semantics: "read", Citation: facts.Refs["summary"], Clause: summary}
	cells := classifySourceLanes(row.Key, facts, &a)
	for _, lane := range []string{"direct_read", "binary_download"} {
		requireSourceLane(t, cells, lane, "applicable")
	}
	for _, lane := range []string{"direct_write", "binary_upload", "etl", "reverse_etl", "sync_transport"} {
		requireSourceLane(t, cells, lane, "not_applicable")
	}
}

func TestSourceLaneCollectionAndMutation(t *testing.T) {
	for _, tc := range []struct {
		name, method, summary, response string
		read, write, etl, reverse       string
	}{
		{"finite collection", "GET", "List widgets", `{"type":"array","items":{"type":"object","properties":{"id":{"type":"string"}}}}`, "applicable", "not_applicable", "applicable", "not_applicable"},
		{"mutation", "POST", "Create widget", `{"type":"object","properties":{"id":{"type":"string"}}}`, "not_applicable", "applicable", "not_applicable", "applicable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			node := json.RawMessage(`{"id":"fixture","protocol":"rest","method":"` + tc.method + `","path":"/widgets","source_operation":{"summary":"` + tc.summary + `","responses":{"200":{"content":{"application/json":{"schema":` + tc.response + `}}}}}}`)
			row := retainedSourceOperation{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "fixture"}, Observed: true, Node: node, Pointer: "/rest/operations/0"}
			doc := retainedSourceDocument{ID: "fixture", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
			facts := normalizeSourceFacts(row, doc, nil)
			cells := classifySourceLanes(row.Key, facts, nil)
			requireSourceLane(t, cells, "direct_read", tc.read)
			requireSourceLane(t, cells, "direct_write", tc.write)
			requireSourceLane(t, cells, "etl", tc.etl)
			requireSourceLane(t, cells, "reverse_etl", tc.reverse)
		})
	}
}

func TestSourceLaneRuleCounterexample(t *testing.T) {
	row, doc := retainedFactFixture(t, "vercel", "vercel.rest.readSessionFile")
	facts := normalizeSourceFacts(row, doc, nil)
	var clause string
	if err := json.Unmarshal(facts.Groups["summary"], &clause); err != nil {
		t.Fatal(err)
	}
	a := sourceSemanticAnnotation{Key: row.Key, Semantics: "read", Citation: facts.Refs["summary"], Clause: clause}
	cells := classifySourceLanes(row.Key, facts, &a)
	want := map[string]string{"direct_read": "applicable", "direct_write": "not_applicable", "binary_download": "applicable", "binary_upload": "not_applicable", "etl": "not_applicable", "reverse_etl": "not_applicable", "sync_transport": "not_applicable"}
	if !sourceLaneOracleMatches(cells, want) {
		t.Fatal("valid independently specified seven-lane control rejected")
	}
	wrong := append([]sourceLaneCell(nil), cells...)
	wrong[1].Applicability = "applicable"
	if sourceLaneOracleMatches(wrong, want) {
		t.Fatal("oracle accepts POST read misclassified as mutation")
	}
	wrong = append([]sourceLaneCell(nil), cells...)
	wrong[2].Applicability = "not_applicable"
	if sourceLaneOracleMatches(wrong, want) {
		t.Fatal("oracle accepts missing binary applicability")
	}
	if sourceLaneOracleMatches(cells[:6], want) {
		t.Fatal("oracle accepts omitted lane")
	}
	a.Citation.ValueSHA256 = "wrong-readable-claim"
	invalid := classifySourceLanes(row.Key, facts, &a)
	if len(invalid[0].Diagnostics) == 0 || invalid[0].Diagnostics[0].Code != "source_annotation_invalid" {
		t.Fatal("changed citation accepted")
	}
}

func sourceLaneOracleMatches(cells []sourceLaneCell, expected map[string]string) bool {
	if len(cells) != 7 || len(expected) != 7 {
		return false
	}
	seen := map[string]bool{}
	for _, cell := range cells {
		if seen[cell.Lane] || expected[cell.Lane] != cell.Applicability {
			return false
		}
		seen[cell.Lane] = true
	}
	return true
}

func TestSourceLanePostReadNotion(t *testing.T) {
	row, doc := retainedFactFixture(t, "notion", "notion.rest.post-database-query")
	facts := normalizeSourceFacts(row, doc, nil)
	var summary string
	if err := json.Unmarshal(facts.Groups["summary"], &summary); err != nil {
		t.Fatal(err)
	}
	a := sourceSemanticAnnotation{Key: row.Key, Semantics: "read", Citation: facts.Refs["summary"], Clause: summary}
	cells := classifySourceLanes(row.Key, facts, &a)
	requireSourceLane(t, cells, "direct_read", "applicable")
	requireSourceLane(t, cells, "direct_write", "not_applicable")
}

func TestSourceLaneUnknownMediaAndPagingNotSync(t *testing.T) {
	node := json.RawMessage(`{"id":"fixture","protocol":"rest","method":"GET","path":"/widgets","source_operation":{"summary":"List widgets","parameters":[{"in":"query","name":"cursor","schema":{"type":"string"}}],"responses":{"200":{"description":"unknown media"}}}}`)
	row := retainedSourceOperation{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "fixture"}, Observed: true, Node: node, Pointer: "/rest/operations/0"}
	doc := retainedSourceDocument{ID: "fixture", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
	facts := normalizeSourceFacts(row, doc, nil)
	cells := classifySourceLanes(row.Key, facts, nil)
	requireSourceLane(t, cells, "sync_transport", "not_applicable")
	// A response with an entirely absent content declaration cannot prove that
	// a body-bearing 200 response has no binary/record representation.
	requireSourceLane(t, cells, "binary_download", "undetermined")
}

func TestSourceLaneBinaryLocalReference(t *testing.T) {
	facts := sourceFacts{Document: json.RawMessage(`{"components":{"schemas":{"File":{"type":"string","format":"binary"}}}}`), Groups: map[string]json.RawMessage{"request_body": json.RawMessage(`{"content":{"multipart/form-data":{"schema":{"type":"object","properties":{"file":{"$ref":"#/components/schemas/File"}}}}}}`)}}
	got := sourceRequestShape(facts)
	if !got.Binary || !got.Known {
		t.Fatalf("local binary part not recognized: %+v", got)
	}
	facts.Groups["request_body"] = json.RawMessage(`{"content":{"application/json":{"schema":{"$ref":"https://not-fetched.example/schema"}}}}`)
	got = sourceRequestShape(facts)
	if got.Known || got.Binary {
		t.Fatalf("external schema became an applicability proof: %+v", got)
	}
}

func TestSourceLaneAnnotationContradiction(t *testing.T) {
	for _, tc := range []struct {
		name, method, summary, interpretation string
		valid                                 bool
	}{
		{"post read claimed mutation", "POST", "Read a file", "mutation", false},
		{"mutation mentions read target", "POST", "Create a read token", "read", false},
		{"substring is not read semantics", "POST", "Update the playlist", "read", false},
		{"negated mutation", "POST", "Do not create a widget", "mutation", false},
		{"positive post read", "POST", "Read a file", "read", true},
		{"positive mutation", "POST", "Create a widget", "mutation", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			node, err := json.Marshal(map[string]any{"id": "fixture", "protocol": "rest", "method": tc.method, "path": "/widgets", "source_operation": map[string]any{"summary": tc.summary, "responses": map[string]any{"204": map[string]any{"description": "No content"}}}})
			if err != nil {
				t.Fatal(err)
			}
			row := retainedSourceOperation{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "fixture"}, Observed: true, Node: node, Pointer: "/rest/operations/0"}
			doc := retainedSourceDocument{ID: "fixture", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
			facts := normalizeSourceFacts(row, doc, nil)
			annotation := sourceSemanticAnnotation{Key: row.Key, Semantics: tc.interpretation, Citation: facts.Refs["summary"], Clause: tc.summary}
			cells := classifySourceLanes(row.Key, facts, &annotation)
			invalid := false
			for _, cell := range cells {
				for _, diagnostic := range cell.Diagnostics {
					if diagnostic.Code == "source_annotation_invalid" {
						invalid = true
					}
				}
			}
			if invalid == tc.valid {
				t.Fatalf("annotation %q for cited %q: invalid=%v, want %v; cells=%+v", tc.interpretation, tc.summary, invalid, !tc.valid, cells)
			}
			if tc.valid {
				lane := "direct_read"
				if tc.interpretation == "mutation" {
					lane = "direct_write"
				}
				requireSourceLane(t, cells, lane, "applicable")
			}
		})
	}
}

func TestSourceLaneRenderedUnknownContracts(t *testing.T) {
	node := json.RawMessage(`{"id":"fixture.file","method":"GET","path":"/file","protocol":"rest"}`)
	row := retainedSourceOperation{Key: sourceOperationKey{Connector: "fixture", Inventory: "supplement", ID: "fixture.file"}, Observed: true, Node: node, Pointer: "/rest/operations/0", SourceLocation: "#read"}
	doc := retainedSourceDocument{ID: "fixture:lock", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
	markup := `<h2 id=read>Read file</h2><p>The file can be read.</p>`
	payload, err := json.Marshal(markup)
	if err != nil {
		t.Fatal(err)
	}
	raw := retainedSourceDocument{ID: "fixture:html", ContentType: "text/html", Payload: payload, RetainedFileSHA256: sourceBytesHash([]byte(markup)), Bytes: int64(len(markup))}
	facts := normalizeSourceFacts(row, doc, &raw)
	cells := classifySourceLanes(row.Key, facts, nil)
	requireSourceLane(t, cells, "direct_read", "applicable")
	for _, lane := range []string{"binary_download", "binary_upload", "etl", "sync_transport"} {
		cell := requireSourceLane(t, cells, lane, "undetermined")
		if len(cell.Diagnostics) == 0 {
			t.Errorf("unresolved %s lacks a source-keyed diagnostic", lane)
		}
	}
}

func TestSourceLaneReceiverDemandVisibility(t *testing.T) {
	for _, tc := range []struct{ connector, id string }{
		{"bitbucket", "bitbucket.rest.post_/repositories/{workspace}/{repo_slug}/hooks"},
		{"bitbucket", "bitbucket.rest.post_/workspaces/{workspace}/hooks"},
		{"bitbucket", "bitbucket.rest.put_/repositories/{workspace}/{repo_slug}/hooks/{uid}"},
		{"bitbucket", "bitbucket.rest.put_/workspaces/{workspace}/hooks/{uid}"},
		{"circleci", "circleci.rest.createWebhook"}, {"circleci", "circleci.rest.updateWebhook"},
		{"gitlab", "postApiV4GroupsIdHooks"}, {"gitlab", "postApiV4Hooks"}, {"gitlab", "postApiV4ProjectsIdHooks"},
		{"jira", "jira.rest.registerDynamicWebhooks"}, {"stripe", "stripe.rest.PostWebhookEndpoints"},
		{"sentry", "sentry.rest.Register a New Service Hook"},
	} {
		t.Run(tc.connector+"/"+tc.id, func(t *testing.T) {
			row, doc := retainedFactFixture(t, tc.connector, tc.id)
			facts := normalizeSourceFacts(row, doc, nil)
			cells := classifySourceLanes(row.Key, facts, nil)
			if len(cells) != 7 {
				t.Fatal("source lost a lane")
			}
			cell := cells[6]
			if cell.Lane != "sync_transport" || cell.Applicability != "undetermined" || cell.State != "mapped_unproven" {
				t.Errorf("registration demand hidden or promoted: %+v", cell)
			}
			if len(cell.GapRefs) != 0 || len(cell.ProofRefs) != 0 || len(cell.References) != 0 {
				t.Errorf("unreconciled demand invented foundation/execution/proof: %+v", cell)
			}
		})
	}
}

func TestSourceLaneReceiverDemandClauseControls(t *testing.T) {
	for _, tc := range []struct {
		name, method, summary string
		want                  bool
	}{
		{"affirmative", "POST", "Create a webhook for the project", true},
		{"read registration config", "GET", "Get a webhook for the project", false},
		{"negative instruction", "POST", "Do not create a webhook", false},
		{"word lookalike", "POST", "Create a hookable widget", false},
		{"later example", "POST", "Create a widget. For updates, create a webhook.", false},
		{"quoted action", "POST", "The create webhook endpoint is deprecated", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			node, err := json.Marshal(map[string]any{"id": "provider.webhook_name_not_authority", "protocol": "rest", "method": tc.method, "path": "/items", "source_operation": map[string]any{"summary": tc.summary, "responses": map[string]any{}}})
			if err != nil {
				t.Fatal(err)
			}
			doc := retainedSourceDocument{ID: "fixture:primary", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
			facts := normalizeSourceFacts(retainedSourceOperation{Observed: true, Node: node, Pointer: "/rest/operations/0"}, doc, nil)
			ref, got := sourceRegistrationDemand(facts)
			if got != tc.want {
				t.Fatalf("registration=%v want=%v", got, tc.want)
			}
			if got && (ref.DocumentID != doc.ID || ref.Pointer != "/rest/operations/0/source_operation/summary") {
				t.Fatalf("wrong source witness %+v", ref)
			}
		})
	}
}

func sourceLaneProjectionFixture() sourceLaneTargetRef {
	a, b := "/properties/a", "/properties/b"
	return sourceLaneTargetRef{Kind: "write", Connector: "fixture", ID: "create", Lane: "direct_write", SchemaRole: sourceLaneSchemaRequest,
		SourceSchema: &sourceFactRef{DocumentID: "fixture:source", Pointer: "/request/schema", ValueSHA256: strings.Repeat("a", 64)},
		FieldMappings: []sourceLaneFieldMapping{
			{Source: sourceFactRef{DocumentID: "fixture:source", Pointer: "/request/schema/properties/a", ValueSHA256: strings.Repeat("b", 64)}, Target: sourceLaneFieldTarget{Kind: sourceLaneFieldSchema, Pointer: &a}},
			{Source: sourceFactRef{DocumentID: "fixture:source", Pointer: "/request/schema/properties/b", ValueSHA256: strings.Repeat("c", 64)}, Target: sourceLaneFieldTarget{Kind: sourceLaneFieldSchema, Pointer: &b}},
		}}
}

func TestSourceLaneTargetRefValueIdentity(t *testing.T) {
	original := sourceLaneProjectionFixture()
	for _, tc := range []struct {
		name   string
		change func(*sourceLaneTargetRef)
		equal  bool
	}{
		{"independent pointers", func(*sourceLaneTargetRef) {}, true},
		{"mapping order", func(r *sourceLaneTargetRef) {
			r.FieldMappings[0], r.FieldMappings[1] = r.FieldMappings[1], r.FieldMappings[0]
		}, true},
		{"source schema digest", func(r *sourceLaneTargetRef) { r.SourceSchema.ValueSHA256 = strings.Repeat("d", 64) }, false},
		{"source schema pointer", func(r *sourceLaneTargetRef) { r.SourceSchema.Pointer = "/other" }, false},
		{"role", func(r *sourceLaneTargetRef) { r.SchemaRole = sourceLaneSchemaRecord }, false},
		{"target kind", func(r *sourceLaneTargetRef) { r.FieldMappings[0].Target.Kind = sourceLaneFieldConfig }, false},
		{"target pointer", func(r *sourceLaneTargetRef) { *r.FieldMappings[0].Target.Pointer = "/properties/other" }, false},
		{"mapping source", func(r *sourceLaneTargetRef) { r.FieldMappings[0].Source.Pointer = "/other" }, false},
		{"mapping digest", func(r *sourceLaneTargetRef) { r.FieldMappings[0].Source.ValueSHA256 = strings.Repeat("d", 64) }, false},
		{"canonical identity", func(r *sourceLaneTargetRef) { r.CanonicalID = "other" }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			other := sourceLaneProjectionFixture()
			tc.change(&other)
			if got := sourceLaneTargetRefEqual(original, other); got != tc.equal {
				t.Fatalf("value identity=%v want%v", got, tc.equal)
			}
		})
	}
	empty := original
	empty.FieldMappings = nil
	other := empty
	other.FieldMappings = []sourceLaneFieldMapping{}
	if !sourceLaneTargetRefEqual(empty, other) {
		t.Fatal("nil/empty mapping lists differ")
	}
	copy := canonicalSourceLaneTargetRef(original)
	*copy.FieldMappings[0].Target.Pointer = "/changed"
	copy.SourceSchema.Pointer = "/changed"
	if original.SourceSchema.Pointer != "/request/schema" || *original.FieldMappings[0].Target.Pointer != "/properties/a" {
		t.Fatal("canonical copy retained caller-owned pointers")
	}
}

func TestSourceLaneProjectionShape(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*sourceLaneTargetRef)
		valid  bool
	}{
		{"valid", func(*sourceLaneTargetRef) {}, true},
		{"root schema", func(r *sourceLaneTargetRef) {
			r.FieldMappings = r.FieldMappings[:1]
			*r.FieldMappings[0].Target.Pointer = ""
		}, true},
		{"escaped fields", func(r *sourceLaneTargetRef) { *r.FieldMappings[0].Target.Pointer = "/properties/a~1b~0c" }, true},
		{"missing pointer", func(r *sourceLaneTargetRef) { r.FieldMappings[0].Target.Pointer = nil }, false},
		{"unknown tag", func(r *sourceLaneTargetRef) { r.FieldMappings[0].Target.Kind = "other" }, false},
		{"bad escape", func(r *sourceLaneTargetRef) { *r.FieldMappings[0].Target.Pointer = "/properties/a~2b" }, false},
		{"parameter index alias", func(r *sourceLaneTargetRef) {
			r.FieldMappings[0].Target.Kind = sourceLaneFieldParameter
			*r.FieldMappings[0].Target.Pointer = "/parameters/00"
		}, false},
		{"parameter coordinate", func(r *sourceLaneTargetRef) {
			r.FieldMappings[0].Target.Kind = sourceLaneFieldParameter
			*r.FieldMappings[0].Target.Pointer = "/parameters/0"
		}, true},
		{"config root", func(r *sourceLaneTargetRef) {
			r.FieldMappings[0].Target.Kind = sourceLaneFieldConfig
			*r.FieldMappings[0].Target.Pointer = ""
		}, false},
		{"duplicate", func(r *sourceLaneTargetRef) { r.FieldMappings = append(r.FieldMappings, r.FieldMappings[0]) }, false},
		{"same target different source", func(r *sourceLaneTargetRef) { *r.FieldMappings[1].Target.Pointer = *r.FieldMappings[0].Target.Pointer }, false},
		{"same source different target", func(r *sourceLaneTargetRef) { r.FieldMappings[1].Source = r.FieldMappings[0].Source }, false},
		{"target overlap", func(r *sourceLaneTargetRef) { *r.FieldMappings[1].Target.Pointer = "/properties/a/properties/b" }, false},
		{"source overlap", func(r *sourceLaneTargetRef) {
			r.FieldMappings[1].Source.Pointer = r.FieldMappings[0].Source.Pointer + "/properties/b"
		}, false},
		{"nonadjacent target overlap", func(r *sourceLaneTargetRef) {
			third := r.FieldMappings[1]
			third.Source.Pointer = "/third"
			p := "/properties/a/properties/b"
			third.Target.Pointer = &p
			r.FieldMappings = append(r.FieldMappings, third)
			*r.FieldMappings[1].Target.Pointer = "/properties/a-other"
		}, false},
		{"transport projection", func(r *sourceLaneTargetRef) { r.Kind = "sync_transport" }, false},
		{"missing role", func(r *sourceLaneTargetRef) { r.SchemaRole = "" }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ref := sourceLaneProjectionFixture()
			tc.change(&ref)
			if got := sourceLaneTargetRefShape(ref) == nil; got != tc.valid {
				t.Fatalf("shape valid=%v want%v", got, tc.valid)
			}
		})
	}
}

func TestSourceLaneProjectionJSONShapes(t *testing.T) {
	base := sourceLaneProjectionFixture()
	for _, tc := range []struct {
		name  string
		edit  func(map[string]any)
		valid bool
	}{
		{"new optional absent", func(m map[string]any) { delete(m, "source_schema"); delete(m, "field_mappings") }, true},
		{"new optional null", func(m map[string]any) { m["source_schema"] = nil; m["field_mappings"] = nil }, true},
		{"empty mapping list", func(m map[string]any) { m["field_mappings"] = []any{} }, true},
		{"empty citation", func(m map[string]any) { m["source_schema"] = map[string]any{} }, false},
		{"citation pointer null", func(m map[string]any) { m["source_schema"].(map[string]any)["pointer"] = nil }, false},
		{"citation pointer missing", func(m map[string]any) { delete(m["source_schema"].(map[string]any), "pointer") }, false},
		{"role null", func(m map[string]any) { m["schema_role"] = nil }, false},
		{"target pointer null", func(m map[string]any) {
			m["field_mappings"].([]any)[0].(map[string]any)["target"].(map[string]any)["pointer"] = nil
		}, false},
		{"target pointer missing", func(m map[string]any) {
			delete(m["field_mappings"].([]any)[0].(map[string]any)["target"].(map[string]any), "pointer")
		}, false},
		{"target null", func(m map[string]any) { m["field_mappings"].([]any)[0].(map[string]any)["target"] = nil }, false},
		{"unknown field", func(m map[string]any) {
			m["field_mappings"].([]any)[0].(map[string]any)["target"].(map[string]any)["extra"] = true
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, _ := json.Marshal(base)
			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}
			tc.edit(doc)
			raw, _ = json.Marshal(doc)
			var ref sourceLaneTargetRef
			err := decodeStrictJSON(raw, &ref)
			if err == nil {
				err = sourceLaneTargetRefShape(ref)
			}
			if (err == nil) != tc.valid {
				t.Fatalf("JSON shape accepted=%v want%v: %s (%v)", err == nil, tc.valid, raw, err)
			}
		})
	}
}

func TestSourceLaneGraphQLCitationShapes(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		valid bool
	}{
		{`{}`, true},
		{`{"graphql":null}`, true},
		{`{"graphql":{}}`, true},
		{`{"graphql":{"operation_name":null,"document":null,"request_schema":null}}`, true},
		{`{"graphql":{"document":{}}}`, false},
		{`{"graphql":{"root_field":"widgets"}}`, false},
	} {
		var annotation sourceSemanticAnnotation
		err := decodeStrictJSON([]byte(tc.raw), &annotation)
		valid := err == nil && sourceLaneGraphQLRefsShape(annotation.GraphQL)
		if valid != tc.valid {
			t.Errorf("GraphQL JSON shape valid=%v want%v for %s: %v", valid, tc.valid, tc.raw, err)
		}
	}
	ref := sourceFactRef{DocumentID: "fixture:graphql", Pointer: "/operation/document", ValueSHA256: strings.Repeat("a", 64)}
	refs := &sourceLaneGraphQLRefs{Document: &ref}
	if !sourceLaneGraphQLRefsShape(refs) {
		t.Fatal("literal JSON citation rejected")
	}
	ref.Section = "#heading"
	if sourceLaneGraphQLRefsShape(refs) {
		t.Fatal("rendered quote used as machine-readable GraphQL document")
	}
}
