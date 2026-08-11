package postgres

import (
	"strconv"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
)

// postgresColumnType maps only closed, lossless PostgreSQL catalog shapes.
// All other types deliberately return ErrUnsupportedCatalogShape rather than
// masquerading as generic strings or objects.
func postgresColumnType(definition database.Definition, column postgresCatalogColumn) (database.NativeType, database.LogicalType, error) {
	if column.typeKind != "b" || column.elementOID != 0 {
		return database.NativeType{}, database.LogicalType{}, ErrUnsupportedCatalogShape
	}

	switch column.nativeName {
	case "text":
		logical, err := database.NewString(0, column.collation)
		if err != nil {
			return database.NativeType{}, database.LogicalType{}, ErrUnsupportedCatalogShape
		}
		return database.NativeType{Name: column.nativeName}, logical, nil
	case "varchar", "bpchar":
		length, modifiers, err := postgresStringLength(column.typeModifier)
		if err != nil {
			return database.NativeType{}, database.LogicalType{}, ErrUnsupportedCatalogShape
		}
		logical, err := database.NewString(length, column.collation)
		if err != nil {
			return database.NativeType{}, database.LogicalType{}, ErrUnsupportedCatalogShape
		}
		return database.NativeType{Name: column.nativeName, Modifiers: modifiers}, logical, nil
	case "numeric":
		precision, scale, err := postgresNumericPrecisionScale(column.typeModifier)
		if err != nil {
			return database.NativeType{}, database.LogicalType{}, ErrUnsupportedCatalogShape
		}
		logical, err := database.NewDecimal(precision, scale)
		if err != nil {
			return database.NativeType{}, database.LogicalType{}, ErrUnsupportedCatalogShape
		}
		return database.NativeType{
			Name:      column.nativeName,
			Modifiers: []string{"precision-" + strconv.Itoa(int(precision)), "scale-" + strconv.Itoa(int(scale))},
		}, logical, nil
	case "time", "timetz", "timestamp", "timestamptz":
		precision, modifiers, err := postgresTemporalPrecision(column.typeModifier)
		if err != nil {
			return database.NativeType{}, database.LogicalType{}, ErrUnsupportedCatalogShape
		}
		var logical database.LogicalType
		if column.nativeName == "time" || column.nativeName == "timetz" {
			logical, err = database.NewTime(precision, column.nativeName == "timetz")
		} else {
			logical, err = database.NewTimestamp(precision, column.nativeName == "timestamptz")
		}
		if err != nil {
			return database.NativeType{}, database.LogicalType{}, ErrUnsupportedCatalogShape
		}
		return database.NativeType{Name: column.nativeName, Modifiers: modifiers}, logical, nil
	}

	if column.typeModifier != -1 {
		return database.NativeType{}, database.LogicalType{}, ErrUnsupportedCatalogShape
	}
	for _, mapping := range definition.TypeMappings() {
		if mapping.Native.Name == column.nativeName && len(mapping.Native.Modifiers) == 0 {
			return database.NativeType{Name: column.nativeName}, mapping.Logical, nil
		}
	}
	return database.NativeType{}, database.LogicalType{}, ErrUnsupportedCatalogShape
}

func postgresStringLength(typeModifier int32) (uint32, []string, error) {
	if typeModifier == -1 {
		return 0, nil, nil
	}
	if typeModifier < 4 {
		return 0, nil, ErrUnsupportedCatalogShape
	}
	length := typeModifier - 4
	if length < 0 {
		return 0, nil, ErrUnsupportedCatalogShape
	}
	return uint32(length), []string{"length-" + strconv.Itoa(int(length))}, nil
}

func postgresNumericPrecisionScale(typeModifier int32) (uint16, uint16, error) {
	if typeModifier < 4 {
		return 0, 0, ErrUnsupportedCatalogShape
	}
	modifier := int64(typeModifier) - 4
	precision := uint16((modifier >> 16) & 0xffff)
	scale := int16(modifier & 0xffff)
	if precision == 0 || scale < 0 || uint16(scale) > precision {
		return 0, 0, ErrUnsupportedCatalogShape
	}
	return precision, uint16(scale), nil
}

func postgresTemporalPrecision(typeModifier int32) (uint16, []string, error) {
	if typeModifier == -1 {
		return 6, nil, nil
	}
	if typeModifier < 0 || typeModifier > 9 {
		return 0, nil, ErrUnsupportedCatalogShape
	}
	precision := uint16(typeModifier)
	return precision, []string{"precision-" + strconv.Itoa(int(precision))}, nil
}

// legacyStreamsFromTypedCatalog is the one-way compatibility projection for
// callers that still consume connectors.Catalog. It never discovers or
// retypes a PostgreSQL table on its own.
func legacyStreamsFromTypedCatalog(catalog database.Catalog) []connectors.Stream {
	relations := catalog.Relations()
	streams := make([]connectors.Stream, 0, len(relations))
	for _, relation := range relations {
		qualified := relation.Ref.Schema.Name + "." + relation.Ref.Name
		stream := connectors.Stream{
			Name:        qualified,
			Description: "PostgreSQL table " + qualified,
			Fields:      make([]connectors.Field, 0, len(relation.Columns)),
		}
		for _, column := range relation.Columns {
			stream.Fields = append(stream.Fields, connectors.Field{
				Name: column.Ref.Name,
				Type: legacyFieldType(column.Type),
			})
		}
		for _, key := range relation.Keys {
			if key.Kind != database.KeyPrimary {
				continue
			}
			stream.PrimaryKey = make([]string, 0, len(key.Columns))
			for _, column := range key.Columns {
				stream.PrimaryKey = append(stream.PrimaryKey, column.Name)
			}
			break
		}
		streams = append(streams, stream)
	}
	return streams
}

func legacyFieldType(logical database.LogicalType) string {
	switch logical.Kind() {
	case database.LogicalSignedInteger, database.LogicalUnsignedInteger:
		return "integer"
	case database.LogicalDecimal, database.LogicalFloat:
		return "number"
	case database.LogicalBoolean:
		return "boolean"
	case database.LogicalDate, database.LogicalTime, database.LogicalTimestamp:
		return "timestamp"
	case database.LogicalJSON:
		return "object"
	case database.LogicalArray:
		return "array"
	case database.LogicalString, database.LogicalBinary, database.LogicalUUID:
		return "string"
	default:
		return "unsupported"
	}
}
