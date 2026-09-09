package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// prepareReadInputs validates a detached selected-consumer snapshot. It never
// resolves credentials, constructs runtime clients, or marks a caller trusted.
func prepareReadInputs(stream StreamSpec, req connectors.ReadRequest, pagination *PaginationSpec) (StreamSpec, connectors.ReadRequest, error) {
	if stream.RequestInputs == nil {
		return stream, req, nil
	}
	plan := stream.inputPlan
	if plan == nil || plan.schema == nil || plan.schema.node == nil {
		return stream, req, fmt.Errorf("request input contract is not compiled")
	}
	req.Config.Config = maps.Clone(req.Config.Config)
	if req.Config.Config == nil {
		req.Config.Config = map[string]string{}
	}
	req.Query = maps.Clone(req.Query)
	envelope := map[string]any{"path": map[string]any{}, "query": map[string]any{}, "header": map[string]any{}}
	knownQuery := map[string]bool{}
	scalarBytes := 0
	stream.Headers = maps.Clone(stream.Headers)
	var body map[string]any
	if plan.schema.node.properties["body"] != nil {
		var err error
		if stream.preparedReadBodyPresent {
			var ok bool
			body, ok = copyRecordValue(stream.preparedReadBody).(map[string]any)
			if !ok {
				return stream, req, fmt.Errorf("prepared request body must be an object")
			}
		} else {
			body, err = resolveStreamBodyMap(stream.Body, requestVars(req.Config, nil, "", req.Query))
		}
		if err != nil {
			return stream, req, err
		}
		if body == nil {
			body = map[string]any{}
		}
	}
	for _, binding := range plan.bindings {
		node := plan.schema.node.properties[binding.In]
		if binding.In == "body" {
			for _, part := range strings.Split(binding.Pointer[1:], "/") {
				if node == nil {
					break
				}
				node = node.properties[part]
			}
		} else if node != nil {
			node = node.properties[binding.Name]
		}
		if node == nil {
			return stream, req, fmt.Errorf("request input binding lacks its compiled schema")
		}
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
			if binding.In == "header" {
				delete(stream.Headers, binding.Name)
			}
			continue
		}
		if len(raw) > maxOperationDirectReadBytes-scalarBytes {
			return stream, req, fmt.Errorf("request inputs exceed aggregate preparation budget")
		}
		scalarBytes += len(raw)
		value, err := connectors.DecodeSourceScalar(node.types[0], raw, 1<<20)
		if err != nil {
			return stream, req, fmt.Errorf("request input %s/%s: %w", binding.In, binding.Name, err)
		}
		if binding.In == "path" && strings.TrimSpace(raw) == "" {
			return stream, req, fmt.Errorf("path request input must be nonempty")
		}
		if binding.In == "body" {
			target := body
			parts := strings.Split(binding.Pointer[1:], "/")
			for _, part := range parts[:len(parts)-1] {
				child, exists := target[part]
				if !exists {
					child = map[string]any{}
					target[part] = child
				}
				object, ok := child.(map[string]any)
				if !ok {
					return stream, req, fmt.Errorf("named body input crosses a non-object")
				}
				target = object
			}
			target[parts[len(parts)-1]] = value
		} else {
			envelope[binding.In].(map[string]any)[binding.Name] = value
		}
		if binding.In == "header" {
			if len(raw) > maxOperationParameterMaxBytes {
				return stream, req, fmt.Errorf("request header exceeds byte bound")
			}
			if _, err := canonicalPreparedRequestHeaders(map[string]string{binding.Name: raw}); err != nil {
				return stream, req, fmt.Errorf("request header value is invalid")
			}
			if stream.Headers == nil {
				stream.Headers = map[string]string{}
			}
			stream.Headers[binding.Name] = raw
		}
		if binding.ConfigKey != "" {
			req.Config.Config[binding.ConfigKey] = raw
		}
	}
	for name := range req.Query {
		if !knownQuery[name] {
			return stream, req, fmt.Errorf("undeclared request query input %q", name)
		}
	}
	if body != nil {
		envelope["body"] = body
		stream.preparedReadBody = body
		stream.preparedReadBodyPresent = true
	}
	root := *plan.schema.node
	if bodyNode := root.properties["body"]; bodyNode != nil && pagination != nil && hasPaginationBody(*pagination) {
		// Only engine-owned body positions can be absent until pagination composes
		// them. All supplied values and all caller-owned required fields still validate.
		copyBody := *bodyNode
		copyBody.required = nil
		for _, name := range bodyNode.required {
			if name != pagination.BodyCursorField && name != pagination.BodyLimitField && name != pagination.BodyOffsetField && name != pagination.BodyPageField {
				copyBody.required = append(copyBody.required, name)
			}
		}
		root.properties = maps.Clone(root.properties)
		root.properties["body"] = &copyBody
	}
	if err := root.validate(envelope, ""); err != nil {
		return stream, req, fmt.Errorf("request inputs violate the selected schema")
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
	pagination := stream.Pagination
	if pagination == nil {
		pagination = bundle.HTTP.Pagination
	}
	return prepareReadInputs(stream, req, pagination)
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
	if plan == nil || plan.schema == nil || plan.schema.node == nil {
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
	_, prepared, err := prepareReadInputs(StreamSpec{RequestInputs: op.REST.RequestInputs, inputPlan: plan}, connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: config}, Query: req.Query}, op.REST.Pagination)
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

// Validate the effective placed values after the paginator and caller merge,
// immediately before selecting a requester. No query is rebuilt here.
func validateEffectiveReadInputs(stream StreamSpec, req connectors.ReadRequest, query url.Values, body any) error {
	if stream.RequestInputs == nil {
		return nil
	}
	plan := stream.inputPlan
	if plan == nil || plan.schema == nil {
		return fmt.Errorf("request input contract is not compiled")
	}
	envelope := map[string]any{"path": map[string]any{}, "query": map[string]any{}, "header": map[string]any{}}
	for _, binding := range plan.bindings {
		if binding.In == "body" {
			continue
		}
		raw, present := req.Config.Config[binding.ConfigKey]
		if binding.In == "query" {
			values, exists := query[binding.Name]
			present = exists
			if exists {
				if len(values) != 1 {
					return fmt.Errorf("scalar request input has repeated values")
				}
				raw = values[0]
			}
		}
		if binding.In == "header" {
			raw, present = stream.Headers[binding.Name]
		}
		if !present {
			continue
		}
		node := plan.schema.node.properties[binding.In].properties[binding.Name]
		if node == nil || len(node.types) != 1 {
			return fmt.Errorf("effective request input lacks its scalar type")
		}
		value, err := connectors.DecodeSourceScalar(node.types[0], raw, 1<<20)
		if err != nil {
			return fmt.Errorf("effective request input is invalid")
		}
		envelope[binding.In].(map[string]any)[binding.Name] = value
	}
	for name := range query {
		if _, exists := plan.schema.node.properties["query"].properties[name]; !exists {
			return fmt.Errorf("effective request contains undeclared query input")
		}
	}
	if body != nil {
		envelope["body"] = body
	}
	if err := plan.schema.Validate(envelope); err != nil {
		return fmt.Errorf("effective request inputs violate the selected schema")
	}
	return nil
}
