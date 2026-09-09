package engine

import (
	"fmt"
	"math/big"
	"strings"

	"polymetrics.ai/internal/connectors"
)

func validateRequestInputCommands(surface *CLISurface, streams []StreamSpec, operations []OperationSpec) error {
	if surface == nil {
		return nil
	}
	for _, flag := range surface.GlobalFlags {
		if flag.InputCodec != "" {
			return fmt.Errorf("source input codec requires a selected command consumer")
		}
	}
	for _, command := range surface.Commands {
		var plan *compiledRequestInputPlan
		for _, stream := range streams {
			if command.Stream == stream.Name {
				plan = stream.inputPlan
			}
		}
		for _, op := range operations {
			if command.Operation == op.ID && op.REST != nil {
				plan = op.REST.inputPlan
			}
		}
		for _, flag := range command.Flags {
			binding, found := requestInputFlagBinding(plan, flag.MapsTo)
			if flag.InputCodec == "" && !found {
				continue
			}
			if found && binding.In == "query" && selectedStructuredQuery(plan, binding.Name) {
				if flag.Format != "" || len(flag.Values) != 0 || flag.Minimum != nil || flag.Maximum != nil || flag.MinItems != 0 || flag.MaxItems != 0 {
					return fmt.Errorf("structured query flag constraints belong to selected input schema")
				}
				container := plan.schema.node.properties["query"]
				if flag.Type != "json" || flag.InputCodec != "source_structured_v1" || flag.Repeatable || flag.AllowBareString || flag.Required != containsRequestInputName(container.required, binding.Name) {
					return fmt.Errorf("structured query flag differs from selected input")
				}
				continue
			}
			if !found || plan == nil || flag.InputCodec != "source_scalar_v1" {
				return fmt.Errorf("command %q flag %q lacks its selected input codec/binding", command.Path, flag.Name)
			}
			if binding.In == "body" {
				return fmt.Errorf("named structured flags require their typed input codec")
			}
			container := plan.schema.node.properties[binding.In]
			node := container.properties[binding.Name]
			if len(node.types) != 1 || node.types[0] != flag.Type || flag.Repeatable || flag.Required != containsRequestInputName(container.required, binding.Name) {
				return fmt.Errorf("command %q flag %q differs from selected input type or presence", command.Path, flag.Name)
			}
			if flag.Format != node.format || !requestInputEnumEqual(flag.Values, node.enum) {
				return fmt.Errorf("command %q flag %q differs from selected input format or enum", command.Path, flag.Name)
			}
			if !requestInputBoundEqual(flag.Minimum, node.minimum, node.hasMinimum) || !requestInputBoundEqual(flag.Maximum, node.maximum, node.hasMaximum) {
				return fmt.Errorf("command %q flag %q differs from selected input bounds", command.Path, flag.Name)
			}
		}
	}
	return nil
}

func requestInputFlagBinding(plan *compiledRequestInputPlan, mapping string) (RequestInputBinding, bool) {
	if plan == nil {
		return RequestInputBinding{}, false
	}
	location, name, ok := strings.Cut(mapping, ".")
	if !ok {
		return RequestInputBinding{}, false
	}
	for _, binding := range plan.bindings {
		if location == "config" && binding.ConfigKey == name || location == binding.In && binding.Name == name {
			return binding, true
		}
	}
	return RequestInputBinding{}, false
}

func requestInputBoundEqual(bound *connectors.ExactNumber, value *big.Rat, present bool) bool {
	if !present {
		return bound == nil
	}
	if bound == nil || value == nil {
		return false
	}
	parsed, ok := connectors.ParseExactJSONNumber(bound.String())
	return ok && parsed.Cmp(value) == 0
}

func requestInputEnumEqual(values []string, expected []any) bool {
	if len(values) != len(expected) {
		return false
	}
	seen := map[string]bool{}
	for _, value := range values {
		if seen[value] {
			return false
		}
		seen[value] = true
		found := false
		for _, want := range expected {
			if text, err := requestInputScalarText(want); err == nil && text == value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
