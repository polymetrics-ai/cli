package glassfrog

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the GlassFrog API resource path (relative
// to base_url) it reads from, the JSON key its records array is nested under
// (GlassFrog wraps each list under a key named after the resource, e.g.
// {"circles":[...]}), and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the GlassFrog list endpoint path segment (e.g. "circles").
	resource string
	// recordsKey is the dotted path to the records array in the response body.
	recordsKey string
	// mapRecord flattens a raw GlassFrog object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// glassfrogStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in glassfrogStreams; the read
// path is fully data-driven from this table.
var glassfrogStreamEndpoints = map[string]streamEndpoint{
	"assignments": {resource: "assignments", recordsKey: "assignments", mapRecord: assignmentRecord},
	"circles":     {resource: "circles", recordsKey: "circles", mapRecord: circleRecord},
	"people":      {resource: "people", recordsKey: "people", mapRecord: personRecord},
	"projects":    {resource: "projects", recordsKey: "projects", mapRecord: projectRecord},
	"roles":       {resource: "roles", recordsKey: "roles", mapRecord: roleRecord},
}

// glassfrogStreams returns the connector's published stream catalog. Every
// GlassFrog object exposes an integer id as primary key. The API supports only
// full-refresh syncs, so no incremental cursor fields are declared.
func glassfrogStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "assignments",
			Description: "GlassFrog role assignments (people filling roles).",
			PrimaryKey:  []string{"id"},
			Fields:      assignmentFields(),
		},
		{
			Name:        "circles",
			Description: "GlassFrog circles (organizational units).",
			PrimaryKey:  []string{"id"},
			Fields:      circleFields(),
		},
		{
			Name:        "people",
			Description: "GlassFrog people (members of the organization).",
			PrimaryKey:  []string{"id"},
			Fields:      personFields(),
		},
		{
			Name:        "projects",
			Description: "GlassFrog projects tracked within circles.",
			PrimaryKey:  []string{"id"},
			Fields:      projectFields(),
		},
		{
			Name:        "roles",
			Description: "GlassFrog roles defined within circles.",
			PrimaryKey:  []string{"id"},
			Fields:      roleFields(),
		},
	}
}

func assignmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "election", Type: "string"},
		{Name: "exclude_from_meetings", Type: "boolean"},
		{Name: "focus", Type: "string"},
		{Name: "person_id", Type: "integer"},
		{Name: "role_id", Type: "integer"},
	}
}

func circleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "short_name", Type: "string"},
		{Name: "organization_id", Type: "integer"},
		{Name: "strategy", Type: "string"},
		{Name: "supported_role_id", Type: "integer"},
	}
}

func personFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "external_id", Type: "string"},
		{Name: "tag_names", Type: "array"},
	}
}

func projectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "description", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "effort", Type: "string"},
		{Name: "value", Type: "string"},
		{Name: "roi", Type: "string"},
		{Name: "private_to_circle", Type: "boolean"},
		{Name: "link", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "archived_at", Type: "string"},
		{Name: "waiting_on_who", Type: "string"},
		{Name: "waiting_on_what", Type: "string"},
	}
}

func roleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "purpose", Type: "string"},
		{Name: "is_core", Type: "boolean"},
		{Name: "elected_until", Type: "string"},
		{Name: "name_with_circle_for_core_roles", Type: "string"},
		{Name: "organization_id", Type: "integer"},
	}
}

func assignmentRecord(item map[string]any) connectors.Record {
	links := nestedObject(item, "links")
	return connectors.Record{
		"id":                    item["id"],
		"election":              item["election"],
		"exclude_from_meetings": item["exclude_from_meetings"],
		"focus":                 item["focus"],
		"person_id":             links["person"],
		"role_id":               links["role"],
	}
}

func circleRecord(item map[string]any) connectors.Record {
	links := nestedObject(item, "links")
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"short_name":        item["short_name"],
		"organization_id":   item["organization_id"],
		"strategy":          item["strategy"],
		"supported_role_id": links["supported_role"],
	}
}

func personRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"email":       item["email"],
		"external_id": item["external_id"],
		"tag_names":   item["tag_names"],
	}
}

func projectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"description":       item["description"],
		"status":            item["status"],
		"effort":            item["effort"],
		"value":             item["value"],
		"roi":               item["roi"],
		"private_to_circle": item["private_to_circle"],
		"link":              item["link"],
		"created_at":        item["created_at"],
		"archived_at":       item["archived_at"],
		"waiting_on_who":    item["waiting_on_who"],
		"waiting_on_what":   item["waiting_on_what"],
	}
}

func roleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                              item["id"],
		"name":                            item["name"],
		"purpose":                         item["purpose"],
		"is_core":                         item["is_core"],
		"elected_until":                   item["elected_until"],
		"name_with_circle_for_core_roles": item["name_with_circle_for_core_roles"],
		"organization_id":                 item["organization_id"],
	}
}

// nestedObject returns item[key] as a map, or an empty map when absent. GlassFrog
// nests foreign keys under a "links" object.
func nestedObject(item map[string]any, key string) map[string]any {
	if obj, ok := item[key].(map[string]any); ok {
		return obj
	}
	return map[string]any{}
}
