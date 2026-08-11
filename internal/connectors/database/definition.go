package database

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"math/big"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/synccontract"
)

// DatabaseDefinitionSchemaVersion is the only accepted database.json version
// for this foundation slice.
const DatabaseDefinitionSchemaVersion uint = 1

//go:embed schema/database.schema.json
var databaseDefinitionSchema []byte

// DefinitionSchema returns a copy of the closed JSON Schema documented by this
// package. Load additionally uses strict decoding and semantic validation.
func DefinitionSchema() []byte {
	return append([]byte(nil), databaseDefinitionSchema...)
}

// DriverDeclaration identifies a database driver contract. It is not an
// executable capability claim: DriverRegistry.Admit still requires a matching
// registered driver and shared native admission evidence.
type DriverDeclaration struct {
	ID         string
	Protocol   string
	APIVersion uint
}

func (d DriverDeclaration) validate() error {
	if !validOpaqueName(d.ID) || !validOpaqueName(d.Protocol) || d.APIVersion == 0 {
		return errors.New("database driver declaration is invalid")
	}
	return nil
}

// QualificationPart is one closed identifier qualification layer.
type QualificationPart string

const (
	QualificationCatalog  QualificationPart = "catalog"
	QualificationSchema   QualificationPart = "schema"
	QualificationRelation QualificationPart = "relation"
)

// CatalogPolicy controls only structured catalog qualification. It contains no
// user SQL or executable operation.
type CatalogPolicy struct {
	QualificationOrder []QualificationPart
}

func (p CatalogPolicy) clone() CatalogPolicy {
	p.QualificationOrder = append([]QualificationPart(nil), p.QualificationOrder...)
	return p
}

func (p CatalogPolicy) validate() error {
	if len(p.QualificationOrder) < 2 || len(p.QualificationOrder) > 3 || p.QualificationOrder[len(p.QualificationOrder)-1] != QualificationRelation {
		return errors.New("database catalog qualification policy is invalid")
	}
	seen := make(map[QualificationPart]struct{}, len(p.QualificationOrder))
	for _, part := range p.QualificationOrder {
		if part != QualificationCatalog && part != QualificationSchema && part != QualificationRelation {
			return errors.New("database catalog qualification policy is invalid")
		}
		if _, exists := seen[part]; exists {
			return errors.New("database catalog qualification policy is invalid")
		}
		seen[part] = struct{}{}
	}
	return nil
}

// QuoteStyle is a closed renderer-owned identifier quote family. F1 only
// validates its declaration; it does not render SQL.
type QuoteStyle string

const (
	QuoteDouble   QuoteStyle = "double_quote"
	QuoteBacktick QuoteStyle = "backtick"
	QuoteBracket  QuoteStyle = "bracket"
)

// CaseFold is a closed database identifier case policy.
type CaseFold string

const (
	CaseFoldLower    CaseFold = "lower"
	CaseFoldUpper    CaseFold = "upper"
	CaseFoldPreserve CaseFold = "preserve"
)

// IdentifierPolicy captures bounded, renderer-owned identifier rules. It
// cannot contain a caller-supplied quote function or raw SQL fragment.
type IdentifierPolicy struct {
	QuoteStyle QuoteStyle
	CaseFold   CaseFold
	MaxBytes   int
}

func (p IdentifierPolicy) validate() error {
	if (p.QuoteStyle != QuoteDouble && p.QuoteStyle != QuoteBacktick && p.QuoteStyle != QuoteBracket) ||
		(p.CaseFold != CaseFoldLower && p.CaseFold != CaseFoldUpper && p.CaseFold != CaseFoldPreserve) ||
		p.MaxBytes <= 0 || p.MaxBytes > 256 {
		return errors.New("database identifier policy is invalid")
	}
	return nil
}

// NativeType describes a native database type name plus structured modifiers.
// It is catalog data, never executable SQL text.
type NativeType struct {
	Name      string
	Modifiers []string
}

func (t NativeType) clone() NativeType {
	t.Modifiers = append([]string(nil), t.Modifiers...)
	return t
}

func (t NativeType) validate() error {
	if !validOpaqueName(t.Name) {
		return errors.New("database native type is invalid")
	}
	for _, modifier := range t.Modifiers {
		if !validOpaqueName(modifier) {
			return errors.New("database native type is invalid")
		}
	}
	return nil
}

