package database

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const TransformPlanVersionV1 uint = 1

var ErrTransformPlanInvalid = errors.New("database transform plan is invalid")

// TransformPlanV1 is the closed, connector-neutral description of a bounded
// typed projection. It is intentionally an immutable value: callers retain
// normalized JSON and a hash, never a user-authored query or a mutable AST.
// Native connectors may compile this vocabulary into their own vector engine,
// but no native syntax crosses this contract.
type TransformPlanV1 struct {
	normalized []byte
	hash       string
	outputs    []TransformOutputColumnV1
}

// TransformOutputColumnV1 is one ordered typed result column. The expression
// is canonical JSON from the closed TransformPlanV1 language, not SQL.
type TransformOutputColumnV1 struct {
	Target     string
	Type       string
	Rounding   string
	Source     string
	Expression []byte
}

func (c TransformOutputColumnV1) clone() TransformOutputColumnV1 {
	clone := c
	clone.Expression = append([]byte(nil), c.Expression...)
	return clone
}

// ParseTransformPlanV1 validates raw JSON, rejects unknown operations before
// any caller can open an endpoint, and returns its one canonical encoding.
func ParseTransformPlanV1(raw []byte) (TransformPlanV1, error) {
	var document transformPlanDocument
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil || decoder.More() {
		return TransformPlanV1{}, ErrTransformPlanInvalid
	}
	if document.Version != TransformPlanVersionV1 || len(document.Select) == 0 {
		return TransformPlanV1{}, ErrTransformPlanInvalid
	}

	outputs := make([]TransformOutputColumnV1, len(document.Select))
	targets := make(map[string]struct{}, len(document.Select))
	for index, projection := range document.Select {
		output, normalized, err := normalizeTransformProjection(projection)
		if err != nil {
			return TransformPlanV1{}, ErrTransformPlanInvalid
		}
		if _, exists := targets[output.Target]; exists {
			return TransformPlanV1{}, ErrTransformPlanInvalid
		}
		targets[output.Target] = struct{}{}
		outputs[index] = output
		document.Select[index] = normalized
	}
	if len(document.Where) != 0 {
		normalized, err := normalizeTransformExpression(document.Where)
		if err != nil || normalized.Operation != "not_equal" {
			return TransformPlanV1{}, ErrTransformPlanInvalid
		}
		document.Where = normalized.JSON
	}
	normalized, err := json.Marshal(document)
	if err != nil {
		return TransformPlanV1{}, ErrTransformPlanInvalid
	}
	digest := sha256.Sum256(normalized)
	return TransformPlanV1{normalized: normalized, hash: hex.EncodeToString(digest[:]), outputs: outputs}, nil
}

// NormalizedJSON returns a fresh canonical plan encoding suitable for durable
// connection state. It never returns the caller's original formatting.
func (p TransformPlanV1) NormalizedJSON() []byte {
	return append([]byte(nil), p.normalized...)
}

// Hash is the stable SHA-256 identity bound into plan, preview, approval, and
// segment receipts. A zero plan has no hash and is not a transform plan.
func (p TransformPlanV1) Hash() string { return p.hash }

// Outputs returns an ordered, defensive projection of typed result columns.
func (p TransformPlanV1) Outputs() []TransformOutputColumnV1 {
	outputs := make([]TransformOutputColumnV1, len(p.outputs))
	for index := range p.outputs {
		outputs[index] = p.outputs[index].clone()
	}
	return outputs
}

