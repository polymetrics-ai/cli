package encharge

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Encharge API resource path (relative
// to base_url), the JSON path where its records array lives, whether it is
// offset-paginated, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Encharge path segment (e.g. "people/all").
	resource string
	// recordsPath is the dotted JSON path to the records array in the response
	// (e.g. "people", "segments", "objects").
	recordsPath string
	// paginated is true for endpoints that support limit/offset pagination.
	paginated bool
	// mapRecord flattens a raw Encharge object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// enchargeStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in enchargeStreams; the read
// path is fully data-driven from this table.
//
// Endpoints, record paths, and pagination come from the Encharge API and the
// Airbyte source-encharge manifest:
//   - peoples       GET /people/all     records at "people"   offset paginated
//   - segments      GET /segments       records at "segments"
//   - fields        GET /fields         records at "items"
//   - account_tags  GET /tags-management records at "tags"
//   - schemas       GET /schemas        records at "objects"
var enchargeStreamEndpoints = map[string]streamEndpoint{
	"peoples":      {resource: "people/all", recordsPath: "people", paginated: true, mapRecord: enchargePersonRecord},
	"segments":     {resource: "segments", recordsPath: "segments", mapRecord: enchargeSegmentRecord},
	"fields":       {resource: "fields", recordsPath: "items", mapRecord: enchargeFieldRecord},
	"account_tags": {resource: "tags-management", recordsPath: "tags", mapRecord: enchargeTagRecord},
	"schemas":      {resource: "schemas", recordsPath: "objects", mapRecord: enchargeSchemaRecord},
}

// enchargeStreams returns the connector's published stream catalog.
func enchargeStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "peoples",
			Description: "Encharge people (contacts).",
			PrimaryKey:  []string{"id"},
			Fields:      enchargePersonFields(),
		},
		{
			Name:        "segments",
			Description: "Encharge people segments.",
			PrimaryKey:  []string{"id"},
			Fields:      enchargeSegmentFields(),
		},
		{
			Name:        "fields",
			Description: "Encharge custom field definitions.",
			PrimaryKey:  []string{"name"},
			Fields:      enchargeFieldFields(),
		},
		{
			Name:        "account_tags",
			Description: "Encharge account-level tags.",
			PrimaryKey:  []string{"tag"},
			Fields:      enchargeTagFields(),
		},
		{
			Name:        "schemas",
			Description: "Encharge object schemas.",
			PrimaryKey:  []string{"name"},
			Fields:      enchargeSchemaFields(),
		},
	}
}

func enchargePersonFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "company", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
		{Name: "userId", Type: "string"},
	}
}

func enchargeSegmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func enchargeFieldFields() []connectors.Field {
	return []connectors.Field{
		{Name: "name", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "format", Type: "string"},
	}
}

func enchargeTagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "tag", Type: "string"},
		{Name: "id", Type: "string"},
		{Name: "createdAt", Type: "string"},
	}
}

func enchargeSchemaFields() []connectors.Field {
	return []connectors.Field{
		{Name: "name", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "type", Type: "string"},
	}
}

func enchargePersonRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"email":     item["email"],
		"name":      item["name"],
		"firstName": item["firstName"],
		"lastName":  item["lastName"],
		"phone":     item["phone"],
		"title":     item["title"],
		"company":   item["company"],
		"country":   item["country"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
		"userId":    item["userId"],
	}
}

func enchargeSegmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"type":      item["type"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
	}
}

func enchargeFieldRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"name":   item["name"],
		"title":  item["title"],
		"type":   item["type"],
		"format": item["format"],
	}
}

func enchargeTagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"tag":       item["tag"],
		"id":        item["id"],
		"createdAt": item["createdAt"],
	}
}

func enchargeSchemaRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"name":  item["name"],
		"title": item["title"],
		"type":  item["type"],
	}
}
