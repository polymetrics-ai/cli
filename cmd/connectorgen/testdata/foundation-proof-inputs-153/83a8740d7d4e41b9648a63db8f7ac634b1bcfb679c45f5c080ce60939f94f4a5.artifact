package engine

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

// dynamicAction builds a write action with a declared dynamic-key region over
// the record field "custom_fields". record_schema stays CLOSED
// (additionalProperties:false) — the region is declared in it as an object and
// its interior is validated separately, which is the whole point: tenant keys
// become expressible without opening the schema.
func dynamicAction(spec *DynamicFieldsSpec) WriteAction {
	return WriteAction{
		Name:          "sync_member",
		Kind:          "upsert",
		Method:        http.MethodPost,
		Path:          "/members",
		DynamicFields: spec,
		RecordSchema: json.RawMessage(`{
			"type": "object",
			"additionalProperties": false,
			"required": ["id"],
			"properties": {
				"id": {"type": "string"},
				"custom_fields": {"type": "object"}
			}
		}`),
	}
}

func defaultDynamicSpec() *DynamicFieldsSpec {
	return &DynamicFieldsSpec{
		Field:         "custom_fields",
		KeyPattern:    "^[A-Za-z][A-Za-z0-9_]{0,63}$",
		MaxKeys:       10,
		ValueTypes:    []string{"string", "number", "boolean", "null"},
		MaxValueBytes: 64,
	}
}

// TestDynamicFieldsAcceptsTypedScalars: the declared region accepts scalar
// values under matching keys, and inline target merges them at the body root.
func TestDynamicFieldsAcceptsTypedScalars(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{}`)
	b := newWriteTestBundle(srv, dynamicAction(defaultDynamicSpec()))
	rec := connectors.Record{"id": "m1", "custom_fields": map[string]any{
		"tenantScore": json.Number("42"),
		"tenantTag":   "gold",
		"tenantFlag":  true,
	}}
	if err := writeOneRecord(t, b, "sync_member", connectors.RuntimeConfig{}, rec); err != nil {
		t.Fatalf("Write: %v", err)
	}
	body := cap.json()
	if body["tenantTag"] != "gold" {
		t.Fatalf("tenantTag not merged inline: %v", body)
	}
	if body["tenantFlag"] != true {
		t.Fatalf("tenantFlag not merged inline: %v", body)
	}
	if _, ok := body["custom_fields"]; ok {
		t.Fatalf("inline target must not also send the container field: %v", body)
	}
	if body["id"] != "m1" {
		t.Fatalf("declared field lost: %v", body)
	}
}

// TestDynamicFieldsNestedTarget keeps the region under its own field.
func TestDynamicFieldsNestedTarget(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{}`)
	spec := defaultDynamicSpec()
	spec.Target = "nested"
	b := newWriteTestBundle(srv, dynamicAction(spec))
	rec := connectors.Record{"id": "m1", "custom_fields": map[string]any{"tenantTag": "gold"}}
	if err := writeOneRecord(t, b, "sync_member", connectors.RuntimeConfig{}, rec); err != nil {
		t.Fatalf("Write: %v", err)
	}
	body := cap.json()
	nested, ok := body["custom_fields"].(map[string]any)
	if !ok {
		t.Fatalf("nested target must keep the container: %v", body)
	}
	if nested["tenantTag"] != "gold" {
		t.Fatalf("nested value missing: %v", nested)
	}
}

// TestDynamicFieldsRejectsNestedValue is THE anti-escape-hatch invariant: a
// caller must never be able to inject request structure. Values stay scalar.
func TestDynamicFieldsRejectsNestedValue(t *testing.T) {
	for name, bad := range map[string]any{
		"object": map[string]any{"nested": "x"},
		"array":  []any{"a", "b"},
	} {
		t.Run(name, func(t *testing.T) {
			srv, _ := captureServer(t, http.StatusOK, `{}`)
			b := newWriteTestBundle(srv, dynamicAction(defaultDynamicSpec()))
			rec := connectors.Record{"id": "m1", "custom_fields": map[string]any{"tenantThing": bad}}
			err := writeOneRecord(t, b, "sync_member", connectors.RuntimeConfig{}, rec)
			if err == nil {
				t.Fatal("want hard error for non-scalar dynamic value, got nil")
			}
			if !strings.Contains(err.Error(), "scalar") {
				t.Fatalf("error should name the scalar constraint, got: %v", err)
			}
		})
	}
}

