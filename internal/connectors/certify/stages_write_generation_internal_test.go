package certify

import (
	"encoding/json"
	"testing"
)

func TestGenerateRecordForGitHubLabelPreservesOptionalColorOverride(t *testing.T) {
	schema, err := writeActionRecordSchema("github", "create_label")
	if err != nil {
		t.Fatalf("writeActionRecordSchema: %v", err)
	}
	var doc schemaDoc
	if err := json.Unmarshal(schema, &doc); err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	if len(doc.Required) != 1 || doc.Required[0] != "name" {
		t.Fatalf("create_label required fields = %v, want exactly locked-provider required field [name]", doc.Required)
	}
	if containsString(doc.Required, "color") {
		t.Fatalf("create_label required fields = %v, want color optional as declared by the locked provider source", doc.Required)
	}
	if doc.Properties["color"].Type != "string" {
		t.Fatalf("create_label color schema = %#v, want declared optional string", doc.Properties["color"])
	}
	var bounded struct {
		Properties map[string]struct {
			MaxLength int `json:"maxLength"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(schema, &bounded); err != nil {
		t.Fatalf("parse bounded schema: %v", err)
	}
	if got := bounded.Properties["color"].MaxLength; got != 8192 {
		t.Fatalf("create_label color maxLength = %d, want locked declaration bound 8192", got)
	}
	rec, err := GenerateRecordWithOverrides(schema, "pm-cert-github-test", "12345678", PairingsFor("github")[0].Overrides)
	if err != nil {
		t.Fatalf("GenerateRecordWithOverrides: %v", err)
	}
	if rec["color"] != "ededed" {
		t.Fatalf("color = %#v, want ededed", rec["color"])
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