func (t NativeType) key() string {
	return t.Name + "\x00" + joinTokens(t.Modifiers)
}

// TypeMapping maps one closed native catalog type to a closed logical type. A
// missing native type has no fallback; it becomes OpaqueNative at discovery.
type TypeMapping struct {
	Native  NativeType
	Logical LogicalType
}

func (m TypeMapping) clone() TypeMapping {
	return TypeMapping{Native: m.Native.clone(), Logical: m.Logical.clone()}
}

func (m TypeMapping) validate() error {
	if err := m.Native.validate(); err != nil {
		return err
	}
	if err := m.Logical.validate(0); err != nil || m.Logical.containsOpaqueNative() {
		return errors.New("database type mapping logical type is invalid")
	}
	return nil
}

// Definition is a loaded immutable database.json projection. Its fields stay
// private so mutable maps/slices cannot escape through a loaded bundle.
type Definition struct {
	schemaVersion uint
	driver        DriverDeclaration
	catalog       CatalogPolicy
	identifiers   IdentifierPolicy
	resources     ResourcePolicy
	typeMappings  []TypeMapping
	admittedModes []synccontract.Mode
}

// SchemaVersion returns the supported document version.
func (d Definition) SchemaVersion() uint { return d.schemaVersion }

// Driver returns the declared driver identity as a value copy.
func (d Definition) Driver() DriverDeclaration { return d.driver }

// CatalogPolicy returns a defensive copy of qualification policy slices.
func (d Definition) CatalogPolicy() CatalogPolicy { return d.catalog.clone() }

// IdentifierPolicy returns the closed identifier policy.
func (d Definition) IdentifierPolicy() IdentifierPolicy { return d.identifiers }

// Resources returns the finite resource policy by value.
func (d Definition) Resources() ResourcePolicy { return d.resources }

// TypeMappings returns a deep defensive projection.
func (d Definition) TypeMappings() []TypeMapping {
	mappings := make([]TypeMapping, len(d.typeMappings))
	for i := range d.typeMappings {
		mappings[i] = d.typeMappings[i].clone()
	}
	return mappings
}

// AdmittedModes returns a defensive copy of the canonical synccontract modes
// declared by a future driver. An empty slice claims no executable mode.
func (d Definition) AdmittedModes() []synccontract.Mode {
	return append([]synccontract.Mode(nil), d.admittedModes...)
}

func (d Definition) clone() Definition {
	clone := d
	clone.catalog = d.catalog.clone()
	clone.typeMappings = d.TypeMappings()
	clone.admittedModes = d.AdmittedModes()
	return clone
}

// Validate confirms the complete immutable projection still obeys the closed
// database definition contract.
func (d Definition) Validate() error {
	if d.schemaVersion != DatabaseDefinitionSchemaVersion {
		return errors.New("database definition schema version is not supported")
	}
	if err := d.driver.validate(); err != nil {
		return err
	}
	if err := d.catalog.validate(); err != nil {
		return err
	}
	if err := d.identifiers.validate(); err != nil {
		return err
	}
	if err := d.resources.validate(); err != nil {
		return err
	}
	if len(d.typeMappings) == 0 {
		return errors.New("database definition requires explicit type mappings")
	}
	seenMappings := make(map[string]struct{}, len(d.typeMappings))
	for _, mapping := range d.typeMappings {
		if err := mapping.validate(); err != nil {
			return err
		}
		key := mapping.Native.key()
		if _, exists := seenMappings[key]; exists {
			return errors.New("database definition contains duplicate native type mappings")
		}
		seenMappings[key] = struct{}{}
	}
	seenModes := make(map[synccontract.Mode]struct{}, len(d.admittedModes))
	for _, mode := range d.admittedModes {
		if err := mode.Validate(); err != nil {
			return errors.New("database definition declares an unsupported sync mode")
		}
		if _, exists := seenModes[mode]; exists {
			return errors.New("database definition declares a duplicate sync mode")
		}
		seenModes[mode] = struct{}{}
	}
	return nil
}

// ErrInvalidDefinition is returned for syntactic or semantic database.json
// failure.
var ErrInvalidDefinition = errors.New("database definition is invalid")

// DefinitionError keeps loader diagnostics bounded to schema-derived details.
type DefinitionError struct {
	reason string
	cause  error
}

func (e *DefinitionError) Error() string {
	if e == nil || e.reason == "" {
		return ErrInvalidDefinition.Error()
	}
	return ErrInvalidDefinition.Error() + ": " + e.reason
}

