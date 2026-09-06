package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/syncplan"
)

const sourceBindingBody099F = `{"type":"object","properties":{"data":{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false}},"required":["data"],"additionalProperties":false}`
const sourceBindingRecord099F = `{"type":"object","properties":{"id":{"type":"string"}},"required":["id"],"additionalProperties":false}`
const sourceBindingEnvelope099F = `{"type":"object","properties":{"data":{"type":"array","items":` + sourceBindingRecord099F + `}},"required":["data"],"additionalProperties":false}`

// The source literals above are retained independently of each target and its
// mutations. Canonicalization and engine.Load must succeed before any oracle.
func sourceBindingFixture099F(t *testing.T, kind string, change func(*vNextSourceLock)) (sourceOperationKey, sourceFacts, sourceSemanticAnnotation) {
	t.Helper()
	lock := minimalVNextLockForTest()
	method, summary, route := "GET", "Get widgets", "/widgets"
	opSource := `"responses":{"200":{"content":{"application/json":{"schema":` + sourceBindingEnvelope099F + `}}}}`
	artifact, id, canonical, field, lane := "streams.json", "widgets", "stream:widgets", "stream", "etl"
	role := sourceLaneSchemaRecord
	anchorSuffix := "/responses/200/content/application~1json/schema"
	if kind == "body" {
		method, summary, route = "POST", "Create widgets", "/widgets/{id}"
		opSource = `"parameters":[{"name":"id","in":"path","required":true,"schema":{"type":"string"}}],"requestBody":{"required":true,"content":{"application/json":{"schema":` + sourceBindingBody099F + `}}},"responses":{"204":{"description":"No content"}}`
		artifact, id, canonical, field, lane = "writes.json", "create_widget", "write:widgets.create", "write", "direct_write"
		role, anchorSuffix = sourceLaneSchemaRequest, "/requestBody/content/application~1json/schema"
		lock.Lanes["etl"] = "unsupported"
		lock.Schemas["schemas/request.json"] = json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"},"data":{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false}},"required":["id","data"],"additionalProperties":false}`)
		lock.Operations = []vNextOperationDescriptor{{ID: canonical, SchemaRefs: vNextSchemaReferences{Request: "schemas/request.json"}, Write: json.RawMessage(`{"name":"create_widget","kind":"create","method":"POST","path":"/widgets/{{ record.id }}","path_fields":["id"],"body_type":"json","body_required":true,"body_fields":["data"],"record_schema":` + string(lock.Schemas["schemas/request.json"]) + `,"risk":"low"}`)}}
	} else {
		lock.Schemas["schemas/widgets.json"] = json.RawMessage(sourceBindingRecord099F)
	}
	if change != nil {
		change(&lock)
	}
	descriptor, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatalf("fixture must reach semantic binding, admission failed: %v", err)
	}
	key := sourceOperationKey{Connector: "acme", Inventory: "primary", ID: "retained.widgets"}
	node := json.RawMessage(`{"id":"retained.widgets","protocol":"rest","method":"` + method + `","path":"` + route + `","source_operation":{"summary":"` + summary + `",` + opSource + `}}`)
	doc := retainedSourceDocument{ID: "fixture:099F", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
	facts := normalizeSourceFacts(retainedSourceOperation{Key: key, Observed: true, Node: node, Pointer: "/rest/operations/0"}, doc, nil)
	facts.bindings = sourceBindingTestInputs(t, map[string][]byte{}, map[string]vNextCanonicalDescriptor{"acme": descriptor})
	authored, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	facts.bindings.Authoring = map[string][]byte{"internal/connectors/defs/acme/source.lock.json": authored}
	if descriptor.Staged.Identity.Digest == "" || facts.bindings.Bundles["acme"].Name != "acme" {
		t.Fatal("actual admission/load witness missing")
	}
	ref := sourceLaneTargetRef{Kind: field, Connector: "acme", ID: id, Lane: lane, Artifact: "internal/connectors/defs/acme/" + artifact, Pointer: "/streams/0", ArtifactSHA256: sourceBytesHash(descriptor.Staged.Outputs[artifact]), CanonicalID: canonical, CanonicalPointer: "/operations/0/" + field, Generation: descriptor.Staged.Identity.Digest, SchemaRole: role}
	if kind == "body" {
		ref.Pointer = "/actions/0"
	}
	base := "/rest/operations/0/source_operation"
	anchor := sourceBindingCitation099F(t, facts, base+anchorSuffix)
	ref.SourceSchema = &anchor
	mapField := func(source, target string) sourceLaneFieldMapping {
		return sourceLaneFieldMapping{Source: sourceBindingCitation099F(t, facts, source), Target: sourceLaneFieldTarget{Kind: sourceLaneFieldSchema, Pointer: &target}}
	}
	if kind == "body" {
		ref.FieldMappings = []sourceLaneFieldMapping{mapField(base+"/parameters/0", "/properties/id"), mapField(anchor.Pointer+"/properties/data", "/properties/data")}
	} else {
		ref.FieldMappings = []sourceLaneFieldMapping{mapField(anchor.Pointer+"/properties/data/items", "")}
	}
	return key, facts, sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: summary, IntendedBindings: []sourceLaneTargetRef{ref}}
}

func sourceBindingCitation099F(t *testing.T, facts sourceFacts, pointer string) sourceFactRef {
	t.Helper()
	raw, err := sourceJSONPointer(facts.Document, pointer)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicalSourceJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	return sourceFactRef{DocumentID: "fixture:099F", Pointer: pointer, ValueSHA256: sourceBytesHash(canonical)}
}

func sourceBindingOutcome099F(t *testing.T, key sourceOperationKey, facts sourceFacts, a sourceSemanticAnnotation, wantAccepted int, wantCodes ...string) sourceLaneCell {
	t.Helper()
	cells := classifySourceLanes(key, facts, &a)
	if len(cells) != 7 {
		t.Fatalf("lost seven cells: %d", len(cells))
	}
	lane := a.IntendedBindings[0].Lane
	cell := requireSourceLane(t, cells, lane, "applicable")
	if len(cell.References) != wantAccepted {
		t.Errorf("accepted=%d want=%d; diagnostics=%+v", len(cell.References), wantAccepted, cell.Diagnostics)
	}
	for _, code := range wantCodes {
		found := false
		for _, d := range cell.Diagnostics {
			if d.Code == code {
				found = true
				if d.Key != key || d.Stage != "reference" || len(d.Lanes) != 1 || d.Lanes[0] != lane {
					t.Errorf("wrong diagnostic scope: %+v", d)
				}
			}
		}
		if !found {
			t.Errorf("want %s; got %+v", code, cell.Diagnostics)
		}
	}
	if cell.State != "mapped_unproven" || len(cell.ProofRefs) != 0 {
		t.Errorf("binding promoted behavior: %+v", cell)
	}
	return cell
}

func TestSourceLaneBinding118AUnresolvedSuccessCoverage(t *testing.T) {
	for _, mode := range []string{"external", "missing local", "cycle", "unsuccessful"} {
		t.Run(mode, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
			before := sourceBindingOutcome099F(t, key, facts, a, 1)
			if len(before.Diagnostics) != 0 || len(before.References) != 1 {
				t.Fatal("admitted complete200 counterpart required")
			}
			good := before.References[0]
			status := "201"
			if mode == "unsuccessful" {
				status = "400"
			}
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				ref := "https://provider.invalid/response"
				if mode == "missing local" {
					ref = "#/responses/missing"
				}
				if mode == "cycle" {
					ref = "#/responses/cycle"
					doc["source_contract"] = map[string]any{"responses": map[string]any{"cycle": map[string]any{"$ref": ref}}}
				}
				op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
				op["responses"].(map[string]any)[status] = map[string]any{"$ref": ref}
			})
			cell := sourceBindingOutcome099F(t, key, facts, a, 1)
			if len(cell.References) != 1 || !sourceLaneTargetRefEqual(cell.References[0], good) {
				t.Fatal("exact accepted200 sibling lost")
			}
			want := mode != "unsuccessful"
			found := false
			for _, d := range cell.Diagnostics {
				if d.Code == "target_required_scope_unverified" && d.Pointer == "/rest/operations/0/source_operation/responses/"+status {
					found = true
					if d.Stage != "reference" || d.Severity != "deficit" || d.Key != key {
						t.Errorf("incorrect scope diagnostic: %+v", d)
					}
				}
			}
			if found != want {
				t.Errorf("unresolved successful scope diagnostic=%v want=%v: %+v", found, want, cell.Diagnostics)
			}
		})
	}
}

func TestSourceLaneBinding118AUnresolvedRootCoverage(t *testing.T) {
	for _, group := range []string{"request_body", "responses"} {
		t.Run(group, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
			before := classifySourceLanes(key, facts, &a)
			cell := requireSourceLane(t, before, "etl", "applicable")
			if len(cell.References) != 1 || len(cell.Diagnostics) != 0 {
				t.Fatal("real admitted coverage input required")
			}
			// Isolate the coverage boundary using the independently accepted reference;
			// the newly retained root is unresolved, not an absent optional group.
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
				name := "requestBody"
				if group == "responses" {
					name = "responses"
				}
				op[name] = map[string]any{"$ref": "https://provider.invalid/root"}
			})
			after := sourceLaneCheckCoverage(key, facts, a, before)
			got := requireSourceLane(t, after, "etl", "applicable")
			found := false
			for _, d := range got.Diagnostics {
				if d.Code == "target_required_scope_unverified" && d.Pointer == facts.Refs[group].Pointer {
					found = true
				}
			}
			if !found || len(after) != 7 || len(got.References) != 1 {
				t.Errorf("unresolved root lost coverage obligation: %+v", got)
			}
		})
	}
}

