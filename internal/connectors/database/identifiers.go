package database

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"polymetrics.ai/internal/warehouse"
)

// CatalogRef is one structured catalog/database identifier component.
type CatalogRef struct {
	Name string
}

func (r CatalogRef) validate() error { return validateIdentifierComponent(r.Name) }

// SchemaRef is a schema scoped to its catalog rather than a dot-concatenated
// string. This prevents a qualification choice from becoming SQL text.
type SchemaRef struct {
	Catalog CatalogRef
	Name    string
}

func (r SchemaRef) validate() error {
	if err := r.Catalog.validate(); err != nil {
		return errors.New("database schema has an invalid catalog reference")
	}
	if err := validateIdentifierComponent(r.Name); err != nil {
		return errors.New("database schema reference is invalid")
	}
	return nil
}

// RelationRef identifies one relation/table through structured catalog and
// schema components. It never carries a raw qualified SQL expression.
type RelationRef struct {
	Schema SchemaRef
	Name   string
}

func (r RelationRef) validate() error {
	if err := r.Schema.validate(); err != nil {
		return errors.New("database relation has an invalid schema reference")
	}
	if err := validateIdentifierComponent(r.Name); err != nil {
		return errors.New("database relation reference is invalid")
	}
	return nil
}

func (r RelationRef) equal(other RelationRef) bool {
	return r.Schema.Catalog.Name == other.Schema.Catalog.Name &&
		r.Schema.Name == other.Schema.Name && r.Name == other.Name
}

// ColumnRef identifies one relation-bound column. A column may not float free
// as an arbitrary expression or qualification fragment.
type ColumnRef struct {
	Relation RelationRef
	Name     string
}

func (r ColumnRef) validate() error {
	if err := r.Relation.validate(); err != nil {
		return errors.New("database column has an invalid relation reference")
	}
	if err := validateIdentifierComponent(r.Name); err != nil {
		return errors.New("database column reference is invalid")
	}
	return nil
}

func (r ColumnRef) equal(other ColumnRef) bool {
	return r.Relation.equal(other.Relation) && r.Name == other.Name
}

// ConnectionIdentity aliases the shared warehouse artifact owner triple. The
// database layer cannot invent a second source/target identity rule; it stays
// aligned with the structural identity already used by warehouse paths.
type ConnectionIdentity = warehouse.ArtifactIdentity

func validateDatabaseConnectionIdentity(identity ConnectionIdentity) error {
	if err := identity.Validate(); err != nil || !validConnectorID(identity.ConnectorID) {
		return errors.New("database connection identity is incomplete or invalid")
	}
	return nil
}

// SourceRef pins a source owner triple and relation.
type SourceRef struct {
	identity ConnectionIdentity
	relation RelationRef
}

// NewSourceRef creates a typed source reference.
func NewSourceRef(identity ConnectionIdentity, relation RelationRef) (SourceRef, error) {
	if err := validateDatabaseConnectionIdentity(identity); err != nil {
		return SourceRef{}, err
	}
	if err := relation.validate(); err != nil {
		return SourceRef{}, errors.New("database source relation is invalid")
	}
	return SourceRef{identity: identity, relation: relation}, nil
}

func (r SourceRef) validate() error {
	if err := validateDatabaseConnectionIdentity(r.identity); err != nil {
		return err
	}
	return r.relation.validate()
}

// Identity returns the source's complete opaque owner triple.
func (r SourceRef) Identity() ConnectionIdentity { return r.identity }

// Relation returns the structured source relation.
func (r SourceRef) Relation() RelationRef { return r.relation }

// TargetRef pins the destination connection identity and logical relation. It
// is not a managed-target owner: F2 derives that owner from the source-owned
// warehouse artifact, then adds the in-database assertion/provisioning state
// machine. This value does not claim that a target exists or has been asserted.
type TargetRef struct {
	identity ConnectionIdentity
	relation RelationRef
}

// NewTargetRef creates a typed destination connection reference without any
// DDL or ownership side effect.
func NewTargetRef(identity ConnectionIdentity, relation RelationRef) (TargetRef, error) {
	if err := validateDatabaseConnectionIdentity(identity); err != nil {
		return TargetRef{}, err
	}
	if err := relation.validate(); err != nil {
		return TargetRef{}, errors.New("database target relation is invalid")
	}
	return TargetRef{identity: identity, relation: relation}, nil
}

func (r TargetRef) validate() error {
	if err := validateDatabaseConnectionIdentity(r.identity); err != nil {
		return err
	}
	return r.relation.validate()
}

