package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
)

var ErrConnectorCatalogNotRepresentable = errors.New("connector catalog is not representable as a managed target schema")

// CatalogForManagedTargetSource resolves one sealed source relation for the
// managed-target boundary. Native databases retain their exact typed catalog;
// declarative APIs are projected from the connector's authoritative JSON
// schema instead of inferring a target schema from provider values.
func CatalogForManagedTargetSource(ctx context.Context, source connectors.Connector, runtime connectors.RuntimeConfig, stream string) (Catalog, error) {
	if ctx == nil || source == nil {
		return Catalog{}, ErrConnectorCatalogNotRepresentable
	}
	if typed, ok := source.(interface {
		TypedCatalog(context.Context, connectors.RuntimeConfig) (Catalog, error)
	}); ok {
		catalog, err := typed.TypedCatalog(ctx, runtime)
		if err != nil {
			return Catalog{}, err
		}
		relation, err := relationForManagedTargetStream(catalog, stream)
		if err != nil {
			return Catalog{}, err
		}
		sealed, err := NewCatalog(catalog.Ref(), []Relation{relation})
		if err != nil {
			return Catalog{}, fmt.Errorf("%w: %v", ErrConnectorCatalogNotRepresentable, err)
		}
		return sealed, nil
	}
	public, err := source.Catalog(ctx, runtime)
	if err != nil {
		return Catalog{}, err
	}
	return CatalogFromConnectorCatalog(public, stream)
}