func TestSourceLaneBinding099FBodyAndEnvelope(t *testing.T) {
	for _, kind := range []string{"body", "envelope"} {
		t.Run(kind, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, kind, nil)
			cell := sourceBindingOutcome099F(t, key, facts, a, 1)
			if len(cell.Diagnostics) != 0 {
				t.Errorf("complete literal counterpart has diagnostics: %+v", cell.Diagnostics)
			}
			if len(cell.References) == 1 && !sourceLaneTargetRefEqual(cell.References[0], a.IntendedBindings[0]) {
				t.Error("accepted identity/projection changed")
			}
		})
	}
	for _, tc := range []struct{ name, old, new string }{
		{"nested required omitted", `"required":["name"]`, `"required":[]`},
		{"nested type changed", `"name":{"type":"string"}`, `"name":{"type":"integer"}`},
		{"body field omitted", `"body_fields":["data"]`, `"body_fields":["id"]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "body", func(lock *vNextSourceLock) {
				lock.Operations[0].Write = json.RawMessage(strings.ReplaceAll(string(lock.Operations[0].Write), tc.old, tc.new))
				lock.Schemas["schemas/request.json"] = json.RawMessage(strings.ReplaceAll(string(lock.Schemas["schemas/request.json"]), tc.old, tc.new))
			})
			code := "target_schema_mismatch"
			if tc.name == "body field omitted" {
				code = "target_body_projection_mismatch"
			}
			sourceBindingOutcome099F(t, key, facts, a, 0, code)
		})
	}
	for _, tc := range []struct {
		name   string
		change func(*sourceLaneTargetRef)
	}{
		{"wrong role", func(r *sourceLaneTargetRef) { r.SchemaRole = sourceLaneSchemaRequest }},
		{"envelope mistaken for item", func(r *sourceLaneTargetRef) { r.FieldMappings[0].Source = *r.SourceSchema }},
		{"false raw citation digest", func(r *sourceLaneTargetRef) { r.SourceSchema.ValueSHA256 = strings.Repeat("0", 64) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
			tc.change(&a.IntendedBindings[0])
			code := "target_schema_role_mismatch"
			if tc.name == "envelope mistaken for item" {
				code = "target_record_projection_mismatch"
			}
			if tc.name == "false raw citation digest" {
				code = "source_binding_citation_mismatch"
			}
			sourceBindingOutcome099F(t, key, facts, a, 0, code)
		})
	}
}

func TestSourceLaneBinding099FSiblingScopes(t *testing.T) {
	for _, badFirst := range []bool{false, true} {
		t.Run(strconv.FormatBool(badFirst), func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
			good := canonicalSourceLaneTargetRef(a.IntendedBindings[0])
			bad := canonicalSourceLaneTargetRef(good)
			bad.ID = "absent"
			a.IntendedBindings = []sourceLaneTargetRef{good, bad}
			if badFirst {
				a.IntendedBindings = []sourceLaneTargetRef{bad, good}
			}
			cell := sourceBindingOutcome099F(t, key, facts, a, 1, "target_absent")
			if len(cell.References) == 1 && !sourceLaneTargetRefEqual(cell.References[0], good) {
				t.Error("wrong sibling accepted")
			}
			if len(cell.References) == 1 {
				*cell.References[0].FieldMappings[0].Target.Pointer = "changed"
				if *good.FieldMappings[0].Target.Pointer != "" {
					t.Error("accepted ref aliases caller memory")
				}
			}
		})
	}
}

func TestSourceLaneBinding099FRegistryCanonical(t *testing.T) {
	t.Run("registry extracted root", func(t *testing.T) {
		key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
		r := &a.IntendedBindings[0]
		r.Kind = "schema"
		r.ID = "schemas/widgets.json"
		r.Artifact = "internal/connectors/defs/acme/" + r.ID
		r.Pointer = ""
		r.ArtifactSHA256 = sourceBytesHash(facts.bindings.Artifacts[r.Artifact])
		r.CanonicalPointer = "/operations/0/schema_refs/record"
		sourceBindingOutcome099F(t, key, facts, a, 1)
	})
	for _, wrong := range []bool{false, true} {
		t.Run("authored versus canonical index/"+strconv.FormatBool(wrong), func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "envelope", func(lock *vNextSourceLock) {
				lock.Operations[0].ID = "z"
				lock.Operations = append(lock.Operations, vNextOperationDescriptor{ID: "a", StreamOrder: 1, SchemaRefs: vNextSchemaReferences{Record: "schemas/widgets.json"}, Stream: json.RawMessage(`{"name":"widgets_a","path":"/widgets","records":{"path":"data"},"schema":"schemas/widgets.json"}`)})
			})
			r := &a.IntendedBindings[0]
			r.Kind = "canonical_operation"
			r.ID = "a"
			r.CanonicalID = "a"
			r.Artifact = "internal/connectors/defs/acme/source.lock.json"
			r.ArtifactSHA256 = sourceBytesHash(facts.bindings.Authoring[r.Artifact])
			r.Pointer = "/operations/1"
			r.CanonicalPointer = "/operations/0"
			if wrong {
				r.CanonicalPointer = "/operations/1"
			}
			want := 1
			codes := []string{}
			if wrong {
				want = 0
				codes = []string{"canonical_provenance_mismatch"}
			}
			sourceBindingOutcome099F(t, key, facts, a, want, codes...)
		})
	}
}

func TestSourceLaneBinding099FGraphQL(t *testing.T) {
	for _, tc := range []struct{ name, document, operation, schema, want string }{
		{"matching exact document", "query FirstWidgets { widgets { id } }", "FirstWidgets", `{"type":"object","properties":{},"additionalProperties":false}`, "target_response_contract_unverified"},
		{"same route other operation", "query SecondWidgets { widgets { id } }", "SecondWidgets", `{"type":"object","properties":{},"additionalProperties":false}`, "target_graphql_operation_mismatch"},
		{"same name other root", "query FirstWidgets { otherWidgets { id } }", "FirstWidgets", `{"type":"object","properties":{},"additionalProperties":false}`, "target_graphql_document_mismatch"},
		{"wrong variables schema", "query FirstWidgets($owner: String!) { widgets(owner: $owner) { id } }", "FirstWidgets", `{"type":"object","properties":{"owner":{"type":"string"}},"required":["owner"],"additionalProperties":false}`, "target_schema_mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lock := graphQLSourceLockForSemanticAdmissionTest()
			lock.Operations = lock.Operations[:1]
			lock.Operations[0].Source = json.RawMessage(`{"provider_operation":"` + tc.operation + `","method":"POST","path":"/graphql"}`)
			var op map[string]any
			if json.Unmarshal(lock.Operations[0].Operation, &op) != nil {
				t.Fatal("fixture operation")
			}
			g := op["graphql"].(map[string]any)
			g["document"] = tc.document
			g["operation_name"] = tc.operation
			var schema any
			if json.Unmarshal([]byte(tc.schema), &schema) != nil {
				t.Fatal("fixture schema")
			}
			g["variables_schema"] = schema
			lock.Operations[0].Operation, _ = json.Marshal(op)
			descriptor, err := canonicalizeVNextSourceLock(lock)
			if err != nil {
				t.Fatalf("must reach actual admitted GraphQL binding: %v", err)
			}
			key := sourceOperationKey{Connector: "acme", Inventory: "primary", ID: "retained.graphql"}
			node := json.RawMessage(`{"id":"retained.graphql","protocol":"graphql","method":"POST","path":"/graphql","source_operation":{"summary":"Read widgets","operation_name":"FirstWidgets","document":"query FirstWidgets { widgets { id } }","variables_schema":{"type":"object","properties":{},"additionalProperties":false},"responses":{"200":{"description":"Response schema unavailable"}}}}`)
			facts := normalizeSourceFacts(retainedSourceOperation{Key: key, Observed: true, Node: node, Pointer: "/rest/operations/0"}, retainedSourceDocument{ID: "fixture:099F", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}, nil)
			facts.bindings = sourceBindingTestInputs(t, map[string][]byte{}, map[string]vNextCanonicalDescriptor{"acme": descriptor})
			if facts.bindings.Bundles["acme"].Operations[0].GraphQL == nil {
				t.Fatal("loaded GraphQL stage missing")
			}
			base := "/rest/operations/0/source_operation/"
			name := sourceBindingCitation099F(t, facts, base+"operation_name")
			document := sourceBindingCitation099F(t, facts, base+"document")
			request := sourceBindingCitation099F(t, facts, base+"variables_schema")
			empty := ""
			ref := sourceLaneTargetRef{Kind: "operation", Connector: "acme", ID: "widgets.first", Lane: "direct_read", Artifact: "internal/connectors/defs/acme/operations.json", Pointer: "/operations/0", ArtifactSHA256: sourceBytesHash(descriptor.Staged.Outputs["operations.json"]), CanonicalID: "source:widgets.first", CanonicalPointer: "/operations/0/operation", Generation: descriptor.Staged.Identity.Digest, SchemaRole: sourceLaneSchemaRequest, SourceSchema: &request, FieldMappings: []sourceLaneFieldMapping{{Source: request, Target: sourceLaneFieldTarget{Kind: sourceLaneFieldSchema, Pointer: &empty}}}}
			a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: "Read widgets", Semantics: "read", GraphQL: &sourceLaneGraphQLRefs{OperationName: &name, Document: &document, RequestSchema: &request}, IntendedBindings: []sourceLaneTargetRef{ref}}
			cell := sourceBindingOutcome099F(t, key, facts, a, 0, tc.want)
			if tc.name == "matching exact document" && len(cell.Diagnostics) != 1 {
				t.Errorf("matched request projection must reach only absent response consumer: %+v", cell.Diagnostics)
			}
		})
	}
}

func sourceBindingRepin099F(t *testing.T, key sourceOperationKey, facts sourceFacts, change func(map[string]any)) sourceFacts {
	t.Helper()
	var document map[string]any
	if json.Unmarshal(facts.Document, &document) != nil {
		t.Fatal("document decode")
	}
	change(document)
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	node, err := sourceJSONPointer(raw, "/rest/operations/0")
	if err != nil {
		t.Fatal(err)
	}
	updated := normalizeSourceFacts(retainedSourceOperation{Key: key, Observed: true, Node: node, Pointer: "/rest/operations/0"}, retainedSourceDocument{ID: "fixture:099F", Payload: raw}, nil)
	updated.bindings = facts.bindings
	return updated
}

func TestSourceLaneBinding099FMultipleScopes(t *testing.T) {
	for _, mode := range []string{"both", "missing second", "bad second first", "bad second last"} {
		t.Run(mode, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
				var response any
				_ = json.Unmarshal([]byte(`{"content":{"application/json":{"schema":`+sourceBindingEnvelope099F+`}}}`), &response)
				op["responses"].(map[string]any)["201"] = response
			})
			second := canonicalSourceLaneTargetRef(a.IntendedBindings[0])
			anchor := sourceBindingCitation099F(t, facts, "/rest/operations/0/source_operation/responses/201/content/application~1json/schema")
			second.SourceSchema = &anchor
			second.FieldMappings[0].Source = sourceBindingCitation099F(t, facts, anchor.Pointer+"/properties/data/items")
			want := 2
			codes := []string{}
			if mode != "both" {
				want = 1
				codes = []string{"target_required_scope_unverified"}
			}
			if mode != "missing second" {
				if strings.HasPrefix(mode, "bad") {
					second.FieldMappings[0].Source = anchor
					codes = append(codes, "target_record_projection_mismatch")
				}
				a.IntendedBindings = append(a.IntendedBindings, second)
				if mode == "bad second first" {
					a.IntendedBindings[0], a.IntendedBindings[1] = a.IntendedBindings[1], a.IntendedBindings[0]
				}
			}
			sourceBindingOutcome099F(t, key, facts, a, want, codes...)
		})
	}
}

func TestSourceLaneBinding099FSourceSelectors(t *testing.T) {
	for _, mode := range []string{"unrelated equal node", "overridden parameter", "invalid absent target", "local ref occurrence", "external ref", "ambiguous component"} {
		t.Run(mode, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "body", nil)
			r := &a.IntendedBindings[0]
			want := 0
			code := "source_binding_scope_mismatch"
			switch mode {
			case "unrelated equal node":
				facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
					var schema any
					_ = json.Unmarshal([]byte(sourceBindingBody099F), &schema)
					doc["unrelated"] = schema
				})
				c := sourceBindingCitation099F(t, facts, "/unrelated")
				r.SourceSchema = &c
			case "overridden parameter":
				facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
					doc["unrelated"] = map[string]any{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "string"}}
				})
				r.FieldMappings[0].Source = sourceBindingCitation099F(t, facts, "/unrelated")
			case "invalid absent target":
				r.ID = "future"
				r.SourceSchema.ValueSHA256 = strings.Repeat("0", 64)
				code = "source_binding_citation_mismatch"
			case "local ref occurrence", "external ref", "ambiguous component":
				facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
					var schema any
					_ = json.Unmarshal([]byte(sourceBindingBody099F), &schema)
					doc["source_contract"] = map[string]any{"components": map[string]any{"schemas": map[string]any{"Body": schema}}}
					op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
					uri := "#/components/schemas/Body"
					if mode == "external ref" {
						uri = "https://provider.example/schema"
					}
					op["requestBody"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"] = map[string]any{"$ref": uri}
				})
				anchor := sourceBindingCitation099F(t, facts, r.SourceSchema.Pointer)
				r.SourceSchema = &anchor
				r.FieldMappings[1].Source = sourceBindingCitation099F(t, facts, "/source_contract/components/schemas/Body/properties/data")
				if mode == "local ref occurrence" {
					want = 1
					code = ""
				}
				if mode == "external ref" {
					code = "source_schema_unverified"
				}
				if mode == "ambiguous component" {
					facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
						body := doc["source_contract"].(map[string]any)["components"].(map[string]any)["schemas"].(map[string]any)["Body"].(map[string]any)
						var field any
						_ = json.Unmarshal([]byte(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false}`), &field)
						doc["source_contract"].(map[string]any)["components"].(map[string]any)["schemas"].(map[string]any)["Shared"] = field
						props := body["properties"].(map[string]any)
						props["data"] = map[string]any{"$ref": "#/components/schemas/Shared"}
						props["other"] = map[string]any{"$ref": "#/components/schemas/Shared"}
					})
					r.FieldMappings[1].Source = sourceBindingCitation099F(t, facts, "/source_contract/components/schemas/Shared")
					code = "source_projection_ambiguous"
				}
			}
			codes := []string{}
			if code != "" {
				codes = append(codes, code)
			}
			sourceBindingOutcome099F(t, key, facts, a, want, codes...)
		})
	}
}

