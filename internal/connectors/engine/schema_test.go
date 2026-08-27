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
			name: "OpenAPI example annotation compiles",
			raw:  `{"type":"string","format":"date-time","default":"x","example":"urn:ietf:params:scim:api:messages:2.0:ListResponse","title":"t","description":"d","$schema":"http://json-schema.org/draft-07/schema#"}`,
		},
		{
			name: "pattern properties compile",
			raw:  `{"type":"object","additionalProperties":false,"patternProperties":{"^x-[a-z]+$":{"type":"integer"}}}`,
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
			name:     "absolute URI valid",
			raw:      `{"type":"string","format":"uri"}`,
			instance: `"https://example.invalid/calendar?channel=1#watch"`,
		},
		{
			name:      "relative URI invalid",
			raw:       `{"type":"string","format":"uri"}`,
			instance:  `"calendar/watch"`,
			wantErr:   true,
			errSubstr: "format",
		},
		{
			name:      "malformed URI escape invalid",
			raw:       `{"type":"string","format":"uri"}`,
			instance:  `"https://example.invalid/%ZZ"`,
			wantErr:   true,
			errSubstr: "format",
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
			name:     "pattern properties validates matching provider metadata",
			raw:      `{"type":"object","additionalProperties":false,"patternProperties":{"^x-[a-z]+$":{"type":"integer"}}}`,
			instance: `{"x-retries":3}`,
		},
		{
			name:      "pattern properties reject a matching value with the wrong type",
			raw:       `{"type":"object","additionalProperties":false,"patternProperties":{"^x-[a-z]+$":{"type":"integer"}}}`,
			instance:  `{"x-retries":"three"}`,
			wantErr:   true,
			errSubstr: "/x-retries",
		},
		{
			name:      "pattern properties do not permit unmatched extra fields",
			raw:       `{"type":"object","additionalProperties":false,"patternProperties":{"^x-[a-z]+$":{"type":"integer"}}}`,
			instance:  `{"retries":3}`,
			wantErr:   true,
			errSubstr: "/retries",
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

func TestSchemaClosedCompositionValidation(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		instance  string
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "nullable scalar oneOf",
			raw:      `{"oneOf":[{"type":"string","minLength":1},{"type":"null"}]}`,
			instance: `null`,
		},
		{
			name:     "nested anyOf selects a source-backed scalar arm",
			raw:      `{"type":"object","additionalProperties":false,"required":["selector"],"properties":{"selector":{"anyOf":[{"type":"string","enum":["name"]},{"type":"integer","minimum":1}]}}}`,
			instance: `{"selector":1}`,
		},
		{
			name:     "compatible allOf intersection",
			raw:      `{"allOf":[{"type":"string","minLength":3},{"type":"string","pattern":"^[a-z]+$"}]}`,
			instance: `"alpha"`,
		},
		{
			name:     "discriminated object selects exactly one alternative",
			raw:      `{"oneOf":[{"type":"object","additionalProperties":false,"required":["kind","radius"],"properties":{"kind":{"type":"string","enum":["circle"]},"radius":{"type":"number","minimum":0}}},{"type":"object","additionalProperties":false,"required":["kind","side"],"properties":{"kind":{"type":"string","enum":["square"]},"side":{"type":"number","minimum":0}}}]}`,
			instance: `{"kind":"circle","radius":2}`,
		},
		{
			name:      "oneOf ambiguous selection fails locally",
			raw:       `{"oneOf":[{"type":"string"},{"type":"string","minLength":1}]}`,
			instance:  `"x"`,
			wantErr:   true,
			errSubstr: "oneOf",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			schema, err := CompileSchema(json.RawMessage(testCase.raw))
			if err != nil {
				t.Fatalf("CompileSchema: %v", err)
			}
			var value any
			if err := json.Unmarshal([]byte(testCase.instance), &value); err != nil {
				t.Fatalf("decode instance: %v", err)
			}
			err = schema.Validate(value)
			if testCase.wantErr {
				if err == nil || !strings.Contains(err.Error(), testCase.errSubstr) {
					t.Fatalf("Validate() error = %v, want %q", err, testCase.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Validate(): %v", err)
			}
		})
	}
}

func TestSchemaCompileRejectsMalformedOrDuplicateComposition(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "empty anyOf", raw: `{"anyOf":[]}`, want: "anyOf"},
		{name: "duplicate oneOf alternatives", raw: `{"oneOf":[{"type":"string"},{"type":"string"}]}`, want: "duplicate"},
		{name: "duplicate anyOf alternatives", raw: `{"anyOf":[{"type":"integer"},{"type":"integer"}]}`, want: "duplicate"},
		{name: "contradictory enum intersection", raw: `{"allOf":[{"type":"string","enum":["a"]},{"type":"string","enum":["b"]}]}`, want: "allOf"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := CompileSchema(json.RawMessage(testCase.raw))
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("CompileSchema() error = %v, want %q", err, testCase.want)
			}
		})
	}
}

func TestSchemaEnumExactNumbersBeyondFloatPrecision(t *testing.T) {
	schema, err := CompileSchema(json.RawMessage(`{"type":"integer","enum":[9007199254740992]}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	if err := schema.Validate(json.Number("9007199254740992.0")); err != nil {
		t.Fatalf("equivalent exact number rejected: %v", err)
	}
	if err := schema.Validate(json.Number("9007199254740993")); err == nil {
		t.Fatal("distinct integer above 2^53 matched enum through float rounding")
	}

	exponent, err := CompileSchema(json.RawMessage(`{"type":"number","enum":[1e3]}`))
	if err != nil {
		t.Fatalf("CompileSchema exponent: %v", err)
	}
	if err := exponent.Validate(json.Number("1000.0")); err != nil {
		t.Fatalf("equivalent exponent rejected: %v", err)
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

func TestSchemaValidateMaxProperties(t *testing.T) {
	sch, err := CompileSchema(json.RawMessage(`{"type":"object","maxProperties":2}`))
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}
	if err := sch.Validate(map[string]any{"one": true, "two": true}); err != nil {
		t.Fatalf("at limit: unexpected error: %v", err)
	}
	err = sch.Validate(map[string]any{"one": true, "two": true, "three": true})
	if err == nil {
		t.Fatal("over limit: want error, got nil")
	}
	if !strings.Contains(err.Error(), "maxProperties") {
		t.Fatalf("over limit: error should name maxProperties, got %v", err)
	}
}

func TestCompileSchemaRejectsPrefixItemsOutsideStructuredREST(t *testing.T) {
	_, err := CompileSchema(json.RawMessage(`{"type":"array","prefixItems":[{"type":"string"}]}`))
	if err == nil || !strings.Contains(err.Error(), "unknown keyword") {
		t.Fatalf("CompileSchema error = %v, want prefixItems rejection", err)
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
