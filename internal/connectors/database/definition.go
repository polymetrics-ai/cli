package database

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
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
	if err := m.Logical.validate(0); err != nil || m.Logical.Kind() == LogicalOpaqueNative {
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
// failure. Its text intentionally never includes a JSON value from the file.
var ErrInvalidDefinition = errors.New("database definition is invalid")

// DefinitionError keeps loader diagnostics in a small fixed vocabulary so a
// malformed manifest cannot exfiltrate a secret-looking value through errors.
type DefinitionError struct {
	reason string
}

func (e *DefinitionError) Error() string {
	if e == nil || e.reason == "" {
		return ErrInvalidDefinition.Error()
	}
	return ErrInvalidDefinition.Error() + ": " + e.reason
}

func (e *DefinitionError) Unwrap() error { return ErrInvalidDefinition }

func invalidDefinition(reason string) error { return &DefinitionError{reason: reason} }

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
	if !json.Valid(databaseDefinitionSchema) {
		return Definition{}, invalidDefinition("embedded schema is invalid")
	}
	raw, err := fs.ReadFile(fsys, "database.json")
	if err != nil {
		return Definition{}, invalidDefinition("database.json is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return Definition{}, err
	}

	var document definitionDocument
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return Definition{}, invalidDefinition("JSON must match the closed schema")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Definition{}, invalidDefinition("JSON must contain one document")
	}
	if err := ctx.Err(); err != nil {
		return Definition{}, err
	}
	definition, err := document.definition()
	if err != nil {
		return Definition{}, invalidDefinition("semantic validation failed")
	}
	return definition.clone(), nil
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
			ConnectTimeout:   time.Duration(d.Resources.ConnectTimeoutMS) * time.Millisecond,
			OperationTimeout: time.Duration(d.Resources.OperationTimeoutMS) * time.Millisecond,
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