// SourceFields returns the sorted, unique source columns that a typed
// extractor must provide to evaluate this closed plan. It exposes field names
// only, never an AST or executable syntax, so a native range reader can
// project narrow Arrow batches without learning a destination dialect.
func (p TransformPlanV1) SourceFields() ([]string, error) {
	if !p.valid() {
		return nil, ErrTransformPlanInvalid
	}
	fields := make(map[string]struct{})
	for _, output := range p.outputs {
		if output.Source != "" {
			fields[output.Source] = struct{}{}
			continue
		}
		if err := collectTransformExpressionFields(output.Expression, fields); err != nil {
			return nil, ErrTransformPlanInvalid
		}
	}
	var document transformPlanDocument
	if err := json.Unmarshal(p.normalized, &document); err != nil {
		return nil, ErrTransformPlanInvalid
	}
	if len(document.Where) != 0 {
		if err := collectTransformExpressionFields(document.Where, fields); err != nil {
			return nil, ErrTransformPlanInvalid
		}
	}
	result := make([]string, 0, len(fields))
	for field := range fields {
		result = append(result, field)
	}
	sort.Strings(result)
	return result, nil
}

func collectTransformExpressionFields(raw []byte, fields map[string]struct{}) error {
	expression, err := normalizeTransformExpression(raw)
	if err != nil {
		return err
	}
	switch expression.Operation {
	case "upper", "date":
		field, err := transformExpressionField(expression.JSON, expression.Operation)
		if err != nil {
			return err
		}
		fields[field] = struct{}{}
		return nil
	case "cast":
		var node struct {
			Cast json.RawMessage `json:"cast"`
		}
		if err := json.Unmarshal(expression.JSON, &node); err != nil {
			return err
		}
		return collectTransformExpressionFields(node.Cast, fields)
	case "multiply", "mod", "not_equal":
		var node map[string][]json.RawMessage
		if err := json.Unmarshal(expression.JSON, &node); err != nil || len(node[expression.Operation]) != 2 {
			return ErrTransformPlanInvalid
		}
		for _, operand := range node[expression.Operation] {
			var field string
			if err := json.Unmarshal(operand, &field); err == nil {
				fields[field] = struct{}{}
				continue
			}
			trimmed := bytes.TrimSpace(operand)
			if len(trimmed) > 0 && trimmed[0] == '{' {
				if err := collectTransformExpressionFields(trimmed, fields); err != nil {
					return err
				}
			}
		}
		return nil
	default:
		return ErrTransformPlanInvalid
	}
}

// OutputMapping returns an identity mapping for the already-transformed,
// typed segment columns. This deliberately maps result names to themselves:
// native extractors and vector engines own expression evaluation, while the
// managed-target provisioning contract owns only its output schema.
func (p TransformPlanV1) OutputMapping() (MappingContractV1, error) {
	if !p.valid() {
		return MappingContractV1{}, ErrTransformPlanInvalid
	}
	columns := make([]MappingColumnV1, len(p.outputs))
	for index, output := range p.outputs {
		logical, err := transformLogicalType(output.Type)
		if err != nil {
			return MappingContractV1{}, ErrTransformPlanInvalid
		}
		typePlan, err := CompileTypePlan(logical, logical)
		if err != nil {
			return MappingContractV1{}, ErrTransformPlanInvalid
		}
		columns[index] = MappingColumnV1{Source: output.Target, Target: output.Target, Type: typePlan, Nullable: true}
	}
	return NewMappingContractV1(columns)
}

// OutputRelation derives the typed target schema owned by a transformed
// stream. It retains only the sealed relation identity; source keys are not
// copied because a projected full-overwrite target has no implicit upsert key.
func (p TransformPlanV1) OutputRelation(source Relation) (Relation, error) {
	if !p.valid() || p.ValidateAgainstRelation(source) != nil {
		return Relation{}, ErrTransformPlanInvalid
	}
	relation := Relation{Ref: source.Ref, NativeIdentity: source.NativeIdentity, Columns: make([]Column, len(p.outputs))}
	for index, output := range p.outputs {
		logical, err := transformLogicalType(output.Type)
		if err != nil {
			return Relation{}, ErrTransformPlanInvalid
		}
		relation.Columns[index] = Column{Ref: ColumnRef{Relation: relation.Ref, Name: output.Target}, Type: logical, Nullable: true, Ordinal: index + 1}
	}
	return relation, nil
}