// TestDynamicFieldsRejectsUnmatchedKey: keys are constrained by the
// bundle-declared pattern, which is metadata and never caller input.
func TestDynamicFieldsRejectsUnmatchedKey(t *testing.T) {
	for _, key := range []string{"9leading_digit", "has-dash", "has space", "", strings.Repeat("x", 100)} {
		srv, _ := captureServer(t, http.StatusOK, `{}`)
		b := newWriteTestBundle(srv, dynamicAction(defaultDynamicSpec()))
		rec := connectors.Record{"id": "m1", "custom_fields": map[string]any{key: "v"}}
		if err := writeOneRecord(t, b, "sync_member", connectors.RuntimeConfig{}, rec); err == nil {
			t.Fatalf("key %q must be rejected by key_pattern", key)
		}
	}
}

// TestDynamicFieldsRejectsCollision: a dynamic key can never shadow a
// structural one.
func TestDynamicFieldsRejectsCollision(t *testing.T) {
	t.Run("path_field", func(t *testing.T) {
		srv, _ := captureServer(t, http.StatusOK, `{}`)
		action := dynamicAction(defaultDynamicSpec())
		action.Path = "/members/{{ record.id }}"
		action.PathFields = []string{"id"}
		b := newWriteTestBundle(srv, action)
		rec := connectors.Record{"id": "m1", "custom_fields": map[string]any{"id": "spoofed"}}
		if err := writeOneRecord(t, b, "sync_member", connectors.RuntimeConfig{}, rec); err == nil {
			t.Fatal("dynamic key colliding with a path_field must be rejected")
		}
	})
	t.Run("body_key", func(t *testing.T) {
		srv, _ := captureServer(t, http.StatusOK, `{}`)
		b := newWriteTestBundle(srv, dynamicAction(defaultDynamicSpec()))
		// "id" is already a body key for this action (no path_fields), so a
		// dynamic key of the same name would silently overwrite it.
		rec := connectors.Record{"id": "m1", "custom_fields": map[string]any{"id": "spoofed"}}
		if err := writeOneRecord(t, b, "sync_member", connectors.RuntimeConfig{}, rec); err == nil {
			t.Fatal("dynamic key colliding with a body key must be rejected")
		}
	})
}

// TestDynamicFieldsEnforcesBounds: max_keys and max_value_bytes.
func TestDynamicFieldsEnforcesBounds(t *testing.T) {
	t.Run("max_keys", func(t *testing.T) {
		srv, _ := captureServer(t, http.StatusOK, `{}`)
		spec := defaultDynamicSpec()
		spec.MaxKeys = 2
		b := newWriteTestBundle(srv, dynamicAction(spec))
		rec := connectors.Record{"id": "m1", "custom_fields": map[string]any{"aa": "1", "bb": "2", "cc": "3"}}
		if err := writeOneRecord(t, b, "sync_member", connectors.RuntimeConfig{}, rec); err == nil {
			t.Fatal("want max_keys rejection")
		}
	})
	t.Run("max_value_bytes", func(t *testing.T) {
		srv, _ := captureServer(t, http.StatusOK, `{}`)
		spec := defaultDynamicSpec()
		spec.MaxValueBytes = 4
		b := newWriteTestBundle(srv, dynamicAction(spec))
		rec := connectors.Record{"id": "m1", "custom_fields": map[string]any{"aa": "far too long"}}
		if err := writeOneRecord(t, b, "sync_member", connectors.RuntimeConfig{}, rec); err == nil {
			t.Fatal("want max_value_bytes rejection")
		}
	})
}

// TestDynamicFieldsRejectsDisallowedValueType: value_types is a declared
// allow-list, not a suggestion.
func TestDynamicFieldsRejectsDisallowedValueType(t *testing.T) {
	srv, _ := captureServer(t, http.StatusOK, `{}`)
	spec := defaultDynamicSpec()
	spec.ValueTypes = []string{"string"}
	b := newWriteTestBundle(srv, dynamicAction(spec))
	rec := connectors.Record{"id": "m1", "custom_fields": map[string]any{"aa": true}}
	if err := writeOneRecord(t, b, "sync_member", connectors.RuntimeConfig{}, rec); err == nil {
		t.Fatal("want rejection of a value type outside value_types")
	}
}

