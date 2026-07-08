package certify

import "testing"

func TestGenerateRecordForGitHubLabelIncludesColor(t *testing.T) {
	schema, err := writeActionRecordSchema("github", "create_label")
	if err != nil {
		t.Fatalf("writeActionRecordSchema: %v", err)
	}
	rec, err := GenerateRecordWithOverrides(schema, "pm-cert-github-test", "12345678", PairingsFor("github")[0].Overrides)
	if err != nil {
		t.Fatalf("GenerateRecordWithOverrides: %v", err)
	}
	if rec["color"] != "ededed" {
		t.Fatalf("color = %#v, want ededed", rec["color"])
	}
}
