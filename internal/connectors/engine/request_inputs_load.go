package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/safety"
)

func loadRequestInputPlans(fsys fs.FS, streams []StreamSpec, operations []OperationSpec, base HTTPBase) error {
	for index := range streams {
		plan, err := loadRequestInputPlan(fsys, streams[index].RequestInputs)
		if err != nil {
			return fmt.Errorf("stream %s request inputs: %w", streams[index].Name, err)
		}
		if err := validateStreamRequestInputBindings(streams[index], plan); err != nil {
			return fmt.Errorf("stream %s request input binding: %w", streams[index].Name, err)
		}
		if plan != nil {
			protected := operationRuntimeHeaderNamesForBase(base)
			for _, binding := range plan.bindings {
				if binding.In == "header" {
					name, err := connectors.CanonicalOperationHeaderName(binding.Name)
					if err != nil {
						return err
					}
					if _, forbidden := protected[name]; forbidden {
						return fmt.Errorf("request input header is runtime-owned")
					}
				}
			}
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
	if err := prepareInputSchemaNode(schema.node); err != nil {
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
				canonical, err := connectors.CanonicalOperationHeaderName(binding.Name)
				if err != nil || connectors.IsProtectedOperationHeaderName(canonical) {
					return nil, fmt.Errorf("request input header is invalid or protected")
				}
				coordinate = binding.In + "/" + canonical
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
	var queryEncoding *FormEncoding
	if contract.QueryEncoding != nil {
		encoded, err := json.Marshal(contract.QueryEncoding)
		if err != nil {
			return nil, fmt.Errorf("invalid query encoding")
		}
		if err := json.Unmarshal(encoded, &queryEncoding); err != nil {
			return nil, fmt.Errorf("invalid query encoding")
		}
		if err := validateFormEncoding(queryEncoding, envelope.Properties["query"], nil); err != nil {
			return nil, fmt.Errorf("query encoding: %w", err)
		}
	}
	return &compiledRequestInputPlan{queryEncoding: queryEncoding, querySchema: append(json.RawMessage(nil), envelope.Properties["query"]...), schema: schema, raw: append(json.RawMessage(nil), raw...), bodySchema: append(json.RawMessage(nil), envelope.Properties["body"]...), bindings: append([]RequestInputBinding(nil), contract.Bindings...)}, nil
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
	boundQuery := map[string]bool{}
	boundPath := stream.Path
	for _, binding := range plan.bindings {
		if binding.In == "query" {
			boundQuery[binding.Name] = true
		}
		if binding.In == "path" {
			boundPath = strings.ReplaceAll(boundPath, "{{ config."+binding.ConfigKey+" }}", "bound")
		}
	}
	if strings.Contains(boundPath, "{{") || strings.Contains(boundPath, "}}") {
		return fmt.Errorf("stream path contains unbound input")
	}
	for name := range stream.Query {
		if !boundQuery[name] {
			return fmt.Errorf("stream query contains unbound input")
		}
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

// Input contracts use the existing validator with the configuration format
// predicates. The nodes belong solely to this newly compiled input schema.
func prepareInputSchemaNode(node *schemaNode) error {
	if node == nil {
		return nil
	}
	node.inputFormats = true
	if node.format != "" && !isSupportedConfigurationFormat(node.format) {
		return fmt.Errorf("unsupported request input format")
	}
	for _, child := range node.properties {
		if err := prepareInputSchemaNode(child); err != nil {
			return err
		}
	}
	for _, child := range node.oneOf {
		if err := prepareInputSchemaNode(child); err != nil {
			return err
		}
	}
	for _, child := range node.prefixItems {
		if err := prepareInputSchemaNode(child); err != nil {
			return err
		}
	}
	for _, child := range node.patternProperties {
		if err := prepareInputSchemaNode(child.schema); err != nil {
			return err
		}
	}
	if err := prepareInputSchemaNode(node.items); err != nil {
		return err
	}
	if node.hasDefault {
		if err := node.validate(node.defaultVal, ""); err != nil {
			return fmt.Errorf("request input default violates its schema")
		}
	}
	return nil
}
