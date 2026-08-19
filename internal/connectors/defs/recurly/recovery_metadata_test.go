// Package recurly holds connector-local regression evidence for the recovered
// Recurly bundle.
package recurly

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type recoveryStream struct {
	Name           string            `json:"name"`
	Query          map[string]string `json:"query"`
	Projection     string            `json:"projection"`
	ComputedFields map[string]string `json:"computed_fields"`
	Schema         string            `json:"schema"`
}

type recoverySchema struct {
	PrimaryKey []string                   `json:"x-primary-key"`
	Cursor     string                     `json:"x-cursor-field"`
	Required   []string                   `json:"required"`
	Properties map[string]json.RawMessage `json:"properties"`
}

type legacyStreamExpectation struct {
	Cursor         string
	ComputedFields map[string]string
	PropertyTypes  map[string]any
	FixtureChecks  map[string]any
}

func TestRecoveredLegacyStreamMetadataPreserved(t *testing.T) {
	t.Parallel()

	streams := loadRecoveryStreams(t)
	expectations := map[string]legacyStreamExpectation{
		"accounts": {
			Cursor: "updated_at",
			ComputedFields: map[string]string{
				"id": "{{ record.id }}", "code": "{{ record.code }}", "email": "{{ record.email }}",
				"state": "{{ record.state }}", "created_at": "{{ record.created_at }}", "updated_at": "{{ record.updated_at }}",
			},
			PropertyTypes: map[string]any{"id": "string", "email": []any{"string", "null"}},
			FixtureChecks: map[string]any{"email": "fixture1@example.com"},
		},
		"invoices": {
			Cursor: "created_at",
			ComputedFields: map[string]string{
				"id": "{{ record.id }}", "account_id": "{{ record.account.id }}", "state": "{{ record.state }}",
				"total": "{{ record.total }}", "created_at": "{{ record.created_at }}",
			},
			PropertyTypes: map[string]any{"id": "string", "account_id": []any{"string", "null"}, "total": []any{"number", "null"}},
			FixtureChecks: map[string]any{"account.id": "acct_fixture_1", "total": 100.0},
		},
		"plans": {
			Cursor: "updated_at",
			ComputedFields: map[string]string{
				"id": "{{ record.id }}", "code": "{{ record.code }}", "name": "{{ record.name }}",
				"state": "{{ record.state }}", "updated_at": "{{ record.updated_at }}",
			},
			PropertyTypes: map[string]any{"id": "string", "name": []any{"string", "null"}},
			FixtureChecks: map[string]any{"name": "Fixture Plan 1"},
		},
		"subscriptions": {
			Cursor: "updated_at",
			ComputedFields: map[string]string{
				"id": "{{ record.id }}", "account_id": "{{ record.account.id }}", "plan_id": "{{ record.plan.id }}",
				"state": "{{ record.state }}", "created_at": "{{ record.created_at }}", "updated_at": "{{ record.updated_at }}",
			},
			PropertyTypes: map[string]any{"id": "string", "account_id": []any{"string", "null"}, "plan_id": []any{"string", "null"}},
			FixtureChecks: map[string]any{"account.id": "acct_fixture_1", "plan.id": "plan_fixture_1"},
		},
		"transactions": {
			Cursor: "created_at",
			ComputedFields: map[string]string{
				"id": "{{ record.id }}", "account_id": "{{ record.account.id }}", "status": "{{ record.status }}",
				"amount": "{{ record.amount }}", "created_at": "{{ record.created_at }}",
			},
			PropertyTypes: map[string]any{"id": "string", "account_id": []any{"string", "null"}, "amount": []any{"number", "null"}},
			FixtureChecks: map[string]any{"account.id": "acct_fixture_1", "amount": 100.0},
		},
	}

	for name, want := range expectations {
		stream, ok := streams[name]
		if !ok {
			t.Errorf("stream %q is missing", name)
			continue
		}
		if got := stream.Query["limit"]; got != "200" {
			t.Errorf("stream %q query.limit = %q, want %q", name, got, "200")
		}
		if stream.Projection != "" {
			t.Errorf("stream %q projection = %q, want schema default", name, stream.Projection)
		}
		if !reflect.DeepEqual(stream.ComputedFields, want.ComputedFields) {
			t.Errorf("stream %q computed_fields = %#v, want %#v", name, stream.ComputedFields, want.ComputedFields)
		}

		schema := loadRecoverySchema(t, stream.Schema)
		if !reflect.DeepEqual(schema.PrimaryKey, []string{"id"}) {
			t.Errorf("schema %q x-primary-key = %v, want [id]", stream.Schema, schema.PrimaryKey)
		}
		if schema.Cursor != want.Cursor {
			t.Errorf("schema %q x-cursor-field = %q, want %q", stream.Schema, schema.Cursor, want.Cursor)
		}
		if !contains(schema.Required, "id") {
			t.Errorf("schema %q required = %v, want id", stream.Schema, schema.Required)
		}
		for property, wantType := range want.PropertyTypes {
			gotType := schemaPropertyType(t, schema.Properties[property])
			if !reflect.DeepEqual(gotType, wantType) {
				t.Errorf("schema %q property %q type = %#v, want %#v", stream.Schema, property, gotType, wantType)
			}
		}

		fixture := loadRecoveryFixture(t, name)
		for path, wantValue := range want.FixtureChecks {
			if gotValue := valueAtPath(fixture, path); !reflect.DeepEqual(gotValue, wantValue) {
				t.Errorf("fixture %q path %q = %#v, want %#v", name, path, gotValue, wantValue)
			}
		}
	}
}

