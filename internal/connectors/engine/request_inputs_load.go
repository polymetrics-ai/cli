package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"polymetrics.ai/internal/safety"
)

func loadRequestInputPlans(fsys fs.FS, streams []StreamSpec, operations []OperationSpec) error {
	for index := range streams {
		plan, err := loadRequestInputPlan(fsys, streams[index].RequestInputs)
		if err != nil {
			return fmt.Errorf("stream %s request inputs: %w", streams[index].Name, err)
		}
		if err := validateStreamRequestInputBindings(streams[index], plan); err != nil {
			return fmt.Errorf("stream %s request input binding: %w", streams[index].Name, err)
		}
		streams[index].inputPlan = plan
	}
	for index := range operations {
		if operations[index].REST == nil {
			continue
		}
		plan, err := loadRequestInputPlan(fsys, operations[index].REST.RequestInputs)
		if err != nil {
			return fmt.Errorf("operation %s request inputs: %w", operations[index].ID, err)
		}
		operations[index].REST.inputPlan = plan
	}
	return nil
}

func loadRequestInputPlan(fsys fs.FS, contract *RequestInputContract) (*compiledRequestInputPlan, error) {
	if contract == nil {
		return nil, nil
	}
	if contract.Version != 1 {
		return nil, fmt.Errorf("unsupported request input version")
	}
	reference := contract.Schema
	if !fs.ValidPath(reference) || path.Clean(reference) != reference || !strings.HasPrefix(reference, "schemas/") || path.Ext(reference) != ".json" || strings.Contains(reference, "\\") {
		return nil, fmt.Errorf("request input schema must be confined to schemas")
	}
	file, err := fsys.Open(reference)
	if err != nil {
		return nil, err
	}
	const schemaBytes = 16 << 20
	raw, readErr := io.ReadAll(io.LimitReader(file, schemaBytes+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if len(raw) > schemaBytes {
		return nil, fmt.Errorf("request input schema exceeds byte budget")
	}
	schema, err := CompileSchema(raw)
	if err != nil {
		return nil, err
	}
	root := schema.node
	if !requestInputObject(root) || len(root.properties) < 3 || len(root.properties) > 4 {
		return nil, fmt.Errorf("request input schema must be a closed placed envelope")
	}
	for name := range root.properties {
		if name != "path" && name != "query" && name != "header" && name != "body" {
			return nil, fmt.Errorf("unknown request input envelope member")
		}
	}
	for _, location := range []string{"path", "query", "header"} {
		if !requestInputObject(root.properties[location]) || !containsRequestInputName(root.required, location) {
			return nil, fmt.Errorf("request input %s container must be required and closed", location)
		}
	}
	if len(contract.Bindings) > 4096 {
		return nil, fmt.Errorf("request input binding count exceeds budget")
	}
	bound := map[string]bool{}
	aliases := map[string]bool{}
	for _, binding := range contract.Bindings {
		if binding.ConfigKey != "" {
			if safety.ValidateIdentifier(binding.ConfigKey, "request input alias") != nil || aliases[binding.ConfigKey] {
				return nil, fmt.Errorf("invalid or duplicate request input alias")
			}
			aliases[binding.ConfigKey] = true
		}
		node := root.properties[binding.In]
		coordinate := binding.In + "/" + binding.Name
		if binding.In == "body" {
			if binding.Name != "" || binding.Pointer == "" || !strings.HasPrefix(binding.Pointer, "/") {
				return nil, fmt.Errorf("body input needs an exclusive named pointer")
			}
			coordinate = "body" + binding.Pointer
			for _, part := range strings.Split(binding.Pointer[1:], "/") {
				if part == "" || strings.Contains(part, "~") || node == nil {
					return nil, fmt.Errorf("body input pointer is not a supported named member")
				}
				node = node.properties[part]
			}
		} else {
			if binding.In != "path" && binding.In != "query" && binding.In != "header" {
				return nil, fmt.Errorf("unknown request input location")
			}
			if binding.Name == "" || binding.Pointer != "" {
				return nil, fmt.Errorf("parameter input needs an exclusive name")
			}
			if binding.In == "header" {
				coordinate = binding.In + "/" + strings.ToLower(binding.Name)
			}
			node = node.properties[binding.Name]
		}
		if node == nil || bound[coordinate] {
			return nil, fmt.Errorf("missing or duplicate request input schema coordinate")
		}
		bound[coordinate] = true
	}
	for _, location := range []string{"path", "query", "header"} {
		for name := range root.properties[location].properties {
			coordinate := location + "/" + name
			if location == "header" {
				coordinate = location + "/" + strings.ToLower(name)
			}
			if !bound[coordinate] {
				return nil, fmt.Errorf("request input schema member lacks binding")
			}
		}
	}
	var envelope struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, err
	}
	return &compiledRequestInputPlan{schema: schema, raw: append(json.RawMessage(nil), raw...), bodySchema: append(json.RawMessage(nil), envelope.Properties["body"]...), bindings: append([]RequestInputBinding(nil), contract.Bindings...)}, nil
}

func requestInputObject(node *schemaNode) bool {
	return node != nil && len(node.types) == 1 && node.types[0] == "object" && node.hasAdditionalProps && !node.additionalProperties
}

func containsRequestInputName(names []string, name string) bool {
	for _, candidate := range names {
		if candidate == name {
			return true
		}
	}
	return false
}

func validateStreamRequestInputBindings(stream StreamSpec, plan *compiledRequestInputPlan) error {
	if plan == nil {
		return nil
	}
	for _, binding := range plan.bindings {
		if binding.In == "body" {
			continue
		}
		if binding.ConfigKey == "" {
			return fmt.Errorf("saved parameter binding lacks an input alias")
		}
		template := "{{ config." + binding.ConfigKey + " }}"
		switch binding.In {
		case "path":
			if !strings.Contains(stream.Path, template) {
				return fmt.Errorf("path alias differs from selected input binding")
			}
		case "query":
			query, exists := stream.Query[binding.Name]
			if !exists || query.Template != template {
				return fmt.Errorf("query alias differs from selected input binding")
			}
			required := containsRequestInputName(plan.schema.node.properties["query"].required, binding.Name)
			if query.OmitWhenAbsent == required || query.Default != "" {
				return fmt.Errorf("query presence/default differs from selected input schema")
			}
		case "header":
			if stream.Headers[binding.Name] != template {
				return fmt.Errorf("header alias differs from selected input binding")
			}
		}
	}
	return nil
}
