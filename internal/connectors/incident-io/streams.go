package incidentio

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the incident.io API resource path
// (relative to base_url), the JSON key holding the records array in the
// response, whether the endpoint paginates via pagination_meta.after, and the
// record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the API path segment, e.g. "v2/incidents".
	resource string
	// recordsKey is the top-level JSON key holding the array, e.g. "incidents".
	recordsKey string
	// paginated is true when the endpoint supports page_size/after cursor
	// pagination and returns pagination_meta.after for the next page.
	paginated bool
	// mapRecord flattens a raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. The read path is fully
// data-driven from this table: adding a stream means adding one entry here plus
// a Stream definition in streams().
var streamEndpoints = map[string]streamEndpoint{
	"incidents":      {resource: "v2/incidents", recordsKey: "incidents", paginated: true, mapRecord: incidentRecord},
	"severities":     {resource: "v1/severities", recordsKey: "severities", paginated: false, mapRecord: severityRecord},
	"incident_roles": {resource: "v2/incident_roles", recordsKey: "incident_roles", paginated: false, mapRecord: incidentRoleRecord},
	"users":          {resource: "v2/users", recordsKey: "users", paginated: true, mapRecord: userRecord},
	"follow_ups":     {resource: "v2/follow_ups", recordsKey: "follow_ups", paginated: true, mapRecord: followUpRecord},
}

// streams returns the connector's published stream catalog. incident.io objects
// carry a string id primary key; objects with created_at/updated_at expose those
// as incremental cursor fields.
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "incidents",
			Description:  "incident.io incidents.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       incidentFields(),
		},
		{
			Name:         "severities",
			Description:  "incident.io severities.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       severityFields(),
		},
		{
			Name:         "incident_roles",
			Description:  "incident.io incident roles.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       incidentRoleFields(),
		},
		{
			Name:         "users",
			Description:  "incident.io users.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       userFields(),
		},
		{
			Name:         "follow_ups",
			Description:  "incident.io follow-ups.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       followUpFields(),
		},
	}
}

func incidentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "reference", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "summary", Type: "string"},
		{Name: "mode", Type: "string"},
		{Name: "visibility", Type: "string"},
		{Name: "severity_id", Type: "string"},
		{Name: "severity_name", Type: "string"},
		{Name: "status_id", Type: "string"},
		{Name: "status_name", Type: "string"},
		{Name: "status_category", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func severityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "rank", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func incidentRoleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "role_type", Type: "string"},
		{Name: "shortform", Type: "string"},
		{Name: "instructions", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "slack_user_id", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "base_role_id", Type: "string"},
		{Name: "base_role_name", Type: "string"},
	}
}

func followUpFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "incident_id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "assignee_id", Type: "string"},
		{Name: "assignee_name", Type: "string"},
		{Name: "completed_at", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func incidentRecord(item map[string]any) connectors.Record {
	severity := nestedObject(item, "severity")
	status := nestedObject(item, "incident_status")
	return connectors.Record{
		"id":              item["id"],
		"reference":       item["reference"],
		"name":            item["name"],
		"summary":         item["summary"],
		"mode":            item["mode"],
		"visibility":      item["visibility"],
		"severity_id":     severity["id"],
		"severity_name":   severity["name"],
		"status_id":       status["id"],
		"status_name":     status["name"],
		"status_category": status["category"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
	}
}

func severityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"rank":        item["rank"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func incidentRoleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"description":  item["description"],
		"role_type":    item["role_type"],
		"shortform":    item["shortform"],
		"instructions": item["instructions"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func userRecord(item map[string]any) connectors.Record {
	baseRole := nestedObject(item, "base_role")
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"email":          item["email"],
		"slack_user_id":  item["slack_user_id"],
		"role":           item["role"],
		"base_role_id":   baseRole["id"],
		"base_role_name": baseRole["name"],
	}
}

func followUpRecord(item map[string]any) connectors.Record {
	assignee := nestedObject(item, "assignee")
	return connectors.Record{
		"id":            item["id"],
		"incident_id":   item["incident_id"],
		"title":         item["title"],
		"description":   item["description"],
		"status":        item["status"],
		"assignee_id":   assignee["id"],
		"assignee_name": assignee["name"],
		"completed_at":  item["completed_at"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

// nestedObject returns item[key] as a map, or an empty map when missing/null or
// not an object. This keeps the record mappers nil-safe against optional nested
// objects in the API payload.
func nestedObject(item map[string]any, key string) map[string]any {
	if obj, ok := item[key].(map[string]any); ok {
		return obj
	}
	return map[string]any{}
}
