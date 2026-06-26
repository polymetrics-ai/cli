package n8n

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the n8n public API resource path (relative
// to the /api/v1 base) it reads from, and the record mapper that flattens its
// objects.
type streamEndpoint struct {
	// resource is the n8n list endpoint path segment (e.g. "workflows").
	resource string
	// mapRecord flattens a raw n8n object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// n8nStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in n8nStreams; the read path is
// fully data-driven from this table. Every endpoint shares n8n's cursor
// pagination ({data:[...], nextCursor:"..."}).
var n8nStreamEndpoints = map[string]streamEndpoint{
	"workflows":  {resource: "workflows", mapRecord: n8nWorkflowRecord},
	"executions": {resource: "executions", mapRecord: n8nExecutionRecord},
	"tags":       {resource: "tags", mapRecord: n8nTagRecord},
	"users":      {resource: "users", mapRecord: n8nUserRecord},
}

// n8nStreams returns the connector's published stream catalog. n8n objects expose
// a string id and createdAt/updatedAt timestamps, so the primary key is ["id"]
// and the incremental cursor field is ["updatedAt"] where present.
func n8nStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "workflows",
			Description:  "n8n workflows.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       n8nWorkflowFields(),
		},
		{
			Name:         "executions",
			Description:  "n8n workflow executions.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"startedAt"},
			Fields:       n8nExecutionFields(),
		},
		{
			Name:         "tags",
			Description:  "n8n workflow tags.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       n8nTagFields(),
		},
		{
			Name:         "users",
			Description:  "n8n instance users.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       n8nUserFields(),
		},
	}
}

func n8nWorkflowFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "isArchived", Type: "boolean"},
		{Name: "createdAt", Type: "timestamp"},
		{Name: "updatedAt", Type: "timestamp"},
		{Name: "triggerCount", Type: "integer"},
		{Name: "versionId", Type: "string"},
	}
}

func n8nExecutionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "workflowId", Type: "string"},
		{Name: "finished", Type: "boolean"},
		{Name: "mode", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "retryOf", Type: "string"},
		{Name: "startedAt", Type: "timestamp"},
		{Name: "stoppedAt", Type: "timestamp"},
	}
}

func n8nTagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "createdAt", Type: "timestamp"},
		{Name: "updatedAt", Type: "timestamp"},
	}
}

func n8nUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "isPending", Type: "boolean"},
		{Name: "role", Type: "string"},
		{Name: "createdAt", Type: "timestamp"},
		{Name: "updatedAt", Type: "timestamp"},
	}
}

func n8nWorkflowRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"active":       item["active"],
		"isArchived":   item["isArchived"],
		"createdAt":    item["createdAt"],
		"updatedAt":    item["updatedAt"],
		"triggerCount": item["triggerCount"],
		"versionId":    item["versionId"],
	}
}

func n8nExecutionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"workflowId": item["workflowId"],
		"finished":   item["finished"],
		"mode":       item["mode"],
		"status":     item["status"],
		"retryOf":    item["retryOf"],
		"startedAt":  item["startedAt"],
		"stoppedAt":  item["stoppedAt"],
	}
}

func n8nTagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
	}
}

func n8nUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"email":     item["email"],
		"firstName": item["firstName"],
		"lastName":  item["lastName"],
		"isPending": item["isPending"],
		"role":      item["role"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
	}
}
