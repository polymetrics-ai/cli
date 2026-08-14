// Package database defines the typed, non-executing foundation shared by
// native database connectors. It intentionally contains no SQL renderer,
// connection opener, write session, or changefeed implementation.
package database

import (
	"errors"
	"fmt"
	"strings"
)

// LogicalKind is the closed scalar/collection vocabulary used by database
// catalogs and type planning. It deliberately has no catch-all string type.
type LogicalKind string

const (
	LogicalSignedInteger   LogicalKind = "signed_integer"
	LogicalUnsignedInteger LogicalKind = "unsigned_integer"
	LogicalDecimal         LogicalKind = "decimal"
	LogicalFloat           LogicalKind = "float"
	LogicalBoolean         LogicalKind = "boolean"
	LogicalString          LogicalKind = "string"
	LogicalBinary          LogicalKind = "binary"
	LogicalDate            LogicalKind = "date"
	LogicalTime            LogicalKind = "time"
	LogicalTimestamp       LogicalKind = "timestamp"
	LogicalUUID            LogicalKind = "uuid"
	LogicalJSON            LogicalKind = "json"
	LogicalArray           LogicalKind = "array"
	LogicalOpaqueNative    LogicalKind = "opaque_native"
)

const (
	maxLogicalDecimalPrecision = 1000
	maxLogicalLength           = 1 << 30
	maxLogicalTemporalScale    = 9
	maxLogicalArrayDepth       = 8
)

// LogicalType is an immutable-by-construction value. Constructors validate all
// parameters, and accessors return copies for its nested values. Nullability
// belongs to Column rather than to this scalar type.
type LogicalType struct {
	kind          LogicalKind
	bits          uint8
	precision     uint16
	scale         uint16
	maxLength     uint32
	collation     string
	withTimezone  bool
	element       *LogicalType
	opaqueEngine  string
	opaqueName    string
	opaqueOptions []string
}

// Kind returns the closed logical kind.
func (t LogicalType) Kind() LogicalKind { return t.kind }

// BitWidth returns the integer or floating-point width, when applicable.
func (t LogicalType) BitWidth() uint8 { return t.bits }

// Precision returns decimal or temporal precision, when applicable.
func (t LogicalType) Precision() uint16 { return t.precision }

// Scale returns decimal scale, when applicable.
func (t LogicalType) Scale() uint16 { return t.scale }

// MaxLength returns the declared string/binary bound. Zero means the logical
// type itself is unbounded; it never means an unbounded runtime resource.
func (t LogicalType) MaxLength() uint32 { return t.maxLength }

// Collation returns a string collation annotation, when present.
func (t LogicalType) Collation() string { return t.collation }

// WithTimezone reports the explicit time/timestamp timezone semantics.
func (t LogicalType) WithTimezone() bool { return t.withTimezone }

// Element returns a defensive copy of an array element type, or nil.
func (t LogicalType) Element() *LogicalType {
	if t.element == nil {
		return nil
	}
	clone := t.element.clone()
	return &clone
}

// OpaqueNativeDetails returns defensive copies of the native type identity.
// Opaque types may exist in a discovered catalog, but can never compile into a
// lossless target type plan until a closed mapping is added.
func (t LogicalType) OpaqueNativeDetails() (engine, name string, options []string) {
	return t.opaqueEngine, t.opaqueName, append([]string(nil), t.opaqueOptions...)
}

// NewSignedInteger creates a signed integer logical type with the stated
// exact bit width.
func NewSignedInteger(bits uint8) (LogicalType, error) {
	return newLogicalType(LogicalType{kind: LogicalSignedInteger, bits: bits})
}

// NewUnsignedInteger creates an unsigned integer logical type with the stated
// exact bit width.
func NewUnsignedInteger(bits uint8) (LogicalType, error) {
	return newLogicalType(LogicalType{kind: LogicalUnsignedInteger, bits: bits})
}

// NewDecimal creates a fixed-precision decimal logical type.
func NewDecimal(precision, scale uint16) (LogicalType, error) {
	return newLogicalType(LogicalType{kind: LogicalDecimal, precision: precision, scale: scale})
}

// NewFloat creates a floating-point logical type with the stated exact width.
func NewFloat(bits uint8) (LogicalType, error) {
	return newLogicalType(LogicalType{kind: LogicalFloat, bits: bits})
}

