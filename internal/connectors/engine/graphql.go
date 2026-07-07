package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func buildGraphQLPayload(spec *GraphQLRequestSpec, vars Vars) (map[string]any, error) {
	if err := validateGraphQLSpec(spec, ""); err != nil {
		return nil, err
	}

	payload := map[string]any{
		"query":         spec.Document,
		"operationName": spec.OperationName,
	}
	if len(spec.Variables) > 0 {
		resolved, err := resolveGraphQLVariables(spec.Variables, vars)
		if err != nil {
			return nil, err
		}
		payload["variables"] = resolved
	}
	return payload, nil
}

func resolveGraphQLVariables(in map[string]any, vars Vars) (map[string]any, error) {
	out := make(map[string]any, len(in))
	for k, v := range in {
		resolved, err := resolveGraphQLValue(v, vars)
		if err != nil {
			return nil, fmt.Errorf("resolve graphql variable %q: %w", k, err)
		}
		out[k] = resolved
	}
	return out, nil
}

func resolveGraphQLValue(v any, vars Vars) (any, error) {
	switch t := v.(type) {
	case string:
		return Interpolate(t, vars)
	case map[string]any:
		if _, ok := t["template"]; ok {
			return resolveGraphQLVariableTemplate(t, vars)
		}
		out := make(map[string]any, len(t))
		for k, child := range t {
			resolved, err := resolveGraphQLValue(child, vars)
			if err != nil {
				return nil, err
			}
			out[k] = resolved
		}
		return out, nil
	case []any:
		out := make([]any, len(t))
		for i, child := range t {
			resolved, err := resolveGraphQLValue(child, vars)
			if err != nil {
				return nil, err
			}
			out[i] = resolved
		}
		return out, nil
	default:
		return v, nil
	}
}

func resolveGraphQLVariableTemplate(spec map[string]any, vars Vars) (any, error) {
	tmpl, ok := spec["template"].(string)
	if !ok {
		return nil, fmt.Errorf("graphql variable template must be a string")
	}
	resolved, err := Interpolate(tmpl, vars)
	if err != nil {
		return nil, err
	}
	typ, _ := spec["type"].(string)
	switch typ {
	case "", "string":
		return resolved, nil
	case "integer":
		v, err := strconv.ParseInt(resolved, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("graphql variable integer: %w", err)
		}
		return v, nil
	case "number":
		v, err := strconv.ParseFloat(resolved, 64)
		if err != nil {
			return nil, fmt.Errorf("graphql variable number: %w", err)
		}
		return v, nil
	case "boolean":
		v, err := strconv.ParseBool(resolved)
		if err != nil {
			return nil, fmt.Errorf("graphql variable boolean: %w", err)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported graphql variable type %q", typ)
	}
}

func graphQLErrors(body []byte) error {
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}

	var envelope struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&envelope); err != nil {
		return nil
	}
	if len(envelope.Errors) == 0 {
		return nil
	}

	msgs := make([]string, 0, len(envelope.Errors))
	for _, item := range envelope.Errors {
		msg := strings.TrimSpace(item.Message)
		if msg == "" {
			msg = "error without message"
		}
		msgs = append(msgs, msg)
		if len(msgs) == 3 {
			break
		}
	}
	if len(envelope.Errors) > len(msgs) {
		msgs = append(msgs, fmt.Sprintf("%d more", len(envelope.Errors)-len(msgs)))
	}
	return fmt.Errorf("graphql errors: %s", strings.Join(msgs, "; "))
}
