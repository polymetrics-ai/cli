package database

import (
	"encoding/json"
	"errors"
	"math"
	"time"
	"unicode/utf8"

	"polymetrics.ai/internal/connectors"
)

const (
	// MappingContractVersionV1 is the initial, closed mapping contract
	// revision. New revisions must be separate types so a sealed plan can never
	// silently change mapping semantics.
	MappingContractVersionV1 uint = 1
)

var (
	// ErrMappingContractInvalid refuses an incomplete, forged, or ambiguous
	// source-to-target mapping before it can be sealed into a write plan.
	ErrMappingContractInvalid = errors.New("database mapping contract is invalid")
	// ErrMappingValueInvalid refuses a record value that cannot be represented
	// by its declared exact or lossless mapping. It intentionally does not
	// render the value or field name.
	ErrMappingValueInvalid = errors.New("database mapping value is not representable")
)

// MappingColumnV1 binds one source record field to one target column and its
// lossless type plan. Every mapped target column has a source field: defaults,
// arbitrary expressions, and inferred columns are deliberately outside this
// shared contract.
type MappingColumnV1 struct {
	Source   string
	Target   string
	Type     TypePlan
	Nullable bool
}

func (c MappingColumnV1) clone() MappingColumnV1 {
	clone := c
	clone.Type = cloneTypePlan(c.Type)
	return clone
}

func (c MappingColumnV1) validate() error {
	if validateIdentifierComponent(c.Source) != nil || validateIdentifierComponent(c.Target) != nil || validateTypePlan(c.Type) != nil {
		return ErrMappingContractInvalid
	}
	return nil
}

// MappingContractV1 is an immutable, versioned target-column set and
// source-to-target mapping. It is driver-neutral: it supplies no DDL, SQL,
// relation name, connection, or provider-specific type.
type MappingContractV1 struct {
	version uint
	columns []MappingColumnV1
}

// NewMappingContractV1 validates and seals the complete ordered target-column
// set. Each type plan is independently recompiled to prove that callers cannot
// inject a narrowing or semantic conversion into a write plan.
func NewMappingContractV1(columns []MappingColumnV1) (MappingContractV1, error) {
	contract := MappingContractV1{
		version: MappingContractVersionV1,
		columns: make([]MappingColumnV1, len(columns)),
	}
	for index := range columns {
		contract.columns[index] = columns[index].clone()
	}
	if err := contract.validate(); err != nil {
		return MappingContractV1{}, ErrMappingContractInvalid
	}
	return contract, nil
}

func (c MappingContractV1) validate() error {
	if c.version != MappingContractVersionV1 || len(c.columns) == 0 {
		return ErrMappingContractInvalid
	}
	sources := make(map[string]struct{}, len(c.columns))
	targets := make(map[string]struct{}, len(c.columns))
	for _, column := range c.columns {
		if err := column.validate(); err != nil {
			return ErrMappingContractInvalid
		}
		if _, exists := sources[column.Source]; exists {
			return ErrMappingContractInvalid
		}
		if _, exists := targets[column.Target]; exists {
			return ErrMappingContractInvalid
		}
		sources[column.Source] = struct{}{}
		targets[column.Target] = struct{}{}
	}
	return nil
}

func (c MappingContractV1) clone() MappingContractV1 {
	clone := MappingContractV1{version: c.version, columns: make([]MappingColumnV1, len(c.columns))}
	for index := range c.columns {
		clone.columns[index] = c.columns[index].clone()
	}
	return clone
}

// Version returns the pinned mapping revision.
func (c MappingContractV1) Version() uint { return c.version }

// Columns returns an independent ordered projection of the complete target
// column set and source mappings.
func (c MappingContractV1) Columns() []MappingColumnV1 {
	columns := make([]MappingColumnV1, len(c.columns))
	for index := range c.columns {
		columns[index] = c.columns[index].clone()
	}
	return columns
}

// HasTarget reports whether one target column belongs to this sealed mapping.
func (c MappingContractV1) HasTarget(target string) bool {
	if c.validate() != nil {
		return false
	}
	for _, column := range c.columns {
		if column.Target == target {
			return true
		}
	}
	return false
}

