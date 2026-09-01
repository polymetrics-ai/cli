// Package failures defines the shared, serializable classification for
// connector configuration, system, and transient failures.
package failures

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxMessageBytes   = 1024
	maxReferenceBytes = 256
)

var (
	codePattern           = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	referenceValuePattern = regexp.MustCompile(`^[A-Za-z0-9._:/-]+$`)
)

// Domain identifies the recovery class of a failure. Only transient failures
// may be retried automatically.
type Domain string

const (
	DomainConfiguration Domain = "configuration"
	DomainSystem        Domain = "system"
	DomainTransient     Domain = "transient"
)

// DispatchKind identifies a non-executable command-dispatch terminal. It is
// valid only for system failures because it describes a local runtime or
// declaration defect rather than caller configuration or a retryable event.
type DispatchKind string

const (
	DispatchKindDirectStub                   DispatchKind = "direct_stub"
	DispatchKindHelperDelegatedRefusal       DispatchKind = "helper_delegated_refusal"
	DispatchKindWrappedTypedUnsupported      DispatchKind = "wrapped_typed_unsupported"
	DispatchKindDeclaredButUnroutableCommand DispatchKind = "declared_but_unroutable_command"
	DispatchKindUnresolvedDynamicTarget      DispatchKind = "unresolved_dynamic_target"
)

// ReferenceKind bounds the context serialized with a classification. It
// deliberately names only stable identifiers, never request values, provider
// response content, or credentials.
type ReferenceKind string

const (
	ReferenceKindSource     ReferenceKind = "source"
	ReferenceKindConnector  ReferenceKind = "connector"
	ReferenceKindCommand    ReferenceKind = "command"
	ReferenceKindOperation  ReferenceKind = "operation"
	ReferenceKindCapability ReferenceKind = "capability"
)

// Reference is a safe identifier used to locate the source of a failure.
// Value must be a bounded single-line identifier, not a request or response
// value.
type Reference struct {
	Kind  ReferenceKind `json:"kind"`
	Value string        `json:"value"`
}

// Input is the serializable portion of a Classification. Cause is supplied
// separately to New so it cannot accidentally enter a report payload.
type Input struct {
	Domain       Domain
	Code         string
	Message      string
	FieldPath    string
	DispatchKind DispatchKind
	References   []Reference
}

// Classification is a typed connector failure. Its message is safe to show to
// a user; its optional cause remains available only through Cause or Unwrap.
type Classification struct {
	domain       Domain
	code         string
	message      string
	fieldPath    string
	dispatchKind DispatchKind
	references   []Reference
	cause        error
}

// New constructs a validated Classification. It rejects invalid wire values
// before they can enter a diagnostic or configuration report.
func New(input Input, cause error) (*Classification, error) {
	normalized, err := normalizeInput(input)
	if err != nil {
		return nil, err
	}
	return &Classification{
		domain:       normalized.Domain,
		code:         normalized.Code,
		message:      normalized.Message,
		fieldPath:    normalized.FieldPath,
		dispatchKind: normalized.DispatchKind,
		references:   cloneReferences(normalized.References),
		cause:        normalizeCause(cause),
	}, nil
}

// Domain returns the closed recovery domain.
func (c *Classification) Domain() Domain {
	if c == nil {
		return ""
	}
	return c.domain
}

// Code returns the stable machine-readable reason code.
func (c *Classification) Code() string {
	if c == nil {
		return ""
	}
	return c.code
}

// Message returns the user-facing explanation without internal cause text.
func (c *Classification) Message() string {
	if c == nil {
		return ""
	}
	return c.message
}

// FieldPath returns the optional exact JSON Pointer for a field-level error.
func (c *Classification) FieldPath() string {
	if c == nil {
		return ""
	}
	return c.fieldPath
}

// DispatchKind returns the optional closed dispatch failure kind.
func (c *Classification) DispatchKind() DispatchKind {
	if c == nil {
		return ""
	}
	return c.dispatchKind
}

// References returns a defensive copy of the stable failure references.
func (c *Classification) References() []Reference {
	if c == nil {
		return nil
	}
	return cloneReferences(c.references)
}