func TestSourceLaneBinding099FSyncTransport(t *testing.T) {
	for _, mode := range []string{"descriptor without factory", "plan does not register factory", "wrong explicit plan coordinate", "wrong executor ID", "wrong role"} {
		t.Run(mode, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "envelope", func(lock *vNextSourceLock) {
				lock.Lanes["sync_transport"] = "implemented"
				lock.Execution = map[string]json.RawMessage{"sync_transport.json": json.RawMessage(`{"schema_version":1,"source_transport":{"executor":{"family":"declarative_api","id":"fixture_source"},"eligible_streams":["widgets"],"modes":["full_overwrite"],"delivery":{"idempotency":"keyed","ordering":"source_ordered","deletes":"tombstone"}}}`)}
			})
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)["callbacks"] = map[string]any{"changes": map[string]any{}}
			})
			r := &a.IntendedBindings[0]
			r.Kind = "sync_transport"
			r.Lane = "sync_transport"
			r.ID = "fixture_source"
			r.Artifact = "internal/connectors/defs/acme/sync_transport.json"
			r.Pointer = "/source_transport"
			r.ArtifactSHA256 = sourceBytesHash(facts.bindings.Artifacts[r.Artifact])
			r.CanonicalPointer = "/execution/sync_transport.json/source_transport"
			r.SchemaRole = ""
			r.SourceSchema = nil
			r.FieldMappings = nil
			code := "target_transport_executor_unverified"
			if strings.Contains(mode, "plan") {
				descriptor := facts.bindings.Canonical["acme"]
				axes, ok := synccontract.ModeAxes(synccontract.ModeFullOverwrite)
				if !ok {
					t.Fatal("mode contract missing")
				}
				plan := syncplan.Plan{ContractVersion: syncplan.ContractVersion, Source: syncplan.BindingRef{Kind: synccontract.BindingKindStream, ID: "widgets"}, Target: syncplan.BindingRef{Kind: synccontract.BindingKindAction, ID: "destination.write"}, Mode: synccontract.ModeFullOverwrite, Axes: axes, GenerationDigest: descriptor.Staged.Identity.Digest, ArtifactDigest: vNextSemanticAdmissionDigest("a"), EvidenceDigest: vNextSemanticAdmissionDigest("b"), Executors: []syncplan.ExecutorRef{{Role: syncplan.ExecutorRoleSource, ID: descriptor.Staged.Manifest.Executor, Digest: descriptor.Staged.Identity.Digest}, {Role: syncplan.ExecutorRoleDestination, ID: "closed_typed/destination.v1", Digest: vNextSemanticAdmissionDigest("c")}}, Foundation: syncplan.FoundationRef{ID: "authoring.source-lock-vnext.v1", Digest: vNextSemanticAdmissionDigest("d"), Available: true, Reference: "docs/connector-canon/foundations/catalog.json"}}
				plan.Executors[0], plan.Executors[1] = plan.Executors[1], plan.Executors[0]
				coordinate := "/operations/0/stream"
				if mode == "wrong explicit plan coordinate" {
					coordinate = "/operations/9/stream"
					code = "target_transport_coordinate_unverified"
				}
				staged, err := admitVNextCanonicalDescriptor(descriptor, vNextSemanticAdmissionInput{Sync: []vNextSyncAdmission{{SourceID: "stream:widgets", FieldPath: coordinate, Plan: plan}}})
				if err != nil {
					t.Fatalf("real supplied sync admission must succeed: %v", err)
				}
				if len(staged.Sync) != 1 || staged.Sync[0].Result.Kind != syncplan.ResultKindExecutable {
					t.Fatal("actual sync phase not reached")
				}
				descriptor.Staged = staged
				facts.bindings.Canonical["acme"] = descriptor
			}
			if mode == "wrong executor ID" {
				r.ID = "api_engine.v1"
				code = "target_identity_mismatch"
			}
			if mode == "wrong role" {
				r.Pointer = "/destination_transport"
				r.CanonicalPointer = "/execution/sync_transport.json/destination_transport"
				code = "target_absent"
			}
			sourceBindingOutcome099F(t, key, facts, a, 0, code)
		})
	}
}

