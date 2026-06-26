package hibob

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the HiBob API resource path (relative to
// base_url), the JSON path to its records array in the response body, and the
// record mapper that flattens its objects. The read path is fully data-driven
// from this table.
type streamEndpoint struct {
	// resource is the HiBob list endpoint path segment (e.g. "profiles").
	resource string
	// recordsPath is the dotted JSON path to the array of records in the
	// response (e.g. "employees", "values").
	recordsPath string
	// mapRecord flattens a raw HiBob object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// paginated is true when the endpoint supports offset/limit paging. HiBob's
	// metadata endpoints (named lists, fields) return the full set in one shot.
	paginated bool
}

// hibobStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in hibobStreams.
var hibobStreamEndpoints = map[string]streamEndpoint{
	"profiles":      {resource: "profiles", recordsPath: "employees", mapRecord: hibobProfileRecord, paginated: true},
	"named_lists":   {resource: "company/named-lists", recordsPath: "values", mapRecord: hibobNamedListRecord},
	"company_lists": {resource: "company/people/fields", recordsPath: ".", mapRecord: hibobFieldRecord},
}

// hibobStreams returns the connector's published stream catalog.
func hibobStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "profiles",
			Description:  "Active employee profiles the service user can access.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       hibobProfileFields(),
		},
		{
			Name:         "named_lists",
			Description:  "Company named lists (e.g. departments, sites) and their values.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       hibobNamedListFields(),
		},
		{
			Name:         "company_lists",
			Description:  "Company people field definitions (HR metadata schema).",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       hibobFieldFields(),
		},
	}
}

func hibobProfileFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "displayName", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "surname", Type: "string"},
		{Name: "fullName", Type: "string"},
		{Name: "personal_pronouns", Type: "string"},
		{Name: "work_title", Type: "string"},
		{Name: "work_department", Type: "string"},
		{Name: "work_site", Type: "string"},
		{Name: "work_startDate", Type: "string"},
		{Name: "work_isManager", Type: "boolean"},
	}
}

func hibobNamedListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "value", Type: "string"},
		{Name: "parentId", Type: "string"},
		{Name: "archived", Type: "boolean"},
		{Name: "children", Type: "object"},
	}
}

func hibobFieldFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "description", Type: "string"},
	}
}

// hibobProfileRecord flattens an employee profile, hoisting the most useful
// nested work fields into top-level work_* keys while preserving raw nested
// objects.
func hibobProfileRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":                nestedString(item, "id"),
		"email":             item["email"],
		"displayName":       item["displayName"],
		"firstName":         item["firstName"],
		"surname":           item["surname"],
		"fullName":          item["fullName"],
		"personal_pronouns": nestedString(mapAt(item, "personal"), "pronouns"),
	}
	work := mapAt(item, "work")
	rec["work_title"] = work["title"]
	rec["work_department"] = work["department"]
	rec["work_site"] = work["site"]
	rec["work_startDate"] = work["startDate"]
	rec["work_isManager"] = work["isManager"]
	rec["work"] = item["work"]
	rec["personal"] = item["personal"]
	return rec
}

// hibobNamedListRecord flattens a named-list value entry.
func hibobNamedListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       nestedString(item, "id"),
		"name":     item["name"],
		"value":    item["value"],
		"parentId": item["parentId"],
		"archived": item["archived"],
		"children": item["children"],
	}
}

// hibobFieldRecord flattens a people field definition.
func hibobFieldRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          nestedString(item, "id"),
		"name":        item["name"],
		"category":    item["category"],
		"type":        item["type"],
		"description": item["description"],
	}
}
