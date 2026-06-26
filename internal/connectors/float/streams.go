package float

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Float API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Float list endpoint path segment (e.g. "people").
	resource string
	// mapRecord flattens a raw Float object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// floatStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in floatStreams; the read path
// is fully data-driven from this table. Float v3 list endpoints return a
// top-level JSON array (no envelope), so the records path is "" for all of them.
var floatStreamEndpoints = map[string]streamEndpoint{
	"people":      {resource: "people", mapRecord: floatPersonRecord},
	"projects":    {resource: "projects", mapRecord: floatProjectRecord},
	"clients":     {resource: "clients", mapRecord: floatClientRecord},
	"tasks":       {resource: "tasks", mapRecord: floatTaskRecord},
	"departments": {resource: "departments", mapRecord: floatDepartmentRecord},
}

// floatStreams returns the connector's published stream catalog. Each Float
// resource exposes its own "<resource>_id" integer primary key. Float v3 list
// endpoints are full-refresh only (no reliable updated_at cursor across all
// resources), matching the catalog's supported_sync_modes: [full_refresh].
func floatStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "people",
			Description: "Float people (team members).",
			PrimaryKey:  []string{"people_id"},
			Fields:      floatPersonFields(),
		},
		{
			Name:        "projects",
			Description: "Float projects.",
			PrimaryKey:  []string{"project_id"},
			Fields:      floatProjectFields(),
		},
		{
			Name:        "clients",
			Description: "Float clients.",
			PrimaryKey:  []string{"client_id"},
			Fields:      floatClientFields(),
		},
		{
			Name:        "tasks",
			Description: "Float project tasks.",
			PrimaryKey:  []string{"task_id"},
			Fields:      floatTaskFields(),
		},
		{
			Name:        "departments",
			Description: "Float departments.",
			PrimaryKey:  []string{"department_id"},
			Fields:      floatDepartmentFields(),
		},
	}
}

func floatPersonFields() []connectors.Field {
	return []connectors.Field{
		{Name: "people_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "job_title", Type: "string"},
		{Name: "role_id", Type: "integer"},
		{Name: "department_id", Type: "integer"},
		{Name: "people_type_id", Type: "integer"},
		{Name: "active", Type: "integer"},
		{Name: "employee_type", Type: "integer"},
		{Name: "start_date", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func floatProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "project_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "client_id", Type: "integer"},
		{Name: "color", Type: "string"},
		{Name: "notes", Type: "string"},
		{Name: "tags", Type: "array"},
		{Name: "budget_type", Type: "integer"},
		{Name: "budget_total", Type: "number"},
		{Name: "default_hourly_rate", Type: "number"},
		{Name: "non_billable", Type: "integer"},
		{Name: "active", Type: "integer"},
		{Name: "project_manager", Type: "integer"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func floatClientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "client_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func floatTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "task_id", Type: "integer"},
		{Name: "project_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "billable", Type: "integer"},
		{Name: "task_meta_id", Type: "integer"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func floatDepartmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "department_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "parent_id", Type: "integer"},
	}
}

func floatPersonRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"people_id":      item["people_id"],
		"name":           item["name"],
		"email":          item["email"],
		"job_title":      item["job_title"],
		"role_id":        item["role_id"],
		"department_id":  item["department_id"],
		"people_type_id": item["people_type_id"],
		"active":         item["active"],
		"employee_type":  item["employee_type"],
		"start_date":     item["start_date"],
		"created":        item["created"],
		"modified":       item["modified"],
	}
}

func floatProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"project_id":          item["project_id"],
		"name":                item["name"],
		"client_id":           item["client_id"],
		"color":               item["color"],
		"notes":               item["notes"],
		"tags":                item["tags"],
		"budget_type":         item["budget_type"],
		"budget_total":        item["budget_total"],
		"default_hourly_rate": item["default_hourly_rate"],
		"non_billable":        item["non_billable"],
		"active":              item["active"],
		"project_manager":     item["project_manager"],
		"created":             item["created"],
		"modified":            item["modified"],
	}
}

func floatClientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"client_id": item["client_id"],
		"name":      item["name"],
		"created":   item["created"],
		"modified":  item["modified"],
	}
}

func floatTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"task_id":      item["task_id"],
		"project_id":   item["project_id"],
		"name":         item["name"],
		"billable":     item["billable"],
		"task_meta_id": item["task_meta_id"],
		"created":      item["created"],
		"modified":     item["modified"],
	}
}

func floatDepartmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"department_id": item["department_id"],
		"name":          item["name"],
		"parent_id":     item["parent_id"],
	}
}
