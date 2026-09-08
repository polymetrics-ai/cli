package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestCoordinate236Expected(t *testing.T) {
	const object = `{"type":"object"}`
	const array = `{"type":"array"}`
	const scalar = `{"type":"string"}`
	const mismatch = "target_record_projection_mismatch"
	const unverified = "source_schema_unverified"
	const projection = "target_record_projection_unverified"
	type example struct {
		name, schema, path, document string
		single                       bool
		want                         []string
		reason                       string
	}
	cases := []example{}
	for _, path := range []string{"", "."} {
		cases = append(cases, example{name: "root-object-" + path, schema: object, path: path, want: []string{}}, example{name: "root-array-" + path, schema: array, path: path, want: []string{"[]"}})
	}
	for _, end := range []struct {
		name, schema string
		suffix       []string
		reason       string
	}{{"object", object, []string{}, ""}, {"array", array, []string{"[]"}, ""}, {"scalar", scalar, nil, projection}} {
		for _, depth := range []int{1, 2} {
			schema := `{"type":"object","properties":{"leaf":` + end.schema + `}}`
			path := "leaf"
			want := []string{"leaf"}
			if depth == 2 {
				schema = `{"type":"object","properties":{"outer":` + schema + `}}`
				path = "outer.leaf"
				want = []string{"outer", "leaf"}
			}
			want = append(want, end.suffix...)
			if end.reason != "" {
				want = nil
			}
			cases = append(cases, example{name: fmt.Sprintf("terminal-%s-%d", end.name, depth), schema: schema, path: path, want: want, reason: end.reason})
		}
	}
	for _, path := range []string{"missing", "missing.leaf", "outer.missing.leaf", "outer.leaf.missing"} {
		cases = append(cases, example{name: "missing-" + path, schema: `{"type":"object","properties":{"outer":{"type":"object","properties":{"leaf":{"type":"object"}}}}}`, path: path, reason: mismatch})
	}
	for _, schema := range []string{scalar, array} {
		cases = append(cases, example{name: "beyond-" + schema, schema: schema, path: "child", reason: mismatch})
	}
	cases = append(cases,
		example{name: "root-escaped-ref", schema: `{"$ref":"#/$defs/a~1b~0c"}`, document: `{"$defs":{"a/b~c":{"type":"object"}}}`, want: []string{}},
		example{name: "leaf-escaped-ref", schema: `{"type":"object","properties":{"leaf":{"$ref":"#/$defs/a~1b~0c"}}}`, path: "leaf", document: `{"$defs":{"a/b~c":{"type":"array"}}}`, want: []string{"leaf", "[]"}},
		example{name: "cycle", schema: `{"$ref":"#/$defs/a"}`, document: `{"$defs":{"a":{"$ref":"#/$defs/a"}}}`, reason: unverified},
		example{name: "invalid-ref", schema: `{"$ref":"#/absent"}`, document: `{}`, reason: unverified},
		example{name: "nullable-true", schema: `{"type":"object","nullable":true}`, want: []string{}},
		example{name: "nullable-false", schema: `{"type":"object","nullable":false}`, want: []string{}},
		example{name: "nullable-null", schema: `{"type":"object","nullable":null}`, reason: unverified},
		example{name: "nullable-string", schema: `{"type":"object","nullable":"true"}`, reason: unverified},
		example{name: "union", schema: `{"type":["object","null"]}`, reason: unverified},
		example{name: "composition", schema: `{"allOf":[{"type":"object"}]}`, reason: unverified},
		example{name: "contradiction", schema: `{"type":"string","properties":{}}`, reason: mismatch},
		example{name: "unsupported-precedes-contradiction", schema: `{"type":"string","properties":{},"allOf":[]}`, reason: unverified},
		example{name: "single-array", schema: array, single: true, reason: mismatch},
		example{name: "single-object", schema: object, single: true, want: []string{}},
		example{name: "empty-prefix", schema: `{"type":"array","prefixItems":[]}`, want: []string{"[]"}},
		example{name: "nonempty-prefix", schema: `{"type":"array","prefixItems":[{"type":"object"}]}`, reason: projection},
	)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stream := engine.StreamSpec{}
			stream.Records.Path = c.path
			stream.Records.SingleObject = c.single
			got, reason := sourceLaneRecordCoordinate(sourceFacts{Document: json.RawMessage(c.document)}, json.RawMessage(c.schema), stream)
			if !reflect.DeepEqual(got, c.want) || reason != c.reason {
				t.Fatalf("coordinate=%#v reason=%q; want %#v %q", got, reason, c.want, c.reason)
			}
		})
	}
}