func TestSourceLaneBinding099FDestinationTransport(t *testing.T) {
	for _, wrong := range []bool{false, true} {
		t.Run(strconv.FormatBool(wrong), func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "body", func(lock *vNextSourceLock) {
				lock.Lanes["sync_transport"] = "implemented"
				lock.Execution = map[string]json.RawMessage{"sync_transport.json": json.RawMessage(`{"schema_version":1,"destination_transport":{"executor":{"family":"declarative_api","id":"fixture_destination"},"eligible_actions":["create_widget"],"modes":["full_append"],"delivery":{"idempotency":"keyed","ordering":"source_ordered","deletes":"tombstone"},"acknowledgement":"durable_warehouse","apply_strategies":[{"mode":"full_append","strategy":"append","action":"create_widget"}]}}`)}
			})
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)["callbacks"] = map[string]any{"changes": map[string]any{}}
			})
			r := &a.IntendedBindings[0]
			r.Kind, r.Lane, r.ID = "sync_transport", "sync_transport", "fixture_destination"
			r.Artifact = "internal/connectors/defs/acme/sync_transport.json"
			r.ArtifactSHA256 = sourceBytesHash(facts.bindings.Artifacts[r.Artifact])
			r.Pointer, r.CanonicalPointer = "/destination_transport", "/execution/sync_transport.json/destination_transport"
			r.SchemaRole, r.SourceSchema, r.FieldMappings = "", nil, nil
			code := "target_transport_executor_unverified"
			if wrong {
				r.ID = "fixture_source"
				code = "target_identity_mismatch"
			}
			sourceBindingOutcome099F(t, key, facts, a, 0, code)
		})
	}
}

func TestSourceLaneBinding099FTemplateControls(t *testing.T) {
	for _, mode := range []string{"matching config record", "swapped fields", "nested config", "missing mapping", "other literal", "unsupported expression"} {
		t.Run(mode, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "body", func(lock *vNextSourceLock) {
				lock.ConfigSchema = json.RawMessage(`{"type":"object","properties":{"workspace_id":{"type":"string"}},"required":["workspace_id"]}`)
				route := "/workspaces/{{ config.workspace_id }}/widgets/{{ record.id }}"
				if mode == "swapped fields" {
					route = "/workspaces/{{ record.id }}/widgets/{{ config.workspace_id }}"
				}
				if mode == "nested config" {
					route = "/workspaces/{{ config.workspace_id.child }}/widgets/{{ record.id }}"
				}
				if mode == "other literal" {
					route = "/other/{{ config.workspace_id }}/widgets/{{ record.id }}"
				}
				if mode == "unsupported expression" {
					route = "/workspaces/{{ config.workspace_id | lower }}/widgets/{{ record.id }}"
				}
				lock.Operations[0].Write = json.RawMessage(strings.Replace(string(lock.Operations[0].Write), "/widgets/{{ record.id }}", route, 1))
			})
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				node := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)
				node["path"] = "/workspaces/{workspace_id}/widgets/{id}"
				op := node["source_operation"].(map[string]any)
				op["parameters"] = append(op["parameters"].([]any), map[string]any{"name": "workspace_id", "in": "path", "required": true, "schema": map[string]any{"type": "string"}})
			})
			pointer := "/properties/workspace_id"
			a.IntendedBindings[0].FieldMappings = append(a.IntendedBindings[0].FieldMappings, sourceLaneFieldMapping{Source: sourceBindingCitation099F(t, facts, "/rest/operations/0/source_operation/parameters/1"), Target: sourceLaneFieldTarget{Kind: sourceLaneFieldConfig, Pointer: &pointer}})
			want := 0
			code := "target_path_projection_mismatch"
			if mode == "matching config record" {
				want = 1
				code = ""
			}
			if mode == "nested config" || mode == "unsupported expression" {
				code = "target_path_projection_unverified"
			}
			if mode == "other literal" {
				code = "target_semantics_mismatch"
			}
			if mode == "missing mapping" {
				a.IntendedBindings[0].FieldMappings = a.IntendedBindings[0].FieldMappings[:2]
			}
			codes := []string{}
			if code != "" {
				codes = append(codes, code)
			}
			sourceBindingOutcome099F(t, key, facts, a, want, codes...)
		})
	}
}

func TestSourceLaneBinding099FSchemaOracle(t *testing.T) {
	for _, tc := range []struct{ name, source, target, want string }{
		{"decimal equivalence", `{"const":1}`, `{"const":1.0}`, ""},
		{"numeric enum equivalence", `{"enum":[1]}`, `{"enum":[1.0]}`, ""},
		{"number versus numeric-looking string", `{"const":1}`, `{"const":"1e0"}`, "target_schema_mismatch"},
		{"unknown composition cannot mask known type", `{"type":"string","allOf":[{}]}`, `{"type":"integer"}`, "target_schema_mismatch"},
		{"external ref remains unknown", `{"$ref":"https://example.invalid/schema"}`, `{"type":"string"}`, "source_schema_unverified"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := sourceLaneSchemaCompare(sourceFacts{}, json.RawMessage(tc.source), json.RawMessage(tc.target))
			if got != tc.want {
				t.Fatalf("literal semantic oracle got=%q want=%q", got, tc.want)
			}
		})
	}
	for _, pointer := range []string{"/properties/a~1b", "/properties/x~0y", "/prefixItems/0", ""} {
		t.Run("coordinate "+pointer, func(t *testing.T) {
			raw := json.RawMessage(`{"type":"object","properties":{"a/b":{"type":"string"},"x~y":{"type":"integer"}},"prefixItems":[{"type":"boolean"}]}`)
			projection, code := sourceLaneTargetProjection(raw, pointer)
			if code != "" || len(projection.Raw) == 0 {
				t.Fatalf("valid literal schema coordinate not resolved: %s", code)
			}
		})
	}
}

func TestSourceLaneBinding099FJSONArray(t *testing.T) {
	key, facts, a := sourceBindingFixture099F(t, "body", func(lock *vNextSourceLock) {
		record := `{"type":"object","properties":{"id":{"type":"string"},"payload":{"type":"array","items":` + sourceBindingRecord099F + `}},"required":["id","payload"],"additionalProperties":false}`
		lock.Schemas["schemas/request.json"] = json.RawMessage(record)
		lock.Operations[0].Write = json.RawMessage(`{"name":"create_widget","kind":"create","method":"POST","path":"/widgets/{{ record.id }}","path_fields":["id"],"body_type":"json_array","body_field":"payload","body_schema":{"type":"array","items":` + sourceBindingRecord099F + `},"record_schema":` + record + `,"risk":"low"}`)
	})
	facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
		op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
		var schema any
		_ = json.Unmarshal([]byte(`{"type":"array","items":`+sourceBindingRecord099F+`}`), &schema)
		op["requestBody"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"] = schema
	})
	r := &a.IntendedBindings[0]
	anchor := sourceBindingCitation099F(t, facts, r.SourceSchema.Pointer)
	r.SourceSchema = &anchor
	r.FieldMappings[1].Source = anchor
	pointer := "/properties/payload"
	r.FieldMappings[1].Target.Pointer = &pointer
	sourceBindingOutcome099F(t, key, facts, a, 1)
}

func TestSourceLaneBinding099FGraphQLVariables(t *testing.T) {
	for _, mode := range []string{"valid", "swapped", "unrelated citation", "linked citation", "ambiguous linked citation"} {
		t.Run(mode, func(t *testing.T) {
			swapped := mode == "swapped"
			key, facts, a := sourceBindingFixture099F(t, "body", func(lock *vNextSourceLock) {
				input := "id"
				if swapped {
					input = "other"
				}
				schema := `{"type":"object","properties":{"id":{"type":"string"},"other":{"type":"string"}},"required":["id"],"additionalProperties":false}`
				lock.Schemas["schemas/request.json"] = json.RawMessage(schema)
				lock.Operations[0].Write = json.RawMessage(`{"name":"create_widget","kind":"create","method":"POST","path":"/graphql","body_type":"graphql","graphql":{"document":"mutation Change($id: String!) { change(id: $id) { id } }","operation_name":"Change","variables":{"id":"{{ record.` + input + ` }}"}},"record_schema":` + schema + `,"risk":"low"}`)
			})
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				node := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)
				node["path"] = "/graphql"
				node["protocol"] = "graphql"
				op := node["source_operation"].(map[string]any)
				delete(op, "parameters")
				delete(op, "requestBody")
				op["operation_name"] = "Change"
				op["document"] = "mutation Change($id: String!) { change(id: $id) { id } }"
				op["variables_schema"] = map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}, "required": []any{"id"}, "additionalProperties": false}
			})
			base := "/rest/operations/0/source_operation/"
			name := sourceBindingCitation099F(t, facts, base+"operation_name")
			document := sourceBindingCitation099F(t, facts, base+"document")
			request := sourceBindingCitation099F(t, facts, base+"variables_schema")
			a.GraphQL = &sourceLaneGraphQLRefs{OperationName: &name, Document: &document, RequestSchema: &request}
			r := &a.IntendedBindings[0]
			r.SourceSchema = &request
			pointer := "/properties/id"
			r.FieldMappings = []sourceLaneFieldMapping{{Source: sourceBindingCitation099F(t, facts, request.Pointer+pointer), Target: sourceLaneFieldTarget{Kind: sourceLaneFieldSchema, Pointer: &pointer}}}
			want := 1
			codes := []string{}
			if swapped {
				want = 0
				codes = []string{"target_graphql_variable_mismatch"}
			}
			if mode == "unrelated citation" {
				facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
					doc["unrelated"] = "Change"
				})
				name = sourceBindingCitation099F(t, facts, "/unrelated")
				a.GraphQL.OperationName = &name
				want = 0
				codes = []string{"source_binding_scope_mismatch"}
			}
			if strings.Contains(mode, "linked citation") {
				facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
					doc["source_contract"] = map[string]any{"contract": map[string]any{"name": "Change"}}
					op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
					op["contract"] = map[string]any{"$ref": "#/contract"}
					if mode == "ambiguous linked citation" {
						op["other_contract"] = map[string]any{"$ref": "#/contract"}
					}
				})
				name = sourceBindingCitation099F(t, facts, "/source_contract/contract/name")
				a.GraphQL.OperationName = &name
				if mode == "ambiguous linked citation" {
					want = 0
					codes = []string{"source_projection_ambiguous"}
				}
			}
			sourceBindingOutcome099F(t, key, facts, a, want, codes...)
		})
	}
}

