package engine

import "testing"

func TestResolveStreamBodyMapOmitsAbsentDeclaredInput(t *testing.T) {
	body := map[string]any{
		"limit": 100,
		"createdAfter": map[string]any{
			"template":         "{{ query.createdAfter }}",
			"omit_when_absent": true,
		},
	}
	withoutInput, err := resolveStreamBodyMap(body, Vars{Query: map[string]string{}})
	if err != nil {
		t.Fatalf("resolve absent body input: %v", err)
	}
	if len(withoutInput) != 1 || withoutInput["limit"] != 100 {
		t.Fatalf("absent body input = %#v, want only source-declared limit", withoutInput)
	}
	withInput, err := resolveStreamBodyMap(body, Vars{Query: map[string]string{"createdAfter": "2026-01-01"}})
	if err != nil {
		t.Fatalf("resolve present body input: %v", err)
	}
	if withInput["createdAfter"] != "2026-01-01" {
		t.Fatalf("present body input = %#v, want declared field", withInput)
	}
}

func TestResolveStreamBodyMapCoercesDeclaredTypedInput(t *testing.T) {
	body := map[string]any{
		"createdAfter": map[string]any{
			"template":         "{{ query.createdAfter }}",
			"omit_when_absent": true,
			"type":             "integer",
		},
		"includeArchived": map[string]any{
			"template":         "{{ query.includeArchived }}",
			"omit_when_absent": true,
			"type":             "boolean",
		},
	}
	resolved, err := resolveStreamBodyMap(body, Vars{Query: map[string]string{"createdAfter": "1700000000", "includeArchived": "true"}})
	if err != nil {
		t.Fatalf("resolve typed body: %v", err)
	}
	if resolved["createdAfter"] != int64(1700000000) || resolved["includeArchived"] != true {
		t.Fatalf("typed body = %#v, want integer and boolean values", resolved)
	}
}
