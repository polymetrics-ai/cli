package deputy

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to its Deputy API path (relative to
// base_url) and the record mapper that flattens raw Deputy objects. paginated is
// true for endpoints that accept Deputy's ?start=N offset paging (the
// /resource/* collection endpoints); the curated my/* and supervise endpoints
// return a single bounded page.
type streamEndpoint struct {
	path      string
	paginated bool
	mapRecord func(map[string]any) connectors.Record
}

// deputyStreamEndpoints is the per-stream routing table. The read path is fully
// data-driven from this table: adding a stream means adding one entry here plus a
// Stream definition in deputyStreams.
//
// Paths and primary keys mirror the official Deputy API and the upstream
// airbyte/source-deputy manifest.
var deputyStreamEndpoints = map[string]streamEndpoint{
	"locations":   {path: "api/v1/resource/Company", paginated: true, mapRecord: deputyLocationRecord},
	"employees":   {path: "api/v1/supervise/employee", paginated: false, mapRecord: deputyEmployeeRecord},
	"departments": {path: "api/v1/resource/OperationalUnit", paginated: true, mapRecord: deputyDepartmentRecord},
	"timesheets":  {path: "api/v1/my/timesheets", paginated: false, mapRecord: deputyTimesheetRecord},
	"tasks":       {path: "api/v1/my/tasks", paginated: false, mapRecord: deputyTaskRecord},
}

// deputyStreams returns the connector's published catalog. Deputy is
// full-refresh only (no incremental cursor field), keyed by the integer Id
// flattened to "id".
func deputyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "locations",
			Description: "Deputy locations (Company resource).",
			PrimaryKey:  []string{"id"},
			Fields:      deputyLocationFields(),
		},
		{
			Name:        "employees",
			Description: "Deputy employees.",
			PrimaryKey:  []string{"id"},
			Fields:      deputyEmployeeFields(),
		},
		{
			Name:        "departments",
			Description: "Deputy departments (OperationalUnit resource).",
			PrimaryKey:  []string{"id"},
			Fields:      deputyDepartmentFields(),
		},
		{
			Name:        "timesheets",
			Description: "Deputy timesheets for the authenticated user's scope.",
			PrimaryKey:  []string{"id"},
			Fields:      deputyTimesheetFields(),
		},
		{
			Name:        "tasks",
			Description: "Deputy tasks for the authenticated user's scope.",
			PrimaryKey:  []string{"id"},
			Fields:      deputyTaskFields(),
		},
	}
}

func deputyLocationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "company_name", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "address", Type: "integer"},
		{Name: "country", Type: "integer"},
		{Name: "creator", Type: "integer"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func deputyEmployeeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "display_name", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "company", Type: "integer"},
		{Name: "role", Type: "integer"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func deputyDepartmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "operational_unit_name", Type: "string"},
		{Name: "company", Type: "integer"},
		{Name: "active", Type: "boolean"},
		{Name: "creator", Type: "integer"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func deputyTimesheetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "employee", Type: "integer"},
		{Name: "operational_unit", Type: "integer"},
		{Name: "date", Type: "string"},
		{Name: "start_time", Type: "integer"},
		{Name: "end_time", Type: "integer"},
		{Name: "total_time", Type: "number"},
		{Name: "is_in_progress", Type: "boolean"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func deputyTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "due_time", Type: "string"},
		{Name: "completed", Type: "boolean"},
		{Name: "priority", Type: "integer"},
		{Name: "creator", Type: "integer"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func deputyLocationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["Id"],
		"company_name": item["CompanyName"],
		"code":         item["Code"],
		"active":       item["Active"],
		"address":      item["Address"],
		"country":      item["Country"],
		"creator":      item["Creator"],
		"created":      item["Created"],
		"modified":     item["Modified"],
	}
}

func deputyEmployeeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["Id"],
		"display_name": item["DisplayName"],
		"first_name":   item["FirstName"],
		"last_name":    item["LastName"],
		"active":       item["Active"],
		"company":      item["Company"],
		"role":         item["Role"],
		"created":      item["Created"],
		"modified":     item["Modified"],
	}
}

func deputyDepartmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                    item["Id"],
		"operational_unit_name": item["OperationalUnitName"],
		"company":               item["Company"],
		"active":                item["Active"],
		"creator":               item["Creator"],
		"created":               item["Created"],
		"modified":              item["Modified"],
	}
}

func deputyTimesheetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["Id"],
		"employee":         item["Employee"],
		"operational_unit": item["OperationalUnit"],
		"date":             item["Date"],
		"start_time":       item["StartTime"],
		"end_time":         item["EndTime"],
		"total_time":       item["TotalTime"],
		"is_in_progress":   item["IsInProgress"],
		"created":          item["Created"],
		"modified":         item["Modified"],
	}
}

func deputyTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["Id"],
		"title":     item["Title"],
		"due_time":  item["DueTime"],
		"completed": item["Completed"],
		"priority":  item["Priority"],
		"creator":   item["Creator"],
		"created":   item["Created"],
		"modified":  item["Modified"],
	}
}