// CatalogFromConnectorCatalog projects exactly one public connector stream
// into the closed database schema vocabulary. The connector's declared JSON
// schema is mandatory: runtime rows never become schema authority.
func CatalogFromConnectorCatalog(catalog connectors.Catalog, streamName string) (Catalog, error) {
	if strings.TrimSpace(catalog.Connector) == "" || strings.TrimSpace(streamName) == "" {
		return Catalog{}, ErrConnectorCatalogNotRepresentable
	}
	var selected *connectors.Stream
	for index := range catalog.Streams {
		if catalog.Streams[index].Name != streamName {
			continue
		}
		if selected != nil {
			return Catalog{}, ErrConnectorCatalogNotRepresentable
		}
		copy := catalog.Streams[index]
		selected = &copy
	}
	if selected == nil || len(selected.Schema) == 0 {
		return Catalog{}, ErrConnectorCatalogNotRepresentable
	}

	var document struct {
		Type       json.RawMessage            `json:"type"`
		Required   []string                   `json:"required"`
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(selected.Schema, &document); err != nil || !schemaTypeIncludesExactly(document.Type, "object") || len(document.Properties) == 0 {
		return Catalog{}, ErrConnectorCatalogNotRepresentable
	}
	catalogRef := CatalogRef{Name: catalog.Connector}
	relationRef := RelationRef{Schema: SchemaRef{Catalog: catalogRef, Name: "api"}, Name: streamName}
	required := make(map[string]bool, len(document.Required))
	for _, name := range document.Required {
		required[name] = true
	}
	names := make([]string, 0, len(document.Properties))
	for name := range document.Properties {
		names = append(names, name)
	}
	sort.Strings(names)
	columns := make([]Column, 0, len(names))
	byName := make(map[string]ColumnRef, len(names))
	for ordinal, name := range names {
		logical, explicitNullable, err := connectorSchemaLogicalType(document.Properties[name])
		if err != nil {
			return Catalog{}, err
		}
		ref := ColumnRef{Relation: relationRef, Name: name}
		columns = append(columns, Column{Ref: ref, Type: logical, Nullable: explicitNullable || !required[name], Ordinal: ordinal + 1})
		byName[name] = ref
	}

	keys := make([]Key, 0, 1)
	if len(selected.PrimaryKey) > 0 {
		keyColumns := make([]ColumnRef, len(selected.PrimaryKey))
		seen := make(map[string]bool, len(selected.PrimaryKey))
		for index, name := range selected.PrimaryKey {
			ref, ok := byName[name]
			if !ok || seen[name] || !required[name] {
				return Catalog{}, ErrConnectorCatalogNotRepresentable
			}
			seen[name] = true
			keyColumns[index] = ref
		}
		keys = append(keys, Key{Name: "api_primary_key", Kind: KeyPrimary, Columns: keyColumns})
	}
	relation := Relation{Ref: relationRef, Columns: columns, Keys: keys}
	typed, err := NewCatalog(catalogRef, []Relation{relation})
	if err != nil {
		return Catalog{}, fmt.Errorf("%w: %v", ErrConnectorCatalogNotRepresentable, err)
	}
	return typed, nil
}

func relationForManagedTargetStream(catalog Catalog, stream string) (Relation, error) {
	parts := strings.Split(stream, ".")
	if len(parts) == 0 || len(parts) > 2 {
		return Relation{}, ErrConnectorCatalogNotRepresentable
	}
	wantSchema, wantRelation := "", parts[len(parts)-1]
	if len(parts) == 2 {
		wantSchema = parts[0]
	}
	for _, relation := range catalog.Relations() {
		if relation.Ref.Name == wantRelation && (wantSchema == "" || relation.Ref.Schema.Name == wantSchema) {
			return relation, nil
		}
	}
	return Relation{}, ErrConnectorCatalogNotRepresentable
}

func connectorSchemaLogicalType(raw json.RawMessage) (LogicalType, bool, error) {
	var node struct {
		Type   json.RawMessage `json:"type"`
		Format string          `json:"format"`
	}
	if err := json.Unmarshal(raw, &node); err != nil {
		return LogicalType{}, false, ErrConnectorCatalogNotRepresentable
	}
	types, err := schemaTypes(node.Type)
	if err != nil {
		return LogicalType{}, false, err
	}
	nullable := false
	concrete := make([]string, 0, len(types))
	for _, value := range types {
		if value == "null" {
			nullable = true
			continue
		}
		concrete = append(concrete, value)
	}
	if len(concrete) != 1 {
		return LogicalType{}, false, ErrConnectorCatalogNotRepresentable
	}
	var logical LogicalType
	switch concrete[0] {
	case "integer":
		logical, err = NewSignedInteger(64)
	case "number":
		logical, err = NewFloat(64)
	case "boolean":
		logical = NewBoolean()
	case "object", "array":
		logical = NewJSON()
	case "string":
		switch node.Format {
		case "date":
			logical = NewDate()
		case "date-time":
			logical, err = NewTimestamp(6, true)
		case "time":
			logical, err = NewTime(6, true)
		case "uuid":
			logical = NewUUID()
		case "":
			logical, err = NewString(0, "")
		default:
			return LogicalType{}, false, ErrConnectorCatalogNotRepresentable
		}
	default:
		return LogicalType{}, false, ErrConnectorCatalogNotRepresentable
	}
	if err != nil {
		return LogicalType{}, false, fmt.Errorf("%w: %v", ErrConnectorCatalogNotRepresentable, err)
	}
	return logical, nullable, nil
}

func schemaTypeIncludesExactly(raw json.RawMessage, want string) bool {
	types, err := schemaTypes(raw)
	return err == nil && len(types) == 1 && types[0] == want
}

func schemaTypes(raw json.RawMessage) ([]string, error) {
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		if single == "" {
			return nil, ErrConnectorCatalogNotRepresentable
		}
		return []string{single}, nil
	}
	var multiple []string
	if err := json.Unmarshal(raw, &multiple); err != nil || len(multiple) == 0 {
		return nil, ErrConnectorCatalogNotRepresentable
	}
	seen := make(map[string]bool, len(multiple))
	for _, value := range multiple {
		if value == "" || seen[value] {
			return nil, ErrConnectorCatalogNotRepresentable
		}
		seen[value] = true
	}
	return multiple, nil
}
