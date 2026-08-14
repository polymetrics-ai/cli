package database

import (
	"context"
	"errors"

	"polymetrics.ai/internal/warehouse"
)

// SortDirection is the closed direction vocabulary for a typed stable read
// order. It is not an arbitrary ORDER BY fragment.
type SortDirection string

const (
	SortAscending  SortDirection = "ascending"
	SortDescending SortDirection = "descending"
)

// OrderTerm is one relation-bound ordering term in a ReadPlan.
type OrderTerm struct {
	Column    ColumnRef
	Direction SortDirection
}

func (o OrderTerm) validate(relation Relation) error {
	if (o.Direction != SortAscending && o.Direction != SortDescending) ||
		!o.Column.Relation.equal(relation.Ref) || !containsColumn(relation.Columns, o.Column) {
		return errors.New("database read order term is invalid")
	}
	return nil
}

// ReadPlanRequest is the typed input for an immutable stable read plan. It
// contains catalog objects, ordering, limits, and one source-to-warehouse leg;
// it has no SQL, target, credential, or raw connection material.
type ReadPlanRequest struct {
	Inbound    WarehouseInboundRef
	Definition Definition
	Catalog    Catalog
	Relation   RelationRef
	Columns    []ColumnRef
	Order      []OrderTerm
	PageSize   int
}

// ReadPlan is a non-executing, immutable description of a stable paged read.
// A future driver may render it only through its own closed primitives.
type ReadPlan struct {
	inbound     WarehouseInboundRef
	relation    RelationRef
	fingerprint SchemaFingerprint
	columns     []ColumnRef
	order       []OrderTerm
	pageSize    int
}

// NewReadPlan validates a deterministic keyset read plan and observes a
// stable catalog fingerprint. It honors cancellation before inspecting input.
func NewReadPlan(ctx context.Context, request ReadPlanRequest) (ReadPlan, error) {
	if ctx == nil {
		return ReadPlan{}, errors.New("database read plan context is required")
	}
	if err := ctx.Err(); err != nil {
		return ReadPlan{}, err
	}
	if err := request.Inbound.validate(); err != nil {
		return ReadPlan{}, errors.New("database read plan warehouse inbound leg is invalid")
	}
	if err := request.Definition.Validate(); err != nil {
		return ReadPlan{}, errors.New("database read plan definition is invalid")
	}
	if err := request.Catalog.validate(); err != nil {
		return ReadPlan{}, errors.New("database read plan catalog is invalid")
	}
	if err := request.Relation.validate(); err != nil || !request.Inbound.Source().Relation().equal(request.Relation) {
		return ReadPlan{}, errors.New("database read plan relation is invalid")
	}
	pageSize, err := request.Definition.Resources().EffectivePageSize(request.PageSize)
	if err != nil {
		return ReadPlan{}, errors.New("database read plan page size is outside the declared resource bound")
	}
	relation, found := request.Catalog.relation(request.Relation)
	if !found {
		return ReadPlan{}, errors.New("database read plan relation is absent from the catalog")
	}
	if err := validateReadColumns(request.Columns, relation); err != nil {
		return ReadPlan{}, err
	}
	if err := validateReadOrder(request.Order, request.Columns, relation); err != nil {
		return ReadPlan{}, err
	}
	if err := ctx.Err(); err != nil {
		return ReadPlan{}, err
	}
	plan := ReadPlan{
		inbound:     request.Inbound,
		relation:    request.Relation,
		fingerprint: request.Catalog.Fingerprint(),
		columns:     append([]ColumnRef(nil), request.Columns...),
		order:       append([]OrderTerm(nil), request.Order...),
		pageSize:    pageSize,
	}
	if err := ctx.Err(); err != nil {
		return ReadPlan{}, err
	}
	return plan, nil
}

func validateReadColumns(columns []ColumnRef, relation Relation) error {
	if len(columns) == 0 {
		return errors.New("database read plan requires selected columns")
	}
	seen := make(map[string]struct{}, len(columns))
	for _, column := range columns {
		if !column.Relation.equal(relation.Ref) || !containsColumn(relation.Columns, column) {
			return errors.New("database read plan selects an unknown column")
		}
		if _, exists := seen[column.Name]; exists {
			return errors.New("database read plan selects a duplicate column")
		}
		seen[column.Name] = struct{}{}
	}
	return nil
}

func validateReadOrder(order []OrderTerm, selected []ColumnRef, relation Relation) error {
	if len(order) == 0 {
		return errors.New("database read plan requires deterministic ordering")
	}
	seen := make(map[string]struct{}, len(order))
	for _, term := range order {
		if err := term.validate(relation); err != nil {
			return err
		}
		if _, exists := seen[term.Column.Name]; exists {
			return errors.New("database read plan contains duplicate ordering columns")
		}
		if !containsColumnRef(selected, term.Column) {
			return errors.New("database read plan must select every ordering column")
		}
		seen[term.Column.Name] = struct{}{}
	}
	if !hasStableKeySuffix(order, relation) {
		return errors.New("database read plan order must end with a declared unique key")
	}
	return nil
}

func containsColumnRef(columns []ColumnRef, expected ColumnRef) bool {
	for _, column := range columns {
		if column.equal(expected) {
			return true
		}
	}
	return false
}

func hasStableKeySuffix(order []OrderTerm, relation Relation) bool {
	for _, key := range relation.Keys {
		if key.Kind != KeyPrimary && key.Kind != KeyUnique || len(key.Columns) > len(order) {
			continue
		}
		offset := len(order) - len(key.Columns)
		matches := true
		for index, column := range key.Columns {
			if !order[offset+index].Column.equal(column) {
				matches = false
				break
			}
		}
		if matches && keyColumnsAreNonNullable(key, relation.Columns) {
			return true
		}
	}
	return false
}

func keyColumnsAreNonNullable(key Key, columns []Column) bool {
	for _, keyColumn := range key.Columns {
		found := false
		for _, column := range columns {
			if column.Ref.equal(keyColumn) {
				found = true
				if column.Nullable {
					return false
				}
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Inbound returns the source-to-warehouse leg this plan may serve. It has no
// database target and therefore cannot encode a direct connector pair.
func (p ReadPlan) Inbound() WarehouseInboundRef { return p.inbound }

// Source returns the typed database source reference for compatibility with
// readers that need source catalog identity. The warehouse remains mandatory
// and is available through Warehouse().
func (p ReadPlan) Source() SourceRef { return p.inbound.Source() }

// Warehouse returns the connector-agnostic artifact into which layer one must
// land the extracted records.
func (p ReadPlan) Warehouse() warehouse.ArtifactRef { return p.inbound.Warehouse() }

// Relation returns the structured relation reference.
func (p ReadPlan) Relation() RelationRef { return p.relation }

// SchemaFingerprint returns the pinned catalog fingerprint.
func (p ReadPlan) SchemaFingerprint() SchemaFingerprint { return p.fingerprint }

// Columns returns a defensive selected-column projection.
func (p ReadPlan) Columns() []ColumnRef { return append([]ColumnRef(nil), p.columns...) }

// Order returns a defensive deterministic-order projection.
func (p ReadPlan) Order() []OrderTerm { return append([]OrderTerm(nil), p.order...) }

// PageSize returns the proven finite page size.
func (p ReadPlan) PageSize() int { return p.pageSize }
