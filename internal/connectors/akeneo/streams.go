package akeneo

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Akeneo REST resource path (relative to
// the /api/rest/v1 base) it reads from, and the record mapper that flattens its
// items. Adding a stream means adding one entry to akeneoStreamEndpoints plus a
// Stream definition in akeneoStreams; the read path is fully data-driven.
type streamEndpoint struct {
	// resource is the Akeneo REST resource path segment (e.g. "products").
	resource string
	// mapRecord flattens a raw Akeneo item into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// akeneoStreamEndpoints is the per-stream routing table.
var akeneoStreamEndpoints = map[string]streamEndpoint{
	"products":   {resource: "products", mapRecord: akeneoProductRecord},
	"categories": {resource: "categories", mapRecord: akeneoCategoryRecord},
	"families":   {resource: "families", mapRecord: akeneoFamilyRecord},
	"attributes": {resource: "attributes", mapRecord: akeneoAttributeRecord},
	"channels":   {resource: "channels", mapRecord: akeneoChannelRecord},
}

// akeneoStreams returns the connector's published stream catalog. Akeneo PIM
// resources are keyed by a string code/identifier and the Akeneo API only
// supports full_refresh syncs, so there are no incremental cursor fields.
func akeneoStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "products",
			Description: "Akeneo catalog products.",
			PrimaryKey:  []string{"id"},
			Fields:      akeneoProductFields(),
		},
		{
			Name:        "categories",
			Description: "Akeneo product categories.",
			PrimaryKey:  []string{"id"},
			Fields:      akeneoCategoryFields(),
		},
		{
			Name:        "families",
			Description: "Akeneo product families.",
			PrimaryKey:  []string{"id"},
			Fields:      akeneoFamilyFields(),
		},
		{
			Name:        "attributes",
			Description: "Akeneo product attributes.",
			PrimaryKey:  []string{"id"},
			Fields:      akeneoAttributeFields(),
		},
		{
			Name:        "channels",
			Description: "Akeneo distribution channels.",
			PrimaryKey:  []string{"id"},
			Fields:      akeneoChannelFields(),
		},
	}
}

func akeneoProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "uuid", Type: "string"},
		{Name: "enabled", Type: "boolean"},
		{Name: "family", Type: "string"},
		{Name: "parent", Type: "string"},
		{Name: "categories", Type: "array"},
		{Name: "groups", Type: "array"},
		{Name: "values", Type: "object"},
		{Name: "created", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func akeneoCategoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "parent", Type: "string"},
		{Name: "labels", Type: "object"},
		{Name: "updated", Type: "string"},
	}
}

func akeneoFamilyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "attribute_as_label", Type: "string"},
		{Name: "attribute_as_image", Type: "string"},
		{Name: "attributes", Type: "array"},
		{Name: "labels", Type: "object"},
	}
}

func akeneoAttributeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "group", Type: "string"},
		{Name: "localizable", Type: "boolean"},
		{Name: "scopable", Type: "boolean"},
		{Name: "labels", Type: "object"},
	}
}

func akeneoChannelFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "currencies", Type: "array"},
		{Name: "locales", Type: "array"},
		{Name: "category_tree", Type: "string"},
		{Name: "labels", Type: "object"},
	}
}

// akeneoCode returns the natural string key of an Akeneo item, preferring
// "code" (categories, families, attributes, channels) and falling back to
// "identifier" then "uuid" (products). The chosen value is published as "id".
func akeneoCode(item map[string]any) any {
	for _, key := range []string{"code", "identifier", "uuid"} {
		if v, ok := item[key]; ok && v != nil && v != "" {
			return v
		}
	}
	return nil
}

func akeneoProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         akeneoCode(item),
		"uuid":       item["uuid"],
		"enabled":    item["enabled"],
		"family":     item["family"],
		"parent":     item["parent"],
		"categories": item["categories"],
		"groups":     item["groups"],
		"values":     item["values"],
		"created":    item["created"],
		"updated":    item["updated"],
	}
}

func akeneoCategoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":      akeneoCode(item),
		"parent":  item["parent"],
		"labels":  item["labels"],
		"updated": item["updated"],
	}
}

func akeneoFamilyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 akeneoCode(item),
		"attribute_as_label": item["attribute_as_label"],
		"attribute_as_image": item["attribute_as_image"],
		"attributes":         item["attributes"],
		"labels":             item["labels"],
	}
}

func akeneoAttributeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          akeneoCode(item),
		"type":        item["type"],
		"group":       item["group"],
		"localizable": item["localizable"],
		"scopable":    item["scopable"],
		"labels":      item["labels"],
	}
}

func akeneoChannelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            akeneoCode(item),
		"currencies":    item["currencies"],
		"locales":       item["locales"],
		"category_tree": item["category_tree"],
		"labels":        item["labels"],
	}
}