func (e *DefinitionError) Unwrap() []error {
	if e == nil || e.cause == nil {
		return []error{ErrInvalidDefinition}
	}
	return []error{ErrInvalidDefinition, e.cause}
}

func invalidDefinition(reason string, causes ...error) error {
	var cause error
	if len(causes) > 0 {
		cause = causes[0]
	}
	return &DefinitionError{reason: reason, cause: cause}
}

type definitionPathError struct {
	path   string
	reason string
	cause  error
}

func (e *definitionPathError) Error() string {
	if e == nil || e.path == "" {
		return "database.json is invalid"
	}
	if e.reason == "" {
		return e.path
	}
	return e.path + ": " + e.reason
}

func (e *definitionPathError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func invalidDefinitionAt(path, reason string, cause error) error {
	pathError := &definitionPathError{path: path, reason: reason, cause: cause}
	return invalidDefinition(pathError.Error(), pathError)
}

// Load reads and validates database.json from one connector bundle directory.
// It checks cancellation before I/O and before it returns an immutable
// projection. It never logs or returns raw manifest values.
func Load(ctx context.Context, fsys fs.FS) (Definition, error) {
	if ctx == nil {
		return Definition{}, invalidDefinition("context is required")
	}
	if err := ctx.Err(); err != nil {
		return Definition{}, err
	}
	schema, err := loadDefinitionSchema()
	if err != nil {
		return Definition{}, invalidDefinition("embedded schema is invalid")
	}
	raw, err := fs.ReadFile(fsys, "database.json")
	if err != nil {
		return Definition{}, invalidDefinition("database.json is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return Definition{}, err
	}
	positions, err := validateDefinitionJSON(raw, schema)
	if err != nil {
		return Definition{}, invalidDefinition(err.Error(), err)
	}

	var document definitionDocument
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return Definition{}, invalidDefinitionAt(definitionDecodePath(err, schema, positions), "JSON must match the closed schema", err)
	}
	if err := ctx.Err(); err != nil {
		return Definition{}, err
	}
	definition, err := document.definition()
	if err != nil {
		return Definition{}, invalidDefinition("semantic validation failed", err)
	}
	projection := definition.clone()
	if err := ctx.Err(); err != nil {
		return Definition{}, err
	}
	return projection, nil
}

type definitionSchemaDocument struct {
	Type                 string                     `json:"type"`
	Reference            string                     `json:"$ref"`
	Properties           map[string]json.RawMessage `json:"properties"`
	Items                json.RawMessage            `json:"items"`
	Minimum              json.RawMessage            `json:"minimum"`
	Maximum              json.RawMessage            `json:"maximum"`
	Enum                 []json.RawMessage          `json:"enum"`
	Required             []string                   `json:"required"`
	AdditionalProperties *bool                      `json:"additionalProperties"`
	Definitions          map[string]json.RawMessage `json:"definitions"`
}

type definitionSchemaNode struct {
	valueType   string
	properties  map[string]*definitionSchemaNode
	required    []string
	items       *definitionSchemaNode
	minimum     *big.Int
	maximum     *big.Int
	integerEnum []*big.Int
}

type definitionJSONPathPosition struct {
	path   string
	offset int64
}

func loadDefinitionSchema() (*definitionSchemaNode, error) {
	if !json.Valid(databaseDefinitionSchema) {
		return nil, errors.New("database definition schema is invalid")
	}
	var root definitionSchemaDocument
	if err := json.Unmarshal(databaseDefinitionSchema, &root); err != nil {
		return nil, err
	}
	return buildDefinitionSchemaNode(databaseDefinitionSchema, root.Definitions, make(map[string]*definitionSchemaNode))
}

func buildDefinitionSchemaNode(raw json.RawMessage, definitions map[string]json.RawMessage, references map[string]*definitionSchemaNode) (*definitionSchemaNode, error) {
	var document definitionSchemaDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	if document.Reference != "" {
		const prefix = "#/definitions/"
		if !strings.HasPrefix(document.Reference, prefix) {
			return nil, errors.New("database definition schema contains an unsupported reference")
		}
		name := strings.TrimPrefix(document.Reference, prefix)
		if name == "" || strings.Contains(name, "/") {
			return nil, errors.New("database definition schema contains an invalid reference")
		}
		if node, exists := references[name]; exists {
			return node, nil
		}
		definition, exists := definitions[name]
		if !exists {
			return nil, errors.New("database definition schema references an unavailable definition")
		}
		node := &definitionSchemaNode{}
		references[name] = node
		if err := populateDefinitionSchemaNode(node, definition, definitions, references); err != nil {
			return nil, err
		}
		return node, nil
	}

	node := &definitionSchemaNode{}
	if err := populateDefinitionSchemaNode(node, raw, definitions, references); err != nil {
		return nil, err
	}
	return node, nil
}

func populateDefinitionSchemaNode(node *definitionSchemaNode, raw json.RawMessage, definitions map[string]json.RawMessage, references map[string]*definitionSchemaNode) error {
	var document definitionSchemaDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		return err
	}
	if document.Reference != "" || node == nil || document.Type == "" {
		return errors.New("database definition schema contains an invalid node")
	}
	node.valueType = document.Type
	switch document.Type {
	case "object":
		if document.AdditionalProperties == nil || *document.AdditionalProperties {
			return errors.New("database definition schema object is not closed")
		}
		node.properties = make(map[string]*definitionSchemaNode, len(document.Properties))
		for name, property := range document.Properties {
			propertyNode, err := buildDefinitionSchemaNode(property, definitions, references)
			if err != nil {
				return err
			}
			for existing := range node.properties {
				if strings.EqualFold(existing, name) {
					return errors.New("database definition schema contains ambiguous property names")
				}
			}
			node.properties[name] = propertyNode
		}
		node.required = append([]string(nil), document.Required...)
		for _, name := range node.required {
			if _, exists := node.properties[name]; !exists {
				return errors.New("database definition schema requires an unavailable property")
			}
		}
	case "array":
		if len(document.Items) == 0 {
			return errors.New("database definition schema array has no item contract")
		}
		items, err := buildDefinitionSchemaNode(document.Items, definitions, references)
		if err != nil {
			return err
		}
		node.items = items
	case "integer":
		minimum, err := definitionSchemaIntegerConstraint(document.Minimum)
		if err != nil {
			return err
		}
		maximum, err := definitionSchemaIntegerConstraint(document.Maximum)
		if err != nil {
			return err
		}
		if minimum != nil && maximum != nil && minimum.Cmp(maximum) > 0 {
			return errors.New("database definition schema integer bounds are invalid")
		}
		integerEnum, err := definitionSchemaIntegerEnum(document.Enum)
		if err != nil {
			return err
		}
		node.minimum = minimum
		node.maximum = maximum
		node.integerEnum = integerEnum
	case "string", "boolean":
	default:
		return errors.New("database definition schema contains an unsupported type")
	}
	return nil
}

