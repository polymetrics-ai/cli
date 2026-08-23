package engine

import (
	"fmt"
	"strings"
)

type sensitiveTransformExecutor struct {
	version string
	apply   func(map[string]any) (map[string]any, error)
}

var sensitiveTransformRegistry = map[string]sensitiveTransformExecutor{
	"none": {
		version: "v1",
		apply: func(body map[string]any) (map[string]any, error) {
			return body, nil
		},
	},
}

func operationSensitiveTransform(op OperationSpec) (string, sensitiveTransformExecutor, bool, error) {
	if op.SensitivePolicy == nil {
		return "", sensitiveTransformExecutor{}, false, nil
	}
	name := strings.ToLower(strings.TrimSpace(op.SensitivePolicy.Transform))
	if name == "" {
		name = "none"
	}
	executor, ok := sensitiveTransformRegistry[name]
	if !ok {
		return "", sensitiveTransformExecutor{}, false, fmt.Errorf("operation %q sensitive transform %q is not registered for execution", op.ID, name)
	}
	return name, executor, true, nil
}

func applyOperationSensitiveTransform(op OperationSpec, body map[string]any) (map[string]any, error) {
	_, executor, present, err := operationSensitiveTransform(op)
	if err != nil || !present {
		return body, err
	}
	transformed, err := executor.apply(body)
	if err != nil {
		return nil, fmt.Errorf("operation %q apply sensitive transform: %w", op.ID, err)
	}
	return transformed, nil
}

func bindOperationSensitiveTransform(definition map[string]any, op OperationSpec) error {
	name, executor, present, err := operationSensitiveTransform(op)
	if err != nil || !present {
		return err
	}
	definition["sensitive_transform"] = map[string]any{"name": name, "version": executor.version}
	return nil
}