// ValidateAgainstRelation proves every referenced source field and operation
// is compatible with a typed source catalog. Callers use it at connection
// creation, before the connection is persisted or a transport can start.
func (p TransformPlanV1) ValidateAgainstRelation(relation Relation) error {
	if !p.valid() || relation.validate(relation.Ref.Schema.Catalog) != nil {
		return ErrTransformPlanInvalid
	}
	columns := make(map[string]Column, len(relation.Columns))
	for _, column := range relation.Columns {
		columns[column.Ref.Name] = column
	}
	for _, output := range p.outputs {
		if output.Source != "" {
			column, found := columns[output.Source]
			if !found || !transformDirectTypeAllowed(column.Type, output.Type) {
				return ErrTransformPlanInvalid
			}
			continue
		}
		if err := validateTransformExpressionAgainstColumns(output.Expression, columns, output.Type); err != nil {
			return ErrTransformPlanInvalid
		}
	}
	var document transformPlanDocument
	if err := json.Unmarshal(p.normalized, &document); err != nil {
		return ErrTransformPlanInvalid
	}
	if len(document.Where) != 0 {
		if err := validateTransformExpressionAgainstColumns(document.Where, columns, "bool"); err != nil {
			return ErrTransformPlanInvalid
		}
	}
	return nil
}

func (p TransformPlanV1) valid() bool {
	if len(p.normalized) == 0 || len(p.outputs) == 0 || len(p.hash) != sha256.Size*2 {
		return false
	}
	parsed, err := ParseTransformPlanV1(p.normalized)
	if err != nil || parsed.hash != p.hash || len(parsed.outputs) != len(p.outputs) {
		return false
	}
	for index := range parsed.outputs {
		left, right := parsed.outputs[index], p.outputs[index]
		if left.Target != right.Target || left.Type != right.Type || left.Rounding != right.Rounding || left.Source != right.Source || !bytes.Equal(left.Expression, right.Expression) {
			return false
		}
	}
	return true
}

type transformPlanDocument struct {
	Version uint                          `json:"version"`
	Select  []transformProjectionDocument `json:"select"`
	Where   json.RawMessage               `json:"where,omitempty"`
}

type transformProjectionDocument struct {
	Source   string          `json:"source,omitempty"`
	Target   string          `json:"target"`
	Type     string          `json:"type"`
	Rounding string          `json:"rounding,omitempty"`
	Expr     json.RawMessage `json:"expr,omitempty"`
}

type normalizedTransformExpression struct {
	Operation string
	JSON      json.RawMessage
}

func normalizeTransformProjection(projection transformProjectionDocument) (TransformOutputColumnV1, transformProjectionDocument, error) {
	if !validTransformIdentifier(projection.Target) || transformLogicalTypeValid(projection.Type) == false || strings.TrimSpace(projection.Rounding) != projection.Rounding {
		return TransformOutputColumnV1{}, transformProjectionDocument{}, ErrTransformPlanInvalid
	}
	if (projection.Source == "") == (len(projection.Expr) == 0) {
		return TransformOutputColumnV1{}, transformProjectionDocument{}, ErrTransformPlanInvalid
	}
	output := TransformOutputColumnV1{Target: projection.Target, Type: projection.Type, Rounding: projection.Rounding}
	if projection.Source != "" {
		if !validTransformIdentifier(projection.Source) || projection.Rounding != "" {
			return TransformOutputColumnV1{}, transformProjectionDocument{}, ErrTransformPlanInvalid
		}
		output.Source = projection.Source
		return output, projection, nil
	}
	expression, err := normalizeTransformExpression(projection.Expr)
	if err != nil || !transformExpressionOutputAllowed(expression.Operation, projection.Type, projection.Rounding) {
		return TransformOutputColumnV1{}, transformProjectionDocument{}, ErrTransformPlanInvalid
	}
	output.Expression = append([]byte(nil), expression.JSON...)
	projection.Expr = append(json.RawMessage(nil), expression.JSON...)
	return output, projection, nil
}

