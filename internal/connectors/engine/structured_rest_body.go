package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"mime"
	"reflect"
	"strings"

	"polymetrics.ai/internal/safety"
)

const (
	maxStructuredRESTBodyDepth  = 16
	maxStructuredRESTBodyFields = 256
	maxStructuredRESTBodyItems  = 1024
)

// PreflightOperationStructuredJSONBodyField proves that a command's one
// structured input names a closed, bounded top-level field of its fixed
// operation. It is deliberately a name-only preflight: callers never provide
// a body template, route, method, media type, or arbitrary nested path.
func PreflightOperationStructuredJSONBodyField(b Bundle, operation, field string) error {
	op, err := findOperation(b, operation)
	if err != nil {
		return err
	}
	return ValidateOperationStructuredJSONBodyField(op, field)
}

// ValidateOperationStructuredJSONBodyField is the common declaration gate for
// a json command flag on a fixed write operation. GraphQL retains its existing
// closed-variable contract; REST uses the equivalent closed body-schema
// contract. Commandrunner and connectorgen call this same function; typed
// execution compiles the same schema at its shared body boundary.
func ValidateOperationStructuredJSONBodyField(op OperationSpec, field string) error {
	field = strings.TrimSpace(field)
	if err := safety.ValidateIdentifier(field, "structured body field"); err != nil {
		return err
	}
	if strings.Contains(field, ".") {
		return fmt.Errorf("operation %q structured body field %q must be top-level", op.ID, field)
	}

	switch op.Kind {
	case "graphql_mutation":
		return ValidateGraphQLOperationStructuredJSONVariable(op, field)
	case "rest_write":
		return validateRESTStructuredJSONBodyField(op, field)
	default:
		return fmt.Errorf("operation %q structured JSON body requires rest_write or graphql_mutation, got %q", op.ID, op.Kind)
	}
}

func validateRESTStructuredJSONBodyField(op OperationSpec, field string) error {
	if op.REST == nil {
		return fmt.Errorf("operation %q rest_write has no rest declaration", op.ID)
	}
	if op.REST.Multipart != nil {
		return fmt.Errorf("operation %q structured JSON body is unavailable for multipart writes", op.ID)
	}
	if _, _, err := operationStructuredJSONContentType(op); err != nil {
		return err
	}
	root, err := structuredRESTBodySchema(op)
	if err != nil {
		return err
	}
	properties, ok := root["properties"].(map[string]any)
	if !ok {
		return fmt.Errorf("operation %q body_schema must declare properties", op.ID)
	}
	raw, ok := properties[field]
	if !ok {
		return fmt.Errorf("operation %q body_schema does not declare structured field %q", op.ID, field)
	}
	node, ok := raw.(map[string]any)
	if !ok || (!isObjectType(node) && !isArrayType(node)) {
		return fmt.Errorf("operation %q body_schema field %q must be an object or array", op.ID, field)
	}
	return nil
}

func operationStructuredJSONContentType(op OperationSpec) (string, string, error) {
	if op.REST == nil {
		return "", "", fmt.Errorf("operation %q has no rest declaration", op.ID)
	}
	declared := strings.TrimSpace(op.REST.ContentType)
	if declared == "" {
		return "application/json", "json", nil
	}
	mediaType, _, err := mime.ParseMediaType(declared)
	if err != nil {
		return "", "", fmt.Errorf("operation %q has invalid rest content_type %q: %w", op.ID, declared, err)
	}
	switch strings.ToLower(mediaType) {
	case "application/json", "application/scim+json":
		return strings.ToLower(mediaType), "json", nil
	default:
		return "", "", fmt.Errorf("operation %q structured JSON body requires application/json or application/scim+json content_type", op.ID)
	}
}

func structuredRESTBodySchema(op OperationSpec) (map[string]any, error) {
	compiled, err := compileStructuredRESTBodySchema(op)
	if err != nil {
		return nil, err
	}
	return compiled.root, nil
}