func (c MappingContractV1) matches(other MappingContractV1) bool {
	if c.validate() != nil || other.validate() != nil || c.version != other.version || len(c.columns) != len(other.columns) {
		return false
	}
	for index := range c.columns {
		left, right := c.columns[index], other.columns[index]
		if left.Source != right.Source || left.Target != right.Target || left.Nullable != right.Nullable || !sameTypePlan(left.Type, right.Type) {
			return false
		}
	}
	return true
}

// MapRecord projects one source record into the exact ordered target column
// set. Every target value is represented using its declared target logical
// type; no undeclared source field can become a target column.
func (c MappingContractV1) MapRecord(source connectors.Record) (connectors.Record, error) {
	return c.projectRecord(source, false)
}

// UnmapRecord is the checked inverse projection for a value that was emitted
// by MapRecord. It is useful to prove lossless mappings and to reconcile
// mapped target evidence without inventing a second field vocabulary.
func (c MappingContractV1) UnmapRecord(target connectors.Record) (connectors.Record, error) {
	return c.projectRecord(target, true)
}

func (c MappingContractV1) projectRecord(record connectors.Record, reverse bool) (connectors.Record, error) {
	if c.validate() != nil || record == nil {
		return nil, ErrMappingContractInvalid
	}
	projected := make(connectors.Record, len(c.columns))
	for _, column := range c.columns {
		from, to := column.Source, column.Target
		sourceType, targetType := column.Type.Source(), column.Type.Target()
		if reverse {
			from, to = to, from
			sourceType, targetType = targetType, sourceType
		}
		value, found := record[from]
		if !found {
			return nil, ErrMappingValueInvalid
		}
		if value == nil {
			if !column.Nullable {
				return nil, ErrMappingValueInvalid
			}
			projected[to] = nil
			continue
		}
		mapped, err := mapLogicalValue(value, sourceType, targetType, !reverse)
		if err != nil {
			return nil, ErrMappingValueInvalid
		}
		projected[to] = mapped
	}
	return projected, nil
}

func cloneTypePlan(plan TypePlan) TypePlan {
	return TypePlan{
		source:         plan.source.clone(),
		target:         plan.target.clone(),
		classification: plan.classification,
	}
}

func validateTypePlan(plan TypePlan) error {
	if plan.source.validate(0) != nil || plan.target.validate(0) != nil {
		return ErrMappingContractInvalid
	}
	compiled, err := CompileTypePlan(plan.source, plan.target)
	if err != nil || !sameTypePlan(compiled, plan) {
		return ErrMappingContractInvalid
	}
	return nil
}

func sameTypePlan(left, right TypePlan) bool {
	return left.classification == right.classification && left.source.Equal(right.source) && left.target.Equal(right.target)
}

func mapLogicalValue(value any, source, target LogicalType, requireLossless bool) (any, error) {
	classification, err := ClassifyTypeCompatibility(source, target)
	if err != nil || (requireLossless && classification != CompatibilityExact && classification != CompatibilityLossless) {
		return nil, ErrMappingValueInvalid
	}
	switch source.Kind() {
	case LogicalSignedInteger:
		integer, err := signedMappingValue(value, source.BitWidth())
		if err != nil {
			return nil, err
		}
		return signedMappingTarget(integer, target.BitWidth())
	case LogicalUnsignedInteger:
		integer, err := unsignedMappingValue(value, source.BitWidth())
		if err != nil {
			return nil, err
		}
		return unsignedMappingTarget(integer, target.BitWidth())
	case LogicalFloat:
		return floatMappingValue(value, source.BitWidth(), target.BitWidth())
	case LogicalBoolean:
		if typed, ok := value.(bool); ok {
			return typed, nil
		}
	case LogicalString:
		if typed, ok := value.(string); ok && utf8.ValidString(typed) && mappingLengthWithin(typed, source.MaxLength()) && mappingLengthWithin(typed, target.MaxLength()) {
			return typed, nil
		}
	case LogicalBinary:
		if typed, ok := value.([]byte); ok && mappingBytesWithin(typed, source.MaxLength()) && mappingBytesWithin(typed, target.MaxLength()) {
			return append([]byte(nil), typed...), nil
		}
	case LogicalDecimal:
		if typed, ok := value.(string); ok {
			return typed, nil
		}
	case LogicalDate:
		if typed, ok := value.(string); ok {
			return typed, nil
		}
	case LogicalTime, LogicalTimestamp:
		switch typed := value.(type) {
		case string:
			return typed, nil
		case time.Time:
			return typed, nil
		}
	case LogicalUUID:
		switch typed := value.(type) {
		case string:
			return typed, nil
		case [16]byte:
			return typed, nil
		}
	case LogicalJSON:
		switch typed := value.(type) {
		case json.RawMessage:
			if json.Valid(typed) {
				return append(json.RawMessage(nil), typed...), nil
			}
		case []byte:
			if json.Valid(typed) {
				return append(json.RawMessage(nil), typed...), nil
			}
		case string:
			if json.Valid([]byte(typed)) {
				return typed, nil
			}
		}
	case LogicalArray:
		values, ok := value.([]any)
		if !ok {
			break
		}
		sourceElement, targetElement := source.Element(), target.Element()
		if sourceElement == nil || targetElement == nil {
			break
		}
		mapped := make([]any, len(values))
		for index := range values {
			if values[index] == nil {
				mapped[index] = nil
				continue
			}
			item, err := mapLogicalValue(values[index], *sourceElement, *targetElement, requireLossless)
			if err != nil {
				return nil, err
			}
			mapped[index] = item
		}
		return mapped, nil
	}
	return nil, ErrMappingValueInvalid
}