func TestRecoveredConnectionSpecPreservesBaseURLFormat(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("spec.json")
	if err != nil {
		t.Fatalf("read spec.json: %v", err)
	}
	var spec struct {
		Properties map[string]struct {
			Format string `json:"format"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(content, &spec); err != nil {
		t.Fatalf("parse spec.json: %v", err)
	}
	if got := spec.Properties["base_url"].Format; got != "uri" {
		t.Errorf("spec base_url format = %q, want %q", got, "uri")
	}
}

func loadRecoveryStreams(t *testing.T) map[string]recoveryStream {
	t.Helper()
	content, err := os.ReadFile("streams.json")
	if err != nil {
		t.Fatalf("read streams.json: %v", err)
	}
	var bundle struct {
		Streams []recoveryStream `json:"streams"`
	}
	if err := json.Unmarshal(content, &bundle); err != nil {
		t.Fatalf("parse streams.json: %v", err)
	}
	result := make(map[string]recoveryStream, len(bundle.Streams))
	for _, stream := range bundle.Streams {
		result[stream.Name] = stream
	}
	return result
}

func loadRecoverySchema(t *testing.T, name string) recoverySchema {
	t.Helper()
	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read schema %q: %v", name, err)
	}
	var schema recoverySchema
	if err := json.Unmarshal(content, &schema); err != nil {
		t.Fatalf("parse schema %q: %v", name, err)
	}
	return schema
}

func schemaPropertyType(t *testing.T, raw json.RawMessage) any {
	t.Helper()
	if len(raw) == 0 {
		return nil
	}
	var property struct {
		Type any `json:"type"`
	}
	if err := json.Unmarshal(raw, &property); err != nil {
		t.Fatalf("parse schema property: %v", err)
	}
	return property.Type
}

func loadRecoveryFixture(t *testing.T, stream string) map[string]any {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("fixtures", "streams", stream, "page_1.json"))
	if err != nil {
		t.Fatalf("read fixture for %q: %v", stream, err)
	}
	var fixture struct {
		Response struct {
			Body struct {
				Data []map[string]any `json:"data"`
			} `json:"body"`
		} `json:"response"`
	}
	if err := json.Unmarshal(content, &fixture); err != nil {
		t.Fatalf("parse fixture for %q: %v", stream, err)
	}
	if len(fixture.Response.Body.Data) == 0 {
		t.Fatalf("fixture for %q has no response data", stream)
	}
	return fixture.Response.Body.Data[0]
}

func valueAtPath(value map[string]any, path string) any {
	current := any(value)
	for _, segment := range splitPath(path) {
		object, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = object[segment]
	}
	return current
}

func splitPath(path string) []string {
	result := make([]string, 0, 2)
	start := 0
	for index, runeValue := range path {
		if runeValue == '.' {
			result = append(result, path[start:index])
			start = index + 1
		}
	}
	return append(result, path[start:])
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