func operationHasStructuredRESTBodyField(op OperationSpec) bool {
	if op.REST == nil || len(op.REST.BodySchema) == 0 {
		return false
	}
	var root map[string]any
	if err := json.Unmarshal(op.REST.BodySchema, &root); err != nil {
		return false
	}
	properties, ok := root["properties"].(map[string]any)
	if !ok {
		return false
	}
	for _, raw := range properties {
		node, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if isObjectType(node) || isArrayType(node) {
			return true
		}
		if _, ok := node["properties"]; ok {
			return true
		}
		if _, ok := node["additionalProperties"]; ok {
			return true
		}
		if _, ok := node["items"]; ok {
			return true
		}
		if _, ok := node["prefixItems"]; ok {
			return true
		}
	}
	return false
}

func operationHasStructuredRESTBodyValue(body map[string]any) bool {
	for _, value := range body {
		if isStructuredJSONBodyValue(value) {
			return true
		}
	}
	return false
}

type structuredRESTBodySchemaCompilation struct {
	root   map[string]any
	schema *Schema
}

func compileStructuredRESTBodySchema(op OperationSpec) (*structuredRESTBodySchemaCompilation, error) {
	if op.REST == nil || len(op.REST.BodySchema) == 0 {
		return nil, fmt.Errorf("operation %q structured JSON body requires body_schema", op.ID)
	}
	var root map[string]any
	if err := json.Unmarshal(op.REST.BodySchema, &root); err != nil {
		return nil, fmt.Errorf("operation %q body_schema is not an object: %w", op.ID, err)
	}
	if root == nil {
		return nil, fmt.Errorf("operation %q body_schema must be an object", op.ID)
	}
	if err := normalizeStructuredRESTBodySchemaNode(op.ID, root, "body_schema"); err != nil {
		return nil, err
	}
	if !isObjectType(root) {
		return nil, fmt.Errorf("operation %q body_schema must be an object", op.ID)
	}
	normalized, err := json.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("operation %q body_schema: normalize: %w", op.ID, err)
	}
	sch, err := compileStructuredRESTBodySchemaDocument(normalized)
	if err != nil {
		return nil, fmt.Errorf("operation %q body_schema: %w", op.ID, err)
	}
	state := structuredRESTBodySchemaState{}
	if err := requireClosedBoundedStructuredRESTBody(op.ID, root, "body_schema", 1, &state); err != nil {
		return nil, err
	}
	return &structuredRESTBodySchemaCompilation{root: root, schema: sch}, nil
}