// NewBoolean creates the boolean logical type.
func NewBoolean() LogicalType { return LogicalType{kind: LogicalBoolean} }

// NewString creates UTF-8 text with an optional maximum length and collation.
func NewString(maxLength uint32, collation string) (LogicalType, error) {
	return newLogicalType(LogicalType{kind: LogicalString, maxLength: maxLength, collation: collation})
}

// NewBinary creates binary data with an optional maximum length.
func NewBinary(maxLength uint32) (LogicalType, error) {
	return newLogicalType(LogicalType{kind: LogicalBinary, maxLength: maxLength})
}

// NewDate creates the calendar-date logical type.
func NewDate() LogicalType { return LogicalType{kind: LogicalDate} }

// NewTime creates a time-of-day logical type with explicit timezone semantics.
func NewTime(precision uint16, withTimezone bool) (LogicalType, error) {
	return newLogicalType(LogicalType{kind: LogicalTime, precision: precision, withTimezone: withTimezone})
}

// NewTimestamp creates a timestamp logical type with explicit timezone
// semantics. A timezone conversion is never implicit in compatibility checks.
func NewTimestamp(precision uint16, withTimezone bool) (LogicalType, error) {
	return newLogicalType(LogicalType{kind: LogicalTimestamp, precision: precision, withTimezone: withTimezone})
}

// NewUUID creates the UUID logical type.
func NewUUID() LogicalType { return LogicalType{kind: LogicalUUID} }

// NewJSON creates the JSON logical type.
func NewJSON() LogicalType { return LogicalType{kind: LogicalJSON} }

// NewArray creates a collection whose element has a closed logical type.
func NewArray(element LogicalType) (LogicalType, error) {
	clone := element.clone()
	return newLogicalType(LogicalType{kind: LogicalArray, element: &clone})
}

// NewOpaqueNative records a catalog-discovered native type that has no trusted
// closed mapping yet. It cannot be used as a string fallback or target plan.
func NewOpaqueNative(engine, name string, options []string) (LogicalType, error) {
	return newLogicalType(LogicalType{
		kind:          LogicalOpaqueNative,
		opaqueEngine:  engine,
		opaqueName:    name,
		opaqueOptions: append([]string(nil), options...),
	})
}

func newLogicalType(t LogicalType) (LogicalType, error) {
	if err := t.validate(0); err != nil {
		return LogicalType{}, err
	}
	return t.clone(), nil
}

func (t LogicalType) clone() LogicalType {
	clone := t
	clone.opaqueOptions = append([]string(nil), t.opaqueOptions...)
	if t.element != nil {
		element := t.element.clone()
		clone.element = &element
	}
	return clone
}

func (t LogicalType) containsOpaqueNative() bool {
	return t.kind == LogicalOpaqueNative || (t.element != nil && t.element.containsOpaqueNative())
}

// Equal reports structural logical-type equality, including all parameters.
func (t LogicalType) Equal(other LogicalType) bool {
	if t.kind != other.kind || t.bits != other.bits || t.precision != other.precision ||
		t.scale != other.scale || t.maxLength != other.maxLength ||
		t.collation != other.collation || t.withTimezone != other.withTimezone ||
		t.opaqueEngine != other.opaqueEngine || t.opaqueName != other.opaqueName ||
		len(t.opaqueOptions) != len(other.opaqueOptions) {
		return false
	}
	for i := range t.opaqueOptions {
		if t.opaqueOptions[i] != other.opaqueOptions[i] {
			return false
		}
	}
	if t.element == nil || other.element == nil {
		return t.element == nil && other.element == nil
	}
	return t.element.Equal(*other.element)
}