func definitionSchemaIntegerConstraint(raw json.RawMessage) (*big.Int, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	value, ok := definitionJSONInteger(strings.TrimSpace(string(raw)))
	if !ok {
		return nil, errors.New("database definition schema integer constraint is invalid")
	}
	return value, nil
}

func definitionSchemaIntegerEnum(values []json.RawMessage) ([]*big.Int, error) {
	if len(values) == 0 {
		return nil, nil
	}
	enum := make([]*big.Int, len(values))
	for index, raw := range values {
		value, err := definitionSchemaIntegerConstraint(raw)
		if err != nil || value == nil {
			return nil, errors.New("database definition schema integer enum is invalid")
		}
		enum[index] = value
	}
	return enum, nil
}

func definitionJSONInteger(value string) (*big.Int, bool) {
	parsed, ok := new(big.Int).SetString(value, 10)
	return parsed, ok
}

func validateDefinitionJSON(raw []byte, schema *definitionSchemaNode) ([]definitionJSONPathPosition, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	positions := make([]definitionJSONPathPosition, 0)
	if err := validateDefinitionJSONValue(decoder, schema, "$", &positions); err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err != nil {
			return nil, &definitionPathError{path: "$", reason: "JSON is malformed", cause: err}
		}
		return nil, &definitionPathError{path: "$", reason: "JSON must contain one document"}
	}
	return positions, nil
}

