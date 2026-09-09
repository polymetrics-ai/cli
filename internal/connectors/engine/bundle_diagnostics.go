package engine

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// BundleDiagnosticError exposes only producer-owned reason text and selected
// execution identity. Cause retains the entire original graph for errors.Is/As
// and must never be serialized into public diagnostics. Field uses fixed JSON
// Pointer components; <member:N> denotes an untrusted member's zero-based
// ordinal among sorted keys, not a literal JSON Pointer or the member's name.
// Database definitions retain their schema-owned $.field[index] coordinates;
// @byte:N identifies the end offset of an unknown member token without its text.
type BundleDiagnosticError struct {
	Connector  string `json:"connector"`
	Generation string `json:"generation"`
	Digest     string `json:"digest"`
	File       string `json:"file"`
	Field      string `json:"field"`
	ReasonCode string `json:"reason_code"`
	Reason     string `json:"reason"`
	Cause      error  `json:"-"`
}

func (e *BundleDiagnosticError) Error() string {
	reason := e.Reason
	if reason == "" {
		reason = "invalid bundle declaration"
	}
	return fmt.Sprintf("load bundle %q generation=%q file=%q field=%q: %s", e.Connector, e.Generation, e.File, e.Field, reason)
}
func (e *BundleDiagnosticError) Unwrap() error { return e.Cause }
func (e *BundleDiagnosticError) BundleLocation() (string, string, string, string) {
	return e.Connector, e.Generation, e.File, e.Field
}

// SafeBundleError selects the outer diagnostic without exposing wrapping or
// joined sibling text. A fresh copy retains the whole incoming graph.
func SafeBundleError(err error) error {
	var d *BundleDiagnosticError
	if !errors.As(err, &d) {
		return err
	}
	bound := *d
	bound.Cause = err
	return &bound
}

type bundleFileError struct {
	file  string
	cause error
}

func (e *bundleFileError) Error() string { return fmt.Sprintf("read %s: %v", e.file, e.cause) }
func (e *bundleFileError) Unwrap() error { return e.cause }

// Private located errors preserve the old direct-helper text. Only the outer
// BundleDiagnosticError publishes safe metadata; producers never parse prose.
type schemaValidationError struct {
	path, code, reason string
	cause              error
}

func (e *schemaValidationError) Error() string { return e.cause.Error() }
func (e *schemaValidationError) Unwrap() error { return e.cause }
func diagnosticAt(path, code, reason string, cause error) error {
	return &schemaValidationError{path: path, code: code, reason: reason, cause: cause}
}
func diagnosticWithin(prefix string, cause error) error {
	var d *schemaValidationError
	if !errors.As(cause, &d) {
		return cause
	}
	return diagnosticAt(prefix+d.path, d.code, d.reason, cause)
}
func diagnosticMember(index int) string { return fmt.Sprintf("/<member:%d>", index) }
func diagnosticProperty(key string) string {
	return "/" + strings.NewReplacer("~", "~0", "/", "~1").Replace(key)
}
func sortedDiagnosticKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Called only while constructing immutable embedded meta-schemas. Authored
// properties and pattern-matched instance members never acquire this trust.
func (n *schemaNode) trustDiagnosticProperties() {
	if n == nil {
		return
	}
	n.trustedProperties = true
	for _, child := range n.properties {
		child.trustDiagnosticProperties()
	}
	for _, pattern := range n.patternProperties {
		pattern.schema.trustDiagnosticProperties()
	}
	n.items.trustDiagnosticProperties()
	for _, child := range n.prefixItems {
		child.trustDiagnosticProperties()
	}
	for _, child := range n.oneOf {
		child.trustDiagnosticProperties()
	}
}

func (e *schemaValidationError) BundleDiagnostic() (string, string, string) {
	return e.path, e.code, e.reason
}