// Cause returns the internal diagnostic cause. It is intentionally omitted
// from JSON and Error so callers never need to parse or expose it.
func (c *Classification) Cause() error {
	if c == nil {
		return nil
	}
	return c.cause
}

// Retryable reports whether automatic retry is permitted. Configuration and
// system failures are never retryable; only transient failures are.
func (c *Classification) Retryable() bool {
	return c != nil && c.domain == DomainTransient
}

// Error returns a user-facing field-scoped message without cause text.
func (c *Classification) Error() string {
	if c == nil {
		return "failure classification is nil"
	}
	if c.fieldPath == "" {
		return c.message
	}
	return c.fieldPath + ": " + c.message
}

// Unwrap exposes the internal cause to Go error inspection without adding it
// to serialized or user-facing output.
func (c *Classification) Unwrap() error {
	return c.Cause()
}

// Validate checks a Classification received from a caller or decoded from
// JSON. Constructors always return a valid value.
func (c *Classification) Validate() error {
	if c == nil {
		return fmt.Errorf("failure classification is nil")
	}
	_, err := normalizeInput(Input{
		Domain:       c.domain,
		Code:         c.code,
		Message:      c.message,
		FieldPath:    c.fieldPath,
		DispatchKind: c.dispatchKind,
		References:   c.references,
	})
	return err
}

// MarshalJSON emits only stable, safe classification data. The internal cause
// is deliberately unavailable to JSON encoders.
func (c Classification) MarshalJSON() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(classificationJSON{
		Domain:       c.domain,
		Code:         c.code,
		Message:      c.message,
		FieldPath:    jsonPointer(c.fieldPath),
		DispatchKind: c.dispatchKind,
		References:   c.references,
	})
}

