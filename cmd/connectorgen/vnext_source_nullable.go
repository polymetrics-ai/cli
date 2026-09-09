package main

import (
	"encoding/json"
	"fmt"
)

// sourceProjectionNullableScalars translates only explicitly typed source
// scalar nullability. Raw bounds, defaults and enum members remain intact.
// It visits schema coordinates only: example/default objects are values.
func sourceProjectionNullableScalars(source map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	nodes := 0
	var lower func(map[string]json.RawMessage, int) (map[string]json.RawMessage, error)
	lower = func(node map[string]json.RawMessage, depth int) (map[string]json.RawMessage, error) {
		nodes++
		if depth > 256 || nodes > 100000 {
			return nil, fmt.Errorf("source schema lowering budget exceeded")
		}
		out := make(map[string]json.RawMessage, len(node))
		for key, raw := range node {
			out[key] = append(json.RawMessage(nil), raw...)
		}
		if raw, exists := node["nullable"]; exists {
			var nullable bool
			if string(raw) == "null" || json.Unmarshal(raw, &nullable) != nil {
				return nil, fmt.Errorf("source nullable must be boolean")
			}
			var kind string
			if json.Unmarshal(node["type"], &kind) != nil {
				return nil, fmt.Errorf("source nullable requires one explicit scalar type")
			}
			switch kind {
			case "string", "integer", "number", "boolean":
			default:
				return nil, fmt.Errorf("source nullable type is not yet lowered")
			}
			delete(out, "nullable")
			if nullable {
				out["type"], _ = json.Marshal([]string{kind, "null"})
			}
		}
		for _, group := range []string{"properties", "patternProperties"} {
			raw, exists := node[group]
			if !exists {
				continue
			}
			var children map[string]json.RawMessage
			if json.Unmarshal(raw, &children) != nil || children == nil {
				return nil, fmt.Errorf("source schema property group must be object")
			}
			for key, child := range children {
				var nested map[string]json.RawMessage
				if json.Unmarshal(child, &nested) != nil || nested == nil {
					return nil, fmt.Errorf("source property schema must be object")
				}
				result, err := lower(nested, depth+1)
				if err != nil {
					return nil, err
				}
				children[key], _ = json.Marshal(result)
			}
			out[group], _ = json.Marshal(children)
		}
		if raw, exists := node["items"]; exists {
			var child map[string]json.RawMessage
			if json.Unmarshal(raw, &child) != nil || child == nil {
				return nil, fmt.Errorf("source items schema must be object")
			}
			result, err := lower(child, depth+1)
			if err != nil {
				return nil, err
			}
			out["items"], _ = json.Marshal(result)
		}
		for _, group := range []string{"oneOf", "prefixItems"} {
			raw, exists := node[group]
			if !exists {
				continue
			}
			var children []map[string]json.RawMessage
			if json.Unmarshal(raw, &children) != nil || children == nil {
				return nil, fmt.Errorf("source schema alternatives must be array")
			}
			for i, child := range children {
				if child == nil {
					return nil, fmt.Errorf("source alternative schema must be object")
				}
				result, err := lower(child, depth+1)
				if err != nil {
					return nil, err
				}
				children[i] = result
			}
			out[group], _ = json.Marshal(children)
		}
		return out, nil
	}
	return lower(source, 0)
}
