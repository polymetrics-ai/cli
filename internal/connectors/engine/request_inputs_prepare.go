package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// prepareReadInputs validates a detached selected-consumer snapshot. It never
// resolves credentials, constructs runtime clients, or marks a caller trusted.
func prepareReadInputs(stream StreamSpec, req connectors.ReadRequest) (StreamSpec, connectors.ReadRequest, error) {
	if stream.RequestInputs == nil {
		return stream, req, nil
	}
	plan := stream.inputPlan
	if plan == nil {
		return stream, req, fmt.Errorf("request input contract is not compiled")
	}
	req.Config.Config = maps.Clone(req.Config.Config)
	if req.Config.Config == nil {
		req.Config.Config = map[string]string{}
	}
	req.Query = maps.Clone(req.Query)
	envelope := map[string]any{"path": map[string]any{}, "query": map[string]any{}, "header": map[string]any{}}
	knownQuery := map[string]bool{}
	for _, binding := range plan.bindings {
		if binding.In == "body" {
			return stream, req, fmt.Errorf("named body input preparation requires its typed consumer")
		}
		node := plan.schema.node.properties[binding.In].properties[binding.Name]
		if len(node.types) != 1 {
			return stream, req, fmt.Errorf("flat request input requires an unambiguous scalar type")
		}
		raw, present := req.Config.Config[binding.ConfigKey]
		if binding.ConfigKey == "" {
			present = false
		}
		if binding.In == "query" {
			knownQuery[binding.Name] = true
			if override, exists := req.Query[binding.Name]; exists {
				raw, present = override, true
			}
		}
		if !present && node.hasDefault {
			var err error
			raw, err = requestInputScalarText(node.defaultVal)
			if err != nil {
				return stream, req, err
			}
			present = true
		}
		if !present {
			continue
		}
		value, err := connectors.DecodeSourceScalar(node.types[0], raw, 1<<20)
		if err != nil {
			return stream, req, fmt.Errorf("request input %s/%s: %w", binding.In, binding.Name, err)
		}
		if binding.In == "path" && strings.TrimSpace(raw) == "" {
			return stream, req, fmt.Errorf("path request input must be nonempty")
		}
		envelope[binding.In].(map[string]any)[binding.Name] = value
		if binding.ConfigKey != "" {
			req.Config.Config[binding.ConfigKey] = raw
		}
	}
	for name := range req.Query {
		if !knownQuery[name] {
			return stream, req, fmt.Errorf("undeclared request query input %q", name)
		}
	}
	if plan.schema.node.properties["body"] != nil {
		// The body placement consumer receives a detached typed body. Template and
		// named-member preparation are handled before exposing that capability.
		if len(stream.Body) > 0 {
			return stream, req, fmt.Errorf("request body requires typed preparation")
		}
	}
	if err := plan.schema.Validate(envelope); err != nil {
		return stream, req, fmt.Errorf("request inputs: %w", err)
	}
	return stream, req, nil
}

func requestInputScalarText(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case bool:
		return strconv.FormatBool(typed), nil
	case json.Number:
		return typed.String(), nil
	default:
		return "", fmt.Errorf("request input default has no lossless flat scalar representation")
	}
}

func prepareConnectorReadInputs(bundle Bundle, req connectors.ReadRequest) (StreamSpec, connectors.ReadRequest, error) {
	stream, err := findStream(bundle, req.Stream)
	if err != nil {
		return stream, req, err
	}
	req.Config = materializeConfigDefaults(bundle, req.Config)
	return prepareReadInputs(stream, req)
}

// ValidateReadInputs implements the optional pre-secret App contract using the
// same selected plan as the exported readers.
func (c *Connector) ValidateReadInputs(ctx context.Context, req connectors.ReadInputValidationRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, _, err := prepareConnectorReadInputs(c.bundle, connectors.ReadRequest{Stream: req.Stream, Config: connectors.RuntimeConfig{Config: req.Config}, Query: req.Query})
	return err
}

func prepareOperationReadInputs(op OperationSpec, req connectors.OperationDirectReadRequest) (connectors.OperationDirectReadRequest, error) {
	if op.REST == nil || op.REST.RequestInputs == nil {
		return req, nil
	}
	plan := op.REST.inputPlan
	if plan == nil {
		return req, fmt.Errorf("request input contract is not compiled")
	}
	req.PathParams = maps.Clone(req.PathParams)
	if req.PathParams == nil {
		req.PathParams = map[string]string{}
	}
	req.Query = maps.Clone(req.Query)
	if req.Query == nil {
		req.Query = map[string]string{}
	}
	req.Headers = maps.Clone(req.Headers)
	if req.Headers == nil {
		req.Headers = map[string]string{}
	}
	// Both direct and saved adapters feed the same selected scalar preparation.
	// Aliases remain local implementation coordinates, never new wire names.
	config := map[string]string{}
	known := map[string]map[string]bool{"path": {}, "query": {}, "header": {}}
	for _, binding := range plan.bindings {
		if binding.In == "body" {
			return req, fmt.Errorf("named body input preparation requires its typed consumer")
		}
		if binding.ConfigKey == "" {
			return req, fmt.Errorf("request input parameter lacks its generated alias")
		}
		known[binding.In][binding.Name] = true
		values := req.Query
		if binding.In == "path" {
			values = req.PathParams
		}
		if binding.In == "header" {
			values = req.Headers
		}
		if value, present := values[binding.Name]; present {
			config[binding.ConfigKey] = value
		}
	}
	for location, values := range map[string]map[string]string{"path": req.PathParams, "query": req.Query, "header": req.Headers} {
		for name := range values {
			if !known[location][name] {
				return req, fmt.Errorf("undeclared %s request input", location)
			}
		}
	}
	if len(req.HeaderValues) != 0 {
		return req, fmt.Errorf("scalar request inputs do not accept repeated header values")
	}
	_, prepared, err := prepareReadInputs(StreamSpec{RequestInputs: op.REST.RequestInputs, inputPlan: plan}, connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: config}, Query: req.Query})
	if err != nil {
		return req, err
	}
	for _, binding := range plan.bindings {
		value, present := prepared.Config.Config[binding.ConfigKey]
		if !present {
			continue
		}
		switch binding.In {
		case "path":
			req.PathParams[binding.Name] = value
		case "query":
			req.Query[binding.Name] = value
		case "header":
			req.Headers[binding.Name] = value
		}
	}
	return req, nil
}
