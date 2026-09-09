package main

import (
	"encoding/json"
	"fmt"
	"polymetrics.ai/internal/connectors/engine"
)

// Only source-explicit serialization enters this arm. Documentary dialects
// such as bracket-suffixed arrays need their separate reviewed semantics.
func sourceProjectionQueryShape(facts sourceFacts, node map[string]json.RawMessage, location, name string, overrides map[string]engine.FormFieldEncoding) (map[string]json.RawMessage, string, engine.FormFieldEncoding, error) {
	override, supplied := overrides[name]
	if schema, kind, ok := sourceProjectionScalarSchema(facts, node["schema"]); ok {
		if supplied {
			return nil, "", override, fmt.Errorf("scalar query does not consume structured dialect")
		}
		return schema, kind, engine.FormFieldEncoding{Mode: "scalar"}, nil
	}
	if location != "query" {
		return nil, "", engine.FormFieldEncoding{}, fmt.Errorf("structured parameter requires query placement")
	}
	schema, ok := sourceResolveObject(facts, node["schema"], map[string]bool{}, 0)
	if !ok {
		return nil, "", engine.FormFieldEncoding{}, fmt.Errorf("structured query schema does not resolve")
	}
	if supplied {
		var kind string
		if json.Unmarshal(schema["type"], &kind) != nil || (kind != "array" && kind != "object") {
			return nil, "", override, fmt.Errorf("query dialect requires explicit structured source type")
		}
		for _, field := range []string{"style", "explode", "content", "allowReserved", "allowEmptyValue"} {
			if len(node[field]) != 0 {
				return nil, "", override, fmt.Errorf("authored query dialect conflicts with explicit source serialization")
			}
		}
		raw, err := json.Marshal(schema)
		if err != nil {
			return nil, "", override, err
		}
		if _, err = engine.CompileSchema(raw); err != nil {
			return nil, "", override, err
		}
		return schema, kind, override, nil
	}
	var kind, style string
	var explode bool
	if json.Unmarshal(schema["type"], &kind) != nil || json.Unmarshal(node["style"], &style) != nil || json.Unmarshal(node["explode"], &explode) != nil || !explode {
		return nil, "", engine.FormFieldEncoding{}, fmt.Errorf("structured query requires explicit source type/style/explode")
	}
	field := engine.FormFieldEncoding{Null: "reject"}
	switch {
	case kind == "array" && style == "form":
		var minimum int
		if json.Unmarshal(schema["minItems"], &minimum) != nil || minimum < 1 {
			return nil, "", field, fmt.Errorf("empty query array requires explicit source disposition")
		}
		if _, _, ok := sourceProjectionScalarSchema(facts, schema["items"]); !ok {
			return nil, "", field, fmt.Errorf("repeated query items require scalar schema")
		}
		field.Mode = "repeated"
		field.EmptyArray = "omit"
	case kind == "object" && style == "deepObject":
		var required []string
		var properties map[string]json.RawMessage
		if json.Unmarshal(schema["required"], &required) != nil || len(required) == 0 || json.Unmarshal(schema["properties"], &properties) != nil {
			return nil, "", field, fmt.Errorf("empty query object requires explicit source disposition")
		}
		for _, raw := range properties {
			if _, _, ok := sourceProjectionScalarSchema(facts, raw); !ok {
				return nil, "", field, fmt.Errorf("deepObject query requires scalar properties")
			}
		}
		field.Mode = "brackets"
	default:
		return nil, "", field, fmt.Errorf("structured query dialect is not yet lowered")
	}
	for _, name := range []string{"content", "allowReserved", "allowEmptyValue"} {
		if len(node[name]) != 0 {
			return nil, "", field, fmt.Errorf("structured query serialization requires source reconciliation")
		}
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		return nil, "", field, err
	}
	if _, err = engine.CompileSchema(raw); err != nil {
		return nil, "", field, err
	}
	return schema, kind, field, nil
}
