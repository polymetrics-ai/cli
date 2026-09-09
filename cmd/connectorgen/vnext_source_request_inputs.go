package main

import (
	"encoding/json"
	"fmt"

	"polymetrics.ai/internal/connectors/engine"
)

func sourceProjectionInputContract(facts sourceFacts) (json.RawMessage, *engine.RequestInputContract, error) {
	containers := map[string]map[string]any{}
	properties := map[string]any{}
	for _, location := range []string{"path", "query", "header"} {
		container := map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false}
		containers[location] = container
		properties[location] = container
	}
	bindings := []engine.RequestInputBinding{}
	for _, parameter := range facts.Parameters {
		container := containers[parameter.In]
		if container == nil {
			return nil, nil, fmt.Errorf("source parameter location requires typed input projection")
		}
		node, ok := sourceResolveObject(facts, parameter.Node, map[string]bool{}, 0)
		if !ok {
			return nil, nil, fmt.Errorf("source parameter does not resolve")
		}
		schema, _, ok := sourceProjectionScalarSchema(facts, node["schema"])
		if !ok {
			return nil, nil, fmt.Errorf("source parameter requires a supported scalar schema")
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
	return raw, &engine.RequestInputContract{Version: 1, Schema: "schemas/" + sourceBytesHash(raw) + ".json", Bindings: bindings}, nil
}