func (t LogicalType) validate(depth int) error {
	if depth > maxLogicalArrayDepth {
		return errors.New("database logical type nesting exceeds the supported maximum")
	}
	switch t.kind {
	case LogicalSignedInteger, LogicalUnsignedInteger:
		if !validIntegerBits(t.bits) || !t.hasNoParametersExcept("bits") {
			return errors.New("database logical integer type has invalid parameters")
		}
	case LogicalDecimal:
		if t.precision == 0 || t.precision > maxLogicalDecimalPrecision || t.scale > t.precision || !t.hasNoParametersExcept("decimal") {
			return errors.New("database logical decimal type has invalid precision or scale")
		}
	case LogicalFloat:
		if (t.bits != 32 && t.bits != 64) || !t.hasNoParametersExcept("bits") {
			return errors.New("database logical float type has invalid parameters")
		}
	case LogicalBoolean, LogicalDate, LogicalUUID, LogicalJSON:
		if !t.hasNoParametersExcept("none") {
			return errors.New("database logical scalar type has invalid parameters")
		}
	case LogicalString:
		if t.maxLength > maxLogicalLength || !validCollation(t.collation) || !t.hasNoParametersExcept("string") {
			return errors.New("database logical string type has invalid parameters")
		}
	case LogicalBinary:
		if t.maxLength > maxLogicalLength || !t.hasNoParametersExcept("binary") {
			return errors.New("database logical binary type has invalid parameters")
		}
	case LogicalTime, LogicalTimestamp:
		if t.precision > maxLogicalTemporalScale || !t.hasNoParametersExcept("temporal") {
			return errors.New("database logical temporal type has invalid parameters")
		}
	case LogicalArray:
		if t.element == nil || !t.hasNoParametersExcept("array") {
			return errors.New("database logical array type has invalid parameters")
		}
		if err := t.element.validate(depth + 1); err != nil {
			return fmt.Errorf("database logical array element: %w", err)
		}
	case LogicalOpaqueNative:
		if !validOpaqueName(t.opaqueEngine) || !validOpaqueName(t.opaqueName) || !t.hasNoParametersExcept("opaque") {
			return errors.New("database opaque native type has invalid parameters")
		}
		for _, option := range t.opaqueOptions {
			if !validOpaqueName(option) {
				return errors.New("database opaque native type has invalid modifiers")
			}
		}
	default:
		return errors.New("database logical type kind is not supported")
	}
	return nil
}

func (t LogicalType) hasNoParametersExcept(kind string) bool {
	switch kind {
	case "bits":
		return t.precision == 0 && t.scale == 0 && t.maxLength == 0 && t.collation == "" && !t.withTimezone && t.element == nil && t.opaqueEngine == "" && t.opaqueName == "" && len(t.opaqueOptions) == 0
	case "decimal":
		return t.bits == 0 && t.maxLength == 0 && t.collation == "" && !t.withTimezone && t.element == nil && t.opaqueEngine == "" && t.opaqueName == "" && len(t.opaqueOptions) == 0
	case "string":
		return t.bits == 0 && t.precision == 0 && t.scale == 0 && !t.withTimezone && t.element == nil && t.opaqueEngine == "" && t.opaqueName == "" && len(t.opaqueOptions) == 0
	case "binary":
		return t.bits == 0 && t.precision == 0 && t.scale == 0 && t.collation == "" && !t.withTimezone && t.element == nil && t.opaqueEngine == "" && t.opaqueName == "" && len(t.opaqueOptions) == 0
	case "temporal":
		return t.bits == 0 && t.scale == 0 && t.maxLength == 0 && t.collation == "" && t.element == nil && t.opaqueEngine == "" && t.opaqueName == "" && len(t.opaqueOptions) == 0
	case "array":
		return t.bits == 0 && t.precision == 0 && t.scale == 0 && t.maxLength == 0 && t.collation == "" && !t.withTimezone && t.opaqueEngine == "" && t.opaqueName == "" && len(t.opaqueOptions) == 0
	case "opaque":
		return t.bits == 0 && t.precision == 0 && t.scale == 0 && t.maxLength == 0 && t.collation == "" && !t.withTimezone && t.element == nil
	default:
		return t.bits == 0 && t.precision == 0 && t.scale == 0 && t.maxLength == 0 && t.collation == "" && !t.withTimezone && t.element == nil && t.opaqueEngine == "" && t.opaqueName == "" && len(t.opaqueOptions) == 0
	}
}

func validIntegerBits(bits uint8) bool {
	return bits == 8 || bits == 16 || bits == 32 || bits == 64
}

func validCollation(value string) bool {
	return value == "" || validOpaqueName(value)
}

func validOpaqueName(value string) bool {
	if value == "" || len(value) > 256 || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			continue
		}
		return false
	}
	return true
}

// Compatibility classifies whether the target logical type can represent the
// source without an implicit semantic transformation.
type Compatibility string

