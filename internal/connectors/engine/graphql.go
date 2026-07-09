package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type omittedGraphQLVariable struct{}

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
		if _, omitted := resolved.(omittedGraphQLVariable); omitted {
			continue
		}
		out[k] = resolved
	}
	return out, nil
}

// ResolveCheckGraphQLVariables statically validates every templated GraphQL
// variable value against the same interpolation namespace rules used by
// ordinary stream/write templates. Constant strings without "{{ }}" are no-ops.
func ResolveCheckGraphQLVariables(in map[string]any, specKeys map[string]bool) error {
	for name, value := range in {
		if err := resolveCheckGraphQLValue(value, specKeys); err != nil {
			return fmt.Errorf("graphql variable %q: %w", name, err)
		}
	}
	return nil
}

func resolveCheckGraphQLValue(v any, specKeys map[string]bool) error {
	switch t := v.(type) {
	case string:
		return ResolveCheck(t, specKeys)
	case map[string]any:
		if template, ok := t["template"]; ok {
			tmpl, _ := template.(string)
			return ResolveCheck(tmpl, specKeys)
		}
		if source, ok := t["source"]; ok {
			ref, _ := source.(string)
			return ResolveCheck("{{ "+ref+" }}", specKeys)
		}
		for _, child := range t {
			if err := resolveCheckGraphQLValue(child, specKeys); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range t {
			if err := resolveCheckGraphQLValue(child, specKeys); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolveGraphQLValue(v any, vars Vars) (any, error) {
	switch t := v.(type) {
	case string:
		return Interpolate(t, vars)
	case map[string]any:
		if _, ok := t["template"]; ok {
			return resolveGraphQLVariableTemplate(t, vars)
		}
		if _, ok := t["source"]; ok {
			return resolveGraphQLVariableSource(t, vars)
		}
		out := make(map[string]any, len(t))
		for k, child := range t {
			resolved, err := resolveGraphQLValue(child, vars)
			if err != nil {
				return nil, err
			}
			if _, omitted := resolved.(omittedGraphQLVariable); omitted {
				continue
			}
			out[k] = resolved
		}
		return out, nil
	case []any:
		out := make([]any, 0, len(t))
		for _, child := range t {
			resolved, err := resolveGraphQLValue(child, vars)
			if err != nil {
				return nil, err
			}
			if _, omitted := resolved.(omittedGraphQLVariable); omitted {
				continue
			}
			out = append(out, resolved)
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
	omit, _ := spec["omit_when_empty"].(bool)
	if err != nil {
		if def, ok := spec["default"].(string); ok && isUnresolvedGraphQLDefaultable(err) {
			resolved = def
		} else if omit && isUnresolvedGraphQLDefaultable(err) {
			return omittedGraphQLVariable{}, nil
		} else {
			return nil, err
		}
	}
	if omit && resolved == "" {
		return omittedGraphQLVariable{}, nil
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

func resolveGraphQLVariableSource(spec map[string]any, vars Vars) (any, error) {
	source, ok := spec["source"].(string)
	if !ok {
		return nil, fmt.Errorf("graphql variable source must be a string")
	}
	raw, err := resolveGraphQLSourceValue(source, vars)
	omit, _ := spec["omit_when_empty"].(bool)
	if err != nil {
		if omit && isUnresolvedGraphQLDefaultable(err) {
			return omittedGraphQLVariable{}, nil
		}
		return nil, err
	}
	if omit && isEmptyGraphQLSourceValue(raw) {
		return omittedGraphQLVariable{}, nil
	}
	if err := rejectGraphQLSourceControlChars(raw); err != nil {
		return nil, err
	}
	return coerceGraphQLSourceValue(raw, typeOfGraphQLVariable(spec))
}

func resolveGraphQLSourceValue(source string, vars Vars) (any, error) {
	ref := strings.TrimSpace(source)
	if ref == "" || strings.ContainsAny(ref, "{}|") {
		return nil, fmt.Errorf("graphql variable source %q must be a plain reference", source)
	}
	if ref == "cursor" {
		return vars.Cursor, nil
	}
	parts := strings.Split(ref, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("graphql variable source %q must be namespaced", source)
	}
	namespace, path := parts[0], parts[1:]
	switch namespace {
	case "record":
		return resolveRecordPathValue(vars.Record, path)
	case "config":
		v, ok := vars.Config[path[0]]
		if !ok {
			return nil, &unresolvedKeyError{Namespace: "config", Key: path[0]}
		}
		return v, nil
	case "query":
		v, ok := vars.Query[path[0]]
		if !ok {
			return nil, &unresolvedKeyError{Namespace: "query", Key: path[0]}
		}
		return v, nil
	case "incremental":
		if path[0] == "lower_bound" {
			return vars.IncrementalLowerBound, nil
		}
		return nil, &unresolvedKeyError{Namespace: "incremental", Key: path[0]}
	default:
		return nil, fmt.Errorf("graphql variable source %q uses unsupported namespace %q", source, namespace)
	}
}

func isEmptyGraphQLSourceValue(v any) bool {
	if v == nil {
		return true
	}
	switch t := v.(type) {
	case string:
		return t == ""
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() == 0
	default:
		return false
	}
}

func rejectGraphQLSourceControlChars(v any) error {
	switch t := v.(type) {
	case string:
		if strings.ContainsAny(t, "\r\n") {
			return fmt.Errorf("graphql variable source value contains CR/LF")
		}
	case []any:
		for _, item := range t {
			if err := rejectGraphQLSourceControlChars(item); err != nil {
				return err
			}
		}
	case map[string]any:
		for _, item := range t {
			if err := rejectGraphQLSourceControlChars(item); err != nil {
				return err
			}
		}
	}
	return nil
}

func typeOfGraphQLVariable(spec map[string]any) string {
	typ, _ := spec["type"].(string)
	return typ
}

func coerceGraphQLSourceValue(raw any, typ string) (any, error) {
	switch typ {
	case "", "json":
		return raw, nil
	case "string":
		return stringify(raw), nil
	case "integer":
		return parseGraphQLInteger(raw)
	case "number":
		return parseGraphQLNumber(raw)
	case "boolean":
		return parseGraphQLBoolean(raw)
	case "string_array", "integer_array", "number_array", "boolean_array":
		return coerceGraphQLArray(raw, strings.TrimSuffix(typ, "_array"))
	default:
		return nil, fmt.Errorf("unsupported graphql variable type %q", typ)
	}
}

func parseGraphQLInteger(raw any) (int64, error) {
	switch t := raw.(type) {
	case int:
		return int64(t), nil
	case int64:
		return t, nil
	case float64:
		return int64(t), nil
	default:
		v, err := strconv.ParseInt(stringify(raw), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("graphql variable integer: %w", err)
		}
		return v, nil
	}
}

func parseGraphQLNumber(raw any) (float64, error) {
	switch t := raw.(type) {
	case int:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case float64:
		return t, nil
	default:
		v, err := strconv.ParseFloat(stringify(raw), 64)
		if err != nil {
			return 0, fmt.Errorf("graphql variable number: %w", err)
		}
		return v, nil
	}
}

func parseGraphQLBoolean(raw any) (bool, error) {
	switch t := raw.(type) {
	case bool:
		return t, nil
	default:
		v, err := strconv.ParseBool(stringify(raw))
		if err != nil {
			return false, fmt.Errorf("graphql variable boolean: %w", err)
		}
		return v, nil
	}
}

func coerceGraphQLArray(raw any, itemType string) ([]any, error) {
	items, err := graphQLArrayItems(raw)
	if err != nil {
		return nil, err
	}
	out := make([]any, 0, len(items))
	for _, item := range items {
		coerced, err := coerceGraphQLSourceValue(item, itemType)
		if err != nil {
			return nil, err
		}
		out = append(out, coerced)
	}
	return out, nil
}

func graphQLArrayItems(raw any) ([]any, error) {
	if raw == nil {
		return nil, nil
	}
	if s, ok := raw.(string); ok {
		if strings.TrimSpace(s) == "" {
			return nil, nil
		}
		parts := strings.Split(s, ",")
		out := make([]any, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
		return out, nil
	}
	rv := reflect.ValueOf(raw)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("graphql variable array: got %T", raw)
	}
	out := make([]any, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		out = append(out, rv.Index(i).Interface())
	}
	return out, nil
}

func isUnresolvedGraphQLDefaultable(err error) bool {
	var unresolved *unresolvedKeyError
	if !errors.As(err, &unresolved) {
		return false
	}
	switch unresolved.Namespace {
	case "config", "query", "record", "incremental":
		return true
	default:
		return false
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
