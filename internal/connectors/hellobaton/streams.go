package hellobaton

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Hellobaton API resource path
// (relative to base_url) and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Hellobaton list endpoint path segment (e.g. "projects").
	resource string
	// mapRecord flattens a raw Hellobaton object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// hellobatonStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in hellobatonStreams; the
// read path is fully data-driven from this table.
var hellobatonStreamEndpoints = map[string]streamEndpoint{
	"projects":   {resource: "projects", mapRecord: hellobatonProjectRecord},
	"milestones": {resource: "milestones", mapRecord: hellobatonMilestoneRecord},
	"tasks":      {resource: "tasks", mapRecord: hellobatonTaskRecord},
	"phases":     {resource: "phases", mapRecord: hellobatonGenericRecord},
	"companies":  {resource: "companies", mapRecord: hellobatonGenericRecord},
	"users":      {resource: "users", mapRecord: hellobatonUserRecord},
}

// hellobatonStreams returns the connector's published stream catalog. Every
// Hellobaton object exposes an integer id and ISO-8601 `created`/`modified`
// timestamps; the API only supports full refresh, so cursor fields are advisory.
func hellobatonStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "projects",
			Description:  "Hellobaton projects.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       hellobatonProjectFields(),
		},
		{
			Name:         "milestones",
			Description:  "Hellobaton project milestones.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       hellobatonMilestoneFields(),
		},
		{
			Name:         "tasks",
			Description:  "Hellobaton tasks.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       hellobatonTaskFields(),
		},
		{
			Name:         "phases",
			Description:  "Hellobaton project phases.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       hellobatonGenericFields(),
		},
		{
			Name:         "companies",
			Description:  "Hellobaton companies.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       hellobatonGenericFields(),
		},
		{
			Name:         "users",
			Description:  "Hellobaton users.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       hellobatonUserFields(),
		},
	}
}

func hellobatonGenericFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "created", Type: "timestamp"},
		{Name: "modified", Type: "timestamp"},
		{Name: "_self", Type: "string"},
	}
}

func hellobatonProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "created", Type: "timestamp"},
		{Name: "modified", Type: "timestamp"},
		{Name: "archived", Type: "boolean"},
		{Name: "completed_datetime", Type: "timestamp"},
		{Name: "cost", Type: "integer"},
		{Name: "annual_contract_value", Type: "string"},
		{Name: "creator", Type: "string"},
		{Name: "_self", Type: "string"},
	}
}

func hellobatonMilestoneFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "created", Type: "timestamp"},
		{Name: "modified", Type: "timestamp"},
		{Name: "description", Type: "string"},
		{Name: "deadline_datetime", Type: "timestamp"},
		{Name: "deadline_fixed", Type: "boolean"},
		{Name: "duration", Type: "integer"},
		{Name: "finish_datetime", Type: "timestamp"},
		{Name: "project", Type: "string"},
		{Name: "_self", Type: "string"},
	}
}

func hellobatonTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "created", Type: "timestamp"},
		{Name: "modified", Type: "timestamp"},
		{Name: "description", Type: "string"},
		{Name: "project", Type: "string"},
		{Name: "_self", Type: "string"},
	}
}

func hellobatonUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "created", Type: "timestamp"},
		{Name: "modified", Type: "timestamp"},
		{Name: "_self", Type: "string"},
	}
}

func hellobatonGenericRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"name":     item["name"],
		"created":  item["created"],
		"modified": item["modified"],
		"_self":    item["_self"],
	}
}

func hellobatonProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                    item["id"],
		"name":                  item["name"],
		"created":               item["created"],
		"modified":              item["modified"],
		"archived":              item["archived"],
		"completed_datetime":    item["completed_datetime"],
		"cost":                  item["cost"],
		"annual_contract_value": item["annual_contract_value"],
		"creator":               item["creator"],
		"_self":                 item["_self"],
	}
}

func hellobatonMilestoneRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"created":           item["created"],
		"modified":          item["modified"],
		"description":       item["description"],
		"deadline_datetime": item["deadline_datetime"],
		"deadline_fixed":    item["deadline_fixed"],
		"duration":          item["duration"],
		"finish_datetime":   item["finish_datetime"],
		"project":           item["project"],
		"_self":             item["_self"],
	}
}

func hellobatonTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"created":     item["created"],
		"modified":    item["modified"],
		"description": item["description"],
		"project":     item["project"],
		"_self":       item["_self"],
	}
}

func hellobatonUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"first_name": item["first_name"],
		"last_name":  item["last_name"],
		"created":    item["created"],
		"modified":   item["modified"],
		"_self":      item["_self"],
	}
}
