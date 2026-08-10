package database

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
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

// ConnectionIdentity is the opaque owner triple shared by source and target
// references. It deliberately excludes display names, server connection data,
// credential revisions, DSNs, and all secret material.
type ConnectionIdentity struct {
	WorkspaceID  string
	ConnectorID  string
	ConnectionID string
}

// Validate rejects an incomplete or unsafe opaque identity before it can reach
// a target path, catalog, or driver boundary.
func (i ConnectionIdentity) Validate() error {
	if !validOpaqueIdentityComponent(i.WorkspaceID) || !validConnectorID(i.ConnectorID) || !validOpaqueIdentityComponent(i.ConnectionID) {
		return errors.New("database connection identity is incomplete or invalid")
	}
	return nil
}

// SameIdentity ignores any display or credential concept because neither is a
// member of this type. It is the only equality rule used by source/target refs.
func (i ConnectionIdentity) SameIdentity(other ConnectionIdentity) bool {
	return i.WorkspaceID == other.WorkspaceID && i.ConnectorID == other.ConnectorID && i.ConnectionID == other.ConnectionID
}

// SourceRef pins a source owner triple and relation.
type SourceRef struct {
	identity ConnectionIdentity
	relation RelationRef
}

// NewSourceRef creates a typed source reference.
func NewSourceRef(identity ConnectionIdentity, relation RelationRef) (SourceRef, error) {
	if err := identity.Validate(); err != nil {
		return SourceRef{}, err
	}
	if err := relation.validate(); err != nil {
		return SourceRef{}, errors.New("database source relation is invalid")
	}
	return SourceRef{identity: identity, relation: relation}, nil
}

func (r SourceRef) validate() error {
	if err := r.identity.Validate(); err != nil {
		return err
	}
	return r.relation.validate()
}

// Identity returns the source's complete opaque owner triple.
func (r SourceRef) Identity() ConnectionIdentity { return r.identity }

// Relation returns the structured source relation.
func (r SourceRef) Relation() RelationRef { return r.relation }

// TargetRef pins a target owner triple and logical managed relation. F2 adds
// the ownership assertion/provisioning state machine; this value does not
// claim that a target exists or has been asserted.
type TargetRef struct {
	identity ConnectionIdentity
	relation RelationRef
}

// NewTargetRef creates a typed target reference without any DDL or ownership
// side effect.
func NewTargetRef(identity ConnectionIdentity, relation RelationRef) (TargetRef, error) {
	if err := identity.Validate(); err != nil {
		return TargetRef{}, err
	}
	if err := relation.validate(); err != nil {
		return TargetRef{}, errors.New("database target relation is invalid")
	}
	return TargetRef{identity: identity, relation: relation}, nil
}

// Identity returns the target's complete opaque owner triple.
func (r TargetRef) Identity() ConnectionIdentity { return r.identity }

// Relation returns the structured logical target relation.
func (r TargetRef) Relation() RelationRef { return r.relation }

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
