package main

import (
	"encoding/json"
	"fmt"

	"polymetrics.ai/internal/connectors/engine"
)

func sourceProjectionInputContract(facts sourceFacts, overrides map[string]engine.FormFieldEncoding) (json.RawMessage, *engine.RequestInputContract, error) {
	containers := map[string]map[string]any{}
	properties := map[string]any{}
	for _, location := range []string{"path", "query", "header"} {
		container := map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false}
		containers[location] = container
		properties[location] = container
	}
	bindings := []engine.RequestInputBinding{}
	queryFields := map[string]engine.FormFieldEncoding{}
	hasStructured := false
	for _, parameter := range facts.Parameters {
		container := containers[parameter.In]
		if container == nil {
			return nil, nil, fmt.Errorf("source parameter location requires typed input projection")
		}
		node, ok := sourceResolveObject(facts, parameter.Node, map[string]bool{}, 0)
		if !ok {
			return nil, nil, fmt.Errorf("source parameter does not resolve")
		}
		schema, kind, encoding, err := sourceProjectionQueryShape(facts, node, parameter.In, parameter.Name, overrides)
		if err != nil {
			return nil, nil, err
		}
		if parameter.In == "query" {
			queryFields[parameter.Name] = encoding
			hasStructured = hasStructured || kind == "object" || kind == "array"
		}
		container["properties"].(map[string]any)[parameter.Name] = schema
		if parameter.Required {
			required, _ := container["required"].([]string)
			container["required"] = append(required, parameter.Name)
		}
		bindings = append(bindings, engine.RequestInputBinding{ConfigKey: parameter.Name, In: parameter.In, Name: parameter.Name})
	}
	raw, err := json.Marshal(map[string]any{"type": "object", "properties": properties, "required": []string{"path", "query", "header"}, "additionalProperties": false})
	if err != nil {
		return nil, nil, err
	}
	contract := &engine.RequestInputContract{Version: 1, Schema: "schemas/" + sourceBytesHash(raw) + ".json", Bindings: bindings}
	if hasStructured {
		contract.QueryEncoding = &engine.FormEncoding{Version: 1, MaxDepth: 32, MaxMembers: 10000, MaxItems: 10000, MaxPairs: 10000, MaxBytes: 64 << 10, Fields: queryFields}
	}
	return raw, contract, nil
}
