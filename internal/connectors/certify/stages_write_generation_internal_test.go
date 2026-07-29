package certify

import (
	"encoding/json"
	"testing"
)

func TestGenerateRecordForGitHubLabelIncludesColor(t *testing.T) {
	schema, err := writeActionRecordSchema("github", "create_label")
	if err != nil {
		t.Fatalf("writeActionRecordSchema: %v", err)
	}
	var doc schemaDoc
	if err := json.Unmarshal(schema, &doc); err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	if !containsString(doc.Required, "color") {
		t.Fatalf("create_label required fields = %v, want color from defs/github/writes.json", doc.Required)
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
