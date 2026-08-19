// Package credential defines provider-neutral secret-input invariants shared by
// CLI, persistence, and request authentication boundaries.
package credential

import (
	"fmt"
	"sort"
	"strings"
)

// EmptySecretError reports a missing credential value without ever exposing a
// secret. Field identifies the non-secret credential field that needs input.
type EmptySecretError struct {
	Field string
}

func (e *EmptySecretError) Error() string {
	if e == nil || e.Field == "" {
		return "credential value is empty; supply a non-empty value"
	}
	return fmt.Sprintf("credential secret field %q is empty; supply a non-empty value", e.Field)
}

// NormalizeStdin removes at most one documented terminal transport delimiter
// from a stdin-supplied secret. It preserves every other byte, including
// leading/trailing whitespace and earlier newlines in multiline credentials.
func NormalizeStdin(value string) string {
	switch {
	case strings.HasSuffix(value, "\r\n"):
		return strings.TrimSuffix(value, "\r\n")
	case strings.HasSuffix(value, "\n"):
		return strings.TrimSuffix(value, "\n")
	default:
		return value
	}
}

// RequirePersistentValue rejects a value that has no credential bytes. It is
// intentionally byte-exact: whitespace is not trimmed or rewritten before
// encrypted persistence.
func RequirePersistentValue(field, value string) error {
	if value == "" {
		return &EmptySecretError{Field: field}
	}
	return nil
}

// RequirePersistentValues applies RequirePersistentValue to every supplied
// field in stable field-name order so a caller receives deterministic,
// non-secret validation when more than one value is empty.
func RequirePersistentValues(values map[string]string) error {
	fields := make([]string, 0, len(values))
	for field := range values {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	for _, field := range fields {
		if err := RequirePersistentValue(field, values[field]); err != nil {
			return err
		}
	}
	return nil
}

// RequireAuthenticationValue rejects a blank value before a shared auth helper
// can produce an empty request credential. Authentication helpers retain their
// existing surrounding-whitespace normalization for non-blank values.
func RequireAuthenticationValue(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return &EmptySecretError{Field: field}
	}
	return nil
}
