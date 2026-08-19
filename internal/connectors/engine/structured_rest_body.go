package engine

import (
	"encoding/json"
	"fmt"
	"math"
	"mime"
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
// contract. Commandrunner and connectorgen call this same function, while the
// typed executor calls it again when a supplied value is structured.
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
	if op.REST == nil || len(op.REST.BodySchema) == 0 {
		return nil, fmt.Errorf("operation %q structured JSON body requires body_schema", op.ID)
	}
	var root map[string]any
	if err := json.Unmarshal(op.REST.BodySchema, &root); err != nil {
		return nil, fmt.Errorf("operation %q body_schema is not an object: %w", op.ID, err)
	}
	if !isObjectType(root) {
		return nil, fmt.Errorf("operation %q body_schema must be an object", op.ID)
	}
	if _, err := CompileSchema(op.REST.BodySchema); err != nil {
		return nil, fmt.Errorf("operation %q body_schema: %w", op.ID, err)
	}
	state := structuredRESTBodySchemaState{}
	if err := requireClosedBoundedStructuredRESTBody(op.ID, root, "body_schema", 1, &state); err != nil {
		return nil, err
	}
	return root, nil
}

type structuredRESTBodySchemaState struct {
	fields int
}

func requireClosedBoundedStructuredRESTBody(operation string, node map[string]any, path string, depth int, state *structuredRESTBodySchemaState) error {
	if depth > maxStructuredRESTBodyDepth {
		return fmt.Errorf("operation %q %s exceeds structured body depth limit %d", operation, path, maxStructuredRESTBodyDepth)
	}
	if structuredRESTBodyNodeMayAcceptObjectOrArray(node) && !structuredRESTBodyNodeHasExplicitType(node) {
		return fmt.Errorf("operation %q %s may accept objects or arrays without an explicit type", operation, path)
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
		rawItems, ok := node["items"]
		if !ok {
			return fmt.Errorf("operation %q %s declares an array without an items schema", operation, path)
		}
		items, ok := rawItems.(map[string]any)
		if !ok || len(items) == 0 {
			return fmt.Errorf("operation %q %s declares an array with an empty items schema", operation, path)
		}
		if err := requireClosedBoundedStructuredRESTBody(operation, items, path+"/items", depth+1, state); err != nil {
			return err
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
	if _, ok := value.(map[string]any); ok {
		return true
	}
	_, ok := arrayElements(value)
	return ok
}