func TestSourceLaneBinding099FCrossReferenceControls(t *testing.T) {
	for _, mode := range []string{"duplicate", "conflict", "missing required parameter", "wrong equal registry", "wrong authored object"} {
		t.Run(mode, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "body", func(lock *vNextSourceLock) {
				lock.Schemas["schemas/unrelated.json"] = append(json.RawMessage(nil), lock.Schemas["schemas/request.json"]...)
			})
			code := "target_duplicate_claim"
			switch mode {
			case "duplicate", "conflict":
				second := canonicalSourceLaneTargetRef(a.IntendedBindings[0])
				if mode == "conflict" {
					pointer := "/properties/missing"
					second.FieldMappings[1].Target.Pointer = &pointer
					code = "target_projection_conflict"
				}
				a.MaterializedBindings = []sourceLaneTargetRef{second}
			case "missing required parameter":
				facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
					op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
					op["parameters"] = append(op["parameters"].([]any), map[string]any{"name": "scope", "in": "query", "required": true, "schema": map[string]any{"type": "string"}})
				})
				code = "target_parameter_mismatch"
			case "wrong equal registry":
				r := &a.IntendedBindings[0]
				r.Kind = "schema"
				r.ID = "schemas/unrelated.json"
				r.Artifact = "internal/connectors/defs/acme/" + r.ID
				r.Pointer = ""
				r.ArtifactSHA256 = sourceBytesHash(facts.bindings.Artifacts[r.Artifact])
				r.CanonicalPointer = "/operations/0/schema_refs/request"
				code = "canonical_provenance_mismatch"
			case "wrong authored object":
				r := &a.IntendedBindings[0]
				r.Kind = "canonical_operation"
				r.ID = r.CanonicalID
				r.Artifact = "internal/connectors/defs/acme/source.lock.json"
				r.Pointer = "/operations/0"
				r.CanonicalPointer = "/operations/0"
				var lock map[string]any
				_ = json.Unmarshal(facts.bindings.Authoring[r.Artifact], &lock)
				lock["operations"].([]any)[0].(map[string]any)["write"].(map[string]any)["path"] = "/other/{{ record.id }}"
				raw, _ := json.Marshal(lock)
				facts.bindings.Authoring[r.Artifact] = raw
				r.ArtifactSHA256 = sourceBytesHash(raw)
				code = "canonical_authoring_mismatch"
			}
			sourceBindingOutcome099F(t, key, facts, a, 0, code)
		})
	}
}

func TestSourceLaneBinding099FOptionalBodyAndSourceErrors(t *testing.T) {
	for _, mode := range []string{"outer false", "outer absent", "error response", "duplicate field", "missing pointer", "unknown tag"} {
		t.Run(mode, func(t *testing.T) {
			kind := "body"
			if mode == "error response" {
				kind = "envelope"
			}
			key, facts, a := sourceBindingFixture099F(t, kind, func(lock *vNextSourceLock) {
				if strings.HasPrefix(mode, "outer") {
					lock.Operations[0].Write = json.RawMessage(strings.Replace(string(lock.Operations[0].Write), `"body_required":true`, `"body_required":false`, 1))
				}
			})
			want := 0
			code := "target_projection_shape_invalid"
			switch mode {
			case "outer false", "outer absent":
				facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
					body := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)["requestBody"].(map[string]any)
					if mode == "outer absent" {
						delete(body, "required")
					} else {
						body["required"] = false
					}
				})
				want = 1
				code = ""
			case "error response":
				facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
					responses := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)["responses"].(map[string]any)
					var response any
					_ = json.Unmarshal([]byte(`{"content":{"application/json":{"schema":`+sourceBindingEnvelope099F+`}}}`), &response)
					responses["400"] = response
				})
				r := &a.IntendedBindings[0]
				anchor := sourceBindingCitation099F(t, facts, "/rest/operations/0/source_operation/responses/400/content/application~1json/schema")
				r.SourceSchema = &anchor
				r.FieldMappings[0].Source = sourceBindingCitation099F(t, facts, anchor.Pointer+"/properties/data/items")
				code = "source_binding_scope_mismatch"
			case "duplicate field":
				r := &a.IntendedBindings[0]
				r.FieldMappings = append(r.FieldMappings, r.FieldMappings[0])
			case "missing pointer":
				a.IntendedBindings[0].FieldMappings[0].Target.Pointer = nil
			case "unknown tag":
				a.IntendedBindings[0].FieldMappings[0].Target.Kind = "body"
			}
			codes := []string{}
			if code != "" {
				codes = append(codes, code)
			}
			sourceBindingOutcome099F(t, key, facts, a, want, codes...)
		})
	}
}

func TestSourceLaneBinding099FIncompleteCoverage(t *testing.T) {
	for _, mode := range []string{"schema only", "unknown success response"} {
		t.Run(mode, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
			code := "target_required_scope_unverified"
			if mode == "schema only" {
				r := &a.IntendedBindings[0]
				r.Kind = "schema"
				r.ID = "schemas/widgets.json"
				r.Artifact = "internal/connectors/defs/acme/" + r.ID
				r.Pointer = ""
				r.ArtifactSHA256 = sourceBytesHash(facts.bindings.Artifacts[r.Artifact])
				r.CanonicalPointer = "/operations/0/schema_refs/record"
				code = "target_executable_coverage_unverified"
			} else {
				facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
					responses := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)["responses"].(map[string]any)
					responses["201"] = map[string]any{"description": "Unknown successful response"}
				})
			}
			sourceBindingOutcome099F(t, key, facts, a, 1, code)
		})
	}
}

func TestSourceLaneBinding099FAuthoringCustody(t *testing.T) {
	root, cohort, _ := sourceBindingRepositoryFixture(t)
	path := "internal/connectors/defs/acme/source.lock.json"
	before, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		t.Fatal(err)
	}
	inputs := collectSourceLaneBindings(context.Background(), root, cohort)
	if !reflect.DeepEqual(inputs.Authoring[path], before) {
		t.Fatal("collector failed exact separately owned authoring bytes")
	}
	if _, exists := inputs.Artifacts[path]; exists {
		t.Fatal("source lock leaked into execution artifacts")
	}
	if len(inputs.Canonical) != 1 || len(inputs.Bundles) != 1 {
		t.Fatal("authoring observation lost admitted executable counterpart")
	}
}

func TestSourceLaneBindingClaims(t *testing.T) {
	const artifact = "internal/connectors/defs/fixture/operations.json"
	raw := []byte(`{"operations":[{"id":"read_widget","kind":"rest_read","rest":{"method":"GET","path":"/widgets/{id}"}},{"id":"other_widget","kind":"rest_read","rest":{"method":"GET","path":"/other/{id}"}}]}`)
	node := json.RawMessage(`{"id":"provider.read","protocol":"rest","method":"GET","path":"/widgets/{id}","source_operation":{"summary":"Get widget","responses":{"204":{"description":"No content"}}}}`)
	key := sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "provider.read"}
	row := retainedSourceOperation{Key: key, Observed: true, Node: node, Pointer: "/rest/operations/0"}
	doc := retainedSourceDocument{ID: "fixture:primary", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
	for _, tc := range []struct {
		name, id, pointer, hash string
		materialized            bool
		want                    string
	}{
		{"matching observed endpoint", "read_widget", "/operations/0", "", false, "target_contract_unverified"},
		{"existing target without authored pointer", "read_widget", "", "", false, "target_contract_unverified"},
		{"existing target with missing pointer", "other_widget", "/operations/9", "", false, "target_pointer_mismatch"},
		{"absent intended", "future_read", "/operations/9", "", false, "target_absent"},
		{"dangling present", "future_read", "/operations/9", sourceBytesHash(raw), true, "materialized_target_absent"},
		{"existing wrong endpoint", "other_widget", "/operations/1", sourceBytesHash(raw), true, "target_semantics_mismatch"},
		{"wrong identity at valid pointer", "other_widget", "/operations/0", sourceBytesHash(raw), true, "target_identity_mismatch"},
		{"changed target bytes", "read_widget", "/operations/0", "stale", true, "target_digest_mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			facts := normalizeSourceFacts(row, doc, nil)
			facts.bindings = sourceBindingTestInputs(t, map[string][]byte{artifact: raw}, nil)
			ref := sourceLaneTargetRef{Kind: "operation", Connector: "fixture", ID: tc.id, Lane: "direct_read", Artifact: artifact, Pointer: tc.pointer, ArtifactSHA256: tc.hash}
			a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: "Get widget"}
			if tc.materialized {
				a.MaterializedBindings = []sourceLaneTargetRef{ref}
			} else {
				a.IntendedBindings = []sourceLaneTargetRef{ref}
			}
			cells := classifySourceLanes(key, facts, &a)
			cell := requireSourceLane(t, cells, "direct_read", "applicable")
			found := false
			for _, d := range cell.Diagnostics {
				if d.Code == tc.want {
					found = true
				}
			}
			if !found {
				t.Fatalf("reference %s: want diagnostic %s, got %+v", tc.name, tc.want, cell.Diagnostics)
			}
			if len(cell.References) != 0 || cell.State != "mapped_unproven" {
				t.Fatalf("invalid/absent target promoted: %+v", cell)
			}
		})
	}
}

