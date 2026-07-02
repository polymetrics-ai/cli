package engine

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchemaCompileKeywordMatrix(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{
			name: "type scalar",
			raw:  `{"type":"string"}`,
		},
		{
			name: "type array with null",
			raw:  `{"type":["string","null"]}`,
		},
		{
			name: "required",
			raw:  `{"type":"object","required":["id"],"properties":{"id":{"type":"integer"}}}`,
		},
		{
			name: "properties",
			raw:  `{"type":"object","properties":{"name":{"type":"string"}}}`,
		},
		{
			name: "items",
			raw:  `{"type":"array","items":{"type":"string"}}`,
		},
		{
			name: "enum",
			raw:  `{"type":"string","enum":["a","b"]}`,
		},
		{
			name: "pattern",
			raw:  `{"type":"string","pattern":"^[a-z]+$"}`,
		},
		{
			name: "minProperties",
			raw:  `{"type":"object","minProperties":1}`,
		},
		{
			name: "additionalProperties false",
			raw:  `{"type":"object","additionalProperties":false,"properties":{"a":{"type":"string"}}}`,
		},
		{
			name: "annotations preserved but not enforced",
			raw:  `{"type":"string","format":"date-time","default":"x","title":"t","description":"d","$schema":"http://json-schema.org/draft-07/schema#"}`,
		},
		{
			name:    "unknown keyword is compile error",
			raw:     `{"type":"string","totallyUnknownKeyword":true}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompileSchema(json.RawMessage(tt.raw))
			if tt.wantErr && err == nil {
				t.Fatalf("expected compile error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected compile error: %v", err)
			}
		})
	}
}

func TestSchemaValidateInstances(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		instance  string
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "valid object",
			raw:      `{"type":"object","required":["id"],"properties":{"id":{"type":"integer"}}}`,
			instance: `{"id":1}`,
		},
		{
			name:      "missing required field",
			raw:       `{"type":"object","required":["id"],"properties":{"id":{"type":"integer"}}}`,
			instance:  `{}`,
			wantErr:   true,
			errSubstr: "id",
		},
		{
			name:      "wrong type",
			raw:       `{"type":"object","properties":{"id":{"type":"integer"}}}`,
			instance:  `{"id":"nope"}`,
			wantErr:   true,
			errSubstr: "/id",
		},
		{
			name:     "nullable type union accepts null",
			raw:      `{"type":"object","properties":{"state":{"type":["string","null"]}}}`,
			instance: `{"state":null}`,
		},
		{
			name:      "nullable type union rejects wrong type",
			raw:       `{"type":"object","properties":{"state":{"type":["string","null"]}}}`,
			instance:  `{"state":5}`,
			wantErr:   true,
			errSubstr: "/state",
		},
		{
			name:     "items valid",
			raw:      `{"type":"array","items":{"type":"string"}}`,
			instance: `["a","b"]`,
		},
		{
			name:      "items invalid element",
			raw:       `{"type":"array","items":{"type":"string"}}`,
			instance:  `["a",5]`,
			wantErr:   true,
			errSubstr: "/1",
		},
		{
			name:     "enum valid",
			raw:      `{"type":"string","enum":["a","b"]}`,
			instance: `"a"`,
		},
		{
			name:      "enum invalid",
			raw:       `{"type":"string","enum":["a","b"]}`,
			instance:  `"c"`,
			wantErr:   true,
			errSubstr: "enum",
		},
		{
			name:     "pattern valid",
			raw:      `{"type":"string","pattern":"^[a-z]+$"}`,
			instance: `"abc"`,
		},
		{
			name:      "pattern invalid",
			raw:       `{"type":"string","pattern":"^[a-z]+$"}`,
			instance:  `"ABC"`,
			wantErr:   true,
			errSubstr: "pattern",
		},
		{
			name:     "minProperties valid",
			raw:      `{"type":"object","minProperties":1,"properties":{"a":{"type":"string"}}}`,
			instance: `{"a":"x"}`,
		},
		{
			name:      "minProperties invalid",
			raw:       `{"type":"object","minProperties":1,"properties":{"a":{"type":"string"}}}`,
			instance:  `{}`,
			wantErr:   true,
			errSubstr: "minProperties",
		},
		{
			name:     "additionalProperties false valid",
			raw:      `{"type":"object","additionalProperties":false,"properties":{"a":{"type":"string"}}}`,
			instance: `{"a":"x"}`,
		},
		{
			name:      "additionalProperties false rejects extra",
			raw:       `{"type":"object","additionalProperties":false,"properties":{"a":{"type":"string"}}}`,
			instance:  `{"a":"x","b":"y"}`,
			wantErr:   true,
			errSubstr: "/b",
		},
		{
			name:      "nested path in error",
			raw:       `{"type":"object","properties":{"user":{"type":"object","properties":{"login":{"type":"string"}}}}}`,
			instance:  `{"user":{"login":5}}`,
			wantErr:   true,
			errSubstr: "/user/login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sch, err := CompileSchema(json.RawMessage(tt.raw))
			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			var v any
			if err := json.Unmarshal([]byte(tt.instance), &v); err != nil {
				t.Fatalf("unmarshal instance: %v", err)
			}
			verr := sch.Validate(v)
			if tt.wantErr && verr == nil {
				t.Fatalf("expected validation error, got nil")
			}
			if !tt.wantErr && verr != nil {
				t.Fatalf("unexpected validation error: %v", verr)
			}
			if tt.wantErr && tt.errSubstr != "" && !strings.Contains(verr.Error(), tt.errSubstr) {
				t.Fatalf("error %q does not contain %q", verr.Error(), tt.errSubstr)
			}
		})
	}
}

func TestSchemaSecretKeys(t *testing.T) {
	raw := `{
		"type": "object",
		"properties": {
			"token": {"type": "string", "x-secret": true},
			"repository": {"type": "string"},
			"private_key": {"type": "string", "x-secret": true}
		}
	}`
	sch, err := CompileSchema(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	keys := sch.SecretKeys()
	want := map[string]bool{"token": true, "private_key": true}
	if len(keys) != len(want) {
		t.Fatalf("SecretKeys() = %v, want keys matching %v", keys, want)
	}
	for _, k := range keys {
		if !want[k] {
			t.Fatalf("unexpected secret key %q", k)
		}
	}
}

// TestSchemaDefaults proves Defaults() returns a stringified
// property-name -> default map for every root property that declares a
// JSON Schema "default" annotation (gap-loop cycle-1 item 6, REVIEW-A.md
// C3), and omits properties with no default at all.
func TestSchemaDefaults(t *testing.T) {
	raw := `{
		"type": "object",
		"properties": {
			"base_url": {"type": "string", "default": "https://api.example.com"},
			"max_pages": {"type": "string", "default": "0"},
			"api_key": {"type": "string", "x-secret": true}
		}
	}`
	sch, err := CompileSchema(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	defaults := sch.Defaults()
	if defaults["base_url"] != "https://api.example.com" {
		t.Fatalf("Defaults()[base_url] = %q, want https://api.example.com", defaults["base_url"])
	}
	if defaults["max_pages"] != "0" {
		t.Fatalf("Defaults()[max_pages] = %q, want 0", defaults["max_pages"])
	}
	if _, ok := defaults["api_key"]; ok {
		t.Fatalf("Defaults()[api_key] present, want absent (no default declared)")
	}
}

// TestSchemaDefaultTypeMismatches proves DefaultTypeMismatches() flags a
// property whose "default" value's JSON type does not match its declared
// "type" (gap-loop cycle-1 item 6 validate rule: "default must
// type-check"), and does not flag a well-typed default.
func TestSchemaDefaultTypeMismatches(t *testing.T) {
	raw := `{
		"type": "object",
		"properties": {
			"base_url": {"type": "string", "default": "https://api.example.com"},
			"max_pages": {"type": "integer", "default": "not-a-number"},
			"enabled": {"type": "boolean", "default": "yes"}
		}
	}`
	sch, err := CompileSchema(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	mismatches := sch.DefaultTypeMismatches()
	want := map[string]bool{"max_pages": true, "enabled": true}
	if len(mismatches) != len(want) {
		t.Fatalf("DefaultTypeMismatches() = %v, want keys matching %v", mismatches, want)
	}
	for _, k := range mismatches {
		if !want[k] {
			t.Fatalf("unexpected mismatch key %q", k)
		}
	}
}

func TestSchemaProperties(t *testing.T) {
	raw := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"name": {"type": "string"}
		}
	}`
	sch, err := CompileSchema(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	props := sch.Properties()
	want := map[string]bool{"id": true, "name": true}
	if len(props) != len(want) {
		t.Fatalf("Properties() = %v, want keys matching %v", props, want)
	}
	for _, p := range props {
		if !want[p] {
			t.Fatalf("unexpected property %q", p)
		}
	}
}

func TestStreamSchemaAccessors(t *testing.T) {
	raw := `{
		"type": "object",
		"x-primary-key": ["id"],
		"x-cursor-field": "updated_at",
		"properties": {
			"id": {"type": "integer"},
			"updated_at": {"type": "string"}
		}
	}`
	sch, err := CompileSchema(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	ss := &StreamSchema{Schema: sch, PrimaryKey: sch.PrimaryKeys(), CursorField: sch.CursorFieldName()}
	if len(ss.PrimaryKey) != 1 || ss.PrimaryKey[0] != "id" {
		t.Fatalf("PrimaryKey = %v", ss.PrimaryKey)
	}
	if ss.CursorField != "updated_at" {
		t.Fatalf("CursorField = %q", ss.CursorField)
	}
}

func TestSchemaCompileErrorMessages(t *testing.T) {
	_, err := CompileSchema(json.RawMessage(`{not valid json`))
	if err == nil {
		t.Fatalf("expected error for malformed json")
	}

	_, err = CompileSchema(json.RawMessage(`{"type":"bogus-type"}`))
	if err == nil {
		t.Fatalf("expected error for unknown type value")
	}
}
