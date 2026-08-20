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

// InvalidSecretValueError reports credential bytes that cannot be safely sent
// in an HTTP header without exposing the credential value itself.
type InvalidSecretValueError struct {
	Field string
}

func (e *InvalidSecretValueError) Error() string {
	if e == nil || e.Field == "" {
		return "credential value contains prohibited header control characters; supply a valid value"
	}
	return fmt.Sprintf("credential secret field %q contains prohibited header control characters; supply a valid value", e.Field)
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

// RequirePersistentValue rejects values with no credential material: empty, a
// single LF, or a single CRLF. It is intentionally byte-exact: whitespace is
// not trimmed or rewritten before encrypted persistence.
func RequirePersistentValue(field, value string) error {
	if transportOnlyCredentialValue(value) {
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

// RequireAuthenticationValue rejects an empty, single-LF, or single-CRLF value
// before a shared auth helper can produce an empty request credential.
func RequireAuthenticationValue(field, value string) error {
	if transportOnlyCredentialValue(value) {
		return &EmptySecretError{Field: field}
	}
	return nil
}

func transportOnlyCredentialValue(value string) bool {
	switch value {
	case "", "\n", "\r\n":
		return true
	default:
		return false
	}
}

// ValidateHTTPHeaderValue rejects credential bytes that HTTP cannot transport
// safely in a header while preserving all transport-valid bytes unchanged.
func ValidateHTTPHeaderValue(field, value string) error {
	for i := 0; i < len(value); i++ {
		byteValue := value[i]
		if (byteValue < 0x20 && byteValue != '\t') || byteValue == 0x7f {
			return &InvalidSecretValueError{Field: field}
		}
	}
	return nil
}