const (
	CompatibilityExact                     Compatibility = "exact"
	CompatibilityLossless                  Compatibility = "lossless"
	CompatibilityExplicitTransformRequired Compatibility = "explicit_transform_required"
	CompatibilityUnsupported               Compatibility = "unsupported"
)

// TypePlan is a defensive typed conversion decision. It is intentionally not a
// SQL cast or executable transformation.
type TypePlan struct {
	source         LogicalType
	target         LogicalType
	classification Compatibility
}

// Source returns a defensive copy of the planned source type.
func (p TypePlan) Source() LogicalType { return p.source.clone() }

// Target returns a defensive copy of the planned target type.
func (p TypePlan) Target() LogicalType { return p.target.clone() }

// Classification returns the closed compatibility result.
func (p TypePlan) Classification() Compatibility { return p.classification }

// ErrLosslessTypePlanRequired prevents a caller from treating a semantic cast
// request as an executable target mapping.
var ErrLosslessTypePlanRequired = errors.New("database type plan requires an exact or lossless mapping")

// TypeCompatibilityError describes a rejected plan without rendering source
// or target values, which can contain provider-defined native names.
type TypeCompatibilityError struct {
	Classification Compatibility
}

func (e *TypeCompatibilityError) Error() string {
	if e == nil || e.Classification == "" {
		return ErrLosslessTypePlanRequired.Error()
	}
	return "database type plan is " + string(e.Classification) + "; " + ErrLosslessTypePlanRequired.Error()
}

func (e *TypeCompatibilityError) Unwrap() error { return ErrLosslessTypePlanRequired }

// CompileTypePlan returns a plan only for an exact or proven-lossless mapping.
// Every other classification is deliberately rejected rather than mapped to a
// generic text type.
func CompileTypePlan(source, target LogicalType) (TypePlan, error) {
	classification, err := ClassifyTypeCompatibility(source, target)
	if err != nil {
		return TypePlan{}, err
	}
	if classification != CompatibilityExact && classification != CompatibilityLossless {
		return TypePlan{}, &TypeCompatibilityError{Classification: classification}
	}
	return TypePlan{source: source.clone(), target: target.clone(), classification: classification}, nil
}

// ClassifyTypeCompatibility reports the closed mapping classification without
// constructing an executable plan.
func ClassifyTypeCompatibility(source, target LogicalType) (Compatibility, error) {
	if err := source.validate(0); err != nil {
		return CompatibilityUnsupported, err
	}
	if err := target.validate(0); err != nil {
		return CompatibilityUnsupported, err
	}
	if source.containsOpaqueNative() || target.containsOpaqueNative() {
		return CompatibilityUnsupported, nil
	}
	if source.Equal(target) {
		return CompatibilityExact, nil
	}
	if source.kind != target.kind {
		return CompatibilityExplicitTransformRequired, nil
	}

	switch source.kind {
	case LogicalSignedInteger:
		if target.bits >= source.bits {
			return CompatibilityLossless, nil
		}
	case LogicalUnsignedInteger:
		if target.bits >= source.bits {
			return CompatibilityLossless, nil
		}
	case LogicalDecimal:
		if target.scale >= source.scale && target.precision-target.scale >= source.precision-source.scale {
			return CompatibilityLossless, nil
		}
	case LogicalFloat:
		if target.bits >= source.bits {
			return CompatibilityLossless, nil
		}
	case LogicalString:
		if source.collation == target.collation && isLengthSuperset(source.maxLength, target.maxLength) {
			return CompatibilityLossless, nil
		}
	case LogicalBinary:
		if isLengthSuperset(source.maxLength, target.maxLength) {
			return CompatibilityLossless, nil
		}
	case LogicalTime, LogicalTimestamp:
		if source.withTimezone == target.withTimezone && target.precision >= source.precision {
			return CompatibilityLossless, nil
		}
	case LogicalArray:
		classification, err := ClassifyTypeCompatibility(*source.element, *target.element)
		if err != nil {
			return CompatibilityUnsupported, err
		}
		if classification == CompatibilityExact || classification == CompatibilityLossless {
			return CompatibilityLossless, nil
		}
		return classification, nil
	}

	return CompatibilityExplicitTransformRequired, nil
}

func isLengthSuperset(source, target uint32) bool {
	if source == target {
		return true
	}
	return target == 0 || (source != 0 && target >= source)
}
