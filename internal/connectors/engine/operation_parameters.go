package engine

import (
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
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
	for _, parameter := range operationParameters(op) {
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

func operationBinaryDownloadPathParameterNames(op OperationSpec) (map[string]struct{}, error) {
	if op.Binary == nil {
		return nil, fmt.Errorf("operation %q has no binary declaration", op.ID)
	}
	path := op.Binary.Path
	remaining := surfacePathVarPattern.ReplaceAllString(path, "")
	if strings.ContainsAny(remaining, "{}") {
		return nil, fmt.Errorf("operation %q has malformed binary path template %q", op.ID, path)
	}
	names := make(map[string]struct{})
	for _, match := range surfacePathVarPattern.FindAllStringSubmatch(path, -1) {
		if len(match) != 2 {
			continue
		}
		name := match[1]
		if err := safety.ValidateIdentifier(name, "binary path parameter"); err != nil {
			return nil, fmt.Errorf("operation %q binary path parameter: %w", op.ID, err)
		}
		names[name] = struct{}{}
	}
	return names, nil
}

func materializeOperationBinaryDownloadPathParams(op OperationSpec, cfg connectors.RuntimeConfig, pathParams map[string]string) (map[string]string, error) {
	declared, err := operationBinaryDownloadPathParameterNames(op)
	if err != nil {
		return nil, err
	}
	return materializeOperationPathParams(op, declared, cfg, pathParams)
}

func materializeOperationPathParams(op OperationSpec, declared map[string]struct{}, cfg connectors.RuntimeConfig, pathParams map[string]string) (map[string]string, error) {
	parameters, err := operationParametersForLocation(op, "path")
	if err != nil {
		return nil, err
	}
	fields := make([]string, 0, len(pathParams))
	for field := range pathParams {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	for _, field := range fields {
		if err := safety.ValidateIdentifier(field, "operation path parameter"); err != nil {
			return nil, fmt.Errorf("operation %q path parameter: %w", op.ID, err)
		}
		if _, ok := declared[field]; !ok {
			return nil, fmt.Errorf("operation %q path parameter %q is not declared", op.ID, field)
		}
	}
	names := make([]string, 0, len(declared))
	for name := range declared {
		names = append(names, name)
	}
	sort.Strings(names)
	effective := make(map[string]string, len(names))
	for _, field := range names {
		value, provided := pathParams[field]
		if !provided || value == "" {
			value, provided = cfg.Config[field]
		}
		if !provided || value == "" {
			return nil, fmt.Errorf("missing path variable %q", field)
		}
		parameter, ok := parameters[field]
		if !ok {
			parameter = OperationParameter{Name: field, In: "path", Type: "string"}
		}
		if err := validateOperationParameterWireValue(op, parameter, "path", value); err != nil {
			return nil, err
		}
		effective[field] = value
	}
	return effective, nil
}

func operationBinaryDownloadQuery(op OperationSpec, requested map[string]string) (map[string]string, error) {
	parameters, err := operationParametersForLocation(op, "query")
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(requested))
	for name := range requested {
		names = append(names, name)
	}
	sort.Strings(names)
	provided := make(map[string]struct{}, len(names))
	query := make(map[string]string, len(names))
	for _, name := range names {
		if err := safety.ValidateIdentifier(name, "binary query parameter"); err != nil {
			return nil, fmt.Errorf("operation %q binary query parameter: %w", op.ID, err)
		}
		parameter, ok := parameters[name]
		if !ok {
			return nil, fmt.Errorf("operation %q binary query parameter %q is not source-declared in binary.parameters", op.ID, name)
		}
		if err := validateOperationParameterWireValue(op, parameter, "query", requested[name]); err != nil {
			return nil, err
		}
		provided[name] = struct{}{}
		query[name] = requested[name]
	}
	for name, parameter := range parameters {
		if parameter.Required {
			if _, ok := provided[name]; !ok {
				return nil, fmt.Errorf("operation %q requires query parameter %q", op.ID, name)
			}
		}
	}
	return query, nil
}

func validateOperationParameterWireValue(op OperationSpec, parameter OperationParameter, location, value string) error {
	if location == "query" && parameter.Required && strings.TrimSpace(value) == "" {
		return fmt.Errorf("operation %q requires non-blank query parameter %q", op.ID, parameter.Name)
	}
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