// TestDynamicFieldsAbsentUnchanged is the regression guard: an action with no
// dynamic_fields keeps closed-schema behaviour exactly as before.
func TestDynamicFieldsAbsentUnchanged(t *testing.T) {
	srv, cap := captureServer(t, http.StatusOK, `{}`)
	action := dynamicAction(nil)
	b := newWriteTestBundle(srv, action)
	if err := writeOneRecord(t, b, "sync_member", connectors.RuntimeConfig{}, connectors.Record{"id": "m1"}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if cap.json()["id"] != "m1" {
		t.Fatalf("unchanged path broken: %v", cap.json())
	}
	// The closed record_schema must still reject an undeclared field.
	err := writeOneRecord(t, b, "sync_member", connectors.RuntimeConfig{}, connectors.Record{"id": "m1", "surprise": "x"})
	if err == nil {
		t.Fatal("closed record_schema must still reject undeclared fields when dynamic_fields is absent")
	}
}

// TestDynamicFieldsRedactionApplies: redact_fields still covers dynamic values.
func TestDynamicFieldsRedactionApplies(t *testing.T) {
	srv, _ := captureServer(t, http.StatusInternalServerError, `{"error":"boom"}`)
	action := dynamicAction(defaultDynamicSpec())
	action.RedactFields = []string{"custom_fields"}
	b := newWriteTestBundle(srv, action)
	rec := connectors.Record{"id": "m1", "custom_fields": map[string]any{"tenantTag": "supersecretvalue"}}
	err := writeOneRecord(t, b, "sync_member", connectors.RuntimeConfig{}, rec)
	if err == nil {
		t.Fatal("want write error from 500")
	}
	if strings.Contains(err.Error(), "supersecretvalue") {
		t.Fatalf("dynamic value leaked into error text: %v", err)
	}
}

// --- bundle-level validation ------------------------------------------------

func TestDynamicFieldsBundleValidation(t *testing.T) {
	base := func() WriteAction {
		a := dynamicAction(defaultDynamicSpec())
		return a
	}
	for name, mutate := range map[string]func(*WriteAction){
		"empty field":         func(a *WriteAction) { a.DynamicFields.Field = "" },
		"empty key_pattern":   func(a *WriteAction) { a.DynamicFields.KeyPattern = "" },
		"bad key_pattern":     func(a *WriteAction) { a.DynamicFields.KeyPattern = "([" },
		"unknown value_type":  func(a *WriteAction) { a.DynamicFields.ValueTypes = []string{"object"} },
		"unknown target":      func(a *WriteAction) { a.DynamicFields.Target = "headers" },
		"field in path_field": func(a *WriteAction) { a.PathFields = []string{"custom_fields"} },
		"field in body_field": func(a *WriteAction) { a.BodyFields = []string{"custom_fields"} },
		"unsupported body_type": func(a *WriteAction) {
			a.BodyType = "form"
		},
	} {
		t.Run(name, func(t *testing.T) {
			a := base()
			mutate(&a)
			if err := validateWriteBodies([]WriteAction{a}); err == nil {
				t.Fatalf("validateWriteBodies must reject: %s", name)
			}
		})
	}

	t.Run("valid", func(t *testing.T) {
		if err := validateWriteBodies([]WriteAction{base()}); err != nil {
			t.Fatalf("valid dynamic_fields rejected: %v", err)
		}
	})
	t.Run("absent", func(t *testing.T) {
		if err := validateWriteBodies([]WriteAction{dynamicAction(nil)}); err != nil {
			t.Fatalf("absent dynamic_fields rejected: %v", err)
		}
	})
}

// TestWritesSchemaAcceptsDynamicFields pins the schema surface.
func TestWritesSchemaAcceptsDynamicFields(t *testing.T) {
	if !strings.Contains(writesSchemaJSON, "dynamic_fields") {
		t.Fatal("writes.schema.json does not declare dynamic_fields")
	}
	sch, err := CompileSchema(json.RawMessage(writesSchemaJSON))
	if err != nil {
		t.Fatalf("compile writes schema: %v", err)
	}
	doc := map[string]any{"actions": []any{map[string]any{
		"name": "sync_member", "kind": "upsert", "method": "POST", "path": "/members",
		"record_schema": map[string]any{"type": "object"}, "risk": "low",
		"dynamic_fields": map[string]any{
			"field": "custom_fields", "key_pattern": "^[A-Za-z][A-Za-z0-9_]*$",
			"max_keys": 100, "value_types": []any{"string", "number"},
			"max_value_bytes": 4096, "target": "inline",
		},
	}}}
	if err := sch.Validate(doc); err != nil {
		t.Fatalf("writes.schema.json must accept dynamic_fields: %v", err)
	}
}