func materializeStructuredRESTBody(op OperationSpec, body map[string]any) (map[string]any, error) {
	compiled, err := compileStructuredRESTBodySchema(op)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("operation %q: encode structured body: %w", op.ID, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var canonical map[string]any
	if err := decoder.Decode(&canonical); err != nil {
		return nil, fmt.Errorf("operation %q: decode structured body: %w", op.ID, err)
	}
	if canonical == nil {
		return nil, fmt.Errorf("operation %q: structured body must be an object", op.ID)
	}
	if err := compiled.schema.Validate(canonical); err != nil {
		return nil, fmt.Errorf("operation %q: body_schema: %w", op.ID, err)
	}
	return canonical, nil
}

type structuredRESTBodySchemaState struct {
	fields int
}

func requireClosedBoundedStructuredRESTBody(operation string, node map[string]any, path string, depth int, state *structuredRESTBodySchemaState) error {
	if depth > maxStructuredRESTBodyDepth {
		return fmt.Errorf("operation %q %s exceeds structured body depth limit %d", operation, path, maxStructuredRESTBodyDepth)
	}
	if structuredRESTBodyNodeMayAcceptObjectOrArray(node) && !structuredRESTBodyNodeHasExplicitType(node) {
		return structuredRESTBodyFoundationGap(operation, path, "may accept objects or arrays; declare an explicit type")
	}
	if isObjectType(node) {
		if closed, ok := node["additionalProperties"].(bool); !ok || closed {
			return fmt.Errorf("operation %q %s is an object and must declare additionalProperties: false", operation, path)
		}
	}
	if isArrayType(node) {
		raw, ok := node["maxItems"]
		if !ok {
			return fmt.Errorf("operation %q %s declares an array without maxItems", operation, path)
		}
		maxItems, ok := raw.(float64)
		if !ok || math.Trunc(maxItems) != maxItems || maxItems < 0 {
			return fmt.Errorf("operation %q %s has invalid maxItems", operation, path)
		}
		if maxItems > maxStructuredRESTBodyItems {
			return fmt.Errorf("operation %q %s maxItems %.0f exceeds structured body limit %d", operation, path, maxItems, maxStructuredRESTBodyItems)
		}
		rawItems, hasItems := node["items"]
		rawPrefixItems, hasPrefixItems := node["prefixItems"]
		if !hasItems && !hasPrefixItems {
			return fmt.Errorf("operation %q %s declares an array without an items schema", operation, path)
		}
		if hasItems {
			items, ok := rawItems.(map[string]any)
			if !ok || len(items) == 0 {
				return fmt.Errorf("operation %q %s declares an array with an empty items schema", operation, path)
			}
			if err := requireClosedBoundedStructuredRESTBody(operation, items, path+"/items", depth+1, state); err != nil {
				return err
			}
		}
		if hasPrefixItems {
			prefixItems, ok := rawPrefixItems.([]any)
			if !ok || len(prefixItems) == 0 {
				return fmt.Errorf("operation %q %s declares an array with an empty prefixItems schema", operation, path)
			}
			for index, rawPrefixItem := range prefixItems {
				prefixItem, ok := rawPrefixItem.(map[string]any)
				if !ok || len(prefixItem) == 0 {
					return fmt.Errorf("operation %q %s/prefixItems/%d must be a non-empty schema object", operation, path, index)
				}
				if err := requireClosedBoundedStructuredRESTBody(operation, prefixItem, fmt.Sprintf("%s/prefixItems/%d", path, index), depth+1, state); err != nil {
					return err
				}
			}
			if !hasItems && maxItems > float64(len(prefixItems)) {
				return structuredRESTBodyFoundationGap(operation, path, "prefixItems does not constrain every allowed array item; declare an items schema for remaining items")
			}
		}
	}
	properties, ok := node["properties"].(map[string]any)
	if !ok {
		return nil
	}
	state.fields += len(properties)
	if state.fields > maxStructuredRESTBodyFields {
		return fmt.Errorf("operation %q %s exceeds structured body field limit %d", operation, path, maxStructuredRESTBodyFields)
	}
	for _, name := range sortedMapKeys(properties) {
		child, ok := properties[name].(map[string]any)
		if !ok {
			return fmt.Errorf("operation %q %s/%s must be a schema object", operation, path, name)
		}
		if err := requireClosedBoundedStructuredRESTBody(operation, child, path+"/"+name, depth+1, state); err != nil {
			return err
		}
	}
	return nil
}

func normalizeStructuredRESTBodySchemaNode(operation string, node map[string]any, path string) error {
	types, hasExplicitType, err := structuredRESTBodyExplicitTypes(operation, path, node)
	if err != nil {
		return err
	}
	_, hasProperties := node["properties"]
	_, hasAdditionalProperties := node["additionalProperties"]
	_, hasItems := node["items"]
	_, hasPrefixItems := node["prefixItems"]
	hasObjectStructure := hasProperties || hasAdditionalProperties
	hasArrayStructure := hasItems || hasPrefixItems

	if !hasExplicitType {
		switch {
		case hasObjectStructure && hasArrayStructure:
			return structuredRESTBodyFoundationGap(operation, path, "has both object and array structure without an explicit type")
		case hasObjectStructure:
			node["type"] = "object"
		case hasArrayStructure:
			node["type"] = "array"
		}
	} else {
		if hasObjectStructure && !structuredRESTBodyTypeAllows(types, "object") {
			return structuredRESTBodyFoundationGap(operation, path, "object structure conflicts with its explicit type")
		}
		if hasArrayStructure && !structuredRESTBodyTypeAllows(types, "array") {
			return structuredRESTBodyFoundationGap(operation, path, "array structure conflicts with its explicit type")
		}
	}

	if hasProperties {
		properties, ok := node["properties"].(map[string]any)
		if !ok {
			return structuredRESTBodyFoundationGap(operation, path, "properties must be an object of schema objects")
		}
		for _, name := range sortedMapKeys(properties) {
			child, ok := properties[name].(map[string]any)
			if !ok {
				return structuredRESTBodyFoundationGap(operation, path+"/"+name, "must be a schema object")
			}
			if err := normalizeStructuredRESTBodySchemaNode(operation, child, path+"/"+name); err != nil {
				return err
			}
		}
	}
	if hasItems {
		items, ok := node["items"].(map[string]any)
		if !ok {
			return structuredRESTBodyFoundationGap(operation, path+"/items", "must be a schema object")
		}
		if err := normalizeStructuredRESTBodySchemaNode(operation, items, path+"/items"); err != nil {
			return err
		}
	}
	if hasPrefixItems {
		prefixItems, ok := node["prefixItems"].([]any)
		if !ok {
			return structuredRESTBodyFoundationGap(operation, path+"/prefixItems", "must be an array of schema objects")
		}
		for index, rawPrefixItem := range prefixItems {
			prefixItem, ok := rawPrefixItem.(map[string]any)
			if !ok {
				return structuredRESTBodyFoundationGap(operation, fmt.Sprintf("%s/prefixItems/%d", path, index), "must be a schema object")
			}
			if err := normalizeStructuredRESTBodySchemaNode(operation, prefixItem, fmt.Sprintf("%s/prefixItems/%d", path, index)); err != nil {
				return err
			}
		}
	}
	return nil
}

func structuredRESTBodyExplicitTypes(operation string, path string, node map[string]any) ([]string, bool, error) {
	raw, ok := node["type"]
	if !ok {
		return nil, false, nil
	}
	switch types := raw.(type) {
	case string:
		if types == "" {
			return nil, false, structuredRESTBodyFoundationGap(operation, path, "type must be a non-empty JSON type")
		}
		return []string{types}, true, nil
	case []any:
		if len(types) == 0 {
			return nil, false, structuredRESTBodyFoundationGap(operation, path, "type must be a non-empty JSON type list")
		}
		declared := make([]string, 0, len(types))
		for _, rawType := range types {
			typeName, ok := rawType.(string)
			if !ok || typeName == "" {
				return nil, false, structuredRESTBodyFoundationGap(operation, path, "type must contain only non-empty JSON type names")
			}
			declared = append(declared, typeName)
		}
		return declared, true, nil
	default:
		return nil, false, structuredRESTBodyFoundationGap(operation, path, "type must be a JSON type name or list")
	}
}

func structuredRESTBodyTypeAllows(types []string, want string) bool {
	for _, declared := range types {
		if declared == want {
			return true
		}
	}
	return false
}

func structuredRESTBodyFoundationGap(operation string, path string, detail string) error {
	return fmt.Errorf("operation %q %s has a structured-body foundation gap: %s", operation, path, detail)
}

func structuredRESTBodyNodeMayAcceptObjectOrArray(node map[string]any) bool {
	if enum, ok := node["enum"].([]any); ok && len(enum) > 0 {
		for _, value := range enum {
			if isStructuredJSONBodyValue(value) {
				return true
			}
		}
		return false
	}

	switch types := node["type"].(type) {
	case string:
		return types == "object" || types == "array"
	case []any:
		if len(types) == 0 {
			return true
		}
		for _, value := range types {
			typeName, ok := value.(string)
			if !ok || typeName == "object" || typeName == "array" {
				return true
			}
		}
		return false
	default:
		return true
	}
}

func structuredRESTBodyNodeHasExplicitType(node map[string]any) bool {
	switch types := node["type"].(type) {
	case string:
		return types != ""
	case []any:
		return len(types) > 0
	default:
		return false
	}
}

func isStructuredJSONBodyValue(value any) bool {
	if value == nil {
		return false
	}
	if reflect.ValueOf(value).Kind() == reflect.Map {
		return true
	}
	_, ok := arrayElements(value)
	return ok
}