// UnmarshalJSON accepts only one complete, valid classification object and
// always leaves the in-memory cause empty.
func (c *Classification) UnmarshalJSON(raw []byte) error {
	if c == nil {
		return fmt.Errorf("unmarshal failure classification into nil receiver")
	}
	if !utf8.Valid(raw) {
		return fmt.Errorf("decode failure classification: JSON is not valid UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var wire classificationJSON
	if err := decoder.Decode(&wire); err != nil {
		return fmt.Errorf("decode failure classification: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("decode failure classification: multiple JSON values")
		}
		return fmt.Errorf("decode failure classification: %w", err)
	}
	classification, err := New(Input{
		Domain:       wire.Domain,
		Code:         wire.Code,
		Message:      wire.Message,
		FieldPath:    string(wire.FieldPath),
		DispatchKind: wire.DispatchKind,
		References:   wire.References,
	}, nil)
	if err != nil {
		return err
	}
	*c = *classification
	return nil
}

type classificationJSON struct {
	Domain       Domain       `json:"domain"`
	Code         string       `json:"code"`
	Message      string       `json:"message"`
	FieldPath    jsonPointer  `json:"field_path,omitempty"`
	DispatchKind DispatchKind `json:"dispatch_kind,omitempty"`
	References   []Reference  `json:"references,omitempty"`
}

type jsonPointer string

func (pointer *jsonPointer) UnmarshalJSON(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	if bytes.Equal(raw, []byte("null")) {
		*pointer = ""
		return nil
	}
	if err := validateJSONPointerSurrogateEscapes(raw); err != nil {
		return err
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	*pointer = jsonPointer(value)
	return nil
}

func normalizeInput(input Input) (Input, error) {
	if !validDomain(input.Domain) {
		return Input{}, fmt.Errorf("failure domain %q is not supported", input.Domain)
	}
	if !codePattern.MatchString(input.Code) {
		return Input{}, fmt.Errorf("failure code must be lowercase snake_case")
	}
	input.Message = strings.TrimSpace(input.Message)
	if err := validateSafeText("failure message", input.Message, maxMessageBytes); err != nil {
		return Input{}, err
	}
	if err := validateJSONPointer(input.FieldPath); err != nil {
		return Input{}, err
	}
	if input.DispatchKind != "" {
		if !validDispatchKind(input.DispatchKind) {
			return Input{}, fmt.Errorf("dispatch failure kind %q is not supported", input.DispatchKind)
		}
		if input.Domain != DomainSystem {
			return Input{}, fmt.Errorf("dispatch failure kind requires system domain")
		}
	}
	for _, reference := range input.References {
		if !validReferenceKind(reference.Kind) {
			return Input{}, fmt.Errorf("failure reference kind %q is not supported", reference.Kind)
		}
		if err := validateSafeText("failure reference", reference.Value, maxReferenceBytes); err != nil {
			return Input{}, err
		}
		if !referenceValuePattern.MatchString(reference.Value) {
			return Input{}, fmt.Errorf("failure reference must be an identifier, not a request or response value")
		}
	}
	input.References = cloneReferences(input.References)
	return input, nil
}

func validDomain(domain Domain) bool {
	switch domain {
	case DomainConfiguration, DomainSystem, DomainTransient:
		return true
	default:
		return false
	}
}

func validDispatchKind(kind DispatchKind) bool {
	switch kind {
	case DispatchKindDirectStub,
		DispatchKindHelperDelegatedRefusal,
		DispatchKindWrappedTypedUnsupported,
		DispatchKindDeclaredButUnroutableCommand,
		DispatchKindUnresolvedDynamicTarget:
		return true
	default:
		return false
	}
}

func validReferenceKind(kind ReferenceKind) bool {
	switch kind {
	case ReferenceKindSource, ReferenceKindConnector, ReferenceKindCommand, ReferenceKindOperation, ReferenceKindCapability:
		return true
	default:
		return false
	}
}

func validateSafeText(label, value string, maxBytes int) error {
	if value == "" {
		return fmt.Errorf("%s is required", label)
	}
	if len(value) > maxBytes {
		return fmt.Errorf("%s exceeds %d bytes", label, maxBytes)
	}
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s is not valid UTF-8", label)
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return fmt.Errorf("%s contains a control character", label)
		}
	}
	return nil
}

func validateJSONPointer(pointer string) error {
	if !utf8.ValidString(pointer) {
		return fmt.Errorf("failure field path is not valid UTF-8")
	}
	if pointer == "" {
		return nil
	}
	if !strings.HasPrefix(pointer, "/") {
		return fmt.Errorf("failure field path must be a JSON Pointer")
	}
	for i := 0; i < len(pointer); i++ {
		if pointer[i] != '~' {
			continue
		}
		if i+1 >= len(pointer) || (pointer[i+1] != '0' && pointer[i+1] != '1') {
			return fmt.Errorf("failure field path has an invalid JSON Pointer escape")
		}
		i++
	}
	return nil
}

func validateJSONPointerSurrogateEscapes(raw []byte) error {
	if len(raw) < 2 || raw[0] != '"' || raw[len(raw)-1] != '"' {
		return nil
	}
	for i := 1; i < len(raw)-1; {
		if raw[i] != '\\' {
			i++
			continue
		}
		codeUnit, ok := jsonUTF16Escape(raw, i)
		if !ok {
			i += 2
			continue
		}
		if codeUnit >= 0xD800 && codeUnit <= 0xDBFF {
			following, ok := jsonUTF16Escape(raw, i+6)
			if !ok || following < 0xDC00 || following > 0xDFFF {
				return fmt.Errorf("failure field path has an unpaired UTF-16 surrogate escape")
			}
			i += 12
			continue
		}
		if codeUnit >= 0xDC00 && codeUnit <= 0xDFFF {
			return fmt.Errorf("failure field path has an unpaired UTF-16 surrogate escape")
		}
		i += 6
	}
	return nil
}

func jsonUTF16Escape(raw []byte, start int) (uint16, bool) {
	if start+6 > len(raw) || raw[start] != '\\' || raw[start+1] != 'u' {
		return 0, false
	}
	var value uint16
	for _, character := range raw[start+2 : start+6] {
		value <<= 4
		switch {
		case character >= '0' && character <= '9':
			value += uint16(character - '0')
		case character >= 'a' && character <= 'f':
			value += uint16(character-'a') + 10
		case character >= 'A' && character <= 'F':
			value += uint16(character-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}

func cloneReferences(references []Reference) []Reference {
	return append([]Reference(nil), references...)
}

func normalizeCause(cause error) error {
	if cause == nil {
		return nil
	}
	value := reflect.ValueOf(cause)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if value.IsNil() {
			return nil
		}
	}
	return cause
}