// Identity returns the target database connection's complete opaque identity.
func (r TargetRef) Identity() ConnectionIdentity { return r.identity }

// Relation returns the structured logical target relation.
func (r TargetRef) Relation() RelationRef { return r.relation }

// WarehouseInboundRef is the database-specific extraction side of the shared
// warehouse mediator: a database source can land only in the warehouse
// artifact owned by that source connection. It deliberately has no target.
type WarehouseInboundRef struct {
	source    SourceRef
	warehouse warehouse.ArtifactRef
}

// NewWarehouseInboundRef binds a database source to its own durable warehouse
// artifact. It does not open the database or write an artifact; shared layer
// one owns those operations.
func NewWarehouseInboundRef(source SourceRef, artifact warehouse.ArtifactRef) (WarehouseInboundRef, error) {
	ref := WarehouseInboundRef{source: source, warehouse: artifact}
	if err := ref.validate(); err != nil {
		return WarehouseInboundRef{}, err
	}
	return ref, nil
}

func (r WarehouseInboundRef) validate() error {
	if err := r.source.validate(); err != nil {
		return errors.New("database warehouse inbound source is invalid")
	}
	if err := r.warehouse.Validate(); err != nil {
		return errors.New("database warehouse inbound artifact is invalid")
	}
	if !r.source.Identity().SameIdentity(r.warehouse.Identity()) {
		return errors.New("database warehouse inbound artifact is not owned by the source connection")
	}
	return nil
}

// Source returns the source side of an inbound-only database leg.
func (r WarehouseInboundRef) Source() SourceRef { return r.source }

// Warehouse returns the neutral shared warehouse artifact for this inbound
// leg. It is not a database-specific materializer.
func (r WarehouseInboundRef) Warehouse() warehouse.ArtifactRef { return r.warehouse }

// WarehouseOutboundRef is the database-specific apply side of the shared
// warehouse mediator: it receives rows from one source-owned warehouse
// artifact and names one database target connection. It deliberately has no
// source reference; source ownership remains on the artifact.
type WarehouseOutboundRef struct {
	warehouse warehouse.ArtifactRef
	target    TargetRef
}

// NewWarehouseOutboundRef binds a warehouse artifact to a database target
// without creating a direct source-to-target route or executing any write.
func NewWarehouseOutboundRef(artifact warehouse.ArtifactRef, target TargetRef) (WarehouseOutboundRef, error) {
	ref := WarehouseOutboundRef{warehouse: artifact, target: target}
	if err := ref.validate(); err != nil {
		return WarehouseOutboundRef{}, err
	}
	return ref, nil
}

func (r WarehouseOutboundRef) validate() error {
	if err := r.warehouse.Validate(); err != nil {
		return errors.New("database warehouse outbound artifact is invalid")
	}
	if err := r.target.validate(); err != nil {
		return errors.New("database warehouse outbound target is invalid")
	}
	return nil
}

// Warehouse returns the shared warehouse side of an outbound-only database
// leg.
func (r WarehouseOutboundRef) Warehouse() warehouse.ArtifactRef { return r.warehouse }

// Target returns the database destination side of an outbound-only leg.
func (r WarehouseOutboundRef) Target() TargetRef { return r.target }

// NativeRelationIdentity is an opaque database-native relation identifier
// (such as a PostgreSQL OID) observed by a future driver. Its value is never
// used as a SQL fragment and errors never render it.
type NativeRelationIdentity struct {
	Kind  string
	Value string
}

func (i NativeRelationIdentity) validate() error {
	if i.Kind == "" && i.Value == "" {
		return nil
	}
	if !validOpaqueName(i.Kind) || !validOpaqueIdentityComponent(i.Value) {
		return errors.New("database native relation identity is invalid")
	}
	return nil
}

func validateIdentifierComponent(value string) error {
	if value == "" || len(value) > 256 || strings.TrimSpace(value) != value || !utf8.ValidString(value) {
		return errors.New("database identifier component is invalid")
	}
	for _, r := range value {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '_', r == '$', r == '-':
			continue
		default:
			return errors.New("database identifier component is invalid")
		}
	}
	return nil
}

func validOpaqueIdentityComponent(value string) bool {
	if value == "" || len(value) > 256 || strings.TrimSpace(value) != value || !utf8.ValidString(value) {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}

func validConnectorID(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for index, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			if index == 0 && r == '-' {
				return false
			}
			continue
		}
		return false
	}
	return true
}
