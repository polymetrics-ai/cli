package insightful

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a published stream name to the Insightful API resource path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects. Adding a stream means adding one entry here plus a Stream definition
// in insightfulStreams; the read path is fully data-driven from this table.
type streamEndpoint struct {
	// resource is the Insightful list endpoint path segment (e.g. "employee").
	resource string
	// query are static query params appended to every request for this stream
	// (e.g. a `select` projection). Optional.
	query map[string]string
	// mapRecord flattens a raw Insightful object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// insightfulStreamEndpoints is the per-stream routing table. The Insightful API
// uses singular resource paths (employee, team, project, directory) even though
// the published stream names follow the upstream upstream connector naming.
var insightfulStreamEndpoints = map[string]streamEndpoint{
	"employee":  {resource: "employee", mapRecord: insightfulEmployeeRecord},
	"team":      {resource: "team", mapRecord: insightfulTeamRecord},
	"projects":  {resource: "project", mapRecord: insightfulProjectRecord},
	"directory": {resource: "directory", mapRecord: insightfulDirectoryRecord},
}

// insightfulStreams returns the connector's published stream catalog. Every
// Insightful object exposes a string `id`; mutable resources carry a numeric
// `updatedAt` (unix millis) used as the incremental cursor.
func insightfulStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "employee",
			Description:  "Insightful employees (monitored users).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       insightfulEmployeeFields(),
		},
		{
			Name:        "team",
			Description: "Insightful teams and their settings.",
			PrimaryKey:  []string{"id"},
			Fields:      insightfulTeamFields(),
		},
		{
			Name:         "projects",
			Description:  "Insightful projects.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       insightfulProjectFields(),
		},
		{
			Name:         "directory",
			Description:  "Insightful directory entries (org directory members).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       insightfulDirectoryFields(),
		},
	}
}

func insightfulEmployeeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "modelName", Type: "string"},
		{Name: "createdAt", Type: "integer"},
		{Name: "updatedAt", Type: "integer"},
	}
}

func insightfulTeamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "default", Type: "boolean"},
		{Name: "employees", Type: "array"},
		{Name: "projects", Type: "array"},
		{Name: "modelName", Type: "string"},
	}
}

func insightfulProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "archived", Type: "boolean"},
		{Name: "billable", Type: "boolean"},
		{Name: "creatorId", Type: "string"},
		{Name: "organizationId", Type: "string"},
		{Name: "employees", Type: "array"},
		{Name: "modelName", Type: "string"},
		{Name: "createdAt", Type: "integer"},
		{Name: "updatedAt", Type: "integer"},
	}
}

func insightfulDirectoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "organizationId", Type: "string"},
		{Name: "modelName", Type: "string"},
		{Name: "createdAt", Type: "integer"},
		{Name: "updatedAt", Type: "integer"},
	}
}

func insightfulEmployeeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"email":     item["email"],
		"modelName": item["modelName"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
	}
}

func insightfulTeamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"default":     item["default"],
		"employees":   item["employees"],
		"projects":    item["projects"],
		"modelName":   item["modelName"],
	}
}

func insightfulProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"archived":       item["archived"],
		"billable":       item["billable"],
		"creatorId":      item["creatorId"],
		"organizationId": item["organizationId"],
		"employees":      item["employees"],
		"modelName":      item["modelName"],
		"createdAt":      item["createdAt"],
		"updatedAt":      item["updatedAt"],
	}
}

func insightfulDirectoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"organizationId": item["organizationId"],
		"modelName":      item["modelName"],
		"createdAt":      item["createdAt"],
		"updatedAt":      item["updatedAt"],
	}
}
