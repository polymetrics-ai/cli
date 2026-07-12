package agentloop

import (
	"reflect"
	"strings"
)

func validateFixtureStrings(fixture Fixture) error {
	return inspectFixtureValue(reflect.ValueOf(fixture))
}

func inspectFixtureValue(value reflect.Value) error {
	if !value.IsValid() {
		return nil
	}
	if value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
		if value.IsNil() {
			return nil
		}
		return inspectFixtureValue(value.Elem())
	}
	switch value.Kind() {
	case reflect.String:
		return inspectFixtureString(value.String())
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if err := inspectFixtureValue(value.Field(i)); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if err := inspectFixtureValue(value.Index(i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func inspectFixtureString(value string) error {
	if len(value) > maxFixtureString {
		return validationError("FIXTURE_STRING_TOO_LARGE", "fixture string exceeds the limit")
	}
	lower := strings.ToLower(value)
	for _, fragment := range []string{"raw command:", "raw prompt:", "/sessions/", "\\sessions\\", ".jsonl"} {
		if strings.Contains(lower, fragment) {
			return validationError("FIXTURE_FORBIDDEN_CONTENT", "fixture contains a forbidden content surface")
		}
	}
	for _, fragment := range []string{"authorization:", "bearer ", "private key"} {
		if strings.Contains(lower, fragment) {
			return validationError("FIXTURE_SENSITIVE_DATA", "fixture contains sensitive-shaped content")
		}
	}
	if hasSensitiveAssignment(value) {
		return validationError("FIXTURE_SENSITIVE_DATA", "fixture contains sensitive-shaped content")
	}
	return nil
}

func hasSensitiveAssignment(value string) bool {
	lower := strings.ToLower(value)
	for i := 0; i < len(lower); i++ {
		if lower[i] != '=' && lower[i] != ':' {
			continue
		}
		if strings.TrimSpace(lower[i+1:]) == "" {
			continue
		}
		end := i
		for end > 0 && lower[end-1] == ' ' {
			end--
		}
		start := end
		for start > 0 && isIdentifierByte(lower[start-1]) {
			start--
		}
		key := lower[start:end]
		if key == "token" || key == "secret" || key == "password" || key == "key" ||
			strings.HasSuffix(key, "token") || strings.HasSuffix(key, "secret") ||
			strings.HasSuffix(key, "password") || strings.HasSuffix(key, "key") {
			return true
		}
	}
	return false
}

func isIdentifierByte(value byte) bool {
	return value == '_' || value >= 'a' && value <= 'z' || value >= '0' && value <= '9'
}