func validateDefinitionJSONValue(decoder *json.Decoder, schema *definitionSchemaNode, path string, positions *[]definitionJSONPathPosition) error {
	if schema == nil {
		return &definitionPathError{path: path, reason: "closed schema is invalid"}
	}
	token, err := decoder.Token()
	if err != nil {
		return &definitionPathError{path: path, reason: "JSON is malformed", cause: err}
	}
	if token == nil {
		return &definitionPathError{path: path, reason: "null is not permitted"}
	}
	offset := decoder.InputOffset()

	switch schema.valueType {
	case "object":
		return validateDefinitionJSONObject(decoder, token, schema, path, positions)
	case "array":
		return validateDefinitionJSONArray(decoder, token, schema, path, positions)
	case "integer":
		if number, ok := token.(json.Number); ok {
			if err := validateDefinitionJSONIntegerConstraints(number, schema, path); err != nil {
				return err
			}
		}
		if err := consumeDefinitionJSONToken(decoder, token); err != nil {
			return &definitionPathError{path: path, reason: "JSON is malformed", cause: err}
		}
	case "string", "boolean":
		if err := consumeDefinitionJSONToken(decoder, token); err != nil {
			return &definitionPathError{path: path, reason: "JSON is malformed", cause: err}
		}
	default:
		return &definitionPathError{path: path, reason: "closed schema is invalid"}
	}
	*positions = append(*positions, definitionJSONPathPosition{path: path, offset: offset})
	return nil
}

func validateDefinitionJSONIntegerConstraints(number json.Number, schema *definitionSchemaNode, path string) error {
	value, ok := definitionJSONInteger(string(number))
	if !ok {
		return nil
	}
	if schema.minimum != nil && value.Cmp(schema.minimum) < 0 {
		return &definitionPathError{path: path, reason: "value " + string(number) + " violates minimum " + schema.minimum.String()}
	}
	if schema.maximum != nil && value.Cmp(schema.maximum) > 0 {
		return &definitionPathError{path: path, reason: "value " + string(number) + " violates maximum " + schema.maximum.String()}
	}
	if len(schema.integerEnum) > 0 {
		for _, allowed := range schema.integerEnum {
			if value.Cmp(allowed) == 0 {
				return nil
			}
		}
		return &definitionPathError{path: path, reason: "value " + string(number) + " violates enum " + definitionSchemaIntegerEnumDescription(schema.integerEnum)}
	}
	return nil
}

func definitionSchemaIntegerEnumDescription(values []*big.Int) string {
	parts := make([]string, len(values))
	for index, value := range values {
		parts[index] = value.String()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func validateDefinitionJSONObject(decoder *json.Decoder, token json.Token, schema *definitionSchemaNode, path string, positions *[]definitionJSONPathPosition) error {
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return &definitionPathError{path: path, reason: "object value is required"}
	}
	seen := make(map[string]struct{}, len(schema.properties))
	for decoder.More() {
		memberToken, err := decoder.Token()
		if err != nil {
			return &definitionPathError{path: path, reason: "JSON is malformed", cause: err}
		}
		member, ok := memberToken.(string)
		if !ok {
			return &definitionPathError{path: path, reason: "object member name is required"}
		}
		canonical, property, exact := definitionSchemaProperty(schema, member)
		if property == nil {
			return &definitionPathError{path: path, reason: "unknown member is not permitted"}
		}
		memberPath := definitionObjectPath(path, canonical)
		if !exact {
			return &definitionPathError{path: memberPath, reason: "case-aliased member is not permitted"}
		}
		if _, exists := seen[canonical]; exists {
			return &definitionPathError{path: memberPath, reason: "duplicate member is not permitted"}
		}
		seen[canonical] = struct{}{}
		if err := validateDefinitionJSONValue(decoder, property, memberPath, positions); err != nil {
			return err
		}
	}
	end, err := decoder.Token()
	if err != nil {
		return &definitionPathError{path: path, reason: "JSON is malformed", cause: err}
	}
	if delimiter, ok := end.(json.Delim); !ok || delimiter != '}' {
		return &definitionPathError{path: path, reason: "JSON object is incomplete"}
	}
	for _, required := range schema.required {
		if _, exists := seen[required]; !exists {
			return &definitionPathError{path: definitionObjectPath(path, required), reason: "required member is missing"}
		}
	}
	return nil
}

func validateDefinitionJSONArray(decoder *json.Decoder, token json.Token, schema *definitionSchemaNode, path string, positions *[]definitionJSONPathPosition) error {
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
		return &definitionPathError{path: path, reason: "array value is required"}
	}
	for index := 0; decoder.More(); index++ {
		if err := validateDefinitionJSONValue(decoder, schema.items, definitionArrayPath(path, index), positions); err != nil {
			return err
		}
	}
	end, err := decoder.Token()
	if err != nil {
		return &definitionPathError{path: path, reason: "JSON is malformed", cause: err}
	}
	if delimiter, ok := end.(json.Delim); !ok || delimiter != ']' {
		return &definitionPathError{path: path, reason: "JSON array is incomplete"}
	}
	return nil
}