func normalizeTransformExpression(raw []byte) (normalizedTransformExpression, error) {
	var object map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&object); err != nil || decoder.More() || len(object) != 1 {
		return normalizedTransformExpression{}, ErrTransformPlanInvalid
	}
	for operation, argument := range object {
		switch operation {
		case "date", "upper":
			var field string
			if err := json.Unmarshal(argument, &field); err != nil || !validTransformIdentifier(field) {
				return normalizedTransformExpression{}, ErrTransformPlanInvalid
			}
			encoded, _ := json.Marshal(map[string]string{operation: field})
			return normalizedTransformExpression{Operation: operation, JSON: encoded}, nil
		case "cast":
			inner, err := normalizeTransformExpression(argument)
			if err != nil || inner.Operation != "multiply" {
				return normalizedTransformExpression{}, ErrTransformPlanInvalid
			}
			encoded, _ := json.Marshal(struct {
				Value json.RawMessage `json:"cast"`
			}{Value: inner.JSON})
			return normalizedTransformExpression{Operation: operation, JSON: encoded}, nil
		case "multiply", "mod", "not_equal":
			values, err := normalizeTransformOperands(argument)
			if err != nil {
				return normalizedTransformExpression{}, err
			}
			encoded, _ := json.Marshal(map[string][]json.RawMessage{operation: values})
			return normalizedTransformExpression{Operation: operation, JSON: encoded}, nil
		default:
			return normalizedTransformExpression{}, ErrTransformPlanInvalid
		}
	}
	return normalizedTransformExpression{}, ErrTransformPlanInvalid
}

func normalizeTransformOperands(raw []byte) ([]json.RawMessage, error) {
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil || len(values) != 2 {
		return nil, ErrTransformPlanInvalid
	}
	normalized := make([]json.RawMessage, len(values))
	for index, value := range values {
		trimmed := bytes.TrimSpace(value)
		if len(trimmed) == 0 {
			return nil, ErrTransformPlanInvalid
		}
		if trimmed[0] == '{' {
			expression, err := normalizeTransformExpression(trimmed)
			if err != nil {
				return nil, err
			}
			normalized[index] = expression.JSON
			continue
		}
		var field string
		if err := json.Unmarshal(trimmed, &field); err == nil {
			if !validTransformIdentifier(field) {
				return nil, ErrTransformPlanInvalid
			}
			normalized[index], _ = json.Marshal(field)
			continue
		}
		decoder := json.NewDecoder(bytes.NewReader(trimmed))
		decoder.UseNumber()
		var number json.Number
		if err := decoder.Decode(&number); err != nil || decoder.More() || !validTransformNumber(number) {
			return nil, ErrTransformPlanInvalid
		}
		normalized[index] = append(json.RawMessage(nil), trimmed...)
	}
	return normalized, nil
}

func validTransformNumber(value json.Number) bool {
	if strings.ContainsAny(value.String(), "eE") || strings.TrimSpace(value.String()) != value.String() {
		return false
	}
	_, err := value.Int64()
	return err == nil
}

func validTransformIdentifier(value string) bool { return validateIdentifierComponent(value) == nil }

func transformLogicalTypeValid(name string) bool {
	_, err := transformLogicalType(name)
	return err == nil
}

func transformLogicalType(name string) (LogicalType, error) {
	switch name {
	case "int64":
		return NewSignedInteger(64)
	case "string":
		return NewString(0, "")
	case "date":
		return NewDate(), nil
	case "timestamp":
		// TransformPlanV1's timestamp is an instant-bearing timestamp with
		// timezone. A zone-neutral source cannot be silently promoted into an
		// instant, so direct timestamp validation below requires the same
		// explicit catalog semantics.
		return NewTimestamp(6, true)
	default:
		return LogicalType{}, ErrTransformPlanInvalid
	}
}