func TestSourceLaneBindingCanonicalPositive(t *testing.T) {
	lock := operationDirectReadLockForSemanticAdmissionTest()
	descriptor, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatal(err)
	}
	artifact := "internal/connectors/defs/acme/operations.json"
	raw := descriptor.Staged.Outputs["operations.json"]
	node := json.RawMessage(`{"id":"provider.widgets.get","protocol":"rest","method":"GET","path":"/widgets","source_operation":{"summary":"Get widgets","responses":{"200":{"description":"Response semantics retained separately"}}}}`)
	key := sourceOperationKey{Connector: "acme", Inventory: "primary", ID: "provider.widgets.get"}
	row := retainedSourceOperation{Key: key, Observed: true, Node: node, Pointer: "/rest/operations/0"}
	doc := retainedSourceDocument{ID: "acme:primary", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
	facts := normalizeSourceFacts(row, doc, nil)
	facts.bindings = sourceBindingTestInputs(t, map[string][]byte{artifact: raw}, map[string]vNextCanonicalDescriptor{"acme": descriptor})
	ref := sourceLaneTargetRef{Kind: "operation", Connector: "acme", ID: "widgets.get", Lane: "direct_read", Artifact: artifact, Pointer: "/operations/0", ArtifactSHA256: sourceBytesHash(raw), CanonicalID: "operation:widgets.get", CanonicalPointer: "/operations/0/operation", Generation: descriptor.Staged.Identity.Digest}
	annotation := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: "Get widgets", IntendedBindings: []sourceLaneTargetRef{ref}}
	cells := classifySourceLanes(key, facts, &annotation)
	cell := requireSourceLane(t, cells, "direct_read", "applicable")
	// Identity/provenance can be observed without promoting unavailable response
	// semantics or behavior. A correct observation must survive that distinction.
	found := false
	for _, d := range cell.Diagnostics {
		if d.Code == "target_response_contract_unverified" {
			found = true
		}
	}
	if !found {
		t.Fatalf("exact admitted positive target not distinguished from missing provenance: %+v", cell.Diagnostics)
	}
	if cell.State != "mapped_unproven" || len(cell.ProofRefs) != 0 {
		t.Fatalf("identity promoted behavior: %+v", cell)
	}
	changed := map[string][]byte{}
	for name, raw := range descriptor.Staged.Outputs {
		changed[name] = raw
	}
	var metadata map[string]any
	if err := json.Unmarshal(changed["metadata.json"], &metadata); err != nil {
		t.Fatal(err)
	}
	metadata["description"] = "Different current generation"
	changed["metadata.json"], err = json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := engine.Load(newVNextExecutionFS("acme", changed), "acme")
	if err != nil {
		t.Fatal(err)
	}
	facts.bindings.Bundles["acme"] = actual
	facts.bindings.Artifacts["internal/connectors/defs/acme/metadata.json"] = changed["metadata.json"]
	drift := requireSourceLane(t, classifySourceLanes(key, facts, &annotation), "direct_read", "applicable")
	found = false
	for _, d := range drift.Diagnostics {
		if d.Code == "execution_generation_mismatch" {
			found = true
		}
	}
	if !found {
		t.Fatalf("other-file generation drift hidden by matching referenced file: %+v", drift.Diagnostics)
	}
}

func TestSourceLaneBindingProviderOperationIdentity(t *testing.T) {
	for _, tc := range []struct{ name, provider, want string }{
		{"exact supplied provider identity", "GetWidgets", "target_response_contract_unverified"},
		{"same route different supplied operation", "GetOtherWidgets", "target_operation_identity_mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lock := operationDirectReadLockForSemanticAdmissionTest()
			lock.Operations[0].Source = json.RawMessage(`{"provider_operation":"` + tc.provider + `","method":"GET","path":"/widgets"}`)
			descriptor, err := canonicalizeVNextSourceLock(lock)
			if err != nil {
				t.Fatal(err)
			}
			key := sourceOperationKey{Connector: "acme", Inventory: "primary", ID: "retained.widgets"}
			node := json.RawMessage(`{"id":"retained.widgets","operation_id":"GetWidgets","method":"GET","path":"/widgets","protocol":"rest","source_operation":{"summary":"Get widgets","responses":{"200":{"description":"Unknown response shape"}}}}`)
			facts := normalizeSourceFacts(retainedSourceOperation{Key: key, Observed: true, Node: node, Pointer: "/rest/operations/0"}, retainedSourceDocument{ID: "acme:primary", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}, nil)
			facts.bindings = sourceBindingTestInputs(t, map[string][]byte{}, map[string]vNextCanonicalDescriptor{"acme": descriptor})
			if len(facts.bindings.Bundles["acme"].Operations) != 1 || descriptor.Staged.Identity.Digest == "" {
				t.Fatal("real admission/load stage not reached")
			}
			artifact := "internal/connectors/defs/acme/operations.json"
			ref := sourceLaneTargetRef{Kind: "operation", Connector: "acme", ID: "widgets.get", Lane: "direct_read", Artifact: artifact, Pointer: "/operations/0", ArtifactSHA256: sourceBytesHash(descriptor.Staged.Outputs["operations.json"]), CanonicalID: "operation:widgets.get", CanonicalPointer: "/operations/0/operation", Generation: descriptor.Staged.Identity.Digest}
			cells := resolveSourceLaneBindings(key, facts, sourceSemanticAnnotation{IntendedBindings: []sourceLaneTargetRef{ref}}, []sourceLaneCell{{Lane: "direct_read", Applicability: "applicable"}})
			found := false
			for _, d := range cells[0].Diagnostics {
				if d.Code == tc.want {
					found = true
				}
			}
			if !found {
				t.Fatalf("admitted target provider identity %q: want %s; got %+v", tc.provider, tc.want, cells[0].Diagnostics)
			}
		})
	}
}

func TestSourceLaneBindingArtifactKindIdentity(t *testing.T) {
	lock := operationDirectReadLockForSemanticAdmissionTest()
	descriptor, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatal(err)
	}
	raw := descriptor.Staged.Outputs["operations.json"]
	for _, tc := range []struct{ name, file, kind, id, want string }{
		{"correct operation artifact", "operations.json", "operation", "widgets.get", ""},
		{"same bytes wrong artifact kind", "writes.json", "operation", "widgets.get", "target_artifact_kind_mismatch"},
		{"same bytes wrong target kind", "operations.json", "stream", "widgets.get", "target_artifact_kind_mismatch"},
		{"same bytes wrong exact ID", "operations.json", "operation", "widgets.other", "target_absent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pointer, code := sourceLaneTargetPointer(raw, sourceLaneTargetRef{Kind: tc.kind, ID: tc.id, Connector: "acme", Artifact: "internal/connectors/defs/acme/" + tc.file})
			if code != tc.want {
				t.Fatalf("existing valid bytes, kind=%s file=%s id=%s: got pointer=%s code=%s, want %s", tc.kind, tc.file, tc.id, pointer, code, tc.want)
			}
			if code == "" && pointer != "/operations/0" {
				t.Fatalf("lost exact target pointer: %s", pointer)
			}
		})
	}
}

func TestSourceLaneGitLabBridge(t *testing.T) {
	for _, tc := range []struct{ name, sourcePath, targetPath, bridge, want string }{
		{"declared exact boundary", "/api/v4/projects/{id}", "/projects/{id}", `{"source_prefix":"/api/v4","connector_prefix":""}`, "target_contract_unverified"},
		{"prefix lookalike", "/api/v40/projects/{id}", "/0/projects/{id}", `{"source_prefix":"/api/v4","connector_prefix":""}`, "target_semantics_mismatch"},
		{"undeclared prefix", "/api/v4/projects/{id}", "/projects/{id}", `null`, "target_semantics_mismatch"},
		{"unrelated prefix declaration", "/private/projects/{id}", "/projects/{id}", `{"source_prefix":"/private","connector_prefix":""}`, "target_semantics_mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			node, err := json.Marshal(map[string]any{"id": "provider.project", "protocol": "rest", "method": "GET", "path": tc.sourcePath, "source_operation": map[string]any{"summary": "Get project", "responses": map[string]any{"204": map[string]any{"description": "No content"}}}})
			if err != nil {
				t.Fatal(err)
			}
			key := sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "provider.project"}
			row := retainedSourceOperation{Key: key, Observed: true, Node: node, Pointer: "/rest/operations/0"}
			doc := retainedSourceDocument{ID: "fixture", Payload: json.RawMessage(`{"rest":{"path_bridge":` + tc.bridge + `,"operations":[` + string(node) + `]}}`)}
			facts := normalizeSourceFacts(row, doc, nil)
			raw, err := json.Marshal(map[string]any{"operations": []any{map[string]any{"id": "project", "kind": "rest_read", "rest": map[string]any{"method": "GET", "path": tc.targetPath}}}})
			if err != nil {
				t.Fatal(err)
			}
			artifact := "internal/connectors/defs/fixture/operations.json"
			facts.bindings = sourceBindingTestInputs(t, map[string][]byte{artifact: raw}, nil)
			ref := sourceLaneTargetRef{Kind: "operation", Connector: "fixture", ID: "project", Lane: "direct_read", Artifact: artifact, Pointer: "/operations/0"}
			a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: "Get project", IntendedBindings: []sourceLaneTargetRef{ref}}
			cells := classifySourceLanes(key, facts, &a)
			cell := requireSourceLane(t, cells, "direct_read", "applicable")
			found := false
			for _, d := range cell.Diagnostics {
				if d.Code == tc.want {
					found = true
				}
			}
			if !found {
				t.Fatalf("bridge %s: want %s got %+v", tc.name, tc.want, cell.Diagnostics)
			}
			if facts.Path != tc.sourcePath {
				t.Fatalf("source route was rewritten: %q", facts.Path)
			}
		})
	}
}