func consumeDefinitionJSONToken(decoder *json.Decoder, token json.Token) error {
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		for decoder.More() {
			if _, err := decoder.Token(); err != nil {
				return err
			}
			value, err := decoder.Token()
			if err != nil {
				return err
			}
			if err := consumeDefinitionJSONToken(decoder, value); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim('}') {
			return errors.New("JSON object is incomplete")
		}
	case '[':
		for decoder.More() {
			value, err := decoder.Token()
			if err != nil {
				return err
			}
			if err := consumeDefinitionJSONToken(decoder, value); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim(']') {
			return errors.New("JSON array is incomplete")
		}
	default:
		return errors.New("JSON value is malformed")
	}
	return nil
}

func definitionSchemaProperty(schema *definitionSchemaNode, member string) (string, *definitionSchemaNode, bool) {
	if schema == nil {
		return "", nil, false
	}
	if property, exists := schema.properties[member]; exists {
		return member, property, true
	}
	for name, property := range schema.properties {
		if strings.EqualFold(name, member) {
			return name, property, false
		}
	}
	return "", nil, false
}

func definitionDecodePath(err error, schema *definitionSchemaNode, positions []definitionJSONPathPosition) string {
	var typeError *json.UnmarshalTypeError
	if !errors.As(err, &typeError) {
		return "$"
	}
	for _, position := range positions {
		if position.offset == typeError.Offset {
			return position.path
		}
	}
	path := "$"
	node := schema
	for _, member := range strings.Split(typeError.Field, ".") {
		for node != nil && node.valueType == "array" {
			node = node.items
		}
		canonical, property, _ := definitionSchemaProperty(node, member)
		if property == nil {
			return path
		}
		path = definitionObjectPath(path, canonical)
		node = property
	}
	return path
}

func definitionObjectPath(path, member string) string {
	if path == "$" {
		return "$." + member
	}
	return path + "." + member
}

func definitionArrayPath(path string, index int) string {
	return path + "[" + strconv.Itoa(index) + "]"
}

type definitionDocument struct {
	SchemaVersion uint                     `json:"schema_version"`
	Driver        driverDocument           `json:"driver"`
	Catalog       catalogPolicyDocument    `json:"catalog"`
	Identifiers   identifierPolicyDocument `json:"identifiers"`
	Resources     resourcePolicyDocument   `json:"resources"`
	TypeMappings  []typeMappingDocument    `json:"type_mappings"`
	AdmittedModes *[]synccontract.Mode     `json:"admitted_modes"`
}

type driverDocument struct {
	ID         string `json:"id"`
	Protocol   string `json:"protocol"`
	APIVersion uint   `json:"api_version"`
}

type catalogPolicyDocument struct {
	QualificationOrder []QualificationPart `json:"qualification_order"`
}

type identifierPolicyDocument struct {
	QuoteStyle QuoteStyle `json:"quote_style"`
	CaseFold   CaseFold   `json:"case_fold"`
	MaxBytes   int        `json:"max_bytes"`
}

type resourceLimitDocument struct {
	Default int `json:"default"`
	Maximum int `json:"maximum"`
}

type resourcePolicyDocument struct {
	ReadPage           resourceLimitDocument `json:"read_page"`
	WriteBatch         resourceLimitDocument `json:"write_batch"`
	Pool               resourceLimitDocument `json:"pool"`
	ConnectTimeoutMS   int                   `json:"connect_timeout_ms"`
	OperationTimeoutMS int                   `json:"operation_timeout_ms"`
	MaxParameters      int                   `json:"max_parameters"`
}

type typeMappingDocument struct {
	Native  nativeTypeDocument  `json:"native"`
	Logical logicalTypeDocument `json:"logical"`
}

type nativeTypeDocument struct {
	Name      string   `json:"name"`
	Modifiers []string `json:"modifiers"`
}

type logicalTypeDocument struct {
	Kind          LogicalKind          `json:"kind"`
	Bits          uint8                `json:"bits"`
	Precision     uint16               `json:"precision"`
	Scale         uint16               `json:"scale"`
	MaxLength     uint32               `json:"max_length"`
	Collation     string               `json:"collation"`
	WithTimezone  *bool                `json:"with_timezone"`
	Element       *logicalTypeDocument `json:"element"`
	OpaqueEngine  string               `json:"opaque_engine"`
	OpaqueName    string               `json:"opaque_name"`
	OpaqueOptions []string             `json:"opaque_options"`
}

func (d definitionDocument) definition() (Definition, error) {
	if d.AdmittedModes == nil {
		return Definition{}, errors.New("admitted modes are required")
	}
	connectTimeout, err := durationFromMilliseconds(d.Resources.ConnectTimeoutMS, hardMaximumConnectTimeout)
	if err != nil {
		return Definition{}, err
	}
	operationTimeout, err := durationFromMilliseconds(d.Resources.OperationTimeoutMS, hardMaximumOperationTimeout)
	if err != nil {
		return Definition{}, err
	}
	mappings := make([]TypeMapping, len(d.TypeMappings))
	for i, mapping := range d.TypeMappings {
		logical, err := mapping.Logical.logicalType()
		if err != nil {
			return Definition{}, err
		}
		mappings[i] = TypeMapping{
			Native:  NativeType{Name: mapping.Native.Name, Modifiers: append([]string(nil), mapping.Native.Modifiers...)},
			Logical: logical,
		}
	}
	definition := Definition{
		schemaVersion: d.SchemaVersion,
		driver: DriverDeclaration{
			ID:         d.Driver.ID,
			Protocol:   d.Driver.Protocol,
			APIVersion: d.Driver.APIVersion,
		},
		catalog: CatalogPolicy{QualificationOrder: append([]QualificationPart(nil), d.Catalog.QualificationOrder...)},
		identifiers: IdentifierPolicy{
			QuoteStyle: d.Identifiers.QuoteStyle,
			CaseFold:   d.Identifiers.CaseFold,
			MaxBytes:   d.Identifiers.MaxBytes,
		},
		resources: ResourcePolicy{
			ReadPage:         Limit{Default: d.Resources.ReadPage.Default, Maximum: d.Resources.ReadPage.Maximum},
			WriteBatch:       Limit{Default: d.Resources.WriteBatch.Default, Maximum: d.Resources.WriteBatch.Maximum},
			Pool:             Limit{Default: d.Resources.Pool.Default, Maximum: d.Resources.Pool.Maximum},
			ConnectTimeout:   connectTimeout,
			OperationTimeout: operationTimeout,
			MaxParameters:    d.Resources.MaxParameters,
		},
		typeMappings:  mappings,
		admittedModes: append([]synccontract.Mode(nil), (*d.AdmittedModes)...),
	}
	if err := definition.Validate(); err != nil {
		return Definition{}, err
	}
	return definition, nil
}

func durationFromMilliseconds(milliseconds int, maximum time.Duration) (time.Duration, error) {
	if milliseconds <= 0 || milliseconds > int(maximum/time.Millisecond) {
		return 0, errors.New("database timeout is outside the finite resource bound")
	}
	return time.Duration(milliseconds) * time.Millisecond, nil
}

func (d logicalTypeDocument) logicalType() (LogicalType, error) {
	typeValue := LogicalType{
		kind:          d.Kind,
		bits:          d.Bits,
		precision:     d.Precision,
		scale:         d.Scale,
		maxLength:     d.MaxLength,
		collation:     d.Collation,
		opaqueEngine:  d.OpaqueEngine,
		opaqueName:    d.OpaqueName,
		opaqueOptions: append([]string(nil), d.OpaqueOptions...),
	}
	if d.WithTimezone != nil {
		typeValue.withTimezone = *d.WithTimezone
	}
	if d.Element != nil {
		element, err := d.Element.logicalType()
		if err != nil {
			return LogicalType{}, err
		}
		typeValue.element = &element
	}
	if (d.Kind == LogicalTime || d.Kind == LogicalTimestamp) && d.WithTimezone == nil {
		return LogicalType{}, errors.New("temporal timezone semantics are required")
	}
	if d.Kind != LogicalTime && d.Kind != LogicalTimestamp && d.WithTimezone != nil {
		return LogicalType{}, errors.New("non-temporal timezone semantics are invalid")
	}
	return newLogicalType(typeValue)
}

func joinTokens(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}
	return strings.Join(tokens, "\x00")
}