func transformExpressionOutputAllowed(operation, outputType, rounding string) bool {
	switch operation {
	case "date":
		return outputType == "date" && rounding == ""
	case "upper":
		return outputType == "string" && rounding == ""
	case "cast":
		return outputType == "int64" && rounding == "exact"
	default:
		return false
	}
}

func transformDirectTypeAllowed(source LogicalType, output string) bool {
	switch output {
	case "int64":
		return source.Kind() == LogicalSignedInteger && source.BitWidth() <= 64
	case "string":
		return source.Kind() == LogicalString
	case "date":
		return source.Kind() == LogicalDate
	case "timestamp":
		return source.Kind() == LogicalTimestamp && source.WithTimezone()
	default:
		return false
	}
}

func validateTransformExpressionAgainstColumns(raw []byte, columns map[string]Column, want string) error {
	expression, err := normalizeTransformExpression(raw)
	if err != nil {
		return err
	}
	switch expression.Operation {
	case "date":
		field, err := transformExpressionField(expression.JSON, "date")
		if err != nil || want != "date" || columns[field].Type.Kind() != LogicalTimestamp {
			return ErrTransformPlanInvalid
		}
	case "upper":
		field, err := transformExpressionField(expression.JSON, "upper")
		if err != nil || want != "string" || columns[field].Type.Kind() != LogicalString {
			return ErrTransformPlanInvalid
		}
	case "cast":
		var node struct {
			Cast json.RawMessage `json:"cast"`
		}
		if err := json.Unmarshal(expression.JSON, &node); err != nil || want != "int64" {
			return ErrTransformPlanInvalid
		}
		if err := validateTransformArithmetic(node.Cast, columns, "multiply"); err != nil {
			return ErrTransformPlanInvalid
		}
	case "not_equal":
		if want != "bool" || validateTransformArithmetic(expression.JSON, columns, "not_equal") != nil {
			return ErrTransformPlanInvalid
		}
	default:
		return ErrTransformPlanInvalid
	}
	return nil
}

func transformExpressionField(raw []byte, operation string) (string, error) {
	var node map[string]string
	if err := json.Unmarshal(raw, &node); err != nil || len(node) != 1 || !validTransformIdentifier(node[operation]) {
		return "", ErrTransformPlanInvalid
	}
	return node[operation], nil
}

func validateTransformArithmetic(raw []byte, columns map[string]Column, operation string) error {
	var node map[string][]json.RawMessage
	if err := json.Unmarshal(raw, &node); err != nil || len(node) != 1 || len(node[operation]) != 2 {
		return ErrTransformPlanInvalid
	}
	for _, operand := range node[operation] {
		trimmed := bytes.TrimSpace(operand)
		if len(trimmed) > 0 && trimmed[0] == '{' {
			expression, err := normalizeTransformExpression(trimmed)
			if err != nil || expression.Operation != "mod" || validateTransformArithmetic(expression.JSON, columns, "mod") != nil {
				return ErrTransformPlanInvalid
			}
			continue
		}
		var field string
		if err := json.Unmarshal(trimmed, &field); err == nil {
			column, found := columns[field]
			if !found || (column.Type.Kind() != LogicalSignedInteger && column.Type.Kind() != LogicalUnsignedInteger && column.Type.Kind() != LogicalDecimal) {
				return ErrTransformPlanInvalid
			}
			continue
		}
		var number json.Number
		decoder := json.NewDecoder(bytes.NewReader(trimmed))
		decoder.UseNumber()
		if err := decoder.Decode(&number); err != nil || !validTransformNumber(number) {
			return ErrTransformPlanInvalid
		}
	}
	return nil
}

func (p TransformPlanV1) String() string { return fmt.Sprintf("transform-plan-v1:%s", p.hash) }
