package engine

import (
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/safety"
)

const (
	defaultOperationParameterMaxBytes = 4096
	maxOperationParameterMaxBytes     = 64 << 10
)

func operationParameterByteCap(parameter OperationParameter) int {
	if parameter.MaxBytes <= 0 {
		return defaultOperationParameterMaxBytes
	}
	if parameter.MaxBytes > maxOperationParameterMaxBytes {
		return maxOperationParameterMaxBytes
	}
	return parameter.MaxBytes
}

func operationParametersForLocation(op OperationSpec, location string) (map[string]OperationParameter, error) {
	parameters := make(map[string]OperationParameter)
	if op.REST == nil {
		return parameters, nil
	}
	for _, parameter := range op.REST.Parameters {
		if !strings.EqualFold(strings.TrimSpace(parameter.In), location) {
			continue
		}
		name := strings.TrimSpace(parameter.Name)
		if name == "" || name != parameter.Name {
			return nil, fmt.Errorf("operation %q rest.parameters has invalid %s parameter name %q", op.ID, location, parameter.Name)
		}
		if _, duplicate := parameters[name]; duplicate {
			return nil, fmt.Errorf("operation %q rest.parameters duplicates %s parameter %q", op.ID, location, name)
		}
		parameters[name] = parameter
	}
	return parameters, nil
}

func validateOperationParameterWireValue(op OperationSpec, parameter OperationParameter, location, value string) error {
	if err := safety.RejectDangerousChars(value, location+" parameter "+parameter.Name); err != nil {
		return err
	}
	switch strings.ToLower(strings.TrimSpace(parameter.Type)) {
	case "", "string", "enum":
	case "boolean":
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("operation %q %s parameter %q must be boolean", op.ID, location, parameter.Name)
		}
	case "integer":
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("operation %q %s parameter %q must be an integer", op.ID, location, parameter.Name)
		}
	case "number":
		number, err := strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
			return fmt.Errorf("operation %q %s parameter %q must be a number", op.ID, location, parameter.Name)
		}
	default:
		return fmt.Errorf("operation %q %s parameter %q has unsupported type %q", op.ID, location, parameter.Name, parameter.Type)
	}
	if len(parameter.Values) > 0 {
		allowed := append([]string(nil), parameter.Values...)
		sort.Strings(allowed)
		matched := false
		for _, candidate := range allowed {
			if value == candidate {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("operation %q %s parameter %q must be one of %s", op.ID, location, parameter.Name, strings.Join(allowed, "|"))
		}
	}

	encoded := ""
	switch location {
	case "path":
		var err error
		encoded, err = encodeSurfacePathValue(parameter.Name, value)
		if err != nil {
			return err
		}
	case "query":
		encoded = url.QueryEscape(value)
	default:
		encoded = value
	}
	capBytes := operationParameterByteCap(parameter)
	if len(encoded) > capBytes {
		return fmt.Errorf("operation %q %s parameter %q encoded value exceeds byte cap %d", op.ID, location, parameter.Name, capBytes)
	}
	return nil
}

func validateConservativeOperationParameterBytes(location, name, encoded string) error {
	if len(encoded) > defaultOperationParameterMaxBytes {
		return fmt.Errorf("%s parameter %q encoded value exceeds byte cap %d", location, name, defaultOperationParameterMaxBytes)
	}
	return nil
}