func TestSourceLaneBindingParameterContract(t *testing.T) {
	for _, tc := range []struct {
		name, sourceType string
		sourceRequired   bool
		sourceMaximum    any
		want             string
	}{
		{"matching bounded parameter", "integer", true, 100, "target_response_contract_unverified"},
		{"equivalent exponent bound", "integer", true, json.Number("1e2"), "target_response_contract_unverified"},
		{"equivalent decimal bound", "integer", true, json.Number("100.0"), "target_response_contract_unverified"},
		{"wrong existing type", "string", true, 100, "target_parameter_mismatch"},
		{"wrong requiredness", "integer", false, 100, "target_parameter_mismatch"},
		{"wrong bound", "integer", true, 10, "target_parameter_mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lock := operationDirectReadLockForSemanticAdmissionTest()
			lock.Operations[0].Operation = json.RawMessage(`{"id":"widgets.get","kind":"rest_read","summary":"Get widgets","risk":"low","approval":"none","output_policy":"json_redacted","rest":{"method":"GET","path":"/widgets","max_bytes":1024,"parameters":[{"name":"limit","in":"query","type":"integer","required":true,"minimum":1,"maximum":100}]}}`)
			lock.Operations[0].Commands[0].Command = json.RawMessage(`{"path":"widgets get","summary":"Get widgets","intent":"direct_read","availability":"implemented","operation":"widgets.get","api_surface":[{"method":"GET","path":"/widgets"}],"output_policy":"json_redacted","flags":[{"name":"limit","maps_to":"query.limit","type":"integer","required":true,"minimum":1,"maximum":100}]}`)
			descriptor, err := canonicalizeVNextSourceLock(lock)
			if err != nil {
				t.Fatal(err)
			}
			node, err := json.Marshal(map[string]any{"id": "provider.widgets.get", "protocol": "rest", "method": "GET", "path": "/widgets", "source_operation": map[string]any{"summary": "Get widgets", "parameters": []any{map[string]any{"name": "limit", "in": "query", "required": tc.sourceRequired, "schema": map[string]any{"type": tc.sourceType, "minimum": 1, "maximum": tc.sourceMaximum}}}, "responses": map[string]any{"200": map[string]any{"description": "Unresolved response"}}}})
			if err != nil {
				t.Fatal(err)
			}
			key := sourceOperationKey{Connector: "acme", Inventory: "primary", ID: "provider.widgets.get"}
			row := retainedSourceOperation{Key: key, Observed: true, Node: node, Pointer: "/rest/operations/0"}
			doc := retainedSourceDocument{ID: "acme:primary", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
			facts := normalizeSourceFacts(row, doc, nil)
			artifact := "internal/connectors/defs/acme/operations.json"
			raw := descriptor.Staged.Outputs["operations.json"]
			facts.bindings = sourceBindingTestInputs(t, map[string][]byte{artifact: raw}, map[string]vNextCanonicalDescriptor{"acme": descriptor})
			ref := sourceLaneTargetRef{Kind: "operation", Connector: "acme", ID: "widgets.get", Lane: "direct_read", Artifact: artifact, Pointer: "/operations/0", ArtifactSHA256: sourceBytesHash(raw), CanonicalID: "operation:widgets.get", CanonicalPointer: "/operations/0/operation", Generation: descriptor.Staged.Identity.Digest}
			a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: "Get widgets", IntendedBindings: []sourceLaneTargetRef{ref}}
			cell := requireSourceLane(t, classifySourceLanes(key, facts, &a), "direct_read", "applicable")
			found := false
			for _, d := range cell.Diagnostics {
				if d.Code == tc.want {
					found = true
				}
			}
			if !found {
				t.Fatalf("%s: want %s got %+v", tc.name, tc.want, cell.Diagnostics)
			}
		})
	}
}

func TestSourceLaneBindingNumericOracle(t *testing.T) {
	for _, tc := range []struct {
		a, b         string
		equal, known bool
	}{
		{"9007199254740993", "9007199254740992", false, true},
		{"0.0001", "1e-4", true, true},
		{"-0.0", "0", true, true},
		{"-10", "10", false, true},
		{"1e999999999", "10e999999998", true, true},
		{"null", "0", false, true},
		{`"100"`, "100", false, false},
	} {
		equal, known := sourceLaneNumericBoundEqual([]byte(tc.a), []byte(tc.b))
		if equal != tc.equal || known != tc.known {
			t.Errorf("%s vs %s: got (%v,%v), want (%v,%v)", tc.a, tc.b, equal, known, tc.equal, tc.known)
		}
	}
}

func sourceBindingRepositoryFixture(t *testing.T) (string, sourceLaneCohort, vNextCanonicalDescriptor) {
	t.Helper()
	lock := operationDirectReadLockForSemanticAdmissionTest()
	descriptor, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	outputs := map[string][]byte{}
	for name, raw := range descriptor.Staged.Outputs {
		outputs[name] = raw
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	outputs["source.lock.json"] = raw
	for name, raw := range outputs {
		file := filepath.Join(root, "internal/connectors/defs/acme", name)
		if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root, sourceLaneCohort{Inventories: []sourceInventoryAnchor{{Connector: "acme", Inventory: "primary"}, {Connector: "acme", Inventory: "supplement"}}}, descriptor
}

func sourceBindingFixtureSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(name string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(name)
			if err != nil {
				return err
			}
			result[rel] = "symlink:" + target
			return nil
		}
		raw, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		result[rel] = string(raw)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestSourceLaneBindingCollector(t *testing.T) {
	root, cohort, descriptor := sourceBindingRepositoryFixture(t)
	before := sourceBindingFixtureSnapshot(t, root)
	inputs := collectSourceLaneBindings(context.Background(), root, cohort)
	if len(inputs.Observations) != 0 {
		t.Fatalf("valid canonical/execution inputs rejected: %+v", inputs.Observations)
	}
	if len(inputs.Canonical) != 1 || inputs.Canonical["acme"].Staged.Identity.Digest != descriptor.Staged.Identity.Digest || len(inputs.Bundles) != 1 {
		t.Fatalf("missing exact canonical/execution observation: %+v", inputs.Observations)
	}
	if len(inputs.Artifacts) != len(descriptor.Staged.Outputs) {
		t.Fatalf("collected%d files, want%d closed runtime artifacts", len(inputs.Artifacts), len(descriptor.Staged.Outputs))
	}
	for name := range inputs.Artifacts {
		if strings.Contains(name, "source.lock") {
			t.Fatalf("authoring lock entered runtime inputs: %s", name)
		}
	}
	if after := sourceBindingFixtureSnapshot(t, root); !reflect.DeepEqual(before, after) {
		t.Fatal("read-only collection changed fixture files")
	}
}

func TestSourceLaneBindingCollectorFaults(t *testing.T) {
	for _, which := range []string{"cancel", "malformed authoring", "unsafe artifact"} {
		t.Run(which, func(t *testing.T) {
			root, cohort, _ := sourceBindingRepositoryFixture(t)
			ctx := context.Background()
			switch which {
			case "cancel":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			case "malformed authoring":
				if err := os.WriteFile(filepath.Join(root, "internal/connectors/defs/acme/source.lock.json"), []byte(`{"schema_version":4,"operations":`), 0600); err != nil {
					t.Fatal(err)
				}
			case "unsafe artifact":
				name := filepath.Join(root, "internal/connectors/defs/acme/metadata.json")
				if err := os.Remove(name); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("spec.json", name); err != nil {
					t.Fatal(err)
				}
			}
			before := sourceBindingFixtureSnapshot(t, root)
			inputs := collectSourceLaneBindings(ctx, root, cohort)
			want := "cancelled"
			if which == "malformed authoring" {
				want = "canonical_input_invalid"
			}
			if which == "unsafe artifact" {
				want = "artifact_input_invalid"
			}
			found := false
			for _, o := range inputs.Observations {
				if o.Code == want {
					found = true
				}
			}
			if !found {
				t.Fatalf("fault%s lacked %s: %+v", which, want, inputs.Observations)
			}
			if which == "malformed authoring" && (len(inputs.Canonical) != 0 || len(inputs.Bundles) != 1) {
				t.Fatalf("authoring fault erased runtime observation or created canonical admission")
			}
			if which == "unsafe artifact" && len(inputs.Bundles) != 0 {
				t.Fatal("unsafe artifact became a loaded bundle")
			}
			if which == "cancel" && (len(inputs.Bundles) != 0 || len(inputs.Canonical) != 0 || len(inputs.Artifacts) != 0) {
				t.Fatal("cancelled collection continued to artifact/admission work")
			}
			if after := sourceBindingFixtureSnapshot(t, root); !reflect.DeepEqual(before, after) {
				t.Fatal("failure changed fixture state")
			}
		})
	}
}

func TestSourceLaneBindingCollectorCurrentClosedInputs(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "data/connector-canon/batch1-source-lane-cohort.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cohort sourceLaneCohort
	if err := json.Unmarshal(raw, &cohort); err != nil {
		t.Fatal(err)
	}
	inputs := collectSourceLaneBindings(context.Background(), root, cohort)
	expected := map[string]bool{"asana": true, "gitlab": true, "bitbucket": true, "circleci": true, "dockerhub": true, "jira": true, "notion": true, "sentry": true, "stripe": true, "vercel": true}
	if len(inputs.Bundles) != len(expected) {
		t.Fatalf("closed input bundles%d, want10; observations=%+v", len(inputs.Bundles), inputs.Observations)
	}
	for name := range expected {
		if _, exists := inputs.Bundles[name]; !exists {
			t.Errorf("closed execution input %s missing", name)
		}
	}
	if len(inputs.Canonical) != 2 {
		t.Fatalf("canonical observations%d, want existing Asana/GitLab pair; observations=%+v", len(inputs.Canonical), inputs.Observations)
	}
	for _, name := range []string{"asana", "gitlab"} {
		if _, exists := inputs.Canonical[name]; !exists {
			t.Errorf("canonical observation missing%s", name)
		}
	}
	for _, observation := range inputs.Observations {
		if observation.Code != "canonical_input_unavailable" || observation.Connector == "asana" || observation.Connector == "gitlab" {
			t.Errorf("unexpected current input observation: %+v", observation)
		}
	}
	// These are static input observations only, never executable lane proof.
}

func TestSourceLaneReferenceCounterexample(t *testing.T) {
	for _, tc := range []struct{ name, raw, id, pointer, code string }{
		{"exact operation", `{"operations":[{"id":"a"},{"id":"b"}]}`, "b", "/operations/1", ""},
		{"duplicate operation", `{"operations":[{"id":"b"},{"id":"b"}]}`, "b", "", "target_identity_ambiguous"},
		{"absent operation", `{"operations":[{"id":"a"}]}`, "b", "", "target_absent"},
		{"wrong collection", `{"actions":[{"name":"b"}]}`, "b", "", "target_shape_invalid"},
	} {
		pointer, code := sourceLaneTargetPointer([]byte(tc.raw), sourceLaneTargetRef{Kind: "operation", ID: tc.id})
		if pointer != tc.pointer || code != tc.code {
			t.Errorf("%s: pointer/code=(%s,%s), want(%s,%s)", tc.name, pointer, code, tc.pointer, tc.code)
		}
	}
}

// sourceBindingTestInputs supplies the collector's typed boundary to unit
// tests. Canonical positive cases use actual engine.Load; small refusal
// fixtures explicitly decode only the target collection, without pretending
// that those minimal fixtures are full admitted execution bundles.
func sourceBindingTestInputs(t *testing.T, artifacts map[string][]byte, canonical map[string]vNextCanonicalDescriptor) *sourceLaneBindingInputs {
	t.Helper()
	inputs := &sourceLaneBindingInputs{Artifacts: artifacts, Canonical: canonical, Bundles: map[string]engine.Bundle{}}
	for name, descriptor := range canonical {
		for file, raw := range descriptor.Staged.Outputs {
			key := "internal/connectors/defs/" + name + "/" + file
			if _, exists := inputs.Artifacts[key]; !exists {
				inputs.Artifacts[key] = raw
			}
		}
		bundle, err := engine.Load(newVNextExecutionFS(name, descriptor.Staged.Outputs), name)
		if err != nil {
			t.Fatal(err)
		}
		inputs.Bundles[name] = bundle
	}
	for name, raw := range artifacts {
		parts := strings.Split(name, "/")
		if len(parts) < 5 {
			t.Fatal("invalid test artifact path")
		}
		connector := parts[3]
		if _, exists := canonical[connector]; exists {
			continue
		}
		bundle := inputs.Bundles[connector]
		bundle.Name = connector
		switch filepath.Base(name) {
		case "operations.json":
			var doc struct {
				Operations []engine.OperationSpec `json:"operations"`
			}
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}
			bundle.Operations = doc.Operations
		case "writes.json":
			var doc struct {
				Actions []engine.WriteAction `json:"actions"`
			}
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}
			bundle.Writes = doc.Actions
		case "streams.json":
			var doc struct {
				Streams []engine.StreamSpec `json:"streams"`
			}
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}
			bundle.Streams = doc.Streams
		}
		inputs.Bundles[connector] = bundle
	}
	return inputs
}

