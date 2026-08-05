package engine

import (
	"encoding/json"
	"fmt"
	"reflect"
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

func TestSchemaMappingRequirementsUseCompiledComposition(t *testing.T) {
	t.Run("allOf required fields", func(t *testing.T) {
		schema, err := CompileSchema(json.RawMessage(`{
			"type":"object",
			"additionalProperties":false,
			"properties":{"name":{"type":"string"}},
			"allOf":[{"required":["name"]}]
		}`))
		if err != nil {
			t.Fatalf("CompileSchema: %v", err)
		}
		paths, err := schema.RequiredMappingPaths()
		if err != nil {
			t.Fatalf("RequiredMappingPaths: %v", err)
		}
		if !reflect.DeepEqual(paths, []string{"name"}) {
			t.Fatalf("RequiredMappingPaths = %v, want [name]", paths)
		}
		info, err := schema.MappingPath("name")
		if err != nil {
			t.Fatalf("MappingPath(name): %v", err)
		}
		if !reflect.DeepEqual(info.Types, []string{"string"}) {
			t.Fatalf("MappingPath(name).Types = %v, want [string]", info.Types)
		}
	})

	t.Run("unsatisfiable allOf type intersection", func(t *testing.T) {
		schema, err := CompileSchema(json.RawMessage(`{
			"type":"object",
			"additionalProperties":false,
			"required":["value"],
			"properties":{"value":{"allOf":[{"type":"string"},{"type":"integer"}]}}
		}`))
		if err != nil {
			t.Fatalf("CompileSchema: %v", err)
		}
		if _, err := schema.RequiredMappingPaths(); err == nil || !strings.Contains(err.Error(), "unsatisfiable") {
			t.Fatalf("RequiredMappingPaths error = %v, want unsatisfiable allOf error", err)
		}
		if _, err := schema.MappingPath("value"); err == nil || !strings.Contains(err.Error(), "unsatisfiable") {
			t.Fatalf("MappingPath(value) error = %v, want unsatisfiable allOf error", err)
		}
	})

	t.Run("number and integer intersect as integer", func(t *testing.T) {
		schema, err := CompileSchema(json.RawMessage(`{
			"type":"object",
			"additionalProperties":false,
			"required":["value"],
			"properties":{"value":{"allOf":[{"type":"number"},{"type":"integer"}]}}
		}`))
		if err != nil {
			t.Fatalf("CompileSchema: %v", err)
		}
		info, err := schema.MappingPath("value")
		if err != nil {
			t.Fatalf("MappingPath(value): %v", err)
		}
		if !reflect.DeepEqual(info.Types, []string{"integer"}) {
			t.Fatalf("MappingPath(value).Types = %v, want [integer]", info.Types)
		}
	})

	t.Run("incompatible closed allOf objects fail compilation", func(t *testing.T) {
		_, err := CompileSchema(json.RawMessage(`{
			"type":"object",
			"allOf":[
				{"type":"object","additionalProperties":false,"properties":{"first":{"type":"string"}}},
				{"type":"object","additionalProperties":false,"properties":{"last":{"type":"string"}}}
			]
		}`))
		if err == nil || !strings.Contains(err.Error(), "incompatible closed-object allOf") {
			t.Fatalf("CompileSchema error = %v, want incompatible closed-object allOf", err)
		}
	})

	t.Run("request pointers preserve literal dotted properties", func(t *testing.T) {
		schema, err := CompileSchema(json.RawMessage(`{
			"type":"object",
			"additionalProperties":false,
			"required":["a.b","a"],
			"properties":{
				"a.b":{"type":"string"},
				"a":{"type":"object","additionalProperties":false,"required":["b"],"properties":{"b":{"type":"string"}}}
			}
		}`))
		if err != nil {
			t.Fatalf("CompileSchema: %v", err)
		}
		paths, err := schema.RequiredRequestBodyPointers()
		if err != nil {
			t.Fatalf("RequiredRequestBodyPointers: %v", err)
		}
		if !reflect.DeepEqual(paths, []string{"/body/a.b", "/body/a/b"}) {
			t.Fatalf("RequiredRequestBodyPointers = %v", paths)
		}
		for _, pointer := range paths {
			got, err := CanonicalRequestFieldPointer(pointer, schema)
			if err != nil {
				t.Fatalf("CanonicalRequestFieldPointer(%q): %v", pointer, err)
			}
			if got != pointer {
				t.Fatalf("CanonicalRequestFieldPointer(%q) = %q", pointer, got)
			}
		}
	})

	for _, keyword := range []string{"anyOf", "oneOf"} {
		t.Run(keyword+" is explicit", func(t *testing.T) {
			schema, err := CompileSchema(json.RawMessage(fmt.Sprintf(`{
				"type":"object",
				"additionalProperties":false,
				"properties":{"name":{"type":"string"},"id":{"type":"string"}},
				%q:[{"required":["name"]},{"required":["id"]}]
			}`, keyword)))
			if err != nil {
				t.Fatalf("CompileSchema: %v", err)
			}
			if _, err := schema.RequiredMappingPaths(); err == nil || !strings.Contains(err.Error(), keyword) {
				t.Fatalf("RequiredMappingPaths error = %v, want %s alternative error", err, keyword)
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

func TestSchemaValidatesRequestBodyInputTypes(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"properties":{
			"count":{"type":"integer"},
			"names":{"type":"array","items":{"type":"string"}},
			"ids":{"type":"array","items":{"type":"integer"}}
		}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	tests := []struct {
		name      string
		pointer   string
		inputType string
		wantErr   bool
	}{
		{name: "integer", pointer: "/body/count", inputType: "integer"},
		{name: "integer rejects string", pointer: "/body/count", inputType: "string", wantErr: true},
		{name: "string array", pointer: "/body/names", inputType: "string_array"},
		{name: "string array rejects integer items", pointer: "/body/ids", inputType: "string_array", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sch.ValidateRequestBodyInputType(tt.pointer, tt.inputType)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateRequestBodyInputType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSchemaValidatesRequestBodyInputDomains(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"properties":{
			"state":{"type":"string","enum":["open","paused"]},
			"empty_only":{"type":"string","enum":[""]},
			"timestamp":{"type":"string","enum":["not-a-date"]},
			"whole":{"type":"number","enum":[1]},
			"fractional":{"type":"number","enum":[1.5]},
			"ids":{"type":"array","minItems":2,"maxItems":4,"items":{"type":"string"}},
			"required_empty":{"type":"array","maxItems":0,"items":{"type":"string"}},
			"patterned_state":{"type":"string","pattern":"^open$"},
			"required_empty_pattern":{"type":"string","pattern":"^$"},
			"compatible_pattern":{"type":"string","pattern":"^[a-z]+$"},
			"empty_pattern_labels":{"type":"array","items":{"type":"string","pattern":"^$"}},
			"labels":{"type":"array","items":{"type":"string","enum":["alpha"]}},
			"delimited_labels":{"type":"array","items":{"type":"string","enum":["a,b"]}},
			"unsafe_labels":{"type":"array","enum":[["line\nbreak"]],"items":{"type":"string"}},
			"numeric_labels":{"type":"array","items":{"enum":[1]}},
			"composed_numeric_labels":{"allOf":[{"items":{"enum":[1]}}]}
		}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	tests := []struct {
		name    string
		pointer string
		input   RequestBodyInput
		wantErr bool
	}{
		{name: "finite enum rejects partial schema enum overlap", pointer: "/body/state", input: RequestBodyInput{Type: "enum", Values: []string{"open", "closed"}}, wantErr: true},
		{name: "finite enum accepts complete schema enum overlap", pointer: "/body/state", input: RequestBodyInput{Type: "enum", Values: []string{"open", "paused"}}},
		{name: "disjoint enum", pointer: "/body/state", input: RequestBodyInput{Type: "enum", Values: []string{"closed"}}, wantErr: true},
		{name: "required string excludes empty enum", pointer: "/body/empty_only", input: RequestBodyInput{Type: "string", Required: true}, wantErr: true},
		{name: "date-time format excludes invalid enum", pointer: "/body/timestamp", input: RequestBodyInput{Type: "string", Format: "date-time"}, wantErr: true},
		{name: "integer intersects whole number enum", pointer: "/body/whole", input: RequestBodyInput{Type: "integer"}},
		{name: "integer excludes fractional number enum", pointer: "/body/fractional", input: RequestBodyInput{Type: "integer"}, wantErr: true},
		{name: "compatible array bounds", pointer: "/body/ids", input: RequestBodyInput{Type: "string_array", MinItems: 1, MaxItems: 3}},
		{name: "cli maximum below schema minimum", pointer: "/body/ids", input: RequestBodyInput{Type: "string_array", MaxItems: 1}, wantErr: true},
		{name: "cli minimum above schema maximum", pointer: "/body/ids", input: RequestBodyInput{Type: "string_array", MinItems: 5}, wantErr: true},
		{name: "required string array excludes schema maximum zero", pointer: "/body/required_empty", input: RequestBodyInput{Type: "string_array", Required: true}, wantErr: true},
		{name: "finite enum rejects partial schema pattern match", pointer: "/body/patterned_state", input: RequestBodyInput{Type: "enum", Values: []string{"open", "closed"}, Required: true}, wantErr: true},
		{name: "finite enum accepts complete schema pattern match", pointer: "/body/patterned_state", input: RequestBodyInput{Type: "enum", Values: []string{"open"}, Required: true}},
		{name: "finite enum excludes schema pattern", pointer: "/body/patterned_state", input: RequestBodyInput{Type: "enum", Values: []string{"closed"}, Required: true}, wantErr: true},
		{name: "required string excludes empty-only pattern", pointer: "/body/required_empty_pattern", input: RequestBodyInput{Type: "string", Required: true}, wantErr: true},
		{name: "string array excludes empty-only item pattern", pointer: "/body/empty_pattern_labels", input: RequestBodyInput{Type: "string_array"}, wantErr: true},
		{name: "non-enum string accepts compatible pattern", pointer: "/body/compatible_pattern", input: RequestBodyInput{Type: "string", Required: true}},
		{name: "string array intersects item enum", pointer: "/body/labels", input: RequestBodyInput{Type: "string_array"}},
		{name: "string array excludes delimited item enum", pointer: "/body/delimited_labels", input: RequestBodyInput{Type: "string_array"}, wantErr: true},
		{name: "string array excludes unsafe root enum item", pointer: "/body/unsafe_labels", input: RequestBodyInput{Type: "string_array"}, wantErr: true},
		{name: "string array excludes numeric item enum", pointer: "/body/numeric_labels", input: RequestBodyInput{Type: "string_array"}, wantErr: true},
		{name: "string array excludes composed numeric item enum", pointer: "/body/composed_numeric_labels", input: RequestBodyInput{Type: "string_array"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sch.ValidateRequestBodyInput(tt.pointer, tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateRequestBodyInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRequestFieldPointerAssignmentsRejectsOverwrites(t *testing.T) {
	tests := []struct {
		name     string
		pointers []string
	}{
		{name: "duplicate path", pointers: []string{"/path/id", "/path/id"}},
		{name: "duplicate query", pointers: []string{"/query/filter", "/query/filter"}},
		{name: "duplicate body", pointers: []string{"/body/name", "/body/name"}},
		{name: "body parent child", pointers: []string{"/body/config", "/body/config/value"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateRequestFieldPointerAssignments(tt.pointers); err == nil {
				t.Fatal("ValidateRequestFieldPointerAssignments() error = nil")
			}
		})
	}
}

func TestSchemaValidatesEffectiveRequestBodyAssembly(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		static   map[string]any
		required []string
		optional []string
		wantErr  bool
	}{
		{
			name:     "required object override replaces static scalar",
			raw:      `{"type":"object","additionalProperties":false,"required":["config"],"properties":{"config":{"type":"object","additionalProperties":false,"required":["value"],"properties":{"value":{"type":"string"}}}}}`,
			static:   map[string]any{"config": "legacy"},
			required: []string{"/body/config/value"},
		},
		{
			name:     "dynamic array replacement must retain required siblings",
			raw:      `{"type":"object","additionalProperties":false,"required":["items"],"properties":{"items":{"type":"array","items":{"type":"object","additionalProperties":false,"required":["name","kind"],"properties":{"name":{"type":"string"},"kind":{"type":"string"}}}}}}`,
			static:   map[string]any{"items": []any{map[string]any{"name": "static", "kind": "widget"}}},
			required: []string{"/body/items/0/name"},
			wantErr:  true,
		},
		{
			name:     "required sparse array mapping",
			raw:      `{"type":"object","additionalProperties":false,"properties":{"items":{"type":"array","items":{"type":"object","additionalProperties":false,"properties":{"name":{"type":"string"}}}}}}`,
			required: []string{"/body/items/1/name"},
			wantErr:  true,
		},
		{
			name:     "optional sparse array mapping",
			raw:      `{"type":"object","additionalProperties":false,"properties":{"items":{"type":"array","items":{"type":"object","additionalProperties":false,"properties":{"name":{"type":"string"}}}}}}`,
			optional: []string{"/body/items/1/name"},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sch, err := CompileSchema(json.RawMessage(tt.raw))
			if err != nil {
				t.Fatalf("CompileSchema: %v", err)
			}
			err = sch.ValidateEffectiveRequestBody(tt.static, tt.required, tt.optional)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateEffectiveRequestBody() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// --- array cardinality (minItems/maxItems) ---------------------------------
//
// The engine dialect had no way to say "this array must not be empty", which
// is why 25 Airtable operations sat blocked behind
// airtable-array-cardinality-foundation, and why drip/writes.json and
// zoho-bigin/writes.json each ship a written apology for the same gap.

func TestCompileSchemaArrayCardinalityKeywords(t *testing.T) {
	for _, raw := range []string{
		`{"type":"array","minItems":1}`,
		`{"type":"array","maxItems":100}`,
		`{"type":"array","minItems":1,"maxItems":100}`,
		`{"type":"object","properties":{"ids":{"type":"array","minItems":1}}}`,
	} {
		if _, err := CompileSchema(json.RawMessage(raw)); err != nil {
			t.Fatalf("CompileSchema(%s): unexpected error: %v", raw, err)
		}
	}
}

func TestSchemaValidateMinItems(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{"type":"object","required":["ids"],"properties":{"ids":{"type":"array","minItems":1,"items":{"type":"string"}}}}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}

	if err := sch.Validate(map[string]any{"ids": []any{"a"}}); err != nil {
		t.Fatalf("one element: unexpected error: %v", err)
	}
	err = sch.Validate(map[string]any{"ids": []any{}})
	if err == nil {
		t.Fatal("empty array: want error, got nil")
	}
	if !strings.Contains(err.Error(), "minItems") || !strings.Contains(err.Error(), "/ids") {
		t.Fatalf("empty array: error should name minItems and the path, got %v", err)
	}

	// minItems governs array CONTENT, never presence: "required and non-empty"
	// is required + minItems, exactly as in real draft-07. A missing optional
	// array must stay valid or every existing bundle silently changes meaning.
	optional, err := CompileSchema(json.RawMessage(`{"type":"object","properties":{"ids":{"type":"array","minItems":1}}}`))
	if err != nil {
		t.Fatalf("CompileSchema optional: %v", err)
	}
	if err := optional.Validate(map[string]any{}); err != nil {
		t.Fatalf("absent optional array: want valid, got %v", err)
	}
}

func TestSchemaValidateMaxItems(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{"type":"array","maxItems":2}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	if err := sch.Validate([]any{"a", "b"}); err != nil {
		t.Fatalf("at limit: unexpected error: %v", err)
	}
	err = sch.Validate([]any{"a", "b", "c"})
	if err == nil {
		t.Fatal("over limit: want error, got nil")
	}
	if !strings.Contains(err.Error(), "maxItems") {
		t.Fatalf("over limit: error should name maxItems, got %v", err)
	}
}

func TestSchemaArrayCardinalityIgnoresNonArrays(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{"minItems":1,"maxItems":2}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	for _, v := range []any{"string", json.Number("7"), true, nil, map[string]any{}} {
		if err := sch.Validate(v); err != nil {
			t.Fatalf("non-array %v: want ignored, got %v", v, err)
		}
	}
}

func TestSchemaMinItemsZeroIsHonored(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{"type":"array","minItems":0}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	if err := sch.Validate([]any{}); err != nil {
		t.Fatalf("explicit minItems 0: want empty array valid, got %v", err)
	}
}

func TestCompileSchemaRejectsInvalidArrayCardinality(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "negative minItems", raw: `{"type":"array","minItems":-1}`, want: "minItems"},
		{name: "negative maxItems", raw: `{"type":"array","maxItems":-1}`, want: "maxItems"},
		{name: "non integer minItems", raw: `{"type":"array","minItems":"1"}`, want: "minItems"},
		{name: "non integer maxItems", raw: `{"type":"array","maxItems":true}`, want: "maxItems"},
		{name: "maxItems below minItems", raw: `{"type":"array","minItems":3,"maxItems":2}`, want: "maxItems"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompileSchema(json.RawMessage(tt.raw))
			if err == nil {
				t.Fatalf("want compile error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error should mention %q, got %v", tt.want, err)
			}
		})
	}
}
