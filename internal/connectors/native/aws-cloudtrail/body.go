package awscloudtrail

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func buildActionBodyFromStrings(action string, raw map[string]string, requireRequired bool) (map[string]any, error) {
	values := make(map[string]any, len(raw))
	for k, v := range raw {
		values[k] = v
	}
	return buildActionBody(action, values, requireRequired)
}

func buildActionBody(action string, raw map[string]any, requireRequired bool) (map[string]any, error) {
	fields := cloudTrailActionFields[action]
	allowed := make(map[string]awsActionField, len(fields))
	for _, field := range fields {
		allowed[field.Name] = field
	}
	body := make(map[string]any, len(raw))
	for key, value := range raw {
		field, ok := allowed[key]
		if !ok {
			return nil, fmt.Errorf("aws-cloudtrail %s field %q is not in the official request schema", action, key)
		}
		coerced, err := coerceActionValue(field, value)
		if err != nil {
			return nil, fmt.Errorf("aws-cloudtrail %s field %q: %w", action, key, err)
		}
		if field.Name == "AdvancedEventSelectors" {
			coerced, err = normalizeAdvancedEventSelectors(coerced)
			if err != nil {
				return nil, fmt.Errorf("aws-cloudtrail %s field %q: %w", action, key, err)
			}
		}
		if err := validateActionField(action, field, coerced); err != nil {
			return nil, fmt.Errorf("aws-cloudtrail %s field %q: %w", action, key, err)
		}
		body[key] = coerced
	}
	if requireRequired {
		for _, field := range fields {
			if requiredActionField(action, field) {
				if value, ok := body[field.Name]; !ok || !requiredActionValuePresent(value) {
					return nil, fmt.Errorf("aws-cloudtrail %s requires field %s", action, field.Name)
				}
			}
		}
		if err := validateActionCrossFields(action, body); err != nil {
			return nil, err
		}
	}
	return body, nil
}

func coerceActionValue(field awsActionField, value any) (any, error) {
	if value == nil {
		return nil, nil
	}
	typeName := strings.ToLower(field.Type)
	switch {
	case strings.Contains(typeName, "boolean"):
		return coerceBool(value)
	case strings.Contains(typeName, "integer"):
		return coerceInt(value)
	case strings.Contains(typeName, "timestamp"):
		return coerceTimestamp(value)
	case strings.Contains(typeName, "array of strings"):
		return coerceStringArray(value)
	case strings.Contains(typeName, "array of"):
		return coerceJSONArray(value)
	case strings.Contains(typeName, "map") || strings.Contains(typeName, "object"):
		return coerceJSONObject(value)
	default:
		return stringifyValue(value), nil
	}
}

func coerceBool(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(v))
		if err != nil {
			return false, fmt.Errorf("want boolean")
		}
		return parsed, nil
	default:
		return false, fmt.Errorf("want boolean")
	}
}

func coerceInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		if v > math.MaxInt || v < math.MinInt {
			return 0, fmt.Errorf("integer out of range")
		}
		return int(v), nil
	case float64:
		if v != math.Trunc(v) || v > math.MaxInt || v < math.MinInt {
			return 0, fmt.Errorf("want integer")
		}
		return int(v), nil
	case json.Number:
		parsed, err := v.Int64()
		if err != nil {
			return 0, fmt.Errorf("want integer")
		}
		return coerceInt(parsed)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, fmt.Errorf("want integer")
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("want integer")
	}
}

func coerceTimestamp(value any) (any, error) {
	switch v := value.(type) {
	case time.Time:
		return v.Unix(), nil
	case int, int64, float64, json.Number:
		return coerceInt(v)
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil, fmt.Errorf("want timestamp")
		}
		if unix, err := strconv.Atoi(trimmed); err == nil {
			return unix, nil
		}
		if t, err := time.Parse(time.RFC3339, trimmed); err == nil {
			return t.Unix(), nil
		}
		if t, err := time.Parse("2006-01-02", trimmed); err == nil {
			return t.Unix(), nil
		}
		return nil, fmt.Errorf("want RFC3339, YYYY-MM-DD, or Unix timestamp")
	default:
		return nil, fmt.Errorf("want timestamp")
	}
}

func coerceStringArray(value any) ([]string, error) {
	switch v := value.(type) {
	case []string:
		return v, nil
	case []any:
		out := make([]string, 0, len(v))
		for _, elem := range v {
			out = append(out, stringifyValue(elem))
		}
		return out, nil
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil, nil
		}
		if strings.HasPrefix(trimmed, "[") {
			var arr []string
			if err := json.Unmarshal([]byte(trimmed), &arr); err != nil {
				return nil, fmt.Errorf("want JSON string array")
			}
			return arr, nil
		}
		parts := strings.Split(trimmed, ",")
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			if item := strings.TrimSpace(part); item != "" {
				out = append(out, item)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("want string array")
	}
}

func coerceJSONArray(value any) ([]any, error) {
	switch v := value.(type) {
	case []any:
		return v, nil
	case []map[string]any:
		out := make([]any, 0, len(v))
		for _, elem := range v {
			out = append(out, elem)
		}
		return out, nil
	case string:
		var arr []any
		if err := json.Unmarshal([]byte(strings.TrimSpace(v)), &arr); err != nil {
			return nil, fmt.Errorf("want JSON array")
		}
		return arr, nil
	default:
		return nil, fmt.Errorf("want JSON array")
	}
}

func coerceJSONObject(value any) (map[string]any, error) {
	switch v := value.(type) {
	case map[string]any:
		return v, nil
	case map[string]string:
		out := make(map[string]any, len(v))
		for k, val := range v {
			out[k] = val
		}
		return out, nil
	case string:
		var obj map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(v)), &obj); err != nil {
			return nil, fmt.Errorf("want JSON object")
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("want JSON object")
	}
}

func stringifyValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}
