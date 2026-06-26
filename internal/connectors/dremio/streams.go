package dremio

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Dremio REST API resource path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects. Every listed Dremio endpoint returns a {"data":[...]} envelope and
// (optionally) a top-level "nextPageToken" for body-cursor pagination.
type streamEndpoint struct {
	// resource is the Dremio API path segment (e.g. "catalog").
	resource string
	// mapRecord flattens a raw Dremio object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// dremioStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in dremioStreams; the read path
// is fully data-driven from this table.
var dremioStreamEndpoints = map[string]streamEndpoint{
	"catalog":     {resource: "catalog", mapRecord: dremioCatalogRecord},
	"reflections": {resource: "reflections", mapRecord: dremioReflectionRecord},
	"sources":     {resource: "source", mapRecord: dremioSourceRecord},
	"users":       {resource: "user", mapRecord: dremioUserRecord},
}

// dremioStreams returns the connector's published stream catalog. Dremio objects
// are keyed by a string "id"; none of these list endpoints expose a reliable
// incremental cursor, so they are full-refresh only (no CursorFields).
func dremioStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "catalog",
			Description: "Top-level Dremio catalog entries (sources, spaces, folders, datasets).",
			PrimaryKey:  []string{"id"},
			Fields:      dremioCatalogFields(),
		},
		{
			Name:        "reflections",
			Description: "Dremio reflections (aggregation and raw acceleration definitions).",
			PrimaryKey:  []string{"id"},
			Fields:      dremioReflectionFields(),
		},
		{
			Name:        "sources",
			Description: "Dremio data sources configured on the instance.",
			PrimaryKey:  []string{"id"},
			Fields:      dremioSourceFields(),
		},
		{
			Name:        "users",
			Description: "Dremio user accounts.",
			PrimaryKey:  []string{"id"},
			Fields:      dremioUserFields(),
		},
	}
}

func dremioCatalogFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "containerType", Type: "string"},
		{Name: "datasetType", Type: "string"},
		{Name: "path", Type: "array"},
		{Name: "tag", Type: "string"},
		{Name: "createdAt", Type: "string"},
	}
}

func dremioReflectionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "datasetId", Type: "string"},
		{Name: "enabled", Type: "boolean"},
		{Name: "status", Type: "object"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func dremioSourceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "path", Type: "array"},
		{Name: "tag", Type: "string"},
		{Name: "createdAt", Type: "string"},
	}
}

func dremioUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "active", Type: "boolean"},
	}
}

func dremioCatalogRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"type":          item["type"],
		"containerType": item["containerType"],
		"datasetType":   item["datasetType"],
		"path":          item["path"],
		"tag":           item["tag"],
		"createdAt":     item["createdAt"],
	}
}

func dremioReflectionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"type":      item["type"],
		"datasetId": item["datasetId"],
		"enabled":   item["enabled"],
		"status":    item["status"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
	}
}

func dremioSourceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"type":      item["type"],
		"path":      item["path"],
		"tag":       item["tag"],
		"createdAt": item["createdAt"],
	}
}

func dremioUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"firstName": item["firstName"],
		"lastName":  item["lastName"],
		"email":     item["email"],
		"active":    item["active"],
	}
}
