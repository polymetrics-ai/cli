package engine

import (
	"fmt"
	"strings"
)

// WithholdOperationDirectWriteBodyFields removes only declaration-owned body
// paths from a canonical structured body before it reaches durable plan state.
func WithholdOperationDirectWriteBodyFields(bundle Bundle, operation string, body map[string]any, fields []string) (map[string]any, []string, error) {
	op, root, canonical, err := canonicalStructuredBodyPlanFragment(bundle, operation, body)
	if err != nil {
		return nil, nil, err
	}
	_ = op
	withheld := make([]string, 0, len(fields))
	for _, raw := range fields {
		path := structuredBodyRelativePath(raw)
		if path == "" {
			continue
		}
		resolved, err := resolveOperationDirectWriteBodySchemaPath(root, path)
		if err != nil {
			return nil, nil, fmt.Errorf("operation %q sensitive body field %q: %w", operation, path, err)
		}
		removed, err := deleteStructuredBodyValue(canonical, resolved.steps, path)
		if err != nil {
			return nil, nil, err
		}
		if removed {
			withheld = append(withheld, path)
		}
	}
	return canonical, withheld, nil
}

func RedactOperationDirectWriteBodyFields(bundle Bundle, operation string, body map[string]any, fields []string) (map[string]any, error) {
	_, root, canonical, err := canonicalStructuredBodyPlanFragment(bundle, operation, body)
	if err != nil {
		return nil, err
	}
	for _, raw := range fields {
		path := structuredBodyRelativePath(raw)
		if path == "" {
			continue
		}
		resolved, err := resolveOperationDirectWriteBodySchemaPath(root, path)
		if err != nil {
			return nil, fmt.Errorf("operation %q sensitive body field %q: %w", operation, path, err)
		}
		if _, found := operationDirectWriteBodyPathValue(canonical, resolved.steps); !found {
			continue
		}
		updated, err := setOperationDirectWriteBodyPathValue(canonical, resolved.steps, "redacted", path)
		if err != nil {
			return nil, err
		}
		var ok bool
		canonical, ok = updated.(map[string]any)
		if !ok || canonical == nil {
			return nil, fmt.Errorf("operation %q redacted body must be an object", operation)
		}
	}
	return canonical, nil
}

func MergeOperationDirectWriteBodyFragments(bundle Bundle, operation string, base, overlay map[string]any) (map[string]any, error) {
	op, _, err := structuredBodyPlanRoot(bundle, operation)
	if err != nil {
		return nil, err
	}
	merged, err := mergeStructuredBodyMaps(base, overlay, "body")
	if err != nil {
		return nil, err
	}
	compiled, err := compileStructuredRESTBodySchema(op)
	if err != nil {
		return nil, err
	}
	return canonicalizeStructuredRESTBodyFragment(compiled, op, merged, "body")
}

func OperationDirectWriteBodyPathContains(bundle Bundle, operation, parent, child string) (bool, error) {
	_, root, err := structuredBodyPlanRoot(bundle, operation)
	if err != nil {
		return false, err
	}
	parentPath, err := resolveOperationDirectWriteBodySchemaPath(root, structuredBodyRelativePath(parent))
	if err != nil {
		return false, err
	}
	childPath, err := resolveOperationDirectWriteBodySchemaPath(root, structuredBodyRelativePath(child))
	if err != nil {
		return false, err
	}
	if len(parentPath.steps) > len(childPath.steps) {
		return false, nil
	}
	for index, step := range parentPath.steps {
		other := childPath.steps[index]
		if step.array != other.array || step.key != other.key || step.index != other.index {
			return false, nil
		}
	}
	return true, nil
}

func (connector *Connector) WithholdOperationDirectWriteBodyFields(operation string, body map[string]any, fields []string) (map[string]any, []string, error) {
	return WithholdOperationDirectWriteBodyFields(connector.bundle, operation, body, fields)
}

func (connector *Connector) RedactOperationDirectWriteBodyFields(operation string, body map[string]any, fields []string) (map[string]any, error) {
	return RedactOperationDirectWriteBodyFields(connector.bundle, operation, body, fields)
}

func (connector *Connector) MergeOperationDirectWriteBodyFragments(operation string, base, overlay map[string]any) (map[string]any, error) {
	return MergeOperationDirectWriteBodyFragments(connector.bundle, operation, base, overlay)
}

