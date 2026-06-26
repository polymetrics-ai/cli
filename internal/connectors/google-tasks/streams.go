package googletasks

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to how it is read. tasklists is a single
// top-level list; tasks is nested under each task list, so it carries a flag the
// read path uses to fan out across every task list.
type streamEndpoint struct {
	// resource is the path segment for top-level lists (relative to base_url),
	// e.g. "users/@me/lists". Empty for nested streams routed per parent.
	resource string
	// nested is true for streams (tasks) that are read per task list rather than
	// from a single endpoint.
	nested bool
	// mapRecord flattens a raw Google Tasks object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// googleTasksStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in googleTasksStreams.
var googleTasksStreamEndpoints = map[string]streamEndpoint{
	"tasklists": {resource: "users/@me/lists", mapRecord: taskListRecord},
	"tasks":     {nested: true, mapRecord: taskRecord},
}

// googleTasksStreams returns the connector's published stream catalog. Every
// Google Tasks resource exposes a string id and an RFC3339 `updated` timestamp,
// so the primary key is ["id"] and the incremental cursor is ["updated"].
func googleTasksStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "tasklists",
			Description:  "Google task lists belonging to the authenticated user.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       taskListFields(),
		},
		{
			Name:         "tasks",
			Description:  "Tasks across every task list belonging to the authenticated user.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       taskFields(),
		},
	}
}

func taskListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "kind", Type: "string"},
		{Name: "id", Type: "string"},
		{Name: "etag", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "updated", Type: "string"},
		{Name: "self_link", Type: "string"},
	}
}

func taskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "kind", Type: "string"},
		{Name: "id", Type: "string"},
		{Name: "etag", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "updated", Type: "string"},
		{Name: "self_link", Type: "string"},
		{Name: "parent", Type: "string"},
		{Name: "position", Type: "string"},
		{Name: "notes", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "due", Type: "string"},
		{Name: "completed", Type: "string"},
		{Name: "deleted", Type: "boolean"},
		{Name: "hidden", Type: "boolean"},
		{Name: "tasklist_id", Type: "string"},
	}
}

func taskListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"kind":      item["kind"],
		"id":        item["id"],
		"etag":      item["etag"],
		"title":     item["title"],
		"updated":   item["updated"],
		"self_link": item["selfLink"],
	}
}

func taskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"kind":      item["kind"],
		"id":        item["id"],
		"etag":      item["etag"],
		"title":     item["title"],
		"updated":   item["updated"],
		"self_link": item["selfLink"],
		"parent":    item["parent"],
		"position":  item["position"],
		"notes":     item["notes"],
		"status":    item["status"],
		"due":       item["due"],
		"completed": item["completed"],
		"deleted":   item["deleted"],
		"hidden":    item["hidden"],
	}
}