func TestCoordinate236EarlyRefusal(t *testing.T) {
	cases := map[string]func(*engine.StreamSpec){
		"filter":      func(s *engine.StreamSpec) { s.Records.Filter = &engine.FilterSpec{} },
		"keyed":       func(s *engine.StreamSpec) { s.Records.KeyedObject = true },
		"wrap":        func(s *engine.StreamSpec) { s.Records.WrapField = "x" },
		"zip":         func(s *engine.StreamSpec) { s.ArrayZipProjection = &engine.ArrayZipProjectionSpec{} },
		"computed":    func(s *engine.StreamSpec) { s.ComputedFields = map[string]string{"x": "y"} },
		"response":    func(s *engine.StreamSpec) { s.ResponseFields = map[string]string{"x": "y"} },
		"passthrough": func(s *engine.StreamSpec) { s.Projection = "passthrough" },
	}
	for name, set := range cases {
		t.Run(name, func(t *testing.T) {
			s := engine.StreamSpec{}
			set(&s)
			a := &sourceShapeAnalysis{}
			got, code := sourceLaneRecordCoordinate(sourceFacts{analysis: a}, json.RawMessage(`invalid`), s)
			if got != nil || code != "target_record_projection_unverified" || a.Visits != 0 {
				t.Fatalf("got %#v %q visits %d", got, code, a.Visits)
			}
		})
	}
}

func TestCoordinate236ResolverBudget(t *testing.T) {
	const raw = `{"type":"object","properties":{"leaf":{"$ref":"#/$defs/a"}}}`
	const doc = `{"$defs":{"a":{"$ref":"#/$defs/b"},"b":{"type":"array"}}}`
	for _, c := range []struct {
		name          string
		start, visits int
		reason        string
		want          []string
		exhausted     bool
	}{
		{"below", 99995, 99999, "", []string{"leaf", "[]"}, false},
		{"at", 99996, 100000, "", []string{"leaf", "[]"}, false},
		{"over", 99997, 100001, "source_schema_unverified", nil, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			a := &sourceShapeAnalysis{Objects: map[string]map[string]json.RawMessage{}, Visits: c.start}
			s := engine.StreamSpec{}
			s.Records.Path = "leaf"
			got, code := sourceLaneRecordCoordinate(sourceFacts{analysis: a, Document: json.RawMessage(doc)}, json.RawMessage(raw), s)
			if !reflect.DeepEqual(got, c.want) || code != c.reason || a.Visits != c.visits || a.Exhausted != c.exhausted {
				t.Fatalf("got %#v %q visits=%d exhausted=%v", got, code, a.Visits, a.Exhausted)
			}
			if c.reason == "" {
				a.Visits = 0
				got, code = sourceLaneRecordCoordinate(sourceFacts{analysis: a, Document: json.RawMessage(doc)}, json.RawMessage(raw), s)
				if !reflect.DeepEqual(got, c.want) || code != "" || a.Visits != 2 {
					t.Fatalf("cached: %#v %q visits=%d", got, code, a.Visits)
				}
			}
		})
	}
}

func TestCoordinate236ReferenceDepth(t *testing.T) {
	for _, count := range []int{255, 256, 257} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			defs := map[string]any{}
			for i := 0; i < count; i++ {
				defs[fmt.Sprint(i)] = map[string]any{"$ref": fmt.Sprintf("#/$defs/%d", i+1)}
			}
			defs[fmt.Sprint(count)] = map[string]any{"type": "object"}
			doc, err := json.Marshal(map[string]any{"$defs": defs})
			if err != nil {
				t.Fatal(err)
			}
			a := &sourceShapeAnalysis{Objects: map[string]map[string]json.RawMessage{}}
			got, code := sourceLaneRecordCoordinate(sourceFacts{Document: doc, analysis: a}, json.RawMessage(`{"$ref":"#/$defs/0"}`), engine.StreamSpec{})
			// The anchor itself consumes a ref step: count=255 ends at depth256.
			if count == 255 {
				if !reflect.DeepEqual(got, []string{}) || code != "" || a.Visits != 257 {
					t.Fatalf("got %#v %q visits=%d", got, code, a.Visits)
				}
			} else if got != nil || code != "source_schema_unverified" || a.Visits != 257 {
				t.Fatalf("got %#v %q visits=%d", got, code, a.Visits)
			}
		})
	}
}

func TestCoordinate236OracleFaults(t *testing.T) {
	want := []string{"outer", "leaf", "[]"}
	for _, c := range []struct {
		name   string
		got    []string
		reason string
	}{
		{"same-count-wrong-coordinate", []string{"other", "leaf", "[]"}, ""},
		{"same-count-wrong-reason", want, "target_record_projection_unverified"},
		{"nil-not-empty", nil, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			expected := want
			if c.name == "nil-not-empty" {
				expected = []string{}
			}
			if reflect.DeepEqual(c.got, expected) && c.reason == "" {
				t.Fatal("fault accepted")
			}
		})
	}
}
