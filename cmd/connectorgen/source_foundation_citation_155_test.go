package main

import (
	"encoding/json"
	"testing"
)

func TestSourceFoundationSelectedGlobalCitation155(t *testing.T) {
	repo, _ := sourceFoundationUniverseFixture(t)
	raw := []byte(`{"schema_version":2,"connector":"acme","source_contract":{"components":{"securitySchemes":{"Chosen/~Key":{"type":"apiKey","in":"header","name":"X-API-Key"},"Unrelated":{"type":"apiKey","in":"header","name":"X-Other"}}}},"rest":{"operations":[{"id":"source.a","operation_id":"createWidget","protocol":"rest","method":"post","path":"/widgets","source_operation":{"summary":"Create widget","security":[{"Chosen/~Key":[]}],"responses":{"204":{"description":"No content"}}}},{"id":"source.b","operation_id":"other","protocol":"rest","method":"get","path":"/other","source_operation":{"summary":"Other","responses":{"204":{"description":"No content"}}}}]},"counts":{"total":2}}`)
	proofWrite(t, repo, "source.json", raw)
	_, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	cohort.Inventories[0].Connector = "acme"
	cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
	manifest, err := buildSourceLaneManifest(t.Context(), repo, cohort, nil)
	if err != nil || manifest.Validation.Status != "valid" {
		t.Fatalf("actual retained source control: %v %+v", err, manifest.Validation)
	}
	facts := manifest.SourceOperations[0].Facts
	group := facts.Refs["security_schemes"]
	var document retainedSourceDocument
	for _, d := range manifest.Documents {
		if d.ID == group.DocumentID {
			document = d
		}
	}
	if !sourceFoundationRequirementCitation(facts, document, group) {
		t.Fatal("normalized group control refused")
	}
	cite := func(pointer string) sourceFactRef {
		t.Helper()
		r := group
		r.Pointer = pointer
		value, err := sourceJSONPointer(raw, pointer)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		canonical, err := canonicalSourceJSON(encoded)
		if err != nil {
			t.Fatal(err)
		}
		r.ValueSHA256 = sourceBytesHash(canonical)
		return r
	}
	selected := cite(group.Pointer + "/Chosen~1~0Key")
	cases := []struct {
		name string
		ref  sourceFactRef
		want bool
	}{
		{"exact_referenced_escaped_scheme", selected, true},
		{"unreferenced_scheme", cite(group.Pointer + "/Unrelated"), false},
		{"scheme_property", cite(selected.Pointer + "/name"), false},
		{"normalized_group", group, true},
	}
	wrong := selected
	wrong.DocumentID = "wrong-document"
	cases = append(cases, struct {
		name string
		ref  sourceFactRef
		want bool
	}{"wrong_document", wrong, false})
	wrong = selected
	wrong.ValueSHA256 = sourceBytesHash([]byte("wrong"))
	cases = append(cases, struct {
		name string
		ref  sourceFactRef
		want bool
	}{"wrong_hash", wrong, false})
	wrong = selected
	wrong.Section = "wrong-section"
	cases = append(cases, struct {
		name string
		ref  sourceFactRef
		want bool
	}{"wrong_section", wrong, false})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sourceFoundationRequirementCitation(facts, document, tc.ref); got != tc.want {
				t.Fatalf("exact source citation admission=%t want=%t", got, tc.want)
			}
		})
	}
}
