package codefresh

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Codefresh API resource path (relative
// to base_url), the dotted JSON path where its records array lives in the
// response body, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Codefresh list endpoint path segment (e.g. "projects").
	resource string
	// recordsPath is the dotted path to the records array in the response body.
	// "" selects the root (a bare top-level JSON array).
	recordsPath string
	// mapRecord flattens a raw Codefresh object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// codefreshStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in codefreshStreams; the
// read path is fully data-driven from this table.
var codefreshStreamEndpoints = map[string]streamEndpoint{
	"projects":  {resource: "projects", recordsPath: "projects", mapRecord: codefreshProjectRecord},
	"pipelines": {resource: "pipelines", recordsPath: "docs", mapRecord: codefreshPipelineRecord},
	"agents":    {resource: "agents", recordsPath: "", mapRecord: codefreshAgentRecord},
	"contexts":  {resource: "contexts", recordsPath: "", mapRecord: codefreshContextRecord},
}

// codefreshStreams returns the connector's published stream catalog.
func codefreshStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "projects",
			Description: "Codefresh projects grouping pipelines.",
			PrimaryKey:  []string{"id"},
			Fields:      codefreshProjectFields(),
		},
		{
			Name:        "pipelines",
			Description: "Codefresh pipelines (workflow definitions).",
			PrimaryKey:  []string{"id"},
			Fields:      codefreshPipelineFields(),
		},
		{
			Name:        "agents",
			Description: "Codefresh runner agents.",
			PrimaryKey:  []string{"id"},
			Fields:      codefreshAgentFields(),
		},
		{
			Name:        "contexts",
			Description: "Codefresh shared configuration contexts (integrations).",
			PrimaryKey:  []string{"id"},
			Fields:      codefreshContextFields(),
		},
	}
}

func codefreshProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "project_name", Type: "string"},
		{Name: "favorite", Type: "boolean"},
		{Name: "pipelines_number", Type: "integer"},
		{Name: "updated_at", Type: "string"},
	}
}

func codefreshPipelineFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "project", Type: "string"},
		{Name: "is_public", Type: "boolean"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func codefreshAgentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func codefreshContextFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "owner", Type: "string"},
	}
}

// stringAt returns item[key] as a string when it is a string, else "".
func stringAt(item map[string]any, key string) string {
	if v, ok := item[key].(string); ok {
		return v
	}
	return ""
}

// metadataField reads a nested value from item["metadata"][key] when present.
func metadataField(item map[string]any, key string) any {
	if meta, ok := item["metadata"].(map[string]any); ok {
		return meta[key]
	}
	return nil
}

// recordID resolves a Codefresh object's identity. Codefresh uses "id", "_id",
// or a name under metadata depending on the resource; the mappers normalise to
// "id" so the catalog primary key is uniform.
func recordID(item map[string]any) any {
	if v := item["id"]; v != nil {
		return v
	}
	if v := item["_id"]; v != nil {
		return v
	}
	if v := metadataField(item, "name"); v != nil {
		return v
	}
	if v := item["projectName"]; v != nil {
		return v
	}
	return nil
}

func codefreshProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               recordID(item),
		"project_name":     item["projectName"],
		"favorite":         item["favorite"],
		"pipelines_number": item["pipelinesNumber"],
		"updated_at":       item["updatedAt"],
	}
}

func codefreshPipelineRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         recordID(item),
		"name":       metadataField(item, "name"),
		"project":    metadataField(item, "project"),
		"is_public":  metadataField(item, "isPublic"),
		"created_at": metadataField(item, "created_at"),
		"updated_at": metadataField(item, "updated_at"),
	}
}

func codefreshAgentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         recordID(item),
		"name":       item["name"],
		"version":    item["version"],
		"status":     item["status"],
		"created_at": item["created_at"],
	}
}

func codefreshContextRecord(item map[string]any) connectors.Record {
	owner := item["owner"]
	if owner == nil {
		owner = metadataField(item, "owner")
	}
	return connectors.Record{
		"id":    recordID(item),
		"name":  metadataField(item, "name"),
		"type":  item["type"],
		"owner": owner,
	}
}