func (connector *Connector) OperationDirectWriteBodyPathContains(operation, parent, child string) (bool, error) {
	return OperationDirectWriteBodyPathContains(connector.bundle, operation, parent, child)
}

func (base Base) WithholdOperationDirectWriteBodyFields(operation string, body map[string]any, fields []string) (map[string]any, []string, error) {
	return WithholdOperationDirectWriteBodyFields(base.bundle, operation, body, fields)
}

func (base Base) RedactOperationDirectWriteBodyFields(operation string, body map[string]any, fields []string) (map[string]any, error) {
	return RedactOperationDirectWriteBodyFields(base.bundle, operation, body, fields)
}

func (base Base) MergeOperationDirectWriteBodyFragments(operation string, first, second map[string]any) (map[string]any, error) {
	return MergeOperationDirectWriteBodyFragments(base.bundle, operation, first, second)
}

func (base Base) OperationDirectWriteBodyPathContains(operation, parent, child string) (bool, error) {
	return OperationDirectWriteBodyPathContains(base.bundle, operation, parent, child)
}

func canonicalStructuredBodyPlanFragment(bundle Bundle, operation string, body map[string]any) (OperationSpec, map[string]any, map[string]any, error) {
	op, root, err := structuredBodyPlanRoot(bundle, operation)
	if err != nil {
		return OperationSpec{}, nil, nil, err
	}
	compiled, err := compileStructuredRESTBodySchema(op)
	if err != nil {
		return OperationSpec{}, nil, nil, err
	}
	canonical, err := canonicalizeStructuredRESTBodyFragment(compiled, op, body, "body")
	if err != nil {
		return OperationSpec{}, nil, nil, err
	}
	return op, root, canonical, nil
}

func structuredBodyPlanRoot(bundle Bundle, operation string) (OperationSpec, map[string]any, error) {
	op, _, err := operationDirectWriteSpec(bundle, operation)
	if err != nil {
		return OperationSpec{}, nil, err
	}
	if op.Kind != "rest_write" || !OperationDirectWriteHasStructuredRESTBody(op) {
		return OperationSpec{}, nil, fmt.Errorf("operation %q does not expose a structured REST body", operation)
	}
	root, err := operationDirectWriteBodySchemaRoot(op)
	if err != nil {
		return OperationSpec{}, nil, err
	}
	return op, root, nil
}

func structuredBodyRelativePath(path string) string {
	return strings.TrimPrefix(strings.TrimSpace(path), "body.")
}

func deleteStructuredBodyValue(current any, steps []operationDirectWriteBodyPathStep, path string) (bool, error) {
	if len(steps) == 0 {
		return false, nil
	}
	step := steps[0]
	if step.array {
		items, ok := current.([]any)
		if !ok {
			return false, fmt.Errorf("body field %q conflicts with existing non-array value", path)
		}
		if step.index < 0 || step.index >= len(items) || items[step.index] == nil {
			return false, nil
		}
		if len(steps) == 1 {
			items[step.index] = nil
			return true, nil
		}
		return deleteStructuredBodyValue(items[step.index], steps[1:], path)
	}
	object, ok := current.(map[string]any)
	if !ok {
		return false, fmt.Errorf("body field %q conflicts with existing non-object value", path)
	}
	value, found := object[step.key]
	if !found || value == nil {
		return false, nil
	}
	if len(steps) == 1 {
		delete(object, step.key)
		return true, nil
	}
	return deleteStructuredBodyValue(value, steps[1:], path)
}

func mergeStructuredBodyMaps(base, overlay map[string]any, path string) (map[string]any, error) {
	result := cloneAnyMap(base)
	for name, overlayValue := range overlay {
		if existing, found := result[name]; found {
			left, leftObject := existing.(map[string]any)
			right, rightObject := overlayValue.(map[string]any)
			if leftObject || rightObject {
				if !leftObject || !rightObject {
					return nil, fmt.Errorf("%s.%s conflicts with an object replacement", path, name)
				}
				merged, err := mergeStructuredBodyMaps(left, right, path+"."+name)
				if err != nil {
					return nil, err
				}
				result[name] = merged
				continue
			}
		}
		result[name] = cloneJSONValue(overlayValue)
	}
	return result, nil
}