func signedMappingValue(value any, bits uint8) (int64, error) {
	switch bits {
	case 8:
		if typed, ok := value.(int8); ok {
			return int64(typed), nil
		}
	case 16:
		if typed, ok := value.(int16); ok {
			return int64(typed), nil
		}
	case 32:
		if typed, ok := value.(int32); ok {
			return int64(typed), nil
		}
	case 64:
		if typed, ok := value.(int64); ok {
			return typed, nil
		}
	}
	return 0, ErrMappingValueInvalid
}

func signedMappingTarget(value int64, bits uint8) (any, error) {
	switch bits {
	case 8:
		if value >= math.MinInt8 && value <= math.MaxInt8 {
			return int8(value), nil
		}
	case 16:
		if value >= math.MinInt16 && value <= math.MaxInt16 {
			return int16(value), nil
		}
	case 32:
		if value >= math.MinInt32 && value <= math.MaxInt32 {
			return int32(value), nil
		}
	case 64:
		return value, nil
	}
	return nil, ErrMappingValueInvalid
}

func unsignedMappingValue(value any, bits uint8) (uint64, error) {
	switch bits {
	case 8:
		if typed, ok := value.(uint8); ok {
			return uint64(typed), nil
		}
	case 16:
		if typed, ok := value.(uint16); ok {
			return uint64(typed), nil
		}
	case 32:
		if typed, ok := value.(uint32); ok {
			return uint64(typed), nil
		}
	case 64:
		if typed, ok := value.(uint64); ok {
			return typed, nil
		}
	}
	return 0, ErrMappingValueInvalid
}

func unsignedMappingTarget(value uint64, bits uint8) (any, error) {
	switch bits {
	case 8:
		if value <= math.MaxUint8 {
			return uint8(value), nil
		}
	case 16:
		if value <= math.MaxUint16 {
			return uint16(value), nil
		}
	case 32:
		if value <= math.MaxUint32 {
			return uint32(value), nil
		}
	case 64:
		return value, nil
	}
	return nil, ErrMappingValueInvalid
}

func floatMappingValue(value any, sourceBits, targetBits uint8) (any, error) {
	switch sourceBits {
	case 32:
		typed, ok := value.(float32)
		if !ok {
			return nil, ErrMappingValueInvalid
		}
		if targetBits == 32 {
			return typed, nil
		}
		return float64(typed), nil
	case 64:
		typed, ok := value.(float64)
		if !ok || targetBits != 64 {
			return nil, ErrMappingValueInvalid
		}
		return typed, nil
	default:
		return nil, ErrMappingValueInvalid
	}
}

func mappingLengthWithin(value string, maximum uint32) bool {
	return maximum == 0 || uint64(len(value)) <= uint64(maximum)
}

func mappingBytesWithin(value []byte, maximum uint32) bool {
	return maximum == 0 || uint64(len(value)) <= uint64(maximum)
}
