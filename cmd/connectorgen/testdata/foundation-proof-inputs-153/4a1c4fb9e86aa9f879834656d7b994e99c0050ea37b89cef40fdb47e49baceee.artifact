package connsdk

import (
	"encoding/json"
	"regexp"
	"sort"
)

// FieldSpec describes one inferred field. Name/Type map directly onto
// connectors.Field, so a connector can build its catalog from inferred specs.
type FieldSpec struct {
	Name string
	Type string
}

// timestampPattern matches common ISO-8601 / RFC3339 date-time strings so they
// can be typed as "timestamp" rather than "string".
var timestampPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}([T ]\d{2}:\d{2}(:\d{2})?(\.\d+)?(Z|[+-]\d{2}:?\d{2})?)?$`)

// InferType classifies a decoded JSON value into a connector field type:
// string, integer, number, boolean, timestamp, object, array, or null.
func InferType(v any) string {
	switch t := v.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case json.Number:
		if _, err := t.Int64(); err == nil {
			return "integer"
		}
		return "number"
	case float64:
		if t == float64(int64(t)) {
			return "integer"
		}
		return "number"
	case string:
		if timestampPattern.MatchString(t) {
			return "timestamp"
		}
		return "string"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	default:
		return "string"
	}
}

// InferFields derives sorted FieldSpecs from a sample record. Nested objects and
// arrays are reported as object/array fields (not flattened) so downstream
// schema mapping stays predictable.
func InferFields(sample Record) []FieldSpec {
	out := make([]FieldSpec, 0, len(sample))
	for name, value := range sample {
		out = append(out, FieldSpec{Name: name, Type: InferType(value)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// InferFieldsFromSamples merges field specs across multiple records, which is
// more robust than a single sample when some fields are intermittently null.
func InferFieldsFromSamples(samples []Record) []FieldSpec {
	types := map[string]string{}
	for _, rec := range samples {
		for name, value := range rec {
			t := InferType(value)
			existing, ok := types[name]
			if !ok || existing == "null" {
				types[name] = t
				continue
			}
			// Prefer a concrete type over null; otherwise keep the first seen.
			if existing != t && t != "null" && existing == "" {
				types[name] = t
			}
		}
	}
	out := make([]FieldSpec, 0, len(types))
	for name, t := range types {
		out = append(out, FieldSpec{Name: name, Type: t})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
