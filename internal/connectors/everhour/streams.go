package everhour

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to its Everhour API resource. Top-level
// streams (projects, clients, users, team time) read a single array endpoint.
// Substreams (tasks) are parented by projects: for each project id the read
// path is templated and the parent id is stitched onto every child record.
type streamEndpoint struct {
	// resource is the API path for top-level streams (e.g. "projects").
	resource string
	// parentResource, when set, marks this as a substream. The read loop first
	// lists parentResource, then requests childPath(parentID) per parent.
	parentResource string
	// childPathFor builds the per-parent child path (e.g. "projects/<id>/tasks").
	childPathFor func(parentID string) string
	// parentIDField is the field on the child record that carries the parent id.
	parentIDField string
	// mapRecord flattens a raw Everhour object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// everhourStreamEndpoints is the per-stream routing table. The read path is
// fully data-driven from this map.
var everhourStreamEndpoints = map[string]streamEndpoint{
	"projects": {resource: "projects", mapRecord: everhourProjectRecord},
	"clients":  {resource: "clients", mapRecord: everhourClientRecord},
	"users":    {resource: "team/users", mapRecord: everhourUserRecord},
	"time":     {resource: "team/time", mapRecord: everhourTimeRecord},
	"tasks": {
		parentResource: "projects",
		childPathFor:   func(id string) string { return "projects/" + id + "/tasks" },
		parentIDField:  "project_id",
		mapRecord:      everhourTaskRecord,
	},
}

// everhourStreams returns the connector's published stream catalog. Every
// Everhour object exposes a string id; the Everhour API is full-refresh only, so
// no cursor fields are published.
func everhourStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "projects",
			Description: "Everhour projects.",
			PrimaryKey:  []string{"id"},
			Fields:      everhourProjectFields(),
		},
		{
			Name:        "clients",
			Description: "Everhour clients.",
			PrimaryKey:  []string{"id"},
			Fields:      everhourClientFields(),
		},
		{
			Name:        "users",
			Description: "Everhour team members.",
			PrimaryKey:  []string{"id"},
			Fields:      everhourUserFields(),
		},
		{
			Name:        "tasks",
			Description: "Everhour tasks, one per project.",
			PrimaryKey:  []string{"id"},
			Fields:      everhourTaskFields(),
		},
		{
			Name:        "time",
			Description: "Everhour team time records.",
			PrimaryKey:  []string{"id"},
			Fields:      everhourTimeFields(),
		},
	}
}

func everhourProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "platform", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "workspaceId", Type: "string"},
		{Name: "workspaceName", Type: "string"},
		{Name: "favorite", Type: "boolean"},
		{Name: "foreign", Type: "boolean"},
		{Name: "createdAt", Type: "string"},
	}
}

func everhourClientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "favorite", Type: "boolean"},
		{Name: "createdAt", Type: "string"},
	}
}

func everhourUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "headline", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "capacity", Type: "integer"},
		{Name: "isEmailVerified", Type: "boolean"},
		{Name: "createdAt", Type: "string"},
	}
}

func everhourTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "project_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "completed", Type: "boolean"},
		{Name: "url", Type: "string"},
		{Name: "createdAt", Type: "string"},
	}
}

func everhourTimeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "time", Type: "integer"},
		{Name: "user", Type: "integer"},
		{Name: "createdAt", Type: "string"},
	}
}

func everhourProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            stringField(item, "id"),
		"name":          item["name"],
		"type":          item["type"],
		"platform":      item["platform"],
		"status":        item["status"],
		"workspaceId":   item["workspaceId"],
		"workspaceName": item["workspaceName"],
		"favorite":      item["favorite"],
		"foreign":       item["foreign"],
		"createdAt":     item["createdAt"],
	}
}

func everhourClientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        stringField(item, "id"),
		"name":      item["name"],
		"email":     item["email"],
		"status":    item["status"],
		"favorite":  item["favorite"],
		"createdAt": item["createdAt"],
	}
}

func everhourUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              stringField(item, "id"),
		"name":            item["name"],
		"email":           item["email"],
		"headline":        item["headline"],
		"role":            item["role"],
		"status":          item["status"],
		"type":            item["type"],
		"capacity":        item["capacity"],
		"isEmailVerified": item["isEmailVerified"],
		"createdAt":       item["createdAt"],
	}
}

func everhourTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        stringField(item, "id"),
		"name":      item["name"],
		"type":      item["type"],
		"status":    item["status"],
		"completed": item["completed"],
		"url":       item["url"],
		"createdAt": item["createdAt"],
	}
}

func everhourTimeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        stringField(item, "id"),
		"date":      item["date"],
		"time":      item["time"],
		"user":      item["user"],
		"createdAt": item["createdAt"],
	}
}