func TestSourceLaneBindingDeclaredTargetKinds(t *testing.T) {
	for _, tc := range []struct{ kind, lane, method, summary, file, collection string }{
		{"write", "direct_write", "POST", "Create widget", "writes.json", "actions"},
		{"stream", "etl", "GET", "List widgets", "streams.json", "streams"},
	} {
		for _, wrong := range []bool{false, true} {
			t.Run(tc.kind+"/wrong="+strconv.FormatBool(wrong), func(t *testing.T) {
				key := sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "provider.widgets"}
				node, err := json.Marshal(map[string]any{"id": key.ID, "method": tc.method, "protocol": "rest", "path": "/widgets", "source_operation": map[string]any{"summary": tc.summary, "responses": map[string]any{"200": map[string]any{"content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "array", "items": map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}}}}}}}}})
				if err != nil {
					t.Fatal(err)
				}
				row := retainedSourceOperation{Key: key, Observed: true, Node: node, Pointer: "/rest/operations/0"}
				doc := retainedSourceDocument{ID: "fixture", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
				facts := normalizeSourceFacts(row, doc, nil)
				targetPath := "/widgets"
				if wrong {
					targetPath = "/other"
				}
				target := map[string]any{"name": "widgets", "method": tc.method, "path": targetPath}
				if tc.kind == "write" {
					target["kind"] = "create"
					target["body_type"] = "none"
				} else {
					target["records"] = map[string]any{"path": "."}
				}
				raw, err := json.Marshal(map[string]any{tc.collection: []any{target}})
				if err != nil {
					t.Fatal(err)
				}
				artifact := "internal/connectors/defs/fixture/" + tc.file
				facts.bindings = sourceBindingTestInputs(t, map[string][]byte{artifact: raw}, nil)
				ref := sourceLaneTargetRef{Kind: tc.kind, ID: "widgets", Connector: "fixture", Lane: tc.lane, Artifact: artifact}
				a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: tc.summary, IntendedBindings: []sourceLaneTargetRef{ref}}
				cell := requireSourceLane(t, classifySourceLanes(key, facts, &a), tc.lane, "applicable")
				want := "target_contract_unverified"
				if wrong {
					want = "target_semantics_mismatch"
				}
				found := false
				for _, d := range cell.Diagnostics {
					if d.Code == want {
						found = true
					}
				}
				if !found {
					t.Fatalf("%s target wrong=%v: want%s got%+v", tc.kind, wrong, want, cell.Diagnostics)
				}
			})
		}
	}
}

func TestSourceLaneBindingCommandIdentity(t *testing.T) {
	lock := operationDirectReadLockForSemanticAdmissionTest()
	lock.Operations = append(lock.Operations, vNextOperationDescriptor{ID: "operation:widgets.other", Operation: json.RawMessage(`{"id":"widgets.other","kind":"rest_read","summary":"Get other widgets","risk":"low","approval":"none","output_policy":"json_redacted","rest":{"method":"GET","path":"/other","max_bytes":1024}}`), Commands: []vNextCommandDescriptor{{Order: 1, Command: json.RawMessage(`{"path":"widgets other","summary":"Get other widgets","intent":"direct_read","availability":"implemented","operation":"widgets.other","api_surface":[{"method":"GET","path":"/other"}],"output_policy":"json_redacted","flags":[]}`)}}})
	descriptor, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, id, canonicalID, canonicalPointer, want string }{
		{"matching command", "widgets get", "operation:widgets.get", "/operations/0/commands/0", "target_response_contract_unverified"},
		{"wrong existing command", "widgets other", "operation:widgets.other", "/operations/1/commands/0", "target_semantics_mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			node := json.RawMessage(`{"id":"provider.widgets","method":"GET","path":"/widgets","protocol":"rest","source_operation":{"summary":"Get widgets","responses":{"200":{"description":"Response contract unresolved"}}}}`)
			key := sourceOperationKey{Connector: "acme", Inventory: "primary", ID: "provider.widgets"}
			row := retainedSourceOperation{Key: key, Observed: true, Node: node, Pointer: "/rest/operations/0"}
			doc := retainedSourceDocument{ID: "acme:primary", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
			facts := normalizeSourceFacts(row, doc, nil)
			artifact := "internal/connectors/defs/acme/cli_surface.json"
			raw := descriptor.Staged.Outputs["cli_surface.json"]
			facts.bindings = sourceBindingTestInputs(t, map[string][]byte{artifact: raw}, map[string]vNextCanonicalDescriptor{"acme": descriptor})
			ref := sourceLaneTargetRef{Kind: "command", Connector: "acme", ID: tc.id, Lane: "direct_read", Artifact: artifact, ArtifactSHA256: sourceBytesHash(raw), CanonicalID: tc.canonicalID, CanonicalPointer: tc.canonicalPointer, Generation: descriptor.Staged.Identity.Digest}
			a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: "Get widgets", IntendedBindings: []sourceLaneTargetRef{ref}}
			cell := requireSourceLane(t, classifySourceLanes(key, facts, &a), "direct_read", "applicable")
			found := false
			for _, d := range cell.Diagnostics {
				if d.Code == tc.want {
					found = true
				}
			}
			if !found {
				t.Fatalf("%s: want%s got%+v", tc.name, tc.want, cell.Diagnostics)
			}
		})
	}
}
